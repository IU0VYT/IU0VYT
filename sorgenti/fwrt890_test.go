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
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// rt890Fake: programma ARM plausibile per l'AT32F421
func rt890Fake(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*7 + i>>8)
	}
	put := func(off int, v uint32) {
		b[off], b[off+1], b[off+2], b[off+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	put(0, 0x20003FF0)
	put(4, 0x0800150F)
	put(8, 0x080014CD)
	put(12, 0x080014CF)
	return b
}

// rt890Emu: caricatore della RT-890 (come descritto da rt_890_flasher.py)
type rt890Emu struct {
	mu      sync.Mutex
	flash   []byte
	erased  bool
	state   string
	errs    []string
	normal  bool // radio accesa normalmente: non risponde ai comandi del caricatore
	nakAt   int  // blocco (da 1) rifiutato
	silence bool // nessuna risposta alla cancellazione
}

func (e *rt890Emu) fail(f string, a ...interface{}) {
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf(f, a...))
	e.state = "fail"
	e.mu.Unlock()
}

func (e *rt890Emu) serve(p Port, baud int) {
	if baud != rt890Baud {
		e.fail("velocità %d", baud)
		return
	}
	if e.normal {
		return
	}
	ctx := context.Background()
	sumOK := func(b []byte) bool {
		var s byte
		for _, x := range b[:len(b)-1] {
			s += x
		}
		return s == b[len(b)-1]
	}
	blocks := 0
	for {
		c := make([]byte, 1)
		if err := p.ReadFull(ctx, c, 10*time.Second); err != nil {
			return
		}
		switch c[0] {
		case rt890CmdRead:
			r := make([]byte, 3)
			if p.ReadFull(ctx, r, time.Second) != nil || !sumOK(append(c, r...)) {
				e.fail("lettura errata")
				return
			}
			_ = p.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
		case rt890CmdErase:
			r := make([]byte, 4)
			if p.ReadFull(ctx, r, time.Second) != nil || !sumOK(append(c, r...)) || r[2] != 0x55 {
				e.fail("cancellazione errata %x", r)
				return
			}
			if e.silence {
				continue
			}
			e.mu.Lock()
			e.erased = true
			e.flash = bytes.Repeat([]byte{0xFF}, rt890MaxLen)
			e.mu.Unlock()
			_ = p.Write([]byte{rt890Ack})
		case rt890CmdWrite:
			r := make([]byte, 2+rt890Block+1)
			if p.ReadFull(ctx, r, time.Second) != nil || !sumOK(append(c, r...)) {
				e.fail("scrittura errata")
				return
			}
			blocks++
			off := int(r[0])<<8 | int(r[1])
			e.mu.Lock()
			if !e.erased || off%rt890Block != 0 || off+rt890Block > rt890MaxLen {
				e.mu.Unlock()
				e.fail("blocco a 0x%04X non ammesso", off)
				_ = p.Write([]byte{0x15})
				return
			}
			copy(e.flash[off:], r[2:2+rt890Block])
			e.state = "data"
			e.mu.Unlock()
			if blocks == e.nakAt {
				_ = p.Write([]byte{0x15})
				continue
			}
			_ = p.Write([]byte{rt890Ack})
		default:
			e.fail("comando sconosciuto %02X", c[0])
			return
		}
	}
}

func TestRT890Update(t *testing.T) {
	oldWait := rt890BootWait
	rt890BootWait = 3 * time.Second
	oldErase := rt890EraseWait
	rt890EraseWait = time.Second
	defer func() { rt890BootWait, rt890EraseWait = oldWait, oldErase }()
	for _, c := range []struct {
		name    string
		size    int
		setup   func(*rt890Emu)
		ok      bool
		started bool
	}{
		{"normale", 54648, nil, true, true},
		{"memoria piena", rt890MaxLen, nil, true, true},
		{"multiplo di 128", 128 * 300, nil, true, true},
		{"radio accesa normalmente", 30000, func(e *rt890Emu) { e.normal = true }, false, false},
		{"blocco rifiutato", 30000, func(e *rt890Emu) { e.nakAt = 9 }, false, true},
		{"cancellazione senza risposta", 30000, func(e *rt890Emu) { e.silence = true }, false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			fw := rt890Fake(c.size)
			e := &rt890Emu{state: "wait"}
			if c.setup != nil {
				c.setup(e)
			}
			old := openPortFunc
			openPortFunc = func(name string, baud int) (Port, error) {
				a, b := newPipe()
				go e.serve(b, baud)
				return a, nil
			}
			defer func() { openPortFunc = old }()
			started := false
			sent, err := rt890Update(context.Background(), "COM9", fw, noProg, func(s string) { t.Log(s) }, func() { started = true })
			time.Sleep(50 * time.Millisecond)
			e.mu.Lock()
			defer e.mu.Unlock()
			t.Logf("errore %v, %d byte, stato %s, avviato %v, %v", err, sent, e.state, started, e.errs)
			if (err == nil) != c.ok || started != c.started {
				t.Fatal("esito inatteso")
			}
			if c.ok {
				want := append(append([]byte{}, fw...), make([]byte, (rt890Block-len(fw)%rt890Block)%rt890Block)...)
				if len(e.errs) > 0 || sent != int64(len(fw)) || !bytes.Equal(e.flash[:len(want)], want) {
					t.Fatal("firmware ricevuto diverso")
				}
				if bytes.Count(e.flash[len(want):], []byte{0xFF}) != rt890MaxLen-len(want) {
					t.Fatal("memoria oltre il firmware modificata")
				}
			}
		})
	}
}

func TestRT890Analyze(t *testing.T) {
	fw := rt890Fake(54648)
	in := analyzeRT890("RT-890_V1.18.bin", "x", fw).Info
	if !in.Valid || in.Confirm != "RT890" || !in.Typed || in.Kind != "rt890" {
		t.Fatalf("%+v", in)
	}
	if sniffKind("RT-890_V1.18.bin", fw) != fwRT890 {
		t.Fatal("tipo del file non riconosciuto")
	}
	bad := append([]byte{}, fw...)
	bad[4] = 0x00 // avvio pari: non è un programma Thumb
	if analyzeRT890("RT890.bin", "x", bad).Info.Valid || rt890Plausible(bad) {
		t.Fatal("file non valido accettato")
	}
	if rt890Plausible(rt890Fake(rt890MaxLen + 16)) {
		t.Fatal("file troppo grande accettato")
	}
	// firmware vero (M7OCM v3.2), se presente: IU0VYT_RT890_FW=percorso
	if p := os.Getenv("IU0VYT_RT890_FW"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		in := analyzeRT890("m7ocm-v3.2-pcb2.0.bin", p, b).Info
		t.Logf("%s: valido %v, conosciuto %q", p, in.Valid, in.Known)
		for _, c := range in.Checks {
			t.Logf("  %s: %s", c.Level, c.Text)
		}
		if !in.Valid || in.Known == "" || sniffKind("m7ocm-v3.2-pcb2.0.bin", b) != fwRT890 {
			t.Fatal("firmware M7OCM non riconosciuto")
		}
	}
}
