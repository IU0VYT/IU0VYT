// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rb28.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RB28 (FRS), RB628 (PMR446), RB48 e RB648 (retevis_rb28.py di
// CHIRP). 9600 baud; chiave «PHOGR…» → 06, 02 → identificativo «P32073…», 06 → 06;
// blocchi 'R' da 32 byte (il primo senza conferma) e 'W' da 16 byte; 'E' per
// uscire. Toni CTCSS e codici DCS come numero d'ordine nelle tabelle PMR (senza
// polarità del DCS); gli ultimi 3 byte di ogni canale contengono le impostazioni
// della radio e si conservano.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rb28Opt struct {
	magic string
	lock  string // "frs" (RB28), "pmr" (RB628), "" (RB48, RB648: tutte le frequenze)
	size  int
}

type rb28Fam struct{}

func rb28o(m *genModel) *rb28Opt { return m.Opt.(*rb28Opt) }

const rb28FormatT = `
#seekto 0x0000;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 rxtone;
  u8 unused_9;
  u8 txtone;
  u8 unused_b;
  u8 compander:1,
     unused_1:1,
     highpower:1,
     wide:1,
     unused_2:4;
  u8 reserved[3];
} memory[%d];
`

var (
	rb28Tones = func() []int {
		drop := map[int]bool{693: true, 1598: true, 1655: true, 1713: true, 1773: true, 1835: true, 1899: true,
			1966: true, 1995: true, 2065: true, 2291: true, 2541: true}
		var out []int
		for _, t := range chirpTones {
			if !drop[t] {
				out = append(out, t)
			}
		}
		return out
	}()
	rb28DCS = func() []int {
		drop := map[int]bool{36: true, 53: true, 122: true, 145: true, 212: true, 225: true, 246: true, 252: true,
			255: true, 266: true, 274: true, 325: true, 332: true, 356: true, 446: true, 452: true, 454: true, 455: true,
			462: true, 523: true, 526: true}
		var out []int
		for _, d := range dtcsCodes {
			if !drop[d] {
				out = append(out, d)
			}
		}
		return out
	}()
)

func init() {
	note := func(s string) string {
		return s + " Toni CTCSS e codici DCS della tabella PMR della radio (senza polarità invertita). Compander e blocco " +
			"del canale occupato si conservano. Driver di CHIRP."
	}
	mk := func(id, model, rclass string, slots int, o *rb28Opt, pw []genPower, bands []band, noTxOff bool, n string) *genModel {
		return &genModel{ID: id, Vendor: "Retevis", Model: model, RClass: rclass, Name: "Retevis " + model, Slots: slots,
			First: 1, Powers: pw, Bands: bands, Narrow: true, DCS: rb28DCS, Baud: 9600, NoTxOff: noTxOff, Fam: rb28Fam{},
			Opt: o, Notes: note(n)}
	}
	genRegister(
		mk("rt-rb28", "RB28", "RB28Radio", 22, &rb28Opt{magic: "PHOGR\x0B\x28", lock: "frs", size: 0x160},
			[]genPower{{"Alta 2,5 W", 2.5, 0}, {"Bassa 0,5 W", 0.5, 2}}, []band{{400000000, 480000000}}, false,
			"Radio FRS per gli Stati Uniti (in Italia la banda FRS non è ammessa): 22 canali FRS fissi a banda stretta, "+
				"canali 8-14 a bassa potenza."),
		mk("rt-rb628", "RB628", "RB628Radio", 16, &rb28Opt{magic: "PHOGR\x28\x0B", lock: "pmr", size: 0x160},
			[]genPower{{"0,5 W (alta)", 0.5, 0}, {"0,5 W", 0.5, 2}}, []band{{400000000, 480000000}}, false,
			"Radio PMR446 (uso libero in Italia): 16 canali PMR fissi, banda stretta, 0,5 W."),
		mk("rt-rb48", "RB48", "RB48Radio", 22, &rb28Opt{magic: "PHOGR[\x17", size: 0x800},
			[]genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}, []band{{400000000, 520000000}}, true,
			"Venduta come radio FRS ma programmabile su tutta la gamma 400-520 MHz (in Italia non è ammessa in "+
				"trasmissione fuori dalle bande consentite)."),
		mk("rt-rb648", "RB648", "RB648Radio", 16, &rb28Opt{magic: "PHOGRQp", size: 0x800},
			[]genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}, []band{{400000000, 520000000}}, true,
			"Venduta come radio PMR ma programmabile su tutta la gamma 400-520 MHz (in Italia è ammessa in "+
				"trasmissione solo sulle frequenze PMR446 con 0,5 W)."),
	)
}

func (rb28Fam) mem(m *genModel) *bwEl          { return bwMust(fmt.Sprintf(rb28FormatT, m.Slots)) }
func (f rb28Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f rb28Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f rb28Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)[:13]...)
}
func (f rb28Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(img[f.rec(m, i).Off:], raw[:13])
}
func (f rb28Fam) Match(m *genModel, img []byte) bool { return false }

func (f rb28Fam) toneDec(v int64) string {
	switch {
	case v == 0 || v == 0xFF:
		return ""
	case v < 0x27:
		if int(v-1) < len(rb28Tones) {
			return fmtTone(rb28Tones[v-1])
		}
		return ""
	}
	if int(v-0x27) < len(rb28DCS) {
		return genTone{Mode: "DTCS", Code: rb28DCS[v-0x27]}.String()
	}
	return ""
}

func (f rb28Fam) toneEnc(s string) (int64, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, err
	}
	switch t.Mode {
	case "Tone":
		for k, v := range rb28Tones {
			if v == t.Hz10 {
				return int64(k + 1), nil
			}
		}
		return 0, fmt.Errorf("tono %s non disponibile su questa radio", s)
	case "DTCS":
		if t.Inv {
			return 0, fmt.Errorf("questa radio non ha il DCS a polarità invertita")
		}
		for k, v := range rb28DCS {
			if v == t.Code {
				return int64(k + 0x27), nil
			}
		}
		return 0, fmt.Errorf("codice DCS %s non disponibile su questa radio", s)
	}
	return 0, nil
}

// fixed: frequenza fissa della posizione (0 se libera)
func (f rb28Fam) fixed(m *genModel, i int) int64 {
	switch rb28o(m).lock {
	case "frs":
		if i < len(btGmrsAll) {
			return btGmrsAll[i]
		}
	case "pmr":
		if i < len(t18PMR) {
			return t18PMR[i]
		}
	}
	return 0
}

func (f rb28Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	rxr := r.F("rxfreq").Raw(img)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(rxr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = !r.F("wide").Bool(img)
	c.TxTone = f.toneDec(r.F("txtone").Int(img))
	c.RxTone = f.toneDec(r.F("rxtone").Int(img))
	if r.F("highpower").Bool(img) {
		c.Power = m.Powers[0].Sem
	} else {
		c.Power = m.Powers[1].Sem
	}
	// canali fissi come li mostra CHIRP
	if fx := f.fixed(m, i); fx != 0 {
		c.Rx, c.Narrow = fx, true
		switch rb28o(m).lock {
		case "frs":
			if !c.TxOff {
				c.Tx = fx
			}
			if n := i + 1; n >= 8 && n <= 14 {
				c.Power = m.Powers[len(m.Powers)-1].Sem
			}
		case "pmr":
			c.Tx, c.TxOff = fx, false
			c.Power = m.Powers[1].Sem
		}
	}
	return c
}

func (f rb28Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	raw := r.Raw(img)
	if c == nil {
		copy(img[r.Off:], bytes.Repeat([]byte{0xFF}, 13))
		return nil
	}
	tt, err := f.toneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := f.toneEnc(c.RxTone)
	if err != nil {
		return err
	}
	comp := int64(0)
	if f.Get(m, img, i) != nil {
		comp = r.F("compander").Int(img)
	}
	_ = raw
	copy(img[r.Off:], make([]byte, 13))
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("txtone").SetInt(img, tt)
	r.F("rxtone").SetInt(img, rt)
	// RB628: le due potenze sono entrambe 0,5 W e CHIRP le confronta in watt:
	// scrive sempre «alta»
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 0 || m.Powers[0].Watts == m.Powers[1].Watts)
	r.F("compander").SetInt(img, comp)
	return nil
}

// chirpInvalid: toni che CHIRP non accetta per questa radio
func (f rb28Fam) chirpInvalid(m *genModel, c *Chan) bool {
	_, e1 := f.toneEnc(c.TxTone)
	_, e2 := f.toneEnc(c.RxTone)
	return e1 != nil || e2 != nil
}

// lockedSlotImg / Locked: canali FRS e PMR fissi (immutable di CHIRP)
func (f rb28Fam) lockedSlotImg(m *genModel, img []byte, i int) bool { return f.fixed(m, i) != 0 }

func (f rb28Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if want != nil {
		for _, t := range []string{want.TxTone, want.RxTone} {
			if _, err := f.toneEnc(t); err != nil {
				return err
			}
		}
	}
	fx := f.fixed(m, i)
	if fx == 0 {
		return nil
	}
	n := i + 1
	kind := map[string]string{"frs": "FRS", "pmr": "PMR446"}[rb28o(m).lock]
	switch {
	case want == nil:
		if old != nil {
			return fmt.Errorf("il canale %d è un canale %s fisso: non si può cancellare", n, kind)
		}
		return nil
	case want.Rx != fx:
		return fmt.Errorf("il canale %d è il canale %s %s MHz: frequenza non modificabile", n, kind, fmtMHz(fx))
	case !want.Narrow:
		return fmt.Errorf("i canali %s sono sempre a banda stretta", kind)
	case rb28o(m).lock == "pmr" && want.TxOff:
		return fmt.Errorf("sui canali PMR446 la trasmissione non si può disabilitare")
	case !want.TxOff && want.Tx != fx:
		return fmt.Errorf("sui canali %s TX e RX coincidono", kind)
	case old != nil && want.TxOff != old.TxOff:
		return fmt.Errorf("sul canale %d la trasmissione non si può attivare o disattivare (come in CHIRP)", n)
	case rb28o(m).lock == "frs" && n >= 8 && n <= 14 && want.Power != m.Powers[len(m.Powers)-1].Sem:
		return fmt.Errorf("i canali FRS 8-14 sono sempre a bassa potenza")
	case rb28o(m).lock == "pmr" && want.Power != m.Powers[1].Sem:
		return fmt.Errorf("i canali PMR446 sono sempre a 0,5 W")
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rb28Enter(o *rb28Opt) error {
	if err := s.send([]byte(o.magic)); err != nil {
		return err
	}
	var ack []byte
	for k := 0; k < 4; k++ {
		ack = readUpTo(s.ctx, s.p, 1, time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(ack) == 1 && ack[0] == 0x06 {
			break
		}
	}
	if len(ack) == 0 {
		return rErr("la radio non risponde")
	}
	if ack[0] != 0x06 {
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, time.Second)
	if !bytes.Contains(id, []byte("P32073\x02\xFF")) && !bytes.Contains(id, []byte("P32073\x00\xFF")) {
		return rErr("identificazione della radio non riuscita (% X)", id)
	}
	return s.ack2()
}

func (f rb28Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rb28o(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.rb28Enter(o); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < o.size; a += 0x20 {
		prog(a, o.size, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x20}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 4+0x20, time.Second)
		if len(b) != 4+0x20 || !bytes.Equal(b[:4], []byte{'W', byte(a >> 8), byte(a), 0x20}) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if a != 0 {
			if err := s.ack2(); err != nil {
				return nil, "", err
			}
		}
		data = append(data, b[4:]...)
	}
	s.send([]byte{'E'})
	prog(o.size, o.size, "Lettura completata")
	return data, "", nil
}

func (f rb28Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rb28o(m)
	if len(img) < 0x160 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x160, "Collegamento con la radio…")
	if err := s.rb28Enter(o); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x160; a += 0x10 {
		prog(a, 0x160, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	s.send([]byte{'E'})
	prog(0x160, 0x160, "Scrittura completata")
	return "", nil
}
