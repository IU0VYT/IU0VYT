// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.28: Radioddity UB15-G, DB-35G, DB35-UV e DB40-G, con il driver
// nuovo ricavato dal programma dei costruttori e dal suo file di configurazione
// (drv_wf.go); la DB40-G è anche nel CHIRP e viene confrontata con quello.

var catalog428 = []catItem{
	{ID: "rdd-ub15g", Brand: "Radioddity", Name: "UB15-G", Aliases: []string{"UB-15G", "UB15G", "UB 15G"}, Models: []string{"UB15-G"}, Type: "gmrs",
		Drivers: []string{"rd-ub15g", "rd-ub15g-110"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Veicolare GMRS da 15 W, 80 canali più le 8 posizioni dei ponti. Due schede, secondo il firmware della radio: prima della V1.10 (senza nomi) e V1.10 (con i nomi). " +
			"Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rdd-db35g", Brand: "Radioddity", Name: "DB35-G", Aliases: []string{"DB-35G", "DB35G", "DB 35G"}, Models: []string{"DB35-G"}, Type: "gmrs",
		Drivers: []string{"rd-db35g"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Veicolare GMRS da 35 W, 80 canali, riceve anche in VHF (il programma del costruttore si chiama «DB-35G»). Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rdd-db35uv", Brand: "Radioddity", Name: "DB35-UV", Aliases: []string{"DB-35UV", "DB35UV", "DB 35UV"}, Models: []string{"DB35-UV"}, Type: "analog",
		Drivers: []string{"rd-db35uv"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Veicolare VHF / UHF da 35 W, 80 canali. Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rdd-db40g", Brand: "Radioddity", Name: "DB40-G", Aliases: []string{"DB-40G", "DB40G", "DB 40G"}, Models: []string{"DB40-G"}, Type: "gmrs",
		Drivers: []string{"rd-db40g"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Veicolare GMRS da 40 W, 199 canali con nome. Canali e impostazioni generali; è anche nel CHIRP (dal novembre 2026) e il file è lo stesso." + cps425},
}
