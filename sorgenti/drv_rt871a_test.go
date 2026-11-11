//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"testing"
)

// 4.19, Radtel RT-817A / RT-871A: si leggono e si scrivono solo i blocchi del
// programma del costruttore (impostazioni 8-11, canali 0x10-0x3F, radio FM 0x358);
// il codice lungo del canale (byte 0x14-0x1F) si conserva
func TestRT871A(t *testing.T) {
	m := genByID["rtl-rt871a"]
	f := m.Fam.(iradioFam)
	p := f.proto()
	base := bytes.Repeat([]byte{0xFF}, p.size)
	want := []*Chan{
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0", Power: 0, Scan: true},
		{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D023N", Name: "DCS", Power: 1, Scan: true, Narrow: true},
		{Rx: 118100000, Tx: 118100000, Name: "TWR", Power: 2},
	}
	for i, c := range want {
		if err := f.Set(m, base, i*500, c); err != nil {
			t.Fatal(err)
		}
		if g := f.Get(m, base, i*500); g == nil || *g != *c {
			t.Fatalf("canale %d: scritto %+v, riletto %+v", i*500, c, g)
		}
	}
	// potenza: 0 bassa, 1 media, 2 alta (byte 0x10 del canale); gamma aeronautica in AM
	if base[0x2000+0x10] != 2 || base[0x2000+500*48+0x10] != 1 || base[0x2000+1000*48+0x10] != 0 || base[0x2000+1000*48] != 1 {
		t.Fatalf("potenze o modo: %02X %02X %02X %02X", base[0x2000+0x10], base[0x2000+500*48+0x10], base[0x2000+1000*48+0x10], base[0x2000+1000*48])
	}
	// codice lungo del canale: resta quando il canale cambia
	code := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	copy(base[0x2000+0x14:], code)
	if err := f.Set(m, base, 0, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "123.0", Name: "R0b", Power: 1, Scan: true}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(base[0x2000+0x14:0x2000+0x20], code) {
		t.Fatalf("codice del canale non conservato: % X", base[0x2000+0x14:0x2000+0x20])
	}
	// fuori dalle aree lette la radio simulata ha altri dati: non devono essere letti né scritti
	e := f.emu(m, base).(*iradioEmu)
	in := func(a int) bool {
		for _, z := range rt871aRegions {
			if a >= z[0] && a < z[0]+z[1]*p.block {
				return true
			}
		}
		return false
	}
	for a := 0; a < p.size; a += p.block {
		if !in(a) {
			for k := 0; k < p.block; k++ {
				e.img[a+k] = byte(a>>10) ^ 0x5A
			}
		}
	}
	radio := append([]byte{}, e.img...)
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	got, _, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, base) {
		t.Fatal("lettura diversa: letti blocchi fuori dalle aree del programma del costruttore?")
	}
	if err := f.Set(m, got, 7, &Chan{Rx: 446006250, Tx: 446006250, Name: "PMR 1", Power: 2, Scan: true, Narrow: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil {
		t.Fatal(err)
	}
	after := e.image()
	for a := 0; a < p.size; a += p.block {
		ref := radio
		if in(a) {
			ref = got
		}
		if !bytes.Equal(after[a:a+p.block], ref[a:a+p.block]) {
			t.Fatalf("blocco 0x%X: scritto fuori dalle aree o non scritto", a)
		}
	}
}
