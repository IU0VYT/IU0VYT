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
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

type wxEmu struct {
	mu     sync.Mutex
	img    []byte
	hello  func(p Port) error // presentazione prima dei blocchi
	base   int                // posizione nell'immagine dell'indirizzo 0 della radio
	noAck  bool               // lettura senza conferma 06 (Retevis H777 V4, RT21H)
	noAck0 bool               // 4.9: primo blocco (indirizzo 0) senza conferma (Retevis RB28)
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
			if e.noAck || e.noAck0 && a == 0 {
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

func (e echoPort) unwrapPort() Port { return e.Port }

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
	emuRequireLine(p, true, false) // 4.9: 9600 baud con parità pari
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
		emuRequireLine(p, false, true) // 4.9: due bit di stop
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

// 4.9: Wouxun KG-UVD1P, KG-UV6, KG-816/818, KG-805G (wouxun.py)
func (wxoFam) emu(m *genModel, img []byte) genEmu {
	q := wxoo(m).query[0]
	return newWxEmu(img, 0x2000, func(p Port) error {
		if err := emuWaitFor(p, q); err != nil {
			return err
		}
		p.Write([]byte("\x02\x00KG669V\x00"))
		if err := emuExpect(p, "\x02\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
}

// 4.9: TYT TH-UV8000 (th_uv8000.py)
func (thuv8000Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, thuv8000Size, func(p Port) error {
		if err := emuWaitFor(p, "PROGRAMa"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "PROGRAMb"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte("\x50TC8800E"))
		if err := emuExpect(p, "MTC88CUMHS3E7BN-"); err != nil {
			return err
		}
		p.Write([]byte{0x80})
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
	e.noAck = true
	return e
}

// 4.9: TDXone TD-Q8A
func (tdq8aFam) emu(m *genModel, img []byte) genEmu {
	return newWxEmu(img, 0x2008, func(p Port) error {
		if err := emuWaitFor(p, "\x02PYNCRAM"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte("P3107Q8A"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	})
}

// 4.9: TIDRADIO TD-M11
type tdm11Emu struct {
	mu    sync.Mutex
	img   []byte
	magic string
}

func (f tdm11Fam) emu(m *genModel, img []byte) genEmu {
	return &tdm11Emu{img: newWxEmu(img, 0x450, nil).img, magic: f.o.magic}
}

func (e *tdm11Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *tdm11Emu) serve(p Port) error {
	if err := emuWaitFor(p, e.magic); err != nil {
		return err
	}
	p.Write([]byte{'A'})
	if err := emuExpect(p, "\xff\xff\xff\xff\xff\xff"); err != nil {
		return err
	}
	p.Write([]byte{'A'})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
			return nil
		}
		switch c {
		case 'R':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			a, n := int(h[0])|int(h[1])<<8, int(h[2])
			if a+n > len(e.img) {
				return fmt.Errorf("indirizzo fuori dalla memoria")
			}
			d := e.image()[a : a+n]
			p.Write(append(append([]byte{}, d...), tdm11Sum(d)))
		case 'W':
			h, err := emuRead(p, 2+16+1)
			if err != nil {
				return err
			}
			a := int(h[0]) | int(h[1])<<8
			if a+16 > len(e.img) || tdm11Sum(h[2:18]) != h[18] {
				return fmt.Errorf("blocco non valido")
			}
			e.mu.Lock()
			copy(e.img[a:], h[2:18])
			e.mu.Unlock()
			p.Write([]byte{'A'})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Retevis C2 (trame STX/ETX con DLE)
type c2Emu struct {
	mu  sync.Mutex
	img []byte
}

func (c2Fam) emu(m *genModel, img []byte) genEmu {
	return &c2Emu{img: newWxEmu(img, c2MemSize, nil).img}
}

func (e *c2Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

// c2Area: posizione e lunghezza di una tabella nella memoria
func c2Area(tbl, idx byte) (int, int, bool) {
	i := int(idx)
	switch tbl {
	case c2TblSet:
		return c2SetOff, c2SetSize, i == 0
	case c2TblChan:
		return c2ChanOff + i*c2ChanSize, c2ChanSize, i < 256
	case c2TblExtra:
		return c2ExtraOff + i*c2ExtraSize, c2ExtraSize, i < c2ExtraCount
	case c2TblFinal:
		return c2FinalOff, c2FinalSize, i == 0
	case c2TblSys:
		return c2SysconfOff, c2SysSize, i == 0
	case c2TblRfcal:
		return c2RfcalOff + i*c2RfcalSize, c2RfcalSize, i < c2RfcalCount
	}
	return 0, 0, false
}

func (e *c2Emu) serve(p Port) error {
	ctx := context.Background()
	for {
		in, err := c2ReadFrame(ctx, p, 30*time.Second)
		if err != nil || in == nil {
			return err
		}
		pl, err := c2Unframe(in)
		if err != nil || len(pl) < 4 {
			return fmt.Errorf("trama non valida")
		}
		cmd, tbl, idx := pl[1], pl[2], pl[3]
		resp := []byte{0x01, cmd, tbl, idx}
		switch cmd {
		case c2CmdIdent:
			resp = append(resp, e.image()[:c2IdentSize]...)
		case c2CmdRead:
			off, n, ok := c2Area(tbl, idx)
			if !ok {
				return fmt.Errorf("tabella %02X/%d sconosciuta", tbl, idx)
			}
			resp = append(resp, e.image()[off:off+n]...)
		case c2CmdWrite:
			off, n, ok := c2Area(tbl, idx)
			d := pl[4:]
			if !ok || len(d) > n {
				return fmt.Errorf("scrittura non valida nella tabella %02X/%d", tbl, idx)
			}
			e.mu.Lock()
			copy(e.img[off:], d)
			e.mu.Unlock()
		case c2CmdVerify:
		case c2CmdFinalize:
			p.Write(c2Frame(resp))
			return nil
		default:
			return fmt.Errorf("comando sconosciuto %02X", cmd)
		}
		p.Write(c2Frame(resp))
	}
}

// 4.9: Retevis RB28, RB628, RB48, RB648 (primo blocco letto senza conferma)
type rb28Emu struct{ *wxEmu }

func (rb28Fam) emu(m *genModel, img []byte) genEmu {
	o := rb28o(m)
	e := newWxEmu(img, o.size, nil)
	e.noAck0 = true
	fp := "P32073\x02\xFF"
	if m.ID == "rt-rb628" {
		fp = "P32073\x00\xFF"
	}
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, o.magic); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write([]byte(fp))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return &rb28Emu{e}
}

// 4.9: Retevis RB15, RB615
type rb15Emu struct {
	mu  sync.Mutex
	img []byte
}

func (rb15Fam) emu(m *genModel, img []byte) genEmu {
	return &rb15Emu{img: newWxEmu(img, 0x7A0, nil).img}
}

func (e *rb15Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rb15Emu) serve(p Port) error {
	if err := emuWaitFor(p, "21\x05\x10x"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case '2':
			if _, err := emuRead(p, 4); err != nil { // «1 05 EE V»: fine
				return err
			}
			return nil
		case 'R':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			a := int(h[0])<<8 | int(h[1])
			if h[2] != tdm11Sum([]byte{'R', h[0], h[1]}) || a+16 > len(e.img) {
				return fmt.Errorf("lettura non valida")
			}
			b := append([]byte{'R', h[0], h[1]}, e.image()[a:a+16]...)
			p.Write(append(b, tdm11Sum(b)))
		case 'W':
			h, err := emuRead(p, 2+16+1)
			if err != nil {
				return err
			}
			a := int(h[0])<<8 | int(h[1])
			if h[18] != tdm11Sum(append([]byte{'W'}, h[:18]...)) || a+16 > len(e.img) {
				return fmt.Errorf("scrittura non valida")
			}
			e.mu.Lock()
			copy(e.img[a:], h[2:18])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Retevis RB17P (lettura senza conferma; la sessione finisce senza 'E')
type rb17pEmu struct{ *wxEmu }

func (rb17pFam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x1800, nil)
	e.noAck = true
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "PGA588"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "M\x02"); err != nil {
			return err
		}
		p.Write(bytes.Repeat([]byte{0xFF}, 8))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return &rb17pEmu{e}
}

// 4.9: Retevis RT23 (blocchi con somma di controllo)
type rt23Emu struct {
	mu  sync.Mutex
	img []byte
}

func (rt23Fam) emu(m *genModel, img []byte) genEmu {
	return &rt23Emu{img: newWxEmu(img, 0x1000, nil).img}
}

func (e *rt23Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rt23Emu) serve(p Port) error {
	if err := emuWaitFor(p, "PROIUAM"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	if err := emuExpect(p, "\x02"); err != nil {
		return err
	}
	p.Write([]byte("P31183\xFF\xFF"))
	if err := emuExpect(p, "\x06"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a := int(h[0])<<8 | int(h[1])
		if a+16 > len(e.img) {
			return fmt.Errorf("indirizzo fuori dalla memoria")
		}
		switch c {
		case 'R':
			d := e.image()[a : a+16]
			p.Write(append(append([]byte{'W', h[0], h[1], h[2]}, d...), tdm11Sum(d)))
		case 'W':
			d, err := emuRead(p, 17)
			if err != nil {
				return err
			}
			if d[16] != tdm11Sum(d[:16]) {
				return fmt.Errorf("somma di controllo errata")
			}
			e.mu.Lock()
			copy(e.img[a:], d[:16])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Retevis RT26
type rt26Emu struct{ *wxEmu }

func (rt26Fam) emu(m *genModel, img []byte) genEmu {
	e := newWxEmu(img, 0x400, nil)
	e.hello = func(p Port) error {
		if err := emuWaitFor(p, "PROGRAMa"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "PROGRAMb"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		p.Write(append([]byte("PDK80"), bytes.Repeat([]byte{0xFF}, 11)...))
		if err := emuExpect(p, "MXT8KCUMHS1X7BN/"); err != nil {
			return err
		}
		p.Write([]byte{0xB2})
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		return p.Write([]byte{0x06})
	}
	return &rt26Emu{e}
}

// 4.9: Retevis RT76P (la radio risponde «R…» e non chiede conferme)
type rt76pEmu struct {
	mu  sync.Mutex
	img []byte
}

func (rt76pFam) emu(m *genModel, img []byte) genEmu {
	return &rt76pEmu{img: newWxEmu(img, 0x2000, nil).img}
}

func (e *rt76pEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rt76pEmu) serve(p Port) error {
	if err := emuWaitFor(p, "PROGROMCD2U"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	if err := emuExpect(p, "F"); err != nil {
		return err
	}
	p.Write([]byte{0x01, 0x36, 0x01, 0x74, 0x04, 0x00, 0x05, 0x20})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
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
			p.Write(append([]byte{'R', h[0], h[1], h[2]}, e.image()[a:a+n]...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Retevis RT87 (la radio ripete ogni byte ricevuto)
type rt87Emu struct {
	mu  sync.Mutex
	img []byte
}

func (rt87Fam) emu(m *genModel, img []byte) genEmu {
	return &rt87Emu{img: newWxEmu(img, 0x2000, nil).img}
}

func (e *rt87Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *rt87Emu) serve(p Port) error {
	if err := emuWaitFor(p, "PGM2017"); err != nil {
		return err
	}
	p.Write([]byte("PGM2017\x06\x30"))
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
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
			p.Write(append([]byte{'R', h[0], h[1], h[2]}, append([]byte{'W', h[0], h[1], h[2]}, e.image()[a:a+n]...)...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write(append(append([]byte{'W', h[0], h[1], h[2]}, d...), 0x06))
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: AnyTone TERMN-8R, OBLTR-8R
type athtEmu struct {
	mu    sync.Mutex
	img   []byte
	ident string
}

func (f athtFam) emu(m *genModel, img []byte) genEmu {
	return &athtEmu{img: newWxEmu(img, 0x8000, nil).img, ident: f.ident}
}

func (e *athtEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *athtEmu) serve(p Port) error {
	if err := emuWaitFor(p, "PROGRAM"); err != nil {
		return err
	}
	p.Write([]byte("QX\x06"))
	if err := emuExpect(p, "\x02"); err != nil {
		return err
	}
	p.Write([]byte(("I" + e.ident + "         ")[:16]))
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == 'E' {
			if err := emuExpect(p, "ND"); err != nil {
				return err
			}
			return p.Write([]byte{0x06})
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
			b := append([]byte{'W', h[0], h[1], h[2]}, e.image()[a:a+n]...)
			p.Write(append(b, tdm11Sum(b[1:]), 0x06))
		case 'W':
			d, err := emuRead(p, n+2)
			if err != nil {
				return err
			}
			if d[n] != tdm11Sum(append(append([]byte{}, h...), d[:n]...)) || d[n+1] != 0x06 {
				return fmt.Errorf("blocco non valido")
			}
			e.mu.Lock()
			copy(e.img[a:], d[:n])
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Rugged RH5R-V2
type rh5rEmu struct {
	mu  sync.Mutex
	img []byte
}

func (rh5rFam) emu(m *genModel, img []byte) genEmu {
	return &rh5rEmu{img: newWxEmu(img, 0x2000, nil).img}
}

func (e *rh5rEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append(append([]byte{}, e.img...), []byte("PGM2015")...)
}

func (e *rh5rEmu) serve(p Port) error {
	if err := emuWaitFor(p, "PGM2015"); err != nil {
		return err
	}
	p.Write([]byte{0x06, 0x30})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
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
			e.mu.Lock()
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: Puxing PX-888K
type px888kEmu struct {
	mu  sync.Mutex
	img []byte
}

func (px888kFam) emu(m *genModel, img []byte) genEmu {
	return &px888kEmu{img: newWxEmu(img, 0x1000, nil).img}
}

func (e *px888kEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *px888kEmu) serve(p Port) error {
	if err := emuWaitFor(p, "XONLINE"); err != nil {
		return err
	}
	p.Write([]byte("PX888D\x00\xff"))
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == 'E' {
			return p.Write([]byte{0x06})
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a := int(h[0])<<8 | int(h[1])
		if h[2] != '@' || a+0x40 > len(e.img) {
			return fmt.Errorf("comando non valido")
		}
		switch c {
		case 'R':
			p.Write(append([]byte{'W', h[0], h[1], '@'}, e.image()[a:a+0x40]...))
		case 'W':
			d, err := emuRead(p, 0x40)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// 4.9: QYT KT-8R
type kt8rEmu struct {
	mu  sync.Mutex
	img []byte
}

func (kt8rFam) emu(m *genModel, img []byte) genEmu {
	return &kt8rEmu{img: newWxEmu(img, 0x4000, nil).img}
}

func (e *kt8rEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kt8rEmu) serve(p Port) error {
	if err := emuWaitFor(p, "\x55\x20\x17\x07\x03\xFF\xDC\x02"); err != nil {
		return err
	}
	p.Write(append([]byte{0x06}, e.image()[0x3F70:0x3F78]...))
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c != 0x06 {
			return fmt.Errorf("comando non valido %02X", c)
		}
		h, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		if h[0] != 'S' && h[0] != 'X' {
			return fmt.Errorf("comando sconosciuto %02X", h[0])
		}
		a3, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(a3[0])<<8|int(a3[1]), int(a3[2])
		if a+n > len(e.img) {
			return fmt.Errorf("indirizzo fuori dalla memoria")
		}
		if h[0] == 'S' {
			p.Write(append([]byte{0x06, 'X', a3[0], a3[1], a3[2]}, e.image()[a:a+n]...))
			continue
		}
		d, err := emuRead(p, n)
		if err != nil {
			return err
		}
		e.mu.Lock()
		copy(e.img[a:], d)
		e.mu.Unlock()
		p.Write([]byte{0x06})
	}
}

// 4.9: Retevis HA1G, HA1UV, HA2 (protocollo RDTP)
type haEmu struct {
	mu  sync.Mutex
	img []byte
	o   *ha1gOpt
}

func (ha1gFam) emu(m *genModel, img []byte) genEmu {
	o := ha1go(m)
	return &haEmu{img: newWxEmu(img, o.size, nil).img, o: o}
}

func (e *haEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *haEmu) region(id byte) (int, int, bool) {
	for _, r := range e.o.regions {
		if r.id == id {
			return r.start, r.len, true
		}
	}
	return 0, 0, false
}

func (e *haEmu) serve(p Port) error {
	ctx := context.Background()
	var wbuf []byte
	for {
		pkt, err := haRead(ctx, p, 30*time.Second)
		if err != nil || len(pkt) < 20 {
			return err
		}
		typ, code := pkt[14], pkt[15]
		idx, count := int(binary.LittleEndian.Uint16(pkt[16:])), int(binary.LittleEndian.Uint16(pkt[18:]))
		data := pkt[20 : len(pkt)-3]
		switch {
		case typ == 111:
			return nil
		case typ == 0:
			ans := append([]byte(e.o.model), make([]byte, 30)...)
			p.Write(haPacket(0, 1, 0, 1, ans))
		case code == 2:
			a, n, ok := e.region(typ)
			if !ok {
				return fmt.Errorf("zona %d sconosciuta", typ)
			}
			k := int(binary.LittleEndian.Uint16(data))
			tot := (n + 1023) / 1024
			end := (k + 1) * 1024
			if end > n {
				end = n
			}
			p.Write(haPacket(typ, 2, k, tot, e.image()[a+k*1024:a+end]))
		case code == 0:
			a, n, ok := e.region(typ)
			if !ok {
				return fmt.Errorf("zona %d sconosciuta", typ)
			}
			if idx == 0 {
				wbuf = nil
			}
			wbuf = append(wbuf, data...)
			if idx == count-1 {
				if len(wbuf) > n {
					return fmt.Errorf("zona %d troppo lunga", typ)
				}
				e.mu.Lock()
				copy(e.img[a:], wbuf)
				e.mu.Unlock()
			}
			p.Write(haPacket(typ, 0, idx, count, []byte{0}))
		default:
			return fmt.Errorf("comando sconosciuto %d/%d", typ, code)
		}
	}
}
