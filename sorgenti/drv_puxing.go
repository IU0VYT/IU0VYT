// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: puxing.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Puxing PX-777 e PX-2R (puxing.py di CHIRP, protocollo «Wouxun»)

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

const px777Format = `
#seekto 0x0000;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 _3_unknown_1;
  u8 _2_unknown_1:2,
     power_high:1,
     iswide:1,
     skip:1,
     bclo:2,
     _2_unknown_2:1;
  u8 _4_unknown1:7,
     pttid:1;
  u8 unknown;
} memory[128];

#seekto 0x080A;
struct {
  u8 limits;
  u8 model;
} model;

#seekto 0x0850;
struct {
  u8 name[6];
  u8 pad[2];
} names[128];
`

const px2rFormat = `
#seekto 0x0010;
struct {
  lbcd freq[4];
  lbcd offset[4];
  u8 rx_tone;
  u8 tx_tone;
  u8 duplex:2,
     txdtcsinv:1,
     rxdtcsinv:1,
     simplex:1,
     unknown2:1,
     iswide:1,
     ishigh:1;
  u8 name[5];
} memory[128];
`

const pxCharset777 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ- " // poi spazi fino al 59 (PUXING_CHARSET)
const pxCharset2R = "0123456789- ABCDEFGHIJKLMNOPQRSTUVWXYZ +"

// gamme del PX-777 secondo il codice «limits» della memoria (0xEE…0xF7)
var px777Bands = []band{{67000000, 72000000}, {136000000, 174000000}, {240000000, 260000000}, {350000000, 390000000},
	{400000000, 430000000}, {430000000, 450000000}, {450000000, 470000000}, {470000000, 490000000},
	{400000000, 470000000}, {460000000, 520000000}}

type puxingFam struct{}

type puxingOpt struct{ px2r bool }

func init() {
	genRegister(&genModel{ID: "pux-px777", Vendor: "Puxing", Model: "PX-777", RClass: "Puxing777Radio", Name: "Puxing PX-777",
		Slots: 128, First: 1, NameMax: 6, Charset: pxCharset777, Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}},
		Bands: px777Bands, Narrow: true, HasScan: true, NoTxOff: true, OneDCS: true, DCS: dtcsCodes, Baud: 9600, Fam: puxingFam{}, Opt: &puxingOpt{},
		Notes: "Accendi la radio con il cavo collegato. 128 canali con nomi di 6 caratteri; la gamma di frequenze è quella scritta nella radio. " +
			"Toni: CTCSS in trasmissione, TSQL (stesso tono) o DCS (stesso codice); il blocco del canale occupato si azzera, come in CHIRP. Driver di CHIRP."})
	genRegister(&genModel{ID: "pux-px2r", Vendor: "Puxing", Model: "PX-2R", RClass: "Puxing2RRadio", Name: "Puxing PX-2R",
		Slots: 128, First: 1, NameMax: 5, Charset: pxCharset2R, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 500000000}}, Narrow: true, HasScan: false, NoTxOff: true, OneDCS: true, DCS: dtcsCodes, Baud: 9600,
		Fam: puxingFam{}, Opt: &puxingOpt{px2r: true},
		Notes: "Accendi la radio con il cavo collegato. 128 canali UHF con nomi di 5 caratteri. Toni: CTCSS in trasmissione, TSQL (stesso tono) " +
			"o DCS (stesso codice). Driver di CHIRP."})
}

func pxo(m *genModel) *puxingOpt { return m.Opt.(*puxingOpt) }

func (puxingFam) mem(m *genModel) *bwEl {
	if pxo(m).px2r {
		return bwMust(px2rFormat)
	}
	return bwMust(px777Format)
}
func (f puxingFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f puxingFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f puxingFam) name(m *genModel, i int) *bwEl {
	if pxo(m).px2r {
		return f.rec(m, i).F("name")
	}
	return f.mem(m).F("names").I(i).F("name")
}
func (f puxingFam) Rec(m *genModel, img []byte, i int) []byte {
	r := append([]byte{}, f.rec(m, i).Raw(img)...)
	if !pxo(m).px2r {
		r = append(r, f.name(m, i).Raw(img)...)
	}
	return r
}
func (f puxingFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(m, i).SetRaw(img, raw[:16])
	if !pxo(m).px2r && len(raw) >= 22 {
		f.name(m, i).SetRaw(img, raw[16:22])
	}
}

func (f puxingFam) Match(m *genModel, img []byte) bool {
	if pxo(m).px2r {
		return len(img) == 0x0FE0 && !bytes.HasSuffix(img, []byte("IcomCloneFormat3"))
	}
	return len(img) == 3168 && (img[0x80B] == 0x3A || (img[0x80B] == 0x38 && img[0x80A] == 0xEE))
}

// Bands: PX-777, gamma indicata dalla memoria (get_features di CHIRP)
func (f puxingFam) Bands(m *genModel, img []byte) []band {
	if pxo(m).px2r || len(img) < 0x80C {
		return m.Bands
	}
	if img[0x80B] == 0x3A {
		if k := int(img[0x80A]) - 0xEE; k >= 0 && k < len(px777Bands) {
			return px777Bands[k : k+1]
		}
	}
	return px777Bands[1:2]
}

func (puxingFam) crossband() bool { return true }

func (f puxingFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	if bytes.HasPrefix(r.Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	if pxo(m).px2r {
		return f.get2r(m, img, i)
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10, Tx: r.F("tx_freq").Int(img) * 10}
	c.Scan = r.F("skip").Bool(img)
	c.Narrow = !r.F("iswide").Bool(img)
	noTone := func(e *bwEl) bool {
		b := e.Raw(img)
		return (b[0] == 0 && b[1] == 0) || (b[0] == 0xFF && b[1] == 0xFF)
	}
	tx, rx := r.F("tx_tone"), r.F("rx_tone")
	switch {
	case noTone(tx):
	case tx.Int(img) > 8000 || (!noTone(rx) && rx.Int(img) > 8000):
		// _do_dtcs di puxing.py: un solo codice per trasmissione e ricezione
		if tx.Int(img) < 8000 || rx.Int(img) < 8000 {
			return c // CHIRP: «Split tone not supported»
		}
		dec := func(e *bwEl) (bool, int) {
			if e.Raw(img)[0] == 0xFF {
				return false, 0
			}
			if v := e.Int(img); v > 12000 {
				return true, int(v - 12000)
			} else {
				return false, int(v - 8000)
			}
		}
		tp, tc := dec(tx)
		rp, rc := dec(rx)
		if rc == 0 {
			rc = tc
		}
		if tc == 0 {
			tc = rc
		}
		if tc != rc {
			return c
		}
		c.TxTone = genTone{Mode: "DTCS", Code: tc, Inv: tp}.String()
		c.RxTone = genTone{Mode: "DTCS", Code: tc, Inv: rp}.String()
	default:
		c.TxTone = fmtTone(int(tx.Int(img)))
		if !noTone(rx) {
			c.RxTone = c.TxTone
		}
	}
	c.Power = m.powerSem(map[bool]int{true: 0, false: 1}[r.F("power_high").Bool(img)])
	var nm strings.Builder
	for _, b := range f.name(m, i).Raw(img) {
		if b == 0xFF {
			break
		}
		if int(b) < len(pxCharset777) {
			nm.WriteByte(pxCharset777[b])
		} else {
			nm.WriteByte(' ')
		}
	}
	c.Name = strings.TrimRight(nm.String(), " \t\n\r\x0b\x0c")
	return c
}

func (f puxingFam) get2r(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	c := &Chan{Rx: r.F("freq").Int(img) * 10, Scan: true}
	off := r.F("offset").Int(img) * 10
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx + off
	case 2:
		c.Tx = c.Rx - off
	default:
		c.Tx = c.Rx
	}
	c.Narrow = !r.F("iswide").Bool(img)
	c.Power = m.powerSem(int(r.F("ishigh").Int(img)))
	tx, rx := r.F("tx_tone").Int(img), r.F("rx_tone").Int(img)
	switch {
	case tx >= 0x33:
		if k := int(tx - 0x33); k < len(dtcsCodes) {
			c.TxTone = genTone{Mode: "DTCS", Code: dtcsCodes[k], Inv: r.F("txdtcsinv").Bool(img)}.String()
			c.RxTone = genTone{Mode: "DTCS", Code: dtcsCodes[k], Inv: r.F("rxdtcsinv").Bool(img)}.String()
		}
	case tx > 0:
		if k := int(tx - 1); k < len(chirpTones) {
			c.TxTone = fmtTone(chirpTones[k])
			if rx != 0 {
				c.RxTone = c.TxTone
			}
		}
	}
	var nm strings.Builder
	for _, b := range r.F("name").Raw(img) {
		if b == 0xFF {
			break
		}
		if int(b) < len(pxCharset2R) {
			nm.WriteByte(pxCharset2R[b])
		} else {
			nm.WriteByte(' ')
		}
	}
	c.Name = strings.TrimRight(nm.String(), " \t\n\r\x0b\x0c")
	return c
}

// pxTones: i toni ammessi da queste radio (nessun tono, CTCSS in trasmissione,
// TSQL con lo stesso tono, DCS con lo stesso codice)
func pxTones(c *Chan) (genTone, genTone, error) {
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return t, t, err
	}
	r, err := genToneParse(c.RxTone)
	if err != nil {
		return t, r, err
	}
	switch {
	case t.Mode == "" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
	default:
		return t, r, fmt.Errorf("questa radio ammette solo un tono CTCSS in trasmissione, lo stesso tono in trasmissione e ricezione oppure lo stesso codice DCS")
	}
	return t, r, nil
}

func (f puxingFam) chirpInvalid(m *genModel, c *Chan) bool {
	_, _, err := pxTones(c)
	return err != nil
}

func (f puxingFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	t, rt, err := pxTones(c)
	if err != nil {
		return err
	}
	cs := pxCharset777
	if pxo(m).px2r {
		cs = pxCharset2R
	}
	var nm []byte
	for k := 0; k < len(c.Name); k++ {
		x := strings.IndexByte(cs, c.Name[k])
		if x < 0 {
			return fmt.Errorf("carattere %q non ammesso nel nome", c.Name[k])
		}
		nm = append(nm, byte(x))
	}
	if pxo(m).px2r {
		return f.set2r(m, img, i, c, t, rt, nm)
	}
	r.F("rx_freq").SetInt(img, c.Rx/10)
	r.F("tx_freq").SetInt(img, c.Tx/10)
	r.F("skip").SetBool(img, c.Scan)
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("rx_tone").Fill(img, 0xFF)
	r.F("tx_tone").Fill(img, 0xFF)
	switch t.Mode {
	case "DTCS":
		r.F("tx_tone").SetInt(img, int64(t.Code))
		r.F("rx_tone").SetInt(img, int64(t.Code))
		fl := func(inv bool) byte {
			if inv {
				return 0xC0
			}
			return 0x80
		}
		img[r.F("tx_tone").Off+1] |= fl(t.Inv)
		img[r.F("rx_tone").Off+1] |= fl(rt.Inv)
	case "Tone":
		r.F("tx_tone").SetInt(img, int64(t.Hz10))
		if rt.Mode == "Tone" {
			r.F("rx_tone").SetInt(img, int64(t.Hz10))
		}
	}
	r.F("power_high").SetBool(img, m.powerIndex(c.Power) == 0)
	r.F("bclo").SetInt(img, 0)
	n := f.name(m, i)
	n.Fill(img, 0xFF)
	copy(img[n.Off:n.Off+6], nm)
	return nil
}

func (f puxingFam) set2r(m *genModel, img []byte, i int, c *Chan, t, rt genTone, nm []byte) error {
	r := f.rec(m, i)
	wasEmpty := f.Get(m, img, i) == nil
	r.F("freq").SetInt(img, c.Rx/10)
	switch {
	case c.Tx == c.Rx:
		// CHIRP riscrive lo scostamento memorizzato (0,6 MHz per un canale nuovo)
		if wasEmpty {
			r.F("offset").SetInt(img, 600000/10)
		}
		r.F("duplex").SetInt(img, 0)
	case c.Tx > c.Rx:
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
		r.F("duplex").SetInt(img, 1)
	default:
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
		r.F("duplex").SetInt(img, 2)
	}
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("ishigh").SetBool(img, m.powerIndex(c.Power) == 1)
	switch t.Mode {
	case "DTCS":
		k := slices.Index(dtcsCodes, t.Code)
		if k < 0 {
			return fmt.Errorf("codice DCS %03d non ammesso", t.Code)
		}
		r.F("tx_tone").SetInt(img, int64(k+0x33))
		r.F("rx_tone").SetInt(img, int64(k+0x33))
		r.F("txdtcsinv").SetBool(img, t.Inv)
		r.F("rxdtcsinv").SetBool(img, rt.Inv)
	case "Tone":
		k := slices.Index(chirpTones, t.Hz10)
		if k < 0 {
			return fmt.Errorf("tono %s non ammesso", fmtTone(t.Hz10))
		}
		r.F("tx_tone").SetInt(img, int64(k+1))
		if rt.Mode == "Tone" {
			r.F("rx_tone").SetInt(img, int64(k+1))
		} else {
			r.F("rx_tone").SetInt(img, 0)
		}
	default:
		r.F("tx_tone").SetInt(img, 0)
		r.F("rx_tone").SetInt(img, 0)
	}
	n := r.F("name")
	n.Fill(img, 0xFF)
	copy(img[n.Off:n.Off+5], nm)
	return nil
}

// presentazioni: PX-777 "\x02PROGRA" → 06, "M\x02" → 8 byte, 06 → 06 (fino a
// 10 tentativi); PX-2R "PROGRAM\x02" → 06, 06 → 16 byte
func (s *rawSess) puxingPrep(m *genModel) error {
	if pxo(m).px2r {
		if err := s.send([]byte("PROGRAM\x02")); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, 500*time.Millisecond)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return rErr("la radio non risponde")
		}
		if b[0] != 0x06 {
			return rErr("risposta inattesa della radio")
		}
		if err := s.send(b); err != nil {
			return err
		}
		readUpTo(s.ctx, s.p, 16, 500*time.Millisecond)
		return nil
	}
	var last error
	for k := 0; k < 10; k++ {
		last = func() error {
			if err := s.send([]byte("\x02PROGRA")); err != nil {
				return err
			}
			if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				return rErr("la radio non ha risposto al primo comando")
			}
			if err := s.send([]byte("M\x02")); err != nil {
				return err
			}
			if id := readUpTo(s.ctx, s.p, 8, time.Second); len(id) != 8 {
				return rErr("la radio non ha inviato l'identificativo")
			}
			if err := s.send([]byte{0x06}); err != nil {
				return err
			}
			if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				return rErr("la radio non ha confermato l'identificativo")
			}
			return nil
		}()
		if last == nil || s.ctx.Err() != nil {
			break
		}
		if err := s.sleep(time.Second); err != nil {
			return err
		}
	}
	if s.ctx.Err() != nil {
		return errCancelled
	}
	return last
}

func pxSizes(m *genModel) (rd, wr, bs int) {
	if pxo(m).px2r {
		return 0x0FE0, 0x0FE0, 0x10
	}
	return 0x0C60, 0x0C40, 0x08
}

func (f puxingFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	rd, _, bs := pxSizes(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, rd, "Collegamento con la radio…")
	if err := s.puxingPrep(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, rd, bs, time.Second, prog)
	return data, "", err
}

func (f puxingFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	rd, wr, bs := pxSizes(m)
	if len(img) < rd {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), rd)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, wr, "Collegamento con la radio…")
	if err := s.puxingPrep(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img, 0, wr, bs, time.Second, prog)
}

// fit: toni che la radio non può memorizzare (toni diversi in trasmissione e
// ricezione, tono solo in ricezione) ridotti a quelli ammessi
func (f puxingFam) fit(m *genModel, c *Chan) (bool, string) {
	if _, _, err := pxTones(c); err == nil {
		return true, ""
	}
	t, _ := genToneParse(c.TxTone)
	old := c.TxTone + "/" + c.RxTone
	switch t.Mode {
	case "DTCS":
		c.RxTone = genTone{Mode: "DTCS", Code: t.Code}.String()
	default:
		c.RxTone = ""
	}
	return true, fmt.Sprintf("toni %s non ammessi da questa radio, impostati %s/%s", old, c.TxTone, c.RxTone)
}
