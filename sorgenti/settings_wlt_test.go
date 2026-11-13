// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"testing"
)

// Impostazioni di Retevis C6, C66 e Radioddity Echo: valori letti dal blocco
// predefinito del programma del costruttore e byte scritti per ogni scelta,
// calcolati a mano dal codice del programma.
func TestWltSettings(t *testing.T) {
	// blocco predefinito (programDataDefault): squelch 3, voce 1, 0x40, 5, TOT 12, risparmio 4, tasti 9 4 2 1
	def := []byte{3, 1, 0x40, 5, 0x0C, 0xFF, 4, 0xFF, 0xFF, 0xFF, 9, 4, 2, 1, 0, 0xFF}
	type w struct {
		key, read string // valore letto dal blocco predefinito ("!" = non previsto)
		set       string // valore da scrivere
		off       int
		want      byte // byte dopo la scrittura
	}
	cases := map[string][]w{
		"rt-c6": {
			{"sql", "3", "9", 0, 9}, {"voice", "1", "0", 1, 0}, {"r_tone", "2", "3", 2, 0x60}, {"roger", "0", "3", 2, 0x4C},
			{"pair", "true", "false", 2, 0x41}, {"clone", "true", "false", 2, 0x42}, {"vox_on", "false", "true", 3, 0x15}, {"vox_level", "5", "9", 3, 0x09},
			{"tot", "11", "0", 4, 1}, {"save", "4", "0", 6, 0}, {"pf1_short", "8", "0", 10, 0}, {"pf1_long", "4", "8", 11, 9},
			{"pf2_short", "2", "7", 12, 7}, {"pf2_long", "1", "5", 13, 5},
		},
		"rt-c66": {
			{"sql", "3", "0", 0, 0}, {"voice", "1", "0", 1, 0}, {"r_tone", "2", "0", 2, 0x00}, {"roger", "0", "2", 2, 0x48},
			{"pair", "true", "false", 2, 0x41}, {"clone", "true", "false", 2, 0x42}, {"vox_on", "false", "true", 3, 0x15}, {"vox_level", "5", "0", 3, 0x00},
			{"tot", "12", "20", 4, 20}, {"save", "4", "1", 6, 1}, {"pf1_short", "!", "8", 10, 8}, {"pf1_long", "4", "0", 11, 0},
			{"pf2_short", "2", "3", 12, 3}, {"pf2_long", "1", "6", 13, 6},
		},
		"rd-echo": {
			{"sql", "3", "5", 0, 5}, {"voice", "1", "2", 1, 2}, {"vox_level", "5", "0", 3, 0}, {"vox_delay", "0", "10", 14, 10},
			{"tot", "12", "1", 4, 1}, {"save", "4", "10", 6, 10}, {"pf1_short", "!", "4", 10, 4}, {"pf1_long", "4", "3", 11, 3},
			{"pf2_short", "2", "0", 12, 0}, {"pf2_long", "1", "2", 13, 2},
		},
	}
	for id, ws := range cases {
		m := genByID[id]
		o := wltO(m)
		if !settingsAvailable(id) || settingsFor(id, make([]byte, o.size()-1)) != nil {
			t.Fatalf("%s: impostazioni assenti o date anche a un'immagine corta", id)
		}
		img := bytes.Repeat([]byte{0xFF}, o.size())
		copy(img[o.set:], def)
		tb := settingsFor(id, img)
		if tb == nil || len(tb.Settings) != len(ws) {
			t.Fatalf("%s: %d impostazioni invece di %d", id, len(tb.Settings), len(ws))
		}
		by := map[string]*setDef{}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			if by[d.Key] != nil || d.ItLabel == "" || d.ItDesc == "" || d.Group == "" || (d.Type == "list" && len(d.Options) != len(d.ItOpts)) {
				t.Fatalf("%s: impostazione doppia o incompleta: %+v", id, d)
			}
			by[d.Key] = d
		}
		for _, c := range ws {
			d := by[c.key]
			if d == nil {
				t.Fatalf("%s: manca %s", id, c.key)
			}
			got, err := readSetting(d, img)
			if c.read == "!" {
				if err == nil {
					t.Fatalf("%s %s: valore non previsto letto come %q", id, c.key, got)
				}
			} else if err != nil || got != c.read {
				t.Fatalf("%s %s: letto %q invece di %q (%v)", id, c.key, got, c.read, err)
			}
			work := append([]byte{}, img...)
			if c.read == "!" {
				work[o.set+c.off] = 0
			}
			ref := append([]byte{}, work...)
			if err := writeSetting(d, work, c.set); err != nil {
				t.Fatalf("%s %s=%s: %v", id, c.key, c.set, err)
			}
			if work[o.set+c.off] != c.want {
				t.Fatalf("%s %s=%s: byte %02X invece di %02X", id, c.key, c.set, work[o.set+c.off], c.want)
			}
			if back, err := readSetting(d, work); err != nil || back != c.set {
				t.Fatalf("%s %s: riletto %q invece di %q (%v)", id, c.key, back, c.set, err)
			}
			work[o.set+c.off] = ref[o.set+c.off]
			if !bytes.Equal(work, ref) {
				t.Fatalf("%s %s: modificati altri byte", id, c.key)
			}
		}
	}
}
