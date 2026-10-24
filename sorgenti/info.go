// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Informazioni sul programma e verifica degli aggiornamenti.
//
// L'indirizzo del pacchetto pubblicato è fisso (non modificabile dall'utente).
// La verifica NON scarica il pacchetto: legge con richieste parziali (HTTP Range)
// solo l'indice dello zip e il file "versione.txt" (righe chiave=valore:
// versione, compilazione), oppure la riga "Versione x.y" del LEGGIMI.txt.
// Solo se la versione pubblicata è più recente il programma propone di
// scaricarla: il pacchetto viene scaricato quando l'utente lo chiede e salvato
// sul PC. La verifica si fa con il pulsante e, in automatico, a ogni avvio.

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appName          = "IU0VYT Flash"
	appAuthor        = `Daniele "IU0VYT" Bonifazi`
	appEmail         = ""
	appQRZ           = "https://www.qrz.com/db/IU0VYT"
	appLicense       = "GNU General Public License versione 3 o successiva (GPL-3.0-or-later)"
	defaultUpdateURL = "https://github.com/IU0VYT/IU0VYT/releases/download/NEW/IU0VYT.Flash.zip"
	// firma letta dalla verifica degli aggiornamenti nell'eseguibile pubblicato
	versionMarker = "IU0VYT-FLASH-VERSIONE:" + appVersion + ";"
)

// data e ora di compilazione: -ldflags "-X main.buildDate=2026-10-20T21:45"
var buildDate = ""

// openExternal apre un indirizzo nel browser o un file in Esplora risorse
// (impostata dalla versione Windows; altrove l'interfaccia apre il link da sé).
var openExternal = func(target string, selectFile bool) bool { return false }

func fmtBuildDate(s string) string {
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			if layout == "2006-01-02" {
				return t.Format("02/01/2006")
			}
			return t.Format("02/01/2006 15:04")
		}
	}
	return s
}

type updState struct {
	mu      sync.Mutex
	version string          // versione pubblicata trovata dall'ultima verifica
	saved   map[string]bool // file salvati dal programma (si possono mostrare in Esplora risorse)
}

var upd = &updState{saved: map[string]bool{}}

// updateSource: indirizzo fisso del pacchetto pubblicato (cambia solo nei test)
var updateSource = defaultUpdateURL

func (a *App) infoHandlers(api func(string, func(*http.Request) (interface{}, error))) {
	api("info", a.apiInfo)
	api("update/check", a.apiUpdateCheck)
	api("update/save", a.apiUpdateSave)
	api("open", a.apiOpen2)
}

func (a *App) apiInfo(r *http.Request) (interface{}, error) {
	bd := fmtBuildDate(buildDate)
	if bd == "" {
		bd = "non indicata"
	}
	return map[string]interface{}{
		"name": appName, "version": appVersion, "buildDate": bd, "author": appAuthor, "email": appEmail,
		"qrz": appQRZ, "dataDir": a.store.dir, "marker": versionMarker, "license": appLicense,
	}, nil
}

var (
	reMarker    = regexp.MustCompile(`IU0VYT-FLASH-VERSIONE:(\d+(?:\.\d+)*);`)
	reLeggimi   = regexp.MustCompile(`(?m)^Versione (\d+(?:\.\d+)+)`)
	reKeyValue  = regexp.MustCompile(`(?m)^\s*([A-Za-z]+)\s*=\s*(.+?)\s*$`)
	errNotFound = errors.New("versione non trovata")
)

// packageVersion legge versione e data di compilazione dal pacchetto zip pubblicato.
func packageVersion(b []byte) (version, built string, err error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", "", fmt.Errorf("il file scaricato non è un archivio .zip valido (%v)", err)
	}
	return zipVersion(zr, true)
}

// zipVersion: versione dal pacchetto aperto; withExe = cerca anche la firma
// nell'eseguibile (solo per un pacchetto già scaricato: l'eseguibile è grande).
func zipVersion(zr *zip.Reader, withExe bool) (version, built string, err error) {
	read := func(f *zip.File, max int64) []byte {
		rc, err := f.Open()
		if err != nil {
			return nil
		}
		defer rc.Close()
		d, _ := io.ReadAll(io.LimitReader(rc, max))
		return d
	}
	var exe, leggimi []*zip.File
	for _, f := range zr.File {
		base := strings.ToLower(filepath.Base(strings.ReplaceAll(f.Name, "\\", "/")))
		switch {
		case base == "versione.txt":
			for _, m := range reKeyValue.FindAllStringSubmatch(string(read(f, 64<<10)), -1) {
				switch strings.ToLower(m[1]) {
				case "versione":
					version = m[2]
				case "compilazione":
					built = m[2]
				}
			}
			if version != "" {
				return version, built, nil
			}
		case strings.HasSuffix(base, ".exe") && withExe:
			exe = append(exe, f)
		case base == "leggimi.txt":
			leggimi = append(leggimi, f)
		}
	}
	for _, f := range exe {
		if m := reMarker.FindSubmatch(read(f, 200<<20)); m != nil {
			return string(m[1]), "", nil
		}
	}
	for _, f := range leggimi {
		if m := reLeggimi.FindSubmatch(read(f, 1<<20)); m != nil {
			return string(m[1]), "", nil
		}
	}
	return "", "", errNotFound
}

// cmpVersion: -1, 0, 1 (numeri separati da punti)
func cmpVersion(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

var updHTTP = &http.Client{Timeout: 5 * time.Minute}

// errore per le risposte del server degli aggiornamenti
type updStatusError struct {
	code int
	msg  string
}

func (e *updStatusError) Error() string { return e.msg }

func updStatusErr(resp *http.Response) error {
	e := &updStatusError{code: resp.StatusCode}
	switch resp.StatusCode {
	case 404:
		e.msg = "il pacchetto degli aggiornamenti non è stato trovato sul server (risposta 404): riprova più tardi o contatta l'autore"
	case 401, 403:
		e.msg = "il server degli aggiornamenti non consente di scaricare il pacchetto (accesso negato): contatta l'autore"
	default:
		e.msg = fmt.Sprintf("il server degli aggiornamenti ha risposto «%s»: riprova più tardi", resp.Status)
	}
	return e
}

func updRequest(u, rng string) (*http.Response, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IU0VYT-Flash/"+appVersion)
	req.Header.Set("Accept", "application/zip, application/octet-stream, */*")
	if rng != "" {
		req.Header.Set("Range", rng)
	}
	resp, err := updHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("collegamento al server degli aggiornamenti non riuscito: controlla la connessione a Internet (%v)", err)
	}
	return resp, nil
}

// downloadUpdate scarica l'intero pacchetto (solo quando l'utente lo chiede).
func downloadUpdate(u string) ([]byte, error) {
	resp, err := updRequest(u, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, updStatusErr(resp)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 300<<20))
	if err != nil {
		return nil, fmt.Errorf("download interrotto: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("PK")) {
		return nil, errors.New("il server ha restituito una pagina web invece del pacchetto .zip")
	}
	return b, nil
}

// remoteZip legge parti del pacchetto pubblicato con richieste HTTP Range
// (io.ReaderAt per archive/zip), a blocchi da 64 KB.
type remoteZip struct {
	url    string // indirizzo finale, dopo i reindirizzamenti
	size   int64
	tail   []byte // ultimi byte del file (prima richiesta)
	tailAt int64
	blocks map[int64][]byte
	read   int64 // byte ricevuti in tutto
	err    error // primo errore di una richiesta parziale
}

const remoteBlock = 64 << 10

var reContentRange = regexp.MustCompile(`^bytes (\d+)-(\d+)/(\d+)$`)

// openRemoteZip: richieste parziali con intervalli espliciti. Il server dei file
// di GitHub (release-assets.githubusercontent.com) rifiuta gli intervalli "gli
// ultimi N byte" (bytes=-N) con «501 Unsupported client range»: prima si chiede
// il primo byte per conoscere la dimensione, poi gli ultimi 64 KB (dove sta
// l'indice dello zip). Se il server non accetta richieste parziali restituisce il
// pacchetto intero (risposta 200).
func openRemoteZip(u string) (*remoteZip, []byte, error) {
	get := func(u, rng string) (*http.Response, []byte, error) {
		resp, err := updRequest(u, rng)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()
		switch resp.StatusCode {
		case 200:
			b, err := io.ReadAll(io.LimitReader(resp.Body, 300<<20))
			if err != nil {
				return nil, nil, fmt.Errorf("lettura interrotta: %v", err)
			}
			return resp, b, nil
		case 206:
			b, err := io.ReadAll(io.LimitReader(resp.Body, remoteBlock+1))
			if err != nil {
				return nil, nil, fmt.Errorf("lettura interrotta: %v", err)
			}
			return resp, b, nil
		}
		return nil, nil, updStatusErr(resp)
	}
	resp, b, err := get(u, "bytes=0-0")
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == 200 {
		return nil, b, nil
	}
	m := reContentRange.FindStringSubmatch(resp.Header.Get("Content-Range"))
	if m == nil {
		return nil, nil, errRangeUnsupported
	}
	size, _ := strconv.ParseInt(m[3], 10, 64)
	if size <= 0 {
		return nil, nil, errRangeUnsupported
	}
	// indirizzo finale dopo i reindirizzamenti (su GitHub vale qualche minuto)
	final := resp.Request.URL.String()
	start := max(0, size-remoteBlock)
	resp, b, err = get(final, fmt.Sprintf("bytes=%d-%d", start, size-1))
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode == 200 {
		return nil, b, nil
	}
	m = reContentRange.FindStringSubmatch(resp.Header.Get("Content-Range"))
	if m == nil || int64(len(b)) != size-start {
		return nil, nil, errRangeUnsupported
	}
	rz := &remoteZip{url: final, size: size, tail: b, tailAt: start, blocks: map[int64][]byte{}, read: 1 + int64(len(b))}
	return rz, nil, nil
}

// errRangeUnsupported: il server non gestisce come previsto le richieste parziali
var errRangeUnsupported = errors.New("il server degli aggiornamenti non gestisce le richieste parziali")

func (rz *remoteZip) block(off int64) ([]byte, error) {
	if b, ok := rz.blocks[off]; ok {
		return b, nil
	}
	end := min(off+remoteBlock, rz.size)
	resp, err := updRequest(rz.url, fmt.Sprintf("bytes=%d-%d", off, end-1))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 206 {
		rz.err = updStatusErr(resp)
		return nil, rz.err
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, end-off+1))
	if err == nil && int64(len(b)) != end-off {
		err = errors.New("risposta parziale del server non valida")
	}
	if err != nil {
		rz.err = err
		return nil, err
	}
	rz.read += int64(len(b))
	rz.blocks[off] = b
	return b, nil
}

func (rz *remoteZip) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for n < len(p) {
		pos := off + int64(n)
		if pos >= rz.size {
			return n, io.EOF
		}
		if pos >= rz.tailAt {
			n += copy(p[n:], rz.tail[pos-rz.tailAt:])
			continue
		}
		start := pos / remoteBlock * remoteBlock
		b, err := rz.block(start)
		if err != nil {
			return n, err
		}
		n += copy(p[n:], b[pos-start:])
	}
	return n, nil
}

// remoteVersion: versione del pacchetto pubblicato senza scaricarlo.
// Restituisce anche la dimensione del pacchetto e i byte effettivamente letti.
func remoteVersion(u string) (version, built string, size, read int64, err error) {
	version, built, size, read, err = remoteVersionOnce(u, false)
	if err != nil && rangeTrouble(err) {
		// richieste parziali non riuscite (server diverso dal previsto): si legge
		// il pacchetto intero, come farebbe un normale download
		version, built, size, read, err = remoteVersionOnce(u, true)
	}
	return
}

// rangeTrouble: errori per cui ha senso riprovare senza richieste parziali
// (non: pacchetto inesistente, accesso negato, nessuna connessione)
func rangeTrouble(err error) bool {
	if err == errRangeUnsupported {
		return true
	}
	var se *updStatusError
	return errors.As(err, &se) && se.code != 404 && se.code != 401 && se.code != 403
}

func remoteVersionOnce(u string, whole bool) (version, built string, size, read int64, err error) {
	var rz *remoteZip
	var all []byte
	if whole {
		all, err = downloadUpdate(u)
	} else {
		rz, all, err = openRemoteZip(u)
	}
	if err != nil {
		return "", "", 0, 0, err
	}
	var zr *zip.Reader
	if all != nil {
		// server senza richieste parziali: il pacchetto è arrivato intero
		size, read = int64(len(all)), int64(len(all))
		if !bytes.HasPrefix(all, []byte("PK")) {
			return "", "", 0, 0, errors.New("il server ha restituito una pagina web invece del pacchetto .zip")
		}
		zr, err = zip.NewReader(bytes.NewReader(all), size)
	} else {
		zr, err = zip.NewReader(rz, rz.size)
		size = rz.size
	}
	if err != nil {
		if rz != nil && rz.err != nil {
			return "", "", 0, 0, errRangeUnsupported
		}
		return "", "", 0, 0, fmt.Errorf("il pacchetto pubblicato non è un archivio .zip valido (%v)", err)
	}
	version, built, err = zipVersion(zr, false)
	if rz != nil {
		read = rz.read
		if rz.err != nil && err != nil {
			return "", "", 0, 0, errRangeUnsupported
		}
	}
	if err == errNotFound {
		err = errors.New("nel pacchetto pubblicato non è indicata la versione del programma")
	}
	return version, built, size, read, err
}

func (a *App) apiUpdateCheck(r *http.Request) (interface{}, error) {
	v, built, size, read, err := remoteVersion(updateSource)
	if err != nil {
		return nil, err
	}
	upd.mu.Lock()
	upd.version = v
	upd.mu.Unlock()
	return map[string]interface{}{"remote": v, "remoteBuild": fmtBuildDate(built), "local": appVersion,
		"cmp": cmpVersion(v, appVersion), "size": size, "read": read}, nil
}

func (a *App) apiUpdateSave(r *http.Request) (interface{}, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	upd.mu.Lock()
	checked := upd.version
	upd.mu.Unlock()
	if checked == "" || cmpVersion(checked, appVersion) <= 0 {
		return nil, errors.New("nessun aggiornamento da scaricare: verifica prima gli aggiornamenti")
	}
	// prima la scelta della cartella, poi il download (una volta sola)
	p := req.Path
	if p == "" {
		name := appName + " " + checked + ".zip"
		dir := ""
		if h, err := os.UserHomeDir(); err == nil {
			if d := filepath.Join(h, "Downloads"); dirExists(d) {
				dir = d
			}
		}
		suggest := name
		if dir != "" {
			suggest = filepath.Join(dir, name)
		}
		var err error
		if p, err = a.dialog(true, "Salva l'aggiornamento di "+appName, "Archivio zip (*.zip)|*.zip", suggest, "zip"); err != nil {
			return nil, err
		}
	}
	b, err := downloadUpdate(updateSource)
	if err != nil {
		return nil, err
	}
	v, _, err := packageVersion(b)
	if err != nil {
		return nil, fmt.Errorf("pacchetto scaricato non valido: %v", err)
	}
	if cmpVersion(v, appVersion) <= 0 {
		return nil, fmt.Errorf("il pacchetto scaricato contiene la versione %s, non più recente di quella installata", v)
	}
	if err := writeAtomic(p, b); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	upd.mu.Lock()
	upd.saved[p] = true
	upd.mu.Unlock()
	return map[string]interface{}{"path": p, "version": v, "size": len(b)}, nil
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// apiOpen2: apre un collegamento (http/https/mailto) o mostra un file salvato
func (a *App) apiOpen2(r *http.Request) (interface{}, error) {
	var req struct {
		URL  string `json:"url"`
		File string `json:"file"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.File != "" {
		upd.mu.Lock()
		ok := upd.saved[req.File]
		upd.mu.Unlock()
		if !ok {
			return nil, errors.New("file non disponibile")
		}
		return map[string]bool{"opened": openExternal(req.File, true)}, nil
	}
	u := strings.TrimSpace(req.URL)
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "mailto:") {
		return nil, errors.New("collegamento non valido")
	}
	return map[string]bool{"opened": openExternal(u, false)}, nil
}
