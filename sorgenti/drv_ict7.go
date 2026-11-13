// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ict7h.py, ict8.py (4.33).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.33: palmari Icom degli anni '90 gestiti da CHIRP, sul collegamento Icom di
// drv_icom.go.
//
// IC-T7H (bibanda, 60 canali numerati da 1, senza nome): canale di 8 byte con la
// frequenza in cifre (centinaia di kHz in due byte, poi una cifra per le decine
// di kHz e mezza cifra per i 500 Hz) e lo scostamento in decine di kHz; un byte
// per canale a 0x0270 con vuoto, modo dei toni, verso dello scostamento ed
// esclusione dalla scansione. I subtoni sono memorizzati come posizione
// nell'elenco + 1. Niente DCS, niente potenza nel canale.
//
// IC-T8A (tribanda 50 / 144 / 430 MHz, 100 canali da 0, nomi di 4 caratteri):
// frequenza e scostamento in sei cifre di kHz, con l'ultima cifra che conta i
// 500 Hz; nomi a 0x0400 (quattro FF = senza nome), un byte per canale a 0x0600.
//
// Come in CHIRP: cancellando un canale dell'IC-T7H i suoi campi tornano ai
// valori di un canale nuovo; nell'IC-T8A si segna solo come vuoto.

import (
	"context"
	"fmt"
	"strings"
)

type ict7Fam struct{}

type ict7Opt struct {
	clone *icomOpt
	t8    bool // IC-T8A
}

const ict7hFormat = `
struct {
  bbcd freq[2];
  u8  lastfreq:4,
      fraction:4;
  bbcd offset[2];
  u8  unknown;
  u8  rtone;
  u8  ctone;
} memory[60];

#seekto 0x0270;
struct {
  u8 empty:1,
     tmode:2,
     duplex:2,
     unknown3:1,
     skip:1,
     unknown4:1;
} flags[60];
`

const ict8aFormat = `
struct {
  bbcd freq[3];
  bbcd offset[3];
  u8 rtone;
  u8 ctone;
} memory[100];

#seekto 0x0400;
struct {
  char name[4];
} names[100];

#seekto 0x0600;
struct {
  u8 empty:1,
     skip:1,
     tmode:2,
     duplex:2,
     unknown2:2;
} flags[100];
`

func init() {
	t7 := &ict7Opt{clone: &icomOpt{model: [4]byte{0x18, 0x10, 0x00, 0x01}, end: "Icom Inc.", size: 0x03B0, ranges: [][3]int{{0x0000, 0x03B0, 16}}}}
	t8 := &ict7Opt{t8: true, clone: &icomOpt{model: [4]byte{0x19, 0x03, 0x00, 0x01}, end: "Icom Inc.", size: 0x07B0, ranges: [][3]int{{0x0000, 0x07B0, 16}}}}
	genRegister(
		&genModel{ID: "ic-t7h", Vendor: "Icom", Model: "IC-T7H", RClass: "ICT7HRadio", Name: "Icom IC-T7H",
			Slots: 60, First: 1, Powers: []genPower{{"Unica", 6, 0}},
			Bands: []band{{118000000, 174000000}, {400000000, 470000000}}, HasScan: true, NoTxOff: true, Baud: 9600, Fam: ict7Fam{}, Opt: t7,
			Notes: "60 canali VHF/UHF (da 1 a 60) senza nome. " + icomCableNote + icomNoDCS + icomNoPower + "Driver di CHIRP."},
		&genModel{ID: "ic-t8a", Vendor: "Icom", Model: "IC-T8A", RClass: "ICT8ARadio", Name: "Icom IC-T8A",
			Slots: 100, First: 0, NameMax: 4, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890", Powers: []genPower{{"Unica", 5, 0}},
			Bands: []band{{50000000, 54000000}, {118000000, 174000000}, {400000000, 470000000}}, HasScan: true, NoTxOff: true, Baud: 9600, Fam: ict7Fam{}, Opt: t8,
			Notes: "100 canali (da 0 a 99) sui 50, 144 e 430 MHz, con nomi di 4 caratteri. " + icomCableNote + icomNoDCS + icomNoPower + "Driver di CHIRP."},
	)
}

// frasi comuni delle note dei palmari Icom della 4.33
const (
	icomNoDCS   = "La radio non ha il DCS. "
	icomNoPower = "La potenza non è memorizzata nel canale. "
	icomModeHz  = "Il modo (FM, FM larga, AM) segue la frequenza. "
	icomFastCp  = "La copia è quella veloce (38400 baud dopo l'identificazione). "
)

func ict7o(m *genModel) *ict7Opt { return m.Opt.(*ict7Opt) }

func (ict7Fam) icom(m *genModel) *icomOpt { return ict7o(m).clone }

func (ict7Fam) mem(m *genModel) *bwEl {
	if ict7o(m).t8 {
		return bwMust(ict8aFormat)
	}
	return bwMust(ict7hFormat)
}

func (f ict7Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f ict7Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f ict7Fam) flg(m *genModel, i int) *bwEl { return f.mem(m).F("flags").I(i) }

func (f ict7Fam) Rec(m *genModel, img []byte, i int) []byte {
	raw := append([]byte{}, f.rec(m, i).Raw(img)...)
	if ict7o(m).t8 {
		raw = append(raw, f.mem(m).F("names").I(i).Raw(img)...)
	}
	return append(raw, f.flg(m, i).Raw(img)...)
}

func (f ict7Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(m, i).ByteLen()
	k := 0
	if ict7o(m).t8 {
		k = 4
	}
	if len(raw) != n+k+1 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:n])
	if k > 0 {
		f.mem(m).F("names").I(i).SetRaw(img, raw[n:n+k])
	}
	f.flg(m, i).SetRaw(img, raw[n+k:])
	f.flg(m, i).F("empty").SetBool(img, false)
}

func (ict7Fam) Match(m *genModel, img []byte) bool { return len(img) == ict7o(m).clone.size }

// combinazioni di toni che la radio ha: nessuno, subtono in trasmissione, lo
// stesso subtono nei due versi (niente DCS)
func (ict7Fam) toneOK(t, r genTone) bool {
	k := icomToneMode(t, r)
	return k >= 0 && k != 3
}
func (ict7Fam) fitBySet() {}

func (ict7Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	k := icomToneMode(t, r)
	d := c.Tx - c.Rx
	return k < 0 || k == 3 || d > 70000000 || d < -70000000
}

// ict7Tone: posizione del subtono nell'elenco per il valore memorizzato
// (posizione + 1). Lo zero, come in CHIRP, è l'ultimo subtono dell'elenco;
// -1 se il valore non è leggibile.
func ict7Tone(raw int64) int64 {
	switch {
	case raw == 0:
		return int64(len(chirpTones)) - 1
	case raw <= int64(len(chirpTones)):
		return raw - 1
	}
	return -1
}

// frequenza dell'IC-T8A: sei cifre di kHz; le ultime quattro cifre binarie non
// sono una cifra decimale ma il numero di passi da 500 Hz (da 0 a 15) oltre le
// decine di kHz. Come in CHIRP (_get_freq / _set_freq).
func ict8Freq(e *bwEl, img []byte) int64 {
	raw := e.Raw(img)
	last := int64(raw[2] & 0x0F)
	v := int64(0)
	for _, b := range raw {
		v = v*100 + int64(b>>4)*10 + int64(b&0x0F)
	}
	return (v-last)*1000 + last*500
}

func ict8SetFreq(e *bwEl, img []byte, hz int64) {
	e.SetInt(img, hz/1000)
	raw := append([]byte{}, e.Raw(img)...)
	raw[2] = raw[2]&0xF0 + byte(hz%10000/500)
	e.SetRaw(img, raw)
}

func (f ict7Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, fl := f.rec(m, i), f.flg(m, i)
	if fl.F("empty").Bool(img) {
		return nil
	}
	c := &Chan{}
	var offs int64
	if ict7o(m).t8 {
		c.Rx, offs = ict8Freq(r.F("freq"), img), ict8Freq(r.F("offset"), img)
		if nm := f.mem(m).F("names").I(i).F("name"); string(nm.Raw(img)) != "\xFF\xFF\xFF\xFF" {
			c.Name = strings.TrimRight(strings.TrimRight(nm.Str(img), " "), "\x00")
		}
	} else {
		c.Rx = r.F("freq").Int(img)*100000 + r.F("lastfreq").Int(img)*10000 + r.F("fraction").Int(img)*500
		offs = r.F("offset").Int(img) * 10000
	}
	c.Tx = c.Rx
	switch r2 := fl.F("duplex").Int(img); r2 {
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
	c.TxTone, c.RxTone = icomTones(mode, ict7Tone(r.F("rtone").Int(img)), ict7Tone(r.F("ctone").Int(img)), 0, 0)
	c.Scan = !fl.F("skip").Bool(img)
	return c
}

func (f ict7Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(m, i), f.flg(m, i)
	t8 := ict7o(m).t8
	if c == nil {
		if !t8 { // come CHIRP: i campi tornano quelli di un canale nuovo
			r.F("freq").SetInt(img, 0)
			r.F("lastfreq").SetInt(img, 0)
			r.F("fraction").SetInt(img, 0)
			r.F("offset").SetInt(img, 60)
			r.F("rtone").SetInt(img, 9)
			r.F("ctone").SetInt(img, 9)
			fl.F("tmode").SetInt(img, 0)
			fl.F("duplex").SetInt(img, 0)
			fl.F("skip").SetBool(img, false)
		}
		fl.F("empty").SetBool(img, true)
		return nil
	}
	// IC-T7H: come in CHIRP, scostamento e subtoni già presenti nella posizione
	// restano anche quando era vuota (se sono valori leggibili)
	was := (!fl.F("empty").Bool(img) || !t8) && ict7Tone(r.F("rtone").Int(img)) >= 0 && ict7Tone(r.F("ctone").Int(img)) >= 0
	kept, rtone, ctone := int64(600000), int64(8), int64(8)
	if was {
		rtone, ctone = ict7Tone(r.F("rtone").Int(img)), ict7Tone(r.F("ctone").Int(img))
		if t8 {
			kept = ict8Freq(r.F("offset"), img)
		} else {
			kept = r.F("offset").Int(img) * 10000
		}
	}
	tmode, rtone, ctone, _, _, err := icomTonesFor(c, rtone, ctone, 0)
	if err != nil {
		return err
	}
	if tmode == 3 {
		return fmt.Errorf("la radio non ha i codici DCS")
	}
	if c.Rx%500 != 0 || t8 && c.Rx%10000/500 > 15 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 3, d
	case d < 0:
		dup, offs = 2, -d
	}
	step, max := int64(10000), int64(99990000)
	if t8 {
		step, max = 500, 999997500
	}
	if dup != 0 && (offs%step != 0 || t8 && offs%10000/500 > 15) {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz", offs)
	}
	if offs > max {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	if t8 {
		r.Fill(img, 0)
		fl.Fill(img, 0)
		ict8SetFreq(r.F("freq"), img, c.Rx)
		ict8SetFreq(r.F("offset"), img, offs)
		nm := f.mem(m).F("names").I(i).F("name")
		if name := m.cleanName(c.Name); strings.TrimSpace(name) != "" {
			nm.SetStr(img, name, ' ')
		} else {
			nm.Fill(img, 0xFF)
		}
	} else {
		r.F("freq").SetInt(img, c.Rx/100000)
		r.F("lastfreq").SetInt(img, c.Rx%100000/10000)
		r.F("fraction").SetInt(img, c.Rx%10000/500)
		r.F("offset").SetInt(img, offs/10000)
	}
	r.F("rtone").SetInt(img, rtone+1)
	r.F("ctone").SetInt(img, ctone+1)
	fl.F("tmode").SetInt(img, []int64{0, 2, 3}[tmode])
	fl.F("duplex").SetInt(img, dup)
	fl.F("skip").SetBool(img, !c.Scan)
	fl.F("empty").SetBool(img, false)
	return nil
}

func (f ict7Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f ict7Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("l'immagine non è di una %s", m.Model)
	}
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
