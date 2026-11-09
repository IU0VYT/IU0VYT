// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

//go:build !windows

package main

import (
	"os"
	"strings"
)

// systemLangs (4.12): fuori da Windows (prove) dalle variabili LANGUAGE, LC_ALL,
// LC_MESSAGES e LANG
func systemLangs() []string {
	var out []string
	for _, k := range []string{"language", "lc_all", "lc_messages", "lang"} {
		for _, v := range strings.Split(os.Getenv(strings.ToUpper(k)), ":") {
			if v != "" && v != "C" && !strings.EqualFold(v, "posix") {
				out = append(out, v)
			}
		}
	}
	return out
}
