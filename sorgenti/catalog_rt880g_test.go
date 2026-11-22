// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"strings"
	"testing"
)

// 4.46: la Radtel RT-880G è un ricetrasmettitore radioamatoriale VHF/UHF, non una
// radio GMRS. Il tipo del catalogo decide l'avviso «Radio venduta come GMRS» prima
// della scrittura dei ponti (radioCategory in ui/archive.js): le schede che usano
// il driver della RT-880G (RT-880G, RT-880, RT-860) devono essere tutte «analog» e
// senza la nota sulle radio per gli Stati Uniti.
func TestRT880GNonGMRS(t *testing.T) {
	found := 0
	for _, c := range catalog {
		uses := false
		for _, d := range c.Drivers {
			if d == "rtl-rt880g" {
				uses = true
			}
		}
		if !uses {
			continue
		}
		found++
		if c.Type != "analog" {
			t.Errorf("%s: tipo %q, atteso \"analog\" (radio radioamatoriale)", c.ID, c.Type)
		}
		if strings.Contains(c.Notes, strings.TrimSpace(usNote)) {
			t.Errorf("%s: la nota dice che è una radio per gli Stati Uniti", c.ID)
		}
	}
	if found < 3 {
		t.Errorf("schede con il driver rtl-rt880g: %d, attese almeno 3 (RT-880G, RT-880, RT-860)", found)
	}
}
