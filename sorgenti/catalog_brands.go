// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.0: radio di altre marche gestite con i driver di CHIRP (classi di
// CHIRP con la stessa memoria e lo stesso protocollo delle famiglie già nel
// programma, oppure famiglie nuove come la Quansheng UV-K5). Ogni modello è
// verificato con i test automatici: lettura, scrittura e byte scambiati via cavo
// identici a CHIRP.
//
// Fonti: driver e tabella dei modelli di CHIRP (https://chirpmyradio.com),
// pagine dei programmi dei costruttori: retevis.com/downloads/software,
// radtels.com/pages/software-download, en.qsfj.com/support/downloads,
// tidradio.com/pages/programming-software.

const (
	rtvDown = "https://www.retevis.com/downloads/software"
	rtlDown = "https://www.radtels.com/pages/software-download"
	qsDown  = "https://en.qsfj.com/support/downloads"
	tidDown = "https://tidradio.com/pages/programming-software"
	rddDown = "https://www.radioddity.com/pages/download-center"
	hiroWeb = "https://globalhiroyasu.com/"
	tytWeb  = "https://www.tyt888.com/"
)

// rt21Fixed30: nota dei modelli GMRS della famiglia Retevis RT21
const rt21Fixed30 = " 30 canali GMRS senza nome." + " I canali con una frequenza GMRS fissa non si cancellano e non cambiano frequenza (come in CHIRP)."

func info(label, url string) catLink { return catLink{"info", label, url} }

// chirpNote: nota comune delle radio gestite con un driver di CHIRP
func chirpNote(cls string) string {
	return "Canali e impostazioni con il driver di CHIRP (classe " + cls + "), verificato con i test automatici: " +
		"lettura, scrittura e byte scambiati con la radio identici a CHIRP. I file .img si aprono anche con CHIRP."
}

func fixedNote(what string) string {
	return " I canali con una frequenza " + what + " fissa non si cancellano e non cambiano frequenza (come in CHIRP)."
}

var catalogOther = []catItem{
	// ------------------------------------------------ Quansheng
	{ID: "qs-uv-k5", Brand: "Quansheng", Name: "UV-K5 / UV-K6 / UV-5R Plus", Aliases: []string{"UV-K5", "UV-K6", "UV-5R Plus"},
		Models: []string{"UV-K5", "UV-K6", "UV-5R Plus"}, Type: "analog", Drivers: []string{"qs-uvk5"}, FW: "k5",
		Links: []catLink{info("Download center Quansheng", qsDown)},
		Notes: chirpNote("UVK5Radio") + " Solo con le versioni del firmware originale conosciute (k5_2.01, 3.00, 4.00, 5.00, 7.00/7.01, 1.02, «1o11»): " +
			"con un firmware modificato (egzumer, IJV, F4HWN…) il programma si ferma senza leggere né scrivere, come CHIRP. " +
			"Firmware: file .bin di Quansheng (es. k5_v2.01.26_publish.bin) con il caricatore della radio (accesa tenendo premuto il PTT); " +
			"il programma controlla che il firmware sia per il tipo di caricatore della radio (UV-K5, UV-K5(8)/UV-K6 o UV-5R Plus). " +
			"Quansheng protegge i download con una password che fornisce il rivenditore."},
	{ID: "qs-uv-k1", Brand: "Quansheng", Name: "UV-K1", Type: "analog", Drivers: []string{"qs-uvk1"},
		Links: []catLink{info("Download center Quansheng", qsDown)},
		Notes: chirpNote("UVK1Radio") + " Solo canali (CHIRP non gestisce ancora le impostazioni) e solo con il firmware 7.03."},
	{ID: "rt-ra79", Brand: "Retevis", Name: "RA79", Type: "analog", Drivers: []string{"rt-ra79"}, FW: "k5",
		Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: "È una Quansheng UV-K5. " + chirpNote("RA79Radio") +
			" Firmware: file .bin di Retevis nel formato di Quansheng (es. «RA79 V3.bin», versione 7.00.16 per la RA79 V3, con il caricatore 7.x della " +
			"UV-K5 V3): radio spenta e riaccesa tenendo premuto il PTT. Il programma legge la versione del caricatore e invia il firmware solo se è " +
			"della stessa famiglia (le RA79 con il caricatore della UV-K5 classica rifiutano i file 7.x)."},
	{ID: "mt-tk6", Brand: "MaxTalker", Name: "TK-6", Type: "analog", Drivers: []string{"mt-tk6"}, Notes: "È una Quansheng UV-K5. " + chirpNote("MaxTalkerTK6")},

	// ------------------------------------------------ Radtel
	// 4.2: Radtel RT-590 / RT-600 (la stessa radio della Quansheng UV-K5 / UV-5R
	// Plus: file del firmware di Radtel nel formato di Quansheng, verificati con uvtools)
	{ID: "rtl-rt-590", Brand: "Radtel", Name: "RT-590 / RT-590 Plus", Aliases: []string{"RT-590", "RT-590 Plus"}, Type: "analog", Drivers: []string{"qs-uvk5"}, FW: "k5",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)},
		Notes: "È la Quansheng UV-K5: canali e impostazioni con il driver di CHIRP della UV-K5 (firmware originale 2.01.xx). " +
			"Firmware: il file .bin di Radtel (es. «590 231110_v2.01.32_publish.bin», in RT590_update_file_v2.01.32.rar) con il caricatore " +
			"della radio (accesa tenendo premuto il PTT, torcia accesa), come il programma di Radtel."},
	{ID: "rtl-rt-600", Brand: "Radtel", Name: "RT-600 / RT-600 Plus", Aliases: []string{"RT-600", "RT-600 Plus"}, Type: "analog", Drivers: []string{"qs-uvk5"}, FW: "k5",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)},
		Notes: "È la Quansheng UV-5R Plus: canali e impostazioni con il driver di CHIRP della UV-K5 (firmware originale 4.00.xx). " +
			"Firmware: il file .bin di Radtel (es. «5RPLUS_v4.00.06_publish.bin», in RT600_update_file_v4.00.06.rar) con il caricatore " +
			"della radio (accesa tenendo premuto il PTT)."},
	{ID: "rtl-rt-900", Brand: "Radtel", Name: "RT-900", Type: "analog", Drivers: []string{"rt-rt900"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT900") + " 999 canali."},
	{ID: "rtl-rt-900-bt", Brand: "Radtel", Name: "RT-900 BT", Aliases: []string{"RT-900 Bluetooth"}, Type: "analog", Drivers: []string{"rt-rt900bt"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT900BT") + " 999 canali."},
	{ID: "rtl-rt-910", Brand: "Radtel", Name: "RT-910 / RT-910B", Aliases: []string{"RT-910", "RT-910B", "RT-910 BT"}, Models: []string{"RT-910", "RT-910B"}, Type: "analog",
		Drivers: []string{"bj-7700", "bj-7700nbt"}, Links: []catLink{info("Programmi e firmware Radtel", rtlDown)},
		Notes: "Stessa radio della Bajeton BJ-7700: driver di CHIRP RT910BT (versione Bluetooth) e RT910."},
	{ID: "rtl-rt-920", Brand: "Radtel", Name: "RT-920", Type: "analog", Drivers: []string{"rt-rt920"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT920") + " 990 canali (i canali ausiliari FM/AM/HF non sono gestiti)."},
	{ID: "rtl-rt-950-pro", Brand: "Radtel", Name: "RT-950 Pro", Type: "analog", Drivers: []string{"bj-9000"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)},
		Notes: "Stessa radio della Bajeton BJ-9000: canali e impostazioni come nella scheda BJ-9000. Il firmware Radtel si aggiorna con il programma del costruttore."},
	// 4.1: famiglia MML JC-8810 (mml_jc8810.py); 4.3: firmware .kdhx di Radtel
	{ID: "rtl-rt-470", Brand: "Radtel", Name: "RT-470", Type: "analog", Drivers: []string{"rt-rt470"}, FW: "kdhx",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470Radio") + " 256 canali, potenze 10/8/4 W. " + rtlFwNote},
	{ID: "rtl-rt-470l", Brand: "Radtel", Name: "RT-470L", Type: "analog", Drivers: []string{"rt-rt470l"}, FW: "kdhx",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470LRadio") + " 256 canali. " + rtlFwNote},
	{ID: "rtl-rt-470x", Brand: "Radtel", Name: "RT-470X", Type: "analog", Drivers: []string{"rt-rt470x"}, FW: "kdhx",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470XRadio") + " 256 canali; in banda aerea (118-136 MHz) la radio riceve in AM. " + rtlFwNote},
	{ID: "rtl-rt-470x-plus", Brand: "Radtel", Name: "RT-470X Plus", Aliases: []string{"RT-470X BT", "RT-470X Plus BT"}, Type: "analog", Drivers: []string{"rt-rt470xbt"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470XPlusRadio") + " Versione con Bluetooth della RT-470X."},
	{ID: "rtl-rt-630", Brand: "Radtel", Name: "RT-630", Type: "analog", Drivers: []string{"rt-rt630"}, FW: "kdhx",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT630Radio") + " 256 canali. " + rtlFwNote},
	{ID: "rtl-rt-495", Brand: "Radtel", Name: "RT-495", Type: "analog", Drivers: []string{"rt-rt495"}, FW: "kdhx",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT495Radio") + " 256 canali, piattaforma RT-630. " + rtlFwNote},
	{ID: "hy-hi-8811", Brand: "Hiroyasu", Name: "HI-8811", Type: "analog", Drivers: []string{"hy-hi8811"},
		Links: []catLink{info("Sito Hiroyasu", hiroWeb)}, Notes: chirpNote("HI8811Radio") + " Piattaforma Radtel RT-470L."},
	{ID: "as-uv-a37", Brand: "Anysecu", Name: "UV-A37", Type: "analog", Drivers: []string{"as-uva37"}, Notes: chirpNote("UVA37Radio") + " 256 canali."},
	{ID: "tp-a36plus", Brand: "Talkpod", Name: "A36plus", Aliases: []string{"A36 Plus"}, Type: "analog", Drivers: []string{"tp-a36plus"},
		Notes: chirpNote("A36plusRadio") + " 512 canali (firmware 1.22 o successivo)."},
	{ID: "tp-a36plus-8w", Brand: "Talkpod", Name: "A36plus 8 W", Aliases: []string{"A36 Plus 8W"}, Type: "analog", Drivers: []string{"tp-a36plus8w"},
		Notes: chirpNote("A36plus8wRadio") + " 512 canali."},
	{ID: "ab-ar-730", Brand: "Abbree", Name: "AR-730", Type: "analog", Drivers: []string{"ab-ar730"}, Notes: chirpNote("AR730Radio") + " Piattaforma Anysecu UV-A37."},
	{ID: "rtl-t18", Brand: "Radtel", Name: "T18", Type: "pmr", Drivers: []string{"rt-t18"}, Notes: chirpNote("T18Radio") + " 16 canali senza nome."},

	// ------------------------------------------------ Radioddity
	{ID: "rdd-gm-30", Brand: "Radioddity", Name: "GM-30", Type: "gmrs", Drivers: []string{"rd-gm30"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: chirpNote("RadioddityGM30") + " Radio GMRS (USA): canali 1-30 GMRS fissi, 23-54 ponti GMRS, 55-250 solo ricezione."},
	{ID: "rdd-gm-30-plus", Brand: "Radioddity", Name: "GM-30 Plus", Type: "gmrs", Drivers: []string{"rd-gm30plus"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: chirpNote("RadioddityGM30Plus") + " Stessa piattaforma della Baofeng GM-21."},
	{ID: "rdd-gm-30-pro", Brand: "Radioddity", Name: "GM-30 Pro", Type: "gmrs", Drivers: []string{"rd-gm30pro"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: chirpNote("RadioddityGM30Pro") + " Stessa piattaforma della Baofeng GM-21, con lo scrambler."},
	{ID: "rdd-gs-10b", Brand: "Radioddity", Name: "GS-10B", Type: "analog", Drivers: []string{"rd-gs10b"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: chirpNote("RadioddityGS10B") + " 256 canali, piattaforma Radtel RT-900."},
	{ID: "rdd-ga-2s", Brand: "Radioddity", Name: "GA-2S", Type: "pmr", Drivers: []string{"rd-ga2s"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: chirpNote("ROGA2SRadio") + " 16 canali senza nome (piattaforma della BF-888S)."},
	{ID: "rdd-ga-510", Brand: "Radioddity", Name: "GA-510", Type: "analog", Drivers: []string{"bf-hseries"}, Links: []catLink{info("Download center Radioddity", rddDown)},
		Notes: "Stessa radio della serie Baofeng H: driver di CHIRP della GA-510 (prima versione, 9600 baud). La GA-510 V2 (57600 baud) non è ancora gestita."},

	// ------------------------------------------------ TIDRADIO
	// 4.1: famiglia TD-H8 (tdh8.py)
	{ID: "tid-td-h8", Brand: "TIDRADIO", Name: "TD-H8", Aliases: []string{"TD-H8 HAM", "TD-H8 GMRS"}, Type: "analog",
		Drivers: []string{"td-tdh8", "td-tdh8ham", "td-tdh8gmrs", "td-tdh8g3", "td-tdh8g3ham", "td-tdh8g3gmrs"}, FW: "tdh",
		Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: chirpNote("TDH8") + " 199 canali con nome di 8 caratteri. Versioni: base, HAM e GMRS, di seconda e di terza generazione " +
			"(la radio comunica quale è: se scegli quella sbagliata non viene letta). " + tdhFwNote},
	{ID: "tid-td-h3", Brand: "TIDRADIO", Name: "TD-H3", Aliases: []string{"TD-H3 HAM", "TD-H3 GMRS"}, Type: "analog",
		Drivers: []string{"td-tdh3", "td-tdh3ham", "td-tdh3gmrs"}, FW: "tdh", Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: chirpNote("TDH3") + " 199 canali con nome di 8 caratteri; versioni base, HAM e GMRS. " + tdhFwNote},
	{ID: "rtl-rt-730", Brand: "Radtel", Name: "RT-730", Type: "analog", Drivers: []string{"rt-rt730"}, FW: "tdh",
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT730") + " Piattaforma TIDRADIO TD-H8, 199 canali. " + tdhFwNote},
	{ID: "tid-td-h6", Brand: "TIDRADIO", Name: "TD-H6", Type: "analog", Drivers: []string{"td-h6"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: chirpNote("TDH6Radio") + " Piattaforma Radioddity GA-510."},

	// ------------------------------------------------ Abbree, MaxTalker, Intek, Maverick
	{ID: "ab-ar-f5", Brand: "Abbree", Name: "AR-F5", Type: "analog", Drivers: []string{"ab-arf5"}, Notes: chirpNote("AbbreeARF5Radio") + " Canali 1-128, piattaforma GA-510."},
	{ID: "mt-p15", Brand: "MaxTalker", Name: "P15", Type: "analog", Drivers: []string{"mt-p15"}, Notes: chirpNote("MaxTalkerP15") + " Stessa piattaforma della Baofeng UV-5RM."},
	{ID: "mt-mt-5rm", Brand: "MaxTalker", Name: "MT-5RM", Type: "analog", Drivers: []string{"mt-mt5rm"}, Notes: chirpNote("MaxTalkerMT5RM") + " Stessa piattaforma della Baofeng UV-5RM."},
	{ID: "mt-mt-8s", Brand: "MaxTalker", Name: "MT-8S", Type: "pmr", Drivers: []string{"mt-mt8s"}, Notes: chirpNote("MT8SRadio") + " 16 canali senza nome (piattaforma della BF-888S)."},
	{ID: "in-kt-980hp", Brand: "Intek", Name: "KT-980HP", Type: "analog", Drivers: []string{"in-kt980hp"}, Notes: chirpNote("IntekKT980Radio") + " Piattaforma Baofeng BF-F8HP, tre potenze."},
	{ID: "mv-ra-100", Brand: "Maverick", Name: "RA-100", Type: "pmr", Drivers: []string{"mv-ra100"}, Notes: chirpNote("MAVERICKRA100Radio") + " Piattaforma Baofeng BF-M4."},
	{ID: "mv-ra-425", Brand: "Maverick", Name: "RA-425", Type: "pmr", Drivers: []string{"mv-ra425"}, Notes: chirpNote("MAVERICKRA425Radio") + " Piattaforma Baofeng BF-1904 (10 W)."},

	// ------------------------------------------------ BTECH
	{ID: "bt-frs-b1", Brand: "BTECH", Name: "FRS-B1", Type: "gmrs", Drivers: []string{"bt-frsb1"}, Notes: chirpNote("FRSB1Radio") + fixedNote("FRS")},

	// ------------------------------------------------ Retevis
	{ID: "rtv-ra685", Brand: "Retevis", Name: "RA685", Type: "analog", Drivers: []string{"rt-ra685"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisRA685Radio") + " Canali 1-128, piattaforma GA-510."},
	{ID: "rtv-ra85", Brand: "Retevis", Name: "RA85", Type: "analog", Drivers: []string{"rt-ra85"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisRA85Radio") + " Canali 1-128, piattaforma GA-510."},
	{ID: "rtv-rt6", Brand: "Retevis", Name: "RT6", Type: "analog", Drivers: []string{"rt-rt6"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RT6") + " È la Baofeng WP970I (famiglia UV-9R)."},
	{ID: "rtv-h777", Brand: "Retevis", Name: "H777", Type: "pmr", Drivers: []string{"rt-h777"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RetevisH777") + " 16 canali senza nome (piattaforma della BF-888S)."},
	{ID: "rtv-h777-plus", Brand: "Retevis", Name: "H777 Plus", Type: "pmr", Drivers: []string{"rt-h777plus"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("H777PlusRadio") + " 16 canali senza nome."},
	{ID: "rtv-h777h", Brand: "Retevis", Name: "H777H", Type: "pmr", Drivers: []string{"rt-h777h-pmr", "rt-h777h-frs"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "Due versioni: PMR446 (Europa) e FRS (USA). " + chirpNote("H777HPMRRadio / H777HFRSRadio") + fixedNote("PMR446 o FRS")},
	{ID: "rtv-rt22", Brand: "Retevis", Name: "RT22", Type: "pmr", Drivers: []string{"rt-rt22"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RT22Radio") + " 16 canali senza nome."},
	{ID: "rtv-rt22frs", Brand: "Retevis", Name: "RT22FRS", Type: "gmrs", Drivers: []string{"rt-rt22frs"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT22FRS")},
	{ID: "rtv-rt622", Brand: "Retevis", Name: "RT622", Type: "pmr", Drivers: []string{"rt-rt622"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RT622") + " Versione PMR446 della RT22."},
	{ID: "rtv-rt22s", Brand: "Retevis", Name: "RT22S", Type: "gmrs", Drivers: []string{"rt-rt22s"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT22SRadio") + fixedNote("FRS")},
	{ID: "rtv-rt20", Brand: "Retevis", Name: "RT20", Type: "pmr", Drivers: []string{"rt-rt20"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT20Radio")},
	{ID: "rtv-rt15", Brand: "Retevis", Name: "RT15", Type: "pmr", Drivers: []string{"rt-rt15"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT15Radio")},
	{ID: "rtv-rt47", Brand: "Retevis", Name: "RT47", Type: "gmrs", Drivers: []string{"rt-rt47"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT47Radio") + fixedNote("FRS")},
	{ID: "rtv-rt47v", Brand: "Retevis", Name: "RT47V", Type: "gmrs", Drivers: []string{"rt-rt47v"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT47VRadio") + fixedNote("MURS")},
	{ID: "rtv-rt647", Brand: "Retevis", Name: "RT647", Type: "pmr", Drivers: []string{"rt-rt647"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT647Radio") + fixedNote("PMR446")},
	{ID: "rtv-rt68", Brand: "Retevis", Name: "RT68", Type: "gmrs", Drivers: []string{"rt-rt68"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT68Radio") + fixedNote("FRS")},
	{ID: "rtv-rt668", Brand: "Retevis", Name: "RT668", Type: "pmr", Drivers: []string{"rt-rt668"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT668Radio") + fixedNote("PMR446")},
	{ID: "rtv-rb17", Brand: "Retevis", Name: "RB17", Type: "gmrs", Drivers: []string{"rt-rb17"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB17Radio") + fixedNote("FRS")},
	{ID: "rtv-rb17v", Brand: "Retevis", Name: "RB17V", Type: "gmrs", Drivers: []string{"rt-rb17v"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB17VRadio") + fixedNote("MURS")},
	{ID: "rtv-rb617", Brand: "Retevis", Name: "RB617", Type: "pmr", Drivers: []string{"rt-rb617"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB617Radio") + fixedNote("PMR446")},
	{ID: "rtv-rb18", Brand: "Retevis", Name: "RB18", Type: "gmrs", Drivers: []string{"rt-rb18"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RB18Radio") + fixedNote("FRS") + " Come in CHIRP la potenza dei canali si legge ma i canali scritti risultano ad alta potenza."},
	{ID: "rtv-rb618", Brand: "Retevis", Name: "RB618", Type: "pmr", Drivers: []string{"rt-rb618"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB618Radio") + fixedNote("PMR446")},
	{ID: "rtv-rb19", Brand: "Retevis", Name: "RB19", Type: "gmrs", Drivers: []string{"rt-rb19"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB19Radio") + fixedNote("FRS")},
	{ID: "rtv-rb19p", Brand: "Retevis", Name: "RB19P", Type: "gmrs", Drivers: []string{"rt-rb19p"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB19PRadio") + fixedNote("GMRS")},
	{ID: "rtv-rb619", Brand: "Retevis", Name: "RB619", Type: "pmr", Drivers: []string{"rt-rb619"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB619Radio") + fixedNote("PMR446")},
	{ID: "rtv-rb29", Brand: "Retevis", Name: "RB29", Type: "gmrs", Drivers: []string{"rt-rb29"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB29Radio") + fixedNote("FRS")},
	{ID: "rtv-rb629", Brand: "Retevis", Name: "RB629", Type: "pmr", Drivers: []string{"rt-rb629"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB629Radio") + fixedNote("PMR446")},
	{ID: "rtv-rb75", Brand: "Retevis", Name: "RB75", Type: "gmrs", Drivers: []string{"rt-rb75"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB75Radio")},
	{ID: "rtv-rb85", Brand: "Retevis", Name: "RB85", Type: "pmr", Drivers: []string{"rt-rb85"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB85Radio") + " 10 W."},
	{ID: "rtv-rb87", Brand: "Retevis", Name: "RB87", Type: "gmrs", Drivers: []string{"rt-rb87"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB87Radio") + fixedNote("GMRS")},

	// ------------------------------------------------ altre marche con la Retevis RT22
	{ID: "wln-kd-c1", Brand: "WLN", Name: "KD-C1", Type: "pmr", Drivers: []string{"wln-kdc1"}, Notes: "È la Retevis RT22. " + chirpNote("KDC1")},
	{ID: "zt-zt-x6", Brand: "Zastone", Name: "ZT-X6", Type: "pmr", Drivers: []string{"zt-ztx6"}, Notes: "È la Retevis RT22. " + chirpNote("ZTX6")},
	{ID: "lt-lt-316", Brand: "Luiton", Name: "LT-316", Type: "pmr", Drivers: []string{"lt-lt316"}, Notes: "È la Retevis RT22. " + chirpNote("LT316")},
	{ID: "tid-td-m8", Brand: "TID", Name: "TD-M8", Type: "pmr", Drivers: []string{"tid-tdm8"}, Notes: "È la Retevis RT22. " + chirpNote("TDM8")},
	// ------------------------------------------------ 4.2: famiglia TYT TH-UV88 (th_uv88.py)
	{ID: "tyt-th-uv88", Brand: "TYT", Name: "TH-UV88", Type: "analog", Drivers: []string{"tyt-thuv88"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("THUV88Radio") + " 200 canali con nome di 10 caratteri, tre potenze (5, 2,5 e 0,5 W)."},
	{ID: "tyt-th-uv98", Brand: "TYT", Name: "TH-UV98", Type: "analog", Drivers: []string{"tyt-thuv98"}, Links: []catLink{info("Sito TYT", tytWeb)},
		Notes: chirpNote("THUV98Radio") + " Piattaforma della TH-UV88 (protocollo «UV98»)."},
	{ID: "rtv-rt85", Brand: "Retevis", Name: "RT85", Type: "analog", Drivers: []string{"rt-rt85"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: "È la TYT TH-UV88. " + chirpNote("RT85")},
	{ID: "rtv-ra89", Brand: "Retevis", Name: "RA89", Type: "analog", Drivers: []string{"rt-ra89"}, FW: "th8600", Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("RA89") + " Piattaforma della TYT TH-UV88 (protocollo «UV99»). Firmware: file .bin di Retevis (es. " +
			"«Eng_RA89_Upadat_20250524_B91.38.bin») con il protocollo del programma «RA89-update-software» («TH8600-UpDAta Soft» 2.3): la radio " +
			"va prima portata in modalità di aggiornamento come indicato da Retevis. Se l'invio si interrompe non accendere la radio normalmente: " +
			"ripeti subito l'aggiornamento."},
	{ID: "rtv-p2", Brand: "Retevis", Name: "P2", Type: "analog", Drivers: []string{"rt-p2"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("P2") + " Radio senza display, piattaforma della TYT TH-UV88."},
	{ID: "rtv-p62", Brand: "Retevis", Name: "P62", Type: "analog", Drivers: []string{"rt-p62"}, Links: []catLink{info("Programmi Retevis", rtvDown)},
		Notes: chirpNote("P62") + " Radio senza display, piattaforma della TYT TH-UV88."},
	{ID: "ex-qrz-1", Brand: "Explorer", Name: "QRZ-1", Type: "analog", Drivers: []string{"ex-qrz1"},
		Notes: chirpNote("QRZ1") + " Piattaforma della TYT TH-UV88 (protocollo «UV78»)."},

	// ------------------------------------------------ 4.2: famiglia Radtel RT-490 (radtel_rt490.py)
	{ID: "rtl-rt-490", Brand: "Radtel", Name: "RT-490", Type: "analog", Drivers: []string{"rtl-rt490"}, Links: []catLink{info("Programmi e firmware Radtel", rtlDown)},
		Notes: chirpNote("RT490Radio") + " 256 canali con nome di 12 caratteri (anche in cinese, GB2312), bande aerea, VHF, 220, 350 e UHF."},
	{ID: "mml-jc-8629", Brand: "MMLradio", Name: "JC-8629", Type: "analog", Drivers: []string{"mml-jc8629"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("MML8629Radio")},
	{ID: "jjcc-jc-8629", Brand: "JJCC", Name: "JC-8629", Type: "analog", Drivers: []string{"jjcc-jc8629"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("JJCC8629Radio")},
	{ID: "soc-jc-8629", Brand: "Socotran", Name: "JC-8629", Type: "analog", Drivers: []string{"soc-jc8629"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("SocotranJC8629Radio")},
	{ID: "soc-fb-8629", Brand: "Socotran", Name: "FB-8629", Type: "analog", Drivers: []string{"soc-fb8629"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("SocotranFB8629Radio")},
	{ID: "jp-8800-plus", Brand: "Jianpai", Name: "8800 Plus", Type: "analog", Drivers: []string{"jp-8800plus"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("Jianpai8629Radio")},
	{ID: "bor-8rs", Brand: "Boristone", Name: "8RS", Type: "analog", Drivers: []string{"bor-8rs"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("Boristone8RSRadio")},
	{ID: "ab-ar-869", Brand: "Abbree", Name: "AR-869", Type: "analog", Drivers: []string{"ab-ar869"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("AbbreeAR869Radio")},
	{ID: "hg-hg-590", Brand: "HamGeek", Name: "HG-590", Type: "analog", Drivers: []string{"hg-hg590"}, Notes: "Stessa radio della Radtel RT-490. " + chirpNote("HamGeekHG590Radio")},

	// ------------------------------------------------ 4.2: famiglia Retevis RT21 (retevis_rt21.py)
	{ID: "rtv-rt21", Brand: "Retevis", Name: "RT21", Type: "pmr", Drivers: []string{"rt-rt21"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT21Radio") + " 16 canali UHF senza nome."},
	{ID: "rtv-rt21v", Brand: "Retevis", Name: "RT21V", Type: "pmr", Drivers: []string{"rt-rt21v"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT21VRadio") + " 5 canali VHF senza nome."},
	{ID: "rtv-rb17a", Brand: "Retevis", Name: "RB17A", Type: "gmrs", Drivers: []string{"rt-rb17a"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB17ARadio") + rt21Fixed30},
	{ID: "rtv-rb23", Brand: "Retevis", Name: "RB23", Type: "gmrs", Drivers: []string{"rt-rb23"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB23Radio") + rt21Fixed30},
	{ID: "rtv-rb26", Brand: "Retevis", Name: "RB26", Type: "gmrs", Drivers: []string{"rt-rb26"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB26Radio") + rt21Fixed30},
	{ID: "rtv-rb626", Brand: "Retevis", Name: "RB626", Type: "pmr", Drivers: []string{"rt-rb626"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB626") + " 32 canali senza nome."},
	{ID: "rtv-rb89", Brand: "Retevis", Name: "RB89", Type: "gmrs", Drivers: []string{"rt-rb89"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB89Radio") + " 30 canali senza nome (in CHIRP senza canali fissi)."},
	{ID: "rtv-rt76", Brand: "Retevis", Name: "RT76", Type: "gmrs", Drivers: []string{"rt-rt76"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT76Radio") + rt21Fixed30},
	{ID: "rtv-rt29", Brand: "Retevis", Name: "RT29", Aliases: []string{"RT29 UHF", "RT29 VHF"}, Type: "analog", Drivers: []string{"rt-rt29u", "rt-rt29v"},
		Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT29UHFRadio / RT29VHFRadio") + " Versioni UHF e VHF, 16 canali senza nome, 10 W."},
	{ID: "rtv-rt19", Brand: "Retevis", Name: "RT19", Type: "gmrs", Drivers: []string{"rt-rt19"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT19Radio") + " 22 canali FRS." + fixedNote("FRS")},
	{ID: "rtv-rt619", Brand: "Retevis", Name: "RT619", Type: "pmr", Drivers: []string{"rt-rt619"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT619Radio") + " 16 canali PMR446." + fixedNote("PMR446")},
	{ID: "ab-ar-63", Brand: "Abbree", Name: "AR-63", Type: "pmr", Drivers: []string{"ab-ar63"}, Notes: chirpNote("AR63Radio") + " 16 canali UHF senza nome."},
	{ID: "rtv-rt40b", Brand: "Retevis", Name: "RT40B", Type: "gmrs", Drivers: []string{"rt-rt40b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT40BRadio") + " 22 canali GMRS." + fixedNote("GMRS")},
	{ID: "rtv-rb28b", Brand: "Retevis", Name: "RB28B", Type: "gmrs", Drivers: []string{"rt-rb28b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB28BRadio") + " 22 canali FRS." + fixedNote("FRS")},
	{ID: "rtv-rb628b", Brand: "Retevis", Name: "RB628B", Type: "pmr", Drivers: []string{"rt-rb628b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RB628BRadio") + " 16 canali PMR446." + fixedNote("PMR446")},
	{ID: "rtv-rt86", Brand: "Retevis", Name: "RT86", Type: "pmr", Drivers: []string{"rt-rt86"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT86Radio") + " 16 canali UHF senza nome."},
	{ID: "rtv-rt86s", Brand: "Retevis", Name: "RT86S", Type: "pmr", Drivers: []string{"rt-rt86s"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT86SRadio") + " 30 canali senza nome, tre potenze."},
	{ID: "rtv-rt86a", Brand: "Retevis", Name: "RT86A", Type: "pmr", Drivers: []string{"rt-rt86a"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: chirpNote("RT86ARadio") + " 30 canali senza nome; potenza alta o media (come CHIRP)."},

	// ------------------------------------------------ 4.2: Quansheng TK11 (tk11.py)
	{ID: "qs-tk11", Brand: "Quansheng", Name: "TK11", Type: "analog", Drivers: []string{"qs-tk11"}, Links: []catLink{info("Download center Quansheng", qsDown)},
		Notes: chirpNote("TK11") + " 999 canali con nome di 16 caratteri, ricezione da 0,15 a 1160 MHz (AM, SSB e CW restano come sono), tre potenze."},
}

// rtlFwNote (4.3): firmware .kdhx di Radtel (radtels.com) per RT-470, RT-470L,
// RT-470X, RT-495 e RT-630
const rtlFwNote = "Firmware: file .kdhx di Radtel con il caricatore «JJCC Bootloader Pro» (stesso protocollo del «Bootloader Pro» della UV-25 Plus, " +
	"presentazione automatica della piattaforma JJCC). Modalità manuale: radio accesa tenendo premuti i tasti laterali 2 e 3. La RT-470 ha due schede: " +
	"il firmware 1.x è solo per la scheda vecchia, il 2.x solo per la nuova (versione nel menu 55): un file sbagliato blocca la radio."

// tdhFwNote (4.3): firmware .BIN di TD-H3, TD-H8 e RT-730
const tdhFwNote = "Firmware: file .BIN del costruttore con il caricatore comune a TD-H3, TD-H8 e RT-730 (lo stesso dei programmi «H8_IAP» di TIDRADIO e " +
	"«RT730 Update tool» di Radtel): radio spenta e riaccesa tenendo premuto il PTT. Il caricatore non controlla il modello del firmware: " +
	"si deve scrivere il modello indicato dal file. Se l'invio si interrompe la radio resta nel caricatore e si ripete l'aggiornamento."
