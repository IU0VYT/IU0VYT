//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// GT-68 e GT-22 confrontate con il codice ORIGINALE dei programmi del
// costruttore («BF_V1_FH» del pacchetto GT-68 e «GT-22»), che ha la stessa
// classe AnaCps.Data.DataProtocol della T17: si usa lo stesso programma di
// prova (verifica/t17/harness).
//   T17_HARNESS=… GT68_EXE=…/BF_V1_FH.exe GT22_EXE=…/GT-22.exe

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// canale come lo mostra il programma: RX|TX|toneRX|toneTX|PowerMode|BandWide|ScanAdd
// (bit della memoria: alta potenza, banda stretta, bit 4)
func h777CpsStr(m *genModel, img []byte, i int) string {
	c := m.Fam.Get(m, img, i)
	if c == nil {
		return ""
	}
	tone := func(s string) string {
		if s == "" {
			return "OFF"
		}
		return s
	}
	f := func(hz int64) string { s := fmtMHz6(hz); return s[:len(s)-1] }
	r := (h777Fam{}).rec(m, i)
	b := func(n string) int { return int(r.F(n).Int(img)) }
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d|%d", f(c.Rx), f(c.Tx), tone(c.RxTone), tone(c.TxTone), b("highpower"), b("narrow"), b("skip"))
}

func TestGT2268CPS(t *testing.T) {
	h := os.Getenv("T17_HARNESS")
	if h == "" {
		t.Skip("T17_HARNESS non impostato")
	}
	for _, v := range []struct{ id, env string }{{"bf-gt68", "GT68_EXE"}, {"bf-gt22", "GT22_EXE"}} {
		exe := os.Getenv(v.env)
		if exe == "" {
			continue
		}
		m := genByID[v.id]
		t.Run(v.id, func(t *testing.T) {
			base := pmrTestImage(t, m, h777o(m).memsize, map[int][]byte{
				0x0000: bytes.Repeat([]byte{0xFF}, 16),
				0x0130: {0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				0x02B0: {0x01, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				0x03C0: {0x03, 0x05, 0x00, 0x06, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				0x03D0: {0x00, 0x00, 0xF7, 0x37, 0x30, 0x31, 0x33, 0x50},
			})
			run := func(e genEmu, args ...string) string {
				addr, stop := serveHarnessFn(func(p Port, baud int) { _ = e.serve(p) })
				defer stop()
				out, err := exec.Command("dotnet", append([]string{h, exe, addr}, args...)...).CombinedOutput()
				if err != nil {
					t.Fatalf("programma del costruttore: %v\n%s", err, out)
				}
				return string(out)
			}
			out := run(genEmuFor(m, base), "read")
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) != m.Slots {
				t.Fatalf("%d righe dal programma del costruttore", len(lines))
			}
			for i := 0; i < m.Slots; i++ {
				f := strings.Split(strings.TrimSpace(lines[i]), "|")
				got := ""
				if f[1] != "" {
					got = strings.Join(f[1:8], "|")
				}
				if want := h777CpsStr(m, base, i); got != want {
					t.Errorf("canale %d: programma del costruttore %q, qui %q", i+1, got, want)
				}
			}
			// modifiche scritte dal programma e da qui
			chs := genChans(m, base)
			edits := []struct {
				n int
				c *Chan
			}{
				{1, nil},
				{2, &Chan{Rx: 446106250, Tx: 446106250, TxTone: "D162N", RxTone: "D162N", Power: 0, Narrow: true, Scan: true}},
				{7, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "131.8", RxTone: "", Power: 2, Narrow: true, Scan: true}},
				{15, &Chan{Rx: 462562500, Tx: 462562500, Power: 0, Narrow: false, Scan: true}},
			}
			var sb strings.Builder
			for _, e := range edits {
				chs[e.n] = e.c
				if e.c == nil {
					fmt.Fprintf(&sb, "%d|\n", e.n)
				}
			}
			want, err := genWrite(m, base, chs, &Report{})
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range edits {
				if e.c != nil {
					r := (h777Fam{}).rec(m, e.n)
					fmt.Fprintf(&sb, "%d|%s|%d|%d\n", e.n, h777CpsStr(m, want, e.n), r.F("bcl").Int(want), (want[r.Off+12]>>7)&1)
				}
			}
			ef := filepath.Join(t.TempDir(), "mod.txt")
			_ = os.WriteFile(ef, []byte(sb.String()), 0o644)
			eo := genEmuFor(m, base)
			run(eo, "write", ef)
			eg := genEmuFor(m, base)
			old := openPortFunc
			openPortFunc = func(string, int) (Port, error) {
				a, b := newPipe()
				go eg.serve(b)
				return a, nil
			}
			defer func() { openPortFunc = old }()
			if _, err := m.Fam.Upload(context.Background(), m, "COM9", want, noProg); err != nil {
				t.Fatal(err)
			}
			// Il programma riscrive tutti i 16 canali: nei canali che qui non si
			// toccano azzera i bit 1, 5 e 6 del byte delle opzioni e nelle posizioni
			// vuote (frequenza FF) scrive anche le opzioni. Per il resto (canali
			// modificati, cancellati, impostazioni) la memoria deve essere identica.
			theirs, mine := eo.image(), eg.image()
			mod := map[int]bool{}
			for _, e := range edits {
				mod[e.n] = true
			}
			norm := 0
			for i := 0; i < m.Slots; i++ {
				r := (h777Fam{}).rec(m, i)
				o := r.Off
				empty := bytes.Equal(mine[o:o+4], []byte{0xFF, 0xFF, 0xFF, 0xFF})
				for k := 8; k < 16; k++ {
					a, b := theirs[o+k], mine[o+k]
					if a == b {
						continue
					}
					if empty || !mod[i] && k == 12 && a&^0x62 == b&^0x62 {
						theirs[o+k] = b
						norm++
					}
				}
			}
			if !bytes.Equal(theirs, mine) {
				t.Fatalf("memoria scritta dal programma del costruttore diversa: %s", diffRanges(theirs, mine))
			}
			t.Logf("%d canali letti uguali; %d modifiche scritte: memoria identica (%d byte normalizzati dal programma nei canali non toccati o vuoti)", m.Slots, len(edits), norm)
		})
	}
}
