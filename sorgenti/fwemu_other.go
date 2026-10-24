//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulate per l'aggiornamento del firmware Baofeng: caricatore KDH V1.0
// (file .BF) e 5RH Pro (file .dat), come risultano dai programmi originali.

import (
	"bytes"
	"context"
	"fmt"
	"time"
	"hash/crc32"
	"sync"
)

// ---------------------------------------------------------------- KDH V1.0 (.BF)

type kdhEmu struct {
	mu      sync.Mutex
	raw     []byte // file atteso
	state   string // app | upd | pre | boot | done
	log     []string
	errs    []string
	r1, r2  []byte
	region  int
	next    int
	n1, n2  int
	nPkt    int
	drop    map[int]bool // pacchetti (contati da 1) senza risposta
	nak     map[int]bool // risposta E2 (errore di controllo)
	corrupt map[int]bool // risposta con CRC errato
	reject  byte         // codice d'errore da dare al primo pacchetto di dati (prove)
	hsDrop  int          // presentazioni ignorate
	silent  bool
	cur     Port
}

func newKdhEmu(raw []byte, manual bool) *kdhEmu {
	e := &kdhEmu{raw: raw, state: "app", drop: map[int]bool{}, nak: map[int]bool{}, corrupt: map[int]bool{}}
	if manual {
		e.state = "boot"
	}
	return e
}

func (e *kdhEmu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }
func (e *kdhEmu) fail(f string, a ...interface{}) { e.errs = append(e.errs, fmt.Sprintf(f, a...)) }

func (e *kdhEmu) snapshot() (log, errs []string, state string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state
}

func (e *kdhEmu) reply(cmd, arg byte, badCRC bool) {
	b := kdhPack(cmd, arg, nil)
	if badCRC {
		b[6] ^= 0x55
	}
	_ = e.cur.Write(b)
}

func (e *kdhEmu) serve(p Port, baud int) error {
	e.mu.Lock()
	e.cur = p
	if baud != kdhBaud {
		e.fail("velocità %d", baud)
	}
	e.mu.Unlock()
	var win []byte
	for {
		e.mu.Lock()
		st := e.state
		e.mu.Unlock()
		if st == "app" || st == "upd" {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			win = append(win, b[0])
			if len(win) > 32 {
				win = win[len(win)-32:]
			}
			e.mu.Lock()
			switch {
			case st == "app" && bytes.HasSuffix(win, []byte(kdhHandshke)):
				win = nil
				e.note("presentazione")
				if e.hsDrop > 0 || e.silent {
					e.hsDrop--
				} else {
					e.state = "upd"
					_ = p.Write([]byte{6})
				}
			case st == "upd" && bytes.HasSuffix(win, []byte("UPDATE")):
				win = nil
				e.note("UPDATE")
				e.state = "pre"
				_ = p.Write([]byte{6})
			}
			e.mu.Unlock()
			continue
		}
		pkt, ok, err := e.readPacket(p)
		if err != nil {
			return nil
		}
		e.mu.Lock()
		if !ok {
			e.note("crc-errato")
			e.reply(0xE2, 0, false)
			e.mu.Unlock()
			continue
		}
		e.nPkt++
		k := e.nPkt
		n := len(pkt) - 8
		e.note("%02X %02X %d %08x", pkt[1], pkt[2], n, crc32.ChecksumIEEE(pkt[5:5+n]))
		switch {
		case e.silent || e.drop[k]:
		case e.nak[k]:
			e.reply(0xE2, 0, false)
		case e.corrupt[k]:
			e.reply(pkt[1], 6, true)
		default:
			e.handle(pkt[1], pkt[2], pkt[5:5+n])
		}
		e.mu.Unlock()
	}
}

// readPacket: ok=false se il CRC è errato
func (e *kdhEmu) readPacket(p Port) ([]byte, bool, error) {
	for {
		b, err := emuRead(p, 1)
		if err != nil {
			return nil, false, err
		}
		if b[0] == 0xAA {
			break
		}
	}
	h, err := emuRead(p, 4)
	if err != nil {
		return nil, false, err
	}
	n := int(h[2])<<8 | int(h[3])
	rest, err := emuRead(p, n+3)
	if err != nil {
		return nil, false, err
	}
	pkt := append(append([]byte{0xAA}, h...), rest...)
	c := crc16x(pkt[1 : 5+n])
	return pkt, c == uint16(rest[n])<<8|uint16(rest[n+1]) && rest[n+2] == 0xEF, nil
}

func (e *kdhEmu) handle(cmd, arg byte, data []byte) {
	if e.state == "pre" {
		if cmd != 0x42 {
			e.fail("comando %02X prima dell'avvio del caricatore", cmd)
			e.reply(cmd, 0xE5, false)
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
	case 0x01:
		if string(data) != "BOOTLOADER" {
			e.reply(0xE1, 0, false)
			return
		}
	case 0x04, 0x05:
		if len(data) != 1 {
			e.fail("comando %02X senza numero di pacchetti", cmd)
		} else if cmd == 4 {
			e.region, e.next, e.n1 = 1, 0, int(data[0])
		} else {
			e.region, e.next, e.n2 = 2, 0, int(data[0])
		}
	case 0x03:
		if e.reject != 0 {
			e.reply(e.reject, 0, false)
			return
		}
		if len(data) != kdhPkt {
			e.fail("pacchetto di %d byte", len(data))
		}
		switch {
		case int(arg) == e.next&0xFF:
			if e.region == 1 {
				e.r1 = append(e.r1, data...)
			} else {
				e.r2 = append(e.r2, data...)
			}
			e.next++
		case int(arg) == (e.next-1)&0xFF:
			// ripetizione dopo una conferma persa
		default:
			e.reply(0xE3, 0, false)
			return
		}
	case 0x45:
		e.state = "done"
		e.check()
	default:
		e.reply(0xE5, 0, false)
		return
	}
	e.reply(cmd, 6, false)
}

func (e *kdhEmu) check() {
	if e.raw == nil {
		return // prove dell'interfaccia: file non noto
	}
	s1, s2 := int(be32(e.raw[1:5])), int(be32(e.raw[5:9]))
	if len(e.r1)/kdhPkt != e.n1 || len(e.r2)/kdhPkt != e.n2 {
		e.fail("pacchetti ricevuti %d+%d, annunciati %d+%d", len(e.r1)/kdhPkt, len(e.r2)/kdhPkt, e.n1, e.n2)
	}
	if len(e.r1) < s1 || !bytes.Equal(e.r1[:s1], e.raw[16:16+s1]) {
		e.fail("area 1 diversa dal file")
	}
	if len(e.r2) < s2 || !bytes.Equal(e.r2[:s2], e.raw[16+s1:16+s1+s2]) {
		e.fail("area 2 diversa dal file")
	}
}

func kdhEmuPorts(e *kdhEmu) func(string, int) (Port, error) {
	return func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b, baud)
		return a, nil
	}
}

// ---------------------------------------------------------------- 5RH Pro (.dat)

type rh5Emu struct {
	mu    sync.Mutex
	raw   []byte
	v2    bool // elettronica della radio
	state string
	log   []string
	errs  []string
	nak   int // blocco (contato da 1) a cui si risponde 'N'
	mute  int // blocco a cui non si risponde
}

func newRh5Emu(raw []byte, v2 bool) *rh5Emu { return &rh5Emu{raw: raw, v2: v2, state: "idle"} }

func (e *rh5Emu) note(f string, a ...interface{}) {
	e.mu.Lock()
	e.log = append(e.log, fmt.Sprintf(f, a...))
	e.mu.Unlock()
}
func (e *rh5Emu) fail(f string, a ...interface{}) {
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf(f, a...))
	e.mu.Unlock()
}

func (e *rh5Emu) snapshot() (log, errs []string, state string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state
}

func (e *rh5Emu) setState(s string) { e.mu.Lock(); e.state = s; e.mu.Unlock() }

func (e *rh5Emu) serve(p Port, baud int) error {
	e.note("porta %d", baud)
	expect := func(want []byte, what string) bool {
		b, err := emuRead(p, len(want))
		if err != nil {
			return false
		}
		e.note("%s %x", what, b)
		if !bytes.Equal(b, want) {
			e.fail("%s: ricevuto %x", what, b)
			return false
		}
		return true
	}
	if !expect([]byte("DOWNLOAD"), "download") {
		return nil
	}
	if e.v2 {
		p.Write(rh5V2)
	} else {
		p.Write(rh5V1)
	}
	if !expect([]byte{'A'}, "ack") {
		return nil
	}
	p.Write([]byte{'A'})
	if !expect(rh5Erase, "erase") {
		return nil
	}
	e.setState("cancellata")
	p.Write([]byte{'A'})
	if !expect([]byte("PROGRAM1"), "program") {
		return nil
	}
	p.Write([]byte{'A'})
	size := rh5Size(e.raw)
	var sum uint32
	var got int64
	for k := 1; ; k++ {
		h, err := emuRead(p, 5)
		if err != nil {
			return nil
		}
		if string(h[:3]) == "END" {
			t, err := emuRead(p, 4)
			if err != nil {
				return nil
			}
			cs := be32(t)
			e.note("end %x %x", h, t)
			if uint32(cs) != sum || got != size {
				e.fail("somma %08x invece di %08x, %d byte invece di %d", cs, sum, got, size)
				p.Write([]byte{'N'})
				return nil
			}
			e.setState("aggiornata")
			p.Write([]byte{'A'})
			return nil
		}
		a := be32(h[:4])
		var d []byte
		var n int64
		if e.raw == nil {
			// prove dell'interfaccia: file non noto, blocco fino a una pausa
			d = readUpTo(context.Background(), p, rh5Block, 200*time.Millisecond)
			n, size = int64(len(d)), got+int64(len(d))
		} else {
			n = size - a
			if n > rh5Block {
				n = rh5Block
			}
			if a != got || h[4] != 0 || n <= 0 {
				e.fail("blocco %d: intestazione %x inattesa", k, h)
				return nil
			}
			if d, err = emuRead(p, int(n)); err != nil {
				return nil
			}
		}
		e.note("blocco %x %d %08x", h, n, crc32.ChecksumIEEE(d))
		if e.raw != nil && !bytes.Equal(d, e.raw[rh5Head+a:rh5Head+a+n]) {
			e.fail("blocco %d: dati diversi dal file", k)
		}
		for _, v := range d {
			sum += uint32(v)
		}
		got += n
		switch k {
		case e.mute:
			continue
		case e.nak:
			p.Write([]byte{'N'})
		default:
			p.Write([]byte{'A'})
		}
	}
}

func rh5EmuPorts(e *rh5Emu) func(string, int) (Port, error) {
	return func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b, baud)
		return a, nil
	}
}
