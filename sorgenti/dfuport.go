// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Collegamento con le radio in modalità USB DFU (DM-1701): il Port riceve le
// richieste di controllo USB nel formato di drv_tyt.go (8 byte + dati in uscita)
// e restituisce l'esito (4 byte) e i dati in entrata; ogni richiesta passa alla
// funzione xfer del driver di Windows in uso (STTub30 o WinUSB, usb_windows.go).

import (
	"context"
	"time"
)

type dfuPort struct {
	// xfer: una richiesta di controllo; esito = byte trasferiti o -1
	xfer  func(in bool, rt, req byte, value, index uint16, out []byte, n int) (int32, []byte)
	close func() error
	rbuf  []byte
	wbuf  []byte
}

func (p *dfuPort) Write(b []byte) error {
	p.wbuf = append(p.wbuf, b...)
	for len(p.wbuf) >= 8 {
		h := p.wbuf
		in := h[0]&0x80 != 0
		n := int(h[6]) | int(h[7])<<8
		need := 8
		if !in {
			need += n
		}
		if len(p.wbuf) < need {
			return nil
		}
		value, index := uint16(h[2])|uint16(h[3])<<8, uint16(h[4])|uint16(h[5])<<8
		var out []byte
		if !in {
			out = append([]byte{}, h[8:8+n]...)
		}
		res, data := p.xfer(in, h[0], h[1], value, index, out, n)
		r := []byte{byte(res), byte(res >> 8), byte(res >> 16), byte(res >> 24)}
		p.rbuf = append(p.rbuf, r...)
		if in && res > 0 {
			p.rbuf = append(p.rbuf, data...)
		}
		p.wbuf = p.wbuf[need:]
	}
	return nil
}

func (p *dfuPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if len(p.rbuf) < len(b) {
		return errTimeout
	}
	copy(b, p.rbuf)
	p.rbuf = p.rbuf[len(b):]
	return nil
}

func (p *dfuPort) Purge() error      { p.rbuf, p.wbuf = nil, nil; return nil }
func (p *dfuPort) SetBaud(int) error { return nil }
func (p *dfuPort) Close() error      { return p.close() }
