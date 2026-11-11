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
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

// 4.23, Radtel RT-69 GPS: canale, nome e byte «valido» come li scrive il
// programma del costruttore (CheckValue, FormatValue, SaveChannelName,
// cmb_Main_SelectedIndexChanged), valori calcolati a mano
func TestRT69Codec(t *testing.T) {
	m := genByID["rtl-rt69"]
	f := m.Fam.(rt69Fam)
	if m.Slots != 999 || m.First != 1 || rt69Blocks != 0x1EE || rt69From+64*rt69Blocks != rt69Size {
		t.Fatal("costanti del modello")
	}
	img := bytes.Repeat([]byte{0xFF}, rt69Size)
	cases := []struct {
		i          int
		c          *Chan
		want, name string
	}{
		// 145.500 / 145.000, tono TX 88.5 (0x8000 + 885), alta, larga, in scansione
		{0, &Chan{Rx: 145500000, Tx: 145000000, TxTone: "88.5", Power: 0, Scan: true, Name: "R0"},
			"00de03f0" + "00dd40a0" + "0000" + "8375" + "06" + "21" + "ffff", "5230" + strings.Repeat("00", 10)},
		// 433.500 stretta, bassa, fuori scansione, DCS: RX D754I, TX D645N; nome di 12 caratteri
		{1, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D754I", TxTone: "D645N", Power: 2, Narrow: true, Name: "PONTE R0 ROM"},
			"029577f0" + "029577f0" + "61ec" + "41a5" + "04" + "00" + "ffff", "504f4e544520523020524f4d"},
		// 446.00625 media, larga, in scansione, sola ricezione, tono RX 254.1, D023N in TX
		{998, &Chan{Rx: 446006250, Tx: 446006250, TxOff: true, RxTone: "254.1", TxTone: "D023N", Power: 1, Scan: true, Name: "PMR 1"},
			"02a88d31" + "ffffffff" + "89ed" + "4013" + "05" + "21" + "ffff", "504d522031" + strings.Repeat("00", 7)},
	}
	for _, k := range cases {
		if err := f.Set(m, img, k.i, k.c); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, k.i)); got != k.want || len(k.want) != 32 {
			t.Errorf("canale %d:\n got %s\nwant %s", k.i, got, k.want)
		}
		if got := hex.EncodeToString(f.name1(img, k.i)) + hex.EncodeToString(f.name2(img, k.i)); got != k.name || len(k.name) != 24 {
			t.Errorf("nome %d:\n got %s\nwant %s", k.i, got, k.name)
		}
		if img[rt69Valid+k.i] != 0x9E {
			t.Errorf("canale %d non segnato come valido", k.i)
		}
		if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
			t.Errorf("canale %d riletto diverso: %+v", k.i, g)
		}
	}
	// posizioni del programma: canali da 0x0910, nomi da 0x4788 e 0x6FB0, validi da 0x6701
	if &f.rec(img, 0)[0] != &img[0x0910] || &f.rec(img, 998)[0] != &img[0x0910+998*16] ||
		&f.name1(img, 998)[0] != &img[0x4788+998*8] || &f.name2(img, 998)[0] != &img[0x6FB0+998*4] || rt69Valid+998 != 0x6AE7 {
		t.Error("posizioni di canali e nomi")
	}
	// le voci non gestite di un canale esistente restano: DCS «normale» (bit 0x04 a
	// zero), silenziamento QT*DTMF (bit 1-2 = 2) e gli altri bit dei byte 12 e 13
	r := f.rec(img, 0)
	r[12], r[13] = 0xF2, 0xD5
	if g := f.Get(m, img, 0); g == nil || g.Power != 0 || g.Narrow || g.Scan {
		t.Errorf("canale 0 letto come %+v", g)
	}
	if err := f.Set(m, img, 0, &Chan{Rx: 145500000, Tx: 145000000, Power: 2, Narrow: true, Scan: true, Name: "R0"}); err != nil {
		t.Fatal(err)
	}
	if r[12] != 0xF0 || r[13] != 0xF4 || binary.BigEndian.Uint16(r[10:12]) != 0 {
		t.Errorf("canale 0 modificato: % X", r[8:])
	}
	// potenza 3 (che il programma mostra come alta)
	r[12] = 0x07
	if g := f.Get(m, img, 0); g == nil || g.Power != 0 {
		t.Errorf("potenza 3: %+v", g)
	}
	// nome con caratteri non scrivibili: resta se non cambia
	copy(f.name1(img, 1), []byte{0xD6, 0xD0, 'A', 0, 0, 0, 0, 0})
	copy(f.name2(img, 1), []byte{0, 0, 0, 0})
	g := f.Get(m, img, 1)
	if g == nil {
		t.Fatal("canale 1")
	}
	g.Power = 0
	if err := f.Set(m, img, 1, g); err != nil || !bytes.Equal(f.name1(img, 1), []byte{0xD6, 0xD0, 'A', 0, 0, 0, 0, 0}) || f.rec(img, 1)[12] != 0x06 {
		t.Errorf("nome dopo una modifica: %x (%v)", f.name1(img, 1), err)
	}
	// cancellazione come «cancella riga» del programma
	if err := f.Set(m, img, 1, nil); err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(f.rec(img, 1)) != "ffffffffffffffff00000000ffffffff" || img[rt69Valid+1] != 0 || f.Get(m, img, 1) != nil ||
		!bytes.Equal(f.name1(img, 1), make([]byte, 8)) || !bytes.Equal(f.name2(img, 1), make([]byte, 4)) {
		t.Errorf("canale 1 non cancellato: %x", f.rec(img, 1))
	}
	// canale non valido anche se ha dei dati
	copy(f.rec(img, 5), f.rec(img, 0))
	img[rt69Valid+5] = 0
	if f.Get(m, img, 5) != nil {
		t.Error("canale non valido visto come usato")
	}
	// rifiutati: frequenze fuori da 66.025-599.995 MHz, DCS non in tabella, tono fuori da 60-254.1 Hz
	for _, c := range []*Chan{
		{Rx: 66024990, Tx: 145500000},
		{Rx: 145500000, Tx: 599995010},
		{Rx: 145500000, Tx: 145500000, TxTone: "D019N"},
		{Rx: 145500000, Tx: 145500000, RxTone: "59.9"},
	} {
		if err := f.Set(m, img, 7, c); err == nil {
			t.Errorf("canale accettato: %+v", c)
		}
	}
	if f.Get(m, img, 7) != nil || img[rt69Valid+7] == 0x9E {
		t.Error("canale rifiutato scritto lo stesso")
	}
	for _, hz := range []int64{66025000, 599995000} {
		if err := f.Set(m, img, 8, &Chan{Rx: hz, Tx: hz}); err != nil {
			t.Errorf("%d Hz: %v", hz, err)
		}
	}
	// limiti della radio (CheckFrequency): due bande a 32 bit, oppure fino a tre in BCD
	b1 := make([]byte, 16)
	b2 := make([]byte, 16)
	binary.BigEndian.PutUint32(b1[0:], 13600000)
	binary.BigEndian.PutUint32(b1[4:], 17400000)
	binary.BigEndian.PutUint32(b2[0:], 40000000)
	binary.BigEndian.PutUint32(b2[4:], 52000000)
	bands, err := rt69Limits(b1, b2, make([]byte, 13))
	if err != nil || len(bands) != 2 || bands[0] != (band{136000000, 174000000}) || bands[1] != (band{400000000, 520000000}) {
		t.Errorf("limiti a 32 bit: %v (%v)", bands, err)
	}
	b3, _ := hex.DecodeString("05" + "13601740" + "22002600" + "40005200")
	bands, err = rt69Limits(bytes.Repeat([]byte{0xFF}, 16), b2, b3)
	if err != nil || len(bands) != 2 || bands[0] != (band{136000000, 174000000}) || bands[1] != (band{400000000, 520000000}) {
		t.Errorf("limiti in BCD: %v (%v)", bands, err)
	}
	b3[1] = 0x1A
	if _, err := rt69Limits(bytes.Repeat([]byte{0xFF}, 16), b2, b3); err == nil {
		t.Error("limiti non in BCD accettati")
	}
}

// lettura e scrittura con la radio simulata, nell'ordine del programma del
// costruttore; sotto 0x0480 non si scrive mai
func TestRT69Clone(t *testing.T) {
	m := genByID["rtl-rt69"]
	f := m.Fam.(rt69Fam)
	base := make([]byte, rt69Size)
	for i := range base {
		base[i] = byte(i*3) ^ byte(i>>6)
	}
	copy(base[rt69Valid:], make([]byte, rt69Slots))
	copy(base[rt69PwdAt:], bytes.Repeat([]byte{0xFF}, 8))
	for i, c := range []*Chan{
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Scan: true},
		{Rx: 433500000, Tx: 433500000, Name: "U", Power: 2, Scan: true, Narrow: true},
	} {
		if err := f.Set(m, base, i*500, c); err != nil {
			t.Fatal(err)
		}
	}
	e := f.emu(m, base).(*rt69Emu)
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		if baud != 19200 {
			t.Errorf("velocità %d", baud)
		}
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	got, info, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[rt69From:], base[rt69From:]) || !bytes.Equal(got[:rt69From], bytes.Repeat([]byte{0xFF}, rt69From)) || !strings.Contains(info, "RT-69GPS") {
		t.Fatalf("lettura diversa (%s)", info)
	}
	// ordine: saluto, password, blocchi da 0x0480 a 0x7FC0, fine
	want := []int{wx8CmdID << 16, wx8CmdRd<<16 | rt69PwdAt}
	for a := rt69From; a < rt69Size; a += 64 {
		want = append(want, wx8CmdRd<<16|a)
	}
	want = append(want, wx8CmdEnd<<16)
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
	if len(want) != 2+0x1EE+1 {
		t.Fatalf("sequenza di %d comandi", len(want))
	}
	reset()
	if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Name: "PMR 1", Power: 2, Scan: true, Narrow: true}); err != nil {
		t.Fatal(err)
	}
	hidden := append([]byte{}, e.mem[:rt69From]...)
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	// saluto, password, i tre blocchi dei limiti, blocchi da 0x0480, fine
	want = []int{wx8CmdID << 16, wx8CmdRd<<16 | rt69PwdAt, wx8CmdRd<<16 | 0x44, wx8CmdRd<<16 | 0x54, wx8CmdRd<<16 | 0x2F0}
	for a := rt69From; a < rt69Size; a += 64 {
		want = append(want, wx8CmdWr<<16|a)
	}
	want = append(want, wx8CmdEnd<<16)
	cmp("scrittura")
	if !bytes.Equal(e.image(), got) || !bytes.Equal(e.mem[:rt69From], hidden) {
		t.Fatal("memoria della radio diversa da quella scritta")
	}
	if g := f.Get(m, e.image(), 7); g == nil || g.Name != "PMR 1" || g.Rx != 446006250 {
		t.Fatalf("canale 8 nella radio: %+v", g)
	}
	// radio con i limiti in BCD (136-174 e 400-520 MHz): il canale a 446 MHz passa
	e.mu.Lock()
	copy(e.mem[0x44:], bytes.Repeat([]byte{0xFF}, 32))
	b3, _ := hex.DecodeString("05" + "13601740" + "22002600" + "40005200")
	copy(e.mem[0x2F0:], b3)
	e.mu.Unlock()
	reset()
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	logged(len(want))
	// un canale fuori dalle bande della radio (220 MHz: la banda c'è ma è spenta):
	// errore con la frequenza e nessun blocco scritto
	bad := append([]byte{}, got...)
	if err := f.Set(m, bad, 9, &Chan{Rx: 220000000, Tx: 220000000}); err != nil {
		t.Fatal(err)
	}
	reset()
	if _, err := f.Upload(context.Background(), m, "COM7", bad, noProg); err == nil || !strings.Contains(err.Error(), "220.00000") {
		t.Fatalf("canale fuori banda: %v", err)
	}
	want = []int{wx8CmdID << 16, wx8CmdRd<<16 | rt69PwdAt, wx8CmdRd<<16 | 0x44, wx8CmdRd<<16 | 0x54, wx8CmdRd<<16 | 0x2F0, wx8CmdEnd << 16}
	cmp("fuori banda")
	if !bytes.Equal(e.image(), got) {
		t.Fatal("scrittura nonostante il canale fuori banda")
	}
	// blocco rifiutato: errore con l'indirizzo e uscita dalla programmazione
	reset()
	e.nak = rt69From + 0x80
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "0x0500") {
		t.Fatalf("blocco rifiutato: %v", err)
	}
	want = append(want[:5], wx8CmdWr<<16|rt69From, wx8CmdWr<<16|(rt69From+0x40), wx8CmdWr<<16|(rt69From+0x80), wx8CmdEnd<<16)
	cmp("rifiuto")
	e.nak = -1
	// password impostata: né lettura né scrittura
	e.mu.Lock()
	copy(e.mem[rt69PwdAt:], "123456\x00\x00")
	e.mu.Unlock()
	reset()
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("lettura con la password: %v", err)
	}
	logged(3)
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("scrittura con la password: %v", err)
	}
	want = []int{wx8CmdID << 16, wx8CmdRd<<16 | rt69PwdAt, wx8CmdEnd << 16, wx8CmdID << 16, wx8CmdRd<<16 | rt69PwdAt, wx8CmdEnd << 16}
	cmp("password")
	e.mu.Lock()
	copy(e.mem[rt69PwdAt:], bytes.Repeat([]byte{0xFF}, 8))
	e.mu.Unlock()
	// risposte con il terzo byte diverso da zero (Wouxun KG-UV8D): non è questa radio
	e.dir = 0xFF
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil {
		t.Fatal("risposta di un'altra radio accettata")
	}
	e.dir = 0
	// radio muta: errore chiaro
	e.deaf = true
	rt69TimeoutTest := time.Now()
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "non risponde") {
		t.Fatalf("radio muta: %v", err)
	}
	if time.Since(rt69TimeoutTest) > 10*time.Second {
		t.Error("attesa troppo lunga")
	}
}
