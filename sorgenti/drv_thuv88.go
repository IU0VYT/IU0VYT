// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: th_uv88.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.2: famiglia TYT TH-UV88 (driver th_uv88.py di CHIRP): TYT TH-UV88 e TH-UV98,
// Retevis RT85, RA89, P2 e P62, Explorer QRZ-1. 200 canali (1-200) di 20 byte da
// 0x0000 (frequenze in decine di Hz, toni a 12 bit, prime 6 lettere del nome),
// altre 10 lettere del nome da 0x1960, canali in uso e canali in scansione in due
// mappe di bit (0x1180 e 0x11A0). Memoria di 0x22A0 byte, 57600 baud.
//
// Protocollo (_do_ident, _download, _upload, _exit_program_mode di th_uv88.py),
// con xxxx = nome del modello nel protocollo («UV88», «UV99», «UV78», «UV29»,
// «UV98»):
//   FE FE EE EF E0 xxxx FD → 36 byte che iniziano con FE FE EF EE E1 xxxx e
//   finiscono con FD;
//   lettura: FE FE EE EF E2 xxxx FD → FE FE EF EE E6 00 FD; per ogni blocco di
//   0x20 byte FE FE EE EF EB indirizzo(4, BE) 00 20 FD → FE FE EF EE E4 indirizzo
//   00 20 + 32 byte + somma + FD (45 byte);
//   scrittura: FE FE EE EF E3 xxxx FD → FE FE EF EE E6 00 FD; per ogni blocco
//   FE FE EE EF E4 indirizzo 00 20 + 32 byte + somma + FD → FE FE EF EE E6 00 FD
//   (i blocchi da 0x1680 a 0x193F non si scrivono, come il programma del
//   costruttore); somma = complemento a 256 della somma di indirizzo, lunghezza
//   e dati; fine: FE FE EE EF E5 xxxx FD.

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"
)

const thuv88Format = `
struct {
  ul32 rxfreq;
  ul32 txfreq;
  ul16 scramble:4,
       rxtone:12;
  ul16 decodeDSCI:1,
       encodeDSCI:1,
       unk1:1,
       unk2:1,
       txtone:12;
  u8   power:2,
       wide:2,
       b_lock:2,
       unk3:2;
  u8   unk4:3,
       signal:2,
       displayName:1,
       unk5:2;
  u8   unk6:2,
       pttid:2,
       unk7:1,
       step:3;
  u8   name[6];
} chan_mem[200];
#seekto 0x1180;
struct {
  u8 bitmap[26];
} chan_avail;
#seekto 0x11A0;
struct {
  u8 bitmap[26];
} chan_skip;
#seekto 0x1960;
struct {
  u8 extra_name[10];
} chan_name[200];
#seekto 0x2180;
struct {
  ul32 rxfreq;
} fm_stations[24];
struct {
  u8 fmset[4];
} fmmap;
struct {
  ul32 fmcur;
} fmfrqs;
`

const (
	thuv88MemSize = 0x22A0
	thuv88Block   = 0x20
	thuv88Wait    = 2 * time.Second
)

type thuv88Opt struct {
	tag  string // nome del modello nel protocollo («UV88»…)
	ra89 bool   // scrambler sì/no (RA89) invece dei livelli 1-8
}

type thuv88Fam struct{}

func init() {
	cs := "ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz0123456789`!\"#$%&'()*+,-./:;<=>?@[]^_"
	pw := []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 0,5 W", 0.5, 2}}
	add := func(id, vendor, model, rclass, name, tag string, ra89 bool, note string) {
		genRegister(&genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: 200, First: 1, NameMax: 10,
			Charset: cs, Powers: pw, Bands: []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: true,
			DCS: dtcsCodes, Baud: 57600, Fam: thuv88Fam{}, Opt: &thuv88Opt{tag: tag, ra89: ra89},
			Notes: "Accendi la radio con il cavo collegato. Canali 1-200. Driver di CHIRP della famiglia TYT TH-UV88 " +
				"(protocollo «" + tag + "»)." + note})
	}
	add("tyt-thuv88", "TYT", "TH-UV88", "THUV88Radio", "TYT TH-UV88", "UV88", false, "")
	add("tyt-thuv98", "TYT", "TH-UV98", "THUV98Radio", "TYT TH-UV98", "UV98", false, "")
	add("rt-rt85", "Retevis", "RT85", "RT85", "Retevis RT85", "UV88", false, " È una TYT TH-UV88.")
	add("rt-ra89", "Retevis", "RA89", "RA89", "Retevis RA89", "UV99", true, "")
	add("rt-p2", "Retevis", "P2", "P2", "Retevis P2", "UV29", false, " Radio senza display.")
	add("rt-p62", "Retevis", "P62", "P62", "Retevis P62", "UV29", false, " Radio senza display, protocollo della Retevis P2.")
	add("ex-qrz1", "Explorer", "QRZ-1", "QRZ1", "Explorer QRZ-1", "UV78", false, "")
}

func thuv88o(m *genModel) *thuv88Opt { return m.Opt.(*thuv88Opt) }

func (thuv88Fam) mem() *bwEl { return bwMust(thuv88Format) }

func (f thuv88Fam) layout(m *genModel) *bwEl { return f.mem() }

func (f thuv88Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("chan_mem").I(i), b.F("chan_name").I(i)
}

// bit del canale i (0 = canale 1) in una delle due mappe
func (f thuv88Fam) bit(img []byte, mapName string, i int) (int, byte) {
	off := f.mem().F(mapName).F("bitmap").Off + i/8
	return off, byte(1) << (i % 8)
}

func (f thuv88Fam) getBit(img []byte, mapName string, i int) bool {
	off, msk := f.bit(img, mapName, i)
	return img[off]&msk != 0
}

func (f thuv88Fam) setBit(img []byte, mapName string, i int, v bool) {
	off, msk := f.bit(img, mapName, i)
	if v {
		img[off] |= msk
	} else {
		img[off] &^= msk
	}
}

// il record e il nome; i due bit (in uso, scansione) in un byte finale
func (f thuv88Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	var st byte
	if f.getBit(img, "chan_avail", i) {
		st |= 1
	}
	if f.getBit(img, "chan_skip", i) {
		st |= 2
	}
	return append(append(append([]byte{}, r.Raw(img)...), n.Raw(img)...), st)
}

func (f thuv88Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:20])
	n.SetRaw(img, raw[20:30])
	f.setBit(img, "chan_avail", i, raw[30]&1 != 0)
	f.setBit(img, "chan_skip", i, raw[30]&2 != 0)
}

func (f thuv88Fam) Match(m *genModel, img []byte) bool { return false }

// tono a 12 bit: oltre 2600 spento, oltre 511 CTCSS in decimi di Hz, altrimenti
// il codice DCS scritto in ottale (es. 023 = 0x13)
func thuv88ToneDec(v int64, inv bool) string {
	switch {
	case v > 2600:
		return ""
	case v > 511:
		return fmtTone(int(v))
	}
	code, _ := strconv.Atoi(strconv.FormatInt(v, 8))
	return genTone{Mode: "DTCS", Code: code, Inv: inv}.String()
}

func thuv88ToneEnc(t genTone) (int64, error) {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10), nil
	case "DTCS":
		v, err := strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		if err != nil {
			return 0, fmt.Errorf("codice DCS %03d non valido", t.Code)
		}
		return v, nil
	}
	return 0xFFF, nil
}

func (f thuv88Fam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.getBit(img, "chan_avail", i) {
		return nil
	}
	r, n := f.recs(i)
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10, Scan: f.getBit(img, "chan_skip", i)}
	if tx := r.F("txfreq").Int(img); tx == 0xFFFFFFFF {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = tx * 10
	}
	// come CHIRP: le 16 lettere (6 nel canale + 10) come latin-1, spazi finali tolti
	c.Name = pyRstrip(latin1Str(string(r.F("name").Raw(img)) + string(n.F("extra_name").Raw(img))))
	c.TxTone = thuv88ToneDec(r.F("txtone").Int(img), r.F("encodeDSCI").Bool(img))
	c.RxTone = thuv88ToneDec(r.F("rxtone").Int(img), r.F("decodeDSCI").Bool(img))
	pw := int(r.F("power").Int(img))
	if pw >= len(m.Powers) {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r.F("wide").Int(img) == 2 // 0 WFM, 1 FM, 2 NFM
	return c
}

func (f thuv88Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if c == nil {
		f.setBit(img, "chan_avail", i, false)
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
	txv, err := thuv88ToneEnc(tt)
	if err != nil {
		return err
	}
	rxv, err := thuv88ToneEnc(rt)
	if err != nil {
		return err
	}
	r, n := f.recs(i)
	was := f.getBit(img, "chan_avail", i)
	f.setBit(img, "chan_avail", i, true)
	f.setBit(img, "chan_skip", i, c.Scan)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0xFFFFFFFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	nb := []byte(latin1Bytes(c.Name))
	if len(nb) > 16 {
		nb = nb[:16]
	}
	nb = append(nb, bytes.Repeat([]byte{' '}, 16-len(nb))...)
	r.F("name").SetRaw(img, nb[:6])
	n.F("extra_name").SetRaw(img, nb[6:])
	r.F("displayName").SetBool(img, c.Name != "")
	r.F("decodeDSCI").SetBool(img, rt.Mode == "DTCS" && rt.Inv)
	r.F("rxtone").SetInt(img, rxv)
	r.F("encodeDSCI").SetBool(img, tt.Mode == "DTCS" && tt.Inv)
	r.F("txtone").SetInt(img, txv)
	wide := int64(1)
	if c.Narrow {
		wide = 2
	}
	r.F("wide").SetInt(img, wide)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	if was {
		// le impostazioni del canale lette e riscritte da CHIRP (mem.extra)
		if v := r.F("b_lock").Int(img); v > 2 {
			r.F("b_lock").SetInt(img, 2)
		}
		sc := r.F("scramble").Int(img)
		if thuv88o(m).ra89 {
			if sc != 0 {
				sc = 1
			}
		} else if sc >= 8 {
			sc = 0
		}
		r.F("scramble").SetInt(img, sc)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type thuv88Sess struct {
	ctx context.Context
	p   Port
	o   *thuv88Opt
}

func (s *thuv88Sess) magic(cmd byte) []byte {
	return append(append([]byte{0xFE, 0xFE, 0xEE, 0xEF, cmd}, s.o.tag...), 0xFD)
}

var thuv88Ack = []byte{0xFE, 0xFE, 0xEF, 0xEE, 0xE6, 0x00, 0xFD}

func (s *thuv88Sess) read(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, thuv88Wait); err != nil {
		return nil, err
	}
	return b, nil
}

// _do_ident: saluto e risposta di 36 byte; poi il comando di lettura (E2) o di
// scrittura (E3) con la conferma
func (s *thuv88Sess) start(cmd byte) ([]byte, error) {
	if err := s.p.Write(s.magic(0xE0)); err != nil {
		return nil, err
	}
	id, err := s.read(36)
	if err != nil {
		return nil, err
	}
	fp := append([]byte{0xFE, 0xFE, 0xEF, 0xEE, 0xE1}, s.o.tag...)
	if !bytes.HasPrefix(id, fp) || id[35] != 0xFD {
		return nil, rErr("risposta inattesa al saluto (% X): è il modello giusto?", id[:min(len(id), 12)])
	}
	if err := s.p.Write(s.magic(cmd)); err != nil {
		return nil, err
	}
	ack, err := s.read(7)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(ack, thuv88Ack) {
		return nil, rErr("la radio non è entrata in programmazione (% X)", ack)
	}
	return id, nil
}

// _exit_program_mode: comando di uscita, poi si scartano i byte residui
func (s *thuv88Sess) exit() {
	_ = s.p.Write(s.magic(0xE5))
	_ = readUpTo(context.Background(), s.p, 256, 50*time.Millisecond)
}

func thuv88Sum(b []byte) byte {
	var n byte
	for _, x := range b {
		n += x
	}
	return -n
}

func (f thuv88Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer p.Close()
	s := &thuv88Sess{ctx: ctx, p: p, o: thuv88o(m)}
	prog(0, thuv88MemSize, "Collegamento con la radio…")
	id, err := s.start(0xE2)
	if err != nil {
		s.exit()
		return nil, "", genFail("Identificazione", err)
	}
	data := make([]byte, 0, thuv88MemSize)
	for a := 0; a < thuv88MemSize; a += thuv88Block {
		if ctx.Err() != nil {
			s.exit()
			return nil, "", errCancelled
		}
		prog(a, thuv88MemSize, "Lettura…")
		fr := []byte{0xFE, 0xFE, 0xEE, 0xEF, 0xEB, byte(a >> 24), byte(a >> 16), byte(a >> 8), byte(a), 0x00, thuv88Block, 0xFD}
		if err := p.Write(fr); err != nil {
			return nil, "", err
		}
		d, err := s.read(thuv88Block + 13)
		if err == nil && (!bytes.HasPrefix(d, []byte{0xFE, 0xFE, 0xEF, 0xEE, 0xE4}) || d[len(d)-1] != 0xFD) {
			// CHIRP lo segnala soltanto; qui ci si ferma: i dati non sarebbero affidabili
			err = rErr("risposta inattesa (% X…)", d[:11])
		}
		if err != nil {
			s.exit()
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, d[11:len(d)-2]...)
	}
	s.exit()
	prog(thuv88MemSize, thuv88MemSize, "Lettura completata")
	return data, fmt.Sprintf("modello «%s»", bytes.TrimRight(id[5:9], "\x00")), nil
}

func (f thuv88Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < thuv88MemSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), thuv88MemSize)
	}
	p, err := openPort(port, m.Baud)
	if err != nil {
		return "", err
	}
	defer p.Close()
	s := &thuv88Sess{ctx: ctx, p: p, o: thuv88o(m)}
	prog(0, thuv88MemSize, "Collegamento con la radio…")
	if _, err := s.start(0xE3); err != nil {
		s.exit()
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < thuv88MemSize; a += thuv88Block {
		if a >= 0x1680 && a < 0x1940 {
			continue // come il programma del costruttore
		}
		if ctx.Err() != nil {
			s.exit()
			return "", errCancelled
		}
		prog(a, thuv88MemSize, "Scrittura…")
		out := append([]byte{byte(a >> 24), byte(a >> 16), byte(a >> 8), byte(a), 0x00, thuv88Block}, img[a:a+thuv88Block]...)
		fr := append(append(append([]byte{0xFE, 0xFE, 0xEE, 0xEF, 0xE4}, out...), thuv88Sum(out)), 0xFD)
		if err := p.Write(fr); err != nil {
			return "", err
		}
		ack, err := s.read(7)
		if err == nil && !bytes.Equal(ack, thuv88Ack) {
			err = rErr("la radio ha rifiutato il blocco (% X)", ack)
		}
		if err != nil {
			s.exit()
			return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
	}
	s.exit()
	prog(thuv88MemSize, thuv88MemSize, "Scrittura completata")
	return "", nil
}
