// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
//
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ar8200.py (4.41).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).
// Parametri della porta e comando EX dall'elenco dei comandi RS-232 di AOR per l'AR8200.

package main

// 4.41: ricevitore palmare AOR AR8200 (in CHIRP modo «live»). Il ricevitore
// acceso, col cavo CC8200, risponde a comandi di testo chiusi da CR (8 bit, 2
// bit di stop, 19200, 9600 o 4800 baud secondo il menu della radio); le
// risposte sono righe chiuse da CR LF, una riga vuota per un comando riuscito.
//
// Qui si gestiscono le 1000 memorie di scansione: dieci coppie di banchi (A/a
// … J/j) da 100 canali; quanti canali ha il banco maiuscolo di ogni coppia lo
// dice la radio (MWA → «MW A:50 a:50»). La posizione del programma è il
// numero della coppia per 100 più il numero del canale, come in CHIRP.
//
//	VR            versione («VR0104»); una riga vuota inviata risponde «?»
//	VF, MC1       un solo VFO e (MC1) come li invia CHIRP all'inizio
//	OL            tabella degli scostamenti, dieci righe per volta: «OL01 RF0000600000»
//	GR            canali della scansione selettiva: «GR00 MXA05», fino a «GRnn ---»
//	MAb, MA       dieci canali del banco b per volta: «MXA00 MP0 RF0145500000
//	              ST005000 AU0 MD1 AT0 TMnome» oppure «MXA00 ---» (vuoto)
//	MRbnn         sceglie il canale; poi OF («OF01-»), SH («SH002500+»), NL, PC
//	              danno scostamento, correzione del passo, limitatore e protezione
//	MXbnn RF… MP. ST… MD. AT. OFnn± AU. TMnome   scrive il canale
//	MRbnn, MQnn   cancella il canale
//	EX            restituisce la tastiera alla radio (CHIRP non lo invia, e la
//	              radio va spenta e riaccesa; qui si invia alla fine)
//
// Modi (MD): 0 FM larga, 1 FM, 2 AM, 3 USB, 4 LSB, 5 CW, 6 FM stretta, 7 AM
// larga, 8 AM stretta; con AU1 modo e passo li sceglie la radio dal suo piano
// di banda.
//
// L'immagine del programma (formato di IU0VYT Flash, CHIRP non ha file per
// questa radio): 256 byte di intestazione (divisione dei banchi e tabella
// degli scostamenti) e 64 byte per canale. In scrittura la radio viene
// riletta e si inviano solo i canali cambiati.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ar82Head    = 256
	ar82Rec     = 64
	ar82Magic   = "IU0VYT-AOR-AR8200\x00"
	ar82SizesAt = 32 // 10 byte: canali del banco maiuscolo di ogni coppia
	ar82OLAt    = 48 // 50 valori da 4 byte: tabella degli scostamenti in Hz
	ar82Banks   = "AaBbCcDdEeFfGgHhIiJj"
	ar82Keep    = 28 // byte del canale confrontati per decidere se scriverlo
)

// tempi di CHIRP: due secondi per ogni riga, attese fra i tentativi da 0,25 s
// raddoppiate fino a 5 s
var (
	ar82Timeout = 2 * time.Second
	ar82Wait    = 250 * time.Millisecond
	ar82WaitMax = 5 * time.Second
)

type ar82Fam struct{}

func init() {
	genRegister(&genModel{ID: "aor-ar8200", Vendor: "AOR", Model: "AR8200", RClass: "IU0VYT_AOR_AR8200", Name: "AOR AR8200",
		Slots: 1000, First: 0, NameMax: 12, Zones: 10, Powers: []genPower{{"Ricevitore", 0, 0}},
		Bands: []band{{100000, 2040000000}}, Narrow: true, HasScan: true, NoChirp: true, Baud: 19200, Fam: ar82Fam{},
		Notes: "Ricevitore: i canali sono di sola ricezione, senza toni. Si gestiscono le 1000 memorie di scansione: dieci coppie di banchi (A/a … J/j) da 100 posizioni, " +
			"divise fra i due banchi come è impostato sulla radio (se cambi la divisione, rileggi). Nomi di 12 caratteri. " +
			"Un canale nuovo, o al quale cambi la frequenza, prende il modo automatico (modo e passo scelti dalla radio), oppure FM stretta se lo chiedi; " +
			"modo, passo, attenuatore, scostamento, limitatore di rumore e protezione restano quelli impostati sulla radio finché la frequenza non cambia. " +
			"«Escluso dalla scansione» è il «pass» del canale. Frequenze di ricerca, frequenze escluse, VFO e impostazioni non si gestiscono da qui. " +
			"Radio accesa, cavo CC8200, velocità impostata sulla radio (19200, 9600 o 4800 baud, provate nell'ordine: con la radio spenta l'errore arriva dopo quasi tre minuti). " +
			"La lettura completa richiede qualche minuto; prima di scrivere la radio viene riletta e si inviano solo i canali cambiati. " +
			"Alla fine il programma restituisce la tastiera alla radio. Comandi del driver di CHIRP, provati con una radio simulata."})
}

// noIdent: «Identifica» non prova questa radio (i tentativi alle tre velocità durano minuti)
func (ar82Fam) noIdent() {}

func ar82Size(m *genModel) int { return ar82Head + m.Slots*ar82Rec }

func ar82Blank(m *genModel) []byte {
	img := make([]byte, ar82Size(m))
	copy(img, ar82Magic)
	for k := 0; k < 10; k++ {
		img[ar82SizesAt+k] = 50
	}
	return img
}

func (ar82Fam) rec(img []byte, i int) []byte { return img[ar82Head+i*ar82Rec : ar82Head+(i+1)*ar82Rec] }

func (f ar82Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}
func (f ar82Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (ar82Fam) Match(m *genModel, img []byte) bool {
	return len(img) == ar82Size(m) && bytes.HasPrefix(img, []byte(ar82Magic))
}

func (ar82Fam) imgSize(m *genModel) int { return ar82Size(m) }

// ar82Split: canali del banco maiuscolo della coppia (50 se l'immagine non lo dice)
func ar82Split(img []byte, pair int) int {
	if n := int(img[ar82SizesAt+pair]); n >= 10 && n <= 90 && n%10 == 0 {
		return n
	}
	return 50
}

// ar82Addr: banco (lettera) e numero del canale della posizione i
func ar82Addr(img []byte, i int) (bank byte, ch int) {
	pair, ch := i/100, i%100
	if n := ar82Split(img, pair); ch >= n {
		return ar82Banks[pair*2+1], ch - n
	}
	return ar82Banks[pair*2], ch
}

// ar82Chan: il canale come lo dà la radio
type ar82Chan struct {
	pass, adjOn, auto, att, limit, protect, sel bool
	freq                                        int64
	step, adj                                   int // Hz
	mode, of, dup, selIdx                       int // dup: 0 nessuno, 1 «+», 2 «-»
	name                                        string
}

// campi del canale nell'immagine: [0] presente, [1] bit, [2:6] frequenza,
// [6:9] passo, [9] modo, [10] scostamento, [11] verso, [12:15] correzione del
// passo, [16:28] nome, [28] scansione selettiva, [29] suo indice
func (f ar82Fam) get(img []byte, i int) *ar82Chan {
	r := f.rec(img, i)
	if r[0] != 1 {
		return nil
	}
	b := r[1]
	return &ar82Chan{pass: b&1 != 0, adjOn: b&2 != 0, auto: b&4 != 0, att: b&8 != 0, limit: b&16 != 0, protect: b&32 != 0,
		freq: int64(binary.BigEndian.Uint32(r[2:6])), step: int(r[6])<<16 | int(r[7])<<8 | int(r[8]), mode: int(r[9]), of: int(r[10]), dup: int(r[11]),
		adj: int(r[12])<<16 | int(r[13])<<8 | int(r[14]), name: strings.TrimRight(string(r[16:28]), " \x00"), sel: r[28] != 0, selIdx: int(r[29])}
}

func (f ar82Fam) put(img []byte, i int, c *ar82Chan) {
	r := f.rec(img, i)
	for k := range r {
		r[k] = 0
	}
	if c == nil {
		return
	}
	r[0] = 1
	for k, on := range []bool{c.pass, c.adjOn, c.auto, c.att, c.limit, c.protect} {
		if on {
			r[1] |= 1 << k
		}
	}
	binary.BigEndian.PutUint32(r[2:6], uint32(c.freq))
	r[6], r[7], r[8] = byte(c.step>>16), byte(c.step>>8), byte(c.step)
	r[9], r[10], r[11] = byte(c.mode), byte(c.of), byte(c.dup)
	r[12], r[13], r[14] = byte(c.adj>>16), byte(c.adj>>8), byte(c.adj)
	copy(r[16:28], fmt.Sprintf("%-12.12s", c.name))
	if c.sel {
		r[28], r[29] = 1, byte(c.selIdx)
	}
}

func (f ar82Fam) Get(m *genModel, img []byte, i int) *Chan {
	c := f.get(img, i)
	if c == nil || c.freq <= 0 {
		return nil
	}
	return &Chan{Rx: c.freq, Tx: c.freq, TxOff: true, Narrow: !c.auto && c.mode == 6, Scan: !c.pass, Name: c.name}
}

// fit: nelle conversioni da un ricetrasmettitore resta la frequenza di ricezione
func (ar82Fam) fit(m *genModel, c *Chan) (bool, string) {
	c.Tx, c.TxOff = c.Rx, true
	return true, ""
}

func (ar82Fam) toneOK(t, r genTone) bool { return t.Mode == "" && r.Mode == "" }
func (ar82Fam) fitBySet()                {}

func (f ar82Fam) chirpInvalid(m *genModel, c *Chan) bool {
	return c.TxTone != "" || c.RxTone != "" || (c.Tx != c.Rx && !c.TxOff) || c.Rx%50 != 0
}

func (f ar82Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if c == nil {
		f.put(img, i, nil)
		return nil
	}
	if c.TxTone != "" || c.RxTone != "" {
		return fmt.Errorf("la radio è un ricevitore: i canali non hanno toni")
	}
	if c.Tx != c.Rx && !c.TxOff {
		return fmt.Errorf("la radio è un ricevitore: i canali non hanno una frequenza di trasmissione")
	}
	if !m.inBands(c.Rx) {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio (da 0,1 a 2040 MHz)", fmtMHz(c.Rx))
	}
	if c.Rx%50 != 0 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %d Hz (multipli di 50 Hz)", c.Rx)
	}
	n := &ar82Chan{step: 5000, auto: true}
	old := f.get(img, i)
	if old != nil {
		*n = *old
		if old.freq != c.Rx {
			// frequenza cambiata: modo e passo li sceglie di nuovo la radio
			n.auto, n.mode = true, 0
		}
	}
	switch {
	case c.Narrow:
		n.auto, n.mode = false, 6
	case old != nil && old.freq == c.Rx && !old.auto && old.mode == 6:
		n.mode = 1 // da FM stretta a FM
	}
	n.freq, n.pass, n.name = c.Rx, !c.Scan, asciiName(c.Name)
	f.put(img, i, n)
	return nil
}

// ---------------------------------------------------------------- collegamento

type ar82Sess struct {
	*rawSess
	lines int // righe ricevute (per le prove)
}

var (
	ar82ReVR = regexp.MustCompile(`^VR([0-9]+)`)
	ar82ReMW = regexp.MustCompile(`^MW [A-J]:([1-9]0) [a-j]:([1-9]0)`)
	ar82ReOL = regexp.MustCompile(`^OL([0-4][0-9]) RF(0[0-9]{7}00)`)
	ar82ReGR = regexp.MustCompile(`^GR([0-9]{2}) (?:(---)$|MX([A-Ja-j])([0-9]{2}))`)
	ar82ReMX = regexp.MustCompile(`^MX([A-Ja-j])([0-9]{2}) (?:(---)$|MP([01]) RF([0-9]{10}) ST([0-9]{6})([+]?) AU([01]) MD([0-8]) AT([01]) TM(.{0,12}))`)
	// CHIRP vuole un carattere dopo il numero (spazio, + o -), ma toglie gli
	// spazi in coda alla riga: qui il segno può anche mancare
	ar82ReOF  = regexp.MustCompile(`^OF([0-4][0-9]) ?([+-]?)`)
	ar82ReSH  = regexp.MustCompile(`^SH([0-9]{6})([+]?)`)
	ar82ReNL  = regexp.MustCompile(`^NL([01])`)
	ar82RePC  = regexp.MustCompile(`^PC([01])`)
	ar82ReOK  = regexp.MustCompile(`^$`)
	ar82ReAny = regexp.MustCompile(`^[?]`)
)

// readline: una riga della radio (fino a LF; due secondi per ogni byte, come
// CHIRP), senza spazi ai lati e senza i caratteri XON/XOFF del controllo di flusso
func (s *ar82Sess) readline() string {
	var b []byte
	for len(b) < 512 {
		c := readUpTo(s.ctx, s.p, 1, ar82Timeout)
		if len(c) == 0 {
			break
		}
		if c[0] == '\n' {
			break
		}
		if c[0] != 0x11 && c[0] != 0x13 {
			b = append(b, c[0])
		}
	}
	s.lines++
	return strings.TrimSpace(string(b))
}

// reset: come CHIRP dopo una risposta sbagliata (scarta quello che c'è, CR, una riga)
func (s *ar82Sess) reset() {
	s.p.Purge()
	s.send([]byte("\r"))
	s.readline()
}

// retry: i cinque tentativi di CHIRP con le sue attese; try dà true se la risposta va bene
func (s *ar82Sess) retry(try func() bool) bool {
	wait := ar82Wait
	for k := 0; k < 5; k++ {
		if s.ctx.Err() != nil {
			return false
		}
		if try() {
			return true
		}
		if s.sleep(wait) != nil {
			return false
		}
		if wait *= 2; wait > ar82WaitMax {
			wait = ar82WaitMax
		}
		s.reset()
	}
	return false
}

// cmd: un comando con una riga di risposta; i campi della risposta o nil
func (s *ar82Sess) cmd(cmd string, re *regexp.Regexp) []string {
	var m []string
	s.retry(func() bool {
		if s.send([]byte(cmd+"\r")) != nil {
			return false
		}
		m = re.FindStringSubmatch(s.readline())
		return m != nil
	})
	return m
}

func (s *ar82Sess) must(cmd string, re *regexp.Regexp, what string) ([]string, error) {
	m := s.cmd(cmd, re)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if m == nil {
		return nil, rErr("la radio non ha risposto come atteso al comando %s (%s)", cmd, what)
	}
	return m, nil
}

func ar82Open(ctx context.Context, m *genModel, port string, prog Progress) (*ar82Sess, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	if x, ok := rs.p.(interface{ SetTwoStopBits(bool) error }); ok {
		if err := x.SetTwoStopBits(true); err != nil {
			rs.p.Close()
			return nil, err
		}
	}
	s := &ar82Sess{rawSess: rs}
	s.reset()
	for _, b := range []int{19200, 9600, 4800} {
		if ctx.Err() != nil {
			rs.p.Close()
			return nil, errCancelled
		}
		prog(0, m.Slots, fmt.Sprintf("Collegamento con la radio a %d baud…", b))
		if err := rs.p.SetBaud(b); err != nil {
			rs.p.Close()
			return nil, err
		}
		s.cmd("", ar82ReAny)
		if s.cmd("VR", ar82ReVR) != nil {
			s.cmd("VF", ar82ReOK)
			s.cmd("MC1", ar82ReOK)
			return s, nil
		}
	}
	rs.p.Close()
	if ctx.Err() != nil {
		return nil, errCancelled
	}
	return nil, genFail("Collegamento", rErr("la radio non risponde: controlla che sia accesa, il cavo e la porta"))
}

// close: restituisce la tastiera alla radio e chiude la porta
func (s *ar82Sess) close() {
	s.send([]byte("EX\r"))
	readUpTo(context.Background(), s.p, 64, 300*time.Millisecond)
	s.p.Close()
}

// page: dieci righe di un elenco (MAb / MA, OL); nil se una non è quella attesa
func (s *ar82Sess) page(cmd string, re *regexp.Regexp) [][]string {
	if s.send([]byte(cmd+"\r")) != nil {
		return nil
	}
	var out [][]string
	for k := 0; k < 10; k++ {
		m := re.FindStringSubmatch(s.readline())
		if m == nil {
			return nil
		}
		out = append(out, m)
	}
	return out
}

// head: divisione dei banchi, tabella degli scostamenti e scansione selettiva
func (s *ar82Sess) head(m *genModel, img []byte) (sel map[string]int, err error) {
	// tabella degli scostamenti: cinque pagine da dieci (come CHIRP)
	var ol [][]string
	ok := s.retry(func() bool {
		ol = nil
		for k := 0; k < 5; k++ {
			p := s.page("OL", ar82ReOL)
			if p == nil {
				return false
			}
			ol = append(ol, p...)
		}
		return true
	})
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if !ok {
		return nil, rErr("la radio non ha dato la tabella degli scostamenti (comando OL)")
	}
	for _, e := range ol {
		n, _ := strconv.Atoi(e[1])
		hz, _ := strconv.ParseInt(e[2], 10, 64)
		if n < 50 {
			binary.BigEndian.PutUint32(img[ar82OLAt+4*n:], uint32(hz))
		}
	}
	// scansione selettiva: un elenco fino alla riga «---» (o cento righe). Se
	// la radio risponde in un altro modo si prosegue senza (come CHIRP, che
	// però ripete il comando altre cinque volte)
	sel = map[string]int{}
	if s.send([]byte("GR\r")) != nil {
		return sel, nil
	}
	for k := 0; k < 100; k++ {
		g := ar82ReGR.FindStringSubmatch(s.readline())
		if g == nil || g[2] != "" {
			break
		}
		n, _ := strconv.Atoi(g[1])
		sel[g[3]+g[4]] = n
	}
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	return sel, nil
}

func ar82Line(g []string) *ar82Chan {
	if g[3] != "" {
		return nil
	}
	c := &ar82Chan{pass: g[4] == "1", adjOn: g[7] == "+", auto: g[8] == "1", att: g[10] == "1", name: strings.TrimRight(g[11], " ")}
	c.freq, _ = strconv.ParseInt(g[5], 10, 64)
	c.step, _ = strconv.Atoi(g[6])
	c.mode, _ = strconv.Atoi(g[9])
	return c
}

func (s *ar82Sess) readAll(m *genModel, prog Progress, what string) ([]byte, error) {
	f := ar82Fam{}
	img := ar82Blank(m)
	var sel map[string]int
	for pair := 0; pair < 10; pair++ {
		up := ar82Banks[pair*2]
		mw, err := s.must("MW"+string(up), ar82ReMW, "divisione dei banchi")
		if err != nil {
			return nil, genFail("Lettura", err)
		}
		n, _ := strconv.Atoi(mw[1])
		img[ar82SizesAt+pair] = byte(n)
		if pair == 0 {
			if sel, err = s.head(m, img); err != nil {
				return nil, genFail("Lettura", err)
			}
		}
		for half, size := range []int{n, 100 - n} {
			bank := ar82Banks[pair*2+half]
			base := pair * 100
			if half == 1 {
				base += n
			}
			// l'elenco del banco, dieci canali per volta
			var rows [][]string
			ok := s.retry(func() bool {
				rows = nil
				for k := 0; k < (size+9)/10; k++ {
					cmd := "MA"
					if k == 0 {
						cmd += string(bank)
					}
					p := s.page(cmd, ar82ReMX)
					if p == nil {
						return false
					}
					rows = append(rows, p...)
				}
				return true
			})
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
			if !ok {
				return nil, genFail("Lettura", rErr("la radio non ha dato l'elenco del banco %c (comando MA)", bank))
			}
			for ch := 0; ch < size; ch++ {
				g := rows[ch]
				if g[1] != string(bank) || g[2] != fmt.Sprintf("%02d", ch) {
					return nil, genFail("Lettura", rErr("la radio ha risposto con un altro canale (%s%s al posto di %c%02d)", g[1], g[2], bank, ch))
				}
				c := ar82Line(g)
				if c == nil {
					continue
				}
				if s.ctx.Err() != nil {
					return nil, errCancelled
				}
				// come CHIRP: si sceglie il canale e si chiedono scostamento,
				// correzione del passo, limitatore e protezione; quello che la
				// radio non dà resta al valore normale
				id := fmt.Sprintf("%c%02d", bank, ch)
				s.cmd("MR"+id, ar82ReOK)
				if r := s.cmd("OF", ar82ReOF); r != nil && r[2] != "" {
					// senza verso lo scostamento non conta (CHIRP lo riscrive a 00)
					c.of, _ = strconv.Atoi(r[1])
					c.dup = map[string]int{"+": 1, "-": 2}[r[2]]
				}
				if r := s.cmd("SH", ar82ReSH); r != nil && r[2] == "+" {
					c.adj, _ = strconv.Atoi(r[1])
				}
				if r := s.cmd("NL", ar82ReNL); r != nil {
					c.limit = r[1] == "1"
				}
				if r := s.cmd("PC", ar82RePC); r != nil {
					c.protect = r[1] == "1"
				}
				if k, in := sel[id]; in {
					c.sel, c.selIdx = true, k
				}
				f.put(img, base+ch, c)
				prog(base+ch, m.Slots, what)
			}
			prog(base+size, m.Slots, what)
		}
	}
	return img, nil
}

func (f ar82Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := ar82Open(ctx, m, port, prog)
	if err != nil {
		return nil, "", err
	}
	defer s.close()
	img, err := s.readAll(m, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	prog(m.Slots, m.Slots, "Lettura completata")
	return img, "", nil
}

// ar82Write: la riga che scrive il canale (come CHIRP)
func ar82Write(bank byte, ch int, c *ar82Chan) string {
	of := c.of
	if c.dup == 0 {
		of = 0 // senza verso CHIRP scrive lo scostamento 00
	}
	return fmt.Sprintf("MX%c%02d RF%010d MP%d ST%06d MD%d AT%d OF%02d%s AU%d TM%-12s", bank, ch, c.freq, b2i(c.pass), c.step, c.mode, b2i(c.att),
		of, []string{"", "+", "-"}[c.dup%3], b2i(c.auto), c.name)
}

func (f ar82Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	for i := 0; i < m.Slots; i++ {
		if c := f.get(img, i); c != nil && (c.mode > 8 || c.of > 49 || c.dup > 2 || c.step <= 0 || c.step > 999950 || c.freq <= 0 || c.adj > 999950) {
			return "", fmt.Errorf("immagine della memoria non valida (posizione %d)", i)
		}
	}
	s, err := ar82Open(ctx, m, port, prog)
	if err != nil {
		return "", err
	}
	defer s.close()
	cur, err := s.readAll(m, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		return "", err
	}
	if !bytes.Equal(cur[ar82SizesAt:ar82SizesAt+10], img[ar82SizesAt:ar82SizesAt+10]) {
		return "", rErr("la divisione dei banchi sulla radio non è più quella di quando è stata letta: rileggi la radio. Nulla è stato scritto")
	}
	if !bytes.Equal(cur[ar82OLAt:ar82OLAt+200], img[ar82OLAt:ar82OLAt+200]) {
		return "", rErr("la tabella degli scostamenti sulla radio non è più quella di quando è stata letta: rileggi la radio. Nulla è stato scritto")
	}
	var todo []int
	for i := 0; i < m.Slots; i++ {
		if !bytes.Equal(f.rec(cur, i)[:ar82Keep], f.rec(img, i)[:ar82Keep]) {
			todo = append(todo, i)
		}
	}
	for k, i := range todo {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(k, len(todo), "Scrittura…")
		bank, ch := ar82Addr(img, i)
		id := fmt.Sprintf("%c%02d", bank, ch)
		c := f.get(img, i)
		if c == nil {
			// cancellazione, come CHIRP: si sceglie il canale e lo si cancella
			if _, err := s.must("MR"+id, ar82ReOK, "scelta del canale"); err != nil {
				return "", genFail("Scrittura", err)
			}
			if _, err := s.must(fmt.Sprintf("MQ%02d", ch), ar82ReOK, "cancellazione del canale"); err != nil {
				return "", genFail("Scrittura", err)
			}
			continue
		}
		if s.cmd(ar82Write(bank, ch, c), ar82ReOK) == nil {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio ha rifiutato il canale %s (posizione %d): è protetto, o il banco è protetto? I canali già scritti restano, gli altri no", id, i)
		}
		s.cmd("MR"+id, ar82ReOK)
		if !c.auto {
			// la correzione del passo si può solo commutare: prima si chiede com'è
			toggle := ""
			if r := s.cmd("SH", ar82ReSH); r != nil && (r[2] == "+") != (c.adj > 0) {
				toggle = "+"
			}
			s.cmd(fmt.Sprintf("SH%06d%s", c.adj, toggle), ar82ReOK)
		}
		s.cmd(fmt.Sprintf("NL%d", b2i(c.limit)), ar82ReOK)
		s.cmd(fmt.Sprintf("PC%d", b2i(c.protect)), ar82ReOK)
	}
	if ctx.Err() != nil {
		return "", errCancelled
	}
	prog(len(todo), len(todo), "Scrittura completata")
	return "", nil
}
