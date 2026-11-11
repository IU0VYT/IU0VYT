//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-69 GPS: 32 KiB di memoria, pacchetti in chiaro della
// KG-UV8D con le risposte «7D comando 00 …» che il programma del costruttore
// pretende (drv_rt69.go); tiene l'elenco dei comandi ricevuti. Sotto 0x0480 la
// radio ha i suoi dati (limiti di frequenza), che non fanno parte dell'immagine.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type rt69Emu struct {
	mu   sync.Mutex
	mem  []byte
	log  []int // comando<<16 | indirizzo, nell'ordine
	nak  int   // indirizzo del blocco rifiutato in scrittura (-1 nessuno)
	dir  byte  // terzo byte delle risposte (0x00 sulla radio vera)
	deaf bool  // la radio non risponde
}

func (rt69Fam) emu(m *genModel, img []byte) genEmu {
	e := &rt69Emu{mem: bytes.Repeat([]byte{0xFF}, rt69Size), nak: -1}
	if len(img) >= rt69Size {
		copy(e.mem[rt69From:], img[rt69From:rt69Size])
	}
	// limiti della radio: due bande uguali, tutto quello che il programma ammette
	for _, a := range []int{0x0044, 0x0054} {
		binary.BigEndian.PutUint32(e.mem[a:], rt69LoHz/10)
		binary.BigEndian.PutUint32(e.mem[a+4:], rt69HiHz/10)
	}
	return e
}

// image: la parte che il programma legge (sotto 0x0480 tutto a FF)
func (e *rt69Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	img := bytes.Repeat([]byte{0xFF}, rt69Size)
	copy(img[rt69From:], e.mem[rt69From:])
	return img
}

func (e *rt69Emu) note(c byte, addr int) {
	e.mu.Lock()
	e.log = append(e.log, int(c)<<16|addr)
	e.mu.Unlock()
}

func (e *rt69Emu) reply(p Port, cmd byte, payload []byte) {
	h := []byte{0x7D, cmd, e.dir, byte(len(payload))}
	cs := sum8(append(append([]byte{}, h[1:]...), payload...))
	p.Write(append(append(h, payload...), cs))
}

func (e *rt69Emu) serve(p Port) error {
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
		body, err := emuRead(p, int(h[2])+1)
		if err != nil {
			return err
		}
		pl, cs := body[:len(body)-1], body[len(body)-1]
		if h[1] != 0xFF || sum8(append(append([]byte{}, h...), pl...)) != cs {
			return fmt.Errorf("pacchetto errato")
		}
		if e.deaf {
			continue
		}
		switch h[0] {
		case wx8CmdID:
			e.note(h[0], 0)
			e.reply(p, wx8CmdID, append([]byte("RT-69GPS"), make([]byte, 35)...))
		case wx8CmdRd:
			a, n := int(pl[0])<<8|int(pl[1]), int(pl[2])
			if a+n > len(e.mem) {
				return fmt.Errorf("lettura fuori dalla memoria %04X", a)
			}
			e.note(h[0], a)
			e.mu.Lock()
			d := append([]byte{pl[0], pl[1]}, e.mem[a:a+n]...)
			e.mu.Unlock()
			e.reply(p, wx8CmdRd, d)
		case wx8CmdWr:
			a := int(pl[0])<<8 | int(pl[1])
			if a < rt69From || a+len(pl)-2 > len(e.mem) {
				return fmt.Errorf("scrittura fuori dall'area dei dati %04X", a)
			}
			e.note(h[0], a)
			if a == e.nak {
				e.reply(p, wx8CmdWr, []byte{0, 0})
				continue
			}
			e.mu.Lock()
			copy(e.mem[a:], pl[2:])
			e.mu.Unlock()
			e.reply(p, wx8CmdWr, []byte{pl[0], pl[1]})
		case wx8CmdEnd:
			e.note(h[0], 0)
			return nil
		default:
			return fmt.Errorf("comando sconosciuto %d", h[0])
		}
	}
}
