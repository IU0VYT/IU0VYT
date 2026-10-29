// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle radio (menu, display, VOX, DTMF, VFO…): lettura e modifica
// nella memoria della radio (la stessa che viene salvata nel file e scritta via
// cavo insieme ai canali).
//
// Ogni impostazione ha una codifica:
//   table  bit della memoria (campi con maschera) e valore dei bit per ogni scelta
//          (elenco, sì/no, numeri piccoli: un valore per ogni numero);
//   num    numero: raw = v·num/den + add, scritto in binario, BCD o una cifra per
//          byte nei bit della maschera;
//   str    testo: un byte per carattere (tabella dei caratteri) e riempimento.
// Le tabelle delle radio gestite con i driver di CHIRP sono ricavate dal codice di
// CHIRP (verifica/chirp/chirpsettings.py) e verificate con esso (settings_test.go);
// le altre radio hanno tabelle ricavate dai programmi dei costruttori.

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type setEnc struct {
	Kind   string           `json:"kind"`
	Fields [][2]interface{} `json:"fields,omitempty"` // [inizio, maschera esadecimale]
	Table  []*string        `json:"table,omitempty"`
	Endian string           `json:"endian,omitempty"`
	Fmt    string           `json:"fmt,omitempty"`
	Num    int64            `json:"num,omitempty"`
	Den    int64            `json:"den,omitempty"`
	Add    int64            `json:"add,omitempty"`
	Scale  int64            `json:"scale,omitempty"`
	Offs   []int            `json:"offs,omitempty"`
	Map    map[string]int   `json:"map,omitempty"`
	Pad    *int             `json:"pad,omitempty"`
}

type setDef struct {
	Key     string   `json:"key"`
	Group   string   `json:"group"`
	Label   string   `json:"label"`
	Type    string   `json:"type"` // bool | list | int | float | str | numstr | info
	Options []string `json:"options,omitempty"`
	Min     float64  `json:"min,omitempty"`
	Max     float64  `json:"max,omitempty"`
	Step    float64  `json:"step,omitempty"`
	Res     float64  `json:"res,omitempty"`
	Dec     int      `json:"dec,omitempty"`
	Width   int      `json:"width,omitempty"`
	MaxLen  int      `json:"maxlen,omitempty"`
	MinLen  int      `json:"minlen,omitempty"`
	Charset string   `json:"charset,omitempty"`
	Autopad bool     `json:"autopad,omitempty"`
	RO      bool     `json:"ro,omitempty"`
	Value   string   `json:"value,omitempty"` // solo lettura (info)
	Warn    string   `json:"warn,omitempty"`  // avviso di CHIRP
	Enc     *setEnc  `json:"enc,omitempty"`
	// testi in italiano già pronti (tabelle dei programmi dei costruttori)
	ItLabel string   `json:"-"`
	ItDesc  string   `json:"-"`
	ItWarn  string   `json:"-"`
	ItOpts  []string `json:"-"` // etichette delle scelte
}

type setTable struct {
	Class    string   `json:"class"`
	Size     int      `json:"size"`
	Settings []setDef `json:"settings"`
	Source   string   `json:"-"` // origine, mostrata nella finestra
}

//go:embed settings_chirp.json
var settingsChirpJSON []byte

var (
	chirpSetOnce sync.Once
	chirpSet     map[string]*setTable
)

func chirpSettings() map[string]*setTable {
	chirpSetOnce.Do(func() {
		chirpSet = map[string]*setTable{}
		var all map[string]*setTable
		if err := json.Unmarshal(settingsChirpJSON, &all); err == nil {
			for k, t := range all {
				t.Source = "CHIRP"
				chirpSet[k] = t
			}
		}
	})
	return chirpSet
}

// settingsProvider: altre radio (tabelle dai programmi dei costruttori)
type settingsProvider struct {
	has func(kind string) bool
	get func(kind string, img []byte) *setTable
}

var settingsProviders []settingsProvider

// settingsAvailable: la radio kind ha impostazioni modificabili
func settingsAvailable(kind string) bool {
	for _, p := range settingsProviders {
		if p.has(kind) {
			return true
		}
	}
	m, ok := genByID[kind]
	return ok && !m.NoChirp && chirpSettings()[m.RClass] != nil
}

// settingsFor: tabella delle impostazioni della radio kind (nil se non gestite)
func settingsFor(kind string, img []byte) *setTable {
	for _, p := range settingsProviders {
		if p.has(kind) {
			return p.get(kind, img)
		}
	}
	m, ok := genByID[kind]
	if !ok || m.NoChirp {
		return nil
	}
	t := chirpSettings()[m.RClass]
	if t == nil || len(img) < t.Size {
		return nil
	}
	return t
}

// ---------------------------------------------------------------- codifiche

func (e *setEnc) fields() ([]int, [][]byte, error) {
	var offs []int
	var masks [][]byte
	for _, f := range e.Fields {
		off, ok1 := f[0].(float64)
		ms, ok2 := f[1].(string)
		if !ok1 || !ok2 {
			return nil, nil, errors.New("codifica non valida")
		}
		m, err := hex.DecodeString(ms)
		if err != nil {
			return nil, nil, err
		}
		offs = append(offs, int(off))
		masks = append(masks, m)
	}
	return offs, masks, nil
}

func (e *setEnc) maskedHex(img []byte) (string, error) {
	offs, masks, err := e.fields()
	if err != nil {
		return "", err
	}
	var b []byte
	for i, off := range offs {
		for k, m := range masks[i] {
			if off+k >= len(img) {
				return "", errors.New("impostazione fuori dalla memoria")
			}
			b = append(b, img[off+k]&m)
		}
	}
	return hex.EncodeToString(b), nil
}

func (e *setEnc) putMasked(img []byte, val string) error {
	offs, masks, err := e.fields()
	if err != nil {
		return err
	}
	v, err := hex.DecodeString(val)
	if err != nil {
		return err
	}
	n := 0
	for i, off := range offs {
		for k, m := range masks[i] {
			if off+k >= len(img) || n >= len(v) {
				return errors.New("impostazione fuori dalla memoria")
			}
			img[off+k] = img[off+k]&^m | v[n]&m
			n++
		}
	}
	return nil
}

// numeri: valore grezzo nei bit della maschera
func (e *setEnc) rawNum(img []byte) (int64, bool) {
	offs, masks, err := e.fields()
	if err != nil || len(offs) != 1 || offs[0]+len(masks[0]) > len(img) {
		return 0, false
	}
	off, ms := offs[0], masks[0]
	bs := make([]byte, len(ms))
	for k := range ms {
		bs[k] = img[off+k] & ms[k]
	}
	if e.Fmt == "digits" {
		order := bs
		if e.Endian != "be" {
			order = reverse(bs)
		}
		var v int64
		for _, x := range order {
			if x > 9 {
				return 0, false
			}
			v = v*10 + int64(x)
		}
		return v, true
	}
	var v, mv uint64
	if e.Endian == "be" {
		for k := range bs {
			v, mv = v<<8|uint64(bs[k]), mv<<8|uint64(ms[k])
		}
	} else {
		for k := len(bs) - 1; k >= 0; k-- {
			v, mv = v<<8|uint64(bs[k]), mv<<8|uint64(ms[k])
		}
	}
	if mv == 0 {
		return 0, false
	}
	for mv&1 == 0 {
		mv >>= 1
		v >>= 1
	}
	if e.Fmt == "bcd" {
		var d, mul int64 = 0, 1
		for v != 0 {
			n := int64(v & 15)
			if n > 9 {
				return 0, false
			}
			d += n * mul
			mul *= 10
			v >>= 4
		}
		return d, true
	}
	return int64(v), true
}

func reverse(b []byte) []byte {
	out := make([]byte, len(b))
	for i := range b {
		out[len(b)-1-i] = b[i]
	}
	return out
}

func (e *setEnc) putNum(img []byte, raw int64) error {
	offs, masks, err := e.fields()
	if err != nil || len(offs) != 1 {
		return errors.New("codifica numerica non valida")
	}
	off, ms := offs[0], masks[0]
	if off+len(ms) > len(img) || raw < 0 {
		return errors.New("valore non rappresentabile")
	}
	if e.Fmt == "digits" {
		digits := make([]byte, len(ms))
		v := raw
		for k := len(ms) - 1; k >= 0; k-- {
			digits[k] = byte(v % 10)
			v /= 10
		}
		if v != 0 {
			return errors.New("numero troppo grande")
		}
		if e.Endian != "be" {
			digits = reverse(digits)
		}
		for k := range ms {
			img[off+k] = img[off+k]&^ms[k] | digits[k]&ms[k]
		}
		return nil
	}
	var u uint64
	if e.Fmt == "bcd" {
		v := raw
		for sh := 0; v != 0; sh += 4 {
			u |= uint64(v%10) << sh
			v /= 10
		}
	} else {
		u = uint64(raw)
	}
	var mv uint64
	if e.Endian == "be" {
		for k := range ms {
			mv = mv<<8 | uint64(ms[k])
		}
	} else {
		for k := len(ms) - 1; k >= 0; k-- {
			mv = mv<<8 | uint64(ms[k])
		}
	}
	sh := 0
	for x := mv; x != 0 && x&1 == 0; x >>= 1 {
		sh++
	}
	w := u << sh
	if w&^mv != 0 || (u<<sh)>>sh != u {
		return errors.New("numero troppo grande")
	}
	for k := range ms {
		var b byte
		if e.Endian == "be" {
			b = byte(w >> (8 * (len(ms) - 1 - k)))
		} else {
			b = byte(w >> (8 * k))
		}
		img[off+k] = img[off+k]&^ms[k] | b&ms[k]
	}
	return nil
}

// ---------------------------------------------------------------- lettura

func fmtSetNum(v float64, dec int) string {
	if dec <= 0 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', dec, 64)
}

// readSetting: valore attuale come testo ("" e errore se non riconoscibile)
func readSetting(d *setDef, img []byte) (string, error) {
	if d.Type == "info" {
		return d.Value, nil
	}
	e := d.Enc
	if e == nil {
		return "", errors.New("codifica assente")
	}
	switch e.Kind {
	case "table":
		cur, err := e.maskedHex(img)
		if err != nil {
			return "", err
		}
		for i, t := range e.Table {
			if t == nil || *t != cur {
				continue
			}
			switch d.Type {
			case "bool":
				return strconv.FormatBool(i == 1), nil
			case "list":
				return d.Options[i], nil
			case "int":
				return fmtSetNum(d.Min+float64(i)*setStep(d), 0), nil
			}
		}
		return "", errors.New("valore non previsto nella memoria")
	case "num":
		raw, ok := e.rawNum(img)
		if !ok || e.Num == 0 {
			return "", errors.New("valore non leggibile")
		}
		x := (raw - e.Add) * e.Den
		if x%e.Num != 0 {
			return "", errors.New("valore non previsto nella memoria")
		}
		v := float64(x/e.Num) / float64(setScale(e))
		dec := d.Dec
		if d.Type == "int" {
			dec = 0
		}
		return fmtSetNum(v, dec), nil
	case "str":
		rev := map[int]rune{}
		for c, b := range e.Map {
			rev[b] = []rune(c)[0]
		}
		var out []rune
		for _, off := range e.Offs {
			if off >= len(img) {
				return "", errors.New("impostazione fuori dalla memoria")
			}
			b := int(img[off])
			if e.Pad != nil && b == *e.Pad {
				break
			}
			c, ok := rev[b]
			if !ok {
				// carattere fuori dalla tabella (es. ideogrammi scritti dal programma
				// del costruttore): il testo si può solo sostituire
				return strings.TrimRight(string(out), " "), errors.New("testo con caratteri non gestiti")
			}
			out = append(out, c)
		}
		return strings.TrimRight(string(out), " "), nil
	}
	return "", errors.New("codifica sconosciuta")
}

func setStep(d *setDef) float64 {
	if d.Step == 0 {
		return 1
	}
	return d.Step
}

func setScale(e *setEnc) int64 {
	if e.Scale == 0 {
		return 1
	}
	return e.Scale
}

// ---------------------------------------------------------------- scrittura

// writeSetting scrive il valore v (testo come lo mostra l'interfaccia)
func writeSetting(d *setDef, img []byte, v string) error {
	if d.RO || d.Type == "info" || d.Enc == nil {
		return errors.New("impostazione di sola lettura")
	}
	e := d.Enc
	if e.Kind == "str" {
		v = strings.TrimRight(v, " ") // gli spazi iniziali restano (come in CHIRP)
	} else {
		v = strings.TrimSpace(v)
	}
	switch e.Kind {
	case "table":
		idx := -1
		switch d.Type {
		case "bool":
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("valore non valido (sì/no)")
			}
			idx = 0
			if b {
				idx = 1
			}
		case "list":
			for i, o := range d.Options {
				if o == v {
					idx = i
				}
			}
		case "int":
			n, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
			if err != nil || n < d.Min || n > d.Max {
				return fmt.Errorf("valore fuori dall'intervallo %s-%s", fmtSetNum(d.Min, 0), fmtSetNum(d.Max, 0))
			}
			k := (n - d.Min) / setStep(d)
			if k != math.Trunc(k) {
				return fmt.Errorf("il valore deve andare a passi di %s", fmtSetNum(setStep(d), 0))
			}
			idx = int(k)
		}
		if idx < 0 || idx >= len(e.Table) || e.Table[idx] == nil {
			return fmt.Errorf("valore non ammesso: %q", v)
		}
		return e.putMasked(img, *e.Table[idx])
	case "num":
		n, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
		if err != nil {
			return fmt.Errorf("numero non valido: %q", v)
		}
		if (d.Type == "int" || d.Type == "float") && (n < d.Min || n > d.Max) {
			return fmt.Errorf("valore fuori dall'intervallo %s-%s", fmtSetNum(d.Min, d.Dec), fmtSetNum(d.Max, d.Dec))
		}
		if d.Type == "int" && d.Step > 1 && math.Mod(n-d.Min, d.Step) != 0 {
			return fmt.Errorf("il valore deve andare a passi di %s", fmtSetNum(d.Step, 0))
		}
		if n < 0 {
			return errors.New("valore negativo non ammesso")
		}
		iv := int64(math.Round(n * float64(setScale(e))))
		x := iv*e.Num + e.Add*e.Den
		if e.Den == 0 || x%e.Den != 0 {
			return fmt.Errorf("valore non rappresentabile: %q", v)
		}
		return e.putNum(img, x/e.Den)
	case "str":
		r := []rune(v)
		if len(r) > len(e.Offs) || (d.MaxLen > 0 && len(r) > d.MaxLen) {
			return fmt.Errorf("al massimo %d caratteri", len(e.Offs))
		}
		if len(r) < d.MinLen {
			return fmt.Errorf("almeno %d caratteri", d.MinLen)
		}
		if d.Autopad {
			for len(r) < len(e.Offs) {
				r = append(r, ' ')
			}
		}
		for i, off := range e.Offs {
			if off >= len(img) {
				return errors.New("impostazione fuori dalla memoria")
			}
			if i < len(r) {
				b, ok := e.Map[string(r[i])]
				if !ok {
					return fmt.Errorf("carattere non ammesso: %q", string(r[i]))
				}
				img[off] = byte(b)
			} else {
				if e.Pad == nil {
					return errors.New("testo troppo corto per questa impostazione")
				}
				img[off] = byte(*e.Pad)
			}
		}
		return nil
	}
	return errors.New("codifica sconosciuta")
}

// ---------------------------------------------------------------- interfaccia

type setItem struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Orig     string   `json:"orig"` // nome originale (CHIRP / programma del costruttore)
	Type     string   `json:"type"`
	Options  []string `json:"options,omitempty"`
	OptLabel []string `json:"optLabels,omitempty"`
	Min      float64  `json:"min"`
	Max      float64  `json:"max"`
	Step     float64  `json:"step"`
	Dec      int      `json:"dec"`
	MaxLen   int      `json:"maxlen,omitempty"`
	Charset  string   `json:"charset,omitempty"`
	Value    string   `json:"value"`
	Unknown  bool     `json:"unknown,omitempty"` // valore in memoria non previsto
	RO       bool     `json:"ro,omitempty"`
	Desc     string   `json:"desc"`
	Warn     string   `json:"warn,omitempty"`
	Critical bool     `json:"critical,omitempty"`
}

type setGroup struct {
	Name  string    `json:"name"`
	Items []setItem `json:"items"`
}

func settingsView(t *setTable, img []byte) []setGroup {
	var groups []setGroup
	gi := map[string]int{}
	for i := range t.Settings {
		d := &t.Settings[i]
		info := setInfoFor(d)
		it := setItem{Key: d.Key, Label: info.Label, Orig: d.Label, Type: d.Type, Options: d.Options, Min: d.Min, Max: d.Max,
			Step: setStep(d), Dec: d.Dec, MaxLen: d.MaxLen, Charset: d.Charset, RO: d.RO || d.Type == "info",
			Desc: info.Desc, Warn: info.Warn, Critical: info.Critical}
		if d.Type == "float" && d.Res > 0 {
			it.Step = d.Res
		}
		if d.Enc != nil && d.Enc.Kind == "str" && it.MaxLen == 0 {
			it.MaxLen = len(d.Enc.Offs)
		}
		if len(d.ItOpts) == len(d.Options) && len(d.ItOpts) > 0 {
			it.OptLabel = d.ItOpts
		} else {
			for _, o := range d.Options {
				it.OptLabel = append(it.OptLabel, setOptLabel(o))
			}
		}
		v, err := readSetting(d, img)
		if err != nil {
			it.Unknown = true
		}
		it.Value = v
		g := setGroupName(d.Group)
		k, ok := gi[g]
		if !ok {
			gi[g] = len(groups)
			k = len(groups)
			groups = append(groups, setGroup{Name: g})
		}
		groups[k].Items = append(groups[k].Items, it)
	}
	return groups
}

func (a *App) apiRadioSettings(r *http.Request) (interface{}, error) {
	var req struct {
		Kind string `json:"kind"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.mu.Lock()
	sess := a.radios[req.Kind]
	var img []byte
	if sess != nil {
		img = append([]byte{}, sess.Base...)
	}
	a.mu.Unlock()
	if !settingsAvailable(req.Kind) {
		return map[string]interface{}{"available": false, "reason": "Per questa radio la modifica delle impostazioni non è ancora disponibile."}, nil
	}
	if img == nil {
		return map[string]interface{}{"available": false, "reason": "Per vedere e modificare le impostazioni leggi prima la radio via cavo (o apri un suo file)."}, nil
	}
	t := settingsFor(req.Kind, img)
	if t == nil {
		return map[string]interface{}{"available": false, "reason": "La memoria aperta non corrisponde alla tabella delle impostazioni di questa radio: leggi di nuovo la radio."}, nil
	}
	return map[string]interface{}{"available": true, "source": t.Source, "groups": settingsView(t, img)}, nil
}

func (a *App) apiRadioSettingsSet(r *http.Request) (interface{}, error) {
	var req struct {
		Kind   string            `json:"kind"`
		Values map[string]string `json:"values"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sess := a.radios[req.Kind]
	if sess == nil {
		return nil, errors.New("leggi prima la radio o apri un suo file")
	}
	img := append([]byte{}, sess.Base...)
	t := settingsFor(req.Kind, img)
	if t == nil {
		return nil, errors.New("impostazioni non disponibili per questa radio")
	}
	byKey := map[string]*setDef{}
	for i := range t.Settings {
		byKey[t.Settings[i].Key] = &t.Settings[i]
	}
	keys := make([]string, 0, len(req.Values))
	for k := range req.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var errs []string
	for _, k := range keys {
		d := byKey[k]
		if d == nil {
			errs = append(errs, k+": impostazione sconosciuta")
			continue
		}
		if err := writeSetting(d, img, req.Values[k]); err != nil {
			errs = append(errs, setInfoFor(d).Label+": "+err.Error())
		}
	}
	if len(errs) > 0 {
		return nil, errors.New("impostazioni non applicate: " + strings.Join(errs, "; "))
	}
	sess.Base = img
	out := map[string]interface{}{"ok": true, "changed": len(keys), "groups": settingsView(t, img)}
	if req.Kind == kindDAT {
		out["single"] = datSingleList(img) // «Divisione dei canali» cambia la vista della tabella
	}
	return out, nil
}
