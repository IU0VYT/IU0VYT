// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.21: Radtel RT-5D (dal programma del costruttore), Kenwood TK-2180 / TK-3180
// e TK-7180 / TK-8180 (tk8180.py di CHIRP).

const kw180CatNote = " Radio professionale con i canali divisi in zone: la lista ha 250 posizioni per ognuna delle prime 32 zone e la " +
	"posizione è il numero del canale nella zona. Si usano solo le zone che la radio ha già; in tutto 512 canali."

var catalog421 = []catItem{
	{ID: "rtl-rt-5d", Brand: "Radtel", Name: "RT-5D", Aliases: []string{"RT5D", "JJCC-888DMR"}, Type: "dmr",
		Drivers: []string{"rtl-rt5d"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "Radio DMR: 16 zone da 64 canali con nomi di 12 caratteri, 18-1000 MHz, 115200 baud. Ricavato dal programma del costruttore " +
			"e provato con una radio simulata, non su una radio vera. Lettura e scrittura riguardano tutta la memoria, come nel programma " +
			"del costruttore: contatti, gruppi e impostazioni non si modificano qui, ma scrivendo un file vecchio tornano quelli di allora."},
	{ID: "kw-tk-2180", Brand: "Kenwood", Name: "TK-2180 / TK-3180", Aliases: []string{"TK-2180", "TK-3180", "TK-3180K", "TK-3180K2", "TK-3180E"},
		Type: "analog", Drivers: []string{"kw-tk2180", "kw-tk3180k", "kw-tk3180k2", "kw-tk3180e"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("KenwoodTK2180Radio") + kw180CatNote},
	{ID: "kw-tk-7180", Brand: "Kenwood", Name: "TK-7180 / TK-8180", Aliases: []string{"TK-7180", "TK-8180", "TK-7180E", "TK-8180E"},
		Type: "analog", Drivers: []string{"kw-tk7180", "kw-tk8180", "kw-tk7180e", "kw-tk8180e"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("KenwoodTK7180Radio") + " Radio veicolare." + kw180CatNote},
}
