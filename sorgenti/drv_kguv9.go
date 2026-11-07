// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// kguv9dplus.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Wouxun KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K e KG-UV9G Pro
// (kguv9dplus.py di CHIRP). 19200 baud, pacchetti «7D comando FF lunghezza dati
// somma»: dati e somma (di 4 bit) mascherati con uno XOR a catena che parte da
// 0x52. Memoria di 32 KiB; la scrittura invia solo le zone conosciute (tabelle
// config_map / config_map2 di CHIRP). 999 canali con nomi di 8 caratteri,
// modulazione FM, FM stretta e AM.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type wx9Fam struct{}

type wx9Opt struct {
	upload [][3]int // zone scritte: indirizzo, dimensione del blocco, numero di blocchi
}

const wx9Format = `
#seekto 0x860;
struct {
    u8 w_mode;
    u16 w_chan;
} a_conf;
#seekto 0x870;
struct {
    u8 w_mode;
    u16 w_chan;
} b_conf;
#seekto 0xa00;
struct {
    u32 rxfreq;
    u32 txfreq;
    u16 txtone;
    u16 rxtone;
    u8  bit7_5:3,
        qt:3,
        bit1_0:2;
    u8  bit7:1,
        scan:1,
        bit5:1,
        pwr:2,
        mod:1,
        fm_dev:2;
    u8  state;
    u8  c3;
} chan_blk[999];
#seekto 0x4900;
struct {
    char name[8];
} chan_name[999];
#seekto 0x7000;
struct {
    u8 ch_valid;
} chan_valid[999];
`

// zone scritte nella radio (config_map e config_map2 di kguv9dplus.py)
var (
	wx9Map = [][3]int{{0x40, 16, 1}, {0x60, 20, 1}, {0x74, 8, 1}, {0x740, 40, 1}, {0x780, 16, 1}, {0x790, 16, 1},
		{0x800, 16, 1}, {0x810, 16, 1}, {0x820, 16, 1}, {0x830, 16, 1}, {0x840, 16, 1}, {0x860, 16, 1}, {0x870, 16, 1},
		{0x880, 16, 1}, {0x890, 16, 1}, {0x8a0, 16, 1}, {0x8b0, 16, 1}, {0x8c0, 16, 1}, {0x8d0, 16, 1}, {0x8e0, 16, 1},
		{0x8f0, 16, 1}, {0x940, 64, 2}, {0xa00, 64, 249}, {0x4840, 48, 1}, {0x4900, 32, 249}, {0x6820, 24, 1},
		{0x7000, 8, 124}, {0x73E0, 1, 7}, {0x7400, 64, 5}}
	wx9Map2 = [][3]int{{0x40, 16, 1}, {0x50, 10, 1}, {0x60, 20, 1}, {0x74, 8, 1}, {0x7c, 4, 1}, {0x740, 40, 1},
		{0x780, 16, 1}, {0x790, 16, 1}, {0x800, 16, 1}, {0x810, 16, 1}, {0x820, 16, 1}, {0x830, 16, 1}, {0x840, 16, 1},
		{0x860, 16, 1}, {0x870, 16, 1}, {0x880, 16, 1}, {0x890, 16, 1}, {0x8a0, 16, 1}, {0x8b0, 16, 1}, {0x8c0, 16, 1},
		{0x8d0, 16, 1}, {0x8e0, 16, 1}, {0x8f0, 16, 1}, {0x940, 64, 2}, {0xa00, 64, 249}, {0x4840, 48, 1},
		{0x4900, 32, 249}, {0x6820, 24, 1}, {0x7000, 8, 124}, {0x73E0, 1, 7}, {0x7400, 64, 5}, {0x7600, 1, 1}}
)

const (
	wx9MemValid    = 0xfc
	wx9MemInvalid  = 0xff
	wx9ChanValid   = 0x41
	wx9ChanInvalid = 0xFC
)

func init() {
	notes := "Accendi la radio con il cavo collegato (19200 baud). 999 canali con nomi di 8 caratteri; " +
		"i canali nella gamma aeronautica (108-136 MHz) si scrivono in AM. Driver di CHIRP (sperimentale)."
	pw := []genPower{{"Bassa 1 W", 1, 2}, {"Media 2 W", 2, 1}, {"Alta 5 W", 5, 0}}
	b9d := []band{{108000000, 135999999}, {136000000, 180000000}, {230000000, 250000000}, {350000000, 400000000},
		{400000000, 520000000}, {700000000, 985000000}}
	b9px := []band{{108000000, 135997500}, {136000000, 180997500}, {219000000, 250997500}, {350000000, 399997500},
		{400000000, 512997500}, {700000000, 986997500}}
	mk := func(id, model, rclass string, bands []band, up [][3]int) *genModel {
		return &genModel{ID: id, Vendor: "Wouxun", Model: model, RClass: rclass, Name: "Wouxun " + model, Slots: 999, First: 1,
			NameMax: 8, Charset: charsetASCII, Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: dtcsCodes,
			Baud: 19200, Fam: wx9Fam{}, Opt: &wx9Opt{upload: up}, Notes: notes}
	}
	genRegister(
		mk("wx-kguv9dplus", "KG-UV9D Plus", "KGUV9DPlusRadio", b9d, wx9Map),
		mk("wx-kguv9px", "KG-UV9PX", "KGUV9PXRadio", b9px, wx9Map2),
		mk("wx-kguv9gx", "KG-UV9GX", "KGUV9GXRadio", b9px, wx9Map2),
		mk("wx-kguv9k", "KG-UV9K", "KGUV9KRadio", b9d, wx9Map),
		mk("wx-kguv9gpro", "KG-UV9G Pro", "KGUV9GProRadio", b9d, wx9Map),
	)
}

func (wx9Fam) mem() *bwEl                 { return bwMust(wx9Format) }
func (f wx9Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f wx9Fam) recs(i int) (*bwEl, *bwEl, *bwEl) {
	b := f.mem()
	return b.F("chan_blk").I(i), b.F("chan_name").I(i), b.F("chan_valid").I(i)
}
func (f wx9Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n, v := f.recs(i)
	return append(append(append([]byte{}, r.Raw(img)...), n.Raw(img)...), v.Raw(img)...)
}
func (f wx9Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n, v := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 25 {
		n.SetRaw(img, raw[16:24])
		v.SetRaw(img, raw[24:25])
	}
}

// Match: solo la KG-UV9D Plus si riconosce dal contenuto (le altre dai metadati)
func (f wx9Fam) Match(m *genModel, img []byte) bool {
	return m.ID == "wx-kguv9dplus" && len(img) == 0x8000 && strings.Contains(strings.ToLower(string(img[0x51:0x59])), "kg-uv9d")
}
func (wx9Fam) crossband() bool { return false }

func (f wx9Fam) empty(img []byte, i int) bool {
	r, _, _ := f.recs(i)
	rx := r.F("rxfreq").Int(img)
	return rx == 0xFFFFFFFF || rx > 99999999
}

func (f wx9Fam) Get(m *genModel, img []byte, i int) *Chan {
	if f.empty(img, i) {
		return nil
	}
	r, n, _ := f.recs(i)
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if tx := r.F("txfreq").Int(img); tx == 0xFFFFFFFF {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = tx * 10
	}
	var nm []byte
	for _, b := range n.F("name").Raw(img) {
		if b != 0 && b != 0xFF {
			nm = append(nm, b)
		}
	}
	c.Name = string(nm)
	c.TxTone = kenwoodToneDec(r.F("txtone").Int(img), 0x8000, 0x4000)
	c.RxTone = kenwoodToneDec(r.F("rxtone").Int(img), 0x8000, 0x4000)
	c.Scan = r.F("scan").Int(img) > 0
	c.Power = m.powerSem(int(r.F("pwr").Int(img)))
	c.Narrow = r.F("mod").Int(img) == 1 || r.F("fm_dev").Int(img) != 0
	return c
}

func (f wx9Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n, v := f.recs(i)
	b := f.mem()
	if c == nil {
		// il canale cancellato non resta come canale di lavoro (come CHIRP)
		for _, a := range []string{"a_conf", "b_conf"} {
			if b.F(a).F("w_chan").Int(img) == int64(i+1) {
				b.F(a).F("w_chan").SetInt(img, 1)
			}
		}
		r.Fill(img, 0xFF)
		n.F("name").Fill(img, 0)
		r.F("state").SetInt(img, wx9MemInvalid)
		v.F("ch_valid").SetInt(img, wx9ChanInvalid)
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
	// AM: nella gamma aeronautica, o se il canale era già in AM sulla stessa frequenza
	am := c.Rx >= 108000000 && c.Rx < 136000000
	if !f.empty(img, i) && r.F("mod").Int(img) == 1 && r.F("rxfreq").Int(img)*10 == c.Rx {
		am = true
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0xFFFFFFFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("scan").SetBool(img, c.Scan)
	switch {
	case am:
		r.F("mod").SetInt(img, 1)
		r.F("fm_dev").SetInt(img, 1)
	case c.Narrow:
		r.F("mod").SetInt(img, 0)
		r.F("fm_dev").SetInt(img, 1)
	default:
		r.F("mod").SetInt(img, 0)
		r.F("fm_dev").SetInt(img, 0)
	}
	r.F("txtone").SetInt(img, kenwoodToneEnc(tt, 0x8000, 0x4000, 0))
	r.F("rxtone").SetInt(img, kenwoodToneEnc(rt, 0x8000, 0x4000, 0))
	r.F("pwr").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("qt").SetInt(img, 0)
	r.F("bit5").SetInt(img, 0)
	nm := strings.TrimRight(c.Name, " \t\r\n\x00\xff")
	raw := make([]byte, 8)
	copy(raw, nm)
	n.F("name").SetRaw(img, raw)
	r.F("state").SetInt(img, wx9MemValid)
	v.F("ch_valid").SetInt(img, wx9ChanValid)
	return nil
}

// ---------------------------------------------------------------- protocollo

const (
	wx9CmdIdent  = 0x80
	wx9CmdHangup = 0x81
	wx9CmdRconf  = 0x82
	wx9CmdWconf  = 0x83
)

// wx9Pkt: _pkt_encode di kguv9dplus.py (somma di 4 bit, XOR a catena da 0x52)
func wx9Pkt(op byte, payload []byte) []byte {
	d := []byte{0x7d, op, 0xff, byte(len(payload))}
	cs := int(op) + 0xff + len(payload)
	for _, b := range payload {
		cs += int(b)
	}
	d = append(append(d, payload...), byte(cs&0xF))
	x := byte(0x52)
	for k := 4; k < len(d); k++ {
		d[k] ^= x
		x = d[k]
	}
	return d
}

// wx9Unpkt: _pkt_decode (dati e somma giusta)
func wx9Unpkt(d []byte) (op byte, payload []byte, ok bool) {
	x := byte(0x52)
	var pl []byte
	for _, b := range d[4:] {
		pl = append(pl, x^b)
		x = b
	}
	cs := int(d[1]) + int(d[2]) + int(d[3])
	for _, b := range pl[:len(pl)-1] {
		cs += int(b)
	}
	return d[1], pl[:len(pl)-1], cs&0xF == int(pl[len(pl)-1])
}

func (s *rawSess) wx9Read() (byte, []byte, bool, error) {
	h := readUpTo(s.ctx, s.p, 4, 2*time.Second)
	if s.ctx.Err() != nil {
		return 0, nil, false, errCancelled
	}
	if len(h) == 0 {
		return 0, nil, false, rErr("la radio non risponde")
	}
	if len(h) < 4 || h[0] != 0x7D || h[1] < 0x80 || h[1] > 0x85 || h[2] != 0 {
		return 0, nil, false, rErr("risposta della radio non valida (% X)", h)
	}
	rest := readUpTo(s.ctx, s.p, int(h[3])+1, 2*time.Second)
	if len(rest) != int(h[3])+1 {
		return 0, nil, false, rErr("risposta della radio incompleta")
	}
	op, pl, ok := wx9Unpkt(append(h, rest...))
	return op, pl, ok, nil
}

func (s *rawSess) wx9Ident() error {
	for k := 0; k < 10; k++ {
		if err := s.send(wx9Pkt(wx9CmdIdent, nil)); err != nil {
			return err
		}
		op, resp, ok, err := s.wx9Read()
		if err != nil {
			return err
		}
		if len(resp) == 0 {
			return rErr("la radio non risponde")
		}
		if len(resp) != 74 || op != wx9CmdIdent {
			continue
		}
		if !ok {
			if err := s.sleep(100 * time.Millisecond); err != nil {
				return err
			}
			continue
		}
		if string(resp[:7]) == "KG-UV9D" {
			return nil
		}
		return rErr("la radio risponde ma non è di questo modello (%q)", string(resp[:9]))
	}
	return rErr("identificazione non riuscita")
}

func (f wx9Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.wx9Ident(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	mem := make([]byte, 0x8000)
	for a := 0; a < 0x8000; a += 64 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0x8000, "Lettura…")
		if err := s.send(wx9Pkt(wx9CmdRconf, []byte{byte(a >> 8), byte(a), 64})); err != nil {
			return nil, "", err
		}
		op, resp, ok, err := s.wx9Read()
		if err != nil {
			return nil, "", err
		}
		if !ok || op != wx9CmdRconf || len(resp) < 2 || int(resp[0])<<8|int(resp[1]) != a {
			return nil, "", rErr("errore di trasmissione leggendo il blocco 0x%04X", a)
		}
		copy(mem[a:], resp[2:])
	}
	_ = s.send(wx9Pkt(wx9CmdHangup, nil))
	prog(0x8000, 0x8000, "Lettura completata")
	return mem, "", nil
}

func (f wx9Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x8000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x8000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.wx9Ident(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, z := range m.Opt.(*wx9Opt).upload {
		for a := z[0]; a < z[0]+z[1]*z[2]; a += z[1] {
			prog(a, 0x8000, "Scrittura…")
			if err := s.send(wx9Pkt(wx9CmdWconf, append([]byte{byte(a >> 8), byte(a)}, img[a:a+z[1]]...))); err != nil {
				return "", err
			}
			op, ack, ok, err := s.wx9Read()
			if err != nil {
				return "", err
			}
			if !ok || op != wx9CmdWconf || len(ack) < 2 || int(ack[0])<<8|int(ack[1]) != a {
				return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	_ = s.send(wx9Pkt(wx9CmdHangup, nil))
	prog(0x8000, 0x8000, "Scrittura completata")
	return "", nil
}
