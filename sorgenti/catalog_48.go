// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.8: radio portatili aggiunte dai driver di CHIRP

var catalog48 = []catItem{
	{ID: "tyt-th-uv3r", Brand: "TYT", Name: "TH-UV3R", Type: "analog", Drivers: []string{"tyt-thuv3r"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("TYTUV3RRadio") + " 128 canali con nomi di 6 caratteri, VHF/UHF."},
	{ID: "tyt-th-uv3r-25", Brand: "TYT", Name: "TH-UV3R-25", Aliases: []string{"TH-UV3R 2,5 kHz"}, Type: "analog", Drivers: []string{"tyt-thuv3r25"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("TYTUV3R25Radio") + " Versione con passo da 2,5 kHz e due potenze."},
	{ID: "kyd-nc-630a", Brand: "KYD", Name: "NC-630A", Aliases: []string{"Plant-Tours MT-700", "MT-700"}, Type: "analog", Drivers: []string{"kyd-nc630a"},
		Notes: chirpNote("NC630aRadio") + " 16 canali UHF. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, " +
			"provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "kyd-ip-620", Brand: "KYD", Name: "IP-620", Aliases: []string{"Argut A-36", "A-36"}, Type: "analog", Drivers: []string{"kyd-ip620"},
		Notes: chirpNote("IP620Radio") + " 200 canali VHF/UHF. Driver di CHIRP sperimentale; la lettura via cavo di CHIRP non funziona con Python 3: qui il protocollo " +
			"è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "tyt-th-uvf1", Brand: "TYT", Name: "TH-UVF1", Type: "analog", Drivers: []string{"tyt-thuvf1"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("TYTTHUVF1Radio") + " 128 canali con nomi di 7 caratteri, VHF/UHF."},
	{ID: "tyt-th-uvf8d", Brand: "TYT", Name: "TH-UVF8D", Aliases: []string{"TH-UVF8"}, Type: "analog", Drivers: []string{"tyt-thuvf8d"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("TYTUVF8DRadio") + " 128 canali (banco A) con nomi di 7 caratteri, VHF/UHF."},
	{ID: "ksun-m6", Brand: "KSUN", Name: "M6", Type: "analog", Drivers: []string{"ksun-m6"},
		Notes: chirpNote("KSunM6Radio") + " 80 canali UHF (400-480 MHz) senza nomi."},
	{ID: "pux-px-777", Brand: "Puxing", Name: "PX-777", Aliases: []string{"PX777"}, Type: "analog", Drivers: []string{"pux-px777"},
		Notes: chirpNote("Puxing777Radio") + " 128 canali con nomi di 6 caratteri, una gamma per radio (quella indicata dalla memoria)."},
	{ID: "pux-px-2r", Brand: "Puxing", Name: "PX-2R", Aliases: []string{"PX2R"}, Type: "analog", Drivers: []string{"pux-px2r"},
		Notes: chirpNote("Puxing2RRadio") + " 128 canali UHF con nomi di 5 caratteri. In CHIRP la scrittura via cavo non funziona con Python 3: " +
			"qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "bob-x3plus", Brand: "Boblov", Name: "X3 Plus", Aliases: []string{"X3Plus"}, Type: "pmr", Drivers: []string{"bob-x3plus"},
		Notes: chirpNote("BoblovX3Plus") + " Radio da casco, 16 canali UHF. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo " +
			"(con l'eco dei dati) è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "lch-hg-uv98", Brand: "Lanchonlh", Name: "HG-UV98", Type: "analog", Drivers: []string{"lch-hguv98"},
		Notes: chirpNote("LanchonlhHG_UV98") + " 128 canali con nomi di 8 caratteri, VHF/UHF, con APRS (le impostazioni APRS sono tra quelle di CHIRP)."},
	{ID: "rd-r2", Brand: "Radioddity", Name: "R2", Type: "pmr", Drivers: []string{"rd-r2"},
		Notes: chirpNote("RadioddityR2") + " 16 canali UHF senza nomi."},
	{ID: "rtv-rt24", Brand: "Retevis", Name: "RT24", Type: "pmr", Drivers: []string{"rt-rt24"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisRT24") + " PMR446, 16 canali senza nomi (CHIRP non blocca le frequenze: in Europa usa solo quelle PMR446)."},
	{ID: "rtv-rt24v", Brand: "Retevis", Name: "RT24V", Type: "pmr", Drivers: []string{"rt-rt24v"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisRT24V") + " Versione Freenet (Germania), 6 canali VHF senza nomi."},
	{ID: "rtv-h777s", Brand: "Retevis", Name: "H777S", Type: "pmr", Drivers: []string{"rt-h777s"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisH777S") + " 16 canali UHF senza nomi."},
	{ID: "rtv-rt1", Brand: "Retevis", Name: "RT1", Type: "analog", Drivers: []string{"rt-rt1"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RT1Radio") + " 16 canali senza nomi, 9 W, versione VHF o UHF."},
	{ID: "rtv-h777-v4", Brand: "Retevis", Name: "H777 V4", Aliases: []string{"H777 nuova versione"}, Type: "pmr", Drivers: []string{"rt-h777v4"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("H777V4") + " Nuova versione della H777 (dal 2024), con un'altra memoria: la scheda della H777 la riconosce e indica questa."},
	{ID: "rtv-rt21h", Brand: "Retevis", Name: "RT21H", Type: "pmr", Drivers: []string{"rt-rt21h"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RT21HRadio") + " 16 canali UHF senza nomi. In CHIRP la lettura via cavo di questa radio non entra in programmazione: " +
			"qui si segue la procedura di scrittura del suo driver, provata con la radio simulata (da verificare con una radio vera)."},
	{ID: "qs-tg-uv2p", Brand: "Quansheng", Name: "TG-UV2+", Aliases: []string{"TG-UV2", "TG-UV2 Plus"}, Type: "analog", Drivers: []string{"qs-tguv2p"},
		Links: []catLink{info("Download center Quansheng", qsDown)},
		Notes: chirpNote("QuanshengTGUV2P") + " 200 canali con nomi di 6 caratteri, 10 W; il blocco della tastiera non è tra le impostazioni (CHIRP lo rifiuta con doppio ascolto attivo)."},
	{ID: "fdx-fd-268a", Brand: "Feidaxin", Name: "FD-268A", Type: "analog", Drivers: []string{"fdx-268a"},
		Notes: chirpNote("FD268ARadio") + " 99 canali VHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "fdx-fd-268b", Brand: "Feidaxin", Name: "FD-268B", Type: "analog", Drivers: []string{"fdx-268b"},
		Notes: chirpNote("FD268BRadio") + " 99 canali UHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "fdx-fd-288a", Brand: "Feidaxin", Name: "FD-288A", Type: "analog", Drivers: []string{"fdx-288a"},
		Notes: chirpNote("FD288ARadio") + " 99 canali VHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
	{ID: "fdx-fd-288b", Brand: "Feidaxin", Name: "FD-288B", Type: "analog", Drivers: []string{"fdx-288b"},
		Notes: chirpNote("FD288BRadio") + " 99 canali UHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera)."},
	{ID: "fdx-fd-150a", Brand: "Feidaxin", Name: "FD-150A", Type: "analog", Drivers: []string{"fdx-150a"},
		Notes: chirpNote("FD150ARadio") + " 99 canali VHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
	{ID: "fdx-fd-160a", Brand: "Feidaxin", Name: "FD-160A", Type: "analog", Drivers: []string{"fdx-160a"},
		Notes: chirpNote("FD160ARadio") + " 99 canali VHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
	{ID: "fdx-fd-450a", Brand: "Feidaxin", Name: "FD-450A", Type: "analog", Drivers: []string{"fdx-450a"},
		Notes: chirpNote("FD450ARadio") + " 99 canali UHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
	{ID: "fdx-fd-460a", Brand: "Feidaxin", Name: "FD-460A", Type: "analog", Drivers: []string{"fdx-460a"},
		Notes: chirpNote("FD460ARadio") + " 99 canali UHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
	{ID: "fdx-fd-460uh", Brand: "Feidaxin", Name: "FD-460UH", Type: "analog", Drivers: []string{"fdx-460uh"},
		Notes: chirpNote("FD460UHRadio") + " 99 canali UHF senza nomi. In CHIRP la lettura via cavo non funziona con Python 3: qui il protocollo è quello del suo driver, provato con la radio simulata (da verificare con una radio vera). Verificata con lo stesso codice dei modelli con immagine di prova di CHIRP (FD-268A/B, FD-288B)."},
}
