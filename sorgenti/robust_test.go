//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.23: un file .img rovinato (o di un'altra radio con la stessa dimensione) non
// deve mai fermare il programma. Per ogni modello si parte da una memoria valida,
// la si riempie di byte a caso e si leggono, si modificano e si cancellano tutti
// i canali: sono ammessi gli errori, non gli arresti (panic).

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func robustBase(t *testing.T, m *genModel) []byte {
	if m.NoChirp {
		return noChirpBase(t, m)
	}
	if os.Getenv("CHIRP_DIR") == "" {
		return nil
	}
	p := chirpImageFor(t, m)
	if p == "" {
		return nil
	}
	return loadImg(t, p)
}

// robustWhere: prima riga del programma (non della libreria) nella pila dell'arresto
func robustWhere() string {
	pcs := make([]uintptr, 32)
	fr := runtime.CallersFrames(pcs[:runtime.Callers(3, pcs)])
	for {
		f, more := fr.Next()
		if strings.Contains(f.File, "/sorgenti/") && !strings.HasSuffix(f.File, "robust_test.go") {
			return fmt.Sprintf("        in %s:%d (%s)", filepath.Base(f.File), f.Line, f.Function)
		}
		if !more {
			return ""
		}
	}
}

func TestRobustGarbage(t *testing.T) {
	n := 0
	for _, m := range genModels {
		m := m
		t.Run(m.ID, func(t *testing.T) {
			base := robustBase(t, m)
			if base == nil {
				t.Skip("nessuna immagine di partenza")
			}
			n++
			bad := 0
			try := func(what string, i int, fn func()) {
				defer func() {
					if r := recover(); r != nil {
						if bad++; bad <= 3 {
							t.Errorf("%s, canale %d: arresto: %v\n%s", what, i, r, robustWhere())
						}
					}
				}()
				fn()
			}
			for seed := int64(1); seed <= 3; seed++ {
				rnd := rand.New(rand.NewSource(seed))
				img := make([]byte, len(base))
				switch seed {
				case 1: // tutto a caso
					rnd.Read(img)
				case 2: // tutto a zero
				case 3: // memoria valida con un byte su otto a caso
					copy(img, base)
					for k := 0; k < len(img)/8; k++ {
						img[rnd.Intn(len(img))] = byte(rnd.Intn(256))
					}
				}
				for i := 0; i < m.Slots && bad <= 3; i++ {
					try(fmt.Sprintf("lettura (prova %d)", seed), i, func() { _ = genGet(m, img, i); _ = m.Fam.Rec(m, img, i) })
					if i%7 == 0 {
						c := &Chan{Rx: 145500000, Tx: 145500000, Name: "PROVA", Scan: true}
						if !m.inBands(c.Rx) && len(m.Bands) > 0 {
							c.Rx, c.Tx = m.Bands[0].Lo+500000, m.Bands[0].Lo+500000
						}
						try(fmt.Sprintf("scrittura (prova %d)", seed), i, func() { _ = m.Fam.Set(m, img, i, c) })
						try(fmt.Sprintf("rilettura (prova %d)", seed), i, func() { _ = genGet(m, img, i) })
					}
					if i%11 == 0 {
						try(fmt.Sprintf("cancellazione (prova %d)", seed), i, func() { _ = m.Fam.Set(m, img, i, nil) })
					}
				}
			}
		})
	}
	t.Logf("%d modelli provati", n)
}

// 4.23: scrittura di un'immagine troncata (file rovinato): mai un arresto. Si
// provano l'immagine a metà e quella senza l'ultimo byte.
func TestRobustShortUpload(t *testing.T) {
	old := openPortFunc
	defer func() { openPortFunc = old }()
	for _, m := range genModels {
		m := m
		t.Run(m.ID, func(t *testing.T) {
			base := robustBase(t, m)
			emu := genEmuFor(m, base)
			if base == nil || emu == nil {
				t.Skip("nessuna immagine di partenza o nessuna radio simulata")
			}
			openPortFunc = func(name string, baud int) (Port, error) {
				a, b := newPipe()
				a.baud = baud
				go emu.serve(b)
				return a, nil
			}
			for _, n := range []int{len(base) / 2, len(base) - 1} {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("immagine di %d byte su %d: arresto: %v\n%s", n, len(base), r, robustWhere())
						}
					}()
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					// molte radio scrivono solo una parte dell'immagine: se i byte che
					// mancano non servono la scrittura può anche riuscire
					_, _ = m.Fam.Upload(ctx, m, "COM7", base[:n], noProg)
				}()
			}
		})
	}
}

// 4.23: nelle conversioni un canale che la radio non può memorizzare va scartato
// con un avviso (genAdapt), non deve fermare tutto con l'errore della scrittura.
// Per i modelli ricavati dai programmi dei costruttori e dai firmware alternativi:
// se genAdapt tiene il canale, la scrittura nella memoria deve riuscire.
func TestAdaptThenSet(t *testing.T) {
	samples := []*Chan{
		{Rx: 27125000, Tx: 27125000, Name: "CB"},
		{Rx: 51510000, Tx: 51510000, Name: "6M"},
		{Rx: 145500000, Tx: 145500000, Name: "V"},
		{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0"},
		{Rx: 145501000, Tx: 145501000, Name: "PASSO"},
		{Rx: 220000000, Tx: 220000000, Name: "220"},
		{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D023N", Name: "U"},
		{Rx: 446006250, Tx: 446006250, Narrow: true, Name: "PMR"},
		{Rx: 162550000, Tx: 162550000, TxOff: true, Name: "RX"},
		{Rx: 868000000, Tx: 868000000, Name: "868"},
		{Rx: 1297500000, Tx: 1297500000, Name: "23CM"},
		{Rx: 2400000000, Tx: 2400000000, Name: "13CM"},
		{Rx: 145500005, Tx: 145500005, Name: "5HZ"},
	}
	n, kept := 0, 0
	for _, m := range genModels {
		if !m.NoChirp {
			continue
		}
		m := m
		t.Run(m.ID, func(t *testing.T) {
			base := noChirpBase(t, m)
			n++
			for k, c := range samples {
				rep := &Report{}
				a := genAdapt(c, m, c.Name, rep)
				if a == nil {
					continue
				}
				kept++
				img := append([]byte{}, base...)
				// posizione libera lontana dalla prima (che noChirpBase usa già)
				slot := 1 + k
				if slot >= m.Slots {
					slot = m.Slots - 1
				}
				if lk, ok := m.Fam.(interface {
					lockedSlotImg(m *genModel, img []byte, i int) bool
				}); ok && lk.lockedSlotImg(m, img, slot) {
					continue
				}
				if err := m.Fam.Set(m, img, slot, a); err != nil {
					t.Errorf("canale «%s» (%s MHz) tenuto da genAdapt ma rifiutato dalla scrittura: %v", c.Name, fmtMHz(c.Rx), err)
				}
			}
		})
	}
	t.Logf("%d modelli, %d canali tenuti", n, kept)
}
