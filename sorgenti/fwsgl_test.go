//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"testing"
)

// sglFakeFile: file .sgl con le intestazioni indicate
func sglFakeFile(group, model, key string, fw []byte) []byte {
	b := bytes.Repeat([]byte{0x5A}, 0x400)
	copy(b, "SGL!")
	off, k := 0x40, [2]byte{0x3C, 0xA7}
	h := []byte{byte(off), byte(off >> 8), k[0], k[1]}
	for i := range h {
		b[0x0C+i] = h[i] ^ "SGL!"[i]
	}
	h2 := bytes.Repeat([]byte{0xFF}, 0x67)
	binary.LittleEndian.PutUint32(h2[0x06:], uint32(len(fw)))
	copy(h2[0x32:0x42], group)
	copy(h2[0x42:0x4A], model)
	copy(h2[0x4A:0x52], "V1.00.01")
	copy(h2[0x5F:0x67], key)
	for i := range h2 {
		b[off+i] = h2[i] ^ k[i%2]
	}
	return append(b, fw...)
}

// sglBoot: caricatore HID della famiglia GD-77
type sglBoot struct {
	mu      sync.Mutex
	key     string // prefisso della chiave della radio
	group   string
	stage   int
	erased  bool
	fw      []byte
	sums    int
	errs    []string
	normal  bool
	badSum  int // KB (da 1) con la somma rifiutata
	cmds    []string
	partial uint32
	kb      int
}

func (e *sglBoot) serve(p Port) error {
	reply := func(d []byte) error {
		var b [42]byte
		b[0], b[2] = 3, byte(len(d))
		copy(b[4:], d)
		return p.Write(b[:])
	}
	for {
		var rq [42]byte
		if err := emuReadInto(p, rq[:]); err != nil {
			return nil
		}
		if rq[0] != 1 || rq[1] != 0 {
			return fmt.Errorf("rapporto non valido")
		}
		n := int(rq[2]) | int(rq[3])<<8
		d := append([]byte{}, rq[4:4+n]...)
		e.mu.Lock()
		var out []byte
		trim := string(bytes.TrimRight(d, "\xff"))
		switch {
		case e.normal:
			out = []byte{0}
		case string(d) == "DOWNLOAD":
			out, e.stage = []byte("#UPDATE?"), 1
		case e.stage == 1 && string(d) == "A":
			out, e.stage = []byte("A"), 2
		case e.stage == 2:
			if !bytes.HasPrefix(d, []byte(e.key)) || len(d) != 8 {
				out = []byte(e.key)
				e.stage = 99
				break
			}
			out, e.stage = d[:4], 3
		case e.stage >= 3 && e.stage < 7:
			want := []string{"F-PROG", e.group, "", "V1.00.01"}[e.stage-3]
			if want != "" && trim != want {
				e.errs = append(e.errs, fmt.Sprintf("atteso %q, ricevuto %q", want, trim))
			}
			e.cmds = append(e.cmds, trim)
			out = []byte("A")
			e.stage++
		case e.stage == 7 && trim == "F-ERASE":
			e.erased, e.stage, out = true, 8, []byte("A")
		case e.stage == 8 && string(d) == "A":
			e.stage, out = 9, []byte("A")
		case e.stage == 9 && string(d) == "PROGRAM\x0f":
			e.stage, out = 10, []byte("A")
		case e.stage == 10 && bytes.HasPrefix(d, []byte("END\xff")):
			e.kb++
			var sum uint32
			for _, b := range e.fw[(e.kb-1)*sglSumBlock:] {
				sum += uint32(b)
			}
			if sum != binary.LittleEndian.Uint32(d[4:]) || e.kb == e.badSum {
				out = []byte{0x15}
				break
			}
			e.sums++
			out = []byte("A")
		case e.stage == 10:
			a := binary.BigEndian.Uint32(d)
			l := int(binary.BigEndian.Uint16(d[4:]))
			if int(a) != len(e.fw) || l != len(d)-6 || l > 32 {
				e.errs = append(e.errs, fmt.Sprintf("blocco a 0x%X lungo %d", a, l))
			}
			e.fw = append(e.fw, d[6:]...)
			out = []byte("A")
		default:
			e.errs = append(e.errs, fmt.Sprintf("comando inatteso %q nella fase %d", d, e.stage))
			out = []byte{0x15}
		}
		e.mu.Unlock()
		if err := reply(out); err != nil {
			return err
		}
	}
}

func TestSGLFirmware(t *testing.T) {
	fw := fakeData(0x31234, 9)
	file := sglFakeFile("BF-5R", "BF-5R", "DV02S67b", fw)
	in := analyzeSGL("RD-5R_V2.1.6.sgl", "x", file).Info
	for _, c := range in.Checks {
		t.Logf("%s: %s", c.Level, c.Text)
	}
	if !in.Valid || in.Confirm != "RD5R" || in.Version != "V2.1.6" || sniffKind("RD-5R_V2.1.6.sgl", file) != fwSGL {
		t.Fatalf("%+v", in)
	}
	if in := analyzeSGL("x.sgl", "x", sglFakeFile("SG-MD-760", "MD-760", "DV01amnb", fw)).Info; in.Valid {
		t.Fatal("firmware della GD-77 accettato")
	}
	if in := analyzeSGL("x.sgl", "x", sglFakeFile("BF-DMR", "1801", "DV03t!D9", fw)).Info; !in.Valid || in.Confirm != "DM1801" {
		t.Fatalf("DM-1801: %+v", in)
	}
	for _, c := range []struct {
		name    string
		setup   func(*sglBoot)
		ok      bool
		started bool
	}{
		{"normale", nil, true, true},
		{"altra radio", func(e *sglBoot) { e.key = "DV03" }, false, false},
		{"radio accesa normalmente", func(e *sglBoot) { e.normal = true }, false, false},
		{"somma rifiutata", func(e *sglBoot) { e.badSum = 5 }, false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := &sglBoot{key: "DV02", group: "BF-5R"}
			if c.setup != nil {
				c.setup(e)
			}
			old := openPortFunc
			openPortFunc = func(name string, baud int) (Port, error) {
				a, b := newPipe()
				go e.serve(b)
				return a, nil
			}
			defer func() { openPortFunc = old }()
			started := false
			sent, err := sglUpdate(context.Background(), file, noProg, func(s string) { t.Log(s) }, func() { started = true })
			e.mu.Lock()
			defer e.mu.Unlock()
			t.Logf("errore %v, %d byte, avviato %v, %v %q", err, sent, started, e.errs, e.cmds)
			if (err == nil) != c.ok || started != c.started {
				t.Fatal("esito inatteso")
			}
			if c.ok && (len(e.errs) > 0 || !bytes.Equal(e.fw, fw) || e.sums != (len(fw)+1023)/1024 || !e.erased) {
				t.Fatal("firmware ricevuto diverso")
			}
			if !c.started && e.erased {
				t.Fatal("memoria cancellata")
			}
		})
	}
	// file vero, se presente: IU0VYT_SGL_FW=percorso
	if p := os.Getenv("IU0VYT_SGL_FW"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		in := analyzeSGL(p, p, b).Info
		for _, c := range in.Checks {
			t.Logf("%s: %s", c.Level, c.Text)
		}
		if !in.Valid {
			t.Fatal("firmware non riconosciuto")
		}
	}
}
