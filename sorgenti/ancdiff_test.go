//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Confronto con il codice ORIGINALE del programma di aggiornamento (BinTe ANC
// Bootloader V1.1): il programma verifica/anc/harness carica il suo eseguibile e
// ne fa girare la classe BootHelper collegata via TCP alla stessa radio
// simulata usata per GestioneCanali. Si confrontano i pacchetti ricevuti dalla
// radio (con la velocità a cui sono arrivati) e i dati del firmware.
//
//   ANC_HARNESS=verifica/anc/harness/bin/Release/net8.0/anc_harness.dll \
//   ANC_EXE="…/BinTe ANC Bootloader.exe" [ANC_REAL=…/BJ8300_V0.12_20251112.anc] \
//   go test -mod=vendor -run TestDiffAnc -v .

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// serveHarness: radio simulata raggiungibile dal sostituto di SerialPort
func serveHarness(e *ancEmu) (string, func()) {
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
				go e.serve(b, int(binary.LittleEndian.Uint32(d)))
				go func(a *pipeEnd) {
					for {
						one := make([]byte, 1)
						if err := a.ReadFull(context.Background(), one, time.Hour); err != nil {
							return
						}
						send(one)
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

func TestDiffAnc(t *testing.T) {
	harness, exe := os.Getenv("ANC_HARNESS"), os.Getenv("ANC_EXE")
	if harness == "" || exe == "" {
		t.Skip("ANC_HARNESS e ANC_EXE non impostati")
	}
	dir := t.TempDir()
	type scen struct {
		name  string
		file  string
		limit int64
		o     ancOpts
		setup func(*ancEmu)
	}
	synth := filepath.Join(dir, "BJ8300_sintetico.anc")
	os.WriteFile(synth, ancDecrypt(fakeFirmware(300000, bj8300Platform)), 0o644)
	odd := filepath.Join(dir, "dispari.anc") // lunghezza non multipla di 1024
	os.WriteFile(odd, ancDecrypt(fakeFirmware(77777, bj8300Platform)), 0o644)
	scens := []scen{
		{"update 115200", synth, 300000, ancOpts{Baud: 115200}, nil},
		{"update 57600", synth, 60000, ancOpts{Baud: 57600}, nil},
		{"update saluto", synth, 60000, ancOpts{Baud: 115200}, func(e *ancEmu) { e.loaderHello = true }},
		{"recovery", synth, 60000, ancOpts{Recovery: true, Baud: 115200}, nil},
		{"lunghezza dispari", odd, 0, ancOpts{Baud: 115200}, nil},
		{"risposta persa", synth, 30000, ancOpts{Baud: 115200}, func(e *ancEmu) { e.drop[9] = true; e.drop[12] = true }},
		{"CRC errato", synth, 30000, ancOpts{Baud: 115200}, func(e *ancEmu) { e.corrupt[7] = true }},
		{"cambio di velocità perso", synth, 30000, ancOpts{Baud: 115200}, func(e *ancEmu) { e.drop[2] = true }},
		{"radio spenta", synth, 30000, ancOpts{Baud: 115200}, func(e *ancEmu) { e.silent = true }},
	}
	if p := os.Getenv("ANC_REAL"); p != "" {
		scens = append(scens, scen{"firmware vero", p, 700000, ancOpts{Baud: 115200}, nil})
	}
	for _, s := range scens {
		t.Run(s.name, func(t *testing.T) {
			raw, err := os.ReadFile(s.file)
			if err != nil {
				t.Fatal(err)
			}
			dec := ancDecrypt(raw)
			mk := func() *ancEmu {
				e := newAncEmu(dec, s.o.Recovery)
				e.defaultPlan(s.limit)
				if s.setup != nil {
					s.setup(e)
				}
				return e
			}
			// codice originale
			eo := mk()
			addr, stop := serveHarness(eo)
			mode, bi := "update", "1"
			if s.o.Recovery {
				mode = "recovery"
			}
			if s.o.Baud == 57600 {
				bi = "0"
			}
			cmd := exec.Command("dotnet", harness, exe, s.file, addr, mode, bi)
			out, herr := cmd.CombinedOutput()
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, errsO, stateO, gotO := eo.snapshot()
			// GestioneCanali
			eg := mk()
			old := openPortFunc
			openPortFunc = ancEmuPorts(eg)
			sent, gerr := ancUpdate(context.Background(), "COM9", dec, s.o, noProg, nil, nil)
			openPortFunc = old
			time.Sleep(100 * time.Millisecond)
			logG, errsG, stateG, gotG := eg.snapshot()

			t.Logf("originale: %s (uscita %v) — stato %s, %d byte, %d pacchetti", strings.TrimSpace(string(out)), herr, stateO, gotO, len(logO))
			t.Logf("GestioneCanali: errore %v — stato %s, %d byte (%d inviati), %d pacchetti", gerr, stateG, gotG, sent, len(logG))
			if len(errsO) > 0 || len(errsG) > 0 {
				t.Fatalf("segnalazioni della radio simulata: originale %v, GestioneCanali %v", errsO, errsG)
			}
			if stateO != stateG || gotO != gotG || (herr == nil) != (gerr == nil) {
				t.Fatalf("esito diverso")
			}
			if strings.Join(logO, "\n") != strings.Join(logG, "\n") {
				n := min(len(logO), len(logG))
				for i := 0; i < n; i++ {
					if logO[i] != logG[i] {
						t.Fatalf("pacchetto %d diverso:\noriginale      %.200s\nGestioneCanali %.200s", i+1, logO[i], logG[i])
					}
				}
				t.Fatalf("numero di pacchetti diverso: %d / %d", len(logO), len(logG))
			}
		})
	}
}
