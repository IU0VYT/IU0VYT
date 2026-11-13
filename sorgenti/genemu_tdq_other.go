// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !windows

package main

// Radio simulata della famiglia «tdq» (drv_tdq.go), per le prove.

import (
	"bytes"
	"fmt"
	"sync"
)

type tdqEmu struct {
	mu  sync.Mutex
	o   *tdqOpt
	mem []byte
	log []string
	// per le prove: risposta alla chiave (07 nella radio vera) e all'ultimo 08 dell'aggancio
	ack07, ack08 byte
}

func (tdqFam) emu(m *genModel, img []byte) genEmu {
	e := &tdqEmu{o: tdqO(m), mem: bytes.Repeat([]byte{0xFF}, tdqSize), ack07: 0x07, ack08: 0x08}
	copy(e.mem, img)
	if tdqBlankSet(e.mem) { // radio nuova: voce in inglese, guadagno 5, squelch 4, TOT 30 s, suono dei tasti, risparmio
		copy(e.mem[tdqSetAddr:], []byte{0x52, 4, 2, 5, 0, 0x50, 2, 0, 1, 1, 1, 9, 0, 12, 0, 1})
		copy(e.mem[0xCC0+7:], []byte{1, 0, 8}) // versione dell'hardware
	}
	return e
}

func (e *tdqEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem...)
}

func (e *tdqEmu) note(s string) {
	e.mu.Lock()
	e.log = append(e.log, s)
	e.mu.Unlock()
}

func (e *tdqEmu) serve(p Port) error {
	// stato del collegamento (di questo collegamento, non della radio): 0 chiuso,
	// 1 dopo la chiave, 2 dopo «123456», 3 in programmazione
	state := 0
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch {
		case state == 0 && c == e.o.magic[0]:
			if b, err := emuRead(p, 6); err != nil || string(b) != e.o.magic[1:] {
				return fmt.Errorf("chiave errata: % X", b)
			}
			e.note("chiave")
			state = 1
			p.Write([]byte{e.ack07})
		case state == 1 && c == tdqPass[0]:
			if b, err := emuRead(p, len(tdqPass)-1); err != nil || !bytes.Equal(b, tdqPass[1:]) {
				return fmt.Errorf("seconda chiave errata: % X", b)
			}
			state = 2
			p.Write([]byte{0x55})
		case state == 2 && c == 0x08:
			state = 3
			p.Write([]byte{e.ack08})
		case state == 3 && c == 0x08:
			p.Write([]byte{0x08})
		case state == 3 && c == 0x52:
			b, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr := int(b[0])<<8 | int(b[1])
			if b[2] != 0x40 || addr%0x40 != 0 || addr+0x40 > tdqSize {
				return fmt.Errorf("lettura non prevista: %#04x, %d byte", addr, b[2])
			}
			e.note(fmt.Sprintf("R%04X", addr))
			e.mu.Lock()
			out := append([]byte{0x57, b[0], b[1], b[2]}, e.mem[addr:addr+0x40]...)
			e.mu.Unlock()
			p.Write(out)
		case state == 3 && c == 0x57:
			b, err := emuRead(p, 3+16)
			if err != nil {
				return err
			}
			addr := int(b[0])<<8 | int(b[1])
			if b[2] != 0x10 || addr%0x10 != 0 || addr+0x10 > tdqSize {
				return fmt.Errorf("scrittura non prevista: %#04x, %d byte", addr, b[2])
			}
			e.note(fmt.Sprintf("W%04X", addr))
			e.mu.Lock()
			if addr == 0xCC0 { // la versione dell'hardware (byte 7-9) non si può cambiare
				copy(e.mem[addr:], b[3:3+7])
				copy(e.mem[addr+10:], b[3+10:])
			} else {
				copy(e.mem[addr:], b[3:])
			}
			e.mu.Unlock()
			p.Write([]byte{0x08})
		case state == 3 && c == 0x45:
			e.note("fine")
			state = 0
		default:
			return fmt.Errorf("comando inatteso %02X nello stato %d", c, state)
		}
	}
}
