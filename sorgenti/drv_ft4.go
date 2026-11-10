// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ft4.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: Yaesu FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E e FT-25R (cavo SCU-35, 9600 baud).
// Il cavo unisce TX e RX: ogni comando torna indietro (eco) prima della risposta.
// «PROGRAM» → «QX» 06; 02 → identificativo di 16 byte (l'ultimo è il sottotipo) 06;
// 'R' indirizzo(2) 10 → 'W' indirizzo(2) 10 dati(16) somma 06;
// 'W' indirizzo(2) 10 dati(16) somma 06 → 06; «END» → 06.
// Memoria: 0x215 blocchi da 16 byte; il blocco 0 non si scrive.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - il codice 4 del campo «duplex» è il simplex della radio (menu RPT.SFT = SIMP),
//     che CHIRP mostra come trasmissione disabilitata: qui il canale è in simplex;
//   - uno scostamento che non è multiplo del passo della radio (25 o 50 kHz) si
//     scrive come frequenza di trasmissione separata, senza arrotondarlo.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type ft4Fam struct{}

type ft4Opt struct {
	id      string      // identificativo della radio fino al primo 0x00
	scale   int64       // scostamento in unità da 25 kHz × scale
	sub3    bool        // FT-65 con sottotipo 3: unità da 25 kHz anche se scale = 2
	eu      bool        // spostamento automatico europeo
	steps   []int64     // passi ammessi (Hz)
	nameLen int         // caratteri del nome mostrati dalla radio
	offBy   [3][2]int64 // FT-65R: trasmissione disabilitata con uno scostamento fuori gamma (segno, Hz) per gamma
}

const ft4Format = `
struct slot {
 u8 tx_pwr;
 bbcd freq[4];
 u8 tx_ctcss;
 u8 rx_ctcss;
 u8 tx_dcs;
 u8 rx_dcs;
 u8 unknown1:5,
    duplex:3;
 ul16 offset;
 u8 unknown2:7,
    tx_width:1;
 u8 step;
 u8 sql_type;
 u8 unused;
};
struct bitmap {
 u8 b8[28];
 u8 unused[4];
};
struct name {
 u8 chrs[8];
};
struct txfreq {
 bbcd freq[4];
};
#seekto 0x0010;
struct slot memory[200];
struct slot pms[20];
struct slot vfo[5];
struct slot home[3];
struct bitmap enable;
struct bitmap scan;
#seekto 0x1000;
struct name names[220];
#seekto 0x1700;
struct txfreq txfreqs[220];
`

const ft4Size = 0x215 * 16

var (
	ft4StepCode = []int64{0, 5000, 6250, 10000, 12500, 15000, 20000, 25000, 50000, 100000}
	ft4StepsEU  = ft4StepCode[1:]
	ft4StepsUS  = []int64{5000, 10000, 12500, 15000, 20000, 25000, 50000, 100000}
	ft4AutoUS   = [][3]int64{{145100000, 145495000, 2}, {146000000, 146395000, 0}, {146600000, 146995000, 2},
		{147000000, 147395000, 0}, {147600000, 147995000, 2}}
	ft4AutoEU = [][3]int64{{145600000, 145800000, 2}, {438200000, 439425000, 2}}
	ft4Dual   = []band{{65000000, 108000000}, {136000000, 174000000}, {400000000, 480000000}}
	ft4VHF    = []band{{65000000, 108000000}, {136000000, 174000000}}
	ft4Powers = []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 0,5 W", 0.5, 2}}
)

const ft4Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz*+-/@"

func init() {
	mk := func(id, model, rclass string, o *ft4Opt, bands []band, notxoff bool, extra string) *genModel {
		name := "Yaesu " + model
		if o.sub3 {
			name += " (sottotipo 3)"
		}
		notes := fmt.Sprintf("200 canali con nomi di %d caratteri. Cavo Yaesu SCU-35 (o compatibile) nella presa del microfono, 9600 baud: "+
			"il cavo rimanda indietro ogni byte inviato. Passo di sintonia, banchi, canali PMS, VFO e tasti programmabili si conservano. ", o.nameLen)
		if notxoff {
			notes += "La radio non ha un blocco della trasmissione per canale. "
		} else {
			notes += "La trasmissione disabilitata si ottiene, come in CHIRP, con uno scostamento che porta fuori gamma. "
		}
		return &genModel{ID: id, Vendor: "Yaesu", Model: model, RClass: rclass, Name: name,
			Slots: 200, First: 1, NameMax: o.nameLen, Charset: ft4Charset, Powers: ft4Powers, Bands: bands,
			Narrow: true, HasScan: true, NoTxOff: notxoff, DCS: dtcsCodes, Baud: 9600, Fam: ft4Fam{}, Opt: o,
			Notes: notes + extra + "Driver di CHIRP."}
	}
	const s3 = "Scheda per le radio che dichiarano il sottotipo 3 (scostamento in unità da 25 kHz): se la radio è di un altro tipo il programma lo dice. "
	off65 := [3][2]int64{{0, 0}, {-1, 99950000}, {1, 99950000}}
	genRegister(
		mk("ys-ft4xr", "FT-4XR", "YaesuFT4XRRadio", &ft4Opt{id: "IFT-35R", scale: 1, steps: ft4StepsUS, nameLen: 6}, ft4Dual, true, ""),
		mk("ys-ft4xe", "FT-4XE", "YaesuFT4XERadio", &ft4Opt{id: "IFT-35R", scale: 1, eu: true, steps: ft4StepsEU, nameLen: 6}, ft4Dual, true, ""),
		mk("ys-ft4vr", "FT-4VR", "YaesuFT4VRRadio", &ft4Opt{id: "IFT-15R", scale: 1, steps: ft4StepsUS, nameLen: 6}, ft4VHF, true, ""),
		mk("ys-ft65r", "FT-65R", "YaesuFT65RRadio", &ft4Opt{id: "IH-420", scale: 2, steps: ft4StepsEU, nameLen: 8, offBy: off65}, ft4Dual, false, ""),
		mk("ys-ft65e", "FT-65E", "YaesuFT65ERadio", &ft4Opt{id: "IH-420", scale: 2, eu: true, steps: ft4StepsEU, nameLen: 8}, ft4Dual, true, ""),
		mk("ys-ft25r", "FT-25R", "YaesuFT25RRadio", &ft4Opt{id: "IFT-25R", scale: 2, steps: ft4StepsUS, nameLen: 8}, ft4VHF, true, ""),
		mk("ys-ft65r-s3", "FT-65R", "YaesuFT65RRadio", &ft4Opt{id: "IH-420", scale: 2, sub3: true, steps: ft4StepsEU, nameLen: 8, offBy: off65}, ft4Dual, false, s3),
		mk("ys-ft65e-s3", "FT-65E", "YaesuFT65ERadio", &ft4Opt{id: "IH-420", scale: 2, sub3: true, eu: true, steps: ft4StepsEU, nameLen: 8}, ft4Dual, true, s3),
		mk("ys-ft25r-s3", "FT-25R", "YaesuFT25RRadio", &ft4Opt{id: "IFT-25R", scale: 2, sub3: true, steps: ft4StepsUS, nameLen: 8}, ft4VHF, true, s3),
	)
}

func ft4o(m *genModel) *ft4Opt { return m.Opt.(*ft4Opt) }

// unità dello scostamento in Hz
func (o *ft4Opt) unit() int64 {
	if o.sub3 {
		return 25000
	}
	return 25000 * o.scale
}

func (o *ft4Opt) hasOffTrick() bool { return o.offBy[1][0] != 0 }

func (ft4Fam) mem() *bwEl                 { return bwMust(ft4Format) }
func (f ft4Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f ft4Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

func (f ft4Fam) Rec(m *genModel, img []byte, i int) []byte {
	r := f.mem()
	out := append([]byte{}, f.rec(i).Raw(img)...)
	out = append(out, r.F("names").I(i).Raw(img)...)
	out = append(out, r.F("txfreqs").I(i).Raw(img)...)
	return append(out, f.bit(img, "enable", i), f.bit(img, "scan", i))
}

func (f ft4Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) != 16+8+4+2 {
		return
	}
	r := f.mem()
	f.rec(i).SetRaw(img, raw[:16])
	r.F("names").I(i).SetRaw(img, raw[16:24])
	r.F("txfreqs").I(i).SetRaw(img, raw[24:28])
	f.setBit(img, "enable", i, raw[28] != 0)
	f.setBit(img, "scan", i, raw[29] != 0)
}

func (ft4Fam) Match(m *genModel, img []byte) bool { return len(img) == ft4Size }

func (f ft4Fam) bit(img []byte, what string, i int) byte {
	return byte(f.mem().F(what).F("b8").I(i/8).Int(img)>>(uint(i)%8)) & 1
}

func (f ft4Fam) setBit(img []byte, what string, i int, on bool) {
	e := f.mem().F(what).F("b8").I(i / 8)
	v := e.Int(img)
	if on {
		v |= 1 << (uint(i) % 8)
	} else {
		v &^= 1 << (uint(i) % 8)
	}
	e.SetInt(img, v)
}

func (m *genModel) ft4Band(hz int64) int {
	for k, b := range m.Bands {
		if hz >= b.Lo && hz <= b.Hi {
			return k
		}
	}
	return -1
}

// ft4Duplex: verso (0 simplex, +1, -1), scostamento e frequenza di trasmissione
// del canale, con «off» vero se la trasmissione è disabilitata (solo FT-65R).
// chirpOff: CHIRP mostra il canale come «off» (anche il simplex con codice 4).
func (f ft4Fam) duplex(m *genModel, img []byte, i int) (tx int64, off bool, chirpOff bool, keptOffset int64) {
	o := ft4o(m)
	r := f.rec(i)
	rx := r.F("freq").Int(img) * 10
	code := r.F("duplex").Int(img)
	offs := r.F("offset").Int(img) * o.unit()
	if code == 5 { // automatico: il verso dipende dalla frequenza
		code = 4
		auto := ft4AutoUS
		if o.eu {
			auto = ft4AutoEU
		}
		for _, a := range auto {
			if rx >= a[0] && rx < a[1] {
				code = a[2]
			}
		}
	}
	switch code {
	case 0, 2:
		sign := int64(1)
		if code == 2 {
			sign = -1
		}
		tx = rx + sign*offs
		if o.hasOffTrick() {
			if b := m.ft4Band(rx); b >= 0 && o.offBy[b][0] != 0 && m.ft4Band(tx) < 0 {
				return rx, true, true, 0
			}
		}
		return tx, false, false, offs
	case 6:
		tx = f.mem().F("txfreqs").I(i).F("freq").Int(img) * 10
		return tx, false, false, tx
	case 4:
		return rx, false, true, offs
	}
	return rx, false, false, offs
}

func (f ft4Fam) tones(img []byte, r *bwEl) (tx, rx string) {
	tone := func(v int64) string {
		if v >= 1 && int(v) <= len(chirpTones) {
			return fmtTone(chirpTones[v-1])
		}
		return "88.5"
	}
	dcs := func(v int64) string {
		if v >= 1 && int(v) <= len(dtcsCodes) {
			return fmt.Sprintf("D%03dN", dtcsCodes[v-1])
		}
		return "D023N"
	}
	tc, rc := r.F("tx_ctcss").Int(img), r.F("rx_ctcss").Int(img)
	td, rd := r.F("tx_dcs").Int(img), r.F("rx_dcs").Int(img)
	switch r.F("sql_type").Int(img) {
	case 1: // solo in ricezione
		return "", tone(rc)
	case 2:
		return tone(tc), ""
	case 3:
		switch {
		case tc == 0 && rc == 0:
		case tc == 0, tc == rc:
			return tone(rc), tone(rc)
		case rc == 0:
			return tone(tc), ""
		default:
			return tone(tc), tone(rc)
		}
	case 5:
		switch {
		case td == 0 && rd == 0:
		case td == 0:
			return "", dcs(rd)
		case rd == 0:
			return dcs(td), ""
		case td == rd:
			return dcs(td), dcs(td)
		default:
			return dcs(td), dcs(rd)
		}
	}
	return "", "" // 0 spento, 4 tono rovesciato, 6 chiamata selettiva (pager)
}

func (f ft4Fam) Get(m *genModel, img []byte, i int) *Chan {
	if f.bit(img, "enable", i) == 0 {
		return nil
	}
	o := ft4o(m)
	r := f.rec(i)
	c := &Chan{Rx: r.F("freq").Int(img) * 10}
	tx, off, _, _ := f.duplex(m, img, i)
	c.Tx, c.TxOff = tx, off
	if off {
		c.Tx = c.Rx
	}
	c.TxTone, c.RxTone = f.tones(img, r)
	p := 2 - int(r.F("tx_pwr").Int(img))
	if p < 0 {
		p = 2
	}
	c.Power = m.powerSem(p)
	c.Narrow = r.F("tx_width").Bool(img)
	c.Scan = f.bit(img, "scan", i) != 0
	raw := f.mem().F("names").I(i).Raw(img)
	var nm []byte
	for _, b := range raw[:o.nameLen] {
		if b == 0 {
			break
		}
		if b == 0x7F {
			b = ' '
		}
		nm = append(nm, b)
	}
	c.Name = string(bytes.TrimRight(nm, " "))
	return c
}

// passo valido per la radio: quello del canale se è ammesso, altrimenti il primo
// che divide la frequenza
func (o *ft4Opt) validStep(step, hz int64) int64 {
	for _, s := range o.steps {
		if s == step {
			return step
		}
	}
	for _, s := range o.steps {
		if hz%s == 0 {
			return s
		}
	}
	return o.steps[0]
}

func ft4Index(list []int, v int) int64 {
	for k, x := range list {
		if x == v {
			return int64(k + 1)
		}
	}
	return -1
}

// chirpInvalid: combinazioni che la radio non ha (CTCSS in un verso e DCS nell'altro, DCS invertito)
func (ft4Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return (t.Mode != "" && r.Mode != "" && t.Mode != r.Mode) || t.Inv || r.Inv
}

// identGroup (4.14, «Identifica»): il saluto «PROGRAM» a 9600 baud è lo stesso delle
// AnyTone TERMN-8R/OBLTR-8R e della KYD NC-630A, ma qui il cavo rimanda l'eco e la
// radio risponde solo «QX» 06 (tre byte). Provata per prima con il driver di
// un'altra famiglia, quei tre byte non bastano a dire che una radio ha risposto e
// tutto il gruppo veniva abbandonato: le Yaesu hanno perciò un gruppo loro.
func (ft4Fam) identGroup() string { return "yaesu-ft4" }

// toneOK (4.14): combinazioni di toni che la radio ha; le altre sono adattate
// da genAdapt prima della scrittura
func (ft4Fam) toneOK(t, r genTone) bool {
	return !t.Inv && !r.Inv && !(t.Mode != "" && r.Mode != "" && t.Mode != r.Mode)
}

func (f ft4Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if c == nil {
		f.setBit(img, "enable", i, false)
		return nil
	}
	o := ft4o(m)
	r := f.rec(i)
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	if t.Inv || rt.Inv {
		return fmt.Errorf("la radio non ha i codici DCS invertiti")
	}
	if t.Mode != "" && rt.Mode != "" && t.Mode != rt.Mode {
		return fmt.Errorf("la radio non permette un subtono in un verso e un codice DCS nell'altro")
	}
	// valori del canale già presente, che restano quando non cambiano (come CHIRP)
	was := f.bit(img, "enable", i) != 0
	oldStep, oldRx, kept := int64(0), int64(0), int64(0)
	tc, rc, td, rd := int64(9), int64(9), int64(1), int64(1) // 88,5 Hz e DCS 023
	pager := false
	if was {
		if k := r.F("step").Int(img); k >= 0 && int(k) < len(ft4StepCode) {
			oldStep = ft4StepCode[k]
		}
		oldRx = r.F("freq").Int(img) * 10
		_, _, _, kept = f.duplex(m, img, i)
		if r.F("sql_type").Int(img) == 6 {
			pager = true
		} else {
			keep := func(name string, max int, def int64) int64 {
				if v := r.F(name).Int(img); v >= 1 && int(v) <= max {
					return v
				}
				return def
			}
			tc, rc = keep("tx_ctcss", len(chirpTones), 9), keep("rx_ctcss", len(chirpTones), 9)
			td, rd = keep("tx_dcs", len(dtcsCodes), 1), keep("rx_dcs", len(dtcsCodes), 1)
		}
	}
	step := o.validStep(oldStep, oldRx)
	idx := func(g genTone) (int64, error) {
		var k int64
		if g.Mode == "Tone" {
			k = ft4Index(chirpTones, g.Hz10)
		} else {
			k = ft4Index(dtcsCodes, g.Code)
		}
		if k < 0 {
			return 0, fmt.Errorf("tono %s non disponibile su questa radio", g.String())
		}
		return k, nil
	}
	sql := int64(0)
	switch {
	case t.Mode == "" && rt.Mode == "":
	case t.Mode == "Tone" && rt.Mode == "":
		if tc, err = idx(t); err != nil {
			return err
		}
		sql = 2
	case t.Mode == "" && rt.Mode == "Tone":
		if rc, err = idx(rt); err != nil {
			return err
		}
		sql = 1
	case t.Mode == "Tone":
		if tc, err = idx(t); err != nil {
			return err
		}
		if rc, err = idx(rt); err != nil {
			return err
		}
		sql = 3
	case t.Mode == "DTCS" && rt.Mode == "":
		if td, err = idx(t); err != nil {
			return err
		}
		sql, rd = 5, 0
	case t.Mode == "" && rt.Mode == "DTCS":
		if rd, err = idx(rt); err != nil {
			return err
		}
		sql, td = 5, 0
	default:
		if td, err = idx(t); err != nil {
			return err
		}
		if rd, err = idx(rt); err != nil {
			return err
		}
		sql = 5
	}
	if pager {
		sql = 6
	}

	// verso e scostamento
	code, offs, txf := int64(1), kept, c.Rx/10
	d := c.Tx - c.Rx
	switch {
	case c.TxOff:
		code, offs = 4, 0
		if b := m.ft4Band(c.Rx); b >= 0 && o.offBy[b][0] != 0 {
			code, offs = 0, o.offBy[b][1]
			if o.offBy[b][0] < 0 {
				code = 2
			}
		}
	case d == 0:
	case d > 70000000 || d < -70000000 || d%o.unit() != 0:
		code, offs, txf = 6, c.Tx, c.Tx/10
	case d > 0:
		code, offs = 0, d
	default:
		code, offs = 2, -d
	}

	r.F("unused").SetInt(img, 0)
	r.F("unknown1").SetInt(img, 0)
	r.F("unknown2").SetInt(img, 0)
	r.F("freq").SetInt(img, c.Rx/10)
	r.F("tx_ctcss").SetInt(img, tc)
	r.F("rx_ctcss").SetInt(img, rc)
	r.F("tx_dcs").SetInt(img, td)
	r.F("rx_dcs").SetInt(img, rd)
	r.F("sql_type").SetInt(img, sql)
	r.F("tx_pwr").SetInt(img, int64(2-m.powerIndex(c.Power)))
	r.F("tx_width").SetBool(img, c.Narrow)
	for k, s := range ft4StepCode {
		if s == step {
			r.F("step").SetInt(img, int64(k))
		}
	}
	r.F("offset").SetInt(img, offs/o.unit())
	f.setBit(img, "enable", i, true)
	f.setBit(img, "scan", i, c.Scan)
	name := []byte(m.cleanName(c.Name) + "        ")[:8]
	f.mem().F("names").I(i).F("chrs").SetRaw(img, name)
	f.mem().F("txfreqs").I(i).F("freq").SetInt(img, txf)
	r.F("duplex").SetInt(img, code)
	return nil
}

// ---------------------------------------------------------------- protocollo

type ft4Sess struct {
	*rawSess
}

var (
	errFt4NoEcho   = rErr("il cavo non rimanda indietro i dati inviati: controlla il cavo e che sia inserito fino in fondo")
	errFt4NoAnswer = rErr("la radio non risponde")
)

func ft4Sum(b []byte) byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return s
}

// cmd: invia il comando, toglie l'eco e legge la risposta (n byte, oppure fino
// al 06 se n < 0) e la conferma 06
func (s *ft4Sess) cmd(c []byte, n int) ([]byte, error) {
	for try := 0; ; try++ {
		if err := s.send(c); err != nil {
			return nil, err
		}
		echo := readUpTo(s.ctx, s.p, len(c), time.Second)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(echo) == 0 {
			return nil, errFt4NoEcho
		}
		if !bytes.Equal(echo, c) {
			return nil, rErr("il cavo non rimanda indietro i dati inviati come dovrebbe: radio spenta o cavo non adatto")
		}
		if n < 0 {
			var resp []byte
			for {
				b := readUpTo(s.ctx, s.p, 1, time.Second)
				if s.ctx.Err() != nil {
					return nil, errCancelled
				}
				if len(b) == 0 {
					return nil, rErr("la radio non risponde")
				}
				if b[0] == 0x06 {
					return resp, nil
				}
				if resp = append(resp, b[0]); len(resp) > 256 {
					return nil, rErr("risposta della radio troppo lunga")
				}
			}
		}
		var resp []byte
		if n > 0 {
			resp = readUpTo(s.ctx, s.p, n, time.Second)
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
			if len(resp) == 0 && string(c) == "PROGRAM" {
				return nil, errFt4NoAnswer
			}
		}
		ack := readUpTo(s.ctx, s.p, 1, time.Second)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(ack) == 1 && ack[0] == 0x06 {
			return resp, nil
		}
		if try >= 5 {
			return nil, rErr("la radio non ha confermato il comando")
		}
	}
}

// start: modalità di programmazione e identificativo; restituisce il sottotipo
func (s *ft4Sess) start(m *genModel) (int, error) {
	o := ft4o(m)
	ok := false
	var last error
enter:
	for round := 0; round < 3 && !ok; round++ {
		for k := 0; k < 3; k++ {
			r, err := s.cmd([]byte("PROGRAM"), 2)
			if err == errCancelled {
				return 0, err
			}
			if err == nil && string(r) == "QX" {
				ok = true
				break enter
			}
			if err == errFt4NoEcho || err == errFt4NoAnswer {
				return 0, err
			}
			last = err
		}
		if _, err := s.cmd([]byte("END"), 0); err == errCancelled {
			return 0, err
		}
	}
	if !ok {
		if last != nil {
			return 0, last
		}
		return 0, rErr("la radio non è entrata in programmazione (attesa la risposta «QX»)")
	}
	id, err := s.cmd([]byte{0x02}, -1)
	if err != nil {
		return 0, err
	}
	if len(id) == 0 {
		return 0, rErr("identificativo della radio vuoto")
	}
	head := id
	if k := bytes.IndexByte(id, 0); k >= 0 {
		head = id[:k]
	}
	if string(head) != o.id {
		s.cmd([]byte("END"), 0)
		for _, x := range genModels {
			if xo, isFt4 := x.Opt.(*ft4Opt); isFt4 && xo.id == string(head) && !xo.sub3 {
				return 0, rErr("la radio è un altro modello (%s): apri la scheda «%s» o quella della stessa serie", head, x.Name)
			}
		}
		return 0, rErr("identificativo della radio sconosciuto (%q)", head)
	}
	sub := int(id[len(id)-1])
	if o.scale == 2 && (sub == 3) != o.sub3 {
		s.cmd([]byte("END"), 0)
		for _, x := range genModels {
			if xo, isFt4 := x.Opt.(*ft4Opt); isFt4 && x.Model == m.Model && xo.sub3 == (sub == 3) {
				return 0, rErr("la radio dichiara il sottotipo %d: apri la scheda «%s»", sub, x.Name)
			}
		}
	}
	return sub, nil
}

func (f ft4Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return img, note, err
}

// DownloadMeta: come Download, con il sottotipo della radio nei metadati del file
// (chiave «ft4_subtype», la stessa di CHIRP)
func (f ft4Fam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, nil, "", err
	}
	defer rs.p.Close()
	s := &ft4Sess{rs}
	prog(0, ft4Size, "Collegamento con la radio…")
	sub, err := s.start(m)
	if err != nil {
		return nil, nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, ft4Size)
	for a := 0; a < ft4Size; a += 16 {
		prog(a, ft4Size, "Lettura…")
		ok := false
		for try := 0; try < 3 && !ok; try++ {
			c := []byte{'R', byte(a >> 8), byte(a), 16}
			r, err := s.cmd(c, 21)
			if err != nil {
				return nil, nil, "", err
			}
			if len(r) == 21 && r[0] == 'W' && bytes.Equal(r[1:4], c[1:4]) && ft4Sum(r[1:20]) == r[20] {
				copy(img[a:], r[4:20])
				ok = true
			}
		}
		if !ok {
			s.cmd([]byte("END"), 0)
			return nil, nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
	}
	if _, err := s.cmd([]byte("END"), 0); err != nil {
		return nil, nil, "", err
	}
	prog(ft4Size, ft4Size, "Lettura completata")
	return img, map[string]interface{}{"ft4_subtype": sub}, "", nil
}

func (f ft4Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != ft4Size {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer rs.p.Close()
	s := &ft4Sess{rs}
	prog(0, ft4Size, "Collegamento con la radio…")
	if _, err := s.start(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 16; a < ft4Size; a += 16 {
		prog(a, ft4Size, "Scrittura…")
		body := append([]byte{byte(a >> 8), byte(a), 16}, img[a:a+16]...)
		msg := append([]byte{'W'}, body...)
		msg = append(msg, ft4Sum(body), 0x06)
		if _, err := s.cmd(msg, 0); err != nil {
			return "", genFail(fmt.Sprintf("Scrittura del blocco 0x%04X", a), err)
		}
	}
	if _, err := s.cmd([]byte("END"), 0); err != nil {
		return "", err
	}
	prog(ft4Size, ft4Size, "Scrittura completata")
	return "", nil
}
