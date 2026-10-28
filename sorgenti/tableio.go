// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura di tabelle da .xls, .xlsx e .csv; scrittura di ponti e canali in
// .xls, .xlsx, CSV (Excel italiano) e CSV compatibile CHIRP.

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// readTable riconosce il formato dal contenuto (non dall'estensione).
func readTable(b []byte, name string) ([]*Sheet, error) {
	switch {
	case isCFB(b):
		return readXLS(b)
	case len(b) > 4 && b[0] == 'P' && b[1] == 'K':
		return readXLSX(b)
	case len(b) == 0:
		return nil, errors.New("file vuoto")
	}
	if bytes.Contains(b[:min(len(b), 2048)], []byte{0}) {
		return nil, errors.New("formato non riconosciuto: usa un file .xls, .xlsx o .csv")
	}
	return readCSV(b, name)
}

func decodeText(b []byte) string {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(b) {
		return string(b)
	}
	// Windows-1252 (il caso tipico dei CSV salvati da Excel italiano)
	var sb strings.Builder
	cp := map[byte]rune{0x80: '€', 0x82: '‚', 0x84: '„', 0x85: '…', 0x91: '‘', 0x92: '’', 0x93: '“', 0x94: '”', 0x96: '–', 0x97: '—', 0x99: '™'}
	for _, c := range b {
		if r, ok := cp[c]; ok {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(rune(c))
		}
	}
	return sb.String()
}

func readCSV(b []byte, name string) ([]*Sheet, error) {
	text := decodeText(b)
	first := text
	if i := strings.IndexAny(text, "\r\n"); i >= 0 {
		first = text[:i]
	}
	sep := ','
	best := strings.Count(first, ",")
	if n := strings.Count(first, ";"); n > best {
		sep, best = ';', n
	}
	if n := strings.Count(first, "\t"); n > best {
		sep = '\t'
	}
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = sep
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	sh := &Sheet{Name: "CSV"}
	for row := 0; ; row++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV non valido alla riga %d: %v", row+1, err)
		}
		cells := make([]Cell, len(rec))
		for i, v := range rec {
			cells[i] = Cell{S: v, Fill: -1, Font: -1}
		}
		sh.Rows = append(sh.Rows, cells)
	}
	return []*Sheet{sh}, nil
}

// ---------------------------------------------------------------- tabelle in uscita

type colDef struct {
	title string
	width int
}

func tCell(s string, st int) WCell  { return WCell{S: s, Style: st} }
func nCell(v float64, st int) WCell { return WCell{N: v, IsNum: true, Style: st} }
func mhz(hz int64) float64          { return float64(hz) / 1e6 }
func toneCell(t string, st int) WCell {
	if v, ok := parseTone(t); ok {
		if v == 0 {
			return tCell("", st)
		}
		return nCell(float64(v)/10, st)
	}
	return tCell(t, st)
}

var pontiCols = []colDef{{"Nome canale", 16}, {"Sigla", 8}, {"Località", 24}, {"Provincia", 6}, {"Regione", 14}, {"Banda", 6},
	{"RX (MHz)", 11}, {"TX (MHz)", 11}, {"Shift (MHz)", 9}, {"Tono TX", 7}, {"Tono RX", 7}, {"Larghezza", 9}, {"Potenza", 7},
	{"Tipo", 18}, {"Rete", 16}, {"Nominativo", 10}, {"Nodo EchoLink", 8}, {"Locatore", 8}, {"Stato", 14}, {"Avvisi", 40}, {"Note", 40}}

func larghezza(n bool) string {
	if n {
		return "Stretta"
	}
	return "Larga"
}

func pontiSheet(pp []*Ponte) WSheet {
	sh := WSheet{Name: "Ponti"}
	var hdr []WCell
	for _, c := range pontiCols {
		hdr = append(hdr, tCell(c.title, stHeader))
		sh.Widths = append(sh.Widths, c.width)
	}
	sh.Rows = append(sh.Rows, hdr)
	for _, p := range pp {
		st := stNormal
		if p.Stato == statoOff || p.Stato == statoTemp {
			st = stOff
		} else if len(p.Avvisi) > 0 {
			st = stWarn
		}
		sh.Rows = append(sh.Rows, []WCell{tCell(p.Nome, st), tCell(p.Sigla, st), tCell(p.Localita, st), tCell(p.Provincia, st),
			tCell(p.Regione, st), tCell(bandOf(p.Rx), st), nCell(mhz(p.Rx), st), nCell(mhz(p.Tx), st), nCell(mhz(p.Tx-p.Rx), st),
			toneCell(p.TxTone, st), toneCell(p.RxTone, st), tCell(larghezza(p.Narrow), st), tCell(powerName(p.Power), st),
			tCell(strings.Join(p.Tipi, ", "), st), tCell(p.Rete, st), tCell(p.Nominativo, st), tCell(p.Nodo, st), tCell(p.Locatore, st),
			tCell(statoText(p.Stato), st), tCell(strings.Join(p.Avvisi, "; "), st), tCell(p.Note, st)})
	}
	return sh
}

func statoText(s string) string {
	if s == "" {
		return "Attivo"
	}
	return s
}

func chanSheet(kind string, chs []*Chan, sel []int) WSheet {
	sh := WSheet{Name: radioName(kind)}
	cols := []colDef{{"Posizione", 9}}
	if kind == kindDAT {
		cols = append(cols, colDef{"Zona", 5}, colDef{"N.", 4})
	}
	cols = append(cols, []colDef{{"Nome canale", 16}, {"RX (MHz)", 11}, {"TX (MHz)", 11}, {"Shift (MHz)", 9}, {"Tono TX", 7},
		{"Tono RX", 7}, {"Potenza", 7}, {"Larghezza", 9}}...)
	if kind == kindDAT {
		cols = append(cols, colDef{"Scan", 5})
	}
	var hdr []WCell
	for _, c := range cols {
		hdr = append(hdr, tCell(c.title, stHeader))
		sh.Widths = append(sh.Widths, c.width)
	}
	sh.Rows = append(sh.Rows, hdr)
	for _, i := range sel {
		if i < 0 || i >= len(chs) || chs[i] == nil {
			continue
		}
		c := chs[i]
		row := []WCell{nCell(float64(i+1), 0)}
		if kind == kindDAT {
			row = append(row, nCell(float64(i/datPerZone+1), 0), nCell(float64(i%datPerZone+1), 0))
		}
		row = append(row, tCell(c.Name, 0), nCell(mhz(c.Rx), 0), nCell(mhz(c.Tx), 0), nCell(mhz(c.Tx-c.Rx), 0),
			toneCell(c.TxTone, 0), toneCell(c.RxTone, 0), tCell(powerName(c.Power), 0), tCell(larghezza(c.Narrow), 0))
		if kind == kindDAT {
			row = append(row, tCell(map[bool]string{true: "Sì", false: "No"}[c.Scan], 0))
		}
		sh.Rows = append(sh.Rows, row)
	}
	return sh
}

// CSV per Excel italiano: separatore ";" e virgola decimale.
func sheetCSV(sh WSheet) []byte {
	var b bytes.Buffer
	b.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&b)
	w.Comma = ';'
	w.UseCRLF = true
	for _, row := range sh.Rows {
		rec := make([]string, len(row))
		for i, c := range row {
			if c.IsNum {
				rec[i] = strings.ReplaceAll(strconv.FormatFloat(c.N, 'f', -1, 64), ".", ",")
			} else {
				rec[i] = c.S
			}
		}
		_ = w.Write(rec)
	}
	w.Flush()
	return b.Bytes()
}

// ---------------------------------------------------------------- CHIRP

// Riga CHIRP: la colonna Power usa i watt (es. "5.0W"), come le versioni attuali di CHIRP.
type chirpMem struct {
	loc     int
	name    string
	rx, tx  int64
	txTone  string
	rxTone  string
	narrow  bool
	power   int
	skip    bool
	comment string
}

var chirpPowers = []string{"5.0W", "2.0W", "1.0W"}

func chirpStep(hz ...int64) string {
	for _, st := range []struct {
		v int64
		s string
	}{{12500, "12.50"}, {6250, "6.25"}, {5000, "5.00"}, {2500, "2.50"}} {
		ok := true
		for _, h := range hz {
			if h%st.v != 0 {
				ok = false
			}
		}
		if ok {
			return st.s
		}
	}
	return "12.50"
}

func chirpCSV(mems []chirpMem) []byte {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.UseCRLF = true
	_ = w.Write(chirpCSVHead)
	for _, m := range mems {
		c := &Chan{Rx: m.rx, Tx: m.tx, TxTone: m.txTone, RxTone: m.rxTone, Narrow: m.narrow, Power: m.power, Name: m.name, Scan: !m.skip}
		_ = w.Write(chirpRowOut(m.loc, c, m.comment))
	}
	w.Flush()
	return b.Bytes()
}

func pontiChirp(pp []*Ponte) []byte {
	var mems []chirpMem
	for i, p := range pp {
		var parts []string
		for _, s := range []string{p.Sigla, p.Localita, strings.Join(p.Tipi, "/"), p.Rete, p.Nominativo} {
			if s = strings.TrimSpace(s); s != "" {
				parts = append(parts, s)
			}
		}
		mems = append(mems, chirpMem{loc: i + 1, name: p.Nome, rx: p.Rx, tx: p.Tx, txTone: p.TxTone, rxTone: p.RxTone,
			narrow: p.Narrow, power: p.Power, comment: strings.Join(parts, " - ")})
	}
	return chirpCSV(mems)
}

// chanChirp: canali di una radio nel CSV di CHIRP (dalla 3.7 con DCS, sola
// ricezione, canali digitali e scansione di tutte le radio)
func chanChirp(kind string, chs []*Chan, sel []int) []byte { return writeChirpCSV(chs, sel) }

// exportSheet scrive una tabella nel formato richiesto: xlsx, xls, csv.
func exportSheet(sh WSheet, format string) ([]byte, error) {
	switch format {
	case "xlsx":
		return writeXLSX([]WSheet{sh})
	case "xls":
		return writeXLS([]WSheet{sh})
	case "csv":
		return sheetCSV(sh), nil
	}
	return nil, fmt.Errorf("formato %s non gestito", format)
}
