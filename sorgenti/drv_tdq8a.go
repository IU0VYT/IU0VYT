// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tdxone_tdq8a.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: TDXone TD-Q8A (tdxone_tdq8a.py di CHIRP). 9600 baud, «\x02PYNCRAM» → 06,
// 02 → identificativo «P3107…», 06 → 06; blocchi 'R' da 64 byte con conferma e
// 'W' da 16 byte. 128 canali con nomi di 6 caratteri; solo subtoni CTCSS.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

type tdq8aFam struct{}

const tdq8aFormat = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:2,
     dtmf:1,
     unknown2:1,
     bcl:1,
     unknown3:3;
  u8 unknown4:1,
     scan:1,
     highpower:1,
     wide:1,
     unknown5:4;
  u8 unknown6[2];
} memory[128];
#seekto 0x1010;
struct {
  u8 name[6];
  u8 unknown[10];
} names[128];
`

const tdq8aCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ?+-*"

func init() {
	genRegister(&genModel{ID: "tdx-tdq8a", Vendor: "TDXone", Model: "TD-Q8A", RClass: "TDXoneTDQ8A", Name: "TDXone TD-Q8A",
		Slots: 128, First: 1, NameMax: 6, Charset: tdq8aCharset, Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}},
		Bands: []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true, DCS: nil, Baud: 9600,
		Fam: tdq8aFam{}, Notes: "128 canali con nomi di 6 caratteri (cifre, lettere maiuscole, ? + - *). La radio ha solo i subtoni " +
			"CTCSS (niente DCS); DTMF e blocco del canale occupato si conservano. Driver di CHIRP (sperimentale)."})
}

func (tdq8aFam) mem() *bwEl                 { return bwMust(tdq8aFormat) }
func (f tdq8aFam) layout(m *genModel) *bwEl { return f.mem() }
func (f tdq8aFam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("names").I(i)
}
func (f tdq8aFam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f tdq8aFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 32 {
		n.SetRaw(img, raw[16:32])
	}
}
func (f tdq8aFam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x2008 && bytes.HasPrefix(img[0x2000:], []byte("TD-Q8A"))
}
func (tdq8aFam) crossband() bool { return false }

func tdq8aToneDec(v int64) string {
	if v == 0 || v == 0xFFFF || v < 0x0258 {
		return ""
	}
	return fmtTone(int(v))
}

func (f tdq8aFam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	var nm []byte
	for _, b := range n.F("name").Raw(img) {
		if int(b) >= len(tdq8aCharset) {
			break
		}
		nm = append(nm, tdq8aCharset[b])
	}
	c.Name = strings.TrimRight(string(nm), " ")
	c.TxTone = tdq8aToneDec(r.F("txtone").Int(img))
	c.RxTone = tdq8aToneDec(r.F("rxtone").Int(img))
	c.Scan = r.F("scan").Int(img) > 0
	if r.F("highpower").Int(img) > 0 {
		c.Power = 0
	} else {
		c.Power = 2
	}
	c.Narrow = r.F("wide").Int(img) == 0
	return c
}

func (f tdq8aFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.SetRaw(img, append(append(bytes.Repeat([]byte{0xFF}, 12), 0xBF), 0xFF, 0xFF, 0xFF))
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
	if tt.Mode == "DTCS" || rt.Mode == "DTCS" {
		return fmt.Errorf("questa radio non ha il DCS: solo subtoni CTCSS")
	}
	var dtmf, bcl int64 = 0, 1
	if f.Get(m, img, i) != nil {
		dtmf, bcl = r.F("dtmf").Int(img), r.F("bcl").Int(img)
	}
	r.SetRaw(img, append(append(bytes.Repeat([]byte{0xFF}, 12), 0x9F), 0xFF, 0xFF, 0xFF))
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	nm := n.F("name")
	for k := 0; k < 6; k++ {
		v := byte(0xFF)
		if k < len(c.Name) {
			x := strings.IndexByte(tdq8aCharset, c.Name[k])
			if x < 0 {
				return fmt.Errorf("carattere %q non ammesso nel nome", c.Name[k])
			}
			v = byte(x)
		}
		img[nm.Off+k] = v
	}
	tv, rv := int64(0), int64(0)
	if tt.Mode == "Tone" {
		tv = int64(tt.Hz10)
	}
	if rt.Mode == "Tone" {
		rv = int64(rt.Hz10)
	}
	r.F("txtone").SetInt(img, tv)
	r.F("rxtone").SetInt(img, rv)
	r.F("scan").SetBool(img, c.Scan)
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 0)
	r.F("dtmf").SetInt(img, dtmf)
	r.F("bcl").SetInt(img, bcl)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) tdq8aIdent() error {
	var last error
	for k := 0; k < 2; k++ {
		if k > 0 {
			if err := s.sleep(2 * time.Second); err != nil {
				return err
			}
		}
		if last = s.tdq8aTry(); last == nil || last == errCancelled {
			return last
		}
	}
	return last
}

func (s *rawSess) tdq8aTry() error {
	if err := s.send([]byte("\x02PYNCRAM")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, 1500*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("risposta inattesa della radio")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, 1500*time.Millisecond)
	if !bytes.HasPrefix(id, []byte("P3107")) {
		return rErr("identificativo inatteso (% X)", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 1500*time.Millisecond); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio ha rifiutato la copia")
	}
	return nil
}

func (s *rawSess) tdq8aReadBlock(a int) ([]byte, error) {
	if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x40}); err != nil {
		return nil, err
	}
	resp := readUpTo(s.ctx, s.p, 4+0x40, 1500*time.Millisecond)
	if len(resp) != 4+0x40 || resp[0] != 'W' || resp[1] != byte(a>>8) || resp[2] != byte(a) || resp[3] != 0x40 {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		return nil, rErr("risposta della radio non valida per il blocco 0x%04X", a)
	}
	if err := s.sleep(50 * time.Millisecond); err != nil {
		return nil, err
	}
	if err := s.send([]byte{0x06}); err != nil {
		return nil, err
	}
	if b := readUpTo(s.ctx, s.p, 1, 1500*time.Millisecond); len(b) != 1 || b[0] != 0x06 {
		return nil, rErr("la radio non ha inviato il blocco 0x%04X", a)
	}
	return resp[4:], nil
}

func (f tdq8aFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.tdq8aIdent(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x2000; a += 0x40 {
		prog(a, 0x2000, "Lettura…")
		b, err := s.tdq8aReadBlock(a)
		if err != nil {
			return nil, "", err
		}
		data = append(data, b...)
	}
	prog(0x2000, 0x2000, "Lettura completata")
	return append(data, []byte("TD-Q8A  ")...), "", nil
}

var tdq8aRanges = [][2]int{{0x0010, 0x0810}, {0x0E20, 0x0E60}, {0x0F10, 0x0F30}, {0x1010, 0x1810}, {0x1F10, 0x1F30}}

func (f tdq8aFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.tdq8aIdent(); err != nil {
		return "", genFail("Identificazione", err)
	}
	// come il programma del costruttore: prima si legge il blocco 0x0F80
	if _, err := s.tdq8aReadBlock(0x0F80); err != nil {
		return "", err
	}
	for _, rg := range tdq8aRanges {
		for a := rg[0]; a < rg[1]; a += 0x10 {
			prog(a, 0x2000, "Scrittura…")
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)); err != nil {
				return "", err
			}
			if b := readUpTo(ctx, s.p, 1, 1500*time.Millisecond); len(b) != 1 || b[0] != 0x06 {
				return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
			}
		}
	}
	prog(0x2000, 0x2000, "Scrittura completata")
	return "", nil
}
