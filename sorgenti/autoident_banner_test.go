//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.4: il riquadro «Identificazione Radio In Corso» compare solo per la porta a cui
// è collegata una radio che risponde. Un cavo senza radio può ricevere disturbi
// (0x00, 0xFF, trame incomplete) e l'eco, anche in ritardo, di quanto inviato: non
// è una radio.

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// noisyCable: cavo senza radio con disturbi sulla linea e l'eco ritardata dei
// byte inviati (come certi cavi con la radio spenta o scollegata)
type noisyCable struct {
	*pipeEnd
	stop chan struct{}
	once sync.Once
}

func (c *noisyCable) Close() error {
	c.once.Do(func() { close(c.stop) })
	return c.pipeEnd.Close()
}

func newNoisyCable(seed int64) Port {
	x, y := newPipe()
	c := &noisyCable{pipeEnd: x, stop: make(chan struct{})}
	noise := []byte{0x00, 0xFF, 0xF8, 0x80, 0xFE, 0xE0, 0x00, 0x00}
	r := rand.New(rand.NewSource(seed))
	var rmu sync.Mutex
	go func() { // disturbi a gruppi
		for {
			select {
			case <-c.stop:
				return
			case <-time.After(25 * time.Millisecond):
				rmu.Lock()
				n := 1 + r.Intn(3)
				b := make([]byte, n)
				for i := range b {
					b[i] = noise[r.Intn(len(noise))]
				}
				rmu.Unlock()
				_ = y.Write(b)
			}
		}
	}()
	type echoed struct {
		v  byte
		at time.Time
	}
	q := make(chan echoed, 1<<16)
	go func() { // eco in ritardo (40 ms, nello stesso ordine): arriva anche dopo la fine del tentativo
		for {
			select {
			case <-c.stop:
				return
			case e := <-q:
				time.Sleep(time.Until(e.at))
				_ = y.Write([]byte{e.v})
			}
		}
	}()
	go func() {
		b := make([]byte, 1)
		for {
			select {
			case <-c.stop:
				return
			default:
			}
			if err := y.ReadFull(context.Background(), b, 50*time.Millisecond); err != nil {
				continue
			}
			q <- echoed{b[0], time.Now().Add(40 * time.Millisecond)}
		}
	}()
	return c
}

func TestAutoIdentBanner(t *testing.T) {
	identFast(t)
	identFirstWait, identSilentMax = 60*time.Millisecond, 180*time.Millisecond
	oldS, oldF, oldU, oldL, oldQ := autoSettle, autoRetryFast, autoRetryUnknown, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() {
		autoSettle, autoRetryFast, autoRetryUnknown, listPortsFunc, autoQuiet = oldS, oldF, oldU, oldL, oldQ
	})
	m := genByID["bf-uv5r"]
	mem := identModelBase(t, m)
	if mem == nil {
		t.Skip("manca l'immagine di prova di CHIRP (CHIRP_DIR)")
	}
	listPortsFunc = func() []PortInfo {
		return []PortInfo{{"COM6", "USB-SERIAL CH340 (COM6)"}, {"COM7", "USB-SERIAL CH340 (COM7)"}}
	}
	var rmu sync.Mutex
	radioOn := ""
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	seed := int64(1)
	openPortFunc = func(name string, baud int) (Port, error) {
		rmu.Lock()
		on := radioOn
		seed++
		s := seed
		rmu.Unlock()
		if strings.EqualFold(name, on) {
			x, y := newPipe()
			x.baud = baud
			go genEmuFor(m, append([]byte{}, mem...)).serve(y)
			return x, nil
		}
		return newNoisyCable(s), nil
	}
	var hmu sync.Mutex
	hooks := 0
	identSureHook = func(key string, clean, echo []byte) {
		hmu.Lock()
		defer hmu.Unlock()
		if hooks++; hooks <= 5 {
			t.Logf("risposta sicura su %s: % X (eco … % X)", key, clean, echo[max(0, len(echo)-24):])
		}
	}
	t.Cleanup(func() { identSureHook = nil })
	c := newAPIClient(t)
	a := c.app
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })

	// campiona lo stato come l'interfaccia (ogni 20 ms) e annota le porte del riquadro
	watch := func(d time.Duration, stopOn func(map[string]interface{}) bool) map[string]int {
		seen := map[string]int{}
		end := time.Now().Add(d)
		for first := true; first || time.Now().Before(end); first = false {
			a.autoStep(time.Now())
			r := c.must("identify/auto", nil)
			if r["searching"] == true {
				seen["banner"]++
			}
			if l, ok := r["answering"].([]interface{}); ok {
				for _, p := range l {
					seen[p.(string)]++
				}
			}
			if stopOn != nil && stopOn(r) {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		return seen
	}
	found := func(port string) func(map[string]interface{}) bool {
		return func(r map[string]interface{}) bool {
			a.auto.mu.Lock()
			defer a.auto.mu.Unlock()
			for _, e := range a.auto.events {
				if e.Port == port {
					return true
				}
			}
			return false
		}
	}

	// 1. due cavi, nessuna radio: mai il riquadro (almeno due giri completi)
	c.must("identify/auto", map[string]bool{"enabled": true})
	seen := watch(0, nil)
	t0 := time.Now()
	for time.Since(t0) < 60*time.Second {
		s := watch(time.Second, nil)
		for k, v := range s {
			seen[k] += v
		}
		a.auto.mu.Lock()
		p6, p7 := a.auto.ports["COM6"], a.auto.ports["COM7"]
		done := p6 != nil && p7 != nil && p6.tries >= 2 && p7.tries >= 2
		a.auto.mu.Unlock()
		if done {
			break
		}
	}
	a.auto.mu.Lock()
	p6, p7 := a.auto.ports["COM6"], a.auto.ports["COM7"]
	a.auto.mu.Unlock()
	if p6 == nil || p7 == nil {
		t.Fatal("porte non cercate")
	}
	a.auto.mu.Lock()
	t6, t7, u6, u7 := p6.tries, p7.tries, p6.unknown, p7.unknown
	a.auto.mu.Unlock()
	t.Logf("nessuna radio: giri COM6 %d, COM7 %d, riquadro %v (%v)", t6, t7, seen["banner"], time.Since(t0).Round(time.Second))
	if t6 < 2 || t7 < 2 {
		t.Fatalf("giri non completati: %d %d", t6, t7)
	}
	if seen["banner"] > 0 || seen["COM6"] > 0 || seen["COM7"] > 0 {
		t.Fatalf("riquadro visibile senza radio: %v", seen)
	}
	if u6 || u7 {
		t.Fatal("cavo senza radio scambiato per una radio sconosciuta")
	}

	// 2. e 3. radio su una sola porta: riquadro solo per quella
	for _, cs := range []struct{ on, off string }{{"COM7", "COM6"}, {"COM6", "COM7"}} {
		rmu.Lock()
		radioOn = cs.on
		rmu.Unlock()
		c.must("identify/auto", map[string]interface{}{"rescan": true})
		s := watch(90*time.Second, found(cs.on))
		t.Logf("radio su %s, nessuna radio su %s: %v", cs.on, cs.off, s)
		if !found(cs.on)(nil) {
			t.Fatalf("radio su %s non riconosciuta", cs.on)
		}
		if s[cs.off] > 0 {
			t.Fatalf("riquadro per %s, dove non c'è nessuna radio: %v", cs.off, s)
		}
		if s[cs.on] == 0 {
			t.Fatalf("riquadro mai visibile per %s, dove c'è la radio", cs.on)
		}
		// la radio torna spenta prima del caso successivo
		rmu.Lock()
		radioOn = ""
		rmu.Unlock()
		a.auto.mu.Lock()
		a.auto.events = nil
		a.auto.mu.Unlock()
	}
}
