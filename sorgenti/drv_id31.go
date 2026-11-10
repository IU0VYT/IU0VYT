// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// id31.py, id51.py, id51plus.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: palmari Icom D-STAR ID-31A/E (UHF), ID-51A/E (VHF/UHF), ID-51 Plus
// («50° anniversario») e ID-51 Plus2: 500 canali con nomi di 16 caratteri, copia
// veloce (38400 baud). Il canale è analogico (FM larga o stretta, AM in banda
// aerea sulle ID-51) o digitale D-STAR (DV); dei canali DV si conservano i
// nominativi (UR, RPT1, RPT2) e il codice digitale, un canale DV nuovo nasce con
// «CQCQCQ» come in CHIRP.
//
// Differenze volute rispetto a CHIRP:
//   - una frequenza che non è multiplo di 5 o 6,25 kHz viene rifiutata (CHIRP la
//     arrotonda verso il basso);
//   - il salto «P» della scansione programmata resta com'è quando il canale
//     modificato resta escluso dalla scansione.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

type id31Fam struct{}

type id31Opt struct {
	clone  *icomOpt
	format string
	am     bool  // modo AM (banda aerea)
	noBank int64 // valore «nessun banco»
	plus   bool  // ID-51 Plus e Plus2: area 0x1AF40 usata (distingue i file dalla ID-51)
	meta   bool  // file riconosciuto solo dai metadati (ID-51 Plus2)
}

const id31Rec = `
struct {
  u24 freq;
  u16 offset;
  u16 rtone:6,
      ctone:6,
      unknown2:1,
      mode:3;
  u8 dtcs;
  u8 tune_step:4,
     dsql_mode:2,
     unknown5:2;
  u8 unknown4;
  u8 tmode:4,
     duplex:2,
     dtcs_polarity:2;
  char name[16];
  u8 digcode;
  u8 urcall[7];
  u8 rpt1call[7];
  u8 rpt2call[7];
} memory[500];
`

const id31Format = id31Rec + `
#seekto 0x69C0;
u8 used_flags[70];
#seekto 0x6A06;
u8 skip_flags[69];
#seekto 0x6A4B;
u8 pskp_flags[69];
#seekto 0x6AC0;
struct {
  u8 bank;
  u8 index;
} banks[500];
#seekto 0x6F50;
struct {
  char name[16];
} bank_names[26];
`

const id51Format = id31Rec + `
#seekto 0x6A40;
u8 used_flags[70];
#seekto 0x6A86;
u8 skip_flags[69];
#seekto 0x6ACB;
u8 pskp_flags[69];
#seekto 0x6B40;
struct {
  u8 bank;
  u8 index;
} banks[500];
#seekto 0x6FD0;
struct {
  char name[16];
} bank_names[26];
`

const id51PlusFormat = id31Rec + `
#seekto 0x6A40;
u8 used_flags[70];
#seekto 0x6A86;
u8 skip_flags[69];
#seekto 0x6ACB;
u8 pskp_flags[69];
#seekto 0x6B40;
struct {
  u8 unknown:3,
     bank:5;
  u8 index;
} banks[500];
#seekto 0x6FD0;
struct {
  char name[16];
} bank_names[26];
`

// canale mai usato come lo lascia la radio (serve quando la memoria è vuota, tutta FF)
var id31Blank = append(append([]byte{0x01, 0x57, 0xC0, 0x03, 0xE8, 0x20, 0x80, 0x00, 0x80, 0xE4, 0x00},
	bytes.Repeat([]byte{' '}, 16)...), 0x00,
	0x87, 0x46, 0x1D, 0x18, 0x74, 0x50, 0x20, 0x40, 0x81, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x81, 0x02, 0x04, 0x08, 0x10, 0x20)

var (
	id31Clone      = &icomOpt{model: [4]byte{0x33, 0x22, 0x00, 0x01}, end: "Icom Inc.A8", size: 0x15500, fast: true, ranges: [][3]int{{0, 0x15500, 32}}}
	id51Clone      = &icomOpt{model: [4]byte{0x33, 0x90, 0x00, 0x01}, end: "Icom Inc.DA", size: 0x1FB40, fast: true, ranges: [][3]int{{0, 0x1FB40, 32}}}
	id51PlusClone  = &icomOpt{model: [4]byte{0x33, 0x90, 0x00, 0x02}, end: "Icom Inc.DA", size: 0x1FB40, fast: true, ranges: [][3]int{{0, 0x1FB40, 32}}}
	id51Plus2Clone = &icomOpt{model: [4]byte{0x33, 0x90, 0x00, 0x03}, end: "Icom Inc.DA", size: 0x1FB40, fast: true, raw: true, flip: true,
		ranges: [][3]int{{0, 0x1FB40, 32}}}
)

const id31Cable = "Cavo dati Icom OPC-2350LU (USB) oppure OPC-2218LU, nella presa «DATA»; la radio va accesa normalmente e durante la copia " +
	"mostra «CLONE». Copia veloce a 38400 baud. "

func init() {
	mk := func(id, model, rclass string, o *id31Opt, bands []band, what string) *genModel {
		return &genModel{ID: id, Vendor: "Icom", Model: model, RClass: rclass, Name: "Icom " + model,
			Slots: 500, First: 0, NameMax: 16, Powers: []genPower{{"Unica", 5, 0}}, Bands: bands, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: dtcsCodes, Baud: 9600, Fam: id31Fam{}, Opt: o, Digital: "D-STAR",
			Notes: "500 canali " + what + " (da 0 a 499) con nomi di 16 caratteri, analogici o D-STAR. " + id31Cable +
				"Un canale D-STAR nuovo nasce con «CQCQCQ». Driver di CHIRP."}
	}
	air := band{108000000, 137000000}
	genRegister(
		mk("ic-id31", "ID-31A", "ID31Radio", &id31Opt{clone: id31Clone, format: id31Format, noBank: 0xFF},
			[]band{{400000000, 479000000}}, "UHF"),
		mk("ic-id51", "ID-51", "ID51Radio", &id31Opt{clone: id51Clone, format: id51Format, am: true, noBank: 0xFF},
			[]band{{137000000, 174000000}, {400000000, 479000000}, air}, "VHF/UHF (banda aerea in AM)"),
		mk("ic-id51plus", "ID-51 Plus", "ID51PLUSRadio", &id31Opt{clone: id51PlusClone, format: id51PlusFormat, am: true, noBank: 0x1F, plus: true},
			[]band{{137000000, 174000000}, {380000000, 479000000}, air}, "VHF/UHF (banda aerea in AM)"),
		mk("ic-id51plus2", "ID-51 Plus2", "ID51PLUS2Radio", &id31Opt{clone: id51Plus2Clone, format: id51PlusFormat, am: true, noBank: 0x1F, plus: true, meta: true},
			[]band{{137000000, 174000000}, {380000000, 479000000}, air}, "VHF/UHF (banda aerea in AM)"),
	)
}

func (id31Fam) opt(m *genModel) *id31Opt        { return m.Opt.(*id31Opt) }
func (f id31Fam) icom(m *genModel) *icomOpt     { return f.opt(m).clone }
func (f id31Fam) mem(m *genModel) *bwEl         { return bwMust(f.opt(m).format) }
func (f id31Fam) layout(m *genModel) *bwEl      { return f.mem(m) }
func (f id31Fam) rec(m *genModel, i int) *bwEl  { return f.mem(m).F("memory").I(i) }
func (f id31Fam) bank(m *genModel, i int) *bwEl { return f.mem(m).F("banks").I(i) }

func (f id31Fam) Rec(m *genModel, img []byte, i int) []byte {
	e := f.mem(m)
	out := append([]byte{}, f.rec(m, i).Raw(img)...)
	out = append(out, icomB(icomBit(e.F("skip_flags"), img, i)), icomB(icomBit(e.F("pskp_flags"), img, i)))
	return append(out, f.bank(m, i).Raw(img)...)
}

func (f id31Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	e := f.mem(m)
	n := f.rec(m, i).ByteLen()
	if len(raw) != n+4 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:n])
	icomSetBit(e.F("skip_flags"), img, i, raw[n] != 0)
	icomSetBit(e.F("pskp_flags"), img, i, raw[n+1] != 0)
	f.bank(m, i).SetRaw(img, raw[n+2:])
	icomSetBit(e.F("used_flags"), img, i, false)
}

// Match: ID-51 e ID-51 Plus hanno la memoria della stessa grandezza; nella ID-51
// l'area 0x1AF40 è vuota (FF). I file della ID-51 Plus2 si riconoscono dai metadati.
func (f id31Fam) Match(m *genModel, img []byte) bool {
	o := f.opt(m)
	if len(img) != o.clone.size || o.meta {
		return false
	}
	if o.clone.size != 0x1FB40 {
		return true
	}
	empty := bytes.Equal(img[0x1AF40:0x1AF60], bytes.Repeat([]byte{0xFF}, 32))
	return empty != o.plus
}

// modi dei toni delle radio D-STAR: 1 subtono in trasmissione, 2 e 3 subtono nei
// due versi, 4 e 5 DCS; 6 e 7 (squelch rovesciato) non si mostrano, come in CHIRP
func id31Tones(tmode, rtone, ctone, dtcs, pol int64) (tx, rx string) {
	switch tmode {
	case 1:
		return icomTones(1, rtone, ctone, dtcs, pol)
	case 2, 3:
		return icomTones(2, rtone, ctone, dtcs, pol)
	case 4, 5:
		return icomTones(3, rtone, ctone, dtcs, pol)
	}
	return "", ""
}

func (id31Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (id31Fam) fitBySet()                {}

func (id31Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 70000000 || d < -70000000
}

// nominativo D-STAR: 8 caratteri da 7 bit in 7 byte
func id31Call(s string) []byte {
	s = (s + "        ")[:8]
	var acc uint64
	for i := 0; i < 8; i++ {
		acc = acc<<7 | uint64(s[i]&0x7F)
	}
	out := make([]byte, 7)
	for i := 6; i >= 0; i-- {
		out[i] = byte(acc)
		acc >>= 8
	}
	return out
}

func (f id31Fam) Get(m *genModel, img []byte, i int) *Chan {
	e := f.mem(m)
	if icomBit(e.F("used_flags"), img, i) {
		return nil
	}
	r := f.rec(m, i)
	raw := r.F("freq").Int(img)
	mult, omult := int64(5000), int64(5000)
	if raw&0x200000 != 0 {
		mult = 6250
	}
	if raw&0x040000 != 0 {
		omult = 6250
	}
	c := &Chan{Rx: (raw & 0x3FFFF) * mult}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * omult
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	mode := r.F("mode").Int(img)
	c.Dig = mode == 5
	c.Narrow = mode == 1
	if !c.Dig {
		c.TxTone, c.RxTone = id31Tones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	}
	c.Scan = !icomBit(e.F("skip_flags"), img, i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f id31Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	e, r, o := f.mem(m), f.rec(m, i), f.opt(m)
	if c == nil {
		icomSetBit(e.F("used_flags"), img, i, true)
		f.bank(m, i).F("bank").SetInt(img, o.noBank)
		return nil
	}
	was := !icomBit(e.F("used_flags"), img, i)
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, r.ByteLen())) {
		was = false
		r.SetRaw(img, id31Blank)
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz)
	tmode, rtone, ctone, dtcs, pol := int64(0), int64(8), int64(8), int64(0), int64(0)
	if was {
		rtone, ctone, dtcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
	}
	if !c.Dig {
		var err error
		if tmode, rtone, ctone, dtcs, pol, err = icomTonesFor(c, rtone, ctone, dtcs); err != nil {
			return err
		}
		tmode = []int64{0, 1, 3, 5}[tmode]
	}
	mult, flag := int64(5000), int64(0)
	switch {
	case c.Rx%5000 == 0:
	case c.Rx%6250 == 0:
		mult, flag = 6250, 0x200000
	default:
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5 o di 6,25 kHz)", fmtMHz(c.Rx))
	}
	if c.Rx/mult > 0x3FFFF {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	// lo scostamento di un canale simplex resta quello che c'era (come in CHIRP)
	dup, offs := int64(0), int64(600000)
	if was {
		offs = r.F("offset").Int(img) * 5000
		if r.F("freq").Int(img)&0x040000 != 0 {
			offs = r.F("offset").Int(img) * 6250
		}
	}
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	omult := int64(5000)
	switch {
	case offs%5000 == 0:
	case offs%6250 == 0:
		omult, flag = 6250, flag|0x040000
	default:
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz (multipli di 5 o di 6,25 kHz)", offs)
	}
	if dup != 0 && offs >= 60000000 {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	oldMode, oldRx := r.F("mode").Int(img), (r.F("freq").Int(img)&0x3FFFF)*5000
	if r.F("freq").Int(img)&0x200000 != 0 {
		oldRx = (r.F("freq").Int(img) & 0x3FFFF) * 6250
	}
	mode := int64(0)
	switch {
	case c.Dig:
		mode = 5
	case o.am && (c.Rx >= 108000000 && c.Rx < 137000000 || was && oldMode == 3 && oldRx == c.Rx):
		mode = 3 // banda aerea: AM
	case c.Narrow:
		mode = 1
	}
	if !(was && oldMode == 5) && (c.Dig || oldMode == 5) {
		// canale D-STAR nuovo (o posizione vuota con i resti di un canale D-STAR):
		// chiamata generale, nessun ripetitore (come CHIRP)
		r.F("urcall").SetRaw(img, id31Call("CQCQCQ"))
		r.F("rpt1call").SetRaw(img, id31Call(""))
		r.F("rpt2call").SetRaw(img, id31Call(""))
		r.F("digcode").SetInt(img, 0)
	}
	r.F("freq").SetInt(img, c.Rx/mult|flag)
	r.F("offset").SetInt(img, offs/omult)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("tmode").SetInt(img, tmode)
	r.F("duplex").SetInt(img, dup)
	r.F("mode").SetInt(img, mode)
	if step := r.F("tune_step").Int(img); !was || step == 2 || step == 3 || step > 13 {
		r.F("tune_step").SetInt(img, 0) // passo di sintonia di un canale nuovo: 5 kHz (come CHIRP)
	}
	icomSetBit(e.F("used_flags"), img, i, false)
	skipped := was && icomBit(e.F("skip_flags"), img, i)
	icomSetBit(e.F("skip_flags"), img, i, !c.Scan)
	if c.Scan || !skipped {
		icomSetBit(e.F("pskp_flags"), img, i, false)
	}
	return nil
}

func (f id31Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f id31Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
