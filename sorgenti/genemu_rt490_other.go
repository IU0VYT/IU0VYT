//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radtel_rt490.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio simulata della famiglia Radtel RT-490 (protocollo di radtel_rt490.py di CHIRP).

import (
	"bytes"
	"fmt"
	"sync"
)

type rt490Emu struct {
	mu  sync.Mutex
	img []byte
}

func (rt490Fam) emu(m *genModel, img []byte) genEmu {
	e := &rt490Emu{img: append([]byte{}, img...)}
	if len(e.img) < rt490MemSize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, rt490MemSize-len(e.img))...)
	}
	e.img = e.img[:rt490MemSize]
	return e
}

func (e *rt490Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rt490Emu) serve(p Port) error {
	magic := []byte("PROGROMJJCCU")
	prog := false
	var win []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if !prog {
			win = append(win, c)
			if len(win) > len(magic) {
				win = win[1:]
			}
			if bytes.Equal(win, magic) {
				win = nil
				p.Write([]byte{0x06})
				if b, err := emuByte(p); err != nil || b != 'F' {
					return fmt.Errorf("comando F atteso (%v)", err)
				}
				p.Write([]byte{0x01, 0x36, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20})
				prog = true
			}
			continue
		}
		if c == 'E' {
			prog = false
			continue
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if addr < len(e.img) {
				copy(d, e.img[addr:])
			}
			e.mu.Unlock()
			p.Write(append([]byte{'R', h[0], h[1], h[2]}, d...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if addr+n <= len(e.img) {
				copy(e.img[addr:], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}
