//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: protocolli dei driver indicati in AUTORI_E_LICENZE.txt.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio simulate per i driver generici: il lato radio dei protocolli, così
// come risulta da CHIRP. Servono ai test (confronto dei byte inviati da CHIRP
// e da questo programma) e alla prova dell'interfaccia su Linux
// (GESTIONE_CANALI_EMU=1).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// recPort registra i byte ricevuti dalla radio simulata (cioè inviati dal PC).
type recPort struct {
	Port
	mu  sync.Mutex
	rx  bytes.Buffer
	ops []string
}

func (r *recPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	err := r.Port.ReadFull(ctx, b, timeout)
	if err == nil {
		r.mu.Lock()
		r.rx.Write(b)
		r.mu.Unlock()
	}
	return err
}

func (r *recPort) got() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]byte{}, r.rx.Bytes()...)
}

// radio simulata generica
type genEmu interface {
	serve(p Port) error // una sessione completa (fino alla chiusura)
	image() []byte      // memoria della radio nel formato dell'immagine
}

func emuByte(p Port) (byte, error) {
	b, err := emuRead(p, 1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// ---------------------------------------------------------------- UV-5R

type uv5rEmu struct {
	mu      sync.Mutex
	img     []byte   // identificativo + principale + ausiliaria (formato CHIRP)
	magic   [][]byte // identificazioni accettate
	ident   []byte   // risposta al comando 0x02 (8 o 12 byte)
	aux     bool
	dropped bool // la radio "perde" il byte 0x1FCF nelle letture da 0x40
}

func newUV5REmu(m *genModel, img []byte) *uv5rEmu {
	o := uv5ro(m)
	e := &uv5rEmu{img: append([]byte{}, img...), magic: [][]byte{o.idents[0]}, ident: append([]byte{}, img[:8]...), aux: o.aux}
	if len(e.img) < uv5rFullSize && o.aux {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, uv5rFullSize-len(e.img))...)
	}
	return e
}

func (e *uv5rEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *uv5rEmu) memAt(addr int) []byte {
	switch {
	case addr < 0x1800:
		return e.img[8+addr:]
	case addr >= 0x1EC0 && addr < 0x2000 && e.aux:
		return e.img[0x1808+addr-0x1EC0:]
	}
	return nil
}

func (e *uv5rEmu) serve(p Port) error {
	// identificazione: 7 byte
	var got []byte
	for {
		b, err := emuByte(p)
		if err != nil {
			return err
		}
		got = append(got, b)
		if len(got) < 7 {
			continue
		}
		ok := false
		for _, m := range e.magic {
			if bytes.Equal(got[len(got)-7:], m) {
				ok = true
			}
		}
		if ok {
			break
		}
		if len(got) > 7 {
			got = got[1:]
		}
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write(e.ident)
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		switch c {
		case 'S':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr, size := int(h[0])<<8|int(h[1]), int(h[2])
			data := bytes.Repeat([]byte{0xFF}, size)
			e.mu.Lock()
			if src := e.memAt(addr); src != nil {
				copy(data, src)
			}
			if e.dropped && size == 0x40 && addr == 0x1FC0 {
				data = append(append(append([]byte{}, data[:15]...), 0xFF), data[15:63]...)
			}
			e.mu.Unlock()
			p.Write(append([]byte{'X', h[0], h[1], h[2]}, data...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma 0x06 attesa")
			}
			p.Write([]byte{0x06})
		case 'X':
			h, err := emuRead(p, 3)
			if err != nil {
				return err
			}
			addr, size := int(h[0])<<8|int(h[1]), int(h[2])
			d, err := emuRead(p, size)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if dst := e.memAt(addr); dst != nil {
				copy(dst, d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- UV-17Pro

type uv17Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
}

func (uv17Fam) emu(m *genModel, img []byte) genEmu {
	o := uv17o(m)
	e := &uv17Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < o.total {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, o.total-len(e.img))...)
	}
	e.img = e.img[:o.total]
	return e
}

func (e *uv17Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

// posizione nell'immagine di un indirizzo della radio
func (e *uv17Emu) pos(addr int) int {
	o := uv17o(e.m)
	p := 0
	for k, s := range o.starts {
		if addr >= s && addr < s+o.sizes[k] {
			return p + addr - s
		}
		p += o.sizes[k]
	}
	return -1
}

func (e *uv17Emu) serve(p Port) error {
	o := uv17o(e.m)
	id, err := emuRead(p, 16)
	if err != nil {
		return err
	}
	if !bytes.Equal(id, o.idents[0]) {
		return fmt.Errorf("identificazione inattesa %q", id)
	}
	p.Write([]byte{0x06})
	for _, mg := range o.magics {
		got, err := emuRead(p, len(mg.cmd))
		if err != nil {
			return err
		}
		if !bytes.Equal(got, mg.cmd) {
			return fmt.Errorf("comando %x atteso, ricevuto %x", mg.cmd, got)
		}
		resp := bytes.Repeat([]byte{0x06}, mg.resp)
		if mg.resp > 1 {
			copy(resp, []byte("SIMULATA-"+e.m.Model))
		}
		if mg.cmd[0] == 'M' && o.modelIDs != nil {
			resp = []byte(o.modelIDs[0])
		}
		p.Write(resp)
	}
	key := []byte(btSymbols[o.encrsym])
	for {
		h, err := emuRead(p, 4)
		if err != nil {
			return nil
		}
		addr, n := int(h[1])<<8|int(h[2]), int(h[3])
		switch h[0] {
		case 'R':
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if q := e.pos(addr); q >= 0 {
				copy(d, e.img[q:])
			}
			e.mu.Unlock()
			if o.encr {
				btCrypt(d, key)
			}
			p.Write(append(append([]byte{}, h...), d...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			if o.encr {
				btCrypt(d, key)
			}
			e.mu.Lock()
			if q := e.pos(addr); q >= 0 {
				copy(e.img[q:q+n], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", h[0])
		}
	}
}

// ---------------------------------------------------------------- BaofengCommonHT (UV-9R, BF-A58…)

type bfcEmu struct {
	mu     sync.Mutex
	img    []byte // 0x2000 di memoria + identificativo
	magic  []byte
	noAck  bool
	identN int
}

func (bfcFam) emu(m *genModel, img []byte) genEmu {
	n := bfco(m).identN
	e := &bfcEmu{img: append([]byte{}, img...), magic: bfco(m).magics[0], noAck: bfco(m).noAck, identN: n}
	if len(e.img) < 0x2000+n {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x2000+n-len(e.img))...)
	}
	e.img = e.img[:0x2000+n]
	if e.img[0x2000] != 0xAA || e.img[0x2000+n-1] != 0xDD {
		for k := 0; k < n; k++ {
			e.img[0x2000+k] = byte(0x30 + k)
		}
		e.img[0x2000], e.img[0x2000+n-1] = 0xAA, 0xDD
	}
	return e
}

func (e *bfcEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img[:0x2000]...)
}

func (e *bfcEmu) serve(p Port) error {
	mg, err := emuRead(p, len(e.magic))
	if err != nil {
		return err
	}
	if !bytes.Equal(mg, e.magic) {
		return fmt.Errorf("identificazione inattesa % X", mg)
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write(e.img[0x2000 : 0x2000+e.identN])
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, size := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'S':
			d := bytes.Repeat([]byte{0xFF}, size)
			e.mu.Lock()
			if addr < 0x2000 {
				copy(d, e.img[addr:0x2000])
			}
			e.mu.Unlock()
			p.Write(append([]byte{'X', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			if !e.noAck {
				p.Write([]byte{0x06})
			}
		case 'X':
			d, err := emuRead(p, size)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if addr+size <= 0x2000 {
				copy(e.img[addr:], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- BF-888S

type h777Emu struct {
	mu    sync.Mutex
	m     *genModel
	img   []byte
	ident []byte
}

func (h777Fam) emu(m *genModel, img []byte) genEmu {
	o := h777o(m)
	e := &h777Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < o.memsize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, o.memsize-len(e.img))...)
	}
	e.img = e.img[:o.memsize]
	e.ident = append([]byte{}, o.idents[0]...)
	for len(e.ident) < 8 {
		e.ident = append(e.ident, 0xF4)
	}
	return e
}

func (e *h777Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *h777Emu) serve(p Port) error {
	o := h777o(e.m)
	done := false
start:
	if b, err := emuByte(p); err != nil || b != 0x02 {
		if done && err != nil {
			return nil // collegamento chiuso dopo la fine della sessione
		}
		return fmt.Errorf("0x02 atteso")
	}
	cmd, err := emuRead(p, len(o.cmd))
	if err != nil || !bytes.Equal(cmd, o.cmd) {
		return fmt.Errorf("comando di programmazione inatteso %q", cmd)
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write(e.ident[:8])
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		if c == 'E' {
			if o.endF {
				// GT-22: 'F', poi il programma ripete la presentazione
				p.Write([]byte{'F'})
			}
			// la radio torna in attesa di una nuova sessione
			done = true
			goto start
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		addr, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			d := bytes.Repeat([]byte{0xFF}, n)
			e.mu.Lock()
			if addr < len(e.img) {
				copy(d, e.img[addr:])
			}
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if addr+n <= len(e.img) {
				copy(e.img[addr:], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- UV-17 (pagine numerate)

type uv17lEmu struct {
	mu    sync.Mutex
	m     *genModel
	pages map[int][]byte // indice fisico (1…) -> 4 KB
	num   map[int]byte   // indice fisico -> numero di pagina (BCD); 0xFF libera
	n     int
}

func (uv17lFam) emu(m *genModel, img []byte) genEmu {
	o := m.Opt.(*uv17lOpt)
	e := &uv17lEmu{m: m, pages: map[int][]byte{}, num: map[int]byte{}}
	e.n = (o.memsize + 0x1000 - 0x1FFF + 0xFFF) / 0x1000
	// pagine in ordine sparso (come dopo l'usura), le altre libere
	phys := []int{3, 1, 7, 2, 9, 5, 4, 11, 6}
	for k := 1; k <= e.n; k++ {
		e.num[k] = 0xFF
		e.pages[k] = bytes.Repeat([]byte{0xFF}, 0x1000)
	}
	for bi, bn := range uv17BlockOrder {
		p := phys[bi]
		e.num[p] = byte(bn/10)<<4 | byte(bn%10)
		if (bi+1)*0x1000 <= len(img) {
			copy(e.pages[p], img[bi*0x1000:(bi+1)*0x1000])
		}
	}
	return e
}

func (e *uv17lEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []byte
	for _, bn := range uv17BlockOrder {
		for k := 1; k <= e.n; k++ {
			if e.num[k] == byte(bn/10)<<4|byte(bn%10) {
				out = append(out, e.pages[k]...)
			}
		}
	}
	return out
}

func (e *uv17lEmu) serve(p Port) error {
	o := e.m.Opt.(*uv17lOpt)
	expect := func(want []byte, resp []byte) error {
		got, err := emuRead(p, len(want))
		if err != nil {
			return err
		}
		if !bytes.Equal(got, want) {
			return fmt.Errorf("atteso % X, ricevuto % X", want, got)
		}
		p.Write(resp)
		return nil
	}
	if err := expect([]byte("PSEARCH"), uv17lFingerprint); err != nil {
		return err
	}
	for _, mg := range uv17lMagics {
		if err := expect(mg.cmd, bytes.Repeat([]byte{0x06}, mg.resp)); err != nil {
			return err
		}
	}
	ms := make([]byte, 11)
	binary.LittleEndian.PutUint32(ms[7:], uint32(o.memsize))
	if err := expect(uv17lMemsizeCmd, ms); err != nil {
		return err
	}
	for _, mg := range uv17lMagics2 {
		if err := expect(mg.cmd, bytes.Repeat([]byte{0x06}, mg.resp)); err != nil {
			return err
		}
	}
	for {
		h, err := emuRead(p, 5)
		if err != nil {
			return nil
		}
		addr := int(h[1]) | int(h[2])<<8 | int(h[3])<<16
		n := int(h[4])
		pg, off := addr/0x1000, addr%0x1000
		switch h[0] {
		case 'R':
			e.mu.Lock()
			var d []byte
			if n == 1 && off == 0xFFF {
				// numero della pagina nell'ultimo byte della pagina stessa (come
				// _get_memory_map di CHIRP: indirizzo 0x1FFF → pagina 1, letta da 0x1000)
				d = []byte{e.num[pg]}
			} else if pd, ok := e.pages[pg]; ok {
				d = append([]byte{}, pd[off:off+n]...)
			} else {
				d = bytes.Repeat([]byte{0xFF}, n)
			}
			e.mu.Unlock()
			p.Write(append(append([]byte{}, h...), d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			if pd, ok := e.pages[pg]; ok {
				copy(pd[off:], d)
			}
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", h[0])
		}
	}
}

// ---------------------------------------------------------------- UV-B5

type uvb5Emu struct {
	mu  sync.Mutex
	img []byte // 0x1030 (intestazione + memoria)
}

func (uvb5Fam) emu(m *genModel, img []byte) genEmu {
	e := &uvb5Emu{img: append([]byte{}, img...)}
	if len(e.img) < 0x1030 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x1030-len(e.img))...)
	}
	e.img = e.img[:0x1030]
	return e
}

func (e *uvb5Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img[0x30:]...)
}

func (e *uvb5Emu) serve(p Port) error {
	if got, err := emuRead(p, 8); err != nil || string(got) != "\x05PROGRAM" {
		return fmt.Errorf("comando di programmazione atteso")
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write([]byte("HKT511\x00\x01"))
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			d := append([]byte{}, e.img[0x30+a:0x30+a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[0x30+a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- BF-T1

type bft1Emu struct {
	mu  sync.Mutex
	img []byte // 0x800
}

func (bft1Fam) emu(m *genModel, img []byte) genEmu {
	e := &bft1Emu{img: append([]byte{}, img...)}
	if len(e.img) < 0x800 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0x800-len(e.img))...)
	}
	e.img = e.img[:0x800]
	return e
}

func (e *bft1Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *bft1Emu) serve(p Port) error {
	if got, err := emuRead(p, 8); err != nil || string(got) != "\x05PROGRAM" {
		return fmt.Errorf("comando di programmazione atteso")
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write([]byte(" BF9100S"))
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil || c == 0x62 {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			if a+n > len(e.img) {
				// indirizzo fuori dalla memoria (richiesta del driver di un altro
				// modello, ad esempio durante l'identificazione): la radio tace
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- protocollo 'R'/'W' con conferma (BF-T8…)

type rbEmu struct {
	mu  sync.Mutex
	sp  *rbSpec
	img []byte
}

func newRBEmu(sp *rbSpec, img []byte) *rbEmu {
	e := &rbEmu{sp: sp, img: append([]byte{}, img...)}
	if len(e.img) < sp.memsize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, sp.memsize-len(e.img))...)
	}
	e.img = e.img[:sp.memsize]
	return e
}

func (bft8Fam) emu(m *genModel, img []byte) genEmu { return newRBEmu(bft8o(m).spec, img) }
func (t18Fam) emu(m *genModel, img []byte) genEmu  { return newRBEmu(t18o(m).spec, img) }

func (e *rbEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *rbEmu) serve(p Port) error {
	if got, err := emuRead(p, len(e.sp.magic)); err != nil || !bytes.Equal(got, e.sp.magic) {
		return fmt.Errorf("comando di programmazione atteso (%q)", got)
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write(e.sp.finger)
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil || c == e.sp.exitCmd()[0] {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			if a+n > len(e.img) {
				// indirizzo fuori dalla memoria (richiesta del driver di un altro
				// modello, ad esempio durante l'identificazione): la radio tace
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- UV-3R (protocollo Wouxun)

type uv3rEmu struct {
	mu  sync.Mutex
	img []byte
}

func (uv3rFam) emu(m *genModel, img []byte) genEmu {
	e := &uv3rEmu{img: append([]byte{}, img...)}
	if len(e.img) < 0xE40 {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, 0xE40-len(e.img))...)
	}
	e.img = e.img[:0xE40]
	return e
}

func (e *uv3rEmu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *uv3rEmu) serve(p Port) error {
	if got, err := emuRead(p, 8); err != nil || string(got) != "\x05PROGRAM" {
		return fmt.Errorf("comando di programmazione atteso")
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	p.Write([]byte("UV3R\x00\x01\x02\x03"))
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	for {
		c, err := emuByte(p)
		if err != nil {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			if a+n > len(e.img) {
				// indirizzo fuori dalla memoria (richiesta del driver di un altro
				// modello, ad esempio durante l'identificazione): la radio tace
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// ---------------------------------------------------------------- BF-T20, W31E

type rt22Emu struct {
	mu  sync.Mutex
	m   *genModel
	img []byte
}

func (rt22Fam) emu(m *genModel, img []byte) genEmu {
	o := rt22o(m)
	e := &rt22Emu{m: m, img: append([]byte{}, img...)}
	if len(e.img) < o.memsize {
		e.img = append(e.img, bytes.Repeat([]byte{0xFF}, o.memsize-len(e.img))...)
	}
	e.img = e.img[:o.memsize]
	return e
}

func (e *rt22Emu) image() []byte { e.mu.Lock(); defer e.mu.Unlock(); return append([]byte{}, e.img...) }

func (e *rt22Emu) serve(p Port) error {
	o := rt22o(e.m)
	if got, err := emuRead(p, 7); err != nil || string(got) != "PROGRGS" {
		return fmt.Errorf("comando di programmazione atteso")
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x02 {
		return fmt.Errorf("0x02 atteso")
	}
	id := []byte("P3T22\xF8")
	if len(o.fileid) > 0 {
		id = append([]byte{}, o.fileid[0]...)
	}
	for len(id) < 8 {
		id = append(id, 0xFF)
	}
	p.Write(id[:8])
	if b, err := emuByte(p); err != nil || b != 0x06 {
		return fmt.Errorf("0x06 atteso")
	}
	p.Write([]byte{0x06})
	if b, err := emuByte(p); err != nil || b != 0x07 {
		return fmt.Errorf("0x07 atteso")
	}
	p.Write([]byte{0x4E})
	for {
		c, err := emuByte(p)
		if err != nil || c == 'E' {
			return nil
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return err
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		switch c {
		case 'R':
			e.mu.Lock()
			if a+n > len(e.img) {
				// indirizzo fuori dalla memoria (richiesta del driver di un altro
				// modello, ad esempio durante l'identificazione): la radio tace
				e.mu.Unlock()
				return fmt.Errorf("indirizzo 0x%04X fuori dalla memoria", a)
			}
			d := append([]byte{}, e.img[a:a+n]...)
			e.mu.Unlock()
			p.Write(append([]byte{'W', h[0], h[1], h[2]}, d...))
			if b, err := emuByte(p); err != nil || b != 0x06 {
				return fmt.Errorf("conferma attesa")
			}
			p.Write([]byte{0x06})
		case 'W':
			d, err := emuRead(p, n)
			if err != nil {
				return err
			}
			e.mu.Lock()
			copy(e.img[a:], d)
			e.mu.Unlock()
			p.Write([]byte{0x06})
		default:
			return fmt.Errorf("comando sconosciuto %02X", c)
		}
	}
}

// genEmuFor: radio simulata per un modello generico
func genEmuFor(m *genModel, img []byte) genEmu {
	switch m.Fam.(type) {
	case uv5rFam:
		return newUV5REmu(m, img)
	}
	if f, ok := m.Fam.(interface {
		emu(m *genModel, img []byte) genEmu
	}); ok {
		return f.emu(m, img)
	}
	return nil
}
