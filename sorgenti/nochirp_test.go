//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
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
	"encoding/binary"
	"fmt"
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
	case *jc8810Opt: // 4.18: Radtel RT-330 e RT-400
		size = o.memsize
	case *h777Opt:
		size = o.memsize
	case *rt22Opt:
		size = o.memsize
	case *bfdOpt:
		size = bfdMemsize
	case *t17Opt:
		size = t17Memsize
	case *dm5kOpt:
		size = dm5kSize()
	case *rt4dOpt: // 4.20: Radtel RT-4D, RT-6D, RT-1000 Pro
		size = rt4dSize
	case *rt5dOpt: // 4.21: Radtel RT-5D
		size = rt5dSize()
	}
	mem := bytes.Repeat([]byte{0xFF}, size)
	if _, ok := m.Opt.(*drOpt); ok {
		mem = drSynthImage()
	}
	if _, ok := m.Fam.(ijvFam); ok { // 4.13: firmware IJV
		mem = ijvSynthImage(m)
	}
	if _, ok := m.Opt.(*rdtyOpt); ok {
		mem = rdtySynthImage(m)
	}
	if _, ok := m.Opt.(*tytOpt); ok {
		mem = tytSynthImage()
	}
	if _, ok := m.Fam.(ogdFam); ok { // 4.16: OpenGD77 (memoria vuota: nessun canale valido)
		mem = bytes.Repeat([]byte{0xFF}, ogdSize)
		for b := 0; b < 8; b++ {
			copy(mem[ogdBank(b):], make([]byte, 16))
		}
		copy(mem[rdtyZoneTab:], make([]byte, 32))
	}
	if _, ok := m.Fam.(kwlFam); ok { // 4.17: Kenwood programmate un canale per volta
		mem = kwlBlank(m)
	}
	if _, ok := m.Fam.(fagciFam); ok { // 4.18: firmware R3b0rn di fagci (EEPROM originale di 8 KiB)
		mem = fagciBlank(2)
	}
	if f, ok := m.Fam.(iradioFam); ok { // 4.19: Radtel RT-817A / RT-871A
		mem = bytes.Repeat([]byte{0xFF}, f.proto().size)
	}
	switch m.Fam.(type) { // 4.15: firmware nicFW
	case nicfwFam:
		mem = nicfwTestImg(true)
	case nicfw880Fam:
		mem = bytes.Repeat([]byte{0xFF}, nicfw880Size)
		binary.LittleEndian.PutUint16(mem[nicfw880SetAt:], 0xABC6)
	}
	first := &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"}
	if m.NameMax == 0 {
		first.Name = ""
	}
	if err := m.Fam.Set(m, mem, 0, first); err != nil {
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
				if _, tyt := m.Opt.(*tytOpt); tyt && c != nil && c.Name == "" && g != nil && g.Name == fmt.Sprintf("CH %d", i+1) {
					g.Name = "" // la DM-1701 vuole un nome: si usa «CH n»
				}
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
