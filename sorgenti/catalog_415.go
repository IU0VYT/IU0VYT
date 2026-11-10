// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: palmari Icom D-STAR e IC-E90, Yaesu VX e System Fusion, Kenwood TH-D72 /
// TH-D74 / TH-D75 (driver di CHIRP riscritti in Go), firmware liberi nicFW (TD-H3,
// RT-880, RT-950 Pro) e F4HWN Fusion; palmari Kenwood PMR446 e DMR delle pagine
// ufficiali, per ora senza gestione.

const (
	kenwoodWeb    = "https://kenwood.it/comm/"
	kenwoodPmr    = "https://kenwood.it/comm/pmr-446/"
	kenwoodDmr    = "https://kenwood.eu/comm/digital/dmr/portable/"
	nicfwH3Docs   = "https://github.com/nicsure/nicfw2docs"
	nicfw880Docs  = "https://github.com/nicsure/nicFW880_950_Docs"
	nicfwSite     = "https://nicsure.co.uk/"
	f4hwnFusion   = "https://github.com/armel/uv-k1-k5v3-firmware-custom/releases"
	dstarNote     = " I canali sono analogici o digitali D-STAR (colonna «Modo»); dei canali D-STAR si conservano nominativi e codice digitale."
	yaesuCloneTip = " La copia parte dalla radio (modalità «CLONE»): le istruzioni sono nella finestra del cavo."
	kenwoodNoCps  = "Radio professionale: si programma con il programma KPG di Kenwood, riservato ai rivenditori; il protocollo non è pubblico."
	icDstarNote   = " 500 canali con nomi di 16 caratteri, copia veloce." + dstarNote
	ysfCatNote    = " 900 canali con nomi di 16 caratteri, analogici o digitali C4FM (colonna «Modo»)." + yaesuCloneTip
	vxCatNote     = " 999 canali con nomi di 6 caratteri." + yaesuCloneTip
	thdCatNote    = " 1000 canali con nomi di 16 caratteri, cavo USB (la radio compare come porta seriale)." + dstarNote
	f5CatNote     = " Firmware libero F4HWN Fusion dalla versione 5: 1024 canali (scheda «firmware F4HWN Fusion 5.x»)."
	nicfwCatNote  = "Firmware libero di nicsure: canali scritti da zero dalla documentazione pubblicata dall'autore e verificati con la radio simulata " +
		"(da verificare con una radio vera). La taratura della radio non viene mai scritta."
)

var catalog415 = []catItem{
	// Icom
	{ID: "ic-id-31", Brand: "Icom", Name: "ID-31A / ID-31E", Aliases: []string{"ID-31A", "ID-31E", "ID-31"}, Type: "analog", Drivers: []string{"ic-id31"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ID31Radio") + icDstarNote},
	{ID: "ic-id-51", Brand: "Icom", Name: "ID-51A / ID-51E", Aliases: []string{"ID-51A", "ID-51E", "ID-51"}, Type: "analog", Drivers: []string{"ic-id51"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ID51Radio") + icDstarNote},
	{ID: "ic-id-51-plus", Brand: "Icom", Name: "ID-51 Plus", Aliases: []string{"ID-51A Plus", "ID-51E Plus", "ID-51 50th Anniversary"}, Type: "analog",
		Drivers: []string{"ic-id51plus"}, Links: []catLink{info("Sito Icom", icomWeb)},
		Notes: chirpNote("ID51PLUSRadio") + icDstarNote},
	{ID: "ic-id-51-plus2", Brand: "Icom", Name: "ID-51 Plus2", Aliases: []string{"ID-51A Plus2", "ID-51E Plus2"}, Type: "analog",
		Drivers: []string{"ic-id51plus2"}, Links: []catLink{info("Sito Icom", icomWeb)},
		Notes: chirpNote("ID51PLUS2Radio") + icDstarNote},
	{ID: "ic-e90", Brand: "Icom", Name: "IC-E90 / IC-T90", Aliases: []string{"IC-E90", "IC-T90", "IC-T90A"}, Type: "analog", Drivers: []string{"ic-e90"},
		Links: []catLink{info("Sito Icom", icomWeb)}, Notes: chirpNote("ICx90Radio") + " Tribanda 50/144/430 MHz con ricevitore a larga banda: 500 canali con nomi di 6 caratteri."},
	// Yaesu
	{ID: "ys-vx-6", Brand: "Yaesu", Name: "VX-6R / VX-6E", Aliases: []string{"VX-6R", "VX-6E", "VX-6"}, Type: "analog", Drivers: []string{"ys-vx6"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX6Radio") + vxCatNote},
	{ID: "ys-vx-3", Brand: "Yaesu", Name: "VX-3R / VX-3E", Aliases: []string{"VX-3R", "VX-3E", "VX-3"}, Type: "analog", Drivers: []string{"ys-vx3"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("VX3Radio") + vxCatNote},
	{ID: "ys-ft-70d", Brand: "Yaesu", Name: "FT-70DR / FT-70DE", Aliases: []string{"FT-70DR", "FT-70DE", "FT-70D"}, Type: "analog", Drivers: []string{"ys-ft70d"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT70Radio") + " 900 canali con nomi di 6 caratteri; il modo digitale C4FM di ogni canale resta " +
			"quello impostato sulla radio." + yaesuCloneTip},
	{ID: "ys-ft1d", Brand: "Yaesu", Name: "FT1DR / FT1DE", Aliases: []string{"FT1DR", "FT1DE", "FT-1D", "FT1XD"}, Type: "analog", Drivers: []string{"ys-ft1d"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT1Radio") + ysfCatNote},
	{ID: "ys-ft2d", Brand: "Yaesu", Name: "FT2DR / FT2DE", Aliases: []string{"FT2DR", "FT2DE", "FT-2D"}, Type: "analog", Drivers: []string{"ys-ft2d", "ys-ft2dv2"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT2D") + ysfCatNote},
	{ID: "ys-ft3d", Brand: "Yaesu", Name: "FT3DR / FT3DE", Aliases: []string{"FT3DR", "FT3DE", "FT-3D"}, Type: "analog", Drivers: []string{"ys-ft3d"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT3D") + ysfCatNote},
	{ID: "ys-ft5d", Brand: "Yaesu", Name: "FT5DR / FT5DE", Aliases: []string{"FT5DR", "FT5DE", "FT-5D"}, Type: "analog", Drivers: []string{"ys-ft5d"},
		Links: []catLink{info("Sito Yaesu", yaesuWeb)}, Notes: chirpNote("FT5D") + ysfCatNote},
	// Kenwood
	{ID: "kw-th-d75", Brand: "Kenwood", Name: "TH-D75A / TH-D75E", Aliases: []string{"TH-D75E", "TH-D75A", "TH-D75"}, Type: "analog", Drivers: []string{"kw-thd75"},
		Links: []catLink{info("Pagina Kenwood della TH-D75E", "https://kenwood.it/comm/amatoriale/vhf-uhf/TH-D75E/")},
		Notes: chirpNote("THD75Radio") + thdCatNote},
	{ID: "kw-th-d74", Brand: "Kenwood", Name: "TH-D74A / TH-D74E", Aliases: []string{"TH-D74E", "TH-D74A", "TH-D74"}, Type: "analog", Drivers: []string{"kw-thd74"},
		Links: []catLink{info("Sito Kenwood", kenwoodWeb)}, Notes: chirpNote("THD74Radio") + thdCatNote},
	{ID: "kw-th-d72", Brand: "Kenwood", Name: "TH-D72A / TH-D72E", Aliases: []string{"TH-D72E", "TH-D72A", "TH-D72"}, Type: "analog", Drivers: []string{"kw-thd72"},
		Links: []catLink{info("Sito Kenwood", kenwoodWeb)}, Notes: chirpNote("THD72Radio") + " 1000 canali con nomi di 8 caratteri, cavo USB."},
	{ID: "kw-tk-3701d", Brand: "Kenwood", Name: "TK-3701DE", Aliases: []string{"TK-3701D", "ProTalk TK-3701D"}, Type: "pmr",
		Links: []catLink{info("Pagina Kenwood (PMR446)", kenwoodPmr)}, Notes: kenwoodNoCps},
	{ID: "kw-tk-3501", Brand: "Kenwood", Name: "TK-3501E", Aliases: []string{"TK-3501", "ProTalk TK-3501"}, Type: "pmr",
		Links: []catLink{info("Pagina Kenwood (PMR446)", kenwoodPmr)}, Notes: kenwoodNoCps},
	{ID: "kw-pkt-300", Brand: "Kenwood", Name: "PKT-300E", Aliases: []string{"PKT-300"}, Type: "pmr",
		Links: []catLink{info("Pagina Kenwood (PMR446)", kenwoodPmr)}, Notes: kenwoodNoCps},
	{ID: "kw-ubz-lj9", Brand: "Kenwood", Name: "UBZ-LJ9", Aliases: []string{"UBZ-LJ9SET"}, Type: "pmr",
		Links: []catLink{info("Pagina Kenwood (PMR446)", kenwoodPmr)}, Notes: "PMR446 analogico di uso libero: i canali si scelgono dalla radio, non c'è un programma di programmazione."},
	{ID: "kw-nx-1200", Brand: "Kenwood", Name: "NX-1200D / NX-1300D", Aliases: []string{"NX-1200DE", "NX-1200DE2", "NX-1200DE3", "NX-1300DE", "NX-1300DE2", "NX-1300DE3"},
		Models: []string{"NX-1200D", "NX-1300D"}, Type: "dmr", Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)}, Notes: kenwoodNoCps},
	{ID: "kw-nx-3200", Brand: "Kenwood", Name: "NX-3200 / NX-3300", Aliases: []string{"NX-3200E", "NX-3200E2", "NX-3200E3", "NX-3300E", "NX-3300E2", "NX-3300E3"},
		Models: []string{"NX-3200", "NX-3300"}, Type: "dmr", Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)}, Notes: kenwoodNoCps},
	{ID: "kw-nx-3220", Brand: "Kenwood", Name: "NX-3220 / NX-3320", Aliases: []string{"NX-3220E", "NX-3220E2", "NX-3220E3", "NX-3320E", "NX-3320E2", "NX-3320E3"},
		Models: []string{"NX-3220", "NX-3320"}, Type: "dmr", Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)}, Notes: kenwoodNoCps},
	{ID: "kw-nx-5200", Brand: "Kenwood", Name: "NX-5200 / NX-5300", Aliases: []string{"NX-5200E", "NX-5200E2", "NX-5300E", "NX-5300E2"},
		Models: []string{"NX-5200", "NX-5300"}, Type: "dmr", Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)}, Notes: kenwoodNoCps},
	{ID: "kw-tk-d240", Brand: "Kenwood", Name: "TK-D240E / TK-D340E", Aliases: []string{"TK-D240", "TK-D340"}, Models: []string{"TK-D240E", "TK-D340E"}, Type: "dmr",
		Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)}, Notes: kenwoodNoCps},
	{ID: "kw-tk-d200", Brand: "Kenwood", Name: "TK-D200E / TK-D300E", Aliases: []string{"TK-D200GE", "TK-D200E2", "TK-D200GE2", "TK-D300GE", "TK-D300E2", "TK-D300GE2"},
		Models: []string{"TK-D200E", "TK-D300E"}, Type: "dmr", Links: []catLink{info("Palmari DMR Kenwood", kenwoodDmr)},
		Notes: kenwoodNoCps},
	// firmware liberi
	{ID: "tid-td-h3-nicfw", Brand: "TIDRADIO", Name: "TD-H3 (firmware nicFW)", Aliases: []string{"TD-H3 nicFW", "nicFW"}, Type: "analog", Drivers: []string{"td-tdh3-nicfw"}, FW: "tdh",
		Links: []catLink{info("Documentazione di nicFW 2.x (nicsure)", nicfwH3Docs), info("Sito di nicsure", nicfwSite), info("Programmi TIDRADIO", tidDown)},
		Notes: nicfwCatNote + " " + tdhFwNote},
	{ID: "rtl-rt-880", Brand: "Radtel", Name: "RT-880", Aliases: []string{"RT-880 (non G)", "iRadio UV-98"}, Type: "analog", Drivers: []string{"rtl-rt880-nicfw"},
		Links: []catLink{info("Programmi Radtel", rtlDown), info("Documentazione di nicFW880 (nicsure)", nicfw880Docs)},
		Notes: "Con il firmware originale di Radtel la RT-880 (senza GPS) si programma con il programma di Radtel: il suo protocollo non è ancora gestito " +
			"(la RT-880G sì, con la sua scheda). " + nicfwCatNote},
	{ID: "rtl-rt-880-nicfw", Brand: "Radtel", Name: "RT-880 / RT-880G (firmware nicFW880)", Aliases: []string{"nicFW880", "RT-880 nicFW", "RT-880G nicFW"}, Type: "analog",
		Drivers: []string{"rtl-rt880-nicfw"}, Links: []catLink{info("Documentazione di nicFW880 (nicsure)", nicfw880Docs), info("Sito di nicsure", nicfwSite)},
		Notes: nicfwCatNote},
	{ID: "rtl-rt-950-pro-nicfw", Brand: "Radtel", Name: "RT-950 Pro (firmware nicFW950)", Aliases: []string{"nicFW950", "RT-950 Pro nicFW"}, Type: "analog",
		Drivers: []string{"rtl-rt950pro-nicfw"}, Links: []catLink{info("Documentazione di nicFW880 / 950 (nicsure)", nicfw880Docs), info("Sito di nicsure", nicfwSite)},
		Notes: nicfwCatNote},
}
