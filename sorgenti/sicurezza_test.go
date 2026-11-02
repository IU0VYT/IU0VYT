// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "testing"

// Mappe della memoria: numeri a 32 bit accettati, oltre il limite rifiutati
// con un errore (non con un valore troncato).
func TestBWNumeriA32Bit(t *testing.T) {
	root, err := bwParse("#seekto 0x40010;\nu8 a;\nu8 b[372];\n#seek 0x10;\nu8 c;\n")
	if err != nil {
		t.Fatalf("mappa valida rifiutata: %v", err)
	}
	if got := root.F("a").Off; got != 0x40010 {
		t.Fatalf("offset di a = %#x, atteso 0x40010", got)
	}
	if got := root.F("c").Off; got != 0x40010+1+372+0x10 {
		t.Fatalf("offset di c = %#x", got)
	}
	for _, src := range []string{
		"#seekto 0x100000000;\nu8 a;\n",
		"#seek 0x100000000;\nu8 a;\n",
		"u8 a[0x100000000];\n",
	} {
		if _, err := bwParse(src); err == nil {
			t.Errorf("numero oltre 32 bit accettato: %q", src)
		}
	}
}

// Stringhe BIFF: una lunghezza dichiarata enorme in un file danneggiato deve
// dare un errore senza riservare gigabyte di memoria.
func TestBIFFStringaDanneggiata(t *testing.T) {
	// flag 0x04 (dati estesi) con lunghezza estesa 0xFFFFFFFF, poi "AB"
	rs := &biffStr{parts: [][]byte{{0x04, 0xFF, 0xFF, 0xFF, 0xFF, 'A', 'B'}}}
	if _, err := rs.str(2); err == nil {
		t.Fatal("stringa con dati estesi impossibili accettata")
	}
	// lunghezza in caratteri oltre i dati presenti
	rs = &biffStr{parts: [][]byte{{0x00, 'A', 'B'}}}
	if _, err := rs.str(65535); err == nil {
		t.Fatal("stringa troncata accettata")
	}
	// stringa valida su due record (CONTINUE con byte di flag) letta come prima
	rs = &biffStr{parts: [][]byte{{0x00, 'C', 'I'}, {0x01, 'A', 0x00, 'O', 0x00}}}
	s, err := rs.str(4)
	if err != nil || s != "CIAO" {
		t.Fatalf("stringa valida: %q, %v", s, err)
	}
}
