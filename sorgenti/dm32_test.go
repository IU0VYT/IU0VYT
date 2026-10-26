//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// DM-32UV: cancellazioni che toccano zone, liste di scansione e numero dei canali,
// canali DMR e un canale nuovo oltre l'ultimo: immagine identica a quella di CHIRP.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDm32ChirpEdits(t *testing.T) {
	d := chirpDir(t)
	m := genByID["bf-dm32uv"]
	p := filepath.Join(d, "tests", "images", "Baofeng_DM-32UV.img")
	img := loadImg(t, p)
	f := dm32Fam{}
	type ed struct {
		n int
		c *Chan
	}
	edits := []ed{
		{10, nil}, {11, nil}, {12, nil}, // tutta la zona «Digital» e la lista «DMR scan»
		{14, nil}, // l'ultimo canale: il numero dei canali scende
		{5, nil},  // in mezzo, in una zona e in una lista
		{20, &Chan{Rx: 433500000, Tx: 433500000, Power: 0, Name: "NUOVO"}},
		{2, &Chan{Rx: 438500000, Tx: 430900000, Power: 2, Narrow: true, Name: "DMR RPT", Dig: true}},
		{3, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "D023N", RxTone: "D023N", Power: 1, Name: "Rpt TSQL"}},
	}
	out := append([]byte{}, img...)
	var js []map[string]interface{}
	for _, e := range edits {
		if err := f.Set(m, out, e.n-1, e.c); err != nil {
			t.Fatal(err)
		}
		var cj interface{}
		if e.c != nil {
			mode := "FM"
			if e.c.Dig {
				mode = "DMR"
			} else if e.c.Narrow {
				mode = "NFM"
			}
			cj = map[string]interface{}{"rx": e.c.Rx, "tx": e.c.Tx, "txTone": e.c.TxTone, "rxTone": e.c.RxTone,
				"power": m.powerIndex(e.c.Power), "mode": mode, "scan": true, "name": e.c.Name}
		}
		js = append(js, map[string]interface{}{"n": e.n, "chan": cj})
	}
	tmp := t.TempDir()
	ef := filepath.Join(tmp, "edits.json")
	of := filepath.Join(tmp, "out.img")
	b, _ := json.Marshal(js)
	_ = os.WriteFile(ef, b, 0o644)
	chirpRef(t, "encode", m.RClass, p, ef, of)
	ref, err := os.ReadFile(of)
	if err != nil {
		t.Fatal(err)
	}
	ref, _ = splitImg(ref)
	if !bytes.Equal(ref, out) {
		t.Fatalf("immagine diversa da quella di CHIRP: %s", diffRanges(ref, out))
	}
	if f.zoneCount(out) != 1 || f.count(out) != 20 {
		t.Fatalf("zone %d, canali %d", f.zoneCount(out), f.count(out))
	}
	t.Logf("%d modifiche (zone, liste, canali DMR): immagine identica a CHIRP", len(edits))
}
