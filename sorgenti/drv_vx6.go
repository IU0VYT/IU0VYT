// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// vx6.py, vx3.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: Yaesu VX-6R/E e VX-3R/E: 999 canali con nomi di 6 caratteri, ricevitore
// a larga banda. Copia a 19200 baud (vedi drv_yaesuclone.go). Il modo (FM, AM, FM
// larga delle radio commerciali) non è una colonna del programma: resta quello
// del canale se la frequenza non cambia, altrimenti segue la frequenza.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare viene rifiutata invece di
//     essere troncata al kHz;
//   - VX-6, DCS in trasmissione con subtono in ricezione: il subtono scritto è
//     quello di ricezione (CHIRP lascia il valore precedente);
//   - uno scostamento che la radio non può memorizzare si scrive come frequenza
//     di trasmissione separata;
//   - il canale «preferenziale» (VX-6) o con il salto «P» (VX-3) resta tale
//     quando lo si modifica.

import (
	"context"
	"fmt"
	"strings"
)

type vxFam struct{}

type vxOpt struct {
	clone  *yaesuOpt
	format string
	chars  string // caratteri nell'ordine dei codici (0 = codice senza carattere)
	narrow string // campo della deviazione ridotta
	vx3    bool   // VX-3: due potenze, quattro modi dei toni, canale nuovo azzerato
}

const vx6Format = `
#seekto 0x1ECA;
struct {
  u8 even_pskip:1,
     even_skip:1,
     even_valid:1,
     even_masked:1,
     odd_pskip:1,
     odd_skip:1,
     odd_valid:1,
     odd_masked:1;
} flags[500];

#seekto 0x21CA;
struct {
  u8 unknown11:1,
     step_changed:1,
     half_deviation:1,
     cpu_shifted:1,
     unknown12:4;
  u8 mode:2,
     duplex:2,
     tune_step:4;
  bbcd freq[3];
  u8 power:2,
     unknown2:2,
     pager:1,
     tmode:3;
  u8 name[6];
  bbcd offset[3];
  u8 tone;
  u8 dcs;
  u8 unknown5;
} memory[999];
`

const vx3Format = `
#seekto 0x20CA;
struct {
  u8 even_pskip:1,
     even_skip:1,
     even_valid:1,
     even_masked:1,
     odd_pskip:1,
     odd_skip:1,
     odd_valid:1,
     odd_masked:1;
} flags[999];

#seekto 0x244A;
struct {
  u8   unknown1a:2,
       txnarrow:1,
       clockshift:1,
       unknown1b:4;
  u8   mode:2,
       duplex:2,
       tune_step:4;
  bbcd freq[3];
  u8   power:2,
       unknown2:4,
       tmode:2;
  u8   name[6];
  bbcd offset[3];
  u8   unknown3:2,
       tone:6;
  u8   unknown4:1,
       dcs:7;
  u8   unknown5;
  u8   smetersquelch;
  u8   unknown7a:2,
       attenuate:1,
       unknown7b:1,
       automode:1,
       unknown8:1,
       bell:2;
} memory[999];
`

var (
	vx6Clone = &yaesuOpt{model: "AH021", size: 32587, head: 10, chunk: 16, sums: [][3]int{{0x0000, 0x7F49, 0x7F4A}},
		send: "[BAND]", recv: "[V/M] («-WAIT-»)"}
	vx3Clone = &yaesuOpt{model: "AH028", size: 32587, head: 10, chunk: 32, sums: [][3]int{{0x0000, 0x7F49, 0x7F4A}},
		send: "[BAND]", recv: "[V/M] («-WAIT-»)"}
)

// caratteri delle radio nell'ordine dei codici; «u» e «o» sono simboli che non si
// scrivono da qui
const (
	vx6Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ +-/?[]u_\x00\x00\x00\x00\x00\x00\x00\x00\x00$%%\x00*#.|=\\o@"
	vx3Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ +-/\x00[](){}\x00\x00_\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00*\x00\x00,'|"
)

// yaesuCloneHow: istruzioni della copia (presa del cavo, tasto da tenere premuto
// all'accensione, tasto che avvia l'invio, tasto che mette la radio in ricezione)
func yaesuCloneHow(jack, hold, send, recv string) string {
	return "La copia parte dalla radio: spegnila, collega il cavo alla presa " + jack + " e accendila tenendo premuto " + hold +
		" (compare «CLONE»). Per leggere: avvia la lettura qui e poi premi " + send + " sulla radio. Per scrivere: premi " + recv +
		" sulla radio (compare «-WAIT-») e poi avvia la scrittura qui; alla fine la radio non conferma, l'esito si vede sul suo schermo. "
}

const vxNote = "999 canali con nomi di 6 caratteri, cavo di programmazione Yaesu, 19200 baud. "

var vxCloneNote = yaesuCloneHow("MIC/SP", "[F/W]", "[BAND]", "[V/M]")

func init() {
	genRegister(
		&genModel{ID: "ys-vx6", Vendor: "Yaesu", Model: "VX-6", RClass: "VX6Radio", Name: "Yaesu VX-6R / VX-6E",
			Slots: 999, First: 1, NameMax: 6, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ +-/?[]_$%*#.|=\\@",
			Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 1 W", 1, 2}, {"Minima 0,3 W", 0.3, 3}},
			Bands:  []band{{500000, 998990000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 19200, Fam: vxFam{},
			Opt:   &vxOpt{clone: vx6Clone, format: vx6Format, chars: vx6Chars, narrow: "half_deviation"},
			Notes: vxNote + vxCloneNote + "Driver di CHIRP."},
		&genModel{ID: "ys-vx3", Vendor: "Yaesu", Model: "VX-3", RClass: "VX3Radio", Name: "Yaesu VX-3R / VX-3E",
			Slots: 999, First: 1, NameMax: 6, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ +-/[](){}_*,'|",
			Powers: []genPower{{"Alta 1,5 W", 1.5, 0}, {"Bassa 0,1 W", 0.1, 2}},
			Bands:  []band{{500000, 999000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 19200, Fam: vx3Fam{},
			Opt:   &vxOpt{clone: vx3Clone, format: vx3Format, chars: vx3Chars, narrow: "txnarrow", vx3: true},
			Notes: vxNote + vxCloneNote + ysfNote},
	)
}

func (vxFam) opt(m *genModel) *vxOpt         { return m.Opt.(*vxOpt) }
func (f vxFam) yaesu(m *genModel) *yaesuOpt  { return f.opt(m).clone }
func (f vxFam) mem(m *genModel) *bwEl        { return bwMust(f.opt(m).format) }
func (f vxFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f vxFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

// flag: campo dei bit del canale (i pari: metà «odd», i dispari: metà «even»)
func (f vxFam) flag(m *genModel, i int, what string) *bwEl {
	half := "odd_"
	if i%2 == 1 {
		half = "even_"
	}
	return f.mem(m).F("flags").I(i / 2).F(half + what)
}

func (f vxFam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(m, i).Raw(img)...)
	return append(out, icomB(f.flag(m, i, "pskip").Bool(img)), icomB(f.flag(m, i, "skip").Bool(img)))
}

func (f vxFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(m, i).ByteLen()
	if len(raw) != n+2 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:n])
	f.flag(m, i, "pskip").SetBool(img, raw[n] != 0)
	f.flag(m, i, "skip").SetBool(img, raw[n+1] != 0)
	f.flag(m, i, "valid").SetBool(img, true)
	f.flag(m, i, "masked").SetBool(img, true)
}

func (f vxFam) Match(m *genModel, img []byte) bool { return yaesuMatch(f.opt(m).clone, img) }

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (vxFam) noIdent() {}

// yaesuRound: frequenza memorizzata al kHz riportata al valore vero (i 500 Hz dei
// passi da 12,5 kHz e i 250/750 Hz di quelli da 6,25 kHz sono sottintesi; in
// banda aerea i canali da 8,33 kHz). Come fix_rounded_step di CHIRP.
func yaesuRound(hz int64) int64 {
	if hz >= 108000000 && hz <= 137000000 {
		if hz%25000 == 0 {
			return hz
		}
		base := hz / 25000 * 25000
		f := hz
		switch k := hz - base; {
		case k == 5000:
			f -= 8333
		case k == 15000:
			f += 8333
		case k >= 20000 && k%5000 == 0:
			return -1 // non è un canale della banda aerea
		}
		best, bd := base, int64(1<<62)
		for i := int64(0); i < 3; i++ {
			c := base + 8333*i
			d := c - f
			if d < 0 {
				d = -d
			}
			if d < bd {
				best, bd = c, d
			}
		}
		return best
	}
	for _, add := range []int64{500, 250, 750, 330} {
		if (hz+add)%12500 == 0 || (hz+add)%6250 == 0 {
			return hz + add
		}
	}
	return hz
}

// yaesuKHz: valore da memorizzare (kHz) per la frequenza hz; false se la radio
// non può memorizzarla
func yaesuKHz(hz int64) (int64, bool) {
	k := hz / 1000
	return k, hz > 0 && k <= 999999 && yaesuRound(k*1000) == hz
}

// vxTmode: modo dei toni per la coppia (trasmissione, ricezione); -1 se la radio
// non ha quella combinazione
func (f vxFam) tmode(m *genModel, t, r genTone) int64 {
	k := ft60Tmode(t, r)
	switch {
	case k == 4:
		return 3
	case k > 4 && f.opt(m).vx3:
		return -1
	}
	return k
}

func (f vxFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	k := f.tmode(m, t, r)
	return k < 0 || k == 7
}

func (vxFam) toneOK(t, r genTone) bool { return ft60Tmode(t, r) >= 0 }

// vx3Fam: VX-3, senza i toni misti
type vx3Fam struct{ vxFam }

func (vx3Fam) toneOK(t, r genTone) bool { k := ft60Tmode(t, r); return k >= 0 && k <= 4 }
func (vxFam) fitBySet()                 {}

func (f vxFam) duplex(m *genModel, img []byte, r *bwEl) (tx int64, kept int64) {
	rx := yaesuRound(r.F("freq").Int(img) * 1000)
	kept = r.F("offset").Int(img) * 1000
	if !f.opt(m).vx3 {
		kept = yaesuRound(kept)
	}
	switch r.F("duplex").Int(img) {
	case 1:
		return rx - kept, kept
	case 2:
		return rx + kept, kept
	case 3:
		kept = yaesuRound(r.F("offset").Int(img) * 1000)
		return kept, kept
	}
	return rx, kept
}

func (f vxFam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.flag(m, i, "masked").Bool(img) || !f.flag(m, i, "valid").Bool(img) {
		return nil
	}
	r, o := f.rec(m, i), f.opt(m)
	c := &Chan{Rx: yaesuRound(r.F("freq").Int(img) * 1000)}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx, _ = f.duplex(m, img, r)
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img) & 0x3F); k < len(chirpTones) {
		tone = fmtTone(chirpTones[k])
	}
	if k := int(r.F("dcs").Int(img) & 0x7F); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	switch r.F("tmode").Int(img) {
	case 1:
		c.TxTone = tone
	case 2:
		c.TxTone, c.RxTone = tone, tone
	case 3:
		c.TxTone, c.RxTone = dcs, dcs
	case 5:
		c.TxTone = dcs
	case 6:
		c.TxTone, c.RxTone = tone, dcs
	case 7:
		c.TxTone, c.RxTone = dcs, tone
	}
	p := r.F("power").Int(img)
	if o.vx3 {
		c.Power = m.powerSem(int(^p & 1))
	} else {
		c.Power = m.powerSem(int(3 - p))
	}
	mode := r.F("mode").Int(img)
	c.Narrow = r.F(o.narrow).Bool(img) && (mode == 0 || mode == 3 && !o.vx3)
	c.Scan = !f.flag(m, i, "skip").Bool(img)
	if o.vx3 && f.flag(m, i, "pskip").Bool(img) {
		c.Scan = false
	}
	var sb strings.Builder
	for _, b := range r.F("name").Raw(img) {
		if b == 0xFF {
			break
		}
		if k := int(b & 0x7F); k < len(o.chars) && o.chars[k] != 0 {
			sb.WriteByte(o.chars[k])
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

func (f vxFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, o := f.rec(m, i), f.opt(m)
	valid, used := f.flag(m, i, "valid").Bool(img), f.flag(m, i, "masked").Bool(img)
	if c == nil {
		// come CHIRP: il canale visibile viene nascosto, quello già nascosto liberato
		if valid && !used {
			f.flag(m, i, "valid").SetBool(img, false)
			return nil
		}
		f.flag(m, i, "masked").SetBool(img, false)
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
	tmode := f.tmode(m, t, rt)
	if tmode < 0 {
		if o.vx3 {
			return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
		}
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione o in entrambi i versi, DCS in " +
			"trasmissione o in entrambi i versi, subtono e DCS misti; mai il solo tono in ricezione o codici invertiti)")
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	kept, tone, dcs, step, oldRx, oldMode := int64(600000), int64(8), int64(0), int64(0), int64(0), int64(0)
	if valid {
		_, kept = f.duplex(m, img, r)
		tone, dcs = r.F("tone").Int(img)&0x3F, r.F("dcs").Int(img)&0x7F
		step = r.F("tune_step").Int(img)
		oldRx, oldMode = yaesuRound(r.F("freq").Int(img)*1000), r.F("mode").Int(img)
	}
	if (!o.vx3 && step > 9) || (o.vx3 && step > 8) {
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
	if rt.Mode == "Tone" && tmode == 7 {
		if tone, err = find(chirpTones, rt.Hz10); err != nil {
			return err
		}
	}
	for _, g := range []genTone{t, rt} {
		if g.Mode == "DTCS" {
			if dcs, err = find(dtcsCodes, g.Code); err != nil {
				return err
			}
		}
	}
	dup, offs := int64(0), kept/1000
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	ok2 := ad%1000 == 0
	if !o.vx3 {
		_, ok2 = yaesuKHz(ad)
	}
	switch {
	case d == 0:
		if k, ok := yaesuKHz(kept); !ok && !o.vx3 || kept%1000 != 0 && o.vx3 {
			offs = 0
		} else if !o.vx3 {
			offs = k
		}
	case ad > 70000000 || !ok2 || ad/1000 > 999999:
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
	// modo: resta quello del canale se la frequenza non cambia, altrimenti segue la frequenza
	mode := icx90Mode(c.Rx) // 0 FM, 1 FM larga, 2 AM
	mode = []int64{0, 2, 1}[mode]
	if valid && oldRx == c.Rx && (oldMode == 1 || oldMode == 2) {
		mode = oldMode
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !valid {
		if o.vx3 {
			r.Fill(img, 0)
			r.F("unknown5").SetInt(img, 0x0D)
			r.F("unknown7b").SetInt(img, 1)
			r.F("automode").SetInt(img, 1)
		} else {
			for _, k := range []string{"unknown11", "step_changed", "cpu_shifted", "unknown12", "unknown2", "pager", "unknown5"} {
				r.F(k).SetInt(img, 0)
			}
		}
	}
	f.flag(m, i, "valid").SetBool(img, true)
	f.flag(m, i, "masked").SetBool(img, true)
	r.F("freq").SetInt(img, fk)
	r.F("offset").SetInt(img, offs)
	r.F("tone").SetInt(img, tone)
	r.F("tmode").SetInt(img, tmode)
	r.F("duplex").SetInt(img, dup)
	r.F("mode").SetInt(img, mode)
	r.F(o.narrow).SetBool(img, c.Narrow && mode == 0)
	r.F("dcs").SetInt(img, dcs)
	r.F("tune_step").SetInt(img, step)
	pi := int64(m.powerIndex(c.Power))
	if o.vx3 {
		r.F("power").SetInt(img, []int64{3, 0}[pi])
	} else {
		r.F("power").SetInt(img, 3-pi)
	}
	wasSkip := valid && used && f.flag(m, i, "skip").Bool(img)
	wasP := valid && used && f.flag(m, i, "pskip").Bool(img)
	switch {
	case !o.vx3:
		// VX-6: «pskip» è il canale preferenziale, che resta tale
		f.flag(m, i, "skip").SetBool(img, !c.Scan)
		f.flag(m, i, "pskip").SetBool(img, wasP && c.Scan)
	case c.Scan:
		f.flag(m, i, "skip").SetBool(img, false)
		f.flag(m, i, "pskip").SetBool(img, false)
	case wasP && !wasSkip:
		// VX-3: il salto «P» resta
	default:
		f.flag(m, i, "skip").SetBool(img, true)
		f.flag(m, i, "pskip").SetBool(img, false)
	}
	name := m.cleanName(c.Name)
	enc := make([]byte, 6)
	for k := range enc {
		ch := byte(' ')
		if k < len(name) {
			ch = name[k]
		}
		p := strings.IndexByte(o.chars, ch)
		if p < 0 || ch == 0 {
			p = strings.IndexByte(o.chars, ' ')
		}
		enc[k] = byte(p)
	}
	if strings.TrimSpace(name) != "" {
		enc[0] |= 0x80
	}
	r.F("name").SetRaw(img, enc)
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

func (f vxFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := yaesuDownload(ctx, m, port, f.opt(m).clone, prog)
	return img, "", err
}

func (f vxFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", yaesuUpload(ctx, m, port, f.opt(m).clone, img, prog)
}
