//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata in modalità di aggiornamento per i programmi della RA89 e della
// RT85 («TH8600-UpDAta Soft»): risponde alla presentazione del suo modello
// ripetendola, come si aspettano i programmi; registra i comandi ricevuti per il
// confronto con i programmi originali (verifica/ra89).

import (
	"bytes"
	"fmt"
	"sync"
	"time"
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
	// 4.5: pausa tra l'apertura o l'ultima risposta e ogni comando (per il
	// confronto delle pause dei programmi originali), con il tipo del comando
	gaps []thGap
	last time.Time
	// latency: tempo di risposta della radio (0: subito). Nelle prove con i
	// programmi originali 2 ms, come una radio vera: i programmi impostano la soglia
	// dell'evento DataReceived (3 byte) subito DOPO l'invio, e una risposta
	// istantanea della porta simulata può arrivare prima (il programma la legge a
	// metà e si ferma con «Communication failed»)
	latency time.Duration
}

type thGap struct {
	what string // aa (AA ..), blocco, fine
	d    time.Duration
}

func newThEmu(hello byte) *thEmu { return &thEmu{hello: hello, state: "wait"} }

func (e *thEmu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }

func (e *thEmu) gapList() []thGap {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]thGap{}, e.gaps...)
}

func (e *thEmu) snapshot() ([]string, []string, string, []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state, append([]byte{}, e.data...)
}

func (e *thEmu) serve(p Port, baud int) {
	e.mu.Lock()
	e.note("apertura %d", baud)
	e.last = time.Now()
	e.mu.Unlock()
	for {
		h, err := emuRead(p, 2)
		if err != nil {
			return
		}
		e.mu.Lock()
		gap := thGap{"aa", time.Since(e.last)}
		if h[0] == 0x55 {
			gap.what = "blocco"
			if h[1] == 0 {
				gap.what = "fine"
			}
		}
		e.gaps = append(e.gaps, gap)
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
				e.state, reply = "start", []byte{0xAA, e.hello, 0xEF} // la radio ripete la presentazione
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
			if e.latency > 0 {
				time.Sleep(e.latency)
			}
			_ = p.Write(reply)
			e.mu.Lock()
			e.last = time.Now()
			e.mu.Unlock()
		}
	}
}
