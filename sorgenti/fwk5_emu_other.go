//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Caricatore della Quansheng UV-K5 simulato (per le prove): annuncio 0x0518
// ripetuto, risposta alla versione, scrittura dei blocchi da 256 byte con la
// conferma 0x051A. Registra i byte ricevuti per il confronto con uvtools.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

type k5BootEmu struct {
	mu    sync.Mutex
	bl    string // versione del caricatore, es. "2.00.06"
	flash []byte
	rx    []byte // byte ricevuti dal PC
	final int
	stop  int // blocco dopo il quale la radio smette di rispondere (-1: mai)
}

func newK5BootEmu(bl string) *k5BootEmu {
	return &k5BootEmu{bl: bl, flash: bytes.Repeat([]byte{0xFF}, 0xF000), stop: -1}
}

func (e *k5BootEmu) beacon() []byte {
	b := []byte{0x18, 0x05, 0x20, 0x00, 0x01, 0x02, 0x02, 0x06, 0x1c, 0x53, 0x50, 0x4a, 0x37, 0x47, 0xff, 0x0f, 0x8c, 0x00, 0x53, 0x00}
	v := make([]byte, 16)
	copy(v, e.bl)
	return append(b, v...)
}

func (e *k5BootEmu) serve(p Port) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wmu sync.Mutex
	write := func(b []byte) { wmu.Lock(); p.Write(k5BootPacket(b)); wmu.Unlock() }
	// annuncio ogni 20 ms finché il PC non presenta la versione
	announcing := make(chan struct{})
	go func() {
		t := time.NewTicker(20 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-announcing:
				return
			case <-ctx.Done():
				return
			case <-t.C:
				write(e.beacon())
			}
		}
	}()
	stopped := false
	blocks := 0
	for {
		f, raw, err := k5EmuFrame(ctx, p)
		if err != nil {
			if !stopped {
				close(announcing)
			}
			return nil
		}
		e.mu.Lock()
		e.rx = append(e.rx, raw...)
		e.mu.Unlock()
		switch {
		case len(f) >= 4 && f[0] == 0x30 && f[1] == 0x05:
			if !stopped {
				close(announcing)
				stopped = true
			}
			write(e.beacon())
		case len(f) >= 16+0x100 && f[0] == 0x19 && f[1] == 0x05:
			if e.stop >= 0 && blocks >= e.stop {
				continue
			}
			a := int(f[8])<<8 | int(f[9])
			e.mu.Lock()
			e.final = int(f[10]) << 8
			if a+0x100 <= len(e.flash) {
				copy(e.flash[a:], f[16:16+0x100])
			}
			e.mu.Unlock()
			blocks++
			write([]byte{0x1a, 0x05, 0x08, 0x00, 0x8a, 0x8d, 0x9f, 0x1d, f[8], f[9], 0x00, 0x00})
		default:
			return fmt.Errorf("comando inatteso % X", f[:min(4, len(f))])
		}
	}
}

// k5EmuFrame: pacchetto del PC (dati decifrati, CRC controllato) e byte grezzi
func k5EmuFrame(ctx context.Context, p Port) ([]byte, []byte, error) {
	h, err := emuRead(p, 4)
	if err != nil {
		return nil, nil, err
	}
	if h[0] != 0xAB || h[1] != 0xCD {
		return nil, nil, fmt.Errorf("intestazione inattesa % X", h)
	}
	n := int(binary.LittleEndian.Uint16(h[2:]))
	d, err := emuRead(p, n+4)
	if err != nil {
		return nil, nil, err
	}
	dec := uvk5Xor(d[:n+2])
	if binary.LittleEndian.Uint16(dec[n:]) != crc16XModem(dec[:n]) || d[n+2] != 0xDC || d[n+3] != 0xBA {
		return nil, nil, fmt.Errorf("CRC o chiusura del pacchetto non validi")
	}
	return dec[:n], append(h, d...), nil
}
