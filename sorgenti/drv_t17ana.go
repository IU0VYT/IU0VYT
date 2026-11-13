// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.32: Baofeng T22 in versione FRS, TIDRADIO TD-M8S e TIDRADIO TD-V2 (famiglia
// t17Fam: protocollo e memoria sono descritti al punto 5 di drv_bfd.go).
//
// Sono registrati qui e non in drv_bfd.go perché l'ordine di registrazione è
// l'ordine in cui «Identifica» prova i modelli con lo stesso saluto: questi
// tre hanno il saluto «\x02PROGRAM» della BF-888S (drv_small.go) e devono venire
// dopo, così la prima lettura resta quella del driver della BF-888S, che è
// quello provato dal vero (CHIRP). I file si caricano in ordine alfabetico.

func init() {
	pw := []genPower{{"Bassa", 0.5, 2}, {"Alta", 0.5, 0}}
	cod := pmrCodec{base: 0, newFlags: 0x01, newTail: []byte{0xFF, 0xFF, 0xFF}}
	// programma «T20 FRS» v1.0.1 (Baofeng T22 in versione FRS e TIDRADIO
	// TD-M8S: lo stesso file) e «CPS_F8_FRS» v1.0 (TIDRADIO TD-V2)
	t20 := t17Opt{cmd: "PROGRAM", pmr: true, noJump: true, pwAddr: 0x2D8, rblk: 8, wblk: 8, size: 0x340, codec: cod,
		gaps:  [][2]int{{0x000, 0x160}, {0x2B0, 0x2C8}, {0x330, 0x338}},
		wgaps: [][2]int{{0x000, 0x160}, {0x2B0, 0x2D0}}, ffill: [][2]int{{0x2C8, 0x2D0}}}
	v2cod := cod
	v2cod.base = 0x10
	v2 := t17Opt{cmd: "PROGRAM", pmr: true, noJump: true, pwAddr: 0x130, rblk: 8, wblk: 8, size: bfdMemsize, codec: v2cod,
		gaps:  [][2]int{{0x000, 0x110}, {0x2B0, 0x2C0}, {0x330, 0x340}, {0x380, 0x3E0}},
		wgaps: [][2]int{{0x000, 0x110}, {0x2B0, 0x2C0}, {0x380, 0x3E0}}}
	ana := func(id, vendor, model, name string, slots int, o t17Opt, notes string) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: "IU0VYT_ANACPS_" + id, Name: name, Slots: slots, First: 1,
			Powers: pw, Bands: []band{{400000000, 480000000}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
			DCS: dtcsAll(), Baud: 9600, Fam: t17Fam{}, Opt: &oo, Notes: notes}
	}
	t20Note := "Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore «T20 FRS» (v1.0.1, lo stesso che Baofeng dà per la T22 " +
		"in versione FRS e TIDRADIO per la TD-M8S); se sulla radio è impostata la password di lettura e scrittura, toglila prima con quel programma."
	genRegister(
		ana("bf-t22frs", "Baofeng", "T22 FRS", "Baofeng T22 (FRS, USA)", 22, t20, t20Note),
		ana("td-m8s", "TIDRADIO", "TD-M8S", "TIDRADIO TD-M8S", 22, t20, t20Note),
		ana("td-v2", "TIDRADIO", "TD-V2", "TIDRADIO TD-V2", 16, v2,
			"Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore «CPS_F8_FRS» (v1.0); "+
				"se sulla radio è impostata la password di lettura e scrittura, toglila prima con quel programma."),
	)
}
