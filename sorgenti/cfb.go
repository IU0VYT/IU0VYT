// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura e scrittura minima del formato "Compound File Binary" (OLE2),
// il contenitore dei file Excel .xls.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
)

const (
	cfbFree    = 0xFFFFFFFF
	cfbEnd     = 0xFFFFFFFE
	cfbFatSect = 0xFFFFFFFD
	cfbDifSect = 0xFFFFFFFC
)

var cfbSig = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

type cfbFile struct {
	data     []byte
	secSize  int
	miniSize int
	cutoff   uint32
	fat      []uint32
	miniFat  []uint32
	dirs     []cfbDir
	miniStrm []byte
}

type cfbDir struct {
	name  string
	typ   byte
	start uint32
	size  uint64
}

func isCFB(b []byte) bool { return len(b) >= 512 && string(b[:8]) == string(cfbSig) }

func (f *cfbFile) sector(id uint32) ([]byte, error) {
	off := (int64(id) + 1) * int64(f.secSize)
	if id >= cfbDifSect || off+int64(f.secSize) > int64(len(f.data)) {
		// l'ultimo settore può essere troncato in file scritti male
		if id < cfbDifSect && off < int64(len(f.data)) {
			s := make([]byte, f.secSize)
			copy(s, f.data[off:])
			return s, nil
		}
		return nil, fmt.Errorf("settore %d fuori dal file", id)
	}
	return f.data[off : off+int64(f.secSize)], nil
}

func (f *cfbFile) chain(start uint32, table []uint32) ([]uint32, error) {
	var out []uint32
	seen := map[uint32]bool{}
	for id := start; id != cfbEnd; {
		if id >= uint32(len(table)) || seen[id] {
			if id == cfbFree && len(out) > 0 {
				break
			}
			return nil, errors.New("catena di settori non valida")
		}
		seen[id] = true
		out = append(out, id)
		id = table[id]
	}
	return out, nil
}

// readChain legge una catena di settori; size < 0 = tutta la catena.
func (f *cfbFile) readChain(start uint32, size int64) ([]byte, error) {
	if start == cfbEnd || start == cfbFree || size == 0 {
		return nil, nil
	}
	ids, err := f.chain(start, f.fat)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(ids)*f.secSize)
	for _, id := range ids {
		s, err := f.sector(id)
		if err != nil {
			return nil, err
		}
		out = append(out, s...)
	}
	if size > 0 && int64(len(out)) > size {
		out = out[:size]
	}
	return out, nil
}

func openCFB(b []byte) (*cfbFile, error) {
	if !isCFB(b) {
		return nil, errors.New("non è un file OLE2/CFB")
	}
	le := binary.LittleEndian
	f := &cfbFile{data: b}
	shift := le.Uint16(b[30:])
	if shift != 9 && shift != 12 {
		return nil, errors.New("dimensione settore non supportata")
	}
	f.secSize = 1 << shift
	f.miniSize = 1 << le.Uint16(b[32:])
	nFat := le.Uint32(b[44:])
	firstDir := le.Uint32(b[48:])
	f.cutoff = le.Uint32(b[56:])
	firstMini := le.Uint32(b[60:])
	nMini := le.Uint32(b[64:])
	firstDif := le.Uint32(b[68:])
	nDif := le.Uint32(b[72:])
	var fatIDs []uint32
	for i := 0; i < 109 && uint32(len(fatIDs)) < nFat; i++ {
		fatIDs = append(fatIDs, le.Uint32(b[76+4*i:]))
	}
	dif := firstDif
	for k := uint32(0); k < nDif && dif < cfbDifSect && uint32(len(fatIDs)) < nFat; k++ {
		s, err := f.sector(dif)
		if err != nil {
			return nil, err
		}
		per := f.secSize/4 - 1
		for i := 0; i < per && uint32(len(fatIDs)) < nFat; i++ {
			fatIDs = append(fatIDs, le.Uint32(s[4*i:]))
		}
		dif = le.Uint32(s[4*per:])
	}
	for _, id := range fatIDs {
		s, err := f.sector(id)
		if err != nil {
			return nil, err
		}
		for i := 0; i < f.secSize/4; i++ {
			f.fat = append(f.fat, le.Uint32(s[4*i:]))
		}
	}
	dirData, err := f.readChain(firstDir, -1)
	if err != nil {
		return nil, fmt.Errorf("directory: %v", err)
	}
	for off := 0; off+128 <= len(dirData); off += 128 {
		e := dirData[off : off+128]
		nl := int(le.Uint16(e[64:]))
		if nl > 64 {
			nl = 64
		}
		u := make([]uint16, 0, 32)
		for i := 0; i+1 < nl; i += 2 {
			c := le.Uint16(e[i:])
			if c == 0 {
				break
			}
			u = append(u, c)
		}
		d := cfbDir{name: string(utf16.Decode(u)), typ: e[66], start: le.Uint32(e[116:]), size: le.Uint64(e[120:])}
		if f.secSize == 512 {
			d.size &= 0xFFFFFFFF
		}
		f.dirs = append(f.dirs, d)
	}
	if len(f.dirs) == 0 || f.dirs[0].typ != 5 {
		return nil, errors.New("radice del file non trovata")
	}
	if nMini > 0 && firstMini < cfbDifSect {
		mf, err := f.readChain(firstMini, -1)
		if err == nil {
			for i := 0; i+4 <= len(mf); i += 4 {
				f.miniFat = append(f.miniFat, le.Uint32(mf[i:]))
			}
		}
		f.miniStrm, _ = f.readChain(f.dirs[0].start, int64(f.dirs[0].size))
	}
	return f, nil
}

func (f *cfbFile) stream(name string) ([]byte, error) {
	for _, d := range f.dirs {
		if d.typ == 2 && strings.EqualFold(d.name, name) {
			if d.size < uint64(f.cutoff) {
				ids, err := f.chain(d.start, f.miniFat)
				if err != nil {
					return nil, err
				}
				out := make([]byte, 0, len(ids)*f.miniSize)
				for _, id := range ids {
					o := int(id) * f.miniSize
					if o+f.miniSize > len(f.miniStrm) {
						return nil, errors.New("mini stream non valido")
					}
					out = append(out, f.miniStrm[o:o+f.miniSize]...)
				}
				if uint64(len(out)) > d.size {
					out = out[:d.size]
				}
				return out, nil
			}
			return f.readChain(d.start, int64(d.size))
		}
	}
	return nil, fmt.Errorf("flusso %s non trovato", name)
}

// writeCFB crea un file OLE2 con un solo flusso (es. "Workbook").
// Il flusso viene portato ad almeno 4096 byte, come fanno altri programmi,
// così non serve il "mini stream".
func writeCFB(name string, stream []byte) ([]byte, error) {
	if len(stream) < 4096 {
		stream = append(stream, make([]byte, 4096-len(stream))...)
	}
	le := binary.LittleEndian
	const ss = 512
	nData := (len(stream) + ss - 1) / ss
	nFat := 1
	for (nData + 1 + nFat) > nFat*(ss/4) {
		nFat++
	}
	if nFat > 109 {
		return nil, errors.New("file troppo grande per il formato .xls")
	}
	dirSec := nData
	fatStart := nData + 1
	total := nData + 1 + nFat
	out := make([]byte, ss*(1+total))
	h := out[:ss]
	copy(h, cfbSig)
	le.PutUint16(h[24:], 0x003E)
	le.PutUint16(h[26:], 0x0003)
	le.PutUint16(h[28:], 0xFFFE)
	le.PutUint16(h[30:], 9)
	le.PutUint16(h[32:], 6)
	le.PutUint32(h[44:], uint32(nFat))
	le.PutUint32(h[48:], uint32(dirSec))
	le.PutUint32(h[56:], 4096)
	le.PutUint32(h[60:], cfbEnd)
	le.PutUint32(h[68:], cfbEnd)
	for i := 0; i < 109; i++ {
		v := uint32(cfbFree)
		if i < nFat {
			v = uint32(fatStart + i)
		}
		le.PutUint32(h[76+4*i:], v)
	}
	copy(out[ss:], stream)
	// directory
	dir := out[ss*(1+dirSec) : ss*(2+dirSec)]
	putDir := func(e []byte, n string, typ byte, child, start uint32, size uint64) {
		u := utf16.Encode([]rune(n))
		for i, c := range u {
			le.PutUint16(e[2*i:], c)
		}
		le.PutUint16(e[64:], uint16(2*len(u)+2))
		e[66] = typ
		e[67] = 1 // nero
		le.PutUint32(e[68:], cfbFree)
		le.PutUint32(e[72:], cfbFree)
		le.PutUint32(e[76:], child)
		le.PutUint32(e[116:], start)
		le.PutUint64(e[120:], size)
	}
	putDir(dir[0:128], "Root Entry", 5, 1, cfbEnd, 0)
	putDir(dir[128:256], name, 2, cfbFree, 0, uint64(len(stream)))
	for k := 2; k < 4; k++ {
		e := dir[128*k : 128*(k+1)]
		le.PutUint32(e[68:], cfbFree)
		le.PutUint32(e[72:], cfbFree)
		le.PutUint32(e[76:], cfbFree)
	}
	// FAT
	fat := make([]uint32, nFat*(ss/4))
	for i := range fat {
		fat[i] = cfbFree
	}
	for i := 0; i < nData; i++ {
		fat[i] = uint32(i + 1)
	}
	fat[nData-1] = cfbEnd
	fat[dirSec] = cfbEnd
	for i := 0; i < nFat; i++ {
		fat[fatStart+i] = cfbFatSect
	}
	for i, v := range fat {
		le.PutUint32(out[ss*(1+fatStart)+4*i:], v)
	}
	return out, nil
}
