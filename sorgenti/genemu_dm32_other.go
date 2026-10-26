//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Baofeng DM-32UV: memoria flash a pagine da 4 KB con l'etichetta
// nell'ultimo byte, in posizioni mescolate (come una radio che ha spostato le
// pagine), protocollo descritto in drv_dm32.go.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

const (
	dm32EmuStart = 0x010000
	dm32EmuEnd   = 0x0FFFFF
)

type dm32Emu struct {
	mu    sync.Mutex
	flash []byte // da dm32EmuStart a dm32EmuEnd
}

func (dm32Fam) emu(m *genModel, img []byte) genEmu {
	e := &dm32Emu{flash: bytes.Repeat([]byte{0xFF}, dm32EmuEnd+1-dm32EmuStart)}
	n := len(e.flash) / dm32Page
	k := 0
	for i := range dm32Tags {
		if (i+1)*dm32Page > len(img) {
			break
		}
		pg := img[i*dm32Page : (i+1)*dm32Page]
		if bytes.Equal(pg, bytes.Repeat([]byte{0xFF}, dm32Page)) {
			continue // pagina che la radio non ha
		}
		// posizioni mescolate, con pagine libere in mezzo
		pos := (k*37 + 11) % n
		for !bytes.Equal(e.flash[pos*dm32Page:(pos+1)*dm32Page], bytes.Repeat([]byte{0xFF}, dm32Page)) {
			pos = (pos + 1) % n
		}
		copy(e.flash[pos*dm32Page:], pg)
		k++
	}
	return e
}

// immagine logica: la prima pagina con ogni etichetta
func (e *dm32Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	img := bytes.Repeat([]byte{0xFF}, len(dm32Tags)*dm32Page)
	for i, t := range dm32Tags {
		for a := 0; a < len(e.flash); a += dm32Page {
			if e.flash[a+dm32Page-1] == t {
				copy(img[i*dm32Page:], e.flash[a:a+dm32Page])
				break
			}
		}
	}
	return img
}

func (e *dm32Emu) serve(p Port) error {
	vinfo := func(i byte) []byte {
		switch i {
		case 1:
			return []byte("DM32.01.01.047")
		case 10:
			b := make([]byte, 8)
			binary.LittleEndian.PutUint32(b, dm32EmuStart)
			binary.LittleEndian.PutUint32(b[4:], dm32EmuEnd)
			return b
		}
		return []byte{i, 0x00}
	}
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
				p.Write(append([]byte{0x06}, dm32ModelID...))
			case "PASSSTA":
				p.Write([]byte{'P', 0x00, 0x00})
			case "SYSINFO", "PROGRAM":
				p.Write([]byte{0x06})
			default:
				return fmt.Errorf("comando inatteso %q", append([]byte{c}, r...))
			}
		case 'V':
			r, err := emuRead(p, 4)
			if err != nil {
				return err
			}
			d := vinfo(r[3])
			if r[2] == 0x40 {
				d = []byte{0x01}
			}
			p.Write(append([]byte{'V', r[3], byte(len(d))}, d...))
		case 'G':
			if _, err := emuRead(p, 5); err != nil {
				return err
			}
			p.Write(append([]byte{'S'}, make([]byte, 0x105)...))
		case 0xFF:
			if _, err := emuRead(p, 4); err != nil {
				return err
			}
		case 0x02:
			p.Write([]byte("DM32UV\x00\x00"))
		case 0x06:
			p.Write([]byte{0x06})
		case 'R', 'W':
			h, err := emuRead(p, 5)
			if err != nil {
				return err
			}
			addr := int(h[0]) | int(h[1])<<8 | int(h[2])<<16
			n := int(h[3]) | int(h[4])<<8
			off := addr - dm32EmuStart
			if off < 0 || off+n > len(e.flash) {
				return fmt.Errorf("indirizzo fuori dalla memoria %06X+%X", addr, n)
			}
			if c == 'R' {
				e.mu.Lock()
				d := append([]byte{}, e.flash[off:off+n]...)
				e.mu.Unlock()
				p.Write(append(append([]byte{'W'}, h...), d...))
				continue
			}
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			if off%dm32Page != 0 || n != dm32Page {
				return fmt.Errorf("scrittura non allineata %06X+%X", addr, n)
			}
			e.mu.Lock()
			copy(e.flash[off:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}
