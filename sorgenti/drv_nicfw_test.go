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
	"encoding/binary"
	"sort"
	"strings"
	"testing"
)

// sessione con una radio simulata
func emuRun(e genEmu, fn func()) {
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return a, nil
	}
	fn()
}

func nicfwTestImg(be bool) []byte {
	img := bytes.Repeat([]byte{0xFF}, nicfwSize)
	if be {
		binary.BigEndian.PutUint16(img[nicfwSetAt:], nicfwMagic5)
	} else {
		binary.LittleEndian.PutUint16(img[nicfwSetAt:], nicfwMagic0)
	}
	img[0x1DFD], img[0x1DFF] = 120, 130 // potenza massima tarata UHF e VHF
	return img
}

var nicfwTestChans = []*Chan{
	{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0 Monte", Scan: true},
	{Rx: 433500000, Tx: 433500000, TxTone: "D754I", RxTone: "D023N", Name: "DCS", Narrow: true, Power: 2, Scan: true},
	{Rx: 446006250, Tx: 446006250, RxTone: "123.0", Name: "PMR 1", Power: 1, Scan: true},
	{Rx: 118100000, Tx: 118100000, TxOff: true, Name: "TWR", Scan: true},
}

// canali scritti e riletti uguali, nei due ordini dei byte; byte come da documentazione
func TestNicfwChannels(t *testing.T) {
	m := genByID["td-tdh3-nicfw"]
	for _, be := range []bool{true, false} {
		img := nicfwTestImg(be)
		for i, c := range nicfwTestChans {
			if err := m.Fam.Set(m, img, i*7, c); err != nil {
				t.Fatalf("be=%v canale %d: %v", be, i, err)
			}
			if g := m.Fam.Get(m, img, i*7); g == nil || *g != *c {
				t.Fatalf("be=%v canale %d riletto diverso:\n %+v\n %+v", be, i, g, c)
			}
		}
		if m.Fam.Get(m, img, 1) != nil {
			t.Error("un canale mai scritto risulta usato")
		}
		// primo canale: 145,600 MHz = 14560000 unità da 10 Hz = 0x00DE2B00
		want := []byte{0x00, 0xDE, 0x2B, 0x00}
		if !be {
			want = []byte{0x00, 0x2B, 0xDE, 0x00}
		}
		if got := img[0x40:0x44]; !bytes.Equal(got, want) {
			t.Errorf("be=%v frequenza scritta % X, attesa % X", be, got, want)
		}
		// secondo canale (posizione 7): DCS 754 invertito in trasmissione = 0x8000 | 0x4000 | 0o754
		r := img[0x40+7*32:]
		var tx, rx uint16
		if be {
			rx, tx = binary.BigEndian.Uint16(r[8:]), binary.BigEndian.Uint16(r[10:])
		} else {
			rx, tx = binary.LittleEndian.Uint16(r[8:]), binary.LittleEndian.Uint16(r[10:])
		}
		if tx != 0xC000|0o754 || rx != 0x8000|0o023 {
			t.Errorf("be=%v toni scritti %04X / %04X", be, tx, rx)
		}
		if r[12] != 24 || r[15]&1 != 1 {
			t.Errorf("be=%v potenza bassa %d (attesa 24 = un quinto di 120), stretta %d", be, r[12], r[15]&1)
		}
		if string(r[20:23]) != "DCS" || r[23] != 0 {
			t.Errorf("nome scritto % X", r[20:32])
		}
		// cancellazione
		if err := m.Fam.Set(m, img, 0, nil); err != nil || m.Fam.Get(m, img, 0) != nil {
			t.Errorf("canale non cancellato (%v)", err)
		}
		// una potenza già impostata sulla radio resta com'è
		c := *nicfwTestChans[2]
		img[0x40+14*32+12] = 77
		c.Name = "PMR UNO"
		if err := m.Fam.Set(m, img, 14, &c); err != nil || img[0x40+14*32+12] != 77 {
			t.Errorf("potenza cambiata: %d (%v)", img[0x40+14*32+12], err)
		}
	}
	// memoria di un altro firmware: niente canali e nessuna scrittura
	img := bytes.Repeat([]byte{0xFF}, nicfwSize)
	if m.Fam.Match(m, img) {
		t.Error("memoria senza numero magico riconosciuta")
	}
	if err := m.Fam.Set(m, img, 0, nicfwTestChans[0]); err == nil {
		t.Error("scrittura accettata su una memoria che non è di nicFW")
	}
	for _, bad := range []*Chan{{Rx: 10000000, Tx: 10000000}, {Rx: 145600005, Tx: 145600005}, {Rx: 145600000, Tx: 145600000, TxTone: "D999N"}} {
		if err := m.Fam.Set(m, nicfwTestImg(true), 0, bad); err == nil {
			t.Errorf("canale non valido accettato: %+v", bad)
		}
	}
}

// lettura e scrittura con la radio simulata: solo i blocchi cambiati, mai la taratura
func TestNicfwClone(t *testing.T) {
	m := genByID["td-tdh3-nicfw"]
	radio := nicfwTestImg(true)
	for i := nicfwCalib; i < nicfwSize; i++ {
		radio[i] = byte(i)
	}
	radio[0x1DFD], radio[0x1DFF] = 120, 130
	e := m.Fam.(nicfwFam).emu(m, radio).(*nicfwEmu)
	var img []byte
	emuRun(e, func() {
		var err error
		if img, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {}); err != nil {
			t.Fatal(err)
		}
	})
	if !bytes.Equal(img, radio) || e.stopped {
		t.Fatalf("lettura diversa dalla memoria della radio (radio ferma: %v)", e.stopped)
	}
	for i, c := range nicfwTestChans {
		if err := m.Fam.Set(m, img, i, c); err != nil {
			t.Fatal(err)
		}
	}
	// un file con un'altra taratura (letto da un'altra radio)
	for i := nicfwCalib; i < nicfwSize; i++ {
		img[i] ^= 0x55
	}
	emuRun(e, func() {
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
			t.Fatal(err)
		}
	})
	got := e.image()
	if !bytes.Equal(got[:nicfwCalib], img[:nicfwCalib]) {
		t.Error("canali non scritti")
	}
	if !bytes.Equal(got[nicfwCalib:], radio[nicfwCalib:]) {
		t.Error("la taratura della radio è stata toccata")
	}
	sort.Ints(e.writes)
	if len(e.writes) != 4 || e.writes[0] != 2 || e.writes[3] != 5 || e.reboots != 1 || e.stopped {
		t.Errorf("blocchi scritti %v, riavvii %d, radio ferma %v", e.writes, e.reboots, e.stopped)
	}
	// seconda scrittura senza modifiche: niente da scrivere, nessun riavvio
	e.writes = nil
	emuRun(e, func() {
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
			t.Fatal(err)
		}
	})
	if len(e.writes) != 0 || e.reboots != 1 {
		t.Errorf("scrittura inutile: %v, riavvii %d", e.writes, e.reboots)
	}
	// immagine di un'altra versione del firmware: rifiutata
	other := nicfwTestImg(false)
	emuRun(e, func() {
		_, err := m.Fam.Upload(context.Background(), m, "EMU", other, func(int, int, string) {})
		if err == nil || !strings.Contains(err.Error(), "versione") {
			t.Errorf("immagine di un'altra versione accettata: %v", err)
		}
	})
	if e.stopped {
		t.Error("radio lasciata ferma dopo il rifiuto")
	}
	// radio con un altro firmware: lettura rifiutata
	stock := m.Fam.(nicfwFam).emu(m, bytes.Repeat([]byte{0xFF}, nicfwSize))
	emuRun(stock, func() {
		if _, _, err := m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {}); err == nil {
			t.Error("radio senza nicFW letta senza errori")
		}
	})
}

func TestNicfw880(t *testing.T) {
	for _, id := range []string{"rtl-rt880-nicfw", "rtl-rt950pro-nicfw"} {
		m := genByID[id]
		radio := bytes.Repeat([]byte{0xFF}, nicfw880Size)
		binary.LittleEndian.PutUint16(radio[nicfw880SetAt:], 0xABC6)
		for i := nicfw880CalLo; i < nicfw880CalHi; i++ {
			radio[i] = byte(i)
		}
		e := m.Fam.(nicfw880Fam).emu(m, radio).(*nicfw880Emu)
		var img []byte
		emuRun(e, func() {
			var err error
			if img, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {}); err != nil {
				t.Fatal(err)
			}
		})
		if !bytes.Equal(img, radio) {
			t.Fatal("lettura diversa dalla memoria della radio")
		}
		for i, c := range nicfwTestChans {
			c := *c
			c.Name += " con un nome lungo"
			if err := m.Fam.Set(m, img, i*100, &c); err != nil {
				t.Fatal(err)
			}
			if g := m.Fam.Get(m, img, i*100); g == nil || *g != c {
				t.Fatalf("canale %d riletto diverso:\n %+v\n %+v", i, g, c)
			}
		}
		// primo canale dopo i tre VFO: frequenza in unità da 10 Hz, numero magico, modulazione automatica
		r := img[3*62:]
		if binary.LittleEndian.Uint32(r) != 14560000 || binary.LittleEndian.Uint16(r[24:]) != 0xABC6 || r[18] != 3 || r[21] != 255 {
			t.Errorf("canale scritto % X", r[:32])
		}
		if string(r[32:40]) != "R0 Monte" {
			t.Errorf("nome scritto %q", r[32:62])
		}
		for i := nicfw880CalLo; i < nicfw880CalHi; i++ {
			img[i] ^= 0x55
		}
		emuRun(e, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
				t.Fatal(err)
			}
		})
		got := e.image()
		if !bytes.Equal(got[:nicfw880ChanHi], img[:nicfw880ChanHi]) {
			t.Error("canali non scritti")
		}
		if !bytes.Equal(got[nicfw880ChanHi:], radio[nicfw880ChanHi:]) {
			t.Error("impostazioni o taratura toccate")
		}
		if len(e.writes) == 0 || len(e.writes) > 8 || e.reboots != 1 {
			t.Errorf("blocchi scritti %d, riavvii %d", len(e.writes), e.reboots)
		}
		if err := m.Fam.Set(m, img, 0, nil); err != nil || m.Fam.Get(m, img, 0) != nil {
			t.Errorf("canale non cancellato (%v)", err)
		}
	}
}
