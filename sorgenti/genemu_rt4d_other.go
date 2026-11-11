//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-4D / RT-6D / RT-1000 Pro: memoria a blocchi da 1 KB
// con i numeri di blocco del programma del costruttore; accetta in scrittura
// solo i comandi che il programma usa e ne tiene l'elenco (drv_rt4d.go).

import (
	"bytes"
	"fmt"
	"sync"
)

type rt4dEmu struct {
	mu     sync.Mutex
	blocks map[int][]byte // blocchi in lettura
	reads  []int          // blocchi letti, nell'ordine
	writes []int          // comando<<16 | blocco, nell'ordine
}

// comando di scrittura → primo blocco letto e numero di blocchi (RT-4D)
var rt4dEmuAreas = map[byte][2]int{
	0x90: {rt4dBlkCfg, 1}, 0x91: {rt4dBlkCh, rt4dChBlocks}, 0x92: {rt4dBlkVfo, 1}, 0x93: {rt4dBlkZone, rt4dZnBlocks},
	0x94: {0x178, 208}, 0x95: {0x318, 20}, 0x96: {0x340, 12}, 0x97: {0x358, 4}, 0x98: {0x3C0, 4},
}

func (rt4dFam) emu(m *genModel, img []byte) genEmu {
	e := &rt4dEmu{blocks: map[int][]byte{}}
	img = append(append([]byte{}, img...), bytes.Repeat([]byte{0xFF}, rt4dSize)...)
	for k := 0; k < rt4dChBlocks; k++ {
		e.blocks[rt4dBlkCh+k] = append([]byte{}, img[k*dm5kBlock:(k+1)*dm5kBlock]...)
	}
	for k := 0; k < rt4dZnBlocks; k++ {
		e.blocks[rt4dBlkZone+k] = append([]byte{}, img[rt4dZoneOff+k*dm5kBlock:rt4dZoneOff+(k+1)*dm5kBlock]...)
	}
	return e
}

func (e *rt4dEmu) block(n int) []byte {
	if d, ok := e.blocks[n]; ok {
		return d
	}
	return bytes.Repeat([]byte{0xFF}, dm5kBlock)
}

func (e *rt4dEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	var img []byte
	for k := 0; k < rt4dChBlocks; k++ {
		img = append(img, e.block(rt4dBlkCh+k)...)
	}
	for k := 0; k < rt4dZnBlocks; k++ {
		img = append(img, e.block(rt4dBlkZone+k)...)
	}
	return img
}

func (e *rt4dEmu) serve(p Port) error {
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
			d := e.block(n)
			e.reads = append(e.reads, n)
			e.mu.Unlock()
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
			a, ok := rt4dEmuAreas[c]
			if !ok || n >= a[1] {
				return fmt.Errorf("blocco %02X %d inatteso", c, n)
			}
			e.mu.Lock()
			e.blocks[a[0]+n] = append([]byte{}, r[2:2+dm5kBlock]...)
			e.writes = append(e.writes, int(c)<<16|n)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}
