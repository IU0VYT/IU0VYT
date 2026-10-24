// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura dei file Excel 97-2003 (.xls, formato BIFF8), compresi i colori
// di sfondo delle celle (servono a riconoscere lo stato dei ponti nel file
// di IK2ANE).

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"unicode/utf16"
)

type Cell struct {
	S     string
	Num   float64
	IsNum bool
	Fill  int // indice colore di sfondo (-1 = nessuno)
	Font  int // indice colore del carattere (-1 = automatico)
}

type Sheet struct {
	Name string
	Rows [][]Cell
}

func (s *Sheet) set(r, c int, cell Cell) {
	if r < 0 || c < 0 || r > 200000 || c > 1000 {
		return
	}
	for len(s.Rows) <= r {
		s.Rows = append(s.Rows, nil)
	}
	row := s.Rows[r]
	for len(row) <= c {
		row = append(row, Cell{Fill: -1, Font: -1})
	}
	row[c] = cell
	s.Rows[r] = row
}

func (s *Sheet) get(r, c int) Cell {
	if r < 0 || r >= len(s.Rows) || c < 0 || c >= len(s.Rows[r]) {
		return Cell{Fill: -1, Font: -1}
	}
	return s.Rows[r][c]
}

type biffRec struct {
	typ  uint16
	data []byte
	cont [][]byte // record CONTINUE successivi
}

func biffRecords(b []byte, pos int) ([]biffRec, error) {
	var out []biffRec
	le := binary.LittleEndian
	for pos+4 <= len(b) {
		t := le.Uint16(b[pos:])
		l := int(le.Uint16(b[pos+2:]))
		if pos+4+l > len(b) {
			return out, errors.New("record troncato")
		}
		d := b[pos+4 : pos+4+l]
		pos += 4 + l
		if t == 0x003C && len(out) > 0 {
			out[len(out)-1].cont = append(out[len(out)-1].cont, d)
			continue
		}
		out = append(out, biffRec{typ: t, data: d})
		if t == 0x000A { // EOF
			break
		}
	}
	return out, nil
}

// lettore di stringhe Unicode BIFF8 che attraversa i record CONTINUE
type biffStr struct {
	parts [][]byte
	pi    int
	p     int
}

func (r *biffStr) avail() int {
	if r.pi >= len(r.parts) {
		return 0
	}
	return len(r.parts[r.pi]) - r.p
}

func (r *biffStr) next() bool {
	if r.pi+1 >= len(r.parts) {
		return false
	}
	r.pi++
	r.p = 0
	return true
}

func (r *biffStr) bytes(n int) ([]byte, error) {
	out := make([]byte, 0, n)
	for n > 0 {
		if r.avail() == 0 && !r.next() {
			return nil, errors.New("stringa troncata")
		}
		k := min(n, r.avail())
		out = append(out, r.parts[r.pi][r.p:r.p+k]...)
		r.p += k
		n -= k
	}
	return out, nil
}

func (r *biffStr) u8() (byte, error) {
	b, err := r.bytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *biffStr) u16() (int, error) {
	b, err := r.bytes(2)
	if err != nil {
		return 0, err
	}
	return int(binary.LittleEndian.Uint16(b)), nil
}

func (r *biffStr) u32() (int, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return int(binary.LittleEndian.Uint32(b)), nil
}

// str legge una XLUnicodeRichExtendedString (cch già letto se lenGiven>=0).
func (r *biffStr) str(cch int) (string, error) {
	flags, err := r.u8()
	if err != nil {
		return "", err
	}
	runs, ext := 0, 0
	if flags&0x08 != 0 {
		if runs, err = r.u16(); err != nil {
			return "", err
		}
	}
	if flags&0x04 != 0 {
		if ext, err = r.u32(); err != nil {
			return "", err
		}
	}
	high := flags&0x01 != 0
	u := make([]uint16, 0, cch)
	for len(u) < cch {
		if r.avail() == 0 {
			// la stringa continua nel record successivo, preceduta da un byte di flag
			if !r.next() {
				return "", errors.New("stringa troncata")
			}
			f, err := r.u8()
			if err != nil {
				return "", err
			}
			high = f&0x01 != 0
		}
		if high {
			if r.avail() < 2 {
				return "", errors.New("carattere spezzato")
			}
			b, _ := r.bytes(2)
			u = append(u, binary.LittleEndian.Uint16(b))
		} else {
			b, _ := r.bytes(1)
			u = append(u, uint16(b[0]))
		}
	}
	if _, err := r.bytes(4*runs + ext); err != nil {
		return "", err
	}
	return string(utf16.Decode(u)), nil
}

func rkValue(rk uint32) float64 {
	var v float64
	if rk&2 != 0 {
		v = float64(int32(rk) >> 2)
	} else {
		v = math.Float64frombits(uint64(rk&0xFFFFFFFC) << 32)
	}
	if rk&1 != 0 {
		v /= 100
	}
	return v
}

func numCell(v float64) Cell {
	return Cell{S: strconv.FormatFloat(v, 'f', -1, 64), Num: v, IsNum: true, Fill: -1, Font: -1}
}

// readXLS legge tutti i fogli di un file .xls.
func readXLS(b []byte) ([]*Sheet, error) {
	f, err := openCFB(b)
	if err != nil {
		return nil, err
	}
	wb, err := f.stream("Workbook")
	if err != nil {
		if _, e2 := f.stream("Book"); e2 == nil {
			return nil, errors.New("file Excel 5.0/95 non supportato: salvalo in formato Excel 97-2003 o .xlsx")
		}
		return nil, err
	}
	return parseBIFF(wb)
}

func parseBIFF(wb []byte) (sheets []*Sheet, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("file .xls non valido: %v", e)
		}
	}()
	le := binary.LittleEndian
	glob, err := biffRecords(wb, 0)
	if err != nil {
		return nil, err
	}
	if len(glob) == 0 || glob[0].typ != 0x0809 || len(glob[0].data) < 4 || le.Uint16(glob[0].data) != 0x0600 {
		return nil, errors.New("solo i file Excel 97-2003 (BIFF8) sono supportati")
	}
	type xf struct{ font, fill int }
	var fonts []int // colore di ciascun font
	var xfs []xf
	var sst []string
	type bs struct {
		name string
		pos  int
		typ  byte
	}
	var bsheets []bs
	for _, r := range glob {
		d := r.data
		switch r.typ {
		case 0x0031: // FONT
			if len(fonts) == 4 {
				fonts = append(fonts, -1) // l'indice 4 non esiste
			}
			c := -1
			if len(d) >= 6 {
				c = int(le.Uint16(d[4:]))
				if c == 0x7FFF {
					c = -1
				}
			}
			fonts = append(fonts, c)
		case 0x00E0: // XF
			x := xf{fill: -1}
			if len(d) >= 20 {
				x.font = int(le.Uint16(d[0:]))
				pat := le.Uint32(d[14:]) >> 26
				if pat != 0 {
					x.fill = int(le.Uint16(d[18:]) & 0x7F)
					if x.fill >= 64 {
						x.fill = -1
					}
				}
			}
			xfs = append(xfs, x)
		case 0x0085: // BOUNDSHEET
			if len(d) < 8 {
				continue
			}
			rs := &biffStr{parts: [][]byte{d[7:]}}
			name, e := rs.str(int(d[6]))
			if e != nil {
				return nil, e
			}
			bsheets = append(bsheets, bs{name: name, pos: int(le.Uint32(d)), typ: d[5]})
		case 0x00FC: // SST
			rs := &biffStr{parts: append([][]byte{d}, r.cont...)}
			if _, e := rs.u32(); e != nil {
				return nil, e
			}
			n, e := rs.u32()
			if e != nil {
				return nil, e
			}
			for i := 0; i < n; i++ {
				cch, e := rs.u16()
				if e != nil {
					break
				}
				s, e := rs.str(cch)
				if e != nil {
					break
				}
				sst = append(sst, s)
			}
		}
	}
	style := func(ixf int) (int, int) {
		if ixf < 0 || ixf >= len(xfs) {
			return -1, -1
		}
		x := xfs[ixf]
		fc := -1
		if x.font >= 0 && x.font < len(fonts) {
			fc = fonts[x.font]
		}
		return x.fill, fc
	}
	for _, bsh := range bsheets {
		sh := &Sheet{Name: bsh.name}
		sheets = append(sheets, sh)
		if bsh.typ != 0 || bsh.pos <= 0 || bsh.pos >= len(wb) {
			continue // grafici, macro...
		}
		recs, _ := biffRecords(wb, bsh.pos)
		cell := func(d []byte, c Cell) {
			row, col := int(le.Uint16(d[0:])), int(le.Uint16(d[2:]))
			c.Fill, c.Font = style(int(le.Uint16(d[4:])))
			sh.set(row, col, c)
		}
		for k, r := range recs {
			d := r.data
			switch r.typ {
			case 0x00FD: // LABELSST
				if len(d) >= 10 {
					i := int(le.Uint32(d[6:]))
					s := ""
					if i < len(sst) {
						s = sst[i]
					}
					cell(d, Cell{S: s})
				}
			case 0x0204: // LABEL
				if len(d) >= 9 {
					rs := &biffStr{parts: append([][]byte{d[8:]}, r.cont...)}
					s, _ := rs.str(int(le.Uint16(d[6:])))
					cell(d, Cell{S: s})
				}
			case 0x0203: // NUMBER
				if len(d) >= 14 {
					cell(d, numCell(math.Float64frombits(le.Uint64(d[6:]))))
				}
			case 0x027E: // RK
				if len(d) >= 10 {
					cell(d, numCell(rkValue(le.Uint32(d[6:]))))
				}
			case 0x00BD: // MULRK
				if len(d) >= 6 {
					row, c0 := int(le.Uint16(d[0:])), int(le.Uint16(d[2:]))
					n := (len(d) - 6) / 6
					for i := 0; i < n; i++ {
						p := 4 + 6*i
						c := numCell(rkValue(le.Uint32(d[p+2:])))
						c.Fill, c.Font = style(int(le.Uint16(d[p:])))
						sh.set(row, c0+i, c)
					}
				}
			case 0x0201: // BLANK (conserva il colore)
				if len(d) >= 6 {
					cell(d, Cell{})
				}
			case 0x00BE: // MULBLANK
				if len(d) >= 6 {
					row, c0 := int(le.Uint16(d[0:])), int(le.Uint16(d[2:]))
					n := (len(d) - 6) / 2
					for i := 0; i < n; i++ {
						c := Cell{}
						c.Fill, c.Font = style(int(le.Uint16(d[4+2*i:])))
						sh.set(row, c0+i, c)
					}
				}
			case 0x0205: // BOOLERR
				if len(d) >= 8 {
					if d[7] == 0 {
						cell(d, Cell{S: map[bool]string{true: "VERO", false: "FALSO"}[d[6] != 0]})
					} else {
						cell(d, Cell{S: "#ERR"})
					}
				}
			case 0x0006: // FORMULA: si usa il risultato
				if len(d) >= 14 {
					if le.Uint16(d[12:]) != 0xFFFF {
						cell(d, numCell(math.Float64frombits(le.Uint64(d[6:]))))
					} else if d[6] == 0 { // risultato testo nel record STRING successivo
						s := ""
						for j := k + 1; j < len(recs) && j <= k+3; j++ {
							if recs[j].typ == 0x0207 && len(recs[j].data) >= 3 {
								rs := &biffStr{parts: append([][]byte{recs[j].data[2:]}, recs[j].cont...)}
								s, _ = rs.str(int(le.Uint16(recs[j].data)))
								break
							}
						}
						cell(d, Cell{S: s})
					} else if d[6] == 1 {
						cell(d, Cell{S: map[bool]string{true: "VERO", false: "FALSO"}[d[8] != 0]})
					}
				}
			}
		}
	}
	return sheets, nil
}
