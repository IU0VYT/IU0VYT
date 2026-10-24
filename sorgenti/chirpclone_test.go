//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Confronto dei protocolli via cavo con CHIRP: la stessa radio simulata viene
// letta e scritta da CHIRP (collegato via TCP) e da questo programma; si
// confrontano tutti i byte inviati alla radio e il contenuto finale della
// memoria. Si esegue solo se CHIRP_DIR indica i sorgenti di CHIRP.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// radio simulata per il modello (nil se la famiglia non ne ha una)
func emuFor(m *genModel, img []byte) genEmu {
	switch m.Fam.(type) {
	case uv5rFam:
		e := newUV5REmu(m, img)
		// radio che non esiste in versione "originale" (firmware < BFB291): la radio
		// simulata dichiara un firmware recente, come le radio vere
		if o := uv5ro(m); o.neverOrig && o.aux && len(e.img) > 0x1848 {
			if i := bytes.Index(e.img[0x1838:0x1848], []byte("BFB")); i >= 0 {
				copy(e.img[0x1838+i+3:], "297")
			}
		}
		return e
	}
	if f, ok := m.Fam.(interface {
		emu(m *genModel, img []byte) genEmu
	}); ok {
		return f.emu(m, img)
	}
	return nil
}

// sessione di CHIRP con la radio simulata: byte ricevuti dalla radio
func chirpSession(t *testing.T, m *genModel, e genEmu, op, file string) []byte {
	var rec *recPort
	var wg sync.WaitGroup
	wg.Add(1)
	addr, stop := serveTCP(func(p Port) error {
		defer wg.Done()
		rec = &recPort{Port: p}
		err := e.serve(rec)
		if err != nil {
			t.Logf("radio simulata (CHIRP): %v", err)
		}
		return err
	})
	defer stop()
	chirpRef(t, "clone", m.RClass, addr, op, file)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("la radio simulata non ha terminato la sessione con CHIRP")
	}
	return rec.got()
}

// sessione di questo programma con la radio simulata
func goSession(t *testing.T, m *genModel, e genEmu, run func() error) []byte {
	a, b := newPipe()
	rec := &recPort{Port: b}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) { return a, nil }
	defer func() { openPortFunc = old }()
	done := make(chan error, 1)
	go func() { done <- e.serve(rec) }()
	if err := run(); err != nil {
		t.Fatalf("questo programma: %v", err)
	}
	_ = a.Close()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("la radio simulata non ha terminato la sessione")
	}
	return rec.got()
}

func TestChirpClone(t *testing.T) {
	chirpDir(t)
	tmp := t.TempDir()
	only := os.Getenv("CHIRP_MODEL")
	for _, m := range genModels {
		if m.NoChirp {
			continue
		}
		m := m
		if only != "" && m.ID != only {
			continue
		}
		t.Run(m.ID, func(t *testing.T) {
			p := chirpImageFor(t, m)
			img := loadImg(t, p)
			if emuFor(m, img) == nil {
				t.Skip("nessuna radio simulata per questa famiglia")
			}
			// ---- lettura
			eC, eG := emuFor(m, img), emuFor(m, img)
			outC := filepath.Join(tmp, m.ID+"_dl.img")
			recC := chirpSession(t, m, eC, "download", outC)
			var got []byte
			recG := goSession(t, m, eG, func() error {
				var err error
				got, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
				return err
			})
			if !bytes.Equal(recC, recG) {
				t.Errorf("lettura: byte inviati alla radio diversi da CHIRP (%d / %d byte): %s", len(recC), len(recG), diffRanges(recC, recG))
			}
			dlC, err := os.ReadFile(outC)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dlC, got) {
				t.Errorf("lettura: immagine diversa da quella letta da CHIRP (%d / %d byte): %s", len(dlC), len(got), diffRanges(dlC, got))
			}
			t.Logf("lettura: %d byte inviati alla radio, identici a CHIRP; immagine di %d byte identica", len(recG), len(got))

			// ---- scrittura di un'immagine modificata
			mod := append([]byte{}, got...)
			for _, e := range chirpEdits(m, mod) {
				if err := m.Fam.Set(m, mod, e.N, e.Chan); err != nil {
					t.Fatal(err)
				}
			}
			inF := filepath.Join(tmp, m.ID+"_ul.img")
			_ = os.WriteFile(inF, mod, 0o644)
			uC, uG := emuFor(m, img), emuFor(m, img)
			recC = chirpSession(t, m, uC, "upload", inF)
			recG = goSession(t, m, uG, func() error {
				_, err := m.Fam.Upload(context.Background(), m, "EMU", mod, func(int, int, string) {})
				return err
			})
			if !bytes.Equal(recC, recG) {
				t.Errorf("scrittura: byte inviati alla radio diversi da CHIRP (%d / %d byte): %s", len(recC), len(recG), diffRanges(recC, recG))
			}
			if !bytes.Equal(uC.image(), uG.image()) {
				t.Errorf("scrittura: memoria finale della radio diversa: %s", diffRanges(uC.image(), uG.image()))
			}
			t.Logf("scrittura: %d byte inviati alla radio, identici a CHIRP; memoria finale identica", len(recG))
		})
	}
}
