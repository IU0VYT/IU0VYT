// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.13: impostazioni principali del firmware IJV (area fissa 0x0000-0x01FF della
// memoria, nomi dei gruppi e dei banchi). Posizioni e valori ricavati dalla
// tabella delle impostazioni di RadioManager 4.3.2 (V4 e FW5) e dal modulo CHIRP
// 3.60 (V3). Le tarature (da 0x1E00), la VFO, le selettive, i contatti, i tasti
// della FW5 e la scansione non sono qui: restano a RadioManager.

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

type ijvSetDef struct {
	off                           int
	mask                          byte
	opts                          []string // vuoto: sì/no
	str                           int      // testo di str caratteri (opts e mask non usati)
	key, group, label, desc, warn string
}

func ijvNums(from, n, step int, unit string) []string {
	o := make([]string, n)
	for i := range o {
		o[i] = strconv.Itoa(from+i*step) + unit
	}
	return o
}

func ijvSetDefs(kind ijvKind) []ijvSetDef {
	gOp, gTx, gRx, gDisp, gKeys, gNames := "Funzionamento", "Trasmissione", "Ricezione e audio", "Display", "Tasti laterali", "Nomi"
	v3 := kind == ijvV3 || kind == ijvV3X
	fw5 := kind == ijvFW5 || kind == ijvV6 // la V6 usa le impostazioni della V5 (come RadioManager)
	d := []ijvSetDef{
		{off: 0x02, opts: []string{"OFF", "30 s", "1 min", "3 min", "5 min"}, key: "talk_time", group: gTx, label: "Tempo massimo di trasmissione (TOT)",
			desc: "Dopo questo tempo la trasmissione si interrompe da sola."},
		{off: 0x04, key: "key_lock", group: gOp, label: "Blocco della tastiera"},
		{off: 0x05, key: "vox_switch", group: gTx, label: "VOX (trasmissione a voce)",
			warn: "Con il VOX acceso la radio trasmette appena sente un suono."},
		{off: 0x09, opts: []string{"Frequenza", "Numero del canale", "Nome del canale", "Nome piccolo e frequenza grande", "Nome grande e frequenza piccola"},
			key: "channel_display", group: gDisp, label: "Visualizzazione del canale"},
		{off: 0x0D, opts: []string{"Spenta", "5 s", "10 s", "20 s", "1 min", "3 min", "In ricezione e trasmissione", "Sempre accesa"},
			key: "backlight", group: gDisp, label: "Retroilluminazione"},
		{off: 0x0E, key: "tail_note_elimination", group: gTx, label: "Eliminazione della coda in trasmissione"},
		{off: 0x0F, key: "vfo_hidden", group: gOp, label: "VFO nascosto"},
		{off: 0x12, key: "auto_key_lock", group: gOp, label: "Blocco automatico della tastiera"},
		{off: 0x16, opts: []string{"OFF", "5 s", "10 s", "30 s", "1 min", "3 min", "6 min", "10 min", "20 min"}, key: "beacon", group: gTx,
			label: "Beacon / chiamata CQ", warn: "Con il beacon acceso la radio trasmette da sola a intervalli regolari."},
		{off: 0x17, mask: 0x01, key: "tx_enable", group: gTx, label: "Trasmissione abilitata",
			desc: "Spento: la radio non trasmette (solo ricezione)."},
		{off: 0x1F, mask: 0x30, opts: []string{"OFF", "50 MHz", "125 MHz", "Personalizzato"}, key: "upconverter", group: gRx, label: "Upconverter"},
		{off: 0x24, opts: []string{"Normale", "Invertito"}, key: "lcd_type", group: gDisp, label: "Display (colori)"},
		{off: 0x150, str: 16, key: "logo_line1", group: gDisp, label: "Messaggio all'accensione, riga 1"},
		{off: 0x160, str: 16, key: "logo_line2", group: gDisp, label: "Messaggio all'accensione, riga 2"},
		{off: 0x170, str: 8, key: "qrz", group: gDisp, label: "Nominativo (QRZ)"},
	}
	welcome := []string{"Nessuno", "Nome del firmware", "Messaggio"}
	vfo := []string{"Singolo", "Doppio", "Doppio bloccato (DW LOCK)", "Doppio collegato (DW LINK)", "Split"}
	if fw5 {
		welcome = []string{"Nessuno", "Logo", "Messaggio"}
		vfo = []string{"1 VFO", "2 VFO", "3 VFO", "Doppio bloccato (DW LOCK)", "Doppio collegato (DW LINK)", "Split",
			"Triplo bloccato (TW LOCK)", "Triplo collegato (TW LINK)"}
	}
	d = append(d,
		ijvSetDef{off: 0x13, opts: welcome, key: "power_on_display", group: gDisp, label: "Schermata all'accensione"},
		ijvSetDef{off: 0x14, opts: vfo, key: "vfo_mode", group: gOp, label: "Modo VFO"})
	if v3 {
		keys := []string{"Nessuna", "Torcia", "Potenza di trasmissione", "Monitor", "Scansione", "Radio FM", "Cambio VFO", "Scambio VFO",
			"Squelch +", "Squelch -", "Prova REGA", "Allarme REGA", "Preset", "AGC manuale", "Elenco dei canali", "Scrambler", "Rubrica", "Chiamata CQ in CW"}
		d = append(d,
			ijvSetDef{off: 0x03, opts: ijvNums(0, 10, 1, ""), key: "tx_dev", group: gTx, label: "Deviazione in trasmissione (0 normale, 9 massima)"},
			ijvSetDef{off: 0x07, opts: []string{"+1,1 dB", "+4,0 dB", "+8,0 dB", "+12,0 dB", "+15,1 dB"}, key: "mic_gain", group: gTx, label: "Guadagno del microfono"},
			ijvSetDef{off: 0x08, key: "beep", group: gOp, label: "Bip dei tasti"},
			ijvSetDef{off: 0x0C, opts: ijvNums(0, 8, 1, ""), key: "afc", group: gRx, label: "AFC (controllo automatico della frequenza)"},
			ijvSetDef{off: 0x10, opts: []string{"OFF", "FCC", "CE", "GB", "430", "438"}, key: "flock", group: gTx, label: "Blocco delle frequenze di trasmissione (F-Lock)"},
			ijvSetDef{off: 0x11, opts: []string{"A tempo (5 s)", "Lenta", "Veloce", "Si ferma sul segnale", "Registro"}, key: "scan_resume", group: gRx, label: "Ripresa della scansione"},
			ijvSetDef{off: 0x17, mask: 0x10, key: "micbar", group: gDisp, label: "Barra del microfono"},
			ijvSetDef{off: 0x1F, mask: 0x01, key: "savedirect", group: gOp, label: "Salvataggio automatico"},
			ijvSetDef{off: 0x22, opts: append([]string{"OFF"}, ijvNums(100, 20, 100, " ms")...), key: "rp_ste", group: gRx, label: "Eliminazione della coda del ponte (RP-STE)"},
			ijvSetDef{off: 0x23, opts: []string{"A", "B"}, key: "bands_tx", group: gTx, label: "VFO di trasmissione"},
			ijvSetDef{off: 0x190, opts: keys, key: "side1_short", group: gKeys, label: "Tasto laterale 1, pressione breve"},
			ijvSetDef{off: 0x191, opts: keys, key: "side1_long", group: gKeys, label: "Tasto laterale 1, pressione lunga"},
			ijvSetDef{off: 0x192, opts: keys, key: "side2_short", group: gKeys, label: "Tasto laterale 2, pressione breve"},
			ijvSetDef{off: 0x193, opts: keys, key: "side2_long", group: gKeys, label: "Tasto laterale 2, pressione lunga"})
	} else {
		dev := ijvNums(0, 10, 1, "")
		if fw5 {
			dev = []string{"NFM-", "NFM", "HAM", "HAM+", "FM", "FM+", "AFSK", "AUDIO", "WIDE", "WFM"}
		}
		d = append(d,
			ijvSetDef{off: 0x03, opts: dev, key: "tx_deviation", group: gTx, label: "Deviazione in trasmissione"},
			ijvSetDef{off: 0x06, opts: ijvNums(1, 10, 1, ""), key: "vox_level", group: gTx, label: "Sensibilità del VOX"},
			ijvSetDef{off: 0x07, opts: []string{"+2,0 dB", "+5,0 dB", "+9,0 dB", "+13,0 dB", "+17,0 dB", "+21,0 dB", "+25,0 dB", "+28,1 dB", "+31,1 dB"},
				key: "mic_gain", group: gTx, label: "Guadagno del microfono"},
			ijvSetDef{off: 0x0A, opts: append([]string{"OFF"}, ijvNums(950, 31, 100, " Hz")...), key: "scrambler_freq", group: gTx, label: "Frequenza dello scrambler"},
			ijvSetDef{off: 0x0C, opts: ijvNums(0, 9, 1, ""), key: "afc", group: gRx, label: "AFC (controllo automatico della frequenza)"},
			ijvSetDef{off: 0x10, opts: []string{"OFF", "FCC", "CE", "GB", "430"}, key: "flock", group: gTx, label: "Blocco delle frequenze di trasmissione (F-Lock)"},
			ijvSetDef{off: 0x17, mask: 0x08, key: "power_on_beep", group: gOp, label: "Bip all'accensione"},
			ijvSetDef{off: 0x22, opts: append([]string{"OFF"}, ijvNums(100, 10, 100, " ms")...), key: "rp_ste", group: gRx, label: "Eliminazione della coda del ponte (RP-STE)"},
			ijvSetDef{off: 0x27, mask: 0x03, opts: []string{"OFF", "Basso", "Medio", "Massimo"}, key: "battery_saver", group: gOp, label: "Risparmio della batteria"},
			ijvSetDef{off: 0x27, mask: 0x08, key: "ptt_toggle", group: gTx, label: "PTT a commutazione (premi una volta per trasmettere)",
				warn: "Con il PTT a commutazione la trasmissione resta accesa finché non premi di nuovo il PTT."})
		if !fw5 {
			keys := []string{"Nessuna", "Torcia", "Potenza di trasmissione", "Monitor", "Scansione", "Radio FM", "VFO successivo", "Modo VFO",
				"Squelch +", "Squelch -", "Prova REGA", "Allarme REGA", "Preset", "AGC manuale", "Elenco dei canali", "Rubrica",
				"Chiamata CQ in CW", "VOX", "Scansione veloce", "Copia della frequenza (FCOPY)", "Scansione dei toni"}
			for k, n := range []string{"1, pressione breve", "1, pressione lunga", "2, pressione breve", "2, pressione lunga"} {
				d = append(d, ijvSetDef{off: 0x190 + k, opts: keys, key: fmt.Sprintf("side%d", k+1), group: gKeys, label: "Tasto laterale " + n})
			}
		}
		for b := 0; b < 16; b++ {
			d = append(d, ijvSetDef{off: ijvBankNames + b*8, str: 8, key: fmt.Sprintf("bank_name_%d", b+1), group: gNames,
				label: fmt.Sprintf("Nome del banco %d", b+1)})
		}
	}
	gn := ijvFam{kind}.groupNames()
	for g := 0; g < 16; g++ {
		lab := fmt.Sprintf("Nome del gruppo %d", g)
		if g == 0 {
			lab = "Nome del gruppo 0 (tutti)"
		}
		if kind == ijvV3X {
			lab += " (banco 0)"
		}
		d = append(d, ijvSetDef{off: gn + g*8, str: 8, key: fmt.Sprintf("group_name_%d", g), group: gNames, label: lab})
	}
	return d
}

func ijvSetTable(kind ijvKind) *setTable {
	src := "tabella delle impostazioni di RadioManager 4.3.2 di IJV (firmware V4)"
	switch kind {
	case ijvFW5, ijvV6:
		src = "tabella delle impostazioni di RadioManager 4.3.2 di IJV (firmware V5 e V6)"
	case ijvV3, ijvV3X:
		src = "modulo CHIRP 3.60 pubblicato da IJV (firmware V3)"
	}
	t := &setTable{Class: "IJV-" + kind.family(), Size: ijvFam{kind}.imgSize(), Source: src}
	ascii := map[string]int{}
	for c := 0x20; c <= 0x7E; c++ {
		ascii[string(rune(c))] = c
	}
	pad := 0
	for _, b := range ijvSetDefs(kind) {
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.desc, ItWarn: b.warn}
		if b.str > 0 {
			offs := make([]int, b.str)
			for k := range offs {
				offs[k] = b.off + k
			}
			sd.Type, sd.MaxLen, sd.Autopad = "str", b.str, true
			sd.Enc = &setEnc{Kind: "str", Offs: offs, Map: ascii, Pad: &pad}
			t.Settings = append(t.Settings, sd)
			continue
		}
		mask := b.mask
		if mask == 0 {
			mask = 0xFF
		}
		sh := 0
		for mask>>sh&1 == 0 {
			sh++
		}
		val := func(i int) *string {
			s := hex.EncodeToString([]byte{byte(i<<sh) & mask})
			return &s
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(b.off), hex.EncodeToString([]byte{mask})}}}
		sd.Enc = enc
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

var ijvSetTables = map[string]*setTable{}

func init() {
	for id, k := range map[string]ijvKind{"qs-uvk5-ijv3": ijvV3, "qs-uvk5-ijv3x": ijvV3X, "qs-uvk5v3-ijv4": ijvV4, "qs-uvk5v3-ijv5": ijvFW5, "qs-uvk5v3-ijv6": ijvV6} {
		ijvSetTables[id] = ijvSetTable(k)
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return ijvSetTables[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := ijvSetTables[kind]
			if t == nil || len(img) != t.Size {
				return nil
			}
			return t
		},
	})
}
