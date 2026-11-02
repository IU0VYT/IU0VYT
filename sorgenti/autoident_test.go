//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
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

// Radio collegata quando l'opzione è attiva: ricerca automatica, radio
// riconosciuta una sola volta mentre resta collegata (controllo della presenza);
// radio staccata dal cavo, che resta nel PC: segnalata come scollegata e
// ritrovata quando si ricollega (anche accesa dopo); radio sostituita con un
// altro modello: riconosciuta; cavo staccato e ricollegato durante una ricerca.
func TestAutoIdentConnect(t *testing.T) {
	identFast(t)
	oldS, oldF, oldSl, oldL := autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc
	oldPE, oldPW, oldQ := autoPresenceEvery, autoPresenceWait, autoQuiet
	autoSettle, autoRetryFast, autoRetrySlow = 10*time.Millisecond, 20*time.Millisecond, 20*time.Millisecond
	autoPresenceEvery, autoPresenceWait, autoQuiet = 60*time.Millisecond, 400*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() {
		autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc = oldS, oldF, oldSl, oldL
		autoPresenceEvery, autoPresenceWait, autoQuiet = oldPE, oldPW, oldQ
	})

	type radio struct {
		m   *genModel
		mem []byte
	}
	mk := func(id string) radio {
		m := genByID[id]
		mem := identModelBase(t, m)
		if mem == nil {
			t.Skip("manca l'immagine di prova di CHIRP (CHIRP_DIR)")
		}
		return radio{m, mem}
	}
	uv5r, k5 := mk("bf-uv5r"), mk("qs-uvk5")
	var cur atomic.Value // radio collegata al cavo (nil: nessuna)
	cur.Store((*radio)(nil))
	var mu sync.Mutex
	ports := []PortInfo{}
	listPortsFunc = func() []PortInfo { mu.Lock(); defer mu.Unlock(); return append([]PortInfo{}, ports...) }
	setPorts := func(p ...PortInfo) { mu.Lock(); ports = p; mu.Unlock() }
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	identAttachRaw(func(p Port) {
		if r := cur.Load().(*radio); r != nil {
			genEmuFor(r.m, append([]byte{}, r.mem...)).serve(p)
		}
	}, "")
	plug := func(r *radio) { cur.Store(r) }

	a := newApp(mustStore(t))
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })
	snap := func() (seq, lost int, last *autoEvent) {
		a.auto.mu.Lock()
		defer a.auto.mu.Unlock()
		return a.auto.seq, a.auto.lostSeq, a.auto.last
	}
	waitFor := func(what string, max time.Duration, cond func() bool) {
		t.Helper()
		end := time.Now().Add(max)
		for !cond() {
			if time.Now().After(end) {
				t.Fatalf("%s: tempo scaduto (stato %q)", what, a.auto.status)
			}
			a.autoStep(time.Now())
			time.Sleep(20 * time.Millisecond)
		}
	}
	cable := PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"}

	// opzione disattivata: nessuna ricerca
	setPorts(cable, PortInfo{"COM1", "Porta di comunicazione (COM1)"})
	plug(&uv5r)
	for i := 0; i < 5; i++ {
		a.autoStep(time.Now())
		time.Sleep(20 * time.Millisecond)
	}
	if seq, _, _ := snap(); seq != 0 || a.auto.run != nil {
		t.Fatalf("ricerca con l'opzione disattivata")
	}

	// attivata: la porta del cavo viene interrogata, COM1 no
	st := a.store.Get()
	st.AutoIdent = true
	_ = a.store.SaveSettings(st)
	waitFor("prima radio", 90*time.Second, func() bool { s, _, _ := snap(); return s == 1 })
	_, _, last := snap()
	if last.Port != "COM6" || !last.Result.Found || last.Result.Kind != "bf-uv5r" || last.Again {
		t.Fatalf("radio non riconosciuta: %+v", last)
	}
	if _, ok := a.auto.ports["COM1"]; ok {
		t.Fatal("COM1 interrogata")
	}
	// radio ancora collegata: controlli della presenza, nessuna nuova ricerca
	time0 := time.Now()
	waitFor("controlli della presenza", 10*time.Second, func() bool { return time.Since(time0) > 1500*time.Millisecond })
	if s, l, _ := snap(); s != 1 || l != 0 {
		t.Fatalf("radio collegata data per scollegata o cercata di nuovo: esiti %d, scollegamenti %d", s, l)
	}

	// radio staccata dal cavo (il cavo resta nel PC): scollegata, poi ricerche
	plug(nil)
	waitFor("scollegamento", 20*time.Second, func() bool { _, l, _ := snap(); return l == 1 })
	waitFor("ricerche senza radio", 60*time.Second, func() bool {
		a.auto.mu.Lock()
		defer a.auto.mu.Unlock()
		p := a.auto.ports["COM6"]
		return p != nil && p.tries >= 1
	})
	// la stessa radio ricollegata (o accesa dopo): riconosciuta di nuovo
	plug(&uv5r)
	waitFor("stessa radio ricollegata", 90*time.Second, func() bool { s, _, _ := snap(); return s == 2 })
	if _, _, last = snap(); !last.Again || last.Result.Kind != "bf-uv5r" {
		t.Fatalf("radio ricollegata: %+v", last)
	}

	// radio sostituita con un altro modello, sempre con il cavo nel PC
	plug(nil)
	waitFor("scollegamento 2", 20*time.Second, func() bool { _, l, _ := snap(); return l == 2 })
	plug(&k5)
	waitFor("radio diversa", 120*time.Second, func() bool { s, _, _ := snap(); return s == 3 })
	if _, _, last = snap(); last.Again || last.Result.Kind != "qs-uvk5" {
		t.Fatalf("radio sostituita: %+v", last)
	}

	// cavo staccato dal PC e ricollegato con un'altra radio mentre si cerca
	plug(nil)
	setPorts()
	waitFor("cavo staccato", 10*time.Second, func() bool { a.auto.mu.Lock(); defer a.auto.mu.Unlock(); return len(a.auto.ports) == 0 })
	setPorts(cable)
	waitFor("ricerca sul cavo ricollegato", 10*time.Second, func() bool { a.auto.mu.Lock(); defer a.auto.mu.Unlock(); return a.auto.run != nil })
	setPorts()
	time.Sleep(30 * time.Millisecond)
	plug(&uv5r)
	setPorts(cable)
	waitFor("radio sul cavo ricollegato", 90*time.Second, func() bool { s, _, _ := snap(); return s == 4 })

	// disattivazione: stato azzerato
	st = a.store.Get()
	st.AutoIdent = false
	_ = a.store.SaveSettings(st)
	a.autoStep(time.Now())
	if len(a.auto.ports) != 0 {
		t.Fatalf("stato non azzerato dopo la disattivazione")
	}
	t.Logf("esiti %d, scollegamenti %d", a.auto.seq, a.auto.lostSeq)
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
