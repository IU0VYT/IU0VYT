// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// id880.py (4.39).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.39: Icom D-STAR ID-80H (palmare) e ID-880H (veicolare): 1000 canali (da 0 a
// 999) con nomi di 8 caratteri, copia veloce (38400 baud). Stessa memoria per i
// due modelli; si distinguono dal modello dichiarato dalla radio e, nei file,
// dalla sigla «API80,» / «API880,». Il canale è analogico (FM larga o stretta,
// AM) o digitale D-STAR (DV); dei canali DV si conservano i nominativi (UR,
// RPT1, RPT2) e il codice digitale, un canale DV nuovo nasce con «CQCQCQ».
//
// Frequenza e scostamento: 18 e 16 bit per un moltiplicatore (5, 6,25, 6,25,
// 8,333 o 9 kHz) scritto nei tre bit alti della frequenza (ricezione) e nei tre
// seguenti (scostamento).
//
// Differenze volute rispetto a CHIRP:
//   - una frequenza che non è multiplo di 5, 6,25 o 9 kHz viene rifiutata
//     (CHIRP si ferma con un errore interno);
//   - il salto «P» della scansione programmata resta com'è quando il canale
//     modificato resta escluso dalla scansione;
//   - i modi «AM», «AM stretta» e (ID-80H) «FM larga da radiodiffusione» di un
//     canale restano quelli che sono se la frequenza non cambia;
//   - CHIRP, quando scrive un canale, accende o spegne anche un «indicatore del
//     nome» (name_flags) all'indirizzo 0x0FB8, che però cade DENTRO i canali
//     98-101: è un errore di CHIRP (nelle memorie lette da radio vere lì ci sono
//     canali, non indicatori) e scriverlo rovinerebbe quei canali. Qui non si
//     scrive: la posizione vera dell'indicatore, se esiste, non è nota.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

type id880Fam struct{}

type id880Opt struct {
	clone *icomOpt
	tag   string // sigla del modello dentro la memoria
	wfm   bool   // ID-80H: riceve anche la FM da radiodiffusione
}

const id880Format = `
struct {
  u24 rxmult:3,
      txmult:3,
      freq:18;
  u16 offset;
  u16 rtone:6,
      ctone:6,
      unknown1:1,
      mode:3;
  u8 dtcs;
  u8 tune_step:4,
     unknown2:4;
  u8 unknown3;
  u8 unknown4:1,
     tmode:3,
     duplex:2,
     dtcs_polarity:2;
  char name[8];
  u8 unknwon5:1,
     digital_code:7;
  u8 urcall[7];
  u8 r1call[7];
  u8 r2call[7];
} memory[1000];
#seekto 0xAA80;
u8 used_flags[132];
#seekto 0xAB04;
u8 skip_flags[132];
u8 pskip_flags[132];
#seekto 0xAD00;
struct {
  u8 bank;
  u8 index;
} bank_info[1000];
#seekto 0xB550;
struct {
  char name[6];
} bank_names[26];
#seekto 0xDE56;
struct {
  char call[8];
  char extension[4];
} mycall[6];
struct {
  char call[8];
} urcall[60];
struct {
  char call[8];
  char extension[4];
} rptcall[99];
`

// id880Mult: moltiplicatori della frequenza (FREQ_MULTIPLIER di CHIRP)
var id880Mult = []int64{5000, 6250, 6250, 8333, 9000}

func id880Clone(m1 byte) *icomOpt {
	return &icomOpt{model: [4]byte{0x31, m1, 0x00, 0x01}, end: "Icom Inc.B1", size: 0xF600, fast: true,
		ranges: [][3]int{{0x0000, 0xF5C0, 32}, {0xF5C0, 0xF5E0, 16}, {0xF5E0, 0xF600, 32}}}
}

func init() {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ !\"#$%&'()*+,-./:;<=>?@[\\]^"
	mk := func(id, model, rclass string, o *id880Opt, watts float64, bands []band, what string) *genModel {
		return &genModel{ID: id, Vendor: "Icom", Model: model, RClass: rclass, Name: "Icom " + model,
			Slots: 1000, First: 0, NameMax: 8, Charset: chars, Powers: []genPower{{"Unica", watts, 0}}, Bands: bands,
			Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: id880Fam{}, Opt: o, Digital: "D-STAR",
			Notes: what + " 1000 canali (da 0 a 999) con nomi di 8 caratteri, analogici o D-STAR. " + id880Cable +
				"Un canale D-STAR nuovo nasce con «CQCQCQ». Driver di CHIRP."}
	}
	genRegister(
		mk("ic-id880", "ID-880H", "ID880Radio", &id880Opt{clone: id880Clone(0x67), tag: "API880,"}, 50,
			[]band{{118000000, 173995000}, {230000000, 549995000}, {810000000, 823990000}, {849000000, 868990000}, {894000000, 999990000}},
			"Radio veicolare VHF/UHF."),
		mk("ic-id80", "ID-80H", "ID80Radio", &id880Opt{clone: id880Clone(0x55), tag: "API80,", wfm: true}, 5,
			[]band{{495000, 1000000000}}, "Palmare VHF/UHF."),
	)
}

const id880Cable = "Cavo di copia Icom (quello indicato nel manuale della radio); la radio va accesa normalmente. Copia veloce a 38400 baud. "

func (id880Fam) opt(m *genModel) *id880Opt      { return m.Opt.(*id880Opt) }
func (f id880Fam) icom(m *genModel) *icomOpt    { return f.opt(m).clone }
func (f id880Fam) mem(m *genModel) *bwEl        { return bwMust(id880Format) }
func (f id880Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f id880Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f id880Fam) bank(m *genModel, i int) *bwEl {
	return f.mem(m).F("bank_info").I(i)
}

func (f id880Fam) Rec(m *genModel, img []byte, i int) []byte {
	e := f.mem(m)
	out := append([]byte{}, f.rec(m, i).Raw(img)...)
	out = append(out, icomB(icomBit(e.F("skip_flags"), img, i)), icomB(icomBit(e.F("pskip_flags"), img, i)))
	return append(out, f.bank(m, i).Raw(img)...)
}

func (f id880Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	e := f.mem(m)
	n := f.rec(m, i).ByteLen()
	if len(raw) != n+4 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:n])
	icomSetBit(e.F("skip_flags"), img, i, raw[n] != 0)
	icomSetBit(e.F("pskip_flags"), img, i, raw[n+1] != 0)
	f.bank(m, i).SetRaw(img, raw[n+2:])
	icomSetBit(e.F("used_flags"), img, i, false)
}

// Match: ID-80H e ID-880H hanno la memoria della stessa grandezza e si
// distinguono dalla sigla scritta dalla radio (come CHIRP)
func (f id880Fam) Match(m *genModel, img []byte) bool {
	o := f.opt(m)
	return len(img) == o.clone.size && bytes.Contains(img, []byte(o.tag))
}

func (id880Fam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }
func (id880Fam) fitBySet()                {}

func (id880Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 70000000 || d < -70000000
}

// id880Freq: frequenza e scostamento del canale in Hz (0 se il moltiplicatore
// non esiste)
func (f id880Fam) freq(img []byte, r *bwEl) (rx, offs int64) {
	rm, tm := r.F("rxmult").Int(img), r.F("txmult").Int(img)
	if int(rm) >= len(id880Mult) || int(tm) >= len(id880Mult) {
		return 0, 0
	}
	return r.F("freq").Int(img) * id880Mult[rm], r.F("offset").Int(img) * id880Mult[tm]
}

// id880Enc: valore e moltiplicatore con cui si scrive una frequenza
func id880Enc(hz int64) (val, mult int64, ok bool) {
	for k, step := range id880Mult {
		if k != 2 && k != 3 && hz%step == 0 {
			return hz / step, int64(k), true
		}
	}
	return 0, 0, false
}

func (f id880Fam) Get(m *genModel, img []byte, i int) *Chan {
	e := f.mem(m)
	if icomBit(e.F("used_flags"), img, i) {
		return nil
	}
	r := f.rec(m, i)
	rx, offs := f.freq(img, r)
	if rx <= 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	mode := r.F("mode").Int(img)
	c.Dig = mode == 5
	c.Narrow = mode == 1 || mode == 4
	if !c.Dig {
		c.TxTone, c.RxTone = id31Tones(r.F("tmode").Int(img), r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img), r.F("dtcs_polarity").Int(img))
	}
	c.Scan = !icomBit(e.F("skip_flags"), img, i) && !icomBit(e.F("pskip_flags"), img, i)
	c.Name = strings.TrimRight(r.F("name").Str(img), " ")
	return c
}

func (f id880Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	e, r, o := f.mem(m), f.rec(m, i), f.opt(m)
	if c == nil {
		icomSetBit(e.F("used_flags"), img, i, true)
		r.Fill(img, 0xFF)
		f.bank(m, i).F("bank").SetInt(img, 0xFF)
		return nil
	}
	was := !icomBit(e.F("used_flags"), img, i)
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz)
	tmode, rtone, ctone, dtcs, pol := int64(0), int64(8), int64(8), int64(0), int64(0)
	oldRx, oldOffs, oldMode := int64(0), int64(600000), int64(0)
	if was {
		rtone, ctone, dtcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs").Int(img)
		oldRx, oldOffs = f.freq(img, r)
		oldMode = r.F("mode").Int(img)
	}
	if !c.Dig {
		var err error
		if tmode, rtone, ctone, dtcs, pol, err = icomTonesFor(c, rtone, ctone, dtcs); err != nil {
			return err
		}
		tmode = []int64{0, 1, 2, 5}[tmode]
	}
	fv, fm, ok := id880Enc(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5, di 6,25 o di 9 kHz)", fmtMHz(c.Rx))
	}
	if fv > 0x3FFFF {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	// lo scostamento di un canale simplex resta quello che c'era (come in CHIRP)
	dup, offs := int64(0), oldOffs
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	ov, om, ok := id880Enc(offs)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz (multipli di 5, di 6,25 o di 9 kHz)", offs)
	}
	if ov > 0xFFFF || dup != 0 && offs >= 60000000 {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	mode := int64(0)
	same := was && oldRx == c.Rx
	switch {
	case c.Dig:
		mode = 5
	case same && (oldMode == 3 && !c.Narrow || oldMode == 4 && c.Narrow || oldMode == 2 && o.wfm && !c.Narrow):
		mode = oldMode // AM, AM stretta, FM da radiodiffusione: restano
	case c.Rx >= 118000000 && c.Rx < 137000000:
		mode = 3 // banda aerea: AM
		if c.Narrow {
			mode = 4
		}
	case c.Narrow:
		mode = 1
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !was {
		r.Fill(img, 0)
	}
	if !(was && oldMode == 5) && c.Dig {
		// canale D-STAR nuovo: chiamata generale, nessun ripetitore (come CHIRP)
		r.F("urcall").SetRaw(img, id31Call("CQCQCQ"))
		r.F("r1call").SetRaw(img, id31Call(""))
		r.F("r2call").SetRaw(img, id31Call(""))
	}
	r.F("freq").SetInt(img, fv)
	r.F("rxmult").SetInt(img, fm)
	r.F("txmult").SetInt(img, om)
	r.F("offset").SetInt(img, ov)
	r.F("name").SetStr(img, m.cleanName(c.Name), ' ')
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs").SetInt(img, dtcs)
	r.F("dtcs_polarity").SetInt(img, pol)
	r.F("tmode").SetInt(img, tmode)
	r.F("duplex").SetInt(img, dup)
	r.F("mode").SetInt(img, mode)
	if r.F("tune_step").Int(img) > 13 {
		r.F("tune_step").SetInt(img, 0) // passo di sintonia sconosciuto: 5 kHz (come CHIRP)
	}
	icomSetBit(e.F("used_flags"), img, i, false)
	wasP := was && icomBit(e.F("pskip_flags"), img, i) && !icomBit(e.F("skip_flags"), img, i)
	switch {
	case c.Scan:
		icomSetBit(e.F("skip_flags"), img, i, false)
		icomSetBit(e.F("pskip_flags"), img, i, false)
	case wasP:
		// il salto «P» resta
	default:
		icomSetBit(e.F("skip_flags"), img, i, true)
		icomSetBit(e.F("pskip_flags"), img, i, false)
	}
	return nil
}

func (f id880Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f id880Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
