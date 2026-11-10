// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: impostazioni principali del firmware nicFW 2.5x della TIDRADIO TD-H3
// (blocco «settingsBlock» a 0x1900, dalla documentazione dell'autore:
// https://github.com/nicsure/nicfw2docs/blob/main/settingsBlock.md). Solo le
// voci di un byte il cui significato è documentato; le versioni 2.0x hanno un
// blocco diverso e qui non si modificano.

import "encoding/binary"

var nicfwSetTable = func() *setTable {
	gOp, gTx, gRx, gDisp := "Funzionamento", "Trasmissione", "Ricezione e audio", "Display"
	at := func(off int) int { return nicfwSetAt + off }
	d := []ijvSetDef{
		{off: at(0x02), opts: ijvNums(0, 10, 1, ""), key: "squelch", group: gRx, label: "Squelch"},
		{off: at(0x03), key: "dual_watch", group: gRx, label: "Doppio ascolto (Dual Watch)"},
		{off: at(0x18), opts: []string{"OFF", "Mostra il tono ricevuto", "Copia il tono ricevuto in trasmissione"}, key: "tone_monitor", group: gRx,
			label: "Controllo del tono ricevuto"},
		{off: at(0x0E), opts: ijvNums(0, 32, 1, ""), key: "mic_gain", group: gTx, label: "Guadagno del microfono"},
		{off: at(0x52), opts: append([]string{"OFF"}, ijvNums(1, 15, 1, "")...), key: "vox", group: gTx, label: "VOX (trasmissione a voce)"},
		{off: at(0x11), opts: []string{"OFF", "Icona", "Percentuale", "Tensione"}, key: "batt_style", group: gDisp, label: "Indicatore della batteria"},
		{off: at(0x19), opts: ijvNums(0, 29, 1, ""), key: "lcd_brightness", group: gDisp, label: "Luminosità dello schermo"},
		{off: at(0x5F), key: "lcd_inverted", group: gDisp, label: "Colori dello schermo invertiti"},
		{off: at(0x46), key: "key_lock", group: gOp, label: "Blocco dei tasti"},
		{off: at(0x47), key: "bluetooth", group: gOp, label: "Bluetooth"},
		{off: at(0x48), key: "power_save", group: gOp, label: "Risparmio della batteria"},
		{off: at(0x49), key: "key_tones", group: gOp, label: "Toni dei tasti"},
	}
	return setTableAdd(&setTable{Class: "nicFW-2.5", Size: nicfwSize,
		Source: "documentazione pubblicata dall'autore del firmware nicFW (nicsure/nicfw2docs)"}, d)
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "td-tdh3-nicfw" },
		get: func(kind string, img []byte) *setTable {
			if len(img) != nicfwSize || binary.BigEndian.Uint16(img[nicfwSetAt:]) != nicfwMagic5 {
				return nil
			}
			return nicfwSetTable
		},
	})
}
