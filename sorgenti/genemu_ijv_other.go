//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.13: radio simulata con il firmware IJV (V3, V4, FW5, V6). Accetta i comandi
// offuscati (protocollo di fabbrica) e, per V4 e FW5, quelli in chiaro, e
// risponde nello stesso modo. La FW5 ha la memoria fisica con 16 banchi; V4 e
// FW5 espongono il banco scelto (byte 0x000B o «CFG5»).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type ijvEmu struct {
	mu       sync.Mutex
	kind     ijvKind
	fw       string
	img      []byte // V3, V4: memoria logica
	flash    []byte // FW5: memoria fisica
	noPlain  bool   // V4 vecchia: solo protocollo di fabbrica
	refuse   bool   // FW5: scrittura dei banchi non abilitata
	v6       bool   // risponde solo al collegamento dati della V6
	beacon   bool   // FW5: annuncio 0x0570 prima della risposta al saluto
	writes   int
	flushes  int
	noReply  map[uint16]bool
	lastMode string
}

const ijvEmuFlash = ijvBankBase + 16*ijvBankSize

func (f ijvFam) emu(m *genModel, img []byte) genEmu {
	// versioni di prova nel formato dei firmware IJV
	e := &ijvEmu{kind: f.kind, fw: [...]string{"V-N3.60", "V-X3.60", "IJV V4.0.4", "IJV FW5 5.2"}[f.kind]}
	size := f.imgSize()
	im := append([]byte{}, img...)
	if len(im) < size {
		im = append(im, bytes.Repeat([]byte{0xFF}, size-len(im))...)
	}
	im = im[:size]
	if f.kind == ijvFW5 {
		e.flash = bytes.Repeat([]byte{0xFF}, ijvEmuFlash)
		copy(e.flash, im[:ijvFixedEnd])
		bk, err := f.imgBank(im)
		if err != nil {
			bk = 0
		}
		copy(e.flash[ijvBankBase+bk*ijvBankSize:], im[ijvFixedEnd:])
	} else {
		e.img = im
	}
	return e
}

func (e *ijvEmu) bank() int {
	if e.kind == ijvFW5 {
		bk, err := ijvCfgBank(e.flash[ijvCfg5 : ijvCfg5+32])
		if err != nil {
			return 0
		}
		return bk
	}
	return int(e.img[0x0B])
}

func (e *ijvEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.kind != ijvFW5 {
		return append([]byte{}, e.img...)
	}
	bk := e.bank()
	out := append([]byte{}, e.flash[:ijvFixedEnd]...)
	return append(out, e.flash[ijvBankBase+bk*ijvBankSize:ijvBankBase+(bk+1)*ijvBankSize]...)
}

func (e *ijvEmu) serve(p Port) error {
	if e.v6 {
		return e.serveV6(p)
	}
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		if h[0] != 0xAB || h[1] != 0xCD {
			return fmt.Errorf("pacchetto non valido % X", h)
		}
		n := int(h[2])
		body, err := emuRead(p, n+4)
		if err != nil {
			return err
		}
		if body[n+2] != 0xDC || body[n+3] != 0xBA {
			return fmt.Errorf("chiusura del pacchetto non valida")
		}
		var cmd []byte
		plain := false
		if d := uvk5Xor(body[:n+2]); binary.LittleEndian.Uint16(d[n:]) == crc16XModem(d[:n]) {
			cmd = d[:n]
		} else if binary.LittleEndian.Uint16(body[n:]) == crc16XModem(body[:n]) && n%4 == 0 {
			cmd, plain = body[:n], true
		} else {
			return fmt.Errorf("CRC del pacchetto non valido")
		}
		if plain && (e.noPlain || e.kind == ijvV3 || e.kind == ijvV3X) {
			continue // la radio non capisce il protocollo in chiaro: nessuna risposta
		}
		if !plain && e.kind == ijvFW5 {
			continue
		}
		e.lastMode = map[bool]string{true: "chiaro", false: "fabbrica"}[plain]
		reply := func(b []byte) {
			if plain {
				pad := append(append([]byte{}, b...), make([]byte, (4-len(b)%4)%4)...)
				out := []byte{0xAB, 0xCD, byte(len(pad)), byte(len(pad) >> 8)}
				out = append(out, pad...)
				crc := crc16XModem(pad)
				p.Write(append(out, byte(crc), byte(crc>>8), 0xDC, 0xBA))
				return
			}
			p.Write(uvk5Packet(b))
		}
		id := binary.LittleEndian.Uint16(cmd)
		if e.noReply[id] {
			continue
		}
		switch id {
		case 0x0514:
			if e.beacon {
				reply([]byte{0x70, 0x05, 0x00, 0x00})
			}
			r := append([]byte{0x15, 0x05, 0x24, 0x00}, []byte(e.fw)...)
			reply(append(r, make([]byte, 0x28-len(r))...))
		case 0x051B:
			off, ln := int(binary.LittleEndian.Uint16(cmd[4:])), int(cmd[6])
			e.mu.Lock()
			if e.kind == ijvFW5 || off+ln > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("lettura logica non valida a 0x%04X", off)
			}
			d := append([]byte{}, e.img[off:off+ln]...)
			e.mu.Unlock()
			reply(append([]byte{0x1c, 0x05, byte(ln + 4), 0x00, cmd[4], cmd[5], byte(ln), 0x00}, d...))
		case 0x051D:
			off, ln := int(binary.LittleEndian.Uint16(cmd[4:])), int(cmd[6])
			e.mu.Lock()
			if e.kind == ijvFW5 || len(cmd) < 12+ln || off+ln > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("scrittura logica non valida a 0x%04X", off)
			}
			if off < ijvSettingEnd+0x200 && off >= ijvSettingEnd {
				e.mu.Unlock()
				return fmt.Errorf("scrittura nelle tarature a 0x%04X", off)
			}
			copy(e.img[off:], cmd[12:12+ln])
			e.writes++
			e.mu.Unlock()
			reply([]byte{0x1e, 0x05, 0x02, 0x00, cmd[4], cmd[5]})
		case 0x0564, 0x0566:
			if e.kind != ijvFW5 {
				return fmt.Errorf("comando fisico con il firmware %s", e.fw)
			}
			addr, ln := int(binary.LittleEndian.Uint32(cmd[4:])), int(cmd[8])
			e.mu.Lock()
			if ln == 0 || ln > 128 || addr+ln > len(e.flash) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo fisico non valido 0x%06X", addr)
			}
			r := make([]byte, 10)
			binary.LittleEndian.PutUint16(r, id+1)
			copy(r[4:], cmd[4:9])
			if id == 0x0564 {
				binary.LittleEndian.PutUint16(r[2:], uint16(6+ln))
				r = append(r, e.flash[addr:addr+ln]...)
			} else {
				binary.LittleEndian.PutUint16(r[2:], 6)
				switch {
				case e.refuse && addr >= ijvBankBase:
					r[9] = 2
				case addr >= 0x1540 && addr < ijvFixedEnd:
					e.mu.Unlock()
					return fmt.Errorf("scrittura nella zona protetta a 0x%06X", addr)
				default:
					copy(e.flash[addr:], cmd[10:10+ln])
					e.writes++
				}
			}
			e.mu.Unlock()
			reply(r)
		case 0x0576:
			e.flushes++
			reply([]byte{0x77, 0x05, 0x01, 0x00, 0x00})
		case 0x05DD:
			return nil
		default:
			return fmt.Errorf("comando sconosciuto 0x%04X", id)
		}
	}
}

// serveV6: risponde solo al collegamento dati della V6 (informazioni e versione)
func (e *ijvEmu) serveV6(p Port) error {
	for {
		// si cerca l'inizio «A5 5A»: i pacchetti della UV-K5 non hanno risposta
		b, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		if b[0] != 0xA5 {
			continue
		}
		if b, err = emuRead(p, 1); err != nil {
			return nil
		} else if b[0] != 0x5A {
			continue
		}
		t, err := emuRead(p, 8)
		if err != nil {
			return nil
		}
		h := append([]byte{0xA5, 0x5A}, t...)
		n := int(binary.LittleEndian.Uint16(h[8:]))
		rest, err := emuRead(p, n+4)
		if err != nil {
			return err
		}
		all := append(append([]byte{}, h...), rest[:n]...)
		if crc32IEEE(all[2:]) != binary.LittleEndian.Uint32(rest[n:]) {
			return fmt.Errorf("CRC V6 non valido")
		}
		var v uint32
		switch h[2] {
		case 8:
			v = 1<<28 | 1<<16 | 27<<8 | 1 // UV-K5 V3, servizi 27
		case 18:
			v = 6<<16 | 1<<8 | 2 // 6.1.2
		default:
			return fmt.Errorf("comando V6 sconosciuto %d", h[2])
		}
		r := []byte{0x5A, 0xA5, 0, h[3], 0, 0, 0, 0}
		binary.LittleEndian.PutUint32(r[4:], v)
		p.Write(binary.LittleEndian.AppendUint32(r, crc32IEEE(r[2:8])))
	}
}

// ijvSynthImage: memoria iniziale di prova (vuota, banco 3 scelto, «CFG5»
// valida per la FW5)
func ijvSynthImage(m *genModel) []byte {
	f := m.Fam.(ijvFam)
	img := bytes.Repeat([]byte{0xFF}, f.imgSize())
	for i := 0; i < f.slots(); i++ {
		ijvBlankRec(f.rec(img, i))
	}
	for k := 0; k < 0x40; k++ {
		img[k] = 0
	}
	switch f.kind {
	case ijvV4:
		img[0x0B] = 2
	case ijvFW5:
		binary.LittleEndian.PutUint32(img[ijvCfg5:], ijvCfg5Magic)
		for k := 4; k < 32; k++ {
			img[ijvCfg5+k] = 0
		}
		img[ijvCfg5Bank] = 2
	}
	return img
}
