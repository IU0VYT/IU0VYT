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
	"strings"
	"testing"
)

func icomRun(e *icomEmu, fn func()) {
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return a, nil
	}
	fn()
}

// immagine di prova: memoria vuota del modello con alcuni canali
func icomTestImg(t *testing.T, m *genModel) []byte {
	o, _ := icomOptOf(m)
	img := bytes.Repeat([]byte{0xFF}, o.size)
	if x, ok := m.Opt.(*icx8xOpt); ok {
		img[0x1930] = x.uhf
	}
	lo := m.Bands[0].Lo + 9000000
	for i, c := range []*Chan{
		{Rx: lo, Tx: lo, Name: "S20", Scan: true},
		{Rx: lo + 100000, Tx: lo - 500000, TxTone: "88.5", Name: "R0", Scan: true, Power: m.Powers[len(m.Powers)-1].Sem},
		{Rx: lo + 112500, Tx: lo + 112500, TxTone: "D023I", RxTone: "D023N", Name: "DCS", Narrow: true},
		{Rx: lo + 200000, Tx: lo + 800000, TxTone: "123.0", RxTone: "123.0", Name: "TSQL", Scan: true},
	} {
		c.Power = m.powerSem(m.powerIndex(c.Power))
		if err := m.Fam.Set(m, img, i*5, c); err != nil {
			t.Fatalf("%s canale %d: %v", m.ID, i, err)
		}
		if g := m.Fam.Get(m, img, i*5); g == nil || *g != *c {
			t.Fatalf("%s canale %d riletto diverso: %+v / %+v", m.ID, i, g, c)
		}
	}
	return img
}

func TestIcomRecord(t *testing.T) {
	for _, id := range []string{"ic-v86", "ic-t10", "ic-v80", "ic-t70", "ic-v82", "ic-u82"} {
		m := genByID[id]
		img := icomTestImg(t, m)
		lo := m.Bands[0].Lo + 9000000
		for _, bad := range []*Chan{
			{Rx: lo, Tx: lo, RxTone: "88.5"},                   // solo tono in ricezione
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "100.0"},  // subtoni diversi
			{Rx: lo, Tx: lo, TxTone: "D023N", RxTone: "D025N"}, // codici diversi
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "D023N"},  // misti
		} {
			if err := m.Fam.Set(m, img, 50, bad); err == nil {
				t.Errorf("%s: toni non validi accettati: %+v", id, bad)
			}
		}
		if id != "ic-v86" && id != "ic-t10" {
			if err := m.Fam.Set(m, img, 50, &Chan{Rx: lo + 1000, Tx: lo + 1000}); err == nil {
				t.Errorf("%s: frequenza non memorizzabile accettata", id)
			}
			// multiplo di 6,25 kHz ma non di 12,5: memorizzato esatto (CHIRP lo arrotonda)
			c := &Chan{Rx: lo + 6250, Tx: lo + 6250, Scan: true, Power: m.Powers[0].Sem}
			if err := m.Fam.Set(m, img, 51, c); err != nil {
				t.Errorf("%s: %v", id, err)
			} else if g := m.Fam.Get(m, img, 51); g == nil || g.Rx != c.Rx {
				t.Errorf("%s: frequenza a 6,25 kHz riletta %+v", id, g)
			}
		}
		if m.Fam.Get(m, img, 50) != nil {
			t.Errorf("%s: un canale rifiutato ha lasciato tracce", id)
		}
		// spostamento e cancellazione
		raw := m.Fam.Rec(m, img, 5)
		m.Fam.PutRec(m, img, 60, raw)
		a, b := m.Fam.Get(m, img, 5), m.Fam.Get(m, img, 60)
		if a == nil || b == nil || *a != *b {
			t.Errorf("%s: canale spostato diverso", id)
		}
		if err := m.Fam.Set(m, img, 5, nil); err != nil || m.Fam.Get(m, img, 5) != nil {
			t.Errorf("%s: canale non cancellato", id)
		}
	}
	// IC-V80: trasmissione disabilitata
	m := genByID["ic-v80"]
	img := icomTestImg(t, m)
	c := &Chan{Rx: 145500000, Tx: 145500000, TxOff: true, Scan: true, Name: "RX"}
	if err := m.Fam.Set(m, img, 70, c); err != nil {
		t.Fatal(err)
	}
	if g := m.Fam.Get(m, img, 70); g == nil || !g.TxOff {
		t.Errorf("trasmissione disabilitata riletta %+v", g)
	}
}

func TestIcomCable(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-v86", "ic-v80", "ic-u82"} {
		m := genByID[id]
		o, _ := icomOptOf(m)
		img := icomTestImg(t, m)
		// lettura e scrittura, con il cavo a un filo e con quello senza eco
		for _, noEcho := range []bool{false, true} {
			e := newIcomEmu(o, img)
			e.noEcho = noEcho
			icomRun(e, func() {
				got, _, err := m.Fam.Download(ctx, m, "COM9", nop)
				if err != nil {
					t.Fatalf("%s: %v", id, err)
				}
				if !bytes.Equal(got, img) {
					t.Errorf("%s: immagine letta diversa", id)
				}
				mod := append([]byte{}, got...)
				m.Fam.Set(m, mod, 9, &Chan{Rx: m.Bands[0].Lo + 9500000, Tx: m.Bands[0].Lo + 9500000, Scan: true, Name: "NEW"})
				if _, err := m.Fam.Upload(ctx, m, "COM9", mod, nop); err != nil {
					t.Fatalf("%s: %v", id, err)
				}
				// si scrivono solo le zone previste: il resto resta com'era
				want := append([]byte{}, img...)
				for _, rg := range o.ranges {
					copy(want[rg[0]:rg[1]], mod[rg[0]:rg[1]])
				}
				if !bytes.Equal(e.image(), want) {
					t.Errorf("%s: memoria finale diversa", id)
				}
			})
		}
		// blocco con la somma di controllo sbagliata
		e := newIcomEmu(o, img)
		e.badSum = 3
		icomRun(e, func() {
			if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), "somma di controllo") {
				t.Errorf("%s, somma errata: %v", id, err)
			}
		})
		// la radio rifiuta i dati
		e = newIcomEmu(o, img)
		e.reject = true
		icomRun(e, func() {
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err == nil || !strings.Contains(err.Error(), "rifiutato") {
				t.Errorf("%s, rifiuto: %v", id, err)
			}
			if e.writes != 0 {
				t.Errorf("%s: dati scritti nonostante il rifiuto", id)
			}
		})
	}
	// modello sbagliato: la radio è una IC-T70, la scheda è la IC-V86
	m := genByID["ic-v86"]
	e := newIcomEmu(ict70Opt, nil)
	icomRun(e, func() {
		if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), "«Icom IC-T70»") {
			t.Errorf("modello sbagliato: %v", err)
		}
		if _, err := m.Fam.Upload(ctx, m, "COM9", make([]byte, icv86Opt.size), nop); err == nil || e.writes != 0 {
			t.Errorf("scrittura su un altro modello: %v", err)
		}
	})
	// IC-V82 e IC-U82: stessa copia, modello diverso nella memoria
	v, u := genByID["ic-v82"], genByID["ic-u82"]
	eu := newIcomEmu(icx8xClone, icomTestImg(t, u))
	icomRun(eu, func() {
		if _, _, err := v.Fam.Download(ctx, v, "COM9", nop); err == nil || !strings.Contains(err.Error(), "IC-U82") {
			t.Errorf("IC-U82 letta con la scheda IC-V82: %v", err)
		}
		if _, err := v.Fam.Upload(ctx, v, "COM9", icomTestImg(t, u), nop); err == nil {
			t.Error("immagine UHF scritta con la scheda VHF")
		}
	})
}
