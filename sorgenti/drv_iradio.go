// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// iradio_common.py, iradio_uv_5118.py, iradio_uv_5118plus.py, radtel_rt880g.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: radio con il protocollo «iRadio»: Abbree AR-518 (iRadio UV-5118),
// Ruyage UV58Plus (iRadio UV-5118plus) e Radtel RT-880G. Ingresso in
// programmazione con una sequenza di 5 byte confermata da 06; blocchi letti con
// 'R' indirizzo (16 bit) somma, risposta 'R' indirizzo dati somma; scrittura a
// blocchi con la somma e conferma 06; uscita con un'altra sequenza.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

type iradioKind int

const (
	irAR518  iradioKind = iota // iRadio UV-5118 (Abbree AR-518)
	irUV58P                    // iRadio UV-5118plus (Ruyage UV58Plus)
	irRT880G                   // Radtel RT-880G
	irRT890M                   // Radtel RT-890 con il firmware M7OCM (driver di M7OCM per CHIRP)
	// 4.19: Radtel RT-817A / RT-871A, dal programma del costruttore («Radtel RT-871A
	// CPS», stessa base di codice di quello della RT-860): memoria della RT-880G senza
	// le zone, con un solo blocco per la radio FM
	irRT871A
)

type iradioFam struct{ kind iradioKind }

const iradio5118Format = `
#seekto 0x0050;
struct memory {
  ul32 rxfreq;
  ul16 rx_tone;
  ul32 txfreq;
  ul16 tx_tone;
  u8 lowpower:1,
     isnarrow:1,
     unknown1:1,
     unknown2:3,
     bcl:1,
     scan:1;
  u8 unknown3[3];
};
struct memory channels[128];
`

const iradio5118pFormat = `
struct memory {
  ul32 rxfreq;
  ul16 rx_tone;
  ul32 txfreq;
  ul16 tx_tone;
  ul24 mutecode;
  u8 unknown_0:2,
     mutetype:2,
     unknown_1:4;
  u8 isnarrow:1,
     lowpower:1,
     scan:1,
     bcl:2,
     is_airband:1,
     unknown_3:1,
     unknown_4:1;
  u8 unknown_5;
  u8 unused_0:4,
     scno:4;
  u8 unknown_6[3];
  char name[10];
};
#seekto 0x1000;
struct memory channels[999];
`

// 4.9: driver di M7OCM per il firmware M7OCM della Radtel RT-890
// (github.com/M7OCM/890, chirp-next-radtel-rt-890-m7ocm.py): modulazione del
// canale (FM, AM, LSB, USB) e otto liste di scansione
const iradio890mFormat = `
struct memory {
  ul32 rxfreq;
  ul16 rx_tone;
  ul32 txfreq;
  ul16 tx_tone;
  ul24 mutecode;
  u8 unknown_0:2,
     mutetype:2,
     unknown_1:4;
  u8 isnarrow:1,
     lowpower:1,
     scan:1,
     bcl:2,
     modulation_type:2,
     unknown_4:1;
  u8 unknown_5;
  u8 unused_0:4,
     scno:4;
  u8 scanlist8:1,
     scanlist7:1,
     scanlist6:1,
     scanlist5:1,
     scanlist4:1,
     scanlist3:1,
     scanlist2:1,
     scanlist1:1;
  u8 unknown_6[2];
  char name[10];
};

#seekto 0x1000;
struct memory channels[999];
`

const rt880gFormat = `
struct tone {
  ul16 unknown:2,
       is_dtcs:1,
       tone_or_inverted:1,
       tone_val:12;
};
struct memory {
  u8 rx_mode;
  u8 limit_rx_tx;
  u8 enabled;
  u8 isnarrow;
  struct tone rx_tone;
  ul32 rxfreq;
  ul32 txfreq;
  struct tone tx_tone;
  u8 power;
  u8 bcl;
  u8 dcs_encrypt:3,
     tot:5;
  u8 scan_remove:1,
     tail_tone:3,
     scrambler:4;
  ul32 mutecode;
  u8 unknown[8];
  char name[16];
};
#seekto 0x2000;
struct memory channels[1024];
`

// codici DCS di queste radio (RB15_DTCS / DTCS_CODES: quelli di CHIRP più 645)
var iradioDCS = withCodes(dtcsCodes, 645)

func init() {
	// gamme come in CHIRP (VALID_BANDS); le gamme solo in ricezione sono comprese
	genRegister(
		&genModel{ID: "ab-ar518", Vendor: "Abbree", Model: "AR-518", RClass: "AbbreeAR518Radio", Name: "Abbree AR-518 (iRadio UV-5118)",
			Slots: 128, First: 1, NameMax: 0, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
			Bands: []band{{20000000, 64000000}, {108000000, 660000000}}, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: iradioDCS, Baud: 9600, Fam: iradioFam{irAR518},
			Notes: "Radio senza nomi dei canali (128 canali); il blocco del canale occupato si conserva. Driver di CHIRP (iRadio UV-5118)."},
		&genModel{ID: "ry-uv58plus", Vendor: "Ruyage", Model: "UV58Plus", RClass: "RuyageUV58PlusRadio", Name: "Ruyage UV58Plus (iRadio UV-5118plus)",
			Slots: 999, First: 1, NameMax: 10, Charset: charsetAlnum + "`{|}!\"#$%&'()*+,-./:;<=>?@[]^_",
			Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
			Bands:  []band{{108000000, 560000000}}, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: iradioDCS, Baud: 115200, Fam: iradioFam{irUV58P},
			Notes: "999 canali con nomi di 10 caratteri; blocco del canale occupato, codice di silenziamento e numero SC si conservano; " +
				"nella gamma aeronautica (sotto 136 MHz) il canale si scrive in AM. Driver di CHIRP (iRadio UV-5118plus)."},
		&genModel{ID: "rtl-rt890-m7ocm", Vendor: "Radtel", Model: "RT-890 M7OCM", RClass: "RuyageUV58PlusRadioM7OCM",
			Name: "Radtel RT-890 (firmware M7OCM)", Slots: 999, First: 1, NameMax: 10, Charset: charsetAlnum + "`{|}!\"#$%&'()*+,-./:;<=>?@[]^_",
			Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
			Bands:  []band{{10000000, 1300000000}}, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: iradioDCS, Baud: 115200, Fam: iradioFam{irRT890M},
			Notes: "Radtel RT-890 (PCB 2.0) con il firmware M7OCM: 999 canali con nomi di 10 caratteri, 10-1300 MHz. Modulazione " +
				"(AM, LSB, USB) dei canali che non cambiano frequenza, liste di scansione, blocco del canale occupato, codice di " +
				"silenziamento e numero SC si conservano. Driver di M7OCM per CHIRP (github.com/M7OCM/890, GPL v2 o successiva)."},
		&genModel{ID: "rtl-rt880g", Vendor: "Radtel", Model: "RT-880G", RClass: "RT880G", Name: "Radtel RT-880G",
			Slots: 1024, First: 1, NameMax: 16, Charset: charsetASCII,
			Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 5 W", 5, 1}, {"Alta 10 W", 10, 0}},
			Bands:  []band{{18000000, 620000000}, {840000000, 1000000000}}, Narrow: true, HasScan: true, NoTxOff: true,
			DCS: iradioDCS, Baud: 115200, Fam: iradioFam{irRT880G},
			Notes: "1024 canali con nomi di 16 caratteri (lettura di circa 860 KiB a 115200 baud: un paio di minuti). " +
				"Modulazione (AM, USB) dei canali che non cambiano frequenza, scrambler, toni di coda, cifratura DCS e limiti RX/TX " +
				"si conservano; nella gamma aeronautica il canale si scrive in AM. Le zone restano come sono. Driver di CHIRP."},
		&genModel{ID: "rtl-rt871a", Vendor: "Radtel", Model: "RT-817A", Variant: "RT-871A", RClass: "IU0VYT_Radtel_RT871A",
			Name:  "Radtel RT-817A / RT-871A",
			Slots: 1024, First: 1, NameMax: 16, Charset: charsetASCII,
			Powers: []genPower{{"Bassa", 0, 2}, {"Media", 0, 1}, {"Alta", 0, 0}},
			Bands:  []band{{18000000, 660000000}, {840000000, 1000000000}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
			DCS: iradioDCS, Baud: 115200, Fam: iradioFam{irRT871A},
			Notes: "1024 canali con nomi di 16 caratteri, 115200 baud. Protocollo e memoria ricavati dal programma del costruttore " +
				"(«Radtel RT-871A CPS»), che usa il collegamento e il canale di 48 byte della Radtel RT-880G senza le zone. Si leggono e si " +
				"scrivono le stesse aree del programma del costruttore (impostazioni, canali, radio FM; non il logo). Modulazione " +
				"(AM, USB) dei canali che non cambiano frequenza, scrambler, toni di coda, codice del canale e le altre opzioni " +
				"si conservano; nella gamma aeronautica il canale si scrive in AM."},
	)
}

func (f iradioFam) fmtStr() string {
	switch f.kind {
	case irUV58P:
		return iradio5118pFormat
	case irRT890M:
		return iradio890mFormat
	case irRT880G, irRT871A:
		return rt880gFormat
	}
	return iradio5118Format
}

func (f iradioFam) mem() *bwEl               { return bwMust(f.fmtStr()) }
func (f iradioFam) layout(m *genModel) *bwEl { return f.mem() }
func (f iradioFam) rec(i int) *bwEl          { return f.mem().F("channels").I(i) }
func (f iradioFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f iradioFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f iradioFam) Match(m *genModel, img []byte) bool                { return false }
func (iradioFam) crossband() bool                                     { return false }

// toni (_decode_tone / _encode_tone di iradio_uv_5118.py): 0x3000 nessuno,
// 0x1000 DCS normale, 0x2000 DCS invertito (codice ottale in 9 bit)
func iradioToneDec(v int64) string {
	switch {
	case v == 0x3000:
		return ""
	case v&0x1000 != 0:
		return iradioDCSTone(v&0x1FF, false)
	case v&0x2000 != 0:
		return iradioDCSTone(v&0x1FF, true)
	}
	return fmtTone(int(v))
}

func iradioDCSTone(oct int64, inv bool) string {
	code := int(oct>>6)*100 + int(oct>>3&7)*10 + int(oct&7)
	return genTone{Mode: "DTCS", Code: code, Inv: inv}.String()
}

func iradioOct(code int) int64 { return int64(code/100)<<6 | int64(code/10%10)<<3 | int64(code%10) }

func iradioToneEnc(t genTone) int64 {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10)
	case "DTCS":
		if t.Inv {
			return iradioOct(t.Code) | 0x2800
		}
		return iradioOct(t.Code) | 0x1800
	}
	return 0x3000
}

// rt880gToneDec: struct tone (is_dtcs, tone_or_inverted, valore)
func rt880gToneDec(t *bwEl, img []byte) string {
	dcs, ti, v := t.F("is_dtcs").Int(img) > 0, t.F("tone_or_inverted").Int(img) > 0, t.F("tone_val").Int(img)
	switch {
	case !dcs && !ti:
		return ""
	case dcs:
		return iradioDCSTone(v, ti)
	}
	return fmtTone(int(v))
}

func rt880gToneEnc(t *bwEl, img []byte, g genTone) {
	var dcs, ti, v int64
	switch g.Mode {
	case "Tone":
		ti, v = 1, int64(g.Hz10)
	case "DTCS":
		dcs, v = 1, iradioOct(g.Code)
		if g.Inv {
			ti = 1
		}
	}
	t.F("is_dtcs").SetInt(img, dcs)
	t.F("tone_or_inverted").SetInt(img, ti)
	t.F("tone_val").SetInt(img, v)
}

func (f iradioFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	rx := r.F("rxfreq").Int(img)
	if rx == 0 || rx == 0xFFFFFFFF {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: r.F("txfreq").Int(img) * 10, Scan: true}
	switch f.kind {
	case irRT880G, irRT871A:
		c.TxTone = rt880gToneDec(r.F("tx_tone"), img)
		c.RxTone = rt880gToneDec(r.F("rx_tone"), img)
		c.Narrow = r.F("rx_mode").Int(img) == 0 && r.F("isnarrow").Int(img) == 1
		switch r.F("power").Int(img) {
		case 2:
			c.Power = 0
		case 1:
			c.Power = 1
		default:
			c.Power = 2
		}
		c.Scan = r.F("scan_remove").Int(img) == 0
		c.Name = strings.ReplaceAll(strings.TrimRight(string(r.F("name").Raw(img)), " "), "\xFF", "")
		c.Name = latin1Str(c.Name)
	default:
		c.TxTone = iradioToneDec(r.F("tx_tone").Int(img))
		c.RxTone = iradioToneDec(r.F("rx_tone").Int(img))
		c.Narrow = r.F("isnarrow").Int(img) > 0
		c.Power = m.powerSem(int(r.F("lowpower").Int(img)))
		c.Scan = r.F("scan").Int(img) > 0
		if f.kind == irUV58P || f.kind == irRT890M {
			c.Name = latin1Str(strings.ReplaceAll(strings.TrimRight(string(r.F("name").Raw(img)), " "), "\xFF", " "))
		}
	}
	return c
}

func (f iradioFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		switch f.kind {
		case irUV58P, irRT890M:
			r.SetRaw(img, append(bytes.Repeat([]byte{0xFF}, 22), bytes.Repeat([]byte{0x20}, 10)...))
		default:
			r.Fill(img, 0xFF)
		}
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
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
	ff4 := bytes.Repeat([]byte{0xFF}, 4)
	tmpl := append(append(append(append(append([]byte{}, ff4...), 0x00, 0x30), ff4...), 0x00, 0x30), make([]byte, 4)...)
	switch f.kind {
	case irAR518:
		bcl := old != nil && r.F("bcl").Int(img) > 0
		r.SetRaw(img, tmpl)
		r.F("bcl").SetBool(img, bcl)
	case irUV58P:
		var bcl, mt, mc, sc int64
		if old != nil {
			bcl, mt, mc, sc = r.F("bcl").Int(img), r.F("mutetype").Int(img), r.F("mutecode").Int(img), r.F("scno").Int(img)
		}
		r.SetRaw(img, append(append(tmpl, make([]byte, 6)...), bytes.Repeat([]byte{0x20}, 10)...))
		r.F("bcl").SetInt(img, bcl)
		r.F("mutetype").SetInt(img, mt)
		r.F("mutecode").SetInt(img, mc)
		r.F("scno").SetInt(img, sc)
		// gamma aeronautica in AM (come la lettura di CHIRP, che lo imposta)
		r.F("is_airband").SetBool(img, c.Rx < 136000000)
		nm := []byte(strings.TrimRight(c.Name, "\xFF"))
		if len(nm) > 10 {
			nm = nm[:10]
		}
		r.F("name").SetRaw(img, append(nm, bytes.Repeat([]byte{0x20}, 10-len(nm))...))
	case irRT890M:
		keep := []string{"bcl", "mutetype", "mutecode", "scno", "scanlist1", "scanlist2", "scanlist3", "scanlist4",
			"scanlist5", "scanlist6", "scanlist7", "scanlist8"}
		vals := make([]int64, len(keep))
		var mod int64
		if old != nil {
			for k, n := range keep {
				vals[k] = r.F(n).Int(img)
			}
			if old.Rx == c.Rx && !c.Narrow && !old.Narrow {
				mod = r.F("modulation_type").Int(img) // AM, LSB o USB del canale si conservano
			}
		}
		r.SetRaw(img, append(append(tmpl, make([]byte, 6)...), bytes.Repeat([]byte{0x20}, 10)...))
		for k, n := range keep {
			r.F(n).SetInt(img, vals[k])
		}
		r.F("modulation_type").SetInt(img, mod)
		nm := []byte(strings.TrimRight(c.Name, "\xFF"))
		if len(nm) > 10 {
			nm = nm[:10]
		}
		r.F("name").SetRaw(img, append(nm, bytes.Repeat([]byte{0x20}, 10-len(nm))...))
	case irRT880G, irRT871A:
		// impostazioni del canale conservate (extra di get_memory)
		var enc, mc, tail, scr, bcl, lim, tot, mode int64
		// RT-871A: il codice del canale occupa anche gli 8 byte che la RT-880G non usa
		var code2 []byte
		if old != nil && f.kind == irRT871A {
			code2 = append([]byte{}, r.F("unknown").Raw(img)...)
		}
		if old != nil {
			enc, mc, tail, scr = r.F("dcs_encrypt").Int(img), r.F("mutecode").Int(img), r.F("tail_tone").Int(img), r.F("scrambler").Int(img)
			if bcl = r.F("bcl").Int(img); bcl > 2 {
				bcl = 0
			}
			lim, tot = r.F("limit_rx_tx").Int(img), r.F("tot").Int(img)
			if old.Rx == c.Rx {
				mode = r.F("rx_mode").Int(img) // AM o USB del canale si conservano
			}
		}
		if c.Rx >= 108000000 && c.Rx <= 135999999 {
			mode = 1 // gamma aeronautica: AM (validate_memory di CHIRP)
		}
		r.SetRaw(img, append(make([]byte, 32), bytes.Repeat([]byte{0x20}, 16)...))
		r.F("enabled").SetInt(img, 1)
		if code2 != nil {
			r.F("unknown").SetRaw(img, code2)
		}
		nm := []byte(strings.TrimRight(c.Name, " "))
		if len(nm) > 16 {
			nm = nm[:16]
		}
		r.F("name").SetRaw(img, append(nm, bytes.Repeat([]byte{0xFF}, 16-len(nm))...))
		r.F("scan_remove").SetBool(img, !c.Scan)
		switch {
		case mode == 1 || mode == 2:
			r.F("rx_mode").SetInt(img, mode)
		case c.Narrow:
			r.F("isnarrow").SetInt(img, 1)
		}
		switch m.powerIndex(c.Power) {
		case 2:
			r.F("power").SetInt(img, 2)
		case 1:
			r.F("power").SetInt(img, 1)
		}
		rt880gToneEnc(r.F("tx_tone"), img, tt)
		rt880gToneEnc(r.F("rx_tone"), img, rt)
		r.F("dcs_encrypt").SetInt(img, enc)
		r.F("mutecode").SetInt(img, mc)
		r.F("tail_tone").SetInt(img, tail)
		r.F("scrambler").SetInt(img, scr)
		r.F("bcl").SetInt(img, bcl)
		r.F("limit_rx_tx").SetInt(img, lim)
		r.F("tot").SetInt(img, tot)
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	r.F("txfreq").SetInt(img, c.Tx/10)
	if f.kind != irRT880G && f.kind != irRT871A {
		r.F("scan").SetBool(img, c.Scan)
		r.F("isnarrow").SetBool(img, c.Narrow)
		r.F("tx_tone").SetInt(img, iradioToneEnc(tt))
		r.F("rx_tone").SetInt(img, iradioToneEnc(rt))
		r.F("lowpower").SetBool(img, m.powerIndex(c.Power) == 1)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type iradioProto struct {
	enter, exit []byte
	enterTO     time.Duration
	block       int
	readOff     int // indirizzo di lettura = blocco (o indirizzo) + readOff
	byBlock     bool
	size        int
}

func (f iradioFam) proto() iradioProto {
	switch f.kind {
	case irUV58P, irRT890M:
		return iradioProto{enter: []byte("58\x05\x10\x82"), exit: []byte("58\x05\xEE\x60"), enterTO: 250 * time.Millisecond,
			block: 0x80, readOff: 0x7820, byBlock: true, size: 0x140 * 0x80}
	case irRT880G:
		return iradioProto{enter: []byte("4R\x05\x10\x9b"), exit: []byte("4R\x05\xEE\x79"), enterTO: 250 * time.Millisecond,
			block: 0x400, readOff: 8, byBlock: true, size: 0x35b * 0x400}
	case irRT871A:
		// la memoria arriva fino al blocco della radio FM (blocco 0x358 della radio)
		return iradioProto{enter: []byte("4R\x05\x10\x9b"), exit: []byte("4R\x05\xEE\x79"), enterTO: 250 * time.Millisecond,
			block: 0x400, readOff: 8, byBlock: true, size: 0x351 * 0x400}
	}
	return iradioProto{enter: []byte("93\x05\x10\x81"), exit: []byte("93\x05\xEE\x5F"), enterTO: 750 * time.Millisecond,
		block: 0x10, readOff: 0x0340, size: 0x0850}
}

func (s *rawSess) iradioEnter(p iradioProto) error {
	for k := 0; k < 5; k++ {
		if err := s.send(p.enter); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, p.enterTO)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 1 && b[0] == 0x06 {
			return nil
		}
		if len(b) == 0 && k == 4 {
			return rErr("la radio non risponde")
		}
	}
	return rErr("la radio non è entrata in programmazione dopo cinque tentativi: controlla il cavo e riaccendi la radio")
}

// iradioRead: un blocco (addr: indirizzo o numero del blocco, già con readOff)
func (s *rawSess) iradioRead(addr, n int) ([]byte, error) {
	cmd := []byte{'R', byte(addr >> 8), byte(addr)}
	if err := s.send(append(cmd, sum8(cmd))); err != nil {
		return nil, err
	}
	resp := readUpTo(s.ctx, s.p, 3+n+1, 2*time.Second)
	if len(resp) != 3+n+1 || !bytes.Equal(resp[:3], cmd) || resp[len(resp)-1] != sum8(resp[:len(resp)-1]) {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		return nil, rErr("lettura del blocco 0x%04X non riuscita", addr)
	}
	return resp[3 : 3+n], nil
}

func (s *rawSess) iradioWrite(cmd byte, addr int, data []byte) error {
	pk := append([]byte{cmd, byte(addr >> 8), byte(addr)}, data...)
	if err := s.send(append(pk, sum8(pk))); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 || b[0] != 0x06 {
		if s.ctx.Err() != nil {
			return errCancelled
		}
		return rErr("la radio non ha accettato il blocco 0x%04X", addr)
	}
	return nil
}

func (f iradioFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	p := f.proto()
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, p.size, "Collegamento con la radio…")
	if err := s.iradioEnter(p); err != nil {
		return nil, "", genFail("Ingresso in programmazione", err)
	}
	defer s.send(p.exit)
	if f.kind == irRT871A {
		// come il programma del costruttore: solo impostazioni, canali e radio FM
		data := bytes.Repeat([]byte{0xFF}, p.size)
		done, total := 0, 0
		for _, z := range rt871aRegions {
			total += z[1]
		}
		for _, z := range rt871aRegions {
			for b := 0; b < z[1]; b++ {
				if ctx.Err() != nil {
					return nil, "", errCancelled
				}
				prog(done*p.block, total*p.block, "Lettura…")
				a := z[0] + b*p.block
				d, err := s.iradioRead(a/p.block+p.readOff, p.block)
				if err != nil {
					return nil, "", err
				}
				copy(data[a:], d)
				done++
			}
		}
		prog(total*p.block, total*p.block, "Lettura completata")
		return data, "", nil
	}
	var data []byte
	for a := 0; a < p.size; a += p.block {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, p.size, "Lettura…")
		addr := a
		if p.byBlock {
			addr = a / p.block
		}
		b, err := s.iradioRead(addr+p.readOff, p.block)
		if err != nil {
			return nil, "", err
		}
		data = append(data, b...)
	}
	prog(p.size, p.size, "Lettura completata")
	return data, "", nil
}

// zone scritte dalla RT-880G: indirizzo, blocchi da 1 KiB, identificativo della zona
var rt880gRegions = [][3]int{{0x0000, 4, 0x90}, {0x2000, 48, 0x91}, {0x1a000, 128, 0x92}, {0xd4000, 4, 0x99}}

// zone lette e scritte dalla RT-871A (programma del costruttore): impostazioni
// (blocchi 8-11 della radio), canali (0x10-0x3F), radio FM (0x358)
var rt871aRegions = [][3]int{{0x0000, 4, 0x90}, {0x2000, 48, 0x91}, {0xd4000, 1, 0x99}}

func (f iradioFam) regions() [][3]int {
	if f.kind == irRT871A {
		return rt871aRegions
	}
	return rt880gRegions
}

func (f iradioFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	p := f.proto()
	if len(img) < p.size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), p.size)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, p.size, "Collegamento con la radio…")
	if err := s.iradioEnter(p); err != nil {
		return "", genFail("Ingresso in programmazione", err)
	}
	defer s.send(p.exit)
	switch f.kind {
	case irAR518:
		for a := 0; a < p.size; a += p.block {
			prog(a, p.size, "Scrittura…")
			if err := s.iradioWrite('W', a+p.readOff, img[a:a+p.block]); err != nil {
				return "", err
			}
		}
	case irUV58P, irRT890M:
		// come il programma del costruttore: prima si legge il primo blocco
		if _, err := s.iradioRead(p.readOff, p.block); err != nil {
			return "", err
		}
		for b := 0; b < p.size/p.block; b++ {
			prog(b*p.block, p.size, "Scrittura…")
			if err := s.iradioWrite('I', b, img[b*p.block:(b+1)*p.block]); err != nil {
				return "", err
			}
		}
	case irRT880G, irRT871A:
		// come il programma del costruttore: prima si legge il blocco 0
		// (do_upload di CHIRP: senza lo scostamento della lettura)
		if _, err := s.iradioRead(0, p.block); err != nil {
			return "", err
		}
		done, total := 0, 0
		for _, z := range f.regions() {
			total += z[1]
		}
		for _, z := range f.regions() {
			for b := 0; b < z[1]; b++ {
				prog(done*p.block, total*p.block, "Scrittura…")
				a := z[0] + b*p.block
				if err := s.iradioWrite(byte(z[2]), b, img[a:a+p.block]); err != nil {
					return "", err
				}
				done++
			}
		}
	}
	prog(p.size, p.size, "Scrittura completata")
	return "", nil
}
