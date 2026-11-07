// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// th_uv8000.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: TYT TH-UV8000D / TH-UV8000E (th_uv8000.py di CHIRP). 4800 baud,
// «PROGRAMa», «PROGRAMb», identificativo «MTC88CUMHS3E7BN-», poi blocchi 'R'/'W'
// da 16 byte (4,75 KiB). 128 canali da 32 byte con nomi di 7 caratteri (codice
// del carattere meno 32), mappe dei canali presenti e da saltare in scansione.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type thuv8000Fam struct{}

const thuv8000Format = `
struct chns {
  ul32 rxfreq;
  ul32 txfreq;
  u8 rxtone[2];
  u8 txtone[2];
  u8  wide:1,
      vox_on:1,
      chunk01:1,
      bcl:1,
      epilogue:1,
      power:1,
      chunk02:1,
      chunk03:1;
  u8  ani:1,
      chunk08:1,
      ptt:2,
      chpad04:4;
  u8  chunk05;
  u16 id_code;
  u8  chunk06;
  u8  name[7];
  ul32 chpad06;
  ul16 chpad07;
  u8  chpad08;
};
#seekto 0x0010;
struct chns chan_mem[128];
#seekto 0x1190;
struct {
  u8  map[16];
} chnmap;
struct {
  u8  map[16];
} skpchns;
`

const thuv8000Size = 0x1300

func init() {
	genRegister(&genModel{ID: "tyt-thuv8000", Vendor: "TYT", Model: "TH-UV8000", RClass: "THUV8000Radio", Name: "TYT TH-UV8000D / TH-UV8000E",
		Slots: 128, First: 1, NameMax: 7, Charset: charsetAlnum + "`!\"#$%&'()*+,-./:;<=>?@[]^_",
		Powers: []genPower{{"Bassa 5 W", 5, 2}, {"Alta 10 W", 10, 0}},
		Bands:  []band{{87500000, 107900000}, {136000000, 174000000}, {220000000, 260000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, DCS: withCodes(dtcsCodes, 645), Baud: 4800, Fam: thuv8000Fam{},
		Notes: "Spegni la radio, collega il cavo, riaccendila con il volume a metà (4800 baud). 128 canali con nomi di 7 caratteri; " +
			"blocco del canale occupato, VOX, ANI, PTT ID e coda di ogni canale si conservano. Driver di CHIRP."})
}

func (thuv8000Fam) mem() *bwEl                 { return bwMust(thuv8000Format) }
func (f thuv8000Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f thuv8000Fam) rec(i int) *bwEl          { return f.mem().F("chan_mem").I(i) }
func (f thuv8000Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.bit(img, "skpchns", i))
}
func (f thuv8000Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(i).SetRaw(img, raw[:32])
	if len(raw) >= 33 {
		f.setBit(img, "skpchns", i, raw[32] != 0)
	}
}
func (f thuv8000Fam) Match(m *genModel, img []byte) bool { return false }

// crossband: CHIRP non ammette «split»; la memoria ha la frequenza di trasmissione
func (thuv8000Fam) crossband() bool { return true }

// bit del canale nelle mappe (_do_map: bit (n-1)%8 del byte (n-1)/8)
func (f thuv8000Fam) bit(img []byte, mp string, i int) byte {
	off := f.mem().F(mp).F("map").Off + i/8
	return (img[off] >> (i % 8)) & 1
}

func (f thuv8000Fam) setBit(img []byte, mp string, i int, on bool) {
	off := f.mem().F(mp).F("map").Off + i/8
	if on {
		img[off] |= 1 << (i % 8)
	} else {
		img[off] &^= 1 << (i % 8)
	}
}

// toni: due byte in BCD (basso, alto); alto con 0x80 = DCS e 0x40 = invertito
func thuv8000ToneDec(lo, hi byte) string {
	if lo == 0xFF {
		return ""
	}
	if hi < 0x26 {
		return fmtTone(int(bcdByte(hi))*100 + int(bcdByte(lo)))
	}
	return genTone{Mode: "DTCS", Code: int(hi&0x3F)*100 + int(bcdByte(lo)), Inv: hi&0x40 != 0}.String()
}

func thuv8000ToneEnc(t genTone) (byte, byte) {
	switch t.Mode {
	case "Tone":
		return decToBCD(int64(t.Hz10) % 100), decToBCD(int64(t.Hz10) / 100)
	case "DTCS":
		hi := decToBCD(int64(t.Code)/100) | 0x80
		if t.Inv {
			hi |= 0x40
		}
		return decToBCD(int64(t.Code) % 100), hi
	}
	return 0xFF, 0xFF
}

func (f thuv8000Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if tx := r.F("txfreq").Int(img); tx == 0xFFFFFFFF {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = tx * 10
	}
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("wide").Int(img) == 0
	rt, tt := r.F("rxtone").Raw(img), r.F("txtone").Raw(img)
	c.RxTone = thuv8000ToneDec(rt[0], rt[1])
	c.TxTone = thuv8000ToneDec(tt[0], tt[1])
	var nm []byte
	for _, b := range r.F("name").Raw(img) {
		nm = append(nm, b+32)
	}
	c.Name = latin1Str(strings.TrimRight(string(nm), " \t\n\r\x0b\x0c"))
	c.Scan = f.bit(img, "skpchns", i) == 0
	return c
}

func (f thuv8000Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.F("rxfreq").SetInt(img, 0xFFFFFFFF)
		f.setBit(img, "chnmap", i, true)
		f.setBit(img, "skpchns", i, true)
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
	r.F("power").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("wide").SetBool(img, !c.Narrow)
	lo, hi := thuv8000ToneEnc(rt)
	r.F("rxtone").SetRaw(img, []byte{lo, hi})
	lo, hi = thuv8000ToneEnc(tt)
	r.F("txtone").SetRaw(img, []byte{lo, hi})
	nm := make([]byte, 7)
	name := c.Name + strings.Repeat(" ", 7)
	for k := 0; k < 7; k++ {
		v := int(name[k]) - 32
		if v < 0 {
			v = 0
		}
		nm[k] = byte(v)
	}
	r.F("name").SetRaw(img, nm)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0xFFFFFFFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	f.setBit(img, "chnmap", i, false)
	f.setBit(img, "skpchns", i, !c.Scan)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) thuv8000Ident() error {
	_ = s.p.Purge()
	for _, cmd := range []string{"PROGRAMa", "PROGRAMb"} {
		if err := s.send([]byte(cmd)); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, 2*time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			_ = s.send([]byte("E"))
			return rErr("la radio non risponde")
		}
		if b[0] != 0x06 {
			_ = s.send([]byte("E"))
			return rErr("risposta inattesa della radio")
		}
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 8, 2*time.Second); len(b) != 8 {
		_ = s.send([]byte("E"))
		return rErr("la radio non ha inviato l'identificativo")
	}
	if err := s.send([]byte("MTC88CUMHS3E7BN-")); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 {
		_ = s.send([]byte("E"))
		return rErr("la radio non ha confermato")
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 {
		_ = s.send([]byte("E"))
		return rErr("la radio non ha confermato")
	}
	return nil
}

func (f thuv8000Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, thuv8000Size, "Collegamento con la radio…")
	if err := s.thuv8000Ident(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < thuv8000Size; a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, thuv8000Size, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 16}); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 4+16, 2*time.Second)
		if len(resp) != 20 {
			_ = s.send([]byte("E"))
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, resp[4:]...)
	}
	_ = s.send([]byte("E"))
	prog(thuv8000Size, thuv8000Size, "Lettura completata")
	return data, "", nil
}

func (f thuv8000Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < thuv8000Size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), thuv8000Size)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, thuv8000Size, "Collegamento con la radio…")
	if err := s.thuv8000Ident(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < thuv8000Size; a += 16 {
		prog(a, thuv8000Size, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 16}, img[a:a+16]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, 2*time.Second); len(b) != 1 || b[0] != 0x06 {
			_ = s.send([]byte("E"))
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	_ = s.send([]byte("E"))
	prog(thuv8000Size, thuv8000Size, "Scrittura completata")
	return "", nil
}
