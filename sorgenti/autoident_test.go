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

// Regole della 4.2: cavo senza radio = ricerche silenziose (nessun riquadro);
// radio collegata = riquadro mentre la radio risponde, poi riconosciuta; dopo il
// riconoscimento la ricerca si ferma (anche se la radio viene cambiata) finché non
// si preme «Identifica Nuova Radio Automaticamente» o non si collega un altro cavo.
func TestAutoIdentConnect(t *testing.T) {
	identFast(t)
	// 4.2: 81 tentativi per ricerca; le radio simulate rispondono subito, quindi
	// attese più brevi di identFast (altrimenti due ricerche senza radio
	// richiedono più di due minuti)
	identFirstWait, identSilentMax = 80*time.Millisecond, 300*time.Millisecond
	oldS, oldF, oldSl, oldL, oldQ := autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoRetrySlow, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() {
		autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc, autoQuiet = oldS, oldF, oldSl, oldL, oldQ
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
	waitFor := func(what string, max time.Duration, cond func() bool) {
		t.Helper()
		end := time.Now().Add(max)
		for !cond() {
			if time.Now().After(end) {
				t.Fatalf("%s: tempo scaduto (stato %q)", what, a.auto.status)
			}
			a.autoStep(time.Now())
			state()
			time.Sleep(20 * time.Millisecond)
		}
	}
	cable := PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"}

	// opzione disattivata: nessuna ricerca; «Identifica Nuova Radio» rifiutato
	setPorts(cable, PortInfo{"COM1", "Porta di comunicazione (COM1)"})
	plug(&uv5r)
	for i := 0; i < 5; i++ {
		a.autoStep(time.Now())
		time.Sleep(20 * time.Millisecond)
	}
	if seq() != 0 || a.auto.run != nil {
		t.Fatal("ricerca con l'opzione disattivata")
	}
	if _, err := c.call("identify/auto", map[string]interface{}{"rescan": true}); err == nil {
		t.Fatal("nuova ricerca accettata con l'opzione disattivata")
	}

	// attivata, cavo senza radio: ricerche silenziose, nessun riquadro
	plug(nil)
	c.must("identify/auto", map[string]interface{}{"enabled": true})
	waitFor("ricerche senza radio", 150*time.Second, func() bool { return tries("COM6") >= 2 })
	if banner.Load() {
		t.Fatal("riquadro visibile senza nessuna radio collegata")
	}
	if tries("COM1") != -1 {
		t.Fatal("COM1 interrogata")
	}

	// radio collegata: riquadro, poi radio riconosciuta e ricerca ferma
	plug(&uv5r)
	waitFor("radio riconosciuta", 90*time.Second, func() bool { return seq() == 1 })
	if !banner.Load() {
		t.Fatal("riquadro mai visibile con la radio collegata")
	}
	if a.auto.last.Result.Kind != "bf-uv5r" || a.auto.last.Port != "COM6" {
		t.Fatalf("radio riconosciuta: %+v", a.auto.last)
	}
	if st := state(); st["searching"] == true || st["busy"] == true {
		t.Fatal("ricerca ancora in corso dopo il riconoscimento")
	}
	// radio cambiata con il cavo nel PC: nessuna nuova ricerca
	plug(&k5)
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
	if a.auto.last.Result.Kind != "qs-uvk5" || !banner.Load() {
		t.Fatalf("nuova radio: %+v", a.auto.last)
	}

	// un altro cavo collegato dopo il riconoscimento: viene cercato
	setPorts(cable, PortInfo{"COM9", "Prolific USB-to-Serial Comm Port (COM9)"})
	plug(&uv5r)
	waitFor("cavo nuovo", 90*time.Second, func() bool { return seq() == 3 })
	if a.auto.last.Port != "COM9" {
		t.Fatalf("cavo nuovo: %+v", a.auto.last)
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
