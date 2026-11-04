//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.4: firmware della Retevis RA89

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func thFake(n int) []byte {
	b := make([]byte, n)
	rand.New(rand.NewSource(int64(n))).Read(b)
	return b
}

// thFirmware: il file vero (RETEVIS_FW = cartella) o uno di prova della stessa dimensione
func thFirmware(t *testing.T) (string, []byte) {
	if d := os.Getenv("RETEVIS_FW"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, "*RA89*.bin"))
		if len(fs) > 0 {
			b, err := os.ReadFile(fs[0])
			if err == nil {
				return filepath.Base(fs[0]), b
			}
		}
	}
	return "Eng_RA89_Upadat_20250524_B91.38.bin", thFake(107552)
}

func TestTH8600Analyze(t *testing.T) {
	name, raw := thFirmware(t)
	if sniffKind(name, raw) != fwTH8600 {
		t.Fatalf("%s non riconosciuto", name)
	}
	in := analyzeTH8600(name, name, raw).Info
	t.Logf("%s: %s %s %s", name, in.Target, in.Version, in.Date)
	if !in.Valid || in.Confirm != "RA89" || !in.Typed || in.Version != "B91.38" || in.Date != "24/05/2025" {
		t.Fatalf("analisi: %+v", in)
	}
	if analyzeTH8600("RA89.bin", "x", make([]byte, 100000)).Info.Valid {
		t.Fatal("file vuoto accettato")
	}
	if sniffKind("firmware.bin", raw) == fwTH8600 {
		t.Fatal("file senza il nome della RA89 riconosciuto come RA89")
	}
}

func TestTH8600Emu(t *testing.T) {
	oldD, oldM, oldW := thDelay, thSleepMul, thAckWait
	thDelay, thSleepMul, thAckWait = 0, 0, 300*time.Millisecond
	t.Cleanup(func() { thDelay, thSleepMul, thAckWait = oldD, oldM, oldW })
	_, fw := thFirmware(t)
	run := func(e *thEmu) (int64, bool, error) {
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			x, y := newPipe()
			go e.serve(y, baud)
			return x, nil
		}
		defer func() { openPortFunc = old }()
		started := false
		n, err := thUpdate(context.Background(), "COM9", fw, thHelloRA89, noProg, func(string) {}, func() { started = true })
		time.Sleep(30 * time.Millisecond)
		return n, started, err
	}
	e := newThEmu(thHelloRA89)
	n, _, err := run(e)
	log, errs, st, data := e.snapshot()
	if err != nil || len(errs) > 0 || st != "done" || n != int64(len(fw)) || !bytes.Equal(data, fw) {
		t.Fatalf("invio: %v, %v, %s, %d byte, %d passi", err, errs, st, n, len(log))
	}
	t.Logf("invio completo: %d blocchi", e.blocks)
	for _, c := range []struct {
		name    string
		set     func(*thEmu)
		started bool
	}{
		{"radio di un altro modello", func(e *thEmu) { e.wrong = true }, false},
		{"errore nel blocco 10", func(e *thEmu) { e.nakAt = 10 }, true},
		{"nessuna risposta al blocco 20", func(e *thEmu) { e.quiet = 20 }, true},
	} {
		e := newThEmu(thHelloRA89)
		c.set(e)
		_, started, err := run(e)
		t.Logf("%s: %v", c.name, err)
		if err == nil || started != c.started {
			t.Fatalf("%s: esito inatteso (avviato %v)", c.name, started)
		}
	}
}

// TestDiffTH8600: stesso dialogo del programma originale «RA89-update-software»
// fatto girare con verifica/ra89/harness (RA89_HARNESS = ra89_harness.dll,
// RA89_EXE = RA89-update-software.exe)
func TestDiffTH8600(t *testing.T) {
	h, exe := os.Getenv("RA89_HARNESS"), os.Getenv("RA89_EXE")
	if h == "" || exe == "" {
		t.Skip("RA89_HARNESS / RA89_EXE non impostati")
	}
	name, fw := thFirmware(t)
	dir := t.TempDir()
	file := filepath.Join(dir, name)
	os.WriteFile(file, fw, 0o644)
	for _, baud := range []int{57600, 115200} {
		// righe preparate dal programma: blocchi da 128 byte con la somma
		out, err := exec.Command("dotnet", h, exe, file, "127.0.0.1:1", "3", fmtInt(baud), "lines").CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		var lines []string
		for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			lines = append(lines, strings.TrimPrefix(l, "RIGA "))
		}
		n := (len(fw) + thBlock - 1) / thBlock
		if len(lines) != n+1 || lines[n] != "END" {
			t.Fatalf("%d righe", len(lines))
		}
		for i := 0; i < n; i++ {
			pk := thPacket(fw[i*thBlock : min((i+1)*thBlock, len(fw))])
			if want := strings.ToUpper(hexStr(pk[2:])); lines[i] != want {
				t.Fatalf("riga %d diversa", i)
			}
		}
		eo := newThEmu(thHelloRA89)
		addr, stop := serveHarnessFn(func(p Port, b int) { eo.serve(p, b) })
		t0 := time.Now()
		o, herr := exec.Command("dotnet", h, exe, file, addr, "3", fmtInt(baud)).CombinedOutput()
		stop()
		time.Sleep(100 * time.Millisecond)
		logO, errsO, stO, dataO := eo.snapshot()
		t.Logf("originale %d baud: %s (%v) %v, stato %s", baud, strings.TrimSpace(string(o)), herr, time.Since(t0).Round(time.Second), stO)
		eg := newThEmu(thHelloRA89)
		old, oldB, oldD := openPortFunc, thBaud, thDelay
		openPortFunc = func(name string, b int) (Port, error) { x, y := newPipe(); go eg.serve(y, b); return x, nil }
		thBaud = baud
		if baud == 115200 {
			thDelay = 20 * time.Millisecond
		}
		_, gerr := thUpdate(context.Background(), "COM9", fw, thHelloRA89, noProg, func(string) {}, func() {})
		openPortFunc, thBaud, thDelay = old, oldB, oldD
		time.Sleep(100 * time.Millisecond)
		logG, errsG, stG, dataG := eg.snapshot()
		if len(errsO) > 0 || len(errsG) > 0 || herr != nil || gerr != nil || stO != "done" || stG != "done" {
			t.Fatalf("esiti: originale %v %v %s, IU0VYT Flash %v %v %s", herr, errsO, stO, gerr, errsG, stG)
		}
		if !bytes.Equal(dataO, fw) || !bytes.Equal(dataG, fw) {
			t.Fatal("firmware ricevuto diverso")
		}
		cmpLogs(t, logO, logG)
	}
	// errore della radio al blocco 10: entrambi si fermano lì
	eo := newThEmu(thHelloRA89)
	eo.nakAt = 10
	addr, stop := serveHarnessFn(func(p Port, b int) { eo.serve(p, b) })
	o, herr := exec.Command("dotnet", h, exe, file, addr, "3", "57600").CombinedOutput()
	stop()
	time.Sleep(100 * time.Millisecond)
	logO, _, _, _ := eo.snapshot()
	eg := newThEmu(thHelloRA89)
	eg.nakAt = 10
	old := openPortFunc
	openPortFunc = func(name string, b int) (Port, error) { x, y := newPipe(); go eg.serve(y, b); return x, nil }
	_, gerr := thUpdate(context.Background(), "COM9", fw, thHelloRA89, noProg, func(string) {}, func() {})
	openPortFunc = old
	time.Sleep(100 * time.Millisecond)
	logG, _, _, _ := eg.snapshot()
	t.Logf("errore al blocco 10: originale %s (%v); IU0VYT Flash %v", strings.TrimSpace(string(o)), herr, gerr)
	if herr == nil || gerr == nil {
		t.Fatal("errore non segnalato")
	}
	cmpLogs(t, logO, logG)
}

func hexStr(b []byte) string {
	const d = "0123456789abcdef"
	out := make([]byte, 0, 2*len(b))
	for _, c := range b {
		out = append(out, d[c>>4], d[c&15])
	}
	return string(out)
}
