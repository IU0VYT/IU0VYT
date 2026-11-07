// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// gmrsuv1.py, gmrsv2.py, mursv1.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: BTECH GMRS-V1, GMRS-V2, MURS-V1 e MURS-V2: Baofeng della famiglia
// baofeng_common (stesso protocollo) con le regole dei canali GMRS e MURS degli
// Stati Uniti come in CHIRP (frequenze fisse, ponti a +5 MHz, canali a banda
// stretta e bassa potenza, trasmissione spenta fuori dalle frequenze ammesse).

import "fmt"

const (
	btGmrsV1 = 1 + iota
	btGmrsV2
	btMursV1
	btMursV2
)

const btGmrsV1Format = `
#seekto %s;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown0:4,
     scode:4;
  u8 unknown1;
  u8 unknown2:7,
     lowpower:1;
  u8 unknown3:1,
     wide:1,
     unknown4:2,
     bcl:1,
     scan:1,
     pttid:2;
} memory[%d];

#seekto %s;
struct {
  char name[7];
  u8 unknown1[9];
} names[%d];

#seekto 0x1EF0;
struct {
  char line1[7];
  char line2[7];
} firmware_msg;
`

var (
	btGmrsF1   = []int64{462562500, 462587500, 462612500, 462637500, 462662500, 462687500, 462712500}
	btGmrsF2   = []int64{467562500, 467587500, 467612500, 467637500, 467662500, 467687500, 467712500}
	btGmrsF3   = []int64{462550000, 462575000, 462600000, 462625000, 462650000, 462675000, 462700000, 462725000}
	btGmrsOrig = concatI64(btGmrsF1, btGmrsF3, btGmrsF3)
	btGmrs2017 = concatI64(btGmrsF1, btGmrsF2, btGmrsF3, btGmrsF3)
	btGmrsAll  = concatI64(btGmrsF1, btGmrsF2, btGmrsF3)
	btMursNFM  = []int64{151820000, 151880000, 151940000}
	btMursFM   = []int64{154570000, 154600000}
	btMursAll  = concatI64(btMursNFM, btMursFM)
	btMurs15   = concatI64(btMursAll, btMursAll, btMursAll)
)

func concatI64(ls ...[]int64) []int64 {
	var out []int64
	for _, l := range ls {
		out = append(out, l...)
	}
	return out
}

func inI64(l []int64, v int64) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

func init() {
	cs := charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./"
	gm := func(id, model, rclass string, slots, first int, memAt, names string, kind int, pw []genPower, bands []band,
		magic []byte, fileid []string, match bool, ranges [][2]int, note string) *genModel {
		o := &bfcOpt{format: fmt.Sprintf(btGmrsV1Format, memAt, slots, names, slots), identN: 8, magics: [][]byte{magic},
			fileid: fileid, match: match, ranges: ranges, dtcs: uv5rDTCS, scodeN: 15, pttidN: 4, gmrs: kind, btech: true}
		return &genModel{ID: id, Vendor: "BTECH", Model: model, RClass: rclass, Name: "BTECH " + model, Slots: slots,
			First: first, NameMax: 7, Charset: cs, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: uv5rDTCS,
			Baud: 9600, Fam: bfcFam{}, Opt: o, Notes: note}
	}
	std := []band{{130000000, 180000000}, {400000000, 521000000}}
	r1 := [][2]int{{0x0000, 0x0DF0}, {0x0E00, 0x1800}, {0x1EE0, 0x1EF0}, {0x1F60, 0x1F70}, {0x1F80, 0x1F90}, {0x1FC0, 0x1FD0}}
	r2 := [][2]int{{0x0000, 0x1EF0}, {0x1F60, 0x1F70}, {0x1F80, 0x1F90}, {0x1FC0, 0x1FD0}}
	mg1 := []byte{0x50, 0x5F, 0x20, 0x15, 0x12, 0x15, 0x4D}
	mg2 := []byte{0x50, 0x5F, 0x20, 0x21, 0x12, 0x15, 0x4D}
	gmrsNote := "Radio GMRS per gli Stati Uniti: in Italia la banda GMRS non è ammessa. "
	genRegister(
		gm("bt-gmrsv1", "GMRS-V1", "GMRSV1", 128, 0, "0x0000", "0x1000", btGmrsV1,
			[]genPower{{"Alta 5 W", 5, 0}, {"Bassa 2 W", 2, 2}}, std, mg1,
			[]string{"US32412", "US32418", "US32416", "US32411"}, true, r1,
			gmrsNote+"Canali GMRS fissi come in CHIRP: 0-22 (firmware originale) oppure 1-30 (firmware con le regole del "+
				"2017, US32418), con frequenza e passo dei ponti non modificabili; fuori da quei canali le frequenze non GMRS "+
				"sono solo in ricezione. Accendi la radio con il volume a metà e su un canale libero."),
		gm("bt-gmrsv2", "GMRS-V2", "GMRSV2", 200, 0, "0x0000", "0x1000", btGmrsV2,
			[]genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}}, std, mg2, []string{"US32419"}, false, r2,
			gmrsNote+"Come in CHIRP: fuori dalle frequenze GMRS la trasmissione è spenta, sui canali dei ponti solo simplex "+
				"o +5 MHz, sui canali 467 MHz «solo portatile» banda stretta e bassa potenza. Accendi la radio con il volume a "+
				"metà e su un canale libero."),
		gm("bt-mursv1", "MURS-V1", "MURSV1", 15, 1, "0x0010", "0x1010", btMursV1,
			[]genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}, []band{{151820000, 154600250}}, mg1,
			[]string{"USM2402"}, true, r1,
			"Radio MURS per gli Stati Uniti (151-154 MHz): in Italia la banda MURS non è ammessa. 15 canali sulle "+
				"frequenze MURS, sempre in simplex; 151,820-151,940 MHz a banda stretta. Accendi la radio con il volume a "+
				"metà e su un canale libero."),
		gm("bt-mursv2", "MURS-V2", "MURSV2", 200, 0, "0x0000", "0x1000", btMursV2,
			[]genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}, std, mg2, []string{"USM2403"}, false, r2,
			"Radio MURS per gli Stati Uniti: in Italia la banda MURS non è ammessa. Come in CHIRP: fuori dalle frequenze "+
				"MURS la trasmissione è spenta, sulle frequenze MURS solo simplex e 151,820-151,940 MHz a banda stretta. "+
				"Accendi la radio con il volume a metà e su un canale libero."),
	)
}

// btGmrsOrigFW: GMRS-V1 con il firmware originale (regole prima del 2017)
func btGmrsOrigFW(img []byte) bool {
	if len(img) < 0x1EF7 {
		return false
	}
	switch string(img[0x1EF0:0x1EF7]) {
	case "US32416", "US32411", "US32412":
		return true
	}
	return false
}

// btGmrsView: il canale come lo mostra CHIRP (frequenze fisse, passo, banda e
// potenza imposti) e i campi che CHIRP non lascia cambiare
func btGmrsView(m *genModel, img []byte, i int, c *Chan) (imm map[string]bool) {
	o := bfco(m)
	n := i + m.First
	imm = map[string]bool{}
	simplex := func() { c.Tx, c.TxOff = c.Rx, false }
	plus5 := func() { c.Tx, c.TxOff = c.Rx+5000000, false }
	low := m.powerSem(1)
	switch o.gmrs {
	case btGmrsV1:
		if btGmrsOrigFW(img) {
			if inI64(btGmrsOrig, c.Rx) {
				if n >= 0 && n <= 22 {
					c.Rx = btGmrsOrig[n]
					imm["empty"], imm["freq"] = true, true
				}
				switch {
				case n >= 0 && n <= 14:
					simplex()
					imm["duplex"] = true
				case n >= 15 && n <= 22:
					plus5()
					imm["duplex"] = true
				}
			} else if n > 22 {
				c.TxOff, c.Tx = true, c.Rx
				imm = map[string]bool{"duplex": true}
			}
		} else {
			if inI64(btGmrs2017, c.Rx) {
				if n >= 1 && n <= 30 {
					c.Rx = btGmrs2017[n-1]
					imm["empty"], imm["freq"] = true, true
				}
				switch {
				case n >= 1 && n <= 7, n >= 15 && n <= 22:
					simplex()
					imm["duplex"] = true
				case n >= 8 && n <= 14:
					simplex()
					c.Narrow, c.Power = true, low
					imm["duplex"], imm["mode"], imm["power"] = true, true, true
				case n >= 23 && n <= 30:
					plus5()
					imm["duplex"] = true
				}
			} else if n < 1 || n > 30 {
				c.TxOff, c.Tx = true, c.Rx
				imm = map[string]bool{"duplex": true}
			}
		}
	case btGmrsV2:
		switch {
		case inI64(btGmrsF1, c.Rx):
			simplex()
		case inI64(btGmrsF2, c.Rx):
			simplex()
			c.Narrow, c.Power = true, low
		case inI64(btGmrsF3, c.Rx):
		default:
			c.TxOff, c.Tx = true, c.Rx
		}
	case btMursV1:
		if inI64(btMurs15, c.Rx) {
			if n >= 1 && n <= 15 {
				c.Rx = btMurs15[n-1]
				simplex()
				imm["empty"], imm["freq"], imm["duplex"] = true, true, true
			}
			if inI64(btMursNFM, c.Rx) {
				c.Narrow = true
				imm["mode"] = true
			}
		}
	case btMursV2:
		switch {
		case inI64(btMursNFM, c.Rx):
			simplex()
			c.Narrow = true
		case inI64(btMursFM, c.Rx):
			simplex()
		default:
			c.TxOff, c.Tx = true, c.Rx
		}
	}
	return imm
}

func (f bfcFam) gmrsGet(m *genModel, img []byte, i int) *Chan {
	o := bfco(m)
	mem, nam := f.recs(m, i)
	if mem.Raw(img)[0] == 0xFF {
		return nil
	}
	rx := mem.F("rxfreq").Int(img)
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	txr := mem.F("txfreq").Raw(img)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	switch o.gmrs {
	case btGmrsV1:
		if string(txr) == string(ff) {
			c.TxOff = true
		} else {
			c.Tx = mem.F("txfreq").Int(img) * 10
		}
	case btGmrsV2, btMursV2:
		if string(txr) == string(ff) || string(txr) == "\x00\x00\x00\x00" {
			c.TxOff = true
		} else if mem.F("txfreq").Int(img) > rx {
			c.Tx = c.Rx + 5000000
		}
	}
	c.Name = pyName(nam.F("name").Raw(img), true)
	c.TxTone = uv5rToneDec(mem.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(mem.F("rxtone").Int(img), true)
	c.Scan = mem.F("scan").Bool(img)
	c.Power = m.powerSem(int(mem.F("lowpower").Int(img)))
	c.Narrow = !mem.F("wide").Bool(img)
	btGmrsView(m, img, i, c)
	return c
}

func (f bfcFam) gmrsSet(m *genModel, img []byte, i int, c *Chan) error {
	o := bfco(m)
	mem, nam := f.recs(m, i)
	if c == nil {
		mem.Fill(img, 0xFF)
		nam.Fill(img, 0xFF)
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
	var bcl, pttid, scode int64
	if mem.Raw(img)[0] != 0xFF {
		bcl, pttid, scode = mem.F("bcl").Int(img), mem.F("pttid").Int(img), mem.F("scode").Int(img)
	}
	mem.Fill(img, 0)
	mem.F("rxfreq").SetInt(img, c.Rx/10)
	switch {
	case o.gmrs == btMursV1:
		mem.F("txfreq").SetInt(img, c.Rx/10)
	case c.TxOff:
		mem.F("txfreq").Fill(img, 0xFF)
	case o.gmrs == btGmrsV1:
		mem.F("txfreq").SetInt(img, c.Tx/10)
	case c.Tx > c.Rx:
		mem.F("txfreq").SetInt(img, (c.Rx+5000000)/10)
	default:
		mem.F("txfreq").SetInt(img, c.Rx/10)
	}
	nf := nam.F("name")
	for k := 0; k < m.NameMax; k++ {
		if k < len(c.Name) {
			img[nf.Off+k] = c.Name[k]
		} else {
			img[nf.Off+k] = 0xFF
		}
	}
	mem.F("txtone").SetInt(img, uv5rToneEnc(tt))
	mem.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	mem.F("scan").SetBool(img, c.Scan)
	mem.F("wide").SetBool(img, !c.Narrow)
	mem.F("lowpower").SetInt(img, int64(m.powerIndex(c.Power)))
	mem.F("bcl").SetInt(img, bcl)
	mem.F("pttid").SetInt(img, pttid)
	mem.F("scode").SetInt(img, scode)
	return nil
}

// LockedImg: i campi che CHIRP non lascia cambiare (immutable) su GMRS-V1 e
// MURS-V1, e su tutte e quattro i canali che la radio (e CHIRP) mostrerebbe in
// un altro modo (trasmissione spenta fuori dalle frequenze ammesse, passo,
// banda e potenza imposti)
func (f bfcFam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	o := bfco(m)
	if o.gmrs == 0 {
		return nil
	}
	n := i + m.First
	if old != nil && (o.gmrs == btGmrsV1 || o.gmrs == btMursV1) {
		oc := *old
		imm := btGmrsView(m, img, i, &oc)
		switch {
		case imm["empty"] && want == nil:
			return fmt.Errorf("il canale %d è un canale fisso: non si può cancellare", n)
		case want == nil:
		case imm["freq"] && want.Rx != old.Rx:
			return fmt.Errorf("il canale %d è un canale fisso a %s MHz: frequenza non modificabile", n, fmtMHz(old.Rx))
		case imm["duplex"] && (want.TxOff != old.TxOff || !old.TxOff && want.Tx-want.Rx != old.Tx-old.Rx):
			if old.TxOff {
				return fmt.Errorf("sul canale %d la trasmissione è spenta (frequenza non ammessa)", n)
			}
			return fmt.Errorf("sul canale %d il passo è fisso (%s)", n, map[bool]string{true: "simplex", false: "+5 MHz"}[old.Tx == old.Rx])
		case imm["mode"] && want.Narrow != old.Narrow:
			return fmt.Errorf("il canale %d è sempre a banda stretta", n)
		case imm["power"] && want.Power != old.Power:
			return fmt.Errorf("il canale %d è sempre a bassa potenza", n)
		}
	}
	if want == nil {
		return nil
	}
	v := *want
	btGmrsView(m, img, i, &v)
	switch {
	case v.Rx != want.Rx:
		return fmt.Errorf("sul canale %d la radio usa la frequenza GMRS/MURS %s MHz", n, fmtMHz(v.Rx))
	case v.TxOff && !want.TxOff:
		return fmt.Errorf("canale %d: %s MHz non è una frequenza ammessa in trasmissione su questa posizione: imposta TX OFF", n, fmtMHz(want.Rx))
	case !v.TxOff && !want.TxOff && v.Tx != want.Tx:
		if v.Tx == v.Rx {
			return fmt.Errorf("canale %d: su questa frequenza solo simplex (TX = RX)", n)
		}
		return fmt.Errorf("canale %d: su questa frequenza solo simplex o +5 MHz", n)
	case v.Narrow != want.Narrow:
		return fmt.Errorf("canale %d: su questa frequenza solo banda stretta", n)
	case v.Power != want.Power:
		return fmt.Errorf("canale %d: su questa frequenza solo bassa potenza", n)
	}
	return nil
}

// lockedSlotImg: posizioni con un canale fisso (frequenza non modificabile)
func (f bfcFam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	o := bfco(m)
	if o.gmrs != btGmrsV1 && o.gmrs != btMursV1 {
		return false
	}
	c := f.Get(m, img, i)
	if c == nil {
		return false
	}
	return btGmrsView(m, img, i, c)["freq"]
}
