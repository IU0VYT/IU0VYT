// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
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

// «Inoltra ad altra radio» (3.9.3) su una radio con posizioni bloccate (canali GMRS
// fissi della GM-15 Pro): le posizioni bloccate restano, i canali vanno solo dove
// la radio li accetta e il risultato è sempre scrivibile.
func TestPlanForwardLockedSlots(t *testing.T) {
	m := genByID["bf-gm15pro"]
	img := bytes.Repeat([]byte{0xFF}, gm30Size)
	for i := 0; i < m.Slots; i++ {
		_ = m.Fam.Set(m, img, i, nil)
	}
	for i := 0; i < 30; i++ { // canali GMRS fissi 1-30, come in una radio vera
		f := int64(462562500 + 25000*(i%7))
		if err := m.Fam.Set(m, img, i, &Chan{Rx: f, Tx: f, Scan: true}); err != nil {
			t.Fatal(err)
		}
	}
	cur := genChans(m, img)
	src := []*Chan{
		{Rx: 145500000, Tx: 145500000, Scan: true, Name: "VHF"},        // canali 55-250 solo in ricezione, 31-54 solo ponti GMRS: escluso
		{Rx: 462600000, Tx: 467600000, Scan: true, Name: "PONTE GMRS"}, // ponte GMRS: canale 31
	}
	for _, mode := range []string{"free", "replace"} {
		rep := &Report{}
		p, err := planForward(m.ID, cur, src, mode, img, nil, rep)
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if _, err := genWrite(m, img, p.chans, &Report{}); err != nil {
			t.Fatalf("%s: piano non scrivibile: %v", mode, err)
		}
		for i := 0; i < 54; i++ {
			if cur[i] != nil && p.chans[i] == nil {
				t.Fatalf("%s: canale bloccato %d cancellato", mode, i+1)
			}
		}
		w := strings.Join(rep.Warnings, " | ")
		if mode == "replace" && !strings.Contains(w, "non si possono cancellare") || !strings.Contains(w, "«VHF»") {
			t.Fatalf("segnalazioni: %v", rep.Warnings)
		}
		if len(p.placed) != 1 || p.placed[0] != 30 || p.skipped != 1 {
			t.Fatalf("%s: posizioni %v, esclusi %d", mode, p.placed, p.skipped)
		}
		t.Logf("%s: posizioni %v, esclusi %d, %v", mode, p.placed, p.skipped, rep.Warnings)
	}
}

func TestFmtPositions(t *testing.T) {
	if s := fmtPositions(kindXLC, []int{0, 1, 2, 5, 7, 8}); s != "1-3, 6, 8-9" {
		t.Fatal(s)
	}
	if s := fmtPositions(kindDAT, []int{62, 63, 64, 65}); s != "1/63-1/64, 2/1-2/2" {
		t.Fatal(s)
	}
}
