// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Formato .xlc del Baofeng K67 Pro (software "5RH PRO" CPS): copia della
// memoria della radio (128 KB; la radio ne usa i primi 0xC000 byte).
// Mappa ricavata dal software originale (classe DataProtocol):
//   0x0000  128 byte di informazioni sulla radio (nome, gamme RX 0x10 e TX 0x30,
//           versioni 0x50, data 0x60, funzioni 0x70)
//   0x0080  640 canali da 48 byte:
//           0-3 freq RX (BCD little-endian, unità 10 Hz), 4-7 freq TX,
//           8-9 tono RX, 10-11 tono TX (BCD big-endian in decimi di Hz;
//           bit7 del primo byte = DCS, bit6 = DCS invertito),
//           16 bit7-6 potenza (10 alta), bit5-4 banda (00 stretta),
//           bit3-2 direzione shift, bit1 inversione, bit0 talk-around,
//           17-20 segnalazioni, 24-30 passo e indici, 32-47 nome (16 byte)
//   0x7900  VFO, 0x7980 impostazioni
//   0x7A20  80 byte: un bit per canale, 0 = canale presente
//   0x7A80  zone (10 x 152 byte da 0x7A90)
//   0x8100  scansione; 0x81A0 80 byte: un bit per canale, 0 = in scansione
//   ...     DTMF, toni, MDC, emergenza, APRS, rubrica GPS
//   0x1FFFE abilitazione zone, 0x1FFFF banda di frequenza della radio
// Quando salva, il programma originale riscrive i record dei canali con la
// stessa normalizzazione applicata qui (byte non usati a zero, ecc.).

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	xlcSize     = 131072
	xlcChBase   = 0x80
	xlcChSize   = 48
	xlcChannels = 640
	xlcValid    = 0x7A20 // bit a 0 = canale presente
	xlcScanFlg  = 0x81A0 // bit a 0 = canale in scansione
	xlcZoneBase = 0x7A90
	xlcZoneSize = 152
	xlcZones    = 10
	xlcRadioLen = 0xC000 // byte trasferiti da e verso la radio
)

// isXLC: 128 KB con la versione del software della radio a 0x50 (es. "v1_0_35").
func isXLC(b []byte) bool {
	if len(b) != xlcSize {
		return false
	}
	if bytes.Contains(b[0x50:0x70], []byte("v1_")) {
		return true
	}
	v := b[0x50:0x58]
	return (v[0] == 'v' || v[0] == 'V') && v[1] >= '0' && v[1] <= '9'
}

func bcdLE(b []byte) (int64, bool) {
	var v int64
	for i := len(b) - 1; i >= 0; i-- {
		hi, lo := b[i]>>4, b[i]&0xf
		if hi > 9 || lo > 9 {
			return 0, false
		}
		v = v*100 + int64(hi)*10 + int64(lo)
	}
	return v, true
}

func bcdBE(b []byte) (int, bool) {
	v := 0
	for _, c := range b {
		hi, lo := c>>4, c&0xf
		if hi > 9 || lo > 9 {
			return 0, false
		}
		v = v*100 + int(hi)*10 + int(lo)
	}
	return v, true
}

func putBCDLE(dst []byte, v int64) {
	s := fmt.Sprintf("%0*d", len(dst)*2, v)
	for i := 0; i < len(dst); i++ {
		p := s[len(s)-2*(i+1):]
		dst[i] = (p[0]-'0')<<4 | (p[1] - '0')
	}
}

func putBCDBE(dst []byte, v int) {
	s := fmt.Sprintf("%0*d", len(dst)*2, v)
	for i := 0; i < len(dst); i++ {
		dst[i] = (s[2*i]-'0')<<4 | (s[2*i+1] - '0')
	}
}

type band struct {
	Lo int64 `json:"lo"`
	Hi int64 `json:"hi"`
}

func xlcBands(b []byte, off int) []band {
	var out []band
	for i := 0; i < 4; i++ {
		lo, ok1 := bcdLE(b[off+8*i : off+8*i+4])
		hi, ok2 := bcdLE(b[off+8*i+4 : off+8*i+8])
		if ok1 && ok2 && hi > lo {
			out = append(out, band{lo * 10, hi * 10})
		}
	}
	return out
}

func inBands(hz int64, bs []band) bool {
	for _, x := range bs {
		if hz >= x.Lo && hz <= x.Hi {
			return true
		}
	}
	return false
}

func bitGet(b []byte, off, i int) byte { return (b[off+i/8] >> (uint(i) % 8)) & 1 }
func bitSet(b []byte, off, i int, v byte) {
	if v != 0 {
		b[off+i/8] |= 1 << (uint(i) % 8)
	} else {
		b[off+i/8] &^= 1 << (uint(i) % 8)
	}
}

func xlcRaw(rec []byte) string { return "k67:" + hex.EncodeToString(rec) }

func xlcUnraw(s string) ([]byte, bool) {
	if !strings.HasPrefix(s, "k67:") {
		return nil, false
	}
	rec, err := hex.DecodeString(s[4:])
	if err != nil || len(rec) != xlcChSize {
		return nil, false
	}
	return rec, true
}

// ---------------------------------------------------------------- toni (SubVoiceConvert / SubAudioToData)

// xlcToneDecode: "" OFF, "88.5" CTCSS, "D023N"/"D023I" DCS; ok=false se il
// valore non è leggibile (il programma originale lo tratta come OFF).
func xlcToneDecode(h, l byte) (string, bool) {
	if h&0x80 == 0 {
		if h == 0 {
			return "", true
		}
		v, ok := bcdBE([]byte{h, l})
		if !ok || v == 0 {
			return "", false
		}
		return toneStr(v), true
	}
	if h == 0xFF {
		return "", true
	}
	code := int(h&7)<<8 | int(l)
	s := strconv.FormatInt(int64(code), 16)
	for len(s) < 3 {
		s = "0" + s
	}
	pol := "N"
	if h&0xC0 == 0xC0 {
		pol = "I"
	}
	return "D" + strings.ToUpper(s) + pol, true
}

// dcsParse: "D023N" -> codice e polarità; valido solo con tre cifre ottali.
func dcsParse(s string) (code string, inv bool, ok bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) != 5 || s[0] != 'D' || (s[4] != 'N' && s[4] != 'I') {
		return "", false, false
	}
	for _, c := range s[1:4] {
		if c < '0' || c > '7' {
			return "", false, false
		}
	}
	return s[1:4], s[4] == 'I', true
}

func xlcToneEncode(t string) ([2]byte, bool) {
	if t == "" {
		return [2]byte{0, 0}, true
	}
	if code, inv, ok := dcsParse(t); ok {
		v, _ := strconv.ParseUint(code, 16, 16)
		h := byte(v>>8) | 0x80
		if inv {
			h |= 0xC0
		}
		return [2]byte{h, byte(v)}, true
	}
	v, ok := parseTone(t)
	if !ok || v == 0 || v > 9999 {
		return [2]byte{0, 0}, false
	}
	var b [2]byte
	putBCDBE(b[:], v)
	return b, true
}

// ---------------------------------------------------------------- nomi (CharSwap2String / StringSwap2Char)

// xlcNameBytes: byte del nome fino al terminatore, letti a coppie come fa il
// programma originale; il resto del campo viene azzerato.
func xlcNameBytes(f []byte) []byte {
	var out []byte
	for i := 0; i+1 < len(f); i += 2 {
		if f[i] == 0 || f[i] == 0xFF {
			break
		}
		out = append(out, f[i], f[i+1])
	}
	return bytes.TrimRight(out, "\x00")
}

func xlcRecDecode(r []byte) (*Chan, string) {
	rx, ok1 := bcdLE(r[0:4])
	tx, ok2 := bcdLE(r[4:8])
	if !ok1 || rx == 0 {
		return nil, "frequenza non leggibile"
	}
	if !ok2 || tx == 0 {
		tx = rx
	}
	pw := 2
	if (r[16]>>6)&3 == 2 {
		pw = 0
	}
	c := &Chan{Rx: rx * 10, Tx: tx * 10, Power: pw, Narrow: r[16]&0x30 == 0}
	var okR, okT bool
	c.RxTone, okR = xlcToneDecode(r[8], r[9])
	c.TxTone, okT = xlcToneDecode(r[10], r[11])
	c.Name = asciiName(string(xlcNameBytes(r[32:48])))
	w := ""
	if !okR || !okT {
		w = "tono non leggibile, considerato OFF (come nel programma originale)"
	}
	return c, w
}

func xlcAllZero(r []byte) bool {
	for _, x := range r {
		if x != 0 {
			return false
		}
	}
	return true
}

func readXLC(b []byte, rep *Report) []*Chan {
	chs := make([]*Chan, xlcChannels)
	var stale []string
	defer func() {
		if len(stale) > 0 {
			rep.warn("K67: le posizioni %s contengono dati ma nel file sono segnate come vuote, quindi la radio e il programma originale le ignorano. "+
				"Se le avevi riempite con la versione 1.0 di questo programma, reinserisci quei canali.", strings.Join(stale, ", "))
		}
	}()
	for i := range chs {
		if bitGet(b, xlcValid, i) != 0 {
			// posizione vuota: eventuali dati residui sono ignorati, come nel programma originale
			if r := b[xlcChBase+xlcChSize*i : xlcChBase+xlcChSize*(i+1)]; !xlcAllZero(r) {
				if c, _ := xlcRecDecode(r); c != nil {
					stale = append(stale, fmt.Sprintf("%d (%s %s MHz)", i+1, c.Name, strconv.FormatFloat(float64(c.Rx)/1e6, 'f', 5, 64)))
				}
			}
			continue
		}
		r := b[xlcChBase+xlcChSize*i : xlcChBase+xlcChSize*(i+1)]
		c, w := xlcRecDecode(r)
		if c == nil {
			rep.warn("K67 canale %d: segnato come presente ma con %s: ignorato", i+1, w)
			continue
		}
		if w != "" {
			rep.warn("K67 canale %d (%s): %s", i+1, c.Name, w)
		}
		if (r[16]&0x0C != 0) && c.Tx == c.Rx {
			rep.warn("K67 canale %d (%s): TX uguale a RX ma il file indica uno shift %s: verificare se il canale deve avere lo shift", i+1, c.Name, map[bool]string{true: "+", false: "-"}[r[16]&0x04 != 0])
		}
		c.Scan = bitGet(b, xlcScanFlg, i) == 0
		c.Raw = xlcRaw(r)
		chs[i] = c
	}
	return chs
}

// xlcEncodeChan costruisce il record di un canale presente, partendo dai
// byte originali (se il canale proviene dallo stesso tipo di radio) e
// applicando la normalizzazione del programma originale (ChnConvert).
func xlcEncodeChan(c *Chan, pos string, rep *Report) []byte {
	rec := make([]byte, xlcChSize)
	raw, hadRaw := xlcUnraw(c.Raw)
	if hadRaw {
		copy(rec, raw)
	}
	var old *Chan
	if hadRaw {
		old, _ = xlcRecDecode(raw)
	}
	if old == nil || old.Rx != c.Rx || bcdBad(rec[0:4]) {
		putBCDLE(rec[0:4], c.Rx/10)
	}
	if old == nil || old.Tx != c.Tx || bcdBad(rec[4:8]) {
		putBCDLE(rec[4:8], c.Tx/10)
	}
	for k, t := range []string{c.RxTone, c.TxTone} {
		v, ok := xlcToneEncode(t)
		if !ok {
			rep.warn("%s: tono %s \"%s\" non gestito sul K67, impostato OFF", pos, []string{"RX", "TX"}[k], t)
		}
		rec[8+2*k], rec[9+2*k] = v[0], v[1]
	}
	// byte 12-15, 21-23 e 31 non usati: il programma originale li azzera
	for _, k := range []int{12, 13, 14, 15, 21, 22, 23, 31} {
		rec[k] = 0
	}
	rec[18] &= 0xE3
	rec[19] &= 0xE0
	rec[20] &= 0xF0
	f := rec[16] & 0x0F
	if c.Power == 1 {
		rep.warn("%s: potenza Media non disponibile sul K67, impostata Alta", pos)
	}
	if c.Power != 2 {
		f |= 0x80
	}
	if !c.Narrow {
		f |= 0x20
	}
	if old == nil || old.Rx != c.Rx || old.Tx != c.Tx {
		// direzione dello shift ricalcolata come fa il programma originale (DivFreqDisp)
		f &^= 0x0C
		if c.Tx > c.Rx {
			f |= 0x04
		} else if c.Tx < c.Rx {
			f |= 0x08
		}
	}
	rec[16] = f
	name := asciiName(c.Name)
	if len(name) > 16 {
		rep.warn("%s: nome accorciato a 16 caratteri", pos)
		name = name[:16]
	}
	var nb []byte
	if old != nil && old.Name == name {
		nb = xlcNameBytes(rec[32:48]) // nome invariato: stessi byte
	} else {
		nb = []byte(name)
	}
	for k := 32; k < 48; k++ {
		rec[k] = 0
	}
	copy(rec[32:48], nb)
	return rec
}

func bcdBad(b []byte) bool { _, ok := bcdLE(b); return !ok }

func writeXLC(base []byte, chs []*Chan, rep *Report) []byte {
	out := make([]byte, len(base))
	copy(out, base)
	rxb, txb := xlcBands(base, 0x10), xlcBands(base, 0x30)
	extra := 0
	for i, c := range chs {
		if i >= xlcChannels && c != nil {
			extra++
		}
	}
	if extra > 0 {
		rep.warn("%d canali oltre la posizione 640 non entrano nel K67 e sono stati esclusi", extra)
	}
	used := make([]bool, xlcChannels)
	for i := 0; i < xlcChannels; i++ {
		r := out[xlcChBase+xlcChSize*i : xlcChBase+xlcChSize*(i+1)]
		var c *Chan
		if i < len(chs) {
			c = chs[i]
		}
		if c != nil {
			pos := fmt.Sprintf("K67 canale %d (%s)", i+1, c.Name)
			if c.Rx <= 0 || c.Rx >= 1000000000 || c.Tx <= 0 || c.Tx >= 1000000000 {
				rep.warn("%s: frequenza fuori scala, canale non scritto", pos)
				c = nil
			} else {
				copy(r, xlcEncodeChan(c, pos, rep))
				used[i] = true
				bitSet(out, xlcScanFlg, i, map[bool]byte{true: 0, false: 1}[c.Scan])
				if len(rxb) > 0 && !inBands(c.Rx, rxb) {
					rep.warn("%s: %s MHz fuori dalle gamme di ricezione della radio", pos, fmtMHz(c.Rx))
				}
				if len(txb) > 0 && !inBands(c.Tx, txb) {
					rep.warn("%s: TX %s MHz fuori dalle gamme di trasmissione della radio", pos, fmtMHz(c.Tx))
				}
			}
		}
		if c == nil {
			for k := range r {
				r[k] = 0 // canale assente: record a zero, come nel programma originale
			}
		}
		bitSet(out, xlcValid, i, map[bool]byte{true: 0, false: 1}[used[i]])
	}
	baseUsed := make([]bool, xlcChannels)
	for i := range baseUsed {
		baseUsed[i] = bitGet(base, xlcValid, i) == 0
	}
	xlcZoneUpdate(out, baseUsed, used)
	return out
}

// xlcZoneUpdate aggiorna zone, canale corrente A/B e allarmi come fa il
// programma originale quando si crea o si cancella un canale
// (Frm_ZoneChn.NewChannelUpdateZone, UpDate_SettingData,
// UpDate_SettingChnData, UpDate_EmergData) e poi riscrive l'area delle zone
// come ZoneConvert.
func xlcZoneUpdate(out []byte, baseUsed, newUsed []bool) {
	zmc, zmn, zsize := 640, 1, 1304 // radio senza zone (K67 Pro)
	if (out[115]>>2)&1 == 1 {
		zmc, zmn, zsize = 64, 10, 152
	}
	chnNum := make([]int, 10)
	ids := make([][]int, zmn)
	names := make([][]byte, zmn)
	for z := 0; z < zmn; z++ {
		rec := out[xlcZoneBase+zsize*z:]
		n := int(rec[0])
		if n > zmc {
			n = 0
		}
		chnNum[z] = n
		ids[z] = make([]int, zmc)
		for j := 0; j < zmc; j++ {
			ids[z][j] = int(rec[2+2*j])<<8 | int(rec[3+2*j])
		}
		names[z] = xlcNameBytes(rec[136:152])
		if n > 0 && len(names[z]) == 0 {
			names[z] = []byte(fmt.Sprintf("Zone %d", z+1))
		}
	}
	zoneTotal := int(out[0x7A80])
	const gs = 0x7980
	chANum := int(out[gs+2])<<8 | int(out[gs+3])
	chBNum := int(out[gs+4])<<8 | int(out[gs+5])
	chAZone, chBZone := int(out[gs+6]), int(out[gs+7])
	const em = 40200
	firstZone := func() int {
		for i := 0; i < 10; i++ {
			if chnNum[i] > 0 {
				return i
			}
		}
		return 0
	}
	usedNow := append([]bool{}, baseUsed...)
	// canali nuovi
	for s := 0; s < xlcChannels; s++ {
		if baseUsed[s] || !newUsed[s] {
			continue
		}
		z, row := s/zmc, s%zmc
		if chnNum[z] == 0 {
			if len(names[z]) == 0 {
				names[z] = []byte(fmt.Sprintf("Zone %d", z+1))
			}
			zoneTotal++
		}
		ids[z][row] = s
		chnNum[z] = (chnNum[z] + 1) & 0xFFFF
		usedNow[s] = true
	}
	// canali cancellati
	for s := 0; s < xlcChannels; s++ {
		if !baseUsed[s] || newUsed[s] {
			continue
		}
		usedNow[s] = false
		z, row := s/zmc, s%zmc
		ids[z][row] = 0xFFFF
		chnNum[z] = (chnNum[z] - 1) & 0xFFFF
		if chnNum[z] == 0 {
			zoneTotal--
			if chAZone == z {
				chAZone = firstZone()
			}
			if chBZone == z {
				chBZone = firstZone()
			}
			nz := firstZone()
			for k := 0; k < 10; k++ {
				e := out[em+16*k:]
				if int(e[15]) == z {
					e[1], e[15] = 1, byte(nz)
				}
			}
		} else {
			for k := 0; k < 10; k++ {
				e := out[em+16*k:]
				if int(e[15]) == z && int(e[14]) == row {
					e[1] = 1
				}
			}
		}
		for _, p := range []*struct{ num, zone *int }{{&chANum, &chAZone}, {&chBNum, &chBZone}} {
			if *p.num != row {
				continue
			}
			b := *p.zone * zmc
			for i := b; i < b+64; i++ {
				if i < xlcChannels && usedNow[i] {
					*p.num = (i - b) & 0xFF
					break
				}
			}
		}
	}
	out[0x7A80] = byte(zoneTotal)
	out[gs+2], out[gs+3] = byte(chANum>>8), byte(chANum)
	out[gs+4], out[gs+5] = byte(chBNum>>8), byte(chBNum)
	out[gs+6], out[gs+7] = byte(chAZone), byte(chBZone)
	for z := 0; z < zmn; z++ {
		rec := out[xlcZoneBase+zsize*z:]
		rec[0] = byte(chnNum[z])
		for j := 0; j < zmc; j++ {
			v := 0xFFFF
			if s := z*zmc + j; s < xlcChannels && newUsed[s] {
				v = ids[z][j]
			}
			rec[2+2*j], rec[3+2*j] = byte(v>>8), byte(v)
		}
		for k := 136; k < 152; k++ {
			rec[k] = 0
		}
		copy(rec[136:152], names[z])
	}
}

// blankXLC: file K67 con le impostazioni del file base ma senza canali.
func blankXLC(base []byte) []byte {
	return writeXLC(base, nil, &Report{})
}

// xlcFreshZones riproduce ciò che fa il programma originale appena avviato
// con una radio senza zone (come il K67 Pro): legge la zona con i parametri
// di una radio a zone (64 canali), quindi al salvataggio le voci dalla 65 in
// poi dei canali presenti diventano 0 e un contatore oltre 64 diventa 0.
// Così i byte inviati alla radio sono gli stessi che invierebbe il
// programma originale aprendo lo stesso file.
func xlcFreshZones(b []byte) {
	if (b[115]>>2)&1 == 1 {
		return // radio a zone: il programma originale le legge correttamente
	}
	z := b[xlcZoneBase:]
	if z[0] > 64 {
		z[0] = 0
	}
	for j := 64; j < xlcChannels; j++ {
		if bitGet(b, xlcValid, j) == 0 {
			z[2+2*j], z[3+2*j] = 0, 0
		} else {
			z[2+2*j], z[3+2*j] = 0xFF, 0xFF
		}
	}
	// il nome della zona copre le voci 67-74 (ZoneConvert lo scrive per ultimo)
	name := xlcNameBytes(z[136:152])
	if z[0] > 0 && len(name) == 0 {
		name = []byte("Zone 1")
	}
	for k := 136; k < 152; k++ {
		z[k] = 0
	}
	copy(z[136:152], name)
}

// xlcImageForRadio: i byte da inviare alla radio (primi 0xC000).
func xlcImageForRadio(file []byte) []byte {
	img := make([]byte, xlcRadioLen)
	copy(img, file[:xlcRadioLen])
	xlcFreshZones(img)
	return img
}

// xlcFileFromRadio costruisce il file .xlc dai dati letti dalla radio come lo
// salverebbe il programma originale: data del salvataggio nelle informazioni,
// zone come sopra, abilitazione zone e banda negli ultimi due byte.
func xlcFileFromRadio(img []byte, freqBand byte, k6 bool, now time.Time) []byte {
	out := make([]byte, xlcSize)
	copy(out, img[:xlcRadioLen])
	copy(out[96:112], now.Format("2006.01.02 15:04"))
	xlcFreshZones(out)
	enZone := (img[115] >> 2) & 1
	if k6 {
		enZone = 0
	}
	out[xlcSize-2] = enZone
	out[xlcSize-1] = freqBand
	return out
}
