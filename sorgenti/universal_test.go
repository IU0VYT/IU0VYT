//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Convertitore universale (3.7): ogni file di partenza verso ogni formato di
// destinazione (tutte le radio), e il CSV di CHIRP confrontato con CHIRP.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// canali di prova con tutti i casi: toni CTCSS e DCS in ogni combinazione, shift,
// bande diverse, sola ricezione, larghezza, potenze, scansione
func uniSample() []*Chan {
	return []*Chan{
		{Rx: 145600000, Tx: 145000000, TxTone: "67.0", Name: "R0", Scan: true},
		{Rx: 430100000, Tx: 431700000, TxTone: "88.5", RxTone: "88.5", Narrow: true, Power: 2, Name: "TSQL"},
		{Rx: 145500000, Tx: 145500000, TxTone: "D023N", RxTone: "D023N", Name: "DCS", Scan: true},
		{Rx: 433500000, Tx: 433500000, TxTone: "D754I", RxTone: "D754N", Name: "DCS POL", Scan: true},
		{Rx: 145550000, Tx: 145550000, TxTone: "D023N", RxTone: "D031N", Name: "DCS X", Scan: true},
		{Rx: 145575000, Tx: 145575000, TxTone: "71.9", RxTone: "D047N", Name: "T>D", Power: 1, Scan: true},
		{Rx: 145587500, Tx: 145587500, TxTone: "D065N", RxTone: "77.0", Name: "D>T", Scan: true},
		{Rx: 145600000, Tx: 145600000, RxTone: "79.7", Name: "RXONLY T", Scan: true},
		{Rx: 145612500, Tx: 145612500, TxTone: "88.5", RxTone: "100.0", Name: "T>T", Scan: true},
		{Rx: 446006250, Tx: 446006250, Narrow: true, Name: "PMR1", Scan: true},
		{Rx: 145800000, Tx: 435000000, Name: "SPLIT", Scan: true},
		{Rx: 162550000, Tx: 162550000, TxOff: true, Name: "SOLO RX", Scan: true},
	}
}

func uniSame(a, b *Chan) string {
	switch {
	case a.Rx != b.Rx || (!a.TxOff && a.Tx != b.Tx):
		return fmt.Sprintf("frequenze %d/%d e %d/%d", a.Rx, a.Tx, b.Rx, b.Tx)
	case a.TxTone != b.TxTone || a.RxTone != b.RxTone:
		return fmt.Sprintf("toni %q/%q e %q/%q", a.TxTone, a.RxTone, b.TxTone, b.RxTone)
	case a.Narrow != b.Narrow || a.Scan != b.Scan || a.TxOff != b.TxOff || a.Dig != b.Dig || a.Power != b.Power:
		return fmt.Sprintf("attributi %+v e %+v", *a, *b)
	case a.Name != b.Name:
		return fmt.Sprintf("nome %q e %q", a.Name, b.Name)
	}
	return ""
}

// formati universali: andata e ritorno senza perdite
func TestUniversalFormats(t *testing.T) {
	src := uniSample()
	for _, tg := range []string{"chirp", "xlsx", "csv"} {
		out, ext, _, err := uniConvert(src, tg, nil, false, nil)
		if err != nil {
			t.Fatal(err)
		}
		back, err := readAny(out, "prova."+ext, "")
		if err != nil {
			t.Fatalf("%s: %v", tg, err)
		}
		got := packed(back.Channels)
		if len(got) != len(src) {
			t.Fatalf("%s: %d canali invece di %d", tg, len(got), len(src))
		}
		for i := range src {
			if d := uniSame(src[i], got[i]); d != "" {
				t.Errorf("%s canale %d: %s", tg, i+1, d)
			}
		}
	}
}

// ogni sorgente verso ogni radio: il file prodotto si rilegge, i canali sono nelle
// gamme e con le regole della radio, e ogni canale scritto viene da quelli di partenza
func TestUniversalAllTargets(t *testing.T) {
	sources := map[string][]*Chan{"esempio": uniSample()}
	for name, p := range map[string]string{"K67": sampleXLC, "BJ-8300": sampleDAT} {
		s, err := readAny(mustRead(t, p), filepath.Base(p), "")
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = s.Channels
	}
	n, skipped := 0, 0
	for sname, src := range sources {
		rx := map[int64]bool{}
		for _, c := range src {
			if c != nil {
				rx[c.Rx] = true
			}
		}
		for _, tg := range uniTargets() {
			var base []byte
			if tg.NeedsBase {
				if base = uniBase(t, genByID[tg.ID]); base == nil {
					skipped++
					continue
				}
			}
			out, ext, rep, err := uniConvert(src, tg.ID, base, false, nil)
			if err != nil {
				t.Fatalf("%s -> %s: %v", sname, tg.ID, err)
			}
			back, err := readAny(out, "uscita."+ext, tg.ID)
			if err != nil {
				t.Fatalf("%s -> %s: file non rileggibile: %v", sname, tg.ID, err)
			}
			if tg.NeedsBase && back.Kind != tg.ID {
				if m, b := genByID[tg.ID], genByID[back.Kind]; b == nil || b.RClass != m.RClass {
					t.Fatalf("%s -> %s: riletto come %s", sname, tg.ID, back.Kind)
				}
			}
			var baseChs []*Chan
			if base != nil {
				if bs, err := readAny(base, "base.img", tg.ID); err == nil {
					baseChs = bs.Channels
				}
			}
			k := 0
			for i, c := range back.Channels {
				if c == nil {
					continue
				}
				k++
				// i canali fissi della radio (posizioni bloccate) restano quelli della base
				if i < len(baseChs) && baseChs[i] != nil && baseChs[i].Rx == c.Rx {
					continue
				}
				if !rx[c.Rx] {
					t.Fatalf("%s -> %s: canale %+v non presente nella sorgente", sname, tg.ID, *c)
				}
			}
			if k == 0 && len(rep.Warnings) == 0 {
				t.Fatalf("%s -> %s: nessun canale e nessuna segnalazione", sname, tg.ID)
			}
			n++
		}
	}
	t.Logf("%d conversioni (%d sorgenti x %d formati; %d senza immagine di base, serve CHIRP_DIR): file rileggibili, canali dalla sorgente",
		n, len(sources), len(uniTargets()), skipped/len(sources))
}

// uniBase: immagine di base della radio (sintetica per i modelli non in CHIRP,
// altrimenti l'immagine di prova di CHIRP, se c'è CHIRP_DIR)
func uniBase(t *testing.T, m *genModel) []byte {
	if m.NoChirp {
		return makeImg(m, noChirpBase(t, m), nil)
	}
	d := os.Getenv("CHIRP_DIR")
	if d == "" {
		return nil
	}
	cands := []string{chirpTestImage[m.ID], strings.ReplaceAll(m.Vendor+"_"+m.Model, " ", "_") + ".img"}
	if f, ok := chirpFamilyImage[fmt.Sprintf("%T", m.Fam)]; ok {
		cands = append(cands, f)
	}
	for _, c := range cands {
		if c == "" {
			continue
		}
		if b, err := os.ReadFile(filepath.Join(d, "tests", "images", c)); err == nil {
			mem, _ := splitImg(b)
			if len(mem) >= genMinSize(m) {
				return makeImg(m, mem, nil)
			}
		}
	}
	return nil
}

// immagine di base di un'altra radio: rifiutata
func TestUniversalWrongBase(t *testing.T) {
	m := genByID["bf-uv25plus"]
	other := genByID["bf-t9"]
	if _, _, _, err := uniConvert(uniSample(), m.ID, makeImg(other, noChirpBase(t, other), nil), false, nil); err == nil {
		t.Fatal("base di un'altra radio accettata")
	}
	if _, _, _, err := uniConvert(uniSample(), m.ID, nil, false, nil); err == nil || !strings.Contains(err.Error(), "serve un'immagine") {
		t.Fatalf("senza base: %v", err)
	}
}

// CSV di CHIRP: scritto qui e letto da CHIRP, scritto da CHIRP e letto qui
func TestUniversalChirpCSV(t *testing.T) {
	dir := os.Getenv("CHIRP_DIR")
	if dir == "" {
		t.Skip("CHIRP_DIR non impostato")
	}
	py := filepath.Join("verifica", "chirp", "csvcheck.py")
	run := func(args ...string) []byte {
		cmd := exec.Command("python3", append([]string{py}, args...)...)
		cmd.Env = append(os.Environ(), "CHIRP_DIR="+dir)
		out, err := cmd.Output()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				t.Fatalf("csvcheck %v: %v\n%s", args, err, ee.Stderr)
			}
			t.Fatal(err)
		}
		return out
	}
	type cm struct {
		Name              string
		Rx, Tx            int64
		TxTone, RxTone    string
		Narrow, Dig, Scan bool
		TxOff             bool `json:"txOff"`
	}
	tmp := t.TempDir()
	// 1) scritto qui, letto da CHIRP
	src := uniSample()
	p := filepath.Join(tmp, "qui.csv")
	_ = os.WriteFile(p, writeChirpCSV(src, nil), 0o644)
	var got []cm
	if err := json.Unmarshal(run("parse", p), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(src) {
		t.Fatalf("CHIRP ha letto %d canali su %d", len(got), len(src))
	}
	for i, c := range src {
		g := &Chan{Name: got[i].Name, Rx: got[i].Rx, Tx: got[i].Tx, TxTone: got[i].TxTone, RxTone: got[i].RxTone, Narrow: got[i].Narrow,
			Dig: got[i].Dig, Scan: got[i].Scan, TxOff: got[i].TxOff, Power: c.Power}
		if d := uniSame(c, g); d != "" {
			t.Errorf("CHIRP, canale %d: %s", i+1, d)
		}
	}
	// 2) scritto da CHIRP (immagini di prova di CHIRP), letto qui
	n := 0
	for cls, img := range map[string]string{"BaofengUV5RGeneric": "Baofeng_UV-5R.img", "BFA58": "Baofeng_UV-9G.img",
		"UV17Pro": "Baofeng_UV-17Pro.img", "BJ7800": "Bajeton_BJ7800.img"} {
		if img == "" {
			continue
		}
		ip := filepath.Join(dir, "tests", "images", img)
		if _, err := os.Stat(ip); err != nil {
			continue
		}
		op := filepath.Join(tmp, cls+".csv")
		run("export", cls, ip, op)
		var want []cm
		if err := json.Unmarshal(run("parse", op), &want); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(op)
		mine, _, err := readChirpCSV(b)
		if err != nil {
			t.Fatalf("%s: %v", cls, err)
		}
		if len(mine) != len(want) {
			t.Fatalf("%s: %d canali qui, %d in CHIRP", cls, len(mine), len(want))
		}
		for i, w := range want {
			c := mine[i]
			wc := &Chan{Name: w.Name, Rx: w.Rx, Tx: w.Tx, TxTone: w.TxTone, RxTone: w.RxTone, Narrow: w.Narrow, Dig: w.Dig, Scan: w.Scan, TxOff: w.TxOff, Power: c.Power}
			if d := uniSame(wc, c); d != "" {
				t.Errorf("%s canale %d: %s", cls, i+1, d)
			}
		}
		n += len(mine)
	}
	t.Logf("%d canali: CSV scritto qui letto da CHIRP uguale; %d canali di CSV scritti da CHIRP letti qui uguali", len(src), n)
}

// immagini .img create dal convertitore per le radio di CHIRP: CHIRP le apre e
// legge gli stessi canali letti qui
func TestUniversalChirpOpens(t *testing.T) {
	if os.Getenv("CHIRP_DIR") == "" {
		t.Skip("CHIRP_DIR non impostato")
	}
	src := uniSample()
	n, models := 0, 0
	for _, m := range genModels {
		if m.NoChirp {
			continue
		}
		base := uniBase(t, m)
		if base == nil {
			continue
		}
		out, _, _, err := uniConvert(src, m.ID, base, false, nil)
		if err != nil {
			t.Fatalf("%s: %v", m.ID, err)
		}
		p := filepath.Join(t.TempDir(), m.ID+".img")
		_ = os.WriteFile(p, out, 0o644)
		mine := genChans(m, loadImg(t, p))
		ref := chirpDecode(t, m, p)
		for i := 0; i < m.Slots && i < len(ref); i++ {
			if d := cmpChan(m, i+m.First, mine[i], ref[i]); d != "" {
				t.Errorf("%s: %s", m.ID, d)
			}
			if mine[i] != nil {
				n++
			}
		}
		models++
	}
	t.Logf("%d modelli, %d canali: le immagini create dal convertitore si aprono con CHIRP con gli stessi canali", models, n)
}
