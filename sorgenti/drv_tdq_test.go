//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func tdqLink(m *genModel, img []byte) (*tdqEmu, func()) {
	e := emuFor(m, img).(*tdqEmu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	return e, func() { openPortFunc = old; _ = a.Close() }
}

func (e *tdqEmu) logCopy() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...)
}

// Toni calcolati a mano come li scrive il programma: 67.0 Hz = 670 = 0x029E → 9E 02;
// D023N = 0x2800 + 023 ottale (19) = 0x2813 → 13 28; D754I = 0xA800 + 754 ottale (492 = 0x1EC) → EC A9.
func TestTdqTones(t *testing.T) {
	for s, want := range map[string][2]byte{
		"67.0": {0x9E, 0x02}, "250.3": {0xC7, 0x09}, "100.0": {0xE8, 0x03}, "D023N": {0x13, 0x28}, "D411N": {0x09, 0x29},
		"D754N": {0xEC, 0x29}, "D023I": {0x13, 0xA8}, "D754I": {0xEC, 0xA9}, "": {0x00, 0x00},
	} {
		var b [2]byte
		if err := tdqToneEnc(b[:], s, 0x00); err != nil || b != want {
			t.Fatalf("%q: % X invece di % X (%v)", s, b, want, err)
		}
		if got := tdqToneDec(want[:]); got != s {
			t.Fatalf("% X: %q invece di %q", want, got, s)
		}
	}
	var b [2]byte
	if err := tdqToneEnc(b[:], "", 0xFF); err != nil || b != [2]byte{0xFF, 0xFF} { // TD-M2 Pro
		t.Fatalf("nessun tono (M2 Pro): % X (%v)", b, err)
	}
	for _, raw := range [][2]byte{{0xFF, 0xFF}, {0x00, 0x28}, {0xF1, 0x28}, {0xFF, 0x02}, {0x05, 0x00}} {
		if got := tdqToneDec(raw[:]); got != "" {
			t.Fatalf("% X letto come %q", raw, got)
		}
	}
	for _, bad := range []string{"D889N", "30.0"} {
		if err := tdqToneEnc(b[:], bad, 0); err == nil {
			t.Fatalf("%s accettato", bad)
		}
	}
}

func TestTdqChannel(t *testing.T) {
	f := tdqFam{}
	pmr := genByID["td-pmr"]
	img := bytes.Repeat([]byte{0xFF}, tdqSize)
	// canale nuovo: 446,00625 MHz = 44600625 (10 Hz) → 25 06 60 44; stretta, bassa, in scansione
	if err := f.Set(pmr, img, 0, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Narrow: true, Scan: true}); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x25, 0x06, 0x60, 0x44, 0x25, 0x06, 0x60, 0x44, 0, 0, 0, 0, 0x0E, 0x41, 0xFF, 0xFF}
	if !bytes.Equal(img[:16], want) {
		t.Fatalf("canale nuovo: % X\n          atteso % X", img[:16], want)
	}
	c := f.Get(pmr, img, 0)
	if c == nil || c.Rx != 446006250 || c.Tx != 446006250 || c.TxOff || c.Power != 2 || !c.Narrow || !c.Scan || c.RxTone != "" || c.TxTone != "" {
		t.Fatalf("riletto %+v", c)
	}
	if f.Get(pmr, img, 1) != nil {
		t.Fatal("canale vuoto letto come pieno")
	}
	// alta su un modello a due livelli = 2 nella radio; larga
	if err := f.Set(pmr, img, 1, &Chan{Rx: 446018750, Tx: 446018750, Power: 0, TxTone: "D754I"}); err != nil {
		t.Fatal(err)
	}
	if got := img[16:32]; got[12] != 0x8F || got[13] != 0x01 || !bytes.Equal(got[8:12], []byte{0, 0, 0xEC, 0xA9}) {
		t.Fatalf("secondo canale: % X", got)
	}
	// canale scritto dal programma: 462,5625 MHz solo ricezione, 67.0 / D023N, media, funzione scrambler,
	// scrambler 4, larga, blocco su occupato, salto, PTT-ID 2, modo 2, codici 5 e 6
	m11 := genByID["td-m11plus"]
	rec := []byte{0x50, 0x62, 0x25, 0x46, 0xFF, 0xFF, 0xFF, 0xFF, 0x9E, 0x02, 0x13, 0x28, 0x57, 0x3A, 5, 6}
	copy(img[32:], rec)
	c = f.Get(m11, img, 2)
	if c == nil || c.Rx != 462562500 || !c.TxOff || c.RxTone != "67.0" || c.TxTone != "D023N" || c.Power != 1 || c.Narrow || c.Scan {
		t.Fatalf("canale del programma: %+v", c)
	}
	if c2 := f.Get(genByID["td-frs"], img, 2); c2 == nil || c2.Power != 2 { // «media» su un modello che non l'ha: primo livello
		t.Fatalf("potenza non prevista: %+v", c2)
	}
	c.Tx, c.TxOff, c.Power, c.Narrow, c.Scan, c.RxTone = 467562500, false, 0, true, true, ""
	if err := f.Set(m11, img, 2, c); err != nil {
		t.Fatal(err)
	}
	want = []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x75, 0x46, 0, 0, 0x13, 0x28, 0x96, 0x7A, 5, 6}
	if !bytes.Equal(img[32:48], want) {
		t.Fatalf("canale modificato: % X\n               atteso % X", img[32:48], want)
	}
	if err := f.Set(m11, img, 2, nil); err != nil || !bytes.Equal(img[32:48], bytes.Repeat([]byte{0xFF}, 16)) {
		t.Fatalf("canale cancellato: % X (%v)", img[32:48], err)
	}
	// frequenze rifiutate
	for _, bad := range []*Chan{{Rx: 300000000, Tx: 300000000}, {Rx: 446001000, Tx: 446001000}, {Rx: 446006250, Tx: 175000000}} {
		if err := f.Set(m11, img, 3, bad); err == nil {
			t.Errorf("%d/%d accettato", bad.Rx, bad.Tx)
		}
	}
	if err := f.Set(genByID["td-m2pro"], img, 3, &Chan{Rx: 145500000, Tx: 145500000}); err == nil {
		t.Error("TD-M2 Pro: VHF accettata")
	}
	if err := f.Set(m11, img, 3, &Chan{Rx: 145500000, Tx: 145500000, TxOff: true}); err != nil || !bytes.Equal(img[48:56], []byte{0, 0, 0x55, 0x14, 0xFF, 0xFF, 0xFF, 0xFF}) {
		t.Errorf("145,5 MHz solo ricezione: % X (%v)", img[48:56], err)
	}
}

func TestTdqModels(t *testing.T) {
	for id, w := range map[string]struct {
		slots, pw int
		magic     string
	}{"td-712plus": {199, 3, "QROGRAM"}, "td-m7": {199, 3, "QROGRAM"}, "td-m11plus": {199, 3, "QROGRAM"}, "td-m11plus-16": {16, 3, "QROGRAM"},
		"td-m11plus-22": {22, 2, "QROGRAM"}, "td-br09": {199, 3, "QROGRAM"}, "td-777s-16": {16, 3, "QROGRAM"}, "td-777s-22": {22, 3, "QROGRAM"},
		"td-v900": {199, 3, "QROGRAM"}, "td-frs": {22, 2, "QROGRAM"}, "td-pmr": {16, 2, "QROGRAM"}, "td-m2pro": {199, 3, "QROGRAO"}} {
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := tdqO(m)
		if m.Slots != w.slots || len(m.Powers) != w.pw || len(o.pow) != w.pw || o.magic != w.magic || m.Baud != 38400 || !m.NoChirp || m.NameMax != 0 {
			t.Errorf("%s: dati del modello errati", id)
		}
		if id == "td-m2pro" != (len(m.Bands) == 1) || o.pow[len(o.pow)-1] != 2 {
			t.Errorf("%s: bande o potenze errate", id)
		}
	}
}

func TestTdqRadio(t *testing.T) {
	np := func(int, int, string) {}
	defer func(d time.Duration) { tdqWait = d }(tdqWait)
	tdqWait = 200 * time.Millisecond
	ctx := context.Background()
	for _, id := range []string{"td-m11plus", "td-m2pro"} {
		m := genByID[id]
		f := tdqFam{}
		base := bytes.Repeat([]byte{0xFF}, tdqSize)
		if err := f.Set(m, base, 0, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Narrow: true}); err != nil {
			t.Fatal(err)
		}
		e, done := tdqLink(m, base)
		got, _, err := f.Download(ctx, m, "COM7", np)
		for i := 0; i < 100 && err == nil && len(e.logCopy()) < 60; i++ { // il 45 finale arriva dopo l'ultima risposta
			time.Sleep(10 * time.Millisecond)
		}
		done()
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if !bytes.Equal(got, e.image()) || got[tdqSetAddr] != 0x52 || got[0xCC0+7] != 1 {
			t.Fatalf("%s: lettura diversa dalla memoria della radio", id)
		}
		lg := e.logCopy()
		if len(lg) != 60 || lg[0] != "chiave" || lg[1] != "R0000" || lg[2] != "R0040" || lg[58] != "R0E40" || lg[59] != "fine" {
			t.Fatalf("%s: sequenza di lettura inattesa (%d passi): %v", id, len(lg), lg[:3])
		}
		// modifica e scrittura: 232 righe da 16 byte, da 0x000 a 0xE70, senza byte di chiusura
		if err := f.Set(m, got, 1, &Chan{Rx: 446018750, Tx: 446018750, Power: 0, TxTone: "88.5"}); err != nil {
			t.Fatal(err)
		}
		got[tdqSetAddr+1] = 7                // squelch
		copy(got[0xCC0+7:], []byte{9, 9, 9}) // la versione dell'hardware non deve cambiare nella radio
		e2, done := tdqLink(m, e.image())
		_, err = f.Upload(ctx, m, "COM7", got, np)
		done()
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		lg = e2.logCopy()
		if len(lg) != 233 || lg[1] != "W0000" || lg[2] != "W0010" || lg[232] != "W0E70" {
			t.Fatalf("%s: sequenza di scrittura inattesa (%d passi)", id, len(lg))
		}
		after := e2.image()
		if !bytes.Equal(after[:0xCC0], got[:0xCC0]) || after[tdqSetAddr+1] != 7 || !bytes.Equal(after[0xCC0+7:0xCC0+10], []byte{1, 0, 8}) {
			t.Fatalf("%s: memoria della radio diversa da quella scritta", id)
		}
		// immagine mai letta da una radio: si scrivono solo i canali (200 righe)
		e3, done := tdqLink(m, e.image())
		_, err = f.Upload(ctx, m, "COM7", base, np)
		done()
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if lg = e3.logCopy(); len(lg) != 201 || lg[200] != "W0C70" {
			t.Fatalf("%s: immagine vuota, %d passi", id, len(lg))
		}
		if after = e3.image(); after[tdqSetAddr] != 0x52 || !bytes.Equal(after[:tdqChans], base[:tdqChans]) {
			t.Fatalf("%s: immagine vuota, impostazioni toccate", id)
		}
		// risposta sbagliata alla chiave
		e4, done := tdqLink(m, base)
		e4.ack07 = 0x06
		_, _, err = f.Download(ctx, m, "COM7", np)
		done()
		if err == nil {
			t.Fatalf("%s: risposta 06 alla chiave accettata", id)
		}
		// immagine troppo corta
		if _, err := f.Upload(ctx, m, "COM7", base[:100], np); err == nil {
			t.Fatalf("%s: immagine corta accettata", id)
		}
	}
	// la radio di un'altra chiave non risponde
	_, done := tdqLink(genByID["td-m2pro"], nil)
	_, _, err := tdqFam{}.Download(ctx, genByID["td-frs"], "COM7", np)
	done()
	if err == nil {
		t.Fatal("chiave di un altro modello accettata")
	}
}
