// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// kg935g.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Wouxun KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus
// e KG-XS20H (kg935g.py di CHIRP). Stesso protocollo della KG-UV8D Plus
// (pacchetti cifrati con lo XOR a catena), memoria di 32 KiB; la scrittura invia
// solo le zone conosciute. 999 canali con nomi di 8 caratteri.

import (
	"context"
	"fmt"
	"strings"
)

type kg935Fam struct{}

type kg935Opt struct {
	wx8Opt
	xs20   bool // KG-XS20G/H: potenza su due livelli (0 o 2), trasmissione spenta anche con frequenza 0
	scr    bool // scrambler per canale (HAS_SCRAMBLER)
	fav    bool // canale preferito (HAS_FAVORITE)
	upload [][3]int
}

const kg935MemFmt = `
#seekto 0x0900;
struct {
    u32     rxfreq;
    u32     txfreq;
    u16     rxtone;
    u16     txtone;
    u8      scrambler:4,
            power:4;
    u8      unknown3:2,
            scan_add:1,
            %s:1,
            compander:1,
            mute_mode:2,
            iswide:1;
    u8      unknown5;
    u8      unknown6;
} memory[1000];
#seekto 0x4780;
struct {
    u8    name[8];
    u8    unknown[4];
} names[1000];
#seekto 0x7670;
u8          valid[1000];
`

var (
	kg935Map = [][3]int{{0x44, 32, 1}, {0x440, 8, 1}, {0x480, 8, 5}, {0x500, 8, 15}, {0x580, 8, 15}, {0x600, 8, 5},
		{0x800, 64, 2}, {0x880, 16, 1}, {0x8C0, 16, 1}, {0x900, 64, 250}, {0x4780, 32, 375}, {0x7670, 8, 125}}
	kg935HMap = append([][3]int{{0x14, 1, 1}}, kg935Map...)
	kgXS20Map = [][3]int{{0x900, 64, 250}, {0x4780, 32, 375}, {0x7670, 8, 125}}
)

func init() {
	bands := []band{{30000000, 299999990}, {300000000, 999999990}}
	notes := "Accendi la radio con il cavo collegato (19200 baud). 999 canali con nomi di 8 caratteri; mute, scrambler, " +
		"compander e preferito di ogni canale si conservano. Driver di CHIRP (sperimentale)."
	mk := func(id, model, rclass string, pw []genPower, o *kg935Opt) *genModel {
		if o.xs20 {
			o.wx8Opt = wx8Opt{start: 0x79, enc: true, model: "KG-UV8D-A"}
		} else {
			o.wx8Opt = wx8Opt{start: 0x7C, enc: true, model: "KG-UV8D-B"}
		}
		f := "unknown4"
		if o.fav {
			f = "favorite"
		}
		o.format = fmt.Sprintf(kg935MemFmt, f)
		o.validOff = 0x7670
		return &genModel{ID: id, Vendor: "Wouxun", Model: model, RClass: rclass, Name: "Wouxun " + model, Slots: 999, First: 1,
			NameMax: 8, Charset: charsetASCII, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: dtcsCodes,
			Baud: 19200, Fam: kg935Fam{}, Opt: o, Notes: notes}
	}
	p935 := []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Media 4,5 W", 4.5, 1}, {"Alta 5,5 W", 5.5, 0}}
	p935h := []genPower{{"Bassa 2 W", 2, 2}, {"Media 5 W", 5, 1}, {"Alta 8 W", 8, 0}}
	pxs := []genPower{{"Bassa 5 W", 5, 2}, {"Alta 20 W", 20, 0}}
	genRegister(
		mk("wx-kg935g", "KG-935G", "KG935GRadio", p935, &kg935Opt{scr: true, upload: kg935Map}),
		mk("wx-kg935gplus", "KG-935G Plus", "KG935GPlusRadio", p935, &kg935Opt{scr: true, fav: true, upload: kg935Map}),
		mk("wx-kguv8h", "KG-UV8H", "KGUV8HRadio", p935, &kg935Opt{scr: true, upload: kg935Map}),
		mk("wx-kg935h", "KG-935H", "KG935HRadio", p935h, &kg935Opt{scr: true, fav: true, upload: kg935HMap}),
		mk("wx-kgxs20g", "KG-XS20G", "KGXS20G", pxs, &kg935Opt{xs20: true, fav: true, upload: kgXS20Map}),
		mk("wx-kgxs20gplus", "KG-XS20G Plus", "KGXS20GPlusRadio", pxs, &kg935Opt{xs20: true, fav: true, upload: kgXS20Map}),
		mk("wx-kgxs20h", "KG-XS20H", "KGXS20HRadio", pxs, &kg935Opt{xs20: true, scr: true, fav: true, upload: kgXS20Map}),
	)
}

func kg935o(m *genModel) *kg935Opt { return m.Opt.(*kg935Opt) }

func (kg935Fam) mem(m *genModel) *bwEl      { return bwMust(kg935o(m).format) }
func (f kg935Fam) layout(m *genModel) *bwEl { return f.mem(m) }
func (f kg935Fam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	b := f.mem(m)
	return b.F("memory").I(i + 1), b.F("names").I(i + 1)
}
func (f kg935Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(m, i)
	return append(append(append([]byte{}, r.Raw(img)...), n.Raw(img)...), img[0x7670+i+1])
}
func (f kg935Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(m, i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 29 {
		n.SetRaw(img, raw[16:28])
		img[0x7670+i+1] = raw[28]
	}
}
func (f kg935Fam) Match(m *genModel, img []byte) bool { return false }
func (kg935Fam) crossband() bool                      { return false }

func (f kg935Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := kg935o(m)
	if img[0x7670+i+1] != wx8Valid {
		return nil
	}
	r, n := f.recs(m, i)
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	tx := r.F("txfreq").Int(img)
	if tx == 0xFFFFFFFF || (o.xs20 && tx == 0) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = tx * 10
	}
	var nm []rune
	for _, b := range n.F("name").Raw(img) {
		if b != 0 {
			nm = append(nm, rune(b))
		}
	}
	c.Name = strings.TrimRight(string(nm), " \t\n\r\x0b\x0c")
	c.TxTone = wx8ToneDec(r.F("txtone").Int(img), 0x4000, 0x2000)
	c.RxTone = wx8ToneDec(r.F("rxtone").Int(img), 0x4000, 0x2000)
	c.Scan = r.F("scan_add").Int(img) > 0
	pv := r.F("power").Int(img)
	switch {
	case o.xs20 && pv == 0:
		c.Power = 2
	case o.xs20:
		c.Power = 0
	default:
		pv &= 3
		if int(pv) >= len(m.Powers) {
			pv = int64(len(m.Powers) - 1)
		}
		c.Power = m.powerSem(int(pv))
	}
	c.Narrow = r.F("iswide").Int(img) == 0
	return c
}

func (f kg935Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := kg935o(m)
	r, n := f.recs(m, i)
	if c == nil {
		r.Fill(img, 0)
		img[0x7670+i+1] = 0
		n.Fill(img, 0)
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
	if img[0x7670+i+1] != wx8Valid {
		r.Fill(img, 0)
		n.Fill(img, 0)
	} else if r.F("mute_mode").Int(img) > 2 {
		r.F("mute_mode").SetInt(img, 2) // get_extra: valore oltre l'elenco → QT*DTMF
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0xFFFFFFFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("scan_add").SetBool(img, c.Scan)
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("rxtone").SetInt(img, wx8ToneEnc(rt, 0x4000, 0x2000, 0))
	r.F("txtone").SetInt(img, wx8ToneEnc(tt, 0x4000, 0x2000, 0))
	k := m.powerIndex(c.Power)
	if o.xs20 {
		if k == 0 {
			r.F("power").SetInt(img, 0)
		} else {
			r.F("power").SetInt(img, 2)
		}
	} else {
		r.F("power").SetInt(img, int64(k))
	}
	r.F("unknown3").SetInt(img, 3)
	r.F("unknown5").SetInt(img, 1)
	r.F("unknown6").SetInt(img, 255)
	nm := n.F("name")
	for k := 0; k < 8; k++ {
		v := byte(0)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	img[0x7670+i+1] = wx8Valid
	return nil
}

func (f kg935Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return wx8Fam{}.dl(ctx, m, &kg935o(m).wx8Opt, port, prog)
}

func (f kg935Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := kg935o(m)
	if len(img) < 0x8000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x8000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.wx8IdentO(&o.wx8Opt); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, z := range o.upload {
		for a := z[0]; a < z[0]+z[1]*z[2]; a += z[1] {
			prog(a, 0x8000, "Scrittura…")
			if err := s.wx8Write(&o.wx8Opt, wx8CmdWr, append([]byte{byte(a >> 8), byte(a)}, img[a:a+z[1]]...)); err != nil {
				return "", err
			}
			resp, ok, err := s.wx8Read(&o.wx8Opt)
			if err != nil {
				return "", err
			}
			if !ok || len(resp) < 2 || int(resp[0])<<8|int(resp[1]) != a {
				return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	_ = s.wx8Write(&o.wx8Opt, wx8CmdEnd, nil)
	prog(0x8000, 0x8000, "Scrittura completata")
	return "", nil
}
