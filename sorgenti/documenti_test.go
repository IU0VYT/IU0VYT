// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// TestCopieDocumenti (4.11): LEGGIMI.txt, AUTORI_E_LICENZE.txt e la licenza esistono
// sia nella cartella principale sia in «sorgenti» e devono restare identici. Gira
// anche su Windows; se la cartella principale non c'è (sorgenti copiati da soli) il
// controllo si salta.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCopieDocumenti(t *testing.T) {
	pairs := [][2]string{
		{"LEGGIMI.txt", "LEGGIMI.txt"},
		{"AUTORI_E_LICENZE.txt", "AUTORI_E_LICENZE.txt"},
		{"LICENSE.txt", "LICENSE"}, // nel repository
	}
	checked := 0
	for _, p := range pairs {
		here, err := os.ReadFile(p[0])
		if err != nil {
			t.Fatalf("%s mancante in sorgenti: %v", p[0], err)
		}
		root := filepath.Join("..", p[1])
		other, err := os.ReadFile(root)
		if err != nil && p[1] == "LICENSE" {
			root = filepath.Join("..", "LICENSE.txt") // nel pacchetto
			other, err = os.ReadFile(root)
		}
		if err != nil {
			continue
		}
		norm := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
		if !bytes.Equal(norm(here), norm(other)) {
			t.Errorf("sorgenti/%s è diverso da %s: le due copie devono essere identiche", p[0], filepath.ToSlash(root))
		}
		checked++
	}
	if checked == 0 {
		t.Skip("cartella principale non presente")
	}
}
