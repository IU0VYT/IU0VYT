//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radtel_rt900.py.

package main

// Radio simulata Bajeton BJ-7800 / BJ-7700 (protocollo di radtel_rt900.py di CHIRP)
// e BJ-9000 (programma del costruttore: comando 'M', blocchi da 128 byte).

import (
	"bytes"
	"fmt"
	"sync"
)

type rt900Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
}

func (rt900Fam) emu(m *genModel, img []byte) genEmu {
	n := (rt900o(m).memsize + 0x7F) &^ 0x7F
	e := &rt900Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < n {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, n-len(e.img))...)
	}
	e.img = e.img[:n]
	return e
}

func (e *rt900Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *rt900Emu) serve(p Port) error {
	o := rt900o(e.m)
	mg, err := emuRead(p, len(o.magic))
	if err != nil {
		return err
	}
	if !bytes.Equal(mg, o.magic) {
		return fmt.Errorf("identificazione inattesa %q", mg)
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 'F' {
		return fmt.Errorf("comando F atteso (%v)", err)
	}
	if o.fp1 != nil {
		p.Write(append(append([]byte{}, o.fp1[0]...), o.fp2[0]...))
	} else {
		p.Write([]byte{0x01, 0x36, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20, 0x02, 0x00, 0x02, 0x60, 0x01, 0x03, 0x30, 0x04})
	}
	if o.model != "" {
		if b, err := emuByte(p); err != nil || b != 'M' {
			return fmt.Errorf("comando M atteso (%v)", err)
		}
		p.Write([]byte(fmt.Sprintf("%-12s", o.model)))
	}
	cs, err := emuRead(p, len(o.crypt))
	if err != nil {
		return err
	}
	if !bytes.Equal(cs, o.crypt) {
		return fmt.Errorf("pacchetto di cifratura inatteso % X", cs)
	}
	p.Write([]byte{0x06})
	key := []byte(btSymbols[1])
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == 'E' {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if addr < len(e.img) {
				copy(d, e.img[addr:])
			}
			e.mu.Unlock()
			if addr < 0xF000 {
				btCrypt(d, key)
			}
			p.Write(append([]byte{'R', h[0], h[1], h[2]}, d...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			if addr < 0xF000 {
				btCrypt(d, key)
			}
			e.mu.Lock()
			if addr+n <= len(e.img) {
				copy(e.img[addr:], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %q", c)
		}
	}
}
