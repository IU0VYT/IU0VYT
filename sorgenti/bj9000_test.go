//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bajeton BJ-9000: canali e impostazioni confrontati con il codice ORIGINALE del
// programma «BJ-9000_CPS» v1.1.1 (BT-RT950PRO_CPS.exe, verifica/bt11) collegato a
// una radio simulata che parla come la radio vera con quel programma (cifratura
// scelta a caso dal programma, area APRS con i comandi 'T'/'X'):
//   BT11_HARNESS=…/bt11_harness.dll BJ9000_CPS=…/BT-RT950PRO_CPS.exe
// 1) canali: il programma legge la memoria scritta qui e la riscrive: canali
//    identici byte per byte;
// 2) impostazioni: letture uguali; combinazioni casuali di modifiche fatte qui e
//    dal programma (proprietà di KDH.FunConfigData, poi scrittura nella radio):
//    bit delle impostazioni identici.

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// radio simulata per il programma del costruttore
type bj9kCPSEmu struct {
	mu    sync.Mutex
	mem   []byte // 0x0000-0xD2FF
	aprs  []byte // area APRS (128 byte, comandi 'T' e 'X')
	ended bool
}

func (e *bj9kCPSEmu) serve(p Port) error {
	mg, err := emuRead(p, 14)
	if err != nil {
		return err
	}
	if string(mg) != "PROGRAMBT9000U" {
		return fmt.Errorf("identificazione inattesa %q", mg)
	}
	p.Write([]byte{0x06})
	if b, err := emuRead(p, 1); err != nil || b[0] != 'F' {
		return errors.New("F atteso")
	}
	p.Write([]byte{0x01, 0x36, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20, 0x02, 0x00, 0x02, 0x60, 0x01, 0x03, 0x30, 0x04})
	if b, err := emuRead(p, 1); err != nil || b[0] != 'M' {
		return errors.New("M atteso")
	}
	p.Write([]byte("RT-950      "))
	pk, err := emuRead(p, 25)
	if err != nil {
		return err
	}
	if string(pk[:4]) != "SEND" {
		return errors.New("SEND atteso")
	}
	idx := 2*int(pk[4]&0x0F) + 1
	if pk[4]&0x20 != 0 {
		idx++
	}
	key := []byte(btSymbols[pk[4+idx]])
	p.Write([]byte{0x06})
	for {
		c, err := emuRead(p, 1)
		if err != nil {
			return err
		}
		if c[0] == 'E' {
			e.mu.Lock()
			e.ended = true
			e.mu.Unlock()
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		e.mu.Lock()
		area := e.mem
		if c[0] == 'T' || c[0] == 'X' {
			area = e.aprs
		}
		e.mu.Unlock()
		switch c[0] {
		case 'R', 'T':
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			copy(d, area[addr:])
			e.mu.Unlock()
			btCrypt(d, key)
			p.Write(append([]byte{c[0], h[0], h[1], h[2]}, d...))
		case 'W', 'X':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			btCrypt(d, key)
			e.mu.Lock()
			copy(area[addr:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}

func bj9kEnv(t *testing.T) (string, string) {
	h, exe := os.Getenv("BT11_HARNESS"), os.Getenv("BJ9000_CPS")
	if h == "" || exe == "" {
		t.Skip("BT11_HARNESS o BJ9000_CPS non impostati")
	}
	return h, exe
}

// bj9kCPS: il programma legge la memoria mem (read) o scrive il file dat in una
// radio con memoria iniziale mem (write); restituisce file letto e memoria finale
func bj9kCPS(t *testing.T, h, exe, op string, mem, dat []byte) ([]byte, []byte) {
	t.Helper()
	e := &bj9kCPSEmu{mem: append([]byte{}, mem...), aprs: bytes.Repeat([]byte{0xFF}, 128)}
	addr, stop := serveHarnessFn(func(p Port, _ int) { _ = e.serve(p) })
	defer stop()
	f := filepath.Join(t.TempDir(), "radio.dat")
	if op == "write" {
		_ = os.WriteFile(f, dat, 0o644)
	}
	bt11Run(t, h, exe, op, addr, f)
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.ended {
		t.Fatal("il programma non ha chiuso la sessione")
	}
	out, _ := os.ReadFile(f)
	return out, append([]byte{}, e.mem...)
}

// bj9kBase: memoria come la scrive il programma partendo da un file nuovo
func bj9kBase(t *testing.T, h, exe string) []byte {
	p := filepath.Join(t.TempDir(), "nuovo.dat")
	bt11Run(t, h, exe, "new", p)
	dat, _ := os.ReadFile(p)
	_, mem := bj9kCPS(t, h, exe, "write", bytes.Repeat([]byte{0xFF}, 0xD300), dat)
	return mem
}

func TestBJ9000CPS(t *testing.T) {
	h, exe := bj9kEnv(t)
	m := genByID["bj-9000"]
	base := bj9kBase(t, h, exe)
	chs := genChans(m, base)
	for _, e := range chirpEdits(m, base) {
		chs[e.N] = e.Chan
	}
	ours, err := genWrite(m, base, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	dat, _ := bj9kCPS(t, h, exe, "read", ours, nil)
	_, mem := bj9kCPS(t, h, exe, "write", bytes.Repeat([]byte{0xFF}, 0xD300), dat)
	if !bytes.Equal(mem[:0x7800], ours[:0x7800]) {
		t.Fatalf("canali riscritti dal programma diversi: %s", diffRanges(ours[:0x7800], mem[:0x7800]))
	}
	n := 0
	for _, c := range chs {
		if c != nil {
			n++
		}
	}
	t.Logf("%d canali scritti qui, letti e riscritti dal programma del costruttore: memoria dei canali identica", n)
}

func TestSettingsBJ9000CPS(t *testing.T) {
	h, exe := bj9kEnv(t)
	tb := settingsFor("bj-9000", make([]byte, 0xD300))
	seen := map[string]bool{}
	for _, d := range tb.Settings {
		if d.Key == "" || d.ItLabel == "" || d.ItDesc == "" || seen[d.Key] {
			t.Fatalf("impostazione senza chiave, nome o descrizione, o ripetuta: %+v", d)
		}
		seen[d.Key] = true
	}
	prop := map[string]string{}
	for _, b := range bj9kSetDefs {
		prop[b.key] = "CbB_" + b.field
	}
	base := bj9kBase(t, h, exe)
	dir := t.TempDir()
	props := func(dat []byte) map[string]string {
		p := filepath.Join(dir, "f.dat")
		_ = os.WriteFile(p, dat, 0o644)
		m := map[string]string{}
		for _, l := range strings.Split(bt11Run(t, h, exe, "fun", p), "\n") {
			if k := strings.Index(l, "="); k > 0 {
				m[l[:k]] = strings.TrimRight(l[k+1:], "\r")
			}
		}
		return m
	}
	pv := func(d *setDef, v string) string {
		if d.Type == "bool" {
			b, _ := strconv.ParseBool(v)
			if *d.Enc.Table[1] == "00" { // ON = 0
				b = !b
			}
			if b {
				return "1"
			}
			return "0"
		}
		return v
	}
	check := func(img []byte, what string) {
		dat, _ := bj9kCPS(t, h, exe, "read", img, nil)
		pm := props(dat)
		for i := range tb.Settings {
			d := &tb.Settings[i]
			v, err := readSetting(d, img)
			if err != nil {
				t.Errorf("%s %s: %v", what, d.Key, err)
				continue
			}
			if got, want := pv(d, v), pm[prop[d.Key]]; got != want {
				t.Errorf("%s %s: qui %q, programma %q", what, d.Key, got, want)
			}
		}
	}
	check(base, "base")
	rnd := rand.New(rand.NewSource(9000))
	total := 0
	for r := 0; r < 8; r++ {
		mine := append([]byte{}, base...)
		var lines []string
		for _, k := range rnd.Perm(len(tb.Settings)) {
			d := &tb.Settings[k]
			var v string
			if d.Type == "bool" {
				v = strconv.FormatBool(rnd.Intn(2) == 1)
			} else {
				v = d.Options[rnd.Intn(len(d.Options))]
			}
			if err := writeSetting(d, mine, v); err != nil {
				t.Fatal(err)
			}
			lines = append(lines, prop[d.Key]+"="+pv(d, v))
		}
		// il programma: legge la radio, cambia le impostazioni e le scrive
		dat, _ := bj9kCPS(t, h, exe, "read", base, nil)
		bp := filepath.Join(dir, "base.dat")
		mp := filepath.Join(dir, "mod.txt")
		op := filepath.Join(dir, "out.dat")
		_ = os.WriteFile(bp, dat, 0o644)
		_ = os.WriteFile(mp, []byte(strings.Join(lines, "\n")), 0o644)
		bt11Run(t, h, exe, "fun", bp, mp, op)
		mod, _ := os.ReadFile(op)
		_, theirs := bj9kCPS(t, h, exe, "write", base, mod)
		for i := range tb.Settings {
			d := &tb.Settings[i]
			a, _ := readSetting(d, mine)
			b, err := readSetting(d, theirs)
			if err != nil || a != b {
				t.Errorf("giro %d, %s: qui %q, programma %q (%v)", r, d.Key, a, b, err)
			}
		}
		if diff := k67SetDiff(mine, theirs, tb); diff != "" {
			t.Fatalf("giro %d: bit diversi: %s", r, diff)
		}
		// i canali non cambiano
		if !bytes.Equal(mine[:0x7800], theirs[:0x7800]) {
			t.Fatalf("giro %d: canali diversi", r)
		}
		if r == 0 {
			check(mine, "modificata")
		}
		total += len(lines)
	}
	t.Logf("%d impostazioni lette come il programma del costruttore; %d modifiche in 8 combinazioni: bit identici", len(tb.Settings), total)
}
