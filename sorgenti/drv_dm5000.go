// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bajeton DM-5000 (radio DMR): protocollo e memoria ricavati dal programma del
// costruttore «BAJETON DM-5000 CPS» (versione 250726, .NET, progetto
// «DM550_NOLCD»), pubblicato sulla pagina dei download di bajeton.com.
//
// Protocollo (115200 baud, 8N1):
//   «34 52 05 10 9B» → 06 (il programma riprova 3 volte a 200 ms)
//   lettura di un blocco da 1 KB: 'R' n(2, big endian) somma(1) →
//     'R' n(2) + 1024 byte + somma (somma = byte meno significativo della somma
//     dei byte precedenti)
//   scrittura: comando(1) n(2) + 1024 byte + somma → 06
//   fine: «34 52 05 EE 79»
// Aree (blocchi letti → comando e blocchi scritti):
//   impostazioni   8-11    → 0x90 0-3   (nel blocco 8: byte 0-11 password, 960-1023 della radio)
//   canali        16-63    → 0x91 0-47  (1024 canali da 48 byte)
//   zone         112-239   → 0x92 0-127 (256 zone da 512 byte)
//   contatti     368-431   → 0x93 0-63
//   gruppi       496-519   → 0x94 0-23
//   cifratura    520-531   → 0x95 0-11
//   messaggi     592-595   → 0x97 0-3
//   radio FM     856       → 0x99 0
// Prima di scrivere il programma legge il blocco 0 e il blocco 8 (controllo
// della password).
//
// Canale (48 byte): 0 ID (digitale: 0 della radio, 1 del canale) / AM-FM
// (analogico), 1 = 1, 2 tipo (0 digitale, 1 analogico), 3 slot / banda (0 larga,
// 1 stretta), 4-5 color code e modo DMR / tono RX, 6-9 RX e 10-13 TX in decine di
// Hz (little endian), 14 monitor / 14-15 tono TX, 16 potenza (0 bassa, 1 alta),
// 17 priorità, 18 (analogico) tipo di squelch << 5 | TOT, 19 bit 7 = 1 escluso
// dalla scansione (analogico: tono di coda << 4 | scrambler), 20-31 parametri
// digitali (TOT, allarme, lista dei gruppi, contatto, cifratura, ID del canale
// in BCD) o codici di silenzio analogici, 32-47 nome (GBK, FF in coda).
// Toni (16 bit, little endian): 0 nessuno, 0x1000 | decimi di Hz (CTCSS),
// 0x2000 | codice DCS ottale (normale), 0x3000 | codice (invertito).
// Zona (512 byte): 0 e 1 numero di canali, 4-19 nome, 20-419 fino a 200 canali
// (16 bit little endian, numero del canale da 0; 0xFFFF libero).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type dm5kArea struct {
	name        string
	read, count int  // primo blocco letto e numero di blocchi
	cmd         byte // comando di scrittura
}

var dm5kAreas = []dm5kArea{
	{"impostazioni", 8, 4, 0x90}, {"canali", 16, 48, 0x91}, {"zone", 112, 128, 0x92}, {"contatti", 368, 64, 0x93},
	{"gruppi", 496, 24, 0x94}, {"cifratura", 520, 12, 0x95}, {"messaggi", 592, 4, 0x97}, {"radio FM", 856, 1, 0x99},
}

const (
	dm5kBlock    = 1024
	dm5kChSize   = 48
	dm5kSlots    = 1024
	dm5kZoneSize = 512
	dm5kZones    = 256
	dm5kZoneMax  = 200
)

// posizione di un'area nell'immagine
func dm5kOff(name string) int {
	o := 0
	for _, a := range dm5kAreas {
		if a.name == name {
			return o
		}
		o += a.count * dm5kBlock
	}
	panic("area sconosciuta " + name)
}

func dm5kSize() int {
	n := 0
	for _, a := range dm5kAreas {
		n += a.count
	}
	return n * dm5kBlock
}

type dm5kOpt struct{}

type dm5kFam struct{}

func init() {
	genRegister(&genModel{ID: "bj-dm5000", Vendor: "Bajeton", Model: "DM-5000", RClass: "IU0VYT_BJDM5000", Name: "Bajeton DM-5000",
		Slots: dm5kSlots, First: 1, NameMax: 16, Charset: charsetASCII,
		Powers: []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}},
		Bands:  []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
		DCS: dtcsCodes, Baud: 115200, Fam: dm5kFam{}, Opt: &dm5kOpt{}, Digital: "DMR",
		Notes: "Radio DMR. Protocollo e memoria ricavati dal programma del costruttore «BAJETON DM-5000 CPS» (versione 250726). " +
			"Qui si modificano frequenze, nome, potenza, scansione e toni; per i canali DMR gruppo, contatto, slot e color code restano " +
			"quelli impostati (un canale DMR nuovo prende i valori predefiniti del programma: primo gruppo e primo contatto, slot 1, color code 0). " +
			"Un canale nuovo viene aggiunto alla zona 1 se non è in nessuna zona; un canale cancellato viene tolto dalle zone. " +
			"Contatti, gruppi, messaggi e impostazioni della radio non si toccano. Se sulla radio è impostata una password, toglila prima con il programma del costruttore."})
}

func (dm5kFam) rec(img []byte, i int) []byte {
	o := dm5kOff("canali") + i*dm5kChSize
	return img[o : o+dm5kChSize]
}

func dm5kToneDec(v uint16) string {
	n := int(v & 0x0FFF)
	switch {
	case v&0xF000 == 0x1000:
		return fmtTone(n)
	case v&0xF000 == 0x2000:
		return genTone{Mode: "DTCS", Code: dm5kOct(n)}.String()
	case v&0x3000 == 0x3000:
		return genTone{Mode: "DTCS", Code: dm5kOct(n), Inv: true}.String()
	}
	return ""
}

// valore ottale → codice DCS scritto in cifre (0o23 → 23)
func dm5kOct(n int) int { return (n>>6)*100 + ((n>>3)&7)*10 + n&7 }

func dm5kToneEnc(t genTone) uint16 {
	switch t.Mode {
	case "Tone":
		return uint16(t.Hz10&0x0FFF) | 0x1000
	case "DTCS":
		c := t.Code
		v := uint16((c/100)*64+(c/10%10)*8+c%10) & 0x0FFF
		if t.Inv {
			return v | 0x3000
		}
		return v | 0x2000
	}
	return 0
}

func (f dm5kFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if r[2] != 0 && r[2] != 1 {
		return nil
	}
	c := &Chan{Rx: int64(binary.LittleEndian.Uint32(r[6:])) * 10, Tx: int64(binary.LittleEndian.Uint32(r[10:])) * 10}
	c.Name = strings.TrimRight(string(bytes.TrimRight(r[32:48], "\xff\x00")), " ")
	if r[16] == 1 {
		c.Power = m.powerSem(0)
	} else {
		c.Power = m.powerSem(1)
	}
	c.Scan = r[19]&0x80 == 0
	if r[2] == 0 {
		c.Dig = true
		return c
	}
	c.Narrow = r[3] == 1
	c.RxTone = dm5kToneDec(binary.LittleEndian.Uint16(r[4:]))
	c.TxTone = dm5kToneDec(binary.LittleEndian.Uint16(r[14:]))
	return c
}

func (f dm5kFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f dm5kFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (f dm5kFam) Match(m *genModel, img []byte) bool { return false }

// ---- zone

func (dm5kFam) zone(img []byte, z int) []byte {
	o := dm5kOff("zone") + z*dm5kZoneSize
	return img[o : o+dm5kZoneSize]
}

func (f dm5kFam) zoneMembers(img []byte, z int) []int {
	r := f.zone(img, z)
	var out []int
	for j := 0; j < dm5kZoneMax; j++ {
		if v := int(binary.LittleEndian.Uint16(r[20+2*j:])); v < dm5kSlots {
			out = append(out, v)
		}
	}
	return out
}

// come il programma: elenco compatto, posizioni libere a FFFF, numero nei byte 0 e 1
func (f dm5kFam) setZoneMembers(img []byte, z int, mem []int) {
	r := f.zone(img, z)
	for j := 0; j < dm5kZoneMax; j++ {
		v := uint16(0xFFFF)
		if j < len(mem) {
			v = uint16(mem[j])
		}
		binary.LittleEndian.PutUint16(r[20+2*j:], v)
	}
	r[0], r[1] = byte(len(mem)), byte(len(mem))
	if bytes.Equal(r[4:20], bytes.Repeat([]byte{0xFF}, 16)) {
		// zona mai salvata: nome predefinito del programma
		n := []byte(fmt.Sprintf("Zone-%03d", z+1))
		copy(r[4:20], append(n, bytes.Repeat([]byte{0xFF}, 16-len(n))...))
	}
}

func (f dm5kFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if c == nil {
		copy(r, bytes.Repeat([]byte{0xFF}, dm5kChSize))
		for z := 0; z < dm5kZones; z++ {
			mem := f.zoneMembers(img, z)
			if k := idxInt(mem, i); k >= 0 {
				f.setZoneMembers(img, z, append(mem[:k:k], mem[k+1:]...))
			}
		}
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
	isNew := r[2] != 0 && r[2] != 1
	typ := byte(1)
	if c.Dig {
		typ = 0
	}
	if isNew || r[2] != typ {
		// valori predefiniti del programma (InitDigital / InitAnalog)
		keepName := append([]byte{}, r[32:48]...)
		copy(r, bytes.Repeat([]byte{0xFF}, dm5kChSize))
		copy(r[32:], keepName)
		if c.Dig {
			for _, k := range []int{0, 3, 4, 5, 14, 17, 20, 21, 22, 23, 24, 25, 26, 27} {
				r[k] = 0
			}
			binary.LittleEndian.PutUint32(r[28:], 1) // ID del canale 1 (BCD)
		} else {
			for k := 0; k < 32; k++ {
				if k < 6 || k > 13 {
					r[k] = 0
				}
			}
		}
	}
	r[1], r[2] = 1, typ
	binary.LittleEndian.PutUint32(r[6:], uint32(c.Rx/10))
	binary.LittleEndian.PutUint32(r[10:], uint32(c.Tx/10))
	if m.powerIndex(c.Power) == 0 {
		r[16] = 1
	} else {
		r[16] = 0
	}
	if c.Dig {
		// il programma scrive 0 o 128 nel byte 19 dei canali digitali
		r[19] = 0
		if !c.Scan {
			r[19] = 0x80
		}
	} else {
		r[19] &^= 0x80
		if !c.Scan {
			r[19] |= 0x80
		}
		r[3] = 0
		if c.Narrow {
			r[3] = 1
		}
		binary.LittleEndian.PutUint16(r[4:], dm5kToneEnc(rt))
		binary.LittleEndian.PutUint16(r[14:], dm5kToneEnc(tt))
	}
	name := bytes.Repeat([]byte{0xFF}, 16)
	copy(name, c.Name)
	copy(r[32:], name)
	if isNew {
		in := false
		for z := 0; z < dm5kZones && !in; z++ {
			in = idxInt(f.zoneMembers(img, z), i) >= 0
		}
		if !in {
			mem := f.zoneMembers(img, 0)
			if len(mem) < dm5kZoneMax {
				f.setZoneMembers(img, 0, append(mem, i))
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

var (
	dm5kConnect = []byte{0x34, 0x52, 0x05, 0x10, 0x9B}
	dm5kEnd     = []byte{0x34, 0x52, 0x05, 0xEE, 0x79}
)

func dm5kSum(b []byte) byte {
	var s byte
	for _, c := range b {
		s += c
	}
	return s
}

type dm5kLink struct {
	ctx context.Context
	p   Port
}

func (l *dm5kLink) recv(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := l.p.ReadFull(l.ctx, b, to); err != nil {
		return nil, err
	}
	return b, nil
}

func (l *dm5kLink) connect() error {
	for k := 0; k < 3; k++ {
		_ = l.p.Purge()
		if err := l.p.Write(dm5kConnect); err != nil {
			return err
		}
		r, err := l.recv(1, 400*time.Millisecond)
		if err == nil && r[0] == 0x06 {
			return nil
		}
		if err != nil && err != errTimeout {
			return err
		}
	}
	return errTimeout
}

func (l *dm5kLink) readBlock(n int) ([]byte, error) {
	cmd := []byte{'R', byte(n >> 8), byte(n)}
	if err := l.p.Write(append(cmd, dm5kSum(cmd))); err != nil {
		return nil, err
	}
	r, err := l.recv(4+dm5kBlock, 2*time.Second)
	if err != nil {
		return nil, err
	}
	if r[0] != 'R' || r[1] != cmd[1] || r[2] != cmd[2] {
		return nil, rErr("risposta inattesa alla lettura del blocco %d", n)
	}
	if dm5kSum(r[:3+dm5kBlock]) != r[3+dm5kBlock] {
		return nil, rErr("dati del blocco %d non validi (somma di controllo)", n)
	}
	return r[3 : 3+dm5kBlock], nil
}

func (l *dm5kLink) writeBlock(cmd byte, n int, d []byte) error {
	f := append([]byte{cmd, byte(n >> 8), byte(n)}, d...)
	f = append(f, dm5kSum(f))
	if err := l.p.Write(f); err != nil {
		return err
	}
	r, err := l.recv(1, 2*time.Second)
	if err != nil {
		return err
	}
	if r[0] != 0x06 {
		return rErr("la radio ha rifiutato il blocco %d (risposta %02X)", n, r[0])
	}
	return nil
}

// password: byte 0-11 del blocco 8 tutti a FF (o a 0) se non impostata
func dm5kCheckPwd(b8 []byte) error {
	if bytes.Equal(b8[:12], bytes.Repeat([]byte{0xFF}, 12)) || bytes.Equal(b8[:12], make([]byte, 12)) {
		return nil
	}
	return rErr("sulla radio è impostata una password: toglila con il programma del costruttore")
}

func dm5kOpen(ctx context.Context, port string) (*dm5kLink, error) {
	p, err := openPortFunc(port, 115200)
	if err != nil {
		return nil, err
	}
	return &dm5kLink{ctx: ctx, p: p}, nil
}

func (f dm5kFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	l, err := dm5kOpen(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer l.p.Close()
	size := dm5kSize()
	prog(0, size, "Collegamento con la radio…")
	if err := l.connect(); err != nil {
		return nil, "", genFail("Collegamento", err)
	}
	img := make([]byte, 0, size)
	for _, a := range dm5kAreas {
		for k := 0; k < a.count; k++ {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			prog(len(img), size, "Lettura: "+a.name+"…")
			d, err := l.readBlock(a.read + k)
			if err != nil {
				return nil, "", genFail("Lettura", err)
			}
			if a.read+k == 8 {
				if err := dm5kCheckPwd(d); err != nil {
					_ = l.p.Write(dm5kEnd)
					return nil, "", genFail("Lettura", err)
				}
			}
			img = append(img, d...)
		}
	}
	_ = l.p.Write(dm5kEnd)
	prog(size, size, "Lettura completata")
	return img, "", nil
}

func (f dm5kFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	size := dm5kSize()
	if len(img) < size {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	l, err := dm5kOpen(ctx, port)
	if err != nil {
		return "", err
	}
	defer l.p.Close()
	prog(0, size, "Collegamento con la radio…")
	if err := l.connect(); err != nil {
		return "", genFail("Collegamento", err)
	}
	// come il programma: blocco 0, poi blocco 8 (password)
	if _, err := l.readBlock(0); err != nil {
		return "", genFail("Identificazione", err)
	}
	b8, err := l.readBlock(8)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	if err := dm5kCheckPwd(b8); err != nil {
		_ = l.p.Write(dm5kEnd)
		return "", genFail("Identificazione", err)
	}
	cfg := append([]byte{}, img[:dm5kBlock]...)
	// il programma conserva password (0-11) e byte 960-1023 della radio, 14-15 a FF
	copy(cfg[0:12], b8[0:12])
	cfg[14], cfg[15] = 0xFF, 0xFF
	copy(cfg[960:], b8[960:])
	off := 0
	for _, a := range dm5kAreas {
		for k := 0; k < a.count; k++ {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			prog(off, size, "Scrittura: "+a.name+"…")
			d := img[off : off+dm5kBlock]
			if off == 0 {
				d = cfg
			}
			if err := l.writeBlock(a.cmd, k, d); err != nil {
				return "", genFail("Scrittura", err)
			}
			off += dm5kBlock
		}
	}
	_ = l.p.Write(dm5kEnd)
	prog(size, size, "Scrittura completata")
	return "", nil
}
