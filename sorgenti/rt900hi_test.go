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

// 4.8: Radtel RT-900 senza Bluetooth, canali 513-999 (firmware V1.14P): si
// rileggono e si scrivono solo se cambiano; con altri dati nella radio (firmware
// da 512 canali) non si scrive nulla
func TestRT900WriteHigh(t *testing.T) {
	m := genByID["rt-rt900"]
	o := rt900o(m)
	if o.hi != [2]int{0x4000, 0x7D00} || o.ranges[0] != [2]int{0, 0x4000} || m.Slots != 999 {
		t.Fatalf("aree: %v %v", o.hi, o.ranges)
	}
	ctx := context.Background()
	prog := func(int, int, string) {}
	blank := func() []byte { return bytes.Repeat([]byte{0xFF}, o.memsize) }
	upload := func(radio, img []byte) (*rt900Emu, error) {
		e := emuFor(m, radio).(*rt900Emu)
		a, b := newPipe()
		old := openPortFunc
		openPortFunc = func(string, int) (Port, error) { return a, nil }
		go func() { _ = e.serve(b) }()
		_, err := m.Fam.Upload(ctx, m, "EMU", img, prog)
		openPortFunc = old
		_ = a.Close()
		return e, err
	}

	// area vuota: il canale 700 si scrive
	radio := blank()
	img := append([]byte{}, radio...)
	rt950Chan(t, m, img, 0, 145500000, "UNO")
	rt950Chan(t, m, img, 699, 433500000, "SETTECENTO")
	e, err := upload(radio, img)
	if err != nil {
		t.Fatal(err)
	}
	if c := genGet(m, e.image(), 699); c == nil || c.Name != "SETTECENTO" {
		t.Fatalf("canale 700 non scritto: %+v", c)
	}
	if c := genGet(m, e.image(), 0); c == nil || c.Name != "UNO" {
		t.Fatalf("canale 1 non scritto: %+v", c)
	}

	// firmware da 512 canali con altri dati dopo il 512: il canale 700 modificato
	// non si scrive e la radio non cambia
	radio = blank()
	for a := 0x5003; a < 0x5100; a += 7 {
		radio[a] = 0x3C
	}
	e, err = upload(radio, img)
	if err == nil || !strings.Contains(err.Error(), "V1.14P") {
		t.Fatalf("scrittura in una radio con il firmware da 512 canali: %v", err)
	}
	if got := e.image(); !bytes.Equal(got[:o.memsize], radio) {
		t.Fatalf("la radio è stata modificata: %s", diffRanges(radio, got[:o.memsize]))
	}

	// stessa radio, canali solo fino al 512: si scrivono, l'area dopo il 512 resta
	img2 := append([]byte{}, radio...)
	rt950Chan(t, m, img2, 0, 145500000, "UNO")
	rt950Chan(t, m, img2, 511, 145600000, "ULTIMO512")
	e, err = upload(radio, img2)
	if err != nil {
		t.Fatal(err)
	}
	got := e.image()
	if !bytes.Equal(got[0x4000:0x7CE0], radio[0x4000:0x7CE0]) {
		t.Fatal("area dopo il canale 512 modificata")
	}
	if c := genGet(m, got, 511); c == nil || c.Name != "ULTIMO512" {
		t.Fatalf("canale 512 non scritto: %+v", c)
	}

	// radio con canali dopo il 512 (V1.14P) e un canale cancellato: si scrive
	radio = blank()
	rt950Chan(t, m, radio, 600, 438500000, "RADIO")
	rt950Chan(t, m, radio, 601, 438600000, "VIA")
	img3 := append([]byte{}, radio...)
	if err := m.Fam.Set(m, img3, 601, nil); err != nil {
		t.Fatal(err)
	}
	rt950Chan(t, m, img3, 998, 439000000, "999")
	e, err = upload(radio, img3)
	if err != nil {
		t.Fatal(err)
	}
	got = e.image()
	if c := genGet(m, got, 998); c == nil || c.Name != "999" {
		t.Fatalf("canale 999 non scritto: %+v", c)
	}
	if c := genGet(m, got, 601); c != nil {
		t.Fatalf("canale 602 non cancellato: %+v", c)
	}
	if c := genGet(m, got, 600); c == nil || c.Name != "RADIO" {
		t.Fatalf("canale 601: %+v", c)
	}
}
