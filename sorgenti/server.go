// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Server HTTP locale: l'interfaccia (HTML/JS incorporati nell'eseguibile)
// parla con il programma tramite queste API. Ascolta solo su 127.0.0.1 e
// richiede un codice casuale generato a ogni avvio.

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed ui
var uiFiles embed.FS

const appVersion = "3.5"

var errNeedPath = errors.New("needPath")
var errCancel = errors.New("annullato")

// fileDialog è impostata dalla versione Windows (finestre native).
// Restituisce errCancel se l'utente annulla, errNeedPath se non disponibile.
var fileDialog = func(save bool, title, filter, initial, defExt string) (string, error) { return "", errNeedPath }

// askQuit: chiusura dell'applicazione (impostata dalla versione Windows)
var quitApp = func() { os.Exit(0) }

type radioSession struct {
	Path string
	Base []byte
	Meta map[string]interface{} // metadati del file .img (radio generiche)
	K6   bool                   // dati letti da una radio che si è identificata come BF-K6
}

type App struct {
	store   *Store
	token   string
	mu      sync.Mutex
	dlgMu   sync.Mutex
	radios  map[string]*radioSession
	dirty   bool
	lastImp struct {
		key    string
		sheets []*Sheet
	}
	lastPing time.Time
	job      *cableJob // ultima operazione con la radio via cavo
	fw       fwSession // aggiornamento del firmware
}

func newApp(st *Store) *App {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return &App{store: st, token: hex.EncodeToString(b), radios: map[string]*radioSession{}, lastPing: time.Now()}
}

func (a *App) IsDirty() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.dirty
}

type apiErr struct {
	Error    string `json:"error"`
	NeedPath bool   `json:"needPath,omitempty"`
	Cancel   bool   `json:"cancel,omitempty"`
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch {
	case errors.Is(err, errNeedPath):
		w.WriteHeader(409)
		_ = json.NewEncoder(w).Encode(apiErr{Error: "serve il percorso del file", NeedPath: true})
	case errors.Is(err, errCancel):
		w.WriteHeader(409)
		_ = json.NewEncoder(w).Encode(apiErr{Error: "annullato", Cancel: true})
	default:
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(apiErr{Error: err.Error()})
	}
}

func (a *App) handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(uiFiles, "ui")
	static := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		static.ServeHTTP(w, r)
	})
	api := func(name string, f func(r *http.Request) (interface{}, error)) {
		mux.HandleFunc("/api/"+name, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.Header.Get("X-Token") != a.token {
				w.WriteHeader(403)
				return
			}
			a.mu.Lock()
			a.lastPing = time.Now()
			a.mu.Unlock()
			defer func() {
				if e := recover(); e != nil {
					fail(w, fmt.Errorf("errore interno: %v", e))
				}
			}()
			v, err := f(r)
			if err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, v)
		})
	}
	api("init", a.apiInit)
	api("catalog", a.apiCatalog)
	api("radio/open", a.apiOpen)
	api("radio/new", a.apiNew)
	api("radio/save", a.apiSave)
	api("adapt", a.apiAdapt)
	api("archive/save", a.apiArchiveSave)
	api("import/load", a.apiImport)
	api("export", a.apiExport)
	api("settings", a.apiSettings)
	api("rename", a.apiRename)
	api("dirty", a.apiDirty)
	api("cable/ports", a.apiCablePorts)
	api("cable/start", a.apiCableStart)
	api("cable/status", a.apiCableStatus)
	api("cable/cancel", a.apiCableCancel)
	a.fwHandlers(mux, api)
	a.infoHandlers(api)
	api("ping", func(r *http.Request) (interface{}, error) { return map[string]bool{"ok": true}, nil })
	api("quit", func(r *http.Request) (interface{}, error) {
		go func() { time.Sleep(200 * time.Millisecond); quitApp() }()
		return map[string]bool{"ok": true}, nil
	})
	return mux
}

func decode(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(io.LimitReader(r.Body, 100<<20))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("richiesta non valida: %v", err)
	}
	return nil
}

func (a *App) dialog(save bool, title, filter, initial, defExt string) (string, error) {
	a.dlgMu.Lock()
	defer a.dlgMu.Unlock()
	if initial == "" && a.store.Settings.LastDir != "" {
		initial = a.store.Settings.LastDir + string(os.PathSeparator)
	}
	p, err := fileDialog(save, title, filter, initial, defExt)
	if err == nil && p != "" {
		st := a.store.Settings
		st.LastDir = filepath.Dir(p)
		_ = a.store.SaveSettings(st)
	}
	if err == nil && p == "" {
		return "", errCancel
	}
	return p, err
}

// ---------------------------------------------------------------- init

func (a *App) apiInit(r *http.Request) (interface{}, error) {
	var tones []string
	for _, t := range ctcssList {
		tones = append(tones, fmtTone(t))
	}
	return map[string]interface{}{
		"version":     appVersion,
		"settings":    a.store.Settings,
		"archivePath": a.store.archivePath(),
		"ponti":       a.store.Ponti,
		"tipi":        tipiNoti,
		"ctcss":       tones,
		"nativeDlg":   fileDialogNative,
		"defaultURL":  defaultURL,
		"dcsBaj":      btDCS,
	}, nil
}

// ---------------------------------------------------------------- radio

var radioFilter = map[string]string{
	kindXLC: "Baofeng K67 Pro (*.xlc)|*.xlc|Tutti i file|*.*",
	kindDAT: "Bajeton 8300 (*.dat)|*.dat|Tutti i file|*.*",
	"":      "File radio (*.xlc, *.dat, *.img)|*.xlc;*.dat;*.img|Baofeng K67 Pro (*.xlc)|*.xlc|Bajeton 8300 (*.dat)|*.dat|Immagini CHIRP (*.img)|*.img|Tutti i file|*.*",
}

type radioResp struct {
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Channels []*Chan  `json:"channels"`
	Warnings []string `json:"warnings"`
	Bands    []band   `json:"bands,omitempty"`
}

func (a *App) apiOpen(r *http.Request) (interface{}, error) {
	var req struct {
		Radio string `json:"radio"`
		Path  string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	p := req.Path
	if p == "" {
		var err error
		t := "Apri il file della radio"
		if req.Radio != "" {
			t = "Apri il file del " + radioName(req.Radio)
		}
		filt := radioFilter[req.Radio]
		if isGenKind(req.Radio) {
			filt = imgFilter
		}
		if p, err = a.dialog(false, t, filt, "", ""); err != nil {
			return nil, err
		}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere il file: %v", err)
	}
	if isGenKind(req.Radio) || (req.Radio == "" && detect(b) == "") {
		return a.genOpen(p, b, req.Radio)
	}
	rep := &Report{}
	kind, chs, err := readRadio(b, rep)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.radios[kind] = &radioSession{Path: p, Base: b}
	a.mu.Unlock()
	st := a.store.Settings
	if kind == kindXLC {
		st.K67Path = p
	} else {
		st.BajPath = p
	}
	_ = a.store.SaveSettings(st)
	resp := radioResp{Kind: kind, Path: p, Channels: chs, Warnings: rep.Warnings}
	if kind == kindXLC {
		resp.Bands = xlcBands(b, 0x10)
	}
	return resp, nil
}

func (a *App) apiNew(r *http.Request) (interface{}, error) {
	var req struct {
		Radio string `json:"radio"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	var b []byte
	var err error
	switch req.Radio {
	case kindXLC:
		b = blankXLC(baseXLC)
	case kindDAT:
		if b, err = blankDAT(baseDAT); err != nil {
			return nil, err
		}
	default:
		if isGenKind(req.Radio) {
			return nil, errors.New("per questa radio leggi prima la radio via cavo oppure apri un suo file .img (anche salvato con CHIRP): le impostazioni della radio non si possono creare da zero")
		}
		return nil, errors.New("radio sconosciuta")
	}
	a.mu.Lock()
	a.radios[req.Radio] = &radioSession{Base: b}
	a.mu.Unlock()
	resp := radioResp{Kind: req.Radio, Channels: make([]*Chan, radioSlots(req.Radio))}
	if req.Radio == kindXLC {
		resp.Bands = xlcBands(b, 0x10)
	}
	return resp, nil
}

func (a *App) apiSave(r *http.Request) (interface{}, error) {
	var req struct {
		Radio    string  `json:"radio"`
		Channels []*Chan `json:"channels"`
		SaveAs   bool    `json:"saveAs"`
		Path     string  `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if isGenKind(req.Radio) {
		return a.genSave(req.Radio, req.Channels, req.SaveAs, req.Path)
	}
	if req.Radio != kindXLC && req.Radio != kindDAT {
		return nil, errors.New("radio sconosciuta")
	}
	a.mu.Lock()
	sess := a.radios[req.Radio]
	a.mu.Unlock()
	if sess == nil {
		return nil, errors.New("nessun file aperto per questa radio")
	}
	p := req.Path
	if p == "" && !req.SaveAs {
		p = sess.Path
	}
	if p == "" {
		ext, filt := "xlc", radioFilter[kindXLC]
		suggest := "K67_canali.xlc"
		if req.Radio == kindDAT {
			ext, filt, suggest = "dat", radioFilter[kindDAT], "Bajeton_canali.dat"
		}
		if sess.Path != "" {
			suggest = sess.Path
		} else if a.store.Settings.LastDir != "" {
			suggest = filepath.Join(a.store.Settings.LastDir, suggest)
		}
		var err error
		if p, err = a.dialog(true, "Salva il file del "+radioName(req.Radio), filt, suggest, ext); err != nil {
			return nil, err
		}
	}
	rep := &Report{}
	out, err := writeRadio(req.Radio, sess.Base, req.Channels, rep)
	if err != nil {
		return nil, err
	}
	// controllo: il file scritto deve rileggersi uguale a quanto richiesto
	if _, back, err := readRadio(out, &Report{}); err != nil {
		return nil, fmt.Errorf("errore interno, file non salvato: %v", err)
	} else if err := verifySaved(req.Channels, back); err != nil {
		return nil, fmt.Errorf("errore interno, file non salvato: %v", err)
	}
	if old, err := os.ReadFile(p); err == nil && !bytes.Equal(old, out) {
		_ = os.WriteFile(p+".bak", old, 0o644) // copia di sicurezza della versione precedente
	}
	if err := writeAtomic(p, out); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	a.mu.Lock()
	a.radios[req.Radio] = &radioSession{Path: p, Base: out}
	a.mu.Unlock()
	return map[string]interface{}{"path": p, "warnings": rep.Warnings}, nil
}

// verifySaved confronta i canali richiesti con quelli riletti dal file.
func verifySaved(want, got []*Chan) error {
	for i := 0; i < len(got); i++ {
		var w *Chan
		if i < len(want) {
			w = want[i]
		}
		g := got[i]
		if (w == nil) != (g == nil) {
			if w != nil && (w.Rx <= 0 || w.Tx <= 0 || w.Rx >= 1000000000 || w.Tx >= 1000000000) {
				continue // canale non scrivibile, già segnalato
			}
			return fmt.Errorf("posizione %d non corrisponde", i+1)
		}
		if w == nil {
			continue
		}
		if w.Rx != g.Rx || w.Tx != g.Tx {
			return fmt.Errorf("posizione %d: frequenze non corrispondono", i+1)
		}
	}
	return nil
}

func (a *App) apiAdapt(r *http.Request) (interface{}, error) {
	var req struct {
		To       string  `json:"to"`
		Channels []*Chan `json:"channels"`
		Pos      []int   `json:"pos"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if m, ok := genByID[req.To]; ok {
		rep := &Report{}
		out := make([]*Chan, len(req.Channels))
		for i, c := range req.Channels {
			if c == nil {
				continue
			}
			label := fmt.Sprintf("«%s»", c.Name)
			if i < len(req.Pos) {
				label = posLabel(req.To, req.Pos[i]) + " " + label
			}
			out[i] = genAdapt(c, m, label, rep)
		}
		return map[string]interface{}{"channels": out, "warnings": rep.Warnings}, nil
	}
	var bands []band
	if req.To == kindXLC {
		a.mu.Lock()
		s := a.radios[kindXLC]
		a.mu.Unlock()
		if s != nil {
			bands = xlcBands(s.Base, 0x10)
		} else {
			bands = xlcBands(baseXLC, 0x10)
		}
	}
	rep := &Report{}
	out := make([]*Chan, len(req.Channels))
	for i, c := range req.Channels {
		if c == nil {
			continue
		}
		label := fmt.Sprintf("«%s»", c.Name)
		if i < len(req.Pos) {
			label = posLabel(req.To, req.Pos[i]) + " " + label
		}
		out[i] = adaptChan(c, req.To, bands, label, rep)
	}
	return map[string]interface{}{"channels": out, "warnings": rep.Warnings}, nil
}

// ---------------------------------------------------------------- archivio

func (a *App) apiArchiveSave(r *http.Request) (interface{}, error) {
	var req struct {
		Ponti []*Ponte `json:"ponti"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Ponti == nil {
		req.Ponti = []*Ponte{}
	}
	if err := a.store.SavePonti(req.Ponti); err != nil {
		return nil, fmt.Errorf("impossibile salvare l'archivio: %v", err)
	}
	return map[string]interface{}{"ok": true, "count": len(req.Ponti)}, nil
}

func (a *App) apiSettings(r *http.Request) (interface{}, error) {
	var st Settings
	if err := decode(r, &st); err != nil {
		return nil, err
	}
	cur := a.store.Settings
	cur.URL, cur.URLs, cur.Modello, cur.ShowEmpty = st.URL, st.URLs, st.Modello, st.ShowEmpty
	if err := a.store.SaveSettings(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func (a *App) apiRename(r *http.Request) (interface{}, error) {
	var req struct {
		Ponti   []*Ponte `json:"ponti"`
		Modello string   `json:"modello"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	names := make([]string, len(req.Ponti))
	for i, p := range req.Ponti {
		if p != nil {
			names[i] = nomeDaModello(p, req.Modello)
		}
	}
	return map[string]interface{}{"names": names}, nil
}

func (a *App) apiDirty(r *http.Request) (interface{}, error) {
	var req struct {
		Dirty bool `json:"dirty"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.dirty = req.Dirty
	a.mu.Unlock()
	return map[string]bool{"ok": true}, nil
}

// ---------------------------------------------------------------- importazione

var tableFilter = "Tabelle (*.xls, *.xlsx, *.csv)|*.xls;*.xlsx;*.csv|Excel 97-2003 (*.xls)|*.xls|Excel (*.xlsx)|*.xlsx|CSV (*.csv)|*.csv|Tutti i file|*.*"

var httpClient = &http.Client{Timeout: 90 * time.Second}

func download(url string) ([]byte, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, errors.New("l'indirizzo deve iniziare con http:// o https://")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("indirizzo non valido: %v", err)
	}
	req.Header.Set("User-Agent", "IU0VYT-Flash/"+appVersion)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download non riuscito: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download non riuscito: il server ha risposto %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 60<<20))
	if err != nil {
		return nil, fmt.Errorf("download interrotto: %v", err)
	}
	if bytes.HasPrefix(bytes.TrimSpace(b[:min(len(b), 512)]), []byte("<")) && !bytes.HasPrefix(b, []byte("PK")) {
		return nil, errors.New("l'indirizzo ha restituito una pagina web, non un file .xls/.xlsx/.csv")
	}
	return b, nil
}

func (a *App) apiImport(r *http.Request) (interface{}, error) {
	var req struct {
		Source  string `json:"source"` // file | url | cache
		Path    string `json:"path"`
		URL     string `json:"url"`
		Sheet   string `json:"sheet"`
		Modello string `json:"modello"`
		ToneRx  bool   `json:"toneRx"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	var sheets []*Sheet
	var label string
	switch req.Source {
	case "cache":
		a.mu.Lock()
		sheets, label = a.lastImp.sheets, a.lastImp.key
		a.mu.Unlock()
		if sheets == nil {
			return nil, errors.New("nessun file caricato")
		}
	case "url":
		b, err := download(req.URL)
		if err != nil {
			return nil, err
		}
		label = strings.TrimSpace(req.URL)
		if sheets, err = readTable(b, label); err != nil {
			return nil, err
		}
		st := a.store.Settings
		st.URL = label
		found := false
		for _, u := range st.URLs {
			if u == label {
				found = true
			}
		}
		if !found {
			st.URLs = append([]string{label}, st.URLs...)
			if len(st.URLs) > 10 {
				st.URLs = st.URLs[:10]
			}
		}
		_ = a.store.SaveSettings(st)
	default:
		p := req.Path
		if p == "" {
			var err error
			if p, err = a.dialog(false, "Scegli il file da importare", tableFilter, "", ""); err != nil {
				return nil, err
			}
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("impossibile leggere il file: %v", err)
		}
		label = p
		if sheets, err = readTable(b, p); err != nil {
			return nil, err
		}
	}
	a.mu.Lock()
	a.lastImp.key, a.lastImp.sheets = label, sheets
	a.mu.Unlock()
	fonte := filepath.Base(label)
	if strings.HasPrefix(label, "http") {
		fonte = label
	}
	fonte += " (" + time.Now().Format("02/01/2006") + ")"
	res, err := importSheets(sheets, req.Sheet, importOpts{Modello: req.Modello, ToneRx: req.ToneRx, Fonte: fonte})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"result": res, "label": label}, nil
}

// ---------------------------------------------------------------- esportazione

func (a *App) apiExport(r *http.Request) (interface{}, error) {
	var req struct {
		Format   string   `json:"format"` // xlsx | xls | csv | chirp
		What     string   `json:"what"`   // ponti | radio
		Ponti    []*Ponte `json:"ponti"`
		Radio    string   `json:"radio"`
		Channels []*Chan  `json:"channels"`
		Sel      []int    `json:"sel"`
		Path     string   `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	var out []byte
	var err error
	base := "ponti"
	if req.What == "radio" {
		base = map[string]string{kindXLC: "K67_canali", kindDAT: "Bajeton_canali"}[req.Radio]
		if m, ok := genByID[req.Radio]; ok {
			base = genFileBase(m) + "_canali"
		}
		if base == "" {
			return nil, errors.New("radio sconosciuta")
		}
		sel := req.Sel
		if sel == nil {
			for i := range req.Channels {
				sel = append(sel, i)
			}
		}
		if req.Format == "chirp" {
			out = chanChirp(req.Radio, req.Channels, sel)
		} else {
			out, err = exportSheet(chanSheet(req.Radio, req.Channels, sel), req.Format)
		}
	} else {
		if req.Format == "chirp" {
			out = pontiChirp(req.Ponti)
		} else {
			out, err = exportSheet(pontiSheet(req.Ponti), req.Format)
		}
	}
	if err != nil {
		return nil, err
	}
	ext := req.Format
	filt := map[string]string{
		"xlsx":  "Excel (*.xlsx)|*.xlsx",
		"xls":   "Excel 97-2003 (*.xls)|*.xls",
		"csv":   "CSV per Excel (*.csv)|*.csv",
		"chirp": "CSV per CHIRP (*.csv)|*.csv",
	}[req.Format]
	if req.Format == "chirp" {
		ext = "csv"
		base += "_CHIRP"
	}
	p := req.Path
	if p == "" {
		suggest := base + "." + ext
		if a.store.Settings.LastDir != "" {
			suggest = filepath.Join(a.store.Settings.LastDir, suggest)
		}
		if p, err = a.dialog(true, "Esporta", filt, suggest, ext); err != nil {
			return nil, err
		}
	}
	if err := writeAtomic(p, out); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	return map[string]interface{}{"path": p, "bytes": len(out)}, nil
}

// ---------------------------------------------------------------- avvio

func (a *App) listen() (net.Listener, string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("http://%s/index.html?t=%s", ln.Addr().String(), a.token)
	return ln, url, nil
}
