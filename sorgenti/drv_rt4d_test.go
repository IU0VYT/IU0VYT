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
)

// 4.20, Radtel RT-4D / RT-6D: record del canale come lo scrive il programma del
// costruttore (SaveCHData), valori calcolati a mano dai campi
func TestRT4DCodec(t *testing.T) {
	ff := func(n int) string { return hex.EncodeToString(bytes.Repeat([]byte{0xFF}, n)) }
	for _, id := range []string{"rtl-rt4d", "rtl-rt6d"} {
		m := genByID[id]
		f := m.Fam.(rt4dFam)
		img := bytes.Repeat([]byte{0xFF}, rt4dSize)
		cases := []struct {
			i    int
			c    *Chan
			want string
		}{
			// analogico 433.500 (43350000 = 0x029577F0), TX 88.5 (0x1000|885 = 0x1375), stretta, alta, in scansione
			{0, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "88.5", Power: 0, Narrow: true, Scan: true, Name: "TEST"},
				"40" + "10" + "4e" + "01" + "40" + "f0779502" + "f0779502" + "0000" + "7513" + "0000" + "00" + "0000" + "01000000" + "00000000" + "0000" + "54455354" + ff(12)},
			// analogico con DCS: RX D754I (0o754 = 0x1EC | 0x3000), TX D023N (0o23 = 0x13 | 0x2000), larga, bassa, escluso dalla scansione
			{1, &Chan{Rx: 145600000, Tx: 145000000, RxTone: "D754I", TxTone: "D023N", Power: 2, Scan: false, Name: "R0 ROMA"},
				"40" + "10" + "0e" + "81" + "00" + "002bde00" + "a040dd00" + "ec31" + "1320" + "0000" + "00" + "0000" + "01000000" + "00000000" + "0000" + "523020524f4d41" + ff(9)},
			// digitale DMR, alta, in scansione: valori predefiniti del programma
			{2, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 0, Scan: true, Name: "DMR"},
				"00" + "10" + "4e" + "01" + "00" + "f0ca9c02" + "30329102" + "0000" + "0000" + "0000" + "00" + "0000" + "01000000" + "00000000" + "0000" + "444d52" + ff(13)},
			// gamma aeronautica: AM (bit 5-4 del byte 4 = 1)
			{1023, &Chan{Rx: 118100000, Tx: 118100000, Power: 0, Scan: true, Name: "TWR"},
				"40" + "10" + "4e" + "01" + "10" + "d034b400" + "d034b400" + "0000" + "0000" + "0000" + "00" + "0000" + "01000000" + "00000000" + "0000" + "545752" + ff(13)},
		}
		for _, k := range cases {
			if err := f.Set(m, img, k.i, k.c); err != nil {
				t.Fatal(err)
			}
			if got := hex.EncodeToString(f.rec(img, k.i)); got != k.want {
				t.Errorf("%s canale %d:\n got %s\nwant %s", id, k.i+1, got, k.want)
			}
			if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
				t.Errorf("%s canale %d riletto diverso: %+v", id, k.i+1, g)
			}
		}
		// il canale 1 sta a 0 dell'immagine (blocco 16 della radio), il 1024 a 1023·48
		if !bytes.Equal(img[1023*48:1023*48+3], []byte{0x40, 0x10, 0x4E}) {
			t.Errorf("%s: posizione del canale 1024", id)
		}
		// zona 1 (a 0xC000 dell'immagine, blocco 0x78): canali scelti a 0, nome «Zone1», i quattro canali
		z := f.zone(img, 0)
		if want := "00000000" + hex.EncodeToString([]byte("Zone1")) + ff(11) + "000001000200ff03" + ff(246*2); hex.EncodeToString(z) != want || &z[0] != &img[0xC000] {
			t.Errorf("%s zona 1:\n got %x\nwant %s", id, z, want)
		}
		// le voci non gestite di un canale esistente restano: solo RX / TOT / priorità / tipo di subtono / contatto / gruppi / ID / SSB
		r := f.rec(img, 0)
		r[0] |= 0x10 | 0x08
		r[1] = 0x35
		r[2] = r[2]&0xC0 | 0x05
		r[3] = 0x80 | 0x40 | 0x10 | 0x03
		r[4] |= 0x20 | 0x06
		copy(r[17:30], []byte{1, 2, 3, 4, 5, 0x78, 0x56, 0x34, 0x12, 9, 8, 7, 6})
		keep := append([]byte{}, r...)
		if err := f.Set(m, img, 0, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "88.5", Power: 2, Narrow: false, Scan: true, Name: "TEST2"}); err != nil {
			t.Fatal(err)
		}
		keep[2] &^= 0xC0 // potenza bassa
		keep[3] &^= 0x80 // in scansione
		keep[4] &^= 0xC0 // larga (SSB resta: la frequenza non cambia)
		keep[36] = '2'
		if !bytes.Equal(r, keep) {
			t.Errorf("%s canale 1 modificato:\n got %x\nwant %x", id, r, keep)
		}
		// frequenza diversa: il modo torna FM
		if err := f.Set(m, img, 0, &Chan{Rx: 433525000, Tx: 433525000, Power: 2, Scan: true, Name: "TEST2"}); err != nil {
			t.Fatal(err)
		}
		if r[4]&0x30 != 0 {
			t.Errorf("%s: modo non riportato a FM (%02X)", id, r[4])
		}
		// da digitale ad analogico: valori predefiniti
		if err := f.Set(m, img, 2, &Chan{Rx: 438300000, Tx: 430700000, Power: 0, Scan: true, Name: "DMR"}); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, 2)[:5]); got != "40104e0100" {
			t.Errorf("%s canale 3 analogico: %s", id, got)
		}
		// cancellazione: canale a FF e tolto dalla zona (elenco compatto, canali scelti e nome invariati)
		copy(z[0:4], []byte{2, 0, 3, 0})
		if err := f.Set(m, img, 1, nil); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, 1)); got != ff(48) {
			t.Errorf("%s canale 2 cancellato: %s", id, got)
		}
		if want := "02000300" + hex.EncodeToString([]byte("Zone1")) + ff(11) + "00000200ff03" + ff(247*2); hex.EncodeToString(z) != want {
			t.Errorf("%s zona 1 dopo la cancellazione:\n got %x\nwant %s", id, z, want)
		}
		// un canale già in un'altra zona non viene aggiunto alla zona 1
		z5 := f.zone(img, 4)
		copy(z5, append([]byte{0, 0, 0, 0}, []byte("Ponti\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff")...))
		z5[20], z5[21] = 9, 0
		if err := f.Set(m, img, 9, &Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", Scan: true, Name: "R6"}); err != nil {
			t.Fatal(err)
		}
		if got := f.zoneMembers(img, 0); len(got) != 3 {
			t.Errorf("%s: canale 10 aggiunto alla zona 1: %v", id, got)
		}
	}
}

// RT-1000 Pro: solo analogica, tre canali scelti e nome di 14 byte nella zona,
// scansione non gestita (byte 3 conservato)
func TestRT1000ProCodec(t *testing.T) {
	m := genByID["rtl-rt1000pro"]
	f := m.Fam.(rt4dFam)
	img := bytes.Repeat([]byte{0xFF}, rt4dSize)
	ff := func(n int) string { return hex.EncodeToString(bytes.Repeat([]byte{0xFF}, n)) }
	// tipo 0 (digitale nelle DMR) qui vuol dire libero
	copy(f.rec(img, 5), make([]byte, 48))
	if f.Get(m, img, 5) != nil {
		t.Fatal("canale con tipo 0 visto come usato")
	}
	c := &Chan{Rx: 446006250, Tx: 446006250, TxTone: "D023N", RxTone: "D023N", Power: 2, Narrow: true, Scan: true, Name: "PMR 1"}
	if err := f.Set(m, img, 5, c); err != nil {
		t.Fatal(err)
	}
	// 446.00625 MHz = 44600625 = 0x02A88D31
	if got, want := hex.EncodeToString(f.rec(img, 5)), "40"+"10"+"0e"+"01"+"40"+"318da802"+"318da802"+"1320"+"1320"+"0000"+"00"+"0000"+"01000000"+"00000000"+"0000"+"504d522031"+ff(11); got != want {
		t.Errorf("canale 6:\n got %s\nwant %s", got, want)
	}
	if g := f.Get(m, img, 5); g == nil || *g != *c {
		t.Errorf("canale 6 riletto diverso: %+v", g)
	}
	// un canale DMR chiesto a una radio analogica resta analogico
	if err := f.Set(m, img, 6, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Name: "X"}); err != nil {
		t.Fatal(err)
	}
	if f.rec(img, 6)[0]>>6 != 1 || f.Get(m, img, 6).Dig {
		t.Error("canale digitale su radio analogica")
	}
	z := f.zone(img, 0)
	if want := "000000000000" + hex.EncodeToString([]byte("Zone1")) + ff(9) + "05000600" + ff(248*2); hex.EncodeToString(z) != want {
		t.Errorf("zona 1:\n got %x\nwant %s", z, want)
	}
	// byte 3 (scansione e priorità, bit incerti nel programma del costruttore): resta
	f.rec(img, 5)[3] = 0xC5
	c.Scan = false
	if err := f.Set(m, img, 5, c); err != nil {
		t.Fatal(err)
	}
	if f.rec(img, 5)[3] != 0xC5 || !f.Get(m, img, 5).Scan {
		t.Errorf("byte 3 cambiato: %02X", f.rec(img, 5)[3])
	}
}

// lettura e scrittura con la radio simulata: si leggono il blocco 8, i canali e
// le zone; si scrivono, nell'ordine del programma, impostazioni e VFO come sono
// sulla radio, canali e zone; contatti, gruppi, cifratura, messaggi e radio FM
// non si toccano
func TestRT4DClone(t *testing.T) {
	for _, id := range []string{"rtl-rt4d", "rtl-rt6d", "rtl-rt1000pro"} {
		m := genByID[id]
		f := m.Fam.(rt4dFam)
		base := bytes.Repeat([]byte{0xFF}, rt4dSize)
		for i, c := range []*Chan{
			{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Scan: true},
			{Rx: 433500000, Tx: 433500000, Name: "U", Power: 2, Scan: true, Narrow: true},
		} {
			if err := f.Set(m, base, i*700, c); err != nil {
				t.Fatal(err)
			}
		}
		e := f.emu(m, base).(*rt4dEmu)
		// tutto il resto della radio ha altri dati
		for n := 0; n < 0x400; n++ {
			if _, ok := e.blocks[n]; !ok {
				e.blocks[n] = bytes.Repeat([]byte{byte(n) ^ 0x5A}, dm5kBlock)
			}
		}
		copy(e.blocks[rt4dBlkCfg], bytes.Repeat([]byte{0xFF}, 12)) // nessuna password
		e.blocks[rt4dBlkCfg][14], e.blocks[rt4dBlkCfg][15] = 0x12, 0x34
		radio := map[int][]byte{}
		for n, d := range e.blocks {
			radio[n] = append([]byte{}, d...)
		}
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			if baud != 115200 {
				t.Errorf("velocità %d", baud)
			}
			a, b := newPipe()
			a.baud = baud
			go e.serve(b)
			return a, nil
		}
		got, _, err := f.Download(context.Background(), m, "COM7", noProg)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, base) {
			t.Fatalf("%s: lettura diversa", id)
		}
		// blocchi letti: 8, poi 16-63, poi 0x78-0xF7
		if len(e.reads) != 1+48+128 || e.reads[0] != 8 || e.reads[1] != 16 || e.reads[48] != 63 || e.reads[49] != 0x78 || e.reads[176] != 0xF7 {
			t.Fatalf("%s: blocchi letti %v", id, e.reads)
		}
		e.reads = nil
		if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Name: "PMR 1", Power: 2, Scan: true, Narrow: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
			t.Fatal(err)
		}
		// prima della scrittura: blocchi 0, 8 e 0x70
		if len(e.reads) != 3 || e.reads[0] != 0 || e.reads[1] != 8 || e.reads[2] != 0x70 {
			t.Fatalf("%s: blocchi letti prima della scrittura %v", id, e.reads)
		}
		// ordine del programma: 0x90 0, 0x91 0-47, 0x92 0, 0x93 0-127
		var want []int
		want = append(want, 0x90<<16)
		for k := 0; k < 48; k++ {
			want = append(want, 0x91<<16|k)
		}
		want = append(want, 0x92<<16)
		for k := 0; k < 128; k++ {
			want = append(want, 0x93<<16|k)
		}
		if len(e.writes) != len(want) {
			t.Fatalf("%s: %d blocchi scritti, attesi %d", id, len(e.writes), len(want))
		}
		for k := range want {
			if e.writes[k] != want[k] {
				t.Fatalf("%s: scrittura %d = %06X, attesa %06X", id, k, e.writes[k], want[k])
			}
		}
		if !bytes.Equal(e.image(), got) {
			t.Fatalf("%s: canali o zone scritti diversi", id)
		}
		for n, d := range radio {
			in := n >= rt4dBlkCh && n < rt4dBlkCh+rt4dChBlocks || n >= rt4dBlkZone && n < rt4dBlkZone+rt4dZnBlocks
			if in {
				continue
			}
			ref := append([]byte{}, d...)
			if n == rt4dBlkCfg {
				ref[14], ref[15] = 0xFF, 0xFF // come il programma del costruttore
			}
			if !bytes.Equal(e.blocks[n], ref) {
				t.Fatalf("%s: blocco %d cambiato", id, n)
			}
		}
		// password impostata: né lettura né scrittura
		copy(e.blocks[rt4dBlkCfg], "123456")
		e.writes = nil
		if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatalf("%s: lettura con password: %v", id, err)
		}
		if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "password") || len(e.writes) != 0 {
			t.Fatalf("%s: scrittura con password: %v (%d blocchi scritti)", id, err, len(e.writes))
		}
		openPortFunc = old
	}
}
