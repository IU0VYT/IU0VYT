//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func identDiagReset() {
	identDiag.mu.Lock()
	identDiag.lines = nil
	identDiag.mu.Unlock()
}

func diagCount(sub string) int {
	n := 0
	for _, l := range strings.Split(identDiag.text(), "\r\n") {
		if strings.Contains(l, sub) {
			n++
		}
	}
	return n
}

// 4.36: il programma ricorda su quale porta ha riconosciuto ogni radio. Tre
// radio, due cavi (come nella prova di Daniele: K67 Pro e BJ-8300 sullo stesso
// cavo, BJ-9000 su un altro): su ogni porta si provano per prime le radio già
// riconosciute lì, l'ultima per prima; prima l'elenco era uno solo e la radio
// collegata poteva essere provata per terza.
func TestIdentPortMemory(t *testing.T) {
	a := newApp(mustStore(t))
	first := func(port string, n int) []string {
		identPortMu.Lock()
		defer identPortMu.Unlock()
		var out []string
		for _, p := range a.identPlan(&identReq{Port: port})[:n] {
			switch {
			case p.k67:
				out = append(out, kindXLC)
			case p.bt:
				out = append(out, kindDAT)
			default:
				out = append(out, p.kinds[0].ID)
			}
		}
		return out
	}
	// impostazioni di una versione precedente: un solo elenco, BJ-9000 l'ultima
	st := a.store.Get()
	st.KnownRadios = []string{"bj-9000", kindDAT, kindXLC}
	if err := a.store.SaveSettings(st); err != nil {
		t.Fatal(err)
	}
	if got := first("COM6", 4); got[0] != "bj-9000" || got[1] != kindDAT || got[2] != kindXLC {
		t.Fatalf("porta senza ricordi: %v", got)
	}
	a.rememberRadio("COM6", kindDAT)
	a.rememberRadio("COM6", kindXLC)
	a.rememberRadio("com9", "bj-9000")
	st = a.store.Get()
	if strings.Join(st.KnownRadios, ",") != "bj-9000,"+kindXLC+","+kindDAT {
		t.Fatalf("elenco generale: %v", st.KnownRadios)
	}
	if strings.Join(st.PortRadios["COM6"], ",") != kindXLC+","+kindDAT || strings.Join(st.PortRadios["COM9"], ",") != "bj-9000" {
		t.Fatalf("radio per porta: %v", st.PortRadios)
	}
	// COM6: K67 Pro (due velocità), BJ-8300, poi la BJ-9000 dell'altro cavo
	if got := first("COM6", 4); strings.Join(got, ",") != kindXLC+","+kindXLC+","+kindDAT+",bj-9000" {
		t.Fatalf("COM6: %v", got)
	}
	// COM9: BJ-9000 per prima, poi le altre note
	if got := first("COM9", 4); strings.Join(got, ",") != "bj-9000,"+kindXLC+","+kindXLC+","+kindDAT {
		t.Fatalf("COM9: %v", got)
	}
	// l'ultima riconosciuta sulla porta passa in testa
	a.rememberRadio("COM6", kindDAT)
	if got := first("COM6", 3); strings.Join(got, ",") != kindDAT+","+kindXLC+","+kindXLC {
		t.Fatalf("COM6 dopo la BJ-8300: %v", got)
	}
	// le impostazioni salvate si rileggono
	s2, err := openStore(a.store.dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(s2.Get().PortRadios["COM6"], ",") != kindDAT+","+kindXLC {
		t.Fatalf("radio per porta non salvate: %v", s2.Get().PortRadios)
	}
	// ricerca automatica: sulla porta si ripetono solo le sue radio (della K67
	// Pro una sola velocità), ogni identKnownEveryPort tentativi degli altri
	identPortMu.Lock()
	plan := a.identPlan(&identReq{Port: "COM6"})
	identPortMu.Unlock()
	idx, every := a.knownProbes(plan, "COM6")
	if len(idx) != 2 || idx[0] != 0 || idx[1] != 1 || every != identKnownEveryPort {
		t.Fatalf("tentativi delle radio note su COM6: %v ogni %d", idx, every)
	}
	il := identInterleaveKnown(plan, idx, every)
	if want := len(plan) + 2*((len(plan)-2-1)/every); len(il) != want {
		t.Fatalf("piano con le radio note ripetute: %d tentativi, attesi %d", len(il), want)
	}
	for i, p := range il {
		// posizioni 0-1, poi ogni «every» altri: 2+every, 2+every+1, …
		k := i % (every + 2)
		if want := k < 2; (p.bt || (p.k67 && p.baud == plan[1].baud)) != want {
			t.Fatalf("posizione %d del piano: %s", i, p.name)
		}
	}
	// porta mai usata: le radio note in generale, ogni identKnownEvery
	idx, every = a.knownProbes(a.identPlan(&identReq{Port: "COM33"}), "COM33")
	if len(idx) != 3 || every != identKnownEvery {
		t.Fatalf("porta nuova: %v ogni %d", idx, every)
	}
	// troppe porte ricordate: restano quelle presenti e quella appena usata
	for i := 0; i < identPortsMax+10; i++ {
		a.rememberRadio("COM"+strings.Repeat("1", 1)+string(rune('A'+i%26))+string(rune('A'+i/26)), "bf-uv5r")
	}
	if n := len(a.store.Get().PortRadios); n > identPortsMax {
		t.Fatalf("porte ricordate: %d", n)
	}
}

// 4.36: la velocità a cui la K67 Pro ha risposto si ricorda anche quando la
// radio è stata letta da «Identifica» (prima solo dalla scheda della radio): la
// ricerca successiva comincia da quella.
func TestIdentK67BaudRemembered(t *testing.T) {
	identFast(t)
	// radio a 115200 baud: alla domanda fatta a 19200 risponde con un byte
	// che non è «A» (il programma passa a 115200, come quello del costruttore)
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "BF-K6", answerA: 0xFE}
	c := newAPIClient(t)
	identAttach(t, func(p Port) { _ = k67.serve(p) }, "")
	c.must("identify/start", map[string]string{"port": "COM6"})
	st := identWait(c)
	if st["state"] != "done" || st["result"].(map[string]interface{})["kind"] != kindXLC {
		t.Fatalf("ricerca: %v", st)
	}
	s := c.app.store.Get()
	if s.K67Baud != 115200 {
		t.Fatalf("velocità ricordata: %d", s.K67Baud)
	}
	if p := s.PortRadios["COM6"]; len(p) != 1 || p[0] != kindXLC {
		t.Fatalf("radio della porta: %v", s.PortRadios)
	}
	identPortMu.Lock()
	plan := c.app.identPlan(&identReq{Port: "COM6"})
	identPortMu.Unlock()
	if !plan[0].k67 || plan[0].baud != 115200 || !plan[1].k67 || plan[1].baud != 19200 {
		t.Fatalf("piano: %s, %s", plan[0].name, plan[1].name)
	}
}

// 4.36: l'ascolto della porta prima dei tentativi (un secondo, per scoprire le
// radio che trasmettono da sole) si faceva per ogni velocità: ora una porta che
// tace si ascolta una volta sola, e il tentativo di una radio già riconosciuta
// su quella porta parte senza ascolto.
func TestIdentListenOnce(t *testing.T) {
	f, s, p, l, l9 := identFirstWait, identSilentMax, identPause, identSpontListen, identSpontListen9600
	identFirstWait, identSilentMax, identPause = 20*time.Millisecond, 40*time.Millisecond, 10*time.Millisecond
	identSpontListen, identSpontListen9600 = 150*time.Millisecond, 150*time.Millisecond
	identSpontReset()
	t.Cleanup(func() {
		identFirstWait, identSilentMax, identPause, identSpontListen, identSpontListen9600 = f, s, p, l, l9
		identSpontReset()
	})
	identAttach(t, func(p Port) {}, "") // cavo senza radio
	run := func(c *apiClient) (listens, noListen int) {
		identDiagReset()
		identSpontReset()
		c.must("identify/start", map[string]string{"port": "COM6", "brand": "Baofeng"})
		if st := identWait(c); st["state"] != "done" || st["result"].(map[string]interface{})["found"] == true {
			t.Fatalf("ricerca senza radio: %v", st)
		}
		return diagCount("ascolto di"), diagCount("senza ascolto")
	}
	c := newAPIClient(t)
	if n, k := run(c); n != 1 || k != 0 {
		t.Fatalf("porta senza ricordi: %d ascolti, %d tentativi senza ascolto\n%s", n, k, identDiag.text())
	}
	if !strings.Contains(strings.Split(identDiag.text(), "\r\n")[2], "ascolto di") {
		t.Fatalf("l'ascolto non è al primo tentativo:\n%s", identDiag.text()[:600])
	}
	// K67 Pro già riconosciuta su questa porta: il suo tentativo parte subito,
	// l'ascolto si fa al tentativo seguente
	c.app.rememberRadio("COM6", kindXLC)
	if n, k := run(c); n != 1 || k != 1 {
		t.Fatalf("radio nota: %d ascolti, %d tentativi senza ascolto", n, k)
	}
	lines := strings.Split(identDiag.text(), "\r\n")
	if !strings.Contains(lines[1], "K67") || !strings.Contains(lines[2], "senza ascolto") {
		t.Fatalf("primo tentativo:\n%s", strings.Join(lines[:4], "\n"))
	}
}

// k67Mezza: K67 Pro «spenta ma collegata». Risponde al saluto (anche con il
// nome del modello) ma non alla richiesta di lettura: è quello che può fare
// una radio spenta che resta alimentata dal cavo. Conta i saluti ricevuti.
func k67Mezza(p Port, n *int32) {
	t, err := emuRead(p, 16)
	if err != nil || !bytes.Equal(t, k67TInfo) {
		return
	}
	atomic.AddInt32(n, 1)
	p.Write([]byte{'A'})
	sync, err := emuRead(p, 8)
	if err != nil {
		return
	}
	seed := sync[7]
	p.Write([]byte{'A' ^ seed})
	if _, err := emuRead(p, 8); err != nil { // password
		return
	}
	p.Write([]byte{'A' ^ seed})
	if _, err := emuRead(p, 11); err != nil { // INFORMATION
		return
	}
	id := bytes.Repeat([]byte{0xFF}, 16)
	copy(id, "BF-K6")
	p.Write(xorBytes(id, seed))
	_, _ = emuRead(p, 1<<20) // alla richiesta di lettura non risponde: tace fino alla chiusura della porta
}

// lentoBlocchi: lato radio di una porta simulata; ogni blocco di dati parte dopo
// una pausa, come su un cavo vero a 19200 baud (la lettura dura abbastanza da
// vedere il riquadro)
type lentoBlocchi struct {
	Port
	d time.Duration
}

func (p *lentoBlocchi) Write(b []byte) error {
	if len(b) > 1000 {
		time.Sleep(p.d)
	}
	return p.Port.Write(b)
}

// 4.36, segnalazione di Daniele: con la K67 Pro collegata ma SPENTA, dopo che
// era già stata identificata, compariva il riquadro «Identificazione Radio In
// Corso». Una radio spenta che risponde al saluto senza farsi leggere non deve
// far comparire il riquadro: il programma avvisa una volta che il dispositivo
// risponde ma non si lascia leggere, riprova in silenzio le radio già
// riconosciute e, quando la radio viene accesa, la riconosce in pochi secondi
// (qui il riquadro compare, con il modello e l'avanzamento).
func TestAutoIdentRadioOff(t *testing.T) {
	identFast(t)
	identFirstWait, identSilentMax = 80*time.Millisecond, 250*time.Millisecond
	oldS, oldF, oldU, oldK, oldL, oldQ := autoSettle, autoRetryFast, autoRetryUnknown, autoRetryKnown, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	autoRetryUnknown, autoRetryKnown = 4*time.Second, 150*time.Millisecond
	t.Cleanup(func() {
		autoSettle, autoRetryFast, autoRetryUnknown, autoRetryKnown, listPortsFunc, autoQuiet = oldS, oldF, oldU, oldK, oldL, oldQ
	})
	identDiagReset()
	listPortsFunc = func() []PortInfo { return []PortInfo{{"COM6", "USB-SERIAL CH340 (COM6)"}} }
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "BF-K6"}
	var mu sync.Mutex
	mode := "mezza"
	var hello int32
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	openPortFunc = func(name string, baud int) (Port, error) {
		x, y := newPipe()
		x.baud = baud
		mu.Lock()
		m := mode
		mu.Unlock()
		switch m {
		case "mezza":
			go k67Mezza(y, &hello)
		case "accesa":
			go func() { _ = k67.serve(&lentoBlocchi{Port: y, d: 60 * time.Millisecond}) }()
		}
		return x, nil
	}
	c := newAPIClient(t)
	a := c.app
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })
	a.rememberRadio("COM6", kindXLC) // la radio era già stata identificata su questo cavo

	banner, notices, lastNotice := 0, 0, 0
	var status string
	var reading []interface{}
	sample := func() map[string]interface{} {
		a.autoStep(time.Now())
		r := c.must("identify/auto", nil)
		if r["searching"] == true {
			banner++
			if rd, _ := r["reading"].([]interface{}); len(rd) > 0 {
				reading = rd
			}
		}
		if n, ok := r["notice"].(map[string]interface{}); ok && int(n["seq"].(float64)) > lastNotice {
			lastNotice = int(n["seq"].(float64))
			notices++
			if n["port"] != "COM6" {
				t.Fatalf("avviso: %v", n)
			}
		}
		status, _ = r["status"].(string)
		return r
	}
	c.must("identify/auto", map[string]interface{}{"enabled": true})

	// radio spenta che risponde al saluto: per tre secondi
	end := time.Now().Add(3 * time.Second)
	firstNotice := time.Duration(0)
	t0 := time.Now()
	for time.Now().Before(end) {
		sample()
		if notices == 1 && firstNotice == 0 {
			firstNotice = time.Since(t0)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if banner != 0 {
		t.Fatalf("riquadro «Identificazione Radio In Corso» visibile %d volte con la radio spenta", banner)
	}
	if notices != 1 {
		t.Fatalf("avvisi «risponde ma non si lascia leggere»: %d (atteso 1)\n%s", notices, identDiag.text())
	}
	// l'avviso arriva appena la radio nota ha risposto senza farsi leggere, non
	// alla fine del giro (che con i tempi veri dura minuti)
	if firstNotice > 2*time.Second {
		t.Fatalf("avviso dopo %v", firstNotice)
	}
	if !strings.Contains(status, "non si lascia leggere") || !strings.Contains(status, "COM6") {
		t.Fatalf("stato: %q", status)
	}
	// il giro continua e la radio nota viene riprovata ogni pochi tentativi
	if h := atomic.LoadInt32(&hello); h < 2 {
		t.Fatalf("la radio nota è stata riprovata solo %d volte in 3 s", h)
	}
	t.Logf("radio spenta: avviso dopo %v, %d saluti in 3 s, nessun riquadro", firstNotice.Round(10*time.Millisecond), atomic.LoadInt32(&hello))

	// la radio viene accesa: riconosciuta in poco tempo, con il riquadro
	mu.Lock()
	mode = "accesa"
	mu.Unlock()
	t0 = time.Now()
	for {
		r := sample()
		if int(r["seq"].(float64)) == 1 {
			break
		}
		if time.Since(t0) > 15*time.Second {
			t.Fatalf("radio accesa non riconosciuta (stato %q)\n%s", status, identDiag.text())
		}
		time.Sleep(20 * time.Millisecond)
	}
	d := time.Since(t0)
	t.Logf("radio accesa dopo l'avviso riconosciuta in %v; riquadro visibile in %d campioni; %v", d.Round(10*time.Millisecond), banner, reading)
	if d > 4*time.Second {
		t.Errorf("radio accesa riconosciuta dopo %v", d)
	}
	if banner == 0 || len(reading) == 0 {
		t.Fatal("riquadro mai visibile durante la lettura della radio accesa")
	}
	rd := reading[0].(map[string]interface{})
	if rd["port"] != "COM6" || rd["name"] != "Baofeng K67 Pro" || rd["pct"].(float64) <= 0 {
		t.Fatalf("riquadro: %v", rd)
	}
	if notices != 1 {
		t.Fatalf("avvisi dopo l'accensione: %d", notices)
	}
	r := sample()
	if r["searching"] == true || strings.Contains(r["status"].(string), "non si lascia leggere") {
		t.Fatalf("stato dopo il riconoscimento: %v", r["status"])
	}
	ev := r["last"].(map[string]interface{})["result"].(map[string]interface{})
	if ev["kind"] != kindXLC {
		t.Fatalf("radio riconosciuta: %v", ev)
	}
}

// 4.36: la diagnosi dell'identificazione riporta tentativi, byte e tempi, ma
// non la password di programmazione della K67 Pro.
func TestIdentDiag(t *testing.T) {
	identFast(t)
	identDiagReset()
	const pwd = "SEGRETO1"
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "BF-K6", password: pwd}
	c := newAPIClient(t)
	st := c.app.store.Get()
	st.K67Pwd = pwd
	if err := c.app.store.SaveSettings(st); err != nil {
		t.Fatal(err)
	}
	identAttach(t, func(p Port) { _ = k67.serve(p) }, "")
	c.must("identify/start", map[string]string{"port": "COM6"})
	if st := identWait(c); st["state"] != "done" || st["result"].(map[string]interface{})["kind"] != kindXLC {
		t.Fatalf("ricerca: %v", st)
	}
	path := filepath.Join(t.TempDir(), "diagnosi.txt")
	r := c.must("identify/diag", map[string]string{"path": path})
	b, err := os.ReadFile(path)
	if err != nil || r["path"] != path {
		t.Fatalf("diagnosi non salvata: %v %v", err, r)
	}
	text := string(b)
	if !strings.HasPrefix(text, "\xEF\xBB\xBFIU0VYT Flash "+appVersion) {
		t.Fatalf("intestazione: %q", text[:60])
	}
	for _, want := range []string{"=== ricerca manuale", "tentativo 1 di", "Baofeng K67 Pro", "apertura a 19200 baud", "→ 16 byte: " + hex.EncodeToString(k67TInfo),
		"← 1 byte dopo", "lettura riuscita; risposta sicura", "riconosciuta Baofeng", "Radio già riconosciute su COM6: " + kindXLC} {
		if !strings.Contains(text, want) {
			t.Errorf("nella diagnosi manca %q", want)
		}
	}
	if t.Failed() {
		t.Logf("\n%s", text)
	}
	// il seme è l'ultimo byte del saluto di 8 byte: la password cifrata con
	// quel seme non deve comparire, e nemmeno in chiaro
	i := strings.Index(text, "→ 8 byte: ")
	if i < 0 {
		t.Fatal("saluto con il seme non trovato")
	}
	sync, err := hex.DecodeString(text[i+len("→ 8 byte: ") : i+len("→ 8 byte: ")+16])
	if err != nil {
		t.Fatal(err)
	}
	enc := hex.EncodeToString(xorBytes([]byte(pwd), sync[7]))
	if strings.Contains(text, enc) || strings.Contains(text, pwd) || strings.Contains(text, hex.EncodeToString([]byte(pwd))) {
		t.Fatal("la password della K67 Pro è finita nella diagnosi")
	}
	if n := strings.Count(text, "→ 8 byte: "); n != 1 {
		t.Fatalf("invii di 8 byte in chiaro: %d", n)
	}
	// le righe sono limitate
	for i := 0; i < identDiagMax+100; i++ {
		identDiag.add("COM6", "riga %d", i)
	}
	if n := identDiag.count(); n > identDiagMax || n < identDiagMax/2 {
		t.Fatalf("righe tenute: %d", n)
	}
	identDiagReset()
}

// 4.36: dopo un giro in cui un dispositivo ha risposto senza essere riconosciuto
// la porta resta in silenzio: un avviso solo, poi le sole radio già note ogni
// autoRetryKnown e tutto il giro ogni autoRetryUnknown (prima: tutto il giro
// ogni 30 secondi, con il riquadro che ricompariva ogni volta).
func TestAutoIdentUnknownRetry(t *testing.T) {
	a := newApp(mustStore(t))
	a.auto.ports = map[string]*autoPort{}
	p := &autoPort{name: "COM6"}
	a.auto.ports["COM6"] = p
	finish := func(answered, knownOnly bool) {
		j := &cableJob{Op: "identify", State: "done", Result: &identResult{}, Auto: true}
		j.answered.Store(answered)
		r := &autoRun{p: p, job: j, knownOnly: knownOnly}
		p.run = r
		a.autoWait(r)
	}
	near := func(got time.Time, want time.Duration) bool {
		d := time.Until(got) - want
		return d > -300*time.Millisecond && d < 300*time.Millisecond
	}
	// nessuna radio nota: dopo la risposta non riconosciuta si aspetta il giro completo
	finish(true, false)
	if !p.unknown || a.auto.noticeSeq != 1 || a.auto.notice.Port != "COM6" || !near(p.next, autoRetryUnknown) {
		t.Fatalf("senza radio note: unknown %v, avvisi %d, prossima fra %v", p.unknown, a.auto.noticeSeq, time.Until(p.next))
	}
	if a.hasKnownRadios("COM6") {
		t.Fatal("radio note inesistenti")
	}
	// con una radio nota: giri brevi delle sole radio note, senza altri avvisi
	a.rememberRadio("COM6", kindXLC)
	p.unknown, p.fullAt = false, time.Time{}
	finish(true, false)
	if !p.unknown || a.auto.noticeSeq != 2 || !near(p.next, autoRetryKnown) || !near(p.fullAt, autoRetryUnknown) {
		t.Fatalf("con radio nota: avvisi %d, prossima fra %v, giro completo fra %v", a.auto.noticeSeq, time.Until(p.next), time.Until(p.fullAt))
	}
	full := p.fullAt
	finish(true, true)  // la radio nota risponde ancora senza farsi leggere
	finish(false, true) // oppure tace: il dispositivo può essere un modello non gestito
	if !p.unknown || a.auto.noticeSeq != 2 || !near(p.next, autoRetryKnown) || !p.fullAt.Equal(full) {
		t.Fatalf("giri delle radio note: unknown %v, avvisi %d, prossima fra %v", p.unknown, a.auto.noticeSeq, time.Until(p.next))
	}
	// vicino al giro completo: la prossima ricerca è il giro completo
	p.fullAt = time.Now().Add(time.Second)
	finish(false, true)
	if !near(p.next, time.Second) {
		t.Fatalf("prossima ricerca fra %v, giro completo fra 1 s", time.Until(p.next))
	}
	// giro completo con risposta: ancora in silenzio, nuovo giro completo fra autoRetryUnknown
	finish(true, false)
	if !p.unknown || a.auto.noticeSeq != 2 || !near(p.fullAt, autoRetryUnknown) {
		t.Fatalf("giro completo: avvisi %d, prossimo fra %v", a.auto.noticeSeq, time.Until(p.fullAt))
	}
	// giro completo senza risposta: il dispositivo non c'è più, ascolto continuo
	finish(false, false)
	if p.unknown || !near(p.next, autoRetryFast) {
		t.Fatalf("dispositivo sparito: unknown %v, prossima fra %v", p.unknown, time.Until(p.next))
	}
	// lo stato dell'interfaccia
	p.unknown = true
	a.auto.mu.Lock()
	a.autoStatusLocked()
	st := a.auto.status
	a.auto.mu.Unlock()
	if !strings.Contains(st, "Su COM6 un dispositivo risponde ma non si lascia leggere") {
		t.Fatalf("stato: %q", st)
	}
}

// 4.36: ricerca automatica con la scheda della K67 Pro già aperta. Per dire
// «Radio già identificata!! Scheda Radio già aperta» non serve rileggere tutta
// la memoria (a 19200 baud sono 26 secondi): basta il controllo breve (saluto,
// primo blocco, chiusura). Con la scheda chiusa la memoria si legge per intero.
func TestAutoIdentAlreadyOpen(t *testing.T) {
	identFast(t)
	identFirstWait, identSilentMax = 80*time.Millisecond, 250*time.Millisecond
	oldS, oldF, oldL, oldQ := autoSettle, autoRetryFast, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { autoSettle, autoRetryFast, listPortsFunc, autoQuiet = oldS, oldF, oldL, oldQ })
	identDiagReset()
	listPortsFunc = func() []PortInfo { return []PortInfo{{"COM6", "USB-SERIAL CH340 (COM6)"}} }
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "BF-K6"}
	identAttach(t, func(p Port) { _ = k67.serve(p) }, "")
	c := newAPIClient(t)
	a := c.app
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })
	wait := func(seq int, open []string) map[string]interface{} {
		t.Helper()
		t0 := time.Now()
		for {
			a.autoStep(time.Now())
			r := c.must("identify/auto", map[string]interface{}{"open": open})
			if int(r["seq"].(float64)) >= seq {
				return r["last"].(map[string]interface{})["result"].(map[string]interface{})
			}
			if time.Since(t0) > 20*time.Second {
				t.Fatalf("esito %d non arrivato (stato %v)\n%s", seq, r["status"], identDiag.text())
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	blocks := xlcRadioLen / 4096
	// prima identificazione: scheda chiusa, memoria letta per intero
	c.must("identify/auto", map[string]interface{}{"enabled": true, "open": []string{}})
	r := wait(1, []string{})
	if r["kind"] != kindXLC || r["already"] == true || k67.reads != blocks || a.autoIdentSession(1) == nil {
		t.Fatalf("prima identificazione: %v, %d blocchi letti", r, k67.reads)
	}
	c.must("identify/open", map[string]interface{}{"kind": kindXLC, "seq": 1})
	// «Identifica Radio» con la scheda aperta: un blocco solo, poi la chiusura
	k67.mu.Lock()
	k67.log = nil
	k67.mu.Unlock()
	t0 := time.Now()
	c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{kindXLC}})
	r = wait(2, []string{kindXLC})
	if r["kind"] != kindXLC || r["already"] != true || k67.reads != blocks+1 {
		t.Fatalf("scheda aperta: %v, %d blocchi letti in più", r, k67.reads-blocks)
	}
	if a.autoIdentSession(2) != nil {
		t.Fatal("scheda aperta: memoria da aprire senza averla letta")
	}
	k67.mu.Lock()
	log := strings.Join(k67.log, " ")
	k67.mu.Unlock()
	if !strings.HasSuffix(log, "R:0000:0 END") || strings.Count(log, "HS:R:") != 1 {
		t.Fatalf("comandi ricevuti dalla radio: %s", log)
	}
	if r["backup"] != "" || r["count"].(float64) != 0 {
		t.Fatalf("scheda aperta: %v", r)
	}
	t.Logf("scheda già aperta: radio riconosciuta in %v con un blocco", time.Since(t0).Round(10*time.Millisecond))
	// scheda chiusa nel frattempo: di nuovo la lettura intera
	c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{}})
	r = wait(3, []string{})
	if r["already"] == true || k67.reads != 2*blocks+1 || a.autoIdentSession(3) == nil {
		t.Fatalf("scheda chiusa: %v, %d blocchi letti", r, k67.reads)
	}
	// la ricerca manuale legge sempre tutto
	c.must("identify/auto", map[string]interface{}{"enabled": false, "open": []string{kindXLC}})
	c.must("identify/start", map[string]string{"port": "COM6"})
	if st := identWait(c); st["state"] != "done" || st["result"].(map[string]interface{})["already"] == true || k67.reads != 3*blocks+1 {
		t.Fatalf("ricerca manuale: %v, %d blocchi", st["state"], k67.reads)
	}
}

// 4.36: una K67 Pro che risponde al saluto ma invia una memoria senza le
// informazioni della radio (tutti i byte uguali, come può fare una radio spenta
// alimentata dal cavo) non è una radio riconosciuta.
func TestIdentK67BlankRead(t *testing.T) {
	identFast(t)
	k67 := &k67Emu{mem: bytes.Repeat([]byte{0xFF}, xlcRadioLen), model: "BF-K6"}
	identAttach(t, func(p Port) { _ = k67.serve(p) }, "")
	a := newApp(mustStore(t))
	identPortMu.Lock()
	defer identPortMu.Unlock()
	req := &identReq{Port: "COM6"}
	plan := a.identPlan(req)[:2] // i due tentativi della K67 Pro
	if !plan[0].k67 || !plan[1].k67 {
		t.Fatalf("piano: %s, %s", plan[0].name, plan[1].name)
	}
	j := &cableJob{Op: "identify", State: "running", Notes: []string{}, Log: []string{}}
	res, err := a.identRun(context.Background(), j, req, plan, time.Now(), &identResult{Fields: []identField{}, Others: []identOther{}, Answers: []string{}})
	if err != nil || res.Found {
		t.Fatalf("memoria vuota riconosciuta come radio: %v %+v", err, res)
	}
	if len(res.Answers) != 2 || !strings.Contains(res.Answers[0], "le informazioni lette dalla radio non sono valide") {
		t.Fatalf("risposte: %v", res.Answers)
	}
	if es, _ := os.ReadDir(a.backupDir()); len(es) != 0 {
		t.Fatalf("copia di una memoria non valida salvata: %d file", len(es))
	}
	if len(a.store.Get().KnownRadios) != 0 {
		t.Fatal("radio non valida ricordata")
	}
}
