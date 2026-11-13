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

// Modelli del programma «TIDRadioCPS» del 2026 (TD-H3 Plus / SE, H7 Plus, H8 di
// quarta generazione, H9, 638UV).
func TestTdh8Cps26(t *testing.T) {
	// identificativi della tabella del programma (8 byte, scritti lì in esadecimale)
	idents := map[string]string{
		"td-h3plus": "503331313833FFFF", "td-h3plusham": "503331313835FFFF", "td-h3plusgmrs": "503331313834FFFF",
		"td-h7plus": "54444837504C554E", "td-h7plusham": "54444837504C5548", "td-h7plusgmrs": "54444837504C5547",
		"td-h8g4": "544448383447454E", "td-h8g4ham": "5444483834474548", "td-h8g4gmrs": "5444483834474547",
		"td-h9": "54444839FFFFFF4E", "td-h9ham": "54444839FFFFFF48", "td-h9gmrs": "54444839FFFFFF47",
		"td-638uv": "544436333855564E", "td-638uvham": "5444363338555648", "td-638uvgmrs": "5444363338555647",
	}
	np := func(int, int, string) {}
	for id, hx := range idents {
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		o := tdh8o(m)
		if strings.ToUpper(hexStr(o.ident)) != hx || !o.cps26 || o.memsize != 0x1F30 || !m.NoChirp || m.Slots != 199 || len(m.Powers) != 3 ||
			!bytes.Equal(o.magic, []byte{0x50, 0x56, 0x4F, 0x4A, 0x48, 0x5C, 0x14}) || settingsAvailable(id) {
			t.Fatalf("%s: dati del modello errati (% X)", id, o.ident)
		}
		size := 8 + 0x1F40
		img := append(append([]byte{}, o.ident...), make([]byte, size-8)...)
		// Righe dei file di fabbrica del programma (indirizzo della radio = 8 + indirizzo):
		//   canale 1 a 0x0010: 50 62 25 46 | 50 62 25 46 | FF FF | FF FF | 00 00 20 00  = 462,5625 MHz, potenza alta, banda larga
		//   canale 2 a 0x0020: 50 62 15 44 | 50 62 15 44 | 44 07 | 44 07 | 05 E4 18 02  = 441,5625 MHz, tono 74,4, media, stretta,
		//                      con scrambler 5, PTT-ID 3, salto, blocco su occupato e modo di ricezione 2
		//   canale 3 a 0x0030: vuoto come lo scrive il programma (frequenze a zero)
		copy(img[8+0x10:], []byte{0x50, 0x62, 0x25, 0x46, 0x50, 0x62, 0x25, 0x46, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x20, 0x00})
		copy(img[8+0x20:], []byte{0x50, 0x62, 0x15, 0x44, 0x50, 0x62, 0x15, 0x44, 0x44, 0x07, 0x44, 0x07, 0x05, 0xE4, 0x18, 0x02})
		copy(img[8+0x30:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0})
		copy(img[8+0xD40:], "GMRS1\xFF\xFF\xFF  PONTE\x00")
		img[8+0x1900], img[8+0x1920] = 0x07, 0x01 // usati 1-3; in scansione solo l'1
		// 462,5625 MHz è fuori dalla gamma di trasmissione dei radioamatori: nella versione HAM è di sola ricezione
		ham := strings.HasSuffix(id, "ham")
		c1, c2 := m.Fam.Get(m, img, 0), m.Fam.Get(m, img, 1)
		if c1 == nil || c1.Rx != 462562500 || c1.Tx != 462562500 || c1.Power != 0 || c1.Narrow || !c1.Scan || c1.Name != "GMRS1" || c1.RxTone != "" || c1.TxOff != ham {
			t.Fatalf("%s: canale 1 %+v", id, c1)
		}
		if c2 == nil || c2.Rx != 441562500 || c2.RxTone != "74.4" || c2.TxTone != "74.4" || c2.Power != 1 || !c2.Narrow || c2.Scan || c2.Name != "  PONTE" || c2.TxOff {
			t.Fatalf("%s: canale 2 %+v", id, c2)
		}
		if m.Fam.Get(m, img, 2) != nil || m.Fam.Get(m, img, 3) != nil {
			t.Fatalf("%s: canale vuoto letto come pieno", id)
		}
		// modifica: i campi che non si gestiscono restano (byte 12, bit 7-5 e 2 del byte 13, byte 15)
		c2.Tx, c2.RxTone, c2.Power, c2.Narrow, c2.Scan, c2.Name = 446562500, "D023N", 2, false, true, "R1"
		if err := m.Fam.Set(m, img, 1, c2); err != nil {
			t.Fatal(err)
		}
		// 446,5625 → 50 62 65 44; bassa = 0 nei bit 4-5 del byte 14, larga = bit 3 a zero
		want := []byte{0x50, 0x62, 0x15, 0x44, 0x50, 0x62, 0x65, 0x44, 0x23, 0x80, 0x44, 0x07, 0x05, 0xE4, 0x00, 0x02}
		if !bytes.Equal(img[8+0x20:8+0x30], want) || img[8+0x1920] != 0x03 || string(img[8+0xD48:8+0xD50]) != "R1\x00\x00\x00\x00\x00\x00" {
			t.Fatalf("%s: canale 2 dopo la modifica % X", id, img[8+0x20:8+0x30])
		}
		// canale nuovo in un posto vuoto: campi in più a zero, segnato come usato
		if err := m.Fam.Set(m, img, 3, &Chan{Rx: 145500000, Tx: 145500000, Power: 0}); err != nil || img[8+0x1900] != 0x0F ||
			!bytes.Equal(img[8+0x40:8+0x50], []byte{0x00, 0x00, 0x55, 0x14, 0x00, 0x00, 0x55, 0x14, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x20, 0x00}) {
			t.Fatalf("%s: canale nuovo % X (%v)", id, img[8+0x40:8+0x50], err)
		}
		// banda aerea e, nella versione HAM, tutto ciò che è fuori da 144-148 / 420-450 MHz: solo ricezione
		if err := m.Fam.Set(m, img, 4, &Chan{Rx: 118100000, Tx: 118100000}); err != nil || !m.Fam.Get(m, img, 4).TxOff {
			t.Fatalf("%s: banda aerea non in sola ricezione (%v)", id, err)
		}
		_ = m.Fam.Set(m, img, 5, &Chan{Rx: 462562500, Tx: 462562500})
		if g := m.Fam.Get(m, img, 5); g.TxOff != ham {
			t.Fatalf("%s: 462,5625 MHz sola ricezione = %v", id, g.TxOff)
		}
		if m.inBands(700000000) || m.inBands(190000000) || !m.inBands(108000000) || !m.inBands(600000000) {
			t.Fatalf("%s: limiti di frequenza errati", id)
		}
		// cancellazione come il programma
		if err := m.Fam.Set(m, img, 0, nil); err != nil || m.Fam.Get(m, img, 0) != nil || img[8+0x1900]&1 != 0 || img[8+0x1920]&1 != 0 ||
			!bytes.Equal(img[8+0x10:8+0x20], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0}) {
			t.Fatalf("%s: cancellazione % X", id, img[8+0x10:8+0x20])
		}

		// radio simulata: lettura e scrittura di 0x0000-0x1F3F, senza lettura della versione, chiusura «CE»
		radio := append([]byte{}, img...)
		for k := 8 + 0x1F30; k < len(radio); k++ {
			radio[k] = byte(k)
		}
		link := func(mm *genModel, mem []byte) (*tdh8Emu, func()) {
			e := emuFor(mm, mem).(*tdh8Emu)
			a, b := newPipe()
			old := openPortFunc
			openPortFunc = func(string, int) (Port, error) { return a, nil }
			go func() { _ = e.serve(b) }()
			return e, func() { openPortFunc = old; _ = a.Close() }
		}
		ended := func(e *tdh8Emu) byte { e.mu.Lock(); defer e.mu.Unlock(); return e.end }
		e, done := link(m, radio)
		got, _, err := m.Fam.Download(context.Background(), m, "EMU", np)
		waitFor(func() bool { return ended(e) != 0 })
		done()
		if err != nil || !bytes.Equal(got, radio) || len(got) != size || ended(e) != 'C' {
			t.Fatalf("%s: lettura: %v, %d byte, chiusura %q", id, err, len(got), ended(e))
		}
		_ = m.Fam.Set(m, got, 10, &Chan{Rx: 433500000, Tx: 433500000, Name: "PROVA"})
		e, done = link(m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", got, np)
		waitFor(func() bool { return ended(e) != 0 })
		done()
		if after := e.image(); err != nil || !bytes.Equal(after, got) || ended(e) != 'C' || e.fwReads != 0 {
			t.Fatalf("%s: scrittura: %v, chiusura %q, letture della versione %d", id, err, ended(e), e.fwReads)
		}
	}
	// versione sbagliata: la radio HAM non viene letta con la scheda base, e il messaggio dice come si presenta
	a, b := genByID["td-h9"], genByID["td-h9ham"]
	e := emuFor(b, append(append([]byte{}, tdh8o(b).ident...), make([]byte, 0x1F40)...)).(*tdh8Emu)
	pa, pb := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return pa, nil }
	go func() { _ = e.serve(pb) }()
	_, _, err := a.Fam.Download(context.Background(), a, "EMU", np)
	openPortFunc = old
	_ = pa.Close()
	if err == nil || !strings.Contains(err.Error(), "TDH9") {
		t.Fatalf("radio HAM letta con la scheda base: %v", err)
	}
	// i modelli di CHIRP non cambiano: chiusura 'E', lettura della versione prima di scrivere, cancellazione a FF
	h3 := genByID["td-tdh3"]
	if o := tdh8o(h3); o.cps26 || !o.fwRead || o.memsize != 0x1fef {
		t.Fatal("TD-H3 di CHIRP modificata")
	}
	img := append(append([]byte{}, tdh8o(h3).ident...), bytes.Repeat([]byte{0xFF}, 0x2000)...)
	_ = h3.Fam.Set(h3, img, 0, &Chan{Rx: 145500000, Tx: 145500000})
	_ = h3.Fam.Set(h3, img, 0, nil)
	if !bytes.Equal(img[8+0x10:8+0x20], bytes.Repeat([]byte{0xFF}, 16)) {
		t.Fatalf("TD-H3 di CHIRP: cancellazione % X", img[8+0x10:8+0x20])
	}
}
