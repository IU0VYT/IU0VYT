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
	"sync/atomic"
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

// k5v3BootEmu (4.4): caricatore 7.x (UV-K5 V3, UV-K1, Retevis RA79 V3) simulato
// come descritto da quansheng_toolkit e Multi-UVTools: annuncio 0x0518 finché il
// PC non risponde tre volte con 0x0530 e le prime 4 lettere della versione del
// caricatore; poi blocchi 0x0519 (ora, numero, totale, 4 zeri, 256 byte) con la
// conferma 0x051A (ora, numero, esito).
type k5v3BootEmu struct {
	mu     sync.Mutex
	bl     string
	flash  []byte
	pages  int
	hs     int
	errs   []string
	nakAt  int // blocco (da 0) a cui rispondere una volta con un errore (-1: mai)
	silent int // blocco (da 0) a cui non rispondere più (-1: mai)
	done   bool
}

func newK5v3BootEmu(bl string) *k5v3BootEmu {
	return &k5v3BootEmu{bl: bl, nakAt: -1, silent: -1}
}

func (e *k5v3BootEmu) beacon() []byte {
	b := []byte{0x18, 0x05, 0x20, 0x00}
	b = append(b, bytes.Repeat([]byte{0x5A}, 16)...)
	v := make([]byte, 16)
	copy(v, e.bl)
	return append(b, v...)
}

func (e *k5v3BootEmu) fail(f string, a ...interface{}) {
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf(f, a...))
	e.mu.Unlock()
}

func (e *k5v3BootEmu) serve(p Port) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wmu sync.Mutex
	write := func(b []byte) { wmu.Lock(); p.Write(k5BootPacket(b)); wmu.Unlock() }
	var announce atomic.Bool
	announce.Store(true)
	go func() {
		t := time.NewTicker(30 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if announce.Load() {
					write(e.beacon())
				}
			}
		}
	}()
	var ts uint32
	next := 0
	nakDone := false
	for {
		f, _, err := k5EmuFrame(ctx, p)
		if err != nil {
			return nil
		}
		switch {
		case len(f) >= 8 && f[0] == 0x30 && f[1] == 0x05:
			if string(f[4:8]) != e.bl[:4] {
				e.fail("presentazione %q invece di %q", f[4:8], e.bl[:4])
				continue
			}
			e.mu.Lock()
			e.hs++
			if e.hs >= 3 {
				announce.Store(false)
			}
			e.mu.Unlock()
		case len(f) >= 16+0x100 && f[0] == 0x19 && f[1] == 0x05:
			e.mu.Lock()
			hs := e.hs
			e.mu.Unlock()
			if hs < 3 {
				e.fail("blocco prima della presentazione")
				continue
			}
			t := binary.LittleEndian.Uint32(f[4:])
			idx := int(binary.LittleEndian.Uint16(f[8:]))
			tot := int(binary.LittleEndian.Uint16(f[10:]))
			if next == 0 {
				ts = t
			}
			if t != ts || !bytes.Equal(f[12:16], make([]byte, 4)) || binary.LittleEndian.Uint16(f[2:]) != 0x010C {
				e.fail("intestazione del blocco %d % X", idx, f[:16])
			}
			if e.silent >= 0 && idx >= e.silent {
				continue
			}
			code := uint16(0)
			if idx == e.nakAt && !nakDone {
				nakDone, code = true, 1
			} else if idx != next {
				e.fail("blocco %d invece di %d", idx, next)
				code = 2
			} else {
				e.mu.Lock()
				e.flash = append(e.flash, f[16:16+0x100]...)
				e.pages, e.done = idx+1, idx+1 == tot
				e.mu.Unlock()
				next++
			}
			r := []byte{0x1a, 0x05, 0x08, 0x00, 0, 0, 0, 0, 0, 0, 0, 0}
			binary.LittleEndian.PutUint32(r[4:], t)
			binary.LittleEndian.PutUint16(r[8:], uint16(idx))
			binary.LittleEndian.PutUint16(r[10:], code)
			write(r)
		default:
			e.fail("comando inatteso % X", f[:min(4, len(f))])
		}
	}
}
