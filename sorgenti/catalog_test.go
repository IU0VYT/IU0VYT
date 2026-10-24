// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
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

// il catalogo usa solo driver esistenti, ogni driver compare nel catalogo e
// identificativi e nomi non sono ripetuti
func TestCatalog(t *testing.T) {
	used := map[string]bool{}
	ids := map[string]bool{}
	names := map[string]bool{}
	for _, c := range catalog {
		if ids[c.ID] {
			t.Errorf("identificativo ripetuto %s", c.ID)
		}
		ids[c.ID] = true
		k := strings.ToUpper(c.Brand + " " + c.Name)
		if names[k] {
			t.Errorf("nome ripetuto %s", k)
		}
		names[k] = true
		for _, d := range c.Drivers {
			if d != kindXLC && d != kindDAT && !isGenKind(d) {
				t.Errorf("%s: driver %s inesistente", c.ID, d)
			}
			used[d] = true
		}
		switch c.FW {
		case "", "anc", "bf", "rh5", "kdhx", "bj9000", "kdh9g":
		case "link":
			ok := false
			for _, l := range c.Links {
				ok = ok || l.Kind == "fw"
			}
			if !ok {
				t.Errorf("%s: firmware «link» senza collegamento", c.ID)
			}
		default:
			t.Errorf("%s: tipo di firmware %s sconosciuto", c.ID, c.FW)
		}
		for _, l := range c.Links {
			if !strings.HasPrefix(l.URL, "http") {
				t.Errorf("%s: collegamento non valido %s", c.ID, l.URL)
			}
		}
	}
	for _, m := range genModels {
		if !used[m.ID] {
			t.Errorf("driver %s (%s) assente dal catalogo", m.ID, m.Name)
		}
	}
	t.Logf("%d modelli nel catalogo, %d driver generici", len(catalog), len(genModels))
}

