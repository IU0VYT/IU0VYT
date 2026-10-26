//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: ga510.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio simulata Baofeng serie H / Radioddity GA-510 (protocollo di ga510.py di CHIRP).

import (
	"bytes"
	"fmt"
	"sync"
)

type ga510Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
}

func (ga510Fam) emu(m *genModel, img []byte) genEmu {
	n := ga510o(m).memsize
	e := &ga510Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < n {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, n-len(e.img))...)
	}
	e.img = e.img[:n]
	return e
}

func (e *ga510Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *ga510Emu) serve(p Port) error {
	o := ga510o(e.m)
	prog := false
	var win []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if !prog {
			if c == 'E' {
				continue
			}
			win = append(win, c)
			if len(win) > len(o.magic) {
				win = win[len(win)-len(o.magic):]
			}
			if bytes.Equal(win, o.magic) {
				win = nil
				p.Write([]byte{0x06})
				if b, err := emuByte(p); err != nil || b != 'F' {
					return fmt.Errorf("comando F atteso (%v)", err)
				}
				p.Write([]byte("BFH\x00\x00\x00\x00\x00"))
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
