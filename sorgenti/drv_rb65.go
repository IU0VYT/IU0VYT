// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.47: Retevis RB65 (sigla «K22GRAM»), ricavata dal programma del costruttore
// (famiglia «Analog1», blocchi da 8 byte). Il dialogo è quello degli altri modelli
// da 8 byte (02 e sigla, 02, 06, 05 con i 6 byte della password, 06, i 126 blocchi
// da 0x0000 a 0x03EF, chiusura «b»): si usa lo stesso collegamento della famiglia
// della T18 (rbSpec). La memoria dei canali invece è sua: 22 canali da 8 byte da
// 0x0000, con UNA sola frequenza (la radio trasmette dove riceve) e UN solo tono,
// che vale in ricezione e in trasmissione.
//
//	byte 0-3  frequenza in BCD (8 cifre, unità 10 Hz), dal byte basso
//	byte 4-5  tono: FF FF nessuno; CTCSS in BCD; DCS come numero con 0x80
//	          (normale) o 0xC0 (invertito) nel byte alto
//	byte 6    opzioni, con gli stessi bit degli altri modelli «Analog1»
//	byte 7    non usato dal programma (si conserva)
//
// Il bit 1 delle opzioni a 0 segna il canale che il programma mostra come
// «cifrato» (frequenza scritta come numero a 32 bit, tono nascosto): si legge con
// la sola frequenza e, se non lo si modifica, resta byte per byte com'è.
// Verificato sul file dati di fabbrica del programma: i 22 canali FRS, la potenza
// bassa sui canali 8-14, toni validi, riscrittura identica.

import (
	"bytes"
	"context"
	"time"
)

const rb65Format = `
struct {
  lbcd freq[4];       // 0-3
  u8 tone[2];         // 4-5
  u8 speccode:1,      // 6   Spec Code
     compand:1,       //     Compander  0 = Yes, 1 = No
     scramb:1,        //     Scramble  0 = Yes, 1 = No
     scanadd:1,       //     Scan Add  0 = Yes, 1 = No
     ishighpower:1,   //     TX Power  0 = Low, 1 = High
     narrow:1,        //     W/N  0 = Wide, 1 = Narrow
     plain:1,         //     0 = canale «cifrato»
     bcl:1;           //     Busy Lock  0 = On, 1 = Off
  u8 unused;          // 7
} memory[22];
`

type rb65Fam struct{}

type rb65Opt struct{ spec *rbSpec }

func rb65o(m *genModel) *rb65Opt { return m.Opt.(*rb65Opt) }

func init() {
	genRegister(&genModel{ID: "rt-rb65", Vendor: "Retevis", Model: "RB65", RClass: "IU0VYT_Retevis_RB65", Name: "Retevis RB65",
		Slots: 22, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
		Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, NoChirp: true, NoTxOff: true, DCS: uv5rDTCS, Baud: 9600, Fam: rb65Fam{},
		Opt: &rb65Opt{spec: &rbSpec{exit: []byte("b"), quiet: true, magic: []byte("\x02K22GRAM"), memsize: 0x03F0, ranges: [][2]int{{0x0000, 0x03F0}},
			block: 8, tries: 1, to: time.Second, noAck: true, anyID: true, pwStep: true}},
		Notes: rb45Note447})
}

func (rb65Fam) mem(m *genModel) *bwEl              { return bwMust(rb65Format) }
func (f rb65Fam) layout(m *genModel) *bwEl         { return f.mem(m) }
func (f rb65Fam) rec(m *genModel, i int) *bwEl     { return f.mem(m).F("memory").I(i) }
func (f rb65Fam) Match(m *genModel, _ []byte) bool { return false }
func (f rb65Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f rb65Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f rb65Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	b := r.F("freq").Raw(img)
	if bytes.Equal(b, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{}
	if r.F("plain").Bool(img) {
		c.Rx = r.F("freq").Int(img) * 10
		c.RxTone = h777v4ToneDec(r.F("tone").Raw(img), false)
		c.TxTone = c.RxTone
	} else {
		// canale «cifrato»: frequenza come numero a 32 bit, tono nascosto
		c.Rx = (int64(b[3])<<24 | int64(b[2])<<16 | int64(b[1])<<8 | int64(b[0])) * 10
	}
	if c.Rx == 0 {
		return nil
	}
	c.Tx = c.Rx
	c.Scan = !r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("ishighpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

func (f rb65Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if ok, why := f.fit(m, c); !ok {
		return rErr("%s", why)
	}
	tn, err := h777v4ToneEnc(c.RxTone, false)
	if err != nil {
		return err
	}
	if old := f.Get(m, img, i); old != nil && !r.F("plain").Bool(img) {
		if old.Rx == c.Rx && old.RxTone == c.RxTone && old.TxTone == c.TxTone && old.Scan == c.Scan && old.Narrow == c.Narrow &&
			m.powerIndex(old.Power) == m.powerIndex(c.Power) {
			return nil // canale «cifrato» non modificato: resta com'è
		}
	}
	r.F("freq").SetInt(img, c.Rx/10)
	r.F("tone").SetRaw(img, tn)
	r.F("plain").SetBool(img, true)
	r.F("scanadd").SetBool(img, !c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// fit: una sola frequenza e un solo tono per canale
func (f rb65Fam) fit(m *genModel, c *Chan) (bool, string) {
	if c.TxOff || c.Tx != c.Rx {
		return false, "questa radio ha una sola frequenza per canale e trasmette dove riceve"
	}
	if c.RxTone != c.TxTone {
		return false, "questa radio ha un solo tono per canale, uguale in ricezione e in trasmissione"
	}
	if c.Rx%10 != 0 || c.Rx/10 > 99999999 {
		return false, "frequenza che questa radio non può memorizzare"
	}
	return true, ""
}

func (f rb65Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	d, err := rbDownload(ctx, port, m.Baud, rb65o(m).spec, prog)
	return d, "", err
}

func (f rb65Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", rbUpload(ctx, port, m.Baud, rb65o(m).spec, img, prog)
}
