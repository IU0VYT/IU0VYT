//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.9: Wouxun KG-UV8D / KG-UV8D Plus / KG-UV8E simulate (pacchetti con somma,
// cifrati con lo XOR a catena sulle Plus e 8E)

import (
	"bytes"
	"fmt"
	"sync"
)

type wx8Emu struct {
	mu  sync.Mutex
	img []byte
	o   *wx8Opt
}

func (wx8Fam) emu(m *genModel, img []byte) genEmu {
	e := &wx8Emu{img: append([]byte{}, img...), o: wx8o(m)}
	if len(e.img) < 0x8000 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x8000-len(e.img))...)
	}
	e.img = e.img[:0x8000]
	return e
}

func (e *wx8Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *wx8Emu) reply(p Port, cmd byte, payload []byte) error {
	h := []byte{e.o.start, cmd, 0xFF, byte(len(payload))}
	cs := sum8(append(append([]byte{}, h[1:]...), payload...))
	body := append(append([]byte{}, payload...), cs)
	if e.o.enc {
		body = wx8Enc(body)
	}
	return p.Write(append(h, body...))
}

func (e *wx8Emu) serve(p Port) error {
	for {
		// attesa dell'inizio del pacchetto (scarta i saluti di altri modelli)
		b, err := emuByte(p)
		if err != nil {
			return nil
		}
		if b != e.o.start {
			continue
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		body, err := emuRead(p, int(h[2])+1)
		if err != nil {
			return err
		}
		if e.o.enc {
			body = wx8Dec(body)
		}
		pl, cs := body[:len(body)-1], body[len(body)-1]
		if sum8(append(append([]byte{}, h...), pl...)) != cs {
			return fmt.Errorf("somma del pacchetto errata")
		}
		switch h[0] {
		case wx8CmdID:
			id := append([]byte(e.o.model), make([]byte, 43-len(e.o.model))...)
			e.reply(p, wx8CmdID, id)
		case wx8CmdRd:
			a, n := int(pl[0])<<8|int(pl[1]), int(pl[2])
			e.mu.Lock()
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			e.reply(p, wx8CmdRd, append([]byte{pl[0], pl[1]}, d...))
		case wx8CmdWr:
			a := int(pl[0])<<8 | int(pl[1])
			e.mu.Lock()
			copy(e.img[a:], pl[2:])
			e.mu.Unlock()
			e.reply(p, wx8CmdWr, []byte{pl[0], pl[1]})
		case wx8CmdEnd:
			return nil
		default:
			return fmt.Errorf("comando sconosciuto %d", h[0])
		}
	}
}

// 4.9: Wouxun KG-UV9D Plus e simili (kguv9dplus.py)
type wx9Emu struct {
	mu  sync.Mutex
	img []byte
}

func (wx9Fam) emu(m *genModel, img []byte) genEmu {
	e := &wx9Emu{img: append([]byte{}, img...)}
	if len(e.img) < 0x8000 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x8000-len(e.img))...)
	}
	e.img = e.img[:0x8000]
	return e
}

func (e *wx9Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func wx9Reply(op byte, payload []byte) []byte {
	d := wx9Pkt(op, payload)
	// verso il PC la direzione è 0x00 e la somma la comprende
	d[2] = 0x00
	cs := int(op) + len(payload)
	for _, b := range payload {
		cs += int(b)
	}
	plain := append(append([]byte{}, payload...), byte(cs&0xF))
	x := byte(0x52)
	for k := range plain {
		plain[k] ^= x
		x = plain[k]
	}
	return append(d[:4], plain...)
}

func (e *wx9Emu) serve(p Port) error {
	for {
		b, err := emuByte(p)
		if err != nil {
			return nil
		}
		if b != 0x7D {
			continue
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		rest, err := emuRead(p, int(h[2])+1)
		if err != nil {
			return err
		}
		op, pl, ok := wx9Unpkt(append([]byte{0x7D}, append(h, rest...)...))
		if !ok {
			return fmt.Errorf("somma del pacchetto errata")
		}
		switch op {
		case wx9CmdIdent:
			id := append([]byte("KG-UV9D02"), make([]byte, 74-9)...)
			p.Write(wx9Reply(op, id))
		case wx9CmdRconf:
			a, n := int(pl[0])<<8|int(pl[1]), int(pl[2])
			e.mu.Lock()
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(wx9Reply(op, append([]byte{pl[0], pl[1]}, d...)))
		case wx9CmdWconf:
			a := int(pl[0])<<8 | int(pl[1])
			e.mu.Lock()
			copy(e.img[a:], pl[2:])
			e.mu.Unlock()
			p.Write(wx9Reply(op, []byte{pl[0], pl[1]}))
		case wx9CmdHangup:
			return nil
		default:
			return fmt.Errorf("comando sconosciuto %02X", op)
		}
	}
}

// 4.9: famiglia KG-935G (stesso protocollo della KG-UV8D Plus)
func (kg935Fam) emu(m *genModel, img []byte) genEmu {
	e := &wx8Emu{img: append([]byte{}, img...), o: &kg935o(m).wx8Opt}
	if len(e.img) < 0x8000 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x8000-len(e.img))...)
	}
	e.img = e.img[:0x8000]
	return e
}
