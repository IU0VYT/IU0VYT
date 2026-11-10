//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// differenza voluta rispetto a CHIRP: il codice 4 del campo «duplex» (e
// l'automatico fuori dalle gamme dei ponti) è il simplex della radio, che CHIRP
// mostra come «off»; qui il canale è in simplex con la trasmissione attiva
func (ft4Fam) chirpRefFix(m *genModel, g *Chan, c *chirpChan) {
	if c.TxOff && !g.TxOff && g.Tx == g.Rx {
		c.TxOff, c.Tx = false, c.Rx
	}
}

func ft4TestImg(t *testing.T, m *genModel) []byte {
	img := make([]byte, ft4Size)
	copy(img, "4123")
	for i, c := range []*Chan{
		{Rx: 145500000, Tx: 145500000, Name: "S20", Scan: true},
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Scan: true, Power: 1},
		{Rx: 145787500, Tx: 145187500, TxTone: "D023N", RxTone: "D023N", Name: "R7A", Narrow: true, Power: 2},
		{Rx: 145000000, Tx: 145012500, Name: "ODD", Scan: true},
	} {
		if err := m.Fam.Set(m, img, i*3, c); err != nil {
			t.Fatal(err)
		}
		g := m.Fam.Get(m, img, i*3)
		if g == nil || *g != *c {
			t.Fatalf("canale %d riletto diverso: %+v / %+v", i, g, c)
		}
	}
	return img
}

func TestFT4Record(t *testing.T) {
	for _, id := range []string{"ys-ft4xe", "ys-ft65e", "ys-ft65r", "ys-ft65r-s3", "ys-ft25r"} {
		m := genByID[id]
		img := ft4TestImg(t, m)
		f := ft4Fam{}
		// scostamento di 600 kHz in unità della radio; quello di 12,5 kHz come frequenza separata
		if got, want := f.rec(3).F("offset").Int(img), 600000/ft4o(m).unit(); got != want || f.rec(3).F("duplex").Int(img) != 2 {
			t.Errorf("%s: scostamento %d (atteso %d), verso %d", id, got, want, f.rec(3).F("duplex").Int(img))
		}
		if f.rec(9).F("duplex").Int(img) != 6 || f.mem().F("txfreqs").I(9).F("freq").Int(img) != 14501250 {
			t.Errorf("%s: scostamento non multiplo non scritto come frequenza separata", id)
		}
		// il simplex scritto dalla radio (codice 4) resta simplex con la trasmissione attiva
		f.rec(0).F("duplex").SetInt(img, 4)
		if g := m.Fam.Get(m, img, 0); g == nil || g.TxOff || g.Tx != g.Rx {
			t.Errorf("%s: simplex della radio letto come %+v", id, g)
		}
		// trasmissione disabilitata: solo dove CHIRP la ottiene con lo scostamento fuori gamma
		c := &Chan{Rx: 433500000, Tx: 433500000, TxOff: true, Scan: true, Name: "RXONLY"}
		if m.NoTxOff || len(m.Bands) < 3 {
			continue
		}
		if err := m.Fam.Set(m, img, 20, c); err != nil {
			t.Fatal(err)
		}
		if g := m.Fam.Get(m, img, 20); g == nil || !g.TxOff {
			t.Errorf("%s: trasmissione disabilitata riletta come %+v", id, g)
		}
		// spostamento del canale: record, nome, frequenza di trasmissione e bit
		raw := m.Fam.Rec(m, img, 3)
		m.Fam.PutRec(m, img, 50, raw)
		a, b := m.Fam.Get(m, img, 3), m.Fam.Get(m, img, 50)
		if a == nil || b == nil || *a != *b {
			t.Errorf("%s: canale spostato diverso: %+v / %+v", id, a, b)
		}
	}
	m := genByID["ys-ft4xe"]
	img := ft4TestImg(t, m)
	if err := m.Fam.Set(m, img, 1, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "D023N"}); err == nil {
		t.Error("subtono e DCS misti accettati")
	}
	if err := m.Fam.Set(m, img, 1, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "D023I"}); err == nil {
		t.Error("DCS invertito accettato")
	}
	if err := m.Fam.Set(m, img, 0, nil); err != nil || m.Fam.Get(m, img, 0) != nil {
		t.Error("canale non cancellato")
	}
}

func ft4Run(t *testing.T, m *genModel, e *ft4Emu, f func(port string)) {
	t.Helper()
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return a, nil
	}
	f("COM9")
}

func TestFT4Cable(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	m := genByID["ys-ft4xe"]
	img := ft4TestImg(t, m)
	e := ft4Fam{}.emu(m, img).(*ft4Emu)
	e.badSum = 2 // due letture disturbate: si ripetono
	ft4Run(t, m, e, func(port string) {
		got, meta, _, err := ft4Fam{}.DownloadMeta(ctx, m, port, nop)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, img) {
			t.Error("immagine letta diversa dalla memoria della radio")
		}
		if meta["ft4_subtype"] != 0 {
			t.Errorf("sottotipo %v", meta["ft4_subtype"])
		}
		// scrittura: tutto tranne il blocco 0
		mod := append([]byte{}, got...)
		copy(mod, "XXXX")
		m.Fam.Set(m, mod, 5, &Chan{Rx: 433500000, Tx: 433500000, Scan: true, Name: "U280"})
		if _, err := m.Fam.Upload(ctx, m, port, mod, nop); err != nil {
			t.Fatal(err)
		}
		fin := e.image()
		if !bytes.Equal(fin[16:], mod[16:]) || !bytes.Equal(fin[:16], img[:16]) {
			t.Error("memoria finale diversa: il blocco 0 non si scrive, il resto sì")
		}
		if e.writes != ft4Size/16-1 {
			t.Errorf("%d blocchi scritti", e.writes)
		}
	})
	// modello sbagliato: la radio è una FT-65, la scheda è la FT-4X
	e65 := ft4Fam{}.emu(genByID["ys-ft65e"], img).(*ft4Emu)
	ft4Run(t, m, e65, func(port string) {
		_, _, err := m.Fam.Download(ctx, m, port, nop)
		if err == nil || !strings.Contains(err.Error(), "FT-65") {
			t.Errorf("modello sbagliato: %v", err)
		}
		if _, err := m.Fam.Upload(ctx, m, port, img, nop); err == nil {
			t.Error("scrittura su un altro modello accettata")
		}
		if e65.writes != 0 {
			t.Error("blocchi scritti su un altro modello")
		}
	})
	// sottotipo 3: la scheda normale rimanda a quella giusta, e viceversa
	m65, m65s := genByID["ys-ft65e"], genByID["ys-ft65e-s3"]
	es := ft4Fam{}.emu(m65s, img).(*ft4Emu)
	ft4Run(t, m65, es, func(port string) {
		if _, _, err := m65.Fam.Download(ctx, m65, port, nop); err == nil || !strings.Contains(err.Error(), "sottotipo 3") {
			t.Errorf("sottotipo 3 con la scheda normale: %v", err)
		}
		if _, meta, _, err := (ft4Fam{}).DownloadMeta(ctx, m65s, port, nop); err != nil || meta["ft4_subtype"] != 3 {
			t.Errorf("sottotipo 3 con la sua scheda: %v %v", err, meta)
		}
	})
	en := ft4Fam{}.emu(m65, img).(*ft4Emu)
	ft4Run(t, m65s, en, func(port string) {
		if _, _, err := m65s.Fam.Download(ctx, m65s, port, nop); err == nil || !strings.Contains(err.Error(), "«Yaesu FT-65E»") {
			t.Errorf("radio normale con la scheda del sottotipo 3: %v", err)
		}
	})
	// cavo senza eco
	ne := ft4Fam{}.emu(m, img).(*ft4Emu)
	ne.noEcho = true
	ft4Run(t, m, ne, func(port string) {
		if _, _, err := m.Fam.Download(ctx, m, port, nop); err == nil || !strings.Contains(err.Error(), "cavo") {
			t.Errorf("cavo senza eco: %v", err)
		}
	})
}

func TestFT60(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	m := genByID["ys-ft60"]
	img := make([]byte, ft60Size)
	copy(img, "AH017")
	for i, c := range []*Chan{
		{Rx: 145500000, Tx: 145500000, Name: "S20", Scan: true},
		{Rx: 145612500, Tx: 145012500, TxTone: "88.5", Name: "R0A", Scan: true, Power: 1},
		{Rx: 430587500, Tx: 432187500, TxTone: "D023N", RxTone: "123.0", Name: "DT", Narrow: true, Power: 2},
		{Rx: 145000000, Tx: 145012500, TxTone: "100.0", RxTone: "D754N", Name: "ODD", Scan: true},
		{Rx: 118300000, Tx: 118300000, TxOff: true, Name: "TWR", Scan: true},
	} {
		if err := m.Fam.Set(m, img, i*7, c); err != nil {
			t.Fatal(err)
		}
		if g := m.Fam.Get(m, img, i*7); g == nil || *g != *c {
			t.Fatalf("canale %d riletto diverso: %+v / %+v", i, g, c)
		}
	}
	f := ft60Fam{}
	if !f.rec(28).F("isam").Bool(img) || f.rec(0).F("isam").Bool(img) {
		t.Error("AM solo nella gamma aeronautica")
	}
	for _, bad := range []*Chan{
		{Rx: 446006250, Tx: 446006250},                  // passo da 6,25 kHz
		{Rx: 145500000, Tx: 145500000, RxTone: "88.5"},  // solo tono in ricezione
		{Rx: 145500000, Tx: 145500000, TxTone: "D023I"}, // DCS invertito
		{Rx: 145500000, Tx: 145506250},                  // trasmissione non memorizzabile
		{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "100.0"},
	} {
		if err := m.Fam.Set(m, img, 100, bad); err == nil {
			t.Errorf("canale non valido accettato: %+v", bad)
		}
	}
	if m.Fam.Get(m, img, 100) != nil {
		t.Error("un canale rifiutato ha lasciato tracce")
	}
	run := func(e *ft60Emu, fn func()) {
		old := openPortFunc
		defer func() { openPortFunc = old }()
		openPortFunc = func(name string, baud int) (Port, error) {
			a, b := newPipe()
			go e.serve(b)
			return a, nil
		}
		fn()
	}
	// lettura con la somma di controllo sbagliata nella radio: rifiutata
	e := ft60Fam{}.emu(m, img).(*ft60Emu)
	run(e, func() {
		if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), "somma di controllo") {
			t.Errorf("somma di controllo errata: %v", err)
		}
	})
	// scrittura: la somma di controllo viene calcolata; poi la lettura torna uguale
	run(e, func() {
		if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err != nil {
			t.Fatal(err)
		}
	})
	want := append([]byte{}, img...)
	ft60Checksum(want)
	if e.writes != 1 || !bytes.Equal(e.image(), want) {
		t.Error("memoria finale diversa dall'immagine scritta")
	}
	run(e, func() {
		got, _, err := m.Fam.Download(ctx, m, "COM9", nop)
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("lettura dopo la scrittura: %v", err)
		}
	})
	// immagine di un'altra radio: non si scrive
	if _, err := m.Fam.Upload(ctx, m, "COM9", make([]byte, ft60Size), nop); err == nil {
		t.Error("immagine senza identificativo accettata")
	}
	// radio non in modalità clone: scrittura senza conferma
	s := &ft60Emu{img: make([]byte, ft60Size), silent: true}
	run(s, func() {
		if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err == nil || !strings.Contains(err.Error(), "cavo") {
			t.Errorf("radio muta: %v", err)
		}
	})
}
