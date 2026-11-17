//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: AOR AR8200 simulato (vedi drv_ar8200.go): i comandi che servono alle
// memorie di scansione, con le risposte nel formato che il driver di CHIRP si
// aspetta.

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type ar82Emu struct {
	mu      sync.Mutex
	m       *genModel
	sizes   [10]int
	ol      [50]uint32
	ch      map[string]*ar82Chan // «A05» → canale (adj: valore anche se la correzione è spenta)
	gr      []string             // scansione selettiva, in ordine
	cur     string               // canale scelto con MR
	maBank  byte
	maNext  int
	olNext  int
	cmds    []string // comandi ricevuti
	writes  int
	deletes int
	ex      int
	baud    int        // velocità impostata sulla radio (0 = qualunque)
	pcBaud  func() int // velocità della porta del PC
	ofSpace bool       // senza verso risponde «OF00 » invece di «OF00»
	xon     bool       // alla scrittura risponde con XOFF e XON
	refuse  string     // canale che la radio non si lascia scrivere
	grBare  bool       // elenco vuoto della scansione selettiva senza numero
	garble  int        // la pagina numero garble di un elenco MA arriva rovinata (una volta)
	pages   int
	noStop2 bool // non controlla i due bit di stop
}

var ar82ReWrite = regexp.MustCompile(`^MX([A-Ja-j])([0-9]{2}) RF([0-9]{10}) MP([01]) ST([0-9]{6}) MD([0-8]) AT([01]) OF([0-9]{2})([+-]?) AU([01]) TM(.{0,12})$`)

func (ar82Fam) blankImg(m *genModel) []byte { return ar82Blank(m) }

func (ar82Fam) emu(m *genModel, img []byte) genEmu {
	f := ar82Fam{}
	e := &ar82Emu{m: m, ch: map[string]*ar82Chan{}}
	for k := range e.sizes {
		e.sizes[k] = 50
	}
	for n, hz := range []uint32{0, 600000, 1600000, 7600000, 5000000, 4600000, 9000000, 10000000} {
		e.ol[n] = hz
	}
	if img != nil && f.Match(m, img) {
		for k := range e.sizes {
			e.sizes[k] = ar82Split(img, k)
		}
		for n := 0; n < 50; n++ {
			e.ol[n] = uint32(img[ar82OLAt+4*n])<<24 | uint32(img[ar82OLAt+4*n+1])<<16 | uint32(img[ar82OLAt+4*n+2])<<8 | uint32(img[ar82OLAt+4*n+3])
		}
		type se struct {
			id  string
			idx int
		}
		var sel []se
		for i := 0; i < m.Slots; i++ {
			if c := f.get(img, i); c != nil {
				bank, ch := ar82Addr(img, i)
				id := fmt.Sprintf("%c%02d", bank, ch)
				e.ch[id] = c
				if c.sel {
					sel = append(sel, se{id, c.selIdx})
				}
			}
		}
		sort.Slice(sel, func(a, b int) bool { return sel[a].idx < sel[b].idx })
		for _, x := range sel {
			e.gr = append(e.gr, x.id)
		}
	}
	return e
}

func (e *ar82Emu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	f := ar82Fam{}
	img := ar82Blank(e.m)
	for k, n := range e.sizes {
		img[ar82SizesAt+k] = byte(n)
	}
	for n, hz := range e.ol {
		img[ar82OLAt+4*n], img[ar82OLAt+4*n+1], img[ar82OLAt+4*n+2], img[ar82OLAt+4*n+3] = byte(hz>>24), byte(hz>>16), byte(hz>>8), byte(hz)
	}
	for i := 0; i < e.m.Slots; i++ {
		bank, ch := ar82Addr(img, i)
		id := fmt.Sprintf("%c%02d", bank, ch)
		c := e.ch[id]
		if c == nil {
			continue
		}
		x := *c
		if !x.adjOn {
			x.adj = 0
		}
		if x.dup == 0 {
			x.of = 0
		}
		x.sel, x.selIdx = false, 0
		for k, g := range e.gr {
			if g == id {
				x.sel, x.selIdx = true, k
			}
		}
		f.put(img, i, &x)
	}
	return img
}

func (e *ar82Emu) size(bank byte) int {
	k := strings.IndexByte(ar82Banks, bank)
	if k < 0 {
		return 0
	}
	if k%2 == 0 {
		return e.sizes[k/2]
	}
	return 100 - e.sizes[k/2]
}

func (e *ar82Emu) line(bank byte, n int) string {
	id := fmt.Sprintf("%c%02d", bank, n)
	c := e.ch[id]
	if c == nil {
		return "MX" + id + " ---"
	}
	adj := ""
	if c.adjOn {
		adj = "+"
	}
	return fmt.Sprintf("MX%s MP%d RF%010d ST%06d%s AU%d MD%d AT%d TM%-12s", id, b2i(c.pass), c.freq, c.step, adj, b2i(c.auto), c.mode, b2i(c.att), c.name)
}

func (e *ar82Emu) serve(p Port) error {
	if !e.noStop2 {
		emuRequireLine(p, false, true)
	}
	for {
		var cmd []byte
		for {
			b, err := emuRead(p, 1)
			if err != nil {
				return nil
			}
			if b[0] == '\r' {
				break
			}
			cmd = append(cmd, b[0])
		}
		e.mu.Lock()
		if e.baud != 0 && e.pcBaud != nil && e.pcBaud() != e.baud {
			e.mu.Unlock()
			continue // velocità sbagliata: la radio non capisce niente
		}
		out := e.handle(string(cmd))
		e.mu.Unlock()
		for _, l := range out {
			p.Write([]byte(l + "\r\n"))
		}
	}
}

func (e *ar82Emu) handle(cmd string) []string {
	e.cmds = append(e.cmds, cmd)
	c := e.ch[e.cur]
	num := func(s string) int { n, _ := strconv.Atoi(s); return n }
	switch {
	case cmd == "":
		return []string{"?"}
	case cmd == "VR":
		return []string{"VR0104"}
	case cmd == "VF", cmd == "MC1":
		return []string{""}
	case cmd == "EX":
		e.ex++
		return []string{""}
	case len(cmd) == 3 && cmd[:2] == "MW" && cmd[2] >= 'A' && cmd[2] <= 'J':
		k := int(cmd[2] - 'A')
		return []string{fmt.Sprintf("MW %c:%d %c:%d", cmd[2], e.sizes[k], cmd[2]+32, 100-e.sizes[k])}
	case cmd == "OL":
		var out []string
		for k := 0; k < 10; k++ {
			n := (e.olNext + k) % 50
			out = append(out, fmt.Sprintf("OL%02d RF%010d", n, e.ol[n]))
		}
		e.olNext = (e.olNext + 10) % 50
		return out
	case cmd == "GR":
		var out []string
		for k, id := range e.gr {
			out = append(out, fmt.Sprintf("GR%02d MX%s", k, id))
		}
		if len(e.gr) == 0 && e.grBare {
			return []string{"GR ---"}
		}
		if len(e.gr) < 100 {
			out = append(out, fmt.Sprintf("GR%02d ---", len(e.gr)))
		}
		return out
	case strings.HasPrefix(cmd, "MA") && len(cmd) <= 3:
		if len(cmd) == 3 {
			if e.size(cmd[2]) == 0 {
				return []string{"?"}
			}
			e.maBank, e.maNext = cmd[2], 0
		}
		if e.maBank == 0 || e.maNext >= e.size(e.maBank) {
			return []string{"?"}
		}
		var out []string
		for k := 0; k < 10; k++ {
			out = append(out, e.line(e.maBank, e.maNext+k))
		}
		e.maNext += 10
		if e.pages++; e.garble > 0 && e.pages == e.garble {
			out[3] = "MX??? rumore sulla linea"
		}
		return out
	case len(cmd) == 5 && cmd[:2] == "MR":
		if e.size(cmd[2]) == 0 || num(cmd[3:]) >= e.size(cmd[2]) {
			return []string{"?"}
		}
		e.cur = cmd[2:]
		return []string{""}
	case cmd == "OF":
		if c == nil {
			return []string{"?"}
		}
		sign := []string{"", "+", "-"}[c.dup]
		if c.dup == 0 && e.ofSpace {
			sign = " "
		}
		return []string{fmt.Sprintf("OF%02d%s", c.of, sign)}
	case cmd == "SH":
		if c == nil {
			return []string{"?"}
		}
		if c.adjOn {
			return []string{fmt.Sprintf("SH%06d+", c.adj)}
		}
		return []string{fmt.Sprintf("SH%06d", c.adj)}
	case cmd == "NL":
		if c == nil {
			return []string{"?"}
		}
		return []string{fmt.Sprintf("NL%d", b2i(c.limit))}
	case cmd == "PC":
		if c == nil {
			return []string{"?"}
		}
		return []string{fmt.Sprintf("PC%d", b2i(c.protect))}
	case regexp.MustCompile(`^SH[0-9]{6}[+]?$`).MatchString(cmd):
		if c == nil {
			return []string{"?"}
		}
		c.adj = num(cmd[2:8])
		if strings.HasSuffix(cmd, "+") {
			c.adjOn = !c.adjOn
		}
		return []string{""}
	case cmd == "NL0", cmd == "NL1":
		if c == nil {
			return []string{"?"}
		}
		c.limit = cmd == "NL1"
		return []string{""}
	case cmd == "PC0", cmd == "PC1":
		if c == nil {
			return []string{"?"}
		}
		c.protect = cmd == "PC1"
		return []string{""}
	case len(cmd) == 4 && cmd[:2] == "MQ":
		if len(e.cur) != 3 {
			return []string{"?"}
		}
		id := fmt.Sprintf("%c%s", e.cur[0], cmd[2:])
		if e.ch[id] == nil {
			return []string{"?"}
		}
		delete(e.ch, id)
		for k, g := range e.gr {
			if g == id {
				e.gr = append(e.gr[:k], e.gr[k+1:]...)
				break
			}
		}
		e.deletes++
		return []string{""}
	case strings.HasPrefix(cmd, "MX"):
		g := ar82ReWrite.FindStringSubmatch(cmd)
		if g == nil {
			return []string{"?"}
		}
		id := g[1] + g[2]
		hz, _ := strconv.ParseInt(g[3], 10, 64)
		if num(g[2]) >= e.size(g[1][0]) || hz < 100000 || hz > 2040000000 || num(g[5]) == 0 || num(g[8]) > 49 || id == e.refuse {
			return []string{"?"}
		}
		n := &ar82Chan{}
		if old := e.ch[id]; old != nil {
			n.adj, n.adjOn, n.limit, n.protect = old.adj, old.adjOn, old.limit, old.protect
		}
		n.freq, n.pass, n.step, n.mode, n.att = hz, g[4] == "1", num(g[5]), num(g[6]), g[7] == "1"
		n.of, n.dup, n.auto, n.name = num(g[8]), map[string]int{"": 0, "+": 1, "-": 2}[g[9]], g[10] == "1", strings.TrimRight(g[11], " ")
		e.ch[id] = n
		e.writes++
		if e.xon {
			return []string{"\x13\x11"}
		}
		return []string{""}
	}
	return []string{"?"}
}
