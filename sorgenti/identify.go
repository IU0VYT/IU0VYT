// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Identificazione della radio collegata (3.8, scheda «Identifica»).
//
// Il programma prova, uno dopo l'altro, i protocolli di LETTURA di tutte le radio
// che conosce (K67 Pro, Bajeton BJ-8300, poi i driver delle altre radio) finché
// una risponde; la radio che risponde viene letta per intero e riconosciuta dal
// contenuto della memoria, come fanno CHIRP e dmrconfig con i file.
//
// Per non provare decine di volte la stessa cosa, i modelli che all'inizio del
// collegamento inviano esattamente gli stessi byte alla stessa velocità (lo
// stesso «saluto») formano un gruppo e se ne prova uno solo: i byte si ricavano
// facendo «parlare» ogni driver con una radio finta che non risponde mai
// (identTranscript). Se la radio risponde al saluto ma la lettura non riesce
// (ad esempio la radio dice di essere un altro modello), si provano gli altri
// modelli del gruppo.
//
// Per abbreviare l'attesa quando la radio non risponde, ogni attesa della
// PRIMA risposta è limitata (identFirstWait) e il tentativo si abbandona se la
// radio non ha risposto entro identSilentMax. I byte che tornano uguali a
// quelli inviati (eco del cavo) non contano come risposta.
//
// Si inviano solo comandi di lettura: nessun dato viene scritto nella radio.

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
)

var (
	identFirstWait = 700 * time.Millisecond  // attesa massima di ogni lettura prima della prima risposta
	identSilentMax = 2500 * time.Millisecond // tentativo abbandonato se la radio tace
	identPause     = 1200 * time.Millisecond // pausa dopo un tentativo a cui la radio ha risposto
	identPortMu    sync.Mutex                // openPortFunc viene sostituita durante l'identificazione
)

// ---------------------------------------------------------------- radio finta muta

// silentPort: radio che non risponde mai; registra i byte inviati dal driver.
type silentPort struct {
	tx  *bytes.Buffer
	max int
	mu  *sync.Mutex
}

var errIdentEnough = errors.New("trascrizione completa")

func (p *silentPort) Write(b []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tx.Write(b)
	if p.tx.Len() >= p.max {
		return errIdentEnough
	}
	return nil
}
func (p *silentPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if ctx.Err() != nil {
		return errCancelled
	}
	return errTimeout
}
func (p *silentPort) Purge() error { return nil }
func (p *silentPort) SetBaud(baud int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.tx, "\x00<baud %d>\x00", baud)
	return nil
}
func (p *silentPort) Close() error { return nil }

const identTxMax = 96

var (
	identTxOnce sync.Once
	identTxKeys map[string]string // modello → «saluto»
)

const identSilentName = "\x00IDENT:"

// identTranscripts: velocità e primi byte inviati dal driver di ogni modello
// collegato via cavo a una radio che non risponde (chiave del «saluto»). Si
// calcolano una volta sola, tutti insieme; va chiamata con identPortMu.
func identTranscripts() map[string]string {
	identTxOnce.Do(func() {
		keys := map[string]string{}
		var mu sync.Mutex
		bufs := map[string]*bytes.Buffer{}
		var list []*genModel
		for _, m := range genModels {
			if m.Conn == "" {
				list = append(list, m)
				bufs[m.ID] = &bytes.Buffer{}
			}
		}
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			if !strings.HasPrefix(name, identSilentName) {
				return old(name, baud)
			}
			tx := bufs[strings.TrimPrefix(name, identSilentName)]
			mu.Lock()
			fmt.Fprintf(tx, "<%d>", baud)
			mu.Unlock()
			return &silentPort{tx: tx, max: identTxMax, mu: &mu}, nil
		}
		var wg sync.WaitGroup
		for _, m := range list {
			wg.Add(1)
			go func(m *genModel) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
				defer cancel()
				defer func() { _ = recover() }()
				_, _, _ = m.Fam.Download(ctx, m, identSilentName+m.ID, func(int, int, string) {})
			}(m)
		}
		wg.Wait()
		openPortFunc = old
		for _, m := range list {
			mu.Lock()
			b := bufs[m.ID].Bytes()
			if len(b) > identTxMax+32 {
				b = b[:identTxMax+32]
			}
			keys[m.ID] = string(b)
			mu.Unlock()
		}
		identTxKeys = keys
	})
	return identTxKeys
}

// ---------------------------------------------------------------- porta controllata

// probePort: porta vera con le attese della prima risposta abbreviate; registra
// i primi byte ricevuti (identificativo della radio).
type probePort struct {
	Port
	st *probeState
}

type probeState struct {
	mu      sync.Mutex
	tx, rx  []byte
	genuine bool // la radio ha risposto (byte diversi dall'eco di quelli inviati)
	first   []byte
}

func (s *probeState) answered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.genuine
}

func (p probePort) Write(b []byte) error {
	p.st.mu.Lock()
	if len(p.st.tx) < 4096 {
		p.st.tx = append(p.st.tx, b...)
	}
	p.st.mu.Unlock()
	return p.Port.Write(b)
}

func (p probePort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if !p.st.answered() && timeout > identFirstWait {
		timeout = identFirstWait
	}
	err := p.Port.ReadFull(ctx, b, timeout)
	if err == nil {
		s := p.st
		s.mu.Lock()
		if !s.genuine {
			s.rx = append(s.rx, b...)
			if !bytes.Contains(s.tx, s.rx) {
				s.genuine = true
			}
		}
		if len(s.first) < 64 {
			s.first = append(s.first, b[:min(len(b), 64-len(s.first))]...)
		}
		s.mu.Unlock()
	}
	return err
}

// probe: esegue fn con la porta controllata; se la radio non risponde entro
// identSilentMax il tentativo si interrompe (silent = true).
func probe(ctx context.Context, fn func(ctx context.Context) error) (st *probeState, silent bool, err error) {
	st = &probeState{}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		p, err := old(name, baud)
		if err != nil {
			return nil, err
		}
		return probePort{p, st}, nil
	}
	defer func() { openPortFunc = old }()
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := make(chan struct{})
	timedOut := false
	var tmu sync.Mutex
	silentMax := identSilentMax
	go func() {
		select {
		case <-time.After(silentMax):
			if !st.answered() {
				tmu.Lock()
				timedOut = true
				tmu.Unlock()
				cancel()
			}
		case <-stop:
		}
	}()
	func() {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("errore interno: %v", e)
			}
		}()
		err = fn(pctx)
	}()
	close(stop)
	if ctx.Err() != nil {
		return st, false, errCancelled
	}
	tmu.Lock()
	to := timedOut
	tmu.Unlock()
	if err != nil && (!st.answered() || to) {
		return st, true, err
	}
	return st, false, err
}

// ---------------------------------------------------------------- risultato

type identOther struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type identField struct {
	L string `json:"l"`
	V string `json:"v"`
}

type identResult struct {
	Found   bool         `json:"found"`
	Kind    string       `json:"kind"`  // scheda della radio (k67, baj o driver generico)
	Name    string       `json:"name"`  // nome completo
	Brand   string       `json:"brand"` // marca
	Model   string       `json:"model"`
	Type    string       `json:"type"`
	Fields  []identField `json:"fields"`
	Others  []identOther `json:"others"`  // altri modelli possibili (stesso collegamento e stessa memoria)
	Answers []string     `json:"answers"` // risposte della radio non andate a buon fine
	Tried   int          `json:"tried"`   // protocolli provati
	// Ambiguous: la radio ha la memoria di più modelli che non si distinguono
	// via cavo (stesso protocollo, nessun segno del modello nella memoria)
	Ambiguous bool   `json:"ambiguous"`
	Seconds   int    `json:"seconds"`
	Backup    string `json:"backup"`
	Count     int    `json:"count"` // canali presenti
}

func (r *identResult) add(l, f string, a ...interface{}) {
	v := strings.TrimSpace(fmt.Sprintf(f, a...))
	if v != "" {
		r.Fields = append(r.Fields, identField{l, v})
	}
}

// radio letta in attesa di essere aperta nella sua scheda
type identSession struct {
	kind  string
	kinds []string // modelli proposti (il primo è quello identificato)
	sess  *radioSession
	mems  map[string][]byte // memoria letta con il driver di ogni modello proposto
}

// ---------------------------------------------------------------- testo dei byte

// printable: testo leggibile dei byte (nil se non è testo)
func identText(b []byte) string {
	b = bytes.TrimRight(b, "\x00\xff ")
	if len(b) == 0 {
		return ""
	}
	n := 0
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			n++
		} else if c != 0 && c != 0xFF {
			return ""
		}
	}
	if n < 2 {
		return ""
	}
	var sb strings.Builder
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			sb.WriteByte(c)
		} else {
			sb.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

// identRuns: tratti di testo leggibile (almeno 4 caratteri) nei byte ricevuti
// dalla radio, es. «BT-11U V1.00» o «PROGRAMCOLORPROU»; "" se non ce ne sono
func identRuns(b []byte) string {
	var runs []string
	cur := []byte{}
	flush := func() {
		if t := strings.TrimSpace(string(cur)); len(t) >= 4 {
			runs = append(runs, strings.Join(strings.Fields(t), " "))
		}
		cur = cur[:0]
	}
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			cur = append(cur, c)
		} else {
			flush()
		}
	}
	flush()
	return strings.Join(runs, " · ")
}

func bandsText(bs []band) string {
	var out []string
	for _, b := range bs {
		out = append(out, fmt.Sprintf("%s–%s MHz", mhzText(b.Lo), mhzText(b.Hi)))
	}
	return strings.Join(out, ", ")
}

func mhzText(hz int64) string {
	s := fmt.Sprintf("%.4f", float64(hz)/1e6)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return strings.Replace(s, ".", ",", 1)
}

// ---------------------------------------------------------------- dati DMR

// BCD a 8 cifre (ID DMR della RD-5R e della DM-1801, come GET_ID di dmrconfig)
func bcdID(b []byte) (uint32, bool) {
	var v uint32
	for _, c := range b {
		hi, lo := c>>4, c&15
		if hi > 9 || lo > 9 {
			return 0, false
		}
		v = v*100 + uint32(hi)*10 + uint32(lo)
	}
	return v, true
}

func bcdDigits(b []byte) (string, bool) {
	var sb strings.Builder
	for _, c := range b {
		if c>>4 > 9 || c&15 > 9 {
			return "", false
		}
		sb.WriteByte('0' + c>>4)
		sb.WriteByte('0' + c&15)
	}
	return sb.String(), true
}

// dmrInfo: ID DMR, nome della radio, data dell'ultima programmazione e versione
// del programma che l'ha scritta, dalla memoria letta (stessi campi e stesse
// posizioni di dmrconfig: rd5r.c / gd77.c / dm1801.c print_id e print_version,
// uv380.c per la DM-1701).
func dmrInfo(m *genModel, mem []byte, r *identResult) {
	switch m.Fam.(type) {
	case rdtyFam:
		const ts, gs = 0x88, 0xE0
		if len(mem) < gs+12 {
			return
		}
		if id, ok := bcdID(mem[gs+8 : gs+12]); ok {
			r.add("ID DMR", "%d", id)
		}
		if n := mem[gs]; n != 0 && n != 0xFF {
			r.add("Nome della radio", "%s", identText(mem[gs:gs+8]))
		}
		if mem[ts] != 0xFF {
			if d, ok := bcdDigits(mem[ts : ts+6]); ok {
				r.add("Ultima programmazione", "%s-%s-%s %s:%s", d[0:4], d[4:6], d[6:8], d[8:10], d[10:12])
			}
		}
	case tytFam:
		const ts, gs = 0x2001, 0x2040
		if len(mem) < gs+0x90 {
			return
		}
		if id := uint32(mem[gs+0x44]) | uint32(mem[gs+0x45])<<8 | uint32(mem[gs+0x46])<<16; id != 0xFFFFFF {
			r.add("ID DMR", "%d", id)
		}
		if u := mem[gs+0x70 : gs+0x90]; !(u[0] == 0 && u[1] == 0) && !(u[0] == 0xFF && u[1] == 0xFF) {
			var w []uint16
			for i := 0; i+1 < len(u); i += 2 {
				c := uint16(u[i]) | uint16(u[i+1])<<8
				if c == 0 || c == 0xFFFF {
					break
				}
				w = append(w, c)
			}
			r.add("Nome della radio", "%s", strings.TrimSpace(string(utf16.Decode(w))))
		}
		if t := mem[ts : ts+11]; t[0] != 0xFF {
			if d, ok := bcdDigits(t[:7]); ok {
				r.add("Ultima programmazione", "%s-%s-%s %s:%s:%s", d[0:4], d[4:6], d[6:8], d[8:10], d[10:12], d[12:14])
			}
			const cm = "0123456789:;<=>?"
			if v := fmt.Sprintf("V%c%c.%c%c", cm[t[7]&15], cm[t[8]&15], cm[t[9]&15], cm[t[10]&15]); !strings.ContainsAny(v, ":;<=>?") {
				r.add("Versione del programma che l'ha scritta", "%s", v)
			}
		}
	}
}

// ---------------------------------------------------------------- descrizione

func genIdentFill(m *genModel, cands []*genModel, mem []byte, note string, st *probeState, port string, r *identResult) {
	r.Found, r.Kind, r.Name, r.Brand = true, m.ID, m.Name, m.Vendor
	r.Model = strings.TrimSpace(m.Model + " " + m.Variant)
	switch {
	case m.Digital != "":
		r.Type = "Ricetrasmettitore " + m.Digital + " (canali analogici FM e digitali " + m.Digital + ")"
	default:
		r.Type = "Ricetrasmettitore analogico FM"
	}
	if pmr := strings.Contains(m.Name, "PMR") || strings.Contains(m.Name, "FRS"); pmr {
		r.Type += ", uso libero " + map[bool]string{true: "PMR446", false: "FRS/GMRS"}[strings.Contains(m.Name, "PMR")]
	}
	r.add("Marca e modello", "%s", m.Name)
	r.add("Tipo", "%s", r.Type)
	bs := m.Bands
	if bf, ok := m.Fam.(interface {
		Bands(m *genModel, img []byte) []band
	}); ok {
		// gamme scritte nella memoria della radio (come CHIRP)
		func() {
			defer func() { _ = recover() }()
			if b := bf.Bands(m, mem); len(b) > 0 {
				bs = b
			}
		}()
	}
	r.add("Gamme di frequenza", "%s", bandsText(bs))
	r.add("Memorie", "%d canali", m.Slots)
	switch m.Conn {
	case "hid":
		r.add("Collegamento", "USB (dispositivo HID, senza driver)")
	case "dfu":
		r.add("Collegamento", "USB in modalità DFU (driver del costruttore o WinUSB)")
	default:
		r.add("Collegamento", "cavo di programmazione sulla porta %s, %d baud", port, m.Baud)
	}
	r.add("Firmware / identificativo", "%s", note)
	dmrInfo(m, mem, r)
	if st != nil {
		r.add("Testo inviato dalla radio", "%s", identRuns(st.first))
	}
	r.add("Numero di serie", "%s", "non disponibile: il protocollo di programmazione di questa radio non lo comunica")
	seen := map[*genModel]bool{m: true}
	for _, o := range cands {
		if !seen[o] {
			seen[o] = true
			r.Others = append(r.Others, identOther{o.ID, o.Name})
		}
	}
	sort.Slice(r.Others, func(i, j int) bool { return r.Others[i].Name < r.Others[j].Name })
	r.Count = len(packed(genChans(m, mem)))
}

// ---------------------------------------------------------------- scansione

type identReq struct {
	Port   string `json:"port"`   // porta COM o "USB"
	Brand  string `json:"brand"`  // "" tutte, Baofeng, Bajeton
	Family string `json:"family"` // "" tutte, analog, dmr
}

type identProbe struct {
	name  string // testo mostrato
	kinds []*genModel
	k67   bool
	baud  int // K67: velocità
	bt    bool
}

func identWanted(req *identReq, brand string, dmr bool) bool {
	if req.Brand != "" && !strings.EqualFold(req.Brand, brand) {
		return false
	}
	switch req.Family {
	case "analog":
		return !dmr
	case "dmr":
		return dmr
	}
	return true
}

// identPlan: tentativi nell'ordine (prima le radio con una scheda aperta)
func (a *App) identPlan(req *identReq) []identProbe {
	usb := strings.EqualFold(req.Port, "USB")
	var out []identProbe
	if !usb {
		if identWanted(req, "Baofeng", false) {
			// come il programma originale: prima la velocità usata l'ultima volta
			b1, b2 := 19200, 115200
			if a.store.Settings.K67Baud == 115200 {
				b1, b2 = b2, b1
			}
			for _, b := range []int{b1, b2} {
				out = append(out, identProbe{name: fmt.Sprintf("Baofeng K67 Pro / 5RH Pro / BF-K6 (%d baud)", b), k67: true, baud: b})
			}
		}
		if identWanted(req, "Bajeton", false) {
			out = append(out, identProbe{name: "Bajeton BJ-8300 (BT-11)", bt: true})
		}
	}
	a.mu.Lock()
	open := map[string]bool{}
	for k := range a.radios {
		open[k] = true
	}
	a.mu.Unlock()
	groups := map[string]*identProbe{}
	var order []string
	var keys map[string]string
	if !usb {
		keys = identTranscripts()
	}
	for _, m := range genModels {
		if usb != (m.Conn != "") || !identWanted(req, m.Vendor, m.Digital != "") {
			continue
		}
		key := "usb " + m.ID // via USB: un tentativo per modello (se la radio manca, l'apertura fallisce subito)
		if !usb {
			key = keys[m.ID]
		}
		g := groups[key]
		if g == nil {
			g = &identProbe{}
			groups[key] = g
			order = append(order, key)
		}
		if open[m.ID] {
			g.kinds = append([]*genModel{m}, g.kinds...)
		} else {
			g.kinds = append(g.kinds, m)
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		return open[groups[order[i]].kinds[0].ID] && !open[groups[order[j]].kinds[0].ID]
	})
	for _, k := range order {
		g := groups[k]
		g.name = g.kinds[0].Name
		if n := len(g.kinds) - 1; n > 0 {
			g.name += fmt.Sprintf(" (e altri %d modelli con lo stesso collegamento)", n)
			if n == 1 {
				g.name = g.kinds[0].Name + " (e " + g.kinds[1].Name + ")"
			}
		}
		out = append(out, *g)
	}
	return out
}

func (a *App) runIdentify(ctx context.Context, j *cableJob, req *identReq) (*identResult, error) {
	identPortMu.Lock()
	defer identPortMu.Unlock()
	start := time.Now()
	res := &identResult{Fields: []identField{}, Others: []identOther{}, Answers: []string{}}
	j.set("Preparazione dei protocolli", true)
	plan := a.identPlan(req)
	if len(plan) == 0 {
		return nil, errors.New("nessuna radio corrisponde alla ricerca")
	}
	return a.identRun(ctx, j, req, plan, start, res)
}

func (a *App) identRun(ctx context.Context, j *cableJob, req *identReq, plan []identProbe, start time.Time, res *identResult) (*identResult, error) {
	answer := func(who string, err error) {
		if err != nil && err != errCancelled {
			res.Answers = append(res.Answers, who+": "+err.Error())
			j.logf("%s ha risposto, ma: %v", who, err)
		}
	}
	for i, p := range plan {
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", i+1, len(plan)), true)
		j.prog(i, len(plan), "Provo: "+p.name)
		res.Tried++
		switch {
		case p.k67 || p.bt:
			kind := kindXLC
			creq := &cableReq{Op: "read", Radio: kindXLC, Port: req.Port, Baud: p.baud, Password: a.store.Settings.K67Pwd}
			if p.bt {
				kind, creq.Radio, creq.Baud, creq.Password = kindDAT, kindDAT, 0, ""
			}
			var file []byte
			var info interface{}
			var k6 bool
			st, silent, err := probe(ctx, func(pctx context.Context) error {
				var e error
				file, info, k6, e = a.cableRead(pctx, j, creq)
				if e == nil {
					return nil
				}
				return e
			})
			if err == errCancelled && ctx.Err() != nil {
				return nil, errCancelled
			}
			if silent {
				continue
			}
			if err != nil {
				answer(p.name, err)
				_ = sleepCtx(ctx, identPause)
				continue
			}
			a.identFillMain(kind, file, info, k6, st, req.Port, res)
			res.Seconds = int(time.Since(start).Seconds())
			if bp, err := a.saveBackup(kind, file, "identificata"); err == nil {
				res.Backup = bp
			}
			a.mu.Lock()
			a.ident = &identSession{kind: kind, kinds: []string{kind}, sess: &radioSession{Base: file, K6: k6}}
			a.mu.Unlock()
			return res, nil
		default:
			readWith := func(m *genModel) (mem []byte, note string, st *probeState, silent bool, err error) {
				j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", i+1, len(plan)), true)
				j.prog(i, len(plan), "Provo: "+m.Name)
				st, silent, err = probe(ctx, func(pctx context.Context) error {
					var e error
					mem, note, e = m.Fam.Download(pctx, m, req.Port, func(done, total int, msg string) {
						j.mu.Lock()
						j.Phase = "Lettura: " + m.Name
						j.Done, j.Total, j.Msg = done, total, msg
						j.mu.Unlock()
					})
					if e == nil && len(mem) < genMinSize(m) {
						e = fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
					}
					return e
				})
				return
			}
			for k, m := range p.kinds {
				mem, note, st, silent, err := readWith(m)
				if err == errCancelled && ctx.Err() != nil {
					return nil, errCancelled
				}
				if silent {
					if k == 0 {
						break // nessuna risposta al saluto del gruppo
					}
					continue
				}
				if err != nil {
					answer(m.Name, err)
					_ = sleepCtx(ctx, identPause)
					continue
				}
				// letta. I modelli del gruppo con lo stesso saluto possono avere
				// memorie diverse (driver o dimensioni): si rilegge la radio una
				// volta per ogni tipo di memoria del gruppo e il modello è quello
				// che riconosce la propria (come CHIRP con i file). Se nessuno la
				// riconosce, i modelli non si distinguono via cavo: si propongono
				// tutti quelli con la stessa memoria di chi ha letto.
				type shapeRead struct {
					m    *genModel
					mem  []byte
					note string
					st   *probeState
				}
				reads := []shapeRead{{m, mem, note, st}}
				seen := map[string]bool{identShape(m): true}
				for _, o := range p.kinds[k+1:] {
					if seen[identShape(o)] || len(reads) > identMaxRereads {
						continue
					}
					seen[identShape(o)] = true
					_ = sleepCtx(ctx, identPause)
					mem2, note2, st2, silent2, err2 := readWith(o)
					if err2 == errCancelled && ctx.Err() != nil {
						return nil, errCancelled
					}
					if silent2 || err2 != nil {
						j.logf("%s: %v", o.Name, err2)
						continue
					}
					reads = append(reads, shapeRead{o, mem2, note2, st2})
				}
				// confermati: il modello riconosce la memoria letta con il suo driver;
				// possibili: modelli che non hanno segni riconoscibili nella memoria
				// (come in CHIRP) letti senza errori con il loro driver
				var best *genModel
				var confirmed, possible []*genModel
				for _, r := range reads {
					for _, o := range p.kinds {
						if identShape(o) != identShape(r.m) {
							continue
						}
						if !identCanMatch(o) {
							possible = append(possible, o)
							continue
						}
						if func() (ok bool) {
							defer func() { _ = recover() }()
							return len(r.mem) >= genMinSize(o) && o.Fam.Match(o, r.mem)
						}() {
							confirmed = append(confirmed, o)
							if best == nil || o == m {
								best, mem, note, st = o, r.mem, r.note, r.st
							}
						}
					}
				}
				if len(confirmed) == 0 {
					// nessun modello riconosce la memoria: sono possibili tutti quelli
					// che l'hanno letta senza errori
					for _, r := range reads {
						for _, o := range p.kinds {
							if identShape(o) == identShape(r.m) && identCanMatch(o) {
								possible = append(possible, o)
							}
						}
					}
				}
				if best == nil && len(possible) > 0 {
					best = possible[0]
					for _, o := range possible {
						if o == m {
							best = o
						}
					}
					for _, r := range reads {
						if identShape(r.m) == identShape(best) {
							mem, note, st = r.mem, r.note, r.st
						}
					}
				}
				if best == nil {
					// nessun modello riconosce la memoria: quello che l'ha letta
					best = m
					mem, note, st = reads[0].mem, reads[0].note, reads[0].st
				}
				// memoria letta per ogni modello proposto (quella del suo tipo)
				mems := map[string][]byte{best.ID: mem}
				var cands []*genModel
				for _, o := range append(confirmed, possible...) {
					for _, r := range reads {
						if _, ok := mems[o.ID]; !ok && identShape(r.m) == identShape(o) {
							mems[o.ID] = r.mem
							cands = append(cands, o)
						}
					}
				}
				genIdentFill(best, cands, mem, note, st, req.Port, res)
				res.Ambiguous = len(confirmed) == 0 && len(res.Others) > 0
				res.Seconds = int(time.Since(start).Seconds())
				if bp, err := a.saveGenBackup(best, mem, nil, "identificata"); err == nil {
					res.Backup = bp
				}
				a.mu.Lock()
				kinds := []string{best.ID}
				for _, o := range res.Others {
					kinds = append(kinds, o.Kind)
				}
				a.ident = &identSession{kind: best.ID, kinds: kinds, sess: &radioSession{Base: mem}, mems: mems}
				a.mu.Unlock()
				return res, nil
			}
		}
	}
	res.Seconds = int(time.Since(start).Seconds())
	return res, nil
}

func famName(m *genModel) string { return fmt.Sprintf("%T", m.Fam) }

// identShape: tipo di memoria letta (driver e dimensione)
func identShape(m *genModel) string {
	n := genMinSize(m)
	switch o := m.Opt.(type) {
	case *h777Opt:
		n = o.memsize
	case *rt22Opt:
		n = o.memsize
	case *uv17Opt:
		n = o.total
	case *ga510Opt:
		n = o.memsize
	case *rt900Opt:
		n = o.memsize
	}
	return fmt.Sprintf("%s/%d", famName(m), n)
}

// identCanMatch: il modello si riconosce dal contenuto della memoria (come
// match_model di CHIRP); per gli altri la memoria non dice di che modello è.
func identCanMatch(m *genModel) bool {
	switch o := m.Opt.(type) {
	case *uv5rOpt:
		return o.match
	case *bfcOpt:
		return o.match || o.fileid5x3
	case *h777Opt:
		return o.match
	}
	switch m.Fam.(type) {
	case rdtyFam, uvb5Fam, bft1Fam, rt22Fam, uv3rFam, tytFam:
		return true
	}
	return false
}

const identMaxRereads = 4

func sleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// K67 Pro e Bajeton BJ-8300
func (a *App) identFillMain(kind string, file []byte, info interface{}, k6 bool, st *probeState, port string, r *identResult) {
	r.Found, r.Kind = true, kind
	if kind == kindXLC {
		ki, _ := info.(k67Info)
		r.Brand, r.Type = "Baofeng", "Ricetrasmettitore analogico FM bibanda"
		r.Model = "K67 Pro (5RH Pro)"
		if ki.K6 {
			r.Model = "BF-K6 (memoria della K67 Pro)"
		}
		r.Name = "Baofeng " + r.Model
		r.add("Marca e modello", "%s", r.Name)
		r.add("Tipo", "%s", r.Type)
		r.add("Modello comunicato dalla radio", "%s", ki.Model)
		r.add("Gamme di frequenza", "%s", bandsText(xlcBands(file, 0x10)))
		r.add("Codice della banda (programma originale)", "%d", ki.FreqBand)
		r.add("Memorie", "%d canali", xlcChannels)
		r.add("Collegamento", "cavo di programmazione sulla porta %s, %d baud", port, ki.Baud)
	} else {
		bi, _ := info.(btInfo)
		r.Brand, r.Type = "Bajeton", "Ricetrasmettitore analogico FM bibanda"
		r.Model = "BJ-8300"
		r.Name = "Bajeton BJ-8300"
		r.add("Marca e modello", "%s", r.Name)
		r.add("Tipo", "%s", r.Type)
		f := strings.Fields(bi.ID)
		if len(f) > 0 {
			r.add("Modello comunicato dalla radio", "%s", f[0])
		}
		if len(f) > 1 {
			r.add("Versione del firmware", "%s", strings.Join(f[1:], " "))
		}
		r.add("Memorie", "%d canali (%d zone da %d)", datZones*datPerZone, datZones, datPerZone)
		r.add("Collegamento", "cavo di programmazione sulla porta %s, 57600 baud", port)
		if v := datVer(file); v != "" {
			r.add("Formato del file", "BT-11_CPS v%s", v)
		}
	}
	r.add("Numero di serie", "%s", "non disponibile: il protocollo di programmazione di questa radio non lo comunica")
	if _, chs, err := readRadio(file, &Report{}); err == nil {
		r.Count = len(packed(chs))
	}
}

// ---------------------------------------------------------------- API

func (a *App) apiIdentifyStart(r *http.Request) (interface{}, error) {
	var req identReq
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione (o «USB» per le radio DMR collegate via USB)")
	}
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: hex.EncodeToString(b), Op: "identify", State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}}
	a.mu.Lock()
	if a.job != nil && a.job.isRunning() {
		a.mu.Unlock()
		cancel()
		return nil, errors.New("è già in corso un'operazione con la radio")
	}
	a.job = j
	a.ident = nil
	a.mu.Unlock()
	if !strings.EqualFold(req.Port, "USB") {
		st := a.store.Settings
		st.GenPort = req.Port
		_ = a.store.SaveSettings(st)
	}
	go func() {
		defer cancel()
		var res *identResult
		var err error
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("errore interno: %v", e)
				}
			}()
			res, err = a.runIdentify(ctx, j, &req)
		}()
		j.mu.Lock()
		defer j.mu.Unlock()
		switch {
		case err == errCancelled:
			j.State, j.Error = "cancelled", "Ricerca interrotta. Se la radio è rimasta in modalità programmazione, spegnila e riaccendila."
		case err != nil:
			j.State, j.Error = "error", err.Error()
		default:
			j.State, j.Result = "done", res
		}
		j.CanStop = false
	}()
	return j.snapshot(), nil
}

// identify/open: la radio identificata nella sua scheda (come dopo «Leggi dalla radio»)
func (a *App) apiIdentifyOpen(r *http.Request) (interface{}, error) {
	var req struct {
		Kind string `json:"kind"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.mu.Lock()
	id := a.ident
	a.mu.Unlock()
	if id == nil {
		return nil, errors.New("nessuna radio identificata da aprire")
	}
	kind := id.kind
	if req.Kind != "" {
		ok := false
		for _, k := range id.kinds {
			ok = ok || k == req.Kind
		}
		if !ok {
			return nil, errors.New("modello non compatibile con la radio identificata")
		}
		kind = req.Kind
	}
	if m, ok := genByID[kind]; ok {
		mem := id.sess.Base
		if b, ok := id.mems[kind]; ok {
			mem = b
		}
		if len(mem) < genMinSize(m) {
			return nil, errors.New("la memoria letta non è quella di questo modello")
		}
		a.mu.Lock()
		a.radios[kind] = &radioSession{Base: mem}
		a.mu.Unlock()
		return map[string]interface{}{"radio": genResp{Kind: m.ID, Channels: genChans(m, mem), Def: m.def(), Warnings: []string{}}}, nil
	}
	rep := &Report{}
	kind, chs, err := readRadio(id.sess.Base, rep)
	if err != nil || kind != id.kind {
		return nil, fmt.Errorf("dati letti non validi: %v", err)
	}
	a.mu.Lock()
	a.radios[id.kind] = &radioSession{Base: id.sess.Base, K6: id.sess.K6}
	a.mu.Unlock()
	resp := radioResp{Kind: kind, Channels: chs, Warnings: rep.Warnings}
	if kind == kindXLC {
		resp.Bands = xlcBands(id.sess.Base, 0x10)
	} else {
		resp.DatVer = datVer(id.sess.Base)
	}
	return map[string]interface{}{"radio": resp}, nil
}
