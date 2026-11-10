//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: radio con il firmware nicFW simulate per le prove (comportamento dalla
// documentazione dell'autore e, per la TD-H3, dal sorgente pubblicato del
// firmware 2.5: uart.c).

import "sync"

// TD-H3 con nicFW 2.x
type nicfwEmu struct {
	mu      sync.Mutex
	img     []byte
	writes  []int // blocchi scritti, nell'ordine
	reboots int
	stopped bool // radio fermata (0x45) e non ancora riavviata (0x46)
}

func (nicfwFam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, nicfwSize)
	copy(mem, img)
	return &nicfwEmu{img: mem}
}

func (e *nicfwEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *nicfwEmu) serve(p Port) error {
	for {
		c, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		switch c[0] {
		case 0x01, 0x02:
			p.Write(c)
		case 0x45, 0x46:
			e.mu.Lock()
			e.stopped = c[0] == 0x45
			e.mu.Unlock()
			p.Write(c)
		case 0x30:
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			e.mu.Lock()
			d := append([]byte{0x30}, e.img[int(b[0])*32:int(b[0])*32+32]...)
			e.mu.Unlock()
			p.Write(append(d, nicfwSum(d[1:])))
		case 0x31:
			b, err := emuRead(p, 34)
			if err != nil {
				return nil
			}
			if nicfwSum(b[1:33]) != b[33] {
				continue // blocco non valido: nessuna conferma
			}
			e.mu.Lock()
			copy(e.img[int(b[0])*32:], b[1:33])
			e.writes = append(e.writes, int(b[0]))
			e.mu.Unlock()
			p.Write([]byte{0x31})
		case 0x49:
			e.mu.Lock()
			e.reboots++
			e.mu.Unlock()
		}
	}
}

// RT-880 / RT-950 Pro con nicFW880: la memoria simulata comincia a 0x50200
type nicfw880Emu struct {
	mu      sync.Mutex
	img     []byte
	pending map[int][]byte
	writes  []int // indirizzi scritti e confermati
	reboots int
}

func (nicfw880Fam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, nicfw880Size)
	copy(mem, img)
	return &nicfw880Emu{img: mem, pending: map[int][]byte{}}
}

func (e *nicfw880Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *nicfw880Emu) serve(p Port) error {
	for {
		c, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		if c[0] != 0xAA {
			continue
		}
		if c, err = emuRead(p, 1); err != nil {
			return nil
		}
		switch c[0] {
		case 0x30:
			a, err := emuRead(p, 4)
			if err != nil {
				return nil
			}
			off := int(a[0]) | int(a[1])<<8 | int(a[2])<<16 | int(a[3])<<24 - nicfw880Base
			d := make([]byte, 128)
			e.mu.Lock()
			if off >= 0 && off+128 <= len(e.img) {
				copy(d, e.img[off:])
			}
			e.mu.Unlock()
			out := append([]byte{0xAA, 0x30}, a...)
			out = append(out, d...)
			p.Write(append(out, nicfwSum(d)))
		case 0x40:
			b, err := emuRead(p, 4+128+1)
			if err != nil {
				return nil
			}
			off := int(b[0]) | int(b[1])<<8 | int(b[2])<<16 | int(b[3])<<24 - nicfw880Base
			if nicfwSum(b[4:132]) != b[132] || off < 0 || off+128 > len(e.img) {
				continue
			}
			e.mu.Lock()
			e.pending[off] = append([]byte{}, b[4:132]...)
			e.mu.Unlock()
			p.Write([]byte{0xAA, 0x40})
		case 0x48:
			e.mu.Lock()
			for off, d := range e.pending {
				copy(e.img[off:], d)
				e.writes = append(e.writes, off)
			}
			e.pending = map[int][]byte{}
			e.mu.Unlock()
			p.Write([]byte{0xAA, 0x48})
		case 0x50:
			e.mu.Lock()
			e.reboots++
			e.pending = map[int][]byte{}
			e.mu.Unlock()
		}
	}
}
