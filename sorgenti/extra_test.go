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
	"testing"
)

// TestEsterniEmu (4.9): modelli con un driver esterno a CHIRP (le prove con
// CHIRP si fanno solo con CHIRP_EXTRA): canali scritti e riletti, lettura e
// scrittura via cavo con la radio simulata.
func TestEsterniEmu(t *testing.T) {
	sizes := map[string]int{"rtl-rt890-m7ocm": 0xA000}
	for id, size := range sizes {
		m := genByID[id]
		t.Run(id, func(t *testing.T) {
			img := bytes.Repeat([]byte{0xFF}, size)
			chs := []*Chan{
				{Rx: 145500000, Tx: 145500000, Power: 0, Scan: true, Name: "S20"},
				{Rx: 438700000, Tx: 431100000, TxTone: "88.5", Power: 2, Narrow: true, Scan: true, Name: "RPT"},
				{Rx: 446006250, Tx: 446006250, TxTone: "D023N", RxTone: "D023I", Power: 2, Narrow: true, Name: "PMR1"},
			}
			for i, c := range chs {
				if err := m.Fam.Set(m, img, i, c); err != nil {
					t.Fatal(err)
				}
				if g := m.Fam.Get(m, img, i); g == nil || !chanEqual(g, c) {
					t.Fatalf("canale %d riletto diverso: %+v / %+v", i+1, g, c)
				}
			}
			e := emuFor(m, img)
			var got []byte
			goSession(t, m, e, func() error {
				var err error
				got, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
				return err
			})
			if !bytes.Equal(got, e.image()) {
				t.Fatalf("lettura: immagine diversa: %s", diffRanges(e.image(), got))
			}
			mod := append([]byte{}, got...)
			if err := m.Fam.Set(m, mod, 5, &Chan{Rx: 144800000, Tx: 144800000, Power: 0, Scan: true, Name: "APRS"}); err != nil {
				t.Fatal(err)
			}
			goSession(t, m, e, func() error {
				_, err := m.Fam.Upload(context.Background(), m, "EMU", mod, func(int, int, string) {})
				return err
			})
			if !bytes.Equal(e.image(), mod) {
				t.Fatalf("scrittura: memoria della radio diversa: %s", diffRanges(mod, e.image()))
			}
		})
	}
}
