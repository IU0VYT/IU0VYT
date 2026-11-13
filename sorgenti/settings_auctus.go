// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni (4.26) delle Retevis RB39 / RB39P e RB639 / RB639P: i dati di
// sistema che i programmi del costruttore («RB39 CPS», «Retevis Analog CPS»,
// classe Auctus_V1._0.Data.WriFreq) leggono con il comando 04/01 e riscrivono
// con 04/00, un byte per voce, nell'ordine:
//
//   0  tempo massimo di trasmissione, in decine di secondi (6, 12, 18)
//   1  livello del VOX (0 = spento, 1-5)
//   2  ritardo del VOX in mezzi secondi (1-6)
//   3  guida vocale (0 spenta, 2 inglese; 1, cinese, non è offerto)
//   4  compander (0/1)
//   5  VOX (0/1)
//   6  monitor (0/1)
//   7  funzione del tasto laterale tenuto premuto (0-3)
//   8  risparmio della batteria (0/1)
//   9  Bluetooth (0/1), solo nelle versioni «P», che mandano dieci byte
//
// Le scelte sono quelle della finestra «SystemInfo» (l'unica che il programma
// apre: «FormSystemInfo», con altre scelte, non è usata), che lega ogni casella
// al valore e non alla posizione nell'elenco. Nell'immagine i dati stanno dopo
// il byte della lunghezza (vedi drv_auctus.go).

import "encoding/hex"

const auctusSrc = "programmi del costruttore «RB39 CPS» e «Retevis Analog CPS» (finestra «System Information»)"

type auctusSetDef struct {
	key               string
	idx               int
	vals              []byte   // valore nella radio di ogni scelta (vuoto: 0 e 1)
	opts              []string // vuoto: casella sì/no
	group, label, dsc string
	warn              string
}

var auctusSetDefs = func() []auctusSetDef {
	g1, g2 := "Ricezione e trasmissione", "Display, suoni e tasti"
	return []auctusSetDef{
		{key: "tot", idx: 0, vals: []byte{6, 12, 18}, opts: []string{"60 s", "120 s", "180 s"}, group: g1, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola."},
		{key: "vox_on", idx: 5, group: g1, label: "VOX (trasmissione a voce)",
			dsc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", warn: wVOX},
		{key: "vox_level", idx: 1, vals: []byte{0, 1, 2, 3, 4, 5}, opts: append([]string{"OFF"}, strNums(5, 1, 1, "")...), group: g1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX)."},
		{key: "vox_delay", idx: 2, vals: []byte{1, 2, 3, 4, 5, 6}, opts: []string{"0,5 s", "1 s", "1,5 s", "2 s", "2,5 s", "3 s"}, group: g1, label: "Ritardo del VOX",
			dsc: "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare (evita di tagliare la fine delle frasi)."},
		{key: "compander", idx: 4, group: g1, label: "Compander",
			dsc: "Comprime la voce in trasmissione e la espande in ricezione per ridurre il rumore di fondo: funziona bene solo se anche le altre radio lo usano."},
		{key: "monitor", idx: 6, group: g1, label: "Monitor (squelch aperto)",
			dsc: "Tiene aperto lo squelch: si sente tutto quello che c'è sul canale, anche il fruscio e i segnali senza il tono giusto."},
		{key: "voice", idx: 3, vals: []byte{0, 2}, opts: []string{"Spenta", "Inglese"}, group: g2, label: "Guida vocale",
			dsc: "La radio annuncia a voce i menu e i canali."},
		{key: "pf_long", idx: 7, vals: []byte{0, 1, 2, 3}, opts: []string{"Nessuna", "Allarme", "Monitor", "Compander"}, group: g2, label: "Tasto laterale (pressione lunga)",
			dsc: "Funzione del tasto laterale tenuto premuto."},
		{key: "save", idx: 8, group: g2, label: "Risparmio della batteria",
			dsc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi."},
		{key: "bt", idx: 9, group: g2, label: "Bluetooth",
			dsc: "Accende o spegne il Bluetooth della radio (solo RB39P e RB639P)."},
	}
}()

// auctusSetTable: n = numero di byte di sistema che la radio ha mandato
func auctusSetTable(n int) *setTable {
	t := &setTable{Class: "Auctus", Size: auctusSize, Source: auctusSrc}
	for _, b := range auctusSetDefs {
		if b.idx >= n {
			continue
		}
		vals := b.vals
		if len(b.opts) == 0 {
			vals = []byte{0, 1}
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(auctusSys + 1 + b.idx), "ff"}}}
		for _, v := range vals {
			s := hex.EncodeToString([]byte{v})
			enc.Table = append(enc.Table, &s)
		}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.dsc, ItWarn: b.warn, Enc: enc, Type: "bool"}
		if len(b.opts) > 0 {
			sd.Type = "list"
			sd.ItOpts = b.opts
			sd.Options = strNums(len(b.opts), 0, 1, "")
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

var auctusSet9, auctusSet10 = auctusSetTable(9), auctusSetTable(10)

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool {
			m, ok := genByID[kind]
			if !ok {
				return false
			}
			_, ok = m.Fam.(auctusFam)
			return ok
		},
		get: func(kind string, img []byte) *setTable {
			if len(img) < auctusSize {
				return nil
			}
			switch n := int(img[auctusSys]); {
			case n >= 10 && n <= 15:
				return auctusSet10
			case n == 9:
				return auctusSet9
			}
			return nil // immagine senza dati di sistema
		},
	})
}
