// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"regexp"
	"strings"
	"testing"
)

// 4.42: la scheda Home deve stare tutta nella finestra. Il comportamento vero si
// prova in un browser (verifica/ui/home_misura.py e home_scenari.py); qui si
// controlla che i pezzi da cui dipende non vengano tolti per sbaglio.
func TestHome442(t *testing.T) {
	rd := func(n string) string {
		b, err := uiFiles.ReadFile("ui/" + n)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	css, js, html := rd("style.css"), rd("home.js"), rd("index.html")
	// l'ultima regola di «.home» non deve farla restringere (era la causa del taglio)
	if !regexp.MustCompile(`(?m)^\.home \{ flex: none; \}`).MatchString(css) {
		t.Error("style.css: manca «.home { flex: none; }»")
	}
	if !strings.Contains(css, ".page#page-home { overflow: auto;") {
		t.Error("style.css: la pagina Home deve poter scorrere")
	}
	if !regexp.MustCompile(`\.hb-home \.hb-grid \{ grid-template-columns: repeat\(auto-fill, minmax\(\d+px, 1fr\)\)`).MatchString(css) {
		t.Error("style.css: il riquadro delle marche della Home deve avere le colonne automatiche")
	}
	if !regexp.MustCompile(`(?m)^\.neon \{ direction: ltr; \}`).MatchString(css) {
		t.Error("style.css: il titolo si scrive sempre da sinistra a destra")
	}
	for _, s := range []string{"fit() {", "watchFit(page) {", "inner.style.zoom", "page.scrollHeight <= page.clientHeight", "new ResizeObserver(", "this.watchFit(page);", "if (this.built) { this.fit(); return; }"} {
		if !strings.Contains(js, s) {
			t.Errorf("home.js: manca %q", s)
		}
	}
	if m := regexp.MustCompile(`minZoom: (0\.\d+),`).FindStringSubmatch(js); m == nil || m[1] < "0.5" || m[1] > "0.8" {
		t.Errorf("home.js: riduzione minima %v", m)
	}
	if !strings.Contains(html, `<link rel="icon" type="image/png" href="icona.png">`) {
		t.Error("index.html: manca l'icona della pagina")
	}
	if _, err := uiFiles.ReadFile("ui/icona.png"); err != nil {
		t.Error(err)
	}
}
