//go:build !windows

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

// 4.6: nomi esatti delle radio con la stessa gestione (es. Radtel RT-950 Pro e
// Bajeton BJ-9000, Baofeng K67 Pro e 5RH Pro)
func TestModelNames(t *testing.T) {
	a := newApp(mustStore(t))
	res := func(kind string) *identResult {
		m := genByID[kind]
		r := &identResult{Found: true, Kind: kind, Name: kindBaseName(kind)}
		if m != nil {
			r.Brand = m.Vendor
		}
		r.add("Marca e modello", "%s", r.Name)
		return r
	}
	// nessuna scelta precedente: si chiede quale sia
	r := res("bj-9000")
	a.identCardNames(r)
	if !r.Ambiguous || r.Name != "Bajeton BJ-9000 (firmware 0.24 e successivi: 990 canali)" || r.Card != "bj-9000" {
		t.Fatalf("senza scelta: %+v", r)
	}
	found := false
	for _, o := range r.Others {
		found = found || o.Kind == "bj-9000#rtl-rt-950-pro" && o.Name == "Radtel RT-950 Pro (firmware 0.24 e successivi: 990 canali)"
	}
	if !found {
		t.Fatalf("la Radtel RT-950 Pro non è tra le scelte: %+v", r.Others)
	}
	// scelta la RT-950 Pro: da qui in poi si chiama così
	a.rememberCard("bj-9000", "rtl-rt-950-pro")
	r = res("bj-9000")
	a.identCardNames(r)
	if r.Ambiguous || r.Name != "Radtel RT-950 Pro (firmware 0.24 e successivi: 990 canali)" || r.Brand != "Radtel" {
		t.Fatalf("dopo la scelta: %+v", r)
	}
	for _, f := range r.Fields {
		if f.L == "Marca e modello" && !strings.HasPrefix(f.V, "Radtel RT-950 Pro") {
			t.Fatalf("campo «Marca e modello»: %q", f.V)
		}
	}
	// l'altra versione (960 canali) ha la sua scelta, indipendente
	r = res("bj-9000-960")
	a.identCardNames(r)
	if !r.Ambiguous || !strings.Contains(r.Name, "960 canali") {
		t.Fatalf("versione con 960 canali: %+v", r)
	}
	// K67 Pro / 5RH Pro: radio diverse con la stessa gestione
	r = res(kindXLC)
	a.identCardNames(r)
	if r.Name != "Baofeng K67 Pro" || !r.Ambiguous {
		t.Fatalf("K67: %+v", r)
	}
	a.rememberCard(kindXLC, "5rh-pro")
	r = res(kindXLC)
	a.identCardNames(r)
	if r.Name != "Baofeng 5RH Pro" || r.Ambiguous {
		t.Fatalf("5RH Pro scelta: %+v", r)
	}
	// una sola radio con questa gestione: nome sicuro
	r = res("rt-rt900")
	a.identCardNames(r)
	if r.Ambiguous || r.Name != "Radtel RT-900" {
		t.Fatalf("RT-900: %+v", r)
	}
	// scelta non valida per la gestione: ignorata
	a.rememberCard("rt-rt900", "rtl-rt-950-pro")
	if a.store.Get().ModelCards["rt-rt900"] != "" {
		t.Fatal("scelta di una radio con un'altra gestione salvata")
	}
	if k, c := splitKindCard("bj-9000#rtl-rt-950-pro"); k != "bj-9000" || c != "rtl-rt-950-pro" {
		t.Fatal(k, c)
	}
}
