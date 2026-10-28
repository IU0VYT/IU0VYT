//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Prova del Port USB DFU (dfuport.go, usato su Windows con i driver STTub30 e
// WinUSB): ogni richiesta di controllo passa da xfer, qui collegata alla radio
// simulata. Lettura e scrittura della DM-1701 come con il collegamento diretto.

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"
)

func TestDFUPortFraming(t *testing.T) {
	m := genByID["bf-dm1701"]
	base := bytes.Repeat([]byte{0xFF}, tytMemSize)
	c := &Chan{Rx: 438500000, Tx: 430900000, Power: 0, Scan: true, Name: "IU0VYT", TxTone: "88.5", RxTone: "88.5"}
	if err := m.Fam.Set(m, base, 0, c); err != nil {
		t.Fatal(err)
	}
	emu := genEmuFor(m, base)
	nreq := 0
	open := func() Port {
		a, b := newPipe()
		go emu.serve(b)
		// richiesta di controllo → stessa richiesta alla radio simulata (come farebbe il driver USB)
		return &dfuPort{close: a.Close, xfer: func(in bool, rt, req byte, value, index uint16, out []byte, n int) (int32, []byte) {
			nreq++
			if !in {
				n = len(out)
			}
			h := []byte{rt, req, byte(value), byte(value >> 8), byte(index), byte(index >> 8), byte(n), byte(n >> 8)}
			if err := a.Write(append(h, out...)); err != nil {
				return -1, nil
			}
			var r [4]byte
			if err := a.ReadFull(context.Background(), r[:], 5*time.Second); err != nil {
				return -1, nil
			}
			res := int32(binary.LittleEndian.Uint32(r[:]))
			if !in || res <= 0 {
				return res, nil
			}
			d := make([]byte, res)
			if err := a.ReadFull(context.Background(), d, 5*time.Second); err != nil {
				return -1, nil
			}
			return res, d
		}}
	}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) { return open(), nil }
	defer func() { openPortFunc = old }()
	mem, info, err := m.Fam.Download(context.Background(), m, "USB", noProg)
	if err != nil {
		t.Fatalf("lettura: %v", err)
	}
	if !bytes.Equal(mem, base) {
		t.Fatalf("memoria letta diversa")
	}
	if g := m.Fam.Get(m, mem, 0); g == nil || g.Name != "IU0VYT" || g.Rx != 438500000 {
		t.Fatalf("canale: %+v", g)
	}
	c2 := *c
	c2.Name = "PROVA"
	if err := m.Fam.Set(m, mem, 1, &c2); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Fam.Upload(context.Background(), m, "USB", mem, noProg); err != nil {
		t.Fatalf("scrittura: %v", err)
	}
	if g := m.Fam.Get(m, emu.image(), 1); g == nil || g.Name != "PROVA" {
		t.Fatalf("radio dopo la scrittura: %+v", g)
	}
	t.Logf("%d richieste di controllo; %s", nreq, info)
}
