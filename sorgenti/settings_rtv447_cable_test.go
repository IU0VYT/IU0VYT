//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "testing"

// 4.47: le impostazioni dei 24 modelli Retevis con la memoria letta dal loro
// driver (radio simulata), non solo con una memoria costruita a mano: la tabella
// deve esserci tutta e nessuna voce deve stare oltre ciò che il driver legge.
// Per RT68B, RT668B e RB669, che leggono solo due intervalli, le voci devono
// stare dentro un intervallo letto.
func TestSettingsRetevisCable447(t *testing.T) {
	for id, defs := range rtvSetDefs {
		m := genByID[id]
		got := cableImage(t, m, noChirpBase(t, m))
		if got == nil {
			t.Fatalf("%s: nessuna radio simulata", id)
		}
		tb := settingsFor(id, got)
		if tb == nil || len(tb.Settings) != len(defs)+1 { // in coda c'è la password di programmazione
			t.Fatalf("%s: tabella assente o incompleta con la memoria letta via cavo (%d byte)", id, len(got))
		}
		for i := range tb.Settings {
			if end := setDefEnd(&tb.Settings[i]); end > len(got) {
				t.Fatalf("%s: la voce %s arriva a %d, oltre la memoria letta (%d)", id, tb.Settings[i].Key, end, len(got))
			}
		}
		if o, ok := m.Opt.(*t18Opt); ok && o.spec.rranges != nil {
			offs := map[int]string{}
			for _, d := range defs {
				offs[d.off] = d.key
			}
			for k := 0; k < rtvPwDefs[id].n; k++ {
				offs[rtvPwDefs[id].off+k] = "progpw"
			}
			for off, key := range offs {
				in := false
				for _, r := range o.spec.rranges {
					in = in || (off >= r[0] && off < r[1])
				}
				if !in {
					t.Fatalf("%s: la voce %s (0x%03X) sta in una parte che il driver non legge", id, key, off)
				}
			}
		}
	}
}
