//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: Icom IC-91AD / IC-92AD simulata (vedi drv_ic9x.go). Risponde al
// risveglio, poi a lettura, scrittura e cancellazione di una posizione delle
// due bande; rifiuta (FA) i canali con campi non validi, come la radio vera
// secondo le note di CHIRP.

import (
	"sync"
	"time"
)

type ic9xEmu struct {
	mu     sync.Mutex
	m      *genModel
	mem    map[int][]byte // banda<<16 | tipo<<12 | numero → canale
	writes int            // canali scritti
	clears int            // posizioni cancellate
	reads  int
	wakes  int
	deaf   bool          // non risponde al risveglio
	echo   bool          // cavo che rimanda indietro i byte inviati
	nakAt  int           // rifiuta la scrittura numero nakAt (1 = la prima)
	stray  bool          // prima di ogni risposta manda un blocco non richiesto
	sleepy time.Duration // dopo questo tempo senza comandi serve di nuovo il risveglio (0 = mai)
	mute   int           // non risponde più dopo questo numero di letture (0 = mai)
	wrong  bool          // risponde alla lettura con un'altra posizione
	awake  bool
	last   time.Time
}

func ic9xKey(band, kind byte, num int) int { return int(band)<<16 | int(kind)<<12 | num }

func (ic9xFam) emu(m *genModel, img []byte) genEmu {
	e := &ic9xEmu{m: m, mem: map[int][]byte{}}
	f := ic9xFam{}
	if img != nil && f.Match(m, img) {
		for i := 0; i < m.Slots; i++ {
			if p := f.payload(m, img, i); p != nil {
				kind, num := ic9xAddr(m, i)
				e.mem[ic9xKey(ic9xo(m).band, kind, num)] = append([]byte{}, p...)
			}
		}
	}
	return e
}

func (e *ic9xEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	f := ic9xFam{}
	img := ic9xBlank(e.m)
	for i := 0; i < e.m.Slots; i++ {
		kind, num := ic9xAddr(e.m, i)
		if p := e.mem[ic9xKey(ic9xo(e.m).band, kind, num)]; p != nil {
			r := f.rec(img, i)
			r[0], r[1] = 1, byte(len(p))
			copy(r[2:], p)
		}
	}
	return img
}

func (e *ic9xEmu) reply(p Port, band byte, body ...byte) {
	p.Write(append(append([]byte{0xFE, 0xFE, 0x80, band}, body...), 0xFD))
}

func (e *ic9xEmu) serve(p Port) error {
	for {
		var fr []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			fr = append(fr, b[0])
			if b[0] == 0xFD {
				break
			}
		}
		if e.echo {
			p.Write(fr)
		}
		k := 0
		for k < len(fr) && fr[k] == 0xFE {
			k++
		}
		if k < 2 {
			continue
		}
		body := fr[k : len(fr)-1]
		e.mu.Lock()
		e.handle(p, body)
		e.mu.Unlock()
	}
}

func (e *ic9xEmu) handle(p Port, body []byte) {
	if len(body) == 3 && body[0] == 0x01 && body[1] == 0x80 && body[2] == 0x19 {
		if e.deaf {
			return
		}
		e.wakes++
		e.awake, e.last = true, time.Now()
		e.reply(p, 0x01, 0x19, 0x00)
		return
	}
	if !e.awake || (e.sleepy > 0 && time.Since(e.last) > e.sleepy) {
		e.awake = false
		return
	}
	if len(body) < 7 || (body[0] != 1 && body[0] != 2) || body[1] != 0x80 || body[2] != 0x1A || body[3] != 0x00 {
		return
	}
	e.last = time.Now()
	band, data := body[0], body[4:]
	num, ok := bcdBE(data[1:3])
	limit := 850
	if band == 2 {
		limit = 450
	}
	if data[0] == 2 {
		limit = 2
	}
	if !ok || (data[0] != 1 && data[0] != 2) || num >= limit {
		e.reply(p, band, 0xFA)
		return
	}
	key := ic9xKey(band, data[0], num)
	if e.stray {
		e.reply(p, band, 0x1A, 0x00, 0x09, 0x00, 0x00)
	}
	switch {
	case len(data) == 3:
		e.reads++
		if e.mute > 0 && e.reads > e.mute {
			return
		}
		if e.wrong {
			e.reply(p, band, 0x1A, 0x00, data[0], 0x09, 0x99, 0xFF)
			return
		}
		if ch := e.mem[key]; ch != nil {
			e.reply(p, band, append([]byte{0x1A, 0x00}, ch...)...)
		} else {
			e.reply(p, band, 0x1A, 0x00, data[0], data[1], data[2], 0xFF)
		}
	case len(data) == 4 && data[3] == 0xFF:
		e.clears++
		delete(e.mem, key)
		e.reply(p, band, 0xFB)
	default:
		if e.nakAt > 0 && e.writes+1 == e.nakAt {
			e.nakAt = -1
			e.reply(p, band, 0xFA)
			return
		}
		name := true
		if len(data) >= 34 {
			for _, c := range data[26:34] {
				name = name && c >= 0x20 && c < 0x7F
			}
		}
		if !ic9xValid(data, band) || !name {
			e.reply(p, band, 0xFA)
			return
		}
		e.writes++
		e.mem[key] = append([]byte{}, data...)
		e.reply(p, band, 0xFB)
	}
}
