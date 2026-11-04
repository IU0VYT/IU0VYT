//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata in modalità di aggiornamento per il programma della RA89
// («TH8600-UpDAta Soft»): registra i comandi ricevuti per il confronto con il
// programma originale (verifica/ra89).

import (
	"bytes"
	"fmt"
	"sync"
)

type thEmu struct {
	mu     sync.Mutex
	hello  byte
	log    []string
	errs   []string
	data   []byte
	state  string // wait | start | data | end | done
	nakAt  int    // blocco (da 1) a cui rispondere «AA 03 EF» (0: mai)
	quiet  int    // blocco (da 1) a cui non rispondere (0: mai)
	wrong  bool   // radio di un altro modello: nessuna risposta alla presentazione
	blocks int
}

func newThEmu(hello byte) *thEmu { return &thEmu{hello: hello, state: "wait"} }

func (e *thEmu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }

func (e *thEmu) snapshot() ([]string, []string, string, []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state, append([]byte{}, e.data...)
}

func (e *thEmu) serve(p Port, baud int) {
	e.mu.Lock()
	e.note("apertura %d", baud)
	e.mu.Unlock()
	for {
		h, err := emuRead(p, 2)
		if err != nil {
			return
		}
		e.mu.Lock()
		var reply []byte
		switch {
		case h[0] == 0xAA:
			rest, err := emuRead(p, 1)
			if err != nil {
				e.mu.Unlock()
				return
			}
			if h[1] == 0x10 { // AA 10 00 00 EF
				r2, err := emuRead(p, 2)
				if err != nil {
					e.mu.Unlock()
					return
				}
				rest = append(rest, r2...)
			}
			cmd := append(append([]byte{}, h...), rest...)
			e.note("% X", cmd)
			switch {
			case bytes.Equal(cmd, []byte{0xAA, e.hello, 0xEF}) && e.state == "wait" && !e.wrong:
				e.state, reply = "start", []byte{0xAA, 0x01, 0xEF}
			case bytes.Equal(cmd, []byte{0xAA, 0x10, 0x00, 0x00, 0xEF}) && e.state == "start":
				e.state, reply = "data", []byte{0xAA, 0x11, 0xEF}
			case bytes.Equal(cmd, []byte{0xAA, 0xFF, 0xEF}) && e.state == "end":
				e.state, reply = "done", []byte{0xAA, 0xFF, 0xEF}
			case e.wrong:
			default:
				e.errs = append(e.errs, fmt.Sprintf("comando inatteso % X nello stato %s", cmd, e.state))
			}
		case h[0] == 0x55:
			n := int(h[1])
			if n == 0 {
				if _, err := emuRead(p, 1); err != nil {
					e.mu.Unlock()
					return
				}
				e.note("55 00 00")
				if e.state != "data" {
					e.errs = append(e.errs, "fine dei dati nello stato "+e.state)
				}
				e.state, reply = "end", []byte{0xAA, 0x02, 0xEF}
				break
			}
			e.mu.Unlock()
			d, err := emuRead(p, n+1)
			e.mu.Lock()
			if err != nil {
				e.mu.Unlock()
				return
			}
			var s byte
			for _, b := range d[:n] {
				s += b
			}
			e.blocks++
			e.note("blocco %d: %d byte, somma %02X", e.blocks, n, d[n])
			switch {
			case e.state != "data":
				e.errs = append(e.errs, "blocco nello stato "+e.state)
			case s != d[n]:
				e.errs = append(e.errs, fmt.Sprintf("somma errata nel blocco %d", e.blocks))
				reply = []byte{0xAA, 0x03, 0xEF}
			case e.blocks == e.quiet:
			case e.blocks == e.nakAt:
				reply = []byte{0xAA, 0x03, 0xEF}
			default:
				e.data = append(e.data, d[:n]...)
				reply = []byte{0xAA, 0x02, 0xEF}
			}
		default:
			e.errs = append(e.errs, fmt.Sprintf("byte inattesi % X", h))
		}
		e.mu.Unlock()
		if reply != nil {
			_ = p.Write(reply)
		}
	}
}
