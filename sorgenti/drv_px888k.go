// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// puxing_px888k.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Puxing PX-888K (puxing_px888k.py di CHIRP): 128 canali con nomi di 6
// caratteri, canali presenti e in scansione in due mappe di bit. 9600 baud;
// «XONLINE» → «PX888D 00 FF»; blocchi 'R'/'W' da 64 byte («R aa aa @»); 'E' → 06.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type px888kFam struct{}

const px888kFormat = `
struct {
  struct {
#seekto 0x0000;
    struct {
      bbcd rx_freq[4];
      bbcd tx_freq[4];
      struct {
        u8 digital:1,
           invert:1,
           high:6;
        u8 low;
      } tone[2];
      u8 duplex_sign:2,
         compander:1,
         txpower:1,
         modulation_width:1,
         txrx_reverse:1,
         bcl:2;
      u8 scrambler_type:3,
         use_scrambler:1,
         opt_signal:2,
         ptt_id_edge:2;
      u8 _unknown_000E[2];
    } data[128];
    struct {
      char entry[6];
      char _unknown_0806[2];
    } names[128];
  } channel_memory;
} mem;
`

var px888kMagic = []byte("PX888D\x00\xff\x13\x40\x17\x60\x40\x00\x48\x00")

func init() {
	genRegister(&genModel{ID: "px-888k", Vendor: "Puxing", Model: "PX-888K", RClass: "Puxing_PX888K_Radio",
		Name: "Puxing PX-888K", Slots: 128, First: 1, NameMax: 6, Charset: charsetASCII,
		Powers: []genPower{{"Bassa 0,6 W", 0.6, 2}, {"Alta 4,5 W", 4.5, 0}},
		Bands:  []band{{134000000, 176000000}, {400000000, 480000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes,
		NoTxOff: true, Baud: 9600, Fam: px888kFam{},
		Notes: "128 canali con nomi di 6 caratteri; la radio non ha canali in sola ricezione. Blocco del canale occupato, " +
			"compander, scrambler, PTT ID e segnalazione si conservano. Driver di CHIRP."})
}

func (px888kFam) mem() *bwEl                 { return bwMust(px888kFormat) }
func (f px888kFam) layout(m *genModel) *bwEl { return f.mem() }
func (f px888kFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem().F("mem").F("channel_memory")
	return b.F("data").I(i), b.F("names").I(i).F("entry")
}
func (f px888kFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f px888kFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 22 {
		n.SetRaw(img, raw[16:22])
	}
}
func (f px888kFam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x1000 && bytes.Equal(img[0xC40:0xC50], px888kMagic)
}

func px888kBit(img []byte, base, i int) bool { return img[base+i/8]&(1<<(i%8)) != 0 }
func px888kSetBit(img []byte, base, i int, on bool) {
	if on {
		img[base+i/8] |= 1 << (i % 8)
	} else {
		img[base+i/8] &^= 1 << (i % 8)
	}
}

func px888kToneDec(hi, lo byte) string {
	h := hi & 0x3F
	if h == 0x3F && lo == 0xFF {
		return ""
	}
	if hi&0x80 != 0 {
		code := int(h&0x0F)*100 + int(lo>>4)*10 + int(lo&0x0F)
		if idxInt(dtcsCodes, code) < 0 {
			return ""
		}
		return genTone{Mode: "DTCS", Code: code, Inv: hi&0x40 != 0}.String()
	}
	hz10 := int(h>>4)*1000 + int(h&0x0F)*100 + int(lo>>4)*10 + int(lo&0x0F)
	if idxInt(chirpTones, hz10) < 0 {
		return ""
	}
	return fmtTone(hz10)
}

func px888kToneEnc(t genTone) (byte, byte) {
	switch t.Mode {
	case "Tone":
		v := t.Hz10
		lo := byte((v/10%10)<<4 | v%10)
		hi := byte((v/1000)<<4 | v/100%10)
		return hi, lo
	case "DTCS":
		v := t.Code
		hi := byte(0x80 | v/100)
		if t.Inv {
			hi |= 0x40
		}
		return hi, byte((v/10%10)<<4 | v%10)
	}
	return 0xFF, 0xFF
}

func (f px888kFam) Get(m *genModel, img []byte, i int) *Chan {
	if !px888kBit(img, 0xC20, i) {
		return nil
	}
	r, n := f.recs(i)
	raw := r.Raw(img)
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10, Tx: r.F("tx_freq").Int(img) * 10}
	var nm []byte
	for _, b := range n.Raw(img) {
		if b == 0xFF {
			break
		}
		nm = append(nm, b)
	}
	c.Name = string(nm)
	c.Scan = px888kBit(img, 0xC30, i)
	c.TxTone = px888kToneDec(raw[8], raw[9])
	c.RxTone = px888kToneDec(raw[10], raw[11])
	c.Narrow = !r.F("modulation_width").Bool(img)
	c.Power = m.powerSem(int(r.F("txpower").Int(img)))
	return c
}

func (f px888kFam) toneOK(s string) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch {
	case t.Mode == "Tone" && idxInt(chirpTones, t.Hz10) < 0:
		return fmt.Errorf("tono %s non disponibile su questa radio", s)
	case t.Mode == "DTCS" && idxInt(dtcsCodes, t.Code) < 0:
		return fmt.Errorf("codice DCS %s non disponibile su questa radio", s)
	}
	return nil
}

func (f px888kFam) chirpInvalid(m *genModel, c *Chan) bool {
	return f.toneOK(c.TxTone) != nil || f.toneOK(c.RxTone) != nil
}

func (f px888kFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		px888kSetBit(img, 0xC20, i, false)
		px888kSetBit(img, 0xC30, i, false)
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	for _, s := range []string{c.TxTone, c.RxTone} {
		if err := f.toneOK(s); err != nil {
			return err
		}
	}
	tt, _ := genToneParse(c.TxTone)
	rt, _ := genToneParse(c.RxTone)
	if !px888kBit(img, 0xC20, i) {
		r.SetRaw(img, []byte{0x14, 0x61, 0x00, 0x00, 0x14, 0x61, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xC8, 0x00, 0xFF, 0xFF})
	}
	nb := n.Raw(img)
	var nm []byte
	for _, b := range bytes.TrimRight([]byte(c.Name), " \t\n\r\x0b\x0c") {
		if b >= 0x20 && b < 0x7F && len(nm) < 6 {
			nm = append(nm, b)
		}
	}
	for k := range nb {
		v := byte(0xFF)
		if k < len(nm) {
			v = nm[k]
		}
		img[n.Off+k] = v
	}
	px888kSetBit(img, 0xC20, i, true)
	px888kSetBit(img, 0xC30, i, c.Scan)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	r.F("tx_freq").SetInt(img, c.Tx/10)
	raw := img[r.Off : r.Off+16]
	h, l := px888kToneEnc(tt)
	raw[8] = h
	raw[9] = l
	h, l = px888kToneEnc(rt)
	raw[10], raw[11] = h, l
	r.F("modulation_width").SetBool(img, !c.Narrow)
	r.F("txpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) px888kHello() error {
	for k := 0; k < 10; k++ {
		if err := s.send([]byte("XONLINE")); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 8, time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if bytes.Equal(b, []byte("PX888D\x00\xff")) {
			return nil
		}
	}
	return rErr("la radio non ha risposto a nessuno dei 10 tentativi di collegamento")
}

func (s *rawSess) px888kEnd() error {
	if err := s.send([]byte{'E'}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio non ha confermato la fine della copia")
	}
	return nil
}

func (f px888kFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	if err := s.px888kHello(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x1000; a += 0x40 {
		prog(a, 0x1000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), '@'}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x44, time.Second)
		if len(b) != 0x44 || !bytes.Equal(b[:4], []byte{'W', byte(a >> 8), byte(a), '@'}) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, b[4:]...)
	}
	if err := s.px888kEnd(); err != nil {
		return nil, "", err
	}
	prog(0x1000, 0x1000, "Lettura completata")
	return data, "", nil
}

func (f px888kFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x1000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x1000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0xFC0, "Collegamento con la radio…")
	if err := s.px888kHello(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0xFC0; a += 0x40 {
		prog(a, 0xFC0, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), '@'}, img[a:a+0x40]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	if err := s.px888kEnd(); err != nil {
		return "", err
	}
	prog(0xFC0, 0xFC0, "Scrittura completata")
	return "", nil
}
