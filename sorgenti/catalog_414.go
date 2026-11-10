// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.14: palmari Yaesu e Icom gestiti con i driver di CHIRP riscritti in Go.

const (
	yaesuWeb = "https://www.yaesu.com/product-listing.aspx?Category=VHF/UHF%20Handhelds"
	icomWeb  = "https://www.icomjapan.com/support/"
)

const ft4Note = " 200 canali, cavo SCU-35 nella presa del microfono. A differenza di CHIRP, il simplex scritto dalla radio resta simplex " +
	"e uno scostamento che non è multiplo del passo della radio diventa una frequenza di trasmissione separata."

const ft4Sub3Note = " Le radio che dichiarano il sottotipo 3 hanno una scheda a parte: il programma indica quella giusta."

var catalog414 = []catItem{
	{ID: "ys-ft-4xr", Brand: "Yaesu", Name: "FT-4XR", Aliases: []string{"FT-4X"}, Type: "analog", Drivers: []string{"ys-ft4xr"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT4XRRadio") + ft4Note + " Versione per gli Stati Uniti."},
	{ID: "ys-ft-4xe", Brand: "Yaesu", Name: "FT-4XE", Type: "analog", Drivers: []string{"ys-ft4xe"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT4XERadio") + ft4Note + " Versione europea."},
	{ID: "ys-ft-4vr", Brand: "Yaesu", Name: "FT-4VR", Aliases: []string{"FT-4V"}, Type: "analog", Drivers: []string{"ys-ft4vr"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT4VRRadio") + ft4Note + " Solo VHF."},
	{ID: "ys-ft-65r", Brand: "Yaesu", Name: "FT-65R", Type: "analog", Drivers: []string{"ys-ft65r", "ys-ft65r-s3"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT65RRadio") + ft4Note + " Versione per Stati Uniti e Asia." + ft4Sub3Note},
	{ID: "ys-ft-65e", Brand: "Yaesu", Name: "FT-65E", Type: "analog", Drivers: []string{"ys-ft65e", "ys-ft65e-s3"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT65ERadio") + ft4Note + " Versione europea." + ft4Sub3Note},
	{ID: "ys-ft-25r", Brand: "Yaesu", Name: "FT-25R", Aliases: []string{"FT-25E"}, Type: "analog", Drivers: []string{"ys-ft25r", "ys-ft25r-s3"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("YaesuFT25RRadio") + ft4Note + " Solo VHF." + ft4Sub3Note},
	{ID: "ys-ft-60", Brand: "Yaesu", Name: "FT-60R", Aliases: []string{"FT-60E", "FT-60"}, Type: "analog", Drivers: []string{"ys-ft60"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT60Radio") + " 1000 canali; la copia parte dalla radio (modalità «F8 CLONE»). " +
			"A differenza di CHIRP, le frequenze che la radio non può memorizzare vengono rifiutate invece di essere arrotondate."},
	// Icom
	{ID: "ic-v86", Brand: "Icom", Name: "IC-V86", Type: "analog", Drivers: []string{"ic-v86"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICV86Radio") + " 200 canali VHF con nomi di 5 caratteri."},
	{ID: "ic-v80", Brand: "Icom", Name: "IC-V80", Aliases: []string{"IC-V80E"}, Type: "analog", Drivers: []string{"ic-v80"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICV80Radio") + " 200 canali VHF con nomi di 5 caratteri, copia veloce."},
	{ID: "ic-t70", Brand: "Icom", Name: "IC-T70", Aliases: []string{"IC-T70A", "IC-T70E"}, Type: "analog", Drivers: []string{"ic-t70"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICT70Radio") + " 300 canali VHF/UHF con nomi di 6 caratteri."},
	{ID: "ic-t10", Brand: "Icom", Name: "IC-T10", Type: "analog", Drivers: []string{"ic-t10"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICVT10Radio") + " 200 canali VHF/UHF con nomi di 6 caratteri."},
	{ID: "ic-v82", Brand: "Icom", Name: "IC-V82", Type: "analog", Drivers: []string{"ic-v82"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICV82Radio") + " 200 canali VHF con nomi di 5 caratteri."},
	{ID: "ic-u82", Brand: "Icom", Name: "IC-U82", Type: "analog", Drivers: []string{"ic-u82"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICU82Radio") + " 200 canali UHF con nomi di 5 caratteri."},
}
