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
	"strconv"
	"testing"
)

// 4.26: impostazioni delle BF-T17 / BF-T18, GT-18 e FS-T3 / PR-T3. Posizioni e
// numero delle scelte trascritti da readInfo / writeInfo e dalle risorse delle
// finestre dei quattro programmi dei costruttori.
func TestT17Settings(t *testing.T) {
	type pos struct {
		off  int
		mask byte
		n    int
	}
	base := map[string]pos{"voice": {0x630, 0xFF, 2}, "lang": {0x631, 0xFF, 2}, "scan_on": {0x632, 0xFF, 2}, "vox_on": {0x633, 0xFF, 2},
		"vox_level": {0x634, 0xFF, 5}, "r_tone": {0x635, 0xFF, 3}, "low_tx": {0x636, 0xFF, 2}, "high_tx": {0x637, 0xFF, 2},
		"beep": {0x640, 0x01, 2}, "save": {0x640, 0x02, 2}, "sql": {0x641, 0xFF, 10}, "alarm_mode": {0x642, 0xFF, 2},
		"tot": {0x643, 0xFF, 11}, "tail": {0x644, 0xFF, 2}}
	for id, txEnd := range map[string]byte{"bf-t17pmr": 0x04, "bf-t17frs": 0x80, "bf-gt18pmr": 0x04, "rd-t3": 0} {
		want := map[string]pos{}
		for k, v := range base {
			want[k] = v
		}
		if txEnd != 0 {
			want["tx_end"] = pos{0x640, txEnd, 2}
		}
		if !settingsAvailable(id) {
			t.Fatalf("%s: impostazioni non disponibili", id)
		}
		img := bytes.Repeat([]byte{0xFF}, t17Memsize)
		tb := settingsFor(id, img)
		if tb == nil || len(tb.Settings) != len(want) {
			t.Fatalf("%s: %d impostazioni invece di %d", id, len(tb.Settings), len(want))
		}
		// i blocchi delle impostazioni sono tra quelli scritti dal driver
		in := func(a int) bool {
			for _, g := range t17o(genByID[id]).gaps {
				if a >= g[0] && a < g[1] {
					return true
				}
			}
			return false
		}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			w, ok := want[d.Key]
			if !ok || d.ItLabel == "" || d.ItDesc == "" || d.Group == "" {
				t.Fatalf("%s: impostazione inattesa o senza nome: %+v", id, d)
			}
			offs, masks, err := d.Enc.fields()
			if err != nil || len(offs) != 1 || offs[0] != w.off || masks[0][0] != w.mask || !in(offs[0]) {
				t.Fatalf("%s %s: posizione %#x maschera %X (attese %#x %X)", id, d.Key, offs, masks, w.off, w.mask)
			}
			n := len(d.Options)
			if d.Type == "bool" {
				n = 2
			}
			if n != w.n {
				t.Fatalf("%s %s: %d scelte invece di %d", id, d.Key, n, w.n)
			}
			sh := 0
			for w.mask>>sh&1 == 0 {
				sh++
			}
			img[offs[0]] = 0xFF
			for v := 0; v < n; v++ {
				val := strconv.Itoa(v)
				if d.Type == "bool" {
					val = strconv.FormatBool(v == 1)
				}
				if err := writeSetting(d, img, val); err != nil {
					t.Fatalf("%s %s=%s: %v", id, d.Key, val, err)
				}
				if got := int(img[offs[0]]&w.mask) >> sh; got != v {
					t.Fatalf("%s %s=%s: byte %02X", id, d.Key, val, img[offs[0]])
				}
				if img[offs[0]]|w.mask != 0xFF {
					t.Fatalf("%s %s: bit fuori dalla maschera modificati (%02X)", id, d.Key, img[offs[0]])
				}
				if back, err := readSetting(d, img); err != nil || back != val {
					t.Fatalf("%s %s: riletto %q invece di %q (%v)", id, d.Key, back, val, err)
				}
			}
		}
		// valori di una radio vera (immagine di prova dei programmi): squelch 5,
		// TOT 180 s (scelta 6), VOX livello 4, guida vocale accesa in inglese
		copy(img[0x630:], []byte{0x01, 0x00, 0x01, 0x00, 0x03, 0x00, 0x00, 0x00})
		copy(img[0x640:], []byte{0x03, 0x05, 0x00, 0x06, 0x01})
		for key, v := range map[string]string{"sql": "5", "tot": "6", "vox_level": "3", "voice": "true", "lang": "0", "beep": "true", "save": "true", "tail": "true", "vox_on": "false"} {
			for i := range tb.Settings {
				if d := &tb.Settings[i]; d.Key == key {
					if got, err := readSetting(d, img); err != nil || got != v {
						t.Fatalf("%s %s: letto %q invece di %q (%v)", id, key, got, v, err)
					}
				}
			}
		}
		for i := range tb.Settings {
			if d := &tb.Settings[i]; d.Key == "tot" && (d.ItOpts[0] != "OFF" || d.ItOpts[1] != "30 s" || d.ItOpts[10] != "300 s") {
				t.Fatalf("%s: TOT %v", id, d.ItOpts)
			}
		}
	}
	// la BF-T9 (altro programma) non ha queste impostazioni
	if settingsAvailable("bf-t9") {
		t.Fatal("BF-T9: impostazioni non verificate sul suo programma")
	}
}
