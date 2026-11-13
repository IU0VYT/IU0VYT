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

// 4.26: Retevis RB39 / RB39P e RB639 / RB639P. Pacchetti e dati trascritti dal
// codice dei programmi del costruttore (tabelle di byte e metodi readInfo,
// writeInfo, getChannelCFG_HexToStr della classe Auctus_V1._0.Data.WriFreq).
func TestAuctusPackets(t *testing.T) {
	// pacchetti fissi del programma, con la loro somma di controllo
	for want, got := range map[string][]byte{
		"AA 06 01 01 08 BB":    auctusPkt(1, 1, nil),
		"AA 06 02 01 09 BB":    auctusPkt(2, 1, nil),
		"AA 06 02 00 08 BB":    auctusPkt(2, 0, nil),
		"AA 06 04 01 0B BB":    auctusPkt(4, 1, nil),
		"AA 06 07 01 0E BB":    auctusPkt(7, 1, nil),
		"AA 06 07 00 0D BB":    auctusPkt(7, 0, nil),
		"AA 07 03 01 00 0B BB": auctusPkt(3, 1, []byte{0}),
		"AA 07 03 01 15 20 BB": auctusPkt(3, 1, []byte{21}),
	} {
		if s := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(strings.Trim(strings.ReplaceAll(fmtHex(got), "  ", " "), " "), "\n", " "))); s != want {
			t.Fatalf("pacchetto %s invece di %s", s, want)
		}
	}
	if p := auctusPkt(3, 0, make([]byte, 16)); len(p) != 0x16 || p[1] != 0x16 || p[21] != 0xBB {
		t.Fatalf("scrittura del canale: % X", p)
	}
	if p := auctusPkt(4, 0, make([]byte, 9)); len(p) != 0x0F || p[1] != 0x0F {
		t.Fatalf("scrittura dei dati di sistema: % X", p)
	}
	// tabelle dei toni: 50 CTCSS dopo «nessuno», 104 codici DCS
	if len(auctusCTCSS) != 51 || auctusCTCSS[1] != "67.0" || auctusCTCSS[50] != "254.1" || len(auctusDCS) != 104 || auctusDCS[0] != 23 || auctusDCS[103] != 754 {
		t.Fatalf("tabelle dei toni: %d e %d", len(auctusCTCSS), len(auctusDCS))
	}
}

func fmtHex(b []byte) string {
	var sb strings.Builder
	for i, x := range b {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strings.ToUpper(strings.TrimSpace(string("0123456789ABCDEF"[x>>4]) + string("0123456789ABCDEF"[x&15]))))
	}
	return sb.String()
}

func TestAuctusRadio(t *testing.T) {
	for _, id := range []string{"rt-rb39", "rt-rb639"} {
		m := genByID[id]
		if m == nil || m.Baud != 2400 {
			t.Fatalf("%s: modello assente o velocità errata", id)
		}
		// canale come lo scrive il programma: TX e RX in decine di Hz, prima il
		// byte più significativo; DCS = posizione + 1 con il bit del tipo
		img := make([]byte, auctusSize)
		c := &Chan{Rx: 462562500, Tx: 462562500, RxTone: "D023N", TxTone: "88.5", Power: m.Powers[1].Sem, Narrow: true}
		if err := m.Fam.Set(m, img, 2, c); err != nil {
			t.Fatal(err)
		}
		want := []byte{2, 0x02, 0xC1, 0xD0, 0x7A, 0x02, 0xC1, 0xD0, 0x7A, 1, 9, 0x01, 1, 0, 0, 1}
		if !bytes.Equal(img[32:48], want) {
			t.Fatalf("%s: canale % X\n       invece di % X", id, img[32:48], want)
		}
		if g := m.Fam.Get(m, img, 2); g == nil || g.Rx != c.Rx || g.Tx != c.Tx || g.RxTone != "D023N" || g.TxTone != "88.5" || !g.Narrow || g.Power != c.Power {
			t.Fatalf("%s: riletto %+v", id, g)
		}
		if err := m.Fam.Set(m, img, 3, &Chan{Rx: 446006250, Tx: 446006250, RxTone: "254.1", TxTone: "D754I", Power: m.Powers[0].Sem}); err != nil {
			t.Fatal(err)
		}
		if r := img[48:64]; r[9] != 50 || r[10] != 208 || r[11] != 0x02 || r[12] != 0 || r[13] != 1 {
			t.Fatalf("%s: toni e potenza % X", id, r)
		}
		if err := m.Fam.Set(m, img, 4, &Chan{Rx: 446006250, Tx: 446006250, RxTone: "D017N"}); err == nil {
			t.Fatalf("%s: tono non previsto accettato", id)
		}
		// canale spento = posizione vuota
		if err := m.Fam.Set(m, img, 2, nil); err != nil || m.Fam.Get(m, img, 2) != nil || img[32+15] != 0 {
			t.Fatalf("%s: cancellazione", id)
		}

		// radio simulata: lettura, modifica, scrittura
		radio := make([]byte, auctusSize)
		for i := 0; i < m.Slots; i++ {
			hz := int64(462562500 + i*25000)
			if err := m.Fam.Set(m, radio, i, &Chan{Rx: hz, Tx: hz, Power: m.Powers[1].Sem}); err != nil {
				t.Fatal(err)
			}
			radio[i*16+14] = 1 // blocco su canale occupato
		}
		radio[auctusSys] = 10 // versione «P»: dieci byte di sistema
		copy(radio[auctusSys+1:], []byte{3, 2, 1, 1, 0, 1, 0, 2, 1, 1})
		link := func(img []byte) (*auctusEmu, func()) {
			e := emuFor(m, img).(*auctusEmu)
			a, b := newPipe()
			old := openPortFunc
			openPortFunc = func(string, int) (Port, error) { return a, nil }
			go func() { _ = e.serve(b) }()
			return e, func() { openPortFunc = old; _ = a.Close() }
		}
		e, done := link(radio)
		got, info, err := m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
		done()
		if err != nil {
			t.Fatalf("%s: lettura: %v", id, err)
		}
		if !bytes.Equal(got[:m.Slots*16], radio[:m.Slots*16]) || !bytes.Equal(got[auctusSys:auctusVer], radio[auctusSys:auctusVer]) || !strings.Contains(info, auctusO(m).country) {
			t.Fatalf("%s: lettura diversa (%s): %s", id, info, diffRanges(radio[:auctusVer], got[:auctusVer]))
		}
		// le posizioni oltre i canali della radio restano spente
		for i := m.Slots; i < auctusSlots; i++ {
			if got[i*16+15] != 0 {
				t.Fatalf("%s: posizione %d non vuota", id, i+1)
			}
		}
		// comandi: versione, inizio lettura, un canale per posizione, sistema, fine
		logLen := func() int { e.mu.Lock(); defer e.mu.Unlock(); return len(e.log) }
		waitFor(func() bool { return logLen() == m.Slots+4 })
		if len(e.log) != m.Slots+4 || e.log[0] != 0x0101 || e.log[1] != 0x0201 || e.log[2] != 0x0301 || e.log[m.Slots+2] != 0x0401 || e.log[m.Slots+3] != 0x0701 {
			t.Fatalf("%s: comandi %X", id, e.log)
		}
		if err := m.Fam.Set(m, got, 0, &Chan{Rx: 467562500, Tx: 467562500, TxTone: "D023I", Power: m.Powers[0].Sem, Narrow: true}); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, got, 5, nil); err != nil {
			t.Fatal(err)
		}
		e, done = link(radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", got, func(int, int, string) {})
		done()
		if err != nil {
			t.Fatalf("%s: scrittura: %v", id, err)
		}
		waitFor(func() bool { return logLen() == m.Slots+4 })
		after := e.image()
		if c := m.Fam.Get(m, after, 0); c == nil || c.Rx != 467562500 || c.TxTone != "D023I" || !c.Narrow {
			t.Fatalf("%s: canale 1 dopo la scrittura: %+v", id, c)
		}
		if m.Fam.Get(m, after, 5) != nil || after[0*16+14] != 1 {
			t.Fatalf("%s: cancellazione o blocco su occupato non conservati", id)
		}
		if !bytes.Equal(after[auctusSys:auctusSys+11], radio[auctusSys:auctusSys+11]) {
			t.Fatalf("%s: dati di sistema modificati", id)
		}
		if len(e.log) != m.Slots+4 || e.log[1] != 0x0200 || e.log[2] != 0x0300 || e.log[m.Slots+2] != 0x0400 || e.log[m.Slots+3] != 0x0700 {
			t.Fatalf("%s: comandi di scrittura %X", id, e.log)
		}

		// radio dell'altro paese: la lettura si ferma e indica l'altra scheda
		e, done = link(radio)
		e.region = "AU147_" + map[string]string{"FCC": "CE", "CE": "FCC"}[auctusO(m).country]
		_, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
		// 4.31: dopo il rifiuto la radio va fatta uscire dalla programmazione (terzo comando: fine)
		for i := 0; i < 50 && len(e.logCopy()) < 3; i++ {
			time.Sleep(10 * time.Millisecond)
		}
		done()
		if lg := e.logCopy(); err == nil || !strings.Contains(err.Error(), "altro paese") || len(lg) != 3 || lg[2] != 0x0701 {
			t.Fatalf("%s: radio di un altro paese: %v (comandi %04X)", id, err, lg)
		}
		// file senza i dati di sistema (non letto dalla radio): scrittura rifiutata
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", make([]byte, auctusSize), func(int, int, string) {}); err == nil {
			t.Fatalf("%s: scrittura di un file non letto dalla radio accettata", id)
		}
	}
}

func (e *auctusEmu) logCopy() []int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]int{}, e.log...)
}
