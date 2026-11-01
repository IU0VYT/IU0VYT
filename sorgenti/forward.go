// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// «Inoltra ad altra radio» (3.9.3): i canali scelti nella scheda di una radio
// vengono scritti in un'altra radio collegata con il cavo. In un'unica
// operazione: lettura della radio di destinazione (salvata come copia di
// sicurezza: servono anche le sue impostazioni e i canali già presenti),
// adattamento dei canali alla radio, inserimento nelle posizioni libere (oppure al
// posto di tutti i canali), scrittura e rilettura di verifica. Se qualcosa non va
// prima della scrittura, nella radio non viene scritto nulla.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// forwardPlan: canali della radio dopo l'inoltro e posizioni usate
type forwardPlan struct {
	chans   []*Chan
	placed  []int
	skipped int // canali non trasferibili o senza posizione adatta
}

// planForward adatta i canali alla radio "kind" e li inserisce nei canali letti
// dalla radio (cur). mode: "free" (posizioni libere, i canali presenti restano)
// oppure "replace" (tutti i canali della radio sostituiti da quelli inoltrati).
// mem: memoria letta (radio generiche, per le posizioni bloccate); bands: gamme
// del K67.
func planForward(kind string, cur, src []*Chan, mode string, mem []byte, bands []band, rep *Report) (*forwardPlan, error) {
	m := genByID[kind]
	final := append([]*Chan{}, cur...)
	// posizioni che la radio non permette di cambiare (es. canali GMRS fissi)
	var lk interface {
		Locked(m *genModel, i int, old, want *Chan) error
	}
	if m != nil {
		lk, _ = m.Fam.(interface {
			Locked(m *genModel, i int, old, want *Chan) error
		})
	}
	fits := func(i int, c *Chan) bool {
		if m == nil {
			return true
		}
		if lk != nil && lk.Locked(m, i, genGet(m, mem, i), c) != nil {
			return false
		}
		return c == nil || genValidate(m, c, i) == nil
	}
	if mode == "replace" {
		kept := 0
		for i := range final {
			if final[i] != nil && fits(i, nil) {
				final[i] = nil
			} else if final[i] != nil {
				kept++
			}
		}
		if kept > 0 {
			rep.warn("%d canali fissi della radio non si possono cancellare: restano al loro posto", kept)
		}
	} else if mode != "free" {
		return nil, errors.New("modo di inserimento sconosciuto")
	}
	p := &forwardPlan{}
	next := 0
	for _, c := range src {
		if c == nil {
			continue
		}
		label := fmt.Sprintf("«%s» (%s MHz)", strings.TrimSpace(c.Name), fmtMHz(c.Rx))
		var a *Chan
		if m != nil {
			a = genAdapt(c, m, label, rep)
		} else {
			a = adaptChan(c, kind, bands, label, rep)
		}
		if a == nil {
			p.skipped++
			continue
		}
		placed := false
		for i := next; i < len(final); i++ {
			if final[i] == nil && fits(i, a) {
				final[i] = a
				p.placed = append(p.placed, i)
				placed, next = true, i+1
				break
			}
		}
		if !placed && next > 0 {
			// posizioni saltate perché non adatte a un canale precedente
			for i := 0; i < next; i++ {
				if final[i] == nil && fits(i, a) {
					final[i] = a
					p.placed = append(p.placed, i)
					placed = true
					break
				}
			}
		}
		if !placed {
			free := 0
			for i := range final {
				if final[i] == nil {
					free++
				}
			}
			if free == 0 {
				return nil, fmt.Errorf("nella %s non ci sono abbastanza posizioni libere: ne servono %d, ne restavano %d. "+
					"Nulla è stato scritto: libera qualche posizione oppure scegli «sostituisci tutti i canali»",
					radioName(kind), countChans(src), len(p.placed))
			}
			rep.warn("%s: non può stare in nessuna posizione libera della %s, escluso", label, radioName(kind))
			p.skipped++
		}
	}
	if len(p.placed) == 0 {
		return nil, fmt.Errorf("nessuno dei canali scelti si può usare sulla %s: nulla è stato scritto", radioName(kind))
	}
	p.chans = final
	return p, nil
}

func countChans(cs []*Chan) int {
	n := 0
	for _, c := range cs {
		if c != nil {
			n++
		}
	}
	return n
}

// posizioni in forma leggibile: «1-5, 8, 12-13»
func fmtPositions(kind string, pos []int) string {
	first := 1
	if m, ok := genByID[kind]; ok {
		first = m.First
	}
	label := func(i int) string {
		if kind == kindDAT {
			return fmt.Sprintf("%d/%d", i/datPerZone+1, i%datPerZone+1)
		}
		return fmt.Sprint(i + first)
	}
	var parts []string
	for k := 0; k < len(pos); {
		j := k
		for j+1 < len(pos) && pos[j+1] == pos[j]+1 && (kind != kindDAT || pos[j+1]%datPerZone != 0) {
			j++
		}
		if j > k {
			parts = append(parts, label(pos[k])+"-"+label(pos[j]))
		} else {
			parts = append(parts, label(pos[k]))
		}
		k = j + 1
	}
	return strings.Join(parts, ", ")
}

func (a *App) forwardStart(req *cableReq) (interface{}, error) {
	m, gen := genByID[req.Radio]
	if !gen && req.Radio != kindXLC && req.Radio != kindDAT {
		return nil, errors.New("radio sconosciuta")
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione")
	}
	if countChans(req.Channels) == 0 {
		return nil, errors.New("scegli prima i canali da inoltrare")
	}
	if req.Mode == "" {
		req.Mode = "free"
	}
	if req.Mode != "free" && req.Mode != "replace" {
		return nil, errors.New("modo di inserimento sconosciuto")
	}
	st := a.store.Get()
	switch {
	case gen && m.Conn == "":
		st.GenPort = req.Port
	case req.Radio == kindXLC:
		st.K67Port, st.K67Baud, st.K67Pwd = req.Port, req.Baud, req.Password
	case req.Radio == kindDAT:
		st.BajPort = req.Port
	}
	_ = a.store.SaveSettings(st)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: hex.EncodeToString(b), Op: "forward", Radio: req.Radio, State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}}
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
			res, err = a.runForward(ctx, j, req)
		}()
		j.mu.Lock()
		defer j.mu.Unlock()
		switch {
		case err == errCancelled:
			j.State, j.Error = "cancelled", "Operazione annullata: nella radio non è stato scritto nulla. Se la radio è rimasta in modalità programmazione, spegnila e riaccendila."
		case err != nil:
			j.State, j.Error = "error", err.Error()
		default:
			j.State, j.Result = "done", res
		}
		j.CanStop = false
	}()
	return j.snapshot(), nil
}

func (a *App) runForward(ctx context.Context, j *cableJob, req *cableReq) (interface{}, error) {
	res := map[string]interface{}{}
	rep := &Report{}
	wreq := &cableReq{Op: "write", Radio: req.Radio, Port: req.Port, Baud: req.Baud, Password: req.Password, Verify: true}
	var payload []byte
	var plan *forwardPlan
	var resp interface{}
	if m, gen := genByID[req.Radio]; gen {
		// 1. lettura della radio di destinazione = copia di sicurezza
		j.set("Copia di sicurezza: lettura della radio di destinazione", true)
		mem, note, err := m.Fam.Download(ctx, m, req.Port, j.prog)
		if note != "" {
			j.note("%s", note)
		}
		if err == nil && len(mem) < genMinSize(m) {
			err = fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
		}
		if err != nil {
			if err == errCancelled || ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, fmt.Errorf("lettura della radio di destinazione non riuscita, nulla è stato scritto. %v", err)
		}
		a.mu.Lock()
		var meta map[string]interface{}
		sess := a.radios[m.ID]
		if sess != nil {
			meta = sess.Meta
		}
		a.mu.Unlock()
		bp, err := a.saveGenBackup(m, mem, meta, "prima_dell_inoltro")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): nulla è stato scritto", err)
		}
		res["backup"] = bp
		j.note("Copia di sicurezza della radio di destinazione salvata in %s", bp)
		// 2. adattamento e posizioni
		plan, err = planForward(m.ID, genChans(m, mem), req.Channels, req.Mode, mem, nil, rep)
		if err != nil {
			notesFrom(j, rep)
			return nil, err
		}
		if payload, err = genWrite(m, mem, plan.chans, rep); err != nil {
			return nil, fmt.Errorf("canali non scrivibili nella radio, nulla è stato scritto: %v", err)
		}
		notesFrom(j, rep)
		// come CHIRP: tra due collegamenti la radio deve uscire dalla programmazione
		select {
		case <-ctx.Done():
			return nil, errCancelled
		case <-time.After(1500 * time.Millisecond):
		}
		// 3. scrittura e verifica (la sessione della scheda aperta resta com'era)
		var prev []byte
		if sess != nil {
			prev = sess.Base
		}
		wr, err := a.runGenCable(ctx, j, m, wreq, payload, meta)
		a.mu.Lock()
		if s := a.radios[m.ID]; s != nil {
			s.Base = prev
		}
		if err == nil && req.Adopt {
			a.radios[m.ID] = &radioSession{Base: payload, Meta: meta}
		}
		a.mu.Unlock()
		if err != nil {
			return nil, err
		}
		for k, v := range wr.(map[string]interface{}) {
			if k != "backup" {
				res[k] = v
			}
		}
		resp = genResp{Kind: m.ID, Channels: genChans(m, payload), Def: m.def(), Warnings: []string{}}
	} else {
		// 1. lettura della radio di destinazione = copia di sicurezza
		j.set("Copia di sicurezza: lettura della radio di destinazione", true)
		file, _, k6, err := a.cableRead(ctx, j, req)
		if err != nil {
			if err == errCancelled || ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, fmt.Errorf("lettura della radio di destinazione non riuscita, nulla è stato scritto. %v", err)
		}
		kind, cur, err := readRadio(file, &Report{})
		if err != nil || kind != req.Radio {
			return nil, fmt.Errorf("dati letti dalla radio non validi, nulla è stato scritto: %v", err)
		}
		bp, err := a.saveBackup(req.Radio, file, "prima_dell_inoltro")
		if err != nil {
			return nil, fmt.Errorf("impossibile salvare la copia di sicurezza (%v): nulla è stato scritto", err)
		}
		res["backup"] = bp
		j.note("Copia di sicurezza della radio di destinazione salvata in %s", bp)
		// 2. adattamento e posizioni
		var bands []band
		if kind == kindXLC {
			bands = xlcBands(file, 0x10)
		}
		plan, err = planForward(kind, cur, req.Channels, req.Mode, nil, bands, rep)
		if err != nil {
			notesFrom(j, rep)
			return nil, err
		}
		if payload, err = writeRadio(kind, file, plan.chans, rep); err != nil {
			return nil, fmt.Errorf("canali non scrivibili nella radio, nulla è stato scritto: %v", err)
		}
		if kind == kindDAT {
			if _, err := datToEEPROM(payload); err != nil {
				return nil, fmt.Errorf("dati non scrivibili nella radio, nulla è stato scritto: %v", err)
			}
		}
		notesFrom(j, rep)
		if err := ctx.Err(); err != nil {
			return nil, errCancelled
		}
		// 3. scrittura e verifica
		wr, err := a.runCable(ctx, j, wreq, payload)
		if err != nil {
			return nil, err
		}
		for k, v := range wr.(map[string]interface{}) {
			if k != "backup" {
				res[k] = v
			}
		}
		chs, _ := readRadioKind(kind, payload)
		r := radioResp{Kind: kind, Channels: chs, Warnings: []string{}}
		if kind == kindXLC {
			r.Bands = xlcBands(payload, 0x10)
		} else {
			r.DatVer, r.Single = datVer(payload), datSingleList(payload)
		}
		resp = r
		if req.Adopt {
			a.mu.Lock()
			a.radios[kind] = &radioSession{Base: payload, K6: k6}
			a.mu.Unlock()
		}
	}
	j.note("Inoltrati %d canali nelle posizioni %s.", len(plan.placed), fmtPositions(req.Radio, plan.placed))
	if plan.skipped > 0 {
		j.note("%d canali non sono stati inoltrati: vedi le segnalazioni.", plan.skipped)
	}
	res["placed"] = plan.placed
	res["count"] = len(plan.placed)
	res["skipped"] = plan.skipped
	res["radio"] = resp
	res["adopted"] = req.Adopt
	return res, nil
}

// notesFrom: le segnalazioni dell'adattamento diventano note dell'operazione
// (anche quando l'inoltro si ferma prima della scrittura)
func notesFrom(j *cableJob, rep *Report) {
	for _, w := range rep.Warnings {
		j.note("%s", w)
	}
}

// readRadioKind: canali di un file .xlc / .dat appena preparato
func readRadioKind(kind string, b []byte) ([]*Chan, error) {
	k, chs, err := readRadio(b, &Report{})
	if err != nil {
		return nil, err
	}
	if k != kind {
		return nil, fmt.Errorf("formato inatteso")
	}
	return chs, nil
}
