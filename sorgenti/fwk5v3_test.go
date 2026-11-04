//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.4: firmware delle radio con il caricatore 7.x (UV-K5 V3, UV-K1, Retevis RA79 V3)

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func k5v3Fake(n int) []byte {
	fw := fakeARM(n, int64(n))
	binary.LittleEndian.PutUint32(fw, 0x20002FA0)
	binary.LittleEndian.PutUint32(fw[4:], 0x080028D5)
	return fw
}

func TestK5v3Analyze(t *testing.T) {
	fw := k5v3Fake(71784)
	raw := k5Pack([]byte("7.00.16"), fw)
	if !k5IsPacked(raw) || sniffKind("RA79 V3.bin", raw) != fwK5 {
		t.Fatal("file 7.x non riconosciuto")
	}
	in := analyzeK5("RA79 V3.bin", "x", raw).Info
	t.Logf("%s, %s, %s", in.Target, in.Version, in.Hw)
	if !in.Valid || !strings.Contains(in.Target, "UV-K5 V3") || !slices.Contains(in.Models, "rt-ra79") {
		t.Fatalf("analisi: %+v", in)
	}
	// programma per un altro microcontrollore (avvio a 0x000000D5, come la UV-K5)
	bad := append([]byte{}, fw...)
	binary.LittleEndian.PutUint32(bad[4:], 0xD5)
	if analyzeK5("x.bin", "x", k5Pack([]byte("7.00.16"), bad)).Info.Valid {
		t.Fatal("firmware 7.x con l'avvio della UV-K5 accettato")
	}
	// la UV-K5 classica resta limitata a 0xEFFF byte
	if analyzeK5("k5.bin", "x", k5Pack([]byte("2.01.26"), fakeARM(0xF100, 1))).Info.Valid {
		t.Fatal("firmware 2.x troppo grande accettato")
	}
	if d := os.Getenv("RETEVIS_FW"); d != "" {
		b, err := os.ReadFile(d + "/RA79 V3.bin")
		if err != nil {
			t.Fatal(err)
		}
		in := analyzeK5("RA79 V3.bin", "x", b).Info
		t.Logf("file vero: %s %s %v", in.Version, in.Target, in.Checks)
		if !in.Valid || in.Version != "7.00.16" {
			t.Fatal("RA79 V3.bin non valido")
		}
	}
}

func TestK5v3Emu(t *testing.T) {
	oldW, oldB := k5BootWait, k5BlockWait
	k5BootWait, k5BlockWait = 3*time.Second, 400*time.Millisecond
	t.Cleanup(func() { k5BootWait, k5BlockWait = oldW, oldB })
	fw := k5v3Fake(71784)
	if d := os.Getenv("RETEVIS_FW"); d != "" {
		b, _ := os.ReadFile(d + "/RA79 V3.bin")
		if _, f, ok, err := k5Unpack(b); err == nil && ok {
			fw = f
			t.Log("con il firmware vero RA79 V3")
		}
	}
	run := func(radio interface{ serve(Port) error }, ver string, fw []byte) (int64, bool, error) {
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			x, y := newPipe()
			go radio.serve(y)
			return x, nil
		}
		defer func() { openPortFunc = old }()
		started := false
		v := make([]byte, 16)
		copy(v, ver)
		n, err := k5Update(context.Background(), "COM9", v, fw, noProg, func(s string) { t.Log(s) }, func() { started = true })
		time.Sleep(50 * time.Millisecond)
		return n, started, err
	}
	// caricatore 7.00.07 (UV-K5 V3 / RA79 V3), firmware 7.00.16: invio completo
	e := newK5v3BootEmu("7.00.07")
	e.nakAt = 5
	n, _, err := run(e, "7.00.16", fw)
	e.mu.Lock()
	if err != nil || n != int64(len(fw)) || !e.done || len(e.errs) > 0 || !bytes.Equal(e.flash[:len(fw)], fw) || e.hs != 3 {
		t.Fatalf("invio: %v, %d byte, %d blocchi, presentazioni %d, %v", err, n, e.pages, e.hs, e.errs)
	}
	e.mu.Unlock()
	// rifiuti, senza scrivere nulla
	for _, c := range []struct {
		name  string
		radio interface{ serve(Port) error }
		ver   string
		fw    []byte
	}{
		{"firmware 7.00 su caricatore UV-K1 7.03.01", newK5v3BootEmu("7.03.01"), "7.00.16", fw},
		{"firmware 7.x su UV-K5 V1 (caricatore 2.00.06)", newK5BootEmu("2.00.06"), "7.00.16", fw},
		{"firmware 7.x su UV-K5 V2 (caricatore 5.00.01)", newK5BootEmu("5.00.01"), "7.00.16", fw},
		{"firmware della UV-K5 (2.01.26) su caricatore 7.00.07", newK5v3BootEmu("7.00.07"), "2.01.26", fakeARM(0x8000, 3)},
	} {
		n, started, err := run(c.radio, c.ver, c.fw)
		t.Logf("%s: %v", c.name, err)
		if err == nil || started || n != 0 {
			t.Fatalf("%s: non rifiutato (%d byte)", c.name, n)
		}
		if v3, ok := c.radio.(*k5v3BootEmu); ok && v3.pages > 0 {
			t.Fatalf("%s: blocchi scritti", c.name)
		}
	}
	// la radio smette di rispondere: errore dopo i nuovi tentativi
	e = newK5v3BootEmu("7.00.07")
	e.silent = 10
	if _, started, err := run(e, "7.00.16", fw); err == nil || !started {
		t.Fatalf("radio muta: %v", err)
	}
}
