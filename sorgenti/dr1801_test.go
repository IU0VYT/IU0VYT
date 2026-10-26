//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// DR-1801UV confrontata con il codice originale del programma del costruttore
// «DR-1801A6» (verifica/dr1801/harness): lettura del file di fabbrica, modifiche
// (scritte da DataLayout.W del programma e da questo programma: byte identici)
// e protocollo via cavo (lettura e scrittura del programma con la radio simulata).
//   DR1801_HARNESS=…/dr1801_harness.dll DR1801_EXE=…/DR-1801A6.exe
//   DR1801_ACCPS=«…/DM-1801A6 Factory setting.accps»

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stesso formato della modalità decode del programma di prova
func drDump(img []byte) string {
	l, err := drParse(img)
	if err != nil {
		return err.Error()
	}
	c := l.ch()
	var sb strings.Builder
	avail := int(binary.LittleEndian.Uint16(img[c.base:]))
	type rec struct {
		idx  int
		line string
	}
	var recs []rec
	n := 0
	for p := 0; p < c.max && n < avail; p++ {
		r := c.rec(img, p)
		idx := int(binary.LittleEndian.Uint16(r))
		if idx == 0xFFFF {
			continue
		}
		n++
		u16 := func(o int) uint16 { return binary.LittleEndian.Uint16(r[o:]) }
		recs = append(recs, rec{idx, fmt.Sprintf("ch|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%s", idx, r[2], r[3],
			binary.LittleEndian.Uint32(r[4:]), binary.LittleEndian.Uint32(r[8:]), r[25], r[30], r[31], u16(28), r[34], r[35], u16(32),
			u16(36), u16(38), drStr(c.name(img, p)))})
	}
	// il programma elenca i canali in ordine di numero
	for i := 0; i < len(recs); i++ {
		for j := i + 1; j < len(recs); j++ {
			if recs[j].idx < recs[i].idx {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
	}
	for _, r := range recs {
		sb.WriteString(r.line + "\n")
	}
	join := func(v []uint16) string {
		var s []string
		for _, x := range v {
			s = append(s, fmt.Sprint(x))
		}
		return strings.Join(s, ",")
	}
	for _, z := range l.zones(img).list {
		fmt.Fprintf(&sb, "zone|%d|%s|%s\n", z.index, z.name, join(z.channels))
	}
	m := l.mods["SL"]
	p := m.params
	nameLen, infoSize := int(p[1]), int(binary.LittleEndian.Uint16(p[6:]))
	d := img[m.off:]
	for i := 0; i < int(d[0]); i++ {
		r := d[4+i*infoSize:]
		var v []uint16
		for j := 0; j < int(r[1]); j++ {
			v = append(v, binary.LittleEndian.Uint16(r[16+nameLen+2*j:]))
		}
		fmt.Fprintf(&sb, "scan|%d|%s\n", r[0], join(v))
	}
	if rm, ok := l.mods["RM"]; ok {
		d := img[rm.off+4:]
		for i := 0; i < 8; i++ {
			r := d[i*20:]
			var v []uint16
			for j := 0; j < int(binary.LittleEndian.Uint16(r[2:])) && j < 8; j++ {
				v = append(v, binary.LittleEndian.Uint16(r[4+2*j:]))
			}
			fmt.Fprintf(&sb, "roam|%d|%s\n", i, join(v))
		}
	}
	return sb.String()
}

func TestDr1801CPS(t *testing.T) {
	h, exe, file := os.Getenv("DR1801_HARNESS"), os.Getenv("DR1801_EXE"), os.Getenv("DR1801_ACCPS")
	if h == "" || exe == "" || file == "" {
		t.Skip("DR1801_HARNESS, DR1801_EXE o DR1801_ACCPS non impostati")
	}
	m := genByID["bf-dr1801uv"]
	base, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		out, err := exec.Command("dotnet", append([]string{h, exe}, args...)...).Output()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				t.Fatalf("programma del costruttore: %v\n%s", err, ee.Stderr)
			}
			t.Fatal(err)
		}
		return string(out)
	}
	dir := t.TempDir()
	// 1) lettura del file di fabbrica: stessi canali, zone e liste
	if got, want := drDump(base), run("decode", file); got != want {
		t.Fatalf("lettura diversa dal programma del costruttore:\nqui:\n%s\nprogramma:\n%s", got, want)
	}
	chs := genChans(m, base)
	if c := chs[20]; c == nil || c.Rx != 144625000 || c.TxTone != "63.0" || c.Dig || c.Narrow || c.Name != "CHL A 21" {
		t.Fatalf("canale 21 letto male: %+v", c)
	}
	if c := chs[0]; c == nil || !c.Dig || c.Rx != 144125000 || c.Name != "CHL DMR 1" {
		t.Fatalf("canale 1 letto male: %+v", c)
	}
	// 2) modifiche: cancellazioni (anche di canali in zone, liste e roaming),
	// canali modificati e nuovi, analogici e DMR
	type edit struct {
		n    int
		c    *Chan
		line string
	}
	edits := []edit{
		{5, nil, "del|5"},
		{0, nil, "del|0"},
		{21, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "88.5", Power: 2, Narrow: true, Name: "R0 ROMA"},
			"set|21|1|0|433500000|433500000|0|1|0|885|2|0|23|0|0|R0 ROMA"},
		{40, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D754I", Power: 0, Name: "RV48"},
			"set|40|1|1|145600000|145000000|1|2|1|754|1|0|885|0|0|RV48"},
		{41, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 0, Name: "DMR IR0"},
			"set|41|3|1|438300000|430700000|0|0|0|0|0|0|0|0|0|DMR IR0"},
		{42, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "69.4", Power: 0, Name: "TONO 69.4"},
			"set|42|1|1|145500000|145500000|1|0|0|0|1|0|694|%P1%|%P2%|TONO 69.4"},
	}
	p1, p2 := drCtcPara(694)
	var sb strings.Builder
	mine := append([]byte{}, base...)
	for _, e := range edits {
		sb.WriteString(strings.NewReplacer("%P1%", fmt.Sprint(p1), "%P2%", fmt.Sprint(p2)).Replace(e.line) + "\n")
		if err := m.Fam.Set(m, mine, e.n, e.c); err != nil {
			t.Fatalf("canale %d: %v", e.n+1, err)
		}
	}
	ef, out := filepath.Join(dir, "mod.txt"), filepath.Join(dir, "out.accps")
	_ = os.WriteFile(ef, []byte(sb.String()), 0o644)
	run("edit", file, ef, out)
	theirs, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mine, theirs) {
		t.Fatalf("memoria diversa da quella scritta dal programma del costruttore: %s", diffRanges(mine, theirs))
	}
	if got, want := drDump(mine), run("decode", out); got != want {
		t.Fatalf("rilettura diversa:\nqui:\n%s\nprogramma:\n%s", got, want)
	}
	for _, e := range edits {
		g := m.Fam.Get(m, mine, e.n)
		if (g == nil) != (e.c == nil) || g != nil && (g.Rx != e.c.Rx || g.Tx != e.c.Tx || g.TxTone != e.c.TxTone || g.RxTone != e.c.RxTone || g.Name != e.c.Name || g.Dig != e.c.Dig) {
			t.Errorf("canale %d riletto: %+v", e.n+1, g)
		}
	}
	// 3) protocollo: lettura e scrittura del programma con la radio simulata
	emu := genEmuFor(m, base)
	addr, stop := serveHarnessFn(func(p Port, baud int) { _ = emu.serve(p) })
	rf := filepath.Join(dir, "letto.accps")
	run("read", addr, rf)
	stop()
	got, _ := os.ReadFile(rf)
	if !bytes.Equal(got, base) {
		t.Fatalf("lettura del programma del costruttore diversa: %s", diffRanges(got, base))
	}
	eo := genEmuFor(m, base)
	addr, stop = serveHarnessFn(func(p Port, baud int) { _ = eo.serve(p) })
	run("write", addr, out)
	stop()
	eg := genEmuFor(m, base)
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) {
		a, b := newPipe()
		go eg.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	if _, err := m.Fam.Upload(context.Background(), m, "COM9", mine, noProg); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(eo.image(), eg.image()) {
		t.Fatalf("radio scritta dal programma del costruttore diversa: %s", diffRanges(eo.image(), eg.image()))
	}
	back, _, err := m.Fam.Download(context.Background(), m, "COM9", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(back, mine) {
		t.Fatalf("rilettura diversa: %s", diffRanges(back, mine))
	}
	t.Logf("%d canali letti uguali; %d modifiche: memoria identica a DataLayout.W; lettura e scrittura del programma del costruttore uguali", strings.Count(drDump(base), "ch|"), len(edits))
}
