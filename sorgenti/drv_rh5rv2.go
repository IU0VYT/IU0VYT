// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// rh5r_v2.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Rugged RH5R-V2 (rh5r_v2.py di CHIRP, base TYT TH-UVF8F V2): portatile
// bibanda costruito in Cina, 128 canali con nomi di 7 caratteri. 9600 baud;
// «PGM2015» → 06 30; blocchi 'R'/'W' da 64 byte, 'E' per uscire. L'immagine
// finisce con «PGM2015».

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rh5rFam struct{ fileid string }

const rh5rFormat = `
struct memory {
  bbcd rx_freq[4];
  bbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown10:5,
     highpower:1,
     unknown11:2;
  u8 unknown20:4,
     narrow:1,
     unknown21:3;
  u8 unknown31:1,
     scanadd:1,
     unknown32:6;
  u8 unknown4;
};

struct name {
  char name[7];
};

#seekto 0x0010;
struct memory channels[128];

#seekto 0x08C0;
struct name names[128];
`

func init() {
	genRegister(&genModel{ID: "rg-rh5rv2", Vendor: "Rugged", Model: "RH5R-V2", RClass: "RH5RV2", Name: "Rugged RH5R-V2",
		Slots: 128, First: 1, NameMax: 7, Charset: charsetUpperNum + "-",
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS,
		Baud: 9600, Fam: rh5rFam{fileid: "RUGGED \xFF"},
		Notes: "Portatile bibanda (versione Rugged della TYT TH-UVF8): 128 canali con nomi di 7 caratteri (maiuscole, " +
			"cifre e trattino); senza spostamenti oltre 70 MHz (split). Driver di CHIRP."})
}

func (rh5rFam) mem() *bwEl                 { return bwMust(rh5rFormat) }
func (f rh5rFam) layout(m *genModel) *bwEl { return f.mem() }
func (f rh5rFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("channels").I(i), b.F("names").I(i)
}
func (f rh5rFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f rh5rFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 23 {
		n.SetRaw(img, raw[16:23])
	}
}
func (f rh5rFam) Match(m *genModel, img []byte) bool {
	return len(img) >= 0x848 && bytes.HasSuffix(img, []byte("PGM2015")) && string(img[0x840:0x848]) == f.fileid
}
func (rh5rFam) crossband() bool { return true }

// fit (4.9, convertitore): canali con uno spostamento oltre 70 MHz non copiati
// (come CHIRP)
func (rh5rFam) fit(m *genModel, c *Chan) (bool, string) {
	if d := c.Tx - c.Rx; !c.TxOff && (d > 70000000 || d < -70000000) {
		return false, "questa radio non ammette spostamenti oltre 70 MHz (split)"
	}
	return true, ""
}

func (f rh5rFam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	raw := r.Raw(img)
	if bytes.HasPrefix(raw, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	c.Tx = c.Rx
	if tx := r.F("tx_freq").Int(img); tx == 0 {
		c.TxOff = true
	} else {
		c.Tx = tx * 10
	}
	c.TxTone = rt87ToneDec(raw[10:12])
	c.RxTone = rt87ToneDec(raw[8:10])
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Name = string(bytes.TrimRight(n.F("name").Raw(img), "\xFF "))
	return c
}

func (f rh5rFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
		return nil
	}
	if d := c.Tx - c.Rx; !c.TxOff && (d > 70000000 || d < -70000000) {
		return fmt.Errorf("questa radio non ammette spostamenti oltre 70 MHz (split)")
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	r.Fill(img, 0)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	if !c.TxOff {
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
	for k := 0; k < 7; k++ {
		v := byte(0xFF)
		if k < len(nm) {
			v = nm[k]
		}
		img[nf.Off+k] = v
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rh5rIdent() error {
	if err := s.send([]byte("PGM2015")); err != nil {
		return err
	}
	ack := readUpTo(s.ctx, s.p, 2, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if !bytes.Equal(ack, []byte{0x06, 0x30}) {
		if len(ack) == 0 {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non ha confermato il primo comando (% X)", ack)
	}
	return nil
}

func (f rh5rFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	defer s.send([]byte{'E'})
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rh5rIdent(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x2000; a += 0x40 {
		prog(a, 0x2000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x40}); err != nil {
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
	prog(0x2000, 0x2000, "Lettura completata")
	return append(data, []byte("PGM2015")...), "", nil
}

func (f rh5rFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	defer s.send([]byte{'E'})
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rh5rIdent(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x2000; a += 0x40 {
		prog(a, 0x2000, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x40}, img[a:a+0x40]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(0x2000, 0x2000, "Scrittura completata")
	return "", nil
}
