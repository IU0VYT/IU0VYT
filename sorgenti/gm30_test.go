// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"strings"
	"testing"
)

// regole dei canali GMRS della GM-15 Pro (come le modifiche ammesse da CHIRP)
func TestGM15ProRules(t *testing.T) {
	m := genByID["bf-gm15pro"]
	img := bytes.Repeat([]byte{0xFF}, gm30Size)
	for i := 0; i < m.Slots; i++ {
		_ = m.Fam.Set(m, img, i, nil)
	}
	set := func(i int, rx, tx int64, off bool) {
		if err := m.Fam.Set(m, img, i, &Chan{Rx: rx, Tx: tx, TxOff: off, Scan: true}); err != nil {
			t.Fatal(err)
		}
	}
	set(0, 462562500, 462562500, false)
	set(9, 467587500, 467587500, true)
	set(30, 462550000, 467550000, false)
	set(60, 145500000, 145500000, true)
	chs := genChans(m, img)
	try := func(i int, c *Chan, want string) {
		t.Helper()
		x := append([]*Chan{}, chs...)
		x[i] = c
		_, err := genWrite(m, img, x, &Report{})
		if want == "" && err != nil || want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
			t.Fatalf("canale %d: errore %v, atteso %q", i+1, err, want)
		}
	}
	try(0, &Chan{Rx: 462587500, Tx: 462587500, Scan: true}, "fissi")
	try(0, nil, "cancellare")
	try(0, &Chan{Rx: 462562500, Tx: 462562500, Scan: true, Name: "CH1"}, "")
	try(9, &Chan{Rx: 467587500, Tx: 467587500, Scan: true}, "solo in ricezione")
	try(30, &Chan{Rx: 446006250, Tx: 451006250, Scan: true}, "ponti")
	try(30, &Chan{Rx: 462600000, Tx: 467600000, Scan: true}, "")
	try(30, &Chan{Rx: 462600000, Tx: 462600000, Scan: true}, "+ 5 MHz")
	try(60, &Chan{Rx: 145500000, Tx: 145500000, Scan: true}, "solo in ricezione")
	try(60, &Chan{Rx: 433500000, Tx: 433500000, TxOff: true, TxTone: "D754I", RxTone: "88.5", Scan: true, Name: "PROVA"}, "")
	try(60, nil, "")
	try(100, &Chan{Rx: 145500000, Tx: 145500000, TxOff: true, Scan: true}, "")
}
