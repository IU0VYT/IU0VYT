// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Nomi, descrizioni e avvisi in italiano delle impostazioni delle radio.
// Le tabelle dei programmi dei costruttori (K67, BJ-8300, BJ-9000) hanno già i
// testi italiani; per quelle ricavate da CHIRP le regole qui sotto riconoscono il
// nome inglese della voce (sono centinaia, con nomi diversi da un driver all'altro)
// e danno nome, descrizione e, per le voci delicate, l'avviso da confermare.

import (
	"regexp"
	"strings"
)

type setInfo struct {
	Label    string
	Desc     string
	Warn     string
	Critical bool
}

// avvisi delle voci delicate
const (
	wPwd = "Con una password impostata la radio non si accende, non si legge o non si programma senza il codice: se lo dimentichi " +
		"potresti non riuscire più a usarla o a riprogrammarla."
	wLimit = "Limiti di frequenza della trasmissione: allargarli oltre le bande per cui la radio è omologata e per cui sei autorizzato " +
		"è illegale e può danneggiare il trasmettitore; restringerli può impedire di trasmettere sui tuoi canali."
	wTxEn   = "Abilita o blocca la trasmissione su un'intera banda: trasmettere fuori dalle bande autorizzate è illegale; se la blocchi la radio su quella banda solo riceve."
	wTxBand = "Abilita la trasmissione su una banda non radioamatoriale: in Italia è illegale senza una specifica autorizzazione."
	wTxOff  = "Con la trasmissione disattivata la radio non trasmette più su nessun canale finché non la riattivi."
	wTOT    = "Senza limite di tempo (OFF) un PTT bloccato o premuto per errore lascia la radio in trasmissione: disturbi agli altri e possibile surriscaldamento."
	wVOX    = "Con il VOX acceso la radio trasmette appena sente un suono: rumori ambientali o conversazioni vicine possono finire in onda."
	wReset  = "Permette di riportare la radio alle condizioni di fabbrica dal menu: un reset cancella canali e impostazioni."
	wMenu   = "Nasconde o limita le voci del menu della radio: alcune regolazioni non saranno più accessibili dalla tastiera."
	wPTTID  = "Il codice viene trasmesso in aria a ogni trasmissione: sui ponti radioamatoriali e nei canali condivisi di solito non è gradito."
	wLock   = "Con la tastiera bloccata i tasti non rispondono finché non la sblocchi con la combinazione indicata nel manuale."
	wRelay  = "Come ripetitore la radio trasmette da sola, anche a lungo: usala solo dove è consentito (in Italia un ponte va autorizzato) e controlla il riscaldamento."
	wKill   = "Con la funzione attiva chi conosce il codice può bloccare la radio a distanza via radio."
	wBT     = "Con la programmazione via Bluetooth attiva chiunque sia vicino può tentare di leggere o modificare la radio."
	wSide   = "Disattivando i tasti laterali si perdono le funzioni assegnate (es. monitor o torcia)."
	wAlarm  = "Nelle modalità che trasmettono, l'allarme manda un segnale in aria sul canale in uso."
	wVolt   = "Con il blocco disattivato la radio trasmette anche con la tensione fuori dai limiti, a rischio di danni o di emissioni non pulite."
)

type itRule struct {
	re                *regexp.Regexp
	label, desc, warn string
}

func itR(re, label, desc, warn string) itRule {
	return itRule{regexp.MustCompile("(?i)^(?:" + re + ")$"), label, desc, warn}
}

// regole sul nome CHIRP della voce (la prima che corrisponde vale); $1… = parti
// tra parentesi del nome
var itRules = []itRule{
	// --- protezione
	itR(`power on password switch`, "Password all'accensione", "Chiede una password quando si accende la radio.", wPwd),
	itR(`read protect`, "Protezione della lettura", "Impedisce di leggere la radio con il cavo senza la password.", wPwd),
	itR(`write protect`, "Protezione della scrittura", "Impedisce di programmare la radio con il cavo senza la password.", wPwd),
	// --- limiti di banda e trasmissione
	itR(`(vhf2?|uhf) (lower|low) limit( \(mhz\))?`, "Limite inferiore di trasmissione $1 (MHz)", "Frequenza più bassa su cui la radio accetta di trasmettere nella banda $1.", wLimit),
	itR(`(vhf2?|uhf) (upper|high) limit( \(mhz\))?`, "Limite superiore di trasmissione $1 (MHz)", "Frequenza più alta su cui la radio accetta di trasmettere nella banda $1.", wLimit),
	itR(`(vhf|uhf) tx enabled`, "Trasmissione abilitata in $1", "Permette la trasmissione nella banda $1.", wTxEn),
	itR(`(220|520)m tx`, "Trasmissione a $1 MHz", "Permette di trasmettere nella banda dei $1 MHz.", wTxBand),
	itR(`disable tx`, "Trasmissione disattivata", "Blocca del tutto la trasmissione: la radio funziona solo come ricevitore.", wTxOff),
	itR(`(high|low) voltage inhibit (tx|transmit)`, "Blocco della trasmissione con tensione $1", "Impedisce di trasmettere quando la tensione della batteria è troppo $1 (protegge la radio).", wVolt),
	// --- squelch, VOX, tempi di trasmissione
	itR(`(carrier )?squelch( level)?`, "Livello squelch", "Soglia di apertura dell'audio: più alto il livello, più forte deve essere un segnale per essere ascoltato. A 0 l'audio resta sempre aperto (si sente il fruscio).", ""),
	itR(`vox( switch)?`, "VOX (trasmissione a voce)", "Con il VOX la radio trasmette da sola quando sente la voce, senza premere il PTT. Se è un livello: OFF = spento, numeri più bassi = più sensibile.", wVOX),
	itR(`vox (level|gain|sensitivity)`, "Sensibilità del VOX", "Livello di voce che fa partire la trasmissione con il VOX (di solito numeri bassi = più sensibile; OFF, se c'è, spegne il VOX).", wVOX),
	itR(`vox delay( time)?( \(old \| new\))?`, "Ritardo del VOX", "Quanto tempo la radio continua a trasmettere dopo che hai smesso di parlare.", ""),
	itR(`(rx disable vox|inhibit vox on receive)`, "VOX disattivato durante la ricezione", "Il VOX non fa trasmettere mentre la radio sta ricevendo un segnale.", ""),
	itR(`(time[- ]?out timer|timeout timer|tot|timeout)( \(s(econds)?\))?`, "Tempo massimo di trasmissione (TOT)", "Dopo questo tempo di trasmissione continua la radio smette di trasmettere da sola.", wTOT),
	itR(`timeout timer alarm`, "Preavviso del TOT", "Beep di avviso qualche secondo prima che scatti il tempo massimo di trasmissione.", ""),
	itR(`(busy channel lockout|busy lockout|bcl)`, "Blocco su canale occupato", "Se attivo, la radio non trasmette mentre sul canale c'è già un segnale.", ""),
	// --- code e ponti
	itR(`(squelch tail eliminate( \(ht to ht\))?|tail clear|ste)`, "Eliminazione della coda di rumore", "Elimina il breve rumore che si sente quando l'altra radio smette di trasmettere (tra radio con la stessa funzione).", ""),
	itR(`squelch tail eliminate \(repeater\)`, "Eliminazione della coda del ponte (RP-STE)", "Elimina la coda di rumore dei ponti ripetitori.", ""),
	itR(`(ste repeater delay|rpt tail delay|repeater tail delay( \(ms\))?)`, "Ritardo della coda del ponte (RPT-RL)", "Ritardo con cui la radio riconosce ed elimina la coda di un ponte ripetitore.", ""),
	itR(`(rpt tail clear|repeater tail revert( \(ms\))?)`, "Eliminazione della coda del ponte", "Durata del tono che elimina la coda di rumore dei ponti ripetitori.", ""),
	itR(`ctcss tail revert`, "Inversione di fase del CTCSS a fine trasmissione", "Tecnica per eliminare la coda di rumore con i toni CTCSS.", ""),
	itR(`(tone burst|tone-burst frequency|pilot tone)`, "Tono di chiamata per i ponti", "Frequenza del tono di apertura inviato ai ponti che lo richiedono (in Europa di solito 1750 Hz).", ""),
	itR(`relay mode`, "Modalità ripetitore", "La radio ritrasmette i segnali ricevuti (funzione di ponte ripetitore).", wRelay),
	itR(`(roger|roger beep)`, "Segnale di fine trasmissione (roger beep)", "Suono inviato in aria quando rilasci il PTT.", ""),
	// --- doppio ascolto e modalità
	itR(`(dual watch|dual standby|tdr - dual frequency standby|tdr)`, "Doppio ascolto", "La radio ascolta alternativamente le due bande A e B.", ""),
	itR(`(dual watch|dual standby) tx priority|tdr tx|tx-a/b`, "Banda di trasmissione in doppio ascolto", "In doppio ascolto, su quale banda trasmettere (quella dell'ultimo segnale ricevuto o sempre la stessa).", ""),
	itR(`dual watch channel`, "Canale del doppio ascolto", "Canale controllato in doppio ascolto.", ""),
	itR(`single (watch|mode)`, "Ascolto singolo", "Mostra e ascolta una sola banda.", ""),
	itR(`(vfo/mr mode|work mode|frequency mode|radio work mode\((a|b)\)|channel (a|b) work mode)`, "Modalità di lavoro $2$3", "Frequenza libera (VFO) o canali memorizzati.", ""),
	itR(`radio operating mode`, "Modalità di funzionamento", "Modalità di funzionamento della radio.", ""),
	itR(`vfo/mr button`, "Tasto VFO/MR", "Abilita il tasto che passa tra frequenza libera e canali.", ""),
	itR(`(mr|mem channel) ?\(?(a|b)\)?( channel)?( #)?`, "Canale attivo della banda $2", "Canale selezionato sulla banda $2.", ""),
	itR(`mr (a|b) channel( #)?`, "Canale attivo della banda $1", "Canale selezionato sulla banda $1.", ""),
	itR(`active channel`, "Canale attivo", "Canale selezionato.", ""),
	itR(`(scan resume( method)?|scan mode|scan type|scan)`, "Modalità di scansione", "Come riprende la scansione dopo un segnale: TO dopo qualche secondo, CO quando il segnale finisce, SE si ferma sul segnale trovato.", ""),
	itR(`scan link zone (\d+)`, "Scansione collegata alla zona $1", "Include la zona $1 nella scansione delle zone collegate.", ""),
	itR(`scan zone linking`, "Scansione su più zone", "La scansione passa da una zona all'altra.", ""),
	itR(`qt save (mode|type)`, "Salvataggio del tono trovato con la scansione", "Dove salvare il tono CTCSS/DCS trovato con la scansione dei toni (ricezione, trasmissione o entrambe).", ""),
	itR(`(broadcast fm radio|fm radio|fm function|fm/am/ssb radio)`, "Radio FM broadcast", "Permette di ascoltare la radio FM commerciale.", ""),
	itR(`disable fm radio`, "Radio FM disattivata", "Disattiva la radio FM commerciale.", ""),
	itR(`fm interrupt`, "Interruzione della radio FM", "Un segnale sul canale interrompe l'ascolto della radio FM.", ""),
	itR(`fm range`, "Gamma della radio FM", "Gamma di frequenze della radio FM commerciale.", ""),
	itR(`fm work mode`, "Modalità della radio FM", "Frequenza libera o canali memorizzati per la radio FM.", ""),
	itR(`noaa wx radio`, "Radio meteo NOAA", "Ascolto del servizio meteo NOAA (solo USA/Canada).", ""),
	itR(`area`, "Area geografica", "Zona per cui è impostata la radio (piani di frequenza diversi).", ""),
	// --- display, luci, suoni
	itR(`(a|b|channel (a|b)) ?channel display type|display mode \((a|b)\)|memory display format (a|b)|(a|b) channel display type|display format\(f(1|2)\)`, "Display dei canali $1$2$3$4$5$6", "Cosa mostra il display per i canali: nome, frequenza o numero.", ""),
	itR(`channel (a|b) display type`, "Display della banda $1", "Cosa mostra il display per i canali della banda $1.", ""),
	itR(`(display|display mode|selected display line)`, "Display", "Modalità di visualizzazione del display.", ""),
	itR(`display ani`, "Mostra il codice ANI", "Mostra sul display il codice identificativo ricevuto.", ""),
	itR(`(backlight|back light|auto backlight|backlight timeout|backlight timer)`, "Retroilluminazione", "Dopo quanto tempo si spegne la luce del display (o se resta sempre accesa).", ""),
	itR(`invert lcd`, "Display invertito", "Inverte i colori del display.", ""),
	itR(`lcd contrast`, "Contrasto del display", "Contrasto del display.", ""),
	itR(`(rx|tx|standby) led color`, "Colore del display: $1", "Colore della retroilluminazione in ricezione (RX), in trasmissione (TX) o in attesa (standby).", ""),
	itR(`(power[- ]on message|power on message)( (\d))?`, "Messaggio di accensione $3", "Testo mostrato all'accensione.", ""),
	itR(`6\+power-on message (\d)`, "Messaggio di accensione $1 (6+)", "Testo mostrato all'accensione (firmware recenti).", ""),
	itR(`boot screen (\d)`, "Messaggio di accensione $1", "Testo mostrato all'accensione.", ""),
	itR(`(boot screen mode|power on display type)`, "Schermata di accensione", "Cosa mostra la radio all'accensione (logo, messaggio o tensione).", ""),
	itR(`firmware message (\d)`, "Messaggio del firmware $1", "Testo del firmware mostrato all'accensione (sola lettura).", ""),
	itR(`power on tone`, "Suono di accensione", "Suono all'accensione.", ""),
	itR(`power on time`, "Durata della schermata di accensione", "Per quanto tempo resta la schermata di accensione.", ""),
	itR(`(beep|key beep|beep tone|beep prompt)`, "Suono dei tasti", "Beep alla pressione dei tasti.", ""),
	itR(`(voice|voice prompts?|voice alert|enable voice)`, "Guida vocale", "La radio annuncia a voce i menu e i canali.", ""),
	itR(`(language|voice language)`, "Lingua", "Lingua dei menu e/o della guida vocale.", ""),
	itR(`volume level`, "Volume", "Volume.", ""),
	itR(`am rx`, "Ricezione AM", "Ricezione in AM (banda aeronautica).", ""),
	itR(`(alarm sound|sound alarm|alarm)`, "Suono dell'allarme", "Suono di allarme (tasto di emergenza).", ""),
	itR(`alarm (mode|time|code)`, "Allarme: $1", "Funzionamento del tasto di allarme (solo sulla radio o con trasmissione di un suono o di un codice).", wAlarm),
	itR(`rx emergency`, "Ricezione delle emergenze", "Reagisce ai segnali di emergenza ricevuti.", ""),
	itR(`ring`, "Squillo", "Durata dello squillo alla ricezione di una chiamata selettiva.", ""),
	itR(`(save mode|battery save( mode)?|battery saver)`, "Risparmio della batteria", "La radio spegne il ricevitore per brevi intervalli quando il canale è libero: più risparmio, ma i primi istanti di una chiamata possono andare persi.", ""),
	itR(`noise reduction`, "Riduzione del rumore", "Riduce il rumore di fondo dell'audio ricevuto.", ""),
	itR(`hang-up time`, "Tempo di riaggancio", "Tempo dopo il quale termina una chiamata selettiva.", ""),
	// --- tasti e menu
	itR(`(key lock|keypad lock|kb lock)`, "Blocco della tastiera", "Blocca la tastiera.", wLock),
	itR(`(auto key lock|automatic key lock|key auto lock)`, "Blocco automatico della tastiera", "Blocca la tastiera dopo qualche secondo senza usarla.", ""),
	itR(`(menu auto quit|menu quit timer)`, "Uscita automatica dal menu", "Dopo quanto tempo senza premere tasti la radio esce dal menu.", ""),
	itR(`(all menus|menu)`, "Menu completo", "Abilita tutte le voci del menu della radio.", wMenu),
	itR(`(reset menu|enable menu rst)`, "Reset dal menu", "Permette il reset della radio dal menu.", wReset),
	itR(`side key disable`, "Tasti laterali disattivati", "Disattiva i tasti laterali.", wSide),
	itR(`(pf|skey) ?(\d)( key)?( \((short|long) press\)| (long|short))?`, "Tasto laterale $2 $5$6", "Funzione del tasto laterale $2 (pressione breve o lunga).", ""),
	itR(`pf(\d) long`, "Tasto laterale $1 (pressione lunga)", "Funzione del tasto laterale $1 tenuto premuto.", ""),
	itR(`side key( function)?`, "Tasto laterale", "Funzione del tasto laterale.", ""),
	itR(`input dtmf`, "Inserimento DTMF", "Permette di inserire toni DTMF dalla tastiera.", ""),
	itR(`(bluetooth)`, "Bluetooth", "Bluetooth della radio (auricolari o app).", wBT),
	itR(`gps on`, "GPS", "Ricevitore GPS acceso.", ""),
	itR(`gps (mode|speed units|timezone)`, "GPS: $1", "Impostazione del ricevitore GPS.", ""),
	itR(`stationid`, "Identificativo di stazione", "Identificativo trasmesso dalla radio.", ""),
	// --- DTMF e codici
	itR(`(ptt[- ]id|when to send ptt id)`, "Invio del codice identificativo (PTT-ID)", "Quando trasmettere il codice DTMF identificativo: alla pressione del PTT (BOT), al rilascio (EOT) o in entrambi i momenti.", wPTTID),
	itR(`(ptt[- ]id code|ptt id code|signal code|dtmf code) ?(\d*)`, "Codice DTMF $2", "Sequenza DTMF memorizzata (codice identificativo o di chiamata).", ""),
	itR(`signal code (\d+) name`, "Nome del codice DTMF $1", "Nome del codice DTMF $1.", ""),
	itR(`(ptt id delay|ptt delay|send id delay)`, "Ritardo dell'invio del codice", "Attesa tra la pressione del PTT e l'invio del codice identificativo.", ""),
	itR(`(ani code|ani id|radio id)`, "Codice identificativo (ANI)", "Codice DTMF che identifica questa radio.", ""),
	itR(`when to send ani id`, "Invio del codice ANI", "Quando trasmettere il codice identificativo.", wPTTID),
	itR(`(dtmf side ?tone|side tone)`, "Ascolto dei toni DTMF trasmessi", "Fa sentire nell'altoparlante i toni DTMF che la radio trasmette.", ""),
	itR(`dtmf speed \((on|off)\)`, "Velocità DTMF (durata $1)", "Durata dei toni (on) e delle pause (off) DTMF.", ""),
	itR(`(digit duration|digit interval duration|delay time) \(ms\)`, "DTMF: $1 (ms)", "Tempi dei toni DTMF trasmessi.", ""),
	itR(`(up|down) code`, "Codice DTMF di $1 trasmissione", "Codice inviato all'inizio o alla fine della trasmissione.", ""),
	itR(`(group call code|separate code|alarm code)`, "DTMF: $1", "Codice DTMF speciale (chiamata di gruppo, separatore, allarme).", ""),
	itR(`ptt (press|release) send`, "Codice inviato alla $1 del PTT", "Codice DTMF inviato alla pressione o al rilascio del PTT.", wPTTID),
	itR(`entry (\d+)`, "Codice DTMF $1", "Codice DTMF memorizzato.", ""),
	itR(`bank name (\d+)`, "Nome del banco $1", "Nome del banco di canali $1.", ""),
	// --- VFO
	itR(`vfo (a|b) (bandwidth|power|shift|tuning step|ptt-id|s-code)`, "VFO $1: $2", "Impostazione del VFO $1 (frequenza libera).", ""),
	itR(`(bandwidth|power|tx power|shift|offset direction|tuning step|step freq|s-code|signal code|rx ctcss/dcs|tx ctcss/dcs|rx modulation|scramble|rx gain|modulation mode)`, "$1", "Impostazione del canale o della frequenza libera (VFO) indicata dal gruppo.", ""),
	itR(`voice privacy - subtone encryption`, "Cifratura dei subtoni", "Cifratura semplice dei toni.", ""),
}

// traduzione delle parti del nome inglese riportate nelle regole ($1, $2…)
var itPhrases = map[string]string{
	"mode": "modalità", "time": "durata", "code": "codice", "speed units": "unità di velocità", "timezone": "fuso orario",
	"high": "alta", "low": "bassa", "up": "inizio", "down": "fine", "press": "pressione", "release": "rilascio",
	"short": "(pressione breve)", "long": "(pressione lunga)", "bandwidth": "larghezza di banda", "power": "potenza",
	"tx power": "potenza", "shift": "scostamento", "offset direction": "direzione dello scostamento", "tuning step": "passo",
	"step freq": "passo", "s-code": "codice PTT-ID", "signal code": "codice PTT-ID", "ptt-id": "invio del codice PTT-ID",
	"rx ctcss/dcs": "tono in ricezione", "tx ctcss/dcs": "tono in trasmissione", "rx modulation": "modulazione in ricezione",
	"scramble": "scrambler", "rx gain": "guadagno in ricezione", "modulation mode": "modo di modulazione",
	"digit duration": "durata delle cifre", "digit interval duration": "pausa tra le cifre", "delay time": "ritardo iniziale",
	"group call code": "codice di chiamata di gruppo", "separate code": "codice separatore", "alarm code": "codice di allarme",
	"rx": "ricezione", "tx": "trasmissione", "standby": "attesa", "on": "ON", "off": "OFF",
}

var reDollar = regexp.MustCompile(`\$(\d)`)

func itExpand(tmpl, s string, m []int) string {
	return reDollar.ReplaceAllStringFunc(tmpl, func(x string) string {
		i := int(x[1] - '0')
		if 2*i+1 >= len(m) || m[2*i] < 0 {
			return ""
		}
		g := strings.TrimSpace(s[m[2*i]:m[2*i+1]])
		if t, ok := itPhrases[strings.ToLower(g)]; ok {
			return t
		}
		return g
	})
}

func setInfoFor(d *setDef) setInfo {
	if d.ItLabel != "" { // tabelle dei programmi dei costruttori: testi già in italiano
		return setInfo{Label: d.ItLabel, Desc: d.ItDesc, Warn: d.ItWarn, Critical: d.ItWarn != ""}
	}
	lab := strings.TrimSpace(d.Label)
	low := strings.ToLower(lab)
	switch {
	case strings.Contains(low, "read password"):
		return setInfo{"Password di lettura", "Password richiesta per leggere la radio con il cavo.", wPwd, true}
	case strings.Contains(low, "write password"):
		return setInfo{"Password di scrittura", "Password richiesta per programmare la radio con il cavo.", wPwd, true}
	}
	for _, ru := range itRules {
		m := ru.re.FindStringSubmatchIndex(lab)
		if m == nil || ru.label == "" {
			continue
		}
		l := strings.Join(strings.Fields(itExpand(ru.label, lab, m)), " ")
		ds := strings.Join(strings.Fields(itExpand(ru.desc, lab, m)), " ")
		if r := []rune(l); len(r) > 0 {
			l = strings.ToUpper(string(r[0])) + string(r[1:])
		}
		if strings.HasPrefix(d.Key, "vfo") && !strings.Contains(strings.ToLower(l), "vfo") {
			l += " (VFO)"
		}
		info := setInfo{Label: l, Desc: ds + " Nome in CHIRP: «" + lab + "».", Warn: ru.warn}
		if d.Warn != "" && info.Warn == "" {
			info.Warn = "Avviso di CHIRP: " + d.Warn
		}
		info.Critical = info.Warn != ""
		return info
	}
	info := setInfo{Label: lab, Desc: "Voce «" + lab + "» delle impostazioni della radio (nome usato da CHIRP)."}
	if d.Warn != "" {
		info.Warn, info.Critical = "Avviso di CHIRP: "+d.Warn, true
	}
	return info
}

var itGroups = map[string]string{
	"basic settings": "Impostazioni di base", "advanced settings": "Impostazioni avanzate", "other settings": "Altre impostazioni",
	"general settings": "Impostazioni generali", "work mode settings": "Modalità di lavoro", "dtmf settings": "DTMF",
	"dtmf encode settings": "Codici DTMF", "dtmf list": "Elenco dei codici DTMF", "dtmf": "DTMF", "fm radio": "Radio FM",
	"bank names": "Nomi dei banchi", "protect": "Protezione", "vfo a/b channel": "VFO A/B", "vfo a channel": "VFO A", "vfo b channel": "VFO B",
	"vfo a": "VFO A", "vfo b": "VFO B", "fm/am/ssb modulation": "Radio FM/AM/SSB", "am settings": "AM", "ssb settings": "SSB", "modes": "Modalità",
	"bj7800 specific": "Funzioni della BJ-7800", "rt-910 specific": "Funzioni della RT-910 / BJ-7700", "key assignment": "Tasti", "service": "Servizio",
}

func setGroupName(g string) string {
	if g == "" {
		return "Impostazioni"
	}
	parts := strings.Split(g, "/")
	for i, p := range parts {
		if t, ok := itGroups[strings.ToLower(strings.TrimSpace(p))]; ok {
			parts[i] = t
		}
	}
	// «FM/AM/SSB Modulation» è un nome solo con le barre
	out := strings.Join(parts, " / ")
	out = strings.ReplaceAll(out, "FM / AM / SSB Modulation", "Radio FM/AM/SSB")
	return out
}

var itOpts = map[string]string{
	"off": "OFF", "on": "ON", "disabled": "Disattivato", "enabled": "Attivo", "low": "Bassa", "high": "Alta", "mid": "Media", "medium": "Media",
	"middle": "Media", "narrow": "Stretta (12,5 kHz)", "wide": "Larga (25 kHz)", "narrow (12.5 khz)": "Stretta (12,5 kHz)", "wide (25 khz)": "Larga (25 kHz)",
	"frequency": "Frequenza", "channel": "Canale", "name": "Nome", "channel number": "Numero del canale", "english": "Inglese", "chinese": "Cinese",
	"both": "Entrambi", "bot": "All'inizio (BOT)", "eot": "Alla fine (EOT)", "to": "TO (a tempo)", "co": "CO (a portante)", "se": "SE (si ferma)",
	"time": "Tempo (TO)", "carrier": "Portante (CO)", "search": "Ricerca (SE)", "time (to)": "Tempo (TO)", "carrier (co)": "Portante (CO)", "search (se)": "Ricerca (SE)",
	"blue": "Blu", "orange": "Arancione", "purple": "Viola", "red": "Rosso", "green": "Verde", "white": "Bianco", "yellow": "Giallo",
	"none": "Nessuno", "all": "Tutti", "default": "Predefinito", "bright": "Sempre accesa", "always on": "Sempre accesa", "voltage": "Tensione",
	"batt voltage": "Tensione della batteria", "logo": "Logo", "message": "Messaggio", "full": "Completo", "local": "Locale", "remote": "Remoto",
	"local alarm": "Allarme locale", "remote alarm": "Allarme a distanza", "on site": "Solo sulla radio", "send sound": "Trasmette il suono",
	"send code": "Trasmette il codice", "send tone": "Trasmette il tono", "site": "Sul posto", "tone": "Tono", "code": "Codice", "key": "Tasto",
	"radio": "Radio FM", "fm radio": "Radio FM", "scan": "Scansione", "monitor": "Monitor", "flashlight": "Torcia", "tx power": "Potenza",
	"transmit power": "Potenza", "high/low power": "Potenza alta/bassa", "spectrum": "Analizzatore di spettro", "zone select": "Scelta della zona",
	"freq. mode": "Frequenza (VFO)", "frequency mode": "Frequenza (VFO)", "vfo mode": "Frequenza (VFO)", "ch. mode": "Canali", "ch mode": "Canali",
	"channel mode": "Canali", "name + frequency": "Nome + frequenza", "name + number": "Nome + numero", "freq. + number": "Frequenza + numero",
	"current frequency": "Frequenza attuale", "single display": "Display singolo", "dual display, dual watch": "Display doppio, doppio ascolto",
	"dual display, single watch": "Display doppio, ascolto singolo", "normal mode": "Normale", "super mode": "Super", "deep mode": "Profondo",
	"super": "Super", "rx": "Ricezione", "tx": "Trasmissione", "ptt only": "Solo PTT", "sk only": "Solo tasto laterale", "sk + ptt": "Tasto laterale + PTT",
	"kb side tone": "Tasti", "ani side tone": "Codice ANI", "kb + ani side tone": "Tasti + ANI", "dt-st": "Tasti", "ani-st": "Codice ANI", "dt+ani": "Tasti + ANI",
	"general": "Generale", "factory": "Fabbrica", "beep": "Beep", "tone1200": "Tono 1200 Hz", "voice": "Voce", "station id": "Identificativo di stazione",
	"off (if supported by radio)": "OFF (se la radio lo prevede)",
}

var reOptUnit = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?) ?(sec|secs|seconds|s|min|mins|minutes|minute|h|hour|hours)$`)
var reOptLevel = regexp.MustCompile(`(?i)^level (\d+)$`)

func setOptLabel(o string) string {
	if t, ok := itOpts[strings.ToLower(strings.TrimSpace(o))]; ok {
		return t
	}
	if m := reOptUnit.FindStringSubmatch(strings.TrimSpace(o)); m != nil {
		u := strings.ToLower(m[2])
		switch {
		case strings.HasPrefix(u, "s"):
			u = "s"
		case strings.HasPrefix(u, "m"):
			u = "min"
		default:
			u = "h"
		}
		return strings.ReplaceAll(m[1], ".", ",") + " " + u
	}
	if m := reOptLevel.FindStringSubmatch(strings.TrimSpace(o)); m != nil {
		return "Livello " + m[1]
	}
	return o
}
