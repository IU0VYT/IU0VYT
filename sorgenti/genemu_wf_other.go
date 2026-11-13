// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

//go:build !windows

package main

// Radio simulata della famiglia Radioddity UB15-G / DB-35G / DB35-UV / DB40-G
// (drv_wf.go), per le prove: tipo di radio, aggancio in tre passi, letture e
// scritture per numero di record. Davanti ai dati letti ripete il comando (52,
// indirizzo, lunghezza), come fa la DB40-G vera secondo il driver del CHIRP; il
// programma del costruttore salta quei 4 byte senza guardarli. La domanda del
// tipo prima dell'aggancio è facoltativa (il CHIRP non la manda).

import (
	"bytes"
	"fmt"
	"sync"
)

type wfEmu struct {
	mu   sync.Mutex
	m    *genModel
	o    *wfOpt
	mem  []byte
	kind []byte
	log  []string
	// badEcho: davanti ai dati letti mette 57 invece di 52 (per le prove)
	badEcho bool
}

func (wfFam) emu(m *genModel, img []byte) genEmu {
	o := wfO(m)
	e := &wfEmu{m: m, o: o, mem: make([]byte, o.size(m)), kind: o.ver}
	copy(e.mem, img)
	if wfBlank(e.mem[o.param(m):]) { // radio nuova: squelch 3, TOT 180 s, suono dei tasti
		if o.page == 32 { // DB40-G: un byte per voce (blocco, suono dei tasti, TOT, squelch)
			copy(e.mem[o.param(m):], []byte{0, 1, 6, 3})
		} else {
			e.mem[o.param(m)], e.mem[o.param(m)+1] = 3, 0x0B
		}
	}
	return e
}

func (e *wfEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem...)
}

func (e *wfEmu) note(s string) {
	e.mu.Lock()
	e.log = append(e.log, s)
	e.mu.Unlock()
}

// row: posizione nell'immagine della riga con quell'indirizzo, -1 se non esiste
func (e *wfEmu) row(addr int) int {
	if addr == wfParamAddr {
		return e.o.param(e.m)
	}
	if r := addr - e.o.start; r >= 0 && r%16 == 0 && r/16 < e.m.Slots {
		return r / 16 * e.o.page
	}
	return -1
}

func (e *wfEmu) serve(p Port) error {
	// stato del collegamento: 0 chiuso, 1 dopo PROGRAM, 2 dopo la sigla, 3 in programmazione.
	// È di questo collegamento e non della radio simulata: il 62 finale del collegamento
	// precedente può arrivare quando quello nuovo è già cominciato.
	state := 0
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch {
		case c == 0x55 && state == 0:
			if b, err := emuRead(p, 3); err != nil || !bytes.Equal(b, []byte{1, 0, 0}) {
				return fmt.Errorf("domanda del tipo errata: % X", b)
			}
			e.note("tipo")
			p.Write(e.kind)
		case c == 0x02 && state == 0:
			if b, err := emuRead(p, 7); err != nil || string(b) != "PROGRAM" {
				return fmt.Errorf("aggancio errato: % X", b)
			}
			state = 1
			p.Write([]byte{0x06})
		case c == 0x02 && state == 1:
			state = 2
			p.Write([]byte("SMP558\xFF\xFF"))
		case c == 0x06 && state == 2:
			state = 3
			p.Write([]byte{0x06})
		case c == 0x52 && state == 3:
			b, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr, n := int(b[0])<<8|int(b[1]), int(b[2])
			at := e.row(addr)
			if at < 0 || n != e.o.page {
				return fmt.Errorf("lettura non prevista: %#04x, %d byte", addr, n)
			}
			e.note(fmt.Sprintf("R%04X", addr))
			e.mu.Lock()
			out := append([]byte{0x52, b[0], b[1], b[2]}, e.mem[at:at+n]...)
			if e.badEcho {
				out[0] = 0x57
			}
			e.mu.Unlock()
			p.Write(out)
		case c == 0x57 && state == 3:
			b, err := emuRead(p, 3+e.o.page)
			if err != nil {
				return err
			}
			addr, n := int(b[0])<<8|int(b[1]), int(b[2])
			at := e.row(addr)
			if at < 0 || n != e.o.page {
				return fmt.Errorf("scrittura non prevista: %#04x, %d byte", addr, n)
			}
			e.note(fmt.Sprintf("W%04X", addr))
			e.mu.Lock()
			copy(e.mem[at:], b[3:])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		case c == 0x62:
			e.note("fine")
			state = 0
		default:
			return fmt.Errorf("comando inatteso %02X nello stato %d", c, state)
		}
	}
}
