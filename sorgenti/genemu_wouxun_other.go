//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.8: radio simulate con il protocollo «Wouxun» (blocchi 'R'/'W')

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"
)

type wxEmu struct {
	mu    sync.Mutex
	img   []byte
	hello func(p Port) error // presentazione prima dei blocchi
	base  int                // posizione nell'immagine dell'indirizzo 0 della radio
	noAck bool               // lettura senza conferma 06 (Retevis H777 V4, RT21H)
}

func newWxEmu(img []byte, size int, hello func(p Port) error) *wxEmu {
	e := &wxEmu{img: append([]byte{}, img...), hello: hello}
	if len(e.img) < size {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, size-len(e.img))...)
	}
	e.img = e.img[:size]
	return e
}

func (e *wxEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *wxEmu) serve(p Port) error {
	if err := e.hello(p); err != nil {
		return err
	}
	for {
		c, err := emuByte(p)
		if err != nil || c == 0x06 || c == 'E' {
			return nil // fine della sessione
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			if e.base+a+n > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			d := append([]byte{}, e.img[e.base+a:e.base+a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if e.noAck {
				continue
			}
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if e.base+a+n > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			copy(e.img[e.base+a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// emuWaitFor: attende la sequenza scartando i byte precedenti (presentazioni di
// altri modelli durante l'identificazione automatica)
func emuWaitFor(p Port, s string) error {
	var win []byte
	for {
		b, err := emuByte(p)
		if err != nil {
			return err
		}
		win = append(win, b)
		if len(win) > len(s) {
			win = win[1:]
		}
		if string(win) == s {
			return nil
		}
	}
}

func emuExpect(p Port, s string) error {
	got, err := emuRead(p, len(s))
	if err != nil {
		return err
	}
	if string(got) != s {
		return fmt.Errorf("attesi %q, ricevuti %q", s, got)
	}
	return nil
}

func (thuv3rFam) emu(m *genModel, img []byte) genEmu {
	return newWxEmu(img, thuv3ro(m).memsize, func(p Port) error {
		if err := emuWaitFor(p, "PROGRAMa"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
}

func (kydFam) emu(m *genModel, img []byte) genEmu {
	n, _, _, _ := kydSizes(m)
	return newWxEmu(img, n, func(p Port) error {
		hello, id := "PROGRAM", "P32073\xF8\xFF"
		if kydo(m).ip620 {
			hello, id = "iUHOUXUN\x02", "\x06KG67\x01V\xF8"
		}
		if err := emuWaitFor(p, hello); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte(id))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
}

func (thuvf1Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x1010, nil)
	e.base = 0x10
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "PROG333"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write(e.image()[:16])
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}

// TH-UVF8D: blocchi da 32 byte confermati con "A", fine con "ENDR"/"ENDW"
type thuvf8dEmu struct {
	mu  sync.Mutex
	img []byte
}

func (thuvf8dFam) emu(m *genModel, img []byte) genEmu {
	return &thuvf8dEmu{img: newWxEmu(img, 0x4020, nil).img}
}

func (e *thuvf8dEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *thuvf8dEmu) serve(p Port) error {
	if err := emuWaitFor(p, "\x02PROGRAM"); err != nil {
		return err
	}
	p.Write([]byte("PG"))
	if err := emuExpect(p, "\x02"); err != nil {
		return err
	}
	p.Write(e.image()[:32])
	if err := emuExpect(p, "A"); err != nil {
		return err
	}
	p.Write([]byte("A"))
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		if string(h) == "ENDR" {
			return nil
		}
		if string(h) == "ENDW" {
			p.Write([]byte{0, 0, 0})
			return nil
		}
		a, n := int(h[1])<<8|int(h[2]), int(h[3])
		if 0x20+a+n > len(e.img) {
			return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
		}
		switch h[0] {
		case 'R':
			p.Write(append([]byte{'W', h[1], h[2], h[3]}, e.image()[0x20+a:0x20+a+n]...))
			if err := emuExpect(p, "A"); err != nil {
				return err
			}
			p.Write([]byte("A"))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[0x20+a:], d)
			e.mu.Unlock()
			p.Write([]byte("A"))
		default:
			return fmt.Errorf("comando sconosciuto %02X", h[0])
		}
	}
}

// KSUN M6: blocchi da 16 byte con somma di controllo, indirizzi da 0x0050
type ksunEmu struct {
	mu  sync.Mutex
	img []byte
}

func (ksunFam) emu(m *genModel, img []byte) genEmu {
	return &ksunEmu{img: newWxEmu(img, ksunSize, nil).img}
}

func (e *ksunEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *ksunEmu) serve(p Port) error {
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 'R', 'W':
			h, err := emuRead(p, 2)
			if err != nil {
				return err
			}
			a := (int(h[0])<<8 | int(h[1])) - ksunStart
			n := 1
			if c == 'W' {
				n += 0x10
			}
			rest, err := emuRead(p, n)
			if err != nil {
				return err
			}
			msg := append(append([]byte{c}, h...), rest[:n-1]...)
			if rest[n-1] != ksunSum(msg) {
				return fmt.Errorf("somma di controllo errata")
			}
			if a < 0 || a+0x10 > len(e.img) {
				return fmt.Errorf("indirizzo fuori dalla memoria")
			}
			e.mu.Lock()
			if c == 'R' {
				p.Write(ksunCmd(append(msg, e.img[a:a+0x10]...)...))
			} else {
				copy(e.img[a:], rest[:0x10])
				p.Write([]byte{0x06})
			}
			e.mu.Unlock()
		case 0x32:
			h, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			if h[1] == 0x05 && h[2] == 0x10 {
				p.Write([]byte{0x06})
			} else if h[2] == 0xEE {
				return nil
			}
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

func (puxingFam) emu(m *genModel, img []byte) genEmu {
	rd, _, _ := pxSizes(m)
	return newWxEmu(img, rd, func(p Port) error {
		if pxo(m).px2r {
			if err := emuWaitFor(p, "PROGRAM\x02"); err != nil {
				return err
			}
			p.Write([]byte{0x06})
			if err := emuExpect(p, "\x06"); err != nil {
				return err
			}
			return p.Write([]byte("PX-2R\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"))
		}
		if err := emuWaitFor(p, "\x02PROGRA"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "M\x02"); err != nil {
			return err
		}
		p.Write([]byte("PX777\x00\x00\x00"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
}

// Boblov X3 Plus: la radio ripete ogni byte ricevuto
type echoPort struct{ Port }

func (e echoPort) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	if err := e.Port.ReadFull(ctx, b, to); err != nil {
		return err
	}
	return e.Port.Write(b)
}

func (x3plusFam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x3F0, nil)
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, ".VKOGRAM"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte("SMP558\x00\x00"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return &x3Emu{e}
}

type x3Emu struct{ *wxEmu }

func (e *x3Emu) serve(p Port) error { return e.wxEmu.serve(echoPort{p}) }

// HG-UV98: blocchi da 64 byte, doppia conferma 06/06
type hguv98Emu struct {
	mu  sync.Mutex
	img []byte
}

func (hguv98Fam) emu(m *genModel, img []byte) genEmu {
	return &hguv98Emu{img: newWxEmu(img, 0x2000, nil).img}
}

func (e *hguv98Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *hguv98Emu) serve(p Port) error {
	ack := func(p Port) error {
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	if err := emuWaitFor(p, "NiNHSG0N"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	if err := emuExpect(p, "\x02"); err != nil {
		return err
	}
	p.Write([]byte("P3107\x00\x00\x00"))
	if err := ack(p); err != nil {
		return err
	}
	for {
		c, err := emuByte(p)
		if err != nil || c == 0x45 {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		if a+n > len(e.img) {
			return fmt.Errorf("indirizzo fuori dalla memoria")
		}
		switch c {
		case 'R':
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, e.image()[a:a+n]...))
			if err := ack(p); err != nil {
				return err
			}
			if err := ack(p); err != nil {
				return err
			}
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
			if err := ack(p); err != nil {
				return err
			}
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// Radioddity R2: "\x02TYOGRAM", identificativo, password (FF), blocchi da 8 byte
func (r2Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x3F0, nil)
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "\x02TYOGRAM"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte("\x06R2\x00\x00\x00\x00\x00"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x05"); err != nil {
			return err
		}
		p.Write(bytes.Repeat([]byte{0xFF}, 6))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}

// Retevis RT1: "PROGRAMa", "PROGRAMb", identificativo, chiave → B2, 06
func (rt1Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x400, nil)
	e.hello = func(p Port) error {
		for _, s := range []string{"PROGRAMa", "PROGRAMb"} {
			if err := emuExpect(p, s); err != nil {
				return err
			}
			p.Write([]byte{0x06})
		}
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write(append(e.image()[0x170:0x178], make([]byte, 8)...))
		if err := emuExpect(p, "MXT8KCUMHS1X7BN/"); err != nil {
			return err
		}
		p.Write([]byte{0xB2})
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}

func (h777v4Fam) emu(m *genModel, img []byte) genEmu {
	o := h777v4o(m)
	e := newWxEmu(img, o.memsize, nil)
	e.noAck = true
	e.hello = func(p Port) error {
		// la classe H777 di CHIRP prova prima il comando della V4: per l'RT21H
		// arriva solo il suo
		if err := emuWaitFor(p, "\x02"+o.cmd); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		id := append([]byte{}, o.idents[0]...)
		p.Write(append(id, make([]byte, 8-len(id))...))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}

func (tguv2pFam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x2030, nil)
	e.base = 0x30
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "\x02PnOGdAM"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x4D\x02"); err != nil {
			return err
		}
		p.Write([]byte("P5555\x00\x00\x00"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}

func (fd268Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x800, nil)
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "\x02PROGRA"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "M\x02"); err != nil {
			return err
		}
		p.Write(append(append([]byte{}, m.Opt.(*fd268Opt).ident...), 0, 0, 0, 0))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return e
}
