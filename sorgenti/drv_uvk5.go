// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: uvk5.py
// (Copyright 2023 Jacek Lipkowski SQ5BPF e altri autori di CHIRP).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Quansheng UV-K5 / UV-K6 / UV-5R Plus, Retevis RA79, MaxTalker TK-6 e
// Quansheng UV-K1 (driver uvk5.py di CHIRP, firmware originale).
//
// Protocollo a 38400 baud: ogni comando è «AB CD lunghezza 00», i dati con il
// CRC-16 XMODEM offuscati con una tabella XOR di 16 byte, poi «DC BA». La radio
// risponde al saluto con la versione del firmware: come CHIRP si legge e si
// scrive solo con le versioni del firmware originale conosciute (per le altre la
// struttura della memoria potrebbe essere diversa).
// Memoria: 200 canali da 16 byte a 0x0000, attributi a 0x0D60 (banda, liste
// di scansione, posizione libera), nomi di 10 caratteri a 0x0F50.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode"
)

const uvk5Format = `
#seekto 0x0000;
struct {
  ul32 freq;
  ul32 offset;
  u8 rxcode;
  u8 txcode;
  u8 txcodeflag:4,
     rxcodeflag:4;
  u8 flags1_unknown7:1,
     flags1_unknown6:1,
     flags1_unknown5:1,
     enable_am:1,
     flags1_unknown3:1,
     is_in_scanlist:1,
     shift:2;
  u8 flags2_unknown7:1,
     flags2_unknown6:1,
     flags2_unknown5:1,
     bclo:1,
     txpower:2,
     bandwidth:1,
     freq_reverse:1;
  u8 dtmf_flags_unknown7:1,
     dtmf_flags_unknown6:1,
     dtmf_flags_unknown5:1,
     dtmf_flags_unknown4:1,
     dtmf_flags_unknown3:1,
     dtmf_pttid:2,
     dtmf_decode:1;
  u8 step;
  u8 scrambler;
} channel[214];
#seekto 0x0D60;
struct {
  u8 is_scanlist1:1,
     is_scanlist2:1,
     compander:2,
     is_free:1,
     band:3;
} channel_attributes[200];
#seekto 0x0F50;
struct {
  char name[16];
} channelname[200];
`

const (
	uvk5MemSize  = 0x2000
	uvk5ProgSize = 0x1D00
	uvk5Block    = 0x80
)

type uvk5Opt struct {
	approve func(fw string) bool
	// detect: CHIRP sceglie il driver dal firmware (detect_from_serial) con un
	// primo saluto prima della lettura (solo la classe UV-K5)
	detect bool
	emuFW  string // 4.9: versione del firmware della radio simulata dei test
}

type uvk5Fam struct{}

// bande del firmware originale (BANDS di uvk5.py) e, per la scelta della banda
// del canale, quelle senza limiti (BANDS_NOLIMITS: CHIRP usa sempre queste)
var uvk5BandsNoLimits = [][2]float64{{18.0, 76.0}, {108.0, 136.9999}, {137.0, 173.9999}, {174.0, 349.9999}, {350.0, 399.9999}, {400.0, 469.9999}, {470.0, 1300.0}}

func uvk5ApproveK5(fw string) bool {
	for _, p := range []string{"k5_2.01.", "app_2.01.", "2.01.", "3.00.", "4.00.", "k5_4.00.", "5.00.", "7.00.", "7.01.", "1o11", "1.02."} {
		if strings.HasPrefix(fw, p) {
			return true
		}
	}
	return false
}

func init() {
	mk := func(id, vendor, model, rclass, name string, approve func(string) bool) *genModel {
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: 200, First: 1,
			NameMax: 10, Charset: charsetASCII,
			Powers: []genPower{{"Bassa 1,5 W", 1.5, 2}, {"Media 3 W", 3, 1}, {"Alta 5 W", 5, 0}},
			Bands:  []band{{50000000, 76000000}, {108000000, 600000000}}, Narrow: true, HasScan: false,
			DCS: dtcsCodes, Baud: 38400, Fam: uvk5Fam{}, Opt: &uvk5Opt{approve: approve},
			Notes: "Accendi la radio, poi collega il cavo al connettore microfono/altoparlante (ben inserito). " +
				"Come CHIRP, la lettura e la scrittura si fanno solo con le versioni del firmware originale conosciute: " +
				"con un firmware modificato il programma si ferma senza scrivere."}
	}
	k5 := mk("qs-uvk5", "Quansheng", "UV-K5", "UVK5Radio", "Quansheng UV-K5 / UV-K6 / UV-5R Plus", uvk5ApproveK5)
	uvk5o(k5).detect = true
	k1 := mk("qs-uvk1", "Quansheng", "UV-K1", "UVK1Radio", "Quansheng UV-K1", func(fw string) bool { return strings.HasPrefix(fw, "7.03.") })
	// 4.9: firmware libero «OSFW» (OSFWUVK5Radio di uvk5.py): memoria del
	// firmware originale, gamme senza limiti
	osfw := mk("qs-uvk5-osfw", "Quansheng", "UV-K5", "OSFWUVK5Radio", "Quansheng UV-K5 / UV-K6 / UV-5R Plus (firmware OSFW)",
		func(fw string) bool { return fw == "OSFW-bd90ca3" })
	osfw.Variant = "OSFW"
	osfw.Bands = []band{{18000000, 1300000000}}
	uvk5o(osfw).emuFW = "OSFW-bd90ca3"
	osfw.Notes = "Firmware libero OSFW (versione bd90ca3, la sola gestita da CHIRP). Accendi la radio, poi collega il cavo al " +
		"connettore microfono/altoparlante. Stessa memoria del firmware originale; gamme senza limiti. Driver di CHIRP."
	genRegister(
		k5,
		osfw,
		mk("rt-ra79", "Retevis", "RA79", "RA79Radio", "Retevis RA79", uvk5ApproveK5),
		mk("mt-tk6", "MaxTalker", "TK-6", "MaxTalkerTK6", "MaxTalker TK-6", uvk5ApproveK5),
		k1,
	)
}

func uvk5o(m *genModel) *uvk5Opt { return m.Opt.(*uvk5Opt) }

func (uvk5Fam) mem() *bwEl                 { return bwMust(uvk5Format) }
func (f uvk5Fam) layout(m *genModel) *bwEl { return f.mem() }

func (f uvk5Fam) recs(i int) (*bwEl, *bwEl, *bwEl) {
	b := f.mem()
	return b.F("channel").I(i), b.F("channel_attributes").I(i), b.F("channelname").I(i)
}

func (f uvk5Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, a, n := f.recs(i)
	return append(append(append([]byte{}, r.Raw(img)...), a.Raw(img)...), n.Raw(img)...)
}

func (f uvk5Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, a, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	a.SetRaw(img, raw[16:17])
	n.SetRaw(img, raw[17:33])
}

func (f uvk5Fam) Match(m *genModel, img []byte) bool { return false }

// uvk5LegalAir: frequenza aeronautica come _legal_airband_freq di uvk5.py.
// 4.9: come CHIRP, la frequenza memorizzata resta com'è; solo una frequenza
// allineata a 5 kHz che non è uno dei tre canali da 8,33 kHz del blocco di
// 25 kHz (fix_rounded_step la rifiuta) diventa il canale più vicino, compreso
// l'inizio del blocco seguente. Prima si arrotondava sempre al canale da
// 8,33 kHz, e una frequenza come 127,0875 MHz veniva letta 127,091666 MHz.
func uvk5LegalAir(hz int64) int64 {
	if hz < 108000000 || hz > 137000000 || hz%25000 == 0 {
		return hz
	}
	base := hz / 25000 * 25000
	if (hz-base)%5000 != 0 || (hz-base)/5000 < 4 {
		return hz
	}
	best := base
	for j := int64(0); j < 4; j++ {
		c := base + (25000/3)*j
		if abs64(c-hz) < abs64(best-hz) {
			best = c
		}
	}
	return best
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func uvk5ToneDec(flag, code int64) string {
	switch flag {
	case 1:
		if int(code) < len(chirpTones) {
			return fmtTone(chirpTones[code])
		}
	case 2, 3:
		if int(code) < len(dtcsCodes) {
			return genTone{Mode: "DTCS", Code: dtcsCodes[code], Inv: flag == 3}.String()
		}
	}
	return ""
}

func uvk5ToneEnc(t genTone) (flag, code int64) {
	switch t.Mode {
	case "Tone":
		if k := idxInt(chirpTones, t.Hz10); k >= 0 {
			return 1, int64(k)
		}
	case "DTCS":
		if k := idxInt(dtcsCodes, t.Code); k >= 0 {
			if t.Inv {
				return 3, int64(k)
			}
			return 2, int64(k)
		}
	}
	return 0, 0
}

func (f uvk5Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, _, n := f.recs(i)
	raw := r.F("freq").Int(img)
	if raw == 0 || raw == 0xFFFFFFFF {
		return nil
	}
	c := &Chan{Rx: uvk5LegalAir(raw * 10)}
	off := r.F("offset").Int(img)
	shift := r.F("shift").Int(img)
	c.Tx = c.Rx
	if off != 0 {
		switch {
		case shift == 2 && raw == off:
			c.TxOff = true
		case shift == 2:
			c.Tx = c.Rx - off*10
		case shift == 1:
			c.Tx = c.Rx + off*10
		}
	}
	c.TxTone = uvk5ToneDec(r.F("txcodeflag").Int(img), r.F("txcode").Int(img))
	c.RxTone = uvk5ToneDec(r.F("rxcodeflag").Int(img), r.F("rxcode").Int(img))
	c.Narrow = r.F("bandwidth").Int(img) > 0
	switch r.F("txpower").Int(img) {
	case 2:
		c.Power = 0
	case 1:
		c.Power = 1
	default:
		c.Power = 2
	}
	// nome: fino al primo 0x00 o 0xFF (latin-1), spazi finali tolti
	nb := n.F("name").Raw(img)
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
	// come str.rstrip() di Python
	c.Name = strings.TrimRightFunc(string(rs), func(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1C && r <= 0x1F) })
	c.Scan = true
	return c
}

// uvk5Band: banda del canale negli attributi (_find_band di uvk5.py, senza limiti)
func uvk5Band(hz int64) int64 {
	mhz := float64(hz) / 1e6
	if mhz < 50.0 {
		return 1
	}
	for k, b := range uvk5BandsNoLimits {
		if mhz >= b[0] && mhz <= b[1] {
			return int64(k)
		}
	}
	return 0
}

// crossband: spostamenti oltre 70 MHz scritti come scostamento (+/-), che la
// radio ammette; CHIRP li tratta come «split», che questa radio non ha
func (f uvk5Fam) crossband() bool { return true }

func (f uvk5Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, a, n := f.recs(i)
	if c == nil {
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
	oldFreq := r.F("freq").Int(img)
	empty := oldFreq == 0 || oldFreq == 0xFFFFFFFF
	// impostazioni del canale conservate come in CHIRP (valori di get_memory)
	bclo := !empty && r.F("bclo").Int(img) > 0
	frev := !empty && r.F("freq_reverse").Int(img) > 0
	pttid := r.F("dtmf_pttid").Int(img)
	dtmf := !empty && r.F("dtmf_decode").Int(img) > 0
	enc := r.F("scrambler").Int(img) & 0x0F
	if enc >= 11 {
		enc = 0
	}
	sl1, sl2 := a.F("is_scanlist1").Int(img), a.F("is_scanlist2").Int(img)
	oldAM := r.F("enable_am").Int(img) > 0
	oldRx := uvk5LegalAir(oldFreq * 10)
	step := int64(1) // canale nuovo: passo 5 kHz (valore predefinito di CHIRP)
	if !empty {
		step = r.F("step").Int(img)
		if step >= 6 {
			step = 0
		}
	}
	if old[0] == 0xFF {
		r.SetRaw(img, make([]byte, 16))
	} else {
		r.SetRaw(img, append(make([]byte, 10), old[10]&0xCC, old[11]&0xEC, old[12]&0xE0, old[13]&0xF8, old[14]&0xF1, old[15]&0xF0))
	}
	r.F("bandwidth").SetBool(img, c.Narrow)
	// AM: nella gamma aeronautica, o se il canale era già in AM sulla stessa frequenza
	r.F("enable_am").SetBool(img, (c.Rx >= 108000000 && c.Rx <= 137000000) || (!empty && oldAM && oldRx == c.Rx))
	freq := c.Rx / 10
	r.F("freq").SetInt(img, freq)
	switch {
	case c.TxOff:
		r.F("shift").SetInt(img, 2)
		r.F("offset").SetInt(img, freq)
	case c.Tx == c.Rx:
		r.F("shift").SetInt(img, 0)
		r.F("offset").SetInt(img, 0)
	case c.Tx > c.Rx:
		r.F("shift").SetInt(img, 1)
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
	default:
		r.F("shift").SetInt(img, 2)
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
	}
	a.F("is_scanlist1").SetInt(img, 0)
	a.F("is_scanlist2").SetInt(img, 0)
	a.F("compander").SetInt(img, 0)
	a.F("is_free").SetInt(img, 0)
	a.F("band").SetInt(img, uvk5Band(c.Rx))
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
	return nil
}

// ---------------------------------------------------------------- protocollo

var uvk5XOR = []byte{22, 108, 20, 230, 46, 145, 13, 64, 33, 53, 213, 64, 19, 3, 233, 128}

func uvk5Xor(b []byte) []byte {
	out := make([]byte, len(b))
	for k, v := range b {
		out[k] = v ^ uvk5XOR[k%len(uvk5XOR)]
	}
	return out
}

func crc16XModem(b []byte) uint16 {
	var crc uint16
	for _, v := range b {
		crc ^= uint16(v) << 8
		for k := 0; k < 8; k++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// uvk5Packet: comando o risposta impacchettato come sul cavo
func uvk5Packet(data []byte) []byte {
	crc := make([]byte, 2)
	binary.LittleEndian.PutUint16(crc, crc16XModem(data))
	out := []byte{0xAB, 0xCD, byte(len(data)), 0x00}
	out = append(out, uvk5Xor(append(append([]byte{}, data...), crc...))...)
	return append(out, 0xDC, 0xBA)
}

type uvk5Sess struct {
	ctx context.Context
	p   Port
}

const uvk5Timeout = 1500 * time.Millisecond

func (s *uvk5Sess) cmd(data []byte) error { return s.p.Write(uvk5Packet(data)) }

func (s *uvk5Sess) reply() ([]byte, error) {
	h := make([]byte, 4)
	if err := s.p.ReadFull(s.ctx, h, uvk5Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	if h[0] != 0xAB || h[1] != 0xCD || h[3] != 0x00 {
		return nil, rErr("risposta della radio non valida")
	}
	body := make([]byte, int(h[2]))
	if err := s.p.ReadFull(s.ctx, body, uvk5Timeout); err != nil {
		return nil, rErr("risposta della radio incompleta")
	}
	foot := make([]byte, 4)
	if err := s.p.ReadFull(s.ctx, foot, uvk5Timeout); err != nil || foot[2] != 0xDC || foot[3] != 0xBA {
		return nil, rErr("risposta della radio incompleta")
	}
	return uvk5Xor(body), nil
}

// hello: versione del firmware
func (s *uvk5Sess) hello() (string, error) {
	if err := s.cmd([]byte{0x14, 0x05, 0x04, 0x00, 0x6a, 0x39, 0x57, 0x64}); err != nil {
		return "", err
	}
	rep, err := s.reply()
	if err != nil {
		return "", err
	}
	if bytes.HasPrefix(rep, []byte{0x18, 0x05}) {
		return "", rErr("la radio è in modalità aggiornamento del firmware: spegnila e riaccendila normalmente")
	}
	var fw []byte
	for k := 4; k < len(rep) && k <= 24; k++ {
		if rep[k] < ' ' || rep[k] > '~' {
			break
		}
		fw = append(fw, rep[k])
	}
	return string(fw), nil
}

func (s *uvk5Sess) readMem(off, n int) ([]byte, error) {
	cmd := []byte{0x1b, 0x05, 0x08, 0x00, byte(off), byte(off >> 8), byte(n), 0, 0x6a, 0x39, 0x57, 0x64}
	if err := s.cmd(cmd); err != nil {
		return nil, err
	}
	rep, err := s.reply()
	if err != nil {
		return nil, err
	}
	if len(rep) < 8+n {
		return nil, rErr("lettura della memoria a 0x%04X incompleta", off)
	}
	return rep[8 : 8+n], nil
}

func (s *uvk5Sess) writeMem(off int, data []byte) error {
	cmd := []byte{0x1d, 0x05, byte(len(data) + 8), 0, byte(off), byte(off >> 8), byte(len(data)), 1, 0x6a, 0x39, 0x57, 0x64}
	if err := s.cmd(append(cmd, data...)); err != nil {
		return err
	}
	rep, err := s.reply()
	if err != nil {
		return err
	}
	if len(rep) < 6 || rep[0] != 0x1e || rep[4] != byte(off) || rep[5] != byte(off>>8) {
		return rErr("la radio non ha confermato la scrittura a 0x%04X", off)
	}
	return nil
}

func (s *uvk5Sess) start(m *genModel) (string, error) {
	fw, err := s.hello()
	if err != nil {
		return "", genFail("Collegamento", err)
	}
	if fw == "" {
		return "", rErr("la radio non ha comunicato la versione del firmware")
	}
	if !uvk5o(m).approve(fw) {
		// 4.9: scheda giusta per il firmware della radio (firmware liberi)
		for _, o := range genModels {
			if op, ok := o.Opt.(*uvk5Opt); ok && o != m && op.approve(fw) {
				return fw, rErr("la radio ha il firmware %s: apri la scheda «%s» (questa scheda è per un altro firmware e, come CHIRP, non legge e non scrive)", fw, o.Name)
			}
		}
		return fw, rErr("il firmware della radio (%s) non è tra quelli conosciuti per la %s: per sicurezza non si legge e non si scrive (come in CHIRP)", fw, m.Name)
	}
	return fw, nil
}

func (f uvk5Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, meta, note, err := f.DownloadMeta(ctx, m, port, prog)
	_ = meta
	return mem, note, err
}

// DownloadMeta: come Download, con la versione del firmware nei metadati del
// file .img (uvk5_firmware, come CHIRP)
func (f uvk5Fam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	p, err := openPort(port, 38400)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	s := &uvk5Sess{ctx: ctx, p: p}
	prog(0, uvk5MemSize, "Collegamento con la radio…")
	// come CHIRP: un primo saluto riconosce la radio e il firmware
	// (detect_from_serial), il secondo apre la lettura (do_download)
	if uvk5o(m).detect {
		if _, err := s.start(m); err != nil {
			return nil, nil, "", err
		}
	}
	fw, err := s.start(m)
	if err != nil {
		return nil, nil, "", err
	}
	var data []byte
	for a := 0; a < uvk5MemSize; a += uvk5Block {
		if ctx.Err() != nil {
			return nil, nil, "", errCancelled
		}
		prog(a, uvk5MemSize, "Lettura…")
		b, err := s.readMem(a, uvk5Block)
		if err != nil {
			return nil, nil, "", err
		}
		data = append(data, b...)
	}
	prog(uvk5MemSize, uvk5MemSize, "Lettura completata")
	return data, map[string]interface{}{"uvk5_firmware": fw}, "Firmware della radio: " + fw, nil
}

func (f uvk5Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < uvk5ProgSize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	p, err := openPort(port, 38400)
	if err != nil {
		return "", err
	}
	defer p.Close()
	s := &uvk5Sess{ctx: ctx, p: p}
	prog(0, uvk5ProgSize, "Collegamento con la radio…")
	fw, err := s.start(m)
	if err != nil {
		return "", err
	}
	for a := 0; a < uvk5ProgSize; a += uvk5Block {
		prog(a, uvk5ProgSize, "Scrittura…")
		if err := s.writeMem(a, img[a:a+uvk5Block]); err != nil {
			return "", err
		}
	}
	// riavvio della radio
	_ = s.cmd([]byte{0xdd, 0x05, 0x00, 0x00})
	prog(uvk5ProgSize, uvk5ProgSize, "Scrittura completata")
	return "Firmware della radio: " + fw, nil
}
