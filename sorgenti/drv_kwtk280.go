// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tk280.py (4.18).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.18: palmari professionali Kenwood TK-280 e TK-380 (tk280.py di CHIRP).
//
// Come le TK-2140 / TK-3140 (drv_kwtk140.go): 250 posizioni fisiche da 48 byte a
// 0x2000, ognuna con gruppo e numero; la tabella a 0x300 (gruppo, numero, indice
// del gruppo, posizione) dice quali posizioni sono in uso, in ordine di gruppo e
// numero; i gruppi stanno a 0x1000. La lista del programma è quella delle
// posizioni fisiche; un canale nuovo entra nel gruppo del canale che lo precede.
//
// 9600 baud, parità pari: «PROGRAM» → 06; 02 → sigla di 8 byte; 06 → 06; 'P' → 10
// byte; 06 → 06; 'O'. Blocchi di 256 byte: 'R' n(2) → 'W' dati somma (oppure 'Z'
// e un byte), 06 → 06; 'W' n(2) dati somma (oppure 'Z' n(2) FF), 06 → 06; i
// blocchi da 0x51 a 0x5F non si scrivono. 'E' per uscire.

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type kwx80Fam struct{}

type kwx80Opt struct{ typ string }

const kwx80Format = `
struct {
  u8 format[2];
  u8 unknown0[12];
  u8 groups;
  u8 channels;
} conv;
#seekto 0x0300;
struct {
    u8 group;
    u8 number;
    u8 group_index;
    u8 index;
} group_mapping[250];
#seekto 0x1000;
struct {
  u8 number;
  u8 channels;
  char name[10];
  u8 data_grp;
  u8 data_chan;
  u8 unknown108;
  u8 unknown109;
} groups[250];
#seekto 0x2000;
struct {
  u8 number;
  u8 group;
  char name[10];
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 unknown_rx:4,
     rx_step:4;
  u8 unknown_tx:4,
     tx_step:4;
  ul16 rx_tone;
  ul16 tx_tone;
  u8 unknown23[5];
  u8 signalling;
  u8 unknown24:1,
     ptt_id:1,
     beat_shift:1,
     busy_lock:1,
     data_en:1,
     power:1,
     compander:1,
     wide:1;
  u8 unknown27:4,
     busy_lock_opt:2,
     unknown28:2;
  u8 unknown29[14];
} memory[250];
`

const (
	kwx80Size = 0x8000
	kwx80Add  = 0x60 // bit «in scansione» delle posizioni
	kwx80ID   = 0xA7
	kwx80N    = 250
)

func init() {
	mk := func(id, model, class, typ string, lo, hi int64, pw []genPower) *genModel {
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: kwx80N, First: 1, NameMax: 10, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz 1234567890_-*()/\\+=.",
			Powers: pw, Bands: []band{{lo, hi}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kwx80Fam{}, Opt: &kwx80Opt{typ: typ},
			Notes: "250 posizioni con nomi di 10 caratteri, divise nei gruppi (zone) della radio. Radio professionale: si programma col " +
				"cavo di programmazione Kenwood, accendendo la radio dopo aver collegato il cavo. La lista mostra le posizioni della " +
				"memoria: un canale modificato resta nel suo gruppo con il suo numero; un canale nuovo entra nel gruppo del canale " +
				"che lo precede nella lista (il primo gruppo se non ce n'è) con il primo numero libero. Segnalazione, compander, " +
				"blocco del canale occupato e «beat shift» si conservano. I gruppi non si creano e non si cancellano da qui. Driver di CHIRP."}
	}
	genRegister(
		mk("kw-tk280", "TK-280", "TK280_Radios", "P0280", 136000000, 174000000, kwtkPowers),
		mk("kw-tk380", "TK-380", "TK380_Radios", "P0380", 400000000, 520000000, []genPower{{"Bassa 1 W", 1, 2}, {"Alta 4 W", 4, 0}}),
	)
}

func (kwx80Fam) mem() *bwEl                 { return bwMust(kwx80Format) }
func (f kwx80Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f kwx80Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f kwx80Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f kwx80Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	old := f.rec(i).Raw(img)
	num, grp := old[0], old[1]
	f.rec(i).SetRaw(img, raw)
	r := f.rec(i).Raw(img)
	r[0], r[1] = num, grp
}
func (f kwx80Fam) Match(m *genModel, img []byte) bool {
	return len(img) == kwx80Size && string(img[kwx80ID:kwx80ID+5]) == m.Opt.(*kwx80Opt).typ
}

type kwx80Map struct{ group, number, gi, index int }

// maps: voci valide della tabella (posizione nella tabella → voce)
func (f kwx80Fam) maps(img []byte) map[int]kwx80Map {
	out := map[int]kwx80Map{}
	gm := f.mem().F("group_mapping")
	for k := 0; k < kwx80N; k++ {
		e := gm.I(k).Raw(img)
		if e[0] < 255 && e[1] < 255 {
			out[k] = kwx80Map{int(e[0]), int(e[1]), int(e[2]), int(e[3])}
		}
	}
	return out
}

// slot: voce della tabella che usa la posizione fisica i (-1 se libera)
func (f kwx80Fam) slot(img []byte, i int) (int, kwx80Map) {
	for k, e := range f.maps(img) {
		if e.index == i {
			return k, e
		}
	}
	return -1, kwx80Map{}
}

// kwx80Step: passo della frequenza come in CHIRP (12,5 - 6,25 - 5 - 2,5 kHz)
func kwx80Step(hz int64) int64 {
	switch {
	case hz%12500 == 0:
		return 5
	case hz%6250 == 0:
		return 2
	case hz%5000 == 0:
		return 1
	case hz%2500 == 0:
		return 0
	}
	return 1
}

func (f kwx80Fam) Get(m *genModel, img []byte, i int) *Chan {
	if k, _ := f.slot(img, i); k < 0 {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	c.Tx = c.Rx
	if r.F("txfreq").Raw(img)[0] == 0xFF {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = strings.TrimRight(r.F("name").Str(img), " \xff")
	c.TxTone = kw60gToneDec(r.F("tx_tone").Int(img))
	c.RxTone = kw60gToneDec(r.F("rx_tone").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = !r.F("wide").Bool(img)
	c.Scan = kwtkBit(img, kwx80Add, i)
	return c
}

func (f kwx80Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	k, cur := f.slot(img, i)
	gr := f.mem().F("groups")
	gm := f.mem().F("group_mapping")
	if c == nil {
		if k >= 0 {
			if cur.gi < kwx80N {
				g := gr.I(cur.gi).F("channels")
				g.SetInt(img, g.Int(img)-1)
			}
			r.Fill(img, 0xFF)
			gm.I(k).Fill(img, 0xFF)
			f.compact(img)
		}
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	if k < 0 {
		// canale nuovo: gruppo del canale che lo precede (il primo gruppo se non c'è),
		// primo numero libero, prima voce libera della tabella
		ms := f.maps(img)
		group := -1
		for p := i - 1; p >= 0 && group < 0; p-- {
			for _, e := range ms {
				if e.index == p {
					group = e.group
				}
			}
		}
		gi := -1
		for x := 0; x < kwx80N; x++ {
			n := int(gr.I(x).F("number").Int(img))
			if n > 250 {
				continue
			}
			if group < 0 {
				group = n
			}
			if n == group {
				gi = x
				break
			}
		}
		if gi < 0 {
			return fmt.Errorf("la radio non ha gruppi: creane uno con il programma del costruttore")
		}
		taken := map[int]bool{}
		for _, e := range ms {
			if e.group == group {
				taken[e.number] = true
			}
		}
		num := 1
		for taken[num] {
			num++
		}
		empty := -1
		for x := 0; x < kwx80N; x++ {
			if gm.I(x).Raw(img)[0] == 0xFF {
				empty = x
				break
			}
		}
		if num > 250 || empty < 0 {
			return fmt.Errorf("il gruppo %d è pieno", group)
		}
		gm.I(empty).SetRaw(img, []byte{byte(group), byte(num), byte(gi), byte(i)})
		g := gr.I(gi).F("channels")
		g.SetInt(img, g.Int(img)+1)
		f.compact(img)
		cur = kwx80Map{group, num, gi, i}
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	r.F("unknown_rx").SetInt(img, 3)
	r.F("unknown_tx").SetInt(img, 3)
	r.F("rx_step").SetInt(img, kwx80Step(c.Rx/10*10))
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
		r.F("tx_step").SetInt(img, 1)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
		r.F("tx_step").SetInt(img, kwx80Step(c.Tx/10*10))
	}
	if err := kw60gToneEnc(r.F("tx_tone"), img, c.TxTone); err != nil {
		return err
	}
	if err := kw60gToneEnc(r.F("rx_tone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("name").SetStr(img, c.Name, ' ')
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("wide").SetBool(img, !c.Narrow)
	kwtkSetBit(img, kwx80Add, i, c.Scan)
	r.F("number").SetInt(img, int64(cur.number))
	r.F("group").SetInt(img, int64(cur.group))
	return nil
}

// compact: tabella in ordine di gruppo e numero, canali di ogni gruppo e totali
// (_compact_mappings di CHIRP)
func (f kwx80Fam) compact(img []byte) {
	var valid []kwx80Map
	counts := map[int]int{}
	gm := f.mem().F("group_mapping")
	for k := 0; k < kwx80N; k++ {
		e := gm.I(k).Raw(img)
		if e[0] < 255 && e[1] < 255 {
			valid = append(valid, kwx80Map{int(e[0]), int(e[1]), int(e[2]), int(e[3])})
			counts[int(e[0])]++
		}
	}
	sort.SliceStable(valid, func(a, b int) bool {
		if valid[a].group != valid[b].group {
			return valid[a].group < valid[b].group
		}
		return valid[a].number < valid[b].number
	})
	gr := f.mem().F("groups")
	for k := 0; k < kwx80N; k++ {
		e := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		if k < len(valid) {
			v := valid[k]
			e = []byte{byte(v.group), byte(v.number), byte(v.gi), byte(v.index)}
		}
		gm.I(k).SetRaw(img, e)
		if n, ok := counts[int(gr.I(k).F("number").Int(img))]; ok {
			gr.I(k).F("channels").SetInt(img, int64(n))
		}
	}
	f.mem().F("conv").F("groups").SetInt(img, int64(len(counts)))
	f.mem().F("conv").F("channels").SetInt(img, int64(len(valid)))
}

func (f kwx80Fam) open(ctx context.Context, m *genModel, port string) (*rawSess, error) {
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
	if err := setEvenParity(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	// come il programma del costruttore: la radio può non rispondere al primo invito
	var b []byte
	for try := 0; try < 5; try++ {
		s.p.Purge()
		s.send([]byte("PROGRAM"))
		b = readUpTo(ctx, s.p, 1, time.Second)
		if len(b) == 1 && b[0] == 0x06 {
			break
		}
		if err := s.sleep(500 * time.Millisecond); err != nil {
			return fail(err)
		}
	}
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
	if !bytes.Contains(id, []byte(m.Opt.(*kwx80Opt).typ)) {
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwx80Opt); ok && bytes.Contains(id, []byte(xo.typ)) {
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
	s.send([]byte{'O'})
	return s, nil
}

func (f kwx80Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, kwx80Size, "Collegamento con la radio…")
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
	for blk := 0; blk < kwx80Size/256; blk++ {
		prog(blk*256, kwx80Size, "Lettura…")
		s.send([]byte{'R', byte(blk >> 8), byte(blk)})
		c := readUpTo(ctx, s.p, 1, time.Second)
		if len(c) != 1 {
			return fail(genFail("Lettura", errTimeout))
		}
		block := bytes.Repeat([]byte{0xFF}, 256)
		if c[0] == 'Z' {
			if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 {
				return fail(genFail("Lettura", errTimeout))
			}
		} else {
			b := readUpTo(ctx, s.p, 257, 2*time.Second)
			if len(b) != 257 {
				return fail(rErr("lettura del blocco 0x%04X non riuscita", blk*256))
			}
			if kwtkSum(b[:256]) != b[256] {
				return fail(rErr("somma di controllo errata: ripeti la lettura"))
			}
			block = b[:256]
		}
		if err := s.ack2(); err != nil {
			return fail(err)
		}
		data = append(data, block...)
	}
	prog(kwx80Size, kwx80Size, "Lettura completata")
	return data, "", nil
}

func (f kwx80Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, kwx80Size, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	defer s.send([]byte{'E'})
	for blk := 0; blk < kwx80Size/256; blk++ {
		if blk > 0x50 && blk < 0x60 {
			continue
		}
		prog(blk*256, kwx80Size, "Scrittura…")
		data := img[blk*256 : blk*256+256]
		frame := []byte{'Z', byte(blk >> 8), byte(blk), 0xFF}
		if !bytes.Equal(data, bytes.Repeat([]byte{0xFF}, 256)) {
			frame = append([]byte{'W', byte(blk >> 8), byte(blk)}, data...)
			frame = append(frame, kwtkSum(data))
		}
		if err := s.send(frame); err != nil {
			return "", err
		}
		if err := s.ack2(); err != nil {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", blk*256)
		}
	}
	prog(kwx80Size, kwx80Size, "Scrittura completata")
	return "", nil
}
