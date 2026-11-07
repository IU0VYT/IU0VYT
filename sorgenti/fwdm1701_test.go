//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"testing"
)

// dm1701FakeFile: file «OutSecurityBin» con le aree indicate
func dm1701FakeFile(magic []byte, regs []dm1701Region) []byte {
	b := bytes.Repeat([]byte{0xFF}, dm1701FwData)
	copy(b, dm1701FwMagic+"\x00\x00")
	copy(b[0x10:], "DM1701\x00\x00")
	for i := range b[0x30 : 0x30+76] {
		b[0x30+i] = byte(i)
	}
	copy(b[0x30:], magic)
	binary.LittleEndian.PutUint32(b[0x7C:], uint32(len(regs)))
	for i, r := range regs {
		binary.LittleEndian.PutUint32(b[0x80+8*i:], r.addr)
		binary.LittleEndian.PutUint32(b[0x84+8*i:], uint32(len(r.data)))
	}
	for _, r := range regs {
		b = append(b, r.data...)
	}
	b = append(b, bytes.Repeat([]byte{0xFF}, 240)...)
	return append(b, dm1701FwEnd...)
}

func fakeData(n int, seed byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*13) ^ seed ^ byte(i>>10)
	}
	return b
}

// dm1701Boot: caricatore DFU della DM-1701 (stati, comandi e memoria flash)
type dm1701Boot struct {
	mu      sync.Mutex
	flash   []byte // 0x08000000-0x080FFFFF
	state   byte
	busy    int // GETSTATUS «occupato» ancora da dare
	ptr     uint32
	fwMode  bool
	reg     bool
	errs    []string
	erased  map[uint32]bool
	ident   string
	rejectN int // blocco di dati (da 1) rifiutato
	nData   int
}

func newDM1701Boot() *dm1701Boot {
	f := make([]byte, 0x100000)
	for i := range f {
		f[i] = byte(i * 7) // caricatore e programma vecchio
	}
	return &dm1701Boot{flash: f, state: dfuStIdle, erased: map[uint32]bool{}, ident: "DM-1701"}
}

func (e *dm1701Boot) fail(f string, a ...interface{}) { e.errs = append(e.errs, fmt.Sprintf(f, a...)) }

func (e *dm1701Boot) serve(p Port) error {
	reply := func(res int32, data []byte) error {
		b := make([]byte, 4, 4+len(data))
		binary.LittleEndian.PutUint32(b, uint32(res))
		return p.Write(append(b, data...))
	}
	for {
		var h [8]byte
		if err := emuReadInto(p, h[:]); err != nil {
			return nil
		}
		rt, rq := h[0], h[1]
		value := int(binary.LittleEndian.Uint16(h[2:]))
		n := int(binary.LittleEndian.Uint16(h[6:]))
		in := rt&0x80 != 0
		var out []byte
		if !in && n > 0 {
			out = make([]byte, n)
			if err := emuReadInto(p, out); err != nil {
				return err
			}
		}
		e.mu.Lock()
		var err error
		switch {
		case in && rq == dfuGetState:
			err = reply(1, []byte{e.state})
		case in && rq == dfuGetStatus:
			if e.state == dfuStDnBusy {
				if e.busy > 0 {
					e.busy--
					err = reply(6, []byte{0, 20, 0, 0, dfuStDnBusy, 0})
					break
				}
				e.state = 5
			}
			code := byte(0)
			if e.state == dfuStError {
				code = 0x0A
			}
			err = reply(6, []byte{code, 0, 0, 0, e.state, 0})
		case !in && (rq == dfuAbort || rq == dfuClrStatus):
			e.state = dfuStIdle
			err = reply(0, nil)
		case !in && rq == dfuDnload && value == 0:
			if e.state != dfuStIdle && e.state != 5 {
				e.fail("DNLOAD nello stato %d", e.state)
			}
			e.state, e.busy = dfuStDnBusy, 1
			switch {
			case bytes.Equal(out, []byte{0xA2, 0x01}):
				e.reg = true
			case bytes.Equal(out, []byte{0x91, 0x31}):
				e.fwMode = true
			case len(out) == 5 && out[0] == 0x21:
				e.ptr = binary.LittleEndian.Uint32(out[1:])
			case len(out) == 5 && out[0] == 0x41:
				a := binary.LittleEndian.Uint32(out[1:])
				ok := false
				for _, s := range dm1701Sectors {
					if a >= s[0] && a < s[0]+s[1] {
						ok = true
						for k := s[0]; k < s[0]+s[1]; k++ {
							e.flash[k-0x08000000] = 0xFF
						}
						e.erased[s[0]] = true
						e.busy = 3
					}
				}
				if !ok || !e.fwMode {
					e.fail("cancellazione a 0x%08X non ammessa (aggiornamento %v)", a, e.fwMode)
					e.state = dfuStError
				}
			default:
				e.fail("comando sconosciuto % X", out)
				e.state = dfuStError
			}
			err = reply(int32(len(out)), nil)
		case !in && rq == dfuDnload:
			e.nData++
			a := e.ptr + uint32(value-2)*dm1701FwBlock
			if !e.fwMode || a < dm1701FlashLo || int(a)-0x08000000+len(out) > len(e.flash) || len(out) > dm1701FwBlock || e.nData == e.rejectN {
				e.fail("blocco a 0x%08X non ammesso", a)
				e.state = dfuStError
				err = reply(int32(len(out)), nil)
				break
			}
			for k, b := range out {
				e.flash[int(a)-0x08000000+k] &= b // senza cancellazione resta sbagliato
			}
			e.state, e.busy = dfuStDnBusy, 1
			err = reply(int32(len(out)), nil)
		case in && rq == dfuUpload && value == 0:
			if !e.reg {
				err = reply(-9, nil)
				break
			}
			id := make([]byte, n)
			copy(id, e.ident)
			e.state = 9
			err = reply(int32(n), id)
		default:
			err = fmt.Errorf("richiesta DFU inattesa %02X %02X", rt, rq)
		}
		e.mu.Unlock()
		if err != nil {
			return err
		}
	}
}

func TestDM1701Firmware(t *testing.T) {
	main := fakeData(0x5C000, 1) // 0x0800C000-0x08068000: settori 3-6 e parte del 7
	spi := fakeData(0x3000, 2)
	file := dm1701FakeFile([]byte{0x01, 0x0F}, []dm1701Region{{0x0800C000, main}, {0x00200000, spi}})
	in := analyzeDM1701("DM1701_V2.04.bin", "x", file).Info
	for _, c := range in.Checks {
		t.Logf("%s: %s", c.Level, c.Text)
	}
	if !in.Valid || in.Confirm != "DM1701" || !in.Typed || sniffKind("DM1701_V2.04.bin", file) != fwDM1701 {
		t.Fatalf("%+v", in)
	}
	if analyzeDM1701("x.bin", "x", dm1701FakeFile([]byte{0x02, 0x17, 0x0C}, []dm1701Region{{0x0800C000, main}})).Info.Valid {
		t.Fatal("firmware della MD-UV380 accettato")
	}
	if analyzeDM1701("x.bin", "x", dm1701FakeFile([]byte{0x01, 0x0F}, []dm1701Region{{0x08000000, main}})).Info.Valid {
		t.Fatal("firmware che tocca il caricatore accettato")
	}
	for _, c := range []struct {
		name  string
		setup func(*dm1701Boot)
		ok    bool
	}{
		{"normale", nil, true},
		{"altra radio", func(e *dm1701Boot) { e.ident = "MD-UV390" }, false},
		{"modello non letto", func(e *dm1701Boot) { e.ident = "" }, true},
		{"blocco rifiutato", func(e *dm1701Boot) { e.rejectN = 40 }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := newDM1701Boot()
			if c.setup != nil {
				c.setup(e)
			}
			boot := append([]byte{}, e.flash[:0xC000]...)
			old := openPortFunc
			openPortFunc = func(name string, baud int) (Port, error) {
				if name != dm1701DfuPort {
					return nil, fmt.Errorf("porta %s", name)
				}
				a, b := newPipe()
				go e.serve(b)
				return a, nil
			}
			defer func() { openPortFunc = old }()
			started := false
			sent, id, err := dm1701Update(context.Background(), file, noProg, func(s string) { t.Log(s) }, func() { started = true })
			e.mu.Lock()
			defer e.mu.Unlock()
			t.Logf("errore %v, %d byte, radio %q, avviato %v, %v", err, sent, id, started, e.errs)
			if (err == nil) != c.ok {
				t.Fatal("esito inatteso")
			}
			if !bytes.Equal(e.flash[:0xC000], boot) {
				t.Fatal("caricatore modificato")
			}
			if c.name == "altra radio" && started {
				t.Fatal("aggiornamento avviato con un'altra radio")
			}
			if c.ok {
				if len(e.errs) > 0 || sent != int64(len(main)) || !bytes.Equal(e.flash[0xC000:0xC000+len(main)], main) {
					t.Fatal("firmware scritto diverso")
				}
				// settori toccati cancellati, gli altri no
				for _, s := range dm1701Sectors {
					want := s[0] < 0x0800C000+uint32(len(main))
					if e.erased[s[0]] != want {
						t.Fatalf("settore 0x%08X cancellato %v", s[0], e.erased[s[0]])
					}
				}
			}
		})
	}
	// file vero del costruttore, se presente: IU0VYT_DM1701_FW=percorso
	if p := os.Getenv("IU0VYT_DM1701_FW"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		in := analyzeDM1701(p, p, b).Info
		for _, c := range in.Checks {
			t.Logf("%s: %s", c.Level, c.Text)
		}
		if !in.Valid {
			t.Fatal("firmware del costruttore non riconosciuto")
		}
		// invio completo alla radio simulata: memoria del programma uguale al file
		e := newDM1701Boot()
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) { a, b2 := newPipe(); go e.serve(b2); return a, nil }
		defer func() { openPortFunc = old }()
		if _, _, err := dm1701Update(context.Background(), b, noProg, func(string) {}, func() {}); err != nil {
			t.Fatal(err)
		}
		_, _, regs, _ := dm1701Parse(b)
		for _, r := range regs {
			if dm1701Writable(r) && !bytes.Equal(e.flash[r.addr-0x08000000:][:len(r.data)], r.data) {
				t.Fatalf("area 0x%08X scritta diversa", r.addr)
			}
		}
	}
}
