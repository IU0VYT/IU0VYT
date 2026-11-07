// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rt23.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RT23 (retevis_rt23.py di CHIRP), bibanda 2 m / 70 cm con ripetitore
// a banda incrociata: 128 canali con nomi di 7 caratteri. 9600 baud; «PROIUAM»
// (un byte ogni 5 ms, fino a 5 volte) → 06, 02 → «P31183…», 06 → 06; blocchi
// 'R'/'W' da 16 byte con somma di controllo; canali usati e in scansione in due
// mappe di bit.

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sort"
	"time"
)

type rt23Fam struct{}

const rt23Format = `
struct memory {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rxtone[2];
  lbcd txtone[2];
  u8 unknown1;
  u8 pttid:2,
     unknown2:1,
     signaling:1,
     unknown3:1,
     bcl:1,
     unknown4:2;
  u8 unknown5:3,
     highpower:1,
     isnarrow:1,
     scan:1,
     unknown6:2;
  u8 unknown7;
};

#seekto 0x0010;
struct memory channels[128];

#seekto 0x08D0;
struct {
  char name[7];
  u8 unknown2[1];
} names[128];

#seekto 0x0D20;
u8 usedflags[16];
u8 scanflags[16];
`

var rt23DCS = func() []int {
	out := append([]int{}, dtcsCodes...)
	for _, c := range []int{17, 50, 55, 135, 217, 254, 305, 645, 765} {
		if !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	sort.Ints(out)
	return out
}()

func init() {
	genRegister(&genModel{ID: "rt-rt23", Vendor: "Retevis", Model: "RT23", RClass: "RT23Radio", Name: "Retevis RT23",
		Slots: 128, First: 1, NameMax: 7, Charset: charsetUpperNum + ":;<=>?@ !\"#$%&'()*+,-./",
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 2,5 W", 2.5, 0}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true, DCS: rt23DCS,
		Baud: 9600, Fam: rt23Fam{},
		Notes: "128 canali con nomi di 7 caratteri (maiuscole, cifre e simboli). Blocco del canale occupato, PTT ID e " +
			"segnalazione DTMF si conservano. Driver di CHIRP."})
}

func (rt23Fam) mem() *bwEl                 { return bwMust(rt23Format) }
func (f rt23Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt23Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("channels").I(i), b.F("names").I(i)
}
func (f rt23Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.F("name").Raw(img)...)
}
func (f rt23Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 23 {
		n.F("name").SetRaw(img, raw[16:23])
	}
}
func (f rt23Fam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x1000 && bytes.HasPrefix(img, []byte("P31183"))
}

func (f rt23Fam) flag(img []byte, base, i int) bool { return img[base+i/8]&(1<<(i%8)) != 0 }
func (f rt23Fam) setFlag(img []byte, base, i int, on bool) {
	if on {
		img[base+i/8] |= 1 << (i % 8)
	} else {
		img[base+i/8] &^= 1 << (i % 8)
	}
}

func (f rt23Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if !f.flag(img, 0xD20, i) {
		return nil
	}
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if r.Raw(img)[4] == 0xFF {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = pyName(n.F("name").Raw(img), true)
	c.Narrow = r.F("isnarrow").Bool(img)
	c.TxTone = x3ToneDec(r.F("txtone").Int(img))
	c.RxTone = x3ToneDec(r.F("rxtone").Int(img))
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Scan = f.flag(img, 0xD30, i)
	return c
}

func (f rt23Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		n.F("name").Fill(img, 0xFF)
		f.setFlag(img, 0xD20, i, false)
		f.setFlag(img, 0xD30, i, false)
		return nil
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	f.setFlag(img, 0xD20, i, true)
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, 16)) {
		r.Fill(img, 0)
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	nm := n.F("name")
	for k := 0; k < 7; k++ {
		v := byte(0xFF)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	r.F("scan").SetBool(img, c.Scan)
	f.setFlag(img, 0xD30, i, c.Scan)
	r.F("isnarrow").SetBool(img, c.Narrow)
	if err := x3ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rt23Enter() error {
	var ack []byte
	ok := false
	for k := 0; k < 5 && !ok; k++ {
		for _, b := range []byte("PROIUAM") {
			if err := s.sleep(5 * time.Millisecond); err != nil {
				return err
			}
			if err := s.send([]byte{b}); err != nil {
				return err
			}
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
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, time.Second); !bytes.HasPrefix(id, []byte("P31183")) {
		return rErr("identificativo sconosciuto (% X)", id)
	}
	return s.ack2()
}

func (f rt23Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	if err := s.rt23Enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x1000; a += 0x10 {
		prog(a, 0x1000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x10}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 4+0x10+1, time.Second)
		if len(b) != 0x15 || !bytes.Equal(b[:4], []byte{'W', byte(a >> 8), byte(a), 0x10}) || b[0x14] != tdm11Sum(b[4:0x14]) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		blk := b[4:0x14]
		if a == 0 && bytes.HasPrefix(blk, bytes.Repeat([]byte{0xFF}, 6)) {
			blk = append([]byte("P31183"), bytes.Repeat([]byte{0xFF}, 10)...)
		}
		data = append(data, blk...)
	}
	s.send([]byte{'E'})
	prog(0x1000, 0x1000, "Lettura completata")
	return data, "", nil
}

func (f rt23Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x1000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x1000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0xEC0, "Collegamento con la radio…")
	if err := s.rt23Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0xEC0; a += 0x10 {
		prog(a, 0xEC0, "Scrittura…")
		d := img[a : a+0x10]
		if err := s.send(append(append([]byte{'W', byte(a >> 8), byte(a), 0x10}, d...), tdm11Sum(d))); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(0xEC0, 0xEC0, "Scrittura completata")
	return "", nil
}
