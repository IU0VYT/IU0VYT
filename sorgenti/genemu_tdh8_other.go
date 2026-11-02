//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// TIDRADIO TD-H8 simulata (protocollo di tdh8.py): saluto, identificativo,
// lettura e scrittura a blocchi di 32 byte con somma di controllo, uscita 'E'.
// Un saluto diverso (es. quello della terza generazione) non riceve risposta.

import (
	"bytes"
	"fmt"
	"sync"
)

type tdh8Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte // immagine come quella di CHIRP: identificativo (8 byte) + dati
}

func (tdh8Fam) emu(m *genModel, img []byte) genEmu {
	o := tdh8o(m)
	n := 8 + (o.memsize+0x1F)/0x20*0x20
	e := &tdh8Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < n {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, n-len(e.img))...)
	}
	e.img = e.img[:n]
	copy(e.img, o.ident)
	return e
}

func (e *tdh8Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *tdh8Emu) serve(p Port) error {
	o := tdh8o(e.m)
	for {
		mg, err := emuRead(p, len(o.magic))
		if err != nil {
			return nil
		}
		if bytes.Equal(mg, o.magic) {
			break
		}
		// saluto di un altro modello: nessuna risposta (la radio ignora i byte)
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write(o.ident)
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
			return nil
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
			if 8+addr < len(e.img) {
				copy(d, e.img[8+addr:])
			}
			e.mu.Unlock()
			var sum byte
			for _, b := range d {
				sum += b
			}
			p.Write(append(append([]byte{'W', h[0], h[1], h[2]}, d...), sum))
		case 'W':
			d, err := emuRead(p, n+1)
			if err != nil {
				return err
			}
			var sum byte
			for _, b := range d[:n] {
				sum += b
			}
			if sum != d[n] {
				return fmt.Errorf("somma di controllo errata nel blocco 0x%04X", addr)
			}
			e.mu.Lock()
			if 8+addr+n <= len(e.img) {
				copy(e.img[8+addr:], d[:n])
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}
