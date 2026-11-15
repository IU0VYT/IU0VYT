// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icw32.py (4.39).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.39: palmari bibanda Icom IC-W32A e IC-W32E. La radio ha due metà separate,
// VHF e UHF, ognuna con 100 canali (da 0 a 99) con nomi di 8 caratteri: in
// CHIRP sono due «sotto-radio», qui una lista sola di 200 posizioni, da 0 a 99
// la metà VHF e da 100 a 199 la metà UHF (100 + il numero che si vede sulla
// radio). In ogni metà la memoria ha 110 posizioni: le ultime dieci sono i
// limiti di scansione (M1A … M5B), che qui non si toccano.
//
// I due modelli hanno la stessa memoria; i file della IC-W32E hanno un byte in
// più in coda, a zero, che la radio non manda (è il modo in cui CHIRP li
// distingue).
//
// Il confronto della struttura con CHIRP (TestChirpLayout) qui non si fa (vedi
// chirpNoLayout nelle prove): in CHIRP gli indirizzi sono quelli delle due
// sotto-radio. Lettura, modifica e copia sono confrontate come per le altre radio.
//
// Differenze volute rispetto a CHIRP:
//   - una frequenza o uno scostamento che la radio non può memorizzare vengono
//     rifiutati (CHIRP li tronca);
//   - la banda aerea (da 118 a 137 MHz, solo nella metà VHF) è in AM; l'AM di
//     un canale resta se la frequenza non cambia.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

type icw32Fam struct{}

const icw32Format = `
struct w32rec {
  bbcd freq[3];
  bbcd offset[3];
  u8 ctone;
  u8 rtone;
  char name[8];
};
struct w32flag {
  u8 empty:1,
     skip:1,
     tmode:2,
     duplex:2,
     unk3:1,
     am:1;
};
#seekto 0x0000;
struct w32rec vhf[110];
#seekto 0x06E0;
struct w32rec uhf[110];
#seekto 0x0DC0;
struct w32flag vflag[110];
#seekto 0x0E2E;
struct w32flag uflag[110];
#seekto 0x0E9C;
struct {
  u8 unknown1:7,
     right_scan_direction:1;
  u8 right_scanning:1,
     unknown2:7;
  u8 unknown3:7,
     left_scan_direction:1;
  u8 left_scanning:1,
     unknown4:7;
} state;
`

const icw32Tag = "IcomCloneFormat3"

// icw32Bands: limiti delle due metà (VHF, UHF)
var icw32Bands = []band{{118000000, 174000000}, {400000000, 470000000}}

func init() {
	mk := func(id, model, rclass string, m3 byte, extra int) *genModel {
		return &genModel{ID: id, Vendor: "Icom", Model: model, RClass: rclass, Name: "Icom " + model,
			Slots: 200, First: 0, NameMax: 8, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-./", Powers: []genPower{{"Unica", 5, 0}},
			Bands: icw32Bands, HasScan: true, NoTxOff: true, Baud: 9600, Fam: icw32Fam{},
			Opt: &icomOpt{model: [4]byte{0x18, 0x82, 0x00, m3}, end: "Icom Inc.", size: 0x0FE0 + extra, extra: extra, ranges: [][3]int{{0x0000, 0x0FE0, 16}}},
			Notes: "200 canali con nomi di 8 caratteri: da 0 a 99 la metà VHF, da 100 a 199 la metà UHF (100 + il numero che si vede sulla radio). " +
				icomCableNote + icomNoDCS + icomNoPower + "Driver di CHIRP."}
	}
	genRegister(mk("ic-w32a", "IC-W32A", "ICW32ARadio", 0x01, 0), mk("ic-w32e", "IC-W32E", "ICW32ERadio", 0x02, 1))
}

func (icw32Fam) icom(m *genModel) *icomOpt { return m.Opt.(*icomOpt) }
func (icw32Fam) mem() *bwEl                { return bwMust(icw32Format) }

// layout: serve anche per la prova di scrittura dei canali (genTrySet)
func (f icw32Fam) layout(m *genModel) *bwEl { return f.mem() }

// rec, flg: canale e indicatori della posizione i (da 0 a 99 VHF, da 100 a 199 UHF)
func (f icw32Fam) rec(i int) *bwEl {
	if i >= 100 {
		return f.mem().F("uhf").I(i - 100)
	}
	return f.mem().F("vhf").I(i)
}

func (f icw32Fam) flg(i int) *bwEl {
	if i >= 100 {
		return f.mem().F("uflag").I(i - 100)
	}
	return f.mem().F("vflag").I(i)
}

func (f icw32Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.flg(i).Raw(img)...)
}

func (f icw32Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(i).ByteLen()
	if len(raw) != n+1 {
		return
	}
	// un canale non passa da una metà all'altra se la frequenza non è della metà di arrivo
	old := append(append([]byte{}, f.rec(i).Raw(img)...), f.flg(i).Raw(img)...)
	f.rec(i).SetRaw(img, raw[:n])
	f.flg(i).SetRaw(img, raw[n:])
	f.flg(i).F("empty").SetBool(img, false)
	if c := f.Get(m, img, i); c == nil || !f.inBand(i, c.Rx) {
		f.rec(i).SetRaw(img, old[:n])
		f.flg(i).SetRaw(img, old[n:])
	}
}

// Match: IC-W32A e IC-W32E si distinguono dalla lunghezza del file (come CHIRP)
func (f icw32Fam) Match(m *genModel, img []byte) bool {
	o := f.icom(m)
	if len(img) != o.size {
		return false
	}
	end := len(img) - o.extra
	return string(img[end-len(icw32Tag):end]) == icw32Tag && bytes.Equal(img[end:], make([]byte, o.extra))
}

func (icw32Fam) toneOK(t, r genTone) bool { return ict7Fam{}.toneOK(t, r) }
func (icw32Fam) fitBySet()                {}

func (icw32Fam) chirpInvalid(m *genModel, c *Chan) bool { return ict7Fam{}.chirpInvalid(m, c) }

// trySlot: posizione su cui provare un canale (vedi genTrySet): una della metà giusta
func (icw32Fam) trySlot(m *genModel, c *Chan) int {
	if c.Rx >= icw32Bands[1].Lo {
		return 100
	}
	return 0
}

func (icw32Fam) inBand(i int, hz int64) bool {
	b := icw32Bands[i/100]
	return hz >= b.Lo && hz <= b.Hi
}

// Locked: ogni metà della radio accetta solo le frequenze della sua banda
func (f icw32Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if want == nil || f.inBand(i, want.Rx) {
		return nil
	}
	if i < 100 {
		return fmt.Errorf("le posizioni da 0 a 99 sono la metà VHF della radio (da 118 a 174 MHz): %s MHz va nelle posizioni da 100 a 199", fmtMHz(want.Rx))
	}
	return fmt.Errorf("le posizioni da 100 a 199 sono la metà UHF della radio (da 400 a 470 MHz): %s MHz va nelle posizioni da 0 a 99", fmtMHz(want.Rx))
}

func (f icw32Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, fl := f.rec(i), f.flg(i)
	if fl.F("empty").Bool(img) {
		return nil
	}
	c := &Chan{Rx: yaesuRound(r.F("freq").Int(img) * 1000)}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * 100
	switch fl.F("duplex").Int(img) {
	case 2:
		c.Tx = c.Rx - offs
	case 3:
		c.Tx = c.Rx + offs
	}
	mode := int64(0)
	switch fl.F("tmode").Int(img) {
	case 2:
		mode = 1
	case 3:
		mode = 2
	}
	c.TxTone, c.RxTone = icomTones(mode, r.F("rtone").Int(img), r.F("ctone").Int(img), 0, 0)
	c.Scan = !fl.F("skip").Bool(img)
	if nm := r.F("name"); nm.Raw(img)[0] != 0xFF {
		c.Name = strings.TrimRight(nm.Str(img), " ")
	}
	return c
}

func (f icw32Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(i), f.flg(i)
	if c == nil {
		fl.F("empty").SetBool(img, true) // come CHIRP: il resto resta com'è
		return nil
	}
	if err := f.Locked(m, i, nil, c); err != nil {
		return err
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, scostamento 600 kHz)
	was := !fl.F("empty").Bool(img)
	kept, rtone, ctone, oldRx := int64(600000), int64(8), int64(8), int64(0)
	if was {
		kept, rtone, ctone = r.F("offset").Int(img)*100, r.F("rtone").Int(img), r.F("ctone").Int(img)
		oldRx = yaesuRound(r.F("freq").Int(img) * 1000)
	}
	tmode, rtone, ctone, _, _, err := icomTonesFor(c, rtone, ctone, 0)
	if err != nil {
		return err
	}
	if tmode == 3 {
		return fmt.Errorf("la radio non ha i codici DCS")
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// lo scostamento di un canale simplex resta quello che c'era (come in CHIRP):
	// in quel caso il campo non si riscrive
	dup, offs := int64(0), kept
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 3, d
	case d < 0:
		dup, offs = 2, -d
	}
	if dup != 0 && offs%100 != 0 {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz", offs)
	}
	if dup != 0 && offs/100 > 999999 {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	am := i < 100 && c.Rx < 137000000 || was && oldRx == c.Rx && fl.F("am").Bool(img)
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	r.F("freq").SetInt(img, fk)
	if dup != 0 || !was {
		r.F("offset").SetInt(img, offs/100)
	}
	if name := strings.TrimRight(m.cleanName(c.Name), " "); name != "" {
		r.F("name").SetStr(img, name, ' ')
	} else {
		r.F("name").Fill(img, 0xFF)
	}
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	fl.F("duplex").SetInt(img, dup)
	fl.F("tmode").SetInt(img, []int64{0, 2, 3}[tmode])
	fl.F("skip").SetBool(img, !c.Scan)
	fl.F("am").SetBool(img, am)
	fl.F("empty").SetBool(img, false)
	// come CHIRP: una scansione rimasta in corso nella memoria viene fermata
	st := f.mem().F("state")
	st.F("left_scanning").SetBool(img, false)
	st.F("right_scanning").SetBool(img, false)
	return nil
}

func (f icw32Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := icomDownload(ctx, port, m.Baud, f.icom(m), prog)
	return img, "", err
}

func (f icw32Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", icomUpload(ctx, port, m.Baud, f.icom(m), img, prog)
}
