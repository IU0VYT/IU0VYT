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
	"testing"
)

// 4.25: verso del bit della potenza nei programmi T17-T18, GT-18 e Radioddity T3
// (scelte «LOW», «HIGH»: la seconda accende il bit 3 del byte 12 del canale)
func TestT17PowerBit(t *testing.T) {
	for _, id := range []string{"bf-t17pmr", "bf-t17frs", "bf-gt18pmr", "rd-t3", "bf-t9"} {
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		if len(m.Powers) != 2 || m.Powers[0].Label != "Bassa" || m.Powers[1].Label != "Alta" {
			t.Fatalf("%s: potenze %+v", id, m.Powers)
		}
		img := bytes.Repeat([]byte{0xFF}, t17Memsize)
		for k, want := range []byte{0x00, 0x08} { // 0 = bassa, 1 = alta
			c := &Chan{Rx: 446006250, Tx: 446006250, Power: m.Powers[k].Sem, Narrow: true, Scan: true}
			if id == "bf-t9" {
				img[0x652] = 0xF7 // gamma 400-470 MHz
			}
			if err := m.Fam.Set(m, img, 3, c); err != nil {
				t.Fatal(err)
			}
			if got := img[3*16+12] & 0x08; got != want {
				t.Fatalf("%s: potenza %s → bit 3 = %02X", id, m.Powers[k].Label, got)
			}
			if g := m.Fam.Get(m, img, 3); g == nil || g.Power != c.Power {
				t.Fatalf("%s: potenza riletta %+v", id, g)
			}
		}
		// byte scritto dal programma per un canale in potenza alta, banda stretta,
		// in scansione, senza blocco del canale occupato: 0x08 | 0x04 | 0x01
		if id != "bf-t9" && img[3*16+12]&0x9F != 0x0D {
			t.Fatalf("%s: byte 12 = %02X", id, img[3*16+12])
		}
	}
}

// Radioddity FS-T3 / PR-T3: stesse aree e stessi blocchi del programma «Radioddity T3»
func TestRadioddityT3(t *testing.T) {
	m := genByID["rd-t3"]
	o := t17o(m)
	if o.cmd != "PROGRAL" || o.pmr || o.wblk != 8 || len(o.gaps) != 1 || o.gaps[0] != [2]int{0, 0x660} || m.Slots != 99 || m.Baud != 9600 {
		t.Fatalf("opzioni: %+v", o)
	}
	for hz, ok := range map[int64]bool{446006250: true, 446200000: true, 462562500: true, 467712500: true, 433500000: false, 446300000: false, 468000000: false} {
		if m.inBands(hz) != ok {
			t.Fatalf("frequenza %d: ammessa = %v", hz, !ok)
		}
	}
	radio := bytes.Repeat([]byte{0xFF}, t17Memsize)
	copy(radio[0x630:], []byte{0x01, 0x00, 0x01, 0x00, 0x03, 0x00, 0x00, 0x00})
	copy(radio[0x640:], []byte{0x03, 0x05, 0x00, 0x06, 0x01, 0xFF, 0xFF, 0xFF})
	copy(radio[0x650:], []byte{0x00, 0x00, 0xF7, 0x33, 0x31, 0x30, 0x33, 0x50})
	for i, hz := range []int64{462562500, 462587500, 462612500} {
		if err := m.Fam.Set(m, radio, i, &Chan{Rx: hz, Tx: hz, TxTone: "67.0", RxTone: "67.0", Power: m.Powers[1].Sem, Narrow: true, Scan: true}); err != nil {
			t.Fatal(err)
		}
	}
	link := func(img []byte) (genEmu, func()) {
		e := emuFor(m, img)
		a, b := newPipe()
		old := openPortFunc
		openPortFunc = func(string, int) (Port, error) { return a, nil }
		go func() { _ = e.serve(b) }()
		return e, func() { openPortFunc = old; _ = a.Close() }
	}
	_, done := link(radio)
	blocks := 0
	got, _, err := m.Fam.Download(context.Background(), m, "EMU", func(d, tot int, s string) {
		if s == "Lettura…" {
			blocks++
		}
	})
	done()
	if err != nil {
		t.Fatal(err)
	}
	if blocks != 0x660/16 || !bytes.Equal(got, radio) {
		t.Fatalf("lettura: %d blocchi, %s", blocks, diffRanges(radio, got))
	}
	if err := m.Fam.Set(m, got, 1, nil); err != nil {
		t.Fatal(err)
	}
	if err := m.Fam.Set(m, got, 21, &Chan{Rx: 462725000, Tx: 462725000, Power: m.Powers[0].Sem, Scan: false}); err != nil {
		t.Fatal(err)
	}
	e, done := link(radio)
	blocks = 0
	_, err = m.Fam.Upload(context.Background(), m, "EMU", got, func(d, tot int, s string) {
		if s == "Scrittura…" {
			blocks++
		}
	})
	done()
	if err != nil {
		t.Fatal(err)
	}
	if blocks != 0x660/8 {
		t.Fatalf("scrittura: %d blocchi", blocks)
	}
	after := e.(interface{ image() []byte }).image()
	if !bytes.Equal(after[:t17Memsize], got) {
		t.Fatalf("radio dopo la scrittura: %s", diffRanges(got, after[:t17Memsize]))
	}
	if m.Fam.Get(m, after, 1) != nil {
		t.Fatal("canale 2 non cancellato")
	}
	// canale cancellato: 16 byte a 0xFF, come li scrive il programma
	if !bytes.Equal(after[16:32], bytes.Repeat([]byte{0xFF}, 16)) {
		t.Fatalf("canale cancellato: % X", after[16:32])
	}
	if c := m.Fam.Get(m, after, 21); c == nil || c.Rx != 462725000 || c.Scan || c.Narrow || c.Power != m.Powers[0].Sem {
		t.Fatalf("canale 22: %+v", c)
	}
	// impostazioni e identificativo restano quelli della radio
	if !bytes.Equal(after[0x630:0x660], radio[0x630:0x660]) {
		t.Fatal("impostazioni modificate")
	}
}
