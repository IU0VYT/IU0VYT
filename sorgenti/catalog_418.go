// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.18: palmari professionali Kenwood con i canali divisi in zone o gruppi
// (TK-2140, TK-3140, TK-280, TK-380), Radtel RT-330 e RT-400 (dai programmi del
// costruttore), firmware R3b0rn di fagci per la Quansheng UV-K5.

const (
	kwZoneCatNote = " Radio professionale con i canali divisi in zone (gruppi): la lista mostra le posizioni della memoria, un canale " +
		"nuovo entra nella zona del canale che lo precede. Le zone non si creano e non si cancellano da qui."
	k5FagciNote = " Dalla 4.18 anche il firmware R3b0rn di fagci (solo i canali: 298 con la memoria originale, fino a 1024 con una memoria più grande)."
	fagciSite   = "https://github.com/fagci/uvk5-fagci-reborn"
	rtlCpsNote  = " Ricavato dal programma del costruttore e provato con una radio simulata, non su una radio vera. Solo i canali: le impostazioni " +
		"lette dalla radio si riscrivono come sono."
)

var catalog418 = []catItem{
	{ID: "kw-tk-2140", Brand: "Kenwood", Name: "TK-2140 / TK-3140", Aliases: []string{"TK-2140", "TK-3140", "TK-2140K", "TK-3140K", "TK-3140K2", "TK-3140K3"},
		Type: "analog", Drivers: []string{"kw-tk2140k", "kw-tk3140k", "kw-tk3140k2", "kw-tk3140k3"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("KenwoodTK2140KRadio") + kwZoneCatNote},
	{ID: "kw-tk-280", Brand: "Kenwood", Name: "TK-280 / TK-380", Aliases: []string{"TK-280", "TK-380"},
		Type: "analog", Drivers: []string{"kw-tk280", "kw-tk380"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK280_Radios") + kwZoneCatNote},
	{ID: "rtl-rt-330", Brand: "Radtel", Name: "RT-330", Aliases: []string{"RT330"}, Type: "analog", Drivers: []string{"rt-rt330"},
		Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "512 canali con nomi di 12 caratteri, 100-520 MHz, 115200 baud. Stesso protocollo e stesso record dei canali delle Radtel RT-470 " +
			"(famiglia MML JC-8810 di CHIRP), con un saluto proprio." + rtlCpsNote},
	{ID: "rtl-rt-400", Brand: "Radtel", Name: "RT-400", Aliases: []string{"RT400", "JJCC-7700"}, Type: "analog", Drivers: []string{"rt-rt400"},
		Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "256 canali con nomi di 12 caratteri, 18-1000 MHz. Stesso saluto, stesso protocollo e stesso record dei canali delle Radtel RT-470 " +
			"(famiglia MML JC-8810 di CHIRP). Il programma del costruttore non controlla l'identificativo della radio: la scheda rifiuta " +
			"le radio che si identificano come un altro modello della famiglia." + rtlCpsNote},
}
