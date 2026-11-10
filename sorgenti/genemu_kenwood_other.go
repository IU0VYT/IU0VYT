//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: radio Kenwood simulate per le prove.

import (
	"strings"
	"sync"
)

// TH-D74 / TH-D75: comandi a righe («ID», «0M PROGRAM»), poi blocchi da 256 byte
type thd74Emu struct {
	mu     sync.Mutex
	o      *thd74Opt
	id     string
	img    []byte
	writes int
	silent bool // radio spenta: nessuna risposta
}

func (thd74Fam) emu(m *genModel, img []byte) genEmu { return newThd74Emu(m, img) }
func (thd72Fam) emu(m *genModel, img []byte) genEmu { return newThd74Emu(m, img) }

func newThd74Emu(m *genModel, img []byte) *thd74Emu {
	o := m.Opt.(*thd74Opt)
	mem := make([]byte, o.bytes())
	copy(mem, img)
	return &thd74Emu{o: o, id: o.id, img: mem}
}

func (e *thd74Emu) block(h []byte) int {
	if e.o.d72 {
		return int(h[1]) | int(h[2])<<8
	}
	return int(h[0])<<8 | int(h[1])
}

func (e *thd74Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *thd74Emu) serve(p Port) error {
	for {
		// comandi a righe
		var line []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if b[0] == '\r' {
				break
			}
			line = append(line, b[0])
		}
		if e.silent {
			continue
		}
		switch cmd := strings.TrimSpace(string(line)); cmd {
		case "ID":
			p.Write([]byte("ID " + e.id + "\r"))
			continue
		case "0M PROGRAM":
			p.Write([]byte("0M\r"))
		default:
			p.Write([]byte("?\r"))
			continue
		}
		// programmazione
		for prog := true; prog; {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			switch c[0] {
			case 'R':
				h, err := emuRead(p, 4)
				if err != nil {
					return nil
				}
				blk := e.block(h)
				if (blk+1)*256 > len(e.img) {
					continue
				}
				e.mu.Lock()
				d := append(e.o.head('W', blk), e.img[blk*256:blk*256+256]...)
				e.mu.Unlock()
				p.Write(d)
				if a, err := emuRead(p, 1); err != nil || a[0] != 0x06 {
					return nil
				}
				p.Write([]byte{0x06})
			case 'W':
				h, err := emuRead(p, 4+256)
				if err != nil {
					return nil
				}
				blk := e.block(h)
				if (blk+1)*256 > len(e.img) {
					p.Write([]byte{0x15})
					continue
				}
				e.mu.Lock()
				copy(e.img[blk*256:], h[4:])
				e.writes++
				e.mu.Unlock()
				p.Write([]byte{0x06})
			case 'E':
				prog = false
			}
		}
	}
}
