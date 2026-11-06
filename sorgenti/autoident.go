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
// Regole (4.3):
//   - con l'opzione attiva il programma, in qualsiasi scheda ci si trovi,
//     controlla ogni secondo le porte dei cavi di programmazione e i dispositivi
//     USB delle radio DMR;
//   - ogni cavo ha la sua ricerca, indipendente dalle altre: più cavi (più radio)
//     si cercano INSIEME, ognuno sulla sua porta;
//   - su un cavo senza radio (o con la radio spenta) la ricerca si ripete subito,
//     di continuo e in silenzio: una radio accesa dopo viene trovata al giro
//     successivo (le radio già riconosciute in passato si provano per prime,
//     vedi identPlan e Settings.KnownRadios);
//   - il riquadro «Identificazione Radio In Corso» compare solo mentre una radio
//     risponde davvero (byte che non sono l'eco del cavo) e sparisce quando è
//     riconosciuta o non viene riconosciuta;
//   - quando una radio è riconosciuta l'interfaccia ne mostra i dati e apre la sua
//     scheda (o avvisa che la scheda è già aperta); la ricerca su QUEL cavo si
//     ferma (l'opzione resta attiva) finché non si preme «Identifica Radio» o non
//     si ricollega il cavo; gli altri cavi continuano;
//   - una radio che risponde ma non viene riconosciuta (modello non gestito) si
//     riprova solo ogni autoRetryUnknown, per non far lampeggiare il riquadro;
//   - si interrogano solo le porte dei cavi di programmazione (adattatori
//     USB-seriale CH340, Prolific, Silicon Labs, FTDI…), mai porte Bluetooth,
//     modem o porte seriali di sistema;
//   - le ricerche automatiche lasciano sempre il posto alle operazioni chieste
//     dall'utente (yieldAuto): lettura, scrittura, firmware, inoltro e ricerca
//     manuale le interrompono subito.

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	autoTick         = time.Second             // controllo di porte e dispositivi
	autoSettle       = 1500 * time.Millisecond // attesa dopo il collegamento del cavo
	autoRetryFast    = 300 * time.Millisecond  // 4.3: pausa tra due giri di ricerca (ascolto continuo)
	autoRetryUnknown = 30 * time.Second        // radio che risponde ma non è riconosciuta
	autoQuiet        = 1500 * time.Millisecond // pausa dopo un saluto interrotto (la radio torna pronta)
	autoEventsMax    = 16                      // esiti ricordati per l'interfaccia
	// usbRadioPresentFunc: c'è un dispositivo USB di una radio DMR conosciuta
	// (impostata dalla versione Windows)
	usbRadioPresentFunc = func() bool { return false }
)

type autoPort struct {
	name, desc string
	strong     bool // adattatore USB-seriale riconosciuto dal nome
	next       time.Time
	tries      int
	done       bool // ricerca ferma (radio riconosciuta su questa porta)
	paused     bool // ricerca interrotta dall'utente («Interrompi»)
	gone       bool // porta sparita (cavo staccato dal PC)
	unknown    bool // l'ultima ricerca ha avuto risposte ma nessun modello riconosciuto
	launching  bool // 4.7: ricerca in partenza (vedi autoLaunch)
	run        *autoRun
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
	events     []*autoEvent          // ultimi esiti (radio riconosciute), il più recente per ultimo
	sessions   map[int]*identSession // radio di ogni esito (identify/open con "seq")
	status     string
	quietUntil time.Time
	stop       chan struct{}
	// live: tutte le ricerche automatiche in corso (con un mutex proprio: si
	// aggiorna da startIdentify, chiamata con mu già preso)
	liveMu sync.Mutex
	live   map[*cableJob]bool
}

func (x *autoIdent) liveAdd(j *cableJob, on bool) {
	x.liveMu.Lock()
	defer x.liveMu.Unlock()
	if x.live == nil {
		x.live = map[*cableJob]bool{}
	}
	if on {
		x.live[j] = true
	} else {
		delete(x.live, j)
	}
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
	// 4.5: in ordine progressivo (COM1, COM2, COM3…): anche le ricerche partono così
	sort.Slice(all, func(i, j int) bool { return portLess(all[i].name, all[j].name) })
	if a.auto.usb != nil {
		all = append(all, a.auto.usb)
	}
	return all
}

// autoStep: un controllo di porte e dispositivi; avvia le ricerche delle porte
// pronte (una per porta, in parallelo) senza aspettarne la fine
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
			if p.run != nil && p.run.job.cancel != nil {
				p.run.job.cancel()
			}
			delete(a.auto.ports, n) // cavo staccato dal PC
		}
	}
	if usb && !a.auto.usbWas {
		a.auto.usb = &autoPort{name: "USB", desc: "radio DMR collegata via USB", strong: true, next: now.Add(autoSettle)}
	}
	if !usb && a.auto.usb != nil {
		a.auto.usb.gone = true
		if r := a.auto.usb.run; r != nil && r.job.cancel != nil {
			r.job.cancel()
		}
		a.auto.usb = nil
	}
	a.auto.usbWas = usb

	// ricerche: una per porta, tutte insieme
	quiet := now.Before(a.auto.quietUntil)
	var launch []*autoPort
	for _, p := range a.autoAllLocked() {
		if r := p.run; r != nil {
			if p.gone && r.job.cancel != nil {
				r.job.cancel() // cavo staccato durante la ricerca
			}
			continue
		}
		if p.launching || p.done || p.paused || p.gone || p.next.After(now) || userBusy || quiet {
			continue
		}
		p.launching = true
		launch = append(launch, p)
	}
	a.autoStatusLocked()
	a.auto.mu.Unlock()
	// 4.7: le ricerche partono senza tenere bloccato lo stato dell'identificazione
	// automatica (prima, se l'avvio aspettava, si fermava anche l'aggiornamento
	// dell'interfaccia, che lo chiede ogni secondo e mezzo)
	for _, p := range launch {
		a.autoLaunch(p)
	}
}

func (a *App) autoStatusLocked() {
	all := a.autoAllLocked()
	sort.Slice(all, func(i, j int) bool { return portLess(all[i].name, all[j].name) })
	var answering, listening, found, unknown, paused []string
	for _, p := range all {
		switch {
		case p.run != nil && p.run.job.answered.Load():
			answering = append(answering, p.name)
		case p.run != nil:
			listening = append(listening, p.name)
		case p.done:
			found = append(found, p.name)
		case p.paused:
			paused = append(paused, p.name)
		case p.unknown:
			unknown = append(unknown, p.name)
		default:
			listening = append(listening, p.name)
		}
	}
	var parts []string
	if len(answering) > 0 {
		parts = append(parts, "Una radio risponde su "+strings.Join(answering, ", ")+": identificazione in corso…")
	}
	if len(listening) > 0 {
		parts = append(parts, "In ascolto su "+strings.Join(listening, ", ")+" (collega la radio e accendila, anche dopo il cavo)")
	}
	if len(unknown) > 0 {
		parts = append(parts, "Su "+strings.Join(unknown, ", ")+" una radio risponde ma non è tra i modelli conosciuti")
	}
	if len(found) > 0 {
		parts = append(parts, "Radio riconosciuta su "+strings.Join(found, ", ")+": per un'altra radio premi «Identifica Radio»")
	}
	if len(paused) > 0 {
		parts = append(parts, "Ricerca interrotta su "+strings.Join(paused, ", "))
	}
	if len(all) == 0 {
		parts = append(parts, "In attesa del collegamento del cavo di programmazione")
	}
	a.auto.status = strings.Join(parts, ". ")
}

// portLess (4.5): ordine progressivo delle porte, su ogni sistema: COM1, COM2,
// COM3, COM4… COM10 (Windows), /dev/ttyUSB0… /dev/ttyUSB10 (Linux),
// /dev/cu.usbserial-110… (macOS). I nomi si confrontano a pezzi: le lettere
// senza distinguere maiuscole e minuscole, i numeri per valore; prima le porte
// COM, poi le altre
func portLess(x, y string) bool {
	_, okx := comNumber(x)
	_, oky := comNumber(y)
	if okx != oky {
		return okx
	}
	if c := natCompare(x, y); c != 0 {
		return c < 0
	}
	return x < y
}

// natCompare: confronto «naturale» (a2 prima di a10)
func natCompare(x, y string) int {
	i, j := 0, 0
	for i < len(x) && j < len(y) {
		dx, dy := isDigit(x[i]), isDigit(y[j])
		if dx && dy {
			si, sj := i, j
			for i < len(x) && isDigit(x[i]) {
				i++
			}
			for j < len(y) && isDigit(y[j]) {
				j++
			}
			a, b := strings.TrimLeft(x[si:i], "0"), strings.TrimLeft(y[sj:j], "0")
			if len(a) != len(b) {
				if len(a) < len(b) {
					return -1
				}
				return 1
			}
			if a != b {
				if a < b {
					return -1
				}
				return 1
			}
			continue
		}
		cx, cy := lowerASCII(x[i]), lowerASCII(y[j])
		if dx != dy {
			// un numero viene prima delle lettere (COM1 prima di COMA)
			if dx {
				return -1
			}
			return 1
		}
		if cx != cy {
			if cx < cy {
				return -1
			}
			return 1
		}
		i++
		j++
	}
	switch {
	case len(x)-i < len(y)-j:
		return -1
	case len(x)-i > len(y)-j:
		return 1
	}
	return 0
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

func comNumber(s string) (int, bool) {
	u := strings.ToUpper(s)
	if !strings.HasPrefix(u, "COM") {
		return 0, false
	}
	n := 0
	for _, c := range u[3:] {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, len(u) > 3
}

// autoLaunch: avvia la ricerca su p in un goroutine (senza a.auto.mu)
func (a *App) autoLaunch(p *autoPort) {
	j, err := a.startIdentify(&identReq{Port: p.name}, true)
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	p.launching = false
	if err != nil {
		return // un'altra operazione è appena partita: si riprova al prossimo controllo
	}
	if p.gone || p.paused || !a.autoEnabledLocked() {
		// cavo staccato, ricerca interrotta o opzione disattivata nel frattempo
		if j.cancel != nil {
			j.cancel()
		}
		return
	}
	r := &autoRun{p: p, job: j}
	p.run = r
	go a.autoWait(r)
	a.autoStatusLocked()
}

// autoEnabledLocked: come autoEnabled (le impostazioni hanno un loro blocco)
func (a *App) autoEnabledLocked() bool { return a.autoEnabled() }

// autoWait: attende la fine della ricerca e ne registra l'esito
func (a *App) autoWait(r *autoRun) {
	j := r.job
	for j.isRunning() {
		time.Sleep(50 * time.Millisecond)
	}
	j.mu.Lock()
	state, res := j.State, j.Result
	j.mu.Unlock()
	now := time.Now()
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	p := r.p
	if p.run == r {
		p.run = nil
	}
	if state == "cancelled" {
		if q := now.Add(autoQuiet); q.After(a.auto.quietUntil) {
			a.auto.quietUntil = q
		}
	}
	if p.gone {
		return // cavo staccato durante la ricerca: la porta ricomparirà come nuova
	}
	p.tries++
	ir, _ := res.(*identResult)
	switch {
	case state == "done" && ir != nil && ir.Found:
		// radio riconosciuta: la ricerca su questa porta si ferma (l'opzione resta
		// attiva; si riprende con «Identifica Radio»)
		p.done, p.unknown = true, false
		a.auto.seq++
		ev := &autoEvent{Seq: a.auto.seq, Port: p.name, Result: ir, When: now.Format("15:04:05")}
		a.auto.events = append(a.auto.events, ev)
		if a.auto.sessions == nil {
			a.auto.sessions = map[int]*identSession{}
		}
		a.auto.sessions[ev.Seq] = j.sess
		for len(a.auto.events) > autoEventsMax {
			delete(a.auto.sessions, a.auto.events[0].Seq)
			a.auto.events = a.auto.events[1:]
		}
	case state == "cancelled":
		p.next = now.Add(autoQuiet)
		p.tries-- // la ricerca non è arrivata in fondo
	case j.answered.Load():
		// una radio ha risposto davvero (vedi probeState.sure) ma non è stata
		// riconosciuta. 4.4: non basta una risposta qualsiasi dei driver (Answers):
		// i disturbi di un cavo senza radio sembravano una radio sconosciuta
		p.unknown = true
		p.next = now.Add(autoRetryUnknown)
	default:
		// nessuna risposta: si ascolta di nuovo subito (la radio può essere accesa dopo)
		p.unknown = false
		p.next = now.Add(autoRetryFast)
	}
	a.autoStatusLocked()
}

// autoRescan: nuova ricerca su tutti i cavi collegati (pulsante «Identifica
// Radio»); restituisce il numero di porte da cercare
func (a *App) autoRescan() int {
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	now := time.Now()
	n := 0
	for _, p := range a.autoAllLocked() {
		p.done, p.paused, p.unknown, p.tries, p.next = false, false, false, 0, now
		n++
	}
	return n
}

// autoIdentSession: radio riconosciuta dall'esito automatico seq (nil se nel
// frattempo ce n'è stato un altro)
func (a *App) autoIdentSession(seq int) *identSession {
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	return a.auto.sessions[seq]
}

// yieldAuto: interrompe la ricerca automatica in corso, per lasciare il cavo
// all'operazione chiesta dall'utente; attende al massimo pochi secondi, più la
// breve pausa che serve alla radio per tornare pronta dopo un saluto interrotto
func (a *App) yieldAuto() {
	var runs []*cableJob
	a.auto.liveMu.Lock()
	for j := range a.auto.live {
		runs = append(runs, j)
	}
	a.auto.liveMu.Unlock()
	for _, j := range runs {
		if j.cancel != nil {
			j.cancel()
		}
	}
	for _, j := range runs {
		for i := 0; i < 80 && j.isRunning(); i++ {
			time.Sleep(50 * time.Millisecond)
		}
	}
	a.auto.mu.Lock()
	if len(runs) > 0 {
		if q := time.Now().Add(autoQuiet); q.After(a.auto.quietUntil) {
			a.auto.quietUntil = q
		}
	}
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
		Rescan  bool  `json:"rescan"` // «Identifica Radio»
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
	a.autoStatusLocked()
	all := a.autoAllLocked()
	sort.Slice(all, func(i, j int) bool { return portLess(all[i].name, all[j].name) })
	ports, searchPorts, answering := []string{}, []string{}, []string{}
	paused := false
	for _, p := range all {
		ports = append(ports, p.name)
		paused = paused || p.paused
		if p.run != nil {
			searchPorts = append(searchPorts, p.name)
			if p.run.job.answered.Load() {
				answering = append(answering, p.name) // una radio ha risposto: riquadro visibile
			}
		}
	}
	events := a.auto.events
	if events == nil {
		events = []*autoEvent{}
	}
	out := map[string]interface{}{"enabled": a.autoEnabled(), "status": a.auto.status, "ports": ports,
		"searching": len(answering) > 0, "answering": answering, "busy": len(searchPorts) > 0, "searchPorts": searchPorts,
		"seq": a.auto.seq, "paused": paused, "events": events}
	if len(searchPorts) > 0 {
		out["searchPort"] = searchPorts[0]
	}
	if rescanned >= 0 {
		out["rescanned"] = rescanned
	}
	if n := len(a.auto.events); n > 0 {
		out["last"] = a.auto.events[n-1]
	}
	return out, nil
}
