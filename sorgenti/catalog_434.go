// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.34: Yaesu VX-170 / VX-177 e le veicolari FT-7800 / FT-7900 e
// FT-8900 (drv_ft7800.go). Solo canali.

var catalog434 = []catItem{
	{ID: "ys-vx-170", Brand: "Yaesu", Name: "VX-170", Type: "analog", Drivers: []string{"ys-vx170"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("VX170Radio") + " Palmare VHF: 200 canali con nomi di 6 caratteri."},
	{ID: "ys-vx-177", Brand: "Yaesu", Name: "VX-177", Type: "analog", Drivers: []string{"ys-vx177"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("VX177Radio") + " Palmare UHF: 200 canali con nomi di 6 caratteri."},
	{ID: "ys-ft-7800", Brand: "Yaesu", Name: "FT-7800", Aliases: []string{"FT-7900", "FT-7800R", "FT-7900R"}, Type: "analog", Drivers: []string{"ys-ft7800"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("FT7800Radio") + " Veicolare bibanda: 999 canali con nomi di 6 caratteri."},
	{ID: "ys-ft-8900", Brand: "Yaesu", Name: "FT-8900", Aliases: []string{"FT-8900R"}, Type: "analog", Drivers: []string{"ys-ft8900"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNoteCh("FT8900Radio") + " Veicolare a quattro bande: 799 canali con nomi di 6 caratteri."},
}
