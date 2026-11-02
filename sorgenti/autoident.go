// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Identificazione automatica (4.0, scheda «Identifica», pulsante «Abilita
// identificazione automatica»; la scelta resta salvata nelle impostazioni).
//
// Con l'opzione attiva il programma controlla ogni pochi secondi le porte
// seriali e i dispositivi USB delle radio DMR:
//   - quando compare una porta nuova (cavo di programmazione appena collegato)
//     o un dispositivo USB di una radio, avvia da solo la stessa ricerca della
//     scheda «Identifica» (solo comandi di lettura, nessuna scrittura);
//   - se la radio non risponde (ad esempio è ancora spenta) la ricerca si ripete
//     finché il cavo resta collegato: prima ogni pochi secondi, poi più di rado.
//     Così la radio viene riconosciuta anche se si accende dopo aver collegato
//     il cavo;
//   - quando la radio è riconosciuta, l'interfaccia ne mostra i dati e apre la
//     sua scheda con i canali letti. La stessa porta non si interroga più finché
//     il cavo non viene scollegato e ricollegato.
//
// Si interrogano solo le porte dei cavi di programmazione (adattatori USB-seriale
// CH340, Prolific, Silicon Labs, FTDI…), mai porte Bluetooth, modem o porte
// seriali di sistema. Una ricerca automatica lascia sempre il posto alle
// operazioni chieste dall'utente (yieldAuto): lettura, scrittura, firmware,
// inoltro e ricerca manuale la interrompono subito.

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	autoTick      = 1500 * time.Millisecond // controllo di porte e dispositivi
	autoSettle    = 1500 * time.Millisecond // attesa dopo il collegamento del cavo
	autoRetryFast = 3 * time.Second         // pausa tra le ricerche nei primi minuti
	autoRetrySlow = 20 * time.Second        // pausa tra le ricerche dopo autoFastTries
	autoFastTries = 8
	// usbRadioPresentFunc: c'è un dispositivo USB di una radio DMR conosciuta
	// (impostata dalla versione Windows)
	usbRadioPresentFunc = func() bool { return false }
)

type autoPort struct {
	name, desc string
	next       time.Time
	tries      int
	done       bool // radio riconosciuta (o ricerca interrotta): si attende lo scollegamento
}

// autoEvent: esito dell'ultima ricerca automatica che ha trovato la radio
type autoEvent struct {
	Seq    int          `json:"seq"`
	Port   string       `json:"port"`
	Result *identResult `json:"result"`
	When   string       `json:"when"`
}

type autoIdent struct {
	mu      sync.Mutex
	ports   map[string]*autoPort
	usbWas  bool
	usb     *autoPort
	seq     int
	last    *autoEvent
	ident   *identSession // radio dell'ultimo esito (per identify/open con "seq")
	status  string
	running bool
	stop    chan struct{}
}

// cableLikePort: la porta è quella di un adattatore USB-seriale (cavo di
// programmazione)? Sconosciuta (nessuna descrizione): sì, salvo le porte
// seriali di sistema COM1/COM2.
func cableLikePort(p PortInfo) bool {
	d := strings.ToLower(p.Desc)
	for _, no := range []string{"bluetooth", "modem", "communications port", "porta di comunicazione", "printer", "stampante",
		"intel(r) active management", "serial over lan"} {
		if strings.Contains(d, no) {
			return false
		}
	}
	if d == "" {
		n := strings.ToUpper(p.Name)
		return n != "COM1" && n != "COM2"
	}
	for _, yes := range []string{"ch340", "ch341", "ch343", "ch9102", "wch", "usb-serial", "usb serial", "usb-to-serial", "usb to serial",
		"pl2303", "prolific", "cp210", "silicon labs", "ftdi", "ft232", "usb uart", "usb-uart", "cdc", "usb-sériel", "seriale usb", "seriell"} {
		if strings.Contains(d, yes) {
			return true
		}
	}
	return false
}

func (a *App) autoEnabled() bool { return a.store.Get().AutoIdent }

// startAutoIdent: avvia il controllo periodico (una volta, all'avvio del programma)
func (a *App) startAutoIdent() {
	a.auto.mu.Lock()
	if a.auto.stop != nil {
		a.auto.mu.Unlock()
		return
	}
	a.auto.stop = make(chan struct{})
	a.auto.ports = map[string]*autoPort{}
	stop := a.auto.stop
	a.auto.mu.Unlock()
	go func() {
		t := time.NewTicker(autoTick)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				a.autoStep(time.Now())
			}
		}
	}()
}

func (a *App) stopAutoIdent() {
	a.auto.mu.Lock()
	if a.auto.stop != nil {
		close(a.auto.stop)
		a.auto.stop = nil
	}
	a.auto.mu.Unlock()
}

func (a *App) autoSetStatus(s string) {
	a.auto.mu.Lock()
	a.auto.status = s
	a.auto.mu.Unlock()
}

// autoStep: un controllo di porte e dispositivi; avvia al massimo una ricerca
func (a *App) autoStep(now time.Time) {
	if !a.autoEnabled() {
		a.auto.mu.Lock()
		a.auto.ports = map[string]*autoPort{}
		a.auto.usbWas, a.auto.usb = false, nil
		a.auto.status = ""
		a.auto.mu.Unlock()
		return
	}
	a.mu.Lock()
	busy := a.job != nil && a.job.isRunning()
	a.mu.Unlock()
	if busy {
		return
	}
	ports := listPortsFunc()
	usb := usbRadioPresentFunc()
	a.auto.mu.Lock()
	seen := map[string]bool{}
	for _, p := range ports {
		if !cableLikePort(p) {
			continue
		}
		seen[p.Name] = true
		if a.auto.ports[p.Name] == nil {
			a.auto.ports[p.Name] = &autoPort{name: p.Name, desc: p.Desc, next: now.Add(autoSettle)}
		}
	}
	for n := range a.auto.ports {
		if !seen[n] {
			delete(a.auto.ports, n) // cavo scollegato
		}
	}
	if usb && !a.auto.usbWas {
		a.auto.usb = &autoPort{name: "USB", desc: "radio DMR collegata via USB", next: now.Add(autoSettle)}
	}
	if !usb {
		a.auto.usb = nil
	}
	a.auto.usbWas = usb
	// prossima ricerca: la porta in attesa da più tempo
	var cands []*autoPort
	for _, p := range a.auto.ports {
		if !p.done {
			cands = append(cands, p)
		}
	}
	if a.auto.usb != nil && !a.auto.usb.done {
		cands = append(cands, a.auto.usb)
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].next.Before(cands[j].next) })
	var due *autoPort
	if len(cands) > 0 && !cands[0].next.After(now) {
		due = cands[0]
	}
	switch {
	case len(cands) == 0 && len(a.auto.ports) == 0 && a.auto.usb == nil:
		a.auto.status = "In attesa del collegamento di una radio"
	case len(cands) == 0:
		a.auto.status = "Radio collegata già riconosciuta: scollega e ricollega il cavo per una nuova ricerca"
	case due == nil:
		a.auto.status = "In attesa che la radio risponda (" + cands[0].name + "): accendila con il volume a metà"
	}
	a.auto.mu.Unlock()
	if due != nil {
		a.autoSearch(due)
	}
}

// autoSearch: una ricerca sulla porta (o via USB), attesa fino alla fine
func (a *App) autoSearch(p *autoPort) {
	a.autoSetStatus("Ricerca della radio su " + p.name + "…")
	j, err := a.startIdentify(&identReq{Port: p.name}, true)
	if err != nil {
		return // un'altra operazione è appena partita: si riprova al prossimo controllo
	}
	a.auto.mu.Lock()
	a.auto.running = true
	a.auto.mu.Unlock()
	for j.isRunning() {
		time.Sleep(100 * time.Millisecond)
	}
	j.mu.Lock()
	state, res := j.State, j.Result
	j.mu.Unlock()
	a.mu.Lock()
	id := a.ident
	a.mu.Unlock()
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	a.auto.running = false
	p.tries++
	r, _ := res.(*identResult)
	switch {
	case state == "done" && r != nil && r.Found:
		p.done = true
		a.auto.seq++
		a.auto.last = &autoEvent{Seq: a.auto.seq, Port: p.name, Result: r, When: time.Now().Format("15:04:05")}
		a.auto.ident = id
		a.auto.status = "Radio riconosciuta: " + r.Name + " (" + p.name + ")"
	case state == "cancelled":
		// interrotta dall'utente o da un'altra operazione: si riprova dopo
		p.next = time.Now().Add(autoRetryFast)
	default:
		d := autoRetryFast
		if p.tries >= autoFastTries {
			d = autoRetrySlow
		}
		p.next = time.Now().Add(d)
	}
}

// autoIdentSession: radio riconosciuta dall'esito automatico seq (nil se nel
// frattempo ce n'è stato un altro)
func (a *App) autoIdentSession(seq int) *identSession {
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	if a.auto.last == nil || a.auto.last.Seq != seq {
		return nil
	}
	return a.auto.ident
}

// yieldAuto: interrompe la ricerca automatica in corso, per lasciare il cavo
// all'operazione chiesta dall'utente (attende al massimo pochi secondi)
func (a *App) yieldAuto() {
	a.mu.Lock()
	j := a.job
	a.mu.Unlock()
	if j == nil || !j.Auto || !j.isRunning() {
		return
	}
	if j.cancel != nil {
		j.cancel()
	}
	for i := 0; i < 80 && j.isRunning(); i++ {
		time.Sleep(50 * time.Millisecond)
	}
}

// ---------------------------------------------------------------- API

// identify/auto: stato (e, con "enabled", attivazione o disattivazione)
func (a *App) apiIdentifyAuto(r *http.Request) (interface{}, error) {
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Enabled != nil {
		st := a.store.Get()
		st.AutoIdent = *req.Enabled
		if err := a.store.SaveSettings(st); err != nil {
			return nil, err
		}
		if !*req.Enabled {
			a.yieldAuto()
		}
	}
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	var ports []string
	for n := range a.auto.ports {
		ports = append(ports, n)
	}
	if a.auto.usb != nil {
		ports = append(ports, "USB")
	}
	sort.Strings(ports)
	out := map[string]interface{}{"enabled": a.autoEnabled(), "status": a.auto.status, "ports": ports, "searching": a.auto.running, "seq": a.auto.seq}
	if a.auto.last != nil {
		out["last"] = a.auto.last
	}
	if ports == nil {
		out["ports"] = []string{}
	}
	return out, nil
}
