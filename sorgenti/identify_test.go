//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove dell'identificazione della radio (3.8):
// 1) ogni modello simulato viene riconosciuto dal tentativo del suo gruppo;
// 2) ricerca completa: K67 Pro, BJ-8300, alcune radio dei driver e radio USB,
//    ognuna messa in fondo all'elenco dei tentativi; nessuna scrittura nella radio;
// 3) nessuna radio collegata: la ricerca finisce con «non trovata» entro i tempi;
// 4) dati DMR (ID, nome, data di programmazione) uguali a quelli di dmrconfig.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func identFast(t *testing.T) {
	f, s, p := identFirstWait, identSilentMax, identPause
	identFirstWait, identSilentMax, identPause = 150*time.Millisecond, 900*time.Millisecond, 50*time.Millisecond
	t.Cleanup(func() { identFirstWait, identSilentMax, identPause = f, s, p })
}

// radio simulata su ogni porta (e sui dispositivi USB del modello)
func identAttach(t *testing.T, serve func(p Port), conn string) {
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		usb := strings.HasPrefix(name, "HID:") || strings.HasPrefix(name, "DFU:")
		if (conn != "") != usb || (conn == "hid" && !strings.HasPrefix(name, "HID:")) || (conn == "dfu" && !strings.HasPrefix(name, "DFU:")) {
			return nil, fmt.Errorf("dispositivo %s non trovato", name)
		}
		a, b := newPipe()
		a.baud = baud
		go serve(b)
		return a, nil
	}
	t.Cleanup(func() { openPortFunc = old })
}

func identWait(c *apiClient) map[string]interface{} {
	c.t.Helper()
	for i := 0; i < 60000; i++ {
		st := c.must("cable/status", nil)
		if st["state"] != "running" {
			return st
		}
		time.Sleep(10 * time.Millisecond)
	}
	c.t.Fatal("ricerca non terminata")
	return nil
}

func identModelBase(t *testing.T, m *genModel) []byte {
	b := uniBase(t, m)
	if b == nil {
		return nil
	}
	mem, _ := splitImg(b)
	return mem
}

func TestIdentifyEachModel(t *testing.T) {
	identFast(t)
	a := newApp(mustStore(t))
	n, other := 0, 0
	for _, m := range genModels {
		mem := identModelBase(t, m)
		if mem == nil {
			t.Logf("%s: nessuna immagine di prova", m.Name)
			continue
		}
		emu := genEmuFor(m, mem)
		port := "COM7"
		if m.Conn != "" {
			port = "USB"
		}
		identPortMu.Lock()
		plan := a.identPlan(&identReq{Port: port})
		identPortMu.Unlock()
		var grp []identProbe
		for _, p := range plan {
			for _, k := range p.kinds {
				if k == m {
					grp = append(grp, p)
				}
			}
		}
		if len(grp) != 1 {
			t.Fatalf("%s: in %d gruppi", m.Name, len(grp))
		}
		func() {
			old := openPortFunc
			defer func() { openPortFunc = old }()
			before := append([]byte{}, emu.image()...)
			identAttachRaw(func(p Port) { emu.serve(p) }, m.Conn)
			j := &cableJob{Notes: []string{}, Log: []string{}}
			identPortMu.Lock()
			res, err := a.identRun(context.Background(), j, &identReq{Port: port}, grp, time.Now(),
				&identResult{Fields: []identField{}, Others: []identOther{}, Answers: []string{}})
			identPortMu.Unlock()
			if err != nil || !res.Found {
				t.Fatalf("%s: non identificata: %v %v", m.Name, err, res.Answers)
			}
			n++
			if res.Kind != m.ID {
				ok := false
				for _, o := range res.Others {
					ok = ok || o.Kind == m.ID
				}
				// immagine di prova della famiglia (di un altro modello): la memoria è di quel modello
				got := genByID[res.Kind]
				// la memoria di prova (spesso quella di CHIRP di un altro modello della
				// famiglia) non è riconosciuta dal modello: va bene un modello che la
				// riconosce
				if !ok && !(identCanMatch(m) && !m.Fam.Match(m, mem) && identCanMatch(got) && !res.Ambiguous) {
					t.Errorf("%s: identificata come %s (altri: %v)", m.Name, res.Name, res.Others)
				} else if !ok {
					t.Logf("%s: la memoria di prova non è riconosciuta da questo modello; riconosciuta da %s", m.Name, res.Name)
				}
				other++
			}
			read := a.ident.sess.Base
			if res.Count != len(packed(genChans(genByID[res.Kind], read))) {
				t.Errorf("%s: %d canali", m.Name, res.Count)
			}
			if b, ok := a.ident.mems[m.ID]; ok && res.Kind != m.ID {
				read = b // proposto tra gli altri modelli: si apre con la sua lettura
			}
			if _, ok := a.ident.mems[m.ID]; ok || res.Kind == m.ID {
				// letta con il driver del modello: stessi canali della radio simulata
				g1, g2 := genChans(m, read), genChans(m, mem)
				for i := range g2 {
					if (g1[i] == nil) != (g2[i] == nil) || g1[i] != nil && *g1[i] != *g2[i] {
						t.Errorf("%s: canale %d letto %+v, nella radio %+v", m.Name, i+1, g1[i], g2[i])
						break
					}
				}
			}
			if !bytes.Equal(emu.image(), before) {
				t.Errorf("%s: la memoria della radio simulata è cambiata", m.Name)
			}
		}()
	}
	t.Logf("%d modelli identificati (%d come modello con la stessa memoria)", n, other)
	if n < 20 {
		t.Fatalf("troppo pochi modelli provati: %d", n)
	}
}

func identAttachRaw(serve func(p Port), conn string) {
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		usb := strings.HasPrefix(name, "HID:") || strings.HasPrefix(name, "DFU:")
		if (conn != "") != usb || (conn == "hid" && !strings.HasPrefix(name, "HID:")) || (conn == "dfu" && !strings.HasPrefix(name, "DFU:")) {
			return old(name, baud)
		}
		a, b := newPipe()
		a.baud = baud
		go serve(b)
		return a, nil
	}
}

func mustStore(t *testing.T) *Store {
	st, err := openStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return st
}

// ricerca completa dall'API, con la radio in fondo all'elenco dei tentativi
func TestIdentifyFullScan(t *testing.T) {
	identFast(t)
	type cas struct {
		name  string
		port  string
		conn  string
		serve func(p Port)
		kind  string
		want  []string // campi attesi (etichetta=valore, valore parziale)
		mem   func() []byte
	}
	k67 := &k67Emu{mem: xlcImageForRadio(baseXLC), model: "BF-K6"}
	btImg, err := datToEEPROM(baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	bt := newBtEmu(btImg)
	cases := []cas{
		{name: "K67 come BF-K6", port: "COM3", serve: func(p Port) { _ = k67.serve(p) }, kind: kindXLC,
			want: []string{"Modello comunicato dalla radio=BF-K6"}},
		{name: "Bajeton BJ-8300", port: "COM4", serve: func(p Port) { _ = bt.serve(p) }, kind: kindDAT,
			want: []string{"Modello comunicato dalla radio=BT-11U", "Versione del firmware=V1.00"}},
	}
	for _, id := range []string{"bf-uv5r", "bf-uv17pro", "bf-888s", "bf-dm32uv", "bj-9000", "bf-dr1801uv", "bf-rd5r", "bf-dm1701", "bf-w31d"} {
		m := genByID[id]
		mem := identModelBase(t, m)
		if mem == nil {
			t.Logf("%s: nessuna immagine di prova", m.Name)
			continue
		}
		emu := genEmuFor(m, mem)
		port := "COM7"
		if m.Conn != "" {
			port = "USB"
		}
		cases = append(cases, cas{name: m.Name, port: port, conn: m.Conn, serve: func(p Port) { emu.serve(p) }, kind: id})
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			c := newAPIClient(t)
			identAttach(t, cs.serve, cs.conn)
			t0 := time.Now()
			c.must("identify/start", map[string]string{"port": cs.port})
			st := identWait(c)
			if st["state"] != "done" {
				t.Fatalf("ricerca: %v %v", st["state"], st["error"])
			}
			res := st["result"].(map[string]interface{})
			if res["found"] != true || res["kind"] != cs.kind {
				t.Fatalf("risultato: %v", res)
			}
			fields := map[string]string{}
			for _, f := range res["fields"].([]interface{}) {
				f := f.(map[string]interface{})
				fields[f["l"].(string)] = f["v"].(string)
			}
			for _, w := range cs.want {
				kv := strings.SplitN(w, "=", 2)
				if !strings.Contains(fields[kv[0]], kv[1]) {
					t.Errorf("campo %q = %q, atteso %q", kv[0], fields[kv[0]], kv[1])
				}
			}
			if res["backup"] == "" {
				t.Errorf("copia della lettura non salvata")
			}
			t.Logf("%s: %v in %v dopo %v tentativi; %v", res["name"], res["type"], time.Since(t0).Round(time.Millisecond), res["tried"], fields)
			// apertura nella scheda: stessi canali della lettura
			r := c.must("identify/open", nil)
			rad := r["radio"].(map[string]interface{})
			if rad["kind"] != cs.kind {
				t.Fatalf("apertura: %v", rad["kind"])
			}
			n := 0
			for _, ch := range rad["channels"].([]interface{}) {
				if ch != nil {
					n++
				}
			}
			if float64(n) != res["count"].(float64) || n == 0 {
				t.Fatalf("canali aperti %d, contati %v", n, res["count"])
			}
		})
	}
	if !bytes.Equal(k67.mem, xlcImageForRadio(baseXLC)) {
		t.Errorf("memoria della K67 simulata cambiata")
	}
	if b2, _ := datToEEPROM(baseDAT); !bytes.Equal(bt.mem[:len(b2)], b2) {
		t.Errorf("memoria della BJ-8300 simulata cambiata")
	}
}

// nessuna radio: porta che non risponde mai, tempi veri
func TestIdentifyNoRadio(t *testing.T) {
	c := newAPIClient(t)
	identAttach(t, func(p Port) {}, "")
	t0 := time.Now()
	c.must("identify/start", map[string]string{"port": "COM5"})
	st := identWait(c)
	d := time.Since(t0)
	if st["state"] != "done" {
		t.Fatalf("ricerca: %v %v", st["state"], st["error"])
	}
	res := st["result"].(map[string]interface{})
	if res["found"] != false || len(res["answers"].([]interface{})) != 0 {
		t.Fatalf("risultato: %v", res)
	}
	t.Logf("nessuna radio: %v tentativi in %v", res["tried"], d.Round(time.Second))
	if d > 3*time.Minute {
		t.Fatalf("ricerca troppo lunga: %v", d)
	}
	// via USB senza dispositivi: subito
	t0 = time.Now()
	c.must("identify/start", map[string]string{"port": "USB"})
	st = identWait(c)
	if res := st["result"].(map[string]interface{}); res["found"] != false || time.Since(t0) > 5*time.Second {
		t.Fatalf("USB: %v in %v", res, time.Since(t0))
	}
}

// ID DMR, nome e data dell'ultima programmazione: gli stessi stampati da dmrconfig
func TestIdentifyDMRInfo(t *testing.T) {
	bin, ex := os.Getenv("DMRCONFIG"), os.Getenv("DMRCONFIG_EX")
	if bin == "" || ex == "" {
		t.Skip("DMRCONFIG o DMRCONFIG_EX non impostati")
	}
	re := regexp.MustCompile(`(?m)^(ID|Name|Last Programmed Date|CPS Software Version): *(.*)$`)
	for _, v := range []struct{ id, conf, img string }{
		{"bf-rd5r", "rd5r-south-bay-area.conf", ""},
		{"bf-dm1801", "dm1801-south-bay-area.conf", ""},
		{"bf-dm1701", "uv380-south-bay-area.conf", ""},
	} {
		m := genByID[v.id]
		dir := t.TempDir()
		mem := dmrconfigImage(t, bin, ex, m, v.conf, dir)
		if mem == nil {
			continue
		}
		out, err := exec.Command(bin, filepath.Join(dir, "device.img")).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: dmrconfig: %v\n%s", m.Name, err, out)
		}
		want := map[string]string{}
		for _, g := range re.FindAllStringSubmatch(string(out), -1) {
			want[g[1]] = strings.TrimSpace(g[2])
		}
		r := &identResult{}
		dmrInfo(m, mem, r)
		got := map[string]string{}
		for _, f := range r.Fields {
			got[f.L] = f.V
		}
		if got["ID DMR"] != want["ID"] || got["Nome della radio"] != want["Name"] || want["ID"] != "2221234" || want["Name"] != "IU0VYT" {
			t.Errorf("%s: ID %q/%q, nome %q/%q", m.Name, got["ID DMR"], want["ID"], got["Nome della radio"], want["Name"])
		}
		if d := want["Last Programmed Date"]; d != "" && got["Ultima programmazione"] != d {
			t.Errorf("%s: data %q, dmrconfig %q", m.Name, got["Ultima programmazione"], d)
		}
		if d := want["CPS Software Version"]; d != "" && !strings.Contains(d, "?") && got["Versione del programma che l'ha scritta"] != d {
			t.Errorf("%s: versione %q, dmrconfig %q", m.Name, got["Versione del programma che l'ha scritta"], d)
		}
		t.Logf("%s: dmrconfig %v; qui %v", m.Name, want, got)
	}
}

// memoria di prova creata da dmrconfig con un suo file di esempio
func dmrconfigImage(t *testing.T, bin, ex string, m *genModel, conf, dir string) []byte {
	var blank []byte
	switch m.Fam.(type) {
	case rdtyFam:
		blank = bytes.Repeat([]byte{0xFF}, rdtyMemSize)
		copy(blank, rdtyo(m).hdr)
	case tytFam:
		blank = bytes.Repeat([]byte{0xFF}, tytMemSize)
	default:
		t.Fatalf("%s: non è una radio di dmrconfig", m.Name)
	}
	if _, err := os.Stat(filepath.Join(ex, conf)); err != nil {
		t.Logf("%s: manca %s", m.Name, conf)
		return nil
	}
	_ = os.WriteFile(filepath.Join(dir, "blank.img"), blank, 0o644)
	// ID e nome della radio impostati (gli esempi non li hanno)
	cb, _ := os.ReadFile(filepath.Join(ex, conf))
	cb = append([]byte("ID: 2221234\nName: IU0VYT\n"), cb...)
	_ = os.WriteFile(filepath.Join(dir, "prova.conf"), cb, 0o644)
	cmd := exec.Command(bin, "-c", "blank.img", "prova.conf")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dmrconfig -c: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "device.img"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
