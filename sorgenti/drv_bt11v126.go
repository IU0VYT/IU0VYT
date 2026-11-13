// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.25: due radio ricavate dal programma «BT-11_CPS» v1.2.6 di Bajeton (marzo
// 2026) e dal suo gemello «GS-10B Pro_EnCPS» v1.0.3 di Radioddity (stesso
// codice, classe KDH.RWDataOperation, con poche differenze). Nessuna delle due è
// nel CHIRP: protocollo, memoria e impostazioni vengono dalla lettura del codice
// dei programmi del costruttore.
//
// Protocollo: quello della famiglia Radtel RT-900 (drv_rt900.go), 57600 baud:
//   "PROGRAMBT11U" → 0x06; 'F' → 16 byte (il programma non li controlla);
//   pacchetto "SEND" (scelta della chiave di cifratura) → 0x06;
//   'R'/'W' indirizzo(2, BE) 0x40, blocchi di 64 byte cifrati; 'E' alla fine.
// Record del canale (Set/GetCHInfo_Parameter): identico a quello della RT-900
// (rt900Format): frequenze BCD, toni, byte 12 gruppo DTMF, 13 PTT-ID,
// 14 scrambler/potenza, 15 banda stretta, cifratura, blocco su occupato,
// scansione, AM, apprendimento; codice FHSS; nome di 12 byte.
//
// Bajeton BJ-8300 con il firmware nuovo (BT-11_CPS v1.2.6). Rispetto alla
// v1.0.9 (gestione originale di questo programma, file .dat) cambiano:
//   canali: 990 in 10 zone da 99 (prima 960 in 15 zone da 64), 0x0000-0x7BBF;
//   nomi delle zone: 0xC800-0xC8FF, 16 byte per zona, 12 byte GB2312, resto 0xFF;
//   impostazioni: tasti laterali in 0x902B-0x902D (prima 0x902A-0x902C) e
//   riduzione del rumore in 0x9032 (prima 0x9031).
// Aree lette e scritte dal programma, in quest'ordine: 0x0000-0x7BBF,
// 0x8000-0x803F (VFO), 0x9000-0x903F (impostazioni), 0xA000-0xA13F (DTMF),
// 0xB000-0xB0BF e 0xC000-0xC2FF (canali FM/AM/HF e loro nomi), 0xC800-0xC8FF.
// Qui i canali 961-990 e i nomi delle zone si scrivono solo se cambiano e dopo
// aver controllato che nella radio quelle aree contengano canali e nomi (o
// siano vuote). Se dopo il canale 960 la radio ha dati che non sono canali
// (firmware precedente) non si scrive nulla e la lettura lo segnala.
//
// Radioddity GS-10B Pro (GS-10B Pro_EnCPS v1.0.3): 999 canali in un unico
// elenco (0x0000-0x7CDF; l'ultimo blocco arriva a 0x7CFF), niente zone e niente
// area dei nomi delle zone, tasto laterale 3 anche con la pressione lunga
// (0x902E); le altre aree sono le stesse.
//
// Verifica: ogni posizione è stata letta due volte nel codice dei programmi,
// nel metodo che legge dalla radio (Get…) e in quello che scrive (Set…).

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

func init() {
	areas := [][2]int{{0x0000, 0x7D00}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}, {0xB000, 0xB0C0}, {0xC000, 0xC300}}
	o := &rt900Opt{
		magic:   []byte("PROGRAMBT11U"),
		crypt:   append([]byte("SEND \x01\x01"), make([]byte, 18)...),
		reads:   areas,
		ranges:  areas,
		memsize: 0xC300,
		airband: band{108000000, 136975000},
	}
	genRegister(&genModel{ID: "rd-gs10bpro", Vendor: "Radioddity", Model: "GS-10B Pro", RClass: "RT900BT", NoChirp: true,
		Name: "Radioddity GS-10B Pro", Slots: 999, First: 1, NameMax: 12,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 5 W", 5, 1}, {"Bassa 2 W", 2, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato, su un canale libero. 999 canali in un unico elenco. " +
			"Canali e impostazioni generali come nel programma del costruttore «GS-10B Pro_EnCPS» v1.0.3; VFO, DTMF e canali delle bande in sola ricezione " +
			"restano quelli della radio. Non compatibile con CHIRP."})

	// Bajeton BJ-8300 con il firmware nuovo (10 zone da 99 canali)
	o83 := &rt900Opt{
		magic:   []byte("PROGRAMBT11U"),
		crypt:   o.crypt,
		reads:   [][2]int{{0x0000, 0x7BC0}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}, {0xB000, 0xB0C0}, {0xC000, 0xC300}},
		ranges:  [][2]int{{0x0000, 0x7800}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}, {0xB000, 0xB0C0}, {0xC000, 0xC300}},
		memsize: 0xC300,
		airband: o.airband,
		// canali 961-990: riletti dalla radio, scritti solo se cambiano (fino a
		// 0x7BBF, come il programma: il blocco 0x7BC0 non si legge e non si scrive)
		hi: [2]int{0x7800, 0x7BC0}, hiStrict: true,
		hiJunk: "nella radio, dopo il canale 960, ci sono dati che non sono canali: la radio sembra avere il firmware precedente (960 canali in 15 zone da 64). " +
			"Nessun dato è stato scritto. Usa la scheda «Bajeton 8300» (file .dat), oppure aggiorna il firmware della radio",
	}
	genRegister(&genModel{ID: "bj-8300-990", Vendor: "Bajeton", Model: "BJ-8300", RClass: "RT920", NoChirp: true,
		Name: "Bajeton BJ-8300 (firmware nuovo: 10 zone da 99 canali)", Slots: 990, First: 1, NameMax: 12, Zones: 10,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta", 8, 0}, {"Media", 4, 1}, {"Bassa", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o83,
		Notes: "Accendi la radio con il cavo collegato, su un canale libero. Per la BJ-8300 con il firmware che divide i canali in 10 zone da 99 " +
			"(programma del costruttore «BT-11_CPS» v1.2.6); se la tua radio ha 15 zone da 64 canali usa la scheda «Bajeton 8300». " +
			"Canali, nomi delle zone e impostazioni generali come nel programma del costruttore; VFO, DTMF e canali delle bande in sola ricezione " +
			"restano quelli della radio. I canali 961-990 e i nomi delle zone si scrivono solo se li modifichi, dopo aver controllato che la radio li abbia. " +
			"Non compatibile con CHIRP."})
	genByID["bj-8300-990"].ZoneNames = &zoneNameSpec{Addr: 0xC800, Count: 10, Stride: 16, Max: 12, GB: true, Pad: 0xFF, Area: 0x100}
}

// Impostazioni della GS-10B Pro: blocco 0x9000-0x903F, come lo leggono e lo
// scrivono Get/SetFunConfigInfo_Part1 (0x9000) e _Part2 (0x9020) del programma;
// scelte della finestra «Optional Features» (FormFunConfig). Il programma legge
// ogni valore con la maschera 0x0F; «Cur Work Mode» e «Channel Divide» non si
// possono cambiare (il programma scrive sempre 0 e 1) e qui non ci sono.
func gs10bpSetDefs() []bj9kSetDef {
	side := []string{"Radio FM", "Potenza di trasmissione", "Scansione", "Ricerca frequenza", "Meteo (NOAA)", "Allarme (SOS)",
		"Analizzatore di spettro", "Monitor (apre lo squelch)", "PTT della banda B"}
	d := []bj9kSetDef{
		{field: "SQL", off: 0x00, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...)},
		{field: "TOT", off: 0x05, opts: []string{"OFF", "15 s", "30 s", "45 s", "60 s", "75 s"}},
		{field: "VOX", off: 0x02, opts: append([]string{"OFF"}, strNums(9, 1, 1, "")...)},
		{field: "VoxDelay", off: 0x20, opts: func() []string {
			o := make([]string, 16)
			for i := range o {
				o[i] = strconv.FormatFloat(0.5+float64(i)*0.1, 'f', 1, 64) + " s"
			}
			return o
		}()},
		{field: "BusyLockout", off: 0x0F},
		{field: "TailNoiseClear", off: 0x14},
		{field: "PassRepetNoiseClear", off: 0x15, opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...)},
		{field: "PassRepetNoiseDetect", off: 0x16, opts: append([]string{"OFF"}, strNums(10, 100, 100, " ms")...)},
		{field: "RTone", off: 0x1E, opts: []string{"1000 Hz", "1450 Hz", "1750 Hz", "2100 Hz"}},
		{field: "SoundTxEnd", off: 0x17, opts: []string{"OFF", "Beep", "Tono 1200 Hz"}},
		{field: "DenoiseSwitch", off: 0x32},

		{field: "TDR", off: 0x04},
		{field: "TDRTx", off: 0x13, opts: []string{"OFF (ultima attiva)", "Banda A", "Banda B"}},
		{field: "WorkModeA", off: 0x1A, mask: 0x0F, opts: []string{"Frequenza (VFO)", "Canali"}},
		{field: "WorkModeB", off: 0x1A, mask: 0xF0, opts: []string{"Frequenza (VFO)", "Canali"}},
		{field: "Scan", off: 0x0A, opts: []string{"TO (a tempo)", "CO (a portante)", "SE (si ferma)"}},
		{field: "FMRadio", off: 0x19, inv: true},
		{field: "WeatherCH", off: 0x27, opts: strNums(10, 1, 1, "")},

		{field: "DisplayModeA", off: 0x0D, opts: []string{"Nome", "Frequenza", "Numero del canale"}},
		{field: "DisplayModeB", off: 0x0E, opts: []string{"Nome", "Frequenza", "Numero del canale"}},
		{field: "AutoBacklight", off: 0x03, opts: []string{"Sempre accesa", "5 s", "10 s", "15 s", "20 s", "30 s", "1 min", "2 min", "3 min"}},
		{field: "PowerMsg", off: 0x1C, opts: []string{"Logo predefinito", "Tensione della batteria"}},
		{field: "Language", off: 0x08, opts: []string{"Inglese", "Cinese"}},
		{field: "VoicePrompt", off: 0x07},
		{field: "BeepPrompt", off: 0x06},
		// «Menu Exit Time»: 5-50 secondi a passi di 5, poi 60
		{field: "TimerMenuQuit", off: 0x21, opts: append(strNums(10, 5, 5, " s"), "60 s")},
		{field: "SaveMode", off: 0x01, opts: []string{"OFF", "Normale", "Super", "Profondo"}},
		{field: "LockKeyBoard", off: 0x1B},
		{field: "AutoKeyLock", off: 0x10, opts: []string{"OFF", "5 s", "10 s", "15 s"}},
		{field: "KeySide2", off: 0x2B, opts: side[:8]},
		{field: "KeySide2L", off: 0x2C, opts: side[:9]},
		{field: "KeySide3", off: 0x2D, opts: side[:8]},
		{field: "KeySide3L", off: 0x2E, opts: side[:8], key: "pf3_long", group: "Display, suoni e tasti", label: "Tasto laterale 3 (pressione lunga)",
			desc: "Funzione del tasto laterale 3 tenuto premuto."},

		{field: "DTMF", off: 0x09, opts: []string{"OFF", "Tasti", "Codice ANI", "Tasti + ANI"}},
		{field: "PTTID", off: 0x0B, opts: []string{"OFF", "All'inizio", "Alla fine", "Inizio e fine"}},
		{field: "SendIDDelay", off: 0x0C, opts: []string{"0 ms", "100 ms", "200 ms", "400 ms", "600 ms", "800 ms", "1000 ms"}},
		{field: "AlarmMode", off: 0x11, opts: []string{"Solo sulla radio", "Trasmette il suono", "Trasmette il codice"}},
		{field: "AlarmSound", off: 0x12},
		{field: "BTRWData", off: 0x1D},
	}
	// nomi, descrizioni, gruppi e avvisi: quelli delle stesse voci della BJ-8300
	common := map[string]bajSetDef{}
	for _, b := range bajSetDefs {
		common[b.field[4:]] = b
	}
	for i := range d {
		c := common[d[i].field]
		if d[i].key == "" {
			d[i].key, d[i].group, d[i].label, d[i].desc, d[i].warn = c.key, c.group, c.label, c.desc, c.warn
		}
		if d[i].mask == 0 {
			d[i].mask = 0x0F
		}
	}
	return d
}

var gs10bpSettings = kdhSetTable("GS-10B Pro", 0x9040, gs10bpSetDefs(),
	fmt.Sprintf("programma del costruttore %s (finestra «Optional Features»)", "GS-10B Pro_EnCPS v1.0.3"))

// Impostazioni della BJ-8300 con il firmware nuovo: le stesse voci della
// gestione originale (bajSetDefs), nelle posizioni di Get/SetFunConfigInfo_Part1
// e _Part2 del programma BT-11_CPS v1.2.6. «Cur Work Mode» non c'è: il
// programma v1.2.6 scrive sempre 0.
func bj83nSetDefs() []bj9kSetDef {
	off := map[string]int{"WorkModeA": 0x1A, "WorkModeB": 0x1A, "VoxDelay": 0x20, "TimerMenuQuit": 0x21, "RemoteKillEnable": 0x24,
		"WeatherCH": 0x27, "DivideCH": 0x28, "KeySide2": 0x2B, "KeySide2L": 0x2C, "KeySide3": 0x2D, "DenoiseSwitch": 0x32}
	for i, f := range btFun1 {
		if f != "" {
			off[f[4:]] = i
		}
	}
	var d []bj9kSetDef
	for _, b := range bajSetDefs {
		f := b.field[4:]
		o, ok := off[f]
		if !ok || f == "CurWorkMode" {
			continue
		}
		x := bj9kSetDef{field: f, off: o, mask: 0x0F, opts: b.opts, inv: b.inv, key: b.key, group: b.group, label: b.label, desc: b.desc, warn: b.warn}
		switch f {
		case "WorkModeB":
			x.mask = 0xF0
		case "DivideCH":
			x.opts = []string{fmt.Sprintf("Zone (%d × %d)", 10, 99), fmt.Sprintf("Tutti i canali (%d)", 990)}
			x.desc = kdhDivideDesc
		}
		d = append(d, x)
	}
	return d
}

// kdhDivideDesc (4.25): descrizione di «Channel Divide» senza il numero delle zone
const kdhDivideDesc = "Canali divisi in zone oppure tutti in un unico elenco."

var bj83nSettings = kdhSetTable("BJ-8300 v1.2.6", 0x9040, bj83nSetDefs(),
	fmt.Sprintf("programma del costruttore %s (finestra «Optional Features»)", "BT-11_CPS v1.2.6"))

// kdhSetTable: tabella delle impostazioni di un blocco 0x9000 dei programmi
// «KDH» (una scelta per byte o per gruppo di bit)
func kdhSetTable(class string, size int, defs []bj9kSetDef, src string) *setTable {
	t := &setTable{Class: class, Size: size, Source: src}
	for _, b := range defs {
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
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "rd-gs10bpro" || kind == "bj-8300-990" },
		get: func(kind string, img []byte) *setTable {
			if len(img) < gs10bpSettings.Size {
				return nil
			}
			if kind == "bj-8300-990" {
				return bj83nSettings
			}
			return gs10bpSettings
		},
	})
}
