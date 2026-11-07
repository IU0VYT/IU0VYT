// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rt76p.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RT76P (retevis_rt76p.py di CHIRP), portatile GMRS: 128 canali con
// nomi di 10 caratteri. 9600 baud; «PROGROMCD2U» (fino a 5 volte) → 06, 'F' →
// identificativo di 8 byte; blocchi 'R' da 64 byte (risposta «R…») e 'W' da 32.
// Come CHIRP, nella scrittura applica le regole GMRS: frequenze non GMRS solo in
// ricezione, canali dei ponti simplex o +5 MHz, 467 MHz a banda stretta e bassa
// potenza.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rt76pFam struct{}

const rt76pFormat = `
#seekto 0x0000;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:4,
     scode:4;
  u8 unknown2:6,
     pttid:2;
  u8 unknown3:7,
     lowpower:1;
  u8 ani:1,
     narrow:1,
     unknown4:2,
     bcl:1,
     scan:1,
     unknown5:1,
     compand:1;
} memory[128];

#seekto 0x0C00;
struct {
  char name[10];
  u8 unused[6];
} names[128];
`

func init() {
	genRegister(&genModel{ID: "rt-rt76p", Vendor: "Retevis", Model: "RT76P", RClass: "RT76PRadio", Name: "Retevis RT76P (GMRS)",
		Slots: 128, First: 1, NameMax: 10, Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS,
		Baud: 9600, Fam: rt76pFam{},
		Notes: "Radio GMRS per gli Stati Uniti: in Italia la banda GMRS non è ammessa. Come in CHIRP: fuori dalle frequenze " +
			"GMRS solo ricezione, sui canali dei ponti simplex o +5 MHz, sui 467 MHz banda stretta e bassa potenza. Blocco del " +
			"canale occupato, PTT ID, compander e ANI si conservano. Driver di CHIRP."})
}

func (rt76pFam) mem() *bwEl                 { return bwMust(rt76pFormat) }
func (f rt76pFam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt76pFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("names").I(i)
}
func (f rt76pFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f rt76pFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 32 {
		n.SetRaw(img, raw[16:32])
	}
}
func (f rt76pFam) Match(m *genModel, img []byte) bool { return false }

func (f rt76pFam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	rx := r.F("rxfreq").Int(img)
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else if r.F("txfreq").Int(img) > rx {
		c.Tx = c.Rx + 5000000
	}
	c.Name = pyName(n.F("name").Raw(img), true)
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	c.Scan = r.F("scan").Bool(img)
	c.Power = m.powerSem(int(r.F("lowpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

// rt76pRule: trasmissione, banda e potenza che CHIRP scrive per il canale
func rt76pRule(m *genModel, c *Chan) (tx int64, off, narrow bool, power int) {
	tx, off, narrow, power = c.Rx, false, c.Narrow, c.Power
	switch {
	case inI64(btGmrsF1, c.Rx):
	case inI64(btGmrsF2, c.Rx):
		narrow, power = true, m.powerSem(1)
	case inI64(btGmrsF3, c.Rx):
		if !c.TxOff && c.Tx > c.Rx && c.Tx-c.Rx <= 70000000 {
			tx = c.Rx + 5000000
		}
	default:
		off = true
	}
	return
}

func (f rt76pFam) Set(m *genModel, img []byte, i int, c *Chan) error {
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
	var x [5]int64
	ks := []string{"bcl", "pttid", "scode", "compand", "ani"}
	if f.Get(m, img, i) != nil {
		for k, s := range ks {
			x[k] = r.F(s).Int(img)
		}
	}
	tx, off, narrow, power := rt76pRule(m, c)
	r.Fill(img, 0)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if off {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, tx/10)
	}
	nm := n.F("name")
	for k := 0; k < 10; k++ {
		v := byte(0xFF)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	r.F("narrow").SetBool(img, narrow)
	r.F("lowpower").SetBool(img, m.powerIndex(power) == 1)
	for k, s := range ks {
		r.F(s).SetInt(img, x[k])
	}
	return nil
}

// Locked: canali che la radio (e CHIRP) scriverebbe in un altro modo
func (f rt76pFam) Locked(m *genModel, i int, old, want *Chan) error {
	if want == nil {
		return nil
	}
	tx, off, narrow, power := rt76pRule(m, want)
	n := i + 1
	switch {
	case off && !want.TxOff:
		return fmt.Errorf("canale %d: %s MHz non è una frequenza GMRS: imposta TX OFF", n, fmtMHz(want.Rx))
	case !off && want.TxOff:
		return fmt.Errorf("canale %d: sulle frequenze GMRS la trasmissione non si può disabilitare", n)
	case !off && tx != want.Tx:
		return fmt.Errorf("canale %d: su questa frequenza GMRS solo simplex o +5 MHz (ponti)", n)
	case narrow != want.Narrow || power != want.Power:
		return fmt.Errorf("canale %d: sui canali GMRS 467 MHz solo banda stretta e bassa potenza", n)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rt76pEnter() error {
	var ack []byte
	ok := false
	for k := 0; k < 5 && !ok; k++ {
		if err := s.send([]byte("PROGROMCD2U")); err != nil {
			return err
		}
		ack = readUpTo(s.ctx, s.p, 1, time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		ok = len(ack) == 1 && ack[0] == 0x06
	}
	if !ok {
		if len(ack) == 0 {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non è entrata in programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
	}
	if err := s.send([]byte{'F'}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, time.Second); !bytes.Equal(id, []byte{0x01, 0x36, 0x01, 0x74, 0x04, 0x00, 0x05, 0x20}) {
		return rErr("identificativo sconosciuto (% X)", id)
	}
	return nil
}

func (f rt76pFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rt76pEnter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x2000; a += 0x40 {
		prog(a, 0x2000, "Lettura…")
		h := []byte{'R', byte(a >> 8), byte(a), 0x40}
		if err := s.send(h); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x44, time.Second)
		if len(b) != 0x44 || !bytes.Equal(b[:4], h) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, b[4:]...)
	}
	s.send([]byte{'E'})
	prog(0x2000, 0x2000, "Lettura completata")
	return data, "", nil
}

func (f rt76pFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.rt76pEnter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, rg := range [][2]int{{0x0000, 0x0820}, {0x0C00, 0x1400}, {0x1A00, 0x1C20}} {
		for a := rg[0]; a < rg[1]; a += 0x20 {
			prog(a, 0x2000, "Scrittura…")
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x20}, img[a:a+0x20]...)); err != nil {
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
	s.send([]byte{'E'})
	prog(0x2000, 0x2000, "Scrittura completata")
	return "", nil
}
