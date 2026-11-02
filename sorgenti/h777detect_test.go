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

// Retevis H777: una radio che non risponde al comando della V4 («C777HAM», provato
// per primo come in CHIRP) si legge con il comando classico; una H777 V4 viene
// riconosciuta e rifiutata senza leggere.
func TestH777DetectFallback(t *testing.T) {
	m := genByID["rt-h777"]
	img := make([]byte, h777o(m).memsize)
	for i := range img {
		img[i] = 0xFF
	}
	copy(img[0x10:], []byte{0x00, 0x25, 0x46, 0x44, 0x00, 0x25, 0x46, 0x44})
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })

	// radio classica: silenzio dopo C777HAM, poi «E» e la sequenza normale
	identAttachRaw(func(p Port) {
		if b, err := emuRead(p, 8); err != nil || !bytes.Equal(b, []byte("\x02C777HAM")) {
			return
		}
		if b, err := emuByte(p); err != nil || b != 'E' {
			return
		}
		genEmuFor(m, append([]byte{}, img...)).serve(p)
	}, "")
	mem, _, err := m.Fam.Download(context.Background(), m, "COM6", func(int, int, string) {})
	if err != nil {
		t.Fatalf("lettura con il comando classico: %v", err)
	}
	if !bytes.Equal(mem[:0x18], img[:0x18]) {
		t.Fatal("memoria letta diversa")
	}

	// H777 V4: risponde al comando con un identificativo di 0xFF
	identAttachRaw(func(p Port) {
		if b, err := emuRead(p, 8); err != nil || !bytes.Equal(b, []byte("\x02C777HAM")) {
			return
		}
		p.Write([]byte{0x06})
		if b, err := emuByte(p); err != nil || b != 0x02 {
			return
		}
		p.Write(bytes.Repeat([]byte{0xFF}, 8))
		if b, err := emuByte(p); err != nil || b != 0x06 {
			return
		}
		p.Write([]byte{0x06})
	}, "")
	_, _, err = m.Fam.Download(context.Background(), m, "COM6", func(int, int, string) {})
	if err == nil || !strings.Contains(err.Error(), "V4") {
		t.Fatalf("H777 V4 non riconosciuta: %v", err)
	}
	t.Log(err)
}
