// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// wouxun.py, wouxun_common.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Wouxun KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D/X, KG-816, KG-818 e KG-805G
// (wouxun.py di CHIRP). Presentazione con «HiWOUXUN» (o le altre del modello):
// la radio risponde con «KG669V»; poi 02 06 → 06 e i blocchi 'R'/'W' del
// protocollo Wouxun (lettura a 64 byte, scrittura a 16 byte, 8 KiB).

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type wxoFam struct{}

type wxoOpt struct {
	query   []string // presentazioni (_querymodels), provate a turno
	format  string
	split   bool // bit «splitdup» (tutti tranne la KG-805G)
	nameRec int  // posizione del nome nel record dei nomi (8 per la KG-UVD1P)
	match   func(img []byte) bool
	ranges  int // posizione dei limiti di frequenza (freq_ranges), 0 = fissi
	fixed   []band
}

const wxoMemFmt = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 _3_unknown_1:4,
     bcl:1,
     _3_unknown_2:3;
  u8 splitdup:1,
     skip:1,
     power_high:1,
     iswide:1,
     _2_unknown_2:4;
  u8 unknown;
  u8 _0_unknown_1:3,
     iswidex:1,
     _0_unknown_2:4;
} memory[%d];
`

var (
	wxoFmtUVD1P = fmt.Sprintf(wxoMemFmt, 199) + `
#seekto 0x0970;
struct {
    u16 vhf_rx_start;
    u16 vhf_rx_stop;
    u16 uhf_rx_start;
    u16 uhf_rx_stop;
    u16 vhf_tx_start;
    u16 vhf_tx_stop;
    u16 uhf_tx_start;
    u16 uhf_tx_stop;
} freq_ranges;
#seekto 0x1008;
struct {
  u8 unknown[8];
  u8 name[6];
  u8 pad[2];
} names[199];
`
	wxoFmtUV6 = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 _3_unknown_1:4,
     bcl:1,
     _3_unknown_2:3;
  u8 splitdup:1,
     skip:1,
     power_high:1,
     iswide:1,
     _2_unknown_2:4;
  u8 pad;
  u8 _0_unknown_1:3,
     iswidex:1,
     _0_unknown_2:4;
} memory[199];
#seekto 0x0ff0;
struct {
    u16 vhf_rx_start;
    u16 vhf_rx_stop;
    u16 uhf_rx_start;
    u16 uhf_rx_stop;
    u16 vhf_tx_start;
    u16 vhf_tx_stop;
    u16 uhf_tx_start;
    u16 uhf_tx_stop;
} freq_ranges;
#seekto 0x1010;
struct {
  u8 name[6];
  u8 pad[10];
} names[199];
`
	wxoFmt816 = fmt.Sprintf(wxoMemFmt, 199) + `
#seekto 0x0d70;
struct {
    u16 vhf_rx_start;
    u16 vhf_rx_stop;
    u16 uhf_rx_start;
    u16 uhf_rx_stop;
    u16 vhf_tx_start;
    u16 vhf_tx_stop;
    u16 uhf_tx_start;
    u16 uhf_tx_stop;
} freq_ranges;
#seekto 0x1010;
struct {
    u8 name[6];
    u8 pad[10];
} names[199];
`
	wxoFmt805G = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 _3_unknown_1:4,
     bcl:1,
     _3_unknown_2:3;
  u8 _2_unknown_1:1,
     skip:1,
     power_high:1,
     iswide:1,
     _2_unknown_2:4;
  u8 unknown;
  u8 _0_unknown_1:3,
     iswidex:1,
     _0_unknown_2:4;
} memory[128];
#seekto 0x1010;
struct {
    u8 name[6];
    u8 pad[10];
} names[128];
`
)

const wxoCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ?+-"

// frequenze dei limiti di banda (encode_freq / decode_freq di wouxun.py): 4 cifre
// decimali codificate con una tabella
var wxoFreqTable = []int{0x7, 0xa, 0x0, 0x9, 0xb, 0x2, 0xe, 0x1, 0x3, 0xf}

func wxoDecodeFreq(v int64) int64 {
	f := int64(0)
	for sh := 12; sh >= 0; sh -= 4 {
		d := int((v >> sh) & 0xF)
		k := idxInt(wxoFreqTable, d)
		if k < 0 {
			return -1
		}
		f = f*10 + int64(k)
	}
	return f
}

func init() {
	ff := bytes.Repeat([]byte{0xFF}, 16)
	notes := "Spegni la radio, collega il cavo, riaccendila e mettila su un canale senza segnali. "
	pw := []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}}
	mk := func(id, model, rclass, name string, slots int, o *wxoOpt, bands []band, note string) *genModel {
		o.fixed = bands
		return &genModel{ID: id, Vendor: "Wouxun", Model: model, RClass: rclass, Name: "Wouxun " + name, Slots: slots, First: 1,
			NameMax: 6, Charset: wxoCharset, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: wxoFam{}, Opt: o, Notes: notes + note}
	}
	genRegister(
		mk("wx-kguvd1p", "KG-UVD1P", "KGUVD1PRadio", "KG-UVD1P / KG-UV2D / KG-UV3D", 128,
			&wxoOpt{query: []string{"HiWOUXUN\x02", "PROGUV6X\x02"}, format: wxoFmtUVD1P, split: true, nameRec: 8, ranges: 0x0970,
				match: func(d []byte) bool {
					return len(d) == 8192 && string(d[0x60:0x64]) != "2009" && string(d[0x170:0x173]) != "LX-" &&
						bytes.Equal(d[0x1f77:0x1f7d], ff[:6]) && bytes.Equal(d[0x0d70:0x0d80], ff)
				}},
			[]band{{136000000, 175000000}, {216000000, 520000000}},
			"128 canali con nomi di 6 caratteri (cifre, lettere maiuscole, ? + -). Driver di CHIRP (sperimentale)."),
		mk("wx-kguv6", "KG-UV6", "KGUV6DRadio", "KG-UV6D / KG-UV6X", 199,
			&wxoOpt{query: []string{"HiWXUVD1\x02", "HiKGUVD1\x02"}, format: wxoFmtUV6, split: true, ranges: 0x0ff0,
				match: func(d []byte) bool { return len(d) == 8192 && string(d[0x1f77:0x1f7d]) == "WELCOM" }},
			[]band{{136000000, 175000000}, {216000000, 520000000}},
			"199 canali con nomi di 6 caratteri. Driver di CHIRP (sperimentale)."),
		mk("wx-kg816", "KG-816", "KG816Radio", "KG-816", 199,
			&wxoOpt{query: []string{"HiWOUXUN\x02"}, format: wxoFmt816, split: true, ranges: 0x0d70,
				match: func(d []byte) bool {
					return len(d) == 8192 && string(d[0x60:0x64]) != "2009" && string(d[0x170:0x173]) != "LX-" &&
						!bytes.Equal(d[0xF7E:0xF80], []byte{0x01, 0xE2}) && bytes.Equal(d[0x1f77:0x1f7d], ff[:6]) &&
						!bytes.Equal(d[0x0d70:0x0d80], ff)
				}},
			[]band{{400000000, 480000000}},
			"199 canali con nomi di 6 caratteri; gamma di frequenza letta dalla radio. Driver di CHIRP (sperimentale)."),
		mk("wx-kg818", "KG-818", "KG818Radio", "KG-818", 199,
			&wxoOpt{query: []string{"HiWOUXUN\x02"}, format: wxoFmt816, split: true, ranges: 0x0d70},
			[]band{{66000000, 520000000}},
			"199 canali con nomi di 6 caratteri; gamma di frequenza letta dalla radio. Driver di CHIRP (sperimentale)."),
		mk("wx-kg805g", "KG-805G", "KG805GRadio", "KG-805G", 128,
			&wxoOpt{query: []string{"HiWOUXUN\x02"}, format: wxoFmt805G},
			[]band{{400000000, 470987500}},
			"128 canali con nomi di 6 caratteri (radio GMRS). Driver di CHIRP (sperimentale)."),
	)
}

func wxoo(m *genModel) *wxoOpt { return m.Opt.(*wxoOpt) }

func (wxoFam) mem(m *genModel) *bwEl          { return bwMust(wxoo(m).format) }
func (f wxoFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f wxoFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f wxoFam) name(m *genModel, i int) *bwEl {
	return f.mem(m).F("names").I(i).F("name")
}
func (f wxoFam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(m, i).Raw(img)...), f.name(m, i).Raw(img)...)
}
func (f wxoFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(m, i).SetRaw(img, raw[:16])
	if len(raw) >= 22 {
		f.name(m, i).SetRaw(img, raw[16:22])
	}
}
func (f wxoFam) Match(m *genModel, img []byte) bool {
	if mt := wxoo(m).match; mt != nil {
		return mt(img)
	}
	return false
}
func (wxoFam) crossband() bool { return false }

// Bands: gamme di ricezione scritte nella radio (valid_freq ricavata da get_settings)
func (f wxoFam) Bands(m *genModel, img []byte) []band {
	o := wxoo(m)
	if o.ranges == 0 || len(img) < 0x2000 {
		return o.fixed
	}
	fr := f.mem(m).F("freq_ranges")
	vl, vh := wxoDecodeFreq(fr.F("vhf_rx_start").Int(img)), wxoDecodeFreq(fr.F("vhf_rx_stop").Int(img))
	if vl < 0 || vh < 0 {
		return o.fixed
	}
	out := []band{{vl * 1000000, (vh+1)*1000000 - 1}}
	if m.ID == "wx-kg816" || m.ID == "wx-kg818" {
		return out
	}
	ul, uh := wxoDecodeFreq(fr.F("uhf_rx_start").Int(img)), wxoDecodeFreq(fr.F("uhf_rx_stop").Int(img))
	if ul < 0 || uh < 0 {
		return o.fixed
	}
	return append(out, band{ul * 1000000, (uh+1)*1000000 - 1})
}

func (f wxoFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	raw := r.Raw(img)
	if bytes.Equal(raw, bytes.Repeat([]byte{0xFF}, 16)) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	txRaw := r.F("tx_freq").Raw(img)
	switch {
	case bytes.Equal(txRaw, []byte{0xFF, 0xFF, 0xFF, 0xFF}) && !(wxoo(m).split && r.F("splitdup").Int(img) > 0):
		c.TxOff, c.Tx = true, c.Rx
	default:
		c.Tx = r.F("tx_freq").Int(img) * 10
	}
	c.Scan = r.F("skip").Int(img) > 0
	c.Narrow = r.F("iswide").Int(img) == 0
	c.TxTone = kenwoodToneDec(r.F("txtone").Int(img), 0x2800, 0x8000)
	c.RxTone = kenwoodToneDec(r.F("rxtone").Int(img), 0x2800, 0x8000)
	if r.F("power_high").Int(img) > 0 {
		c.Power = 0
	} else {
		c.Power = 2
	}
	var nm []byte
	for _, b := range f.name(m, i).Raw(img) {
		if b == 0xFF || int(b) >= len(wxoCharset) {
			break
		}
		nm = append(nm, wxoCharset[b])
	}
	c.Name = string(nm)
	return c
}

func (f wxoFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF) // il nome resta (wipe_memory solo del record)
		return nil
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	var nm []byte
	for k := 0; k < 6; k++ {
		if k < len(c.Name) {
			x := bytes.IndexByte([]byte(wxoCharset), c.Name[k])
			if x < 0 {
				return fmt.Errorf("carattere %q non ammesso nel nome", c.Name[k])
			}
			nm = append(nm, byte(x))
		} else {
			nm = append(nm, 0xFF)
		}
	}
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, 16)) {
		r.Fill(img, 0x00)
	}
	r.F("rx_freq").SetInt(img, c.Rx/10)
	split := wxoo(m).split && abs64(c.Tx-c.Rx) > 70000000
	switch {
	case c.TxOff:
		r.F("tx_freq").Fill(img, 0xFF)
	default:
		r.F("tx_freq").SetInt(img, c.Tx/10)
	}
	if wxoo(m).split {
		r.F("splitdup").SetBool(img, split && !c.TxOff)
	}
	r.F("skip").SetBool(img, c.Scan)
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("rxtone").SetInt(img, kenwoodToneEnc(rt, 0x2800, 0x8000, 0xFFFF))
	r.F("txtone").SetInt(img, kenwoodToneEnc(tt, 0x2800, 0x8000, 0xFFFF))
	r.F("power_high").SetBool(img, m.powerIndex(c.Power) == 0)
	f.name(m, i).SetRaw(img, nm)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) wxoIdent(m *genModel) error {
	q := wxoo(m).query
	var resp []byte
	for k := 0; k < 10; k++ {
		if err := s.send([]byte(q[k%len(q)])); err != nil {
			return err
		}
		resp = readUpTo(s.ctx, s.p, 9, 1500*time.Millisecond)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(resp) != 9 {
			if err := s.sleep(time.Second); err != nil {
				return err
			}
			continue
		}
		if string(resp[2:8]) != "KG669V" {
			return rErr("la radio risponde ma non è di questo modello (% X)", resp)
		}
		if err := s.send([]byte{0x02, 0x06}); err != nil {
			return err
		}
		if err := s.sleep(50 * time.Millisecond); err != nil {
			return err
		}
		if b := readUpTo(s.ctx, s.p, 1, 1500*time.Millisecond); len(b) != 1 || b[0] != 0x06 {
			return rErr("la radio non è entrata in modalità di programmazione")
		}
		return nil
	}
	if len(resp) == 0 {
		return rErr("la radio non risponde")
	}
	return rErr("la radio non si è identificata")
}

func (f wxoFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.wxoIdent(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, 0x2000, 0x40, 1500*time.Millisecond, prog)
	return data, "", err
}

func (f wxoFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.wxoIdent(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img, 0, 0x2000, 0x10, 1500*time.Millisecond, prog)
}
