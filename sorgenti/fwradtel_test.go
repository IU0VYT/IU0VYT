//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.3: firmware .kdhx di Radtel (RT-470, RT-470L, RT-470X, RT-495, RT-630) con
// il caricatore «JJCC Bootloader Pro».

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRadtelKdhxAnalyze(t *testing.T) {
	raw := fakeKDHX(64072)
	cases := []struct {
		name, target string
		models       []string
		bad          bool
	}{
		{"470 serial MFM_V2.13C_20240425.kdhx", "Radtel RT-470 / RT-470L / RT-470X", []string{"rt-rt470", "rt-rt470l", "rt-rt470x"}, false},
		{"RT470_AllBandTx_V1.27A_20240523.kdhx", "Radtel RT-470 / RT-470L / RT-470X", []string{"rt-rt470"}, false},
		{"RT470_senza_versione.kdhx", "Radtel RT-470 / RT-470L / RT-470X", nil, true},
		{"495_NUM_ALLBAND_V0.07_20231207.kdhx", "Radtel RT-495", []string{"rt-rt495"}, false},
		{"rt-630 8.33AM137_AMFM.kdhx", "Radtel RT-630", []string{"rt-rt630"}, false},
	}
	for _, c := range cases {
		img := analyzeKDHX(c.name, c.name, raw)
		in := img.Info
		t.Logf("%s → %s, %s, valido %v, %s", c.name, in.Target, in.Platform, in.Valid, in.Hw)
		if in.Target != c.target || in.Platform != "JJCC" || !in.Typed {
			t.Errorf("%s: destinazione %q piattaforma %q", c.name, in.Target, in.Platform)
		}
		if in.Valid == c.bad {
			t.Errorf("%s: valido %v", c.name, in.Valid)
		}
		for _, m := range c.models {
			if !slices.Contains(in.Models, m) {
				t.Errorf("%s: manca il modello %s", c.name, m)
			}
		}
	}
	// i file della UV-25 Plus e quelli senza modello restano alla famiglia Baofeng
	for _, n := range []string{"UV25Pro_NRF_401+_V0.23_250522.kdhx", "firmware_20240630.kdhx", "H802_RT470.kdhx"} {
		if in := analyzeKDHX(n, n, raw).Info; in.Platform == "JJCC" {
			t.Errorf("%s: riconosciuto come Radtel", n)
		}
	}
}

func TestRadtelKdhxEmu(t *testing.T) {
	raw := fakeKDHX(64072)
	run := func(radio *kdh2Variant, manual bool) error {
		e := newKdh2Emu(radio, raw, manual)
		old := openPortFunc
		openPortFunc = kdh2EmuPorts(e)
		defer func() { openPortFunc = old }()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		sent, err := kdh2Update(ctx, "COM9", raw, kdhxJJProto, manual, noProg, nil, nil)
		time.Sleep(50 * time.Millisecond)
		_, errs, state := e.snapshot()
		if len(errs) > 0 {
			t.Fatalf("segnalazioni della radio: %v", errs)
		}
		if err == nil && (state != "done" || sent != int64(len(raw))) {
			t.Fatalf("stato %s, %d byte", state, sent)
		}
		return err
	}
	if err := run(kdhxJJProto, false); err != nil {
		t.Fatalf("automatico: %v", err)
	}
	if err := run(kdhxJJProto, true); err != nil {
		t.Fatalf("manuale: %v", err)
	}
	// una radio Baofeng (H802) non risponde alla presentazione JJCC
	if err := run(kdhxProto, false); err == nil {
		t.Fatal("radio H802 aggiornata con la presentazione JJCC")
	}
	if !bytes.Equal(kdhxJJProto.Hello, []byte("PROGROMJJCCU")) || kdhxJJProto.MaxPkt != kdhxProto.MaxPkt {
		t.Fatal("variante JJCC diversa dal programma di Radtel")
	}
}

// TestRadtelKdhxReal: pacchetti scaricati da radtels.com (RADTEL_FW=cartella)
func TestRadtelKdhxReal(t *testing.T) {
	d := os.Getenv("RADTEL_FW")
	if d == "" {
		t.Skip("RADTEL_FW non impostato")
	}
	fs, _ := filepath.Glob(filepath.Join(d, "*.rar"))
	n := 0
	for _, f := range fs {
		b, _ := os.ReadFile(f)
		e := newExpander(filepath.Base(f), f, b)
		if err := e.run(); err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		for _, c := range e.found {
			if !strings.HasSuffix(strings.ToLower(c.Name), ".kdhx") {
				continue
			}
			in := analyzeKDHX(c.Name, c.Where, c.data).Info
			t.Logf("%s: %s, %s, %s, conosciuto %q", c.Name, in.Target, in.Version, in.Hw, in.Known)
			if !in.Valid || in.Platform != "JJCC" || in.Known == "" {
				t.Errorf("%s: valido %v piattaforma %s", c.Name, in.Valid, in.Platform)
			}
			for _, ch := range in.Checks {
				if ch.Level == "bad" || strings.Contains(ch.Text, "danneggiato") {
					t.Errorf("%s: %s", c.Name, ch.Text)
				}
			}
			// rinominato: il contenuto vale più del nome
			if r := analyzeKDHX("firmware.kdhx", "x", c.data).Info; r.Target != in.Target {
				t.Errorf("%s rinominato: %s", c.Name, r.Target)
			}
			n++
		}
	}
	if n < 4 {
		t.Errorf("trovati %d firmware Radtel", n)
	}
}
