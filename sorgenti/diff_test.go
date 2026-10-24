//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Confronto con il codice originale dei programmi del costruttore (decompilato
// e compilato in un programma di prova, vedi sorgenti/verifica): lo stesso
// file viene scritto in una radio simulata sia dal codice originale sia da
// questo programma, e si confrontano i byte arrivati alla radio.
// Si esegue solo se la variabile HARNESS indica il programma di prova.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func harness(t *testing.T, args ...string) {
	t.Helper()
	h := os.Getenv("HARNESS")
	out, err := exec.Command("dotnet", append([]string{h}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("programma originale %v: %v\n%s", args, err, out)
	}
}

func needHarness(t *testing.T) {
	if os.Getenv("HARNESS") == "" {
		t.Skip("HARNESS non impostata")
	}
}

func diffRanges(a, b []byte) string {
	var parts []string
	n := 0
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			j := i
			for j < len(a) && a[j] != b[j] {
				j++
			}
			n += j - i
			if len(parts) < 12 {
				parts = append(parts, fmt.Sprintf("0x%04X-0x%04X (%x -> %x)", i, j-1, a[i:min(j, i+8)], b[i:min(j, i+8)]))
			}
			i = j
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d byte diversi: %s", n, strings.Join(parts, "; "))
}

// file modificato con questo programma: spostamenti, cancellazioni, nuovi
// canali, nomi, toni (anche DCS), potenza, banda, scansione
func k67Modified(t *testing.T) []byte {
	src := mustRead(t, sampleXLC)
	chs := readXLC(src, &Report{})
	chs[299], chs[85] = chs[85], nil
	chs[10] = nil
	chs[160] = &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "NUOVO R0", Scan: true}
	chs[161] = &Chan{Rx: 430100000, Tx: 431700000, TxTone: "D023N", RxTone: "D023N", Power: 2, Narrow: true, Name: "DCS"}
	chs[90].Name = "PROVA ALTO"
	chs[90].TxTone = "71.9"
	chs[100].Rx, chs[100].Tx = 145575000, 144975000
	chs[5].Scan = true
	return writeXLC(src, chs, &Report{})
}

func TestDiffK67Write(t *testing.T) {
	needHarness(t)
	dir := t.TempDir()
	for _, tc := range []struct {
		name string
		file []byte
	}{{"esempio", mustRead(t, sampleXLC)}, {"modificato", k67Modified(t)}} {
		path := filepath.Join(dir, tc.name+".xlc")
		os.WriteFile(path, tc.file, 0o644)
		radioMem := k67Sample(t)
		radioMem[0x10] ^= 0x01 // informazioni della radio diverse dal file
		e1 := &k67Emu{mem: append([]byte{}, radioMem...), model: "5RH-PRO"}
		addr, stop := serveTCP(e1.serve)
		harness(t, "k67-write", addr, path)
		stop()
		e2 := &k67Emu{mem: append([]byte{}, radioMem...), model: "5RH-PRO"}
		defer withEmu(t, e2.serve)()
		chs := readXLC(tc.file, &Report{})
		img := xlcImageForRadio(writeXLC(tc.file, chs, &Report{}))
		if _, err := k67Write(context.Background(), "COM1", img, &k67Opts{FreqBand: tc.file[xlcSize-1]}, noProg); err != nil {
			t.Fatal(err)
		}
		waitFor(func() bool { e2.mu.Lock(); defer e2.mu.Unlock(); return len(e2.log) >= len(e1.log) })
		if d := diffRanges(e1.mem, e2.mem); d != "" {
			t.Errorf("%s: memoria della radio diversa (originale -> questo programma): %s", tc.name, d)
		} else {
			t.Logf("%s: memoria della radio identica (%d byte)", tc.name, len(e1.mem))
		}
		if strings.Join(e1.log, "\n") != strings.Join(e2.log, "\n") {
			t.Errorf("%s: sequenza dei comandi diversa:\n%v\n%v", tc.name, e1.log, e2.log)
		} else {
			t.Logf("%s: stessa sequenza di %d comandi", tc.name, len(e1.log))
		}
	}
}

func TestDiffK67Read(t *testing.T) {
	needHarness(t)
	dir := t.TempDir()
	mem := k67Sample(t)
	e1 := &k67Emu{mem: append([]byte{}, mem...), model: "5RH-PRO", band: 0}
	addr, stop := serveTCP(e1.serve)
	out := filepath.Join(dir, "letto.xlc")
	harness(t, "k67-read", addr, out)
	stop()
	cps := mustRead(t, out)
	e2 := &k67Emu{mem: append([]byte{}, mem...), model: "5RH-PRO", band: 0}
	defer withEmu(t, e2.serve)()
	img, info, err := k67Read(context.Background(), "COM1", &k67Opts{}, noProg)
	if err != nil {
		t.Fatal(err)
	}
	mine := xlcFileFromRadio(img, info.FreqBand, info.K6, time.Now())
	copy(mine[96:112], cps[96:112]) // data e ora del salvataggio
	if d := diffRanges(cps, mine); d != "" {
		t.Errorf("file salvato dal programma originale dopo la lettura vs file di questo programma: %s", d)
	} else {
		t.Logf("file dopo la lettura identico a quello del programma originale (data a parte)")
	}
	// i canali devono coincidere
	a, b := readXLC(cps, &Report{}), readXLC(mine, &Report{})
	for i := range a {
		if (a[i] == nil) != (b[i] == nil) || (a[i] != nil && (a[i].Rx != b[i].Rx || a[i].Tx != b[i].Tx || a[i].Name != b[i].Name || a[i].TxTone != b[i].TxTone || a[i].RxTone != b[i].RxTone || a[i].Power != b[i].Power || a[i].Narrow != b[i].Narrow || a[i].Scan != b[i].Scan)) {
			t.Fatalf("canale %d diverso: %+v / %+v", i+1, a[i], b[i])
		}
	}
	if strings.Join(e1.log, "\n") != strings.Join(e2.log, "\n") {
		t.Errorf("sequenza dei comandi diversa:\n%v\n%v", e1.log, e2.log)
	}
}

func btModified(t *testing.T) []byte {
	src := mustRead(t, sampleDAT)
	chs, _ := readDAT(src, &Report{})
	chs[9*64+4], chs[64+21] = chs[64+21], nil
	chs[0].Name = "NUOVO NOME"
	chs[0].Power = 1
	chs[0].Scan = false
	chs[5*64] = &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D023I", Name: "R0 PROVA", Scan: true, Narrow: true}
	out, err := writeDAT(src, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDiffBtWrite(t *testing.T) {
	needHarness(t)
	dir := t.TempDir()
	for _, tc := range []struct {
		name string
		dat  []byte
	}{{"esempio", mustRead(t, sampleDAT)}, {"modificato", btModified(t)}} {
		path := filepath.Join(dir, tc.name+".dat")
		os.WriteFile(path, tc.dat, 0o644)
		start := bytes.Repeat([]byte{0x5A}, btImgSize)
		e1 := newBtEmu(start)
		addr, stop := serveTCP(e1.serve)
		harness(t, "bt-write", addr, path)
		stop()
		e2 := newBtEmu(start)
		defer withEmu(t, e2.serve)()
		img, err := datToEEPROM(tc.dat)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := btWriteImage(context.Background(), "COM1", img, noProg, nil); err != nil {
			t.Fatal(err)
		}
		waitFor(func() bool { e2.mu.Lock(); defer e2.mu.Unlock(); return e2.ended })
		if d := diffRanges(e1.mem, e2.mem); d != "" {
			t.Errorf("%s: memoria della radio diversa: %s", tc.name, d)
		} else {
			t.Logf("%s: memoria della radio identica (%d byte)", tc.name, len(e1.mem))
		}
		if strings.Join(e1.log, "\n") != strings.Join(e2.log, "\n") {
			t.Errorf("%s: sequenza dei comandi diversa (%d / %d)", tc.name, len(e1.log), len(e2.log))
		} else {
			t.Logf("%s: stessa sequenza di %d comandi", tc.name, len(e1.log))
		}
	}
}

func TestDiffBtRead(t *testing.T) {
	needHarness(t)
	dir := t.TempDir()
	img := btSampleImage(t)
	// alcuni dati in più rispetto all'esempio: DCS, FHSS, nomi FM
	mod, err := datToEEPROM(btModified(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		img  []byte
	}{{"esempio", img}, {"modificato", mod}} {
		e1 := newBtEmu(tc.img)
		addr, stop := serveTCP(e1.serve)
		out := filepath.Join(dir, tc.name+".dat")
		harness(t, "bt-read", addr, out)
		stop()
		e2 := newBtEmu(tc.img)
		defer withEmu(t, e2.serve)()
		got, _, err := btReadImage(context.Background(), "COM1", noProg, nil)
		if err != nil {
			t.Fatal(err)
		}
		mine, err := eepromToDAT(got, baseDAT)
		if err != nil {
			t.Fatal(err)
		}
		fa, fb := datFields(t, mustRead(t, out)), datFields(t, mine)
		diff := 0
		for k, v := range fa {
			if fb[k] != v {
				diff++
				if diff <= 10 {
					t.Logf("%s: campo %s: originale %q, questo programma %q", tc.name, k, v, fb[k])
				}
			}
		}
		if diff > 0 || len(fa) != len(fb) {
			t.Errorf("%s: %d campi diversi (%d/%d)", tc.name, diff, len(fa), len(fb))
		} else {
			t.Logf("%s: %d campi identici", tc.name, len(fa))
		}
		waitFor(func() bool { e2.mu.Lock(); defer e2.mu.Unlock(); return e2.ended })
		if strings.Join(e1.log, "\n") != strings.Join(e2.log, "\n") {
			t.Errorf("%s: sequenza dei comandi diversa (%d / %d)", tc.name, len(e1.log), len(e2.log))
		}
	}
}

// Il programma originale apre il file salvato da questo programma e lo
// risalva: a parte data e voci di zona (vedi xlcFreshZones) deve restare uguale.
func TestDiffK67SaveCompat(t *testing.T) {
	needHarness(t)
	dir := t.TempDir()
	for _, tc := range []struct {
		name string
		file []byte
	}{{"esempio", mustRead(t, sampleXLC)}, {"modificato", k67Modified(t)}} {
		in, out := filepath.Join(dir, tc.name+".xlc"), filepath.Join(dir, tc.name+"_cps.xlc")
		os.WriteFile(in, tc.file, 0o644)
		addr, stop := serveTCP(func(Port) error { select {} })
		harness(t, "k67-save", addr, in, out)
		stop()
		cps := mustRead(t, out)
		want := append([]byte{}, tc.file...)
		xlcFreshZones(want)
		copy(want[96:112], cps[96:112])
		if d := diffRanges(want, cps); d != "" {
			t.Errorf("%s: file risalvato dal programma originale diverso: %s", tc.name, d)
		} else {
			t.Logf("%s: il programma originale risalva il file senza altre modifiche", tc.name)
		}
	}
}
