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

// 4.33: Icom IC-T7H, IC-T8A, IC-Q7A, IC-P7 (il confronto byte per byte con
// CHIRP è nelle prove TestChirp*; qui i casi che quelle non toccano)
func TestIcom433(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-t7h", "ic-t8a", "ic-q7a", "ic-p7"} {
		m := genByID[id]
		o, ok := icomOptOf(m)
		if m == nil || !ok || m.NoChirp || !m.NoTxOff || !m.HasScan || m.Narrow {
			t.Fatalf("%s: dati del modello", id)
		}
		dcs := len(m.DCS) > 0
		if dcs != (id == "ic-p7") {
			t.Fatalf("%s: DCS", id)
		}
		img := bytes.Repeat([]byte{0xFF}, o.size)
		for i := 0; i < m.Slots; i++ {
			if m.Fam.Get(m, img, i) != nil {
				t.Fatalf("%s: memoria a FF, canale %d non vuoto", id, i)
			}
		}
		named := m.NameMax > 0
		nm := func(s string) string {
			if named {
				return s
			}
			return ""
		}
		chs := []*Chan{
			{Rx: 145500000, Tx: 145500000, Scan: true, Name: nm("S20")},
			{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true, Name: nm("R0")},
			{Rx: 430012500, Tx: 431612500, TxTone: "123.0", RxTone: "123.0", Name: nm("TSQL")},
			{Rx: 433500000, Tx: 433500000, TxTone: "254.1", Scan: true},
		}
		if dcs {
			chs = append(chs, &Chan{Rx: 145212500, Tx: 145212500, TxTone: "D754I", RxTone: "D754N", Scan: true, Name: "DCS"})
		}
		for k, c := range chs {
			if err := m.Fam.Set(m, img, k*3, c); err != nil {
				t.Fatalf("%s canale %d: %v", id, k, err)
			}
			if g := m.Fam.Get(m, img, k*3); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso: %+v / %+v", id, k, g, c)
			}
		}
		// combinazioni che la radio non ha
		bad := []*Chan{
			{Rx: 145500000, Tx: 145500000, RxTone: "88.5"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "100.0"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "D023N"},
			{Rx: 145500100, Tx: 145500100},
		}
		if !dcs {
			bad = append(bad, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "D023N", RxTone: "D023N"})
		}
		if id == "ic-p7" {
			bad = bad[:3] // la frequenza è in terzi di Hz: ogni valore è memorizzabile
		}
		before := append([]byte{}, img...)
		for _, c := range bad {
			if err := m.Fam.Set(m, img, 40, c); err == nil {
				t.Errorf("%s: canale non valido accettato: %+v", id, c)
			}
		}
		if !bytes.Equal(before, img) {
			t.Errorf("%s: un canale rifiutato ha lasciato tracce", id)
		}
		// spostamento, cancellazione e canale nuovo nello stesso posto
		m.Fam.PutRec(m, img, 50, m.Fam.Rec(m, img, 3))
		if a, b := m.Fam.Get(m, img, 3), m.Fam.Get(m, img, 50); a == nil || b == nil || *a != *b {
			t.Errorf("%s: canale spostato diverso", id)
		}
		if err := m.Fam.Set(m, img, 3, nil); err != nil || m.Fam.Get(m, img, 3) != nil {
			t.Errorf("%s: canale non cancellato", id)
		}
		if err := m.Fam.Set(m, img, 3, chs[0]); err != nil {
			t.Fatal(err)
		}
		if g := m.Fam.Get(m, img, 3); g == nil || *g != *chs[0] {
			t.Errorf("%s: canale nuovo dopo la cancellazione %+v", id, g)
		}
		// canale simplex modificato: lo scostamento che c'era resta memorizzato
		rec := append([]byte{}, m.Fam.Rec(m, img, 6)...)
		c2 := *chs[2]
		c2.Tx = c2.Rx
		if err := m.Fam.Set(m, img, 6, &c2); err != nil {
			t.Fatal(err)
		}
		c2.Tx = c2.Rx + 1600000
		if err := m.Fam.Set(m, img, 6, &c2); err != nil || !bytes.Equal(rec, m.Fam.Rec(m, img, 6)) {
			t.Errorf("%s: canale tornato com'era ma con dati diversi (%v)", id, err)
		}
		// cavo: lettura e scrittura con la radio simulata
		e := newIcomEmu(o, img)
		icomRun(e, func() {
			got, _, err := m.Fam.Download(ctx, m, "COM9", nop)
			if err != nil || !bytes.Equal(got, img) {
				t.Fatalf("%s: lettura: %v", id, err)
			}
			mod := append([]byte{}, got...)
			if err := m.Fam.Set(m, mod, 20, &Chan{Rx: 145700000, Tx: 145100000, TxTone: "100.0", Scan: true, Name: nm("NEW")}); err != nil {
				t.Fatal(err)
			}
			if _, err := m.Fam.Upload(ctx, m, "COM9", mod, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
			if !bytes.Equal(e.image(), mod) {
				t.Errorf("%s: memoria finale diversa", id)
			}
			if _, err := m.Fam.Upload(ctx, m, "COM9", mod[:len(mod)-1], nop); err == nil {
				t.Errorf("%s: immagine corta accettata", id)
			}
		})
		// un'altra radio Icom collegata: non si legge e non si scrive
		other := icx8xClone
		e = newIcomEmu(other, nil)
		icomRun(e, func() {
			if _, _, err := m.Fam.Download(ctx, m, "COM9", nop); err == nil {
				t.Errorf("%s: letta una radio di un altro modello", id)
			}
			if _, err := m.Fam.Upload(ctx, m, "COM9", img, nop); err == nil || e.writes != 0 {
				t.Errorf("%s: scrittura su un altro modello: %v", id, err)
			}
		})
	}

	// IC-T7H: codifica della frequenza e valori di un canale cancellato
	m := genByID["ic-t7h"]
	img := bytes.Repeat([]byte{0xFF}, 0x03B0)
	if err := m.Fam.Set(m, img, 0, &Chan{Rx: 145612500, Tx: 145012500, TxTone: "67.0", Scan: false}); err != nil {
		t.Fatal(err)
	}
	// 1456 | 1 e 2500/500 = 5 | scostamento 0060 | tono 67,0 = primo dell'elenco (1)
	if got := img[0:8]; !bytes.Equal(got[:5], []byte{0x14, 0x56, 0x15, 0x00, 0x60}) || got[6] != 1 || img[0x270]&0xFA != 0x52 {
		t.Fatalf("IC-T7H: canale % X, byte di stato %02X", got, img[0x270])
	}
	_ = m.Fam.Set(m, img, 0, nil)
	if got := img[0:8]; !bytes.Equal(got[:5], []byte{0, 0, 0, 0x00, 0x60}) || got[6] != 9 || got[7] != 9 || img[0x270]&0x80 == 0 {
		t.Fatalf("IC-T7H: canale cancellato % X", got)
	}
	if !m.inBands(118000000) || !m.inBands(470000000) || m.inBands(200000000) || m.First != 1 || m.Slots != 60 {
		t.Fatal("IC-T7H: limiti")
	}

	// IC-T8A: l'ultima mezza cifra conta i 500 Hz; nome assente = quattro FF
	m = genByID["ic-t8a"]
	img = bytes.Repeat([]byte{0xFF}, 0x07B0)
	if err := m.Fam.Set(m, img, 2, &Chan{Rx: 52525000, Tx: 52525000, Scan: true}); err != nil {
		t.Fatal(err)
	}
	if got := img[16:24]; !bytes.Equal(got, []byte{0x05, 0x25, 0x2A, 0x00, 0x06, 0x00, 9, 9}) || !bytes.Equal(img[0x408:0x40C], []byte{0xFF, 0xFF, 0xFF, 0xFF}) || img[0x602] != 0 {
		t.Fatalf("IC-T8A: canale % X, nome % X, stato %02X", got, img[0x408:0x40C], img[0x602])
	}
	if g := m.Fam.Get(m, img, 2); g == nil || g.Rx != 52525000 || g.Name != "" {
		t.Fatalf("IC-T8A: riletto %+v", g)
	}
	if err := m.Fam.Set(m, img, 2, &Chan{Rx: 145009500, Tx: 145009500}); err == nil {
		t.Error("IC-T8A: frequenza non memorizzabile accettata")
	}

	// IC-Q7A: passi fini, oltre 1 GHz, modo che segue la frequenza, salto «P»
	m = genByID["ic-q7a"]
	f := icq7Fam{}
	img = bytes.Repeat([]byte{0xFF}, 0x07C0)
	for k, c := range []struct {
		hz   int64
		raw  []byte
		mode int64
	}{
		{145012500, []byte{0x14, 0x50, 0x12, 0x80}, 0}, {433006250, []byte{0x43, 0x30, 0x06, 0x80}, 0}, {433018750, []byte{0x43, 0x30, 0x18, 0x80}, 0},
		{1296200000, []byte{0xC9, 0x62, 0x00, 0x00}, 0}, {98500000, []byte{0x09, 0x85, 0x00, 0x00}, 1}, {7100000, []byte{0x00, 0x71, 0x00, 0x00}, 2},
		{121500000, []byte{0x12, 0x15, 0x00, 0x00}, 2},
	} {
		if err := m.Fam.Set(m, img, k, &Chan{Rx: c.hz, Tx: c.hz, Scan: true}); err != nil {
			t.Fatalf("IC-Q7A %d: %v", c.hz, err)
		}
		if got := img[k*8 : k*8+4]; !bytes.Equal(got, c.raw) || f.flg(k).F("mode").Int(img) != c.mode {
			t.Fatalf("IC-Q7A %d: % X, modo %d", c.hz, got, f.flg(k).F("mode").Int(img))
		}
		if g := m.Fam.Get(m, img, k); g == nil || g.Rx != c.hz {
			t.Fatalf("IC-Q7A %d: riletto %+v", c.hz, g)
		}
	}
	// stessa frequenza: il modo scelto sulla radio resta; salto «P» conservato finché resta escluso
	f.flg(0).F("mode").SetInt(img, 3)
	f.flg(0).F("skip").SetBool(img, false)
	f.flg(0).F("pskip").SetBool(img, true)
	f.rec(0).F("tune_step").SetInt(img, 5)
	g := m.Fam.Get(m, img, 0)
	if g == nil || g.Scan {
		t.Fatalf("IC-Q7A: salto P letto %+v", g)
	}
	g.TxTone = "100.0"
	if err := m.Fam.Set(m, img, 0, g); err != nil || f.flg(0).F("mode").Int(img) != 3 || !f.flg(0).F("pskip").Bool(img) || f.flg(0).F("skip").Bool(img) || f.rec(0).F("tune_step").Int(img) != 5 {
		t.Fatalf("IC-Q7A: modo, salto P o passo non conservati (%v)", err)
	}
	g.Scan = true
	_ = m.Fam.Set(m, img, 0, g)
	g.Scan = false
	_ = m.Fam.Set(m, img, 0, g)
	if f.flg(0).F("pskip").Bool(img) || !f.flg(0).F("skip").Bool(img) {
		t.Fatal("IC-Q7A: dopo il rientro nella scansione il salto deve essere quello normale")
	}
	if err := m.Fam.Set(m, img, 9, &Chan{Rx: 145600000, Tx: 135600000}); err == nil || !strings.Contains(err.Error(), "scostamento") {
		t.Errorf("IC-Q7A: scostamento di 10 MHz: %v", err)
	}

	// IC-P7: terzi di Hz conservati, passo automatico conservato, valori del canale cancellato
	m = genByID["ic-p7"]
	fp := icp7Fam{}
	img = bytes.Repeat([]byte{0xFF}, 0x7500)
	if err := m.Fam.Set(m, img, 7, &Chan{Rx: 118008333, Tx: 118008333, Scan: true, Name: "AIR"}); err != nil {
		t.Fatal(err)
	}
	r := fp.rec(7)
	r.F("freq").SetInt(img, 354025000) // 118,008333… MHz esatti
	r.F("tuning_step").SetInt(img, 13)
	g = m.Fam.Get(m, img, 7)
	if g == nil || g.Rx != 118008333 || r.F("mode").Int(img) != 2 {
		t.Fatalf("IC-P7: letto %+v, modo %d", g, r.F("mode").Int(img))
	}
	g.Name = "TWR"
	if err := m.Fam.Set(m, img, 7, g); err != nil || r.F("freq").Int(img) != 354025000 || r.F("tuning_step").Int(img) != 13 || r.F("train_tone").Int(img) != 228 || r.F("unknown1").Int(img) != 63 {
		t.Fatalf("IC-P7: frequenza o passo non conservati (%v)", err)
	}
	_ = m.Fam.Set(m, img, 7, nil)
	if r.F("freq").Int(img) != 15000 || r.F("offset").Int(img) != 479985000 || r.F("name").Str(img) != "      " || !icomBit(fp.mem().F("used"), img, 7) ||
		!icomBit(fp.mem().F("skips"), img, 7) || !icomBit(fp.mem().F("pskips"), img, 7) {
		t.Fatal("IC-P7: canale cancellato diverso da quello di CHIRP")
	}
}
