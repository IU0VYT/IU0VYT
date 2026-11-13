// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni (4.26) delle Baofeng BF-T17 / BF-T18 (PMR446 e FRS), GT-18 PMR446
// e Radioddity FS-T3 / PR-T3: blocchi 0x630 e 0x640 della memoria, come li
// leggono e li scrivono readInfo e writeInfo dei programmi dei costruttori
// (classe AnaCps.Data.DataProtocol di «T17-T18 PMR446 PRO» e «T17-T18 FRS»
// v1.2.2, «BAOFENG GT18 PMR» v1.2.1, «Radioddity T3» v1.3: stesse posizioni in
// tutti e quattro), con le scelte della finestra delle funzioni di sistema.
//
//   0x630  guida vocale (0/1)            0x640  bit 0 suono dei tasti
//   0x631  lingua della guida vocale            bit 1 risparmio della batteria
//   0x632  scansione (0/1)                      bit 2 tono di fine trasmissione (PMR)
//   0x633  VOX (0/1)                            bit 7 tono di fine trasmissione (T17 FRS)
//   0x634  livello del VOX (0-4 = 1-5)   0x641  squelch (0-9)
//   0x635  tono di chiamata              0x642  modalità dell'allarme
//   0x636  blocco TX a batteria scarica  0x643  tempo massimo di trasmissione
//   0x637  blocco TX a tensione alta     0x644  eliminazione della coda di rumore
//
// Ogni voce occupa un byte intero, tranne quelle del byte 0x640. La FS-T3 /
// PR-T3 non ha il tono di fine trasmissione. Le scelte non vengono riempite
// all'avvio dei programmi (a differenza della potenza dei canali, vedi
// drv_bfd.go): l'ordine è quello delle risorse delle finestre.
//
// 4.32: programmi «T20 FRS» v1.0.1 (Baofeng T22 in versione FRS, TIDRADIO
// TD-M8S) e «CPS_F8_FRS» v1.0 (TIDRADIO TD-V2): stessa finestra e stesse voci
// nei due, tre blocchi invece di due (a = 0x2C0 nel primo, 0x3C0 nel secondo):
//
//   0x2B0  guida vocale (0/1)              0x2B8  «Rx Emergency» (0/1)
//   0x2B1  lingua della guida vocale       a+0    bit 0 suono dei tasti
//   0x2B2  scansione (0/1)                        bit 1 risparmio della batteria
//   0x2B3  VOX (0/1)                       a+1    squelch (0-9)
//   0x2B4  livello del VOX (1-5)           a+2    tasto laterale
//   0x2B5  VOX spento in ricezione         a+3    tempo massimo di trasmissione
//   0x2B6  blocco TX a tensione alta
//   0x2B7  blocco TX a tensione bassa
//
// (le due voci del blocco della trasmissione sono in ordine inverso rispetto
// alle T17). Il livello del VOX è il numero stesso (casella numerica 1-5).

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

type t17SetDef struct {
	key               string
	off               int
	mask              byte
	opts              []string // vuoto: OFF/ON
	first             int      // valore della prima scelta (0 di solito)
	group, label, dsc string
	warn              string
}

// t17SetDefsFor: txEnd = maschera del tono di fine trasmissione (0 = non c'è)
func t17SetDefsFor(txEnd byte) []t17SetDef {
	g1, g2, g3 := "Ricezione e trasmissione", "Display, suoni e tasti", "DTMF, allarme e funzioni remote"
	d := []t17SetDef{
		{key: "sql", off: 0x641, opts: strNums(10, 0, 1, ""), group: g1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: 0x643, opts: append([]string{"OFF"}, strNums(10, 30, 30, " s")...), group: g1, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "vox_on", off: 0x633, group: g1, label: "VOX (trasmissione a voce)",
			dsc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", warn: wVOX},
		{key: "vox_level", off: 0x634, opts: strNums(5, 1, 1, ""), group: g1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX)."},
		{key: "tail", off: 0x644, group: g1, label: "Eliminazione della coda di rumore",
			dsc: "Elimina il breve rumore che si sente quando l'altra radio smette di trasmettere (tra radio con la stessa funzione)."},
		{key: "r_tone", off: 0x635, opts: []string{"1450 Hz", "1750 Hz", "2100 Hz"}, group: g1, label: "Tono di chiamata per i ponti",
			dsc: "Frequenza del tono di apertura inviato ai ponti che lo richiedono (in Europa di solito 1750 Hz)."},
		{key: "scan_on", off: 0x632, group: g1, label: "Scansione",
			dsc: "Funzione di scansione dei canali accesa o spenta."},
		{key: "low_tx", off: 0x636, group: g1, label: "Blocco della trasmissione con tensione bassa",
			dsc: "Impedisce di trasmettere quando la tensione della batteria è troppo bassa (protegge la radio).", warn: wVolt},
		{key: "high_tx", off: 0x637, group: g1, label: "Blocco della trasmissione con tensione alta",
			dsc: "Impedisce di trasmettere quando la tensione della batteria è troppo alta (protegge la radio).", warn: wVolt},

		{key: "voice", off: 0x630, group: g2, label: "Guida vocale", dsc: "La radio annuncia a voce i menu e i canali."},
		{key: "lang", off: 0x631, opts: []string{"Inglese", "Cinese"}, group: g2, label: "Lingua", dsc: "Lingua dei menu e/o della guida vocale."},
		{key: "beep", off: 0x640, mask: 0x01, group: g2, label: "Suono dei tasti", dsc: "Beep alla pressione dei tasti."},
		{key: "save", off: 0x640, mask: 0x02, group: g2, label: "Risparmio della batteria",
			dsc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi."},

		{key: "alarm_mode", off: 0x642, opts: []string{"Solo sulla radio", "Trasmette il suono"}, group: g3, label: "Modalità dell'allarme (SOS)",
			dsc: "Funzionamento del tasto di allarme (solo sulla radio o con trasmissione di un suono o di un codice).", warn: wAlarm},
	}
	if txEnd != 0 {
		d = append(d, t17SetDef{key: "tx_end", off: 0x640, mask: txEnd, group: g1, label: "Segnale di fine trasmissione (roger beep)",
			dsc: "Suono inviato in aria quando rilasci il PTT."})
	}
	for i := range d {
		if d[i].mask == 0 {
			d[i].mask = 0xFF
		}
	}
	return d
}

// anaSetDefs (4.32): voci dei programmi «T20 FRS» e «CPS_F8_FRS»; a = indirizzo
// del blocco con suoni, squelch, tasto laterale e TOT
func anaSetDefs(a int) []t17SetDef {
	g1, g2, g3 := "Ricezione e trasmissione", "Display, suoni e tasti", "DTMF, allarme e funzioni remote"
	t := t17SetDefsFor(0)
	pick := func(key string, off int, mask byte) t17SetDef {
		for _, d := range t {
			if d.key == key {
				d.off, d.mask = off, mask
				return d
			}
		}
		panic(key)
	}
	vl := pick("vox_level", 0x2B4, 0xFF)
	vl.first = 1
	return []t17SetDef{
		pick("sql", a+1, 0xFF), pick("tot", a+3, 0xFF), pick("vox_on", 0x2B3, 0xFF), vl,
		{key: "vox_rx", off: 0x2B5, mask: 0xFF, group: g1, label: "VOX disattivato durante la ricezione",
			dsc: "Il VOX non fa trasmettere mentre la radio sta ricevendo un segnale."},
		pick("scan_on", 0x2B2, 0xFF), pick("low_tx", 0x2B7, 0xFF), pick("high_tx", 0x2B6, 0xFF),
		pick("voice", 0x2B0, 0xFF), pick("lang", 0x2B1, 0xFF), pick("beep", a, 0x01), pick("save", a, 0x02),
		{key: "side_key", off: a + 2, mask: 0xFF, opts: []string{"Nessuna", "Monitor", "Cambio di potenza", "Allarme"}, group: g2,
			label: "Tasto laterale", dsc: "Funzione del tasto laterale."},
		{key: "rx_emerg", off: 0x2B8, mask: 0xFF, group: g3, label: "Allarme («Rx Emergency»)",
			dsc: "Accende o spegne la funzione di allarme: nel programma del costruttore la voce si chiama «Rx Emergency» in inglese e «interruttore dell'allarme» in cinese."},
	}
}

func t17SetTable(defs []t17SetDef, src string) *setTable {
	return t17SetTableSize(defs, src, t17Memsize)
}

func t17SetTableSize(defs []t17SetDef, src string, size int) *setTable {
	t := &setTable{Class: "AnaCps", Size: size, Source: src}
	for _, b := range defs {
		sh := 0
		for b.mask>>sh&1 == 0 {
			sh++
		}
		val := func(i int) *string {
			s := hex.EncodeToString([]byte{byte((i+b.first)<<sh) & b.mask})
			return &s
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(b.off), hex.EncodeToString([]byte{b.mask})}}}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.dsc, ItWarn: b.warn, Enc: enc}
		if len(b.opts) == 0 {
			sd.Type = "bool"
			enc.Table = []*string{val(0), val(1)}
		} else {
			sd.Type = "list"
			sd.ItOpts = b.opts
			for i := range b.opts {
				sd.Options = append(sd.Options, strconv.Itoa(i))
				enc.Table = append(enc.Table, val(i))
			}
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

// t17Settings: tabella per modello (maschera del tono di fine trasmissione e programma)
var t17Settings = map[string]*setTable{
	"bf-t17pmr":  t17SetTable(t17SetDefsFor(0x04), t17Src("«T17-T18 PMR446 PRO» v1.2.2")),
	"bf-t17frs":  t17SetTable(t17SetDefsFor(0x80), t17Src("«T17-T18 FRS» v1.2.2")),
	"bf-gt18pmr": t17SetTable(t17SetDefsFor(0x04), t17Src("«BAOFENG GT18 PMR» v1.2.1")),
	"rd-t3":      t17SetTable(t17SetDefsFor(0), t17Src("«Radioddity T3» v1.3")),
	"bf-t22frs":  t17SetTableSize(anaSetDefs(0x2C0), t17Src("«T20 FRS» v1.0.1"), 0x340),
	"td-m8s":     t17SetTableSize(anaSetDefs(0x2C0), t17Src("«T20 FRS» v1.0.1"), 0x340),
	"td-v2":      t17SetTableSize(anaSetDefs(0x3C0), t17Src("«CPS_F8_FRS» v1.0"), bfdMemsize),
}

func t17Src(prog string) string { return fmt.Sprintf("programma del costruttore %s", prog) }

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return t17Settings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := t17Settings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
}
