//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: Yaesu VX-1 simulata. La copia è un blocco solo di 4818 byte: se entro
// un secondo e mezzo non arriva niente è la radio a inviare (come se fosse
// stato premuto [DWN]); se arrivano dei dati li riceve, controlla le due somme
// e conferma con 0x21.

import (
	"context"
	"fmt"
	"sync"
)

type vx1Emu struct {
	mu     sync.Mutex
	img    []byte
	writes int
	silent bool // radio non in modalità clone
	noEcho bool // cavo a due fili: i dati inviati non tornano indietro
	zeros  int  // 00 inviati prima della conferma (alcune versioni della radio)
	nak    bool // la radio non conferma
	cut    int  // in lettura la radio si ferma dopo questi byte (0 = tutta la memoria)
}

func (vx1Fam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, vx1Size)
	if img == nil {
		copy(mem[0xE6:], "YAESU")
	}
	copy(mem, img)
	vx1Fam{}.fixSums(mem) // la radio manda sempre le somme di controllo giuste
	return &vx1Emu{img: mem}
}

func (e *vx1Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *vx1Emu) serve(p Port) error {
	if e.silent {
		for {
			if _, err := emuRead(p, 1); err != nil {
				return nil
			}
		}
	}
	first := make([]byte, 1)
	if err := p.ReadFull(context.Background(), first, yaesuEmuWait); err != nil {
		if err != errTimeout {
			return nil
		}
		src := e.image()
		if e.cut > 0 {
			src = src[:e.cut]
		}
		p.Write(src)
		return nil
	}
	buf := append([]byte{}, first...)
	if !e.noEcho {
		p.Write(first)
	}
	for len(buf) < vx1Size {
		n := vx1Size - len(buf)
		if n > 64 {
			n = 64
		}
		b, err := emuRead(p, n)
		if err != nil {
			return fmt.Errorf("scrittura interrotta dopo %d byte", len(buf))
		}
		if !e.noEcho {
			p.Write(b)
		}
		buf = append(buf, b...)
	}
	if a, b := vx1Sums(buf); buf[0x00A0] != a || buf[0x12D1] != b {
		return fmt.Errorf("somme di controllo errate nella memoria ricevuta")
	}
	if e.nak {
		return nil
	}
	e.mu.Lock()
	copy(e.img, buf)
	e.writes++
	e.mu.Unlock()
	p.Write(append(make([]byte, e.zeros), 0x21))
	return nil
}
