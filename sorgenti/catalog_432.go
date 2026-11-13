// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.32: TIDRADIO TD-M8S e TD-V2 (drv_t17ana.go; la Baofeng T22 in
// versione FRS è una seconda scheda della voce «T22» di catalog.go) e TD-H4 e
// TD-H11 del programma «TIDRadioCPS» (drv_tdh8.go, opzione cps26).

const ana432 = "Canali senza nome e impostazioni generali. Non è in CHIRP." + cps425

var catalog432 = []catItem{
	{ID: "tid-td-m8s", Brand: "TIDRADIO", Name: "TD-M8S", Aliases: []string{"TD-M8s", "M8S", "TDM8S"}, Models: []string{"TD-M8S"},
		Type: "gmrs", Drivers: []string{"td-m8s"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "22 canali FRS (Stati Uniti). Il programma del costruttore è lo stesso della Baofeng T22 in versione FRS. " +
			"Non è la TD-M8 senza «S», che ha un'altra scheda. " + ana432},
	{ID: "tid-td-v2", Brand: "TIDRADIO", Name: "TD-V2", Aliases: []string{"V2", "TDV2"}, Models: []string{"TD-V2"},
		Type: "gmrs", Drivers: []string{"td-v2"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "16 canali FRS (Stati Uniti). Le istruzioni del programma del costruttore lo indicano anche per Pofung F8, F8i, F8U, F8X e BF-V8. " + ana432},
	{ID: "tid-td-h4", Brand: "TIDRADIO", Name: "TD-H4", Aliases: []string{"H4", "TDH4"}, Models: []string{"TD-H4"},
		Type: "analog", Drivers: []string{"td-h4", "td-h4ham", "td-h4gmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali e impostazioni generali. Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. Non è in CHIRP." + cps425},
	{ID: "tid-td-h11", Brand: "TIDRADIO", Name: "TD-H11", Aliases: []string{"H11", "TDH11"}, Models: []string{"TD-H11"},
		Type: "analog", Drivers: []string{"td-h11", "td-h11ham", "td-h11gmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali e impostazioni generali. Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. Non è in CHIRP." + cps425},
}
