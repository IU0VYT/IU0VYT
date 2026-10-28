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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type countWriter struct {
	http.ResponseWriter
	n int64
}

func (c *countWriter) Write(b []byte) (int, error) {
	n, err := c.ResponseWriter.Write(b)
	c.n += int64(n)
	return n, err
}

func TestCmpVersion(t *testing.T) {
	for _, c := range []struct {
		a, b string
		w    int
	}{{"2.2", "2.2", 0}, {"2.10", "2.9", 1}, {"2.1", "2.2", -1}, {"3", "2.9.9", 1}, {"2.2.0", "2.2", 0}} {
		if g := cmpVersion(c.a, c.b); g != c.w {
			t.Errorf("%s %s: %d", c.a, c.b, g)
		}
	}
}

func TestPackageVersion(t *testing.T) {
	// versione.txt
	z := makeZip(t, map[string][]byte{"IU0VYT Flash/versione.txt": []byte("IU0VYT Flash\nversione=2.3\ncompilazione=2026-11-02T10:15\n"), "IU0VYT Flash/IU0VYT Flash.exe": []byte("MZ")}, false)
	if v, b, err := packageVersion(z); err != nil || v != "2.3" || fmtBuildDate(b) != "02/11/2026 10:15" {
		t.Fatalf("%s %s %v", v, b, err)
	}
	// firma nell'eseguibile
	z = makeZip(t, map[string][]byte{"x/IU0VYT Flash.exe": []byte("MZ....IU0VYT-FLASH-VERSIONE:2.4;....")}, false)
	if v, _, err := packageVersion(z); err != nil || v != "2.4" {
		t.Fatalf("%s %v", v, err)
	}
	// pacchetti fino alla 2.1: riga del LEGGIMI
	z = makeZip(t, map[string][]byte{"GestioneCanali_K67_Bajeton/LEGGIMI.txt": []byte("GESTIONE CANALI\n=====\nVersione 2.1 - un solo file\n"), "GestioneCanali_K67_Bajeton/GestioneCanali.exe": []byte("MZ")}, false)
	if v, _, err := packageVersion(z); err != nil || v != "2.1" {
		t.Fatalf("%s %v", v, err)
	}
	if _, _, err := packageVersion(makeZip(t, map[string][]byte{"a.txt": []byte("x")}, false)); err != errNotFound {
		t.Fatal(err)
	}
	if _, _, err := packageVersion([]byte("<html>")); err == nil {
		t.Fatal("pagina web accettata")
	}
	// pacchetto vero pubblicato (facoltativo): PKG_21=percorso dello zip 2.1
	if p := os.Getenv("PKG_21"); p != "" {
		b, _ := os.ReadFile(p)
		if v, _, err := packageVersion(b); err != nil || v != "2.1" {
			t.Fatalf("pacchetto 2.1: %s %v", v, err)
		}
	}
	// l'eseguibile compilato contiene la firma della versione
	if !strings.Contains(versionMarker, appVersion) {
		t.Fatal(versionMarker)
	}
}

// pacchetto con un eseguibile grande (non compresso), per controllare che la
// verifica legga solo una piccola parte del file
func bigPackage(t *testing.T, version string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	exe := make([]byte, 3<<20)
	rand.New(rand.NewSource(1)).Read(exe)
	f, _ := w.CreateHeader(&zip.FileHeader{Name: "IU0VYT Flash/IU0VYT Flash.exe", Method: zip.Store})
	f.Write(exe)
	for k := 0; k < 300; k++ { // molti file, come i sorgenti
		f, _ = w.Create(fmt.Sprintf("IU0VYT Flash/sorgenti/file%03d.go", k))
		f.Write([]byte("package main\n"))
	}
	f, _ = w.Create("IU0VYT Flash/versione.txt")
	f.Write([]byte("versione=" + version + "\ncompilazione=2030-01-01T08:00\n"))
	w.Close()
	return buf.Bytes()
}

func TestUpdateAPI(t *testing.T) {
	cur := makeZip(t, map[string][]byte{"IU0VYT Flash/versione.txt": []byte("versione=" + appVersion + "\n")}, false)
	newer := bigPackage(t, "9.0")
	old := makeZip(t, map[string][]byte{"G/LEGGIMI.txt": []byte("Versione 2.1 - x\n")}, false)
	var served int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		switch strings.TrimSuffix(r.URL.Path, "/norange") {
		case "/uguale.zip":
			body = cur
		case "/nuova.zip":
			body = newer
		case "/oggetto":
			// come release-assets.githubusercontent.com: «501 Unsupported client
			// range» per le richieste degli ultimi N byte (bytes=-N)
			if strings.HasPrefix(r.Header.Get("Range"), "bytes=-") {
				http.Error(w, "Unsupported client range", http.StatusNotImplemented)
				return
			}
			body = newer
		case "/rifiuta.zip": // server che risponde 501 a qualsiasi richiesta parziale
			if r.Header.Get("Range") != "" {
				http.Error(w, "Not Implemented", http.StatusNotImplemented)
				return
			}
			body = newer
		case "/vecchia.zip":
			body = old
		case "/github.zip": // come GitHub: reindirizzamento all'archivio dei file
			http.Redirect(w, r, "/oggetto?firma=x", http.StatusFound)
			return
		case "/pagina.zip":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<!doctype html><p>Login</p>"))
			return
		case "/privato.zip":
			w.WriteHeader(403)
			return
		default:
			w.WriteHeader(404)
			return
		}
		cw := &countWriter{ResponseWriter: w}
		if strings.HasSuffix(r.URL.Path, "/norange") {
			cw.Write(body)
		} else {
			http.ServeContent(cw, r, "p.zip", time.Time{}, bytes.NewReader(body))
		}
		served += cw.n
	}))
	defer srv.Close()
	defer func() { updateSource = defaultUpdateURL }()
	dir := t.TempDir()
	st, _ := openStore(dir)
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
	info, _ := call("info", nil)
	if info["name"] != "IU0VYT Flash" || info["author"] != `Daniele "IU0VYT" Bonifazi` || !strings.Contains(info["license"].(string), "GPL") {
		t.Fatalf("%v", info)
	}
	// l'indirizzo degli aggiornamenti non è esposto all'interfaccia (dalla 3.5 c'è
	// solo il collegamento alla pagina del repository)
	if b, _ := json.Marshal(info); bytes.Contains(b, []byte(defaultUpdateURL)) || bytes.Contains(b, []byte("releases")) || bytes.Contains(b, []byte(".zip")) {
		t.Fatalf("indirizzo esposto: %s", b)
	}
	if info["repo"] != "https://github.com/IU0VYT/IU0VYT" {
		t.Fatalf("repository: %v", info["repo"])
	}
	if updateSource != defaultUpdateURL || !strings.HasPrefix(defaultUpdateURL, "https://github.com/IU0VYT/") {
		t.Fatal(defaultUpdateURL)
	}
	for path, want := range map[string]float64{"/uguale.zip": 0, "/nuova.zip": 1, "/vecchia.zip": -1, "/github.zip": 1,
		"/uguale.zip/norange": 0, "/nuova.zip/norange": 1, "/rifiuta.zip": 1} {
		updateSource = srv.URL + path
		served = 0
		// l'indirizzo indicato dall'interfaccia viene ignorato
		r, code := call("update/check", map[string]string{"url": srv.URL + "/vecchia.zip"})
		if code != 200 || r["cmp"] != want {
			t.Fatalf("%s: %d %v", path, code, r)
		}
		if want == 1 && !strings.HasSuffix(path, "/norange") && path != "/rifiuta.zip" {
			// verifica senza scaricare il pacchetto: solo indice e versione.txt
			if served > 200<<10 || int64(r["read"].(float64)) != served || int64(r["size"].(float64)) != int64(len(newer)) {
				t.Fatalf("%s: letti %d byte su %d (%v)", path, served, len(newer), r)
			}
		}
	}
	for p, msg := range map[string]string{"/manca.zip": "404", "/privato.zip": "accesso negato", "/pagina.zip": "pagina web"} {
		updateSource = srv.URL + p
		r, code := call("update/check", nil)
		if code != 400 || !strings.Contains(r["error"].(string), msg) {
			t.Fatalf("%s: %v", p, r)
		}
	}
	// nessun download se la versione pubblicata non è più recente
	updateSource = srv.URL + "/uguale.zip"
	call("update/check", nil)
	if r, code := call("update/save", map[string]string{"path": filepath.Join(dir, "x.zip")}); code != 400 || !strings.Contains(r["error"].(string), "nessun aggiornamento") {
		t.Fatalf("%d %v", code, r)
	}
	// salvataggio: il pacchetto si scarica solo ora
	updateSource = srv.URL + "/github.zip"
	call("update/check", nil)
	out := filepath.Join(dir, "IU0VYT Flash 9.0.zip")
	r, code := call("update/save", map[string]string{"path": out})
	if code != 200 || r["path"] != out || r["version"] != "9.0" {
		t.Fatalf("%d %v", code, r)
	}
	if b, _ := os.ReadFile(out); !bytes.Equal(b, newer) {
		t.Fatal("file salvato diverso")
	}
	// apertura: solo collegamenti web/posta e file salvati dal programma
	if _, code := call("open", map[string]string{"url": "file:///etc/passwd"}); code != 400 {
		t.Fatal("collegamento non web accettato")
	}
	if _, code := call("open", map[string]string{"file": "/etc/passwd"}); code != 400 {
		t.Fatal("file qualsiasi accettato")
	}
	if _, code := call("open", map[string]string{"url": appQRZ}); code != 200 {
		t.Fatal("QRZ rifiutato")
	}
	if _, code := call("open", map[string]string{"file": out}); code != 200 {
		t.Fatal("file salvato rifiutato")
	}
}

// 3.9: indirizzo dell'autore mai in chiaro (né nell'interfaccia né nelle
// costanti del programma); scrittura con il programma di posta, copia negli
// appunti, immagine; indirizzo degli aggiornamenti «latest»; cartella dei dati
// dell'utente che usa il programma.
func TestContactAndDataDir(t *testing.T) {
	// l'indirizzo non è scritto neppure qui: se ne confronta l'impronta SHA-256
	want := authorEmail()
	if h := sha256.Sum256([]byte(want)); hex.EncodeToString(h[:]) != "d523bb2624c69a508f02042f68a02109f5900dd12ada55cc8f9da3dc23e1b50a" ||
		strings.Count(want, "@") != 1 {
		t.Fatal("indirizzo decifrato errato")
	}
	src, _ := os.ReadFile("info.go")
	for _, l := range strings.Split(string(src), "\n") {
		if strings.Contains(l, want) && !strings.HasPrefix(strings.TrimSpace(l), "//") {
			t.Fatalf("indirizzo in chiaro nel codice: %s", l)
		}
	}
	if defaultUpdateURL != "https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip" {
		t.Fatalf("indirizzo degli aggiornamenti %s", defaultUpdateURL)
	}
	c := newAPIClient(t)
	info := c.must("info", nil)
	b, _ := json.Marshal(info)
	if bytes.Contains(b, []byte(want)) || bytes.Contains(b, []byte("@")) {
		t.Fatalf("indirizzo esposto nella scheda Info: %s", b)
	}
	// nessun file del pacchetto lo contiene (sorgenti, interfaccia, documenti, verifiche)
	_ = filepath.WalkDir("..", func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() && (d.Name() == "vendor" || d.Name() == "bin" || d.Name() == "obj") {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			if b, _ := os.ReadFile(p); bytes.Contains(b, []byte(want)) {
				t.Errorf("indirizzo in chiaro in %s", p)
			}
		}
		return nil
	})
	ui, _ := os.ReadFile("ui/info.js")
	for _, l := range strings.Split(string(ui), "\n") {
		if strings.Contains(l, "@") && !strings.HasPrefix(strings.TrimSpace(l), "//") && !strings.Contains(l, "'@'") {
			t.Fatalf("riga della scheda Info con un indirizzo: %s", l)
		}
	}
	// programma di posta
	var opened []string
	oldOpen, oldCopy := openExternal, copyText
	defer func() { openExternal, copyText = oldOpen, oldCopy }()
	openExternal = func(target string, sel bool) bool { opened = append(opened, target); return true }
	if r := c.must("contact", map[string]string{"action": "mail"}); r["opened"] != true || len(opened) != 1 ||
		!strings.HasPrefix(opened[0], "mailto:"+want+"?subject=IU0VYT%20Flash%20") {
		t.Fatalf("mailto: %v %v", r, opened)
	}
	openExternal = func(string, bool) bool { return false }
	if r := c.must("contact", map[string]string{"action": "mail"}); r["opened"] != false {
		t.Fatalf("senza programma di posta: %v", r)
	}
	var copied string
	copyText = func(s string) bool { copied = s; return true }
	if r := c.must("contact", map[string]string{"action": "copy"}); r["copied"] != true || copied != want {
		t.Fatalf("appunti: %v %q", r, copied)
	}
	if r := c.must("contact", map[string]string{"action": "show"}); r["user"].(string)+"@"+r["domain"].(string) != want || strings.Contains(r["user"].(string), "@") {
		t.Fatalf("immagine: %v", r)
	}
	// cartella dei dati: quella dell'utente che avvia il programma
	t.Setenv("GESTIONE_CANALI_DIR", "")
	home := t.TempDir()
	t.Setenv("APPDATA", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	if d := appDataDir(); d != filepath.Join(home, appDirName) {
		t.Fatalf("cartella dei dati %s, attesa sotto %s", d, home)
	}
	if s := shortDataDir(filepath.Join(home, appDirName)); s != filepath.Join("%APPDATA%", appDirName) {
		t.Fatalf("forma breve %q", s)
	}
	if s := shortDataDir("/altrove/dati"); s != "" {
		t.Fatalf("forma breve fuori da APPDATA %q", s)
	}
	st, err := openStore(appDataDir())
	if err != nil || st.dir != filepath.Join(home, appDirName) {
		t.Fatalf("archivio aperto in %v (%v)", st, err)
	}
	if _, err := os.Stat(filepath.Join(home, appDirName)); err != nil {
		t.Fatal(err)
	}
}
