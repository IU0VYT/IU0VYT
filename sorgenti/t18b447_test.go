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
	"fmt"
	"strings"
	"testing"
)

// 4.47, Retevis RT68B, RT668B («T68BRAM») e RB669 («H32GRAM»): famiglia della T18,
// programmi «Analog1» della generazione successiva. Il canale è quello già provato
// in TestT18Vend447; qui si prova ciò che cambia: l'identificazione che deve
// cominciare con «S», i due soli intervalli letti e scritti (il resto della memoria
// non si chiede e non si tocca), la conferma 86 di un blocco scritto, la RB669 che
// dopo la lettura non manda «b», i 32 canali della RB669.
func TestT18Gen2v447(t *testing.T) {
	nop := func(int, int, string) {}
	for _, x := range []struct {
		id, cmd string
		slots   int
		chEnd   int
		bye     bool // «b» dopo la lettura
	}{
		{"rt-rt68b", "T68BRAM", 16, 0x110, true},
		{"rt-rt668b", "T68BRAM", 16, 0x110, true},
		{"rt-rb669", "H32GRAM", 32, 0x220, false},
	} {
		id := x.id
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := t18o(m)
		sp := o.spec
		rr := [][2]int{{0, x.chEnd}, {0x3A0, 0x3F0}}
		if !m.NoChirp || m.Slots != x.slots || !o.vend || !o.reserved || o.lock != "" || o.power != "rw" || string(sp.magic) != "\x02"+x.cmd ||
			sp.anyID || len(sp.fingers) != 1 || string(sp.fingers[0]) != "S" || !sp.pwStep || sp.block != 8 || sp.memsize != 0x3F0 ||
			!sp.noAck || sp.echo || string(sp.exit) != "b" || !sp.ack86 || sp.noExitRead == x.bye ||
			fmt.Sprint(sp.rranges) != fmt.Sprint(rr) || fmt.Sprint(sp.ranges) != fmt.Sprint(rr) {
			t.Fatalf("%s: modello con altri parametri: %+v", id, sp)
		}
		// memoria della radio scritta a mano: canali, resti nella parte che il programma
		// non legge (da ritrovare intatti nella radio e mai nell'immagine), impostazioni
		radio := bytes.Repeat([]byte{0xFF}, 0x3F0)
		// canale 1 (0x0010): 462,5625 MHz, tono 67,0, opzioni EF, coda FF
		copy(radio[0x10:], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0x70, 0x06, 0x70, 0x06, 0xEF, 0xFF, 0xFF, 0xFF})
		// ultimo canale: 446,19375 MHz (RB669, canale 32: 149,0875 MHz), D026N, opzioni E7
		last := 0x10 + 16*(x.slots-1)
		if x.slots == 32 {
			copy(radio[last:], []byte{0x50, 0x87, 0x90, 0x14, 0x50, 0x87, 0x90, 0x14, 0x26, 0x80, 0x26, 0x80, 0xE7, 0xFF, 0xFF, 0xFF})
		} else {
			copy(radio[last:], []byte{0x75, 0x93, 0x61, 0x44, 0x75, 0x93, 0x61, 0x44, 0x26, 0x80, 0x26, 0x80, 0xE7, 0xFF, 0xFF, 0xFF})
		}
		if last+16 > x.chEnd || last+16 < x.chEnd-16 { // la RB669 legge anche i 16 byte dopo l'ultimo canale
			t.Fatalf("%s: i canali finiscono a 0x%X", id, last+16)
		}
		for k := x.chEnd; k < 0x3A0; k++ {
			radio[k] = byte(k*7 + 1)
		}
		copy(radio[0x3B8:], []byte("SMP558\xFF\xFF"))
		copy(radio[0x3C0:], []byte{0x87, 0x05, 0x00, 0x06, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0xFE, 0x00, 0xFF})
		want := append([]byte{}, radio...) // l'immagine attesa: la parte non letta resta a FF
		for k := x.chEnd; k < 0x3A0; k++ {
			want[k] = 0xFF
		}
		lastRx, lastLow := int64(446193750), true
		if x.slots == 32 {
			lastRx = 149087500
		}
		if g := m.Fam.Get(m, radio, 0); g == nil || g.Rx != 462562500 || g.RxTone != "67.0" || !g.Scan || m.powerIndex(g.Power) != 0 {
			t.Fatalf("%s canale 1: %+v", id, g)
		}
		if g := m.Fam.Get(m, radio, x.slots-1); g == nil || g.Rx != lastRx || g.Tx != lastRx || g.RxTone != "D026N" || g.TxTone != "D026N" ||
			(m.powerIndex(g.Power) == 1) != lastLow {
			t.Fatalf("%s ultimo canale: %+v", id, g)
		}

		ident := []byte("SMP558\xFF\xFF")
		newEmu := func(mem []byte) *rbEmu {
			e := m.Fam.(t18Fam).emu(m, mem).(*rbEmu)
			e.ident = ident
			return e
		}
		e := newEmu(radio)
		var tx, rx, img []byte
		emuRunRec(e, &tx, &rx, func() {
			var err error
			if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
				t.Fatalf("%s: lettura: %v", id, err)
			}
		})
		if !bytes.Equal(img, want) {
			t.Fatalf("%s: immagine letta diversa da quella attesa", id)
		}
		var wtx, wrx []byte
		step := func(send, reply []byte) {
			wtx = append(wtx, send...)
			wrx = append(wrx, reply...)
		}
		hello := func() {
			step(append([]byte{0x02}, x.cmd...), []byte{0x06})
			step([]byte{0x02}, ident)
			step([]byte{0x06}, []byte{0x06})
			step([]byte{0x05}, radio[0x3B0:0x3B6])
			step([]byte{0x06}, []byte{0x06})
		}
		hello()
		nblk := 0
		for _, r := range rr {
			for a := r[0]; a < r[1]; a += 8 {
				step([]byte{'R', byte(a >> 8), byte(a), 8}, append([]byte{'W', byte(a >> 8), byte(a), 8}, radio[a:a+8]...))
				nblk++
			}
		}
		if x.bye {
			wtx = append(wtx, 'b')
		}
		if nblk != x.chEnd/8+10 {
			t.Fatalf("%s: %d blocchi", id, nblk)
		}
		if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
			t.Fatalf("%s: dialogo di lettura diverso dalla scheda (%d/%d byte mandati, %d/%d ricevuti)", id, len(tx), len(wtx), len(rx), len(wrx))
		}
		outside := func(e *rbEmu, what string) {
			e.mu.Lock()
			defer e.mu.Unlock()
			for _, l := range e.log {
				var a int
				fmt.Sscanf(l[1:], "%X", &a)
				if a >= x.chEnd && a < 0x3A0 {
					t.Fatalf("%s: %s: chiesto alla radio l'indirizzo %s, che il programma del costruttore non tocca", id, what, l)
				}
			}
		}
		outside(e, "lettura")

		// modifiche: canale nuovo nell'ultima posizione libera prima dell'ultimo canale, primo canale cancellato
		nc := &Chan{Rx: 462725000, Tx: 462725000, Narrow: true, Scan: true, Power: 2, RxTone: "D023N", TxTone: "D023N"}
		if x.id == "rt-rt668b" || x.slots == 32 {
			nc.Rx, nc.Tx = 446006250, 446006250
		}
		if !m.inBands(nc.Rx) {
			t.Fatalf("%s: %d Hz fuori dalle gamme del modello", id, nc.Rx)
		}
		pos := x.slots - 2
		if err := m.Fam.Set(m, img, pos, nc); err != nil {
			t.Fatal(err)
		}
		at := 0x10 + 16*pos
		wrec := []byte{0x00, 0x25, 0x27, 0x46, 0x00, 0x25, 0x27, 0x46, 0x23, 0x80, 0x23, 0x80, 0xE7, 0xFF, 0xFF, 0xFF}
		if nc.Rx == 446006250 {
			copy(wrec, []byte{0x25, 0x06, 0x60, 0x44, 0x25, 0x06, 0x60, 0x44})
		}
		if !bytes.Equal(img[at:at+16], wrec) {
			t.Fatalf("%s: canale nuovo scritto come % X, atteso % X", id, img[at:at+16], wrec)
		}
		if err := m.Fam.Set(m, img, 0, nil); err != nil {
			t.Fatal(err)
		}
		// la radio conferma i blocchi scritti con 06 oppure con 86: il programma del costruttore accetta tutti e due
		for _, ak := range []byte{0x06, 0x86} {
			ew := newEmu(radio)
			ew.ackW = ak
			tx, rx, wtx, wrx = nil, nil, nil, nil
			emuRunRec(ew, &tx, &rx, func() {
				if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
					t.Fatalf("%s: scrittura (conferma %02X): %v", id, ak, err)
				}
			})
			hello()
			for _, r := range rr {
				for a := r[0]; a < r[1]; a += 8 {
					step(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...), []byte{ak})
				}
			}
			wtx = append(wtx, 'b')
			if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
				t.Fatalf("%s: dialogo di scrittura (conferma %02X) diverso dalla scheda (%d/%d byte mandati, %d/%d ricevuti)", id, ak, len(tx), len(wtx), len(rx), len(wrx))
			}
			outside(ew, "scrittura")
			got := ew.image()
			if !bytes.Equal(got[:x.chEnd], img[:x.chEnd]) || !bytes.Equal(got[x.chEnd:], radio[x.chEnd:]) {
				t.Fatalf("%s: dopo la scrittura (conferma %02X) canali non arrivati, oppure toccata la parte che non si scrive o le impostazioni", id, ak)
			}
			if g := m.Fam.Get(m, got, pos); g == nil || g.Rx != nc.Rx || g.RxTone != "D023N" || m.Fam.Get(m, got, 0) != nil {
				t.Fatalf("%s: canale nuovo o cancellato non arrivati alla radio", id)
			}
		}
		// una conferma diversa ferma la scrittura al primo blocco, senza toccare la radio
		er := newEmu(radio)
		er.ackW = 0x15
		emuRun(er, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err == nil || !strings.Contains(err.Error(), "0x0000") {
				t.Fatalf("%s: blocco rifiutato: %v", id, err)
			}
		})
		if !bytes.Equal(er.image(), radio) {
			t.Fatalf("%s: radio modificata da una scrittura rifiutata", id)
		}
		// identificazione che non comincia con «S»: il programma del costruttore si ferma, e anche qui
		ei := newEmu(radio)
		ei.ident = []byte("PMS558\xFF\xFF")
		emuRun(ei, func() {
			if _, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err == nil || !strings.Contains(err.Error(), "identificativo sconosciuto") {
				t.Fatalf("%s: identificazione sbagliata, lettura: %v", id, err)
			}
		})
		emuRun(ei, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err == nil {
				t.Fatalf("%s: identificazione sbagliata, scrittura accettata", id)
			}
		})
		if !bytes.Equal(ei.image(), radio) || len(ei.log) != 0 {
			t.Fatalf("%s: radio con un'altra identificazione letta o modificata", id)
		}
		// password di programmazione (6 byte a 0x03B0, dentro la parte che si legge e si
		// scrive): si legge e si scrive lo stesso, e resta com'è
		locked := append([]byte{}, radio...)
		copy(locked[0x3B0:], []byte{12, 13, 14, 72, 21, 0xFF})
		pwThrough447(t, m, newEmu(locked), 0x3B0, []byte{12, 13, 14, 72, 21, 0xFF}, func(mem []byte) {
			if err := m.Fam.Set(m, mem, 1, nc); err != nil {
				t.Fatal(err)
			}
		})
	}
	// gamme prese dai file di configurazione dei programmi (estremi compresi)
	for _, c := range []struct {
		id string
		hz int64
		ok bool
	}{
		{"rt-rt68b", 462550000, true}, {"rt-rt68b", 462725000, true}, {"rt-rt68b", 467562500, false}, {"rt-rt68b", 446006250, false},
		{"rt-rt668b", 446006250, true}, {"rt-rt668b", 446193750, true}, {"rt-rt668b", 462562500, false},
		{"rt-rb669", 149025000, true}, {"rt-rb669", 446006250, true}, {"rt-rb669", 470000000, true}, {"rt-rb669", 480000000, false},
	} {
		if genByID[c.id].inBands(c.hz) != c.ok {
			t.Fatalf("%s: %d Hz ammessa = %v", c.id, c.hz, !c.ok)
		}
	}
	// una RT68B non si collega con la sigla della RB669, né il contrario
	a, b := genByID["rt-rt68b"], genByID["rt-rb669"]
	emuRun(b.Fam.(t18Fam).emu(b, bytes.Repeat([]byte{0xFF}, 0x3F0)), func() {
		if _, _, err := a.Fam.Download(context.Background(), a, "EMU", nop); err == nil {
			t.Fatal("RB669 letta con la sigla della RT68B")
		}
	})
	// i modelli già presenti non cambiano: tutta la memoria, conferma 06 soltanto, «b» dopo la lettura
	for _, id := range []string{"rt-rt68h", "rt-b3h", "rt-rb29", "rt-rb617", "rt-rb17"} {
		if sp := t18o(genByID[id]).spec; sp.rranges != nil || sp.ack86 || sp.noExitRead {
			t.Fatalf("%s: collegamento cambiato", id)
		}
	}
}
