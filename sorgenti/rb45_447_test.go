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

// 4.47, Retevis RB45, RB645 («M59GRAM») e RB47 («182GRAM»): canali da 11 byte.
// La memoria di partenza è scritta a mano dalla scheda del protocollo (non con il
// driver): frequenze a 24 bit da 0x2000000, toni, opzioni; poi lettura e scrittura
// con la radio simulata, dialogo byte per byte, password, canale «cifrato».
func TestRB45Group447(t *testing.T) {
	nop := func(int, int, string) {}
	for _, x := range []struct {
		id, cmd string
		slots   int
		dig     bool
	}{
		{"rt-rb45", "M59GRAM", 22, false},
		{"rt-rb645", "M59GRAM", 16, false},
		{"rt-rb47", "182GRAM", 22, true},
	} {
		id := x.id
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := h777v4o(m)
		if !m.NoChirp || m.Slots != x.slots || o.cmd != x.cmd || !o.rb45 || o.dcsDig != x.dig || o.bs != 11 || o.memsize != 0xFD ||
			!o.pass || !o.bye || o.ack != 0 || o.echo || o.base != 0x2000000 || o.bits != 24 {
			t.Fatalf("%s: modello con altri parametri: %+v", id, o)
		}
		radio := bytes.Repeat([]byte{0xFF}, 0xFD)
		// canale 1: 462,5625 MHz, tono 67,0, opzioni FF (non in scansione, alta, stretta)
		copy(radio[0:], []byte{0x7A, 0xD0, 0xC1, 0x7A, 0xD0, 0xC1, 0x70, 0x06, 0x70, 0x06, 0xFF})
		// canale 8: 467,5625 MHz, D243N, bassa potenza, in scansione, larga
		dcs := []byte{0xA3, 0x80}
		inv := []byte{0xE3, 0xC1} // D743I
		if x.dig {
			dcs, inv = []byte{0x43, 0x82}, []byte{0x43, 0xC7}
		}
		copy(radio[7*11:], []byte{0x9A, 0x71, 0xC9, 0x9A, 0x71, 0xC9, dcs[0], dcs[1], dcs[0], dcs[1], 0xE3})
		// canale 16: 446,00625 MHz in ricezione, trasmissione spenta, D743I solo in ricezione
		copy(radio[15*11:], []byte{0x31, 0x8D, 0xA8, 0xFF, 0xFF, 0xFF, inv[0], inv[1], 0xFF, 0xFF, 0xFF})
		// impostazioni «di fabbrica» (valori qualsiasi), nessuna password
		for k := 0; k < 8; k++ {
			radio[0xF2+k] = byte(0x10 + k)
		}
		type want struct {
			i              int
			rx, tx         int64
			off            bool
			rt, tt         string
			scan, nar, low bool
		}
		for _, w := range []want{
			{0, 462562500, 462562500, false, "67.0", "67.0", false, true, false},
			{7, 467562500, 467562500, false, "D243N", "D243N", true, false, true},
			{15, 446006250, 446006250, true, "D743I", "", false, true, false},
		} {
			g := m.Fam.Get(m, radio, w.i)
			if g == nil || g.Rx != w.rx || g.Tx != w.tx || g.TxOff != w.off || g.RxTone != w.rt || g.TxTone != w.tt || g.Scan != w.scan ||
				g.Narrow != w.nar || (m.powerIndex(g.Power) == 0) != w.low {
				t.Fatalf("%s canale %d: %+v", id, w.i+1, g)
			}
		}
		if m.Fam.Get(m, radio, 1) != nil {
			t.Fatalf("%s: canale libero letto come occupato", id)
		}
		// un canale è libero solo se tutti e tre i byte della frequenza sono FF
		odd := append([]byte{}, radio...)
		copy(odd[11:], []byte{0xFF, 0xD0, 0xC1, 0xFF, 0xD0, 0xC1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		if g := m.Fam.Get(m, odd, 1); g == nil || g.Rx != (0x2000000+0xC1D0FF)*10 {
			t.Fatalf("%s: frequenza con il primo byte FF letta come canale libero: %+v", id, g)
		}

		// lettura: la radio simulata, come nei file dati del costruttore, da 0xD1 risponde con 01 nel byte alto
		e := m.Fam.(h777v4Fam).emu(m, radio).(*wxEmu)
		e.hi01 = 0xD1
		var tx, rx []byte
		var img []byte
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
			step([]byte{0x02}, make([]byte, 8))
			step([]byte{0x06}, []byte{0x06})
			step([]byte{0x05}, radio[0xFA:0xFD]) // password: 3 byte
			step([]byte{0x06}, []byte{0x06})
		}
		hello()
		for a := 0; a < 0xFD; a += 11 {
			hi := byte(0)
			if a >= 0xD1 {
				hi = 1
			}
			step([]byte{'R', 0, byte(a), 11}, append([]byte{'W', hi, byte(a), 11}, radio[a:a+11]...))
		}
		// dopo la lettura il programma del costruttore non manda la chiusura
		if !bytes.Equal(tx, wtx) {
			t.Fatalf("%s: byte mandati in lettura\n% X\nattesi\n% X", id, tx, wtx)
		}
		if !bytes.Equal(rx, wrx) {
			t.Fatalf("%s: byte ricevuti in lettura\n% X\nattesi\n% X", id, rx, wrx)
		}
		// una radio di questa famiglia non accetta la sigla di un modello da 13 byte
		b3b := genByID["rt-b3b"]
		emuRun(m.Fam.(h777v4Fam).emu(m, radio), func() {
			if _, _, err := b3b.Fam.Download(context.Background(), b3b, "EMU", nop); err == nil {
				t.Fatalf("%s: letta con la sigla della B3B", id)
			}
		})

		// modifiche: canale nuovo, canale cancellato, toni cambiati
		nc := &Chan{Rx: 446093750, Tx: 446093750, Narrow: true, Scan: true, Power: 2, RxTone: "D023N", TxTone: "123.0"}
		if err := m.Fam.Set(m, img, 2, nc); err != nil {
			t.Fatal(err)
		}
		// posizioni e codifica scritte a mano: 446,09375 MHz = 0x2A8AF5F; D023N = 13 80 (numero) oppure 23 80 (cifre)
		d23 := []byte{0x13, 0x80}
		if x.dig {
			d23 = []byte{0x23, 0x80}
		}
		wrec := []byte{0x5F, 0xAF, 0xA8, 0x5F, 0xAF, 0xA8, d23[0], d23[1], 0x30, 0x12, 0xEF &^ 0x08}
		if !bytes.Equal(img[22:33], wrec) {
			t.Fatalf("%s: canale nuovo scritto come % X, atteso % X", id, img[22:33], wrec)
		}
		if err := m.Fam.Set(m, img, 7, nil); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(img[77:88], bytes.Repeat([]byte{0xFF}, 11)) {
			t.Fatalf("%s: canale cancellato non tutto a FF", id)
		}
		tx, rx, wtx, wrx = nil, nil, nil, nil
		emuRunRec(e, &tx, &rx, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
		})
		hello()
		for a := 0; a < 0xFD; a += 11 {
			step(append([]byte{'W', 0, byte(a), 11}, img[a:a+11]...), []byte{0x06})
		}
		wtx = append(wtx, 'b')
		if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
			t.Fatalf("%s: dialogo di scrittura diverso dalla scheda\n% X\n% X", id, tx, rx)
		}
		got := e.image()
		if !bytes.Equal(got, img) || !bytes.Equal(got[0xF2:], radio[0xF2:]) {
			t.Fatalf("%s: memoria della radio diversa dopo la scrittura (o impostazioni toccate)", id)
		}
		if g := m.Fam.Get(m, got, 2); g == nil || g.Rx != nc.Rx || g.RxTone != "D023N" || g.TxTone != "123.0" || !g.Scan || m.Fam.Get(m, got, 7) != nil {
			t.Fatalf("%s: canale nuovo o cancellato non arrivati alla radio", id)
		}

		// password di programmazione (3 byte a 0xFA): si legge e si scrive lo stesso, e resta com'è
		locked := append([]byte{}, radio...)
		copy(locked[0xFA:], []byte{1, 61, 0xFF})
		pwThrough447(t, m, m.Fam.(h777v4Fam).emu(m, locked), 0xFA, []byte{1, 61, 0xFF}, func(mem []byte) {
			if err := m.Fam.Set(m, mem, 4, &Chan{Rx: 462637500, Tx: 462637500, Narrow: true, Scan: true}); err != nil {
				t.Fatal(err)
			}
		})

		// canale «cifrato» (bit 1 delle opzioni a 0): si vede la sola frequenza; se non
		// lo si modifica resta byte per byte com'è; se lo si modifica diventa normale
		enc := append([]byte{}, radio...)
		crec := []byte{0x7A, 0xD0, 0xC1, 0x12, 0x34, 0x56, 0x9A, 0xBC, 0xDE, 0xF0, 0xFD}
		copy(enc[44:], crec)
		g := m.Fam.Get(m, enc, 4)
		if g == nil || g.Rx != 462562500 || g.Tx != g.Rx || g.TxOff || g.RxTone != "" || g.TxTone != "" {
			t.Fatalf("%s: canale «cifrato» letto come %+v", id, g)
		}
		same := *g
		if err := m.Fam.Set(m, enc, 4, &same); err != nil || !bytes.Equal(enc[44:55], crec) {
			t.Fatalf("%s: canale «cifrato» non modificato riscritto come % X (%v)", id, enc[44:55], err)
		}
		chg := *g
		chg.RxTone, chg.TxTone = "88.5", "88.5"
		if err := m.Fam.Set(m, enc, 4, &chg); err != nil || enc[54]&0x02 == 0 || !bytes.Equal(enc[44:54], []byte{0x7A, 0xD0, 0xC1, 0x7A, 0xD0, 0xC1, 0x85, 0x08, 0x85, 0x08}) {
			t.Fatalf("%s: canale «cifrato» modificato: % X (%v)", id, enc[44:55], err)
		}
		// frequenza fuori dalla memoria a 24 bit
		if err := m.Fam.Set(m, enc, 5, &Chan{Rx: 144000000, Tx: 144000000}); err == nil {
			t.Fatalf("%s: accettata una frequenza che la memoria non può contenere", id)
		}
	}
	// le due sigle non si scambiano: una RB47 rifiuta il collegamento di una RB45
	a, b := genByID["rt-rb45"], genByID["rt-rb47"]
	emuRun(b.Fam.(h777v4Fam).emu(b, bytes.Repeat([]byte{0xFF}, 0xFD)), func() {
		if _, _, err := a.Fam.Download(context.Background(), a, "EMU", nop); err == nil {
			t.Fatal("RB47 letta con la sigla della RB45")
		}
	})
	// e i due modi di scrivere il DCS restano distinti
	if h777v4ToneDec([]byte{0x43, 0x82}, true) != "D243N" || h777v4ToneDec([]byte{0xA3, 0x80}, false) != "D243N" ||
		h777v4ToneDec([]byte{0x93, 0x82}, true) != "" || h777v4ToneDec([]byte{0x43, 0xC7}, true) != "D743I" {
		t.Fatal("codifica del DCS")
	}
}
