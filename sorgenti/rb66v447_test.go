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

// 4.47, Retevis RB66 e RB66P (sigla «D22GRA5»): canali da 8 byte con una sola
// frequenza e un solo tono, collegamento delle RB666 (conferma 66, eco). La
// memoria di partenza è scritta a mano dalla scheda del protocollo (non con il
// driver); poi lettura e scrittura con la radio simulata, dialogo byte per byte
// con e senza eco, password, canale «cifrato», byte che i programmi non toccano.
func TestRB66v447(t *testing.T) {
	nop := func(int, int, string) {}
	for _, x := range []struct {
		id       string
		size, pw int
	}{
		{"rt-rb66", 0xE0, 6},
		{"rt-rb66p", 0xE8, 4},
	} {
		id := x.id
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := h777v4o(m)
		if !m.NoChirp || !m.NoTxOff || m.Slots != 22 || o.cmd != "D22GRA5" || !o.rb66 || o.rb45 || o.dcsDig || o.bs != 8 || o.memsize != x.size ||
			o.pwN != x.pw || !o.pass || !o.bye || o.ack != 0x66 || !o.echo || o.base != 0x2000000 || o.bits != 24 {
			t.Fatalf("%s: modello con altri parametri: %+v", id, o)
		}
		fam := m.Fam.(rb66Fam)
		radio := bytes.Repeat([]byte{0xFF}, x.size)
		// canale 1: 462,5625 MHz, tono 67,0, opzioni FF (non in scansione, alta, stretta)
		copy(radio[0:], []byte{0x7A, 0xD0, 0xC1, 0xFF, 0xFF, 0x70, 0x06, 0xFF})
		// canale 8: 467,5625 MHz, D243N, bassa potenza (F7)
		copy(radio[7*8:], []byte{0x9A, 0x71, 0xC9, 0xFF, 0xFF, 0xA3, 0x80, 0xF7})
		// canale 16: 462,5750 MHz, D743I
		copy(radio[15*8:], []byte{0x5C, 0xD5, 0xC1, 0xFF, 0xFF, 0xE3, 0xC1, 0xFF})
		// canale 20: 462,6750 MHz, nessun tono, in scansione, alta, larga; i byte 3-4 con valori qualsiasi
		copy(radio[19*8:], []byte{0x6C, 0xFC, 0xC1, 0x12, 0x34, 0xFF, 0xFF, 0xEB})
		// impostazioni «di fabbrica» (valori qualsiasi), nessuna password
		copy(radio[0xD0:], []byte{0x37, 0x03, 0x06, 0x05, 0x00, 0x15, 0xFF, 0xFF, 0xFF, 0x06, 0x05})
		type want struct {
			i              int
			rx             int64
			tone           string
			scan, nar, low bool
		}
		for _, w := range []want{
			{0, 462562500, "67.0", false, true, false},
			{7, 467562500, "D243N", false, true, true},
			{15, 462575000, "D743I", false, true, false},
			{19, 462675000, "", true, false, false},
		} {
			g := fam.Get(m, radio, w.i)
			if g == nil || g.Rx != w.rx || g.Tx != w.rx || g.TxOff || g.RxTone != w.tone || g.TxTone != w.tone || g.Scan != w.scan ||
				g.Narrow != w.nar || (m.powerIndex(g.Power) == 0) != w.low {
				t.Fatalf("%s canale %d: %+v", id, w.i+1, g)
			}
		}
		for _, i := range []int{1, 21} {
			if fam.Get(m, radio, i) != nil {
				t.Fatalf("%s: canale libero %d letto come occupato", id, i+1)
			}
		}
		// un canale è libero solo se tutti e tre i byte della frequenza sono FF
		odd := append([]byte{}, radio...)
		copy(odd[8:], []byte{0xFF, 0xD0, 0xC1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		if g := fam.Get(m, odd, 1); g == nil || g.Rx != (0x2000000+0xC1D0FF)*10 {
			t.Fatalf("%s: frequenza con il primo byte FF letta come canale libero: %+v", id, g)
		}

		// modifiche: canale nuovo, canale cancellato, tono cambiato su un canale con i byte 3-4 pieni
		img := append([]byte{}, radio...)
		nc := &Chan{Rx: 462700000, Tx: 462700000, Narrow: false, Scan: true, Power: 2, RxTone: "D023N", TxTone: "D023N"}
		if err := fam.Set(m, img, 2, nc); err != nil {
			t.Fatal(err)
		}
		// scritto a mano: 462,7000 MHz = 0x2C20630; D023N = 13 80; opzioni: tutto a 1 tranne
		// scansione (bit 4), potenza (bit 3) e banda (bit 2)
		if !bytes.Equal(img[16:24], []byte{0x30, 0x06, 0xC2, 0xFF, 0xFF, 0x13, 0x80, 0xE3}) {
			t.Fatalf("%s: canale nuovo scritto come % X", id, img[16:24])
		}
		if err := fam.Set(m, img, 7, nil); err != nil || !bytes.Equal(img[56:64], bytes.Repeat([]byte{0xFF}, 8)) {
			t.Fatalf("%s: canale cancellato non tutto a FF (%v)", id, err)
		}
		c20 := *fam.Get(m, img, 19)
		c20.RxTone, c20.TxTone = "123.0", "123.0"
		if err := fam.Set(m, img, 19, &c20); err != nil || !bytes.Equal(img[152:160], []byte{0x6C, 0xFC, 0xC1, 0x12, 0x34, 0x30, 0x12, 0xEB}) {
			t.Fatalf("%s: canale 20 con il tono cambiato: % X (%v)", id, img[152:160], err)
		}
		// canale nuovo su una posizione libera con dei resti negli altri byte: parte pulito
		// (byte 3-4 a FF, opzioni come una riga nuova)
		dirty := append([]byte{}, radio...)
		copy(dirty[8:], []byte{0xFF, 0xFF, 0xFF, 0x01, 0x02, 0x03, 0x04, 0x00})
		if fam.Get(m, dirty, 1) != nil {
			t.Fatalf("%s: posizione libera con dei resti letta come canale", id)
		}
		if err := fam.Set(m, dirty, 1, &Chan{Rx: 462562500, Tx: 462562500, Narrow: true, Power: 0}); err != nil ||
			!bytes.Equal(dirty[8:16], []byte{0x7A, 0xD0, 0xC1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}) {
			t.Fatalf("%s: canale nuovo su una posizione con dei resti: % X (%v)", id, dirty[8:16], err)
		}
		// la radio non può contenere: trasmissione altrove o spenta, toni diversi, frequenza fuori dai 24 bit
		for k, bad := range []*Chan{
			{Rx: 462562500, Tx: 467562500},
			{Rx: 462562500, Tx: 462562500, TxOff: true},
			{Rx: 462562500, Tx: 462562500, RxTone: "67.0"},
			{Rx: 462562500, Tx: 462562500, RxTone: "67.0", TxTone: "D023N"},
			{Rx: 144000000, Tx: 144000000},
			{Rx: (0x2000000 + 0xFFFFFF) * 10, Tx: (0x2000000 + 0xFFFFFF) * 10},
		} {
			before := append([]byte{}, img...)
			if ok, _ := fam.fit(m, bad); ok {
				t.Fatalf("%s: canale non ammesso %d accettato da fit", id, k)
			}
			if err := fam.Set(m, img, 5, bad); err == nil || !bytes.Equal(img, before) {
				t.Fatalf("%s: canale non ammesso %d scritto (%v)", id, k, err)
			}
		}

		// il dialogo, byte per byte, come nella scheda: 02 e sigla insieme, 02, 06, 05 (6 byte
		// la RB66, 4 la RB66P), 06, i blocchi da 8 byte e la chiusura «b»; la radio conferma
		// con 66 e, con l'eco, prima di ogni risposta rimanda quello che ha ricevuto
		pwd := bytes.Repeat([]byte{0xFF}, x.pw) // da 0xDD; oltre la memoria letta (RB66) vale FF
		copy(pwd, radio[0xDD:])
		for _, echo := range []bool{false, true} {
			er := fam.emu(m, radio).(*wxEmu)
			er.echo = echo
			var tx, rx, wtx, wrx []byte
			step := func(send, reply []byte) {
				wtx = append(wtx, send...)
				if echo {
					wrx = append(wrx, send...)
				}
				wrx = append(wrx, reply...)
			}
			hello := func() {
				step(append([]byte{0x02}, "D22GRA5"...), []byte{0x66})
				step([]byte{0x02}, make([]byte, 8))
				step([]byte{0x06}, []byte{0x66})
				step([]byte{0x05}, pwd)
				step([]byte{0x06}, []byte{0x66})
			}
			var got []byte
			emuRunRec(er, &tx, &rx, func() {
				var err error
				if got, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
					t.Fatalf("%s (eco %v): lettura: %v", id, echo, err)
				}
			})
			if !bytes.Equal(got, radio) {
				t.Fatalf("%s (eco %v): lettura diversa dalla memoria della radio", id, echo)
			}
			hello()
			blocks := 0
			for a := 0; a < x.size; a += 8 {
				step([]byte{'R', 0, byte(a), 8}, append([]byte{'W', 0, byte(a), 8}, radio[a:a+8]...))
				blocks++
			}
			wtx = append(wtx, 'b')
			if blocks != x.size/8 || (id == "rt-rb66" && blocks != 28) || (id == "rt-rb66p" && blocks != 29) {
				t.Fatalf("%s: %d blocchi", id, blocks)
			}
			if !bytes.Equal(tx, wtx) {
				t.Fatalf("%s (eco %v): byte mandati in lettura\n% X\nattesi\n% X", id, echo, tx, wtx)
			}
			if !bytes.Equal(rx, wrx) {
				t.Fatalf("%s (eco %v): byte ricevuti in lettura\n% X\nattesi\n% X", id, echo, rx, wrx)
			}
			tx, rx, wtx, wrx = nil, nil, nil, nil
			emuRunRec(er, &tx, &rx, func() {
				if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
					t.Fatalf("%s (eco %v): scrittura: %v", id, echo, err)
				}
			})
			hello()
			for a := 0; a < x.size; a += 8 {
				step(append([]byte{'W', 0, byte(a), 8}, img[a:a+8]...), []byte{0x66})
			}
			wtx = append(wtx, 'b')
			if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
				t.Fatalf("%s (eco %v): dialogo di scrittura diverso dalla scheda\n% X\n% X", id, echo, tx, rx)
			}
			after := er.image()
			if !bytes.Equal(after, img) || !bytes.Equal(after[0xB0:], radio[0xB0:]) {
				t.Fatalf("%s (eco %v): memoria della radio diversa dopo la scrittura (o impostazioni toccate)", id, echo)
			}
			if g := fam.Get(m, after, 2); g == nil || g.Rx != nc.Rx || g.RxTone != "D023N" || g.TxTone != "D023N" || !g.Scan || g.Narrow ||
				m.powerIndex(g.Power) != 0 || fam.Get(m, after, 7) != nil {
				t.Fatalf("%s: canale nuovo o cancellato non arrivati alla radio", id)
			}
		}

		// una radio che conferma con 06 non è una di queste: niente lettura, niente scrittura
		e6 := fam.emu(m, radio).(*wxEmu)
		e6.ack = 0x06
		emuRun(e6, func() {
			if _, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err == nil || strings.Contains(err.Error(), "serie più vecchia") {
				t.Fatalf("%s: conferma 06, lettura: %v", id, err)
			}
		})
		emuRun(e6, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err == nil {
				t.Fatalf("%s: conferma 06, scrittura accettata", id)
			}
		})
		if !bytes.Equal(e6.image(), radio) {
			t.Fatalf("%s: radio con la conferma 06 modificata", id)
		}
		// e una conferma 06 a un blocco scritto non vale: la scrittura si ferma al primo blocco
		eb := fam.emu(m, radio).(*wxEmu)
		eb.ackW = 0x06
		emuRun(eb, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err == nil || !strings.Contains(err.Error(), "0x0000") {
				t.Fatalf("%s: blocco confermato con 06: %v", id, err)
			}
		})
		if !bytes.Equal(eb.image(), radio) {
			t.Fatalf("%s: radio modificata da una scrittura rifiutata", id)
		}
		// una RB666 (altra sigla) non accetta il collegamento di questo modello, né il contrario
		rb666 := genByID["rt-rb666p"]
		emuRun(rb666.Fam.(h777v4Fam).emu(rb666, bytes.Repeat([]byte{0xFF}, 0xEA)), func() {
			if _, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err == nil {
				t.Fatalf("%s: letta una RB666P", id)
			}
		})
		emuRun(fam.emu(m, radio), func() {
			if _, _, err := rb666.Fam.Download(context.Background(), rb666, "EMU", nop); err == nil {
				t.Fatalf("%s: letta con la sigla della RB666P", id)
			}
		})

		// password di programmazione (da 0xDD; della RB66 la memoria ne tiene tre byte,
		// della RB66P quattro): si legge e si scrive lo stesso, e resta com'è
		pwb := []byte{12, 13, 14} // RB66: «123» nella tabella del suo programma
		if id == "rt-rb66p" {
			pwb = []byte{1, 2, 3, 4} // RB66P: «1234»
		}
		if len(pwb) != x.pw && len(pwb) != x.size-0xDD {
			t.Fatalf("%s: lunghezza della password di prova", id)
		}
		locked := append([]byte{}, radio...)
		copy(locked[0xDD:], pwb)
		pwThrough447(t, m, fam.emu(m, locked), 0xDD, pwb, func(mem []byte) {
			if err := m.Fam.Set(m, mem, 20, &Chan{Rx: 462637500, Tx: 462637500, Narrow: true, Scan: true}); err != nil {
				t.Fatal(err)
			}
		})

		// canale «cifrato» (bit 1 delle opzioni a 0): si vede la sola frequenza; se non lo si
		// modifica resta byte per byte com'è; se lo si modifica diventa normale
		enc := append([]byte{}, radio...)
		crec := []byte{0x7A, 0xD0, 0xC1, 0x56, 0x78, 0x70, 0x06, 0xFD} // con un tono valido (67,0) che deve restare nascosto
		copy(enc[32:], crec)
		g := fam.Get(m, enc, 4)
		if g == nil || g.Rx != 462562500 || g.Tx != g.Rx || g.TxOff || g.RxTone != "" || g.TxTone != "" {
			t.Fatalf("%s: canale «cifrato» letto come %+v", id, g)
		}
		same := *g
		if err := fam.Set(m, enc, 4, &same); err != nil || !bytes.Equal(enc[32:40], crec) {
			t.Fatalf("%s: canale «cifrato» non modificato riscritto come % X (%v)", id, enc[32:40], err)
		}
		chg := *g
		chg.RxTone, chg.TxTone = "88.5", "88.5"
		if err := fam.Set(m, enc, 4, &chg); err != nil || !bytes.Equal(enc[32:40], []byte{0x7A, 0xD0, 0xC1, 0x56, 0x78, 0x85, 0x08, 0xFF}) {
			t.Fatalf("%s: canale «cifrato» modificato: % X (%v)", id, enc[32:40], err)
		}
	}
	// RB66 e RB66P hanno la stessa sigla ma non la stessa memoria: con la scheda
	// sbagliata la lettura si ferma (cambia il numero di byte della password), e
	// «Identifica» non dà mai all'una la memoria letta con il driver dell'altra
	a, b := genByID["rt-rb66"], genByID["rt-rb66p"]
	if identShape(a) == identShape(b) {
		t.Fatal("RB66 e RB66P: stesso tipo di memoria per «Identifica»")
	}
	if identShape(genByID["rt-rb669"]) == identShape(genByID["rt-rt68h"]) {
		t.Fatal("RB669 e RT68H: stesso tipo di memoria per «Identifica»")
	}
	emuRun(b.Fam.(rb66Fam).emu(b, bytes.Repeat([]byte{0xFF}, 0xE8)), func() {
		if _, _, err := a.Fam.Download(context.Background(), a, "EMU", nop); err == nil {
			t.Fatal("RB66P letta con la scheda della RB66")
		}
	})
}
