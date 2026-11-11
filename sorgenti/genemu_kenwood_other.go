//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: radio Kenwood simulate per le prove.

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// TH-D74 / TH-D75: comandi a righe («ID», «0M PROGRAM»), poi blocchi da 256 byte
type thd74Emu struct {
	mu     sync.Mutex
	o      *thd74Opt
	id     string
	img    []byte
	writes int
	silent bool // radio spenta: nessuna risposta
}

func (thd74Fam) emu(m *genModel, img []byte) genEmu { return newThd74Emu(m, img) }
func (thd72Fam) emu(m *genModel, img []byte) genEmu { return newThd74Emu(m, img) }

func newThd74Emu(m *genModel, img []byte) *thd74Emu {
	o := m.Opt.(*thd74Opt)
	mem := make([]byte, o.bytes())
	copy(mem, img)
	return &thd74Emu{o: o, id: o.id, img: mem}
}

func (e *thd74Emu) block(h []byte) int {
	if e.o.d72 {
		return int(h[1]) | int(h[2])<<8
	}
	return int(h[0])<<8 | int(h[1])
}

func (e *thd74Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *thd74Emu) serve(p Port) error {
	for {
		// comandi a righe
		var line []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if b[0] == '\r' {
				break
			}
			line = append(line, b[0])
		}
		if e.silent {
			continue
		}
		switch cmd := strings.TrimSpace(string(line)); cmd {
		case "ID":
			p.Write([]byte("ID " + e.id + "\r"))
			continue
		case "0M PROGRAM":
			p.Write([]byte("0M\r"))
		default:
			p.Write([]byte("?\r"))
			continue
		}
		// programmazione
		for prog := true; prog; {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			switch c[0] {
			case 'R':
				h, err := emuRead(p, 4)
				if err != nil {
					return nil
				}
				blk := e.block(h)
				if (blk+1)*256 > len(e.img) {
					continue
				}
				e.mu.Lock()
				d := append(e.o.head('W', blk), e.img[blk*256:blk*256+256]...)
				e.mu.Unlock()
				p.Write(d)
				if a, err := emuRead(p, 1); err != nil || a[0] != 0x06 {
					return nil
				}
				p.Write([]byte{0x06})
			case 'W':
				h, err := emuRead(p, 4+256)
				if err != nil {
					return nil
				}
				blk := e.block(h)
				if (blk+1)*256 > len(e.img) {
					p.Write([]byte{0x15})
					continue
				}
				e.mu.Lock()
				copy(e.img[blk*256:], h[4:])
				e.writes++
				e.mu.Unlock()
				p.Write([]byte{0x06})
			case 'E':
				prog = false
			}
		}
	}
}

// 4.17: palmari Kenwood programmati un canale per volta (TH-F6/F7, TH-G71,
// TH-D7, TH-K2): righe «MR», «MW», «MNA» («ME», «MN» sulla TH-K2).
type kwlEmu struct {
	mu   sync.Mutex
	m    *genModel
	img  []byte
	log  []string // comandi di scrittura ricevuti
	zero bool     // risponde «MNA 0,nnn,nome» alla richiesta «MNA nnn»
	mute bool     // radio che non risponde
}

func (kwlFam) emu(m *genModel, img []byte) genEmu {
	mem := kwlBlank(m)
	if len(img) == len(mem) {
		copy(mem, img)
	}
	return &kwlEmu{m: m, img: mem}
}

func (e *kwlEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kwlEmu) serve(p Port) error {
	for {
		var line []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if b[0] == '\r' {
				break
			}
			line = append(line, b[0])
		}
		if e.mute {
			continue
		}
		p.Write([]byte(e.answer(string(line)) + "\r"))
	}
}

func (e *kwlEmu) answer(line string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	o := kwlo(e.m)
	ix := kwlIdxs[o.kind]
	switch line {
	case "":
		return "?"
	case "ID":
		return "ID " + o.id
	case "AI 0":
		return "AI 0"
	}
	k := strings.IndexByte(line, ' ')
	if k < 0 {
		return "?"
	}
	cmd, args := line[:k], line[k+1:]
	fl := strings.Split(args, ",")
	rd, wr, nm := "MR", "MW", "MNA"
	numAt, splitAt, nameNum := ix.num, ix.num-1, ix.num-1
	if o.kind == kwlK2 {
		rd, wr, nm = "ME", "ME", "MN"
		splitAt, nameNum = -1, 0
	}
	slot := func(at int) (int, bool) {
		if at >= len(fl) || len(fl[at]) != ix.numW {
			return 0, false
		}
		n := int(kwlInt(fl[at]))
		return n, n >= 0 && n < e.m.Slots
	}
	f := kwlFam{}
	if cmd == nm {
		n, ok := slot(nameNum)
		if !ok {
			return "N"
		}
		r := f.rec(e.img, n)
		if r[0] != 1 {
			return "N"
		}
		if len(fl) == nameNum+1 {
			if e.zero && o.kind == kwlF6 {
				return cmd + " 0," + args + "," + kwlText(r[kwlNameAt:kwlNameAt+kwlNameLen])
			}
			return line + "," + kwlText(r[kwlNameAt:kwlNameAt+kwlNameLen])
		}
		name := strings.Join(fl[nameNum+1:], ",")
		if len(name) > e.m.NameMax {
			return "N"
		}
		e.log = append(e.log, line)
		copy(r[kwlNameAt:kwlNameAt+kwlNameLen], make([]byte, kwlNameLen))
		copy(r[kwlNameAt:], name)
		return line
	}
	if cmd != rd && cmd != wr {
		return "?"
	}
	n, ok := slot(numAt)
	if !ok {
		return "N"
	}
	r := f.rec(e.img, n)
	split := splitAt >= 0 && fl[splitAt] == "1"
	if len(fl) == numAt+1 && cmd == rd {
		// lettura (sulla TH-K2 lettura e scrittura hanno lo stesso comando)
		if r[0] != 1 || split && r[1] != 1 {
			return "N"
		}
		if split {
			return line + "," + kwlText(r[kwlSplitAt:kwlSplitAt+11]) + ",0"
		}
		return cmd + " " + kwlText(r[kwlSpecAt:kwlSpecAt+kwlSpecLen])
	}
	e.log = append(e.log, line)
	if split {
		if r[0] != 1 || len(fl) != numAt+3 || len(fl[numAt+1]) != 11 {
			return "N"
		}
		r[1] = 1
		copy(r[kwlSplitAt:], fl[numAt+1])
		return line
	}
	if len(fl) == numAt+1 || len(fl) == numAt+2 && fl[numAt+1] == "" {
		copy(r, make([]byte, kwlRec)) // cancellazione
		return line
	}
	if len(fl) != ix.n || len(args) > kwlSpecLen || len(fl[ix.freq]) != ix.freqW || len(fl[ix.offs]) != ix.offsW {
		return "N"
	}
	name := append([]byte{}, r[kwlNameAt:kwlNameAt+kwlNameLen]...)
	copy(r, make([]byte, kwlRec))
	r[0] = 1
	copy(r[kwlSpecAt:], args)
	copy(r[kwlNameAt:], name)
	return line
}

// 4.17: palmari Kenwood TK della serie 60 (blocchi di 8 byte)
type kwp60Emu struct {
	mu  sync.Mutex
	img []byte
}

func (kwp60Fam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, kwp60Size)
	copy(mem, img)
	return &kwp60Emu{img: mem}
}

func (e *kwp60Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kwp60Emu) serve(p Port) error {
	if err := emuWaitFor(p, "PROGRAM"); err != nil {
		return nil
	}
	p.Write([]byte{0x06})
	if err := emuExpect(p, "\x02"); err != nil {
		return err
	}
	e.mu.Lock()
	id := kwp60Ident(e.img)
	e.mu.Unlock()
	p.Write(id)
	if err := emuExpect(p, "\x06"); err != nil {
		return err
	}
	p.Write([]byte{0x06})
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		a := int(h[1])<<8 | int(h[2])
		if h[3] != 8 || a+8 > kwp60Size || a%8 != 0 {
			return fmt.Errorf("comando non valido % X", h)
		}
		switch h[0] {
		case 'R':
			e.mu.Lock()
			p.Write(append([]byte{'W', h[1], h[2], 8}, e.img[a:a+8]...))
			e.mu.Unlock()
			if err := emuExpect(p, "\x06"); err != nil {
				return err
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, 8)
			if err != nil {
				return nil
			}
			if a > kwp60Top {
				return fmt.Errorf("scrittura oltre la zona scrivibile (0x%04X)", a)
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", h[0])
		}
	}
}

// 4.17: palmari Kenwood TK della serie 60G (blocchi di 256 byte con somma). La
// sigla del programma che ha scritto la radio (5 byte a 0xBB) resta fissa: CHIRP
// ci scrive «Chirp», IU0VYT Flash non la tocca, e nel confronto non conta.
type kw60gEmu struct {
	mu  sync.Mutex
	img []byte
}

func (kw60gFam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, kw60gSize)
	copy(mem, img)
	copy(mem[kw60gSoft:], "Chirp")
	return &kw60gEmu{img: mem}
}

func (e *kw60gEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kw60gEmu) serve(p Port) error {
	for {
		if err := emuWaitFor(p, "PROGRAM"); err != nil {
			return nil
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		e.mu.Lock()
		id := append([]byte{}, e.img[kw60gID:kw60gID+8]...)
		e.mu.Unlock()
		p.Write(id)
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
	session:
		for {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if c[0] == 'E' {
				break session
			}
			h, err := emuRead(p, 2)
			if err != nil {
				return nil
			}
			blk := int(h[0])<<8 | int(h[1])
			if blk >= kw60gSize/256 {
				return fmt.Errorf("blocco non valido %d", blk)
			}
			switch c[0] {
			case 'R':
				e.mu.Lock()
				d := append([]byte{}, e.img[blk*256:blk*256+256]...)
				e.mu.Unlock()
				if bytes.Equal(d, bytes.Repeat([]byte{0xFF}, 256)) {
					p.Write([]byte{'Z', 0xFF})
				} else {
					p.Write(append(append([]byte{'W'}, d...), kwtkSum(d)))
				}
				if err := emuExpect(p, "\x06"); err != nil {
					return err
				}
				p.Write([]byte{0x06})
			case 'W', 'Z':
				d := bytes.Repeat([]byte{0xFF}, 256)
				if c[0] == 'W' {
					b, err := emuRead(p, 257)
					if err != nil {
						return nil
					}
					if kwtkSum(b[:256]) != b[256] {
						p.Write([]byte{0x15})
						continue
					}
					d = b[:256]
				} else if _, err := emuRead(p, 1); err != nil {
					return nil
				}
				e.mu.Lock()
				copy(e.img[blk*256:], d)
				copy(e.img[kw60gSoft:], "Chirp")
				e.mu.Unlock()
				p.Write([]byte{0x06})
			default:
				return fmt.Errorf("comando sconosciuto %02X", c[0])
			}
		}
	}
}

// 4.18: Kenwood TK-2140 / TK-3140 (blocchi di 256 byte, bit della scansione a parte)
type kwx140Emu struct {
	mu  sync.Mutex
	img []byte
}

func (kwx140Fam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, kwx140Size)
	copy(mem, img)
	return &kwx140Emu{img: mem[:kwx140Size]}
}

func (e *kwx140Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kwx140Emu) serve(p Port) error {
	for {
		if err := emuWaitFor(p, "PROGRAM"); err != nil {
			return nil
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		e.mu.Lock()
		id := append([]byte{}, e.img[kwx140ID:kwx140ID+8]...)
		e.mu.Unlock()
		p.Write(id)
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "P"); err != nil {
			return err
		}
		p.Write([]byte("V1.00\x00\x00\x00\x00\x00"))
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
	session:
		for {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			switch c[0] {
			case 'E':
				break session
			case 'R', 'W', 'Z':
				h, err := emuRead(p, 2)
				if err != nil {
					return nil
				}
				blk := int(h[1])
				if h[0] != 0 || blk >= 0x80 {
					return fmt.Errorf("blocco non valido % X", h)
				}
				e.mu.Lock()
				d := append([]byte{}, e.img[blk*256:blk*256+256]...)
				e.mu.Unlock()
				if c[0] == 'R' {
					if bytes.Equal(d, bytes.Repeat(d[:1], 256)) {
						p.Write([]byte{'Z', d[0]})
					} else {
						p.Write(append(append([]byte{'W'}, d...), kwtkSum(d)))
					}
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
					p.Write([]byte{0x06})
					continue
				}
				if blk >= 0x54 && blk < 0x60 {
					return fmt.Errorf("scrittura di un blocco protetto (0x%02X)", blk)
				}
				if c[0] == 'W' {
					b, err := emuRead(p, 257)
					if err != nil {
						return nil
					}
					if kwtkSum(b[:256]) != b[256] {
						p.Write([]byte{0x15})
						continue
					}
					d = b[:256]
				} else {
					b, err := emuRead(p, 1)
					if err != nil {
						return nil
					}
					d = bytes.Repeat(b, 256)
				}
				e.mu.Lock()
				copy(e.img[blk*256:], d)
				e.mu.Unlock()
				p.Write([]byte{0x06})
			case 'S', 'X':
				h, err := emuRead(p, 3)
				if err != nil {
					return nil
				}
				a := int(h[0])<<8 | int(h[1])
				if h[2] != 0x10 || (a != 0x320 && a != 0x330) {
					return fmt.Errorf("zona della scansione non valida % X", h)
				}
				off := kwx140Scan + a - 0x320
				if c[0] == 'S' {
					e.mu.Lock()
					p.Write(append([]byte{'X'}, e.img[off:off+16]...))
					e.mu.Unlock()
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
					p.Write([]byte{0x06})
					continue
				}
				b, err := emuRead(p, 17)
				if err != nil {
					return nil
				}
				sum := 0
				for _, x := range b[:16] {
					sum += int(x)
				}
				if byte(sum%255) != b[16] {
					p.Write([]byte{0x15})
					continue
				}
				e.mu.Lock()
				copy(e.img[off:], b[:16])
				e.mu.Unlock()
				p.Write([]byte{0x06})
				if err := emuExpect(p, "\x06"); err != nil {
					return err
				}
			default:
				return fmt.Errorf("comando sconosciuto %02X", c[0])
			}
		}
	}
}

// 4.18: Kenwood TK-280 / TK-380 (blocchi di 256 byte con somma, conferma 06/06
// dopo ogni blocco)
type kwx80Emu struct {
	mu  sync.Mutex
	img []byte
}

func (kwx80Fam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, kwx80Size)
	copy(mem, img)
	return &kwx80Emu{img: mem[:kwx80Size]}
}

func (e *kwx80Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *kwx80Emu) serve(p Port) error {
	ack := func() error {
		if err := emuExpect(p, "\x06"); err != nil {
			return err
		}
		p.Write([]byte{0x06})
		return nil
	}
	for {
		if err := emuWaitFor(p, "PROGRAM"); err != nil {
			return nil
		}
		p.Write([]byte{0x06})
		if err := emuExpect(p, "\x02"); err != nil {
			return err
		}
		e.mu.Lock()
		id := append([]byte{}, e.img[kwx80ID:kwx80ID+8]...)
		e.mu.Unlock()
		p.Write(id)
		if err := ack(); err != nil {
			return err
		}
		if err := emuExpect(p, "P"); err != nil {
			return err
		}
		p.Write([]byte("v2.00k\xff\xff\xff\xff"))
		if err := ack(); err != nil {
			return err
		}
		if err := emuExpect(p, "O"); err != nil {
			return err
		}
	session:
		for {
			c, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if c[0] == 'E' {
				break session
			}
			h, err := emuRead(p, 2)
			if err != nil {
				return nil
			}
			blk := int(h[0])<<8 | int(h[1])
			if blk >= kwx80Size/256 {
				return fmt.Errorf("blocco non valido %d", blk)
			}
			switch c[0] {
			case 'R':
				e.mu.Lock()
				d := append([]byte{}, e.img[blk*256:blk*256+256]...)
				e.mu.Unlock()
				if bytes.Equal(d, bytes.Repeat([]byte{0xFF}, 256)) {
					p.Write([]byte{'Z', 0xFF})
				} else {
					p.Write(append(append([]byte{'W'}, d...), kwtkSum(d)))
				}
				if err := ack(); err != nil {
					return err
				}
			case 'W', 'Z':
				d := bytes.Repeat([]byte{0xFF}, 256)
				if c[0] == 'W' {
					b, err := emuRead(p, 257)
					if err != nil {
						return nil
					}
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
					if kwtkSum(b[:256]) != b[256] || (blk > 0x50 && blk < 0x60) {
						p.Write([]byte{0x15})
						continue
					}
					d = b[:256]
				} else {
					if _, err := emuRead(p, 1); err != nil {
						return nil
					}
					if err := emuExpect(p, "\x06"); err != nil {
						return err
					}
				}
				e.mu.Lock()
				copy(e.img[blk*256:], d)
				e.mu.Unlock()
				p.Write([]byte{0x06})
			default:
				return fmt.Errorf("comando sconosciuto %02X", c[0])
			}
		}
	}
}
