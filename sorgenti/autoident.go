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
//     si fa LEGGERE (4.36: il saluto è andato a buon fine e stanno arrivando i
//     dati della memoria) e sparisce quando è riconosciuta o la lettura cade.
//     Prima bastava una risposta qualsiasi che non fosse l'eco del cavo: una
//     radio spenta ma collegata può rispondere al saluto (resta alimentata dal
//     cavo) senza farsi leggere, e il riquadro compariva per una radio spenta;
//   - quando una radio è riconosciuta l'interfaccia ne mostra i dati e apre la sua
//     scheda (o avvisa che la scheda è già aperta); la ricerca su QUEL cavo si
//     ferma (l'opzione resta attiva) finché non si preme «Identifica Radio» o non
//     si ricollega il cavo; gli altri cavi continuano;
//   - un dispositivo che risponde ma non si lascia leggere (radio spenta o non
//     pronta, oppure modello non gestito): l'interfaccia lo dice una volta, con
//     un avviso (autoNotice); poi si riprovano ogni autoRetryKnown, in silenzio,
//     solo le radio già riconosciute (la radio spenta, una volta accesa, si
//     trova in pochi secondi) e ogni autoRetryUnknown tutto il giro;
//   - si interrogano solo le porte dei cavi di programmazione (adattatori
//     USB-seriale CH340, Prolific, Silicon Labs, FTDI…), mai porte Bluetooth,
//     modem o porte seriali di sistema;
//   - le ricerche automatiche lasciano sempre il posto alle operazioni chieste
//     dall'utente (yieldAuto): lettura, scrittura, firmware, inoltro e ricerca
//     manuale le interrompono subito.

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	autoTick         = time.Second             // controllo di porte e dispositivi
	autoSettle       = 1500 * time.Millisecond // attesa dopo il collegamento del cavo
	autoRetryFast    = 300 * time.Millisecond  // 4.3: pausa tra due giri di ricerca (ascolto continuo)
	autoRetryUnknown = 30 * time.Second        // radio che risponde ma non è riconosciuta: giro completo
	autoRetryKnown   = 3 * time.Second         // 4.36: come sopra, giro delle sole radio già riconosciute
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
	// fullAt (4.36): con unknown, momento del prossimo giro completo; fino ad
	// allora si riprovano solo le radio già riconosciute
	fullAt time.Time
}

// autoNotice (4.36): un dispositivo risponde su una porta ma non si lascia
// leggere; l'interfaccia mostra l'avviso una volta per ogni seq
type autoNotice struct {
	Seq  int    `json:"seq"`
	Port string `json:"port"`
	When string `json:"when"`
}

// autoEvent: esito dell'ultima ricerca automatica che ha trovato la radio
type autoEvent struct {
	Seq    int          `json:"seq"`
	Port   string       `json:"port"`
	Result *identResult `json:"result"`
	When   string       `json:"when"`
}

type autoRun struct {
	p         *autoPort
	job       *cableJob
	knownOnly bool // 4.36: giro delle sole radio già riconosciute
}

type autoIdent struct {
	mu        sync.Mutex
	ports     map[string]*autoPort
	usbWas    bool
	usb       *autoPort
	seq       int
	events    []*autoEvent          // ultimi esiti (radio riconosciute), il più recente per ultimo
	sessions  map[int]*identSession // radio di ogni esito (identify/open con "seq")
	notice    *autoNotice           // 4.36: ultimo avviso «risponde ma non si lascia leggere»
	noticeSeq int
	// openTabs (4.36): schede delle radio aperte nell'interfaccia, come le ha
	// dette l'ultima richiesta di stato (openAt); vedi autoTabOpen
	openTabs   map[string]bool
	openAt     time.Time
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

// autoNoticeLocked (4.36): su p un dispositivo risponde ma non si lascia
// leggere. L'interfaccia mostra l'avviso una volta; da qui in poi la porta è
// «unknown» finché una radio non viene riconosciuta o il dispositivo tace
func (a *App) autoNoticeLocked(p *autoPort, now time.Time) {
	p.unknown = true
	p.fullAt = now.Add(autoRetryUnknown)
	a.auto.noticeSeq++
	a.auto.notice = &autoNotice{Seq: a.auto.noticeSeq, Port: p.name, When: now.Format("15:04:05")}
	identDiag.add(probeKey(p.name), "*** un dispositivo risponde ma non si lascia leggere: avviso all'utente")
}

func (a *App) autoStatusLocked() {
	all := a.autoAllLocked()
	sort.Slice(all, func(i, j int) bool { return portLess(all[i].name, all[j].name) })
	for _, p := range all {
		// una radio già nota ha risposto senza farsi leggere: avviso subito,
		// mentre il giro continua (la radio accesa nel frattempo viene trovata)
		if p.run != nil && !p.unknown && p.run.job.knownFail.Load() {
			a.autoNoticeLocked(p, time.Now())
		}
	}
	var answering, checking, listening, found, unknown, paused []string
	for _, p := range all {
		switch {
		case p.run != nil && p.run.job.answered.Load() && p.run.job.reading.Load():
			answering = append(answering, p.name)
		case p.unknown && !p.paused:
			// anche durante i giri di controllo, che sono silenziosi
			unknown = append(unknown, p.name)
		case p.run != nil && p.run.job.answered.Load():
			checking = append(checking, p.name)
		case p.run != nil:
			listening = append(listening, p.name)
		case p.done:
			found = append(found, p.name)
		case p.paused:
			paused = append(paused, p.name)
		default:
			listening = append(listening, p.name)
		}
	}
	var parts []string
	if len(answering) > 0 {
		parts = append(parts, "Una radio risponde su "+strings.Join(answering, ", ")+": identificazione in corso…")
	}
	if len(checking) > 0 {
		parts = append(parts, "Un dispositivo risponde su "+strings.Join(checking, ", ")+": controllo in corso…")
	}
	if len(listening) > 0 {
		parts = append(parts, "In ascolto su "+strings.Join(listening, ", ")+" (collega la radio e accendila, anche dopo il cavo)")
	}
	if len(unknown) > 0 {
		parts = append(parts, "Su "+strings.Join(unknown, ", ")+" un dispositivo risponde ma non si lascia leggere: se la radio è spenta accendila, se è accesa spegnila e riaccendila (la ricerca riprende da sola); se non viene riconosciuta nemmeno così, il modello non è tra quelli gestiti")
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
	a.auto.mu.Lock()
	// 4.36: dopo una risposta non riconosciuta, fino al prossimo giro completo
	// si riprovano solo le radio già riconosciute
	knownOnly := p.unknown && time.Now().Before(p.fullAt) && a.hasKnownRadios(p.name)
	a.auto.mu.Unlock()
	j, err := a.startIdentify(&identReq{Port: p.name, KnownOnly: knownOnly}, true)
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
	r := &autoRun{p: p, job: j, knownOnly: knownOnly}
	p.run = r
	go a.autoWait(r)
	a.autoStatusLocked()
}

// autoTabOpen (4.36): l'interfaccia ha aperta, con dei dati, la scheda di questa
// radio (l'ha detto negli ultimi secondi: l'interfaccia chiede lo stato ogni
// secondo e mezzo)
func (a *App) autoTabOpen(kind string) bool {
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	return a.auto.openTabs[kind] && time.Since(a.auto.openAt) < 10*time.Second
}

// hasKnownRadios: ci sono radio già riconosciute da riprovare su questa porta
func (a *App) hasKnownRadios(port string) bool {
	st := a.store.Get()
	return len(st.PortRadios[probeKey(port)]) > 0 || len(st.KnownRadios) > 0
}

// autoUnknownNext: prossima ricerca su una porta con un dispositivo che
// risponde ma non si lascia leggere
func (a *App) autoUnknownNext(p *autoPort, now time.Time) time.Time {
	if t := now.Add(autoRetryKnown); t.Before(p.fullAt) && a.hasKnownRadios(p.name) {
		return t
	}
	if p.fullAt.After(now) {
		return p.fullAt
	}
	return now.Add(autoRetryKnown)
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
		// i disturbi di un cavo senza radio sembravano una radio sconosciuta.
		// 4.36: l'interfaccia lo dice una volta (autoNotice), poi si riprova in
		// silenzio: spesso le radio già riconosciute, ogni tanto tutto il giro
		if !p.unknown {
			a.autoNoticeLocked(p, now)
		} else if !r.knownOnly {
			p.fullAt = now.Add(autoRetryUnknown)
		}
		p.next = a.autoUnknownNext(p, now)
	case p.unknown && r.knownOnly:
		// giro delle sole radio note senza risposta: il dispositivo può essere
		// un modello non gestito, che a quelle non risponde. Si resta in silenzio
		// fino al prossimo giro completo
		p.next = a.autoUnknownNext(p, now)
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
		p.fullAt = time.Time{}
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
		// Open (4.36): schede delle radio aperte nell'interfaccia (vedi autoTabOpen)
		Open *[]string `json:"open"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Open != nil {
		a.auto.mu.Lock()
		a.auto.openTabs = map[string]bool{}
		for _, k := range *req.Open {
			a.auto.openTabs[k] = true
		}
		a.auto.openAt = time.Now()
		a.auto.mu.Unlock()
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
	reading := []map[string]interface{}{}
	paused := false
	for _, p := range all {
		ports = append(ports, p.name)
		paused = paused || p.paused
		if p.run != nil {
			searchPorts = append(searchPorts, p.name)
			// 4.36: riquadro visibile solo mentre la radio che ha risposto si fa leggere
			if j := p.run.job; j.answered.Load() && j.reading.Load() {
				answering = append(answering, p.name)
				name, pct := j.readState()
				reading = append(reading, map[string]interface{}{"port": p.name, "name": name, "pct": pct})
			}
		}
	}
	events := a.auto.events
	if events == nil {
		events = []*autoEvent{}
	}
	out := map[string]interface{}{"enabled": a.autoEnabled(), "status": a.auto.status, "ports": ports,
		"searching": len(answering) > 0, "answering": answering, "busy": len(searchPorts) > 0, "searchPorts": searchPorts,
		"seq": a.auto.seq, "paused": paused, "events": events, "reading": reading, "noticeSeq": a.auto.noticeSeq}
	if a.auto.notice != nil {
		out["notice"] = a.auto.notice
	}
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

// identify/diag: salva la diagnosi in un file di testo scelto dall'utente
func (a *App) apiIdentifyDiag(r *http.Request) (interface{}, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	out := []byte("\xEF\xBB\xBF" + a.identDiagReport()) // con BOM: il Blocco note lo apre in UTF-8
	p := req.Path
	if p == "" {
		suggest := "IU0VYT_diagnosi_identifica_" + time.Now().Format("2006-01-02_15-04-05") + ".txt"
		if d := a.store.Get().LastDir; d != "" {
			suggest = filepath.Join(d, suggest)
		}
		var err error
		if p, err = a.dialog(true, "Salva la diagnosi", "File di testo (*.txt)|*.txt", suggest, "txt"); err != nil {
			return nil, err
		}
	}
	if err := writeAtomic(p, out); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	return map[string]interface{}{"path": p, "bytes": len(out), "lines": identDiag.count()}, nil
}
