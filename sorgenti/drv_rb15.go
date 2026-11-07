// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rb15.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RB15 e RB615 (retevis_rb15.py di CHIRP): 99 canali senza nome,
// 400-480 MHz. 9600 baud; chiave «21 05 10 x» (fino a 5 volte) → 06; blocchi
// 'R'/'W' da 16 byte con indirizzo e somma di controllo; «21 05 EE V» per uscire.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rb15Fam struct{}

const rb15Format = `
struct memory {
  u32 rxfreq;
  u16 decQT;
  u32 txfreq;
  u16 encQT;
  u8 lowpower:1,
     unknown1:1,
     isnarrow:1,
     bcl:2,
     scan:1,
     encode:1,
     isunused:1;
  u8 unknown3[3];
};

#seekto 0x0170;
struct memory channels[99];
`

func init() {
	for _, x := range [][3]string{{"rt-rb15", "RB15", "RB15Radio"}, {"rt-rb615", "RB615", "RB615RadioBase"}} {
		genRegister(&genModel{ID: x[0], Vendor: "Retevis", Model: x[1], RClass: x[2], Name: "Retevis " + x[1], Slots: 99,
			First: 1, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}, Bands: []band{{400000000, 480000000}},
			Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: rb15Fam{},
			Notes: "99 canali senza nome, 400-480 MHz (venduta come radio " + map[string]string{"RB15": "FRS", "RB615": "PMR"}[x[1]] +
				" ma programmabile su tutta la gamma: in trasmissione rispetta le bande consentite). Blocco del canale " +
				"occupato e codifica si conservano. Driver di CHIRP."})
	}
}

func (rb15Fam) mem() *bwEl                 { return bwMust(rb15Format) }
func (f rb15Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rb15Fam) rec(i int) *bwEl          { return f.mem().F("channels").I(i) }
func (f rb15Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f rb15Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f rb15Fam) Match(m *genModel, img []byte) bool                { return false }

func rb15ToneDec(v int64) string {
	if v == 0xC000 || v == 0xFFFF {
		return ""
	}
	code := v & 0x3FFF
	switch {
	case v&0x4000 != 0:
		return genTone{Mode: "DTCS", Code: octDigits(int(code))}.String()
	case v&0x8000 != 0:
		return genTone{Mode: "DTCS", Code: octDigits(int(code)), Inv: true}.String()
	}
	return fmtTone(int(code))
}

// octDigits: 19 → 23 (il codice DCS scritto in ottale)
func octDigits(v int) int {
	r, mul := 0, 1
	for v > 0 {
		r += (v % 8) * mul
		v /= 8
		mul *= 10
	}
	return r
}

// octValue: 23 → 19
func octValue(d int) int {
	r, mul := 0, 1
	for d > 0 {
		r += (d % 10) * mul
		d /= 10
		mul *= 8
	}
	return r
}

func rb15ToneEnc(t genTone) int64 {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10)
	case "DTCS":
		if t.Inv {
			return int64(octValue(t.Code)) | 0x8000
		}
		return int64(octValue(t.Code)) | 0x4000
	}
	return 0xC000
}

func (f rb15Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), ff) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("isnarrow").Bool(img)
	c.TxTone = rb15ToneDec(r.F("encQT").Int(img))
	c.RxTone = rb15ToneDec(r.F("decQT").Int(img))
	c.Power = m.powerSem(int(r.F("lowpower").Int(img)))
	c.Scan = r.F("scan").Bool(img)
	return c
}

func (f rb15Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
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
	r.F("isunused").SetInt(img, 0)
	r.F("unknown1").SetInt(img, 0)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("scan").SetBool(img, c.Scan)
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("decQT").SetInt(img, rb15ToneEnc(rt))
	r.F("encQT").SetInt(img, rb15ToneEnc(tt))
	r.F("lowpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rb15Enter() error {
	var ack []byte
	for k := 0; k < 5; k++ {
		if err := s.send([]byte("21\x05\x10x")); err != nil {
			return err
		}
		ack = readUpTo(s.ctx, s.p, 1, 750*time.Millisecond)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(ack) == 1 && ack[0] == 0x06 {
			return nil
		}
	}
	if len(ack) == 0 {
		return rErr("la radio non risponde")
	}
	return rErr("la radio non è entrata in programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
}

func (f rb15Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x7A0, "Collegamento con la radio…")
	if err := s.rb15Enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x7A0; a += 0x10 {
		prog(a, 0x7A0, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a)}
		if err := s.send(append(cmd, tdm11Sum(cmd))); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 3+0x10+1, 500*time.Millisecond)
		if len(b) != 0x14 || !bytes.Equal(b[:3], cmd) || b[0x13] != tdm11Sum(b[:0x13]) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, b[3:0x13]...)
	}
	s.send([]byte("21\x05\xEEV"))
	prog(0x7A0, 0x7A0, "Lettura completata")
	return data, "", nil
}

func (f rb15Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x7A0 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x7A0)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x7A0, "Collegamento con la radio…")
	if err := s.rb15Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0x150; a < 0x7A0; a += 0x10 {
		prog(a, 0x7A0, "Scrittura…")
		b := append([]byte{'W', byte(a >> 8), byte(a)}, img[a:a+0x10]...)
		if err := s.send(append(b, tdm11Sum(b))); err != nil {
			return "", err
		}
		if r := readUpTo(ctx, s.p, 1, 500*time.Millisecond); len(r) != 1 || r[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	s.send([]byte("21\x05\xEEV"))
	prog(0x7A0, 0x7A0, "Scrittura completata")
	return "", nil
}
