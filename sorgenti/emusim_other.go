//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulate per i test dei protocolli: implementano il lato radio così
// come risulta dal programma originale (stessi byte attesi e inviati).

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ---------------------------------------------------------------- collegamento simulato

type pipeEnd struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	other  *pipeEnd
	baud   int
	log    *bytes.Buffer // byte inviati da questo lato
	closed bool
}

func newPipe() (*pipeEnd, *pipeEnd) {
	a, b := &pipeEnd{baud: 0, log: &bytes.Buffer{}}, &pipeEnd{log: &bytes.Buffer{}}
	a.cond, b.cond = sync.NewCond(&a.mu), sync.NewCond(&b.mu)
	a.other, b.other = b, a
	return a, b
}

func (p *pipeEnd) Write(b []byte) error {
	p.log.Write(b)
	o := p.other
	o.mu.Lock()
	o.buf = append(o.buf, b...)
	o.mu.Unlock()
	o.cond.Broadcast()
	return nil
}

func (p *pipeEnd) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	timer := time.AfterFunc(timeout, func() { p.cond.Broadcast() })
	defer timer.Stop()
	stop := context.AfterFunc(ctx, func() { p.mu.Lock(); p.cond.Broadcast(); p.mu.Unlock() })
	defer stop()
	p.mu.Lock()
	defer p.mu.Unlock()
	for len(p.buf) < len(b) {
		if p.closed {
			return io.EOF
		}
		if ctx.Err() != nil {
			return errCancelled
		}
		if time.Now().After(deadline) {
			return errTimeout
		}
		p.cond.Wait()
	}
	copy(b, p.buf)
	p.buf = p.buf[len(b):]
	return nil
}

func (p *pipeEnd) Purge() error {
	p.mu.Lock()
	p.buf = nil
	p.mu.Unlock()
	return nil
}

func (p *pipeEnd) SetBaud(b int) error { p.baud = b; return nil }

func (p *pipeEnd) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	p.cond.Broadcast()
	o := p.other
	o.mu.Lock()
	o.closed = true
	o.mu.Unlock()
	o.cond.Broadcast()
	return nil
}

// lettura bloccante per i simulatori (30 s massimo)
func emuRead(p Port, n int) ([]byte, error) {
	b := make([]byte, n)
	err := p.ReadFull(context.Background(), b, 30*time.Second)
	return b, err
}

// ---------------------------------------------------------------- K67 simulato

type k67Emu struct {
	mem      []byte // 0xC000
	password string
	model    string
	band     byte
	writes   int // blocchi scritti
	reads    int
	answerA  byte // primo byte di risposta al passo 1 (normalmente 'A')
	rejectW  bool
	mu       sync.Mutex
	log      []string // operazioni in chiaro, per il confronto
}

func (e *k67Emu) note(s string) { e.mu.Lock(); e.log = append(e.log, s); e.mu.Unlock() }

var errEmuStop = errors.New("fine")

func (e *k67Emu) serve(p Port) error {
	for {
		if err := e.session(p); err != nil {
			return err
		}
	}
}

func (e *k67Emu) session(p Port) error {
	t, err := emuRead(p, 16)
	if err != nil {
		return err
	}
	if !bytes.Equal(t, k67TInfo) {
		return errors.New("passo 1 errato")
	}
	a := e.answerA
	if a == 0 {
		a = 'A'
	}
	p.Write([]byte{a})
	sync, err := emuRead(p, 8)
	if err != nil {
		return err
	}
	seed := sync[7]
	for i := 0; i < 7; i++ {
		if sync[i]^seed != k67Sync[i] {
			return errors.New("sync errato")
		}
	}
	p.Write([]byte{'A' ^ seed})
	pw, err := emuRead(p, 8)
	if err != nil {
		return err
	}
	want := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	copy(want, e.password)
	if !bytes.Equal(xorBytes(pw, seed), want) {
		p.Write([]byte{'N' ^ seed})
		return nil
	}
	p.Write([]byte{'A' ^ seed})
	inf, err := emuRead(p, 11)
	if err != nil {
		return err
	}
	if !bytes.Equal(xorBytes(inf, seed), k67Header) {
		return errors.New("INFORMATION errato")
	}
	id := bytes.Repeat([]byte{0xFF}, 16)
	copy(id, e.model)
	id[8] = e.band
	p.Write(xorBytes(id, seed))
	op, err := emuRead(p, 1)
	if err != nil {
		return err
	}
	if o := op[0] ^ seed; o != 'R' && o != 'W' {
		return errors.New("modo errato")
	} else {
		e.note("HS:" + string(o) + ":" + string(bytes.TrimRight(xorBytes(pw, seed), "\xff")))
	}
	p.Write([]byte{'A' ^ seed})
	for {
		c, err := emuRead(p, 4)
		if err != nil {
			return err
		}
		d := xorBytes(c, seed)
		switch {
		case bytes.Equal(d, k67End):
			e.note("END")
			p.Write([]byte{'A' ^ seed})
			return nil
		case d[0] == 'R':
			addr := int(d[1])<<8 | int(d[2])
			resp := append(append([]byte{}, d...), e.mem[addr:addr+4096]...)
			e.reads++
			e.note(fmt.Sprintf("R:%04X:%x", addr, d[3]))
			p.Write(xorBytes(resp, seed))
		case d[0] == 'W':
			addr := int(d[1])<<8 | int(d[2])
			data, err := emuRead(p, 4096)
			if err != nil {
				return err
			}
			if e.rejectW {
				p.Write([]byte{'N' ^ seed})
				continue
			}
			copy(e.mem[addr:], xorBytes(data, seed))
			e.note(fmt.Sprintf("W:%04X:%x", addr, sha1.Sum(xorBytes(data, seed))))
			e.writes++
			p.Write([]byte{'A' ^ seed})
		default:
			return errors.New("comando sconosciuto")
		}
	}
}

// ---------------------------------------------------------------- Bajeton simulata

type btEmu struct {
	mem     []byte // 0xC340 (compreso il blocco finale)
	id      []byte
	written map[int]bool
	ended   bool
	mu      sync.Mutex
	log     []string
}

func (e *btEmu) note(s string) { e.mu.Lock(); e.log = append(e.log, s); e.mu.Unlock() }

func newBtEmu(img []byte) *btEmu {
	m := bytes.Repeat([]byte{0xFF}, btImgSize+64)
	copy(m, img)
	return &btEmu{mem: m, id: []byte("BT-11U  V1.00   "), written: map[int]bool{}}
}

func (e *btEmu) serve(p Port) error {
	for {
		if err := e.session(p); err != nil {
			return err
		}
	}
}

func (e *btEmu) session(p Port) error {
	h, err := emuRead(p, 12)
	if err != nil {
		return err
	}
	if string(h) != "PROGRAMBT11U" {
		return errors.New("handshake errato")
	}
	p.Write([]byte{6})
	f, err := emuRead(p, 1)
	if err != nil {
		return err
	}
	if f[0] != 'F' {
		return errors.New("F atteso")
	}
	p.Write(e.id)
	pk, err := emuRead(p, 25)
	if err != nil {
		return err
	}
	if string(pk[:4]) != "SEND" {
		return errors.New("SEND atteso")
	}
	idx := 2*int(pk[4]&0x0F) + 1
	if pk[4]&0x20 != 0 {
		idx++
	}
	key := []byte(btSymbols[pk[4+idx]])
	e.note("HS")
	p.Write([]byte{6})
	for {
		c, err := emuRead(p, 1)
		if err != nil {
			return err
		}
		switch c[0] {
		case 'E':
			e.note("E")
			e.ended = true
			return nil
		case 'R':
			hdr, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr := int(hdr[0])<<8 | int(hdr[1])
			data := append([]byte{}, e.mem[addr:addr+64]...)
			btCrypt(data, key)
			e.note(fmt.Sprintf("R:%04X:%x", addr, hdr[2]))
			p.Write(append([]byte{'W', hdr[0], hdr[1], hdr[2]}, data...))
		case 'W':
			hdr, err := emuRead(p, 67)
			if err != nil {
				return err
			}
			addr := int(hdr[0])<<8 | int(hdr[1])
			data := append([]byte{}, hdr[3:]...)
			btCrypt(data, key)
			copy(e.mem[addr:], data)
			e.note(fmt.Sprintf("W:%04X:%x:%x", addr, hdr[2], data))
			e.written[addr] = true
			p.Write([]byte{6})
		default:
			return errors.New("comando sconosciuto")
		}
	}
}

// ---------------------------------------------------------------- simulatore su TCP (per il confronto con il codice originale)

type tcpPort struct{ c net.Conn }

func (t *tcpPort) Write(b []byte) error { _, err := t.c.Write(b); return err }
func (t *tcpPort) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	t.c.SetReadDeadline(time.Now().Add(to))
	_, err := io.ReadFull(t.c, b)
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return errTimeout
	}
	return err
}
func (t *tcpPort) Purge() error      { return nil }
func (t *tcpPort) SetBaud(int) error { return nil }
func (t *tcpPort) Close() error      { return t.c.Close() }

// serveTCP: simulatore raggiungibile via TCP (una sessione per connessione)
func serveTCP(serve func(Port) error) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go serve(&tcpPort{c: c})
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}
