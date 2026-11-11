// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.23: Kenwood TK-481 e TK-981 (driver di CHIRP, radio a 900 MHz divise in
// sistemi), TK-780 e TK-880 (veicolari della famiglia TK-280 / TK-380), Radtel
// RT-13 e RT-69 GPS (dai programmi del costruttore).

const kwSysCatNote = " Radio a 900 MHz (banda non usata in Italia), sempre a banda stretta. I canali sono divisi in sistemi: la lista ha 250 " +
	"posizioni per ognuno dei 32 sistemi e la posizione è il numero del canale nel sistema. Si usano solo i sistemi che la radio ha " +
	"già; quelli a risorse condivise (trunking) non si modificano. Si programma col cavo di programmazione Kenwood."

var catalog423 = []catItem{
	{ID: "kw-tk-481", Brand: "Kenwood", Name: "TK-481", Aliases: []string{"TK481"},
		Type: "analog", Drivers: []string{"kw-tk481"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK481") + kwSysCatNote},
	{ID: "kw-tk-981", Brand: "Kenwood", Name: "TK-981", Aliases: []string{"TK981"},
		Type: "analog", Drivers: []string{"kw-tk981"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK981") + kwSysCatNote},
	{ID: "kw-tk-780", Brand: "Kenwood", Name: "TK-780 / TK-880", Aliases: []string{"TK-780", "TK-880", "TK780", "TK880"},
		Type: "analog", Drivers: []string{"kw-tk780", "kw-tk880"}, Links: []catLink{info("Sito Kenwood", kenwoodWeb)},
		Notes: chirpNote("TK780_Radios") + kwZoneCatNote},
	{ID: "rtl-rt-13", Brand: "Radtel", Name: "RT-13", Aliases: []string{"RT13"}, Type: "analog",
		Drivers: []string{"rtl-rt13"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt13Note},
	{ID: "rtl-rt-69-gps", Brand: "Radtel", Name: "RT-69 GPS", Aliases: []string{"RT-69", "RT69", "RT69GPS", "889G"}, Type: "analog",
		Drivers: []string{"rtl-rt69"}, Links: []catLink{info("Programmi Radtel", rtlDown)},
		Notes: rt69Note},
}
