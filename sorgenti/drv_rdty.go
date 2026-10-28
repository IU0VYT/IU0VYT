// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da dmrconfig (https://github.com/OpenRTX/dmrconfig, licenza BSD a
// 3 clausole, Copyright (C) 2018 Serge Vakulenko KK6ABQ): rd5r.c, dm1801.c, hid.c;
// e da qdmr (https://github.com/hmatuschek/qdmr, GNU GPL v3, Hannes Matuschek
// DM3MAT): radioddity_interface.cc.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio DMR collegate via USB come dispositivo HID (VID 15A2, PID 0073), con la
// memoria della famiglia Radioddity GD-77: Baofeng RD-5R (anche DM-5R dalla
// versione «Tier II») e Baofeng DM-1801. Memoria e protocollo come in dmrconfig
// (rd5r.c, dm1801.c, hid.c) e qdmr (radioddity_interface.cc).
//
// Protocollo: ogni scambio è un rapporto HID di 42 byte: 01 00 lunghezza(2, LE)
// dati; risposta 03 00 lunghezza 00 dati.
//   "\x02PROGRA" → 'A';  "M\x02" → 16 byte di identificazione («BF-5R…», «1801…»);
//   'A' → 'A';  "CWB\x04\x00" banco(1) "\x00\x00" → 'A' (banco 0: indirizzi sotto
//   0x10000, banco 1: da 0x10000);  'R' addrH addrL 32 → 36 byte (32 di dati dal
//   quinto);  'W' addrH addrL 32 + 32 byte → 'A';  "ENDR"/"ENDW" → 'A' (fine).
// Si leggono e si scrivono i blocchi da 128 byte dal numero 1 a nblk-1 (indirizzi
// 0x80 … fine), tranne 0x7C00…0x7FFF, come dmrconfig.
//
// Memoria (128 KB; l'immagine ha in testa l'identificativo come dmrconfig):
//   0x0088 data di programmazione, 0x00E0 nome e ID della radio
//   0x3780 canali 1-128, 0xB1B0 canali 129-1024 (banchi di 16 byte di bitmap +
//   128 canali da 56 byte), 0x8010 zone (bitmap di 32 byte + zone), liste di
//   scansione (RD-5R 0x17620, DM-1801 0x1790: byte di validità + liste da 88 byte).
// Canale: 0 nome (16, FF), 16 RX e 20 TX (BCD, 8 cifre da 10 Hz, little endian),
//   24 modo (0 analogico, 1 DMR), 32 tono RX e 34 tono TX (BCD: 2 bit di tipo,
//   CTCSS in decimi di Hz, DCS 2 = normale, 3 = invertito; FFFF nessuno),
//   42 color code TX, 44 color code RX, 49 bit 6 slot 2, 51 bit 1 banda larga,
//   bit 2 solo ricezione, bit 7 potenza alta.
//
// Qui si gestiscono i canali (frequenze, nome, modo, potenza, banda e toni); le
// impostazioni DMR dei canali esistenti, i contatti, i gruppi e le impostazioni
// della radio restano come sono. Un canale cancellato viene tolto anche dalle zone
// e dalle liste di scansione; un canale nuovo entra nella prima zona con posto
// libero (o in una zona nuova), perché la radio mostra i canali per zone.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	rdtyMemSize = 0x20000
	rdtyChSize  = 56
	rdtyBank0   = 0x3780
	rdtyBank1   = 0xB1B0
	rdtyBankLen = 16 + 128*rdtyChSize
	rdtyZoneTab = 0x8010
	rdtyScanLen = 88
	rdtyVID     = 0x15A2
	rdtyPID     = 0x0073
)

type rdtyOpt struct {
	ident     string // identificazione della radio (risposta a «M\x02»)
	hdr       string // intestazione dell'immagine (come dmrconfig)
	nblk      int    // blocchi da 128 byte (dal numero 1)
	zoneMem   int    // canali per zona
	nzones    int
	scanOff   int
	nscan     int
	scanValid int  // byte di validità delle liste di scansione
	sqBit     bool // DM-1801: squelch «normale» nel bit 0 del byte 51
}

type rdtyFam struct{}

func init() {
	mk := func(id, model, name string, o rdtyOpt, notes string) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: "IU0VYT_" + strings.ReplaceAll(model, "-", ""), Name: name,
			Slots: 1024, First: 1, NameMax: 16, Charset: charsetASCII, Powers: []genPower{{"Alta", 5, 0}, {"Bassa", 1, 2}},
			Bands: []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: false, NoChirp: true,
			DCS: dtcsCodes, Baud: 0, Fam: rdtyFam{}, Opt: &oo, Digital: "DMR", Conn: "hid", Notes: notes}
	}
	common := "Radio DMR collegata via USB (dispositivo HID, nessuna porta COM): memoria e protocollo come nei programmi liberi " +
		"dmrconfig e qdmr, confrontati con il codice di dmrconfig. Si modificano frequenze, nomi (16 caratteri), modo (analogico / DMR), " +
		"potenza, banda e toni; per i canali DMR color code, slot, contatto e lista dei gruppi restano quelli impostati (i canali DMR " +
		"nuovi: color code 1, slot 1, senza contatto). Contatti, gruppi, messaggi e impostazioni della radio non si toccano."
	genRegister(
		mk("bf-rd5r", "RD-5R", "Baofeng RD-5R / DM-5R Tier II", rdtyOpt{ident: "BF-5R", hdr: "BF-5R", nblk: 966, zoneMem: 16, nzones: 250,
			scanOff: 0x17620, nscan: 250, scanValid: 256}, common),
		mk("bf-dm1801", "DM-1801", "Baofeng DM-1801", rdtyOpt{ident: "1801", hdr: "1801", nblk: 989, zoneMem: 32, nzones: 150,
			scanOff: 0x1790, nscan: 64, scanValid: 64, sqBit: true}, common),
	)
}

func rdtyo(m *genModel) *rdtyOpt { return m.Opt.(*rdtyOpt) }

func rdtyBank(b int) int {
	if b == 0 {
		return rdtyBank0
	}
	return rdtyBank1 + (b-1)*rdtyBankLen
}

func rdtyRecOff(i int) int { return rdtyBank(i>>7) + 16 + (i%128)*rdtyChSize }

func rdtyValid(img []byte, i int) bool {
	return img[rdtyBank(i>>7)+(i%128)/8]>>(i&7)&1 == 1
}

func rdtySetValid(img []byte, i int, v bool) {
	p := rdtyBank(i>>7) + (i%128)/8
	if v {
		img[p] |= 1 << (i & 7)
	} else {
		img[p] &^= 1 << (i & 7)
	}
}

// frequenze BCD «abcdefgh» (10 Hz), little endian
func rdtyFreq(b []byte) (int64, bool) {
	v := binary.LittleEndian.Uint32(b)
	var hz int64
	for k := 7; k >= 0; k-- {
		d := (v >> (4 * k)) & 15
		if d > 9 {
			return 0, false
		}
		hz = hz*10 + int64(d)
	}
	return hz * 10, true
}

func rdtyPutFreq(b []byte, hz int64) {
	v := hz / 10
	var bcd uint32
	for k := 0; k < 8; k++ {
		bcd |= uint32(v%10) << (4 * k)
		v /= 10
	}
	binary.LittleEndian.PutUint32(b, bcd)
}

// tono: FFFF nessuno; tipo nei 2 bit alti (0 CTCSS, 2 DCS normale, 3 DCS
// invertito), poi 4 cifre BCD (CTCSS in decimi di Hz, DCS ottale come decimale)
func rdtyTone(v uint16) string {
	if v == 0xFFFF {
		return ""
	}
	tag := v >> 14
	a, b, c, d := int(v>>12&3), int(v>>8&15), int(v>>4&15), int(v&15)
	switch tag {
	case 2, 3:
		s := fmt.Sprintf("D%d%d%d", b, c, d)
		if tag == 3 {
			return s + "I"
		}
		return s + "N"
	}
	return fmtTone(a*1000 + b*100 + c*10 + d)
}

func rdtyToneEnc(t genTone) uint16 {
	switch t.Mode {
	case "Tone":
		v := t.Hz10
		return uint16(v/1000)<<12 | uint16(v/100%10)<<8 | uint16(v/10%10)<<4 | uint16(v%10)
	case "DTCS":
		v := t.Code
		tag := uint16(2)
		if t.Inv {
			tag = 3
		}
		return tag<<14 | uint16(v/100)<<8 | uint16(v/10%10)<<4 | uint16(v%10)
	}
	return 0xFFFF
}

func rdtyName(b []byte) string {
	n := bytes.IndexAny(b, "\xff\x00")
	if n < 0 {
		n = len(b)
	}
	return strings.TrimRight(string(b[:n]), " ")
}

func (f rdtyFam) Get(m *genModel, img []byte, i int) *Chan {
	if len(img) < rdtyMemSize || !rdtyValid(img, i) {
		return nil
	}
	r := img[rdtyRecOff(i):][:rdtyChSize]
	rx, ok1 := rdtyFreq(r[16:])
	tx, ok2 := rdtyFreq(r[20:])
	if !ok1 || !ok2 || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: tx, Name: rdtyName(r[0:16])}
	c.Dig = r[24] == 1
	c.Narrow = r[51]>>1&1 == 0
	if r[51]>>7&1 == 1 {
		c.Power = 0
	} else {
		c.Power = 2
	}
	c.TxOff = r[51]>>2&1 == 1
	c.Scan = r[31] != 0
	if !c.Dig {
		c.RxTone = rdtyTone(binary.LittleEndian.Uint16(r[32:]))
		c.TxTone = rdtyTone(binary.LittleEndian.Uint16(r[34:]))
	}
	return c
}

// rdtyErase: canale cancellato come erase_channel di dmrconfig
func rdtyErase(o *rdtyOpt, r []byte) {
	for k := 0; k < 16; k++ {
		r[k] = 0xFF
	}
	binary.LittleEndian.PutUint32(r[16:], 0x40000000)
	binary.LittleEndian.PutUint32(r[20:], 0x40000000)
	copy(r[24:], []byte{0, 0, 0, 0, 5, 0, 0x50, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0x16, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0})
	// byte 51: banda larga, potenza alta; 52-55
	r[51] = 1<<1 | 1<<7
	r[52], r[53], r[54], r[55] = 0, 0, 0, 0
	if o.sqBit {
		r[51] |= 1 // squelch «normale»
	} else {
		r[55] = 5
	}
}

func (f rdtyFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := rdtyo(m)
	r := img[rdtyRecOff(i):][:rdtyChSize]
	if c == nil {
		if rdtyValid(img, i) {
			rdtyErase(o, r)
			rdtySetValid(img, i, false)
			rdtyUnlink(o, img, i)
		}
		return nil
	}
	var tt, rt genTone
	if !c.Dig {
		var err error
		if tt, err = genToneParse(c.TxTone); err != nil {
			return err
		}
		if rt, err = genToneParse(c.RxTone); err != nil {
			return err
		}
	}
	isNew := !rdtyValid(img, i)
	if isNew {
		rdtyErase(o, r)
		// valori di setup_channel di dmrconfig per un canale nuovo
		if c.Dig {
			r[42], r[44] = 1, 1 // color code 1
			r[29] = 2           // ammissione: color code
		} else {
			r[42], r[44] = 0, 0
		}
	}
	copy(r[0:16], bytes.Repeat([]byte{0xFF}, 16))
	copy(r[0:16], c.Name)
	rdtyPutFreq(r[16:], c.Rx)
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	rdtyPutFreq(r[20:], tx)
	if c.Dig {
		r[24] = 1
		binary.LittleEndian.PutUint16(r[32:], 0xFFFF)
		binary.LittleEndian.PutUint16(r[34:], 0xFFFF)
		if isNew {
			r[51] &^= 1 << 1 // banda stretta (12,5 kHz) per i canali DMR
		}
	} else {
		r[24] = 0
		binary.LittleEndian.PutUint16(r[32:], rdtyToneEnc(rt))
		binary.LittleEndian.PutUint16(r[34:], rdtyToneEnc(tt))
		if c.Narrow {
			r[51] &^= 1 << 1
		} else {
			r[51] |= 1 << 1
		}
	}
	if m.powerIndex(c.Power) == 0 {
		r[51] |= 1 << 7
	} else {
		r[51] &^= 1 << 7
	}
	if c.TxOff {
		r[51] |= 1 << 2
	} else {
		r[51] &^= 1 << 2
	}
	if isNew {
		rdtySetValid(img, i, true)
		rdtyZoneAdd(o, img, i)
	}
	return nil
}

// ---------------------------------------------------------------- zone e liste

func rdtyZoneLen(o *rdtyOpt) int { return 16 + 2*o.zoneMem }

func rdtyZoneOff(o *rdtyOpt, z int) int { return rdtyZoneTab + 32 + z*rdtyZoneLen(o) }

func rdtyZoneValid(img []byte, z int) bool { return img[rdtyZoneTab+z/8]>>(z&7)&1 == 1 }

// rdtyUnlink toglie il canale i (numero i+1) da zone e liste di scansione,
// compattando gli elenchi
func rdtyUnlink(o *rdtyOpt, img []byte, i int) {
	cnum := uint16(i + 1)
	drop := func(list []byte, n int, v uint16) {
		var keep []uint16
		for k := 0; k < n; k++ {
			x := binary.LittleEndian.Uint16(list[2*k:])
			if x != 0 && x != v {
				keep = append(keep, x)
			}
		}
		for k := 0; k < n; k++ {
			var x uint16
			if k < len(keep) {
				x = keep[k]
			}
			binary.LittleEndian.PutUint16(list[2*k:], x)
		}
	}
	for z := 0; z < o.nzones; z++ {
		if rdtyZoneValid(img, z) {
			drop(img[rdtyZoneOff(o, z)+16:], o.zoneMem, cnum)
		}
	}
	for s := 0; s < o.nscan; s++ {
		if img[o.scanOff+s] == 0 || img[o.scanOff+s] == 0xFF {
			continue
		}
		sl := img[o.scanOff+o.scanValid+s*rdtyScanLen:][:rdtyScanLen]
		// primo elemento: «canale selezionato» (1), poi i canali come numero+1
		drop(sl[18:], 31, cnum+1)
		for _, p := range []int{80, 82, 84} {
			if binary.LittleEndian.Uint16(sl[p:]) == cnum+1 {
				binary.LittleEndian.PutUint16(sl[p:], 0)
			}
		}
	}
}

// rdtyZoneAdd: canale nuovo nella prima zona con posto libero, altrimenti in una
// zona nuova («Zone n»)
func rdtyZoneAdd(o *rdtyOpt, img []byte, i int) {
	cnum := uint16(i + 1)
	for z := 0; z < o.nzones; z++ {
		if !rdtyZoneValid(img, z) {
			continue
		}
		mem := img[rdtyZoneOff(o, z)+16:]
		for k := 0; k < o.zoneMem; k++ {
			if binary.LittleEndian.Uint16(mem[2*k:]) == cnum {
				return
			}
		}
	}
	for z := 0; z < o.nzones; z++ {
		if !rdtyZoneValid(img, z) {
			continue
		}
		mem := img[rdtyZoneOff(o, z)+16:]
		for k := 0; k < o.zoneMem; k++ {
			if binary.LittleEndian.Uint16(mem[2*k:]) == 0 {
				binary.LittleEndian.PutUint16(mem[2*k:], cnum)
				return
			}
		}
	}
	for z := 0; z < o.nzones; z++ {
		if rdtyZoneValid(img, z) {
			continue
		}
		zr := img[rdtyZoneOff(o, z):][:rdtyZoneLen(o)]
		for k := range zr {
			zr[k] = 0
		}
		name := fmt.Sprintf("Zone %d", z+1)
		for k := 0; k < 16; k++ {
			zr[k] = 0xFF
		}
		copy(zr, name)
		binary.LittleEndian.PutUint16(zr[16:], cnum)
		img[rdtyZoneTab+z/8] |= 1 << (z & 7)
		return
	}
}

func (f rdtyFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, img[rdtyRecOff(i):][:rdtyChSize]...)
}

func (f rdtyFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(img[rdtyRecOff(i):][:rdtyChSize], raw)
	was := rdtyValid(img, i)
	rdtySetValid(img, i, true)
	if !was {
		rdtyZoneAdd(rdtyo(m), img, i)
	}
}

func (f rdtyFam) Match(m *genModel, img []byte) bool {
	return len(img) == rdtyMemSize && strings.HasPrefix(string(img), rdtyo(m).hdr)
}

// ---------------------------------------------------------------- protocollo

type rdtySess struct {
	ctx  context.Context
	p    Port
	bank int
}

func (s *rdtySess) xfer(data []byte, rlen int) ([]byte, error) {
	var buf [42]byte
	buf[0], buf[1] = 1, 0
	buf[2], buf[3] = byte(len(data)), byte(len(data)>>8)
	copy(buf[4:], data)
	if err := s.p.Write(buf[:]); err != nil {
		return nil, err
	}
	var rb [42]byte
	if err := s.p.ReadFull(s.ctx, rb[:], 2*time.Second); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	if rb[0] != 3 || rb[1] != 0 || rb[3] != 0 || int(rb[2]) != rlen {
		return nil, rErr("risposta della radio non valida (% X)", rb[:4])
	}
	return append([]byte{}, rb[4:4+rlen]...), nil
}

func (s *rdtySess) ack(data []byte, what string) error {
	r, err := s.xfer(data, 1)
	if err != nil {
		return err
	}
	if r[0] != 'A' {
		return rErr("la radio non ha confermato %s", what)
	}
	return nil
}

func (s *rdtySess) enter(m *genModel) (string, error) {
	if err := s.ack([]byte("\x02PROGRA"), "la modalità di programmazione"); err != nil {
		return "", err
	}
	id, err := s.xfer([]byte("M\x02"), 16)
	if err != nil {
		return "", err
	}
	if err := s.ack([]byte("A"), "l'identificazione"); err != nil {
		return "", err
	}
	n := bytes.IndexByte(id, 0xFF)
	if n < 0 {
		n = len(id)
	}
	ident := string(id[:n])
	if ident != rdtyo(m).ident {
		return ident, rErr("modello errato: la radio si identifica come %q", ident)
	}
	ver := strings.TrimRight(string(id[8:12]), "\x00\xff")
	s.bank = -1
	return ident + " " + ver, nil
}

func (s *rdtySess) selBank(addr int) error {
	b := 0
	if addr >= 0x10000 {
		b = 1
	}
	if b == s.bank {
		return nil
	}
	if err := s.ack([]byte{'C', 'W', 'B', 4, 0, byte(b), 0, 0}, "il cambio di banco"); err != nil {
		return err
	}
	s.bank = b
	return nil
}

func rdtyOpen(ctx context.Context) (*rdtySess, error) {
	p, err := openPortFunc(fmt.Sprintf("HID:%04X:%04X", rdtyVID, rdtyPID), 0)
	if err != nil {
		return nil, err
	}
	return &rdtySess{ctx: ctx, p: p}, nil
}

func rdtySkip(bno int) bool { return bno >= 248 && bno < 256 }

func (f rdtyFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rdtyo(m)
	s, err := rdtyOpen(ctx)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.nblk, "Collegamento con la radio…")
	info, err := s.enter(m)
	if err != nil {
		return nil, info, genFail("Identificazione", err)
	}
	img := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
	for bno := 1; bno < o.nblk; bno++ {
		if rdtySkip(bno) {
			continue
		}
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		prog(bno, o.nblk, "Lettura…")
		addr := bno * 128
		if err := s.selBank(addr); err != nil {
			return nil, info, err
		}
		for n := 0; n < 128; n += 32 {
			a := addr + n
			r, err := s.xfer([]byte{'R', byte(a >> 8), byte(a), 32}, 36)
			if err != nil {
				return nil, info, err
			}
			copy(img[a:], r[4:])
		}
	}
	_ = s.ack([]byte("ENDR"), "la fine della lettura")
	copy(img, o.hdr)
	prog(o.nblk, o.nblk, "Lettura completata")
	return img, info, nil
}

func (f rdtyFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rdtyo(m)
	if !f.Match(m, img) {
		return "", fmt.Errorf("i dati non sono una memoria della %s", m.Name)
	}
	s, err := rdtyOpen(ctx)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.nblk, "Collegamento con la radio…")
	info, err := s.enter(m)
	if err != nil {
		return info, genFail("Identificazione", err)
	}
	for bno := 1; bno < o.nblk; bno++ {
		if rdtySkip(bno) {
			continue
		}
		if ctx.Err() != nil {
			return info, errCancelled
		}
		prog(bno, o.nblk, "Scrittura…")
		addr := bno * 128
		if err := s.selBank(addr); err != nil {
			return info, err
		}
		for n := 0; n < 128; n += 32 {
			a := addr + n
			if err := s.ack(append([]byte{'W', byte(a >> 8), byte(a), 32}, img[a:a+32]...), fmt.Sprintf("il blocco 0x%05X", a)); err != nil {
				return info, err
			}
		}
	}
	if err := s.ack([]byte("ENDW"), "la fine della scrittura"); err != nil {
		return info, err
	}
	prog(o.nblk, o.nblk, "Scrittura completata")
	return info, nil
}
