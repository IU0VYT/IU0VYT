//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-13: memoria da 0x0000 a 0x03FF, saluto e blocchi di
// 8 byte come li usa il programma del costruttore (drv_rt13.go); tiene l'elenco
// dei comandi ricevuti.

import (
	"bytes"
	"fmt"
	"sync"
)

type rt13Emu struct {
	mu    sync.Mutex
	mem   []byte // memoria per indirizzo
	log   []int  // comando<<16 | indirizzo, nell'ordine ('H' saluto, 'E' con indirizzo 0)
	noise int    // byte estranei mandati prima dello 0x06 del saluto
	nak   int    // indirizzo del blocco rifiutato in scrittura (-1 nessuno)
	noF   bool   // la radio non risponde 'F' a 'E'
	hello []byte // saluto accettato
}

func (rt13Fam) emu(m *genModel, img []byte) genEmu {
	e := &rt13Emu{mem: bytes.Repeat([]byte{0xFF}, 0x0400), nak: -1, hello: rt13Hello}
	copy(e.mem, img)
	return e
}

func (e *rt13Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem[:rt13Size]...)
}

func (e *rt13Emu) note(c byte, addr int) {
	e.mu.Lock()
	e.log = append(e.log, int(c)<<16|addr)
	e.mu.Unlock()
}

func (e *rt13Emu) serve(p Port) error {
	emuRequireLine(p, false, true) // due bit di stop
	prog := false
	var win []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if !prog {
			win = append(win, c)
			if len(win) > len(e.hello) {
				win = win[len(win)-len(e.hello):]
			}
			if !bytes.Equal(win, e.hello) {
				continue
			}
			win = nil
			e.note('H', 0)
			p.Write(append(bytes.Repeat([]byte{0x00}, e.noise), 0x06))
			if b, err := emuByte(p); err != nil || b != 0x02 {
				return fmt.Errorf("0x02 atteso (%v)", err)
			}
			p.Write([]byte("P3107\xF7\x00\x00"))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("0x06 atteso (%v)", err)
			}
			p.Write([]byte{0x06})
			prog = true
			continue
		}
		if c == 'E' {
			e.note('E', 0)
			if !e.noF {
				p.Write([]byte{'F'})
			}
			prog = false
			continue
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		if n != 8 || addr%8 != 0 || addr+n > len(e.mem) {
			return fmt.Errorf("blocco inatteso %c %04X %02X", c, addr, n)
		}
		switch c {
		case 'R':
			e.note(c, addr)
			e.mu.Lock()
			d := append([]byte{'W', h[0], h[1], h[2]}, e.mem[addr:addr+n]...)
			e.mu.Unlock()
			p.Write(d)
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("0x06 atteso dopo la lettura (%v)", err)
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.note(c, addr)
			if addr == e.nak {
				p.Write([]byte{0x15})
				continue
			}
			e.mu.Lock()
			copy(e.mem[addr:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}
