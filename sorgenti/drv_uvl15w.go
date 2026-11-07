// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tyt_uvl15w.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: TYT UVL-15W (tyt_uvl15w.py di CHIRP, dal protocollo pubblicato da TYT).
// 115200 baud, pacchetti FE FE EE EF comando dati+LRC (ogni byte +0x80, i valori
// oltre 0xF9 come FF + 4 bit) FD. Si legge e si scrive la zona dati della flash
// 0x8000-0x20FFF (100 KiB): 1000 canali da 48 byte (numerati da 0) con nomi di
// 16 caratteri, bit «valido», 2 bit di scansione per canale, zone e liste di
// scansione (due versioni della memoria secondo il firmware: 3.03 e 3.09).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type uvl15wFam struct{}

const uvl15wFormat = `
struct channel {
  u32 rxfreq;
  u32 txfreq;
  char name[24];
  u32 offset;
  u8 step:4,
     reverse:2,
     duplex:2;
  u8 bcl:2,
     rxonly:1,
     power:2,
     mode:3;
  u8 txtone_type:3,
     sql_type:3,
     unknown26:2;
  u8 unknown27:1,
     compander:2,
     dcs_pol:2,
     rxtone_type:3;
  u8 txtone;
  u8 rxtone;
  u8 sig_type;
  u8 sig_index;
  u8 scramble;
  u8 pttid;
  u8 aprs_rx;
  u8 reserved;
};
#seekto 0x0000;
struct channel memory[1000];
#seekto 0xBB80;
struct channel vfo[2];
#seekto 0xBC40;
struct channel temp[2];
`

const (
	uvl15wFlashStart = 0x8000
	uvl15wFlashEnd   = 0x21000
	uvl15wSize       = uvl15wFlashEnd - uvl15wFlashStart
	uvl15wValid      = 0xD000
	uvl15wScan       = 0xD080
)

// zone e liste di scansione (GROUP_FORMATS): posizione delle liste dei canali
// (128 posizioni da 16 bit), delle mappe per canale e numero dei gruppi
type uvl15wGroups struct {
	zones, scans  int  // liste dei canali delle zone e delle liste di scansione
	chZones, chSc int  // mappe (bit a 0 = il canale è nel gruppo)
	n             int  // gruppi
	be            bool // liste in big endian (versione 2)
	mapW          int  // byte per mappa (2 o 4)
}

var uvl15wLayouts = map[int]uvl15wGroups{
	1: {zones: 0x10000, scans: 0x11000, chZones: 0x12000, chSc: 0x12800, n: 16, mapW: 2},
	2: {zones: 0x10000, scans: 0x12000, chZones: 0x14000, chSc: 0x15000, n: 32, be: true, mapW: 4},
}

var uvl15wTones = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230, 1273,
	1318, 1365, 1413, 1462, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835, 1862, 1899, 1928, 1966, 1995,
	2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

var uvl15wDCS = []int{23, 25, 26, 31, 32, 36, 43, 47, 51, 53, 54, 65, 71, 72, 73, 74, 114, 115, 116, 122, 125, 131, 132, 134, 143,
	145, 152, 155, 156, 162, 165, 172, 174, 205, 212, 223, 225, 226, 243, 244, 245, 246, 251, 252, 255, 261, 263, 265, 266, 271,
	274, 306, 311, 315, 325, 331, 332, 343, 346, 351, 356, 364, 365, 371, 411, 412, 413, 423, 431, 432, 445, 446, 452, 454, 455,
	462, 464, 465, 466, 503, 506, 516, 523, 526, 532, 546, 565, 606, 612, 624, 627, 631, 632, 654, 662, 664, 703, 712, 723, 731,
	732, 734, 743, 754}

func init() {
	genRegister(&genModel{ID: "tyt-uvl15w", Vendor: "TYT", Model: "UVL-15W", RClass: "UVL15WRadio", Name: "TYT UVL-15W",
		Slots: 1000, First: 0, NameMax: 16, Charset: charsetASCII,
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 5 W", 5, 1}, {"Alta 15 W", 15, 0}},
		Bands:  []band{{108000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true,
		DCS: uvl15wDCS, Baud: 115200, Fam: uvl15wFam{},
		Notes: "1000 canali (da 0 a 999) con nomi di 16 caratteri. Accendi la radio con il cavo collegato: alla fine di ogni lettura " +
			"e scrittura la radio si riavvia. Se sulla radio c'è una password di lettura o scrittura, toglila prima. Si scrive solo " +
			"un'immagine letta dalla radio (contiene tutta la zona dati). Modulazione AM dei canali che non cambiano frequenza, passo, " +
			"canale prioritario in scansione, BCL, scrambler, compander, PTT ID e APRS si conservano; un canale cancellato esce " +
			"dalle zone e dalle liste di scansione. Driver di CHIRP (sperimentale)."})
}

func (uvl15wFam) mem() *bwEl                 { return bwMust(uvl15wFormat) }
func (f uvl15wFam) layout(m *genModel) *bwEl { return f.mem() }
func (f uvl15wFam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f uvl15wFam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), uvl15wBit(img, i), uvl15wSkip(img, i))
}
func (f uvl15wFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(i).SetRaw(img, raw[:48])
	if len(raw) >= 50 {
		uvl15wSetBit(img, i, raw[48] != 0)
		uvl15wSetSkip(img, i, int(raw[49]))
	}
}
func (f uvl15wFam) Match(m *genModel, img []byte) bool { return false }
func (uvl15wFam) crossband() bool                      { return false }

func uvl15wBit(img []byte, i int) byte { return (img[uvl15wValid+i/8] >> (i % 8)) & 1 }
func uvl15wSetBit(img []byte, i int, on bool) {
	if on {
		img[uvl15wValid+i/8] |= 1 << (i % 8)
	} else {
		img[uvl15wValid+i/8] &^= 1 << (i % 8)
	}
}
func uvl15wSkip(img []byte, i int) byte { return (img[uvl15wScan+i/4] >> (2 * (i % 4))) & 3 }
func uvl15wSetSkip(img []byte, i int, v int) {
	sh := 2 * (i % 4)
	img[uvl15wScan+i/4] = img[uvl15wScan+i/4]&^(3<<sh) | byte(v<<sh)
}

// uvl15wLayout: versione della memoria (_detect_layout): 0 sconosciuta
func uvl15wLayout(img []byte) int {
	if len(img) < 0x16B05 {
		return 0
	}
	switch {
	case bytes.Equal(img[0x13B00:0x13B05], []byte("EDG1\x01")):
		return 1
	case bytes.Equal(img[0x16B00:0x16B05], []byte("EDG1\x02")):
		return 2
	case bytes.Equal(img[0x13B00:0x13B05], bytes.Repeat([]byte{0xFF}, 5)):
		return 1
	}
	return 0
}

// mappa dei gruppi del canale ch (bit invertiti) e lista dei canali di un gruppo
func (g uvl15wGroups) chMap(img []byte, base, ch int) uint32 {
	off := base + ch*g.mapW
	if g.mapW == 2 {
		return uint32(binary.LittleEndian.Uint16(img[off:]))
	}
	return binary.LittleEndian.Uint32(img[off:])
}

func (g uvl15wGroups) setChMap(img []byte, base, ch int, v uint32) {
	off := base + ch*g.mapW
	if g.mapW == 2 {
		binary.LittleEndian.PutUint16(img[off:], uint16(v))
	} else {
		binary.LittleEndian.PutUint32(img[off:], v)
	}
}

func (g uvl15wGroups) slot(img []byte, base, grp, k int) uint16 {
	off := base + grp*256 + k*2
	if g.be {
		return binary.BigEndian.Uint16(img[off:])
	}
	return binary.LittleEndian.Uint16(img[off:])
}

func (g uvl15wGroups) setSlot(img []byte, base, grp, k int, v uint16) {
	off := base + grp*256 + k*2
	if g.be {
		binary.BigEndian.PutUint16(img[off:], v)
	} else {
		binary.LittleEndian.PutUint16(img[off:], v)
	}
}

// clearGroups: _clear_groups (il canale esce da zone e liste di scansione)
func uvl15wClearGroups(img []byte, ch int) {
	g, ok := uvl15wLayouts[uvl15wLayout(img)]
	if !ok {
		return
	}
	for _, kind := range [][2]int{{g.zones, g.chZones}, {g.scans, g.chSc}} {
		bits := ^g.chMap(img, kind[1], ch)
		for grp := 0; grp < g.n; grp++ {
			if bits&(1<<grp) == 0 {
				continue
			}
			var keep []uint16
			for k := 0; k < 128; k++ {
				if s := g.slot(img, kind[0], grp, k); int(s) < 1000 && int(s) != ch {
					keep = append(keep, s)
				}
			}
			for k := 0; k < 128; k++ {
				v := uint16(0xFFFF)
				if k < len(keep) {
					v = keep[k]
				}
				g.setSlot(img, kind[0], grp, k, v)
			}
			g.setChMap(img, kind[1], ch, g.chMap(img, kind[1], ch)|1<<grp)
		}
	}
}

// _decode_name: fino al primo 0, UTF-8; 0xFF iniziale = nome vuoto
func uvl15wName(raw []byte) string {
	if k := bytes.IndexByte(raw, 0); k >= 0 {
		raw = raw[:k]
	}
	if len(raw) > 0 && raw[0] == 0xFF {
		return ""
	}
	var b strings.Builder
	for len(raw) > 0 {
		r, n := utf8.DecodeRune(raw)
		b.WriteRune(r)
		raw = raw[n:]
	}
	return strings.TrimRightFunc(b.String(), func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f' })
}

var uvl15wPol = []string{"NN", "NR", "RN", "RR"}

func uvl15wToneDec(kind, idx int64, pol byte) string {
	switch {
	case (kind == 1 || kind == 3) && int(idx) < len(uvl15wTones):
		return fmtTone(uvl15wTones[idx])
	case (kind == 2 || kind == 4) && int(idx) < len(uvl15wDCS):
		return genTone{Mode: "DTCS", Code: uvl15wDCS[idx], Inv: pol == 'R'}.String()
	}
	return ""
}

func (f uvl15wFam) Get(m *genModel, img []byte, i int) *Chan {
	if uvl15wBit(img, i) == 0 {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("rxfreq").Int(img)}
	tx := r.F("txfreq").Int(img)
	switch d := r.F("duplex").Int(img); {
	case r.F("rxonly").Int(img) > 0:
		c.TxOff, c.Tx = true, c.Rx
	case d == 1:
		c.Tx = c.Rx - r.F("offset").Int(img)
	case d == 2:
		c.Tx = c.Rx + r.F("offset").Int(img)
	default:
		c.Tx = tx
	}
	c.Name = uvl15wName(r.F("name").Raw(img))
	md := r.F("mode").Int(img)
	c.Narrow = md == 1 || md == 3
	c.Power = m.powerSem(int(min(r.F("power").Int(img), 2)))
	pol := uvl15wPol[r.F("dcs_pol").Int(img)]
	c.TxTone = uvl15wToneDec(r.F("txtone_type").Int(img), r.F("txtone").Int(img), pol[0])
	if sq := r.F("sql_type").Int(img); sq == 1 || sq == 3 || sq == 4 {
		c.RxTone = uvl15wToneDec(r.F("rxtone_type").Int(img), r.F("rxtone").Int(img), pol[1])
	}
	c.Scan = uvl15wSkip(img, i) != 1
	return c
}

func uvl15wToneEnc(t genTone) (int64, int64) {
	switch t.Mode {
	case "Tone":
		if k := idxInt(uvl15wTones, t.Hz10); k >= 0 {
			return 1, int64(k)
		}
	case "DTCS":
		if k := idxInt(uvl15wDCS, t.Code); k >= 0 {
			return 2, int64(k)
		}
	}
	return 0, 0
}

func (f uvl15wFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		uvl15wSetBit(img, i, false)
		r.Fill(img, 0)
		uvl15wSetSkip(img, i, 3)
		uvl15wClearGroups(img, i)
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
	for _, t := range []genTone{tt, rt} {
		if t.Mode == "Tone" && idxInt(uvl15wTones, t.Hz10) < 0 {
			return fmt.Errorf("subtono %s non disponibile su questa radio", fmtTone(t.Hz10))
		}
		if t.Mode == "DTCS" && idxInt(uvl15wDCS, t.Code) < 0 {
			return fmt.Errorf("codice DCS D%03d non disponibile su questa radio", t.Code)
		}
	}
	valid := uvl15wBit(img, i) != 0
	// valori conservati (get_memory e i suoi extra)
	var bcl, rev, comp, scr, pttid, aprs int64
	step := int64(2) // canale nuovo: 5 kHz (passo predefinito di CHIRP)
	am := false
	oldSkip := uvl15wSkip(img, i)
	if valid {
		bcl, rev, comp = min(r.F("bcl").Int(img), 2), min(r.F("reverse").Int(img), 2), r.F("compander").Int(img)
		scr = min(r.F("scramble").Int(img), 8)
		pttid, aprs = min(r.F("pttid").Int(img)&0x0F, 8), min(r.F("aprs_rx").Int(img)&3, 2)
		if s := r.F("step").Int(img); s < 12 {
			step = s
		}
		md := r.F("mode").Int(img)
		am = (md == 2 || md == 3) && r.F("rxfreq").Int(img) == c.Rx
	} else {
		r.Fill(img, 0)
		uvl15wClearGroups(img, i)
		r.F("step").SetInt(img, 6)
		r.F("power").SetInt(img, 2)
		r.F("offset").SetInt(img, 600000)
	}
	if c.Rx >= 108000000 && c.Rx < 137000000 {
		am = true // gamma aeronautica in AM
	}
	uvl15wSetBit(img, i, true)
	nm := []byte(c.Name)
	if len(nm) > 24 {
		nm = nm[:24]
	}
	r.F("name").SetRaw(img, append(nm, make([]byte, 24-len(nm))...))
	r.F("rxfreq").SetInt(img, c.Rx)
	r.F("rxonly").SetBool(img, c.TxOff)
	switch {
	case c.TxOff || c.Tx == c.Rx:
		r.F("duplex").SetInt(img, 0)
		r.F("txfreq").SetInt(img, c.Rx)
	case abs64(c.Tx-c.Rx) > 70000000:
		r.F("duplex").SetInt(img, 3)
		r.F("txfreq").SetInt(img, c.Tx)
		r.F("offset").SetInt(img, abs64(c.Tx-c.Rx))
	case c.Tx < c.Rx:
		r.F("duplex").SetInt(img, 1)
		r.F("offset").SetInt(img, c.Rx-c.Tx)
		r.F("txfreq").SetInt(img, c.Tx)
	default:
		r.F("duplex").SetInt(img, 2)
		r.F("offset").SetInt(img, c.Tx-c.Rx)
		r.F("txfreq").SetInt(img, c.Tx)
	}
	md := int64(0)
	if c.Narrow {
		md = 1
	}
	if am {
		md += 2
	}
	r.F("mode").SetInt(img, md)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("step").SetInt(img, step)
	ty, v := uvl15wToneEnc(tt)
	r.F("txtone_type").SetInt(img, ty)
	r.F("txtone").SetInt(img, v)
	ry, rv := uvl15wToneEnc(rt)
	r.F("rxtone_type").SetInt(img, ry)
	r.F("rxtone").SetInt(img, rv)
	pol := "N"
	if tt.Mode == "DTCS" && tt.Inv {
		pol = "R"
	}
	if rt.Mode == "DTCS" && rt.Inv {
		pol += "R"
	} else {
		pol += "N"
	}
	r.F("dcs_pol").SetInt(img, int64(idxStr(uvl15wPol, pol)))
	sq := r.F("sql_type").Int(img)
	withTone := sq == 1 || sq == 3 || sq == 4
	if (tt.Mode != "" || rt.Mode != "") && !withTone {
		r.F("sql_type").SetInt(img, 1)
	} else if tt.Mode == "" && rt.Mode == "" && withTone {
		r.F("sql_type").SetInt(img, 0)
	}
	sk := 0
	switch {
	case !c.Scan:
		sk = 1
	case oldSkip == 2 && valid:
		sk = 2 // canale prioritario («P») conservato
	}
	uvl15wSetSkip(img, i, sk)
	if valid {
		r.F("bcl").SetInt(img, bcl)
		r.F("reverse").SetInt(img, rev)
		r.F("compander").SetInt(img, comp)
		r.F("scramble").SetInt(img, scr)
		r.F("pttid").SetInt(img, r.F("pttid").Int(img)&0xF0|pttid)
		r.F("aprs_rx").SetInt(img, r.F("aprs_rx").Int(img)&0xFC|aprs)
	}
	return nil
}

// trimLeftByte: toglie i byte b iniziali (bytes.TrimLeft lavora sui caratteri
// UTF-8 e con un byte non valido come 0xFE toglierebbe anche altri byte)
func trimLeftByte(d []byte, b byte) []byte {
	for len(d) > 0 && d[0] == b {
		d = d[1:]
	}
	return d
}

func idxStr(l []string, v string) int {
	for k, x := range l {
		if x == v {
			return k
		}
	}
	return 0
}

// ---------------------------------------------------------------- protocollo

const (
	uvl15wCmdInfo       = 0xE0
	uvl15wCmdInfoResp   = 0xE1
	uvl15wCmdReadStart  = 0xE2
	uvl15wCmdWriteStart = 0xE3
	uvl15wCmdWrite      = 0xE4
	uvl15wCmdFinish     = 0xE5
	uvl15wCmdRead       = 0xE6
	uvl15wCmdError      = 0xEE
)

func uvl15wEscape(d []byte) []byte {
	var out []byte
	for _, b := range d {
		t := b + 0x80
		if t > 0xF9 {
			out = append(out, 0xFF, t&0x0F)
		} else {
			out = append(out, t)
		}
	}
	return out
}

func uvl15wUnescape(d []byte) ([]byte, error) {
	var out []byte
	for k := 0; k < len(d); k++ {
		if d[k] == 0xFF {
			if k+1 >= len(d) {
				return nil, rErr("risposta della radio troncata")
			}
			out = append(out, 0xF0+d[k+1]+0x80)
			k++
		} else {
			out = append(out, d[k]+0x80)
		}
	}
	return out, nil
}

func uvl15wLRC(d []byte) byte { return byte(0x100 - int(sum8(d))) }

func uvl15wFrame(cmd byte, data []byte) []byte {
	out := []byte{0xFE, 0xFE, 0xEE, 0xEF, cmd}
	out = append(out, uvl15wEscape(append(append([]byte{}, data...), uvl15wLRC(data)))...)
	return append(out, 0xFD)
}

func (s *rawSess) uvl15wRecv(to time.Duration) (byte, []byte, error) {
	var fr []byte
	for k := 0; k < 2*(0x800+8)+64; k++ {
		b := readUpTo(s.ctx, s.p, 1, to)
		if s.ctx.Err() != nil {
			return 0, nil, errCancelled
		}
		if len(b) == 0 {
			if fr == nil {
				return 0, nil, errTimeout
			}
			return 0, nil, rErr("risposta della radio incompleta")
		}
		if fr == nil && b[0] != 0xFE {
			continue
		}
		fr = append(fr, b[0])
		if b[0] == 0xFD {
			fr = trimLeftByte(fr, 0xFE)
			if len(fr) < 5 || fr[0] != 0xEF || fr[1] != 0xEE {
				return 0, nil, rErr("risposta della radio non valida")
			}
			pl, err := uvl15wUnescape(fr[3 : len(fr)-1])
			if err != nil {
				return 0, nil, err
			}
			if len(pl) == 0 || sum8(pl) != 0 {
				return 0, nil, rErr("somma di controllo errata nella risposta")
			}
			return fr[2], pl[:len(pl)-1], nil
		}
	}
	return 0, nil, rErr("risposta della radio troppo lunga")
}

// uvl15wCmd: _command (3 tentativi per gli errori di trasmissione)
func (s *rawSess) uvl15wCmd(cmd byte, data []byte, expect int, to time.Duration) ([]byte, error) {
	for k := 0; k < 3; k++ {
		if err := s.send(uvl15wFrame(cmd, data)); err != nil {
			return nil, err
		}
		rc, rd, err := s.uvl15wRecv(to)
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			continue
		}
		if rc == uvl15wCmdError {
			msg := string(rd)
			if strings.Contains(msg, "Lrc") || strings.Contains(msg, "Frame") {
				continue
			}
			return nil, rErr("la radio ha segnalato un errore: %s", msg)
		}
		if expect >= 0 && int(rc) != expect {
			return nil, rErr("risposta %02X inattesa al comando %02X", rc, cmd)
		}
		return rd, nil
	}
	return nil, rErr("la radio non ha accettato il comando %02X", cmd)
}

func (s *rawSess) uvl15wIdent(write bool) error {
	_, err := s.uvl15wInfo(write)
	return err
}

func (s *rawSess) uvl15wInfo(write bool) ([]byte, error) {
	info, err := s.uvl15wCmd(uvl15wCmdInfo, []byte("UVL-15W"), uvl15wCmdInfoResp, 2*time.Second)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(info, []byte("UVL-15W")) {
		return nil, rErr("la radio non è una UVL-15W (%q)", info[:min(16, len(info))])
	}
	off, op := 10, "lettura"
	if write {
		off, op = 11, "scrittura"
	}
	if len(info) > off && info[off] != 0 {
		return nil, rErr("sulla radio è impostata una password di %s: toglila dalla radio, poi riprova", op)
	}
	return info, nil
}

func uvl15wRange() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uvl15wFlashStart)
	binary.BigEndian.PutUint32(b[4:], uvl15wFlashEnd-1)
	return b
}

func (f uvl15wFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return mem, note, err
}

// DownloadMeta: come Download, con la versione del firmware nei metadati del
// file .img (uvl15w_version, come CHIRP)
func (f uvl15wFam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, nil, "", err
	}
	defer s.p.Close()
	prog(0, uvl15wSize, "Collegamento con la radio…")
	info, err := s.uvl15wInfo(false)
	if err != nil {
		return nil, nil, "", genFail("Identificazione", err)
	}
	ver := ""
	if len(info) >= 20 {
		ver = strings.Trim(string(info[12:20]), "\x00")
	}
	meta := map[string]interface{}{"uvl15w_version": ver}
	r, err := s.uvl15wCmd(uvl15wCmdReadStart, uvl15wRange(), -1, 2*time.Second)
	if err != nil {
		return nil, nil, "", err
	}
	if !bytes.Contains(r, []byte("READ START OK")) {
		return nil, nil, "", rErr("la radio ha rifiutato la lettura (%q)", r)
	}
	var data []byte
	for a := uvl15wFlashStart; a < uvl15wFlashEnd; a += 0x800 {
		prog(a-uvl15wFlashStart, uvl15wSize, "Lettura…")
		n := min(0x800, uvl15wFlashEnd-a)
		req := make([]byte, 6)
		binary.BigEndian.PutUint32(req, uint32(a))
		binary.BigEndian.PutUint16(req[4:], uint16(n))
		resp, err := s.uvl15wCmd(uvl15wCmdRead, req, uvl15wCmdWrite, 5*time.Second)
		if err != nil {
			return nil, nil, "", err
		}
		if !bytes.Equal(resp[:min(6, len(resp))], req) || len(resp) != 6+n {
			return nil, nil, "", rErr("blocco 0x%06X non valido", a)
		}
		data = append(data, resp[6:]...)
	}
	_, _ = s.uvl15wCmd(uvl15wCmdFinish, []byte("Read Complete"), -1, 5*time.Second)
	if uvl15wLayout(data) == 0 {
		return data, meta, "La memoria letta ha una versione che questo programma (come CHIRP) non conosce: i canali si leggono, " +
			"ma prima di scrivere controlla il firmware della radio.", nil
	}
	prog(uvl15wSize, uvl15wSize, "Lettura completata")
	return data, meta, "Firmware della radio: " + ver, nil
}

func (f uvl15wFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < uvl15wSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), uvl15wSize)
	}
	if uvl15wLayout(img) == 0 {
		return "", fmt.Errorf("l'immagine non è una memoria letta da una UVL-15W con un firmware conosciuto: leggi prima la radio")
	}
	img = append([]byte{}, img[:uvl15wSize]...)
	// come il programma del costruttore: VFO A/B copiati in Temp A/B
	b := f.mem()
	b.F("temp").I(0).SetRaw(img, b.F("vfo").I(0).Raw(img))
	b.F("temp").I(1).SetRaw(img, b.F("vfo").I(1).Raw(img))
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, uvl15wSize, "Collegamento con la radio…")
	if err := s.uvl15wIdent(true); err != nil {
		return "", genFail("Identificazione", err)
	}
	r, err := s.uvl15wCmd(uvl15wCmdWriteStart, uvl15wRange(), -1, 2*time.Second)
	if err != nil {
		return "", err
	}
	if !bytes.Contains(r, []byte("WRITE START OK")) {
		return "", rErr("la radio ha rifiutato la scrittura (%q)", r)
	}
	for off := 0; off < uvl15wSize; off += 0x200 {
		prog(off, uvl15wSize, "Scrittura…")
		blk := img[off:min(off+0x200, uvl15wSize)]
		hdr := make([]byte, 6)
		binary.BigEndian.PutUint32(hdr, uint32(uvl15wFlashStart+off))
		binary.BigEndian.PutUint16(hdr[4:], uint16(len(blk)))
		resp, err := s.uvl15wCmd(uvl15wCmdWrite, append(hdr, blk...), -1, 5*time.Second)
		if err != nil {
			return "", err
		}
		if !bytes.HasPrefix(resp, []byte("WF OK")) || len(resp) < 11 || !bytes.Equal(resp[5:11], hdr) {
			return "", rErr("la radio ha rifiutato il blocco 0x%06X", uvl15wFlashStart+off)
		}
	}
	_, _ = s.uvl15wCmd(uvl15wCmdFinish, []byte("Write Complete"), -1, 5*time.Second)
	prog(uvl15wSize, uvl15wSize, "Scrittura completata")
	return "", nil
}
