// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// Conteggi dei modelli del catalogo, gli stessi della schermata «Home»: radio
// con i canali gestiti, radio con le impostazioni modificabili, modelli della
// galleria e radio con l'aggiornamento del firmware. Il sito del programma li
// mostra: a ogni modifica del ramo main il flusso .github/workflows/sito.yml li
// scrive in versione.txt con TestConteggiVersione e pubblica il file insieme alla
// cartella docs.

type catCounts struct {
	Radio        int // radio con i canali gestiti (almeno un driver)
	Impostazioni int // radio con le impostazioni modificabili
	Galleria     int // modelli nella galleria (compatibili e non)
	Firmware     int // radio con l'aggiornamento del firmware fatto dal programma
}

func catalogCounts() catCounts {
	n := catCounts{Galleria: len(catalog)}
	for _, c := range catalog {
		if len(c.Drivers) > 0 {
			n.Radio++
		}
		for _, d := range c.Drivers {
			if settingsAvailable(d) {
				n.Impostazioni++
				break
			}
		}
		switch c.FW {
		case "anc", "bf", "rh5", "kdhx", "bj9000", "kdh9g": // come FW_KINDS in ui/gallery.js («link» escluso)
			n.Firmware++
		}
	}
	return n
}

// Con IU0VYT_SCRIVI_CONTEGGI=1 il test scrive in ../versione.txt i conteggi del
// catalogo (righe radio=, impostazioni=, galleria=, firmware=) lasciando com'è il
// resto del file; senza la variabile controlla quelli presenti (se ci sono):
//
//	IU0VYT_SCRIVI_CONTEGGI=1 go test -mod=vendor -count=1 -run TestConteggiVersione .
func TestConteggiVersione(t *testing.T) {
	const path = "../versione.txt"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("versione.txt assente: ", err)
	}
	n := catalogCounts()
	want := []struct {
		key string
		val int
	}{{"radio", n.Radio}, {"impostazioni", n.Impostazioni}, {"galleria", n.Galleria}, {"firmware", n.Firmware}}
	t.Logf("radio=%d impostazioni=%d galleria=%d firmware=%d", n.Radio, n.Impostazioni, n.Galleria, n.Firmware)

	got := map[string]string{}
	var keep []string // righe diverse dai conteggi, nell'ordine originale
	isCount := map[string]bool{}
	for _, w := range want {
		isCount[w.key] = true
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		if m := reKeyValue.FindStringSubmatch(line); m != nil && isCount[strings.ToLower(m[1])] {
			got[strings.ToLower(m[1])] = m[2]
			continue
		}
		if strings.TrimSpace(line) != "" {
			keep = append(keep, line)
		}
	}

	if os.Getenv("IU0VYT_SCRIVI_CONTEGGI") == "1" {
		out := strings.Join(keep, "\n") + "\n"
		for _, w := range want {
			out += w.key + "=" + strconv.Itoa(w.val) + "\n"
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if len(got) == 0 {
		t.Skip("versione.txt senza conteggi (li scrive il flusso del sito)")
	}
	for _, w := range want {
		if got[w.key] != strconv.Itoa(w.val) {
			t.Errorf("versione.txt: %s=%q, nel catalogo %d (aggiornalo con IU0VYT_SCRIVI_CONTEGGI=1)", w.key, got[w.key], w.val)
		}
	}
}
