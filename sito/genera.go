// IU0VYT Flash — pagine del sito nelle lingue del programma
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA.
//
// docs/index.html è l'originale italiano e l'unica pagina da modificare a mano.
// I testi da tradurre sono segnati con l'attributo data-t (si traduce il contenuto
// dell'elemento) o data-ta="attributo …" (si traducono quegli attributi); anche la
// descrizione dei dati strutturati si traduce. Il riquadro delle marche, rigenerato
// da sorgenti/verifica/sito, viene segnato in automatico.
//
// Le traduzioni sono in lang/<lingua>.tsv: «chiave<TAB>testo», con la chiave FNV-1a a
// 32 bit del testo italiano (spazi ridotti a uno, numeri dei conteggi scritti «#»),
// come ui/lang del programma; lang/it.tsv contiene i testi italiani.
//
// Uso, dalla cartella sito:
//
//	go run . -scrivi-it   riscrive lang/it.tsv dopo aver cambiato i testi italiani
//	go run . -check       controlla traduzioni e pagina italiana senza scrivere nulla
//	go run .              scrive docs/<lingua>/index.html (10 lingue), aggiorna in
//	                      docs/index.html le parti generate e scrive docs/sitemap.xml
//
// Il flusso «Sito» (.github/workflows/sito.yml) esegue «go run .» prima di ogni
// pubblicazione: le pagine delle altre lingue non si salvano nel repository. Se manca
// una traduzione il programma si ferma con l'elenco dei testi e il sito non cambia.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	siteURL  = "https://iu0vyt.github.io/IU0VYT/"
	repoBlob = "https://github.com/IU0VYT/IU0VYT/blob/main/"
	itHeader = "# IU0VYT Flash: testi italiani del sito (chiave FNV-1a 32 bit, testo); scritto da «go run . -scrivi-it»\n"
)

// lingue nello stesso ordine del selettore del programma (sorgenti/ui/i18n.js)
type lang struct {
	Code    string // codice della lingua e cartella in docs/ e lingue/
	Name    string // nome nella lingua stessa
	OG      string // og:locale
	Intl    string // formato di date e numeri nello script della pagina
	Guide   string // guida tradotta in lingue/<codice>/
	Authors string // autori e licenze in lingue/<codice>/
	Label   string // «Lingua» nella lingua stessa (selettore)
}

var langs = []lang{
	{"it", "Italiano", "it_IT", "it-IT", "LEGGIMI.txt", "AUTORI_E_LICENZE.txt", "Lingua"},
	{"en", "English", "en_GB", "en-GB", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "Language"},
	{"fr", "Français", "fr_FR", "fr-FR", "LISEZMOI.txt", "AUTEURS_ET_LICENCES.txt", "Langue"},
	{"de", "Deutsch", "de_DE", "de-DE", "LIESMICH.txt", "AUTOREN_UND_LIZENZEN.txt", "Sprache"},
	{"es", "Español", "es_ES", "es-ES", "LEEME.txt", "AUTORES_Y_LICENCIAS.txt", "Idioma"},
	{"pt", "Português", "pt_PT", "pt-PT", "LEIAME.txt", "AUTORES_E_LICENCAS.txt", "Idioma"},
	{"ru", "Русский", "ru_RU", "ru-RU", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "Язык"},
	{"zh", "中文", "zh_CN", "zh-CN", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "语言"},
	{"ja", "日本語", "ja_JP", "ja-JP", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "言語"},
	{"ko", "한국어", "ko_KR", "ko-KR", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "언어"},
	{"ar", "العربية", "ar_AE", "ar-u-nu-latn", "GUIDE.txt", "AUTHORS_AND_LICENSES.txt", "اللغة"},
}

func (l lang) url() string {
	if l.Code == "it" {
		return siteURL
	}
	return siteURL + l.Code + "/"
}

func (l lang) rtl() bool { return l.Code == "ar" }

var errs []string

func fail(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }

func stopIfErrors() {
	if len(errs) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "sito: errori:")
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "  -", e)
	}
	os.Exit(1)
}

// ---------------------------------------------------------------- testi

var (
	reStart   = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)((?:\s+[a-zA-Z_:][-a-zA-Z0-9_:.]*(?:\s*=\s*"[^"]*")?)*)\s*/?>`)
	reAttr    = regexp.MustCompile(`([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*"([^"]*)")?`)
	reSkip    = regexp.MustCompile(`(?s)<(script|style)\b[^>]*>.*?</(?:script|style)>`)
	reTag     = regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)([^>]*)>`)
	reNum     = regexp.MustCompile(`<span data-n="([a-z]+)">\d+</span>`)
	reNumHash = regexp.MustCompile(`<span data-n="([a-z]+)">#</span>`)
	reTailNum = regexp.MustCompile(`: (\d+)$`)
	reHTML    = regexp.MustCompile(`<html[^>]*>`)
)

const (
	kInner = iota // contenuto di un elemento con data-t (HTML)
	kAttr         // valore di un attributo elencato in data-ta (testo di attributo)
	kJSON         // descrizione dei dati strutturati (testo semplice)
)

type unit struct {
	pos  int
	kind int
	key  string
	text string
}

// elemento segnato nella pagina italiana
type elem struct {
	name                 string
	tagStart, tagEnd     int // tag di apertura
	innerStart, innerEnd int // contenuto (solo con data-t)
	t                    bool
	ta                   []string
	attrs                map[string]string
}

func key(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%08x", h.Sum32())
}

func normInner(s string) string {
	return reNum.ReplaceAllString(strings.Join(strings.Fields(s), " "), `<span data-n="$1">#</span>`)
}

// testo di un attributo: «… gestite: 44» -> «… gestite: #» (il numero resta quello della pagina)
func normAttr(s string) (text, num string) {
	s = strings.Join(strings.Fields(s), " ")
	if m := reTailNum.FindStringSubmatch(s); m != nil {
		return reTailNum.ReplaceAllString(s, ": #"), m[1]
	}
	return s, ""
}

// il riquadro delle marche si rigenera da sorgenti/verifica/sito senza segni: li
// rimette qui
func autoMark(s string) string {
	s = regexp.MustCompile(`(<div class="hb-title"><svg[^>]*>.*?</svg>)<span>`).ReplaceAllString(s, `$1<span data-t>`)
	return strings.Replace(s, `<div class="hb" role="group"`, `<div class="hb" data-ta="aria-label" role="group"`, 1)
}

func findClose(s, name string, from int) int {
	depth, i := 1, from
	for {
		cc := strings.Index(s[i:], "</"+name+">")
		if cc < 0 {
			return -1
		}
		oc := -1
		for j := i; ; {
			k := strings.Index(s[j:], "<"+name)
			if k < 0 {
				break
			}
			k += j
			if c := s[k+1+len(name)]; c == ' ' || c == '>' || c == '\t' || c == '\n' || c == '/' {
				oc = k
				break
			}
			j = k + 1
		}
		if oc >= 0 && oc < i+cc {
			depth++
			i = oc + 1
			continue
		}
		depth--
		if depth == 0 {
			return i + cc
		}
		i += cc + 1
	}
}

func parse(s string) []elem {
	skip := reSkip.FindAllStringIndex(s, -1)
	inSkip := func(p int) bool {
		for _, r := range skip {
			if p >= r[0] && p < r[1] {
				return true
			}
		}
		return false
	}
	var out []elem
	for _, m := range reStart.FindAllStringSubmatchIndex(s, -1) {
		if inSkip(m[0]) {
			continue
		}
		attrs := map[string]string{}
		for _, a := range reAttr.FindAllStringSubmatch(s[m[4]:m[5]], -1) {
			attrs[a[1]] = a[2]
		}
		_, t := attrs["data-t"]
		ta := strings.Fields(attrs["data-ta"])
		if !t && len(ta) == 0 {
			continue
		}
		e := elem{name: strings.ToLower(s[m[2]:m[3]]), tagStart: m[0], tagEnd: m[1], t: t, ta: ta, attrs: attrs}
		if t {
			c := findClose(s, e.name, m[1])
			if c < 0 {
				fail("<%s data-t> senza chiusura (posizione %d)", e.name, m[0])
				continue
			}
			e.innerStart, e.innerEnd = m[1], c
		}
		for _, a := range ta {
			if _, ok := attrs[a]; !ok {
				fail("<%s data-ta=%q>: manca l'attributo %s", e.name, attrs["data-ta"], a)
			}
		}
		out = append(out, e)
	}
	// un elemento segnato non può stare dentro il contenuto di un altro
	for _, a := range out {
		if !a.t {
			continue
		}
		for _, b := range out {
			if b.tagStart > a.innerStart && b.tagStart < a.innerEnd {
				fail("elemento segnato dentro un altro: %q", s[b.tagStart:b.tagEnd])
			}
		}
	}
	return out
}

// testi da tradurre, nell'ordine della pagina
func units(s string, els []elem) []unit {
	var us []unit
	for _, e := range els {
		for _, a := range e.ta {
			t, _ := normAttr(e.attrs[a])
			us = append(us, unit{e.tagStart, kAttr, key(t), t})
		}
		if e.t {
			t := normInner(s[e.innerStart:e.innerEnd])
			us = append(us, unit{e.innerStart, kInner, key(t), t})
		}
	}
	if d, p, ok := ldDescription(s); ok {
		t := strings.Join(strings.Fields(d), " ")
		us = append(us, unit{p, kJSON, key(t), t})
	}
	sort.SliceStable(us, func(i, j int) bool { return us[i].pos < us[j].pos })
	return us
}

// tag e attributi di un testo HTML, senza ordine (i link possono cambiare indirizzo)
func sig(s string) string {
	var parts []string
	for _, m := range reTag.FindAllStringSubmatch(s, -1) {
		if m[1] == "/" {
			parts = append(parts, "/"+m[2])
			continue
		}
		var as []string
		for _, a := range reAttr.FindAllStringSubmatch(m[3], -1) {
			if a[1] == "href" {
				as = append(as, "href")
			} else {
				as = append(as, a[1]+"="+a[2])
			}
		}
		sort.Strings(as)
		parts = append(parts, m[2]+"["+strings.Join(as, " ")+"]")
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func checkTranslation(code string, u unit, tr string) {
	short := u.text
	if len([]rune(short)) > 60 {
		short = string([]rune(short)[:60]) + "…"
	}
	if strings.TrimSpace(tr) == "" {
		fail("%s: traduzione vuota per %q", code, short)
		return
	}
	switch u.kind {
	case kInner:
		if sig(tr) != sig(u.text) {
			fail("%s: tag o attributi diversi dall'italiano in %q -> %q", code, short, tr)
		}
		if len(reNumHash.FindAllString(tr, -1)) != len(reNumHash.FindAllString(u.text, -1)) {
			fail("%s: conteggi <span data-n> diversi in %q", code, short)
		}
	case kAttr:
		if strings.ContainsAny(tr, "\"<>") {
			fail("%s: virgolette o < > nel testo di un attributo: %q", code, tr)
		}
		if strings.HasSuffix(u.text, ": #") != strings.HasSuffix(tr, ": #") {
			fail("%s: «: #» finale diverso in %q -> %q", code, short, tr)
		}
	case kJSON:
		if strings.ContainsAny(tr, "<>") {
			fail("%s: < > nella descrizione dei dati strutturati", code)
		}
	}
}

// ---------------------------------------------------------------- pagine

func ldBounds(s string) (int, int, bool) {
	const open = `<script type="application/ld+json">`
	i := strings.Index(s, open)
	if i < 0 {
		return 0, 0, false
	}
	j := strings.Index(s[i:], "</script>")
	if j < 0 {
		return 0, 0, false
	}
	return i + len(open), i + j, true
}

type kv struct {
	k string
	v json.RawMessage
}

func ldFields(body string) ([]kv, error) {
	dec := json.NewDecoder(strings.NewReader(body))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, fmt.Errorf("dati strutturati: non è un oggetto JSON")
	}
	var out []kv
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		k, _ := t.(string)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		out = append(out, kv{k, raw})
	}
	return out, nil
}

func ldDescription(s string) (string, int, bool) {
	a, b, ok := ldBounds(s)
	if !ok {
		fail("dati strutturati (application/ld+json) non trovati")
		return "", 0, false
	}
	fs, err := ldFields(s[a:b])
	if err != nil {
		fail("%v", err)
		return "", 0, false
	}
	for _, f := range fs {
		if f.k == "description" {
			var d string
			if json.Unmarshal(f.v, &d) == nil {
				return d, a, true
			}
		}
	}
	fail("dati strutturati senza «description»")
	return "", 0, false
}

func js(v any) json.RawMessage {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
	return json.RawMessage(strings.TrimRight(b.String(), "\n"))
}

func plain(h string) string {
	return strings.Join(strings.Fields(html.UnescapeString(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(h, ""))), " ")
}

// dati strutturati della pagina: indirizzo, descrizione, funzioni (titoli delle
// schede) e parole chiave (marche del riquadro); gli altri campi restano come sono
func rebuildLD(s string, l lang, desc string) string {
	a, b, ok := ldBounds(s)
	if !ok {
		return s
	}
	fs, err := ldFields(s[a:b])
	if err != nil {
		fail("%s: %v", l.Code, err)
		return s
	}
	var feats []string
	for _, m := range regexp.MustCompile(`(?s)<article class="card">.*?<h3[^>]*>(.*?)</h3>`).FindAllStringSubmatch(s, -1) {
		feats = append(feats, plain(m[1]))
	}
	var brands []string
	if i := strings.Index(s, `<ul class="hb-grid">`); i >= 0 {
		if j := strings.Index(s[i:], "</ul>"); j >= 0 {
			for _, m := range regexp.MustCompile(`</span>([^<]+)</li>`).FindAllStringSubmatch(s[i:i+j], -1) {
				brands = append(brands, html.UnescapeString(m[1]))
			}
		}
	}
	var out strings.Builder
	out.WriteString("\n{\n")
	for i, f := range fs {
		v := f.v
		switch f.k {
		case "url":
			v = js(l.url())
		case "description":
			if l.Code != "it" {
				v = js(desc)
			}
		case "featureList":
			parts := make([]string, len(feats))
			for k, x := range feats {
				parts[k] = string(js(x))
			}
			v = json.RawMessage("[" + strings.Join(parts, ", ") + "]")
		case "keywords":
			v = js(strings.Join(brands, ", "))
		}
		fmt.Fprintf(&out, "  %s: %s", js(f.k), v)
		if i < len(fs)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	out.WriteString("}\n")
	return s[:a] + out.String() + s[b:]
}

func rel(from, to lang) string {
	switch {
	case from.Code == to.Code:
		return "./"
	case from.Code == "it":
		return to.Code + "/"
	case to.Code == "it":
		return "../"
	}
	return "../" + to.Code + "/"
}

func headBlock(l lang) string {
	var b strings.Builder
	b.WriteString("<!-- sito:testa: generato da sito/genera.go (indirizzo e lingue della pagina) -->\n")
	fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"%s\">\n", l.url())
	for _, x := range langs {
		fmt.Fprintf(&b, "<link rel=\"alternate\" hreflang=\"%s\" href=\"%s\">\n", x.Code, x.url())
	}
	fmt.Fprintf(&b, "<link rel=\"alternate\" hreflang=\"x-default\" href=\"%s\">\n", siteURL)
	fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"%s\">\n", l.url())
	fmt.Fprintf(&b, "<meta property=\"og:locale\" content=\"%s\">\n", l.OG)
	for _, x := range langs {
		if x.Code != l.Code {
			fmt.Fprintf(&b, "<meta property=\"og:locale:alternate\" content=\"%s\">\n", x.OG)
		}
	}
	b.WriteString("<!-- /sito:testa -->")
	return b.String()
}

func selector(cur lang, flags map[string]string) string {
	var b strings.Builder
	b.WriteString("<!-- sito:lingue: generato da sito/genera.go (bandiere di sorgenti/ui/i18n.js) -->\n")
	b.WriteString("    <details class=\"lang\">\n")
	fmt.Fprintf(&b, "      <summary title=\"%s\" aria-label=\"%s: %s\">%s<span class=\"lang-name\">%s</span><span class=\"lang-code\" aria-hidden=\"true\">%s</span></summary>\n",
		cur.Label, cur.Label, cur.Name, flags[cur.Code], cur.Name, strings.ToUpper(cur.Code))
	b.WriteString("      <ul class=\"lang-list\">\n")
	for _, x := range langs {
		extra := ""
		if x.Code == cur.Code {
			extra += ` aria-current="page"`
		}
		fmt.Fprintf(&b, "        <li><a href=\"%s\" hreflang=\"%s\" lang=\"%s\" data-lang=\"%s\"%s>%s<span>%s</span></a></li>\n",
			rel(cur, x), x.Code, x.Code, x.Code, extra, flags[x.Code], x.Name)
	}
	b.WriteString("      </ul>\n    </details>\n    <!-- /sito:lingue -->")
	return b.String()
}

func replaceBlock(s, name, block string) string {
	i := strings.Index(s, "<!-- sito:"+name)
	end := "<!-- /sito:" + name + " -->"
	j := strings.Index(s, end)
	if i < 0 || j < i {
		fail("blocco «sito:%s» non trovato in docs/index.html", name)
		return s
	}
	return s[:i] + block + s[j+len(end):]
}

func mustReplace(s *string, code, old, new string) {
	if !strings.Contains(*s, old) {
		fail("%s: %q non trovato nella pagina", code, old)
		return
	}
	*s = strings.ReplaceAll(*s, old, new)
}

func render(src string, els []elem, l lang, tr map[string]string, flags, counts map[string]string) string {
	type edit struct {
		a, b int
		s    string
	}
	var eds []edit
	desc := ""
	if l.Code != "it" {
		restore := func(t string) string {
			return reNumHash.ReplaceAllStringFunc(t, func(x string) string {
				n := reNumHash.FindStringSubmatch(x)[1]
				if counts[n] == "" {
					fail("%s: conteggio %q assente nella pagina italiana", l.Code, n)
				}
				return `<span data-n="` + n + `">` + counts[n] + `</span>`
			})
		}
		for _, e := range els {
			if len(e.ta) > 0 {
				tag := src[e.tagStart:e.tagEnd]
				for _, a := range e.ta {
					t, num := normAttr(e.attrs[a])
					v := tr[key(t)]
					if num != "" {
						v = strings.TrimSuffix(v, "#") + num
					}
					re := regexp.MustCompile(`(\s` + regexp.QuoteMeta(a) + `=")[^"]*(")`)
					tag = re.ReplaceAllLiteralString(tag, " "+a+`="`+v+`"`)
				}
				eds = append(eds, edit{e.tagStart, e.tagEnd, tag})
			}
			if e.t {
				eds = append(eds, edit{e.innerStart, e.innerEnd, restore(tr[key(normInner(src[e.innerStart:e.innerEnd]))])})
			}
		}
		if d, _, ok := ldDescription(src); ok {
			desc = tr[key(strings.Join(strings.Fields(d), " "))]
		}
	}
	sort.Slice(eds, func(i, j int) bool { return eds[i].a < eds[j].a })
	var b strings.Builder
	last := 0
	for _, e := range eds {
		b.WriteString(src[last:e.a])
		b.WriteString(e.s)
		last = e.b
	}
	b.WriteString(src[last:])
	out := b.String()

	h := `<html lang="` + l.Code + `"`
	if l.rtl() {
		h += ` dir="rtl"`
	}
	h += ` data-intl="` + l.Intl + `">`
	if loc := reHTML.FindStringIndex(out); loc != nil {
		out = out[:loc[0]] + h + out[loc[1]:]
	} else {
		fail("tag <html> non trovato")
	}
	out = replaceBlock(out, "testa", headBlock(l))
	out = replaceBlock(out, "lingue", selector(l, flags))
	if l.Code != "it" {
		mustReplace(&out, l.Code, `src="icona.png"`, `src="../icona.png"`)
		mustReplace(&out, l.Code, `href="icona.png"`, `href="../icona.png"`)
		mustReplace(&out, l.Code, `withTimeout('versione.txt'`, `withTimeout('../versione.txt'`)
		mustReplace(&out, l.Code, repoBlob+"changelog.txt", repoBlob+"lingue/"+l.Code+"/CHANGELOG.txt")
		mustReplace(&out, l.Code, repoBlob+"AUTORI_E_LICENZE.txt", repoBlob+"lingue/"+l.Code+"/"+l.Authors)
	}
	return rebuildLD(out, l, desc)
}

func sitemap(old string) string {
	date := "2026-11-08"
	if m := regexp.MustCompile(`<lastmod>([^<]*)</lastmod>`).FindStringSubmatch(old); m != nil {
		date = m[1]
	}
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<!-- scritta da sito/genera.go; data aggiornata dal flusso «Sito» -->\n")
	b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\" xmlns:xhtml=\"http://www.w3.org/1999/xhtml\">\n")
	for _, l := range langs {
		fmt.Fprintf(&b, "  <url>\n    <loc>%s</loc>\n    <lastmod>%s</lastmod>\n", l.url(), date)
		for _, x := range langs {
			fmt.Fprintf(&b, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", x.Code, x.url())
		}
		fmt.Fprintf(&b, "    <xhtml:link rel=\"alternate\" hreflang=\"x-default\" href=\"%s\"/>\n  </url>\n", siteURL)
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

// ---------------------------------------------------------------- file

// testo con fine riga «\n»; nl è il fine riga del file («\r\n» su Windows)
func readText(p string) (string, string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", "\n", err
	}
	s := strings.TrimPrefix(string(b), "\xef\xbb\xbf")
	nl := "\n"
	if strings.Contains(s, "\r\n") {
		nl = "\r\n"
	}
	return strings.ReplaceAll(s, "\r\n", "\n"), nl, nil
}

func writeText(p, s, nl string) {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		fail("%v", err)
		return
	}
	if err := os.WriteFile(p, []byte(strings.ReplaceAll(s, "\n", nl)), 0o644); err != nil {
		fail("%v", err)
	}
}

func loadTSV(p string) map[string]string {
	s, _, err := readText(p)
	if err != nil {
		fail("%v", err)
		return nil
	}
	m := map[string]string{}
	for n, line := range strings.Split(s, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "\t")
		if !ok || len(k) != 8 {
			fail("%s:%d: riga non valida (serve «chiave<TAB>testo»)", filepath.Base(p), n+1)
			continue
		}
		if _, dup := m[k]; dup {
			fail("%s:%d: chiave %s ripetuta", filepath.Base(p), n+1, k)
		}
		m[k] = v
	}
	return m
}

// bandiere del selettore del programma (funzione flag di sorgenti/ui/i18n.js)
func loadFlags(p string) map[string]string {
	s, _, err := readText(p)
	if err != nil {
		fail("bandiere: %v", err)
		return nil
	}
	m := map[string]string{}
	for _, x := range regexp.MustCompile(`(?m)^\s*(it|en|fr|de|es|pt|ru|zh|ja|ko|ar): '(<[^']*)',?\s*$`).FindAllStringSubmatch(s, -1) {
		m[x[1]] = `<svg class="flag" viewBox="0 0 24 16" width="22" height="15" aria-hidden="true">` + x[2] +
			`<rect width="24" height="16" fill="none" stroke="rgba(0,0,0,.25)" stroke-width=".8"/></svg>`
	}
	for _, l := range langs {
		if m[l.Code] == "" {
			fail("bandiera di %s non trovata in %s", l.Code, p)
		}
	}
	return m
}

func main() {
	docsFlag := flag.String("docs", "../docs", "cartella del sito")
	scriviIt := flag.Bool("scrivi-it", false, "riscrive lang/it.tsv con i testi italiani di docs/index.html")
	check := flag.Bool("check", false, "controlla soltanto, senza scrivere file")
	flag.Parse()
	docs, _ := filepath.Abs(*docsFlag)
	root := filepath.Dir(docs)
	langDir := filepath.Join(root, "sito", "lang")

	orig, nl, err := readText(filepath.Join(docs, "index.html"))
	if err != nil {
		fail("%v", err)
		stopIfErrors()
	}
	src := autoMark(orig)
	flags := loadFlags(filepath.Join(root, "sorgenti", "ui", "i18n.js"))
	els := parse(src)
	us := units(src, els)
	counts := map[string]string{}
	for _, m := range reNum.FindAllStringSubmatch(src, -1) {
		if counts[m[1]] == "" {
			counts[m[1]] = regexp.MustCompile(`\d+`).FindString(m[0])
		}
	}
	stopIfErrors()

	// testi italiani
	var it strings.Builder
	it.WriteString(itHeader)
	seen := map[string]string{}
	var uniq []unit
	for _, u := range us {
		if o, ok := seen[u.key]; ok {
			if o != u.text {
				fail("stessa chiave %s per %q e %q", u.key, o, u.text)
			}
			continue
		}
		seen[u.key] = u.text
		uniq = append(uniq, u)
		if strings.ContainsAny(u.text, "\t\n") {
			fail("testo italiano con tabulazioni: %q", u.text)
		}
		fmt.Fprintf(&it, "%s\t%s\n", u.key, u.text)
	}
	itPath := filepath.Join(langDir, "it.tsv")
	if *scriviIt && !*check {
		writeText(itPath, it.String(), "\n")
	} else if old, _, _ := readText(itPath); old != it.String() {
		fail("lang/it.tsv non aggiornato: esegui «go run . -scrivi-it» e traduci i testi nuovi")
	}

	// traduzioni
	trs := map[string]map[string]string{}
	for _, l := range langs[1:] {
		m := loadTSV(filepath.Join(langDir, l.Code+".tsv"))
		var missing []string
		for _, u := range uniq {
			tr, ok := m[u.key]
			if !ok {
				missing = append(missing, u.key+"\t"+u.text)
				continue
			}
			checkTranslation(l.Code, u, tr)
		}
		if len(missing) > 0 {
			fail("lang/%s.tsv: %d testi senza traduzione:\n      %s", l.Code, len(missing), strings.Join(missing, "\n      "))
		}
		for k := range m {
			if _, ok := seen[k]; !ok {
				fail("lang/%s.tsv: traduzione %s di un testo che non c'è più", l.Code, k)
			}
		}
		for _, f := range []string{"CHANGELOG.txt", l.Guide, l.Authors, "README.md"} {
			if _, err := os.Stat(filepath.Join(root, "lingue", l.Code, f)); err != nil {
				fail("lingue/%s/%s non trovato", l.Code, f)
			}
		}
		trs[l.Code] = m
	}
	stopIfErrors()

	pages := map[string]string{}
	for _, l := range langs {
		pages[l.Code] = render(src, els, l, trs[l.Code], flags, counts)
	}
	smOld, smNL, _ := readText(filepath.Join(docs, "sitemap.xml"))
	sm := sitemap(smOld)
	stopIfErrors()

	if *check {
		if pages["it"] != orig {
			fail("docs/index.html non aggiornato: esegui «go run .»")
		}
		if sm != smOld {
			fail("docs/sitemap.xml non aggiornata: esegui «go run .»")
		}
		stopIfErrors()
		fmt.Printf("sito: controllo superato (%d testi, %d lingue)\n", len(uniq), len(langs))
		return
	}
	if pages["it"] != orig {
		writeText(filepath.Join(docs, "index.html"), pages["it"], nl)
	}
	for _, l := range langs[1:] {
		writeText(filepath.Join(docs, l.Code, "index.html"), pages[l.Code], "\n")
	}
	if sm != smOld {
		writeText(filepath.Join(docs, "sitemap.xml"), sm, smNL)
	}
	stopIfErrors()
	fmt.Printf("sito: %d pagine scritte (%d testi)\n", len(langs), len(uniq))
}
