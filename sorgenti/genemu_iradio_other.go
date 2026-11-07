//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.9: radio simulate con il protocollo iRadio (AR-518, UV58Plus, RT-880G)

import (
	"bytes"
	"fmt"
	"sync"
)

type iradioEmu struct {
	mu  sync.Mutex
	img []byte
	f   iradioFam
}

func (f iradioFam) emu(m *genModel, img []byte) genEmu {
	p := f.proto()
	e := &iradioEmu{img: append([]byte{}, img...), f: f}
	if len(e.img) < p.size {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, p.size-len(e.img))...)
	}
	e.img = e.img[:p.size]
	return e
}

func (e *iradioEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *iradioEmu) serve(p Port) error {
	pr := e.f.proto()
	in := false
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == pr.enter[0] && !in {
			rest, err := emuRead(p, len(pr.enter)-1)
			if err != nil {
				return err
			}
			if !bytes.Equal(append([]byte{c}, rest...), pr.enter) {
				continue
			}
			in = true
			p.Write([]byte{0x06})
			continue
		}
		if !in {
			continue // presentazioni di altri modelli
		}
		if c == pr.exit[0] {
			if _, err := emuRead(p, len(pr.exit)-1); err != nil {
				return err
			}
			return nil
		}
		h, err := emuRead(p, 2)
		if err != nil {
			return err
		}
		addr := int(h[0])<<8 | int(h[1])
		switch {
		case c == 'R':
			if _, err := emuRead(p, 1); err != nil {
				return err
			}
			a := addr - pr.readOff
			if pr.byBlock {
				a *= pr.block
			}
			if e.f.kind == irRT880G && addr == 0 {
				a = 0 // lettura iniziale della scrittura (blocco 0)
			}
			e.mu.Lock()
			if a < 0 || a+pr.block > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("blocco 0x%04X fuori dalla memoria", addr)
			}
			d := append([]byte{}, e.img[a:a+pr.block]...)
			e.mu.Unlock()
			out := append([]byte{'R', h[0], h[1]}, d...)
			p.Write(append(out, sum8(out)))
		default:
			d, err := emuRead(p, pr.block+1)
			if err != nil {
				return err
			}
			if sum8(append(append([]byte{c}, h...), d[:pr.block]...)) != d[pr.block] {
				return fmt.Errorf("somma del blocco errata")
			}
			var a int
			switch {
			case c == 'W':
				a = addr - pr.readOff
			case c == 'I':
				a = addr * pr.block
			default:
				base := -1
				for _, z := range rt880gRegions {
					if byte(z[2]) == c {
						base = z[0]
					}
				}
				if base < 0 {
					return fmt.Errorf("comando sconosciuto %02X", c)
				}
				a = base + addr*pr.block
			}
			e.mu.Lock()
			copy(e.img[a:], d[:pr.block])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		}
	}
}
