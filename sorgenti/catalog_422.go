// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.22: Radtel RT-480, RT-68, RT-830 e RT-850 (dai programmi del costruttore);
// Radtel RT-20M (stessa memoria della iRadio UV-5118plus).

const rt480CatNote = "256 canali (da 0 a 255) con nomi di 12 caratteri, 9600 baud. Ricavato dal programma del costruttore e provato con una " +
	"radio simulata, non su una radio vera. Si scrivono solo canali e nomi: impostazioni, VFO e DTMF della radio non si toccano."

const rt830CatNote = "128 canali senza nome, 9600 baud. Ricavato dal programma del costruttore e provato con una radio simulata, non su una radio vera. " +
	"Si scrivono le aree che scrive il programma del costruttore; la password della radio non si scrive mai."

var catalog422 = []catItem{
	{ID: "rtl-rt-480", Brand: "Radtel", Name: "RT-480", Aliases: []string{"RT480"}, Type: "analog",
		Drivers: []string{"rtl-rt480"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt480CatNote},
	{ID: "rtl-rt-68", Brand: "Radtel", Name: "RT-68", Aliases: []string{"RT68"}, Type: "analog",
		Drivers: []string{"rtl-rt68"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt480CatNote},
	{ID: "rtl-rt-830", Brand: "Radtel", Name: "RT-830", Aliases: []string{"RT830", "UV-83"}, Type: "analog",
		Drivers: []string{"rtl-rt830"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt830CatNote},
	{ID: "rtl-rt-850", Brand: "Radtel", Name: "RT-850", Aliases: []string{"RT850", "UV5118"}, Type: "analog",
		Drivers: []string{"rtl-rt850"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt830CatNote},
	{ID: "rtl-rt-20m", Brand: "Radtel", Name: "RT-20M", Aliases: []string{"RT20M"}, Type: "analog",
		Drivers: []string{"ry-uv58plus"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: "Il programma del costruttore («RT-20M Radio Program Software») usa lo stesso collegamento, gli stessi blocchi e lo stesso " +
			"canale di 32 byte della iRadio UV-5118plus: si apre con la scheda Ruyage UV58Plus (999 canali con nomi di 10 caratteri, " +
			"115200 baud). Confronto fatto sul programma, non su una radio vera."},
}
