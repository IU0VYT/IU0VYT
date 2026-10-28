//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata DM-1701 in modalità DFU (protocollo in drv_tyt.go): stati DFU,
// comandi del caricatore (91 01, A2 01, 21 indirizzo, 41 cancellazione, 91 05)
// e memoria flash (la scrittura può solo azzerare bit: un blocco non cancellato
// prima di scriverlo resta sbagliato). Stesso formato del collegamento TCP usato
// per far parlare dmrconfig con la radio simulata (verifica/dmrconfig/tcpdev.c).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type tytEmu struct {
	mu       sync.Mutex
	m        *genModel
	mem      []byte // memoria come l'immagine (0xD0000)
	state    byte
	addr     uint32
	cmd      []byte // ultimo DNLOAD a valore 0 (comando)
	ident    bool
	rebooted bool
}

func (tytFam) emu(m *genModel, img []byte) genEmu {
	mem := bytes.Repeat([]byte{0xFF}, tytMemSize)
	copy(mem, img)
	return &tytEmu{m: m, mem: mem, state: dfuStIdle}
}

func (e *tytEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.mem...)
}

// indirizzo fisico (blocchi da 1 KB) -> posizione nell'immagine
func tytEmuOff(phys int) int {
	switch {
	case phys < 256:
		return phys * 1024
	case phys >= 1088 && phys < 1088+(tytMemSize/1024-256):
		return (phys - 832) * 1024
	}
	return -1
}

func (e *tytEmu) serve(p Port) error {
	ident := make([]byte, 64)
	copy(ident, e.m.Opt.(*tytOpt).ident)
	reply := func(res int32, data []byte) error {
		b := make([]byte, 4, 4+len(data))
		binary.LittleEndian.PutUint32(b, uint32(res))
		return p.Write(append(b, data...))
	}
	for {
		var h [8]byte
		if err := emuReadInto(p, h[:]); err != nil {
			return nil
		}
		rt, rq := h[0], h[1]
		value := int(binary.LittleEndian.Uint16(h[2:]))
		n := int(binary.LittleEndian.Uint16(h[6:]))
		in := rt&0x80 != 0
		var out []byte
		if !in && n > 0 {
			out = make([]byte, n)
			if err := emuReadInto(p, out); err != nil {
				return err
			}
		}
		e.mu.Lock()
		var err error
		switch {
		case in && rq == dfuGetState:
			err = reply(1, []byte{e.state})
		case in && rq == dfuGetStatus:
			// dopo un DNLOAD la radio passa a «dnload idle», dopo un UPLOAD resta
			// in «upload idle»: wait_dfu_idle li riporta a «idle» con ABORT
			if e.state == 3 {
				e.state = 5
			}
			err = reply(6, []byte{0, 0, 0, 0, e.state, 0})
		case !in && (rq == dfuAbort || rq == dfuClrStatus):
			e.state = dfuStIdle
			err = reply(0, nil)
		case !in && rq == dfuDnload && value == 0:
			e.state = 3
			switch {
			case len(out) == 2 && out[0] == 0x91 && out[1] == 0x01:
			case len(out) == 2 && out[0] == 0xA2 && out[1] == 0x01:
				e.ident = true
			case len(out) == 2 && out[0] == 0x91 && out[1] == 0x05:
				e.rebooted = true
			case len(out) == 5 && out[0] == 0x21:
				e.addr = binary.LittleEndian.Uint32(out[1:])
			case len(out) == 5 && out[0] == 0x41:
				a := binary.LittleEndian.Uint32(out[1:])
				off := -1
				switch {
				case a < 0x40000:
					off = int(a)
				case a >= 0x110000 && a < 0x1E0000:
					off = int(a) - 0x110000 + 0x40000
				}
				if off < 0 || a%0x10000 != 0 {
					e.state = dfuStError
				} else {
					for k := off; k < off+0x10000 && k < tytMemSize; k++ {
						e.mem[k] = 0xFF
					}
				}
			default:
				e.state = dfuStError
			}
			err = reply(int32(len(out)), nil)
		case !in && rq == dfuDnload:
			off := tytEmuOff(value - 2)
			if e.addr != 0 || off < 0 || len(out) != 1024 {
				e.state = dfuStError
				err = reply(-9, nil)
				break
			}
			for k, b := range out {
				e.mem[off+k] &= b
			}
			e.state = 3
			err = reply(1024, nil)
		case in && rq == dfuUpload && value == 0:
			if !e.ident || n != 64 {
				err = reply(-9, nil)
				break
			}
			e.state = 9
			err = reply(64, ident)
		case in && rq == dfuUpload:
			off := tytEmuOff(value - 2)
			if e.addr != 0 || off < 0 || n != 1024 {
				err = reply(-9, nil)
				break
			}
			e.state = 9
			err = reply(1024, append([]byte{}, e.mem[off:off+1024]...))
		case !in && rq == dfuDetach:
			err = reply(0, nil)
		default:
			err = fmt.Errorf("richiesta DFU inattesa %02X %02X", rt, rq)
		}
		e.mu.Unlock()
		if err != nil {
			return err
		}
	}
}

// tytSynthImage: memoria vuota (canali, zone e liste cancellati come fa dmrconfig)
func tytSynthImage() []byte {
	img := bytes.Repeat([]byte{0xFF}, tytMemSize)
	for i := 0; i < tytNCh; i++ {
		tytErase(tytRec(img, i))
	}
	for z := 0; z < tytNZones; z++ {
		zr, ze := tytZone(img, z), tytZoneExt(img, z)
		for k := range zr {
			zr[k] = 0
		}
		for k := range ze {
			ze[k] = 0
		}
	}
	for s := 0; s < tytNScan; s++ {
		sl := tytScan(img, s)
		for k := range sl {
			sl[k] = 0
		}
		binary.LittleEndian.PutUint16(sl[32:], 0xFFFF)
		binary.LittleEndian.PutUint16(sl[34:], 0xFFFF)
		binary.LittleEndian.PutUint16(sl[36:], 0xFFFF)
		sl[38], sl[39], sl[40], sl[41] = 0xF1, 500/25, 2000/250, 0xFF
	}
	return img
}
