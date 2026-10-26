//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulate della serie D («BF Series CPS»), delle BF-T17 / BF-T18 e della
// BF-T9, con i
// protocolli descritti in drv_bfd.go.

import (
	"bytes"
	"fmt"
	"sync"
)

type bfdEmu struct {
	mu   sync.Mutex
	m    *genModel
	img  []byte
	size int
}

func (bfdFam) emu(m *genModel, img []byte) genEmu {
	e := &bfdEmu{m: m, img: append([]byte{}, img...), size: bfdMemsize}
	e.fit()
	return e
}

func (t17Fam) emu(m *genModel, img []byte) genEmu {
	e := &bfdEmu{m: m, img: append([]byte{}, img...), size: t17Memsize}
	e.fit()
	return e
}

func (e *bfdEmu) fit() {
	if len(e.img) < e.size {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, e.size-len(e.img))...)
	}
	e.img = e.img[:e.size]
}

func (e *bfdEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *bfdEmu) peek(addr, n int) []byte {
	d := bytes.Repeat([]byte{0xFF}, n)
	e.mu.Lock()
	if addr < len(e.img) {
		copy(d, e.img[addr:])
	}
	e.mu.Unlock()
	return d
}

func (e *bfdEmu) poke(addr int, d []byte) {
	e.mu.Lock()
	if addr+len(d) <= len(e.img) {
		copy(e.img[addr:], d)
	}
	e.mu.Unlock()
}

func (e *bfdEmu) serve(p Port) error {
	if _, ok := e.m.Opt.(*t17Opt); ok {
		return e.serveT17(p)
	}
	o := bfdo(e.m)
	ident, query, unlocked := false, false, !o.unlock
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 0x02:
			if ident {
				ident = false
				p.Write(bfdIdent)
				query = true
				continue
			}
			cmd, err := emuRead(p, 7)
			if err != nil {
				return err
			}
			switch {
			case string(cmd) == "ucbfpwd":
				p.Write([]byte{0x06})
				if k, err := emuRead(p, 4); err != nil || !bytes.Equal(k, []byte{0x10, 0xC5, 0xEA, 0x35}) {
					return fmt.Errorf("sblocco inatteso")
				}
				p.Write([]byte{0x06})
				if k, err := emuRead(p, 16); err != nil || !bytes.Equal(k, []byte("Q041OTUwMA==\x00\x00\x00\x00")) {
					return fmt.Errorf("password inattesa")
				}
				p.Write([]byte{0x06})
				unlocked = true
			case string(cmd) == "prOGRAM":
				if unlocked {
					p.Write([]byte{0x00, o.key, 0x01, 0x10})
				}
			case string(cmd[2:]) == "OGRAM" && (cmd[0] == o.magic[0] && cmd[1] == o.magic[1] || cmd[0] == 0x80 && cmd[1] == 0x82):
				p.Write([]byte{0x06})
				ident = true
			default:
				return fmt.Errorf("comando inatteso %q", cmd)
			}
		case 0x06:
			p.Write([]byte{0x06})
		case 'E':
			p.Write([]byte{'F'})
		case 'R':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr := int(h[0])<<8 | int(h[1])
			d := e.peek(addr, int(h[2]))
			if query {
				// prima lettura dopo l'identificativo: richiesta fissa (0x0130), risposta 8×FF
				query = false
				if addr != 0x130 {
					return fmt.Errorf("richiesta 0x0130 attesa, ricevuto 0x%04X", addr)
				}
				d = bytes.Repeat([]byte{0xFF}, 8)
			}
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
		case 'W':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			d, err := emuRead(p, int(h[2]))
			if err != nil {
				return err
			}
			e.poke(int(h[0])<<8|int(h[1]), d)
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

func (e *bfdEmu) serveT17(p Port) error {
	o := t17o(e.m)
	ident := false
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 0x02:
			if ident {
				ident = false
				if e.m.ID == "bf-t9" {
					p.Write([]byte("P3107\xF4AM")) // BF-T9
				} else {
					p.Write([]byte("P3013\xF7\x00\x00"))
				}
				continue
			}
			cmd, err := emuRead(p, 7)
			if err != nil {
				return err
			}
			if string(cmd) != o.cmd {
				return fmt.Errorf("comando inatteso %q", cmd)
			}
			p.Write([]byte{0x06})
			ident = true
		case 0x06:
			p.Write([]byte{0x06})
		case 'E':
			p.Write([]byte{'F'})
		case 0xFE:
			// codice di salto di frequenza (799 = 0x031F, valore predefinito del programma)
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			p.Write([]byte{0xFE, h[0], h[1], h[2], 0x1F, 0x03, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		case 'R':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, e.peek(int(h[0])<<8|int(h[1]), int(h[2]))...))
		case 'W':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			d, err := emuRead(p, int(h[2]))
			if err != nil {
				return err
			}
			e.poke(int(h[0])<<8|int(h[1]), d)
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}
