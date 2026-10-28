// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// API del convertitore universale e dell'importazione da qualsiasi file (3.7).

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const anyFilter = "Tutti i formati (*.xlc, *.dat, *.img, *.csv, *.xlsx, *.xls)|*.xlc;*.dat;*.img;*.csv;*.xlsx;*.xls;*.txt|" +
	"Baofeng K67 Pro (*.xlc)|*.xlc|Bajeton BJ-8300 (*.dat)|*.dat|Immagini delle radio, anche di CHIRP (*.img)|*.img|" +
	"CSV di CHIRP o tabella CSV (*.csv)|*.csv|Excel (*.xlsx, *.xls)|*.xlsx;*.xls|Tutti i file|*.*"

// canali presenti (senza le posizioni vuote), per l'inserimento in una radio
func packed(chs []*Chan) []*Chan {
	out := []*Chan{}
	for _, c := range chs {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}

func (a *App) openAny(path, title, prefer string) (string, *uniSource, error) {
	p := path
	if p == "" {
		var err error
		if p, err = a.dialog(false, title, anyFilter, "", ""); err != nil {
			return "", nil, err
		}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", nil, fmt.Errorf("impossibile leggere il file: %v", err)
	}
	src, err := readAny(b, filepath.Base(p), prefer)
	if err != nil {
		return "", nil, err
	}
	if src.Warnings == nil {
		src.Warnings = []string{}
	}
	return p, src, nil
}

// convert/open: file di partenza del convertitore
func (a *App) apiConvertOpen(r *http.Request) (interface{}, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	p, src, err := a.openAny(req.Path, "Converti: scegli il file di partenza", "")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"path": p, "name": filepath.Base(p), "kind": src.Kind, "desc": src.Desc, "count": src.count(),
		"channels": src.Channels, "warnings": src.Warnings}, nil
}

// convert/targets: formati di destinazione (e per le radio .img se c'è una scheda aperta)
func (a *App) apiConvertTargets(r *http.Request) (interface{}, error) {
	ts := uniTargets()
	a.mu.Lock()
	tabs := map[string]bool{}
	for k, s := range a.radios {
		if s != nil && s.Base != nil {
			tabs[k] = true
		}
	}
	a.mu.Unlock()
	out := make([]map[string]interface{}, 0, len(ts))
	for _, t := range ts {
		tab := tabs[t.ID] || (strings.HasPrefix(t.ID, "baj-") && tabs[kindDAT])
		out = append(out, map[string]interface{}{"id": t.ID, "name": t.Name, "group": t.Group, "ext": t.Ext, "needsBase": t.NeedsBase,
			"slots": t.Slots, "tab": tab})
	}
	return map[string]interface{}{"targets": out}, nil
}

// convert/run: scrive i canali nel formato scelto e salva il file
func (a *App) apiConvertRun(r *http.Request) (interface{}, error) {
	var req struct {
		Channels []*Chan `json:"channels"`
		Target   string  `json:"target"`
		Base     string  `json:"base"` // "" | tab | file
		BasePath string  `json:"basePath"`
		Pack     bool    `json:"pack"`
		Name     string  `json:"name"` // nome del file di partenza (per il nome proposto)
		Path     string  `json:"path"` // file di uscita (prove automatiche)
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if len(packed(req.Channels)) == 0 {
		return nil, errors.New("nessun canale da convertire")
	}
	var tgt *uniTarget
	for _, t := range uniTargets() {
		if t.ID == req.Target {
			t := t
			tgt = &t
		}
	}
	if tgt == nil {
		return nil, errors.New("scegli il formato di destinazione")
	}
	var base []byte
	switch req.Base {
	case "tab":
		kind := req.Target
		if strings.HasPrefix(kind, "baj-") {
			kind = kindDAT
		}
		a.mu.Lock()
		if s := a.radios[kind]; s != nil {
			base = append([]byte{}, s.Base...)
			if m, ok := genByID[kind]; ok {
				base = makeImg(m, base, s.Meta)
			}
		}
		a.mu.Unlock()
		if base == nil {
			return nil, errors.New("la scheda di questa radio non ha dati: leggi la radio o apri un suo file")
		}
	case "file":
		p := req.BasePath
		if p == "" {
			var err error
			filt := map[string]string{"img": imgFilter, "xlc": radioFilter[kindXLC], "dat": radioFilter[kindDAT]}[tgt.Ext]
			if p, err = a.dialog(false, "File della radio da usare come base", filt, "", ""); err != nil {
				return nil, err
			}
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("impossibile leggere il file di base: %v", err)
		}
		base = b
	}
	var band *byte
	if st := a.store.Settings; st.K67Known {
		b := st.K67Band
		band = &b
	}
	out, ext, rep, err := uniConvert(req.Channels, req.Target, base, req.Pack, band)
	if err != nil {
		return nil, err
	}
	// controllo: il file prodotto si rilegge
	if back, err := readAny(out, "uscita."+ext, req.Target); err != nil {
		return nil, fmt.Errorf("errore interno, file non salvato: %v", err)
	} else if back.count() == 0 {
		return nil, errors.New("nessun canale trasferibile nel formato scelto: file non salvato")
	}
	p := req.Path
	if p == "" {
		stem := strings.TrimSuffix(req.Name, filepath.Ext(req.Name))
		if stem == "" {
			stem = "canali"
		}
		suffix := map[string]string{"chirp": "_CHIRP", "xlsx": "", "csv": "", kindXLC: "_K67", "baj-1.0.8": "_BJ8300_v108", "baj-1.0.9": "_BJ8300_v109"}[req.Target]
		if m, ok := genByID[req.Target]; ok {
			suffix = "_" + genFileBase(m)
		}
		suggest := stem + suffix + "." + ext
		if a.store.Settings.LastDir != "" {
			suggest = filepath.Join(a.store.Settings.LastDir, suggest)
		}
		filt := map[string]string{"csv": "CSV (*.csv)|*.csv", "xlsx": "Excel (*.xlsx)|*.xlsx", "xlc": radioFilter[kindXLC],
			"dat": radioFilter[kindDAT], "img": imgFilter}[ext]
		if p, err = a.dialog(true, "Salva il file convertito", filt, suggest, ext); err != nil {
			return nil, err
		}
	}
	if err := writeAtomic(p, out); err != nil {
		return nil, fmt.Errorf("impossibile salvare: %v", err)
	}
	n := 0
	if back, err := readAny(out, "uscita."+ext, req.Target); err == nil {
		n = back.count()
	}
	return map[string]interface{}{"path": p, "count": n, "warnings": rep.Warnings}, nil
}

// radio/import: canali da un file qualsiasi da inserire nella scheda della radio
func (a *App) apiRadioImport(r *http.Request) (interface{}, error) {
	var req struct {
		Kind string `json:"kind"`
		Path string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	p, src, err := a.openAny(req.Path, "Importa canali da un file", req.Kind)
	if err != nil {
		return nil, err
	}
	chs := packed(src.Channels)
	if len(chs) == 0 {
		return nil, errors.New("il file non contiene canali")
	}
	return map[string]interface{}{"path": p, "kind": src.Kind, "desc": src.Desc, "channels": chs, "warnings": src.Warnings}, nil
}
