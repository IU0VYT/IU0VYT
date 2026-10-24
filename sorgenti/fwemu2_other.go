//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulate per i caricatori KDH più recenti (vedi fwkdhx.go): UV-25 Plus
// (.kdhx), Bajeton BJ-9000 (.Firmware) e UV-9G (.kdh), come risultano dai
// programmi originali. Simulano anche la velocità della porta: i byte inviati a
// una velocità diversa da quella della radio arrivano rovinati e sono ignorati.

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"sync"
)

type kdh2Emu struct {
	mu         sync.Mutex
	v          *kdh2Variant
	raw        []byte // file atteso (nil: prove dell'interfaccia)
	state      string // app | upd | pre | boot | done
	radioBaud  int
	log        []string
	errs       []string
	data       []byte
	next       int
	announced  []byte
	greeted    bool
	nPkt       int
	drop       map[int]bool // pacchetti (contati da 1) senza risposta
	nak        map[int]bool // risposta E2
	corrupt    map[int]bool // risposta con CRC errato
	reject     byte         // codice d'errore al primo pacchetto di dati
	wrongModel bool         // la radio è di un altro modello (E6, o nessuna risposta alla presentazione)
	hsDrop     int
	silent     bool
	cur        Port
	curBaud    int
	noisy      bool // byte rovinati già annotati per questa connessione
}

func newKdh2Emu(v *kdh2Variant, raw []byte, manual bool) *kdh2Emu {
	e := &kdh2Emu{v: v, raw: raw, state: "app", radioBaud: v.HelloBaud, drop: map[int]bool{}, nak: map[int]bool{}, corrupt: map[int]bool{}}
	if manual {
		e.state, e.radioBaud = "boot", v.BootBaud
	}
	return e
}

func (e *kdh2Emu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }
func (e *kdh2Emu) fail(f string, a ...interface{}) { e.errs = append(e.errs, fmt.Sprintf(f, a...)) }

func (e *kdh2Emu) snapshot() (log, errs []string, state string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state
}

func (e *kdh2Emu) reply(cmd, code byte, badCRC bool) {
	var b []byte
	if !e.v.V3 {
		b = kdhPack(cmd, code, nil)
	} else {
		b = e.v.pack(cmd, uint16(code), nil)
	}
	if badCRC {
		b[len(b)-2] ^= 0x55
	}
	if e.curBaud != e.radioBaud {
		b = bytes.Repeat([]byte{0x80}, len(b)/2+1)
	}
	_ = e.cur.Write(b)
}

func (e *kdh2Emu) errReply(cmd, code byte) {
	if e.v.V3 {
		e.reply(cmd, code, false)
	} else {
		e.reply(code, 0, false)
	}
}

// garbled: true se la connessione non è alla velocità della radio
func (e *kdh2Emu) garbled(baud int) bool {
	if baud == e.radioBaud {
		return false
	}
	if !e.noisy {
		e.noisy = true
		e.note("%d: byte rovinati", baud)
	}
	return true
}

func (e *kdh2Emu) serve(p Port, baud int) error {
	e.mu.Lock()
	e.cur, e.curBaud, e.noisy = p, baud, false
	e.mu.Unlock()
	var win []byte
	for {
		b, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		e.mu.Lock()
		if e.garbled(baud) {
			e.mu.Unlock()
			continue
		}
		st := e.state
		if st == "app" || st == "upd" {
			win = append(win, b[0])
			if len(win) > 32 {
				win = win[len(win)-32:]
			}
			switch {
			case st == "app" && bytes.HasSuffix(win, e.v.Hello):
				win = nil
				e.note("%d: presentazione", baud)
				if e.hsDrop > 0 || e.silent || e.wrongModel && !e.v.V3 {
					e.hsDrop--
				} else {
					e.state = "upd"
					_ = p.Write([]byte{6})
				}
			case st == "upd" && bytes.HasSuffix(win, []byte("UPDATE")):
				win = nil
				e.note("%d: UPDATE", baud)
				e.state = "pre"
				_ = p.Write([]byte{6})
				e.radioBaud = e.v.BootBaud
			}
			e.mu.Unlock()
			continue
		}
		e.mu.Unlock()
		if b[0] != 0xAA {
			continue
		}
		pkt, ok, err := e.readPacket(p)
		if err != nil {
			return nil
		}
		e.mu.Lock()
		hl := e.v.hdrLen()
		if !ok {
			e.note("crc-errato")
			e.errReply(0, 0xE2)
			e.mu.Unlock()
			continue
		}
		e.nPkt++
		k := e.nPkt
		n := len(pkt) - 1 - hl - 3
		d := pkt[1+hl : 1+hl+n]
		arg := int(pkt[2])
		if e.v.V3 {
			arg = int(pkt[2])<<8 | int(pkt[3])
		}
		e.note("%d: %02X %d %d %08x", baud, pkt[1], arg, n, crc32.ChecksumIEEE(d))
		switch {
		case e.silent || e.drop[k]:
		case e.nak[k]:
			e.errReply(pkt[1], 0xE2)
		case e.corrupt[k]:
			e.reply(pkt[1], 6, true)
		default:
			e.handle(pkt[1], arg, d)
		}
		e.mu.Unlock()
	}
}

// readPacket (dopo AA): ok=false se il CRC è errato
func (e *kdh2Emu) readPacket(p Port) ([]byte, bool, error) {
	hl := e.v.hdrLen()
	h, err := emuRead(p, hl)
	if err != nil {
		return nil, false, err
	}
	n := int(h[hl-2])<<8 | int(h[hl-1])
	rest, err := emuRead(p, n+3)
	if err != nil {
		return nil, false, err
	}
	pkt := append(append([]byte{0xAA}, h...), rest...)
	c := crc16x(pkt[1 : 1+hl+n])
	tail := byte(0xEF)
	if e.v.V3 {
		tail = 0x55
	}
	return pkt, c == uint16(rest[n])<<8|uint16(rest[n+1]) && rest[n+2] == tail, nil
}

func (e *kdh2Emu) handle(cmd byte, arg int, data []byte) {
	if e.state == "pre" {
		if cmd != 0x42 {
			e.fail("comando %02X prima dell'avvio del caricatore", cmd)
			e.errReply(cmd, 0xE5)
			return
		}
		e.state = "boot"
		e.reply(cmd, 6, false)
		return
	}
	if e.state != "boot" {
		e.fail("comando %02X nello stato %s", cmd, e.state)
		return
	}
	switch cmd {
	case e.v.GreetCmd:
		if string(data) != e.v.Greet {
			e.errReply(cmd, 0xE1)
			return
		}
		e.greeted = true
	case 0x02:
		if !e.v.V3 || !e.greeted {
			e.errReply(cmd, 0xE5)
			return
		}
		if e.wrongModel || (e.raw != nil && !bytes.Equal(data, e.raw[e.v.ModelAt:e.v.ModelAt+32])) {
			e.errReply(cmd, 0xE6)
			return
		}
	case 0x04:
		e.announced, e.next, e.data = append([]byte{}, data...), 0, nil
	case 0x03:
		if e.reject != 0 {
			e.errReply(cmd, e.reject)
			return
		}
		if len(data) != kdhPkt {
			e.fail("pacchetto di %d byte", len(data))
		}
		switch {
		case arg == e.next&e.argMask():
			e.data = append(e.data, data...)
			e.next++
		case arg == (e.next-1)&e.argMask():
			// ripetizione dopo una conferma persa
		default:
			e.errReply(cmd, 0xE3)
			return
		}
	case 0x45:
		e.state = "done"
		e.check()
	default:
		e.errReply(cmd, 0xE5)
		return
	}
	e.reply(cmd, 6, false)
}

func (e *kdh2Emu) argMask() int {
	if e.v.V3 {
		return 0xFFFF
	}
	return 0xFF
}

func (e *kdh2Emu) check() {
	if e.raw == nil {
		return
	}
	n := (len(e.raw) + kdhPkt - 1) / kdhPkt
	if want := e.v.countData(n); !bytes.Equal(e.announced, want) {
		e.fail("numero di pacchetti annunciato %x invece di %x", e.announced, want)
	}
	if e.next != n {
		e.fail("ricevuti %d pacchetti invece di %d", e.next, n)
	}
	exp := append([]byte{}, e.raw...)
	for len(exp)%kdhPkt != 0 {
		exp = append(exp, kdh2Fill)
	}
	if !bytes.Equal(e.data, exp) {
		e.fail("dati diversi dal file")
	}
}

func kdh2EmuPorts(e *kdh2Emu) func(string, int) (Port, error) {
	return func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b, baud)
		return a, nil
	}
}

// ---------------------------------------------------------------- UV-9G

type k9gEmu struct {
	mu        sync.Mutex
	raw       []byte
	state     string // app | u | hs | data | done
	radioBaud int
	log       []string
	errs      []string
	data      []byte
	next      int
	nFr       int
	drop      map[int]bool // blocchi (contati da 1) senza risposta
	nak       map[int]bool // risposta E1
	reject    byte         // risposta a tutti i blocchi di dati
	hsDrop    int
	silent    bool
	curBaud   int
	noisy     bool
}

func newK9gEmu(raw []byte, manual bool) *k9gEmu {
	e := &k9gEmu{raw: raw, state: "app", radioBaud: 9600, drop: map[int]bool{}, nak: map[int]bool{}}
	if manual {
		e.state, e.radioBaud = "hs", k9gBaud
	}
	return e
}

func (e *k9gEmu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }
func (e *k9gEmu) fail(f string, a ...interface{}) { e.errs = append(e.errs, fmt.Sprintf(f, a...)) }

func (e *k9gEmu) snapshot() (log, errs []string, state string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state
}

func (e *k9gEmu) serve(p Port, baud int) error {
	e.mu.Lock()
	e.curBaud, e.noisy = baud, false
	e.mu.Unlock()
	var win []byte
	for {
		b, err := emuRead(p, 1)
		if err != nil {
			return nil
		}
		e.mu.Lock()
		if baud != e.radioBaud {
			if !e.noisy {
				e.noisy = true
				e.note("%d: byte rovinati", baud)
			}
			e.mu.Unlock()
			continue
		}
		switch e.state {
		case "app":
			win = append(win, b[0])
			if len(win) > 16 {
				win = win[len(win)-16:]
			}
			if bytes.HasSuffix(win, k9gHello) {
				win = nil
				e.note("%d: presentazione", baud)
				if e.hsDrop > 0 || e.silent {
					e.hsDrop--
				} else {
					e.state = "u"
					_ = p.Write([]byte{6})
				}
			}
		case "u":
			if b[0] == 'U' {
				e.note("%d: U", baud)
				e.state, e.radioBaud = "hs", k9gBaud
			}
		case "hs":
			win = append(win, b[0])
			if len(win) > 16 {
				win = win[len(win)-16:]
			}
			if bytes.HasSuffix(win, k9gHs) {
				win = nil
				e.note("%d: caricatore", baud)
				if e.hsDrop > 0 || e.silent {
					e.hsDrop--
				} else {
					e.state = "data"
					_ = p.Write([]byte{6})
				}
			}
		case "data":
			if b[0] == 'E' {
				e.note("%d: E", baud)
				e.state = "done"
				if e.raw != nil && len(e.data) < len(e.raw) {
					e.state = "interrotto" // fine dopo un errore
				} else {
					e.check()
				}
				break
			}
			if b[0] != 0xAA {
				e.note("%d: byte %02X inatteso", baud, b[0])
				break
			}
			e.mu.Unlock()
			rest, err := emuRead(p, 37)
			if err != nil {
				return nil
			}
			e.mu.Lock()
			fr := append([]byte{0xAA}, rest...)
			e.nFr++
			k := e.nFr
			addr := int(fr[1])<<8 | int(fr[2])
			d := fr[3:35]
			e.note("%d: %04X %08x", baud, addr, crc32.ChecksumIEEE(d))
			c := crc16x(fr[1:35])
			switch {
			case e.silent || e.drop[k]:
			case c != uint16(fr[35])<<8|uint16(fr[36]) || fr[37] != 0xEF:
				_ = p.Write([]byte{0xE1})
			case e.nak[k]:
				_ = p.Write([]byte{0xE1})
			case e.reject != 0:
				_ = p.Write([]byte{e.reject})
			case addr == e.next*k9gBlock:
				e.data = append(e.data, d...)
				e.next++
				_ = p.Write([]byte{6})
			case addr == (e.next-1)*k9gBlock:
				_ = p.Write([]byte{6})
			default:
				_ = p.Write([]byte{0xE2})
			}
		case "done":
			e.note("%d: byte %02X dopo la fine", baud, b[0])
		}
		e.mu.Unlock()
	}
}

func (e *k9gEmu) check() {
	if e.raw == nil {
		return
	}
	exp := append([]byte{}, e.raw...)
	for len(exp)%k9gBlock != 0 {
		exp = append(exp, 0)
	}
	if !bytes.Equal(e.data, exp) {
		e.fail("dati diversi dal file (%d byte ricevuti, %d attesi)", len(e.data), len(exp))
	}
}

func k9gEmuPorts(e *k9gEmu) func(string, int) (Port, error) {
	return func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b, baud)
		return a, nil
	}
}
