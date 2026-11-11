//go:build !windows

// IU0VYT Flash - gestione canali, ponti e firmware di radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

// 4.16: radio simulata con il firmware OpenGD77 (protocollo descritto in qdmr):
// EEPROM, flash a settori di 4096 byte con il buffer, comandi dello schermo.
type ogdEmu struct {
	mu    sync.Mutex
	o     *ogdOpt
	typ   byte
	eep   []byte
	flash []byte
	buf   []byte
	sec   int
	bad   []string // scritture fuori da canali e zone
	wrote int
	cmds  []byte
}

func (ogdFam) emu(m *genModel, img []byte) genEmu {
	o := ogdo(m)
	e := &ogdEmu{o: o, typ: o.types[0], eep: bytes.Repeat([]byte{0xA5}, 0x10000), flash: bytes.Repeat([]byte{0x5A}, o.flash+ogdFlashLen+ogdSector), sec: -1}
	full := append([]byte{}, img...)
	if len(full) < ogdSize {
		full = append(full, bytes.Repeat([]byte{0xFF}, ogdSize-len(full))...)
	}
	for _, r := range ogdEepRanges {
		copy(e.eep[r[0]:r[1]], full[r[0]:r[1]])
	}
	copy(e.flash[o.flash:], full[ogdEep:ogdSize])
	return e
}

func (e *ogdEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	img := bytes.Repeat([]byte{0xFF}, ogdSize)
	for _, r := range ogdEepRanges {
		copy(img[r[0]:r[1]], e.eep[r[0]:r[1]])
	}
	copy(img[ogdEep:], e.flash[e.o.flash:e.o.flash+ogdFlashLen])
	return img
}

func (e *ogdEmu) serve(p Port) error {
	for {
		c, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		switch {
		case c[0] == 'C':
			b, err := emuRead(p, 22)
			if err != nil {
				return nil
			}
			e.mu.Lock()
			e.cmds = append(e.cmds, b[0])
			e.mu.Unlock()
			p.Write([]byte{'-'})
		case c[0] == 'R':
			b, err := emuRead(p, 7)
			if err != nil {
				return nil
			}
			addr, n := int(binary.BigEndian.Uint32(b[1:])), int(binary.BigEndian.Uint16(b[5:]))
			d := make([]byte, n)
			e.mu.Lock()
			switch b[0] {
			case 1:
				if addr+n <= len(e.flash) {
					copy(d, e.flash[addr:])
				}
			case 2:
				if addr+n <= len(e.eep) {
					copy(d, e.eep[addr:])
				}
			case 9:
				binary.LittleEndian.PutUint32(d[0:], 3)
				binary.LittleEndian.PutUint32(d[4:], uint32(e.typ))
				copy(d[8:], "R20240908")
				copy(d[24:], "20240908120000")
			}
			e.mu.Unlock()
			p.Write(append([]byte{'R', byte(n >> 8), byte(n)}, d...))
		case c[0] == e.o.wr:
			k, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			switch k[0] {
			case 1:
				b, err := emuRead(p, 3)
				if err != nil {
					return nil
				}
				e.mu.Lock()
				e.sec = int(b[0])<<16 | int(b[1])<<8 | int(b[2])
				e.buf = append([]byte{}, e.flash[e.sec*ogdSector:][:ogdSector]...)
				e.mu.Unlock()
			case 2, 4:
				h, err := emuRead(p, 6)
				if err != nil {
					return nil
				}
				addr, n := int(binary.BigEndian.Uint32(h)), int(binary.BigEndian.Uint16(h[4:]))
				d, err := emuRead(p, n)
				if err != nil {
					return nil
				}
				e.mu.Lock()
				if k[0] == 4 {
					if !ogdWritable(addr) || !ogdWritable(addr+n-1) {
						e.bad = append(e.bad, fmt.Sprintf("EEPROM 0x%05X", addr))
					}
					copy(e.eep[addr:], d)
					e.wrote++
				} else if e.sec < 0 || addr/ogdSector != e.sec {
					e.bad = append(e.bad, fmt.Sprintf("flash 0x%05X fuori dal settore", addr))
				} else {
					if off := ogdEep + addr - e.o.flash; !ogdWritable(off) || !ogdWritable(off+n-1) {
						e.bad = append(e.bad, fmt.Sprintf("flash 0x%05X", addr))
					}
					copy(e.buf[addr-e.sec*ogdSector:], d)
					e.wrote++
				}
				e.mu.Unlock()
			case 3:
				e.mu.Lock()
				if e.sec >= 0 {
					copy(e.flash[e.sec*ogdSector:], e.buf)
				}
				e.sec = -1
				e.mu.Unlock()
			default:
				continue
			}
			p.Write([]byte{e.o.wr, k[0]})
		}
	}
}
