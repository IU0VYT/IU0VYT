// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: baofeng_common.py, baofeng_wp970i.py, uv6r.py, uv5x3.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia Baofeng "BaofengCommonHT" (driver baofeng_common.py e
// baofeng_wp970i.py di CHIRP): UV-9R, UV-9R Pro, BF-A58, BF-A58S, UV-82III,
// UV-82WP, GT-3WP, UV-S9X3, 5RX, UV-9G, WP970I. 128 canali, 9600 baud.
//
// Protocollo: come la famiglia UV-5R, ma la stringa di identificazione è
// inviata in un'unica scrittura, la radio risponde con 8 byte (AA … DD) e la
// memoria (0x0000-0x1FFF) è seguita nell'immagine dall'identificativo.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const wp970iFormat = `
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

#seekto 0x1000;
struct {
  char name[7];
  u8 unknown1[9];
} names[128];
`

var (
	mstringWP970I = []byte{0x50, 0xBB, 0xFF, 0x20, 0x14, 0x04, 0x13}
	mstringUV9G   = []byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x05, 0x25}
	mstringUVS9X3 = []byte{0x50, 0xBB, 0xFF, 0x20, 0x12, 0x07, 0x25}
)

type bfcOpt struct {
	format    string // mappa dei canali (diversa sulla UV-6R)
	noAck     bool   // la radio non conferma prima di inviare un blocco (UV-6R)
	ranges    [][2]int
	uv6r      bool // riconoscimento del file come in uv6r.py
	identN    int  // lunghezza della risposta di identificazione (8, 14 sulla UV-5X3)
	btech     bool // marca BTECH in CHIRP: area ausiliaria completa solo con il "byte perso"
	fileid5x3 bool
	magics    [][]byte
	triBand   bool // potenza codificata come 0 (alta) / 2 (bassa)
	fileid    []string
	match     bool
	dtcs      []int
	scodeN    int // elementi della lista S-CODE
	pttidN    int
}

type bfcFam struct{}

func init() {
	cs := charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./"
	pw3 := []genPower{{"Alta 5 W", 5, 0}, {"Media 3 W", 3, 1}, {"Bassa 1 W", 1, 2}}
	pw2 := []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}}
	std := []band{{130000000, 180000000}, {400000000, 521000000}}
	mk := func(id, model, rclass, name string, o bfcOpt, nameLen int, pw []genPower, bands []band) *genModel {
		oo := o
		if oo.magics == nil {
			oo.magics = [][]byte{mstringWP970I}
		}
		if oo.format == "" {
			oo.format = wp970iFormat
		}
		if oo.identN == 0 {
			oo.identN = 8
		}
		if oo.ranges == nil {
			oo.ranges = [][2]int{{0x0000, 0x0DF0}, {0x0E00, 0x1800}, {0x1EE0, 0x1EF0}, {0x1F60, 0x1F70}, {0x1F80, 0x1F90}, {0x1FC0, 0x1FD0}}
		}
		oo.dtcs, oo.scodeN, oo.pttidN = uv5rDTCS, 15, 4
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: 128, First: 0,
			NameMax: nameLen, Charset: cs, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600,
			Fam: bfcFam{}, Opt: &oo, Notes: "Accendi la radio con il volume a metà e su un canale libero."}
	}
	a58s := []band{{130000000, 180000000}, {200000000, 260000000}, {400000000, 521000000}}
	b5rx := []band{{108000000, 136000000}, {136000000, 174000000}, {200000000, 260000000}, {350000000, 390000000}, {400000000, 520000000}}
	genRegister(
		mk("bf-a58", "BF-A58", "BFA58", "Baofeng BF-A58 / UV-9R Pro", bfcOpt{match: true, fileid: []string{"BFT515 ", "BFT517 "}}, 7, pw3, std),
		mk("bf-uv82wp", "UV-82WP", "UV82WP", "Baofeng UV-82WP", bfcOpt{match: true}, 6, pw3, std),
		mk("bf-gt3wp", "GT-3WP", "GT3WP", "Baofeng GT-3WP", bfcOpt{match: true}, 7, pw3, std),
		mk("bf-a58s", "BF-A58S", "BFA58S", "Baofeng BF-A58S / UV-82III", bfcOpt{match: true, triBand: true}, 7, pw2, a58s),
		mk("bf-uvs9x3", "UV-S9X3", "UVS9X3", "Baofeng UV-S9X3", bfcOpt{match: true, triBand: true, magics: [][]byte{mstringUVS9X3}}, 7, pw2, a58s),
		mk("bf-5rx", "5RX", "BF5RXRadio", "Baofeng UV-5RX (5RX)", bfcOpt{match: true, triBand: true, magics: [][]byte{mstringUVS9X3}}, 7, pw2, b5rx),
		mk("bf-uv9r", "UV-9R", "UV9R", "Baofeng UV-9R / UV-9R Plus", bfcOpt{match: true}, 7, pw3, std),
		mk("bf-uv9g", "UV-9G", "UV9G", "Baofeng UV-9G", bfcOpt{magics: [][]byte{mstringUV9G}}, 7,
			[]genPower{{"Alta 5 W", 5, 0}, {"Media 1 W", 1, 1}, {"Bassa 0,5 W", 0.5, 2}}, std),
		mk("bf-uv6r", "UV-6R", "UV6R", "Baofeng UV-6R", bfcOpt{format: uv6rFormat, noAck: true, uv6r: true, match: true,
			magics: [][]byte{{0x50, 0xBB, 0xFF, 0x20, 0x14, 0x11, 0x22}},
			ranges: [][2]int{{0x0000, 0x1800}, {0x1F40, 0x1F50}, {0x1FC0, 0x1FD0}, {0x1FE0, 0x1FF0}}}, 7, pw2,
			[]band{{136000000, 174000000}, {400000000, 520000000}}),
	)
	x3 := mk("bf-uv5rx3-btech", "UV-5X3", "UV5X3", "Baofeng UV-5RX3 / BTECH UV-5X3", bfcOpt{format: uv6rFormat, identN: 14, btech: true,
		fileid5x3: true, match: true, magics: [][]byte{{0x50, 0x0D, 0x0C, 0x20, 0x16, 0x03, 0x28}},
		ranges: [][2]int{{0x0000, 0x0DF0}, {0x0E00, 0x1800}, {0x1E00, 0x1E70}, {0x1EE0, 0x1EF0}, {0x1F60, 0x1F70}, {0x1F80, 0x1F90}, {0x1FA0, 0x1FB0}, {0x1FE0, 0x2000}}},
		7, pw2, []band{{130000000, 180000000}, {220000000, 226000000}, {400000000, 521000000}})
	x3.Vendor = "BTECH"
	genRegister(x3)
}

func bfco(m *genModel) *bfcOpt { return m.Opt.(*bfcOpt) }

const uv6rFormat = `
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
} memory[128];

#seekto 0x1000;
struct {
  char name[7];
  u8 unknown[9];
} names[128];
`

func (bfcFam) mem(m *genModel) *bwEl { return bwMust(bfco(m).format) }

func (f bfcFam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f bfcFam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	r := f.mem(m)
	return r.F("memory").I(i), r.F("names").I(i)
}

func (f bfcFam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(m, i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}

func (f bfcFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(m, i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}

func (f bfcFam) Match(m *genModel, img []byte) bool {
	o := bfco(m)
	if !o.match || (len(img) != 0x2008 && len(img) != 0x2010 && !o.fileid5x3) {
		return false
	}
	if o.fileid5x3 {
		if len(img) != 0x200E {
			return false
		}
		rid := string(img[0x1EF0:0x1EF7])
		return rid == "UVVG302" || rid == "UVVG301" || rid == "UVVG306"
	}
	if o.uv6r {
		rid1 := string(img[0x1FF8:0x2000])
		return (rid1 == " BF230#1" || rid1 == " BF230#2") && string(img[0x1FD0:0x1FD5]) == m.Model
	}
	if len(img) > 0x2008 {
		return bytes.HasPrefix(img[0x2008:0x2010], []byte(m.Model))
	}
	rid := string(img[0x1EF0:0x1EF7])
	for _, id := range o.fileid {
		if rid == id {
			return true
		}
	}
	return false
}

func (f bfcFam) Get(m *genModel, img []byte, i int) *Chan {
	o := bfco(m)
	mem, nam := f.recs(m, i)
	if mem.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: mem.F("rxfreq").Int(img) * 10}
	txr := mem.F("txfreq").Raw(img)
	if bytes.Equal(txr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(txr, []byte{0, 0, 0, 0}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = mem.F("txfreq").Int(img) * 10
	}
	c.Name = pyName(nam.F("name").Raw(img), true)
	c.TxTone = uv5rToneDec(mem.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(mem.F("rxtone").Int(img), true)
	c.Scan = mem.F("scan").Bool(img)
	lp := int(mem.F("lowpower").Int(img))
	if o.triBand {
		if lp == 2 {
			lp = 1
		} else {
			lp = 0
		}
	} else if lp >= len(m.Powers) {
		lp = 0
	}
	c.Power = m.powerSem(lp)
	c.Narrow = !mem.F("wide").Bool(img)
	return c
}

func (f bfcFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := bfco(m)
	mem, nam := f.recs(m, i)
	if c == nil {
		mem.Fill(img, 0xFF)
		nam.Fill(img, 0xFF)
		return nil
	}
	wasEmpty := mem.Raw(img)[0] == 0xFF
	var bcl, pttid, scode int64
	if !wasEmpty {
		bcl, pttid, scode = mem.F("bcl").Int(img), mem.F("pttid").Int(img), mem.F("scode").Int(img)
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
	pi := m.powerIndex(c.Power)
	if o.triBand {
		if pi == 1 {
			pi = 2
		} else {
			pi = 0
		}
	}
	mem.F("lowpower").SetInt(img, int64(pi))
	// impostazioni del canale conservate (come CHIRP, che le rilegge e le riscrive)
	mem.F("bcl").SetInt(img, bcl)
	mem.F("pttid").SetInt(img, pttid)
	mem.F("scode").SetInt(img, scode)
	return nil
}

// ---------------------------------------------------------------- protocollo

type bfcSess struct {
	ctx context.Context
	p   Port
	m   *genModel
}

func (s *bfcSess) recv(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, 1500*time.Millisecond); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

func (s *bfcSess) sleep(d time.Duration) error {
	select {
	case <-s.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

func (s *bfcSess) doIdent(magic []byte) ([]byte, error) {
	_ = s.p.Purge()
	if err := s.p.Write(magic); err != nil {
		return nil, err
	}
	ack, err := s.recv(1)
	if err != nil {
		return nil, err
	}
	if ack[0] != 0x06 {
		return nil, rErr("risposta inattesa dalla radio")
	}
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	n := bfco(s.m).identN
	id, err := s.recv(n)
	if err != nil {
		return nil, err
	}
	if id[0] != 0xAA || id[n-1] != 0xDD {
		return nil, rErr("risposta inattesa dalla radio (% X)", id)
	}
	if err := s.p.Write([]byte{0x06}); err != nil {
		return nil, err
	}
	ack, err = s.recv(1)
	if err != nil {
		return nil, err
	}
	if ack[0] != 0x06 {
		return nil, rErr("la radio ha rifiutato la copia")
	}
	return id, nil
}

func (s *bfcSess) identRadio() ([]byte, error) {
	var last error
	for _, mg := range bfco(s.m).magics {
		id, err := s.doIdent(mg)
		if err == nil {
			return id, nil
		}
		if _, ok := err.(radioErr); !ok {
			return nil, err
		}
		last = err
		if err := s.sleep(2 * time.Second); err != nil {
			return nil, err
		}
	}
	if last == nil {
		last = rErr("la radio non risponde")
	}
	return nil, last
}

// lettura di un blocco (bfc._read_block / _recv con conferma)
func (s *bfcSess) readBlock(start, size int, first bool) ([]byte, error) {
	if err := s.p.Write([]byte{'S', byte(start >> 8), byte(start), byte(size)}); err != nil {
		return nil, err
	}
	if !first && !bfco(s.m).noAck {
		ack, err := s.recv(1)
		if err != nil {
			return nil, err
		}
		if ack[0] != 0x06 {
			return nil, rErr("la radio ha rifiutato di inviare il blocco 0x%04X", start)
		}
	}
	hdr, err := s.recv(4)
	if err != nil {
		return nil, err
	}
	if hdr[0] != 'X' || int(hdr[1])<<8|int(hdr[2]) != start || int(hdr[3]) != size {
		return nil, rErr("risposta sconosciuta dalla radio")
	}
	d, err := s.recv(size)
	if err != nil {
		return nil, err
	}
	if err := s.p.Write([]byte{0x06}); err != nil {
		return nil, err
	}
	if err := s.sleep(50 * time.Millisecond); err != nil {
		return nil, err
	}
	return d, nil
}

func bfcOpen(ctx context.Context, m *genModel, port string) (*bfcSess, error) {
	p, err := openPortFunc(port, 9600)
	if err != nil {
		return nil, err
	}
	return &bfcSess{ctx: ctx, p: p, m: m}, nil
}

func (f bfcFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := bfcOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	ident, err := s.identRadio()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	// versione del firmware e "byte perso"
	if _, err := s.readBlock(0x1E80, 0x40, true); err != nil {
		return nil, "", genFail("Lettura della versione", err)
	}
	b1, err := s.readBlock(0x1EC0, 0x40, false)
	if err != nil {
		return nil, "", genFail("Lettura della versione", err)
	}
	b2, err := s.readBlock(0x1FC0, 0x40, false)
	if err != nil {
		return nil, "", genFail("Lettura della versione", err)
	}
	info := fmt.Sprintf("firmware %q", string(bytes.TrimRight(b1[48:62], "\x00\xff ")))
	dropped := b2[15] == 0xFF
	end := 0x2000
	if dropped {
		end = 0x2000 - 0x40
	}
	var data []byte
	for a := 0; a < end; a += 0x40 {
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		prog(a, 0x2000, "Lettura…")
		d, err := s.readBlock(a, 0x40, false)
		if err != nil {
			return nil, info, genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, d...)
	}
	if dropped {
		for a := 0x2000 - 0x40; a < 0x2000; a += 0x10 {
			d, err := s.readBlock(a, 0x10, false)
			if err != nil {
				return nil, info, genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			data = append(data, d...)
		}
	}
	data = append(data, ident...)
	prog(0x2000, 0x2000, "Lettura completata")
	return data, info, nil
}

func (f bfcFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := bfcOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if _, err := s.identRadio(); err != nil {
		return "", genFail("Identificazione", err)
	}
	var blk [6][]byte
	for k, a := range []int{0x1E80, 0x1EC0, 0x1F00, 0x1F40, 0x1F80, 0x1FC0} {
		if blk[k], err = s.readBlock(a, 0x40, k == 0); err != nil {
			return "", genFail("Lettura dell'area ausiliaria", err)
		}
	}
	info := fmt.Sprintf("firmware %q", string(bytes.TrimRight(blk[1][48:62], "\x00\xff ")))
	r1 := append(append([]byte{}, blk[2]...), blk[3][0:32]...)
	r2, r3 := blk[3][48:64], blk[4][16:64]
	dropped := blk[5][15] == 0xFF
	matched := bytes.Equal(img[0x1F00:0x1F60], r1) && bytes.Equal(img[0x1F70:0x1F80], r2) && bytes.Equal(img[0x1F90:0x1FC0], r3)
	if dropped && !matched {
		return info, rErr("immagine non adatta a questa radio: leggi prima la radio, modifica i canali e scrivili nella stessa radio")
	}
	ranges := bfco(m).ranges
	if dropped || (matched && !bfco(m).btech) {
		ranges = [][2]int{{0x0000, 0x0DF0}, {0x0E00, 0x1800}, {0x1E80, 0x2000}}
	}
	for _, r := range ranges {
		for a := r[0]; a < r[1]; a += 0x10 {
			prog(a, 0x2000, "Scrittura…")
			if err := s.p.Write(append([]byte{'X', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)); err != nil {
				return info, err
			}
			if err := s.sleep(50 * time.Millisecond); err != nil {
				return info, err
			}
			ack, err := s.recv(1)
			if err != nil {
				return info, genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
			if ack[0] != 0x06 {
				return info, rErr("Scrittura (0x%04X): la radio ha rifiutato il blocco", a)
			}
		}
	}
	prog(0x2000, 0x2000, "Scrittura completata")
	return info, nil
}
