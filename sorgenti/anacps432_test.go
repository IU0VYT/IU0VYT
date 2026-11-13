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
	"sync"
	"testing"
	"time"
)

// 4.32: Baofeng T22 (FRS) / TIDRADIO TD-M8S (programma «T20 FRS» v1.0.1) e
// TIDRADIO TD-V2 (programma «CPS_F8_FRS» v1.0). La radio finta qui sotto è
// scritta a parte da quella di genemu_bfd_other.go e segue handShake, readInfo
// e writeInfo dei due programmi: annota ogni comando ricevuto.
type anaFake struct {
	mu  sync.Mutex
	img []byte
	log []string
}

func (f *anaFake) note(format string, a ...interface{}) {
	f.mu.Lock()
	f.log = append(f.log, fmt.Sprintf(format, a...))
	f.mu.Unlock()
}

func (f *anaFake) serve(p Port) {
	ident := false
	for {
		c, err := emuByte(p)
		if err != nil {
			return
		}
		switch c {
		case 0x02:
			if ident {
				ident = false
				f.note("02")
				p.Write([]byte("P3107\xF7\x00\x00"))
				continue
			}
			cmd, err := emuRead(p, 7)
			if err != nil {
				return
			}
			f.note("%s", cmd)
			if string(cmd) != "PROGRAM" {
				return
			}
			p.Write([]byte{0x06})
			ident = true
		case 0x06:
			f.note("06")
			p.Write([]byte{0x06})
		case 'E':
			f.note("E")
			p.Write([]byte{'F'})
		case 'R', 'W':
			h, err := emuRead(p, 3)
			if err != nil {
				return
			}
			a, n := int(h[0])<<8|int(h[1]), int(h[2])
			f.note("%c%03X/%d", c, a, n)
			f.mu.Lock()
			if c == 'R' {
				d := append([]byte{'W', h[0], h[1], h[2]}, f.img[a:a+n]...)
				f.mu.Unlock()
				p.Write(d)
				continue
			}
			f.mu.Unlock()
			d, err := emuRead(p, n)
			if err != nil {
				return
			}
			f.mu.Lock()
			copy(f.img[a:], d)
			f.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			f.note("?%02X", c)
			return
		}
	}
}

func (f *anaFake) cmds() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.log, " ")
}

// done: comandi ricevuti, dopo aver atteso l'ultimo ('E' parte quando il driver
// ha già finito)
func (f *anaFake) done() string {
	for i := 0; i < 400 && !strings.HasSuffix(f.cmds(), " E"); i++ {
		time.Sleep(5 * time.Millisecond)
	}
	return f.cmds()
}

// sequenza attesa: aggancio, password, blocchi con la conferma dopo ognuno
func anaWant(pw int, write bool, areas [][2]int) string {
	hs := "PROGRAM 02 06"
	out := []string{hs, fmt.Sprintf("R%03X/8", pw), "06"}
	if write {
		out = append(out, "E", hs)
	}
	for _, g := range areas {
		for a := g[0]; a < g[1]; a += 8 {
			if write {
				out = append(out, fmt.Sprintf("W%03X/8", a))
			} else {
				out = append(out, fmt.Sprintf("R%03X/8", a), "06")
			}
		}
	}
	return strings.Join(append(out, "E"), " ")
}

func anaLink(t *testing.T, f *anaFake) func() {
	t.Helper()
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go f.serve(b)
	return func() { openPortFunc = old; _ = a.Close() }
}

func TestAnaCps432(t *testing.T) {
	type spec struct {
		id       string
		slots    int
		base, pw int
		size     int
		rd, wr   [][2]int
		set      int // blocco con suoni, squelch, tasto laterale e TOT
	}
	t20 := spec{slots: 22, base: 0, pw: 0x2D8, size: 0x340,
		rd: [][2]int{{0x000, 0x160}, {0x2B0, 0x2C8}, {0x330, 0x338}},
		wr: [][2]int{{0x000, 0x160}, {0x2B0, 0x2D0}}, set: 0x2C0}
	v2 := spec{slots: 16, base: 0x10, pw: 0x130, size: 0x3E0,
		rd: [][2]int{{0x000, 0x110}, {0x2B0, 0x2C0}, {0x330, 0x340}, {0x380, 0x3E0}},
		wr: [][2]int{{0x000, 0x110}, {0x2B0, 0x2C0}, {0x380, 0x3E0}}, set: 0x3C0}
	for _, sp := range []spec{func() spec { s := t20; s.id = "bf-t22frs"; return s }(), func() spec { s := t20; s.id = "td-m8s"; return s }(),
		func() spec { s := v2; s.id = "td-v2"; return s }()} {
		t.Run(sp.id, func(t *testing.T) {
			m := genByID[sp.id]
			if m == nil {
				t.Fatal("modello assente")
			}
			if m.Slots != sp.slots || m.Baud != 9600 || !m.NoChirp || t17o(m).memsize() != sp.size || !m.inBands(462562500) || !m.inBands(400000000) || m.inBands(480000001) || m.inBands(399999999) {
				t.Fatalf("dati del modello: %d canali, %d baud, %d byte", m.Slots, m.Baud, t17o(m).memsize())
			}
			// memoria della radio: ogni byte diverso, password assente
			radio := make([]byte, 0x400)
			for i := range radio {
				radio[i] = byte(i*7 + i>>8)
			}
			for i := 0; i < sp.slots*16; i++ {
				radio[sp.base+i] = 0xFF
			}
			copy(radio[sp.pw:], bytes.Repeat([]byte{0xFF}, 8))
			// canale 1 come lo scrive il programma: 462,5625 MHz, tono 67,0 Hz in
			// ricezione, D023N in trasmissione, potenza alta, banda stretta, in
			// scansione, senza blocco del canale occupato
			copy(radio[sp.base:], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0x70, 0x06, 0x23, 0x80, 0x0D, 0xFF, 0xFF, 0xFF})
			// canale 2: 467,7125 MHz, D754I in ricezione, niente in trasmissione,
			// potenza bassa, banda larga, fuori dalla scansione, con il salto di frequenza
			copy(radio[sp.base+16:], []byte{0x50, 0x12, 0x77, 0x46, 0x50, 0x12, 0x77, 0x46, 0x54, 0xC7, 0xFF, 0xFF, 0x91, 0xFF, 0xFF, 0xFF})
			f := &anaFake{img: append([]byte{}, radio...)}
			done := anaLink(t, f)
			img, _, err := m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
			done()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := f.done(), anaWant(sp.pw, false, sp.rd); got != want {
				t.Fatalf("lettura: comandi diversi dal programma (%d e %d caratteri)", len(got), len(want))
			}
			if len(img) != sp.size {
				t.Fatalf("immagine di %d byte", len(img))
			}
			for a := range img {
				in := false
				for _, g := range sp.rd {
					in = in || a >= g[0] && a < g[1]
				}
				if in && img[a] != radio[a] || !in && img[a] != 0xFF {
					t.Fatalf("immagine, byte %#x = %02X", a, img[a])
				}
			}
			c1, c2 := m.Fam.Get(m, img, 0), m.Fam.Get(m, img, 1)
			if c1 == nil || c1.Rx != 462562500 || c1.Tx != 462562500 || c1.RxTone != "67.0" || c1.TxTone != "D023N" || c1.Power != m.Powers[1].Sem || !c1.Narrow || !c1.Scan {
				t.Fatalf("canale 1: %+v", c1)
			}
			if c2 == nil || c2.Rx != 467712500 || c2.RxTone != "D754I" || c2.TxTone != "" || c2.Power != m.Powers[0].Sem || c2.Narrow || c2.Scan {
				t.Fatalf("canale 2: %+v", c2)
			}
			if m.Fam.Get(m, img, 2) != nil || m.Fam.Get(m, img, sp.slots-1) != nil {
				t.Fatal("canali vuoti letti come pieni")
			}
			// modifiche: canale 2 cambiato (il bit del salto resta), canale 3 nuovo, canale 1 cancellato
			c2.Narrow, c2.Scan, c2.Power = true, true, m.Powers[1].Sem
			if err := m.Fam.Set(m, img, 1, c2); err != nil {
				t.Fatal(err)
			}
			if err := m.Fam.Set(m, img, 2, &Chan{Rx: 462587500, Tx: 462587500, TxTone: "88.5", Power: m.Powers[0].Sem, Scan: true}); err != nil {
				t.Fatal(err)
			}
			if err := m.Fam.Set(m, img, 0, nil); err != nil {
				t.Fatal(err)
			}
			if got := img[sp.base+16+12]; got != 0x8D {
				t.Fatalf("canale 2, byte 12 = %02X", got)
			}
			if got, want := img[sp.base+32:sp.base+48], []byte{0x50, 0x87, 0x25, 0x46, 0x50, 0x87, 0x25, 0x46, 0xFF, 0xFF, 0x85, 0x08, 0x01, 0xFF, 0xFF, 0xFF}; !bytes.Equal(got, want) {
				t.Fatalf("canale nuovo: % X", got)
			}
			if !bytes.Equal(img[sp.base:sp.base+16], bytes.Repeat([]byte{0xFF}, 16)) {
				t.Fatalf("canale cancellato: % X", img[sp.base:sp.base+16])
			}
			// impostazioni
			tab := settingsFor(sp.id, img)
			if !settingsAvailable(sp.id) || tab == nil || len(tab.Settings) != 14 {
				t.Fatal("impostazioni assenti o in numero inatteso")
			}
			pos := map[string][3]int{"voice": {0x2B0, 0xFF, 2}, "lang": {0x2B1, 0xFF, 2}, "scan_on": {0x2B2, 0xFF, 2}, "vox_on": {0x2B3, 0xFF, 2},
				"vox_level": {0x2B4, 0xFF, 5}, "vox_rx": {0x2B5, 0xFF, 2}, "high_tx": {0x2B6, 0xFF, 2}, "low_tx": {0x2B7, 0xFF, 2}, "rx_emerg": {0x2B8, 0xFF, 2},
				"beep": {sp.set, 0x01, 2}, "save": {sp.set, 0x02, 2}, "sql": {sp.set + 1, 0xFF, 10}, "side_key": {sp.set + 2, 0xFF, 4}, "tot": {sp.set + 3, 0xFF, 11}}
			for i := range tab.Settings {
				d := &tab.Settings[i]
				w, ok := pos[d.Key]
				offs, masks, err := d.Enc.fields()
				if !ok || err != nil || len(offs) != 1 || offs[0] != w[0] || int(masks[0][0]) != w[1] || d.ItLabel == "" || d.ItDesc == "" {
					t.Fatalf("%s: posizione %#x maschera %X", d.Key, offs, masks)
				}
				n := len(d.Options)
				if d.Type == "bool" {
					n = 2
				}
				if n != w[2] {
					t.Fatalf("%s: %d scelte", d.Key, n)
				}
				delete(pos, d.Key)
			}
			if len(pos) != 0 {
				t.Fatalf("voci mancanti: %v", pos)
			}
			set := func(key, v string) {
				t.Helper()
				for i := range tab.Settings {
					if d := &tab.Settings[i]; d.Key == key {
						if err := writeSetting(d, img, v); err != nil {
							t.Fatalf("%s: %v", key, err)
						}
						if got, err := readSetting(d, img); err != nil || got != v {
							t.Fatalf("%s: riletto %q (%v)", key, got, err)
						}
						return
					}
				}
				t.Fatalf("%s: voce assente", key)
			}
			img[sp.set] = 0xA4
			set("vox_level", "4") // quinta scelta: livello 5
			set("side_key", "2")
			set("tot", "10")
			set("sql", "9")
			set("beep", "true")
			set("save", "false")
			set("vox_rx", "true")
			set("high_tx", "true")
			set("low_tx", "false")
			if img[0x2B4] != 5 || img[sp.set+2] != 2 || img[sp.set+3] != 10 || img[sp.set+1] != 9 || img[sp.set] != 0xA5 || img[0x2B5] != 1 || img[0x2B6] != 1 || img[0x2B7] != 0 {
				t.Fatalf("impostazioni scritte: % X / % X", img[0x2B0:0x2B9], img[sp.set:sp.set+4])
			}
			// scrittura
			f = &anaFake{img: append([]byte{}, radio...)}
			done = anaLink(t, f)
			_, err = m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
			done()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := f.done(), anaWant(sp.pw, true, sp.wr); got != want {
				t.Fatalf("scrittura: comandi diversi dal programma (%d e %d caratteri)", len(got), len(want))
			}
			for a := range f.img {
				in, rd := false, false
				for _, g := range sp.wr {
					in = in || a >= g[0] && a < g[1]
				}
				for _, g := range sp.rd {
					rd = rd || a >= g[0] && a < g[1]
				}
				want := radio[a]
				switch {
				case in && rd:
					want = img[a]
				case in:
					want = 0xFF // 0x2C8 del «T20 FRS»: FF come fa il programma
				}
				if f.img[a] != want {
					t.Fatalf("radio dopo la scrittura, byte %#x = %02X invece di %02X", a, f.img[a], want)
				}
			}
			// password impostata: ci si ferma prima di leggere o scrivere qualsiasi blocco
			for _, write := range []bool{false, true} {
				f = &anaFake{img: append([]byte{}, radio...)}
				f.img[sp.pw+2] = '1'
				done = anaLink(t, f)
				if write {
					_, err = m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
				} else {
					_, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
				}
				done()
				if err == nil || !strings.Contains(err.Error(), "password") {
					t.Fatalf("password impostata: %v", err)
				}
				if got, want := f.cmds(), fmt.Sprintf("PROGRAM 02 06 R%03X/8", sp.pw); got != want {
					t.Fatalf("password impostata, comandi: %s", got)
				}
				f.img[sp.pw+2] = 0xFF
				if !bytes.Equal(f.img, radio) {
					t.Fatal("password impostata: memoria della radio cambiata")
				}
			}
		})
	}
	// le T17 non cambiano: stesse opzioni di prima
	for id, want := range map[string][3]int{"bf-t17pmr": {0x7B0, 16, 0x660}, "bf-t17frs": {0x7B0, 16, 0x660}, "bf-gt18pmr": {0x7B0, 16, 0x660}, "rd-t3": {0x7B0, 16, 0x660}, "bf-t9": {0x7B0, 16, 0x660}} {
		o := t17o(genByID[id])
		if o.pw() != want[0] || o.rb() != want[1] || o.memsize() != want[2] || o.noJump || len(o.ffill) != 0 || len(o.wr()) != len(o.gaps) {
			t.Fatalf("%s: opzioni cambiate", id)
		}
	}
}
