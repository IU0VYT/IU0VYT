// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali (4.27) di Retevis C6, Retevis C66 e Radioddity Echo-DE /
// Echo-IT, dal codice dei programmi dei costruttori (Form1.displaySet per la
// lettura e i gestori «SelectedIndexChanged» per la scrittura). Indirizzi
// relativi all'inizio delle impostazioni (0x600 nel C6, 0x300 negli altri):
//
//   +0   squelch 0-9
//   +1   guida vocale (C6, C66: spenta, inglese; Echo: spenta, cinese, inglese)
//   +2   solo C6 e C66: bit 0 abbinamento con un tasto (0 = acceso), bit 1 copia
//        senza fili (0 = accesa), bit 2-4 suono di trasmissione, bit 5-7 tono
//   +3   C6 e C66: bit 4 VOX, bit 0-3 livello; Echo: livello 0-9 (0 = spento)
//   +4   tempo massimo di trasmissione a passi di 15 s: nel C6 1 = 15 s (lo 0 è
//        mostrato come 15 s e non si può scegliere), negli altri 0 = spento
//   +6   risparmio della batteria
//   +10…+13  tasti laterali 1 e 2, pressione breve e lunga; nel C6 «NOAA» vale 9
//        e l'8 è mostrato come «nessuna»
//   +14  solo Echo: ritardo del VOX a passi di mezzo secondo

import "encoding/hex"

type wltSetDef struct {
	key   string
	off   int
	mask  byte     // 0 = tutto il byte
	vals  []byte   // valore del campo per ogni scelta (vuoto: 0, 1, 2…)
	opts  []string // vuoto: casella sì/no
	inv   bool     // casella sì/no rovesciata: 0 = acceso
	group string
	label string
	dsc   string
	warn  string
}

func wltSetDefs(id string) []wltSetDef {
	g1, g2 := "Ricezione e trasmissione", "Display, suoni e tasti"
	echo, c6 := id == "rd-echo", id == "rt-c6"
	d := []wltSetDef{
		{key: "sql", off: 0, opts: strNums(10, 0, 1, ""), group: g1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
	}
	tot := wltSetDef{key: "tot", off: 4, group: g1, label: "Tempo massimo di trasmissione (TOT)",
		dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola."}
	if c6 {
		tot.opts = strNums(20, 15, 15, " s")
		for v := 1; v <= 20; v++ {
			tot.vals = append(tot.vals, byte(v))
		}
	} else {
		tot.opts, tot.warn = append([]string{"OFF"}, strNums(20, 15, 15, " s")...), wTOT
	}
	d = append(d, tot)
	vox := "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT."
	lev := "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX)."
	if echo {
		d = append(d,
			wltSetDef{key: "vox_level", off: 3, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...), group: g1, label: "Sensibilità del VOX", dsc: lev, warn: wVOX},
			wltSetDef{key: "vox_delay", off: 14, opts: []string{"0 s", "0,5 s", "1 s", "1,5 s", "2 s", "2,5 s", "3 s", "3,5 s", "4 s", "4,5 s", "5 s"}, group: g1, label: "Ritardo del VOX",
				dsc: "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare (evita di tagliare la fine delle frasi)."})
	} else {
		d = append(d,
			wltSetDef{key: "vox_on", off: 3, mask: 0x10, group: g1, label: "VOX (trasmissione a voce)", dsc: vox, warn: wVOX},
			wltSetDef{key: "vox_level", off: 3, mask: 0x0F, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...), group: g1, label: "Sensibilità del VOX", dsc: lev},
			wltSetDef{key: "r_tone", off: 2, mask: 0xE0, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}, group: g1, label: "Tono di chiamata per i ponti",
				dsc: "Frequenza del tono di apertura inviato ai ponti che lo richiedono (in Europa di solito 1750 Hz)."},
			wltSetDef{key: "roger", off: 2, mask: 0x1C, opts: []string{"Spento", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}, group: g2, label: "Suono di trasmissione (roger)",
				dsc: "Breve suono mandato in aria quando premi o rilasci il PTT, per far capire agli altri quando inizi o finisci di parlare."},
			wltSetDef{key: "pair", off: 2, mask: 0x01, inv: true, group: g2, label: "Abbinamento con un tasto",
				dsc: "Permette di copiare dalla radio di un compagno, con una combinazione di tasti, la frequenza e i toni del canale in uso."},
			wltSetDef{key: "clone", off: 2, mask: 0x02, inv: true, group: g2, label: "Copia senza fili",
				dsc: "Permette di copiare via radio tutti i canali da una radio uguale all'altra."})
	}
	voice := wltSetDef{key: "voice", off: 1, opts: []string{"Spenta", "Inglese"}, group: g2, label: "Guida vocale", dsc: "La radio annuncia a voce i menu e i canali."}
	save := wltSetDef{key: "save", off: 6, opts: []string{"OFF", "1:1", "1:2", "1:3", "1:4"}, group: g2, label: "Risparmio della batteria",
		dsc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi."}
	keys := []string{"Nessuna", "Monitor", "Scansione", "Allarme sulla radio", "Allarme trasmesso", "VOX acceso / spento", "Potenza", "Diretta (senza ponte)"}
	var kv []byte
	switch {
	case echo:
		voice.opts = []string{"Spenta", "Cinese", "Inglese"}
		save.opts = append([]string{"OFF"}, "1:1", "1:2", "1:3", "1:4", "1:5", "1:6", "1:7", "1:8", "1:9", "1:10")
		keys = []string{"Nessuna", "Monitor", "Scansione", "Allarme", "Blocco dei tasti"}
	case c6:
		keys, kv = append(keys, "Meteo NOAA"), []byte{0, 1, 2, 3, 4, 5, 6, 7, 9}
	default:
		keys = append(keys, "Frequenze invertite")
	}
	d = append(d, voice, save)
	for i, k := range []struct{ key, label, dsc string }{
		{"pf1_short", "Tasto laterale 1 (pressione breve)", "Funzione del tasto laterale 1 premuto brevemente."},
		{"pf1_long", "Tasto laterale 1 (pressione lunga)", "Funzione del tasto laterale 1 tenuto premuto."},
		{"pf2_short", "Tasto laterale 2 (pressione breve)", "Funzione del tasto laterale 2 premuto brevemente."},
		{"pf2_long", "Tasto laterale 2 (pressione lunga)", "Funzione del tasto laterale 2 tenuto premuto."},
	} {
		d = append(d, wltSetDef{key: k.key, off: 10 + i, vals: kv, opts: keys, group: g2, label: k.label, dsc: k.dsc})
	}
	return d
}

func wltSetTable(m *genModel) *setTable {
	o := wltO(m)
	t := &setTable{Class: "WLT", Size: o.size(), Source: "programma del costruttore " + o.prog}
	for _, b := range wltSetDefs(m.ID) {
		mask, sh := b.mask, 0
		if mask == 0 {
			mask = 0xFF
		}
		for mask>>sh&1 == 0 {
			sh++
		}
		val := func(v byte) *string {
			s := hex.EncodeToString([]byte{v << sh & mask})
			return &s
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(o.set + b.off), hex.EncodeToString([]byte{mask})}}}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.dsc, ItWarn: b.warn, Enc: enc, Type: "bool"}
		switch {
		case len(b.opts) == 0 && b.inv:
			enc.Table = []*string{val(1), val(0)}
		case len(b.opts) == 0:
			enc.Table = []*string{val(0), val(1)}
		default:
			sd.Type, sd.ItOpts = "list", b.opts
			sd.Options = strNums(len(b.opts), 0, 1, "")
			for i := range b.opts {
				v := byte(i)
				if len(b.vals) > 0 {
					v = b.vals[i]
				}
				enc.Table = append(enc.Table, val(v))
			}
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

var wltSettings = map[string]*setTable{}

func init() {
	for _, id := range []string{"rt-c6", "rt-c66", "rd-echo"} {
		wltSettings[id] = wltSetTable(genByID[id])
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return wltSettings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := wltSettings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
}
