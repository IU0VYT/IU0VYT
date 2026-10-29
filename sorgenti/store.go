// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Archivio ponti e impostazioni, salvati in formato JSON nella cartella
// dell'utente (%APPDATA%\GestioneCanaliRadio su Windows; la cartella mantiene il
// nome delle versioni precedenti, così archivio e impostazioni restano gli stessi).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const appDirName = "GestioneCanaliRadio"

type Settings struct {
	URL       string   `json:"url"`
	URLs      []string `json:"urls"`
	Modello   string   `json:"modello"`
	LastDir   string   `json:"lastDir"`
	K67Path   string   `json:"k67Path"`
	BajPath   string   `json:"bajPath"`
	ShowEmpty bool     `json:"showEmpty"`
	// collegamento via cavo
	K67Port string `json:"k67Port"`
	K67Baud int    `json:"k67Baud"`
	K67Pwd  string `json:"k67Pwd"`
	// modello dell'ultima radio K67 letta: si è identificata come «BF-K6»
	K67K6       bool   `json:"k67K6"`
	K67Known    bool   `json:"k67Known"`
	K67Band     byte   `json:"k67Band"` // banda di frequenza dell'ultima radio K67 letta
	BajPort     string `json:"bajPort"`
	BajVer      string `json:"bajVer"`  // formato dei nuovi file .dat: "1.0.8" (predefinito) o "1.0.9"
	GenPort     string `json:"genPort"` // radio gestite con i driver di CHIRP
	CableBackup bool   `json:"cableBackup"`
	CableVerify bool   `json:"cableVerify"`
	// aggiornamento del firmware della Bajeton
	FwURL      string `json:"fwURL"`
	FwBaud     int    `json:"fwBaud"`
	FwNoBackup bool   `json:"fwNoBackup"`
	// non più usato dalla 3.1.1 (l'indirizzo degli aggiornamenti è fisso): resta per
	// leggere le impostazioni salvate dalle versioni precedenti
	UpdateURL string `json:"updateURL"`
}

const defaultURL = "https://www.ik2ane.it/pontixls.xls"

type archiveFile struct {
	Versione int      `json:"versione"`
	Ponti    []*Ponte `json:"ponti"`
}

type Store struct {
	mu       sync.Mutex
	dir      string
	Ponti    []*Ponte
	Settings Settings
}

func appDataDir() string {
	if d := os.Getenv("GESTIONE_CANALI_DIR"); d != "" {
		return d
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	return filepath.Join(base, appDirName)
}

func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, Settings: Settings{URL: defaultURL, Modello: "{LOC}", K67Baud: 19200, CableBackup: true, CableVerify: true}}
	if b, err := os.ReadFile(s.archivePath()); err == nil {
		var a archiveFile
		if err := json.Unmarshal(b, &a); err != nil {
			// archivio rovinato: lo si mette da parte e si riparte dalla copia di sicurezza
			_ = os.Rename(s.archivePath(), s.archivePath()+".danneggiato")
			if b2, err2 := os.ReadFile(s.archivePath() + ".bak"); err2 == nil && json.Unmarshal(b2, &a) == nil {
				s.Ponti = a.Ponti
			}
		} else {
			s.Ponti = a.Ponti
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, "impostazioni.json")); err == nil {
		_ = json.Unmarshal(b, &s.Settings)
	}
	if s.Settings.URL == "" {
		s.Settings.URL = defaultURL
	}
	if s.Settings.Modello == "" {
		s.Settings.Modello = "{LOC}"
	}
	for _, p := range s.Ponti {
		fixPonte(p)
	}
	return s, nil
}

func fixPonte(p *Ponte) {
	if p.ID == "" {
		p.ID = newID()
	}
	if p.Tipi == nil {
		p.Tipi = []string{}
	}
	if p.Avvisi == nil {
		p.Avvisi = []string{}
	}
	if p.Tx == 0 {
		p.Tx = p.Rx
	}
	p.Riga = 0
}

func (s *Store) archivePath() string { return filepath.Join(s.dir, "archivio_ponti.json") }

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (s *Store) SavePonti(pp []*Ponte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range pp {
		if p == nil {
			return fmt.Errorf("archivio non valido")
		}
		fixPonte(p)
	}
	b, err := json.MarshalIndent(archiveFile{Versione: 1, Ponti: pp}, "", " ")
	if err != nil {
		return err
	}
	if _, err := os.Stat(s.archivePath()); err == nil {
		if old, err := os.ReadFile(s.archivePath()); err == nil {
			_ = os.WriteFile(s.archivePath()+".bak", old, 0o644)
		}
	}
	if err := writeAtomic(s.archivePath(), b); err != nil {
		return err
	}
	s.Ponti = pp
	return nil
}

// Get: copia delle impostazioni (lette anche dalle operazioni con la radio in
// corso, in parallelo con le richieste dell'interfaccia)
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Settings
}

func (s *Store) SaveSettings(st Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Settings = st
	b, _ := json.MarshalIndent(st, "", " ")
	return writeAtomic(filepath.Join(s.dir, "impostazioni.json"), b)
}
