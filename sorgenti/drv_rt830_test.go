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
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

// 4.22, Radtel RT-830 e RT-850: record del canale come lo scrive il programma
// del costruttore (SaveData), valori calcolati a mano dai campi
func TestRT830Codec(t *testing.T) {
	type tc struct {
		i    int
		c    *Chan
		want string
	}
	for _, x := range []struct {
		id         string
		size, chAt int
		cases      []tc
	}{
		{"rtl-rt830", 0x870, 0x40, []tc{
			// 145.600 (0x00DE2B00) / 145.000 (0x00DD40A0), tono TX 88.5 (0x0375), alta, larga, in scansione: gamma 136-174 = 0x04
			{0, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true}, "00de2b00" + "3000" + "00dd40a0" + "0375" + "05" + "000000"},
			// 433.500 (0x029577F0) stretta, bassa, fuori scansione, RX D754I (0x21EC), TX D023N (0x1013): gamma 400-520 = 0x10
			{127, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D754I", TxTone: "D023N", Power: 2, Narrow: true}, "029577f0" + "21ec" + "029577f0" + "1013" + "d0" + "000000"},
			// 118.100 (0x00B434D0): gamma 100-136 = 0x00
			{5, &Chan{Rx: 118100000, Tx: 118100000, Scan: true}, "00b434d0" + "3000" + "00b434d0" + "3000" + "01" + "000000"},
			// 220.000 (0x014FB180): gamma 174-240 = 0x08; 300.000 (0x01C9C380): gamma 240-400 = 0x0C
			{6, &Chan{Rx: 220000000, Tx: 220000000}, "014fb180" + "3000" + "014fb180" + "3000" + "08" + "000000"},
			{7, &Chan{Rx: 300000000, Tx: 300000000}, "01c9c380" + "3000" + "01c9c380" + "3000" + "0c" + "000000"},
		}},
		{"rtl-rt850", 0x850, 0x50, []tc{
			// gli stessi valori in little endian; byte 12: 0x10 se RX ≥ 240 MHz, 0x08 se TX ≥ 240 MHz
			{0, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true}, "002bde00" + "0030" + "a040dd00" + "7503" + "01" + "000000"},
			{127, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D754I", TxTone: "D023N", Power: 2, Narrow: true}, "f0779502" + "ec21" + "f0779502" + "1310" + "d8" + "000000"},
			{5, &Chan{Rx: 433500000, Tx: 145000000, Scan: true}, "f0779502" + "0030" + "a040dd00" + "0030" + "11" + "000000"},
			{6, &Chan{Rx: 145000000, Tx: 433500000}, "a040dd00" + "0030" + "f0779502" + "0030" + "08" + "000000"},
		}},
	} {
		m := genByID[x.id]
		f := m.Fam.(rt830Fam)
		o := rt830o(m)
		if o.size() != x.size || o.chOff() != x.chAt || m.Slots != 128 {
			t.Fatalf("%s: immagine di %d byte, canali a %d", x.id, o.size(), o.chOff())
		}
		img := bytes.Repeat([]byte{0xFF}, o.size())
		for _, k := range x.cases {
			if err := f.Set(m, img, k.i, k.c); err != nil {
				t.Fatal(err)
			}
			if got := hex.EncodeToString(f.rec(m, img, k.i)); got != k.want || len(k.want) != 32 {
				t.Errorf("%s canale %d:\n got %s\nwant %s", x.id, k.i+1, got, k.want)
			}
			if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
				t.Errorf("%s canale %d riletto diverso: %+v", x.id, k.i+1, g)
			}
		}
		if &f.rec(m, img, 127)[0] != &img[x.chAt+127*16] {
			t.Error("posizione del canale 128")
		}
		// le voci non gestite di un canale esistente restano: tipo di DCS (2 bit alti
		// del tono RX), cifratura (2 bit alti del tono TX), limite di trasmissione
		// (0x02) e codice di silenziamento
		r := f.rec(m, img, 0)
		o.put16(r[4:6], o.u16(r[4:6])|0x8000)
		o.put16(r[10:12], o.u16(r[10:12])|0x4000)
		r[12] |= 0x02
		copy(r[13:16], []byte{1, 2, 3})
		if g := f.Get(m, img, 0); g == nil || g.RxTone != "" || g.TxTone != "88.5" {
			t.Errorf("%s: toni con i bit alti: %+v", x.id, g)
		}
		if err := f.Set(m, img, 0, &Chan{Rx: 145600000, Tx: 145000000, RxTone: "D023I", TxTone: "123.0", Power: 2}); err != nil {
			t.Fatal(err)
		}
		// D023I = 0x2013 con il tipo di DCS 0x8000 = 0xA013; 123.0 = 0x04CE con la cifratura 0x4000 = 0x44CE
		band := byte(0x04)
		if !o.be {
			band = 0
		}
		if o.u16(r[4:6]) != 0xA013 || o.u16(r[10:12]) != 0x44CE || r[12] != 0x80|0x02|band || !bytes.Equal(r[13:16], []byte{1, 2, 3}) {
			t.Errorf("%s canale 1 modificato: %x", x.id, r)
		}
		// cancellazione: tutto a FF; libero anche con la frequenza a zero o con il bit 0x20
		if err := f.Set(m, img, 5, nil); err != nil || !bytes.Equal(f.rec(m, img, 5), bytes.Repeat([]byte{0xFF}, 16)) || f.Get(m, img, 5) != nil {
			t.Errorf("%s: canale 6 non cancellato (%v)", x.id, err)
		}
		copy(f.rec(m, img, 5), make([]byte, 16))
		f.rec(m, img, 6)[12] |= 0x20
		if f.Get(m, img, 5) != nil || f.Get(m, img, 6) != nil {
			t.Errorf("%s: canale libero visto come usato", x.id)
		}
		// rifiutati: sola ricezione, fuori dai limiti del programma, codice DCS non in tabella
		for _, c := range []*Chan{
			{Rx: 145500000, Tx: 145500000, TxOff: true},
			{Rx: o.lo - 10, Tx: o.lo - 10},
			{Rx: 145500000, Tx: o.hi + 10},
			{Rx: 145500000, Tx: 145500000, TxTone: "D017N"},
		} {
			if err := f.Set(m, img, 9, c); err == nil {
				t.Errorf("%s: canale accettato: %+v", x.id, c)
			}
		}
	}
}

// lettura e scrittura con la radio simulata, nell'ordine dei programmi del
// costruttore: la password non viene mai scritta, il resto della memoria della
// radio non viene toccato
func TestRT830Clone(t *testing.T) {
	for _, x := range []struct {
		id                 string
		rd0                bool // la lettura passa per l'indirizzo 0x0000 (controllo del modello)
		first, lastR, last int  // primo blocco dopo la password, ultimo letto, ultimo scritto
	}{
		{"rtl-rt830", false, 0x01E0, 0x0A30, 0x0A20},
		{"rtl-rt850", true, 0x0350, 0x0B80, 0x0B80},
	} {
		m := genByID[x.id]
		f := m.Fam.(rt830Fam)
		o := rt830o(m)
		base := make([]byte, o.size())
		for i := range base {
			base[i] = byte(i*3) ^ byte(i>>6)
		}
		copy(base, bytes.Repeat([]byte{0xFF}, 16)) // nessuna password
		copy(base[o.chOff():], bytes.Repeat([]byte{0xFF}, 128*16))
		for i, c := range []*Chan{
			{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true},
			{Rx: 433500000, Tx: 433500000, Power: 2, Scan: true, Narrow: true},
		} {
			if err := f.Set(m, base, i*100, c); err != nil {
				t.Fatal(err)
			}
		}
		e := f.emu(m, base).(*rt830Emu)
		radio := append([]byte{}, e.mem...)
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			if baud != 9600 {
				t.Errorf("velocità %d", baud)
			}
			a, b := newPipe()
			a.baud = baud
			go e.serve(b)
			return a, nil
		}
		logged := func(n int) []int {
			for k := 0; ; k++ {
				e.mu.Lock()
				l := append([]int{}, e.log...)
				e.mu.Unlock()
				if len(l) >= n || k > 200 {
					return l
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
		cmp := func(what string, want []int) {
			t.Helper()
			l := logged(len(want))
			e.mu.Lock()
			e.log = nil
			e.mu.Unlock()
			if len(l) != len(want) {
				t.Fatalf("%s %s: %d comandi, attesi %d", x.id, what, len(l), len(want))
			}
			for k, v := range l {
				if v != want[k] {
					t.Fatalf("%s %s: comando %d = %06X, atteso %06X", x.id, what, k, v, want[k])
				}
			}
		}
		got, _, err := f.Download(context.Background(), m, "COM7", noProg)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, base) {
			t.Fatalf("%s: lettura diversa", x.id)
		}
		want := []int{'I' << 16, 'R'<<16 | o.base}
		if x.rd0 {
			want = append(want, 'R'<<16)
		}
		for a := x.first; a <= x.lastR; a += 16 {
			want = append(want, 'R'<<16|a)
		}
		cmp("lettura", append(want, 'E'<<16))
		if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Scan: true, Narrow: true}); err != nil {
			t.Fatal(err)
		}
		// la password nel file è diversa: non deve arrivare alla radio
		copy(got, bytes.Repeat([]byte{0x31}, 16))
		if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
			t.Fatal(err)
		}
		want = []int{'I' << 16, 'R'<<16 | o.base, 'R' << 16}
		for a := x.first; a <= x.last; a += 16 {
			want = append(want, 'W'<<16|a)
		}
		cmp("scrittura", append(want, 'E'<<16))
		// nella radio: scritto da «first» a «last», il resto com'era
		e.mu.Lock()
		after := append([]byte{}, e.mem...)
		e.mu.Unlock()
		for a := 0; a < len(after); a += 16 {
			ref := radio[a : a+16]
			if a >= x.first && a <= x.last {
				ref = got[a-o.base : a-o.base+16]
			}
			if !bytes.Equal(after[a:a+16], ref) {
				t.Fatalf("%s: blocco 0x%04X scritto fuori dalle aree del programma o non scritto", x.id, a)
			}
		}
		if g := f.Get(m, e.image(), 7); g == nil || g.Rx != 446006250 || !g.Narrow {
			t.Fatalf("%s: canale 8 nella radio: %+v", x.id, g)
		}
		// radio con la password: né lettura né scrittura
		e.mu.Lock()
		copy(e.mem[o.base:], "1234")
		e.mu.Unlock()
		if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatalf("%s: lettura con la password: %v", x.id, err)
		}
		cmp("lettura con la password", []int{'I' << 16, 'R'<<16 | o.base, 'E' << 16})
		if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatalf("%s: scrittura con la password: %v", x.id, err)
		}
		cmp("scrittura con la password", []int{'I' << 16, 'R'<<16 | o.base, 'E' << 16})
		// RT-850: un'altra radio (primo byte dell'indirizzo 0x0000 diverso da 0x58) viene rifiutata
		e.mu.Lock()
		copy(e.mem[o.base:], bytes.Repeat([]byte{0xFF}, 16))
		e.mem[0] = 0x57
		e.mu.Unlock()
		_, _, err = f.Download(context.Background(), m, "COM7", noProg)
		if x.rd0 != (err != nil) || (err != nil && !strings.Contains(err.Error(), "(57)")) {
			t.Fatalf("%s: controllo del modello: %v", x.id, err)
		}
		openPortFunc = old
	}
}
