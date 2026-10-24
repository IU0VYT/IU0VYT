//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// immagine sintetica con la struttura del firmware della BJ-8300 (in chiaro)
func fakeFirmware(size int, platform string) []byte {
	r := rand.New(rand.NewSource(int64(size)))
	b := make([]byte, size)
	r.Read(b)
	copy(b[fwPlatformAt:], platform)
	copy(b[0x24d0:], "\x00PB01_00_0\x04UPDATE_JUMP\x00")
	copy(b[size/2:], "UPDATE_JUMP")
	copy(b[size-16:], "JLUFW\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	return b
}

func TestAncDecrypt(t *testing.T) {
	// 6 blocchi: 0,1 in chiaro; 2 con "RBGI"; 3 in chiaro; 4,5 (ultimi due) in chiaro
	b := bytes.Repeat([]byte{0x11}, 6*1024)
	d := ancDecrypt(b)
	for i := 0; i < 6*1024; i++ {
		want := byte(0x11)
		if i/1024 == 2 {
			want = 0x11 ^ "RBGI"[i%4]
		}
		if d[i] != want {
			t.Fatalf("byte %d: %02x invece di %02x", i, d[i], want)
		}
	}
	// 9 blocchi + resto: blocco 4 "KDHT", blocco 5 "RBGI", 2 e 3... ; byte speciali non toccati
	b = make([]byte, 9*1024+100)
	for i := range b {
		b[i] = byte(i * 7)
	}
	b[4*1024] = 'K' // uguale alla chiave: resta
	b[4*1024+1] = 'D' ^ 0xFF
	d = ancDecrypt(b)
	if d[4*1024] != 'K' || d[4*1024+1] != 'D'^0xFF {
		t.Fatal("byte uguali alla chiave o al complemento modificati")
	}
	if d[4*1024+2] != b[4*1024+2]^'H' && b[4*1024+2] != 0 && b[4*1024+2] != 0xFF {
		t.Fatal("blocco 4 non cifrato con KDHT")
	}
	if !bytes.Equal(d[3*1024:4*1024], b[3*1024:4*1024]) || !bytes.Equal(d[8*1024:], b[8*1024:]) {
		t.Fatal("blocchi in chiaro modificati")
	}
	// operazione simmetrica
	r := make([]byte, 50000)
	rand.New(rand.NewSource(1)).Read(r)
	if !bytes.Equal(ancDecrypt(ancDecrypt(r)), r) {
		t.Fatal("cifratura non simmetrica")
	}
}

func TestAncPack(t *testing.T) {
	if c := crc16x([]byte("123456789")); c != 0x31C3 {
		t.Fatalf("CRC %04x", c)
	}
	p := ancPack([]byte{6})
	if !bytes.Equal(p[:5], []byte{0xAA, 0x55, 1, 0, 6}) || len(p) != 7 {
		t.Fatalf("%x", p)
	}
	c := crc16x(p[:5])
	if p[5] != byte(c) || p[6] != byte(c>>8) {
		t.Fatal("CRC in coda errato")
	}
}

type ancRun struct {
	emu  *ancEmu
	sent int64
	err  error
	log  []string
	strt bool
}

func runAnc(t *testing.T, fw []byte, o ancOpts, setup func(e *ancEmu)) ancRun {
	t.Helper()
	e := newAncEmu(fw, o.Recovery)
	e.defaultPlan(40000)
	if setup != nil {
		setup(e)
	}
	old := openPortFunc
	openPortFunc = ancEmuPorts(e)
	defer func() { openPortFunc = old }()
	var r ancRun
	r.emu = e
	r.sent, r.err = ancUpdate(context.Background(), "COM9", fw, o, noProg, func(s string) { r.log = append(r.log, s) }, func() { r.strt = true })
	time.Sleep(50 * time.Millisecond)
	return r
}

func expectedBytes(e *ancEmu) int64 {
	var n int64
	for _, r := range e.plan {
		w := r.n
		if r.addr+w > int64(len(e.fw)) {
			w = int64(len(e.fw)) - r.addr
		}
		n += w
	}
	return n
}

func bauds(log []string) []string {
	var out []string
	for _, l := range log {
		b := l[:strings.Index(l, ":")]
		if len(out) == 0 || out[len(out)-1] != b {
			out = append(out, b)
		}
	}
	return out
}

func TestAncUpdate(t *testing.T) {
	fw := fakeFirmware(120000, bj8300Platform)
	cases := []struct {
		name  string
		o     ancOpts
		setup func(*ancEmu)
		bauds string
	}{
		{"normale", ancOpts{Baud: 115200}, nil, "9600 10000 115200"},
		{"normale 57600", ancOpts{Baud: 57600}, nil, "9600 10000 57600"},
		{"saluto del caricatore", ancOpts{Baud: 115200}, func(e *ancEmu) { e.loaderHello = true }, "9600 10000 115200"},
		{"recupero", ancOpts{Recovery: true, Baud: 115200}, nil, "9600 115200"},
		{"risposta persa", ancOpts{Baud: 115200}, func(e *ancEmu) { e.drop[8] = true }, "9600 10000 115200"},
		{"CRC errato", ancOpts{Baud: 115200}, func(e *ancEmu) { e.corrupt[10] = true }, "9600 10000 115200"},
		{"comando sconosciuto", ancOpts{Baud: 115200}, func(e *ancEmu) { e.loaderHello, e.unknownCmd = true, true }, "9600 10000 115200"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := runAnc(t, fw, c.o, c.setup)
			log, errs, state, got := r.emu.snapshot()
			if r.err != nil {
				t.Fatalf("errore: %v\n%s\n%s", r.err, strings.Join(r.log, "\n"), strings.Join(log, "\n"))
			}
			if len(errs) > 0 {
				t.Fatalf("la radio simulata segnala: %v", errs)
			}
			if state != "done" || got != expectedBytes(r.emu) || r.sent != got || !r.strt {
				t.Fatalf("stato %s, ricevuti %d inviati %d attesi %d", state, got, r.sent, expectedBytes(r.emu))
			}
			if b := strings.Join(bauds(log), " "); b != c.bauds {
				t.Fatalf("velocità %s invece di %s", b, c.bauds)
			}
		})
	}
	// radio che non risponde: errore senza aver iniziato; tentativi come l'originale
	t.Run("radio spenta", func(t *testing.T) {
		start := time.Now()
		r := runAnc(t, fw, ancOpts{Baud: 115200}, func(e *ancEmu) { e.silent = true })
		log, _, _, _ := r.emu.snapshot()
		if r.err != errTimeout || r.strt || r.sent != 0 {
			t.Fatalf("%v %v %d", r.err, r.strt, r.sent)
		}
		if len(log) != 6 { // primo invio + 5 nuovi tentativi
			t.Fatalf("%d invii: %v", len(log), log)
		}
		if d := time.Since(start); d < 11*time.Second || d > 14*time.Second {
			t.Fatalf("durata %v (attese 6 × 2 s)", d)
		}
	})
}

func TestAnalyzeANC(t *testing.T) {
	plain := fakeFirmware(200000, bj8300Platform)
	raw := ancDecrypt(plain)
	img := analyzeANC("BJ8300_V0.13_20260101.anc", "x", raw)
	in := img.Info
	if !in.Valid || in.Warnings || in.Platform != bj8300Platform || in.Version != "V0.13" || in.Date != "01/01/2026" {
		t.Fatalf("%+v", in)
	}
	if !bytes.Equal(img.Dec, plain) {
		t.Fatal("decifratura")
	}
	// nome e piattaforma diversi: valido ma con segnalazioni
	img = analyzeANC("RT999_V1.0.anc", "x", ancDecrypt(fakeFirmware(200000, "ZZ12AB_soundbox")))
	if !img.Info.Valid || !img.Info.Warnings {
		t.Fatalf("%+v", img.Info)
	}
	// non è un firmware
	img = analyzeANC("BJ8300.anc", "x", bytes.Repeat([]byte{1, 2, 3}, 30000))
	if img.Info.Valid {
		t.Fatal("file non valido accettato")
	}
	// file cifrato due volte (cioè in chiaro): non valido
	img = analyzeANC("BJ8300.anc", "x", plain)
	if img.Info.Valid {
		t.Fatalf("firmware non cifrato accettato: %+v", img.Info.Checks)
	}
	// firmware vero (facoltativo): ANC_REAL=percorso del file .anc
	if p := os.Getenv("ANC_REAL"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		img := analyzeANC(filepath.Base(p), p, b)
		if !img.Info.Valid || img.Info.Warnings || img.Info.Known == "" || img.Info.Platform != bj8300Platform || img.Info.Copies != 6 {
			t.Fatalf("%+v", img.Info)
		}
		t.Logf("%+v", img.Info)
	}
}

func makeZip(t *testing.T, files map[string][]byte, gbk bool) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for n, d := range files {
		name := n
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		if gbk {
			s, _ := simplifiedchinese.GBK.NewEncoder().String(n)
			h.Name = s
			h.NonUTF8 = true
		}
		f, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		f.Write(d)
	}
	w.Close()
	return buf.Bytes()
}

func TestFwExpander(t *testing.T) {
	anc := ancDecrypt(fakeFirmware(100000, bj8300Platform))
	inner := makeZip(t, map[string][]byte{"BJ8300_V0.12_20251112.anc": anc}, false)
	tool := makeZip(t, map[string][]byte{"升级器/BinTe ANC Bootloader.exe": []byte("MZ...")}, true)
	outer := makeZip(t, map[string][]byte{"BJ8300_V0.12_20251112.zip": inner, "BinTalk ANC版本升级器.zip": tool, "leggimi.txt": []byte("x")}, true)
	e := newExpander("pacchetto.zip", "C:/pacchetto.zip", outer)
	if err := e.run(); err != nil || len(e.found) != 1 {
		t.Fatalf("%v %+v", err, e.found)
	}
	if e.found[0].Name != "BJ8300_V0.12_20251112.anc" || !strings.Contains(e.found[0].Where, "BJ8300_V0.12_20251112.zip › BJ8300_V0.12_20251112.anc") {
		t.Fatalf("%+v", e.found[0])
	}
	// firmware dei caricatori più recenti dentro uno .zip
	kz := makeZip(t, map[string][]byte{"Guida/UV25Pro_NRF_401+_V0.23_250522.kdhx": []byte("x"), "Guida/BJ9000_V0.27.Firmware": []byte("y"),
		"Guida/T52GMRS_M8_V02_231004.kdh": []byte("z"), "Guida/guida.pdf": []byte("%PDF")}, false)
	e = newExpander("guida.zip", "guida.zip", kz)
	if err := e.run(); err != nil || len(e.found) != 3 {
		t.Fatalf("%v %+v", err, e.found)
	}
	// .rar (RAR5, anche "solido", e .rar dentro .rar): aperti dal programma
	for _, c := range []struct{ file, where string }{
		{"fw_rar5.rar", "fw_rar5.rar › firmware/BJ8300_V0.12_20251112.zip › BJ8300_V0.12_20251112.anc"},
		{"fw_rar5_nested.rar", "fw_rar5_nested.rar › pacchetto interno.rar › BJ8300_V0.12_20251112.zip › BJ8300_V0.12_20251112.anc"},
	} {
		b, err := os.ReadFile("testdata/" + c.file)
		if err != nil {
			t.Fatal(err)
		}
		e = newExpander(c.file, c.file, b)
		if err := e.run(); err != nil || len(e.found) != 1 || e.found[0].Where != c.where {
			t.Fatalf("%s: %v %+v", c.file, err, e.found)
		}
		// stesso contenuto del file messo nell'archivio
		if h := sha256.Sum256(e.found[0].data); hex.EncodeToString(h[:]) != "2dd5c9e9a2e0c58b62c618751d123727dee65c561c7183b1b68e028a784537dc" {
			t.Fatalf("%s: firmware estratto diverso", c.file)
		}
	}
	// protetti da password (file o intestazioni cifrati)
	for _, f := range []string{"fw_pwd.rar", "fw_hpwd.rar"} {
		b, _ := os.ReadFile("testdata/" + f)
		e = newExpander(f, f, b)
		err := e.run()
		if err == nil && len(e.found) == 0 && len(e.skipped) == 1 && strings.Contains(e.skipped[0], "password") {
			continue
		}
		if err == nil || !strings.Contains(err.Error(), "password") {
			t.Fatalf("%s: %v %+v %v", f, err, e.found, e.skipped)
		}
	}
	// .rar rovinato
	b, _ := os.ReadFile("testdata/fw_rar5.rar")
	e = newExpander("rotto.rar", "rotto.rar", b[:len(b)/2])
	if err := e.run(); err == nil && len(e.found) > 0 {
		t.Fatal("archivio troncato accettato")
	}
	// archivi .rar veri dei costruttori (facoltativo): RAR_SAMPLES=cartella
	if d := os.Getenv("RAR_SAMPLES"); d != "" {
		fs, _ := filepath.Glob(filepath.Join(d, "*.rar"))
		for _, f := range fs {
			b, _ := os.ReadFile(f)
			e = newExpander(filepath.Base(f), f, b)
			err := e.run()
			t.Logf("%s: errore %v, firmware %d, saltati %v", filepath.Base(f), err, len(e.found), e.skipped)
			if err != nil {
				t.Errorf("%s: %v", f, err)
			}
		}
	}
	// file qualsiasi
	e = newExpander("foto.jpg", "foto.jpg", []byte("\xff\xd8\xff"))
	if err := e.run(); err == nil {
		t.Fatal("file non riconosciuto accettato")
	}
}

func TestFwNormalizeURL(t *testing.T) {
	cases := map[string]string{
		"https://drive.google.com/file/d/1AbCdEfGhIjKlMnOp/view?usp=sharing": "https://drive.usercontent.google.com/download?export=download&confirm=t&id=1AbCdEfGhIjKlMnOp",
		"https://drive.google.com/open?id=1AbCdEfGhIjKlMnOp":                 "https://drive.usercontent.google.com/download?export=download&confirm=t&id=1AbCdEfGhIjKlMnOp",
		"https://www.dropbox.com/s/abc/BJ8300.zip?dl=0":                      "https://www.dropbox.com/s/abc/BJ8300.zip?dl=1",
		"https://example.com/fw/BJ8300.rar":                                  "https://example.com/fw/BJ8300.rar",
	}
	for in, want := range cases {
		got, err := fwNormalizeURL(in)
		if err != nil || got != want {
			t.Errorf("%s → %s (%v)", in, got, err)
		}
	}
	for _, bad := range []string{"https://drive.google.com/drive/folders/1J4fUng8bC1uRyRvbDzZNlxdm9puzuANW?usp=sharing", "ftp://x/y", "BJ8300.anc"} {
		if _, err := fwNormalizeURL(bad); err == nil {
			t.Errorf("%s accettato", bad)
		}
	}
}

func TestFwDownload(t *testing.T) {
	anc := ancDecrypt(fakeFirmware(60000, bj8300Platform))
	z := makeZip(t, map[string][]byte{"BJ8300_V0.12_20251112.anc": anc}, false)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fw.zip":
			w.Header().Set("Content-Disposition", `attachment; filename="BJ8300 pacchetto.zip"`)
			w.Write(z)
		case "/pagina":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html>login</html>"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	b, name, err := fwDownload(srv.URL + "/fw.zip")
	if err != nil || name != "BJ8300 pacchetto.zip" || !bytes.Equal(b, z) {
		t.Fatalf("%v %q", err, name)
	}
	if _, _, err := fwDownload(srv.URL + "/pagina"); err == nil || !strings.Contains(err.Error(), "pagina web") {
		t.Fatal(err)
	}
	if _, _, err := fwDownload(srv.URL + "/manca"); err == nil {
		t.Fatal("404 accettato")
	}
}

// percorso completo attraverso le API: firmware da archivio, controlli, aggiornamento
func TestFwAPI(t *testing.T) {
	dir := t.TempDir()
	st, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := newApp(st)
	h := app.handler()
	call := func(name string, body interface{}) (map[string]interface{}, int) {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/"+name, bytes.NewReader(b))
		req.Header.Set("X-Token", app.token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		var out map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &out)
		return out, w.Code
	}
	plain := fakeFirmware(150000, bj8300Platform)
	anc := ancDecrypt(plain)
	zp := filepath.Join(dir, "pacchetto.zip")
	os.WriteFile(zp, makeZip(t, map[string][]byte{
		"a/BJ8300_V0.12_20251112.anc": anc,
		"b/ALTRO_V1.0.anc":            ancDecrypt(fakeFirmware(90000, "XX11_soundbox")),
	}, false), 0o644)
	r, code := call("fw/load", map[string]string{"source": "file", "path": zp})
	if code != 200 || r["state"] != "choose" || len(r["choices"].([]interface{})) != 2 {
		t.Fatalf("%d %v", code, r)
	}
	if r["choices"].([]interface{})[0].(map[string]interface{})["name"] != "BJ8300_V0.12_20251112.anc" {
		t.Fatalf("ordine delle scelte: %v", r["choices"])
	}
	r, code = call("fw/pick", map[string]int{"index": 0})
	if code != 200 || r["state"] != "ready" {
		t.Fatalf("%d %v", code, r)
	}
	fwi := r["firmware"].(map[string]interface{})
	if fwi["valid"] != true || fwi["platform"] != bj8300Platform {
		t.Fatalf("%v", fwi)
	}
	sha := fwi["sha256"].(string)

	// radio simulata: CPS (57600) e caricatore (altre velocità) sulla stessa porta
	bt := newBtEmu(mustBtImage(t))
	e := newAncEmu(plain, false)
	e.defaultPlan(30000)
	old := openPortFunc
	defer func() { openPortFunc = old }()
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		e.mu.Lock()
		active := e.state != "app"
		e.mu.Unlock()
		if baud == 57600 && !active {
			go bt.serve(b)
		} else {
			go e.serve(b, baud)
		}
		return a, nil
	}
	base := map[string]interface{}{"port": "COM4", "mode": "update", "baud": 115200, "backup": true, "sha256": sha, "confirmModel": true}
	// controlli: conferma mancante, recupero senza testo, firmware diverso
	for _, bad := range []map[string]interface{}{{"confirmModel": false}, {"mode": "recovery"}, {"sha256": "abc"}, {"baud": 9600}} {
		req := map[string]interface{}{}
		for k, v := range base {
			req[k] = v
		}
		for k, v := range bad {
			req[k] = v
		}
		if r, code := call("fw/start", req); code != 400 {
			t.Fatalf("%v accettato: %v", bad, r)
		}
	}
	if r, code := call("fw/start", base); code != 200 {
		t.Fatalf("%d %v", code, r)
	}
	var stt map[string]interface{}
	for i := 0; i < 400; i++ {
		stt, _ = call("cable/status", nil)
		if stt["state"] != "running" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if stt["state"] != "done" {
		t.Fatalf("%v", stt)
	}
	res := stt["result"].(map[string]interface{})
	if res["backup"] == nil || res["radioId"] != "BT-11U  V1.00" || int64(res["sent"].(float64)) != expectedBytes(e) {
		t.Fatalf("%v", res)
	}
	if _, errs, state, _ := e.snapshot(); state != "done" || len(errs) > 0 {
		t.Fatalf("%s %v", state, errs)
	}
	if !bt.ended {
		t.Fatal("la sessione di programmazione non è stata chiusa con E")
	}
	// radio diversa (non risponde al protocollo della Bajeton): aggiornamento non avviato
	e2 := newAncEmu(plain, false)
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		go e2.serve(b, baud)
		return a, nil
	}
	base["backup"] = false
	if r, code := call("fw/start", base); code != 200 {
		t.Fatalf("%d %v", code, r)
	}
	for i := 0; i < 400; i++ {
		stt, _ = call("cable/status", nil)
		if stt["state"] != "running" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if stt["state"] != "error" || !strings.Contains(stt["error"].(string), "NON avviato") {
		t.Fatalf("%v", stt)
	}
	e2.mu.Lock()
	n2, st2 := e2.nPkt, e2.state
	e2.mu.Unlock()
	if st2 != "app" || n2 != 0 {
		t.Fatalf("la radio non doveva ricevere pacchetti: %s %d", st2, n2)
	}
}

func mustBtImage(t *testing.T) []byte {
	img, err := datToEEPROM(baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// firmware .anc della BJ-7800: riconosciuto dal nome del file
func TestAnalyzeANC7800(t *testing.T) {
	raw := ancDecrypt(fakeFirmware(300000, "AC69XX_soundbox"))
	img := analyzeANC("BJ7800_CN_EN_V0.17_20251126.anc", "", raw)
	if !img.Info.Valid || img.Info.Target != "Bajeton BJ-7800" || img.Info.Confirm != "BJ7800" || !img.Info.Warnings || img.Info.Version != "V0.17" {
		t.Fatalf("%+v", img.Info)
	}
	img = analyzeANC("BJ8300_V0.12_20251112.anc", "", ancDecrypt(fakeFirmware(300000, bj8300Platform)))
	if !img.Info.Valid || img.Info.Confirm != "BJ8300" {
		t.Fatalf("%+v", img.Info)
	}
}
