// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.25: radio con un driver nuovo, ricavato leggendo il codice dei
// programmi dei costruttori (drv_bt11v126.go, drv_bfd.go).

// cps425: fine comune delle note dei driver ricavati dai programmi dei costruttori
const cps425 = " Driver ricavato dal codice del programma del costruttore: non è ancora provato con una radio vera. " +
	"Leggi prima la radio e conserva la copia della lettura."

var catalog425 = []catItem{
	{ID: "rdd-gs-10b-pro", Brand: "Radioddity", Name: "GS-10B Pro", Aliases: []string{"GS10B Pro", "GS-10BPro"}, Type: "analog",
		Drivers: []string{"rd-gs10bpro"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "999 canali in un unico elenco e impostazioni generali, come nel programma del costruttore «GS-10B Pro_EnCPS» v1.0.3 " +
			"(stessa famiglia della Radtel RT-900 e della Bajeton BJ-8300). Non è nel CHIRP." + cps425},
	{ID: "rdd-fs-t3", Brand: "Radioddity", Name: "FS-T3 / PR-T3", Aliases: []string{"FST3", "PRT3", "T3"}, Models: []string{"FS-T3", "PR-T3"}, Type: "pmr",
		Drivers: []string{"rd-t3"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Canali come nel programma del costruttore «Radioddity T3» v1.3 (stesso protocollo della Baofeng BF-T17, senza password): " +
			"la FS-T3 è la versione FRS (22 canali), la PR-T3 quella PMR446 (16 canali). Non è nel CHIRP." + cps425},
}
