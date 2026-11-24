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
	"time"
)

// 4.47, Retevis B3B, B63B, RT22P, RT622P (sigla «C160RAM»): stessa memoria dei
// canali della H777 nuova versione; lettura e scrittura con la radio simulata;
// la password di programmazione non ferma niente (decisione del 23/11/2026: come
// CHIRP) e resta com'è; una radio con un'altra sigla rifiuta
func TestH777Twins447(t *testing.T) {
	h777 := genByID["rt-h777v4"]
	nop := func(int, int, string) {}
	for _, id := range []string{"rt-b3b", "rt-b63b", "rt-rt22p", "rt-rt622p", "rt-rb666", "rt-rb666p"} {
		m := genByID[id]
		cmd, ack := "C160RAM", byte(0)
		if strings.HasPrefix(id, "rt-rb666") {
			cmd, ack = "DLHGRA5", 0x66 // RB666 e RB666P: altra sigla, conferma 66, eco possibile
		}
		if m == nil || !m.NoChirp || h777v4o(m).cmd != cmd || h777v4o(m).memsize != 0xEA || h777v4o(m).bs != 13 || !h777v4o(m).pass ||
			!h777v4o(m).bye || h777v4o(m).ack != ack || h777v4o(m).echo != (ack != 0) {
			t.Fatalf("%s: modello assente o con altri parametri", id)
		}
		// memoria di partenza: canali scritti con il driver della H777, impostazioni «di fabbrica»
		radio := bytes.Repeat([]byte{0xFF}, 0xEA)
		for k := 0xD0; k < 0xDD; k++ {
			radio[k] = byte(k)
		}
		chans := []*Chan{
			{Rx: 446006250, Tx: 446006250, Narrow: true, Scan: true, Power: 0, RxTone: "88.5", TxTone: "88.5"},
			{Rx: 462562500, Tx: 467562500, Scan: false, Power: 2, RxTone: "D023N", TxTone: "D754I"},
			{Rx: 400000000, Tx: 400000000, TxOff: true, Scan: true, Power: 0},
		}
		for i, c := range chans {
			cc := *c
			if err := h777.Fam.Set(h777, radio, i*5, &cc); err != nil {
				t.Fatal(err)
			}
		}
		// posizioni dei campi, scritte a mano dalla scheda: canale 1 a 0x00, 13 byte
		if !bytes.Equal(radio[0:4], []byte{0x31, 0x8D, 0xA8, 0x02}) || radio[8] != 0x85 || radio[9] != 0x08 || radio[12]&0x1C != 0x0C {
			t.Fatalf("canale 1: % X", radio[:13])
		}
		e := m.Fam.(h777v4Fam).emu(m, radio).(*wxEmu)
		var img []byte
		emuRun(e, func() {
			var err error
			if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
				t.Fatalf("%s: lettura: %v", id, err)
			}
		})
		if !bytes.Equal(img, radio) {
			t.Fatalf("%s: lettura diversa dalla memoria della radio", id)
		}
		for i, c := range chans {
			g := m.Fam.Get(m, img, i*5)
			w := h777.Fam.Get(h777, radio, i*5)
			if g == nil || w == nil || g.Rx != c.Rx || g.RxTone != c.RxTone || g.TxTone != c.TxTone || g.Narrow != c.Narrow || g.Scan != c.Scan ||
				g.TxOff != c.TxOff || g.Rx != w.Rx || g.Tx != w.Tx || m.powerIndex(g.Power) != h777.powerIndex(w.Power) {
				t.Fatalf("%s canale %d: %+v (H777: %+v)", id, i*5+1, g, w)
			}
		}
		if m.Fam.Get(m, img, 1) != nil {
			t.Fatalf("%s: canale libero letto come occupato", id)
		}
		// modifica, cancellazione e canale nuovo; impostazioni e password intatte
		nc := &Chan{Rx: 446093750, Tx: 446093750, Narrow: true, Scan: true, Power: 2, RxTone: "123.0", TxTone: "123.0"}
		if err := m.Fam.Set(m, img, 7, nc); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, img, 5, nil); err != nil {
			t.Fatal(err)
		}
		emuRun(e, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
				t.Fatalf("%s: scrittura: %v", id, err)
			}
		})
		got := e.image()
		if !bytes.Equal(got[:0xEA], img) || !bytes.Equal(got[0xD0:0xEA], radio[0xD0:0xEA]) {
			t.Fatalf("%s: memoria della radio diversa dopo la scrittura", id)
		}
		if g := m.Fam.Get(m, got, 7); g == nil || g.Rx != nc.Rx || g.RxTone != "123.0" || m.Fam.Get(m, got, 5) != nil {
			t.Fatalf("%s: canale nuovo o cancellato non arrivati alla radio", id)
		}
		// radio con la password di programmazione: si legge e si scrive lo stesso (come
		// CHIRP), e la password resta quella della radio
		locked := append([]byte{}, radio...)
		copy(locked[0xDD:], []byte{1, 2, 3, 4, 0xFF, 0xFF})
		el := m.Fam.(h777v4Fam).emu(m, locked).(*wxEmu)
		var limg []byte
		emuRun(el, func() {
			var err error
			if limg, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
				t.Fatalf("%s: lettura con la password: %v", id, err)
			}
		})
		if !bytes.Equal(limg, locked) {
			t.Fatalf("%s: lettura con la password diversa dalla memoria della radio", id)
		}
		if err := m.Fam.Set(m, limg, 9, nc); err != nil {
			t.Fatal(err)
		}
		emuRun(el, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", limg, nop); err != nil {
				t.Fatalf("%s: scrittura con la password: %v", id, err)
			}
		})
		lgot := el.image()[:0xEA]
		if !bytes.Equal(lgot, limg) || !bytes.Equal(lgot[0xD0:], locked[0xD0:]) || !bytes.Equal(lgot[0xDD:0xE3], []byte{1, 2, 3, 4, 0xFF, 0xFF}) {
			t.Fatalf("%s: radio con la password: memoria o password diverse dopo la scrittura", id)
		}
		if g := m.Fam.Get(m, lgot, 9); g == nil || g.Rx != nc.Rx {
			t.Fatalf("%s: radio con la password: canale nuovo non arrivato", id)
		}
		// e scrivendo una memoria senza password (letta da un'altra radio) la password
		// della radio diventa quella della memoria scritta: si scrive tutta la memoria
		emuRun(el, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
				t.Fatalf("%s: scrittura di un'altra memoria sulla radio con la password: %v", id, err)
			}
		})
		if !bytes.Equal(el.image()[:0xEA], img) {
			t.Fatalf("%s: la memoria scritta non è arrivata tutta alla radio", id)
		}
		// una H777 (altra sigla) non accetta il collegamento di questo modello
		eh := h777.Fam.(h777v4Fam).emu(h777, radio)
		emuRun(eh, func() {
			if _, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err == nil {
				t.Fatalf("%s: letta una radio con un'altra sigla", id)
			}
		})
		// il dialogo, byte per byte, come nella scheda: 02 e sigla insieme, 02, 06, 05, 06,
		// i 18 blocchi e la chiusura «b»; la radio conferma con 06 oppure (RB666) con 66
		// e, con l'eco, prima di ogni risposta rimanda quello che ha ricevuto
		for _, echo := range []bool{false, true} {
			if echo && ack == 0 {
				continue // i programmi dei quattro modelli «C160RAM» non prevedono l'eco
			}
			er := m.Fam.(h777v4Fam).emu(m, radio).(*wxEmu)
			er.echo = echo
			var tx, rx []byte
			emuRunRec(er, &tx, &rx, func() {
				if g, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err != nil || !bytes.Equal(g, radio) {
					t.Fatalf("%s (eco %v): lettura: %v", id, echo, err)
				}
			})
			ak := byte(0x06)
			if ack != 0 {
				ak = ack
			}
			var wtx, wrx []byte
			step := func(send, reply []byte) {
				wtx = append(wtx, send...)
				if echo {
					wrx = append(wrx, send...)
				}
				wrx = append(wrx, reply...)
			}
			step(append([]byte{0x02}, cmd...), []byte{ak})
			step([]byte{0x02}, make([]byte, 8))
			step([]byte{0x06}, []byte{ak})
			step([]byte{0x05}, radio[0xDD:0xE3])
			step([]byte{0x06}, []byte{ak})
			for a := 0; a < 0xEA; a += 13 {
				step([]byte{'R', 0, byte(a), 13}, append([]byte{'W', 0, byte(a), 13}, radio[a:a+13]...))
			}
			wtx = append(wtx, 'b')
			if !bytes.Equal(tx, wtx) {
				t.Fatalf("%s (eco %v): byte mandati in lettura\n% X\nattesi\n% X", id, echo, tx, wtx)
			}
			if !bytes.Equal(rx, wrx) {
				t.Fatalf("%s (eco %v): byte ricevuti in lettura\n% X\nattesi\n% X", id, echo, rx, wrx)
			}
			tx, rx, wtx, wrx = nil, nil, wtx[:0:0], wrx[:0:0]
			emuRunRec(er, &tx, &rx, func() {
				if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
					t.Fatalf("%s (eco %v): scrittura: %v", id, echo, err)
				}
			})
			step(append([]byte{0x02}, cmd...), []byte{ak})
			step([]byte{0x02}, make([]byte, 8))
			step([]byte{0x06}, []byte{ak})
			step([]byte{0x05}, radio[0xDD:0xE3])
			step([]byte{0x06}, []byte{ak})
			for a := 0; a < 0xEA; a += 13 {
				step(append([]byte{'W', 0, byte(a), 13}, img[a:a+13]...), []byte{ak})
			}
			wtx = append(wtx, 'b')
			if !bytes.Equal(tx, wtx) || !bytes.Equal(rx, wrx) {
				t.Fatalf("%s (eco %v): dialogo di scrittura diverso dalla scheda\n% X\n% X", id, echo, tx, rx)
			}
			if !bytes.Equal(er.image()[:0xEA], img) {
				t.Fatalf("%s (eco %v): memoria diversa dopo la scrittura", id, echo)
			}
		}
		if ack == 0x66 {
			// serie più vecchia (conferma 06 al collegamento, frequenze in BCD): ci si ferma
			// subito, senza leggere e senza scrivere; e una conferma 06 a un blocco non vale
			eo := m.Fam.(h777v4Fam).emu(m, radio).(*wxEmu)
			eo.ack = 0x06
			emuRun(eo, func() {
				if _, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err == nil || !strings.Contains(err.Error(), "serie più vecchia") {
					t.Fatalf("%s: serie più vecchia, lettura: %v", id, err)
				}
			})
			emuRun(eo, func() {
				if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err == nil || !strings.Contains(err.Error(), "serie più vecchia") {
					t.Fatalf("%s: serie più vecchia, scrittura: %v", id, err)
				}
			})
			if !bytes.Equal(eo.image()[:0xEA], radio) {
				t.Fatalf("%s: radio della serie più vecchia modificata", id)
			}
			// e un modello «C160RAM» non si collega a una RB666 (altra sigla), né il contrario
			b3b := genByID["rt-b3b"]
			emuRun(m.Fam.(h777v4Fam).emu(m, radio), func() {
				if _, _, err := b3b.Fam.Download(context.Background(), b3b, "EMU", nop); err == nil {
					t.Fatalf("%s: letta con la sigla della B3B", id)
				}
			})
		}
	}
	// la H777 nuova versione resta sul dialogo di CHIRP (com'era fino alla 4.46): la
	// password non si chiede (la radio simulata, a cui non è stato detto di rispondere
	// a 05, rifiuterebbe la richiesta) e, se c'è, non ferma né la lettura né la scrittura
	if h777v4o(h777).pass {
		t.Fatal("H777 nuova versione: dialogo diverso da quello di CHIRP (richiesta della password)")
	}
	free := bytes.Repeat([]byte{0xFF}, 0xEA)
	c0 := &Chan{Rx: 446006250, Tx: 446006250, Narrow: true, Scan: true}
	if err := h777.Fam.Set(h777, free, 0, c0); err != nil {
		t.Fatal(err)
	}
	ef := h777.Fam.(h777v4Fam).emu(h777, free).(*wxEmu)
	emuRun(ef, func() {
		if img, _, err := h777.Fam.Download(context.Background(), h777, "EMU", nop); err != nil || !bytes.Equal(img, free) {
			t.Fatalf("H777 senza password: %v", err)
		}
	})
	lock := append([]byte{}, free...)
	copy(lock[0xDD:], []byte{10, 11, 12, 0xFF, 0xFF, 0xFF})
	el := h777.Fam.(h777v4Fam).emu(h777, lock).(*wxEmu)
	var limg []byte
	emuRun(el, func() {
		var err error
		if limg, _, err = h777.Fam.Download(context.Background(), h777, "EMU", nop); err != nil || !bytes.Equal(limg, lock) {
			t.Fatalf("H777 con la password, lettura: %v", err)
		}
	})
	c1 := &Chan{Rx: 446018750, Tx: 446018750, Narrow: true, Scan: true}
	if err := h777.Fam.Set(h777, limg, 1, c1); err != nil {
		t.Fatal(err)
	}
	emuRun(el, func() {
		if _, err := h777.Fam.Upload(context.Background(), h777, "EMU", limg, nop); err != nil {
			t.Fatalf("H777 con la password, scrittura: %v", err)
		}
	})
	if got := el.image()[:0xEA]; !bytes.Equal(got, limg) || !bytes.Equal(got[0xDD:0xE3], []byte{10, 11, 12, 0xFF, 0xFF, 0xFF}) {
		t.Fatal("H777 con la password: memoria o password diverse dopo la scrittura")
	}
}

// portRec447: registra, dal lato del programma, i byte mandati e quelli ricevuti
type portRec447 struct {
	Port
	tx, rx *[]byte
}

func (r portRec447) Write(b []byte) error {
	*r.tx = append(*r.tx, b...)
	return r.Port.Write(b)
}

func (r portRec447) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	err := r.Port.ReadFull(ctx, b, timeout)
	if err == nil {
		*r.rx = append(*r.rx, b...)
	}
	return err
}

// emuRunRec: come emuRun, registrando il dialogo visto dal programma
func emuRunRec(e genEmu, tx, rx *[]byte, fn func()) {
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return portRec447{a, tx, rx}, nil
	}
	fn()
}
