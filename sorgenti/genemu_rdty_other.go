//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata della famiglia RD-5R / DM-1801 (protocollo HID in drv_rdty.go):
// rapporti di 42 byte, lo stesso formato del collegamento TCP usato per far
// parlare dmrconfig con la radio simulata (verifica/dmrconfig/tcpdev.c).

import (
	"bytes"
	"fmt"
	"sync"
)

type rdtyEmu struct {
	mu   sync.Mutex
	m    *genModel
	mem  []byte
	bank int
	prog bool
}

func (rdtyFam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
	copy(mem, img)
	return &rdtyEmu{m: m, mem: mem}
}

// image: la memoria come la legge il programma (intestazione e zone non lette a FF)
func (e *rdtyEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	o := rdtyo(e.m)
	img := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
	for bno := 1; bno < o.nblk; bno++ {
		if !rdtySkip(bno) {
			copy(img[bno*128:(bno+1)*128], e.mem[bno*128:])
		}
	}
	copy(img, o.hdr)
	return img
}

func (e *rdtyEmu) serve(p Port) error {
	o := rdtyo(e.m)
	ident := make([]byte, 16)
	copy(ident, bytes.Repeat([]byte{0xFF}, 8))
	copy(ident, o.ident)
	copy(ident[8:], "V210\x00\x04\x80\x04")
	for {
		var rq [42]byte
		if err := emuReadInto(p, rq[:]); err != nil {
			return nil
		}
		if rq[0] != 1 || rq[1] != 0 {
			return fmt.Errorf("rapporto HID non valido % X", rq[:4])
		}
		n := int(rq[2]) | int(rq[3])<<8
		d := rq[4 : 4+n]
		var out []byte
		e.mu.Lock()
		switch {
		case string(d) == "\x02PROGRA":
			e.prog, e.bank = true, 0
			out = []byte{'A'}
		case string(d) == "M\x02":
			out = ident
		case string(d) == "A":
			out = []byte{'A'}
		case n == 8 && string(d[:5]) == "CWB\x04\x00":
			e.bank = int(d[5])
			out = []byte{'A'}
		case n == 4 && d[0] == 'R':
			a := e.bank<<16 | int(d[1])<<8 | int(d[2])
			out = append([]byte{'W', d[1], d[2], d[3]}, e.mem[a:a+int(d[3])]...)
		case n == 36 && d[0] == 'W':
			a := e.bank<<16 | int(d[1])<<8 | int(d[2])
			copy(e.mem[a:a+32], d[4:36])
			out = []byte{'A'}
		case string(d) == "ENDR" || string(d) == "ENDW":
			e.prog = false
			out = []byte{'A'}
		default:
			e.mu.Unlock()
			return fmt.Errorf("comando sconosciuto % X", d)
		}
		e.mu.Unlock()
		var rb [42]byte
		rb[0], rb[1], rb[2], rb[3] = 3, 0, byte(len(out)), 0
		copy(rb[4:], out)
		if err := p.Write(rb[:]); err != nil {
			return err
		}
	}
}

func emuReadInto(p Port, b []byte) error {
	r, err := emuRead(p, len(b))
	if err != nil {
		return err
	}
	copy(b, r)
	return nil
}

// rdtySynthImage: memoria vuota della famiglia (canali, zone e liste vuote, come
// dopo le cancellazioni di dmrconfig)
func rdtySynthImage(m *genModel) []byte {
	o := rdtyo(m)
	img := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
	copy(img, o.hdr)
	for b := 0; b < 8; b++ {
		bk := rdtyBank(b)
		for k := 0; k < 16; k++ {
			img[bk+k] = 0
		}
		for i := 0; i < 128; i++ {
			rdtyErase(o, img[bk+16+i*rdtyChSize:][:rdtyChSize])
		}
	}
	for k := 0; k < 32; k++ {
		img[rdtyZoneTab+k] = 0
	}
	for z := 0; z < o.nzones; z++ {
		zr := img[rdtyZoneOff(o, z):][:rdtyZoneLen(o)]
		for k := range zr {
			zr[k] = 0
		}
	}
	for s := 0; s < o.scanValid; s++ {
		img[o.scanOff+s] = 0
	}
	return img
}
