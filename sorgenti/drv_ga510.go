// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: ga510.py.

package main

// Baofeng serie H (programma del costruttore «BAOFENG H-Series Programming
// Software», SQ5R): stesso protocollo e stessa memoria della Radioddity GA-510
// (driver ga510.py di CHIRP, classe RadioddityGA510Radio, prodotta da Baofeng).
// 128 canali (0-127) di 16 byte da 0x0000, nomi di 10 caratteri da 0x0C00,
// 9600 baud.
//
// Protocollo (start_program, do_download, do_upload di ga510.py):
//   'E' (uscita da un'eventuale sessione precedente), pausa;
//   "PROGROMBFHU" → … 0x06; 'F' → identificativo di 8 byte;
//   lettura: 'R' indirizzo(2, BE) 0x40 → 'R' indirizzo 0x40 + 64 byte (0x0000-0x1C3F);
//   scrittura: 'W' indirizzo 0x20 + 32 byte → 0x06 (0x0000-0x1C1F, come il
//   programma del costruttore); fine: 'E'.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const ga510Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 signal;
  u8 unknown1:6,
     pttid:2;
  u8 unknown2:6,
     power:2;
  u8 unknown3_0:1,
     narrow:1,
     unknown3_1:2,
     bcl:1,
     scan:1,
     allow_tx:1,
     fhss:1;
} memories[128];
#seekto 0x0C00;
struct {
  char name[10];
  u8 pad[6];
} names[128];
`

type ga510Opt struct {
	magic   []byte
	memsize int // area letta
	wsize   int // area scritta
}

type ga510Fam struct{}

func init() {
	o := &ga510Opt{magic: []byte("PROGROMBFHU"), memsize: 0x1C40, wsize: 0x1C20}
	genRegister(&genModel{ID: "bf-hseries", Vendor: "Radioddity", Model: "GA-510", RClass: "RadioddityGA510Radio",
		Name: "Baofeng serie H (H-31, H-32, BF-H…)", Slots: 128, First: 0, NameMax: 10,
		Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!\"#$%&'()~+-,./:;<=>?@[\\]^`{}*| ",
		Powers:  []genPower{{"Alta", 10, 0}, {"Bassa", 1, 2}, {"Media", 5, 1}},
		Bands:   []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 9600, Fam: ga510Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato. Canali 0-127. Il programma del costruttore della serie H " +
			"usa lo stesso protocollo e la stessa memoria della Radioddity GA-510 (prodotta da Baofeng): driver di CHIRP " +
			"della GA-510; nei file .img compare come Radioddity GA-510."})
}

func ga510o(m *genModel) *ga510Opt { return m.Opt.(*ga510Opt) }

func (ga510Fam) mem() *bwEl { return bwMust(ga510Format) }

func (f ga510Fam) layout(m *genModel) *bwEl { return f.mem() }

func (f ga510Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memories").I(i), b.F("names").I(i)
}

func (f ga510Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}

func (f ga510Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	n.SetRaw(img, raw[16:32])
}

func (f ga510Fam) Match(m *genModel, img []byte) bool { return false }

var ff4 = []byte{0xFF, 0xFF, 0xFF, 0xFF}

func (f ga510Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if bytes.Equal(r.F("rxfreq").Raw(img), ff4) {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), ff4) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	// come CHIRP: solo i caratteri < 127, spazi finali tolti
	var nb []byte
	for _, b := range n.F("name").Raw(img) {
		if b < 127 {
			nb = append(nb, b)
		}
	}
	c.Name = strings.TrimRight(string(nb), " \t\n\r\x0b\x0c")
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	pw := int(r.F("power").Int(img))
	if pw >= len(m.Powers) {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = r.F("scan").Bool(img)
	return c
}

func (f ga510Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
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
	if bytes.Equal(r.F("rxfreq").Raw(img), ff4) {
		r.Fill(img, 0) // nuovo canale (come CHIRP)
	}
	name := c.Name
	if len(name) > 10 {
		name = name[:10]
	}
	n.F("name").SetStr(img, name, ' ')
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("scan").SetBool(img, c.Scan)
	return nil
}

// ---------------------------------------------------------------- protocollo

const ga510Timeout = 1500 * time.Millisecond

// start_program di ga510.py
func (s *rawSess) ga510Start(o *ga510Opt) ([]byte, error) {
	if err := s.send([]byte{'E'}); err != nil {
		return nil, err
	}
	// la radio può rispondere con dei byte residui: si scartano
	_ = readUpTo(s.ctx, s.p, 256, 300*time.Millisecond)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if err := s.send(o.magic); err != nil {
		return nil, err
	}
	ack := readUpTo(s.ctx, s.p, 256, ga510Timeout)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(ack) == 0 {
		return nil, errTimeout
	}
	if ack[len(ack)-1] != 0x06 {
		return nil, rErr("risposta inattesa alla richiesta di programmazione (% X): è il modello giusto?", ack)
	}
	if err := s.send([]byte{'F'}); err != nil {
		return nil, err
	}
	id := make([]byte, 8)
	if err := s.p.ReadFull(s.ctx, id, ga510Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("identificativo della radio incompleto")
		}
		return nil, err
	}
	return id, nil
}

func (f ga510Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := ga510o(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	id, err := s.ga510Start(o)
	if err != nil {
		_ = s.send([]byte{'E'})
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < o.memsize; a += 0x40 {
		if ctx.Err() != nil {
			_ = s.send([]byte{'E'})
			return nil, "", errCancelled
		}
		prog(a, o.memsize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), 0x40}
		if err := s.send(cmd); err != nil {
			return nil, "", err
		}
		d := make([]byte, 0x44)
		err := s.p.ReadFull(ctx, d, ga510Timeout)
		if err == nil && (d[1] != cmd[1] || d[2] != cmd[2] || d[3] != 0x40) {
			err = rErr("risposta inattesa")
		}
		if err != nil {
			_ = s.send([]byte{'E'})
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, d[4:]...)
	}
	_ = s.send([]byte{'E'})
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, fmt.Sprintf("identificativo %q", id), nil
}

func (f ga510Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := ga510o(m)
	// immagini più corte (es. salvate da versioni di CHIRP che leggevano fino a
	// 0x1C00): si scrive solo quello che c'è
	wsize := min(o.wsize, len(img)/0x20*0x20)
	if wsize < 0x1000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, wsize, "Collegamento con la radio…")
	if _, err := s.ga510Start(o); err != nil {
		_ = s.send([]byte{'E'})
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < wsize; a += 0x20 {
		prog(a, wsize, "Scrittura…")
		if err := s.send([]byte{'W', byte(a >> 8), byte(a), 0x20}); err != nil {
			return "", err
		}
		if err := s.send(img[a : a+0x20]); err != nil {
			return "", err
		}
		ack, err := readByte(ctx, s.p, ga510Timeout)
		if err == nil && ack != 0x06 {
			err = rErr("la radio ha rifiutato il blocco")
		}
		if err != nil {
			_ = s.send([]byte{'E'})
			return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
	}
	_ = s.send([]byte{'E'})
	prog(wsize, wsize, "Scrittura completata")
	return "", nil
}
