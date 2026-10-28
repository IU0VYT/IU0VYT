//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prove dell'API via cavo (/api/cable/...) con radio simulate: lettura,
// scrittura con copia di sicurezza e verifica, errori e interruzioni.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type apiClient struct {
	t   *testing.T
	srv *httptest.Server
	app *App
}

func newAPIClient(t *testing.T) *apiClient {
	st, err := openStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := newApp(st)
	srv := httptest.NewServer(a.handler())
	t.Cleanup(srv.Close)
	return &apiClient{t: t, srv: srv, app: a}
}

func (c *apiClient) call(name string, body interface{}) (map[string]interface{}, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", c.srv.URL+"/api/"+name, bytes.NewReader(b))
	req.Header.Set("X-Token", c.app.token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer r.Body.Close()
	var out map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&out)
	if r.StatusCode != 200 {
		return out, fmt.Errorf("%v", out["error"])
	}
	return out, nil
}

func (c *apiClient) must(name string, body interface{}) map[string]interface{} {
	c.t.Helper()
	out, err := c.call(name, body)
	if err != nil {
		c.t.Fatalf("%s: %v", name, err)
	}
	return out
}

// attende la fine dell'operazione; maxStop: se >0 prova a interrompere quando
// la fase indicata è in corso
func (c *apiClient) wait(stopPhase string) (map[string]interface{}, []string) {
	c.t.Helper()
	var stopErrs []string
	for i := 0; i < 3000; i++ {
		st := c.must("cable/status", nil)
		if st["state"] != "running" {
			return st, stopErrs
		}
		if stopPhase != "" && st["phase"] == stopPhase && st["done"].(float64) > 2 {
			if _, err := c.call("cable/cancel", nil); err != nil {
				stopErrs = append(stopErrs, err.Error())
			}
			stopPhase = ""
		}
		time.Sleep(10 * time.Millisecond)
	}
	c.t.Fatal("operazione non terminata")
	return nil, nil
}

func chansJSON(t *testing.T, chs []*Chan) []interface{} {
	b, _ := json.Marshal(chs)
	var out []interface{}
	_ = json.Unmarshal(b, &out)
	return out
}

func TestCableK67(t *testing.T) {
	file := mustRead(t, sampleXLC)
	radio := &k67Emu{mem: xlcImageForRadio(file), model: "5RH-PRO", password: "77"}
	defer withEmu(t, radio.serve)()
	// radio simulata più lenta, per poter tentare l'interruzione durante la scrittura
	fast := openPortFunc
	openPortFunc = func(n string, b int) (Port, error) { p, err := fast(n, b); return verySlowPort{p}, err }
	c := newAPIClient(t)

	// password errata: errore chiaro, nessuna sessione
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM3", "baud": 19200, "password": "12"})
	st, _ := c.wait("")
	if st["state"] != "error" || !strings.Contains(st["error"].(string), "assword") {
		t.Fatalf("password errata: %v", st)
	}
	// lettura
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM3", "baud": 19200, "password": "77"})
	st, _ = c.wait("")
	if st["state"] != "done" {
		t.Fatalf("lettura: %v", st)
	}
	res := st["result"].(map[string]interface{})
	got := res["radio"].(map[string]interface{})["channels"].([]interface{})
	want := readXLC(file, &Report{})
	n := 0
	for i, w := range want {
		if (w == nil) != (got[i] == nil) {
			t.Fatalf("canale %d", i+1)
		}
		if w != nil {
			n++
			g := got[i].(map[string]interface{})
			if g["name"] != w.Name || int64(g["rx"].(float64)) != w.Rx {
				t.Fatalf("canale %d: %v / %+v", i+1, g, w)
			}
		}
	}
	bk, _ := res["backup"].(string)
	if _, err := os.Stat(bk); err != nil {
		t.Fatalf("copia della lettura non salvata: %v", err)
	}
	if c.app.store.Settings.K67Port != "COM3" || c.app.store.Settings.K67Pwd != "77" || !c.app.store.Settings.CableBackup || !c.app.store.Settings.CableVerify {
		t.Fatal("impostazioni non ricordate")
	}
	t.Logf("letti %d canali; copia in %s", n, bk)

	// scrittura di canali modificati, con copia di sicurezza e verifica
	want[0].Name = "VIA CAVO"
	want[3] = nil
	want[200] = &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D023N", Name: "NUOVO", Scan: true}
	radio.mem[0x10] ^= 1 // la radio conserva i suoi primi 128 byte
	head := append([]byte{}, radio.mem[:128]...)
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "baud": 19200, "password": "77",
		"backup": true, "verify": true, "channels": chansJSON(t, want)})
	st, stopErrs := c.wait("Scrittura nella radio")
	if st["state"] != "done" || len(stopErrs) != 1 || !strings.Contains(stopErrs[0], "non si può interrompere") {
		t.Fatalf("scrittura: %v %v", st, stopErrs)
	}
	res = st["result"].(map[string]interface{})
	if res["verified"] != true {
		t.Fatalf("verifica: %v", st)
	}
	if b, err := os.ReadFile(res["backup"].(string)); err != nil || !isXLC(b) {
		t.Fatal("copia di sicurezza mancante")
	}
	back := readXLC(xlcFileFromRadio(radio.mem, 0, false, time.Now()), &Report{})
	if back[0].Name != "VIA CAVO" || back[3] != nil || back[200] == nil || back[200].RxTone != "D023N" || !bytes.Equal(radio.mem[:128], head) {
		t.Fatalf("radio dopo la scrittura: %+v %+v %+v", back[0], back[3], back[200])
	}
	t.Logf("note: %v", st["notes"])

	// interruzione durante la lettura: consentita
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "k67", "port": "COM3", "password": "77"})
	st, stopErrs = c.wait("Lettura della radio")
	if st["state"] != "cancelled" || len(stopErrs) != 0 {
		t.Fatalf("interruzione: %v %v", st, stopErrs)
	}
	// nessun canale: rifiutato prima di collegarsi
	if _, err := c.call("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "channels": []interface{}{nil}}); err == nil {
		t.Fatal("scrittura senza canali accettata")
	}
}

func TestCableBajeton(t *testing.T) {
	img := btSampleImage(t)
	radio := newBtEmu(img)
	defer withEmu(t, radio.serve)()
	c := newAPIClient(t)
	// scrittura senza aver aperto nulla: rifiutata
	if _, err := c.call("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM4", "channels": []interface{}{map[string]interface{}{"rx": 145600000, "tx": 145600000}}}); err == nil {
		t.Fatal("scrittura senza file accettata")
	}
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM4"})
	st, _ := c.wait("")
	if st["state"] != "done" {
		t.Fatalf("lettura: %v", st)
	}
	got := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"].([]interface{})
	want, _ := readDAT(mustRead(t, sampleDAT), &Report{})
	for i := range want {
		if (want[i] == nil) != (got[i] == nil) {
			t.Fatalf("canale %d", i)
		}
	}
	want[0].Name = "CAVO"
	want[64*3+5] = &Chan{Rx: 433500000, Tx: 433500000, Name: "PMR", Narrow: true, TxTone: "D754I", Scan: true}
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM4", "backup": true, "verify": true, "channels": chansJSON(t, want)})
	st, _ = c.wait("")
	if st["state"] != "done" || st["result"].(map[string]interface{})["verified"] != true {
		t.Fatalf("scrittura: %v", st)
	}
	dat, err := eepromToDAT(radio.mem[:btImgSize], baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	back, _ := readDAT(dat, &Report{})
	if back[0].Name != "CAVO" || back[64*3+5] == nil || back[64*3+5].TxTone != "D754I" {
		t.Fatalf("radio dopo la scrittura: %+v %+v", back[0], back[64*3+5])
	}
	// porta senza radio
	openPortFunc = func(string, int) (Port, error) { a, _ := newPipe(); return a, nil }
	c.must("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM5"})
	st, _ = c.wait("")
	if st["state"] != "error" || !strings.Contains(st["error"].(string), "non risponde") {
		t.Fatalf("radio spenta: %v", st)
	}
	t.Logf("errore mostrato: %s", st["error"])
}

type verySlowPort struct{ Port }

func (s verySlowPort) Write(b []byte) error {
	time.Sleep(25 * time.Millisecond)
	return s.Port.Write(b)
}

// K67 Pro che si identifica come «BF-K6» (segnalato da IU0VYT con una radio vera,
// 3.6): file aperto da disco e scritto con la copia di sicurezza: la lettura della
// copia ricorda il modello (come flagK6 del programma originale) e la scrittura
// parte. Senza alcuna lettura precedente la scrittura viene rifiutata con un
// messaggio che spiega cosa fare; dopo una lettura parte anche senza copia.
func TestCableK67IdentifiedAsK6(t *testing.T) {
	file := mustRead(t, sampleXLC)
	radio := &k67Emu{mem: xlcImageForRadio(file), model: "BF-K6", band: file[xlcSize-1]}
	defer withEmu(t, radio.serve)()
	c := newAPIClient(t)
	p := t.TempDir() + "/k67.xlc"
	_ = os.WriteFile(p, file, 0o644)
	r := c.must("radio/open", map[string]interface{}{"radio": "k67", "path": p})
	chs := r["channels"].([]interface{})
	chs[0].(map[string]interface{})["name"] = "K6 OK"
	// 1) senza copia di sicurezza e senza letture: rifiutata, con istruzioni
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "channels": chs})
	st, _ := c.wait("")
	if st["state"] != "error" || !strings.Contains(st["error"].(string), "Leggi prima questa radio") {
		t.Fatalf("senza lettura: %v", st)
	}
	// 2) con la copia di sicurezza (impostazione predefinita): scritta
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "backup": true, "verify": true, "channels": chs})
	st, _ = c.wait("")
	if st["state"] != "done" || st["result"].(map[string]interface{})["verified"] != true {
		t.Fatalf("con copia di sicurezza: %v", st)
	}
	if back := readXLC(xlcFileFromRadio(radio.mem, 0, true, time.Now()), &Report{}); back[0] == nil || back[0].Name != "K6 OK" {
		t.Fatalf("canale non scritto: %+v", back[0])
	}
	// 3) riaperto il file, senza copia: il modello letto è ricordato
	c.must("radio/open", map[string]interface{}{"radio": "k67", "path": p})
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "channels": chs})
	if st, _ = c.wait(""); st["state"] != "done" {
		t.Fatalf("dopo una lettura: %v", st)
	}
	// 4) radio diversa (non BF-K6) dopo aver letto un BF-K6: rifiutata
	radio.model = "5RH-PRO"
	c.must("cable/start", map[string]interface{}{"op": "write", "radio": "k67", "port": "COM3", "channels": chs})
	if st, _ = c.wait(""); st["state"] != "error" || !strings.Contains(st["error"].(string), "dell'ultima radio letta") {
		t.Fatalf("modello diverso: %v", st)
	}
}
