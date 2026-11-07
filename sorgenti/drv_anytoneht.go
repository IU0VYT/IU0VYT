// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// anytone_ht.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: AnyTone TERMN-8R e OBLTR-8R (anytone_ht.py di CHIRP): 200 canali con
// nomi di 6 caratteri, frequenze in BCD a passi di 100 Hz (con i 50 Hz dei
// passi 6,25/12,5 kHz ricostruiti), tutti i 512 codici DCS. 9600 baud; «PROGRAM»
// → «QX 06», 02 → identificativo; blocchi 'R'/'W' da 16 byte con somma di
// controllo e 06; «END» per finire.
// Differenza voluta da CHIRP: CHIRP non toglie mai la sola ricezione (tx_off)
// da un canale; qui si toglie quando il canale torna a trasmettere.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type athtFam struct{ ident string }

const athtFormat = `
struct memory {
  bbcd freq[4];
  bbcd offset[4];
  u8 unknown1:4,
     tune_step:4;
  u8 unknown2:2,
     txdcsextra:1,
     txinv:1,
     channel_width:2,
     unknown3:1,
     tx_off:1;
  u8 unknown4:2,
     rxdcsextra:1,
     rxinv:1,
     power:2,
     duplex:2;
  u8 unknown5:4,
     rxtmode:2,
     txtmode:2;
  u8 unknown6:2,
     txtone:6;
  u8 unknown7:2,
     rxtone:6;
  u8 txcode;
  u8 rxcode;
  u8 unknown8[3];
  char name[6];
  u8 squelch:4,
     unknown9:2,
     bcl:2;
  u8 unknownA;
  u8 unknownB:7,
     sqlmode:1;
  u8 unknownC[4];
};

#seekto 0x0100;
u8 used_flags[50];

#seekto 0x0120;
u8 skip_flags[50];

#seekto 0x2000;
struct memory memory[200];
`

var (
	athtTones = append([]int{625}, chirpTones...)
	allDCS    = func() []int {
		var out []int
		for a := 0; a < 8; a++ {
			for b := 0; b < 8; b++ {
				for c := 0; c < 8; c++ {
					out = append(out, a*100+b*10+c)
				}
			}
		}
		return out
	}()
)

func init() {
	for _, x := range [][3]string{{"at-termn8r", "TERMN-8R", "AnyToneTERMN8RRadio"}, {"at-obltr8r", "OBLTR-8R", "AnyToneOBLTR8RRadio"}} {
		id := map[string]string{"TERMN-8R": "TERMN8R", "OBLTR-8R": "OBLTR8R"}[x[1]]
		genRegister(&genModel{ID: x[0], Vendor: "AnyTone", Model: x[1], RClass: x[2], Name: "AnyTone " + x[1], Slots: 200,
			First: 0, NameMax: 6, Charset: charsetASCII, Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2 W", 2, 1}, {"Bassa 1 W", 1, 2}},
			Bands: []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true, DCS: allDCS, Baud: 9600,
			Fam: athtFam{ident: id},
			Notes: "200 canali con nomi di 6 caratteri, tutti i 512 codici DCS e il tono 62,5 Hz. Blocco del canale occupato e " +
				"squelch del canale si conservano. Driver di CHIRP (sperimentale)."})
	}
}

func (athtFam) mem() *bwEl                 { return bwMust(athtFormat) }
func (f athtFam) layout(m *genModel) *bwEl { return f.mem() }
func (f athtFam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f athtFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f athtFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f athtFam) Match(m *genModel, img []byte) bool {
	return len(img) >= 0x20 && bytes.Contains(img[0x10:0x20], []byte(f.ident))
}

func (f athtFam) Get(m *genModel, img []byte, i int) *Chan {
	if img[0x100+i/8]&(1<<(i%8)) != 0 {
		return nil
	}
	r := f.rec(i)
	fq := r.F("freq").Int(img)
	c := &Chan{Rx: fq * 100}
	if d := fq % 10; d == 2 || d == 7 {
		c.Rx += 50
	}
	off := r.F("offset").Int(img) * 100
	c.Tx = c.Rx
	switch r.F("duplex").Int(img) {
	case 2:
		c.Tx = c.Rx - off
	case 3:
		c.Tx = c.Rx + off
	}
	if r.F("tx_off").Bool(img) {
		c.TxOff, c.Tx = true, c.Rx
	}
	var nm []byte
	for _, b := range bytes.TrimRight(r.F("name").Raw(img), " \t\n\r\x0b\x0c") {
		if b >= 0x20 && b < 0x7F {
			nm = append(nm, b)
		}
	}
	c.Name = string(nm)
	c.Narrow = r.F("channel_width").Int(img) == 1
	tone := func(mode, tv, code, extra, inv string) string {
		switch r.F(mode).Int(img) {
		case 1:
			if k := int(r.F(tv).Int(img)); k < len(athtTones) {
				return fmtTone(athtTones[k])
			}
		case 2:
			k := int(r.F(extra).Int(img))<<8 | int(r.F(code).Int(img))
			if k < len(allDCS) {
				return genTone{Mode: "DTCS", Code: allDCS[k], Inv: r.F(inv).Bool(img)}.String()
			}
		}
		return ""
	}
	c.TxTone = tone("txtmode", "txtone", "txcode", "txdcsextra", "txinv")
	c.RxTone = tone("rxtmode", "rxtone", "rxcode", "rxdcsextra", "rxinv")
	c.Scan = img[0x120+i/8]&(1<<(i%8)) == 0
	p := int(r.F("power").Int(img))
	if p > 2 {
		p = 0
	}
	c.Power = m.powerSem(p)
	return c
}

func (f athtFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	bit := byte(1 << (i % 8))
	if c == nil {
		img[0x100+i/8] |= bit
		img[0x120+i/8] |= bit
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
	old := f.Get(m, img, i)
	img[0x100+i/8] &^= bit
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, 32)) {
		r.Fill(img, 0)
		r.F("squelch").SetInt(img, 3)
	}
	r.F("freq").SetInt(img, c.Rx/100)
	r.F("tx_off").SetBool(img, c.TxOff)
	d := c.Tx - c.Rx
	switch {
	case c.TxOff:
		r.F("duplex").SetInt(img, 0)
		r.F("offset").SetInt(img, 0)
	case d == 0:
		// come CHIRP: simplex con lo scostamento memorizzato (canale nuovo: quello
		// predefinito di CHIRP, 600 kHz)
		switch {
		case old == nil:
			r.F("offset").SetInt(img, 6000)
		case old.TxOff:
			r.F("offset").SetInt(img, 0)
		}
		r.F("duplex").SetInt(img, 0)
	case d > 0:
		r.F("duplex").SetInt(img, 3)
		r.F("offset").SetInt(img, d/100)
	default:
		r.F("duplex").SetInt(img, 2)
		r.F("offset").SetInt(img, -d/100)
	}
	nf := r.F("name")
	for k := 0; k < 6; k++ {
		v := byte(' ')
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nf.Off+k] = v
	}
	w := int64(0)
	if c.Narrow {
		w = 1
	}
	r.F("channel_width").SetInt(img, w)
	set := func(t genTone, mode, tv, code, extra, inv string) error {
		switch t.Mode {
		case "Tone":
			k := idxInt(athtTones, t.Hz10)
			if k < 0 {
				return fmt.Errorf("tono %s non disponibile su questa radio", t.String())
			}
			r.F(mode).SetInt(img, 1)
			r.F(tv).SetInt(img, int64(k))
		case "DTCS":
			k := idxInt(allDCS, t.Code)
			if k < 0 {
				return fmt.Errorf("codice DCS %s non valido", t.String())
			}
			r.F(mode).SetInt(img, 2)
			r.F(code).SetInt(img, int64(k&0xFF))
			r.F(extra).SetInt(img, int64(k>>8))
		default:
			r.F(mode).SetInt(img, 0)
		}
		r.F(inv).SetBool(img, t.Mode == "DTCS" && t.Inv)
		return nil
	}
	if err := set(tt, "txtmode", "txtone", "txcode", "txdcsextra", "txinv"); err != nil {
		return err
	}
	if err := set(rt, "rxtmode", "rxtone", "rxcode", "rxdcsextra", "rxinv"); err != nil {
		return err
	}
	r.F("sqlmode").SetBool(img, rt.Mode != "")
	if c.Scan {
		img[0x120+i/8] &^= bit
	} else {
		img[0x120+i/8] |= bit
	}
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	return nil
}

// chirpInvalid: toni che CHIRP non accetta per questa radio
func (f athtFam) chirpInvalid(m *genModel, c *Chan) bool {
	for _, s := range []string{c.TxTone, c.RxTone} {
		if t, err := genToneParse(s); err != nil || t.Mode == "Tone" && idxInt(athtTones, t.Hz10) < 0 {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) athtIdent(ident string) error {
	if err := s.send([]byte("PROGRAM")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 3, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if !bytes.Equal(b, []byte("QX\x06")) {
		return rErr("risposta inattesa della radio")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 16, time.Second); !bytes.Contains(id, []byte(ident)) {
		return rErr("modello non supportato (% X)", id)
	}
	return nil
}

func (s *rawSess) athtFinish() error {
	if err := s.send([]byte("END")); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio non ha chiuso la sessione correttamente")
	}
	return nil
}

func (f athtFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.athtIdent(f.ident); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x8000; a += 0x10 {
		prog(a, 0x8000, "Lettura…")
		h := []byte{'R', byte(a >> 8), byte(a), 0x10}
		if err := s.send(h); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x16, time.Second)
		if len(b) != 0x16 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X incompleta", a)
		}
		if b[0x15] != 0x06 {
			return nil, "", rErr("la radio ha rifiutato il blocco 0x%04X", a)
		}
		if int(b[1])<<8|int(b[2]) != a {
			return nil, "", rErr("la radio ha inviato un blocco inatteso")
		}
		if tdm11Sum(b[1:0x14]) != b[0x14] {
			return nil, "", rErr("somma di controllo errata nel blocco 0x%04X", a)
		}
		data = append(data, b[4:0x14]...)
	}
	if err := s.athtFinish(); err != nil {
		return nil, "", err
	}
	prog(0x8000, 0x8000, "Lettura completata")
	return data, "", nil
}

func (f athtFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x8000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x8000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.athtIdent(f.ident); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0x100; a < 0x8000; a += 0x10 {
		prog(a, 0x8000, "Scrittura…")
		fr := append([]byte{'W', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)
		fr = append(fr, tdm11Sum(fr[1:]), 0x06)
		if err := s.send(fr); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	if err := s.athtFinish(); err != nil {
		return "", err
	}
	prog(0x8000, 0x8000, "Scrittura completata")
	return "", nil
}
