// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lingue (4.10): estrazione dei testi dell'interfaccia (ui/*.js, ui/index.html e
// le stringhe dei file .go che non sono prove) e controllo delle traduzioni.
//   IU0VYT_STRINGHE=file  scrive l'elenco dei testi (chiave, testo, file:riga)
//   IU0VYT_LANG_IT=1      riscrive ui/lang/it.tsv (testi italiani)
// TestTraduzioni controlla che ogni testo abbia la traduzione in tutte le lingue,
// con le stesse parti variabili {0}, {1}…

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

type uiStr struct {
	text string // testo normalizzato (parti variabili come {0}, {1}…)
	pos  string // file:riga della prima volta
}

var (
	reLetters3 = regexp.MustCompile(`\p{L}{3,}`)
	reTag      = regexp.MustCompile(`(?s)<[^<>]*>`)
	reAttr     = regexp.MustCompile(`(?s)\b(title|placeholder|aria-label|alt)="([^"]*)"`)
	reIdent    = regexp.MustCompile(`^[a-z0-9_.\-/#:\[\]=*>+~^$()'"]+$`)
	reGoVerb   = regexp.MustCompile(`%[-+# 0]*(\*|\d+)?(\.(\*|\d+))?[vTtbcdoOqxXUeEfFgGsp]`)
	rePH       = regexp.MustCompile(`\{(\d+)\}`)
)

var htmlEnt = strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&nbsp;", " ", "&hellip;", "…", "&mdash;", "—", "&ndash;", "–", "&laquo;", "«", "&raquo;", "»", "&times;", "×", "&rarr;", "→", "&larr;", "←", "&#9654;", "▶")

// testo da tradurre? (almeno una parola di 3 lettere, non un nome di codice)
func uiCandidate(s string) bool {
	if !reLetters3.MatchString(s) {
		return false
	}
	if strings.Contains(s, "://") || strings.HasPrefix(s, "www.") {
		return strings.Count(s, " ") >= 2
	}
	if reIdent.MatchString(s) {
		return false // parole minuscole senza spazi: chiavi, classi, selettori
	}
	if !strings.ContainsAny(s, " ") {
		r := []rune(s)
		// una sola parola: con l'iniziale maiuscola o con lettere accentate
		hasAcc := strings.ContainsAny(s, "àèéìòùÀÈÉÌÒÙ")
		if !(unicode.IsUpper(r[0]) || hasAcc) {
			return false
		}
		// nomi di codice in CamelCase (RClass, chiavi JSON)
		up := 0
		for _, c := range r[1:] {
			if unicode.IsUpper(c) {
				up++
			}
		}
		if up > 0 && !hasAcc && strings.IndexFunc(s, func(c rune) bool { return c == '-' || c == ' ' || c == '.' }) < 0 {
			return up >= len(r)-1 // sigle tutte maiuscole sì (es. «ATTENZIONE»)
		}
	}
	return true
}

func addUI(out map[string]*uiStr, order *[]string, s, pos string) {
	// byte dei protocolli delle radio (non testo): fuori
	if !utf8.ValidString(s) || strings.IndexFunc(s, func(c rune) bool { return c < 0x20 && c != '\n' && c != '\t' && c != '\r' || c == 0xFFFD }) >= 0 {
		return
	}
	s = i18nNorm(htmlEnt.Replace(s))
	s = strings.Trim(s, " ")
	if !uiCandidate(s) {
		return
	}
	// solo parti variabili e punteggiatura: niente da tradurre
	if strings.TrimSpace(rePH.ReplaceAllString(s, "")) == "" || !reLetters3.MatchString(rePH.ReplaceAllString(s, "")) {
		return
	}
	if _, ok := out[s]; ok {
		return
	}
	out[s] = &uiStr{text: s, pos: pos}
	*order = append(*order, s)
}

// testo di un modello HTML/JS: frammenti fra i tag e attributi visibili
func addHTMLText(out map[string]*uiStr, order *[]string, s, pos string) {
	for _, m := range reAttr.FindAllStringSubmatch(s, -1) {
		addUI(out, order, renumber(m[2]), pos)
	}
	// i commenti HTML non si vedono
	s = regexp.MustCompile(`(?s)<!--.*?-->`).ReplaceAllString(s, "")
	for _, frag := range reTag.Split(s, -1) {
		addUI(out, order, renumber(frag), pos)
	}
}

// parti variabili numerate da 0 in ogni frammento
func renumber(s string) string {
	n := 0
	return regexp.MustCompile(`\x00`).ReplaceAllStringFunc(s, func(string) string {
		r := fmt.Sprintf("{%d}", n)
		n++
		return r
	})
}

type jsLit struct {
	q          byte
	text       string // con \x00 al posto di ${…}
	start, end int    // posizione nel sorgente (virgolette comprese)
}

// letterali di stringa di un file JavaScript (anche dentro ${…})
func jsLiterals(src string) []jsLit {
	var out []jsLit
	var scan func(i, end int, inTmpl bool) int
	unesc := func(s string) string {
		r := strings.NewReplacer(`\n`, "\n", `\t`, " ", `\'`, "'", `\"`, `"`, "\\`", "`", `\\`, `\`, ` `, " ", `…`, "…")
		return r.Replace(s)
	}
	regexOK := func(i int) bool {
		j := i - 1
		for j >= 0 && (src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r') {
			j--
		}
		if j < 0 {
			return true
		}
		return strings.IndexByte("(,=:[!&|?{};+-*%<>~^", src[j]) >= 0 || strings.HasSuffix(src[:j+1], "return")
	}
	scan = func(i, end int, inExpr bool) int {
		depth := 0
		for i < end {
			c := src[i]
			switch {
			case c == '/' && i+1 < end && src[i+1] == '/':
				for i < end && src[i] != '\n' {
					i++
				}
			case c == '/' && i+1 < end && src[i+1] == '*':
				k := strings.Index(src[i+2:], "*/")
				if k < 0 {
					return end
				}
				i += k + 4
			case c == '/' && regexOK(i):
				i++
				cls := false
				for i < end && src[i] != '\n' {
					if src[i] == '\\' {
						i += 2
						continue
					}
					if src[i] == '[' {
						cls = true
					} else if src[i] == ']' {
						cls = false
					} else if src[i] == '/' && !cls {
						break
					}
					i++
				}
				i++
			case c == '\'' || c == '"':
				st := i
				i++
				var b strings.Builder
				for i < end && src[i] != c && src[i] != '\n' {
					if src[i] == '\\' && i+1 < end {
						b.WriteString(src[i : i+2])
						i += 2
						continue
					}
					b.WriteByte(src[i])
					i++
				}
				i++
				out = append(out, jsLit{c, unesc(b.String()), st, i})
			case c == '`':
				st := i
				i++
				var b strings.Builder
				for i < end && src[i] != '`' {
					if src[i] == '\\' && i+1 < end {
						b.WriteString(src[i : i+2])
						i += 2
						continue
					}
					if src[i] == '$' && i+1 < end && src[i+1] == '{' {
						i = scan(i+2, end, true)
						b.WriteByte(0)
						continue
					}
					b.WriteByte(src[i])
					i++
				}
				i++
				out = append(out, jsLit{'`', unesc(b.String()), st, i})
			case inExpr && c == '{':
				depth++
				i++
			case inExpr && c == '}':
				if depth == 0 {
					return i + 1
				}
				depth--
				i++
			default:
				i++
			}
		}
		return i
	}
	scan(0, len(src), false)
	sort.Slice(out, func(a, b int) bool { return out[a].start < out[b].start })
	return out
}

var (
	reJSJoin  = regexp.MustCompile(`^\s*\+\s*[A-Za-z_$][\w$.]*(\([^()'"\x60]*\))?(\.[\w$]+(\([^()'"\x60]*\))?)*\s*\+\s*$`)
	reJSTail  = regexp.MustCompile(`^\s*\+\s*[A-Za-z_$][\w$.]*(\([^()'"\x60]*\))?(\.[\w$]+(\([^()'"\x60]*\))?)*\s*[),;:\]}]`)
	reJSHead  = regexp.MustCompile(`[(,=:?]\s*[A-Za-z_$][\w$.]*(\([^()'"\x60]*\))?(\.[\w$]+(\([^()'"\x60]*\))?)*\s*\+\s*$`)
	reLineCnt = func(src string, off int) int { return strings.Count(src[:off], "\n") + 1 }
)

func extractJS(out map[string]*uiStr, order *[]string, file, src string) {
	lits := jsLiterals(src)
	for i, l := range lits {
		pos := fmt.Sprintf("%s:%d", file, reLineCnt(src, l.start))
		t := l.text
		if strings.ContainsAny(t, "<>") || l.q == '`' {
			addHTMLText(out, order, t, pos)
		} else {
			addUI(out, order, t, pos)
		}
		if l.q == '`' || strings.ContainsAny(t, "<>") {
			continue
		}
		// testi composti con +: 'Leggi ' + n + ' canali' -> «Leggi {0} canali»
		joined := t
		k := i
		for k+1 < len(lits) && lits[k+1].q != '`' && lits[k].end <= lits[k+1].start && reJSJoin.MatchString(src[lits[k].end:lits[k+1].start]) {
			joined += "\x00" + lits[k+1].text
			k++
		}
		if k == len(lits)-1 || !reJSJoin.MatchString(src[lits[k].end:min(len(src), lits[k].end+200)]) {
			if reJSTail.MatchString(src[lits[k].end:min(len(src), lits[k].end+200)]) {
				joined += "\x00"
			}
		}
		if i > 0 && reJSHead.MatchString(src[max(0, l.start-120):l.start]) && (lits[i-1].end > l.start || !reJSJoin.MatchString(src[lits[i-1].end:l.start])) {
			joined = "\x00" + joined
		}
		if joined != t {
			addUI(out, order, renumber(joined), pos)
		}
	}
}

func extractHTML(out map[string]*uiStr, order *[]string, file, src string) {
	src = regexp.MustCompile(`(?s)<script.*?</script>|<style.*?</style>`).ReplaceAllString(src, "")
	addHTMLText(out, order, src, file)
}

// stringhe dei file .go (non di prova): i formati di fmt diventano {0}, {1}…
func extractGo(t testing.TB, out map[string]*uiStr, order *[]string, file string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	skip := map[*ast.BasicLit]bool{}
	joinedVal := map[*ast.BasicLit]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			skip[x.Path] = true
		case *ast.Field:
			if x.Tag != nil {
				skip[x.Tag] = true
			}
		case *ast.CallExpr:
			// espressioni regolari, chiavi di mappe, nomi di file: non sono testi
			if se, ok := x.Fun.(*ast.SelectorExpr); ok {
				if id, ok := se.X.(*ast.Ident); ok && (id.Name == "regexp" || id.Name == "filepath" || id.Name == "os" || id.Name == "exec" || id.Name == "syscall" || id.Name == "strings" && se.Sel.Name != "Join") {
					for _, a := range x.Args {
						if bl, ok := a.(*ast.BasicLit); ok {
							skip[bl] = true
						}
					}
				}
			}
		}
		return true
	})
	// 4.10: testi composti con +: "a" + "b" diventa «ab», "Firmware: " + nome
	// diventa «Firmware: {0}» (come li mostra il programma)
	joined := map[*ast.BasicLit]string{}
	var flat func(e ast.Expr) ([]ast.Expr, bool)
	flat = func(e ast.Expr) ([]ast.Expr, bool) {
		if p, ok := e.(*ast.ParenExpr); ok {
			return flat(p.X)
		}
		if b, ok := e.(*ast.BinaryExpr); ok && b.Op == token.ADD {
			l, okl := flat(b.X)
			r, okr := flat(b.Y)
			return append(l, r...), okl || okr
		}
		bl, ok := e.(*ast.BasicLit)
		return []ast.Expr{e}, ok && bl.Kind == token.STRING
	}
	ast.Inspect(f, func(n ast.Node) bool {
		b, ok := n.(*ast.BinaryExpr)
		if !ok || b.Op != token.ADD {
			return true
		}
		parts, any := flat(b)
		if !any || len(parts) < 2 {
			return true
		}
		var sb strings.Builder
		var first *ast.BasicLit
		lits := 0
		for _, p := range parts {
			if bl, ok := p.(*ast.BasicLit); ok && bl.Kind == token.STRING {
				v, err := strconv.Unquote(bl.Value)
				if err != nil {
					return true
				}
				sb.WriteString(v)
				skip[bl] = true
				if first == nil {
					first = bl
				}
				lits++
			} else {
				sb.WriteString("\x00")
			}
		}
		if first != nil {
			joined[first] = sb.String()
		}
		return false // le parti interne sono già nel testo composto
	})
	for bl, v := range joined {
		skip[bl] = false
		joinedVal[bl] = v
	}
	ast.Inspect(f, func(n ast.Node) bool {
		bl, ok := n.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING || skip[bl] {
			return true
		}
		s, err := strconv.Unquote(bl.Value)
		if err != nil {
			return true
		}
		if v, ok := joinedVal[bl]; ok {
			s = v
		}
		// formati della memoria (CHIRP), HTML e CSS non sono testi a sé
		if strings.Contains(s, "#seekto") || strings.Contains(s, "u8 ") || strings.Contains(s, "ul16") || strings.Contains(s, "lbcd") {
			return true
		}
		pos := fmt.Sprintf("%s:%d", file, fset.Position(bl.Pos()).Line)
		if strings.Contains(s, "%") || strings.Contains(s, "\x00") {
			s = strings.ReplaceAll(s, "\x00", "%v")
			n := 0
			s = reGoVerb.ReplaceAllStringFunc(s, func(string) string {
				r := fmt.Sprintf("{%d}", n)
				n++
				return r
			})
			s = strings.ReplaceAll(s, "%%", "%")
		}
		if strings.ContainsAny(s, "<>") && reTag.MatchString(s) {
			addHTMLText(out, order, s, pos)
			return true
		}
		// i messaggi di Windows vanno a capo: si traducono riga per riga
		for _, line := range strings.Split(s, "\n") {
			addUI(out, order, renumberLocal(line), pos)
		}
		return true
	})
}

// 4.11: la variante fra parentesi in fondo al nome di una radio («… (firmware fino
// a 0.23: 960 canali)») si usa anche da sola nei nomi delle schede: è un testo a sé
var reNameVariant = regexp.MustCompile(`^.*\S \(([^()]*\pL\pL[^()]*)\)$`)

// {3} {4} in una riga -> {0} {1}
func renumberLocal(s string) string {
	m := map[string]string{}
	n := 0
	return rePH.ReplaceAllStringFunc(s, func(x string) string {
		if r, ok := m[x]; ok {
			return r
		}
		r := fmt.Sprintf("{%d}", n)
		n++
		m[x] = r
		return r
	})
}

func extractUI(t testing.TB) []*uiStr {
	out := map[string]*uiStr{}
	var order []string
	files, _ := filepath.Glob("ui/*.js")
	sort.Strings(files)
	for _, f := range files {
		if filepath.ToSlash(f) == "ui/i18n.js" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		extractJS(out, &order, f, string(b))
	}
	b, err := os.ReadFile("ui/index.html")
	if err != nil {
		t.Fatal(err)
	}
	extractHTML(out, &order, "ui/index.html", string(b))
	gofiles, _ := filepath.Glob("*.go")
	sort.Strings(gofiles)
	for _, f := range gofiles {
		if strings.HasSuffix(f, "_test.go") || strings.HasPrefix(f, "genemu") {
			continue
		}
		extractGo(t, out, &order, f)
	}
	ids := make([]string, 0, len(genByID))
	for id := range genByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if m := reNameVariant.FindStringSubmatch(genByID[id].Name); m != nil {
			addUI(out, &order, m[1], "nome della radio "+id)
		}
	}
	// parti dei nomi delle impostazioni inserite nelle etichette ($1, $2…): anche parole singole minuscole
	phr := make([]string, 0, len(itPhrases))
	for _, v := range itPhrases {
		phr = append(phr, v)
	}
	sort.Strings(phr)
	for _, v := range phr {
		if _, ok := out[v]; !ok && reLetters3.MatchString(v) && v != strings.ToUpper(v) {
			out[v] = &uiStr{text: v, pos: "settings_it.go:itPhrases"}
			order = append(order, v)
		}
	}
	res := make([]*uiStr, 0, len(order))
	for _, s := range order {
		res = append(res, out[s])
	}
	return res
}

func phSet(s string) string {
	m := rePH.FindAllString(s, -1)
	sort.Strings(m)
	return strings.Join(m, "")
}

func TestTraduzioni(t *testing.T) {
	strs := extractUI(t)
	keys := map[string]string{}
	for _, s := range strs {
		k := i18nKey(s.text)
		if o, ok := keys[k]; ok && o != s.text {
			t.Fatalf("stessa chiave %s per %q e %q", k, o, s.text)
		}
		keys[k] = s.text
	}
	if p := os.Getenv("IU0VYT_STRINGHE"); p != "" {
		var b strings.Builder
		for _, s := range strs {
			fmt.Fprintf(&b, "%s\t%s\t%s\n", i18nKey(s.text), s.text, s.pos)
		}
		if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if os.Getenv("IU0VYT_LANG_IT") != "" {
		var b strings.Builder
		b.WriteString("# IU0VYT Flash: testi italiani dell'interfaccia (chiave FNV-1a 32 bit, testo)\n")
		for _, s := range strs {
			fmt.Fprintf(&b, "%s\t%s\n", i18nKey(s.text), s.text)
		}
		if err := os.WriteFile("ui/lang/it.tsv", []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	it, err := os.ReadFile("ui/lang/it.tsv")
	if err != nil {
		t.Fatal(err)
	}
	itm := parseLangTSV(it)
	for k, s := range keys {
		if itm[k] != s {
			t.Fatalf("ui/lang/it.tsv non aggiornato (IU0VYT_LANG_IT=1): %q", s)
		}
	}
	t.Logf("%d testi dell'interfaccia", len(strs))
	for _, l := range uiLangs[1:] {
		b, err := os.ReadFile("ui/lang/" + l + ".tsv")
		if err != nil {
			t.Errorf("%s: %v", l, err)
			continue
		}
		d := parseLangTSV(b)
		miss := 0
		var ex []string
		for k, s := range keys {
			tr, ok := d[k]
			if !ok {
				miss++
				if len(ex) < 5 {
					ex = append(ex, s)
				}
				continue
			}
			if phSet(tr) != phSet(s) {
				t.Errorf("%s: parti variabili diverse in %q -> %q", l, s, tr)
			}
		}
		extra := 0
		for k := range d {
			if _, ok := keys[k]; !ok {
				extra++
			}
		}
		if miss > 0 {
			t.Errorf("%s: %d testi senza traduzione (es. %q)", l, miss, ex)
		}
		if extra > 0 {
			t.Errorf("%s: %d traduzioni di testi che non ci sono più", l, extra)
		}
	}
}
