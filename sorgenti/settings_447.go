// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni aggiunte nella 4.47 alle radio che avevano solo i canali.
//
// Famiglia AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II, BTECH DMR-6X2 e DMR-6X2
// PRO). Posizioni e valori riscritti da zero dalla descrizione della memoria di
// qdmr (anytone_codeplug.hh, d868uv_codeplug.hh, d878uv_codeplug.hh,
// dmr6x2uv_codeplug.hh: GeneralSettingsElement, BootSettingsElement,
// RadioIDElement; GNU GPL v3, Hannes Matuschek DM3MAT) e di dmrconfig
// (anytone_ht.c: general_settings_t, radioid_t; licenza BSD, Serge Vakulenko
// KK6ABQ). Si usano solo le voci che in qdmr hanno la stessa posizione nei cinque
// modelli; fanno eccezione il tono dei tasti e il tono di canale libero, che
// nelle DMR-6X2 sono scambiati di posto (0x36 e 0x00) e qui hanno la posizione
// del loro modello. Dove c'è anche dmrconfig (ID, nome, righe di accensione,
// schermata di accensione a 0x06, modo di lavoro a 0x34) i due concordano.
//   impostazioni generali: indirizzo 0x02500000 (immagine 0x071600)
//   righe di accensione:   indirizzo 0x02500600 (immagine 0x071C00), 2 × 16 byte
//   ID della radio n. 1:   indirizzo 0x02580000 (immagine 0x073D00): ID in BCD a
//                          8 cifre, un byte a 0, nome di 16 caratteri
// Le password (0x02500620) non si toccano.
//
// Radtel RT-880 / RT-950 Pro con il firmware nicFW880 / nicFW950: struttura
// «settingsInfo» pubblicata dall'autore (nicsure/nicFW880_950_Docs, pagine
// «Data Structures», «Squelch Menu», «Display Menu», «VFO Menu», «Advanced
// Menu»). Tutti i campi sono numeri di 16 bit, dal byte basso; si usano solo le
// voci di cui la documentazione dà i valori ammessi, e solo se il blocco
// comincia con il numero di controllo 0xABC6 (versioni con un'altra struttura
// hanno un altro numero).
//
// TID TD-UV68: è la TD-H8 con un altro nome (stessa risposta «P31183» e stessa
// mappa della memoria nella classe UV68 di CHIRP, che non ridefinisce nulla della
// classe TDH8). In CHIRP la scheda delle impostazioni della UV68 non si apre per
// un errore del driver (il modello manca dall'elenco H8_LIST e il codice cerca
// campi della TD-H3): qui si usa la tabella della classe TDH8.

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

// ---------------------------------------------------------------- AnyTone

const (
	atOffBoot = atOffSettings + 0x600 // righe di accensione
	// parti della memoria che le impostazioni possono cambiare
	atSetLen   = 0x60 // impostazioni generali usate qui (0x00-0x57)
	atBootLen  = 0x20
	atRadioLen = 0x20
	atAddrSet  = 0x02500000
	atAddrBoot = 0x02500600
	atAddrID   = 0x02580000
)

func atSettings(x6 bool) *setTable {
	gID, gOp, gTone, gDisp := "Radio e identificativo DMR", "Funzionamento", "Toni e segnalazioni", "Display"
	at := func(off int) int { return atOffSettings + off }
	keyTone, idleTone := 0x00, 0x36
	if x6 {
		keyTone, idleTone = 0x36, 0x00
	}
	lvl := func(n int) []string { return append([]string{"Spento"}, ijvNums(1, n, 1, "")...) }
	t := &setTable{Class: "ANYTONE", Size: atSize,
		Source: "descrizione della memoria di qdmr e dmrconfig (impostazioni generali della famiglia AnyTone AT-D868UV / AT-D878UV)"}
	t.Settings = []setDef{
		{Key: "radio_id", Group: gID, Label: "ID DMR della radio", Type: "int", Min: 1, Max: 16776415, ItLabel: "ID DMR della radio",
			ItDesc: "Numero DMR personale (ID radio), assegnato da RadioID.net al nominativo.",
			Enc:    &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(atOffRadioID), "ffffffff"}}, Endian: "be", Fmt: "bcd", Num: 1, Den: 1}},
	}
	// testi: un byte per carattere, il resto a 0 (come qdmr e dmrconfig)
	str := func(key, label, desc string, off, n int) setDef {
		d := rdtyStr(key, gID, label, desc, off, n)
		d.Enc.Pad = &zeroPad
		return d
	}
	t.Settings = append(t.Settings,
		str("radio_name", "Nome della radio", "Nome mostrato dalla radio (di solito il nominativo), fino a 16 caratteri.", atOffRadioID+5, 16),
		str("intro1", "Testo di accensione, riga 1", "", atOffBoot, 14),
		str("intro2", "Testo di accensione, riga 2", "", atOffBoot+16, 14))
	return setTableAdd(t, []ijvSetDef{
		{off: at(0x06), opts: []string{"Logo", "Testo (le due righe)", "Immagine"}, key: "boot_disp", group: gID, label: "Schermata di accensione"},
		{off: at(0x34), opts: []string{"Amatoriale", "Professionale"}, key: "work_mode", group: gOp, label: "Modalità di funzionamento",
			desc: "In modalità professionale la radio blocca alcuni menu e alcune funzioni dei tasti."},
		{off: at(0x01), opts: []string{"Canale", "Frequenza"}, key: "disp_mode", group: gDisp, label: "Canale mostrato come"},
		{off: at(0x26), opts: ijvNums(1, 5, 1, ""), key: "brightness", group: gDisp, label: "Luminosità del display"},
		{off: at(0x51), key: "show_clock", group: gDisp, label: "Orologio sul display"},
		{off: at(0x02), key: "auto_keylock", group: gOp, label: "Blocco automatico dei tasti"},
		{off: at(0x03), opts: []string{"Spento", "10 minuti", "30 minuti", "60 minuti", "120 minuti"}, key: "auto_off", group: gOp,
			label: "Spegnimento automatico"},
		{off: at(0x0B), opts: []string{"Spento", "1:1", "2:1"}, key: "power_save", group: gOp, label: "Risparmio della batteria"},
		{off: at(0x09), opts: lvl(5), key: "squelch_a", group: gOp, label: "Squelch della banda A"},
		{off: at(0x0A), opts: lvl(5), key: "squelch_b", group: gOp, label: "Squelch della banda B"},
		{off: at(0x0C), opts: lvl(3), key: "vox", group: gOp, label: "Livello del VOX"},
		{off: at(0x0F), opts: ijvNums(1, 5, 1, ""), key: "mic_gain", group: gOp, label: "Guadagno del microfono"},
		{off: at(0x57), key: "enh_audio", group: gOp, label: "Audio migliorato"},
		{off: at(0x28), key: "gps", group: gOp, label: "GPS"},
		{off: at(keyTone), key: "key_tone", group: gTone, label: "Toni dei tasti"},
		{off: at(idleTone), key: "idle_tone", group: gTone, label: "Tono di canale libero"},
		{off: at(0x31), mask: 0x01, key: "permit_dmr", group: gTone, label: "Tono di permesso (digitale)"},
		{off: at(0x31), mask: 0x02, key: "permit_fm", group: gTone, label: "Tono di permesso (analogico)"},
		{off: at(0x32), key: "reset_tone", group: gTone, label: "Tono di ripristino"},
		{off: at(0x39), key: "startup_tone", group: gTone, label: "Tono di accensione"},
		{off: at(0x3A), key: "call_end", group: gTone, label: "Tono di fine chiamata"},
		{off: at(0x29), key: "sms_alert", group: gTone, label: "Avviso dei messaggi"},
		{off: at(0x2F), key: "call_alert", group: gTone, label: "Avviso di chiamata"},
	})
}

var atSetTables = [2]*setTable{atSettings(false), atSettings(true)}

// atSetSpans: le tre parti della memoria che le impostazioni possono cambiare
// (indirizzo nella radio, posizione nell'immagine, lunghezza). In scrittura si
// rileggono dalla radio e si scrivono solo le righe da 16 byte diverse: chi non
// ha cambiato le impostazioni non le riscrive.
var atSetSpans = []atSpan{
	{atAddrSet, atOffSettings, atSetLen},
	{atAddrBoot, atOffBoot, atBootLen},
	{atAddrID, atOffRadioID, atRadioLen},
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool {
			m := genByID[kind]
			if m == nil {
				return false
			}
			_, ok := m.Fam.(atFam)
			return ok
		},
		get: func(kind string, img []byte) *setTable {
			if len(img) != atSize {
				return nil
			}
			if ato(genByID[kind]).x6 {
				return atSetTables[1]
			}
			return atSetTables[0]
		},
	})
}

// ---------------------------------------------------------------- nicFW880 / nicFW950

// campo di 16 bit (dal byte basso) della struttura settingsInfo
func nic16Tbl(off, n int) *setEnc {
	t := make([]*string, n)
	for i := range t {
		s := hex.EncodeToString([]byte{byte(i), byte(i >> 8)})
		t[i] = &s
	}
	return &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(off), "ffff"}}, Table: t}
}

func nic16List(key, group, label, desc string, off int, opts []string) setDef {
	codes := make([]string, len(opts))
	for i := range opts {
		codes[i] = strconv.Itoa(i)
	}
	return setDef{Key: key, Group: group, Label: label, Type: "list", Options: codes, ItOpts: opts, ItLabel: label, ItDesc: desc,
		Enc: nic16Tbl(nicfw880SetAt+off, len(opts))}
}

func nic16Bool(key, group, label, desc string, off int) setDef {
	return setDef{Key: key, Group: group, Label: label, Type: "bool", ItLabel: label, ItDesc: desc, Enc: nic16Tbl(nicfw880SetAt+off, 2)}
}

func nic16Int(key, group, label, desc string, off, min, max int) setDef {
	return setDef{Key: key, Group: group, Label: label, Type: "int", Min: float64(min), Max: float64(max), Step: 1, ItLabel: label, ItDesc: desc,
		Enc: &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(nicfw880SetAt + off), "ffff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: 1}}
}

var nicfw880SetTable = func() *setTable {
	gRx, gTx, gDisp := "Ricezione e audio", "Trasmissione", "Display"
	t := &setTable{Class: "nicFW880", Size: nicfw880Size,
		Source: "documentazione pubblicata dall'autore del firmware nicFW880 / nicFW950 (nicsure/nicFW880_950_Docs)"}
	t.Settings = []setDef{
		nic16List("squelch", gRx, "Squelch", "", 4, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)),
		nic16Int("noise_trigger", gRx, "Squelch: soglia del rumore", "Livello di rumore sotto il quale lo squelch si apre (0-127, predefinito 48).", 6, 0, 127),
		nic16Int("noise_ceiling", gRx, "Squelch: tetto del rumore", "Con un rumore uguale o superiore lo squelch si chiude subito (0-127, predefinito 60).", 8, 0, 127),
		nic16Int("noise_hyst", gRx, "Squelch: isteresi del rumore", "Distanza fra apertura e chiusura dello squelch (0-20, predefinito 4).", 10, 0, 20),
		nic16List("sq_throttle", gRx, "Squelch: rallentamento", "", 12, ijvNums(0, 10, 1, "")),
		nic16List("ste", gRx, "Eliminazione della coda dello squelch (STE)", "", 40, []string{"Spento", "Ricezione", "Trasmissione", "Ricezione e trasmissione"}),
		nic16Bool("noise_gate", gRx, "Filtro del rumore (noise gate)", "", 270),
		nic16Int("tone_time", gRx, "Durata del controllo del tono ricevuto (s)", "", 232, 2, 99),
		nic16List("mic_gain", gTx, "Guadagno del microfono", "", 18, ijvNums(0, 32, 1, "")),
		nic16List("vox", gTx, "VOX (trasmissione a voce)", "", 36, append([]string{"OFF"}, ijvNums(1, 15, 1, "")...)),
		nic16Int("lcd_timeout", gDisp, "Durata del display acceso", "Secondi dopo i quali lo schermo si spegne (0 = sempre acceso).", 26, 0, 999),
		nic16Int("heartbeat", gDisp, "Lampeggio di radio accesa (s)", "Ogni quanti secondi il LED lampeggia a schermo spento (0 = mai).", 28, 0, 500),
		nic16Bool("lcd_inverted", gDisp, "Colori dello schermo invertiti", "", 30),
		nic16List("lcd_gamma", gDisp, "Gamma dello schermo", "", 32, ijvNums(0, 4, 1, "")),
		nic16Int("lcd_day", gDisp, "Luminosità di giorno", "", 258, 1, 31),
		nic16Int("lcd_night", gDisp, "Luminosità di notte", "", 260, 1, 31),
		nic16List("lcd_mode", gDisp, "Luminosità in uso", "", 262, []string{"Giorno", "Notte"}),
	}
	return t
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool {
			m := genByID[kind]
			if m == nil {
				return false
			}
			_, ok := m.Fam.(nicfw880Fam)
			return ok
		},
		get: func(kind string, img []byte) *setTable {
			if len(img) != nicfw880Size || binary.LittleEndian.Uint16(img[nicfw880SetAt:]) != nicfw880Magic {
				return nil
			}
			return nicfw880SetTable
		},
	})
}

// ---------------------------------------------------------------- TD-UV68

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "td-uv68" && chirpSettings()["TDH8"] != nil },
		get: func(kind string, img []byte) *setTable {
			return fitSetTable(genByID[kind], chirpSettings()["TDH8"], img)
		},
	})
}
