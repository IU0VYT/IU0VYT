// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// API per le radio gestite con i driver portati da CHIRP ("generiche"):
// catalogo dei modelli, apertura e salvataggio dei file .img (compatibili con
// CHIRP), adattamento dei canali, lettura e scrittura via cavo con copia di
// sicurezza e verifica.

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
	"regexp"
	"strings"
	"time"
)

const imgFilter = "Immagine della radio, compatibile con CHIRP (*.img)|*.img|Tutti i file|*.*"

// definizioni dei modelli per l'interfaccia (galleria e schede)
func (a *App) apiCatalog(r *http.Request) (interface{}, error) {
	defs := map[string]genDef{}
	for _, m := range genModels {
		defs[m.ID] = m.def()
	}
	return map[string]interface{}{"catalog": catalog, "defs": defs, "noDriver": catalogNoDriver}, nil
}

func genModelOf(kind string) (*genModel, error) {
	m, ok := genByID[kind]
	if !ok {
		return nil, errors.New("radio sconosciuta")
	}
	return m, nil
}

type genResp struct {
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	Channels []*Chan  `json:"channels"`
	Warnings []string `json:"warnings"`
	Def      genDef   `json:"def"`
}

// genOpen: file .img (anche salvato da CHIRP) per la radio indicata (o riconosciuta)
func (a *App) genOpen(p string, b []byte, prefer string) (interface{}, error) {
	m, mem, meta, err := detectGen(b, prefer)
	if err != nil {
		return nil, err
	}
	rep := &Report{}
	if prefer != "" && isGenKind(prefer) && m.ID != prefer {
		pm := genByID[prefer]
		if pm.Fam2() == m.Fam2() || fmt.Sprintf("%T", pm.Fam) == fmt.Sprintf("%T", m.Fam) && pm.RClass == m.RClass {
			m = pm
		} else {
			rep.warn("Il file è di un %s, non di un %s: aperto come %s", m.Name, pm.Name, m.Name)
		}
	}
	if len(mem) < genMinSize(m) {
		return nil, fmt.Errorf("file incompleto: %d byte, troppo pochi per un %s", len(mem), m.Name)
	}
	a.mu.Lock()
	a.radios[m.ID] = &radioSession{Path: p, Base: mem, Meta: meta}
	a.mu.Unlock()
	return genResp{Kind: m.ID, Path: p, Channels: genChans(m, mem), Warnings: rep.Warnings, Def: m.def()}, nil
}

// dimensione minima dell'immagine: deve contenere tutti i canali
func genMinSize(m *genModel) int {
	if lf, ok := m.Fam.(interface{ layout(m *genModel) *bwEl }); ok {
		return lf.layout(m).ByteLen()
	}
	return 0
}

var reBadFileChar = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func genFileBase(m *genModel) string {
	return strings.Trim(reBadFileChar.ReplaceAllString(m.Vendor+"_"+m.Model, "_"), "_")
}

func (a *App) genSave(kind string, chs []*Chan, saveAs bool, path string) (interface{}, error) {
	m, err := genModelOf(kind)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	sess := a.radios[kind]
	a.mu.Unlock()
	if sess == nil {
		return nil, errors.New("nessun file aperto per questa radio")
	}
	p := path
	if p == "" && !saveAs {
		p = sess.Path
	}
	if p == "" {
		suggest := genFileBase(m) + ".img"
		if sess.Path != "" {
			suggest = sess.Path
		} else if a.store.Settings.LastDir != "" {
			suggest = filepath.Join(a.store.Settings.LastDir, suggest)
		}
		if p, err = a.dialog(true, "Salva il file della "+m.Name, imgFilter, suggest, "img"); err != nil {
			return nil, err
		}
	}
	rep := &Report{}
	mem, err := genWrite(m, sess.Base, chs, rep)
	if err != nil {
		return nil, err
	}
	if err := verifySaved(chs, genChans(m, mem)); err != nil {
		return nil, fmt.Errorf("errore interno, file non salvato: %v", err)
	}
	out := makeImg(m, mem, sess.Meta)
	if old, err := os.ReadFile(p); err == nil && !bytes.Equal(old, out) {
		_ = os.WriteFile(p+".bak", old, 0o644)
	}
	if err := writeAtomic(p, out); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	a.mu.Lock()
	a.radios[kind] = &radioSession{Path: p, Base: mem, Meta: sess.Meta}
	a.mu.Unlock()
	return map[string]interface{}{"path": p, "warnings": rep.Warnings}, nil
}

// genAdapt prepara un canale (da un'altra radio o dall'archivio) per il modello m.
func genAdapt(c *Chan, m *genModel, label string, rep *Report) *Chan {
	n := *c
	if _, ok := genUnraw(m, n.Raw); !ok {
		n.Raw = "" // i dati non interpretati appartengono a un'altra radio
	}
	name := m.cleanName(n.Name)
	// nome troppo lungo: prima si tolgono gli spazi («PMR CH 12» → «PMRCH12»),
	// così i nomi restano distinti anche sulle radio con pochi caratteri
	if m.NameMax > 0 && len(strings.TrimSpace(asciiName(n.Name))) > m.NameMax {
		if alt := m.cleanName(strings.ReplaceAll(strings.TrimSpace(n.Name), " ", "")); alt != "" {
			name = alt
		}
	}
	if name != n.Name {
		rep.warn("%s: nome «%s» → «%s»", label, n.Name, name)
	}
	n.Name = name
	if len(m.Powers) > 0 {
		k := m.powerIndex(n.Power)
		if m.Powers[k].Sem != n.Power {
			rep.warn("%s: potenza %s non disponibile, impostata %s", label, powerName(n.Power), m.Powers[k].Label)
		}
		n.Power = m.Powers[k].Sem
	}
	if n.Narrow && !m.Narrow {
		rep.warn("%s: banda stretta non disponibile, impostata larga", label)
		n.Narrow = false
	}
	if !m.HasScan {
		n.Scan = true
	}
	if n.Dig && m.Digital == "" {
		rep.warn("%s: canale digitale, ma la %s ha solo canali analogici: canale non copiato", label, m.Name)
		return nil
	}
	if n.TxOff && m.NoTxOff {
		rep.warn("%s: questa radio non permette canali di sola ricezione: canale non copiato", label)
		return nil
	}
	if n.TxOff {
		n.Tx = n.Rx
	}
	for _, t := range []*string{&n.TxTone, &n.RxTone} {
		if *t == "" {
			continue
		}
		if v, ok := parseTone(*t); ok {
			*t = toneStr(v)
			continue
		}
		code, inv, ok := dcsParse(*t)
		if ok && m.dcsOK(code) {
			s := "D" + code + "N"
			if inv {
				s = "D" + code + "I"
			}
			*t = s
			continue
		}
		rep.warn("%s: tono «%s» non disponibile su questa radio, impostato OFF", label, *t)
		*t = ""
	}
	if m.OneDCS {
		tc, tinv, tok := dcsParse(n.TxTone)
		rc, _, rok := dcsParse(n.RxTone)
		if tok && rok && tc != rc {
			s := "D" + rc + "N"
			if tinv {
				s = "D" + rc + "I"
			}
			rep.warn("%s: la %s usa lo stesso codice DCS in trasmissione e in ricezione: TX %s → %s", label, m.Name, n.TxTone, s)
			n.TxTone = s
		}
	}
	if !m.inBands(n.Rx) {
		rep.warn("%s: %s MHz fuori dalle gamme della radio", label, fmtMHz(n.Rx))
	}
	return &n
}

// ---------------------------------------------------------------- cavo

func (a *App) saveGenBackup(m *genModel, mem []byte, meta map[string]interface{}, what string) (string, error) {
	p := filepath.Join(a.backupDir(), fmt.Sprintf("%s_%s_%s.img", genFileBase(m), what, time.Now().Format("2006-01-02_15-04-05")))
	return p, os.WriteFile(p, makeImg(m, mem, meta), 0o644)
}

func (a *App) genCableStart(req *cableReq) (interface{}, error) {
	m, err := genModelOf(req.Radio)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione")
	}
	a.mu.Lock()
	sess := a.radios[req.Radio]
	a.mu.Unlock()
	var payload []byte
	var rep Report
	switch req.Op {
	case "write":
		if sess == nil {
			return nil, errors.New("leggi prima la radio o apri un suo file .img: le impostazioni della radio vengono conservate e si modificano solo i canali")
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
		if payload, err = genWrite(m, sess.Base, req.Channels, &rep); err != nil {
			return nil, err
		}
	case "read":
	default:
		return nil, errors.New("operazione sconosciuta")
	}
	st := a.store.Settings
	if m.Conn == "" {
		st.GenPort = req.Port
	}
	if req.Op == "write" {
		st.CableBackup, st.CableVerify = req.Backup, req.Verify
	}
	_ = a.store.SaveSettings(st)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: hex.EncodeToString(b), Op: req.Op, Radio: req.Radio, State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}}
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
	var meta map[string]interface{}
	if sess != nil {
		meta = sess.Meta
	}
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
			res, err = a.runGenCable(ctx, j, m, req, payload, meta)
		}()
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

func (a *App) runGenCable(ctx context.Context, j *cableJob, m *genModel, req *cableReq, payload []byte, meta map[string]interface{}) (interface{}, error) {
	read := func(phase string) ([]byte, error) {
		j.set(phase, true)
		mem, note, err := m.Fam.Download(ctx, m, req.Port, j.prog)
		if err != nil {
			return nil, err
		}
		if note != "" {
			j.note("%s", note)
		}
		if len(mem) < genMinSize(m) {
			return nil, fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
		}
		return mem, nil
	}
	if req.Op == "read" {
		mem, err := read("Lettura della radio")
		if err != nil {
			return nil, err
		}
		bp, berr := a.saveGenBackup(m, mem, meta, "letta")
		if berr == nil {
			j.note("Copia della lettura salvata in %s", bp)
		}
		a.mu.Lock()
		a.radios[m.ID] = &radioSession{Base: mem, Meta: meta}
		a.mu.Unlock()
		resp := genResp{Kind: m.ID, Channels: genChans(m, mem), Def: m.def(), Warnings: []string{}}
		return map[string]interface{}{"radio": resp, "backup": bp}, nil
	}
	res := map[string]interface{}{}
	if req.Backup {
		mem, err := read("Copia di sicurezza: lettura della radio")
		if err != nil {
			return nil, fmt.Errorf("copia di sicurezza non riuscita, nessun dato è stato scritto. %v", err)
		}
		bp, err := a.saveGenBackup(m, mem, meta, "prima_della_scrittura")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): nessun dato è stato scritto", err)
		}
		res["backup"] = bp
		j.note("Copia di sicurezza della radio salvata in %s", bp)
		// come CHIRP: tra due collegamenti la radio deve uscire dalla programmazione
		select {
		case <-ctx.Done():
			return nil, errCancelled
		case <-time.After(1500 * time.Millisecond):
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, errCancelled
	}
	j.set("Scrittura nella radio", false)
	note, err := m.Fam.Upload(context.Background(), m, req.Port, payload, j.prog)
	if note != "" {
		j.note("%s", note)
	}
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	if s := a.radios[m.ID]; s != nil {
		s.Base = payload
	}
	a.mu.Unlock()
	if req.Verify {
		select {
		case <-ctx.Done():
		case <-time.After(1500 * time.Millisecond):
		}
		mem, err := read("Verifica: rilettura della radio")
		switch {
		case err == errCancelled:
			j.note("Verifica annullata: la scrittura era già completata.")
		case err != nil:
			j.note("Scrittura completata, ma la rilettura di verifica non è riuscita: %v", err)
			res["verified"] = false
		default:
			var bad []string
			for i := 0; i < m.Slots; i++ {
				w, g := m.Fam.Get(m, payload, i), m.Fam.Get(m, mem, i)
				if !chanEqual(w, g) {
					bad = append(bad, fmt.Sprint(i+m.First))
				}
			}
			res["verified"] = len(bad) == 0
			if len(bad) == 0 {
				j.note("Verifica: i canali riletti dalla radio corrispondono a quelli scritti.")
			} else {
				if len(bad) > 20 {
					bad = append(bad[:20], "…")
				}
				j.note("Verifica: differenze nei canali %s", strings.Join(bad, ", "))
			}
		}
	}
	return res, nil
}
