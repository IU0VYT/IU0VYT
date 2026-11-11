//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-480 / RT-68: 16 KiB di memoria agli indirizzi del
// programma del costruttore; tiene l'elenco dei comandi ricevuti (drv_rt480.go).

import (
	"bytes"
	"fmt"
	"sync"
)

type rt480Emu struct {
	mu    sync.Mutex
	mem   []byte // memoria per indirizzo (0x0000-0x3FFF)
	log   []int  // comando<<16 | indirizzo, nell'ordine ('E' e 'F' con indirizzo 0)
	noise int    // byte estranei mandati prima dello 0x06 dell'aggancio
	deaf  int    // agganci ignorati (il programma rimanda il comando)
	nak   int    // indirizzo del blocco rifiutato in scrittura (-1 nessuno)
	hello []byte // comando di aggancio accettato
}

func (rt480Fam) emu(m *genModel, img []byte) genEmu {
	e := &rt480Emu{mem: bytes.Repeat([]byte{0xFF}, 0x4000), nak: -1, hello: []byte(rt480Hello)}
	o := 0
	for _, a := range rt480Areas {
		if o < len(img) {
			copy(e.mem[a[0]:a[1]], img[o:])
		}
		o += a[1] - a[0]
	}
	return e
}

func (e *rt480Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	var img []byte
	for _, a := range rt480Areas {
		img = append(img, e.mem[a[0]:a[1]]...)
	}
	return img
}

func (e *rt480Emu) note(c byte, addr int) {
	e.mu.Lock()
	e.log = append(e.log, int(c)<<16|addr)
	e.mu.Unlock()
}

func (e *rt480Emu) serve(p Port) error {
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
			if e.deaf > 0 {
				e.deaf--
				continue
			}
			p.Write(append(bytes.Repeat([]byte{0x00}, e.noise), 0x06))
			if b, err := emuByte(p); err != nil || b != 'F' {
				return fmt.Errorf("comando F atteso (%v)", err)
			}
			e.note('F', 0)
			p.Write([]byte("S19\x00\x01\x02\x03\x04"))
			prog = true
			continue
		}
		if c == 'E' {
			e.note('E', 0)
			prog = false
			continue
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		if n != 0x40 || addr%0x40 != 0 || addr+n > len(e.mem) {
			return fmt.Errorf("blocco inatteso %c %04X %02X", c, addr, n)
		}
		switch c {
		case 'R':
			e.note(c, addr)
			e.mu.Lock()
			// il programma controlla solo l'indirizzo: il primo byte qui è diverso apposta
			d := append([]byte{'W', h[0], h[1], h[2]}, e.mem[addr:addr+n]...)
			e.mu.Unlock()
			p.Write(d)
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
