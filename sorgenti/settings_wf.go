// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali (4.28) di Radioddity UB15-G, DB-35G, DB35-UV e DB40-G: la
// riga letta all'indirizzo E0 00 (drv_wf.go). Posizione e larghezza di ogni
// campo vengono dalla sezione «param» del file di configurazione del programma
// del costruttore, le scelte dal suo file delle lingue; un campo «check» largo
// 4 bit vale 0 o 1.

import "encoding/hex"

type wfSetDef struct {
	key   string
	off   int
	shift uint
	bits  uint
	opts  []string // vuoto: casella sì/no
	group string
	label string
	dsc   string
	warn  string
}

const (
	wfG1 = "Ricezione e trasmissione"
	wfG2 = "Display, suoni e tasti"
)

func wfSetCommon(nrc wfSetDef) []wfSetDef {
	return []wfSetDef{
		{key: "sql", off: 0, bits: 4, opts: strNums(10, 0, 1, ""), group: wfG1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: 1, bits: 3, opts: []string{"OFF", "60 s", "120 s", "180 s"}, group: wfG1, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		nrc,
		{key: "pri_watch", off: 1, shift: 5, bits: 1, group: wfG1, label: "Ascolto del canale prioritario",
			dsc: "La radio controlla ogni tanto i canali segnati come prioritari mentre è su un altro canale."},
		{key: "beep", off: 1, shift: 3, bits: 1, group: wfG2, label: "Suono dei tasti", dsc: "Beep alla pressione dei tasti."},
		{key: "roger", off: 1, shift: 4, bits: 1, group: wfG2, label: "Segnale di fine trasmissione (roger beep)", dsc: "Suono inviato in aria quando rilasci il PTT."},
	}
}

var (
	wfNrcOnOff = wfSetDef{key: "nrc", off: 0, shift: 4, bits: 4, group: wfG1, label: "Riduzione del rumore", dsc: "Riduce il rumore di fondo dell'audio ricevuto."}
	wfNrcLevel = wfSetDef{key: "nrc_level", off: 0, shift: 4, bits: 4, opts: strNums(5, 1, 1, ""), group: wfG1, label: "Livello della riduzione del rumore",
		dsc: "Quanto viene ridotto il rumore di fondo dell'audio ricevuto: più alto il livello, più forte la riduzione."}
)

func wfSetUB15() []wfSetDef { return wfSetCommon(wfNrcOnOff) }

func wfSetUB15v110() []wfSetDef {
	return append(wfSetCommon(wfNrcOnOff),
		wfSetDef{key: "anc", off: 2, bits: 4, group: wfG1, label: "Riduzione del rumore del microfono",
			dsc: "Riduce il rumore dell'ambiente raccolto dal microfono mentre trasmetti."},
		wfSetDef{key: "vox_level", off: 2, shift: 4, bits: 4, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...), group: wfG1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX).", warn: wVOX})
}

func wfSetDB35G() []wfSetDef { return wfSetCommon(wfNrcLevel) }

func wfSetDB35UV() []wfSetDef {
	return append(wfSetCommon(wfNrcLevel),
		wfSetDef{key: "vox_level", off: 2, bits: 3, opts: strNums(5, 1, 1, ""), group: wfG1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX)."},
		wfSetDef{key: "step", off: 2, shift: 3, bits: 3, opts: []string{"2,5 kHz", "5 kHz", "6,25 kHz", "12,5 kHz", "25 kHz", "50 kHz", "100 kHz"}, group: wfG1, label: "Passo di frequenza",
			dsc: "Di quanto cambia la frequenza a ogni scatto quando la radio è in modo VFO."},
		wfSetDef{key: "bl_level", off: 2, shift: 6, bits: 2, opts: []string{"OFF", "1", "2", "3"}, group: wfG2, label: "Retroilluminazione",
			dsc: "Luminosità della retroilluminazione del display."},
		wfSetDef{key: "bl_time", off: 3, bits: 2, opts: []string{"Sempre accesa", "10 s", "20 s", "30 s"}, group: wfG2, label: "Spegnimento della retroilluminazione",
			dsc: "Dopo quanto tempo senza toccare la radio la retroilluminazione si spegne."},
		wfSetDef{key: "bl_color", off: 3, shift: 2, bits: 3, opts: []string{"Bianco", "Blu", "Verde", "Rosso", "Ciano", "Viola", "Giallo"}, group: wfG2, label: "Colore della retroilluminazione",
			dsc: "Colore della luce del display."})
}

func wfSetDB40G() []wfSetDef {
	return []wfSetDef{
		{key: "sql", off: 3, bits: 8, opts: strNums(10, 0, 1, ""), group: wfG1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: 2, bits: 8, opts: []string{"OFF", "30 s", "60 s", "90 s", "120 s", "150 s", "180 s", "210 s", "240 s", "270 s", "300 s"}, group: wfG1,
			label: "Tempo massimo di trasmissione (TOT)", dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "bcl", off: 0, bits: 8, opts: []string{"OFF", "ON"}, group: wfG1, label: "Blocco su canale occupato",
			dsc: "Se attivo, la radio non trasmette mentre sul canale c'è già un segnale."},
		{key: "beep", off: 1, bits: 8, opts: []string{"OFF", "ON"}, group: wfG2, label: "Suono dei tasti", dsc: "Beep alla pressione dei tasti."},
	}
}

func wfSetTable(m *genModel) *setTable {
	o := wfO(m)
	t := &setTable{Class: "WF-" + m.ID, Size: o.size(m), Source: "programma del costruttore " + o.prog}
	for _, b := range o.set() {
		mask := byte(1<<b.bits-1) << b.shift
		val := func(v int) *string {
			s := hex.EncodeToString([]byte{byte(v) << b.shift & mask})
			return &s
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(o.param(m) + b.off), hex.EncodeToString([]byte{mask})}}}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.dsc, ItWarn: b.warn, Enc: enc, Type: "bool"}
		n := 2
		if len(b.opts) > 0 {
			n = len(b.opts)
			sd.Type, sd.ItOpts = "list", b.opts
			sd.Options = strNums(n, 0, 1, "")
		}
		for v := 0; v < n; v++ {
			enc.Table = append(enc.Table, val(v))
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

var wfSettings = map[string]*setTable{}

func init() {
	for _, id := range []string{"rd-ub15g", "rd-ub15g-110", "rd-db35g", "rd-db35uv", "rd-db40g"} {
		wfSettings[id] = wfSetTable(genByID[id])
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return wfSettings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := wfSettings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			if wfBlank(img[t.Size-wfO(genByID[kind]).page : t.Size]) {
				return nil // immagine mai letta da una radio: non ci sono impostazioni da mostrare
			}
			return t
		},
	})
}
