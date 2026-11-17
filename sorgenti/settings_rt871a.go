// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"encoding/hex"
	"strconv"
)

// Impostazioni della Radtel RT-817A / RT-871A (4.43). Ricavate dal programma del
// costruttore «Radtel RT-871A CPS» (250707): le due funzioni che riempiono le
// caselle dalla memoria letta e che preparano la scrittura usano le stesse
// posizioni, tutte nel blocco delle impostazioni (i primi 4096 byte dell'immagine,
// che il driver legge e riscrive per intero).
//
// La radio è parente della RT-880G di CHIRP, ma solo le prime voci stanno allo
// stesso posto (0x5C-0x69 e il DTMF a 0x200): da 0x6E in poi la mappa è diversa
// (nella RT-871A il doppio ascolto è in 0x66 e la priorità di trasmissione in 0x7E;
// squelch, VOX e guadagni sono a 0x100-0x114). Per questo la tabella è propria e non
// «gemella» di quella di CHIRP.
//
//   0x11  salvaschermo           0x80  passo             0xA2  direzione scansione
//   0x5C  guida vocale           0x83  banda principale  0xA3  modo scansione
//   0x5D  suono dei tasti        0x84  modo banda A      0xA4  ritorno scansione
//   0x5E  blocco tasti           0x85  display banda A   0xA5  sosta scansione
//   0x5F  blocco automatico      0x89  modo banda B      0xA6  aggiornamento voce (16 bit)
//   0x62  durata del display     0x8A  display banda B   0xA8  aggiornamento segnale (16 bit)
//   0x63  risparmio batteria     0x8E, 0x93, 0x98, 0x9D  0xAA-0xAD tasti laterali
//   0x64  avvio del risparmio      gamme bloccate: tipo, 0xB0-0xB9 tasti 0-9
//   0x65  uscita dal menu          inizio e fine (16 bit) 0xEA  cifre della frequenza
//   0x66  doppio ascolto         0x100 tono singolo (16 bit, Hz)
//   0x67  diretta / inversione   0x102 squelch           0x10B tono di inizio
//   0x68  tipo di allarme        0x105 microfono         0x10C tono di fine
//   0x7E  priorità TX            0x106 volume            0x10D VOX, 0x10E soglia, 0x10F ritardo
//   0x7F  banda del PTT          0x110 gamma del frequenzimetro
//   0x111 ritardo del ponte (16 bit)  0x114 soglia del canale adiacente
//   0x200-0x209 DTMF
//
// Non mostrate: immagine, suoneria, messaggio e password di accensione, nome della
// radio, spegnimento automatico (un tempo di 32 bit con ore, minuti e secondi), i
// canali correnti delle due bande, la casella «SPTT» (significato non ricavabile
// dal programma) e i byte che il programma scrive sempre con un valore fisso.

const rt871aSetSize = 0x1000

// rsLE: valore su «width» byte, il meno significativo per primo
func rsLE(v, width int) *string {
	b := make([]byte, width)
	for i := range b {
		b[i] = byte(v >> (8 * i))
	}
	s := hex.EncodeToString(b)
	return &s
}

func rsEnc(off, width, n, first int) *setEnc {
	mask := ""
	for i := 0; i < width; i++ {
		mask += "ff"
	}
	e := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(off), mask}}}
	for i := 0; i < n; i++ {
		e.Table = append(e.Table, rsLE(first+i, width))
	}
	return e
}

func rsBool(key, group, label, desc string, off int) setDef {
	return setDef{Key: key, Group: group, Label: label, ItLabel: label, ItDesc: desc, Type: "bool", Enc: rsEnc(off, 1, 2, 0)}
}

func rsList(key, group, label, desc string, off, width int, opts []string) setDef {
	d := setDef{Key: key, Group: group, Label: label, ItLabel: label, ItDesc: desc, Type: "list", ItOpts: opts, Enc: rsEnc(off, width, len(opts), 0)}
	for i := range opts {
		d.Options = append(d.Options, strconv.Itoa(i))
	}
	return d
}

func rsInt(key, group, label, desc string, off, width, min, max int) setDef {
	return setDef{Key: key, Group: group, Label: label, ItLabel: label, ItDesc: desc, Type: "int",
		Min: float64(min), Max: float64(max), Step: 1, Enc: rsEnc(off, width, max-min+1, min)}
}

// rsNum: numero su 16 bit (meno significativo per primo) senza tabella
func rsNum(key, group, label, desc string, off, min, max int) setDef {
	return setDef{Key: key, Group: group, Label: label, ItLabel: label, ItDesc: desc, Type: "int",
		Min: float64(min), Max: float64(max), Step: 1,
		Enc: &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(off), "ffff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: 1, Scale: 1}}
}

func rt871aSetTable() *setTable {
	gRT, gAB, gDK, gSave, gKeys, gLock, gDTMF := "Ricezione e trasmissione", "Doppio ascolto e modalità", "Display, suoni e tasti",
		"Risparmio energetico", "Tasti programmabili", "Gamme bloccate", "DTMF, allarme e funzioni remote"
	// tempi delle caselle «timer» del programma: spento, 5, 10, 15 e poi di 15 in 15 fino a 600 secondi
	timer := func(first string) []string {
		o := []string{first, "5 s", "10 s", "15 s"}
		for s := 30; s <= 600; s += 15 {
			o = append(o, strconv.Itoa(s)+" s")
		}
		return o
	}
	ms := func(n, start, step int) []string {
		o := strNums(n, start, step, " ms")
		if start == 0 {
			o[0] = "0"
		}
		return o
	}
	keys := []string{"Nessuna funzione", "Monitor", "Cambio di potenza", "Doppio ascolto sì/no", "Priorità di trasmissione", "Scansione",
		"Retroilluminazione sì/no", "Tono di fine trasmissione", "Radio FM", "Diretta / inversione", "Allarme di emergenza",
		"Lettura della frequenza", "Ricerca del tono (CTCSS/DCS)", "Invio del tono singolo", "Stato della radio", "Ricezione AM/FM",
		"Squelch", "Passo di frequenza", "Canali meteo NOAA", "Salva canale"}
	mode := []string{"Frequenza (VFO)", "Canale", "Zona"}
	show := []string{"Numero del canale", "Frequenza"}

	s := []setDef{
		rsInt("sql", gRT, "Livello squelch",
			"Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio).",
			0x102, 1, 0, 10),
		rsList("step", gRT, "Passo di frequenza", "Passo con cui i tasti cambiano la frequenza in modo VFO.", 0x80, 1,
			[]string{"0,25 kHz", "1,25 kHz", "2,5 kHz", "5 kHz", "6,25 kHz", "10 kHz", "12,5 kHz", "20 kHz", "25 kHz", "50 kHz", "100 kHz", "500 kHz", "1 MHz", "5 MHz"}),
		rsList("talkaround", gRT, "Diretta / inversione",
			"«Diretta» trasmette sulla frequenza di ricezione, senza passare dal ponte; «Frequenze invertite» scambia ricezione e trasmissione (per ascoltare l'ingresso del ponte).",
			0x67, 1, []string{"OFF", "Diretta (senza ponte)", "Frequenze invertite"}),
		rsInt("mic_gain", gRT, "Guadagno del microfono", "Amplificazione del microfono in trasmissione (0-31).", 0x105, 1, 0, 31),
		rsInt("rx_volume", gRT, "Volume di ricezione", "Livello dell'audio in ricezione (0-63).", 0x106, 1, 0, 63),
		rsInt("glitch", gRT, "Soglia del canale adiacente",
			"Soglia con cui la radio scarta i disturbi dei canali vicini (0-10). Se non hai un motivo preciso, lascia il valore che c'è.", 0x114, 1, 0, 10),
		rsList("relay_delay", gRT, "Ritardo dopo la coda del ponte",
			"Tempo di attesa dopo la fine della trasmissione prima di riaprire l'ascolto: serve a non sentire la coda del ponte.", 0x111, 2, ms(21, 0, 100)),
		rsBool("vox", gRT, "VOX (trasmissione a voce)", "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT.", 0x10D),
		rsInt("vox_th", gRT, "Soglia del VOX", "Livello di voce che fa partire la trasmissione con il VOX.", 0x10E, 1, 0, 255),
		rsList("vox_delay", gRT, "Ritardo del VOX", "Per quanto tempo la radio resta in trasmissione dopo che la voce è cessata.", 0x10F, 1,
			[]string{"0 s", "0,5 s", "1 s", "1,5 s", "2 s", "2,5 s", "3 s", "3,5 s", "4 s", "4,5 s", "5 s"}),
		rsBool("tx_start_beep", gRT, "Tono di inizio trasmissione", "La radio emette un tono quando inizia a trasmettere.", 0x10B),
		rsList("roger", gRT, "Tono di fine trasmissione",
			"Che cosa la radio invia quando rilasci il PTT: niente, un tono (roger beep), il nominativo oppure la posizione GPS.", 0x10C, 1,
			[]string{"OFF", "Tono 1", "Tono 2", "Invia il nominativo", "Invia la posizione GPS"}),
		rsNum("tone", gRT, "Frequenza del tono singolo (Hz)", "Frequenza del tono inviato con la funzione «Invio del tono singolo» (per esempio 1750 Hz per aprire i ponti).", 0x100, 0, 20000),
		rsList("detect_range", gRT, "Gamma del frequenzimetro", "Gamma in cui la funzione «Lettura della frequenza» cerca il segnale.", 0x110, 1,
			[]string{"18-64 MHz", "64-136 MHz", "136-174 MHz", "174-240 MHz", "240-320 MHz", "320-400 MHz", "400-480 MHz", "480-560 MHz", "560-640 MHz", "840-920 MHz", "920-1000 MHz"}),

		rsBool("dw", gAB, "Doppio ascolto", "La radio ascolta le due bande (A e B) alternandole.", 0x66),
		rsList("main_range", gAB, "Banda corrente", "Banda su cui la radio lavora.", 0x83, 1, []string{"A", "B"}),
		rsList("tx_pri", gAB, "Priorità di trasmissione",
			"Con il doppio ascolto: premendo il PTT si trasmette sulla banda selezionata oppure su quella che sta ricevendo.", 0x7E, 1,
			[]string{"Banda selezionata", "Banda che sta ricevendo"}),
		rsList("main_ptt", gAB, "Banda del PTT principale", "Su quale banda trasmette il tasto PTT principale.", 0x7F, 1, []string{"Banda A", "Banda principale"}),
		rsList("mode_a", gAB, "Banda A: modo di lavoro", "", 0x84, 1, mode),
		rsList("show_a", gAB, "Banda A: sul display", "", 0x85, 1, show),
		rsList("mode_b", gAB, "Banda B: modo di lavoro", "", 0x89, 1, mode),
		rsList("show_b", gAB, "Banda B: sul display", "", 0x8A, 1, show),
		rsList("scan_dir", gAB, "Direzione della scansione", "", 0xA2, 1, []string{"In su", "In giù"}),
		rsList("scan_mode", gAB, "Ripresa della scansione",
			"CO: la scansione riparte quando il segnale sparisce; TO: riparte dopo il tempo di sosta; SE: si ferma sul primo segnale.", 0xA3, 1, []string{"CO", "TO", "SE"}),
		rsList("scan_return", gAB, "Al termine della scansione torna a", "", 0xA4, 1, []string{"Canale di partenza", "Canale corrente"}),
		rsList("scan_dwell", gAB, "Sosta della scansione", "Tempo di sosta su un canale occupato (modo TO).", 0xA5, 1, strNums(31, 0, 1, " s")),

		rsBool("voice", gDK, "Guida vocale", "La radio annuncia a voce canali e funzioni.", 0x5C),
		rsBool("beep", gDK, "Suono dei tasti", "Segnale acustico a ogni pressione di un tasto.", 0x5D),
		rsBool("key_lock", gDK, "Tasti bloccati", "Stato del blocco della tastiera.", 0x5E),
		rsList("lock_timer", gDK, "Blocco automatico dei tasti", "Dopo questo tempo senza toccare i tasti la tastiera si blocca da sola.", 0x5F, 1, timer("OFF")),
		rsList("lcd_timer", gDK, "Durata del display acceso", "Dopo questo tempo senza toccare i tasti il display si spegne.", 0x62, 1, timer("OFF")),
		rsList("saver_timer", gDK, "Avvio del salvaschermo", "Dopo questo tempo parte il salvaschermo.", 0x11, 1, timer("OFF")),
		rsList("menu_timer", gDK, "Uscita automatica dal menu", "Dopo questo tempo senza toccare i tasti la radio esce dal menu.", 0x65, 1, timer("OFF")),
		rsList("freq_digits", gDK, "Cifre per inserire la frequenza", "Numero di cifre da battere sulla tastiera per inserire una frequenza.", 0xEA, 1, []string{"6 cifre", "8 cifre"}),
		rsList("voice_refresh", gDK, "Aggiornamento dell'indicatore della voce", "Ogni quanto si aggiorna sul display il livello della voce in trasmissione.", 0xA6, 2, ms(21, 0, 100)),
		rsList("rssi_refresh", gDK, "Aggiornamento dell'indicatore del segnale", "Ogni quanto si aggiorna sul display la forza del segnale ricevuto.", 0xA8, 2, ms(21, 0, 100)),

		rsList("save_mode", gSave, "Risparmio della batteria",
			"In attesa la radio spegne il ricevitore a intervalli: più alto il rapporto, più batteria si risparmia (e più facile è perdere l'inizio di una chiamata).",
			0x63, 1, []string{"OFF", "1:1", "1:2", "1:3"}),
		rsList("save_start", gSave, "Avvio del risparmio", "Tempo senza segnali dopo il quale parte il risparmio della batteria.", 0x64, 1, timer("0 s")),
	}
	knames := []struct {
		key, label string
		off        int
	}{
		{"fs1_s", "Tasto laterale 1, pressione breve", 0xAA}, {"fs1_l", "Tasto laterale 1, pressione lunga", 0xAB},
		{"fs2_s", "Tasto laterale 2, pressione breve", 0xAC}, {"fs2_l", "Tasto laterale 2, pressione lunga", 0xAD},
	}
	for _, k := range knames {
		s = append(s, rsList(k.key, gKeys, k.label, "", k.off, 1, keys))
	}
	for i := 0; i < 10; i++ {
		s = append(s, rsList("key"+strconv.Itoa(i), gKeys, "Tasto "+strconv.Itoa(i)+", pressione lunga", "", 0xB0+i, 1, keys))
	}
	for i := 0; i < 4; i++ {
		n, off := strconv.Itoa(i+1), 0x8E+5*i
		s = append(s,
			rsList("lock"+n, gLock, "Gamma "+n,
				"Che cosa è permesso fra le due frequenze indicate sotto: tutto, solo la ricezione oppure niente.", off, 1,
				[]string{"Libera", "Solo ricezione", "Bloccata"}),
			rsNum("lock"+n+"_start", gLock, "Gamma "+n+": da (MHz)", "", off+1, 18, 1000),
			rsNum("lock"+n+"_end", gLock, "Gamma "+n+": a (MHz)", "", off+3, 18, 1000))
	}
	s = append(s,
		rsList("alarm_type", gDTMF, "Tipo di allarme",
			"Che cosa fa il tasto di allarme: suona solo la radio, trasmette l'allarme oppure tutte e due le cose.", 0x68, 1,
			[]string{"Locale", "Remoto (trasmesso)", "Locale e remoto"}),
		rsList("dtmf_mode", gDTMF, "Invio automatico del DTMF", "Quando la radio invia da sola il codice DTMF scelto sotto.", 0x203, 1,
			[]string{"OFF", "A inizio trasmissione", "A fine trasmissione", "A inizio e a fine"}),
		rsList("dtmf_select", gDTMF, "Codice DTMF da inviare", "", 0x204, 1, func() []string {
			o := make([]string, 16)
			for i := range o {
				o[i] = "DTMF-" + string(rune('0'+(i+1)/10)) + string(rune('0'+(i+1)%10))
			}
			return o
		}()),
		rsList("dtmf_delay", gDTMF, "DTMF: ritardo prima dell'invio", "", 0x200, 1, ms(21, 0, 100)),
		rsList("dtmf_duration", gDTMF, "DTMF: durata di ogni cifra", "", 0x201, 1, ms(18, 30, 10)),
		rsList("dtmf_interval", gDTMF, "DTMF: pausa fra le cifre", "", 0x202, 1, ms(18, 30, 10)),
		rsBool("dtmf_display", gDTMF, "Mostra i DTMF ricevuti", "", 0x205),
		rsInt("dtmf_gain", gDTMF, "DTMF: livello di trasmissione", "Livello dei toni DTMF trasmessi (0-127).", 0x206, 1, 0, 127),
		rsInt("dtmf_th", gDTMF, "DTMF: soglia di decodifica", "Livello minimo perché un tono DTMF ricevuto sia riconosciuto (0-63).", 0x207, 1, 0, 63),
		rsBool("dtmf_remote", gDTMF, "Comandi remoti via DTMF",
			"La radio accetta i comandi remoti ricevuti in DTMF.", 0x208),
		rsBool("time_cal", gDTMF, "Regolazione dell'ora via radio", "La radio regola l'orologio con i dati ricevuti.", 0x209),
	)
	for i := range s {
		switch s[i].Key {
		case "vox":
			s[i].ItWarn = wVOX
		case "dtmf_remote":
			s[i].ItWarn = "Con i comandi remoti attivi chi conosce i codici può agire sulla radio a distanza: attivali solo se ti servono."
		}
	}
	return &setTable{Class: "IU0VYT_Radtel_RT871A", Size: rt871aSetSize, Source: t17Src("«Radtel RT-871A CPS» (250707)"), Settings: s}
}

var rt871aSettings = rt871aSetTable()

// rs871Pick (4.44): le voci «keys» della tabella della RT-871A, spostate di
// «delta» byte, per le radio dello stesso progetto del costruttore («DM550_NOLCD»)
// che le hanno alla stessa posizione nel blocco delle impostazioni e con le
// stesse scelte (verificato programma per programma: settings_rt4d.go)
func rs871Pick(keys []string, delta int) []setDef {
	var out []setDef
	for _, key := range keys {
		found := false
		for _, d := range rt871aSettings.Settings {
			if d.Key != key {
				continue
			}
			if len(d.Enc.Fields) != 1 {
				panic("rs871Pick: " + key)
			}
			e := *d.Enc
			e.Fields = [][2]interface{}{{d.Enc.Fields[0][0].(float64) + float64(delta), d.Enc.Fields[0][1]}}
			d.Enc = &e
			out = append(out, d)
			found = true
		}
		if !found {
			panic("rs871Pick: manca " + key)
		}
	}
	return out
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "rtl-rt871a" },
		get: func(kind string, img []byte) *setTable {
			if kind != "rtl-rt871a" || len(img) < rt871aSetSize {
				return nil
			}
			return rt871aSettings
		},
	})
}
