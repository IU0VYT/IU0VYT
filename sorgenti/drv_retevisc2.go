// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_c2.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis C2 (retevis_c2.py di CHIRP), portatile GMRS. 115200 baud con RTS
// spento; trame STX … ETX con «byte stuffing» DLE e somma di controllo; la
// memoria è fatta di tabelle (identificativo, impostazioni, 256 canali da 56
// byte, blocchi extra, tabella finale, configurazione e taratura) lette e
// scritte una per volta, con una verifica dopo ogni scrittura.

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"time"
)

type c2Fam struct{}

const c2Format = `
#seekto 0x0360;
struct {
    u8 cfg_unk0:1,
       highpower:1,
       cfg_unk1:1,
       busylock:1,
       talkaround:1,
       cfg_unk2:3;
    ul16 bandwidth;
    u8 pad0;
    ul32 rxfreq;
    ul32 rxtone;
    ul32 txfreq;
    ul32 txtone;
    u8 unknown1[20];
    char name[12];
    ul16 unk15:1,
         compand:1,
         unk13_12:2,
         unk11:1,
         unk10:1,
         dtmf_group:4,
         ani:1,
         scanadd:1,
         unk3:1,
         pttid:3;
    ul16 scramble;
} memory[256];
`

const (
	c2MemSize     = 17396
	c2SetOff      = 0x001C
	c2SetSize     = 836
	c2ChanOff     = 0x0360
	c2ChanSize    = 56
	c2ExtraOff    = 0x3B60
	c2FinalOff    = 0x3C60
	c2FinalSize   = 480
	c2FinalWrite  = 448
	c2SysSize     = 1100
	c2RfcalCount  = 10
	c2ExtraCount  = 8
	c2ExtraSize   = 32
	c2IdentSize   = 28
	c2RfcalSize   = 36
	c2SysconfOff  = 0x3E40
	c2RfcalOff    = 0x428C
	c2CmdIdent    = 0x01
	c2CmdRead     = 0x02
	c2CmdWrite    = 0x03
	c2CmdVerify   = 0x04
	c2CmdFinalize = 0x05
	c2TblSys      = 0x01
	c2TblRfcal    = 0x02
	c2TblSet      = 0x03
	c2TblChan     = 0x04
	c2TblExtra    = 0x05
	c2TblFinal    = 0x07
)

func init() {
	genRegister(&genModel{ID: "rt-c2", Vendor: "Retevis", Model: "C2", RClass: "RetevisC2", Name: "Retevis C2 (GMRS)",
		Slots: 256, First: 1, NameMax: 12, Charset: charsetAlnum + " !@#$%&*()-+=",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}}, Bands: []band{{100000000, 520000000}},
		Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 115200, Fam: c2Fam{},
		Notes: "Radio GMRS per gli Stati Uniti: in Italia la banda GMRS non è ammessa. Copia a 115200 baud con il cavo " +
			"USB della radio. Come in CHIRP, i canali che trasmettono sulle frequenze GMRS 467,5625-467,7125 MHz sono sempre " +
			"a banda stretta e bassa potenza. Compander, talk around, blocco del canale occupato, ANI, PTT ID, gruppo DTMF e " +
			"scrambler si conservano. Driver di CHIRP."})
}

func (c2Fam) mem() *bwEl                 { return bwMust(c2Format) }
func (f c2Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f c2Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f c2Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f c2Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f c2Fam) Match(m *genModel, img []byte) bool                { return len(img) == c2MemSize }

// ---------------------------------------------------------------- toni

func c2Rev(v uint32, n int) uint32 {
	var r uint32
	for i := 0; i < n; i++ {
		if v&(1<<i) != 0 {
			r |= 1 << (n - 1 - i)
		}
	}
	return r
}

func c2Golay(d12 uint32) uint32 {
	msg := d12 << 11
	for i := 11; i >= 0; i-- {
		if msg&(1<<(i+11)) != 0 {
			msg ^= 0xAE3 << i
		}
	}
	return msg & 0x7FF
}

func c2DcsEnc(code int, inv bool) uint32 {
	dec, _ := strconv.ParseInt(strconv.Itoa(code), 8, 32)
	c9, low3 := uint32(dec), uint32(0x01)
	if inv {
		c9, low3 = ^uint32(dec)&0x1FF, 0x06
	}
	d12 := c2Rev(c9, 9)<<3 | low3
	cw := d12<<11 | c2Golay(d12)
	t := c2Rev(cw, 23)<<2 | 0x02
	if inv {
		t |= 0x02 << 24
	}
	return t
}

func c2ToneDec(v uint32) string {
	if v == 0 {
		return ""
	}
	if v&0x02 == 0 {
		return fmtTone(int(math.Round(float64(int64(v)-1) / 4)))
	}
	raw := (v >> 2) & 0x1FF
	inv := (v>>25)&1 == 1
	if inv {
		raw = ^raw & 0x1FF
	}
	code, _ := strconv.Atoi(strconv.FormatInt(int64(raw), 8))
	return genTone{Mode: "DTCS", Code: code, Inv: inv}.String()
}

func c2ToneEnc(t genTone) uint32 {
	switch t.Mode {
	case "Tone":
		return uint32(t.Hz10)*4 + 1
	case "DTCS":
		return c2DcsEnc(t.Code, t.Inv)
	}
	return 0
}

// ---------------------------------------------------------------- canali

func c2HHOnly(tx int64) bool { return inI64(btGmrsF2, tx) }

func (f c2Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	rx := r.F("rxfreq").Int(img)
	if rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if tx := r.F("txfreq").Int(img); tx == 0 || tx == 0xFFFFFFFF {
		c.TxOff = true
	} else {
		c.Tx = tx
	}
	c.Narrow = r.F("bandwidth").Int(img) != 25000
	if r.F("highpower").Bool(img) {
		c.Power = m.powerSem(0)
	} else {
		c.Power = m.powerSem(1)
	}
	c.TxTone = c2ToneDec(uint32(r.F("txtone").Int(img)))
	c.RxTone = c2ToneDec(uint32(r.F("rxtone").Int(img)))
	c.Name = string(bytes.TrimRight(bytes.TrimRight(r.F("name").Raw(img), "\x00"), " \t\n\r\x0b\x0c"))
	c.Scan = r.F("scanadd").Bool(img)
	if !c.TxOff && c2HHOnly(c.Tx) {
		c.Narrow, c.Power = true, m.powerSem(1)
	}
	return c
}

func (f c2Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0)
		r.F("bandwidth").SetInt(img, 25000)
		r.F("highpower").SetInt(img, 1)
		r.F("cfg_unk2").SetInt(img, 5)
		r.F("scanadd").SetInt(img, 1)
		r.F("pttid").SetInt(img, 1)
		r.F("unk11").SetInt(img, 1)
		r.F("scramble").SetInt(img, 0xFFFF)
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
	for _, ch := range []byte(c.Name) {
		if ch >= 0x80 {
			return fmt.Errorf("carattere non ammesso nel nome")
		}
	}
	existed := f.Get(m, img, i) != nil
	r.F("rxfreq").SetInt(img, c.Rx)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0)
	} else {
		r.F("txfreq").SetInt(img, c.Tx)
	}
	hh := !c.TxOff && c2HHOnly(c.Tx)
	if hh || c.Narrow {
		r.F("bandwidth").SetInt(img, 12500)
	} else {
		r.F("bandwidth").SetInt(img, 25000)
	}
	r.F("highpower").SetBool(img, !hh && m.powerIndex(c.Power) == 0)
	r.F("txtone").SetInt(img, int64(c2ToneEnc(tt)))
	r.F("rxtone").SetInt(img, int64(c2ToneEnc(rt)))
	nm := r.F("name")
	for k := 0; k < 12; k++ {
		v := byte(0)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	r.F("scanadd").SetBool(img, c.Scan)
	r.F("pad0").SetInt(img, 0)
	if !existed {
		// canale nuovo: opzioni come le imposta CHIRP
		for _, k := range []string{"busylock", "talkaround", "compand", "dtmf_group", "ani"} {
			r.F(k).SetInt(img, 0)
		}
		r.F("pttid").SetInt(img, 1)
		r.F("scramble").SetInt(img, 0xFFFF)
	}
	return nil
}

// Locked: come CHIRP (_apply_gmrs_limits), sulle frequenze GMRS 467 MHz «solo
// portatile» la radio usa sempre banda stretta e bassa potenza
func (f c2Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if want == nil || want.TxOff || !c2HHOnly(want.Tx) {
		return nil
	}
	if !want.Narrow {
		return fmt.Errorf("canale %d: sulle frequenze GMRS 467,5625-467,7125 MHz solo banda stretta", i+1)
	}
	if want.Power != m.powerSem(1) {
		return fmt.Errorf("canale %d: sulle frequenze GMRS 467,5625-467,7125 MHz solo bassa potenza", i+1)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

var c2Esc = map[byte]byte{0x00: 0x13, 0x02: 0x01, 0x03: 0x11, 0x10: 0x12, 0xFF: 0x14}

func c2Frame(payload []byte) []byte {
	var sum byte
	for _, b := range payload {
		sum += b
	}
	out := []byte{0x02}
	for _, b := range append(append([]byte{}, payload...), sum) {
		if e, ok := c2Esc[b]; ok {
			out = append(out, 0x10, e)
		} else {
			out = append(out, b)
		}
	}
	return append(out, 0x03)
}

// c2Unframe: contenuto di una trama (senza STX/ETX), controllo della somma
func c2Unframe(fr []byte) ([]byte, error) {
	var in []byte
	for k := 0; k < len(fr); k++ {
		if fr[k] == 0x10 && k+1 < len(fr) {
			k++
			v := fr[k]
			for o, e := range c2Esc {
				if e == fr[k] {
					v = o
				}
			}
			in = append(in, v)
		} else {
			in = append(in, fr[k])
		}
	}
	if len(in) < 2 {
		return nil, rErr("trama troppo corta")
	}
	p := in[:len(in)-1]
	var sum byte
	for _, b := range p {
		sum += b
	}
	if sum != in[len(in)-1] {
		return nil, rErr("somma di controllo errata nella risposta della radio")
	}
	return p, nil
}

// c2ReadFrame: una trama STX … ETX (tenendo conto dei DLE), entro 5 secondi
func c2ReadFrame(ctx context.Context, p Port, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	var data []byte
	esc := false
	for time.Now().Before(deadline) {
		b, err := readByte(ctx, p, time.Until(deadline))
		if err != nil {
			if ctx.Err() != nil {
				return nil, errCancelled
			}
			break
		}
		if len(data) == 0 {
			if b == 0x02 {
				data = append(data, b)
			}
			continue
		}
		data = append(data, b)
		switch {
		case esc:
			esc = false
		case b == 0x10:
			esc = true
		case b == 0x03:
			return data[1 : len(data)-1], nil
		}
	}
	return nil, nil
}

type c2Sess struct {
	*rawSess
	saw bool
}

func (s *c2Sess) xfer(cmd, tbl, idx byte, data []byte, retry bool) ([]byte, error) {
	hdr := []byte{0x01, cmd, tbl, idx}
	fr := c2Frame(append(append([]byte{}, hdr...), data...))
	tries := 1
	if retry {
		tries = 3
	}
	for k := 0; k < tries; k++ {
		if err := s.send(fr); err != nil {
			return nil, err
		}
		in, err := c2ReadFrame(s.ctx, s.p, 5*time.Second)
		if err != nil {
			return nil, err
		}
		if in != nil {
			s.saw = true
			resp, err := c2Unframe(in)
			if err != nil {
				return nil, err
			}
			if len(resp) < 4 || !bytes.Equal(resp[1:4], hdr[1:4]) {
				return nil, rErr("risposta inattesa della radio (comando %02X, tabella %02X, blocco %d)", cmd, tbl, idx)
			}
			return resp[4:], nil
		}
		_ = s.p.Purge()
	}
	if !s.saw {
		return nil, rErr("la radio non risponde")
	}
	return nil, rErr("la radio non risponde più (comando %02X, tabella %02X, blocco %d)", cmd, tbl, idx)
}

func c2Open(ctx context.Context, m *genModel, port string) (*c2Sess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	if err := setRTSOff(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	return &c2Sess{rawSess: s}, nil
}

func (f c2Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := c2Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, c2MemSize, "Collegamento con la radio…")
	img := make([]byte, 0, c2MemSize)
	rd := func(cmd, tbl, idx byte) error {
		d, err := s.xfer(cmd, tbl, idx, nil, true)
		if err != nil {
			return err
		}
		img = append(img, d...)
		prog(len(img), c2MemSize, "Lettura…")
		return nil
	}
	if err := rd(c2CmdIdent, 0, 0); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	if err := rd(c2CmdRead, c2TblSet, 0); err != nil {
		return nil, "", err
	}
	for i := 0; i < 256; i++ {
		if err := rd(c2CmdRead, c2TblChan, byte(i)); err != nil {
			return nil, "", err
		}
	}
	for i := 0; i < c2ExtraCount; i++ {
		if err := rd(c2CmdRead, c2TblExtra, byte(i)); err != nil {
			return nil, "", err
		}
	}
	if err := rd(c2CmdRead, c2TblFinal, 0); err != nil {
		return nil, "", err
	}
	if err := rd(c2CmdRead, c2TblSys, 0); err != nil {
		return nil, "", err
	}
	for i := 0; i < c2RfcalCount; i++ {
		if err := rd(c2CmdRead, c2TblRfcal, byte(i)); err != nil {
			return nil, "", err
		}
	}
	if len(img) != c2MemSize {
		return nil, "", rErr("memoria letta di %d byte invece di %d", len(img), c2MemSize)
	}
	prog(c2MemSize, c2MemSize, "Lettura completata")
	return img, "", nil
}

// DownloadMeta: come CHIRP, la versione del formato dell'immagine nei metadati
func (f c2Fam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	img, note, err := f.Download(ctx, m, port, prog)
	if err != nil {
		return nil, nil, "", err
	}
	return img, map[string]interface{}{"image_version": 1}, note, nil
}

func (f c2Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < c2MemSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), c2MemSize)
	}
	s, err := c2Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, c2MemSize, "Collegamento con la radio…")
	if _, err := s.xfer(c2CmdIdent, 0, 0, nil, true); err != nil {
		return "", genFail("Identificazione", err)
	}
	if _, err := s.xfer(c2CmdRead, c2TblSet, 0, nil, true); err != nil {
		return "", err
	}
	wr := func(tbl, idx byte, off, n int) error {
		if _, err := s.xfer(c2CmdWrite, tbl, idx, img[off:off+n], false); err != nil {
			return err
		}
		if _, err := s.xfer(c2CmdVerify, tbl, idx, nil, true); err != nil {
			return err
		}
		prog(off+n, c2MemSize, "Scrittura…")
		return nil
	}
	if err := wr(c2TblSet, 0, c2SetOff, c2SetSize); err != nil {
		return "", err
	}
	for i := 0; i < 256; i++ {
		if err := wr(c2TblChan, byte(i), c2ChanOff+i*c2ChanSize, c2ChanSize); err != nil {
			return "", err
		}
	}
	for i := 0; i < c2ExtraCount; i++ {
		if err := wr(c2TblExtra, byte(i), c2ExtraOff+i*c2ExtraSize, c2ExtraSize); err != nil {
			return "", err
		}
	}
	if err := wr(c2TblFinal, 0, c2FinalOff, c2FinalWrite); err != nil {
		return "", err
	}
	if _, err := s.xfer(c2CmdFinalize, 0, 0, nil, true); err != nil {
		return "", err
	}
	prog(c2MemSize, c2MemSize, "Scrittura completata")
	return "", nil
}
