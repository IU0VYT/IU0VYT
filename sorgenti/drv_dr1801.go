// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Baofeng DR-1801UV (piattaforma Auctus A6, «BF1801»): protocollo e memoria
// ricavati dal programma del costruttore «DR-1801A6» (.NET, pacchetto
// «Baofeng_DR-1801UV_Programming_Software»; classi Auctus.BF1801.Cps.Protocol,
// Core.PipeLine e Data.DataLayout) e dalla libreria Auctus.NetStream.
//
// Pacchetti: AA lunghezza comando(2, little endian) dati somma BB; lunghezza =
// 6 + dati; somma = XOR di lunghezza, byte del comando e dati. Le risposte
// hanno il comando con il bit 15 a 1 scritto in big endian (es. 81 04).
//   9600 baud:  01 04 (collegamento)          → 81 04 [01 = pronta]
//               00 2B [n, password]            → 80 2B [0 = password errata, 1/2 = ok]
//   lettura:    01 00 [velocità, 4 byte BE]    → 81 00 [01, lunghezza BE4,
//                                                 versione dei dati BE4 (104),
//                                                 versione dell'hardware BE4, CRC 2]
//               la porta si riapre a 115200 baud (pausa di 1 s), 01 01 → tutti i
//               dati, senza pacchetto; CRC = XOR dei byte pari e dei byte dispari
//   scrittura:  01 02 [versione BE2 (1), lunghezza BE4, CRC 2, velocità BE4] → 81 02
//               115200 baud (pausa di 1 s), dati a blocchi di 200 byte ogni 30 ms
//               → 81 03 [01, errore 2]
// La lettura restituisce il file completo (intestazione «AUCTUSA6», come i file
// .accps del programma); la scrittura invia solo il corpo (dopo «#BD#»).
//
// Memoria: intestazione (lunghezza, versione 104, elenco dei moduli «#MP#»:
// indice, attivo, codice di 2 lettere, posizione nel corpo (3 byte BE),
// parametri) e corpo con i moduli. Si usano:
//   CH canali: parametri MAX(2) lunghezza del nome(1) -(1) dimensione(2)
//      dimensione del nome(2); dati: numero di canali(2), riservato(2), MAX
//      record da 52 byte, MAX nomi (GBK, 0 in coda). Record: 0 numero del canale
//      (FFFF libero), 2 tipo (1 analogico, 2 PDT, 3 DMR), 3 potenza (1 alta),
//      4 RX e 8 TX in Hz, 12 contatto TX, 14 PTT, 15 riservato, 16 color code,
//      17 slot, …, 25 banda (0 = 12,5 kHz, 1 = 25 kHz), 26 scansione, 27 lista,
//      28 tono RX (2) tipo(1: 0 nessuno, 1 CTCSS in decimi di Hz, 2 DCS) invertito(1),
//      32 tono TX come sopra, 36-39 parametri dei CTCSS non standard, 42 gruppi RX…
//   ZD zone, SL liste di scansione, RM roaming: un canale cancellato viene tolto
//      dalle tre liste come fa il programma (le zone vuote vengono eliminate).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	drRecSize   = 52
	drDataVer   = 104
	drWriteVer  = 1
	drLowBaud   = 9600
	drHighBaud  = 115200
	drChunk     = 200
	drChunkWait = 30 * time.Millisecond
)

type drMod struct {
	code   string
	off    int // posizione assoluta dei dati nell'immagine
	params []byte
}

type drLayout struct {
	hl   int
	mods map[string]drMod
}

func drParse(img []byte) (*drLayout, error) {
	if len(img) < 64 || string(img[:8]) != "AUCTUSA6" {
		return nil, fmt.Errorf("dati della radio non riconosciuti (manca l'intestazione AUCTUSA6)")
	}
	hl := int(binary.LittleEndian.Uint16(img[8:]))
	cnt := int(binary.LittleEndian.Uint16(img[16:]))
	mo := int(binary.LittleEndian.Uint16(img[18:]))
	if hl+4 > len(img) || mo+4 > len(img) || string(img[mo:mo+4]) != "#MP#" || string(img[hl:hl+4]) != "#BD#" {
		return nil, fmt.Errorf("struttura dei dati della radio non valida")
	}
	l := &drLayout{hl: hl, mods: map[string]drMod{}}
	o := mo + 4
	for i := 0; i < cnt; i++ {
		if o+8 > len(img) {
			return nil, fmt.Errorf("elenco dei moduli troppo corto")
		}
		code := string(img[o+2 : o+4])
		off := int(img[o+4])<<16 | int(img[o+5])<<8 | int(img[o+6])
		pl := int(img[o+7])
		if o+8+pl > len(img) {
			return nil, fmt.Errorf("elenco dei moduli troppo corto")
		}
		l.mods[code] = drMod{code: code, off: hl + 4 + off, params: append([]byte{}, img[o+8:o+8+pl]...)}
		o += 8 + pl
	}
	for _, c := range []string{"CH", "ZD", "SL"} {
		if _, ok := l.mods[c]; !ok {
			return nil, fmt.Errorf("modulo %s mancante nei dati della radio", c)
		}
	}
	return l, nil
}

// canali
type drCh struct {
	max, nameLen, size, nameSize int
	base                         int // dati del modulo
}

func (l *drLayout) ch() drCh {
	m := l.mods["CH"]
	p := m.params
	return drCh{max: int(binary.LittleEndian.Uint16(p)), nameLen: int(p[2]), size: int(binary.LittleEndian.Uint16(p[4:])),
		nameSize: int(binary.LittleEndian.Uint16(p[6:])), base: m.off}
}

func (c drCh) rec(img []byte, pos int) []byte {
	o := c.base + 4 + pos*c.size
	return img[o : o+c.size]
}

func (c drCh) name(img []byte, pos int) []byte {
	o := c.base + 4 + c.max*c.size + pos*c.nameSize
	return img[o : o+c.nameSize]
}

// posizione del record con il canale idx (come la lettura del programma: si
// scorrono i record finché non si trovano tutti i canali dichiarati)
func (c drCh) find(img []byte, idx int) int {
	avail := int(binary.LittleEndian.Uint16(img[c.base:]))
	n := 0
	for p := 0; p < c.max && n < avail; p++ {
		v := int(binary.LittleEndian.Uint16(c.rec(img, p)))
		if v == 0xFFFF {
			continue
		}
		n++
		if v == idx {
			return p
		}
	}
	return -1
}

func (c drCh) exists(img []byte, idx int) bool { return c.find(img, idx) >= 0 }

func (c drCh) recount(img []byte) {
	n := 0
	for p := 0; p < c.max; p++ {
		if binary.LittleEndian.Uint16(c.rec(img, p)) != 0xFFFF {
			n++
		}
	}
	binary.LittleEndian.PutUint16(img[c.base:], uint16(n))
}

// zone
type drZone struct {
	name     string
	reserved byte
	index    uint16
	channels []uint16
}

type drZones struct {
	max, nameLen, size, maxCh int
	base                      int
	up, down                  uint16
	r1, r2                    byte
	list                      []drZone
}

func (l *drLayout) zones(img []byte) *drZones {
	m := l.mods["ZD"]
	p := m.params
	z := &drZones{max: int(binary.LittleEndian.Uint16(p)), nameLen: int(p[3]), size: int(binary.LittleEndian.Uint16(p[4:])),
		maxCh: int(binary.LittleEndian.Uint16(p[6:])), base: m.off}
	d := img[m.off:]
	cnt := int(binary.LittleEndian.Uint16(d))
	z.up, z.down = binary.LittleEndian.Uint16(d[2:]), binary.LittleEndian.Uint16(d[4:])
	z.r1, z.r2 = d[6], d[7]
	o := 8
	for i := 0; i < cnt && i < z.max; i++ {
		r := d[o : o+z.size]
		zi := drZone{name: drStr(r[:z.nameLen]), reserved: r[z.nameLen+1], index: binary.LittleEndian.Uint16(r[z.nameLen+4:])}
		n := int(binary.LittleEndian.Uint16(r[z.nameLen+2:]))
		for j := 0; j < n && j < z.maxCh; j++ {
			zi.channels = append(zi.channels, binary.LittleEndian.Uint16(r[z.nameLen+8+2*j:]))
		}
		z.list = append(z.list, zi)
		o += z.size
	}
	return z
}

// scrittura come ModuleDataZone.W (zone vuote eliminate e zona corrente
// corretta come in DataLayout.W)
func (z *drZones) write(img []byte) {
	var keep []drZone
	for _, zi := range z.list {
		if len(zi.channels) > 0 {
			keep = append(keep, zi)
		}
	}
	z.list = keep
	has := func(i uint16) bool {
		for _, zi := range z.list {
			if zi.index == i {
				return true
			}
		}
		return false
	}
	first := uint16(0)
	if len(z.list) > 0 {
		first = z.list[0].index
	}
	if !has(z.up) {
		z.up = first
	}
	if !has(z.down) {
		z.down = first
	}
	area := img[z.base : z.base+8+z.max*z.size]
	for i := range area {
		area[i] = 0
	}
	binary.LittleEndian.PutUint16(area, uint16(len(z.list)))
	binary.LittleEndian.PutUint16(area[2:], z.up)
	binary.LittleEndian.PutUint16(area[4:], z.down)
	area[6], area[7] = z.r1, z.r2
	o := 8
	for _, zi := range z.list {
		r := area[o : o+z.size]
		nb := []byte(zi.name)
		if len(nb) > z.nameLen {
			nb = nb[:z.nameLen]
		}
		copy(r, nb)
		r[z.nameLen] = byte(len(nb))
		r[z.nameLen+1] = zi.reserved
		binary.LittleEndian.PutUint16(r[z.nameLen+2:], uint16(len(zi.channels)))
		binary.LittleEndian.PutUint16(r[z.nameLen+4:], zi.index)
		for j, c := range zi.channels {
			binary.LittleEndian.PutUint16(r[z.nameLen+8+2*j:], c)
		}
		o += z.size
	}
}

func drStr(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

func drRemove(l []uint16, v uint16) ([]uint16, bool) {
	var out []uint16
	found := false
	for _, x := range l {
		if x == v {
			found = true
			continue
		}
		out = append(out, x)
	}
	return out, found
}

// liste di scansione (ModuleDataScanList): un canale cancellato viene tolto;
// la lista modificata si riscrive con i canali in ordine, come il programma
func (l *drLayout) scanRemove(img []byte, idx uint16) {
	m := l.mods["SL"]
	p := m.params
	maxCh, nameLen, infoSize := int(p[0]), int(p[1]), int(binary.LittleEndian.Uint16(p[6:]))
	d := img[m.off:]
	cnt := int(d[0])
	for i := 0; i < cnt; i++ {
		r := d[4+i*infoSize : 4+(i+1)*infoSize]
		n := int(r[1])
		at := 16 + nameLen
		var chs []uint16
		for j := 0; j < n && j < maxCh; j++ {
			chs = append(chs, binary.LittleEndian.Uint16(r[at+2*j:]))
		}
		chs, found := drRemove(chs, idx)
		if !found {
			continue
		}
		sort.Slice(chs, func(a, b int) bool { return chs[a] < chs[b] })
		r[1] = byte(len(chs))
		for j := 0; j < maxCh; j++ {
			v := uint16(0)
			if j < len(chs) {
				v = chs[j]
			}
			binary.LittleEndian.PutUint16(r[at+2*j:], v)
		}
	}
}

// liste di roaming (ModuleDataRoaming): 8 liste da 8 canali
func (l *drLayout) roamRemove(img []byte, idx uint16) {
	m, ok := l.mods["RM"]
	if !ok {
		return
	}
	d := img[m.off+4:]
	for i := 0; i < 8; i++ {
		r := d[i*20 : (i+1)*20]
		n := int(int16(binary.LittleEndian.Uint16(r[2:])))
		if n < 0 {
			n = 0
		}
		if n > 8 {
			n = 8
		}
		var chs []int16
		found := false
		for j := 0; j < n; j++ {
			v := int16(binary.LittleEndian.Uint16(r[4+2*j:]))
			if v == int16(idx) {
				found = true
				continue
			}
			chs = append(chs, v)
		}
		if !found {
			continue
		}
		sort.Slice(chs, func(a, b int) bool { return chs[a] < chs[b] })
		binary.LittleEndian.PutUint16(r[2:], uint16(len(chs)))
		for j := 0; j < 8; j++ {
			v := int16(0)
			if j < len(chs) {
				v = chs[j]
			}
			binary.LittleEndian.PutUint16(r[4+2*j:], uint16(v))
		}
	}
}

// ---------------------------------------------------------------- famiglia

type drOpt struct{}

type drFam struct{}

func init() {
	genRegister(&genModel{ID: "bf-dr1801uv", Vendor: "Baofeng", Model: "DR-1801UV", RClass: "IU0VYT_DR1801UV", Name: "Baofeng DR-1801UV",
		Slots: 1024, First: 1, NameMax: 16, Charset: charsetASCII,
		Powers: []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, NoTxOff: true, NoChirp: true,
		DCS: dtcsCodes, Baud: drLowBaud, Fam: drFam{}, Opt: &drOpt{}, Digital: "DMR",
		Notes: "Radio DMR (piattaforma Auctus A6). Protocollo e memoria ricavati dal programma del costruttore «DR-1801A6» e provati con il suo codice originale. " +
			"Qui si modificano frequenze, nome, potenza, banda e toni; per i canali DMR contatto, gruppi, slot e color code restano quelli impostati " +
			"(un canale DMR nuovo prende i valori del programma: primo contatto, slot 1, color code 0). Un canale nuovo viene aggiunto alla prima zona con posto; " +
			"un canale cancellato viene tolto da zone, liste di scansione e roaming. Contatti, messaggi e impostazioni non si toccano. " +
			"Se sulla radio è impostata una password di programmazione, toglila prima con il programma del costruttore."})
}

func (drFam) lay(img []byte) (*drLayout, drCh, bool) {
	l, err := drParse(img)
	if err != nil {
		return nil, drCh{}, false
	}
	return l, l.ch(), true
}

func drToneDec(typ, inv byte, v uint16) string {
	switch typ {
	case 1:
		if v == 0 {
			return ""
		}
		return fmtTone(int(v))
	case 2:
		return genTone{Mode: "DTCS", Code: int(v), Inv: inv == 1}.String()
	}
	return ""
}

func (f drFam) Get(m *genModel, img []byte, i int) *Chan {
	_, c, ok := f.lay(img)
	if !ok || i >= c.max {
		return nil
	}
	p := c.find(img, i)
	if p < 0 {
		return nil
	}
	r := c.rec(img, p)
	ch := &Chan{Rx: int64(binary.LittleEndian.Uint32(r[4:])), Tx: int64(binary.LittleEndian.Uint32(r[8:])), Name: drStr(c.name(img, p))}
	if r[3] == 1 {
		ch.Power = m.powerSem(0)
	} else {
		ch.Power = m.powerSem(1)
	}
	if r[2] == 2 || r[2] == 3 {
		ch.Dig = true
		return ch
	}
	ch.Narrow = r[25] == 0
	ch.RxTone = drToneDec(r[30], r[31], binary.LittleEndian.Uint16(r[28:]))
	ch.TxTone = drToneDec(r[34], r[35], binary.LittleEndian.Uint16(r[32:]))
	return ch
}

func (f drFam) Rec(m *genModel, img []byte, i int) []byte {
	_, c, ok := f.lay(img)
	if !ok {
		return nil
	}
	p := c.find(img, i)
	if p < 0 {
		return make([]byte, drRecSize+c.nameSize)
	}
	return append(append([]byte{}, c.rec(img, p)...), c.name(img, p)...)
}

func (f drFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	_, c, ok := f.lay(img)
	if !ok || len(raw) != c.size+c.nameSize {
		return
	}
	p := f.slot(img, c, i)
	if p < 0 {
		return
	}
	copy(c.rec(img, p), raw[:c.size])
	copy(c.name(img, p), raw[c.size:])
	binary.LittleEndian.PutUint16(c.rec(img, p), uint16(i))
	c.recount(img)
}

func (f drFam) Match(m *genModel, img []byte) bool { return false }

// posizione per il canale i: quella che ha già, altrimenti la posizione i se è
// libera (come scrive il programma), altrimenti la prima libera
func (drFam) slot(img []byte, c drCh, i int) int {
	if p := c.find(img, i); p >= 0 {
		return p
	}
	if i < c.max && binary.LittleEndian.Uint16(c.rec(img, i)) == 0xFFFF {
		return i
	}
	for p := 0; p < c.max; p++ {
		if binary.LittleEndian.Uint16(c.rec(img, p)) == 0xFFFF {
			return p
		}
	}
	return -1
}

func drToneEnc(t genTone) (typ, inv byte, v uint16) {
	switch t.Mode {
	case "Tone":
		return 1, 0, uint16(t.Hz10)
	case "DTCS":
		if t.Inv {
			inv = 1
		}
		return 2, inv, uint16(t.Code)
	}
	return 0, 0, 0
}

// parametri dei CTCSS fuori dall'elenco del programma (CtdcssItem_Fb): coseno e
// seno della frequenza a 8 kHz in Q14; per i toni dell'elenco valgono 0
var drCTCSS = []int{630, 670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230,
	1273, 1318, 1365, 1413, 1462, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835, 1862, 1899, 1928, 1966,
	1995, 2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

func drCtcPara(hz10 int) (uint16, uint16) {
	if idxInt(drCTCSS, hz10) >= 0 {
		return 0, 0
	}
	a := math.Pi * 2 * (float64(float32(hz10) * float32(0.1))) / 8000
	p1 := math.Round(2 * math.Cos(a) * 16384)
	p2 := math.Round(math.Sin(a) * 16384)
	return uint16(uint32(p1)), uint16(uint32(p2))
}

func (f drFam) Set(m *genModel, img []byte, i int, ch *Chan) error {
	l, c, ok := f.lay(img)
	if !ok {
		return fmt.Errorf("dati della radio non validi")
	}
	if i >= c.max {
		return fmt.Errorf("la radio ha %d canali", c.max)
	}
	idx := uint16(i)
	if ch == nil {
		p := c.find(img, i)
		if p < 0 {
			return nil
		}
		if binary.LittleEndian.Uint16(img[c.base:]) <= 1 {
			return fmt.Errorf("la radio deve avere almeno un canale")
		}
		r := c.rec(img, p)
		for k := range r {
			r[k] = 0
		}
		binary.LittleEndian.PutUint16(r, 0xFFFF)
		nm := c.name(img, p)
		for k := range nm {
			nm[k] = 0
		}
		c.recount(img)
		z := l.zones(img)
		changed := false
		for k := range z.list {
			var found bool
			z.list[k].channels, found = drRemove(z.list[k].channels, idx)
			changed = changed || found
		}
		if changed {
			z.write(img)
		}
		l.scanRemove(img, idx)
		l.roamRemove(img, idx)
		return nil
	}
	tt, err := genToneParse(ch.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(ch.RxTone)
	if err != nil {
		return err
	}
	p := c.find(img, i)
	isNew := p < 0
	if isNew {
		if p = f.slot(img, c, i); p < 0 {
			return fmt.Errorf("memoria dei canali piena")
		}
		// valori del programma per un canale nuovo (ModuleDataChannel.CreateNew)
		r := c.rec(img, p)
		for k := range r {
			r[k] = 0
		}
		binary.LittleEndian.PutUint16(r, idx)
	}
	r := c.rec(img, p)
	typ := byte(1)
	if ch.Dig {
		typ = 3
		if r[2] == 2 {
			typ = 2
		}
	}
	r[2] = typ
	r[3] = 0
	if m.powerIndex(ch.Power) == 0 {
		r[3] = 1
	}
	binary.LittleEndian.PutUint32(r[4:], uint32(ch.Rx))
	binary.LittleEndian.PutUint32(r[8:], uint32(ch.Tx))
	if !ch.Dig {
		r[25] = 1
		if ch.Narrow {
			r[25] = 0
		}
		ty, inv, v := drToneEnc(rt)
		binary.LittleEndian.PutUint16(r[28:], v)
		r[30], r[31] = ty, inv
		ty, inv, v = drToneEnc(tt)
		binary.LittleEndian.PutUint16(r[32:], v)
		r[34], r[35] = ty, inv
		var p1, p2 uint16
		if tt.Mode == "Tone" {
			p1, p2 = drCtcPara(tt.Hz10)
		}
		binary.LittleEndian.PutUint16(r[36:], p1)
		binary.LittleEndian.PutUint16(r[38:], p2)
	}
	nm := c.name(img, p)
	for k := range nm {
		nm[k] = 0
	}
	name := ch.Name
	if len(name) > c.nameLen {
		name = name[:c.nameLen]
	}
	copy(nm, name)
	c.recount(img)
	// un canale che non è in nessuna zona non si raggiunge dalla radio: va nella
	// prima zona con posto (o in una zona nuova, come «Crea» del programma)
	z := l.zones(img)
	for _, zi := range z.list {
		for _, v := range zi.channels {
			if v == idx {
				return nil
			}
		}
	}
	for k := range z.list {
		if len(z.list[k].channels) < z.maxCh {
			z.list[k].channels = append(z.list[k].channels, idx)
			z.write(img)
			return nil
		}
	}
	if len(z.list) < z.max {
		n := uint16(0)
		for ; int(n) < z.max; n++ {
			used := false
			for _, zi := range z.list {
				used = used || zi.index == n
			}
			if !used {
				break
			}
		}
		z.list = append(z.list, drZone{name: fmt.Sprintf("Zone %d", n), index: n, channels: []uint16{idx}})
		z.write(img)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func drPkt(cmd uint16, data []byte) []byte {
	n := 6 + len(data)
	b := []byte{0xAA, byte(n), byte(cmd), byte(cmd >> 8)}
	b = append(b, data...)
	s := byte(n) ^ byte(cmd) ^ byte(cmd>>8)
	for _, x := range data {
		s ^= x
	}
	return append(b, s, 0xBB)
}

func drCrc(d []byte) [2]byte {
	var e, o byte
	for i, x := range d {
		if i%2 == 1 {
			o ^= x
		} else {
			e ^= x
		}
	}
	return [2]byte{e, o}
}

type drLink struct {
	ctx  context.Context
	port string
	p    Port
}

func (l *drLink) open(baud int) error {
	if l.p != nil {
		l.p.Close()
	}
	p, err := openPortFunc(l.port, baud)
	if err != nil {
		return err
	}
	l.p = p
	return nil
}

func (l *drLink) sleep(d time.Duration) error {
	select {
	case <-l.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// risposta: pacchetto completo con il comando atteso (gli altri si scartano)
func (l *drLink) reply(want uint16, to time.Duration) ([]byte, error) {
	end := time.Now().Add(to)
	for {
		left := time.Until(end)
		if left <= 0 {
			return nil, errTimeout
		}
		h := make([]byte, 2)
		if err := l.p.ReadFull(l.ctx, h[:1], left); err != nil {
			return nil, err
		}
		if h[0] != 0xAA {
			continue
		}
		if err := l.p.ReadFull(l.ctx, h[1:], time.Second); err != nil {
			return nil, err
		}
		n := int(h[1])
		if n < 6 {
			continue
		}
		rest := make([]byte, n-2)
		if err := l.p.ReadFull(l.ctx, rest, time.Second); err != nil {
			return nil, err
		}
		pk := append(h, rest...)
		if pk[n-1] != 0xBB {
			continue
		}
		if uint16(pk[2])<<8|uint16(pk[3]) == want {
			return pk[4 : n-2], nil
		}
	}
}

func (l *drLink) xfer(cmd uint16, data []byte, want uint16, to time.Duration) ([]byte, error) {
	_ = l.p.Purge()
	if err := l.p.Write(drPkt(cmd, data)); err != nil {
		return nil, err
	}
	return l.reply(want, to)
}

// collegamento e password (vuota), come PipeLineCpsData*Task
func (l *drLink) enter(write bool) error {
	if err := l.open(drLowBaud); err != nil {
		return err
	}
	var last error
	ok := false
	for k := 0; k < 10 && !ok; k++ {
		r, err := l.xfer(0x0401, nil, 0x8104, 300*time.Millisecond)
		if err == errCancelled {
			return err
		}
		if err == nil && len(r) > 0 && r[0] == 1 {
			ok = true
		} else if err != nil {
			last = err
		}
	}
	if !ok {
		if last == nil {
			last = rErr("la radio non è pronta al collegamento")
		}
		return last
	}
	to := time.Second
	if write {
		to = 2 * time.Second
	}
	for k := 0; k < 4; k++ {
		r, err := l.xfer(0x2B00, []byte{0}, 0x802B, to)
		if err != nil {
			return err
		}
		if len(r) > 0 && (r[0] == 1 || r[0] == 2) {
			return nil
		}
		if write {
			break
		}
	}
	return rErr("sulla radio è impostata una password di programmazione: toglila con il programma del costruttore")
}

func drOpen(ctx context.Context, port string) *drLink { return &drLink{ctx: ctx, port: port} }

func (f drFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	l := drOpen(ctx, port)
	defer func() {
		if l.p != nil {
			l.p.Close()
		}
	}()
	prog(0, 1, "Collegamento con la radio…")
	if err := l.enter(false); err != nil {
		return nil, "", genFail("Collegamento", err)
	}
	if err := l.sleep(time.Second); err != nil {
		return nil, "", err
	}
	baud := make([]byte, 4)
	binary.BigEndian.PutUint32(baud, drHighBaud)
	r, err := l.xfer(0x0001, baud, 0x8100, 3*time.Second)
	if err != nil {
		return nil, "", genFail("Richiesta di lettura", err)
	}
	if len(r) < 15 || r[0] != 1 {
		return nil, "", genFail("Richiesta di lettura", rErr("la radio ha rifiutato la lettura"))
	}
	n := int(binary.BigEndian.Uint32(r[1:]))
	ver := int(binary.BigEndian.Uint32(r[5:]))
	crc := [2]byte{r[13], r[14]}
	info := fmt.Sprintf("dati versione %d, hardware %d", ver, binary.BigEndian.Uint32(r[9:]))
	if ver != drDataVer {
		return nil, info, genFail("Richiesta di lettura", rErr("versione dei dati della radio %d diversa da quella del programma del costruttore (%d)", ver, drDataVer))
	}
	if n <= 0 || n > 4<<20 {
		return nil, info, genFail("Richiesta di lettura", rErr("lunghezza dei dati non valida (%d)", n))
	}
	if err := l.open(drHighBaud); err != nil {
		return nil, info, err
	}
	if err := l.sleep(time.Second); err != nil {
		return nil, info, err
	}
	if err := l.p.Write(drPkt(0x0101, nil)); err != nil {
		return nil, info, err
	}
	img := make([]byte, 0, n)
	for len(img) < n {
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		k := n - len(img)
		if k > 1024 {
			k = 1024
		}
		b := make([]byte, k)
		if err := l.p.ReadFull(ctx, b, 30*time.Second); err != nil {
			return nil, info, genFail("Lettura", err)
		}
		img = append(img, b...)
		prog(len(img), n, "Lettura…")
	}
	if drCrc(img) != crc {
		return nil, info, genFail("Lettura", rErr("dati ricevuti non validi (controllo CRC)"))
	}
	lay, err := drParse(img)
	if err != nil {
		return nil, info, genFail("Lettura", err)
	}
	if c := lay.ch(); c.max != m.Slots || c.size != drRecSize {
		return nil, info, genFail("Lettura", rErr("memoria dei canali diversa da quella prevista (%d canali da %d byte)", c.max, c.size))
	}
	prog(n, n, "Lettura completata")
	return img, info, nil
}

func (f drFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	lay, err := drParse(img)
	if err != nil {
		return "", err
	}
	body := img[lay.hl+4:]
	l := drOpen(ctx, port)
	defer func() {
		if l.p != nil {
			l.p.Close()
		}
	}()
	prog(0, len(body), "Collegamento con la radio…")
	if err := l.enter(true); err != nil {
		return "", genFail("Collegamento", err)
	}
	if err := l.sleep(time.Second); err != nil {
		return "", err
	}
	crc := drCrc(body)
	req := make([]byte, 12)
	binary.BigEndian.PutUint16(req, drWriteVer)
	binary.BigEndian.PutUint32(req[2:], uint32(len(body)))
	req[6], req[7] = crc[0], crc[1]
	binary.BigEndian.PutUint32(req[8:], drHighBaud)
	if _, err := l.xfer(0x0201, req, 0x8102, 2*time.Second); err != nil {
		return "", genFail("Richiesta di scrittura", err)
	}
	if err := l.open(drHighBaud); err != nil {
		return "", err
	}
	if err := l.sleep(time.Second); err != nil {
		return "", err
	}
	for o := 0; o < len(body); o += drChunk {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		e := o + drChunk
		if e > len(body) {
			e = len(body)
		}
		if err := l.p.Write(body[o:e]); err != nil {
			return "", err
		}
		prog(e, len(body), "Scrittura…")
		if e < len(body) {
			if err := l.sleep(drChunkWait); err != nil {
				return "", err
			}
		}
	}
	r, err := l.reply(0x8103, time.Duration(len(body)/200*100+10000)*time.Millisecond)
	if err != nil {
		return "", genFail("Scrittura", err)
	}
	info := ""
	if len(r) >= 3 && (r[0] != 1 || r[1] != 0 || r[2] != 0) {
		info = fmt.Sprintf("risposta della radio alla fine della scrittura: % X", r[:3])
	}
	prog(len(body), len(body), "Scrittura completata")
	return info, nil
}
