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
	"fmt"
	"strings"
	"testing"
	"time"
)

func wfLink(m *genModel, img []byte) (*wfEmu, func()) {
	e := emuFor(m, img).(*wfEmu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	return e, func() { openPortFunc = old; _ = a.Close() }
}

func (e *wfEmu) logLen() int { e.mu.Lock(); defer e.mu.Unlock(); return len(e.log) }

// Toni come nella tabella «ctctb» del file di configurazione del costruttore:
// 67.0 = 02 9E, 250.3 = 09 C7, D023N = 28 13, D411N = 29 09, D754N = 29 EC,
// D023I = A8 13, D754I = A9 EC, nessuno = FF FF.
func TestWfTones(t *testing.T) {
	for s, want := range map[string][2]byte{
		"67.0": {0x02, 0x9E}, "250.3": {0x09, 0xC7}, "100.0": {0x03, 0xE8}, "D023N": {0x28, 0x13}, "D411N": {0x29, 0x09},
		"D754N": {0x29, 0xEC}, "D023I": {0xA8, 0x13}, "D754I": {0xA9, 0xEC}, "D265I": {0xA8, 0xB5}, "": {0xFF, 0xFF},
	} {
		var b [2]byte
		if err := wfToneEnc(b[:], s); err != nil || b != want {
			t.Fatalf("%q: % X invece di % X (%v)", s, b, want, err)
		}
		if got := wfToneDec(want[:]); got != s {
			t.Fatalf("% X: %q invece di %q", want, got, s)
		}
	}
	// in lettura per il DCS basta il bit 0x2000 (come nel CHIRP); 00 00 e valori senza senso = nessun tono
	for want, raw := range map[string][2]byte{"D023N": {0x20, 0x13}, "D754I": {0xA1, 0xEC}, "": {0x00, 0x00}} {
		if got := wfToneDec(raw[:]); got != want {
			t.Fatalf("% X: %q invece di %q", raw, got, want)
		}
	}
	if got := wfToneDec([]byte{0x10, 0x00}); got != "" {
		t.Fatalf("valore senza senso letto come %q", got)
	}
	var b [2]byte
	for _, bad := range []string{"D889N", "30.0"} {
		if err := wfToneEnc(b[:], bad); err == nil {
			t.Fatalf("%s accettato", bad)
		}
	}
}

func TestWfRadio(t *testing.T) {
	np := func(int, int, string) {}
	defer func(d time.Duration) { wfWait = d }(wfWait)
	wfWait = 200 * time.Millisecond
	for _, tc := range []struct {
		id         string
		slots      int
		page, baud int
		start      int
		kind       []byte
	}{
		{"rd-ub15g", 88, 16, 115200, 0, []byte{0, 6, 2, 0}}, {"rd-ub15g-110", 88, 24, 115200, 0, []byte{0, 6, 2, 0}},
		{"rd-db35g", 80, 16, 115200, 0, []byte{0, 6, 2, 0}}, {"rd-db35uv", 80, 16, 115200, 0, []byte{0, 6, 2, 0}},
		{"rd-db40g", 199, 32, 19200, 0x10, []byte{0, 8, 9, 3}},
	} {
		id := tc.id
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := wfO(m)
		if m.Slots != tc.slots || o.page != tc.page || m.Baud != tc.baud || o.start != tc.start || !bytes.Equal(o.ver, tc.kind) || o.size(m) != (tc.slots+1)*tc.page {
			t.Fatalf("%s: dati del modello errati", id)
		}
		// 462,5625 MHz = 462562500 Hz = 1B 92 24 C4 (prima il byte alto), come nel file del costruttore
		img := make([]byte, o.size(m))
		c := &Chan{Rx: 462562500, Tx: 467562500, RxTone: "67.0", TxTone: "D023N", Power: m.Powers[len(m.Powers)-1].Sem, Scan: true, Name: "ponte 1"}
		if err := m.Fam.Set(m, img, 1, c); err != nil {
			t.Fatal(err)
		}
		r := img[o.page : 2*o.page]
		if !bytes.Equal(r[:12], []byte{0x1B, 0x92, 0x24, 0xC4, 0x1B, 0xDE, 0x70, 0x04, 0x02, 0x9E, 0x28, 0x13}) {
			t.Fatalf("%s: frequenze e toni % X", id, r[:12])
		}
		switch id {
		case "rd-db40g": // banda nel byte 12 (1 = larga), potenza nel byte 13 (2 = alta), nome dal byte 14 riempito con 00 (come il CHIRP)
			if r[12] != 0x01 || r[13] != 0x02 || string(r[14:22]) != "PONTE 1\x00" {
				t.Fatalf("%s: bit e nome % X", id, r[12:22])
			}
		case "rd-ub15g-110": // potenza bit 0-1 (1 = alta), larga bit 2, scansione bit 4; nome dal byte 13
			if r[12] != 0x15 || string(r[13:21]) != "PONTE 1 " {
				t.Fatalf("%s: bit e nome % X", id, r[12:21])
			}
		default:
			if r[12] != 0x15 || !bytes.Equal(r[13:], []byte{0, 0, 0}) {
				t.Fatalf("%s: bit % X", id, r[12:])
			}
		}
		g := m.Fam.Get(m, img, 1)
		if g == nil || g.Rx != c.Rx || g.Tx != c.Tx || g.RxTone != "67.0" || g.TxTone != "D023N" || g.Power != c.Power || g.Narrow || g.TxOff ||
			(o.scan != nil && !g.Scan) || (o.name >= 0 && g.Name != "PONTE 1") {
			t.Fatalf("%s: riletto %+v", id, g)
		}
		// solo ricezione (trasmissione a zero), stretta, bassa; i bit che non si gestiscono restano
		// (nella DB40-G banda e potenza sono byte interi, come nel CHIRP)
		r[12] |= 0x60
		g.TxOff, g.Narrow, g.Power, g.Scan = true, true, m.Powers[0].Sem, false
		if err := m.Fam.Set(m, img, 1, g); err != nil || !bytes.Equal(r[4:8], []byte{0, 0, 0, 0}) {
			t.Fatalf("%s: solo ricezione: % X (%v)", id, r[:12], err)
		}
		if id == "rd-db40g" {
			if r[12] != 0 || r[13] != 0 {
				t.Fatalf("%s: bit dopo la modifica % X", id, r[12:14])
			}
		} else if r[12] != 0x60 {
			t.Fatalf("%s: bit dopo la modifica %02X", id, r[12])
		}
		if g2 := m.Fam.Get(m, img, 1); g2 == nil || !g2.TxOff || g2.Tx != g2.Rx || !g2.Narrow || g2.Power != m.Powers[0].Sem {
			t.Fatalf("%s: solo ricezione riletta %+v", id, g2)
		}
		// errori e cancellazione (riga vuota come la scrive il programma)
		if err := m.Fam.Set(m, img, 2, &Chan{Rx: 300000000, Tx: 300000000}); err == nil {
			t.Fatalf("%s: frequenza fuori banda accettata", id)
		}
		if o.name >= 0 {
			if err := m.Fam.Set(m, img, 2, &Chan{Rx: 462562500, Tx: 462562500, Name: "nometroppolungo"}); err == nil {
				t.Fatalf("%s: nome lungo accettato", id)
			}
		}
		empty := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF}
		if o.chirp { // riga cancellata tutta a zero, come il CHIRP
			empty = make([]byte, 12)
		}
		if err := m.Fam.Set(m, img, 1, nil); err != nil || m.Fam.Get(m, img, 1) != nil || !bytes.Equal(r[:12], empty) {
			t.Fatalf("%s: cancellazione % X", id, r)
		}
		if o.name >= 0 && !o.chirp && string(r[o.name:o.name+8]) != "        " {
			t.Fatalf("%s: nome della riga cancellata % X", id, r[o.name:o.name+8])
		}

		// radio simulata
		radio := make([]byte, o.size(m))
		for i := 0; i < m.Slots; i += 3 {
			if err := m.Fam.Set(m, radio, i, &Chan{Rx: 462550000 + int64(i)*12500, Tx: 462550000 + int64(i)*12500, Power: m.Powers[0].Sem, Scan: true}); err != nil {
				t.Fatal(err)
			}
		}
		copy(radio[o.param(m):], []byte{0x23, 0x2A, 0x51, 0x05})
		if id == "rd-db40g" {
			copy(radio[o.param(m):], []byte{1, 0, 4, 7})
		}
		e, done := wfLink(m, radio)
		got, _, err := m.Fam.Download(context.Background(), m, "EMU", np)
		done()
		if err != nil || !bytes.Equal(got, radio) {
			t.Fatalf("%s: lettura: %v %s", id, err, diffRanges(radio, got))
		}
		// comandi: domanda del tipo (non per la DB40-G, come il CHIRP), una lettura per riga, impostazioni, fine
		n, first := 1+m.Slots+1+1, []string{"tipo"}
		if o.chirp {
			n, first = n-1, nil
		}
		first = append(first, fmt.Sprintf("R%04X", o.start), fmt.Sprintf("R%04X", o.start+16))
		waitFor(func() bool { return e.logLen() == n })
		if len(e.log) != n || strings.Join(e.log[:len(first)], " ") != strings.Join(first, " ") || e.log[n-2] != "RE000" || e.log[n-1] != "fine" {
			t.Fatalf("%s: comandi %v", id, e.log[:4])
		}
		if err := m.Fam.Set(m, got, 0, &Chan{Rx: 462700000, Tx: 467700000, TxTone: "88.5", Power: m.Powers[len(m.Powers)-1].Sem}); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, got, 3, nil); err != nil {
			t.Fatal(err)
		}
		got[o.param(m)+1] ^= 0x08
		e, done = wfLink(m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", got, np)
		done()
		if err != nil {
			t.Fatalf("%s: scrittura: %v", id, err)
		}
		waitFor(func() bool { return e.logLen() == n })
		if after := e.image(); !bytes.Equal(after, got) || len(e.log) != n || e.log[n-2] != "WE000" {
			t.Fatalf("%s: memoria dopo la scrittura: %s", id, diffRanges(got, e.image()))
		}
		// immagine mai letta (impostazioni vuote): si scrivono solo i canali
		blank := make([]byte, o.size(m))
		_ = m.Fam.Set(m, blank, 0, &Chan{Rx: 462700000, Tx: 462700000, Power: m.Powers[0].Sem})
		e, done = wfLink(m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", blank, np)
		done()
		waitFor(func() bool { return e.logLen() == n-1 })
		if after := e.image(); err != nil || !bytes.Equal(after[o.param(m):], radio[o.param(m):]) || !bytes.Equal(after[:o.param(m)], blank[:o.param(m)]) {
			t.Fatalf("%s: immagine senza impostazioni: %v", id, err)
		}
		if o.chirp {
			// risposta alla lettura con un'intestazione diversa dall'eco del comando: errore
			e, done = wfLink(m, radio)
			e.badEcho = true
			_, _, err = m.Fam.Download(context.Background(), m, "EMU", np)
			done()
			if err == nil || !strings.Contains(err.Error(), "risposta inattesa") {
				t.Fatalf("%s: eco errata accettata: %v", id, err)
			}
			continue
		}
		// radio di un altro tipo: si ferma prima dell'aggancio e dice quale tipo ha trovato
		e, done = wfLink(m, radio)
		e.kind = []byte{0, 0, 0, 5}
		_, _, err = m.Fam.Download(context.Background(), m, "EMU", np)
		done()
		if err == nil || !strings.Contains(err.Error(), "H5") || len(e.log) != 1 {
			t.Fatalf("%s: radio di un altro tipo: %v %v", id, err, e.log)
		}
	}
	// scheda sbagliata fra UB15-G e DB40-G: tipo diverso
	a, b := genByID["rd-ub15g"], genByID["rd-db40g"]
	_, done := wfLink(b, make([]byte, wfO(b).size(b)))
	_, _, err := a.Fam.Download(context.Background(), a, "EMU", np)
	done()
	if err == nil || !strings.Contains(err.Error(), "DB40-G") {
		t.Fatalf("scheda UB15-G con una DB40-G: %v", err)
	}
	// nomi: maiuscole, e in lettura si tolgono sia gli spazi sia gli 00 finali
	for _, id := range []string{"rd-ub15g-110", "rd-db40g"} {
		m := genByID[id]
		o := wfO(m)
		img := make([]byte, o.size(m))
		if m.cleanName("ponte x") != "PONTE X" {
			t.Fatalf("%s: nome non convertito in maiuscolo", id)
		}
		_ = m.Fam.Set(m, img, 0, &Chan{Rx: 462562500, Tx: 462562500, Name: "AB"})
		copy(img[o.name+2:], "\x00 \x00\x00  ")
		if g := m.Fam.Get(m, img, 0); g == nil || g.Name != "AB" {
			t.Fatalf("%s: nome letto %q", id, g.Name)
		}
	}
}
