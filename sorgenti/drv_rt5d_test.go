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

// 4.21, Radtel RT-5D: record del canale come lo scrive il programma del
// costruttore (SetChannel), valori calcolati a mano dai campi
func TestRT5DCodec(t *testing.T) {
	m := genByID["rtl-rt5d"]
	f := m.Fam.(rt5dFam)
	img := bytes.Repeat([]byte{0xFF}, rt5dSize())
	ff := func(n int) string { return hex.EncodeToString(bytes.Repeat([]byte{0xFF}, n)) }
	// sezioni: DTMF 272, chiavi 264, contatti 64000, gruppi 4096 → canali a 68632
	if rt5dOff("canali") != 68632 || rt5dSize() != 134424 {
		t.Fatalf("posizioni: canali a %d, immagine di %d byte", rt5dOff("canali"), rt5dSize())
	}
	cases := []struct {
		i    int
		c    *Chan
		want string
	}{
		// analogico stretto 433.500 (43350000 = 0x029577F0), TX 88.5 (885 = 0x0375), alta, in scansione
		{0, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "88.5", Power: 0, Narrow: true, Scan: true, Name: "TEST"},
			"f0779502" + "f0779502" + "0000" + "7503" + "00" + "00" + "0101" + "00" + "000000" + "01" + "00" + "01" + "00" + "ff" + "000000" + ff(4) + "54455354" + ff(8) + "0000" + ff(18)},
		// analogico largo con DCS: RX D754I (posizione 105+105 = 210 = 0xD2), TX D645N (posizione 94 = 0x5E), bassa, fuori scansione
		{1, &Chan{Rx: 145600000, Tx: 145000000, RxTone: "D754I", TxTone: "D645N", Power: 2, Scan: false, Name: "R0 ROMA"},
			"002bde00" + "a040dd00" + "d200" + "5e00" + "00" + "00" + "0100" + "02" + "000000" + "00" + "00" + "01" + "00" + "ff" + "000000" + ff(4) + "523020524f4d41" + ff(5) + "0000" + ff(18)},
		// digitale DMR, media, in scansione: valori predefiniti del programma
		{2, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 1, Scan: true, Name: "DMR"},
			"f0ca9c02" + "30329102" + "0000" + "0000" + "00" + "00" + "0000" + "01" + "000000" + "01" + "00" + "01" + "00" + "ff" + "000000" + ff(4) + "444d52" + ff(9) + "0000" + ff(18)},
		// solo ricezione: TX a FF; DCS D023N = posizione 1, D023I = posizione 106
		{1023, &Chan{Rx: 118100000, Tx: 118100000, TxOff: true, RxTone: "D023N", TxTone: "D023I", Power: 0, Scan: true, Name: "TWR"},
			"d034b400" + ff(4) + "0100" + "6a00" + "00" + "00" + "0100" + "00" + "000000" + "01" + "00" + "01" + "00" + "ff" + "000000" + ff(4) + "545752" + ff(9) + "0000" + ff(18)},
	}
	for _, k := range cases {
		if err := f.Set(m, img, k.i, k.c); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, k.i)); got != k.want || len(k.want) != 128 {
			t.Errorf("canale %d:\n got %s\nwant %s", k.i+1, got, k.want)
		}
		if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
			t.Errorf("canale %d riletto diverso: %+v", k.i+1, g)
		}
	}
	if &f.rec(img, 1023)[0] != &img[68632+1023*64] {
		t.Error("posizione del canale 1024")
	}
	// le voci non gestite di un canale esistente restano
	r := f.rec(img, 0)
	copy(r[12:14], []byte{3, 2})
	copy(r[17:20], []byte{5, 2, 1})
	copy(r[21:32], []byte{1, 7, 2, 0xFF, 1, 1, 1, 0x56, 0x34, 0x12, 0})
	copy(r[44:46], []byte{0x2C, 0x01})
	keep := append([]byte{}, r...)
	if err := f.Set(m, img, 0, &Chan{Rx: 433500000, Tx: 431900000, TxTone: "123.0", Power: 2, Scan: false, Name: "TEST2"}); err != nil {
		t.Fatal(err)
	}
	copy(keep[4:8], []byte{0xF0, 0x06, 0x93, 0x02}) // 43190000 = 0x029306F0
	copy(keep[10:12], []byte{0xCE, 0x04})           // 1230 = 0x04CE
	keep[15], keep[16], keep[20], keep[36] = 0, 2, 0, '2'
	if !bytes.Equal(r, keep) {
		t.Errorf("canale 1 modificato:\n got %x\nwant %x", r, keep)
	}
	// cancellazione
	if err := f.Set(m, img, 1, nil); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(f.rec(img, 1), bytes.Repeat([]byte{0xFF}, 64)) || f.Get(m, img, 1) != nil {
		t.Error("canale 2 non cancellato")
	}
	// primi tre byte a zero: libero, come nel programma
	copy(f.rec(img, 5), make([]byte, 64))
	if f.Get(m, img, 5) != nil {
		t.Error("canale a zero visto come usato")
	}
	// codice DCS non in tabella: rifiutato
	if err := f.Set(m, img, 6, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "D017N"}); err == nil {
		t.Error("codice DCS 017 accettato")
	}
	// pacchetto: A5 comando parametro lunghezza dati CRC16 XMODEM («123456789» → 31C3),
	// calcolato da «comando» alla fine dei dati
	if got := rt5dFrame('1', '2'<<8|'3', nil); crc16XModem([]byte("123456789")) != 0x31C3 || len(got) != 8 || got[0] != 0xA5 ||
		crc16XModem(got[1:6]) != uint16(got[6])<<8|uint16(got[7]) {
		t.Errorf("pacchetto vuoto: % X", got)
	}
}

// lettura e scrittura con la radio simulata, nell'ordine del programma del
// costruttore; i dati della radio (modello) non vengono mai scritti
func TestRT5DClone(t *testing.T) {
	m := genByID["rtl-rt5d"]
	f := m.Fam.(rt5dFam)
	base := make([]byte, rt5dSize())
	for i := range base {
		base[i] = byte(i*7) ^ byte(i>>9)
	}
	copy(base[rt5dOff("canali"):], bytes.Repeat([]byte{0xFF}, 1024*64))
	for i, c := range []*Chan{
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Scan: true},
		{Rx: 433500000, Tx: 433500000, Name: "U", Power: 2, Scan: true, Narrow: true},
	} {
		if err := f.Set(m, base, i*700, c); err != nil {
			t.Fatal(err)
		}
	}
	e := f.emu(m, base).(*rt5dEmu)
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		if baud != 115200 {
			t.Errorf("velocità %d", baud)
		}
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	e.badCRC = 2 // due risposte guaste: il pacchetto va rimandato
	got, _, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, base) {
		t.Fatal("lettura diversa")
	}
	// ordine: 02, 05, 46, poi 16, 15, 13 ×80, 14 ×4, 10 ×64, 11, 12, 19, 01 (più i due nuovi invii)
	seq := func(wr bool) []int {
		s := []int{0x02 << 16, 0x05 << 16, 0x46 << 16}
		for _, x := range rt5dSecs {
			c := x.rd
			if wr {
				c = x.wr
			}
			for k := 0; k < x.n && c != 0; k++ {
				s = append(s, int(c)<<16|k)
			}
		}
		return append(s, 0x01<<16)
	}
	dedup := func(l []int) []int {
		var o []int
		for i, v := range l {
			if i == 0 || v != l[i-1] {
				o = append(o, v)
			}
		}
		return o
	}
	want := seq(false)
	if len(e.log) != len(want)+2 {
		t.Fatalf("lettura: %d pacchetti, attesi %d + 2 nuovi invii", len(e.log), len(want))
	}
	for k, v := range dedup(e.log) {
		if v != want[k] {
			t.Fatalf("lettura: pacchetto %d = %06X, atteso %06X", k, v, want[k])
		}
	}
	if len(want) != 3+1+1+80+4+64+1+1+1+1 {
		t.Fatalf("sequenza di %d pacchetti", len(want))
	}
	e.log = nil
	if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Name: "PMR 1", Power: 2, Scan: true, Narrow: true}); err != nil {
		t.Fatal(err)
	}
	// i dati della radio nel file sono diversi: non devono arrivare alla radio
	bi := rt5dOff("dati della radio")
	copy(got[bi:], bytes.Repeat([]byte{0x33}, 64))
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	want = seq(true)
	if len(e.log) != len(want) || len(want) != 3+1+1+80+4+64+1+1+1 {
		t.Fatalf("scrittura: %d pacchetti, attesi %d", len(e.log), len(want))
	}
	for k, v := range e.log {
		if v != want[k] {
			t.Fatalf("scrittura: pacchetto %d = %06X, atteso %06X", k, v, want[k])
		}
	}
	after := e.image()
	if !bytes.Equal(after[:bi], got[:bi]) {
		t.Fatal("memoria scritta diversa dal file")
	}
	if !bytes.Equal(after[bi:], base[bi:]) {
		t.Fatal("dati della radio scritti")
	}
	// rifiuto della radio (comando EE): errore, niente altro inviato
	e.log, e.refuse = nil, 0x05
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "rifiutato") || len(e.log) != 2 {
		t.Fatalf("rifiuto: %v (%d pacchetti)", err, len(e.log))
	}
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !bytes.Equal(e.image(), after) {
		t.Fatalf("scrittura dopo un rifiuto: %v", err)
	}
}
