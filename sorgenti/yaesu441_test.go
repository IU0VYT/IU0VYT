//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: Yaesu VX-1R (drv_vx1.go). Lettura, modifica e copia sono confrontate
// con CHIRP nelle prove TestChirp*, che però modificano solo le memorie del
// gruppo 1; qui le altre zone (limiti di scansione, canali «home», gruppo 2,
// onde medie) confrontate con CHIRP, le regole di ogni zona, la maschera di
// banda, i nomi, le somme di controllo e il cavo (con e senza eco, con gli 00
// prima della conferma, radio che non conferma o che si ferma a metà).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func vx1Blank() []byte {
	img := bytes.Repeat([]byte{0xFF}, vx1Size)
	copy(img[0xE6:], "YAESU")
	for k := 0x10D1; k < 0x10D1+256; k++ {
		img[k] = 0
	}
	vx1Fam{}.fixSums(img)
	return img
}

// bit della maschera di banda accesi per la memoria k del gruppo grp
func vx1MaskBits(img []byte, grp, k int) []int {
	base := 8
	if grp == 1 {
		base = 184
	}
	var out []int
	for b := 0; b < 8; b++ {
		bit := base + b*256 + k
		if img[0x10D1+bit/8]&(1<<(bit%8)) != 0 {
			out = append(out, b)
		}
	}
	return out
}

func TestVX1Model(t *testing.T) {
	m := genByID["ys-vx1"]
	if m == nil {
		t.Fatal("modello mancante")
	}
	f := m.Fam.(vx1Fam)
	if m.Slots != 260 || m.First != 1 || m.NameMax != 6 || m.Baud != 9600 || !m.NoTxOff || !m.HasScan || m.NoChirp || m.RClass != "VX1Radio" ||
		len(m.Powers) != 2 || m.Powers[0].Watts != 0.2 || m.Powers[1].Watts != 1 || genMinSize(m) > vx1Size {
		t.Errorf("dati del modello")
	}
	// le zone: confini e posto nella memoria (CHIRP: 0x0BD1 gruppo 1 a 16 byte,
	// 0x03D1 gruppo 2 a 12 byte, 0x11F1 onde medie a 8 byte)
	for _, c := range []struct{ i, grp, idx, mask, off, size int }{
		{0, 1, 8, 0, 0x0BD1 + 8*16, 16}, {51, 1, 59, 51, 0x0BD1 + 59*16, 16}, {52, 1, 60, 52, 0x0BD1 + 60*16, 16},
		{71, 1, 79, 71, 0x0BD1 + 79*16, 16}, {72, 1, 0, -1, 0x0BD1, 16}, {79, 1, 7, -1, 0x0BD1 + 7*16, 16},
		{80, 2, 8, 0, 0x03D1 + 8*12, 12}, {221, 2, 149, 141, 0x03D1 + 149*12, 12}, {222, 2, 150, 142, 0x03D1 + 150*12, 12},
		{241, 2, 169, 161, 0x03D1 + 169*12, 12}, {242, 2, 0, -1, 0x03D1, 12}, {249, 2, 7, -1, 0x03D1 + 7*12, 12},
		{250, 0, 0, -1, 0x11F1, 8}, {259, 0, 9, -1, 0x11F1 + 9*8, 8},
	} {
		grp, idx, mask := f.zone(c.i)
		r := f.rec(c.i)
		if grp != c.grp || idx != c.idx || mask != c.mask || r.Off != c.off || r.ByteLen() != c.size {
			t.Errorf("posizione %d: gruppo %d, indice %d, maschera %d, indirizzo %#x, %d byte", c.i+1, grp, idx, mask, r.Off, r.ByteLen())
		}
	}
	// le zone non si sovrappongono e stanno prima della somma finale
	seen := map[int]int{}
	for i := 0; i < m.Slots; i++ {
		r := f.rec(i)
		for k := r.Off; k < r.Off+r.ByteLen(); k++ {
			if o, dup := seen[k]; dup {
				t.Fatalf("posizioni %d e %d sullo stesso byte %#x", o+1, i+1, k)
			}
			seen[k] = i
		}
		if r.Off+r.ByteLen() > 0x12D1 || (r.Off < 0x10D1+256 && r.Off+r.ByteLen() > 0x10D1) {
			t.Fatalf("posizione %d fuori posto (%#x)", i+1, r.Off)
		}
	}
	// riconoscimento della memoria
	img := vx1Blank()
	if !f.Match(m, img) || f.Match(m, img[:vx1Size-1]) || f.Match(m, append(append([]byte{}, img...), 0)) {
		t.Errorf("riconoscimento dalla lunghezza")
	}
	bad := append([]byte{}, img...)
	bad[0xE6] = 'X'
	if f.Match(m, bad) {
		t.Errorf("riconosciuta una memoria senza la scritta YAESU")
	}
	// somme di controllo: cambiano con ogni byte delle loro zone, non con gli altri
	a0, b0 := vx1Sums(img)
	for _, c := range []struct {
		at   int
		a, b bool
	}{{0x0000, false, false}, {0x0001, true, true}, {0x009F, true, true}, {0x00A1, false, true}, {0x12D0, false, true}} {
		x := append([]byte{}, img...)
		x[c.at] += 3
		a, b := vx1Sums(x)
		if (a != a0) != c.a || (b != b0) != c.b {
			t.Errorf("somme di controllo con il byte %#x cambiato: prima %v, seconda %v", c.at, a != a0, b != b0)
		}
	}
	// le bande della maschera coprono senza buchi tutto quello che la radio riceve
	for k := 1; k < len(vx1Bands); k++ {
		if vx1Bands[k][0] != vx1Bands[k-1][1] {
			t.Errorf("bande %d e %d non contigue", k-1, k)
		}
	}
	for _, hz := range []int64{76000000, 107999000, 108000000, 145000000, 433000000, 800000000, 998999000, 999000000} {
		if vx1Band(hz) < 0 || !m.inBands(hz) {
			t.Errorf("%d Hz senza banda", hz)
		}
	}
	if vx1Band(75999000) >= 0 || vx1Band(999001000) >= 0 || vx1Band(999000000) != 7 {
		t.Errorf("limiti delle bande")
	}
}

func TestVX1Channels(t *testing.T) {
	m := genByID["ys-vx1"]
	f := m.Fam.(vx1Fam)
	lo, hi := 2, 0 // significato nel programma: 2 bassa, 0 alta
	img := vx1Blank()
	set := func(i int, c *Chan) error { return f.Set(m, img, i, c) }
	same := func(i int, w Chan) {
		t.Helper()
		g := f.Get(m, img, i)
		if g == nil {
			t.Fatalf("posizione %d vuota dopo la scrittura", i+1)
		}
		g.Raw = ""
		if *g != w {
			t.Errorf("posizione %d:\n  scritto %+v\n  riletto %+v", i+1, w, *g)
		}
	}
	// gruppo 1: memorie, limiti di scansione e canali «home» accettano tutto
	for _, i := range []int{0, 51, 52, 71, 72, 79} {
		for _, c := range []Chan{
			{Rx: 145500000, Tx: 145500000, Power: hi, Scan: true, Name: "SIMPLX"},
			{Rx: 145750000, Tx: 145150000, TxTone: "88.5", Power: lo, Scan: true, Name: "R6"},
			{Rx: 430100000, Tx: 431700000, TxTone: "123.0", RxTone: "123.0", Power: hi, Name: "RU"},
			{Rx: 433500000, Tx: 433500000, TxTone: "D754N", RxTone: "D754N", Power: hi, Scan: true},
			{Rx: 145200000, Tx: 435000000, Power: hi, Scan: true, Name: "SAT"}, // trasmissione separata
			{Rx: 88000000, Tx: 88000000, Power: hi, Scan: true, Name: "FM"},
			{Rx: 118100000, Tx: 118100000, Power: hi, Scan: true, Name: "TWR"},
			{Rx: 145337500, Tx: 145337500, Power: hi, Scan: true, Name: "12K5"}, // passo da 12,5 kHz
		} {
			if err := set(i, &c); err != nil {
				t.Fatalf("posizione %d, %+v: %v", i+1, c, err)
			}
			same(i, c)
			_, _, mask := f.zone(i)
			bits := []int(nil)
			if mask >= 0 {
				bits = vx1MaskBits(img, 1, mask)
			}
			if want := vx1Band(c.Rx); (mask >= 0 && (len(bits) != 1 || bits[0] != want)) || (mask < 0 && bits != nil) {
				t.Errorf("posizione %d a %d Hz: bit della maschera di banda %v (banda %d)", i+1, c.Rx, bits, want)
			}
		}
	}
	// il modo segue la frequenza quando cambia, resta quando non cambia
	r := f.rec(3)
	for _, c := range []struct {
		hz   int64
		mode int64
	}{{100000000, 1}, {120000000, 2}, {145000000, 0}, {76000000, 1}, {136975000, 2}, {137000000, 0}} {
		if err := set(3, &Chan{Rx: c.hz, Tx: c.hz, Power: hi, Scan: true}); err != nil {
			t.Fatal(err)
		}
		if got := r.F("mode").Int(img); got != c.mode {
			t.Errorf("%d Hz: modo %d, atteso %d", c.hz, got, c.mode)
		}
	}
	r.F("mode").SetInt(img, 2) // AM scelto sulla radio a 137 MHz
	if err := set(3, &Chan{Rx: 137000000, Tx: 137000000, Power: lo, Scan: true, Name: "AM"}); err != nil {
		t.Fatal(err)
	}
	if r.F("mode").Int(img) != 2 {
		t.Errorf("modo scelto sulla radio perso modificando il nome")
	}
	// combinazioni di toni che la radio non ha
	for _, c := range []Chan{
		{Rx: 145000000, Tx: 145000000, RxTone: "88.5"},
		{Rx: 145000000, Tx: 145000000, TxTone: "88.5", RxTone: "100.0"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023N", RxTone: "D754N"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023I", RxTone: "D023I"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023N"},
		{Rx: 145000000, Tx: 145000000, TxTone: "88.5", RxTone: "D023N"},
	} {
		before := append([]byte{}, img...)
		c.Power, c.Scan = hi, true
		if err := set(1, &c); err == nil {
			t.Errorf("toni %q/%q accettati", c.TxTone, c.RxTone)
		}
		if !bytes.Equal(before, img) {
			t.Fatalf("toni %q/%q rifiutati ma la memoria è cambiata", c.TxTone, c.RxTone)
		}
	}
	// frequenze fuori dalla radio
	for _, hz := range []int64{75000000, 1000000000, 30000000, 145001234} {
		before := append([]byte{}, img...)
		if err := set(1, &Chan{Rx: hz, Tx: hz, Power: hi, Scan: true}); err == nil {
			t.Errorf("%d Hz accettati", hz)
		}
		if !bytes.Equal(before, img) {
			t.Fatalf("%d Hz rifiutati ma la memoria è cambiata", hz)
		}
	}
	// gruppo 2: niente toni, solo lo scostamento normale della banda
	for _, i := range []int{80, 221, 222, 241, 242, 249} {
		for _, c := range []Chan{
			{Rx: 145500000, Tx: 145500000, Power: hi, Scan: true, Name: "V"},
			{Rx: 145750000, Tx: 145150000, Power: lo, Scan: true, Name: "R6"},
			{Rx: 145000000, Tx: 145600000, Power: hi, Name: "PIU"},
			{Rx: 439000000, Tx: 434000000, Power: hi, Scan: true, Name: "U-"},
			{Rx: 430000000, Tx: 435000000, Power: hi, Scan: true, Name: "U+"},
			{Rx: 99900000, Tx: 99900000, Power: hi, Scan: true, Name: "RADIO"},
		} {
			if err := set(i, &c); err != nil {
				t.Fatalf("posizione %d, %+v: %v", i+1, c, err)
			}
			same(i, c)
			_, _, mask := f.zone(i)
			if mask >= 0 {
				if bits := vx1MaskBits(img, 2, mask); len(bits) != 1 || bits[0] != vx1Band(c.Rx) {
					t.Errorf("posizione %d a %d Hz: bit della maschera di banda %v", i+1, c.Rx, bits)
				}
			}
		}
		for _, c := range []Chan{
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5"},
			{Rx: 145500000, Tx: 145500000, TxTone: "D023N", RxTone: "D023N"},
			{Rx: 145750000, Tx: 145250000},                  // 500 kHz
			{Rx: 430000000, Tx: 431600000},                  // 1,6 MHz
			{Rx: 145200000, Tx: 435000000},                  // trasmissione separata
			{Rx: 120000000, Tx: 120600000},                  // banda senza scostamento normale
			{Rx: 145750000, Tx: 145150000, RxTone: "D023N"}, // tono solo in ricezione
			{Rx: 800000, Tx: 800000},                        // onde medie
			{Rx: 1000000000, Tx: 1000000000},                // fuori dalla radio
		} {
			before := append([]byte{}, img...)
			c.Power, c.Scan = hi, true
			if err := set(i, &c); err == nil {
				t.Errorf("posizione %d: %+v accettato", i+1, c)
			}
			if !bytes.Equal(before, img) {
				t.Fatalf("posizione %d: %+v rifiutato ma la memoria è cambiata", i+1, c)
			}
		}
	}
	// il tipo di tono già presente in una memoria del gruppo 2 resta com'è
	r2 := f.rec(100)
	if err := set(100, &Chan{Rx: 145000000, Tx: 145000000, Power: hi, Scan: true}); err != nil {
		t.Fatal(err)
	}
	r2.F("selcal").SetInt(img, 2)
	r2.F("clk_shift").SetBool(img, true)
	r2.F("step").SetInt(img, 3)
	if err := set(100, &Chan{Rx: 145025000, Tx: 145025000, Power: lo, Scan: true, Name: "X"}); err != nil {
		t.Fatal(err)
	}
	if r2.F("selcal").Int(img) != 2 || !r2.F("clk_shift").Bool(img) || r2.F("step").Int(img) != 3 {
		t.Errorf("gruppo 2: tipo di tono, spostamento del clock o passo persi con la modifica")
	}
	// onde medie: solo nelle posizioni 251-260, e lì nient'altro
	for _, i := range []int{250, 259} {
		for _, hz := range []int64{500000, 999000, 1296000, 1700000} {
			c := Chan{Rx: hz, Tx: hz, Power: hi, Scan: true, Name: "OM"}
			if err := set(i, &c); err != nil {
				t.Fatalf("posizione %d a %d Hz: %v", i+1, hz, err)
			}
			g := f.Get(m, img, i)
			// un byte per 1,2 MHz: la radio arrotonda al passo più vicino (4,7 kHz)
			if g == nil || g.Rx-hz > 2400 || hz-g.Rx > 2400 || g.Tx != g.Rx || g.Name != "OM" {
				t.Errorf("posizione %d a %d Hz: riletto %+v", i+1, hz, g)
			}
		}
		for _, c := range []Chan{
			{Rx: 145000000, Tx: 145000000}, {Rx: 1701000, Tx: 1701000}, {Rx: 499000, Tx: 499000},
			{Rx: 900000, Tx: 1000000}, {Rx: 900000, Tx: 900000, TxTone: "88.5"}, {Rx: 900000, Tx: 900000, RxTone: "D023N"},
		} {
			before := append([]byte{}, img...)
			c.Power, c.Scan = hi, true
			if err := set(i, &c); err == nil {
				t.Errorf("posizione %d: %+v accettato", i+1, c)
			}
			if !bytes.Equal(before, img) {
				t.Fatalf("posizione %d: %+v rifiutato ma la memoria è cambiata", i+1, c)
			}
		}
	}
	// dal programma: la frequenza scritta è la più vicina che la radio ha, con una nota
	chs := genChans(m, img)
	chs[251] = &Chan{Rx: 900000, Tx: 900000, Power: hi, Scan: true, Name: "RAI"}
	rep := &Report{}
	out, err := genWrite(m, img, chs, rep)
	if err != nil {
		t.Fatalf("onde medie dal programma: %v", err)
	}
	if g := f.Get(m, out, 251); g == nil || g.Rx != vx1BCHz(vx1BCByte(900000)) || g.Rx == 900000 || len(rep.Notes) != 1 || !strings.Contains(rep.Notes[0], "0.901575") {
		t.Errorf("onde medie dal programma: %+v, note %v", g, rep.Notes)
	}
	// e una frequenza già al passo della radio si scrive senza note
	chs[252] = &Chan{Rx: vx1BCHz(100), Tx: vx1BCHz(100), Power: hi, Scan: true}
	chs[251] = f.Get(m, out, 251)
	rep = &Report{}
	if _, err := genWrite(m, out, chs, rep); err != nil || len(rep.Notes) != 0 {
		t.Errorf("onde medie al passo della radio: %v, note %v", err, rep.Notes)
	}
	// ogni byte delle onde medie torna uguale dopo lettura e scrittura
	for b := 0; b <= 0xFE; b++ {
		if got := vx1BCByte(vx1BCHz(byte(b))); got != byte(b) {
			t.Fatalf("onde medie: byte %#x riscritto %#x", b, got)
		}
	}
	if vx1BCHz(0) != 500000 || vx1BCHz(0xFE) != 1700000 {
		t.Errorf("onde medie: estremi %d e %d Hz", vx1BCHz(0), vx1BCHz(0xFE))
	}
	// nomi: tutti i caratteri della radio, maiuscole, taglio a 6
	for _, name := range []string{"0AZ9 (", ")+-=*/", "|?%&_$", "\\<>", "A"} {
		if err := set(10, &Chan{Rx: 145000000, Tx: 145000000, Power: hi, Scan: true, Name: name}); err != nil {
			t.Fatal(err)
		}
		if g := f.Get(m, img, 10); g.Name != strings.TrimRight(name, " ") {
			t.Errorf("nome %q riletto %q", name, g.Name)
		}
	}
	for _, ch := range []byte(vx1Charset) {
		if strings.IndexByte(vx1Chars, ch) < 0 {
			t.Errorf("carattere %q non presente nella tabella della radio", ch)
		}
	}
	if err := set(10, &Chan{Rx: 145000000, Tx: 145000000, Power: hi, Scan: true, Name: "abcdefgh"}); err != nil {
		t.Fatal(err)
	}
	if g := f.Get(m, img, 10); g.Name != "ABCDEF" {
		t.Errorf("nome lungo in minuscole riletto %q", g.Name)
	}
	// cancellazione: solo il primo byte e la maschera di banda; i vicini restano
	for _, i := range []int{0, 60, 75, 90, 230, 245, 255} {
		grp, _, mask := f.zone(i)
		hz := int64(145000000)
		if grp == 0 {
			hz = 900000
		}
		for _, j := range []int{i - 1, i, i + 1} {
			if g, _, _ := f.zone(j); g == grp {
				if err := set(j, &Chan{Rx: hz, Tx: hz, Power: hi, Scan: true, Name: "N"}); err != nil {
					t.Fatal(err)
				}
			}
		}
		before := append([]byte{}, img...)
		if err := set(i, nil); err != nil {
			t.Fatal(err)
		}
		if f.Get(m, img, i) != nil {
			t.Errorf("posizione %d cancellata ma ancora presente", i+1)
		}
		rr := f.rec(i)
		changed := 0
		for k := range img {
			if img[k] != before[k] {
				changed++
				if k != rr.Off && (k < 0x10D1 || k >= 0x10D1+256) {
					t.Errorf("posizione %d: la cancellazione ha cambiato il byte %#x", i+1, k)
				}
			}
		}
		if mask >= 0 {
			if bits := vx1MaskBits(img, grp, mask); len(bits) != 0 || changed != 2 {
				t.Errorf("posizione %d cancellata: bit di banda %v, %d byte cambiati", i+1, bits, changed)
			}
			for _, d := range []int{-1, 1} {
				if g, _, mk := f.zone(i + d); g == grp && mk >= 0 {
					if bits := vx1MaskBits(img, grp, mk); len(bits) != 1 {
						t.Errorf("posizione %d cancellata: bit di banda del vicino %v", i+1, bits)
					}
				}
			}
		} else if changed != 1 {
			t.Errorf("posizione %d cancellata: %d byte cambiati", i+1, changed)
		}
	}
	// i bit di banda dei due gruppi non si toccano fra loro, né toccano gli otto
	// byte di banda di ogni blocco
	img = vx1Blank()
	for i := 0; i < 250; i++ {
		grp, _, mask := f.zone(i)
		if mask < 0 {
			continue
		}
		hz := vx1Bands[i%8][0] + 1000000
		if err := set(i, &Chan{Rx: hz, Tx: hz, Power: hi, Scan: true}); err != nil {
			t.Fatalf("posizione %d a %d Hz: %v", i+1, hz, err)
		}
		_ = grp
	}
	for i := 0; i < 250; i++ {
		grp, _, mask := f.zone(i)
		if mask < 0 {
			continue
		}
		if bits := vx1MaskBits(img, grp, mask); len(bits) != 1 || bits[0] != i%8 {
			t.Fatalf("posizione %d: bit di banda %v, atteso %d", i+1, bits, i%8)
		}
	}
	for b := 0; b < 8; b++ {
		if img[0x10D1+b*32] != 0 || img[0x10D1+b*32+22] != 0 {
			t.Errorf("blocco %d della maschera: byte di banda cambiati (%#x, %#x)", b, img[0x10D1+b*32], img[0x10D1+b*32+22])
		}
	}
}

// le zone che la prova generale non modifica, confrontate con CHIRP
func TestVX1ChirpZones(t *testing.T) {
	chirpDir(t)
	m := genByID["ys-vx1"]
	f := m.Fam.(vx1Fam)
	lo, hi := 2, 0 // significato nel programma: 2 bassa, 0 alta
	p := chirpImageFor(t, m)
	base := loadImg(t, p)
	type ed struct {
		n int
		c *Chan
	}
	edits := []ed{
		// limiti di scansione e «home» del gruppo 1
		{52, &Chan{Rx: 144000000, Tx: 144000000, Power: hi, Scan: true, Name: "L1"}},
		{53, &Chan{Rx: 146000000, Tx: 145400000, TxTone: "100.0", Power: lo, Scan: true, Name: "U1"}},
		{71, &Chan{Rx: 439975000, Tx: 439975000, TxTone: "D071N", RxTone: "D071N", Power: hi, Name: "U10"}},
		{72, &Chan{Rx: 88500000, Tx: 88500000, Power: hi, Scan: true, Name: "HOME1"}},
		{74, &Chan{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "88.5", Power: hi, Scan: true, Name: "HOME3"}},
		{79, &Chan{Rx: 850000000, Tx: 850000000, Power: lo, Scan: true}},
		// gruppo 2: memorie, limiti e «home»
		{80, &Chan{Rx: 145500000, Tx: 145500000, Power: hi, Scan: true, Name: "G2-1"}},
		{81, &Chan{Rx: 145762500, Tx: 145162500, Power: lo, Scan: true, Name: "R6A"}},
		{150, &Chan{Rx: 430250000, Tx: 435250000, Power: hi, Name: "UP"}},
		{221, &Chan{Rx: 121500000, Tx: 121500000, Power: hi, Scan: true, Name: "EMG"}},
		{222, &Chan{Rx: 430000000, Tx: 430000000, Power: hi, Scan: true, Name: "L1"}},
		{241, &Chan{Rx: 440000000, Tx: 435000000, Power: hi, Scan: true, Name: "U10"}},
		{242, &Chan{Rx: 98100000, Tx: 98100000, Power: hi, Scan: true, Name: "H1"}},
		{249, &Chan{Rx: 900000000, Tx: 900000000, Power: lo, Scan: true, Name: "H8"}},
		// onde medie
		{250, &Chan{Rx: 900000, Tx: 900000, Power: hi, Scan: true, Name: "RAI1"}},
		{259, &Chan{Rx: 1575000, Tx: 1575000, Power: hi, Scan: true}},
		// cancellazioni in ogni zona
		{60, nil}, {73, nil}, {100, nil}, {230, nil}, {243, nil}, {255, nil},
	}
	// prima le posizioni da cancellare si riempiono (qui e nell'immagine di partenza)
	for _, e := range edits {
		if e.c == nil {
			hz := int64(145000000)
			if e.n >= 250 {
				hz = 1000000
			}
			if err := f.Set(m, base, e.n, &Chan{Rx: hz, Tx: hz, Power: hi, Scan: true, Name: "DEL"}); err != nil {
				t.Fatal(err)
			}
		}
	}
	// CHIRP parte dai due bit rimasti nella posizione vuota («mostra il nome» e
	// spostamento del clock: quelli del canale cancellato, o tutti a 1 nella
	// memoria mai usata); qui un canale nuovo mostra il nome se ne ha uno e non
	// sposta il clock, che è il valore che il driver di CHIRP dichiara come
	// predefinito. Per il confronto i bit rimasti si mettono uguali nei due
	for _, e := range edits {
		if e.c != nil && f.Get(m, base, e.n) == nil && e.n < 250 {
			r := f.rec(e.n)
			r.F("name_flag").SetBool(base, e.c.Name != "")
			r.F("clk_shift").SetBool(base, false)
			if f.Get(m, base, e.n) != nil {
				t.Fatalf("posizione %d: preparazione sbagliata", e.n+1)
			}
		}
	}
	tmp := t.TempDir()
	bp := filepath.Join(tmp, "base.img")
	if err := os.WriteFile(bp, base, 0o644); err != nil {
		t.Fatal(err)
	}
	out := append([]byte{}, base...)
	var js []map[string]interface{}
	for _, e := range edits {
		var cj interface{}
		if e.c != nil {
			// qui il modo segue la frequenza (il programma non ha un campo per sceglierlo)
			mode := "FM"
			switch vx1Band(e.c.Rx) {
			case 0:
				mode = "WFM"
			case 1:
				mode = "AM"
			}
			cj = map[string]interface{}{"rx": e.c.Rx, "tx": e.c.Tx, "txOff": false, "txTone": e.c.TxTone, "rxTone": e.c.RxTone,
				"power": chirpPowerOf(m, out, e.n, e.c.Power), "mode": mode, "scan": e.c.Scan, "name": e.c.Name}
		}
		js = append(js, map[string]interface{}{"n": e.n + m.First, "chan": cj})
		if err := f.Set(m, out, e.n, e.c); err != nil {
			t.Fatalf("posizione %d: %v", e.n+1, err)
		}
	}
	ej, _ := json.Marshal(js)
	ef, of := filepath.Join(tmp, "edits.json"), filepath.Join(tmp, "out.img")
	_ = os.WriteFile(ef, ej, 0o644)
	res := chirpRef(t, "encode", m.RClass, bp, ef, of)
	var rr struct {
		Errors [][]interface{} `json:"errors"`
	}
	_ = json.Unmarshal(res, &rr)
	for _, e := range rr.Errors {
		t.Errorf("CHIRP segnala per la posizione %v: %v", e[0], e[1])
	}
	ref, err := os.ReadFile(of)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ref, out) {
		t.Errorf("immagine diversa da quella di CHIRP: %s", diffRanges(ref, out))
	}
	// e la lettura di tutte le 260 posizioni dell'immagine modificata coincide
	mp := filepath.Join(tmp, "mod.img")
	if err := os.WriteFile(mp, out, 0o644); err != nil {
		t.Fatal(err)
	}
	dref := chirpDecode(t, m, mp)
	if len(dref) != m.Slots {
		t.Fatalf("CHIRP ha %d posizioni, qui %d", len(dref), m.Slots)
	}
	used := 0
	for i := 0; i < m.Slots; i++ {
		g := genGet(m, out, i)
		pi := -1
		if g != nil {
			used++
			pi = chirpPowerOf(m, out, i, g.Power)
		}
		if d := cmpChanP(m, i+m.First, g, dref[i], pi); d != "" {
			t.Error(d)
		}
	}
	if used < len(edits)-6 {
		t.Errorf("solo %d canali presenti dopo le modifiche", used)
	}
}

func TestVX1Cable(t *testing.T) {
	m := genByID["ys-vx1"]
	f := m.Fam.(vx1Fam)
	ctx := context.Background()
	nop := func(int, int, string) {}
	base := vx1Blank()
	for i, hz := range []int64{145500000, 433500000, 88000000} {
		if err := f.Set(m, base, i, &Chan{Rx: hz, Tx: hz, Power: 0, Scan: true, Name: "CH"}); err != nil {
			t.Fatal(err)
		}
	}
	f.fixSums(base)
	mk := func() *vx1Emu { return f.emu(m, base).(*vx1Emu) }
	// lettura: niente inviato alla radio, immagine identica
	e := mk()
	var got []byte
	err, tx := alSession(t, e, nil, func() error {
		var err error
		got, _, err = f.Download(ctx, m, "EMU", nop)
		return err
	})
	if err != nil || !bytes.Equal(got, base) || len(tx) != 0 {
		t.Fatalf("lettura: %v, %d byte inviati alla radio, immagine uguale: %v", err, len(tx), bytes.Equal(got, base))
	}
	// la radio si ferma a metà: errore con il numero dei byte, niente immagine
	e = mk()
	e.cut = 2000
	err, _ = alSession(t, e, nil, func() error {
		var err error
		got, _, err = f.Download(ctx, m, "EMU", nop)
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "2000") || got != nil {
		t.Errorf("lettura interrotta: %v", err)
	}
	// somma di controllo errata e dati di un'altra radio
	for _, c := range []struct {
		at   int
		want string
	}{{0x0050, "somma"}, {0x0500, "somma"}, {0x12D1, "somma"}} {
		e = mk()
		e.img[c.at] ^= 0x10
		err, _ = alSession(t, e, nil, func() error {
			_, _, err := f.Download(ctx, m, "EMU", nop)
			return err
		})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("byte %#x rovinato: %v", c.at, err)
		}
	}
	e = mk()
	copy(e.img[0xE6:], "XXXXX")
	f.fixSums(e.img)
	err, _ = alSession(t, e, nil, func() error {
		_, _, err := f.Download(ctx, m, "EMU", nop)
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "VX-1") {
		t.Errorf("dati di un'altra radio: %v", err)
	}
	// scrittura: con e senza eco, con e senza gli 00 prima della conferma
	mod := append([]byte{}, base...)
	if err := f.Set(m, mod, 5, &Chan{Rx: 146000000, Tx: 145400000, TxTone: "88.5", Power: 2, Scan: true, Name: "NUOVO"}); err != nil {
		t.Fatal(err)
	}
	mod[0x00A0], mod[0x12D1] = 0x55, 0xAA // somme sbagliate nel file: si ricalcolano
	want := append([]byte{}, mod...)
	f.fixSums(want)
	var sent [][]byte
	for _, noEcho := range []bool{false, true} {
		for _, zeros := range []int{0, 1, 5} {
			e = mk()
			e.noEcho, e.zeros = noEcho, zeros
			err, tx = alSession(t, e, nil, func() error {
				_, err := f.Upload(ctx, m, "EMU", mod, nop)
				return err
			})
			if err != nil {
				t.Fatalf("scrittura (senza eco: %v, %d zeri): %v", noEcho, zeros, err)
			}
			if e.writes != 1 || !bytes.Equal(e.image(), want) {
				t.Fatalf("scrittura (senza eco: %v, %d zeri): memoria finale diversa", noEcho, zeros)
			}
			sent = append(sent, tx)
		}
	}
	for _, s := range sent {
		if !bytes.Equal(s, want) {
			t.Fatalf("scrittura: inviati %d byte diversi dalla memoria con le somme giuste", len(s))
		}
	}
	if mod[0x00A0] != 0x55 {
		t.Errorf("la scrittura ha cambiato l'immagine in memoria")
	}
	// la radio non conferma (non è in «CLN IN»): errore, con e senza eco
	for _, noEcho := range []bool{false, true} {
		e = mk()
		e.noEcho, e.nak = noEcho, true
		err, _ = alSession(t, e, nil, func() error {
			_, err := f.Upload(ctx, m, "EMU", mod, nop)
			return err
		})
		if err == nil || !strings.Contains(err.Error(), "non conferma") || e.writes != 0 || !bytes.Equal(e.image(), base) {
			t.Errorf("radio che non conferma (senza eco: %v): %v", noEcho, err)
		}
	}
	// immagine di un'altra radio: niente inviato
	err, tx = alSession(t, mk(), nil, func() error {
		_, err := f.Upload(ctx, m, "EMU", mod[:vx1Size-2], nop)
		return err
	})
	if err == nil || len(tx) != 0 {
		t.Errorf("immagine troppo corta: %v, %d byte inviati", err, len(tx))
	}
	// risposta che non è la conferma
	for _, c := range []struct {
		img, got []byte
		rest     []byte
		pending  bool
	}{
		{[]byte{1, 2, 3}, nil, nil, true},
		{[]byte{1, 2, 3}, []byte{1, 2}, []byte{1, 2}, true},
		{[]byte{1, 2, 3}, []byte{1, 2, 3}, nil, false},
		{[]byte{1, 2, 3}, []byte{1, 2, 3, 0x21}, []byte{0x21}, false},
		{[]byte{1, 2, 3}, []byte{1, 2, 3, 0, 0, 0x21}, []byte{0x21}, false},
		{[]byte{1, 2, 3}, []byte{0x21}, []byte{0x21}, false},
		{[]byte{1, 2, 3}, []byte{0, 0x21}, []byte{0x21}, false},
		{[]byte{1, 2, 3}, []byte{0, 0}, nil, false},
		{[]byte{1, 2, 3}, []byte{1, 2, 3, 0x15}, []byte{0x15}, false},
		{[]byte{0, 0, 3}, []byte{0, 0}, nil, true},
		{[]byte{0x21, 2, 3}, []byte{0x21}, []byte{0x21}, true},
	} {
		rest, pending := vx1Answer(c.img, c.got)
		if !bytes.Equal(rest, c.rest) || pending != c.pending {
			t.Errorf("risposta %v dopo %v: resto %v, in attesa %v", c.got, c.img, rest, pending)
		}
	}
}

// conversione da un'altra radio: si riempiono solo le memorie normali; limiti di
// scansione e canali «home» restano quelli della radio; i canali con il tono o
// con uno scostamento particolare vanno nel gruppo 1 finché c'è posto
func TestVX1Convert(t *testing.T) {
	m := genByID["ys-vx1"]
	f := m.Fam.(vx1Fam)
	base := vx1Blank()
	keep := map[int]Chan{
		60:  {Rx: 144000000, Tx: 144000000, Power: 0, Scan: true, Name: "L5"},
		75:  {Rx: 145500000, Tx: 145500000, Power: 0, Scan: true, Name: "HOME"},
		230: {Rx: 430000000, Tx: 430000000, Power: 0, Scan: true, Name: "L5B"},
		245: {Rx: 433500000, Tx: 433500000, Power: 0, Scan: true, Name: "HOMEB"},
	}
	for i, c := range keep {
		c := c
		if err := f.Set(m, base, i, &c); err != nil {
			t.Fatal(err)
		}
	}
	var src []*Chan
	for k := 0; k < 120; k++ {
		c := &Chan{Rx: 145000000 + int64(k)*12500, Tx: 145000000 + int64(k)*12500, Power: 0, Scan: true, Name: fmt.Sprintf("C%03d", k)}
		switch k % 4 {
		case 1:
			c.TxTone = "88.5" // con il tono: solo nel gruppo 1
		case 2:
			c.Tx = c.Rx - 600000 // scostamento normale: va bene dappertutto
		case 3:
			c.Tx = c.Rx - 1000000 // scostamento particolare: solo nel gruppo 1
		}
		src = append(src, c)
	}
	src = append(src, &Chan{Rx: 900000, Tx: 900000, Power: 0, Scan: true, Name: "OM"})
	out, _, rep, err := uniConvert(src, m.ID, makeImg(m, base, map[string]interface{}{}), false, nil)
	if err != nil {
		t.Fatal(err)
	}
	back, err := readAny(out, "uscita.img", m.ID)
	if err != nil {
		t.Fatal(err)
	}
	got := back.Channels
	// le posizioni speciali restano com'erano
	for i := 52; i < 80; i++ {
		for _, j := range []int{i, i + 170} {
			w, has := keep[j]
			if g := got[j]; has != (g != nil) || (has && (g.Rx != w.Rx || g.Name != w.Name)) {
				t.Errorf("posizione speciale %d: %+v", j+1, g)
			}
		}
	}
	// nel gruppo 2 niente toni né scostamenti particolari; le onde medie al loro posto
	n1, n2, tone2 := 0, 0, 0
	for i, c := range got {
		if c == nil || i >= 250 || f.placeSkip(m, i) {
			continue
		}
		if i < 52 {
			n1++
			continue
		}
		n2++
		if d := c.Tx - c.Rx; c.TxTone != "" || c.RxTone != "" || (d != 0 && d != -600000) {
			tone2++
		}
	}
	if n1 != 52 || tone2 != 0 || got[250] == nil || got[250].Name != "OM" {
		t.Errorf("gruppo 1: %d canali; gruppo 2: %d con tono o scostamento particolare; onde medie: %+v", n1, tone2, got[250])
	}
	// i canali che non trovano posto sono esclusi con un avviso, non fermano la conversione
	placed := 0
	for i, c := range got {
		if c != nil && i < 250 && !f.placeSkip(m, i) {
			placed++
		}
	}
	excluded := 0
	for _, w := range rep.Warnings {
		if strings.Contains(w, "escluso") {
			excluded++
		}
	}
	if placed+excluded != 120 || excluded == 0 || n2 == 0 {
		t.Errorf("%d canali messi (%d nel gruppo 2), %d esclusi su 120: %v", placed, n2, excluded, rep.Warnings[:min(len(rep.Warnings), 3)])
	}
}
