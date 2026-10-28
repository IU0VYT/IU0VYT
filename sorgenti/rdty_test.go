//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// RD-5R e DM-1801 confrontate con il codice di dmrconfig (rd5r.c, dm1801.c,
// hid.c), compilato con il collegamento TCP di verifica/dmrconfig/tcpdev.c:
//   DMRCONFIG=…/dmrconfig-tcp DMRCONFIG_EX=…/dmrconfig/examples
// 1) memoria di prova creata da dmrconfig con un suo file di esempio;
// 2) lettura: stessi canali (frequenze, nomi, potenza, toni, banda, modo);
// 3) modifiche fatte qui e le stesse tabelle (canali, zone, liste) applicate da
//    dmrconfig alla stessa memoria: memoria identica (tranne la data di
//    programmazione, che dmrconfig aggiorna);
// 4) protocollo: lettura e scrittura di dmrconfig con la radio simulata uguali a
//    quelle di questo programma.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func rdtyConfName(s string) string {
	if s == "" {
		return "-"
	}
	return strings.ReplaceAll(s, " ", "_")
}

func rdtyConfTone(v uint16) string {
	t := rdtyTone(v)
	if t == "" {
		return "-"
	}
	return t
}

// rdtyConf: tabelle di canali, zone e liste di scansione nella sintassi di dmrconfig
func rdtyConf(m *genModel, img []byte) string {
	o := rdtyo(m)
	var dig, ana strings.Builder
	dig.WriteString("Digital Name Receive Transmit Power Scan TOT RO Admit Color Slot RxGL TxContact\n")
	ana.WriteString("Analog Name Receive Transmit Power Scan TOT RO Admit Squelch RxTone TxTone Width\n")
	nd, na := 0, 0
	for i := 0; i < m.Slots; i++ {
		if !rdtyValid(img, i) {
			continue
		}
		r := img[rdtyRecOff(i):][:rdtyChSize]
		rx, _ := rdtyFreq(r[16:])
		tx, _ := rdtyFreq(r[20:])
		pw := "Low"
		if r[51]>>7&1 == 1 {
			pw = "High"
		}
		scan := "-"
		if r[31] != 0 {
			scan = strconv.Itoa(int(r[31]))
		}
		ro := "-"
		if r[51]>>2&1 == 1 {
			ro = "+"
		}
		row := fmt.Sprintf("  %d %s %.5f %.5f %s %s %d %s", i+1, rdtyConfName(rdtyName(r[0:16])), float64(rx)/1e6, float64(tx)/1e6, pw, scan, int(r[27])*15, ro)
		if r[24] == 1 {
			admit := []string{"-", "Free", "Color"}[r[29]]
			slot := 1
			if r[49]>>6&1 == 1 {
				slot = 2
			}
			gl, ct := "-", "-"
			if r[43] != 0 {
				gl = strconv.Itoa(int(r[43]))
			}
			if c := binary.LittleEndian.Uint16(r[46:]); c != 0 {
				ct = strconv.Itoa(int(c))
			}
			fmt.Fprintf(&dig, "%s %s %d %d %s %s\n", row, admit, r[42], slot, gl, ct)
			nd++
		} else {
			admit := []string{"-", "Free", "Tone"}[r[29]]
			sq := strconv.Itoa(int(r[55]))
			if o.sqBit {
				sq = []string{"Tight", "Normal"}[r[51]&1]
			}
			w := "12.5"
			if r[51]>>1&1 == 1 {
				w = "25"
			}
			fmt.Fprintf(&ana, "%s %s %s %s %s %s\n", row, admit, sq, rdtyConfTone(binary.LittleEndian.Uint16(r[32:])),
				rdtyConfTone(binary.LittleEndian.Uint16(r[34:])), w)
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
	list := func(b []byte, n, sub int) string {
		var s []string
		for k := 0; k < n; k++ {
			if v := binary.LittleEndian.Uint16(b[2*k:]); v != 0 {
				s = append(s, strconv.Itoa(int(v)-sub))
			}
		}
		return strings.Join(s, ",")
	}
	sb.WriteString("Zone Name Channels\n")
	for z := 0; z < o.nzones; z++ {
		if !rdtyZoneValid(img, z) {
			continue
		}
		zr := img[rdtyZoneOff(o, z):]
		ch := list(zr[16:], o.zoneMem, 0)
		if ch == "" {
			ch = "-"
		}
		fmt.Fprintf(&sb, "  %d %s %s\n", z+1, rdtyConfName(rdtyName(zr[:16])), ch)
	}
	sb.WriteString("\nScanlist Name PCh1 PCh2 TxCh Channels\n")
	for s := 0; s < o.nscan; s++ {
		if img[o.scanOff+s] != 1 {
			continue
		}
		sl := img[o.scanOff+o.scanValid+s*rdtyScanLen:][:rdtyScanLen]
		pr := func(v uint16, zero string) string {
			switch v {
			case 0:
				return zero
			case 1:
				return "Sel"
			}
			return strconv.Itoa(int(v) - 1)
		}
		ch := list(sl[18:], 31, 1)
		if ch == "" {
			ch = "Sel"
		}
		fmt.Fprintf(&sb, "  %d %s %s %s %s %s\n", s+1, rdtyConfName(rdtyName(sl[:15])), pr(binary.LittleEndian.Uint16(sl[80:]), "-"),
			pr(binary.LittleEndian.Uint16(sl[82:]), "-"), pr(binary.LittleEndian.Uint16(sl[84:]), "Last"), ch)
	}
	return sb.String()
}

// righe dei canali stampate da dmrconfig: numero -> campi
func rdtyDmrconfigChans(conf string) map[int][]string {
	out := map[int][]string{}
	table := ""
	for _, l := range strings.Split(conf, "\n") {
		if i := strings.Index(l, "#"); i >= 0 {
			l = l[:i]
		}
		f := strings.Fields(l)
		if len(f) == 0 {
			table = ""
			continue
		}
		if f[0] == "Digital" || f[0] == "Analog" {
			table = f[0]
			continue
		}
		if table == "" {
			continue
		}
		n, err := strconv.Atoi(f[0])
		if err != nil {
			continue
		}
		out[n] = append([]string{table}, f[1:]...)
	}
	return out
}

func TestRdtyDmrconfig(t *testing.T) {
	bin, ex := os.Getenv("DMRCONFIG"), os.Getenv("DMRCONFIG_EX")
	if bin == "" || ex == "" {
		t.Skip("DMRCONFIG o DMRCONFIG_EX non impostati")
	}
	for _, v := range []struct{ id, conf string }{{"bf-rd5r", "rd5r-south-bay-area.conf"}, {"bf-dm1801", "dm1801-south-bay-area.conf"}} {
		m := genByID[v.id]
		o := rdtyo(m)
		t.Run(v.id, func(t *testing.T) {
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
			blank := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
			copy(blank, o.hdr)
			_ = os.WriteFile(filepath.Join(dir, "blank.img"), blank, 0o644)
			run(nil, "-c", "blank.img", filepath.Join(ex, v.conf))
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
				if f[1] == "_" {
					name = ""
				}
				pw := map[string]int{"High": 0, "Low": 2}[f[4]]
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
			// 3) modifiche: cancellazioni (anche di canali in zone e liste),
			// canali modificati, nuovi analogici e DMR, solo ricezione
			mine := append([]byte{}, base...)
			var del []int
			for _, z := range []int{0, 11, 12} {
				if rdtyZoneValid(base, z) {
					if c := binary.LittleEndian.Uint16(base[rdtyZoneOff(o, z)+16:]); c != 0 {
						del = append(del, int(c)-1)
					}
				}
			}
			type edit struct {
				n int
				c *Chan
			}
			edits := []edit{}
			for _, d := range del {
				edits = append(edits, edit{d, nil})
			}
			// un canale qualsiasi da modificare: il primo analogico rimasto
			mod := -1
			for i := 0; i < m.Slots && mod < 0; i++ {
				if c := m.Fam.Get(m, base, i); c != nil && !c.Dig {
					found := false
					for _, d := range del {
						found = found || d == i
					}
					if !found {
						mod = i
					}
				}
			}
			free := []int{}
			for i := 900; i < 1024 && len(free) < 3; i++ {
				if !rdtyValid(base, i) {
					free = append(free, i)
				}
			}
			edits = append(edits,
				edit{mod, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D023N", TxTone: "88.5", Power: 2, Narrow: true, Name: "R0 ROMA"}},
				edit{free[0], &Chan{Rx: 145600000, Tx: 145000000, TxTone: "D754I", Power: 0, Name: "RV48"}},
				edit{free[1], &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 0, Name: "DMR IR0UGN"}},
				edit{free[2], &Chan{Rx: 145500000, Tx: 145500000, TxOff: true, Power: 2, Narrow: true, Name: "ASCOLTO"}},
			)
			for _, e := range edits {
				if err := m.Fam.Set(m, mine, e.n, e.c); err != nil {
					t.Fatal(err)
				}
			}
			for _, e := range edits {
				g := m.Fam.Get(m, mine, e.n)
				if (g == nil) != (e.c == nil) || g != nil && (g.Rx != e.c.Rx || g.Name != e.c.Name || g.Dig != e.c.Dig || g.TxTone != e.c.TxTone || g.RxTone != e.c.RxTone || g.TxOff != e.c.TxOff) {
					t.Errorf("canale %d riletto: %+v (scritto %+v)", e.n+1, g, e.c)
				}
			}
			_ = os.WriteFile(filepath.Join(dir, "mod.conf"), []byte(rdtyConf(m, mine)), 0o644)
			run(nil, "-c", "base.img", "mod.conf")
			theirs, _ := os.ReadFile(filepath.Join(dir, "device.img"))
			for k := 0x88; k < 0x8E; k++ {
				theirs[k] = mine[k]
			}
			if !bytes.Equal(theirs, mine) {
				t.Fatalf("memoria diversa da quella scritta da dmrconfig: %s", diffRanges(mine, theirs))
			}
			// 4) protocollo
			emu := genEmuFor(m, base)
			addr, stop := serveRawTCP(func(p Port) { _ = emu.serve(p) })
			run([]string{"DMRCONFIG_HID=" + addr}, "-r")
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
			got, _, err := m.Fam.Download(context.Background(), m, "USB", noProg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(read, got) {
				t.Fatalf("lettura di dmrconfig diversa: %s", diffRanges(read, got))
			}
			_ = os.WriteFile(filepath.Join(dir, "mine.img"), mine, 0o644)
			eo := genEmuFor(m, base)
			addr, stop = serveRawTCP(func(p Port) { _ = eo.serve(p) })
			run([]string{"DMRCONFIG_HID=" + addr}, "-w", "mine.img")
			stop()
			cur = genEmuFor(m, base)
			if _, err := m.Fam.Upload(context.Background(), m, "USB", mine, noProg); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(eo.image(), cur.image()) {
				t.Fatalf("radio scritta da dmrconfig diversa: %s", diffRanges(eo.image(), cur.image()))
			}
			t.Logf("%d canali letti uguali a dmrconfig; %d modifiche: memoria identica a quella di dmrconfig; lettura e scrittura via USB uguali", n, len(edits))
		})
	}
}

// netPort: la radio simulata servita su un collegamento TCP senza intestazioni
// (per dmrconfig compilato con verifica/dmrconfig/tcpdev.c)
type netPort struct{ c net.Conn }

func (p *netPort) Write(b []byte) error { _, err := p.c.Write(b); return err }
func (p *netPort) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	_ = p.c.SetReadDeadline(time.Now().Add(to))
	_, err := io.ReadFull(p.c, b)
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return errTimeout
	}
	return err
}
func (p *netPort) Purge() error      { return nil }
func (p *netPort) SetBaud(int) error { return nil }
func (p *netPort) Close() error      { return p.c.Close() }

func serveRawTCP(serve func(p Port)) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		serve(&netPort{c})
	}()
	return ln.Addr().String(), func() { ln.Close(); <-done }
}
