// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// vx150.py (4.37).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.37: Yaesu / Vertex Standard VX-150: 199 canali VHF con nomi di 7 caratteri,
// 9600 baud. La copia parte dalla radio con lo stesso scambio della FT-7800
// (8 byte di intestazione, blocchi da 32 byte confermati, somma di controllo
// finale; il cavo deve rimandare indietro i dati), ma in scrittura la radio
// non conferma la somma di controllo. La potenza non è per canale (si sceglie
// nelle impostazioni della radio). Un canale è vuoto quando i suoi 16 byte
// sono quelli di fabbrica.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che non è multipla di 2,5 kHz, o uno scostamento che non
//     è multiplo di 10 kHz, vengono rifiutati invece di essere troncati;
//   - un subtono che la radio non ha viene rifiutato (CHIRP scrive 67,0 Hz).

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

type vx150Fam struct{}

const vx150Format = `
#seekto 0x0013;
u8 step;

#seekto 0x0026;
struct {
    u8 mem_only:1,
       unknown1:5,
       power:2;
} settings3;

#seekto 0x0049;
struct {
    u8 unknown1:2,
       tx_led:1,
       unknown2:1,
       scan_lamp:1,
       ars:1,
       scan_resume:2;
    u8 bclo:1,
       edge_beep:1,
       key_beep:1,
       arts_interval:1,
       tx_save:1,
       rx_save:3;
    u8 unknown4a:1,
       dtmf_speed:1,
       smt_mod:1,
       ani_on:1,
       unknown5:1,
       apo:3;
    u8 unknown5b:6,
       arts_bp:2;
    u8 unknown6:3,
       rev_hm:1,
       moni_tcall:1,
       bell:3;
    u8 unknown7:5,
       lamp:3;
    u8 unknown8:5,
       tot:3;
    u8 lock_mode;
    u8 unknown9[3];
    u8 dtmf_delay;
} settings2;

#seekto 0x0056;
struct {
    u8 p1_key;
    u8 p2_key;
} p_keys;

struct channel {
    u8 active:1,
       skip:1,
       unknown1:2,
       duplex:2,
       tmode:2;
    u8 sub_khz:2,
       unknown_flags:5,
       band:1;
    bbcd freq_khz;
    bbcd freq_mhz;
    u8 unknown2:2,
       narrow:1,
       clock_shift:1,
       unknown3:4;
    u8 unknown4;
    bbcd offset_low;
    bbcd offset_high;
    u8 tone;
    u8 dtcs;
    u8 unknown5[6];
};

struct channel home;

u8 cwid[16];
u8 ani[16];
struct {
    u8 digits[16];
} dtmf[9];

#seekto 0x0150;
struct {
    u8 name[7];
    u8 show_name;
} names[199];

struct {
    u8 name[7];
    u8 show_name;
} limit_names[10];

struct channel memory[199];

struct channel limits[10];
`

// vx150Empty: i 16 byte di un canale vuoto
var vx150Empty = []byte{0x00, 0x01, 0x00, 0x44, 0x00, 0x00, 0x60, 0x00, 0x0c, 0x00, 0x00, 0x00, 0xaa, 0x00, 0x02, 0x00}

var vx150Clone = &ftxOpt{id: "\x0a\x01\x04\x00\x01\x24", size: 5353, blk: 32, noSumAck: true,
	send: "[PTT]", recv: "[MONI] («SAVING»)", name: "VX-150"}

// caratteri della radio nell'ordine dei codici (tre lettere greche non si
// scrivono da qui)
var vx150Chars = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/ΔγΣ|")

func init() {
	genRegister(
		&genModel{ID: "ys-vx150", Vendor: "Yaesu", Model: "VX-150", RClass: "VX150Radio", Name: "Yaesu VX-150",
			Slots: 199, First: 1, NameMax: 7, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/|",
			Powers: []genPower{{"Unica", 5, 0}}, Bands: []band{{136000000, 174000000}},
			Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: vx150Fam{}, Opt: vx150Clone,
			Notes: yaesuNote(199, 7, 9600) + "La copia parte dalla radio: spegnila, collega il cavo alla presa MIC/SP e accendila tenendo premuti [PTT] e [LAMP]; " +
				"con la manopola scegli «CLONE» e premi [F]. Per leggere: avvia la lettura qui e poi tieni premuto [PTT] per un momento (compare «SENDING»). " +
				"Per scrivere: premi [MONI] sulla radio (compare «SAVING») e poi avvia la scrittura qui. La potenza si sceglie sulla radio, non per canale. Driver di CHIRP."},
	)
}

func (vx150Fam) mem() *bwEl                 { return bwMust(vx150Format) }
func (f vx150Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f vx150Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f vx150Fam) nam(i int) *bwEl          { return f.mem().F("names").I(i) }

func (f vx150Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.nam(i).Raw(img)...)
}

func (f vx150Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(i).ByteLen()
	if len(raw) != n+f.nam(i).ByteLen() || bytes.Equal(raw[:n], vx150Empty) {
		return
	}
	f.rec(i).SetRaw(img, raw[:n])
	f.nam(i).SetRaw(img, raw[n:])
}

func (vx150Fam) Match(m *genModel, img []byte) bool {
	o := m.Opt.(*ftxOpt)
	return len(img) == o.size && bytes.HasPrefix(img, []byte(o.id))
}

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (vx150Fam) noIdent() {}

func (vx150Fam) toneOK(t, r genTone) bool { return vx5ToneOK(t, r) }
func (vx150Fam) fitBySet()                {}

// powerFixed: la potenza non è per canale
func (vx150Fam) powerFixed(m *genModel) bool { return true }

func (vx150Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	if d < 0 {
		d = -d
	}
	return !vx5ToneOK(t, r) || c.Rx%2500 != 0 || d%10000 != 0 || d > 70000000
}

func (f vx150Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if bytes.Equal(r.Raw(img), vx150Empty) {
		return nil
	}
	c := &Chan{Rx: (100+r.F("freq_mhz").Int(img))*1000000 + r.F("freq_khz").Int(img)*10000 + r.F("sub_khz").Int(img)*2500}
	off := r.F("offset_high").Int(img)*1000000 + r.F("offset_low").Int(img)*10000
	switch r.F("duplex").Int(img) {
	case 2:
		c.Tx = c.Rx - off
	case 3:
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
	c.Power = m.powerSem(0)
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = !r.F("skip").Bool(img)
	var sb strings.Builder
	for _, b := range f.nam(i).F("name").Raw(img) {
		if b == 0xFF {
			break
		}
		if int(b) < len(vx150Chars) {
			sb.WriteRune(vx150Chars[b])
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

func (f vx150Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, nm := f.rec(i), f.nam(i)
	if c == nil {
		r.SetRaw(img, vx150Empty)
		nm.Fill(img, 0xFF)
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
	if c.Rx%2500 != 0 || c.Rx < 100000000 || c.Rx >= 200000000 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	existing := !bytes.Equal(r.Raw(img), vx150Empty)
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz)
	tone, dcs := int64(8), int64(0)
	offHi, offLo := int64(0), int64(60)
	if existing {
		if k := r.F("tone").Int(img); int(k) < len(chirpOldTones) {
			tone = k
		}
		if k := r.F("dtcs").Int(img); int(k) < len(dtcsCodes) {
			dcs = k
		}
		offHi, offLo = r.F("offset_high").Int(img), r.F("offset_low").Int(img)
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
	dup := int64(0)
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	if d != 0 {
		if ad%10000 != 0 || ad >= 100000000 {
			return fmt.Errorf("la radio non può memorizzare questo scostamento fra ricezione e trasmissione (%s MHz): deve essere un multiplo di 10 kHz", fmtMHz(ad))
		}
		offHi, offLo = ad/1000000, ad%1000000/10000
		dup = 2
		if d > 0 {
			dup = 3
		}
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !existing {
		r.F("clock_shift").SetInt(img, 0)
		r.F("active").SetInt(img, 1)
	}
	rem := c.Rx % 1000000
	r.F("freq_mhz").SetInt(img, c.Rx/1000000-100)
	r.F("freq_khz").SetInt(img, rem/10000)
	r.F("sub_khz").SetInt(img, rem%10000/2500)
	r.F("band").SetInt(img, 1)
	r.F("skip").SetBool(img, !c.Scan)
	r.F("duplex").SetInt(img, dup)
	r.F("tmode").SetInt(img, tmode)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("unknown4").SetInt(img, 0x10)
	r.F("offset_high").SetInt(img, offHi)
	r.F("offset_low").SetInt(img, offLo)
	r.F("tone").SetInt(img, tone)
	r.F("dtcs").SetInt(img, dcs)
	r.F("unknown5").SetRaw(img, []byte{0x00, 0x00, 0xaa, 0x00, 0x02, 0x00})
	name := strings.TrimRight(m.cleanName(c.Name), " ")
	enc := bytes.Repeat([]byte{0xFF}, 7)
	for k, ch := range []rune(name) {
		if k >= 7 {
			break
		}
		for p, x := range vx150Chars {
			if x == ch {
				enc[k] = byte(p)
				break
			}
		}
	}
	nm.F("name").SetRaw(img, enc)
	if name != "" {
		nm.F("show_name").SetInt(img, 0x00)
	} else {
		nm.F("show_name").SetInt(img, 0xFF)
	}
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

func (f vx150Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return ftxDownload(ctx, m, port, m.Opt.(*ftxOpt), prog)
}

func (f vx150Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return ftxUpload(ctx, m, port, m.Opt.(*ftxOpt), img, prog)
}
