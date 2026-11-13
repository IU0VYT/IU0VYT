//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.14: radio Yaesu simulate per le prove.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FT-4X / FT-65 (cavo SCU-35): ogni byte ricevuto torna indietro come eco
type ft4Emu struct {
	mu      sync.Mutex
	img     []byte
	id      []byte // 16 byte, l'ultimo è il sottotipo
	writes  int
	noEcho  bool // cavo senza eco
	badSum  int  // prime letture con la somma di controllo sbagliata
	program bool
}

func (ft4Fam) emu(m *genModel, img []byte) genEmu {
	o := ft4o(m)
	id := make([]byte, 16)
	copy(id, o.id)
	copy(id[9:], "V100")
	if o.sub3 {
		id[15] = 3
	}
	mem := make([]byte, ft4Size)
	copy(mem, img)
	return &ft4Emu{img: mem, id: id}
}

func (e *ft4Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *ft4Emu) serve(p Port) error {
	get := func(n int) ([]byte, error) {
		b, err := emuRead(p, n)
		if err == nil && !e.noEcho {
			p.Write(b)
		}
		return b, err
	}
	for {
		c, err := get(1)
		if err != nil {
			return nil
		}
		switch c[0] {
		case 'P':
			r, err := get(6)
			if err != nil {
				return err
			}
			if string(r) != "ROGRAM" {
				return fmt.Errorf("atteso PROGRAM, ricevuto P%q", r)
			}
			e.program = true
			p.Write([]byte("QX\x06"))
		case 0x02:
			if !e.program {
				continue
			}
			p.Write(append(append([]byte{}, e.id...), 0x06))
		case 'R':
			h, err := get(3)
			if err != nil {
				return err
			}
			a := int(h[0])<<8 | int(h[1])
			if !e.program || h[2] != 16 || a+16 > len(e.img) {
				continue
			}
			e.mu.Lock()
			body := append([]byte{h[0], h[1], 16}, e.img[a:a+16]...)
			sum := ft4Sum(body)
			if e.badSum > 0 {
				e.badSum--
				sum++
			}
			e.mu.Unlock()
			out := append([]byte{'W'}, body...)
			p.Write(append(out, sum, 0x06))
		case 'W':
			b, err := get(3 + 16 + 2)
			if err != nil {
				return err
			}
			a := int(b[0])<<8 | int(b[1])
			if !e.program || b[2] != 16 || a+16 > len(e.img) || a == 0 || ft4Sum(b[:19]) != b[19] || b[20] != 0x06 {
				continue // nessuna conferma
			}
			e.mu.Lock()
			copy(e.img[a:], b[3:19])
			e.writes++
			e.mu.Unlock()
			p.Write([]byte{0x06})
		case 'E':
			r, err := get(2)
			if err != nil {
				return err
			}
			if string(r) == "ND" {
				e.program = false
				p.Write([]byte{0x06})
			}
		}
	}
}

// yaesuEmuWait (4.34): quanto la radio simulata aspetta i primi dati prima di
// mettersi a inviare. Con 300 ms la scrittura fatta da CHIRP a volte arrivava
// tardi (il programma di confronto apre il collegamento e poi carica
// l'immagine) e la prova cadeva una volta su cinque.
const yaesuEmuWait = 1500 * time.Millisecond

// FT-60: la copia parte dalla radio. Se entro poco non arriva nulla la radio
// simulata invia la sua memoria (come dopo [PTT]); se arrivano dati li riceve
// (come dopo [MONI]). Ogni byte ricevuto torna indietro come eco.
type ft60Emu struct {
	mu     sync.Mutex
	img    []byte
	writes int
	silent bool // radio non in modalità clone: non invia e non conferma
}

func (ft60Fam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, ft60Size)
	copy(mem, img)
	return &ft60Emu{img: mem}
}

func (e *ft60Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *ft60Emu) serve(p Port) error {
	if e.silent {
		for {
			if _, err := emuRead(p, 1); err != nil {
				return nil
			}
		}
	}
	get := func(n int) ([]byte, error) {
		b, err := emuRead(p, n)
		if err == nil {
			p.Write(b)
		}
		return b, err
	}
	first := make([]byte, 1)
	if err := p.ReadFull(context.Background(), first, yaesuEmuWait); err != nil {
		if err != errTimeout {
			return nil
		}
		// invio
		src := e.image()
		wait := func() error {
			b, err := get(1)
			if err != nil {
				return err
			}
			if b[0] != 0x06 {
				return fmt.Errorf("attesa la conferma, ricevuto %02X", b[0])
			}
			return nil
		}
		p.Write(src[:8])
		if err := wait(); err != nil {
			return err
		}
		for a := 8; a < len(src); a += 64 {
			end := a + 64
			if end > len(src) {
				end = len(src)
			}
			p.Write(src[a:end])
			if err := wait(); err != nil {
				return err
			}
		}
		return nil
	}
	// ricezione
	p.Write(first)
	rest, err := get(7)
	if err != nil {
		return err
	}
	buf := append(first, rest...)
	if string(buf[:5]) != "AH017" {
		return fmt.Errorf("intestazione %q", buf[:5])
	}
	p.Write([]byte{0x06})
	for len(buf) < ft60Size {
		n := 64
		if ft60Size-len(buf) < n {
			n = ft60Size - len(buf)
		}
		b, err := get(n)
		if err != nil {
			return err
		}
		buf = append(buf, b...)
		p.Write([]byte{0x06})
	}
	var sum byte
	for _, b := range buf[:0x6FC8] {
		sum += b
	}
	if sum != buf[0x6FC8] {
		return fmt.Errorf("somma di controllo errata")
	}
	e.mu.Lock()
	copy(e.img, buf)
	e.writes++
	e.mu.Unlock()
	return nil
}

// 4.15: palmari Yaesu con la copia in due blocchi (VX-3, VX-6, FT-70D, FT1D…):
// come il cavo a un filo, ogni byte ricevuto torna indietro
type yaesuEmu struct {
	mu     sync.Mutex
	o      *yaesuOpt
	img    []byte
	writes int
	silent bool // radio non in modalità clone
	noEcho bool // cavo a due fili
}

func newYaesuEmu(o *yaesuOpt, img []byte) *yaesuEmu {
	mem := make([]byte, o.size)
	copy(mem, img)
	yaesuSums(o, mem) // la radio manda sempre le somme di controllo giuste
	return &yaesuEmu{o: o, img: mem}
}

func (e *yaesuEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *yaesuEmu) serve(p Port) error {
	if e.silent {
		for {
			if _, err := emuRead(p, 1); err != nil {
				return nil
			}
		}
	}
	get := func(n int) ([]byte, error) {
		b, err := emuRead(p, n)
		if err == nil && !e.noEcho {
			p.Write(b)
		}
		return b, err
	}
	first := make([]byte, 1)
	// la radio vera parte quando si preme il tasto: qui, se entro un secondo e mezzo
	// non arriva nulla, è la radio a inviare
	if err := p.ReadFull(context.Background(), first, 1500*time.Millisecond); err != nil {
		if err != errTimeout {
			return nil
		}
		// invio: primo blocco, conferma, tutto il resto
		src := e.image()
		p.Write(src[:e.o.head])
		b, err := get(1)
		if err != nil {
			return err
		}
		if b[0] != 0x06 {
			return fmt.Errorf("attesa la conferma, ricevuto %02X", b[0])
		}
		p.Write(src[e.o.head:])
		return nil
	}
	// ricezione
	if !e.noEcho {
		p.Write(first)
	}
	rest, err := get(e.o.head - 1)
	if err != nil {
		return err
	}
	buf := append(first, rest...)
	if string(buf[:len(e.o.model)]) != e.o.model {
		return fmt.Errorf("intestazione %q", buf[:5])
	}
	p.Write([]byte{0x06})
	for len(buf) < e.o.size {
		n := e.o.chunk
		if e.o.size-len(buf) < n {
			n = e.o.size - len(buf)
		}
		b, err := get(n)
		if err != nil {
			return err
		}
		buf = append(buf, b...)
	}
	if !yaesuSumsOK(e.o, buf) {
		return fmt.Errorf("somma di controllo errata")
	}
	e.mu.Lock()
	copy(e.img, buf)
	e.writes++
	e.mu.Unlock()
	return nil
}

func (f vxFam) emu(m *genModel, img []byte) genEmu { return newYaesuEmu(f.opt(m).clone, img) }

func (f ysfFam) emu(m *genModel, img []byte) genEmu { return newYaesuEmu(f.opt(m).clone, img) }

// 4.34: VX-170 / VX-177, FT-7800, FT-8900: come la FT-60, con modello, lunghezza
// e blocchi propri. Se entro poco non arriva nulla la radio simulata invia la
// sua memoria; se arrivano dati li riceve. Ogni byte ricevuto torna indietro.
type ftxEmu struct {
	mu     sync.Mutex
	o      *ftxOpt
	img    []byte
	writes int
	silent bool // radio non in modalità clone: non invia e non conferma
}

func (ftxFam) emu(m *genModel, img []byte) genEmu {
	o := ftxo(m)
	mem := make([]byte, o.size)
	copy(mem, img)
	return &ftxEmu{o: o, img: mem}
}

func (e *ftxEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

// blocchi: 8 byte, poi blocchi interi, poi la somma di controllo
func (e *ftxEmu) next(a int) int {
	switch {
	case a == 0:
		return 8
	case a == e.o.size-1:
		return e.o.size
	case a+e.o.blk > e.o.size-1:
		return e.o.size - 1
	}
	return a + e.o.blk
}

func (e *ftxEmu) serve(p Port) error {
	if e.silent {
		for {
			if _, err := emuRead(p, 1); err != nil {
				return nil
			}
		}
	}
	get := func(n int) ([]byte, error) {
		b, err := emuRead(p, n)
		if err == nil {
			p.Write(b)
		}
		return b, err
	}
	first := make([]byte, 1)
	if err := p.ReadFull(context.Background(), first, yaesuEmuWait); err != nil {
		if err != errTimeout {
			return nil
		}
		// invio
		src := e.image()
		for a := 0; a < len(src); {
			end := e.next(a)
			p.Write(src[a:end])
			b, err := get(1)
			if err != nil {
				return err
			}
			if b[0] != 0x06 {
				return fmt.Errorf("attesa la conferma, ricevuto %02X", b[0])
			}
			a = end
		}
		return nil
	}
	// ricezione
	p.Write(first)
	rest, err := get(7)
	if err != nil {
		return err
	}
	buf := append(first, rest...)
	if string(buf[:len(e.o.id)]) != e.o.id {
		return fmt.Errorf("intestazione %q", buf[:6])
	}
	p.Write([]byte{0x06})
	for len(buf) < e.o.size {
		b, err := get(e.next(len(buf)) - len(buf))
		if err != nil {
			return err
		}
		buf = append(buf, b...)
		p.Write([]byte{0x06})
	}
	if ftxSum(buf) != buf[e.o.size-1] {
		return fmt.Errorf("somma di controllo errata")
	}
	e.mu.Lock()
	copy(e.img, buf)
	e.writes++
	e.mu.Unlock()
	return nil
}
