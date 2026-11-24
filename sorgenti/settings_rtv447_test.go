// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
)

// 4.47: impostazioni dei 24 modelli Retevis ricavati dai programmi del costruttore.
// Per ogni modello e ogni voce: ogni scelta si scrive e si rilegge uguale, cambia
// solo i bit della sua maschera nel suo byte, e la voce sta dentro la parte di
// memoria che il driver scrive nella radio. Poi posizioni e valori scritti a mano
// dalla scheda (ANALISI_Retevis_Analog1.md), uno o più per gruppo di programmi.
func TestSettingsRetevis447(t *testing.T) {
	ids := []string{"rt-b3b", "rt-b63b", "rt-rt22p", "rt-rt622p", "rt-rb666", "rt-rb666p", "rt-rb45", "rt-rb645", "rt-rb47",
		"rt-b3h", "rt-b63h", "rt-rb647", "rt-rt68h", "rt-rt668h", "rt-nr10", "rt-nr610", "rt-rb66s", "rt-rb666s", "rt-rb65",
		"rt-rb66", "rt-rb66p", "rt-rt68b", "rt-rt668b", "rt-rb669"}
	if len(rtvSetDefs) != len(ids) {
		t.Fatalf("%d tabelle, attese %d", len(rtvSetDefs), len(ids))
	}
	total := 0
	for _, id := range ids {
		m := genByID[id]
		if m == nil || !m.NoChirp || !settingsAvailable(id) {
			t.Fatalf("%s: modello assente o senza impostazioni", id)
		}
		size := rtvMemSize(m)
		if size == 0 {
			t.Fatalf("%s: memoria sconosciuta", id)
		}
		if settingsFor(id, make([]byte, size-1)) != nil {
			t.Fatalf("%s: tabella data per una memoria troppo corta", id)
		}
		base := make([]byte, size)
		for i := range base {
			base[i] = byte(i*37 + 11)
		}
		tb := settingsFor(id, base)
		// le voci della tabella generata e, sempre per ultima, la password di programmazione
		// (prove sue in settings_rtvpw447_test.go)
		if tb == nil || len(tb.Settings) != len(rtvSetDefs[id])+1 || len(rtvSetDefs[id]) < 5 || tb.Settings[len(tb.Settings)-1].Key != "progpw" {
			t.Fatalf("%s: tabella assente o corta", id)
		}
		pw := rtvPwDefs[id]
		// parte di memoria che il driver scrive
		var ranges [][2]int
		switch o := m.Opt.(type) {
		case *h777v4Opt:
			ranges = [][2]int{{0, o.memsize}}
		case *t18Opt:
			ranges = o.spec.ranges
		case *rb65Opt:
			ranges = o.spec.ranges
		}
		used := map[int]byte{} // bit già occupati da un'altra voce
		keys := map[string]bool{}
		for i := range rtvSetDefs[id] {
			d := &tb.Settings[i]
			src := rtvSetDefs[id][i]
			total++
			if src.off >= pw.off && src.off < pw.off+pw.n {
				t.Fatalf("%s/%s: 0x%03X è un byte della password di programmazione", id, d.Key, src.off)
			}
			if keys[d.Key] || d.Label == "" || d.ItLabel != d.Label || src.mask == 0 {
				t.Fatalf("%s/%s: chiave ripetuta o voce incompleta", id, d.Key)
			}
			keys[d.Key] = true
			if wantGroup := map[bool]string{true: "Tasti laterali", false: "Funzionamento"}[strings.HasPrefix(d.Key, "key") && d.Key != "keylock"]; d.Group != wantGroup {
				t.Fatalf("%s/%s: gruppo %q", id, d.Key, d.Group)
			}
			if used[src.off]&src.mask != 0 {
				t.Fatalf("%s/%s: bit già usati da un'altra voce (0x%03X)", id, d.Key, src.off)
			}
			used[src.off] |= src.mask
			in := false
			for _, r := range ranges {
				in = in || (src.off >= r[0] && src.off < r[1])
			}
			if !in {
				t.Fatalf("%s/%s: 0x%03X fuori dalla parte che si scrive nella radio", id, d.Key, src.off)
			}
			vals := []string{"false", "true"}
			raw := []byte{0, src.mask}
			if d.Type == "list" {
				if len(d.Options) != len(src.opts) || len(d.ItOpts) != len(src.opts) || len(src.vals) != len(src.opts) || len(src.opts) < 2 {
					t.Fatalf("%s/%s: scelte e valori di lunghezza diversa", id, d.Key)
				}
				vals, raw = d.Options, src.vals
				seen := map[byte]bool{}
				for k, v := range raw {
					if v&^src.mask != 0 || seen[v] || src.opts[k] == "" {
						t.Fatalf("%s/%s: valore %02X fuori dalla maschera, ripetuto o senza nome", id, d.Key, v)
					}
					seen[v] = true
				}
			} else if d.Type != "bool" || src.mask&(src.mask-1) != 0 {
				t.Fatalf("%s/%s: una casella deve avere un solo bit", id, d.Key)
			}
			for k, v := range vals {
				img := append([]byte{}, base...)
				if err := writeSetting(d, img, v); err != nil {
					t.Fatalf("%s/%s=%s: %v", id, d.Key, v, err)
				}
				if got, err := readSetting(d, img); err != nil || got != v {
					t.Fatalf("%s/%s: scritto %s, riletto %q (%v)", id, d.Key, v, got, err)
				}
				want := append([]byte{}, base...)
				want[src.off] = want[src.off]&^src.mask | raw[k]
				if !bytes.Equal(img, want) {
					t.Fatalf("%s/%s=%s: byte diversi da quelli attesi (0x%03X: %02X invece di %02X)", id, d.Key, v, src.off, img[src.off], want[src.off])
				}
			}
		}
	}
	if total != 241 {
		t.Fatalf("%d voci in tutto: se sono cambiate va aggiornata anche la scheda", total)
	}
	// posizioni e valori scritti a mano dalla scheda
	for _, c := range []struct {
		id, key, val string
		off          int
		mask, want   byte
	}{
		{"rt-b3b", "voice", "1", 0xD0, 0x0C, 0x08},     // «English» = 08
		{"rt-b3b", "voxsw", "true", 0xD0, 0x10, 0x10},  // «Voice switch»
		{"rt-b3b", "voxdelay", "5", 0xD4, 0xFF, 0x05},  // 3 s
		{"rt-rt22p", "key", "2", 0xD8, 0xFF, 0x02},     // allarme di emergenza
		{"rt-rb666", "voice", "1", 0xD0, 0x0C, 0x08},   // la seconda scelta del programma (vuota) non si offre
		{"rt-rb666p", "key2l", "3", 0xDA, 0xFF, 0x03},  // VOX
		{"rt-rb45", "tot", "6", 0xF4, 0x0F, 0x06},      // mezzo byte basso: 180 s
		{"rt-rb45", "voxdelay", "5", 0xF4, 0xF0, 0x50}, // mezzo byte alto: 3 s
		{"rt-rb645", "key1l", "5", 0xF6, 0xF0, 0x50},   // tasto 1 nel mezzo byte alto
		{"rt-rb645", "key2l", "1", 0xF6, 0x0F, 0x01},   // tasto 2 nel mezzo byte basso
		{"rt-rb47", "codesw", "1", 0xF2, 0x80, 0x80},
		{"rt-b3h", "voxlevel", "0", 0x3C4, 0xFF, 0x01}, // elenco da 1 a 9: il valore è la posizione più 1
		{"rt-b3h", "voxlevel", "8", 0x3C4, 0xFF, 0x09},
		{"rt-b63h", "voxsw", "1", 0x3C0, 0x10, 0x10},
		{"rt-rb647", "specmode", "1", 0x3C2, 0xFF, 0x01},
		{"rt-rt68h", "key1s", "1", 0x3C2, 0xFF, 0x02}, // tasto breve: «Scan» = 2 (il valore 1 è solo del tasto lungo)
		{"rt-rt68h", "key1s", "3", 0x3C2, 0xFF, 0x04}, // blocco del canale
		{"rt-rt68h", "key1l", "1", 0x3C8, 0xFF, 0x01}, // monitor
		{"rt-rt68h", "roger", "true", 0x3C9, 0x01, 0x01},
		{"rt-rt668h", "codesw", "1", 0x3C0, 0x20, 0x20},
		{"rt-rt668h", "specmode", "1", 0x3C5, 0xFF, 0x01},
		{"rt-nr10", "voice", "2", 0x3C0, 0x0C, 0x08}, // Nessuna, Cinese, Inglese
		{"rt-nr610", "key1l", "4", 0x3C5, 0xFF, 0x04},
		{"rt-rb66s", "roger", "true", 0x3C0, 0x20, 0x20},
		{"rt-rb66s", "keylock", "true", 0x3C0, 0x10, 0x10},
		{"rt-rb666s", "abr", "10", 0x3CA, 0xFF, 0x0A},
		{"rt-rb666s", "lcd", "4", 0x3CB, 0xFF, 0x04},
		{"rt-rb65", "tot", "6", 0x3C3, 0xFF, 0x06},
		{"rt-rb66", "voice", "1", 0xD0, 0x0C, 0x04}, // in questi due modelli «English» = 04 (come nel file di fabbrica)
		{"rt-rb66p", "voice", "1", 0xD0, 0x0C, 0x04},
		{"rt-rb66p", "tot", "10", 0xD2, 0xFF, 0x0A},
		{"rt-rt68b", "voice", "1", 0x3C0, 0x04, 0x04}, // un solo bit
		{"rt-rt68b", "roger", "true", 0x3C0, 0x10, 0x10},
		{"rt-rt68b", "voxsw", "true", 0x3C0, 0x20, 0x20},
		{"rt-rt68b", "chlock", "true", 0x3C0, 0x08, 0x08},
		{"rt-rt668b", "codesw", "1", 0x3C0, 0x80, 0x80},
		{"rt-rt668b", "voxlevel", "7", 0x3C4, 0xFF, 0x07},
		{"rt-rb669", "specmode", "1", 0x3C8, 0xFF, 0x01},
		{"rt-rb669", "key1s", "2", 0x3C2, 0xFF, 0x02},
		{"rt-rb669", "codesw", "1", 0x3C0, 0x20, 0x20},
	} {
		m := genByID[c.id]
		img := make([]byte, rtvMemSize(m))
		tb := settingsFor(c.id, img)
		var d *setDef
		for i := range tb.Settings {
			if tb.Settings[i].Key == c.key {
				d = &tb.Settings[i]
			}
		}
		if d == nil {
			t.Fatalf("%s: voce %s assente", c.id, c.key)
		}
		if err := writeSetting(d, img, c.val); err != nil {
			t.Fatalf("%s/%s=%s: %v", c.id, c.key, c.val, err)
		}
		want := make([]byte, len(img))
		want[c.off] = c.want
		if !bytes.Equal(img, want) || c.want&^c.mask != 0 {
			t.Fatalf("%s/%s=%s: atteso %02X a 0x%03X, trovato %02X (o altri byte toccati)", c.id, c.key, c.val, c.want, c.off, img[c.off])
		}
	}
	// voci che non si devono offrire: nascoste nel programma del modello o non chiare
	for _, c := range []struct{ id, key string }{
		{"rt-rt68b", "codesw"}, {"rt-rt68b", "key1s"}, {"rt-rb45", "voice"}, {"rt-rb645", "voice"}, {"rt-rb47", "tot"}, {"rt-rb47", "voice"},
		{"rt-rt22p", "voice"}, {"rt-rb65", "voice"}, {"rt-b3h", "beep"}, {"rt-b3h", "codesw"}, {"rt-rb669", "voice"}, {"rt-rb669", "beep"},
	} {
		for _, d := range settingsFor(c.id, make([]byte, rtvMemSize(genByID[c.id]))).Settings {
			if d.Key == c.key {
				t.Fatalf("%s: la voce %s non va offerta", c.id, c.key)
			}
		}
	}
	// i nomi delle scelte del ritardo del VOX
	if h := rtvHalf(6); strings.Join(h, "|") != "0,5 s|1 s|1,5 s|2 s|2,5 s|3 s" {
		t.Fatalf("rtvHalf: %v", h)
	}
	// i modelli di CHIRP della stessa famiglia tengono la loro tabella
	if rtvSetDefs["rt-rb29"] != nil || rtvSetDefs["rt-h777v4"] != nil || !settingsAvailable("rt-rb29") {
		t.Fatal("tabelle dei modelli di CHIRP cambiate")
	}
}

// Controprova con i file dati di fabbrica dei programmi (non stanno nel repository):
// IU0VYT_FABBRICA_RTV="id=file;id=file…". Ogni voce deve leggere una scelta ammessa.
func TestSettingsRetevisFabbrica(t *testing.T) {
	list := os.Getenv("IU0VYT_FABBRICA_RTV")
	if list == "" {
		t.Skip("IU0VYT_FABBRICA_RTV non impostata")
	}
	n := 0
	for _, kv := range strings.Split(list, ";") {
		p := strings.SplitN(kv, "=", 2)
		m := genByID[p[0]]
		raw, err := os.ReadFile(p[1])
		if m == nil || err != nil {
			t.Fatalf("%s: %v", kv, err)
		}
		img := bytes.Repeat([]byte{0xFF}, rtvMemSize(m))
		for _, ln := range strings.Split(string(raw), "\n") {
			f := strings.Fields(ln)
			if len(f) < 5 || f[0] != "57" {
				continue
			}
			hi, _ := strconv.ParseUint(f[1], 16, 8)
			lo, _ := strconv.ParseUint(f[2], 16, 8)
			a := int(hi)<<8 | int(lo)
			if len(img) <= 0x100 {
				a &= 0xFF // nei file dei modelli da 11 byte il byte alto vale 01 da 0xD1 in poi
			}
			for k, x := range f[4:] {
				v, _ := strconv.ParseUint(x, 16, 8)
				if a+k < len(img) {
					img[a+k] = byte(v)
				}
			}
		}
		var out []string
		for i, d := range settingsFor(p[0], img).Settings {
			if d.Key == "progpw" {
				// nei file di fabbrica la password non c'è. Nei tre modelli da 11 byte il
				// primo dei tre byte (0xFA) vale 00, che nella tabella è il carattere «0»:
				// è il valore lasciato da un elenco nascosto del programma («Side key 1
				// (short)») che usa la stessa posizione; gli altri due byte sono FF e il
				// programma, prima di scrivere la radio, riscrive sempre tutti e tre
				want := ""
				if p[0] == "rt-rb45" || p[0] == "rt-rb645" || p[0] == "rt-rb47" {
					want = "0"
					if !bytes.Equal(img[0xFA:0xFD], []byte{0x00, 0xFF, 0xFF}) {
						t.Errorf("%s: byte della password nel file di fabbrica: % X", p[0], img[0xFA:0xFD])
					}
				}
				if v, err := readSetting(&d, img); err != nil || v != want {
					t.Errorf("%s: password di programmazione nel file di fabbrica: %q (%v)", p[0], v, err)
				}
				continue
			}
			v, err := readSetting(&d, img)
			if err != nil {
				t.Errorf("%s/%s: il valore di fabbrica (%02X a 0x%03X) non è una scelta ammessa: %v", p[0], d.Key, img[rtvSetDefs[p[0]][i].off], rtvSetDefs[p[0]][i].off, err)
				continue
			}
			if d.Type == "list" {
				k, _ := strconv.Atoi(v)
				v = d.ItOpts[k]
			}
			out = append(out, d.Key+"="+v)
			n++
		}
		t.Logf("%s: %s", p[0], strings.Join(out, ", "))
	}
	t.Logf("%d valori di fabbrica letti", n)
}
