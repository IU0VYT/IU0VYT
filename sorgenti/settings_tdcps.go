// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali (4.30) dei TIDRADIO del programma «TIDRadioCPS» del 2026:
// TD-H3 Plus / SE, TD-H7 Plus, TD-H8 di 4ª generazione, TD-H9 e TD-638UV
// (drv_tdh8.go, opzione cps26). Indirizzi e bit vengono dalla finestra «Optional
// Settings» del programma del costruttore; per ogni voce coincidono con la
// struttura del TD-H3 descritta dal driver tdh8.py di CHIRP. Le voci su cui le due
// fonti non concordano, o che il programma non mostra, non sono offerte: quei bit
// restano come li ha lasciati la radio.

import (
	"encoding/hex"
	"strconv"
)

type tdcDef struct {
	key   string
	addr  int    // indirizzo nella memoria della radio (nel file: + 8)
	mask  byte   // bit usati dalla voce
	vals  []byte // valori delle scelte (già senza spostamento); vuoto: 0, 1, 2…
	opts  []string
	n     int // lunghezza di un testo (voce di testo)
	group string
	label string
	dsc   string
	warn  string
}

const (
	tdcG1 = "Ricezione e trasmissione"
	tdcG2 = "Display, suoni e tasti"
	tdcG3 = "Tasti programmabili"
	tdcG4 = "Scansione e modo di lavoro"

	tdcKeyDsc = "Funzione del tasto, con i nomi del programma del costruttore."
)

var (
	tdcKeyLong  = []string{"Nessuna", "Radio FM", "Torcia", "Monitor (squelch aperto)", "Tono", "Allarme", "Meteo"}
	tdcStep     = []string{"2,5 kHz", "5 kHz", "6,25 kHz", "10 kHz", "12,5 kHz", "25 kHz", "50 kHz", "0,5 kHz", "8,33 kHz"}
	tdcWork     = []string{"VFO (frequenza)", "VFO e canali", "Solo canali"}
	tdcLang     = []string{"English", "中文", "Türkçe", "Русский", "Deutsch", "Español", "Italiano", "Français"}
	tdcLangThai = append(append([]string{}, tdcLang...), "ภาษาไทย")
	tdcColors   = []string{"Blu", "Rosso", "Verde", "Giallo", "Viola", "Arancione", "Azzurro", "Ciano", "Grigio", "Blu scuro", "Verde chiaro",
		"Marrone", "Magenta", "Bordeaux", "Blu ardesia", "Grigio chiaro", "Blu ardesia chiaro", "Blu acciaio chiaro"}
)

// tdcHang: 0,5 s … 10,0 s a passi di mezzo secondo
func tdcHang() []string {
	out := make([]string, 20)
	for i := range out {
		out[i] = strconv.Itoa((i+1)/2) + ",0 s"
		if i%2 == 0 {
			out[i] = strconv.Itoa((i+1)/2) + ",5 s"
		}
	}
	return out
}

// tdcDefs: voci del modello. top: ha il tasto superiore; gnss: sulla pressione
// breve la funzione 2 è il GNSS e fra le lingue c'è il tailandese; se: la scheda
// vale anche per il TD-H3 SE, che non ha né la scelta della lingua né «OD PTT».
func tdcDefs(top, gnss, se bool) []tdcDef {
	short := []string{"Nessuna", "Radio FM", "Torcia", "Tono", "Allarme", "Meteo", "PTT2", "OD PTT"}
	if gnss {
		short[2] = "GNSS (posizione)"
	}
	shortVals := []byte{0, 1, 2, 4, 5, 6, 7, 8}
	dShort, lang, langDsc := tdcKeyDsc, tdcLang, "Lingua dei menu della radio."
	if se {
		dShort += " Sul TD-H3 SE la funzione «OD PTT» non c'è."
		langDsc += " Sul TD-H3 SE questa scelta non c'è: lasciala com'è."
	}
	if gnss {
		lang = tdcLangThai
	}
	var d []tdcDef
	if top {
		d = append(d,
			tdcDef{key: "key_top_short", addr: 0xC90, mask: 0xFF, vals: shortVals, opts: short, group: tdcG3, label: "Tasto superiore, pressione breve", dsc: dShort},
			tdcDef{key: "key_top_long", addr: 0xC93, mask: 0xFF, opts: tdcKeyLong, group: tdcG3, label: "Tasto superiore, pressione lunga", dsc: tdcKeyDsc})
	}
	d = append(d,
		tdcDef{key: "key_side1_short", addr: 0xC91, mask: 0xFF, vals: shortVals, opts: short, group: tdcG3, label: "Tasto laterale 1, pressione breve", dsc: dShort},
		tdcDef{key: "key_side1_long", addr: 0xC94, mask: 0xFF, opts: tdcKeyLong, group: tdcG3, label: "Tasto laterale 1, pressione lunga", dsc: tdcKeyDsc},
		tdcDef{key: "key_side2_short", addr: 0xC92, mask: 0xFF, vals: shortVals, opts: short, group: tdcG3, label: "Tasto laterale 2, pressione breve", dsc: dShort},
		tdcDef{key: "key_side2_long", addr: 0xC95, mask: 0xFF, opts: tdcKeyLong, group: tdcG3, label: "Tasto laterale 2, pressione lunga", dsc: tdcKeyDsc},

		tdcDef{key: "squelch", addr: 0xCA9, mask: 0xFF, opts: strNums(10, 0, 1, ""), group: tdcG1, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		tdcDef{key: "tot", addr: 0xCAA, mask: 0xFF, opts: []string{"OFF", "30 s", "60 s", "90 s", "120 s", "150 s", "180 s", "210 s"}, group: tdcG1,
			label: "Tempo massimo di trasmissione (TOT)", dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		tdcDef{key: "vox_level", addr: 0xCA7, mask: 0x07, opts: []string{"OFF", "1", "2", "3", "4", "5"}, group: tdcG1, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX).", warn: wVOX},
		tdcDef{key: "vox_delay", addr: 0xCAE, mask: 0x0F, opts: []string{"1 s", "2 s", "3 s"}, group: tdcG1, label: "Ritardo del VOX",
			dsc: "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare."},
		tdcDef{key: "mic_gain", addr: 0x1F20, mask: 0xFF, opts: strNums(10, 0, 1, ""), group: tdcG1, label: "Guadagno del microfono",
			dsc: "Sensibilità del microfono: più alto il numero, più forte l'audio trasmesso."},
		tdcDef{key: "save", addr: 0xCAC, mask: 0xFF, opts: []string{"OFF", "1:1", "1:2", "1:3", "1:4"}, group: tdcG1, label: "Risparmio della batteria",
			dsc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi."},
		tdcDef{key: "dual_watch", addr: 0xCA3, mask: 0x04, group: tdcG1, label: "Doppio ascolto", dsc: "La radio ascolta alternativamente le due bande A e B."},
		tdcDef{key: "pri_tx", addr: 0xCA0, mask: 0x01, opts: []string{"Banda selezionata", "Banda dell'ultima chiamata ricevuta"}, group: tdcG1,
			label: "Trasmissione con il doppio ascolto", dsc: "In doppio ascolto, su quale banda trasmettere (quella dell'ultimo segnale ricevuto o sempre la stessa)."},
		tdcDef{key: "tail", addr: 0xCA7, mask: 0x40, group: tdcG1, label: "Eliminazione della coda di rumore",
			dsc: "A fine trasmissione la radio manda un segnale che fa chiudere subito lo squelch di chi ascolta con il tono, senza il fruscio finale."},
		tdcDef{key: "roger", addr: 0xCAB, mask: 0xC0, opts: []string{"OFF", "Tono 1", "Tono 2"}, group: tdcG1, label: "Segnale di fine trasmissione (roger beep)",
			dsc: "Suono inviato in aria quando rilasci il PTT."},
		tdcDef{key: "tone_burst", addr: 0xCA2, mask: 0x30, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}, group: tdcG1, label: "Tono di chiamata per i ponti",
			dsc: "Frequenza del tono di apertura inviato ai ponti che lo richiedono (in Europa di solito 1750 Hz)."},
		tdcDef{key: "dtmf_st", addr: 0xCA0, mask: 0x02, group: tdcG1, label: "Ascolto dei toni DTMF trasmessi",
			dsc: "Fa sentire nell'altoparlante i toni DTMF che la radio trasmette."},
		tdcDef{key: "tx200", addr: 0xCAB, mask: 0x10, group: tdcG1, label: "Trasmissione sulla banda dei 200 MHz", dsc: "Voci «TX 200», «TX 350» e «TX 500» del programma del costruttore: abilitano la trasmissione su quella banda.", warn: wTxBand},
		tdcDef{key: "tx350", addr: 0xCAB, mask: 0x08, group: tdcG1, label: "Trasmissione sulla banda dei 350 MHz", dsc: "Voci «TX 200», «TX 350» e «TX 500» del programma del costruttore: abilitano la trasmissione su quella banda.", warn: wTxBand},
		tdcDef{key: "tx500", addr: 0xCAB, mask: 0x04, group: tdcG1, label: "Trasmissione sulla banda dei 500 MHz", dsc: "Voci «TX 200», «TX 350» e «TX 500» del programma del costruttore: abilitano la trasmissione su quella banda.", warn: wTxBand},

		tdcDef{key: "work_a", addr: 0xCA2, mask: 0x03, opts: tdcWork, group: tdcG4, label: "Modo di lavoro della banda A",
			dsc: "Se sulla banda si sceglie una frequenza libera (VFO), un canale memorizzato o tutte e due le cose."},
		tdcDef{key: "work_b", addr: 0xCA3, mask: 0x03, opts: tdcWork, group: tdcG4, label: "Modo di lavoro della banda B",
			dsc: "Se sulla banda si sceglie una frequenza libera (VFO), un canale memorizzato o tutte e due le cose."},
		tdcDef{key: "only_ch", addr: 0xCAF, mask: 0x80, group: tdcG4, label: "Solo canali (modo a canale singolo)",
			dsc: "La radio lavora soltanto sui canali memorizzati.", warn: wMenu},
		tdcDef{key: "step_a", addr: 0xCA8, mask: 0xF0, opts: tdcStep, group: tdcG4, label: "Passo di frequenza della banda A",
			dsc: "Di quanto cambia la frequenza a ogni scatto quando la radio è in modo VFO."},
		tdcDef{key: "step_b", addr: 0xCA8, mask: 0x0F, opts: tdcStep, group: tdcG4, label: "Passo di frequenza della banda B",
			dsc: "Di quanto cambia la frequenza a ogni scatto quando la radio è in modo VFO."},
		tdcDef{key: "scan_mode", addr: 0xCA1, mask: 0xC0, opts: []string{"TO", "CO", "SE"}, group: tdcG4, label: "Modalità di scansione",
			dsc: "Come riprende la scansione dopo un segnale: TO dopo qualche secondo, CO quando il segnale finisce, SE si ferma sul segnale trovato."},
		tdcDef{key: "scan_hang", addr: 0x1F2F, mask: 0xFF, opts: tdcHang(), group: tdcG4, label: "Ripresa della scansione",
			dsc: "Quanto resta ferma la scansione, dopo la fine del segnale, prima di ripartire."},
		tdcDef{key: "fm_rx", addr: 0xCA2, mask: 0x08, group: tdcG4, label: "Radio FM interrotta dalla ricezione",
			dsc: "Mentre ascolti la radio FM, una chiamata sul canale la interrompe e viene fatta sentire."},

		tdcDef{key: "bl_level", addr: 0xC9D, mask: 0xFF, opts: []string{"5", "4", "3", "2", "1"}, group: tdcG2, label: "Luminosità dello schermo",
			dsc: "Livello di luminosità della retroilluminazione, da 1 (minima) a 5 (massima)."},
		tdcDef{key: "bl_time", addr: 0xCAD, mask: 0xFF, opts: []string{"Sempre accesa", "5 s", "10 s", "15 s", "20 s", "25 s", "30 s"}, group: tdcG2,
			label: "Spegnimento della retroilluminazione", dsc: "Dopo quanto tempo si spegne la luce del display (o se resta sempre accesa)."},
		tdcDef{key: "tx_led", addr: 0xCA0, mask: 0x80, group: tdcG2, label: "Schermo acceso in trasmissione", dsc: "Accende il display mentre trasmetti."},
		tdcDef{key: "rx_led", addr: 0xCA0, mask: 0x40, group: tdcG2, label: "Schermo acceso in ricezione", dsc: "Accende il display quando arriva un segnale."},
		tdcDef{key: "breath", addr: 0xCAF, mask: 0x70, opts: []string{"OFF", "5 s", "10 s", "15 s", "30 s"}, group: tdcG2, label: "Lampeggio della spia in attesa",
			dsc: "Voce «Breath Led» del programma del costruttore: ogni quanto si accende la spia mentre la radio è in attesa."},
		tdcDef{key: "name_a", addr: 0xCA2, mask: 0x04, group: tdcG2, label: "Banda A: mostra il nome del canale", dsc: "Cosa mostra il display per i canali della banda A."},
		tdcDef{key: "name_b", addr: 0xCA3, mask: 0x10, group: tdcG2, label: "Banda B: mostra il nome del canale", dsc: "Cosa mostra il display per i canali della banda B."},
		tdcDef{key: "style", addr: 0x1F29, mask: 0xFF, opts: []string{"Singolo", "Doppio", "Classico"}, group: tdcG2, label: "Stile dello schermo",
			dsc: "Aspetto della schermata principale (voci «Single», «Dual» e «Classic» del programma del costruttore)."},
		tdcDef{key: "menu_color", addr: 0x1F2A, mask: 0xFF, opts: tdcColors, group: tdcG2, label: "Colore dei menu", dsc: "Colore usato per i menu della radio."},
		tdcDef{key: "lang", addr: 0x1F28, mask: 0xFF, opts: lang, group: tdcG2, label: "Lingua", dsc: langDsc},
		tdcDef{key: "boot", addr: 0xCA3, mask: 0xC0, opts: []string{"Tensione della batteria", "Messaggio", "Immagine"}, group: tdcG2, label: "Schermata all'accensione",
			dsc: "Che cosa mostra la radio quando la accendi."},
		tdcDef{key: "boot_msg1", addr: 0x1C00, n: 16, group: tdcG2, label: "Messaggio all'accensione, riga 1", dsc: "Testo mostrato all'accensione (fino a 16 caratteri, senza lettere accentate)."},
		tdcDef{key: "boot_msg2", addr: 0x1C10, n: 16, group: tdcG2, label: "Messaggio all'accensione, riga 2", dsc: "Testo mostrato all'accensione (fino a 16 caratteri, senza lettere accentate)."},
		tdcDef{key: "boot_msg3", addr: 0x1C20, n: 16, group: tdcG2, label: "Messaggio all'accensione, riga 3", dsc: "Testo mostrato all'accensione (fino a 16 caratteri, senza lettere accentate)."},
		tdcDef{key: "beep", addr: 0xCA1, mask: 0x04, group: tdcG2, label: "Suono dei tasti", dsc: "Beep alla pressione dei tasti."},
		tdcDef{key: "voice", addr: 0xCA1, mask: 0x01, group: tdcG2, label: "Guida vocale", dsc: "La radio annuncia a voce i menu e i canali."},
		tdcDef{key: "autolock", addr: 0xCA1, mask: 0x10, group: tdcG2, label: "Blocco automatico della tastiera",
			dsc: "La tastiera si blocca da sola dopo qualche secondo di inattività, per evitare pressioni accidentali.", warn: wLock},
		tdcDef{key: "alarm", addr: 0xCAF, mask: 0x01, opts: []string{"Solo sulla radio", "Trasmesso in aria"}, group: tdcG2, label: "Modalità dell'allarme (SOS)",
			dsc: "Se l'allarme suona soltanto dall'altoparlante della radio oppure viene anche trasmesso.", warn: wAlarm})
	return d
}

const tdcBase = 8 // il file comincia con gli 8 byte del nome letto dalla radio

func tdcTable(m *genModel, top, gnss, se bool) *setTable {
	o := tdh8o(m)
	return tdcBuild("TDCPS-"+m.ID, tdcBase+(o.memsize+0x1F)/0x20*0x20, tdcBase, "programma del costruttore TIDRadioCPS", tdcDefs(top, gnss, se))
}

// tdcBuild: tabella delle impostazioni da un elenco di voci; base = posizione
// nel file dell'indirizzo 0 della radio
func tdcBuild(class string, size, base int, source string, defs []tdcDef) *setTable {
	t := &setTable{Class: class, Size: size, Source: source}
	for _, b := range defs {
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.dsc, ItWarn: b.warn}
		if b.n > 0 { // testo: caratteri ASCII stampabili, il resto a zero
			pad, chars, offs := 0, map[string]int{}, make([]int, b.n)
			for i := range offs {
				offs[i] = base + b.addr + i
			}
			for _, c := range asciiPrintable {
				chars[string(c)] = int(c)
			}
			sd.Type, sd.MaxLen, sd.Charset = "str", b.n, asciiPrintable
			sd.Enc = &setEnc{Kind: "str", Offs: offs, Map: chars, Pad: &pad}
			t.Settings = append(t.Settings, sd)
			continue
		}
		shift := uint(0)
		for b.mask>>shift&1 == 0 {
			shift++
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(base + b.addr), hex.EncodeToString([]byte{b.mask})}}}
		sd.Enc, sd.Type = enc, "bool"
		n := 2
		if len(b.opts) > 0 {
			n = len(b.opts)
			sd.Type, sd.ItOpts, sd.Options = "list", b.opts, strNums(n, 0, 1, "")
		}
		for i := 0; i < n; i++ {
			v := byte(i)
			if len(b.vals) > 0 {
				v = b.vals[i]
			}
			s := hex.EncodeToString([]byte{v << shift & b.mask})
			enc.Table = append(enc.Table, &s)
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

var tdcSettings = map[string]*setTable{}

func init() {
	for _, x := range []struct {
		id            string
		top, gnss, se bool
	}{{"td-h3plus", false, false, true}, {"td-h7plus", false, false, false}, {"td-h8g4", true, false, false},
		{"td-h9", true, true, false}, {"td-638uv", true, true, false},
		// 4.32: TD-H4 e TD-H11 (Settings.HasTopKey e KeyCode2IsGNSS valgono anche per loro)
		{"td-h4", true, true, false}, {"td-h11", true, true, false}} {
		for _, suf := range []string{"", "ham", "gmrs"} {
			if m := genByID[x.id+suf]; m != nil {
				tdcSettings[m.ID] = tdcTable(m, x.top, x.gnss, x.se)
			}
		}
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return tdcSettings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := tdcSettings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			blk := img[tdcBase+0xCA0 : tdcBase+0xCB0]
			if wfBlank(blk) {
				return nil // immagine mai letta da una radio: non ci sono impostazioni da mostrare
			}
			return t
		},
	})
}
