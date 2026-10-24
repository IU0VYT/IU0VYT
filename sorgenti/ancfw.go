// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// File del firmware della Bajeton 8300 (.anc): controlli, estrazione dagli
// archivi .zip e .rar (anche uno dentro l'altro), scaricamento da Internet.
//
// Il file .anc è un'immagine JieLi (contenitore "JLUFW") cifrata dal programma
// del costruttore (vedi ancDecrypt). I controlli servono a non inviare alla
// radio un file sbagliato: formato riconosciuto, decifratura coerente,
// piattaforma uguale a quella del firmware ufficiale della BJ-8300, nome del
// file, impronta SHA-256 dei firmware ufficiali conosciuti.
//
// Gli archivi .zip e .rar sono aperti qui; i .rar con il decodificatore libero
// github.com/nwaples/rardecode (licenza BSD, formati RAR 1.5-5).

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nwaples/rardecode/v2"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	fwMaxFile    = 32 << 20  // dimensione massima di un .anc
	fwMaxEntry   = 128 << 20 // dimensione massima di un file dentro un archivio
	fwMaxTotal   = 256 << 20 // totale estratto da un archivio
	fwMaxDepth   = 4         // archivi uno dentro l'altro
	fwMaxDown    = 128 << 20 // download
	fwPlatformAt = 0x410     // nome della piattaforma JieLi nell'immagine decifrata
)

// piattaforma del firmware ufficiale BJ-8300
const bj8300Platform = "YHD15C8_soundbox"

// firmware ufficiali conosciuti (SHA-256 del file .anc)
var fwKnown = map[string]string{
	"4952585480cab69c035804e748e512405763abedee966e795b25885e303cf52d": "Firmware ufficiale Bajeton BJ-8300 V0.12 del 12/11/2025 (BJ8300_V0.12_20251112.anc)",
	"1509800ac1e6b88d8fe028acc8239c5df696c208ab1be7e5a5a3626c872f5a41": "Firmware ufficiale Bajeton BJ-7800 V0.17 del 26/11/2025 (BJ7800_CN_EN_V0.17_20251126.anc)",
	"db74b6c4d497712c0f2db67a66b939b9c21f336bba1976354edf1e26c4ba6aba": "Firmware ufficiale Bajeton BJ-7700 con Bluetooth V0.06 del 24/11/2025 (BJ7700PRO_V0.06_251124.anc)",
}

type fwCheck struct {
	Level string `json:"level"` // ok | warn | bad
	Text  string `json:"text"`
}

type fwInfo struct {
	Kind     string    `json:"kind"`   // anc (Bajeton 8300/7800) | bf (Baofeng KDH .BF) | rh5 (Baofeng 5RH Pro .dat) | kdhx (UV-25 Plus) | bj9000 | kdh9g (UV-9G)
	Target   string    `json:"target"` // radio a cui è destinato
	Confirm  string    `json:"confirm"`
	Hw       string    `json:"hw,omitempty"`
	Name     string    `json:"name"`
	Origin   string    `json:"origin"`
	Size     int       `json:"size"`
	SHA256   string    `json:"sha256"`
	Platform string    `json:"platform"`
	Copies   int       `json:"copies"`
	Version  string    `json:"version"`
	Date     string    `json:"date"`
	Known    string    `json:"known"`
	Checks   []fwCheck `json:"checks"`
	Valid    bool      `json:"valid"`    // si può inviare alla radio
	Warnings bool      `json:"warnings"` // servono conferme in più
	Saved    string    `json:"saved,omitempty"`
}

type fwImage struct {
	Info fwInfo
	Raw  []byte // file .anc come distribuito
	Dec  []byte // decifrato (quello che riceve la radio)
}

var (
	reModel   = regexp.MustCompile(`(?i)BJ[-_ ]?(8300|7800|7700)`)
	reVersion = regexp.MustCompile(`(?i)(?:^|[_\- ])V(\d+(?:\.\d+)+)`)
	reDate    = regexp.MustCompile(`(?:^|[_\- ])(20\d{2})(\d{2})(\d{2})(?:$|[_\-. ])`)
	rePlat    = regexp.MustCompile(`[A-Za-z0-9]{3,24}_soundbox`)
)

// analyzeANC controlla un file .anc e lo prepara per l'invio.
func analyzeANC(name, origin string, raw []byte) *fwImage {
	info := fwInfo{Kind: "anc", Target: "Bajeton BJ-8300", Confirm: "BJ8300", Name: name, Origin: origin, Size: len(raw)}
	model := "8300" // radio indicata dal nome del file (BJ-8300 se non indicata)
	if m := reModel.FindStringSubmatch(strings.TrimSuffix(name, path.Ext(name))); m != nil {
		model = m[1]
	}
	switch model {
	case "7800":
		info.Target, info.Confirm = "Bajeton BJ-7800", "BJ7800"
	case "7700":
		info.Target, info.Confirm = "Bajeton BJ-7700 (con Bluetooth)", "BJ7700"
	}
	sum := sha256.Sum256(raw)
	info.SHA256 = hex.EncodeToString(sum[:])
	add := func(level, f string, a ...interface{}) {
		info.Checks = append(info.Checks, fwCheck{Level: level, Text: fmt.Sprintf(f, a...)})
	}
	img := &fwImage{Raw: raw}
	valid := true
	if len(raw) < 8*1024 || len(raw) > fwMaxFile {
		add("bad", "Dimensione non plausibile per un firmware (%s)", fmtBytes(len(raw)))
		info.Valid = false
		img.Info = info
		return img
	}
	tail := raw[len(raw)-min(len(raw), 256):]
	if bytes.Contains(tail, []byte("JLUFW")) {
		add("ok", "Formato riconosciuto: firmware per il programma di aggiornamento BinTe/KDH (immagine JieLi «JLUFW»)")
	} else {
		add("bad", "Formato non riconosciuto: manca la firma «JLUFW» dei firmware Bajeton (BJ-8300, BJ-7800). Non è un file .anc valido")
		valid = false
	}
	dec := ancDecrypt(raw)
	img.Dec = dec
	if bytes.Contains(dec, []byte("UPDATE_JUMP")) && countAt(dec, []byte("UPDATE_JUMP")) > countAt(raw, []byte("UPDATE_JUMP")) {
		add("ok", "Decifratura riuscita: le intestazioni interne sono leggibili")
	} else if valid {
		add("bad", "La decifratura non dà un contenuto riconoscibile: file danneggiato o di un altro tipo")
		valid = false
	}
	// piattaforma
	if len(dec) > fwPlatformAt+16 {
		p := strings.TrimRight(string(dec[fwPlatformAt:fwPlatformAt+16]), "\x00 ")
		if !isPrintableASCII(p) || len(p) < 4 {
			p = ""
			if m := rePlat.Find(dec[:min(len(dec), 16384)]); m != nil {
				p = string(m)
			}
		}
		info.Platform = p
		if p != "" {
			info.Copies = countAt(dec, []byte(p))
		}
	}
	warn := false
	switch {
	case info.Platform == bj8300Platform:
		add("ok", "Piattaforma %s: la stessa dei firmware ufficiali di BJ-8300, BJ-7800 e BJ-7700 (non distingue il modello: contano il nome del file, l'identificazione della radio e la tua conferma)", info.Platform)
	case info.Platform == "":
		add("warn", "Piattaforma non indicata nel file: impossibile confrontarla con quella dei firmware ufficiali")
		warn = true
	default:
		add("warn", "Piattaforma %s, diversa da quella dei firmware ufficiali Bajeton (%s): potrebbe essere il firmware di un'altra radio", info.Platform, bj8300Platform)
		warn = true
	}
	base := strings.TrimSuffix(name, path.Ext(name))
	if reModel.MatchString(base) {
		add("ok", "Il nome del file indica la %s", info.Target)
	} else {
		add("warn", "Il nome del file non indica la radio (atteso ad es. «BJ8300_V0.12_20251112.anc», «BJ7800_CN_EN_V0.17_20251126.anc» o «BJ7700PRO_V0.06_251124.anc»): considerato per la BJ-8300")
		warn = true
	}
	if m := reVersion.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1]
	}
	if m := reDate.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/" + m[1]
	}
	if k, ok := fwKnown[info.SHA256]; ok {
		info.Known = k
		add("ok", "%s: impronta SHA-256 verificata", k)
		warn = false
	} else if valid {
		add("warn", "Firmware non presente nell'elenco dei firmware ufficiali conosciuti da questo programma (controllo dell'impronta SHA-256)")
	}
	info.Valid = valid
	info.Warnings = warn
	img.Info = info
	return img
}

func countAt(b, sub []byte) int { return bytes.Count(b, sub) }

func isPrintableASCII(s string) bool {
	for _, c := range []byte(s) {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func fmtBytes(n int) string {
	switch {
	case n >= 1<<20:
		return strings.Replace(fmt.Sprintf("%.2f MB", float64(n)/(1<<20)), ".", ",", 1)
	case n >= 1<<10:
		return strings.Replace(fmt.Sprintf("%.1f KB", float64(n)/(1<<10)), ".", ",", 1)
	}
	return fmt.Sprintf("%d byte", n)
}

// ---------------------------------------------------------------- archivi

type fwKind int

const (
	fwOther fwKind = iota
	fwANC
	fwZIP
	fwRAR
	fwBF     // Baofeng, caricatore KDH (.BF)
	fwRH5    // Baofeng 5RH Pro (.dat)
	fwKDHX   // Baofeng UV-25 Plus (.kdhx)
	fwBJ9000 // Bajeton BJ-9000 (.Firmware)
	fwKDH9G  // Baofeng UV-9G (.kdh)
)

func sniffKind(name string, b []byte) fwKind {
	switch {
	case bytes.HasPrefix(b, []byte("PK\x03\x04")) || bytes.HasPrefix(b, []byte("PK\x05\x06")):
		return fwZIP
	case bytes.HasPrefix(b, []byte("Rar!\x1a\x07")):
		return fwRAR
	}
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".anc":
		return fwANC
	case ".kdhx":
		return fwKDHX
	case ".kdh":
		return fwKDH9G
	case ".firmware", ".btf":
		return fwBJ9000
	}
	if bytes.HasPrefix(b, []byte("BaoFeng")) && len(b) > rh5Head && string(b[16:22]) == "BF_5RH" {
		return fwRH5
	}
	if ext == ".bf" || (len(b) > kdhHdr+kdhPkt && b[0] == 2 && kdhHdr+be32(b[1:5])+be32(b[5:9]) == int64(len(b))) {
		return fwBF
	}
	if len(b) > 1024 && bytes.Contains(b[len(b)-min(len(b), 256):], []byte("JLUFW")) {
		return fwANC
	}
	return fwOther
}

// fwItem: file in attesa di essere esaminato (eventualmente dentro archivi)
type fwItem struct {
	Name  string // nome del file
	Where string // percorso: origine › archivio › file
	Data  []byte
	Depth int
}

type fwCandidate struct {
	Name  string `json:"name"`
	Where string `json:"where"`
	Size  int    `json:"size"`
	data  []byte
}

// fwExpander estrae i file firmware da file e archivi .zip/.rar (anche uno
// dentro l'altro).
type fwExpander struct {
	queue    []fwItem
	found    []fwCandidate
	total    int
	skipped  []string
	archives int
}

// fwWanted: file da estrarre dagli archivi (firmware conosciuti e altri archivi)
func fwWanted(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".anc", ".bf", ".dat", ".kdhx", ".kdh", ".firmware", ".btf", ".zip", ".rar":
		return true
	}
	return false
}

func newExpander(name, where string, data []byte) *fwExpander {
	return &fwExpander{queue: []fwItem{{Name: name, Where: where, Data: data}}}
}

// run elabora la coda.
func (e *fwExpander) run() error {
	for len(e.queue) > 0 {
		it := e.queue[0]
		e.queue = e.queue[1:]
		switch sniffKind(it.Name, it.Data) {
		case fwANC, fwBF, fwRH5, fwKDHX, fwBJ9000, fwKDH9G:
			e.found = append(e.found, fwCandidate{Name: it.Name, Where: it.Where, Size: len(it.Data), data: it.Data})
		case fwZIP:
			if it.Depth >= fwMaxDepth {
				e.skipped = append(e.skipped, it.Where+": troppi archivi uno dentro l'altro")
				continue
			}
			e.archives++
			if err := e.expandZip(it); err != nil {
				if it.Depth == 0 {
					return err
				}
				e.skipped = append(e.skipped, fmt.Sprintf("%s: %v", it.Where, err))
			}
		case fwRAR:
			if it.Depth >= fwMaxDepth {
				e.skipped = append(e.skipped, it.Where+": troppi archivi uno dentro l'altro")
				continue
			}
			e.archives++
			if err := e.expandRar(it); err != nil {
				if it.Depth == 0 {
					return err
				}
				e.skipped = append(e.skipped, fmt.Sprintf("%s: %v", it.Where, err))
			}
		default:
			if it.Depth == 0 {
				return fmt.Errorf("«%s» non è un firmware riconosciuto (.anc, .BF, .dat, .kdhx, .kdh, .Firmware) né un archivio .zip o .rar", it.Name)
			}
		}
	}
	return nil
}

// expandRar apre un archivio .rar (RAR 1.5-5, anche "solido"); archivi divisi
// in più volumi e protetti da password non sono supportati.
func (e *fwExpander) expandRar(it fwItem) error {
	rr, err := rardecode.NewReader(bytes.NewReader(it.Data), rardecode.MaxDictionarySize(256<<20))
	if err != nil {
		return rarErr(err)
	}
	for {
		h, err := rr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return rarErr(err)
		}
		if h.IsDir || h.LinkType != 0 {
			continue
		}
		name := strings.ReplaceAll(h.Name, "\\", "/")
		if !fwWanted(name) {
			continue
		}
		if h.UnPackedSize > fwMaxEntry {
			e.skipped = append(e.skipped, it.Where+" › "+name+": file troppo grande")
			continue
		}
		if h.Encrypted {
			e.skipped = append(e.skipped, it.Where+" › "+name+": protetto da password")
			continue
		}
		b, err := io.ReadAll(io.LimitReader(rr, fwMaxEntry+1))
		if err != nil {
			e.skipped = append(e.skipped, fmt.Sprintf("%s › %s: %v", it.Where, name, rarErr(err)))
			continue
		}
		e.total += len(b)
		if e.total > fwMaxTotal {
			return errors.New("l'archivio contiene troppi dati")
		}
		e.queue = append(e.queue, fwItem{Name: path.Base(name), Where: it.Where + " › " + name, Data: b, Depth: it.Depth + 1})
	}
}

func rarErr(err error) error {
	switch {
	case errors.Is(err, rardecode.ErrArchiveEncrypted), errors.Is(err, rardecode.ErrArchivedFileEncrypted), errors.Is(err, rardecode.ErrBadPassword):
		return errors.New("archivio .rar protetto da password (non supportato)")
	case errors.Is(err, rardecode.ErrMultiVolume):
		return errors.New("archivio .rar diviso in più parti (non supportato: unisci le parti o estrai il firmware)")
	}
	return fmt.Errorf("archivio .rar non leggibile (%v)", strings.TrimPrefix(err.Error(), "rardecode: "))
}

func zipName(f *zip.File) string {
	n := f.Name
	if f.NonUTF8 || !utf8.ValidString(n) {
		// nomi scritti con la codifica cinese (GBK), frequenti negli archivi dei costruttori
		if s, err := simplifiedchinese.GBK.NewDecoder().String(n); err == nil && utf8.ValidString(s) {
			n = s
		}
	}
	return strings.ReplaceAll(n, "\\", "/")
}

func (e *fwExpander) expandZip(it fwItem) error {
	zr, err := zip.NewReader(bytes.NewReader(it.Data), int64(len(it.Data)))
	if err != nil {
		return fmt.Errorf("archivio .zip non leggibile: %v", err)
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := zipName(f)
		if !fwWanted(name) {
			continue
		}
		if f.UncompressedSize64 > fwMaxEntry {
			e.skipped = append(e.skipped, it.Where+" › "+name+": file troppo grande")
			continue
		}
		if f.Flags&0x1 != 0 {
			e.skipped = append(e.skipped, it.Where+" › "+name+": protetto da password")
			continue
		}
		rc, err := f.Open()
		if err != nil {
			e.skipped = append(e.skipped, fmt.Sprintf("%s › %s: %v", it.Where, name, err))
			continue
		}
		b, err := io.ReadAll(io.LimitReader(rc, fwMaxEntry+1))
		rc.Close()
		if err != nil {
			e.skipped = append(e.skipped, fmt.Sprintf("%s › %s: %v", it.Where, name, err))
			continue
		}
		e.total += len(b)
		if e.total > fwMaxTotal {
			return errors.New("l'archivio contiene troppi dati")
		}
		e.queue = append(e.queue, fwItem{Name: path.Base(name), Where: it.Where + " › " + name, Data: b, Depth: it.Depth + 1})
	}
	return nil
}

// ---------------------------------------------------------------- Internet

var reDriveFile = regexp.MustCompile(`/file/d/([A-Za-z0-9_-]{10,})`)

// fwNormalizeURL converte i collegamenti di condivisione di Google Drive e
// Dropbox nel collegamento per scaricare il file.
func fwNormalizeURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(strings.ToLower(s), "http://") && !strings.HasPrefix(strings.ToLower(s), "https://") {
		return "", errors.New("l'indirizzo deve iniziare con http:// o https://")
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", errors.New("indirizzo non valido")
	}
	host := strings.ToLower(u.Host)
	switch {
	case host == "drive.google.com" || host == "docs.google.com":
		if strings.Contains(u.Path, "/folders/") || strings.Contains(u.Path, "/drive/") && !strings.Contains(u.Path, "/file/") {
			return "", errors.New("è il collegamento a una cartella di Google Drive, che non si può scaricare direttamente: apri la cartella nel browser e copia il collegamento del singolo file (tasto destro › Condividi › Copia link), oppure scarica il file e caricalo dal PC")
		}
		id := ""
		if m := reDriveFile.FindStringSubmatch(u.Path); m != nil {
			id = m[1]
		} else if v := u.Query().Get("id"); v != "" {
			id = v
		}
		if id != "" {
			return "https://drive.usercontent.google.com/download?export=download&confirm=t&id=" + url.QueryEscape(id), nil
		}
	case strings.HasSuffix(host, "dropbox.com"):
		q := u.Query()
		q.Set("dl", "1")
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
	return s, nil
}

var fwHTTP = &http.Client{Timeout: 5 * time.Minute}

// fwDownload scarica il file; restituisce dati e nome del file.
func fwDownload(raw string) ([]byte, string, error) {
	u, err := fwNormalizeURL(raw)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, "", fmt.Errorf("indirizzo non valido: %v", err)
	}
	req.Header.Set("User-Agent", "IU0VYT-Flash/"+appVersion)
	resp, err := fwHTTP.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download non riuscito: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("download non riuscito: il server ha risposto %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, fwMaxDown+1))
	if err != nil {
		return nil, "", fmt.Errorf("download interrotto: %v", err)
	}
	if len(b) > fwMaxDown {
		return nil, "", errors.New("il file da scaricare è troppo grande")
	}
	name := ""
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		name = dispositionName(cd)
	}
	if name == "" {
		name = path.Base(resp.Request.URL.Path)
		if un, err := url.PathUnescape(name); err == nil {
			name = un
		}
	}
	if name == "" || name == "/" || name == "." || name == "download" || name == "uc" {
		name = "firmware_scaricato"
	}
	head := bytes.TrimSpace(b[:min(len(b), 512)])
	if sniffKind(name, b) == fwOther && (bytes.HasPrefix(head, []byte("<")) || strings.Contains(resp.Header.Get("Content-Type"), "text/html")) {
		return nil, "", errors.New("l'indirizzo ha restituito una pagina web, non un file: usa il collegamento diretto al file (.anc, .zip o .rar)")
	}
	return b, sanitizeName(name), nil
}

var reDispUTF8 = regexp.MustCompile(`(?i)filename\*\s*=\s*UTF-8''([^;]+)`)
var reDisp = regexp.MustCompile(`(?i)filename\s*=\s*"?([^";]+)"?`)

func dispositionName(cd string) string {
	if m := reDispUTF8.FindStringSubmatch(cd); m != nil {
		if s, err := url.PathUnescape(strings.TrimSpace(m[1])); err == nil {
			return s
		}
	}
	if m := reDisp.FindStringSubmatch(cd); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func sanitizeName(n string) string {
	n = strings.ReplaceAll(n, "\\", "/")
	n = path.Base(n)
	n = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) || r < 32 {
			return '_'
		}
		return r
	}, n)
	if n == "" || n == "." || n == ".." {
		n = "firmware"
	}
	return n
}

func sha(b []byte) []byte {
	s := sha256.Sum256(b)
	return s[:]
}
