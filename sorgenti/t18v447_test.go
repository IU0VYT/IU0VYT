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
	"testing"
)

// 4.47, Retevis B3H, B63H («C170RAM»), RB647 («C90GRAM»), RT68H e RT668H
// («H32GRAM»), NR10 e NR610 («KY666AM»), RB66S e RB666S («HZ3GRAh»): blocchi da 8 byte, canali da 16 byte da 0x0010 come nella famiglia
// della T18. La memoria di partenza è scritta a mano dalla scheda del protocollo;
// poi lettura e scrittura con la radio simulata (dialogo byte per byte), password
// di programmazione, canale «cifrato», ultimi tre byte del canale conservati.
func TestT18Vend447(t *testing.T) {
	nop := func(int, int, string) {}
	for _, x := range []struct{ id, cmd string }{
		{"rt-b3h", "C170RAM"}, {"rt-b63h", "C170RAM"}, {"rt-rb647", "C90GRAM"}, {"rt-rt68h", "H32GRAM"}, {"rt-rt668h", "H32GRAM"},
		{"rt-nr10", "KY666AM"}, {"rt-nr610", "KY666AM"}, {"rt-rb66s", "HZ3GRAh"}, {"rt-rb666s", "HZ3GRAh"},
	} {
		id := x.id
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := t18o(m)
		sp := o.spec
		if !m.NoChirp || (m.Slots != 16 && id != "rt-rb66s") || (m.Slots != 22 && id == "rt-rb66s") || !o.vend || !o.reserved || o.lock != "" || o.power != "rw" || o.reset != 0xFB ||
			string(sp.magic) != "\x02"+x.cmd || !sp.anyID || !sp.pwStep || sp.block != 8 || sp.memsize != 0x3F0 || !sp.noAck || sp.echo ||
			string(sp.exit) != "b" {
			t.Fatalf("%s: modello con altri parametri", id)
		}
		radio := bytes.Repeat([]byte{0xFF}, 0x3F0)
		// canale 1 (0x0010): 462,5625 MHz in BCD, tono 67,0, opzioni EF (in scansione, alta, stretta), coda 25 51 43
		copy(radio[0x10:], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0x70, 0x06, 0x70, 0x06, 0xEF, 0x25, 0x51, 0x43})
		// canale 9 (0x0090): 446,10625 MHz, D026N, opzioni E7 (bassa)
		copy(radio[0x90:], []byte{0x25, 0x06, 0x61, 0x44, 0x25, 0x06, 0x61, 0x44, 0x26, 0x80, 0x26, 0x80, 0xE7, 0xFF, 0xFF, 0xFF})
		// canale 16 (0x0100): 462,7250 MHz, D462I in ricezione, trasmissione spenta, fuori dalla scansione, larga
		copy(radio[0x100:], []byte{0x00, 0x25, 0x27, 0x46, 0xFF, 0xFF, 0xFF, 0xFF, 0x62, 0xC4, 0xFF, 0xFF, 0xFB, 0xFF, 0xFF, 0xFF})
		for k := 0x3C0; k < 0x3D0; k++ { // impostazioni: valori qualsiasi, da ritrovare intatti
			radio[k] = byte(k)
		}
		type want struct {
			i              int
			rx             int64
			off            bool
			rt, tt         string
			scan, nar, low bool
		}
		for _, w := range []want{
			{0, 462562500, false, "67.0", "67.0", true, true, false},
			{8, 446106250, false, "D026N", "D026N", true, true, true},
			{15, 462725000, true, "D462I", "", false, false, false},
		} {
			g := m.Fam.Get(m, radio, w.i)
			if g == nil || g.Rx != w.rx || g.Tx != w.rx || g.TxOff != w.off || g.RxTone != w.rt || g.TxTone != w.tt || g.Scan != w.scan ||
				g.Narrow != w.nar || (m.powerIndex(g.Power) == 1) != w.low {
				t.Fatalf("%s canale %d: %+v", id, w.i+1, g)
			}
		}
		if m.Fam.Get(m, radio, 1) != nil {
			t.Fatalf("%s: canale libero letto come occupato", id)
		}

		e := m.Fam.(t18Fam).emu(m, radio)
		var tx, rx, img []byte
		emuRunRec(e, &tx, &rx, func() {
			var err error
			if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
				t.Fatalf("%s: lettura: %v", id, err)
			}
		})
		if !bytes.Equal(img, radio) {
			t.Fatalf("%s: lettura diversa dalla memoria della radio", id)
		}
		var wtx, wrx []byte
		step := func(send, reply []byte) {
			wtx = append(wtx, send...)
			wrx = append(wrx, reply...)
		}
		hello := func() {
			step(append([]byte{0x02}, x.cmd...), []byte{0x06})
			step([]byte{0x02}, []byte("SMP558\x00\x00"))
			step([]byte{0x06}, []byte{0x06})
			step([]byte{0x05}, radio[0x3B0:0x3B6])
			step([]byte{0x06}, []byte{0x06})
		}
		hello()
		for a := 0; a < 0x3F0; a += 8 {
			step([]byte{'R', byte(a >> 8), byte(a), 8}, append([]byte{'W', byte(a >> 8), byte(a), 8}, radio[a:a+8]...))
		}
		wtx = append(wtx, 'b')
		if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
			t.Fatalf("%s: dialogo di lettura diverso dalla scheda (%d/%d byte mandati, %d/%d ricevuti)", id, len(tx), len(wtx), len(rx), len(wrx))
		}

		// canale nuovo: opzioni del programma (FB) con scansione, potenza e banda del canale; coda FF
		nc := &Chan{Rx: 446093750, Tx: 446093750, Narrow: true, Scan: true, Power: 2, RxTone: "D023N", TxTone: "123.0"}
		if err := m.Fam.Set(m, img, 2, nc); err != nil {
			t.Fatal(err)
		}
		wrec := []byte{0x75, 0x93, 0x60, 0x44, 0x75, 0x93, 0x60, 0x44, 0x23, 0x80, 0x30, 0x12, 0xE7, 0xFF, 0xFF, 0xFF}
		if !bytes.Equal(img[0x30:0x40], wrec) {
			t.Fatalf("%s: canale nuovo scritto come % X, atteso % X", id, img[0x30:0x40], wrec)
		}
		// canale modificato e canale cancellato: gli ultimi tre byte restano
		mod := *m.Fam.Get(m, img, 0)
		mod.RxTone, mod.TxTone, mod.Scan = "88.5", "88.5", false
		if err := m.Fam.Set(m, img, 0, &mod); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(img[0x10:0x20], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0x85, 0x08, 0x85, 0x08, 0xFF, 0x25, 0x51, 0x43}) {
			t.Fatalf("%s: canale modificato scritto come % X", id, img[0x10:0x20])
		}
		if err := m.Fam.Set(m, img, 8, nil); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, img, 0, nil); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(img[0x10:0x20], append(bytes.Repeat([]byte{0xFF}, 13), 0x25, 0x51, 0x43)) || m.Fam.Get(m, img, 0) != nil || m.Fam.Get(m, img, 8) != nil {
			t.Fatalf("%s: canale cancellato: % X", id, img[0x10:0x20])
		}
		tx, rx, wtx, wrx = nil, nil, nil, nil
		emuRunRec(e, &tx, &rx, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
		})
		hello()
		for a := 0; a < 0x3F0; a += 8 {
			step(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...), []byte{0x06})
		}
		wtx = append(wtx, 'b')
		if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
			t.Fatalf("%s: dialogo di scrittura diverso dalla scheda (%d/%d byte mandati, %d/%d ricevuti)", id, len(tx), len(wtx), len(rx), len(wrx))
		}
		got := e.(interface{ image() []byte }).image()
		if !bytes.Equal(got, img) || !bytes.Equal(got[0x3B0:], radio[0x3B0:]) {
			t.Fatalf("%s: memoria della radio diversa dopo la scrittura (o impostazioni e password toccate)", id)
		}

		// password di programmazione (6 byte a 0x03B0): si legge e si scrive lo stesso, e resta com'è
		locked := append([]byte{}, radio...)
		copy(locked[0x3B0:], []byte{12, 13, 14, 72, 0xFF, 0xFF})
		pwThrough447(t, m, m.Fam.(t18Fam).emu(m, locked), 0x3B0, []byte{12, 13, 14, 72, 0xFF, 0xFF}, func(mem []byte) {
			if err := m.Fam.Set(m, mem, 5, &Chan{Rx: 446093750, Tx: 446093750, Narrow: true, Scan: true, Power: 2}); err != nil {
				t.Fatal(err)
			}
		})

		// canale «cifrato» (bit 1 delle opzioni a 0): frequenza come numero a 32 bit;
		// non modificato resta com'è, modificato diventa un canale normale in BCD
		enc := append([]byte{}, radio...)
		crec := []byte{0x7A, 0xD0, 0xC1, 0x02, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0xED, 0x11, 0x22, 0x33}
		copy(enc[0x50:], crec)
		g := m.Fam.Get(m, enc, 4)
		if g == nil || g.Rx != 462562500 || g.Tx != g.Rx || g.TxOff || g.RxTone != "" || g.TxTone != "" {
			t.Fatalf("%s: canale «cifrato» letto come %+v", id, g)
		}
		same := *g
		if err := m.Fam.Set(m, enc, 4, &same); err != nil || !bytes.Equal(enc[0x50:0x60], crec) {
			t.Fatalf("%s: canale «cifrato» non modificato riscritto come % X (%v)", id, enc[0x50:0x60], err)
		}
		chg := *g
		chg.RxTone = "88.5"
		if err := m.Fam.Set(m, enc, 4, &chg); err != nil || enc[0x5C]&0x02 == 0 ||
			!bytes.Equal(enc[0x50:0x5C], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0x85, 0x08, 0xFF, 0xFF}) || !bytes.Equal(enc[0x5D:0x60], crec[13:]) {
			t.Fatalf("%s: canale «cifrato» modificato: % X (%v)", id, enc[0x50:0x60], err)
		}
	}
	// un modello con un'altra sigla non si collega
	a, b := genByID["rt-b3h"], genByID["rt-rt68h"]
	emuRun(b.Fam.(t18Fam).emu(b, bytes.Repeat([]byte{0xFF}, 0x3F0)), func() {
		if _, _, err := a.Fam.Download(context.Background(), a, "EMU", nop); err == nil {
			t.Fatal("RT68H letta con la sigla della B3H")
		}
	})
	// i modelli di CHIRP della stessa famiglia restano com'erano: nessun passo della password
	if o := t18o(genByID["rt-rb29"]); o.vend || o.spec.pwStep || o.spec.anyID {
		t.Fatal("RB29 (CHIRP) cambiata")
	}
}

// 4.47, Retevis RB65 («K22GRAM»): canali da 8 byte da 0x0000 con una sola
// frequenza e un solo tono; stesso dialogo dei modelli qui sopra
func TestRB65v447(t *testing.T) {
	nop := func(int, int, string) {}
	m := genByID["rt-rb65"]
	if m == nil || !m.NoChirp || !m.NoTxOff || m.Slots != 22 {
		t.Fatal("rt-rb65: modello assente o con altri parametri")
	}
	sp := rb65o(m).spec
	if string(sp.magic) != "\x02K22GRAM" || !sp.anyID || !sp.pwStep || sp.block != 8 || sp.memsize != 0x3F0 || !sp.noAck || string(sp.exit) != "b" {
		t.Fatal("rt-rb65: collegamento con altri parametri")
	}
	radio := bytes.Repeat([]byte{0xFF}, 0x3F0)
	// canale 1: 462,5625 MHz, 67,0; canale 8: 467,5625 MHz, D243N (numero: A3 80), bassa; canale 16: 462,5750, D743I
	copy(radio[0x00:], []byte{0x50, 0x62, 0x25, 0x46, 0x70, 0x06, 0xFF, 0xFF})
	copy(radio[0x38:], []byte{0x50, 0x62, 0x75, 0x46, 0xA3, 0x80, 0xE7, 0xFF})
	copy(radio[0x78:], []byte{0x00, 0x75, 0x25, 0x46, 0xE3, 0xC1, 0xFB, 0x5A})
	for k := 0x3C0; k < 0x3D0; k++ {
		radio[k] = byte(k)
	}
	for _, w := range []struct {
		i              int
		rx             int64
		tone           string
		scan, nar, low bool
	}{{0, 462562500, "67.0", false, true, false}, {7, 467562500, "D243N", true, true, true}, {15, 462575000, "D743I", false, false, false}} {
		g := m.Fam.Get(m, radio, w.i)
		if g == nil || g.Rx != w.rx || g.Tx != w.rx || g.TxOff || g.RxTone != w.tone || g.TxTone != w.tone || g.Scan != w.scan || g.Narrow != w.nar ||
			(m.powerIndex(g.Power) == 0) != w.low {
			t.Fatalf("canale %d: %+v", w.i+1, g)
		}
	}
	if m.Fam.Get(m, radio, 1) != nil {
		t.Fatal("canale libero letto come occupato")
	}
	e := m.Fam.(rb65Fam).emu(m, radio)
	var tx, rx, img []byte
	emuRunRec(e, &tx, &rx, func() {
		var err error
		if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
			t.Fatalf("lettura: %v", err)
		}
	})
	if !bytes.Equal(img, radio) {
		t.Fatal("lettura diversa dalla memoria della radio")
	}
	var wtx, wrx []byte
	step := func(send, reply []byte) { wtx, wrx = append(wtx, send...), append(wrx, reply...) }
	hello := func() {
		step([]byte("\x02K22GRAM"), []byte{0x06})
		step([]byte{0x02}, []byte("SMP558\x00\x00"))
		step([]byte{0x06}, []byte{0x06})
		step([]byte{0x05}, radio[0x3B0:0x3B6])
		step([]byte{0x06}, []byte{0x06})
	}
	hello()
	for a := 0; a < 0x3F0; a += 8 {
		step([]byte{'R', byte(a >> 8), byte(a), 8}, append([]byte{'W', byte(a >> 8), byte(a), 8}, radio[a:a+8]...))
	}
	wtx = append(wtx, 'b')
	if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
		t.Fatalf("dialogo di lettura diverso dalla scheda (%d/%d, %d/%d)", len(tx), len(wtx), len(rx), len(wrx))
	}
	// canale nuovo, canale modificato (il byte 7 resta), canale cancellato
	if err := m.Fam.Set(m, img, 2, &Chan{Rx: 446093750, Tx: 446093750, Narrow: true, Scan: true, Power: 2, RxTone: "D023N", TxTone: "D023N"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(img[0x10:0x18], []byte{0x75, 0x93, 0x60, 0x44, 0x13, 0x80, 0xE7, 0xFF}) {
		t.Fatalf("canale nuovo scritto come % X", img[0x10:0x18])
	}
	mod := *m.Fam.Get(m, img, 15)
	mod.RxTone, mod.TxTone, mod.Narrow = "88.5", "88.5", true
	if err := m.Fam.Set(m, img, 15, &mod); err != nil || !bytes.Equal(img[0x78:0x80], []byte{0x00, 0x75, 0x25, 0x46, 0x85, 0x08, 0xFF, 0x5A}) {
		t.Fatalf("canale modificato: % X (%v)", img[0x78:0x80], err)
	}
	if err := m.Fam.Set(m, img, 0, nil); err != nil || m.Fam.Get(m, img, 0) != nil {
		t.Fatal("canale non cancellato")
	}
	tx, rx, wtx, wrx = nil, nil, nil, nil
	emuRunRec(e, &tx, &rx, func() {
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
			t.Fatalf("scrittura: %v", err)
		}
	})
	hello()
	for a := 0; a < 0x3F0; a += 8 {
		step(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...), []byte{0x06})
	}
	wtx = append(wtx, 'b')
	if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
		t.Fatalf("dialogo di scrittura diverso dalla scheda (%d/%d, %d/%d)", len(tx), len(wtx), len(rx), len(wrx))
	}
	if got := e.(interface{ image() []byte }).image(); !bytes.Equal(got, img) || !bytes.Equal(got[0x3B0:], radio[0x3B0:]) {
		t.Fatal("memoria della radio diversa dopo la scrittura (o impostazioni e password toccate)")
	}
	// password di programmazione (6 byte a 0x03B0): si legge e si scrive lo stesso, e resta com'è
	locked := append([]byte{}, radio...)
	copy(locked[0x3B0:], []byte{12, 13, 14, 15, 16, 17})
	pwThrough447(t, m, m.Fam.(rb65Fam).emu(m, locked), 0x3B0, []byte{12, 13, 14, 15, 16, 17}, func(mem []byte) {
		if err := m.Fam.Set(m, mem, 20, &Chan{Rx: 462637500, Tx: 462637500, Narrow: true, Scan: true}); err != nil {
			t.Fatal(err)
		}
	})
	// una sola frequenza e un solo tono: il resto si rifiuta, anche in fit
	for _, c := range []*Chan{
		{Rx: 430000000, Tx: 431600000}, {Rx: 446006250, Tx: 446006250, TxTone: "88.5"}, {Rx: 446006250, Tx: 446006250, RxTone: "67.0", TxTone: "88.5"},
	} {
		b := append([]byte{}, radio...)
		if ok, _ := m.Fam.(rb65Fam).fit(m, c); ok {
			t.Fatalf("fit accetta %+v", c)
		}
		if err := m.Fam.Set(m, b, 3, c); err == nil || !bytes.Equal(b, radio) {
			t.Fatalf("Set accetta %+v", c)
		}
	}
	// canale «cifrato»
	enc := append([]byte{}, radio...)
	crec := []byte{0x7A, 0xD0, 0xC1, 0x02, 0x9A, 0xBC, 0xED, 0x33}
	copy(enc[0x20:], crec)
	g := m.Fam.Get(m, enc, 4)
	if g == nil || g.Rx != 462562500 || g.Tx != g.Rx || g.RxTone != "" || g.TxTone != "" {
		t.Fatalf("canale «cifrato» letto come %+v", g)
	}
	same := *g
	if err := m.Fam.Set(m, enc, 4, &same); err != nil || !bytes.Equal(enc[0x20:0x28], crec) {
		t.Fatalf("canale «cifrato» non modificato riscritto come % X (%v)", enc[0x20:0x28], err)
	}
	chg := *g
	chg.RxTone, chg.TxTone = "88.5", "88.5"
	if err := m.Fam.Set(m, enc, 4, &chg); err != nil || !bytes.Equal(enc[0x20:0x28], []byte{0x50, 0x62, 0x25, 0x46, 0x85, 0x08, 0xEF, 0x33}) {
		t.Fatalf("canale «cifrato» modificato: % X (%v)", enc[0x20:0x28], err)
	}
}
