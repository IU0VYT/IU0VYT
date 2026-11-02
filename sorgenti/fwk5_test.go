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
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// firmware .bin di prova (byte casuali, non un vero firmware) con la versione ver
func k5TestFirmware(ver string, size int) ([]byte, []byte) {
	r := rand.New(rand.NewSource(int64(size)))
	fw := make([]byte, size)
	r.Read(fw)
	return k5Pack([]byte(ver), fw), fw
}

func TestK5FirmwarePack(t *testing.T) {
	packed, fw := k5TestFirmware("2.01.26", 0xE3A7)
	if !k5IsPacked(packed) {
		t.Fatal("file .bin non riconosciuto")
	}
	ver, got, ok, err := k5Unpack(packed)
	if err != nil || !ok || !bytes.Equal(got, fw) || k5VerString(ver) != "2.01.26" {
		t.Fatalf("decifratura: %v %v %q", err, ok, ver)
	}
	if sniffKind("k5_v2.01.26_publish.bin", packed) != fwK5 {
		t.Fatal("tipo di firmware non riconosciuto")
	}
	img := analyzeK5("k5_v2.01.26_publish.bin", "prova", packed)
	if !img.Info.Valid || img.Info.Version != "2.01.26" || !bytes.Equal(img.Dec, fw) || img.Info.Target != "Quansheng UV-K5" {
		t.Fatalf("analisi: %+v", img.Info)
	}
	bad := append([]byte{}, packed...)
	bad[100] ^= 1
	if k5IsPacked(bad) || analyzeK5("x.bin", "prova", bad).Info.Valid {
		t.Fatal("file danneggiato accettato")
	}
	if sniffKind("altro.bin", []byte("qualsiasi file .bin")) == fwK5 {
		t.Fatal("file .bin qualsiasi preso per un firmware della UV-K5")
	}
}

func k5Attach(t *testing.T, e *k5BootEmu) {
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	identAttachRaw(func(p Port) { _ = e.serve(p) }, "")
}

// Aggiornamento completo con il caricatore simulato; firmware di un'altra radio
// rifiutato prima di scrivere; radio che smette di rispondere; radio assente.
func TestK5FirmwareUpdate(t *testing.T) {
	ow, ob := k5BootWait, k5BlockWait
	k5BootWait, k5BlockWait = 3*time.Second, 500*time.Millisecond
	t.Cleanup(func() { k5BootWait, k5BlockWait = ow, ob })
	packed, fw := k5TestFirmware("2.01.26", 0xE3A7)
	img := analyzeK5("k5_v2.01.26_publish.bin", "prova", packed)
	noProg := func(int, int, string) {}

	e := newK5BootEmu("2.00.06")
	k5Attach(t, e)
	started := false
	sent, err := k5Update(context.Background(), "COM5", img.Ver, img.Dec, noProg, func(s string) { t.Log(s) }, func() { started = true })
	if err != nil || !started || sent != int64(len(fw)) {
		t.Fatalf("aggiornamento: %v (inviati %d)", err, sent)
	}
	if !bytes.Equal(e.flash[:len(fw)], fw) || e.final != 0xE400 {
		t.Fatalf("memoria della radio diversa dal firmware (fine 0x%04X)", e.final)
	}

	// firmware della UV-K6 (versione 3.x) su una UV-K5: rifiutato, nessun blocco scritto
	p6, _ := k5TestFirmware("3.00.22", 0x9000)
	img6 := analyzeK5("k6_v3.00.22_publish.bin", "prova", p6)
	e2 := newK5BootEmu("2.00.06")
	k5Attach(t, e2)
	started = false
	_, err = k5Update(context.Background(), "COM5", img6.Ver, img6.Dec, noProg, func(string) {}, func() { started = true })
	if err == nil || started || !strings.Contains(err.Error(), "non è per questa radio") || !bytes.Equal(e2.flash[:16], bytes.Repeat([]byte{0xFF}, 16)) {
		t.Fatalf("firmware di un'altra radio: %v (avviato %v)", err, started)
	}
	t.Log(err)

	// radio che non conferma più i blocchi
	e3 := newK5BootEmu("2.00.06")
	e3.stop = 3
	k5Attach(t, e3)
	_, err = k5Update(context.Background(), "COM5", img.Ver, img.Dec, noProg, func(string) {}, func() {})
	if err == nil || !strings.Contains(err.Error(), "0x0300") {
		t.Fatalf("blocco senza conferma: %v", err)
	}

	// nessuna radio nel caricatore
	identAttachRaw(func(p Port) {}, "")
	k5BootWait = 300 * time.Millisecond
	_, err = k5Update(context.Background(), "COM5", img.Ver, img.Dec, noProg, func(string) {}, func() {})
	if err == nil || !strings.Contains(err.Error(), "caricatore") {
		t.Fatalf("senza radio: %v", err)
	}
}

// TestK5FirmwareUvtools (con K5REF_JS = cartella js di uvtools e node): stessa
// decifratura del file e stessi byte inviati alla radio del programma uvtools.
func TestK5FirmwareUvtools(t *testing.T) {
	dir := os.Getenv("K5REF_JS")
	if dir == "" {
		t.Skip("K5REF_JS non impostata")
	}
	ow := k5BootWait
	k5BootWait = 3 * time.Second
	t.Cleanup(func() { k5BootWait = ow })
	for _, c := range []struct {
		ver  string
		size int
	}{{"2.01.26", 0xE3A7}, {"3.00.22", 0xEFFF}, {"4.00.09", 0x2100}} {
		packed, _ := k5TestFirmware(c.ver, c.size)
		f := filepath.Join(t.TempDir(), "fw.bin")
		if err := os.WriteFile(f, packed, 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("node", "verifica/k5/k5ref.js", dir, f).Output()
		if err != nil {
			t.Fatalf("uvtools: %v", err)
		}
		var ref struct {
			Version  string   `json:"version"`
			Firmware string   `json:"firmware"`
			Packets  []string `json:"packets"`
		}
		if err := json.Unmarshal(out, &ref); err != nil {
			t.Fatal(err)
		}
		img := analyzeK5("fw.bin", "prova", packed)
		if hex.EncodeToString(img.Ver) != ref.Version || hex.EncodeToString(img.Dec) != ref.Firmware {
			t.Fatalf("%s: decifratura diversa da uvtools", c.ver)
		}
		e := newK5BootEmu(c.ver[:1] + ".00.06")
		k5Attach(t, e)
		if _, err := k5Update(context.Background(), "COM5", img.Ver, img.Dec, func(int, int, string) {}, func(string) {}, func() {}); err != nil {
			t.Fatal(err)
		}
		want, _ := hex.DecodeString(strings.Join(ref.Packets, ""))
		if !bytes.Equal(e.rx, want) {
			t.Fatalf("%s: byte inviati diversi da uvtools (%d / %d)", c.ver, len(e.rx), len(want))
		}
		t.Logf("versione %s: decifratura e %d pacchetti (%d byte) identici a uvtools", c.ver, len(ref.Packets), len(want))
	}
}
