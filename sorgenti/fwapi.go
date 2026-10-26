// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware (Bajeton BJ-8300/BJ-7800 .anc e BJ-9000 .Firmware, Baofeng .BF, .dat, .kdhx e .kdh): API per l'interfaccia.
//
//   fw/load      sceglie il firmware: file del PC (.anc, .zip, .rar) o indirizzo Internet
//   fw/pick      scelta tra più firmware trovati negli archivi
//   fw/identify  identificazione della radio (protocollo della Bajeton 8300)
//   fw/start     aggiornamento (operazione in background: cable/status, cable/cancel)

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

type fwSession struct {
	mu     sync.Mutex
	exp    *fwExpander
	source string // file | url
	cands  []fwCandidate
	img    *fwImage
	ident  *fwIdent
}

type fwIdent struct {
	Port  string    `json:"port"`
	ID    string    `json:"id"`
	Proto string    `json:"proto"`
	At    time.Time `json:"at"`
}

func randID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *App) fwDir() string {
	d := filepath.Join(a.store.dir, "firmware")
	_ = os.MkdirAll(d, 0o755)
	return d
}

func (a *App) fwHandlers(mux *http.ServeMux, api func(string, func(*http.Request) (interface{}, error))) {
	api("fw/load", a.apiFwLoad)
	api("fw/pick", a.apiFwPick)
	api("fw/identify", a.apiFwIdentify)
	api("fw/start", a.apiFwStart)
	api("fw/info", func(r *http.Request) (interface{}, error) {
		return map[string]interface{}{"settings": a.store.Settings, "firmwareDir": a.fwDir(), "backupDir": a.backupDir()}, nil
	})
}

// esito di fw/load, fw/pick
func (a *App) fwResult() (interface{}, error) {
	s := &a.fw
	exp := s.exp
	if exp == nil {
		return nil, errors.New("nessun firmware in elaborazione")
	}
	if err := exp.run(); err != nil {
		s.exp = nil
		return nil, err
	}
	skipped := append([]string{}, exp.skipped...)
	s.exp = nil
	cands := exp.found
	if len(cands) == 0 {
		msg := "nell'archivio non c'è nessun file firmware riconosciuto (.anc, .BF, .dat, .kdhx, .kdh, .Firmware)"
		if len(skipped) > 0 {
			msg += " (" + strings.Join(skipped, "; ") + ")"
		}
		return nil, errors.New(msg)
	}
	// lo stesso file presente più volte (es. sia sciolto sia dentro un archivio) conta una volta
	seen := map[string]bool{}
	uniq := cands[:0:0]
	for _, c := range cands {
		h := hex.EncodeToString(sha(c.data))
		if !seen[h] {
			seen[h] = true
			uniq = append(uniq, c)
		}
	}
	cands = uniq
	// prima i file che indicano la BJ-8300, poi per nome
	sort.SliceStable(cands, func(i, j int) bool {
		bi, bj := reModel.MatchString(cands[i].Name), reModel.MatchString(cands[j].Name)
		if bi != bj {
			return bi
		}
		return cands[i].Where < cands[j].Where
	})
	s.cands = cands
	if len(cands) > 1 {
		return map[string]interface{}{"state": "choose", "choices": cands, "skipped": skipped}, nil
	}
	return a.fwSelect(0, skipped)
}

func (a *App) fwSelect(i int, skipped []string) (interface{}, error) {
	s := &a.fw
	if i < 0 || i >= len(s.cands) {
		return nil, errors.New("scelta non valida")
	}
	c := s.cands[i]
	var img *fwImage
	switch sniffKind(c.Name, c.data) {
	case fwBF:
		img = analyzeBF(c.Name, c.Where, c.data)
	case fwRH5:
		img = analyze5RH(c.Name, c.Where, c.data)
	case fwKDHX:
		img = analyzeKDHX(c.Name, c.Where, c.data)
	case fwBJ9000:
		img = analyzeBJ9000(c.Name, c.Where, c.data)
	case fwKDH9G:
		img = analyzeKDH9G(c.Name, c.Where, c.data)
	default:
		img = analyzeANC(c.Name, c.Where, c.data)
	}
	if img.Info.Valid && s.source == "url" {
		p := filepath.Join(a.fwDir(), sanitizeName(c.Name))
		if old, err := os.ReadFile(p); err != nil || hex.EncodeToString(sha(old)) != img.Info.SHA256 {
			if err := writeAtomic(p, c.data); err == nil {
				img.Info.Saved = p
			}
		} else {
			img.Info.Saved = p
		}
	}
	s.img = img
	s.cands = nil
	if skipped == nil {
		skipped = []string{}
	}
	return map[string]interface{}{"state": "ready", "firmware": img.Info, "skipped": skipped}, nil
}

func (a *App) apiFwLoad(r *http.Request) (interface{}, error) {
	var req struct {
		Source string `json:"source"` // file | url
		Path   string `json:"path"`
		URL    string `json:"url"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	var data []byte
	var name, where string
	if req.Source == "url" {
		b, n, err := fwDownload(req.URL)
		if err != nil {
			return nil, err
		}
		data, name, where = b, n, strings.TrimSpace(req.URL)
		st := a.store.Settings
		st.FwURL = where
		_ = a.store.SaveSettings(st)
	} else {
		p := req.Path
		if p == "" {
			var err error
			if p, err = a.dialog(false, "Scegli il firmware",
				"Firmware o archivio (*.anc, *.BF, *.dat, *.kdhx, *.kdh, *.Firmware, *.zip, *.rar)|*.anc;*.bf;*.dat;*.kdhx;*.kdh;*.firmware;*.btf;*.zip;*.rar|Bajeton BJ-8300/BJ-7800 (*.anc)|*.anc|Bajeton BJ-9000 (*.Firmware)|*.firmware;*.btf|Baofeng UV-5RM/5RH (*.BF)|*.bf|Baofeng 5RH Pro (*.dat)|*.dat|Baofeng UV-25 Plus (*.kdhx)|*.kdhx|Baofeng UV-9G (*.kdh)|*.kdh|Archivi (*.zip, *.rar)|*.zip;*.rar|Tutti i file|*.*", "", ""); err != nil {
				return nil, err
			}
		}
		fi, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("impossibile leggere il file: %v", err)
		}
		if fi.Size() > fwMaxDown {
			return nil, errors.New("il file è troppo grande per essere un firmware o un suo archivio")
		}
		if data, err = os.ReadFile(p); err != nil {
			return nil, fmt.Errorf("impossibile leggere il file: %v", err)
		}
		name, where = filepath.Base(p), p
	}
	s := &a.fw
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exp = newExpander(name, where, data)
	s.source = req.Source
	s.img = nil
	s.cands = nil
	return a.fwResult()
}

func (a *App) apiFwPick(r *http.Request) (interface{}, error) {
	var req struct {
		Index int `json:"index"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	s := &a.fw
	s.mu.Lock()
	defer s.mu.Unlock()
	return a.fwSelect(req.Index, nil)
}

// ---------------------------------------------------------------- operazioni con la radio

// startJob avvia un'operazione in background (come cable/start).
func (a *App) startJob(op string, canStop bool, run func(ctx context.Context, j *cableJob) (interface{}, error)) (map[string]interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: randID(), Op: op, Radio: kindDAT, State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}, CanStop: canStop}
	a.mu.Lock()
	if a.job != nil && a.job.isRunning() {
		a.mu.Unlock()
		cancel()
		return nil, errors.New("è già in corso un'operazione con la radio")
	}
	a.job = j
	a.mu.Unlock()
	go func() {
		defer cancel()
		var res interface{}
		var err error
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("errore interno: %v", e)
				}
			}()
			res, err = run(ctx, j)
		}()
		j.mu.Lock()
		defer j.mu.Unlock()
		switch {
		case err == errCancelled:
			j.State, j.Error = "cancelled", "Operazione annullata."
		case err != nil:
			j.State, j.Error = "error", err.Error()
		default:
			j.State, j.Result = "done", res
		}
		j.CanStop = false
	}()
	return j.snapshot(), nil
}

func (a *App) apiFwIdentify(r *http.Request) (interface{}, error) {
	var req struct {
		Port string `json:"port"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione")
	}
	st := a.store.Settings
	st.BajPort = req.Port
	_ = a.store.SaveSettings(st)
	return a.startJob("identify", true, func(ctx context.Context, j *cableJob) (interface{}, error) {
		j.set("Identificazione della radio", true)
		a.fw.mu.Lock()
		img := a.fw.img
		a.fw.mu.Unlock()
		var rid, proto string
		var err error
		if img != nil && img.Info.Confirm == "BJ7700" {
			proto = "BJ-7700 (Radtel RT-910)"
			err, rid = rt900Identify(ctx, genByID["bj-7700"], req.Port), "PROGRAMBT80U"
		} else {
			proto = "BT-11 (BJ-8300 / BJ-7800)"
			var info btInfo
			info, err = btIdentify(ctx, req.Port, j.prog)
			rid = strings.TrimSpace(info.ID)
		}
		if err != nil {
			if err == errCancelled {
				return nil, err
			}
			return nil, fmt.Errorf("la radio collegata a %s non ha risposto al protocollo %s. %v", req.Port, proto, err)
		}
		id := &fwIdent{Port: req.Port, ID: rid, Proto: proto, At: time.Now()}
		a.fw.mu.Lock()
		a.fw.ident = id
		a.fw.mu.Unlock()
		return id, nil
	})
}

type fwStartReq struct {
	Port           string `json:"port"`
	Mode           string `json:"mode"`  // anc: update | recovery; bf, kdhx, bj9000, kdh9g: auto | manual; rh5: update
	Model          string `json:"model"` // bf: driver per la copia di sicurezza (es. bf-uv5rm)
	Baud           int    `json:"baud"`
	Backup         bool   `json:"backup"`
	SkipIdentify   bool   `json:"skipIdentify"`
	SHA256         string `json:"sha256"`
	ConfirmModel   bool   `json:"confirmModel"`
	ConfirmText    string `json:"confirmText"`
	AcceptWarnings bool   `json:"acceptWarnings"`
}

func normModel(s string) string {
	s = strings.ToUpper(s)
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '_' || r == '.' {
			return -1
		}
		return r
	}, s)
}

// fwCheckStart: controlli prima di toccare la radio
func fwCheckStart(req *fwStartReq, img *fwImage) error {
	if img == nil {
		return errors.New("scegli prima il file del firmware")
	}
	if req.SHA256 != img.Info.SHA256 {
		return errors.New("il firmware scelto è cambiato: ricaricalo")
	}
	if !img.Info.Valid {
		return errors.New("il file scelto non è un firmware valido per la radio a cui è destinato")
	}
	if img.Info.Warnings && !req.AcceptWarnings {
		return errors.New("il firmware ha delle segnalazioni: conferma di averle lette")
	}
	if strings.TrimSpace(req.Port) == "" {
		return errors.New("scegli la porta del cavo di programmazione")
	}
	switch img.Info.Kind {
	case "bf":
		if req.Mode != "auto" && req.Mode != "manual" {
			return errors.New("modalità sconosciuta")
		}
		if !req.ConfirmModel {
			return errors.New("conferma il modello della radio e il chip indicato dal firmware")
		}
		if (img.Info.Typed || fwOtherModel(req, img)) && normModel(req.ConfirmText) != img.Info.Confirm {
			return fmt.Errorf("firmware per una radio diversa dalla UV-5RM/5RH o da quella scelta: scrivi %s nella casella di conferma", img.Info.Confirm)
		}
		if req.Backup && req.Mode == "auto" {
			if _, ok := genByID[req.Model]; !ok {
				return errors.New("scegli il modello della radio per la copia di sicurezza")
			}
		}
		return nil
	case "kdhx", "bj9000", "kdh9g":
		if req.Mode != "auto" && req.Mode != "manual" {
			return errors.New("modalità sconosciuta")
		}
		if !req.ConfirmModel {
			return fmt.Errorf("conferma che la radio collegata è una %s", img.Info.Target)
		}
		if (img.Info.Typed || fwOtherModel(req, img) || kdh2NeedTyped(img.Info.Kind, req.Mode)) && normModel(req.ConfirmText) != img.Info.Confirm {
			return fmt.Errorf("il caricatore non controlla il modello del firmware che riceve: scrivi %s nella casella di conferma", img.Info.Confirm)
		}
		if req.Backup && req.Mode == "auto" {
			if _, ok := genByID[req.Model]; !ok {
				return errors.New("scegli il modello della radio per la copia di sicurezza")
			}
		}
		return nil
	case "rh5":
		if req.Mode != "update" {
			return errors.New("modalità sconosciuta")
		}
		if req.Baud != 115200 && req.Baud != 38400 {
			return errors.New("velocità non valida")
		}
		if !req.ConfirmModel {
			return errors.New("conferma che la radio collegata è una 5RH Pro (o compatibile, es. K67 Pro)")
		}
		return nil
	}
	if req.Mode != "update" && req.Mode != "recovery" {
		return errors.New("modalità sconosciuta")
	}
	if req.Baud != 115200 && req.Baud != 57600 {
		return errors.New("velocità non valida")
	}
	if !req.ConfirmModel {
		return fmt.Errorf("conferma che la radio collegata è una %s", img.Info.Target)
	}
	if (req.Mode == "recovery" || req.SkipIdentify) && normModel(req.ConfirmText) != img.Info.Confirm {
		return fmt.Errorf("senza l'identificazione automatica della radio, scrivi %s nella casella di conferma", img.Info.Confirm)
	}
	return nil
}

func (a *App) apiFwStart(r *http.Request) (interface{}, error) {
	var req fwStartReq
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.fw.mu.Lock()
	img := a.fw.img
	a.fw.mu.Unlock()
	if err := fwCheckStart(&req, img); err != nil {
		return nil, err
	}
	recovery := req.Mode == "recovery"
	st := a.store.Settings
	if img.Info.Kind == "anc" {
		st.BajPort, st.FwBaud = req.Port, req.Baud
	} else {
		st.GenPort = req.Port
	}
	if img.Info.Kind != "rh5" {
		st.FwNoBackup = !req.Backup
	}
	_ = a.store.SaveSettings(st)
	return a.startJob("firmware", true, func(ctx context.Context, j *cableJob) (interface{}, error) {
		return a.runFirmware(ctx, j, &req, img, recovery)
	})
}

func (a *App) runFirmware(ctx context.Context, j *cableJob, req *fwStartReq, img *fwImage, recovery bool) (interface{}, error) {
	switch img.Info.Kind {
	case "bf":
		return a.runFirmwareBF(ctx, j, req, img)
	case "rh5":
		return a.runFirmwareRH5(ctx, j, req, img)
	case "kdhx", "bj9000", "kdh9g":
		return a.runFirmwareKDH2(ctx, j, req, img)
	}
	res := map[string]interface{}{"firmware": img.Info.Name, "version": img.Info.Version}
	j.logf("Firmware: %s (%s, SHA-256 %s…)", img.Info.Name, fmtBytes(img.Info.Size), img.Info.SHA256[:16])
	if !recovery && !req.SkipIdentify {
		j.set("Identificazione della radio", true)
		var info btInfo
		var err error
		proto := "il protocollo di programmazione BT-11 di BJ-8300/BJ-7800"
		if img.Info.Confirm == "BJ7700" {
			proto = "il protocollo di programmazione della BJ-7700 (Radtel RT-910)"
		}
		if img.Info.Confirm == "BJ7700" {
			// BJ-7700: protocollo della RT-910 («PROGRAMBT80U»), anche per la copia
			m := genByID["bj-7700"]
			if req.Backup {
				j.logf("Lettura della radio con il driver della BJ-7700 (identificazione e copia di sicurezza)…")
				mem, _, err2 := m.Fam.Download(ctx, m, req.Port, j.prog)
				err = err2
				if err == nil {
					bp, err3 := a.saveGenBackup(m, mem, nil, "prima_del_firmware")
					if err3 != nil {
						return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): aggiornamento NON avviato", err3)
					}
					res["backup"] = bp
					j.note("Copia dei canali e delle impostazioni della radio salvata in %s", bp)
					j.logf("Copia di sicurezza salvata: %s", bp)
				}
			} else {
				j.logf("Identificazione della radio con il protocollo della BJ-7700…")
				err = rt900Identify(ctx, m, req.Port)
			}
			info.ID = "PROGRAMBT80U"
		} else if req.Backup && img.Info.Confirm == "BJ7800" {
			// BJ-7800: stesso protocollo di identificazione, memoria diversa: copia con il driver della BJ-7800
			m := genByID["bj-7800"]
			j.logf("Identificazione della radio (protocollo BT-11)…")
			if info, err = btIdentify(ctx, req.Port, j.prog); err == nil {
				select {
				case <-ctx.Done():
					return nil, errCancelled
				case <-time.After(1500 * time.Millisecond):
				}
				j.logf("Lettura della radio con il driver della BJ-7800 (copia di sicurezza)…")
				mem, _, err2 := m.Fam.Download(ctx, m, req.Port, j.prog)
				if err2 == errCancelled {
					return nil, err2
				}
				if err2 != nil {
					return nil, fmt.Errorf("copia di sicurezza non riuscita (%v): aggiornamento NON avviato. Riprova o togli la spunta della copia di sicurezza", err2)
				}
				bp, err3 := a.saveGenBackup(m, mem, nil, "prima_del_firmware")
				if err3 != nil {
					return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): aggiornamento NON avviato", err3)
				}
				res["backup"] = bp
				j.note("Copia dei canali e delle impostazioni della radio salvata in %s", bp)
				j.logf("Copia di sicurezza salvata: %s", bp)
			}
		} else if req.Backup {
			j.logf("Lettura della radio con il protocollo della Bajeton 8300 (identificazione e copia di sicurezza)…")
			var mem []byte
			mem, info, err = btReadImage(ctx, req.Port, j.prog, nil)
			if err == nil {
				if dat, err2 := eepromToDAT(mem, baseDAT); err2 == nil {
					if bp, err3 := a.saveBackup(kindDAT, dat, "prima_del_firmware"); err3 == nil {
						res["backup"] = bp
						j.note("Copia dei canali e delle impostazioni della radio salvata in %s", bp)
						j.logf("Copia di sicurezza salvata: %s", bp)
					} else {
						j.note("Copia di sicurezza non salvata: %v", err3)
					}
				} else {
					j.note("Radio letta, ma i dati non sono convertibili in un file .dat: %v", err2)
				}
			}
		} else {
			j.logf("Identificazione della radio con il protocollo della Bajeton 8300…")
			info, err = btIdentify(ctx, req.Port, j.prog)
		}
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("la radio collegata a %s non ha risposto come una %s (%s): aggiornamento NON avviato, la radio non è stata modificata. "+
				"Controlla che la radio sia accesa, il cavo e la porta. Se la radio non si avvia più, usa la modalità di recupero. Dettaglio: %v", req.Port, img.Info.Target, proto, err)
		}
		rid := strings.TrimSpace(info.ID)
		res["radioId"] = rid
		j.logf("Radio identificata: risponde al%s (identificativo «%s»)", strings.TrimPrefix(proto, "il"), rid)
		j.note("Radio identificata: risponde al%s (identificativo «%s»); modello confermato dall'utente: %s", strings.TrimPrefix(proto, "il"), rid, img.Info.Target)
		// la radio esce dalla modalità di programmazione
		select {
		case <-ctx.Done():
			return nil, errCancelled
		case <-time.After(1500 * time.Millisecond):
		}
	} else if recovery {
		j.logf("Modalità di recupero: la radio non può essere identificata, confermata dall'utente come %s", img.Info.Target)
	} else {
		j.logf("Identificazione automatica saltata: radio confermata dall'utente come %s", img.Info.Target)
	}
	if err := ctx.Err(); err != nil {
		return nil, errCancelled
	}
	j.set("Aggiornamento del firmware", recovery)
	started := false
	t0 := time.Now()
	sent, err := ancUpdate(ctx, req.Port, img.Dec, ancOpts{Recovery: recovery, Baud: req.Baud}, j.prog,
		func(s string) { j.logf("%s", s) },
		func() {
			started = true
			j.mu.Lock()
			j.CanStop = false
			j.mu.Unlock()
		})
	res["sent"] = sent
	if err == errCancelled {
		return nil, err
	}
	if err != nil {
		if !started {
			if recovery {
				return nil, fmt.Errorf("la radio non si è presentata in modalità aggiornamento entro il tempo previsto: nessun dato è stato inviato. Ripeti la procedura (%v)", err)
			}
			hint := ""
			if !req.SkipIdentify {
				hint = " Se la radio era stata identificata ma non entra in modalità aggiornamento, riprova dopo averla riaccesa spuntando «salta l'identificazione automatica»."
			}
			return nil, fmt.Errorf("la radio non ha risposto alla richiesta di aggiornamento: nessun dato è stato inviato e il firmware della radio non è cambiato. "+
				"Spegni e riaccendi la radio e riprova; se non si avvia più, usa la modalità di recupero.%s (%v)", hint, err)
		}
		return nil, fmt.Errorf("aggiornamento interrotto (%s inviati): la radio potrebbe non avviarsi. Non scollegare nulla, spegni e riaccendi la radio e ripeti l'aggiornamento; "+
			"se la radio non si avvia, ripetilo in modalità di recupero (%v)", fmtBytes(int(sent)), err)
	}
	res["seconds"] = int(time.Since(t0).Seconds())
	j.note("Firmware inviato: %s in %d secondi. Spegni e riaccendi la radio e controlla la versione all'accensione.", fmtBytes(int(sent)), int(time.Since(t0).Seconds()))
	return res, nil
}

// ---------------------------------------------------------------- Baofeng

// fwStarted: da qui l'aggiornamento non si può più interrompere
func fwStarted(j *cableJob, flag *bool) func() {
	return func() {
		*flag = true
		j.mu.Lock()
		j.CanStop = false
		j.mu.Unlock()
	}
}

func (a *App) runFirmwareBF(ctx context.Context, j *cableJob, req *fwStartReq, img *fwImage) (interface{}, error) {
	res := map[string]interface{}{"firmware": img.Info.Name, "version": img.Info.Version}
	j.logf("Firmware: %s (%s, SHA-256 %s…)", img.Info.Name, fmtBytes(img.Info.Size), img.Info.SHA256[:16])
	manual := req.Mode == "manual"
	if !manual && req.Backup {
		m := genByID[req.Model]
		j.set("Copia di sicurezza: lettura della radio", true)
		j.logf("Lettura della radio come %s (copia di canali e impostazioni)…", m.Name)
		mem, _, err := m.Fam.Download(ctx, m, req.Port, j.prog)
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("la radio collegata a %s non ha risposto come una %s: aggiornamento NON avviato, la radio non è stata modificata. Controlla radio, cavo e porta (%v)", req.Port, m.Name, err)
		}
		bp, err := a.saveGenBackup(m, mem, nil, "prima_del_firmware")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): aggiornamento NON avviato", err)
		}
		res["backup"] = bp
		j.note("Copia dei canali e delle impostazioni della radio salvata in %s", bp)
		j.logf("Copia di sicurezza salvata: %s", bp)
		select {
		case <-ctx.Done():
			return nil, errCancelled
		case <-time.After(1500 * time.Millisecond):
		}
	}
	if manual {
		j.logf("Modalità manuale: la radio deve essere già nel caricatore (accesa tenendo premuto il PTT)")
	}
	j.set("Aggiornamento del firmware", true)
	started := false
	t0 := time.Now()
	sent, err := kdhUpdate(ctx, req.Port, img.Raw, kdhOpts{Manual: manual}, j.prog, func(s string) { j.logf("%s", s) }, fwStarted(j, &started))
	res["sent"] = sent
	if err == errCancelled && !started {
		return nil, err
	}
	if err != nil {
		if !started {
			if manual {
				return nil, fmt.Errorf("il caricatore della radio non ha risposto: nessun dato è stato inviato. Spegni la radio, riaccendila tenendo premuto il PTT (schermo spento) e riprova (%v)", err)
			}
			return nil, fmt.Errorf("la radio non ha accettato la richiesta di aggiornamento: nessun dato è stato inviato e il firmware non è cambiato. Controlla che la radio sia accesa, il cavo e la porta, oppure usa la modalità manuale (%v)", err)
		}
		return nil, fmt.Errorf("aggiornamento interrotto (%s inviati): la radio potrebbe non avviarsi. Spegnila, riaccendila tenendo premuto il PTT e ripeti l'aggiornamento in modalità manuale (%v)", fmtBytes(int(sent)), err)
	}
	res["seconds"] = int(time.Since(t0).Seconds())
	j.note("Firmware inviato: %s in %d secondi. Spegni e riaccendi la radio e controlla la versione.", fmtBytes(int(sent)), int(time.Since(t0).Seconds()))
	return res, nil
}

func (a *App) runFirmwareRH5(ctx context.Context, j *cableJob, req *fwStartReq, img *fwImage) (interface{}, error) {
	res := map[string]interface{}{"firmware": img.Info.Name, "version": img.Info.Version}
	j.logf("Firmware: %s (%s, %s, SHA-256 %s…)", img.Info.Name, fmtBytes(img.Info.Size), img.Info.Hw, img.Info.SHA256[:16])
	j.set("Aggiornamento del firmware", true)
	j.logf("La radio deve essere in modalità aggiornamento (accesa tenendo premuti i tasti indicati dal costruttore)")
	started := false
	t0 := time.Now()
	sent, err := rh5Update(ctx, req.Port, img.Raw, req.Baud, j.prog, func(s string) { j.logf("%s", s) }, fwStarted(j, &started))
	res["sent"] = sent
	if err == errCancelled && !started {
		return nil, err
	}
	if err != nil {
		if !started {
			return nil, fmt.Errorf("aggiornamento non avviato, la radio non è stata modificata: %v", err)
		}
		return nil, fmt.Errorf("aggiornamento interrotto dopo la cancellazione della memoria (%s inviati): la radio non si avvierà finché non ripeti l'aggiornamento (riportala in modalità aggiornamento e riprova). %v", fmtBytes(int(sent)), err)
	}
	res["seconds"] = int(time.Since(t0).Seconds())
	j.note("Firmware inviato: %s in %d secondi. Spegni e riaccendi la radio e controlla la versione.", fmtBytes(int(sent)), int(time.Since(t0).Seconds()))
	return res, nil
}

// fwOtherModel: la radio scelta nella finestra (modello della copia di
// sicurezza) non è tra quelle a cui è destinato il file: conferma scritta
func fwOtherModel(req *fwStartReq, img *fwImage) bool {
	return req.Model != "" && len(img.Info.Models) > 0 && !slices.Contains(img.Info.Models, req.Model)
}

// kdh2NeedTyped: conferma scritta quando la radio non controlla il firmware
// (UV-9G sempre; caricatore «Bootloader Pro» nella modalità manuale, senza il
// codice H802; sempre se il firmware non è quello della UV-25 Plus: fwInfo.Typed)
func kdh2NeedTyped(kind, mode string) bool {
	return kind == "kdh9g" || (kind == "kdhx" && mode == "manual")
}

// kdh2Manual: istruzioni del costruttore per portare la radio nel caricatore
var kdh2Manual = map[string]string{
	"kdhx":   "radio spenta: tieni premuti SK2 e SK3 (i due tasti sotto il PTT) e accendila; l'indicatore diventa verde (procedura di Baofeng per la UV-25 Plus; per le altre radio segui le istruzioni fornite con il firmware)",
	"bj9000": "radio spenta: tieni premuti i tasti laterali 3 e 4 e accendila; compare la schermata «Update»",
	"kdh9g":  "radio spenta: tieni premuti i tasti 12 e 13 e accendila; si accende la torcia",
}

func (a *App) runFirmwareKDH2(ctx context.Context, j *cableJob, req *fwStartReq, img *fwImage) (interface{}, error) {
	res := map[string]interface{}{"firmware": img.Info.Name, "version": img.Info.Version}
	j.logf("Firmware: %s (%s, SHA-256 %s…)", img.Info.Name, fmtBytes(img.Info.Size), img.Info.SHA256[:16])
	manual := req.Mode == "manual"
	if !manual && req.Backup {
		m := genByID[req.Model]
		j.set("Copia di sicurezza: lettura della radio", true)
		j.logf("Lettura della radio come %s (copia di canali e impostazioni)…", m.Name)
		mem, _, err := m.Fam.Download(ctx, m, req.Port, j.prog)
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("la radio collegata a %s non ha risposto come una %s: aggiornamento NON avviato, la radio non è stata modificata. Controlla radio, cavo e porta (%v)", req.Port, m.Name, err)
		}
		bp, err := a.saveGenBackup(m, mem, nil, "prima_del_firmware")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): aggiornamento NON avviato", err)
		}
		res["backup"] = bp
		j.note("Copia dei canali e delle impostazioni della radio salvata in %s", bp)
		j.logf("Copia di sicurezza salvata: %s", bp)
		select {
		case <-ctx.Done():
			return nil, errCancelled
		case <-time.After(1500 * time.Millisecond):
		}
	}
	if manual {
		j.logf("Modalità manuale: la radio deve essere già nel caricatore (%s)", kdh2Manual[img.Info.Kind])
	}
	j.set("Aggiornamento del firmware", true)
	started := false
	t0 := time.Now()
	logf := func(s string) { j.logf("%s", s) }
	var sent int64
	var err error
	switch img.Info.Kind {
	case "kdh9g":
		sent, err = kdh9gUpdate(ctx, req.Port, img.Raw, manual, j.prog, logf, fwStarted(j, &started))
	case "bj9000":
		sent, err = kdh2Update(ctx, req.Port, img.Raw, bj9000Proto, manual, j.prog, logf, fwStarted(j, &started))
	default:
		sent, err = kdh2Update(ctx, req.Port, img.Raw, kdhxProto, manual, j.prog, logf, fwStarted(j, &started))
	}
	res["sent"] = sent
	if err == errCancelled && !started {
		return nil, err
	}
	how := kdh2Manual[img.Info.Kind]
	if err != nil {
		if !started {
			if manual {
				return nil, fmt.Errorf("il caricatore della radio non ha risposto: nessun dato è stato inviato. Spegni la radio, riportala nel caricatore (%s) e riprova (%v)", how, err)
			}
			return nil, fmt.Errorf("la radio non ha accettato la richiesta di aggiornamento: nessun dato è stato inviato e il firmware non è cambiato. Controlla che la radio sia accesa, il cavo e la porta, oppure usa la modalità manuale (%v)", err)
		}
		if sent == 0 {
			return nil, fmt.Errorf("aggiornamento non riuscito prima dell'invio dei dati: %v. Spegni e riaccendi la radio; se non si avvia, ripeti in modalità manuale (%s)", err, how)
		}
		return nil, fmt.Errorf("aggiornamento interrotto (%s inviati): la radio potrebbe non avviarsi. Spegnila, riportala nel caricatore (%s) e ripeti l'aggiornamento in modalità manuale (%v)", fmtBytes(int(sent)), how, err)
	}
	res["seconds"] = int(time.Since(t0).Seconds())
	j.note("Firmware inviato: %s in %d secondi. Spegni e riaccendi la radio e controlla la versione.", fmtBytes(int(sent)), int(time.Since(t0).Seconds()))
	return res, nil
}
