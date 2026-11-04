//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.4: firmware della Retevis RA89; 4.5: Retevis RT85 e risposta alla
// presentazione corretta

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func thFake(n int) []byte {
	b := make([]byte, n)
	rand.New(rand.NewSource(int64(n))).Read(b)
	return b
}

// thHexOf: file Intel HEX con righe da rec byte (come quelli di Retevis)
func thHexOf(d []byte, rec int) []byte {
	var sb strings.Builder
	for a := 0; a < len(d); a += rec {
		l := d[a:min(a+rec, len(d))]
		r := append([]byte{byte(len(l)), byte(a >> 8), byte(a), 0}, l...)
		var s byte
		for _, c := range r {
			s += c
		}
		r = append(r, -s)
		sb.WriteString(":" + strings.ToUpper(hexStr(r)) + "\r\n")
	}
	sb.WriteString(":00000001FF\r\n")
	return []byte(sb.String())
}

// thFirmware: il file vero (RETEVIS_FW = cartella) o uno di prova della stessa
// dimensione; rt85: file .hex della RT85
func thFirmware(t *testing.T, rt85 bool) (string, []byte) {
	pat := "*RA89*.bin"
	if rt85 {
		pat = "*RT85*.hex"
	}
	if d := os.Getenv("RETEVIS_FW"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, pat))
		if len(fs) > 0 {
			b, err := os.ReadFile(fs[0])
			if err == nil {
				return filepath.Base(fs[0]), b
			}
		}
	}
	if rt85 {
		return "RT85-Firmware-V1.53.hex", thHexOf(thFake(65524), 16)
	}
	return "Eng_RA89_Upadat_20250524_B91.38.bin", thFake(107552)
}

// thProfileFor: programma, presentazione e blocchi di un firmware analizzato
func thProfileFor(t *testing.T, name string, raw []byte, power string) (thProfile, byte, [][]byte, []byte) {
	t.Helper()
	img := analyzeTH8600(name, name, raw)
	if !img.Info.Valid {
		t.Fatalf("%s non valido: %+v", name, img.Info.Checks)
	}
	b, err := thFwBlocks(img)
	if err != nil {
		t.Fatal(err)
	}
	var all []byte
	for _, x := range b {
		all = append(all, x...)
	}
	if img.Info.Power {
		if power == "10W" {
			return thRT85, thHelloRT85_10, b, all
		}
		return thRT85, thHelloRT85_5W, b, all
	}
	return thRA89, thHelloRA89, b, all
}

func TestTH8600Analyze(t *testing.T) {
	name, raw := thFirmware(t, false)
	if sniffKind(name, raw) != fwTH8600 {
		t.Fatalf("%s non riconosciuto", name)
	}
	in := analyzeTH8600(name, name, raw).Info
	t.Logf("%s: %s %s %s", name, in.Target, in.Version, in.Date)
	if !in.Valid || in.Confirm != "RA89" || !in.Typed || in.Power || in.Version != "B91.38" || in.Date != "24/05/2025" {
		t.Fatalf("analisi: %+v", in)
	}
	if analyzeTH8600("RA89.bin", "x", make([]byte, 100000)).Info.Valid {
		t.Fatal("file vuoto accettato")
	}
	if sniffKind("firmware.bin", raw) == fwTH8600 {
		t.Fatal("file senza il nome della RA89 riconosciuto come RA89")
	}

	// RT85
	name, raw = thFirmware(t, true)
	if sniffKind(name, raw) != fwTH8600 || !fwWanted(name) {
		t.Fatalf("%s non riconosciuto", name)
	}
	img := analyzeTH8600(name, name, raw)
	in = img.Info
	t.Logf("%s: %s %s, %s", name, in.Target, in.Version, in.Known)
	if !in.Valid || in.Confirm != "RT85" || !in.Typed || !in.Power || in.Version != "1.53" || in.Models[0] != "rt-rt85" {
		t.Fatalf("analisi: %+v", in)
	}
	// blocchi: 4 righe da 16 byte, l'ultimo con le 4 righe finali (l'ultima da 4
	// byte) e la loro somma (riga di fine con indice multiplo di 4)
	b, err := thFwBlocks(img)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1024 || len(b[0]) != 64 || len(b[1023]) != 53 || !bytes.Equal(b[0], img.Dec[:64]) || !bytes.Equal(b[1023][:52], img.Dec[1023*64:]) {
		t.Fatalf("blocchi: %d, %d, %d", len(b), len(b[0]), len(b[len(b)-1]))
	}
	var s byte
	for _, c := range b[1023][:52] {
		s += c
	}
	if b[1023][52] != s {
		t.Fatal("somma nell'ultimo blocco")
	}
	// numero di righe non multiplo di 4: l'ultima riga riceve una sola somma
	d := thFake(16*4*300 + 16*2)
	bl, data, err := thHexBlocks(thHexOf(d, 16))
	if err != nil || len(bl) != 301 || len(bl[300]) != 32 || !bytes.Equal(data, d) {
		t.Fatalf("file di prova: %v, %d blocchi", err, len(bl))
	}
	// file sbagliati
	good := string(thHexOf(thFake(65524), 16))
	for what, bad := range map[string]string{
		"lunghezza di una riga": strings.Replace(good, ":10001000", ":11001000", 1),
		"indirizzo saltato":     strings.Replace(good, ":10002000", ":10003000", 1),
		"indirizzo esteso":      ":020000040800F2\r\n" + good,
		"senza fine":            strings.Replace(good, ":00000001FF\r\n", "", 1),
		"riga vuota":            strings.Replace(good, "\r\n:10002000", "\r\n\r\n:10002000", 1),
		"caratteri sbagliati":   strings.Replace(good, ":10002000", ":1000200G", 1),
	} {
		if _, _, err := thHexBlocks([]byte(bad)); err == nil {
			t.Fatalf("%s: file accettato", what)
		} else {
			t.Logf("%s: %v", what, err)
		}
		if analyzeTH8600("RT85.hex", "x", []byte(bad)).Info.Valid {
			t.Fatalf("%s: analisi valida", what)
		}
	}
	// righe con LF soltanto: come con CR LF
	bl2, _, err := thHexBlocks([]byte(strings.ReplaceAll(good, "\r\n", "\n")))
	bl3, _, _ := thHexBlocks([]byte(good))
	if err != nil || fmt.Sprint(bl2) != fmt.Sprint(bl3) {
		t.Fatal("righe con LF")
	}
	if sniffKind("firmware.hex", raw) == fwTH8600 {
		t.Fatal("file .hex senza il nome della RT85 riconosciuto come RT85")
	}
}

// thRun: invio di IU0VYT Flash alla radio simulata
func thRun(e *thEmu, pr thProfile, hello byte, blocks [][]byte) (int64, bool, error) {
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		x, y := newPipe()
		go e.serve(y, baud)
		return x, nil
	}
	defer func() { openPortFunc = old }()
	started := false
	n, err := thUpdate(context.Background(), "COM9", pr, hello, blocks, noProg, func(string) {}, func() { started = true })
	time.Sleep(30 * time.Millisecond)
	return n, started, err
}

func TestTH8600Emu(t *testing.T) {
	oldM, oldW := thSleepMul, thAckTest
	thSleepMul, thAckTest = 0, 300*time.Millisecond
	t.Cleanup(func() { thSleepMul, thAckTest = oldM, oldW })
	for _, c := range []struct {
		rt85      bool
		power     string
		radioHelo byte
	}{{false, "", thHelloRA89}, {true, "5W", thHelloRT85_5W}, {true, "10W", thHelloRT85_10}} {
		name, raw := thFirmware(t, c.rt85)
		pr, hello, blocks, all := thProfileFor(t, name, raw, c.power)
		e := newThEmu(c.radioHelo)
		n, _, err := thRun(e, pr, hello, blocks)
		log, errs, st, data := e.snapshot()
		if err != nil || len(errs) > 0 || st != "done" || n != int64(len(all)) || !bytes.Equal(data, all) {
			t.Fatalf("%s %s invio: %v, %v, %s, %d byte, %d passi", pr.name, c.power, err, errs, st, n, len(log))
		}
		if log[1] != fmt.Sprintf("AA %02X EF", hello) {
			t.Fatalf("presentazione: %s", log[1])
		}
		t.Logf("%s %s: invio completo, %d blocchi", pr.name, c.power, e.blocks)
		for _, x := range []struct {
			name    string
			set     func(*thEmu)
			started bool
		}{
			{"radio di un altro modello", func(e *thEmu) { e.wrong = true }, false},
			{"errore nel blocco 10", func(e *thEmu) { e.nakAt = 10 }, true},
			{"nessuna risposta al blocco 20", func(e *thEmu) { e.quiet = 20 }, true},
		} {
			e := newThEmu(c.radioHelo)
			x.set(e)
			_, started, err := thRun(e, pr, hello, blocks)
			t.Logf("%s: %v", x.name, err)
			if err == nil || started != x.started {
				t.Fatalf("%s: esito inatteso (avviato %v)", x.name, started)
			}
		}
	}
	// RT85 con la potenza sbagliata: la radio non risponde, nulla viene scritto
	name, raw := thFirmware(t, true)
	pr, _, blocks, _ := thProfileFor(t, name, raw, "10W")
	e := newThEmu(thHelloRT85_5W)
	_, started, err := thRun(e, pr, thHelloRT85_10, blocks)
	_, _, st, data := e.snapshot()
	t.Logf("RT85 5 W con la scelta 10 W: %v", err)
	if err == nil || started || st != "wait" || len(data) > 0 || !strings.Contains(err.Error(), "potenza diversa") {
		t.Fatal("potenza sbagliata non segnalata")
	}
}

// thCmpGaps: stesse pause dei programmi originali (a pochi millisecondi)
func thCmpGaps(t *testing.T, a, b []thGap) {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("pause: %d / %d", len(a), len(b))
	}
	var ba, bb []time.Duration
	worst := time.Duration(0)
	outliers := 0
	for i := range a {
		if a[i].what != b[i].what {
			t.Fatalf("passo %d: %s / %s", i+1, a[i].what, b[i].what)
		}
		if a[i].what == "blocco" && i > 0 && b[i-1].what == "blocco" {
			ba, bb = append(ba, a[i].d), append(bb, b[i].d)
			continue
		}
		diff := a[i].d - b[i].d
		if diff < 0 {
			diff = -diff
		}
		// prima pausa: dopo l'apertura della porta il programma originale (.NET)
		// compila al volo i suoi metodi (fino a qualche decina di millisecondi)
		tol := 15 * time.Millisecond
		if i == 0 {
			tol = 60 * time.Millisecond
		}
		if diff > tol {
			// con la macchina carica (suite completa) una pausa isolata può
			// allungarsi di qualche decina di millisecondi da una delle due parti:
			// se ne accettano al massimo due, entro 100 ms, segnalate
			outliers++
			msg := fmt.Sprintf("pausa prima del passo %d (%s): originale %v, IU0VYT Flash %v", i+1, a[i].what, a[i].d.Round(time.Millisecond), b[i].d.Round(time.Millisecond))
			if outliers > 2 || diff > 100*time.Millisecond {
				t.Fatal(msg)
			}
			t.Logf("%s (ritardo isolato accettato)", msg)
			continue
		}
		worst = max(worst, diff)
	}
	med := func(l []time.Duration) time.Duration {
		if len(l) == 0 {
			return 0
		}
		l = append([]time.Duration{}, l...)
		sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
		return l[len(l)/2]
	}
	ma, mb := med(ba), med(bb)
	t.Logf("pause: differenza massima %v; tra i blocchi (mediana) originale %v, IU0VYT Flash %v", worst.Round(time.Millisecond), ma.Round(100*time.Microsecond), mb.Round(100*time.Microsecond))
	if d := ma - mb; d > 4*time.Millisecond || d < -4*time.Millisecond {
		t.Fatal("pausa tra i blocchi diversa")
	}
}

// TestDiffTH8600: stesso dialogo dei programmi originali «RA89-update-software»
// e «RT85 10W_5W_UpData» fatti girare con verifica/ra89/harness (RA89_HARNESS =
// ra89_harness.dll, RA89_EXE = RA89-update-software.exe, RT85_EXE = RT85
// 10W_5W_UpData.exe)
func TestDiffTH8600(t *testing.T) {
	h := os.Getenv("RA89_HARNESS")
	if h == "" || (os.Getenv("RA89_EXE") == "" && os.Getenv("RT85_EXE") == "") {
		t.Skip("RA89_HARNESS / RA89_EXE / RT85_EXE non impostati")
	}
	type cs struct {
		exe, model, power string
		rt85              bool
		radio             byte
	}
	var cases []cs
	if x := os.Getenv("RA89_EXE"); x != "" {
		cases = append(cases, cs{x, "3", "", false, thHelloRA89})
	}
	if x := os.Getenv("RT85_EXE"); x != "" {
		cases = append(cases, cs{x, "0", "5W", true, thHelloRT85_5W}, cs{x, "1", "10W", true, thHelloRT85_10})
	}
	for _, c := range cases {
		name, raw := thFirmware(t, c.rt85)
		dir := t.TempDir()
		file := filepath.Join(dir, name)
		os.WriteFile(file, raw, 0o644)
		pr, hello, blocks, all := thProfileFor(t, name, raw, c.power)
		if hello != c.radio {
			t.Fatal("presentazione")
		}
		harness := func(args ...string) ([]byte, error) {
			cmd := exec.Command("dotnet", append([]string{h, c.exe, file}, args...)...)
			cmd.Dir = dir // il programma della RT85 scrive Log.ini nella cartella di lavoro
			return cmd.CombinedOutput()
		}
		for _, baud := range []int{57600, 115200} {
			// righe preparate dal programma: per ogni blocco i dati e la loro somma
			out, err := harness("127.0.0.1:1", c.model, fmtInt(baud), "lines")
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			var lines []string
			for _, l := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(l, "RIGA ") {
					lines = append(lines, strings.TrimSpace(strings.TrimPrefix(l, "RIGA ")))
				} else if strings.HasPrefix(l, "RISPOSTA ") {
					t.Logf("%s %s %d: %s", pr.name, c.power, baud, strings.TrimSpace(l))
					if want := fmt.Sprintf("RISPOSTA AA-%02X-EF", hello); !strings.HasPrefix(l, want) {
						t.Fatalf("risposta attesa dal programma: %s invece di %s", l, want)
					}
				}
			}
			n := len(blocks)
			if len(lines) < n+1 || lines[n] != "END" {
				t.Fatalf("%d righe", len(lines))
			}
			for i, b := range blocks {
				pk := thPacket(b)
				if want := hexStr(pk[2:]); !strings.EqualFold(lines[i], want) {
					t.Fatalf("riga %d diversa:\n%s\n%s", i, lines[i], want)
				}
			}
			t.Logf("%s %s: %d righe preparate dal programma identiche", pr.name, c.power, n)
			eo := newThEmu(c.radio)
			eo.latency = 2 * time.Millisecond
			addr, stop := serveHarnessFn(func(p Port, b int) { eo.serve(p, b) })
			t0 := time.Now()
			o, herr := harness(addr, c.model, fmtInt(baud))
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, errsO, stO, dataO := eo.snapshot()
			t.Logf("originale %s %s %d baud: %s (%v) %v, stato %s", pr.name, c.power, baud, strings.TrimSpace(string(o)), herr, time.Since(t0).Round(time.Second), stO)
			eg := newThEmu(c.radio)
			eg.latency = 2 * time.Millisecond
			old, oldB := openPortFunc, thBaud
			openPortFunc = func(name string, b int) (Port, error) { x, y := newPipe(); go eg.serve(y, b); return x, nil }
			thBaud = baud
			t1 := time.Now()
			_, gerr := thUpdate(context.Background(), "COM9", pr, hello, blocks, noProg, func(string) {}, func() {})
			openPortFunc, thBaud = old, oldB
			time.Sleep(100 * time.Millisecond)
			logG, errsG, stG, dataG := eg.snapshot()
			t.Logf("IU0VYT Flash %s %s %d baud: %v", pr.name, c.power, baud, time.Since(t1).Round(time.Second))
			if len(errsO) > 0 || len(errsG) > 0 || herr != nil || gerr != nil || stO != "done" || stG != "done" {
				t.Fatalf("esiti: originale %v %v %s, IU0VYT Flash %v %v %s", herr, errsO, stO, gerr, errsG, stG)
			}
			if !bytes.Equal(dataO, all) || !bytes.Equal(dataG, all) {
				t.Fatal("firmware ricevuto diverso")
			}
			cmpLogs(t, logO, logG)
			thCmpGaps(t, eo.gapList(), eg.gapList())
		}
		// errore della radio al blocco 10 e radio di un altro modello: entrambi si
		// fermano allo stesso punto
		for _, x := range []struct {
			what string
			set  func(*thEmu)
		}{{"errore al blocco 10", func(e *thEmu) { e.nakAt = 10 }}, {"radio di un altro modello", func(e *thEmu) { e.wrong = true }}} {
			eo := newThEmu(c.radio)
			eo.latency = 2 * time.Millisecond
			x.set(eo)
			addr, stop := serveHarnessFn(func(p Port, b int) { eo.serve(p, b) })
			o, herr := harness(addr, c.model, "57600")
			stop()
			time.Sleep(100 * time.Millisecond)
			logO, _, _, _ := eo.snapshot()
			eg := newThEmu(c.radio)
			eg.latency = 2 * time.Millisecond
			x.set(eg)
			_, _, gerr := thRun(eg, pr, hello, blocks)
			logG, _, _, _ := eg.snapshot()
			t.Logf("%s %s %s: originale %s (%v); IU0VYT Flash %v", pr.name, c.power, x.what, strings.TrimSpace(string(o)), herr, gerr)
			if herr == nil || gerr == nil {
				t.Fatal("errore non segnalato")
			}
			cmpLogs(t, logO, logG)
		}
	}
}

func hexStr(b []byte) string {
	const d = "0123456789abcdef"
	out := make([]byte, 0, 2*len(b))
	for _, c := range b {
		out = append(out, d[c>>4], d[c&15])
	}
	return string(out)
}
