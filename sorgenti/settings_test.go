//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle radio gestite con i driver di CHIRP: la tabella
// settings_chirp.json confrontata con il codice di CHIRP
// (verifica/chirp/chirpsettings.py), con CHIRP_DIR:
// 1) lettura: ogni valore letto qui uguale a quello mostrato da CHIRP;
// 2) scrittura: combinazioni casuali di modifiche applicate qui e da CHIRP
//    (set_settings) alla stessa memoria: memoria identica byte per byte.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestSettingsTables(t *testing.T) {
	n := 0
	for cls, tb := range chirpSettings() {
		seen := map[string]bool{}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			if seen[d.Key] {
				t.Errorf("%s: chiave ripetuta %s", cls, d.Key)
			}
			seen[d.Key] = true
			if d.Type != "info" && d.Enc == nil {
				t.Errorf("%s/%s: codifica assente", cls, d.Key)
			}
			if d.Enc != nil && d.Enc.Kind == "table" && d.Type == "list" && len(d.Enc.Table) != len(d.Options) {
				t.Errorf("%s/%s: tabella e scelte di lunghezza diversa", cls, d.Key)
			}
			n++
		}
	}
	for _, m := range genModels {
		if !m.NoChirp && chirpSettings()[m.RClass] == nil {
			t.Logf("%s (%s): nessuna impostazione", m.ID, m.RClass)
		}
	}
	t.Logf("%d classi, %d impostazioni", len(chirpSettings()), n)
}

// chirpSetVal: valore per chirpsettings.py (JSON con il tipo di CHIRP)
func chirpSetVal(d *setDef, v string) interface{} {
	switch d.Type {
	case "bool":
		b, _ := strconv.ParseBool(v)
		return b
	case "int":
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	case "float":
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case "numstr":
		if d.Autopad {
			for len(v) < d.Width {
				v += " "
			}
		}
		return v
	case "str":
		if d.Autopad && d.MaxLen > 0 {
			for len([]rune(v)) < d.MaxLen {
				v += " "
			}
		}
		return v
	}
	return v
}

func randSetVal(rnd *rand.Rand, d *setDef, cur string) (string, bool) {
	switch d.Type {
	case "bool":
		return strconv.FormatBool(rnd.Intn(2) == 1), true
	case "list":
		var ok []int
		for i := range d.Options {
			if d.Enc.Kind != "table" || d.Enc.Table[i] != nil {
				ok = append(ok, i)
			}
		}
		if len(ok) == 0 {
			return "", false
		}
		return d.Options[ok[rnd.Intn(len(ok))]], true
	case "int":
		st := setStep(d)
		n := int(math.Floor((d.Max-d.Min)/st)) + 1
		v := d.Min + float64(rnd.Intn(n))*st
		if d.Enc.Kind == "table" && d.Enc.Table[int((v-d.Min)/st)] == nil {
			return "", false
		}
		return fmtSetNum(v, 0), true
	case "float":
		steps := int(math.Round((d.Max - d.Min) / d.Res))
		return fmtSetNum(d.Min+float64(rnd.Intn(steps+1))*d.Res, d.Dec), true
	case "numstr":
		f, err := strconv.ParseFloat(strings.TrimSpace(cur), 64)
		if err != nil {
			return "", false
		}
		if d.Dec > 0 {
			f += float64(rnd.Intn(11)-5) * math.Pow(10, -float64(d.Dec)) * 100
		} else {
			f += float64(rnd.Intn(11) - 5)
		}
		if f <= 0 {
			return "", false
		}
		return fmtSetNum(f, d.Dec), true
	case "str":
		cs := []rune(d.Charset)
		if len(cs) == 0 {
			return "", false
		}
		ml := len(d.Enc.Offs)
		if d.MaxLen > 0 && d.MaxLen < ml {
			ml = d.MaxLen
		}
		ln := d.MinLen + rnd.Intn(ml-d.MinLen+1)
		var r []rune
		for i := 0; i < ln; i++ {
			r = append(r, cs[rnd.Intn(len(cs))])
		}
		s := strings.TrimRight(string(r), " ")
		if len([]rune(s)) < d.MinLen {
			return "", false
		}
		if d.Enc.Pad == nil && !d.Autopad && len([]rune(s)) < len(d.Enc.Offs) {
			return "", false
		}
		return s, true
	}
	return "", false
}

func TestSettingsChirp(t *testing.T) {
	dir := os.Getenv("CHIRP_DIR")
	if dir == "" {
		t.Skip("CHIRP_DIR non impostato")
	}
	py := filepath.Join("verifica", "chirp", "chirpsettings.py")
	classes := make([]string, 0, len(chirpSettings()))
	for c := range chirpSettings() {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	if only := os.Getenv("SETTINGS_CLASS"); only != "" {
		classes = []string{only}
	}
	for _, cls := range classes {
		tb := chirpSettings()[cls]
		var m *genModel
		for _, g := range genModels {
			if g.RClass == cls && !g.NoChirp {
				m = g
				break
			}
		}
		if m == nil {
			t.Errorf("%s: nessun modello", cls)
			continue
		}
		t.Run(cls, func(t *testing.T) {
			img := chirpImageFor(t, m)
			tmp := t.TempDir()
			run := func(args ...string) []byte {
				cmd := exec.Command("python3", append([]string{py}, args...)...)
				cmd.Env = append(os.Environ(), "CHIRP_DIR="+dir)
				out, err := cmd.Output()
				if err != nil {
					if ee, ok := err.(*exec.ExitError); ok {
						t.Fatalf("chirpsettings %v: %v\n%s", args, err, ee.Stderr)
					}
					t.Fatal(err)
				}
				return out
			}
			bp := filepath.Join(tmp, "base.img")
			run("base", cls, img, bp)
			base, _ := os.ReadFile(bp)
			// 1) lettura
			var want map[string]interface{}
			if err := json.Unmarshal(run("values", cls, bp), &want); err != nil {
				t.Fatal(err)
			}
			for i := range tb.Settings {
				d := &tb.Settings[i]
				if d.Type == "info" {
					continue
				}
				got, err := readSetting(d, base)
				w := want[d.Key]
				if err != nil {
					// testi con caratteri non ASCII (es. nomi in cinese dell'immagine di
					// prova): mostrati come «valore non previsto», si possono riscrivere
					if x, ok := w.(string); ok && d.Type == "str" && strings.IndexFunc(x, func(r rune) bool { return r > 126 }) >= 0 {
						continue
					}
					t.Errorf("%s: %v", d.Key, err)
					continue
				}
				ok := false
				switch x := w.(type) {
				case bool:
					ok = got == strconv.FormatBool(x)
				case float64:
					g, _ := strconv.ParseFloat(got, 64)
					ok = math.Abs(g-x) < 1e-6
				case string:
					if d.Type == "numstr" {
						a, _ := strconv.ParseFloat(strings.TrimSpace(x), 64)
						g, _ := strconv.ParseFloat(got, 64)
						ok = math.Abs(a-g) < 1e-6
					} else {
						ok = got == strings.TrimRight(x, " ") || got == x
					}
				}
				if !ok {
					t.Errorf("%s: qui %q, CHIRP %v", d.Key, got, w)
				}
			}
			// 2) modifiche casuali
			rnd := rand.New(rand.NewSource(int64(len(cls)) * 7919))
			var mutable []*setDef
			for i := range tb.Settings {
				if d := &tb.Settings[i]; d.Type != "info" && !d.RO {
					mutable = append(mutable, d)
				}
			}
			rounds := 4
			total := 0
			for r := 0; r < rounds && len(mutable) > 0; r++ {
				ch := map[string]interface{}{}
				mine := append([]byte{}, base...)
				for _, k := range rnd.Perm(len(mutable)) {
					if len(ch) >= 14 {
						break
					}
					d := mutable[k]
					cur, _ := readSetting(d, mine)
					v, ok := randSetVal(rnd, d, cur)
					if !ok {
						continue
					}
					if err := writeSetting(d, mine, v); err != nil {
						t.Errorf("%s = %q: %v", d.Key, v, err)
						continue
					}
					ch[d.Key] = chirpSetVal(d, v)
				}
				b, _ := json.Marshal(ch)
				cp := filepath.Join(tmp, fmt.Sprintf("mod%d.json", r))
				op := filepath.Join(tmp, fmt.Sprintf("out%d.img", r))
				_ = os.WriteFile(cp, b, 0o644)
				run("apply", cls, bp, cp, op)
				theirs, _ := os.ReadFile(op)
				if !bytes.Equal(theirs, mine) {
					t.Fatalf("modifiche %s: memoria diversa da CHIRP: %s", b, diffRanges(mine, theirs))
				}
				total += len(ch)
			}
			t.Logf("%d impostazioni lette come CHIRP; %d modifiche in %d combinazioni: memoria identica", len(tb.Settings), total, rounds)
		})
	}
}
