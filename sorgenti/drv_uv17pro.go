// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: baofeng_uv17Pro.py.

package main

// Famiglia Baofeng UV-17Pro (driver baofeng_uv17Pro.py di CHIRP): UV-17Pro,
// UV-17Pro GPS, UV-17R Plus, UV-5RM, UV-5RM Plus, K5 Plus, UV-5RH, GM-5RH,
// UV-5G Plus, UV-21Pro, UV-25, UV-28 Plus, UV-32, GM-21, BF-F8HP Pro, K6,
// UV-5R Mini, UV-5G Mini. Fino a 1000 canali, 115200 baud.
//
// Protocollo:
//   identificazione: stringa di 16 byte (es. "PROGRAMBFNORMALU") → 0x06;
//   poi i comandi 'F' (risposta 16 byte), 'M' e il pacchetto "SEND…" con la
//   chiave di cifratura, ciascuno con la sua risposta.
//   lettura: 'R' indirizzo(2, BE) 0x40 → 4 + 64 byte (dati cifrati);
//   scrittura: 'W' indirizzo 0x40 + 64 byte cifrati → 0x06.
//   La memoria è letta in più aree (es. 0x0000-0x803F, 0x9000, 0xA000, 0xD000)
//   che nell'immagine stanno una dopo l'altra.
// Cifratura: la stessa della Bajeton 8300 (btCrypt) con una delle 20 chiavi.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const uv17ProFormat = `
struct memory_obj {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 scode;
  u8 pttid;
  u8 unknown7:2,
     scramble:2,
     unknown8:2,
     lowpower:2;
  u8 unknown1:1,
     wide:1,
     sqmode:2,
     bcl:1,
     scan:1,
     unknown2:1,
     fhss:1;
  u8 unknown3;
  u8 unknown4;
  u8 unknown5;
  u8 unknown6;
  char name[12];
};
struct memory_obj memory[%d];
`

var (
	uv17MagicSendStd = []byte("SEND\x21\x05\x0D\x01\x01\x01\x04\x11\x08\x05\x0D\x0D\x01\x11\x0F\x09\x12\x09\x10\x04\x00")
	uv17MagicSend5RM = []byte("SEND\x13\x0F\x0D\x13\x0E\x0E\x0E\x0D\x0A\x0E\x0E\x02\x0F\x08\x04\x11\x0C\x00\x0D\x0E\x00")
	uv17MagicSendF8  = []byte("SEND\x12\x0D\x0A\x0A\x10\x03\x0D\x02\x11\x0C\x12\x0A\x11\x06\x04\x0E\x02\x09\x0D\x00\x00")
	gmrsHHOnly       = []int64{467562500, 467587500, 467612500, 467637500, 467662500, 467687500, 467712500}
)

type uv17Magic struct {
	cmd  []byte
	resp int
}

type uv17Opt struct {
	idents   [][]byte
	magics   []uv17Magic
	starts   []int
	sizes    []int
	total    int
	encr     bool
	encrsym  int
	am       bool   // modo AM tra quelli della radio
	airbands []band // gamme mostrate in AM
	scodeN   int    // elementi della lista S-CODE (per la normalizzazione di CHIRP)
	scramble bool
	sqExtra  string // "list" (BF-F8HP Pro) o "bool" (K6, UV-5R Mini): campi sqmode/fhss conservati
	gmrs     bool   // limiti GMRS (GM-21)
	lowIdx   int
	uv5gmini bool // UV-5G Mini: canali GMRS 8-14 sempre stretti e a bassa potenza
	// modelli ricavati dal programma "BF-UVs CPS" del costruttore: dopo 'M' la radio
	// invia il nome del modello, che deve essere uno di questi (come nel programma)
	modelIDs []string
}

type uv17Fam struct{}

func init() {
	air := band{108000000, 135999999}
	vhf := band{136000000, 174000000}
	vhf2 := band{200000000, 260000000}
	uhf := band{400000000, 520000000}
	uhf2 := band{350000000, 390000000}
	b21 := []band{{108000000, 135999999}, {136000000, 174000000}, {220000000, 260000000}, {350000000, 390000000}, {400000000, 479999999}, {480000000, 520000000}}
	stdMag := []uv17Magic{{[]byte{0x46}, 16}, {[]byte{0x4D}, 15}, {uv17MagicSendStd, 1}}
	gpsMag := []uv17Magic{{[]byte{0x46}, 16}, {[]byte{0x4D}, 7}, {uv17MagicSendStd, 1}}
	base := uv17Opt{idents: [][]byte{[]byte("PROGRAMBFNORMALU")}, magics: stdMag,
		starts: []int{0x0000, 0x9000, 0xA000, 0xD000}, sizes: []int{0x8040, 0x0040, 0x02C0, 0x0040}, total: 0x8380,
		encr: true, encrsym: 1, airbands: []band{air}, scodeN: 20, lowIdx: 1}
	p2 := func(h, l float64) []genPower {
		return []genPower{{fmt.Sprintf("Alta %g W", h), h, 0}, {fmt.Sprintf("Bassa %g W", l), l, 2}}
	}
	p3 := func(h, l, m float64) []genPower {
		return []genPower{{fmt.Sprintf("Alta %g W", h), h, 0}, {fmt.Sprintf("Bassa %g W", l), l, 2}, {fmt.Sprintf("Media %g W", m), m, 1}}
	}
	cs := charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./"
	mk := func(id, vendor, model, rclass, name string, o uv17Opt, slots int, pw []genPower, bands []band) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: slots, First: 1,
			NameMax: 12, Charset: cs, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 115200,
			Fam: uv17Fam{}, Opt: &oo, Notes: "Spegni la radio, collega il cavo, poi riaccendila."}
	}
	var ms []*genModel
	o := base
	ms = append(ms, mk("bf-uv17pro", "Baofeng", "UV-17Pro", "UV17Pro", "Baofeng UV-17Pro", o, 1000, p2(5, 1), []band{vhf, vhf2, uhf}))
	o21 := base
	o21.am = true
	ms = append(ms, mk("bf-uv21prov2", "Baofeng", "UV-21ProV2", "UV21ProV2", "Baofeng UV-21Pro V2", o21, 1000, p2(5, 1), b21))
	ms = append(ms, mk("bf-uv25", "Baofeng", "UV-25", "UV25", "Baofeng UV-25", base, 1000, p2(5, 1), []band{vhf, vhf2, uhf}))
	og := base
	og.idents, og.magics, og.am = [][]byte{[]byte("PROGRAMCOLORPROU")}, gpsMag, true
	ms = append(ms, mk("bf-uv17progps", "Baofeng", "UV-17ProGPS", "UV17ProGPS", "Baofeng UV-17Pro GPS", og, 1000, p2(5, 1), []band{air, vhf, vhf2, uhf, uhf2}))
	ogm := og
	ogm.gmrs = true
	ms = append(ms, mk("bf-gm21", "Baofeng", "GM-21", "BaofengGM21", "Baofeng GM-21 (GMRS)", ogm, 1000, p2(5, 1), []band{air, vhf, vhf2, uhf, uhf2}))
	ms = append(ms, mk("bf-uv32", "Baofeng", "UV-32", "UV32", "Baofeng UV-32", og, 1000, p3(10, 2, 5), []band{air, vhf, vhf2, uhf, uhf2}))
	ms = append(ms, mk("bf-uv21progps", "Baofeng", "UV-21ProGPS", "UV21ProGPS", "Baofeng UV-21Pro GPS", og, 1000, p2(5, 1), b21))
	ms = append(ms, mk("bf-uv28plus", "Baofeng", "UV-28Plus", "UV28Plus", "Baofeng UV-28 Plus", og, 1000, p3(10, 2, 5), b21))
	o5 := base
	o5.am, o5.scodeN = true, 15
	b5rm := []band{air, vhf, vhf2, uhf, uhf2}
	ms = append(ms, mk("bf-uv5rm", "Baofeng", "5RM", "BF5RM", "Baofeng UV-5RM", o5, 1000, p3(8, 1, 5), b5rm))
	ms = append(ms, mk("bf-k5plus", "Baofeng", "K5-Plus", "BFK5Plus", "Baofeng K5 Plus", o5, 1000, p3(8, 1, 5), b5rm))
	o5p := o5
	o5p.magics = []uv17Magic{{[]byte{0x46}, 16}, {[]byte{0x4D}, 15}, {uv17MagicSend5RM, 1}}
	o5p.encrsym, o5p.scramble = 13, true
	ms = append(ms, mk("bf-uv5rmplus", "Baofeng", "UV-5RM Plus", "BaofengUV5RMPlus", "Baofeng UV-5RM Plus", o5p, 1000, p3(8, 1, 5), b5rm))
	ogr := base
	ogr.idents, ogr.scodeN = [][]byte{[]byte("PROGRAMBFGMRS05U")}, 15
	ms = append(ms, mk("bf-gm5rh", "Baofeng", "GM-5RH", "GM5RH", "Baofeng GM-5RH (GMRS)", ogr, 1000, p3(5, 0.5, 3), []band{vhf, vhf2, uhf}))
	ms = append(ms, mk("bf-uv5gplus", "Radioddity", "UV-5G Plus", "UV5GPlus", "Baofeng UV-5G Plus", ogr, 1000, p3(5, 0.5, 3), []band{vhf, vhf2, uhf}))
	o17r := base
	o17r.am = true
	ms = append(ms, mk("bf-uv17rplus", "Baofeng", "UV-17R-Plus", "UV17RPlus", "Baofeng UV-17R Plus", o17r, 1000, p2(5, 1), b5rm))
	of8 := base
	of8.idents = [][]byte{[]byte("PROGRAMBF5RTECHU")}
	of8.magics = []uv17Magic{{[]byte{0x46}, 16}, {[]byte{0x4D}, 6}, {uv17MagicSendF8, 1}}
	of8.encrsym, of8.am, of8.sqExtra = 3, true, "list"
	of8.starts, of8.sizes, of8.total = []int{0x0000, 0x9000, 0xA000, 0xD000}, []int{0x8040, 0x0080, 0x02C0, 0x00C0}, 0x8440
	of8.airbands = []band{{108000000, 136999999}, {174000000, 218999999}, {225000000, 399999999}}
	ms = append(ms, mk("bf-f8hppro", "Baofeng", "BF-F8HP-PRO", "F8HPPro", "Baofeng BF-F8HP Pro", of8, 1000, []genPower{{"Alta 10 W", 10, 0}, {"Bassa 1 W", 1, 2}, {"Media 3 W", 3, 1}},
		[]band{{108000000, 136999999}, {137000000, 173999999}, {174000000, 218999999}, {219000000, 224999999}, {225000000, 399999999}, uhf}))
	o5h := base
	o5h.am, o5h.scodeN = true, 15
	ms = append(ms, mk("bf-uv5rh", "Baofeng", "UV-5RH", "UV5RH", "Baofeng UV-5RH", o5h, 1000, p3(10, 2, 5), b5rm))
	ok6 := base
	ok6.idents, ok6.magics = [][]byte{[]byte("PROGRAMBF-K6PROU")}, []uv17Magic{{[]byte{0x46}, 16}}
	ok6.encr, ok6.scramble, ok6.am, ok6.sqExtra = false, true, true, "bool"
	ok6.starts, ok6.sizes, ok6.total = []int{0x0000, 0x9000, 0xA000, 0xC000, 0xD000}, []int{0x8040, 0x0080, 0x02C0, 0x0040, 0x0040}, 0x8400
	ms = append(ms, mk("bf-k6", "Baofeng", "K6", "BFK6", "Baofeng K6 / K63", ok6, 999, p2(5, 1), []band{air, vhf, vhf2, {400000000, 600000000}, uhf2}))
	om := base
	om.idents, om.am, om.sqExtra = [][]byte{[]byte("PROGRAMCOLORPROU")}, true, "bool"
	om.starts, om.sizes, om.total = []int{0x0000, 0x9000, 0xA000}, []int{0x8040, 0x0040, 0x01C0}, 0x8240
	bmini := []band{air, vhf, {350000000, 390000000}, {400000000, 480000000}, {480000000, 520000000}}
	ms = append(ms, mk("bf-uv5rmini", "Baofeng", "UV-5R Mini", "UV5RMini", "Baofeng UV-5R Mini / MINI-5", om, 999, p2(5, 1), bmini))
	omg := om
	omg.idents, omg.sqExtra, omg.uv5gmini = [][]byte{[]byte("PROGRAMGMRS5RMIU"), []byte("PROGRAMCOLORPROU")}, "", true
	ms = append(ms, mk("bf-uv5gmini", "Baofeng", "UV-5G Mini", "UV5GMini", "Baofeng UV-5G Mini (GMRS)", omg, 999, p2(5, 1), bmini))
	// modelli ricavati dai programmi del costruttore (non presenti in CHIRP): stessa
	// mappa della memoria e stesso protocollo della famiglia UV-17Pro.
	// Programma "BF-UVs CPS" (UV-25 Plus, AR-152 Pro, UV-29 Pro, UV-7B): presentazione
	// "PROGRAMCOLORPROU", poi 'F' (16 byte) e 'M' (nome del modello, controllato).
	pw2 := []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}}
	pw3 := []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}, {"Media", 0, 1}}
	cps := func(id, model, rclass, name string, o uv17Opt, pw []genPower, ids ...string) *genModel {
		o.modelIDs = ids
		g := mk(id, "Baofeng", model, rclass, name, o, 1000, pw, b21)
		g.NoChirp = true
		g.Notes += " Protocollo e mappa della memoria ricavati dal programma del costruttore «BF-UVs CPS» (stessi della UV-17Pro GPS di CHIRP); la radio deve identificarsi come " + strings.Join(ids, " o ") + "."
		return g
	}
	oc := base
	oc.idents = [][]byte{[]byte("PROGRAMCOLORPROU")}
	oc.magics = []uv17Magic{{[]byte{0x46}, 16}, {[]byte{0x4D}, 0}, {uv17MagicSendStd, 1}}
	oc.am = true
	genRegister(ms...)
	ms = append(ms[:0:0],
		cps("bf-uv25plus", "UV-25 Plus", "UV17ProGPS", "Baofeng UV-25 Plus", oc, pw3, "UV25PRO"),
		cps("bf-ar152pro", "AR-152 Pro", "UV17ProGPS", "Baofeng AR-152 Pro", oc, pw3, "AR152"),
		cps("bf-uv29pro", "UV-29 Pro", "UV17ProGPS", "Baofeng UV-29 Pro / SW-9", oc, pw3, "UV29PRO", "UV36PRO"),
		cps("bf-uv7b", "UV-7B", "UV17ProGPS", "Baofeng UV-7B", oc, pw2, "UV17PRO"))
	// Programma "T6UV Series CPS" (UV-17R, UV-21R, UV-21L): presentazione
	// "PROGRAMBFNORMALU" come la UV-17Pro; dopo 'M' la radio invia «nome+versione».
	for _, x := range [][3]string{{"bf-uv17r", "UV-17R", "Baofeng UV-17R"}, {"bf-uv21r", "UV-21R", "Baofeng UV-21R"}, {"bf-uv21l", "UV-21L", "Baofeng UV-21L"}} {
		g := mk(x[0], "Baofeng", x[1], "UV17Pro", x[2], base, 1000, pw2, []band{vhf, vhf2, uhf})
		g.NoChirp = true
		g.Notes += " Protocollo e mappa della memoria ricavati dal programma del costruttore «T6UV Series CPS» (stessi della UV-17Pro di CHIRP)."
		ms = append(ms, g)
	}
	genRegister(ms...)
}

func uv17o(m *genModel) *uv17Opt { return m.Opt.(*uv17Opt) }

func (uv17Fam) mem(m *genModel) *bwEl { return bwMust(fmt.Sprintf(uv17ProFormat, m.Slots)) }

func (f uv17Fam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f uv17Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

func (f uv17Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}

func (f uv17Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f uv17Fam) Match(m *genModel, img []byte) bool { return false }

func inAnyBand(hz int64, bs []band) bool {
	for _, b := range bs {
		if hz >= b.Lo && hz <= b.Hi {
			return true
		}
	}
	return false
}

func int64In(l []int64, v int64) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

func (f uv17Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := uv17o(m)
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	txr := r.F("txfreq").Raw(img)
	if bytes.Equal(txr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(txr, []byte{0, 0, 0, 0}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	c.Scan = r.F("scan").Bool(img)
	lp := int(r.F("lowpower").Int(img))
	if lp >= len(m.Powers) {
		lp = 0
	}
	c.Power = m.powerSem(lp)
	// MODES = ["NFM", "FM"]: il bit "wide" a 1 indica la banda stretta
	c.Narrow = r.F("wide").Bool(img)
	nb := r.F("name").Raw(img)
	nm := make([]byte, len(nb))
	for k, b := range nb {
		if b == 0xFF || b == 0 {
			b = ' '
		}
		nm[k] = b
	}
	c.Name = pyName(nm, false)
	if o.gmrs && !c.TxOff && int64In(gmrsHHOnly, c.Tx) {
		c.Narrow, c.Power = true, m.powerSem(o.lowIdx)
	}
	if o.uv5gmini && i+1 >= 8 && i+1 <= 14 {
		c.Narrow, c.Power = true, m.powerSem(o.lowIdx)
	}
	return c
}

func pyMod(a, n int64) int64 {
	r := a % n
	if r < 0 {
		r += n
	}
	return r
}

func (f uv17Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := uv17o(m)
	r := f.rec(m, i)
	wasEmpty := r.Raw(img)[0] == 0xFF
	var bcl, pttid, scode, scramble, sqmode, fhss int64
	if !wasEmpty {
		bcl = r.F("bcl").Int(img)
		pttid = r.F("pttid").Int(img)
		scode = pyMod(r.F("scode").Int(img), int64(o.scodeN))
		scramble = r.F("scramble").Int(img)
		sqmode = r.F("sqmode").Int(img)
		fhss = r.F("fhss").Int(img)
		if o.sqExtra == "bool" && sqmode != 0 {
			sqmode = 1
		}
	}
	if c == nil {
		r.Fill(img, 0xFF)
		if o.gmrs {
			// come CHIRP (GM-21): dopo la cancellazione imposta la banda del
			// canale vuoto secondo i limiti GMRS (modo FM → bit "wide" a 0)
			r.F("wide").SetInt(img, 0)
		}
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
	raw := append(bytes.Repeat([]byte{0}, 16), bytes.Repeat([]byte{0xFF}, 16)...)
	r.SetRaw(img, raw)
	r.F("name").SetStr(img, c.Name, 0xFF)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	narrow := c.Narrow
	r.F("wide").SetBool(img, narrow)
	r.F("lowpower").SetInt(img, int64(m.powerIndex(c.Power)))
	if !wasEmpty {
		r.F("bcl").SetInt(img, bcl)
		r.F("pttid").SetInt(img, pttid)
		r.F("scode").SetInt(img, scode)
		if o.scramble {
			r.F("scramble").SetInt(img, scramble)
		}
		if o.sqExtra != "" {
			r.F("sqmode").SetInt(img, sqmode)
			r.F("fhss").SetInt(img, fhss)
		}
	}
	if o.gmrs && !c.TxOff && int64In(gmrsHHOnly, c.Tx) {
		r.F("lowpower").SetInt(img, int64(o.lowIdx))
		r.F("wide").SetInt(img, 1)
	}
	return nil
}

// lockedSlot: UV-5G Mini, canali GMRS 1-30 fissi (come in CHIRP)
func (f uv17Fam) lockedSlot(m *genModel, i int) bool { return uv17o(m).uv5gmini && i < 30 }

// Locked: modifiche non ammesse dalla radio (canali fissi)
func (f uv17Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if !f.lockedSlot(m, i) {
		return nil
	}
	if old == nil || want == nil || old.Rx != want.Rx || old.Tx != want.Tx || old.TxOff != want.TxOff {
		return fmt.Errorf("i canali 1-30 della UV-5G Mini sono canali GMRS fissi: frequenze non modificabili")
	}
	if i+1 >= 8 && i+1 <= 14 && (old.Narrow != want.Narrow || old.Power != want.Power) {
		return fmt.Errorf("i canali GMRS 8-14 della UV-5G Mini sono sempre a banda stretta e bassa potenza")
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type uv17Sess struct {
	ctx context.Context
	p   Port
	m   *genModel
}

const uv17Timeout = 1500 * time.Millisecond

func (s *uv17Sess) recv(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, uv17Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

// identificazione: _do_ident di CHIRP
func (s *uv17Sess) ident() error {
	o := uv17o(s.m)
	sawData := false
	for _, id := range o.idents {
		_ = s.p.Purge()
		if err := s.p.Write(id); err != nil {
			return err
		}
		ack, err := s.recv(1)
		if err == errCancelled {
			return err
		}
		if err != nil {
			continue
		}
		sawData = true
		if ack[0] != 0x06 {
			continue
		}
		for _, mg := range o.magics {
			if err := s.p.Write(mg.cmd); err != nil {
				return err
			}
			if mg.cmd[0] == 'M' && o.modelIDs != nil {
				// come il programma del costruttore: dopo una breve pausa si leggono
				// tutti i byte arrivati e si confronta il nome del modello
				got := string(bytes.TrimRight(readUpTo(s.ctx, s.p, 32, 400*time.Millisecond), "\x00\xff "))
				if s.ctx.Err() != nil {
					return errCancelled
				}
				if got == "" {
					return rErr("la radio non risponde")
				}
				ok := false
				for _, id := range o.modelIDs {
					ok = ok || got == id
				}
				if !ok {
					return rErr("la radio si è identificata come «%s», non come %s: lettura e scrittura non eseguite", got, s.m.Name)
				}
				continue
			}
			if _, err := s.recv(mg.resp); err != nil {
				return err
			}
		}
		return nil
	}
	if sawData {
		return rErr("risposta inattesa dalla radio (è il modello giusto?)")
	}
	return rErr("la radio non risponde")
}

func (s *uv17Sess) key() []byte { return []byte(btSymbols[uv17o(s.m).encrsym]) }

func uv17Open(ctx context.Context, m *genModel, port string) (*uv17Sess, error) {
	p, err := openPortFunc(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &uv17Sess{ctx: ctx, p: p, m: m}, nil
}

func (f uv17Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := uv17o(m)
	s, err := uv17Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.total, "Collegamento con la radio…")
	if err := s.ident(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for k, start := range o.starts {
		for a := start; a < start+o.sizes[k]; a += 0x40 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			prog(len(data), o.total, "Lettura…")
			if err := s.p.Write([]byte{'R', byte(a >> 8), byte(a), 0x40}); err != nil {
				return nil, "", err
			}
			d, err := s.recv(0x44)
			if err != nil {
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			blk := append([]byte{}, d[4:]...)
			if o.encr {
				btCrypt(blk, s.key())
			}
			data = append(data, blk...)
		}
	}
	prog(o.total, o.total, "Lettura completata")
	return data, "", nil
}

func (f uv17Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := uv17o(m)
	if len(img) < o.total {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.total)
	}
	s, err := uv17Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.total, "Collegamento con la radio…")
	if err := s.ident(); err != nil {
		return "", genFail("Identificazione", err)
	}
	pos := 0
	for k, start := range o.starts {
		for a := start; a < start+o.sizes[k]; a += 0x40 {
			prog(pos, o.total, "Scrittura…")
			blk := append([]byte{}, img[pos:pos+0x40]...)
			if o.encr {
				btCrypt(blk, s.key())
			}
			pos += 0x40
			if err := s.p.Write(append([]byte{'W', byte(a >> 8), byte(a), 0x40}, blk...)); err != nil {
				return "", err
			}
			ack, err := s.recv(1)
			if err != nil {
				return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
			if ack[0] != 0x06 {
				return "", rErr("Scrittura (0x%04X): la radio ha rifiutato il blocco", a)
			}
		}
	}
	prog(o.total, o.total, "Scrittura completata")
	return "", nil
}
