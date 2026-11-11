// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tk3140.py (4.18).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.18: palmari professionali Kenwood TK-2140 e TK-3140 (tk3140.py di CHIRP).
//
// La memoria ha 250 posizioni fisiche; ogni canale porta scritti il gruppo (zona)
// e il numero dentro il gruppo, e una tabella (0x300) elenca i canali in ordine di
// gruppo e numero. CHIRP mostra ogni gruppo come una radio a parte; qui la lista
// è quella delle 250 posizioni fisiche: un canale modificato resta nel suo gruppo
// con il suo numero, un canale nuovo entra nel gruppo del canale che lo precede
// nella lista (il primo gruppo se non ce n'è) con il primo numero libero.
//
// 9600 baud, 8N2: «PROGRAM» → 06; 02 → sigla di 8 byte; 06 → 06; 'P' → 10 byte;
// 06 → 06. Blocchi di 256 byte: 'R' 00 n → 'W' dati somma (oppure 'Z' e il byte di
// riempimento), 06 → 06; 'W' 00 n dati somma → 06 ('Z' 00 n byte per un blocco
// uniforme); i blocchi da 0x54 a 0x5F non si scrivono. I bit della scansione
// stanno a parte: 'S' indirizzo(2) 10 → 'X' e 16 byte, 06 → 06; 'X' indirizzo(2)
// 10 dati somma(modulo 255) → 06, 06. 'E' per uscire.

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type kwx140Fam struct{}

type kwx140Opt struct{ id string }

const kwx140Format = `
#seekto 0x00E;
u8 zone_count;
u8 memory_count;
#seekto 0x300;
struct {
    u8 zone;
    u8 memory;
    u8 zoneinfo_index;
    u8 slot;
} index[250];
#seekto 0x1000;
struct {
    u8 zone;
    u8 count;
    char name[10];
    u8 unknown[4];
} zoneinfo[250];
#seekto 0x2000;
struct {
    u8 memory;
    u8 zone;
    char name[10];
    lbcd rx_freq[4];
    lbcd tx_freq[4];
    u8 rxsomething:4,
       rxstep:4;
    u8 txsomething:4,
       txstep:4;
    ul16 rxtone;
    ul16 txtone;
    u8 unknown2[5];
    u8 optsig;
    u8 unknown2_1:1,
       pttid:1,
       beatshift:1,
       bcl:1,
       unknown1_2:1,
       highpower:1,
       compander:1,
       wide:1;
    u8 unknown3_1:4,
       bcl_optsig:2,
       mode2tone:2;
    u8 unknown4[14];
} memories[250];
#seekto 0x8000;
u8 scanbits[32];
`

const (
	kwx140Size  = 0x8020
	kwx140Scan  = 0x8000
	kwx140ID    = 0xA7
	kwx140Slots = 250
)

func init() {
	mk := func(id, model, class, ident string, lo, hi int64) *genModel {
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: kwx140Slots, First: 1, NameMax: 10, Charset: charsetASCII, Powers: kwtkPowers,
			Bands: []band{{lo, hi}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kwx140Fam{}, Opt: &kwx140Opt{id: ident},
			Notes: "250 posizioni con nomi di 10 caratteri, divise nei gruppi (zone) della radio. Radio professionale: si programma col " +
				"cavo di programmazione Kenwood, accendendo la radio dopo aver collegato il cavo. La lista mostra le posizioni della " +
				"memoria: un canale modificato resta nel suo gruppo con il suo numero; un canale nuovo entra nel gruppo del canale " +
				"che lo precede nella lista (il primo gruppo se non ce n'è) con il primo numero libero. Segnalazione, compander, " +
				"blocco del canale occupato e «beat shift» si conservano. I gruppi non si creano e non si cancellano da qui. Driver di CHIRP."}
	}
	genRegister(
		mk("kw-tk2140k", "TK-2140K", "KenwoodTK2140KRadio", "P2140\x04\xFF\xF1", 136000000, 174000000),
		mk("kw-tk3140k", "TK-3140K", "KenwoodTK3140KRadio", "P3140\x06\xFF\xF1", 400000000, 520000000),
		mk("kw-tk3140k2", "TK-3140K2", "KenwoodTK3140K2Radio", "P3140\x07\xFF\xF1", 400000000, 520000000),
		mk("kw-tk3140k3", "TK-3140K3", "KenwoodTK3140K3Radio", "P3140\x08\xFF\xF1", 400000000, 520000000),
	)
}

func kwx140o(m *genModel) *kwx140Opt { return m.Opt.(*kwx140Opt) }

func (kwx140Fam) mem() *bwEl                 { return bwMust(kwx140Format) }
func (f kwx140Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f kwx140Fam) rec(i int) *bwEl          { return f.mem().F("memories").I(i) }
func (f kwx140Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f kwx140Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	// il canale spostato prende gruppo e numero dalla nuova posizione (vedi Set)
	old := f.rec(i).Raw(img)
	mem, zone := old[0], old[1]
	f.rec(i).SetRaw(img, raw)
	r := f.rec(i).Raw(img)
	r[0], r[1] = mem, zone
}
func (f kwx140Fam) Match(m *genModel, img []byte) bool {
	return len(img) >= kwx140Size && string(img[kwx140ID:kwx140ID+8]) == kwx140o(m).id
}

func (f kwx140Fam) used(img []byte, i int) bool { return f.rec(i).Raw(img)[0] != 0xFF }

// zones: gruppi della radio, nell'ordine della tabella dei gruppi
func (f kwx140Fam) zones(img []byte) []int {
	var out []int
	zi := f.mem().F("zoneinfo")
	for k := 0; k < kwx140Slots; k++ {
		z := zi.I(k).F("zone").Int(img)
		if z == 0xFF {
			break
		}
		out = append(out, int(z))
	}
	return out
}

func kwx140ToneDec(v int64) string {
	switch {
	case v == 0xFFFF:
		return ""
	case v&0x2000 != 0:
		var code int
		fmt.Sscanf(fmt.Sprintf("%o", v&0x1FF), "%d", &code)
		return genTone{Mode: "DTCS", Code: code, Inv: v&0x8000 != 0}.String()
	}
	return fmtTone(int(v))
}

func kwx140ToneEnc(e *bwEl, img []byte, s string) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch t.Mode {
	case "Tone":
		e.SetInt(img, int64(t.Hz10))
	case "DTCS":
		var v int64
		fmt.Sscanf(fmt.Sprintf("%d", t.Code), "%o", &v)
		v |= 0x2800
		if t.Inv {
			v |= 0x8000
		}
		e.SetInt(img, v)
	default:
		e.Fill(img, 0xFF)
	}
	return nil
}

// kwx140Step: passo della frequenza come in CHIRP (il primo, nell'ordine 6,25 -
// 5 - 2,5 - 10 - 7,5 kHz, di cui la frequenza è multipla; 5 kHz se nessuno)
func kwx140Step(hz int64) int64 {
	switch {
	case hz%6250 == 0:
		return 2
	case hz%5000 == 0:
		return 1
	case hz%2500 == 0:
		return 0
	case hz%10000 == 0:
		return 4
	case hz%7500 == 0:
		return 6
	}
	return 1
}

func (f kwx140Fam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.used(img, i) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	c.Tx = c.Rx
	if bytes.Equal(r.F("tx_freq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = r.F("tx_freq").Int(img) * 10
	}
	c.RxTone = kwx140ToneDec(r.F("rxtone").Int(img))
	c.TxTone = kwx140ToneDec(r.F("txtone").Int(img))
	c.Name = strings.TrimRight(r.F("name").Str(img), " \xff")
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Narrow = !r.F("wide").Bool(img)
	c.Scan = kwtkBit(img, kwx140Scan, i)
	return c
}

func (f kwx140Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		if f.used(img, i) {
			r.Fill(img, 0xFF)
			f.sortIndex(img)
		}
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	if !f.used(img, i) {
		// canale nuovo: gruppo del canale che lo precede, primo numero libero
		zs := f.zones(img)
		if len(zs) == 0 {
			return fmt.Errorf("la radio non ha gruppi: creane uno con il programma del costruttore")
		}
		zone := zs[0]
		for k := i - 1; k >= 0; k-- {
			if f.used(img, k) {
				zone = int(f.rec(k).F("zone").Int(img))
				break
			}
		}
		taken := map[int]bool{}
		for k := 0; k < kwx140Slots; k++ {
			if f.used(img, k) && int(f.rec(k).F("zone").Int(img)) == zone {
				taken[int(f.rec(k).F("memory").Int(img))] = true
			}
		}
		num := 1
		for taken[num] {
			num++
		}
		if num > 250 {
			return fmt.Errorf("il gruppo %d è pieno", zone)
		}
		r.Fill(img, 0xFF)
		r.F("memory").SetInt(img, int64(num))
		r.F("zone").SetInt(img, int64(zone))
	}
	r.F("rxsomething").SetInt(img, 3)
	r.F("txsomething").SetInt(img, 3)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	r.F("rxstep").SetInt(img, kwx140Step(c.Rx/10*10))
	if c.TxOff {
		r.F("tx_freq").Fill(img, 0xFF)
		r.F("txstep").SetInt(img, 1)
	} else {
		r.F("tx_freq").SetInt(img, c.Tx/10)
		r.F("txstep").SetInt(img, kwx140Step(c.Tx/10*10))
	}
	if err := kwx140ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	if err := kwx140ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
		return err
	}
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("highpower").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("name").SetStr(img, c.Name, ' ')
	kwtkSetBit(img, kwx140Scan, i, c.Scan)
	f.sortIndex(img)
	return nil
}

// sortIndex: tabella dei canali in ordine di gruppo e numero, conteggi dei gruppi
// e totali (sort_index di CHIRP). Un canale di un gruppo che non c'è viene tolto.
func (f kwx140Fam) sortIndex(img []byte) {
	zs := f.zones(img)
	pos := map[int]int{}
	for k, z := range zs {
		if _, ok := pos[z]; !ok {
			pos[z] = k
		}
	}
	type ent struct{ zone, mem, zi, slot int }
	var es []ent
	for i := 0; i < kwx140Slots; i++ {
		r := f.rec(i)
		z := int(r.F("zone").Int(img))
		if z == 0xFF {
			continue
		}
		k, ok := pos[z]
		if !ok {
			r.Fill(img, 0xFF)
			continue
		}
		es = append(es, ent{z, int(r.F("memory").Int(img)), k, i})
	}
	sort.Slice(es, func(a, b int) bool {
		x, y := es[a], es[b]
		if x.zone != y.zone {
			return x.zone < y.zone
		}
		if x.mem != y.mem {
			return x.mem < y.mem
		}
		if x.zi != y.zi {
			return x.zi < y.zi
		}
		return x.slot < y.slot
	})
	counts := map[int]int{}
	ix := f.mem().F("index")
	for i := 0; i < kwx140Slots; i++ {
		e := ent{0xFF, 0xFF, 0xFF, 0xFF}
		if i < len(es) {
			e = es[i]
			counts[e.zone]++
		}
		x := ix.I(i)
		x.F("zone").SetInt(img, int64(e.zone))
		x.F("memory").SetInt(img, int64(e.mem))
		x.F("zoneinfo_index").SetInt(img, int64(e.zi))
		x.F("slot").SetInt(img, int64(e.slot))
	}
	zi := f.mem().F("zoneinfo")
	total := 0
	seen := map[int]bool{}
	for k, z := range zs {
		zi.I(k).F("count").SetInt(img, int64(counts[z]))
		if !seen[z] {
			total += counts[z]
			seen[z] = true
		}
	}
	f.mem().F("memory_count").SetInt(img, int64(total))
	f.mem().F("zone_count").SetInt(img, int64(len(zs)))
}

func (f kwx140Fam) open(ctx context.Context, m *genModel, port string) (*rawSess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (*rawSess, error) {
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
	if b[0] != 0x06 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
	}
	s.send([]byte{0x02})
	id := readUpTo(ctx, s.p, 8, time.Second)
	if len(id) != 8 {
		return fail(errTimeout)
	}
	if string(id) != kwx140o(m).id {
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwx140Opt); ok && xo.id == string(id) {
				return fail(rErr("la radio è un altro modello: apri la scheda «%s»", x.Name))
			}
		}
		return fail(rErr("la radio è un altro modello (%s)", strings.TrimRight(string(id[:5]), "\xff\x00")))
	}
	if err := s.ack2(); err != nil {
		return fail(err)
	}
	s.send([]byte{'P'})
	if v := readUpTo(ctx, s.p, 10, time.Second); len(v) != 10 {
		return fail(errTimeout)
	}
	if err := s.ack2(); err != nil {
		return fail(err)
	}
	return s, nil
}

func (f kwx140Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, kwx140Size, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	defer s.send([]byte{'E'})
	fail := func(e error) ([]byte, string, error) {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		return nil, "", e
	}
	var data []byte
	for blk := 0; blk < 0x80; blk++ {
		prog(blk*256, kwx140Size, "Lettura…")
		if err := s.send([]byte{'R', 0, byte(blk)}); err != nil {
			return nil, "", err
		}
		c := readUpTo(ctx, s.p, 1, time.Second)
		switch {
		case len(c) != 1:
			return fail(genFail("Lettura", errTimeout))
		case c[0] == 'Z':
			b := readUpTo(ctx, s.p, 1, time.Second)
			if len(b) != 1 {
				return fail(genFail("Lettura", errTimeout))
			}
			data = append(data, bytes.Repeat(b, 256)...)
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
	for a := 0x0320; a < 0x0340; a += 0x10 {
		if err := s.send([]byte{'S', byte(a >> 8), byte(a), 0x10}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 17, time.Second)
		if len(b) != 17 || b[0] != 'X' {
			return fail(rErr("lettura del blocco 0x%04X non riuscita", a))
		}
		data = append(data, b[1:]...)
		s.send([]byte{0x06})
		readUpTo(ctx, s.p, 1, time.Second)
	}
	prog(kwx140Size, kwx140Size, "Lettura completata")
	return data, "", nil
}

func (f kwx140Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, kwx140Size, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	defer s.send([]byte{'E'})
	ack := func(a int) error {
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) != 1 || b[0] != 0x06 {
			return rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
		return nil
	}
	for blk := 0; blk < 0x80; blk++ {
		if blk >= 0x54 && blk < 0x60 {
			continue
		}
		prog(blk*256, kwx140Size, "Scrittura…")
		chunk := img[blk*256 : blk*256+256]
		frame := []byte{'Z', 0, byte(blk), chunk[0]}
		if !bytes.Equal(chunk, bytes.Repeat(chunk[:1], 256)) {
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
	for k := 0; k < 2; k++ {
		a := 0x0320 + k*0x10
		chunk := img[kwx140Scan+k*0x10 : kwx140Scan+k*0x10+0x10]
		sum := 0
		for _, x := range chunk {
			sum += int(x)
		}
		frame := append([]byte{'X', byte(a >> 8), byte(a), 0x10}, chunk...)
		frame = append(frame, byte(sum%255))
		if err := s.send(frame); err != nil {
			return "", err
		}
		if err := ack(a); err != nil {
			return "", err
		}
		s.send([]byte{0x06})
	}
	prog(kwx140Size, kwx140Size, "Scrittura completata")
	return "", nil
}
