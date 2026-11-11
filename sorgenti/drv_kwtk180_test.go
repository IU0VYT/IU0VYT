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
	"fmt"
	"testing"
)

// 4.21: prove dei Kenwood TK-x180. CHIRP ha le immagini di prova della TK-3180K2 e
// della TK-8180: per tutti i modelli si usa la prima con la loro sigla.
func init() {
	chirpFamilyImage["main.kwx180Fam"] = "Kenwood_TK-3180K2.img"
	for _, m := range genModels {
		o, ok := m.Opt.(*kwx180Opt)
		if !ok {
			continue
		}
		id := o.id
		chirpImagePatch[m.ID] = func(b []byte) { copy(b[kwx180ID:], id) }
	}
}

// memoria di prova: tre zone convenzionali vuote
func kwx180TestImg(id string) []byte {
	img := bytes.Repeat([]byte{0xFF}, kwx180Size)
	copy(img[kwx180ID:], id)
	for z := 0; z < 3; z++ {
		a := kwx180Sys + z*kwx180Hdr
		binary.LittleEndian.PutUint16(img[kwx180Starts+2*z:], uint16(a))
		h := img[a : a+kwx180Hdr]
		copy(h, make([]byte, kwx180Hdr))
		h[0], h[1], h[4] = byte(z+1), 0x31, 0
		binary.LittleEndian.PutUint16(h[2:], 1)
		copy(h[5:17], fmt.Sprintf("Zona %d", z+1))
	}
	copy(img[kwx180Skip:kwx180Skip+64], make([]byte, 64))
	return img
}

// 4.21, Kenwood TK-x180: aggiunte e cancellazioni spostano le zone che seguono
// senza cambiare gli altri canali; limiti e rifiuti
func TestKwx180Zones(t *testing.T) {
	m := genByID["kw-tk3180k2"]
	f := m.Fam.(kwx180Fam)
	img := kwx180TestImg("P3180\x07")
	want := map[int]*Chan{}
	put := func(i int, c *Chan) {
		t.Helper()
		if err := f.Set(m, img, i, c); err != nil {
			t.Fatalf("posizione %d: %v", i, err)
		}
		if c == nil {
			delete(want, i)
		} else {
			want[i] = c
		}
		zs, ok := f.zones(img)
		if !ok || len(zs) != 3 {
			t.Fatalf("zone non più leggibili dopo la posizione %d", i)
		}
		// zone una dopo l'altra, bit di scansione in fila, canali in ordine di numero
		addr, scan, n := kwx180Sys, 1, 0
		for _, z := range zs {
			if z.addr != addr || z.flagOff(img) != scan {
				t.Fatalf("zona a 0x%X (attesa 0x%X), primo bit %d (atteso %d)", z.addr, addr, z.flagOff(img), scan)
			}
			for k := 1; k < z.count; k++ {
				if z.rec(img, k-1)[0] >= z.rec(img, k)[0] {
					t.Fatalf("canali della zona non in ordine")
				}
			}
			addr += kwx180Hdr + z.count*kwx180Mem
			scan += z.count
			n += z.count
		}
		if n != len(want) {
			t.Fatalf("%d canali in memoria, attesi %d", n, len(want))
		}
		for j, w := range want {
			if g := f.Get(m, img, j); g == nil || *g != *w {
				t.Fatalf("dopo la posizione %d il canale %d è %+v, atteso %+v", i, j, g, w)
			}
		}
	}
	ch := func(k int) *Chan {
		return &Chan{Rx: 440000000 + int64(k)*12500, Tx: 445000000 + int64(k)*12500, TxTone: "88.5", Name: fmt.Sprintf("CH %d", k),
			Power: m.Powers[k%2].Sem, Narrow: k%3 == 0, Scan: k%4 != 0}
	}
	// zona 2 per prima, poi zona 1 e zona 3, numeri non in ordine
	for k, i := range []int{250 + 9, 250 + 2, 0 + 249, 0 + 0, 500 + 5, 250 + 4, 0 + 100, 500 + 0} {
		put(i, ch(k))
	}
	// modifica senza spostamenti, cancellazioni nella prima, nell'ultima e in mezzo
	put(250+2, &Chan{Rx: 446006250, Tx: 446006250, TxOff: true, RxTone: "D023I", Name: "PMR", Power: 0, Scan: false})
	put(0+0, nil)
	put(500+5, nil)
	put(250+4, nil)
	put(0+0, ch(20))
	// cancellare una posizione vuota o di una zona che non c'è non fa nulla
	before := append([]byte{}, img...)
	if err := f.Set(m, img, 0+7, nil); err != nil || !bytes.Equal(img, before) {
		t.Fatalf("cancellazione di una posizione vuota: %v", err)
	}
	if err := f.Set(m, img, 3*250+1, nil); err != nil || !bytes.Equal(img, before) {
		t.Fatalf("cancellazione in una zona che non c'è: %v", err)
	}
	// zona che la radio non ha: rifiutata e bloccata
	if err := f.Set(m, img, 3*250, ch(1)); err == nil || !bytes.Equal(img, before) {
		t.Fatalf("canale in una zona che non c'è: %v", err)
	}
	if !f.lockedSlotImg(m, img, 3*250) || f.lockedSlotImg(m, img, 3*250-1) || f.LockedImg(m, img, 3*250, nil, ch(1)) == nil {
		t.Fatal("posizioni bloccate")
	}
	// 512 canali in tutto
	for k := 0; len(want) < kwx180Max; k++ {
		if _, ok := want[k]; !ok {
			put(k, ch(k))
		}
	}
	before = append([]byte{}, img...)
	free := 0
	for want[free] != nil {
		free++
	}
	if err := f.Set(m, img, free, ch(1)); err == nil || !bytes.Equal(img, before) {
		t.Fatalf("canale numero 513: %v", err)
	}
	// zona non convenzionale: niente lettura dei canali, niente modifiche
	zs, _ := f.zones(img)
	img[zs[1].addr+1] = 0x32
	if f.Get(m, img, 0) != nil || f.Set(m, img, 0, ch(1)) == nil || !f.lockedSlotImg(m, img, 0) {
		t.Fatal("zona non convenzionale accettata")
	}
}
