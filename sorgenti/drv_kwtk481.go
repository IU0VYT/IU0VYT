// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tk481.py, tk280.py (4.23).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.23: Kenwood TK-481 (palmare) e TK-981 (veicolare), radio a 900 MHz con i
// canali divisi in «sistemi» (tk481.py di CHIRP). Stesso collegamento e stessa
// memoria di 32 KiB delle TK-280 / TK-380 (drv_kwtk280.go).
//
// A 0x0000 la tabella degli indirizzi dei 32 sistemi (16 bit little endian, FFFF
// = il sistema non c'è); a 0x40 il loro numero; a 0x4E un bit per sistema: 1 se è
// a risorse condivise (trunking). Ogni sistema è un'intestazione di 32 byte
// (numero, numero di canali, nome) seguita dai suoi canali di 32 byte; i sistemi
// stanno uno dopo l'altro: aggiungere o togliere un canale sposta di 32 byte
// tutto quello che segue (_expand_system e _reduce_system di CHIRP, rifatti
// uguali). Un canale nuovo va in coda al suo sistema.
// La lista ha 250 posizioni per ognuno dei 32 sistemi: la posizione è il numero
// del canale nel sistema. I sistemi non si creano e non si cancellano; quelli a
// risorse condivise si leggono ma non si modificano.
//
// Canale: 0 numero; 1-4 RX e 5-8 TX in BCD little endian, decine di Hz (TX a FF =
// sola ricezione); 9 e 10 passo di sintonia nei 4 bit bassi, 3 negli alti;
// 11-12 tono RX, 13-14 tono TX (come le TK-280); 15-24 nome; 26 bit 0x08 alta
// potenza; 27 bit 0x20 «talk around» (acceso se TX è diverso da RX), 0x10
// chiamata, 0x04 in scansione.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	kwx481Systems = 32
	kwx481Per     = 250
	kwx481Rec     = 32
	kwx481Trunk   = 0x4E
)

type kwx481Fam struct{}

type kwx481Sys struct{ addr, count int }

func init() {
	mk := func(id, model, class, typ string, pw []genPower) *genModel {
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: kwx481Systems * kwx481Per, First: 1, Zones: kwx481Systems, NameMax: 10,
			Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz 1234567890_-*()/\\+=.",
			Powers:  pw, Bands: []band{{896000000, 941000000}}, Narrow: false, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kwx481Fam{}, Opt: &kwx80Opt{typ: typ}, Notes: strings.TrimSpace(kwSysCatNote)}
	}
	genRegister(
		mk("kw-tk481", "TK-481", "TK481", "PG481", []genPower{{"Bassa 1 W", 1, 2}, {"Alta 2,5 W", 2.5, 0}}),
		mk("kw-tk981", "TK-981", "TK981", "M0981", []genPower{{"Unica", 0, 0}}),
	)
}

// systems: i 32 sistemi (addr < 0 se il sistema non c'è); ok è falso se la
// tabella non è leggibile
func (kwx481Fam) systems(img []byte) (ss []kwx481Sys, ok bool) {
	if len(img) != kwx80Size {
		return nil, false
	}
	ss = make([]kwx481Sys, kwx481Systems)
	for i := range ss {
		a := int(binary.LittleEndian.Uint16(img[2*i:]))
		if a == 0xFFFF {
			ss[i].addr = -1
			continue
		}
		if a < 0x60 || a+kwx481Rec > kwx80Size {
			return nil, false
		}
		n := int(img[a+1])
		if a+kwx481Rec*(n+1) > kwx80Size {
			return nil, false
		}
		ss[i] = kwx481Sys{a, n}
	}
	return ss, true
}

func (s kwx481Sys) rec(img []byte, k int) []byte {
	o := s.addr + kwx481Rec*(k+1)
	return img[o : o+kwx481Rec]
}

// find: posizione nel sistema del canale numero n, -1 se non c'è
func (s kwx481Sys) find(img []byte, n int) int {
	for k := 0; k < s.count; k++ {
		if int(s.rec(img, k)[0]) == n {
			return k
		}
	}
	return -1
}

func kwx481Trunked(img []byte, i int) bool { return img[kwx481Trunk+i/8]>>(i%8)&1 == 1 }

// slot: sistema e posizione del canale della posizione i della lista
func (f kwx481Fam) slot(img []byte, i int) (kwx481Sys, int) {
	ss, ok := f.systems(img)
	if !ok || ss[i/kwx481Per].addr < 0 {
		return kwx481Sys{addr: -1}, -1
	}
	s := ss[i/kwx481Per]
	return s, s.find(img, i%kwx481Per+1)
}

func (f kwx481Fam) Get(m *genModel, img []byte, i int) *Chan {
	s, k := f.slot(img, i)
	if k < 0 {
		return nil
	}
	r := s.rec(img, k)
	// la radio lavora solo a banda stretta (in CHIRP il modo è sempre NFM)
	c := &Chan{Rx: kwx180Bcd(r[1:5]) * 10, Narrow: true}
	c.Tx = c.Rx
	if r[5] == 0xFF {
		c.TxOff = true
	} else {
		c.Tx = kwx180Bcd(r[5:9]) * 10
	}
	c.RxTone = kw60gToneDec(int64(binary.LittleEndian.Uint16(r[11:])))
	c.TxTone = kw60gToneDec(int64(binary.LittleEndian.Uint16(r[13:])))
	c.Name = strings.TrimRight(string(r[15:25]), " \xff")
	if len(m.Powers) > 1 {
		c.Power = m.powerSem(int(r[26] >> 3 & 1))
	}
	c.Scan = r[27]&0x04 != 0
	return c
}

func (f kwx481Fam) Rec(m *genModel, img []byte, i int) []byte {
	if s, k := f.slot(img, i); k >= 0 {
		return append([]byte{}, s.rec(img, k)...)
	}
	return bytes.Repeat([]byte{0xFF}, kwx481Rec)
}

func (f kwx481Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if s, k := f.slot(img, i); k >= 0 && len(raw) == kwx481Rec {
		n := s.rec(img, k)[0]
		copy(s.rec(img, k), raw)
		s.rec(img, k)[0] = n
	}
}

func (f kwx481Fam) Match(m *genModel, img []byte) bool { return kwx80Fam{}.Match(m, img) }

// posizioni dei sistemi che la radio non ha, o a risorse condivise: non si possono usare
func (f kwx481Fam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	ss, ok := f.systems(img)
	return !ok || ss[i/kwx481Per].addr < 0 || kwx481Trunked(img, i/kwx481Per)
}

func (f kwx481Fam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	ss, ok := f.systems(img)
	switch {
	case !ok:
		return fmt.Errorf("la tabella dei sistemi della radio non è leggibile: la memoria non si può modificare")
	case ss[i/kwx481Per].addr < 0:
		if want == nil {
			return nil
		}
		return fmt.Errorf("la radio non ha il sistema %d: va creato con il programma del costruttore", i/kwx481Per+1)
	case kwx481Trunked(img, i/kwx481Per):
		return fmt.Errorf("il sistema %d è a risorse condivise (trunking): non si modifica da qui", i/kwx481Per+1)
	}
	return nil
}

// shift: sposta di delta byte (±32) gli indirizzi dei sistemi che seguono idx
func (kwx481Fam) shift(img []byte, idx, delta int) {
	for j := idx + 1; j < kwx481Systems; j++ {
		if a := binary.LittleEndian.Uint16(img[2*j:]); a != 0xFFFF {
			binary.LittleEndian.PutUint16(img[2*j:], uint16(int(a)+delta))
		}
	}
}

func kwx481Tone(s string) (uint16, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, err
	}
	switch t.Mode {
	case "Tone":
		return uint16(t.Hz10), nil
	case "DTCS":
		var v int
		fmt.Sscanf(fmt.Sprintf("%d", t.Code), "%o", &v)
		v += 0x2800
		if t.Inv {
			v += 0xA000
		}
		return uint16(v), nil
	}
	return 0xFFFF, nil
}

func (f kwx481Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if err := f.LockedImg(m, img, i, nil, c); err != nil {
		return err
	}
	idx, n := i/kwx481Per, i%kwx481Per+1
	s, k := f.slot(img, i)
	if c == nil {
		if k < 0 {
			return nil
		}
		// _reduce_system: i canali che seguono scendono di un posto, tutto il
		// resto della memoria arretra di 32 byte
		end := s.addr + kwx481Rec*(s.count+1)
		copy(img[s.addr+kwx481Rec*(k+1):], img[s.addr+kwx481Rec*(k+2):end])
		f.shift(img, idx, -kwx481Rec)
		copy(img[end-kwx481Rec:], img[end:])
		copy(img[len(img)-kwx481Rec:], bytes.Repeat([]byte{0xFF}, kwx481Rec))
		img[s.addr+1]--
		return nil
	}
	rt, err := kwx481Tone(c.RxTone)
	if err != nil {
		return err
	}
	tt, err := kwx481Tone(c.TxTone)
	if err != nil {
		return err
	}
	if k < 0 {
		// _expand_system: posto per un canale in coda al sistema
		if s.count >= kwx481Per {
			return fmt.Errorf("il sistema %d è pieno", idx+1)
		}
		if !bytes.Equal(img[len(img)-kwx481Rec:], bytes.Repeat([]byte{0xFF}, kwx481Rec)) {
			return fmt.Errorf("la memoria della radio è piena")
		}
		end := s.addr + kwx481Rec*(s.count+1)
		f.shift(img, idx, kwx481Rec)
		copy(img[end+kwx481Rec:], img[end:len(img)-kwx481Rec])
		copy(img[end:end+kwx481Rec], bytes.Repeat([]byte{0xFF}, kwx481Rec))
		img[s.addr+1]++
		s.count++
		k = s.count - 1
	}
	r := s.rec(img, k)
	r[0] = byte(n)
	kwx180PutBcd(r[1:5], c.Rx/10)
	r[9] = 0x30 | byte(kwx80Step(c.Rx/10*10))
	if c.TxOff {
		copy(r[5:9], []byte{0xFF, 0xFF, 0xFF, 0xFF})
		r[10] = 0x30 | 1
	} else {
		kwx180PutBcd(r[5:9], c.Tx/10)
		r[10] = 0x30 | byte(kwx80Step(c.Tx/10*10))
	}
	binary.LittleEndian.PutUint16(r[11:], rt)
	binary.LittleEndian.PutUint16(r[13:], tt)
	name := []byte(c.Name)
	if len(name) > 10 {
		name = name[:10]
	}
	copy(r[15:25], append(name, bytes.Repeat([]byte{' '}, 10-len(name))...))
	if len(m.Powers) > 1 {
		r[26] = r[26]&^0x08 | byte(m.powerIndex(c.Power))<<3
	}
	b := r[27] &^ 0x24
	if c.TxOff || c.Tx != c.Rx {
		b |= 0x20
	}
	if c.Scan {
		b |= 0x04
	}
	r[27] = b
	return nil
}

// ---------------------------------------------------------------- protocollo (quello delle TK-280 / TK-380)

func (f kwx481Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return kwx80Fam{}.Download(ctx, m, port, prog)
}

func (f kwx481Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return kwx80Fam{}.Upload(ctx, m, port, img, prog)
}
