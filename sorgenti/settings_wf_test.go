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

// Impostazioni di Radioddity UB15-G, DB-35G, DB35-UV e DB40-G: valori letti da
// una riga di prova e byte scritti per ogni scelta, calcolati a mano dalle
// posizioni del file di configurazione del costruttore (byte, bit di partenza,
// larghezza).
func TestWfSettings(t *testing.T) {
	// riga di prova: 23 2A 51 05
	//   byte 0 = 0x23: squelch 3 (bit 0-3), riduzione del rumore 2 (bit 4-7)
	//   byte 1 = 0x2A: TOT 2 (bit 0-2), tasti 1 (bit 3), roger 0 (bit 4), prioritario 1 (bit 5)
	//   byte 2 = 0x51: UB15-G V1.10: microfono 1 (bit 0-3), VOX 5 (bit 4-7)
	//                  DB35-UV: VOX 1 (bit 0-2), passo 2 (bit 3-5), luce 1 (bit 6-7)
	//   byte 3 = 0x05: DB35-UV: spegnimento 1 (bit 0-1), colore 1 (bit 2-4)
	row := []byte{0x23, 0x2A, 0x51, 0x05}
	type w struct {
		key, read string // valore letto dalla riga di prova ("!" = non previsto)
		set       string
		off       int
		want      byte
	}
	common := func(nrc w) []w {
		return []w{{"sql", "3", "9", 0, 0x29}, {"tot", "2", "3", 1, 0x2B}, nrc, {"pri_watch", "true", "false", 1, 0x0A},
			{"beep", "true", "false", 1, 0x22}, {"roger", "false", "true", 1, 0x3A}}
	}
	onOff := w{"nrc", "!", "true", 0, 0x10}    // 2 non è né spento né acceso
	level := w{"nrc_level", "2", "4", 0, 0x43} // livelli 1-5 = valori 0-4
	cases := map[string][]w{
		"rd-ub15g":     common(onOff),
		"rd-ub15g-110": append(common(onOff), w{"anc", "true", "false", 2, 0x50}, w{"vox_level", "5", "9", 2, 0x91}),
		"rd-db35g":     common(level),
		"rd-db35uv": append(common(level), w{"vox_level", "1", "4", 2, 0x54}, w{"step", "2", "6", 2, 0x71}, w{"bl_level", "1", "3", 2, 0xD1},
			w{"bl_time", "1", "0", 3, 0x04}, w{"bl_color", "1", "6", 3, 0x19}),
		// DB40-G: un byte per voce (blocco 1, tasti 0, TOT 4 = 120 s, squelch 7), come nel driver del CHIRP
		"rd-db40g": {{"sql", "7", "0", 3, 0}, {"tot", "4", "10", 2, 0x0A}, {"bcl", "1", "0", 0, 0}, {"beep", "0", "1", 1, 1}},
	}
	for id, ws := range cases {
		m := genByID[id]
		o := wfO(m)
		if id == "rd-db40g" {
			row = []byte{1, 0, 4, 7}
		} else {
			row = []byte{0x23, 0x2A, 0x51, 0x05}
		}
		img := make([]byte, o.size(m))
		if !settingsAvailable(id) || settingsFor(id, img[:len(img)-1]) != nil || settingsFor(id, img) != nil {
			t.Fatalf("%s: impostazioni assenti, o date a un'immagine corta o mai letta da una radio", id)
		}
		copy(img[o.param(m):], row)
		tb := settingsFor(id, img)
		if tb == nil || len(tb.Settings) != len(ws) || tb.Size != o.size(m) {
			t.Fatalf("%s: impostazioni %v invece di %d", id, tb, len(ws))
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
			at := o.param(m) + c.off
			if c.read == "!" {
				work[at] = 0
			}
			ref := append([]byte{}, work...)
			if err := writeSetting(d, work, c.set); err != nil {
				t.Fatalf("%s %s=%s: %v", id, c.key, c.set, err)
			}
			if work[at] != c.want {
				t.Fatalf("%s %s=%s: byte %02X invece di %02X", id, c.key, c.set, work[at], c.want)
			}
			if back, err := readSetting(d, work); err != nil || back != c.set {
				t.Fatalf("%s %s: riletto %q invece di %q (%v)", id, c.key, back, c.set, err)
			}
			work[at] = ref[at]
			if !bytes.Equal(work, ref) {
				t.Fatalf("%s %s: modificati altri byte", id, c.key)
			}
		}
	}
	// le scelte fuori elenco non si scrivono
	m := genByID["rd-ub15g"]
	img := make([]byte, wfO(m).size(m))
	img[wfO(m).param(m)] = 3
	tb := settingsFor("rd-ub15g", img)
	if err := writeSetting(&tb.Settings[1], img, "4"); err == nil {
		t.Fatal("TOT fuori elenco accettato")
	}
}
