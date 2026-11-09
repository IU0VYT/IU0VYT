// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "testing"

// TestSysLang (4.12): al primo avvio la lingua del sistema, poi quella scelta
func TestSysLang(t *testing.T) {
	for _, c := range []struct {
		prefs []string
		want  string
	}{
		{[]string{"it-IT"}, "it"},
		{[]string{"de-DE", "en-US"}, "de"},
		{[]string{"es-MX"}, "es"},
		{[]string{"pt-BR"}, "pt"},
		{[]string{"zh-Hans-CN"}, "zh"},
		{[]string{"ja-JP"}, "ja"},
		{[]string{"ko-KR"}, "ko"},
		{[]string{"ar-SA"}, "ar"},
		{[]string{"nl-NL", "fr-BE"}, "fr"}, // la prima lingua che il programma ha
		{[]string{"nl-NL"}, "en"},          // nessuna: inglese
		{[]string{"ru_RU.UTF-8"}, "ru"},    // forma di Linux
		{nil, "en"},
	} {
		if got := sysLangPick(c.prefs); got != c.want {
			t.Errorf("%v: %q, atteso %q", c.prefs, got, c.want)
		}
	}
	if got := startLang("fr"); got != "fr" {
		t.Errorf("lingua scelta: %q, atteso fr", got)
	}
	if got := startLang("it"); got != "it" {
		t.Errorf("italiano scelto: %q, atteso it", got)
	}
	if got, want := startLang(""), sysLangPick(systemLangs()); got != want {
		t.Errorf("primo avvio: %q, atteso %q (lingua del sistema)", got, want)
	}
}
