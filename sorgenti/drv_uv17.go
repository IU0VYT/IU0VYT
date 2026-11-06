// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: baofeng_uv17.py,
// radtel_rt620.py e ga510.py (classe RadioddityGA510v2, 4.7).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia Baofeng UV-17 (driver baofeng_uv17.py di CHIRP): UV-17, UV-13Pro.
// 999 canali, 57600 baud.
//
// La memoria della radio è divisa in pagine da 4 KB numerate (la radio le
// sposta per distribuire l'usura): il programma legge prima la tabella delle
// pagine e poi le pagine 16, 17, 18, 19, 24, 25, 26, 4, 6 in quest'ordine
// (l'immagine è la loro sequenza).
//   identificazione: "PSEARCH" → 0x06 "UV15999"; poi "PASSSTA", "SYSINFO" e
//   altri comandi con le risposte attese; dimensione della memoria
//   (V 00 00 00 0A) e ulteriori comandi.
//   tabella: 'R' indirizzo(3, LE) 1 per ogni pagina → 6 byte (numero in BCD); 0x06.
//   lettura: 'R' indirizzo(3, LE) 0x40 → 5 + 64 byte; 0x06 → 1 byte.
//   scrittura: 'W' indirizzo(3, LE) 0x40 + 64 byte → 0x06.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
)

const uv17Format = `
struct channel {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 unused1;
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:1,
     bcl:1,
     pttid:2,
     unknown2:1,
     wide:1,
     lowpower:1,
     unknown:1;
  u8 scode:4,
     unknown3:3,
     scan:1;
  u8 unknown4;
};
struct channelname {
  char name[11];
};
#seekto 0x0030;
struct {
  struct channel mem[252];
} mem1;
#seek 0x10;
struct {
  struct channel mem[255];
} mem2;
#seek 0x10;
struct {
  struct channel mem[255];
} mem3;
#seek 0x10;
struct {
  struct channel mem[237];
} mem4;
#seekto 0x4000;
struct channelname names1[372];
#seek 0x4;
struct channelname names2[372];
#seek 0x4;
struct channelname names3[255];
`

var uv17BlockOrder = []int{16, 17, 18, 19, 24, 25, 26, 4, 6}

type uv17lOpt struct {
	memsize int // dimensione della memoria dichiarata dalla radio
	// 4.7: altre radio della stessa famiglia (classe UV17 di baofeng_uv17.py):
	// blocks: pagine lette, nell'ordine dell'immagine (BLOCK_O_READ; nil = UV-17);
	// wblocks: pagine scritte (BLOCK_ORDER; nil = blocks)
	blocks, wblocks []int
	// magics, magics2, fingerprint (nil = UV-17); noMemsize: nessuna richiesta
	// della dimensione della memoria (_magic_memsize vuoto, GA-510 V2)
	magics, magics2 []uv17Magic
	finger          []byte
	noMemsize       bool
	format          string // record dei canali (nil = uv17Format)
	ch128           bool   // GA-510 V2: 128 canali (mem1 e names1)
	// narrowBit: MODES = ["NFM", "FM", "WFM"] (Radtel RT-620): wide a 1 = NFM,
	// come CHIRP (_mem.wide = mem.mode == MODES[0])
	narrowBit bool
}

func (o *uv17lOpt) rblocks() []int {
	if o.blocks != nil {
		return o.blocks
	}
	return uv17BlockOrder
}

func (o *uv17lOpt) wblks() []int {
	if o.wblocks != nil {
		return o.wblocks
	}
	return o.rblocks()
}

func (o *uv17lOpt) mg() []uv17Magic {
	if o.magics != nil {
		return o.magics
	}
	return uv17lMagics
}

func (o *uv17lOpt) mg2() []uv17Magic {
	if o.magics2 != nil {
		return o.magics2
	}
	if o.noMemsize {
		return nil
	}
	return uv17lMagics2
}

func (o *uv17lOpt) fp() []byte {
	if o.finger != nil {
		return o.finger
	}
	return uv17lFingerprint
}

// 4.7: Radtel RT-620 (radtel_rt620.py): record senza il byte libero dopo le
// frequenze, potenza e banda su due bit, passo nell'ultimo byte
const rt620Format = `
struct channel {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown12_7:1,
     bcl:1,
     pttid:2,
     wide:2,
     lowpower:2;
  u8 unknown13_5:3,
     compander:1,
     unknown_13_2:2,
     scrambler:2;
  u8 scode:4,
     unknown14_1:3,
     scan:1;
  u8 step;
};
struct channelname {
  char name[11];
};
#seekto 0x0030;
struct {
  struct channel mem[252];
} mem1;
#seek 0x10;
struct {
  struct channel mem[255];
} mem2;
#seek 0x10;
struct {
  struct channel mem[255];
} mem3;
#seek 0x10;
struct {
  struct channel mem[237];
} mem4;
#seekto 0x4000;
struct channelname names1[372];
#seek 0x4;
struct channelname names2[372];
#seek 0x4;
struct channelname names3[255];
`

// 4.7: Radioddity GA-510 V2 (classe RadioddityGA510v2 di ga510.py): 128 canali
const ga510v2Format = `
struct channel {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 unused1;
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:1,
     bcl:1,
     pttid:2,
     unknown2:1,
     wide:1,
     lowpower:2;
  u8 scode:4,
     unknown3:3,
     scan:1;
  u8 unknown4;
};
struct channelname {
  char name[11];
};
#seekto 0x3040;
struct {
  struct channel mem[128];
} mem1;
#seekto 0x400B;
struct channelname names1[128];
`

type uv17lFam struct{}

func init() {
	pw := []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}}
	cs := charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./"
	mk := func(id, model, rclass, name string, memsize int) *genModel {
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: 999, First: 1,
			NameMax: 11, Charset: cs, Powers: pw, Bands: []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true,
			HasScan: true, DCS: uv5rDTCS, Baud: 57600, Fam: uv17lFam{}, Opt: &uv17lOpt{memsize: memsize},
			Notes: "Spegni la radio, collega il cavo, poi riaccendila."}
	}
	genRegister(mk("bf-uv17", "UV-17", "UV17", "Baofeng UV-17", 0xFFFF), mk("bf-uv13pro", "UV-13Pro", "UV13Pro", "Baofeng UV-13Pro", 0x31FFF))

	// 4.7: Radtel RT-620 (radtel_rt620.py): pagine dell'UV-17 più la 2 (taratura,
	// solo letta), identificativo «PROGR6F», tre potenze
	rt620 := mk("rtl-rt620", "RT-620", "RT620", "Radtel RT-620", 0x2EFFF)
	rt620.Vendor = "Radtel"
	rt620.Powers = []genPower{{"Bassa 1 W", 1, 2}, {"Media 5 W", 5, 1}, {"Alta 10 W", 10, 0}}
	rt620.Bands = []band{{108000000, 135999999}, {136000000, 299999999}, {300000000, 559999999}}
	ro := rt620.Opt.(*uv17lOpt)
	ro.blocks = append(append([]int{}, uv17BlockOrder...), 2)
	ro.wblocks = uv17BlockOrder
	ro.finger = []byte("\x06PROGR6F")
	ro.magics2 = []uv17Magic{{[]byte{0x06}, 1}, {[]byte("\xFF\xFF\xFF\xFF\x0CPROGR6F"), 1}, {[]byte{0x02}, 8}, {[]byte{0x06}, 1}}
	ro.format, ro.narrowBit = rt620Format, true
	rt620.Notes = "Spegni la radio, collega il cavo, poi riaccendila. La pagina di taratura si legge e si conserva nella copia, ma non si scrive."
	// 4.7: Radioddity GA-510 V2 (ga510.py, classe RadioddityGA510v2)
	ga := mk("rd-ga510v2", "GA-510", "RadioddityGA510v2", "Radioddity GA-510 V2", 0x10000)
	ga.Vendor, ga.Slots, ga.NameMax = "Radioddity", 128, 8
	ga.Powers = []genPower{{"Bassa 1 W", 1, 2}, {"Media 5 W", 5, 1}, {"Alta 10 W", 10, 0}}
	go2 := ga.Opt.(*uv17lOpt)
	go2.blocks, go2.noMemsize, go2.format, go2.ch128 = []int{2, 4, 6, 16, 24}, true, ga510v2Format, true
	go2.finger = []byte("\x06DMR1702")
	go2.magics = []uv17Magic{{[]byte("PASSSTA"), 3}, {[]byte("SYSINFO"), 1},
		{[]byte{0x56, 0x00, 0x00, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1},
		{[]byte{0x56, 0x00, 0x10, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1},
		{[]byte{0x56, 0x00, 0x20, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1},
		{[]byte{0x56, 0x00, 0x00, 0x00, 0x0A}, 11}, {[]byte{0x06}, 1},
		{[]byte("\xFF\xFF\xFF\xFF\x0CDMR1702"), 1},
		{[]byte{0x02}, 8}, {[]byte{0x06}, 1}}
	genRegister(rt620, ga)
}

func uv17lo(m *genModel) *uv17lOpt { return m.Opt.(*uv17lOpt) }

func (uv17lFam) mem(m *genModel) *bwEl {
	if f := uv17lo(m).format; f != "" {
		return bwMust(f)
	}
	return bwMust(uv17Format)
}

func (f uv17lFam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f uv17lFam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	r := f.mem(m)
	if uv17lo(m).ch128 {
		return r.F("mem1").F("mem").I(i), r.F("names1").I(i)
	}
	var mem, nam *bwEl
	switch {
	case i >= 762:
		mem = r.F("mem4").F("mem").I(i - 762)
	case i >= 507:
		mem = r.F("mem3").F("mem").I(i - 507)
	case i >= 252:
		mem = r.F("mem2").F("mem").I(i - 252)
	default:
		mem = r.F("mem1").F("mem").I(i)
	}
	switch {
	case i >= 744:
		nam = r.F("names3").I(i - 744)
	case i >= 372:
		nam = r.F("names2").I(i - 372)
	default:
		nam = r.F("names1").I(i)
	}
	return mem, nam
}

func (f uv17lFam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(m, i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}

func (f uv17lFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(m, i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}

func (f uv17lFam) Match(m *genModel, img []byte) bool { return false }

// toni in BCD: bit 15 = DCS, bit 14 = invertito
func uv17ToneDec(v int64) string {
	if v == 0 || v == 0xFFFF {
		return ""
	}
	if v&0x8000 != 0 {
		code := (v & 0x0F) + (v>>4&0x0F)*10 + (v>>8&0x0F)*100
		return genTone{Mode: "DTCS", Code: int(code), Inv: v&0xC000 == 0xC000}.String()
	}
	t := (v & 0x0F) + (v>>4&0x0F)*10 + (v>>8&0x0F)*100 + (v>>12&0x0F)*1000
	return fmtTone(int(t))
}

func uv17ToneEnc(t genTone) int64 {
	digits := func(n int) int64 {
		s := fmt.Sprintf("%04d", n)
		return int64(s[0]-'0')<<12 + int64(s[1]-'0')<<8 + int64(s[2]-'0')<<4 + int64(s[3]-'0')
	}
	switch t.Mode {
	case "Tone":
		return digits(t.Hz10)
	case "DTCS":
		v := 0x8000 + digits(t.Code)
		if t.Inv {
			v += 0x4000
		}
		return v & 0xFFFF
	}
	return 0
}

func (f uv17lFam) Get(m *genModel, img []byte, i int) *Chan {
	r, nam := f.recs(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	txr := r.F("txfreq").Raw(img)
	if bytes.Equal(txr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(txr, []byte{0, 0, 0, 0}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = uv17ToneDec(r.F("txtone").Int(img))
	c.RxTone = uv17ToneDec(r.F("rxtone").Int(img))
	c.Scan = r.F("scan").Bool(img)
	lp := int(r.F("lowpower").Int(img))
	if lp >= len(m.Powers) {
		lp = 0 // come CHIRP (IndexError → POWER_LEVELS[0])
	}
	c.Power = m.powerSem(lp)
	c.Narrow = !r.F("wide").Bool(img)
	if uv17lo(m).narrowBit {
		// RT-620: MODES[wide] (0 NFM, 1 FM, 2 WFM); oltre, come la classe UV17Pro
		w := r.F("wide").Int(img)
		c.Narrow = w == 0 || w >= 3
	}
	nb := nam.F("name").Raw(img)
	nm := make([]byte, len(nb))
	for k, b := range nb {
		if b == 0xFF || b == 0 {
			b = ' '
		}
		nm[k] = b
	}
	c.Name = pyName(nm, false)
	return c
}

func (f uv17lFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, nam := f.recs(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		if uv17lo(m).narrowBit {
			// RT-620: come CHIRP (erase_memory), anche il canale cancellato riceve il
			// modo FM e il passo di 5 kHz del canale vuoto predefinito
			r.F("wide").SetInt(img, 1)
			r.F("step").SetInt(img, 1)
		}
		return nil
	}
	wasEmpty := r.Raw(img)[0] == 0xFF
	var bcl, pttid, scode int64 = 0, 0, 1
	var step, comp int64
	if uv17lo(m).narrowBit {
		// RT-620: passo del canale com'è nella memoria (anche se vuoto, come CHIRP)
		if step = r.F("step").Int(img); step >= 8 {
			step = 0
		}
	}
	if !wasEmpty {
		bcl, pttid = r.F("bcl").Int(img), r.F("pttid").Int(img)
		scode = pyMod(r.F("scode").Int(img)-1, 15) + 1
		if uv17lo(m).narrowBit {
			comp = r.F("compander").Int(img)
		}
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	r.Fill(img, 0)
	nam.F("name").SetStr(img, c.Name, 0)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("txtone").SetInt(img, uv17ToneEnc(tt))
	r.F("rxtone").SetInt(img, uv17ToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	if uv17lo(m).narrowBit {
		// RT-620 (radtel_rt620.py): wide = MODES.index(modo), passo e compander
		// conservati, toni assenti a 0xFFFF
		r.F("wide").SetInt(img, map[bool]int64{true: 0, false: 1}[c.Narrow])
		r.F("step").SetInt(img, step)
		r.F("compander").SetInt(img, comp)
		if r.F("rxtone").Int(img) == 0 {
			r.F("rxtone").SetInt(img, 0xFFFF)
		}
		if r.F("txtone").Int(img) == 0 {
			r.F("txtone").SetInt(img, 0xFFFF)
		}
	} else {
		r.F("wide").SetBool(img, !c.Narrow)
	}
	r.F("lowpower").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("bcl").SetInt(img, bcl)
	r.F("pttid").SetInt(img, pttid)
	r.F("scode").SetInt(img, scode)
	return nil
}

// ---------------------------------------------------------------- protocollo

var (
	uv17lMagics = []uv17Magic{{[]byte("PASSSTA"), 3}, {[]byte("SYSINFO"), 1},
		{[]byte{0x56, 0x00, 0x00, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1},
		{[]byte{0x56, 0x00, 0x10, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1},
		{[]byte{0x56, 0x00, 0x20, 0x0A, 0x0D}, 13}, {[]byte{0x06}, 1}}
	uv17lMemsizeCmd = []byte{0x56, 0x00, 0x00, 0x00, 0x0A}
	uv17lMagics2    = []uv17Magic{{[]byte{0x06}, 1},
		{[]byte("\xFF\xFF\xFF\xFF\x0C\x55\x56\x31\x35\x39\x39\x39"), 1},
		{[]byte{0x02}, 8}, {[]byte{0x06}, 1}}
	uv17lFingerprint = []byte("\x06UV15999")
)

func uv17lFrame(cmd byte, addr, n int) []byte {
	return []byte{cmd, byte(addr), byte(addr >> 8), byte(addr >> 16), byte(n)}
}

func (s *uv17Sess) uv17lIdent() error {
	_ = s.p.Purge()
	if err := s.p.Write([]byte("PSEARCH")); err != nil {
		return err
	}
	o := uv17lo(s.m)
	ack, err := s.recv(len(o.fp()))
	if err == errCancelled {
		return err
	}
	if err != nil {
		return rErr("la radio non risponde")
	}
	if !bytes.HasPrefix(ack, o.fp()) {
		return rErr("risposta inattesa dalla radio (è il modello giusto?)")
	}
	for _, mg := range o.mg() {
		if err := s.p.Write(mg.cmd); err != nil {
			return err
		}
		if _, err := s.recv(mg.resp); err != nil {
			return err
		}
	}
	return nil
}

func (s *uv17Sess) send(cmd []byte, n int) ([]byte, error) {
	if err := s.p.Write(cmd); err != nil {
		return nil, err
	}
	return s.recv(n)
}

func (s *uv17Sess) uv17lMap() ([]int, error) {
	o := s.m.Opt.(*uv17lOpt)
	size := o.memsize
	if !o.noMemsize {
		resp, err := s.send(uv17lMemsizeCmd, 11)
		if err != nil {
			return nil, err
		}
		size = int(binary.LittleEndian.Uint32(resp[7:11]))
	}
	for _, mg := range o.mg2() {
		if _, err := s.send(mg.cmd, mg.resp); err != nil {
			return nil, err
		}
	}
	if size != o.memsize {
		return nil, rErr("modello errato o non supportato (dimensione della memoria diversa)")
	}
	var mp []int
	for a := 0x1FFF; a < size+0x1000; a += 0x1000 {
		d, err := s.send(uv17lFrame('R', a, 1), 6)
		if err != nil {
			return nil, err
		}
		b := int(d[5])
		mp = append(mp, (b>>4&0xF)*10+(b&0xF))
		if _, err := s.send([]byte{0x06}, 1); err != nil {
			return nil, err
		}
	}
	return mp, nil
}

func (f uv17lFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := uv17Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	blocks := uv17lo(m).rblocks()
	total := len(blocks) * 0x1000
	prog(0, total, "Collegamento con la radio…")
	if err := s.uv17lIdent(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	mp, err := s.uv17lMap()
	if err != nil {
		return nil, "", genFail("Lettura della tabella della memoria", err)
	}
	var data []byte
	for _, bn := range blocks {
		k := idxInt(mp, bn)
		if k < 0 {
			return nil, "", rErr("la memoria della radio è danneggiata (manca la pagina %d): scrivi nella radio una copia di sicurezza", bn)
		}
		start := (k + 1) * 0x1000
		for a := start; a < start+0x1000; a += 0x40 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			prog(len(data), total, "Lettura…")
			d, err := s.send(uv17lFrame('R', a, 0x40), 0x45)
			if err != nil {
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%05X)", a), err)
			}
			data = append(data, d[5:]...)
			if _, err := s.send([]byte{0x06}, 1); err != nil {
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%05X)", a), err)
			}
		}
	}
	prog(total, total, "Lettura completata")
	return data, "", nil
}

func (f uv17lFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	blocks := uv17lo(m).wblks()
	total := len(blocks) * 0x1000
	if len(img) < total {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := uv17Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, total, "Collegamento con la radio…")
	if err := s.uv17lIdent(); err != nil {
		return "", genFail("Identificazione", err)
	}
	mp, err := s.uv17lMap()
	if err != nil {
		return "", genFail("Lettura della tabella della memoria", err)
	}
	for bi, bn := range blocks {
		if idxInt(mp, bn) < 0 {
			// pagina mancante (scrittura interrotta in precedenza): si usa una pagina libera (165)
			free := idxInt(mp, 165)
			if free < 0 {
				return "", rErr("la memoria della radio è danneggiata e non ci sono pagine libere")
			}
			mp[free] = bn
		}
		start := (idxInt(mp, bn) + 1) * 0x1000
		for a := start; a < start+0x1000; a += 0x40 {
			da := bi*0x1000 + a - start
			prog(da, total, "Scrittura…")
			ack, err := s.send(append(uv17lFrame('W', a, 0x40), img[da:da+0x40]...), 1)
			if err != nil {
				return "", genFail(fmt.Sprintf("Scrittura (0x%05X)", a), err)
			}
			if ack[0] != 0x06 {
				return "", rErr("Scrittura (0x%05X): la radio ha rifiutato il blocco", a)
			}
		}
	}
	prog(total, total, "Scrittura completata")
	return "", nil
}
