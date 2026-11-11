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
	"encoding/binary"
	"testing"
)

// 4.23: prove dei Kenwood TK-481 / TK-981. CHIRP ha le immagini di prova di
// tutti e due e ogni modello usa la sua (quella della TK-981 ha nei sistemi 2 e
// 3 un record tutto a FF, con il numero 255, che né CHIRP né questo programma
// mostrano); quella della TK-481 resta come immagine della famiglia.
func init() {
	chirpFamilyImage["main.kwx481Fam"] = "Kenwood_TK-481.img"
}

func kwx481TestImg(typ string) []byte {
	img := bytes.Repeat([]byte{0xFF}, kwx80Size)
	copy(img[kwx80ID:], typ)
	// sistemi 1, 2 e 5 (il 3 e il 4 non ci sono); il 5 è a risorse condivise
	a := 0x300
	for _, i := range []int{0, 1, 4} {
		binary.LittleEndian.PutUint16(img[2*i:], uint16(a))
		copy(img[a:], []byte{byte(i + 1), 0, 0xFF})
		copy(img[a+3:], "Sistema   ")
		a += kwx481Rec
	}
	img[0x40] = 3
	copy(img[kwx481Trunk:], []byte{0x10, 0, 0, 0})
	return img
}

func TestKwx481Systems(t *testing.T) {
	m := genByID["kw-tk481"]
	f := m.Fam.(kwx481Fam)
	img := kwx481TestImg("PG481")
	want := map[int]*Chan{}
	check := func(what string) {
		t.Helper()
		for i := 0; i < m.Slots; i++ {
			g := f.Get(m, img, i)
			w := want[i]
			if (g == nil) != (w == nil) || (g != nil && *g != *w) {
				t.Fatalf("%s: posizione %d: %+v, atteso %+v", what, i, g, w)
			}
		}
		ss, ok := f.systems(img)
		if !ok {
			t.Fatalf("%s: tabella dei sistemi non leggibile", what)
		}
		// i sistemi restano uno dopo l'altro, senza buchi
		a := 0x300
		for i, s := range ss {
			if s.addr < 0 {
				continue
			}
			if s.addr != a || int(img[s.addr]) != i+1 {
				t.Fatalf("%s: sistema %d a 0x%04X, atteso 0x%04X", what, i+1, s.addr, a)
			}
			a += kwx481Rec * (s.count + 1)
		}
		if !bytes.Equal(img[a:], bytes.Repeat([]byte{0xFF}, len(img)-a)) {
			t.Fatalf("%s: dati dopo l'ultimo sistema", what)
		}
	}
	put := func(i int, c *Chan) {
		t.Helper()
		if err := f.Set(m, img, i, c); err != nil {
			t.Fatalf("posizione %d: %v", i, err)
		}
		if c == nil {
			delete(want, i)
		} else {
			x := *c
			x.Narrow = true
			want[i] = &x
		}
		check("dopo la posizione")
	}
	// canali in ordine sparso in due sistemi: gli altri non devono cambiare
	put(250+9, &Chan{Rx: 935012500, Tx: 896012500, Name: "B10", Scan: true})
	put(0, &Chan{Rx: 935250000, Tx: 935250000, RxTone: "88.5", TxTone: "D023N", Name: "A1", Power: 2})
	put(249, &Chan{Rx: 940000000, Tx: 940000000, TxOff: true, Name: "A250", Scan: true})
	put(250, &Chan{Rx: 936000000, Tx: 897000000, TxTone: "D754I", Name: "B1"})
	put(5, &Chan{Rx: 937006250, Tx: 898006250, Name: "A6", Scan: true, Power: 2})
	put(0, &Chan{Rx: 935262500, Tx: 896262500, Name: "A1 bis"})
	put(249, nil)
	put(250+9, nil)
	put(1, &Chan{Rx: 938000000, Tx: 899000000, Name: "A2"})
	// record come lo scrive CHIRP: A1 bis = canale 1 del sistema 1 (primo record)
	ss, _ := f.systems(img)
	r := ss[0].rec(img, 0)
	// 935.2625 → 93526250, 896.2625 → 89626250 (BCD little endian); passo 12,5 kHz (5); toni FFFF;
	// potenza alta (bit 0x08 acceso), talk around acceso (0x20), fuori scansione (0x04 spento)
	wantRec := []byte{1, 0x50, 0x62, 0x52, 0x93, 0x50, 0x62, 0x62, 0x89, 0x35, 0x35, 0xFF, 0xFF, 0xFF, 0xFF,
		'A', '1', ' ', 'b', 'i', 's', ' ', ' ', ' ', ' ', 0xFF, 0xFF, 0xFB, 0xFF, 0xFF, 0xFF, 0xFF}
	if !bytes.Equal(r, wantRec) {
		t.Errorf("record:\n got % X\nwant % X", r, wantRec)
	}
	// sistema che la radio non ha, sistema a risorse condivise: rifiutati
	if err := f.Set(m, img, 2*250, &Chan{Rx: 935000000, Tx: 935000000}); err == nil {
		t.Error("canale accettato in un sistema che non c'è")
	}
	if err := f.Set(m, img, 4*250, &Chan{Rx: 935000000, Tx: 935000000}); err == nil {
		t.Error("canale accettato in un sistema a risorse condivise")
	}
	if !f.lockedSlotImg(m, img, 2*250) || !f.lockedSlotImg(m, img, 4*250) || f.lockedSlotImg(m, img, 250) {
		t.Error("posizioni bloccate")
	}
	check("alla fine")
	// sistema pieno: 250 canali, il 251° non esiste nella lista; memoria piena rifiutata
	for n := 0; n < 250; n++ {
		if err := f.Set(m, img, 250+n, &Chan{Rx: 935000000 + int64(n)*12500, Tx: 896000000, Name: "X"}); err != nil {
			t.Fatalf("canale %d: %v", n+1, err)
		}
	}
	if ss, _ := f.systems(img); ss[1].count != 250 {
		t.Fatalf("sistema 2: %d canali", ss[1].count)
	}
	copy(img[len(img)-4:], []byte{1, 2, 3, 4})
	if err := f.Set(m, img, 9, &Chan{Rx: 935000000, Tx: 935000000}); err == nil {
		t.Error("canale accettato con la memoria piena")
	}
	// TK-981: senza livelli di potenza, il bit resta com'è
	m9 := genByID["kw-tk981"]
	img9 := kwx481TestImg("M0981")
	if err := f.Set(m9, img9, 0, &Chan{Rx: 935000000, Tx: 935000000, Scan: true}); err != nil {
		t.Fatal(err)
	}
	if ss, _ := f.systems(img9); ss[0].rec(img9, 0)[26] != 0xFF || ss[0].rec(img9, 0)[27] != 0xDF {
		t.Errorf("TK-981: byte 26-27 % X", ss[0].rec(img9, 0)[26:28])
	}
}
