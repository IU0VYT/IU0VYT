// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icp7.py (4.33).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.33: Icom IC-P7, bibanda 144 / 430 MHz con ricevitore da 0,495 a
// 999,990 MHz; 1000 canali da 0 a 999 con nomi di 6 caratteri (seguono, non
// gestiti qui, i limiti di scansione, i canali scritti in automatico e i due
// canali di chiamata). Canale di 21 byte: frequenza e scostamento in terzi di
// Hz, modo dei toni, verso dello scostamento, passo, subtoni, modo, DCS e
// polarità, nome. Tre aree di bit (vuoto, salto, salto «P»), poi banchi e nomi
// dei banchi, che non si toccano.
//
// Il modo (FM, FM larga, AM, automatico) non è una colonna del programma: resta
// quello del canale se la frequenza non cambia, altrimenti segue la frequenza,
// come per l'IC-E90. Se frequenza o scostamento non cambiano restano scritti
// come sono (i terzi di Hz dei passi da 8,33 kHz non si perdono).
//
// Differenze volute rispetto a CHIRP:
//   - il salto «P» resta quando il canale modificato resta fuori dalla scansione;
//   - un canale con il passo di sintonia «automatico» resta così (CHIRP, quando
//     lo modifica, lo sostituisce con il passo adatto alla vecchia frequenza).

import (
	"context"
	"fmt"
	"strings"
)

type icp7Fam struct{}

const icp7Format = `
struct {
  ul32 freq;
  ul32 offset;
  ul16 train_sql:2,
       tmode:3,
       duplex:2,
       train_tone:9;
  ul16 tuning_step:4,
       rtone:6,
       ctone:6;
  ul16 unknown0:6,
       mode:3,
       dtcs:7;
  u8   unknown1:6,
       dtcs_polarity:2;
  char name[6];
} memory[1251];

#seekto 0x6b1e;
struct {
  u8 bank;
  u8 index;
} banks[1050];

#seekto 0x689e;
u8 used[132];

#seekto 0x6922;
u8 skips[132];

#seekto 0x69a6;
u8 pskips[132];

#seekto 0x7352;
struct {
  char name[6];
} bank_names[18];
`

var icp7Opt = &icomOpt{model: [4]byte{0x28, 0x69, 0x00, 0x01}, end: "Icom Inc.A8", size: 0x7500, fast: true, ranges: [][3]int{{0x0000, 0x7500, 32}}}

func init() {
	genRegister(&genModel{ID: "ic-p7", Vendor: "Icom", Model: "IC-P7", RClass: "ICP7Radio", Name: "Icom IC-P7",
		Slots: 1000, First: 0, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890", Powers: []genPower{{"Unica", 1.5, 0}},
		Bands: []band{{495000, 999990000}}, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: icp7Fam{},
		Notes: "1000 canali (da 0 a 999) con nomi di 6 caratteri; ricevitore a larga banda. " + icomCableNote + icomFastCp + icomNoPower + icomModeHz + "Driver di CHIRP."})
}

func (icp7Fam) icom(m *genModel) *icomOpt  { return icp7Opt }
func (icp7Fam) mem() *bwEl                 { return bwMust(icp7Format) }
func (f icp7Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f icp7Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

func (f icp7Fam) Rec(m *genModel, img []byte, i int) []byte {
	e := f.mem()
	return append(append([]byte{}, f.rec(i).Raw(img)...), icomB(icomBit(e.F("skips"), img, i)), icomB(icomBit(e.F("pskips"), img, i)))
}

func (f icp7Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	e := f.mem()
	if n := f.rec(i).ByteLen(); len(raw) == n+2 {
		f.rec(i).SetRaw(img, raw[:n])
		icomSetBit(e.F("skips"), img, i, raw[n] != 0)
		icomSetBit(e.F("pskips"), img, i, raw[n+1] != 0)
		icomSetBit(e.F("used"), img, i, false)
	}
}

func (icp7Fam) Match(m *genModel, img []byte) bool { return len(img) == icp7Opt.size }

func (icp7Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (icp7Fam) fitBySet()                {}

func (icp7Fam) chirpInvalid(m *genModel, c *Chan) bool { return icv80Fam{}.chirpInvalid(m, c) }

// chirpEncodeFix (solo prove): il passo «automatico» (13) che qui resta, vedi sopra
func (f icp7Fam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	for i := 0; i < m.Slots; i++ {
		st := f.rec(i).F("tuning_step")
		if st.Int(base) == 13 && st.Int(out) == 13 && !icomBit(f.mem().F("used"), ref, i) {
			st.SetInt(ref, 13)
		}
	}
}

func (f icp7Fam) Get(m *genModel, img []byte, i int) *Chan {
	e := f.mem()
	if icomBit(e.F("used"), img, i) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("freq").Int(img) / 3}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) / 3
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	mode := r.F("tmode").Int(img)
	switch {
	case mode == 4:
		mode = 3
	case mode > 2:
		mode = 0
	}
	c.TxTone, c.RxTone = icomTones(mode, r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	c.Scan = !icomBit(e.F("skips"), img, i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f icp7Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	e, r := f.mem(), f.rec(i)
	if c == nil { // come CHIRP: valori di un canale mai usato
		icomSetBit(e.F("used"), img, i, true)
		r.F("freq").SetInt(img, 15000)
		r.F("offset").SetInt(img, 479985000)
		r.F("train_sql").SetInt(img, 3)
		r.F("tmode").SetInt(img, 7)
		r.F("duplex").SetInt(img, 3)
		r.F("train_tone").SetInt(img, 0x1FF)
		r.F("tuning_step").SetInt(img, 15)
		r.F("rtone").SetInt(img, 63)
		r.F("ctone").SetInt(img, 63)
		r.F("unknown0").SetInt(img, 0)
		r.F("mode").SetInt(img, 7)
		r.F("dtcs").SetInt(img, 127)
		r.F("unknown1").SetInt(img, 63)
		r.F("dtcs_polarity").SetInt(img, 3)
		r.F("name").SetStr(img, "", ' ')
		icomSetBit(e.F("skips"), img, i, true)
		icomSetBit(e.F("pskips"), img, i, true)
		return nil
	}
	was := !icomBit(e.F("used"), img, i)
	rawF, rawO := c.Rx*3, int64(600000*3)
	rtone, ctone, dtcs, step, mode := int64(8), int64(8), int64(0), int64(0), icx90Mode(c.Rx)
	pskip, sameOffs := false, false
	if was {
		old := f.Get(m, img, i)
		rawO = r.F("offset").Int(img)
		rtone, ctone, dtcs, step = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("tuning_step").Int(img)
		if rtone >= int64(len(chirpTones)) {
			rtone = 8
		}
		if ctone >= int64(len(chirpTones)) {
			ctone = 8
		}
		if dtcs >= int64(len(dtcsCodes)) {
			dtcs = 0
		}
		if old.Rx == c.Rx {
			rawF, mode = r.F("freq").Int(img), r.F("mode").Int(img)
		}
		sameOffs = old.Tx-old.Rx == c.Tx-c.Rx
		pskip = icomBit(e.F("pskips"), img, i) && icomBit(e.F("skips"), img, i) && !c.Scan
	}
	tmode, rtone, ctone, dtcs, pol, err := icomTonesFor(c, rtone, ctone, dtcs)
	if err != nil {
		return err
	}
	dup := int64(0)
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup = 2
		if !sameOffs {
			rawO = d * 3
		}
	case d < 0:
		dup = 1
		if !sameOffs {
			rawO = -d * 3
		}
	}
	if rawF > 0xFFFFFFFF || rawO > 0xFFFFFFFF {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	if tmode == 3 {
		tmode = 4
	}
	icomSetBit(e.F("used"), img, i, false)
	r.F("freq").SetInt(img, rawF)
	r.F("offset").SetInt(img, rawO)
	r.F("train_sql").SetInt(img, 0)
	r.F("tmode").SetInt(img, tmode)
	r.F("duplex").SetInt(img, dup)
	r.F("train_tone").SetInt(img, 228)
	r.F("tuning_step").SetInt(img, step)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("unknown0").SetInt(img, 0)
	r.F("mode").SetInt(img, mode)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("unknown1").SetInt(img, 63)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	icomSetBit(e.F("skips"), img, i, !c.Scan)
	icomSetBit(e.F("pskips"), img, i, pskip)
	return nil
}

func (f icp7Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icp7Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("l'immagine non è di una %s", m.Model)
	}
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
