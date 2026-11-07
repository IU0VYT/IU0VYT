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
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakeBTU: file .BTU di prova (un'area; i byte 5-15 dell'intestazione sono
// casuali come nei file di Radtel), programma da org
func fakeBTU(size int, org uint32) []byte {
	r := rand.New(rand.NewSource(int64(size) + int64(org)))
	b := make([]byte, 16+size)
	b[0] = 1
	b[1], b[2], b[3], b[4] = byte(size>>24), byte(size>>16), byte(size>>8), byte(size)
	for i := 5; i < 16; i++ {
		b[i] = byte(r.Intn(256))
	}
	r.Read(b[16:])
	v := []uint32{0x20004a58, org + 0xd5}
	for i := 2; i < 16; i++ {
		v = append(v, org+0x101+uint32(i)*4)
	}
	for i, x := range v {
		b[16+i*4], b[17+i*4], b[18+i*4], b[19+i*4] = byte(x), byte(x>>8), byte(x>>16), byte(x>>24)
	}
	return b
}

func TestAnalyzeBTU(t *testing.T) {
	for _, c := range []struct {
		name, confirm string
		models        []string
		org           uint32
	}{
		{"RT910_H1.00G_V0.11_260305.BTU", "RT910", []string{"bj-7700nbt"}, 0x08003000},
		{"RT900_H3.01_OPEN_V1.14P_251103.BTU", "RT900", []string{"rt-rt900"}, 0x08003000},
		{"WT_RT_900_512_V1.04P_20241212.BTU", "RT900", []string{"rt-rt900"}, 0x08001000},
		{"BJ7700_5W_V1.07_251103.BTU", "BJ7700", []string{"bj-7700nbt"}, 0x08003000},
	} {
		raw := fakeBTU(100000, c.org)
		if k := sniffKind(c.name, raw); k != fwBTU {
			t.Fatalf("%s: tipo %v", c.name, k)
		}
		if k := sniffKind("senza_estensione", raw); k != fwBTU {
			t.Fatalf("%s senza estensione: tipo %v", c.name, k)
		}
		img := analyzeBTU(c.name, "", raw)
		in := img.Info
		if !in.Valid || !in.Typed || !in.Warnings || in.Confirm != c.confirm || !slices.Equal(in.Models, c.models) || in.Platform != "BT80" {
			t.Fatalf("%s: %+v", c.name, in)
		}
		old := strings.Contains(in.Hw, "0x08001000")
		if old != (c.org == 0x08001000) {
			t.Errorf("%s: elettronica %q", c.name, in.Hw)
		}
	}
	if v := analyzeBTU("RT910_V0.11_260305.BTU", "", fakeBTU(100000, 0x08003000)).Info.Version; v != "V0.11" {
		t.Errorf("versione %q", v)
	}
	// errori: lunghezza sbagliata, vettori non validi, nome senza modello
	b := fakeBTU(50000, 0x08003000)
	b[4]++
	if analyzeBTU("RT910_V0.11.BTU", "", b).Info.Valid {
		t.Error("lunghezza errata accettata")
	}
	b = fakeBTU(50000, 0x08003000)
	b[16+4] = 0x00 // reset pari
	if analyzeBTU("RT910_V0.11.BTU", "", b).Info.Valid {
		t.Error("vettori non validi accettati")
	}
	if in := analyzeBTU("firmware.BTU", "", fakeBTU(50000, 0x08003000)).Info; !in.Valid || len(in.Models) != 0 || !in.Typed {
		t.Errorf("nome senza modello: %+v", in)
	}
	// un .BF (due aree) non è un .BTU
	if sniffKind("x", fakeBF(20000, 1000)) == fwBTU {
		t.Error(".BF scambiato per .BTU")
	}
}

// presentazione «PROGRAMBT80U» a 57600 baud, porta riaperta a 115200 per il
// caricatore; modalità manuale a 115200; ultimo pacchetto completato con FF
func TestBtuEmu(t *testing.T) {
	for _, s := range []struct {
		name   string
		raw    []byte
		manual bool
		opts   kdhOpts
		setup  func(*kdhEmu)
		ok     bool
	}{
		{"automatico", fakeBTU(107860, 0x08003000), false, btuOpts, nil, true},
		{"manuale", fakeBTU(107860, 0x08003000), true, btuOpts, nil, true},
		{"multiplo di 1024", fakeBTU(40960, 0x08003000), false, btuOpts, nil, true},
		{"risposta persa", fakeBTU(20000, 0x08003000), false, btuOpts, func(e *kdhEmu) { e.drop[5] = true }, true},
		{"presentazione persa", fakeBTU(20000, 0x08003000), false, btuOpts, func(e *kdhEmu) { e.hsDrop = 2 }, true},
		// presentazione della UV-5RM: la radio non risponde, nessun dato inviato
		{"presentazione sbagliata", fakeBTU(20000, 0x08003000), false, kdhOpts{}, nil, false},
		{"radio spenta", fakeBTU(20000, 0x08003000), false, btuOpts, func(e *kdhEmu) { e.silent = true }, false},
	} {
		t.Run(s.name, func(t *testing.T) {
			e := newKdhEmu(s.raw, s.manual)
			e.hello, e.helloBaud, e.padFF = "PROGRAMBT80U", 57600, true
			if s.setup != nil {
				s.setup(e)
			}
			old := openPortFunc
			openPortFunc = kdhEmuPorts(e)
			defer func() { openPortFunc = old }()
			o := s.opts
			o.Manual = s.manual
			if s.name == "presentazione sbagliata" {
				o.HelloBaud = 57600
			}
			sent, err := kdhUpdate(context.Background(), "COM9", s.raw, o, noProg, func(m string) { t.Log(m) }, nil)
			time.Sleep(50 * time.Millisecond)
			_, errs, state := e.snapshot()
			if len(errs) > 0 {
				t.Fatalf("segnalazioni della radio: %v", errs)
			}
			if (err == nil) != s.ok || (s.ok && (state != "done" || sent != int64(len(s.raw)-16))) || (!s.ok && sent != 0) {
				t.Fatalf("esito inatteso: %v, stato %s, %d byte", err, state, sent)
			}
		})
	}
}

// firmware veri (BTU_FW = cartella con i file .BTU di Radtel e Bajeton):
// riconosciuti, con la radio giusta e l'impronta conosciuta, e inviati per
// intero alla radio simulata
func TestBTURealFirmware(t *testing.T) {
	dir := os.Getenv("BTU_FW")
	if dir == "" {
		t.Skip("BTU_FW non impostata")
	}
	fs, _ := filepath.Glob(filepath.Join(dir, "*.BTU"))
	if len(fs) < len(fwKnownBTU) {
		t.Fatalf("solo %d file", len(fs))
	}
	for _, f := range fs {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(f)
		in := analyzeBTU(name, f, raw).Info
		up := strings.ToUpper(name)
		want, model := "RT900", "rt-rt900"
		switch {
		case strings.HasPrefix(up, "RT910"):
			want, model = "RT910", "bj-7700nbt"
		case strings.HasPrefix(up, "BJ7700"):
			want, model = "BJ7700", "bj-7700nbt"
		}
		if !in.Valid || in.Known == "" || in.Confirm != want || !slices.Equal(in.Models, []string{model}) || in.Version == "" {
			t.Errorf("%s: %+v", name, in)
			continue
		}
		if strings.Contains(up, "V1.04P") != strings.Contains(in.Hw, "0x08001000") {
			t.Errorf("%s: elettronica %q", name, in.Hw)
		}
		e := newKdhEmu(raw, false)
		e.hello, e.helloBaud, e.padFF = "PROGRAMBT80U", 57600, true
		old := openPortFunc
		openPortFunc = kdhEmuPorts(e)
		sent, err := kdhUpdate(context.Background(), "COM9", raw, btuOpts, noProg, nil, nil)
		openPortFunc = old
		time.Sleep(30 * time.Millisecond)
		_, errs, state := e.snapshot()
		if err != nil || state != "done" || len(errs) > 0 || sent != int64(len(raw)-16) {
			t.Errorf("%s: %v, stato %s, %v, %d byte", name, err, state, errs, sent)
		}
		t.Logf("%s: %s, %s, %s", name, in.Version, in.Date, in.Hw)
	}
}
