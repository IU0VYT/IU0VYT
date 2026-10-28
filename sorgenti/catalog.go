// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "strings"

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
	// Models: modelli distinti venduti con nomi diversi e gestiti allo stesso modo:
	// nella galleria e negli elenchi diventano schede separate (vedi expandCatalog)
	Models []string `json:"-"`
}

const (
	s3     = "https://baofeng.s3.amazonaws.com/"
	bjDown = "https://www.bajeton.com/it/download/"
	gd     = "https://drive.google.com/"
	rdDown = "https://www.radioddity.com/pages/baofeng-download"
	bfDown = "https://www.baofengradio.com/pages/download"
)

// radio PMR446 digitali della serie D (programma «BF Series CPS»)
const notesSerieD = "Serie D (PMR446 con canali analogici e digitali criptati). Protocollo e memoria ricavati dal programma del costruttore " +
	"«BF Series CPS» (app «intercom-configuration»): la radio comunica il proprio modello e il programma lo controlla."

func cps(url string) catLink    { return catLink{"cps", "Programma del costruttore (CPS)", url} }
func manual(url string) catLink { return catLink{"manual", "Manuale", url} }

var catalog = []catItem{
	// ------------------------------------------------ Bajeton
	// collegamenti: pagina dei download di bajeton.com (cartelle Google Drive del costruttore)
	{ID: "bj-8300", Brand: "Bajeton", Name: "BJ-8300", Type: "analog", Drivers: []string{"baj"}, FW: "anc",
		Links: []catLink{{"cps", "Programma BT-11_CPS v1.0.8 (Google Drive Bajeton)", gd + "drive/folders/1l8XTP9j3STu5SsiWcWrcdblccSGgGGrn"},
			{"cps", "Programma BT-11_CPS v1.0.9 (Google Drive Bajeton)", gd + "drive/folders/1jj9HCZTsBrlM9O29fhuKQZVm8GYmRkmP"},
			{"fw", "Firmware BJ8300_V0.12_20251112.anc e BinTe ANC Bootloader (Google Drive Bajeton)", gd + "drive/folders/1J4fUng8bC1uRyRvbDzZNlxdm9puzuANW"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Gestione originale di questo programma (file .dat del programma BT-11_CPS, versioni 1.0.8 e 1.0.9: le due versioni salvano i file in formati diversi e non aprono quelli dell'altra, qui si aprono entrambi e si sceglie il formato in cui salvare), " +
			"impostazioni della radio (finestra «Optional Features» del programma) e aggiornamento del firmware (.anc)."},
	{ID: "bj-7800", Brand: "Bajeton", Name: "BJ-7800", Aliases: []string{"BJ7800", "SHU H28Y Pro"}, Type: "analog", Drivers: []string{"bj-7800"}, FW: "anc",
		Links: []catLink{{"cps", "Programma BT-11_CPS v1.0.9 (Google Drive Bajeton)", gd + "drive/folders/1jj9HCZTsBrlM9O29fhuKQZVm8GYmRkmP"},
			{"fw", "Firmware BJ7800_CN_EN_V0.17_20251126.anc e BinTe ANC Bootloader (Google Drive Bajeton)", gd + "drive/folders/1VpGx-Actyg_fi-GzeMg1rXhaf5VlEHQm"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "960 canali in 15 zone da 64 (driver di CHIRP, scritto per il firmware V0.15P; i canali FM/AM/HF ausiliari non sono gestiti). Il firmware si aggiorna come quello della BJ-8300 (stesso programma «BinTe ANC Bootloader», file .anc)."},
	{ID: "bj-9000", Brand: "Bajeton", Name: "BJ-9000", Aliases: []string{"BJ9000 GPS APRS", "RT-950 Pro"}, Type: "analog", Drivers: []string{"bj-9000"}, FW: "bj9000",
		Links: []catLink{{"cps", "Programma BJ-9000_CPS v1.1.1 (Google Drive Bajeton)", gd + "drive/folders/1jJh1gEOniEV3Ud0ODhcH1JeKueAzq4Vm"},
			{"fw", "Firmware BJ9000_V0.27_260313 e programma BJ9000_EnUPDATE (Google Drive Bajeton)", gd + "drive/folders/1DF17RtchbKreT_QEAsQkON3qauP0gUp-"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "È la Radtel RT-950 Pro. Canali (960 in 15 zone) e impostazioni generali secondo il programma del costruttore BJ-9000_CPS v1.1.1 (VFO, DTMF e APRS con il programma del costruttore). " +
			"Firmware .Firmware con il protocollo del programma «BJ9000_EnUPDATE» (la radio controlla il modello prima di aggiornarsi)."},
	{ID: "bj-8200", Brand: "Bajeton", Name: "BJ-8200", Type: "analog", Drivers: []string{"bj-7700", "bj-7700nbt"}, FW: "anc",
		Links: []catLink{{"fw", "Programma BJ-7700_EnCPS e firmware — versione con Bluetooth (stesso pacchetto della BJ-7700)", gd + "file/d/1ngQoWCWzLFUfmDSW1JLDcqKwQWxfz_Ph/view"},
			{"fw", "Programma BJ-7700_EnCPS e firmware — versione senza Bluetooth (stesso pacchetto della BJ-7700)", gd + "file/d/14pLUetAJkNEECAuUJHYkgBM7EYvL2S2s/view"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Per la BJ-8200 il costruttore pubblica esattamente gli stessi pacchetti della BJ-7700 (programma «BT-RT910_CPS», firmware BJ7700PRO_….anc): " +
			"si usano gli stessi driver (Radtel RT-910 / RT-910 BT). Se la radio si identifica in modo diverso il programma si ferma senza scrivere. " +
			"Il costruttore avverte: il pacchetto per la versione con Bluetooth può danneggiare quella senza, e viceversa."},
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
	{ID: "bj-dm5000", Brand: "Bajeton", Name: "DM-5000", Type: "dmr", Drivers: []string{"bj-dm5000"}, FW: "link",
		Links: []catLink{{"cps", "Programma di programmazione (Google Drive Bajeton)", gd + "file/d/1FaK6oXUZbwBtn5ooTPsXkmaAc5SDnxh0/view"},
			{"fw", "DMR R5000 Update Tool V2.29 (Google Drive Bajeton)", gd + "drive/folders/1HZ3NaUr8Fyr9vEUN2rSN6du5GyeWnTjP"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Canali analogici e DMR con il protocollo e la memoria ricavati dal programma del costruttore «BAJETON DM-5000 CPS» (versione 250726): " +
			"frequenze, nomi, potenza, scansione e toni; contatti, gruppi, zone e impostazioni restano come sono (un canale nuovo entra nella zona 1). " +
			"Il firmware si aggiorna con il programma del costruttore (attenzione: alcuni antivirus segnalano come pericolosi i pacchetti «DMR R5000 Update Tool»)."},
	{ID: "bj-m20d", Brand: "Bajeton", Name: "M20D", Type: "dmr",
		Links: []catLink{{"cps", "Programma di programmazione (Google Drive Bajeton)", gd + "drive/folders/1e-h9e8sjjnaAACswRFy9L3FpaPXA4KWg"},
			{"info", "Pagina dei download Bajeton", bjDown}},
		Notes: "Radio DMR veicolare (piattaforma «HD 520D»): programmazione con il programma del costruttore «520D» (v1.13) e firmware con " +
			"«Firmware Update-HD-V3000S-HD520D» (v1.2.5, gennaio 2026). Per l'aggiornamento il costruttore indica di accendere la radio " +
			"tenendo premuti A/B e CH+."},

	// ------------------------------------------------ Baofeng con display TFT / serie UV-17 Pro
	{ID: "uv-5rm", Brand: "Baofeng", Name: "UV-5RM", Type: "analog", Drivers: []string{"bf-uv5rm"}, FW: "bf",
		Links: []catLink{cps(s3 + "Baofeng_UV-5RM_CPS_20231124.zip")},
		Notes: "Firmware .BF (caricatore KDH): attenzione al chip radio (BK4819 / AT1846S), le versioni non sono intercambiabili."},
	{ID: "uv-5rh", Brand: "Baofeng", Name: "UV-5RH", Type: "analog", Drivers: []string{"bf-uv5rh"}, FW: "bf",
		Notes: "Firmware .BF (caricatore KDH): attenzione al chip radio (BK4819 / AT1846S / FD6818)."},
	{ID: "uv-5rm-plus", Brand: "Baofeng", Name: "UV-5RM Plus", Type: "analog", Drivers: []string{"bf-uv5rmplus"}, FW: "bf",
		Links: []catLink{cps(s3 + "BAOFENG_UV-5RM_Plus_CPS_20240327.zip"), manual(s3 + "Baofeng_UV-5RM_PLUS_K5PLUS_User_Manual_20250427.pdf")}},
	{ID: "k5-plus", Brand: "Baofeng", Name: "K5 Plus", Aliases: []string{"UV-K5 Plus", "BF-K5 Plus"}, Type: "analog", Drivers: []string{"bf-k5plus"}, FW: "bf",
		Links: []catLink{cps(s3 + "BAOFENG_K5_Plus_CPS_20240902.zip"), manual(s3 + "Baofeng_UV-5RM_PLUS_K5PLUS_User_Manual_20250427.pdf")}},
	{ID: "uv-17pro", Brand: "Baofeng", Name: "UV-17 Pro", Aliases: []string{"UV-17Pro"}, Type: "analog", Drivers: []string{"bf-uv17pro"}},
	{ID: "uv-17pro-gps", Brand: "Baofeng", Name: "UV-17 Pro GPS", Type: "analog", Drivers: []string{"bf-uv17progps"}},
	{ID: "uv-17r-plus", Brand: "Baofeng", Name: "UV-17R Plus", Type: "analog", Drivers: []string{"bf-uv17rplus"}, FW: "bf",
		Links: []catLink{cps(s3 + "BAOFENG_UV-17R_CPS_20230801.zip"), manual(s3 + "BAOFENG_UV_17R_PLUS_User_Manual_EN_20240306.pdf")}},
	{ID: "uv-21pro-v2", Brand: "Baofeng", Name: "UV-21 Pro V2", Type: "analog", Drivers: []string{"bf-uv21prov2"}},
	{ID: "uv-21pro-gps", Brand: "Baofeng", Name: "UV-21 Pro GPS", Type: "analog", Drivers: []string{"bf-uv21progps"}},
	{ID: "uv-25", Brand: "Baofeng", Name: "UV-25", Type: "analog", Drivers: []string{"bf-uv25"},
		Links: []catLink{cps(s3 + "Baofeng_UV-25_CPS.zip"), manual(s3 + "Baofeng_UV-25_User_Manual_EU_Version_240524.pdf")}},
	{ID: "uv-28plus", Brand: "Baofeng", Name: "UV-28 Plus", Type: "analog", Drivers: []string{"bf-uv28plus"},
		Links: []catLink{cps(s3 + "Baofeng_UV-28_Plus_CPS.zip"), manual(s3 + "Baofeng_UV-28_Plus_User_Manual.pdf")}},
	{ID: "uv-28pro", Brand: "Baofeng", Name: "UV-28 Pro", Aliases: []string{"UV-28 Pro GPS"}, Type: "analog", Drivers: []string{"bf-uv28pro"}, FW: "kdhx",
		Links: []catLink{{"cps", "Programma «BF-UVs CPS» (pacchetto «Baofeng UV-28 Pro CPS», pagina dei download Baofeng)", bfDown}},
		Notes: "Protocollo e memoria del programma del costruttore «BF-UVs CPS» (la radio si identifica come UV28PRO), gli stessi della UV-17 Pro GPS di CHIRP."},
	{ID: "uv-32", Brand: "Baofeng", Name: "UV-32", Aliases: []string{"UV-32 Pro", "UV-32 GPS"}, Type: "analog", Drivers: []string{"bf-uv32"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_UV-32_CPS_v1.4.7.zip"), manual(s3 + "Baofeng_UV-32_User_Manual.pdf")}},
	{ID: "bf-f8hp-pro", Brand: "Baofeng", Name: "BF-F8HP Pro", Type: "analog", Drivers: []string{"bf-f8hppro"}},
	{ID: "k6", Brand: "Baofeng", Name: "K6 / K63", Aliases: []string{"K6", "K63"}, Models: []string{"K6", "K63"}, Type: "analog", Drivers: []string{"bf-k6"},
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
	{ID: "uv-13pro", Brand: "Baofeng", Name: "UV-13 Pro / TP-8Plus", Aliases: []string{"UV-13Pro", "TP-8Plus"}, Models: []string{"UV-13 Pro", "TP-8Plus"}, Type: "analog", Drivers: []string{"bf-uv13pro"}},
	{ID: "5rh-pro", Brand: "Baofeng", Name: "5RH Pro / K67 Pro", Aliases: []string{"5RH-PRO", "K67", "K67 Pro", "UV-5RH Pro"}, Models: []string{"5RH Pro", "K67 Pro", "UV-5RH Pro", "UV-5RH Pro GPS"}, Type: "analog", Drivers: []string{"k67"}, FW: "rh5",
		Notes: "Gestione originale di questo programma (file .xlc del CPS «5RH_PRO»). Firmware .dat «BF_5RH»: la versione dell'elettronica (V1/V2) deve corrispondere."},

	// ------------------------------------------------ Baofeng serie UV-5R / UV-82 / GT
	{ID: "uv-5r", Brand: "Baofeng", Name: "UV-5R", Aliases: []string{"UV-5RA", "UV-5RE", "UV-5RQ", "UV-5R EX", "UV-5RB", "GT-3", "GT-3 MkII"}, Models: []string{"UV-5R", "UV-5RA", "UV-5RE", "UV-5RQ", "UV-5R EX", "UV-5RB", "GT-3", "GT-3 Mk II"}, Type: "analog", Drivers: []string{"bf-uv5r"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5_Series_UV-5R_UV-5RA_UV-5RE_UV-5R_GT-3Series_GT-3_GT-3TP_Programming_Software_update.rar"), manual(s3 + "Baofeng_UV-5R_UV-5R-Plus_GT-3_Users_Manual.pdf")}},
	{ID: "uv-5r-pro", Brand: "Baofeng", Name: "UV-5R+Pro / UV-5XP", Aliases: []string{"UV-5R Pro", "UV-5XP"}, Models: []string{"UV-5R+Pro", "UV-5XP"}, Type: "analog", Drivers: []string{"bf-uv5r", "bf-f8hp"},
		Notes: "Elettronica della serie UV-5R: scegli UV-5R oppure BF-F8HP (8 W) e leggi prima la radio; se il modello non è quello giusto la radio non risponde e non viene modificata."},
	{ID: "bf-f8hp", Brand: "Baofeng", Name: "BF-F8HP / UV-5RTP", Aliases: []string{"UV-5RTP", "GT-3TP", "UV-5R 8W"}, Models: []string{"BF-F8HP", "UV-5RTP", "GT-3TP", "UV-5R 8W"}, Type: "analog", Drivers: []string{"bf-f8hp"},
		Links: []catLink{cps(s3 + "Baofeng_UV5RTP_CPS_ENG.zip")}},
	{ID: "uv-5rx3", Brand: "Baofeng", Name: "UV-5RX3", Aliases: []string{"BTECH UV-5X3"}, Type: "analog", Drivers: []string{"bf-uv5rx3-btech", "bf-uv5rx3"},
		Links: []catLink{cps("https://baofeng-uv-5rx3.s3.amazonaws.com/UV-5RX3%20Programming%20Software%20v2.7.zip")},
		Notes: "Due versioni del firmware: tribanda (protocollo BTECH) o tipo UV-5R. Se una non risponde prova l'altra."},
	{ID: "uv-82", Brand: "Baofeng", Name: "UV-82 / UV-82L", Aliases: []string{"UV-82", "UV-82L", "UV-82C"}, Models: []string{"UV-82", "UV-82L", "UV-82C"}, Type: "analog", Drivers: []string{"bf-uv82"},
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
	{ID: "bf-a58", Brand: "Baofeng", Name: "BF-A58 / UV-9R Pro", Aliases: []string{"UV-9R Pro"}, Models: []string{"BF-A58", "UV-9R Pro"}, Type: "analog", Drivers: []string{"bf-a58"}},
	{ID: "bf-a58s", Brand: "Baofeng", Name: "BF-A58S / UV-82III", Aliases: []string{"UV-82III"}, Models: []string{"BF-A58S", "UV-82III"}, Type: "analog", Drivers: []string{"bf-a58s"}},
	{ID: "uv-s9x3", Brand: "Baofeng", Name: "UV-S9X3", Type: "analog", Drivers: []string{"bf-uvs9x3"}},
	{ID: "uv-5rx", Brand: "Baofeng", Name: "UV-5RX (5RX)", Aliases: []string{"5RX"}, Type: "analog", Drivers: []string{"bf-5rx"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5RX_CPS_20231124.zip")}},
	{ID: "uv-9r", Brand: "Baofeng", Name: "UV-9R / UV-9R Plus", Aliases: []string{"UV-9R Plus"}, Models: []string{"UV-9R", "UV-9R Plus"}, Type: "analog", Drivers: []string{"bf-uv9r"},
		Links: []catLink{cps(s3 + "BAOFENG_UV-9R_CPS_20231019.zip"), manual(s3 + "BAOFENG_UV_9R_User_Manual_20231019_EN.pdf")}},
	{ID: "uv-9g", Brand: "Baofeng", Name: "UV-9G", Type: "gmrs", Drivers: []string{"bf-uv9g"}, FW: "kdh9g",
		Links: []catLink{cps(s3 + "BAOFENG_UV-9G_Programming_Software_V2.5_20210916.rar"),
			{"fw", "Aggiornamento del costruttore (trasmissione sul canale 15)", s3 + "Baofeng_UV-9G_Update_Transmit_on_Channel_15_202311.zip"},
			manual(s3 + "BAOFENG_UV-9G_GMRS_User_Manual_20210806.pdf")},
		Notes: "Firmware .kdh (T52GMRS_…) con il protocollo del programma del costruttore «Bootloader_BaudRate38400». Attenzione: il caricatore non controlla il modello."},

	// ------------------------------------------------ Baofeng PMR / FRS / piccole
	{ID: "bf-888s", Brand: "Baofeng", Name: "BF-888S / BF-777S", Aliases: []string{"BF-888", "H-777", "BF-777S"}, Models: []string{"BF-888S", "BF-777S"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Links: []catLink{cps(s3 + "Baofeng_BF-888S_CPS_NEW_20240301.zip"), manual("http://s3.image.ro.s3.amazonaws.com/download/BaoFeng-BF888s-Manual.pdf")}},
	{ID: "bf-m4", Brand: "Baofeng", Name: "BF-M4", Type: "pmr", Drivers: []string{"bf-m4"}},
	{ID: "mp31", Brand: "Baofeng", Name: "MP31", Type: "gmrs", Drivers: []string{"bf-mp31"}},
	{ID: "bf-1901", Brand: "Baofeng", Name: "BF-1901", Type: "pmr", Drivers: []string{"bf-1901"}},
	{ID: "bf-1904", Brand: "Baofeng", Name: "BF-1904", Type: "pmr", Drivers: []string{"bf-1904"}},
	{ID: "bf-1909", Brand: "Baofeng", Name: "BF-1909", Type: "pmr", Drivers: []string{"bf-1909"}},
	{ID: "uv-b5", Brand: "Baofeng", Name: "UV-B5 / UV-B6", Aliases: []string{"UV-B6"}, Models: []string{"UV-B5", "UV-B6"}, Type: "analog", Drivers: []string{"bf-uvb5"}},
	{ID: "uv-3r", Brand: "Baofeng", Name: "UV-3R", Type: "analog", Drivers: []string{"bf-uv3r"}},
	{ID: "uv-3r-plus", Brand: "Baofeng", Name: "UV-3R+", Aliases: []string{"UV-3R Plus"}, Type: "analog", Drivers: []string{"bf-uv3r"},
		Links: []catLink{{"cps", "Programma «UV3R» (pagina dei download di baofengradio.co.uk)", "https://baofengradio.co.uk/software/"}},
		Notes: "Sperimentale: secondo la segnalazione #239 di CHIRP la UV-3R+ funziona con il driver della UV-3R (i problemi segnalati erano del cavo). Compatibilità indicata dagli utenti di CHIRP, non dal costruttore."},
	{ID: "bf-t1", Brand: "Baofeng", Name: "BF-T1", Type: "pmr", Drivers: []string{"bf-t1"}},
	{ID: "bf-t9", Brand: "Baofeng", Name: "BF-T9", Aliases: []string{"T9"}, Type: "pmr", Drivers: []string{"bf-t9"},
		Links: []catLink{{"cps", "Programma «BF-Ts» v1.04 (pacchetto «BF-T9 Software», pagina dei download Baofeng)", bfDown}},
		Notes: "Protocollo e memoria ricavati dal programma del costruttore «BF-Ts» e provati con il suo codice originale: 99 canali senza nome; " +
			"la gamma di frequenze ammessa è quella memorizzata nella radio."},
	{ID: "gt-18", Brand: "Baofeng", Name: "GT-18 (PMR446)", Aliases: []string{"GT-18", "GT18"}, Type: "pmr", Drivers: []string{"bf-gt18pmr"},
		Links: []catLink{{"cps", "Programma «BAOFENG GT18 PMR» v1.2.1 (pagina dei download Baofeng)", bfDown}},
		Notes: "Il programma del costruttore è lo stesso della BF-T17 PMR446 (cambia solo il comando di aggancio): protocollo e memoria provati con il suo codice originale. " +
			"Se è impostata la password di lettura e scrittura, toglila prima con il programma del costruttore."},
	{ID: "bf-t8", Brand: "Baofeng", Name: "BF-T8 / BF-U9", Aliases: []string{"BF-U9"}, Models: []string{"BF-T8", "BF-U9"}, Type: "pmr", Drivers: []string{"bf-t8"}},
	{ID: "bf-v8a", Brand: "Baofeng", Name: "BF-V8A", Type: "pmr", Drivers: []string{"bf-v8a"}},
	{ID: "bf-t20frs", Brand: "Baofeng", Name: "BF-T20FRS", Type: "pmr", Drivers: []string{"bf-t20frs"}},
	{ID: "bf-t20", Brand: "Baofeng", Name: "BF-T20", Type: "pmr", Drivers: []string{"bf-t20"}},
	{ID: "w31e", Brand: "Baofeng", Name: "W31E", Aliases: []string{"W31"}, Type: "pmr", Drivers: []string{"bf-w31e"},
		Notes: "Il programma del costruttore per la W31 («CPS», aggancio «PROGRGS», 16 canali da 0x0010 ed elenco di scansione in 0x0178) usa lo stesso protocollo e la stessa memoria del driver W31E di CHIRP."},
	{ID: "w31d", Brand: "Baofeng", Name: "W31D (GMRS)", Type: "gmrs", Drivers: []string{"bf-w31d"},
		Notes: "Versione digitale (serie D) per gli USA: 30 canali. Protocollo del programma del costruttore «BF Series CPS», presente anche in CHIRP («Baofeng digital»)."},
	{ID: "bf-t20d", Brand: "Baofeng", Name: "BF-T20D", Aliases: []string{"T20D"}, Type: "pmr", Drivers: []string{"bf-t20d"},
		Notes: "Versione digitale (serie D) della BF-T20: 30 canali. Protocollo del programma del costruttore «BF Series CPS», presente anche in CHIRP («Baofeng digital»)."},

	// ------------------------------------------------ Baofeng solo con il programma del costruttore
	{ID: "uv-25-plus", Brand: "Baofeng", Name: "UV-25 Plus", Type: "analog", Drivers: []string{"bf-uv25plus"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_UV_25_Plus_CPS_20250227.zip"), {"fw", "Aggiornamento del firmware (guida e programma)", s3 + "Baofeng_UV-25_Plus_Firware_Update_Guide_20250526.zip"},
			manual(s3 + "Baofeng_UV-25_PLUS_User_Manual_EN_IT_DE_ES_FR_20250227.pdf")},
		Notes: "Canali secondo il programma del costruttore «BF-UVs CPS». Firmware .kdhx con il protocollo del programma «Baofeng Bootloader Pro»."},
	{ID: "uv-5rm-plus-gps", Brand: "Baofeng", Name: "UV-5RM Plus GPS", Type: "analog", Drivers: []string{"k67"},
		Links: []catLink{cps(s3 + "Baofeng_UV-5RM_Plus_GPS_CPS_V1.17.zip")},
		Notes: "Il programma ufficiale di questa radio è il CPS «5RH PRO» (lo stesso della 5RH Pro / K67 Pro): si gestisce dalla scheda 5RH Pro / K67 Pro (file .xlc)."},
	{ID: "bf-t60", Brand: "Baofeng", Name: "BF-T60", Models: []string{"BF-T60", "TR-988UV"}, Type: "analog", Drivers: []string{"bf-uv17r"}, FW: "bf",
		Links: []catLink{{"cps", "Programma «T6UV Series EN CPS» (pacchetto «T6UV Series Software», pagina dei download di baofengradio.co.uk)", "https://baofengradio.co.uk/software/"}},
		Notes: "Il costruttore pubblica per questa radio il programma «T6UV Series CPS», lo stesso di UV-17R, UV-21R e UV-21L (aggancio «PROGRAMBFNORMALU», memoria della UV-17Pro): si usa lo stesso driver. " +
			"Il programma riconosce i modelli dal nome inviato dalla radio (UV17…UV29, 5RH, 1801, 1702/06): se la radio non risponde come una radio di questa famiglia, la lettura si ferma senza modificarla."},
	{ID: "uv-17r", Brand: "Baofeng", Name: "UV-17R", Type: "analog", Drivers: []string{"bf-uv17r"}, FW: "bf",
		Links: []catLink{cps(s3 + "BAOFENG_UV-17R_CPS_20230801.zip"), manual(s3 + "BAOFENG_UV_17R_User_Manual_20230815_EN.pdf")}},
	{ID: "uv-21r", Brand: "Baofeng", Name: "UV-21R", Type: "analog", Drivers: []string{"bf-uv21r"}, FW: "bf",
		Links: []catLink{cps(s3 + "BAOFENG_UV-21R_CPS_20240902.zip"), manual(s3 + "Baofeng_UV-21R_User_Manual_EN_20240902.pdf")}},
	{ID: "uv-21l", Brand: "Baofeng", Name: "UV-21L", Type: "analog", Drivers: []string{"bf-uv21l"}, FW: "bf", Links: []catLink{cps(s3 + "Baofeng_UV-21L_CPS_20231130.zip")}},
	{ID: "uv-21", Brand: "Baofeng", Name: "UV-21", Type: "analog", Drivers: []string{"bf-uv21r"}, FW: "bf",
		Notes: "Il programma del costruttore «T6UV Series CPS» (UV-17R, UV-21R, UV-21L) gestisce anche il modello «UV21»: stesso protocollo e stessa memoria."},
	{ID: "uv-29pro", Brand: "Baofeng", Name: "UV-29 Pro", Models: []string{"UV-29 Pro", "SW-9"}, Type: "analog", Drivers: []string{"bf-uv29pro"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_UV-29_PRO_SW-9_CPS_v1.6.1.zip"), manual(s3 + "Baofeng_UV-29_Pro_User_Manual_EN_DE_FR_IT_ES_2026417.pdf")}},
	{ID: "uv-7b", Brand: "Baofeng", Name: "UV-7B", Type: "analog", Drivers: []string{"bf-uv7b"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_UV-7B_Programming_Software_20260205.zip"), manual(s3 + "Baofeng_UV-7B_User_Manual_EN_20251024.pdf")}},
	{ID: "gm-15pro", Brand: "Baofeng", Name: "GM-15 Pro", Type: "gmrs", Drivers: []string{"bf-gm15pro"},
		Links: []catLink{cps(s3 + "BAOFENG_GM-15_PRO_CPS_20240902.zip"), manual(s3 + "Baofeng_GM-15_Pro_User_Manual_EN_20240902.pdf")}},
	{ID: "ar-152", Brand: "Baofeng", Name: "AR-152", Type: "analog", Drivers: []string{"bf-uv5r"},
		Links: []catLink{{"cps", "Programma «AR152» (pacchetto «AR-152 Software», pagina dei download Baofeng)", bfDown}},
		Notes: "Il programma del costruttore della AR-152 usa l'aggancio della UV-5R e nella tabella dei modelli equivalenti di CHIRP la AR-152 corrisponde alla UV-5R: si usa il driver della UV-5R. " +
			"La AR-152 Pro è una radio diversa (scheda separata)."},
	{ID: "ar-152pro", Brand: "Baofeng", Name: "AR-152 Pro", Type: "analog", Drivers: []string{"bf-ar152pro"}, FW: "kdhx",
		Links: []catLink{cps(s3 + "Baofeng_AR-152_PRO_programming_software_v1.4.6.zip")}},
	{ID: "uv-98pro", Brand: "Baofeng", Name: "UV-98 Pro", Type: "analog", Drivers: []string{"bf-uv9r"},
		Links: []catLink{{"cps", "Programma «WP970I (T51) CPS» (pacchetto «Baofeng UV-98 Pro Software», pagina dei download Baofeng)", bfDown}},
		Notes: "Sperimentale: per la UV-98 Pro il costruttore pubblica il programma della WP970I (T51). In CHIRP la WP970I è la base della famiglia " +
			"UV-9R / BF-A58 / UV-82WP (stessa memoria e stesso protocollo): si usa il driver della UV-9R (identica alla WP970I, nomi di 7 caratteri). " +
			"Se la radio non risponde come una WP970I, la lettura si ferma senza modificarla: prima di scrivere conserva una copia della lettura."},
	{ID: "uv-5r3", Brand: "Baofeng", Name: "UV-5R III", Aliases: []string{"UV-5RIII", "UV-5R3"}, Type: "analog", Drivers: []string{"bf-uv5rx3"},
		Links: []catLink{{"cps", "Programma «BF_5R_ThreeBand» (pacchetto «UV-5R III Software», pagina dei download Baofeng)", bfDown}},
		Notes: "Radio a tre bande: nella tabella dei modelli equivalenti di CHIRP la UV-5RIII corrisponde alla Radioddity UV-5RX3."},
	{ID: "bf-r5", Brand: "Baofeng", Name: "BF-R5", Aliases: []string{"BF-R5 Mini", "R5"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Links: []catLink{{"cps", "Programma «BF_480» v1.5 (pagina dei download Baofeng)", bfDown}},
		Notes: "Il programma del costruttore («BF_480») usa il protocollo della BF-888S e nella tabella dei modelli equivalenti di CHIRP la BF-R5 Mini corrisponde alla BF-888: si usa il driver della BF-888S."},
	{ID: "uv-16", Brand: "Baofeng", Name: "UV-16 / UV-16H Pro", Aliases: []string{"UV-16H Pro"}, Models: []string{"UV-16", "UV-16H Pro"}, Type: "analog", Drivers: []string{"bf-f8hp"},
		Notes: "Sperimentale: con CHIRP diversi utenti leggono e scrivono la UV-16 con il driver della BF-F8HP (segnalazione #9833 di CHIRP). Compatibilità segnalata dagli utenti di CHIRP, non confermata dal costruttore: prima della scrittura il programma salva sempre una copia della radio, controlla che i canali letti siano corretti prima di modificare."},
	{ID: "p52uv", Brand: "Baofeng", Name: "P52UV", Type: "analog", Drivers: []string{"bf-f8hp", "bf-uv5r"},
		Notes: "Sperimentale: con CHIRP la P52UV è stata programmata con il driver della BF-F8HP (segnalazione #9405 di CHIRP); con due sole potenze è indicato anche quello della UV-5R. Compatibilità segnalata dagli utenti di CHIRP, non confermata dal costruttore: prima della scrittura il programma salva sempre una copia della radio, controlla che i canali letti siano corretti prima di modificare."},
	{ID: "gmrs-9r", Brand: "Baofeng", Name: "GMRS-9R", Type: "gmrs", Drivers: []string{"bf-uv9g"},
		Notes: "Stessa radio della UV-9G: CHIRP la gestisce con il driver della UV-9G (segnalazione #10388 di CHIRP). Come in CHIRP, fuori dalle frequenze GMRS il canale è solo in ricezione."},
	{ID: "g11s", Brand: "Baofeng", Name: "G11S", Aliases: []string{"Retevis RB27"}, Type: "gmrs", Drivers: []string{"bf-g11s"},
		Notes: "È la Retevis RB27 (tabella dei modelli equivalenti di CHIRP): driver della RB27 di CHIRP."},
	{ID: "5rham", Brand: "Baofeng", Name: "5RHam", Type: "analog"},
	{ID: "bf-9700", Brand: "Baofeng", Name: "BF-9700 / BF-S56 Max", Aliases: []string{"BF-S56 Max", "BF-S56"}, Models: []string{"BF-9700", "BF-S56 Max"}, Type: "analog", Drivers: []string{"bf-9700"},
		Notes: "Stessa memoria della BF-1909 di CHIRP (identificativi «P3107» e «P320h»), con la memoria letta fino a 0x0400: driver proposto a CHIRP da un utente che lo usa in lettura e scrittura sulle sue BF-9700 e BF-S56 Max (segnalazione #11696 di CHIRP).",
		Links: []catLink{cps("https://s3.amazonaws.com/s3.image.ro/download/BF9700Setup_EN.rar"), manual("http://s3.image.ro.s3.amazonaws.com/download/BF-9700USERMANUAL.pdf")}},
	{ID: "uv-9s", Brand: "Baofeng", Name: "UV-9S", Type: "analog", Drivers: []string{"bf-uv5rx3"},
		Notes: "Nella tabella dei modelli equivalenti di CHIRP la UV-9S corrisponde alla Radioddity UV-5RX3."},
	{ID: "uv-8d", Brand: "Baofeng", Name: "UV-8D", Type: "analog"},
	{ID: "uv-e70", Brand: "Baofeng", Name: "UV-E70", Type: "analog", Drivers: []string{"bf-uv82wp"},
		Notes: "Sperimentale: con CHIRP la UV-E70 è stata letta e scritta con il driver della UV-82WP, tenendo il volume a metà (segnalazione #11280 di CHIRP). Compatibilità segnalata dagli utenti di CHIRP, non confermata dal costruttore: prima della scrittura il programma salva sempre una copia della radio, controlla che i canali letti siano corretti prima di modificare."},
	{ID: "bf-uv8", Brand: "Baofeng", Name: "BF-UV8", Type: "analog"},
	{ID: "bf-r3", Brand: "Baofeng", Name: "BF-R3", Type: "analog", Drivers: []string{"bf-uv5rx3"},
		Notes: "Nella tabella dei modelli equivalenti di CHIRP la BF-R3 corrisponde alla Radioddity UV-5RX3."},
	{ID: "h-series", Brand: "Baofeng", Name: "Serie H (H-31, H-32, BF-H2/H5/H6)", Aliases: []string{"H-31", "H-32", "BF-H2", "BF-H5", "BF-H6"}, Models: []string{"H-31", "H-32", "BF-H2", "BF-H5", "BF-H6"}, Type: "analog", Drivers: []string{"bf-hseries"},
		Links: []catLink{cps(s3 + "BAOFENG%20H-Series%20Programming%20Software.rar")}},
	{ID: "bf-888sd", Brand: "Baofeng", Name: "BF-888SD", Aliases: []string{"Serie D"}, Type: "pmr", Drivers: []string{"bf-888sd"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-1901d", Brand: "Baofeng", Name: "BF-1901D", Aliases: []string{"Serie D"}, Type: "pmr", Drivers: []string{"bf-1901d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-889d", Brand: "Baofeng", Name: "BF-889D", Aliases: []string{"Serie D"}, Type: "pmr", Drivers: []string{"bf-889d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v2d", Brand: "Baofeng", Name: "BF-V2D", Aliases: []string{"V2D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v2d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v4d", Brand: "Baofeng", Name: "BF-V4D", Aliases: []string{"V4D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v4d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v5d", Brand: "Baofeng", Name: "BF-V5D", Aliases: []string{"V5D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v5d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v6d", Brand: "Baofeng", Name: "BF-V6D", Aliases: []string{"V6D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v6d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v8d", Brand: "Baofeng", Name: "BF-V8D", Aliases: []string{"V8D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v8d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v9d", Brand: "Baofeng", Name: "BF-V9D", Aliases: []string{"V9D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v9d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v10d", Brand: "Baofeng", Name: "BF-V10D", Aliases: []string{"V10D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v10d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-v12d", Brand: "Baofeng", Name: "BF-V12D", Aliases: []string{"V12D", "EU-V12D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-v12d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-1904d", Brand: "Baofeng", Name: "MRA-1904 / BF-1904D", Aliases: []string{"MRA-1904", "BF-1904D", "1904D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-1904d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-1909d", Brand: "Baofeng", Name: "BF-1909D", Aliases: []string{"1909D", "Serie D"}, Type: "pmr", Drivers: []string{"bf-1909d"},
		Links: []catLink{cps(s3 + "Baofeng_BF-D_Series_CPS.zip")}, Notes: notesSerieD},
	{ID: "bf-c2", Brand: "Baofeng", Name: "BF-C2 (BF-V2)", Aliases: []string{"BF-V2"}, Type: "pmr", Drivers: []string{"bf-c2"},
		Links: []catLink{cps(s3 + "BAOFENG_BF-V2_CPS_20231019.zip"), manual(s3 + "BAOFENG_BF_V2_User_Manual_20231019_EN.pdf")}},
	{ID: "bf-v", Brand: "Baofeng", Name: "BF-V1 / V3 / V4 / V5 / V6A", Aliases: []string{"BF-V1", "BF-V3", "BF-V4", "BF-V5", "BF-V6A"}, Models: []string{"BF-V1", "BF-V3", "BF-V4", "BF-V5", "BF-V6A"}, Type: "pmr"},
	{ID: "bf-88st", Brand: "Baofeng", Name: "BF-88ST / BF-88ST Pro", Aliases: []string{"BF-88ST Pro"}, Models: []string{"BF-88ST", "BF-88ST Pro"}, Type: "pmr", Drivers: []string{"bf-88st"},
		Links: []catLink{cps(s3 + "Baofeng%20BF-88ST%20Programming%20Software_20200722.zip"), manual(s3 + "Baofeng%20BF-88ST%20User%20Manual_20200723.pdf")}},
	{ID: "bf-88a", Brand: "Baofeng", Name: "BF-88A / BF-88E", Aliases: []string{"BF-88E"}, Models: []string{"BF-88A", "BF-88E"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Links: []catLink{cps("https://bf-88a88e.s3.amazonaws.com/Programming%20Software%20%26%20Chirp%20BF-88A%20%26%2088E.zip")},
		Notes: "Il pacchetto del costruttore indica di programmarla con CHIRP come «Baofeng BF-888»: si usa lo stesso driver della BF-888S."},
	{ID: "t22", Brand: "Baofeng", Name: "T22", Type: "pmr", Drivers: []string{"bf-t22"},
		Links: []catLink{cps(s3 + "Baofeng_T22_Programming_Software_PMR.zip"), manual(s3 + "Baofeng_T22_User_Manual_PMR_20250122.pdf")}},
	{ID: "bf-33c", Brand: "Baofeng", Name: "BF-33C", Type: "pmr", Links: []catLink{manual(s3 + "Baofeng_BF-33C_User_Manual_EU_Version.pdf")}},
	{ID: "v1d", Brand: "Baofeng", Name: "V1D", Aliases: []string{"BF-V1D"}, Type: "pmr", Drivers: []string{"bf-v1d"},
		Links: []catLink{cps(s3 + "Baofeng_V1D_CPS.zip"), manual(s3 + "Baofeng_V1D_User_Manual.pdf")},
		Notes: "Protocollo e memoria ricavati dal programma del costruttore («BF Series CPS»): canali analogici e digitali (criptati)."},
	{ID: "bf-t17", Brand: "Baofeng", Name: "BF-T17 / BF-T18", Aliases: []string{"BF-T17", "BF-T18", "T17", "T18"}, Models: []string{"BF-T17", "BF-T18"}, Type: "pmr", Drivers: []string{"bf-t17pmr", "bf-t17frs"},
		Links: []catLink{{"cps", "Programmi «T17-T18 PMR446 PRO» e «T17-T18 FRS» v1.2.2 (pagina dei download Radioddity)", rdDown}},
		Notes: "Protocollo e memoria ricavati dai programmi del costruttore e provati con il loro codice originale: 99 canali (versione PMR446) o 35 (versione FRS per gli USA). " +
			"Se è impostata la password di lettura e scrittura, toglila prima con il programma del costruttore."},
	{ID: "bf-pmr-old", Brand: "Baofeng", Name: "BF-528 / 600S / 666S / 700S / 800S", Aliases: []string{"BF-528", "BF-600S", "BF-666S", "BF-700S", "BF-800S"}, Models: []string{"BF-528", "BF-600S", "BF-666S", "BF-700S", "BF-800S"}, Type: "pmr",
		Notes: "La BF-777S è gestita con la BF-888S (tabella dei modelli equivalenti di CHIRP). Per questi modelli non c'è una conferma di compatibilità: la BF-666S è indicata come probabile clone della BF-888S, ma senza prove."},
	{ID: "bf-t-series", Brand: "Baofeng", Name: "BF-T6 / T7 / T11 / T12 / T15 / T16 / T26 / T51 / T57", Aliases: []string{"BF-T6", "BF-T7", "BF-T11", "BF-T12", "BF-T15", "BF-T16", "BF-T26", "BF-T51", "BF-T57"}, Models: []string{"BF-T6", "BF-T7", "BF-T11", "BF-T12", "BF-T15", "BF-T16", "BF-T26", "BF-T51", "BF-T57"}, Type: "pmr"},
	{ID: "bf-other", Brand: "Baofeng", Name: "BF-358 / A68 / B5 / B6 / E51 / S88", Aliases: []string{"BF-358", "BF-A68", "BF-B5", "BF-B6", "BF-E51", "BF-S88"}, Models: []string{"BF-358", "BF-A68", "BF-B5", "BF-B6", "BF-E51", "BF-S88"}, Type: "pmr"},
	{ID: "ar888s", Brand: "Baofeng", Name: "AR888S / TH-888S", Aliases: []string{"AR888S", "TH888S", "TH-888S"}, Models: []string{"AR888S", "TH-888S"}, Type: "pmr", Drivers: []string{"bf-888s"},
		Notes: "Nella tabella dei modelli equivalenti di CHIRP la TH-888S corrisponde alla BF-888: si usa il driver della BF-888S."},
	{ID: "bf-t3", Brand: "Baofeng", Name: "BF-T3 (per bambini)", Type: "kids"},
	{ID: "gt-22", Brand: "Baofeng", Name: "GT-22", Type: "pmr", Drivers: []string{"bf-gt22"},
		Links: []catLink{{"cps", "Programma «GT-22 CPS» (pacchetto «GT-22 Programming Software», pagina dei download Radioddity)", s3 + "GT-22%20Programming%20Software_20200318.zip"},
			{"info", "Scheda del prodotto (baofengradio.com)", "https://baofengradio.com/products/gt-22"}},
		Notes: "PMR446 senza licenza, 16 canali. Ricavata dal programma del costruttore «GT-22» e verificata con il suo codice originale: stessa memoria " +
			"e stesso protocollo della BF-888S (in CHIRP la GT-22 si legge come Retevis H777, segnalazione #9633)."},
	{ID: "gt-68", Brand: "Baofeng", Name: "GT-68", Type: "gmrs", Drivers: []string{"bf-gt68"},
		Links: []catLink{cps(s3 + "Baofeng_GT-68_CPS.zip"), {"info", "Scheda del prodotto (baofengradio.com)", "https://baofengradio.com/products/gt-68"}},
		Notes: "Radio FRS (USA), 16 canali. Ricavata dal programma del costruttore «BF_V1_FH» (pacchetto GT-68, lo stesso della BF-C2 / BF-V2) e " +
			"verificata con il suo codice originale: stessa memoria e stesso protocollo della BF-888S."},
	{ID: "dm-1701", Brand: "Baofeng", Name: "DM-1701", Aliases: []string{"Retevis RT84"}, Type: "dmr", Drivers: []string{"bf-dm1701"},
		Links: []catLink{cps(s3 + "Baofeng_DM-1701_CPS_1.05.zip"),
			catLink{"info", "qdmr (programma libero per radio DMR, GPL)", "https://dm3mat.de/software/qdmr"}, catLink{"info", "dmrconfig (programma libero da riga di comando, licenza BSD)", "https://github.com/OpenRTX/dmrconfig"}},
		Notes: "Radio DMR della famiglia TYT MD-UV380, collegata via USB in modalità DFU: serve il driver USB del costruttore (cartella " +
			"«DM-1701 USB Driver» del pacchetto «DM-1701 CPS»), lo stesso del suo programma. Memoria e protocollo dei programmi liberi " +
			"dmrconfig e qdmr, verificati con il codice originale di dmrconfig (lettura, modifiche, cancellazione e scrittura identiche). " +
			"Con il firmware libero OpenGD77 la radio usa un altro protocollo: in quel caso usa qdmr."},
	{ID: "dm-1702", Brand: "Baofeng", Name: "DM-1702", Type: "dmr",
		Links: []catLink{{"cps", "Programma «DM-1702 MD Setup» v1.00.76 (pagina dei download Baofeng)", bfDown}},
		Notes: "Radio DMR: il programma del costruttore («DM-1702 MD Setup», famiglia TYT MD) non è documentato; il codeplug si gestisce con quel programma."},
	{ID: "dm-1801", Brand: "Baofeng", Name: "DM-1801", Type: "dmr", Drivers: []string{"bf-dm1801"},
		Links: []catLink{catLink{"info", "dmrconfig (programma libero da riga di comando, licenza BSD)", "https://github.com/OpenRTX/dmrconfig"}, catLink{"info", "qdmr (programma libero per radio DMR, GPL)", "https://dm3mat.de/software/qdmr"}},
		Notes: "Radio DMR (piattaforma della Radioddity GD-77), collegata via USB come dispositivo HID: nessuna porta COM e nessun driver da installare. " +
			"Memoria e protocollo dei programmi liberi dmrconfig e qdmr, verificati con il codice originale di dmrconfig (lettura, modifiche e scrittura " +
			"identiche). Con il firmware libero OpenGD77 la radio usa un altro protocollo: in quel caso usa qdmr."},
	{ID: "dm-1801pro", Brand: "Baofeng", Name: "DM-1801 Pro", Type: "dmr", Drivers: []string{"bf-dm32uv"},
		Links: []catLink{{"cps", "Programma «CPS DMR Radio» v1.60 (pagina dei download Radioddity)", rdDown}},
		Notes: "Sperimentale. Il costruttore pubblica per la DM-1801 Pro lo stesso programma della DM-32UV (stesso eseguibile «DMR CPS.exe», v1.60): si usa il driver della DM-32UV. " +
			"Se la radio si identifica con un modello diverso, il programma si ferma prima di leggere o scrivere."},
	{ID: "dm-1703", Brand: "Baofeng", Name: "DM-1703", Type: "dmr"},
	{ID: "dr-1801uv", Brand: "Baofeng", Name: "DR-1801UV", Aliases: []string{"BF-1801A6", "DR-1801", "BTECH BF-1801A6"}, Type: "dmr", Drivers: []string{"bf-dr1801uv"},
		Links: []catLink{{"cps", "Programma «DR-1801A6» (pacchetto «Baofeng_DR-1801UV_Programming_Software», pagina dei download Baofeng)", bfDown},
			catLink{"info", "qdmr (programma libero per radio DMR, GPL)", "https://dm3mat.de/software/qdmr"}},
		Notes: "Radio DMR (piattaforma Auctus A6). Protocollo e memoria ricavati dal programma del costruttore «DR-1801A6» e verificati con il suo codice originale: " +
			"lettura del file di fabbrica, modifiche scritte con byte identici a quelli del programma e lettura e scrittura via cavo del programma con la radio simulata. " +
			"Si modificano frequenze, nomi, potenza, banda e toni; per i canali DMR contatti, gruppi, slot e color code restano quelli impostati. " +
			"Differenza dal programma del costruttore: cancellando un canale gli altri non vengono rinumerati."},
	{ID: "dm-5r", Brand: "Baofeng", Name: "DM-5R", Aliases: []string{"DM-5R Tier I"}, Type: "dmr",
		Links: []catLink{{"cps", "Programma «CPS_5R» (archivio «DM_5R_Programming_Software», password 123456, pagina dei download Baofeng)", bfDown}},
		Notes: "Radio DMR Tier I: il programma del costruttore («CPS_5R», scritto in C++Builder, porta seriale a 57600 baud) è un programma " +
			"nativo che non si può eseguire per un confronto e non esistono programmi liberi per questa versione: il codeplug si gestisce con " +
			"il programma del costruttore. La versione Tier II (DM-5R Plus) è la RD-5R, compatibile."},
	{ID: "rd-5r", Brand: "Baofeng", Name: "RD-5R", Aliases: []string{"DM-5R Tier II", "Radioddity RD5R", "TD-5R"}, Type: "dmr", Drivers: []string{"bf-rd5r"},
		Links: []catLink{{"cps", "Programma «RD-5R Program Software» V1.0.0.4 e firmware V2.1.6 (pagina dei download Baofeng)", bfDown},
			catLink{"info", "qdmr (programma libero per radio DMR, GPL)", "https://dm3mat.de/software/qdmr"}, catLink{"info", "dmrconfig (programma libero da riga di comando, licenza BSD)", "https://github.com/OpenRTX/dmrconfig"}},
		Notes: "Radio DMR collegata via USB come dispositivo HID: nessuna porta COM e nessun driver da installare. Memoria e protocollo dei programmi " +
			"liberi dmrconfig e qdmr (la radio si identifica come «BF-5R», anche la TD-5R), verificati con il codice originale di dmrconfig (lettura, " +
			"modifiche e scrittura identiche). Con il firmware libero OpenGD77 (OpenRD5R) la radio usa un altro protocollo: in quel caso usa qdmr."},
	{ID: "dm-v1", Brand: "Baofeng", Name: "DM-V1", Type: "dmr"},
	{ID: "dm-32uv", Brand: "Baofeng", Name: "DM-32UV", Type: "dmr", Drivers: []string{"bf-dm32uv"},
		Links: []catLink{{"cps", "Programma «DM-32UV CPS» v1.60 e firmware v1.01.048 (pagina dei download Radioddity)", rdDown},
			catLink{"info", "CHIRP (driver Baofeng DM-32UV)", "https://chirpmyradio.com"}},
		Notes: "Canali (analogici e DMR) con il driver di CHIRP, ricavato dal programma del costruttore v1.60 e provato su una radio vera: " +
			"si modificano frequenze, nomi, potenza e toni; le impostazioni DMR dei canali, le zone, i contatti e le liste restano come sono. " +
			"Il firmware si aggiorna con il programma del costruttore."},
}

// catalogDefault: nota per i modelli senza driver
const catalogNoDriver = "Protocollo non documentato o radio digitale: la programmazione si fa con il programma del costruttore."

// Modelli equivalenti secondo la tabella di CHIRP (chirp/share/model_alias_map.yaml,
// sezione Baofeng/Pofung): stessa memoria e stesso protocollo del modello di
// riferimento del catalogo; ognuno ha la sua scheda.
var chirpEquiv = []struct{ name, base, chirp string }{
	{"K61", "k6", "K6"}, {"K6A", "k6", "K6"}, {"AR-5RM", "uv-5rm", "5RM"}, {"UV-5RH Pro Max", "uv-5rm", "5RM"},
	{"BF-1902", "bf-1901", "BF-1901"}, {"BF-1903", "bf-1901", "BF-1901"}, {"P15UV (Pofung)", "uv-13pro", "UV-13Pro"},
	{"B-580T", "uv-5r", "UV-5R"}, {"BF-E500S", "uv-5r", "UV-5R"}, {"BF-F8", "uv-5r", "UV-5R"}, {"BF-F8+", "uv-5r", "UV-5R"},
	{"BF-F9", "uv-5r", "UV-5R"}, {"BF-UVB2+", "uv-5r", "UV-5R"}, {"FF-12P", "uv-5r", "UV-5R"}, {"UV-5R V2+", "uv-5r", "UV-5R"},
	{"UV-5R2", "uv-5r", "UV-5R"}, {"UV-5R++", "uv-5r", "UV-5R"}, {"UV-5RA+", "uv-5r", "UV-5R"}, {"UV-5RAX", "uv-5r", "UV-5R"},
	{"UV-5RAX+", "uv-5r", "UV-5R"}, {"UV-5RC", "uv-5r", "UV-5R"}, {"UV-5RCX+", "uv-5r", "UV-5R"}, {"UV-5RD", "uv-5r", "UV-5R"},
	{"UV-5RE+", "uv-5r", "UV-5R"}, {"UV-5RG", "uv-5r", "UV-5R"}, {"UV-5RK", "uv-5r", "UV-5R"}, {"UV-5RS", "uv-5r", "UV-5R"},
	{"UV-5RT", "uv-5r", "UV-5R"}, {"UV-5RU", "uv-5r", "UV-5R"}, {"UV-5S", "uv-5r", "UV-5R"}, {"UV-5X (2014)", "uv-5r", "UV-5R"},
	{"UV-860", "uv-5r", "UV-5R"}, {"UV-920", "uv-5r", "UV-5R"}, {"UV-B2+", "uv-5r", "UV-5R"}, {"UV-B3+", "uv-5r", "UV-5R"},
	{"UV-17M", "uv-17r-plus", "UV-17R-Plus"}, {"BF-18H", "uv-25", "UV-25"}, {"BF-18L", "uv-25", "UV-25"},
	{"BF-C5", "bf-888s", "BF-888"}, {"BF-V9", "bf-888s", "BF-888"}, {"GT-1", "bf-888s", "BF-888"},
	{"BF-F8+III", "uv-5rx3", "Radioddity UV-5RX3"}, {"BF-X9", "uv-5rx3", "Radioddity UV-5RX3"}, {"UV-5RA3", "uv-5rx3", "Radioddity UV-5RX3"},
	{"UV-59T", "uv-5rx3", "Radioddity UV-5RX3"},
	{"BF-F8X3", "bf-a58s", "BF-A58S"}, {"BF-UVF10", "bf-a58s", "BF-A58S"}, {"BF-X3 Plus", "bf-a58s", "BF-A58S"}, {"BF-S5 Plus", "bf-a58s", "UV-82III"},
	{"BF-F9V2+", "bf-f8hp", "BF-F8HP"}, {"BF-UV10R", "bf-f8hp", "BF-F8HP"}, {"UV-10R", "bf-f8hp", "BF-F8HP"}, {"UV-5R HTQ", "bf-f8hp", "BF-F8HP"},
	{"UV-5R MK4", "bf-f8hp", "BF-F8HP"}, {"UV-5R MK5", "bf-f8hp", "BF-F8HP"}, {"UV-5R7W", "bf-f8hp", "BF-F8HP"}, {"UV-5RHP", "bf-f8hp", "BF-F8HP"},
	{"UV-5RM HP", "bf-f8hp", "BF-F8HP"}, {"UV-5R+", "bf-f8hp", "BF-F8HP"}, {"UV-S9 (tre potenze)", "bf-f8hp", "BF-F8HP"}, {"GT-3TP Mark III", "bf-f8hp", "BF-F8HP"},
	{"BF-UV9R+HP", "uv-82wp", "UV-82WP"}, {"UV-5RWP", "uv-82wp", "UV-82WP"}, {"BF-V85", "uv-b5", "UV-B5"}, {"M-5R", "uv-5rx", "5RX"},
	{"P8UV (Pofung)", "uv-5g", "Radioddity UV-5G"}, {"P10UV (Pofung)", "h-series", "Radioddity GA-510"}, {"P11UV (Pofung)", "g11s", "Retevis RB27"},
	{"BF-5RH", "uv-5rh", "Baofeng UV-5RH"}, {"UV-82T", "uv-82x3", "Radioddity UV-82X3"}, {"UV-82III (2 PTT)", "uv-82x3", "Radioddity UV-82X3"},
	{"UV-B2", "uv-82", "UV-82"}, {"GT-5", "uv-82", "UV-82"}, {"GT-5TP", "uv-82hp", "UV-82HP"}, {"UV-8R", "uv-82hp", "UV-82HP"},
	{"UV-9X+", "uv-82hp", "UV-82HP"}, {"UV-X9", "uv-82hp", "UV-82HP"}, {"UV-S9T (tre bande)", "uv-s9x3", "UV-S9X3"},
}

func catSlug(s string) string {
	var b []byte
	for _, c := range []byte(strings.ReplaceAll(strings.ToLower(s), "+", "-plus")) {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b = append(b, c)
		case len(b) > 0 && b[len(b)-1] != '-':
			b = append(b, '-')
		}
	}
	return strings.Trim(string(b), "-")
}

// expandCatalog: una scheda per ogni modello (Models e modelli equivalenti di CHIRP)
func expandCatalog(in []catItem) []catItem {
	var out []catItem
	byID := map[string]catItem{}
	for _, c := range in {
		byID[c.ID] = c
		if len(c.Models) == 0 {
			out = append(out, c)
			continue
		}
		for k, n := range c.Models {
			x := c
			x.Models = nil
			x.Name = n
			if k > 0 {
				x.ID = c.ID + "--" + catSlug(n)
			}
			var others []string
			for _, o := range c.Models {
				if o != n {
					others = append(others, o)
				}
			}
			// gli altri modelli hanno la loro scheda: tra i nomi alternativi restano gli altri
			x.Aliases = nil
			for _, a := range c.Aliases {
				dup := false
				for _, o := range c.Models {
					dup = dup || catSlug(a) == catSlug(o)
				}
				if !dup {
					x.Aliases = append(x.Aliases, a)
				}
			}
			note := "Stessa gestione di: " + strings.Join(others, ", ") + "."
			if x.Notes != "" {
				note += " " + x.Notes
			}
			x.Notes = note
			out = append(out, x)
		}
	}
	for _, e := range chirpEquiv {
		b, ok := byID[e.base]
		if !ok {
			panic("modello di riferimento sconosciuto " + e.base)
		}
		ref := b.Name
		if len(b.Models) > 0 {
			ref = b.Models[0]
		}
		out = append(out, catItem{ID: "eq-" + catSlug(e.name), Brand: b.Brand, Name: e.name, Type: b.Type, Drivers: b.Drivers,
			Notes: "Nella tabella dei modelli equivalenti di CHIRP la " + e.name + " corrisponde alla " + e.chirp +
				": si usa la stessa gestione della " + ref + ". Compatibilità indicata da CHIRP, non dal costruttore."})
	}
	return out
}

// fwFamily: note sul firmware delle radio che usano lo stesso caricatore di una
// radio di cui si conoscono i file (aggiunte alle note della scheda)
var fwFamily = map[string]string{
	"bf": "Firmware: stesso caricatore KDH V1.0 della UV-5RM/UV-5RH (radio del programma del costruttore «T6UV Series CPS», che gestisce anche la 5RH; " +
		"la K5 Plus e la UV-5RM Plus in CHIRP sono derivate dalla UV-5RM). Baofeng non pubblica firmware .BF: servono i file forniti dal costruttore o dal " +
		"rivenditore per questa radio. Nella modalità automatica, se la radio non risponde alla richiesta di aggiornamento della UV-5RM, non viene modificata; " +
		"prima dell'invio si deve scrivere il modello indicato dal nome del file. Se il costruttore fornisce un file .kdhx, la finestra usa il caricatore «Bootloader Pro».",
	"kdhx": "Firmware: stesso caricatore «Baofeng Bootloader Pro» della UV-25 Plus (codice di piattaforma H802; la radio usa lo stesso programma del costruttore " +
		"«BF-UVs Pro CPS» della UV-25 Plus). Baofeng pubblica il firmware .kdhx solo per la UV-25 Plus: per questa radio servono i file forniti dal costruttore " +
		"o dal rivenditore. Nella modalità automatica la radio si aggiorna solo se risponde come piattaforma H802; prima dell'invio si deve scrivere il modello " +
		"indicato dal nome del file.",
}

func init() {
	for i := range catalog {
		c := &catalog[i]
		if n, ok := fwFamily[c.FW]; ok && c.ID != "uv-25-plus" && c.ID != "uv-5rm" && c.ID != "uv-5rh" {
			if c.Notes != "" {
				c.Notes += " "
			}
			c.Notes += n
		}
	}
	catalog = expandCatalog(catalog)
}
