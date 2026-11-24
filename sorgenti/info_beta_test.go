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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestChooseUpdate(t *testing.T) {
	for _, c := range []struct {
		local, stable, pre, want string
		isPre                    bool
	}{
		{"4.47.0", "4.46.0", "", "", false},
		{"4.47.0", "4.46.0", "4.47.0", "", false},      // la pre-release in uso: niente da proporre
		{"4.47.0", "4.46.0", "4.48.0", "4.48.0", true}, // pre-release più recente
		{"4.47.0", "4.48.0", "", "4.48.0", false},
		{"4.47.0", "4.48.0", "4.49.0", "4.49.0", true}, // entrambe più recenti: il numero più alto
		{"4.47.0", "4.49.0", "4.48.0", "4.49.0", false},
		{"4.47.0", "4.48.0", "4.48.0", "4.48.0", false}, // a parità la stabile
		{"4.47.0", "4.47.0", "4.46.0", "", false},
		{"4.9.0", "4.8.0", "4.10.0", "4.10.0", true}, // numeri, non lettere
	} {
		v, p := chooseUpdate(c.local, c.stable, c.pre)
		if v != c.want || p != c.isPre {
			t.Errorf("%+v: %q %v", c, v, p)
		}
	}
}

func TestLatestPreRelease(t *testing.T) {
	rc := []ghAsset{{Name: rcAssetName}}
	list := []ghRelease{
		{Tag: "V.4.46.0", Assets: []ghAsset{{Name: "IU0VYT.Flash.zip"}}},       // stabile
		{Tag: "V.4.10.0-rc", Prerelease: true, Assets: rc},                     // 4.10 > 4.9
		{Tag: "V.4.9.0-rc", Prerelease: true, Assets: rc},                      //
		{Tag: "V.4.60.0-rc", Prerelease: true, Draft: true, Assets: rc},        // bozza
		{Tag: "V.4.61.0-rc", Prerelease: true, Assets: []ghAsset{{Name: "x"}}}, // senza l'allegato
		{Tag: "V.4.62.0", Prerelease: true, Assets: rc},                        // tag senza -rc
		{Tag: "V.4.63.0-rc/../x", Prerelease: true, Assets: rc},                // tag strano
		{Tag: "V.4.64.0-rc", Assets: rc},                                       // non pre-release
	}
	if v, tag := latestPreRelease(list); v != "4.10.0" || tag != "V.4.10.0-rc" {
		t.Fatalf("%q %q", v, tag)
	}
	// la più alta anche se nell'elenco viene dopo
	list[1], list[2] = list[2], list[1]
	if v, tag := latestPreRelease(list); v != "4.10.0" || tag != "V.4.10.0-rc" {
		t.Fatalf("ordine: %q %q", v, tag)
	}
	if v, _ := latestPreRelease(list[:1]); v != "" {
		t.Fatal(v)
	}
}

func TestUpdateBeta(t *testing.T) {
	if !strings.HasPrefix(defaultReleasesURL, "https://api.github.com/repos/IU0VYT/IU0VYT/") ||
		defaultRcBase != "https://github.com/IU0VYT/IU0VYT/releases/download/" || releasesSource != defaultReleasesURL {
		t.Fatal("indirizzi delle pre-release")
	}
	// versioni relative a quella del programma
	older, newer := "4.0", "99.0"
	var mu sync.Mutex
	stableV := appVersion // versione del pacchetto «stabile»
	rcV := newer          // versione scritta nel pacchetto della pre-release
	releases := []ghRelease{}
	listCode := 200
	stableCode := 200
	listCalls := 0
	pkgs := map[string][]byte{}
	pkg := func(v string) []byte {
		if b, ok := pkgs[v]; ok {
			return b
		}
		b := makeZip(t, map[string][]byte{"IU0VYT Flash/versione.txt": []byte("versione=" + v + "\ncompilazione=2030-01-02T09:00\n")}, false)
		pkgs[v] = b
		return b
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.URL.Path == "/stabile.zip":
			if stableCode != 200 {
				w.WriteHeader(stableCode)
				return
			}
			http.ServeContent(w, r, "s.zip", time.Time{}, bytes.NewReader(pkg(stableV)))
		case r.URL.Path == "/releases":
			listCalls++
			if r.Header.Get("Accept") != "application/vnd.github+json" || r.URL.RawQuery != "" {
				w.WriteHeader(400)
				return
			}
			if listCode != 200 {
				w.WriteHeader(listCode)
				return
			}
			json.NewEncoder(w).Encode(releases)
		case strings.HasPrefix(r.URL.Path, "/dl/") && strings.HasSuffix(r.URL.Path, "/"+rcAssetName):
			http.ServeContent(w, r, "rc.zip", time.Time{}, bytes.NewReader(pkg(rcV)))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	updateSource, releasesSource, rcDownloadBase = srv.URL+"/stabile.zip", srv.URL+"/releases", srv.URL+"/dl/"
	defer func() {
		updateSource, releasesSource, rcDownloadBase = defaultUpdateURL, defaultReleasesURL, defaultRcBase
	}()

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
	set := func(f func()) { mu.Lock(); f(); mu.Unlock() }
	calls := func() int { mu.Lock(); defer mu.Unlock(); return listCalls }
	rcList := func(vs ...string) []ghRelease {
		l := []ghRelease{{Tag: "V." + older, Assets: []ghAsset{{Name: "IU0VYT.Flash.zip"}}}}
		for _, v := range vs {
			l = append(l, ghRelease{Tag: "V." + v + "-rc", Prerelease: true, Assets: []ghAsset{{Name: rcAssetName}}})
		}
		return l
	}
	set(func() { releases = rcList(newer) })

	// 1. chi usa la stabile (stessa versione): nessuna barra, l'elenco non si legge
	r, code := call("update/check", nil)
	if code != 200 || r["beta"] != false || r["pre"] != false || r["cmp"] != 0.0 || calls() != 0 {
		t.Fatalf("stabile: %d %v (elenco letto %d volte)", code, r, calls())
	}
	// 2. stabile più recente: proposta come sempre, l'elenco non si legge
	set(func() { stableV = "99.1" })
	r, _ = call("update/check", nil)
	if r["beta"] != false || r["pre"] != false || r["cmp"] != 1.0 || r["remote"] != "99.1" || calls() != 0 {
		t.Fatalf("stabile più recente: %v", r)
	}
	// 3. copia di prova con una pre-release più recente: proposta la pre-release
	set(func() { stableV = older })
	r, _ = call("update/check", nil)
	if r["beta"] != true || r["pre"] != true || r["cmp"] != 1.0 || r["remote"] != newer || r["stable"] != older ||
		r["remoteBuild"] != "02/01/2030 09:00" || calls() != 1 {
		t.Fatalf("pre-release: %v", r)
	}
	out := filepath.Join(dir, "beta.zip")
	if r, code := call("update/save", map[string]string{"path": out}); code != 200 || r["version"] != newer || r["pre"] != true {
		t.Fatalf("salvataggio: %d %v", code, r)
	}
	if b, _ := os.ReadFile(out); !bytes.Equal(b, pkg(newer)) {
		t.Fatal("salvato un pacchetto diverso dalla pre-release")
	}
	// 4. copia di prova, unica pre-release = quella in uso: solo la barra
	set(func() { releases = rcList(older, appVersion) })
	r, _ = call("update/check", nil)
	if r["beta"] != true || r["pre"] != false || r["cmp"] != -1.0 || r["remote"] != older {
		t.Fatalf("pre-release in uso: %v", r)
	}
	if r, code := call("update/save", map[string]string{"path": out}); code != 400 || !strings.Contains(r["error"].(string), "nessun aggiornamento") {
		t.Fatalf("%d %v", code, r)
	}
	// 5. pacchetto della pre-release con un numero diverso dal tag: ignorata
	set(func() { releases, rcV = rcList(newer), "98.0" })
	r, _ = call("update/check", nil)
	if r["beta"] != true || r["pre"] != false {
		t.Fatalf("pacchetto diverso dal tag: %v", r)
	}
	// 6. elenco non raggiungibile (es. limite di richieste di GitHub): la barra resta
	set(func() { listCode, rcV = 403, newer })
	r, code = call("update/check", nil)
	if code != 200 || r["beta"] != true || r["pre"] != false {
		t.Fatalf("elenco non raggiungibile: %d %v", code, r)
	}
	// 7. la stessa versione diventa stabile: la barra sparisce, l'elenco non si legge più
	set(func() { listCode, stableV = 200, appVersion })
	n := calls()
	r, _ = call("update/check", nil)
	if r["beta"] != false || r["pre"] != false || r["cmp"] != 0.0 || calls() != n {
		t.Fatalf("diventata stabile: %v", r)
	}
	// 8. senza Internet vale l'ultima stabile vista, salvata nelle impostazioni
	set(func() { stableV = older })
	call("update/check", nil)
	if s2, _ := openStore(dir); s2.Get().LastStable != older {
		t.Fatalf("ultima stabile salvata: %q", s2.Get().LastStable)
	}
	set(func() { stableCode = 404 })
	n = calls()
	r, code = call("update/check", nil)
	if code != 200 || r["offline"] != true || r["beta"] != true || r["stable"] != older || r["error"] == nil || calls() != n {
		t.Fatalf("senza Internet, copia di prova: %d %v", code, r)
	}
	setStable := func(v string) { s := app.store.Get(); s.LastStable = v; app.store.SaveSettings(s) }
	setStable(appVersion)
	r, _ = call("update/check", nil)
	if r["offline"] != true || r["beta"] != false {
		t.Fatalf("senza Internet, stabile: %v", r)
	}
	// senza nessuna stabile salvata: l'errore di sempre
	setStable("")
	if r, code := call("update/check", nil); code != 400 || !strings.Contains(r["error"].(string), "404") {
		t.Fatalf("senza stabile salvata: %d %v", code, r)
	}
}
