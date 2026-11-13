// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

//go:build !windows

package main

// Radio simulata della famiglia Retevis C6 / C66 e Radioddity Echo (drv_wlt.go),
// per le prove: risponde all'aggancio calcolando la risposta dai due numeri che
// trova nel pacchetto (non da quelli che il driver si aspetta), controlla la
// password, le somme e i limiti degli indirizzi.

import (
	"bytes"
	"fmt"
	"sync"
)

type wltEmu struct {
	mu    sync.Mutex
	o     *wltOpt
	mem   []byte // memoria fino alla password compresa
	pwd   []byte
	log   []string
	noise int // risposte sbagliate da dare ai primi agganci
}

func (wltFam) emu(m *genModel, img []byte) genEmu {
	o := wltO(m)
	e := &wltEmu{o: o, mem: bytes.Repeat([]byte{0xFF}, o.size()), pwd: bytes.Repeat([]byte{0xFF}, 6)}
	copy(e.mem, img)
	if _, _, ok := wltRange(m, e.mem); !ok { // radio nuova: 400-480 MHz, modello 0, area 0
		copy(e.mem[o.rng():], []byte{0x00, 0x40, 0x00, 0x48, 0x00, 0x00})
	}
	copy(e.pwd, e.mem[o.pwd():o.pwd()+6])
	return e
}

func (e *wltEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem...)
}

func (e *wltEmu) note(s string) {
	e.mu.Lock()
	e.log = append(e.log, s)
	e.mu.Unlock()
}

func (e *wltEmu) serve(p Port) error {
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 'S':
			b, err := emuRead(p, 10)
			if err != nil {
				return err
			}
			// r2 è l'ultimo byte; dice dove sta r1 fra i byte della chiave
			r2 := b[9]
			k := int(r2>>3) & 7
			r1 := b[k]
			key := append(append([]byte{}, b[:k]...), b[k+1:9]...)
			for i := range key {
				key[i] ^= byte(int(r1) << uint(i))
			}
			e.note("S")
			if !bytes.Equal(key, e.o.key) {
				continue // chiave di un altro modello: la radio tace
			}
			e.mu.Lock()
			bad := e.noise > 0
			if bad {
				e.noise--
			}
			e.mu.Unlock()
			if bad {
				p.Write([]byte{0x55})
				continue
			}
			_, want := wltHello(e.o.key, r1, r2)
			p.Write(want)
			pw, err := emuRead(p, 6)
			if err != nil {
				return err
			}
			if !bytes.Equal(pw, e.pwd) {
				e.note("password errata")
				continue
			}
			p.Write([]byte{'A'})
		case 'R':
			b, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			a, n := int(b[0])|int(b[1])<<8, int(b[2])
			if a+n > len(e.mem) {
				return fmt.Errorf("lettura fuori dalla memoria: %#x+%d", a, n)
			}
			e.note(fmt.Sprintf("R%04X", a))
			e.mu.Lock()
			d := append([]byte{}, e.mem[a:a+n]...)
			e.mu.Unlock()
			p.Write(append(d, wltSum(d)))
		case 'W':
			b, err := emuRead(p, 2+16+1)
			if err != nil {
				return err
			}
			a := int(b[0]) | int(b[1])<<8
			if a%16 != 0 || a+16 > e.o.pwd() {
				return fmt.Errorf("scrittura a un indirizzo non previsto: %#x", a)
			}
			if wltSum(b[2:18]) != b[18] {
				return fmt.Errorf("somma di controllo errata nella scrittura a %#x", a)
			}
			e.note(fmt.Sprintf("W%04X", a))
			e.mu.Lock()
			copy(e.mem[a:], b[2:18])
			e.mu.Unlock()
			p.Write([]byte{'A'})
		case 'E':
			e.note("E")
		default:
			return fmt.Errorf("comando inatteso %02X", c)
		}
	}
}
