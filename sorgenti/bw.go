// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: bitwise.py (formato delle mappe della memoria).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Descrizione della memoria delle radio (mappe dei campi), nello stesso
// linguaggio usato da CHIRP per documentare le memorie delle radio
// ("bitwise"): così le mappe dei driver si confrontano riga per riga con
// quelle pubblicate da CHIRP, e un test verifica che la posizione di ogni
// campo sia la stessa.
//
// Sintassi supportata:
//   #seekto 0x0100;   #seek 4;
//   u8 / u16 / ul16 / u24 / ul24 / u32 / ul32 / i8 / i16 / il16 / i24 / il24 / i32 / il32
//   char / lbcd / bbcd / bit / lbit,  con array  nome[N]
//   campi di bit:   u8 a:3, b:5;   (il primo campo occupa i bit più alti)
//   struct { ... } nome[N];   struct tipo { ... };   struct tipo nome[N];
//   union { ... } nome;
// Gli indirizzi di #seekto sono assoluti anche dentro le strutture (come CHIRP).

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type bwKind int

const (
	bwPrim bwKind = iota
	bwArray
	bwStruct
)

// bwEl è un elemento della mappa: un campo, un array o una struttura.
type bwEl struct {
	Kind  bwKind
	Type  string // tipo del campo (u8, lbcd, char, ...)
	Off   int    // indirizzo del primo byte
	Size  int    // byte occupati dal tipo (per i bit: 1)
	Bits  int    // campo di bit: numero di bit (0 = campo intero)
	Shift int    // campo di bit: posizione del bit più alto + 1 (come CHIRP)
	Elems []*bwEl
	Names []string
	Kids  map[string]*bwEl
}

var bwTypeSize = map[string]int{"u8": 1, "u16": 2, "ul16": 2, "u24": 3, "ul24": 3, "u32": 4, "ul32": 4,
	"i8": 1, "i16": 2, "il16": 2, "i24": 3, "il24": 3, "i32": 4, "il32": 4, "char": 1, "lbcd": 1, "bbcd": 1, "bit": 1, "lbit": 1}

// ---------------------------------------------------------------- analisi del testo

type bwTok struct {
	s    string
	line int
}

func bwTokens(src string) []bwTok {
	var out []bwTok
	lines := strings.Split(src, "\n")
	inComment := false
	for ln, line := range lines {
		i := 0
		for i < len(line) {
			if inComment {
				if j := strings.Index(line[i:], "*/"); j >= 0 {
					inComment = false
					i += j + 2
					continue
				}
				break
			}
			c := line[i]
			switch {
			case strings.HasPrefix(line[i:], "//"):
				i = len(line)
			case strings.HasPrefix(line[i:], "/*"):
				inComment = true
				i += 2
			case c == ' ' || c == '\t' || c == '\r':
				i++
			case c == '"':
				j := strings.IndexByte(line[i+1:], '"')
				if j < 0 {
					j = len(line) - i - 2
				}
				out = append(out, bwTok{line[i : i+j+2], ln + 1})
				i += j + 2
			case strings.ContainsRune("{}[];:,#", rune(c)):
				out = append(out, bwTok{string(c), ln + 1})
				i++
			default:
				j := i
				for j < len(line) && (unicode.IsLetter(rune(line[j])) || unicode.IsDigit(rune(line[j])) || line[j] == '_') {
					j++
				}
				if j == i {
					j = i + 1
				}
				out = append(out, bwTok{line[i:j], ln + 1})
				i = j
			}
		}
	}
	return out
}

type bwParser struct {
	toks  []bwTok
	pos   int
	off   int
	types map[string][]bwTok // strutture definite con nome (testo del blocco)
}

func (p *bwParser) peek() string {
	if p.pos < len(p.toks) {
		return p.toks[p.pos].s
	}
	return ""
}

func (p *bwParser) next() string {
	s := p.peek()
	p.pos++
	return s
}

func (p *bwParser) fail(f string, a ...interface{}) {
	line := 0
	if p.pos < len(p.toks) {
		line = p.toks[p.pos].line
	} else if len(p.toks) > 0 {
		line = p.toks[len(p.toks)-1].line
	}
	panic(fmt.Errorf("mappa della memoria, riga %d: %s", line, fmt.Sprintf(f, a...)))
}

func (p *bwParser) expect(s string) {
	if t := p.next(); t != s {
		p.fail("atteso «%s», trovato «%s»", s, t)
	}
}

func bwCount(s string) (int, bool) {
	v, err := strconv.ParseInt(s, 0, 64)
	return int(v), err == nil && v > 0
}

// block analizza le istruzioni fino a "}" (o alla fine) e le aggiunge a st.
func (p *bwParser) block(st *bwEl, end string) {
	for p.pos < len(p.toks) && p.peek() != end {
		p.statement(st)
	}
	if end != "" {
		p.expect(end)
	}
}

func (st *bwEl) add(name string, e *bwEl) {
	if _, dup := st.Kids[name]; dup {
		name = fmt.Sprintf("%s_%06x", name, e.Off)
	}
	st.Kids[name] = e
	st.Names = append(st.Names, name)
}

func (p *bwParser) statement(st *bwEl) {
	t := p.next()
	switch {
	case t == "#":
		d := p.next()
		v := p.next()
		p.expect(";")
		switch d {
		case "seekto":
			n, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				p.fail("indirizzo non valido %q", v)
			}
			p.off = int(n)
		case "seek":
			n, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				p.fail("valore non valido %q", v)
			}
			p.off += int(n)
		case "printoffset":
		default:
			p.fail("direttiva sconosciuta %q", d)
		}
	case t == "struct":
		p.structStmt(st)
	case t == "union":
		p.unionStmt(st)
	case bwTypeSize[t] > 0:
		p.definition(st, t)
	default:
		p.fail("istruzione non riconosciuta %q", t)
	}
}

func (p *bwParser) declName() (string, int, bool) {
	name := p.next()
	if p.peek() == "[" {
		p.next()
		c := p.next()
		n, ok := bwCount(c)
		if !ok {
			p.fail("dimensione non valida %q", c)
		}
		p.expect("]")
		return name, n, true
	}
	return name, 1, false
}

func (p *bwParser) definition(st *bwEl, typ string) {
	size := bwTypeSize[typ]
	// campo di bit?
	if p.pos+1 < len(p.toks) && p.toks[p.pos+1].s == ":" {
		if typ == "bit" || typ == "lbit" || typ == "char" || typ == "lbcd" || typ == "bbcd" {
			p.fail("campo di bit non valido su %s", typ)
		}
		left := size * 8
		for {
			name := p.next()
			p.expect(":")
			n, ok := bwCount(p.next())
			if !ok {
				p.fail("numero di bit non valido")
			}
			if n > left {
				p.fail("campi di bit oltre la dimensione del tipo")
			}
			st.add(name, &bwEl{Kind: bwPrim, Type: typ, Off: p.off, Size: size, Bits: n, Shift: left})
			left -= n
			if p.peek() != "," {
				break
			}
			p.next()
		}
		p.expect(";")
		p.off += size
		return
	}
	name, count, isArr := p.declName()
	p.expect(";")
	if typ == "bit" || typ == "lbit" {
		if count%8 != 0 {
			p.fail("un array di bit deve avere un numero di elementi multiplo di 8")
		}
		arr := &bwEl{Kind: bwArray, Off: p.off}
		for i := 0; i < count; i++ {
			sh := 8 - i%8
			if typ == "lbit" {
				sh = i%8 + 1
			}
			arr.Elems = append(arr.Elems, &bwEl{Kind: bwPrim, Type: "u8", Off: p.off + i/8, Size: 1, Bits: 1, Shift: sh})
		}
		p.off += count / 8
		if isArr {
			st.add(name, arr)
		} else {
			st.add(name, arr.Elems[0])
		}
		return
	}
	if !isArr {
		st.add(name, &bwEl{Kind: bwPrim, Type: typ, Off: p.off, Size: size})
		p.off += size
		return
	}
	arr := &bwEl{Kind: bwArray, Type: typ, Off: p.off}
	for i := 0; i < count; i++ {
		arr.Elems = append(arr.Elems, &bwEl{Kind: bwPrim, Type: typ, Off: p.off, Size: size})
		p.off += size
	}
	st.add(name, arr)
}

// blockToks restituisce i token di un blocco { ... } (senza parentesi).
func (p *bwParser) blockToks() []bwTok {
	p.expect("{")
	start := p.pos
	depth := 1
	for p.pos < len(p.toks) {
		switch p.next() {
		case "{":
			depth++
		case "}":
			depth--
			if depth == 0 {
				return p.toks[start : p.pos-1]
			}
		}
	}
	p.fail("blocco non chiuso")
	return nil
}

func (p *bwParser) sub(toks []bwTok, st *bwEl) {
	q := &bwParser{toks: toks, off: p.off, types: p.types}
	q.block(st, "")
	p.off = q.off
}

func (p *bwParser) structStmt(st *bwEl) {
	var body []bwTok
	if p.peek() == "{" {
		body = p.blockToks()
	} else {
		tname := p.next()
		if p.peek() == "{" {
			// definizione di un tipo: struct nome { ... };
			p.types[tname] = p.blockToks()
			p.expect(";")
			return
		}
		b, ok := p.types[tname]
		if !ok {
			p.fail("struttura %q non definita", tname)
		}
		body = b
	}
	name, count, isArr := p.declName()
	p.expect(";")
	arr := &bwEl{Kind: bwArray, Off: p.off}
	for i := 0; i < count; i++ {
		el := &bwEl{Kind: bwStruct, Off: p.off, Kids: map[string]*bwEl{}}
		p.sub(body, el)
		arr.Elems = append(arr.Elems, el)
	}
	if isArr {
		st.add(name, arr)
	} else {
		st.add(name, arr.Elems[0])
	}
}

func (p *bwParser) unionStmt(st *bwEl) {
	body := p.blockToks()
	name, count, isArr := p.declName()
	p.expect(";")
	arr := &bwEl{Kind: bwArray, Off: p.off}
	for i := 0; i < count; i++ {
		start := p.off
		el := &bwEl{Kind: bwStruct, Off: start, Kids: map[string]*bwEl{}}
		// ogni membro parte dallo stesso indirizzo
		q := &bwParser{toks: body, off: start, types: p.types}
		size := -1
		for q.pos < len(q.toks) {
			q.off = start
			q.statement(el)
			if size < 0 {
				size = q.off - start
			} else if q.off-start != size {
				p.fail("membri dell'unione di dimensioni diverse")
			}
		}
		if size <= 0 {
			p.fail("unione vuota")
		}
		p.off = start + size
		arr.Elems = append(arr.Elems, el)
	}
	if isArr {
		st.add(name, arr)
	} else {
		st.add(name, arr.Elems[0])
	}
}

// bwParse analizza una mappa della memoria.
func bwParse(src string) (root *bwEl, err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
				return
			}
			panic(e)
		}
	}()
	p := &bwParser{toks: bwTokens(src), types: map[string][]bwTok{}}
	root = &bwEl{Kind: bwStruct, Kids: map[string]*bwEl{}}
	p.block(root, "")
	return root, nil
}

var bwCache sync.Map

// bwMust: mappa analizzata una sola volta (le mappe sono costanti del programma).
func bwMust(src string) *bwEl {
	if v, ok := bwCache.Load(src); ok {
		return v.(*bwEl)
	}
	r, err := bwParse(src)
	if err != nil {
		panic(err)
	}
	bwCache.Store(src, r)
	return r
}

// ---------------------------------------------------------------- accesso ai campi

// F: campo di una struttura (panic se non esiste: errore di programmazione).
func (e *bwEl) F(name string) *bwEl {
	if e == nil || e.Kids == nil {
		panic(fmt.Errorf("campo %q richiesto su un elemento che non è una struttura", name))
	}
	k, ok := e.Kids[name]
	if !ok {
		panic(fmt.Errorf("campo %q inesistente", name))
	}
	return k
}

// Has: la struttura ha il campo?
func (e *bwEl) Has(name string) bool {
	if e == nil || e.Kids == nil {
		return false
	}
	_, ok := e.Kids[name]
	return ok
}

// I: elemento i-esimo di un array.
func (e *bwEl) I(i int) *bwEl {
	if e.Kind != bwArray || i < 0 || i >= len(e.Elems) {
		panic(fmt.Errorf("indice %d fuori dall'array", i))
	}
	return e.Elems[i]
}

func (e *bwEl) Len() int { return len(e.Elems) }

// P: percorso "a.b[3].c"
func (e *bwEl) P(path string) *bwEl {
	cur := e
	for _, part := range strings.Split(path, ".") {
		idx := -1
		if j := strings.IndexByte(part, '['); j >= 0 {
			n, err := strconv.Atoi(strings.TrimSuffix(part[j+1:], "]"))
			if err != nil {
				panic(fmt.Errorf("percorso non valido %q", path))
			}
			idx = n
			part = part[:j]
		}
		cur = cur.F(part)
		if idx >= 0 {
			cur = cur.I(idx)
		}
	}
	return cur
}

// ByteLen: byte occupati dall'elemento.
func (e *bwEl) ByteLen() int {
	switch e.Kind {
	case bwPrim:
		return e.Size
	case bwArray:
		if len(e.Elems) == 0 {
			return 0
		}
		last := e.Elems[len(e.Elems)-1]
		return last.Off + last.ByteLen() - e.Off
	}
	end := e.Off
	for _, k := range e.Kids {
		if x := k.Off + k.ByteLen(); x > end {
			end = x
		}
	}
	return end - e.Off
}

func (e *bwEl) Raw(b []byte) []byte { return b[e.Off : e.Off+e.ByteLen()] }

func (e *bwEl) SetRaw(b []byte, v []byte) {
	n := e.ByteLen()
	if len(v) != n {
		panic(fmt.Errorf("dati di %d byte per un elemento di %d byte", len(v), n))
	}
	copy(b[e.Off:e.Off+n], v)
}

// Fill: tutti i byte dell'elemento al valore v.
func (e *bwEl) Fill(b []byte, v byte) {
	n := e.ByteLen()
	for i := 0; i < n; i++ {
		b[e.Off+i] = v
	}
}

func bwIsLE(t string) bool { return strings.HasPrefix(t, "ul") || strings.HasPrefix(t, "il") }

// raw value of an integer type (without bit masking)
func (e *bwEl) word(b []byte) uint64 {
	var v uint64
	if bwIsLE(e.Type) {
		for i := e.Size - 1; i >= 0; i-- {
			v = v<<8 | uint64(b[e.Off+i])
		}
	} else {
		for i := 0; i < e.Size; i++ {
			v = v<<8 | uint64(b[e.Off+i])
		}
	}
	return v
}

func (e *bwEl) putWord(b []byte, v uint64) {
	if bwIsLE(e.Type) {
		for i := 0; i < e.Size; i++ {
			b[e.Off+i] = byte(v >> (8 * i))
		}
	} else {
		for i := e.Size - 1; i >= 0; i-- {
			b[e.Off+i] = byte(v)
			v >>= 8
		}
	}
}

// Int: valore numerico. Per un array di lbcd/bbcd è il numero decimale
// rappresentato (come int() in CHIRP); per un campo di bit il valore dei bit.
func (e *bwEl) Int(b []byte) int64 {
	if e.Kind == bwArray {
		switch e.Type {
		case "lbcd":
			var v int64
			for i := len(e.Elems) - 1; i >= 0; i-- {
				c := b[e.Elems[i].Off]
				v = v*100 + int64(c>>4)*10 + int64(c&0xF)
			}
			return v
		case "bbcd":
			var v int64
			for _, x := range e.Elems {
				c := b[x.Off]
				v = v*100 + int64(c>>4)*10 + int64(c&0xF)
			}
			return v
		}
		panic(fmt.Errorf("valore numerico di un array di %s", e.Type))
	}
	if e.Kind != bwPrim {
		panic(fmt.Errorf("valore numerico di una struttura"))
	}
	switch e.Type {
	case "lbcd", "bbcd":
		c := b[e.Off]
		return int64(c>>4)*10 + int64(c&0xF)
	}
	w := e.word(b)
	if e.Bits > 0 {
		lo := uint(e.Shift - e.Bits)
		return int64((w >> lo) & (1<<uint(e.Bits) - 1))
	}
	if e.Type[0] == 'i' {
		sh := uint(64 - 8*e.Size)
		return int64(w<<sh) >> sh
	}
	return int64(w)
}

// SetInt imposta il valore (per i campi di bit lascia invariati gli altri bit).
func (e *bwEl) SetInt(b []byte, v int64) {
	if e.Kind == bwArray {
		switch e.Type {
		case "lbcd":
			for _, x := range e.Elems {
				d := v % 100
				b[x.Off] = byte(d/10)<<4 | byte(d%10)
				v /= 100
			}
			return
		case "bbcd":
			for i := len(e.Elems) - 1; i >= 0; i-- {
				d := v % 100
				b[e.Elems[i].Off] = byte(d/10)<<4 | byte(d%10)
				v /= 100
			}
			return
		}
		panic(fmt.Errorf("valore numerico in un array di %s", e.Type))
	}
	switch e.Type {
	case "lbcd", "bbcd":
		b[e.Off] = byte(v%100/10)<<4 | byte(v%10)
		return
	}
	if e.Bits > 0 {
		w := e.word(b)
		lo := uint(e.Shift - e.Bits)
		mask := uint64(1<<uint(e.Bits)-1) << lo
		w = w&^mask | (uint64(v)<<lo)&mask
		e.putWord(b, w)
		return
	}
	e.putWord(b, uint64(v))
}

func (e *bwEl) Bool(b []byte) bool { return e.Int(b) != 0 }

func (e *bwEl) SetBool(b []byte, v bool) {
	if v {
		e.SetInt(b, 1)
	} else {
		e.SetInt(b, 0)
	}
}

// Str: array di char come stringa (byte grezzi, senza conversioni).
func (e *bwEl) Str(b []byte) string { return string(e.Raw(b)) }

// SetStr: scrive la stringa riempiendo con pad fino alla lunghezza del campo.
func (e *bwEl) SetStr(b []byte, s string, pad byte) {
	n := e.ByteLen()
	for i := 0; i < n; i++ {
		if i < len(s) {
			b[e.Off+i] = s[i]
		} else {
			b[e.Off+i] = pad
		}
	}
}

// Leaves: tutti i campi elementari con il loro percorso (per i test di
// confronto con CHIRP).
func (e *bwEl) Leaves(prefix string, f func(path string, x *bwEl)) {
	switch e.Kind {
	case bwPrim:
		f(prefix, e)
	case bwArray:
		for i, x := range e.Elems {
			x.Leaves(fmt.Sprintf("%s[%d]", prefix, i), f)
		}
	case bwStruct:
		for _, n := range e.Names {
			p := n
			if prefix != "" {
				p = prefix + "." + n
			}
			e.Kids[n].Leaves(p, f)
		}
	}
}
