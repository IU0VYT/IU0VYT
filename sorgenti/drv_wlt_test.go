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
	"strings"
	"testing"
	"time"
)

// Aggancio della famiglia Retevis C6 / C66 / Radioddity Echo: pacchetti calcolati
// a mano dal codice del programma del costruttore (SendHandshakeData,
// GetDataByRand, IsHandshakeSuccess).
func TestWltHello(t *testing.T) {
	key := []byte("XMU=E8\xFF\xFF")
	for _, tc := range []struct {
		r1, r2    byte
		pkt, want []byte
	}{
		// r2 = 0x2B: r1 dopo i primi 5 byte della chiave
		{0x12, 0x2B, []byte{0x53, 0x4A, 0x69, 0x1D, 0xAD, 0x65, 0x12, 0x78, 0x7F, 0xFF, 0x2B},
			[]byte{0x45, 0x1D, 0xC7, 0x41, 0x15, 0x15, 0x98, 0xBF, 0x7F}},
		// r2 = 0x05: r1 subito dopo 'S'
		{0x12, 0x05, []byte{0x53, 0x12, 0x4A, 0x69, 0x1D, 0xAD, 0x65, 0x78, 0x7F, 0xFF, 0x05},
			[]byte{0x40, 0x18, 0xCD, 0x55, 0x3D, 0x45, 0x38, 0xFF, 0xFF}},
		// r2 = 0x38: r1 prima dell'ultimo byte della chiave
		{0x12, 0x38, []byte{0x53, 0x4A, 0x69, 0x1D, 0xAD, 0x65, 0x78, 0x7F, 0x12, 0xFF, 0x38},
			[]byte{0x47, 0x1F, 0xC3, 0x49, 0x05, 0x35, 0xD8, 0x3F, 0x7F}},
	} {
		pkt, want := wltHello(key, tc.r1, tc.r2)
		if !bytes.Equal(pkt, tc.pkt) || !bytes.Equal(want, tc.want) {
			t.Fatalf("r1=%02X r2=%02X:\n pacchetto % X\n   atteso  % X\n risposta  % X\n   attesa  % X", tc.r1, tc.r2, pkt, tc.pkt, want, tc.want)
		}
	}
	// per ogni coppia di numeri il pacchetto è lungo 11 byte e chi lo riceve
	// ritrova la chiave sapendo solo dove guardare (come fa la radio)
	for r1 := 1; r1 < 255; r1++ {
		for r2 := 0; r2 < 255; r2++ {
			pkt, want := wltHello(key, byte(r1), byte(r2))
			if len(pkt) != 11 || len(want) != 9 || pkt[0] != 'S' || pkt[10] != byte(r2) {
				t.Fatalf("r1=%d r2=%d: pacchetto % X", r1, r2, pkt)
			}
			b := pkt[1:10]
			k := r2 >> 3 & 7
			if b[k] != byte(r1) {
				t.Fatalf("r1=%d r2=%d: r1 non è alla posizione %d", r1, r2, k)
			}
			got := append(append([]byte{}, b[:k]...), b[k+1:]...)
			for i := range got {
				got[i] ^= byte(r1 << uint(i))
			}
			if !bytes.Equal(got, key) {
				t.Fatalf("r1=%d r2=%d: chiave ricostruita % X", r1, r2, got)
			}
		}
	}
}

func wltLink(m *genModel, img []byte) (*wltEmu, func()) {
	e := emuFor(m, img).(*wltEmu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	return e, func() { openPortFunc = old; _ = a.Close() }
}

func (e *wltEmu) logLen() int { e.mu.Lock(); defer e.mu.Unlock(); return len(e.log) }

func TestWltRadio(t *testing.T) {
	np := func(int, int, string) {}
	defer func(d time.Duration) { wltWait = d }(wltWait)
	wltWait = 150 * time.Millisecond
	for _, id := range []string{"rt-c6", "rt-c66", "rd-echo"} {
		m := genByID[id]
		o := wltO(m)
		if m == nil || m.Baud != 9600 || len(o.key) != 8 || o.size() != m.Slots*16+map[int]int{30: 0x630 - 0x1E0, 16: 0x330 - 0x100}[m.Slots] {
			t.Fatalf("%s: modello assente o dati errati", id)
		}
		base := int64(462562500)
		if id != "rt-c6" {
			base = 446006250
		}
		// canale come lo scrive il programma: decine di Hz in BCD dalle cifre
		// basse, toni in BCD, DCS con 0x80 (0xC0 invertito) nel secondo byte
		radio := bytes.Repeat([]byte{0xFF}, o.size())
		copy(radio[o.rng():], []byte{0x00, 0x40, 0x00, 0x52, 0x00, 0x07})
		if id == "rt-c6" {
			copy(radio[o.rng():], []byte{0x00, 0x40, 0x00, 0x48, 0x00, 0x00})
		}
		copy(radio[o.set:], []byte{3, 1, 0x40, 5, 0x0C, 0xFF, 4, 1, 2, 3})
		img := append([]byte{}, radio...)
		c := &Chan{Rx: base, Tx: base, RxTone: "67.0", TxTone: "D023N", Power: m.Powers[0].Sem, Narrow: true, Scan: true}
		if err := m.Fam.Set(m, img, 2, c); err != nil {
			t.Fatal(err)
		}
		f := []byte{0x50, 0x62, 0x25, 0x46}
		if id != "rt-c6" {
			f = []byte{0x25, 0x06, 0x60, 0x44}
		}
		// canale nuovo: 0x18 nel C6 e nel C66 (alta, in scansione), 0x08 nell'Echo; qui bassa, stretta, in scansione
		want := append(append(append([]byte{}, f...), f...), 0x70, 0x06, 0x23, 0x80, 0x14, 0x00, 0x00, 0xFF)
		if !bytes.Equal(img[32:48], want) {
			t.Fatalf("%s: canale % X\n       invece di % X", id, img[32:48], want)
		}
		if g := m.Fam.Get(m, img, 2); g == nil || g.Rx != base || g.Tx != base || g.RxTone != "67.0" || g.TxTone != "D023N" || !g.Narrow || !g.Scan || g.Power != c.Power {
			t.Fatalf("%s: riletto %+v", id, g)
		}
		if err := m.Fam.Set(m, img, 3, &Chan{Rx: base, Tx: base + 12500, RxTone: "254.1", TxTone: "D754I", Power: m.Powers[1].Sem}); err != nil {
			t.Fatal(err)
		}
		if r := img[48:64]; !bytes.Equal(r[8:12], []byte{0x41, 0x25, 0x54, 0xC7}) || r[12] != 0x08 || r[13] != 0 || r[14] != 0 {
			t.Fatalf("%s: toni e bit % X", id, r)
		}
		if g := m.Fam.Get(m, img, 3); g == nil || g.Tx != base+12500 || g.RxTone != "254.1" || g.TxTone != "D754I" || g.Narrow || g.Scan {
			t.Fatalf("%s: riletto %+v", id, g)
		}
		// gli altri bit del canale restano; con il salto di frequenza con scansione
		// (byte 14 = 1) i toni non si vedono e non si perdono finché non ne scegli uno
		copy(img[64:80], append(append(append([]byte{}, f...), f...), 0x85, 0x08, 0x85, 0x08, 0xE3, 0x15, 0x01, 0x00))
		g := m.Fam.Get(m, img, 4)
		if g == nil || g.RxTone != "" || g.TxTone != "" || g.Narrow || g.Scan || g.Power != m.Powers[0].Sem {
			t.Fatalf("%s: canale con salto di frequenza: %+v", id, g)
		}
		g.Narrow = true
		if err := m.Fam.Set(m, img, 4, g); err != nil || !bytes.Equal(img[72:80], []byte{0x85, 0x08, 0x85, 0x08, 0xE7, 0x15, 0x01, 0x00}) {
			t.Fatalf("%s: canale con salto di frequenza dopo la modifica: % X (%v)", id, img[64:80], err)
		}
		g.TxTone = "88.5"
		if err := m.Fam.Set(m, img, 4, g); err != nil || !bytes.Equal(img[72:80], []byte{0xFF, 0xFF, 0x85, 0x08, 0xE7, 0x15, 0x00, 0x00}) {
			t.Fatalf("%s: tono scelto su un canale con salto di frequenza: % X (%v)", id, img[64:80], err)
		}
		// fuori gamma, tono non valido, cancellazione
		if err := m.Fam.Set(m, img, 5, &Chan{Rx: base + 5, Tx: base + 5}); err == nil {
			t.Fatalf("%s: frequenza con i 5 Hz accettata", id)
		}
		if err := m.Fam.Set(m, img, 5, &Chan{Rx: 530000000, Tx: 530000000}); err == nil {
			t.Fatalf("%s: frequenza fuori gamma accettata", id)
		}
		if err := m.Fam.Set(m, img, 5, &Chan{Rx: base, Tx: base, RxTone: "D999N"}); err == nil {
			t.Fatalf("%s: codice DCS non valido accettato", id)
		}
		if err := m.Fam.Set(m, img, 2, nil); err != nil || m.Fam.Get(m, img, 2) != nil || !bytes.Equal(img[32:46], bytes.Repeat([]byte{0xFF}, 14)) || img[46] != 0 {
			t.Fatalf("%s: cancellazione", id)
		}

		// radio simulata: lettura
		for i := 0; i < m.Slots; i++ {
			if err := m.Fam.Set(m, radio, i, &Chan{Rx: base + int64(i)*12500, Tx: base + int64(i)*12500, Power: m.Powers[1].Sem, Scan: true}); err != nil {
				t.Fatal(err)
			}
		}
		for a := m.Slots * 16; a < o.set; a++ {
			radio[a] = byte(a * 7) // codici del salto di frequenza: dati qualsiasi
		}
		e, done := wltLink(m, radio)
		e.noise = 2 // i primi due agganci falliscono: il terzo deve riuscire
		got, info, err := m.Fam.Download(context.Background(), m, "EMU", np)
		done()
		if err != nil {
			t.Fatalf("%s: lettura: %v", id, err)
		}
		if !bytes.Equal(got, radio) || !strings.Contains(info, "v1.2.3") {
			t.Fatalf("%s: lettura diversa (%s): %s", id, info, diffRanges(radio, got))
		}
		n := 3 + 1 + o.size()/16 + 1 // tre agganci, gamma, blocchi, fine
		waitFor(func() bool { return e.logLen() == n })
		if len(e.log) != n || e.log[0] != "S" || e.log[2] != "S" || e.log[3] != "R"+strings.ToUpper(hex4(o.rng())) || e.log[4] != "R0000" || e.log[n-1] != "E" {
			t.Fatalf("%s: comandi %v", id, e.log)
		}

		// modifica e scrittura
		if err := m.Fam.Set(m, got, 0, &Chan{Rx: base + 25000, Tx: base + 25000, TxTone: "D023I", Power: m.Powers[0].Sem, Narrow: true}); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, got, 5, nil); err != nil {
			t.Fatal(err)
		}
		got[o.set] = 7        // squelch
		got[o.pwd()+6] = 0x77 // un byte del blocco della password: non deve arrivare alla radio
		e, done = wltLink(m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", got, np)
		done()
		if err != nil {
			t.Fatalf("%s: scrittura: %v", id, err)
		}
		n = 1 + 1 + o.pwd()/16 + 1
		waitFor(func() bool { return e.logLen() == n })
		after := e.image()
		if c := m.Fam.Get(m, after, 0); c == nil || c.Rx != base+25000 || c.TxTone != "D023I" || !c.Narrow || c.Scan {
			t.Fatalf("%s: canale 1 dopo la scrittura: %+v", id, c)
		}
		if m.Fam.Get(m, after, 5) != nil || after[o.set] != 7 {
			t.Fatalf("%s: canale 6 o squelch non scritti", id)
		}
		if !bytes.Equal(after[o.pwd():], radio[o.pwd():]) {
			t.Fatalf("%s: il blocco della password è stato scritto", id)
		}
		if !bytes.Equal(after[:o.pwd()], got[:o.pwd()]) {
			t.Fatalf("%s: memoria dopo la scrittura: %s", id, diffRanges(got[:o.pwd()], after[:o.pwd()]))
		}
		for _, l := range e.log {
			if strings.HasPrefix(l, "W") && l >= "W"+strings.ToUpper(hex4(o.pwd())) {
				t.Fatalf("%s: scrittura oltre le impostazioni: %s", id, l)
			}
		}
		if len(e.log) != n || e.log[n-1] != "E" {
			t.Fatalf("%s: comandi della scrittura %v", id, e.log)
		}

		// immagine di un'altra radio (area diversa): si ferma prima di scrivere
		other := append([]byte{}, got...)
		other[o.rng()+5] ^= 1
		e, done = wltLink(m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", other, np)
		done()
		if err == nil || !strings.Contains(err.Error(), "un'altra radio") {
			t.Fatalf("%s: immagine di un'altra radio: %v", id, err)
		}
		for _, l := range e.log {
			if strings.HasPrefix(l, "W") {
				t.Fatalf("%s: ha scritto con l'immagine di un'altra radio", id)
			}
		}
		// immagine mai letta da una radio
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", bytes.Repeat([]byte{0xFF}, o.size()), np); err == nil {
			t.Fatalf("%s: immagine vuota accettata", id)
		}

		// radio con la password: la lettura si ferma e lo dice
		locked := append([]byte{}, radio...)
		copy(locked[o.pwd():], "123456")
		_, done = wltLink(m, locked)
		_, _, err = m.Fam.Download(context.Background(), m, "EMU", np)
		done()
		if err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatalf("%s: radio con password: %v", id, err)
		}
		// tre agganci falliti: errore
		e, done = wltLink(m, radio)
		e.noise = 3
		_, _, err = m.Fam.Download(context.Background(), m, "EMU", np)
		done()
		if err == nil {
			t.Fatalf("%s: lettura riuscita dopo tre agganci falliti", id)
		}
	}
	// due canali predefiniti del programma del C6 (campo defaultData): il canale 8
	// (467,5625 MHz, D243N, bassa, stretta, in scansione) e il 23 (ponte: ricezione
	// 462,550, trasmissione 467,550, senza toni, alta, larga)
	c6 := genByID["rt-c6"]
	fac := bytes.Repeat([]byte{0xFF}, wltO(c6).size())
	copy(fac[7*16:], []byte{0x50, 0x62, 0x75, 0x46, 0x50, 0x62, 0x75, 0x46, 0x43, 0x82, 0x43, 0x82, 0x14, 0, 0, 0})
	copy(fac[22*16:], []byte{0x00, 0x50, 0x25, 0x46, 0x00, 0x50, 0x75, 0x46, 0xFF, 0xFF, 0xFF, 0xFF, 0x18, 0, 0, 0})
	if c := c6.Fam.Get(c6, fac, 7); c == nil || c.Rx != 467562500 || c.Tx != 467562500 || c.RxTone != "D243N" || c.TxTone != "D243N" || !c.Narrow || !c.Scan || c.Power != c6.Powers[0].Sem {
		t.Fatalf("canale 8 di fabbrica del C6: %+v", c)
	} else if err := c6.Fam.Set(c6, fac, 7, c); err != nil || fac[7*16+12] != 0x14 {
		t.Fatalf("canale 8 di fabbrica del C6 riscritto: %v, %02X", err, fac[7*16+12])
	}
	if c := c6.Fam.Get(c6, fac, 22); c == nil || c.Rx != 462550000 || c.Tx != 467550000 || c.RxTone != "" || c.TxTone != "" || c.Narrow || !c.Scan || c.Power != c6.Powers[1].Sem {
		t.Fatalf("canale 23 di fabbrica del C6: %+v", c)
	}
	// scheda sbagliata: la radio non risponde alla chiave di un altro modello
	m, other := genByID["rt-c6"], genByID["rt-c66"]
	radio := bytes.Repeat([]byte{0xFF}, wltO(other).size())
	e := emuFor(other, radio).(*wltEmu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	_, _, err := m.Fam.Download(context.Background(), m, "EMU", np)
	openPortFunc = old
	_ = a.Close()
	if err == nil {
		t.Fatal("la scheda del C6 ha letto un C66")
	}
}

func hex4(v int) string {
	const d = "0123456789abcdef"
	return string([]byte{d[v>>12&15], d[v>>8&15], d[v>>4&15], d[v&15]})
}
