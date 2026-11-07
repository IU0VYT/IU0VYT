// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// uvk5.py, uvk5_egzumer.py, uvk5_cn.py (4.9), e dal driver CHIRP per il firmware
// F4HWN di armel (https://github.com/armel/uv-k5-chirp-driver, GNU GPL v2 o
// successiva): uvk5_egzumer_f4hwn_ver_4_3_0.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Quansheng UV-K5 / UV-K6 / UV-5R Plus con i firmware liberi:
//   - egzumer (uvk5_egzumer.py di CHIRP): stessa memoria del firmware originale,
//     ma modulazione FM/AM/USB, 24 passi, compander per canale e gamme larghe;
//   - F4HWN (driver di armel, fuori da CHIRP): memoria dell'egzumer con 8
//     livelli di potenza, blocco della trasmissione e 3 liste di scansione per
//     canale; si scrivono anche le impostazioni F4HWN a 0x1FF2-0x1FFF;
//   - LOSEHU e derivati cinesi (uvk5_cn.py di CHIRP): memoria di 256 KiB letta e
//     scritta con i comandi a indirizzo lungo 0x2B / 0x38, nomi in GB2312.
// Il saluto è quello della UV-K5: la radio risponde con la versione del
// firmware e ogni scheda legge e scrive solo il firmware per cui è fatta.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type uvk5xKind int

const (
	k5Egz uvk5xKind = iota // egzumer
	k5F4                   // F4HWN
	k5CN                   // LOSEHU e derivati (firmware cinesi)
)

type uvk5xFam struct{ kind uvk5xKind }

// uvk5egzFormat: campi di uvk5_egzumer.py usati qui (canali, attributi, nomi e
// opzioni di compilazione del firmware)
const uvk5egzFormat = `
struct {
  ul32 freq;
  ul32 offset;
  u8 rxcode;
  u8 txcode;
  u8 txcodeflag:4,
  rxcodeflag:4;
  u8 modulation:4,
  shift:4;
  u8 __UNUSED1:3,
  bclo:1,
  txpower:2,
  bandwidth:1,
  freq_reverse:1;
  u8 __UNUSED2:4,
  dtmf_pttid:3,
  dtmf_decode:1;
  u8 step;
  u8 scrambler;
} channel[214];
struct {
u8 is_scanlist1:1,
is_scanlist2:1,
compander:2,
is_free:1,
band:3;
} channel_attributes[207];
#seekto 0xf50;
struct {
char name[16];
} channelname[200];
#seekto 0x1FF0;
struct {
u8 ENABLE_DTMF_CALLING:1,
   ENABLE_PWRON_PASSWORD:1,
   ENABLE_TX1750:1,
   ENABLE_ALARM:1,
   ENABLE_VOX:1,
   ENABLE_VOICE:1,
   ENABLE_NOAA:1,
   ENABLE_FMRADIO:1;
u8 __UNUSED:2,
   ENABLE_SPECTRUM:1,
   ENABLE_AM_FIX:1,
   ENABLE_BLMIN_TMP_OFF:1,
   ENABLE_RAW_DEMODULATORS:1,
   ENABLE_WIDE_RX:1,
   ENABLE_FLASHLIGHT:1;
} BUILD_OPTIONS;
`

// uvk5f4Format: campi del driver F4HWN usati qui
const uvk5f4Format = `
struct {
  ul32 freq;
  ul32 offset;
  u8 rxcode;
  u8 txcode;
  u8 txcodeflag:4,
  rxcodeflag:4;
  u8 modulation:4,
  offsetDir:4;
  u8 __UNUSED1:1,
  txLock:1,
  busyChLockout:1,
  txpower:3,
  bandwidth:1,
  freq_reverse:1;
  u8 __UNUSED2:4,
  dtmf_pttid:3,
  dtmf_decode:1;
  u8 step;
  u8 scrambler;
} channel[214];
struct {
u8 is_scanlistx:3,
compander:2,
band:3;
} ch_attr[207];
#seekto 0xf50;
struct {
char name[16];
} channelname[200];
#seekto 0x1FF0;
struct {
u8 ENABLE_DTMF_CALLING:1,
   ENABLE_PWRON_PASSWORD:1,
   ENABLE_TX1750:1,
   ENABLE_ALARM:1,
   ENABLE_VOX:1,
   ENABLE_VOICE:1,
   ENABLE_NOAA:1,
   ENABLE_FMRADIO:1;
u8 __UNUSED5:1,
   ENABLE_FEAT_F4HWN_RESCUE_OPS:1,
   ENABLE_BANDSCOPE:1,
   ENABLE_AM_FIX:1,
   ENABLE_FEAT_F4HWN_GAME:1,
   ENABLE_RAW_DEMODULATORS:1,
   ENABLE_WIDE_RX:1,
   ENABLE_FLASHLIGHT:1;
} BUILD_OPTIONS;
`

// gamme dei firmware egzumer e F4HWN (BANDS_STANDARD / BANDS_WIDE)
var (
	uvk5egzBandsStd  = []band{{50000000, 76000000}, {108000000, 600000000}}
	uvk5egzBandsWide = []band{{18000000, 1300000000}}
	// LOSEHU (CN_BANDS): 18-1300 MHz
	uvk5cnBands = []band{{18000000, 1299999999}}
)

// passi dei firmware liberi (24 valori: _steps di egzumer, CN_STEPS)
const uvk5xSteps = 24

// LOSEHU: profili del firmware (CN_PROFILES di uvk5_cn.py, in ordine; vale la
// prima corrispondenza) e quelli gestiti
var uvk5cnProfiles = []struct {
	re      *regexp.Regexp
	profile string
}{
	{regexp.MustCompile(`L_BL[0-9][0-9][0-9]`), "losehubl"},
	{regexp.MustCompile(`LOSEHU.*D`), "losehud"},
	{regexp.MustCompile(`UVE.*`), "losehu124h"},
	{regexp.MustCompile(`LOSEHU13[0-9].*HS`), "losehu124h"},
	{regexp.MustCompile(`LOSEHU13[0-9].*H`), "losehu124h"},
	{regexp.MustCompile(`LOSEHU13[0-9].*KS`), "losehu120k"},
	{regexp.MustCompile(`LOSEHU13[0-9].*K`), "losehu120k"},
	{regexp.MustCompile(`LOSEHU12[4-9].*HS`), "losehu124h"},
	{regexp.MustCompile(`LOSEHU12[4-9].*H`), "losehu124h"},
	{regexp.MustCompile(`LOSEHU12[0-3].*H`), "losehu118h"},
	{regexp.MustCompile(`LOSEHU12[0-9].*KS`), "losehu120k"},
	{regexp.MustCompile(`LOSEHU12[0-9].*K`), "losehu120k"},
	{regexp.MustCompile(`LOSEHU11[8-9].*H`), "losehu118h"},
	{regexp.MustCompile(`LOSEHU11[8-9].*K`), "losehu118k"},
	{regexp.MustCompile(`LOSEHU1[0-9][0-9].*`), "todo"},
	{regexp.MustCompile(`7\..*`), "k1"},
}

var uvk5cnOK = map[string]bool{"losehubl": true, "losehu118h": true, "losehu118k": true, "losehu120k": true, "losehu124h": true, "losehud": true}

func uvk5cnProfile(fw string) string {
	for _, p := range uvk5cnProfiles {
		if p.re.MatchString(fw) {
			return p.profile
		}
	}
	return "todo"
}

const (
	uvk5cnMemSize   = 0x40000 // memoria letta (sopra c'è il caricatore multiplo: mai toccato)
	uvk5cnWriteBlk  = 0x40
	uvk5cnNameBytes = 13 // byte del nome modificabili dal menu della radio
)

func init() {
	mk := func(id, model, variant, rclass, name string, kind uvk5xKind, approve func(string) bool, fw string) *genModel {
		m := &genModel{ID: id, Vendor: "Quansheng", Model: model, Variant: variant, RClass: rclass, Name: name, Slots: 200, First: 1,
			NameMax: 10, Charset: charsetASCII,
			Powers: []genPower{{"Bassa 1,5 W", 1.5, 2}, {"Media 3 W", 3, 1}, {"Alta 5 W", 5, 0}},
			Bands:  uvk5egzBandsWide, Narrow: true, HasScan: false, NoTxOff: false,
			DCS: dtcsCodes, Baud: 38400, Fam: uvk5xFam{kind}, Opt: &uvk5Opt{approve: approve, emuFW: fw}}
		return m
	}
	egz := mk("qs-uvk5-egz", "UV-K5", "egzumer", "UVK5RadioEgzumer", "Quansheng UV-K5 / UV-K6 / UV-5R Plus (firmware egzumer)", k5Egz,
		func(fw string) bool { return strings.HasPrefix(fw, "EGZUMER ") }, "EGZUMER v0.22")
	egz.Notes = "Firmware libero egzumer (la radio risponde «EGZUMER v0.xx»). Accendi la radio, poi collega il cavo al connettore " +
		"microfono/altoparlante (ben inserito). Si legge e si scrive solo con il firmware egzumer; modulazione (FM, AM, USB), passo, " +
		"compander e liste di scansione di ogni canale si conservano. Driver di CHIRP."
	f4 := mk("qs-uvk5-f4hwn", "UV-K5 (egzumer + f4hwn)", "", "UVK5RadioEgzumerF4HWN", "Quansheng UV-K5 / UV-K6 / UV-5R Plus (firmware F4HWN)", k5F4,
		func(fw string) bool { return strings.HasPrefix(fw, "F4HWN") }, "F4HWN v4.3")
	f4.NoTxOff = true
	f4.Powers = []genPower{{"Bassa 1 W (o inferiore, se già impostata)", 1, 2}, {"Media 2 W", 2, 1}, {"Alta 5 W", 5, 0}}
	f4.Notes = "Firmware libero F4HWN (la radio risponde «F4HWN v4.x»). Accendi la radio, poi collega il cavo al connettore " +
		"microfono/altoparlante. Si legge e si scrive solo con il firmware F4HWN; oltre ai canali si scrivono le impostazioni F4HWN " +
		"(0x1FF2-0x1FFF), come il driver di F4HWN per CHIRP. Le potenze sotto 1 W (USER, LOW1-LOW4) impostate sulla radio si " +
		"conservano finché non cambi la potenza del canale; blocco della trasmissione, modulazione e liste di scansione si conservano."
	cn := mk("qs-uvk5-losehu", "UV-K5", "losehu", "UVK5LoseHuRadio", "Quansheng UV-K5 / UV-K6 (firmware LOSEHU, LTS, UVE)", k5CN,
		func(fw string) bool { return uvk5cnOK[uvk5cnProfile(fw)] }, "LOSEHU131HS")
	cn.Bands = uvk5cnBands
	cn.NameMax = uvk5cnNameBytes
	cn.Notes = "Firmware cinesi ampliati LOSEHU (dalla 1.18 H/K), LTS e UVE (nomi in caratteri cinesi GB2312). Accendi la radio, poi " +
		"collega il cavo. Si leggono 256 KiB di memoria (circa 80 secondi) e si scrivono solo i canali e le impostazioni fino a " +
		"0x1CFF: caratteri, logo, radio FM e il caricatore oltre 0x40000 non si toccano mai. Driver di CHIRP (sperimentale)."
	genRegister(egz, f4, cn)
}

func uvk5xFmt(k uvk5xKind) string {
	switch k {
	case k5Egz:
		return uvk5egzFormat
	case k5F4:
		return uvk5f4Format
	}
	return uvk5Format
}

func (f uvk5xFam) mem() *bwEl               { return bwMust(uvk5xFmt(f.kind)) }
func (f uvk5xFam) layout(m *genModel) *bwEl { return f.mem() }

func (f uvk5xFam) attrName() string {
	if f.kind == k5F4 {
		return "ch_attr"
	}
	return "channel_attributes"
}

func (f uvk5xFam) recs(i int) (*bwEl, *bwEl, *bwEl) {
	b := f.mem()
	return b.F("channel").I(i), b.F(f.attrName()).I(i), b.F("channelname").I(i)
}

func (f uvk5xFam) Rec(m *genModel, img []byte, i int) []byte {
	r, a, n := f.recs(i)
	return append(append(append([]byte{}, r.Raw(img)...), a.Raw(img)...), n.Raw(img)...)
}

func (f uvk5xFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, a, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	a.SetRaw(img, raw[16:17])
	n.SetRaw(img, raw[17:33])
}

// Match: le immagini dei firmware liberi si riconoscono dai metadati (versione
// del firmware), non dal contenuto
func (f uvk5xFam) Match(m *genModel, img []byte) bool { return false }

func (f uvk5xFam) crossband() bool { return true }

// wide: gamme larghe (opzione di compilazione ENABLE_WIDE_RX)
func (f uvk5xFam) wide(img []byte) bool {
	if f.kind == k5CN {
		return true
	}
	if len(img) < 0x2000 {
		return true
	}
	return f.mem().F("BUILD_OPTIONS").F("ENABLE_WIDE_RX").Int(img) > 0
}

func (f uvk5xFam) Bands(m *genModel, img []byte) []band {
	switch {
	case f.kind == k5CN:
		return uvk5cnBands
	case f.wide(img):
		return uvk5egzBandsWide
	}
	return uvk5egzBandsStd
}

// banda dell'attributo (_find_band): egzumer / F4HWN dalle gamme del firmware
// (False → 0), LOSEHU dalla funzione del firmware originale senza limiti
func (f uvk5xFam) band(img []byte, hz int64) int64 {
	if f.kind == k5Egz || f.kind == k5CN {
		// set_memory di uvk5.py chiama la funzione _find_band(self, hz) del
		// modulo: «self» fa da nolimits, quindi sempre BANDS_NOLIMITS
		return uvk5Band(hz)
	}
	mhz := float64(hz) / 1e6
	bands := [][2]float64{{50.0, 76.0}, {108.0, 136.9999}, {137.0, 173.9999}, {174.0, 349.9999}, {350.0, 399.9999}, {400.0, 469.9999}, {470.0, 600.0}}
	if f.wide(img) {
		bands = [][2]float64{{18.0, 108.0}, {108.0, 136.9999}, {137.0, 173.9999}, {174.0, 349.9999}, {350.0, 399.9999}, {400.0, 469.9999}, {470.0, 1300.0}}
	}
	for k, b := range bands {
		if b[0] <= mhz && mhz <= b[1] {
			return int64(k)
		}
	}
	return 0
}

// ---------------------------------------------------------------- nomi

func uvk5xNameLatin1(nb []byte) string {
	end := len(nb)
	for k, b := range nb {
		if b == 0x00 || b == 0xFF {
			end = k
			break
		}
	}
	rs := make([]rune, 0, end)
	for _, b := range nb[:end] {
		rs = append(rs, rune(b))
	}
	return strings.TrimRightFunc(string(rs), func(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1C && r <= 0x1F) })
}

// gb2312Pair: coppia di byte valida in GB2312 (EUC-CN)
func gb2312Pair(a, b byte) bool { return a >= 0xA1 && a <= 0xF7 && b >= 0xA1 && b <= 0xFE }

// uvk5cnDecodeName: decode_name di uvk5_cn.py
func uvk5cnDecodeName(raw []byte) string {
	var out strings.Builder
	dec := simplifiedchinese.GBK.NewDecoder()
	for k := 0; k < len(raw); {
		b := raw[k]
		if b == 0x00 || b == 0xFF {
			break
		}
		if b < 0x80 {
			out.WriteByte(b)
			k++
			continue
		}
		if k+1 >= len(raw) || raw[k+1] == 0x00 || raw[k+1] == 0xFF {
			break
		}
		s := "?"
		if gb2312Pair(b, raw[k+1]) {
			if d, err := dec.Bytes(raw[k : k+2]); err == nil && !strings.ContainsRune(string(d), '�') && len([]rune(string(d))) == 1 {
				s = string(d)
			}
		}
		out.WriteString(s)
		k += 2
	}
	return out.String()
}

func uvk5cnEncodeChar(r rune) []byte {
	if r < 0x80 {
		return []byte{byte(r)}
	}
	e, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(string(r)))
	if err != nil || len(e) != 2 || !gb2312Pair(e[0], e[1]) {
		return []byte{'?'}
	}
	return e
}

// uvk5cnEncodeName: encode_name di uvk5_cn.py (al massimo 13 byte, senza
// spezzare un carattere di due byte, poi 0x00 fino a 16)
func uvk5cnEncodeName(name string) []byte {
	var out []byte
	for _, r := range name {
		e := uvk5cnEncodeChar(r)
		if len(out)+len(e) > uvk5cnNameBytes {
			break
		}
		out = append(out, e...)
	}
	return append(out, make([]byte, 16-len(out))...)
}

// ---------------------------------------------------------------- potenza F4HWN

// livelli del firmware F4HWN (txpower, 3 bit): 7 alta 5 W, 6 media 2 W, 5 bassa
// 1 W, 4..1 da 500 a 20 mW, 0 «USER». Qui le tre potenze del programma; i
// livelli sotto 1 W si mostrano come «Bassa» e si conservano.
func f4PowerSem(v int64) int {
	switch v {
	case 7:
		return 0
	case 6:
		return 1
	}
	return 2
}

func f4PowerRaw(sem int) int64 {
	switch sem {
	case 0:
		return 7
	case 1:
		return 6
	}
	return 5
}

// f4PowerKeep: livello da scrivere per il canale i con la potenza sem (quello
// già impostato se ha lo stesso significato)
func (f uvk5xFam) f4PowerKeep(img []byte, i int, sem int) int64 {
	r, _, _ := f.recs(i)
	fr := r.F("freq").Int(img)
	if fr != 0 && fr != 0xFFFFFFFF {
		if old := r.F("txpower").Int(img); f4PowerSem(old) == sem {
			return old
		}
	}
	return f4PowerRaw(sem)
}

// chirpPower: indice della potenza in CHIRP (valid_power_levels del driver
// F4HWN: USER, LOW1-LOW5, MID, HIGH = indici 0-7, cioè il valore del campo)
func (f uvk5xFam) chirpPower(m *genModel, img []byte, i int, sem int) int {
	if f.kind != k5F4 {
		return m.powerIndex(sem)
	}
	return int(f.f4PowerKeep(img, i, sem))
}

// ---------------------------------------------------------------- canali

func (f uvk5xFam) empty(img []byte, i int) bool {
	r, a, _ := f.recs(i)
	fr := r.F("freq").Int(img)
	if fr == 0 || fr == 0xFFFFFFFF {
		return true
	}
	// egzumer e LOSEHU (get_memory di uvk5.py): posizione segnata come libera
	return f.kind != k5F4 && a.F("is_free").Int(img) > 0
}

func (f uvk5xFam) Get(m *genModel, img []byte, i int) *Chan {
	if f.empty(img, i) {
		return nil
	}
	r, _, n := f.recs(i)
	raw := r.F("freq").Int(img)
	c := &Chan{Rx: raw * 10, Scan: true}
	if f.kind != k5F4 {
		c.Rx = uvk5LegalAir(raw * 10)
	}
	off := r.F("offset").Int(img)
	c.Tx = c.Rx
	if f.kind == k5F4 {
		switch r.F("offsetDir").Int(img) {
		case 2:
			c.Tx = c.Rx - off*10
		case 1:
			c.Tx = c.Rx + off*10
		}
	} else if off != 0 {
		switch shift := r.F("shift").Int(img); {
		case shift == 2 && raw == off:
			c.TxOff = true
		case shift == 2:
			c.Tx = c.Rx - off*10
		case shift == 1:
			c.Tx = c.Rx + off*10
		}
	}
	tf, rf := r.F("txcodeflag").Int(img), r.F("rxcodeflag").Int(img)
	if f.kind == k5CN && (tf > 3 || rf > 3) {
		tf, rf = 0, 0 // record scritto a metà (_get_tone di uvk5_cn.py)
	}
	c.TxTone = uvk5ToneDec(tf, r.F("txcode").Int(img))
	c.RxTone = uvk5ToneDec(rf, r.F("rxcode").Int(img))
	c.Narrow = r.F("bandwidth").Int(img) > 0
	if f.kind == k5F4 {
		c.Power = f4PowerSem(r.F("txpower").Int(img))
	} else {
		switch r.F("txpower").Int(img) {
		case 2:
			c.Power = 0
		case 1:
			c.Power = 1
		default:
			c.Power = 2
		}
	}
	nb := n.F("name").Raw(img)
	if f.kind == k5CN {
		c.Name = strings.TrimRightFunc(uvk5cnDecodeName(nb), unicode.IsSpace)
	} else {
		c.Name = uvk5xNameLatin1(nb)
	}
	return c
}

func (f uvk5xFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	switch f.kind {
	case k5F4:
		return f.setF4(m, img, i, c)
	}
	return f.setEgzCN(m, img, i, c)
}

// setEgzCN: set_memory di uvk5.py con le aggiunte di uvk5_egzumer.py (modulazione,
// compander) o di uvk5_cn.py (nomi GB2312)
func (f uvk5xFam) setEgzCN(m *genModel, img []byte, i int, c *Chan) error {
	r, a, n := f.recs(i)
	if c == nil {
		if f.kind == k5CN && f.alreadyEmpty(img, i) {
			return nil // _already_empty: i byte dicono già «vuoto»
		}
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
		a.F("is_scanlist1").SetInt(img, 0)
		a.F("is_scanlist2").SetInt(img, 0)
		a.F("compander").SetInt(img, 0)
		a.F("is_free").SetInt(img, 1)
		a.F("band").SetInt(img, 7)
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
	old := append([]byte{}, r.Raw(img)...)
	oldName := append([]byte{}, n.Raw(img)...)
	empty := f.empty(img, i)
	oldFreq := r.F("freq").Int(img)
	// valori conservati come in CHIRP (get_memory e _get_mem_extra)
	bclo := !empty && r.F("bclo").Int(img) > 0
	frev := !empty && r.F("freq_reverse").Int(img) > 0
	dtmf := !empty && r.F("dtmf_decode").Int(img) > 0
	pttid := r.F("dtmf_pttid").Int(img)
	if f.kind == k5Egz && pttid >= 5 {
		pttid = 0
	}
	enc := r.F("scrambler").Int(img) & 0x0F
	if enc >= 11 {
		enc = 0
	}
	sl1, sl2 := a.F("is_scanlist1").Int(img), a.F("is_scanlist2").Int(img)
	comp := a.F("compander").Int(img)
	step := int64(1) // canale nuovo: 5 kHz
	if !empty {
		if step = r.F("step").Int(img); step >= uvk5xSteps {
			step = 0
		}
	}
	// modulazione: quella del canale se la frequenza non cambia, AM nella gamma
	// aeronautica, altrimenti FM
	mod := int64(0)
	if f.kind == k5Egz {
		if !empty && uvk5LegalAir(oldFreq*10) == c.Rx {
			mod = r.F("modulation").Int(img)
		} else if c.Rx >= 108000000 && c.Rx <= 137000000 {
			mod = 1
		}
	}
	oldAM := f.kind == k5CN && r.F("enable_am").Int(img) > 0
	if old[0] == 0xFF {
		r.SetRaw(img, make([]byte, 16))
	} else {
		r.SetRaw(img, append(make([]byte, 10), old[10]&0xCC, old[11]&0xEC, old[12]&0xE0, old[13]&0xF8, old[14]&0xF1, old[15]&0xF0))
	}
	if f.kind == k5Egz {
		r.F("modulation").SetInt(img, mod)
		r.F("bandwidth").SetBool(img, c.Narrow || mod == 2)
	} else {
		r.F("bandwidth").SetBool(img, c.Narrow)
		r.F("enable_am").SetBool(img, (c.Rx >= 108000000 && c.Rx <= 137000000) || (!empty && oldAM && uvk5LegalAir(oldFreq*10) == c.Rx))
	}
	freq := c.Rx / 10
	r.F("freq").SetInt(img, freq)
	shiftF := "shift"
	switch {
	case c.TxOff:
		r.F(shiftF).SetInt(img, 2)
		r.F("offset").SetInt(img, freq)
	case c.Tx == c.Rx:
		r.F(shiftF).SetInt(img, 0)
		r.F("offset").SetInt(img, 0)
	case c.Tx > c.Rx:
		r.F(shiftF).SetInt(img, 1)
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
	default:
		r.F(shiftF).SetInt(img, 2)
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
	}
	a.F("is_scanlist1").SetInt(img, 0)
	a.F("is_scanlist2").SetInt(img, 0)
	a.F("compander").SetInt(img, 0)
	a.F("is_free").SetInt(img, 0)
	a.F("band").SetInt(img, f.band(img, c.Rx))
	if f.kind == k5CN {
		// _store_name: se il nome non cambia si rimettono i byte di prima
		if !bytes.Equal(oldName, bytes.Repeat([]byte{0xFF}, 16)) &&
			strings.TrimRightFunc(uvk5cnDecodeName(oldName), unicode.IsSpace) == strings.TrimRightFunc(c.Name, unicode.IsSpace) {
			n.SetRaw(img, oldName)
		} else {
			n.SetRaw(img, uvk5cnEncodeName(c.Name))
		}
	} else {
		nm := make([]byte, 16)
		for k := 0; k < 10; k++ {
			nm[k] = ' '
		}
		for k, ch := range []rune(c.Name) {
			if k >= 10 {
				break
			}
			if ch > 0xFF {
				ch = '?'
			}
			nm[k] = byte(ch)
		}
		n.SetRaw(img, nm)
	}
	fl, cd := uvk5ToneEnc(rt)
	r.F("rxcodeflag").SetInt(img, fl)
	r.F("rxcode").SetInt(img, cd)
	fl, cd = uvk5ToneEnc(tt)
	r.F("txcodeflag").SetInt(img, fl)
	r.F("txcode").SetInt(img, cd)
	r.F("step").SetInt(img, step)
	switch m.powerIndex(c.Power) {
	case 2:
		r.F("txpower").SetInt(img, 2)
	case 1:
		r.F("txpower").SetInt(img, 1)
	default:
		r.F("txpower").SetInt(img, 0)
	}
	r.F("bclo").SetBool(img, bclo)
	r.F("dtmf_pttid").SetInt(img, pttid)
	r.F("freq_reverse").SetBool(img, frev)
	r.F("dtmf_decode").SetBool(img, dtmf)
	r.F("scrambler").SetInt(img, (r.F("scrambler").Int(img)&0xF0)|enc)
	a.F("is_scanlist1").SetInt(img, sl1)
	a.F("is_scanlist2").SetInt(img, sl2)
	if f.kind == k5Egz {
		a.F("compander").SetInt(img, comp)
	}
	return nil
}

// alreadyEmpty: _already_empty di uvk5_cn.py
func (f uvk5xFam) alreadyEmpty(img []byte, i int) bool {
	r, a, n := f.recs(i)
	ff, zz := bytes.Repeat([]byte{0xFF}, 16), make([]byte, 16)
	rec, nm := r.Raw(img), n.Raw(img)
	if !bytes.Equal(rec, ff) && !bytes.Equal(rec, zz) {
		return false
	}
	if !bytes.Equal(nm, ff) && !bytes.Equal(nm, zz) {
		return false
	}
	return a.F("is_free").Int(img) > 0
}

// setF4: set_memory del driver F4HWN. Il record non si azzera: si scrivono i
// campi (restano solo i bit non usati).
func (f uvk5xFam) setF4(m *genModel, img []byte, i int, c *Chan) error {
	r, a, n := f.recs(i)
	if c != nil && c.TxOff {
		return fmt.Errorf("con il firmware F4HWN la trasmissione si blocca dal menu della radio (TxLock), non dal canale")
	}
	empty := f.empty(img, i)
	// valori conservati (extra di get_memory: per una posizione vuota i valori predefiniti)
	var txLock, bcl, pttid, frev, dtmf, scr, comp, scan int64
	step := int64(1)
	offset := int64(60000) // canale nuovo: lo scostamento predefinito di CHIRP (600 kHz)
	if !empty {
		txLock = r.F("txLock").Int(img)
		bcl = r.F("busyChLockout").Int(img)
		if pttid = r.F("dtmf_pttid").Int(img); pttid >= 5 {
			pttid = 0
		}
		frev = r.F("freq_reverse").Int(img)
		if f.mem().F("BUILD_OPTIONS").F("ENABLE_DTMF_CALLING").Int(img) > 0 {
			dtmf = r.F("dtmf_decode").Int(img)
		}
		if scr = r.F("scrambler").Int(img); scr >= 11 {
			scr = 0
		}
		comp = a.F("compander").Int(img)
		scan = a.F("is_scanlistx").Int(img)
		if step = r.F("step").Int(img); step >= uvk5xSteps {
			step = 0
		}
		offset = r.F("offset").Int(img)
	}
	var power, mod int64
	if c != nil {
		power = f.f4PowerKeep(img, i, c.Power)
		if !empty && r.F("freq").Int(img)*10 == c.Rx {
			mod = r.F("modulation").Int(img)
		} else if c.Rx >= 108000000 && c.Rx <= 137000000 {
			mod = 1
		}
	}
	a.F("is_scanlistx").SetInt(img, 7)
	a.F("compander").SetInt(img, 0)
	a.F("band").SetInt(img, 7)
	if c == nil {
		r.Fill(img, 0xFF)
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
	r.F("modulation").SetInt(img, mod)
	r.F("bandwidth").SetBool(img, c.Narrow || mod == 2)
	r.F("freq").SetInt(img, c.Rx/10)
	switch {
	case c.Tx > c.Rx:
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
		r.F("offsetDir").SetInt(img, 1)
	case c.Tx < c.Rx:
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
		r.F("offsetDir").SetInt(img, 2)
	default:
		// simplex: come CHIRP, lo scostamento memorizzato resta
		r.F("offset").SetInt(img, offset)
		r.F("offsetDir").SetInt(img, 0)
	}
	a.F("band").SetInt(img, f.band(img, c.Rx))
	nm := make([]byte, 16)
	for k := 0; k < 10; k++ {
		nm[k] = ' '
	}
	for k, ch := range []rune(c.Name) {
		if k >= 10 {
			break
		}
		if ch > 0xFF {
			ch = '?'
		}
		nm[k] = byte(ch)
	}
	n.SetRaw(img, nm)
	fl, cd := uvk5ToneEnc(rt)
	r.F("rxcodeflag").SetInt(img, fl)
	r.F("rxcode").SetInt(img, cd)
	fl, cd = uvk5ToneEnc(tt)
	r.F("txcodeflag").SetInt(img, fl)
	r.F("txcode").SetInt(img, cd)
	r.F("step").SetInt(img, step)
	r.F("txpower").SetInt(img, power)
	r.F("txLock").SetInt(img, txLock)
	r.F("busyChLockout").SetInt(img, bcl)
	r.F("dtmf_pttid").SetInt(img, pttid)
	r.F("freq_reverse").SetInt(img, frev)
	r.F("dtmf_decode").SetInt(img, dtmf)
	r.F("scrambler").SetInt(img, scr)
	a.F("compander").SetInt(img, comp)
	a.F("is_scanlistx").SetInt(img, scan)
	return nil
}

// ---------------------------------------------------------------- protocollo

// comandi a indirizzo lungo dei firmware LOSEHU (readmem_command e
// writemem_command di uvk5_cn.py)
func uvk5cnReadCmd(off, n int) []byte {
	b := []byte{0x2b, 0x05, 0x08, 0x00, 0, 0, byte(n), 0, 0x6a, 0x39, 0x57, 0x64, 0, 0}
	binary.LittleEndian.PutUint16(b[4:], uint16(off>>16))
	binary.LittleEndian.PutUint16(b[12:], uint16(off))
	return b
}

func uvk5cnWriteCmd(off int, data []byte) []byte {
	b := []byte{0x38, 0x05, 0, 0, 0, 0, byte(len(data) + 2), 0, 0x6a, 0x39, 0x57, 0x64, 0, 0}
	binary.LittleEndian.PutUint16(b[2:], uint16(10+len(data)))
	binary.LittleEndian.PutUint16(b[4:], uint16(off>>16))
	binary.LittleEndian.PutUint16(b[12:], uint16(off))
	return append(b, data...)
}

func (s *uvk5Sess) cnReadMem(off, n int) ([]byte, error) {
	if off < 0 || off+n > uvk5cnMemSize {
		return nil, fmt.Errorf("lettura oltre 0x%X rifiutata (caricatore della radio)", uvk5cnMemSize)
	}
	if err := s.cmd(uvk5cnReadCmd(off, n)); err != nil {
		return nil, err
	}
	rep, err := s.reply()
	if err != nil {
		return nil, err
	}
	if len(rep) < 8+n {
		return nil, rErr("lettura della memoria a 0x%05X incompleta", off)
	}
	return rep[8 : 8+n], nil
}

func (s *uvk5Sess) cnWriteMem(off int, data []byte) error {
	if off < 0 || off+len(data) > uvk5cnMemSize {
		return fmt.Errorf("scrittura oltre 0x%X rifiutata: lì c'è il caricatore della radio", uvk5cnMemSize)
	}
	if err := s.cmd(uvk5cnWriteCmd(off, data)); err != nil {
		return err
	}
	rep, err := s.reply()
	if err != nil {
		return err
	}
	if len(rep) < 1 || rep[0] != 0x1e {
		return rErr("la radio non ha confermato la scrittura a 0x%05X", off)
	}
	return nil
}

func (f uvk5xFam) memSize() int {
	if f.kind == k5CN {
		return uvk5cnMemSize
	}
	return uvk5MemSize
}

func (f uvk5xFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return mem, note, err
}

func (f uvk5xFam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	p, err := openPort(port, 38400)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	s := &uvk5Sess{ctx: ctx, p: p}
	size := f.memSize()
	prog(0, size, "Collegamento con la radio…")
	fw, err := s.start(m)
	if err != nil {
		return nil, nil, "", err
	}
	var data []byte
	for a := 0; a < size; a += uvk5Block {
		if ctx.Err() != nil {
			return nil, nil, "", errCancelled
		}
		prog(a, size, "Lettura…")
		var b []byte
		if f.kind == k5CN {
			b, err = s.cnReadMem(a, uvk5Block)
		} else {
			b, err = s.readMem(a, uvk5Block)
		}
		if err != nil {
			return nil, nil, "", err
		}
		data = append(data, b...)
	}
	prog(size, size, "Lettura completata")
	return data, map[string]interface{}{"uvk5_firmware": fw}, "Firmware della radio: " + fw, nil
}

// zone scritte: i canali e le impostazioni fino a 0x1CFF; con F4HWN anche le
// sue impostazioni a 0x1FF2-0x1FFF (do_upload del driver F4HWN)
func (f uvk5xFam) writeRanges() [][2]int {
	if f.kind == k5F4 {
		return [][2]int{{0, uvk5ProgSize}, {0x1FF2, uvk5MemSize}}
	}
	return [][2]int{{0, uvk5ProgSize}}
}

func (f uvk5xFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != f.memSize() {
		if f.kind == k5CN {
			return "", fmt.Errorf("l'immagine è di %d byte: per questi firmware deve essere di %d byte (leggi prima la radio)", len(img), uvk5cnMemSize)
		}
		if len(img) < uvk5MemSize {
			return "", fmt.Errorf("immagine della memoria incompleta")
		}
	}
	p, err := openPort(port, 38400)
	if err != nil {
		return "", err
	}
	defer p.Close()
	s := &uvk5Sess{ctx: ctx, p: p}
	total := 0
	for _, r := range f.writeRanges() {
		total += r[1] - r[0]
	}
	prog(0, total, "Collegamento con la radio…")
	fw, err := s.start(m)
	if err != nil {
		return "", err
	}
	done := 0
	for _, rg := range f.writeRanges() {
		blk := uvk5Block
		if f.kind == k5CN {
			blk = uvk5cnWriteBlk
		}
		for a := rg[0]; a < rg[1]; a += blk {
			prog(done+a-rg[0], total, "Scrittura…")
			end := min(a+blk, len(img))
			if f.kind == k5F4 {
				end = min(a+blk, uvk5MemSize) // come il driver: il blocco finisce con la memoria
			}
			if f.kind == k5CN {
				err = s.cnWriteMem(a, img[a:end])
			} else {
				err = s.writeMem(a, img[a:end])
			}
			if err != nil {
				return "", err
			}
		}
		done += rg[1] - rg[0]
	}
	_ = s.cmd([]byte{0xdd, 0x05, 0x00, 0x00})
	prog(total, total, "Scrittura completata")
	return "Firmware della radio: " + fw, nil
}
