// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali del Baofeng K67 Pro / 5RH Pro (file .xlc): area di 128
// byte a 0x7980, come le legge e le scrive il programma del costruttore «5RH PRO»
// (CPS.exe, classe CPS_5RH.DataProtocol, ConvertSetting e SettingConvert; scelte
// dei menu dalla finestra Frm_GenralSet e da lang-en.xml). Verificate con il
// codice originale del programma (verifica/k67set, settings_k67_test.go).
// Non si toccano: modalità di fabbrica (Engineering), impostazioni Bluetooth di
// associazione e i puntatori ai canali e alle zone attivi.

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

const k67SetBase = 0x7980

func k67Table(off int, mask byte, vals []byte) *setEnc {
	t := make([]*string, len(vals))
	for i, v := range vals {
		s := hex.EncodeToString([]byte{v & mask})
		t[i] = &s
	}
	return &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(k67SetBase + off), hex.EncodeToString([]byte{mask})}}, Table: t}
}

func seqBytes(n int, start, step byte, shift uint) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = (start + byte(i)*step) << shift
	}
	return out
}

func strNums(n, start, step int, suffix string) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = strconv.Itoa(start+i*step) + suffix
	}
	return out
}

func k67List(key, group, label, desc string, off int, mask byte, vals []byte, opts []string) setDef {
	codes := make([]string, len(opts))
	for i := range opts {
		codes[i] = strconv.Itoa(i)
	}
	return setDef{Key: key, Group: group, Label: label, Type: "list", Options: codes, ItOpts: opts, ItLabel: label, ItDesc: desc, Enc: k67Table(off, mask, vals)}
}

func k67Bool(key, group, label, desc string, off int, bit uint) setDef {
	return setDef{Key: key, Group: group, Label: label, Type: "bool", ItLabel: label, ItDesc: desc,
		Enc: k67Table(off, 1<<bit, []byte{0, 1 << bit})}
}

func k67Str(key, group, label, desc string, off, n int, pad int, charset string) setDef {
	m := map[string]int{}
	offs := make([]int, n)
	for i := range offs {
		offs[i] = k67SetBase + off + i
	}
	for _, c := range charset {
		m[string(c)] = int(c)
	}
	p := pad
	return setDef{Key: key, Group: group, Label: label, Type: "str", MaxLen: n, Charset: charset, ItLabel: label, ItDesc: desc,
		Enc: &setEnc{Kind: "str", Offs: offs, Map: m, Pad: &p}}
}

const asciiPrintable = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

var k67Settings = func() *setTable {
	onOff := []string{"Spento", "Acceso"}
	disp := []string{"Frequenza", "Nome del canale", "Numero del canale", "Frequenza e nome"}
	keys := []string{"Nessuna", "Scansione", "Monitor (apre lo squelch)", "Torcia", "Radio FM", "Allarme di emergenza", "GPS",
		"Misura di frequenza", "Bluetooth", "Tono 1750 Hz", "Allarme di caduta", "Chiamata rapida", "Cambio zona", "Tensione della batteria",
		"Potenza di trasmissione", "VOX acceso/spento"}
	keyVals := seqBytes(16, 0, 1, 0)
	blTime := append([]string{"Sempre accesa"}, strNums(26, 5, 1, " s")...)
	blVals := append([]byte{0}, seqBytes(26, 5, 1, 0)...)
	tz := make([]string, 25)
	for i := range tz {
		tz[i] = fmt.Sprintf("UTC%+d", i-12)
	}
	gpsID := strNums(80, 1, 1, "")
	g1, g2, g3, g4, g5, g6 := "Display e illuminazione", "Ricezione e squelch", "Trasmissione", "Audio, voce e toni", "Tasti", "Sicurezza e messaggio di accensione"
	g7, g8 := "Risparmio energetico", "GPS, APRS e altre funzioni"
	t := &setTable{Class: "K67", Size: xlcSize, Source: "programma del costruttore «5RH PRO» (CPS.exe) e verificata con il suo codice originale"}
	t.Settings = []setDef{
		k67List("cha_mode", g1, "Modo di lavoro della banda A", "La banda A (superiore) lavora in modo VFO (frequenza libera, impostata da tastiera) oppure MR (memorie: canali salvati).", 0, 0xFF, []byte{0, 1}, []string{"VFO (frequenza)", "MR (canali in memoria)"}),
		k67List("chb_mode", g1, "Modo di lavoro della banda B", "La banda B (inferiore) lavora in modo VFO (frequenza libera) oppure MR (canali in memoria).", 1, 0xFF, []byte{0, 1}, []string{"VFO (frequenza)", "MR (canali in memoria)"}),
		k67List("blight_time", g1, "Durata della retroilluminazione", "Dopo quanti secondi senza usare la radio si spegne l'illuminazione del display. «Sempre accesa» consuma di più la batteria.", 8, 0xFF, blVals, blTime),
		k67List("blight_lv", g1, "Luminosità del display", "Livello di luminosità della retroilluminazione, da 1 (minima) a 5 (massima).", 9, 0xFF, seqBytes(5, 0, 1, 0), strNums(5, 1, 1, "")),
		k67List("cha_disp", g1, "Visualizzazione dei canali della banda A", "Che cosa mostra il display per i canali in memoria della banda A: la frequenza, il nome, il numero del canale o frequenza e nome insieme.", 10, 0xF0, seqBytes(4, 0, 1, 4), disp),
		k67List("chb_disp", g1, "Visualizzazione dei canali della banda B", "Che cosa mostra il display per i canali in memoria della banda B.", 10, 0x0F, seqBytes(4, 0, 1, 0), disp),
		k67List("disp_dir", g1, "Orientamento del display", "Display normale o capovolto (utile se la radio è appesa a testa in giù, ad esempio al taschino).", 38, 0x08, []byte{0, 0x08}, []string{"Normale", "Capovolto"}),
		k67List("pwon_face", g1, "Schermata di accensione", "Che cosa compare all'accensione: un'immagine, il testo del messaggio di accensione o la tensione della batteria.", 37, 0x03, []byte{0, 1, 2}, []string{"Immagine", "Testo (messaggio di accensione)", "Tensione della batteria"}),
		k67List("lang", g1, "Lingua dei menu", "Lingua dei menu della radio.", 37, 0x04, []byte{0, 0x04}, []string{"Inglese", "Cinese"}),
		k67List("dual_mode", g2, "Doppio ascolto (dual watch)", "Come la radio ascolta le due bande: una sola banda, doppio ascolto (controlla entrambe le bande e si ferma su quella attiva) o due bande visualizzate con ascolto della sola banda principale.", 11, 0xFF, []byte{0, 1, 2}, []string{"Banda singola", "Due bande, doppio ascolto", "Due bande, ascolto singolo"}),
		k67List("main_band", g2, "Banda principale", "Banda (A superiore, B inferiore) usata per trasmettere e mostrata come principale.", 12, 0xFF, []byte{0, 1}, []string{"A", "B"}),
		k67List("sql", g2, "Livello dello squelch", "Soglia del segnale sotto la quale l'audio resta muto. 0 = squelch aperto (si sente sempre il fruscio), 9 = solo segnali molto forti. Valori tipici 2-4.", 13, 0xFF, seqBytes(10, 0, 1, 0), append([]string{"0 (aperto)"}, strNums(9, 1, 1, "")...)),
		k67List("tail", g2, "Eliminazione della coda (STE)", "Elimina il rumore che si sente quando l'altra stazione smette di trasmettere: con tono a 55 Hz o con inversione di fase del subtono (120°, 180°, 240°). Deve essere uguale su tutte le radio del gruppo.", 38, 0xE0, seqBytes(5, 0, 1, 5), []string{"Spenta", "55 Hz", "Fase 120°", "Fase 180°", "Fase 240°"}),
		k67List("busy_lock", g3, "Blocco del canale occupato (BCL)", "Se attivo, la radio non trasmette quando sul canale c'è già un segnale: evita di disturbare una comunicazione in corso.", 32, 0x03, []byte{0, 1}, onOff),
		k67List("tot", g3, "Tempo massimo di trasmissione (TOT)", "Dopo quanti secondi di trasmissione continua la radio smette di trasmettere da sola. Protegge da trasmissioni involontarie (PTT bloccato) e dal surriscaldamento. 0 = nessun limite.", 21, 0xFF, seqBytes(15, 0, 1, 0), append([]string{"Nessun limite"}, strNums(14, 15, 15, " s")...)),
		k67List("pretot", g3, "Preavviso del TOT", "Quanti secondi prima della fine del tempo massimo di trasmissione la radio avvisa con un segnale acustico.", 22, 0xFF, seqBytes(11, 0, 1, 0), append([]string{"Nessun preavviso"}, strNums(10, 1, 1, " s")...)),
		k67Bool("vox", g3, "VOX (trasmissione a comando vocale)", "Se attivo, la radio trasmette da sola quando si parla vicino al microfono, senza premere il PTT. Utile a mani libere, ma rumori ambientali possono far partire trasmissioni indesiderate.", 32, 7),
		k67List("vox_lv", g3, "Sensibilità del VOX", "Livello di voce necessario a far partire la trasmissione col VOX: 1 = molto sensibile (basta un suono debole), 9 = poco sensibile.", 14, 0xFF, seqBytes(9, 1, 1, 0), strNums(9, 1, 1, "")),
		k67List("vox_dly", g3, "Ritardo del VOX", "Dopo quanto tempo di silenzio la trasmissione col VOX si interrompe.", 15, 0xFF, seqBytes(19, 10, 5, 0),
			[]string{"1,0 s", "1,5 s", "2,0 s", "2,5 s", "3,0 s", "3,5 s", "4,0 s", "4,5 s", "5,0 s", "5,5 s", "6,0 s", "6,5 s", "7,0 s", "7,5 s", "8,0 s", "8,5 s", "9,0 s", "9,5 s", "10,0 s"}),
		k67List("hz1750", g3, "Tono di chiamata (burst)", "Frequenza del tono di chiamata inviato con il tasto «1750 Hz», usato per aprire alcuni ponti radio analogici (in Italia di solito 1750 Hz).", 26, 0xFF, []byte{0, 1, 2, 3}, []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}),
		k67List("voice", g4, "Guida vocale", "Annunci vocali della radio (numero del canale, funzioni dei menu).", 32, 0x0C, seqBytes(3, 0, 1, 2), []string{"Spenta", "Cinese", "Inglese"}),
		k67Bool("beep", g4, "Toni dei tasti", "Segnale acustico alla pressione dei tasti.", 34, 7),
		k67List("end_tone", g4, "Tono di fine trasmissione (roger beep)", "Breve tono inviato quando si rilascia il PTT, per segnalare la fine del proprio passaggio. Su molti ponti radio non è gradito.", 34, 0x60, seqBytes(4, 0, 1, 5), []string{"Spento", "Tipo 1", "Tipo 2", "Tipo 3"}),
		k67Bool("record", g4, "Registrazione audio", "Abilita la registrazione delle comunicazioni nella memoria della radio (se prevista).", 37, 7),
		k67List("record_mode", g4, "Che cosa registrare", "Registra solo la ricezione, solo la trasmissione o entrambe.", 37, 0x60, seqBytes(3, 0, 1, 5), []string{"Ricezione", "Trasmissione", "Entrambe"}),
		k67Bool("key_lock", g5, "Blocco della tastiera", "Tastiera bloccata (si sblocca con la combinazione di tasti indicata nel manuale).", 33, 7),
		k67Bool("auto_lock", g5, "Blocco automatico della tastiera", "La tastiera si blocca da sola dopo qualche secondo di inattività, per evitare pressioni accidentali.", 33, 6),
		k67List("skey1", g5, "Tasto laterale 1, pressione breve", "Funzione assegnata alla pressione breve del primo tasto laterale.", 48, 0xFF, keyVals, keys),
		k67List("lkey1", g5, "Tasto laterale 1, pressione lunga", "Funzione assegnata alla pressione lunga del primo tasto laterale.", 50, 0xFF, keyVals, keys),
		k67List("skey2", g5, "Tasto laterale 2, pressione breve", "Funzione assegnata alla pressione breve del secondo tasto laterale.", 49, 0xFF, keyVals, keys),
		k67List("lkey2", g5, "Tasto laterale 2, pressione lunga", "Funzione assegnata alla pressione lunga del secondo tasto laterale.", 51, 0xFF, keyVals, keys),
		k67Str("radioname", g6, "Messaggio di accensione", "Testo mostrato all'accensione quando la schermata di accensione è «Testo» (fino a 16 caratteri).", 80, 16, 0, asciiPrintable),
		k67Str("pow_pwd", g6, "Password di accensione", "Se impostata, la radio chiede questo codice numerico (fino a 8 cifre) a ogni accensione. Lasciare vuoto per non usarla.", 64, 8, 0xFF, "0123456789"),
		k67Str("wr_pwd", g6, "Password di lettura e scrittura", "Se impostata, il programma di programmazione (anche questo) deve fornire questo codice per leggere o scrivere la radio. Lasciare vuoto per non usarla.", 72, 8, 0xFF, "0123456789"),
		k67List("save", g7, "Risparmio della batteria", "Modo di risparmio energetico in ricezione: la radio «dorme» per brevi periodi (rapporto acceso:spento). Rapporti più alti risparmiano di più ma possono far perdere l'inizio delle chiamate.", 16, 0xFF, []byte{0, 1, 2, 3}, []string{"Spento", "1:1", "1:2", "1:4"}),
		k67List("save_dly", g7, "Attesa prima del risparmio", "Dopo quanti secondi senza segnali parte il risparmio della batteria.", 17, 0xFF, seqBytes(5, 0, 1, 0), strNums(5, 5, 5, " s")),
		k67List("apo", g7, "Spegnimento automatico", "La radio si spegne da sola dopo questo tempo senza essere usata.", 20, 0xFF, seqBytes(6, 0, 1, 0), []string{"Spento", "30 min", "60 min", "120 min", "240 min", "480 min"}),
		k67Bool("gps", g8, "GPS", "Attiva il ricevitore di posizione satellitare (consuma batteria).", 35, 7),
		k67List("gps_mode", g8, "Sistema satellitare", "Sistema usato per la posizione: GPS, BeiDou o entrambi.", 35, 0x60, seqBytes(3, 0, 1, 5), []string{"GPS", "BeiDou", "GPS + BeiDou"}),
		k67Bool("gps_share", g8, "Condivisione della posizione", "Invia la propria posizione alle altre radio compatibili.", 35, 4),
		k67Bool("gps_req", g8, "Risposta alla richiesta di posizione", "Risponde automaticamente quando un'altra radio chiede la posizione.", 35, 3),
		k67List("gps_zone", g8, "Fuso orario", "Fuso orario per l'ora ricavata dal GPS (Italia: UTC+1, UTC+2 con l'ora legale).", 24, 0xFF, seqBytes(25, 0, 1, 0), tz),
		k67List("gps_id", g8, "Identificativo GPS", "Numero (1-80) che identifica questa radio nella condivisione della posizione.", 31, 0xFF, seqBytes(80, 0, 1, 0), gpsID),
		k67Bool("aprs", g8, "APRS", "Attiva l'invio della posizione con il protocollo APRS (per l'uso radioamatoriale serve il nominativo nelle impostazioni APRS del programma del costruttore).", 32, 6),
		k67Bool("lone_work", g8, "Lavoratore isolato (lone worker)", "Funzione di sicurezza: se la radio non viene usata per il tempo indicato, chiede una conferma e poi invia un allarme.", 32, 5),
		{Key: "lone_time", Group: g8, Label: "Lavoratore isolato: tempo", Type: "int", Min: 0, Max: 255, Step: 1, ItLabel: "Lavoratore isolato: tempo", ItDesc: "Minuti di inattività prima della richiesta di conferma (funzione «lavoratore isolato»).",
			Enc: &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(k67SetBase + 18), "ff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: 1}},
		{Key: "lone_resp", Group: g8, Label: "Lavoratore isolato: tempo di risposta", Type: "int", Min: 0, Max: 255, Step: 1, ItLabel: "Lavoratore isolato: tempo di risposta", ItDesc: "Secondi a disposizione per confermare prima che parta l'allarme (funzione «lavoratore isolato»).",
			Enc: &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(k67SetBase + 19), "ff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: 1}},
		k67Bool("fall_alarm", g8, "Allarme di caduta (man down)", "Invia un allarme se la radio resta inclinata o ferma a terra per un certo tempo.", 32, 4),
		k67Bool("noaa", g8, "Allerta meteo NOAA", "Ascolto dei canali meteo NOAA (solo negli Stati Uniti; in Italia non serve).", 38, 4),
		k67List("noaa_ch", g8, "Canale meteo NOAA", "Canale meteo NOAA da controllare (solo Stati Uniti).", 30, 0xFF, seqBytes(10, 0, 1, 0), strNums(10, 1, 1, "")),
		k67Bool("weather", g8, "Allarme meteo", "Segnalazione degli allarmi meteo ricevuti sul canale NOAA (solo Stati Uniti).", 37, 3),
		k67Bool("bt", g8, "Bluetooth", "Attiva il Bluetooth della radio (cuffie, programmazione con l'app).", 36, 7),
		k67Bool("bt_app", g8, "Bluetooth: collegamento con l'app", "Permette la programmazione e il controllo della radio dall'app per smartphone via Bluetooth.", 36, 4),
	}
	for i := range t.Settings {
		d := &t.Settings[i]
		switch d.Key {
		case "pow_pwd":
			d.ItWarn = "Se imposti una password di accensione e la dimentichi, non potrai più usare la radio finché non viene azzerata dal rivenditore. Annotala in un posto sicuro."
		case "wr_pwd":
			d.ItWarn = "Con una password di lettura e scrittura la radio si programma solo conoscendo il codice (in IU0VYT Flash va scritto nella finestra di lettura/scrittura). Se la dimentichi non potrai più programmarla."
		case "tot":
			d.ItWarn = "Senza limite di tempo un PTT bloccato o premuto per errore lascia la radio in trasmissione: disturbi agli altri e possibile surriscaldamento."
		case "vox":
			d.ItWarn = "Con il VOX acceso la radio trasmette appena sente un suono: rumori ambientali o conversazioni vicine possono finire in onda."
		case "key_lock":
			d.ItWarn = "Con la tastiera bloccata i tasti non rispondono finché non la sblocchi con la combinazione indicata nel manuale."
		}
	}
	return t
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == kindXLC },
		get: func(kind string, img []byte) *setTable {
			if len(img) < xlcSize {
				return nil
			}
			return k67Settings
		},
	})
}
