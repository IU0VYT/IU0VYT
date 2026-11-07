// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rb17p.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RB17P (retevis_rb17p.py di CHIRP), portatile GMRS: 128 canali con
// nomi di 10 caratteri. 9600 baud; 02, «PGA588» → 06, «M 02» → 8 × FF, 06 → 06;
// blocchi 'R' da 64 byte senza conferma, 'W' con conferma. Le regole GMRS sono
// quelle di CHIRP: fuori dalle frequenze GMRS la trasmissione è spenta, sui
// canali dei ponti solo simplex o +5 MHz, sui 467 MHz banda stretta e bassa potenza.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rb17pFam struct{}

const rb17pFormat = `
#seekto 0x0010;
struct {
  char name[10];
} names[128];

#seekto 0x1000;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 unknown1:4,
       compander:1,
       unknown2:1,
       highpower:1,
       unknown3:1;
    u8 unknown4:3,
       wide:1,
       scan:1,
       unknown5:1,
       bcl:2;
    u8 unknown6[2];
} memory[128];
`

func init() {
	genRegister(&genModel{ID: "rt-rb17p", Vendor: "Retevis", Model: "RB17P", RClass: "RB17PRadio", Name: "Retevis RB17P (GMRS)",
		Slots: 128, First: 1, NameMax: 10, Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}}, Bands: []band{{400000000, 470000000}},
		Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600, Fam: rb17pFam{},
		Notes: "Radio GMRS per gli Stati Uniti: in Italia la banda GMRS non è ammessa. Come in CHIRP: fuori dalle " +
			"frequenze GMRS la trasmissione è spenta, sui canali dei ponti solo simplex o +5 MHz, sui 467 MHz banda " +
			"stretta e bassa potenza. Blocco del canale occupato e compander si conservano. Driver di CHIRP."})
}

func (rb17pFam) mem() *bwEl                 { return bwMust(rb17pFormat) }
func (f rb17pFam) layout(m *genModel) *bwEl { return f.mem() }
func (f rb17pFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("names").I(i)
}
func (f rb17pFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f rb17pFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 26 {
		n.SetRaw(img, raw[16:26])
	}
}
func (f rb17pFam) Match(m *genModel, img []byte) bool { return false }

// rb17pView: il canale come lo mostra CHIRP (regole GMRS)
func rb17pView(m *genModel, c *Chan) {
	switch {
	case inI64(btGmrsF1, c.Rx):
		if !c.TxOff {
			c.Tx = c.Rx
		}
	case inI64(btGmrsF2, c.Rx):
		if !c.TxOff {
			c.Tx = c.Rx
		}
		c.Narrow, c.Power = true, m.powerSem(1)
	case inI64(btGmrsF3, c.Rx):
		if c.TxOff || c.Tx <= c.Rx {
			c.Tx, c.TxOff = c.Rx, false
		} else {
			c.Tx = c.Rx + 5000000
		}
	default:
		c.Tx, c.TxOff = c.Rx, true
	}
}

func (f rb17pFam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	rx := r.F("rxfreq").Int(img) * 10
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), ff) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = pyName(n.F("name").Raw(img), true)
	c.Narrow = !r.F("wide").Bool(img)
	c.Scan = r.F("scan").Bool(img)
	c.TxTone = x3ToneDec(r.F("txtone").Int(img))
	c.RxTone = x3ToneDec(r.F("rxtone").Int(img))
	if r.F("highpower").Bool(img) {
		c.Power = m.powerSem(0)
	} else {
		c.Power = m.powerSem(1)
	}
	rb17pView(m, c)
	return c
}

func (f rb17pFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
		return nil
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	var bcl, comp int64
	if f.Get(m, img, i) != nil {
		bcl, comp = r.F("bcl").Int(img), r.F("compander").Int(img)
	}
	r.Fill(img, 0)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	nm := n.F("name")
	for k := 0; k < 10; k++ {
		v := byte(0xFF)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	if err := x3ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 0)
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("scan").SetBool(img, c.Scan)
	r.F("bcl").SetInt(img, bcl)
	r.F("compander").SetInt(img, comp)
	return nil
}

// Locked: canali che la radio (e CHIRP) mostrerebbe in un altro modo
func (f rb17pFam) Locked(m *genModel, i int, old, want *Chan) error {
	if want == nil {
		return nil
	}
	v := *want
	rb17pView(m, &v)
	n := i + 1
	switch {
	case v.TxOff && !want.TxOff:
		return fmt.Errorf("canale %d: %s MHz non è una frequenza GMRS: imposta TX OFF", n, fmtMHz(want.Rx))
	case !v.TxOff && want.TxOff:
		return fmt.Errorf("canale %d: sulle frequenze GMRS dei ponti la trasmissione non si può disabilitare", n)
	case !v.TxOff && v.Tx != want.Tx:
		return fmt.Errorf("canale %d: su questa frequenza GMRS solo simplex o +5 MHz", n)
	case v.Narrow != want.Narrow || v.Power != want.Power:
		return fmt.Errorf("canale %d: sui canali GMRS 467 MHz solo banda stretta e bassa potenza", n)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rb17pEnter() error {
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if err := s.sleep(10 * time.Millisecond); err != nil {
		return err
	}
	if err := s.send([]byte("PGA588")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.send([]byte("M\x02")); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, time.Second); !bytes.Equal(id, bytes.Repeat([]byte{0xFF}, 8)) {
		return rErr("identificativo sconosciuto (% X)", id)
	}
	return s.ack2()
}

func (f rb17pFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x1800, "Collegamento con la radio…")
	if err := s.rb17pEnter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x1800; a += 0x40 {
		prog(a, 0x1800, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x40}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x44, time.Second)
		if len(b) != 0x44 || !bytes.Equal(b[:4], []byte{'W', byte(a >> 8), byte(a), 0x40}) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, b[4:]...)
	}
	prog(0x1800, 0x1800, "Lettura completata")
	return data, "", nil
}

func (f rb17pFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x1800 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x1800)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x1800, "Collegamento con la radio…")
	if err := s.rb17pEnter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, rg := range [][3]int{{0x0010, 0x0510, 0x40}, {0x0810, 0x0830, 0x30}, {0x1000, 0x1800, 0x40}} {
		for a := rg[0]; a < rg[1]; a += rg[2] {
			prog(a, 0x1800, "Scrittura…")
			n := rg[2]
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), byte(n)}, img[a:a+n]...)); err != nil {
				return "", err
			}
			if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				if ctx.Err() != nil {
					return "", errCancelled
				}
				return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	prog(0x1800, 0x1800, "Scrittura completata")
	return "", nil
}
