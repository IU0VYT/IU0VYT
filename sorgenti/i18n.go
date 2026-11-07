// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lingue dell'interfaccia (4.10). L'interfaccia si traduce nella pagina
// (ui/i18n.js); qui si traducono i testi che il programma mostra con le finestre
// di Windows (messaggi, titoli e filtri di «Apri» e «Salva»), con le stesse
// traduzioni (ui/lang/<lingua>.tsv, chiave = impronta FNV-1a a 32 bit del testo
// italiano con gli spazi ridotti a uno).

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

var uiLangs = []string{"it", "en", "fr", "de", "es", "pt", "ru", "zh", "ja"}

func uiLangKnown(l string) bool {
	for _, x := range uiLangs {
		if x == l {
			return true
		}
	}
	return false
}

// lingua corrente dei testi di Windows ("" o "it" = italiano)
var uiLangCur atomic.Value

func setUILang(l string) {
	if !uiLangKnown(l) {
		l = "it"
	}
	uiLangCur.Store(l)
}

func uiLang() string {
	if v, ok := uiLangCur.Load().(string); ok {
		return v
	}
	return "it"
}

func i18nKey(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(i18nNorm(s)))
	return fmt.Sprintf("%08x", h.Sum32())
}

func i18nNorm(s string) string { return strings.Join(strings.Fields(s), " ") }

func parseLangTSV(b []byte) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 9 || line[8] != '\t' {
			continue
		}
		v := strings.NewReplacer(`\n`, "\n", `\t`, "\t").Replace(line[9:])
		m[line[:8]] = v
	}
	return m
}

type i18nPat struct {
	re    *regexp.Regexp
	key   string
	order []int
}

var i18nData struct {
	sync.Mutex
	dicts map[string]map[string]string
	pats  []i18nPat
	ready bool
}

var i18nPH = regexp.MustCompile(`\{(\d+)\}`)

func i18nLoad() {
	i18nData.Lock()
	defer i18nData.Unlock()
	if i18nData.ready {
		return
	}
	i18nData.ready = true
	i18nData.dicts = map[string]map[string]string{}
	it, err := uiFiles.ReadFile("ui/lang/it.tsv")
	if err != nil {
		return
	}
	for k, t := range parseLangTSV(it) {
		if !i18nPH.MatchString(t) {
			continue
		}
		var re strings.Builder
		re.WriteString("^")
		var order []int
		last := 0
		for _, m := range i18nPH.FindAllStringSubmatchIndex(t, -1) {
			re.WriteString(regexp.QuoteMeta(t[last:m[0]]))
			re.WriteString("(.+?)")
			n := 0
			fmt.Sscanf(t[m[2]:m[3]], "%d", &n)
			order = append(order, n)
			last = m[1]
		}
		re.WriteString(regexp.QuoteMeta(t[last:]) + "$")
		if c, err := regexp.Compile(re.String()); err == nil {
			i18nData.pats = append(i18nData.pats, i18nPat{c, k, order})
		}
	}
}

func i18nDict(l string) map[string]string {
	i18nLoad()
	i18nData.Lock()
	defer i18nData.Unlock()
	if d, ok := i18nData.dicts[l]; ok {
		return d
	}
	d := map[string]string{}
	if b, err := uiFiles.ReadFile("ui/lang/" + l + ".tsv"); err == nil {
		d = parseLangTSV(b)
	}
	i18nData.dicts[l] = d
	return d
}

// T traduce un testo dell'interfaccia nella lingua corrente (o lo lascia com'è)
func T(s string) string { return trLang(uiLang(), s) }

func trLang(l, s string) string {
	if l == "it" || l == "" || strings.TrimSpace(s) == "" {
		return s
	}
	// riga per riga (i messaggi di Windows vanno a capo)
	if strings.Contains(s, "\n") {
		lines := strings.Split(s, "\n")
		for i, x := range lines {
			lines[i] = trLang(l, x)
		}
		return strings.Join(lines, "\n")
	}
	d := i18nDict(l)
	n := i18nNorm(s)
	if t, ok := d[i18nKey(n)]; ok {
		return t
	}
	for _, p := range i18nData.pats {
		m := p.re.FindStringSubmatch(n)
		if m == nil {
			continue
		}
		t, ok := d[p.key]
		if !ok {
			continue
		}
		return i18nPH.ReplaceAllStringFunc(t, func(x string) string {
			k := 0
			fmt.Sscanf(x[1:len(x)-1], "%d", &k)
			for j, o := range p.order {
				if o == k {
					return m[j+1]
				}
			}
			return x
		})
	}
	// messaggio composto «testo: dettaglio»
	if i := strings.Index(n, ": "); i > 0 {
		a, b := n[:i+1], n[i+2:]
		ta, tb := trLang(l, a), trLang(l, b)
		if ta != a || tb != b {
			return ta + " " + tb
		}
	}
	return s
}

// trFilter traduce le descrizioni di un filtro «descrizione|*.est|…»
func trFilter(f string) string {
	p := strings.Split(f, "|")
	for i := 0; i+1 < len(p); i += 2 {
		p[i] = T(p[i])
	}
	return strings.Join(p, "|")
}
