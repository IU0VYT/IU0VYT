//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove delle radio della serie D e delle BF-T17 / BF-T18 (drv_bfd.go):
//   - TestT17CPS: il codice ORIGINALE dei programmi «T17-T18 PMR446 PRO» e
//     «T17-T18 FRS» (verifica/t17/harness) legge e scrive la stessa radio
//     simulata: canali letti uguali, memoria finale identica a quella scritta qui.
//       T17_HARNESS=…/t17_harness.dll T17_EXE=…/T17_PMR446_PRO.exe T17FRS_EXE=…/CPS.exe
//   - TestBfdVendorJS: le funzioni di lettura e scrittura dei canali del
//     programma «BF Series CPS» (prese dal file dell'applicazione indicato in
//     BFD_JS, eseguite con node) danno gli stessi canali e gli stessi byte.
//   - TestBfdChirp: canali delle immagini di prova di CHIRP (W31D, BF-V12D) e
//     byte inviati alla radio uguali a quelli di CHIRP (CHIRP_DIR).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------- canali di prova

func pmrTestChans(m *genModel) []*Chan {
	pw := func(i int) int { return m.Powers[i].Sem }
	f := func(mhz float64) int64 { return int64(mhz*1e6 + 0.5) }
	return []*Chan{
		{Rx: f(446.00625), Tx: f(446.00625), Power: pw(1), Narrow: true, Scan: true},
		{Rx: f(446.01875), Tx: f(446.01875), TxTone: "88.5", RxTone: "88.5", Power: pw(0), Narrow: true, Scan: false},
		{Rx: f(446.03125), Tx: f(446.03125), TxTone: "D023N", RxTone: "D023N", Power: pw(1), Narrow: false, Scan: true},
		{Rx: f(446.04375), Tx: f(446.04375), TxTone: "D754I", RxTone: "67.0", Power: pw(0), Narrow: true, Scan: true},
		{Rx: f(433.500), Tx: f(438.100), TxTone: "123.0", Power: pw(1), Narrow: false, Scan: false},
		{Rx: f(446.09375), Tx: f(446.09375), RxTone: "254.1", Power: pw(1), Narrow: true, Scan: true},
	}
}

func pmrTestImage(t *testing.T, m *genModel, size int, settings map[int][]byte) []byte {
	img := bytes.Repeat([]byte{0xFF}, size)
	for a, d := range settings {
		copy(img[a:], d)
	}
	for i, c := range pmrTestChans(m) {
		if err := m.Fam.Set(m, img, i, c); err != nil {
			t.Fatal(err)
		}
	}
	return img
}

// ---------------------------------------------------------------- T17 / T18

func t17Str(c *Chan) string {
	if c == nil {
		return ""
	}
	tone := func(s string) string {
		if s == "" {
			return "OFF"
		}
		return s
	}
	f := func(hz int64) string { s := fmtMHz6(hz); return s[:len(s)-1] } // 446.00625
	pw, bw, sc := 0, 1, 0
	if c.Power == 2 {
		pw = 1
	}
	if c.Narrow {
		bw = 0
	}
	if c.Scan {
		sc = 1
	}
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d|%d", f(c.Rx), f(c.Tx), tone(c.RxTone), tone(c.TxTone), pw, bw, sc)
}

func TestT17CPS(t *testing.T) {
	h := os.Getenv("T17_HARNESS")
	if h == "" {
		t.Skip("T17_HARNESS non impostato")
	}
	for _, v := range []struct{ id, env string }{{"bf-t17pmr", "T17_EXE"}, {"bf-t17frs", "T17FRS_EXE"}, {"bf-gt18pmr", "GT18_EXE"}} {
		exe := os.Getenv(v.env)
		if exe == "" {
			continue
		}
		m := genByID[v.id]
		t.Run(v.id, func(t *testing.T) {
			b0 := byte(0x03)
			if v.id == "bf-t17frs" {
				b0 = 0x83
			}
			base := pmrTestImage(t, m, t17Memsize, map[int][]byte{
				0x630: {0x01, 0x00, 0x01, 0x00, 0x03, 0x00, 0x00, 0x00},
				0x640: {b0, 0x05, 0x00, 0x06, 0x01, 0xFF, 0xFF, 0xFF},
				0x650: {0x00, 0x00, 0xF7, 0x33, 0x31, 0x30, 0x33, 0x50},
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
			// lettura: stessi canali
			out := run(genEmuFor(m, base), "read")
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) != 99 {
				t.Fatalf("%d righe dal programma del costruttore", len(lines))
			}
			for i := 0; i < m.Slots; i++ {
				f := strings.Split(strings.TrimSpace(lines[i]), "|")
				got := ""
				if f[1] != "" {
					got = strings.Join(f[1:8], "|")
				}
				if want := t17Str(m.Fam.Get(m, base, i)); got != want {
					t.Errorf("canale %d: programma del costruttore %q, qui %q", i+1, got, want)
				}
			}
			// scrittura delle stesse modifiche: memoria finale identica
			chs := genChans(m, base)
			edits := []struct {
				n int
				c *Chan
			}{
				{1, nil},
				{2, &Chan{Rx: 446106250, Tx: 446106250, TxTone: "D162N", RxTone: "D162N", Power: 0, Narrow: true, Scan: false}},
				{7, &Chan{Rx: 446181250, Tx: 446181250, TxTone: "131.8", RxTone: "131.8", Power: 2, Narrow: true, Scan: true}},
				{m.Slots - 1, &Chan{Rx: 446193750, Tx: 446193750, Power: 2, Narrow: false, Scan: false}},
			}
			var sb strings.Builder
			for _, e := range edits {
				chs[e.n] = e.c
				if e.c == nil {
					fmt.Fprintf(&sb, "%d|\n", e.n)
				} else {
					fmt.Fprintf(&sb, "%d|%s|0|0\n", e.n, t17Str(e.c))
				}
			}
			ef := filepath.Join(t.TempDir(), "mod.txt")
			_ = os.WriteFile(ef, []byte(sb.String()), 0o644)
			eo := genEmuFor(m, base)
			run(eo, "write", ef)
			want, err := genWrite(m, base, chs, &Report{})
			if err != nil {
				t.Fatal(err)
			}
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
			if !bytes.Equal(eo.image(), eg.image()) {
				t.Fatalf("memoria scritta dal programma del costruttore diversa: %s", diffRanges(eo.image(), eg.image()))
			}
			t.Logf("%d canali letti uguali; %d modifiche scritte: memoria identica (%d byte)", m.Slots, len(edits), len(want))
		})
	}
}

// ---------------------------------------------------------------- BF-T9

// canale nel formato del programma «BF-Ts» (vedi verifica/t9/harness)
func t9Str(c *Chan) string {
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
	pw, bw, sc := 1, 1, 0
	if c.Power == 2 {
		pw = 0
	}
	if c.Narrow {
		bw = 0
	}
	if c.Scan {
		sc = 1
	}
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d|%d", f(c.Rx), f(c.Tx), tone(c.RxTone), tone(c.TxTone), pw, bw, sc)
}

// il programma «BF-Ts» rilegge sempre «beat shift» attivo (errore nel controllo
// del bit 1) e quindi in scrittura azzera il bit 1 di tutti i canali: il
// confronto lo ignora; un canale cancellato dal programma conserva toni e
// impostazioni (con le frequenze a FF), qui diventa tutto FF: nel confronto
// conta solo che sia vuoto
func t9Mask(img []byte) []byte {
	out := append([]byte{}, img...)
	for i := 0; i < 99; i++ {
		r := out[16*i : 16*i+16]
		if bytes.Equal(r[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
			copy(r, bytes.Repeat([]byte{0xFF}, 16))
		}
		r[12] |= 0x02
	}
	return out
}

func TestT9CPS(t *testing.T) {
	h, exe := os.Getenv("T9_HARNESS"), os.Getenv("T9_EXE")
	if h == "" || exe == "" {
		t.Skip("T9_HARNESS o T9_EXE non impostati")
	}
	m := genByID["bf-t9"]
	run := func(e genEmu, args ...string) string {
		addr, stop := serveHarnessFn(func(p Port, baud int) { _ = e.serve(p) })
		defer stop()
		out, err := exec.Command("dotnet", append([]string{h, exe, addr}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("programma del costruttore: %v\n%s", err, out)
		}
		return string(out)
	}
	img := bytes.Repeat([]byte{0xFF}, t17Memsize)
	copy(img[0x630:], []byte{0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x00, 0x00})
	copy(img[0x650:], []byte{0x00, 0x00, 0xF7, 0x37, 0x30, 0x31, 0x33, 0x50})
	chans := []*Chan{
		{Rx: 446006250, Tx: 446006250, Power: 0, Scan: true},
		{Rx: 433500000, Tx: 433500000, TxTone: "88.5", RxTone: "88.5", Power: 2, Narrow: true, Scan: true},
		nil,
		{Rx: 430225000, Tx: 431825000, TxTone: "D023N", Power: 0, Scan: false},
		{Rx: 462562500, Tx: 467562500, TxTone: "123.0", RxTone: "D754I", Power: 2, Narrow: true, Scan: true},
	}
	for i, c := range chans {
		if err := m.Fam.Set(m, img, i, c); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Fam.Set(m, img, 98, &Chan{Rx: 469975000, Tx: 469975000, Scan: true}); err != nil {
		t.Fatal(err)
	}
	// fuori dalla gamma memorizzata (400-470 MHz): rifiutato
	if err := m.Fam.Set(m, img, 10, &Chan{Rx: 145500000, Tx: 145500000}); err == nil {
		t.Fatal("145.5 MHz accettato su una radio 400-470 MHz")
	}
	// immagine «normalizzata» dal programma: lettura e scrittura senza modifiche
	e0 := genEmuFor(m, img)
	vuoto := filepath.Join(t.TempDir(), "vuoto.txt")
	_ = os.WriteFile(vuoto, nil, 0o644)
	run(e0, "write", vuoto)
	base := e0.image()
	// canali identici (le impostazioni in 0x630-0x64F le ricalcola il programma)
	if a, b := t9Mask(base)[:0x630], t9Mask(img)[:0x630]; !bytes.Equal(a, b) {
		t.Fatalf("il programma del costruttore riscrive i canali in modo diverso: %s", diffRanges(a, b))
	}
	out := run(genEmuFor(m, base), "read")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 99 {
		t.Fatalf("%d righe dal programma del costruttore", len(lines))
	}
	for i := 0; i < m.Slots; i++ {
		f := strings.Split(strings.TrimSpace(lines[i]), "|")
		got := ""
		if f[1] != "" {
			got = strings.Join(f[1:8], "|")
		}
		if want := t9Str(m.Fam.Get(m, base, i)); got != want {
			t.Errorf("canale %d: programma del costruttore %q, qui %q", i+1, got, want)
		}
	}
	chs := genChans(m, base)
	edits := []struct {
		n int
		c *Chan
	}{
		{0, nil},
		{2, &Chan{Rx: 446106250, Tx: 446106250, TxTone: "D162N", RxTone: "D162N", Power: 0, Narrow: true, Scan: false}},
		{7, &Chan{Rx: 438625000, Tx: 431025000, TxTone: "131.8", Power: 2, Narrow: false, Scan: true}},
		{m.Slots - 2, &Chan{Rx: 400012500, Tx: 400012500, Power: 2, Narrow: false, Scan: false}},
	}
	var sb strings.Builder
	for _, e := range edits {
		chs[e.n] = e.c
		if e.c == nil {
			fmt.Fprintf(&sb, "%d|\n", e.n)
		} else {
			fmt.Fprintf(&sb, "%d|%s|0|1\n", e.n, t9Str(e.c))
		}
	}
	ef := filepath.Join(t.TempDir(), "mod.txt")
	_ = os.WriteFile(ef, []byte(sb.String()), 0o644)
	eo := genEmuFor(m, base)
	run(eo, "write", ef)
	want, err := genWrite(m, base, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
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
	if !bytes.Equal(t9Mask(eo.image()), t9Mask(eg.image())) {
		t.Fatalf("memoria scritta dal programma del costruttore diversa: %s", diffRanges(t9Mask(eo.image()), t9Mask(eg.image())))
	}
	// e la lettura qui di quanto ha scritto il programma del costruttore
	got, _, err := m.Fam.Download(context.Background(), m, "COM9", noProg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range chs {
		if a, b := t9Str(m.Fam.Get(m, got, i)), t9Str(chs[i]); a != b {
			t.Errorf("canale %d: letto %q, atteso %q", i+1, a, b)
		}
	}
	t.Logf("%d canali letti uguali; %d modifiche scritte: memoria identica (%d byte)", m.Slots, len(edits), len(want))
}

// ---------------------------------------------------------------- serie D: funzioni del programma

// estrae dal file dell'applicazione il testo di una funzione «nome = (…) => {…}» o
// «function nome(…) {…}»
func jsFunc(t *testing.T, src, name string) string {
	re := regexp.MustCompile(`(?:function ` + name + `\(|[,\s]` + name + ` = \()`)
	loc := re.FindStringIndex(src)
	if loc == nil {
		t.Fatalf("funzione %s non trovata", name)
	}
	start := loc[0] + 1
	if strings.HasPrefix(src[loc[0]:], "function") {
		start = loc[0]
	}
	i := strings.IndexByte(src[start:], '{') + start
	depth := 0
	for k := i; k < len(src); k++ {
		switch src[k] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				body := src[start : k+1]
				if !strings.HasPrefix(body, "function") {
					body = "var " + body + ";"
				}
				return body + "\n"
			}
		}
	}
	t.Fatalf("funzione %s incompleta", name)
	return ""
}

func TestBfdVendorJS(t *testing.T) {
	p := os.Getenv("BFD_JS")
	if p == "" {
		t.Skip("BFD_JS non impostato")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node non disponibile")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	var js strings.Builder
	js.WriteString(`var localStorage={getItem:function(){return null}};
var Cn=["low","high"],_n=["wide","narrow"],On=["have","notHave"],Rn=["have","notHave"],cn=["close","pulseOn"],er=["analog","digital"];
`)
	for _, n := range []string{"Kr", "Tt", "du", "U2", "a1", "Ro", "xf", "q2", "Y2", "Gb"} {
		js.WriteString(jsFunc(t, src, n))
	}
	for _, id := range []string{"bf-v1d", "bf-1901d", "bf-w31d"} {
		m := genByID[id]
		t.Run(id, func(t *testing.T) {
			img := pmrTestImage(t, m, bfdMemsize, map[int][]byte{
				0x2B0: {0x01, 0x01, 0x01, 0x00, 0x03, 0x00, 0x01, 0x01},
				0x2B8: {0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				0x3C0: {0x07, 0x03, 0x01, 0x04, 0xFF, 0xFF, 0xFF, 0xFF},
			})
			// un canale digitale
			if err := m.Fam.Set(m, img, 6, &Chan{Rx: 446106250, Tx: 446106250, Power: 0, Narrow: true, Scan: true, Dig: true}); err != nil {
				t.Fatal(err)
			}
			blocks := func(a, b int) []string {
				var out []string
				for x := a; x < b; x += 8 {
					out = append(out, fmt.Sprintf("%x", img[x:x+8]))
				}
				return out
			}
			end := 0x110
			if m.Slots == 30 {
				end = 0x1F0
			}
			in, _ := json.Marshal([][]string{blocks(0x10, end), blocks(0x2B0, 0x2C0), blocks(0x3C0, 0x3D0)})
			code := js.String() + `var a=` + string(in) + `;var r=Y2(a[0],a[1],a[2],[],[],"A");` +
				`var g=Gb(r.list,r.formData,"` + map[bool]string{true: "D", false: "A"}[m.Slots == 30] + `");` +
				`console.log(JSON.stringify({list:r.list,blocks:g}));`
			f := filepath.Join(t.TempDir(), "bfd.js")
			_ = os.WriteFile(f, []byte(code), 0o644)
			out, err := exec.Command("node", f).CombinedOutput()
			if err != nil {
				t.Fatalf("node: %v\n%s", err, out)
			}
			var res struct {
				List   [][]interface{} `json:"list"`
				Blocks []string        `json:"blocks"`
			}
			if err := json.Unmarshal(out, &res); err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			if len(res.List) != m.Slots {
				t.Fatalf("%d canali dal programma", len(res.List))
			}
			tone := func(s string) string {
				if s == "" {
					return "OFF"
				}
				return s
			}
			for i, row := range res.List {
				c := m.Fam.Get(m, img, i)
				if c == nil {
					if row[1] != "" {
						t.Errorf("canale %d vuoto qui, %v per il programma", i+1, row)
					}
					continue
				}
				pw := "low"
				if m.powerIndex(c.Power) == 1 {
					pw = "high"
				}
				bw := map[bool]string{true: "narrow", false: "wide"}[c.Narrow]
				sc := map[bool]string{true: "have", false: "notHave"}[c.Scan]
				dg := map[bool]string{true: "digital", false: "analog"}[c.Dig]
				f := func(hz int64) string { s := fmtMHz6(hz); return s[:len(s)-1] }
				want := []string{f(c.Rx), f(c.Tx), tone(c.RxTone), tone(c.TxTone), pw, bw, sc, dg}
				got := []string{fmt.Sprint(row[1]), fmt.Sprint(row[2]), fmt.Sprint(row[3]), fmt.Sprint(row[4]),
					fmt.Sprint(row[5]), fmt.Sprint(row[6]), fmt.Sprint(row[7]), fmt.Sprint(row[10])}
				if c.Dig {
					got[2], got[3] = "OFF", "OFF"
				}
				if strings.Join(got, "|") != strings.Join(want, "|") {
					t.Errorf("canale %d: programma %v, qui %v", i+1, got, want)
				}
			}
			// byte scritti dal programma per gli stessi canali (byte 14-15 a parte: il programma scrive FF)
			for i := 0; i < m.Slots; i++ {
				g := res.Blocks[2+2*i] + res.Blocks[3+2*i]
				q := fmt.Sprintf("%x", img[0x10+16*i:0x10+16*i+14])
				if !strings.HasPrefix(g, q) && !(m.Fam.Get(m, img, i) == nil && strings.HasPrefix(g, strings.Repeat("f", 28))) {
					t.Errorf("canale %d: programma %s, qui %s", i+1, g[:28], q)
				}
			}
			t.Logf("%d canali: lettura e codifica uguali al programma del costruttore", m.Slots)
		})
	}
}

// ---------------------------------------------------------------- serie D: confronto con CHIRP

func TestBfdChirp(t *testing.T) {
	d := chirpDir(t)
	for _, v := range []struct{ id, img string }{{"bf-w31d", "Baofeng_W31D.img"}, {"bf-v12d", "Baofeng_BF-V12D.img"}} {
		m := genByID[v.id]
		t.Run(v.id, func(t *testing.T) {
			p := filepath.Join(d, "tests", "images", v.img)
			img := loadImg(t, p)
			ref := chirpDecode(t, m, p)
			n := 0
			for i := 0; i < m.Slots; i++ {
				g := m.Fam.Get(m, img, i)
				var c *chirpChan
				if i < len(ref) {
					c = ref[i]
				}
				if g != nil && g.Dig {
					continue // CHIRP: modo FM/NFM senza toni
				}
				if g != nil && c != nil {
					// la scansione è letta al contrario da CHIRP
					if g.Scan == c.Scan {
						t.Errorf("canale %d: scansione %v anche per CHIRP (atteso il contrario)", i+1, g.Scan)
					}
					c.Scan = g.Scan
				}
				if s := cmpChan(m, i+1, g, c); s != "" {
					t.Error(s)
				}
				n++
			}
			// protocollo: byte inviati alla radio uguali a CHIRP
			eC, eG := genEmuFor(m, img), genEmuFor(m, img)
			out := filepath.Join(t.TempDir(), "dl.img")
			recC := chirpSession(t, m, eC, "download", out)
			var got []byte
			recG := goSession(t, m, eG, func() error {
				var err error
				got, _, err = m.Fam.Download(context.Background(), m, "EMU", noProg)
				return err
			})
			if !bytes.Equal(recC, recG) {
				t.Errorf("lettura: byte diversi da CHIRP: %s", diffRanges(recC, recG))
			}
			dl, _ := splitImg(mustRead(t, out))
			if !bytes.Equal(dl, got) {
				t.Errorf("lettura: immagine diversa da CHIRP: %s", diffRanges(dl, got))
			}
			mod := append([]byte{}, got...)
			if err := m.Fam.Set(m, mod, 2, &Chan{Rx: 446056250, Tx: 446056250, TxTone: "D023N", RxTone: "D023N", Power: 0, Narrow: true, Scan: true}); err != nil {
				t.Fatal(err)
			}
			inF := filepath.Join(t.TempDir(), "ul.img")
			_ = os.WriteFile(inF, makeImg(m, mod, nil), 0o644)
			uC, uG := genEmuFor(m, img), genEmuFor(m, img)
			recC = chirpSession(t, m, uC, "upload", inF)
			recG = goSession(t, m, uG, func() error {
				_, err := m.Fam.Upload(context.Background(), m, "EMU", mod, noProg)
				return err
			})
			if !bytes.Equal(recC, recG) {
				t.Errorf("scrittura: byte diversi da CHIRP: %s", diffRanges(recC, recG))
			}
			if !bytes.Equal(uC.image(), uG.image()) {
				t.Errorf("scrittura: memoria finale diversa: %s", diffRanges(uC.image(), uG.image()))
			}
			t.Logf("%d canali uguali a CHIRP (scansione a parte); lettura e scrittura: byte identici a CHIRP", n)
		})
	}
}
