//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: tk11.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio simulata Quansheng TK11 (protocollo di tk11.py di CHIRP).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type tk11Emu struct {
	mu  sync.Mutex
	img []byte
}

func (tk11Fam) emu(m *genModel, img []byte) genEmu {
	e := &tk11Emu{img: append([]byte{}, img...)}
	if len(e.img) < tk11MemSize {
		e.img = append(e.img, make([]byte, tk11MemSize-len(e.img))...)
	}
	e.img = e.img[:tk11MemSize]
	return e
}

func (e *tk11Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

// risposta impacchettata e lunga n byte in tutto (dati completati con zeri)
func tk11Reply(data []byte, n int) []byte {
	data = append(data, make([]byte, n-8-len(data))...)
	return tk11Packet(data)
}

func (e *tk11Emu) serve(p Port) error {
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		if h[0] != 0xAB || h[1] != 0xCD {
			return fmt.Errorf("inizio del pacchetto atteso")
		}
		n := int(binary.LittleEndian.Uint16(h[2:]))
		body, err := emuRead(p, n+4)
		if err != nil {
			return err
		}
		d := uvk5Xor(body[:n+2])
		if !bytes.Equal(body[n+2:], []byte{0xDC, 0xBA}) || binary.LittleEndian.Uint16(d[n:]) != crc16XModem(d[:n]) {
			return fmt.Errorf("pacchetto non valido")
		}
		d = d[:n]
		switch binary.LittleEndian.Uint16(d) {
		case 0x01F4:
			rep := make([]byte, 20)
			binary.LittleEndian.PutUint16(rep, 0x01F5)
			binary.LittleEndian.PutUint16(rep[2:], 0x34)
			copy(rep[4:], "TK11 V1.0.42")
			p.Write(tk11Reply(rep, 64))
		case 0x01FB, 0x01FD:
			a := int(binary.LittleEndian.Uint32(d[4:])) - 0x080000
			l := int(binary.LittleEndian.Uint16(d[8:]))
			if a < 0 || a+l > tk11MemSize {
				return fmt.Errorf("indirizzo 0x%X fuori dalla memoria", a)
			}
			rep := append([]byte{}, d[:12]...)
			if binary.LittleEndian.Uint16(d) == 0x01FB {
				binary.LittleEndian.PutUint16(rep, 0x01FC)
				e.mu.Lock()
				rep = append(rep, e.img[a:a+l]...)
				e.mu.Unlock()
				p.Write(tk11Reply(rep, l+20))
			} else {
				e.mu.Lock()
				copy(e.img[a:], d[16:16+l])
				e.mu.Unlock()
				binary.LittleEndian.PutUint16(rep, 0x01FE)
				p.Write(tk11Reply(rep[:8], 16))
			}
		case 0x05DD:
			return nil
		default:
			return fmt.Errorf("comando sconosciuto % X", d[:2])
		}
	}
}
