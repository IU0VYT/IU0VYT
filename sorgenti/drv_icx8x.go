// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icx8x.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: Icom IC-V82 (VHF) e IC-U82 (UHF): 200 canali con nomi di 5 caratteri.
// Stessa memoria; il byte 0x1930 dice se la radio è UHF (frequenze a partire da
// 400 MHz). La radio vuole l'identificazione due volte prima della scrittura.
//
// Differenza voluta rispetto a CHIRP: un canale in modo digitale DV (radio con il
// modulo UT-118) resta DV quando lo si modifica; CHIRP lo riporta in FM.

import (
	"context"
	"fmt"
	"strings"
)

type icx8xFam struct{}

type icx8xOpt struct {
	base int64 // Hz da aggiungere alla frequenza memorizzata
	uhf  byte
}

const icx8xFormat = `
struct {
  ul16 freq;
  ul16 offset;
  char name[5];
  u8 unknown1:2,
     rtone:6;
  u8 unknown2:2,
     ctone:6;
  u8 dtcs;
  u8 unknown3[5];
  u8 tuning_step:4,
    unknown4:4;
  u8 unknown5[3];
  u8 mult:1,
     unknown6_0:1,
     narrow:1,
     unknown6:1,
     unknown7:2,
     tmode:2;
  u8 dtcs_pol:2,
     duplex:2,
     unknown9:4;
  u8 unknown10:1,
     txinhibit:1,
     unknown11:2,
     dvmode:1,
     unknown12:3;
} memories[207];
#seekto 0x1370;
struct {
    u8 unknown:2,
       empty:1,
       skip:1,
       bank:4;
} flags[206];
#seekto 0x1930;
u8 isuhf;
`

var icx8xClone = &icomOpt{model: [4]byte{0x28, 0x26, 0x00, 0x01}, end: "Icom Inc.CD", size: 6464, twice: true,
	ranges: [][3]int{{0x0000, 0x1340, 32}, {0x1340, 0x1360, 16}, {0x1360, 0x136B, 8}, {0x1370, 0x1440, 32},
		{0x1460, 0x15D0, 32}, {0x15E0, 0x1930, 32}, {0x1938, 0x1940, 8}}}

func init() {
	mk := func(id, model, rclass string, o *icx8xOpt, b band, what string) *genModel {
		return &genModel{ID: id, Vendor: "Icom", Model: model, RClass: rclass, Name: "Icom " + model,
			Slots: 200, First: 0, NameMax: 5, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890", Powers: []genPower{{"Unica", 5, 0}},
			Bands: []band{b}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: icx8xFam{}, Opt: o,
			Notes: "200 canali " + what + " (da 0 a 199) con nomi di 5 caratteri. " + icomCableNote + "La potenza non è memorizzata nel canale. " +
				"Banchi, nominativi D-STAR, passo di sintonia e canali in modo digitale DV si conservano. Driver di CHIRP."}
	}
	genRegister(
		mk("ic-v82", "IC-V82", "ICV82Radio", &icx8xOpt{base: 0, uhf: 0}, band{118000000, 176000000}, "VHF"),
		mk("ic-u82", "IC-U82", "ICU82Radio", &icx8xOpt{base: 400000000, uhf: 1}, band{420000000, 470000000}, "UHF"),
	)
}

func (icx8xFam) icom(m *genModel) *icomOpt  { return icx8xClone }
func (icx8xFam) mem() *bwEl                 { return bwMust(icx8xFormat) }
func (f icx8xFam) layout(m *genModel) *bwEl { return f.mem() }
func (f icx8xFam) rec(i int) *bwEl          { return f.mem().F("memories").I(i) }
func (f icx8xFam) flg(i int) *bwEl          { return f.mem().F("flags").I(i) }

func (f icx8xFam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), icomB(f.flg(i).F("skip").Bool(img)))
}

func (f icx8xFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if n := f.rec(i).ByteLen(); len(raw) == n+1 {
		f.rec(i).SetRaw(img, raw[:n])
		f.flg(i).F("skip").SetBool(img, raw[n] != 0)
		f.flg(i).F("empty").SetBool(img, false)
	}
}

func (icx8xFam) Match(m *genModel, img []byte) bool {
	return len(img) == icx8xClone.size && img[0x1930] == m.Opt.(*icx8xOpt).uhf
}

func (icx8xFam) chirpInvalid(m *genModel, c *Chan) bool { return icv80Fam{}.chirpInvalid(m, c) }

func (f icx8xFam) Get(m *genModel, img []byte, i int) *Chan {
	if f.flg(i).F("empty").Bool(img) {
		return nil
	}
	r := f.rec(i)
	mult := int64(5000)
	if r.F("mult").Bool(img) {
		mult = 6250
	}
	c := &Chan{Rx: r.F("freq").Int(img)*mult + m.Opt.(*icx8xOpt).base}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * 5000
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	c.TxTone, c.RxTone = icomTones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_pol").Int(img))
	c.Narrow = r.F("narrow").Bool(img) && !r.F("dvmode").Bool(img)
	c.Scan = !f.flg(i).F("skip").Bool(img)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f icx8xFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(i), f.flg(i)
	if c == nil {
		fl.F("empty").SetBool(img, true)
		return nil
	}
	base := m.Opt.(*icx8xOpt).base
	was := !fl.F("empty").Bool(img)
	kept, rtone, ctone, dtcs, step := int64(600000), int64(8), int64(8), int64(0), int64(0)
	dv := false
	if was {
		kept = r.F("offset").Int(img) * 5000
		rtone, ctone, dtcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
		step, dv = r.F("tuning_step").Int(img), r.F("dvmode").Bool(img)
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
	fv := (c.Rx - base) / mult
	if c.Rx < base || fv > 0xFFFF {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
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
	fl.F("skip").SetBool(img, !c.Scan)
	fl.F("empty").SetBool(img, false)
	r.F("mult").SetBool(img, mult == 6250)
	r.F("freq").SetInt(img, fv)
	r.F("offset").SetInt(img, offs/5000)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("dvmode").SetBool(img, dv)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("duplex").SetInt(img, dup)
	r.F("tmode").SetInt(img, tmode)
	r.F("tuning_step").SetInt(img, step)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("dtcs_pol").SetInt(img, pol)
	return nil
}

func (f icx8xFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	if err == nil && !f.Match(m, img) {
		other := "IC-U82"
		if m.Opt.(*icx8xOpt).uhf == 1 {
			other = "IC-V82"
		}
		return nil, "", rErr("la radio è l'altro modello della serie: apri la scheda «Icom %s»", other)
	}
	return img, "", err
}

func (f icx8xFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("l'immagine non è di una %s", m.Model)
	}
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
