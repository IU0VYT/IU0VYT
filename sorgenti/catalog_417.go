// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.17: palmari Kenwood programmati un canale per volta (TH-F6A, TH-F7E, TH-G71,
// TH-D7, TH-K2), palmari professionali Kenwood TK delle serie 60 e 60G, Radtel
// RT-860 (stesso collegamento e stessa memoria dei canali della RT-880G).

const (
	kwLiveCatNote = "Canali letti e scritti un canale per volta con i comandi del PC, come fanno i driver di CHIRP per queste radio: la radio " +
		"resta accesa normalmente (9600 baud) e si scrivono solo i canali cambiati. Provato con una radio simulata, non su una radio vera."
	kwTkCatNote = " Radio professionale: banda e numero dei canali dipendono dalla versione, letta dalla memoria della radio."
)

var catalog417 = []catItem{
	{ID: "kw-th-f6", Brand: "Kenwood", Name: "TH-F6A / TH-F7E", Aliases: []string{"TH-F6", "TH-F7", "TH-F6A", "TH-F7E"}, Type: "analog",
		Drivers: []string{"kw-thf7", "kw-thf6"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: kwLiveCatNote},
	{ID: "kw-th-g71", Brand: "Kenwood", Name: "TH-G71A / TH-G71E", Aliases: []string{"TH-G71", "TH-G71A", "TH-G71E"}, Type: "analog",
		Drivers: []string{"kw-thg71"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: kwLiveCatNote},
	{ID: "kw-th-d7", Brand: "Kenwood", Name: "TH-D7A / TH-D7E", Aliases: []string{"TH-D7", "TH-D7A", "TH-D7E", "TH-D7A(G)", "TH-D7E(G)", "TH-D7G"}, Type: "analog",
		Drivers: []string{"kw-thd7", "kw-thd7g"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: kwLiveCatNote},
	{ID: "kw-th-k2", Brand: "Kenwood", Name: "TH-K2AT / TH-K2E / TH-K2ET", Aliases: []string{"TH-K2", "TH-K2AT", "TH-K2E", "TH-K2ET"}, Type: "analog",
		Drivers: []string{"kw-thk2"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: kwLiveCatNote},
	{ID: "kw-tk-270", Brand: "Kenwood", Name: "TK-260 / TK-270 / TK-272 / TK-278", Aliases: []string{"TK-260", "TK-270", "TK-272", "TK-278"}, Type: "analog",
		Drivers: []string{"kw-tk270", "kw-tk260", "kw-tk272", "kw-tk278"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK270_Radio") + kwTkCatNote},
	{ID: "kw-tk-370", Brand: "Kenwood", Name: "TK-360 / TK-370 / TK-372 / TK-378", Aliases: []string{"TK-360", "TK-370", "TK-372", "TK-378"}, Type: "analog",
		Drivers: []string{"kw-tk370", "kw-tk360", "kw-tk372", "kw-tk378"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK370_Radio") + kwTkCatNote},
	{ID: "kw-tk-270g", Brand: "Kenwood", Name: "TK-260G / TK-270G / TK-272G / TK-278G", Aliases: []string{"TK-260G", "TK-270G", "TK-272G", "TK-278G"}, Type: "analog",
		Drivers: []string{"kw-tk270g", "kw-tk260g", "kw-tk272g", "kw-tk278g"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK270G_Radios") + kwTkCatNote},
	{ID: "kw-tk-370g", Brand: "Kenwood", Name: "TK-360G / TK-370G / TK-372G / TK-378G / TK-388G",
		Aliases: []string{"TK-360G", "TK-370G", "TK-372G", "TK-378G", "TK-388G"}, Type: "analog",
		Drivers: []string{"kw-tk370g", "kw-tk360g", "kw-tk372g", "kw-tk378g", "kw-tk388g"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK370G_Radios") + kwTkCatNote},
	// il programma di Radtel della RT-860 («RT-860 Radio CPS» V1.06) usa gli stessi comandi, gli stessi blocchi e lo stesso
	// canale di 48 byte del driver di CHIRP della RT-880G (senza il limite RX/TX), con gli stessi limiti di frequenza
	{ID: "rtl-rt-860", Brand: "Radtel", Name: "RT-860", Aliases: []string{"RT860"}, Type: "analog", Drivers: []string{"rtl-rt880g"},
		Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "La RT-860 si legge e si scrive con la scheda «Radtel RT-880G»: il programma di Radtel usa lo stesso collegamento e la stessa " +
			"memoria dei canali della RT-880G (confronto con il programma del costruttore, non provato su una radio vera)."},
}
