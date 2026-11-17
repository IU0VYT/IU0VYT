// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle Radtel RT-480 e RT-68 (4.44). Ricavate dal programma del
// costruttore «Radio Programing Software» (1.5.2 / 1.5.3, progetto «SQ5R»): la
// funzione che prepara il blocco 0x3A00 in scrittura (WriteCHData) e le due che
// lo interpretano in lettura (GetFunCfgImf_Part1 e _Part2) usano le stesse
// posizioni. Le scelte vengono dalle risorse del programma (inglese).
//
// Nell'immagine il blocco 0x3A00 comincia a 0x2000 (drv_rt480.go); dalla 4.44 il
// driver lo riscrive così com'è nell'immagine, subito dopo i nomi, come fa il
// programma del costruttore.
//
//   +0x00 squelch (0-9)              +0x10 blocco automatico dei tasti
//   +0x01 risparmio (0/1)            +0x11 tipo di allarme
//   +0x02 VOX (0 spento, 1-10)       +0x12 suono dell'allarme (0/1)
//   +0x03 retroilluminazione         +0x13 trasmissione in doppio ascolto
//   +0x04 doppio ascolto (0/1)       +0x14 eliminazione della coda (0/1)
//   +0x05 TOT                        +0x15 eliminazione della coda del ponte
//   +0x06 suono dei tasti (0/1)      +0x16 ritardo della coda del ponte
//   +0x07 guida vocale (0/1)         +0x17 tono di fine trasmissione (0/1)
//   +0x08 lingua                     +0x18 sempre 0
//   +0x09 ascolto dei DTMF           +0x19 radio FM: 0 = abilitata, 1 = no
//   +0x0A ripresa della scansione    +0x1A modo di lavoro
//   +0x0B PTT ID                     +0x1B tasti bloccati (0/1)
//   +0x0C ritardo del PTT ID         +0x1C bit 0-1 schermata di accensione,
//   +0x0D display A                        bit 6 «Auto Pwr» (non mostrato)
//   +0x0E display B                  +0x1D bit 0 «Tx Beep», bit 4-7 spegnimento
//   +0x0F blocco su canale occupato        automatico
//                                    +0x1E tono di chiamata
//                                    +0x25 bit 0 ANI
//
// Non mostrati: «Auto Pwr» (la casella non ha scelte leggibili nel programma), il
// codice ANI e tutto ciò che sta nel secondo blocco (0x3A40: VFO e tasti
// laterali) e nel DTMF, che il driver non scrive.

func rt480SetDefs() []t17SetDef {
	const o = rt480SetOff
	gRT, gAB, gDK, gDT := "Ricezione e trasmissione", "Doppio ascolto e modalità", "Display, suoni e tasti", "DTMF, allarme e funzioni remote"
	off10 := append([]string{"OFF"}, strNums(10, 1, 1, "")...)
	disp := []string{"Frequenza", "Numero del canale", "Nome del canale"}
	no := []int{1, 0}
	return []t17SetDef{
		{key: "sql", off: o + 0x00, mask: 0xFF, opts: strNums(10, 0, 1, ""), group: gRT, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: o + 0x05, mask: 0xFF, opts: []string{"OFF", "30 s", "60 s", "120 s", "240 s", "480 s"}, group: gRT,
			label: "Tempo massimo di trasmissione (TOT)", dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "vox", off: o + 0x02, mask: 0xFF, opts: off10, group: gRT, label: "Sensibilità del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX).", warn: wVOX},
		{key: "bcl", off: o + 0x0F, mask: 0xFF, group: gRT, label: "Blocco su canale occupato",
			dsc: "La radio non trasmette se sul canale c'è già un segnale."},
		{key: "ste", off: o + 0x14, mask: 0xFF, group: gRT, label: "Eliminazione della coda del ponte (STE)"},
		{key: "rp_ste", off: o + 0x15, mask: 0xFF, opts: off10, group: gRT, label: "Eliminazione della coda del ponte (RP-STE)"},
		{key: "rpt_rl", off: o + 0x16, mask: 0xFF, opts: off10, group: gRT, label: "Ritardo della coda del ponte (RPT-RL)"},
		{key: "roger", off: o + 0x17, mask: 0xFF, group: gRT, label: "Tono di fine trasmissione"},
		{key: "tx_beep", off: o + 0x1D, mask: 0x01, group: gRT, label: "Bip di trasmissione («Tx Beep»)"},
		{key: "tone", off: o + 0x1E, mask: 0xFF, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}, group: gRT, label: "Tono di chiamata dei ponti",
			dsc: "Frequenza del tono che apre i ponti ripetitori (in Europa di solito 1750 Hz)."},
		{key: "save", off: o + 0x01, mask: 0xFF, group: gRT, label: "Risparmio della batteria"},

		{key: "tdr", off: o + 0x04, mask: 0xFF, group: gAB, label: "Doppio ascolto", dsc: "La radio ascolta le due bande (A e B) alternandole."},
		{key: "tdr_tx", off: o + 0x13, mask: 0xFF, opts: []string{"OFF", "Banda A", "Banda B"}, group: gAB, label: "Banda di trasmissione in doppio ascolto"},
		{key: "work_mode", off: o + 0x1A, mask: 0xFF, opts: []string{"Frequenza (VFO)", "Canali"}, group: gAB, label: "Modo di lavoro"},
		{key: "disp_a", off: o + 0x0D, mask: 0xFF, opts: disp, group: gAB, label: "Banda A: sul display"},
		{key: "disp_b", off: o + 0x0E, mask: 0xFF, opts: disp, group: gAB, label: "Banda B: sul display"},
		{key: "scan_mode", off: o + 0x0A, mask: 0xFF, opts: []string{"TO", "CO", "SE"}, group: gAB, label: "Ripresa della scansione",
			dsc: "CO: la scansione riparte quando il segnale sparisce; TO: riparte dopo il tempo di sosta; SE: si ferma sul primo segnale."},
		{key: "fm_radio", off: o + 0x19, mask: 0xFF, vals: no, group: gAB, label: "Radio FM", dsc: "Abilita la funzione radio FM."},

		{key: "beep", off: o + 0x06, mask: 0xFF, group: gDK, label: "Suono dei tasti", dsc: "Segnale acustico a ogni pressione di un tasto."},
		{key: "voice", off: o + 0x07, mask: 0xFF, group: gDK, label: "Guida vocale", dsc: "La radio annuncia a voce canali e funzioni."},
		{key: "lang", off: o + 0x08, mask: 0xFF, opts: []string{"Inglese", "Cinese"}, group: gDK, label: "Lingua"},
		{key: "light", off: o + 0x03, mask: 0xFF, opts: []string{"OFF", "5 s", "10 s", "15 s", "Sempre accesa"}, group: gDK, label: "Durata della retroilluminazione"},
		{key: "auto_lock", off: o + 0x10, mask: 0xFF, opts: []string{"OFF", "5 s", "10 s", "15 s"}, group: gDK, label: "Blocco automatico dei tasti",
			dsc: "Dopo questo tempo senza toccare i tasti la tastiera si blocca da sola."},
		{key: "key_lock", off: o + 0x1B, mask: 0xFF, group: gDK, label: "Tasti bloccati", dsc: "Stato del blocco della tastiera."},
		{key: "pwr_msg", off: o + 0x1C, mask: 0x03, opts: []string{"Logo", "Messaggio", "Tensione"}, group: gDK, label: "Schermata di accensione",
			dsc: "Cosa mostra la radio all'accensione (logo, messaggio o tensione)."},
		{key: "auto_off", off: o + 0x1D, mask: 0xF0, opts: []string{"OFF", "4 ore", "8 ore", "16 ore", "24 ore"}, group: gDK, label: "Spegnimento automatico"},

		{key: "dtmf_st", off: o + 0x09, mask: 0xFF, opts: []string{"OFF", "Toni della tastiera", "Toni dell'identificativo", "Tastiera e identificativo"}, group: gDT,
			label: "Ascolto dei toni DTMF trasmessi", dsc: "Fa sentire nell'altoparlante i toni DTMF che la radio trasmette."},
		{key: "ptt_id", off: o + 0x0B, mask: 0xFF, opts: []string{"OFF", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}, group: gDT,
			label: "Invio del codice identificativo (PTT-ID)"},
		{key: "ptt_delay", off: o + 0x0C, mask: 0xFF, opts: []string{"0 ms", "100 ms", "200 ms", "400 ms", "600 ms", "800 ms", "1000 ms"}, group: gDT,
			label: "Ritardo del codice identificativo", dsc: "Attesa tra la pressione del PTT e l'invio del codice identificativo."},
		{key: "ani", off: o + 0x25, mask: 0x01, group: gDT, label: "Invio del codice ANI"},
		{key: "alarm_mode", off: o + 0x11, mask: 0xFF, opts: []string{"Sul posto (SITE)", "Tono trasmesso (TONE)", "Codice trasmesso (CODE)"}, group: gDT, label: "Tipo di allarme"},
		{key: "alarm_sound", off: o + 0x12, mask: 0xFF, group: gDT, label: "Suono dell'allarme"},
	}
}

var rt480Settings = t17SetTableSize(rt480SetDefs(), t17Src("«Radio Programing Software» v1.5.2 / v1.5.3"), rt480SetOff+0x40)

func init() {
	is := func(kind string) bool {
		m := genByID[kind]
		if m == nil {
			return false
		}
		_, ok := m.Fam.(rt480Fam)
		return ok
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: is,
		get: func(kind string, img []byte) *setTable {
			if !is(kind) || len(img) < rt480Settings.Size {
				return nil
			}
			return rt480Settings
		},
	})
}
