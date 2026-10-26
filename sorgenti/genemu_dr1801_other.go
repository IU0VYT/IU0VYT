//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Baofeng DR-1801UV (protocollo in drv_dr1801.go) e memoria di
// prova con la stessa struttura di quella di fabbrica (moduli BI, ZD, SL, CH, RM
// con i parametri del file «DM-1801A6 Factory setting.accps»).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

// drSynthImage: intestazione e corpo con zone, liste di scansione, canali (1024
// da 52 byte, nomi da 20) e roaming vuoti
func drSynthImage() []byte {
	type mod struct {
		code   string
		params []byte
		size   int
	}
	mods := []mod{
		{"BI", []byte{0x18, 0x18, 0x18, 0x18, 0x18, 0x20, 0, 0, 0, 0, 0, 0}, 176},
		{"ZD", []byte{0x96, 0x00, 0x00, 0x20, 0x68, 0x00, 0x20, 0x00, 0, 0, 0, 0}, 8 + 150*104},
		{"SL", []byte{0x10, 0x20, 0x24, 0x03, 0x0a, 0x00, 0x50, 0x00, 0, 0, 0, 0}, 4 + 10*80},
		{"CH", []byte{0x00, 0x04, 0x10, 0x00, 0x34, 0x00, 0x14, 0x00, 0, 0, 0, 0}, 4 + 1024*52 + 1024*20},
		{"RM", []byte{0, 0, 0, 0}, 4 + 8*20},
	}
	const hl, mo = 768, 320
	body := 0
	for _, m := range mods {
		body += m.size
	}
	img := make([]byte, hl+4+body)
	copy(img, "AUCTUSA6")
	binary.LittleEndian.PutUint16(img[8:], hl)
	binary.LittleEndian.PutUint16(img[10:], drDataVer)
	binary.LittleEndian.PutUint16(img[16:], uint16(len(mods)))
	binary.LittleEndian.PutUint16(img[18:], mo)
	binary.LittleEndian.PutUint32(img[20:], uint32(len(img)-4))
	copy(img[24:], []byte{1, 0, 1, 0, 0, 1})
	copy(img[mo:], "#MP#")
	copy(img[hl:], "#BD#")
	o, off := mo+4, 0
	for k, m := range mods {
		img[o], img[o+1] = byte(k+1), 1
		copy(img[o+2:], m.code)
		img[o+4], img[o+5], img[o+6] = byte(off>>16), byte(off>>8), byte(off)
		img[o+7] = byte(len(m.params))
		copy(img[o+8:], m.params)
		o += 8 + len(m.params)
		if m.code == "BI" {
			copy(img[hl+4+off+48:], "A6-0000-XXXX")
			copy(img[hl+4+off+24:], "BF1801")
			copy(img[hl+4+off+96:], "136M-174M,400M-480M")
		}
		if m.code == "CH" {
			for p := 0; p < 1024; p++ {
				binary.LittleEndian.PutUint16(img[hl+4+off+4+p*52:], 0xFFFF)
			}
		}
		off += m.size
	}
	return img
}

type drEmu struct {
	mu  sync.Mutex
	img []byte
	// scrittura annunciata: i dati arrivano dopo la riapertura della porta
	wantLen int
	wantCrc [2]byte
}

func (drFam) emu(m *genModel, img []byte) genEmu {
	if _, err := drParse(img); err != nil {
		img = drSynthImage()
	}
	return &drEmu{img: append([]byte{}, img...)}
}

func (e *drEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *drEmu) serve(p Port) error {
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if e.wantLen > 0 {
			// dati della scrittura (senza pacchetto)
			rest, err := emuRead(p, e.wantLen-1)
			if err != nil {
				return err
			}
			body := append([]byte{c}, rest...)
			e.wantLen = 0
			res := []byte{1, 0, 0}
			if drCrc(body) != e.wantCrc {
				res = []byte{0, 0, 1}
			} else {
				e.mu.Lock()
				l, _ := drParse(e.img)
				e.img = append(e.img[:l.hl+4:l.hl+4], body...)
				e.mu.Unlock()
			}
			p.Write(drReply(0x8103, res))
			continue
		}
		if c != 0xAA {
			return fmt.Errorf("byte inatteso %02X", c)
		}
		n, err := emuByte(p)
		if err != nil {
			return err
		}
		rest, err := emuRead(p, int(n)-2)
		if err != nil {
			return err
		}
		pk := append([]byte{0xAA, n}, rest...)
		if !bytes.Equal(drPkt(binary.LittleEndian.Uint16(pk[2:]), pk[4:len(pk)-2]), pk) {
			return fmt.Errorf("pacchetto non valido % X", pk)
		}
		data := pk[4 : len(pk)-2]
		switch binary.LittleEndian.Uint16(pk[2:]) {
		case 0x0401:
			p.Write(drReply(0x8104, []byte{1}))
		case 0x2B00:
			p.Write(drReply(0x802B, []byte{1}))
		case 0x0001:
			e.mu.Lock()
			img := e.img
			e.mu.Unlock()
			r := make([]byte, 15)
			r[0] = 1
			binary.BigEndian.PutUint32(r[1:], uint32(len(img)))
			binary.BigEndian.PutUint32(r[5:], drDataVer)
			crc := drCrc(img)
			r[13], r[14] = crc[0], crc[1]
			p.Write(drReply(0x8100, r))
		case 0x0101:
			p.Write(e.image())
		case 0x0201:
			e.wantLen = int(binary.BigEndian.Uint32(data[2:]))
			e.wantCrc = [2]byte{data[6], data[7]}
			p.Write(drReply(0x8102, []byte{1, 0}))
		default:
			return fmt.Errorf("comando sconosciuto % X", pk[2:4])
		}
	}
}

// risposta: comando in big endian
func drReply(cmd uint16, data []byte) []byte {
	b := drPkt(cmd, data)
	b[2], b[3] = byte(cmd>>8), byte(cmd)
	b[len(b)-2] = byte(len(b)) ^ byte(cmd) ^ byte(cmd>>8)
	for _, x := range data {
		b[len(b)-2] ^= x
	}
	return b
}
