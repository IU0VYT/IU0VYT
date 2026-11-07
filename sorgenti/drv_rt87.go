// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rt87.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RT87 (retevis_rt87.py di CHIRP), portatile impermeabile bibanda:
// 128 canali con nomi di 10 caratteri. 9600 baud; la radio ripete ogni byte
// ricevuto (eco); «PGM2017» → 06 30; blocchi 'R'/'W' da 64 byte, 'E' per uscire.
// Differenza voluta da CHIRP: CHIRP, a ogni modifica di un canale, aumenta di 1
// il valore dello scrambler (difetto del driver); qui si conserva.

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"
)

type rt87Fam struct{}

const rt87Format = `
struct memory {
  bbcd rx_freq[4];
  bbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown10:2,
     pttid:2,
     unknown11:1,
     highpower:1,
     option:2;
  u8 optsigtype:3,
     unknown20:1,
     narrow:1,
     vox:1,
     bcl:2;
  u8 unknown31:1,
     scanadd:1,
     unknown32:3,
     scrambler:3;
  u8 unknown4;
};

struct name {
  char name[10];
};

#seekto 0x0010;
struct memory channels[128];

#seekto 0x1900;
struct name names[128];
`

func init() {
	genRegister(&genModel{ID: "rt-rt87", Vendor: "Retevis", Model: "RT87", RClass: "RT87", Name: "Retevis RT87",
		Slots: 128, First: 1, NameMax: 10, Charset: charsetAlnum + "!\"#$%&'()*+,-./:;<=>?@[\\]^_`",
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 4 W", 4, 0}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS,
		Baud: 9600, Fam: rt87Fam{},
		Notes: "128 canali con nomi di 10 caratteri. Blocco del canale occupato, PTT ID, VOX, compander/scrambler si " +
			"conservano. Driver di CHIRP."})
}

func (rt87Fam) mem() *bwEl                 { return bwMust(rt87Format) }
func (f rt87Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt87Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("channels").I(i), b.F("names").I(i)
}
func (f rt87Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f rt87Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 26 {
		n.SetRaw(img, raw[16:26])
	}
}
func (f rt87Fam) Match(m *genModel, img []byte) bool { return false }

func rt87ToneDec(b []byte) string {
	raw := int(b[1])<<8 | int(b[0])
	switch {
	case b[0] == 0xFF:
		return ""
	case b[1]&0xC0 == 0xC0 || b[1]&0x80 != 0:
		code, _ := strconv.Atoi(fmt.Sprintf("%x", raw&0x3FFF))
		return genTone{Mode: "DTCS", Code: code, Inv: b[1]&0xC0 == 0xC0}.String()
	}
	hz10, _ := strconv.Atoi(fmt.Sprintf("%x", raw))
	return fmtTone(hz10)
}

func rt87ToneEnc(b []byte, t genTone) {
	v := int64(0xFFFF)
	switch t.Mode {
	case "Tone":
		v, _ = strconv.ParseInt(strconv.Itoa(t.Hz10), 16, 32)
	case "DTCS":
		v, _ = strconv.ParseInt(strconv.Itoa(t.Code), 16, 32)
		v |= 0x8000
		if t.Inv {
			v |= 0x4000
		}
	}
	b[0], b[1] = byte(v), byte(v>>8)
}

func (f rt87Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	raw := r.Raw(img)
	if bytes.HasPrefix(raw, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	c.Tx = c.Rx
	if bytes.Equal(r.F("tx_freq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = r.F("tx_freq").Int(img) * 10
	}
	c.TxTone = rt87ToneDec(raw[10:12])
	c.RxTone = rt87ToneDec(raw[8:10])
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Name = string(bytes.TrimRight(n.F("name").Raw(img), "\xFF "))
	return c
}

func (f rt87Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
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
	ks := []string{"bcl", "pttid", "vox", "option", "scrambler"}
	var x [5]int64
	if f.Get(m, img, i) != nil {
		for k, s := range ks {
			x[k] = r.F(s).Int(img)
		}
	}
	r.Fill(img, 0)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("tx_freq").Fill(img, 0xFF)
	} else {
		r.F("tx_freq").SetInt(img, c.Tx/10)
	}
	raw := img[r.Off : r.Off+16]
	rt87ToneEnc(raw[10:12], tt)
	rt87ToneEnc(raw[8:10], rt)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("scanadd").SetBool(img, c.Scan)
	r.F("highpower").SetInt(img, int64(m.powerIndex(c.Power)))
	nm := bytes.TrimRight([]byte(c.Name), " ")
	nf := n.F("name")
	for k := 0; k < 10; k++ {
		v := byte(0xFF)
		if k < len(nm) {
			v = nm[k]
		}
		img[nf.Off+k] = v
	}
	for k, s := range ks {
		r.F(s).SetInt(img, x[k])
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

// rt87Send: invio con l'eco della radio
func (s *rawSess) rt87Send(b []byte) error {
	if err := s.send(b); err != nil {
		return err
	}
	echo := readUpTo(s.ctx, s.p, len(b), time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if !bytes.Equal(echo, b) {
		if len(echo) == 0 {
			return rErr("la radio non risponde")
		}
		return rErr("eco della radio non corretta")
	}
	return nil
}

func (s *rawSess) rt87Ident() error {
	if err := s.rt87Send([]byte("PGM2017")); err != nil {
		return err
	}
	ack := readUpTo(s.ctx, s.p, 2, time.Second)
	if len(ack) == 0 {
		return rErr("la radio non risponde")
	}
	if !bytes.Equal(ack, []byte{0x06, 0x30}) {
		return rErr("la radio non ha confermato il primo comando (% X)", ack)
	}
	return nil
}

func (f rt87Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rt87Ident(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x2000; a += 0x40 {
		prog(a, 0x2000, "Lettura…")
		if err := s.rt87Send([]byte{'R', byte(a >> 8), byte(a), 0x40}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x44, time.Second)
		if len(b) != 0x44 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("la radio ha inviato un blocco incompleto (0x%04X)", a)
		}
		data = append(data, b[4:]...)
	}
	s.send([]byte{'E'})
	prog(0x2000, 0x2000, "Lettura completata")
	return data, "", nil
}

func (f rt87Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rt87Ident(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, rg := range [][3]int{{0x0000, 0x0A00, 0x40}, {0x0BC0, 0x0C00, 0x08}, {0x0C00, 0x0F00, 0x40}, {0x1000, 0x2000, 0x40}} {
		for a := rg[0]; a < rg[1]; a += rg[2] {
			prog(a, 0x2000, "Scrittura…")
			n := rg[2]
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), byte(n)}, img[a:a+n]...)); err != nil {
				return "", err
			}
			readUpTo(ctx, s.p, n+4, time.Second) // eco
			if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				if ctx.Err() != nil {
					return "", errCancelled
				}
				return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	s.send([]byte{'E'})
	prog(0x2000, 0x2000, "Scrittura completata")
	return "", nil
}
