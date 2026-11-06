// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radtel_rt900.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Bajeton BJ7800 (driver radtel_rt900.py di CHIRP, classe BJ7800, derivata
// dalla Radtel RT-920; stesso apparato della SHU H28Y Pro) e Bajeton BJ-7700
// (Radtel RT-910: il programma del costruttore «BJ-7700_EnCPS» è il
// «BT-RT910_CPS», per «RT910» e «BJ-7700»; classi RT910BT e RT910 di CHIRP).
// 960 canali in 15 zone da 64, record di 32 byte, 57600 baud.
//
// Protocollo (mml_jc8810 + radtel_rt900 di CHIRP):
//   "PROGRAMBT11U" (BJ-7800) o "PROGRAMBT80U" (BJ-7700) → 0x06 (fino a 5 tentativi);
//   'F' → identificativo di 8 byte, seguito da un secondo identificativo di 8 byte;
//   pacchetto "SEND …" (impostazione della cifratura) → 0x06;
//   lettura: 'R' indirizzo(2, BE) 0x40 → 'R' indirizzo 0x40 + 64 byte;
//   scrittura: 'W' indirizzo 0x40 + 64 byte → 0x06;
//   uscita: 'E'.
// Dati cifrati con btCrypt (chiave "CO 7") sotto l'indirizzo 0xF000 (sopra c'è
// la taratura, letta in chiaro e mai riscritta).
// L'immagine copre 0x0000-0xF27F: CHIRP legge a blocchi di 64 byte fino a
// 0xF250 e l'ultimo blocco va oltre.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const rt900Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:4,
     scode:4;
  u8 unknown2:6,
     pttid:2;
  u8 scramble:6,
     txpower:2;
  u8 unknown4:1,
     narrow:1,
     encrypt:2,
     bcl:1,
     scan:1,
     am_modulation:1,
     learning:1;
  ul24 code;
  u8 code_flag;
  char name[12];
} memory[%d];
`

// BJ-9000 (Radtel RT-950 Pro, programma «BT-RT950PRO_CPS»): stesso record di
// 32 byte, ma nel byte 15 l'apprendimento è il bit 7, la modulazione AM il bit 0
// e il bit 1 vale sempre 1.
const rt950Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:4,
     scode:4;
  u8 unknown2:6,
     pttid:2;
  u8 scramble:6,
     txpower:2;
  u8 learning:1,
     narrow:1,
     encrypt:2,
     bcl:1,
     scan:1,
     txen:1,
     am_modulation:1;
  ul24 code;
  u8 code_flag;
  char name[12];
} memory[%d];
`

// 4.7: il byte 15 si scrive come i programmi del costruttore fino alla v1.2.x
// (bit 1 sempre a 1) con tutti i firmware. La v1.3.3 legge la modulazione come
// (byte & 3) % 2, cioè solo il bit 0, e il bit 1 lo ignora (lo riscrive a 0); i
// programmi precedenti, anche la v1.2.1 per il firmware 0.24 e successivi, lo
// scrivono sempre a 1. Con il bit a 1 il canale va bene per ogni versione del
// firmware; senza (come scriveva la 4.6 nella scheda da 990 canali) una radio con
// il firmware fino a 0.23 potrebbe non trasmettere su quel canale.

type rt900Opt struct {
	magic   []byte
	fp1     [][]byte // risposta al comando 'F' (nil: non controllata)
	fp2     [][]byte // secondo identificativo
	crypt   []byte   // pacchetto "SEND …"
	memsize int      // area letta (CHIRP _memsize)
	ranges  [][2]int // aree scritte (CHIRP _ranges, nell'ordine)
	airband band     // gamma aeronautica: canali in AM
	format  string   // record dei canali (predefinito rt900Format)
	model   string   // risposta attesa al comando 'M' (12 byte, BJ-9000)
	block   int      // byte per blocco (predefinito 0x40)
	reads   [][2]int // aree lette (nil: da 0 a memsize)
	txen    bool     // bit "txen" sempre a 1 (BJ-9000)
	// plain: protocollo di mml_jc8810.py (4.1, Radtel RT-470 e simili): dopo
	// l'identificativo non c'è il secondo identificativo né il pacchetto "SEND",
	// e i dati non sono cifrati
	plain bool
	// 4.6 (BJ-9000 / Radtel RT-950 Pro): area dei canali scritta solo quando
	// cambia, dopo averla riletta dalla radio nella stessa sessione (canali
	// 961-990 del firmware 0.24 e successivi, 0x7800-0x7BFF)
	hi [2]int
	// layout (4.6): canali della generazione di firmware (960 o 990); le due
	// versioni della stessa radio sono modelli distinti (vedi rt950PickLayout)
	layout int
}

func (o *rt900Opt) blk() int {
	if o.block == 0 {
		return 0x40
	}
	return o.block
}

type rt900Fam struct{}

func init() {
	o := &rt900Opt{
		magic: []byte("PROGRAMBT11U"),
		fp1:   [][]byte{{0x01, 0x36, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20}, {0x01, 0x00, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20}},
		fp2:   [][]byte{{0x02, 0x00, 0x02, 0x60, 0x01, 0x03, 0x30, 0x04}},
		crypt: append([]byte("SEND \x01\x01"), make([]byte, 18)...),
		// 0x0000-0x77FF: 15 zone da 64 canali di 32 byte
		ranges: [][2]int{{0x0000, 0x7800}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xB000, 0xB400},
			{0xC000, 0xC400}, {0xA000, 0xA140}, {0xD000, 0xD040}},
		memsize: 0xF250,
		airband: band{108000000, 136975000},
	}
	genRegister(&genModel{ID: "bj-7800", Vendor: "Bajeton", Model: "BJ7800", RClass: "BJ7800",
		Name: "Bajeton BJ-7800", Slots: 960, First: 1, NameMax: 12, Zones: 15,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato, su un canale libero. " +
			"Canali 1-960 = 15 zone da 64 (zona 1: canali 1-64, zona 2: 65-128, …). " +
			"Driver di CHIRP per il firmware V0.15P."})

	// BJ-7700 = Radtel RT-910 (stesso protocollo della RT-900 BT, memoria della RT-910)
	o910 := &rt900Opt{
		magic:   []byte("PROGRAMBT80U"),
		fp1:     o.fp1,
		fp2:     o.fp2,
		crypt:   o.crypt,
		ranges:  [][2]int{{0x0000, 0x7800}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}, {0xD000, 0xD040}},
		memsize: 0xF250,
		airband: o.airband,
	}
	for _, v := range []struct{ id, model, rclass, name, fw string }{
		{"bj-7700", "RT-910_BT", "RT910BT", "Bajeton BJ-7700 (Bluetooth)", "V0.24P della RT-910 BT"},
		{"bj-7700nbt", "RT-910", "RT910", "Bajeton BJ-7700 (senza Bluetooth)", "V0.11 della RT-910"},
	} {
		genRegister(&genModel{ID: v.id, Vendor: "Radtel", Model: v.model, RClass: v.rclass,
			Name: v.name, Slots: 960, First: 1, NameMax: 12, Zones: 15,
			Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
			Powers:  []genPower{{"Alta", 8, 0}, {"Media", 4, 1}, {"Bassa", 1, 2}},
			Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
			DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o910,
			Notes: "Accendi la radio con il cavo collegato, su un canale libero. " +
				"Canali 1-960 = 15 zone da 64 (zona 1: canali 1-64, zona 2: 65-128, …). " +
				"La BJ-7700 è la Radtel RT-910 (il programma del costruttore è lo stesso): driver di CHIRP per il firmware " + v.fw + "; " +
				"nei file .img compare come Radtel " + v.model + "."})
	}

	// 4.0: Radtel RT-900, RT-900 BT, RT-920 e Radioddity GS-10B (radtel_rt900.py
	// di CHIRP): stesso protocollo, cambiano il numero di canali e le aree scritte
	// aree scritte nell'ordine di CHIRP (_ranges)
	mkO := func(magic string, chans int, extra ...[2]int) *rt900Opt {
		r := [][2]int{{0x0000, chans}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}}
		r = append(append(r, extra...), [2]int{0xD000, 0xD040})
		return &rt900Opt{magic: []byte(magic), fp1: o.fp1, fp2: o.fp2, crypt: o.crypt, ranges: r, memsize: 0xF250, airband: o.airband}
	}
	for _, v := range []struct {
		id, vendor, model, rclass, name string
		slots                           int
		o                               *rt900Opt
	}{
		{"rt-rt900", "Radtel", "RT-900", "RT900", "Radtel RT-900", 999, mkO("PROGRAMBT80U", 0x7CE0)},
		{"rt-rt900bt", "Radtel", "RT-900_BT", "RT900BT", "Radtel RT-900 BT", 999, mkO("PROGRAMBT80U", 0x7CE0)},
		{"rd-gs10b", "Radioddity", "GS-10B", "RadioddityGS10B", "Radioddity GS-10B", 256, mkO("PROGRAMBT80U", 0x2000)},
		{"rt-rt920", "Radtel", "RT-920", "RT920", "Radtel RT-920", 990, mkO("PROGRAMBT11U", 0x7BC0, [2]int{0xB000, 0xB400}, [2]int{0xC000, 0xC400}, [2]int{0xC800, 0xC8A0})},
	} {
		zones := 0
		if v.id == "rt-rt920" {
			// 4.6: firmware V0.17 e successivi (quello del driver di CHIRP, V0.18):
			// 10 zone da 99 canali, come le mostra la radio
			zones = 10
		}
		genRegister(&genModel{ID: v.id, Vendor: v.vendor, Model: v.model, RClass: v.rclass,
			Name: v.name, Slots: v.slots, First: 1, NameMax: 12, Zones: zones,
			Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
			Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
			Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
			DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: v.o,
			Notes: "Accendi la radio con il cavo collegato, su un canale libero. Driver di CHIRP della famiglia Radtel RT-900."})
	}
}

func init() {
	// BJ-9000 = Radtel RT-950 Pro (programmi del costruttore «BT-RT950PRO_CPS»,
	// per la Bajeton «BJ-9000_CPS»): 115200 baud, "PROGRAMBT9000U", 'F' (16 byte),
	// 'M' → "RT-950", blocchi da 128 byte.
	//
	// 4.6: due generazioni di firmware con memorie dei canali diverse. Il firmware
	// Radtel V0.24 (dicembre 2025) è passato da 15 zone da 64 canali a 10 zone da 99
	// e il programma del costruttore v1.2.1 da Channel[960] (0x0000-0x77FF) a
	// Channel[990] (0x0000-0x7BBF); stessi record di 32 byte, stesse impostazioni
	// (verificato confrontando il codice dei programmi v1.1.1, v1.2.0, v1.2.1 e
	// v1.3.3). In entrambe la scelta «Channel Divide» (byte 0x9026) mostra i canali
	// in zone o in un unico elenco (960 o 990). La radio non comunica la versione del
	// firmware (risposte a 'F' e 'M' identiche con il firmware 0.18 e 0.29): la
	// versione si riconosce dai dati (rt950Layout), altrimenti si chiede.
	// Si legge sempre fino a 0x7BFF, come il programma v1.2.1 (lettura innocua anche
	// con il firmware precedente); si scrivono i canali (0x0000-0x77FF), i canali
	// 961-990 solo se cambiano e dopo averli riletti dalla radio, e le impostazioni
	// generali (blocco 0x9000, rileggendo dalla radio i byte non modificati qui).
	reads := [][2]int{{0x0000, 0x7C00}, {0x8000, 0x8080}, {0x9000, 0x9080}, {0xA000, 0xA180}, {0xB000, 0xB100}, {0xD000, 0xD300}}
	mk := func(layout int) *rt900Opt {
		o := &rt900Opt{
			magic:   []byte("PROGRAMBT9000U"),
			crypt:   append([]byte("SEND \x01\x01"), make([]byte, 18)...),
			model:   "RT-950",
			block:   0x80,
			reads:   reads,
			ranges:  [][2]int{{0x0000, 0x7800}, {0x9000, 0x9080}},
			memsize: 0xD300,
			airband: band{108000000, 136975000},
			format:  rt950Format,
			txen:    true,
			layout:  layout,
		}
		if layout == 990 {
			o.hi = [2]int{0x7800, 0x7C00}
		}
		return o
	}
	common := "Accendi la radio con il cavo collegato. Canali e impostazioni generali (finestra «Optional Features») come nel programma del costruttore; " +
		"VFO, DTMF e APRS restano quelli della radio. Non compatibile con CHIRP. "
	genRegister(&genModel{ID: "bj-9000", Vendor: "Bajeton", Model: "BJ-9000", RClass: "RT950Pro", NoChirp: true,
		Name: "Bajeton BJ-9000 (firmware 0.24 e successivi: 990 canali)", Slots: 990, First: 1, NameMax: 12, Zones: 10,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 115200, Fam: rt900Fam{}, Opt: mk(990),
		Notes: common + "Firmware 0.24 e successivi (programma del costruttore v1.2.1 e successivi): canali 1-990 = 10 zone da 99, " +
			"oppure un unico elenco di 990 canali (impostazione «Divisione dei canali»). I canali 961-990 si scrivono solo se cambiano, " +
			"dopo aver controllato che la radio li abbia."},
		&genModel{ID: "bj-9000-960", Vendor: "Bajeton", Model: "BJ-9000", RClass: "RT950Pro", NoChirp: true,
			Name: "Bajeton BJ-9000 (firmware fino a 0.23: 960 canali)", Slots: 960, First: 1, NameMax: 12, Zones: 15,
			Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
			Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
			Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
			DCS: uv5rDTCS, Baud: 115200, Fam: rt900Fam{}, Opt: mk(960),
			Notes: common + "Firmware fino a 0.23 (programma del costruttore fino a v1.2.0): canali 1-960 = 15 zone da 64, " +
				"oppure un unico elenco di 960 canali (impostazione «Divisione dei canali»)."})
}

func init() {
	// 4.6: Radtel RT-950 (senza «Pro»): programma del costruttore «BT-9000_CPS»
	// (RT-950_EnCPS v1.1.4-v1.3.0), stesso codice dei programmi della RT-950 Pro fino
	// alla v1.2.0 (canali, impostazioni, cifratura) ma la radio risponde a 'M' con
	// "BT-9000" e la memoria ha meno aree: canali 0x0000-0x77FF (960, 15 zone da 64,
	// anche con il firmware V0.29), VFO 0x8000, impostazioni 0x9000, DTMF 0xA000,
	// modulazione 0xB000-0xB07F. Si scrivono i canali e le impostazioni generali.
	o := &rt900Opt{
		magic:   []byte("PROGRAMBT9000U"),
		crypt:   append([]byte("SEND \x01\x01"), make([]byte, 18)...),
		model:   "BT-9000",
		block:   0x80,
		reads:   [][2]int{{0x0000, 0x7800}, {0x8000, 0x8080}, {0x9000, 0x9080}, {0xA000, 0xA180}, {0xB000, 0xB080}},
		ranges:  [][2]int{{0x0000, 0x7800}, {0x9000, 0x9080}},
		memsize: 0xB080,
		airband: band{108000000, 136975000},
		format:  rt950Format,
		txen:    true,
	}
	genRegister(&genModel{ID: "rtl-rt950", Vendor: "Radtel", Model: "RT-950", RClass: "RT950", NoChirp: true,
		Name: "Radtel RT-950", Slots: 960, First: 1, NameMax: 12, Zones: 15,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 115200, Fam: rt900Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato. Canali 1-960 = 15 zone da 64, oppure un unico elenco di 960 canali (impostazione " +
			"«Divisione dei canali»). Canali e impostazioni generali come nel programma del costruttore «BT-9000_CPS» v1.3.0; VFO, DTMF e APRS " +
			"restano quelli della radio. Non compatibile con CHIRP."})
}

// rt950Siblings: le due versioni (960 e 990 canali) della stessa radio
func rt950Siblings(m *genModel) (m960, m990 *genModel) {
	o, ok := m.Opt.(*rt900Opt)
	if !ok || o.layout == 0 {
		return nil, nil
	}
	return genByID["bj-9000-960"], genByID["bj-9000"]
}

// rt950RecPlausible: il record di 32 byte è un canale (frequenze BCD valide)?
func rt950RecPlausible(r []byte) bool {
	bcd := func(b []byte) (int64, bool) {
		var v int64
		for k := len(b) - 1; k >= 0; k-- {
			hi, lo := b[k]>>4, b[k]&0x0F
			if hi > 9 || lo > 9 {
				return 0, false
			}
			v = v*100 + int64(hi)*10 + int64(lo)
		}
		return v * 10, true
	}
	rx, ok := bcd(r[0:4])
	if !ok || rx < 18000000 || rx > 1000000000 {
		return false
	}
	if bytes.Equal(r[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return true
	}
	tx, ok := bcd(r[4:8])
	return ok && tx >= 18000000 && tx <= 1000000000
}

// rt950Layout (4.6): generazione del firmware riconosciuta dai dati letti dalla
// radio: 990 (0.24 e successivi), 960 (fino a 0.23) o 0 se i dati non bastano.
//   - dopo il canale 960 (0x7800-0x7BBF) il firmware 0.24 tiene i canali 961-990:
//     un record che è un canale dice 990; dati che non sono canali (il firmware
//     precedente usa quest'area per altro) dicono 960;
//   - zona attiva di una banda (0x902D-0x902F) oltre la decima: 960.
func rt950Layout(mem []byte) (int, string) {
	if len(mem) < 0x9080 {
		return 0, ""
	}
	ch, junk := rt950HiCount(mem[0x7800:0x7BC0])
	switch {
	case junk > 0 && ch == 0:
		return 960, "dopo il canale 960 la memoria contiene dati che non sono canali"
	case ch > 0 && junk == 0:
		if ch == 1 {
			return 990, "la radio ha un canale oltre il 960"
		}
		return 990, fmt.Sprintf("la radio ha %d canali oltre il 960", ch)
	}
	for k := 0; k < 3; k++ {
		if z := mem[0x902D+k]; z != 0xFF && z&0x0F >= 10 {
			return 960, fmt.Sprintf("la zona attiva della banda %c è la %d", 'A'+k, z&0x0F+1)
		}
	}
	return 0, ""
}

// rt950HiCount: record dell'area dei canali 961-990 che sono canali e che non lo sono
func rt950HiCount(area []byte) (ch, junk int) {
	for a := 0; a+32 <= len(area); a += 32 {
		r := area[a : a+32]
		// posizione vuota: solo 0xFF (o 0x00, come un record azzerato)
		clean := true
		for _, b := range r {
			clean = clean && (b == 0xFF || b == 0x00)
		}
		if clean {
			continue
		}
		if r[0] != 0xFF && rt950RecPlausible(r) {
			ch++
		} else {
			junk++
		}
	}
	return ch, junk
}

// rt950PickLayout (4.6): versione della radio (960 o 990 canali) per la memoria
// letta: quella riconosciuta dai dati, altrimenti quella proposta (m).
func rt950PickLayout(m *genModel, mem []byte) (*genModel, string) {
	m960, m990 := rt950Siblings(m)
	if m960 == nil {
		return m, ""
	}
	l, why := rt950Layout(mem)
	switch l {
	case 990:
		if m != m990 {
			return m990, "Firmware 0.24 o successivo (" + why + "): canali gestiti come 10 zone da 99 (990 canali)."
		}
		return m, ""
	case 960:
		if m != m960 {
			return m960, "Firmware fino a 0.23 (" + why + "): canali gestiti come 15 zone da 64 (960 canali)."
		}
		return m, ""
	}
	o := rt900o(m)
	return m, fmt.Sprintf("Dai dati della radio non si capisce la versione del firmware: canali gestiti come %s. "+
		"Controlla la versione sulla radio (menu: informazioni sulla radio): con il firmware fino a 0.23 i canali sono 960 (15 zone da 64), "+
		"dalla 0.24 sono 990 (10 zone da 99). Se non corrisponde, apri la scheda dell'altra versione dalla galleria e rileggi la radio.",
		map[int]string{960: "firmware fino a 0.23 (960 canali, 15 zone da 64)", 990: "firmware 0.24 e successivi (990 canali, 10 zone da 99)"}[o.layout])
}

// rt900Identify: ingresso in programmazione e uscita (identificazione della radio)
func rt900Identify(ctx context.Context, m *genModel, port string) error {
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return err
	}
	defer s.p.Close()
	err = s.enter()
	s.exit()
	if err != nil {
		return genFail("Identificazione", err)
	}
	return nil
}

func rt900o(m *genModel) *rt900Opt {
	if o, ok := m.Opt.(*jc8810Opt); ok {
		return &o.rt900Opt
	}
	return m.Opt.(*rt900Opt)
}

func (rt900Fam) mem(m *genModel) *bwEl {
	f := rt900o(m).format
	if f == "" {
		f = rt900Format
	}
	return bwMust(fmt.Sprintf(f, m.Slots))
}

func (f rt900Fam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f rt900Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

func (f rt900Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}

func (f rt900Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f rt900Fam) Match(m *genModel, img []byte) bool { return false }

// Finish (4.7): RT-950 Pro / BJ-9000 / RT-950: il bit 1 del byte 15 («trasmissione
// abilitata» per i programmi fino alla v1.2.x, ignorato dalla v1.3.3) si mette a 1
// in tutti i canali presenti, anche in quelli non modificati: ripara i canali
// scritti senza il bit (4.6 nella scheda da 990 canali, programma v1.3.x) che con
// il firmware fino a 0.23 potrebbero non trasmettere. Per il firmware recente non
// cambia nulla (la v1.3.3 legge solo il bit 0, la modulazione).
func (f rt900Fam) Finish(m *genModel, img []byte) {
	o := rt900o(m)
	if !o.txen || o.format != rt950Format {
		return
	}
	// solo i canali 1-960: dopo il 960 una radio con il firmware fino a 0.23 può
	// avere altri dati, che non si toccano (i canali 961-990 modificati hanno già
	// il bit, vedi Set)
	for i := 0; i < min(m.Slots, 960); i++ {
		r := f.rec(m, i)
		if len(img) < r.Off+r.Size || !rt950RecUsed(r.Raw(img)) {
			continue
		}
		r.F("txen").SetInt(img, 1)
	}
}

// rt950RecUsed: record di canale presente (frequenza di ricezione valida in BCD)
func rt950RecUsed(raw []byte) bool {
	if len(raw) < 4 {
		return false
	}
	empty := true
	for _, b := range raw[:4] {
		if b != 0xFF && b != 0x00 {
			empty = false
		}
		if b&0x0F > 9 || b>>4 > 9 {
			return false
		}
	}
	return !empty
}

func (f rt900Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if c.Rx == 0 {
		return nil
	}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	c.Scan = r.F("scan").Bool(img)
	pw := int(r.F("txpower").Int(img)) // 0 alta, 1 media, 2 bassa (3: non valido)
	if pw > 2 {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r.F("narrow").Bool(img)
	// come CHIRP: 0xFF → spazio, spazi finali tolti, poi via i NUL
	c.Name = strings.ReplaceAll(pyName(r.F("name").Raw(img), true), "\x00", "")
	return c
}

func (f rt900Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := rt900o(m)
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	// impostazioni del canale che CHIRP conserva (mem.extra)
	wasEmpty := r.Raw(img)[0] == 0xFF
	var encrypt, scramble, bcl, learning, pttid, scode, code int64
	oldAM, oldRx := false, int64(-1)
	code = 0xFFFFFF
	if !wasEmpty {
		encrypt, scramble = r.F("encrypt").Int(img), r.F("scramble").Int(img)
		bcl, learning = r.F("bcl").Int(img), r.F("learning").Int(img)
		pttid, scode = r.F("pttid").Int(img), r.F("scode").Int(img)
		if v := r.F("code").Int(img); v != 0xFFFFFF {
			code = v & 0x7FFFFF
		}
		oldAM = r.F("am_modulation").Bool(img)
		oldRx = r.F("rxfreq").Int(img) * 10
	}
	raw := append(bytes.Repeat([]byte{0}, 16), bytes.Repeat([]byte{0xFF}, 16)...)
	r.SetRaw(img, raw)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("name").SetStr(img, c.Name, 0xFF)
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("txpower").SetInt(img, int64(m.powerIndex(c.Power)))
	// AM: obbligatoria in banda aerea; altrimenti si conserva quella del
	// canale se la frequenza non cambia (es. canali CB in AM)
	am := (c.Rx >= o.airband.Lo && c.Rx <= o.airband.Hi) || (oldAM && oldRx == c.Rx)
	r.F("am_modulation").SetBool(img, am)
	if o.txen {
		r.F("txen").SetInt(img, 1)
	}
	if !wasEmpty {
		r.F("encrypt").SetInt(img, encrypt)
		r.F("scramble").SetInt(img, scramble)
		r.F("bcl").SetInt(img, bcl)
		r.F("learning").SetInt(img, learning)
		r.F("pttid").SetInt(img, pttid)
		r.F("scode").SetInt(img, scode)
		r.F("code").SetInt(img, code)
		if code == 0xFFFFFF {
			r.F("code_flag").SetInt(img, 0xFF)
		} else {
			r.F("code_flag").SetInt(img, 0xA0)
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type rt900Sess struct {
	ctx context.Context
	p   Port
	m   *genModel
}

const rt900Timeout = 1500 * time.Millisecond

func (s *rt900Sess) recv(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, rt900Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

func bytesIn(l [][]byte, b []byte) bool {
	for _, x := range l {
		if bytes.Equal(x, b) {
			return true
		}
	}
	return false
}

// ingresso in programmazione (_enter_programming_mode di radtel_rt900/mml_jc8810)
func (s *rt900Sess) enter() error {
	o := rt900o(s.m)
	ok, saw := false, false
	for k := 0; k < 5 && !ok; k++ {
		if err := s.p.Write(o.magic); err != nil {
			return err
		}
		b, err := s.recv(1)
		if err == errCancelled {
			return err
		}
		if err == nil {
			saw = true
			ok = b[0] == 0x06
		}
	}
	if !ok {
		if saw {
			return rErr("la radio non ha accettato la modalità di programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
		}
		return rErr("la radio non risponde")
	}
	if err := s.p.Write([]byte{'F'}); err != nil {
		return err
	}
	id, err := s.recv(8)
	if err != nil {
		return err
	}
	if o.fp1 != nil && !bytesIn(o.fp1, id) {
		return rErr("identificativo della radio sconosciuto (% X): è il modello giusto?", id)
	}
	if o.plain {
		return nil
	}
	id2, err := s.recv(8)
	if err != nil {
		return err
	}
	if o.model != "" {
		// modello (programma del costruttore: 'M' → 12 byte)
		if err := s.p.Write([]byte{'M'}); err != nil {
			return err
		}
		mb, err := s.recv(12)
		if err != nil {
			return err
		}
		if got := strings.TrimSpace(string(trimBytesBoth(mb, 0x00, 0xFF))); got != o.model {
			return rErr("la radio si è identificata come «%s», non come %s: è il modello giusto?", got, o.model)
		}
	}
	if err := s.p.Write(o.crypt); err != nil {
		return err
	}
	ack, err := s.recv(1)
	if err != nil {
		return err
	}
	if ack[0] != 0x06 {
		return rErr("errore nell'impostazione della cifratura")
	}
	if o.fp2 != nil && !bytesIn(o.fp2, id2) {
		return rErr("identificativo secondario della radio sconosciuto (% X)", id2)
	}
	return nil
}

func (s *rt900Sess) exit() { _ = s.p.Write([]byte{'E'}) }

func rt900Open(ctx context.Context, m *genModel, port string) (*rt900Sess, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &rt900Sess{ctx: ctx, p: p, m: m}, nil
}

func (f rt900Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rt900o(m)
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	err = s.enter()
	if err != nil {
		s.exit()
		return nil, "", genFail("Identificazione", err)
	}
	key := []byte(btSymbols[1])
	bs := o.blk()
	reads := o.reads
	if reads == nil {
		reads = [][2]int{{0, (o.memsize + bs - 1) / bs * bs}}
	}
	total := 0
	for _, r := range reads {
		total += r[1] - r[0]
	}
	data := bytes.Repeat([]byte{0xFF}, o.memsize)
	if o.reads == nil {
		data = nil
	}
	done := 0
	for _, r := range reads {
		for a := r[0]; a < r[1]; a += bs {
			if ctx.Err() != nil {
				s.exit()
				return nil, "", errCancelled
			}
			prog(done, total, "Lettura…")
			cmd := []byte{'R', byte(a >> 8), byte(a), byte(bs)}
			if err := s.p.Write(cmd); err != nil {
				return nil, "", err
			}
			d, err := s.recv(4 + bs)
			if err == nil && !bytes.Equal(d[:4], cmd) {
				err = rErr("risposta inattesa")
			}
			if err != nil {
				s.exit()
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			blk := append([]byte{}, d[4:]...)
			if a < 0xF000 && !o.plain {
				btCrypt(blk, key)
			}
			if o.reads == nil {
				data = append(data, blk...)
			} else {
				copy(data[a:], blk)
			}
			done += bs
		}
	}
	s.exit()
	prog(total, total, "Lettura completata")
	return data, "", nil
}

func (f rt900Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rt900o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.memsize)
	}
	total := 0
	for _, r := range o.ranges {
		total += r[1] - r[0]
	}
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		s.exit()
		return "", genFail("Identificazione", err)
	}
	key := []byte(btSymbols[1])
	bs := o.blk()
	ranges := o.ranges
	if o.hi[1] > 0 {
		// 4.6: canali 961-990 (firmware 0.24 e successivi): si rileggono dalla radio e
		// si scrivono solo se cambiano; se la radio in quell'area ha dati che non
		// sono canali (firmware fino a 0.23) non si scrive nulla
		cur := make([]byte, 0, o.hi[1]-o.hi[0])
		for a := o.hi[0]; a < o.hi[1]; a += bs {
			prog(0, total, "Controllo dei canali 961-990…")
			cmd := []byte{'R', byte(a >> 8), byte(a), byte(bs)}
			if err := s.p.Write(cmd); err != nil {
				return "", err
			}
			d, err := s.recv(4 + bs)
			if err == nil && !bytes.Equal(d[:4], cmd) {
				err = rErr("risposta inattesa")
			}
			if err != nil {
				s.exit()
				return "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			blk := append([]byte{}, d[4:]...)
			btCrypt(blk, key)
			cur = append(cur, blk...)
		}
		want := img[o.hi[0]:o.hi[1]]
		if !bytes.Equal(cur, want) {
			if _, junk := rt950HiCount(cur[:0x3C0]); junk > 0 {
				s.exit()
				return "", rErr("nella radio, dopo il canale 960, ci sono dati che non sono canali: la radio sembra avere il firmware fino a 0.23 (960 canali). " +
					"Nessun dato è stato scritto. Usa la scheda della versione con 960 canali, oppure aggiorna il firmware della radio")
			}
			ranges = append([][2]int{ranges[0], o.hi}, ranges[1:]...)
			total += o.hi[1] - o.hi[0]
		}
	}
	done := 0
	for _, r := range ranges {
		for a := r[0]; a < r[1]; a += bs {
			prog(done, total, "Scrittura…")
			blk := append([]byte{}, img[a:a+bs]...)
			if a < 0xF000 && !o.plain {
				btCrypt(blk, key)
			}
			if err := s.p.Write(append([]byte{'W', byte(a >> 8), byte(a), byte(bs)}, blk...)); err != nil {
				return "", err
			}
			ack, err := s.recv(1)
			if err == nil && ack[0] != 0x06 {
				err = rErr("la radio ha rifiutato il blocco")
			}
			if err != nil {
				s.exit()
				return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
			done += bs
		}
	}
	s.exit()
	prog(total, total, "Scrittura completata")
	return "", nil
}
