// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.39: Icom D-STAR ID-80H (palmare) e ID-880H (veicolare), drv_id880.go;
// palmari bibanda IC-W32A e IC-W32E, drv_icw32.go.

var catalog439 = []catItem{
	{ID: "ic-id-80", Brand: "Icom", Name: "ID-80H", Aliases: []string{"ID-80", "ID-E80"}, Type: "analog", Drivers: []string{"ic-id80"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ID80Radio") + id880CatNote},
	{ID: "ic-id-880", Brand: "Icom", Name: "ID-880H", Aliases: []string{"ID-880", "ID-E880"}, Type: "analog", Drivers: []string{"ic-id880"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ID880Radio") + id880CatNote},
	{ID: "ic-w32a", Brand: "Icom", Name: "IC-W32A", Aliases: []string{"IC-W32", "W32A"}, Type: "analog", Drivers: []string{"ic-w32a"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICW32ARadio") + icw32CatNote},
	{ID: "ic-w32e", Brand: "Icom", Name: "IC-W32E", Aliases: []string{"W32E"}, Type: "analog", Drivers: []string{"ic-w32e"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICW32ERadio") + icw32CatNote},
}

const icw32CatNote = " Palmare bibanda con due metà separate: 100 canali VHF e 100 canali UHF con nomi di 8 caratteri, in una lista sola (da 0 a 99 VHF, da 100 a 199 UHF)."

const id880CatNote = " 1000 canali con nomi di 8 caratteri, analogici o digitali D-STAR (colonna «Modo»). La radio dichiara il suo modello: se è l'altro il programma lo dice."
