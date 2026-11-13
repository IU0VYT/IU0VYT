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
	"strconv"
	"testing"
)

// Impostazioni delle Retevis RB39 / RB639: un byte per voce nei dati di sistema,
// con i valori (non le posizioni) del programma del costruttore; il Bluetooth
// compare solo se la radio manda dieci byte; le impostazioni cambiate arrivano
// alla radio nel pacchetto 04/00 senza toccare gli altri byte.
func TestAuctusSettings(t *testing.T) {
	want := map[string]struct {
		idx  int
		vals []byte
	}{
		"tot": {0, []byte{6, 12, 18}}, "vox_level": {1, []byte{0, 1, 2, 3, 4, 5}}, "vox_delay": {2, []byte{1, 2, 3, 4, 5, 6}},
		"voice": {3, []byte{0, 2}}, "compander": {4, []byte{0, 1}}, "vox_on": {5, []byte{0, 1}}, "monitor": {6, []byte{0, 1}},
		"pf_long": {7, []byte{0, 1, 2, 3}}, "save": {8, []byte{0, 1}}, "bt": {9, []byte{0, 1}},
	}
	for _, id := range []string{"rt-rb39", "rt-rb639"} {
		if !settingsAvailable(id) {
			t.Fatalf("%s: impostazioni non disponibili", id)
		}
		for _, n := range []int{9, 10} {
			img := make([]byte, auctusSize)
			if settingsFor(id, img) != nil {
				t.Fatalf("%s: tabella anche senza dati di sistema", id)
			}
			img[auctusSys] = byte(n)
			for i := 1; i < 16; i++ {
				img[auctusSys+i] = 0xEE
			}
			ref := append([]byte(nil), img...)
			tb := settingsFor(id, img)
			if tb == nil || len(tb.Settings) != n {
				t.Fatalf("%s con %d byte: tabella %v", id, n, tb)
			}
			seen := map[string]bool{}
			for i := range tb.Settings {
				d := &tb.Settings[i]
				w, ok := want[d.Key]
				if !ok || seen[d.Key] || d.ItLabel == "" || d.ItDesc == "" || d.Group == "" {
					t.Fatalf("%s: impostazione inattesa, doppia o senza nome: %+v", id, d)
				}
				seen[d.Key] = true
				if w.idx >= n {
					t.Fatalf("%s: %s oltre i %d byte della radio", id, d.Key, n)
				}
				off := auctusSys + 1 + w.idx
				if _, err := readSetting(d, img); err == nil {
					t.Fatalf("%s %s: un valore non previsto (EE) viene letto senza errore", id, d.Key)
				}
				for k, v := range w.vals {
					val := strconv.Itoa(k)
					if d.Type == "bool" {
						val = strconv.FormatBool(k == 1)
					} else if len(d.Options) != len(w.vals) || len(d.ItOpts) != len(w.vals) {
						t.Fatalf("%s %s: %d scelte invece di %d", id, d.Key, len(d.Options), len(w.vals))
					}
					if err := writeSetting(d, img, val); err != nil {
						t.Fatalf("%s %s=%s: %v", id, d.Key, val, err)
					}
					if img[off] != v {
						t.Fatalf("%s %s=%s: byte %02X invece di %02X", id, d.Key, val, img[off], v)
					}
					if back, err := readSetting(d, img); err != nil || back != val {
						t.Fatalf("%s %s: riletto %q invece di %q (%v)", id, d.Key, back, val, err)
					}
				}
				img[off] = 0xEE
				if !bytes.Equal(img, ref) {
					t.Fatalf("%s %s: modificati byte fuori dalla propria posizione", id, d.Key)
				}
			}
		}
	}
	// valori iniziali del programma: TOT 60 s, guida vocale in inglese, ritardo 0,5 s,
	// compander e VOX accesi, tasto = monitor, risparmio acceso
	img := make([]byte, auctusSize)
	img[auctusSys] = 9
	copy(img[auctusSys+1:], []byte{6, 0, 1, 2, 1, 1, 0, 2, 1})
	got := map[string]string{}
	for i, tb := 0, settingsFor("rt-rb39", img); i < len(tb.Settings); i++ {
		v, err := readSetting(&tb.Settings[i], img)
		if err != nil {
			t.Fatalf("%s: %v", tb.Settings[i].Key, err)
		}
		got[tb.Settings[i].Key] = v
	}
	for k, v := range map[string]string{"tot": "0", "vox_level": "0", "vox_delay": "0", "voice": "1", "compander": "true", "vox_on": "true",
		"monitor": "false", "pf_long": "2", "save": "true"} {
		if got[k] != v {
			t.Fatalf("valori iniziali: %s = %q invece di %q", k, got[k], v)
		}
	}
}

// Dalla radio simulata alla radio simulata: una impostazione cambiata arriva nel
// pacchetto di sistema, gli altri byte restano quelli letti.
func TestAuctusSettingsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		id string
		n  int
	}{{"rt-rb39", 9}, {"rt-rb39", 10}, {"rt-rb639", 10}} {
		m := genByID[tc.id]
		radio := make([]byte, auctusSize)
		radio[auctusSys] = byte(tc.n)
		copy(radio[auctusSys+1:auctusSys+1+tc.n], []byte{12, 3, 4, 2, 0, 1, 0, 1, 1, 1})
		link := func(img []byte) (*auctusEmu, func()) {
			e := emuFor(m, img).(*auctusEmu)
			a, b := newPipe()
			old := openPortFunc
			openPortFunc = func(string, int) (Port, error) { return a, nil }
			go func() { _ = e.serve(b) }()
			return e, func() { openPortFunc = old; _ = a.Close() }
		}
		_, done := link(radio)
		img, _, err := m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
		done()
		if err != nil {
			t.Fatalf("%s: lettura: %v", tc.id, err)
		}
		tb := settingsFor(tc.id, img)
		if tb == nil || len(tb.Settings) != tc.n {
			t.Fatalf("%s: tabella con %d byte: %v", tc.id, tc.n, tb)
		}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			switch d.Key {
			case "tot":
				if v, err := readSetting(d, img); err != nil || v != "1" {
					t.Fatalf("%s: TOT letto %q (%v)", tc.id, v, err)
				}
				err = writeSetting(d, img, "2") // 180 s
			case "monitor":
				err = writeSetting(d, img, "true")
			}
			if err != nil {
				t.Fatalf("%s %s: %v", tc.id, d.Key, err)
			}
		}
		e, done := link(radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
		done()
		if err != nil {
			t.Fatalf("%s: scrittura: %v", tc.id, err)
		}
		waitFor(func() bool { e.mu.Lock(); defer e.mu.Unlock(); return len(e.log) == m.Slots+4 })
		after := e.image()
		want := append([]byte(nil), radio[auctusSys:auctusVer]...)
		want[1], want[7] = 18, 1
		if !bytes.Equal(after[auctusSys:auctusVer], want) {
			t.Fatalf("%s: dati di sistema nella radio % X invece di % X", tc.id, after[auctusSys:auctusVer], want)
		}
	}
}
