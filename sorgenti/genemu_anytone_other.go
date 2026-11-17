//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata della famiglia AnyTone (drv_anytone.go): memoria per indirizzo
// nelle aree dell'immagine, ingresso «PROGRAM» / 02, letture «R» da 16 o 64 byte
// (qdmr e questo programma usano 16, dmrconfig 64), scritture «W» da 16 byte con la
// somma di controllo, «END». Tiene l'elenco degli indirizzi scritti; le scritture
// fuori dalle aree dell'immagine (l'indice dei contatti di dmrconfig) sono
// accettate e tenute a parte.

import (
	"bytes"
	"fmt"
	"sync"
)

type atEmu struct {
	mu     sync.Mutex
	mem    []byte            // come l'immagine, ma con tutto quello che la radio contiene
	ident  string            // modello dichiarato
	extra  map[uint32][]byte // scritture fuori dalle aree dell'immagine
	writes []uint32          // indirizzi scritti, nell'ordine
	badSum int               // letture a cui rispondere con la somma sbagliata
	nak    uint32            // indirizzo di cui rifiutare la scrittura (0 nessuno)
	ended  int               // «END» ricevuti
}

func (atFam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, atSize)
	copy(mem, img)
	return &atEmu{mem: mem, ident: ato(m).ident, extra: map[uint32][]byte{}}
}

func (atFam) blankImg(m *genModel) []byte { return atSynthImage(m) }

// atSynthImage: memoria senza canali, zone, liste di scansione e contatti
func atSynthImage(m *genModel) []byte {
	img := bytes.Repeat([]byte{0xFF}, atSize)
	zero := func(off, n int) {
		for k := 0; k < n; k++ {
			img[off+k] = 0
		}
	}
	zero(0, 0x40)
	copy(img, ato(m).ident)
	zero(atOffZoneMap, 0x80)
	zero(atOffChMap, 0x240)
	zero(atOffSettings, 0x640)
	zero(atOffRadioID, 0x1F40)
	return img
}

// image: la memoria come la legge il programma (le voci non valide restano a FF)
func (e *atEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := append([]byte{}, e.mem...)
	for off := 0; off < atSize; off += 64 {
		if atSkip(e.mem, off) {
			for k := 0; k < 64; k++ {
				out[off+k] = 0xFF
			}
		}
	}
	return out
}

// posizione nell'immagine dell'indirizzo, -1 se è fuori dalle aree
func atEmuOff(addr uint32, n int) int {
	for _, rg := range atRegions {
		if addr >= rg.addr && addr+uint32(n) <= rg.addr+uint32(rg.size) {
			return rg.off + int(addr-rg.addr)
		}
	}
	return -1
}

func (e *atEmu) serve(p Port) error {
	prog := false
	var win []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if !prog {
			win = append(win, c)
			if len(win) > 7 {
				win = win[len(win)-7:]
			}
			if string(win) != "PROGRAM" {
				continue
			}
			win = nil
			prog = true
			p.Write([]byte("QX\x06"))
			continue
		}
		switch c {
		case 'P': // «PROGRAM» ripetuto
			b, err := emuRead(p, 6)
			if err != nil {
				return err
			}
			if string(b) != "ROGRAM" {
				return fmt.Errorf("comando inatteso P% X", b)
			}
			p.Write([]byte("QX\x06"))
		case 0x02:
			id := make([]byte, 16)
			id[0] = 'I'
			copy(id[1:8], e.ident)
			copy(id[9:], "V100")
			id[15] = 0x06
			p.Write(id)
		case 'R':
			h, err := emuRead(p, 5)
			if err != nil {
				return err
			}
			addr := uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
			n := int(h[4])
			if n != 16 && n != 64 {
				return fmt.Errorf("lettura di %d byte", n)
			}
			data := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if off := atEmuOff(addr, n); off >= 0 {
				copy(data, e.mem[off:off+n])
			} else if d, ok := e.extra[addr]; ok && len(d) == n {
				copy(data, d)
			}
			bad := e.badSum > 0
			if bad {
				e.badSum--
			}
			e.mu.Unlock()
			r := append(append([]byte{'W'}, h...), data...)
			sum := atSum(r[1:])
			if bad {
				sum ^= 0x55
			}
			p.Write(append(r, sum, 0x06))
		case 'W':
			h, err := emuRead(p, 5)
			if err != nil {
				return err
			}
			addr := uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
			n := int(h[4])
			d, err := emuRead(p, n+2)
			if err != nil {
				return err
			}
			if n != 16 || d[n+1] != 0x06 || atSum(append(append([]byte{}, h...), d[:n]...)) != d[n] {
				return fmt.Errorf("scrittura non valida a %08X", addr)
			}
			e.mu.Lock()
			if e.nak != 0 && e.nak == addr {
				e.mu.Unlock()
				p.Write([]byte{0x15})
				continue
			}
			e.writes = append(e.writes, addr)
			if off := atEmuOff(addr, n); off >= 0 {
				copy(e.mem[off:], d[:n])
			} else {
				e.extra[addr] = append([]byte{}, d[:n]...)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		case 'E':
			b, err := emuRead(p, 2)
			if err != nil {
				return err
			}
			if string(b) != "ND" {
				return fmt.Errorf("comando inatteso E% X", b)
			}
			e.mu.Lock()
			e.ended++
			e.mu.Unlock()
			prog = false
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando inatteso %02X", c)
		}
	}
}
