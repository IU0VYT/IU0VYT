// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"os"
	"testing"
)

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

const sampleXLC = "testdata/k67.xlc"
const sampleDAT = "testdata/bajeton.dat"

func TestNRBFRoundTrip(t *testing.T) {
	for _, b := range [][]byte{baseDAT, mustRead(t, sampleDAT)} {
		d, err := ParseNRBF(b)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(d.Dump(), b) {
			t.Fatal("round-trip NRBF diverso")
		}
	}
}

func TestShort(t *testing.T) {
	for in, want := range map[string]string{"CAMPOLUNGHINO 3": "CAMPOLUNGH 3", "VALDIFIORITA 2": "VALDIFIORI 2", "SANTA FIORENZA": "S.FIORENZA", "PROVA MONTE ALTO": "PROVA M.ALTO", "PMR CH 1": "PMR CH 1", "LPD CH 45": "LPD CH 45"} {
		if g := shortName(in, 12); g != want || len(g) > 12 {
			t.Errorf("%q -> %q (atteso %q)", in, g, want)
		}
	}
}

func TestBCD(t *testing.T) {
	b := make([]byte, 4)
	putBCDLE(b, 44600625)
	v, _ := bcdLE(b)
	if v != 44600625 || b[0] != 0x25 || b[3] != 0x44 {
		t.Fatal("bcd LE", b)
	}
	c := make([]byte, 2)
	putBCDBE(c, 1567)
	w, _ := bcdBE(c)
	if w != 1567 || c[0] != 0x15 {
		t.Fatal("bcd BE", c)
	}
}

func count(chs []*Chan) int {
	n := 0
	for _, c := range chs {
		if c != nil {
			n++
		}
	}
	return n
}

// Aprire e salvare senza modifiche deve dare lo stesso identico file.
func TestIdentityXLC(t *testing.T) {
	src := mustRead(t, sampleXLC)
	rep := &Report{}
	chs := readXLC(src, rep)
	if n := count(chs); n != 153 {
		t.Fatalf("canali K67 letti: %d (attesi 153)", n)
	}
	out := writeXLC(src, chs, &Report{})
	if !bytes.Equal(out, src) {
		for i := range out {
			if out[i] != src[i] {
				t.Fatalf("K67: primo byte diverso a 0x%X", i)
			}
		}
	}
}

func TestIdentityDAT(t *testing.T) {
	src := mustRead(t, sampleDAT)
	chs, err := readDAT(src, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	if n := count(chs); n != 152 {
		t.Fatalf("canali Bajeton letti: %d (attesi 152)", n)
	}
	out, err := writeDAT(src, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, src) {
		t.Fatalf("Bajeton: file diverso dopo apri/salva (%d vs %d byte)", len(out), len(src))
	}
}

// Spostare un canale porta con sé i byte che il programma originale conserva,
// aggiorna le tabelle dei canali presenti, della scansione e della zona.
func TestMoveKeepsRawXLC(t *testing.T) {
	src := mustRead(t, sampleXLC)
	chs := readXLC(src, &Report{})
	// canale 86 (PMR DI PROVA) ha il byte 17 = 01
	moved := chs[85]
	chs[85] = nil
	chs[299] = moved
	out := writeXLC(src, chs, &Report{})
	r := out[xlcChBase+xlcChSize*299 : xlcChBase+xlcChSize*300]
	o := src[xlcChBase+xlcChSize*85 : xlcChBase+xlcChSize*86]
	if !bytes.Equal(r, o) {
		t.Fatalf("record spostato diverso:\n%x\n%x", r, o)
	}
	if !bytes.Equal(out[xlcChBase+xlcChSize*85:xlcChBase+xlcChSize*86], make([]byte, 48)) {
		t.Fatal("posizione originale non svuotata")
	}
	if bitGet(out, xlcValid, 299) != 0 || bitGet(out, xlcValid, 85) != 1 {
		t.Fatal("tabella dei canali presenti non aggiornata")
	}
	if bitGet(out, xlcScanFlg, 299) != bitGet(src, xlcScanFlg, 85) {
		t.Fatal("scansione non spostata")
	}
	// zona unica (radio senza zone): voce 299 = 299, voce 85 = FFFF, contatore invariato
	z := out[xlcZoneBase:]
	if int(z[2+2*299])<<8|int(z[3+2*299]) != 299 || z[2+2*85] != 0xFF || z[3+2*85] != 0xFF || z[0] != src[xlcZoneBase] {
		t.Fatalf("zona non aggiornata: %x %x n=%d", z[2+2*299:4+2*299], z[2+2*85:4+2*85], z[0])
	}
	// nient'altro deve cambiare fuori da canali, tabelle e zona
	for i := range out {
		inCh := i >= xlcChBase && i < xlcChBase+xlcChSize*xlcChannels
		inTab := (i >= xlcValid && i < xlcValid+80) || (i >= xlcScanFlg && i < xlcScanFlg+80) || (i >= xlcZoneBase && i < xlcZoneBase+1304)
		if !inCh && !inTab && out[i] != src[i] {
			t.Fatalf("byte 0x%X cambiato", i)
		}
	}
	back := readXLC(out, &Report{})
	if back[299] == nil || back[299].Name != "PMR DI PROVA" || back[85] != nil {
		t.Fatal("rilettura errata")
	}
}

// Il file di esempio è già "normalizzato" dal programma originale: apri e
// salva senza modifiche non deve cambiare nulla (verificato in TestIdentityXLC);
// qui si controlla che cancellare e ricreare tutti i canali lo riporti identico
// tranne che per i contatori della zona.
func TestDeleteRecreateXLC(t *testing.T) {
	src := mustRead(t, sampleXLC)
	chs := readXLC(src, &Report{})
	blank := writeXLC(src, nil, &Report{})
	if count(readXLC(blank, &Report{})) != 0 {
		t.Fatal("canali rimasti")
	}
	again := writeXLC(blank, chs, &Report{})
	for i := range src {
		inZone := i >= xlcZoneBase && i < xlcZoneBase+1304
		inEmerg := i >= 40200 && i < 40200+160 // allarmi: ChSel impostato quando la zona resta vuota (UpDate_EmergData)
		if again[i] != src[i] && !inZone && !inEmerg && i != 0x7A80 {
			t.Fatalf("byte 0x%X diverso dopo cancella/ricrea: %02x -> %02x", i, src[i], again[i])
		}
	}
	// zona: ogni canale ricreato ha la propria voce (tranne 67-74, coperte dal nome della zona come nel programma originale)
	z := again[xlcZoneBase:]
	for j, c := range chs {
		v := int(z[2+2*j])<<8 | int(z[3+2*j])
		switch {
		case j >= 67 && j <= 74:
		case c != nil && v != j:
			t.Fatalf("voce di zona %d = %d", j, v)
		case c == nil && v != 0xFFFF:
			t.Fatalf("voce di zona %d (vuota) = %d", j, v)
		}
	}
	// il contatore della zona è a 8 bit nel file, come nel programma originale: dopo
	// 153 cancellazioni e 153 inserimenti torna al valore iniziale
	if string(z[136:142]) != "Zone 1" || z[0] != src[xlcZoneBase] {
		t.Fatalf("nome/contatori zona: %q n=%d tot=%d", z[136:142], z[0], again[0x7A80])
	}
}

func TestToneK67(t *testing.T) {
	for _, c := range []struct {
		h, l byte
		s    string
	}{{0x08, 0x85, "88.5"}, {0x10, 0x00, "100.0"}, {0x25, 0x41, "254.1"}, {0, 0, ""}, {0x80, 0x23, "D023N"}, {0xC0, 0x23, "D023I"}, {0x87, 0x54, "D754N"}} {
		got, ok := xlcToneDecode(c.h, c.l)
		if !ok || got != c.s {
			t.Errorf("decode %02x%02x -> %q", c.h, c.l, got)
		}
		b, ok := xlcToneEncode(c.s)
		if !ok || b[0] != c.h || b[1] != c.l {
			t.Errorf("encode %q -> %x", c.s, b)
		}
	}
	if _, ok := xlcToneDecode(0x0A, 0x85); ok {
		t.Error("BCD non valido accettato")
	}
}

func TestEditXLC(t *testing.T) {
	src := mustRead(t, sampleXLC)
	chs := readXLC(src, &Report{})
	c := chs[90] // PROVA MONTE ALTO
	c.Name = "PROVA"
	c.TxTone = "88.5"
	c.RxTone = "88.5"
	c.Power = 2
	c.Narrow = true
	c.Rx = 145500000
	c.Tx = 145500000
	out := writeXLC(src, chs, &Report{})
	b := readXLC(out, &Report{})[90]
	if b.Name != "PROVA" || b.TxTone != "88.5" || b.RxTone != "88.5" || b.Power != 2 || !b.Narrow || b.Rx != 145500000 || b.Tx != 145500000 {
		t.Fatalf("modifica non riletta: %+v", b)
	}
	r := out[xlcChBase+xlcChSize*90:]
	if r[16]&0x0C != 0 {
		t.Fatal("flag shift non azzerato su simplex")
	}
}

func TestEditDAT(t *testing.T) {
	src := mustRead(t, sampleDAT)
	chs, _ := readDAT(src, &Report{})
	// sposta zona 2 n.22 in zona 10 n.5, modifica un altro
	chs[9*64+4] = chs[64+21]
	chs[64+21] = nil
	chs[0].Name = "NUOVO NOME"
	chs[0].Power = 1
	chs[0].Scan = false
	out, err := writeDAT(src, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := readDAT(out, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	if b[64+21] != nil || b[9*64+4] == nil || b[9*64+4].Name != "PMR DI PROVA" || b[9*64+4].TxTone != "141.3" || !b[9*64+4].Narrow {
		t.Fatalf("spostamento errato: %+v", b[9*64+4])
	}
	if b[0].Name != "NUOVO NOME" || b[0].Power != 1 || b[0].Scan {
		t.Fatalf("modifica errata: %+v", b[0])
	}
	if count(b) != 152 {
		t.Fatal("numero canali cambiato")
	}
}

func sameChan(a, b *Chan) bool {
	return a.Rx == b.Rx && a.Tx == b.Tx && a.TxTone == b.TxTone && a.RxTone == b.RxTone && a.Narrow == b.Narrow
}

// Andata e ritorno K67 -> Bajeton -> K67: frequenze, toni, banda e potenza restano.
func TestRoundTripFormats(t *testing.T) {
	src := mustRead(t, sampleXLC)
	orig := readXLC(src, &Report{})
	dat, kind, rep, err := Convert(src, nil)
	if err != nil || kind != kindDAT {
		t.Fatal(err, kind)
	}
	t.Logf("K67->Bajeton: %d avvisi", len(rep.Warnings))
	mid, err := readDAT(dat, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	back, kind, _, err := Convert(dat, src)
	if err != nil || kind != kindXLC {
		t.Fatal(err)
	}
	fin := readXLC(back, &Report{})
	for i := range orig {
		a, m, f := orig[i], mid[i], fin[i]
		if (a == nil) != (m == nil) || (a == nil) != (f == nil) {
			t.Fatalf("pos %d: presenza diversa", i+1)
		}
		if a == nil {
			continue
		}
		if !sameChan(a, m) || !sameChan(a, f) || a.Power != m.Power || a.Power != f.Power {
			t.Fatalf("pos %d: %+v / %+v / %+v", i+1, a, m, f)
		}
		if len(m.Name) > 12 || f.Name != m.Name {
			t.Fatalf("pos %d: nomi %q %q %q", i+1, a.Name, m.Name, f.Name)
		}
	}
	// Bajeton -> K67 -> Bajeton
	src2 := mustRead(t, sampleDAT)
	o2, _ := readDAT(src2, &Report{})
	x, _, _, err := Convert(src2, nil)
	if err != nil {
		t.Fatal(err)
	}
	d2, _, _, err := Convert(x, src2)
	if err != nil {
		t.Fatal(err)
	}
	f2, _ := readDAT(d2, &Report{})
	for i := range o2 {
		a, f := o2[i], f2[i]
		if (a == nil) != (f == nil) {
			t.Fatalf("Bajeton pos %d: presenza diversa", i)
		}
		if a != nil && (!sameChan(a, f) || a.Name != f.Name || (a.Power == 2) != (f.Power == 2)) {
			t.Fatalf("Bajeton pos %d: %+v / %+v", i, a, f)
		}
	}
}

func TestBlank(t *testing.T) {
	x := blankXLC(baseXLC)
	if count(readXLC(x, &Report{})) != 0 {
		t.Fatal("K67 vuoto con canali")
	}
	d, err := blankDAT(baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	chs, err := readDAT(d, &Report{})
	if err != nil || count(chs) != 0 {
		t.Fatal("Bajeton vuoto con canali", err)
	}
}

func TestAdaptWarnings(t *testing.T) {
	rep := &Report{}
	c := &Chan{Rx: 145600000, Tx: 145000000, TxTone: "XYZ", RxTone: "D023N", Power: 1, Name: "VERY LONG NAME FOR TEST", Raw: "k67:xx"}
	a := adaptChan(c, kindXLC, xlcBands(baseXLC, 0x10), "test", rep)
	if a.Power != 0 || a.TxTone != "" || a.RxTone != "D023N" || len(a.Name) != 16 || a.Raw != "" || len(rep.Warnings) != 3 {
		t.Fatalf("%+v %v", a, rep.Warnings)
	}
	rep = &Report{}
	b := adaptChan(&Chan{Rx: 430100000, Tx: 431700000, TxTone: "67", RxTone: "D017N", Name: "MONTE ALTO RIPETITORE"}, kindDAT, nil, "t", rep)
	if b.TxTone != "67.0" || b.RxTone != "" || len(b.Name) > 12 || len(rep.Warnings) != 2 {
		t.Fatalf("%+v %v", b, rep.Warnings)
	}
}
