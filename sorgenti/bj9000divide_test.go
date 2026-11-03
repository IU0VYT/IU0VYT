// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "testing"

// 4.3: BJ-9000 con «Divisione dei canali» = tutti i canali (960 senza zone),
// come la BJ-8300
func TestBJ9000Divide(t *testing.T) {
	d := genByID["bj-9000"].def()
	if d.Divide != "divide_ch" || d.DivideAll == "" || d.DivideZon == "" {
		t.Fatalf("definizione senza la scelta della divisione: %+v", d)
	}
	img := make([]byte, 0xD300)
	tb := settingsFor("bj-9000", img)
	var sd *setDef
	for i := range tb.Settings {
		if tb.Settings[i].Key == "divide_ch" {
			sd = &tb.Settings[i]
		}
	}
	if sd != nil && (len(sd.ItOpts) != 2 || sd.ItOpts[1] != "Tutti i canali (960)" || sd.Options[1] != d.DivideAll) {
		t.Fatalf("scelte della divisione: %q %q", sd.Options, sd.ItOpts)
	}
	if sd == nil {
		t.Fatal("impostazione divide_ch assente")
	}
	for _, c := range []struct {
		v    string
		want bool
	}{{d.DivideAll, true}, {d.DivideZon, false}, {d.DivideAll, true}} {
		if err := writeSetting(sd, img, c.v); err != nil {
			t.Fatal(err)
		}
		if got := genSingle("bj-9000", img); got != c.want {
			t.Fatalf("divide_ch=%s: tutti i canali %v", c.v, got)
		}
	}
}
