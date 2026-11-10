//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.14: radio Icom simulate (copia con i pacchetti FE FE … FD); come il cavo a un
// filo, la radio simulata rimanda indietro ogni pacchetto ricevuto.

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

type icomEmu struct {
	mu     sync.Mutex
	o      *icomOpt
	img    []byte
	writes int  // blocchi scritti
	noEcho bool // cavo a due fili: nessun pacchetto torna indietro
	reject bool // la radio rifiuta i dati alla fine
	badSum int  // blocco (in lettura) con la somma di controllo sbagliata; 0 = nessuno
}

func newIcomEmu(o *icomOpt, img []byte) *icomEmu {
	mem := make([]byte, o.size)
	copy(mem, img)
	return &icomEmu{o: o, img: mem}
}

func (icv86Fam) emu(m *genModel, img []byte) genEmu { return newIcomEmu(icv86Opt, img) }
func (icv80Fam) emu(m *genModel, img []byte) genEmu { return newIcomEmu(icv80Opt, img) }
func (ict70Fam) emu(m *genModel, img []byte) genEmu { return newIcomEmu(ict70Opt, img) }
func (ict10Fam) emu(m *genModel, img []byte) genEmu { return newIcomEmu(ict10Opt, img) }
func (icx8xFam) emu(m *genModel, img []byte) genEmu { return newIcomEmu(icx8xClone, img) }

func (e *icomEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

func (e *icomEmu) serve(p Port) error {
	alen := 2
	if e.o.size >= 0x10000 {
		alen = 4
	}
	reply := func(cmd byte, payload []byte) { p.Write(icomFrame(icomRadio, icomPC, cmd, payload)) }
	writing, fast := false, false
	pending := map[int][]byte{}
	for {
		var f []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			f = append(f, b[0])
			if b[0] == 0xFD {
				break
			}
		}
		if !e.noEcho {
			p.Write(f)
		}
		for len(f) > 2 && f[2] == 0xFE {
			f = f[1:]
		}
		if len(f) == 1 && fast {
			continue // secondo FD del pacchetto della copia veloce
		}
		if len(f) < 6 || f[0] != 0xFE || f[1] != 0xFE || f[2] != icomPC || f[3] != icomRadio {
			return fmt.Errorf("pacchetto non valido % X", f)
		}
		cmd, pl := f[4], f[5:len(f)-1]
		// copia veloce: dopo E8 il modello arriva con l'ultimo byte a zero
		model := e.o.model[:]
		if fast {
			model = []byte{e.o.model[0], e.o.model[1], e.o.model[2], 0}
		}
		switch cmd {
		case icomFast:
			if !e.o.fast || !bytes.HasPrefix(pl, e.o.model[:]) {
				return fmt.Errorf("copia veloce non prevista")
			}
			fast = true
		case icomID:
			reply(icomModel, append(append([]byte{}, e.o.model[:]...), 1, 'I', 'U', '0', 'V', 'Y', 'T'))
		case icomOut:
			if !bytes.Equal(pl, model) || fast != e.o.fast {
				continue
			}
			src := e.image()
			blk := 0
			for a := 0; a < len(src); a += 32 {
				n := 32
				if a+n > len(src) {
					n = len(src) - a
				}
				var d []byte
				if alen == 4 {
					d = []byte{byte(a >> 24), byte(a >> 16), byte(a >> 8), byte(a), byte(n)}
				} else {
					d = []byte{byte(a >> 8), byte(a), byte(n)}
				}
				d = append(d, src[a:a+n]...)
				sum := icomSum(d)
				if blk++; blk == e.badSum {
					sum++
				}
				reply(icomDat, []byte(strings.ToUpper(hex.EncodeToString(append(d, sum)))))
			}
			reply(icomEnd, []byte(e.o.end))
		case icomIn:
			writing = bytes.Equal(pl, model) && fast == e.o.fast
			pending = map[int][]byte{}
		case icomDat:
			if !writing {
				continue
			}
			d := make([]byte, len(pl)/2)
			if _, err := hex.Decode(d, pl[:len(d)*2]); err != nil || len(d) < alen+2 {
				writing = false
				continue
			}
			a := 0
			for _, b := range d[:alen] {
				a = a<<8 | int(b)
			}
			n := int(d[alen])
			if len(d) != alen+1+n+1 || icomSum(d[:len(d)-1]) != d[len(d)-1] || a+n > len(e.img) {
				writing = false
				continue
			}
			pending[a] = d[alen+1 : alen+1+n]
		case icomEnd:
			if !writing || e.reject || string(pl) != e.o.end {
				reply(icomOK, []byte{1})
				writing = false
				continue
			}
			e.mu.Lock()
			for a, d := range pending {
				copy(e.img[a:], d)
				e.writes++
			}
			e.mu.Unlock()
			writing = false
			reply(icomOK, []byte{0})
		}
	}
}
