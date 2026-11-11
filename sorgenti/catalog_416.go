// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.16: firmware libero OpenGD77 (GD-77, DM-1801, RD-5R, DM-1701, MD-UV380…) e
// altri firmware liberi della Quansheng UV-K5 (NUNU di kamilsss655, UV-KX di
// joaquim.org, firmware originali modificati).

const (
	ogd77Site  = "https://www.opengd77.com/"
	qdmrSite   = "https://dm3mat.de/software/qdmr"
	ogd77Note  = " Con il firmware libero OpenGD77 la radio compare come porta seriale USB e usa un altro protocollo: apri la scheda «OpenGD77»."
	k5MoreNote = " Dalla 4.16 anche i firmware NUNU di kamilsss655 (memoria di egzumer, 115200 baud), UV-KX di joaquim.org (230 canali) e, con la scheda " +
		"«OSFW o originale modificato», i firmware originali ritoccati con le bande ampliate (uvmod-kitchen, raccolta di Tunas1337)."
	ogd77CatNote = "Con il firmware libero OpenGD77 la radio accesa compare come porta seriale USB: 1024 canali analogici o DMR, letti e " +
		"scritti come nel programma libero qdmr. Si scrivono solo i canali e le zone; contatti, gruppi, impostazioni e taratura non si toccano."
)

var catalog416 = []catItem{
	{ID: "rdd-gd-77", Brand: "Radioddity", Name: "GD-77 / GD-77S (firmware OpenGD77)", Aliases: []string{"GD77", "OpenGD77", "TYT MD-760"}, Type: "dmr",
		Drivers: []string{"ogd77"},
		Links:   []catLink{info("Sito di OpenGD77", ogd77Site), info("qdmr (programma libero per radio DMR, GPL)", qdmrSite)},
		Notes:   ogd77CatNote + " Con il firmware originale di Radioddity la GD-77 non è gestita."},
	{ID: "tyt-md-uv380-ogd", Brand: "TYT", Name: "MD-UV380 / MD-UV390 / MD-380 / MD-2017 / MD-9600 (firmware OpenGD77)",
		Aliases: []string{"OpenUV380", "OpenMD9600", "Retevis RT3S", "Retevis RT3", "Retevis RT82", "Retevis RT90"}, Type: "dmr",
		Drivers: []string{"ogd77-uv380"},
		Links:   []catLink{info("Sito di OpenGD77", ogd77Site), info("qdmr (programma libero per radio DMR, GPL)", qdmrSite)},
		Notes:   ogd77CatNote + " Con il firmware originale di TYT queste radio non sono gestite."},
}
