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
	"math/rand"
	"testing"
)

// TestSettingsRdty (4.10): impostazioni di RD-5R e DM-1801: ogni valore scritto
// si rilegge uguale e cambia solo i suoi bit; nome e ID DMR uguali a quelli
// letti da «Identifica» (dmrInfo, confrontato con dmrconfig in TestIdentifyDMRInfo)
func TestSettingsRdty(t *testing.T) {
	for _, id := range []string{"bf-rd5r", "bf-dm1801"} {
		m := genByID[id]
		img := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
		copy(img, rdtyo(m).hdr)
		// valori iniziali come li scrive il programma del costruttore
		for i := rdtySetBase + 8; i < rdtySetBase+0x28; i++ {
			img[i] = 0
		}
		img[0x7518] = 0
		tb := settingsFor(id, img)
		if tb == nil || len(tb.Settings) < 25 {
			t.Fatalf("%s: impostazioni assenti", id)
		}
		rnd := rand.New(rand.NewSource(7))
		for round := 0; round < 40; round++ {
			for i := range tb.Settings {
				d := &tb.Settings[i]
				cur, _ := readSetting(d, img)
				v, ok := randSetVal(rnd, d, cur)
				if !ok {
					continue
				}
				before := append([]byte{}, img...)
				if err := writeSetting(d, img, v); err != nil {
					t.Fatalf("%s/%s = %q: %v", id, d.Key, v, err)
				}
				got, err := readSetting(d, img)
				if err != nil || got != v {
					t.Fatalf("%s/%s: scritto %q, letto %q (%v)", id, d.Key, v, got, err)
				}
				// cambiano solo i byte dell'impostazione
				for k := range img {
					if img[k] != before[k] && !setTouches(d, k) {
						t.Fatalf("%s/%s: cambiato il byte 0x%X", id, d.Key, k)
					}
				}
			}
		}
		set := func(k, v string) {
			for i := range tb.Settings {
				if tb.Settings[i].Key == k {
					if err := writeSetting(&tb.Settings[i], img, v); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		set("radio_name", "IU0VYT")
		set("radio_id", "2221234")
		r := &identResult{}
		dmrInfo(m, img, r)
		got := map[string]string{}
		for _, f := range r.Fields {
			got[f.L] = f.V
		}
		if got["ID DMR"] != "2221234" || got["Nome della radio"] != "IU0VYT" {
			t.Fatalf("%s: identificazione %v", id, got)
		}
		if !bytes.Equal(img[rdtySetBase+8:rdtySetBase+12], []byte{0x02, 0x22, 0x12, 0x34}) {
			t.Fatalf("%s: ID in memoria % X", id, img[rdtySetBase+8:rdtySetBase+12])
		}
	}
}

func TestSettingsDM1701(t *testing.T) {
	m := genByID["bf-dm1701"]
	img := tytSynthImage()
	tb := settingsFor("bf-dm1701", img)
	if tb == nil {
		t.Fatal("impostazioni assenti")
	}
	rnd := rand.New(rand.NewSource(3))
	for round := 0; round < 60; round++ {
		for i := range tb.Settings {
			d := &tb.Settings[i]
			v, ok := randSetVal(rnd, d, "")
			if !ok {
				t.Fatalf("%s: nessun valore di prova", d.Key)
			}
			before := append([]byte{}, img...)
			if err := writeSetting(d, img, v); err != nil {
				t.Fatalf("%s = %q: %v", d.Key, v, err)
			}
			if got, err := readSetting(d, img); err != nil || got != v {
				t.Fatalf("%s: scritto %q, letto %q (%v)", d.Key, v, got, err)
			}
			for k := range img {
				if img[k] != before[k] && !setTouches(d, k) && !(d.Enc.Fmt == "utf16le" && setTouches(d, k-1)) {
					t.Fatalf("%s: cambiato il byte 0x%X", d.Key, k)
				}
			}
		}
	}
	for i := range tb.Settings {
		switch tb.Settings[i].Key {
		case "radio_name":
			_ = writeSetting(&tb.Settings[i], img, "IU0VYT Daniele")
		case "radio_id":
			_ = writeSetting(&tb.Settings[i], img, "2221234")
		}
	}
	r := &identResult{}
	dmrInfo(m, img, r)
	got := map[string]string{}
	for _, f := range r.Fields {
		got[f.L] = f.V
	}
	if got["ID DMR"] != "2221234" || got["Nome della radio"] != "IU0VYT Daniele" {
		t.Fatalf("identificazione %v", got)
	}
}

// setTouches: il byte k appartiene all'impostazione d
func setTouches(d *setDef, k int) bool {
	e := d.Enc
	for _, o := range e.Offs {
		if o == k {
			return true
		}
	}
	offs, masks, _ := e.fields()
	for i, o := range offs {
		if k >= o && k < o+len(masks[i]) {
			return true
		}
	}
	return false
}
