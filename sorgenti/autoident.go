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
// Con l'opzione attiva il programma, in qualsiasi scheda ci si trovi:
//   - controlla ogni secondo le porte seriali e i dispositivi USB delle radio
//     DMR, anche mentre una ricerca è in corso (4.1: prima il controllo si
//     fermava durante la ricerca e un cavo scollegato e ricollegato in quel
//     momento non veniva notato);
//   - quando compare una porta nuova (cavo di programmazione appena collegato)
//     o un dispositivo USB di una radio, avvia da solo la stessa ricerca della
//     scheda «Identifica» (solo comandi di lettura, nessuna scrittura);
//   - se la radio non risponde (ad esempio è ancora spenta) la ricerca si ripete
//     finché il cavo resta collegato: prima ogni pochi secondi, poi più di rado;
//   - quando la radio è riconosciuta l'interfaccia ne mostra i dati e apre la sua
//     scheda con i canali letti. Da quel momento (4.1) il programma controlla ogni
//     pochi secondi che la radio sia ancora collegata, con il solo saluto iniziale
//     del suo protocollo (lo stesso con cui comincia la ricerca): se la radio non
//     risponde più (cavo staccato dalla radio, radio spenta o sostituita) la
//     ricerca riparte da sola e riconosce la radio collegata dopo, anche se è la
//     stessa. Prima la porta restava ferma finché il cavo non veniva staccato dal
//     PC: una radio scollegata e ricollegata con il cavo ancora nel PC non veniva
//     più cercata.
//
// Si interrogano solo le porte dei cavi di programmazione (adattatori USB-seriale
// CH340, Prolific, Silicon Labs, FTDI…), mai porte Bluetooth, modem o porte
// seriali di sistema. Le porte con il nome di un adattatore noto hanno la
// precedenza; un cavo appena collegato (o una radio appena scollegata) interrompe
// una ricerca ripetuta su un'altra porta. Una ricerca automatica lascia sempre il
// posto alle operazioni chieste dall'utente (yieldAuto): lettura, scrittura,
// firmware, inoltro e ricerca manuale la interrompono subito.

import (
	"context"
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
	// controllo della presenza della radio riconosciuta
	autoPresenceEvery = 4 * time.Second
	autoPresenceWait  = 4 * time.Second // attesa massima della risposta al saluto
	autoPresenceMiss  = 2               // controlli senza risposta prima di dire «scollegata»
	autoQuiet         = 1500 * time.Millisecond
	// usbRadioPresentFunc: c'è un dispositivo USB di una radio DMR conosciuta
	// (impostata dalla versione Windows)
	usbRadioPresentFunc = func() bool { return false }
)

type autoPort struct {
	name, desc string
	strong     bool // adattatore USB-seriale riconosciuto dal nome
	next       time.Time
	tries      int
	done       bool   // radio riconosciuta: si controlla solo la presenza
	kind       string // radio riconosciuta (per il controllo della presenza)
	miss       int
	checkNext  time.Time
	lostKind   string    // radio scollegata (controllo della presenza)
	lostAt     time.Time //
	falseLoss  int       // radio «scollegata» e ritrovata uguale dopo pochi secondi
	gone       bool      // porta sparita (cavo staccato dal PC)
	paused     bool      // ricerca interrotta dall'utente («Interrompi»): fino a «Riprendi» o al ricollegamento del cavo
}

// autoEvent: esito dell'ultima ricerca automatica che ha trovato la radio
type autoEvent struct {
	Seq    int          `json:"seq"`
	Port   string       `json:"port"`
	Result *identResult `json:"result"`
	When   string       `json:"when"`
	Again  bool         `json:"again"` // stessa radio ritrovata dopo uno scollegamento
}

// autoLost: ultima radio data per scollegata
type autoLost struct {
	Seq  int    `json:"seq"`
	Port string `json:"port"`
	Name string `json:"name"`
	When string `json:"when"`
}

type autoRun struct {
	p        *autoPort
	job      *cableJob
	presence bool
}

type autoIdent struct {
	mu         sync.Mutex
	ports      map[string]*autoPort
	usbWas     bool
	usb        *autoPort
	seq        int
	last       *autoEvent
	ident      *identSession // radio dell'ultimo esito (per identify/open con "seq")
	lostSeq    int
	lost       *autoLost
	status     string
	run        *autoRun
	quietUntil time.Time // pausa dopo un saluto interrotto (la radio torna pronta)
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

// autoStep: un controllo di porte e dispositivi; avvia al massimo una ricerca o
// un controllo della presenza, senza aspettarne la fine
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

	var all []*autoPort
	for _, p := range a.auto.ports {
		all = append(all, p)
	}
	if a.auto.usb != nil {
		all = append(all, a.auto.usb)
	}
	// ricerche da fare: prima le porte appena collegate (o con la radio appena
	// scollegata), poi gli adattatori noti, poi la più vecchia
	var cands []*autoPort
	for _, p := range all {
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
		// porta staccata dal PC durante la ricerca, oppure una porta appena
		// collegata mentre si ripete la ricerca su un'altra: si interrompe
		if r.p.gone || (due != nil && due != r.p && due.tries == 0 && (r.presence || r.p.tries > 0)) {
			if r.job.cancel != nil {
				r.job.cancel()
			}
		}
		a.autoStatusLocked(cands, due)
		return
	}
	if userBusy {
		return
	}
	if due != nil {
		a.autoLaunchLocked(due, false)
		a.autoStatusLocked(cands, due)
		return
	}
	// controllo della presenza delle radio riconosciute
	for _, p := range all {
		if p.done && p.kind != "" && p.name != "USB" && !p.checkNext.After(now) {
			a.autoLaunchLocked(p, true)
			break
		}
	}
	a.autoStatusLocked(cands, nil)
}

func (a *App) autoStatusLocked(cands []*autoPort, due *autoPort) {
	if r := a.auto.run; r != nil && !r.presence {
		a.auto.status = "Ricerca della radio su " + r.p.name + "…"
		return
	}
	var done []string
	for _, p := range a.auto.ports {
		if p.done {
			done = append(done, p.name)
		}
	}
	if a.auto.usb != nil && a.auto.usb.done {
		done = append(done, "USB")
	}
	switch {
	case len(cands) == 0 && a.autoPausedLocked():
		a.auto.status = "Ricerca interrotta: premi «Riprendi la ricerca» o ricollega il cavo"
	case len(cands) == 0 && len(done) == 0:
		a.auto.status = "In attesa del collegamento di una radio"
	case len(cands) == 0:
		sort.Strings(done)
		a.auto.status = "Radio collegata già riconosciuta (" + strings.Join(done, ", ") + "): se la scolleghi o la cambi, la ricerca riparte da sola"
	case due == nil:
		a.auto.status = "In attesa che la radio risponda (" + cands[0].name + "): accendila con il volume a metà"
	}
}

func (a *App) autoPausedLocked() bool {
	for _, p := range a.auto.ports {
		if p.paused {
			return true
		}
	}
	return a.auto.usb != nil && a.auto.usb.paused
}

// autoLaunchLocked: avvia la ricerca (o il controllo della presenza) su p in un
// goroutine; con a.auto.mu già preso
func (a *App) autoLaunchLocked(p *autoPort, presence bool) {
	var j *cableJob
	var err error
	if presence {
		j, err = a.startPresence(p.name, p.kind)
	} else {
		j, err = a.startIdentify(&identReq{Port: p.name}, true)
	}
	if err != nil {
		return // un'altra operazione è appena partita: si riprova al prossimo controllo
	}
	r := &autoRun{p: p, job: j, presence: presence}
	a.auto.run = r
	if !presence {
		a.auto.status = "Ricerca della radio su " + p.name + "…"
	}
	go a.autoWait(r)
}

// autoWait: attende la fine della ricerca o del controllo e ne registra l'esito
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
	p := r.p
	if r.presence {
		present, _ := res.(bool)
		switch {
		case state == "cancelled":
			p.checkNext = now.Add(autoPresenceEvery)
		case state == "done" && present:
			p.miss = 0
			p.checkNext = now.Add(autoPresenceEvery)
			if p.falseLoss >= 3 {
				p.checkNext = now.Add(8 * autoPresenceEvery)
			}
			a.auto.quietUntil = now.Add(autoQuiet)
		default:
			p.miss++
			p.checkNext = now.Add(autoPresenceEvery / 2)
			if p.miss >= autoPresenceMiss {
				// la radio non risponde più: scollegata, spenta o sostituita
				a.auto.lostSeq++
				name := p.kind
				if a.auto.last != nil && a.auto.last.Port == p.name && a.auto.last.Result != nil {
					name = a.auto.last.Result.Name
				}
				a.auto.lost = &autoLost{Seq: a.auto.lostSeq, Port: p.name, Name: name, When: now.Format("15:04:05")}
				p.lostKind, p.lostAt = p.kind, now
				p.done, p.kind, p.miss, p.tries, p.next = false, "", 0, 0, now
				a.auto.status = "Radio scollegata da " + p.name + ": in attesa di una radio"
			}
		}
		return
	}
	if state == "cancelled" {
		a.auto.quietUntil = now.Add(autoQuiet)
	}
	if p.gone {
		return // cavo staccato durante la ricerca: la porta ricomparirà come nuova
	}
	p.tries++
	ir, _ := res.(*identResult)
	switch {
	case state == "done" && ir != nil && ir.Found:
		again := p.lostKind != "" && p.lostKind == ir.Kind
		if again && now.Sub(p.lostAt) < 2*autoPresenceEvery {
			// «scollegata» e ritrovata uguale dopo pochi secondi: forse la radio
			// non aveva risposto al saluto per un attimo; dalla terza volta il
			// controllo della presenza si fa più di rado (mai disattivato)
			p.falseLoss++
		}
		p.done, p.kind, p.miss = true, ir.Kind, 0
		p.checkNext = now.Add(autoPresenceEvery)
		p.lostKind = ""
		a.auto.seq++
		a.auto.last = &autoEvent{Seq: a.auto.seq, Port: p.name, Result: ir, When: now.Format("15:04:05"), Again: again}
		a.auto.ident = id
		a.auto.status = "Radio riconosciuta: " + ir.Name + " (" + p.name + ")"
	case state == "cancelled":
		// interrotta dall'utente o da un'altra operazione: si riprova dopo
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

// yieldAuto: interrompe la ricerca automatica (o il controllo della presenza)
// in corso, per lasciare il cavo all'operazione chiesta dall'utente; attende al
// massimo pochi secondi, più la breve pausa che serve alla radio per tornare
// pronta dopo un saluto interrotto
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

// ---------------------------------------------------------------- presenza

// startPresence: controllo della presenza della radio kind sulla porta (lavoro
// automatico: lascia il posto a qualsiasi operazione dell'utente)
func (a *App) startPresence(port, kind string) (*cableJob, error) {
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: "presenza", Op: "presence", State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}, Auto: true}
	a.mu.Lock()
	if a.job != nil && a.job.isRunning() {
		a.mu.Unlock()
		cancel()
		return nil, errors.New("è già in corso un'operazione con la radio")
	}
	a.job = j
	a.mu.Unlock()
	go func() {
		defer cancel()
		present := false
		func() {
			defer func() { _ = recover() }()
			present = a.runPresence(ctx, port, kind)
		}()
		j.mu.Lock()
		defer j.mu.Unlock()
		if ctx.Err() != nil {
			j.State = "cancelled"
		} else {
			j.State, j.Result = "done", present
		}
	}()
	return j, nil
}

// runPresence: la radio kind risponde al saluto iniziale del suo protocollo?
// Si invia solo l'inizio della lettura (lo stesso della ricerca) e ci si ferma
// alla prima risposta, senza leggere la memoria.
func (a *App) runPresence(ctx context.Context, port, kind string) bool {
	identPortMu.Lock()
	defer identPortMu.Unlock()
	plan := a.identPlan(&identReq{Port: port})
	tried := false
	for _, pr := range plan {
		var fn func(ctx context.Context) error
		switch {
		case pr.k67 && kind == kindXLC, pr.bt && kind == kindDAT:
			creq := &cableReq{Op: "read", Radio: kindXLC, Port: port, Baud: pr.baud, Password: a.store.Get().K67Pwd}
			if pr.bt {
				creq.Radio, creq.Baud, creq.Password = kindDAT, 0, ""
			}
			fn = func(c context.Context) error {
				_, _, _, e := a.cableRead(c, &cableJob{Notes: []string{}, Log: []string{}}, creq)
				return e
			}
		default:
			for _, m := range pr.kinds {
				if m.ID == kind {
					mm := m
					fn = func(c context.Context) error {
						_, _, e := mm.Fam.Download(c, mm, port, func(int, int, string) {})
						return e
					}
				}
			}
		}
		if fn == nil {
			continue
		}
		tried = true
		if presenceProbe(ctx, fn) {
			return true
		}
		if ctx.Err() != nil {
			return false
		}
	}
	return !tried // radio non verificabile: si considera ancora collegata
}

// presenceProbe: esegue fn con la porta controllata e la interrompe appena la
// radio risponde (true) o dopo autoPresenceWait senza risposta (false)
func presenceProbe(ctx context.Context, fn func(ctx context.Context) error) bool {
	st := &probeState{}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		p, err := old(name, baud)
		if err != nil {
			return nil, err
		}
		return probePort{p, st}, nil
	}
	defer func() { openPortFunc = old }()
	pctx, cancel := context.WithTimeout(ctx, autoPresenceWait)
	defer cancel()
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(20 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if st.answered() {
					cancel()
					return
				}
			}
		}
	}()
	func() {
		defer func() { _ = recover() }()
		_ = fn(pctx)
	}()
	close(done)
	return st.answered()
}

// ---------------------------------------------------------------- API

// identify/auto: stato (e, con "enabled", attivazione o disattivazione)
func (a *App) apiIdentifyAuto(r *http.Request) (interface{}, error) {
	var req struct {
		Enabled *bool `json:"enabled"`
		Stop    bool  `json:"stop"`   // interrompe la ricerca in corso (banner «Interrompi»)
		Resume  bool  `json:"resume"` // riprende le ricerche interrotte
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Stop {
		// tutte le ricerche in attesa (le radio già riconosciute restano controllate)
		a.auto.mu.Lock()
		for _, p := range a.auto.ports {
			if !p.done {
				p.paused = true
			}
		}
		if a.auto.usb != nil && !a.auto.usb.done {
			a.auto.usb.paused = true
		}
		a.auto.mu.Unlock()
		a.yieldAuto()
	}
	if req.Resume {
		a.auto.mu.Lock()
		for _, p := range a.auto.ports {
			p.paused = false
		}
		if a.auto.usb != nil {
			a.auto.usb.paused = false
		}
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
	a.auto.mu.Lock()
	defer a.auto.mu.Unlock()
	ports := []string{}
	for n := range a.auto.ports {
		ports = append(ports, n)
	}
	if a.auto.usb != nil {
		ports = append(ports, "USB")
	}
	sort.Strings(ports)
	searching, port := false, ""
	if r := a.auto.run; r != nil && !r.presence {
		searching, port = true, r.p.name
	}
	out := map[string]interface{}{"enabled": a.autoEnabled(), "status": a.auto.status, "ports": ports, "searching": searching,
		"searchPort": port, "seq": a.auto.seq, "lostSeq": a.auto.lostSeq, "paused": a.autoPausedLocked()}
	if a.auto.last != nil {
		out["last"] = a.auto.last
	}
	if a.auto.lost != nil {
		out["lost"] = a.auto.lost
	}
	return out, nil
}
