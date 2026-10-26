//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove dell'aggiornamento del firmware Baofeng (KDH V1.0 .BF e 5RH Pro .dat)
// con le radio simulate; i test TestDiffKdh / TestDiffRh5 confrontano i byte
// con quelli del codice originale (vedi verifica/LEGGIMI_VERIFICA.txt).

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeBF: file .BF sintetico (tabella dei vettori plausibile, area 2 di config.)
func fakeBF(s1, s2 int) []byte {
	r := rand.New(rand.NewSource(int64(s1*7 + s2)))
	b := make([]byte, 16+s1+s2)
	b[0] = 2
	b[1], b[2], b[3], b[4] = byte(s1>>24), byte(s1>>16), byte(s1>>8), byte(s1)
	b[5], b[6], b[7], b[8] = byte(s2>>24), byte(s2>>16), byte(s2>>8), byte(s2)
	for i := 9; i < 16; i++ {
		b[i] = byte(r.Intn(256))
	}
	r.Read(b[16:])
	// vettori: SP in RAM, reset e gestori nella flash dopo il caricatore
	v := []uint32{0x20003de0, 0x080010e1}
	for i := 2; i < 16; i++ {
		v = append(v, 0x08001101+uint32(i)*4)
	}
	for i, x := range v {
		b[16+i*4], b[17+i*4], b[18+i*4], b[19+i*4] = byte(x), byte(x>>8), byte(x>>16), byte(x>>24)
	}
	return b
}

func fake5RH(size int, v2 bool) []byte {
	b := make([]byte, rh5Head+size)
	copy(b, "BaoFeng")
	for i := 7; i < 16; i++ {
		b[i] = 0xFF
	}
	copy(b[16:], "BF_5RH")
	for i := 22; i < 32; i++ {
		b[i] = 0xFF
	}
	copy(b[32:], "V1.0.0.0")
	for i := 40; i < 48; i++ {
		b[i] = 0xFF
	}
	if v2 {
		copy(b[48:], "V2.0.0.0")
		for i := 56; i < 64; i++ {
			b[i] = 0xFF
		}
	}
	b[0x40], b[0x41], b[0x42], b[0x43] = byte(size>>24), byte(size>>16), byte(size>>8), byte(size)
	rand.New(rand.NewSource(int64(size))).Read(b[rh5Head:])
	return b
}

func TestKdhEmu(t *testing.T) {
	type sc struct {
		name   string
		raw    []byte
		manual bool
		setup  func(*kdhEmu)
		ok     bool
	}
	scs := []sc{
		{"automatico", fakeBF(59720, 1540), false, nil, true},
		{"manuale", fakeBF(59720, 1540), true, nil, true},
		{"aree multiple di 1024", fakeBF(4096, 2048), false, nil, true},
		{"risposta persa", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.drop[5] = true }, true},
		{"errore di controllo", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.nak[6] = true }, true},
		{"CRC errato", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.corrupt[7] = true }, true},
		{"presentazione persa", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.hsDrop = 2 }, true},
		{"radio spenta", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.silent = true }, false},
		{"errore di scrittura", fakeBF(20000, 1000), true, func(e *kdhEmu) { e.reject = 0xE4 }, false},
	}
	for _, s := range scs {
		t.Run(s.name, func(t *testing.T) {
			e := newKdhEmu(s.raw, s.manual)
			if s.setup != nil {
				s.setup(e)
			}
			old := openPortFunc
			openPortFunc = kdhEmuPorts(e)
			defer func() { openPortFunc = old }()
			st := false
			t0 := time.Now()
			sent, err := kdhUpdate(context.Background(), "COM9", s.raw, kdhOpts{Manual: s.manual}, noProg, func(m string) { t.Log(m) }, func() { st = true })
			time.Sleep(50 * time.Millisecond)
			log, errs, state := e.snapshot()
			t.Logf("errore %v, %d byte, stato %s, %d pacchetti, %v, avviato %v", err, sent, state, len(log), time.Since(t0).Round(time.Millisecond), st)
			if len(errs) > 0 {
				t.Fatalf("segnalazioni della radio: %v", errs)
			}
			if (err == nil) != s.ok || (s.ok && state != "done") {
				t.Fatalf("esito inatteso")
			}
		})
	}
}

func TestRh5Emu(t *testing.T) {
	for _, s := range []struct {
		name     string
		raw      []byte
		radioV2  bool
		setup    func(*rh5Emu)
		ok       bool
		baud     int
		finalSt  string
		wantText string
	}{
		{"V1", fake5RH(458752, false), false, nil, true, 115200, "aggiornata", ""},
		{"V2 38400", fake5RH(401408+300, true), true, nil, true, 38400, "aggiornata", ""},
		{"elettronica diversa", fake5RH(401408, true), false, nil, false, 115200, "idle", "elettronica V1"},
		{"blocco rifiutato", fake5RH(50000, false), false, func(e *rh5Emu) { e.nak = 7 }, false, 115200, "cancellata", ""},
		{"nessuna risposta", fake5RH(50000, false), false, func(e *rh5Emu) { e.mute = 3 }, false, 115200, "cancellata", ""},
	} {
		t.Run(s.name, func(t *testing.T) {
			e := newRh5Emu(s.raw, s.radioV2)
			if s.setup != nil {
				s.setup(e)
			}
			old := openPortFunc
			openPortFunc = rh5EmuPorts(e)
			defer func() { openPortFunc = old }()
			sent, err := rh5Update(context.Background(), "COM9", s.raw, s.baud, noProg, nil, nil)
			time.Sleep(50 * time.Millisecond)
			log, errs, state := e.snapshot()
			t.Logf("errore %v, %d byte, stato %s, %d passi", err, sent, state, len(log))
			if len(errs) > 0 {
				t.Fatalf("segnalazioni della radio: %v", errs)
			}
			if (err == nil) != s.ok || state != s.finalSt || (s.wantText != "" && (err == nil || !strings.Contains(err.Error(), s.wantText))) {
				t.Fatalf("esito inatteso")
			}
			if s.ok && log[0] != "porta "+fmtInt(s.baud) {
				t.Fatalf("velocità: %s", log[0])
			}
		})
	}
}

func fmtInt(n int) string {
	var b []byte
	if n == 0 {
		return "0"
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// ---------------------------------------------------------------- confronto con il codice originale

// TestDiffKdh: BF_HARNESS=verifica/bf/harness/bin/Release/net8.0/bf_harness.dll
// BF_EXE=…/BF_Upgrade_Tool_for_TFT_330L_Series.exe [BF_REAL=cartella con file .BF]
func TestDiffKdh(t *testing.T) {
	harness, exe := os.Getenv("BF_HARNESS"), os.Getenv("BF_EXE")
	if harness == "" || exe == "" {
		t.Skip("BF_HARNESS e BF_EXE non impostati")
	}
	dir := t.TempDir()
	type scen struct {
		name   string
		raw    []byte
		manual bool
		setup  func(*kdhEmu)
	}
	scens := []scen{
		{"automatico", fakeBF(59720, 1540), false, nil},
		{"manuale", fakeBF(59392, 1296), true, nil},
		{"aree multiple di 1024", fakeBF(4096, 2048), false, nil},
		{"risposta persa", fakeBF(20000, 1000), false, func(e *kdhEmu) { e.drop[5] = true; e.drop[9] = true }},
		{"errore di controllo", fakeBF(20000, 1000), true, func(e *kdhEmu) { e.nak[6] = true }},
		{"CRC errato", fakeBF(20000, 1000), true, func(e *kdhEmu) { e.corrupt[4] = true }},
		{"presentazione persa", fakeBF(9000, 700), false, func(e *kdhEmu) { e.hsDrop = 2 }},
		{"errore di scrittura", fakeBF(9000, 700), true, func(e *kdhEmu) { e.reject = 0xE4 }},
		{"radio spenta", fakeBF(9000, 700), false, func(e *kdhEmu) { e.silent = true }},
	}
	if d := os.Getenv("BF_REAL"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, "*.BF"))
		for _, f := range fs {
			b, _ := os.ReadFile(f)
			if analyzeBF(filepath.Base(f), f, b).Info.Valid {
				scens = append(scens, scen{"vero " + filepath.Base(f), b, false, nil})
			}
		}
	}
	for _, s := range scens {
		t.Run(s.name, func(t *testing.T) {
			file := filepath.Join(dir, "fw.BF")
			os.WriteFile(file, s.raw, 0o644)
			mk := func() *kdhEmu {
				e := newKdhEmu(s.raw, s.manual)
				if s.setup != nil {
					s.setup(e)
				}
				return e
			}
			eo := mk()
			addr, stop := serveHarnessFn(func(p Port, baud int) { eo.serve(p, baud) })
			mode := "auto"
			if s.manual {
				mode = "manual"
			}
			out, herr := exec.Command("dotnet", harness, exe, file, addr, mode).CombinedOutput()
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, errsO, stateO := eo.snapshot()

			eg := mk()
			old := openPortFunc
			openPortFunc = kdhEmuPorts(eg)
			_, gerr := kdhUpdate(context.Background(), "COM9", s.raw, kdhOpts{Manual: s.manual}, noProg, nil, nil)
			openPortFunc = old
			time.Sleep(100 * time.Millisecond)
			logG, errsG, stateG := eg.snapshot()
			t.Logf("originale: %s (uscita %v) — stato %s, %d pacchetti", strings.TrimSpace(string(out)), herr, stateO, len(logO))
			t.Logf("IU0VYT Flash: errore %v — stato %s, %d pacchetti", gerr, stateG, len(logG))
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

// TestDiffRh5: RH5_HARNESS=verifica/rh5/harness/bin/Release/net8.0/rh5_harness.dll [RH5_REAL=cartella con file .dat]
func TestDiffRh5(t *testing.T) {
	harness := os.Getenv("RH5_HARNESS")
	if harness == "" {
		t.Skip("RH5_HARNESS non impostato")
	}
	dir := t.TempDir()
	type scen struct {
		name    string
		raw     []byte
		radioV2 bool
		baud    int
		setup   func(*rh5Emu)
	}
	scens := []scen{
		{"V1", fake5RH(458752, false), false, 115200, nil},
		{"V2 38400", fake5RH(401408+300, true), true, 38400, nil},
		{"elettronica diversa", fake5RH(401408, true), false, 115200, nil},
		{"elettronica diversa 2", fake5RH(40140, false), true, 115200, nil},
		{"blocco rifiutato", fake5RH(50000, false), false, 115200, func(e *rh5Emu) { e.nak = 7 }},
	}
	if d := os.Getenv("RH5_REAL"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, "*.dat"))
		for _, f := range fs {
			b, _ := os.ReadFile(f)
			scens = append(scens, scen{"vero " + filepath.Base(f), b, rh5IsV2(b), 115200, nil})
		}
	}
	for _, s := range scens {
		t.Run(s.name, func(t *testing.T) {
			file := filepath.Join(dir, "fw.dat")
			os.WriteFile(file, s.raw, 0o644)
			mk := func() *rh5Emu {
				e := newRh5Emu(s.raw, s.radioV2)
				if s.setup != nil {
					s.setup(e)
				}
				return e
			}
			eo := mk()
			addr, stop := serveHarnessFn(func(p Port, baud int) { eo.serve(p, baud) })
			out, herr := exec.Command("dotnet", harness, file, addr, fmtInt(s.baud)).CombinedOutput()
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, errsO, stateO := eo.snapshot()

			eg := mk()
			old := openPortFunc
			openPortFunc = rh5EmuPorts(eg)
			_, gerr := rh5Update(context.Background(), "COM9", s.raw, s.baud, noProg, nil, nil)
			openPortFunc = old
			time.Sleep(100 * time.Millisecond)
			logG, errsG, stateG := eg.snapshot()
			t.Logf("originale: %s (uscita %v) — stato %s, %d passi", strings.TrimSpace(string(out)), herr, stateO, len(logO))
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

func cmpLogs(t *testing.T, a, b []string) {
	t.Helper()
	if strings.Join(a, "\n") == strings.Join(b, "\n") {
		t.Logf("%d passi identici", len(a))
		return
	}
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			t.Fatalf("passo %d diverso:\noriginale    %.200s\nIU0VYT Flash %.200s", i+1, a[i], b[i])
		}
	}
	t.Fatalf("numero di passi diverso: %d / %d (ultimi: %v / %v)", len(a), len(b), a[max(0, len(a)-3):], b[max(0, len(b)-3):])
}

var _ = bytes.Equal

// serveHarnessFn: radio simulata raggiungibile dal sostituto di SerialPort dei
// programmi di confronto (stesso formato di serveHarness)
func serveHarnessFn(serve func(p Port, baud int)) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		var wmu sync.Mutex
		send := func(d []byte) {
			f := make([]byte, 5+len(d))
			f[0] = 'D'
			binary.LittleEndian.PutUint32(f[1:], uint32(len(d)))
			copy(f[5:], d)
			wmu.Lock()
			c.Write(f)
			wmu.Unlock()
		}
		var host *pipeEnd
		r := bufio.NewReader(c)
		for {
			h := make([]byte, 5)
			if _, err := io.ReadFull(r, h); err != nil {
				if host != nil {
					host.Close()
				}
				return
			}
			d := make([]byte, binary.LittleEndian.Uint32(h[1:]))
			if _, err := io.ReadFull(r, d); err != nil {
				return
			}
			switch h[0] {
			case 'O':
				a, b := newPipe()
				host = a
				go serve(b, int(binary.LittleEndian.Uint32(d)))
				go func(a *pipeEnd) {
					for {
						buf := make([]byte, 1)
						if err := a.ReadFull(context.Background(), buf, time.Hour); err != nil {
							return
						}
						send(buf)
					}
				}(a)
			case 'C':
				if host != nil {
					host.Close()
					host = nil
				}
			case 'W':
				if host != nil {
					host.Write(d)
				}
			}
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

// TestAnalyzeBaofeng: controlli dei file veri (BF_REAL, RH5_REAL) e di quelli sintetici
func TestAnalyzeBaofeng(t *testing.T) {
	show := func(img *fwImage) {
		t.Logf("%s: valido %v, segnalazioni %v, %s %s %s", img.Info.Name, img.Info.Valid, img.Info.Warnings, img.Info.Version, img.Info.Date, img.Info.Hw)
		for _, c := range img.Info.Checks {
			t.Logf("   %s %s", c.Level, c.Text)
		}
	}
	if !analyzeBF("x.BF", "", fakeBF(59720, 1540)).Info.Valid || analyzeBF("x.BF", "", fake5RH(5000, false)).Info.Valid {
		t.Fatal("controllo .BF sintetico")
	}
	if !analyze5RH("x.dat", "", fake5RH(50000, true)).Info.Valid || analyze5RH("x.dat", "", fakeBF(59720, 1540)).Info.Valid {
		t.Fatal("controllo .dat sintetico")
	}
	for _, d := range []string{os.Getenv("BF_REAL"), os.Getenv("RH5_REAL")} {
		if d == "" {
			continue
		}
		fs, _ := filepath.Glob(filepath.Join(d, "*"))
		for _, f := range fs {
			b, _ := os.ReadFile(f)
			switch strings.ToLower(filepath.Ext(f)) {
			case ".bf":
				img := analyzeBF(filepath.Base(f), f, b)
				show(img)
				if !img.Info.Valid && !strings.Contains(f, "dumper") {
					t.Errorf("%s non valido", f)
				}
			case ".dat":
				img := analyze5RH(filepath.Base(f), f, b)
				show(img)
				if !img.Info.Valid {
					t.Errorf("%s non valido", f)
				}
			}
		}
	}
}
