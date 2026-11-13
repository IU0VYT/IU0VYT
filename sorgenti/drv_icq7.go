// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icq7.py (4.33).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.33: Icom IC-Q7A, bibanda 144 / 430 MHz con ricevitore a larga
// banda; 200 canali da 0 a 199, senza nome. Canale di 8 byte: frequenza in kHz
// (sei cifre; oltre 1 GHz le prime quattro cifre binarie contano le centinaia
// di MHz, 10-13), un bit per i passi da 6,25 e 12,5 kHz (la parte sotto il kHz
// non è memorizzata), scostamento in kHz, subtoni e passo di sintonia. Un byte
// per canale a 0x0690 (FF = vuoto) con modo dei toni, verso dello scostamento,
// salti e modo (FM, FM larga, AM, automatico).
//
// Il modo non è una colonna del programma: resta quello del canale se la
// frequenza non cambia, altrimenti segue la frequenza (AM sotto i 30 MHz e in
// banda aerea, FM larga da 76 a 108 MHz, FM altrove), come per l'IC-E90.
//
// Differenze volute rispetto a CHIRP:
//   - sopra 1 GHz CHIRP calcola la frequenza in modo errato; qui si scrive come
//     la legge CHIRP stesso;
//   - una frequenza con il bit dei passi fini in banda aerea viene letta come
//     multiplo di 6,25 kHz (CHIRP la porta sulla griglia a 8,33 kHz, che questa
//     radio non ha);
//   - il salto «P» resta quando il canale modificato resta fuori dalla scansione.

import (
	"context"
	"fmt"
)

type icq7Fam struct{}

const icq7Format = `
struct {
  bbcd freq[3];
  u8  fractional:1,
      unknown:7;
  bbcd offset[2];
  u16 ctone:6,
      rtone:6,
      tune_step:4;
} memory[200];

#seekto 0x0690;
struct {
  u8 tmode:2,
     duplex:2,
     skip:1,
     pskip:1,
     mode:2;
} flags[200];
`

var icq7Opt = &icomOpt{model: [4]byte{0x19, 0x95, 0x00, 0x01}, end: "Icom Inc.", size: 0x07C0, ranges: [][3]int{{0x0000, 0x07C0, 16}}}

func init() {
	genRegister(&genModel{ID: "ic-q7a", Vendor: "Icom", Model: "IC-Q7A", RClass: "ICQ7Radio", Name: "Icom IC-Q7A",
		Slots: 200, First: 0, Powers: []genPower{{"Unica", 0.5, 0}},
		Bands: []band{{1000000, 823995000}, {849000000, 868995000}, {894000000, 1309995000}}, HasScan: true, NoTxOff: true, Baud: 9600, Fam: icq7Fam{},
		Notes: "200 canali (da 0 a 199) senza nome; ricevitore a larga banda. " + icomCableNote + icomNoDCS + icomNoPower + icomModeHz + "Driver di CHIRP."})
}

func (icq7Fam) icom(m *genModel) *icomOpt  { return icq7Opt }
func (icq7Fam) mem() *bwEl                 { return bwMust(icq7Format) }
func (f icq7Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f icq7Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f icq7Fam) flg(i int) *bwEl          { return f.mem().F("flags").I(i) }

func (f icq7Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.flg(i).Raw(img)...)
}

func (f icq7Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+1 && raw[n] != 0xFF {
		f.rec(i).SetRaw(img, raw[:n])
		f.flg(i).SetRaw(img, raw[n:])
	}
}

func (icq7Fam) Match(m *genModel, img []byte) bool { return len(img) == icq7Opt.size }

func (icq7Fam) toneOK(t, r genTone) bool { return ict7Fam{}.toneOK(t, r) }
func (icq7Fam) fitBySet()                {}

func (icq7Fam) chirpInvalid(m *genModel, c *Chan) bool { return ict7Fam{}.chirpInvalid(m, c) }

// kHz memorizzati: cifre decimali, tranne le prime quattro cifre binarie che
// oltre 1 GHz valgono 10-13
func icq7KHz(raw []byte) int64 {
	v := int64(0)
	for _, b := range raw {
		v = v*100 + int64(b>>4)*10 + int64(b&0x0F)
	}
	return v
}

func icq7Mode(hz int64) int64 { return icx90Mode(hz) }

func (f icq7Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, fl := f.rec(i), f.flg(i)
	if fl.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: icq7KHz(r.F("freq").Raw(img)) * 1000}
	if r.F("fractional").Bool(img) {
		for _, add := range []int64{500, 250, 750} {
			if (c.Rx+add)%6250 == 0 {
				c.Rx += add
				break
			}
		}
	}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * 1000
	switch fl.F("duplex").Int(img) {
	case 2:
		c.Tx = c.Rx - offs
	case 3:
		c.Tx = c.Rx + offs
	}
	mode := int64(0)
	switch fl.F("tmode").Int(img) {
	case 2:
		mode = 1
	case 3:
		mode = 2
	}
	c.TxTone, c.RxTone = icomTones(mode, r.F("rtone").Int(img), r.F("ctone").Int(img), 0, 0)
	c.Scan = !fl.F("skip").Bool(img) && !fl.F("pskip").Bool(img)
	return c
}

func (f icq7Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(i), f.flg(i)
	if c == nil {
		fl.Fill(img, 0xFF)
		return nil
	}
	was := fl.Raw(img)[0] != 0xFF
	kept, rtone, ctone, step, mode := int64(600000), int64(8), int64(8), int64(0), icq7Mode(c.Rx)
	pskip := false
	if was {
		old := f.Get(m, img, i)
		kept = r.F("offset").Int(img) * 1000
		rtone, ctone, step = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("tune_step").Int(img)
		if rtone >= int64(len(chirpTones)) {
			rtone = 8
		}
		if ctone >= int64(len(chirpTones)) {
			ctone = 8
		}
		if step > 9 {
			step = 0
		}
		if old.Rx == c.Rx {
			mode = fl.F("mode").Int(img)
		}
		pskip = fl.F("pskip").Bool(img) && !fl.F("skip").Bool(img) && !c.Scan
	}
	tmode, rtone, ctone, _, _, err := icomTonesFor(c, rtone, ctone, 0)
	if err != nil {
		return err
	}
	if tmode == 3 {
		return fmt.Errorf("la radio non ha i codici DCS")
	}
	frac := c.Rx%5000 != 0
	if frac && c.Rx%6250 != 0 || !frac && c.Rx%1000 != 0 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5 o di 6,25 kHz)", fmtMHz(c.Rx))
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 3, d
	case d < 0:
		dup, offs = 2, -d
	}
	if dup != 0 && offs%1000 != 0 {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz (multipli di 1 kHz)", offs)
	}
	if offs > 9999000 {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	r.Fill(img, 0)
	khz := c.Rx / 1000
	r.F("freq").SetRaw(img, []byte{byte(khz/100000)<<4 | byte(khz/10000%10), byte(khz/1000%10)<<4 | byte(khz/100%10), byte(khz/10%10)<<4 | byte(khz%10)})
	r.F("fractional").SetBool(img, frac)
	r.F("offset").SetInt(img, offs/1000)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("tune_step").SetInt(img, step)
	fl.F("tmode").SetInt(img, []int64{0, 2, 3}[tmode])
	fl.F("duplex").SetInt(img, dup)
	fl.F("mode").SetInt(img, mode)
	fl.F("skip").SetBool(img, !c.Scan && !pskip)
	fl.F("pskip").SetBool(img, pskip)
	return nil
}

func (f icq7Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icq7Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("l'immagine non è di una %s", m.Model)
	}
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
