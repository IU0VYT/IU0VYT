// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icv86.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: Icom IC-V86 (VHF, 200 canali con nomi di 5 caratteri).

import (
	"context"
	"fmt"
	"strings"
)

type icv86Fam struct{}

const icv86Format = `
#seekto 0x040;
u8 skips[26];
#seekto 0x060;
u8 used[26];
#seekto 0x0200;
struct {
  ul32 freq;
  ul32 offset;
  char name[5];
  u8 reserved1:2,
     rtone:6;
  u8 reserved2:2,
     ctone:6;
  u8 reserved3:1,
     dtcs:7;
  u8 reserved4:5,
     tuning_step:3;
  u8 reserved5:2,
     mode:1,
     rev:1,
     duplex:2,
     reserved6:2;
  u8 reserved7:2,
     dtcs_polarity:2,
     tmode:4;
  u8 reserved8:5,
     tx:1,
     power:2;
  u8 reserved9;
  u8 reserved10;
  u8 reserved11;
  u8 reserved12;
} memory[207];
`

var icv86Opt = &icomOpt{model: [4]byte{0x40, 0x66, 0x00, 0x01}, end: "Icom Inc.AC", size: 5504, ranges: [][3]int{{0, 5504, 32}}}

func init() {
	genRegister(&genModel{ID: "ic-v86", Vendor: "Icom", Model: "IC-V86", RClass: "ICV86Radio", Name: "Icom IC-V86",
		Slots: 200, First: 0, NameMax: 5, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890",
		Powers: []genPower{{"Alta 5,5 W", 5.5, 0}, {"Bassa 0,5 W", 0.5, 2}, {"Media 2,5 W", 2.5, 1}, {"Massima 7 W", 7, -1}},
		Bands:  []band{{136000000, 174000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: icv86Fam{},
		Notes: "200 canali VHF (da 0 a 199) con nomi di 5 caratteri. " + icomCableNote + "Passo di sintonia, canali di scansione e di chiamata, " +
			"inversione e permesso di trasmissione del canale si conservano. Driver di CHIRP."})
}

func (icv86Fam) icom(m *genModel) *icomOpt       { return icv86Opt }
func (icv86Fam) mem() *bwEl                      { return bwMust(icv86Format) }
func (f icv86Fam) layout(m *genModel) *bwEl      { return f.mem() }
func (f icv86Fam) rec(i int) *bwEl               { return f.mem().F("memory").I(i) }
func (f icv86Fam) flag(what string, i int) *bwEl { return f.mem().F(what).I(i / 8) }

func (f icv86Fam) bit(img []byte, what string, i int) bool {
	return f.flag(what, i).Int(img)&(1<<(uint(i)%8)) != 0
}

func (f icv86Fam) setBit(img []byte, what string, i int, on bool) {
	e := f.flag(what, i)
	v := e.Int(img)
	if on {
		v |= 1 << (uint(i) % 8)
	} else {
		v &^= 1 << (uint(i) % 8)
	}
	e.SetInt(img, v)
}

func (f icv86Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(i).Raw(img)...)
	b := byte(0)
	if f.bit(img, "skips", i) {
		b = 1
	}
	return append(out, b)
}

func (f icv86Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+1 {
		f.rec(i).SetRaw(img, raw[:n])
		f.setBit(img, "skips", i, raw[n] != 0)
		f.setBit(img, "used", i, false)
	}
}

func (icv86Fam) Match(m *genModel, img []byte) bool { return len(img) == icv86Opt.size }

// toni come nei driver Icom di CHIRP: modo 0 spento, 1 subtono in trasmissione,
// 2 subtono in entrambi i versi, 3 DCS (con le polarità), 4 DCS rovesciato
func icomTones(tmode, rtone, ctone, dtcs, pol int64) (tx, rx string) {
	tone := func(k int64) string {
		if k >= 0 && int(k) < len(chirpTones) {
			return fmtTone(chirpTones[k])
		}
		return "88.5"
	}
	switch tmode {
	case 1:
		return tone(rtone), ""
	case 2:
		return tone(ctone), tone(ctone)
	case 3:
		code := 23
		if dtcs >= 0 && int(dtcs) < len(dtcsCodes) {
			code = dtcsCodes[dtcs]
		}
		d := func(inv bool) string {
			if inv {
				return fmt.Sprintf("D%03dI", code)
			}
			return fmt.Sprintf("D%03dN", code)
		}
		return d(pol&2 != 0), d(pol&1 != 0)
	}
	return "", ""
}

// icomToneMode: modo dei toni per la coppia (trasmissione, ricezione); -1 se la
// radio non ha quella combinazione
func icomToneMode(t, r genTone) int64 {
	switch {
	case t.Mode == "" && r.Mode == "":
		return 0
	case t.Mode == "Tone" && r.Mode == "":
		return 1
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
		return 2
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
		return 3
	}
	return -1
}

func icomIndex(list []int, v int) int64 {
	for k, x := range list {
		if x == v {
			return int64(k)
		}
	}
	return -1
}

// toneOK (4.14): combinazioni di toni che le radio Icom hanno (vedi genAdapt)
func (icv86Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (ict10Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (icv80Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (ict70Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (icx8xFam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }

// fitBySet (4.14): i canali che la radio non può memorizzare si scartano prima
// della scrittura (vedi genAdapt)
func (ft4Fam) fitBySet()   {}
func (ft60Fam) fitBySet()  {}
func (icv86Fam) fitBySet() {}
func (ict10Fam) fitBySet() {}
func (icv80Fam) fitBySet() {}
func (ict70Fam) fitBySet() {}
func (icx8xFam) fitBySet() {}

func (icv86Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 70000000 || d < -70000000
}

func (f icv86Fam) Get(m *genModel, img []byte, i int) *Chan {
	if f.bit(img, "used", i) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("freq").Int(img)}
	c.Tx = c.Rx
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - r.F("offset").Int(img)
	case 2:
		c.Tx = c.Rx + r.F("offset").Int(img)
	}
	c.TxTone, c.RxTone = icomTones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("mode").Bool(img)
	c.Scan = !f.bit(img, "skips", i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f icv86Fam) fill(img []byte, r *bwEl) {
	r.Fill(img, 0)
	r.F("freq").SetInt(img, 146010000)
	r.F("offset").SetInt(img, 146010000)
	r.F("name").SetStr(img, "", ' ')
	r.F("rtone").SetInt(img, 8)
	r.F("ctone").SetInt(img, 8)
	r.F("tx").SetInt(img, 1)
}

func (f icv86Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		f.fill(img, r)
		f.setBit(img, "used", i, true)
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := icomToneMode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
	}
	was := !f.bit(img, "used", i)
	kept, rtone, ctone, dtcs := int64(600000), int64(8), int64(8), int64(0) // valori di CHIRP per un canale nuovo
	if was {
		kept, rtone, ctone, dtcs = r.F("offset").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
	}
	pol := int64(0)
	switch tmode {
	case 1:
		rtone = icomIndex(chirpTones, t.Hz10)
	case 2:
		rtone = icomIndex(chirpTones, t.Hz10)
		ctone = rtone
	case 3:
		dtcs = icomIndex(dtcsCodes, t.Code)
		if t.Inv {
			pol |= 2
		}
		if rt.Inv {
			pol |= 1
		}
	}
	if rtone < 0 || ctone < 0 || dtcs < 0 {
		return fmt.Errorf("tono non disponibile su questa radio")
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if !was {
		f.fill(img, r)
	}
	r.F("freq").SetInt(img, c.Rx)
	r.F("offset").SetInt(img, offs)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("mode").SetBool(img, c.Narrow)
	r.F("duplex").SetInt(img, dup)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("tmode").SetInt(img, tmode)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	f.setBit(img, "used", i, false)
	f.setBit(img, "skips", i, !c.Scan)
	return nil
}

func (f icv86Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icv86Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}

// ---------------------------------------------------------------- IC-T10

// 4.14: Icom IC-T10 (VHF/UHF, 200 canali con nomi di 6 caratteri): come la
// IC-V86, con le aree dei bit in un altro punto e il canale di 24 byte.
type ict10Fam struct{}

const ict10Format = `
#seekto 0x020;
lbit pskips[208];
#seekto 0x040;
lbit skips[208];
#seekto 0x060;
lbit used[208];
#seekto 0x0240;
struct {
  ul32 freq;
  ul32 offset;
  char name[6];
  u8 reserved1:2,
     rtone:6;
  u8 reserved2:2,
     ctone:6;
  u8 reserved3:1,
     dtcs:7;
  u8 reserved4:4,
     tuning_step:4;
  u8 reserved5:2,
     mode:1,
     rev:1,
     duplex:2,
     reserved6:2;
  u8 reserved7:2,
     dtcs_polarity:2,
     tmode:4;
  u8 reserved8:5,
     tx_inhibit:1,
     power:2;
  u8 reserved9;
  u8 reserved10;
  u8 reserved11;
} memory[208];
`

var ict10Opt = &icomOpt{model: [4]byte{0x43, 0x28, 0x00, 0x01}, end: "Icom Inc.B8", size: 5888, ranges: [][3]int{{0, 5888, 32}}}

func init() {
	genRegister(&genModel{ID: "ic-t10", Vendor: "Icom", Model: "IC-T10", RClass: "ICVT10Radio", Name: "Icom IC-T10",
		Slots: 200, First: 0, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}, {"Media 2,5 W", 2.5, 1}},
		Bands:  []band{{136000000, 174000000}, {400000000, 479000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ict10Fam{},
		Notes: "200 canali VHF/UHF (da 0 a 199) con nomi di 6 caratteri. " + icomCableNote + "Passo di sintonia, canali di scansione e di chiamata, " +
			"inversione e blocco della trasmissione del canale si conservano. Driver di CHIRP."})
}

func (ict10Fam) icom(m *genModel) *icomOpt      { return ict10Opt }
func (ict10Fam) mem() *bwEl                     { return bwMust(ict10Format) }
func (f ict10Fam) layout(m *genModel) *bwEl     { return f.mem() }
func (f ict10Fam) rec(i int) *bwEl              { return f.mem().F("memory").I(i) }
func (f ict10Fam) bit(what string, i int) *bwEl { return f.mem().F(what).I(i) }

func (f ict10Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), icomB(f.bit("skips", i).Bool(img)), icomB(f.bit("pskips", i).Bool(img)))
}

func (f ict10Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+2 {
		f.rec(i).SetRaw(img, raw[:n])
		f.bit("skips", i).SetBool(img, raw[n] != 0)
		f.bit("pskips", i).SetBool(img, raw[n+1] != 0)
		f.bit("used", i).SetBool(img, false)
	}
}

func (ict10Fam) Match(m *genModel, img []byte) bool { return len(img) == ict10Opt.size }

func (ict10Fam) chirpInvalid(m *genModel, c *Chan) bool { return icv86Fam{}.chirpInvalid(m, c) }

func (f ict10Fam) Get(m *genModel, img []byte, i int) *Chan {
	if f.bit("used", i).Bool(img) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("freq").Int(img)}
	c.Tx = c.Rx
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - r.F("offset").Int(img)
	case 2:
		c.Tx = c.Rx + r.F("offset").Int(img)
	}
	c.TxTone, c.RxTone = icomTones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("mode").Bool(img)
	c.Scan = f.bit("pskips", i).Bool(img) || !f.bit("skips", i).Bool(img)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f ict10Fam) fill(img []byte, r *bwEl) {
	r.Fill(img, 0)
	r.F("freq").SetInt(img, 146010000)
	r.F("offset").SetInt(img, 146010000)
	r.F("name").SetStr(img, "", ' ')
	r.F("rtone").SetInt(img, 8)
	r.F("ctone").SetInt(img, 8)
}

func (f ict10Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		f.fill(img, r)
		f.bit("used", i).SetBool(img, true)
		return nil
	}
	was := !f.bit("used", i).Bool(img)
	kept, rtone, ctone, dtcs := int64(600000), int64(8), int64(8), int64(0)
	if was {
		kept, rtone, ctone, dtcs = r.F("offset").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
	}
	tmode, rtone, ctone, dtcs, pol, err := icomTonesFor(c, rtone, ctone, dtcs)
	if err != nil {
		return err
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if !was {
		f.fill(img, r)
	}
	r.F("freq").SetInt(img, c.Rx)
	r.F("offset").SetInt(img, offs)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("mode").SetBool(img, c.Narrow)
	r.F("duplex").SetInt(img, dup)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("tmode").SetInt(img, tmode)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	f.bit("used", i).SetBool(img, false)
	f.bit("skips", i).SetBool(img, !c.Scan)
	f.bit("pskips", i).SetBool(img, false)
	return nil
}

func (f ict10Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f ict10Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
