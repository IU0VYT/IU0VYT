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
)

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
	{ID: "rt-ra79", Brand: "Retevis", Name: "RA79", Type: "analog", Drivers: []string{"rt-ra79"},
		Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: "È una Quansheng UV-K5. " + chirpNote("RA79Radio")},
	{ID: "mt-tk6", Brand: "MaxTalker", Name: "TK-6", Type: "analog", Drivers: []string{"mt-tk6"}, Notes: "È una Quansheng UV-K5. " + chirpNote("MaxTalkerTK6")},

	// ------------------------------------------------ Radtel
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
	// 4.1: famiglia MML JC-8810 (mml_jc8810.py)
	{ID: "rtl-rt-470", Brand: "Radtel", Name: "RT-470", Type: "analog", Drivers: []string{"rt-rt470"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470Radio") + " 256 canali, potenze 10/8/4 W."},
	{ID: "rtl-rt-470l", Brand: "Radtel", Name: "RT-470L", Type: "analog", Drivers: []string{"rt-rt470l"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470LRadio") + " 256 canali."},
	{ID: "rtl-rt-470x", Brand: "Radtel", Name: "RT-470X", Type: "analog", Drivers: []string{"rt-rt470x"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470XRadio") + " 256 canali; in banda aerea (118-136 MHz) la radio riceve in AM."},
	{ID: "rtl-rt-470x-plus", Brand: "Radtel", Name: "RT-470X Plus", Aliases: []string{"RT-470X BT", "RT-470X Plus BT"}, Type: "analog", Drivers: []string{"rt-rt470xbt"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT470XPlusRadio") + " Versione con Bluetooth della RT-470X."},
	{ID: "rtl-rt-630", Brand: "Radtel", Name: "RT-630", Type: "analog", Drivers: []string{"rt-rt630"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT630Radio") + " 256 canali."},
	{ID: "rtl-rt-495", Brand: "Radtel", Name: "RT-495", Type: "analog", Drivers: []string{"rt-rt495"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT495Radio") + " 256 canali, piattaforma RT-630."},
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
		Drivers: []string{"td-tdh8", "td-tdh8ham", "td-tdh8gmrs", "td-tdh8g3", "td-tdh8g3ham", "td-tdh8g3gmrs"},
		Links:   []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: chirpNote("TDH8") + " 199 canali con nome di 8 caratteri. Versioni: base, HAM e GMRS, di seconda e di terza generazione " +
			"(la radio comunica quale è: se scegli quella sbagliata non viene letta)."},
	{ID: "tid-td-h3", Brand: "TIDRADIO", Name: "TD-H3", Aliases: []string{"TD-H3 HAM", "TD-H3 GMRS"}, Type: "analog",
		Drivers: []string{"td-tdh3", "td-tdh3ham", "td-tdh3gmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: chirpNote("TDH3") + " 199 canali con nome di 8 caratteri; versioni base, HAM e GMRS."},
	{ID: "rtl-rt-730", Brand: "Radtel", Name: "RT-730", Type: "analog", Drivers: []string{"rt-rt730"},
		Links: []catLink{info("Programmi e firmware Radtel", rtlDown)}, Notes: chirpNote("RT730") + " Piattaforma TIDRADIO TD-H8, 199 canali."},
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
}
