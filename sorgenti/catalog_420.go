// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.20: Radtel RT-4D, RT-6D (DMR) e RT-1000 Pro (dai programmi del costruttore).

const rtl4dNote = " Ricavato dal programma del costruttore e provato con una radio simulata, non su una radio vera. Si scrivono solo " +
	"canali e zone; impostazioni e VFO letti dalla radio si riscrivono come sono. «Identifica» non la distingue dalle altre " +
	"Radtel e Bajeton con lo stesso collegamento: scegli tu la scheda."

var catalog420 = []catItem{
	{ID: "rtl-rt-4d", Brand: "Radtel", Name: "RT-4D", Aliases: []string{"RT4D"}, Type: "dmr",
		Drivers: []string{"rtl-rt4d"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "Radio DMR: 1024 canali con nomi di 16 caratteri, 250 zone, 18-660 e 840-1000 MHz, 115200 baud. Contatti, gruppi e " +
			"messaggi non si toccano." + rtl4dNote},
	{ID: "rtl-rt-6d", Brand: "Radtel", Name: "RT-6D", Aliases: []string{"RT6D"}, Type: "dmr",
		Drivers: []string{"rtl-rt6d"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "Radio DMR: 1024 canali con nomi di 16 caratteri, 250 zone, 18-660 e 840-1000 MHz, 115200 baud. Stessa memoria della " +
			"RT-4D. Contatti, gruppi e messaggi non si toccano." + rtl4dNote},
	{ID: "rtl-rt-1000-pro", Brand: "Radtel", Name: "RT-1000 Pro", Aliases: []string{"RT-1000", "RT1000", "RT1000 Pro"}, Type: "analog",
		Drivers: []string{"rtl-rt1000pro"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "1024 canali con nomi di 16 caratteri, 250 zone, 18-660 e 840-1000 MHz, 115200 baud. Stesso canale della RT-4D, solo " +
			"analogico; l'esclusione dalla scansione non si gestisce (resta com'è sulla radio)." + rtl4dNote},
}
