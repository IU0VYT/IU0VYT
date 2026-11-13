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
	"strings"
	"testing"
	"time"
)

// 4.32: Baofeng T22 (PMR446). Il programma del costruttore manda «\x02PROGRGS»,
// CHIRP (W31E e le altre della famiglia) solo «PROGRGS». Due radio finte: una
// accetta il saluto solo con 0x02 davanti, l'altra solo senza.
func t22Fake(p Port, needSTX bool, img []byte) {
	var last []byte
	for {
		c, err := emuByte(p)
		if err != nil {
			return
		}
		last = append(last, c)
		if len(last) > 8 {
			last = last[1:]
		}
		if !bytes.HasSuffix(last, []byte("PROGRGS")) {
			continue
		}
		stx := len(last) == 8 && last[0] == 0x02
		if stx == needSTX {
			break
		}
		last = nil
	}
	p.Write([]byte{0x06})
	for _, st := range []struct {
		want byte
		resp []byte
	}{{0x02, []byte("P3T22\xF8\xFF\xFF")}, {0x06, []byte{0x06}}, {0x07, []byte{0x4E}}} {
		if c, err := emuByte(p); err != nil || c != st.want {
			return
		}
		p.Write(st.resp)
	}
	for {
		c, err := emuByte(p)
		if err != nil || c != 'R' {
			return
		}
		h, err := emuRead(p, 3)
		if err != nil {
			return
		}
		a, n := int(h[0])<<8|int(h[1]), int(h[2])
		p.Write(append([]byte{'W', h[0], h[1], h[2]}, img[a:a+n]...))
		if c, err := emuByte(p); err != nil || c != 0x06 {
			return
		}
		p.Write([]byte{0x06})
	}
}

func TestT22Saluto432(t *testing.T) {
	m := genByID["bf-t22"]
	if m == nil || !rt22o(m).stx || rt22o(m).identJoin() != "bf-w31e" {
		t.Fatal("Baofeng T22: opzioni")
	}
	for id, o := range map[string]*genModel{"bf-w31e": genByID["bf-w31e"], "bf-t20": genByID["bf-t20"], "rt-rt22": genByID["rt-rt22"]} {
		if rt22o(o).stx || rt22o(o).identJoin() != "" {
			t.Fatalf("%s: non deve cambiare", id)
		}
	}
	img := make([]byte, 0x200)
	for i := range img {
		img[i] = byte(i * 3)
	}
	for _, needSTX := range []bool{true, false} {
		a, b := newPipe()
		old := openPortFunc
		openPortFunc = func(string, int) (Port, error) { return a, nil }
		go t22Fake(b, needSTX, img)
		start := time.Now()
		got, _, err := m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
		openPortFunc = old
		_ = a.Close()
		if err != nil || !bytes.Equal(got, img) {
			t.Fatalf("radio che vuole 0x02 = %v: %v", needSTX, err)
		}
		// con 0x02 risponde al primo tentativo, senza al secondo (dopo un secondo di attesa)
		if d := time.Since(start); needSTX && d > 900*time.Millisecond || !needSTX && d < 900*time.Millisecond {
			t.Fatalf("radio che vuole 0x02 = %v: lettura in %v", needSTX, d)
		}
	}
	// «Identifica»: saluto proprio, e la T22 resta fra le candidate del gruppo della W31E
	keys := identTranscripts()
	if !strings.HasPrefix(keys["bf-t22"], "<9600>\x02PROGRGS") || !strings.HasPrefix(keys["bf-w31e"], "<9600>PROGRGS") {
		t.Fatalf("saluti: %q e %q", keys["bf-t22"], keys["bf-w31e"])
	}
	app := newApp(mustStore(t))
	identPortMu.Lock()
	plan := app.identPlan(&identReq{Port: "COM7"})
	identPortMu.Unlock()
	own, joined := -1, -1
	for i, p := range plan {
		hasT22, hasW31E := false, false
		for _, k := range p.kinds {
			hasT22 = hasT22 || k.ID == "bf-t22"
			hasW31E = hasW31E || k.ID == "bf-w31e"
		}
		switch {
		case hasT22 && hasW31E:
			joined = i
			if p.kinds[len(p.kinds)-1].ID != "bf-t22" {
				t.Fatal("la T22 deve essere in coda al gruppo della W31E")
			}
		case hasT22:
			own = i
			if len(p.kinds) != 1 {
				t.Fatalf("gruppo della T22 con %d modelli", len(p.kinds))
			}
		}
	}
	if own < 0 || joined < 0 || joined > own {
		t.Fatalf("gruppi della T22: proprio %d, della W31E %d", own, joined)
	}
}
