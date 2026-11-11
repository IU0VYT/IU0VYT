//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180: ingresso a 9600
// baud con due bit di stop, passaggio a 19200 baud, blocchi di 256 byte per
// numero e pezzi di 64 byte per indirizzo (drv_kwtk180.go).

import (
	"bytes"
	"fmt"
	"sync"
)

type kwx180Emu struct {
	mu  sync.Mutex
	img []byte
}

func (kwx180Fam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, kwx180Size)
	copy(mem, img)
	return &kwx180Emu{img: mem[:kwx180Size]}
}

func (e *kwx180Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

// velocità scelta dal PC (0 se la porta simulata non la riporta)
func kwx180PeerBaud(p Port) int {
	for {
		if u, ok := p.(portUnwrapper); ok {
			p = u.unwrapPort()
			continue
		}
		break
	}
	pe, ok := p.(*pipeEnd)
	if !ok {
		return 0
	}
	pe.other.mu.Lock()
	defer pe.other.mu.Unlock()
	return pe.other.baud
}

func (e *kwx180Emu) serve(p Port) error {
	emuRequireLine(p, false, true)
	for {
		if err := emuWaitFor(p, "PROGRAM"); err != nil {
			return nil
		}
		if b := kwx180PeerBaud(p); b != 0 && b != 9600 {
			return fmt.Errorf("«PROGRAM» a %d baud", b)
		}
		p.Write([]byte{0x16})
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		if b := kwx180PeerBaud(p); b != 0 && b != 19200 {
			return fmt.Errorf("dopo l'ingresso il PC è a %d baud", b)
		}
		e.mu.Lock()
		id := append([]byte{}, e.img[kwx180ID:kwx180ID+8]...)
		e.mu.Unlock()
		p.Write(id)
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
	session:
		for {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			switch c[0] {
			case 'E':
				p.Write([]byte{0x06})
				break session
			case 'R', 'W', 'Z':
				h, err := emuRead(p, 2)
				if err != nil {
					return nil
				}
				blk := int(h[1])
				if h[0] != 0 || blk >= kwx180Blocks {
					return fmt.Errorf("blocco non valido % X", h)
				}
				if c[0] == 'R' {
					e.mu.Lock()
					d := append([]byte{}, e.img[blk*256:blk*256+256]...)
					e.mu.Unlock()
					if bytes.Equal(d, bytes.Repeat([]byte{0xFF}, 256)) {
						p.Write([]byte{'Z', 0xFF})
					} else {
						p.Write(append(append([]byte{'W'}, d...), kwtkSum(d)))
					}
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
					p.Write([]byte{0x06})
					continue
				}
				var d []byte
				if c[0] == 'W' {
					b, err := emuRead(p, 257)
					if err != nil {
						return nil
					}
					if kwtkSum(b[:256]) != b[256] {
						p.Write([]byte{0x15})
						continue
					}
					d = b[:256]
				} else {
					b, err := emuRead(p, 1)
					if err != nil {
						return nil
					}
					d = bytes.Repeat(b, 256)
				}
				e.mu.Lock()
				copy(e.img[blk*256:], d)
				e.mu.Unlock()
				p.Write([]byte{0x06})
			case 'S', 'X':
				h, err := emuRead(p, 3)
				if err != nil {
					return nil
				}
				a := int(h[0])<<8 | int(h[1])
				if h[2] != 0x40 || a < 0x0100 || a >= 0x1200 || a%0x40 != 0 {
					return fmt.Errorf("indirizzo non valido % X", h)
				}
				off := a + 0xBF00
				if c[0] == 'S' {
					e.mu.Lock()
					p.Write(append([]byte{'X'}, e.img[off:off+0x40]...))
					e.mu.Unlock()
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
					p.Write([]byte{0x06})
					continue
				}
				b, err := emuRead(p, 0x40)
				if err != nil {
					return nil
				}
				e.mu.Lock()
				copy(e.img[off:], b)
				e.mu.Unlock()
				p.Write([]byte{0x06})
			default:
				return fmt.Errorf("comando sconosciuto %02X", c[0])
			}
		}
	}
}
