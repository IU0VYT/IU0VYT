//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Verifica dopo la scrittura con radio che si riavviano (3.9.2): la rilettura
// riprova finché la radio torna a rispondere (Bajeton BJ-8300, K67 Pro, radio
// dei driver); se la radio non torna, la scrittura resta riuscita e la nota lo
// spiega.

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// attese brevi in tutte le prove (i valori veri sono in cable.go)
func init() {
	verifyFirstWait, verifyRetryWait, verifyMaxWait = 100*time.Millisecond, 150*time.Millisecond, 25*time.Second
}

// radio simulata che «si riavvia»: le aperture della porta dalla numero from
// (contate da 1) fino a to non trovano la radio (nessuna risposta, o porta
// assente se gone); to = 0: non torna più
func withRebootingEmu(t *testing.T, serve func(Port) error, from, to int, gone bool) func() int {
	t.Helper()
	old := openPortFunc
	var mu sync.Mutex
	n := 0
	openPortFunc = func(name string, baud int) (Port, error) {
		mu.Lock()
		n++
		k := n
		mu.Unlock()
		a, b := newPipe()
		a.baud = baud
		if k >= from && (to == 0 || k <= to) {
			if gone {
				return nil, errPortGone
			}
			return a, nil // radio spenta: nessuna risposta
		}
		go serve(b)
		return a, nil
	}
	t.Cleanup(func() { openPortFunc = old })
	return func() int { mu.Lock(); defer mu.Unlock(); return n }
}

var errPortGone = &portErr{"impossibile aprire la porta COM4: il dispositivo non è collegato"}

type portErr struct{ s string }

func (e *portErr) Error() string { return e.s }

func verifyNotes(st map[string]interface{}) string {
	var out []string
	for _, n := range st["notes"].([]interface{}) {
		out = append(out, n.(string))
	}
	return strings.Join(out, " | ")
}

func TestVerifyAfterRebootBajeton(t *testing.T) {
	for _, gone := range []bool{false, true} {
		radio := newBtEmu(btSampleImage(t))
		// aperture: 1 lettura, 2 copia di sicurezza, 3 scrittura, 4-5 radio in riavvio, 6 verifica
		opens := withRebootingEmu(t, radio.serve, 4, 5, gone)
		c := newAPIClient(t)
		c.must("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM4"})
		st, _ := c.wait("")
		if st["state"] != "done" {
			t.Fatalf("lettura: %v", st["error"])
		}
		chs := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"]
		c.must("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM4", "backup": true, "verify": true, "channels": chs})
		st, _ = c.wait("")
		res := st["result"].(map[string]interface{})
		if st["state"] != "done" || res["verified"] != true || !strings.Contains(verifyNotes(st), "tentativo 3") {
			t.Fatalf("porta assente %v: %v %v / %s", gone, st["state"], res, verifyNotes(st))
		}
		if opens() != 6 {
			t.Fatalf("aperture %d", opens())
		}
		t.Logf("porta assente %v: %s", gone, verifyNotes(st))
	}
}

func TestVerifyRadioNotBack(t *testing.T) {
	radio := newBtEmu(btSampleImage(t))
	withRebootingEmu(t, radio.serve, 4, 0, false)
	c := newAPIClient(t)
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM4"})
	st, _ := c.wait("")
	chs := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"]
	t0 := time.Now()
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM4", "backup": true, "verify": true, "channels": chs})
	st, _ = c.wait("")
	res := st["result"].(map[string]interface{})
	n := verifyNotes(st)
	if st["state"] != "done" || res["verified"] != false || res["verifyNoRead"] != true || !strings.Contains(n, "Scrittura completata") ||
		!strings.Contains(n, "Leggi dalla radio") {
		t.Fatalf("%v %v / %s", st["state"], res, n)
	}
	if d := time.Since(t0); d > verifyMaxWait+20*time.Second {
		t.Fatalf("attesa troppo lunga: %v", d)
	}
	t.Logf("%s", n)
}

func TestVerifyAfterRebootK67AndDriver(t *testing.T) {
	// K67 Pro
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "5RH-PRO"}
	withRebootingEmu(t, k67.serve, 4, 4, false)
	c := newAPIClient(t)
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM3"})
	st, _ := c.wait("")
	chs := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"]
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "backup": true, "verify": true, "channels": chs})
	st, _ = c.wait("")
	if r := st["result"].(map[string]interface{}); st["state"] != "done" || r["verified"] != true || !strings.Contains(verifyNotes(st), "tentativo 2") {
		t.Fatalf("K67: %v %v / %s", st["state"], r, verifyNotes(st))
	}
	// radio dei driver (UV-5R)
	m := genByID["bf-uv5r"]
	mem := identModelBase(t, m)
	if mem == nil {
		t.Skip("CHIRP_DIR non impostata: manca la memoria di prova della UV-5R")
	}
	emu := genEmuFor(m, mem)
	withRebootingEmu(t, func(p Port) error { emu.serve(p); return nil }, 4, 5, true)
	c2 := newAPIClient(t)
	c2.must("cable/start", map[string]interface{}{"op": "read", "radio": m.ID, "port": "COM7"})
	st = identWait(c2)
	if st["state"] != "done" {
		t.Fatalf("UV-5R lettura: %v", st["error"])
	}
	chs = st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"]
	c2.must("cable/start", map[string]interface{}{"op": "write", "radio": m.ID, "port": "COM7", "backup": true, "verify": true, "channels": chs})
	st = identWait(c2)
	if r := st["result"].(map[string]interface{}); st["state"] != "done" || r["verified"] != true || !strings.Contains(verifyNotes(st), "tentativo 3") {
		t.Fatalf("UV-5R: %v %v / %s", st["state"], r, verifyNotes(st))
	}
}
