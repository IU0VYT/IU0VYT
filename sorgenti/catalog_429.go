// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.29: i modelli del programma «TIDRadioCPS» del 2026 (drv_tdh8.go,
// opzione cps26): TD-H3 Plus e TD-H3 SE, TD-H7 Plus, TD-H8 di quarta
// generazione, TD-H9, TD-638UV. Solo canali.

var catalog429 = []catItem{
	{ID: "tid-td-h3plus", Brand: "TIDRADIO", Name: "TD-H3 Plus", Aliases: []string{"H3 Plus", "TDH3 Plus", "TD-H3 SE Mini", "H3 SE"}, Models: []string{"TD-H3 Plus", "TD-H3 SE"},
		Type: "analog", Drivers: []string{"td-h3plus", "td-h3plusham", "td-h3plusgmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo canali (non le impostazioni). Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. " +
			"Si presenta come la TD-H3, ma ha tre potenze: usa queste schede e non quelle della TD-H3. Non è in CHIRP." + cps425},
	{ID: "tid-td-h7plus", Brand: "TIDRADIO", Name: "TD-H7 Plus", Aliases: []string{"H7 Plus", "TDH7 Plus"}, Models: []string{"TD-H7 Plus"},
		Type: "analog", Drivers: []string{"td-h7plus", "td-h7plusham", "td-h7plusgmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo canali (non le impostazioni). Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. Non è in CHIRP." + cps425},
	{ID: "tid-td-h8g4", Brand: "TIDRADIO", Name: "TD-H8 (4ª generazione)", Aliases: []string{"TD-H8 4th Gen", "H8-4rd", "TD-H8 4"}, Models: []string{"TD-H8 4ª gen."},
		Type: "analog", Drivers: []string{"td-h8g4", "td-h8g4ham", "td-h8g4gmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo canali (non le impostazioni). Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. " +
			"Per le TD-H8 delle prime tre generazioni usa la scheda «TD-H8». Non è in CHIRP." + cps425},
	{ID: "tid-td-h9", Brand: "TIDRADIO", Name: "TD-H9", Aliases: []string{"H9", "TDH9"}, Models: []string{"TD-H9"},
		Type: "analog", Drivers: []string{"td-h9", "td-h9ham", "td-h9gmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo canali (non le impostazioni). Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. Non è in CHIRP." + cps425},
	{ID: "tid-td-638uv", Brand: "TIDRADIO", Name: "TD-638UV", Aliases: []string{"638UV", "TD638UV"}, Models: []string{"TD-638UV"},
		Type: "analog", Drivers: []string{"td-638uv", "td-638uvham", "td-638uvgmrs"}, Links: []catLink{info("Programmi TIDRADIO", tidDown)},
		Notes: "199 canali, solo canali (non le impostazioni). Tre schede: base, HAM e GMRS; leggendo la radio il programma riconosce la versione. Non è in CHIRP." + cps425},
}
