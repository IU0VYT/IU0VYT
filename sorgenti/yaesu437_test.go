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
	"sort"
	"strings"
	"testing"
	"time"
)

func init() {
	// nelle prove le pause della scrittura un byte alla volta non servono
	yaesuByteDelay = [2]time.Duration{}
}

// yaesu437Run: fn con la radio simulata e sulla porta
func yaesu437Run(e genEmu, fn func()) {
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return a, nil
	}
	fn()
}

// 4.37: palmari Yaesu VX-5, VX-7, VX-2, FT-50 e VX-150 (il confronto byte per
// byte con CHIRP è nelle prove TestChirp*; qui i casi che quelle non toccano)
func TestYaesu437(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	type cas struct {
		id       string
		size     int
		lo       int64
		cross    bool // toni misti
		split    bool // trasmissione su una frequenza separata
		oddShift bool // scostamento di 612,5 kHz memorizzabile così com'è
		lower    bool // nomi con le minuscole
	}
	for _, cs := range []cas{
		{"ys-vx5", 8123, 145000000, false, true, false, true},
		{"ys-vx7", 16211, 145000000, true, true, false, true},
		{"ys-vx2", 32595, 145000000, false, true, false, false},
		{"ys-ft50", 3723, 145000000, false, false, true, false},
		{"ys-vx150", 5353, 145000000, false, false, false, false},
	} {
		m := genByID[cs.id]
		if m == nil || m.NoChirp || !m.NoTxOff || !m.HasScan || len(m.DCS) == 0 {
			t.Fatalf("%s: dati del modello", cs.id)
		}
		if _, ok := m.Fam.(interface{ noIdent() }); !ok {
			t.Errorf("%s: la copia parte dalla radio, «Identifica» non deve interrogarla", cs.id)
		}
		img := make([]byte, cs.size)
		if cs.id == "ys-vx150" {
			// memoria di fabbrica: tutti i canali vuoti
			copy(img, vx150Clone.id)
			f := vx150Fam{}
			for i := 0; i < m.Slots; i++ {
				f.rec(i).SetRaw(img, vx150Empty)
				f.nam(i).Fill(img, 0xFF)
			}
		}
		if cs.id == "ys-vx2" {
			copy(img, "AH015")
		}
		lo := cs.lo
		pw := func(k int) int { return m.Powers[k%len(m.Powers)].Sem }
		name := func(s string) string {
			if !cs.lower {
				s = strings.ToUpper(s)
			}
			if len(s) > m.NameMax {
				s = s[:m.NameMax]
			}
			return s
		}
		chs := []*Chan{
			{Rx: lo, Tx: lo, Name: name("S20"), Scan: true, Power: pw(0)},
			{Rx: lo + 12500, Tx: lo + 612500, TxTone: "88.5", Name: name("R0a"), Scan: true, Power: pw(1)},
			{Rx: lo + 600000, Tx: lo, TxTone: "123.0", RxTone: "123.0", Name: name("Tsq"), Narrow: m.Narrow, Power: pw(2)},
			{Rx: lo + 25000, Tx: lo + 25000, TxTone: "D754N", RxTone: "D754N", Name: name("DCS+"), Scan: true, Power: pw(3)},
			{Rx: lo + 50000, Tx: lo + 50000, Name: "", Scan: true, Power: pw(0)},
		}
		if cs.oddShift || cs.split {
			chs = append(chs, &Chan{Rx: lo + 100000, Tx: lo + 712500, Name: name("ODD"), Scan: true, Power: pw(0)})
		}
		if cs.split {
			chs = append(chs, &Chan{Rx: lo, Tx: 433000000, Name: name("FAR"), Scan: true, Power: pw(0)})
		}
		if cs.cross {
			chs = append(chs,
				&Chan{Rx: lo, Tx: lo, TxTone: "D023N", Name: name("DT"), Scan: true, Power: pw(0)},
				&Chan{Rx: lo, Tx: lo, TxTone: "100.0", RxTone: "D032N", Name: name("TD"), Scan: true, Power: pw(0)},
				&Chan{Rx: lo, Tx: lo, TxTone: "D051N", RxTone: "131.8", Name: name("DTn"), Scan: true, Power: pw(0)})
		}
		for i, c := range chs {
			if err := m.Fam.Set(m, img, i*3, c); err != nil {
				t.Fatalf("%s canale %d: %v", cs.id, i, err)
			}
			if g := m.Fam.Get(m, img, i*3); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso:\n  %+v\n  %+v", cs.id, i, g, c)
			}
		}
		bad := []*Chan{
			{Rx: lo + 1234, Tx: lo + 1234},                     // frequenza non memorizzabile
			{Rx: lo, Tx: lo, RxTone: "88.5"},                   // solo tono in ricezione
			{Rx: lo, Tx: lo, TxTone: "D023I", RxTone: "D023I"}, // DCS invertito
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "100.0"},
			{Rx: lo, Tx: lo, TxTone: "D023N", RxTone: "D025N"},
		}
		if !cs.cross {
			bad = append(bad, &Chan{Rx: lo, Tx: lo, TxTone: "D023N"}, &Chan{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "D023N"})
		}
		if !cs.split {
			bad = append(bad, &Chan{Rx: lo, Tx: lo + 1234567})
		}
		if cs.id != "ys-vx7" && cs.id != "ys-vx2" {
			bad = append(bad, &Chan{Rx: lo, Tx: lo, TxTone: "159.8"}) // subtono che queste radio non hanno
		}
		free, copyTo := m.Slots-1, m.Slots-5
		for _, b := range bad {
			if err := m.Fam.Set(m, img, free, b); err == nil {
				t.Errorf("%s: canale non valido accettato: %+v", cs.id, b)
			}
		}
		if m.Fam.Get(m, img, free) != nil {
			t.Errorf("%s: un canale rifiutato ha lasciato tracce", cs.id)
		}
		// cancellazione, poi il canale si può rimettere
		if err := m.Fam.Set(m, img, 3, nil); err != nil || m.Fam.Get(m, img, 3) != nil {
			t.Fatalf("%s: cancellazione", cs.id)
		}
		if err := m.Fam.Set(m, img, 3, chs[1]); err != nil {
			t.Fatal(err)
		}
		if g := m.Fam.Get(m, img, 3); g == nil || *g != *chs[1] {
			t.Errorf("%s: canale rimesso dopo la cancellazione: %+v", cs.id, g)
		}
		// dati grezzi del canale: copia in un'altra posizione
		m.Fam.PutRec(m, img, copyTo, m.Fam.Rec(m, img, 6))
		if g := m.Fam.Get(m, img, copyTo); g == nil || *g != *chs[2] {
			t.Errorf("%s: copia grezza diversa: %+v", cs.id, g)
		}
		// radio simulata: scrittura (le somme di controllo vengono calcolate), poi
		// la lettura torna uguale
		e := genEmuFor(m, make([]byte, cs.size))
		yaesu437Run(e, func() {
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", cs.id, err)
			}
		})
		time.Sleep(50 * time.Millisecond)
		var got []byte
		yaesu437Run(e, func() {
			var err error
			if got, _, err = m.Fam.Download(ctx, m, "COM9", nop); err != nil {
				t.Fatalf("%s: lettura: %v", cs.id, err)
			}
		})
		if len(got) != cs.size || !m.Fam.Match(m, got) {
			t.Fatalf("%s: memoria letta non riconosciuta (%d byte)", cs.id, len(got))
		}
		for i, c := range chs {
			if i == 1 {
				continue // riscritto più sotto
			}
			if g := m.Fam.Get(m, got, i*3); g == nil || *g != *c {
				t.Errorf("%s: canale %d dopo scrittura e lettura: %+v", cs.id, i, g)
			}
		}
		// la memoria letta differisce da quella scritta solo nelle somme di controllo
		diff := 0
		for k := range got {
			if got[k] != img[k] {
				diff++
			}
		}
		if diff > 3 {
			t.Errorf("%s: %d byte diversi fra la memoria scritta e quella riletta", cs.id, diff)
		}
		// somma di controllo sbagliata nella radio: lettura rifiutata
		switch x := e.(type) {
		case *yaesuEmu:
			x.mu.Lock()
			x.img[x.o.sums[len(x.o.sums)-1][2]]++
			x.mu.Unlock()
		case *ftxEmu:
			x.mu.Lock()
			x.img[len(x.img)-1]++
			x.mu.Unlock()
		}
		yaesu437Run(e, func() {
			if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil || !strings.Contains(err.Error(), "somma di controllo") {
				t.Errorf("%s: somma di controllo errata: %v", cs.id, err)
			}
		})
		// immagine di lunghezza sbagliata: non si scrive
		if _, err := m.Fam.Upload(ctx, m, "COM9", make([]byte, cs.size-1), nop); err == nil {
			t.Errorf("%s: immagine troppo corta accettata", cs.id)
		}
	}
}

// 4.37: copia in più blocchi con un cavo a due fili (senza eco) e con un blocco
// che comincia con 06: il byte non va scambiato per l'eco della conferma
func TestYaesuMultiBlock(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	m := genByID["ys-vx7"]
	o := vx7Clone
	img := make([]byte, o.size)
	for k := range img {
		img[k] = byte(k * 7)
	}
	img[10] = 0x06 // primo byte del secondo blocco
	img[18] = 0x06 // primo byte del terzo blocco
	yaesuSums(o, img)
	for _, noEcho := range []bool{false, true} {
		e := newYaesuEmu(o, img)
		e.noEcho = noEcho
		yaesu437Run(e, func() {
			got, _, err := m.Fam.Download(ctx, m, "COM9", nop)
			if err != nil || !bytes.Equal(got, img) {
				t.Errorf("senza eco %v: lettura: %v", noEcho, err)
			}
		})
		e2 := newYaesuEmu(o, make([]byte, o.size))
		e2.noEcho = noEcho
		yaesu437Run(e2, func() {
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err != nil {
				t.Errorf("senza eco %v: scrittura: %v", noEcho, err)
			}
		})
		time.Sleep(50 * time.Millisecond)
		if e2.writes != 1 || !bytes.Equal(e2.image(), img) {
			t.Errorf("senza eco %v: memoria scritta diversa", noEcho)
		}
	}
}

// VX-8: CHIRP toglie il canale dai banchi a ogni modifica, qui solo quando viene
// cancellato (differenza voluta, vedi drv_ysf.go). Il confronto rimette nei banchi
// di CHIRP i canali che esistono ancora; i canali cancellati restano controllati.
func (f ysfFam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	if !f.opt(m).vx8 {
		return
	}
	e := f.mem(m)
	for b := 0; b < 24; b++ {
		list := func(img []byte) []int {
			var l []int
			if e.F("bank_used").I(b).F("in_use").Int(img) == 0xFFFF {
				return l
			}
			for _, x := range e.F("bank_members").I(b).F("channel").Elems {
				if v := int(x.Int(img)); v != 0xFFFF {
					l = append(l, v)
				}
			}
			return l
		}
		have := map[int]bool{}
		all := list(ref)
		for _, v := range all {
			have[v] = true
		}
		added := false
		for _, v := range list(out) {
			if have[v] || v >= m.Slots {
				continue
			}
			if fl := f.flag(m, v); fl.F("used").Bool(out) && fl.F("valid").Bool(out) {
				all, added = append(all, v), true
			}
		}
		if !added {
			continue
		}
		sort.Ints(all)
		for j, x := range e.F("bank_members").I(b).F("channel").Elems {
			v := int64(0xFFFF)
			if j < len(all) {
				v = int64(all[j])
			}
			x.SetInt(ref, v)
		}
		e.F("bank_used").I(b).F("in_use").SetInt(ref, e.F("bank_used").I(b).F("in_use").Int(out))
	}
}

// VX-8: i casi che il confronto con CHIRP non tocca (canale a 6,25 kHz, banchi
// conservati dopo una modifica e liberati dopo una cancellazione)
func TestYaesuVX8(t *testing.T) {
	for _, id := range []string{"ys-vx8r", "ys-vx8dr", "ys-vx8ge"} {
		m := genByID[id]
		if m == nil || m.Narrow || m.Digital != "" || m.Slots != 900 || m.NameMax != 16 {
			t.Fatalf("%s: dati del modello", id)
		}
		f := m.Fam.(ysfFam)
		o := f.opt(m)
		if !o.vx8 || !o.coded || o.clone.size != 65227 {
			t.Fatalf("%s: opzioni", id)
		}
		img := make([]byte, o.clone.size)
		copy(img, o.clone.model)
		e := f.mem(m)
		for b := 0; b < 24; b++ {
			e.F("bank_used").I(b).F("in_use").SetInt(img, 0xFFFF)
			e.F("bank_members").I(b).Fill(img, 0xFF)
		}
		chs := []*Chan{
			{Rx: 145000000, Tx: 145000000, Name: "Simplex 145", Scan: true, Power: m.Powers[0].Sem},
			{Rx: 145606250, Tx: 145006250, TxTone: "88.5", Name: "R0a 6.25", Scan: true, Power: m.Powers[1].Sem},
			{Rx: 430012500, Tx: 431612500, TxTone: "D754N", RxTone: "D754N", Name: "dcs", Power: m.Powers[3].Sem},
			{Rx: 145500000, Tx: 433500000, TxTone: "123.0", RxTone: "123.0", Name: "Split", Scan: true, Power: m.Powers[2].Sem},
		}
		for i, c := range chs {
			if err := f.Set(m, img, i, c); err != nil {
				t.Fatalf("%s canale %d: %v", id, i, err)
			}
			if g := f.Get(m, img, i); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso:\n  %+v\n  %+v", id, i, g, c)
			}
		}
		if v := f.rec(m, 0).F("unknown1b").Int(img); v != 0 {
			t.Errorf("%s: 145,000 MHz con il valore dei 6,25 kHz (%d)", id, v)
		}
		if v := f.rec(m, 1).F("unknown1b").Int(img); v != 5 {
			t.Errorf("%s: 145,60625 MHz senza il valore dei 6,25 kHz (%d)", id, v)
		}
		// banda stretta: la radio non l'ha, il canale si scrive lo stesso
		n := *chs[0]
		n.Narrow = true
		if err := f.Set(m, img, 0, &n); err != nil {
			t.Fatal(err)
		}
		if g := f.Get(m, img, 0); g == nil || g.Narrow || f.rec(m, 0).F("half_deviation").Bool(img) {
			t.Errorf("%s: banda stretta rimasta nel canale", id)
		}
		// canali 0 e 1 nel banco 3
		e.F("bank_used").I(3).F("in_use").SetInt(img, 6)
		e.F("bank_members").I(3).F("channel").I(0).SetInt(img, 0)
		e.F("bank_members").I(3).F("channel").I(1).SetInt(img, 1)
		c := *chs[1]
		c.Name = "Cambiato"
		if err := f.Set(m, img, 1, &c); err != nil {
			t.Fatal(err)
		}
		if v := e.F("bank_members").I(3).F("channel").I(1).Int(img); v != 1 {
			t.Errorf("%s: la modifica ha tolto il canale dal banco (%d)", id, v)
		}
		if err := f.Set(m, img, 1, nil); err != nil {
			t.Fatal(err)
		}
		if f.Get(m, img, 1) != nil || f.flag(m, 1).F("valid").Bool(img) || f.flag(m, 1).F("used").Bool(img) {
			t.Errorf("%s: canale cancellato ancora presente", id)
		}
		if a, b := e.F("bank_members").I(3).F("channel").I(0).Int(img), e.F("bank_members").I(3).F("channel").I(1).Int(img); a != 0 || b != 0xFFFF {
			t.Errorf("%s: banco dopo la cancellazione: %d %d", id, a, b)
		}
		if err := f.Set(m, img, 0, nil); err != nil {
			t.Fatal(err)
		}
		if v := e.F("bank_used").I(3).F("in_use").Int(img); v != 0xFFFF {
			t.Errorf("%s: banco vuoto ancora in uso (%d)", id, v)
		}
		// un'altra VX-8 non viene presa per questa
		for _, other := range []string{"ys-vx8r", "ys-vx8dr", "ys-vx8ge"} {
			if got := genByID[other].Fam.Match(genByID[other], img); got != (other == id) {
				t.Errorf("%s: immagine riconosciuta come %s = %v", id, other, got)
			}
		}
	}
}
