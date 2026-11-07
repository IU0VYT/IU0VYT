// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// btech.py, classi QYTColorHT e KT8R (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: QYT KT-8R (btech.py di CHIRP), portatile quadribanda con display a
// colori: 200 canali con nomi di 8 caratteri. 9600 baud; chiave di 8 byte →
// 06 + identificativo («MCB264»…); blocchi 06 'S' da 64 byte in lettura e
// 06 'X' da 16 byte in scrittura (fino a 0x3100). Le gamme di frequenza sono
// quelle scritte nella memoria della radio (0x3D80), come in CHIRP.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type kt8rFam struct{}

const kt8rFormat = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown0:4,
     scode:4;
  u8 unknown1:2,
     spmute:2,
     unknown2:2,
     optsig:2;
  u8 unknown3:3,
     scramble:1,
     unknown4:3,
     power:1;
  u8 unknown5:1,
     wide:1,
     unknown6:2,
     bcl:1,
     add:1,
     pttid:2;
} memory[200];

#seekto 0x1000;
struct {
  char name[8];
  u8 unknown1[8];
} names[200];
`

var kt8rFileids = []string{"MCB264", "MCB284", "MC5264", "MCB254"}

func init() {
	genRegister(&genModel{ID: "qyt-kt8r", Vendor: "QYT", Model: "KT-8R", RClass: "KT8R", Name: "QYT KT-8R",
		Slots: 200, First: 0, NameMax: 8, Charset: charsetAlnum + "`{|}!\"#$%&'()*+,-./:;<=>?@[]^_",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}},
		Bands:  []band{{136000000, 175000000}, {400000000, 481050001}, {200000000, 261000000}, {350000000, 391000000}},
		Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: kt8rFam{},
		Notes: "Portatile quadribanda (VHF, 220, 350 e UHF) con display a colori: 200 canali con nomi di 8 caratteri; le " +
			"gamme sono quelle impostate nella radio. Scrambler, blocco del canale occupato, PTT ID, segnalazione e " +
			"silenziamento dell'altoparlante si conservano. Driver di CHIRP (sperimentale)."})
}

func (kt8rFam) mem() *bwEl                 { return bwMust(kt8rFormat) }
func (f kt8rFam) layout(m *genModel) *bwEl { return f.mem() }
func (f kt8rFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("names").I(i)
}
func (f kt8rFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f kt8rFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 32 {
		n.SetRaw(img, raw[16:32])
	}
}
func (f kt8rFam) Match(m *genModel, img []byte) bool {
	if len(img) < 0x3F76 {
		return false
	}
	for _, id := range kt8rFileids {
		if string(img[0x3F70:0x3F76]) == id {
			return true
		}
	}
	return false
}

// Bands: le gamme scritte nella radio (set_options di btech.py)
func (f kt8rFam) Bands(m *genModel, img []byte) []band {
	if len(img) < 0x3DC0 {
		return m.Bands
	}
	rng := func(off int) (int64, int64) {
		d := func(b []byte) int64 { return (int64(b[0])*100 + int64(b[1])*10 + int64(b[2])) * 1000000 }
		return d(img[off : off+3]), d(img[off+3 : off+6])
	}
	vl, vh := rng(0x3D80)
	v2l, v2h := rng(0x3D90)
	ul, uh := rng(0x3DA0)
	u2l, u2h := rng(0x3DB0)
	if vl == 0 || vh <= vl || uh <= ul {
		return m.Bands
	}
	return []band{{vl, vh}, {ul, uh + 50001}, {v2l, v2h}, {u2l, u2h}}
}

func (f kt8rFam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	c.Tx = c.Rx
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = string(bytes.ReplaceAll(bytes.TrimRight(n.F("name").Raw(img), "\xFF"), []byte{0xFF}, []byte{' '}))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("wide").Bool(img)
	c.Scan = r.F("add").Bool(img)
	dec := func(v int64) string {
		if v > 0x258 {
			return fmtTone(int(v))
		}
		return uv5rToneDec(v, false)
	}
	c.TxTone = dec(r.F("txtone").Int(img))
	c.RxTone = dec(r.F("rxtone").Int(img))
	return c
}

func (f kt8rFam) Set(m *genModel, img []byte, i int, c *Chan) error {
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
	if r.Raw(img)[0] == 0xFF {
		r.Fill(img, 0)
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	nm := bytes.TrimRight([]byte(c.Name), " ")
	nf := n.F("name")
	for k := 0; k < 8; k++ {
		v := byte(0xFF)
		if k < len(nm) {
			v = nm[k]
		}
		img[nf.Off+k] = v
	}
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("wide").SetBool(img, c.Narrow)
	r.F("add").SetBool(img, c.Scan)
	for k := 0; k <= 6; k++ {
		r.F(fmt.Sprintf("unknown%d", k)).SetInt(img, 0)
	}
	// come CHIRP: il codice PTT ID 15 diventa 0
	if r.F("scode").Int(img) == 15 {
		r.F("scode").SetInt(img, 0)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) kt8rIdent() error {
	_ = s.p.Purge()
	if err := s.send([]byte{0x55, 0x20, 0x17, 0x07, 0x03, 0xFF, 0xDC, 0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 50, 1500*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(id) == 0 {
		return rErr("la radio non risponde")
	}
	if id[0] != 0x06 {
		return rErr("risposta inattesa della radio")
	}
	for _, f := range kt8rFileids {
		if bytes.Contains(id, []byte(f)) {
			return s.sleep(100 * time.Millisecond)
		}
	}
	return rErr("identificazione della radio non riuscita (% X)", id)
}

func (f kt8rFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x4000, "Collegamento con la radio…")
	if err := s.kt8rIdent(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x4000; a += 0x40 {
		prog(a, 0x4000, "Lettura…")
		if err := s.send([]byte{0x06, 'S', byte(a >> 8), byte(a), 0x40}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x45, 1500*time.Millisecond)
		if len(b) != 0x45 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X incompleta", a)
		}
		if b[0] != 0x06 && b[0] != 0x05 {
			return nil, "", rErr("conferma errata della radio nel blocco 0x%04X", a)
		}
		if !bytes.Equal(b[1:5], []byte{'X', byte(a >> 8), byte(a), 0x40}) {
			return nil, "", rErr("intestazione non valida nel blocco 0x%04X", a)
		}
		data = append(data, b[5:]...)
	}
	s.send([]byte{0x06})
	prog(0x4000, 0x4000, "Lettura completata")
	return data, "", nil
}

func (f kt8rFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x3100 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x3100, "Collegamento con la radio…")
	if err := s.kt8rIdent(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x3100; a += 0x10 {
		prog(a, 0x3100, "Scrittura…")
		if err := s.send(append([]byte{0x06, 'X', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 1, 1500*time.Millisecond)
		if len(b) == 1 && b[0] != 0x06 {
			b = readUpTo(ctx, s.p, 1, 1500*time.Millisecond)
		}
		if len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha confermato il blocco 0x%04X", a)
		}
	}
	prog(0x3100, 0x3100, "Scrittura completata")
	return "", nil
}
