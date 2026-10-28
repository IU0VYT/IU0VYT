// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni della Bajeton BJ-9000 (Radtel RT-950 Pro): blocco 0x9000-0x907F
// della memoria della radio, come lo leggono e lo scrivono i metodi
// Get/SetFunConfigInfo_Part1..3 del programma del costruttore «BJ-9000_CPS»
// v1.1.1 (BT-RT950PRO_CPS.exe, classe KDH.RWDataOperation); scelte dalla finestra
// «Optional Features» (FormFunConfig). Verificate con il codice originale del
// programma collegato alla radio simulata (verifica/bt11, settings_bj9000_test.go).
// Si modificano solo i bit di ogni impostazione: gli altri byte del blocco restano
// quelli letti dalla radio (il programma originale li riscrive a 0xFF).

import (
	"encoding/hex"
	"strconv"
)

const bj9kSetBase = 0x9000

type bj9kSetDef struct {
	field                         string // proprietà di KDH.FunConfigData (senza «CbB_»)
	off                           int    // posizione nel blocco 0x9000
	mask                          byte
	opts                          []string // etichette italiane (vuoto: OFF/ON)
	inv                           bool     // ON = 0 (radio FM)
	key, group, label, desc, warn string
}

func bj9kKeyOpts() []string {
	return []string{"Nessuna", "Radio FM", "VOX", "Ricerca frequenza", "Analizzatore di spettro", "Meteo (NOAA)", "Scansione dei toni",
		"Squelch", "Passo di frequenza", "Potenza di trasmissione", "Memorizza canale", "Scelta della zona", "Doppio ascolto",
		"CTCSS/DCS", "Scostamento (offset)", "Direzione dello scostamento", "Modulazione in ricezione", "Tono di chiamata",
		"Ripetitore (transfer)", "GPS acceso/spento", "APRS acceso/spento", "Segnale di fine trasmissione", "Programmazione Bluetooth"}
}

func bj9kSideOpts(n int) []string {
	return []string{"Radio FM", "Monitor (apre lo squelch)", "Scansione", "Ricerca frequenza", "Allarme (SOS)", "Analizzatore di spettro",
		"Invio beacon APRS", "PTT della banda C"}[:n]
}

var bj9kSetDefs = func() []bj9kSetDef {
	g1, g2, g3, g4 := "Ricezione e trasmissione", "Bande, zone e modalità", "Display, suoni e tasti", "DTMF, allarme, APRS e funzioni remote"
	bands := []string{"Nome", "Frequenza", "Numero del canale"}
	zones := strNums(15, 1, 1, "")
	d := []bj9kSetDef{
		{field: "SQL", off: 0, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...), group: g1},
		{field: "SaveMode", off: 1, opts: []string{"OFF", "Normale", "Super", "Profondo"}, group: g3},
		{field: "VOXSwitch", off: 0x28, key: "vox_on", group: g1, label: "VOX (trasmissione a voce)",
			desc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.",
			warn: "Con il VOX acceso la radio trasmette appena sente un suono: rumori ambientali o conversazioni vicine possono finire in onda."},
		{field: "VOX", off: 2, opts: strNums(9, 1, 1, ""), key: "vox_level", group: g1, label: "Sensibilità del VOX",
			desc: "Livello di voce che fa partire la trasmissione con il VOX acceso (1 = più sensibile)."},
		{field: "VoxDelay", off: 0x20, opts: func() []string {
			o := make([]string, 16)
			for i := range o {
				o[i] = strconv.FormatFloat(0.5+float64(i)*0.1, 'f', 1, 64) + " s"
			}
			return o
		}(), group: g1},
		{field: "AutoBacklight", off: 3, opts: []string{"Sempre accesa", "5 s", "10 s", "15 s", "20 s", "30 s", "1 min", "2 min", "3 min"}, group: g3},
		{field: "TDR", off: 4, key: "tdr", group: g2, label: "Doppio ascolto", desc: "La radio ascolta alternativamente più bande."},
		{field: "TOT", off: 5, opts: append([]string{"OFF"}, strNums(8, 30, 30, " s")...), group: g1},
		{field: "BeepPrompt", off: 6, group: g3},
		{field: "VoicePrompt", off: 7, group: g3},
		{field: "Language", off: 8, opts: []string{"Inglese", "Cinese"}, group: g3},
		{field: "DTMF", off: 9, opts: []string{"OFF", "Tasti", "Codice ANI", "Tasti + ANI"}, group: g4},
		{field: "Scan", off: 10, opts: []string{"TO (a tempo)", "CO (a portante)", "SE (si ferma)"}, group: g2},
		{field: "PTTID", off: 11, opts: []string{"OFF", "All'inizio", "Alla fine", "Inizio e fine"}, group: g4},
		{field: "SendIDDelay", off: 12, opts: []string{"0 ms", "100 ms", "200 ms", "400 ms", "600 ms", "800 ms", "1000 ms"}, group: g4},
		{field: "DisplayModeA", off: 13, opts: bands, group: g3},
		{field: "DisplayModeB", off: 14, opts: bands, group: g3},
		{field: "DisplayModeC", off: 15, opts: bands, key: "disp_c", group: g3, label: "Display della banda C", desc: "Cosa mostra il display per i canali della banda C."},
		{field: "AutoKeyLock", off: 16, opts: []string{"OFF", "5 s", "10 s", "15 s"}, group: g3},
		{field: "AlarmMode", off: 17, opts: []string{"Solo sulla radio", "Trasmette il suono", "Trasmette il codice"}, group: g4},
		{field: "AlarmSound", off: 18, group: g4},
		{field: "TailNoiseClear", off: 20, group: g1},
		{field: "PassRepetNoiseClear", off: 21, opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...), group: g1},
		{field: "PassRepetNoiseDetect", off: 22, opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...), group: g1},
		{field: "SoundTxEnd", off: 23, opts: []string{"OFF", "Beep", "Tono 1200 Hz"}, group: g1},
		{field: "CurWorkMode", off: 24, opts: []string{"Banda A", "Banda B", "Banda C"}, group: g2},
		{field: "FMRadio", off: 25, inv: true, group: g2},
		{field: "WorkModeA", off: 26, mask: 0x03, opts: []string{"Frequenza (VFO)", "Canali"}, group: g2},
		{field: "WorkModeB", off: 26, mask: 0x0C, opts: []string{"Frequenza (VFO)", "Canali"}, group: g2},
		{field: "WorkModeC", off: 26, mask: 0x30, opts: []string{"Frequenza (VFO)", "Canali"}, key: "mode_c", group: g2, label: "Modalità della banda C",
			desc: "Frequenza libera (VFO) o canali memorizzati."},
		{field: "LockKeyBoard", off: 27, group: g3},
		{field: "PowerMsg", off: 28, opts: []string{"Logo predefinito", "Tensione della batteria"}, group: g3},
		{field: "BTWriteSwitch", off: 29, key: "bt_prog", group: g4},
		{field: "RTone", off: 30, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}, group: g1},
		{field: "TimerMenuQuit", off: 0x21, opts: strNums(11, 5, 5, " s"), group: g3},
		{field: "WeatherCH", off: 0x25, opts: strNums(10, 1, 1, ""), group: g2},
		{field: "DivideCH", off: 0x26, opts: []string{"Zone (15 × 64)", "Tutti i canali (960)"}, group: g2},
		{field: "SubaudioScanSave", off: 0x27, opts: []string{"Ricezione e trasmissione", "Solo ricezione", "Solo trasmissione"}, key: "qt_scan_save", group: g2,
			label: "Salvataggio del tono trovato con la scansione", desc: "Dove salvare il tono CTCSS/DCS trovato con la scansione dei toni: in ricezione e trasmissione, solo in ricezione o solo in trasmissione."},
		{field: "KeySide1", off: 0x29, opts: bj9kSideOpts(8), key: "pf1_short", group: g3, label: "Tasto laterale 1 (pressione breve)", desc: "Funzione del tasto laterale 1 premuto brevemente."},
		{field: "KeySide1L", off: 0x2A, opts: bj9kSideOpts(8), key: "pf1_long", group: g3, label: "Tasto laterale 1 (pressione lunga)", desc: "Funzione del tasto laterale 1 tenuto premuto."},
		{field: "KeySide2", off: 0x2B, opts: bj9kSideOpts(7), group: g3},
		{field: "KeySide2L", off: 0x2C, opts: bj9kSideOpts(7), group: g3},
		{field: "CurWorkAreaA", off: 0x2D, opts: zones, key: "zone_a", group: g2, label: "Zona attiva della banda A", desc: "Zona di canali selezionata sulla banda A."},
		{field: "CurWorkAreaB", off: 0x2E, opts: zones, key: "zone_b", group: g2, label: "Zona attiva della banda B", desc: "Zona di canali selezionata sulla banda B."},
		{field: "CurWorkAreaC", off: 0x2F, opts: zones, key: "zone_c", group: g2, label: "Zona attiva della banda C", desc: "Zona di canali selezionata sulla banda C."},
		{field: "ABUVTransfer", off: 0x39, key: "ab_transfer", group: g4, label: "Ripetitore tra le bande A e B (AB-UV transfer)",
			desc: "La radio ritrasmette su una banda quello che riceve sull'altra (funzione di ponte ripetitore incrociato).",
			warn: "Come ripetitore la radio trasmette da sola, anche a lungo: usala solo dove è consentito (in Italia un ponte va autorizzato) e controlla il riscaldamento."},
		{field: "SoundTransfer", off: 0x3A, key: "transfer_sound", group: g4, label: "Altoparlante durante la ripetizione",
			desc: "Fa sentire nell'altoparlante il segnale ritrasmesso quando la radio funziona da ripetitore."},
	}
	for k := 0; k < 10; k++ {
		d = append(d, bj9kSetDef{field: "Key" + strconv.Itoa(k) + "L", off: 0x3B + k, mask: 0x1F, opts: bj9kKeyOpts(),
			key: "key" + strconv.Itoa(k) + "_long", group: g3, label: "Tasto " + strconv.Itoa(k) + " (pressione lunga)",
			desc: "Funzione rapida del tasto numerico " + strconv.Itoa(k) + " tenuto premuto."})
	}
	// nomi, descrizioni e avvisi comuni con la BJ-8300 (stesse voci del programma)
	common := map[string]bajSetDef{}
	for _, b := range bajSetDefs {
		common[b.field[4:]] = b
	}
	common["BTWriteSwitch"] = common["BTRWData"]
	common["VOXSwitch"] = bajSetDef{}
	for i := range d {
		c := common[d[i].field]
		if d[i].key == "" {
			d[i].key = c.key
		}
		if d[i].label == "" {
			d[i].label = c.label
		}
		if d[i].desc == "" {
			d[i].desc = c.desc
		}
		if d[i].warn == "" && d[i].field != "VOX" {
			d[i].warn = c.warn
		}
		if d[i].mask == 0 {
			d[i].mask = 0x0F
		}
	}
	return d
}()

var bj9kSettings = func() *setTable {
	t := &setTable{Class: "BJ-9000", Size: 0x9080, Source: "programma BJ-9000_CPS del costruttore (finestra «Optional Features»)"}
	for _, b := range bj9kSetDefs {
		sh := 0
		for b.mask>>sh&1 == 0 {
			sh++
		}
		val := func(i int) *string {
			s := hex.EncodeToString([]byte{byte(i<<sh) & b.mask})
			return &s
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(bj9kSetBase + b.off), hex.EncodeToString([]byte{b.mask})}}}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.desc, ItWarn: b.warn, Enc: enc}
		if len(b.opts) == 0 {
			sd.Type = "bool"
			if b.inv {
				enc.Table = []*string{val(1), val(0)}
			} else {
				enc.Table = []*string{val(0), val(1)}
			}
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
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "bj-9000" },
		get: func(kind string, img []byte) *setTable {
			if len(img) < bj9kSettings.Size {
				return nil
			}
			return bj9kSettings
		},
	})
}
