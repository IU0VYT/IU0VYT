// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ft70.py, ft1d.py, ft2d.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: palmari Yaesu System Fusion (C4FM): FT-70D, FT1D, FT2D, FT3D e FT5D.
// 900 canali; copia a 38400 baud con il cavo dati USB (vedi drv_yaesuclone.go).
// Sulle FT1D/FT2D/FT3D/FT5D il canale è analogico oppure digitale C4FM (un canale
// digitale nuovo nasce in «AMS», come in CHIRP); sulla FT-70D il modo digitale
// del canale resta quello che c'è (un canale nuovo nasce analogico).
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare viene rifiutata invece di
//     essere troncata al kHz; uno scostamento che non è un numero intero di kHz
//     si scrive come frequenza di trasmissione separata;
//   - modificare un canale non lo toglie dai banchi (CHIRP lo toglie a ogni
//     modifica); cancellarlo sì;
//   - il salto «P» resta quando il canale modificato resta escluso dalla scansione.

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type ysfFam struct{}

type ysfOpt struct {
	clone  *yaesuOpt
	format string
	ft70   bool // FT-70D: nome di 6 caratteri, nessun modo digitale noto
	coded  bool // FT1D: i nomi sono scritti con i codici dei caratteri
}

const ysfFlags = `
struct flagslot {
  u8 nosubvfo:1,
     unknown:3,
     pskip:1,
     skip:1,
     used:1,
     valid:1;
};
`

const ft1dFormat = ysfFlags + `
struct memslot {
  u8 unknown0:2,
     mode_alt:1,
     clock_shift:1,
     unknown1:4;
  u8 mode:2,
     duplex:2,
     tune_step:4;
  bbcd freq[3];
  u8 power:2,
     digmode:2,
     tone_mode:4;
  u16 charsetbits;
  char label[16];
  bbcd offset[3];
  u8 unknown5:2,
     tone:6;
  u8 unknown6:1,
     dcs:7;
  u16 unknown7;
  u8 unknown8:2,
     att:1,
     autostep:1,
     automode:1,
     unknown9:3;
};
#seekto 0x054A;
struct {
    u16 in_use;
} bank_used[24];
#seekto 0x064A;
struct {
  u8 unknown0[4];
  u8 frequency_band;
  u8 unknown1:6,
     manual_or_mr:2;
  u8 unknown2:7,
     mr_banks:1;
  u8 unknown3;
  u16 mr_index;
  u16 bank_index;
  u16 bank_enable;
  u8 unknown4[5];
  u8 unknown5:6,
     power:2;
  u8 unknown6:4,
     tune_step:4;
  u8 unknown7:6,
     duplex:2;
  u8 unknown8:6,
     tone_mode:2;
  u8 unknown9:2,
     tone:6;
  u8 unknown10;
  u8 unknown11:6,
     mode:2;
  bbcd freq0[4];
  bbcd offset_freq[4];
  u8 unknown12[2];
  char label[16];
  u8 unknown13[6];
  bbcd band_lower[4];
  bbcd band_upper[4];
  bbcd rx_freq[4];
  u8 unknown14[22];
  bbcd freq1[4];
  u8 unknown15[11];
  u8 unknown16:3,
     volume:5;
  u8 unknown17[18];
  u8 active_menu_item;
  u8 checksum;
} vfo_info[6];
#seekto 0x154A;
struct {
    u16 channel[100];
} bank_members[24];
struct flagslot flag[900];
#seekto 0x2D4A;
struct memslot memory[900];
`

const ft70Format = `
#seekto 0x54a;
struct {
  u16 in_use;
} bank_used[24];
struct bank_aux {
  u8 unknown1;
  u8 max_bank;
  u8 unknown2;
  u8 unknown3;
};
#seekto 0x654;
struct bank_aux bank_aux1;
#seekto 0x6D4;
struct bank_aux bank_aux2;
#seekto 0x154a;
struct {
  u16 channel[100];
} bank_members[24];
struct {
  u8 nosubvfo:1,
     unknown:3,
     pskip:1,
     skip:1,
     used:1,
     valid:1;
} flag[900];
#seekto 0x2D4A;
struct {
  u8 display_tag:1,
     unknown0:1,
     deviation:1,
     clock_shift:1,
     unknown1:4;
  u8 mode:2,
     duplex:2,
     tune_step:4;
  bbcd freq[3];
  u8 power:2,
     unknown2:1,
     ams:1,
     tone_mode:4;
  u8 charsetbits[2];
  char label[6];
  char unknown7[10];
  bbcd offset[3];
  u8 unknown5:2,
     tone:6;
  u8 unknown6:1,
     dcs:7;
  u8 unknown9;
  u8 ams_on_dn_vw_fm:2,
     unknown8_3:1,
     unknown8_4:1,
     smeter:4;
  u8 unknown10:2,
     att:1,
     auto_step:1,
     auto_mode:1,
     unknown11:2,
     bell:1;
} memory[900];
`

// caratteri della FT1D nell'ordine dei codici
const ft1dChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz.,:;*#_-/&()@!?^ "

var ft1dSums = [][3]int{{0x064A, 0x06C8, 0x06C9}, {0x06CA, 0x0748, 0x0749}, {0x074A, 0x07C8, 0x07C9}, {0x07CA, 0x0848, 0x0849},
	{0x0000, 0x1FDC9, 0x1FDCA}}

var ysfTouchNote = yaesuCloneHow("DATA", "[DISP]", "[SEND]", "[RECEIVE]")

const ysfNote = "Toni ammessi: subtono in trasmissione o nei due versi, lo stesso codice DCS nei due versi. Driver di CHIRP."

func init() {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz.,:;*#_-/&()@!?^"
	p4 := []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W (L3)", 2.5, 1}, {"Bassa 1 W (L2)", 1, 2}, {"Minima (L1)", 0.05, 3}}
	bands := []band{{500000, 999900000}}
	mk := func(id, model, variant, rclass, name, ident string, coded bool, send, recv, how string) *genModel {
		return &genModel{ID: id, Vendor: "Yaesu", Model: model, Variant: variant, RClass: rclass, Name: name,
			Slots: 900, First: 1, NameMax: 16, Charset: chars, Powers: p4, Bands: bands, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: dtcsCodes, Baud: 38400, Fam: ysfFam{}, Digital: "C4FM",
			Opt: &ysfOpt{clone: &yaesuOpt{model: ident, size: 130507, head: 10, chunk: 32, sums: ft1dSums, send: send, recv: recv},
				format: ft1dFormat, coded: coded},
			Notes: "900 canali con nomi di 16 caratteri, analogici o digitali C4FM (un canale digitale nuovo nasce in «AMS»), cavo dati USB, 38400 baud. " +
				how + ysfNote}
	}
	touch := func(id, model, variant, rclass, name, ident string) *genModel {
		return mk(id, model, variant, rclass, name, ident, false, "[SEND]", "[RECEIVE] («-WAIT-»)", ysfTouchNote)
	}
	genRegister(
		&genModel{ID: "ys-ft70d", Vendor: "Yaesu", Model: "FT-70D", RClass: "FT70Radio", Name: "Yaesu FT-70DR / FT-70DE",
			Slots: 900, First: 1, NameMax: 6, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ .,:;*#_-/&()@!?^",
			Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2 W", 2, 1}, {"Bassa 0,5 W", 0.5, 2}},
			Bands:  bands, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 38400, Fam: ysfFam{},
			Opt: &ysfOpt{clone: &yaesuOpt{model: "AH51G", size: 65227, head: 10, chunk: 32, sums: [][3]int{{0x0000, 0xFEC9, 0xFECA}},
				send: "[BAND]", recv: "[MODE] («-WAIT-»)"}, format: ft70Format, ft70: true},
			Notes: "900 canali con nomi di 6 caratteri, cavo dati USB, 38400 baud. La copia parte dalla radio: stacca la batteria e rimettila tenendo premuti " +
				"[AMS] e il tasto di accensione (compare «ADMS»); per leggere premi [BAND] dopo aver avviato la lettura, per scrivere premi [MODE] («-WAIT-») " +
				"prima di avviare la scrittura. Un canale nuovo nasce in FM analogica. " + ysfNote},
		mk("ys-ft1d", "FT-1D", "R", "FT1Radio", "Yaesu FT1DR / FT1DE", "AH44M", true, "[BAND]", "[Dx] («-WAIT-»)",
			yaesuCloneHow("DATA", "[F]", "[BAND]", "[Dx]")),
		touch("ys-ft2d", "FT2D", "R", "FT2D", "Yaesu FT2DR / FT2DE", "AH60M"),
		touch("ys-ft2dv2", "FT2D", "Rv2", "FT2Dv2", "Yaesu FT2DR / FT2DE (seconda serie)", "AH60G"),
		touch("ys-ft3d", "FT3D", "R", "FT3D", "Yaesu FT3DR / FT3DE", "AH72M"),
		touch("ys-ft5d", "FT5D", "R", "FT5D", "Yaesu FT5DR / FT5DE", "AH82M"),
	)
}

func (ysfFam) opt(m *genModel) *ysfOpt         { return m.Opt.(*ysfOpt) }
func (f ysfFam) yaesu(m *genModel) *yaesuOpt   { return f.opt(m).clone }
func (f ysfFam) mem(m *genModel) *bwEl         { return bwMust(f.opt(m).format) }
func (f ysfFam) layout(m *genModel) *bwEl      { return f.mem(m) }
func (f ysfFam) rec(m *genModel, i int) *bwEl  { return f.mem(m).F("memory").I(i) }
func (f ysfFam) flag(m *genModel, i int) *bwEl { return f.mem(m).F("flag").I(i) }

func (f ysfFam) narrow(m *genModel) string {
	if f.opt(m).ft70 {
		return "deviation"
	}
	return "mode_alt"
}

func (f ysfFam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(m, i).Raw(img)...), f.flag(m, i).Raw(img)...)
}

func (f ysfFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(m, i).ByteLen()
	if len(raw) != n+1 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:n])
	fl := f.flag(m, i)
	fl.SetRaw(img, raw[n:])
	fl.F("used").SetBool(img, true)
	fl.F("valid").SetBool(img, true)
}

// Match: FT1D, FT2D, FT3D e FT5D hanno la memoria della stessa grandezza e si
// distinguono dai primi 5 byte
func (f ysfFam) Match(m *genModel, img []byte) bool { return yaesuMatch(f.opt(m).clone, img) }

func (ysfFam) noIdent()  {}
func (ysfFam) fitBySet() {}

func (ysfFam) toneOK(t, r genTone) bool { k := ft60Tmode(t, r); return k >= 0 && k <= 4 }

func (f ysfFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !f.toneOK(t, r)
}

func (f ysfFam) duplex(img []byte, r *bwEl) (tx int64, kept int64) {
	rx := yaesuRound(r.F("freq").Int(img) * 1000)
	kept = r.F("offset").Int(img) * 1000
	switch r.F("duplex").Int(img) {
	case 1:
		return rx - kept, kept
	case 2:
		return rx + kept, kept
	case 3:
		kept = yaesuRound(kept)
		return kept, kept
	}
	return rx, kept
}

func (f ysfFam) Get(m *genModel, img []byte, i int) *Chan {
	fl, r, o := f.flag(m, i), f.rec(m, i), f.opt(m)
	if !fl.F("used").Bool(img) || !fl.F("valid").Bool(img) {
		return nil
	}
	c := &Chan{Rx: yaesuRound(r.F("freq").Int(img) * 1000)}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx, _ = f.duplex(img, r)
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img)); k < len(chirpTones) {
		tone = fmtTone(chirpTones[k])
	}
	if k := int(r.F("dcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	mode := r.F("mode").Int(img)
	if !o.ft70 {
		c.Dig = mode == 0 && r.F("digmode").Int(img) != 0
	}
	if !c.Dig {
		switch r.F("tone_mode").Int(img) {
		case 1:
			c.TxTone = tone
		case 2:
			c.TxTone, c.RxTone = tone, tone
		case 3:
			c.TxTone, c.RxTone = dcs, dcs
		}
	}
	p := int(3 - r.F("power").Int(img))
	if p >= len(m.Powers) {
		p = len(m.Powers) - 1
	}
	c.Power = m.powerSem(p)
	c.Narrow = mode == 0 && !c.Dig && r.F(f.narrow(m)).Bool(img)
	c.Scan = !fl.F("skip").Bool(img) && !fl.F("pskip").Bool(img)
	var sb strings.Builder
	for _, b := range r.F("label").Raw(img) {
		if b == 0xFF {
			break
		}
		switch {
		case o.coded:
			if int(b) < len(ft1dChars) {
				sb.WriteByte(ft1dChars[b])
			}
		case b >= 0x20 && b < 0x7F:
			sb.WriteByte(b)
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

// debank: toglie il canale dai banchi (come CHIRP quando il canale viene cancellato)
func (f ysfFam) debank(m *genModel, img []byte, i int) {
	e, o := f.mem(m), f.opt(m)
	changed := false
	members := func(b int) []int {
		if e.F("bank_used").I(b).F("in_use").Int(img) == 0xFFFF {
			return nil
		}
		var out []int
		for _, x := range e.F("bank_members").I(b).F("channel").Elems {
			if v := int(x.Int(img)); v != 0xFFFF {
				out = append(out, v)
			}
		}
		sort.Ints(out)
		return out
	}
	for b := 0; b < 24; b++ {
		ch := members(b)
		k := sort.SearchInts(ch, i)
		if k >= len(ch) || ch[k] != i {
			continue
		}
		ch = append(ch[:k], ch[k+1:]...)
		for j, x := range e.F("bank_members").I(b).F("channel").Elems {
			v := int64(0xFFFF)
			if j < len(ch) {
				v = int64(ch[j])
			}
			x.SetInt(img, v)
		}
		if len(ch) == 0 {
			e.F("bank_used").I(b).F("in_use").SetInt(img, 0xFFFF)
		}
		changed = true
	}
	if !changed {
		return
	}
	if o.ft70 {
		maxBank := -1
		for b := 0; b < 24; b++ {
			if e.F("bank_used").I(b).F("in_use").Int(img) != 0xFFFF {
				maxBank = b
			}
		}
		for _, name := range []string{"bank_aux1", "bank_aux2"} {
			if maxBank < 0 {
				e.F(name).Fill(img, 0xFF)
			} else {
				e.F(name).Fill(img, 0)
				e.F(name).F("max_bank").SetInt(img, int64(maxBank))
			}
		}
		return
	}
	// FT1D e successive: il banco scelto nei due VFO (come update_vfo di CHIRP)
	bank, mr := [2]int{-1, -1}, [2]int{-1, -1}
	for b := 0; b < 24 && bank[1] < 0; b++ {
		for _, ch := range members(b) {
			n := ch + 1
			bank[0], mr[0] = b, n
			if n&0x7000 != 0 {
				break
			}
			if ch < 900 && !f.flag(m, ch).F("nosubvfo").Bool(img) {
				bank[1], mr[1] = b, n
				break
			}
		}
	}
	for v := 0; v < 2; v++ {
		vfo, bak := e.F("vfo_info").I(v*2), e.F("vfo_info").I(v*2+1)
		if vfo.F("checksum").Int(img) != bak.F("checksum").Int(img) {
			continue
		}
		switch has := vfo.F("bank_index").Int(img) != 0xFFFF; {
		case bank[v] < 0 && has:
			vfo.F("bank_index").SetInt(img, 0xFFFF)
			vfo.F("mr_index").SetInt(img, 0xFFFF)
			vfo.F("bank_enable").SetInt(img, 0xFFFF)
		case bank[v] >= 0 && !has:
			vfo.F("bank_index").SetInt(img, int64(bank[v]))
			vfo.F("mr_index").SetInt(img, int64(mr[v]))
			vfo.F("bank_enable").SetInt(img, 0)
		}
		for _, k := range []string{"bank_index", "mr_index", "bank_enable"} {
			bak.F(k).SetInt(img, vfo.F(k).Int(img))
		}
	}
}

func (f ysfFam) wipe(img []byte, r *bwEl) {
	r.Fill(img, 0)
	r.F("unknown1").SetInt(img, 5)
}

func (f ysfFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	fl, r, o := f.flag(m, i), f.rec(m, i), f.opt(m)
	valid, used := fl.F("valid").Bool(img), fl.F("used").Bool(img)
	if c == nil {
		f.debank(m, img, i)
		if o.ft70 {
			if !(valid && !used) {
				fl.F("used").SetBool(img, false)
			}
			fl.F("valid").SetBool(img, false)
			return nil
		}
		f.wipe(img, r)
		fl.F("used").SetBool(img, false)
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
	tmode := ft60Tmode(t, rt)
	if tmode < 0 || tmode > 4 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
	}
	if tmode == 4 {
		tmode = 3
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, passo 5 kHz)
	was := valid && (used || o.ft70)
	kept, tone, dcs, step, oldRx, oldMode := int64(0), int64(8), int64(0), int64(0), int64(0), int64(0)
	if o.ft70 {
		kept = 600000
	}
	if was {
		_, kept = f.duplex(img, r)
		tone, dcs, step = r.F("tone").Int(img), r.F("dcs").Int(img), r.F("tune_step").Int(img)
		oldRx, oldMode = yaesuRound(r.F("freq").Int(img)*1000), r.F("mode").Int(img)
	}
	if (o.ft70 && step > 9) || step > 10 {
		step = 0
	}
	find := func(list []int, v int) (int64, error) {
		for k, x := range list {
			if x == v {
				return int64(k), nil
			}
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	if t.Mode == "Tone" {
		if tone, err = find(chirpTones, t.Hz10); err != nil {
			return err
		}
	}
	if t.Mode == "DTCS" {
		if dcs, err = find(dtcsCodes, t.Code); err != nil {
			return err
		}
	}
	dup, offs := int64(0), kept/1000
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	switch {
	case d == 0:
	case ad > 70000000 || ad%1000 != 0:
		tk, ok := yaesuKHz(c.Tx)
		if !ok {
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
		}
		dup, offs = 3, tk
	case d > 0:
		dup, offs = 2, ad/1000
	default:
		dup, offs = 1, ad/1000
	}
	// modo: resta quello del canale se la frequenza non cambia, altrimenti segue la
	// frequenza (0 FM, 1 AM, 2 FM larga; sulla FT-70D il terzo modo non è noto)
	mode := []int64{0, 2, 1}[icx90Mode(c.Rx)]
	if o.ft70 && mode == 2 {
		mode = 0
	}
	if was && oldRx == c.Rx && oldMode != 0 {
		mode = oldMode
	}
	if c.Dig {
		mode = 0
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if o.ft70 && !valid {
		f.wipe(img, r)
	}
	fl.F("used").SetBool(img, true)
	fl.F("valid").SetBool(img, true)
	fl.F("nosubvfo").SetBool(img, c.Rx < 30000000 || c.Rx > 88000000 && c.Rx < 108000000 || c.Rx > 580000000)
	r.F("power").SetInt(img, 3-int64(m.powerIndex(c.Power)))
	r.F("tone").SetInt(img, tone)
	r.F("tone_mode").SetInt(img, tmode)
	r.F("dcs").SetInt(img, dcs)
	r.F("tune_step").SetInt(img, step)
	r.F("duplex").SetInt(img, dup)
	r.F(f.narrow(m)).SetBool(img, c.Narrow && mode == 0 && !c.Dig)
	if !o.ft70 {
		switch dm := r.F("digmode").Int(img); {
		case !c.Dig:
			r.F("digmode").SetInt(img, 0)
		case dm == 0 || !was:
			r.F("digmode").SetInt(img, 1) // canale digitale nuovo: AMS
		}
	}
	r.F("mode").SetInt(img, mode)
	wasSkip, wasP := was && used && fl.F("skip").Bool(img), was && used && fl.F("pskip").Bool(img)
	switch {
	case c.Scan:
		fl.F("skip").SetBool(img, false)
		fl.F("pskip").SetBool(img, false)
	case wasP && !wasSkip:
		// il salto «P» resta
	default:
		fl.F("skip").SetBool(img, true)
		fl.F("pskip").SetBool(img, false)
	}
	r.F("freq").SetInt(img, fk)
	r.F("offset").SetInt(img, offs)
	name := strings.TrimRight(m.cleanName(c.Name), " ")
	lab := r.F("label")
	enc := make([]byte, lab.ByteLen())
	for k := range enc {
		enc[k] = 0xFF
		if k < len(name) {
			enc[k] = name[k]
			if o.coded {
				p := strings.IndexByte(ft1dChars, name[k])
				if p < 0 {
					p = strings.IndexByte(ft1dChars, ' ')
				}
				enc[k] = byte(p)
			}
		}
	}
	lab.SetRaw(img, enc)
	if o.ft70 {
		r.F("charsetbits").Fill(img, 0)
		r.F("display_tag").SetInt(img, 1)
	} else {
		r.F("charsetbits").SetInt(img, 0)
	}
	return nil // le somme di controllo si aggiornano quando si scrive nella radio (come CHIRP)
}

func (f ysfFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := yaesuDownload(ctx, m, port, f.opt(m).clone, prog)
	return img, "", err
}

func (f ysfFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", yaesuUpload(ctx, m, port, f.opt(m).clone, img, prog)
}
