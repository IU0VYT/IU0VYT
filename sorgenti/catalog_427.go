// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.27: Retevis C6 e C66 e Radioddity Echo-DE / Echo-IT, con il driver
// nuovo ricavato dal codice dei programmi dei costruttori (drv_wlt.go).

var catalog427 = []catItem{
	{ID: "rtv-c6", Brand: "Retevis", Name: "C6", Aliases: []string{"MateTalk C6", "C-6"}, Models: []string{"C6"}, Type: "gmrs",
		Drivers: []string{"rt-c6"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "FRS / GMRS, 30 canali (22 più 8 per i ponti). Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rtv-c66", Brand: "Retevis", Name: "C66", Aliases: []string{"MateTalk C66", "C-66"}, Models: []string{"C66"}, Type: "analog",
		Drivers: []string{"rt-c66"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "16 canali UHF (400-480 MHz). Canali e impostazioni generali come nel programma del costruttore. Non è nel CHIRP." + cps425},
	{ID: "rdd-echo", Brand: "Radioddity", Name: "Echo-DE / Echo-IT", Aliases: []string{"Echo", "ECHO-DE", "ECHO-IT", "Echo DE", "Echo IT"}, Models: []string{"Echo-DE", "Echo-IT"}, Type: "pmr",
		Drivers: []string{"rd-echo"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "PMR446, 16 canali, guida vocale in tedesco o in italiano. Canali e impostazioni generali come nel programma del costruttore «Radioddity ECHO». Non è nel CHIRP." + cps425},
}
