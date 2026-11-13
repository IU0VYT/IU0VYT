// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.33: palmari Icom degli anni '90 e 2000 gestiti da CHIRP che
// mancavano (drv_ict7.go, drv_icq7.go, drv_icp7.go). Solo canali.

var catalog433 = []catItem{
	{ID: "ic-t7h", Brand: "Icom", Name: "IC-T7H", Type: "analog", Drivers: []string{"ic-t7h"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICT7HRadio") + " 60 canali VHF/UHF senza nome."},
	{ID: "ic-t8a", Brand: "Icom", Name: "IC-T8A", Aliases: []string{"IC-T8"}, Type: "analog", Drivers: []string{"ic-t8a"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICT8ARadio") + " Tribanda 50/144/430 MHz: 100 canali con nomi di 4 caratteri."},
	{ID: "ic-q7a", Brand: "Icom", Name: "IC-Q7A", Aliases: []string{"IC-Q7"}, Type: "analog", Drivers: []string{"ic-q7a"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICQ7Radio") + " Bibanda 144/430 MHz con ricevitore a larga banda: 200 canali senza nome."},
	{ID: "ic-p7", Brand: "Icom", Name: "IC-P7", Type: "analog", Drivers: []string{"ic-p7"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNoteCh("ICP7Radio") + " Bibanda 144/430 MHz con ricevitore a larga banda: 1000 canali con nomi di 6 caratteri."},
}
