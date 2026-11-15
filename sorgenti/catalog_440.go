// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.40: Alinco (marca nuova), drv_alinco.go: palmari DJ-596, DJ-175 e
// DJ-G7 (tribanda), veicolari DR-135T, DR-235T, DR-435T, DR-03T e DR-06T.

const alincoWeb = "https://www.alinco.com/"

var catalog440 = func() []catItem {
	it := func(id, name string, aliases []string, drv, cls, note string) catItem {
		return catItem{ID: id, Brand: "Alinco", Name: name, Aliases: aliases, Type: "analog", Drivers: []string{drv},
			Links: []catLink{info("Sito Alinco", alincoWeb)}, Notes: chirpNoteCh(cls) + note}
	}
	const c100 = " 100 canali con nomi di 7 caratteri."
	const mob = " Radio veicolare:" + c100
	return []catItem{
		it("al-dj-596", "DJ-596", []string{"DJ-596T", "DJ-596E", "DJ596"}, "al-dj596", "DJ596Radio", " Palmare bibanda 144/430 MHz:"+c100),
		it("al-dj-175", "DJ-175", []string{"DJ-175T", "DJ-175E", "DJ175"}, "al-dj175", "DJ175Radio", " Palmare VHF:"+c100),
		it("al-dj-g7eg", "DJ-G7EG", []string{"DJ-G7", "DJ-G7E"}, "al-djg7eg", "AlincoDJG7EG", alg7CatNote),
		it("al-dj-g7t", "DJ-G7T", nil, "al-djg7t", "AlincoDJG7T", alg7CatNote),
		it("al-dr-135t", "DR-135T", []string{"DR-135", "DR135T"}, "al-dr135t", "DR135Radio", mob),
		it("al-dr-235t", "DR-235T", []string{"DR-235", "DR235T"}, "al-dr235t", "DR235Radio", mob),
		it("al-dr-435t", "DR-435T", []string{"DR-435", "DR435T"}, "al-dr435t", "DR435Radio", mob),
		it("al-dr-03t", "DR-03T", []string{"DR-03", "DR03T"}, "al-dr03t", "DR03Radio", mob),
		it("al-dr-06t", "DR-06T", []string{"DR-06", "DR06T"}, "al-dr06t", "DR06Radio", mob),
	}
}()

const alg7CatNote = " Palmare tribanda 144/430/1200 MHz con ricevitore a larga banda: 1000 canali con nomi di 16 caratteri. Serve il firmware 4_10 o successivo."
