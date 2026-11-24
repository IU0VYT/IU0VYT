// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.47: Retevis RB66 e RB66P (sigla «D22GRA5»), ricavate dai programmi del
// costruttore (famiglia «Analog1»). Il collegamento è quello delle RB666 (driver
// della H777 nuova versione, drv_h777v4.go): 02 e sigla, 02, 06, 05 con la
// password di programmazione, 06, poi i blocchi; la radio conferma con 66 e può
// rimandare ciò che riceve. Di diverso ci sono i blocchi da 8 byte, la lunghezza
// della memoria (RB66: 28 blocchi, 0x00-0xDF; RB66P: 29 blocchi, 0x00-0xE7), la
// password (RB66: 6 byte; RB66P: 4 byte) e il canale, che ha UNA sola frequenza
// (la radio trasmette dove riceve) e UN solo tono, valido nei due versi:
//
//	byte 0-2  frequenza: 24 bit dal byte basso, da sommare a 0x2000000, unità 10 Hz
//	byte 3-4  mai letti né scritti dai programmi (si conservano)
//	byte 5-6  tono: FF FF nessuno; CTCSS in BCD; DCS come numero con 0x80
//	          (normale) o 0xC0 (invertito) nel byte alto
//	byte 7    opzioni, con gli stessi bit degli altri modelli «Analog1»
//
// Tre byte di frequenza a FF sono un canale libero. Il bit 1 delle opzioni a 0
// segna il canale che il programma mostra come «cifrato» (tono nascosto): si
// legge con la sola frequenza e, se non lo si modifica, resta byte per byte com'è.
// Verificato sui file dati di fabbrica dei due programmi: i 22 canali FRS, la
// potenza bassa sui canali 8-14, toni validi, riscrittura identica.
// Le impostazioni (da 0xD0) sono in settings_rtv447.go. La password di
// programmazione non ferma la lettura né la scrittura (come CHIRP); della RB66 la
// memoria che il programma legge e scrive tiene solo i primi tre byte (0xDD-0xDF).

const rb66Format = `
struct {
  ul24 freq;          // 0-2
  u8 keep[2];         // 3-4
  u8 tone[2];         // 5-6
  u8 speccode:1,      // 7   Spec Code
     compand:1,       //     Compander
     scramb:1,        //     Scramble
     scanadd:1,       //     Scan Add  0 = Yes, 1 = No
     ishighpower:1,   //     TX Power  0 = Low, 1 = High
     narrow:1,        //     W/N  0 = Wide, 1 = Narrow
     plain:1,         //     0 = canale «cifrato»
     bcl:1;           //     Busy Lock
} memory[22];
`

// rb66Fam: collegamento, lettura e scrittura della famiglia della H777 nuova
// versione; memoria dei canali propria
type rb66Fam struct{ h777v4Fam }

func init() {
	for _, x := range []struct {
		id, model string
		size, pw  int
	}{
		{"rt-rb66", "RB66", 0xE0, 6},
		{"rt-rb66p", "RB66P", 0xE8, 4},
	} {
		genRegister(&genModel{ID: x.id, Vendor: "Retevis", Model: x.model, RClass: "IU0VYT_Retevis_" + x.model, Name: "Retevis " + x.model,
			Slots: 22, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
			Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, NoChirp: true, NoTxOff: true, DCS: uv5rDTCS, Baud: 9600, Fam: rb66Fam{},
			Opt: &h777v4Opt{cmd: "D22GRA5", pass: true, ack: 0x66, echo: true, bye: true, rb66: true, pwN: x.pw, bs: 8, memsize: x.size,
				base: 0x2000000, bits: 24},
			Notes: rb45Note447})
	}
}

func (rb66Fam) mem(m *genModel) *bwEl          { return bwMust(rb66Format) }
func (f rb66Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f rb66Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f rb66Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f rb66Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f rb66Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	o := h777v4o(m)
	if r.F("freq").Int(img) == 1<<o.bits-1 {
		return nil // canale libero: tutti e tre i byte della frequenza a FF
	}
	c := &Chan{Rx: (r.F("freq").Int(img) + o.base) * 10}
	c.Tx = c.Rx
	if r.F("plain").Bool(img) { // altrimenti canale «cifrato»: tono nascosto
		c.RxTone = h777v4ToneDec(r.F("tone").Raw(img), false)
		c.TxTone = c.RxTone
	}
	c.Scan = !r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("ishighpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

func (f rb66Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := h777v4o(m)
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
	old := f.Get(m, img, i)
	if old == nil {
		r.Fill(img, 0xFF) // canale nuovo: come una riga vuota del programma del costruttore
	} else if !r.F("plain").Bool(img) && old.Rx == c.Rx && old.RxTone == c.RxTone && old.TxTone == c.TxTone && old.Scan == c.Scan &&
		old.Narrow == c.Narrow && m.powerIndex(old.Power) == m.powerIndex(c.Power) {
		return nil // canale «cifrato» non modificato: resta com'è
	}
	r.F("freq").SetInt(img, c.Rx/10-o.base)
	r.F("tone").SetRaw(img, tn)
	r.F("plain").SetBool(img, true)
	r.F("scanadd").SetBool(img, !c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// fit: una sola frequenza e un solo tono per canale; frequenza nei 24 bit da 0x2000000
func (f rb66Fam) fit(m *genModel, c *Chan) (bool, string) {
	if c.TxOff || c.Tx != c.Rx {
		return false, "questa radio ha una sola frequenza per canale e trasmette dove riceve"
	}
	if c.RxTone != c.TxTone {
		return false, "questa radio ha un solo tono per canale, uguale in ricezione e in trasmissione"
	}
	o := h777v4o(m)
	if v := c.Rx/10 - o.base; c.Rx%10 != 0 || v < 0 || v >= 1<<o.bits-1 {
		return false, "frequenza che questa radio non può memorizzare"
	}
	return true, ""
}
