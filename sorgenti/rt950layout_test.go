//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.6: Bajeton BJ-9000 / Radtel RT-950 Pro con il firmware fino a 0.23 (960
// canali, 15 zone da 64) e dalla 0.24 (990 canali, 10 zone da 99): versione
// riconosciuta dai dati, scelta della scheda giusta, scrittura dei canali 961-990
// solo dopo averli riletti dalla radio.

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"
)

func rt950Blank() []byte { return bytes.Repeat([]byte{0xFF}, 0xD300) }

func rt950Chan(t *testing.T, m *genModel, img []byte, i int, mhz int64, name string) {
	t.Helper()
	c := &Chan{Rx: mhz, Tx: mhz, Name: name, Power: m.Powers[0].Sem}
	if err := m.Fam.Set(m, img, i, c); err != nil {
		t.Fatal(err)
	}
}

func TestRT950Layout(t *testing.T) {
	m990, m960 := genByID["bj-9000"], genByID["bj-9000-960"]
	if m990.Slots != 990 || m990.Zones != 10 || m960.Slots != 960 || m960.Zones != 15 {
		t.Fatalf("versioni: %d/%d e %d/%d", m990.Slots, m990.Zones, m960.Slots, m960.Zones)
	}
	// vuota: non si capisce
	img := rt950Blank()
	if l, _ := rt950Layout(img); l != 0 {
		t.Fatalf("memoria vuota: %d", l)
	}
	if m, why := genPickLayout(m960, img); m != m960 || !strings.Contains(why, "non si capisce") {
		t.Fatalf("memoria vuota: %s %q", m.ID, why)
	}
	// canale 975: firmware 0.24
	rt950Chan(t, m990, img, 974, 145500000, "Z10")
	if l, why := rt950Layout(img); l != 990 {
		t.Fatalf("canale 975: %d %s", l, why)
	}
	if m, _ := genPickLayout(m960, img); m != m990 {
		t.Fatalf("canale 975: scelta %s", m.ID)
	}
	// dati che non sono canali dopo il 960 (come quelli del firmware precedente)
	img = rt950Blank()
	for a := 0x7A87; a < 0x7AD3; a += 3 {
		img[a] = 0x03
	}
	if l, why := rt950Layout(img); l != 960 {
		t.Fatalf("area di servizio: %d %s", l, why)
	}
	if m, _ := genPickLayout(m990, img); m != m960 {
		t.Fatalf("area di servizio: scelta %s", m.ID)
	}
	// zona attiva oltre la decima
	img = rt950Blank()
	img[0x902E] = 12
	if l, _ := rt950Layout(img); l != 960 {
		t.Fatalf("zona 13: %d", l)
	}
	// record azzerati: posizioni vuote
	img = rt950Blank()
	for a := 0x7800; a < 0x7BC0; a++ {
		img[a] = 0
	}
	if l, _ := rt950Layout(img); l != 0 {
		t.Fatalf("record azzerati: %d", l)
	}
	// file .img: la versione scritta nei metadati prevale
	img = rt950Blank()
	f := makeImg(m960, img, nil)
	if m, _, _, err := detectGen(f, ""); err != nil || m != m960 {
		t.Fatalf("file della versione con 960 canali aperto come %v (%v)", m, err)
	}
	f = makeImg(m990, img, nil)
	if m, _, _, err := detectGen(f, "bj-9000-960"); err != nil || m != m990 {
		t.Fatalf("file della versione con 990 canali aperto come %v (%v)", m, err)
	}
	// file senza l'indicazione (versioni precedenti del programma): dai dati,
	// altrimenti la scheda in cui si apre
	oldFile := append(append(rt950Blank(), chirpMagic...),
		[]byte(base64.StdEncoding.EncodeToString([]byte(`{"rclass": "RT950Pro", "vendor": "Bajeton", "model": "BJ-9000"}`)))...)
	if m, _, _, err := detectGen(oldFile, "bj-9000-960"); err != nil || m != m960 {
		t.Fatalf("file senza versione aperto nella scheda con 960 canali come %v (%v)", m, err)
	}
	if m, _, _, err := detectGen(oldFile, ""); err != nil || m != m990 {
		t.Fatalf("file senza versione aperto come %v (%v)", m, err)
	}
	mem, _ := splitImg(oldFile)
	rt950Chan(t, m990, mem, 980, 145500000, "Z10")
	oldFile = append(append(mem, chirpMagic...), oldFile[len(mem)+len(chirpMagic):]...)
	if m, _, _, err := detectGen(oldFile, "bj-9000-960"); err != nil || m != m990 {
		t.Fatalf("file senza versione con il canale 981 aperto come %v (%v)", m, err)
	}
}

// scrittura dei canali 961-990 nella radio simulata
func TestRT950WriteHigh(t *testing.T) {
	m990 := genByID["bj-9000"]
	ctx := context.Background()
	prog := func(int, int, string) {}

	// radio con il firmware 0.24 (area vuota): il canale 975 si scrive
	radio := rt950Blank()
	img := append([]byte{}, radio...)
	rt950Chan(t, m990, img, 0, 145500000, "UNO")
	rt950Chan(t, m990, img, 974, 433500000, "Z10")
	e := emuFor(m990, radio)
	goSession(t, m990, e, func() error { _, err := m990.Fam.Upload(ctx, m990, "EMU", img, prog); return err })
	got := e.(*rt900Emu).image()
	if c := genGet(m990, got, 974); c == nil || c.Name != "Z10" {
		t.Fatalf("canale 975 non scritto: %+v", c)
	}
	if c := genGet(m990, got, 0); c == nil || c.Name != "UNO" {
		t.Fatalf("canale 1 non scritto: %+v", c)
	}

	// radio con dati che non sono canali dopo il 960: nessun dato scritto
	radio = rt950Blank()
	for a := 0x7A87; a < 0x7AD3; a += 3 {
		radio[a] = 0x03
	}
	e = emuFor(m990, radio)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	_, err := m990.Fam.Upload(ctx, m990, "EMU", img, prog)
	openPortFunc = old
	_ = a.Close()
	if err == nil || !strings.Contains(err.Error(), "fino a 0.23") {
		t.Fatalf("scrittura in una radio con il firmware precedente: %v", err)
	}
	if got := e.(*rt900Emu).image(); !bytes.Equal(got[:0xD300], radio) {
		t.Fatalf("la radio è stata modificata: %s", diffRanges(radio, got[:0xD300]))
	}

	// stessa radio, nessun canale oltre il 960: si scrivono i canali 1-960 e
	// l'area dopo il 960 resta quella della radio
	img2 := append([]byte{}, radio...)
	rt950Chan(t, m990, img2, 0, 145500000, "UNO")
	e = emuFor(m990, radio)
	goSession(t, m990, e, func() error { _, err := m990.Fam.Upload(ctx, m990, "EMU", img2, prog); return err })
	got = e.(*rt900Emu).image()
	if !bytes.Equal(got[0x7800:0x7C00], radio[0x7800:0x7C00]) {
		t.Fatal("area dopo il canale 960 modificata")
	}
	if c := genGet(m990, got, 0); c == nil || c.Name != "UNO" {
		t.Fatalf("canale 1 non scritto: %+v", c)
	}

	// lettura: si legge fino a 0x7BFF e la versione si riconosce
	radio = rt950Blank()
	rt950Chan(t, m990, radio, 989, 438500000, "ULTIMO")
	e = emuFor(m990, radio)
	var mem []byte
	m960 := genByID["bj-9000-960"]
	goSession(t, m960, e, func() error {
		var err error
		mem, _, err = m960.Fam.Download(ctx, m960, "EMU", prog)
		return err
	})
	if m, _ := genPickLayout(m960, mem); m != m990 {
		t.Fatalf("radio con il canale 990 letta come %s", m.ID)
	}
	if c := genGet(m990, mem, 989); c == nil || c.Name != "ULTIMO" {
		t.Fatalf("canale 990 non letto: %+v", c)
	}
}
