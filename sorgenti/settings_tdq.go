// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali (4.30) della famiglia «tdq» (drv_tdq.go): la riga di 16
// byte all'indirizzo 0xC80, come la leggono e la scrivono i programmi «TD-FRS» /
// «TD-PMR» e «TD-M2Pro» (un indice per voce; le scelte vengono dalle risorse dei
// programmi). Non sono offerte le voci di cui non è chiaro l'effetto («Display
// Time», «Rx Only», «DCS Normal/Special») né il canale corrente.

import "strconv"

// tdqDefs: voci del modello. short: il programma mostra anche la pressione breve
// dei tasti laterali; m2: elenco delle funzioni del programma «TD-M2Pro».
func tdqDefs(short, m2 bool) []tdcDef {
	tot := []string{"OFF"}
	for s := 15; s <= 600; s += 15 {
		tot = append(tot, strconv.Itoa(s)+" s")
	}
	keys := []string{"Nessuna", "Livello squelch", "Scansione", "Cambio di potenza", "VOX", "Blocco del canale", "Chiamata 1", "Chiamata 2",
		"Scansione delle frequenze", "Bluetooth", "Meteo", "Riduzione del rumore"}
	vals := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12}
	keysS, keysL, valsSL := keys, keys, vals
	if m2 { // il programma del TD-M2 Pro ha anche la copia via radio e non ha il monitor
		keysS = append(append([]string{}, keys[:11]...), "Copia via radio", keys[11])
		keysL, valsSL = keysS, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	} else {
		keysS = append(append([]string{}, keys...), "Monitor (squelch aperto)")
		keysL = append(append([]string{}, keys...), "Monitor finché tieni premuto")
		valsSL = append(append([]byte{}, vals...), 14)
	}
	kd := tdcKeyDsc
	a := tdqSetAddr
	d := []tdcDef{
		{key: "squelch", addr: a + 1, mask: 0xFF, opts: strNums(10, 0, 1, ""), group: tdcG1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", addr: a + 2, mask: 0xFF, opts: tot, group: tdcG1, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "tot_alert", addr: a + 3, mask: 0xFF, opts: append([]string{"OFF"}, strNums(10, 1, 1, " s")...), group: tdcG1, label: "Preavviso del tempo massimo",
			dsc: "Quanti secondi prima della fine del tempo massimo di trasmissione la radio avvisa con un segnale acustico."},
		{key: "vox", addr: a + 4, mask: 0xFF, group: tdcG1, label: "VOX (trasmissione a voce)", dsc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", warn: wVOX},
		{key: "vox_level", addr: a + 5, mask: 0x07, opts: strNums(5, 1, 1, ""), group: tdcG1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX)."},
		{key: "vox_delay", addr: a + 6, mask: 0xFF, opts: []string{"0,5 s", "1 s", "1,5 s", "2 s", "2,5 s", "3 s"}, group: tdcG1, label: "Ritardo del VOX",
			dsc: "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare."},
		{key: "mic_gain", addr: a, mask: 0xF0, opts: strNums(10, 0, 1, ""), group: tdcG1, label: "Guadagno del microfono",
			dsc: "Sensibilità del microfono: più alto il numero, più forte l'audio trasmesso."},
		{key: "save", addr: a + 9, mask: 0xFF, group: tdcG1, label: "Risparmio della batteria",
			dsc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi."},
		{key: "voice", addr: a, mask: 0x03, opts: []string{"OFF", "Cinese", "Inglese"}, group: tdcG2, label: "Guida vocale",
			dsc: "La radio annuncia a voce i menu e i canali."},
		{key: "beep", addr: a + 8, mask: 0xFF, group: tdcG2, label: "Suono dei tasti", dsc: "Beep alla pressione dei tasti."},
	}
	if short {
		d = append(d, tdcDef{key: "key1_short", addr: a + 11, mask: 0xFF, vals: valsSL, opts: keysS, group: tdcG3, label: "Tasto laterale 1, pressione breve", dsc: kd})
	}
	d = append(d, tdcDef{key: "key1_long", addr: a + 12, mask: 0xFF, vals: valsSL, opts: keysL, group: tdcG3, label: "Tasto laterale 1, pressione lunga", dsc: kd})
	if short {
		d = append(d, tdcDef{key: "key2_short", addr: a + 13, mask: 0xFF, vals: valsSL, opts: keysS, group: tdcG3, label: "Tasto laterale 2, pressione breve", dsc: kd})
	}
	return append(d, tdcDef{key: "key2_long", addr: a + 14, mask: 0xFF, vals: valsSL, opts: keysL, group: tdcG3, label: "Tasto laterale 2, pressione lunga", dsc: kd})
}

var tdqSettings = map[string]*setTable{}

func init() {
	for _, m := range genModels {
		o, ok := m.Opt.(*tdqOpt)
		if !ok {
			continue
		}
		// nel file «models.json» del programma i modelli a due potenze non mostrano la pressione breve
		tdqSettings[m.ID] = tdcBuild("TDQ-"+m.ID, tdqSize, 0, "programma del costruttore "+o.prog, tdqDefs(len(o.pow) == 3, o.magic == "QROGRAO"))
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return tdqSettings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			if t := tdqSettings[kind]; t != nil && !tdqBlankSet(img) {
				return t
			}
			return nil // immagine mai letta da una radio: non ci sono impostazioni da mostrare
		},
	})
}
