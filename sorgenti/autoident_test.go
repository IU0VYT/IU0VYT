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
// riconosciuta una sola volta, nuova ricerca solo dopo aver ricollegato il cavo.
// Radio accesa dopo il collegamento del cavo: riconosciuta alla ricerca successiva.
func TestAutoIdentConnect(t *testing.T) {
	identFast(t)
	oldS, oldF, oldSl, oldL := autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc
	autoSettle, autoRetryFast, autoRetrySlow = 10*time.Millisecond, 20*time.Millisecond, 20*time.Millisecond
	t.Cleanup(func() { autoSettle, autoRetryFast, autoRetrySlow, listPortsFunc = oldS, oldF, oldSl, oldL })

	m := genByID["bf-uv5r"]
	mem := identModelBase(t, m)
	if mem == nil {
		t.Skip("manca l'immagine di prova della UV-5R")
	}
	var on atomic.Bool
	var mu sync.Mutex
	ports := []PortInfo{}
	listPortsFunc = func() []PortInfo { mu.Lock(); defer mu.Unlock(); return append([]PortInfo{}, ports...) }
	setPorts := func(p ...PortInfo) { mu.Lock(); ports = p; mu.Unlock() }
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	identAttachRaw(func(p Port) {
		if on.Load() {
			genEmuFor(m, append([]byte{}, mem...)).serve(p)
		}
	}, "")

	a := newApp(mustStore(t))
	step := func() { a.autoStep(time.Now()); time.Sleep(15 * time.Millisecond) }
	searches := func() int {
		a.auto.mu.Lock()
		defer a.auto.mu.Unlock()
		if p := a.auto.ports["COM6"]; p != nil {
			return p.tries
		}
		return -1
	}

	// opzione disattivata: nessuna ricerca
	setPorts(PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"}, PortInfo{"COM1", "Porta di comunicazione (COM1)"})
	on.Store(true)
	step()
	step()
	if a.auto.seq != 0 || searches() != -1 {
		t.Fatalf("ricerca con l'opzione disattivata")
	}

	// attivata: la porta del cavo viene interrogata, COM1 no
	st := a.store.Get()
	st.AutoIdent = true
	_ = a.store.SaveSettings(st)
	for i := 0; i < 50 && a.auto.seq == 0; i++ {
		step()
	}
	if a.auto.seq != 1 || a.auto.last == nil || !a.auto.last.Result.Found || a.auto.last.Port != "COM6" {
		t.Fatalf("radio non riconosciuta: %+v", a.auto.last)
	}
	if _, ok := a.auto.ports["COM1"]; ok {
		t.Fatal("COM1 interrogata")
	}
	t.Logf("riconosciuta: %s", a.auto.last.Result.Name)
	// già riconosciuta: nessuna nuova ricerca
	n := searches()
	for i := 0; i < 5; i++ {
		step()
	}
	if searches() != n || a.auto.seq != 1 {
		t.Fatalf("nuova ricerca sulla stessa radio")
	}

	// cavo scollegato e ricollegato con la radio spenta: ricerche ripetute
	on.Store(false)
	setPorts()
	step()
	setPorts(PortInfo{"COM6", "USB-SERIAL CH340 (COM6)"})
	for i := 0; i < 200 && searches() < 1; i++ {
		step()
	}
	if searches() < 1 || a.auto.seq != 1 {
		t.Fatalf("radio spenta: ricerche %d, esiti %d", searches(), a.auto.seq)
	}
	// la radio si accende: riconosciuta alla ricerca successiva
	on.Store(true)
	for i := 0; i < 200 && a.auto.seq == 1; i++ {
		step()
	}
	if a.auto.seq != 2 {
		t.Fatalf("radio accesa dopo il cavo non riconosciuta")
	}

	// disattivazione: stato azzerato
	st = a.store.Get()
	st.AutoIdent = false
	_ = a.store.SaveSettings(st)
	step()
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
