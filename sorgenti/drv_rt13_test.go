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

// 4.23, Radtel RT-13: record del canale come lo scrive il programma del
// costruttore (writeInfo, FreqToData, SubAudioToData), valori calcolati a mano
func TestRT13Codec(t *testing.T) {
	m := genByID["rtl-rt13"]
	f := m.Fam.(rt13Fam)
	if m.Slots != 22 || m.First != 1 || rt13Base+22*16 != 0x0170 || len(rt13Hello) != 8 || rt13Hello[0] != 0x6A {
		t.Fatal("costanti del modello")
	}
	img := bytes.Repeat([]byte{0xFF}, rt13Size)
	cases := []struct {
		i    int
		c    *Chan
		want string
	}{
		// 446.00625 stretta, bassa, in scansione, senza toni: bit 0x04 e blocco spento (0x01)
		{0, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Narrow: true, Scan: true},
			"25066044" + "25066044" + "ffff" + "ffff" + "05" + "00" + "ffff"},
		// 433.500 / 431.900 larga, alta, fuori scansione, tono TX 88.5 (BCD 0885), RX 100.0 (BCD 1000)
		{1, &Chan{Rx: 433500000, Tx: 431900000, Power: 0, TxTone: "88.5", RxTone: "100.0"},
			"00003543" + "00001943" + "0010" + "8508" + "19" + "00" + "ffff"},
		// DCS: RX D023N (0x0023 | 0x8000), TX D754I (0x0754 | 0xC000); sola ricezione
		{21, &Chan{Rx: 400000000, Tx: 400000000, TxOff: true, Power: 0, Scan: true, RxTone: "D023N", TxTone: "D754I"},
			"00000040" + "ffffffff" + "2380" + "54c7" + "09" + "00" + "ffff"},
	}
	for _, k := range cases {
		if err := f.Set(m, img, k.i, k.c); err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(f.rec(img, k.i)); got != k.want || len(k.want) != 32 {
			t.Errorf("canale %d:\n got %s\nwant %s", k.i, got, k.want)
		}
		if g := f.Get(m, img, k.i); g == nil || *g != *k.c {
			t.Errorf("canale %d riletto diverso: %+v", k.i, g)
		}
	}
	if &f.rec(img, 0)[0] != &img[0x0010] || &f.rec(img, 21)[0] != &img[0x0160] {
		t.Error("posizione dei canali")
	}
	// salto di frequenza (0x80), «cambio di frequenza» (0x40), blocco su canale
	// occupato acceso (bit 0x01 a zero) e scrambler di un canale esistente restano
	r := f.rec(img, 0)
	r[12], r[13] = 0xC4, 5
	if g := f.Get(m, img, 0); g == nil || g.Power != 2 || !g.Narrow || !g.Scan {
		t.Errorf("canale 0 letto come %+v", g)
	}
	if err := f.Set(m, img, 0, &Chan{Rx: 446006250, Tx: 446006250, Power: 0, Scan: false}); err != nil {
		t.Fatal(err)
	}
	if r[12] != 0xC0|0x10|0x08 || r[13] != 5 || r[14] != 0xFF || r[15] != 0xFF {
		t.Errorf("canale 0 modificato: % X", r[12:])
	}
	// cancellazione: come scrive il programma un canale vuoto
	if err := f.Set(m, img, 1, nil); err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(f.rec(img, 1)) != strings.Repeat("ff", 12)+"0900ffff" || f.Get(m, img, 1) != nil {
		t.Errorf("canale 1 non cancellato: %x", f.rec(img, 1))
	}
	// canale ricreato: valori di partenza del programma, non quelli di prima
	if err := f.Set(m, img, 1, &Chan{Rx: 433500000, Tx: 433500000, Power: 2, Scan: true}); err != nil || f.rec(img, 1)[12] != 0x01 {
		t.Errorf("canale 1 ricreato: % X (%v)", f.rec(img, 1)[12:], err)
	}
	// toni come li legge il programma: secondo byte a zero = nessuno, FFFF = nessuno
	copy(f.rec(img, 2), f.rec(img, 1))
	copy(f.rec(img, 2)[8:12], []byte{0x85, 0x00, 0xFF, 0xFF})
	if g := f.Get(m, img, 2); g == nil || g.RxTone != "" || g.TxTone != "" {
		t.Errorf("toni vuoti: %+v", g)
	}
	// frequenza vuota come nel programma: quarto byte a FF
	if f.Get(m, img, 5) != nil {
		t.Error("canale vuoto visto come usato")
	}
	// rifiutati: frequenze fuori da 400-480 MHz, codice DCS non valido
	for _, c := range []*Chan{
		{Rx: 399999990, Tx: 433500000},
		{Rx: 433500000, Tx: 480000010},
		{Rx: 145500000, Tx: 145500000},
		{Rx: 433500000, Tx: 433500000, TxTone: "D019N"},
	} {
		if err := f.Set(m, img, 7, c); err == nil {
			t.Errorf("canale accettato: %+v", c)
		}
	}
	if f.Get(m, img, 7) != nil {
		t.Error("canale rifiutato scritto lo stesso")
	}
	for _, hz := range []int64{400000000, 480000000} {
		if err := f.Set(m, img, 8, &Chan{Rx: hz, Tx: hz}); err != nil {
			t.Errorf("%d Hz: %v", hz, err)
		}
	}
	// sola ricezione con la trasmissione fuori banda: conta solo la ricezione
	if err := f.Set(m, img, 9, &Chan{Rx: 433500000, Tx: 145000000, TxOff: true}); err != nil {
		t.Errorf("sola ricezione: %v", err)
	}
}

// lettura e scrittura con la radio simulata, nell'ordine del programma del
// costruttore
func TestRT13Clone(t *testing.T) {
	m := genByID["rtl-rt13"]
	f := m.Fam.(rt13Fam)
	mem := make([]byte, 0x0400)
	for i := range mem {
		mem[i] = byte(i*7) ^ byte(i>>5)
	}
	copy(mem[rt13Base:], bytes.Repeat([]byte{0xFF}, 22*16))
	copy(mem[rt13PwdAddr:], bytes.Repeat([]byte{0xFF}, 8))
	e := f.emu(m, mem).(*rt13Emu)
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
	e.noise = 2
	got, info, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != rt13Size || !strings.Contains(info, "P3107") {
		t.Fatalf("lettura: %d byte (%s)", len(got), info)
	}
	// le aree lette sono quelle della radio, il resto è a FF
	in := func(a int, areas [][2]int) bool {
		for _, ar := range areas {
			if a >= ar[0] && a < ar[1] {
				return true
			}
		}
		return false
	}
	for a := range got {
		if w := mem[a]; in(a, rt13ReadAreas) && got[a] != w || !in(a, rt13ReadAreas) && got[a] != 0xFF {
			t.Fatalf("byte 0x%04X = %02X", a, got[a])
		}
	}
	// ordine del programma: saluto, password, blocchi, 'E'
	want := []int{'H' << 16, 'R'<<16 | rt13PwdAddr}
	for _, ar := range [][2]int{{0x0000, 0x0170}, {0x02B0, 0x02C0}, {0x0330, 0x0340}, {0x0380, 0x03E0}} {
		for a := ar[0]; a < ar[1]; a += 8 {
			want = append(want, 'R'<<16|a)
		}
	}
	want = append(want, 'E'<<16)
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
	if len(want) != 2+46+2+2+12+1 {
		t.Fatalf("sequenza di %d comandi", len(want))
	}
	reset()
	e.noise = 0
	if err := f.Set(m, got, 3, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Scan: true, Narrow: true}); err != nil {
		t.Fatal(err)
	}
	// byte fuori dalle aree scritte: nel file sono diversi, non devono arrivare alla radio
	copy(got[0x0170:0x02B0], bytes.Repeat([]byte{0x33}, 0x02B0-0x0170))
	copy(got[0x0330:0x0340], bytes.Repeat([]byte{0x33}, 16))
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	// saluto, password, 'E', di nuovo il saluto, blocchi, 'E'
	want = []int{'H' << 16, 'R'<<16 | rt13PwdAddr, 'E' << 16, 'H' << 16}
	for _, ar := range [][2]int{{0x0000, 0x0170}, {0x02B0, 0x02C0}, {0x0380, 0x03E0}} {
		for a := ar[0]; a < ar[1]; a += 8 {
			want = append(want, 'W'<<16|a)
		}
	}
	want = append(want, 'E'<<16)
	cmp("scrittura")
	after := e.image()
	for a := range after {
		w := mem[a]
		if in(a, rt13WriteAreas) {
			w = got[a]
		}
		if after[a] != w {
			t.Fatalf("radio dopo la scrittura: byte 0x%04X = %02X, atteso %02X", a, after[a], w)
		}
	}
	if g := f.Get(m, after, 3); g == nil || g.Rx != 446006250 || !g.Narrow {
		t.Fatalf("canale 4 nella radio: %+v", g)
	}
	// blocco rifiutato: errore con l'indirizzo e uscita dalla programmazione
	reset()
	e.nak = 0x0018
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "0x0018") {
		t.Fatalf("blocco rifiutato: %v", err)
	}
	want = []int{'H' << 16, 'R'<<16 | rt13PwdAddr, 'E' << 16, 'H' << 16, 'W' << 16, 'W'<<16 | 8, 'W'<<16 | 0x10, 'W'<<16 | 0x18, 'E' << 16}
	cmp("rifiuto")
	e.nak = -1
	// la radio non conferma la fine: errore
	reset()
	e.noF = true
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil {
		t.Fatal("scrittura senza conferma accettata")
	}
	e.noF = false
	// password impostata: né lettura né scrittura, e nessun blocco scritto
	reset()
	e.mu.Lock()
	copy(e.mem[rt13PwdAddr:], "123456")
	e.mu.Unlock()
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("lettura con la password: %v", err)
	}
	logged(3) // l'uscita ('E') arriva alla radio simulata dopo la fine della funzione
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("scrittura con la password: %v", err)
	}
	want = []int{'H' << 16, 'R'<<16 | rt13PwdAddr, 'E' << 16, 'H' << 16, 'R'<<16 | rt13PwdAddr, 'E' << 16}
	cmp("password")
	e.mu.Lock()
	copy(e.mem[rt13PwdAddr:], bytes.Repeat([]byte{0xFF}, 8))
	e.mu.Unlock()
	// un'altra radio (saluto diverso): nessuna risposta, errore chiaro
	reset()
	e.hello = []byte("\x02PROGRAM")
	if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "non risponde") {
		t.Fatalf("radio sbagliata: %v", err)
	}
}
