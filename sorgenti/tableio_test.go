// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func loadIK2ANE(t *testing.T) *ImportResult {
	t.Helper()
	// elenco dei ponti di IK2ANE: non incluso nel pacchetto (dati di terzi);
	// per queste prove scaricalo da https://www.ik2ane.it/pontixls.xls in testdata
	b, err := os.ReadFile("testdata/pontixls.xls")
	if err != nil {
		t.Skip("testdata/pontixls.xls (elenco IK2ANE) non presente")
	}
	sh, err := readTable(b, "pontixls.xls")
	if err != nil {
		t.Fatal(err)
	}
	res, err := importSheets(sh, "", importOpts{Modello: "{LOC}"})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestImportIK2ANE(t *testing.T) {
	res := loadIK2ANE(t)
	if res.Sheet != "ponti" || res.Formato != "IK2ANE" {
		t.Fatalf("foglio/formato: %s %s", res.Sheet, res.Formato)
	}
	// 1594 righe dati: 1585 ponti + 4 segnaposto 9999999 + 4 righe di testo + 1 vuota
	if len(res.Ponti) != 1585 || len(res.Scartate) != 8 {
		t.Fatalf("ponti %d scartate %d", len(res.Ponti), len(res.Scartate))
	}
	find := func(sigla, loc string, rx int64) *Ponte {
		for _, p := range res.Ponti {
			if p.Sigla == sigla && strings.HasPrefix(p.Localita, loc) && p.Rx == rx {
				return p
			}
		}
		t.Fatalf("ponte %s %s non trovato", sigla, loc)
		return nil
	}
	p := find("RV", "Giarre", 145475000)
	if p.Tx != 144875000 || p.TxTone != "127.3" || p.Regione != "Sicilia" || p.Provincia != "CT" || p.Nominativo != "IR9BL" || p.Nome != "GIARRE" {
		t.Fatalf("Giarre: %+v", p)
	}
	p = find("RU12", "Sarre", 430300000) // "+5.0 MHz"
	if p.Tx != 435300000 || p.TxTone != "74.4" || p.Regione != "Valle d'Aosta" {
		t.Fatalf("Sarre: %+v", p)
	}
	p = find("RV", "Valcava", 144625000) // "+1.362,5 MHz"
	if p.Tx != 145987500 || len(p.Avvisi) == 0 {
		t.Fatalf("Valcava: %+v", p)
	}
	p = find("RU", "Lamporecchio", 430600000) // "DCS023"
	if p.TxTone != "" || len(p.Avvisi) == 0 {
		t.Fatalf("DCS: %+v", p)
	}
	p = find("RU8", "Verbania Suna", 430200000) // "CC1", DMR
	if p.TxTone != "" || len(p.Avvisi) != 0 || !strings.Contains(p.Note, "color code") || p.Tipi[0] != "DMR" {
		t.Fatalf("CC1: %+v", p)
	}
	// stato da colori
	n := map[string]int{}
	for _, p := range res.Ponti {
		n[p.Stato]++
	}
	if n[statoOff] != 9 || n[statoTemp] != 52 || n[statoH24] != 3 {
		t.Fatalf("stati: %v", n)
	}
}

func TestParseFields(t *testing.T) {
	sh := []struct {
		in  string
		off int64
		dub bool
	}{{"+5.0MHz", 5000000, false}, {"-600 kHz", -600000, false}, {"+1.6 MHz", 1600000, false}, {"0", 0, false}, {"0.0", 0, false},
		{"+1.362,5 MHz", 1362500, true}, {"-30MHz", -30000000, false}, {"input", 0, true}, {"+4.675 MHz", 4675000, false}, {"600", -600000, true}}
	for _, c := range sh {
		off, d := parseShift(c.in, 145600000)
		if off != c.off || (d != "") != c.dub {
			t.Errorf("shift %q -> %d %q", c.in, off, d)
		}
	}
	tones := []struct {
		in, out string
		dub     bool
	}{{"79,7", "79.7", false}, {"77.0-141.3", "77.0", true}, {"CC1", "", false}, {"", "", false}, {" ", "", false}, {"123", "123.0", false},
		{"82.2", "82.5", true}, {"DCS023", "", true}, {"1750", "", false}, {"110.9/162.2 loc", "110.9", true}, {"71.9 ", "71.9", false},
		{"162.2tx", "162.2", false}, {"107.2-dcss 266", "107.2", true}, {"DG-ID:10", "", false}, {"CC1-131.8", "131.8", false}, {"xyz", "", true}}
	for _, c := range tones {
		tn, d, _ := parseToneField(c.in)
		if tn != c.out || (len(d) > 0) != c.dub {
			t.Errorf("tono %q -> %q %v", c.in, tn, d)
		}
	}
	for in, want := range map[string]string{"0 lazio": "Lazio", "1piemonte": "Piemonte", "1 Piemonte": "Piemonte", "7 pugla": "Puglia", "6 marche ": "Marche",
		"3n trentino a.a.": "Trentino-Alto Adige", "1x v.aosta": "Valle d'Aosta", "7/8 basilicata": "Basilicata", "Slovenia": "Slovenia", "slovenia": "Slovenia", "4 emilia r.": "Emilia-Romagna"} {
		if g := normRegione(in); g != want {
			t.Errorf("regione %q -> %q", in, g)
		}
	}
	if g := parseTipi("EL-SVX", "R1"); strings.Join(g, ",") != "Analogico,EchoLink,SvxLink" {
		t.Errorf("tipi %v", g)
	}
	if g := parseTipi("C4FM+A", "RU"); strings.Join(g, ",") != "Analogico,C4FM" {
		t.Errorf("tipi %v", g)
	}
	if g := parseTipi("", "HotSpot"); strings.Join(g, ",") != "HotSpot" {
		t.Errorf("tipi %v", g)
	}
}

// Esportazione nei 4 formati e reimportazione: i dati devono tornare uguali.
func TestExportReimport(t *testing.T) {
	res := loadIK2ANE(t)
	pp := res.Ponti
	dir := t.TempDir()
	for _, f := range []string{"xlsx", "xls", "csv", "chirp"} {
		var b []byte
		var err error
		if f == "chirp" {
			b = pontiChirp(pp)
		} else {
			b, err = exportSheet(pontiSheet(pp), f)
		}
		if err != nil {
			t.Fatal(f, err)
		}
		name := filepath.Join(dir, "ponti_"+f+map[string]string{"xlsx": ".xlsx", "xls": ".xls", "csv": ".csv", "chirp": ".csv"}[f])
		_ = os.WriteFile(name, b, 0o644)
		if d := os.Getenv("EXPORT_DIR"); d != "" {
			_ = os.WriteFile(filepath.Join(d, filepath.Base(name)), b, 0o644)
		}
		sh, err := readTable(b, name)
		if err != nil {
			t.Fatal(f, err)
		}
		r2, err := importSheets(sh, "", importOpts{})
		if err != nil {
			t.Fatal(f, err)
		}
		if len(r2.Ponti) != len(pp) || len(r2.Scartate) != 0 {
			t.Fatalf("%s: %d ponti (attesi %d), scartate %d %v", f, len(r2.Ponti), len(pp), len(r2.Scartate), r2.Scartate[:min(3, len(r2.Scartate))])
		}
		for i, a := range pp {
			b := r2.Ponti[i]
			if a.Rx != b.Rx || a.Tx != b.Tx || a.TxTone != b.TxTone || a.RxTone != b.RxTone || a.Nome != b.Nome || a.Narrow != b.Narrow {
				t.Fatalf("%s riga %d:\n%+v\n%+v", f, i, a, b)
			}
			if f != "chirp" && (a.Localita != b.Localita || a.Regione != b.Regione || a.Provincia != b.Provincia || a.Stato != b.Stato ||
				strings.Join(a.Tipi, ",") != strings.Join(b.Tipi, ",") || a.Rete != b.Rete || a.Nominativo != b.Nominativo || a.Sigla != b.Sigla) {
				t.Fatalf("%s riga %d:\n%+v\n%+v", f, i, a, b)
			}
		}
		t.Logf("%s: %d byte, %d ponti riletti", f, len(b), len(r2.Ponti))
	}
}

func TestChirpTones(t *testing.T) {
	out := string(chirpCSV([]chirpMem{
		{loc: 1, name: "A", rx: 145600000, tx: 145000000, txTone: "67.0"},
		{loc: 2, name: "B", rx: 430100000, tx: 431700000, txTone: "88.5", rxTone: "88.5", narrow: true, power: 2},
		{loc: 3, name: "C", rx: 145600000, tx: 145600000, txTone: "67.0", rxTone: "71.9"},
		{loc: 4, name: "D,E", rx: 145600000, tx: 435000000, rxTone: "71.9", power: 1},
	}))
	want := []string{
		"1,A,145.600000,-,0.600000,Tone,67.0,88.5,023,NN,023,Tone->Tone,FM,12.50,,5.0W,",
		"2,B,430.100000,+,1.600000,TSQL,88.5,88.5,023,NN,023,Tone->Tone,NFM,12.50,,1.0W,",
		"3,C,145.600000,,0.000000,Cross,67.0,71.9,023,NN,023,Tone->Tone,FM,12.50,,5.0W,",
		"4,\"D,E\",145.600000,split,435.000000,Cross,71.9,71.9,023,NN,023,->Tone,FM,12.50,,2.0W,",
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("manca riga CHIRP:\n%s\nin\n%s", w, out)
		}
	}
	sh, _ := readTable([]byte(out), "x.csv")
	r, err := importSheets(sh, "", importOpts{})
	if err != nil || len(r.Ponti) != 4 {
		t.Fatal(err)
	}
	exp := []struct{ tx, txt, rxt string }{{"145000000", "67.0", ""}, {"431700000", "88.5", "88.5"}, {"145600000", "67.0", "71.9"}, {"435000000", "", "71.9"}}
	for i, e := range exp {
		p := r.Ponti[i]
		if itoa(p.Tx) != e.tx || p.TxTone != e.txt || p.RxTone != e.rxt {
			t.Errorf("CHIRP riga %d: %+v", i+1, p)
		}
	}
	if r.Ponti[1].Power != 2 || !r.Ponti[1].Narrow || r.Ponti[3].Power != 1 || r.Ponti[3].Nome != "D,E" {
		t.Errorf("CHIRP potenza/banda/nome: %+v %+v", r.Ponti[1], r.Ponti[3])
	}
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

func TestChannelExport(t *testing.T) {
	chs, _ := readDAT(mustRead(t, sampleDAT), &Report{})
	var sel []int
	for i := range chs {
		sel = append(sel, i)
	}
	for _, f := range []string{"xlsx", "xls", "csv"} {
		b, err := exportSheet(chanSheet(kindDAT, chs, sel), f)
		if err != nil {
			t.Fatal(err)
		}
		if d := os.Getenv("EXPORT_DIR"); d != "" {
			_ = os.WriteFile(filepath.Join(d, "bajeton_canali."+f), b, 0o644)
		}
		sh, err := readTable(b, "x."+f)
		if err != nil {
			t.Fatal(err)
		}
		r, err := importSheets(sh, "", importOpts{})
		if err != nil {
			t.Fatal(f, err)
		}
		if len(r.Ponti) != 152 {
			t.Fatalf("%s: %d", f, len(r.Ponti))
		}
		k := 0
		for _, c := range chs {
			if c == nil {
				continue
			}
			p := r.Ponti[k]
			k++
			if p.Rx != c.Rx || p.Tx != c.Tx || p.TxTone != c.TxTone || p.RxTone != c.RxTone || (c.Name != "" && p.Nome != truncName(strings.ToUpper(c.Name))) || p.Narrow != c.Narrow || p.Power != c.Power {
				t.Fatalf("%s: %+v vs %+v", f, c, p)
			}
		}
	}
	x := readXLC(mustRead(t, sampleXLC), &Report{})
	if !strings.Contains(string(chanChirp(kindXLC, x, []int{90})), "91,PROVA MONTE ALTO,145.637500,-,0.600000,Tone,67.0,88.5") {
		t.Error("CHIRP K67")
	}
}

func TestNomi(t *testing.T) {
	// località di fantasia (stessa forma dei nomi dell'elenco dei ponti)
	for loc, want := range map[string]string{"Alpe Fiorita - Borgo Nuovo (XX)": "ALPE FIORITA", "M.Alto (XX)": "M.ALTO", "Bric di Sopra - Valfiorita (XX)": "BRIC DI SOPRA",
		"Castelvecchio Prataverde (XX)": "CASTELVECCHIO PR", "Monte Prova - Paese": "MONTE PROVA", "Sant'Angelo in Valle (XX)": "SANT'ANGELO IN V"} {
		if g := nomeDaModello(&Ponte{Localita: loc}, "{LOC}"); g != want {
			t.Errorf("%q -> %q (atteso %q)", loc, g, want)
		}
	}
	if g := nomeDaModello(&Ponte{Localita: "M.Alto (XX)", Sigla: "RV", Nominativo: "IR0XX"}, "{NOMINATIVO} {LOC}"); g != "IR0XX M.ALTO" {
		t.Error(g)
	}
	if g := nomeDaModello(&Ponte{Sigla: "RU5", Rx: 430125000}, "{LOC}"); g != "RU5" {
		t.Error(g)
	}
}
