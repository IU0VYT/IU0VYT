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
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tdhFake(n int) []byte {
	b := fakeARM(n, int64(n))
	b[0], b[1], b[2] = 0x02, 0x01, 0x00
	return b
}

func TestTdhHello(t *testing.T) {
	want, _ := hex.DecodeString("A0EE74710774" + "555555555555555555555555555555555555555555555555555555555555")
	if !bytes.Equal(tdhHello, want) {
		t.Fatalf("presentazione %x", tdhHello)
	}
	// blocchi della cattura pubblicata (somme FE, 6B, A1 dell'ultimo)
	fw := make([]byte, 0x7AF*32+31)
	a, _ := hex.DecodeString("02C3CAE4FBFD7F2402ECAD0276A4E4F53875395185383AF537AF3A053A756AFF")
	b, _ := hex.DecodeString("AD3912008F948040F085383AE5392427F53BE4F537AF3A053A756AFFAD3B1200")
	z, _ := hex.DecodeString("7F0102BC597F0202BC597F0302BC597F2DDFFE227F0C02CE3AB2C622DFFE22")
	copy(fw, a)
	copy(fw[32:], b)
	copy(fw[0x7AF*32:], z)
	for _, c := range []struct {
		n    int
		head string
	}{{0, "A10000FE"}, {1, "A100016B"}, {0x7AF, "A207AFA1"}} {
		if got := hex.EncodeToString(tdhPacket(fw, c.n)[:4]); got != hex.EncodeToString(mustHex(c.head)) {
			t.Errorf("blocco %d: %s invece di %s", c.n, got, c.head)
		}
	}
}

func mustHex(s string) []byte { b, _ := hex.DecodeString(s); return b }

func TestTdhEmu(t *testing.T) {
	oldB, oldQ := tdhBootWait, tdhQuiet
	tdhBootWait = 3 * time.Second
	defer func() { tdhBootWait, tdhQuiet = oldB, oldQ }()
	for _, c := range []struct {
		name  string
		size  int
		setup func(*tdhEmu)
		ok    bool
	}{
		{"normale", 58160, nil, true},
		{"multiplo di 32", 32 * 400, nil, true},
		{"0xA5 a lungo dopo la presentazione", 20000, func(e *tdhEmu) { e.extraA5 = 2 * time.Second }, true},
		{"conferma persa", 20000, func(e *tdhEmu) { e.dropAck = 7 }, false},
		{"radio spenta", 20000, func(e *tdhEmu) { e.powerOff = true }, false},
		{"altra presentazione", 20000, func(e *tdhEmu) { e.hello = []byte("PROGRAM") }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			fw := tdhFake(c.size)
			e := &tdhEmu{hello: tdhHello, state: "wait", window: 500 * time.Millisecond}
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
			sent, err := tdhUpdate(context.Background(), "COM9", fw, noProg, func(s string) { t.Log(s) }, func() { started = true })
			time.Sleep(50 * time.Millisecond)
			e.mu.Lock()
			defer e.mu.Unlock()
			t.Logf("errore %v, %d byte, stato %s, avviato %v, %v", err, sent, e.state, started, e.errs)
			if (err == nil) != c.ok {
				t.Fatal("esito inatteso")
			}
			if c.ok {
				want := append(append([]byte{}, fw...), make([]byte, (32-len(fw)%32)%32)...)
				if e.state != "done" || !bytes.Equal(e.data, want) || sent != int64(len(fw)) || e.late || len(e.errs) > 0 {
					t.Fatalf("firmware ricevuto diverso (stato %s, in ritardo %v)", e.state, e.late)
				}
			}
			if c.name == "altra presentazione" && started {
				t.Fatal("invio avviato senza la presentazione accettata")
			}
		})
	}
}

func TestTdhAnalyze(t *testing.T) {
	fw := tdhFake(58160)
	for _, c := range []struct{ name, confirm string }{
		{"730Step improvement.BIN", "RT730"}, {"RT730_DTMF_20240704.BIN", "RT730"},
		{"H3_240530.BIN", "TDH3"}, {"137AB_H3_240730.BIN", "TDH3"}, {"TD-H8_231031.bin", "TDH8"}, {"firmware.bin", "TDH"},
	} {
		if sniffKind(c.name, fw) != fwTDH {
			t.Fatalf("%s: non riconosciuto", c.name)
		}
		in := analyzeTDH(c.name, c.name, fw).Info
		t.Logf("%s → %s (%s) %s %s", c.name, in.Target, in.Confirm, in.Version, in.Date)
		if !in.Valid || in.Confirm != c.confirm || !in.Typed {
			t.Errorf("%s: %s valido %v", c.name, in.Confirm, in.Valid)
		}
	}
	if sniffKind("x.bin", []byte("PROGRAMMA")) == fwTDH {
		t.Error("file qualsiasi riconosciuto come firmware TD-H8")
	}
	bad := tdhFake(20000)
	bad[1], bad[2] = 0xFF, 0xFF
	if sniffKind("x.bin", bad) == fwTDH {
		t.Error("salto iniziale fuori dal file accettato")
	}
}

// TestTdhReal: file veri (TDH_FW=cartella con .BIN di Radtel e TIDRADIO)
func TestTdhReal(t *testing.T) {
	d := os.Getenv("TDH_FW")
	if d == "" {
		t.Skip("TDH_FW non impostato")
	}
	fs, _ := filepath.Glob(filepath.Join(d, "*"))
	n := 0
	for _, f := range fs {
		b, err := os.ReadFile(f)
		if err != nil || sniffKind(filepath.Base(f), b) != fwTDH {
			continue
		}
		in := analyzeTDH(filepath.Base(f), f, b).Info
		t.Logf("%s: %s, %s, %s, conosciuto %q", filepath.Base(f), in.Target, in.Date, fmtBytes(in.Size), in.Known)
		if !in.Valid {
			t.Errorf("%s non valido", f)
		}
		e := &tdhEmu{hello: tdhHello, window: 500 * time.Millisecond}
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			a, p := newPipe()
			go e.serve(p, baud)
			return a, nil
		}
		_, err = tdhUpdate(context.Background(), "COM9", b, noProg, func(string) {}, func() {})
		openPortFunc = old
		time.Sleep(30 * time.Millisecond)
		if err != nil || !bytes.Equal(e.data[:len(b)], b) {
			t.Errorf("%s: invio %v", f, err)
		}
		n++
	}
	if n == 0 {
		t.Error("nessun firmware trovato")
	}
}
