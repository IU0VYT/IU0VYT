//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAutoIdentCablePorts(t *testing.T) {
	for _, c := range []struct {
		p  PortInfo
		ok bool
	}{
		{PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"}, true},
		{PortInfo{"COM3", "Prolific USB-to-Serial Comm Port (COM3)"}, true},
		{PortInfo{"COM9", "Silicon Labs CP210x USB to UART Bridge (COM9)"}, true},
		{PortInfo{"COM4", "USB Serial Port (COM4)"}, true},
		{PortInfo{"COM5", "Collegamento seriale standard tramite Bluetooth (COM5)"}, false},
		{PortInfo{"COM1", "Porta di comunicazione (COM1)"}, false},
		{PortInfo{"COM1", ""}, false},
		{PortInfo{"COM8", ""}, true},
		{PortInfo{"COM10", "Intel(R) Active Management Technology - SOL (COM10)"}, false},
	} {
		if got := cableLikePort(c.p); got != c.ok {
			t.Errorf("%v: %v, atteso %v", c.p, got, c.ok)
		}
	}
}

// Regole della 4.3: cavo senza radio = ascolto continuo e silenzioso (nessun
// riquadro); radio accesa dopo = trovata al giro successivo, riquadro mentre
// risponde; dopo il riconoscimento la ricerca su quel cavo si ferma finché non si
// preme «Identifica Nuova Radio Automaticamente»; più cavi con più radio si
// cercano e si riconoscono insieme.
func TestAutoIdentConnect(t *testing.T) {
	identFast(t)
	// le radio simulate rispondono subito: attese brevi (81 tentativi per giro)
	identFirstWait, identSilentMax = 80*time.Millisecond, 250*time.Millisecond
	oldS, oldF, oldU, oldL, oldQ := autoSettle, autoRetryFast, autoRetryUnknown, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() {
		autoSettle, autoRetryFast, autoRetryUnknown, listPortsFunc, autoQuiet = oldS, oldF, oldU, oldL, oldQ
	})

	type radio struct {
		m   *genModel
		mem []byte
	}
	mk := func(id string) *radio {
		m := genByID[id]
		mem := identModelBase(t, m)
		if mem == nil {
			t.Skip("manca l'immagine di prova di CHIRP (CHIRP_DIR)")
		}
		return &radio{m, mem}
	}
	uv5r, k5, rt490 := mk("bf-uv5r"), mk("qs-uvk5"), mk("rtl-rt490")
	// radio collegata a ogni cavo (nessuna: cavo senza radio o radio spenta)
	var rmu sync.Mutex
	onPort := map[string]*radio{}
	plug := func(port string, r *radio) { rmu.Lock(); onPort[port] = r; rmu.Unlock() }
	var mu sync.Mutex
	ports := []PortInfo{}
	listPortsFunc = func() []PortInfo { mu.Lock(); defer mu.Unlock(); return append([]PortInfo{}, ports...) }
	setPorts := func(p ...PortInfo) { mu.Lock(); ports = p; mu.Unlock() }
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	openPortFunc = func(name string, baud int) (Port, error) {
		x, y := newPipe()
		x.baud = baud
		rmu.Lock()
		r := onPort[strings.ToUpper(name)]
		rmu.Unlock()
		if r != nil {
			go genEmuFor(r.m, append([]byte{}, r.mem...)).serve(y)
		}
		return x, nil
	}

	c := newAPIClient(t)
	a := c.app
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })
	var banner atomic.Bool // il riquadro è stato visibile (searching) almeno una volta
	state := func() map[string]interface{} {
		r := c.must("identify/auto", nil)
		if r["searching"] == true {
			banner.Store(true)
		}
		return r
	}
	seq := func() int { a.auto.mu.Lock(); defer a.auto.mu.Unlock(); return a.auto.seq }
	tries := func(n string) int {
		a.auto.mu.Lock()
		defer a.auto.mu.Unlock()
		if p := a.auto.ports[n]; p != nil {
			return p.tries
		}
		return -1
	}
	event := func(port string) *autoEvent {
		a.auto.mu.Lock()
		defer a.auto.mu.Unlock()
		for k := len(a.auto.events) - 1; k >= 0; k-- {
			if a.auto.events[k].Port == port {
				return a.auto.events[k]
			}
		}
		return nil
	}
	waitFor := func(what string, max time.Duration, cond func() bool) time.Duration {
		t.Helper()
		t0 := time.Now()
		end := t0.Add(max)
		for !cond() {
			if time.Now().After(end) {
				t.Fatalf("%s: tempo scaduto (stato %q)", what, a.auto.status)
			}
			a.autoStep(time.Now())
			state()
			time.Sleep(20 * time.Millisecond)
		}
		return time.Since(t0)
	}
	cable := PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"}

	// opzione disattivata: nessuna ricerca; «Identifica Nuova Radio» rifiutato
	setPorts(cable, PortInfo{"COM1", "Porta di comunicazione (COM1)"})
	plug("COM6", uv5r)
	for i := 0; i < 5; i++ {
		a.autoStep(time.Now())
		time.Sleep(20 * time.Millisecond)
	}
	if seq() != 0 || len(c.must("identify/auto", nil)["searchPorts"].([]interface{})) != 0 {
		t.Fatal("ricerca con l'opzione disattivata")
	}
	if _, err := c.call("identify/auto", map[string]interface{}{"rescan": true}); err == nil {
		t.Fatal("nuova ricerca accettata con l'opzione disattivata")
	}

	// attivata, cavo con la radio spenta: ascolto continuo, nessun riquadro
	plug("COM6", nil)
	c.must("identify/auto", map[string]interface{}{"enabled": true})
	d := waitFor("ricerche senza radio", 150*time.Second, func() bool { return tries("COM6") >= 2 })
	t.Logf("due giri di ascolto senza radio in %v", d.Round(time.Millisecond))
	if banner.Load() {
		t.Fatal("riquadro visibile senza nessuna radio collegata")
	}
	if tries("COM1") != -1 {
		t.Fatal("COM1 interrogata")
	}

	// radio accesa dopo il cavo: riquadro, poi radio riconosciuta e ricerca ferma
	plug("COM6", uv5r)
	d = waitFor("radio accesa dopo il cavo", 90*time.Second, func() bool { return seq() == 1 })
	t.Logf("radio accesa dopo il cavo riconosciuta in %v", d.Round(time.Millisecond))
	if !banner.Load() {
		t.Fatal("riquadro mai visibile con la radio collegata")
	}
	if ev := event("COM6"); ev == nil || ev.Result.Kind != "bf-uv5r" {
		t.Fatalf("radio riconosciuta: %+v", ev)
	}
	if k := a.store.Get().KnownRadios; len(k) == 0 || k[0] != "bf-uv5r" {
		t.Fatalf("radio non ricordata: %v", k)
	}
	if st := state(); st["searching"] == true || st["busy"] == true {
		t.Fatal("ricerca ancora in corso dopo il riconoscimento")
	}
	// radio cambiata con il cavo nel PC: nessuna nuova ricerca
	plug("COM6", k5)
	n := tries("COM6")
	time0 := time.Now()
	waitFor("pausa", 10*time.Second, func() bool { return time.Since(time0) > time.Second })
	if tries("COM6") != n || seq() != 1 {
		t.Fatal("nuova ricerca dopo il riconoscimento senza «Identifica Nuova Radio»")
	}
	// «Identifica Nuova Radio Automaticamente»: la nuova radio viene riconosciuta
	banner.Store(false)
	r := c.must("identify/auto", map[string]interface{}{"rescan": true})
	if r["rescanned"] != float64(1) {
		t.Fatalf("nuova ricerca: %v", r)
	}
	waitFor("nuova radio", 120*time.Second, func() bool { return seq() == 2 })
	if ev := event("COM6"); ev.Result.Kind != "qs-uvk5" || !banner.Load() {
		t.Fatalf("nuova radio: %+v", ev)
	}
	// la radio riconosciuta prima ora si prova per prima: la ricerca è più breve
	plug("COM6", uv5r)
	c.must("identify/auto", map[string]interface{}{"rescan": true})
	d = waitFor("radio già nota", 60*time.Second, func() bool { return seq() == 3 })
	t.Logf("radio già riconosciuta in passato trovata in %v", d.Round(time.Millisecond))
	if ev := event("COM6"); ev.Result.Kind != "bf-uv5r" || ev.Result.Tried > 3 {
		t.Fatalf("radio nota non provata per prima: %+v (tentativi %d)", ev, ev.Result.Tried)
	}

	// due cavi nuovi con due radio, collegati insieme: riconosciute insieme
	setPorts(cable, PortInfo{"COM9", "Prolific USB-to-Serial Comm Port (COM9)"}, PortInfo{"COM10", "USB-SERIAL CH340 (COM10)"})
	plug("COM9", k5)
	plug("COM10", rt490)
	parallel := false
	waitFor("due cavi", 120*time.Second, func() bool {
		if sp, _ := c.must("identify/auto", nil)["searchPorts"].([]interface{}); len(sp) == 2 {
			parallel = true
		}
		return seq() == 5
	})
	if !parallel {
		t.Fatal("le due porte non sono state cercate insieme")
	}
	if e9, e10 := event("COM9"), event("COM10"); e9 == nil || e10 == nil || e9.Result.Kind != "qs-uvk5" || e10.Result.Kind != "rtl-rt490" {
		t.Fatalf("due cavi: %+v %+v", e9, e10)
	}
	for _, sq := range []int{4, 5} {
		if a.autoIdentSession(sq) == nil {
			t.Fatalf("radio dell'esito %d non apribile", sq)
		}
	}
	if st := state(); len(st["events"].([]interface{})) != 5 {
		t.Fatalf("esiti: %v", st["events"])
	}

	// disattivazione: stato azzerato
	c.must("identify/auto", map[string]interface{}{"enabled": false})
	a.autoStep(time.Now())
	if len(a.auto.ports) != 0 {
		t.Fatalf("stato non azzerato dopo la disattivazione")
	}
}

// Un'operazione chiesta dall'utente interrompe la ricerca automatica in corso.
func TestAutoIdentYield(t *testing.T) {
	identFast(t)
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	identAttachRaw(func(p Port) {}, "") // nessuna radio: ricerca lunga
	c := newAPIClient(t)
	a := c.app
	j, err := a.startIdentify(&identReq{Port: "COM6"}, true)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if !j.isRunning() {
		t.Fatal("ricerca automatica non in corso")
	}
	t0 := time.Now()
	c.must("identify/start", map[string]string{"port": "COM7"})
	if j.isRunning() {
		t.Fatal("la ricerca automatica non ha lasciato il posto")
	}
	t.Logf("ricerca automatica interrotta in %v", time.Since(t0).Round(time.Millisecond))
	c.must("cable/cancel", nil)
	identWait(c)
	// API: attivazione memorizzata nelle impostazioni
	r := c.must("identify/auto", map[string]interface{}{"enabled": true})
	if r["enabled"] != true || !a.store.Get().AutoIdent {
		t.Fatalf("attivazione: %v", r)
	}
	r = c.must("identify/auto", map[string]interface{}{"enabled": false})
	if r["enabled"] != false || a.store.Get().AutoIdent {
		t.Fatalf("disattivazione: %v", r)
	}
}
