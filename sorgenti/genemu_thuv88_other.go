//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: th_uv88.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio simulata della famiglia TYT TH-UV88 (protocollo di th_uv88.py di CHIRP).
// Un saluto con il nome di un altro modello non riceve risposta.

import (
	"bytes"
	"fmt"
	"sync"
)

type thuv88Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
}

func (thuv88Fam) emu(m *genModel, img []byte) genEmu {
	e := &thuv88Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < thuv88MemSize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, thuv88MemSize-len(e.img))...)
	}
	e.img = e.img[:thuv88MemSize]
	return e
}

func (e *thuv88Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

// un comando: FE FE EE EF + byte del comando + resto fino a FD (per E4 la
// lunghezza è fissa: indirizzo, lunghezza, dati, somma)
func (e *thuv88Emu) frame(p Port) (byte, []byte, error) {
	var win []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return 0, nil, err
		}
		win = append(win, c)
		if len(win) > 4 {
			win = win[1:]
		}
		if bytes.Equal(win, []byte{0xFE, 0xFE, 0xEE, 0xEF}) {
			break
		}
	}
	cmd, err := emuByte(p)
	if err != nil {
		return 0, nil, err
	}
	if cmd == 0xE4 {
		b, err := emuRead(p, 6+thuv88Block+2)
		return cmd, b, err
	}
	var rest []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return 0, nil, err
		}
		if c == 0xFD {
			return cmd, rest, nil
		}
		rest = append(rest, c)
	}
}

func (e *thuv88Emu) serve(p Port) error {
	tag := []byte(thuv88o(e.m).tag)
	mode := byte(0)
	for {
		cmd, b, err := e.frame(p)
		if err != nil {
			return nil
		}
		switch cmd {
		case 0xE0:
			if !bytes.Equal(b, tag) {
				continue // saluto di un altro modello
			}
			id := append(append([]byte{0xFE, 0xFE, 0xEF, 0xEE, 0xE1}, tag...), bytes.Repeat([]byte{0x00}, 26)...)
			p.Write(append(id, 0xFD))
			mode = 0xE0
		case 0xE2, 0xE3:
			if mode != 0xE0 || !bytes.Equal(b, tag) {
				return fmt.Errorf("comando %02X senza saluto", cmd)
			}
			mode = cmd
			p.Write(thuv88Ack)
		case 0xE5:
			mode = 0
		case 0xEB:
			if mode != 0xE2 || len(b) != 6 {
				return fmt.Errorf("lettura fuori dalla modalità di lettura")
			}
			a, n := int(b[2])<<8|int(b[3]), int(b[4])<<8|int(b[5])
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if a < len(e.img) {
				copy(d, e.img[a:])
			}
			e.mu.Unlock()
			out := append(append([]byte{}, b...), d...)
			p.Write(append(append(append([]byte{0xFE, 0xFE, 0xEF, 0xEE, 0xE4}, out...), thuv88Sum(out)), 0xFD))
		case 0xE4:
			if mode != 0xE3 {
				return fmt.Errorf("scrittura fuori dalla modalità di scrittura")
			}
			a := int(b[2])<<8 | int(b[3])
			if thuv88Sum(b[:6+thuv88Block]) != b[6+thuv88Block] || b[7+thuv88Block] != 0xFD {
				return fmt.Errorf("somma di controllo errata nel blocco 0x%04X", a)
			}
			e.mu.Lock()
			if a+thuv88Block <= len(e.img) {
				copy(e.img[a:], b[6:6+thuv88Block])
			}
			e.mu.Unlock()
			p.Write(thuv88Ack)
		default:
			return fmt.Errorf("comando inatteso %02X", cmd)
		}
	}
}
