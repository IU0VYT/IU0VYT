//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Radtel RT-5D: pacchetti A5 con CRC, ingresso in programmazione
// nell'ordine del programma del costruttore, sezioni lette e scritte un
// pacchetto per volta (drv_rt5d.go). Tiene l'elenco dei comandi ricevuti.

import (
	"fmt"
	"sync"
)

type rt5dEmu struct {
	mu     sync.Mutex
	img    []byte
	log    []int // comando<<16 | parametro, nell'ordine
	badCRC int   // numero di risposte da guastare (CRC errato)
	refuse byte  // comando da rifiutare con EE
}

func (rt5dFam) emu(m *genModel, img []byte) genEmu {
	e := &rt5dEmu{img: make([]byte, rt5dSize())}
	for i := range e.img {
		e.img[i] = 0xFF
	}
	copy(e.img, img)
	return e
}

func (e *rt5dEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rt5dEmu) serve(p Port) error {
	step := 0 // 0 attesa del saluto, 1 password, 2 versione, 3 sezioni
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c != 0xA5 {
			return fmt.Errorf("byte inatteso %02X", c)
		}
		h, err := emuRead(p, 5)
		if err != nil {
			return err
		}
		n := int(h[3])<<8 | int(h[4])
		d, err := emuRead(p, n+2)
		if err != nil {
			return err
		}
		if crc16XModem(append(append([]byte{}, h...), d[:n]...)) != uint16(d[n])<<8|uint16(d[n+1]) {
			return fmt.Errorf("CRC errato nel comando %02X", h[0])
		}
		cmd, par, data := h[0], int(h[1])<<8|int(h[2]), d[:n]
		e.mu.Lock()
		e.log = append(e.log, int(cmd)<<16|par)
		var out []byte
		ok := true
		switch {
		case cmd == e.refuse:
			ok = false
		case cmd == 0x02:
			ok = string(data) == rt5dHello
			step = 1
		case cmd == 0x05 && step >= 1: // anche ripetuto (nuovo invio dopo una risposta persa)
			ok = len(data) == 6
			if step < 2 {
				step = 2
			}
		case cmd == 0x46 && step >= 2:
			out = make([]byte, 0x80)
			copy(out, "RT-5D V2.0")
			step = 3
		case cmd == 0x01 && step == 3:
			step = 0
		case step == 3:
			ok = false
			off := 0
			for _, s := range rt5dSecs {
				if par < s.n && n == s.size {
					a := off + par*s.size
					if cmd == s.rd {
						out, ok = append([]byte{}, e.img[a:a+s.size]...), true
					} else if cmd == s.wr && s.wr != 0 {
						copy(e.img[a:], data)
						ok = true
					}
				}
				off += s.size * s.n
			}
		default:
			ok = false
		}
		bad := false
		if e.badCRC > 0 && ok {
			e.badCRC--
			bad = true
		}
		e.mu.Unlock()
		if !ok {
			cmd, out = 0xEE, nil
		}
		f := rt5dFrame(cmd, par, out)
		if bad {
			f[len(f)-1] ^= 0x55
		}
		p.Write(f)
	}
}
