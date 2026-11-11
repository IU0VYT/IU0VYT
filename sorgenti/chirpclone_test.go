//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
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

// 4.6: classi di CHIRP con cui confrontare il collegamento quando quella del
// modello non lo permette. UV68 (TID TD-UV68, la TD-H8 con un altro nome): in CHIRP
// la scrittura si ferma con un errore interno (il modello non è tra quelli di cui
// _get_radio_firmware_version legge la versione), quindi si confronta con la TD-H8
var chirpCloneClass = map[string]string{"td-uv68": "TDH8"}

// 4.8: driver di CHIRP la cui lettura/scrittura via cavo non funziona con Python 3
// (kyd.py, kyd_IP620.py: struct.pack(">cHb", 'R', …) con una stringa, confronti
// tra byte e stringhe; puxing.py, PX-2R: la scrittura unisce stringhe e byte;
// boblov_x3plus.py: comandi come stringhe; retevis_h777v4.py, RT21H: la lettura
// non entra in programmazione perché la classe risulta già «riconosciuta»;
// fd268.py: confronti tra byte e stringhe). Qui si prova solo questo programma con la radio
// simulata: immagine letta uguale alla memoria della radio, memoria finale
// uguale all'immagine scritta
var chirpCloneBroken = map[string]bool{"NC630aRadio": true, "IP620Radio": true, "Puxing2RRadio": true, "BoblovX3Plus": true, "RT21HRadio": true, "FD268ARadio": true, "FD268BRadio": true, "FD288BRadio": true,
	// 4.10: stessa famiglia, ora provata con l'immagine di un modello vicino
	"FD288ARadio": true, "FD150ARadio": true, "FD160ARadio": true, "FD450ARadio": true, "FD460ARadio": true, "FD460UHRadio": true,
	// 4.9: tdxone_tdq8a.py confronta bytes con str (Python 2): con CHIRP la copia non parte mai
	"TDXoneTDQ8A": true,
	// 4.9: uv5r.py con la BJ-UV55 usa has_dropped_byte senza averlo calcolato:
	// la lettura di CHIRP si interrompe sempre
	"BaojieBJUV55Radio": true}

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
	rc := m.RClass
	if c, ok := chirpCloneClass[m.ID]; ok {
		rc = c
	}
	chirpRef(t, "clone", rc, addr, op, file)
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

// 4.9: metadati che il driver di CHIRP vuole nell'immagine da scrivere
var chirpUploadMeta = map[string]map[string]interface{}{
	"tyt-uvl15w": {"uvl15w_version": "3.09.21"},
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
			if chirpCloneBroken[m.RClass] {
				e := emuFor(m, img)
				var got []byte
				goSession(t, m, e, func() error {
					var err error
					got, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
					return err
				})
				if want := e.image(); !bytes.Equal(got, want) {
					t.Fatalf("lettura: immagine diversa dalla memoria della radio: %s", diffRanges(want, got))
				}
				mod := append([]byte{}, got...)
				for _, ed := range chirpEdits(m, mod) {
					if err := m.Fam.Set(m, mod, ed.N, ed.Chan); err != nil {
						t.Fatal(err)
					}
				}
				u := emuFor(m, img)
				goSession(t, m, u, func() error {
					_, err := m.Fam.Upload(context.Background(), m, "EMU", mod, func(int, int, string) {})
					return err
				})
				if fin := u.image(); !bytes.Equal(fin[:len(mod)], mod) {
					t.Errorf("scrittura: memoria finale diversa dall'immagine scritta: %s", diffRanges(mod, fin))
				}
				t.Logf("lettura e scrittura verificate solo con la radio simulata (driver di CHIRP non utilizzabile via cavo con Python 3)")
				return
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
			if meta, ok := chirpUploadMeta[m.ID]; ok {
				// 4.9: CHIRP scrive solo un'immagine con i metadati della lettura
				_ = os.WriteFile(inF, makeImg(m, mod, meta), 0o644)
			} else {
				_ = os.WriteFile(inF, mod, 0o644)
			}
			uC, uG := emuFor(m, img), emuFor(m, img)
			recC = chirpSession(t, m, uC, "upload", inF)
			recG = goSession(t, m, uG, func() error {
				_, err := m.Fam.Upload(context.Background(), m, "EMU", mod, func(int, int, string) {})
				return err
			})
			// 4.8, Radtel RT-900: i canali 513-999 si rileggono e si scrivono solo se
			// cambiano (firmware da 512 canali, vedi rt900hi_test.go): la sequenza è
			// diversa da CHIRP di proposito, la memoria finale deve essere identica
			hiOwn := false
			if _, ok := m.Fam.(rt900Fam); ok && rt900o(m).hiFirst != 0 {
				hiOwn = true
			}
			// 4.11: prima di scrivere, l'area dei nomi delle zone si rilegge dalla
			// radio (si scrive solo se cambia): tolti quei comandi di lettura, la
			// sequenza deve essere quella di CHIRP
			if zs := m.ZoneNames; zs != nil && zs.Area > 0 && len(mod) >= zs.Addr+zs.Area {
				bs := rt900o(m).blk()
				for a := zs.Addr; a < zs.Addr+zs.Area; a += bs {
					cmd := []byte{'R', byte(a >> 8), byte(a), byte(bs)}
					if bytes.Count(recG, cmd) != 1 {
						t.Errorf("scrittura: lettura dell'area dei nomi delle zone (0x%04X) non trovata", a)
					}
					recG = bytes.Replace(recG, cmd, nil, 1)
				}
			}
			// 4.16: famiglie che di proposito scrivono meno di CHIRP (es. UV-KX: mai la
			// taratura): la memoria finale deve comunque essere identica
			if f, ok := m.Fam.(interface{ chirpOwnWrite() bool }); ok && f.chirpOwnWrite() {
				hiOwn = true
			}
			if !hiOwn && !bytes.Equal(recC, recG) {
				t.Errorf("scrittura: byte inviati alla radio diversi da CHIRP (%d / %d byte): %s", len(recC), len(recG), diffRanges(recC, recG))
			}
			if !bytes.Equal(uC.image(), uG.image()) {
				t.Errorf("scrittura: memoria finale della radio diversa: %s", diffRanges(uC.image(), uG.image()))
			}
			t.Logf("scrittura: %d byte inviati alla radio, identici a CHIRP; memoria finale identica", len(recG))
		})
	}
}
