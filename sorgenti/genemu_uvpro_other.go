//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: BTECH UV-PRO simulata (vedi drv_uvpro.go): informazioni, stato, nome
// del gruppo, lettura e scrittura di un canale del gruppo attivo.

import (
	"encoding/binary"
	"sync"
)

type uvproEmu struct {
	mu      sync.Mutex
	m       *genModel
	info    []byte   // 9 byte
	group   int      // gruppo attivo
	gname   string   // nome del gruppo
	recs    [][]byte // dati dei canali (25 byte; altra lunghezza = canale di un altro tipo)
	missing map[int]bool
	writes  int
	reads   int
	sent    [][]byte // canali ricevuti con il comando di scrittura
	noState bool     // non dà lo stato (comando 20)
	stray   bool     // prima di ogni risposta manda un blocco non richiesto, con il byte di controllo
	refuse  int      // rifiuta la scrittura numero refuse (1 = la prima)
	flipAt  int      // alla scrittura numero flipAt salva un bit diverso
	stuck   bool     // con flipAt: anche la scrittura successiva (il ripristino) non riesce
	ackOff  bool     // conferma la scrittura con un altro numero di canale
	mute    int      // non risponde più dopo questo numero di letture (0 = mai)
	garbage bool     // risponde con dati che non sono della radio
}

func (uvproFam) emu(m *genModel, img []byte) genEmu {
	f := uvproFam{}
	e := &uvproEmu{m: m, info: []byte{0x00, 0x00, 0x01, 0x02, 0x00, uvproSoftVer, 0x00, 0x60, 30}, group: 2, gname: "GRUPPO 3", missing: map[int]bool{}}
	for i := 0; i < m.Slots; i++ {
		e.recs = append(e.recs, uvproEmpty(i))
	}
	if img != nil && f.Match(m, img) {
		copy(e.info, img[uvproInfoAt:uvproInfoAt+9])
		if g := img[uvproGroupAt]; g != 0xFF {
			e.group = int(g)
		}
		for i := 0; i < m.Slots; i++ {
			switch r := f.rec(img, i); r[0] {
			case uvproOK:
				e.recs[i] = append([]byte{}, r[1:1+uvproLen]...)
			case uvproNone:
				e.missing[i] = true
			default:
				e.recs[i] = make([]byte, uvproLen+8)
			}
		}
	}
	return e
}

func (e *uvproEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	f := uvproFam{}
	img := make([]byte, uvproSize(e.m))
	copy(img, uvproMagic)
	copy(img[uvproInfoAt:], e.info)
	img[uvproGroupAt] = 0xFF
	if !e.noState {
		img[uvproGroupAt] = byte(e.group)
		copy(img[uvproNameAt:uvproNameAt+10], e.gname)
	}
	for i := 0; i < e.m.Slots; i++ {
		r := f.rec(img, i)
		switch {
		case e.missing[i]:
			r[0] = uvproNone
		case len(e.recs[i]) == uvproLen:
			r[0] = uvproOK
			copy(r[1:], e.recs[i])
		default:
			r[0] = uvproForeign
		}
	}
	return img
}

func (e *uvproEmu) reply(p Port, cmd uint16, body ...byte) {
	if e.stray {
		// un avviso della radio (altro comando), con il byte di controllo in coda
		p.Write([]byte{0xFF, 0x01, 0x01, 0x02, 0x00, 0x02, 0x00, 0x09, 0x01, 0x02, 0xAA})
	}
	out := []byte{0xFF, 0x01, 0x00, byte(len(body)), 0x00, 0x02, byte(cmd>>8) | 0x80, byte(cmd)}
	p.Write(append(out, body...))
}

func (e *uvproEmu) serve(p Port) error {
	for {
		h, err := emuRead(p, 8)
		if err != nil {
			return nil
		}
		if h[0] != 0xFF || h[1] != 0x01 {
			return nil
		}
		body, err := emuRead(p, int(h[3])+int(h[2]&1))
		if err != nil {
			return nil
		}
		body = body[:int(h[3])]
		cmd := binary.BigEndian.Uint16(h[6:8])
		e.mu.Lock()
		e.handle(p, cmd, body)
		e.mu.Unlock()
	}
}

func (e *uvproEmu) handle(p Port, cmd uint16, body []byte) {
	if e.garbage {
		p.Write([]byte("AT+OK\r\n\x00"))
		return
	}
	switch cmd {
	case 4:
		e.reply(p, cmd, append([]byte{0}, e.info...)...)
	case 20:
		if e.noState {
			e.reply(p, cmd, 5)
			return
		}
		e.reply(p, cmd, 0, 0x80, 0x00, byte(e.group>>2), byte(e.group<<6))
	case 73:
		if len(body) != 1 {
			e.reply(p, cmd, 5)
			return
		}
		name := make([]byte, 10)
		copy(name, e.gname)
		e.reply(p, cmd, append([]byte{0, body[0]}, name...)...)
	case 13:
		e.reads++
		if e.mute > 0 && e.reads > e.mute {
			return
		}
		if len(body) != 1 || int(body[0]) >= len(e.recs) || e.missing[int(body[0])] {
			e.reply(p, cmd, 5)
			return
		}
		e.reply(p, cmd, append([]byte{0}, e.recs[body[0]]...)...)
	case 14:
		e.sent = append(e.sent, append([]byte{}, body...))
		if len(body) != uvproLen || int(body[0]) >= len(e.recs) {
			e.reply(p, cmd, 5)
			return
		}
		e.writes++
		if e.refuse > 0 && e.writes == e.refuse || e.stuck && e.flipAt > 0 && e.writes > e.flipAt {
			e.reply(p, cmd, 1)
			return
		}
		rec := append([]byte{}, body...)
		if e.flipAt > 0 && e.writes == e.flipAt {
			rec[13] ^= 0x20
		}
		e.recs[body[0]] = rec
		ack := body[0]
		if e.ackOff {
			ack++
		}
		e.reply(p, cmd, 0, ack)
	}
}
