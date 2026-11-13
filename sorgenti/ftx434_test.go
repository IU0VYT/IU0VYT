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

func init() {
	// nelle prove la pausa fra i blocchi non serve
	ftxBlockDelay = 0
	// CHIRP non ha un'immagine di prova della VX-170: quella della VX-177 con
	// la sigla della VX-170 (stessa memoria)
	chirpTestImage["ys-ft7800"] = "Yaesu_FT-7800_7900.img"
	chirpTestImage["ys-vx170"] = "Yaesu_VX-177.img"
	chirpImagePatch["ys-vx170"] = func(b []byte) {
		copy(b, "AH022$")
		b[len(b)-1] = ftxSum(b)
	}
}

// 4.34: Yaesu VX-170 / VX-177, FT-7800 / FT-7900, FT-8900 (il confronto byte per
// byte con CHIRP è nelle prove TestChirp*; qui i casi che quelle non toccano)
func TestFtx434(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ys-vx170", "ys-vx177", "ys-ft7800", "ys-ft8900"} {
		m := genByID[id]
		if m == nil || m.NoChirp || !m.NoTxOff || !m.HasScan || !m.Narrow || len(m.DCS) == 0 {
			t.Fatalf("%s: dati del modello", id)
		}
		o := ftxo(m)
		f := ftxFam{}
		img := make([]byte, o.size)
		copy(img, o.id)
		lo := m.Bands[0].Lo + 1000000
		if o.am >= 0 {
			lo = 145000000 // fuori dalla gamma aeronautica, che resta in AM
		}
		chs := []*Chan{
			{Rx: lo, Tx: lo, Name: "S20", Scan: true},
			{Rx: lo + 12500, Tx: lo + 612500, TxTone: "88.5", Name: "R0A", Scan: true, Power: 1},
			{Rx: lo + 587500, Tx: lo - 1012500, TxTone: "123.0", RxTone: "123.0", Name: "TSQ", Narrow: true, Power: 2},
			{Rx: lo + 7500, Tx: lo + 7500, TxTone: "D754N", RxTone: "D754N", Name: "DCS*+_", Scan: true},
			{Rx: lo, Tx: lo + 625000, Name: "ODD", Scan: true},       // scostamento non multiplo di 50 kHz
			{Rx: lo, Tx: lo + 13000000, Name: "FAR", Scan: true},     // scostamento oltre 12,75 MHz
			{Rx: lo + 5000, Tx: lo + 12755000, Name: "", Scan: true}, // scostamento massimo, senza nome
		}
		for i, c := range chs {
			if err := m.Fam.Set(m, img, i*7, c); err != nil {
				t.Fatalf("%s canale %d: %v", id, i, err)
			}
			if g := m.Fam.Get(m, img, i*7); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso: %+v / %+v", id, i, g, c)
			}
		}
		// i due scostamenti «strani» sono scritti come trasmissione separata, l'ultimo no
		for k, want := range map[int]int64{28: 4, 35: 4, 42: 3, 7: 3, 14: 2, 0: 0} {
			if d := f.rec(m, k).F("duplex").Int(img); d != want {
				t.Errorf("%s canale %d: duplex %d, atteso %d", id, k, d, want)
			}
		}
		// frequenza: 6 cifre in 10 kHz più i due bit per 5 e 2,5 kHz
		raw := f.rec(m, 21).F("freq").Raw(img)
		if raw[0]&0xC0 != 0xC0 || ftxDecFreq(raw) != lo+7500 {
			t.Errorf("%s: frequenza con 7,5 kHz scritta % X", id, raw)
		}
		for _, bad := range []*Chan{
			{Rx: lo + 6250, Tx: lo + 6250},                     // passo da 6,25 kHz
			{Rx: lo, Tx: lo, RxTone: "88.5"},                   // solo tono in ricezione
			{Rx: lo, Tx: lo, TxTone: "D023I", RxTone: "D023I"}, // DCS invertito
			{Rx: lo, Tx: lo + 81250},                           // trasmissione non memorizzabile
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "100.0"},
			{Rx: lo, Tx: lo, TxTone: "D023N", RxTone: "D025N"},
			{Rx: lo, Tx: lo, TxTone: "D023N"},
			{Rx: lo, Tx: lo, TxOff: true},
		} {
			if err := m.Fam.Set(m, img, 100, bad); err == nil {
				t.Errorf("%s: canale non valido accettato: %+v", id, bad)
			}
		}
		if m.Fam.Get(m, img, 100) != nil {
			t.Errorf("%s: un canale rifiutato ha lasciato tracce", id)
		}
		// modifica di un canale con trasmissione separata verso il simplex e ritorno
		c := *chs[5]
		c.Tx = c.Rx
		if err := m.Fam.Set(m, img, 35, &c); err != nil {
			t.Fatal(err)
		}
		if g := m.Fam.Get(m, img, 35); g == nil || *g != c {
			t.Errorf("%s: da separata a simplex: %+v", id, g)
		}
		// cancellazione: resta tutto tranne il segno «in uso»; il canale si può rimettere
		before := append([]byte{}, f.rec(m, 7).Raw(img)...)
		if err := m.Fam.Set(m, img, 7, nil); err != nil || m.Fam.Get(m, img, 7) != nil {
			t.Fatalf("%s: cancellazione", id)
		}
		after := f.rec(m, 7).Raw(img)
		if !bytes.Equal(before[1:], after[1:]) || before[0]&0x80 == 0 || after[0]&0x80 != 0 {
			t.Errorf("%s: la cancellazione ha toccato altri byte", id)
		}
		if err := m.Fam.Set(m, img, 7, chs[1]); err != nil {
			t.Fatal(err)
		}
		// dati grezzi del canale: copia in un'altra posizione
		m.Fam.PutRec(m, img, 150, m.Fam.Rec(m, img, 14))
		if g := m.Fam.Get(m, img, 150); g == nil || *g != *chs[2] {
			t.Errorf("%s: copia grezza diversa: %+v", id, g)
		}
		// AM: solo nelle veicolari e solo nella gamma aeronautica
		if o.am >= 0 {
			tw := &Chan{Rx: 118300000, Tx: 118300000, Name: "TWR", Scan: true}
			if err := m.Fam.Set(m, img, 160, tw); err != nil {
				t.Fatal(err)
			}
			if f.rec(m, 160).F("mode").Int(img) != o.am || f.rec(m, 0).F("mode").Int(img) != 0 {
				t.Errorf("%s: AM solo nella gamma aeronautica", id)
			}
			// stesso canale, altro nome: resta in AM
			tw.Name = "TOWER"
			if err := m.Fam.Set(m, img, 160, tw); err != nil || f.rec(m, 160).F("mode").Int(img) != o.am {
				t.Errorf("%s: l'AM non è rimasta", id)
			}
		}
		if o.banks {
			// il canale cancellato esce dai banchi, gli altri no
			w := f.mem(m).F("bank_channels").I(3).F("bitmap").I(0)
			w.SetInt(img, 0xFFFFFFFF)
			if err := m.Fam.Set(m, img, 21, nil); err != nil {
				t.Fatal(err)
			}
			if got := w.Int(img); got != 0xFFFFFFFF&^(1<<uint(31-21)) {
				t.Errorf("%s: banchi dopo la cancellazione %08X", id, got)
			}
		}
		if o.k89 {
			// FT-8900: fuori da 108-480 MHz il canale non va sulla banda di destra
			hf := &Chan{Rx: 29600000, Tx: 29500000, Name: "10M", Scan: true}
			if err := m.Fam.Set(m, img, 170, hf); err != nil {
				t.Fatal(err)
			}
			if g := m.Fam.Get(m, img, 170); g == nil || *g != *hf {
				t.Errorf("%s: 29 MHz riletto %+v", id, g)
			}
			if f.rec(m, 170).F("sub_used").Bool(img) || !f.rec(m, 14).F("sub_used").Bool(img) {
				t.Errorf("%s: segno della banda di destra", id)
			}
		}
		run := func(e *ftxEmu, fn func()) {
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
		e := f.emu(m, img).(*ftxEmu)
		e.img[o.size-1] = ftxSum(img) + 1
		run(e, func() {
			if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), "somma di controllo") {
				t.Errorf("%s: somma di controllo errata: %v", id, err)
			}
		})
		// scrittura: la somma di controllo viene calcolata; poi la lettura torna uguale
		run(e, func() {
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err != nil {
				t.Fatal(err)
			}
		})
		want := append([]byte{}, img...)
		want[o.size-1] = ftxSum(want)
		if e.writes != 1 || !bytes.Equal(e.image(), want) {
			t.Errorf("%s: memoria finale diversa dall'immagine scritta", id)
		}
		run(e, func() {
			got, _, err := m.Fam.Download(ctx, m, "COM9", nop)
			if err != nil || !bytes.Equal(got, want) {
				t.Errorf("%s: lettura dopo la scrittura: %v", id, err)
			}
		})
		// immagine di un'altra radio: non si scrive
		if _, err := m.Fam.Upload(ctx, m, "COM9", make([]byte, o.size), nop); err == nil {
			t.Errorf("%s: immagine senza identificativo accettata", id)
		}
		// radio di un altro modello della famiglia: la lettura lo dice
		other := "ys-vx177"
		if id == other {
			other = "ys-vx170"
		}
		om := genByID[other]
		oi := make([]byte, ftxo(om).size)
		copy(oi, ftxo(om).id)
		oi[len(oi)-1] = ftxSum(oi)
		run(f.emu(om, oi).(*ftxEmu), func() {
			if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), om.Name) {
				t.Errorf("%s: radio di un altro modello: %v", id, err)
			}
		})
		// radio non in modalità clone: scrittura senza conferma
		s := &ftxEmu{o: o, img: make([]byte, o.size), silent: true}
		run(s, func() {
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err == nil || !strings.Contains(err.Error(), "cavo") {
				t.Errorf("%s: radio muta: %v", id, err)
			}
		})
	}
}
