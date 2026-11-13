// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.24: radio che i programmi dei costruttori (scaricati dalle loro
// pagine e confrontati istruzione per istruzione con quelli delle radio già
// gestite) mostrano uguali a un modello già nel programma, e i modelli
// equivalenti di CHIRP che mancavano dal catalogo. Nessun driver nuovo: ogni
// scheda usa un driver già verificato con i test automatici.

// cps424: fine comune delle note delle radio riconosciute dal confronto dei programmi
const cps424 = " Il confronto è stato fatto sul codice dei programmi dei costruttori: non è ancora provato con una radio vera. " +
	"Leggi prima la radio: se il modello non è quello giusto la radio non risponde e non viene modificata."

// alias424: nota dei modelli equivalenti di CHIRP
const alias424 = "Nella tabella dei modelli equivalenti di CHIRP questa radio corrisponde a un modello già gestito dal programma: " +
	"si usa lo stesso driver, verificato con i test automatici."

var catalog424 = []catItem{
	// --- dal confronto dei programmi dei costruttori ---
	{ID: "bjt-bt-8000", Brand: "Bajeton", Name: "BT-8000", Aliases: []string{"BT8000", "BJ-8000"}, Type: "analog",
		Drivers: []string{"rd-gs10b"}, Links: []catLink{{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Il programma del costruttore («BT-8000 CPS» v1.10.4) usa lo stesso comando di aggancio e legge e scrive le stesse aree di memoria " +
			"del programma della Radioddity GS-10B: si usa il driver della GS-10B (256 canali)." + cps424},
	{ID: "rdd-gc-5", Brand: "Radioddity", Name: "GC-5", Aliases: []string{"GC5"}, Type: "analog",
		Drivers: []string{"bf-uv17r"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Il programma del costruttore («Radioddity GC-5 CPS») è lo stesso della serie Baofeng UV-17R (cambiano solo le scritte): " +
			"si usa il driver della UV-17R." + cps424},
	{ID: "rdd-trivox", Brand: "Radioddity", Name: "Trivox", Type: "analog",
		Drivers: []string{"bj-9000", "bj-9000-960"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Il programma del costruttore («Radioddity Trivox CPS» v1.3.3) è lo stesso della Radtel RT-950 Pro e della Bajeton BJ-9000, " +
			"con la stessa identificazione della radio: si usano gli stessi driver (990 canali; 960 con i firmware più vecchi)." + cps424},
	{ID: "rdd-pr-t5", Brand: "Radioddity", Name: "PR-T5", Aliases: []string{"PRT5"}, Type: "pmr",
		Drivers: []string{"bf-gt18pmr"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Il programma del costruttore («PR-T5») usa lo stesso protocollo e la stessa memoria di quello della Baofeng GT-18 PMR446 " +
			"(99 canali): si usa il suo driver. Se è impostata la password di lettura e scrittura, toglila prima con il programma del costruttore." + cps424},
	{ID: "rdd-gt-82", Brand: "Radioddity", Name: "GT-82", Aliases: []string{"GT82"}, Type: "analog",
		Drivers: []string{"bf-uv82"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Il programma del costruttore («CPS_UV82») si presenta alla radio con la stessa chiave della Baofeng UV-82: " +
			"si usa il driver della UV-82." + cps424},
	{ID: "rtv-rt28", Brand: "Retevis", Name: "RT28", Aliases: []string{"RT-28"}, Type: "pmr",
		Drivers: []string{"bf-888s"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "Il programma del costruttore («RT28») è lo stesso della Baofeng BF-R5 («BF_480»: cambiano solo i nomi) e usa il protocollo " +
			"della BF-888S: si usa il driver della BF-888S (16 canali, da 400 a 490 MHz; la versione VHF non è gestita)." + cps424},

	// --- modelli equivalenti di CHIRP che mancavano ---
	{ID: "rdd-ga-5s", Brand: "Radioddity", Name: "GA-5S", Aliases: []string{"GA5S"}, Type: "analog",
		Drivers: []string{"bf-f8hp"}, Links: []catLink{info("Download center Radioddity", rddDown)}, Notes: alias424},
	{ID: "rtv-rt5-8w", Brand: "Retevis", Name: "RT5 (tre potenze)", Aliases: []string{"RT5 tri-power", "RT5 8W"}, Type: "analog",
		Drivers: []string{"bf-f8hp"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: alias424},
	{ID: "rtv-rt5r", Brand: "Retevis", Name: "RT5R / RT5RV / RT5", Aliases: []string{"RT-5R", "RT-5RV", "RT-5"}, Models: []string{"RT5R", "RT5RV", "RT5"}, Type: "analog",
		Drivers: []string{"bf-uv5r"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: alias424},
	{ID: "tid-td-uv5r-tripower", Brand: "TIDRADIO", Name: "TD-UV5R TriPower", Aliases: []string{"TD-UV5R", "UV5R TriPower"}, Type: "analog",
		Drivers: []string{"bf-f8hp"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)}, Notes: alias424},
	{ID: "tcs-ti-f8", Brand: "TechSide", Name: "TI-F8+ / TS-T9+", Aliases: []string{"TI-F8 Plus", "TS-T9 Plus"}, Models: []string{"TI-F8+", "TS-T9+"}, Type: "analog",
		Drivers: []string{"bf-f8hp"}, Notes: alias424},
	{ID: "rgd-rh5r", Brand: "Rugged", Name: "RH5R", Type: "analog", Drivers: []string{"bf-uv5r"}, Notes: alias424},
	{ID: "rgd-rh5x", Brand: "Rugged", Name: "RH5X", Type: "analog", Drivers: []string{"bf-a58"}, Notes: alias424},
	{ID: "ank-a-5r", Brand: "Ansoko", Name: "A-5R", Aliases: []string{"A5R"}, Type: "analog", Drivers: []string{"bf-uv5r"}, Notes: alias424},
	{ID: "ank-a-8s", Brand: "Ansoko", Name: "A-8S", Aliases: []string{"A8S"}, Type: "pmr", Drivers: []string{"bf-888s"}, Notes: alias424},
	{ID: "arc-ar-5", Brand: "Arcshell", Name: "AR-5 / AR-6", Aliases: []string{"AR5", "AR6"}, Models: []string{"AR-5", "AR-6"}, Type: "pmr",
		Drivers: []string{"bf-888s"}, Notes: alias424},
	{ID: "arc-ar-8", Brand: "Arcshell", Name: "AR-8", Aliases: []string{"AR8"}, Type: "pmr", Drivers: []string{"bf-t8"}, Notes: alias424},
	{ID: "grv-gv-8s", Brand: "Greaval", Name: "GV-8S / GV-9S", Aliases: []string{"GV8S", "GV9S"}, Models: []string{"GV-8S", "GV-9S"}, Type: "pmr",
		Drivers: []string{"bf-888s"}, Notes: alias424},
	{ID: "tnw-tw-325", Brand: "Tenway", Name: "TW-325", Aliases: []string{"TW325"}, Type: "pmr", Drivers: []string{"bf-888s"}, Notes: alias424},
	{ID: "tnw-uv-82-pro", Brand: "Tenway", Name: "UV-82 Pro", Aliases: []string{"UV-82Pro"}, Type: "analog", Drivers: []string{"bf-uv82hp"}, Notes: alias424},
	{ID: "tnw-uv-5r-pro", Brand: "Tenway", Name: "UV-5R Pro", Aliases: []string{"UV-5RPro"}, Type: "analog", Drivers: []string{"bf-f8hp"}, Notes: alias424},
}
