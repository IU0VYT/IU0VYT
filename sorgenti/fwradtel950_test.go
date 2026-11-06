//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// 4.7: firmware di Radtel RT-950 Pro / RT-950 con il caricatore della BJ-9000, e
// firmware .anc di Radtel RT-920 / RT-910B (stesse radio di BJ-7800 / BJ-7700)
func TestAnalyzeRadtel950(t *testing.T) {
	pro := analyzeBJ9000("RT_950Pro_V0.29_260617.BTF", "", fakeBJ9000(300000))
	if !pro.Info.Valid || pro.Info.Confirm != "BJ9000" || !slices.Contains(pro.Info.Models, "bj-9000") || pro.Info.Version != "V0.29" {
		t.Fatalf("RT-950 Pro: %+v", pro.Info)
	}
	b := fakeBJ9000(300000)
	copy(b[992:], "BT-9000     ")
	np := analyzeBJ9000("RT_950_V0.29_251104.BTF", "", b)
	if !np.Info.Valid || np.Info.Confirm != "RT950" || !slices.Equal(np.Info.Models, []string{"rtl-rt950"}) {
		t.Fatalf("RT-950: %+v", np.Info)
	}
	// firmware della RT-950 senza «Pro» scelto per la BJ-9000: conferma scritta
	if !fwOtherModel(&fwStartReq{Model: "bj-9000"}, np) || fwOtherModel(&fwStartReq{Model: "rtl-rt950"}, np) {
		t.Fatal("radio scelta e firmware non confrontati")
	}
	copy(b[992:], "XX-123      ")
	if o := analyzeBJ9000("strano.BTF", "", b); !o.Info.Valid || !o.Info.Typed || !o.Info.Warnings {
		t.Fatalf("modello sconosciuto: %+v", o.Info)
	}
	for _, c := range []struct{ name, confirm, family string }{
		{"RT920_V0.18_260310.anc", "RT920", "BJ7800"},
		{"RT_910B_V0.14_250808.anc", "RT910", "BJ7700"},
		{"BJ7800_CN_EN_V0.17_20251126.anc", "BJ7800", "BJ7800"},
		{"BJ8300_V0.12_20251112.anc", "BJ8300", "BJ8300"},
	} {
		img := analyzeANC(c.name, "", ancDecrypt(fakeFirmware(300000, bj8300Platform)))
		if !img.Info.Valid || img.Info.Confirm != c.confirm || ancFamily(img) != c.family {
			t.Errorf("%s: %+v (famiglia %s)", c.name, img.Info, ancFamily(img))
		}
	}
}

// firmware veri (RT950_FW = cartella con i file pubblicati da Radtel e Bajeton):
// tutti riconosciuti, con la radio giusta e l'impronta conosciuta
func TestRadtel950RealFirmware(t *testing.T) {
	dir := os.Getenv("RT950_FW")
	if dir == "" {
		t.Skip("RT950_FW non impostata")
	}
	fs, _ := filepath.Glob(filepath.Join(dir, "*"))
	n := 0
	for _, f := range fs {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(f)
		var img *fwImage
		switch sniffKind(name, b) {
		case fwBJ9000:
			img = analyzeBJ9000(name, f, b)
		case fwANC:
			img = analyzeANC(name, f, b)
		default:
			continue
		}
		in := img.Info
		up := strings.ToUpper(name)
		want := map[bool]string{true: "RT950", false: "BJ9000"}[strings.HasPrefix(up, "RT_950_")]
		switch {
		case strings.HasPrefix(up, "RT920"):
			want = "RT920"
		case strings.HasPrefix(up, "RT_910B"):
			want = "RT910"
		}
		if !in.Valid || in.Known == "" || in.Confirm != want {
			t.Errorf("%s: valido %v, conosciuto %q, conferma %s (attesa %s)", name, in.Valid, in.Known, in.Confirm, want)
		}
		n++
	}
	if n < 10 {
		t.Fatalf("solo %d firmware esaminati", n)
	}
	t.Logf("%d firmware riconosciuti", n)
}

// firmware vero della RT-950 Pro inviato alla radio simulata con il caricatore
// della BJ-9000; quello della RT-950 senza «Pro» rifiutato prima dei dati (E6)
func TestRadtel950RealUpdate(t *testing.T) {
	dir := os.Getenv("RT950_FW")
	if dir == "" {
		t.Skip("RT950_FW non impostata")
	}
	pro, err := os.ReadFile(filepath.Join(dir, "RT_950Pro_V0.29_260617.BTF"))
	if err != nil {
		t.Skip(err)
	}
	np, err := os.ReadFile(filepath.Join(dir, "RT_950_V0.29_251104.BTF"))
	if err != nil {
		t.Skip(err)
	}
	run := func(radio, file []byte) (int64, error, string) {
		e := newKdh2Emu(bj9000Proto, radio, true)
		old := openPortFunc
		openPortFunc = kdh2EmuPorts(e)
		defer func() { openPortFunc = old }()
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		sent, err := kdh2Update(ctx, "COM9", file, bj9000Proto, true, noProg, nil, nil)
		time.Sleep(50 * time.Millisecond)
		_, _, state := e.snapshot()
		return sent, err, state
	}
	if sent, err, state := run(pro, pro); err != nil || state != "done" || sent != int64(len(pro)) {
		t.Fatalf("RT-950 Pro: %v, stato %s, %d byte", err, state, sent)
	}
	if sent, err, _ := run(pro, np); err == nil || sent != 0 {
		t.Fatalf("firmware della RT-950 accettato da una RT-950 Pro: %v, %d byte", err, sent)
	}
}
