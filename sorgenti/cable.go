// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura e scrittura delle radio via cavo: operazioni in background con
// avanzamento, copia di sicurezza prima della scrittura e verifica finale.

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type cableJob struct {
	mu      sync.Mutex
	ID      string      `json:"id"`
	Op      string      `json:"op"`
	Radio   string      `json:"radio"`
	State   string      `json:"state"` // running | done | error | cancelled
	Phase   string      `json:"phase"`
	Done    int         `json:"done"`
	Total   int         `json:"total"`
	Msg     string      `json:"msg"`
	Error   string      `json:"error"`
	Notes   []string    `json:"notes"`
	Log     []string    `json:"log"` // messaggi in tempo reale (aggiornamento del firmware)
	Result  interface{} `json:"result"`
	CanStop bool        `json:"canStop"`
	cancel  context.CancelFunc
	started time.Time
}

func (j *cableJob) set(phase string, canStop bool) {
	j.mu.Lock()
	j.Phase, j.CanStop, j.Done, j.Total, j.Msg = phase, canStop, 0, 0, ""
	j.mu.Unlock()
}

func (j *cableJob) isRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.State == "running"
}

func (j *cableJob) prog(done, total int, msg string) {
	j.mu.Lock()
	j.Done, j.Total, j.Msg = done, total, msg
	j.mu.Unlock()
}

func (j *cableJob) note(f string, a ...interface{}) {
	j.mu.Lock()
	j.Notes = append(j.Notes, fmt.Sprintf(f, a...))
	j.mu.Unlock()
}

func (j *cableJob) logf(f string, a ...interface{}) {
	j.mu.Lock()
	j.Log = append(j.Log, time.Now().Format("15:04:05")+"  "+fmt.Sprintf(f, a...))
	j.mu.Unlock()
}

func (j *cableJob) snapshot() map[string]interface{} {
	j.mu.Lock()
	defer j.mu.Unlock()
	return map[string]interface{}{"id": j.ID, "op": j.Op, "radio": j.Radio, "state": j.State, "phase": j.Phase,
		"done": j.Done, "total": j.Total, "msg": j.Msg, "error": j.Error, "notes": j.Notes, "result": j.Result,
		"canStop": j.CanStop, "seconds": int(time.Since(j.started).Seconds()), "log": append([]string{}, j.Log...)}
}

type cableReq struct {
	Op       string  `json:"op"`    // read | write
	Radio    string  `json:"radio"` // k67 | baj
	Port     string  `json:"port"`
	Baud     int     `json:"baud"`
	Password string  `json:"password"`
	Backup   bool    `json:"backup"`
	Verify   bool    `json:"verify"`
	Channels []*Chan `json:"channels"`
}

func (a *App) backupDir() string {
	d := filepath.Join(a.store.dir, "copie_radio")
	_ = os.MkdirAll(d, 0o755)
	return d
}

func (a *App) saveBackup(kind string, data []byte, what string) (string, error) {
	ext, name := ".xlc", "K67"
	if kind == kindDAT {
		ext, name = ".dat", "Bajeton"
	}
	p := filepath.Join(a.backupDir(), fmt.Sprintf("%s_%s_%s%s", name, what, time.Now().Format("2006-01-02_15-04-05"), ext))
	return p, os.WriteFile(p, data, 0o644)
}

func (a *App) apiCablePorts(r *http.Request) (interface{}, error) {
	ports := listPortsFunc()
	if ports == nil {
		ports = []PortInfo{}
	}
	return map[string]interface{}{"ports": ports, "settings": a.store.Settings, "backupDir": a.backupDir()}, nil
}

func (a *App) apiCableStart(r *http.Request) (interface{}, error) {
	var req cableReq
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if isGenKind(req.Radio) {
		return a.genCableStart(&req)
	}
	if req.Radio != kindXLC && req.Radio != kindDAT {
		return nil, errors.New("radio sconosciuta")
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione")
	}
	a.mu.Lock()
	busy := a.job != nil && a.job.isRunning()
	sess := a.radios[req.Radio]
	a.mu.Unlock()
	if busy {
		return nil, errors.New("è già in corso un'operazione con la radio")
	}
	// preparazione dei dati da scrivere prima di toccare la radio
	var payload []byte
	var rep Report
	if req.Op == "write" {
		if sess == nil {
			return nil, errors.New("nessun file aperto per questa radio")
		}
		n := 0
		for _, c := range req.Channels {
			if c != nil {
				n++
			}
		}
		if n == 0 {
			return nil, errors.New("la radio deve avere almeno un canale")
		}
		file, err := writeRadio(req.Radio, sess.Base, req.Channels, &rep)
		if err != nil {
			return nil, err
		}
		if req.Radio == kindXLC {
			payload = file
		} else {
			if _, err := datToEEPROM(file); err != nil {
				return nil, fmt.Errorf("dati non scrivibili nella radio: %v", err)
			}
			payload = file
		}
	} else if req.Op != "read" {
		return nil, errors.New("operazione sconosciuta")
	}
	// impostazioni ricordate
	st := a.store.Settings
	if req.Radio == kindXLC {
		st.K67Port, st.K67Baud, st.K67Pwd = req.Port, req.Baud, req.Password
	} else {
		st.BajPort = req.Port
	}
	if req.Op == "write" {
		st.CableBackup, st.CableVerify = req.Backup, req.Verify
	}
	_ = a.store.SaveSettings(st)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: hex.EncodeToString(b), Op: req.Op, Radio: req.Radio, State: "running", cancel: cancel, started: time.Now(), Notes: []string{}}
	for _, w := range rep.Warnings {
		j.note("%s", w)
	}
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
		res, err := a.runCable(ctx, j, &req, payload)
		j.mu.Lock()
		defer j.mu.Unlock()
		switch {
		case err == errCancelled:
			j.State, j.Error = "cancelled", "Operazione annullata. Se la radio è rimasta in modalità programmazione, spegnila e riaccendila."
		case err != nil:
			j.State, j.Error = "error", err.Error()
		default:
			j.State, j.Result = "done", res
		}
		j.CanStop = false
	}()
	return j.snapshot(), nil
}

func (a *App) apiCableStatus(r *http.Request) (interface{}, error) {
	a.mu.Lock()
	j := a.job
	a.mu.Unlock()
	if j == nil {
		return map[string]interface{}{"state": "none"}, nil
	}
	return j.snapshot(), nil
}

func (a *App) apiCableCancel(r *http.Request) (interface{}, error) {
	a.mu.Lock()
	j := a.job
	a.mu.Unlock()
	if j == nil {
		return map[string]bool{"ok": true}, nil
	}
	j.mu.Lock()
	can := j.CanStop
	j.mu.Unlock()
	if !can {
		if j.Op == "firmware" {
			return nil, errors.New("l'aggiornamento del firmware non si può interrompere: la radio resterebbe senza un firmware completo")
		}
		return nil, errors.New("la scrittura non si può interrompere: la radio resterebbe con dati incompleti")
	}
	j.cancel()
	return map[string]bool{"ok": true}, nil
}

// lettura della radio (immagine + file nel formato del programma originale)
func (a *App) cableRead(ctx context.Context, j *cableJob, req *cableReq) (file []byte, info interface{}, k6 bool, err error) {
	if req.Radio == kindXLC {
		img, ki, err := k67Read(ctx, req.Port, &k67Opts{Password: req.Password, Baud: req.Baud}, j.prog)
		if err != nil {
			return nil, ki, false, err
		}
		return xlcFileFromRadio(img, ki.FreqBand, ki.K6, time.Now()), ki, ki.K6, nil
	}
	img, bi, err := btReadImage(ctx, req.Port, j.prog, nil)
	if err != nil {
		return nil, bi, false, err
	}
	dat, err := eepromToDAT(img, baseDAT)
	if err != nil {
		return nil, bi, false, fmt.Errorf("dati letti non convertibili: %v", err)
	}
	return dat, bi, false, nil
}

func (a *App) runCable(ctx context.Context, j *cableJob, req *cableReq, payload []byte) (interface{}, error) {
	if req.Op == "read" {
		j.set("Lettura della radio", true)
		file, info, k6, err := a.cableRead(ctx, j, req)
		if err != nil {
			return nil, err
		}
		rep := &Report{}
		kind, chs, err := readRadio(file, rep)
		if err != nil || kind != req.Radio {
			return nil, fmt.Errorf("dati letti non validi: %v", err)
		}
		bp, berr := a.saveBackup(req.Radio, file, "letta")
		if berr == nil {
			j.note("Copia della lettura salvata in %s", bp)
		}
		a.mu.Lock()
		a.radios[req.Radio] = &radioSession{Base: file, K6: k6}
		a.mu.Unlock()
		resp := radioResp{Kind: kind, Channels: chs, Warnings: rep.Warnings}
		if kind == kindXLC {
			resp.Bands = xlcBands(file, 0x10)
		}
		return map[string]interface{}{"radio": resp, "info": info, "backup": bp}, nil
	}

	// scrittura
	res := map[string]interface{}{}
	if req.Backup {
		j.set("Copia di sicurezza: lettura della radio", true)
		file, _, _, err := a.cableRead(ctx, j, req)
		if err != nil {
			return nil, fmt.Errorf("copia di sicurezza non riuscita, nessun dato è stato scritto. %v", err)
		}
		bp, err := a.saveBackup(req.Radio, file, "prima_della_scrittura")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): nessun dato è stato scritto", err)
		}
		res["backup"] = bp
		j.note("Copia di sicurezza della radio salvata in %s", bp)
	}
	if err := ctx.Err(); err != nil {
		return nil, errCancelled
	}
	j.set("Scrittura nella radio", false)
	var written []byte
	if req.Radio == kindXLC {
		a.mu.Lock()
		sess := a.radios[kindXLC]
		a.mu.Unlock()
		k6 := sess != nil && sess.K6
		written = xlcImageForRadio(payload)
		ki, err := k67Write(context.Background(), req.Port, written, &k67Opts{Password: req.Password, Baud: req.Baud,
			FreqBand: payload[xlcSize-1], K6: k6}, j.prog)
		res["info"] = ki
		if err != nil {
			return nil, err
		}
	} else {
		img, err := datToEEPROM(payload)
		if err != nil {
			return nil, err
		}
		written = img
		bi, err := btWriteImage(context.Background(), req.Port, img, j.prog, nil)
		res["info"] = bi
		if err != nil {
			return nil, err
		}
	}
	if req.Verify {
		j.set("Verifica: rilettura della radio", true)
		file, _, _, err := a.cableRead(ctx, j, req)
		if err == errCancelled {
			j.note("Verifica annullata: la scrittura era già completata.")
		} else if err != nil {
			j.note("Scrittura completata, ma la rilettura di verifica non è riuscita: %v", err)
			res["verified"] = false
		} else {
			bad := verifyWritten(req.Radio, written, file)
			res["verified"] = len(bad) == 0
			if len(bad) == 0 {
				j.note("Verifica: i canali riletti dalla radio corrispondono a quelli scritti.")
			} else {
				j.note("Verifica: differenze nei canali %s", strings.Join(bad, ", "))
			}
		}
	}
	return res, nil
}

// verifyWritten confronta i canali scritti con quelli riletti.
func verifyWritten(kind string, written, readFile []byte) []string {
	var bad []string
	if kind == kindXLC {
		for i := 0; i < xlcChannels; i++ {
			o := xlcChBase + xlcChSize*i
			if !bytes.Equal(written[o:o+xlcChSize], readFile[o:o+xlcChSize]) ||
				bitGet(written, xlcValid, i) != bitGet(readFile, xlcValid, i) || bitGet(written, xlcScanFlg, i) != bitGet(readFile, xlcScanFlg, i) {
				bad = append(bad, fmt.Sprint(i+1))
			}
		}
		return bad
	}
	img, err := datToEEPROM(readFile)
	if err != nil {
		return []string{"(rilettura non convertibile)"}
	}
	for i := 0; i < datZones*datPerZone; i++ {
		if !bytes.Equal(written[i*32:i*32+32], img[i*32:i*32+32]) {
			bad = append(bad, fmt.Sprintf("%d/%d", i/datPerZone+1, i%datPerZone+1))
		}
	}
	return bad
}
