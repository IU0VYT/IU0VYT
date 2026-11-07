//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.9: TYT UVL-15W simulata (protocollo pubblicato da TYT, tyt_uvl15w.py)

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type uvl15wEmu struct {
	mu  sync.Mutex
	img []byte
}

func (uvl15wFam) emu(m *genModel, img []byte) genEmu {
	e := &uvl15wEmu{img: append([]byte{}, img...)}
	if len(e.img) < uvl15wSize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, uvl15wSize-len(e.img))...)
	}
	e.img = e.img[:uvl15wSize]
	return e
}

func (e *uvl15wEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *uvl15wEmu) reply(p Port, cmd byte, data []byte) {
	out := []byte{0xFE, 0xFE, 0xEF, 0xEE, cmd}
	out = append(out, uvl15wEscape(append(append([]byte{}, data...), uvl15wLRC(data)))...)
	p.Write(append(out, 0xFD))
}

func (e *uvl15wEmu) serve(p Port) error {
	for {
		var fr []byte
		for {
			b, err := emuByte(p)
			if err != nil {
				return nil
			}
			if fr == nil && b != 0xFE {
				continue
			}
			fr = append(fr, b)
			if b == 0xFD {
				break
			}
		}
		fr = trimLeftByte(fr, 0xFE)
		if len(fr) < 5 || fr[0] != 0xEE || fr[1] != 0xEF {
			return fmt.Errorf("pacchetto non valido % X", fr)
		}
		pl, err := uvl15wUnescape(fr[3 : len(fr)-1])
		if err != nil || sum8(pl) != 0 {
			return fmt.Errorf("somma errata")
		}
		cmd, d := fr[2], pl[:len(pl)-1]
		switch cmd {
		case uvl15wCmdInfo:
			info := append([]byte("UVL-15W"), make([]byte, 13)...)
			copy(info[12:], "3.09.21")
			e.reply(p, uvl15wCmdInfoResp, info)
		case uvl15wCmdReadStart:
			e.reply(p, cmd, []byte("READ START OK"))
		case uvl15wCmdWriteStart:
			e.reply(p, cmd, []byte("WRITE START OK"))
		case uvl15wCmdRead:
			a, n := int(binary.BigEndian.Uint32(d))-uvl15wFlashStart, int(binary.BigEndian.Uint16(d[4:]))
			e.mu.Lock()
			blk := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			e.reply(p, uvl15wCmdWrite, append(append([]byte{}, d[:6]...), blk...))
		case uvl15wCmdWrite:
			a, n := int(binary.BigEndian.Uint32(d))-uvl15wFlashStart, int(binary.BigEndian.Uint16(d[4:]))
			e.mu.Lock()
			copy(e.img[a:a+n], d[6:])
			e.mu.Unlock()
			e.reply(p, cmd, append([]byte("WF OK"), d[:6]...))
		case uvl15wCmdFinish:
			e.reply(p, cmd, []byte("OK"))
			return nil
		default:
			return fmt.Errorf("comando sconosciuto %02X", cmd)
		}
	}
}
