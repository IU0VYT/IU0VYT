//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Modelli ricavati dai programmi dei costruttori (non presenti in CHIRP): lettura,
// modifica e scrittura via cavo con la radio simulata; la memoria finale deve
// contenere esattamente le modifiche fatte.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func noChirpBase(t *testing.T, m *genModel) []byte {
	if _, uv17 := m.Opt.(*uv17Opt); uv17 && os.Getenv("CHIRP_DIR") != "" {
		d := os.Getenv("CHIRP_DIR")
		// immagine di prova di CHIRP della classe con la stessa mappa
		for _, n := range []string{"Baofeng_UV-17ProGPS.img", "Baofeng_UV-17Pro.img"} {
			if b, err := os.ReadFile(filepath.Join(d, "tests/images", n)); err == nil {
				mem, _ := splitImg(b)
				if lf, ok := m.Fam.(interface{ layout(m *genModel) *bwEl }); ok && len(mem) >= lf.layout(m).ByteLen() {
					return mem
				}
			}
		}
	}
	size := 0
	switch o := m.Opt.(type) {
	case *uv17Opt:
		size = o.total
	case *rt900Opt:
		size = o.memsize
	case *h777Opt:
		size = o.memsize
	case *rt22Opt:
		size = o.memsize
	}
	mem := bytes.Repeat([]byte{0xFF}, size)
	if err := m.Fam.Set(m, mem, 0, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"}); err != nil {
		t.Fatal(err)
	}
	return mem
}

func TestNoChirpModels(t *testing.T) {
	n := 0
	for _, m := range genModels {
		if !m.NoChirp {
			continue
		}
		n++
		m := m
		t.Run(m.ID, func(t *testing.T) {
			base := noChirpBase(t, m)
			emu := genEmuFor(m, base)
			old := openPortFunc
			openPortFunc = func(name string, baud int) (Port, error) {
				a, b := newPipe()
				a.baud = baud
				go emu.serve(b)
				return a, nil
			}
			defer func() { openPortFunc = old }()
			got, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, emu.image()) {
				t.Fatal("lettura diversa dalla memoria della radio")
			}
			chs := genChans(m, got)
			for _, e := range chirpEdits(m, got) {
				chs[e.N] = e.Chan
			}
			rep := &Report{}
			want, err := genWrite(m, got, chs, rep)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := m.Fam.Upload(context.Background(), m, "COM7", want, noProg); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(emu.image(), want) {
				t.Fatal("memoria della radio diversa da quella scritta")
			}
			for i, c := range chs {
				g := m.Fam.Get(m, want, i)
				if c == nil && g != nil || c != nil && (g == nil || g.Rx != c.Rx || g.Name != c.Name || g.TxTone != c.TxTone || g.RxTone != c.RxTone) {
					t.Fatalf("canale %d: scritto %+v, riletto %+v", i, c, g)
				}
			}
		})
	}
	if n == 0 {
		t.Skip("nessun modello")
	}
}

// la radio che si identifica con un altro modello viene rifiutata
func TestNoChirpWrongModel(t *testing.T) {
	m := genByID["bf-uv25plus"]
	other := genByID["bf-ar152pro"]
	emu := genEmuFor(other, noChirpBase(t, other))
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go emu.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	_, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("AR152")) {
		t.Fatalf("errore atteso, ottenuto %v", err)
	}
}
