// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icx90.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: Icom IC-E90 (Europa) e IC-T90 / IC-T90A (altri mercati): tribanda
// 50/144/430 MHz con ricevitore da 0,495 a 999,990 MHz, 500 canali con nomi di 6
// caratteri. Il modo (FM, FM larga delle radio commerciali, AM) non è una
// colonna del programma: resta quello del canale se la frequenza non cambia,
// altrimenti segue la frequenza (AM sotto i 30 MHz e in banda aerea, FM larga da
// 76 a 108 MHz, FM altrove).
//
// Differenze volute rispetto a CHIRP:
//   - una frequenza che non è multiplo di 5 o 6,25 kHz viene rifiutata (CHIRP la
//     arrotonda verso il basso);
//   - il salto «P» resta com'è quando il canale modificato resta escluso dalla
//     scansione.

import (
	"context"
	"fmt"
	"strings"
)

type icx90Fam struct{}

const icx90Format = `
struct mem_item {
  ul24 freq;
  u8 dtcs_polarity:2,
     unknown_1:2,
     offset_freq_mult:1,
     unknown_2:2,
     freq_mult:1;
  u8 unknown_3:1,
     duplex:2,
     mode:2,
     tone_mode:3;
  ul16 offset_freq;
  u8 dtcs;
  u8 tx_tone_lo:4,
     tune_step:4;
  u8 rx_tone:6,
     tx_tone_hi:2;
  char name[6];
};
struct bank_item {
  u8 invisible_channel:1,
     prog_skip:1,
     mem_skip:1,
     bank_index:5;
  u8 bank_channel;
};
struct mem_item memory[500];
struct mem_item scan_edges[50];
struct bank_item banks[500];
`

var icx90Opt = &icomOpt{model: [4]byte{0x25, 0x07, 0x00, 0x01}, end: "Icom Inc.", size: 0x2D40, ranges: [][3]int{{0, 0x2D40, 32}}}

func init() {
	genRegister(&genModel{ID: "ic-e90", Vendor: "Icom", Model: "IC-E90", RClass: "ICx90Radio", Name: "Icom IC-E90 / IC-T90",
		Slots: 500, First: 0, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789()*+-,/|= ", Powers: []genPower{{"Unica", 5, 0}},
		Bands: []band{{495000, 999990000}}, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: icx90Fam{},
		Notes: "500 canali (da 0 a 499) con nomi di 6 caratteri; ricevitore a larga banda. " + icomCableNote +
			"Il modo (FM, FM larga, AM) segue la frequenza. Driver di CHIRP."})
}

func (icx90Fam) icom(m *genModel) *icomOpt  { return icx90Opt }
func (icx90Fam) mem() *bwEl                 { return bwMust(icx90Format) }
func (f icx90Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f icx90Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f icx90Fam) bank(i int) *bwEl         { return f.mem().F("banks").I(i) }

func (f icx90Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.bank(i).Raw(img)...)
}

func (f icx90Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+2 {
		f.rec(i).SetRaw(img, raw[:n])
		f.bank(i).SetRaw(img, raw[n:])
		f.bank(i).F("invisible_channel").SetBool(img, false)
	}
}

func (icx90Fam) Match(m *genModel, img []byte) bool { return len(img) == icx90Opt.size }

func (icx90Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (icx90Fam) fitBySet()                {}

func (icx90Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 70000000 || d < -70000000
}

func icx90Mult(on bool) int64 {
	if on {
		return 6250
	}
	return 5000
}

func (f icx90Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	rx := r.F("freq").Int(img) * icx90Mult(r.F("freq_mult").Bool(img))
	if rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	offs := r.F("offset_freq").Int(img) * icx90Mult(r.F("offset_freq_mult").Bool(img))
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	rtone := r.F("tx_tone_hi").Int(img)<<4 | r.F("tx_tone_lo").Int(img)
	c.TxTone, c.RxTone = icomTones(r.F("tone_mode").Int(img), rtone, r.F("rx_tone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	b := f.bank(i)
	c.Scan = !b.F("mem_skip").Bool(img) && !b.F("prog_skip").Bool(img)
	c.Name = strings.TrimRight(r.F("name").Str(img), "\x00 ")
	return c
}

// modo che segue la frequenza: 0 FM, 1 FM larga (radio commerciali), 2 AM
func icx90Mode(hz int64) int64 {
	switch {
	case hz < 30000000, hz >= 108000000 && hz < 137000000:
		return 2
	case hz >= 76000000 && hz < 108000000:
		return 1
	}
	return 0
}

func (f icx90Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, b := f.rec(i), f.bank(i)
	if c == nil {
		r.Fill(img, 0)
		b.F("bank_index").SetInt(img, 0x1F)
		b.F("prog_skip").SetBool(img, false)
		b.F("mem_skip").SetBool(img, false)
		b.F("bank_channel").SetInt(img, 0)
		b.F("invisible_channel").SetBool(img, true)
		return nil
	}
	oldRx := r.F("freq").Int(img) * icx90Mult(r.F("freq_mult").Bool(img))
	was := oldRx != 0
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	rtone, ctone, dtcs, offs, step := int64(8), int64(8), int64(0), int64(600000), int64(0)
	if was {
		rtone = r.F("tx_tone_hi").Int(img)<<4 | r.F("tx_tone_lo").Int(img)
		ctone, dtcs = r.F("rx_tone").Int(img), r.F("dtcs").Int(img)
		offs = r.F("offset_freq").Int(img) * icx90Mult(r.F("offset_freq_mult").Bool(img))
		step = r.F("tune_step").Int(img)
	}
	tmode, rtone, ctone, dtcs, pol, err := icomTonesFor(c, rtone, ctone, dtcs)
	if err != nil {
		return err
	}
	if c.Rx%5000 != 0 && c.Rx%6250 != 0 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5 o di 6,25 kHz)", fmtMHz(c.Rx))
	}
	fmult := c.Rx%5000 != 0
	dup := int64(0)
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if offs%5000 != 0 && offs%6250 != 0 {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz (multipli di 5 o di 6,25 kHz)", offs)
	}
	omult := offs%5000 != 0
	if offs/icx90Mult(omult) > 0xFFFF {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	mode := icx90Mode(c.Rx)
	if was && oldRx == c.Rx {
		mode = r.F("mode").Int(img)
	}
	r.F("freq").SetInt(img, c.Rx/icx90Mult(fmult))
	r.F("freq_mult").SetBool(img, fmult)
	r.F("offset_freq").SetInt(img, offs/icx90Mult(omult))
	r.F("offset_freq_mult").SetBool(img, omult)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("tx_tone_hi").SetInt(img, rtone>>4)
	r.F("tx_tone_lo").SetInt(img, rtone&0x0F)
	r.F("rx_tone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("duplex").SetInt(img, dup)
	r.F("tone_mode").SetInt(img, tmode)
	r.F("tune_step").SetInt(img, step)
	r.F("mode").SetInt(img, mode)
	b.F("invisible_channel").SetBool(img, false)
	skipped := was && b.F("mem_skip").Bool(img)
	b.F("mem_skip").SetBool(img, !c.Scan)
	if c.Scan || !skipped {
		b.F("prog_skip").SetBool(img, false)
	}
	return nil
}

func (f icx90Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icx90Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
