//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Confronto dei driver delle radio con CHIRP (vedi verifica/chirp/chirpref.py).
// Si eseguono solo se CHIRP_DIR indica una copia dei sorgenti di CHIRP:
//   CHIRP_DIR=/percorso/chirp go test -mod=vendor -run TestChirp -v .
// Per ogni modello:
//   - mappa: la posizione di ogni campo descritto qui è la stessa di CHIRP;
//   - lettura: i canali letti da un'immagine sono gli stessi che legge CHIRP;
//   - scrittura: la stessa serie di modifiche (canali nuovi, modificati,
//     cancellati, con toni CTCSS/DCS di ogni tipo, potenze, banda stretta,
//     scansione, nomi, TX disabilitato, split) produce un'immagine identica
//     byte per byte a quella prodotta da CHIRP.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func chirpDir(t *testing.T) string {
	d := os.Getenv("CHIRP_DIR")
	if d == "" {
		t.Skip("CHIRP_DIR non impostata")
	}
	return d
}

func chirpRef(t *testing.T, args ...string) []byte {
	t.Helper()
	d := chirpDir(t)
	script, _ := filepath.Abs("verifica/chirp/chirpref.py")
	cmd := exec.Command("python3", append([]string{"-I", script}, args...)...)
	cmd.Env = append(os.Environ(), "CHIRP_DIR="+d)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("chirpref %v: %v\n%s\n%s", args, err, out, stderr.String())
	}
	return out
}

// immagine di prova per il modello: quella di CHIRP del modello, o della famiglia
var chirpTestImage = map[string]string{
	"bf-v8a":  "Radtel_T18.img",
	"bf-g11s": "Retevis_RB27.img",
}

func chirpImageFor(t *testing.T, m *genModel) string {
	d := chirpDir(t)
	cands := []string{chirpTestImage[m.ID], strings.ReplaceAll(m.Vendor+"_"+m.Model, " ", "_") + ".img"}
	if f, ok := chirpFamilyImage[fmt.Sprintf("%T", m.Fam)]; ok {
		cands = append(cands, f)
	}
	for _, c := range cands {
		if c == "" {
			continue
		}
		p := filepath.Join(d, "tests", "images", c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skipf("nessuna immagine di prova per %s", m.Name)
	return ""
}

// immagine di riferimento per i modelli della famiglia senza un'immagine propria
var chirpFamilyImage = map[string]string{
	"main.uv5rFam":  "Baofeng_UV-5R.img",
	"main.uv17Fam":  "Baofeng_UV-17Pro.img",
	"main.bfcFam":   "Baofeng_UV-9G.img",
	"main.h777Fam":  "Baofeng_BF-888.img",
	"main.uv17lFam": "Baofeng_UV-17.img",
	"main.gm30Fam":  "Radioddity_GM-30.img",
}

type chirpChan struct {
	Rx     int64  `json:"rx"`
	Tx     int64  `json:"tx"`
	TxOff  bool   `json:"txOff"`
	TxTone string `json:"txTone"`
	RxTone string `json:"rxTone"`
	Power  int    `json:"power"`
	Mode   string `json:"mode"`
	Scan   bool   `json:"scan"`
	Name   string `json:"name"`
	Dup    string `json:"dup,omitempty"`
}

func loadImg(t *testing.T, p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	mem, _ := splitImg(b)
	return mem
}

func cmpChan(m *genModel, n int, g *Chan, c *chirpChan) string {
	if (g == nil) != (c == nil) {
		return fmt.Sprintf("canale %d: vuoto qui=%v CHIRP=%v (%+v / %+v)", n, g == nil, c == nil, g, c)
	}
	if g == nil {
		return ""
	}
	var diffs []string
	if g.Rx != c.Rx {
		diffs = append(diffs, fmt.Sprintf("rx %d/%d", g.Rx, c.Rx))
	}
	if g.TxOff != c.TxOff {
		diffs = append(diffs, fmt.Sprintf("txoff %v/%v", g.TxOff, c.TxOff))
	}
	if !c.TxOff && g.Tx != c.Tx {
		diffs = append(diffs, fmt.Sprintf("tx %d/%d", g.Tx, c.Tx))
	}
	if g.TxTone != c.TxTone || g.RxTone != c.RxTone {
		diffs = append(diffs, fmt.Sprintf("toni %q,%q / %q,%q", g.TxTone, g.RxTone, c.TxTone, c.RxTone))
	}
	if c.Power >= 0 && m.powerIndex(g.Power) != c.Power {
		diffs = append(diffs, fmt.Sprintf("potenza %d/%d", m.powerIndex(g.Power), c.Power))
	}
	if c.Mode == "FM" || c.Mode == "NFM" {
		if g.Narrow != (c.Mode == "NFM") {
			diffs = append(diffs, fmt.Sprintf("banda %v/%s", g.Narrow, c.Mode))
		}
	}
	if m.HasScan && g.Scan != c.Scan {
		diffs = append(diffs, fmt.Sprintf("scan %v/%v", g.Scan, c.Scan))
	}
	if g.Name != c.Name {
		diffs = append(diffs, fmt.Sprintf("nome %q/%q", g.Name, c.Name))
	}
	if len(diffs) > 0 {
		return fmt.Sprintf("canale %d: %s", n, strings.Join(diffs, ", "))
	}
	return ""
}

func chirpDecode(t *testing.T, m *genModel, img string) []*chirpChan {
	var res []*chirpChan
	if err := json.Unmarshal(chirpRef(t, "decode", m.RClass, img), &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func TestChirpDecode(t *testing.T) {
	chirpDir(t)
	for _, m := range genModels {
		if m.NoChirp {
			continue
		}
		m := m
		t.Run(m.ID, func(t *testing.T) {
			p := chirpImageFor(t, m)
			img := loadImg(t, p)
			ref := chirpDecode(t, m, p)
			if len(ref) != m.Slots {
				t.Fatalf("CHIRP ha %d canali, qui %d", len(ref), m.Slots)
			}
			bad := 0
			used := 0
			for i := 0; i < m.Slots; i++ {
				g := genGet(m, img, i)
				if g != nil {
					used++
				}
				if d := cmpChan(m, i+m.First, g, ref[i]); d != "" {
					bad++
					if bad < 15 {
						t.Error(d)
					}
				}
			}
			t.Logf("%s: %d canali letti, %d differenze", filepath.Base(p), used, bad)
		})
	}
}

// serie di modifiche di prova per un modello
func chirpEdits(m *genModel, img []byte) []struct {
	N    int
	Chan *Chan
} {
	bands := m.Bands
	if b, ok := m.Fam.(interface {
		Bands(*genModel, []byte) []band
	}); ok {
		bands = b.Bands(m, img)
	}
	// una frequenza nella prima gamma "radioamatoriale" disponibile
	pick := func(pref []int64) int64 {
		for _, f := range pref {
			for _, b := range bands {
				if f >= b.Lo && f <= b.Hi {
					return f
				}
			}
		}
		return bands[0].Lo + 1000000
	}
	vhf := pick([]int64{145600000, 146520000, 155000000, 162400000})
	uhf := pick([]int64{430700000, 433500000, 446006250, 462562500, 145750000})
	dcs1 := fmt.Sprintf("D%03d", m.DCS[0])
	dcs2 := fmt.Sprintf("D%03d", m.DCS[len(m.DCS)/2])
	np := len(m.Powers)
	pw := func(k int) int { return m.Powers[k%np].Sem }
	name := func(s string) string { return m.cleanName(s) }
	var free, used []int
	for i := 0; i < m.Slots; i++ {
		if lk, ok := m.Fam.(interface{ lockedSlot(*genModel, int) bool }); ok && lk.lockedSlot(m, i) {
			continue
		}
		if m.Fam.Get(m, img, i) == nil {
			free = append(free, i)
		} else {
			used = append(used, i)
		}
	}
	chs := []*Chan{
		{Rx: vhf, Tx: vhf, Power: pw(0), Scan: true, Name: name("SIMPLEX")},
		{Rx: vhf, Tx: vhf - 600000, TxTone: "88.5", Power: pw(1), Scan: true, Name: name("RPT TONE")},
		{Rx: uhf, Tx: uhf + 5000000, TxTone: "123.0", RxTone: "123.0", Power: pw(2), Narrow: m.Narrow, Scan: false, Name: name("TSQL")},
		{Rx: uhf, Tx: uhf, TxTone: dcs1 + "N", RxTone: dcs1 + "N", Power: pw(0), Scan: true, Name: name("DCS NN")},
		{Rx: vhf, Tx: vhf, TxTone: dcs2 + "I", RxTone: dcs2 + "N", Power: pw(1), Scan: true, Name: name("DCS IN")},
		{Rx: vhf, Tx: vhf, TxTone: "100.0", RxTone: dcs1 + "I", Power: pw(0), Scan: true, Name: name("T-D")},
		{Rx: uhf, Tx: uhf, TxTone: dcs2 + "N", RxTone: "67.0", Power: pw(0), Scan: true, Name: name("D-T")},
		{Rx: vhf, Tx: vhf, RxTone: "254.1", Power: pw(0), Scan: true, Name: name("RXONLY")},
		{Rx: uhf, Tx: uhf, RxTone: dcs2 + "I", Power: pw(2), Scan: true, Name: ""},
		{Rx: vhf, Tx: vhf, TxTone: "77.0", RxTone: "203.5", Power: pw(0), Narrow: m.Narrow, Scan: true, Name: name("XTONE")},
		{Rx: vhf, Tx: vhf, TxOff: true, Power: pw(0), Scan: true, Name: name("NOTX")},
		{Rx: vhf, Tx: uhf, Power: pw(0), Scan: true, Name: name("SPLIT")},
		{Rx: uhf, Tx: uhf, Power: pw(np - 1), Narrow: m.Narrow, Scan: false, Name: name("ABCDEFGHIJKLMNOPQRST")},
		{Rx: vhf + 12500, Tx: vhf + 12500, Power: pw(0), Scan: true, Name: name("a-b.c/1")},
	}
	if m.NoTxOff {
		for k, c := range chs {
			if c.TxOff {
				chs = append(chs[:k], chs[k+1:]...)
				break
			}
		}
	}
	var out []struct {
		N    int
		Chan *Chan
	}
	k := 0
	for _, c := range chs {
		var slot int
		if k%2 == 0 && len(free) > 0 {
			slot, free = free[0], free[1:]
		} else if len(used) > 0 {
			slot, used = used[0], used[1:]
		} else if len(free) > 0 {
			slot, free = free[0], free[1:]
		} else {
			break
		}
		k++
		out = append(out, struct {
			N    int
			Chan *Chan
		}{slot, c})
	}
	// cancellazione di un canale presente, poi un canale nuovo nella stessa
	// posizione e la cancellazione di un altro
	if len(used) > 1 {
		out = append(out, struct {
			N    int
			Chan *Chan
		}{used[len(used)-1], nil})
		out = append(out, struct {
			N    int
			Chan *Chan
		}{used[len(used)-1], &Chan{Rx: vhf, Tx: vhf, Power: pw(0), Scan: true, Name: name("NEW")}})
		out = append(out, struct {
			N    int
			Chan *Chan
		}{used[len(used)-2], nil})
	}
	// ultima posizione
	out = append(out, struct {
		N    int
		Chan *Chan
	}{m.Slots - 1, &Chan{Rx: uhf, Tx: uhf - 1600000, TxTone: "151.4", Power: pw(0), Scan: true, Name: name("LAST")}})
	// posizioni in sola ricezione (es. GM-15 Pro): stesso canale senza TX
	if ro, ok := m.Fam.(interface{ rxOnlySlot(*genModel, int) bool }); ok {
		for k := range out {
			if c := out[k].Chan; c != nil && ro.rxOnlySlot(m, out[k].N) {
				cc := *c
				cc.TxOff, cc.Tx = true, cc.Rx
				out[k].Chan = &cc
			}
		}
	}
	return out
}

func TestChirpEncode(t *testing.T) {
	chirpDir(t)
	tmp := t.TempDir()
	for _, m := range genModels {
		if m.NoChirp {
			continue
		}
		m := m
		t.Run(m.ID, func(t *testing.T) {
			p := chirpImageFor(t, m)
			img := loadImg(t, p)
			edits := chirpEdits(m, img)
			var js []map[string]interface{}
			out := append([]byte{}, img...)
			for _, e := range edits {
				var cj interface{}
				if e.Chan != nil {
					c := *e.Chan
					mode := "FM"
					if c.Narrow {
						mode = "NFM"
					}
					cj = map[string]interface{}{"rx": c.Rx, "tx": c.Tx, "txOff": c.TxOff, "txTone": c.TxTone, "rxTone": c.RxTone,
						"power": m.powerIndex(c.Power), "mode": mode, "scan": c.Scan, "name": c.Name}
				}
				js = append(js, map[string]interface{}{"n": e.N + m.First, "chan": cj})
				if err := m.Fam.Set(m, out, e.N, e.Chan); err != nil {
					t.Fatalf("canale %d: %v", e.N, err)
				}
				// rilettura
				g := m.Fam.Get(m, out, e.N)
				if e.Chan == nil {
					if g != nil {
						t.Errorf("canale %d cancellato ma ancora presente", e.N)
					}
					continue
				}
				if g == nil {
					t.Errorf("canale %d scritto ma vuoto", e.N)
					continue
				}
				w := *e.Chan
				w.Raw, g.Raw = "", ""
				if w.TxOff {
					w.Tx = w.Rx
				}
				if !m.HasScan {
					w.Scan = g.Scan
				}
				if !m.Narrow {
					w.Narrow = g.Narrow
				}
				if *g != w {
					t.Errorf("canale %d riletto diverso:\n  scritto %+v\n  riletto %+v", e.N, w, *g)
				}
			}
			ej, _ := json.Marshal(js)
			ef := filepath.Join(tmp, m.ID+"_edits.json")
			of := filepath.Join(tmp, m.ID+"_out.img")
			_ = os.WriteFile(ef, ej, 0o644)
			res := chirpRef(t, "encode", m.RClass, p, ef, of)
			var rr struct {
				Errors [][]interface{} `json:"errors"`
			}
			_ = json.Unmarshal(res, &rr)
			for _, e := range rr.Errors {
				t.Logf("CHIRP segnala per il canale %v: %v", e[0], e[1])
			}
			ref, err := os.ReadFile(of)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ref, out) {
				t.Errorf("immagine diversa da quella di CHIRP: %s", diffRanges(ref, out))
			} else {
				t.Logf("%d modifiche: immagine identica a quella di CHIRP (%d byte)", len(edits), len(out))
			}
		})
	}
}

// la posizione dei campi della mappa coincide con quella di CHIRP
func TestChirpLayout(t *testing.T) {
	chirpDir(t)
	done := map[string]bool{}
	for _, m := range genModels {
		if m.NoChirp {
			continue
		}
		lf, ok := m.Fam.(interface{ layout(m *genModel) *bwEl })
		if !ok {
			continue
		}
		root := lf.layout(m)
		key := fmt.Sprintf("%p", root)
		if done[key] {
			continue
		}
		done[key] = true
		m := m
		t.Run(m.ID, func(t *testing.T) {
			p := chirpImageFor(t, m)
			var ref map[string][]interface{}
			if err := json.Unmarshal(chirpRef(t, "layout", m.RClass, p), &ref); err != nil {
				t.Fatal(err)
			}
			n, bad := 0, 0
			root.Leaves("", func(path string, x *bwEl) {
				// array di tipi semplici: CHIRP li riporta come un unico elemento
				base := path
				if i := strings.LastIndexByte(path, '['); i > 0 && strings.HasSuffix(path, "]") && x.Bits == 0 {
					if _, ok := ref[path]; !ok {
						base = path[:i]
					}
				}
				r, ok := ref[base]
				if !ok {
					bad++
					if bad < 10 {
						t.Errorf("campo %s non presente in CHIRP", path)
					}
					return
				}
				n++
				off := int(r[0].(float64))
				nb := int(r[2].(float64))
				sh := int(r[3].(float64))
				if base != path {
					// elemento di un array semplice
					if x.Off < off || x.Off >= off+int(r[1].(float64)) {
						bad++
						t.Errorf("%s: indirizzo 0x%X fuori dall'array CHIRP 0x%X", path, x.Off, off)
					}
					return
				}
				if x.Off != off || x.Bits != nb || (nb > 0 && x.Shift != sh) {
					bad++
					if bad < 10 {
						t.Errorf("%s: qui 0x%X bit %d/%d, CHIRP 0x%X bit %d/%d", path, x.Off, x.Bits, x.Shift, off, nb, sh)
					}
				}
			})
			t.Logf("%d campi confrontati, %d differenze", n, bad)
		})
	}
}
