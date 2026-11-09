// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.13: Quansheng UV-K5 / UV-K5 V3 / UV-K1 con il firmware IJV (di IU0IJV,
// www.ijvradio.com): V3 (UV-K5 hardware V1, 200 o 999 canali), V4 e V5 («FW5»)
// (UV-K5 V3 e UV-K1, 500 canali per banco, 16 banchi).
//
// Codice scritto da zero per l'interoperabilità. La disposizione della memoria e
// i comandi sono ricavati dai moduli CHIRP pubblicati da IJV (3.60, V4.0, V5.0) e
// dall'analisi di RadioManager 4.3.2 (note in APPOGGIO\IJV\ANALISI_IJV.md, non
// nel pacchetto). Come per gli altri firmware della UV-K5, ogni scheda legge e
// scrive solo il suo firmware.
//
// Collegamento a 38400 baud con la cornice della UV-K5 «AB CD … DC BA»:
//   - V3: dati offuscati con la tabella XOR della UV-K5 (protocollo di fabbrica);
//   - V4 e FW5: protocollo IJV in chiaro (dati non offuscati, allungati a un
//     multiplo di 4 byte, CRC16 XMODEM), quello di RadioManager; se la radio non
//     risponde si riprova con il protocollo di fabbrica (modulo CHIRP V4).
// Memoria logica (V4 e FW5): 0x0000-0x1FFF area fissa (impostazioni), 0x2000-
// 0x5FFF banco attivo (500 canali da 32 byte, 16 nomi dei gruppi a 0x5E80, nome
// del banco a 0x5F00). La V4 la legge con i comandi della UV-K5 (0x051B/0x051D);
// la FW5 con la lettura fisica (0x0564/0x0566): area fissa a 0x000000, banco n a
// 0x020000 + n·0x4000. Si legge e si scrive solo il banco scelto sulla radio.
// V6 (4.13.1): stessa memoria della FW5, ma con il collegamento dati della V6
// (drv_ijv6.go).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"
	"time"
)

type ijvKind int

const (
	ijvV3  ijvKind = iota // UV-K5 hardware V1, 200 canali a 0x0500
	ijvV3X                // UV-K5 hardware V1, 999 canali a 0x2000 (versioni «X3»)
	ijvV4                 // UV-K5 V3 / UV-K1, memoria logica
	ijvFW5                // UV-K5 V3 / UV-K1, memoria fisica
	ijvV6                 // 4.13.1: UV-K5 V3 / UV-K1, memoria fisica con il collegamento dati della V6
)

type ijvFam struct{ kind ijvKind }

const (
	ijvRecSize    = 32
	ijvAppSize    = 0x6000 // V4 e FW5: area fissa + banco attivo
	ijvFixedEnd   = 0x2000
	ijvBankChans  = 500
	ijvGroupNames = 0x5E80
	ijvBankName   = 0x5F00
	ijvBankNames  = 0x14C0 // 16 × 8 nell'area fissa
	ijvBankBase   = 0x020000
	ijvBankSize   = 0x4000
	ijvCfg5       = 0x1550 // FW5: configurazione «CFG5» nell'area fissa
	ijvCfg5Magic  = 0x35474643
	ijvCfg5Bank   = ijvCfg5 + 19
	ijvV3Size     = 0x2000
	ijvV3XSize    = 0xA000
	ijvV3Chan     = 0x0500
	ijvV3XChan    = 0x2000
	ijvSettingEnd = 0x1E00 // da qui le tarature: mai scritte
)

// ijvFamily: famiglia del firmware IJV dalla versione del saluto ("" se non è
// IJV). Le regole sono quelle di RadioManager (DetectFirmwareFamily), ma più
// strette: niente versioni numeriche sole (anche il firmware originale ha 3.00,
// 4.00, 5.00) e niente «V4» da solo (c'è anche in «F4HWN v4.x»).
func ijvFamily(fw string) string {
	up := strings.ToUpper(strings.TrimSpace(fw))
	switch {
	case up == "":
		return ""
	case strings.HasPrefix(up, "IJV V6"):
		return "V6"
	case strings.Contains(up, "FW5"):
		return "FW5"
	case strings.Contains(up, "V-X3") || strings.Contains(up, "X3."):
		return "V3X"
	case strings.Contains(up, "FW4") || strings.HasPrefix(up, "V-X") || (strings.Contains(up, "IJV") && strings.Contains(up, "V4")):
		// RadioManager accetta ogni versione con «V4», ma anche «F4HWN v4.3» la
		// contiene: qui servono «FW4», «V-X…» (moduli CHIRP V4) o «IJV … V4»
		return "V4"
	case strings.Contains(up, "N3.") || (strings.Contains(up, "IJV") && ijvV3re.MatchString(up)):
		return "V3"
	}
	return ""
}

var ijvV3re = regexp.MustCompile(`(^|[^0-9])3\.\d+`)

func (k ijvKind) family() string {
	return [...]string{"V3", "V3X", "V4", "FW5", "V6"}[k]
}

func init() {
	mk := func(id, model, variant, rclass, name string, kind ijvKind, slots int) *genModel {
		m := &genModel{ID: id, Vendor: "Quansheng", Model: model, Variant: variant, RClass: rclass, Name: name,
			Slots: slots, First: 1, NameMax: 10, Charset: charsetASCII,
			Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 2,5 W", 2.5, 1}, {"Alta 5 W", 5, 0}},
			Bands:  []band{{13000000, 1300000000}}, Narrow: true, HasScan: true, NoChirp: true,
			DCS: dtcsCodes, Baud: 38400, Fam: ijvFam{kind},
			Opt: &uvk5Opt{approve: func(v string) bool { return ijvFamily(v) == kind.family() }}}
		return m
	}
	how := "Accendi la radio, poi collega il cavo al connettore microfono/altoparlante (ben inserito). "
	v3 := mk("qs-uvk5-ijv3", "UV-K5", "IJV V3", "IJVV3Radio", "Quansheng UV-K5 (firmware IJV V3, 200 canali)", ijvV3, 200)
	v3.Notes = "Firmware IJV V3 per la UV-K5 classica (hardware V1), versione con 200 canali. " + how +
		"Si legge tutta la memoria (8 KiB); si scrivono canali, gruppi e impostazioni, mai le tarature (da 0x1E00). " +
		"Protocollo e memoria ricavati dai moduli CHIRP 3.60 pubblicati da IJV."
	v3x := mk("qs-uvk5-ijv3x", "UV-K5", "IJV V3 X", "IJVV3XRadio", "Quansheng UV-K5 (firmware IJV V3 X, 999 canali)", ijvV3X, 999)
	v3x.Notes = "Firmware IJV V3 «X» (999 canali) per la UV-K5 classica (hardware V1). " + how +
		"Si leggono e si scrivono i canali del banco scelto sulla radio (0x2000-0x9FFF) e le impostazioni; mai le tarature. " +
		"Protocollo e memoria ricavati dai moduli CHIRP X3.60 pubblicati da IJV."
	v4 := mk("qs-uvk5v3-ijv4", "UV-K5 V3 / UV-K1", "IJV V4", "IJVV4Radio", "Quansheng UV-K5 V3 / UV-K1 (firmware IJV V4)", ijvV4, ijvBankChans)
	v4.Notes = "Firmware IJV V4 (completo) per UV-K5 V3 e UV-K1. " + how +
		"Si leggono e si scrivono i 500 canali del banco scelto sulla radio, con nomi dei gruppi e impostazioni: per un altro banco " +
		"sceglilo sulla radio e rileggi. Si scrivono solo i blocchi cambiati; tarature mai. Prima di scrivere il programma controlla che " +
		"sulla radio sia ancora scelto lo stesso banco. Protocollo di RadioManager e del modulo CHIRP V4.0 di IJV."
	f5 := mk("qs-uvk5v3-ijv5", "UV-K5 V3 / UV-K1", "IJV FW5", "IJVFW5Radio", "Quansheng UV-K5 V3 / UV-K1 (firmware IJV V5)", ijvFW5, ijvBankChans)
	f5.Bands = []band{{100000, 1300000000}}
	f5.Notes = "Firmware IJV V5 («FW5», a moduli) per UV-K5 V3 e UV-K1. " + how +
		"Si leggono e si scrivono i 500 canali del banco scelto sulla radio, con nomi dei gruppi e impostazioni principali. " +
		"Sotto 13 MHz la radio riceve solo con la scheda HF. Se la radio rifiuta la scrittura, abilita sulla radio la scrittura dei " +
		"banchi dal PC. Moduli, SMS, tasti, SYSTEM e licenza restano a RadioManager di IJV."
	v6 := mk("qs-uvk5v3-ijv6", "UV-K5 V3 / UV-K1", "IJV V6", "IJVV6Radio", "Quansheng UV-K5 V3 / UV-K1 (firmware IJV V6)", ijvV6, ijvBankChans)
	v6.Bands = f5.Bands
	v6.Notes = "Firmware IJV V6 (a moduli) per UV-K5 V3 e UV-K1. " + how +
		"Collegamento dati della V6 (DTR e RTS bassi): se la radio dice che manca il collegamento dati, aggiornala con FLASH V6 di " +
		"RadioManager. Si leggono e si scrivono i 500 canali del banco scelto sulla radio, con nomi dei gruppi e impostazioni principali, " +
		"a settori di 4 KiB rilevati e verificati uno per uno; tarature, licenza, moduli, voci e file dei moduli mai. Moduli e firmware " +
		"restano a RadioManager di IJV."
	genRegister(v3, v3x, v4, f5, v6)
}

// ---------------------------------------------------------------- memoria

func (f ijvFam) imgSize() int {
	switch f.kind {
	case ijvV3:
		return ijvV3Size
	case ijvV3X:
		return ijvV3XSize
	}
	return ijvAppSize
}

func (f ijvFam) chanBase() int {
	if f.kind == ijvV3 {
		return ijvV3Chan
	}
	return ijvV3XChan // anche V4 e FW5: 0x2000
}

func (f ijvFam) slots() int {
	switch f.kind {
	case ijvV3:
		return 200
	case ijvV3X:
		return 999
	}
	return ijvBankChans
}

// groupNames: inizio dei 16 nomi dei gruppi (V3: area 0x0300, banco 0 per la «X»)
func (f ijvFam) groupNames() int {
	if f.kind == ijvV3 || f.kind == ijvV3X {
		return 0x0300
	}
	return ijvGroupNames
}

func (f ijvFam) rec(img []byte, i int) []byte {
	o := f.chanBase() + i*ijvRecSize
	if i < 0 || i >= f.slots() || o+ijvRecSize > len(img) {
		return nil
	}
	return img[o : o+ijvRecSize]
}

func (f ijvFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f ijvFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if r := f.rec(img, i); r != nil && len(raw) == ijvRecSize {
		copy(r, raw)
	}
}

func (f ijvFam) Match(m *genModel, img []byte) bool { return false }
func (f ijvFam) crossband() bool                    { return true }

// bande del firmware (BANDS dei moduli CHIRP; la V5 dalla 0,1 MHz)
var ijvBands = [][2]int64{{13000000, 108000000}, {108000000, 137000000}, {137000000, 174000000}, {174000000, 350000000},
	{350000000, 400000000}, {400000000, 470000000}, {470000000, 1300000000}}

func (f ijvFam) band(hz int64) int {
	for k, b := range ijvBands {
		lo := b[0]
		if k == 0 && (f.kind == ijvFW5 || f.kind == ijvV6) {
			lo = 100000
		}
		if hz >= lo && hz < b[1] {
			return k
		}
	}
	return -1
}

func ijvEmpty(r []byte) bool {
	if len(r) < ijvRecSize {
		return true
	}
	fr := binary.LittleEndian.Uint32(r[16:])
	return fr == 0 || fr == 0xFFFFFFFF || r[15]&0x0F > 6
}

// modulazioni (indice del campo, bit 4-2 del byte 27)
const (
	ijvModFM = 0
	ijvModAM = 1
)

func ijvAir(hz int64) bool { return hz >= 108000000 && hz < 137000000 }

func (f ijvFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if r == nil || ijvEmpty(r) {
		return nil
	}
	freq := int64(binary.LittleEndian.Uint32(r[16:]))
	off := int64(binary.LittleEndian.Uint32(r[20:]))
	c := &Chan{Rx: freq * 10}
	c.Tx = c.Rx
	if off != 0 {
		switch r[27] & 3 {
		case 2:
			if off == freq {
				c.TxOff = true
			} else {
				c.Tx = c.Rx - off*10
			}
		case 1:
			c.Tx = c.Rx + off*10
		}
	}
	c.RxTone = uvk5ToneDec(int64(r[26]&0x0F), int64(r[24]))
	c.TxTone = uvk5ToneDec(int64(r[26]>>4), int64(r[25]))
	switch (r[28] >> 5) & 3 {
	case 2:
		c.Power = 0
	case 1:
		c.Power = 1
	default:
		c.Power = 2
	}
	c.Narrow = (r[28]>>1)&0x0F >= 6
	c.Scan = r[27]&0x20 == 0
	end := 10
	for k := 0; k < 10; k++ {
		if r[k] == 0x00 || r[k] == 0xFF {
			end = k
			break
		}
	}
	nm := make([]rune, 0, end)
	for _, b := range r[:end] {
		if b < 0x20 || b > 0x7E {
			b = '?'
		}
		nm = append(nm, rune(b))
	}
	c.Name = strings.TrimRight(string(nm), " ")
	return c
}

// ijvBlankRec: posizione libera (0xFF, cifre della selettiva «vuote» 0xE)
func ijvBlankRec(r []byte) {
	for k := range r {
		r[k] = 0xFF
	}
	for k := 10; k < 15; k++ {
		r[k] = 0xEE
	}
}

func (f ijvFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if r == nil {
		return fmt.Errorf("posizione %d fuori dalla memoria", i+1)
	}
	if c == nil {
		ijvBlankRec(r)
		return nil
	}
	bd := f.band(c.Rx)
	if bd < 0 {
		return fmt.Errorf("frequenza %s MHz fuori dalle gamme del firmware IJV", fmtMHz(c.Rx))
	}
	if !c.TxOff && c.Tx != c.Rx && f.band(c.Tx) < 0 {
		return fmt.Errorf("frequenza di trasmissione %s MHz fuori dalle gamme del firmware IJV", fmtMHz(c.Tx))
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	old := append([]byte{}, r...)
	wasEmpty := ijvEmpty(old)
	oldRx := int64(binary.LittleEndian.Uint32(old[16:])) * 10
	nr := make([]byte, ijvRecSize)
	if wasEmpty {
		// canale nuovo: valori iniziali di RadioManager (passo 12,5 kHz, squelch 1,
		// gruppo 0, nessuna selettiva)
		for k := 10; k < 15; k++ {
			nr[k] = 0xEE
		}
		nr[30] = 1<<4 | 11
	} else {
		copy(nr, old)
	}
	// nome: 10 caratteri ASCII, spazi in fondo (come i moduli CHIRP)
	for k := 0; k < 10; k++ {
		nr[k] = ' '
	}
	for k, ch := range []rune(c.Name) {
		if k >= 10 {
			break
		}
		if ch < 0x20 || ch > 0x7E {
			ch = '?'
		}
		nr[k] = byte(ch)
	}
	nr[15] = nr[15]&0xF0 | byte(bd)
	if wasEmpty {
		nr[15] = byte(bd)
	}
	freq := c.Rx / 10
	binary.LittleEndian.PutUint32(nr[16:], uint32(freq))
	var shift byte
	var off int64
	switch {
	case c.TxOff:
		shift, off = 2, freq
	case c.Tx > c.Rx:
		shift, off = 1, (c.Tx-c.Rx)/10
	case c.Tx < c.Rx:
		shift, off = 2, (c.Rx-c.Tx)/10
	}
	binary.LittleEndian.PutUint32(nr[20:], uint32(off))
	fl, cd := uvk5ToneEnc(rt)
	nr[24] = byte(cd)
	nr[26] = byte(fl)
	fl, cd = uvk5ToneEnc(tt)
	nr[25] = byte(cd)
	nr[26] |= byte(fl) << 4
	// modulazione: quella del canale se la frequenza non cambia; AM nella gamma
	// aeronautica, FM per un canale nuovo o che esce dall'AM aeronautica
	mod := (old[27] >> 2) & 7
	switch {
	case wasEmpty:
		mod = ijvModFM
		if ijvAir(c.Rx) {
			mod = ijvModAM
		}
	case oldRx != c.Rx && ijvAir(c.Rx):
		mod = ijvModAM
	case oldRx != c.Rx && mod == ijvModAM && !ijvAir(c.Rx):
		mod = ijvModFM
	}
	if f.kind == ijvV3 || f.kind == ijvV3X {
		if mod > 5 {
			mod = ijvModFM
		}
	} else if mod > 6 {
		mod = ijvModFM
	}
	b27 := nr[27] & 0xC0 // blocco TX e protezione conservati
	if wasEmpty {
		b27 = 0
	}
	if !c.Scan {
		b27 |= 0x20
	}
	nr[27] = b27 | mod<<2 | shift
	// banda passante: quella del canale se è dello stesso tipo, altrimenti
	// W 26k (larga) o N 10k (stretta)
	bw := (old[28] >> 1) & 0x0F
	if wasEmpty || bw > 9 || (bw >= 6) != c.Narrow {
		bw = 0
		if c.Narrow {
			bw = 6
		}
	}
	var pw byte
	switch m.powerIndex(c.Power) {
	case 2:
		pw = 2
	case 1:
		pw = 1
	}
	b28 := nr[28] & 0x81 // canale occupato e inversione conservati
	if wasEmpty {
		b28 = 0
	}
	nr[28] = b28 | pw<<5 | bw<<1
	copy(r, nr)
	return nil
}

// ---------------------------------------------------------------- collegamento

// ijvSess: sessione con la radio. plain = protocollo IJV in chiaro (V4, FW5)
type ijvSess struct {
	ctx   context.Context
	p     Port
	plain bool
}

const ijvTimeout = 1500 * time.Millisecond

var ijvSession = []byte{0x6a, 0x39, 0x57, 0x64}

// ijvPlainPacket: comando IJV in chiaro (dati allungati a un multiplo di 4)
func ijvPlainPacket(data []byte) []byte {
	pad := append(append([]byte{}, data...), make([]byte, (4-len(data)%4)%4)...)
	out := []byte{0xAB, 0xCD, byte(len(pad)), 0x00}
	out = append(out, pad...)
	crc := crc16XModem(pad)
	return append(out, byte(crc), byte(crc>>8), 0xDC, 0xBA)
}

func (s *ijvSess) send(data []byte) error {
	if s.plain {
		return s.p.Write(ijvPlainPacket(data))
	}
	return s.p.Write(uvk5Packet(data))
}

func (s *ijvSess) recv(timeout time.Duration) ([]byte, error) {
	h := make([]byte, 4)
	if err := s.p.ReadFull(s.ctx, h, timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	if h[0] != 0xAB || h[1] != 0xCD || (!s.plain && h[3] != 0) {
		return nil, rErr("risposta della radio non valida")
	}
	n := int(h[2])
	if s.plain {
		n |= int(h[3]) << 8
	}
	body := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, body, timeout); err != nil {
		return nil, rErr("risposta della radio incompleta")
	}
	foot := make([]byte, 4)
	if err := s.p.ReadFull(s.ctx, foot, timeout); err != nil || foot[2] != 0xDC || foot[3] != 0xBA {
		return nil, rErr("risposta della radio incompleta")
	}
	if !s.plain {
		return uvk5Xor(body), nil
	}
	if crc := binary.LittleEndian.Uint16(foot); crc != 0xFFFF && crc != crc16XModem(body) {
		return nil, rErr("risposta della radio con CRC non valido")
	}
	return body, nil
}

// request: comando e risposta con l'identificativo atteso (3 tentativi, come
// RadioManager e i moduli CHIRP)
func (s *ijvSess) request(data []byte, want uint16, timeout time.Duration) ([]byte, error) {
	return s.requestN(data, want, timeout, 3)
}

func (s *ijvSess) requestN(data []byte, want uint16, timeout time.Duration, tries int) ([]byte, error) {
	var last error
	for t := 0; t < tries; t++ {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		if t > 0 {
			time.Sleep(150 * time.Millisecond)
			_ = s.p.Purge()
		}
		if err := s.send(data); err != nil {
			return nil, err
		}
		// la FW5 può mandare anche pacchetti suoi (es. l'annuncio 0x0570):
		// si saltano finché arriva la risposta attesa
		for skip := 0; skip < 8; skip++ {
			rep, err := s.recv(timeout)
			if err != nil {
				last = err
				break
			}
			if len(rep) >= 2 && binary.LittleEndian.Uint16(rep) == want {
				return rep, nil
			}
			last = rErr("risposta inattesa della radio")
		}
	}
	return nil, last
}

func ijvAscii(b []byte, from, n int) string {
	var out []byte
	for k := from; k < len(b) && k < from+n; k++ {
		if b[k] < ' ' || b[k] > '~' {
			break
		}
		out = append(out, b[k])
	}
	return string(out)
}

func (s *ijvSess) hello() (string, error) {
	// 2 tentativi per protocollo: la ricerca delle radio («Identifica») non si
	// allunga troppo con le radio che non rispondono
	rep, err := s.requestN(append([]byte{0x14, 0x05, 0x04, 0x00}, ijvSession...), 0x0515, ijvTimeout, 2)
	if err != nil {
		return "", err
	}
	return ijvAscii(rep, 4, 16), nil
}

func (s *ijvSess) readMem(off, n int) ([]byte, error) {
	cmd := append([]byte{0x1b, 0x05, 0x08, 0x00, byte(off), byte(off >> 8), byte(n), 0}, ijvSession...)
	rep, err := s.request(cmd, 0x051C, 3*time.Second)
	if err != nil {
		return nil, err
	}
	if len(rep) < 8+n || int(binary.LittleEndian.Uint16(rep[4:])) != off || int(rep[6]) != n {
		return nil, rErr("lettura della memoria a 0x%04X non valida", off)
	}
	return rep[8 : 8+n], nil
}

func (s *ijvSess) writeMem(off int, data []byte) error {
	cmd := append([]byte{0x1d, 0x05, byte(len(data) + 8), 0, byte(off), byte(off >> 8), byte(len(data)), 1}, ijvSession...)
	rep, err := s.request(append(cmd, data...), 0x051E, 5*time.Second)
	if err != nil {
		return err
	}
	if len(rep) < 6 || int(binary.LittleEndian.Uint16(rep[4:])) != off {
		return rErr("la radio non ha confermato la scrittura a 0x%04X", off)
	}
	return nil
}

func (s *ijvSess) readPhys(addr, n int) ([]byte, error) {
	cmd := make([]byte, 10)
	binary.LittleEndian.PutUint16(cmd, 0x0564)
	binary.LittleEndian.PutUint16(cmd[2:], 6)
	binary.LittleEndian.PutUint32(cmd[4:], uint32(addr))
	cmd[8] = byte(n)
	rep, err := s.request(cmd, 0x0565, 3*time.Second)
	if err != nil {
		return nil, err
	}
	if len(rep) < 10+n || int(binary.LittleEndian.Uint32(rep[4:])) != addr || int(rep[8]) != n || rep[9] != 0 {
		return nil, rErr("lettura della memoria a 0x%06X non riuscita", addr)
	}
	return rep[10 : 10+n], nil
}

func (s *ijvSess) writePhys(addr int, data []byte) error {
	cmd := make([]byte, 10, 10+len(data))
	binary.LittleEndian.PutUint16(cmd, 0x0566)
	binary.LittleEndian.PutUint16(cmd[2:], uint16(6+len(data)))
	binary.LittleEndian.PutUint32(cmd[4:], uint32(addr))
	cmd[8] = byte(len(data))
	rep, err := s.request(append(cmd, data...), 0x0567, 5*time.Second)
	if err != nil {
		return err
	}
	if len(rep) < 10 || int(binary.LittleEndian.Uint32(rep[4:])) != addr || int(rep[8]) != len(data) {
		return rErr("la radio non ha confermato la scrittura a 0x%06X", addr)
	}
	if rep[9] != 0 {
		return rErr("la radio ha rifiutato la scrittura a 0x%06X: abilita sulla radio la scrittura dei banchi dal PC e riprova", addr)
	}
	return nil
}

// flush: la FW5 salva nella flash i dati ancora in memoria (comando 0x0576)
func (s *ijvSess) flush() error {
	rep, err := s.request([]byte{0x76, 0x05, 0x00, 0x00}, 0x0577, 5*time.Second)
	if err != nil {
		return err
	}
	if len(rep) < 5 || rep[4] != 0 {
		return rErr("la radio non ha confermato il salvataggio della memoria")
	}
	return nil
}

func (s *ijvSess) reset() { _ = s.send([]byte{0xdd, 0x05, 0x00, 0x00}) }

// start: saluto e controllo del firmware. V3 con il protocollo di fabbrica; V4 e
// FW5 prima in chiaro, poi con quello di fabbrica
func (f ijvFam) start(ctx context.Context, m *genModel, p Port) (*ijvSess, string, error) {
	_ = setRTSOff(p) // RTS spento e DTR acceso, come RadioManager e i moduli CHIRP
	time.Sleep(200 * time.Millisecond)
	_ = p.Purge()
	modes := []bool{true, false}
	if f.kind == ijvV3 || f.kind == ijvV3X {
		modes = []bool{false}
	}
	var fw string
	var s *ijvSess
	var err error
	for _, plain := range modes {
		s = &ijvSess{ctx: ctx, p: p, plain: plain}
		if fw, err = s.hello(); err == nil {
			break
		}
		if err == errCancelled {
			return nil, "", err
		}
		_ = p.Purge()
	}
	if err != nil {
		if f.kind == ijvV4 || f.kind == ijvFW5 {
			if info := ijvV6Probe(ctx, p); info != "" {
				return nil, "", rErr("la radio ha il firmware IJV V6 (%s): apri la scheda «%s»", info, genByID["qs-uvk5v3-ijv6"].Name)
			}
		}
		return nil, "", genFail("Collegamento", err)
	}
	if fw == "" {
		return nil, "", rErr("la radio non ha comunicato la versione del firmware")
	}
	if ijvFamily(fw) != f.kind.family() {
		for _, o := range genModels {
			if op, ok := o.Opt.(*uvk5Opt); ok && o != m && op.approve(fw) {
				return nil, fw, rErr("la radio ha il firmware %s: apri la scheda «%s» (questa scheda è per un altro firmware e non legge e non scrive)", fw, o.Name)
			}
		}
		return nil, fw, rErr("il firmware della radio (%s) non è %s: per sicurezza non si legge e non si scrive", fw, "IJV "+f.kind.family())
	}
	return s, fw, nil
}

// bank: banco attivo sulla radio (V4: byte logico 0x000B; FW5: «CFG5»)
func (f ijvFam) bank(s *ijvSess) (int, error) {
	if f.kind == ijvV4 {
		b, err := s.readMem(0x000B, 1)
		if err != nil {
			return 0, err
		}
		return int(b[0]), nil
	}
	cfg, err := s.readPhys(ijvCfg5, 32)
	if err != nil {
		return 0, err
	}
	return ijvCfgBank(cfg)
}

func ijvCfgBank(cfg []byte) (int, error) {
	if len(cfg) < 20 || binary.LittleEndian.Uint32(cfg) != ijvCfg5Magic || cfg[19] > 15 {
		return 0, rErr("la radio non indica il banco di memoria scelto (configurazione «CFG5» non valida)")
	}
	return int(cfg[19]), nil
}

// imgBank: banco della memoria letta
func (f ijvFam) imgBank(img []byte) (int, error) {
	if f.kind == ijvV4 {
		if len(img) <= 0x0B || img[0x0B] > 15 {
			return 0, rErr("la memoria non indica il banco")
		}
		return int(img[0x0B]), nil
	}
	if len(img) < ijvCfg5+32 {
		return 0, rErr("la memoria non indica il banco")
	}
	return ijvCfgBank(img[ijvCfg5 : ijvCfg5+32])
}

func (f ijvFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return mem, note, err
}

func (f ijvFam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	if f.kind == ijvV6 {
		return f.v6Download(ctx, m, port, prog)
	}
	p, err := openPort(port, 38400)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	size := f.imgSize()
	prog(0, size, "Collegamento con la radio…")
	s, fw, err := f.start(ctx, m, p)
	if err != nil {
		return nil, nil, "", err
	}
	data := make([]byte, 0, size)
	read := func(a, n int) error {
		if ctx.Err() != nil {
			return errCancelled
		}
		prog(a, size, "Lettura…")
		var b []byte
		var err error
		switch {
		case f.kind == ijvFW5 && a < ijvFixedEnd:
			b, err = s.readPhys(a, n)
		case f.kind == ijvFW5:
			bk, e := f.imgBank(data)
			if e != nil {
				return e
			}
			b, err = s.readPhys(ijvBankBase+bk*ijvBankSize+a-ijvFixedEnd, n)
		default:
			b, err = s.readMem(a, n)
		}
		if err == nil {
			data = append(data, b...)
		}
		return err
	}
	// blocchi: 128 byte (V3, FW5); 32 byte per la V4 (con 128 byte il collegamento
	// USB della PY32F071 non è affidabile, nota del modulo CHIRP V4)
	blk := 0x80
	if f.kind == ijvV4 {
		blk = 0x20
	}
	for a := 0; a < size; a += blk {
		if err := read(a, blk); err != nil {
			return nil, nil, "", err
		}
	}
	prog(size, size, "Lettura completata")
	meta := map[string]interface{}{"uvk5_firmware": fw}
	note := "Firmware della radio: " + fw
	if f.kind == ijvV4 || f.kind == ijvFW5 {
		bk, err := f.imgBank(data)
		if err != nil {
			return nil, nil, "", err
		}
		meta["ijv_bank"] = bk
		note += fmt.Sprintf(" — banco %d «%s»", bk+1, ijvBankLabel(data, bk))
	}
	return data, meta, note, nil
}

// ijvBankLabel: nome del banco (area fissa, poi quello del banco attivo)
func ijvBankLabel(img []byte, bk int) string {
	clean := func(b []byte) string {
		return strings.TrimSpace(strings.Trim(ijvAscii(b, 0, len(b)), "\x00\xff"))
	}
	if o := ijvBankNames + bk*8; o+8 <= len(img) {
		if n := clean(img[o : o+8]); n != "" {
			return n
		}
	}
	if ijvBankName+8 <= len(img) {
		if n := clean(img[ijvBankName : ijvBankName+8]); n != "" {
			return n
		}
	}
	return fmt.Sprintf("BANK %02d", bk)
}

// writeRanges: zone scritte (mai le tarature da 0x1E00 né, per la FW5, la
// configurazione e la licenza oltre 0x1540)
func (f ijvFam) writeRanges() [][2]int {
	switch f.kind {
	case ijvV3:
		return [][2]int{{0, 0x50}, {0x140, ijvSettingEnd}}
	case ijvV3X:
		return [][2]int{{0, 0x50}, {0x140, ijvSettingEnd}, {ijvV3XChan, ijvV3XSize}}
	case ijvV4:
		return [][2]int{{ijvFixedEnd, ijvAppSize}, {0, ijvSettingEnd}}
	}
	return [][2]int{{ijvFixedEnd, ijvAppSize}, {0, 0x1540}}
}

func (f ijvFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != f.imgSize() {
		return "", fmt.Errorf("l'immagine è di %d byte: per questo firmware deve essere di %d byte (leggi prima la radio)", len(img), f.imgSize())
	}
	if f.kind == ijvV6 {
		return f.v6Upload(ctx, m, port, img, prog)
	}
	p, err := openPort(port, 38400)
	if err != nil {
		return "", err
	}
	defer p.Close()
	total := 0
	for _, r := range f.writeRanges() {
		total += r[1] - r[0]
	}
	prog(0, total, "Collegamento con la radio…")
	s, fw, err := f.start(ctx, m, p)
	if err != nil {
		return "", err
	}
	bk := 0
	if f.kind == ijvV4 || f.kind == ijvFW5 {
		want, err := f.imgBank(img)
		if err != nil {
			return "", err
		}
		if bk, err = f.bank(s); err != nil {
			return "", err
		}
		if bk != want {
			return "", rErr("sulla radio è scelto il banco %d, ma i canali da scrivere sono del banco %d: scegli sulla radio il banco %d "+
				"(o rileggi la radio). La radio non è stata modificata", bk+1, want+1, want+1)
		}
	}
	addr := func(a int) int {
		if f.kind == ijvFW5 && a >= ijvFixedEnd {
			return ijvBankBase + bk*ijvBankSize + a - ijvFixedEnd
		}
		return a
	}
	blk := 0x80
	if f.kind == ijvV4 {
		blk = 0x20
	}
	done, wrote, fixed := 0, 0, false
	for _, rg := range f.writeRanges() {
		for a := rg[0]; a < rg[1]; a += blk {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			prog(done+a-rg[0], total, "Scrittura…")
			end := min(a+blk, rg[1])
			want := img[a:end]
			if f.kind == ijvV3 || f.kind == ijvV3X {
				// EEPROM: si scrive tutto, come i moduli CHIRP 3.60
				if err := s.writeMem(a, want); err != nil {
					return "", err
				}
				continue
			}
			// V4 e FW5: si scrivono solo i blocchi cambiati (meno cancellazioni
			// della flash), come RadioManager e i moduli CHIRP
			var cur []byte
			if f.kind == ijvFW5 {
				cur, err = s.readPhys(addr(a), end-a)
			} else {
				cur, err = s.readMem(a, end-a)
			}
			if err != nil {
				return "", err
			}
			if bytes.Equal(cur, want) {
				continue
			}
			if f.kind == ijvFW5 {
				err = s.writePhys(addr(a), want)
			} else {
				err = s.writeMem(a, want)
			}
			if err != nil {
				return "", err
			}
			wrote++
			fixed = fixed || a < ijvFixedEnd
		}
		done += rg[1] - rg[0]
	}
	note := "Firmware della radio: " + fw
	switch {
	case f.kind == ijvV4 && fixed:
		// come RadioManager dopo le impostazioni: si riscrivono 8 byte a 0x1000 e a
		// 0x2000 perché la radio chiuda i settori aperti
		for _, a := range []int{0x1000, 0x2000} {
			b, err := s.readMem(a, 8)
			if err == nil {
				err = s.writeMem(a, b)
			}
			if err != nil {
				return "", err
			}
		}
	case f.kind == ijvFW5 && wrote > 0:
		if err := s.flush(); err != nil {
			note += " (la radio non ha confermato il salvataggio finale: la verifica dopo la scrittura dirà se i dati ci sono)"
		}
	}
	if f.kind != ijvV3 && f.kind != ijvV3X {
		note += fmt.Sprintf(" — blocchi cambiati e scritti: %d", wrote)
	}
	s.reset()
	prog(total, total, "Scrittura completata")
	return note, nil
}

// ---------------------------------------------------------------- V6

// ijvV6Probe: la radio risponde al collegamento dati della V6? Restituisce una
// descrizione (modello, versione) o "". DTR e RTS bassi, richiesta «A5 5A»,
// risposta di 12 byte «5A A5» con CRC32.
func ijvV6Probe(ctx context.Context, p Port) string {
	setDTR := func(on bool) {
		if x, ok := p.(interface{ SetDTR(bool) error }); ok {
			_ = x.SetDTR(on)
		}
	}
	setDTR(false)
	defer setDTR(true)
	_ = setRTSOff(p)
	time.Sleep(120 * time.Millisecond)
	_ = p.Purge()
	ask := func(cmd, seq byte) (uint32, bool) {
		_ = p.Write(ijvV6Request(cmd, seq, 0, nil))
		b := make([]byte, 12)
		if err := p.ReadFull(ctx, b, 800*time.Millisecond); err != nil {
			return 0, false
		}
		if b[0] != 0x5A || b[1] != 0xA5 || b[3] != seq || b[2] != 0 || crc32IEEE(b[2:8]) != binary.LittleEndian.Uint32(b[8:]) {
			return 0, false
		}
		return binary.LittleEndian.Uint32(b[4:]), true
	}
	v, ok := ask(8, 1)
	if !ok {
		return ""
	}
	model := map[uint32]string{1: "UV-K5 V3", 2: "UV-K1", 3: "UV-K10"}[(v>>28)&3]
	if model == "" {
		model = "UV-K5"
	}
	desc := model
	if ver, ok := ask(18, 2); ok {
		desc += fmt.Sprintf(", versione %d.%d.%d", (ver>>16)&0xFF, (ver>>8)&0xFF, ver&0xFF)
	}
	return desc
}

func ijvV6Request(cmd, seq byte, addr uint32, data []byte) []byte {
	b := make([]byte, 10, 14+len(data))
	b[0], b[1], b[2], b[3] = 0xA5, 0x5A, cmd, seq
	binary.LittleEndian.PutUint32(b[4:], addr)
	binary.LittleEndian.PutUint16(b[8:], uint16(len(data)))
	b = append(b, data...)
	crc := crc32IEEE(b[2:])
	return binary.LittleEndian.AppendUint32(b, crc)
}

// crc32IEEE: CRC-32 della V6 (lo stesso di zip, polinomio riflesso 0xEDB88320)
func crc32IEEE(b []byte) uint32 { return crc32.ChecksumIEEE(b) }
