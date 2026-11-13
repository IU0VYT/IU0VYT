// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.30: i TIDRADIO dei programmi «TD-FRS» / «TD-PMR» e «TD-M2Pro»
// (drv_tdq.go): canali e impostazioni generali.

const tdq430 = "Canali senza nome e impostazioni generali. Non è in CHIRP. «Identifica» non distingue i modelli di questa famiglia: scegli tu la scheda." + cps425

var catalog430 = []catItem{
	{ID: "tid-td-m11plus", Brand: "TIDRADIO", Name: "TD-M11 Plus", Aliases: []string{"TD-M11Plus", "M11 Plus", "TD-M11Plus-CH16", "TD-M11Plus-CH22"}, Models: []string{"TD-M11 Plus"},
		Type: "pmr", Drivers: []string{"td-m11plus", "td-m11plus-16", "td-m11plus-22"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "Tre schede, come nel programma del costruttore: 199 canali, 16 canali e 22 canali (quest'ultima con due sole potenze). " +
			"Non è la TD-M11 senza «Plus», che ha un'altra scheda. " + tdq430},
	{ID: "tid-td-712plus", Brand: "TIDRADIO", Name: "TID-712Plus", Aliases: []string{"TD-V712 Plus", "TD-V712Plus", "712 Plus", "TID-712 Plus"}, Models: []string{"TID-712Plus"},
		Type: "analog", Drivers: []string{"td-712plus"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali. Lo stesso programma del costruttore elenca, con la stessa memoria, anche RT128, Airmob A10 Ver 2.0, Ranger V9 e DV-4478+. " + tdq430},
	{ID: "tid-td-m7", Brand: "TIDRADIO", Name: "TD-M7", Aliases: []string{"M7", "TDM7"}, Models: []string{"TD-M7"},
		Type: "analog", Drivers: []string{"td-m7"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)}, Notes: "199 canali. " + tdq430},
	{ID: "tid-td-br09", Brand: "TIDRADIO", Name: "TD-BR09", Aliases: []string{"BR09", "TDBR09"}, Models: []string{"TD-BR09"},
		Type: "analog", Drivers: []string{"td-br09"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)}, Notes: "199 canali. " + tdq430},
	{ID: "tid-td-v900", Brand: "TIDRADIO", Name: "TD-V900", Aliases: []string{"V900", "TDV900"}, Models: []string{"TD-V900"},
		Type: "analog", Drivers: []string{"td-v900"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)}, Notes: "199 canali. " + tdq430},
	{ID: "tid-td-777s", Brand: "TIDRADIO", Name: "TD-777S", Aliases: []string{"777S", "TD777S", "TD-777S-CH16", "TD-777S-CH22"}, Models: []string{"TD-777S"},
		Type: "pmr", Drivers: []string{"td-777s-16", "td-777s-22"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "Due schede, come nel programma del costruttore: 16 canali e 22 canali. " + tdq430},
	{ID: "tid-td-frs", Brand: "TIDRADIO", Name: "TD-FRS", Aliases: []string{"TD FRS"}, Models: []string{"TD-FRS"},
		Type: "pmr", Drivers: []string{"td-frs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "22 canali FRS (Stati Uniti), due potenze. " + tdq430},
	{ID: "tid-td-pmr", Brand: "TIDRADIO", Name: "TD-PMR", Aliases: []string{"TD PMR"}, Models: []string{"TD-PMR"},
		Type: "pmr", Drivers: []string{"td-pmr"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "16 canali PMR446, due potenze. " + tdq430},
	{ID: "tid-td-m2pro", Brand: "TIDRADIO", Name: "TD-M2 Pro", Aliases: []string{"TD-M2Pro", "M2 Pro", "M2Pro"}, Models: []string{"TD-M2 Pro"},
		Type: "pmr", Drivers: []string{"td-m2pro"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo UHF (400-520 MHz nel programma del costruttore). " + tdq430},
}
