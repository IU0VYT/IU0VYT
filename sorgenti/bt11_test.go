//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bajeton BJ-8300 / BJ-7800: file .dat e lettura/scrittura via cavo confrontati
// con il codice ORIGINALE del programma «BT-11_CPS» v1.0.8 e v1.0.9 (verifica/bt11),
// collegato alla radio simulata:
//   BT11_HARNESS=…/bt11_harness.dll BT108_EXE=…/v1.0.8/BT-11_CPS.exe BT11_EXE=…/v1.0.9/BT-11_CPS.exe
// 1) lettura: il file salvato dal programma dopo aver letto la radio ha gli stessi
//    valori (campo per campo) del file creato qui dalla stessa memoria;
// 2) scrittura: il programma scrive quel file nella radio: memoria identica a quella
//    scritta qui;
// 3) impostazioni: letture uguali; combinazioni casuali di modifiche fatte qui e dal
//    programma (proprietà di KDH.FunConfigData, RadioData.SaveToFile): file identici.

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type bt11Ver struct {
	name, exe, tmpl string
}

func bt11Env(t *testing.T) (string, []bt11Ver) {
	h := os.Getenv("BT11_HARNESS")
	if h == "" || (os.Getenv("BT108_EXE") == "" && os.Getenv("BT11_EXE") == "") {
		t.Skip("BT11_HARNESS e BT108_EXE / BT11_EXE non impostati")
	}
	var vs []bt11Ver
	if e := os.Getenv("BT108_EXE"); e != "" {
		vs = append(vs, bt11Ver{"v1.0.8", e, "templates/bajeton_base.dat"})
	}
	if e := os.Getenv("BT11_EXE"); e != "" {
		vs = append(vs, bt11Ver{"v1.0.9", e, "templates/bajeton_base109.dat"})
	}
	return h, vs
}

func bt11Run(t *testing.T, h string, args ...string) string {
	t.Helper()
	cmd := exec.Command("dotnet", append([]string{h}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("programma del costruttore %v: %v\n%s", args[1:], err, ee.Stderr)
		}
		t.Fatal(err)
	}
	return string(out)
}

// bt11CPSRead: il programma legge la radio simulata (memoria img) e salva il file
func bt11CPSRead(t *testing.T, h, exe string, img []byte) []byte {
	t.Helper()
	emu := newBtEmu(img)
	addr, stop := serveHarnessFn(func(p Port, _ int) { _ = emu.serve(p) })
	defer stop()
	out := filepath.Join(t.TempDir(), "letto.dat")
	bt11Run(t, h, exe, "read", addr, out)
	if !emu.ended {
		t.Fatal("il programma non ha chiuso la sessione")
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// bt11CPSWrite: il programma scrive il file nella radio simulata (memoria vuota)
func bt11CPSWrite(t *testing.T, h, exe string, dat []byte) []byte {
	t.Helper()
	emu := newBtEmu(bytes.Repeat([]byte{0xFF}, btImgSize))
	addr, stop := serveHarnessFn(func(p Port, _ int) { _ = emu.serve(p) })
	defer stop()
	in := filepath.Join(t.TempDir(), "da_scrivere.dat")
	_ = os.WriteFile(in, dat, 0o644)
	bt11Run(t, h, exe, "write", addr, in)
	if !emu.ended {
		t.Fatal("il programma non ha chiuso la sessione")
	}
	return emu.mem[:btImgSize]
}

func bt11Dump(t *testing.T, h, exe string, dat []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.dat")
	_ = os.WriteFile(p, dat, 0o644)
	return bt11Run(t, h, exe, "dump", p)
}

func diffLines(a, b string) string {
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	var out []string
	for i := 0; i < len(la) || i < len(lb); i++ {
		var x, y string
		if i < len(la) {
			x = la[i]
		}
		if i < len(lb) {
			y = lb[i]
		}
		if x != y {
			out = append(out, fmt.Sprintf("qui %q / programma %q", x, y))
			if len(out) >= 8 {
				break
			}
		}
	}
	return strings.Join(out, "\n")
}

// bt11Images: memorie di prova (esempio reale e varianti con impostazioni casuali)
func bt11Images(t *testing.T) map[string][]byte {
	base := btSampleImage(t)
	out := map[string][]byte{"esempio": base}
	rnd := rand.New(rand.NewSource(8300))
	for k := 0; k < 2; k++ {
		img := append([]byte{}, base...)
		for i, m := range btFun1Mod {
			if m > 0 {
				img[0x9000+i] = byte(rnd.Intn(m))
			}
		}
		img[0x9000+26] = byte(rnd.Intn(2)<<4 | rnd.Intn(2))
		for _, x := range btFun2 {
			img[0x9020+x.pos] = byte(rnd.Intn(x.mod))
		}
		img[0x9020+8] = byte(rnd.Intn(2))   // cbB_DivideCH (v1.0.9)
		img[0xB000+151] = byte(rnd.Intn(4)) // passo AM: 3 valori nella v1.0.8, 4 nella v1.0.9
		out[fmt.Sprintf("impostazioni %d", k+1)] = img
	}
	return out
}

func TestBT11CPS(t *testing.T) {
	h, vs := bt11Env(t)
	imgs := bt11Images(t)
	for _, v := range vs {
		t.Run(v.name, func(t *testing.T) {
			if os.Getenv("BT11_MKTEMPLATE") != "" && v.name == "v1.0.9" {
				// modello per i file della v1.0.9: il programma legge una radio con il
				// contenuto del modello della v1.0.8
				img, err := datToEEPROM(mustRead(t, "templates/bajeton_base.dat"))
				if err != nil {
					t.Fatal(err)
				}
				_ = os.WriteFile(v.tmpl, bt11CPSRead(t, h, v.exe, img), 0o644)
			}
			tmpl := mustRead(t, v.tmpl)
			if _, _, err := readRadio(tmpl, &Report{}); err != nil {
				t.Fatalf("modello: %v", err)
			}
			// file dell'altra versione convertito per questa: il programma lo apre e lo
			// scrive nella radio come l'originale
			src := mustRead(t, sampleDAT)
			if v.name == "v1.0.8" {
				var err error
				if src, err = datToVer(src, "1.0.9"); err != nil {
					t.Fatal(err)
				}
			}
			want0, err := datToEEPROM(src)
			if err != nil {
				t.Fatal(err)
			}
			conv, err := datToVer(src, strings.TrimPrefix(v.name, "v"))
			if err != nil || datVer(conv) != strings.TrimPrefix(v.name, "v") {
				t.Fatalf("conversione: %v (%s)", err, datVer(conv))
			}
			_ = bt11Dump(t, h, v.exe, conv)
			if mem := bt11CPSWrite(t, h, v.exe, conv); !bytes.Equal(mem, want0) {
				t.Fatalf("file convertito: scrittura diversa: %s", diffRanges(want0, mem))
			}
			for name, img := range imgs {
				theirs := bt11CPSRead(t, h, v.exe, img)
				ours, err := eepromToDAT(img, tmpl)
				if err != nil {
					t.Fatal(err)
				}
				if d := diffLines(bt11Dump(t, h, v.exe, ours), bt11Dump(t, h, v.exe, theirs)); d != "" {
					t.Fatalf("%s: lettura diversa dal programma:\n%s", name, d)
				}
				same := bytes.Equal(ours, theirs)
				mem := bt11CPSWrite(t, h, v.exe, ours)
				want, err := datToEEPROM(ours)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(mem, want) {
					t.Fatalf("%s: scrittura diversa dal programma: %s", name, diffRanges(want, mem))
				}
				// e il file del programma, convertito qui, dà la stessa memoria
				want2, err := datToEEPROM(theirs)
				if err != nil || !bytes.Equal(want2, want) {
					t.Fatalf("%s: file del programma convertito in modo diverso (%v)", name, err)
				}
				t.Logf("%s: lettura uguale campo per campo (file identico byte per byte: %v), scrittura identica", name, same)
			}
		})
	}
}

// valore della proprietà di KDH.FunConfigData per il valore dell'impostazione
func bajPropVal(d *setDef, v string) string {
	if d.Type == "bool" {
		b, _ := strconv.ParseBool(v)
		if d.Enc.Table[1] != nil && *d.Enc.Table[1] == "00000000" { // scelta invertita (radio FM)
			b = !b
		}
		if b {
			return "1"
		}
		return "0"
	}
	for i, o := range d.Options {
		if o == v {
			return strconv.Itoa(i)
		}
	}
	return "?"
}

func TestSettingsBajCPS(t *testing.T) {
	h, vs := bt11Env(t)
	for _, v := range vs {
		t.Run(v.name, func(t *testing.T) {
			dir := t.TempDir()
			base := mustRead(t, v.tmpl)
			tb := settingsFor(kindDAT, base)
			if tb == nil {
				t.Fatal("tabella assente")
			}
			props := func(file string) map[string]string {
				m := map[string]string{}
				for _, l := range strings.Split(bt11Run(t, h, v.exe, "fun", file), "\n") {
					if k := strings.Index(l, "="); k > 0 {
						m[l[:k]] = strings.TrimRight(l[k+1:], "\r")
					}
				}
				return m
			}
			check := func(img []byte, file string) {
				pm := props(file)
				for i := range tb.Settings {
					d := &tb.Settings[i]
					val, err := readSetting(d, img)
					if err != nil {
						t.Errorf("%s: %v", d.Key, err)
						continue
					}
					if got, want := bajPropVal(d, val), pm[bajProp(d.Key)]; got != want {
						t.Errorf("%s: qui %q, programma %q", d.Key, got, want)
					}
				}
			}
			bp := filepath.Join(dir, "base.dat")
			_ = os.WriteFile(bp, base, 0o644)
			check(base, bp)
			rnd := rand.New(rand.NewSource(8300))
			total := 0
			for r := 0; r < 10; r++ {
				tb := settingsFor(kindDAT, base)
				mine := append([]byte{}, base...)
				var lines []string
				for _, k := range rnd.Perm(len(tb.Settings)) {
					d := &tb.Settings[k]
					var val string
					switch d.Type {
					case "bool":
						val = strconv.FormatBool(rnd.Intn(2) == 1)
					case "list":
						val = d.Options[rnd.Intn(len(d.Options))]
					}
					if err := writeSetting(d, mine, val); err != nil {
						t.Fatalf("%s = %q: %v", d.Key, val, err)
					}
					lines = append(lines, bajProp(d.Key)+"="+bajPropVal(d, val))
				}
				mp := filepath.Join(dir, fmt.Sprintf("mod%d.txt", r))
				op := filepath.Join(dir, fmt.Sprintf("out%d.dat", r))
				_ = os.WriteFile(mp, []byte(strings.Join(lines, "\n")), 0o644)
				bt11Run(t, h, v.exe, "fun", bp, mp, op)
				theirs, _ := os.ReadFile(op)
				if !bytes.Equal(mine, theirs) {
					t.Fatalf("giro %d: file diverso da quello del programma: %s", r, diffRanges(mine, theirs))
				}
				// il file modificato resta leggibile e si scrive nella radio come lo
				// scriverebbe il programma
				if _, _, err := readRadio(mine, &Report{}); err != nil {
					t.Fatal(err)
				}
				if r == 0 {
					want, err := datToEEPROM(mine)
					if err != nil {
						t.Fatal(err)
					}
					if mem := bt11CPSWrite(t, h, v.exe, mine); !bytes.Equal(mem, want) {
						t.Fatalf("scrittura delle impostazioni diversa dal programma: %s", diffRanges(want, mem))
					}
				}
				total += len(lines)
			}
			t.Logf("%d impostazioni lette come il programma del costruttore; %d modifiche in 10 combinazioni: file identici byte per byte", len(tb.Settings), total)
		})
	}
}
