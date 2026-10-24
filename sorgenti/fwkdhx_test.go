//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove dei caricatori KDH più recenti (UV-25 Plus .kdhx, BJ-9000 .Firmware,
// UV-9G .kdh) con le radio simulate; TestDiffKdh2 confronta i byte con quelli
// del codice originale dei programmi del costruttore (verifica/LEGGIMI_VERIFICA.txt).

import (
	"context"
	"encoding/binary"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeARM: firmware sintetico con tabella dei vettori plausibile (base 0x08001000)
func fakeARM(size int, seed int64) []byte {
	b := make([]byte, size)
	rand.New(rand.NewSource(seed)).Read(b)
	v := []uint32{0x200059d0, 0x080010d5}
	for i := 2; i < 48; i++ {
		v = append(v, 0x08001105+uint32(i%5)*0x40)
	}
	for i, x := range v {
		binary.LittleEndian.PutUint32(b[i*4:], x)
	}
	return b
}

func fakeKDHX(size int) []byte { return fakeARM(size, int64(size)) }

func fakeBJ9000(size int) []byte {
	b := fakeARM(size, int64(size)*3)
	copy(b[992:], "RT-950      ")
	return b
}

func fake9G(size int) []byte {
	b := make([]byte, size)
	rand.New(rand.NewSource(int64(size) * 5)).Read(b)
	return b
}

type kdh2Scen struct {
	name   string
	v      *kdh2Variant
	raw    []byte
	manual bool
	setup  func(*kdh2Emu)
	ok     bool
}

func kdh2Scens() []kdh2Scen {
	return []kdh2Scen{
		{"kdhx automatico", kdhxProto, fakeKDHX(105400), false, nil, true},
		{"kdhx manuale", kdhxProto, fakeKDHX(105400), true, nil, true},
		{"kdhx multiplo di 1024", kdhxProto, fakeKDHX(8192), false, nil, true},
		{"kdhx risposta persa", kdhxProto, fakeKDHX(20000), true, func(e *kdh2Emu) { e.drop[5] = true }, true},
		{"kdhx errore di controllo", kdhxProto, fakeKDHX(20000), true, func(e *kdh2Emu) { e.nak[6] = true }, true},
		{"kdhx CRC errato", kdhxProto, fakeKDHX(20000), true, func(e *kdh2Emu) { e.corrupt[4] = true }, true},
		{"kdhx presentazione persa", kdhxProto, fakeKDHX(9000), false, func(e *kdh2Emu) { e.hsDrop = 2 }, true},
		{"kdhx errore di scrittura", kdhxProto, fakeKDHX(9000), true, func(e *kdh2Emu) { e.reject = 0xE4 }, false},
		{"kdhx altra radio", kdhxProto, fakeKDHX(9000), false, func(e *kdh2Emu) { e.wrongModel = true }, false},
		{"bj9000 automatico", bj9000Proto, fakeBJ9000(60000), false, nil, true},
		{"bj9000 manuale", bj9000Proto, fakeBJ9000(60000), true, nil, true},
		{"bj9000 risposta persa", bj9000Proto, fakeBJ9000(20480), false, func(e *kdh2Emu) { e.drop[7] = true }, true},
		{"bj9000 errore di controllo", bj9000Proto, fakeBJ9000(20000), true, func(e *kdh2Emu) { e.nak[5] = true }, true},
		{"bj9000 CRC errato", bj9000Proto, fakeBJ9000(20000), true, func(e *kdh2Emu) { e.corrupt[3] = true }, true},
		{"bj9000 modello errato", bj9000Proto, fakeBJ9000(20000), true, func(e *kdh2Emu) { e.wrongModel = true }, false},
		{"bj9000 errore di scrittura", bj9000Proto, fakeBJ9000(20000), true, func(e *kdh2Emu) { e.reject = 0xE4 }, false},
	}
}

func TestKdh2Emu(t *testing.T) {
	for _, s := range append(kdh2Scens(),
		kdh2Scen{"kdhx radio spenta", kdhxProto, fakeKDHX(9000), false, func(e *kdh2Emu) { e.silent = true }, false},
		kdh2Scen{"bj9000 radio spenta", bj9000Proto, fakeBJ9000(9000), true, func(e *kdh2Emu) { e.silent = true }, false}) {
		t.Run(s.name, func(t *testing.T) {
			e := newKdh2Emu(s.v, s.raw, s.manual)
			if s.setup != nil {
				s.setup(e)
			}
			old := openPortFunc
			openPortFunc = kdh2EmuPorts(e)
			defer func() { openPortFunc = old }()
			st := false
			t0 := time.Now()
			sent, err := kdh2Update(context.Background(), "COM9", s.raw, s.v, s.manual, noProg, func(m string) { t.Log(m) }, func() { st = true })
			time.Sleep(50 * time.Millisecond)
			log, errs, state := e.snapshot()
			t.Logf("errore %v, %d byte, stato %s, %d passi, %v, avviato %v", err, sent, state, len(log), time.Since(t0).Round(time.Millisecond), st)
			if len(errs) > 0 {
				t.Fatalf("segnalazioni della radio: %v", errs)
			}
			if (err == nil) != s.ok || (s.ok && (state != "done" || sent != int64(len(s.raw)))) {
				t.Fatalf("esito inatteso")
			}
			for _, l := range log {
				if strings.Contains(l, "rovinati") {
					t.Fatalf("byte inviati alla velocità sbagliata: %v", log)
				}
			}
		})
	}
}

type k9gScen struct {
	name   string
	raw    []byte
	manual bool
	setup  func(*k9gEmu)
	ok     bool
}

func k9gScens() []k9gScen {
	return []k9gScen{
		{"manuale", fake9G(35332), true, nil, true},
		{"automatico", fake9G(35332), false, nil, true},
		{"multiplo di 32", fake9G(4096), true, nil, true},
		{"risposta persa", fake9G(3000), true, func(e *k9gEmu) { e.drop[5] = true; e.drop[9] = true }, true},
		{"errore nei dati", fake9G(3000), true, func(e *k9gEmu) { e.nak[4] = true }, true},
		{"errore di scrittura", fake9G(3000), true, func(e *k9gEmu) { e.reject = 0xE3 }, false},
	}
}

func TestK9gEmu(t *testing.T) {
	for _, s := range append(k9gScens(),
		k9gScen{"presentazione persa", fake9G(3000), false, func(e *k9gEmu) { e.hsDrop = 2 }, true},
		k9gScen{"radio spenta", fake9G(3000), true, func(e *k9gEmu) { e.silent = true }, false}) {
		t.Run(s.name, func(t *testing.T) {
			e := newK9gEmu(s.raw, s.manual)
			if s.setup != nil {
				s.setup(e)
			}
			old := openPortFunc
			openPortFunc = k9gEmuPorts(e)
			defer func() { openPortFunc = old }()
			t0 := time.Now()
			sent, err := kdh9gUpdate(context.Background(), "COM9", s.raw, s.manual, noProg, func(m string) { t.Log(m) }, nil)
			time.Sleep(50 * time.Millisecond)
			log, errs, state := e.snapshot()
			t.Logf("errore %v, %d byte, stato %s, %d passi, %v", err, sent, state, len(log), time.Since(t0).Round(time.Millisecond))
			if len(errs) > 0 {
				t.Fatalf("segnalazioni della radio: %v", errs)
			}
			if (err == nil) != s.ok || (s.ok && state != "done") {
				t.Fatalf("esito inatteso")
			}
		})
	}
}

// ---------------------------------------------------------------- confronto con il codice originale

// TestDiffKdh2: BF_HARNESS=…/bf_harness.dll e uno o più tra
// KDHX_EXE=…/Baofeng Bootloader Pro.exe, BJ9000_EXE=…/BJ9000_EnUPDATE.exe,
// K9G_EXE=…/Bootloader_BaudRate38400.exe; KDH2_REAL=cartella con firmware veri.
func TestDiffKdh2(t *testing.T) {
	harness := os.Getenv("BF_HARNESS")
	if harness == "" {
		t.Skip("BF_HARNESS non impostato")
	}
	exes := map[string]string{"kdhx": os.Getenv("KDHX_EXE"), "bj9000": os.Getenv("BJ9000_EXE"), "kdh9g": os.Getenv("K9G_EXE")}
	dir := t.TempDir()
	scens := kdh2Scens()
	var real9g []k9gScen
	if d := os.Getenv("KDH2_REAL"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, "*"))
		for _, f := range fs {
			b, _ := os.ReadFile(f)
			switch strings.ToLower(filepath.Ext(f)) {
			case ".kdhx":
				scens = append(scens, kdh2Scen{"vero " + filepath.Base(f), kdhxProto, b, false, nil, true},
					kdh2Scen{"vero manuale " + filepath.Base(f), kdhxProto, b, true, nil, true})
			case ".firmware", ".btf":
				scens = append(scens, kdh2Scen{"vero " + filepath.Base(f), bj9000Proto, b, false, nil, true})
			case ".kdh":
				real9g = append(real9g, k9gScen{"vero " + filepath.Base(f), b, true, nil, true})
			}
		}
	}
	run := func(t *testing.T, exe, file, addr, mode string, baud int) (string, error) {
		out, err := exec.Command("dotnet", harness, exe, file, addr, mode, fmtInt(baud)).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	for _, s := range scens {
		exe := exes[s.v.ID]
		if exe == "" {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			file := filepath.Join(dir, "fw.bin")
			os.WriteFile(file, s.raw, 0o644)
			mk := func() *kdh2Emu {
				e := newKdh2Emu(s.v, s.raw, s.manual)
				if s.setup != nil {
					s.setup(e)
				}
				return e
			}
			eo := mk()
			addr, stop := serveHarnessFn(func(p Port, baud int) { eo.serve(p, baud) })
			mode, baud := "auto", s.v.HelloBaud
			if s.manual {
				mode, baud = "manual", s.v.BootBaud
			}
			out, herr := run(t, exe, file, addr, mode, baud)
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, errsO, stateO := eo.snapshot()

			eg := mk()
			old := openPortFunc
			openPortFunc = kdh2EmuPorts(eg)
			_, gerr := kdh2Update(context.Background(), "COM9", s.raw, s.v, s.manual, noProg, nil, nil)
			openPortFunc = old
			time.Sleep(100 * time.Millisecond)
			logG, errsG, stateG := eg.snapshot()
			t.Logf("originale: %s (uscita %v) — stato %s, %d passi", out, herr, stateO, len(logO))
			t.Logf("IU0VYT Flash: errore %v — stato %s, %d passi", gerr, stateG, len(logG))
			if len(errsO) > 0 || len(errsG) > 0 {
				t.Fatalf("segnalazioni della radio simulata: originale %v, IU0VYT Flash %v", errsO, errsG)
			}
			if stateO != stateG || (herr == nil) != (gerr == nil) {
				t.Fatalf("esito diverso")
			}
			cmpLogs(t, logO, logG)
		})
	}
	if exe := exes["kdh9g"]; exe != "" {
		for _, s := range append(k9gScens(), real9g...) {
			t.Run("kdh9g "+s.name, func(t *testing.T) {
				file := filepath.Join(dir, "fw.kdh")
				os.WriteFile(file, s.raw, 0o644)
				mk := func() *k9gEmu {
					e := newK9gEmu(s.raw, s.manual)
					if s.setup != nil {
						s.setup(e)
					}
					return e
				}
				eo := mk()
				addr, stop := serveHarnessFn(func(p Port, baud int) { eo.serve(p, baud) })
				mode, baud := "auto", 9600
				if s.manual {
					mode, baud = "manual", k9gBaud
				}
				out, herr := run(t, exe, file, addr, mode, baud)
				stop()
				time.Sleep(100 * time.Millisecond)
				logO, errsO, stateO := eo.snapshot()

				eg := mk()
				old := openPortFunc
				openPortFunc = k9gEmuPorts(eg)
				_, gerr := kdh9gUpdate(context.Background(), "COM9", s.raw, s.manual, noProg, nil, nil)
				openPortFunc = old
				time.Sleep(100 * time.Millisecond)
				logG, errsG, stateG := eg.snapshot()
				t.Logf("originale: %s (uscita %v) — stato %s, %d passi", out, herr, stateO, len(logO))
				t.Logf("IU0VYT Flash: errore %v — stato %s, %d passi", gerr, stateG, len(logG))
				if len(errsO) > 0 || len(errsG) > 0 {
					t.Fatalf("segnalazioni della radio simulata: originale %v, IU0VYT Flash %v", errsO, errsG)
				}
				if stateO != stateG || (herr == nil) != (gerr == nil) {
					t.Fatalf("esito diverso")
				}
				cmpLogs(t, logO, logG)
			})
		}
	}
}
