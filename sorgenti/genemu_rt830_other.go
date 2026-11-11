//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-830 / RT-850: 4 KiB di memoria agli indirizzi del
// programma del costruttore; tiene l'elenco dei comandi ricevuti (drv_rt830.go).

import (
	"bytes"
	"fmt"
	"sync"
)

type rt830Emu struct {
	mu  sync.Mutex
	o   *rt830Opt
	mem []byte // memoria per indirizzo (0x0000-0x0FFF)
	log []int  // comando<<16 | indirizzo, nell'ordine (ingresso 'I', uscita 'E')
}

func (rt830Fam) emu(m *genModel, img []byte) genEmu {
	o := rt830o(m)
	e := &rt830Emu{o: o, mem: bytes.Repeat([]byte{0xA5}, 0x1000)}
	copy(e.mem[o.base:o.base+o.size()], bytes.Repeat([]byte{0xFF}, o.size()))
	copy(e.mem[o.base:o.base+o.size()], img)
	e.mem[0] = 0x58
	return e
}

func (e *rt830Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem[e.o.base:e.o.base+e.o.size()]...)
}

func (e *rt830Emu) note(c byte, addr int) {
	e.mu.Lock()
	e.log = append(e.log, int(c)<<16|addr)
	e.mu.Unlock()
}

func (e *rt830Emu) serve(p Port) error {
	in := false
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == rt830Enter[0] {
			rest, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			switch {
			case bytes.Equal(rest, rt830Enter[1:]):
				in = true
				e.note('I', 0)
				p.Write([]byte{0x06})
			case bytes.Equal(rest, rt830Exit[1:]):
				e.note('E', 0)
				in = false
			}
			continue
		}
		if !in {
			continue // presentazioni di altri modelli
		}
		h, err := emuRead(p, 2)
		if err != nil {
			return err
		}
		addr := int(h[0])<<8 | int(h[1])
		if addr%16 != 0 || addr+16 > len(e.mem) {
			return fmt.Errorf("indirizzo inatteso %04X", addr)
		}
		switch c {
		case 'R':
			s, err := emuByte(p)
			if err != nil {
				return err
			}
			if s != sum8([]byte{c, h[0], h[1]}) {
				return fmt.Errorf("somma della richiesta errata")
			}
			e.note(c, addr)
			e.mu.Lock()
			out := append([]byte{'R', h[0], h[1]}, e.mem[addr:addr+16]...)
			e.mu.Unlock()
			p.Write(append(out, sum8(out)))
		case 'W':
			d, err := emuRead(p, 17)
			if err != nil {
				return err
			}
			if sum8(append([]byte{c, h[0], h[1]}, d[:16]...)) != d[16] {
				return fmt.Errorf("somma del blocco errata")
			}
			e.note(c, addr)
			e.mu.Lock()
			copy(e.mem[addr:], d[:16])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}
