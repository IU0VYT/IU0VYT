// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: retevis_h777v4.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Retevis H777 V4 (nuova versione della H777, comando «C777HAM») e Retevis
// RT21H («T210GAM», frequenze a 24 bit da 0x2000000)

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"
)

const h777v4Format = `
struct {
  ul32 rxfreq;        // 0-3
  ul32 txfreq;        // 4-7
  lbcd rx_tone[2];    // 8-9
  lbcd tx_tone[2];    // A-B
  u8 unknown0:2,      // C
     scramb:1,        //     Scramble  0 = Off, 1 = On
     scanadd:1,       //     Scan Add  0 = Scan, 1 = Skip
     ishighpower:1,   //     Power Level  0 = Low, 1 = High
     narrow:1,        //     Bandwidth  0 = Wide, 1 = Narrow
     unknown1:1,      //
     bcl:1;           //     Busy Channel Lockout  0 = On, 1 = Off
} memory[16];

struct {
  u8 codesw:1,        // 00D0  Code Switch
     scanm:1,         //       Scan Mode
     voxs:1,          //       VOX Switch
     roger:1,         //       Roger
     voice:1,         //       Voice Annunciation
     unknown_0:1,
     save:1,          //       Battery Save
     beep:1;          //       Beep Tone
  u8 squelch;         // 00D1  Squelch Level
  u8 tot;             // 00D2  Time-out Timer
  u8 vox;             // 00D3  VOX Level
  u8 voxd;            // 00D4  Vox Delay
  u8 unknown_1;
  u8 unknown_2;
  u8 skey2;           // 00D7  Side Key 2 (long)
  u8 unknown_3[5];    // 00D8 - 00DC
  u8 password[6];     // 00DD - 00E2  Password
} settings;
`

const rt21hFormat = `
struct {
  ul24 rxfreq;        // 0-2
  ul24 txfreq;        // 3-5
  lbcd rx_tone[2];    // 6-7
  lbcd tx_tone[2];    // 8-9
  u8 speccode:1,      // A   Spec Code  0 = On, 1 = Off
     compand:1,       //     Compander  0 = On, 1 = Off
     scramb:1,        //     Scramble  0 = Off, 1 = On
     scanadd:1,       //     Scan Add  0 = Scan, 1 = Skip
     ishighpower:1,   //     Power Level  0 = Low, 1 = High
     narrow:1,        //     Bandwidth  0 = Wide, 1 = Narrow
     unknown1:1,      //
     bcl:1;           //     Busy Channel Lockout  0 = On, 1 = Off
  u8 scrambopt;       // B   Scramble Options
} memory[16];

struct {
  u8 unknown_4:1,                  // 00C0
     scanm:1,                      //       Scan Mode
     roger:1,                      //       Roger
     lowbattwarn:1,                //       Low Battery Warning
     unknown_0:1,
     voice:1,                      //       Voice Annunciation
     save:1,                       //       Battery Save
     voxs:1;                       //       VOX Switch
  u8 squelch;                      // 00C1  Squelch Level
  u8 tot;                          // 00C2  Time-out Timer
  u8 voxd;                         // 00C3  Vox Delay
  u8 vox_level;                    // 00C4  VOX Level
  u8 unknown_1;
  u8 unknown_2;
  u8 skey1L:4,                     // 00C7  Side Key 1 (long)
     skey1S:4;                     //       Side Key 1 (short)
  u8 unknown_3:5,                  // 00C8
     removectdcs:1,                //       Remove CT/DCS
     beep:1,                       //       Beep Tone
     specmode:1;                   //       Spec Mode
} settings;
`

type h777v4Fam struct{}

type h777v4Opt struct {
	rt21h   bool
	cmd     string
	idents  [][]byte
	bs      int
	memsize int
	base    int64 // CHIRP base_freq_offset
	bits    int   // bit della frequenza
}

func init() {
	genRegister(&genModel{ID: "rt-h777v4", Vendor: "Retevis", Model: "H777", Variant: "V4", RClass: "H777V4", Name: "Retevis H777 V4 (nuova versione)",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
		Opt: &h777v4Opt{cmd: "C777HAM", idents: [][]byte{bytes.Repeat([]byte{0}, 6), bytes.Repeat([]byte{0xFF}, 6)}, bs: 0x0D, memsize: 0xEA, bits: 32},
		Notes: "Nuova versione della Retevis H777 (versioni FRS e PMR446, CHIRP permette tutta la gamma 400-520 MHz): 16 canali senza nomi; " +
			"blocco del canale occupato e scrambler si conservano. Driver di CHIRP."})
	genRegister(&genModel{ID: "rt-rt21h", Vendor: "Retevis", Model: "RT21H", RClass: "RT21HRadio", Name: "Retevis RT21H",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 502000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
		Opt: &h777v4Opt{rt21h: true, cmd: "T210GAM", idents: [][]byte{[]byte("SMP558\x02\xFF"), []byte("SMP558\x02\x00")}, bs: 0x0C, memsize: 0xD8,
			base: 0x2000000, bits: 24},
		Notes: "Versioni FRS e PMR446 (CHIRP permette tutta la gamma 400-502 MHz): 16 canali senza nomi; blocco del canale occupato, scrambler " +
			"(con le 8 varianti), compander e codice speciale si conservano. Driver di CHIRP."})
}

func h777v4o(m *genModel) *h777v4Opt { return m.Opt.(*h777v4Opt) }

func (h777v4Fam) mem(m *genModel) *bwEl {
	if h777v4o(m).rt21h {
		return bwMust(rt21hFormat)
	}
	return bwMust(h777v4Format)
}
func (f h777v4Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f h777v4Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f h777v4Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f h777v4Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f h777v4Fam) Match(m *genModel, img []byte) bool                { return false }

func h777v4ToneDec(b []byte) string {
	if b[0] == 0xFF {
		return ""
	}
	raw := int64(b[1])<<8 | int64(b[0])
	if b[1]&0x80 != 0 {
		n, _ := strconv.Atoi(strconv.FormatInt(raw&0x1FF, 8))
		if !slices.Contains(uv5rDTCS, n) {
			return ""
		}
		return genTone{Mode: "DTCS", Code: n, Inv: b[1]&0xC0 == 0xC0}.String()
	}
	v := int(b[1]>>4)*1000 + int(b[1]&0xF)*100 + int(b[0]>>4)*10 + int(b[0]&0xF)
	if !slices.Contains(chirpTones, v) {
		return ""
	}
	return fmtTone(v)
}

func h777v4ToneEnc(s string) ([]byte, error) {
	t, err := genToneParse(s)
	if err != nil {
		return nil, err
	}
	var v int64 = 0xFFFF
	switch t.Mode {
	case "Tone":
		v = thHex(t.Hz10)
	case "DTCS":
		v, _ = strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		v |= 0x8000
		if t.Inv {
			v |= 0x4000
		}
	}
	return []byte{byte(v), byte(v >> 8)}, nil
}

func (f h777v4Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	o := h777v4o(m)
	if img[r.Off] == 0xFF {
		return nil
	}
	c := &Chan{Rx: (r.F("rxfreq").Int(img) + o.base) * 10}
	if r.F("txfreq").Int(img) == 1<<o.bits-1 {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = (r.F("txfreq").Int(img) + o.base) * 10
	}
	c.TxTone = h777v4ToneDec(r.F("tx_tone").Raw(img))
	c.RxTone = h777v4ToneDec(r.F("rx_tone").Raw(img))
	c.Scan = !r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("ishighpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

func (f h777v4Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := h777v4o(m)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	enc := func(hz int64) (int64, error) {
		v := hz/10 - o.base
		if v < 0 || v > 1<<o.bits-1 {
			return 0, fmt.Errorf("frequenza %s MHz fuori dalla memoria della radio", fmtMHzShort(hz))
		}
		return v, nil
	}
	rx, err := enc(c.Rx)
	if err != nil {
		return err
	}
	tx := int64(1<<o.bits - 1)
	if !c.TxOff {
		if tx, err = enc(c.Tx); err != nil {
			return err
		}
	}
	tt, err := h777v4ToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := h777v4ToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	if o.rt21h && f.Get(m, img, i) != nil && r.F("scrambopt").Int(img) > 7 {
		// come CHIRP: variante dello scrambler fuori dalle 8 ammesse → la prima
		r.F("scrambopt").SetInt(img, 0)
	}
	r.F("rxfreq").SetInt(img, rx)
	r.F("txfreq").SetInt(img, tx)
	r.F("tx_tone").SetRaw(img, tt)
	r.F("rx_tone").SetRaw(img, rt)
	r.F("scanadd").SetBool(img, !c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// h777v4Enter: come _h777_enter_programming_mode di h777.py con il comando del modello
func (s *rawSess) h777v4Enter(m *genModel) error {
	o := h777v4o(m)
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if err := s.sleep(100 * time.Millisecond); err != nil {
		return err
	}
	if err := s.send([]byte(o.cmd)); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, 500*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, 500*time.Millisecond)
	if len(id) == 0 {
		return rErr("la radio non risponde")
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 500*time.Millisecond); len(b) != 1 || b[0] != 0x06 {
		return rErr("nessuna conferma dopo l'identificazione")
	}
	for _, x := range o.idents {
		if bytes.Contains(id, x) {
			return nil
		}
	}
	return rErr("modello errato: la radio si identifica come %q", id)
}

func (f h777v4Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := h777v4o(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	if err := s.h777v4Enter(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < o.memsize; a += o.bs {
		prog(a, o.memsize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), byte(o.bs)}
		if err := s.send(cmd); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 4+o.bs, time.Second)
		if len(resp) != 4+o.bs || !bytes.Equal(resp[:4], []byte{'W', cmd[1], cmd[2], cmd[3]}) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, resp[4:]...)
	}
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, "", nil
}

func (f h777v4Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := h777v4o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.memsize)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	if err := s.h777v4Enter(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img, 0, o.memsize, o.bs, time.Second, prog)
}

// fit: frequenze che la memoria della radio può contenere (RT21H: da 335,54432 MHz)
func (f h777v4Fam) fit(m *genModel, c *Chan) (bool, string) {
	o := h777v4o(m)
	ok := func(hz int64) bool { v := hz/10 - o.base; return v >= 0 && v < 1<<o.bits-1 }
	if !ok(c.Rx) || (!c.TxOff && !ok(c.Tx)) {
		return false, "frequenza che questa radio non può memorizzare"
	}
	return true, ""
}
