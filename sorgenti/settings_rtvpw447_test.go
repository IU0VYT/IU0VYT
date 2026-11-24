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
)

// rtvPwSet: la voce della password di programmazione nella tabella del modello
func rtvPwSet(t *testing.T, id string, img []byte) *setDef {
	t.Helper()
	tb := settingsFor(id, img)
	if tb == nil || len(tb.Settings) == 0 {
		t.Fatalf("%s: tabella delle impostazioni assente", id)
	}
	d := &tb.Settings[len(tb.Settings)-1]
	n := 0
	for _, x := range tb.Settings {
		if x.Key == "progpw" {
			n++
		}
	}
	if d.Key != "progpw" || n != 1 {
		t.Fatalf("%s: la password di programmazione deve esserci una volta sola, per ultima", id)
	}
	return d
}

// 4.47, decisione di Daniele del 23/11/2026: la password di programmazione dei
// modelli Retevis «Analog1» (e della H777 nuova versione) si vede in chiaro fra le
// impostazioni e si può cambiare o togliere, con i vincoli dei programmi del
// costruttore: posizione, lunghezza, tabella dei caratteri, caratteri ammessi.
func TestSettingsRetevisPassword447(t *testing.T) {
	ids := []string{"rt-h777v4", "rt-b3b", "rt-b63b", "rt-rt22p", "rt-rt622p", "rt-rb666", "rt-rb666p", "rt-rb45", "rt-rb645", "rt-rb47",
		"rt-b3h", "rt-b63h", "rt-rb647", "rt-rt68h", "rt-rt668h", "rt-nr10", "rt-nr610", "rt-rb66s", "rt-rb666s", "rt-rb65",
		"rt-rb66", "rt-rb66p", "rt-rt68b", "rt-rt668b", "rt-rb669"}
	if len(rtvPwDefs) != len(ids) || len(rtvSetDefs)+1 != len(ids) {
		t.Fatalf("%d modelli con la password, attesi %d", len(rtvPwDefs), len(ids))
	}
	const all = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// lunghezza, caratteri e avviso di ogni modello, scritti a mano dalla scheda
	type lim struct {
		n              int
		digits, hidden bool
	}
	lims := map[string]lim{"rt-rb45": {3, false, true}, "rt-rb645": {3, false, true}, "rt-rb47": {3, false, false}, "rt-rb66p": {4, false, false},
		"rt-rb66": {3, true, true}, "rt-rb65": {6, true, false}, "rt-rb666": {6, false, true}, "rt-b3h": {6, false, true}, "rt-b63h": {6, false, true}}
	for _, id := range ids {
		m := genByID[id]
		if m == nil || !settingsAvailable(id) {
			t.Fatalf("%s: modello assente o senza impostazioni", id)
		}
		size := rtvMemSize(m)
		base := make([]byte, size)
		for i := range base {
			base[i] = byte(i*29 + 7)
		}
		d := rtvPwSet(t, id, base)
		p := rtvPwDefs[id]
		l, ok := lims[id]
		if !ok {
			l = lim{6, false, false}
		}
		wantSet := all
		if l.digits {
			wantSet = all[:10]
		}
		if d.Type != "str" || d.MaxLen != l.n || p.n != l.n || d.Charset != wantSet || d.Group != "Password" || d.ItLabel != "Password di programmazione" ||
			d.RO || d.Enc == nil || d.Enc.Kind != "str" || d.Enc.Fmt != "skip" || len(d.Enc.Offs) != l.n || d.Enc.Pad == nil || *d.Enc.Pad != 0xFF ||
			len(d.Enc.Map) != 62 || p.digits != l.digits || p.hidden != l.hidden {
			t.Fatalf("%s: voce della password con altri limiti: %+v", id, d)
		}
		// la tabella: 62 caratteri, 62 byte diversi, nessuno uguale al riempimento
		seen := map[int]bool{}
		for _, c := range all {
			b, ok := d.Enc.Map[string(c)]
			if !ok || seen[b] || b < 0 || b >= 0xFF {
				t.Fatalf("%s: carattere %q senza byte, o byte %d ripetuto", id, c, b)
			}
			seen[b] = true
		}
		// avviso: è una voce delicata; nei modelli il cui programma nasconde le caselle l'avviso lo dice
		info := setInfoFor(d)
		if !info.Critical || info.Warn == "" || strings.Contains(info.Warn, "non mostra la casella") != l.hidden ||
			strings.Contains(info.Desc, "solo cifre") != l.digits || !strings.Contains(info.Desc, "come CHIRP") {
			t.Fatalf("%s: avviso o descrizione della password: %+v", id, info)
		}
		// i byte stanno nella memoria che il driver scrive nella radio
		var ranges [][2]int
		switch o := m.Opt.(type) {
		case *h777v4Opt:
			ranges = [][2]int{{0, o.memsize}}
		case *t18Opt:
			ranges = o.spec.ranges
		case *rb65Opt:
			ranges = o.spec.ranges
		}
		for _, off := range d.Enc.Offs {
			in := false
			for _, r := range ranges {
				in = in || (off >= r[0] && off < r[1])
			}
			if !in || off < p.off || off >= p.off+p.n {
				t.Fatalf("%s: byte 0x%03X della password fuori dalla parte che si scrive nella radio", id, off)
			}
		}
		// ogni carattere ammesso, da solo e in ogni posizione: scritto, riletto, e cambiano solo i byte della password
		for k := 0; k < l.n; k++ {
			for _, c := range wantSet {
				v := strings.Repeat(string(wantSet[0]), k) + string(c)
				img := append([]byte{}, base...)
				if err := writeSetting(d, img, v); err != nil {
					t.Fatalf("%s: password %q: %v", id, v, err)
				}
				if got, err := readSetting(d, img); err != nil || got != v {
					t.Fatalf("%s: scritta %q, riletta %q (%v)", id, v, got, err)
				}
				want := append([]byte{}, base...)
				for j := 0; j < l.n; j++ {
					want[p.off+j] = 0xFF
					if j < len(v) {
						want[p.off+j] = byte(d.Enc.Map[string(v[j])])
					}
				}
				if !bytes.Equal(img, want) {
					t.Fatalf("%s: password %q: byte diversi da quelli attesi", id, v)
				}
			}
		}
		// password piena, password tolta
		full := wantSet[len(wantSet)-l.n:]
		img := append([]byte{}, base...)
		if err := writeSetting(d, img, full); err != nil {
			t.Fatalf("%s: password piena %q: %v", id, full, err)
		}
		if got, _ := readSetting(d, img); got != full {
			t.Fatalf("%s: password piena riletta %q", id, got)
		}
		if err := writeSetting(d, img, ""); err != nil {
			t.Fatalf("%s: password tolta: %v", id, err)
		}
		if got, err := readSetting(d, img); err != nil || got != "" || !bytes.Equal(img[p.off:p.off+l.n], bytes.Repeat([]byte{0xFF}, l.n)) {
			t.Fatalf("%s: password tolta: %q, byte % X", id, got, img[p.off:p.off+l.n])
		}
		for j := range img {
			if (j < p.off || j >= p.off+l.n) && img[j] != base[j] {
				t.Fatalf("%s: togliendo la password è cambiato il byte 0x%03X", id, j)
			}
		}
		// rifiutate: troppo lunga, caratteri fuori dalla tabella, lettere dove il programma vuole solo cifre
		bad := []string{strings.Repeat("1", l.n+1), "-", "1 2", "é", "1.", "#"}
		if l.digits {
			bad = append(bad, "a", "Z", "12A")
		}
		for _, v := range bad {
			img := append([]byte{}, base...)
			if err := writeSetting(d, img, v); err == nil || !bytes.Equal(img, base) {
				t.Fatalf("%s: password %q accettata (o memoria toccata)", id, v)
			}
		}
		// lettura come il programma del costruttore: un byte fuori dalla tabella vale «vuoto» e si salta
		img = append([]byte{}, base...)
		out := 0xFE // un byte che nessuna delle tre tabelle usa
		if seen[out] {
			t.Fatalf("%s: il byte FE è nella tabella", id)
		}
		for j := 0; j < l.n; j++ {
			img[p.off+j] = byte(out)
		}
		if got, err := readSetting(d, img); err != nil || got != "" {
			t.Fatalf("%s: byte fuori dalla tabella letti come %q (%v)", id, got, err)
		}
		img[p.off+l.n-1] = byte(d.Enc.Map["7"])
		img[p.off] = 0xFF
		if got, err := readSetting(d, img); err != nil || got != "7" {
			t.Fatalf("%s: carattere dopo byte vuoti letto come %q (%v)", id, got, err)
		}
		// anche dove si possono scrivere solo cifre, le lettere già nella radio si leggono
		if l.digits {
			img[p.off], img[p.off+1] = byte(d.Enc.Map["A"]), byte(d.Enc.Map["b"])
			if got, err := readSetting(d, img); err != nil || !strings.HasPrefix(got, "Ab") {
				t.Fatalf("%s: lettere nella radio lette come %q (%v)", id, got, err)
			}
		}
		// l'interfaccia la riceve come testo in chiaro, modificabile, con lunghezza e avviso
		var it *setItem
		for _, g := range settingsView(settingsFor(id, base), base) {
			for i := range g.Items {
				if g.Items[i].Key == "progpw" {
					it = &g.Items[i]
					if g.Name != "Password" {
						t.Fatalf("%s: gruppo %q", id, g.Name)
					}
				}
			}
		}
		if it == nil || it.Type != "str" || it.RO || it.MaxLen != l.n || !it.Critical || it.Warn == "" || it.Label != "Password di programmazione" {
			t.Fatalf("%s: voce della password nell'interfaccia: %+v", id, it)
		}
	}
	// byte scritti a mano dalle tabelle dei programmi (non dalla formula del codice)
	for _, c := range []struct {
		id, pw string
		off    int
		want   []byte
	}{
		{"rt-h777v4", "Ab12zZ", 0xDD, []byte{0x0A, 0x25, 0x01, 0x02, 0x3D, 0x23}}, // tabella A: il byte è la posizione in 0-9 A-Z a-z
		{"rt-b3b", "Ab12zZ", 0xDD, []byte{0x0A, 0x25, 0x01, 0x02, 0x3D, 0x23}},
		{"rt-rb666p", "9", 0xDD, []byte{0x09, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"rt-rb45", "9aZ", 0xFA, []byte{0x09, 0x24, 0x23}},
		{"rt-rb47", "y0", 0xFA, []byte{0x3C, 0x00, 0xFF}},
		{"rt-rb66p", "12ab", 0xDD, []byte{0x01, 0x02, 0x24, 0x25}},
		{"rt-rb66", "907", 0xDD, []byte{0x14, 0x0B, 0x12}}, // tabella B: cifre 11-20
		{"rt-rb65", "123456", 0x3B0, []byte{0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11}},
		{"rt-b3h", "0Az9aZ", 0x3B0, []byte{0x0B, 0x48, 0x15, 0x14, 0x2E, 0x2F}}, // tabella B: A = 72, Z = 47, a = 46, z = 21
		{"rt-rt68b", "S1", 0x3B0, []byte{0x36, 0x0C, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"rt-nr610", "Mm", 0x3B0, []byte{0x3C, 0x22, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"rt-rb66s", "0mF9", 0x3B0, []byte{0x30, 0x00, 0x01, 0x6F, 0xFF, 0xFF}}, // tabella C (mescolata)
		{"rt-rb666s", "cfruz8", 0x3B0, []byte{0x80, 0x8D, 0x99, 0x9C, 0x58, 0x7A}},
	} {
		img := make([]byte, rtvMemSize(genByID[c.id]))
		d := rtvPwSet(t, c.id, img)
		if err := writeSetting(d, img, c.pw); err != nil {
			t.Fatalf("%s: password %q: %v", c.id, c.pw, err)
		}
		want := make([]byte, len(img))
		copy(want[c.off:], c.want)
		if !bytes.Equal(img, want) {
			t.Fatalf("%s: password %q scritta come % X a 0x%03X, attesi % X (o altri byte toccati)", c.id, c.pw, img[c.off:c.off+len(c.want)], c.off, c.want)
		}
		if got, err := readSetting(d, img); err != nil || got != c.pw {
			t.Fatalf("%s: riletta %q (%v)", c.id, got, err)
		}
	}
	// la H777 nuova versione: tutte le voci di CHIRP, nello stesso ordine, più la password; la tabella di CHIRP non cambia
	h := genByID["rt-h777v4"]
	ct := chirpSettings()[h.RClass]
	ht := settingsFor("rt-h777v4", make([]byte, rtvMemSize(h)))
	if ct == nil || ht == nil || len(ht.Settings) != len(ct.Settings)+1 || ht.Source != "CHIRP" {
		t.Fatal("H777 nuova versione: tabella diversa da quella di CHIRP più la password")
	}
	for i, d := range ct.Settings {
		if d.Key == "progpw" || ht.Settings[i].Key != d.Key || ht.Settings[i].Enc != d.Enc {
			t.Fatalf("H777 nuova versione: voce %s di CHIRP cambiata", d.Key)
		}
	}
	if settingsFor("rt-h777v4", make([]byte, rtvMemSize(h))) != ht {
		t.Fatal("H777 nuova versione: tabella ricostruita a ogni richiesta")
	}
	// gli altri modelli di CHIRP della famiglia non ricevono la voce
	for _, id := range []string{"rt-rt21h", "rt-rb29", "rt-rb617", "rt-rt18"} {
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello assente", id)
		}
		if tb := settingsFor(id, make([]byte, 0x400)); tb != nil {
			for _, d := range tb.Settings {
				if d.Key == "progpw" {
					t.Fatalf("%s: ha ricevuto la voce della password dei modelli «Analog1»", id)
				}
			}
		}
	}
}

// La password cambiata fra le impostazioni arriva alla radio con la scrittura, e la
// radio (simulata) con la password nuova si legge ancora: un modello per ogni tipo
// di collegamento e di tabella.
func TestSettingsRetevisPasswordCable447(t *testing.T) {
	nop := func(int, int, string) {}
	for _, c := range []struct{ id, old, pw string }{
		{"rt-h777v4", "", "Ab12zZ"}, {"rt-b3b", "777", "Ab12zZ"}, {"rt-rb666p", "1", "zzzzzz"}, {"rt-rb45", "1a", "9aZ"}, {"rt-rb66", "", "907"},
		{"rt-rb66p", "12", "12ab"}, {"rt-b3h", "abc", "0Az9aZ"}, {"rt-rt68b", "S1", "q"}, {"rt-rb669", "", "RB669x"}, {"rt-rb65", "5", "123456"},
		{"rt-rb66s", "m", "0mF9"},
	} {
		m := genByID[c.id]
		p := rtvPwDefs[c.id]
		radio := noChirpBase(t, m)
		if m.ID == "rt-h777v4" {
			radio = bytes.Repeat([]byte{0xFF}, rtvMemSize(m))
			if err := m.Fam.Set(m, radio, 0, &Chan{Rx: 446006250, Tx: 446006250, Narrow: true, Scan: true}); err != nil {
				t.Fatal(err)
			}
		}
		radio = append([]byte{}, radio...)
		if err := writeSetting(rtvPwSet(t, c.id, radio), radio, c.old); err != nil {
			t.Fatalf("%s: password di partenza: %v", c.id, err)
		}
		e := emuFor(m, radio)
		if e == nil {
			t.Fatalf("%s: nessuna radio simulata", c.id)
		}
		read := func(what string) []byte {
			var img []byte
			emuRun(e, func() {
				var err error
				if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
					t.Fatalf("%s: %s: %v", c.id, what, err)
				}
			})
			return img
		}
		write := func(what string, img []byte) {
			emuRun(e, func() {
				if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
					t.Fatalf("%s: %s: %v", c.id, what, err)
				}
			})
		}
		img := read("lettura con la password di partenza")
		d := rtvPwSet(t, c.id, img)
		if got, err := readSetting(d, img); err != nil || got != c.old {
			t.Fatalf("%s: password letta dalla radio %q (%v), attesa %q", c.id, got, err, c.old)
		}
		// cambiata
		before := e.image()
		if err := writeSetting(d, img, c.pw); err != nil {
			t.Fatalf("%s: %v", c.id, err)
		}
		write("scrittura con la password nuova", img)
		after := e.image()
		want := make([]byte, p.n)
		for j := range want {
			want[j] = 0xFF
			if j < len(c.pw) {
				want[j] = byte(d.Enc.Map[string(c.pw[j])])
			}
		}
		if !bytes.Equal(after[p.off:p.off+p.n], want) {
			t.Fatalf("%s: nella radio la password è % X, attesi % X", c.id, after[p.off:p.off+p.n], want)
		}
		for j := range after {
			if (j < p.off || j >= p.off+p.n) && after[j] != before[j] {
				t.Fatalf("%s: cambiando la password è cambiato nella radio il byte 0x%03X", c.id, j)
			}
		}
		img = read("lettura con la password nuova")
		if got, err := readSetting(rtvPwSet(t, c.id, img), img); err != nil || got != c.pw {
			t.Fatalf("%s: password nuova riletta dalla radio %q (%v)", c.id, got, err)
		}
		// tolta
		if err := writeSetting(rtvPwSet(t, c.id, img), img, ""); err != nil {
			t.Fatal(err)
		}
		write("scrittura senza la password", img)
		if got := e.image()[p.off : p.off+p.n]; !bytes.Equal(got, bytes.Repeat([]byte{0xFF}, p.n)) {
			t.Fatalf("%s: password non tolta dalla radio: % X", c.id, got)
		}
		img = read("lettura dopo aver tolto la password")
		if got, err := readSetting(rtvPwSet(t, c.id, img), img); err != nil || got != "" {
			t.Fatalf("%s: dopo averla tolta la password si legge %q (%v)", c.id, got, err)
		}
	}
}
