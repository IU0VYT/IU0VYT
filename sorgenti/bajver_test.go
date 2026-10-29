//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
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

// file .dat delle due versioni di BT-11_CPS: lettura, conversione e impostazioni
func TestDatVersions(t *testing.T) {
	if datVer(baseDAT) != "1.0.8" || datVer(baseDAT109) != "1.0.9" {
		t.Fatalf("versioni dei modelli: %q %q", datVer(baseDAT), datVer(baseDAT109))
	}
	src := mustRead(t, sampleDAT)
	img, err := datToEEPROM(src)
	if err != nil {
		t.Fatal(err)
	}
	_, chs, err := readRadio(src, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	v9, err := datToVer(src, "1.0.9")
	if err != nil || datVer(v9) != "1.0.9" {
		t.Fatalf("conversione: %v", err)
	}
	back, err := datToVer(v9, "1.0.8")
	if err != nil || datVer(back) != "1.0.8" {
		t.Fatalf("conversione inversa: %v", err)
	}
	for _, f := range [][]byte{v9, back} {
		k, got, err := readRadio(f, &Report{})
		if err != nil || k != kindDAT {
			t.Fatal(err)
		}
		if err := verifySaved(chs, got); err != nil {
			t.Fatal(err)
		}
		im, err := datToEEPROM(f)
		if err != nil || !bytes.Equal(im, img) {
			t.Fatalf("memoria diversa dopo la conversione (%v)", err)
		}
		// salvataggio dei canali nel formato del file
		out, err := writeRadio(kindDAT, f, got, &Report{})
		if err != nil || datVer(out) != datVer(f) {
			t.Fatalf("salvataggio: %v", err)
		}
	}
	if fa, fb := datFields(t, src), datFields(t, back); len(fa) != len(fb) {
		t.Fatalf("campi: %d / %d", len(fa), len(fb))
	} else {
		for k, v := range fa {
			if fb[k] != v {
				t.Errorf("campo %s: %q -> %q", k, v, fb[k])
			}
		}
	}
	n8, n9 := len(settingsFor(kindDAT, src).Settings), len(settingsFor(kindDAT, v9).Settings)
	if n9 != n8+1 {
		t.Fatalf("impostazioni: %d (1.0.8) e %d (1.0.9)", n8, n9)
	}
}

// 3.9.2: canali in zone o in un unico elenco come il programma BT-11_CPS v1.0.9
// (cbB_DivideCH: 0 zone, 1 tutti i canali); i file della v1.0.8 sono sempre a zone
func TestDatSingleList(t *testing.T) {
	if datSingleList(baseDAT) || !datSingleList(baseDAT109) {
		t.Fatalf("modelli: v1.0.8 %v, v1.0.9 %v", datSingleList(baseDAT), datSingleList(baseDAT109))
	}
	src := mustRead(t, sampleDAT) // v1.0.8
	v9, err := datToVer(src, "1.0.9")
	if err != nil || datSingleList(src) || datSingleList(v9) {
		t.Fatalf("conversione di un file a zone: %v %v %v", err, datSingleList(src), datSingleList(v9))
	}
	c := newAPIClient(t)
	c.must("radio/datver", map[string]string{"ver": "1.0.9"})
	r := c.must("radio/new", map[string]string{"radio": "baj"})
	if r["datVer"] != "1.0.9" || r["single"] != true {
		t.Fatalf("nuovo file v1.0.9: %v %v", r["datVer"], r["single"])
	}
	s := c.must("radio/settings/set", map[string]interface{}{"kind": "baj", "values": map[string]string{"divide_ch": "0"}})
	if s["single"] != false {
		t.Fatalf("divisione in zone: %v", s["single"])
	}
	s = c.must("radio/settings/set", map[string]interface{}{"kind": "baj", "values": map[string]string{"divide_ch": "1"}})
	if s["single"] != true {
		t.Fatalf("tutti i canali: %v", s["single"])
	}
	// ritorno alla v1.0.8: sempre a zone
	if d := c.must("radio/datver", map[string]string{"ver": "1.0.8"}); d["single"] != false {
		t.Fatalf("v1.0.8: %v", d)
	}
	c.must("radio/datver", map[string]string{"ver": "1.0.8"})
	if r := c.must("radio/new", map[string]string{"radio": "baj"}); r["single"] != nil {
		t.Fatalf("nuovo file v1.0.8: %v", r["single"])
	}
}
