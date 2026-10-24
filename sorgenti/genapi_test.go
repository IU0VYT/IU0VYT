//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove dell'API per le radio generiche (driver di CHIRP) e del firmware
// Baofeng: catalogo, apertura e salvataggio .img, adattamento, lettura e
// scrittura via cavo con radio simulate, aggiornamento .BF e .dat.

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func chansFrom(t *testing.T, v interface{}) []*Chan {
	b, _ := json.Marshal(v)
	var out []*Chan
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestGenAPI(t *testing.T) {
	m := genByID["bf-uv5rm"]
	dir := t.TempDir()
	// immagine di partenza: quella di CHIRP se disponibile, altrimenti una sintetica
	var mem []byte
	if d := os.Getenv("CHIRP_DIR"); d != "" {
		b, err := os.ReadFile(filepath.Join(d, "tests/images/Baofeng_5RM.img"))
		if err == nil {
			mem, _ = splitImg(b)
		}
	}
	if mem == nil {
		mem = bytes.Repeat([]byte{0xFF}, uv17o(m).total)
		if err := m.Fam.Set(m, mem, 0, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"}); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(dir, "radio.img")
	os.WriteFile(src, makeImg(m, mem, nil), 0o644)

	c := newAPIClient(t)
	cat := c.must("catalog", nil)
	if len(cat["catalog"].([]interface{})) < 50 {
		t.Fatal("catalogo incompleto")
	}
	if _, err := c.call("radio/new", map[string]string{"radio": m.ID}); err == nil {
		t.Fatal("radio/new dovrebbe rifiutare le radio generiche")
	}
	r := c.must("radio/open", map[string]string{"path": src})
	if r["kind"] != m.ID {
		t.Fatalf("radio riconosciuta come %v", r["kind"])
	}
	chs := chansFrom(t, r["channels"])
	if len(chs) != m.Slots {
		t.Fatalf("%d canali", len(chs))
	}
	// salvataggio senza modifiche: file identico (a parte i metadati, conservati)
	out := filepath.Join(dir, "copia.img")
	c.must("radio/save", map[string]interface{}{"radio": m.ID, "channels": chs, "path": out})
	b1, _ := os.ReadFile(out)
	m1, _ := splitImg(b1)
	if !bytes.Equal(m1, mem) {
		t.Fatal("salvataggio senza modifiche: memoria cambiata")
	}
	// adattamento di canali da un'altra radio
	ad := c.must("adapt", map[string]interface{}{"to": m.ID, "channels": []*Chan{
		{Rx: 430500000, Tx: 431100000, TxTone: "D754I", RxTone: "67.0", Power: 1, Narrow: true, Scan: true, Name: "Ponte con un nome lungo", Raw: "xlc:00"},
		{Rx: 145500000, Tx: 145500000, TxOff: true, Name: "solo rx"},
	}, "pos": []int{3, 4}})
	ac := chansFrom(t, ad["channels"])
	if ac[0] == nil || len(ac[0].Name) > m.NameMax || ac[0].Raw != "" || ac[0].Power != 1 {
		t.Fatalf("adattamento: %+v", ac[0])
	}
	chs[3], chs[4] = ac[0], ac[1]
	c.must("radio/save", map[string]interface{}{"radio": m.ID, "channels": chs, "path": out})
	b2, _ := os.ReadFile(out)
	_, meta := splitImg(b2)
	if meta["rclass"] != m.RClass {
		t.Fatalf("metadati: %v", meta)
	}
	r2 := c.must("radio/open", map[string]string{"path": out, "radio": m.ID})
	back := chansFrom(t, r2["channels"])
	if back[3] == nil || back[3].Rx != 430500000 || back[3].TxTone != "D754I" || back[4] == nil || !back[4].TxOff {
		t.Fatalf("riletti: %+v %+v", back[3], back[4])
	}

	// via cavo con la radio simulata
	emu := genEmuFor(m, mem)
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go emu.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": m.ID, "port": "COM7"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("lettura: %v", st["error"])
	}
	rd := chansFrom(t, st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"])
	rd[10] = &Chan{Rx: 433500000, Tx: 433500000, Power: 0, Scan: true, Name: "PROVA"}
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": m.ID, "port": "COM7", "channels": rd, "backup": true, "verify": true})
	st, _ = c.wait("")
	if st["state"] != "done" {
		t.Fatalf("scrittura: %v", st["error"])
	}
	res := st["result"].(map[string]interface{})
	if res["verified"] != true || res["backup"] == nil {
		t.Fatalf("scrittura: %v / note %v", res, st["notes"])
	}
	if g := m.Fam.Get(m, emu.image(), 10); g == nil || g.Name != "PROVA" {
		t.Fatalf("radio simulata: canale 11 %+v", g)
	}
}

func TestFwAPIBaofeng(t *testing.T) {
	dir := t.TempDir()
	c := newAPIClient(t)
	// .BF dentro uno zip
	bf := fakeBF(30000, 1200)
	zp := filepath.Join(dir, "fw.zip")
	writeZip(t, zp, map[string][]byte{"5RH_BK_V0.20_250101.BF": bf, "leggimi.txt": []byte("x")})
	r := c.must("fw/load", map[string]string{"source": "file", "path": zp})
	fw := r["firmware"].(map[string]interface{})
	if r["state"] != "ready" || fw["kind"] != "bf" || fw["valid"] != true {
		t.Fatalf("fw/load: %v", r)
	}
	emu := newKdhEmu(bf, false)
	old := openPortFunc
	openPortFunc = kdhEmuPorts(emu)
	defer func() { openPortFunc = old }()
	req := map[string]interface{}{"port": "COM8", "mode": "auto", "sha256": fw["sha256"], "confirmModel": true, "acceptWarnings": true}
	c.must("fw/start", req)
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("aggiornamento .BF: %v", st["error"])
	}
	if _, errs, s := emu.snapshot(); len(errs) > 0 || s != "done" {
		t.Fatalf("radio simulata: %v %s", errs, s)
	}
	// 5RH Pro
	dat := fake5RH(20000, true)
	dp := filepath.Join(dir, "BF_5RH_501_v2_0_99.dat")
	os.WriteFile(dp, dat, 0o644)
	r = c.must("fw/load", map[string]string{"source": "file", "path": dp})
	fw = r["firmware"].(map[string]interface{})
	if fw["kind"] != "rh5" || fw["valid"] != true {
		t.Fatalf("fw/load .dat: %v", r)
	}
	e5 := newRh5Emu(dat, true)
	openPortFunc = rh5EmuPorts(e5)
	if _, err := c.call("fw/start", map[string]interface{}{"port": "COM9", "mode": "update", "baud": 115200, "sha256": fw["sha256"]}); err == nil {
		t.Fatal("serve la conferma del modello")
	}
	c.must("fw/start", map[string]interface{}{"port": "COM9", "mode": "update", "baud": 115200, "sha256": fw["sha256"], "confirmModel": true, "acceptWarnings": true})
	st, _ = c.wait("")
	if st["state"] != "done" {
		t.Fatalf("aggiornamento .dat: %v", st["error"])
	}
	if _, errs, s := e5.snapshot(); len(errs) > 0 || s != "aggiornata" {
		t.Fatalf("radio simulata: %v %s", errs, s)
	}
}

func writeZip(t *testing.T, p string, files map[string][]byte) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for n, d := range files {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		w.Write(d)
	}
	zw.Close()
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
