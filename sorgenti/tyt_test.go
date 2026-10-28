//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// DM-1701 confrontata con il codice di dmrconfig (uv380.c, dfu-libusb.c),
// compilato con il collegamento TCP di verifica/dmrconfig/tcpdev.c:
//   DMRCONFIG=…/dmrconfig-tcp DMRCONFIG_EX=…/dmrconfig/examples
// 1) memoria di prova creata da dmrconfig con un suo file di esempio;
// 2) lettura: stessi canali; 3) modifiche fatte qui e le stesse tabelle applicate
// da dmrconfig: memoria identica (tranne la data di programmazione); 4) protocollo
// DFU: lettura e scrittura di dmrconfig con la radio simulata uguali a quelle di
// questo programma.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func tytConfName(s string) string {
	if s == "" {
		return "-"
	}
	return strings.ReplaceAll(s, " ", "_")
}

// tytConf: tabelle di canali, zone e liste di scansione nella sintassi di dmrconfig
func tytConf(img []byte) string {
	admit := []string{"-", "Free", "Tone", "Color"}
	power := []string{"Low", "Low", "Mid", "High"}
	width := []string{"12.5", "20", "25", "25"}
	var dig, ana strings.Builder
	dig.WriteString("Digital Name Receive Transmit Power Scan TOT RO Admit Color Slot RxGL TxContact\n")
	ana.WriteString("Analog Name Receive Transmit Power Scan TOT RO Admit Squelch RxTone TxTone Width\n")
	nd, na := 0, 0
	for i := 0; i < tytNCh; i++ {
		if !tytValid(img, i) {
			continue
		}
		r := tytRec(img, i)
		rx, _ := rdtyFreq(r[16:])
		tx, _ := rdtyFreq(r[20:])
		scan := "-"
		if r[11] != 0 {
			scan = strconv.Itoa(int(r[11]))
		}
		ro := "-"
		if r[1]>>1&1 == 1 {
			ro = "+"
		}
		row := fmt.Sprintf("  %d %s %.5f %.5f %s %s %d %s %s", i+1, tytConfName(tytStr(r[32:], 16)), float64(rx)/1e6, float64(tx)/1e6,
			power[r[30]&3], scan, int(r[8]&0x3F)*15, ro, admit[r[4]>>6])
		if r[0]&3 == 2 {
			gl, ct := "-", "-"
			if r[12] != 0 {
				gl = strconv.Itoa(int(r[12]))
			}
			if c := binary.LittleEndian.Uint16(r[6:]); c != 0 {
				ct = strconv.Itoa(int(c))
			}
			fmt.Fprintf(&dig, "%s %d %d %s %s\n", row, r[1]>>4, r[1]>>2&3, gl, ct)
			nd++
		} else {
			fmt.Fprintf(&ana, "%s %d %s %s %s\n", row, r[15], rdtyConfTone(binary.LittleEndian.Uint16(r[24:])),
				rdtyConfTone(binary.LittleEndian.Uint16(r[26:])), width[r[0]>>2&3])
			na++
		}
	}
	var sb strings.Builder
	if nd > 0 {
		sb.WriteString(dig.String() + "\n")
	}
	if na > 0 {
		sb.WriteString(ana.String() + "\n")
	}
	list := func(v []uint16) string {
		if len(v) == 0 {
			return "-"
		}
		var s []string
		for _, x := range v {
			s = append(s, strconv.Itoa(int(x)))
		}
		return strings.Join(s, ",")
	}
	sb.WriteString("Zone Name Channels\n")
	for z := 0; z < tytNZones; z++ {
		if !tytZoneValid(img, z) {
			continue
		}
		fmt.Fprintf(&sb, "  %da %s %s\n", z+1, tytConfName(tytStr(tytZone(img, z), 16)), list(tytZoneA(img, z).get()))
		fmt.Fprintf(&sb, "  %db - %s\n", z+1, list(tytZoneB(img, z).get()))
	}
	sb.WriteString("\nScanlist Name PCh1 PCh2 TxCh Channels\n")
	for s := 0; s < tytNScan; s++ {
		if !tytScanValid(img, s) {
			continue
		}
		sl := tytScan(img, s)
		pr := func(v uint16, none string) string {
			switch v {
			case 0xFFFF:
				return none
			case 0:
				return "Sel"
			}
			return strconv.Itoa(int(v))
		}
		fmt.Fprintf(&sb, "  %d %s %s %s %s %s\n", s+1, tytConfName(tytStr(sl, 16)), pr(binary.LittleEndian.Uint16(sl[32:]), "-"),
			pr(binary.LittleEndian.Uint16(sl[34:]), "-"), pr(binary.LittleEndian.Uint16(sl[36:]), "Last"), list(tytList{sl[42:104]}.get()))
	}
	return sb.String()
}

func TestTytDmrconfig(t *testing.T) {
	bin, ex := os.Getenv("DMRCONFIG"), os.Getenv("DMRCONFIG_EX")
	if bin == "" || ex == "" {
		t.Skip("DMRCONFIG o DMRCONFIG_EX non impostati")
	}
	m := genByID["bf-dm1701"]
	dir := t.TempDir()
	run := func(env []string, args ...string) string {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("dmrconfig %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	// 1) memoria di prova
	_ = os.WriteFile(filepath.Join(dir, "blank.img"), bytes.Repeat([]byte{0xFF}, tytMemSize), 0o644)
	run(nil, "-c", "blank.img", filepath.Join(ex, "uv380-south-bay-area.conf"))
	base, err := os.ReadFile(filepath.Join(dir, "device.img"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "base.img"), base, 0o644)
	// 2) lettura
	rows := rdtyDmrconfigChans(run(nil, "base.img"))
	n := 0
	for i := 0; i < m.Slots; i++ {
		c := m.Fam.Get(m, base, i)
		f, ok := rows[i+1]
		if (c != nil) != ok {
			t.Fatalf("canale %d: qui %v, dmrconfig %v", i+1, c, f)
		}
		if c == nil {
			continue
		}
		n++
		rx, _ := strconv.ParseFloat(f[2], 64)
		name := strings.ReplaceAll(f[1], "_", " ")
		pw := map[string]int{"High": 0, "Mid": 1, "Low": 2}[f[4]]
		if int64(rx*1e6+0.5) != c.Rx || name != c.Name || pw != c.Power || (f[0] == "Digital") != c.Dig || (f[7] == "+") != c.TxOff {
			t.Errorf("canale %d: dmrconfig %v, qui %+v", i+1, f, c)
		}
		if f[0] == "Analog" {
			tone := func(s string) string {
				if s == "-" {
					return ""
				}
				if strings.HasPrefix(s, "D") {
					return s
				}
				v, _ := strconv.ParseFloat(s, 64)
				return fmtTone(int(v*10 + 0.5))
			}
			if tone(f[10]) != c.RxTone || tone(f[11]) != c.TxTone || (f[12] == "12.5") != c.Narrow {
				t.Errorf("canale %d (analogico): dmrconfig %v, qui %+v", i+1, f, c)
			}
		}
	}
	if n < 10 {
		t.Fatalf("solo %d canali nella memoria di prova", n)
	}
	// 3) modifiche: cancellazioni (anche di canali in zone e liste), canali
	// modificati (anche da DMR ad analogico e viceversa), nuovi, solo ricezione
	mine := append([]byte{}, base...)
	type edit struct {
		n int
		c *Chan
	}
	var edits []edit
	var inZone, inScan, ana, dig = -1, -1, -1, -1
	for z := 0; z < tytNZones && inZone < 0; z++ {
		if tytZoneValid(base, z) {
			if v := tytZoneA(base, z).get(); len(v) > 1 {
				inZone = int(v[1]) - 1
			}
		}
	}
	for s := 0; s < tytNScan && inScan < 0; s++ {
		if tytScanValid(base, s) {
			if v := (tytList{tytScan(base, s)[42:104]}).get(); len(v) > 0 && int(v[0])-1 != inZone {
				inScan = int(v[0]) - 1
			}
		}
	}
	for i := 0; i < m.Slots; i++ {
		c := m.Fam.Get(m, base, i)
		if c == nil || i == inZone || i == inScan {
			continue
		}
		if c.Dig && dig < 0 {
			dig = i
		}
		if !c.Dig && ana < 0 {
			ana = i
		}
	}
	if ana < 0 {
		// nessun canale analogico: un altro canale DMR diventa analogico
		for i := dig + 1; i < m.Slots && ana < 0; i++ {
			if c := m.Fam.Get(m, base, i); c != nil && i != inZone && i != inScan {
				ana = i
			}
		}
	}
	free := []int{}
	for i := 2900; i < 3000 && len(free) < 3; i++ {
		if !tytValid(base, i) {
			free = append(free, i)
		}
	}
	for _, d := range []int{inZone, inScan} {
		if d >= 0 {
			edits = append(edits, edit{d, nil})
		}
	}
	edits = append(edits,
		edit{ana, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D023N", TxTone: "88.5", Power: 1, Narrow: true, Name: "R0 ROMA"}},
		edit{dig, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "D754I", Power: 2, Name: "RV48"}},
		edit{free[0], &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 0, Name: "DMR IR0UGN"}},
		edit{free[1], &Chan{Rx: 145500000, Tx: 145500000, TxOff: true, Power: 2, Narrow: true, Name: "ASCOLTO"}},
		edit{free[2], &Chan{Rx: 145750000, Tx: 145150000, TxTone: "67.0", Power: 0, Name: "RV60"}},
	)
	if ana < 0 || dig < 0 || len(free) < 3 || inZone < 0 {
		t.Fatalf("memoria di prova inadatta: %d %d %d %v", ana, dig, inZone, free)
	}
	for _, e := range edits {
		if err := m.Fam.Set(m, mine, e.n, e.c); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range edits {
		g := m.Fam.Get(m, mine, e.n)
		if (g == nil) != (e.c == nil) || g != nil && (g.Rx != e.c.Rx || g.Name != e.c.Name || g.Dig != e.c.Dig || g.TxTone != e.c.TxTone ||
			g.RxTone != e.c.RxTone || g.TxOff != e.c.TxOff || g.Power != e.c.Power) {
			t.Errorf("canale %d riletto: %+v (scritto %+v)", e.n+1, g, e.c)
		}
	}
	_ = os.WriteFile(filepath.Join(dir, "mod.conf"), []byte(tytConf(mine)), 0o644)
	run(nil, "-c", "base.img", "mod.conf")
	theirs, _ := os.ReadFile(filepath.Join(dir, "device.img"))
	for k := 0x2001; k < 0x200C; k++ { // data di programmazione
		theirs[k] = mine[k]
	}
	if !bytes.Equal(theirs, mine) {
		t.Fatalf("memoria diversa da quella scritta da dmrconfig: %s", diffRanges(mine, theirs))
	}
	// 4) protocollo DFU
	emu := genEmuFor(m, base)
	addr, stop := serveRawTCP(func(p Port) { _ = emu.serve(p) })
	run([]string{"DMRCONFIG_DFU=" + addr}, "-r")
	stop()
	read, _ := os.ReadFile(filepath.Join(dir, "device.img"))
	old := openPortFunc
	defer func() { openPortFunc = old }()
	var cur genEmu
	openPortFunc = func(string, int) (Port, error) {
		a, b := newPipe()
		go cur.serve(b)
		return a, nil
	}
	cur = genEmuFor(m, base)
	got, info, err := m.Fam.Download(context.Background(), m, "USB", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read, got) || !bytes.Equal(got, base) {
		t.Fatalf("lettura di dmrconfig diversa: %s", diffRanges(read, got))
	}
	_ = os.WriteFile(filepath.Join(dir, "mine.img"), mine, 0o644)
	eo := genEmuFor(m, tytSynthImage())
	addr, stop = serveRawTCP(func(p Port) { _ = eo.serve(p) })
	run([]string{"DMRCONFIG_DFU=" + addr}, "-w", "mine.img")
	stop()
	cur = genEmuFor(m, tytSynthImage())
	if _, err := m.Fam.Upload(context.Background(), m, "USB", mine, noProg); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(eo.image(), cur.image()) || !bytes.Equal(cur.image(), mine) {
		t.Fatalf("radio scritta da dmrconfig diversa: %s", diffRanges(eo.image(), cur.image()))
	}
	t.Logf("%s: %d canali letti uguali a dmrconfig; %d modifiche: memoria identica a quella di dmrconfig; lettura e scrittura DFU uguali", info, n, len(edits))
}
