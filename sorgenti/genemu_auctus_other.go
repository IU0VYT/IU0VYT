//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Retevis RB39 / RB639 (drv_auctus.go): pacchetti AA … BB a 2400
// baud; tiene l'elenco dei comandi ricevuti (comando<<8 | sottocomando).

import (
	"fmt"
	"sync"
)

type auctusEmu struct {
	mu     sync.Mutex
	mem    []byte // immagine (canali, dati di sistema, testi)
	ver    string
	region string
	slots  int
	log    []int
}

func (auctusFam) emu(m *genModel, img []byte) genEmu {
	e := &auctusEmu{mem: make([]byte, auctusSize), ver: "AU147_US.V1.0", region: "AU147_" + auctusO(m).country, slots: m.Slots}
	copy(e.mem, img)
	for i := 0; i < auctusSys/16; i++ {
		e.mem[i*16] = byte(i) // la radio risponde sempre con il numero del canale nel primo byte
	}
	if n := e.mem[auctusSys]; n < 9 || n > 15 {
		e.mem[auctusSys] = 9 // radio nuova: nove byte di sistema
	}
	return e
}

func (e *auctusEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := append([]byte{}, e.mem...)
	for k := range out[auctusVer:] {
		out[auctusVer+k] = 0
	}
	copy(out[auctusVer:auctusVer+32], e.ver)
	copy(out[auctusRegion:auctusRegion+32], e.region)
	return out
}

func (e *auctusEmu) serve(p Port) error {
	for {
		h, err := emuRead(p, 2)
		if err != nil {
			return nil
		}
		if h[0] != 0xAA || h[1] < 6 {
			return fmt.Errorf("pacchetto inatteso % X", h)
		}
		rest, err := emuRead(p, int(h[1])-2)
		if err != nil {
			return err
		}
		n := len(rest)
		if rest[n-1] != 0xBB || auctusSum(append([]byte{h[1]}, rest[:n-2]...)) != rest[n-2] {
			return fmt.Errorf("somma di controllo errata")
		}
		cmd, sub, data := rest[0], rest[1], rest[2:n-2]
		e.mu.Lock()
		e.log = append(e.log, int(cmd)<<8|int(sub))
		var reply []byte
		switch {
		case cmd == 1:
			reply = []byte(e.ver)
		case cmd == 2:
			reply = []byte(e.region)
		case cmd == 3 && sub == 1 && len(data) == 1 && int(data[0]) < e.slots:
			reply = append([]byte{}, e.mem[int(data[0])*16:int(data[0])*16+16]...)
			reply[0] = data[0]
		case cmd == 3 && sub == 0 && len(data) == 16 && int(data[0]) < e.slots:
			copy(e.mem[int(data[0])*16:], data)
			reply = []byte{0}
		case cmd == 4 && sub == 1:
			reply = append([]byte{}, e.mem[auctusSys+1:auctusSys+1+int(e.mem[auctusSys])]...)
		case cmd == 4 && sub == 0 && len(data) == int(e.mem[auctusSys]):
			copy(e.mem[auctusSys+1:], data)
			reply = []byte{0}
		case cmd == 7:
			e.mu.Unlock()
			continue // fine: nessuna risposta
		default:
			e.mu.Unlock()
			return fmt.Errorf("comando inatteso %02X %02X (%d byte)", cmd, sub, len(data))
		}
		e.mu.Unlock()
		p.Write(auctusPkt(cmd, sub, reply))
	}
}
