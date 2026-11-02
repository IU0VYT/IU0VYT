//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// «Inoltra ad altra radio» (3.9.3) con le radio simulate: lettura e copia di
// sicurezza della radio di destinazione, adattamento, posizioni libere o
// sostituzione di tutti i canali, scrittura, verifica; nessuna scrittura se
// qualcosa non va prima.

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// canali restituiti dall'API come []*Chan
func apiChans(t *testing.T, v interface{}) []*Chan {
	t.Helper()
	b, _ := json.Marshal(v)
	var out []*Chan
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func readVia(t *testing.T, c *apiClient, radio, port string) []*Chan {
	t.Helper()
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": radio, "port": port})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("lettura %s: %v", radio, st["error"])
	}
	return apiChans(t, st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"])
}

func someSource(t *testing.T, n int) []*Chan {
	t.Helper()
	src := readXLC(mustRead(t, sampleXLC), &Report{})
	var out []*Chan
	for _, c := range src {
		if c != nil && len(out) < n {
			out = append(out, c)
		}
	}
	if len(out) < n {
		t.Fatalf("canali di prova insufficienti: %d", len(out))
	}
	return out
}

func TestForwardK67ToBajetonFree(t *testing.T) {
	radio := newBtEmu(btSampleImage(t))
	withRebootingEmu(t, radio.serve, 1<<30, 0, false)
	c := newAPIClient(t)
	before := readVia(t, c, "baj", "COM4")
	src := someSource(t, 5)
	// una zona del K67 con un nome lungo: sulla Bajeton viene accorciato a 12 caratteri
	src[0].Name = "NOME MOLTO LUNGO"
	c.must("cable/start", map[string]interface{}{"op": "forward", "radio": "baj", "port": "COM4", "channels": src, "mode": "free"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("inoltro: %v / %s", st["error"], verifyNotes(st))
	}
	res := st["result"].(map[string]interface{})
	if res["verified"] != true || int(res["count"].(float64)) != 5 || res["backup"] == nil {
		t.Fatalf("risultato: %v / %s", res, verifyNotes(st))
	}
	if _, err := os.Stat(res["backup"].(string)); err != nil || !strings.Contains(res["backup"].(string), "prima_dell_inoltro") {
		t.Fatalf("copia di sicurezza: %v %v", res["backup"], err)
	}
	var placed []int
	for _, p := range res["placed"].([]interface{}) {
		placed = append(placed, int(p.(float64)))
	}
	// posizioni: le prime libere della radio, in ordine
	var free []int
	for i, ch := range before {
		if ch == nil && len(free) < 5 {
			free = append(free, i)
		}
	}
	if len(placed) != 5 || !equalInts(placed, free) {
		t.Fatalf("posizioni %v, attese %v", placed, free)
	}
	after := readVia(t, c, "baj", "COM4")
	isPlaced := map[int]int{}
	for k, p := range placed {
		isPlaced[p] = k
	}
	for i := range after {
		k, ok := isPlaced[i]
		if !ok {
			if !chanEqual(after[i], before[i]) {
				t.Fatalf("canale %d cambiato: %+v → %+v", i, before[i], after[i])
			}
			continue
		}
		want := adaptChan(src[k], kindDAT, nil, "", &Report{})
		if after[i] == nil || after[i].Rx != want.Rx || after[i].Tx != want.Tx || after[i].Name != want.Name || after[i].TxTone != want.TxTone {
			t.Fatalf("canale %d: %+v, atteso %+v", i, after[i], want)
		}
	}
	if !strings.Contains(verifyNotes(st), "accorciato") || !strings.Contains(verifyNotes(st), "Inoltrati 5 canali") {
		t.Fatalf("segnalazioni: %s", verifyNotes(st))
	}
	t.Logf("%s", verifyNotes(st))
}

func TestForwardBajetonToK67Replace(t *testing.T) {
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "5RH-PRO"}
	withRebootingEmu(t, k67.serve, 1<<30, 0, false)
	c := newAPIClient(t)
	before := readVia(t, c, "k67", "COM3")
	src := readDATChans(t)
	var send []*Chan
	for _, ch := range src {
		if ch != nil && len(send) < 4 {
			send = append(send, ch)
		}
	}
	// un canale digitale non si può scrivere sul K67: escluso con una segnalazione
	dig := *send[0]
	dig.Dig, dig.Name = true, "DIGITALE"
	send = append(send, &dig)
	c.must("cable/start", map[string]interface{}{"op": "forward", "radio": "k67", "port": "COM3", "channels": send, "mode": "replace"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("inoltro: %v / %s", st["error"], verifyNotes(st))
	}
	res := st["result"].(map[string]interface{})
	if res["verified"] != true || int(res["count"].(float64)) != 4 || int(res["skipped"].(float64)) != 1 {
		t.Fatalf("risultato: %v / %s", res, verifyNotes(st))
	}
	after := readVia(t, c, "k67", "COM3")
	used := 0
	for i, ch := range after {
		if ch == nil {
			continue
		}
		used++
		if i >= 4 {
			t.Fatalf("canale %d non cancellato", i)
		}
		if ch.Rx != send[i].Rx || ch.Tx != send[i].Tx {
			t.Fatalf("canale %d: %+v", i, ch)
		}
	}
	if used != 4 || countChans(before) <= 4 {
		t.Fatalf("canali dopo la sostituzione: %d (prima %d)", used, countChans(before))
	}
	if !strings.Contains(verifyNotes(st), "canale digitale") {
		t.Fatalf("segnalazioni: %s", verifyNotes(st))
	}
}

func readDATChans(t *testing.T) []*Chan {
	t.Helper()
	chs, err := readDAT(mustRead(t, sampleDAT), &Report{})
	if err != nil {
		t.Fatal(err)
	}
	return chs
}

// posizioni libere insufficienti: errore prima della scrittura, radio invariata
func TestForwardNoRoomWritesNothing(t *testing.T) {
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "5RH-PRO"}
	withRebootingEmu(t, k67.serve, 1<<30, 0, false)
	c := newAPIClient(t)
	before := readVia(t, c, "k67", "COM3")
	free := 0
	for _, ch := range before {
		if ch == nil {
			free++
		}
	}
	memBefore := append([]byte{}, k67.mem...)
	src := someSource(t, 3)
	var many []*Chan
	for len(many) < free+1 {
		many = append(many, src[len(many)%len(src)])
	}
	c.must("cable/start", map[string]interface{}{"op": "forward", "radio": "k67", "port": "COM3", "channels": many, "mode": "free"})
	st, _ := c.wait("")
	if st["state"] != "error" || !strings.Contains(st["error"].(string), "posizioni libere") || !strings.Contains(st["error"].(string), "Nulla è stato scritto") {
		t.Fatalf("atteso errore: %v %v", st["state"], st["error"])
	}
	if k67.writes != 0 || !bytes.Equal(memBefore, k67.mem) {
		t.Fatalf("la radio è stata modificata (%d blocchi scritti)", k67.writes)
	}
}

// adopt: la scheda della radio di destinazione non è aperta, i dati scritti diventano
// i suoi; altrimenti il file aperto nella scheda resta com'era
func TestForwardAdoptSession(t *testing.T) {
	for _, adopt := range []bool{false, true} {
		radio := newBtEmu(btSampleImage(t))
		withRebootingEmu(t, radio.serve, 1<<30, 0, false)
		c := newAPIClient(t)
		c.must("radio/new", map[string]interface{}{"radio": "baj"})
		c.app.mu.Lock()
		old := c.app.radios[kindDAT]
		c.app.mu.Unlock()
		c.must("cable/start", map[string]interface{}{"op": "forward", "radio": "baj", "port": "COM4", "channels": someSource(t, 2), "mode": "free", "adopt": adopt})
		st, _ := c.wait("")
		if st["state"] != "done" {
			t.Fatalf("inoltro: %v", st["error"])
		}
		c.app.mu.Lock()
		now := c.app.radios[kindDAT]
		c.app.mu.Unlock()
		if adopt == (now == old) {
			t.Fatalf("adopt %v: sessione %v", adopt, now == old)
		}
		if adopt {
			_, chs, _ := readRadio(now.Base, &Report{})
			rr := st["result"].(map[string]interface{})["radio"].(map[string]interface{})
			if !chansEqual(chs, apiChans(t, rr["channels"])) {
				t.Fatal("canali della sessione diversi da quelli restituiti")
			}
		}
	}
}

// radio dei driver (UV-5R): posizioni libere, scrittura e verifica con il protocollo di CHIRP
func TestForwardToDriverRadio(t *testing.T) {
	m := genByID["bf-uv5r"]
	mem := identModelBase(t, m)
	if mem == nil {
		t.Skip("CHIRP_DIR non impostata: manca la memoria di prova della UV-5R")
	}
	emu := genEmuFor(m, mem)
	withRebootingEmu(t, func(p Port) error { emu.serve(p); return nil }, 1<<30, 0, false)
	c := newAPIClient(t)
	before := readVia(t, c, m.ID, "COM7")
	src := someSource(t, 3)
	c.must("cable/start", map[string]interface{}{"op": "forward", "radio": m.ID, "port": "COM7", "channels": src, "mode": "free"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("inoltro: %v / %s", st["error"], verifyNotes(st))
	}
	res := st["result"].(map[string]interface{})
	if res["verified"] != true || int(res["count"].(float64)) != 3 {
		t.Fatalf("risultato %v / %s", res, verifyNotes(st))
	}
	after := readVia(t, c, m.ID, "COM7")
	changed := 0
	for i := range after {
		if !chanEqual(after[i], before[i]) {
			changed++
			if before[i] != nil {
				t.Fatalf("canale %d già presente sovrascritto", i)
			}
		}
	}
	if changed != 3 {
		t.Fatalf("canali cambiati %d", changed)
	}
	t.Logf("%s", verifyNotes(st))
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func chansEqual(a, b []*Chan) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !chanEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}
