// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tk270.py, tk760g.py (4.17).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.17: versioni dei palmari Kenwood TK delle serie 60 e 60G (tabelle di tk270.py
// e tk760g.py di CHIRP): sigla scritta nella memoria della radio, numero dei
// canali, limiti di frequenza del costruttore in MHz e nome della versione.
type kwVar struct {
	id     string
	n      int
	lo, hi int64
	kind   string
}

var kwTkVariants = map[string][]kwVar{
	"TK260_Radio": {
		{"\x50\x30\x32\x36\x30\x20\x00\x00", 4, 136, 150, "F2"},
		{"\x50\x30\x32\x36\x30\x21\x00\x00", 4, 150, 174, "F1"},
	},
	"TK270_Radio": {
		{"\x50\x30\x32\x37\x30\x10\x00\x00", 32, 136, 150, "F2"},
		{"\x50\x30\x32\x37\x30\x11\x00\x00", 32, 150, 174, "F1"},
	},
	"TK272_Radio": {
		{"\x50\x30\x32\x37\x32\x10\x00\x00", 10, 136, 150, "F2"},
		{"\x50\x30\x32\x37\x32\x11\x00\x00", 10, 150, 174, "F1"},
	},
	"TK278_Radio": {
		{"\x50\x30\x32\x37\x38\x00\x00\x00", 32, 136, 150, "F2"},
		{"\x50\x30\x32\x37\x38\x01\x00\x00", 32, 150, 174, "F1"},
	},
	"TK360_Radio": {
		{"\x50\x30\x33\x36\x30\x24\x00\x00", 4, 450, 470, "F1"},
		{"\x50\x30\x33\x36\x30\x25\x00\x00", 4, 470, 490, "F2"},
		{"\x50\x30\x33\x36\x30\x26\x00\x00", 4, 490, 512, "F3"},
		{"\x50\x30\x33\x36\x30\x23\x00\x00", 4, 406, 430, "F4"},
	},
	"TK370_Radio": {
		{"\x50\x30\x33\x37\x30\x14\x00\x00", 32, 450, 470, "F1"},
		{"\x50\x30\x33\x37\x30\x15\x00\x00", 32, 470, 490, "F2"},
		{"\x50\x30\x33\x37\x30\x16\x00\x00", 32, 490, 512, "F3"},
		{"\x50\x30\x33\x37\x30\x13\x00\x00", 32, 406, 430, "F4"},
	},
	"TK372_Radio": {
		{"\x50\x30\x33\x37\x32\x14\x00\x00", 10, 450, 470, "F1"},
		{"\x50\x30\x33\x37\x32\x15\x00\x00", 10, 470, 490, "F2"},
	},
	"TK378_Radio": {
		{"\x50\x30\x33\x37\x38\x04\x00\x00", 32, 370, 470, "SP1"},
		{"\x50\x30\x33\x37\x38\x02\x00\x00", 32, 350, 427, "SP2"},
	},
	"TK260G_Radios": {
		{"\x50\x32\x36\x30\x30\x55\xFF", 8, 136, 150, "N1"},
		{"\x50\x32\x36\x30\x30\x54\xFF", 8, 146, 174, "N"},
		{"\x50\x32\x36\x30\x30\x24\xFF", 8, 150, 174, "E"},
		{"\x50\x32\x36\x30\x30\x14\xFF", 8, 150, 174, "M"},
		{"\x50\x32\x36\x30\x30\x05\xFF", 8, 136, 150, "K1"},
		{"\x50\x32\x36\x30\x30\x04\xFF", 8, 150, 174, "K"},
		{"\x70\x32\x36\x30\x30\x24\xFB", 8, 150, 174, "E?"},
	},
	"TK270G_Radios": {
		{"\x50\x32\x37\x30\x30\x54\xFF", 128, 146, 174, "NE/NT"},
		{"\x50\x32\x37\x30\x30\x24\xFF", 128, 146, 174, "E"},
		{"\x50\x32\x37\x30\x30\x14\xFF", 128, 150, 174, "M"},
		{"\x50\x32\x37\x30\x30\x05\xFF", 128, 136, 150, "K1"},
		{"\x50\x32\x37\x30\x30\x04\xFF", 128, 150, 174, "K"},
	},
	"TK272G_Radios": {
		{"\x50\x32\x37\x32\x30\x05\xFB", 32, 136, 174, "K1"},
		{"\x50\x32\x37\x32\x30\x04\xFB", 32, 150, 174, "K"},
	},
	"TK278G_Radios": {
		{"\x50\x32\x37\x38\x30\x35\xFF", 128, 136, 150, "C1"},
		{"\x50\x32\x37\x38\x30\x34\xFF", 128, 150, 174, "C"},
		{"\x50\x32\x37\x38\x30\x15\xFF", 16, 136, 150, "M1"},
		{"\x50\x32\x37\x38\x30\x14\xFF", 16, 150, 174, "M"},
	},
	"TK360G_Radios": {
		{"\x50\x33\x36\x30\x30\x06\xFF", 8, 450, 470, "K"},
		{"\x50\x33\x36\x30\x30\x07\xFF", 8, 470, 490, "K1"},
		{"\x50\x33\x36\x30\x30\x08\xFF", 8, 490, 512, "K2"},
		{"\x50\x33\x36\x30\x30\x09\xFF", 8, 403, 430, "K3"},
		{"\x50\x33\x36\x30\x30\x26\xFF", 8, 440, 470, "E"},
		{"\x50\x33\x36\x30\x30\x29\xFF", 8, 406, 430, "E1"},
		{"\x50\x33\x36\x30\x30\x16\xFF", 8, 450, 470, "M"},
		{"\x50\x33\x36\x30\x30\x17\xFF", 8, 470, 490, "M1"},
		{"\x50\x33\x36\x30\x30\x19\xFF", 8, 403, 430, "M2"},
		{"\x50\x33\x36\x30\x30\x56\xFF", 8, 440, 470, "NE"},
		{"\x50\x33\x36\x30\x30\x59\xFF", 8, 403, 430, "NE1"},
	},
	"TK370G_Radios": {
		{"\x50\x33\x37\x30\x30\x06\xFF", 128, 450, 470, "K"},
		{"\x50\x33\x37\x30\x30\x07\xFF", 128, 470, 490, "K1"},
		{"\x50\x33\x37\x30\x30\x08\xFF", 128, 490, 512, "K2"},
		{"\x50\x33\x37\x30\x30\x09\xFF", 128, 403, 430, "K3"},
		{"\x50\x33\x37\x30\x30\x16\xFF", 128, 450, 470, "M"},
		{"\x50\x33\x37\x30\x30\x17\xFF", 128, 470, 490, "M1"},
		{"\x50\x33\x37\x30\x30\x18\xFF", 128, 490, 520, "M2"},
		{"\x50\x33\x37\x30\x30\x19\xFF", 128, 403, 430, "M3"},
		{"\x50\x33\x37\x30\x30\x26\xFF", 128, 440, 470, "E"},
		{"\x50\x33\x37\x30\x30\x56\xFF", 128, 440, 470, "NE"},
	},
	"TK372G_Radios": {
		{"\x50\x33\x37\x32\x30\x06\xFB", 32, 450, 470, "K"},
		{"\x50\x33\x37\x32\x30\x07\xFB", 32, 470, 490, "K2"},
		{"\x50\x33\x37\x32\x30\x08\xFB", 32, 490, 512, "K3"},
		{"\x50\x33\x37\x32\x30\x09\xFB", 32, 403, 430, "K4"},
	},
	"TK378G_Radios": {
		{"\x50\x33\x37\x38\x30\x16\xFF", 16, 450, 470, "M"},
		{"\x50\x33\x37\x38\x30\x17\xFF", 16, 400, 420, "M1"},
		{"\x50\x33\x37\x38\x30\x36\xFF", 128, 490, 512, "C"},
		{"\x50\x33\x37\x38\x30\x39\xFF", 128, 403, 430, "C1"},
	},
	"TK388G_Radios": {
		{"\x50\x33\x38\x38\x30\x1B\xFF", 128, 350, 370, "M"},
	},
}
