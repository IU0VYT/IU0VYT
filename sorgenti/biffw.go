// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Scrittura di file Excel 97-2003 (.xls, BIFF8). Struttura dei record
// modellata su quella prodotta da xlwt, che Excel apre senza problemi.

import (
	"encoding/binary"
	"math"
	"unicode/utf16"
)

// stili delle celle in uscita
const (
	stNormal = iota
	stHeader
	stWarn // riga con avvisi (giallo chiaro)
	stOff  // ponte non attivo (grigio)
)

type WCell struct {
	S     string
	N     float64
	IsNum bool
	Style int
}

type WSheet struct {
	Name   string
	Rows   [][]WCell
	Widths []int // larghezza colonne in caratteri
}

type biffW struct{ b []byte }

func (w *biffW) rec(t uint16, d []byte) {
	// i record più lunghi di 8224 byte vanno spezzati in CONTINUE (gestito per SST)
	w.b = binary.LittleEndian.AppendUint16(w.b, t)
	w.b = binary.LittleEndian.AppendUint16(w.b, uint16(len(d)))
	w.b = append(w.b, d...)
}

func u16b(v ...int) []byte {
	var b []byte
	for _, x := range v {
		b = binary.LittleEndian.AppendUint16(b, uint16(x))
	}
	return b
}

func u32b(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }

// stringa breve (BOUNDSHEET, FONT): cch a 8 bit, sempre UTF-16
func shortUStr(s string) []byte {
	u := utf16.Encode([]rune(s))
	if len(u) > 255 {
		u = u[:255]
	}
	b := []byte{byte(len(u)), 1}
	for _, c := range u {
		b = binary.LittleEndian.AppendUint16(b, c)
	}
	return b
}

func biffFont(bold bool) []byte {
	g := 0
	if bold {
		g = 1 // come xlwt/Excel: bit "grassetto" oltre al peso 700
	}
	d := u16b(200, g, 0x7FFF)
	if bold {
		d = append(d, u16b(700)...)
	} else {
		d = append(d, u16b(400)...)
	}
	d = append(d, u16b(0)...) // sss
	d = append(d, 0, 0, 0, 0) // uls, family, charset, riservato
	d = append(d, shortUStr("Arial")...)
	return d
}

func biffXF(font int, style bool, fill int) []byte {
	d := u16b(font, 0)
	if style {
		d = append(d, u16b(0xFFF5)...)
	} else {
		d = append(d, u16b(0x0001)...)
	}
	d = append(d, 0x20, 0, 0)
	if style {
		d = append(d, 0xF4)
	} else {
		d = append(d, 0xF8)
	}
	d = append(d, 0, 0, 0, 0) // bordi
	if fill >= 0 {
		d = append(d, u32b(1<<26)...)
		d = append(d, u16b(fill|0x41<<7)...)
	} else {
		d = append(d, u32b(0)...)
		d = append(d, u16b(0x20C0)...)
	}
	return d
}

func writeXLS(sheets []WSheet) ([]byte, error) {
	// tabella delle stringhe condivise
	sstIdx := map[string]int{}
	var sst []string
	total := 0
	for _, sh := range sheets {
		for _, r := range sh.Rows {
			for _, c := range r {
				if c.IsNum || c.S == "" {
					continue
				}
				total++
				if _, ok := sstIdx[c.S]; !ok {
					sstIdx[c.S] = len(sst)
					sst = append(sst, c.S)
				}
			}
		}
	}
	g := &biffW{}
	g.rec(0x0809, append(u16b(0x0600, 0x0005, 0x0DBB, 0x07CC), append(u32b(0), u32b(6)...)...))
	g.rec(0x0042, u16b(0x04B0))                                               // CODEPAGE UTF-16
	g.rec(0x003D, u16b(0x01E0, 0x005A, 0x3FCF, 0x2A4E, 0x38, 0, 0, 1, 0x258)) // WINDOW1
	g.rec(0x0022, u16b(0))                                                    // DATEMODE 1900
	for i := 0; i < 5; i++ {
		g.rec(0x0031, biffFont(false)) // indici 0,1,2,3,5
	}
	g.rec(0x0031, biffFont(true)) // indice 6: grassetto
	for i := 0; i < 15; i++ {
		g.rec(0x00E0, biffXF(0, true, -1))
	}
	g.rec(0x00E0, biffXF(0, false, -1))           // 15 normale
	g.rec(0x00E0, biffXF(6, false, 22))           // 16 intestazione: grassetto su grigio
	g.rec(0x00E0, biffXF(0, false, 43))           // 17 avviso: giallo chiaro
	g.rec(0x00E0, biffXF(0, false, 45))           // 18 non attivo: rosa
	g.rec(0x0293, []byte{0x00, 0x80, 0x00, 0xFF}) // STYLE Normale
	bsPos := make([]int, len(sheets))
	for i, sh := range sheets {
		bsPos[i] = len(g.b) + 4
		name := sh.Name
		if name == "" {
			name = "Foglio1"
		}
		g.rec(0x0085, append(append(u32b(0), 0, 0), shortUStr(name)...))
	}
	// SST con eventuali CONTINUE
	{
		const max = 8224
		cur := append(u32b(uint32(total)), u32b(uint32(len(sst)))...)
		first := true
		flush := func() {
			if first {
				g.rec(0x00FC, cur)
				first = false
			} else {
				g.rec(0x003C, cur)
			}
			cur = nil
		}
		for _, s := range sst {
			u := utf16.Encode([]rune(s))
			if len(cur)+3+2 > max { // l'intestazione della stringa non si spezza
				flush()
			}
			cur = append(cur, u16b(len(u))...)
			cur = append(cur, 1)
			for i := 0; i < len(u); i++ {
				if len(cur)+2 > max {
					flush()
					cur = append(cur, 1) // flag: caratteri a 16 bit
				}
				cur = binary.LittleEndian.AppendUint16(cur, u[i])
			}
		}
		flush()
	}
	g.rec(0x000A, nil)
	for i, sh := range sheets {
		binary.LittleEndian.PutUint32(g.b[bsPos[i]:], uint32(len(g.b)))
		w := &biffW{}
		w.rec(0x0809, append(u16b(0x0600, 0x0010, 0x0DBB, 0x07CC), append(u32b(0), u32b(6)...)...))
		ncol := 0
		for _, r := range sh.Rows {
			ncol = max(ncol, len(r))
		}
		ncol = min(ncol, 256)
		nrow := min(len(sh.Rows), 65536)
		w.rec(0x0200, append(append(u32b(0), u32b(uint32(nrow))...), u16b(0, ncol, 0)...))
		for c, wd := range sh.Widths {
			if c >= 256 || wd <= 0 {
				continue
			}
			w.rec(0x007D, u16b(c, c, min(wd*256+200, 65535), 15, 0, 0))
		}
		for r := 0; r < nrow; r++ {
			for c, cell := range sh.Rows[r] {
				if c >= 256 {
					break
				}
				xf := 15 + cell.Style
				if cell.IsNum {
					d := u16b(r, c, xf)
					d = binary.LittleEndian.AppendUint64(d, math.Float64bits(cell.N))
					w.rec(0x0203, d)
				} else if cell.S != "" {
					w.rec(0x00FD, append(u16b(r, c, xf), u32b(uint32(sstIdx[cell.S]))...))
				} else if cell.Style != stNormal {
					w.rec(0x0201, u16b(r, c, xf)) // BLANK colorata
				}
			}
		}
		grbit := 0x06B6
		if i != 0 {
			grbit &^= 0x0600 // solo il primo foglio selezionato
		}
		w.rec(0x023E, u16b(grbit|0x0108, 0, 0, 64, 0, 0, 0, 0, 0)) // WINDOW2 con riquadro bloccato
		w.rec(0x0041, append(u16b(0, 1, 1, 0), 2, 0))              // PANE: blocca la prima riga
		w.rec(0x000A, nil)
		g.b = append(g.b, w.b...)
	}
	return writeCFB("Workbook", g.b)
}
