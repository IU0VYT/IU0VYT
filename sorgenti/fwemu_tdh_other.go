//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"
)

// tdhEmu: caricatore di TD-H3 / TD-H8 / RT-730 secondo la cattura delle porte
// pubblicata da nicsure (firmware_flash_protocol.txt)
type tdhEmu struct {
	mu       sync.Mutex
	hello    []byte // presentazione attesa
	data     []byte
	state    string // wait | hs | data | done | fail
	errs     []string
	dropAck  int           // blocco (da 1) senza conferma
	window   time.Duration // tempo concesso dopo l'ultimo 0xA5
	late     bool
	extraA5  time.Duration // 0xA5 dopo la presentazione
	powerOff bool          // radio spenta: nessun 0xA5
}

func (e *tdhEmu) fail(f string, a ...interface{}) {
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf(f, a...))
	e.state = "fail"
	e.mu.Unlock()
}

func (e *tdhEmu) serve(p Port, baud int) {
	if baud != tdhBaud {
		e.fail("velocità %d", baud)
		return
	}
	if e.powerOff {
		return
	}
	stopA5 := make(chan struct{})
	go func() { // la radio accesa con il PTT ripete 0xA5
		for {
			select {
			case <-stopA5:
				return
			case <-time.After(15 * time.Millisecond):
				_ = p.Write([]byte{tdhReady})
			}
		}
	}()
	h := make([]byte, len(e.hello))
	if err := p.ReadFull(context.Background(), h, 30*time.Second); err != nil {
		close(stopA5)
		return
	}
	if !bytes.Equal(h, e.hello) {
		e.fail("presentazione errata %x", h)
		// continua a inviare 0xA5 come la radio vera
		time.Sleep(6 * time.Second)
		close(stopA5)
		return
	}
	time.Sleep(e.extraA5)
	close(stopA5)
	last := time.Now()
	e.mu.Lock()
	e.state = "data"
	e.mu.Unlock()
	for n := 0; ; n++ {
		pk := make([]byte, 36)
		if err := p.ReadFull(context.Background(), pk, 10*time.Second); err != nil {
			e.fail("blocco %d mancante", n)
			return
		}
		if n == 0 && time.Since(last) > e.window+15*time.Millisecond {
			e.mu.Lock()
			e.late = true
			e.mu.Unlock()
		}
		if pk[0] != 0xA1 && pk[0] != 0xA2 {
			e.fail("tipo del blocco %d: %02X", n, pk[0])
			return
		}
		if int(pk[1])<<8|int(pk[2]) != n {
			e.fail("numero del blocco %d: %02X%02X", n, pk[1], pk[2])
			return
		}
		var s byte
		for _, b := range pk[4:] {
			s += b
		}
		if s != pk[3] {
			e.fail("somma del blocco %d", n)
			return
		}
		e.mu.Lock()
		e.data = append(e.data, pk[4:]...)
		e.mu.Unlock()
		if e.dropAck == n+1 {
			continue
		}
		_ = p.Write([]byte{tdhAck})
		if pk[0] == 0xA2 {
			e.mu.Lock()
			e.state = "done"
			e.mu.Unlock()
			return
		}
	}
}
