// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.41: i palmari di CHIRP che mancavano. Yaesu VX-1R (drv_vx1.go),
// Icom IC-91AD e IC-92AD (drv_ic9x.go), BTECH UV-PRO (drv_uvpro.go) e il
// ricevitore AOR AR8200 (drv_ar8200.go; AOR è una marca nuova).

const aorWeb = "https://www.aorja.com/"

// radio che CHIRP programma un canale per volta, senza file: il confronto è
// fatto sul cavo, con il driver di CHIRP collegato alla radio simulata
const liveCatNote = "Canali con i comandi del driver di CHIRP, che programma questa radio un canale per volta (senza file .img): verificato con i test automatici " +
	"collegando il driver di CHIRP alla radio simulata (stessi canali letti, stessi comandi di scrittura). I file sono nel formato di IU0VYT Flash. "

var catalog441 = []catItem{
	{ID: "ys-vx-1", Brand: "Yaesu", Name: "VX-1R", Aliases: []string{"VX-1", "VX1R", "VX1"}, Type: "analog", Drivers: []string{"ys-vx1"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)},
		Notes: chirpNote("VX1Radio") + " Palmare bibanda 144/430 MHz con ricevitore a larga banda: 260 posizioni (due gruppi di memorie, limiti di scansione, " +
			"canali «home» e onde medie), nomi di 6 caratteri."},
	{ID: "ic-91ad", Brand: "Icom", Name: "IC-91AD", Aliases: []string{"IC-91A", "IC-E91", "IC-91", "IC91AD"}, Type: "analog", Drivers: []string{"ic-9x-a", "ic-9x-b"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: liveCatNote + ic9xCatNote},
	{ID: "ic-92ad", Brand: "Icom", Name: "IC-92AD", Aliases: []string{"IC-E92D", "IC-92", "IC92AD"}, Type: "analog", Drivers: []string{"ic-9x-a", "ic-9x-b"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: liveCatNote + ic9xCatNote},
	{ID: "btech-uv-pro", Brand: "BTECH", Name: "UV-PRO", Aliases: []string{"UV-Pro", "UVPRO"}, Type: "analog", Drivers: []string{"bt-uvpro"},
		Links: []catLink{info("Sito BTECH", btchWeb)},
		Notes: liveCatNote + "Via Bluetooth (porta seriale SPP), senza cavo: i 30 canali del gruppo attivo sulla radio, nomi di 10 caratteri. " +
			"Scrittura ammessa solo con la versione del firmware provata dall'autore del driver di CHIRP (147)."},
	{ID: "aor-ar8200", Brand: "AOR", Name: "AR8200", Aliases: []string{"AR-8200", "AR 8200"}, Type: "analog", Drivers: []string{"aor-ar8200"},
		Links: []catLink{info("Sito AOR", aorWeb)},
		Notes: liveCatNote + "Ricevitore palmare a larga banda (da 0,1 a 2040 MHz): le 1000 memorie di scansione, nomi di 12 caratteri; canali di sola ricezione."},
}

const ic9xCatNote = "Palmare bibanda con ricevitore a larga banda e D-STAR: banda A 800 memorie, banda B 400 memorie (anche D-STAR), più limiti di scansione " +
	"e canali di chiamata; nomi di 8 caratteri. Le due bande si scelgono come due radio distinte."
