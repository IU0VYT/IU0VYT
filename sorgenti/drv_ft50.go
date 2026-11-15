// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ft50.py (4.37).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.37: Yaesu FT-50R: 100 canali con nomi di 4 caratteri, 9600 baud. La copia
// parte dalla radio ed è in otto blocchi (10, 16, 112, 16, 16, 1776, 1776 e 1
// byte), tutti confermati tranne l'ultimo. In scrittura i byte partono uno
// alla volta e la radio li rimanda indietro (come fa CHIRP con questa radio:
// la scrittura dura un minuto e mezzo). La radio non dichiara il modello: la
// memoria si riconosce dalla lunghezza, come in CHIRP. Niente banda stretta e
// niente trasmissione su una frequenza separata. Il modo (FM, AM, FM larga)
// non è una colonna del programma: resta quello del canale se la frequenza non
// cambia, altrimenti segue la frequenza.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare viene rifiutata invece di
//     essere troncata al kHz; lo stesso per lo scostamento;
//   - uno scostamento oltre i 70 MHz viene rifiutato (CHIRP si ferma con un
//     errore interno).

import (
	"context"
	"fmt"
	"strings"
)

type ft50Fam struct{}

const ft50Format = `
struct flag_struct {
  u8 unknown1f:5,
     skip:1,
     mask:1,
     used:1;
};

struct mem_struct {
  u8 showname:1,
     unknown1:3,
     unknown2:2,
     unknown3:2;
  u8 ishighpower:1,
     power:2,
     unknown4:1,
     tuning_step:4;
  u8 codememno:4,
     codeorpage:2,
     duplex:2;
  u8 tmode:2,
     tone:6;
  u8 unknown5:1,
     dtcs:7;
  u8 unknown6:6,
     mode:2;
  bbcd freq[3];
  bbcd offset[3];
  u8 name[4];
};

#seekto 0x000C;
struct {
  u8 extendedrx_flg;    // Seems to be set to 03 when extended rx is enabled
  u8 extendedrx;        // Seems to be set to 01 when extended rx is enabled
} extendedrx_struct;    // UNFINISHED!!

#seekto 0x001A;
struct flag_struct flag[100];

#seekto 0x079C;
struct flag_struct flag_repeat[100];

#seekto 0x00AA;
struct mem_struct memory[100];
struct mem_struct special[11];

#seekto 0x08C7;
struct {
  u8 sub_display;
  u8 unknown1s;
  u8 apo;
  u8 timeout;
  u8 lock;
  u8 rxsave;
  u8 lamp;
  u8 bell;
  u8 cwid[16];
  u8 unknown2s;
  u8 artsmode;
  u8 artsbeep;
  u8 unknown3s;
  u8 unknown4s;
  struct {
    u8 header[3];
    u8 mem_num;
    u8 digits[16];
  } autodial[8];
  struct {
    u8 header[3];
    u8 mem_num;
    u8 digits[32];
  } autodial9_ro;
  bbcd pagingcodec_ro[2];
  bbcd pagingcodep[2];
  struct {
    bbcd digits[2];
  } pagingcode[6];
  u8 code_dec_c_en:1,
     code_dec_p_en:1,
     code_dec_1_en:1,
     code_dec_2_en:1,
     code_dec_3_en:1,
     code_dec_4_en:1,
     code_dec_5_en:1,
     code_dec_6_en:1;
  u8 pagingspeed;
  u8 pagingdelay;
  u8 pagingbell;
  u8 paginganswer;

  #seekto 0x0E30;
  u8 squelch;       // squelch
  u8 unknown0c;
  u8 rptl:1,        // repeater input tracking
     amod:1,        // auto mode
     scnl:1,        // scan lamp
     resm:1,        // scan resume mode 0=5sec, 1=carr
     ars:1,         // automatic repeater shift
     keybeep:1,     // keypad beep
     lck:1,         // lock
     unknown1c:1;
  u8 lgt:1,
     pageamsg:1,
     unknown2c:1,
     bclo:1,        // Busy channel lock out
     unknown3c:2,
     cwid_en:1,     // CWID off/on
     tsav:1;        // TX save
  u8 unknown4c:4,
     artssped:1,    // ARTS/SPED: 0=15s, 1=25s
     unknown5c:1,
     rvhm:1,        // RVHM: 0=home, 1=rev
     mon:1;         // MON: 0=mon, 1=tcal
} settings;

#seekto 0x080E;
struct mem_struct vfo_mem[10];


`

var ft50Clone = &yaesuOpt{model: "", size: 3723, head: 10, blocks: []int{10, 16, 112, 16, 16, 1776, 1776, 1}, chunk: 8,
	sums: [][3]int{{0x0000, 0x0E89, 0x0E8A}}, send: "[PTT]", recv: "[MONI] («WAIT»)", byteEcho: true}

// caratteri della radio nell'ordine dei codici (le cifre compaiono due volte:
// si scrive il primo codice, come CHIRP)
const ft50Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/\x00\x00\x00|0123456789"

func init() {
	genRegister(
		&genModel{ID: "ys-ft50", Vendor: "Yaesu", Model: "FT-50", RClass: "FT50Radio", Name: "Yaesu FT-50R",
			Slots: 100, First: 1, NameMax: 4, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/|",
			Powers:  []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 1 W", 1, 2}, {"Minima 0,1 W", 0.1, 3}},
			Bands:   []band{{76000000, 200000000}, {300000000, 540000000}, {590000000, 999000000}},
			HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ft50Fam{}, Opt: ft50Clone,
			Notes: yaesuNote(100, 4, 9600) + yaesuCloneHow("MIC/SP", "[PTT] + [DIAL]", "[PTT]", "[MONI]") + "Driver di CHIRP."},
	)
}

func (ft50Fam) yaesu(m *genModel) *yaesuOpt { return m.Opt.(*yaesuOpt) }
func (ft50Fam) mem() *bwEl                  { return bwMust(ft50Format) }
func (f ft50Fam) layout(m *genModel) *bwEl  { return f.mem() }
func (f ft50Fam) rec(i int) *bwEl           { return f.mem().F("memory").I(i) }
func (f ft50Fam) flag(i int) *bwEl          { return f.mem().F("flag").I(i) }
func (f ft50Fam) flag2(i int) *bwEl         { return f.mem().F("flag_repeat").I(i) }

func (f ft50Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.flag(i).Raw(img)...)
}

func (f ft50Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(i).ByteLen()
	if len(raw) != n+1 {
		return
	}
	f.rec(i).SetRaw(img, raw[:n])
	f.flag(i).SetRaw(img, raw[n:])
	f.flag(i).F("used").SetBool(img, true)
	f.flag(i).F("mask").SetBool(img, true)
	f.copyFlag(img, i)
}

// copyFlag: la radio tiene due copie dei segni del canale
func (f ft50Fam) copyFlag(img []byte, i int) {
	for _, k := range []string{"skip", "mask", "used"} {
		f.flag2(i).F(k).SetInt(img, f.flag(i).F(k).Int(img))
	}
}

func (f ft50Fam) Match(m *genModel, img []byte) bool { return yaesuMatch(f.yaesu(m), img) }

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (ft50Fam) noIdent() {}

func (ft50Fam) toneOK(t, r genTone) bool { return vx5ToneOK(t, r) }
func (ft50Fam) fitBySet()                {}

func (ft50Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	if d < 0 {
		d = -d
	}
	return !vx5ToneOK(t, r) || d > 70000000
}

func (f ft50Fam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.flag(i).F("used").Bool(img) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: yaesuRound(r.F("freq").Int(img) * 1000)}
	if c.Rx <= 0 {
		return nil
	}
	off := yaesuRound(r.F("offset").Int(img) * 1000)
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - off
	case 2:
		c.Tx = c.Rx + off
	default:
		c.Tx = c.Rx
	}
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img)); k < len(chirpOldTones) {
		tone = fmtTone(chirpOldTones[k])
	}
	if k := int(r.F("dtcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	switch r.F("tmode").Int(img) {
	case 1:
		c.TxTone = tone
	case 2:
		c.TxTone, c.RxTone = tone, tone
	case 3:
		c.TxTone, c.RxTone = dcs, dcs
	}
	if r.F("ishighpower").Bool(img) {
		c.Power = m.powerSem(0)
	} else {
		c.Power = m.powerSem(int(3 - r.F("power").Int(img)))
	}
	c.Scan = !f.flag(i).F("skip").Bool(img)
	var sb strings.Builder
	for _, b := range r.F("name").Raw(img) {
		if k := int(b & 0x7F); k < len(ft50Chars) && ft50Chars[k] != 0 {
			sb.WriteByte(ft50Chars[k])
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

func (f ft50Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(i), f.flag(i)
	used := fl.F("used").Bool(img)
	if c == nil {
		// come CHIRP: si toglie solo il segno «in uso» (la copia dei segni resta com'è)
		fl.F("used").SetBool(img, false)
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := vx5Tmode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	kept, tone, dcs, step, oldRx, oldMode := int64(600), int64(8), int64(0), int64(0), int64(0), int64(0)
	if used {
		kept = r.F("offset").Int(img)
		if k := r.F("tone").Int(img); int(k) < len(chirpOldTones) {
			tone = k
		}
		if k := r.F("dtcs").Int(img); int(k) < len(dtcsCodes) {
			dcs = k
		}
		if k := r.F("tuning_step").Int(img); k < 7 {
			step = k
		}
		oldRx, oldMode = yaesuRound(r.F("freq").Int(img)*1000), r.F("mode").Int(img)
	}
	if t.Mode == "Tone" {
		if tone = yaesuOldTone(t.Hz10); tone < 0 {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	if t.Mode == "DTCS" {
		dcs = -1
		for k, x := range dtcsCodes {
			if x == t.Code {
				dcs = int64(k)
			}
		}
		if dcs < 0 {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	dup, offs := int64(0), kept
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	if d != 0 {
		ok2 := false
		if offs, ok2 = yaesuKHz(ad); !ok2 || ad > 70000000 {
			return fmt.Errorf("la radio non può memorizzare questo scostamento fra ricezione e trasmissione (%s MHz): non ha la trasmissione su una frequenza separata", fmtMHz(ad))
		}
		dup = 1
		if d > 0 {
			dup = 2
		}
	}
	// modo: resta quello del canale se la frequenza non cambia, altrimenti segue la frequenza
	mode := []int64{0, 2, 1}[icx90Mode(c.Rx)] // 0 FM, 1 AM, 2 FM larga
	if used && oldRx == c.Rx && (oldMode == 1 || oldMode == 2) {
		mode = oldMode
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	name := m.cleanName(c.Name)
	if len(name) == 0 {
		r.F("name").SetRaw(img, []byte{0x24, 0x24, 0x24, 0x24})
		r.F("showname").SetInt(img, 0)
	} else {
		r.F("showname").SetInt(img, 1)
		enc := make([]byte, 4)
		for k := range enc {
			ch := byte(' ')
			if k < len(name) {
				ch = name[k]
			}
			p := strings.IndexByte(ft50Chars, ch)
			if p < 0 || ch == 0 {
				p = strings.IndexByte(ft50Chars, ' ')
			}
			enc[k] = byte(p)
		}
		r.F("name").SetRaw(img, enc)
	}
	r.F("freq").SetInt(img, fk)
	r.F("duplex").SetInt(img, dup)
	r.F("offset").SetInt(img, offs)
	r.F("mode").SetInt(img, mode)
	r.F("tuning_step").SetInt(img, step)
	if pi := int64(m.powerIndex(c.Power)); pi == 0 {
		r.F("ishighpower").SetInt(img, 1)
		if r.F("power").Int(img) == 3 {
			r.F("power").SetInt(img, 2)
		}
	} else {
		r.F("ishighpower").SetInt(img, 0)
		r.F("power").SetInt(img, 3-pi)
	}
	r.F("tmode").SetInt(img, tmode)
	r.F("tone").SetInt(img, tone)
	r.F("dtcs").SetInt(img, dcs)
	fl.F("skip").SetBool(img, !c.Scan)
	if !used {
		fl.F("used").SetBool(img, true)
		fl.F("mask").SetBool(img, true) // visibile sulla radio
		for _, k := range []string{"unknown1", "unknown2", "unknown3", "unknown4", "unknown5", "unknown6", "codeorpage"} {
			r.F(k).SetInt(img, 0)
		}
		r.F("codememno").SetInt(img, 2)
	}
	f.copyFlag(img, i)
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

func (f ft50Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := yaesuDownload(ctx, m, port, f.yaesu(m), prog)
	return img, "", err
}

func (f ft50Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", yaesuUpload(ctx, m, port, f.yaesu(m), img, prog)
}
