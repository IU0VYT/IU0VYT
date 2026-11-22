//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.46: verifica «in silenzio» per la Radtel RT-880G e le sue gemelle. Dopo la
// scrittura la radio salva i dati e si riavvia (un paio di secondi); la rilettura
// fatta a mano subito dopo il riavvio riesce, mentre la verifica che la cercava
// dopo 3 secondi e poi ogni 3 secondi non otteneva mai risposta (prove di Daniele
// del 20 e 21/11/2026). Ora, per questa famiglia, dopo la scrittura il programma
// non manda niente per verifyQuietFirst e fra un tentativo e l'altro lascia la
// radio in pace per verifyQuietRetry.

import (
	"sync"
	"testing"
	"time"
)

// attese brevi nelle prove (i valori veri sono in cable.go)
func init() {
	verifyQuietFirst, verifyQuietRetry, verifyQuietMax = 400*time.Millisecond, 500*time.Millisecond, 25*time.Second
}

// portClock: registra quando ogni apertura della porta si apre e si chiude
type portClock struct {
	mu     sync.Mutex
	opened []time.Time
	closed []time.Time
}

type clockPort struct {
	Port
	c *portClock
	k int
}

func (p clockPort) Close() error {
	p.c.mu.Lock()
	p.c.closed[p.k] = time.Now()
	p.c.mu.Unlock()
	return p.Port.Close()
}

// withClockEmu: radio simulata; le aperture da from a to (contate da 1) trovano
// la radio muta (to = 0: non torna più)
func withClockEmu(t *testing.T, emu genEmu, from, to int) *portClock {
	t.Helper()
	c := &portClock{}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		c.mu.Lock()
		k := len(c.opened)
		c.opened = append(c.opened, time.Now())
		c.closed = append(c.closed, time.Time{})
		c.mu.Unlock()
		if !(k+1 >= from && (to == 0 || k+1 <= to)) {
			go emu.serve(b)
		}
		return clockPort{Port: a, c: c, k: k}, nil
	}
	t.Cleanup(func() { openPortFunc = old })
	return c
}

// rt880gWriteJob: lettura, poi scrittura con copia di sicurezza e verifica
func rt880gWriteJob(t *testing.T, id string) map[string]interface{} {
	t.Helper()
	c := newAPIClient(t)
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": id, "port": "COM7"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("%s lettura: %v", id, st["error"])
	}
	rd := chansFrom(t, st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"])
	rd[0] = &Chan{Rx: 145500000, Tx: 145500000, Power: 0, Scan: true, Name: "PROVA"}
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": id, "port": "COM7", "channels": rd, "backup": true, "verify": true})
	st, _ = c.wait("")
	if st["state"] != "done" {
		t.Fatalf("%s scrittura: %v", id, st["error"])
	}
	return st
}

func TestVerifyQuietRT880G(t *testing.T) {
	for _, id := range []string{"rtl-rt880g", "rtl-rt871a"} {
		m := genByID[id]
		if m == nil {
			t.Fatalf("modello %s assente", id)
		}
		// aperture: 1 lettura, 2 copia di sicurezza, 3 scrittura, 4 primo tentativo
		// della verifica (radio muta), 5 secondo tentativo (radio pronta)
		clk := withClockEmu(t, genEmuFor(m, nil), 4, 4)
		st := rt880gWriteJob(t, id)
		res := st["result"].(map[string]interface{})
		if res["verified"] != true {
			t.Fatalf("%s: verifica non riuscita: %v / %s", id, res, verifyNotes(st))
		}
		clk.mu.Lock()
		if len(clk.opened) != 5 {
			clk.mu.Unlock()
			t.Fatalf("%s: aperture della porta %d, attese 5", id, len(clk.opened))
		}
		first := clk.opened[3].Sub(clk.closed[2])
		gap := clk.opened[4].Sub(clk.closed[3])
		clk.mu.Unlock()
		// nessun comando alla radio mentre salva e si riavvia
		if first < verifyQuietFirst {
			t.Errorf("%s: primo tentativo %v dopo la fine della scrittura, atteso almeno %v", id, first, verifyQuietFirst)
		}
		// fra un tentativo e l'altro la radio resta in pace
		if gap < verifyQuietRetry {
			t.Errorf("%s: secondo tentativo %v dopo il primo, atteso almeno %v", id, gap, verifyQuietRetry)
		}
	}
}

// le altre radio restano con le attese di prima (BJ-8300, K67 Pro, UV-5R: le
// prove di verify_test.go), anche nella stessa famiglia di driver
func TestVerifyPlanFamilies(t *testing.T) {
	quiet := map[string]bool{"rtl-rt880g": true, "rtl-rt871a": true}
	for id, m := range genByID {
		got := verifyPlanFor(m.Fam)
		want := stdVerifyPlan()
		if quiet[id] {
			want = verifyPlan{verifyQuietFirst, verifyQuietRetry, verifyQuietMax}
		}
		if got != want {
			t.Errorf("%s: attese della verifica %+v, attese %+v", id, got, want)
		}
	}
	for _, id := range []string{"ar-518", "rtl-rt890m", "bj-9000", "bj-8300", "bf-uv5r"} {
		if m := genByID[id]; m != nil && verifyPlanFor(m.Fam) != stdVerifyPlan() {
			t.Errorf("%s: deve restare con le attese di prima", id)
		}
	}
}
