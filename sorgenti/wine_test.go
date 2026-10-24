//go:build linux

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prova dell'eseguibile Windows sotto Wine con porte seriali "vere": due
// pseudo-terminali collegati a COM7 (K67 simulato) e COM8 (Bajeton simulata).
// Così si prova anche il codice della porta COM di Windows (serial_windows.go).
// Si esegue solo con WINE_EXE=percorso di GestioneCanali.exe (e DISPLAY attivo).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type ptyPort struct{ f *os.File }

func (p *ptyPort) Write(b []byte) error { _, err := p.f.Write(b); return err }
func (p *ptyPort) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	_ = p.f.SetReadDeadline(time.Now().Add(to))
	_, err := io.ReadFull(p.f, b)
	if err != nil && os.IsTimeout(err) {
		return errTimeout
	}
	return err
}
func (p *ptyPort) Purge() error      { return nil }
func (p *ptyPort) SetBaud(int) error { return nil }
func (p *ptyPort) Close() error      { return nil }

func openPty(t *testing.T) (*os.File, string) {
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}
	_ = unix.SetNonblock(fd, true)
	m := os.NewFile(uintptr(fd), "ptmx") // non bloccante: le scadenze di lettura funzionano
	return m, fmt.Sprintf("/dev/pts/%d", n)
}

func listening() map[string]bool {
	out, _ := os.ReadFile("/proc/net/tcp")
	m := map[string]bool{}
	for _, l := range strings.Split(string(out), "\n")[1:] {
		f := strings.Fields(l)
		if len(f) >= 4 && f[3] == "0A" && strings.HasPrefix(f[1], "0100007F:") {
			var port int
			fmt.Sscanf(f[1][9:], "%X", &port)
			m[fmt.Sprintf("127.0.0.1:%d", port)] = true
		}
	}
	return m
}

func TestWineSerial(t *testing.T) {
	exe := os.Getenv("WINE_EXE")
	if exe == "" {
		t.Skip("WINE_EXE non impostata")
	}
	wine := os.Getenv("WINE")
	if wine == "" {
		wine = "/usr/lib/wine/wine64"
	}
	prefix := os.Getenv("WINEPREFIX")
	if prefix == "" {
		t.Fatal("WINEPREFIX non impostata")
	}
	// radio simulate sui pseudo-terminali
	file := mustRead(t, sampleXLC)
	k67 := &k67Emu{mem: xlcImageForRadio(file), model: "5RH-PRO"}
	bt := newBtEmu(btSampleImage(t))
	for _, x := range []struct {
		com   string
		serve func(Port) error
	}{{"com7", k67.session}, {"com8", bt.session}} {
		m, slave := openPty(t)
		// tiene aperto il lato "slave" (altrimenti il master riceve EIO a ogni chiusura)
		keep, err := os.OpenFile(slave, os.O_RDWR|unix.O_NOCTTY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer keep.Close()
		link := filepath.Join(prefix, "dosdevices", x.com)
		_ = os.Remove(link)
		if err := os.Symlink(slave, link); err != nil {
			t.Fatal(err)
		}
		p := &ptyPort{m}
		serve := x.serve
		go func() {
			for {
				if err := serve(p); err != nil {
					// errore di protocollo o attesa scaduta: si scarta quanto ricevuto e si riparte
					_ = m.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
					_, _ = io.Copy(io.Discard, m)
				}
			}
		}()
	}
	before := listening()
	data := t.TempDir()
	cmd := exec.Command(wine, exe)
	cmd.Env = append(os.Environ(), "GESTIONE_CANALI_TOKEN=provawine", "GESTIONE_CANALI_DIR="+filepath.Join(prefix, "drive_c", "dati"), "WINEDEBUG=-all")
	_ = data
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = exec.Command("wineserver", "-k").Run() }()
	var addr string
	for i := 0; i < 600 && addr == ""; i++ {
		time.Sleep(200 * time.Millisecond)
		for a := range listening() {
			if !before[a] {
				addr = a
			}
		}
	}
	if addr == "" {
		t.Fatal("il programma non è partito")
	}
	call := func(name string, body interface{}) map[string]interface{} {
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "http://"+addr+"/api/"+name, bytes.NewReader(b))
		req.Header.Set("X-Token", "provawine")
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var out map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&out)
		if r.StatusCode != 200 {
			t.Fatalf("%s: %v", name, out)
		}
		return out
	}
	wait := func() map[string]interface{} {
		for i := 0; i < 1200; i++ {
			st := call("cable/status", nil)
			if st["state"] != "running" {
				return st
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatal("operazione non terminata")
		return nil
	}
	ports := call("cable/ports", nil)
	t.Logf("porte viste dal programma Windows: %v", ports["ports"])

	// porta inesistente: messaggio in italiano
	call("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM42"})
	st := wait()
	t.Logf("COM42: %v", st["error"])
	if st["state"] != "error" {
		t.Fatal("porta inesistente accettata")
	}
	// lettura K67 da COM7
	call("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM7"})
	st = wait()
	if st["state"] != "done" {
		t.Fatalf("lettura K67: %v", st)
	}
	got := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"].([]interface{})
	want := readXLC(file, &Report{})
	for i := range want {
		if (want[i] == nil) != (got[i] == nil) {
			t.Fatalf("K67 canale %d diverso", i+1)
		}
	}
	t.Logf("K67 letto via COM7 in %v s", st["seconds"])
	// scrittura con copia e verifica
	want[0].Name = "WINE COM7"
	want[300] = &Chan{Rx: 145600000, Tx: 145000000, TxTone: "D023N", Name: "NUOVO", Scan: true}
	call("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM7", "backup": true, "verify": true, "channels": chansJSON(t, want)})
	st = wait()
	if st["state"] != "done" || st["result"].(map[string]interface{})["verified"] != true {
		t.Fatalf("scrittura K67: %v", st)
	}
	back := readXLC(xlcFileFromRadio(k67.mem, 0, false, time.Now()), &Report{})
	if back[0].Name != "WINE COM7" || back[300] == nil || back[300].TxTone != "D023N" {
		t.Fatal("radio K67 non aggiornata")
	}
	t.Logf("K67 scritto e verificato via COM7 in %v s; note: %v", st["seconds"], st["notes"])
	// Bajeton da COM8
	call("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM8"})
	st = wait()
	if st["state"] != "done" {
		t.Fatalf("lettura Bajeton: %v", st)
	}
	bch, _ := readDAT(mustRead(t, sampleDAT), &Report{})
	bch[1].Name = "WINE COM8"
	call("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM8", "backup": true, "verify": true, "channels": chansJSON(t, bch)})
	st = wait()
	if st["state"] != "done" || st["result"].(map[string]interface{})["verified"] != true {
		t.Fatalf("scrittura Bajeton: %v", st)
	}
	dat, err := eepromToDAT(bt.mem[:btImgSize], baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	bb, _ := readDAT(dat, &Report{})
	if bb[1] == nil || bb[1].Name != "WINE COM8" {
		t.Fatal("radio Bajeton non aggiornata")
	}
	t.Logf("Bajeton letta, scritta e verificata via COM8 (%v s)", st["seconds"])
}

// Aggiornamento del firmware con l'eseguibile Windows sotto Wine: Bajeton
// simulata su COM9 (pseudo-terminale). Le velocità impostate dal programma
// sulla porta (SetCommState) si leggono dal lato "master" del pseudo-terminale.
func TestWineFirmware(t *testing.T) {
	exe := os.Getenv("WINE_EXE")
	if exe == "" {
		t.Skip("WINE_EXE non impostata")
	}
	wine := os.Getenv("WINE")
	if wine == "" {
		wine = "/usr/lib/wine/wine64"
	}
	prefix := os.Getenv("WINEPREFIX")
	if prefix == "" {
		t.Fatal("WINEPREFIX non impostata")
	}
	plain := fakeFirmware(200000, bj8300Platform)
	fwPath := filepath.Join(prefix, "drive_c", "BJ8300_V0.12_prova.anc")
	if err := os.WriteFile(fwPath, ancDecrypt(plain), 0o644); err != nil {
		t.Fatal(err)
	}
	before := listening()
	cmd := exec.Command(wine, exe)
	cmd.Env = append(os.Environ(), "GESTIONE_CANALI_TOKEN=provawine", "GESTIONE_CANALI_DIR="+filepath.Join(prefix, "drive_c", "dati"), "WINEDEBUG=-all")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = exec.Command("wineserver", "-k").Run() }()
	var addr string
	for i := 0; i < 600 && addr == ""; i++ {
		time.Sleep(200 * time.Millisecond)
		for a := range listening() {
			if !before[a] {
				addr = a
			}
		}
	}
	if addr == "" {
		t.Fatal("il programma non è partito")
	}
	call := func(name string, body interface{}) (map[string]interface{}, int) {
		b, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "http://"+addr+"/api/"+name, bytes.NewReader(b))
		req.Header.Set("X-Token", "provawine")
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var out map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&out)
		return out, r.StatusCode
	}
	// i file dell'interfaccia sono nell'eseguibile
	for _, f := range []string{"firmware.js", "info.js"} {
		r, err := http.Get("http://" + addr + "/" + f)
		if err != nil || r.StatusCode != 200 {
			t.Fatalf("%s non servito: %v", f, err)
		}
		r.Body.Close()
	}
	r, code := call("fw/load", map[string]string{"source": "file", "path": fwPath})
	if code != 200 || r["state"] != "ready" {
		t.Fatalf("%d %v", code, r)
	}
	sha := r["firmware"].(map[string]interface{})["sha256"].(string)
	for _, c := range []struct {
		mode   string
		speeds string
	}{{"update", "9600 10000 115200"}, {"recovery", "9600 115200"}} {
		e := newAncEmu(plain, c.mode == "recovery")
		e.defaultPlan(60000)
		e.anyBaud = true
		// pseudo-terminale nuovo per ogni prova, collegato a COM9
		m, slave := openPty(t)
		keep, err := os.OpenFile(slave, os.O_RDWR|unix.O_NOCTTY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer keep.Close()
		link := filepath.Join(prefix, "dosdevices", "com9")
		_ = os.Remove(link)
		if err := os.Symlink(slave, link); err != nil {
			t.Fatal(err)
		}
		// velocità impostate dal programma sulla porta
		var spMu sync.Mutex
		var speeds []uint32
		stop := make(chan struct{})
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if tio, err := unix.IoctlGetTermios(int(m.Fd()), unix.TCGETS2); err == nil {
					spMu.Lock()
					if len(speeds) == 0 || speeds[len(speeds)-1] != tio.Ospeed {
						speeds = append(speeds, tio.Ospeed)
					}
					spMu.Unlock()
				}
				time.Sleep(500 * time.Microsecond)
			}
		}()
		p := &ptyPort{m}
		go func() {
			for {
				if err := e.serve(p, 0); err == io.EOF {
					return
				}
			}
		}()
		r, code := call("fw/start", map[string]interface{}{"port": "COM9", "mode": c.mode, "baud": 115200, "skipIdentify": true,
			"sha256": sha, "confirmModel": true, "confirmText": "BJ8300"})
		if code != 200 {
			t.Fatalf("%d %v", code, r)
		}
		var st map[string]interface{}
		for i := 0; i < 600; i++ {
			st, _ = call("cable/status", nil)
			if st["state"] != "running" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		_, errs, state, got := e.snapshot()
		spMu.Lock()
		// prima dell'apertura il pseudo-terminale è a 38400 (valore iniziale); Wine imposta le
		// velocità non standard (10000) come 38400 più un divisore personalizzato (TIOCSSERIAL),
		// che un pseudo-terminale non conserva: dopo il primo 9600, 38400 vale 10000
		var sp []string
		seen9600 := false
		for _, v := range speeds {
			s := fmt.Sprint(v)
			if v == 9600 {
				seen9600 = true
			}
			if !seen9600 {
				continue
			}
			if v == 38400 {
				s = "10000"
			}
			if len(sp) == 0 || sp[len(sp)-1] != s {
				sp = append(sp, s)
			}
		}
		all := fmt.Sprint(speeds)
		spMu.Unlock()
		t.Logf("%s: stato %v, radio %s, %d byte, velocità %v (tutte: %s)", c.mode, st["state"], state, got, sp, all)
		if st["state"] != "done" || state != "done" || len(errs) > 0 || got != expectedBytes(e) {
			t.Fatalf("%v %v", st, errs)
		}
		if strings.Join(sp, " ") != c.speeds {
			t.Fatalf("velocità %v invece di %s", sp, c.speeds)
		}
		close(stop)
		time.Sleep(300 * time.Millisecond)
	}
}
