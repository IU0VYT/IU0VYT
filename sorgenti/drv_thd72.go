// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// thd72.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: Kenwood TH-D72A/E (cavo USB; la radio compare come porta seriale). 1000
// canali con nomi di 8 caratteri. Stesso collegamento della TH-D74 (vedi
// drv_thd74.go), con il numero del blocco scritto al contrario e RTS attivo.
// Come in CHIRP si scrive tutta la memoria tranne gli ultimi due blocchi.
//
// Differenza voluta rispetto a CHIRP: anche il canale 999 ha il nome e il salto
// di scansione (CHIRP li scrive solo fino al canale 998).

import (
	"context"
	"fmt"
	"strings"
)

type thd72Fam struct{}

const thd72Format = `
#seekto 0x0c00;
struct {
  u8 disabled:4,
     prog_vfo:4;
  u8 skip;
} flag[1032];

#seekto 0x1500;
struct {
  ul32 freq;
  u8 unknown1:4,
     tune_step:4;
  u8 mode;
  u8 tone_mode:4,
     duplex:4;
  u8 rtone;
  u8 ctone;
  u8 dtcs;
  u8 cross_mode;
  ul32 offset;
  u8 unknown2:4,
     split_tune_step:4;
} memory[1032];

#seekto 0x5e00;
struct {
    char name[8];
} channel_name[1000];
`

// subtoni delle radio Kenwood (quelli di CHIRP senza 159,8 165,5 171,3 177,3
// 183,5 189,9 196,6 199,5 Hz)
var kenwoodTones = func() []int {
	skip := map[int]bool{1598: true, 1655: true, 1713: true, 1773: true, 1835: true, 1899: true, 1966: true, 1995: true}
	var out []int
	for _, t := range chirpTones {
		if !skip[t] {
			out = append(out, t)
		}
	}
	return out
}()

// gamme dei VFO programmabili (indice memorizzato accanto al canale)
var thd72Vfo = [][3]int64{{136000000, 174000000, 0}, {410000000, 470000000, 1}, {118000000, 136000000, 2},
	{320000000, 400000000, 4}, {400000000, 524000000, 5}}

func init() {
	genRegister(&genModel{ID: "kw-thd72", Vendor: "Kenwood", Model: "TH-D72 (clone mode)", RClass: "THD72Radio", Name: "Kenwood TH-D72A / TH-D72E",
		Slots: 1000, First: 0, NameMax: 8, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz 1234567890",
		Powers: []genPower{{"Unica", 5, 0}}, Bands: []band{{136000000, 174000000}, {320000000, 524000000}, {118000000, 136000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: thd72Fam{}, Opt: &thd74Opt{id: "TH-D72", size: 0x10000, d72: true},
		Notes: "1000 canali (da 0 a 999) con nomi di 8 caratteri. " + thdCableNote +
			"Subtoni Kenwood (mancano 159,8 165,5 171,3 177,3 183,5 189,9 196,6 e 199,5 Hz). Driver di CHIRP."})
}

func (thd72Fam) mem() *bwEl                 { return bwMust(thd72Format) }
func (f thd72Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f thd72Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f thd72Fam) flag(i int) *bwEl         { return f.mem().F("flag").I(i) }
func (f thd72Fam) name(i int) *bwEl         { return f.mem().F("channel_name").I(i).F("name") }

func (f thd72Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append(append([]byte{}, f.rec(i).Raw(img)...), f.flag(i).Raw(img)...)
	return append(out, f.name(i).Raw(img)...)
}

func (f thd72Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) != 16+2+8 {
		return
	}
	f.rec(i).SetRaw(img, raw[:16])
	f.flag(i).SetRaw(img, raw[16:18])
	f.name(i).SetRaw(img, raw[18:])
}

// Match: i file si riconoscono dai metadati (molte radio hanno 64 KiB di memoria)
func (thd72Fam) Match(m *genModel, img []byte) bool { return false }

// thd72Tmode: modo dei toni (0 nessuno, 1 subtono in trasmissione, 2 subtono nei
// due versi, 3 DCS nei due versi, poi i misti nell'ordine dell'elenco di CHIRP + 4:
// 4 due subtoni diversi, 5 DCS in trasmissione, 6 DCS in ricezione, 7 subtono e
// DCS, 8 DCS e subtono, 9 subtono in ricezione); -1 se la radio non ce l'ha
func thd72Tmode(t, r genTone) int {
	switch {
	case t.Inv || r.Inv:
		return -1
	case t.Mode == "" && r.Mode == "":
		return 0
	case t.Mode == "Tone" && r.Mode == "":
		return 1
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
		return 2
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
		return 3
	case t.Mode == "Tone" && r.Mode == "Tone":
		return 4
	case t.Mode == "DTCS" && r.Mode == "":
		return 5
	case t.Mode == "" && r.Mode == "DTCS":
		return 6
	case t.Mode == "Tone" && r.Mode == "DTCS":
		return 7
	case t.Mode == "DTCS" && r.Mode == "Tone":
		return 8
	case t.Mode == "" && r.Mode == "Tone":
		return 9
	}
	return -1
}

func (thd72Fam) toneOK(t, r genTone) bool {
	ok := func(g genTone) bool { return g.Mode != "Tone" || icomIndex(kenwoodTones, g.Hz10) >= 0 }
	return thd72Tmode(t, r) >= 0 && ok(t) && ok(r)
}
func (thd72Fam) fitBySet() {}

func (f thd72Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !f.toneOK(t, r)
}

func (f thd72Fam) Get(m *genModel, img []byte, i int) *Chan {
	fl, r := f.flag(i), f.rec(i)
	if fl.F("disabled").Int(img) == 0xF {
		return nil
	}
	c := &Chan{Rx: r.F("freq").Int(img)}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img)
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx + offs
	case 2:
		c.Tx = c.Rx - offs
	case 4:
		c.Tx = offs
	}
	tone := func(k int64) string {
		if k >= 0 && int(k) < len(kenwoodTones) {
			return fmtTone(kenwoodTones[k])
		}
		return "88.5"
	}
	dcs := "D023N"
	if k := int(r.F("dtcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	rtone, ctone := tone(r.F("rtone").Int(img)), tone(r.F("ctone").Int(img))
	switch r.F("tone_mode").Int(img) {
	case 8:
		c.TxTone = rtone
	case 4:
		c.TxTone, c.RxTone = ctone, ctone
	case 2:
		c.TxTone, c.RxTone = dcs, dcs
	case 1:
		switch r.F("cross_mode").Int(img) {
		case 0:
			c.TxTone, c.RxTone = rtone, ctone
		case 1:
			c.TxTone = dcs
		case 2:
			c.RxTone = dcs
		case 3:
			c.TxTone, c.RxTone = rtone, dcs
		case 4:
			c.TxTone, c.RxTone = dcs, ctone
		case 5:
			c.RxTone = ctone
		case 6:
			c.TxTone, c.RxTone = dcs, dcs
		case 7:
			c.TxTone = rtone
		}
	}
	c.Narrow = r.F("mode").Int(img) == 1
	c.Scan = fl.F("skip").Int(img) == 0
	s := f.name(i).Str(img)
	if k := strings.IndexByte(s, 0xFF); k >= 0 {
		s = s[:k]
	}
	c.Name = strings.TrimRight(s, " ")
	return c
}

func (f thd72Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	fl, r := f.flag(i), f.rec(i)
	wasEmpty := fl.F("disabled").Int(img) == 0xF
	if c == nil {
		fl.F("disabled").SetInt(img, 0xF)
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
	tmode := thd72Tmode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (codici DCS diversi nei due versi o invertiti)")
	}
	vfo := int64(-1)
	for _, v := range thd72Vfo {
		if c.Rx >= v[0] && c.Rx < v[1] {
			vfo = v[2]
			break
		}
	}
	if vfo < 0 {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtoni 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	rtone, ctone, dcs, cross, offs, step, oldMode, oldRx, skip := int64(8), int64(8), int64(0), int64(0), int64(600000), int64(0), int64(0), int64(0), int64(0)
	if !wasEmpty {
		rtone, ctone, dcs, cross = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("cross_mode").Int(img)
		offs, step = r.F("offset").Int(img), r.F("tune_step").Int(img)
		oldMode, oldRx, skip = r.F("mode").Int(img), r.F("freq").Int(img), fl.F("skip").Int(img)
	}
	if step > 10 {
		step = 0
	}
	if cross > 7 {
		cross = 0
	}
	find := func(list []int, v int) (int64, error) {
		if k := icomIndex(list, v); k >= 0 {
			return k, nil
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	switch tmode {
	case 1, 4, 7:
		rtone, err = find(kenwoodTones, t.Hz10)
	case 2:
		if rtone, err = find(kenwoodTones, t.Hz10); err == nil {
			ctone = rtone
		}
	}
	if err == nil && (tmode == 4 || tmode == 8 || tmode == 9) {
		ctone, err = find(kenwoodTones, rt.Hz10)
	}
	if err == nil && (tmode == 3 || tmode == 5 || tmode == 8) {
		dcs, err = find(dtcsCodes, t.Code)
	}
	if err == nil && (tmode == 6 || tmode == 7) {
		dcs, err = find(dtcsCodes, rt.Code)
	}
	if err != nil {
		return err
	}
	tm := []int64{0, 8, 4, 2}
	tonemode := int64(1)
	if tmode < 4 {
		tonemode = tm[tmode]
	} else {
		cross = []int64{0, 1, 2, 3, 4, 5}[tmode-4]
	}
	dup, sstep := int64(0), step
	d := c.Tx - c.Rx
	switch {
	case d == 0:
	case d > 70000000 || d < -70000000:
		switch {
		case c.Tx%5000 == 0:
			sstep = 0
		case c.Tx%12500 == 0:
			sstep = 4
		case c.Tx%6250 == 0:
			sstep = 1
		case thd74SplitStep(c.Tx) == 2:
			sstep = 2
		default:
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
		}
		dup, offs = 4, c.Tx
	case d > 0:
		dup, offs = 1, d
	default:
		dup, offs = 2, -d
	}
	// modo: 0 FM, 1 FM stretta, 2 AM (resta se la frequenza non cambia; banda aerea)
	mode := int64(0)
	switch {
	case !wasEmpty && oldRx == c.Rx && oldMode == 2, c.Rx < 136000000:
		mode = 2
	case c.Narrow:
		mode = 1
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	fl.F("disabled").SetInt(img, 0)
	r.Fill(img, 0)
	r.F("freq").SetInt(img, c.Rx)
	enc := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	copy(enc, m.cleanName(c.Name))
	f.name(i).SetRaw(img, enc)
	r.F("tone_mode").SetInt(img, tonemode)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dcs)
	r.F("cross_mode").SetInt(img, cross)
	r.F("duplex").SetInt(img, dup)
	r.F("offset").SetInt(img, offs)
	r.F("mode").SetInt(img, mode)
	r.F("tune_step").SetInt(img, step)
	r.F("split_tune_step").SetInt(img, sstep)
	fl.F("prog_vfo").SetInt(img, vfo)
	switch {
	case c.Scan:
		fl.F("skip").SetInt(img, 0)
	case skip != 2:
		fl.F("skip").SetInt(img, 1)
	}
	return nil
}

func (f thd72Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return thd74Download(ctx, m, port, prog)
}

func (f thd72Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return thd74Upload(ctx, m, port, img, prog)
}
