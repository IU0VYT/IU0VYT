// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.37: palmari Yaesu VX-5, VX-7 (drv_vx5.go, drv_vx7.go: solo canali),
// VX-2 (drv_vx6.go), FT-50 (drv_ft50.go), VX-150 (drv_vx150.go) e VX-8R, VX-8DR,
// VX-8GE (drv_ysf.go), questi con le impostazioni.

var catalog437 = []catItem{
	{ID: "ys-vx-5", Brand: "Yaesu", Name: "VX-5R", Aliases: []string{"VX-5", "VX-5RS"}, Type: "analog", Drivers: []string{"ys-vx5"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("VX5Radio") + " Palmare a tre bande: 220 canali con nomi di 8 caratteri."},
	{ID: "ys-vx-7", Brand: "Yaesu", Name: "VX-7R", Aliases: []string{"VX-7", "VX-7RB"}, Type: "analog", Drivers: []string{"ys-vx7"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("VX7Radio") + " Palmare a quattro bande: 450 canali con nomi di 8 caratteri."},
	{ID: "ys-vx-2", Brand: "Yaesu", Name: "VX-2R", Aliases: []string{"VX-2", "VX-2E"}, Type: "analog", Drivers: []string{"ys-vx2"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX2Radio") + " Palmare bibanda: 1000 canali con nomi di 6 caratteri."},
	{ID: "ys-ft-50", Brand: "Yaesu", Name: "FT-50R", Aliases: []string{"FT-50", "FT-50RD"}, Type: "analog", Drivers: []string{"ys-ft50"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT50Radio") + " Palmare bibanda: 100 canali con nomi di 4 caratteri."},
	{ID: "ys-vx-150", Brand: "Yaesu", Name: "VX-150", Aliases: []string{"Vertex Standard VX-150"}, Type: "analog", Drivers: []string{"ys-vx150"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX150Radio") + " Palmare VHF: 199 canali con nomi di 7 caratteri."},
	{ID: "ys-vx-8r", Brand: "Yaesu", Name: "VX-8R", Aliases: []string{"VX-8", "VX-8E"}, Type: "analog", Drivers: []string{"ys-vx8r"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX8Radio") + vx8CatNote},
	{ID: "ys-vx-8dr", Brand: "Yaesu", Name: "VX-8DR", Aliases: []string{"VX-8D", "VX-8DE"}, Type: "analog", Drivers: []string{"ys-vx8dr"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX8DRadio") + vx8CatNote},
	{ID: "ys-vx-8ge", Brand: "Yaesu", Name: "VX-8GE", Aliases: []string{"VX-8G", "VX-8GR"}, Type: "analog", Drivers: []string{"ys-vx8ge"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX8GERadio") + vx8CatNote},
}

const vx8CatNote = " Palmare con APRS: 900 canali con nomi di 16 caratteri. La radio dichiara il suo modello: se è un'altra VX-8 il programma lo dice."
