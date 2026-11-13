// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.26: Retevis RB39 / RB39P e RB639 / RB639P, con il driver nuovo
// ricavato dal codice dei programmi del costruttore (drv_auctus.go).

var catalog426 = []catItem{
	{ID: "rtv-rb39", Brand: "Retevis", Name: "RB39 / RB39P", Aliases: []string{"RB-39", "RB-39P"}, Models: []string{"RB39", "RB39P"}, Type: "gmrs",
		Drivers: []string{"rt-rb39"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "Versione FRS (22 canali). Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rtv-rb639", Brand: "Retevis", Name: "RB639 / RB639P", Aliases: []string{"RB-639", "RB-639P"}, Models: []string{"RB639", "RB639P"}, Type: "pmr",
		Drivers: []string{"rt-rb639"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "Versione PMR446 (16 canali). Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
}
