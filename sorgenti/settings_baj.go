// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni della Bajeton BJ-8300 (file .dat del programma «BT-11_CPS», finestra
// «Optional Features» = oggetto KDH.FunConfigData). Ogni impostazione è un intero
// a 32 bit dell'oggetto serializzato: la tabella indica la sua posizione nel file
// aperto, così la modifica cambia solo quei 4 byte (il file resta identico a quello
// che salverebbe il programma, verificato con verifica/bt11 e settings_baj_test).
// Scelte, ordine e limiti come nella finestra del programma (v1.0.8 e v1.0.9: la
// v1.0.9 aggiunge «Channel Divide», cbB_DivideCH). Le impostazioni vengono scritte
// nella radio con i canali (memoria 0x9000, RWDataOperation.SetFunConfigInfo_*).

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

type bajSetDef struct {
	field, key, group, label, desc, warn string
	opts                                 []string // etichette italiane, valore = indice
	bool                                 bool     // OFF/ON (opts vuoto)
	inv                                  bool     // sì/no invertito nel file (0 = sì)
}

var bajSetDefs = []bajSetDef{
	// ---- ricezione e trasmissione
	{field: "cbB_SQL", key: "sql", group: "Ricezione e trasmissione", label: "Livello squelch",
		desc: "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. Con OFF l'audio resta sempre aperto (si sente il fruscio).",
		opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...)},
	{field: "cbB_TOT", key: "tot", group: "Ricezione e trasmissione", label: "Tempo massimo di trasmissione (TOT)",
		desc: "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.",
		warn: "Senza limite di tempo (OFF) un PTT bloccato o premuto per errore lascia la radio in trasmissione: disturbi agli altri e possibile surriscaldamento.",
		opts: []string{"OFF", "15 s", "30 s", "45 s", "60 s", "75 s"}},
	{field: "cbB_VOX", key: "vox", group: "Ricezione e trasmissione", label: "VOX (trasmissione a voce)",
		desc: "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT. Il livello indica la sensibilità (1 = più sensibile). OFF = VOX spento.",
		warn: "Con il VOX acceso la radio trasmette appena sente un suono: rumori ambientali o conversazioni vicine possono finire in onda.",
		opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...)},
	{field: "cbB_VoxDelay", key: "vox_delay", group: "Ricezione e trasmissione", label: "Ritardo del VOX",
		desc: "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare (evita di tagliare la fine delle frasi).",
		opts: func() []string {
			o := make([]string, 16)
			for i := range o {
				o[i] = strconv.FormatFloat(0.5+float64(i)*0.1, 'f', 1, 64) + " s"
			}
			return o
		}()},
	{field: "cbB_BusyLockout", key: "busy_lock", group: "Ricezione e trasmissione", label: "Blocco su canale occupato",
		desc: "Se attivo, la radio non trasmette mentre sul canale c'è già un segnale: evita di sovrapporsi a chi sta parlando.", bool: true},
	{field: "cbB_TailNoiseClear", key: "tail", group: "Ricezione e trasmissione", label: "Eliminazione della coda di rumore",
		desc: "Elimina il breve rumore («squelch tail») che si sente quando l'altra radio smette di trasmettere (funziona tra radio che usano la stessa funzione).", bool: true},
	{field: "cbB_PassRepetNoiseClear", key: "rp_ste", group: "Ricezione e trasmissione", label: "Eliminazione della coda del ponte (RP-STE)",
		desc: "Elimina la coda di rumore dei ponti ripetitori: durata del tono di eliminazione inviato. OFF = funzione spenta.",
		opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...)},
	{field: "cbB_PassRepetNoiseDetect", key: "rpt_rl", group: "Ricezione e trasmissione", label: "Rilevamento della coda del ponte (RPT-RL)",
		desc: "Ritardo con cui la radio riconosce la coda di un ponte ripetitore per eliminarla. OFF = funzione spenta.",
		opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...)},
	{field: "cbB_RTone", key: "r_tone", group: "Ricezione e trasmissione", label: "Tono di chiamata per i ponti",
		desc: "Frequenza del tono di apertura inviato ai ponti ripetitori che lo richiedono (in Europa di solito 1750 Hz).",
		opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}},
	{field: "cbB_SoundTxEnd", key: "tx_end", group: "Ricezione e trasmissione", label: "Segnale di fine trasmissione",
		desc: "Suono inviato in aria quando rilasci il PTT (roger beep): nessuno, un beep o un tono a 1200 Hz.",
		opts: []string{"OFF", "Beep", "Tono 1200 Hz"}},
	{field: "cbB_DenoiseSwitch", key: "denoise", group: "Ricezione e trasmissione", label: "Riduzione del rumore",
		desc: "Riduce il rumore di fondo dell'audio ricevuto.", bool: true},

	// ---- doppio ascolto e modalità
	{field: "cbB_TDR", key: "tdr", group: "Doppio ascolto e modalità", label: "Doppio ascolto (TDR)",
		desc: "La radio ascolta alternativamente le due bande A e B.", bool: true},
	{field: "cbB_TDRTx", key: "tdr_tx", group: "Doppio ascolto e modalità", label: "Banda di trasmissione in doppio ascolto",
		desc: "In doppio ascolto, su quale banda trasmettere: quella in cui è arrivato l'ultimo segnale (OFF) o sempre la A o la B.",
		opts: []string{"OFF (ultima attiva)", "Banda A", "Banda B"}},
	{field: "cbB_CurWorkMode", key: "cur_band", group: "Doppio ascolto e modalità", label: "Banda attiva",
		desc: "Banda selezionata all'accensione.", opts: []string{"Banda A", "Banda B"}},
	{field: "cbB_WorkModeA", key: "mode_a", group: "Doppio ascolto e modalità", label: "Modalità della banda A",
		desc: "Frequenza libera (VFO) o canali memorizzati.", opts: []string{"Frequenza (VFO)", "Canali"}},
	{field: "cbB_WorkModeB", key: "mode_b", group: "Doppio ascolto e modalità", label: "Modalità della banda B",
		desc: "Frequenza libera (VFO) o canali memorizzati.", opts: []string{"Frequenza (VFO)", "Canali"}},
	{field: "cbB_DivideCH", key: "divide_ch", group: "Doppio ascolto e modalità", label: "Divisione dei canali",
		desc: "Canali divisi in 15 zone da 64 oppure tutti i 960 canali in un unico elenco (impostazione presente dalla versione 1.0.9 del programma BT-11_CPS).",
		warn: "Cambia il modo in cui la radio mostra e scorre i canali: con «tutti i canali» le zone non si usano più.",
		opts: []string{"Zone (15 × 64)", "Tutti i canali (960)"}},
	{field: "cbB_Scan", key: "scan_mode", group: "Doppio ascolto e modalità", label: "Modalità di scansione",
		desc: "Come riprende la scansione dopo un segnale: TO dopo qualche secondo, CO quando il segnale finisce, SE si ferma sul segnale trovato.",
		opts: []string{"TO (a tempo)", "CO (a portante)", "SE (si ferma)"}},
	{field: "cbB_FMRadio", key: "fm_radio", group: "Doppio ascolto e modalità", label: "Radio FM broadcast",
		desc: "Permette di ascoltare la radio FM commerciale.", bool: true, inv: true},
	{field: "cbB_WeatherCH", key: "weather_ch", group: "Doppio ascolto e modalità", label: "Canale meteo (NOAA)",
		desc: "Canale del servizio meteo NOAA (solo USA/Canada; in Italia non è usato).", opts: strNums(10, 1, 1, "")},

	// ---- display, suoni e tasti
	{field: "cbB_DisplayModeA", key: "disp_a", group: "Display, suoni e tasti", label: "Display della banda A",
		desc: "Cosa mostra il display per i canali della banda A.", opts: []string{"Nome", "Frequenza", "Numero del canale"}},
	{field: "cbB_DisplayModeB", key: "disp_b", group: "Display, suoni e tasti", label: "Display della banda B",
		desc: "Cosa mostra il display per i canali della banda B.", opts: []string{"Nome", "Frequenza", "Numero del canale"}},
	{field: "cbB_AutoBacklight", key: "backlight", group: "Display, suoni e tasti", label: "Spegnimento della retroilluminazione",
		desc: "Dopo quanto tempo senza usare la radio si spegne la luce del display (Sempre accesa = non si spegne).",
		opts: []string{"Sempre accesa", "5 s", "10 s", "15 s", "20 s", "30 s", "1 min", "2 min", "3 min"}},
	{field: "cbB_PowerMsg", key: "power_msg", group: "Display, suoni e tasti", label: "Schermata di accensione",
		desc: "Cosa mostra la radio all'accensione: il logo predefinito o la tensione della batteria.", opts: []string{"Logo predefinito", "Tensione della batteria"}},
	{field: "cbB_Language", key: "lang", group: "Display, suoni e tasti", label: "Lingua",
		desc: "Lingua dei menu e delle voci della radio.", opts: []string{"Inglese", "Russo"}},
	{field: "cbB_VoicePrompt", key: "voice", group: "Display, suoni e tasti", label: "Guida vocale",
		desc: "La radio annuncia a voce i menu e i canali.", bool: true},
	{field: "cbB_BeepPrompt", key: "beep", group: "Display, suoni e tasti", label: "Suono dei tasti",
		desc: "Beep alla pressione dei tasti.", bool: true},
	{field: "cbB_TimerMenuQuit", key: "menu_time", group: "Display, suoni e tasti", label: "Uscita automatica dal menu",
		desc: "Dopo quanto tempo senza premere tasti la radio esce dal menu.", opts: strNums(11, 5, 5, " s")},
	{field: "cbB_SaveMode", key: "save", group: "Display, suoni e tasti", label: "Risparmio della batteria",
		desc: "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi.",
		opts: []string{"OFF", "Normale", "Super", "Profondo"}},
	{field: "cbB_LockKeyBoard", key: "key_lock", group: "Display, suoni e tasti", label: "Blocco della tastiera",
		desc: "Tastiera bloccata all'accensione.",
		warn: "Con la tastiera bloccata i tasti non rispondono finché non la sblocchi con la combinazione indicata nel manuale.", bool: true},
	{field: "cbB_AutoKeyLock", key: "auto_lock", group: "Display, suoni e tasti", label: "Blocco automatico della tastiera",
		desc: "Blocca la tastiera dopo il tempo indicato senza usarla.", opts: []string{"OFF", "5 s", "10 s", "15 s"}},
	{field: "cbB_KeySide2", key: "pf2_short", group: "Display, suoni e tasti", label: "Tasto laterale 2 (pressione breve)",
		desc: "Funzione del tasto laterale 2 premuto brevemente.",
		opts: []string{"Radio FM", "Potenza di trasmissione", "Scansione", "Ricerca frequenza", "Meteo (NOAA)", "Allarme (SOS)", "Analizzatore di spettro", "PTT della banda B"}},
	{field: "cbB_KeySide2L", key: "pf2_long", group: "Display, suoni e tasti", label: "Tasto laterale 2 (pressione lunga)",
		desc: "Funzione del tasto laterale 2 tenuto premuto.",
		opts: []string{"Radio FM", "Potenza di trasmissione", "Scansione", "Ricerca frequenza", "Meteo (NOAA)", "Allarme (SOS)", "Analizzatore di spettro"}},
	{field: "cbB_KeySide3", key: "pf3_short", group: "Display, suoni e tasti", label: "Tasto laterale 3 (pressione breve)",
		desc: "Funzione del tasto laterale 3 premuto brevemente.",
		opts: []string{"Radio FM", "Potenza di trasmissione", "Scansione", "Ricerca frequenza", "Meteo (NOAA)", "Allarme (SOS)", "Analizzatore di spettro"}},

	// ---- DTMF, allarme e funzioni remote
	{field: "cbB_DTMF", key: "dtmf_st", group: "DTMF, allarme e funzioni remote", label: "Ascolto dei toni DTMF trasmessi",
		desc: "Fa sentire nell'altoparlante i toni DTMF che la radio trasmette: quelli dei tasti, quelli del codice identificativo (ANI) o entrambi.",
		opts: []string{"OFF", "Tasti", "Codice ANI", "Tasti + ANI"}},
	{field: "cbB_PTTID", key: "ptt_id", group: "DTMF, allarme e funzioni remote", label: "Invio del codice identificativo (PTT-ID)",
		desc: "Quando trasmettere il codice DTMF identificativo: alla pressione del PTT, al rilascio o in entrambi i momenti.",
		warn: "Il codice viene trasmesso in aria a ogni trasmissione: sui ponti radioamatoriali e nei canali condivisi di solito non è gradito.",
		opts: []string{"OFF", "All'inizio", "Alla fine", "Inizio e fine"}},
	{field: "cbB_SendIDDelay", key: "id_delay", group: "DTMF, allarme e funzioni remote", label: "Ritardo dell'invio del codice",
		desc: "Attesa tra la pressione del PTT e l'invio del codice identificativo.",
		opts: []string{"0 ms", "100 ms", "200 ms", "400 ms", "600 ms", "800 ms", "1000 ms"}},
	{field: "cbB_AlarmMode", key: "alarm_mode", group: "DTMF, allarme e funzioni remote", label: "Modalità dell'allarme (SOS)",
		desc: "Cosa fa il tasto di allarme: suona solo sulla radio, trasmette il suono di allarme o trasmette il codice di allarme.",
		warn: "Nelle modalità che trasmettono, premere l'allarme manda un segnale in aria sul canale in uso.",
		opts: []string{"Solo sulla radio", "Trasmette il suono", "Trasmette il codice"}},
	{field: "cbB_AlarmSound", key: "alarm_sound", group: "DTMF, allarme e funzioni remote", label: "Suono dell'allarme",
		desc: "Suono di allarme acceso o spento.", bool: true},
	{field: "cbB_RemoteKillEnable", key: "remote_kill", group: "DTMF, allarme e funzioni remote", label: "Disattivazione a distanza",
		desc: "Permette di disattivare la radio a distanza con un comando DTMF (radio smarrite o rubate).",
		warn: "Con la funzione attiva chi conosce il codice può bloccare la radio via radio; per riattivarla serve di nuovo il comando o la riprogrammazione.", bool: true},
	{field: "cbB_BTRWData", key: "bt_prog", group: "DTMF, allarme e funzioni remote", label: "Programmazione via Bluetooth",
		desc: "Permette di leggere e scrivere la radio via Bluetooth (con l'app del costruttore).",
		warn: "Con la programmazione Bluetooth attiva chiunque sia vicino può tentare di leggere o modificare la radio.", bool: true},
}

// bajProp: proprietà di KDH.FunConfigData per la chiave (per il confronto con il programma)
func bajProp(key string) string {
	for _, b := range bajSetDefs {
		if b.key == key {
			return "CbB_" + b.field[4:]
		}
	}
	return ""
}

func int32Hex(v int) *string {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(int32(v)))
	s := hex.EncodeToString(b)
	return &s
}

// bajSettings: tabella delle impostazioni per il file .dat aperto (posizioni
// degli interi di KDH.FunConfigData in questo file)
func bajSettings(img []byte) *setTable {
	d, err := ParseNRBF(img)
	if err != nil {
		return nil
	}
	var fun *Rec
	for _, r := range d.Recs {
		if (r.T == 1 || r.T == 5) && d.MetaOf(r) != nil && d.MetaOf(r).Name == "KDH.FunConfigData" {
			if fun != nil {
				return nil // più di un oggetto: struttura inattesa
			}
			fun = r
		}
	}
	if fun == nil {
		return nil
	}
	m := d.MetaOf(fun)
	pos := map[string]int{}
	for i, n := range m.Names {
		v := fun.Vals[i]
		if v.Prim && v.PT == 8 && len(v.Raw) == 4 && v.Off > 0 && v.Off+4 <= len(img) {
			pos[n] = v.Off
		}
	}
	t := &setTable{Class: "KDH.FunConfigData", Size: len(img), Source: "programma BT-11_CPS del costruttore (finestra «Optional Features»)"}
	for _, b := range bajSetDefs {
		off, ok := pos[b.field]
		if !ok {
			if b.field == "cbB_DivideCH" {
				continue // file della v1.0.8: impostazione assente
			}
			return nil
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(off), "ffffffff"}}}
		sd := setDef{Key: b.key, Group: b.group, Label: b.label, ItLabel: b.label, ItDesc: b.desc, ItWarn: b.warn, Enc: enc}
		if b.bool {
			sd.Type = "bool"
			if b.inv {
				enc.Table = []*string{int32Hex(1), int32Hex(0)}
			} else {
				enc.Table = []*string{int32Hex(0), int32Hex(1)}
			}
		} else {
			sd.Type = "list"
			sd.ItOpts = b.opts
			for i := range b.opts {
				sd.Options = append(sd.Options, strconv.Itoa(i))
				enc.Table = append(enc.Table, int32Hex(i))
			}
		}
		t.Settings = append(t.Settings, sd)
	}
	return t
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == kindDAT },
		get: func(kind string, img []byte) *setTable { return bajSettings(img) },
	})
}
