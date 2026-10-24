//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radioddity_gm30.py.

package main

// Radio simulata Baofeng GM-15 Pro / Radioddity GM-30 (protocollo di radioddity_gm30.py
// di CHIRP). Le aree di memoria stanno in posizioni scelte dalla radio (mappa dei tipi).

import (
	"bytes"
	"fmt"
	"sync"
)

type gm30Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
	loc map[int]int // posizione (i) → indice dell'area nell'immagine
}

func (gm30Fam) emu(m *genModel, img []byte) genEmu {
	e := &gm30Emu{m: m, img: append([]byte{}, img...), loc: map[int]int{}}
	if len(e.img) < gm30Size {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, gm30Size-len(e.img))...)
	}
	e.img = e.img[:gm30Size]
	// posizioni mescolate, come una radio che ha spostato le aree
	for k, i := range []int{3, 7, 1, 10, 5, 12, 2, 9, 14, 6} {
		e.loc[i] = k
	}
	return e
}

func (e *gm30Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *gm30Emu) off(blk, pre byte) (int, bool) {
	k, ok := e.loc[int(blk>>4)]
	if !ok {
		return 0, false
	}
	return k*0x1000 + int(blk&0x0F)*0x100 + int(pre), true
}

func (e *gm30Emu) ack(p Port) error {
	b, err := emuByte(p)
	if err != nil {
		return err
	}
	if b != 0x06 {
		return fmt.Errorf("conferma 06 attesa, ricevuto %02X", b)
	}
	p.Write([]byte{0x06})
	return nil
}

func (e *gm30Emu) serve(p Port) error {
	ident := gm30o(e.m).ident
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 'P', 'S':
			r, err := emuRead(p, 6)
			if err != nil {
				return err
			}
			switch string(append([]byte{c}, r...)) {
			case "PSEARCH":
				p.Write(append([]byte{0x06}, ident...))
			case "PASSSTA":
				p.Write([]byte{0x50, 0, 0})
			case "SYSINFO", string(ident):
				p.Write([]byte{0x06})
			default:
				return fmt.Errorf("comando inatteso %q", append([]byte{c}, r...))
			}
		case 0x56:
			a, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			n := 13
			if bytes.Equal(a, []byte{0, 0, 0, 0x0A}) {
				n = 11
			}
			p.Write(append([]byte{0x56}, bytes.Repeat([]byte{0x30}, n-1)...))
			if err := e.ack(p); err != nil {
				return err
			}
		case 0xFF:
			if _, err := emuRead(p, 4); err != nil {
				return err
			}
		case 0x02:
			p.Write([]byte("GM15PRO\x00"))
			if err := e.ack(p); err != nil {
				return err
			}
		case 0x52:
			h, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			out := append([]byte{0x57}, h...)
			if h[0] == 0xFF {
				t := byte(0xFF)
				if k, ok := e.loc[int(h[1]>>4)]; ok {
					t = gm30Types[k]
				}
				out = append(out, t)
			} else {
				d := bytes.Repeat([]byte{0xFF}, int(h[3]))
				if x, ok := e.off(h[1], h[0]); ok {
					e.mu.Lock()
					copy(d, e.img[x:])
					e.mu.Unlock()
				}
				out = append(out, d...)
			}
			p.Write(out)
			if err := e.ack(p); err != nil {
				return err
			}
		case 0x57:
			h, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			d, err := emuRead(p, int(h[3]))
			if err != nil {
				return err
			}
			if x, ok := e.off(h[1], h[0]); ok {
				e.mu.Lock()
				copy(e.img[x:x+len(d)], d)
				e.mu.Unlock()
			}
			p.Write([]byte{0x06})
		case 0x06, 0x00:
			// uscita dalla programmazione (06 06 00)
		default:
			return fmt.Errorf("comando inatteso %02X", c)
		}
	}
}
