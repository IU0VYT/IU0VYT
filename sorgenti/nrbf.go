// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura/scrittura del formato MS-NRBF (.NET BinaryFormatter) usato dal
// software BT-11_CPS della Bajeton 8300. Il parser conserva tutti i record
// così com'erano, per cui rileggere e riscrivere un file senza modifiche
// produce gli stessi identici byte.

import (
	"encoding/binary"
	"fmt"
)

type ClassTypeInfo struct {
	Name string
	Lib  int32
}

type Val struct {
	Prim bool
	PT   byte
	Raw  []byte
	R    *Rec
	Off  int // posizione di Raw nel file letto (membri primitivi; 0 se creato dopo)
}

type Rec struct {
	T     byte
	Hdr   [4]int32
	ID    int32
	MID   int32
	Lib   int32
	Name  string
	Names []string
	BTs   []byte
	Add   []interface{}
	Vals  []Val
	S     string
	AT    byte
	Lens  []int32
	LB    []int32
	BT    byte
	Add1  interface{}
	Els   []Val
	PT    byte
	PRaw  []byte
	Ref   int32
	N     int32
	PEls  [][]byte
}

type Doc struct {
	Recs  []*Rec
	Meta  map[int32]*Rec
	Objs  map[int32]*Rec
	maxID int32
}

var primSize = map[byte]int{1: 1, 2: 1, 6: 8, 7: 2, 8: 4, 9: 8, 10: 1, 11: 4, 12: 8, 13: 8, 14: 2, 15: 4, 16: 8, 17: 0}

type reader struct {
	b []byte
	p int
}

func (r *reader) need(n int) {
	if n < 0 || r.p+n > len(r.b) {
		panic(fmt.Errorf("file troncato o non valido (posizione %d)", r.p))
	}
}
func (r *reader) u8() byte { r.need(1); v := r.b[r.p]; r.p++; return v }
func (r *reader) i32() int32 {
	r.need(4)
	v := int32(binary.LittleEndian.Uint32(r.b[r.p:]))
	r.p += 4
	return v
}
func (r *reader) len7() int {
	n, sh := 0, 0
	for i := 0; i < 5; i++ {
		b := r.u8()
		n |= int(b&0x7f) << sh
		sh += 7
		if b&0x80 == 0 {
			return n
		}
	}
	panic(fmt.Errorf("lunghezza stringa non valida"))
}
func (r *reader) lps() string {
	n := r.len7()
	r.need(n)
	s := string(r.b[r.p : r.p+n])
	r.p += n
	return s
}
func (r *reader) primRaw(pt byte) []byte {
	start := r.p
	switch pt {
	case 18, 5:
		n := r.len7()
		r.need(n)
		r.p += n
	case 3:
		r.need(1)
		c := r.b[r.p]
		n := 1
		switch {
		case c >= 0xf0:
			n = 4
		case c >= 0xe0:
			n = 3
		case c >= 0x80:
			n = 2
		}
		r.need(n)
		r.p += n
	default:
		n, ok := primSize[pt]
		if !ok {
			panic(fmt.Errorf("tipo primitivo sconosciuto %d", pt))
		}
		r.need(n)
		r.p += n
	}
	out := make([]byte, r.p-start)
	copy(out, r.b[start:r.p])
	return out
}

func ParseNRBF(b []byte) (d *Doc, err error) {
	defer func() {
		if e := recover(); e != nil {
			d = nil
			err = fmt.Errorf("%v", e)
		}
	}()
	d = &Doc{Meta: map[int32]*Rec{}, Objs: map[int32]*Rec{}}
	r := &reader{b: b}
	for {
		rec := d.rd(r)
		d.Recs = append(d.Recs, rec)
		if rec.T == 11 {
			break
		}
	}
	if r.p != len(b) {
		return nil, fmt.Errorf("dati in eccesso alla fine del file")
	}
	return d, nil
}

func (d *Doc) reg(rec *Rec) {
	d.Objs[rec.ID] = rec
	if rec.ID > d.maxID {
		d.maxID = rec.ID
	}
}

func (d *Doc) addOne(r *reader, bt byte) interface{} {
	switch bt {
	case 0, 7:
		return r.u8()
	case 3:
		return r.lps()
	case 4:
		n := r.lps()
		return ClassTypeInfo{n, r.i32()}
	}
	return nil
}

func (d *Doc) members(r *reader, m *Rec) []Val {
	out := make([]Val, len(m.Names))
	for i := range m.Names {
		if m.T == 5 || m.T == 4 {
			if m.BTs[i] == 0 {
				pt := m.Add[i].(byte)
				off := r.p
				out[i] = Val{Prim: true, PT: pt, Raw: r.primRaw(pt), Off: off}
				continue
			}
		}
		out[i] = Val{R: d.rd(r)}
	}
	return out
}

func (d *Doc) elems(r *reader, tot int, bt byte, add interface{}) []Val {
	var els []Val
	for c := 0; c < tot; {
		if bt == 0 {
			pt := add.(byte)
			els = append(els, Val{Prim: true, PT: pt, Raw: r.primRaw(pt)})
			c++
			continue
		}
		e := d.rd(r)
		els = append(els, Val{R: e})
		if e.T == 13 || e.T == 14 {
			c += int(e.N)
		} else {
			c++
		}
	}
	return els
}

func (d *Doc) rd(r *reader) *Rec {
	t := r.u8()
	rec := &Rec{T: t}
	switch t {
	case 0:
		for i := range rec.Hdr {
			rec.Hdr[i] = r.i32()
		}
	case 12:
		rec.ID = r.i32()
		rec.Name = r.lps()
	case 5, 4, 3, 2:
		rec.ID = r.i32()
		rec.Name = r.lps()
		n := int(r.i32())
		if n < 0 || n > 10000 {
			panic(fmt.Errorf("numero membri non valido"))
		}
		for i := 0; i < n; i++ {
			rec.Names = append(rec.Names, r.lps())
		}
		if t == 5 || t == 4 {
			for i := 0; i < n; i++ {
				rec.BTs = append(rec.BTs, r.u8())
			}
			for i := 0; i < n; i++ {
				rec.Add = append(rec.Add, d.addOne(r, rec.BTs[i]))
			}
		}
		if t == 5 || t == 3 {
			rec.Lib = r.i32()
		}
		d.Meta[rec.ID] = rec
		rec.Vals = d.members(r, rec)
		d.reg(rec)
	case 1:
		rec.ID = r.i32()
		rec.MID = r.i32()
		m, ok := d.Meta[rec.MID]
		if !ok {
			panic(fmt.Errorf("metadati %d mancanti", rec.MID))
		}
		d.Meta[rec.ID] = m
		rec.Vals = d.members(r, m)
		d.reg(rec)
	case 6:
		rec.ID = r.i32()
		rec.S = r.lps()
		d.reg(rec)
	case 7:
		rec.ID = r.i32()
		rec.AT = r.u8()
		rk := int(r.i32())
		if rk < 1 || rk > 32 {
			panic(fmt.Errorf("rank array non valido"))
		}
		tot := 1
		for i := 0; i < rk; i++ {
			l := r.i32()
			rec.Lens = append(rec.Lens, l)
			tot *= int(l)
		}
		if rec.AT == 3 || rec.AT == 4 || rec.AT == 5 {
			for i := 0; i < rk; i++ {
				rec.LB = append(rec.LB, r.i32())
			}
		}
		rec.BT = r.u8()
		rec.Add1 = d.addOne(r, rec.BT)
		rec.Els = d.elems(r, tot, rec.BT, rec.Add1)
		d.reg(rec)
	case 8:
		rec.PT = r.u8()
		rec.PRaw = r.primRaw(rec.PT)
	case 9:
		rec.Ref = r.i32()
	case 10, 11:
	case 13:
		rec.N = int32(r.u8())
	case 14:
		rec.N = r.i32()
	case 15:
		rec.ID = r.i32()
		n := int(r.i32())
		rec.PT = r.u8()
		for i := 0; i < n; i++ {
			rec.PEls = append(rec.PEls, r.primRaw(rec.PT))
		}
		d.reg(rec)
	case 16, 17:
		rec.ID = r.i32()
		n := int(r.i32())
		rec.Els = d.elems(r, n, 2, nil)
		d.reg(rec)
	default:
		panic(fmt.Errorf("record di tipo %d non supportato (posizione %d)", t, r.p-1))
	}
	return rec
}

// ---------------- scrittura

type writer struct{ b []byte }

func (w *writer) u8(v byte)   { w.b = append(w.b, v) }
func (w *writer) i32(v int32) { w.b = binary.LittleEndian.AppendUint32(w.b, uint32(v)) }
func (w *writer) lps(s string) {
	n := len(s)
	for {
		c := byte(n & 0x7f)
		n >>= 7
		if n > 0 {
			w.u8(c | 0x80)
		} else {
			w.u8(c)
			break
		}
	}
	w.b = append(w.b, s...)
}
func (w *writer) addOne(bt byte, a interface{}) {
	switch bt {
	case 0, 7:
		w.u8(a.(byte))
	case 3:
		w.lps(a.(string))
	case 4:
		c := a.(ClassTypeInfo)
		w.lps(c.Name)
		w.i32(c.Lib)
	}
}

func (d *Doc) Dump() []byte {
	w := &writer{}
	for _, r := range d.Recs {
		d.wr(w, r)
	}
	return w.b
}

func (d *Doc) wval(w *writer, v Val) {
	if v.Prim {
		w.b = append(w.b, v.Raw...)
	} else {
		d.wr(w, v.R)
	}
}

func (d *Doc) wr(w *writer, r *Rec) {
	w.u8(r.T)
	switch r.T {
	case 0:
		for _, x := range r.Hdr {
			w.i32(x)
		}
	case 12:
		w.i32(r.ID)
		w.lps(r.Name)
	case 5, 4, 3, 2:
		w.i32(r.ID)
		w.lps(r.Name)
		w.i32(int32(len(r.Names)))
		for _, n := range r.Names {
			w.lps(n)
		}
		if r.T == 5 || r.T == 4 {
			for _, b := range r.BTs {
				w.u8(b)
			}
			for i, b := range r.BTs {
				w.addOne(b, r.Add[i])
			}
		}
		if r.T == 5 || r.T == 3 {
			w.i32(r.Lib)
		}
		for _, v := range r.Vals {
			d.wval(w, v)
		}
	case 1:
		w.i32(r.ID)
		w.i32(r.MID)
		for _, v := range r.Vals {
			d.wval(w, v)
		}
	case 6:
		w.i32(r.ID)
		w.lps(r.S)
	case 7:
		w.i32(r.ID)
		w.u8(r.AT)
		w.i32(int32(len(r.Lens)))
		for _, l := range r.Lens {
			w.i32(l)
		}
		for _, l := range r.LB {
			w.i32(l)
		}
		w.u8(r.BT)
		w.addOne(r.BT, r.Add1)
		for _, e := range r.Els {
			d.wval(w, e)
		}
	case 8:
		w.u8(r.PT)
		w.b = append(w.b, r.PRaw...)
	case 9:
		w.i32(r.Ref)
	case 13:
		w.u8(byte(r.N))
	case 14:
		w.i32(r.N)
	case 15:
		w.i32(r.ID)
		w.i32(int32(len(r.PEls)))
		w.u8(r.PT)
		for _, e := range r.PEls {
			w.b = append(w.b, e...)
		}
	case 16, 17:
		w.i32(r.ID)
		tot := 0
		for _, e := range r.Els {
			if e.R.T == 13 || e.R.T == 14 {
				tot += int(e.R.N)
			} else {
				tot++
			}
		}
		w.i32(int32(tot))
		for _, e := range r.Els {
			d.wr(w, e.R)
		}
	}
}

// ---------------- utilità

func (d *Doc) Resolve(v Val) *Rec {
	if v.Prim || v.R == nil {
		return nil
	}
	if v.R.T == 9 {
		return d.Objs[v.R.Ref]
	}
	return v.R
}

func (d *Doc) MetaOf(r *Rec) *Rec {
	if r.T == 1 {
		return d.Meta[r.MID]
	}
	return r
}

func (d *Doc) NewString(s string) Val {
	d.maxID++
	return Val{R: &Rec{T: 6, ID: d.maxID, S: s}}
}

func (d *Doc) StrOf(v Val) string {
	r := d.Resolve(v)
	if r != nil && r.T == 6 {
		return r.S
	}
	return ""
}

func walk(r *Rec, f func(*Rec)) {
	if r == nil {
		return
	}
	f(r)
	for _, v := range r.Vals {
		if !v.Prim {
			walk(v.R, f)
		}
	}
	for _, v := range r.Els {
		if !v.Prim {
			walk(v.R, f)
		}
	}
}

// FixDangling: dopo le modifiche, un riferimento può puntare a una stringa
// che era definita dentro un valore sostituito. La sostituisce con una copia.
func (d *Doc) FixDangling(old map[int32]*Rec) error {
	defined := map[int32]bool{}
	for _, r := range d.Recs {
		walk(r, func(x *Rec) {
			switch x.T {
			case 1, 2, 3, 4, 5, 6, 7, 15, 16, 17:
				defined[x.ID] = true
			}
		})
	}
	var err error
	fix := func(vals []Val) {
		for i, v := range vals {
			if v.Prim || v.R == nil || v.R.T != 9 || defined[v.R.Ref] {
				continue
			}
			o := old[v.R.Ref]
			if o == nil || o.T != 6 {
				err = fmt.Errorf("riferimento %d non risolvibile", v.R.Ref)
				continue
			}
			nv := d.NewString(o.S)
			vals[i] = nv
			defined[nv.R.ID] = true
		}
	}
	for _, r := range d.Recs {
		walk(r, func(x *Rec) { fix(x.Vals); fix(x.Els) })
	}
	return err
}
