// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: uv5r.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia Baofeng UV-5R (UV-5R, UV-5RA/RE/Plus, UV-5X, BF-F8HP, UV-82,
// UV-82HP, UV-82X3, UV-5RX3, UV-6, F-11, GT-5R, UV-5G Pro): 128 canali,
// 9600 baud. Protocollo e mappa della memoria come nel driver uv5r.py di CHIRP.
//
// Protocollo:
//   identificazione: 7 byte "magici" inviati uno alla volta (10 ms tra l'uno e
//   l'altro) → 0x06; 0x02 → identificativo (8 byte, o 12 nelle UV-6 recenti,
//   fino al byte 0xDD); 0x06 → 0x06.
//   lettura: 'S' indirizzo(2, BE) lunghezza → 'X' indirizzo lunghezza + dati;
//   il PC conferma con 0x06 e la radio risponde 0x06 (letto prima del blocco successivo).
//   scrittura: 'X' indirizzo lunghezza(0x10) + 16 byte → 0x06.
// L'immagine (come in CHIRP) è: identificativo (8 byte) + memoria principale
// 0x0000-0x17FF + area ausiliaria 0x1EC0-0x1FFF (versione del firmware, limiti
// di banda, soglie dello squelch).

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"
)

const uv5rFormat = `
#seekto 0x0008;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unused1:3,
     isuhf:1,
     scode:4;
  u8 unknown1:7,
     txtoneicon:1;
  u8 mailicon:3,
     unknown2:3,
     lowpower:2;
  u8 unknown3:1,
     wide:1,
     unknown4:2,
     bcl:1,
     scan:1,
     pttid:2;
} memory[128];

#seekto 0x1008;
struct {
  char name[7];
  u8 unknown2[9];
} names[128];
`

const (
	uv5rCharset       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890!@#$%^&*()+-=[]:\";'<>?,./"
	charsetASCII      = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	charsetAlnum      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz 1234567890"
	charsetUpperNum   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890"
	uv5rVHF220Radio   = 0x02
	uv5rFwVerStart    = 0x1838
	uv5rFwVerStop     = 0x1846
	uv5rMainSize      = 0x1808 // identificativo + memoria principale
	uv5rFullSize      = 0x1948
	uv5rFullSizeModel = 0x1950
)

var (
	uv5rModelOrig = []byte{0x50, 0xBB, 0xFF, 0x01, 0x25, 0x98, 0x4D}
	uv5rModel291  = []byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x07, 0x25}
	uv5rModelF11  = []byte{0x50, 0xBB, 0xFF, 0x13, 0xA1, 0x11, 0xDD}
	uv5rModelUV82 = []byte{0x50, 0xBB, 0xFF, 0x20, 0x13, 0x01, 0x05}
	uv5rModelUV6  = []byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x08, 0x23}
	uv5rModelUV6O = []byte{0x50, 0xBB, 0xFF, 0x12, 0x03, 0x98, 0x4D}
	uv5rModelA58  = []byte{0x50, 0xBB, 0xFF, 0x20, 0x14, 0x04, 0x13}
	uv5rModelUV5G = []byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x06, 0x25}
	uv5rBlacklist = []struct {
		magic  []byte
		reason string
	}{
		{[]byte{0x50, 0x0D, 0x0C, 0x20, 0x16, 0x03, 0x28}, "la radio si identifica come BTECH UV-5X3 (Baofeng UV-5RX3)"},
		{[]byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x06, 0x25}, "la radio si identifica come Radioddity UV-5G"},
	}
	uv5rDTCS = withCodes(dtcsCodes, 645)
)

var (
	basetypeUV5R   = []string{"BFS", "BFB", "N5R-2", "N5R2", "N5RV", "BTS", "D5R2", "B5R2"}
	basetypeF11    = []string{"USA"}
	basetypeUV82   = []string{"US2S2", "B82S", "BF82", "N82-2", "N822"}
	basetypeUV6    = []string{"BF1", "UV6"}
	basetypeF8HP   = []string{"BFP3V3 F", "N5R-3", "N5R3", "F5R3", "BFT", "N5RV"}
	basetypeUV82HP = []string{"N82-3", "N823", "N5R2"}
	basetypeUV82X3 = []string{"HN5RV01"}
)

type uv5rOpt struct {
	idents    [][]byte
	basetype  []string
	vhf, uhf  band
	r220      band
	aux       bool
	triPower  bool
	neverOrig bool   // la radio non è mai una "originale" (firmware < BFB291)
	match     bool   // riconoscibile dal contenuto del file
	bands     []band // gamme fisse (se diverse da quelle calcolate)
	airAM     bool   // UV-5G Pro: AM in 108-136 MHz
	gt5r      bool   // GT-5R: trasmissione solo nelle bande radioamatoriali
}

type uv5rFam struct{}

var uv5rPow2 = []genPower{{"Alta 4 W", 4, 0}, {"Bassa 1 W", 1, 2}}
var uv5rPow3 = []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}}

func init() {
	std := uv5rOpt{vhf: band{130000000, 176000000}, uhf: band{400000000, 520000000}, r220: band{220000000, 260000000}, aux: true, match: true}
	mk := func(id, vendor, model, rclass, name string, o uv5rOpt, pw []genPower, cs string, first int) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: 128, First: first,
			NameMax: 7, Charset: cs, Powers: pw, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: uv5rFam{}, Opt: &oo,
			Notes: "Accendi la radio con il volume a metà e su un canale libero."}
	}
	o := std
	o.idents, o.basetype = [][]byte{uv5rModel291, uv5rModelOrig}, basetypeUV5R
	uv5r := mk("bf-uv5r", "Baofeng", "UV-5R", "BaofengUV5RGeneric", "Baofeng UV-5R", o, uv5rPow2, uv5rCharset, 0)

	o5g := o
	o5g.match = false
	o5g.airAM = true
	o5g.bands = []band{{108000000, 174000000}, {220000000, 260000000}, {350000000, 390000000}, {400000000, 520000000}}
	uv5gpro := mk("bf-uv5gpro", "Baofeng", "UV-5G Pro", "BaofengUV5GPro", "Baofeng UV-5G Pro", o5g, uv5rPow2, uv5rCharset, 0)

	of := std
	of.idents, of.basetype, of.neverOrig = [][]byte{uv5rModelF11}, basetypeF11, true
	f11 := mk("bf-f11", "Baofeng", "F-11", "BaofengF11Radio", "Baofeng F-11", of, uv5rPow2, uv5rCharset, 0)

	o82 := std
	o82.idents, o82.basetype, o82.neverOrig = [][]byte{uv5rModelUV82}, basetypeUV82, true
	o82.uhf = band{400000000, 521000000}
	uv82 := mk("bf-uv82", "Baofeng", "UV-82", "BaofengUV82Radio", "Baofeng UV-82 / UV-82L", o82, uv5rPow2, charsetASCII, 0)

	o82x3 := o82
	o82x3.basetype = basetypeUV82X3
	o82x3.bands = []band{{130000000, 176000000}, {200000000, 260000000}, {400000000, 521000000}}
	uv82x3 := mk("bf-uv82x3", "Radioddity", "UV-82X3", "Radioddity82X3Radio", "Baofeng UV-82X3", o82x3, uv5rPow2, charsetASCII, 0)

	o6 := std
	o6.idents, o6.basetype, o6.aux, o6.neverOrig = [][]byte{uv5rModelUV6, uv5rModelUV6O}, basetypeUV6, false, true
	uv6 := mk("bf-uv6", "Baofeng", "UV-6", "BaofengUV6Radio", "Baofeng UV-6 / UV-7", o6, uv5rPow2, uv5rCharset, 1)

	o8 := std
	o8.idents, o8.basetype, o8.triPower, o8.neverOrig = [][]byte{uv5rModel291, uv5rModelA58}, basetypeF8HP, true, true
	o8.vhf, o8.uhf = band{130000000, 180000000}, band{400000000, 521000000}
	f8hp := mk("bf-f8hp", "Baofeng", "BF-F8HP", "BaofengBFF8HPRadio", "Baofeng BF-F8HP / UV-5RTP / UV-5R 8W", o8, uv5rPow3, uv5rCharset, 0)

	o82hp := std
	o82hp.idents, o82hp.basetype, o82hp.triPower, o82hp.neverOrig = [][]byte{uv5rModelUV82}, basetypeUV82HP, true, true
	o82hp.vhf, o82hp.uhf = band{136000000, 175000000}, band{400000000, 521000000}
	uv82hp := mk("bf-uv82hp", "Baofeng", "UV-82HP", "BaofengUV82HPRadio", "Baofeng UV-82HP", o82hp, uv5rPow3, charsetAlnum+"!@#$%^&*()+-=[]:\";'<>?,./", 0)

	ox3 := o
	ox3.match = false
	ox3.bands = []band{{130000000, 176000000}, {200000000, 260000000}, {400000000, 520000000}}
	uv5rx3 := mk("bf-uv5rx3", "Radioddity", "UV-5RX3", "RadioddityUV5RX3Radio", "Baofeng UV-5RX3 (firmware UV-5R)", ox3, uv5rPow2, uv5rCharset, 0)

	ogt := o
	ogt.match, ogt.gt5r = false, true
	gt5r := mk("bf-gt5r", "Baofeng", "GT-5R", "RadioddityGT5RRadio", "Baofeng GT-5R", ogt, uv5rPow2, uv5rCharset, 0)

	o5gr := o
	o5gr.idents, o5gr.match = [][]byte{uv5rModelUV5G}, false
	uv5g := mk("bf-uv5g", "Radioddity", "UV-5G", "RadioddityUV5GRadio", "Baofeng UV-5G (GMRS)", o5gr, uv5rPow2, uv5rCharset, 0)

	genRegister(uv5r, uv5gpro, f11, uv82, uv82x3, uv6, f8hp, uv82hp, uv5rx3, gt5r, uv5g)
}

func uv5ro(m *genModel) *uv5rOpt { return m.Opt.(*uv5rOpt) }

func (uv5rFam) mem() *bwEl { return bwMust(uv5rFormat) }

func (f uv5rFam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	r := f.mem()
	return r.F("memory").I(i), r.F("names").I(i)
}

func (f uv5rFam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(m, i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}

func (f uv5rFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(m, i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}

// versione del firmware scritta nell'immagine
func uv5rImgVersion(img []byte) []byte {
	if len(img) < uv5rFwVerStop {
		return nil
	}
	return img[uv5rFwVerStart:uv5rFwVerStop]
}

func uv5rIsOrigTag(tag []byte) (bool, error) {
	if i := bytes.Index(tag, []byte("BFB")); i >= 0 {
		if i+6 > len(tag) {
			return false, fmt.Errorf("versione del firmware non leggibile")
		}
		v, err := strconv.Atoi(string(tag[i+3 : i+6]))
		if err != nil {
			return false, fmt.Errorf("versione del firmware non leggibile (%q)", tag)
		}
		return v < 291, nil
	}
	return false, nil
}

func (f uv5rFam) isOrig(m *genModel, img []byte) bool {
	if uv5ro(m).neverOrig {
		return false
	}
	o, _ := uv5rIsOrigTag(uv5rImgVersion(img))
	return o
}

// gamme di frequenza della radio (dipendono dall'immagine, come in CHIRP)
func (f uv5rFam) Bands(m *genModel, img []byte) []band {
	o := uv5ro(m)
	if o.bands != nil {
		return o.bands
	}
	if img != nil && len(img) > 3 && !f.isOrig(m, img) && img[3] == uv5rVHF220Radio {
		return []band{o.vhf, o.r220}
	}
	return []band{o.vhf, o.uhf}
}

func (f uv5rFam) Get(m *genModel, img []byte, i int) *Chan {
	mem, nam := f.recs(m, i)
	if mem.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{}
	rx := mem.F("rxfreq").Int(img)
	tx := mem.F("txfreq").Int(img)
	c.Rx = rx * 10
	txr := mem.F("txfreq").Raw(img)
	if bytes.Equal(txr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(txr, []byte{0, 0, 0, 0}) {
		c.TxOff = true
		c.Tx = c.Rx
	} else {
		c.Tx = tx * 10
	}
	c.Name = pyName(nam.F("name").Raw(img), true)
	c.TxTone = uv5rToneDec(mem.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(mem.F("rxtone").Int(img), true)
	c.Scan = mem.F("scan").Bool(img)
	lp := int(mem.F("lowpower").Int(img))
	if lp >= len(m.Powers) {
		lp = 0
	}
	c.Power = m.powerSem(lp)
	c.Narrow = !mem.F("wide").Bool(img)
	if uv5ro(m).gt5r && !uv5rGT5RTx(c.Rx) {
		c.TxOff, c.Tx = true, c.Rx
	}
	return c
}

func uv5rGT5RTx(hz int64) bool {
	return (hz >= 144000000 && hz < 148000000) || (hz >= 420000000 && hz < 450000000)
}

// toni: 0/0xFFFF spento; >= 0x0258 CTCSS in decimi di Hz; altrimenti DCS:
// indice+1 (normale) o indice+0x6A (invertito).
// (CHIRP usa per il TX la soglia > 0x69 e per l'RX >= 0x6A: è la stessa cosa)
func uv5rToneDec(v int64, rx bool) string {
	switch {
	case v == 0 || v == 0xFFFF:
		return ""
	case v >= 0x0258:
		return fmtTone(int(v))
	}
	var idx int64
	inv := false
	if v > 0x69 {
		idx, inv = v-0x6A, true
	} else {
		idx = v - 1
	}
	if idx < 0 || int(idx) >= len(uv5rDTCS) {
		return ""
	}
	return genTone{Mode: "DTCS", Code: uv5rDTCS[idx], Inv: inv}.String()
}

func uv5rToneEnc(t genTone) int64 {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10)
	case "DTCS":
		v := int64(idxInt(uv5rDTCS, t.Code) + 1)
		if t.Inv {
			v += 0x69
		}
		return v
	}
	return 0
}

func (f uv5rFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	mem, nam := f.recs(m, i)
	if c == nil {
		mem.Fill(img, 0xFF)
		nam.Fill(img, 0xFF)
		return nil
	}
	wasEmpty := mem.Raw(img)[0] == 0xFF
	var bcl, scode, pttid int64
	if !wasEmpty {
		bcl, scode, pttid = mem.F("bcl").Int(img), mem.F("scode").Int(img), mem.F("pttid").Int(img)
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	mem.Fill(img, 0)
	mem.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		mem.F("txfreq").Fill(img, 0xFF)
	} else {
		mem.F("txfreq").SetInt(img, c.Tx/10)
	}
	name := c.Name
	nf := nam.F("name")
	for k := 0; k < 7; k++ {
		if k < len(name) {
			img[nf.Off+k] = name[k]
		} else {
			img[nf.Off+k] = 0xFF
		}
	}
	mem.F("txtone").SetInt(img, uv5rToneEnc(tt))
	mem.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	mem.F("scan").SetBool(img, c.Scan)
	mem.F("wide").SetBool(img, !c.Narrow)
	mem.F("lowpower").SetInt(img, int64(m.powerIndex(c.Power)))
	if !wasEmpty {
		mem.F("bcl").SetInt(img, bcl)
		mem.F("scode").SetInt(img, scode)
		mem.F("pttid").SetInt(img, pttid)
	}
	return nil
}

func (f uv5rFam) Match(m *genModel, img []byte) bool {
	o := uv5ro(m)
	if !o.match {
		return false
	}
	switch len(img) {
	case uv5rFullSizeModel:
		return bytes.HasPrefix(img[0x1948:0x1950], []byte(m.Model))
	case uv5rFullSize:
		rid := img[uv5rFwVerStart:uv5rFwVerStop]
		for _, t := range o.basetype {
			if bytes.Contains(rid, []byte(t)) {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------- protocollo

type uv5rSess struct {
	ctx context.Context
	p   Port
	m   *genModel
}

func (s *uv5rSess) read(n int, timeout time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, timeout); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *uv5rSess) sleep(d time.Duration) error {
	select {
	case <-s.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// doIdent: _do_ident di CHIRP
func (s *uv5rSess) doIdent(magic []byte, secondAck bool) ([]byte, error) {
	for _, b := range magic {
		if err := s.p.Write([]byte{b}); err != nil {
			return nil, err
		}
		if err := s.sleep(10 * time.Millisecond); err != nil {
			return nil, err
		}
	}
	ack, err := s.read(1, time.Second)
	if err == errCancelled {
		return nil, err
	}
	if err != nil || ack[0] != 0x06 {
		return nil, rErr("la radio non risponde")
	}
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	var resp []byte
	for k := 1; k < 13; k++ {
		b, err := s.read(1, time.Second)
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			continue
		}
		resp = append(resp, b[0])
		if b[0] == 0xDD {
			break
		}
	}
	var ident []byte
	switch len(resp) {
	case 8:
		ident = resp
	case 12:
		ident = append([]byte{resp[0], resp[3], resp[5]}, resp[7:]...)
	default:
		return nil, rErr("risposta inattesa dalla radio (% X)", resp)
	}
	if secondAck {
		if err := s.p.Write([]byte{0x06}); err != nil {
			return nil, err
		}
		ack, err := s.read(1, time.Second)
		if err == errCancelled {
			return nil, err
		}
		if err != nil || ack[0] != 0x06 {
			return nil, rErr("la radio ha rifiutato la copia")
		}
	}
	return ident, nil
}

func (s *uv5rSess) identRadio() ([]byte, error) {
	var last error
	for _, magic := range uv5ro(s.m).idents {
		id, err := s.doIdent(magic, true)
		if err == nil {
			return id, nil
		}
		if err == errCancelled {
			return nil, err
		}
		if _, ok := err.(radioErr); !ok {
			return nil, err
		}
		last = err
		if err := s.sleep(2 * time.Second); err != nil {
			return nil, err
		}
	}
	for _, bl := range uv5rBlacklist {
		if _, err := s.doIdent(bl.magic, false); err != nil {
			if err == errCancelled {
				return nil, err
			}
			if _, ok := err.(radioErr); ok {
				continue
			}
			return nil, err
		}
		return nil, rErr("%s: scegli il modello giusto e riprova", bl.reason)
	}
	if last != nil {
		return nil, last
	}
	return nil, rErr("la radio non risponde")
}

func (s *uv5rSess) readBlock(start, size int, first bool) ([]byte, error) {
	if err := s.p.Write([]byte{'S', byte(start >> 8), byte(start), byte(size)}); err != nil {
		return nil, err
	}
	if !first {
		ack, err := s.read(1, time.Second)
		if err == errCancelled {
			return nil, err
		}
		if err != nil || ack[0] != 0x06 {
			return nil, rErr("la radio ha rifiutato di inviare il blocco 0x%04X", start)
		}
	}
	hdr, err := s.read(4, time.Second)
	if err == errCancelled {
		return nil, err
	}
	if err != nil {
		return nil, rErr("la radio non ha inviato il blocco 0x%04X", start)
	}
	if hdr[0] != 'X' || int(hdr[1])<<8|int(hdr[2]) != start || int(hdr[3]) != size {
		return nil, rErr("risposta sconosciuta dalla radio")
	}
	chunk, err := s.read(size, time.Second)
	if err == errCancelled {
		return nil, err
	}
	if err != nil {
		return nil, rErr("la radio ha inviato il blocco 0x%04X incompleto", start)
	}
	if err := s.p.Write([]byte{0x06}); err != nil {
		return nil, err
	}
	if err := s.sleep(50 * time.Millisecond); err != nil {
		return nil, err
	}
	return chunk, nil
}

func (s *uv5rSess) radioVersion() ([]byte, bool, error) {
	if _, err := s.readBlock(0x1E80, 0x40, true); err != nil {
		return nil, false, err
	}
	b1, err := s.readBlock(0x1EC0, 0x40, false)
	if err != nil {
		return nil, false, err
	}
	b2, err := s.readBlock(0x1FC0, 0x40, false)
	if err != nil {
		return nil, false, err
	}
	return b1[48:62], b2[15] == 0xFF, nil
}

func (s *uv5rSess) auxFromRadio() (version, a1, a2, a3 []byte, dropped bool, err error) {
	var blk [6][]byte
	for k, addr := range []int{0x1E80, 0x1EC0, 0x1F00, 0x1F40, 0x1F80, 0x1FC0} {
		if blk[k], err = s.readBlock(addr, 0x40, k == 0); err != nil {
			return
		}
	}
	version = blk[1][48:62]
	a1 = append(append([]byte{}, blk[2]...), blk[3][0:32]...)
	a2 = blk[3][48:64]
	a3 = blk[4][16:64]
	dropped = blk[5][15] == 0xFF
	return
}

func uv5rOpen(ctx context.Context, m *genModel, port string) (*uv5rSess, error) {
	p, err := openPortFunc(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &uv5rSess{ctx: ctx, p: p, m: m}, nil
}

func genFail(step string, err error) error {
	if err == errCancelled {
		return err
	}
	if err == errTimeout {
		return fmt.Errorf("%s: la radio non risponde. Controlla che sia accesa, che il cavo sia inserito bene e che la porta sia quella giusta", step)
	}
	if re, ok := err.(radioErr); ok {
		return fmt.Errorf("%s: %s", step, re.msg)
	}
	return fmt.Errorf("%s: %v", step, err)
}

func (f uv5rFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := uv5rOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	o := uv5ro(m)
	prog(0, 0x1800, "Collegamento con la radio…")
	data, err := s.identRadio()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	info := fmt.Sprintf("identificativo % X", data)
	dropped := false
	if !o.aux {
		if _, err := s.readBlock(0, 0x40, true); err != nil {
			return nil, info, genFail("Lettura", err)
		}
	} else {
		var ver []byte
		if ver, dropped, err = s.radioVersion(); err != nil {
			return nil, info, genFail("Lettura della versione", err)
		}
		info += fmt.Sprintf(", firmware %q", string(bytes.TrimRight(ver, "\x00\xff ")))
	}
	for a := 0; a < 0x1800; a += 0x40 {
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		prog(a, 0x1800, "Lettura…")
		b, err := s.readBlock(a, 0x40, false)
		if err != nil {
			return nil, info, genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, b...)
	}
	if o.aux {
		end := 0x2000
		if dropped {
			end = 0x1FC0
		}
		for a := 0x1EC0; a < end; a += 0x40 {
			b, err := s.readBlock(a, 0x40, false)
			if err != nil {
				return nil, info, genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			data = append(data, b...)
		}
		if dropped {
			for a := 0x1FC0; a < 0x2000; a += 0x10 {
				b, err := s.readBlock(a, 0x10, false)
				if err != nil {
					return nil, info, genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
				}
				data = append(data, b...)
			}
		}
	}
	prog(0x1800, 0x1800, "Lettura completata")
	return data, info, nil
}

func (s *uv5rSess) sendBlock(addr int, data []byte) error {
	msg := append([]byte{'X', byte(addr >> 8), byte(addr), byte(len(data))}, data...)
	if err := s.p.Write(msg); err != nil {
		return err
	}
	if err := s.sleep(50 * time.Millisecond); err != nil {
		return err
	}
	ack, err := s.read(1, time.Second)
	if err == errCancelled {
		return err
	}
	if err != nil || ack[0] != 0x06 {
		return rErr("la radio ha rifiutato il blocco 0x%04X", addr)
	}
	return nil
}

func (f uv5rFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := uv5ro(m)
	if len(img) < uv5rMainSize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := uv5rOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x1800, "Collegamento con la radio…")
	ident, err := s.identRadio()
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	info := fmt.Sprintf("identificativo % X", ident)
	radioUpper, imageUpper := ident[3], img[3]
	if (imageUpper == uv5rVHF220Radio || radioUpper == uv5rVHF220Radio) && imageUpper != radioUpper {
		return info, rErr("l'immagine non è adatta a questa radio (bande di frequenza diverse): leggi prima la radio e lavora sui canali letti")
	}
	type rng struct{ a, b int }
	mainR := []rng{{0x0008, 0x0CF8}, {0x0D08, 0x0DF8}, {0x0E08, 0x1808}}
	var auxR []rng
	if o.aux {
		if len(img) < uv5rFullSize {
			return info, fmt.Errorf("immagine senza l'area ausiliaria: leggi prima la radio")
		}
		imgOrig, err := uv5rIsOrigTag(uv5rImgVersion(img))
		if o.neverOrig {
			imgOrig, err = false, nil
		}
		if err != nil {
			return info, rErr("versione del firmware dell'immagine non leggibile")
		}
		rver, r1, r2, r3, dropped, err := s.auxFromRadio()
		if err != nil {
			return info, genFail("Lettura dell'area ausiliaria", err)
		}
		info += fmt.Sprintf(", firmware %q", string(bytes.TrimRight(rver, "\x00\xff ")))
		radioOrig := false
		if i := bytes.Index(rver, []byte("BFB")); i >= 0 && i+6 <= len(rver) {
			v, err := strconv.Atoi(string(rver[i+3 : i+6]))
			if err != nil {
				return info, rErr("versione del firmware della radio non leggibile")
			}
			radioOrig = v < 291
		}
		if imgOrig != radioOrig {
			return info, rErr("l'immagine non è adatta a questa radio (versione del firmware diversa): leggi prima la radio e lavora sui canali letti")
		}
		matched := bytes.Equal(img[0x1848:0x18A8], r1) && bytes.Equal(img[0x18B8:0x18C8], r2) && bytes.Equal(img[0x18D8:0x1908], r3)
		if dropped && !matched {
			return info, rErr("immagine non adatta a questa radio: leggi prima la radio, modifica i canali e scrivili nella stessa radio")
		}
		switch {
		case imgOrig:
			auxR = []rng{{0x1EE0, 0x1EF0}, {0x1FC0, 0x1FE0}}
		case dropped || matched:
			auxR = []rng{{0x1EC0, 0x2000}}
		default:
			auxR = []rng{{0x1EE0, 0x1EF0}, {0x1F60, 0x1F70}, {0x1F80, 0x1F90}, {0x1FC0, 0x1FD0}}
		}
	}
	total := 0
	for _, r := range mainR {
		total += r.b - r.a
	}
	done := 0
	for _, r := range mainR {
		for a := r.a; a < r.b; a += 0x10 {
			prog(done, total, "Scrittura…")
			if err := s.sendBlock(a-0x08, img[a:a+0x10]); err != nil {
				return info, genFail(fmt.Sprintf("Scrittura (0x%04X)", a-8), err)
			}
			done += 0x10
		}
	}
	if len(img) == uv5rMainSize || !o.aux {
		prog(total, total, "Scrittura completata")
		return info, nil
	}
	for _, r := range auxR {
		for a := r.a; a < r.b; a += 0x10 {
			off := 0x1808 + (a - 0x1EC0)
			if err := s.sendBlock(a, img[off:off+0x10]); err != nil {
				return info, genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
		}
	}
	prog(total, total, "Scrittura completata")
	return info, nil
}

func (f uv5rFam) layout(m *genModel) *bwEl { return f.mem() }
