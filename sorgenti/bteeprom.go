// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Conversione tra il file .dat della Bajeton 8300 (oggetto KDH.RadioData) e
// la memoria della radio, ricavata riga per riga dal software originale
// BT-11_CPS (classe KDH.RWDataOperation, metodi Set*/Get*).
//
// Mappa della memoria (blocchi da 64 byte):
//   0x0000-0x77FF  960 canali da 32 byte (15 zone x 64)
//   0x8000         VFO A (32 byte) + VFO B (32 byte)
//   0x9000         impostazioni, parte 1 + parte 2
//   0xA000-0xA13F  DTMF (32 byte) + 18 gruppi di codici da 16 byte
//   0xB000-0xB0BF  radio FM / AM / SSB
//   0xC000-0xC2FF  nomi dei canali FM / AM / SSB
//
// Le poche differenze rispetto all'originale riguardano solo dati non validi,
// per i quali il programma originale si interrompe con un errore a metà della
// scrittura: qui l'errore viene segnalato prima di iniziare.

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const btImgSize = 0xC300

// btBlocks restituisce gli indirizzi dei blocchi da 64 byte nello stesso
// ordine del programma originale, compreso il blocco finale 0xC300 che il
// programma originale legge (e ignora) e scrive (tutto a 0xFF) prima del
// comando di fine.
func btBlocks() []int {
	var out []int
	for a := 0; a < btImgSize; a += 64 {
		switch {
		case a < 0x7800, a == 0x8000, a == 0x9000,
			a >= 0xA000 && a < 0xA140, a >= 0xB000 && a < 0xB0C0, a >= 0xC000:
			out = append(out, a)
		}
	}
	return append(out, btImgSize)
}

// ---------------- accesso agli oggetti NRBF per nome

type btErr struct{ error }

func btFail(f string, a ...interface{}) { panic(btErr{fmt.Errorf(f, a...)}) }

type nobj struct {
	d   *Doc
	r   *Rec
	idx map[string]int
	cls string
}

func (d *Doc) nobjOf(v Val, want string) *nobj {
	r := d.Resolve(v)
	if r == nil || (r.T != 1 && r.T != 5 && r.T != 4 && r.T != 3 && r.T != 2) {
		btFail("oggetto %s mancante nel file", want)
	}
	m := d.MetaOf(r)
	if m == nil || (want != "" && m.Name != want) {
		btFail("oggetto %s non trovato nel file", want)
	}
	o := &nobj{d: d, r: r, idx: map[string]int{}, cls: m.Name}
	for i, n := range m.Names {
		o.idx[n] = i
	}
	return o
}

func (o *nobj) field(f string) int {
	i, ok := o.idx[f]
	if !ok {
		btFail("campo %s.%s mancante: versione del file non supportata", o.cls, f)
	}
	return i
}

func (o *nobj) has(f string) bool { _, ok := o.idx[f]; return ok }

func (o *nobj) geti(f string) int {
	v := o.r.Vals[o.field(f)]
	if !v.Prim {
		btFail("campo %s.%s non numerico", o.cls, f)
	}
	switch len(v.Raw) {
	case 4:
		return int(int32(binary.LittleEndian.Uint32(v.Raw)))
	case 2:
		return int(int16(binary.LittleEndian.Uint16(v.Raw)))
	case 1:
		return int(v.Raw[0])
	}
	btFail("campo %s.%s di tipo non gestito", o.cls, f)
	return 0
}

func (o *nobj) seti(f string, val int) {
	i := o.field(f)
	v := &o.r.Vals[i]
	if !v.Prim {
		btFail("campo %s.%s non numerico", o.cls, f)
	}
	raw := make([]byte, len(v.Raw))
	switch len(v.Raw) {
	case 4:
		binary.LittleEndian.PutUint32(raw, uint32(int32(val)))
	case 2:
		binary.LittleEndian.PutUint16(raw, uint16(int16(val)))
	case 1:
		raw[0] = byte(val)
	default:
		btFail("campo %s.%s di tipo non gestito", o.cls, f)
	}
	v.Raw = raw
}

func (o *nobj) gets(f string) string { return o.d.StrOf(o.r.Vals[o.field(f)]) }

func (o *nobj) sets(f, s string) {
	i := o.field(f)
	if r := o.d.Resolve(o.r.Vals[i]); r != nil && r.T == 6 && r.S == s {
		return
	}
	o.r.Vals[i] = o.d.NewString(s)
}

func (o *nobj) obj(f, cls string) *nobj { return o.d.nobjOf(o.r.Vals[o.field(f)], cls) }

func (o *nobj) arr(f string) *Rec {
	r := o.d.Resolve(o.r.Vals[o.field(f)])
	if r == nil || (r.T != 7 && r.T != 16 && r.T != 17) {
		btFail("elenco %s.%s mancante nel file", o.cls, f)
	}
	return r
}

// strArr: elementi di un array di stringhe (i null multipli vengono espansi
// in null singoli, così l'indice corrisponde alla posizione)
func (d *Doc) strArrExpand(r *Rec) {
	var out []Val
	for _, e := range r.Els {
		if !e.Prim && e.R != nil && (e.R.T == 13 || e.R.T == 14) {
			for k := int32(0); k < e.R.N; k++ {
				out = append(out, Val{R: &Rec{T: 10}})
			}
			continue
		}
		out = append(out, e)
	}
	r.Els = out
}

// btRadio: le parti del file .dat usate dalla conversione
type btRadio struct {
	d        *Doc
	chans    []*nobj // 960, zona per zona
	vfo      [2]*nobj
	fun      *nobj
	dtmf     *nobj
	codes    *Rec // string[]
	mod      *nobj
	modChans []*nobj
	v109     bool // file della v1.0.9 (Channel[960], cbB_DivideCH)
}

func btOpen(d *Doc) *btRadio {
	root := d.nobjOf(Val{R: &Rec{T: 9, Ref: 1}}, "KDH.RadioData")
	br := &btRadio{d: d}
	els, flat, err := datChannelVals(d, root.arr("channelData"))
	if err != nil {
		btFail("%v", err)
	}
	br.v109 = flat
	for _, ce := range els {
		br.chans = append(br.chans, d.nobjOf(ce, "KDH.Channel"))
	}
	fm := root.obj("freqModeData", "KDH.FreqModeData")
	br.vfo[0] = fm.obj("vfoA", "KDH.VFOData")
	br.vfo[1] = fm.obj("vfoB", "KDH.VFOData")
	br.fun = root.obj("funConfigData", "KDH.FunConfigData")
	br.dtmf = root.obj("dtmfData", "KDH.DTMFData")
	br.codes = br.dtmf.arr("dtmfCodeGroup")
	if br.codes.T != 17 {
		btFail("gruppi DTMF: struttura non supportata")
	}
	d.strArrExpand(br.codes)
	if len(br.codes.Els) > 18 {
		btFail("troppi gruppi DTMF")
	}
	br.mod = root.obj("modulationData", "KDH.ModulationData")
	mc := br.mod.arr("modulationChannels")
	if mc.T != 7 || len(mc.Els) != 16 {
		btFail("canali FM/AM/SSB: struttura non supportata")
	}
	for _, e := range mc.Els {
		br.modChans = append(br.modChans, d.nobjOf(e, "KDH.ModulationChannel"))
	}
	return br
}

func (br *btRadio) code(i int) string { return br.d.StrOf(br.codes.Els[i]) }
func (br *btRadio) setCode(i int, s string) {
	if r := br.d.Resolve(br.codes.Els[i]); r != nil && r.T == 6 && r.S == s {
		return
	}
	br.codes.Els[i] = br.d.NewString(s)
}

// ---------------- GB2312 (in .NET è la code page 936)

func gbEncode(s string) []byte {
	enc := simplifiedchinese.GBK.NewEncoder()
	var out []byte
	for _, r := range s {
		b, err := enc.Bytes([]byte(string(r)))
		if err != nil || len(b) == 0 {
			b = []byte{'?'}
		}
		out = append(out, b...)
	}
	return out
}

func gbDecode(b []byte) string {
	s, err := simplifiedchinese.GBK.NewDecoder().Bytes(b)
	if err != nil {
		return strings.Map(func(r rune) rune {
			if r > 0x7e {
				return '?'
			}
			return r
		}, string(b))
	}
	return strings.ReplaceAll(string(s), "�", "?")
}

// ---------------- numeri decimali (equivalente di Convert.ToDouble + Math.Round)

// decUnits converte "145.26250" in unità da 10^-dec (arrotondamento al pari).
func decUnits(s string, dec int) (int64, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	ip, fp := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		ip, fp = s[:i], s[i+1:]
	}
	if ip == "" && fp == "" {
		return 0, false
	}
	for _, c := range ip + fp {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	if len(ip) > 12 {
		return 0, false
	}
	keep := fp
	rest := ""
	if len(keep) > dec {
		keep, rest = fp[:dec], fp[dec:]
	}
	for len(keep) < dec {
		keep += "0"
	}
	v, err := strconv.ParseInt("0"+ip+keep, 10, 64)
	if err != nil {
		return 0, false
	}
	if rest != "" {
		up := false
		switch {
		case rest[0] > '5':
			up = true
		case rest[0] == '5':
			if strings.Trim(rest[1:], "0") != "" {
				up = true
			} else {
				up = v%2 == 1
			}
		}
		if up {
			v++
		}
	}
	return v, true
}

// ---------------- canali

// btDCSIndex: posizione del codice nella tabella DCS della Bajeton (-1 se assente).
func btDCSIndex(s string) int { return idxOf(btDCS, s) }

var btDCS = func() []string {
	base := []string{"023", "025", "026", "031", "032", "036", "043", "047", "051", "053", "054", "065", "071", "072", "073", "074",
		"114", "115", "116", "122", "125", "131", "132", "134", "143", "145", "152", "155", "156", "162", "165", "172", "174",
		"205", "212", "223", "225", "226", "243", "244", "245", "246", "251", "252", "255", "261", "263", "265", "266", "271", "274",
		"306", "311", "315", "325", "331", "332", "343", "346", "351", "356", "364", "365", "371",
		"411", "412", "413", "423", "431", "432", "445", "446", "452", "454", "455", "462", "464", "465", "466",
		"503", "506", "516", "523", "526", "532", "546", "565",
		"606", "612", "624", "627", "631", "632", "645", "654", "662", "664",
		"703", "712", "723", "731", "732", "734", "743", "754"}
	var out []string
	for _, suf := range []string{"N", "I"} {
		for _, c := range base {
			out = append(out, "D"+c+suf)
		}
	}
	return out
}()

// SetFreq_StrToHex
func btEncFreq(s, what string) []byte {
	b := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	if s == "" {
		return b
	}
	v, ok := decUnits(s, 5)
	if !ok || v > 99999999 {
		btFail("%s \"%s\" non valida", what, s)
	}
	for i := 0; i < 4; i++ {
		r := v % 100
		v /= 100
		b[i] = byte(r/10<<4 | r%10)
	}
	return b
}

// GetFreq_HexToStr
func btDecFreq(b []byte) string {
	if allBytes(b, 0xFF) || allBytes(b, 0) {
		return ""
	}
	v := 0
	for i := 3; i >= 0; i-- {
		v = v*100 + int(b[i]>>4&15)*10 + int(b[i]&15)
	}
	return fmt.Sprintf("%d.%05d", v/100000, v%100000)
}

func allBytes(b []byte, v byte) bool {
	for _, x := range b {
		if x != v {
			return false
		}
	}
	return true
}

// SetSubaudio_StrToHex
func btEncQT(s, what string) [2]byte {
	if s == "" {
		btFail("%s vuoto", what)
	}
	if s[0] == 'D' {
		for i, d := range btDCS {
			if d == s {
				return [2]byte{byte(i + 1), 0}
			}
		}
		btFail("%s \"%s\": codice DCS sconosciuto", what, s)
	}
	if s == "OFF" {
		return [2]byte{0, 0}
	}
	if len(s) < 2 || s[len(s)-2] != '.' {
		btFail("%s \"%s\" non valido", what, s)
	}
	v, err := strconv.Atoi(strings.TrimSpace(s[:len(s)-2] + s[len(s)-1:]))
	if err != nil || v < 0 || v > 0xFFFF {
		btFail("%s \"%s\" non valido", what, s)
	}
	return [2]byte{byte(v), byte(v >> 8)}
}

// GetQTInfo_HexToStr / GetCTCSS_HexToStr / GetCDCSS_HexToStr
func btDecQT(b0, b1 byte) string {
	if b1 != 0 { // CTCSS
		if b0 == 0 || b0 == 255 {
			return "OFF"
		}
		s := strconv.Itoa(int(b1)<<8 + int(b0))
		return s[:len(s)-1] + "." + s[len(s)-1:]
	}
	if b0 != 0 && b0 <= 210 {
		return btDCS[b0-1]
	}
	return "OFF"
}

const hexDigits = "0123456789ABCDEF"

func btEncChannel(c *nobj, pos string) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = 0xFF
	}
	rx := c.gets("rxFreq")
	if rx == "" {
		return b
	}
	copy(b[0:], btEncFreq(rx, pos+": frequenza RX"))
	copy(b[4:], btEncFreq(c.gets("txFreq"), pos+": frequenza TX"))
	q := btEncQT(c.gets("rxQT"), pos+": tono RX")
	b[8], b[9] = q[0], q[1]
	q = btEncQT(c.gets("txQT"), pos+": tono TX")
	b[10], b[11] = q[0], q[1]
	b[12] = byte(c.geti("signallingGroup"))
	b[13] = byte(c.geti("pttId"))
	b[14] = byte(c.geti("scram")<<4 | c.geti("txPower"))
	b[15] = byte(c.geti("bandWide")<<6 | c.geti("encrypt")<<4 | c.geti("busyLockout")<<3 |
		c.geti("scanAdd")<<2 | c.geti("rxModulation")<<1 | c.geti("learnFHSS"))
	if f := c.gets("fhssCode"); f != "" {
		rs := []rune(f)
		if len(rs) < 6 {
			btFail("%s: codice FHSS \"%s\" troppo corto", pos, f)
		}
		ix := func(k int) int { return strings.IndexRune(hexDigits, rs[k]) }
		b[19] = 0xA0
		b[18] = byte(ix(0)<<4 | ix(1))
		b[17] = byte(ix(2)<<4 | ix(3))
		b[16] = byte(ix(4)<<4 | ix(5))
	}
	nb := gbEncode(c.gets("chName"))
	if len(nb) > 12 {
		nb = nb[:12]
	}
	copy(b[20:], nb)
	return b
}

// conteggio dei byte del nome come nel programma originale (stop a 0xFF,
// i byte >= 0xA1 valgono due)
func btNameLen(b []byte) int {
	n, i := 0, 0
	for i < 12 {
		if b[i] == 0xFF {
			break
		}
		if b[i] >= 0xA1 {
			n += 2
			i += 2
		} else {
			n++
			i++
		}
	}
	return n
}

func btDecChannel(c *nobj, b []byte) {
	if allBytes(b[0:4], 0xFF) || allBytes(b[0:4], 0) {
		btDefaultChannel(c)
		return
	}
	c.sets("rxFreq", btDecFreq(b[0:4]))
	c.sets("txFreq", btDecFreq(b[4:8]))
	c.sets("rxQT", btDecQT(b[8], b[9]))
	c.sets("txQT", btDecQT(b[10], b[11]))
	c.seti("signallingGroup", int(b[12]&15)%15)
	c.seti("pttId", int(b[13]&15)%4)
	c.seti("txPower", int(b[14]&15)%3)
	c.seti("scram", int(b[14]>>4)%2)
	c.seti("bandWide", int(b[15]&64>>6)%2)
	c.seti("encrypt", int(b[15]&48>>4)%4)
	c.seti("busyLockout", int(b[15]&8>>3)%2)
	c.seti("scanAdd", int(b[15]&4>>2)%2)
	c.seti("rxModulation", int(b[15]&2>>1)%2)
	c.seti("learnFHSS", int(b[15]&1)%2)
	f := ""
	if b[19] == 0xA0 {
		f = string([]byte{hexDigits[b[18]&0x70>>4], hexDigits[b[18]&15], hexDigits[b[17]>>4], hexDigits[b[17]&15],
			hexDigits[b[16]>>4], hexDigits[b[16]&15]})
	}
	c.sets("fhssCode", f)
	n := btNameLen(b[20:32])
	if n > 12 {
		n = 12
	}
	c.sets("chName", gbDecode(b[20:20+n]))
}

func btDefaultChannel(c *nobj) {
	c.sets("rxFreq", "")
	c.sets("txFreq", "")
	c.sets("rxQT", "OFF")
	c.sets("txQT", "OFF")
	for _, f := range []string{"signallingGroup", "pttId", "txPower", "scram", "bandWide", "encrypt",
		"busyLockout", "scanAdd", "rxModulation", "learnFHSS"} {
		c.seti(f, 0)
	}
	c.sets("fhssCode", "")
	c.sets("chName", "")
}

// ---------------- VFO

var btStep1 = []string{"2.5", "5", "6.25", "8.33", "10", "12.5", "20", "25", "50"} // ordine dei valori nel file
var btStep2 = []string{"2.5", "5", "6.25", "10", "12.5", "20", "25", "50", "8.33"} // ordine nella radio

func idxOf(l []string, s string) int {
	for i, x := range l {
		if x == s {
			return i
		}
	}
	return -1
}

// cifre decimali una per byte, completate a sinistra con zeri
func btDigits(b []byte, off int, s string, dec, width int, what string) {
	v, ok := decUnits(s, dec)
	if !ok {
		btFail("%s \"%s\" non valida", what, s)
	}
	p := 1
	for i := 0; i < dec; i++ {
		p *= 10
	}
	str := fmt.Sprintf("%d%0*d", v/int64(p), dec, v%int64(p))
	for len(str) < width {
		str = "0" + str
	}
	if off+len(str) > 32 {
		btFail("%s \"%s\" troppo grande", what, s)
	}
	for i, c := range str {
		b[off+i] = byte(c - '0')
	}
}

func btEncVFO(v *nobj, what string) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = 0xFF
	}
	btDigits(b, 0, v.gets("tB_RxFreq"), 5, 8, what+": frequenza")
	q := btEncQT(v.gets("cbB_RxQT"), what+": tono RX")
	b[8], b[9] = q[0], q[1]
	q = btEncQT(v.gets("cbB_TxQT"), what+": tono TX")
	b[10], b[11] = q[0], q[1]
	b[14] = byte(v.geti("cbB_OffsetDir")<<4 | v.geti("cbB_SignallingGroup"))
	b[16] = byte(v.geti("cbB_Scram")<<4 | v.geti("cbB_TxPower"))
	b[17] = byte(v.geti("cbB_BandWide")<<6 | v.geti("cbB_Encrypt")<<4 | v.geti("cbB_RxModulation")<<1)
	b[18] = byte(v.geti("cbB_FreqBand"))
	st := v.geti("cbB_StepFreq")
	if st < 0 || st >= len(btStep1) {
		btFail("%s: passo di sintonia non valido", what)
	}
	b[19] = byte(idxOf(btStep2, btStep1[st]))
	btDigits(b, 20, v.gets("tB_OffsetFreq"), 4, 7, what+": scostamento")
	return b
}

func decStr(b []byte) string {
	var sb strings.Builder
	for _, x := range b {
		sb.WriteString(strconv.Itoa(int(x)))
	}
	return sb.String()
}

func btDecVFO(v *nobj, b []byte) {
	f := "400.12500"
	if !allBytes(b[0:8], 0xFF) && !allBytes(b[0:8], 0) {
		n, err := strconv.ParseInt(decStr(b[0:8]), 10, 32)
		if err != nil || n < 0 {
			n = 40012500
		}
		f = fmt.Sprintf("%d.%05d", n/100000, n%100000)
	}
	v.sets("tB_RxFreq", f)
	v.sets("cbB_RxQT", btDecQT(b[8], b[9]))
	v.sets("cbB_TxQT", btDecQT(b[10], b[11]))
	v.seti("cbB_OffsetDir", int(b[14]&48>>4)%3)
	v.seti("cbB_SignallingGroup", int(b[14]&15)%15)
	v.seti("cbB_TxPower", int(b[16]&15)%3)
	v.seti("cbB_Scram", int(b[16]&240>>4)%2)
	v.seti("cbB_BandWide", int(b[17]&64>>6)%2)
	v.seti("cbB_Encrypt", int(b[17]&48>>4)%4)
	v.seti("cbB_RxModulation", int(b[17]&2>>1)%2)
	v.seti("cbB_FreqBand", int(b[18]&15)%2)
	v.seti("cbB_StepFreq", idxOf(btStep1, btStep2[int(b[19]&15)%9]))
	o := decStr(b[20:27])
	v.sets("tB_OffsetFreq", o[:len(o)-4]+"."+o[len(o)-4:])
	// campi riservati: valori di un RadioData nuovo
	for _, fr := range []string{"cbB_FrVFOByte12to13", "cbB_FrVFOByte15", "cbB_FrVFOByte27to31"} {
		v.seti(fr, 0)
	}
}

// ---------------- impostazioni

var btFun1 = []string{"cbB_SQL", "cbB_SaveMode", "cbB_VOX", "cbB_AutoBacklight", "cbB_TDR", "cbB_TOT", "cbB_BeepPrompt",
	"cbB_VoicePrompt", "cbB_Language", "cbB_DTMF", "cbB_Scan", "cbB_PTTID", "cbB_SendIDDelay", "cbB_DisplayModeA",
	"cbB_DisplayModeB", "cbB_BusyLockout", "cbB_AutoKeyLock", "cbB_AlarmMode", "cbB_AlarmSound", "cbB_TDRTx",
	"cbB_TailNoiseClear", "cbB_PassRepetNoiseClear", "cbB_PassRepetNoiseDetect", "cbB_SoundTxEnd", "cbB_CurWorkMode",
	"cbB_FMRadio", "", "cbB_LockKeyBoard", "cbB_PowerMsg", "cbB_BTRWData", "cbB_RTone"}
var btFun1Mod = []int{10, 4, 10, 9, 2, 6, 2, 2, 2, 4, 3, 4, 7, 3, 3, 2, 4, 3, 2, 3, 2, 11, 11, 3, 2, 2, 0, 2, 2, 2, 4}

type btPos struct {
	f        string
	pos, mod int
}

var btFun2 = []btPos{{"cbB_VoxDelay", 0, 16}, {"cbB_TimerMenuQuit", 1, 11}, {"cbB_RemoteKillEnable", 4, 2},
	{"cbB_WeatherCH", 7, 10}, {"cbB_KeySide2", 10, 8}, {"cbB_KeySide2L", 11, 7}, {"cbB_KeySide3", 12, 7},
	{"cbB_DenoiseSwitch", 17, 2}}

// riservati: cbB_FrTwoByte8to9 fino alla v1.0.8; dalla v1.0.9 il byte 8 è
// cbB_DivideCH (canali divisi in zone o tutti insieme) e resta cbB_FrTwoByte9
var btFunReserved = []string{"cbB_FrOneByte31", "cbB_FrTwoByte2to3", "cbB_FrTwoByte5to6", "cbB_FrTwoByte8to9", "cbB_FrTwoByte9",
	"cbB_FrTwoByte13", "cbB_FrTwoByte14to15", "cbB_FrTwoByte16", "cbB_FrTwoByte18to31"}

func btEncFun(f *nobj) ([]byte, []byte) {
	p1 := make([]byte, 32)
	for i, n := range btFun1 {
		if n != "" {
			p1[i] = byte(f.geti(n))
		}
	}
	p1[26] = byte(f.geti("cbB_WorkModeB")<<4 | f.geti("cbB_WorkModeA"))
	p1[31] = 0
	p2 := make([]byte, 32) // i byte non usati vengono scritti a 0
	for _, x := range btFun2 {
		p2[x.pos] = byte(f.geti(x.f))
	}
	if f.has("cbB_DivideCH") {
		p2[8] = byte(f.geti("cbB_DivideCH"))
	}
	return p1, p2
}

func btDecFun(f *nobj, p1, p2 []byte) {
	for i, n := range btFun1 {
		if n != "" {
			f.seti(n, int(p1[i]&15)%btFun1Mod[i])
		}
	}
	f.seti("cbB_WorkModeA", int(p1[26]&15)%2)
	f.seti("cbB_WorkModeB", int(p1[26]&240>>4)%2)
	for _, x := range btFun2 {
		f.seti(x.f, int(p2[x.pos]&15)%x.mod)
	}
	if f.has("cbB_DivideCH") {
		f.seti("cbB_DivideCH", int(p2[8]&15)%2)
	}
	for _, n := range btFunReserved {
		if f.has(n) {
			f.seti(n, 0)
		}
	}
}

// ---------------- DTMF

const btDTMFChars = "0123456789ABCD*#"

func btDTMFPut(b []byte, off int, s, what string) {
	for i, r := range []rune(s) {
		if off+i >= len(b) {
			btFail("%s \"%s\" troppo lungo", what, s)
		}
		b[off+i] = byte(strings.IndexRune(btDTMFChars, r))
	}
}

func btDTMFGet(b []byte, off, n int) string {
	if b[off] == 255 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		x := b[off+i]
		if x == 255 {
			break
		}
		if int(x) < len(btDTMFChars) {
			sb.WriteByte(btDTMFChars[x])
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}

func (br *btRadio) encDTMF() []byte {
	b := make([]byte, 320)
	for i := range b {
		b[i] = 0xFF
	}
	btDTMFPut(b[:32], 0, br.dtmf.gets("tB_DTMFCurId"), "codice DTMF personale")
	b[6] = byte(br.dtmf.geti("cbB_PTTID"))
	b[7] = byte(br.dtmf.geti("cbB_LastTimeSend"))
	b[8] = byte(br.dtmf.geti("cbB_LastTimeStop"))
	for i := range br.codes.Els {
		btDTMFPut(b[32:], i*16, br.code(i), fmt.Sprintf("gruppo DTMF %d", i+1))
	}
	return b
}

func (br *btRadio) decDTMF(b []byte) {
	br.dtmf.sets("tB_DTMFCurId", btDTMFGet(b, 0, 5))
	br.dtmf.seti("cbB_PTTID", int(b[6]&15)%4)
	br.dtmf.seti("cbB_LastTimeSend", int(b[7]&15)%5)
	br.dtmf.seti("cbB_LastTimeStop", int(b[8]&15)%5)
	for _, n := range []string{"cbB_FrByte5", "cbB_FrByte9to15", "cbB_FrByte16to23", "cbB_FrByte24to31"} {
		br.dtmf.seti(n, 0)
	}
	for i := range br.codes.Els {
		br.setCode(i, btDTMFGet(b[32:], i*16, 6))
	}
}

// ---------------- radio FM / AM / SSB

func put16(b []byte, off, v int) { b[off], b[off+1] = byte(v), byte(v>>8) }

func (br *btRadio) encMod() ([]byte, []byte) {
	b := make([]byte, 192)
	for i := range b {
		b[i] = 0xFF
	}
	mc := br.modChans
	for i := 0; i < 16; i++ {
		f := mc[i].geti("fmFreq")
		if i == 0 && f == 0 {
			f = 6400
		}
		put16(b, i*2, f)
	}
	b[32] = byte(br.mod.geti("cbB_FMCurChID"))
	b[33] = byte(br.mod.geti("cbB_WorkMode"))
	for i := 0; i < 16; i++ {
		f := mc[i].geti("amFreq")
		if i == 0 && f == 0 {
			f = 153
		}
		put16(b, 34+i*2, f)
	}
	b[66] = byte(br.mod.geti("cbB_AMCurChID"))
	b[67] = byte(br.mod.geti("cbB_ModulationMode"))
	b[68] = byte(br.mod.geti("cbB_AMRxGain"))
	for i := 0; i < 16; i++ {
		f := mc[i].geti("ssbFreq")
		if i == 0 && f == 0 {
			f = 150
		}
		put16(b, 69+i*5, f)
		b[71+i*5] = byte(mc[i].geti("ssbBandwidth"))
		o := mc[i].geti("ssbBeatFreqOffset")
		if o < -32760 || o > 27240 {
			o = 0
		}
		put16(b, 72+i*5, o)
	}
	b[149] = byte(br.mod.geti("cbB_SSBCurChID"))
	b[150] = byte(br.mod.geti("cbB_SSBStepFreq"))
	b[151] = byte(br.mod.geti("cbB_AMStepFreq"))
	b[152] = byte(br.mod.geti("cbB_SSBRxGain"))

	n := make([]byte, 768)
	for i := range n {
		n[i] = 0xFF
	}
	for i, c := range mc {
		for k, f := range []string{"fmName", "amName", "ssbName"} {
			nb := gbEncode(c.gets(f))
			if len(nb) > 12 {
				nb = nb[:12]
			}
			copy(n[k*256+i*16:], nb)
		}
	}
	return b, n
}

func get16(b []byte, off int) int {
	if (b[off] == 0 && b[off+1] == 0) || (b[off] == 255 && b[off+1] == 255) {
		return 0
	}
	return int(b[off+1])<<8 + int(b[off])
}

func (br *btRadio) decMod(b, n []byte) {
	mc := br.modChans
	for i := 0; i < 16; i++ {
		mc[i].seti("fmFreq", get16(b, i*2))
	}
	if mc[0].geti("fmFreq") == 0 {
		mc[0].seti("fmFreq", 640) // valore usato dal programma originale
	}
	br.mod.seti("cbB_FMCurChID", int(b[32])%15)
	br.mod.seti("cbB_WorkMode", int(b[33])%2)
	for i := 0; i < 16; i++ {
		mc[i].seti("amFreq", get16(b, 34+i*2))
	}
	if mc[0].geti("amFreq") == 0 {
		mc[0].seti("amFreq", 153)
	}
	br.mod.seti("cbB_AMCurChID", int(b[66])%15)
	br.mod.seti("cbB_ModulationMode", int(b[67])%5)
	br.mod.seti("cbB_AMRxGain", int(b[68])%37)
	for i := 0; i < 16; i++ {
		mc[i].seti("ssbFreq", get16(b, 69+i*5))
		mc[i].seti("ssbBandwidth", int(b[71+i*5])%6)
		o := int(int16(uint16(b[73+i*5])<<8 | uint16(b[72+i*5])))
		if o < -32760 || o > 27240 {
			o = 0
		}
		mc[i].seti("ssbBeatFreqOffset", o)
	}
	if mc[0].geti("ssbFreq") == 0 {
		mc[0].seti("ssbFreq", 150)
	}
	br.mod.seti("cbB_SSBCurChID", int(b[149])%15)
	br.mod.seti("cbB_SSBStepFreq", int(b[150])%6)
	if br.v109 {
		br.mod.seti("cbB_AMStepFreq", int(b[151])%4) // v1.0.9: quattro passi
	} else {
		br.mod.seti("cbB_AMStepFreq", int(b[151])%3)
	}
	br.mod.seti("cbB_SSBRxGain", int(b[152])%37)
	for i, c := range mc {
		for k, f := range []string{"fmName", "amName", "ssbName"} {
			off := k*256 + i*16
			c.sets(f, gbDecode(n[off:off+btNameLen(n[off:off+14])]))
		}
	}
}

// ---------------- conversioni complete

func btRecover(err *error) {
	if e := recover(); e != nil {
		if be, ok := e.(btErr); ok {
			*err = be.error
			return
		}
		*err = fmt.Errorf("errore interno: %v", e)
	}
}

// datToEEPROM prepara l'immagine della memoria da scrivere nella radio.
func datToEEPROM(dat []byte) (img []byte, err error) {
	defer btRecover(&err)
	d, err := ParseNRBF(dat)
	if err != nil {
		return nil, fmt.Errorf("file .dat non leggibile: %v", err)
	}
	br := btOpen(d)
	img = make([]byte, btImgSize)
	for i := range img {
		img[i] = 0xFF
	}
	for i, c := range br.chans {
		copy(img[i*32:], btEncChannel(c, fmt.Sprintf("Zona %d canale %d", i/datPerZone+1, i%datPerZone+1)))
	}
	copy(img[0x8000:], btEncVFO(br.vfo[0], "VFO A"))
	copy(img[0x8020:], btEncVFO(br.vfo[1], "VFO B"))
	p1, p2 := btEncFun(br.fun)
	copy(img[0x9000:], p1)
	copy(img[0x9020:], p2)
	copy(img[0xA000:], br.encDTMF())
	m, n := br.encMod()
	copy(img[0xB000:], m)
	copy(img[0xC000:], n)
	return img, nil
}

// eepromToDAT costruisce il file .dat dalla memoria letta, partendo da un
// file modello di cui vengono riscritti tutti i campi.
func eepromToDAT(img, tmpl []byte) (out []byte, err error) {
	defer btRecover(&err)
	if len(img) < btImgSize {
		return nil, fmt.Errorf("dati letti incompleti")
	}
	d, err := ParseNRBF(tmpl)
	if err != nil {
		return nil, fmt.Errorf("modello .dat non leggibile: %v", err)
	}
	old := map[int32]*Rec{}
	for k, v := range d.Objs {
		old[k] = v
	}
	br := btOpen(d)
	for i, c := range br.chans {
		btDecChannel(c, img[i*32:i*32+32])
	}
	btDecVFO(br.vfo[0], img[0x8000:0x8020])
	btDecVFO(br.vfo[1], img[0x8020:0x8040])
	btDecFun(br.fun, img[0x9000:0x9020], img[0x9020:0x9040])
	br.decDTMF(img[0xA000:0xA140])
	br.decMod(img[0xB000:0xB0C0], img[0xC000:0xC300])
	if err := d.FixDangling(old); err != nil {
		return nil, err
	}
	out = d.Dump()
	if _, err := ParseNRBF(out); err != nil {
		return nil, fmt.Errorf("errore interno, file non valido: %v", err)
	}
	return out, nil
}
