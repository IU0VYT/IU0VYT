// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle Radtel RT-830 e RT-850 (4.43). Ricavate dai programmi del
// costruttore «Radtel RT830» («UV-83 Radio Program Software») e «Radtel RT-850
// programming software» («UV5118 Radio Program Software»): la funzione che
// prepara la scrittura (SaveData) e quella che riempie le caselle dalla memoria
// letta (SetPanel). Nell'immagine il blocco delle impostazioni viene dopo i 16
// byte del blocco della password (drv_rt830.go): posizione nell'immagine =
// 0x10 + posizione nel blocco del programma. Il driver riscrive il blocco così
// com'è nell'immagine, quindi le voci cambiate qui arrivano alla radio.
//
// RT-830 (blocco di 48 byte; le voci sono negli ultimi 16):
//   0x31  bit 7 lingua, 6 guida vocale, 5 suono dei tasti, 4 VOX, 3 priorità di
//         trasmissione, 0-2 risparmio della batteria
//   0x32  bit 7 doppio ascolto, 6 area corrente, 4-5 tono di trasmissione, 0-3 squelch
//   0x33  bit 4-7 livello del VOX, 0-3 ritardo del VOX
//   0x34  bit 7 modo del display, 6 canale mostrato come, 0-5 TOT
//   0x35  bit 6-7 direzione dello scostamento, 0-5 blocco automatico dei tasti
//   0x36  bit 7 retroilluminazione, 0-5 durata della retroilluminazione
//   0x37-0x3A scostamento (32 bit, non mostrato)
//   0x3B  bit 4-7 passo, 0-3 codice dello scrambler
//   0x3C  canale dell'area A, 0x3D canale dell'area B
//   0x3E  bit 7 tipo di radio (FM/NOAA), 6 ascolto durante la radio, 5 coda;
//         0-4 e 0x3F frequenza della radio FM (non mostrata)
//
// RT-850 (blocco di 64 byte):
//   0x39  durata del display, 0x3A durata del LED, 0x3B blocco automatico
//   0x3C-0x3F tasti laterali 1 e 2, pressione breve e lunga
//   0x41  bit 6-7 guida vocale, 5 suono dei tasti, 3-4 tono di trasmissione, 0-2 risparmio
//   0x42  bit 7 modo del display, 6 doppio ascolto, 5 gamma corrente, 4 ascolto
//         durante la radio, 0-3 squelch
//   0x43  bit 4-7 livello del VOX, 0-3 ritardo del VOX
//   0x44  bit 6-7 luminosità, 5 direzione dello scostamento, 4 allarme, 0-3 scrambler
//   0x45  bit 7 display, 6 LED, 5 blocco tasti, 4 radio FM, 3 compander, 2 scostamento,
//         1 blocco della frequenza, 0 VOX
//   0x46  bit 4-7 passo, 3 priorità di trasmissione, 0-2 coda
//   0x47-0x4C frequenza della radio FM e scostamento (non mostrati)
//   0x4D  canale della gamma A, 0x4E canale della gamma B, 0x4F TOT
//
// Due incongruenze dei programmi del costruttore, uguali a quella già vista nel
// «BF Series CPS»: in lettura il tono di trasmissione della RT-830 è preso da un
// solo bit (ma le scelte sono quattro e in scrittura occupano due bit) e l'«ascolto
// durante la radio» della RT-850 è letto dal bit 5 (quello della gamma) mentre in
// scrittura va nel bit 4. Qui si segue la scrittura: le due voci hanno un avviso.
//
// Non mostrate: password, scostamento e frequenza della radio FM (numeri su più
// byte legati ai VFO).

const wRT830Prog = "Nel programma del costruttore questa voce è letta in modo diverso da come è scritta: qui si segue la scrittura. " +
	"Se sulla radio l'effetto non è quello atteso, rimetti il valore di prima."

func rt830Times(first string) []string { return append([]string{first}, strNums(40, 15, 15, " s")...) }

func rt830SetDefs() []t17SetDef {
	gRT, gAB, gDK := "Ricezione e trasmissione", "Doppio ascolto e modalità", "Display, suoni e tasti"
	ch := strNums(128, 1, 1, "")
	return []t17SetDef{
		{key: "sql", off: 0x32, mask: 0x0F, opts: strNums(10, 0, 1, ""), group: gRT, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: 0x34, mask: 0x3F, opts: rt830Times("OFF"), group: gRT, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "vox_on", off: 0x31, mask: 0x10, group: gRT, label: "VOX (trasmissione a voce)",
			dsc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", warn: wVOX},
		{key: "vox_level", off: 0x33, mask: 0xF0, opts: strNums(10, 0, 1, ""), group: gRT, label: "Livello del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX."},
		{key: "vox_delay", off: 0x33, mask: 0x0F, opts: strNums(10, 0, 1, ""), group: gRT, label: "Ritardo del VOX",
			dsc: "Per quanto tempo la radio resta in trasmissione dopo che la voce è cessata."},
		{key: "roger", off: 0x32, mask: 0x30, opts: []string{"OFF", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}, group: gRT,
			label: "Tono di trasmissione", dsc: "Tono emesso quando si preme o si rilascia il PTT.", warn: wRT830Prog},
		{key: "tail", off: 0x3E, mask: 0x20, opts: []string{"Corta", "Lunga"}, group: gRT, label: "Coda di fine trasmissione",
			dsc: "Tono che elimina il fruscio a fine trasmissione."},
		{key: "step", off: 0x3B, mask: 0xF0, opts: []string{"0,25 kHz", "1,25 kHz", "2,5 kHz", "5 kHz", "6,25 kHz", "10 kHz", "12,5 kHz", "25 kHz", "50 kHz", "100 kHz"},
			group: gRT, label: "Passo di frequenza", dsc: "Passo con cui i tasti cambiano la frequenza in modo VFO."},
		{key: "offset_dir", off: 0x35, mask: 0xC0, opts: []string{"OFF", "+ (trasmissione più alta)", "− (trasmissione più bassa)"}, group: gRT,
			label: "Direzione dello scostamento", dsc: "In modo VFO: se la trasmissione è spostata rispetto alla ricezione (ponti ripetitori) e in quale verso."},
		{key: "scramble", off: 0x3B, mask: 0x0F, opts: append([]string{"OFF"}, strNums(8, 1, 1, "")...), group: gRT, label: "Scrambler (codice)",
			dsc: "Inversione della voce: si capisce solo con radio che usano lo stesso codice."},
		{key: "save", off: 0x31, mask: 0x07, opts: []string{"OFF", "1:1", "1:2", "1:3", "1:4", "1:5"}, group: gRT, label: "Risparmio della batteria",
			dsc: "In attesa la radio spegne il ricevitore a intervalli: più alto il rapporto, più batteria si risparmia (e più facile è perdere l'inizio di una chiamata)."},

		{key: "dual", off: 0x32, mask: 0x80, group: gAB, label: "Doppio ascolto", dsc: "La radio ascolta le due bande (A e B) alternandole."},
		{key: "area", off: 0x32, mask: 0x40, opts: []string{"A", "B"}, group: gAB, label: "Banda corrente", dsc: "Banda su cui la radio lavora."},
		{key: "tx_pri", off: 0x31, mask: 0x08, opts: []string{"Banda selezionata", "Banda che sta ricevendo"}, group: gAB, label: "Priorità di trasmissione",
			dsc: "Con il doppio ascolto: premendo il PTT si trasmette sulla banda selezionata oppure su quella che sta ricevendo."},
		{key: "display_mode", off: 0x34, mask: 0x80, opts: []string{"Canali", "Frequenza (VFO)"}, group: gAB, label: "Modo di lavoro"},
		{key: "ch_state", off: 0x34, mask: 0x40, opts: []string{"Numero del canale", "Frequenza"}, group: gAB, label: "Canale mostrato come"},
		{key: "ch_a", off: 0x3C, mask: 0xFF, opts: ch, group: gAB, label: "Canale della banda A"},
		{key: "ch_b", off: 0x3D, mask: 0xFF, opts: ch, group: gAB, label: "Canale della banda B"},
		{key: "radio_type", off: 0x3E, mask: 0x80, opts: []string{"Radio FM", "Canali meteo NOAA"}, group: gAB, label: "Tasto radio",
			dsc: "Che cosa si ascolta con la funzione radio: le stazioni FM oppure i canali meteo NOAA (attivi solo in Nord America)."},
		{key: "radio_standby", off: 0x3E, mask: 0x40, group: gAB, label: "Ascolto dei canali durante la radio",
			dsc: "Mentre si ascolta la radio FM la radio continua a sorvegliare il canale e passa alla chiamata quando arriva."},

		{key: "lang", off: 0x31, mask: 0x80, opts: []string{"Cinese", "Inglese"}, group: gDK, label: "Lingua"},
		{key: "voice", off: 0x31, mask: 0x40, group: gDK, label: "Guida vocale", dsc: "La radio annuncia a voce canali e funzioni."},
		{key: "beep", off: 0x31, mask: 0x20, group: gDK, label: "Suono dei tasti", dsc: "Segnale acustico a ogni pressione di un tasto."},
		{key: "light_on", off: 0x36, mask: 0x80, group: gDK, label: "Retroilluminazione"},
		{key: "light_time", off: 0x36, mask: 0x3F, opts: rt830Times("OFF"), group: gDK, label: "Durata della retroilluminazione",
			dsc: "Dopo questo tempo senza toccare i tasti il display si spegne."},
		{key: "auto_lock", off: 0x35, mask: 0x3F, opts: rt830Times("OFF"), group: gDK, label: "Blocco automatico dei tasti",
			dsc: "Dopo questo tempo senza toccare i tasti la tastiera si blocca da sola."},
	}
}

func rt850SetDefs() []t17SetDef {
	gRT, gAB, gDK, gKeys := "Ricezione e trasmissione", "Doppio ascolto e modalità", "Display, suoni e tasti", "Tasti programmabili"
	ch := strNums(128, 1, 1, "")
	keys := []string{"Nessuna funzione", "Monitor", "Allarme locale", "Allarme remoto", "Display acceso/spento", "Blocco della frequenza", "Ricerca di frequenza e tono"}
	d := []t17SetDef{
		{key: "sql", off: 0x42, mask: 0x0F, opts: strNums(10, 0, 1, ""), group: gRT, label: "Livello squelch",
			dsc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio)."},
		{key: "tot", off: 0x4F, mask: 0xFF, opts: rt830Times("OFF"), group: gRT, label: "Tempo massimo di trasmissione (TOT)",
			dsc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", warn: wTOT},
		{key: "vox_on", off: 0x45, mask: 0x01, group: gRT, label: "VOX (trasmissione a voce)",
			dsc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", warn: wVOX},
		{key: "vox_level", off: 0x43, mask: 0xF0, opts: strNums(10, 0, 1, ""), group: gRT, label: "Livello del VOX",
			dsc: "Livello di voce che fa partire la trasmissione con il VOX."},
		{key: "vox_delay", off: 0x43, mask: 0x0F, opts: strNums(10, 0, 1, ""), group: gRT, label: "Ritardo del VOX",
			dsc: "Per quanto tempo la radio resta in trasmissione dopo che la voce è cessata."},
		{key: "roger", off: 0x41, mask: 0x18, opts: []string{"OFF", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}, group: gRT,
			label: "Tono di trasmissione", dsc: "Tono emesso quando si preme o si rilascia il PTT."},
		{key: "tail", off: 0x46, mask: 0x07, opts: strNums(4, 0, 1, ""), group: gRT, label: "Coda di fine trasmissione",
			dsc: "Tono che elimina il fruscio a fine trasmissione."},
		{key: "step", off: 0x46, mask: 0xF0, opts: []string{"0,25 kHz", "1,25 kHz", "2,5 kHz", "5 kHz", "6,25 kHz", "10 kHz", "12,5 kHz", "25 kHz", "50 kHz", "100 kHz"},
			group: gRT, label: "Passo di frequenza", dsc: "Passo con cui i tasti cambiano la frequenza in modo VFO."},
		{key: "offset_on", off: 0x45, mask: 0x04, group: gRT, label: "Scostamento per i ponti",
			dsc: "In modo VFO la trasmissione è spostata rispetto alla ricezione (ponti ripetitori)."},
		{key: "offset_dir", off: 0x44, mask: 0x20, opts: []string{"+ (trasmissione più alta)", "− (trasmissione più bassa)"}, group: gRT, label: "Direzione dello scostamento"},
		{key: "scramble", off: 0x44, mask: 0x0F, opts: append([]string{"OFF"}, strNums(8, 1, 1, "")...), group: gRT, label: "Scrambler (codice)",
			dsc: "Inversione della voce: si capisce solo con radio che usano lo stesso codice."},
		{key: "compander", off: 0x45, mask: 0x08, group: gRT, label: "Compander",
			dsc: "Comprime la voce in trasmissione e la espande in ricezione per ridurre il rumore: serve solo fra radio che lo usano entrambe."},
		{key: "freq_lock", off: 0x45, mask: 0x02, group: gRT, label: "Blocco della frequenza", dsc: "Impedisce di cambiare frequenza dai tasti."},
		{key: "save", off: 0x41, mask: 0x07, opts: []string{"OFF", "1:1", "1:2", "1:3", "1:4", "1:5"}, group: gRT, label: "Risparmio della batteria",
			dsc: "In attesa la radio spegne il ricevitore a intervalli: più alto il rapporto, più batteria si risparmia (e più facile è perdere l'inizio di una chiamata)."},
		{key: "alarm", off: 0x44, mask: 0x10, opts: []string{"Locale", "Remoto (trasmesso)"}, group: gRT, label: "Tipo di allarme",
			dsc: "L'allarme suona solo sulla radio oppure viene anche trasmesso."},

		{key: "dual", off: 0x42, mask: 0x40, group: gAB, label: "Doppio ascolto", dsc: "La radio ascolta le due bande (A e B) alternandole."},
		{key: "area", off: 0x42, mask: 0x20, opts: []string{"A", "B"}, group: gAB, label: "Banda corrente", dsc: "Banda su cui la radio lavora."},
		{key: "tx_pri", off: 0x46, mask: 0x08, opts: []string{"Banda selezionata", "Banda che sta ricevendo"}, group: gAB, label: "Priorità di trasmissione",
			dsc: "Con il doppio ascolto: premendo il PTT si trasmette sulla banda selezionata oppure su quella che sta ricevendo."},
		{key: "display_mode", off: 0x42, mask: 0x80, opts: []string{"Frequenza (VFO)", "Canali"}, group: gAB, label: "Modo di lavoro"},
		{key: "ch_a", off: 0x4D, mask: 0xFF, opts: ch, group: gAB, label: "Canale della banda A"},
		{key: "ch_b", off: 0x4E, mask: 0xFF, opts: ch, group: gAB, label: "Canale della banda B"},
		{key: "radio_on", off: 0x45, mask: 0x10, group: gAB, label: "Radio FM", dsc: "Abilita la funzione radio FM."},
		{key: "radio_standby", off: 0x42, mask: 0x10, group: gAB, label: "Ascolto dei canali durante la radio",
			dsc: "Mentre si ascolta la radio FM la radio continua a sorvegliare il canale e passa alla chiamata quando arriva.", warn: wRT830Prog},

		{key: "voice", off: 0x41, mask: 0xC0, opts: []string{"OFF", "Cinese", "Inglese"}, group: gDK, label: "Guida vocale", dsc: "La radio annuncia a voce canali e funzioni."},
		{key: "beep", off: 0x41, mask: 0x20, group: gDK, label: "Suono dei tasti", dsc: "Segnale acustico a ogni pressione di un tasto."},
		{key: "screen_on", off: 0x45, mask: 0x80, group: gDK, label: "Display acceso"},
		{key: "screen_light", off: 0x44, mask: 0xC0, opts: []string{"Bassa", "Media", "Alta"}, group: gDK, label: "Luminosità del display"},
		{key: "screen_time", off: 0x39, mask: 0xFF, opts: rt830Times("OFF"), group: gDK, label: "Durata del display acceso",
			dsc: "Dopo questo tempo senza toccare i tasti il display si spegne."},
		{key: "led_on", off: 0x45, mask: 0x40, group: gDK, label: "LED acceso"},
		{key: "led_time", off: 0x3A, mask: 0xFF, opts: rt830Times("OFF"), group: gDK, label: "Durata del LED acceso"},
		{key: "key_lock", off: 0x45, mask: 0x20, group: gDK, label: "Tasti bloccati", dsc: "Stato del blocco della tastiera."},
		{key: "auto_lock", off: 0x3B, mask: 0xFF, opts: rt830Times("OFF"), group: gDK, label: "Blocco automatico dei tasti",
			dsc: "Dopo questo tempo senza toccare i tasti la tastiera si blocca da sola."},
	}
	for i, k := range []struct{ key, label string }{{"fs1_s", "Tasto laterale 1, pressione breve"}, {"fs1_l", "Tasto laterale 1, pressione lunga"},
		{"fs2_s", "Tasto laterale 2, pressione breve"}, {"fs2_l", "Tasto laterale 2, pressione lunga"}} {
		d = append(d, t17SetDef{key: k.key, off: 0x3C + i, mask: 0xFF, opts: keys, group: gKeys, label: k.label})
	}
	return d
}

var rt830Settings = map[string]*setTable{
	"rtl-rt830": t17SetTableSize(rt830SetDefs(), t17Src("«Radtel RT830» (UV-83 Radio Program Software)"), 0x10+0x30),
	"rtl-rt850": t17SetTableSize(rt850SetDefs(), t17Src("«Radtel RT-850 programming software» (UV5118)"), 0x10+0x40),
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return rt830Settings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := rt830Settings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
}
