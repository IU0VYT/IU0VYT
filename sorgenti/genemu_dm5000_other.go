//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Bajeton DM-5000: memoria a blocchi da 1 KB con i numeri di
// blocco del programma del costruttore (lettura e scrittura con numerazioni
// diverse), protocollo descritto in drv_dm5000.go.

import (
	"bytes"
	"fmt"
	"sync"
)

type dm5kEmu struct {
	mu     sync.Mutex
	blocks map[int][]byte // blocchi in lettura
}

func (dm5kFam) emu(m *genModel, img []byte) genEmu {
	e := &dm5kEmu{blocks: map[int][]byte{}}
	img = append(append([]byte{}, img...), bytes.Repeat([]byte{0xFF}, dm5kSize())...)
	off := 0
	for _, a := range dm5kAreas {
		for k := 0; k < a.count; k++ {
			e.blocks[a.read+k] = append([]byte{}, img[off:off+dm5kBlock]...)
			off += dm5kBlock
		}
	}
	// blocco 0: dati della radio (non letti dal programma come area)
	b0 := bytes.Repeat([]byte{0xFF}, dm5kBlock)
	copy(b0, "DM-5000")
	e.blocks[0] = b0
	return e
}

func (e *dm5kEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	var img []byte
	for _, a := range dm5kAreas {
		for k := 0; k < a.count; k++ {
			img = append(img, e.blocks[a.read+k]...)
		}
	}
	return img
}

func (e *dm5kEmu) serve(p Port) error {
	connected := false
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch {
		case c == 0x34:
			r, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			switch {
			case bytes.Equal(r, dm5kConnect[1:]):
				connected = true
				p.Write([]byte{0x06})
			case bytes.Equal(r, dm5kEnd[1:]):
				connected = false
			default:
				return fmt.Errorf("comando inatteso % X", r)
			}
		case c == 'R' && connected:
			r, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			if dm5kSum([]byte{c, r[0], r[1]}) != r[2] {
				return fmt.Errorf("somma di controllo errata nella lettura")
			}
			n := int(r[0])<<8 | int(r[1])
			e.mu.Lock()
			d, ok := e.blocks[n]
			e.mu.Unlock()
			if !ok {
				d = bytes.Repeat([]byte{0xFF}, dm5kBlock)
			}
			f := append([]byte{'R', r[0], r[1]}, d...)
			p.Write(append(f, dm5kSum(f)))
		case c >= 0x90 && c <= 0x9A && connected:
			r, err := emuRead(p, 3+dm5kBlock)
			if err != nil {
				return err
			}
			f := append([]byte{c}, r[:2+dm5kBlock]...)
			if dm5kSum(f) != r[2+dm5kBlock] {
				return fmt.Errorf("somma di controllo errata nella scrittura")
			}
			n := int(r[0])<<8 | int(r[1])
			ok := false
			for _, a := range dm5kAreas {
				if a.cmd == c && n < a.count {
					e.mu.Lock()
					e.blocks[a.read+n] = append([]byte{}, r[2:2+dm5kBlock]...)
					e.mu.Unlock()
					ok = true
				}
			}
			if !ok {
				return fmt.Errorf("blocco %02X %d inatteso", c, n)
			}
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}
