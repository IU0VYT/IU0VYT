// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "strconv"

// Impostazioni della Radtel RT-5D (4.44). Ricavate dal programma del costruttore
// «RT-5D CPS» 2.0.5 («JJCC-888DMR_CPS»): le funzioni SetOptionalFunOne / Two (che
// preparano i 64 byte della sezione «funzioni opzionali») e GetOptionalFunOne /
// Two (che li interpretano) usano le stesse posizioni; le scelte vengono da
// InitializeOptionalFunOptions, e il loro numero coincide con il resto usato in
// lettura (valore % numero di scelte). La sezione è nell'immagine a
// rt5dOff("funzioni") e il driver la riscrive com'è (drv_rt5d.go).
//
//   0x00 squelch analogico      0x10 blocco automatico    0x20 ritardo del VOX
//   0x01 risparmio              0x11 tipo di allarme      0x21 uscita dal menu
//   0x02 VOX                    0x12 suono dell'allarme   0x22 squelch digitale
//   0x03 retroilluminazione     0x13 TX in doppio ascolto 0x26 frequenza della coda
//   0x04 doppio ascolto         0x14 coda (STE)           0x27 canale meteo
//   0x05 TOT                    0x15 coda del ponte       0x2A tasto in alto
//   0x06 suono dei tasti        0x16 ritardo della coda   0x2B tasto 2 breve
//   0x07 guida vocale           0x17 tono di fine         0x2C tasto 2 lungo
//   0x08 lingua                 0x18 banda corrente       0x2D tasto 3 breve
//   0x09 ascolto dei DTMF       0x19 radio FM (0 = sì)    0x2E tasto 3 lungo
//   0x0A scansione              0x1A modo A (bit 0-3),    0x31 durata della chiamata
//   0x0B PTT ID                      modo B (bit 4-7)     0x33 ritorno dal doppio
//   0x0C ritardo del PTT ID     0x1B tasti bloccati            ascolto
//   0x0D display A              0x1C schermata iniziale
//   0x0E display B              0x1E tono di chiamata
//                               0x1F tono di inizio

func rt5dSetDefs() []t17SetDef {
	o := rt5dOff("funzioni")
	gRT, gAB, gDK, gKeys, gDT := "Ricezione e trasmissione", "Doppio ascolto e modalità", "Display, suoni e tasti", "Tasti programmabili",
		"DTMF, allarme e funzioni remote"
	lev := append([]string{"OFF"}, strNums(9, 1, 1, "")...)
	ms11 := strNums(11, 0, 100, " ms")
	disp := []string{"Nome del canale", "Frequenza", "Numero del canale"}
	mode := []string{"Frequenza (VFO)", "Canali"}
	keys := []string{"Radio FM", "Monitor", "Scansione", "Lettura della frequenza", "Allarme (SOS)", "Canali meteo NOAA", "Ricerca del tono (CTCSS/DCS)"}
	dec := func(n, first, step int) []string { // decimi di secondo
		out := make([]string, n)
		for i := range out {
			v := first + i*step
			out[i] = strconv.Itoa(v/10) + "," + strconv.Itoa(v%10) + " s"
		}
		return out
	}
	d := []t17SetDef{
		{key: "sql", off: o + 0x00, mask: 0xFF, opts: lev, group: gRT, label: "Livello squelch"},
		{key: "dsql", off: o + 0x22, mask: 0xFF, opts: lev, group: gRT, label: "Livello squelch (canali digitali)"},
		{key: "tot", off: o + 0x05, mask: 0xFF, opts: []string{"OFF", "30 s", "60 s", "120 s", "240 s", "480 s"}, group: gRT,
			label: "Tempo massimo di trasmissione (TOT)", dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "vox", off: o + 0x02, mask: 0xFF, opts: lev, group: gRT, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX).", warn: wVOX},
		{key: "vox_delay", off: o + 0x20, mask: 0xFF, opts: dec(16, 5, 1), group: gRT, label: "Ritardo del VOX",
			dsc: "Per quanto tempo la radio resta in trasmissione dopo che la voce è cessata."},
		{key: "ste", off: o + 0x14, mask: 0xFF, group: gRT, label: "Eliminazione della coda del ponte (STE)"},
		{key: "ste_freq", off: o + 0x26, mask: 0xFF, opts: []string{"55 Hz", "62,5 Hz"}, group: gRT, label: "Frequenza del tono di coda"},
		{key: "rp_ste", off: o + 0x15, mask: 0xFF, opts: ms11, group: gRT, label: "Eliminazione della coda del ponte (RP-STE)"},
		{key: "rpt_rl", off: o + 0x16, mask: 0xFF, opts: ms11, group: gRT, label: "Ritardo della coda del ponte (RPT-RL)"},
		{key: "tx_start", off: o + 0x1F, mask: 0xFF, group: gRT, label: "Tono di inizio trasmissione", dsc: "La radio emette un tono quando inizia a trasmettere."},
		{key: "roger", off: o + 0x17, mask: 0xFF, group: gRT, label: "Tono di fine trasmissione"},
		{key: "tone", off: o + 0x1E, mask: 0xFF, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}, group: gRT, label: "Tono di chiamata dei ponti",
			dsc: "Frequenza del tono che apre i ponti ripetitori (in Europa di solito 1750 Hz)."},
		{key: "keep_call", off: o + 0x31, mask: 0x1F, opts: strNums(20, 1, 1, " s"), group: gRT, label: "Durata della chiamata digitale",
			dsc: "Per quanto tempo, dopo una chiamata digitale ricevuta, il PTT risponde a quella chiamata."},
		{key: "save", off: o + 0x01, mask: 0xFF, opts: []string{"OFF", "1:1", "1:2", "1:3", "1:4"}, group: gRT, label: "Risparmio della batteria",
			dsc: "In attesa la radio spegne il ricevitore a intervalli: più alto il rapporto, più batteria si risparmia (e più facile è perdere l'inizio di una chiamata)."},

		{key: "tdr", off: o + 0x04, mask: 0xFF, group: gAB, label: "Doppio ascolto", dsc: "La radio ascolta le due bande (A e B) alternandole."},
		{key: "tdr_tx", off: o + 0x13, mask: 0xFF, opts: []string{"OFF", "Banda A", "Banda B"}, group: gAB, label: "Banda di trasmissione in doppio ascolto"},
		{key: "tdr_back", off: o + 0x33, mask: 0xFF, opts: append([]string{"0,5 s"}, strNums(10, 1, 1, " s")...), group: gAB, label: "Ritorno dal doppio ascolto",
			dsc: "Tempo dopo il quale, finita una chiamata sull'altra banda, la radio torna alla banda principale."},
		{key: "band", off: o + 0x18, mask: 0xFF, opts: []string{"A", "B"}, group: gAB, label: "Banda corrente", dsc: "Banda su cui la radio lavora."},
		{key: "mode_a", off: o + 0x1A, mask: 0x0F, opts: mode, group: gAB, label: "Banda A: modo di lavoro"},
		{key: "mode_b", off: o + 0x1A, mask: 0xF0, opts: mode, group: gAB, label: "Banda B: modo di lavoro"},
		{key: "disp_a", off: o + 0x0D, mask: 0xFF, opts: disp, group: gAB, label: "Banda A: sul display"},
		{key: "disp_b", off: o + 0x0E, mask: 0xFF, opts: disp, group: gAB, label: "Banda B: sul display"},
		{key: "scan_mode", off: o + 0x0A, mask: 0xFF, opts: []string{"TO", "CO", "SE"}, group: gAB, label: "Ripresa della scansione",
			dsc: "CO: la scansione riparte quando il segnale sparisce; TO: riparte dopo il tempo di sosta; SE: si ferma sul primo segnale."},
		{key: "fm_radio", off: o + 0x19, mask: 0xFF, vals: []int{1, 0}, group: gAB, label: "Radio FM", dsc: "Abilita la funzione radio FM."},
		{key: "wx_ch", off: o + 0x27, mask: 0xFF, opts: strNums(10, 1, 1, ""), group: gAB, label: "Canale meteo NOAA"},

		{key: "beep", off: o + 0x06, mask: 0xFF, group: gDK, label: "Suono dei tasti", dsc: "Segnale acustico a ogni pressione di un tasto."},
		{key: "voice", off: o + 0x07, mask: 0xFF, group: gDK, label: "Guida vocale", dsc: "La radio annuncia a voce canali e funzioni."},
		{key: "lang", off: o + 0x08, mask: 0xFF, opts: []string{"Inglese", "Cinese"}, group: gDK, label: "Lingua"},
		{key: "light", off: o + 0x03, mask: 0xFF, opts: []string{"Sempre accesa", "5 s", "10 s", "15 s", "20 s", "30 s", "1 min", "2 min", "3 min"}, group: gDK,
			label: "Durata della retroilluminazione"},
		{key: "menu_quit", off: o + 0x21, mask: 0xFF, opts: append(strNums(10, 5, 5, " s"), "60 s"), group: gDK, label: "Uscita automatica dal menu",
			dsc: "Dopo questo tempo senza toccare i tasti la radio esce dal menu."},
		{key: "auto_lock", off: o + 0x10, mask: 0xFF, opts: []string{"OFF", "5 s", "10 s", "15 s"}, group: gDK, label: "Blocco automatico dei tasti",
			dsc: "Dopo questo tempo senza toccare i tasti la tastiera si blocca da sola."},
		{key: "key_lock", off: o + 0x1B, mask: 0xFF, group: gDK, label: "Tasti bloccati", dsc: "Stato del blocco della tastiera."},
		{key: "boot", off: o + 0x1C, mask: 0xFF, opts: []string{"Logo", "Tensione"}, group: gDK, label: "Schermata di accensione",
			dsc: "Cosa mostra la radio all'accensione: il logo predefinito o la tensione della batteria."},

		{key: "top_key", off: o + 0x2A, mask: 0xFF, opts: keys, group: gKeys, label: "Tasto in alto, pressione breve"},
		{key: "key2_s", off: o + 0x2B, mask: 0xFF, opts: append(append([]string{}, keys...), "PTT della banda B"), group: gKeys, label: "Tasto laterale 2, pressione breve"},
		{key: "key2_l", off: o + 0x2C, mask: 0xFF, opts: keys, group: gKeys, label: "Tasto laterale 2, pressione lunga"},
		{key: "key3_s", off: o + 0x2D, mask: 0xFF, opts: keys, group: gKeys, label: "Tasto laterale 3, pressione breve"},
		{key: "key3_l", off: o + 0x2E, mask: 0xFF, opts: keys, group: gKeys, label: "Tasto laterale 3, pressione lunga"},

		{key: "dtmf_st", off: o + 0x09, mask: 0xFF, opts: []string{"OFF", "Toni della tastiera", "Toni dell'identificativo", "Tastiera e identificativo"}, group: gDT,
			label: "Ascolto dei toni DTMF trasmessi", dsc: "Fa sentire nell'altoparlante i toni DTMF che la radio trasmette."},
		{key: "ptt_id", off: o + 0x0B, mask: 0xFF, opts: []string{"OFF", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}, group: gDT,
			label: "Invio del codice identificativo (PTT-ID)"},
		{key: "ptt_delay", off: o + 0x0C, mask: 0xFF, opts: []string{"0 ms", "100 ms", "200 ms", "400 ms", "600 ms", "800 ms", "1000 ms"}, group: gDT,
			label: "Ritardo del codice identificativo", dsc: "Attesa tra la pressione del PTT e l'invio del codice identificativo."},
		{key: "alarm_mode", off: o + 0x11, mask: 0xFF, opts: []string{"Sul posto (SITE)", "Tono trasmesso (TONE)", "Codice trasmesso (CODE)"}, group: gDT, label: "Tipo di allarme"},
		{key: "alarm_sound", off: o + 0x12, mask: 0xFF, group: gDT, label: "Suono dell'allarme"},
	}
	return d
}

var rt5dSettings = t17SetTableSize(rt5dSetDefs(), t17Src("«RT-5D CPS» 2.0.5"), rt5dOff("funzioni")+0x40)

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "rtl-rt5d" },
		get: func(kind string, img []byte) *setTable {
			if kind != "rtl-rt5d" || len(img) < rt5dSettings.Size {
				return nil
			}
			return rt5dSettings
		},
	})
}
