// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icv80.py, ict70.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: Icom IC-V80 (VHF, 200 canali) e IC-T70 (VHF/UHF, 300 canali).
// Le frequenze sono in unità da 5 kHz o da 6,25 kHz.
//
// Differenza voluta rispetto a CHIRP (controllata nelle prove): una frequenza
// multipla di 6,25 kHz ma non di 12,5 kHz si scrive in unità da 6,25 kHz invece
// di essere arrotondata; frequenze e scostamenti non memorizzabili si rifiutano.

import (
	"context"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------- IC-V80

type icv80Fam struct{}

const icv80Format = `
struct {
  ul16 freq;
  ul16 offset;
  char name[5];
  u8 unknown0:2,
     rtone:6;
  u8 unknown1:2,
     ctone:6;
  u8 unknown2:1,
     dtcs:7;
  u8 unknown3:5,
     tuning_step:3;
  u8 unknown4:2,
     mode:1,
     reverse_duplex:1,
     duplex:2,
     mult:2;
  u8 unknown6:2,
     dtcs_polarity:2,
     unknown8:2,
     tmode:2;
  u8 unknown9:5,
     tx_inhibit:1,
     power:2;
} memory[207];
u8 skip[32];
u8 unused[32];
`

var icv80Opt = &icomOpt{model: [4]byte{0x32, 0x54, 0x00, 0x01}, end: "Icom Inc.7C", size: 0x0E80, fast: true,
	ranges: [][3]int{{0x0000, 0x0CE0, 32}, {0x0CE0, 0x0D40, 16}, {0x0D40, 0x0E00, 32}, {0x0E00, 0x0E20, 16}, {0x0E20, 0x0E60, 32},
		{0x0E60, 0x0E70, 16}, {0x0E70, 0x0E72, 2}, {0x0E72, 0x0E77, 5}, {0x0E77, 0x0E78, 1}, {0x0E78, 0x0E80, 8}}}

var icomPowersV = []genPower{{"Alta 5,5 W", 5.5, 0}, {"Bassa 0,5 W", 0.5, 2}, {"Media 2,5 W", 2.5, 1}}

func init() {
	genRegister(&genModel{ID: "ic-v80", Vendor: "Icom", Model: "IC-V80", RClass: "ICV80Radio", Name: "Icom IC-V80",
		Slots: 200, First: 0, NameMax: 5, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890", Powers: icomPowersV,
		Bands: []band{{136000000, 174000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600, Fam: icv80Fam{},
		Notes: "200 canali VHF (da 0 a 199) con nomi di 5 caratteri. " + icomCableNote + "La copia è quella veloce (38400 baud dopo l'identificazione). " +
			"Passo di sintonia, canali di scansione e di chiamata e inversione del canale si conservano. Driver di CHIRP."})
	genRegister(&genModel{ID: "ic-t70", Vendor: "Icom", Model: "IC-T70", RClass: "ICT70Radio", Name: "Icom IC-T70",
		Slots: 300, First: 0, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}, {"Media 1 W", 1, 1}},
		Bands:  []band{{136000000, 174000000}, {400000000, 479000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ict70Fam{},
		Notes: "300 canali VHF/UHF (da 0 a 299) con nomi di 6 caratteri. " + icomCableNote + "I banchi si conservano. Driver di CHIRP."})
}

func (icv80Fam) icom(m *genModel) *icomOpt  { return icv80Opt }
func (icv80Fam) mem() *bwEl                 { return bwMust(icv80Format) }
func (f icv80Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f icv80Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

// bit i di un'area di bit (un byte ogni 8 canali)
func icomBit(e *bwEl, img []byte, i int) bool { return e.I(i/8).Int(img)&(1<<(uint(i)%8)) != 0 }

func icomSetBit(e *bwEl, img []byte, i int, on bool) {
	x := e.I(i / 8)
	v := x.Int(img)
	if on {
		v |= 1 << (uint(i) % 8)
	} else {
		v &^= 1 << (uint(i) % 8)
	}
	x.SetInt(img, v)
}

func icomB(on bool) byte {
	if on {
		return 1
	}
	return 0
}

func (f icv80Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), icomB(icomBit(f.mem().F("skip"), img, i)))
}

func (f icv80Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+1 {
		f.rec(i).SetRaw(img, raw[:n])
		icomSetBit(f.mem().F("skip"), img, i, raw[n] != 0)
		icomSetBit(f.mem().F("unused"), img, i, false)
	}
}

func (icv80Fam) Match(m *genModel, img []byte) bool { return len(img) == icv80Opt.size }

func (icv80Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 70000000 || d < -70000000
}

func (f icv80Fam) Get(m *genModel, img []byte, i int) *Chan {
	if icomBit(f.mem().F("unused"), img, i) {
		return nil
	}
	r := f.rec(i)
	mult := int64(5000)
	if r.F("mult").Int(img) != 0 {
		mult = 6250
	}
	c := &Chan{Rx: r.F("freq").Int(img) * mult}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * mult
	switch {
	case !r.F("tx_inhibit").Bool(img): // il bit vale 1 quando la trasmissione è permessa
		c.TxOff = true
	case r.F("duplex").Int(img) == 1:
		c.Tx = c.Rx - offs
	case r.F("duplex").Int(img) == 2:
		c.Tx = c.Rx + offs
	}
	c.TxTone, c.RxTone = icomTones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("mode").Bool(img)
	c.Scan = !icomBit(f.mem().F("skip"), img, i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

// icomTonesFor: valori dei campi dei toni per il canale c, partendo da quelli
// presenti (che restano quando il modo non li usa, come in CHIRP)
func icomTonesFor(c *Chan, rtone, ctone, dtcs int64) (tmode, rt, ct, dc, pol int64, err error) {
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return
	}
	r, err := genToneParse(c.RxTone)
	if err != nil {
		return
	}
	tmode, rt, ct, dc = icomToneMode(t, r), rtone, ctone, dtcs
	switch tmode {
	case -1:
		err = fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
		return
	case 1:
		rt = icomIndex(chirpTones, t.Hz10)
	case 2:
		rt = icomIndex(chirpTones, t.Hz10)
		ct = rt
	case 3:
		dc = icomIndex(dtcsCodes, t.Code)
		if t.Inv {
			pol |= 2
		}
		if r.Inv {
			pol |= 1
		}
	}
	if rt < 0 || ct < 0 || dc < 0 {
		err = fmt.Errorf("tono non disponibile su questa radio")
	}
	return
}

func (f icv80Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	e := f.mem()
	if c == nil {
		r.Fill(img, 0)
		r.F("freq").SetInt(img, 146010000/5000)
		r.F("offset").SetInt(img, 600000/5000)
		r.F("name").SetStr(img, "", ' ')
		r.F("tx_inhibit").SetInt(img, 1)
		r.F("rtone").SetInt(img, 8)
		r.F("ctone").SetInt(img, 8)
		icomSetBit(e.F("unused"), img, i, true)
		icomSetBit(e.F("skip"), img, i, true)
		return nil
	}
	was := !icomBit(e.F("unused"), img, i)
	kept, rtone, ctone, dtcs, step, rev := int64(600000), int64(8), int64(8), int64(0), int64(0), int64(0)
	if was {
		mult := int64(5000)
		if r.F("mult").Int(img) != 0 {
			mult = 6250
		}
		kept = r.F("offset").Int(img) * mult
		rtone, ctone, dtcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
		step, rev = r.F("tuning_step").Int(img), r.F("reverse_duplex").Int(img)
	}
	tmode, rtone, ctone, dtcs, pol, err := icomTonesFor(c, rtone, ctone, dtcs)
	if err != nil {
		return err
	}
	mult := int64(5000)
	switch {
	case c.Rx%5000 == 0:
	case c.Rx%6250 == 0:
		mult = 6250
	default:
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5 o di 6,25 kHz)", fmtMHz(c.Rx))
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case c.TxOff:
		offs = 0
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if dup != 0 && offs%mult != 0 {
		return fmt.Errorf("con la frequenza %s MHz la radio non può memorizzare uno scostamento di %d Hz", fmtMHz(c.Rx), offs)
	}
	if c.Rx/mult > 0xFFFF || offs/mult > 0xFFFF {
		return fmt.Errorf("frequenza o scostamento fuori dai limiti della radio")
	}
	r.Fill(img, 0)
	if mult == 6250 {
		r.F("mult").SetInt(img, 3)
	}
	r.F("freq").SetInt(img, c.Rx/mult)
	r.F("offset").SetInt(img, offs/mult)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("duplex").SetInt(img, dup)
	r.F("tx_inhibit").SetBool(img, !c.TxOff)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("tuning_step").SetInt(img, step)
	r.F("mode").SetBool(img, c.Narrow)
	r.F("tmode").SetInt(img, tmode)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("reverse_duplex").SetInt(img, rev)
	icomSetBit(e.F("unused"), img, i, false)
	icomSetBit(e.F("skip"), img, i, !c.Scan)
	return nil
}

func (f icv80Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icv80Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}

// ---------------------------------------------------------------- IC-T70

type ict70Fam struct{}

const ict70Format = `
struct {
  u24 freq;
  ul16 offset;
  char name[6];
  u8 unknown2:2,
     rtone:6;
  u8 unknown3:2,
     ctone:6;
  u8 unknown4:1,
     dtcs:7;
  u8 tuning_step:4,
     narrow:1,
     unknown5:1,
     duplex:2;
  u8 unknown6:1,
     power:2,
     dtcs_polarity:2,
     tmode:3;
} memory[300];
#seekto 0x12E0;
u8 used[38];
#seekto 0x1306;
u8 skips[38];
#seekto 0x132C;
u8 pskips[38];
`

var ict70Opt = &icomOpt{model: [4]byte{0x32, 0x53, 0x00, 0x01}, end: "Icom Inc.CF", size: 0x19E0, ranges: [][3]int{{0, 0x19E0, 32}}}

func (ict70Fam) icom(m *genModel) *icomOpt  { return ict70Opt }
func (ict70Fam) mem() *bwEl                 { return bwMust(ict70Format) }
func (f ict70Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f ict70Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

func (f ict70Fam) Rec(m *genModel, img []byte, i int) []byte {
	e := f.mem()
	return append(append([]byte{}, f.rec(i).Raw(img)...), icomB(icomBit(e.F("skips"), img, i)), icomB(icomBit(e.F("pskips"), img, i)))
}

func (f ict70Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	e := f.mem()
	if n := f.rec(i).ByteLen(); len(raw) == n+2 {
		f.rec(i).SetRaw(img, raw[:n])
		icomSetBit(e.F("skips"), img, i, raw[n] != 0)
		icomSetBit(e.F("pskips"), img, i, raw[n+1] != 0)
		icomSetBit(e.F("used"), img, i, false)
	}
}

func (ict70Fam) Match(m *genModel, img []byte) bool { return len(img) == ict70Opt.size }

func (ict70Fam) chirpInvalid(m *genModel, c *Chan) bool { return icv80Fam{}.chirpInvalid(m, c) }

func (f ict70Fam) Get(m *genModel, img []byte, i int) *Chan {
	e := f.mem()
	if icomBit(e.F("used"), img, i) {
		return nil
	}
	r := f.rec(i)
	v := r.F("freq").Int(img)
	c := &Chan{Rx: v * 5000}
	if v&0x800000 != 0 {
		c.Rx = (v &^ 0x800000) * 6250
	}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * 5000
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	// modi 3 e 5: subtono e DCS con l'avviso acustico; qui valgono come 2 e 4
	tm := map[int64]int64{1: 1, 2: 2, 3: 2, 4: 3, 5: 3}[r.F("tmode").Int(img)]
	c.TxTone, c.RxTone = icomTones(tm, r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = icomBit(e.F("pskips"), img, i) || !icomBit(e.F("skips"), img, i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f ict70Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	e := f.mem()
	if c == nil {
		r.Fill(img, 0)
		icomSetBit(e.F("used"), img, i, true)
		return nil
	}
	was := !icomBit(e.F("used"), img, i)
	kept, rtone, ctone, dtcs, step := int64(600000), int64(8), int64(8), int64(0), int64(0)
	if was {
		kept = r.F("offset").Int(img) * 5000
		rtone, ctone, dtcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
		if step = r.F("tuning_step").Int(img); step < 4 {
			step = 0 // i codici 0-3 sono tutti il passo da 5 kHz
		}
	}
	tmode, rtone, ctone, dtcs, pol, err := icomTonesFor(c, rtone, ctone, dtcs)
	if err != nil {
		return err
	}
	if tmode == 3 {
		tmode = 4
	}
	var fv int64
	switch {
	case c.Rx%12500 == 0, c.Rx%5000 != 0 && c.Rx%6250 == 0:
		fv = c.Rx/6250 | 0x800000
	case c.Rx%5000 == 0:
		fv = c.Rx / 5000
	default:
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5 o di 6,25 kHz)", fmtMHz(c.Rx))
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if dup != 0 && offs%5000 != 0 {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz (multipli di 5 kHz)", offs)
	}
	if offs/5000 > 0xFFFF {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	r.Fill(img, 0)
	icomSetBit(e.F("used"), img, i, false)
	r.F("freq").SetInt(img, fv)
	r.F("offset").SetInt(img, offs/5000)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("tuning_step").SetInt(img, step)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("duplex").SetInt(img, dup)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("tmode").SetInt(img, tmode)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	icomSetBit(e.F("skips"), img, i, !c.Scan)
	icomSetBit(e.F("pskips"), img, i, false)
	return nil
}

func (f ict70Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f ict70Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
