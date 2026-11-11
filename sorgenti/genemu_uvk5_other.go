//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Quansheng UV-K5 simulata (protocollo di uvk5.py): saluto con la versione
// del firmware, lettura e scrittura della memoria a blocchi, riavvio.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type uvk5Emu struct {
	mu  sync.Mutex
	img []byte
	fw  string
}

func (uvk5Fam) emu(m *genModel, img []byte) genEmu {
	e := &uvk5Emu{img: append([]byte{}, img...), fw: "k5_2.01.32"}
	if m.ID == "qs-uvk1" {
		e.fw = "7.03.01"
	}
	if f := uvk5o(m).emuFW; f != "" {
		e.fw = f
	}
	if len(e.img) < uvk5MemSize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, uvk5MemSize-len(e.img))...)
	}
	e.img = e.img[:uvk5MemSize]
	return e
}

// 4.9: firmware liberi (egzumer, F4HWN, LOSEHU)
func (f uvk5xFam) emu(m *genModel, img []byte) genEmu {
	size := f.memSize()
	e := &uvk5Emu{img: append([]byte{}, img...), fw: uvk5o(m).emuFW}
	if len(e.img) < size {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, size-len(e.img))...)
	}
	e.img = e.img[:size]
	return e
}

func (e *uvk5Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *uvk5Emu) serve(p Port) error {
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		if h[0] != 0xAB || h[1] != 0xCD {
			return fmt.Errorf("pacchetto non valido")
		}
		body, err := emuRead(p, int(h[2])+2+2)
		if err != nil {
			return err
		}
		data := uvk5Xor(body[:int(h[2])+2])
		cmd, crc := data[:int(h[2])], data[int(h[2]):]
		if binary.LittleEndian.Uint16(crc) != crc16XModem(cmd) || body[len(body)-2] != 0xDC || body[len(body)-1] != 0xBA {
			return fmt.Errorf("CRC o chiusura del pacchetto non validi")
		}
		reply := func(b []byte) { p.Write(uvk5Packet(b)) }
		switch {
		case bytes.HasPrefix(cmd, []byte{0x14, 0x05}):
			r := append([]byte{0x15, 0x05, 0x24, 0x00}, []byte(e.fw)...)
			reply(append(r, make([]byte, 0x28-len(r))...))
		case bytes.HasPrefix(cmd, []byte{0x1b, 0x05}):
			off, n := int(cmd[4])|int(cmd[5])<<8, int(cmd[6])
			e.mu.Lock()
			if off+n > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", off)
			}
			d := append([]byte{}, e.img[off:off+n]...)
			e.mu.Unlock()
			reply(append([]byte{0x1c, 0x05, byte(n + 4), 0x00, cmd[4], cmd[5], byte(n), 0x00}, d...))
		case bytes.HasPrefix(cmd, []byte{0x1d, 0x05}):
			off, n := int(cmd[4])|int(cmd[5])<<8, int(cmd[6])
			if len(cmd) < 12+n {
				return fmt.Errorf("scrittura incompleta")
			}
			e.mu.Lock()
			copy(e.img[off:], cmd[12:12+n])
			e.mu.Unlock()
			reply([]byte{0x1e, 0x05, 0x02, 0x00, cmd[4], cmd[5]})
		case bytes.HasPrefix(cmd, []byte{0x2b, 0x05}) && len(cmd) >= 14:
			// 4.9: lettura a indirizzo lungo dei firmware LOSEHU
			off := int(binary.LittleEndian.Uint16(cmd[4:]))<<16 | int(binary.LittleEndian.Uint16(cmd[12:]))
			n := int(cmd[6])
			e.mu.Lock()
			if off+n > len(e.img) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%05X fuori dalla memoria", off)
			}
			d := append([]byte{}, e.img[off:off+n]...)
			e.mu.Unlock()
			reply(append([]byte{0x2c, 0x05, byte(n + 4), 0x00, cmd[12], cmd[13], byte(n), 0x00}, d...))
		case bytes.HasPrefix(cmd, []byte{0x38, 0x05}) && len(cmd) >= 14:
			off := int(binary.LittleEndian.Uint16(cmd[4:]))<<16 | int(binary.LittleEndian.Uint16(cmd[12:]))
			n := int(cmd[6]) - 2
			if n < 0 || len(cmd) < 14+n || off+n > len(e.img) {
				return fmt.Errorf("scrittura non valida")
			}
			e.mu.Lock()
			copy(e.img[off:], cmd[14:14+n])
			e.mu.Unlock()
			reply([]byte{0x1e, 0x05, 0x02, 0x00, cmd[12], cmd[13]})
		case bytes.HasPrefix(cmd, []byte{0xdd, 0x05}):
			return nil // riavvio
		default:
			return fmt.Errorf("comando sconosciuto % X", cmd[:2])
		}
	}
}

// 4.16: firmware UV-KX di joaquim.org (stesso protocollo, 8 KiB)
func (uvkxFam) emu(m *genModel, img []byte) genEmu {
	e := &uvk5Emu{img: append([]byte{}, img...), fw: uvk5o(m).emuFW}
	if len(e.img) < uvk5MemSize {
		e.img = append(e.img, make([]byte, uvk5MemSize-len(e.img))...)
	}
	e.img = e.img[:uvk5MemSize]
	return e
}

// 4.18: firmware R3b0rn di fagci. La radio simulata ha una EEPROM intera
// (impostazioni, VFO, preset, canali in fondo, correzione SSB): indirizzi a 32
// bit, scrittura a gruppi di 8 byte. image() restituisce quello che il programma
// legge (testa e canali).
type fagciEmu struct {
	mu    sync.Mutex
	eep   []byte
	typ   int
	old16 bool // firmware vecchio: non risponde ai comandi con indirizzi a 32 bit
	log   [][2]int
}

func (f fagciFam) emu(m *genModel, img []byte) genEmu {
	typ := 2
	if f.count(img) > 0 {
		typ = int(img[fagciHdr] & 7)
	} else {
		img = fagciBlank(typ)
	}
	e := &fagciEmu{typ: typ, eep: bytes.Repeat([]byte{0xA5}, fagciSizes[typ])}
	copy(e.eep, img[fagciHdr:fagciChanAt])
	end, n := fagciGeom(len(e.eep))
	copy(e.eep[end-n*fagciRec:end], img[fagciChanAt:])
	return e
}

func (e *fagciEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	img := fagciNew(e.eep[:fagciHead], e.typ)
	end, n := fagciGeom(len(e.eep))
	copy(img[fagciChanAt:], e.eep[end-n*fagciRec:end])
	return img
}

func (e *fagciEmu) serve(p Port) error {
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		if h[0] != 0xAB || h[1] != 0xCD {
			return fmt.Errorf("pacchetto non valido")
		}
		body, err := emuRead(p, int(h[2])+2+2)
		if err != nil {
			return err
		}
		data := uvk5Xor(body[:int(h[2])+2])
		cmd, crc := data[:int(h[2])], data[int(h[2]):]
		if binary.LittleEndian.Uint16(crc) != crc16XModem(cmd) || body[len(body)-2] != 0xDC || body[len(body)-1] != 0xBA {
			return fmt.Errorf("CRC o chiusura del pacchetto non validi")
		}
		reply := func(b []byte) { p.Write(uvk5Packet(b)) }
		ts := []byte{0x6a, 0x39, 0x57, 0x64}
		switch {
		case bytes.HasPrefix(cmd, []byte{0x14, 0x05}):
			r := append([]byte{0x15, 0x05, 0x24, 0x00}, []byte("R3b0rn")...)
			reply(append(r, make([]byte, 0x28-len(r))...))
		case bytes.HasPrefix(cmd, []byte{0x1b, 0x05}):
			// come il firmware: un comando con l'ora sbagliata resta senza risposta
			if e.old16 || len(cmd) < 16 || !bytes.Equal(cmd[12:16], ts) {
				continue
			}
			off, n := int(binary.LittleEndian.Uint32(cmd[4:])), int(cmd[8])
			e.mu.Lock()
			if off+n > len(e.eep) {
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%05X fuori dalla memoria", off)
			}
			d := append([]byte{}, e.eep[off:off+n]...)
			e.mu.Unlock()
			reply(append([]byte{0x1c, 0x05, byte(n + 8), 0x00, cmd[4], cmd[5], cmd[6], cmd[7], byte(n), 0, 0, 0}, d...))
		case bytes.HasPrefix(cmd, []byte{0x1d, 0x05}):
			if e.old16 || len(cmd) < 16 || !bytes.Equal(cmd[12:16], ts) {
				continue
			}
			off, n := int(binary.LittleEndian.Uint32(cmd[4:])), int(cmd[8])
			if len(cmd) < 16+n || off+n > len(e.eep) {
				return fmt.Errorf("scrittura non valida")
			}
			e.mu.Lock()
			copy(e.eep[off:], cmd[16:16+n/8*8]) // solo gruppi interi di 8 byte
			e.log = append(e.log, [2]int{off, n / 8 * 8})
			e.mu.Unlock()
			reply([]byte{0x1e, 0x05, 0x06, 0x00, cmd[4], cmd[5], cmd[6], cmd[7], 0, 0})
		case bytes.HasPrefix(cmd, []byte{0xdd, 0x05}):
			return nil // riavvio
		default:
			return fmt.Errorf("comando sconosciuto % X", cmd[:2])
		}
	}
}
