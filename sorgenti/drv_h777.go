// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: h777.py.

package main

// Famiglia Baofeng BF-888S (driver h777.py di CHIRP): BF-888S, BF-M4, MP31,
// BF-1901, BF-1904, BF-1909. 16 canali (38 sulla MP31), senza nomi, 9600 baud.
//
// Protocollo: 0x02, pausa 100 ms, "PROGRAM" (o "PWPG970" sulle BF-19xx) →
// 0x06; 0x02 → identificativo (8 byte, es. "P3107…"); 0x06 → 0x06.
// lettura: 'R' indirizzo(2) 0x08 → 'W' indirizzo 0x08 + 8 byte; 0x06 → 0x06.
// scrittura: 'W' indirizzo 0x08 + 8 byte → 0x06. Fine: 'E'.

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"
)

const h777Format = `
#seekto 0x%04X;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 unknown3:1,
       unknown2:1,
       unknown1:1,
       skip:1,
       highpower:1,
       narrow:1,
       beatshift:1,
       bcl:1;
    u8 unknown4[3];
} memory[%d];
`

// subtoni CTCSS di CHIRP (chirp_common.TONES)
var chirpTones = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148,
	1188, 1230, 1273, 1318, 1365, 1413, 1462, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835,
	1862, 1899, 1928, 1966, 1995, 2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

// subtono più vicino tra quelli di CHIRP (a parità, il primo)
func chirpClosestTone(t10 float64) int {
	best, bd := chirpTones[0], math.Inf(1)
	for _, v := range chirpTones {
		if d := math.Abs(float64(v) - t10); d < bd {
			best, bd = v, d
		}
	}
	return best
}

type h777Opt struct {
	cmd      []byte
	idents   [][]byte
	ranges   [][2]int
	memsize  int
	scramble bool // impostazione "beatshift" conservata come quella del canale
	match    bool
	base     int // indirizzo del primo canale
}

type h777Fam struct{}

func init() {
	pw := func(h float64) []genPower {
		return []genPower{{"Bassa 1 W", 1, 2}, {fmt.Sprintf("Alta %g W", h), h, 0}}
	}
	mk := func(id, model, rclass, name string, slots int, o h777Opt, p []genPower, bands []band) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: slots, First: 1,
			NameMax: 0, Powers: p, Bands: bands, Narrow: true, HasScan: true, DCS: dtcsAll(), Baud: 9600,
			Fam: h777Fam{}, Opt: &oo, Notes: "Radio senza display: i canali non hanno nome."}
	}
	std := h777Opt{cmd: []byte("PROGRAM"), idents: [][]byte{[]byte("P3107")}, ranges: [][2]int{{0x0000, 0x0110}, {0x0380, 0x03E0}, {0x02B0, 0x02C0}},
		memsize: 0x03E0, scramble: true, base: 0x0010}
	o888 := std
	o888.match = true
	b19 := std
	b19.cmd, b19.scramble = []byte("PWPG970"), false
	b1909 := b19
	b1909.idents = [][]byte{[]byte("P320h"), []byte("P3107\xF4AM")}
	b1909.ranges = [][2]int{{0x0000, 0x0110}, {0x0250, 0x0260}, {0x02B0, 0x02C0}, {0x03C0, 0x03E0}}
	b1909.memsize = 0x03F0
	mp := std
	mp.idents, mp.ranges, mp.memsize, mp.base = [][]byte{[]byte("P3107\xf7\x00\x00")}, [][2]int{{0x0000, 0x0400}}, 0x0400, 0
	genRegister(
		mk("bf-888s", "BF-888", "H777Radio", "Baofeng BF-888S", 16, o888, pw(5), []band{{400000000, 490000000}}),
		mk("bf-m4", "BF-M4", "BFM4Radio", "Baofeng BF-M4", 16, std, pw(5), []band{{400000000, 490000000}}),
		mk("bf-mp31", "MP31", "MP31Radio", "Baofeng MP31", 38, mp, pw(5), []band{{400000000, 490000000}}),
		mk("bf-1901", "BF-1901", "BF1901Radio", "Baofeng BF-1901", 16, b19, pw(5), []band{{400000000, 520000000}}),
		mk("bf-1904", "BF-1904", "BF1904Radio", "Baofeng BF-1904", 16, b19, pw(10), []band{{400000000, 520000000}}),
		mk("bf-1909", "BF-1909", "BF1909Radio", "Baofeng BF-1909", 16, b1909, pw(5), []band{{400000000, 520000000}}),
	)
	// Baofeng BF-C2 / BF-V2 e BF-88ST: i programmi del costruttore (BF_V1_FH,
	// «RT69») usano lo stesso protocollo, la stessa memoria e lo stesso
	// identificativo («P3107», riscritto al contrario in 0x3D0) della BF-888S.
	pmr := []genPower{{"Bassa", 0.5, 2}, {"Alta", 0.5, 0}}
	for _, v := range []struct{ id, model, name string }{
		{"bf-c2", "BF-C2", "Baofeng BF-C2 / BF-V2"},
		{"bf-88st", "BF-88ST", "Baofeng BF-88ST"},
	} {
		x := mk(v.id, v.model, "H777Radio", v.name, 16, std, pmr, []band{{400000000, 490000000}})
		x.NoChirp = true
		x.Notes = "Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore: " +
			"stessa memoria della BF-888S di CHIRP (nei file .img compare come BF-888)."
		genRegister(x)
	}
}

// tutti i codici DCS ottali a tre cifre (radio che accettano qualsiasi codice)
func dtcsAll() []int { return dtcsCodes }

func h777o(m *genModel) *h777Opt { return m.Opt.(*h777Opt) }

func (h777Fam) mem(m *genModel) *bwEl { return bwMust(fmt.Sprintf(h777Format, h777o(m).base, m.Slots)) }

func (f h777Fam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f h777Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

func (f h777Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}

func (f h777Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f h777Fam) Match(m *genModel, img []byte) bool {
	return h777o(m).match && len(img) == h777o(m).memsize
}

func bcdByte(b byte) int64 { return int64(b>>4)*10 + int64(b&0x0F) }

// tono (lbcd[2], bit 0x80 = DCS, 0x40 = invertito)
func h777ToneDec(raw []byte) string {
	if raw[0] == 0xFF && raw[1] == 0xFF {
		return ""
	}
	v := bcdByte(raw[0]) + 100*bcdByte(raw[1]&0x3F)
	if raw[1]&0x80 != 0 {
		return genTone{Mode: "DTCS", Code: int(v), Inv: raw[1]&0x40 != 0}.String()
	}
	t := float64(v)
	tone := int(v)
	if idxInt(chirpTones, tone) < 0 {
		tone = chirpClosestTone(t)
	}
	return fmtTone(tone)
}

func decToBCD(v int64) byte { return byte(v/10%10)<<4 | byte(v%10) }

func h777ToneEnc(dst []byte, t genTone) {
	switch t.Mode {
	case "":
		dst[0], dst[1] = 0xFF, 0xFF
	case "Tone":
		v := int64(t.Hz10)
		dst[1] &^= 0xC0
		dst[0] = decToBCD(v % 100)
		dst[1] = dst[1]&0xC0 | decToBCD(v/100%100)&^0xC0
	case "DTCS":
		dst[1] |= 0x80
		if t.Inv {
			dst[1] |= 0x40
		} else {
			dst[1] &^= 0x40
		}
		v := int64(t.Code)
		dst[0] = decToBCD(v % 100)
		dst[1] = dst[1]&0xC0 | decToBCD(v/100%100)&^0xC0
	}
}

func (f h777Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("narrow").Bool(img)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Scan = !r.F("skip").Bool(img)
	c.TxTone = h777ToneDec(r.F("txtone").Raw(img))
	c.RxTone = h777ToneDec(r.F("rxtone").Raw(img))
	return c
}

func (f h777Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
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
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	h777ToneEnc(r.F("txtone").Raw(img), tt)
	h777ToneEnc(r.F("rxtone").Raw(img), rt)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("skip").SetBool(img, !c.Scan)
	// bcl e beatshift (impostazioni del canale) restano come sono
	r.F("unknown1").SetInt(img, 0)
	r.F("unknown2").SetInt(img, 0)
	r.F("unknown3").SetInt(img, 0)
	return nil
}

// ---------------------------------------------------------------- protocollo

type h777Sess struct {
	ctx context.Context
	p   Port
}

func (s *h777Sess) sleep(d time.Duration) error {
	select {
	case <-s.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

func (s *h777Sess) read(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, to); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *h777Sess) enter(m *genModel) ([]byte, error) {
	o := h777o(m)
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	if err := s.sleep(100 * time.Millisecond); err != nil {
		return nil, err
	}
	if err := s.p.Write(o.cmd); err != nil {
		return nil, err
	}
	ack, err := s.read(1, 500*time.Millisecond)
	if err == errTimeout {
		return nil, rErr("la radio non risponde")
	}
	if err != nil {
		return nil, err
	}
	if ack[0] != 0x06 {
		return nil, rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	ident := readUpTo(s.ctx, s.p, 8, 500*time.Millisecond)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(ident) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if err := s.p.Write([]byte{0x06}); err != nil {
		return nil, err
	}
	ack, err = s.read(1, 500*time.Millisecond)
	if err != nil || ack[0] != 0x06 {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("nessuna conferma dopo l'identificazione")
	}
	ok := false
	for _, id := range o.idents {
		if bytes.Contains(ident, id) {
			ok = true
		}
	}
	if !ok {
		return ident, rErr("modello errato: la radio si identifica come %q", ident)
	}
	return ident, nil
}

func h777Open(ctx context.Context, port string) (*h777Sess, error) {
	p, err := openPortFunc(port, 9600)
	if err != nil {
		return nil, err
	}
	return &h777Sess{ctx: ctx, p: p}, nil
}

func (f h777Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := h777o(m)
	s, err := h777Open(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	ident, err := s.enter(m)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	info := fmt.Sprintf("identificativo %q", ident)
	var data []byte
	for a := 0; a < o.memsize; a += 8 {
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		prog(a, o.memsize, "Lettura…")
		if err := s.p.Write([]byte{'R', byte(a >> 8), byte(a), 8}); err != nil {
			return nil, info, err
		}
		resp, err := s.read(12, 500*time.Millisecond)
		if err != nil || !bytes.Equal(resp[:4], []byte{'W', byte(a >> 8), byte(a), 8}) {
			if err == errCancelled {
				return nil, info, err
			}
			return nil, info, rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.p.Write([]byte{0x06}); err != nil {
			return nil, info, err
		}
		ack, err := s.read(1, 500*time.Millisecond)
		if err != nil || ack[0] != 0x06 {
			if err == errCancelled {
				return nil, info, err
			}
			return nil, info, rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	if err := s.p.Write([]byte{'E'}); err != nil {
		return nil, info, err
	}
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, info, nil
}

func (f h777Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := h777o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := h777Open(ctx, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	ident, err := s.enter(m)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	info := fmt.Sprintf("identificativo %q", ident)
	for _, r := range o.ranges {
		for a := r[0]; a < r[1]; a += 8 {
			prog(a, o.memsize, "Scrittura…")
			if err := s.p.Write(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...)); err != nil {
				return info, err
			}
			ack, err := s.read(1, 500*time.Millisecond)
			if err != nil || ack[0] != 0x06 {
				if err == errCancelled {
					return info, err
				}
				return info, rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	if err := s.p.Write([]byte{'E'}); err != nil {
		return info, err
	}
	prog(o.memsize, o.memsize, "Scrittura completata")
	return info, nil
}
