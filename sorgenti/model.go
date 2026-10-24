// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Chan è un canale come lo vede il programma (comune alle due radio).
// Le frequenze sono in Hz. I toni sono stringhe: "" = OFF, "88.5" = CTCSS,
// altrimenti un valore originale non interpretato (conservato com'è).
// Raw conserva i dati del canale che il programma non modifica, così da non
// perderli quando si sposta o si copia un canale dentro la stessa radio.
type Chan struct {
	Rx     int64  `json:"rx"`
	Tx     int64  `json:"tx"`
	TxTone string `json:"txTone"`
	RxTone string `json:"rxTone"`
	Power  int    `json:"power"` // 0 alta, 1 media, 2 bassa
	Narrow bool   `json:"narrow"`
	Scan   bool   `json:"scan"`
	Name   string `json:"name"`
	Raw    string `json:"raw,omitempty"`
	// trasmissione disabilitata sul canale (solo ricezione); Tx = Rx
	TxOff bool `json:"txOff,omitempty"`
}

type Report struct {
	Warnings []string `json:"warnings"`
	Notes    []string `json:"notes"`
}

func (r *Report) warn(f string, a ...interface{}) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(f, a...))
}
func (r *Report) note(f string, a ...interface{}) { r.Notes = append(r.Notes, fmt.Sprintf(f, a...)) }

func fmtMHz(hz int64) string { return fmt.Sprintf("%d.%05d", hz/1000000, (hz%1000000)/10) }

// fmtMHz6 scrive i MHz con 6 decimali (formato CHIRP).
func fmtMHz6(hz int64) string { return fmt.Sprintf("%d.%06d", hz/1000000, hz%1000000) }

// fmtMHzShort: MHz senza zeri finali inutili (almeno 3 decimali), per tabelle e Excel.
func fmtMHzShort(hz int64) string {
	s := fmtMHz6(hz)
	for strings.HasSuffix(s, "0") && len(s)-strings.IndexByte(s, '.') > 4 {
		s = s[:len(s)-1]
	}
	return s
}

func parseMHz(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	parts := strings.SplitN(s, ".", 2)
	ip, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || ip < 0 {
		return 0, false
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 6 {
		frac = frac[:6]
	}
	for len(frac) < 6 {
		frac += "0"
	}
	fp, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || fp < 0 {
		return 0, false
	}
	return ip*1000000 + fp, true
}

// Elenco CTCSS standard (decimi di Hz).
var ctcssList = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230,
	1273, 1318, 1365, 1413, 1462, 1500, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835, 1862, 1899, 1928,
	1966, 1995, 2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

func isStdCTCSS(t int) bool {
	for _, v := range ctcssList {
		if v == t {
			return true
		}
	}
	return false
}

func nearestCTCSS(t int) int {
	best, bd := 0, 1<<30
	for _, v := range ctcssList {
		d := v - t
		if d < 0 {
			d = -d
		}
		if d < bd {
			best, bd = v, d
		}
	}
	return best
}

func fmtTone(t int) string {
	if t == 0 {
		return "OFF"
	}
	return fmt.Sprintf("%d.%d", t/10, t%10)
}

// toneStr: decimi di Hz -> stringa del modello ("" = OFF).
func toneStr(t int) string {
	if t == 0 {
		return ""
	}
	return fmtTone(t)
}

// parseTone restituisce i decimi di Hz; ok=false se il valore non è un CTCSS.
func parseTone(s string) (int, bool) {
	s = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(s, ",", ".")))
	if s == "" || s == "OFF" || s == "NONE" || s == "-" {
		return 0, true
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 50 || f > 260 || math.IsNaN(f) {
		return 0, false
	}
	return int(f*10 + 0.5), true
}

var accenti = strings.NewReplacer("À", "A", "È", "E", "É", "E", "Ì", "I", "Ò", "O", "Ù", "U",
	"à", "a", "è", "e", "é", "e", "ì", "i", "ò", "o", "ù", "u", "’", "'", "‘", "'")

func asciiName(s string) string {
	s = accenti.Replace(s)
	b := make([]byte, 0, len(s))
	for _, c := range s {
		if c >= 32 && c < 127 {
			b = append(b, byte(c))
		} else {
			b = append(b, '?')
		}
	}
	return strings.TrimRight(string(b), " ")
}

// shortName accorcia un nome a max caratteri con abbreviazioni comuni.
func shortName(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	repl := []struct{ a, b string }{
		{"MONTE ", "M."}, {"Monte ", "M."}, {"MONTI ", "M."}, {"Monti ", "M."},
		{"SANTA ", "S."}, {"Santa ", "S."}, {"SANTO ", "S."}, {"Santo ", "S."}, {"SAN ", "S."}, {"San ", "S."},
		{" DI ", " D."}, {" di ", " d."}, {" DEL ", " D."}, {" del ", " d."},
	}
	for _, r := range repl {
		if len(s) <= max {
			break
		}
		s = strings.ReplaceAll(s, r.a, r.b)
	}
	if len(s) > max {
		// conserva un suffisso numerico finale ("CAMPOLUNGHINO 3" → "CAMPOLUNGH 3")
		if i := strings.LastIndexByte(s, ' '); i > 0 && len(s)-i-1 <= 3 && isDigits(s[i+1:]) {
			suf := s[i:]
			body := strings.TrimSpace(s[:i])
			if len(body) > max-len(suf) {
				body = strings.TrimSpace(body[:max-len(suf)])
			}
			return body + suf
		}
		s = strings.TrimSpace(s[:max])
	}
	return s
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func powerName(p int) string {
	switch p {
	case 1:
		return "Media"
	case 2:
		return "Bassa"
	}
	return "Alta"
}

// bandOf classifica una frequenza (Hz).
func bandOf(hz int64) string {
	switch {
	case hz >= 136000000 && hz <= 174000000:
		return "VHF"
	case hz >= 400000000 && hz <= 480000000:
		return "UHF"
	}
	return "Altre"
}
