//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio a cui è destinato un firmware .BF o .kdhx (dal nome del file) e
// conferme richieste prima dell'invio.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func sha256hex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func TestFwTargets(t *testing.T) {
	kd := fakeKDHX(64 * 1024)
	bf := fakeBF(59720, 1540)
	cases := []struct {
		img     *fwImage
		confirm string
		typed   bool
		driver  string
	}{
		{analyzeKDHX("UV25Pro_NRF_401+_V0.23_250522.kdhx", "", kd), "UV25PLUS", false, "bf-uv25plus"},
		{analyzeKDHX("UV29Pro_V0.12_260101.kdhx", "", kd), "UV29PRO", true, "bf-uv29pro"},
		{analyzeKDHX("SW-9_V0.10.kdhx", "", kd), "UV29PRO", true, "bf-uv29pro"},
		{analyzeKDHX("AR152Pro_V0.31.kdhx", "", kd), "AR152PRO", true, "bf-ar152pro"},
		{analyzeKDHX("UV-7B_V0.05.kdhx", "", kd), "UV7B", true, "bf-uv7b"},
		{analyzeKDHX("UV32_V0.08.kdhx", "", kd), "UV32", true, "bf-uv32"},
		{analyzeKDHX("UV28Pro_V0.09.kdhx", "", kd), "UV28PRO", true, "bf-uv28pro"},
		{analyzeKDHX("firmware.kdhx", "", kd), "H802", true, ""},
		{analyzeBF("5RH_BK_V0.14_231123.BF", "", bf), "5RM", false, "bf-uv5rm"},
		{analyzeBF("x.BF", "", bf), "5RM", false, "bf-uv5rh"},
		{analyzeBF("K5Plus_V0.10_240101.BF", "", bf), "K5PLUS", true, "bf-k5plus"},
		{analyzeBF("5RM_Plus_V0.10.BF", "", bf), "5RMPLUS", true, "bf-uv5rmplus"},
		{analyzeBF("UV17R_V0.10.BF", "", bf), "UV17R", true, "bf-uv17r"},
		{analyzeBF("UV-21L_V0.10.BF", "", bf), "UV21L", true, "bf-uv21l"},
	}
	// file conosciuti rinominati: conta il contenuto, non il nome
	if real := os.Getenv("KDH2_REAL"); real != "" {
		if b, err := os.ReadFile(filepath.Join(real, "UV25Pro_NRF_401+_V0.23_250522.kdhx")); err == nil {
			cases = append(cases, struct {
				img     *fwImage
				confirm string
				typed   bool
				driver  string
			}{analyzeKDHX("UV29Pro_V0.12.kdhx", "", b), "UV25PLUS", false, "bf-uv25plus"})
		}
	}
	// .BF conosciuto (qui uno di prova aggiunto all'elenco) con il nome di un'altra radio
	kb := fakeBF(59720, 1544)
	fwKnownBaofeng[sha256hex(kb)] = "prova"
	defer delete(fwKnownBaofeng, sha256hex(kb))
	cases = append(cases, struct {
		img     *fwImage
		confirm string
		typed   bool
		driver  string
	}{analyzeBF("K5Plus_V0.10.BF", "", kb), "5RM", false, "bf-uv5rm"})
	for _, c := range cases {
		i := c.img.Info
		if !i.Valid || i.Confirm != c.confirm || i.Typed != c.typed || c.driver != "" && !slices.Contains(i.Models, c.driver) || c.driver == "" && len(i.Models) != 0 {
			t.Errorf("%s: valido %v, conferma %q, scritta %v, radio %v", i.Name, i.Valid, i.Confirm, i.Typed, i.Models)
			continue
		}
		if c.driver != "" {
			if _, ok := genByID[c.driver]; !ok {
				t.Errorf("%s: driver %s inesistente", i.Name, c.driver)
			}
		}
		// avvio: conferma del modello e, se richiesta, conferma scritta (anche in automatico)
		req := &fwStartReq{Port: "COM9", Mode: "auto", SHA256: i.SHA256, ConfirmModel: true, AcceptWarnings: true}
		err := fwCheckStart(req, c.img)
		if c.typed != (err != nil) {
			t.Errorf("%s: avvio senza conferma scritta: %v", i.Name, err)
		}
		req.ConfirmText = c.confirm
		if err := fwCheckStart(req, c.img); err != nil {
			t.Errorf("%s: avvio con %s: %v", i.Name, c.confirm, err)
		}
		req.ConfirmText = "UV5R"
		if c.typed && fwCheckStart(req, c.img) == nil {
			t.Errorf("%s: accettata una conferma sbagliata", i.Name)
		}
	}
	// radio scelta nella finestra diversa da quella del file: conferma scritta
	u25 := analyzeKDHX("UV25Pro_NRF_401+_V0.23_250522.kdhx", "", kd)
	req := &fwStartReq{Port: "COM9", Mode: "auto", SHA256: u25.Info.SHA256, ConfirmModel: true, AcceptWarnings: true, Model: "bf-uv29pro"}
	if fwCheckStart(req, u25) == nil {
		t.Error("UV-25 Plus su UV-29 Pro accettato senza conferma scritta")
	}
	req.ConfirmText = "UV25PLUS"
	if err := fwCheckStart(req, u25); err != nil {
		t.Error(err)
	}
	req.Model, req.ConfirmText = "bf-uv25plus", ""
	if err := fwCheckStart(req, u25); err != nil {
		t.Error(err)
	}
	// ogni radio indicata dai nomi dei file ha nel catalogo il caricatore giusto
	for _, tb := range []struct {
		kind string
		tbl  []fwTarget
	}{{"bf", bfTargets}, {"kdhx", kdhxTargets}} {
		for _, x := range tb.tbl {
			for _, d := range x.drivers {
				found := false
				for _, c := range catalog {
					if slices.Contains(c.Drivers, d) && (c.FW == "bf" || c.FW == "kdhx") {
						found = true
					}
				}
				if !found {
					t.Errorf("%s: nel catalogo nessuna radio con il driver %s e il firmware .BF/.kdhx", tb.kind, d)
				}
			}
		}
	}
}
