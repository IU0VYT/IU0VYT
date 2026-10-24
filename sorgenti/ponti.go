// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Archivio ponti: interpretazione degli elenchi (IK2ANE, CHIRP, export del
// programma, fogli generici) con segnalazione dei valori dubbi.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Ponte struct {
	ID         string            `json:"id"`
	Nome       string            `json:"nome"` // nome del canale nella radio
	Sigla      string            `json:"sigla"`
	Localita   string            `json:"localita"`
	Provincia  string            `json:"provincia"`
	Regione    string            `json:"regione"`
	Rx         int64             `json:"rx"`
	Tx         int64             `json:"tx"`
	TxTone     string            `json:"txTone"`
	RxTone     string            `json:"rxTone"`
	Narrow     bool              `json:"narrow"`
	Power      int               `json:"power"`
	Tipi       []string          `json:"tipi"`
	TipoOrig   string            `json:"tipoOrig,omitempty"`
	Rete       string            `json:"rete"`
	Nominativo string            `json:"nominativo"`
	Nodo       string            `json:"nodo"`
	Locatore   string            `json:"locatore"`
	Stato      string            `json:"stato"`
	Avvisi     []string          `json:"avvisi"`
	Note       string            `json:"note"`
	Orig       map[string]string `json:"orig,omitempty"`
	Fonte      string            `json:"fonte,omitempty"`
	Riga       int               `json:"riga,omitempty"` // riga nel file di origine (solo in importazione)
}

const (
	statoAttivo = ""
	statoTemp   = "Temporaneamente non attivo"
	statoOff    = "Non attivo"
	statoH24    = "Non H24"
)

// Tipi riconosciuti (ordine usato nei filtri)
var tipiNoti = []string{"Analogico", "EchoLink", "SvxLink", "DMR", "C4FM", "D-Star", "NXDN", "Multimodo", "HotSpot", "Transponder", "Winlink", "Altro"}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------- utilità testo

var senzaAccenti = strings.NewReplacer("à", "a", "è", "e", "é", "e", "ì", "i", "ò", "o", "ù", "u", "À", "A", "È", "E", "É", "E", "Ì", "I", "Ò", "O", "Ù", "U")

func normKey(s string) string {
	s = strings.ToLower(senzaAccenti.Replace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		r := []rune(w)
		for j := range r {
			if j == 0 || r[j-1] == '.' || r[j-1] == '-' || r[j-1] == '\'' {
				r[j] = unicode.ToUpper(r[j])
			}
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// ---------------------------------------------------------------- numeri

// parseNumIT interpreta "145.600", "145,600", "1.362,5", "145600".
func parseNumIT(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
	s = strings.ReplaceAll(s, "−", "-")
	if s == "" {
		return 0, false
	}
	hasDot, hasComma := strings.Contains(s, "."), strings.Contains(s, ",")
	switch {
	case hasDot && hasComma:
		if strings.LastIndex(s, ",") > strings.LastIndex(s, ".") {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	case hasComma:
		if strings.Count(s, ",") > 1 {
			s = strings.ReplaceAll(s, ",", "")
		} else {
			s = strings.ReplaceAll(s, ",", ".")
		}
	case strings.Count(s, ".") > 1:
		s = strings.ReplaceAll(s, ".", "")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// parseFreq: valore di cella -> Hz. unit: "mhz", "khz", "hz" o "" (deduce dal valore).
func parseFreq(c Cell, unit string) (int64, string) {
	var v float64
	txt := strings.ToLower(strings.TrimSpace(c.S))
	if c.IsNum {
		v = c.Num
	} else {
		if txt == "" {
			return 0, "frequenza mancante"
		}
		switch {
		case strings.HasSuffix(txt, "mhz"):
			unit, txt = "mhz", strings.TrimSuffix(txt, "mhz")
		case strings.HasSuffix(txt, "khz"):
			unit, txt = "khz", strings.TrimSuffix(txt, "khz")
		case strings.HasSuffix(txt, "ghz"):
			unit, txt = "ghz", strings.TrimSuffix(txt, "ghz")
		}
		var ok bool
		v, ok = parseNumIT(txt)
		if !ok {
			return 0, fmt.Sprintf("frequenza non valida (\"%s\")", strings.TrimSpace(c.S))
		}
	}
	if v <= 0 {
		return 0, "frequenza mancante"
	}
	switch unit {
	case "ghz":
		v *= 1e9
	case "mhz":
		v *= 1e6
	case "khz":
		v *= 1e3
	case "hz":
	default:
		switch {
		case v < 3000:
			v *= 1e6
		case v < 1e7:
			v *= 1e3
		}
	}
	hz := int64(math.Round(v/10)) * 10
	if hz < 1000000 || hz > 3000000000 {
		return 0, fmt.Sprintf("frequenza fuori scala (\"%s\")", strings.TrimSpace(c.S))
	}
	return hz, ""
}

var reShiftNum = regexp.MustCompile(`([0-9]+(?:[.,][0-9]+)*)\s*(mhz|khz|hz)?`)

// parseShift interpreta lo shift (es. "-600 kHz", "+5.0MHz", "+1.362,5 MHz").
// Restituisce lo scostamento in Hz, un eventuale dubbio e una nota.
func parseShift(s string, rx int64) (off int64, dubbio string) {
	t := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, "−", "-")))
	if t == "" {
		return 0, ""
	}
	switch t {
	case "0", "0.0", "0,0", "-", "simplex", "iso", "isofrequenza", "nessuno", "off":
		return 0, ""
	}
	if strings.Contains(t, "input") || strings.Contains(t, "output") || strings.Contains(t, "ingresso") || strings.Contains(t, "uscita") {
		return 0, fmt.Sprintf("shift \"%s\" non interpretabile: impostata isofrequenza", s)
	}
	sign := 0
	switch {
	case strings.HasPrefix(t, "+"):
		sign = 1
	case strings.HasPrefix(t, "-"):
		sign = -1
	}
	m := reShiftNum.FindStringSubmatch(t)
	if m == nil {
		return 0, fmt.Sprintf("shift \"%s\" non interpretabile: impostata isofrequenza", s)
	}
	v, ok := parseNumIT(m[1])
	if !ok {
		return 0, fmt.Sprintf("shift \"%s\" non interpretabile: impostata isofrequenza", s)
	}
	if v == 0 {
		return 0, ""
	}
	unit := m[2]
	if unit == "" {
		if v < 100 {
			unit = "mhz"
		} else {
			unit = "khz"
		}
	}
	var hz float64
	note := ""
	switch unit {
	case "mhz":
		hz = v * 1e6
		if v >= 100 { // es. "+1.362,5 MHz": il valore è in kHz
			hz = v * 1e3
			note = fmt.Sprintf("shift \"%s\" interpretato come %s MHz", s, strconv.FormatFloat(v/1000, 'f', -1, 64))
		}
	case "khz":
		hz = v * 1e3
	case "hz":
		hz = v
	}
	if sign == 0 {
		sign = -1
		if rx >= 400000000 {
			sign = 1
		}
		note = fmt.Sprintf("segno dello shift mancante in \"%s\": assunto %s", s, map[int]string{1: "+", -1: "-"}[sign])
	}
	off = int64(math.Round(hz/10)) * 10 * int64(sign)
	return off, note
}

var (
	reCC   = regexp.MustCompile(`CC\s*-?\s*([0-9]{1,2})`)
	reDCS  = regexp.MustCompile(`DCS{1,2}\s*-?\s*([0-7]{3})|\bD([0-7]{3})[NI]?\b`)
	reDGID = regexp.MustCompile(`DG-?ID\s*[:=]?\s*([0-9]+)`)
	reNum  = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?`)
)

// parseToneField interpreta la colonna del tono (es. "79,7", "77.0-141.3", "CC1", "DCS023").
func parseToneField(s string) (tone string, dubbi, note []string) {
	t := strings.ToUpper(strings.TrimSpace(s))
	t = strings.ReplaceAll(t, ",", ".")
	if t == "" || t == "OFF" || t == "NO" || t == "-" || t == "NESSUNO" {
		return "", nil, nil
	}
	for _, m := range reCC.FindAllStringSubmatch(t, -1) {
		note = append(note, "color code DMR "+m[1])
	}
	t = reCC.ReplaceAllString(t, " ")
	for _, m := range reDCS.FindAllStringSubmatch(t, -1) {
		code := m[1] + m[2]
		dubbi = append(dubbi, "codice DCS "+code+" indicato ma non gestito dalle radio del programma")
	}
	t = reDCS.ReplaceAllString(t, " ")
	for _, m := range reDGID.FindAllStringSubmatch(t, -1) {
		note = append(note, "DG-ID "+m[1]+" (C4FM)")
	}
	t = reDGID.ReplaceAllString(t, " ")
	var tones []string
	seen := map[string]bool{}
	for _, n := range reNum.FindAllString(t, -1) {
		v, err := strconv.ParseFloat(n, 64)
		if err != nil {
			continue
		}
		switch {
		case v == 1750:
			note = append(note, "accesso con nota a 1750 Hz")
		case v >= 60 && v <= 260:
			ti := int(math.Round(v * 10))
			ts := fmtTone(ti)
			if !isStdCTCSS(ti) {
				nt := nearestCTCSS(ti)
				d := nt - ti
				if d < 0 {
					d = -d
				}
				if d <= 15 {
					dubbi = append(dubbi, fmt.Sprintf("tono %s non standard: usato %s", ts, fmtTone(nt)))
					ts = fmtTone(nt)
				} else {
					dubbi = append(dubbi, fmt.Sprintf("tono %s non valido: ignorato", ts))
					continue
				}
			}
			if !seen[ts] {
				seen[ts] = true
				tones = append(tones, ts)
			}
		}
	}
	if len(tones) > 0 {
		tone = tones[0]
		if len(tones) > 1 {
			dubbi = append(dubbi, fmt.Sprintf("più toni indicati (%s): usato %s", strings.Join(tones, ", "), tone))
		}
	}
	if tone == "" && len(note) == 0 && len(dubbi) == 0 {
		dubbi = append(dubbi, fmt.Sprintf("tono \"%s\" non interpretabile: impostato OFF", strings.TrimSpace(s)))
	}
	return tone, dubbi, note
}

var regioniNote = map[string]string{
	"lazio": "Lazio", "v.aosta": "Valle d'Aosta", "vaosta": "Valle d'Aosta", "valle d'aosta": "Valle d'Aosta", "valledaosta": "Valle d'Aosta",
	"emilia r.": "Emilia-Romagna", "emiliaromagna": "Emilia-Romagna", "emilia-romagna": "Emilia-Romagna", "emiliar": "Emilia-Romagna",
	"trentino a.a.": "Trentino-Alto Adige", "trentinoaa": "Trentino-Alto Adige", "trentinoaltoadige": "Trentino-Alto Adige",
	"friuli v.g.": "Friuli-Venezia Giulia", "friulivg": "Friuli-Venezia Giulia", "friuliveneziagiulia": "Friuli-Venezia Giulia",
	"pugla": "Puglia", "san marino": "San Marino", "sanmarino": "San Marino",
}

var (
	reRegPrefix = regexp.MustCompile(`^[0-9][a-z]?(?:/[0-9])?\s+`)
	reRegDigits = regexp.MustCompile(`^[0-9]+`)
)

func normRegione(s string) string {
	t := strings.ToLower(collapse(s))
	if reRegPrefix.MatchString(t) {
		t = reRegPrefix.ReplaceAllString(t, "")
	} else {
		t = reRegDigits.ReplaceAllString(t, "")
	}
	t = strings.TrimSpace(t)
	if t == "" {
		return ""
	}
	if v, ok := regioniNote[t]; ok {
		return v
	}
	if v, ok := regioniNote[normKey(t)]; ok {
		return v
	}
	return titleCase(t)
}

// tipi dalla colonna tipo/modo e dalla sigla
func parseTipi(tipo, sigla string) []string {
	g := strings.ToUpper(strings.TrimSpace(tipo))
	sg := strings.ToUpper(strings.TrimSpace(sigla))
	set := map[string]bool{}
	if g == "" {
		switch sg {
		case "HOTSPOT":
			set["HotSpot"] = true
		default:
			set["Analogico"] = true
		}
	}
	for _, p := range strings.FieldsFunc(g, func(r rune) bool { return r == '+' || r == '-' || r == '/' || r == ',' || r == ';' || r == ' ' }) {
		switch p {
		case "A", "FM", "NFM", "ANALOGICO", "ANALOG", "AM":
			set["Analogico"] = true
		case "EL", "ECHOLINK":
			set["EchoLink"] = true
			set["Analogico"] = true
		case "SVX", "SVXLINK":
			set["SvxLink"] = true
			set["Analogico"] = true
		case "DMR", "TDMA":
			set["DMR"] = true
		case "DS", "DSTAR", "D-STAR", "DV":
			set["D-Star"] = true
		case "C4FM", "YSF", "FUSION", "WIRES", "WIRESX":
			set["C4FM"] = true
		case "NXDN":
			set["NXDN"] = true
		case "3D":
			set["DMR"], set["D-Star"], set["C4FM"], set["Multimodo"] = true, true, true, true
		case "2D", "MULTIMODO":
			set["Multimodo"] = true
		case "ALLMODE":
			set["Multimodo"], set["Analogico"] = true, true
		case "WL", "WINLINK":
			set["Winlink"] = true
		case "HOTSPOT":
			set["HotSpot"] = true
		case "TRANSPONDER", "TRASPONDER":
			set["Transponder"] = true
		case "D":
			set["D-Star"] = true // "D-STAR" spezzato dal trattino
		case "STAR":
		default:
			set["Altro"] = true
		}
	}
	switch sg {
	case "HOTSPOT":
		set["HotSpot"] = true
	case "E":
		set["EchoLink"], set["Analogico"] = true, true
	case "T":
		set["Transponder"], set["Analogico"] = true, true
	}
	var out []string
	for _, t := range tipiNoti {
		if set[t] {
			out = append(out, t)
		}
	}
	return out
}

// ---------------------------------------------------------------- nomi dei canali

var reProvSuffix = regexp.MustCompile(`\s*\([A-Za-z]{2,3}\)\s*$`)

func pulisciLocalita(s string) string {
	s = reProvSuffix.ReplaceAllString(collapse(s), "")
	return strings.TrimSpace(s)
}

// nomeDaModello: modelli {LOC} {SIGLA} {NOMINATIVO} {PROV}
func nomeDaModello(p *Ponte, modello string) string {
	if modello == "" {
		modello = "{LOC}"
	}
	loc := strings.ToUpper(asciiName(pulisciLocalita(p.Localita)))
	build := func(loc string) string {
		r := strings.NewReplacer("{LOC}", loc, "{SIGLA}", strings.ToUpper(asciiName(collapse(p.Sigla))),
			"{NOMINATIVO}", strings.ToUpper(asciiName(collapse(p.Nominativo))), "{PROV}", strings.ToUpper(collapse(p.Provincia)))
		return collapse(r.Replace(modello))
	}
	n := build(loc)
	// "Alpe Colombino - Giaveno": se il nome è troppo lungo si tiene la prima parte
	if len(n) > 16 && strings.Contains(loc, " - ") {
		if short := build(strings.TrimSpace(strings.SplitN(loc, " - ", 2)[0])); short != "" {
			n = short
		}
	}
	if n == "" {
		for _, alt := range []string{p.Nome, p.Nominativo, p.Sigla, loc} {
			if a := strings.ToUpper(asciiName(collapse(alt))); a != "" {
				return truncName(a)
			}
		}
		return fmtMHz(p.Rx)
	}
	return truncName(n)
}

func truncName(n string) string {
	if len(n) > 16 {
		n = strings.TrimRight(shortName(n, 16), " -,/")
	}
	return n
}

// ---------------------------------------------------------------- importazione da tabella

type ImportResult struct {
	Sheets    []string       `json:"sheets"`
	Sheet     string         `json:"sheet"`
	Formato   string         `json:"formato"`
	HeaderRow int            `json:"headerRow"`
	Colonne   []string       `json:"colonne"` // colonne riconosciute
	Ponti     []*Ponte       `json:"ponti"`
	Scartate  []RigaScartata `json:"scartate"`
}

type RigaScartata struct {
	Riga   int    `json:"riga"`
	Testo  string `json:"testo"`
	Motivo string `json:"motivo"`
}

var sinonimi = map[string]string{
	"nomecanale": "nome", "nome": "nome", "name": "nome", "canale": "nome", "channel": "nome", "channelname": "nome", "alias": "nome",
	"sigla": "sigla", "ripetitore": "sigla", "denominazione": "sigla", "canaleiaru": "sigla",
	"localita": "localita", "qth": "localita", "luogo": "localita", "citta": "localita", "sito": "localita", "location": "localita",
	"provincia": "provincia", "prov": "provincia",
	"regione": "regione", "region": "regione",
	"rxmhz": "rx", "rx": "rx", "frequenza": "rx", "freq": "rx", "frequency": "rx", "frequenzarx": "rx", "uscita": "rx", "output": "rx", "rxfreq": "rx", "rxfrequency": "rx", "receivefrequency": "rx", "frequenzamhz": "rx", "frequenzarxmhz": "rx",
	"txmhz": "tx", "tx": "tx", "frequenzatx": "tx", "entrata": "tx", "input": "tx", "txfreq": "tx", "txfrequency": "tx", "transmitfrequency": "tx", "frequenzatxmhz": "tx",
	"shift": "shift", "shiftmhz": "shift", "offset": "shift", "offsetmhz": "shift", "scostamento": "shift",
	"tonotx": "txTone", "tono": "txTone", "ctcss": "txTone", "tone": "txTone", "subtono": "txTone", "ctcsstx": "txTone", "tonetx": "txTone", "subtonotx": "txTone",
	"tonorx": "rxTone", "ctcssrx": "rxTone", "tonerx": "rxTone", "subtonorx": "rxTone",
	"tipo": "tipo", "modo": "tipo", "mode": "tipo", "grp": "tipo", "sistema": "tipo", "tipologia": "tipo", "modulazione": "tipo",
	"rete": "rete", "network": "rete", "link": "rete",
	"nominativo": "nominativo", "call": "nominativo", "callsign": "nominativo", "nominativoponte": "nominativo",
	"nodo": "nodo", "nodoecholink": "nodo", "node": "nodo", "nnodo": "nodo",
	"locatore": "locatore", "locator": "locatore", "qra": "locatore", "qraloc": "locatore", "qralocator": "locatore",
	"stato": "stato", "status": "stato",
	"note": "note", "comment": "note", "commento": "note", "notes": "note",
	"potenza": "power", "power": "power",
	"larghezza": "narrow", "larghezzabanda": "narrow", "bandwidth": "narrow", "ampiezza": "narrow",
}

// colonne di IK2ANE
var sinonimiIK2ANE = map[string]string{
	"nome": "sigla", "freq": "rx", "shift": "shift", "tono": "txTone", "regione": "regione", "prov": "provincia",
	"localita": "localita", "grp": "tipo", "id": "nominativo", "nnodo": "nodo", "rete": "rete", "qraloc": "locatore",
}

var chirpCols = []string{"location", "name", "frequency", "duplex", "offset", "tone", "rtonefreq", "ctonefreq", "dtcscode", "dtcspolarity", "rxdtcscode", "crossmode", "mode", "tstep", "skip", "power", "comment"}

type colMap struct {
	formato string
	header  int
	cols    map[string]int // campo -> colonna
	unitRx  string
	unitTx  string
	chirp   map[string]int
}

func headerKey(s string) string {
	k := normKey(s)
	return k
}

func detectColumns(sh *Sheet) (*colMap, error) {
	best, bestScore := -1, 0
	var bestMap *colMap
	for r := 0; r < len(sh.Rows) && r < 40; r++ {
		row := sh.Rows[r]
		keys := make([]string, len(row))
		raw := make([]string, len(row))
		for c, cell := range row {
			keys[c] = headerKey(cell.S)
			raw[c] = strings.ToLower(cell.S)
		}
		// CHIRP
		ch := map[string]int{}
		for c, k := range keys {
			for _, n := range chirpCols {
				if k == n {
					if _, ok := ch[n]; !ok {
						ch[n] = c
					}
				}
			}
		}
		if _, ok := ch["frequency"]; ok && len(ch) >= 5 {
			if _, ok2 := ch["duplex"]; ok2 {
				return &colMap{formato: "CHIRP", header: r, chirp: ch}, nil
			}
		}
		// IK2ANE: "(N)ome", "(F)req"
		ik := false
		for _, x := range raw {
			if strings.Contains(x, "(f)req") {
				ik = true
			}
		}
		m := &colMap{formato: "Generico", header: r, cols: map[string]int{}}
		syn := sinonimi
		if ik {
			m.formato = "IK2ANE"
			syn = sinonimiIK2ANE
		}
		for c, k := range keys {
			if k == "" {
				continue
			}
			f, ok := syn[k]
			if !ok {
				// varianti con unità: "rx (mhz)", "frequenza (khz)"
				for _, u := range []string{"mhz", "khz"} {
					if strings.HasSuffix(k, u) {
						if f2, ok2 := syn[strings.TrimSuffix(k, u)]; ok2 && (f2 == "rx" || f2 == "tx" || f2 == "shift") {
							f, ok = f2, true
						}
					}
				}
			}
			if !ok {
				continue
			}
			if _, dup := m.cols[f]; dup {
				continue
			}
			m.cols[f] = c
			if f == "rx" || f == "tx" {
				u := ""
				if strings.Contains(k, "khz") {
					u = "khz"
				} else if strings.Contains(k, "mhz") {
					u = "mhz"
				}
				if f == "rx" {
					m.unitRx = u
				} else {
					m.unitTx = u
				}
			}
		}
		if _, ok := m.cols["rx"]; !ok {
			continue
		}
		if len(m.cols) > bestScore {
			best, bestScore, bestMap = r, len(m.cols), m
		}
	}
	if best < 0 || bestScore < 2 {
		return nil, fmt.Errorf("intestazione non riconosciuta nel foglio \"%s\": servono almeno una colonna con la frequenza e una con il nome o la località", sh.Name)
	}
	return bestMap, nil
}

// legenda dei colori di IK2ANE (foglio "istruzioni"), con valori predefiniti
type legenda struct{ temp, off, h24, shiftNS int }

func leggiLegenda(sheets []*Sheet) legenda {
	l := legenda{temp: 51, off: 10, h24: 54, shiftNS: 47}
	for _, sh := range sheets {
		if normKey(sh.Name) != "istruzioni" {
			continue
		}
		for r := range sh.Rows {
			txt := strings.ToLower(sh.get(r, 1).S)
			f := sh.get(r, 0).Fill
			if f < 0 {
				continue
			}
			switch {
			case strings.Contains(txt, "momentaneamente non in funzione"):
				l.temp = f
			case strings.Contains(txt, "non piu") && strings.Contains(txt, "attivi"):
				l.off = f
			case strings.Contains(txt, "h24"):
				l.h24 = f
			case strings.Contains(txt, "shift non standard"):
				l.shiftNS = f
			}
		}
	}
	return l
}

type importOpts struct {
	Modello string // modello del nome canale
	ToneRx  bool   // usa il tono anche in ricezione
	Fonte   string
}

func cellText(c Cell) string { return strings.TrimSpace(c.S) }

func rowText(row []Cell) string {
	var parts []string
	for _, c := range row {
		if t := cellText(c); t != "" {
			parts = append(parts, t)
		}
	}
	s := strings.Join(parts, " | ")
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}

func importSheets(sheets []*Sheet, want string, o importOpts) (*ImportResult, error) {
	res := &ImportResult{Ponti: []*Ponte{}, Scartate: []RigaScartata{}, Colonne: []string{}}
	for _, s := range sheets {
		res.Sheets = append(res.Sheets, s.Name)
	}
	var sh *Sheet
	var cm *colMap
	if want != "" {
		for _, s := range sheets {
			if s.Name == want {
				sh = s
			}
		}
		if sh == nil {
			return nil, fmt.Errorf("foglio \"%s\" non trovato", want)
		}
		var err error
		if cm, err = detectColumns(sh); err != nil {
			return nil, err
		}
	} else {
		// preferisce un foglio chiamato "ponti", altrimenti il primo riconoscibile
		order := append([]*Sheet{}, sheets...)
		sort.SliceStable(order, func(i, j int) bool { return normKey(order[i].Name) == "ponti" && normKey(order[j].Name) != "ponti" })
		var firstErr error
		for _, s := range order {
			m, err := detectColumns(s)
			if err == nil {
				sh, cm = s, m
				break
			}
			if firstErr == nil {
				firstErr = err
			}
		}
		if sh == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("il file non contiene fogli")
			}
			return nil, firstErr
		}
	}
	res.Sheet, res.Formato, res.HeaderRow = sh.Name, cm.formato, cm.header+1
	for k := range cm.cols {
		res.Colonne = append(res.Colonne, k)
	}
	for k := range cm.chirp {
		res.Colonne = append(res.Colonne, k)
	}
	sort.Strings(res.Colonne)
	leg := leggiLegenda(sheets)
	for r := cm.header + 1; r < len(sh.Rows); r++ {
		row := sh.Rows[r]
		if rowText(row) == "" {
			continue
		}
		var p *Ponte
		var motivo string
		if cm.formato == "CHIRP" {
			p, motivo = chirpRow(sh, r, cm)
		} else {
			p, motivo = genericRow(sh, r, cm, leg, o)
		}
		if p == nil {
			res.Scartate = append(res.Scartate, RigaScartata{Riga: r + 1, Testo: rowText(row), Motivo: motivo})
			continue
		}
		p.Riga = r + 1
		p.ID = newID()
		p.Fonte = o.Fonte
		if p.Nome == "" {
			p.Nome = nomeDaModello(p, o.Modello)
		} else {
			p.Nome = truncName(strings.ToUpper(asciiName(collapse(p.Nome))))
		}
		if o.ToneRx && p.RxTone == "" {
			p.RxTone = p.TxTone
		}
		if p.Avvisi == nil {
			p.Avvisi = []string{}
		}
		if p.Tipi == nil {
			p.Tipi = []string{}
		}
		res.Ponti = append(res.Ponti, p)
	}
	return res, nil
}

func genericRow(sh *Sheet, r int, cm *colMap, leg legenda, o importOpts) (*Ponte, string) {
	get := func(f string) Cell {
		if c, ok := cm.cols[f]; ok {
			return sh.get(r, c)
		}
		return Cell{Fill: -1, Font: -1}
	}
	txt := func(f string) string { return collapse(cellText(get(f))) }
	p := &Ponte{Orig: map[string]string{}}
	rxCell := get("rx")
	rx, why := parseFreq(rxCell, cm.unitRx)
	if rx == 0 {
		if why == "frequenza mancante" && rowText(sh.Rows[r]) != "" {
			why = "nessuna frequenza (riga di testo o vuota)"
		}
		return nil, why
	}
	p.Rx, p.Tx = rx, rx
	p.Sigla = txt("sigla")
	p.Localita = txt("localita")
	p.Provincia = strings.ToUpper(txt("provincia"))
	p.Regione = normRegione(txt("regione"))
	p.Rete = txt("rete")
	p.Nominativo = strings.ToUpper(txt("nominativo"))
	p.Nodo = txt("nodo")
	p.Locatore = strings.ToUpper(txt("locatore"))
	p.Nome = txt("nome")
	notes := []string{}
	if t := txt("note"); t != "" {
		notes = append(notes, t)
	}
	// TX: colonna esplicita oppure shift
	if _, ok := cm.cols["tx"]; ok && cellText(get("tx")) != "" {
		tx, why := parseFreq(get("tx"), cm.unitTx)
		if tx == 0 {
			p.Avvisi = append(p.Avvisi, "TX: "+why+": impostata isofrequenza")
		} else {
			p.Tx = tx
		}
	} else if _, ok := cm.cols["shift"]; ok {
		sc := get("shift")
		s := cellText(sc)
		if sc.IsNum {
			s = strconv.FormatFloat(sc.Num, 'f', -1, 64)
			if sc.Num > 0 {
				s = "+" + s
			}
		}
		off, dub := parseShift(s, rx)
		if dub != "" {
			p.Avvisi = append(p.Avvisi, dub)
			p.Orig["shift"] = s
		}
		if s == "" && cm.formato == "IK2ANE" && !strings.EqualFold(p.Sigla, "hotspot") && p.Sigla != "E" && p.Sigla != "T" {
			p.Avvisi = append(p.Avvisi, "shift non indicato: impostata isofrequenza")
		}
		p.Tx = rx + off
		if p.Tx < 1000000 || p.Tx > 3000000000 {
			p.Avvisi = append(p.Avvisi, fmt.Sprintf("shift \"%s\" porta fuori scala: impostata isofrequenza", s))
			p.Tx = rx
		}
		if leg.shiftNS >= 0 && sc.Fill == leg.shiftNS && cm.formato == "IK2ANE" && off != 0 {
			notes = append(notes, "shift non standard (segnalato nel file)")
		}
	}
	// toni
	tc := get("txTone")
	ts := cellText(tc)
	if tc.IsNum {
		ts = strconv.FormatFloat(tc.Num, 'f', -1, 64)
	}
	tone, dubbi, tnote := parseToneField(ts)
	p.TxTone = tone
	p.Avvisi = append(p.Avvisi, dubbi...)
	notes = append(notes, tnote...)
	if len(dubbi) > 0 || (ts != "" && tone != "" && ts != tone) {
		p.Orig["tono"] = ts
	}
	if cm.formato == "IK2ANE" && tc.Font == 10 && tone != "" {
		notes = append(notes, "tono non regionale (in rosso nel file)")
	}
	if _, ok := cm.cols["rxTone"]; ok {
		rc := get("rxTone")
		rs := cellText(rc)
		if rc.IsNum {
			rs = strconv.FormatFloat(rc.Num, 'f', -1, 64)
		}
		rt, rd, _ := parseToneField(rs)
		p.RxTone = rt
		for _, d := range rd {
			p.Avvisi = append(p.Avvisi, "RX: "+d)
		}
	}
	// tipo
	tipo := txt("tipo")
	p.TipoOrig = tipo
	if _, ok := cm.cols["tipo"]; ok || cm.formato == "IK2ANE" {
		p.Tipi = parseTipi(tipo, p.Sigla)
	} else {
		p.Tipi = []string{"Analogico"}
	}
	// stato
	if cm.formato == "IK2ANE" {
		for _, c := range []Cell{get("sigla"), rxCell} {
			switch c.Fill {
			case leg.off:
				p.Stato = statoOff
			case leg.temp:
				if p.Stato == "" {
					p.Stato = statoTemp
				}
			case leg.h24:
				if p.Stato == "" {
					p.Stato = statoH24
				}
			}
		}
	} else {
		p.Stato = normStato(txt("stato"))
	}
	// potenza e larghezza di banda (export del programma o fogli generici)
	if pw := strings.ToLower(txt("power")); pw != "" {
		p.Power = parsePower(pw)
	}
	if nw := strings.ToLower(txt("narrow")); nw != "" {
		p.Narrow = strings.HasPrefix(nw, "s") || strings.HasPrefix(nw, "n") || strings.Contains(nw, "12.5") || strings.Contains(nw, "12,5")
	}
	p.Note = strings.Join(notes, "; ")
	if len(p.Orig) == 0 {
		p.Orig = nil
	}
	return p, ""
}

func normStato(s string) string {
	k := normKey(s)
	switch {
	case k == "", k == "attivo", k == "ok", k == "si":
		return statoAttivo
	case strings.Contains(k, "temp") || strings.Contains(k, "moment"):
		return statoTemp
	case strings.Contains(k, "h24"):
		return statoH24
	case strings.Contains(k, "non") || strings.Contains(k, "off") || strings.Contains(k, "spento"):
		return statoOff
	}
	return statoAttivo
}

// parsePower: "Alta", "Media", "Bassa", "High", "Low", "5.0W", "1W"...
func parsePower(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "a"), strings.HasPrefix(s, "h"):
		return 0
	case strings.HasPrefix(s, "m"):
		return 1
	case strings.HasPrefix(s, "b"), strings.HasPrefix(s, "l"):
		return 2
	}
	v, ok := parseNumIT(strings.TrimSuffix(strings.TrimSpace(s), "w"))
	if ok {
		switch {
		case v >= 4:
			return 0
		case v >= 1.5:
			return 1
		default:
			return 2
		}
	}
	return 0
}

func chirpRow(sh *Sheet, r int, cm *colMap) (*Ponte, string) {
	get := func(f string) string {
		if c, ok := cm.chirp[f]; ok {
			return cellText(sh.get(r, c))
		}
		return ""
	}
	fc := Cell{S: get("frequency")}
	if c, ok := cm.chirp["frequency"]; ok {
		fc = sh.get(r, c)
	}
	rx, why := parseFreq(fc, "mhz")
	if rx == 0 {
		return nil, why
	}
	p := &Ponte{Rx: rx, Tx: rx, Nome: collapse(get("name")), Tipi: []string{"Analogico"}}
	offC := Cell{S: get("offset")}
	if c, ok := cm.chirp["offset"]; ok {
		offC = sh.get(r, c)
	}
	var off int64
	if strings.TrimSpace(offC.S) != "" {
		off, _ = parseFreq(offC, "mhz")
		if off == 0 { // valori piccoli (es. 0.600) sono sotto la soglia di parseFreq
			if v, ok := parseNumIT(offC.S); ok {
				off = int64(math.Round(v*1e5)) * 10
			}
		}
	}
	switch strings.ToLower(get("duplex")) {
	case "+":
		p.Tx = rx + off
	case "-":
		p.Tx = rx - off
	case "split":
		if off > 0 {
			p.Tx = off
		}
	case "off":
		p.Avvisi = append(p.Avvisi, "trasmissione disattivata nel file CHIRP (Duplex=off): impostata isofrequenza")
	}
	rt, ct := get("rtonefreq"), get("ctonefreq")
	norm := func(s string) string {
		if v, ok := parseTone(s); ok {
			return toneStr(v)
		}
		return ""
	}
	switch strings.ToLower(get("tone")) {
	case "tone":
		p.TxTone = norm(rt)
	case "tsql":
		p.TxTone, p.RxTone = norm(ct), norm(ct)
		if p.TxTone == "" {
			p.TxTone, p.RxTone = norm(rt), norm(rt)
		}
	case "cross":
		cmode := get("crossmode")
		switch {
		case strings.HasPrefix(cmode, "Tone->Tone"):
			p.TxTone, p.RxTone = norm(rt), norm(ct)
		case strings.HasPrefix(cmode, "->Tone"):
			p.RxTone = norm(ct)
		case strings.HasPrefix(cmode, "Tone->"):
			p.TxTone = norm(rt)
		default:
			p.Avvisi = append(p.Avvisi, "modo toni CHIRP \""+cmode+"\" non gestito: toni OFF")
		}
	case "dtcs":
		p.Avvisi = append(p.Avvisi, "tono DCS "+get("dtcscode")+" non gestito: impostato OFF")
	}
	switch strings.ToUpper(get("mode")) {
	case "NFM":
		p.Narrow = true
	case "DV":
		p.Tipi = []string{"D-Star"}
	case "DMR":
		p.Tipi = []string{"DMR"}
	case "AM":
		p.Avvisi = append(p.Avvisi, "modulazione AM")
	}
	if pw := get("power"); pw != "" {
		p.Power = parsePower(pw)
	}
	p.Note = get("comment")
	return p, ""
}

// ---------------------------------------------------------------- conversione ponte -> canale

func ponteChan(p *Ponte) *Chan {
	return &Chan{Rx: p.Rx, Tx: p.Tx, TxTone: p.TxTone, RxTone: p.RxTone, Power: p.Power, Narrow: p.Narrow, Scan: true, Name: p.Nome}
}

func chanPonte(c *Chan) *Ponte {
	return &Ponte{ID: newID(), Nome: c.Name, Rx: c.Rx, Tx: c.Tx, TxTone: c.TxTone, RxTone: c.RxTone, Power: c.Power, Narrow: c.Narrow,
		Tipi: []string{"Analogico"}, Avvisi: []string{}}
}
