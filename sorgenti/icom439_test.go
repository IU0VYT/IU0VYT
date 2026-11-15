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

// 4.39: Icom ID-80H e ID-880H (il confronto byte per byte con CHIRP è nelle
// prove TestChirp*; qui i casi che quelle non toccano)
func TestIcom439(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-id80", "ic-id880"} {
		m := genByID[id]
		o, ok := icomOptOf(m)
		if m == nil || !ok || m.NoChirp || !m.NoTxOff || !m.HasScan || !m.Narrow || m.Digital != "D-STAR" || m.Slots != 1000 || m.First != 0 {
			t.Fatalf("%s: dati del modello", id)
		}
		f := m.Fam.(id880Fam)
		img := bytes.Repeat([]byte{0xFF}, o.size)
		copy(img[0xF000:], f.opt(m).tag)
		for i := 0; i < m.Slots; i++ {
			if f.Get(m, img, i) != nil {
				t.Fatalf("%s: memoria a FF, canale %d non vuoto", id, i)
			}
		}
		chs := []*Chan{
			{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"},
			{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true, Name: "R0"},
			{Rx: 430012500, Tx: 431612500, TxTone: "123.0", RxTone: "123.0", Name: "TSQL", Narrow: true},
			{Rx: 145606250, Tx: 145006250, TxTone: "D754I", RxTone: "D754N", Scan: true, Name: "DCS 6.25"},
			{Rx: 145375000, Tx: 145375000, Scan: true, Name: "DV", Dig: true},
			{Rx: 430000000, Tx: 435000000, Scan: true, Name: "DV RPT", Dig: true},
			{Rx: 433500000, Tx: 433500000, TxTone: "254.1", Scan: true},
		}
		// i canali 98-101 sono quelli che l'«indicatore del nome» di CHIRP rovinerebbe
		pos := []int{0, 3, 98, 99, 100, 101, 999}
		for k, c := range chs {
			if err := f.Set(m, img, pos[k], c); err != nil {
				t.Fatalf("%s canale %d: %v", id, pos[k], err)
			}
		}
		for k, c := range chs {
			if g := f.Get(m, img, pos[k]); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso: %+v / %+v", id, pos[k], g, c)
			}
		}
		// frequenze: 5, 6,25 e 9 kHz; canale D-STAR nuovo con la chiamata generale
		if v := f.rec(m, 99).F("rxmult").Int(img); v != 1 {
			t.Errorf("%s: 145,60625 MHz con il moltiplicatore %d", id, v)
		}
		if !bytes.Equal(f.rec(m, 100).F("urcall").Raw(img), id31Call("CQCQCQ")) {
			t.Errorf("%s: canale D-STAR nuovo senza «CQCQCQ»", id)
		}
		// combinazioni che la radio non ha
		bad := []*Chan{
			{Rx: 145500000, Tx: 145500000, RxTone: "88.5"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "100.0"},
			{Rx: 145500100, Tx: 145500100},
			{Rx: 145500000, Tx: 145501234},
			{Rx: 145500000, Tx: 215500000},
		}
		before := append([]byte{}, img...)
		for _, c := range bad {
			if err := f.Set(m, img, 40, c); err == nil {
				t.Errorf("%s: canale non valido accettato: %+v", id, c)
			}
		}
		if !bytes.Equal(before, img) {
			t.Errorf("%s: un canale rifiutato ha lasciato tracce", id)
		}
		// AM in banda aerea (anche a 9 kHz sulle onde medie per la ID-80H) e modo che resta
		air := &Chan{Rx: 118100000, Tx: 118100000, Scan: true, Name: "TWR"}
		if err := f.Set(m, img, 41, air); err != nil || f.rec(m, 41).F("mode").Int(img) != 3 {
			t.Errorf("%s: banda aerea non in AM (%v)", id, err)
		}
		if id == "ic-id80" {
			mw := &Chan{Rx: 909000, Tx: 909000, Scan: true, Name: "MW"}
			if err := f.Set(m, img, 42, mw); err != nil || f.rec(m, 42).F("rxmult").Int(img) != 4 {
				t.Errorf("%s: 909 kHz (multiplo di 9 kHz): %v", id, err)
			}
			if g := f.Get(m, img, 42); g == nil || g.Rx != 909000 {
				t.Errorf("%s: 909 kHz riletto %+v", id, g)
			}
		}
		// salto «P»: resta finché il canale resta escluso dalla scansione
		e := f.mem(m)
		icomSetBit(e.F("pskip_flags"), img, 0, true)
		c0 := *chs[0]
		c0.Scan = false
		if g := f.Get(m, img, 0); g == nil || g.Scan {
			t.Errorf("%s: canale con il salto «P» dato per incluso nella scansione", id)
		}
		if err := f.Set(m, img, 0, &c0); err != nil || !icomBit(e.F("pskip_flags"), img, 0) || icomBit(e.F("skip_flags"), img, 0) {
			t.Errorf("%s: salto «P» perso (%v)", id, err)
		}
		c0.Scan = true
		if err := f.Set(m, img, 0, &c0); err != nil || icomBit(e.F("pskip_flags"), img, 0) || icomBit(e.F("skip_flags"), img, 0) {
			t.Errorf("%s: salto non tolto (%v)", id, err)
		}
		// cancellazione: record a FF, fuori dal banco, e canale nuovo nello stesso posto
		f.bank(m, 3).F("bank").SetInt(img, 2)
		if err := f.Set(m, img, 3, nil); err != nil || f.Get(m, img, 3) != nil || f.bank(m, 3).F("bank").Int(img) != 0xFF {
			t.Errorf("%s: canale non cancellato", id)
		}
		if err := f.Set(m, img, 3, chs[1]); err != nil {
			t.Fatal(err)
		}
		if g := f.Get(m, img, 3); g == nil || *g != *chs[1] {
			t.Errorf("%s: canale nuovo dopo la cancellazione %+v", id, g)
		}
		// i due modelli si distinguono dalla sigla
		for _, other := range []string{"ic-id80", "ic-id880"} {
			if got := genByID[other].Fam.Match(genByID[other], img); got != (other == id) {
				t.Errorf("%s: immagine riconosciuta come %s = %v", id, other, got)
			}
		}
		// cavo: lettura e scrittura con la radio simulata; l'altro modello viene rifiutato
		emu := newIcomEmu(o, img)
		icomRun(emu, func() {
			got, _, err := f.Download(ctx, m, "COM9", nop)
			if err != nil || !bytes.Equal(got, img) {
				t.Fatalf("%s: lettura: %v", id, err)
			}
			mod := append([]byte{}, got...)
			if err := f.Set(m, mod, 20, &Chan{Rx: 145700000, Tx: 145100000, TxTone: "100.0", Scan: true, Name: "NEW"}); err != nil {
				t.Fatal(err)
			}
			if _, err := f.Upload(ctx, m, "COM9", mod, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
			if !bytes.Equal(emu.image(), mod) {
				t.Errorf("%s: la memoria della radio non è quella scritta", id)
			}
			other := genByID["ic-id880"]
			if id == "ic-id880" {
				other = genByID["ic-id80"]
			}
			if _, _, err := other.Fam.Download(ctx, other, "COM9", nop); err == nil || !strings.Contains(err.Error(), "modello") {
				t.Errorf("%s letta con la scheda %s: %v", id, other.ID, err)
			}
		})
	}
}

// 4.39: Icom IC-W32A e IC-W32E: due metà (VHF e UHF) in una lista sola
func TestIcomW32(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-w32a", "ic-w32e"} {
		m := genByID[id]
		o, ok := icomOptOf(m)
		if m == nil || !ok || m.NoChirp || !m.NoTxOff || !m.HasScan || m.Narrow || len(m.DCS) != 0 || m.Slots != 200 || m.First != 0 {
			t.Fatalf("%s: dati del modello", id)
		}
		f := m.Fam.(icw32Fam)
		img := bytes.Repeat([]byte{0xFF}, o.size)
		copy(img[0x0FE0-len(icw32Tag):], icw32Tag)
		for k := 0x0FE0; k < o.size; k++ {
			img[k] = 0
		}
		for i := 0; i < m.Slots; i++ {
			if f.Get(m, img, i) != nil {
				t.Fatalf("%s: memoria a FF, canale %d non vuoto", id, i)
			}
		}
		limits := append([]byte{}, img[0x0640:0x06E0]...) // posizioni 100-109 della metà VHF: limiti di scansione
		chs := map[int]*Chan{
			0:   {Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"},
			99:  {Rx: 145612500, Tx: 145012500, TxTone: "88.5", Scan: true, Name: "R0A-12.5"},
			100: {Rx: 430012500, Tx: 431612500, TxTone: "123.0", RxTone: "123.0", Name: "TSQL"},
			199: {Rx: 433500000, Tx: 433500000, TxTone: "254.1", Scan: true},
			50:  {Rx: 118100000, Tx: 118100000, Scan: true, Name: "TWR"},
		}
		for i, c := range chs {
			if err := f.Set(m, img, i, c); err != nil {
				t.Fatalf("%s canale %d: %v", id, i, err)
			}
		}
		for i, c := range chs {
			if g := f.Get(m, img, i); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso: %+v / %+v", id, i, g, c)
			}
		}
		if !bytes.Equal(limits, img[0x0640:0x06E0]) {
			t.Errorf("%s: i limiti di scansione della metà VHF sono stati toccati", id)
		}
		if !f.flg(50).F("am").Bool(img) || f.flg(0).F("am").Bool(img) {
			t.Errorf("%s: AM solo in banda aerea", id)
		}
		// ogni metà accetta solo la sua banda; niente DCS, niente frequenze non memorizzabili
		bad := map[int]*Chan{
			10:  {Rx: 433500000, Tx: 433500000},
			110: {Rx: 145500000, Tx: 145500000},
			11:  {Rx: 145500000, Tx: 145500000, TxTone: "D023N", RxTone: "D023N"},
			12:  {Rx: 145500000, Tx: 145500000, RxTone: "88.5"},
			13:  {Rx: 145500100, Tx: 145500100},
			14:  {Rx: 145500000, Tx: 433500000},
		}
		before := append([]byte{}, img...)
		for i, c := range bad {
			if err := f.Set(m, img, i, c); err == nil {
				t.Errorf("%s: canale non valido accettato in %d: %+v", id, i, c)
			}
		}
		if lk := genLockedFn(m, img); lk == nil || lk(10, nil, bad[10]) == nil || lk(110, nil, bad[110]) == nil || lk(10, nil, chs[0]) != nil {
			t.Errorf("%s: regola della banda per posizione", id)
		}
		// spostamento: dentro la stessa metà sì, da una metà all'altra no
		f.PutRec(m, img, 20, f.Rec(m, img, 0))
		if a, b := f.Get(m, img, 0), f.Get(m, img, 20); a == nil || b == nil || *a != *b {
			t.Errorf("%s: canale spostato diverso", id)
		}
		if err := f.Set(m, img, 20, nil); err != nil || f.Get(m, img, 20) != nil {
			t.Errorf("%s: canale non cancellato", id)
		}
		before = append(before[:0], img...)
		f.PutRec(m, img, 120, f.Rec(m, img, 0))
		if !bytes.Equal(before, img) {
			t.Errorf("%s: un canale VHF è finito nella metà UHF", id)
		}
		// i due modelli si distinguono dal byte in coda al file
		for _, other := range []string{"ic-w32a", "ic-w32e"} {
			if got := genByID[other].Fam.Match(genByID[other], img); got != (other == id) {
				t.Errorf("%s: immagine riconosciuta come %s = %v", id, other, got)
			}
		}
		// cavo: lettura e scrittura con la radio simulata
		emu := newIcomEmu(o, img)
		icomRun(emu, func() {
			got, _, err := f.Download(ctx, m, "COM9", nop)
			if err != nil || !bytes.Equal(got, img) {
				t.Fatalf("%s: lettura: %v", id, err)
			}
			mod := append([]byte{}, got...)
			if err := f.Set(m, mod, 130, &Chan{Rx: 431000000, Tx: 432600000, TxTone: "100.0", Scan: true, Name: "NEW"}); err != nil {
				t.Fatal(err)
			}
			if _, err := f.Upload(ctx, m, "COM9", mod, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
			if !bytes.Equal(emu.image(), mod) {
				t.Errorf("%s: la memoria della radio non è quella scritta", id)
			}
		})
	}
}
