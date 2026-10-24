// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo dei modelli Baofeng e Bajeton mostrato nella galleria: per ogni radio
// il driver per canali e ponti (driver portati da CHIRP, oppure i due originali
// del programma: K67 Pro e Bajeton 8300), il tipo di aggiornamento del firmware
// disponibile e i collegamenti ai programmi e ai manuali dei costruttori.
//
// Fonti: https://www.baofengradio.com/pages/download,
// https://www.baofengradio.com/collections/catalog-products, elenco dei
// modelli di CHIRP (https://chirpmyradio.com), bajeton.com.

type catLink struct {
	Kind  string `json:"kind"` // cps | fw | manual | info
	Label string `json:"label"`
	URL   string `json:"url"`
}

type catItem struct {
	ID      string    `json:"id"`
	Brand   string    `json:"brand"`
	Name    string    `json:"name"`
	Aliases []string  `json:"aliases,omitempty"`
	Type    string    `json:"type"` // analog | gmrs | pmr | dmr | kids
	Drivers []string  `json:"drivers,omitempty"`
	FW      string    `json:"fw,omitempty"` // anc | bf | rh5 | kdhx | bj9000 | kdh9g | link
	Links   []catLink `json:"links,omitempty"`
	Notes   string    `json:"notes,omitempty"`
}

const (
	s3     = "https://baofeng.s3.amazonaws.com/"
	bjDown = "https://www.bajeton.com/it/download/"
	gd     = "https://drive.google.com/"
)

func cps(url string) catLink    { return catLink{"cps", "Programma del costruttore (CPS)", url} }
func manual(url string) catLink { return catLink{"manual", "Manuale", url} }

var catalog = []catItem{
	// ------------------------------------------------ Bajeton
	// collegamenti: pagina dei download di bajeton.com (cartelle Google Drive del costruttore)
	{ID: "bj-8300", Brand: "Bajeton", Name: "BJ-8300", Type: "analog", Drivers: []string{"baj"}, FW: "anc",
		Links: []catLink{{"cps", "Programma BT-11_CPS v1.0.8 (Google Drive Bajeton)", gd + "drive/folders/1l8XTP9j3STu5SsiWcWrcdblccSGgGGrn"},
			{"fw", "Firmware BJ8300_V0.12_20251112.anc e BinTe ANC Bootloader (Google Drive Bajeton)", gd + "drive/folders/1J4fUng8bC1uRyRvbDzZNlxdm9puzuANW"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Gestione originale di questo programma (file .dat del programma BT-11) e aggiornamento del firmware (.anc)."},
	{ID: "bj-7800", Brand: "Bajeton", Name: "BJ-7800", Aliases: []string{"BJ7800", "SHU H28Y Pro"}, Type: "analog", Drivers: []string{"bj-7800"}, FW: "anc",
		Links: []catLink{{"cps", "Programma BT-11_CPS v1.0.9 (Google Drive Bajeton)", gd + "drive/folders/1jj9HCZTsBrlM9O29fhuKQZVm8GYmRkmP"},
			{"fw", "Firmware BJ7800_CN_EN_V0.17_20251126.anc e BinTe ANC Bootloader (Google Drive Bajeton)", gd + "drive/folders/1VpGx-Actyg_fi-GzeMg1rXhaf5VlEHQm"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "960 canali in 15 zone da 64 (driver di CHIRP, scritto per il firmware V0.15P; i canali FM/AM/HF ausiliari non sono gestiti). Il firmware si aggiorna come quello della BJ-8300 (stesso programma «BinTe ANC Bootloader», file .anc)."},
	{ID: "bj-9000", Brand: "Bajeton", Name: "BJ-9000", Aliases: []string{"BJ9000 GPS APRS", "RT-950 Pro"}, Type: "analog", Drivers: []string{"bj-9000"}, FW: "bj9000",
		Links: []catLink{{"cps", "Programma BJ-9000_CPS v1.1.1 (Google Drive Bajeton)", gd + "drive/folders/1jJh1gEOniEV3Ud0ODhcH1JeKueAzq4Vm"},
			{"fw", "Firmware BJ9000_V0.27_260313 e programma BJ9000_EnUPDATE (Google Drive Bajeton)", gd + "drive/folders/1DF17RtchbKreT_QEAsQkON3qauP0gUp-"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "È la Radtel RT-950 Pro. Canali (960 in 15 zone) secondo il programma del costruttore BJ-9000_CPS v1.1.1 (si scrivono solo i canali; impostazioni e APRS con il programma del costruttore). " +
			"Firmware .Firmware con il protocollo del programma «BJ9000_EnUPDATE» (la radio controlla il modello prima di aggiornarsi)."},
	{ID: "bj-8200", Brand: "Bajeton", Name: "BJ-8200", Type: "analog", FW: "link",
		Links: []catLink{{"fw", "Programma e firmware — versione con Bluetooth", gd + "file/d/1ngQoWCWzLFUfmDSW1JLDcqKwQWxfz_Ph/view"},
			{"fw", "Programma e firmware — versione senza Bluetooth", gd + "file/d/14pLUetAJkNEECAuUJHYkgBM7EYvL2S2s/view"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Il costruttore avverte: il pacchetto per la versione con Bluetooth può danneggiare quella senza, e viceversa."},
	{ID: "bj-7700", Brand: "Bajeton", Name: "BJ-7700", Aliases: []string{"RT-910", "RT-910 BT"}, Type: "analog", Drivers: []string{"bj-7700", "bj-7700nbt"}, FW: "anc",
		Links: []catLink{{"fw", "Programma BJ-7700_EnCPS v1.0.2 e firmware — versione con Bluetooth", gd + "file/d/1ngQoWCWzLFUfmDSW1JLDcqKwQWxfz_Ph/view"},
			{"fw", "Programma BJ-7700_EnCPS v1.0.2 e firmware — versione senza Bluetooth", gd + "file/d/14pLUetAJkNEECAuUJHYkgBM7EYvL2S2s/view"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "È la Radtel RT-910 (il programma del costruttore è il «BT-RT910_CPS»): canali con i driver di CHIRP RT-910 BT / RT-910. " +
			"Firmware: solo per la versione con Bluetooth (BJ7700PRO_….anc, stesso programma «BinTe ANC Bootloader» della BJ-8300). " +
			"Il firmware della versione senza Bluetooth (.BTU) richiede un programma che il costruttore non pubblica. Il costruttore avverte: i pacchetti delle due versioni non sono intercambiabili."},
	{ID: "bj-98plus", Brand: "Bajeton", Name: "BJ-98 Plus", Aliases: []string{"BJ98 Plus", "UV-98 Plus"}, Type: "analog",
		Links: []catLink{{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Nessun programma pubblicato sulla pagina dei download del costruttore."},
	{ID: "bj-dm5000", Brand: "Bajeton", Name: "DM-5000", Type: "dmr", FW: "link",
		Links: []catLink{{"cps", "Programma di programmazione (Google Drive Bajeton)", gd + "file/d/1FaK6oXUZbwBtn5ooTPsXkmaAc5SDnxh0/view"},
			{"fw", "DMR R5000 Update Tool V2.29 (Google Drive Bajeton)", gd + "drive/folders/1HZ3NaUr8Fyr9vEUN2rSN6du5GyeWnTjP"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Radio DMR: programmazione (codeplug) e firmware solo con i programmi del costruttore."},
	{ID: "bj-m20d", Brand: "Bajeton", Name: "M20D", Type: "dmr",
		Links: []catLink{{"cps", "Programma di programmazione (Google Drive Bajeton)", gd + "drive/folders/1e-h9e8sjjnaAACswRFy9L3FpaPXA4KWg"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Programmazione solo con il programma del costruttore."},

	// ------------------------------------------------ Baofeng con display TFT / serie UV-17 Pro
	{ID: "uv-5rm", Brand: "Baofeng", Name: "UV-5RM", Type: "analog", Drivers: []string{"bf-uv5rm"}, FW: "bf",
		Links: []catLink{cps(s3 + "Baofeng_UV-5RM_CPS_20231124.zip")},
		Notes: "Firmware .BF (caricatore KDH): attenzione al chip radio (BK4819 / AT1846S), le versioni non sono intercambiabili."},
	{ID: "uv-5rh", Brand: "Baofeng", Name: "UV-5RH", Type: "analog", Drivers: []string{"bf-uv5rh"}, FW: "bf",
		Notes: "Firmware .BF (caricatore KDH): attenzione al chip radio (BK4819 / AT1846S / FD6818)."},
	{ID: "uv-5rm-plus", Brand: "Baofeng", Name: "UV-5RM Plus", Type: "analog", Drivers: []string{"bf-uv5rmplus"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-5RM_Plus_CPS_20240327.zip"), manual(s3 + "Baofeng_UV-5RM_PLUS_K5PLUS_User_Manual_20250427.pdf")}},
	{ID: "k5-plus", Brand: "Baofeng", Name: "K5 Plus", Type: "analog", Drivers: []string{"bf-k5plus"},
		Links: []catLink{cps(s3 + "BAOFENG_K5_Plus_CPS_20240902.zip"), manual(s3 + "Baofeng_UV-5RM_PLUS_K5PLUS_User_Manual_20250427.pdf")}},
	{ID: "uv-17pro", Brand: "Baofeng", Name: "UV-17 Pro", Aliases: []string{"UV-17Pro"}, Type: "analog", Drivers: []string{"bf-uv17pro"}},
	{ID: "uv-17pro-gps", Brand: "Baofeng", Name: "UV-17 Pro GPS", Type: "analog", Drivers: []string{"bf-uv17progps"}},
	{ID: "uv-17r-plus", Brand: "Baofeng", Name: "UV-17R Plus", Type: "analog", Drivers: []string{"bf-uv17rplus"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-17R_CPS_20230801.zip"), manual(s3 + "BAOFENG_UV_17R_PLUS_User_Manual_EN_20240306.pdf")}},
	{ID: "uv-21pro-v2", Brand: "Baofeng", Name: "UV-21 Pro V2", Type: "analog", Drivers: []string{"bf-uv21prov2"}},
	{ID: "uv-21pro-gps", Brand: "Baofeng", Name: "UV-21 Pro GPS", Type: "analog", Drivers: []string{"bf-uv21progps"}},
	{ID: "uv-25", Brand: "Baofeng", Name: "UV-25", Type: "analog", Drivers: []string{"bf-uv25"},
		Links: []catLink{cps(s3 + "Baofeng_UV-25_CPS.zip"), manual(s3 + "Baofeng_UV-25_User_Manual_EU_Version_240524.pdf")}},
	{ID: "uv-28plus", Brand: "Baofeng", Name: "UV-28 Plus", Type: "analog", Drivers: []string{"bf-uv28plus"},
		Links: []catLink{cps(s3 + "Baofeng_UV-28_Plus_CPS.zip"), manual(s3 + "Baofeng_UV-28_Plus_User_Manual.pdf")}},
	{ID: "uv-32", Brand: "Baofeng", Name: "UV-32", Type: "analog", Drivers: []string{"bf-uv32"},
		Links: []catLink{cps(s3 + "Baofeng_UV-32_CPS_v1.4.7.zip"), manual(s3 + "Baofeng_UV-32_User_Manual.pdf")}},
	{ID: "bf-f8hp-pro", Brand: "Baofeng", Name: "BF-F8HP Pro", Type: "analog", Drivers: []string{"bf-f8hppro"}},
	{ID: "k6", Brand: "Baofeng", Name: "K6 / K63", Aliases: []string{"K6", "K63"}, Type: "analog", Drivers: []string{"bf-k6"},
		Links: []catLink{cps(s3 + "BAOFENG_K6_CPS_1.1.2_20240913.zip"), manual(s3 + "Baofeng_K6_User_Manual_2024-08-02.pdf")}},
	{ID: "uv-5r-mini", Brand: "Baofeng", Name: "UV-5R Mini / MINI-5", Aliases: []string{"MINI-5", "UV-5R Mini"}, Type: "analog", Drivers: []string{"bf-uv5rmini"},
		Links: []catLink{cps(s3 + "Baofeng_UV_5R_Mini_CPS_V1.0.7.zip"), manual(s3 + "Baofeng_UV-5R_MINI_User_Manual.pdf")}},
	{ID: "uv-5g-mini", Brand: "Baofeng", Name: "UV-5G Mini", Aliases: []string{"MINI-5G"}, Type: "gmrs", Drivers: []string{"bf-uv5gmini"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5G_Mini_CPS_v1.0.3_202508.zip"), manual(s3 + "Baofeng_UV-5G_Mini_User_Manual_202508.pdf")}},
	{ID: "gm-21", Brand: "Baofeng", Name: "GM-21", Type: "gmrs", Drivers: []string{"bf-gm21"}},
	{ID: "gm-5rh", Brand: "Baofeng", Name: "GM-5RH", Type: "gmrs", Drivers: []string{"bf-gm5rh"}},
	{ID: "uv-5g-plus", Brand: "Baofeng", Name: "UV-5G Plus", Type: "gmrs", Drivers: []string{"bf-uv5gplus"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-5G_Plus_CPS_20240910.zip"), manual(s3 + "BAOFENG_UV_5G_PLUS_User_Manual_EN_20231012.pdf")}},
	{ID: "uv-17", Brand: "Baofeng", Name: "UV-17", Type: "analog", Drivers: []string{"bf-uv17"}},
	{ID: "uv-13pro", Brand: "Baofeng", Name: "UV-13 Pro / TP-8Plus", Aliases: []string{"UV-13Pro", "TP-8Plus"}, Type: "analog", Drivers: []string{"bf-uv13pro"}},
	{ID: "5rh-pro", Brand: "Baofeng", Name: "5RH Pro / K67 Pro", Aliases: []string{"5RH-PRO", "K67", "K67 Pro"}, Type: "analog", Drivers: []string{"k67"}, FW: "rh5",
		Notes: "Gestione originale di questo programma (file .xlc del CPS «5RH_PRO»). Firmware .dat «BF_5RH»: la versione dell'elettronica (V1/V2) deve corrispondere."},

	// ------------------------------------------------ Baofeng serie UV-5R / UV-82 / GT
	{ID: "uv-5r", Brand: "Baofeng", Name: "UV-5R", Aliases: []string{"UV-5RA", "UV-5RE", "UV-5R Plus", "UV-5R+", "UV-5RQ", "UV-5R EX", "UV-5RB", "GT-3", "GT-3 MkII"}, Type: "analog", Drivers: []string{"bf-uv5r"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5_Series_UV-5R_UV-5RA_UV-5RE_UV-5R_GT-3Series_GT-3_GT-3TP_Programming_Software_update.rar"), manual(s3 + "Baofeng_UV-5R_UV-5R-Plus_GT-3_Users_Manual.pdf")}},
	{ID: "uv-5r-pro", Brand: "Baofeng", Name: "UV-5R+Pro / UV-5XP", Aliases: []string{"UV-5R Pro", "UV-5XP"}, Type: "analog", Drivers: []string{"bf-uv5r", "bf-f8hp"},
		Notes: "Elettronica della serie UV-5R: scegli UV-5R oppure BF-F8HP (8 W) e leggi prima la radio; se il modello non è quello giusto la radio non risponde e non viene modificata."},
	{ID: "bf-f8hp", Brand: "Baofeng", Name: "BF-F8HP / UV-5RTP", Aliases: []string{"UV-5RTP", "GT-3TP", "UV-5R 8W", "BF-F8+"}, Type: "analog", Drivers: []string{"bf-f8hp"},
		Links: []catLink{cps(s3 + "Baofeng_UV5RTP_CPS_ENG.zip")}},
	{ID: "uv-5rx3", Brand: "Baofeng", Name: "UV-5RX3", Aliases: []string{"BTECH UV-5X3"}, Type: "analog", Drivers: []string{"bf-uv5rx3-btech", "bf-uv5rx3"},
		Links: []catLink{cps("https://baofeng-uv-5rx3.s3.amazonaws.com/UV-5RX3%20Programming%20Software%20v2.7.zip")},
		Notes: "Due versioni del firmware: tribanda (protocollo BTECH) o tipo UV-5R. Se una non risponde prova l'altra."},
	{ID: "uv-82", Brand: "Baofeng", Name: "UV-82 / UV-82L", Aliases: []string{"UV-82", "UV-82L", "UV-82C"}, Type: "analog", Drivers: []string{"bf-uv82"},
		Links: []catLink{cps("https://cdn.shopify.com/s/files/1/0251/9475/8224/files/UV_82_CPS.rar?v=1680765756"), manual("http://s3.image.ro.s3.amazonaws.com/download/UV82L_Manual.pdf")}},
	{ID: "uv-82hp", Brand: "Baofeng", Name: "UV-82HP", Type: "analog", Drivers: []string{"bf-uv82hp"}},
	{ID: "uv-82x3", Brand: "Baofeng", Name: "UV-82X3", Type: "analog", Drivers: []string{"bf-uv82x3"},
		Links: []catLink{cps("https://radioddity.s3.amazonaws.com/Baofeng%20UV-82X3%20Programming%20Software.zip")}},
	{ID: "uv-82wp", Brand: "Baofeng", Name: "UV-82WP", Type: "analog", Drivers: []string{"bf-uv82wp"}},
	{ID: "gt-3wp", Brand: "Baofeng", Name: "GT-3WP", Type: "analog", Drivers: []string{"bf-gt3wp"}},
	{ID: "uv-6", Brand: "Baofeng", Name: "UV-6 / UV-7", Type: "analog", Drivers: []string{"bf-uv6"}},
	{ID: "uv-6r", Brand: "Baofeng", Name: "UV-6R", Type: "analog", Drivers: []string{"bf-uv6r"},
		Links: []catLink{cps("http://s3.image.ro.s3.amazonaws.com/download/UV6R_Setup_EN_V1.10.rar"), manual("http://s3.image.ro.s3.amazonaws.com/download/UV-6R-User-Manual.pdf")}},
	{ID: "uv-5g", Brand: "Baofeng", Name: "UV-5G (UV-5X)", Aliases: []string{"UV-5X"}, Type: "gmrs", Drivers: []string{"bf-uv5g"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5X_Programming_Software_V2.0_20210308.zip")}},
	{ID: "uv-5g-pro", Brand: "Baofeng", Name: "UV-5G Pro", Type: "gmrs", Drivers: []string{"bf-uv5gpro"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5G_PRO_CPS.zip"), manual(s3 + "Baofeng_UV-5G_PRO_User_Manual.pdf")}},
	{ID: "f-11", Brand: "Baofeng", Name: "F-11", Type: "analog", Drivers: []string{"bf-f11"}},
	{ID: "gt-5r", Brand: "Baofeng", Name: "GT-5R", Aliases: []string{"GT-5R Pro"}, Type: "analog", Drivers: []string{"bf-gt5r"}},

	// ------------------------------------------------ Baofeng impermeabili / UV-9R
	{ID: "bf-a58", Brand: "Baofeng", Name: "BF-A58 / UV-9R Pro", Aliases: []string{"UV-9R Pro"}, Type: "analog", Drivers: []string{"bf-a58"}},
	{ID: "bf-a58s", Brand: "Baofeng", Name: "BF-A58S / UV-82III", Aliases: []string{"UV-82III"}, Type: "analog", Drivers: []string{"bf-a58s"}},
	{ID: "uv-s9x3", Brand: "Baofeng", Name: "UV-S9X3", Type: "analog", Drivers: []string{"bf-uvs9x3"}},
	{ID: "uv-5rx", Brand: "Baofeng", Name: "UV-5RX (5RX)", Aliases: []string{"5RX"}, Type: "analog", Drivers: []string{"bf-5rx"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5RX_CPS_20231124.zip")}},
	{ID: "uv-9r", Brand: "Baofeng", Name: "UV-9R / UV-9R Plus", Aliases: []string{"UV-9R Plus"}, Type: "analog", Drivers: []string{"bf-uv9r"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-9R_CPS_20231019.zip"), manual(s3 + "BAOFENG_UV_9R_User_Manual_20231019_EN.pdf")}},
	{ID: "uv-9g", Brand: "Baofeng", Name: "UV-9G", Type: "gmrs", Drivers: []string{"bf-uv9g"}, FW: "kdh9g",
		Links: []catLink{cps(s3 + "BAOFENG_UV-9G_Programming_Software_V2.5_20210916.rar"),
			{"fw", "Aggiornamento del costruttore (trasmissione sul canale 15)", s3 + "Baofeng_UV-9G_Update_Transmit_on_Channel_15_202311.zip"},
			manual(s3 + "BAOFENG_UV-9G_GMRS_User_Manual_20210806.pdf")},
		Notes: "Firmware .kdh (T52GMRS_…) con il protocollo del programma del costruttore «Bootloader_BaudRate38400». Attenzione: il caricatore non controlla il modello."},

	// ------------------------------------------------ Baofeng PMR / FRS / piccole
	{ID: "bf-888s", Brand: "Baofeng", Name: "BF-888S", Aliases: []string{"BF-888", "H-777"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Links: []catLink{cps(s3 + "Baofeng_BF-888S_CPS_NEW_20240301.zip"), manual("http://s3.image.ro.s3.amazonaws.com/download/BaoFeng-BF888s-Manual.pdf")}},
	{ID: "bf-m4", Brand: "Baofeng", Name: "BF-M4", Type: "pmr", Drivers: []string{"bf-m4"}},
	{ID: "mp31", Brand: "Baofeng", Name: "MP31", Type: "gmrs", Drivers: []string{"bf-mp31"}},
	{ID: "bf-1901", Brand: "Baofeng", Name: "BF-1901", Type: "pmr", Drivers: []string{"bf-1901"}},
	{ID: "bf-1904", Brand: "Baofeng", Name: "BF-1904", Type: "pmr", Drivers: []string{"bf-1904"}},
	{ID: "bf-1909", Brand: "Baofeng", Name: "BF-1909", Type: "pmr", Drivers: []string{"bf-1909"}},
	{ID: "uv-b5", Brand: "Baofeng", Name: "UV-B5 / UV-B6", Aliases: []string{"UV-B6"}, Type: "analog", Drivers: []string{"bf-uvb5"}},
	{ID: "uv-3r", Brand: "Baofeng", Name: "UV-3R", Type: "analog", Drivers: []string{"bf-uv3r"}},
	{ID: "bf-t1", Brand: "Baofeng", Name: "BF-T1", Type: "pmr", Drivers: []string{"bf-t1"}},
	{ID: "bf-t8", Brand: "Baofeng", Name: "BF-T8 / BF-U9", Aliases: []string{"BF-U9"}, Type: "pmr", Drivers: []string{"bf-t8"}},
	{ID: "bf-v8a", Brand: "Baofeng", Name: "BF-V8A", Type: "pmr", Drivers: []string{"bf-v8a"}},
	{ID: "bf-t20frs", Brand: "Baofeng", Name: "BF-T20FRS", Type: "pmr", Drivers: []string{"bf-t20frs"}},
	{ID: "bf-t20", Brand: "Baofeng", Name: "BF-T20", Type: "pmr", Drivers: []string{"bf-t20"}},
	{ID: "w31e", Brand: "Baofeng", Name: "W31E", Type: "pmr", Drivers: []string{"bf-w31e"}},

	// ------------------------------------------------ Baofeng solo con il programma del costruttore
	{ID: "uv-25-plus", Brand: "Baofeng", Name: "UV-25 Plus", Type: "analog", Drivers: []string{"bf-uv25plus"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_UV_25_Plus_CPS_20250227.zip"), {"fw", "Aggiornamento del firmware (guida e programma)", s3 + "Baofeng_UV-25_Plus_Firware_Update_Guide_20250526.zip"},
			manual(s3 + "Baofeng_UV-25_PLUS_User_Manual_EN_IT_DE_ES_FR_20250227.pdf")},
		Notes: "Canali secondo il programma del costruttore «BF-UVs CPS». Firmware .kdhx con il protocollo del programma «Baofeng Bootloader Pro»."},
	{ID: "uv-5rm-plus-gps", Brand: "Baofeng", Name: "UV-5RM Plus GPS", Type: "analog", Drivers: []string{"k67"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5RM_Plus_GPS_CPS_V1.17.zip")},
		Notes: "Il programma ufficiale di questa radio è il CPS «5RH PRO» (lo stesso della 5RH Pro / K67 Pro): si gestisce dalla scheda 5RH Pro / K67 Pro (file .xlc)."},
	{ID: "uv-17r", Brand: "Baofeng", Name: "UV-17R", Type: "analog", Drivers: []string{"bf-uv17r"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-17R_CPS_20230801.zip"), manual(s3 + "BAOFENG_UV_17R_User_Manual_20230815_EN.pdf")}},
	{ID: "uv-21r", Brand: "Baofeng", Name: "UV-21R", Type: "analog", Drivers: []string{"bf-uv21r"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-21R_CPS_20240902.zip"), manual(s3 + "Baofeng_UV-21R_User_Manual_EN_20240902.pdf")}},
	{ID: "uv-21l", Brand: "Baofeng", Name: "UV-21L", Type: "analog", Drivers: []string{"bf-uv21l"}, Links: []catLink{cps(s3 + "Baofeng_UV-21L_CPS_20231130.zip")}},
	{ID: "uv-21", Brand: "Baofeng", Name: "UV-21", Type: "analog"},
	{ID: "uv-29pro", Brand: "Baofeng", Name: "UV-29 Pro", Type: "analog", Drivers: []string{"bf-uv29pro"},
		Links: []catLink{cps(s3 + "Baofeng_UV-29_PRO_SW-9_CPS_v1.6.1.zip"), manual(s3 + "Baofeng_UV-29_Pro_User_Manual_EN_DE_FR_IT_ES_2026417.pdf")}},
	{ID: "uv-7b", Brand: "Baofeng", Name: "UV-7B", Type: "analog", Drivers: []string{"bf-uv7b"},
		Links: []catLink{cps(s3 + "Baofeng_UV-7B_Programming_Software_20260205.zip"), manual(s3 + "Baofeng_UV-7B_User_Manual_EN_20251024.pdf")}},
	{ID: "gm-15pro", Brand: "Baofeng", Name: "GM-15 Pro", Type: "gmrs", Drivers: []string{"bf-gm15pro"},
		Links: []catLink{cps(s3 + "BAOFENG_GM-15_PRO_CPS_20240902.zip"), manual(s3 + "Baofeng_GM-15_Pro_User_Manual_EN_20240902.pdf")}},
	{ID: "ar-152", Brand: "Baofeng", Name: "AR-152 / AR-152 Pro", Type: "analog", Drivers: []string{"bf-ar152pro"},
		Links: []catLink{cps(s3 + "Baofeng_AR-152_PRO_programming_software_v1.4.6.zip")}},
	{ID: "uv-16", Brand: "Baofeng", Name: "UV-16 / UV-16H Pro", Aliases: []string{"UV-16H Pro"}, Type: "analog"},
	{ID: "p52uv", Brand: "Baofeng", Name: "P52UV", Type: "analog"},
	{ID: "gmrs-9r", Brand: "Baofeng", Name: "GMRS-9R", Type: "gmrs"},
	{ID: "g11s", Brand: "Baofeng", Name: "G11S", Type: "gmrs"},
	{ID: "5rham", Brand: "Baofeng", Name: "5RHam", Type: "analog"},
	{ID: "bf-9700", Brand: "Baofeng", Name: "BF-9700", Type: "analog",
		Notes: "Il programma del costruttore (Wp970) è un eseguibile nativo senza protocollo documentato e la radio non è in CHIRP: usa il programma del costruttore.",
		Links: []catLink{cps("https://s3.amazonaws.com/s3.image.ro/download/BF9700Setup_EN.rar"), manual("http://s3.image.ro.s3.amazonaws.com/download/BF-9700USERMANUAL.pdf")}},
	{ID: "uv-9s", Brand: "Baofeng", Name: "UV-9S", Type: "analog"},
	{ID: "uv-8d", Brand: "Baofeng", Name: "UV-8D", Type: "analog"},
	{ID: "uv-e70", Brand: "Baofeng", Name: "UV-E70", Type: "analog"},
	{ID: "bf-uv8", Brand: "Baofeng", Name: "BF-UV8", Type: "analog"},
	{ID: "bf-r3", Brand: "Baofeng", Name: "BF-R3", Type: "analog"},
	{ID: "h-series", Brand: "Baofeng", Name: "Serie H (H-31, H-32, BF-H2/H5/H6)", Aliases: []string{"H-31", "H-32", "BF-H2", "BF-H5", "BF-H6"}, Type: "analog", Drivers: []string{"bf-hseries"},
		Links: []catLink{cps(s3 + "BAOFENG%20H-Series%20Programming%20Software.rar")}},
	{ID: "d-series", Brand: "Baofeng", Name: "Serie BF-D", Type: "analog", Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")},
		Notes: "Il programma del costruttore (app «intercom-configuration») usa un protocollo proprio, non documentato e non presente in CHIRP: senza una radio per provarlo non è gestito qui, usa il programma del costruttore."},
	{ID: "bf-c2", Brand: "Baofeng", Name: "BF-C2 (BF-V2)", Aliases: []string{"BF-V2"}, Type: "pmr", Drivers: []string{"bf-c2"},
		Links: []catLink{cps(s3 + "BAOFENG_BF-V2_CPS_20231019.zip"), manual(s3 + "BAOFENG_BF_V2_User_Manual_20231019_EN.pdf")}},
	{ID: "bf-v", Brand: "Baofeng", Name: "BF-V1 / V3 / V4 / V5 / V6A", Aliases: []string{"BF-V1", "BF-V3", "BF-V4", "BF-V5", "BF-V6A"}, Type: "pmr"},
	{ID: "bf-88st", Brand: "Baofeng", Name: "BF-88ST / BF-88ST Pro", Aliases: []string{"BF-88ST Pro"}, Type: "pmr", Drivers: []string{"bf-88st"},
		Links: []catLink{cps(s3 + "Baofeng%20BF-88ST%20Programming%20Software_20200722.zip"), manual(s3 + "Baofeng%20BF-88ST%20User%20Manual_20200723.pdf")}},
	{ID: "bf-88a", Brand: "Baofeng", Name: "BF-88A / BF-88E", Aliases: []string{"BF-88E"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Links: []catLink{cps("https://bf-88a88e.s3.amazonaws.com/Programming%20Software%20%26%20Chirp%20BF-88A%20%26%2088E.zip")},
		Notes: "Il pacchetto del costruttore indica di programmarla con CHIRP come «Baofeng BF-888»: si usa lo stesso driver della BF-888S."},
	{ID: "t22", Brand: "Baofeng", Name: "T22", Type: "pmr", Drivers: []string{"bf-t22"},
		Links: []catLink{cps(s3 + "Baofeng_T22_Programming_Software_PMR.zip"), manual(s3 + "Baofeng_T22_User_Manual_PMR_20250122.pdf")}},
	{ID: "bf-33c", Brand: "Baofeng", Name: "BF-33C", Type: "pmr", Links: []catLink{manual(s3 + "Baofeng_BF-33C_User_Manual_EU_Version.pdf")}},
	{ID: "v1d", Brand: "Baofeng", Name: "V1D", Type: "pmr", Links: []catLink{cps(s3 + "Baofeng_V1D_CPS.zip"), manual(s3 + "Baofeng_V1D_User_Manual.pdf")},
		Notes: "Il programma del costruttore (app «intercom-configuration») usa un protocollo proprio, non documentato e non presente in CHIRP: senza una radio per provarlo non è gestito qui, usa il programma del costruttore."},
	{ID: "bf-pmr-old", Brand: "Baofeng", Name: "BF-528 / 600S / 666S / 700S / 777S / 800S", Aliases: []string{"BF-528", "BF-600S", "BF-666S", "BF-700S", "BF-777S", "BF-800S"}, Type: "pmr"},
	{ID: "bf-t-series", Brand: "Baofeng", Name: "BF-T6 / T7 / T11 / T12 / T15 / T16 / T26 / T51 / T57", Aliases: []string{"BF-T6", "BF-T7", "BF-T11", "BF-T12", "BF-T15", "BF-T16", "BF-T26", "BF-T51", "BF-T57"}, Type: "pmr"},
	{ID: "bf-other", Brand: "Baofeng", Name: "BF-358 / A68 / B5 / B6 / E51 / S88", Aliases: []string{"BF-358", "BF-A68", "BF-B5", "BF-B6", "BF-E51", "BF-S88"}, Type: "pmr"},
	{ID: "ar888s", Brand: "Baofeng", Name: "AR888S / TH888S", Aliases: []string{"AR888S", "TH888S"}, Type: "pmr"},
	{ID: "bf-t3", Brand: "Baofeng", Name: "BF-T3 (per bambini)", Type: "kids"},
	{ID: "dm-1701", Brand: "Baofeng", Name: "DM-1701", Type: "dmr"},
	{ID: "dm-1801pro", Brand: "Baofeng", Name: "DM-1801 Pro", Type: "dmr"},
	{ID: "dm-1703", Brand: "Baofeng", Name: "DM-1703", Type: "dmr"},
	{ID: "dr-1801uv", Brand: "Baofeng", Name: "DR-1801UV", Type: "dmr"},
	{ID: "rd-5r", Brand: "Baofeng", Name: "RD-5R", Type: "dmr"},
	{ID: "dm-v1", Brand: "Baofeng", Name: "DM-V1", Type: "dmr"},
}

// catalogDefault: nota per i modelli senza driver
const catalogNoDriver = "Protocollo non documentato o radio digitale: la programmazione si fa con il programma del costruttore."
