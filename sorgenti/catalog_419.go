// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.19: Radtel RT-817A / RT-871A (dal programma del costruttore).

var catalog419 = []catItem{
	{ID: "rtl-rt-817a", Brand: "Radtel", Name: "RT-817A / RT-871A", Aliases: []string{"RT-817A", "RT-871A", "RT817A", "RT871A"}, Type: "analog",
		Drivers: []string{"rtl-rt871a"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "1024 canali con nomi di 16 caratteri, 18-660 e 840-1000 MHz, 115200 baud. Stesso collegamento e stesso canale della " +
			"Radtel RT-880G, senza le zone." + rtlCpsNote + " «Identifica» non la distingue dalla RT-880G: scegli tu la scheda."},
}
