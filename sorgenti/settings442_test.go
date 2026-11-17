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
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 4.42: impostazioni delle radio che avevano solo i canali, ricavate dai programmi
// dei costruttori. Posizioni, maschere, numero e valore delle scelte sono trascritti
// qui una seconda volta, direttamente dal codice dei programmi (readInfo / writeInfo
// dei programmi «AnaCps», funzione di scrittura di «BF Series CPS»).

type pos442 struct {
	off  int
	mask byte
	vals []int // valore di ogni scelta (bool: 0 e 1)
}

func b442() []int { return []int{0, 1} }

func seq442(first, n int) []int {
	v := make([]int, n)
	for i := range v {
		v[i] = first + i
	}
	return v
}

// check442: la tabella di kind ha esattamente le voci di want, ognuna alla sua
// posizione, con le sue scelte; ogni scelta si scrive e si rilegge e tocca solo i
// bit della maschera; le posizioni stanno nelle aree che il driver scrive
func check442(t *testing.T, kind string, want map[string]pos442, written func(a int) bool) (*setTable, []byte) {
	t.Helper()
	m := genByID[kind]
	if m == nil {
		t.Fatalf("%s: modello assente", kind)
	}
	if !settingsAvailable(kind) {
		t.Fatalf("%s: impostazioni non disponibili", kind)
	}
	img := noChirpBase(t, m)
	tb := settingsFor(kind, img)
	if tb == nil {
		t.Fatalf("%s: nessuna tabella per un'immagine di %d byte", kind, len(img))
	}
	if len(tb.Settings) != len(want) {
		t.Fatalf("%s: %d impostazioni invece di %d", kind, len(tb.Settings), len(want))
	}
	if settingsFor(kind, img[:len(img)/2]) != nil {
		t.Fatalf("%s: tabella anche per un'immagine troppo corta", kind)
	}
	seen := map[string]bool{}
	for i := range tb.Settings {
		d := &tb.Settings[i]
		w, ok := want[d.Key]
		if !ok || seen[d.Key] || d.ItLabel == "" || d.ItDesc == "" || d.Group == "" {
			t.Fatalf("%s: impostazione inattesa, ripetuta o senza nome: %+v", kind, d)
		}
		seen[d.Key] = true
		offs, masks, err := d.Enc.fields()
		if err != nil || len(offs) != 1 || offs[0] != w.off || masks[0][0] != w.mask {
			t.Fatalf("%s %s: posizione %#x maschera %X (attese %#x %X)", kind, d.Key, offs, masks, w.off, w.mask)
		}
		if !written(w.off) {
			t.Fatalf("%s %s: %#x non è fra le aree scritte dal driver", kind, d.Key, w.off)
		}
		n := len(d.Options)
		if d.Type == "bool" {
			n = 2
		} else if len(d.ItOpts) != n {
			t.Fatalf("%s %s: %d nomi per %d scelte", kind, d.Key, len(d.ItOpts), n)
		}
		if n != len(w.vals) {
			t.Fatalf("%s %s: %d scelte invece di %d", kind, d.Key, n, len(w.vals))
		}
		sh := 0
		for w.mask>>sh&1 == 0 {
			sh++
		}
		for _, start := range []byte{0xFF, 0x00, 0xA5} {
			img[w.off] = start
			for v := 0; v < n; v++ {
				val := strconv.Itoa(v)
				if d.Type == "bool" {
					val = strconv.FormatBool(v == 1)
				}
				before := append([]byte(nil), img...)
				if err := writeSetting(d, img, val); err != nil {
					t.Fatalf("%s %s=%s: %v", kind, d.Key, val, err)
				}
				if got := int(img[w.off]&w.mask) >> sh; got != w.vals[v] {
					t.Fatalf("%s %s=%s: nella memoria %d invece di %d", kind, d.Key, val, got, w.vals[v])
				}
				before[w.off] = before[w.off]&^w.mask | img[w.off]&w.mask
				if !bytes.Equal(before, img) {
					t.Fatalf("%s %s=%s: modificati bit fuori dalla maschera", kind, d.Key, val)
				}
				if back, err := readSetting(d, img); err != nil || back != val {
					t.Fatalf("%s %s: riletto %q invece di %q (%v)", kind, d.Key, back, val, err)
				}
			}
		}
	}
	return tb, img
}

func read442(t *testing.T, kind string, tb *setTable, img []byte, want map[string]string) {
	t.Helper()
	for key, v := range want {
		found := false
		for i := range tb.Settings {
			if d := &tb.Settings[i]; d.Key == key {
				found = true
				if got, err := readSetting(d, img); err != nil || got != v {
					t.Fatalf("%s %s: letto %q invece di %q (%v)", kind, key, got, v, err)
				}
			}
		}
		if !found {
			t.Fatalf("%s: voce %s assente", kind, key)
		}
	}
}

// gemelle della BF-888S con un programma «AnaCps» (BF-889, BF_V1_FH, GT-22)
func TestSettings442Ana(t *testing.T) {
	all := map[string]pos442{
		"voice": {0x2B0, 0xFF, b442()}, "lang": {0x2B1, 0xFF, b442()}, "scan_on": {0x2B2, 0xFF, b442()},
		"vox_on": {0x2B3, 0xFF, b442()}, "vox_level": {0x2B4, 0xFF, seq442(1, 5)}, "vox_rx": {0x2B5, 0xFF, b442()},
		"high_tx": {0x2B6, 0xFF, b442()}, "low_tx": {0x2B7, 0xFF, b442()}, "rx_emerg": {0x2B8, 0xFF, b442()},
		"beep": {0x3C0, 0x01, b442()}, "save": {0x3C0, 0x02, b442()}, "sql": {0x3C1, 0xFF, seq442(0, 10)},
		"side_key": {0x3C2, 0xFF, seq442(0, 4)}, "tot": {0x3C3, 0xFF, seq442(0, 11)},
	}
	for kind, drop := range map[string][]string{"bf-88st": nil, "bf-c2": nil, "bf-gt68": nil, "bf-gt22": {"side_key", "scan_on", "rx_emerg"}} {
		want := map[string]pos442{}
		for k, v := range all {
			want[k] = v
		}
		for _, k := range drop {
			delete(want, k)
		}
		o := genByID[kind].Opt.(*h777Opt)
		tb, img := check442(t, kind, want, func(a int) bool {
			for _, r := range o.ranges {
				if a >= r[0] && a < r[1] {
					return true
				}
			}
			return false
		})
		// blocchi come li scrive il programma: guida vocale accesa in cinese, VOX
		// acceso a livello 4, blocco con tensione bassa; beep spento e risparmio
		// acceso (0x02), squelch 7, tasto laterale «Allarme», TOT 120 s
		copy(img[0x2B0:], []byte{1, 1, 0, 1, 4, 0, 0, 1})
		img[0x2B8] = 1
		copy(img[0x3C0:], []byte{0x02, 7, 3, 4})
		exp := map[string]string{"voice": "true", "lang": "1", "vox_on": "true", "vox_level": "3", "vox_rx": "false", "high_tx": "false",
			"low_tx": "true", "beep": "false", "save": "true", "sql": "7", "tot": "4"}
		if len(drop) == 0 {
			exp["side_key"], exp["scan_on"], exp["rx_emerg"] = "3", "false", "true"
		}
		read442(t, kind, tb, img, exp)
		if !strings.Contains(tb.Source, "programma del costruttore «") {
			t.Fatalf("%s: origine %q", kind, tb.Source)
		}
	}
	// la BF-9700 resta senza impostazioni (programma non leggibile)
	if settingsAvailable("bf-9700") {
		t.Fatal("bf-9700: impostazioni non verificate")
	}
}

// Radtel RT-13: programma «RT-13» V1.7 (AnaCps con tre voci in più)
func TestSettings442RT13(t *testing.T) {
	want := map[string]pos442{
		"voice": {0x2B0, 0xFF, b442()}, "lang": {0x2B1, 0xFF, b442()}, "scan_on": {0x2B2, 0xFF, b442()},
		"vox_on": {0x2B3, 0xFF, b442()}, "vox_level": {0x2B4, 0xFF, seq442(0, 9)}, "vox_rx": {0x2B5, 0xFF, b442()},
		"high_tx": {0x2B6, 0xFF, b442()}, "low_tx": {0x2B7, 0xFF, b442()}, "rx_emerg": {0x2B8, 0xFF, b442()},
		"light": {0x2B9, 0xFF, b442()}, "beep": {0x3C0, 0x01, b442()}, "save": {0x3C0, 0x0E, seq442(0, 6)},
		"sql": {0x3C1, 0xFF, seq442(0, 10)}, "side_key": {0x3C2, 0xFF, seq442(0, 3)}, "tot": {0x3C3, 0xFF, seq442(0, 11)},
		"side_key2": {0x3C4, 0xFF, seq442(0, 4)},
	}
	tb, img := check442(t, "rtl-rt13", want, func(a int) bool {
		for _, r := range rt13WriteAreas {
			if a >= r[0] && a < r[1] {
				return true
			}
		}
		return false
	})
	// come li scrive il programma: VOX a livello 7 (nella memoria 6), spia accesa,
	// beep acceso e risparmio 1:8 (1 + 3·2 = 7), tasto 1 «scansione», tasto 2 «allarme»
	copy(img[0x2B0:], []byte{1, 0, 1, 1, 6, 0, 0, 0, 0, 1})
	copy(img[0x3C0:], []byte{0x07, 4, 2, 10, 3})
	read442(t, "rtl-rt13", tb, img, map[string]string{"voice": "true", "lang": "0", "vox_level": "6", "light": "true", "beep": "true",
		"save": "3", "sql": "4", "side_key": "2", "tot": "10", "side_key2": "3"})
	for i := range tb.Settings {
		if d := &tb.Settings[i]; d.Key == "vox_level" && (d.ItOpts[0] != "1" || d.ItOpts[8] != "9") {
			t.Fatalf("livelli del VOX: %v", d.ItOpts)
		}
	}
}

// serie «D»: programma «BF Series CPS» v1.5.1
func TestSettings442Bfd(t *testing.T) {
	protoOf := map[string]byte{"bf-v1d": 'A', "bf-888sd": 'A', "bf-1901d": 'C', "bf-889d": 'B', "bf-w31d": 'D', "bf-t20d": 'D',
		"bf-v4d": 'B', "bf-1904d": 'A', "bf-1909d": 'B', "bf-v12d": 'A', "bf-v8d": 'B', "bf-v9d": 'A', "bf-v10d": 'A',
		"bf-v2d": 'A', "bf-v5d": 'A', "bf-v6d": 'A'}
	n := 0
	for _, m := range genModels {
		if _, ok := m.Fam.(bfdFam); !ok {
			continue
		}
		n++
		proto, ok := protoOf[m.ID]
		if !ok {
			t.Fatalf("%s: modello della serie D non previsto dalla prova", m.ID)
		}
		want := map[string]pos442{
			"voice": {0x2B0, 0xFF, b442()}, "lang": {0x2B1, 0xFF, b442()}, "vox_on": {0x2B3, 0xFF, b442()},
			"vox_level": {0x2B4, 0xFF, seq442(1, 5)}, "low_tx": {0x2B6, 0xFF, b442()}, "high_tx": {0x2B7, 0xFF, b442()},
			"rx_emerg": {0x2B8, 0xFF, b442()}, "noise": {0x2B9, 0xFF, b442()},
			"beep": {0x3C0, 0x01, b442()}, "save": {0x3C0, 0x02, b442()}, "sql": {0x3C1, 0xFF, []int{0, 1, 5, 9}},
			"tot": {0x3C3, 0xFF, seq442(0, 11)},
		}
		switch proto {
		case 'A':
			want["side_key"] = pos442{0x3C2, 0xFF, seq442(0, 4)}
		case 'B', 'D':
			want["side_key"] = pos442{0x3C2, 0xFF, []int{0, 1, 3}}
		case 'C':
			want["alarm_mode"] = pos442{0x3C0, 0x04, b442()}
		}
		tb, img := check442(t, m.ID, want, func(a int) bool {
			for _, r := range bfdRanges(m) {
				if a >= r[0] && a < r[1] && r[2] == 0 {
					return true
				}
			}
			return false
		})
		// i blocchi come li prepara il programma per: guida vocale accesa, seconda
		// lingua, VOX spento con livello 3, blocco con batteria scarica, «RXEmergency»
		// spento, riduzione del rumore accesa; beep spento, risparmio acceso (e
		// allarme remoto nel tipo C), squelch «Strengthen» (5), tasto «Alarm» (3),
		// TOT 120 s (4)
		copy(img[0x2B0:], []byte{1, 1, 1, 0, 3, 0, 1, 0, 0, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		b0 := byte(0x02)
		if proto == 'C' {
			b0 = 0x06
		}
		copy(img[0x3C0:], []byte{b0, 5, 3, 4, 0xFF, 0xFF, 0xFF, 0xFF})
		exp := map[string]string{"voice": "true", "lang": "1", "vox_on": "false", "vox_level": "2", "low_tx": "true", "high_tx": "false",
			"rx_emerg": "false", "noise": "true", "beep": "false", "save": "true", "sql": "2", "tot": "4"}
		switch proto {
		case 'A':
			exp["side_key"] = "3"
		case 'B', 'D':
			exp["side_key"] = "2"
		case 'C':
			exp["alarm_mode"] = "1"
		}
		read442(t, m.ID, tb, img, exp)
		// le due voci scambiate nel programma del costruttore hanno l'avviso
		for i := range tb.Settings {
			d := &tb.Settings[i]
			if (d.Key == "voice" || d.Key == "beep") != (d.ItWarn == wBfdSwap) {
				t.Fatalf("%s %s: avviso %q", m.ID, d.Key, d.ItWarn)
			}
		}
		// squelch scritto dalla radio con un valore fuori elenco: resta com'è
		img[0x3C1] = 3
		for i := range tb.Settings {
			if d := &tb.Settings[i]; d.Key == "sql" {
				if _, err := readSetting(d, img); err == nil {
					t.Fatalf("%s: squelch 3 letto come una delle scelte", m.ID)
				}
			}
		}
	}
	if n != len(protoOf) {
		t.Fatalf("%d modelli della serie D invece di %d", n, len(protoOf))
	}
}

// gemelle con la tabella di CHIRP ristretta alle voci confermate dal programma
func TestSettings442Twins(t *testing.T) {
	wantN := map[string]int{"bf-uv17r": 85, "bf-uv21r": 85, "bf-uv21l": 85, "bf-uv25plus": 99, "bf-uv28pro": 99,
		"bf-uv29pro": 36, "bf-uv7b": 36, "bf-ar152pro": 36, "bf-t22": 7}
	same := 0
	for _, tw := range twinSettings {
		if !tw.anyClass { // quelle con una classe propria hanno la loro prova (settings443_test.go)
			same++
		}
	}
	if len(wantN) != same {
		t.Fatalf("%d gemelle nella prova, %d nel programma", len(wantN), same)
	}
	for kind, n := range wantN {
		tw := twinSettings[kind]
		m := genByID[kind]
		if m == nil || !m.NoChirp || m.RClass != tw.class {
			t.Fatalf("%s: non è una gemella di %s", kind, tw.class)
		}
		if !settingsAvailable(kind) {
			t.Fatalf("%s: impostazioni non disponibili", kind)
		}
		img := noChirpBase(t, m)
		tb := settingsFor(kind, img)
		src := chirpSettings()[tw.class]
		if tb == nil || src == nil {
			t.Fatalf("%s: nessuna tabella (immagine di %d byte)", kind, len(img))
		}
		if len(tb.Settings) != n {
			t.Fatalf("%s: %d impostazioni invece di %d", kind, len(tb.Settings), n)
		}
		if !strings.HasPrefix(tb.Source, "CHIRP+programma del costruttore «") || tb.Size != src.Size {
			t.Fatalf("%s: origine %q, dimensione %d", kind, tb.Source, tb.Size)
		}
		byKey := map[string]*setDef{}
		for i := range src.Settings {
			byKey[src.Settings[i].Key] = &src.Settings[i]
		}
		rnd := rand.New(rand.NewSource(int64(len(kind)) * 104729))
		for i := range tb.Settings {
			d := &tb.Settings[i]
			o := byKey[d.Key]
			if o == nil {
				t.Fatalf("%s: %s non è nella tabella di %s", kind, d.Key, tw.class)
			}
			if twinMatch(d.Key, tw.drop) || (len(tw.only) > 0 && !twinMatch(d.Key, tw.only)) {
				t.Fatalf("%s: %s doveva essere esclusa", kind, d.Key)
			}
			// stessa posizione e stessi valori della tabella di CHIRP
			oo, om, _ := o.Enc.fields()
			do, dm, _ := d.Enc.fields()
			if len(oo) != len(do) || o.Enc.Kind != d.Enc.Kind || o.Type != d.Type || o.Label != d.Label {
				t.Fatalf("%s %s: codifica diversa da CHIRP", kind, d.Key)
			}
			for j := range oo {
				if oo[j] != do[j] || !bytes.Equal(om[j], dm[j]) {
					t.Fatalf("%s %s: campo %d diverso da CHIRP", kind, d.Key, j)
				}
			}
			lim, limited := tw.limit[d.Key]
			if limited {
				if len(d.Options) != lim || len(d.Enc.Table) != lim || len(o.Options) <= lim {
					t.Fatalf("%s %s: %d scelte (limite %d, CHIRP %d)", kind, d.Key, len(d.Options), lim, len(o.Options))
				}
			} else if len(d.Options) != len(o.Options) {
				t.Fatalf("%s %s: scelte diverse da CHIRP", kind, d.Key)
			}
			for j := range d.Options {
				if d.Options[j] != o.Options[j] || (d.Enc.Kind == "table" && *d.Enc.Table[j] != *o.Enc.Table[j]) {
					t.Fatalf("%s %s: scelta %d diversa da CHIRP", kind, d.Key, j)
				}
			}
			// scrittura e rilettura: cambiano solo i byte della voce
			if d.Type == "info" || d.RO {
				continue
			}
			cur, _ := readSetting(d, img)
			for r := 0; r < 6; r++ {
				val, ok := randSetVal(rnd, d, cur)
				if !ok {
					break
				}
				before := append([]byte(nil), img...)
				if err := writeSetting(d, img, val); err != nil {
					t.Fatalf("%s %s=%q: %v", kind, d.Key, val, err)
				}
				allowed := map[int]bool{}
				for _, a := range do {
					allowed[a] = true
				}
				for _, a := range d.Enc.Offs {
					allowed[a] = true
				}
				for a := range img {
					if img[a] != before[a] && !allowed[a] {
						t.Fatalf("%s %s=%q: modificato il byte %#x, che non è della voce", kind, d.Key, val, a)
					}
				}
				back, err := readSetting(d, img)
				if err != nil {
					t.Fatalf("%s %s=%q: rilettura: %v", kind, d.Key, val, err)
				}
				if back != val {
					bf, e1 := strconv.ParseFloat(back, 64)
					vf, e2 := strconv.ParseFloat(strings.TrimSpace(val), 64)
					if e1 != nil || e2 != nil || bf != vf {
						t.Fatalf("%s %s: riletto %q invece di %q", kind, d.Key, back, val)
					}
				}
				cur = back
			}
		}
		// la tabella di CHIRP non è stata toccata dalle riduzioni
		for k, lim := range tw.limit {
			if o := byKey[k]; o == nil || len(o.Options) <= lim || len(o.Enc.Table) != len(o.Options) {
				t.Fatalf("%s: la riduzione di %s ha modificato la tabella di CHIRP", kind, k)
			}
		}
	}
	// voci tolte perché il programma del costruttore non concorda con CHIRP
	for _, kind := range []string{"bf-uv25plus", "bf-uv28pro", "bf-uv29pro", "bf-uv7b", "bf-ar152pro"} {
		tb := twinTablesMap()[kind]
		for i := range tb.Settings {
			k := tb.Settings[i].Key
			if k == "settings.key2short" || k == "settings.gpstimezone" {
				t.Fatalf("%s: %s presente", kind, k)
			}
			if k == "settings.beep" && len(tb.Settings[i].Options) != 2 {
				t.Fatalf("%s: beep con %d scelte", kind, len(tb.Settings[i].Options))
			}
		}
	}
	// le radio di CHIRP delle stesse classi hanno ancora la tabella intera
	for _, kind := range []string{"bf-uv32", "bf-uv28plus", "bf-w31e"} {
		m := genByID[kind]
		if m == nil || m.NoChirp {
			t.Fatalf("%s: non è un modello di CHIRP", kind)
		}
		if tb := chirpSettings()[m.RClass]; tb == nil || settingsFor(kind, make([]byte, tb.Size)) != tb {
			t.Fatalf("%s: tabella di CHIRP non più intera", kind)
		}
	}
}

// le impostazioni nuove passano dall'interfaccia: file aperto, elenco delle voci,
// modifica, salvataggio e riapertura
func TestSettings442API(t *testing.T) {
	cases := []struct {
		kind, key, val string
		groups         int
	}{
		{"bf-889d", "sql", "3", 3}, {"bf-1901d", "alarm_mode", "1", 3}, {"bf-88st", "side_key", "2", 3},
		{"bf-gt22", "tot", "5", 2}, {"rtl-rt13", "save", "4", 3}, {"bf-uv17r", "settings.squelch", "3", 0},
		{"bf-uv7b", "settings.beep", "Beep", 0}, {"bf-uv25plus", "settings.pttdly", "700 ms", 0}, {"bf-t22", "voxgain", "5", 0},
	}
	for _, tc := range cases {
		apiCase442(t, tc.kind, tc.key, tc.val, tc.groups, noChirpBase(t, genByID[tc.kind]))
	}
}

// apiCase442: la radio kind, aperta da un file con la memoria mem, mostra le
// impostazioni; la voce key si cambia in val, resta nel file salvato e i canali
// non si perdono
func apiCase442(t *testing.T, kind, key, val string, ngroups int, mem []byte) {
	t.Helper()
	tc := struct {
		kind, key, val string
		groups         int
	}{kind, key, val, ngroups}
	{
		m := genByID[tc.kind]
		dir := t.TempDir()
		if err := m.Fam.Set(m, mem, 0, &Chan{Rx: 446006250, Tx: 446006250, Narrow: true, Scan: m.HasScan}); err != nil {
			t.Fatalf("%s: %v", tc.kind, err)
		}
		src := filepath.Join(dir, "radio.img")
		if err := os.WriteFile(src, makeImg(m, mem, nil), 0o644); err != nil {
			t.Fatal(err)
		}
		c := newAPIClient(t)
		r := c.must("radio/open", map[string]string{"path": src, "radio": tc.kind})
		if r["kind"] != tc.kind {
			t.Fatalf("%s: aperta come %v", tc.kind, r["kind"])
		}
		s := c.must("radio/settings", map[string]string{"kind": tc.kind})
		if s["available"] != true {
			t.Fatalf("%s: impostazioni non disponibili: %v", tc.kind, s["reason"])
		}
		groups := s["groups"].([]interface{})
		if tc.groups > 0 && len(groups) != tc.groups {
			t.Fatalf("%s: %d gruppi invece di %d", tc.kind, len(groups), tc.groups)
		}
		twin := twinSettings[tc.kind].class != ""
		if src := s["source"].(string); strings.HasPrefix(src, "CHIRP+") != twin {
			t.Fatalf("%s: origine %q", tc.kind, src)
		}
		n, found := 0, false
		for _, g := range groups {
			for _, it := range g.(map[string]interface{})["items"].([]interface{}) {
				x := it.(map[string]interface{})
				n++
				if x["label"] == "" || x["label"] == nil {
					t.Fatalf("%s: voce %v senza nome", tc.kind, x["key"])
				}
				if x["key"] == tc.key {
					found = true
				}
			}
		}
		if !found || n != len(settingsFor(tc.kind, mem).Settings) {
			t.Fatalf("%s: %d voci, %s presente: %v", tc.kind, n, tc.key, found)
		}
		c.must("radio/settings/set", map[string]interface{}{"kind": tc.kind, "values": map[string]string{tc.key: tc.val}})
		s = c.must("radio/settings", map[string]string{"kind": tc.kind})
		got := ""
		for _, g := range s["groups"].([]interface{}) {
			for _, it := range g.(map[string]interface{})["items"].([]interface{}) {
				if x := it.(map[string]interface{}); x["key"] == tc.key {
					got, _ = x["value"].(string)
				}
			}
		}
		if got != tc.val {
			t.Fatalf("%s: %s = %q dopo la modifica (atteso %q)", tc.kind, tc.key, got, tc.val)
		}
		// salvataggio e riapertura: la modifica è nel file, i canali sono rimasti
		out := filepath.Join(dir, "out.img")
		c.must("radio/save", map[string]interface{}{"radio": tc.kind, "channels": chansFrom(t, r["channels"]), "path": out})
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		mem2, _ := splitImg(b)
		tb := settingsFor(tc.kind, mem2)
		if tb == nil {
			t.Fatalf("%s: file salvato senza tabella", tc.kind)
		}
		for i := range tb.Settings {
			if d := &tb.Settings[i]; d.Key == tc.key {
				if v, err := readSetting(d, mem2); err != nil || v != tc.val {
					t.Fatalf("%s: nel file salvato %s = %q (%v)", tc.kind, tc.key, v, err)
				}
			}
		}
		if ch := m.Fam.Get(m, mem2, 0); ch == nil || ch.Rx != 446006250 {
			t.Fatalf("%s: canale 1 perso nel salvataggio: %+v", tc.kind, ch)
		}
	}
}
