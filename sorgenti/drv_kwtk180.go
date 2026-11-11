// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
//
// Derivato dal driver di CHIRP (https://chirpmyradio.com), GPL-3.0-or-later:
//   chirp/drivers/tk8180.py   Copyright 2019 Dan Smith <dsmith@danplanet.com>
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.21: Kenwood TK-2180 / TK-3180 (palmari) e TK-7180 / TK-8180 (veicolari),
// tk8180.py di CHIRP.
//
// La memoria non ha posizioni fisse: le zone stanno una dopo l'altra da 0x0B00
// (tabella degli indirizzi a 0x0A00), ognuna con un'intestazione di 32 byte e
// tanti canali di 48 byte quanti ne contiene; ogni canale porta il suo numero
// (1-250) dentro la zona. Aggiungere o togliere un canale sposta tutte le zone
// che seguono (expand_mmap di CHIRP, rifatto qui uguale byte per byte). I bit di
// esclusione dalla scansione stanno a 0xC570, uno per canale nell'ordine della
// memoria; ogni zona ricorda da quale bit parte (flagoffset, da 1).
// CHIRP mostra ogni zona come una radio a parte; qui la lista ha 250 posizioni
// per ognuna delle prime 32 zone: la posizione è il numero del canale nella
// zona. Le zone non si creano e non si cancellano da qui; in tutto la radio
// tiene 512 canali.
//
// 9600 baud 8N2: «PROGRAM» → 16; si passa a 19200 baud → 06; 02 → sigla di 8
// byte; 06 → 06. Blocchi di 256 byte da 0x00 a 0xBF: 'R' 00 n → 'W' dati somma
// (oppure 'Z' e un byte), 06 → 06; in scrittura 'W' 00 n dati somma → 06 ('Z' 00
// n FF per un blocco tutto a FF). Da 0xC000 a 0xD0FF a pezzi di 64 byte con
// l'indirizzo meno 0xBF00: 'S' indirizzo(2) 40 → 'X' e 64 byte, 06 → 06; 'X'
// indirizzo(2) 40 dati → 06. 'E' → 06 per uscire.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	kwx180Size    = 0xD100
	kwx180Starts  = 0x0A00
	kwx180Sys     = 0x0B00
	kwx180Hdr     = 0x20
	kwx180Mem     = 0x30
	kwx180Skip    = 0xC570
	kwx180Max     = 512 // canali in tutta la radio (bit di scansione)
	kwx180PerZone = 250
	kwx180Zones   = 32 // zone mostrate
	kwx180ID      = 0x0120
	kwx180Blocks  = 0xC0 // blocchi da 256 byte letti per numero

	kwx180Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!\"#$%&'()~+-,./:;<=>?@[\\]^`{}*| "
)

type kwx180Fam struct{}

type kwx180Opt struct{ id string }

type kwx180Zone struct{ addr, count int }

func init() {
	mk := func(id, model, class, ident string, lo, hi int64, pw []genPower, what string) *genModel {
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: kwx180Zones * kwx180PerZone, First: 1, NameMax: 12, Charset: kwx180Charset, Zones: kwx180Zones, Powers: pw,
			Bands: []band{{lo, hi}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kwx180Fam{}, Opt: &kwx180Opt{id: ident},
			Notes: what + " La lista ha 250 posizioni per ognuna delle prime 32 zone della radio: la posizione è il numero del canale " +
				"nella zona. Si possono usare solo le zone che la radio ha già (non si creano e non si cancellano da qui) e in tutto " +
				"la radio tiene 512 canali. PTT ID e le altre opzioni di un canale esistente si conservano. Le zone di tipo non " +
				"convenzionale (trunking) non sono gestite: in quel caso la lettura viene rifiutata. Driver di CHIRP."}
	}
	hand := "Radio professionale portatile: si programma col cavo di programmazione Kenwood."
	mob := "Radio professionale veicolare: si programma col cavo di programmazione Kenwood."
	mobPw := []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}}
	genRegister(
		mk("kw-tk2180", "TK-2180", "KenwoodTK2180Radio", "P2180\x04", 136000000, 174000000, kwtkPowers, hand),
		mk("kw-tk3180k", "TK-3180K", "KenwoodTK3180K1Radio", "P3180\x06", 400000000, 520000000, kwtkPowers, hand),
		mk("kw-tk3180k2", "TK-3180K2", "KenwoodTK3180K2Radio", "P3180\x07", 400000000, 520000000, kwtkPowers, hand),
		mk("kw-tk3180e", "TK-3180E", "KenwoodTK3180ERadio", "P3189\x27", 400000000, 520000000, kwtkPowers, hand),
		mk("kw-tk7180", "TK-7180", "KenwoodTK7180Radio", "M7180\x04", 136000000, 174000000, mobPw, mob),
		mk("kw-tk8180", "TK-8180", "KenwoodTK8180Radio", "M8180\x06", 400000000, 520000000, mobPw, mob),
		mk("kw-tk7180e", "TK-7180E", "KenwoodTK7180ERadio", "M7189$", 136000000, 174000000, mobPw, mob),
		mk("kw-tk8180e", "TK-8180E", "KenwoodTK8180E", "M8189'", 400000000, 520000000, mobPw, mob),
	)
}

func kwx180o(m *genModel) *kwx180Opt { return m.Opt.(*kwx180Opt) }

// zone presenti (probe_layout di CHIRP); ok è falso se la memoria non è
// leggibile o ha una zona non convenzionale
func (kwx180Fam) zones(img []byte) (zs []kwx180Zone, ok bool) {
	if len(img) < kwx180Size {
		return nil, false
	}
	for i := 0; i < 128; i++ {
		a := int(binary.LittleEndian.Uint16(img[kwx180Starts+2*i:]))
		if a == 0xFFFF {
			break
		}
		if a < kwx180Sys || a+kwx180Hdr > kwx180Blocks*256 || img[a+1] != 0x31 {
			return nil, false
		}
		n := int(img[a+4])
		if a+kwx180Hdr+n*kwx180Mem > kwx180Blocks*256 {
			return nil, false
		}
		zs = append(zs, kwx180Zone{a, n})
	}
	return zs, true
}

func (z kwx180Zone) rec(img []byte, k int) []byte {
	o := z.addr + kwx180Hdr + k*kwx180Mem
	return img[o : o+kwx180Mem]
}

// posizione nella zona del canale numero n, -1 se non c'è
func (z kwx180Zone) find(img []byte, n int) int {
	for k := 0; k < z.count; k++ {
		if int(z.rec(img, k)[0]) == n {
			return k
		}
	}
	return -1
}

func (z kwx180Zone) flagOff(img []byte) int { return int(binary.LittleEndian.Uint16(img[z.addr+2:])) }

// bit di esclusione dalla scansione del canale in posizione k di una zona che
// parte dal bit off (da 1); fuori dalla tabella non c'è
func kwx180SkipGet(img []byte, off, k int) (skip, ok bool) {
	bit := off + k - 1
	if bit < 0 || bit >= kwx180Max {
		return false, false
	}
	return kwtkBit(img, kwx180Skip, bit), true
}

func kwx180SkipSet(img []byte, off, k int, skip bool) {
	if bit := off + k - 1; bit >= 0 && bit < kwx180Max {
		kwtkSetBit(img, kwx180Skip, bit, skip)
	}
}

func kwx180Bcd(b []byte) int64 {
	var v, mul int64 = 0, 1
	for _, x := range b {
		v += (int64(x>>4)*10 + int64(x&0x0F)) * mul
		mul *= 100
	}
	return v
}

func kwx180PutBcd(b []byte, v int64) {
	for k := range b {
		d := v % 100
		b[k] = byte(d/10<<4 | d%10)
		v /= 100
	}
}

// passo della frequenza come in CHIRP: il primo, nell'ordine 12,5 - 10 - 6,25 -
// 5 - 2,5 kHz, di cui la frequenza è multipla; 5 kHz se nessuno
func kwx180Step(hz int64) byte {
	switch {
	case hz%12500 == 0:
		return 6
	case hz%10000 == 0:
		return 5
	case hz%6250 == 0:
		return 3
	case hz%5000 == 0:
		return 2
	case hz%2500 == 0:
		return 1
	}
	return 2
}

func kwx180ToneEnc(t genTone) uint16 {
	switch t.Mode {
	case "Tone":
		return uint16(t.Hz10)
	case "DTCS":
		var v uint16
		fmt.Sscanf(fmt.Sprintf("%d", t.Code), "%o", &v)
		v |= 0x2800
		if t.Inv {
			v |= 0x8000
		}
		return v
	}
	return 0xFFFF
}

func (f kwx180Fam) Get(m *genModel, img []byte, i int) *Chan {
	zs, ok := f.zones(img)
	z, n := i/kwx180PerZone, i%kwx180PerZone+1
	if !ok || z >= len(zs) {
		return nil
	}
	k := zs[z].find(img, n)
	if k < 0 {
		return nil
	}
	r := zs[z].rec(img, k)
	c := &Chan{Rx: kwx180Bcd(r[1:5]) * 10}
	c.Tx = c.Rx
	if bytes.Equal(r[5:9], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = kwx180Bcd(r[5:9]) * 10
	}
	c.RxTone = kwx140ToneDec(int64(binary.LittleEndian.Uint16(r[11:])))
	c.TxTone = kwx140ToneDec(int64(binary.LittleEndian.Uint16(r[13:])))
	c.Name = strings.TrimRight(string(r[15:27]), "\x00")
	c.Power = m.powerSem(int(r[46] >> 3 & 1))
	c.Narrow = r[46]>>1&1 == 0
	skip, _ := kwx180SkipGet(img, zs[z].flagOff(img), k)
	c.Scan = !skip
	return c
}

// Rec / PutRec: il record del canale presente in quella posizione (vuoto se non c'è)
func (f kwx180Fam) Rec(m *genModel, img []byte, i int) []byte {
	if zs, ok := f.zones(img); ok && i/kwx180PerZone < len(zs) {
		if k := zs[i/kwx180PerZone].find(img, i%kwx180PerZone+1); k >= 0 {
			return append([]byte{}, zs[i/kwx180PerZone].rec(img, k)...)
		}
	}
	return bytes.Repeat([]byte{0xFF}, kwx180Mem)
}

func (f kwx180Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if zs, ok := f.zones(img); ok && i/kwx180PerZone < len(zs) && len(raw) == kwx180Mem {
		if k := zs[i/kwx180PerZone].find(img, i%kwx180PerZone+1); k >= 0 {
			n := zs[i/kwx180PerZone].rec(img, k)[0]
			copy(zs[i/kwx180PerZone].rec(img, k), raw)
			zs[i/kwx180PerZone].rec(img, k)[0] = n
		}
	}
}

func (f kwx180Fam) Match(m *genModel, img []byte) bool {
	return len(img) == kwx180Size && string(img[kwx180ID:kwx180ID+len(kwx180o(m).id)]) == kwx180o(m).id
}

// posizioni delle zone che la radio non ha: non si possono usare
func (f kwx180Fam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	zs, ok := f.zones(img)
	return !ok || i/kwx180PerZone >= len(zs)
}

func (f kwx180Fam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	if want == nil {
		return nil
	}
	zs, ok := f.zones(img)
	if !ok {
		return fmt.Errorf("la memoria ha zone non convenzionali: non si può modificare")
	}
	if i/kwx180PerZone >= len(zs) {
		return fmt.Errorf("la radio ha %d zone: la zona %d va creata con il programma del costruttore", len(zs), i/kwx180PerZone+1)
	}
	return nil
}

// shuffle_zone di CHIRP: i canali della zona in ordine di numero, con i loro
// bit di scansione
func (f kwx180Fam) shuffle(img []byte, z kwx180Zone) {
	idx := make([]int, z.count)
	for k := range idx {
		idx[k] = k
	}
	sort.SliceStable(idx, func(a, b int) bool { return z.rec(img, idx[a])[0] < z.rec(img, idx[b])[0] })
	same := true
	for k, v := range idx {
		same = same && k == v
	}
	if same {
		return
	}
	off := z.flagOff(img)
	raw := make([][]byte, z.count)
	type flag struct{ v, ok bool }
	skip := make([]flag, z.count)
	for k, src := range idx {
		raw[k] = append([]byte{}, z.rec(img, src)...)
		skip[k].v, skip[k].ok = kwx180SkipGet(img, off, src)
	}
	for k := range idx {
		copy(z.rec(img, k), raw[k])
		if skip[k].ok {
			kwx180SkipSet(img, off, k, skip[k].v)
		}
	}
}

// expand_mmap di CHIRP: le zone con il nuovo numero di canali, a partire dalla
// memoria attuale (quello che resta oltre la fine non viene toccato)
func (f kwx180Fam) relayout(img []byte, old []kwx180Zone, counts []int) []kwx180Zone {
	src := append([]byte{}, img...)
	zs := make([]kwx180Zone, len(counts))
	addr := kwx180Sys
	for k, n := range counts {
		zs[k] = kwx180Zone{addr, n}
		addr += kwx180Hdr + n*kwx180Mem
	}
	for k := 0; k < 128; k++ {
		v := uint16(0xFFFF)
		if k < len(zs) {
			v = uint16(zs[k].addr)
		}
		binary.LittleEndian.PutUint16(img[kwx180Starts+2*k:], v)
	}
	scan := 1
	for k, z := range zs {
		o := old[k]
		copy(img[z.addr:z.addr+kwx180Hdr], src[o.addr:o.addr+kwx180Hdr])
		img[z.addr+4] = byte(z.count)
		oldOff := o.flagOff(src)
		for j := 0; j < z.count && j < o.count; j++ {
			copy(z.rec(img, j), o.rec(src, j))
			if v, ok := kwx180SkipGet(src, oldOff, j); ok {
				kwx180SkipSet(img, scan, j, v)
			}
		}
		binary.LittleEndian.PutUint16(img[z.addr+2:], uint16(scan))
		scan += z.count
	}
	return zs
}

func (f kwx180Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	zs, ok := f.zones(img)
	z, n := i/kwx180PerZone, i%kwx180PerZone+1
	if !ok {
		return fmt.Errorf("la memoria ha zone non convenzionali: non si può modificare")
	}
	if z >= len(zs) {
		if c == nil {
			return nil
		}
		return fmt.Errorf("la radio ha %d zone: la zona %d va creata con il programma del costruttore", len(zs), z+1)
	}
	counts := make([]int, len(zs))
	total := 0
	for k, x := range zs {
		counts[k] = x.count
		total += x.count
	}
	k := zs[z].find(img, n)
	if c == nil {
		if k < 0 {
			return nil
		}
		// come CHIRP: il canale in fondo alla zona, poi la zona più corta di uno
		zs[z].rec(img, k)[0] = 0xFF
		f.shuffle(img, zs[z])
		counts[z]--
		f.relayout(img, zs, counts)
		return nil
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	if k < 0 {
		if total >= kwx180Max {
			return fmt.Errorf("la radio tiene al massimo %d canali in tutto", kwx180Max)
		}
		// come CHIRP: la zona più lunga di uno, il canale nuovo in fondo e poi al
		// suo posto in ordine di numero
		counts[z]++
		zs = f.relayout(img, zs, counts)
		r := zs[z].rec(img, zs[z].count-1)
		r[0] = byte(n)
		f.shuffle(img, zs[z])
		k = zs[z].find(img, n)
		r = zs[z].rec(img, k)
		// valori predefiniti delle voci non note
		r[9], r[10] = 0x30, 0x30
		r[27], r[28], r[29] = 0xFF, 0xFF, 0xFF
		copy(r[30:46], bytes.Repeat([]byte{0xFF}, 16))
		r[46], r[47] = 0xF4, 0xFF
	}
	r := zs[z].rec(img, k)
	name := make([]byte, 12)
	copy(name, strings.TrimRight(c.Name, " \t\r\n\v\f"))
	copy(r[15:27], name)
	kwx180PutBcd(r[1:5], c.Rx/10)
	binary.LittleEndian.PutUint16(r[13:], kwx180ToneEnc(tt))
	binary.LittleEndian.PutUint16(r[11:], kwx180ToneEnc(rt))
	r[46] &^= 0x0A
	if !c.Narrow {
		r[46] |= 0x02
	}
	if m.powerIndex(c.Power) == 1 {
		r[46] |= 0x08
	}
	txStep := byte(2) // CHIRP calcola il passo anche sul campo a FF: ne esce 5 kHz
	if c.TxOff {
		copy(r[5:9], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		kwx180PutBcd(r[5:9], c.Tx/10)
		txStep = kwx180Step(c.Tx / 10 * 10)
	}
	r[9] = r[9]&0xF0 | kwx180Step(c.Rx/10*10)
	r[10] = r[10]&0xF0 | txStep
	kwx180SkipSet(img, zs[z].flagOff(img), k, !c.Scan)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (f kwx180Fam) open(ctx context.Context, m *genModel, port string) (*rawSess, error) {
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (*rawSess, error) {
		// come CHIRP: 'E' alle due velocità
		_ = s.p.SetBaud(9600)
		s.send([]byte{'E'})
		time.Sleep(50 * time.Millisecond)
		_ = s.p.SetBaud(19200)
		s.send([]byte{'E'})
		s.p.Close()
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		return nil, genFail("Identificazione", e)
	}
	if err := setTwoStopBits(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	s.p.Purge()
	if err := s.send([]byte("PROGRAM")); err != nil {
		s.p.Close()
		return nil, err
	}
	b := readUpTo(ctx, s.p, 1, time.Second)
	if len(b) == 0 {
		return fail(errTimeout)
	}
	if b[0] != 0x16 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione veloce"))
	}
	if err := s.p.SetBaud(19200); err != nil {
		return fail(err)
	}
	b = readUpTo(ctx, s.p, 1, time.Second)
	if len(b) == 0 {
		return fail(errTimeout)
	}
	if b[0] != 0x06 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
	}
	s.send([]byte{0x02})
	id := readUpTo(ctx, s.p, 8, time.Second)
	if len(id) != 8 {
		return fail(errTimeout)
	}
	s.send([]byte{0x06})
	b = readUpTo(ctx, s.p, 1, time.Second)
	if len(b) == 0 {
		return fail(errTimeout)
	}
	if b[0] != 0x06 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
	}
	if string(id[:6]) != kwx180o(m).id {
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwx180Opt); ok && xo.id == string(id[:6]) {
				return fail(rErr("la radio è un altro modello: apri la scheda «%s»", x.Name))
			}
		}
		return fail(rErr("la radio è un altro modello (%s)", strings.TrimRight(string(id[:5]), "\xff\x00")))
	}
	return s, nil
}

func (f kwx180Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, kwx180Size, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	fail := func(e error) ([]byte, string, error) {
		s.send([]byte{'E'})
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		return nil, "", e
	}
	var data []byte
	for blk := 0; blk < kwx180Blocks; blk++ {
		prog(len(data), kwx180Size, "Lettura…")
		if err := s.send([]byte{'R', 0, byte(blk)}); err != nil {
			return nil, "", err
		}
		c := readUpTo(ctx, s.p, 1, time.Second)
		switch {
		case len(c) != 1:
			return fail(genFail("Lettura", errTimeout))
		case c[0] == 'Z':
			if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 {
				return fail(genFail("Lettura", errTimeout))
			}
			data = append(data, bytes.Repeat([]byte{0xFF}, 256)...)
		case c[0] == 'W':
			b := readUpTo(ctx, s.p, 257, 2*time.Second)
			if len(b) != 257 {
				return fail(rErr("lettura del blocco 0x%04X non riuscita", blk*256))
			}
			if kwtkSum(b[:256]) != b[256] {
				return fail(rErr("somma di controllo errata: ripeti la lettura"))
			}
			data = append(data, b[:256]...)
		default:
			return fail(rErr("risposta inattesa della radio"))
		}
		if err := s.ack2(); err != nil {
			return fail(err)
		}
	}
	for a := 0x0100; a < 0x1200; a += 0x40 {
		prog(len(data), kwx180Size, "Lettura…")
		if err := s.send([]byte{'S', byte(a >> 8), byte(a), 0x40}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x41, time.Second)
		if len(b) != 0x41 || b[0] != 'X' {
			return fail(rErr("lettura del blocco 0x%04X non riuscita", a+0xBF00))
		}
		data = append(data, b[1:]...)
		if err := s.ack2(); err != nil {
			return fail(err)
		}
	}
	s.send([]byte{'E'})
	if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return fail(rErr("la radio non ha confermato la fine della lettura"))
	}
	if _, ok := f.zones(data); !ok {
		return nil, "", rErr("la radio ha zone non convenzionali (trunking): non sono gestite")
	}
	prog(kwx180Size, kwx180Size, "Lettura completata")
	return data, "", nil
}

func (f kwx180Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != kwx180Size {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, kwx180Size, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	ack := func(a int) error {
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) != 1 || b[0] != 0x06 {
			s.send([]byte{'E'})
			return rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
		return nil
	}
	for blk := 0; blk < kwx180Blocks; blk++ {
		prog(blk*256, kwx180Size, "Scrittura…")
		chunk := img[blk*256 : blk*256+256]
		frame := []byte{'Z', 0, byte(blk), 0xFF}
		if !bytes.Equal(chunk, bytes.Repeat([]byte{0xFF}, 256)) {
			frame = append([]byte{'W', 0, byte(blk)}, chunk...)
			frame = append(frame, kwtkSum(chunk))
		}
		if err := s.send(frame); err != nil {
			return "", err
		}
		if err := ack(blk * 256); err != nil {
			return "", err
		}
	}
	for a := kwx180Blocks * 256; a < kwx180Size; a += 0x40 {
		prog(a, kwx180Size, "Scrittura…")
		b := a - 0xBF00
		frame := append([]byte{'X', byte(b >> 8), byte(b), 0x40}, img[a:a+0x40]...)
		if err := s.send(frame); err != nil {
			return "", err
		}
		if err := ack(a); err != nil {
			return "", err
		}
	}
	s.send([]byte{'E'})
	if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return "", rErr("la radio non ha confermato la fine della scrittura")
	}
	prog(kwx180Size, kwx180Size, "Scrittura completata")
	return "", nil
}
