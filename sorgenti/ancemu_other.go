//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bajeton 8300 simulata per l'aggiornamento del firmware: lato radio del
// protocollo del programma "BinTe ANC Bootloader" (vedi ancproto.go). Serve ai
// test e al confronto con il codice originale. Il comportamento della radio
// vera non è documentato: il simulatore segue ciò che il programma originale
// si aspetta (risposte 01 ai cambi di velocità, richieste 02, fine con 04 e 03)
// e simula una UART: i byte inviati a una velocità diversa da quella della
// radio arrivano rovinati.

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

type ancReq struct{ addr, n int64 }

type ancEmu struct {
	mu          sync.Mutex
	fw          []byte // contenuto atteso (firmware decifrato)
	state       string // app | upd | boot | loader | done
	radioBaud   int
	loaderHello bool // il caricatore si presenta con 01 dopo il cambio di velocità
	plan        []ancReq
	next        int
	got         int64
	log         []string // pacchetti ricevuti dal PC: "baud:hex" (o "baud:rovinati")
	errs        []string
	drop        map[int]bool // pacchetti del PC (contati da 1) ignorati
	corrupt     map[int]bool // pacchetti del PC a cui si risponde con CRC errato
	nPkt        int
	cur         Port
	curBaud     int
	conns       int
	silent      bool // radio spenta: non risponde mai
	unknownCmd  bool // risponde una volta con un comando sconosciuto
	delay       time.Duration
	hello       *time.Ticker
	reboot      string // stato in cui ripartire dopo un aggiornamento (prove dell'interfaccia)
	anyBaud     bool   // collegamento senza informazioni sulla velocità (pseudo-terminale): byte sempre validi
}

// idle: la radio non è in aggiornamento (risponde alla programmazione dei canali)
func (e *ancEmu) idle() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state == "app" || e.state == "done"
}

// newAncEmu: radio accesa (state "app") o già in modalità aggiornamento ("boot").
func newAncEmu(fw []byte, recovery bool) *ancEmu {
	e := &ancEmu{fw: fw, state: "app", radioBaud: ancBaudStart, drop: map[int]bool{}, corrupt: map[int]bool{}, delay: 15 * time.Millisecond}
	if recovery {
		e.state = "boot"
	}
	return e
}

// piano di richieste predefinito: intestazione, poi blocchi da 4 KB della parte
// iniziale, una richiesta a cavallo della fine del file
func (e *ancEmu) defaultPlan(limit int64) {
	e.planFor(int64(len(e.fw)), limit)
}

// planFor: piano per un firmware lungo n byte (anche senza conoscerne il contenuto)
func (e *ancEmu) planFor(n, limit int64) {
	if limit <= 0 || limit > n {
		limit = n
	}
	e.plan = []ancReq{{0, 512}, {512, 1536}}
	for a := int64(2048); a < limit; a += 4096 {
		l := int64(4096)
		e.plan = append(e.plan, ancReq{a, l})
	}
	e.plan = append(e.plan, ancReq{n - 100, 600})
}

func (e *ancEmu) note(f string, a ...interface{}) { e.log = append(e.log, fmt.Sprintf(f, a...)) }
func (e *ancEmu) fail(f string, a ...interface{}) { e.errs = append(e.errs, fmt.Sprintf(f, a...)) }

func (e *ancEmu) snapshot() (log, errs []string, state string, got int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string{}, e.log...), append([]string{}, e.errs...), e.state, e.got
}

// sendLater: risposta della radio dopo d, sulla connessione aperta in quel momento
func (e *ancEmu) sendLater(d time.Duration, payload []byte, badCRC bool) {
	time.AfterFunc(d, func() {
		e.mu.Lock()
		p, ok := e.cur, e.curBaud == e.radioBaud || e.anyBaud
		e.mu.Unlock()
		if p == nil {
			return
		}
		b := ancPack(payload)
		if badCRC {
			b[len(b)-1] ^= 0x5A
		}
		if !ok {
			b = bytes.Repeat([]byte{0x80}, len(b)/2+1) // velocità diversa: byte rovinati
		}
		_ = p.Write(b)
	})
}

func (e *ancEmu) request() {
	if e.next < len(e.plan) {
		r := e.plan[e.next]
		e.sendLater(e.delay, []byte{2, byte(r.addr), byte(r.addr >> 8), byte(r.addr >> 16), byte(r.addr >> 24),
			byte(r.n), byte(r.n >> 8), byte(r.n >> 16), byte(r.n >> 24)}, false)
		return
	}
	e.sendLater(e.delay, []byte{4}, false)
}

// serve: una connessione (apertura della porta) a una certa velocità
func (e *ancEmu) serve(p Port, hostBaud int) error {
	e.mu.Lock()
	if e.state == "done" && e.reboot != "" && hostBaud == ancBaudStart {
		// prove dell'interfaccia: la radio è stata riaccesa
		e.state, e.radioBaud, e.next, e.got, e.hello = e.reboot, ancBaudStart, 0, 0, nil
	}
	e.cur, e.curBaud = p, hostBaud
	e.conns++
	if e.state == "boot" && e.hello == nil && !e.silent {
		// caricatore in attesa: si presenta con 01 a 9600 baud finché il PC non risponde
		e.hello = time.NewTicker(700 * time.Millisecond)
		go func(t *time.Ticker) {
			for range t.C {
				e.mu.Lock()
				st := e.state
				e.mu.Unlock()
				if st != "boot" {
					t.Stop()
					return
				}
				e.sendLater(0, []byte{1}, false)
			}
		}(e.hello)
	}
	e.mu.Unlock()
	for {
		pkt, err := e.readPacket(p)
		if err != nil {
			return err
		}
		e.mu.Lock()
		if hostBaud != e.radioBaud && !e.anyBaud {
			e.note("%d:rovinati", hostBaud)
			e.mu.Unlock()
			continue
		}
		if pkt == nil {
			e.note("%d:crc-errato", hostBaud)
			e.mu.Unlock()
			continue
		}
		e.nPkt++
		e.note("%d:%x", hostBaud, ancPack(pkt))
		k := e.nPkt
		if e.silent || e.drop[k] {
			e.mu.Unlock()
			continue
		}
		if e.corrupt[k] {
			e.sendLater(e.delay, []byte{1}, true)
			e.mu.Unlock()
			continue
		}
		e.handle(pkt)
		e.mu.Unlock()
	}
}

func (e *ancEmu) readPacket(p Port) ([]byte, error) {
	for {
		b, err := emuRead(p, 1)
		if err != nil {
			return nil, err
		}
		if b[0] == 0xAA {
			break
		}
	}
	h, err := emuRead(p, 3)
	if err != nil {
		return nil, err
	}
	n := int(h[1]) | int(h[2])<<8
	rest, err := emuRead(p, n+2)
	if err != nil {
		return nil, err
	}
	frame := append([]byte{0xAA}, h...)
	frame = append(frame, rest[:n]...)
	if h[0] != 0x55 || n == 0 || crc16x(frame) != uint16(rest[n])|uint16(rest[n+1])<<8 {
		return nil, nil
	}
	return rest[:n], nil
}

func (e *ancEmu) handle(pkt []byte) {
	cmd := pkt[0]
	switch e.state {
	case "app":
		if cmd == 6 {
			e.state = "upd"
			e.sendLater(e.delay, []byte{1}, false)
		}
	case "upd", "boot":
		if cmd != 1 || len(pkt) < 5 {
			e.fail("stato %s: comando %d inatteso", e.state, cmd)
			return
		}
		b := int(le32(pkt[1:5]))
		switch {
		case b == ancBaudMid && e.radioBaud != ancBaudMid && e.state == "upd":
			e.radioBaud = ancBaudMid
			e.sendLater(60*time.Millisecond, []byte{1}, false)
		case b == ancBaudMid && e.state == "upd":
			e.sendLater(e.delay, []byte{1}, false)
		case b == 115200 || b == 57600:
			e.state = "loader"
			e.radioBaud = b
			if e.loaderHello {
				e.sendLater(80*time.Millisecond, []byte{1}, false)
			} else {
				e.next = 0
				time.AfterFunc(80*time.Millisecond, func() { e.mu.Lock(); e.request(); e.mu.Unlock() })
			}
		default:
			e.fail("velocità %d inattesa", b)
		}
	case "loader":
		switch cmd {
		case 1:
			if e.unknownCmd {
				e.unknownCmd = false
				e.sendLater(e.delay, []byte{9, 9}, false)
				return
			}
			e.request()
		case 2:
			if e.next >= len(e.plan) {
				e.fail("dati non richiesti")
				return
			}
			r := e.plan[e.next]
			if len(pkt) != int(9+r.n) {
				e.fail("richiesta %d: lunghezza %d invece di %d", e.next, len(pkt), 9+r.n)
			}
			addr, cnt := le32(pkt[1:5]), le32(pkt[5:9])
			want := r.n
			if e.fw == nil { // contenuto sconosciuto (prove dell'interfaccia)
				want = cnt
				if cnt < 0 || cnt > r.n {
					e.fail("richiesta %d: byte validi %d", e.next, cnt)
					return
				}
			} else if r.addr+want > int64(len(e.fw)) {
				want = int64(len(e.fw)) - r.addr
			}
			if addr != r.addr || cnt != want {
				e.fail("richiesta %d: indirizzo %d/%d byte validi %d/%d", e.next, addr, r.addr, cnt, want)
			}
			if e.fw != nil && len(pkt) >= int(9+cnt) && cnt >= 0 && !bytes.Equal(pkt[9:9+cnt], e.fw[r.addr:r.addr+cnt]) {
				e.fail("richiesta %d: dati diversi dal firmware", e.next)
			}
			for _, z := range pkt[min(len(pkt), int(9+cnt)):] {
				if z != 0 {
					e.fail("richiesta %d: riempimento non a zero", e.next)
					break
				}
			}
			e.got += cnt
			e.next++
			e.request()
		case 3:
			e.state = "done"
			e.sendLater(e.delay, []byte{3}, false)
		default:
			e.fail("caricatore: comando %d inatteso", cmd)
		}
	case "done":
		e.fail("pacchetto dopo la fine")
	}
}

// ancEmuPorts: openPortFunc che collega ogni apertura della porta al simulatore
func ancEmuPorts(e *ancEmu) func(string, int) (Port, error) {
	return func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b, baud)
		return a, nil
	}
}
