// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Formati universali e convertitore (3.7): i canali di qualsiasi file gestito dal
// programma (file .xlc del K67, .dat della Bajeton v1.0.8 / v1.0.9, immagini .img
// di tutte le radio compatibili, anche salvate con CHIRP, file CSV di CHIRP, tabelle
// Excel / CSV esportate da qui o scritte a mano) si leggono nello stesso modello di
// canale (Chan) e si scrivono nel formato di qualsiasi altra radio, adattati alle
// sue regole (adaptChan / genAdapt: gamme, toni, potenze, nomi, larghezza, scansione).
//
// Il formato CSV di CHIRP è quello universale: lo leggono e lo scrivono CHIRP e
// questo programma per ogni radio. Per creare l'immagine .img di una radio gestita
// con i driver serve sempre un'immagine di quella radio (letta via cavo o aperta da
// file): l'immagine contiene anche le impostazioni e la taratura della radio, che
// non si possono inventare (scrivere in una radio un'immagine «vuota» la
// danneggerebbe). I file .xlc e .dat si creano invece anche dai modelli inclusi.

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// uniSource: canali letti da un file qualsiasi
type uniSource struct {
	Kind     string   `json:"kind"` // k67, baj, id del modello, "chirp" (CSV di CHIRP), "table" (tabella)
	Desc     string   `json:"desc"` // descrizione del file
	Channels []*Chan  `json:"channels"`
	Warnings []string `json:"warnings"`
}

func (s *uniSource) count() int {
	n := 0
	for _, c := range s.Channels {
		if c != nil {
			n++
		}
	}
	return n
}

// readAny riconosce il file e ne legge i canali (nella posizione originale).
// prefer: radio da preferire per le immagini .img ambigue.
func readAny(b []byte, name, prefer string) (*uniSource, error) {
	if len(b) == 0 {
		return nil, errors.New("file vuoto")
	}
	rep := &Report{}
	switch detect(b) {
	case kindXLC, kindDAT:
		kind, chs, err := readRadio(b, rep)
		if err != nil {
			return nil, err
		}
		desc := radioName(kind) + " (file " + map[string]string{kindXLC: ".xlc", kindDAT: ".dat"}[kind] + ")"
		if kind == kindDAT {
			desc = "Bajeton BJ-8300 (file .dat per BT-11_CPS v" + datVer(b) + ")"
		}
		return &uniSource{Kind: kind, Desc: desc, Channels: chs, Warnings: rep.Warnings}, nil
	}
	ext := strings.ToLower(filepath.Ext(name))
	isText := !bytes.Contains(b[:min(len(b), 4096)], []byte{0})
	if isText && (ext == ".csv" || ext == ".txt" || ext == "" || isChirpCSV(b)) {
		if isChirpCSV(b) {
			chs, w, err := readChirpCSV(b)
			if err != nil {
				return nil, err
			}
			return &uniSource{Kind: "chirp", Desc: "File CSV di CHIRP", Channels: chs, Warnings: w}, nil
		}
	}
	if ext == ".xlsx" || ext == ".xls" || ext == ".csv" || ext == ".txt" || isCFB(b) || (len(b) > 4 && b[0] == 'P' && b[1] == 'K') {
		sheets, err := readTable(b, name)
		if err != nil {
			return nil, err
		}
		chs, w, err := readChanTable(sheets)
		if err != nil {
			return nil, err
		}
		return &uniSource{Kind: "table", Desc: "Tabella dei canali (" + strings.TrimPrefix(ext, ".") + ")", Channels: chs, Warnings: w}, nil
	}
	m, mem, _, err := detectGen(b, prefer)
	if err != nil {
		return nil, fmt.Errorf("formato non riconosciuto: si leggono file .xlc (K67), .dat (Bajeton), .img (radio compatibili, anche di CHIRP), CSV di CHIRP e tabelle Excel/CSV (%v)", err)
	}
	if len(mem) < genMinSize(m) {
		return nil, fmt.Errorf("immagine incompleta: %d byte, troppo pochi per un %s", len(mem), m.Name)
	}
	return &uniSource{Kind: m.ID, Desc: m.Name + " (immagine .img)", Channels: genChans(m, mem)}, nil
}

// ---------------------------------------------------------------- CSV di CHIRP

var chirpCSVHead = []string{"Location", "Name", "Frequency", "Duplex", "Offset", "Tone", "rToneFreq", "cToneFreq", "DtcsCode",
	"DtcsPolarity", "RxDtcsCode", "CrossMode", "Mode", "TStep", "Skip", "Power", "Comment", "URCALL", "RPT1CALL", "RPT2CALL", "DVCODE"}

// isChirpCSV: intestazione di CHIRP («Location,Name,Frequency…») in una delle prime
// righe (CHIRP può scrivere righe di commento prima dell'intestazione)
func isChirpCSV(b []byte) bool {
	t := decodeText(b[:min(len(b), 8192)])
	for k, line := range strings.Split(t, "\n") {
		if k > 20 {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Location,") && strings.Contains(line, "Frequency") {
			return true
		}
	}
	return false
}

// tono del modello -> parti di CHIRP
type chirpTone struct {
	kind string // "", "Tone", "DTCS"
	ctcs int    // decimi di Hz
	dcs  int    // codice ottale (es. 23)
	inv  bool
}

func toneParts(s string) chirpTone {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "OFF" {
		return chirpTone{}
	}
	if strings.HasPrefix(s, "D") && len(s) >= 4 {
		code, err := strconv.Atoi(s[1:4])
		if err == nil {
			return chirpTone{kind: "DTCS", dcs: code, inv: strings.HasSuffix(s, "I") || strings.HasSuffix(s, "R")}
		}
	}
	if v, ok := parseTone(s); ok && v > 0 {
		return chirpTone{kind: "Tone", ctcs: v}
	}
	return chirpTone{}
}

func (t chirpTone) str() string {
	switch t.kind {
	case "Tone":
		return toneStr(t.ctcs)
	case "DTCS":
		p := "N"
		if t.inv {
			p = "I"
		}
		return fmt.Sprintf("D%03d%s", t.dcs, p)
	}
	return ""
}

func pol(inv bool) string {
	if inv {
		return "R"
	}
	return "N"
}

// chirpRowOut: una riga del CSV di CHIRP per il canale
func chirpRowOut(loc int, c *Chan, comment string) []string {
	duplex, off := "", c.Tx-c.Rx
	switch {
	case c.TxOff:
		duplex, off = "off", 0
	case off > 50000000 || off < -50000000: // bande diverse: CHIRP usa "split" con la frequenza TX
		duplex, off = "split", c.Tx
	case off > 0:
		duplex = "+"
	case off < 0:
		duplex, off = "-", -off
	}
	tx, rx := toneParts(c.TxTone), toneParts(c.RxTone)
	tmode, cross := "", "Tone->Tone"
	rt, ct, dc, rdc := 885, 885, 23, 23
	dpol := "NN"
	switch {
	case c.Dig || (tx.kind == "" && rx.kind == ""):
	case tx.kind == "Tone" && rx.kind == "":
		tmode, rt = "Tone", tx.ctcs
	case tx.kind == "Tone" && rx.kind == "Tone" && tx.ctcs == rx.ctcs:
		tmode, rt, ct = "TSQL", tx.ctcs, rx.ctcs
	case tx.kind == "DTCS" && rx.kind == "DTCS" && tx.dcs == rx.dcs:
		tmode, dc, rdc, dpol = "DTCS", tx.dcs, rx.dcs, pol(tx.inv)+pol(rx.inv)
	default:
		tmode = "Cross"
		var a, b string
		if tx.kind == "Tone" {
			a, rt = "Tone", tx.ctcs
		} else if tx.kind == "DTCS" {
			a, dc = "DTCS", tx.dcs
		}
		if rx.kind == "Tone" {
			b, ct = "Tone", rx.ctcs
			if a == "" {
				rt = rx.ctcs // come CHIRP: rToneFreq uguale al tono in ricezione
			}
		} else if rx.kind == "DTCS" {
			b, rdc = "DTCS", rx.dcs
		}
		cross = a + "->" + b
		dpol = pol(tx.inv) + pol(rx.inv)
	}
	mode := "FM"
	switch {
	case c.Dig:
		mode = "DMR"
	case c.Narrow:
		mode = "NFM"
	}
	skip := ""
	if !c.Scan {
		skip = "S"
	}
	p := c.Power
	if p < 0 || p > 2 {
		p = 0
	}
	return []string{strconv.Itoa(loc), c.Name, fmtMHz6(c.Rx), duplex, fmtMHz6(off), tmode, fmtTone(rt), fmtTone(ct),
		fmt.Sprintf("%03d", dc), dpol, fmt.Sprintf("%03d", rdc), cross, mode, chirpStep(c.Rx, off), skip, chirpPowers[p], comment, "", "", "", ""}
}

// writeChirpCSV: canali (nelle posizioni sel; nil = tutti) nel CSV di CHIRP
func writeChirpCSV(chs []*Chan, sel []int) []byte {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.UseCRLF = true
	_ = w.Write(chirpCSVHead)
	if sel == nil {
		for i := range chs {
			sel = append(sel, i)
		}
	}
	for _, i := range sel {
		if i >= 0 && i < len(chs) && chs[i] != nil {
			_ = w.Write(chirpRowOut(i+1, chs[i], ""))
		}
	}
	w.Flush()
	return b.Bytes()
}

func chirpPower(s string) int {
	t := strings.ToLower(strings.TrimSpace(s))
	switch {
	case t == "" || strings.HasPrefix(t, "hi"):
		return 0
	case strings.HasPrefix(t, "mid") || strings.HasPrefix(t, "med"):
		return 1
	case strings.HasPrefix(t, "lo"):
		return 2
	}
	t = strings.TrimSuffix(strings.TrimSuffix(t, "w"), " ")
	if v, err := strconv.ParseFloat(t, 64); err == nil {
		switch {
		case v >= 3.5:
			return 0
		case v >= 1.5:
			return 1
		}
		return 2
	}
	return 0
}

// readChirpCSV legge il CSV di CHIRP (anche esportato da altri programmi).
func readChirpCSV(b []byte) ([]*Chan, []string, error) {
	r := csv.NewReader(strings.NewReader(decodeText(b)))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	recs, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("CSV di CHIRP non valido: %v", err)
	}
	if len(recs) < 2 {
		return nil, nil, errors.New("il file CSV non contiene canali")
	}
	for len(recs) > 0 && (len(recs[0]) == 0 || strings.TrimSpace(recs[0][0]) != "Location") {
		recs = recs[1:] // righe di commento prima dell'intestazione
	}
	if len(recs) < 2 {
		return nil, nil, errors.New("il file CSV non contiene canali")
	}
	col := map[string]int{}
	for i, h := range recs[0] {
		col[strings.TrimSpace(h)] = i
	}
	get := func(rec []string, k string) string {
		if i, ok := col[k]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	var chs []*Chan
	var warns []string
	for n, rec := range recs[1:] {
		line := n + 2
		rx, ok := parseMHz(get(rec, "Frequency"))
		if !ok || rx <= 0 {
			if strings.Join(rec, "") != "" {
				warns = append(warns, fmt.Sprintf("riga %d: frequenza non valida, riga saltata", line))
			}
			continue
		}
		c := &Chan{Rx: rx, Tx: rx, Name: get(rec, "Name"), Scan: get(rec, "Skip") == ""}
		off, _ := parseMHz(get(rec, "Offset"))
		switch strings.ToLower(get(rec, "Duplex")) {
		case "+":
			c.Tx = rx + off
		case "-":
			c.Tx = rx - off
		case "split":
			c.Tx = off
		case "off":
			c.TxOff = true
		}
		tone := func(k string) int {
			v, _ := parseTone(get(rec, k))
			return v
		}
		dcs := func(k string) int {
			v, _ := strconv.Atoi(get(rec, k))
			return v
		}
		pp := get(rec, "DtcsPolarity") + "NN"
		mk := func(kind string, ctcs, code int, p byte) string {
			return chirpTone{kind: kind, ctcs: ctcs, dcs: code, inv: p == 'R'}.str()
		}
		switch get(rec, "Tone") {
		case "Tone":
			c.TxTone = mk("Tone", tone("rToneFreq"), 0, 'N')
		case "TSQL":
			c.TxTone = mk("Tone", tone("cToneFreq"), 0, 'N')
			c.RxTone = c.TxTone
		case "DTCS":
			c.TxTone = mk("DTCS", 0, dcs("DtcsCode"), pp[0])
			c.RxTone = mk("DTCS", 0, dcs("DtcsCode"), pp[1])
		case "Cross":
			parts := strings.SplitN(get(rec, "CrossMode"), "->", 2)
			for len(parts) < 2 {
				parts = append(parts, "")
			}
			switch parts[0] {
			case "Tone":
				c.TxTone = mk("Tone", tone("rToneFreq"), 0, 'N')
			case "DTCS":
				c.TxTone = mk("DTCS", 0, dcs("DtcsCode"), pp[0])
			}
			switch parts[1] {
			case "Tone":
				c.RxTone = mk("Tone", tone("cToneFreq"), 0, 'N')
			case "DTCS":
				code := dcs("RxDtcsCode")
				if code == 0 {
					code = dcs("DtcsCode")
				}
				c.RxTone = mk("DTCS", 0, code, pp[1])
			}
		case "", "TSQL-R", "DTCS-R":
		default:
			warns = append(warns, fmt.Sprintf("riga %d (%s): tipo di tono «%s» non gestito, canale senza toni", line, c.Name, get(rec, "Tone")))
		}
		switch strings.ToUpper(get(rec, "Mode")) {
		case "NFM", "NAM":
			c.Narrow = true
		case "DMR", "DIG", "DV":
			c.Dig, c.Narrow = true, true
		}
		c.Power = chirpPower(get(rec, "Power"))
		chs = append(chs, c)
	}
	if len(chs) == 0 {
		return nil, warns, errors.New("il file CSV non contiene canali validi")
	}
	return chs, warns, nil
}

// ---------------------------------------------------------------- tabelle

var reCol = regexp.MustCompile(`[^a-z0-9]+`)

func colKey(s string) string { return reCol.ReplaceAllString(strings.ToLower(accenti.Replace(s)), "") }

// colonne riconosciute (tabelle esportate da qui, archivio ponti, tabelle a mano)
var tableCols = map[string]string{
	"nomecanale": "name", "nome": "name", "name": "name", "canale": "name", "sigla": "sigla",
	"rxmhz": "rx", "rx": "rx", "frequenzarx": "rx", "frequenza": "rx", "frequency": "rx", "freqrx": "rx", "ricezione": "rx",
	"txmhz": "tx", "tx": "tx", "frequenzatx": "tx", "freqtx": "tx", "trasmissione": "tx",
	"shiftmhz": "shift", "shift": "shift", "offset": "shift", "offsetmhz": "shift",
	"tonotx": "txtone", "tonetx": "txtone", "txtone": "txtone", "ctcsstx": "txtone", "subtonotx": "txtone", "tono": "txtone",
	"tonorx": "rxtone", "tonerx": "rxtone", "rxtone": "rxtone", "ctcssrx": "rxtone", "subtonorx": "rxtone",
	"potenza": "power", "power": "power", "larghezza": "width", "banda": "width", "bandwidth": "width", "mode": "width",
	"scan": "scan", "scansione": "scan", "solorx": "rxonly", "digitale": "dig", "posizione": "pos", "n": "num", "zona": "zone",
}

func cellStr(c Cell) string {
	if c.IsNum {
		return strconv.FormatFloat(c.Num, 'f', -1, 64)
	}
	return strings.TrimSpace(c.S)
}

func cellTone(c Cell) string {
	if c.IsNum {
		if c.Num <= 0 {
			return ""
		}
		return toneStr(int(math.Round(c.Num * 10)))
	}
	s := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(c.S, ",", ".")))
	if s == "" || s == "OFF" || s == "NO" || s == "-" {
		return ""
	}
	if t := toneParts(s); t.kind != "" {
		return t.str()
	}
	return ""
}

func yes(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "si", "sì", "s", "yes", "y", "x", "1", "true", "on":
		return true
	}
	return false
}

// readChanTable: canali da una tabella (prima riga con RX/Frequenza come intestazione)
func readChanTable(sheets []*Sheet) ([]*Chan, []string, error) {
	for _, sh := range sheets {
		for hr := 0; hr < len(sh.Rows) && hr < 20; hr++ {
			cols := map[string]int{}
			for i, c := range sh.Rows[hr] {
				if f, ok := tableCols[colKey(cellStr(c))]; ok {
					if _, dup := cols[f]; !dup {
						cols[f] = i
					}
				}
			}
			ri, ok := cols["rx"]
			if !ok {
				continue
			}
			var chs []*Chan
			var warns []string
			for r := hr + 1; r < len(sh.Rows); r++ {
				row := sh.Rows[r]
				g := func(f string) Cell {
					if i, ok := cols[f]; ok && i < len(row) {
						return row[i]
					}
					return Cell{Fill: -1, Font: -1}
				}
				if ri >= len(row) || cellStr(row[ri]) == "" {
					continue
				}
				rx, msg := parseFreq(g("rx"), "mhz")
				if msg != "" {
					warns = append(warns, fmt.Sprintf("riga %d: %s, riga saltata", r+1, msg))
					continue
				}
				c := &Chan{Rx: rx, Tx: rx, Scan: true}
				if _, ok := cols["tx"]; ok && cellStr(g("tx")) != "" {
					if tx, msg := parseFreq(g("tx"), "mhz"); msg == "" {
						c.Tx = tx
					}
				} else if _, ok := cols["shift"]; ok && cellStr(g("shift")) != "" {
					if v, ok := parseNumIT(cellStr(g("shift"))); ok {
						c.Tx = rx + int64(math.Round(v*1e6))
					}
				}
				c.Name = cellStr(g("name"))
				if c.Name == "" {
					c.Name = cellStr(g("sigla"))
				}
				c.TxTone, c.RxTone = cellTone(g("txtone")), cellTone(g("rxtone"))
				switch strings.ToLower(cellStr(g("power"))) {
				case "media", "mid", "medium":
					c.Power = 1
				case "bassa", "low":
					c.Power = 2
				}
				w := strings.ToLower(cellStr(g("width")))
				c.Narrow = strings.HasPrefix(w, "str") || strings.HasPrefix(w, "narr") || w == "nfm" || strings.Contains(w, "12,5") || strings.Contains(w, "12.5")
				if _, ok := cols["scan"]; ok {
					c.Scan = yes(cellStr(g("scan")))
				}
				c.TxOff = yes(cellStr(g("rxonly")))
				c.Dig = yes(cellStr(g("dig")))
				chs = append(chs, c)
			}
			if len(chs) == 0 {
				continue
			}
			return chs, warns, nil
		}
	}
	return nil, nil, errors.New("nella tabella non c'è una colonna con la frequenza (es. «RX (MHz)», «Frequenza» o «Frequency»)")
}

// tabella dei canali per qualsiasi radio (esportazione e convertitore)
func uniSheet(name string, chs []*Chan, sel []int) WSheet {
	sh := WSheet{Name: name}
	cols := []colDef{{"Posizione", 9}, {"Nome canale", 16}, {"RX (MHz)", 11}, {"TX (MHz)", 11}, {"Shift (MHz)", 9}, {"Tono TX", 7},
		{"Tono RX", 7}, {"Potenza", 7}, {"Larghezza", 9}, {"Scan", 5}, {"Solo RX", 7}, {"Digitale", 8}}
	var hdr []WCell
	for _, c := range cols {
		hdr = append(hdr, tCell(c.title, stHeader))
		sh.Widths = append(sh.Widths, c.width)
	}
	sh.Rows = append(sh.Rows, hdr)
	yn := map[bool]string{true: "Sì", false: "No"}
	if sel == nil {
		for i := range chs {
			sel = append(sel, i)
		}
	}
	for _, i := range sel {
		if i < 0 || i >= len(chs) || chs[i] == nil {
			continue
		}
		c := chs[i]
		sh.Rows = append(sh.Rows, []WCell{nCell(float64(i+1), 0), tCell(c.Name, 0), nCell(mhz(c.Rx), 0), nCell(mhz(c.Tx), 0),
			nCell(mhz(c.Tx-c.Rx), 0), toneCell(c.TxTone, 0), toneCell(c.RxTone, 0), tCell(powerName(c.Power), 0),
			tCell(larghezza(c.Narrow), 0), tCell(yn[c.Scan], 0), tCell(yn[c.TxOff], 0), tCell(yn[c.Dig], 0)})
	}
	return sh
}

// ---------------------------------------------------------------- destinazioni

type uniTarget struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Group     string `json:"group"`
	Ext       string `json:"ext"`
	NeedsBase bool   `json:"needsBase"` // serve un'immagine della radio (.img)
	Slots     int    `json:"slots,omitempty"`
}

func uniTargets() []uniTarget {
	out := []uniTarget{
		{ID: "chirp", Name: "CSV di CHIRP (universale: si apre con CHIRP e con questo programma per ogni radio)", Group: "Formati universali", Ext: "csv"},
		{ID: "xlsx", Name: "Tabella Excel (.xlsx)", Group: "Formati universali", Ext: "xlsx"},
		{ID: "csv", Name: "Tabella CSV per Excel (separatore ;)", Group: "Formati universali", Ext: "csv"},
		{ID: kindXLC, Name: "Baofeng K67 Pro / 5RH Pro (file .xlc del programma «5RH PRO»)", Group: "File dei programmi dei costruttori", Ext: "xlc", Slots: xlcChannels},
		{ID: "baj-1.0.8", Name: "Bajeton BJ-8300 (file .dat per BT-11_CPS v1.0.8)", Group: "File dei programmi dei costruttori", Ext: "dat", Slots: datZones * datPerZone},
		{ID: "baj-1.0.9", Name: "Bajeton BJ-8300 (file .dat per BT-11_CPS v1.0.9)", Group: "File dei programmi dei costruttori", Ext: "dat", Slots: datZones * datPerZone},
	}
	var gen []uniTarget
	for _, m := range genModels {
		gen = append(gen, uniTarget{ID: m.ID, Name: m.Name + " (immagine .img)", Group: "Radio con immagine .img (serve un'immagine della radio)", Ext: "img", NeedsBase: true, Slots: m.Slots})
	}
	sort.Slice(gen, func(i, j int) bool { return strings.ToLower(gen[i].Name) < strings.ToLower(gen[j].Name) })
	return append(out, gen...)
}

// uniPlace: canali nelle posizioni della radio di destinazione (stessa posizione
// se c'è; altrimenti, o con pack, nelle prime posizioni libere)
func uniPlace(src []*Chan, slots int, pack bool, rep *Report) []*Chan {
	out := make([]*Chan, slots)
	var later []*Chan
	for i, c := range src {
		if c == nil {
			continue
		}
		if !pack && i < slots && out[i] == nil {
			out[i] = c
		} else {
			later = append(later, c)
		}
	}
	free := 0
	lost := 0
	for _, c := range later {
		for free < slots && out[free] != nil {
			free++
		}
		if free >= slots {
			lost++
			continue
		}
		out[free] = c
	}
	if lost > 0 {
		rep.warn("%d canali non entrano nelle %d posizioni della radio di destinazione e sono stati esclusi", lost, slots)
	}
	return out
}

// uniConvert scrive i canali nel formato di destinazione. base: immagine o file
// della radio di destinazione (obbligatoria per le radio .img; facoltativa per
// K67 e Bajeton, che altrimenti usano i modelli inclusi).
func uniConvert(src []*Chan, target string, base []byte, pack bool, k67Band *byte) ([]byte, string, *Report, error) {
	rep := &Report{}
	switch target {
	case "chirp":
		return writeChirpCSV(src, nil), "csv", rep, nil
	case "xlsx", "csv":
		b, err := exportSheet(uniSheet("Canali", src, nil), target)
		return b, target, rep, err
	case kindXLC:
		if base == nil {
			base = blankXLC(baseXLC)
			if k67Band != nil {
				base[xlcSize-1] = *k67Band
			}
		} else if !isXLC(base) {
			return nil, "", nil, errors.New("il file di base non è un .xlc del K67")
		}
		bands := xlcBands(base, 0x10)
		chs := uniPlace(src, xlcChannels, pack, rep)
		for i, c := range chs {
			if c != nil {
				chs[i] = adaptChan(c, kindXLC, bands, posLabel(kindXLC, i)+fmt.Sprintf(" «%s»", c.Name), rep)
			}
		}
		out, err := writeRadio(kindXLC, base, chs, rep)
		return out, "xlc", rep, err
	case "baj-1.0.8", "baj-1.0.9", kindDAT:
		ver := strings.TrimPrefix(target, "baj-")
		if target == kindDAT {
			ver = "1.0.8"
		}
		if base == nil {
			b, err := blankDAT(datTemplate(ver))
			if err != nil {
				return nil, "", nil, err
			}
			base = b
		} else {
			if !isDAT(base) {
				return nil, "", nil, errors.New("il file di base non è un .dat della Bajeton")
			}
			b, err := datToVer(base, ver)
			if err != nil {
				return nil, "", nil, err
			}
			base = b
		}
		chs := uniPlace(src, datZones*datPerZone, pack, rep)
		for i, c := range chs {
			if c != nil {
				chs[i] = adaptChan(c, kindDAT, nil, posLabel(kindDAT, i)+fmt.Sprintf(" «%s»", c.Name), rep)
			}
		}
		out, err := writeRadio(kindDAT, base, chs, rep)
		return out, "dat", rep, err
	}
	m, ok := genByID[target]
	if !ok {
		return nil, "", nil, errors.New("formato di destinazione sconosciuto")
	}
	if base == nil {
		return nil, "", nil, fmt.Errorf("per creare l'immagine .img del %s serve un'immagine di quella radio (letta via cavo o aperta da file): contiene impostazioni e taratura della radio, che non si possono inventare. In alternativa usa il CSV di CHIRP", m.Name)
	}
	bm, mem, meta, err := detectGen(base, m.ID)
	if err != nil {
		return nil, "", nil, fmt.Errorf("immagine di base non riconosciuta: %v", err)
	}
	if bm.ID != m.ID && !(bm.RClass == m.RClass && fmt.Sprintf("%T", bm.Fam) == fmt.Sprintf("%T", m.Fam)) {
		return nil, "", nil, fmt.Errorf("l'immagine di base è di un %s, non di un %s", bm.Name, m.Name)
	}
	if len(mem) < genMinSize(m) {
		return nil, "", nil, fmt.Errorf("immagine di base incompleta per un %s", m.Name)
	}
	chs := uniPlace(src, m.Slots, pack, rep)
	for i, c := range chs {
		if c != nil {
			chs[i] = genAdapt(c, m, posLabel(m.ID, i)+fmt.Sprintf(" «%s»", c.Name), rep)
		}
	}
	chs = genPlace(m, mem, chs, rep)
	out, err := genWrite(m, mem, chs, rep)
	if err != nil {
		return nil, "", nil, err
	}
	return makeImg(m, out, meta), "img", rep, nil
}

// genPlace: rispetta le posizioni bloccate della radio (es. canali GMRS fissi della
// GM-15 Pro) e le regole delle singole posizioni: un canale che non può stare nella
// sua posizione va nella prima posizione libera adatta, altrimenti viene escluso;
// le posizioni bloccate conservano il canale dell'immagine di base.
func genPlace(m *genModel, mem []byte, chs []*Chan, rep *Report) []*Chan {
	lk, hasLock := m.Fam.(interface {
		Locked(m *genModel, i int, old, want *Chan) error
	})
	ok := func(i int, c *Chan) bool {
		cur := genGet(m, mem, i)
		if hasLock && lk.Locked(m, i, cur, c) != nil {
			return false
		}
		return c == nil || genValidate(m, c, i) == nil
	}
	final := make([]*Chan, m.Slots)
	fixed := make([]bool, m.Slots)
	var pending []*Chan
	for i := 0; i < m.Slots; i++ {
		var c *Chan
		if i < len(chs) {
			c = chs[i]
		}
		switch {
		case ok(i, c):
			final[i] = c
		case c != nil && ok(i, nil):
			pending = append(pending, c)
		default: // posizione bloccata: resta il canale della radio
			final[i], fixed[i] = genGet(m, mem, i), true
			if c != nil {
				pending = append(pending, c)
			}
		}
	}
	for _, c := range pending {
		placed := false
		for j := 0; j < m.Slots && !placed; j++ {
			if final[j] == nil && !fixed[j] && ok(j, c) {
				final[j], placed = c, true
			}
		}
		if !placed {
			rep.warn("«%s» (%s MHz): non può stare in nessuna posizione libera del %s, escluso", c.Name, fmtMHz(c.Rx), m.Name)
		}
	}
	return final
}
