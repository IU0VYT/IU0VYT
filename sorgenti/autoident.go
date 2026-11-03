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
// Regole (4.2, come richiesto):
//   - con l'opzione attiva il programma, in qualsiasi scheda ci si trovi,
//     controlla ogni secondo le porte dei cavi di programmazione e i dispositivi
//     USB delle radio DMR, anche mentre una ricerca è in corso;
//   - un cavo appena collegato (porta nuova) avvia la ricerca della radio, con gli
//     stessi comandi di sola lettura della scheda «Identifica». Se nessuna radio
//     risponde la ricerca si ripete in silenzio (la radio può essere accesa dopo);
//   - il riquadro «Identificazione Radio In Corso» compare solo quando una radio
//     risponde davvero (byte ricevuti che non sono l'eco del cavo) e sparisce
//     quando la radio è riconosciuta o la ricerca finisce senza riconoscerla: con
//     il cavo senza radio non compare;
//   - quando la radio è riconosciuta l'interfaccia ne mostra i dati e apre la sua
//     scheda; subito dopo la ricerca si ferma (l'opzione resta attiva): nessuna
//     porta viene più interrogata finché non si preme «Identifica Nuova Radio
//     Automaticamente» (barra in alto) o non si collega un altro cavo;
//   - si interrogano solo le porte dei cavi di programmazione (adattatori
//     USB-seriale CH340, Prolific, Silicon Labs, FTDI…), mai porte Bluetooth,
//     modem o porte seriali di sistema; gli adattatori noti hanno la precedenza;
//   - una ricerca automatica lascia sempre il posto alle operazioni chieste
//     dall'utente (yieldAuto): lettura, scrittura, firmware, inoltro e ricerca
//     manuale la interrompono subito.

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	autoTick      = time.Second             // controllo di porte e dispositivi
	autoSettle    = 1500 * time.Millisecond // attesa dopo il collegamento del cavo
	autoRetryFast = 3 * time.Second         // pausa tra le ricerche nei primi minuti
	autoRetrySlow = 20 * time.Second        // pausa tra le ricerche dopo autoFastTries
	autoFastTries = 8
	autoQuiet     = 1500 * time.Millisecond // pausa dopo un saluto interrotto (la radio torna pronta)
	// usbRadioPresentFunc: c'è un dispositivo USB di una radio DMR conosciuta
	// (impostata dalla versione Windows)
	usbRadioPresentFunc = func() bool { return false }
)

type autoPort struct {
	name, desc string
	strong     bool // adattatore USB-seriale riconosciuto dal nome
	next       time.Time
	tries      int
	done       bool // ricerca ferma (radio riconosciuta qui o altrove)
	paused     bool // ricerca interrotta dall'utente («Interrompi»)
	gone       bool // porta sparita (cavo staccato dal PC)
}

// autoEvent: esito dell'ultima ricerca automatica che ha trovato la radio
type autoEvent struct {
	Seq    int          `json:"seq"`
	Port   string       `json:"port"`
	Result *identResult `json:"result"`
	When   string       `json:"when"`
}

type autoRun struct {
	p   *autoPort
	job *cableJob
}

type autoIdent struct {
	mu         sync.Mutex
	ports      map[string]*autoPort
	usbWas     bool
	usb        *autoPort
	seq        int
	last       *autoEvent
	ident      *identSession // radio dell'ultimo esito (per identify/open con "seq")
	status     string
	run        *autoRun
	quietUntil time.Time
	stop       chan struct{}
}

// cableLikePort: la porta è quella di un adattatore USB-seriale (cavo di
// programmazione)? Sconosciuta (nessuna descrizione): sì, salvo le porte
// seriali di sistema COM1/COM2.
func cableLikePort(p PortInfo) bool { ok, _ := cablePortKind(p); return ok }

// cablePortKind: come cableLikePort; strong: il nome è quello di un adattatore
// USB-seriale noto (CH340, Prolific, Silicon Labs, FTDI)
func cablePortKind(p PortInfo) (ok, strong bool) {
	d := strings.ToLower(p.Desc)
	for _, no := range []string{"bluetooth", "modem", "communications port", "porta di comunicazione", "printer", "stampante",
		"intel(r) active management", "serial over lan"} {
		if strings.Contains(d, no) {
			return false, false
		}
	}
	if d == "" {
		n := strings.ToUpper(p.Name)
		return n != "COM1" && n != "COM2", false
	}
	for _, yes := range []string{"ch340", "ch341", "ch343", "ch9102", "wch", "pl2303", "prolific", "cp210", "silicon labs", "ftdi", "ft232"} {
		if strings.Contains(d, yes) {
			return true, true
		}
	}
	for _, yes := range []string{"usb-serial", "usb serial", "usb-to-serial", "usb to serial", "usb uart", "usb-uart", "cdc", "usb-sériel", "seriale usb", "seriell"} {
		if strings.Contains(d, yes) {
			return true, false
		}
	}
	return false, false
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
	if a.auto.ports == nil {
		a.auto.ports = map[string]*autoPort{}
	}
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

func (a *App) autoAllLocked() []*autoPort {
	var all []*autoPort
	for _, p := range a.auto.ports {
		all = append(all, p)
	}
	if a.auto.usb != nil {
		all = append(all, a.auto.usb)
	}
	return all
}

// autoStep: un controllo di porte e dispositivi; avvia al massimo una ricerca,
// senza aspettarne la fine
func (a *App) autoStep(now time.Time) {
	if !a.autoEnabled() {
		a.auto.mu.Lock()
		a.auto.ports = map[string]*autoPort{}
		a.auto.usbWas, a.auto.usb = false, nil
		a.auto.status = ""
		a.auto.mu.Unlock()
		return
	}
	ports := listPortsFunc()
	usb := usbRadioPresentFunc()
	a.mu.Lock()
	userBusy := a.job != nil && a.job.isRunning() && !a.job.Auto
	a.mu.Unlock()

	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	if a.auto.ports == nil {
		a.auto.ports = map[string]*autoPort{}
	}
	seen := map[string]bool{}
	for _, p := range ports {
		ok, strong := cablePortKind(p)
		if !ok {
			continue
		}
		seen[p.Name] = true
		if a.auto.ports[p.Name] == nil {
			a.auto.ports[p.Name] = &autoPort{name: p.Name, desc: p.Desc, strong: strong, next: now.Add(autoSettle)}
		}
	}
	for n, p := range a.auto.ports {
		if !seen[n] {
			p.gone = true
			delete(a.auto.ports, n) // cavo staccato dal PC
		}
	}
	if usb && !a.auto.usbWas {
		a.auto.usb = &autoPort{name: "USB", desc: "radio DMR collegata via USB", strong: true, next: now.Add(autoSettle)}
	}
	if !usb && a.auto.usb != nil {
		a.auto.usb.gone = true
		a.auto.usb = nil
	}
	a.auto.usbWas = usb

	// ricerche da fare: prima le porte appena collegate, poi gli adattatori
	// noti, poi la più vecchia
	var cands []*autoPort
	for _, p := range a.autoAllLocked() {
		if !p.done && !p.paused {
			cands = append(cands, p)
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		x, y := cands[i], cands[j]
		if (x.tries == 0) != (y.tries == 0) {
			return x.tries == 0
		}
		if x.strong != y.strong {
			return x.strong
		}
		return x.next.Before(y.next)
	})
	var due *autoPort
	for _, p := range cands {
		if !p.next.After(now) {
			due = p
			break
		}
	}
	if r := a.auto.run; r != nil {
		// cavo staccato dal PC durante la ricerca, oppure un cavo appena collegato
		// mentre si ripete la ricerca su un'altra porta: si interrompe
		if r.p.gone || (due != nil && due != r.p && due.tries == 0 && r.p.tries > 0) {
			if r.job.cancel != nil {
				r.job.cancel()
			}
		}
		a.autoStatusLocked(cands, due)
		return
	}
	if !userBusy && due != nil {
		a.autoLaunchLocked(due)
	}
	a.autoStatusLocked(cands, due)
}

func (a *App) autoStatusLocked(cands []*autoPort, due *autoPort) {
	if r := a.auto.run; r != nil {
		if r.job.answered.Load() {
			a.auto.status = "Una radio risponde su " + r.p.name + ": identificazione in corso…"
		} else {
			a.auto.status = "Controllo del cavo su " + r.p.name + " (nessuna radio ha ancora risposto)"
		}
		return
	}
	all := a.autoAllLocked()
	done, paused := false, false
	for _, p := range all {
		done = done || p.done
		paused = paused || p.paused
	}
	switch {
	case len(all) == 0:
		a.auto.status = "In attesa del collegamento del cavo di programmazione"
	case len(cands) == 0 && done && a.auto.last != nil:
		a.auto.status = "Radio riconosciuta: " + a.auto.last.Result.Name + " (" + a.auto.last.Port +
			"). Ricerca ferma: per un'altra radio premi «Identifica Nuova Radio Automaticamente»"
	case len(cands) == 0 && paused:
		a.auto.status = "Ricerca interrotta: premi «Identifica Nuova Radio Automaticamente» o ricollega il cavo"
	case len(cands) == 0:
		a.auto.status = "Ricerca ferma: per cercare una radio premi «Identifica Nuova Radio Automaticamente»"
	case due == nil:
		a.auto.status = "Nessuna radio risponde (" + cands[0].name + "): collegala al cavo e accendila con il volume a metà"
	}
}

// autoLaunchLocked: avvia la ricerca su p in un goroutine; con a.auto.mu già preso
func (a *App) autoLaunchLocked(p *autoPort) {
	j, err := a.startIdentify(&identReq{Port: p.name}, true)
	if err != nil {
		return // un'altra operazione è appena partita: si riprova al prossimo controllo
	}
	r := &autoRun{p: p, job: j}
	a.auto.run = r
	go a.autoWait(r)
}

// autoWait: attende la fine della ricerca e ne registra l'esito
func (a *App) autoWait(r *autoRun) {
	j := r.job
	for j.isRunning() {
		time.Sleep(50 * time.Millisecond)
	}
	j.mu.Lock()
	state, res := j.State, j.Result
	j.mu.Unlock()
	a.mu.Lock()
	id := a.ident
	a.mu.Unlock()
	now := time.Now()
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	if a.auto.run == r {
		a.auto.run = nil
	}
	if state == "cancelled" {
		a.auto.quietUntil = now.Add(autoQuiet)
	}
	p := r.p
	if p.gone {
		return // cavo staccato durante la ricerca: la porta ricomparirà come nuova
	}
	p.tries++
	ir, _ := res.(*identResult)
	switch {
	case state == "done" && ir != nil && ir.Found:
		// radio riconosciuta: la ricerca si ferma su tutte le porte (l'opzione
		// resta attiva; si riprende con «Identifica Nuova Radio Automaticamente»)
		for _, x := range a.autoAllLocked() {
			x.done = true
		}
		a.auto.seq++
		a.auto.last = &autoEvent{Seq: a.auto.seq, Port: p.name, Result: ir, When: now.Format("15:04:05")}
		a.auto.ident = id
	case state == "cancelled":
		p.next = now.Add(autoRetryFast)
		p.tries-- // la ricerca non è arrivata in fondo
	default:
		d := autoRetryFast
		if p.tries >= autoFastTries {
			d = autoRetrySlow
		}
		p.next = now.Add(d)
	}
}

// autoRescan: nuova ricerca su tutti i cavi collegati (pulsante «Identifica Nuova
// Radio Automaticamente»); restituisce il numero di porte da cercare
func (a *App) autoRescan() int {
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	now := time.Now()
	n := 0
	for _, p := range a.autoAllLocked() {
		p.done, p.paused, p.tries, p.next = false, false, 0, now
		n++
	}
	return n
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
// all'operazione chiesta dall'utente; attende al massimo pochi secondi, più la
// breve pausa che serve alla radio per tornare pronta dopo un saluto interrotto
func (a *App) yieldAuto() {
	a.mu.Lock()
	j := a.job
	a.mu.Unlock()
	if j != nil && j.Auto && j.isRunning() {
		if j.cancel != nil {
			j.cancel()
		}
		for i := 0; i < 80 && j.isRunning(); i++ {
			time.Sleep(50 * time.Millisecond)
		}
		a.auto.mu.Lock()
		if q := time.Now().Add(autoQuiet); q.After(a.auto.quietUntil) {
			a.auto.quietUntil = q
		}
		a.auto.mu.Unlock()
	}
	a.auto.mu.Lock()
	wait := time.Until(a.auto.quietUntil)
	a.auto.mu.Unlock()
	if wait > 0 && wait <= 2*autoQuiet {
		time.Sleep(wait)
	}
}

// ---------------------------------------------------------------- API

// identify/auto: stato; con "enabled" attivazione o disattivazione, con "stop"
// interruzione della ricerca, con "rescan" nuova ricerca su tutti i cavi
func (a *App) apiIdentifyAuto(r *http.Request) (interface{}, error) {
	var req struct {
		Enabled *bool `json:"enabled"`
		Stop    bool  `json:"stop"`   // riquadro: «Interrompi»
		Rescan  bool  `json:"rescan"` // «Identifica Nuova Radio Automaticamente»
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
	if req.Stop {
		a.auto.mu.Lock()
		for _, p := range a.autoAllLocked() {
			if !p.done {
				p.paused = true
			}
		}
		a.auto.mu.Unlock()
		a.yieldAuto()
	}
	rescanned := -1
	if req.Rescan {
		if !a.autoEnabled() {
			return nil, errors.New("attiva prima l'identificazione automatica nella scheda «Identifica»")
		}
		rescanned = a.autoRescan()
		a.autoStep(time.Now())
	}
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	ports := []string{}
	for _, p := range a.autoAllLocked() {
		ports = append(ports, p.name)
	}
	sort.Strings(ports)
	searching, busy, port := false, false, ""
	if r := a.auto.run; r != nil {
		busy, port = true, r.p.name
		searching = r.job.answered.Load() // una radio ha risposto: riquadro visibile
	}
	paused := false
	for _, p := range a.autoAllLocked() {
		paused = paused || p.paused
	}
	out := map[string]interface{}{"enabled": a.autoEnabled(), "status": a.auto.status, "ports": ports, "searching": searching,
		"busy": busy, "searchPort": port, "seq": a.auto.seq, "paused": paused}
	if rescanned >= 0 {
		out["rescanned"] = rescanned
	}
	if a.auto.last != nil {
		out["last"] = a.auto.last
	}
	return out, nil
}
