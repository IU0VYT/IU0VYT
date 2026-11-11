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

// 4.22, Radtel RT-480 / RT-68: record del canale e nome come li scrive il
// programma del costruttore (GetCHImf_StrToHex, SetCHNaemToHex), valori
// calcolati a mano dai campi
func TestRT480Codec(t *testing.T) {
	m := genByID["rtl-rt480"]
	f := m.Fam.(rt480Fam)
	if rt480Size() != 0x21C0 || genByID["rtl-rt68"] == nil || m.First != 0 || m.Slots != 256 {
		t.Fatalf("immagine di %d byte", rt480Size())
	}
	// tabella DCS del programma: 105 codici, 645 compreso, in ordine (D023N = 1,
	// D645N = 94, D754N = 105; invertiti da 106 a 210)
	if len(uv5rDTCS) != 105 || uv5rDTCS[0] != 23 || uv5rDTCS[93] != 645 || uv5rDTCS[104] != 754 {
		t.Fatalf("tabella DCS: %d codici", len(uv5rDTCS))
	}
	img := bytes.Repeat([]byte{0xFF}, rt480Size())
	ff := func(n int) string { return hex.EncodeToString(bytes.Repeat([]byte{0xFF}, n)) }
	cases := []struct {
		i          int
		c          *Chan
		want, name string
	}{
		// 145.600 / 145.000, tono TX 88.5 (885 = 0x0375), alta, larga, in scansione
		{0, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Power: 0, Scan: true, Name: "R0 ROMA"},
			"00005614" + "00005014" + "0000" + "7503" + "00" + "00" + "00" + "04", "523020524f4d41" + ff(9)},
		// 433.500 stretta, bassa, fuori scansione, DCS: RX D754I (posizione 210 = 0xD2), TX D645N (94 = 0x5E)
		{1, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D754I", TxTone: "D645N", Power: 2, Narrow: true, Name: "U"},
			"00003543" + "00003543" + "d200" + "5e00" + "00" + "00" + "01" + "40", "55" + ff(15)},
		// 446.00625 (passo di 6,25 kHz), DCS D023N = 1 e D023I = 106 (0x6A), nome di 12 caratteri
		{255, &Chan{Rx: 446006250, Tx: 446006250, RxTone: "D023N", TxTone: "D023I", Power: 2, Narrow: true, Scan: true, Name: "PMR 1 ABCDEF"},
			"25066044" + "25066044" + "0100" + "6a00" + "00" + "00" + "01" + "44", "504d522031204142434445" + "46" + ff(4)},
	}
	for _, k := range cases {
		if err := f.Set(m, img, k.i, k.c); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, k.i)); got != k.want || len(k.want) != 32 {
			t.Errorf("canale %d:\n got %s\nwant %s", k.i, got, k.want)
		}
		if got := hex.EncodeToString(f.name(img, k.i)); got != k.name || len(k.name) != 32 {
			t.Errorf("nome %d:\n got %s\nwant %s", k.i, got, k.name)
		}
		if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
			t.Errorf("canale %d riletto diverso: %+v", k.i, g)
		}
	}
	if &f.rec(img, 255)[0] != &img[0x0FF0] || &f.name(img, 255)[0] != &img[0x1FF0] {
		t.Error("posizione del canale 255")
	}
	// le voci non gestite di un canale esistente restano: segnalazione, PTT ID,
	// scrambler (4 bit alti del byte 14) e blocco su canale occupato (0x08); gli
	// altri bit del byte 15 vengono azzerati, come fa il programma
	r := f.rec(img, 0)
	r[12], r[13], r[14], r[15] = 4, 3, 0x52, 0xFF
	keep := append([]byte{}, r...)
	if g := f.Get(m, img, 0); g == nil || g.Power != 2 || !g.Narrow || !g.Scan {
		t.Errorf("potenza 2 (media) letta come %+v", g)
	}
	if err := f.Set(m, img, 0, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "123.0", Power: 0, Scan: false, Name: "R0 ROMA"}); err != nil {
		t.Fatal(err)
	}
	copy(keep[10:12], []byte{0xCE, 0x04}) // 1230 = 0x04CE
	keep[14], keep[15] = 0x50, 0x08
	if !bytes.Equal(r, keep) {
		t.Errorf("canale 0 modificato:\n got %x\nwant %x", r, keep)
	}
	// nome in GB2312 (中 = D6 D0, 文 = CE C4): si legge, resta se non cambia e non
	// viene spezzato a metà di un carattere
	copy(f.name(img, 1), append([]byte{0xD6, 0xD0, 0xCE, 0xC4, 'A'}, bytes.Repeat([]byte{0xFF}, 11)...))
	g := f.Get(m, img, 1)
	if g == nil || g.Name != "中文A" {
		t.Fatalf("nome GB2312: %+v", g)
	}
	g.Power = 0
	if err := f.Set(m, img, 1, g); err != nil || hex.EncodeToString(f.name(img, 1)) != "d6d0cec441"+ff(11) || f.rec(img, 1)[14] != 0 {
		t.Errorf("nome GB2312 dopo una modifica: %x (%v)", f.name(img, 1), err)
	}
	if got := hex.EncodeToString(rt480Name("ABCDEFGHIJK中文")); got != "4142434445464748494a4b"+ff(5) {
		t.Errorf("nome troncato: %s", got)
	}
	// cancellazione
	if err := f.Set(m, img, 1, nil); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(f.rec(img, 1), bytes.Repeat([]byte{0xFF}, 16)) || !bytes.Equal(f.name(img, 1), bytes.Repeat([]byte{0xFF}, 16)) || f.Get(m, img, 1) != nil {
		t.Error("canale 1 non cancellato")
	}
	// frequenza vuota come nel programma: ultimo byte a zero o a FF
	copy(f.rec(img, 5), make([]byte, 16))
	if f.Get(m, img, 5) != nil {
		t.Error("canale a zero visto come usato")
	}
	// trasmissione vuota: sola ricezione in lettura
	copy(f.rec(img, 6), f.rec(img, 0))
	copy(f.rec(img, 6)[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	if g := f.Get(m, img, 6); g == nil || !g.TxOff || g.Tx != g.Rx {
		t.Errorf("TX vuoto: %+v", g)
	}
	// rifiutati: sola ricezione, passo diverso da 1,25 kHz, frequenze fuori dai limiti
	// del programma (134-176, 200-260, 300-520 MHz), codice DCS non in tabella
	for _, c := range []*Chan{
		{Rx: 145500000, Tx: 145500000, TxOff: true},
		{Rx: 145501000, Tx: 145501000},
		{Rx: 176000000, Tx: 176000000},
		{Rx: 145500000, Tx: 280000000},
		{Rx: 133998750, Tx: 133998750},
		{Rx: 145500000, Tx: 145500000, TxTone: "D017N"},
	} {
		if err := f.Set(m, img, 7, c); err == nil {
			t.Errorf("canale accettato: %+v", c)
		}
	}
	if err := genValidate(m, &Chan{Rx: 145500000, Tx: 145500000, TxOff: true}, 7); err == nil {
		t.Error("sola ricezione accettata")
	}
	// ai limiti: 134.000, 175.99875, 200.000, 259.99875, 300.000 e 520.000 MHz sono ammessi
	for _, hz := range []int64{134000000, 175998750, 200000000, 259998750, 300000000, 520000000} {
		if err := f.Set(m, img, 8, &Chan{Rx: hz, Tx: hz}); err != nil {
			t.Errorf("%d Hz: %v", hz, err)
		}
	}
}

// lettura e scrittura con la radio simulata, nell'ordine del programma del
// costruttore; impostazioni, VFO e DTMF non vengono mai scritti
func TestRT480Clone(t *testing.T) {
	m := genByID["rtl-rt68"]
	f := m.Fam.(rt480Fam)
	base := make([]byte, rt480Size())
	for i := range base {
		base[i] = byte(i*5) ^ byte(i>>7)
	}
	copy(base, bytes.Repeat([]byte{0xFF}, rt480ChEnd))
	for i, c := range []*Chan{
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Scan: true},
		{Rx: 433500000, Tx: 433500000, Name: "U", Power: 2, Scan: true, Narrow: true},
	} {
		if err := f.Set(m, base, i*200, c); err != nil {
			t.Fatal(err)
		}
	}
	e := f.emu(m, base).(*rt480Emu)
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		if baud != 9600 {
			t.Errorf("velocità %d", baud)
		}
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	e.noise, e.deaf = 3, 1 // un aggancio perso e byte estranei prima dello 0x06
	got, info, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, base) || !strings.Contains(info, "S19") {
		t.Fatalf("lettura diversa (%s)", info)
	}
	// ordine: F, poi R da 0x0000 a 0x1FC0, 0x3A00, 0x3A40, da 0x3B00 a 0x3C00, E
	want := []int{'F' << 16}
	for a := 0; a < 0x2000; a += 0x40 {
		want = append(want, 'R'<<16|a)
	}
	for _, a := range []int{0x3A00, 0x3A40, 0x3B00, 0x3B40, 0x3B80, 0x3BC0, 0x3C00} {
		want = append(want, 'R'<<16|a)
	}
	want = append(want, 'E'<<16)
	// comandi ricevuti dalla radio simulata (l'ultimo, 'E', arriva dopo la fine
	// della funzione: lo si aspetta per un momento)
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
	reset := func() {
		e.mu.Lock()
		e.log = nil
		e.mu.Unlock()
	}
	cmp := func(what string) {
		t.Helper()
		l := logged(len(want))
		if len(l) != len(want) {
			t.Fatalf("%s: %d comandi, attesi %d", what, len(l), len(want))
		}
		for k, v := range l {
			if v != want[k] {
				t.Fatalf("%s: comando %d = %06X, atteso %06X", what, k, v, want[k])
			}
		}
	}
	cmp("lettura")
	if len(want) != 1+128+7+1 {
		t.Fatalf("sequenza di %d comandi", len(want))
	}
	reset()
	e.noise = 0
	if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Name: "PMR 1", Power: 2, Scan: true, Narrow: true}); err != nil {
		t.Fatal(err)
	}
	// le impostazioni nel file sono diverse: non devono arrivare alla radio
	copy(got[rt480ChEnd:], bytes.Repeat([]byte{0x33}, rt480Size()-rt480ChEnd))
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	want = []int{'F' << 16}
	for a := 0; a < 0x2000; a += 0x40 {
		want = append(want, 'W'<<16|a)
	}
	want = append(want, 'E'<<16)
	cmp("scrittura")
	after := e.image()
	if !bytes.Equal(after[:rt480ChEnd], got[:rt480ChEnd]) {
		t.Fatal("canali scritti diversi dal file")
	}
	if !bytes.Equal(after[rt480ChEnd:], base[rt480ChEnd:]) {
		t.Fatal("impostazioni, VFO o DTMF scritti")
	}
	if g := f.Get(m, after, 7); g == nil || g.Name != "PMR 1" || g.Rx != 446006250 {
		t.Fatalf("canale 7 nella radio: %+v", g)
	}
	// blocco rifiutato: errore con l'indirizzo, uscita dalla programmazione e niente altro
	reset()
	e.nak = 0x0080
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "0x0080") {
		t.Fatalf("blocco rifiutato: %v", err)
	}
	want = []int{'F' << 16, 'W' << 16, 'W'<<16 | 0x40, 'W'<<16 | 0x80, 'E' << 16}
	cmp("rifiuto")
	// un'altra radio (comando di aggancio diverso): nessuna risposta, errore chiaro
	reset()
	e.nak, e.hello = -1, []byte("PROGROMBFHU")
	if testing.Short() {
		return
	}
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "non risponde") || len(logged(0)) != 0 {
		t.Fatalf("radio diversa: %v", err)
	}
}
