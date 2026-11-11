// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// mml_jc8810.py (Copyright 2023 Jim Unroe KC9HI e altri autori di CHIRP).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia MML JC-8810 (4.1): Radtel RT-470, RT-470L, RT-470X, RT-470X BT,
// RT-630, RT-495, Hiroyasu HI-8811, Anysecu UV-A37, Talkpod A36plus e A36plus
// 8 W, Abbree AR-730. Protocollo come la famiglia RT-900 senza cifratura
// (rt900Opt.plain): "PROGRAMJC81U" o "PROGRAMJC37U" → 0x06 (fino a 5
// tentativi); 'F' → identificativo di 8 byte; lettura 'R' indirizzo 0x40 →
// 'R' indirizzo 0x40 + 64 byte; scrittura 'W' … → 0x06; uscita 'E'. 57600 baud.
// Record di 32 byte come la RT-900; potenze: 0 alta, 1 bassa, 2 media.

import (
	"bytes"
	"strconv"
	"strings"
)

const jc8810Format = `
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
     unknown5:1,
     learning:1;
  lbcd code[3];
  u8 unknown6;
  char name[12];
} memory[%d];
`

// jc8810Fam: come rt900Fam (protocollo, memoria, radio simulata) con le regole
// dei canali di mml_jc8810.py
type jc8810Fam struct{ rt900Fam }

type jc8810Opt struct {
	rt900Opt
	pow     []int64 // valore di txpower per ogni potenza della radio (nell'ordine di Powers)
	airband *band   // RT-470X: banda aerea, sempre in AM (nessun bit AM nel canale)
}

func jco(m *genModel) *jc8810Opt { return m.Opt.(*jc8810Opt) }

func (f jc8810Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	c.Scan = r.F("scan").Bool(img)
	pv := r.F("txpower").Int(img)
	c.Power = m.Powers[0].Sem // valore sconosciuto: CHIRP lascia la potenza predefinita (la prima)
	for k, v := range jco(m).pow {
		if v == pv {
			c.Power = m.Powers[k].Sem
		}
	}
	c.Narrow = r.F("narrow").Bool(img)
	// come str() di CHIRP: un carattere per byte (latin-1, es. nomi cinesi GBK)
	c.Name = latin1Str(strings.ReplaceAll(pyName(r.F("name").Raw(img), true), "\x00", ""))
	return c
}

func (f jc8810Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := jco(m)
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
	var encrypt, scramble, bcl, pttid, scode int64
	if !wasEmpty {
		encrypt, scramble, bcl = r.F("encrypt").Int(img), r.F("scramble").Int(img), r.F("bcl").Int(img)
		pttid, scode = r.F("pttid").Int(img), r.F("scode").Int(img)
	}
	r.SetRaw(img, append(bytes.Repeat([]byte{0}, 16), bytes.Repeat([]byte{0xFF}, 16)...))
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("name").SetStr(img, latin1Bytes(c.Name), 0xFF)
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	// banda aerea della RT-470X: modo AM/NAM in CHIRP, che non scrive mai il bit
	// della banda stretta (narrow vale solo per «NFM»)
	narrow := c.Narrow
	if o.airband != nil && c.Rx >= o.airband.Lo && c.Rx <= o.airband.Hi {
		narrow = false
	}
	r.F("narrow").SetBool(img, narrow)
	r.F("txpower").SetInt(img, o.pow[m.powerIndex(c.Power)])
	if !wasEmpty {
		r.F("encrypt").SetInt(img, encrypt)
		r.F("scramble").SetInt(img, scramble)
		r.F("bcl").SetInt(img, bcl)
		r.F("pttid").SetInt(img, pttid)
		r.F("scode").SetInt(img, scode)
	}
	return nil
}

func init() {
	fp := func(l ...string) [][]byte {
		var out [][]byte
		for _, s := range l {
			out = append(out, []byte(s))
		}
		return out
	}
	base := [][2]int{{0x0000, 0x2000}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}}
	rng := func(chans int, tail ...[2]int) [][2]int {
		r := append([][2]int{{0x0000, chans}}, base[1:]...)
		return append(r, tail...)
	}
	h3 := func(h, m, l float64) []genPower {
		return []genPower{{"Alta " + fmtW(h), h, 0}, {"Media " + fmtW(m), m, 1}, {"Bassa " + fmtW(l), l, 2}}
	}
	h2 := func(h, l float64) []genPower {
		return []genPower{{"Alta " + fmtW(h), h, 0}, {"Bassa " + fmtW(l), l, 2}}
	}
	p3, p2 := []int64{0, 2, 1}, []int64{0, 1} // H, M, L → 0, 2, 1; H, L → 0, 1
	bands := func(l ...float64) []band {
		var out []band
		for k := 0; k+1 < len(l); k += 2 {
			out = append(out, band{int64(l[k] * 1e6), int64(l[k+1] * 1e6)})
		}
		return out
	}
	rt470fp := fp("\x00\x00\x00\x26\x00\x20\xD8\x04", "\x00\x00\x00\x42\x00\x20\xF0\x04", "\x00\x00\x00\x4A\x00\x20\xF8\x04",
		"\x00\x00\x00\x3A\x00\x20\xE8\x04", "\x00\x00\x00\x42\x00\x20\xec\x04",
		"\x00\x00\x00\x28\x00\x20\xD4\x04", "\x00\x00\x00\x2C\x00\x20\xD8\x04", "\x00\x00\x00\x36\x00\x20\xDC\x04")
	rt470lfp := fp("\x00\x00\x00\xfe\x00\x20\xAC\x04", "\x00\x00\x00\x20\x00\x20\xCC\x04", "\x00\x00\x00\x20\x00\x20\x07\x00")
	rt470xfp := fp("\x00\x00\x00\x20\x00\x20\xCC\x04", "\x00\x00\x00\x2C\x00\x20\xD8\x04", "\x00\x00\x00\x36\x00\x20\xDC\x04")
	air := &band{118000000, 136000000}
	b470l := bands(108, 136, 136, 179, 220, 260, 330, 400, 400, 520)
	type v struct {
		id, vendor, model, rclass, name string
		slots                           int
		magic                           string
		fps                             [][]byte
		ranges                          [][2]int
		memsize                         int
		powers                          []genPower
		pow                             []int64
		bands                           []band
		air                             *band
	}
	list := []v{
		{"rt-rt470", "Radtel", "RT-470", "RT470Radio", "Radtel RT-470", 256, "PROGRAMJC81U", rt470fp, rng(0x2000, [2]int{0xB000, 0xB300}), 0xB300,
			h3(10, 8, 4), p3, bands(16, 100, 100, 136, 136, 200, 200, 300, 300, 400, 400, 560, 740, 1000), nil},
		{"rt-rt470l", "Radtel", "RT-470L", "RT470LRadio", "Radtel RT-470L", 256, "PROGRAMJC81U", rt470lfp, rng(0x2000, [2]int{0xB000, 0xB300}), 0xB300,
			h3(5, 4, 2), p3, b470l, nil},
		{"rt-rt470x", "Radtel", "RT-470X", "RT470XRadio", "Radtel RT-470X", 256, "PROGRAMJC81U", rt470xfp, rng(0x2000, [2]int{0xB000, 0xB300}), 0xB300,
			h3(5, 4, 2), p3, bands(18, 1000), air},
		{"rt-rt470xbt", "Radtel", "RT-470X_BT", "RT470XPlusRadio", "Radtel RT-470X Plus (Bluetooth)", 256, "PROGRAMJC81U",
			fp("\x01\x36\x01\x80\x04\x00\x05\x20"), rng(0x2000, [2]int{0xB000, 0xB300}), 0xB300, h3(5, 4, 2), p3, bands(18, 1000), air},
		{"hy-hi8811", "Hiroyasu", "HI-8811", "HI8811Radio", "Hiroyasu HI-8811", 256, "PROGRAMJC81U",
			fp("\x00\x00\x00\xfe\x00\x20\xAC\x04", "\x00\x00\x00\x20\x00\x20\xCC\x04", "\x00\x00\x00\x20\x00\x20\x07\x00",
				"\x00\x00\x00\x28\x00\x20\xD4\x04", "\x00\x00\x00\x28\x00\x20\x07\x00"),
			rng(0x2000, [2]int{0xB000, 0xB300}), 0xB300, h3(5, 4, 2), p3, b470l, nil},
		{"as-uva37", "Anysecu", "UV-A37", "UVA37Radio", "Anysecu UV-A37", 256, "PROGRAMJC37U",
			fp("\x00\x00\x00\xE4\x00\x20\x94\x04", "\x00\x00\x00\xE8\x00\x20\x98\x04"), rng(0x2000, [2]int{0xB000, 0xB440}), 0xB440,
			h2(5, 1), p2, bands(108, 136, 136, 174, 200, 260, 350, 390, 400, 520), nil},
		{"ab-ar730", "Abbree", "AR-730", "AR730Radio", "Abbree AR-730", 256, "PROGRAMJC37U",
			fp("\x00\x00\x00\xE4\x00\x20\x94\x04", "\x00\x00\x00\xE8\x00\x20\x98\x04"), rng(0x2000, [2]int{0xB000, 0xB440}), 0xB440,
			h2(5, 1), p2, bands(108, 136, 136, 180, 200, 260, 350, 390, 400, 520), nil},
		{"tp-a36plus", "Talkpod", "A36plus", "A36plusRadio", "Talkpod A36plus", 512, "PROGRAMJC37U",
			fp("\x00\x00\x00\x42\x00\x20\xF0\x04", "\x00\x00\x00\x5A\x00\x20\x08\x05", "\x00\x00\x00\x9E\x00\x20\x0C\x05",
				"\x00\x00\x00\xFA\x00\x20\x40\x05", "\x00\x00\x00\x9C\x00\x20\x04\x05"),
			rng(0x4000, [2]int{0xB000, 0xB440}), 0xB440, h2(5, 1), p2, bands(108, 136, 136, 180, 200, 260, 350, 400, 400, 520), nil},
		{"tp-a36plus8w", "Talkpod", "A36plus_8w", "A36plus8wRadio", "Talkpod A36plus 8 W", 512, "PROGRAMJC37U",
			fp("\x00\x00\x00\xFA\x00\x20\x40\x05", "\x00\x00\x00\xD8\x00\x20\x58\x05"),
			rng(0x4000, [2]int{0xB000, 0xB440}), 0xB440, h2(8, 1), p2, bands(108, 136, 136, 180, 200, 260, 350, 400, 400, 520), nil},
		{"rt-rt630", "Radtel", "RT-630", "RT630Radio", "Radtel RT-630", 256, "PROGRAMJC81U",
			fp("\x00\x00\x00\x32\x00\x20\xD8\x04", "\x00\x00\x00\x36\x00\x20\xDC\x04"),
			rng(0x2000, [2]int{0xB000, 0xB2C0}, [2]int{0xB500, 0xB740}), 0xB740, h3(5, 4, 2), p3, bands(18, 108, 108, 136, 136, 300, 300, 660, 840, 1000), nil},
		{"rt-rt495", "Radtel", "RT-495", "RT495Radio", "Radtel RT-495", 256, "PROGRAMJC81U",
			fp("\x00\x00\x00\x24\x00\x20\xD0\x04", "\x00\x00\x00\x32\x00\x20\xD8\x04"),
			rng(0x2000, [2]int{0xB000, 0xB2C0}, [2]int{0xB500, 0xB740}), 0xB740, h3(5, 4, 2), p3, bands(18, 108, 108, 136, 136, 300, 300, 660, 840, 1000), nil},
	}
	// 4.18: Radtel RT-330 e RT-400, dai programmi del costruttore («Radio CPS» e
	// «JJCC-7700 CPS», stessa base di codice): stesso protocollo e stesso record
	// dei canali della famiglia. Il programma non controlla l'identificativo: la
	// RT-330 ha un saluto solo suo («PROGRAMYF68U», 115200 baud, 512 canali, potenze
	// alta 0 - media 1 - bassa 2); la RT-400 usa il saluto delle RT-470
	// («PROGRAMJC81U»), per cui si rifiutano gli identificativi degli altri modelli.
	var jc81 [][]byte
	for _, x := range list {
		if x.magic == "PROGRAMJC81U" {
			jc81 = append(jc81, x.fps...)
		}
	}
	pn := func(l ...string) []genPower {
		sem := map[string]int{"Alta": 0, "Media": 1, "Bassa": 2}
		var out []genPower
		for _, s := range l {
			out = append(out, genPower{s, 0, sem[s]})
		}
		return out
	}
	type vx struct {
		baud  int      // velocità (0: 57600)
		fpNot [][]byte // identificativi di altri modelli con lo stesso saluto: rifiutati
		note  string
	}
	extra := map[string]vx{
		"rt-rt330": {baud: 115200,
			note: "Accendi la radio con il cavo collegato. 512 canali, 115200 baud. Protocollo e memoria ricavati dal programma del " +
				"costruttore («Radio CPS» della RT-330), che usa lo stesso record dei canali della famiglia MML JC-8810 (Radtel RT-470). " +
				"Si leggono e si scrivono le stesse aree del programma del costruttore."},
		"rt-rt400": {fpNot: jc81,
			note: "Accendi la radio con il cavo collegato. 256 canali. Protocollo e memoria ricavati dal programma del costruttore " +
				"(«JJCC-7700 CPS» della RT-400), che usa lo stesso saluto e lo stesso record dei canali delle Radtel RT-470. Il programma " +
				"del costruttore non controlla l'identificativo della radio: questa scheda rifiuta le radio che si identificano come " +
				"uno degli altri modelli della famiglia (RT-470, RT-630, RT-495, HI-8811) e accetta le altre. Si leggono e si scrivono " +
				"le stesse aree del programma del costruttore."},
	}
	list = append(list,
		v{"rt-rt330", "Radtel", "RT-330", "IU0VYT_Radtel_RT330", "Radtel RT-330", 512, "PROGRAMYF68U", nil,
			rng(0x4000, [2]int{0xB000, 0xB180}), 0xB180, pn("Alta", "Media", "Bassa"), []int64{0, 1, 2}, bands(100, 520), nil},
		v{"rt-rt400", "Radtel", "RT-400", "IU0VYT_Radtel_RT400", "Radtel RT-400", 256, "PROGRAMJC81U", nil,
			rng(0x2000, [2]int{0xB000, 0xB2C0}), 0xB300, pn("Alta", "Media", "Bassa"), p3, bands(18, 1000), nil})
	for _, x := range list {
		ex, isNew := extra[x.id]
		o := &jc8810Opt{rt900Opt: rt900Opt{magic: []byte(x.magic), fp1: x.fps, fpNot: ex.fpNot, ranges: x.ranges, memsize: x.memsize, format: jc8810Format, plain: true},
			pow: x.pow, airband: x.air}
		note := "Accendi la radio con il cavo collegato. Driver di CHIRP della famiglia MML JC-8810 (Radtel RT-470)."
		if x.air != nil {
			note += " In banda aerea (118-136 MHz) la radio riceve in AM."
		}
		baud := 57600
		if isNew {
			// si leggono solo le aree che il programma del costruttore legge
			o.reads, note = x.ranges, ex.note
			if ex.baud != 0 {
				baud = ex.baud
			}
		}
		genRegister(&genModel{ID: x.id, Vendor: x.vendor, Model: x.model, RClass: x.rclass,
			Name: x.name, Slots: x.slots, First: 1, NameMax: 12,
			Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
			Powers:  x.powers, Bands: x.bands, Narrow: true, HasScan: true, NoChirp: isNew,
			DCS: uv5rDTCS, Baud: baud, Fam: jc8810Fam{}, Opt: o, Notes: note})
	}
}

// fmtW: potenza in watt come etichetta («10 W», «0,5 W»)
func fmtW(w float64) string {
	return strings.Replace(strconv.FormatFloat(w, 'f', -1, 64), ".", ",", 1) + " W"
}

// latin1Str: byte → caratteri latin-1 (come str() di CHIRP sui campi char)
func latin1Str(b string) string {
	rs := make([]rune, 0, len(b))
	for i := 0; i < len(b); i++ {
		rs = append(rs, rune(b[i]))
	}
	return string(rs)
}

// latin1Bytes: caratteri fino a U+00FF → un byte ciascuno (gli altri: «?»)
func latin1Bytes(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
			r = '?'
		}
		out = append(out, byte(r))
	}
	return string(out)
}
