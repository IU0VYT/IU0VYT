// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "sync"

// Impostazioni delle Radtel RT-4D, RT-6D e RT-1000 Pro (4.44). Ricavate dai tre
// programmi del costruttore («RT-4D CPS V2.07», «DMR RT-6D V2.06», «Radtel RT-1000
// Pro CPS»), dello stesso progetto «DM550_NOLCD» della RT-871A: DataToPanel (che
// riempie le caselle dal blocco delle impostazioni) e SaveAllData (che lo
// riscrive) usano le stesse posizioni in tutti e tre, confrontate in automatico
// casella per casella (48, 48 e 41 caselle a scelta, lettura = scrittura).
//
// Si mostrano le voci che i tre programmi hanno in comune, alla stessa posizione
// e con le stesse scelte. Per 28 di queste posizione e scelte sono anche quelle
// della RT-871A: nomi e descrizioni vengono dalla sua tabella
// (settings_rt871a.go). In più: 0x10 logo, 0x13 suono e 0x14 messaggio di
// accensione, 0x61 luminosità (0-4), 0x10C tono di fine trasmissione (qui la
// quarta scelta è «MDC1200»), 0x111 ritardo dopo la coda del ponte (un byte).
//
// Non mostrate: password di accensione, spegnimento e accensione automatici (con
// un orario), gamme bloccate, tasti programmabili (elenchi diversi nei tre
// programmi), le voci del DMR (toni di chiamata, formato degli SMS…) e le voci
// presenti in un solo programma. Si potranno aggiungere.
//
// Il blocco sta in coda all'immagine (rt4dCfgOff, drv_rt4d.go); in scrittura il
// driver prende dall'immagine solo i bit di queste voci.

var rt4dFrom871 = []string{"sql", "talkaround", "mic_gain", "rx_volume", "glitch", "tx_start_beep", "detect_range", "tx_pri",
	"scan_mode", "scan_return", "scan_dwell", "voice", "beep", "lock_timer", "lcd_timer", "menu_timer", "freq_digits", "save_mode",
	"alarm_type", "dtmf_mode", "dtmf_select", "dtmf_delay", "dtmf_duration", "dtmf_interval", "dtmf_display", "dtmf_gain", "dtmf_th",
	"dtmf_remote"}

func rt4dSetTable() *setTable {
	const o = rt4dCfgOff
	gRT, gDK := "Ricezione e trasmissione", "Display, suoni e tasti"
	s := rs871Pick(rt4dFrom871, o)
	ms := strNums(21, 0, 100, " ms")
	ms[0] = "0"
	s = append(s,
		rsList("relay_delay", gRT, "Ritardo dopo la coda del ponte",
			"Tempo di attesa dopo la fine della trasmissione prima di riaprire l'ascolto: serve a non sentire la coda del ponte.", o+0x111, 1, ms),
		rsList("roger", gRT, "Tono di fine trasmissione",
			"Che cosa la radio invia quando rilasci il PTT: niente, un tono (roger beep), un identificativo MDC1200 oppure la posizione GPS.", o+0x10C, 1,
			[]string{"OFF", "Tono 1", "Tono 2", "MDC1200", "Invia la posizione GPS"}),
		rsList("light", gDK, "Luminosità del display", "", o+0x61, 1, strNums(5, 0, 1, "")),
		rsBool("boot_logo", gDK, "Logo all'accensione", "", o+0x10),
		rsBool("boot_tone", gDK, "Bip all'accensione", "", o+0x13),
		rsBool("boot_text", gDK, "Messaggio di accensione", "", o+0x14),
	)
	return &setTable{Class: "IU0VYT_Radtel_RT4D", Size: rt4dSizeCfg, Source: t17Src("«RT-4D CPS» V2.07, «DMR RT-6D» V2.06, «RT-1000 Pro CPS»"), Settings: s}
}

var (
	rt4dSetOnce  sync.Once
	rt4dSetTab   *setTable
	rt4dSetField [][2]int
)

func rt4dSettings() *setTable {
	rt4dSetOnce.Do(func() {
		rt4dSetTab = rt4dSetTable()
		for i := range rt4dSetTab.Settings {
			offs, masks, _ := rt4dSetTab.Settings[i].Enc.fields()
			for k, m := range masks[0] {
				rt4dSetField = append(rt4dSetField, [2]int{offs[0] - rt4dCfgOff + k, int(m)})
			}
		}
	})
	return rt4dSetTab
}

// rt4dCfgFields: (posizione nel blocco delle impostazioni, maschera) di ogni voce modificabile
func rt4dCfgFields() [][2]int {
	rt4dSettings()
	return rt4dSetField
}

func init() {
	is := func(kind string) bool {
		m := genByID[kind]
		if m == nil {
			return false
		}
		_, ok := m.Fam.(rt4dFam)
		return ok
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: is,
		get: func(kind string, img []byte) *setTable {
			if !is(kind) || len(img) < rt4dSizeCfg {
				return nil
			}
			return rt4dSettings()
		},
	})
}

// Bajeton DM-5000 (4.44): stesso progetto del costruttore («BAJETON DM-5000 CPS»
// 250726, DataToPanel e SaveData sul blocco delle impostazioni, che qui è
// all'inizio dell'immagine come nella RT-871A). Il confronto automatico con il
// programma della RT-871A dà 35 caselle a scelta alla stessa posizione e con le
// stesse scelte, più i campi numerici (squelch, microfono, volume, soglia del
// canale adiacente, DTMF), il tono singolo (0x100), il ritardo del ponte (0x111) e
// le quattro gamme bloccate. Diverse, quindi non mostrate: passo (15 scelte),
// display delle bande (3 scelte), tasti programmabili (39 funzioni), VOX (non
// c'è); il tono di fine trasmissione ha «MDC1200» al posto del nominativo.
// In scrittura il driver tiene dalla radio password (0-11), byte 14-15 e 960-1023
// (drv_dm5000.go): tutte le voci stanno fra 0x10 e 0x209.
var dm5kFrom871 = []string{"sql", "talkaround", "mic_gain", "rx_volume", "glitch", "relay_delay", "tx_start_beep", "tone", "detect_range",
	"dw", "main_range", "tx_pri", "main_ptt", "mode_a", "mode_b", "scan_dir", "scan_mode", "scan_return", "scan_dwell",
	"voice", "beep", "key_lock", "lock_timer", "lcd_timer", "saver_timer", "menu_timer", "freq_digits", "save_mode", "save_start",
	"lock1", "lock1_start", "lock1_end", "lock2", "lock2_start", "lock2_end", "lock3", "lock3_start", "lock3_end", "lock4", "lock4_start", "lock4_end",
	"alarm_type", "dtmf_mode", "dtmf_select", "dtmf_delay", "dtmf_duration", "dtmf_interval", "dtmf_display", "dtmf_gain", "dtmf_th",
	"dtmf_remote", "time_cal"}

var (
	dm5kSetOnce sync.Once
	dm5kSetTab  *setTable
)

func dm5kSettings() *setTable {
	dm5kSetOnce.Do(func() {
		s := rs871Pick(dm5kFrom871, 0)
		s = append(s, rsList("roger", "Ricezione e trasmissione", "Tono di fine trasmissione",
			"Che cosa la radio invia quando rilasci il PTT: niente, un tono (roger beep), un identificativo MDC1200 oppure la posizione GPS.", 0x10C, 1,
			[]string{"OFF", "Tono 1", "Tono 2", "MDC1200", "Invia la posizione GPS"}))
		dm5kSetTab = &setTable{Class: "IU0VYT_BJDM5000", Size: 4 * dm5kBlock, Source: t17Src("«BAJETON DM-5000 CPS» (250726)"), Settings: s}
	})
	return dm5kSetTab
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "bj-dm5000" },
		get: func(kind string, img []byte) *setTable {
			if kind != "bj-dm5000" || len(img) < 4*dm5kBlock {
				return nil
			}
			return dm5kSettings()
		},
	})
}
