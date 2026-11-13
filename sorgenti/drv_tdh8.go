// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: tdh8.py
// (Copyright 2012 Dan Smith e altri autori di CHIRP).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia TIDRADIO TD-H8 (4.1): TD-H8 (seconda e terza generazione), TD-H3,
// versioni HAM e GMRS, Radtel RT-730. 199 canali con nome di 8 caratteri.
//
// Protocollo (tdh8.py di CHIRP), 38400 baud:
//   saluto di 7 byte (es. "PVOJH\x1c\x14") → 0x06; 0x02 → identificativo di 8
//   byte (es. "P31183\xFF\xFF": 83 base, 84 GMRS, 85 HAM); 0x06 → 0x06;
//   lettura: 'R' indirizzo(2, BE) 0x20 → 'W' indirizzo 0x20 + 32 byte + somma;
//   scrittura: 'W' indirizzo 0x20 + 32 byte + somma (byte basso) → 0x06; uscita 'E'.
// L'immagine (come in CHIRP) comincia con l'identificativo di 8 byte: i dati
// della radio sono dall'indirizzo 8. Prima della scrittura si rilegge il blocco
// 0x1B40 (versione del firmware), come CHIRP.

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// record dei canali (byte 13 e 14 come TD-H8; sulla RT-730 alcuni bit hanno un
// altro significato, ma bcl, potenza e banda sono negli stessi punti)
const tdh8ChanFormat = `
#seekto 0x0008;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rxtone[2];
  lbcd txtone[2];
  u8 unused1;
%s
} memory[200];

#seekto 0x0D48;
struct {
  char name[8];%s
} names[200];

#seekto 0x%04X;
lbit usedflags[200];

#seekto 0x%04X;
lbit scanadd[200];
`

// byte 13 e 14 del canale: TD-H8 (MEM_FORMAT), TD-H3 e TD-H8 di terza
// generazione (MEM_FORMAT_H3), RT-730 (MEM_FORMAT_RT730)
var (
	tdh8Bits = `  u8 pttid:2,
     freqhop:1,
     unused3:1,
     unused4:1,
     bcl:1,
     unused5:1,
     unused2:1;
  u8 unused6:1,
     unused7:1,
     lowpower:2,
     wide:1,
     unused8:1,
     offset:2;
  u8 unused10;`
	tdh3Bits = `  u8 pttid:2,
     freqhop:1,
     unused3:1,
     unused4:1,
     bcl:1,
     unused5:1,
     unused2:1;
  u8 unused6:1,
     scramble:1,
     lowpower:2,
     wide:1,
     unused8:1,
     offset:2;
  u8 unused10;`
	rt730Bits = `  u8 unused2:4,
     spec:1,
     bcl:1,
     unused3:2;
  u8 scramble:1,
     freqhop:1,
     lowpower:2,
     wide:1,
     unused4:3;
  u8 unused5;`
)

type tdh8Fam struct{}

type tdh8Opt struct {
	magic   []byte // saluto (CHIRP _idents[0])
	ident   []byte // identificativo atteso (CHIRP ident_mode)
	memsize int    // CHIRP _memsize (dati della radio, senza l'identificativo)
	flags   int    // indirizzo di usedflags (scanadd 0x20 dopo)
	gmrs    bool   // canali 1-30 GMRS fissi, 8-14 a banda stretta e bassa potenza
	pow     []int64
	tx, rx  []band // gamme di trasmissione e di sola ricezione (CHIRP _txbands, _rxbands)
	rt730   bool   // RT-730: senza PTT-ID e salto di frequenza nei canali
	fwRead  bool   // prima della scrittura CHIRP legge la versione del firmware (0x1B40)
	// gmrsDup: regola di CHIRP (validate_memory) sui ponti GMRS: "3154" canali
	// 31-54 simplex o +5 MHz (TD-H8 GMRS), "all" tutti i canali (TD-H8 GMRS di
	// terza generazione), "" nessuna (TD-H3 GMRS)
	gmrsDup string
	// cps26: modelli del programma «TIDRadioCPS» del 2026 (TD-H3 Plus / SE, H7 Plus,
	// H8 di quarta generazione, H9, 638UV), non in CHIRP. Stessa mappa della TD-H3;
	// il programma legge e scrive 0x0000-0x1F3F, non legge la versione del firmware
	// prima di scrivere, chiude con un blocco di 32 byte «CE» e nel canale ha due
	// campi in più (gruppo dello scrambler nel byte 12, modo di ricezione nei bit 0-1
	// del byte 15) che qui si conservano.
	cps26 bool
}

// tdh8End: chiusura del collegamento. CHIRP manda 'E'; il programma del 2026 manda
// un blocco di 32 byte che comincia con «CE» (anche dopo la lettura).
func tdh8End(p Port, o *tdh8Opt, afterRead bool) {
	switch {
	case o.cps26:
		_ = p.Write(append([]byte{'C', 'E'}, make([]byte, 30)...))
	case !afterRead:
		_ = p.Write([]byte{'E'})
	}
}

// frequenze GMRS (bandplan_na.ALL_GMRS_FREQS di CHIRP)
var tdh8GMRSFreqs = []int64{462562500, 462587500, 462612500, 462637500, 462662500, 462687500, 462712500,
	467562500, 467587500, 467612500, 467637500, 467662500, 467687500, 467712500,
	462550000, 462575000, 462600000, 462625000, 462650000, 462675000, 462700000, 462725000}

var tdh8Air = band{108000000, 135999999}

func tdh8o(m *genModel) *tdh8Opt { return m.Opt.(*tdh8Opt) }

func (tdh8Fam) mem(m *genModel) *bwEl {
	o := tdh8o(m)
	bits, extra := tdh3Bits, ""
	switch {
	case o.rt730:
		bits = rt730Bits
	case o.memsize == 0x1eef: // TD-H8 di seconda generazione: nomi di 16 byte
		bits, extra = tdh8Bits, "\n  u8 unknown2[8];"
	}
	return bwMust(fmt.Sprintf(tdh8ChanFormat, bits, extra, o.flags, o.flags+0x20))
}

func (f tdh8Fam) layout(m *genModel) *bwEl { return f.mem(m) }

// canale i (0 = canale 1) → memory[i+1], names[i]
func (f tdh8Fam) rec(m *genModel, i int) *bwEl  { return f.mem(m).F("memory").I(i + 1) }
func (f tdh8Fam) name(m *genModel, i int) *bwEl { return f.mem(m).F("names").I(i).F("name") }

func (f tdh8Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(m, i).Raw(img)...), f.name(m, i).Raw(img)...)
}

func (f tdh8Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(m, i).SetRaw(img, raw[:16])
	f.name(m, i).SetRaw(img, raw[16:24])
}

func (f tdh8Fam) Match(m *genModel, img []byte) bool { return false }

func tdh8In(hz int64, bs []band) bool {
	for _, b := range bs {
		if hz >= b.Lo && hz <= b.Hi {
			return true
		}
	}
	return false
}

// rxOnly: frequenza solo di ricezione (in una gamma _rxbands e non in _txbands)
func (o *tdh8Opt) rxOnly(hz int64) bool { return tdh8In(hz, o.rx) && !tdh8In(hz, o.tx) }

func (f tdh8Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := tdh8o(m)
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF || f.cps26Empty(m, img, i) {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("wide").Bool(img) // bit «wide» a 1 = banda stretta (NFM in CHIRP)
	pv := r.F("lowpower").Int(img)
	c.Power = m.Powers[0].Sem // valore non valido: la più bassa, come CHIRP
	for k, v := range o.pow {
		if v == pv {
			c.Power = m.Powers[k].Sem
		}
	}
	// nome: 0x00 e 0xFF tolti, spazi finali tolti (come CHIRP)
	var nb []byte
	for _, b := range f.name(m, i).Raw(img) {
		if b != 0x00 && b != 0xFF {
			nb = append(nb, b)
		}
	}
	c.Name = pyRstrip(latin1Str(string(nb))) // come str.rstrip() di Python, anche 0x85 e 0xA0
	c.TxTone = t18ToneDec(r.F("txtone").Raw(img))
	c.RxTone = t18ToneDec(r.F("rxtone").Raw(img))
	c.Scan = f.mem(m).F("scanadd").I(i).Bool(img)
	if o.gmrs && i+1 >= 8 && i+1 <= 14 {
		c.Narrow = true
		c.Power = m.Powers[0].Sem
	}
	if o.rxOnly(c.Rx) {
		c.TxOff, c.Tx = true, c.Rx
	}
	if c.Rx >= tdh8Air.Lo && c.Rx <= tdh8Air.Hi {
		c.Narrow = false // AM in CHIRP
	}
	return c
}

// cps26Empty: nei modelli del programma del 2026 un canale è vuoto anche quando non è
// segnato fra quelli usati o ha la ricezione a zero (è così che il programma scrive i
// canali vuoti: 8 byte a 00, toni FF FF FF FF, il resto a 00).
func (f tdh8Fam) cps26Empty(m *genModel, img []byte, i int) bool {
	if !tdh8o(m).cps26 {
		return false
	}
	r := f.rec(m, i)
	return !f.mem(m).F("usedflags").I(i).Bool(img) || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0, 0, 0, 0})
}

func (f tdh8Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := tdh8o(m)
	mm := f.mem(m)
	r := f.rec(m, i)
	wasEmpty := r.Raw(img)[0] == 0xFF || f.cps26Empty(m, img, i)
	mm.F("usedflags").I(i).SetBool(img, c != nil)
	if c == nil {
		if o.cps26 { // come il programma: frequenze a zero, nessun tono, nome vuoto, fuori dalla scansione
			r.SetRaw(img, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0})
			f.name(m, i).Fill(img, 0x00)
			mm.F("scanadd").I(i).SetBool(img, false)
			return nil
		}
		r.Fill(img, 0xFF)
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
	// impostazioni del canale che CHIRP conserva (mem.extra)
	var pttid, bcl, freqhop int64
	old := append([]byte{}, r.Raw(img)...)
	if !wasEmpty {
		bcl = r.F("bcl").Int(img)
		if !o.rt730 {
			pttid, freqhop = r.F("pttid").Int(img), r.F("freqhop").Int(img)
		}
	}
	r.Fill(img, 0x00)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff || o.rxOnly(c.Rx) {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	f.name(m, i).SetStr(img, latin1Bytes(c.Name), 0x00)
	t18ToneEnc(r.F("txtone").Raw(img), tt)
	t18ToneEnc(r.F("rxtone").Raw(img), rt)
	// modo: FM → wide 0, NFM (e AM in banda aerea) → wide 1
	r.F("wide").SetBool(img, c.Narrow || (c.Rx >= tdh8Air.Lo && c.Rx <= tdh8Air.Hi))
	r.F("lowpower").SetInt(img, o.pow[m.powerIndex(c.Power)])
	if o.cps26 && !wasEmpty { // gruppo dello scrambler e modo di ricezione restano quelli del canale
		cur := append([]byte{}, r.Raw(img)...)
		cur[12], cur[15] = old[12], old[15]
		r.SetRaw(img, cur)
	}
	mm.F("scanadd").I(i).SetBool(img, c.Scan)
	if !wasEmpty {
		if !o.cps26 && string(o.ident) == "P31184\xff\xff" && i+1 >= 189 && i+1 <= 199 {
			pttid, bcl, freqhop = 0, 0, 0
		}
		r.F("bcl").SetInt(img, bcl)
		if !o.rt730 {
			r.F("pttid").SetInt(img, pttid)
			r.F("freqhop").SetInt(img, freqhop)
		}
	}
	return nil
}

// lockedSlot: versioni GMRS, canali 1-30 a frequenza fissa (come in CHIRP)
func (f tdh8Fam) lockedSlot(m *genModel, i int) bool { return tdh8o(m).gmrs && i < 30 }

// Locked: canali GMRS fissi (frequenza; 8-14 anche banda e potenza)
func (f tdh8Fam) Locked(m *genModel, i int, old, want *Chan) error {
	o := tdh8o(m)
	if o.gmrs && want != nil {
		// come validate_memory di CHIRP: nei canali 31-54 solo frequenze GMRS e,
		// dove previsto, solo simplex o ponte con +5 MHz
		n := i + 1
		if n >= 31 && n <= 54 && !t18In(tdh8GMRSFreqs, want.Rx) {
			return fmt.Errorf("nei canali 31-54 della %s sono ammesse solo le frequenze GMRS (462,550-467,7125 MHz)", m.Name)
		}
		dupOK := want.TxOff || want.Tx == want.Rx || want.Tx == want.Rx+5000000
		if !dupOK && (o.gmrsDup == "all" || (o.gmrsDup == "3154" && n >= 31 && n <= 54)) {
			return fmt.Errorf("canale %d della %s: solo simplex o ponte GMRS con +5 MHz", n, m.Name)
		}
	}
	if !f.lockedSlot(m, i) || old == nil {
		return nil
	}
	if want == nil || want.Rx != old.Rx {
		return fmt.Errorf("i canali 1-30 della %s sono canali GMRS fissi: frequenza non modificabile", m.Name)
	}
	if i+1 >= 8 && i+1 <= 14 && (!want.Narrow || want.Power != old.Power) {
		return fmt.Errorf("i canali GMRS 8-14 della %s sono sempre a banda stretta e bassa potenza", m.Name)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type tdh8Sess struct {
	ctx context.Context
	p   Port
}

func (s *tdh8Sess) recv(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, to); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

// ident: saluto e identificativo (_do_ident di tdh8.py)
func (s *tdh8Sess) ident(m *genModel) ([]byte, error) {
	o := tdh8o(m)
	if err := s.p.Write(o.magic); err != nil {
		return nil, err
	}
	ack, err := s.recv(1, time.Second)
	if err != nil {
		return nil, err
	}
	if ack[0] != 0x06 {
		return nil, rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	var resp []byte
	for k := 0; k < 8; k++ {
		b, err := s.recv(1, time.Second)
		if err == errCancelled {
			return nil, err
		}
		if err != nil {
			break
		}
		resp = append(resp, b[0])
		if b[0] == 0xDD {
			break
		}
	}
	if len(resp) != 8 {
		return nil, rErr("risposta inattesa della radio (% X)", resp)
	}
	if err := s.p.Write([]byte{0x06}); err != nil {
		return nil, err
	}
	if b, err := s.recv(1, time.Second); err != nil || b[0] != 0x06 {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("la radio ha rifiutato la copia")
	}
	if !bytes.Equal(resp, o.ident) {
		return resp, rErr("la radio si identifica come %q, non come %s: scegli la versione giusta", resp, m.Name)
	}
	return resp, nil
}

func (s *tdh8Sess) read(a int) ([]byte, error) {
	cmd := []byte{'R', byte(a >> 8), byte(a), 0x20}
	if err := s.p.Write(cmd); err != nil {
		return nil, err
	}
	d, err := s.recv(5+0x20, 2*time.Second)
	if err != nil {
		return nil, err
	}
	if d[0] != 'W' || !bytes.Equal(d[1:4], cmd[1:]) {
		return nil, rErr("risposta inattesa al blocco 0x%04X", a)
	}
	return d[4 : 4+0x20], nil
}

func (f tdh8Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := tdh8o(m)
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer p.Close()
	s := &tdh8Sess{ctx: ctx, p: p}
	prog(0, o.memsize, "Collegamento con la radio…")
	if _, err := s.ident(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := append([]byte{}, o.ident...)
	for a := 0; a < o.memsize; a += 0x20 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, o.memsize, "Lettura…")
		b, err := s.read(a)
		if err != nil {
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, b...)
	}
	tdh8End(p, o, true)
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, "", nil
}

func (f tdh8Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := tdh8o(m)
	end := 8 + (o.memsize+0x1F)/0x20*0x20
	if len(img) < end {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), end)
	}
	p, err := openPort(port, m.Baud)
	if err != nil {
		return "", err
	}
	defer p.Close()
	s := &tdh8Sess{ctx: ctx, p: p}
	prog(0, o.memsize, "Collegamento con la radio…")
	if _, err := s.ident(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	if o.fwRead {
		if _, err := s.read(0x1B40); err != nil {
			return "", genFail("Versione del firmware", err)
		}
	}
	for a := 0; a < o.memsize; a += 0x20 {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(a, o.memsize, "Scrittura…")
		d := img[a+8 : a+40]
		var sum byte
		for _, b := range d {
			sum += b
		}
		pk := append(append([]byte{'W', byte(a >> 8), byte(a), 0x20}, d...), sum)
		if err := p.Write(pk); err != nil {
			return "", err
		}
		ack, err := s.recv(1, 2*time.Second)
		if err == nil && ack[0] != 0x06 {
			err = rErr("la radio ha rifiutato il blocco")
		}
		if err != nil {
			return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
	}
	tdh8End(p, o, false)
	prog(o.memsize, o.memsize, "Scrittura completata")
	return "", nil
}

func init() {
	const (
		tdH8   = "PVOJH\x1c\x14"
		tdH3   = "PVOJH\x5c\x14"
		rt730  = "PGOJH\xc3D"
		tdH8G3 = "PVOJH<\x14"
	)
	mhz := func(l ...float64) []band {
		var out []band
		for k := 0; k+1 < len(l); k += 2 {
			out = append(out, band{int64(math.Round(l[k] * 1e6)), int64(math.Round(l[k+1] * 1e6))})
		}
		return out
	}
	p3 := func(l, m, h float64) ([]genPower, []int64) {
		return []genPower{{"Bassa " + fmtW(l), l, 2}, {"Media " + fmtW(m), m, 1}, {"Alta " + fmtW(h), h, 0}}, []int64{0, 1, 2}
	}
	p2 := func(l, h float64) ([]genPower, []int64) {
		return []genPower{{"Bassa " + fmtW(l), l, 2}, {"Alta " + fmtW(h), h, 0}}, []int64{0, 1}
	}
	type v struct {
		id, vendor, model, rclass, name string
		magic, ident                    string
		memsize, flags                  int
		powers                          []genPower
		pow                             []int64
		tx, rx                          []band
		gmrs, rt                        bool
	}
	h8p, h8v := p3(1, 4, 8)
	h8gp, h8gv := p2(1, 8)
	h3p, h3v := p2(2, 5)
	g3p, g3v := p3(2, 6, 10)
	rtp, rtv := p3(1, 4, 8)
	hamTx, gmrsTx := mhz(144, 149, 420, 451), mhz(136, 175, 400, 521)
	h3HamRx := mhz(18, 107.999, 108, 136, 149.99, 419.99, 451, 600)
	list := []v{
		{"td-tdh8", "TIDRADIO", "TD-H8", "TDH8", "TIDRADIO TD-H8", tdH8, "P31183\xff\xff", 0x1eef, 0x1A08, h8p, h8v, mhz(136, 175, 400, 521), nil, false, false},
		// 4.6: TID TD-UV68 (classe UV68 di tdh8.py: la TD-H8 con un altro nome)
		{"td-uv68", "TID", "TD-UV68", "UV68", "TID TD-UV68", tdH8, "P31183\xff\xff", 0x1eef, 0x1A08, h8p, h8v, mhz(136, 175, 400, 521), nil, false, false},
		{"td-tdh8ham", "TIDRADIO", "TD-H8-HAM", "TDH8_HAM", "TIDRADIO TD-H8 HAM", tdH8, "P31185\xff\xff", 0x1eef, 0x1A08, h8p, h8v, hamTx,
			mhz(136, 143.999, 149.000001, 174, 400, 419.999, 451.000001, 521), false, false},
		{"td-tdh8gmrs", "TIDRADIO", "TD-H8-GMRS", "TDH8_GMRS", "TIDRADIO TD-H8 GMRS", tdH8, "P31184\xff\xff", 0x1eef, 0x1A08, h8gp, h8gv, gmrsTx, nil, true, false},
		{"td-tdh8g3", "TIDRADIO", "TD-H8", "TDH8_3rd_Gen", "TIDRADIO TD-H8 (3ª generazione)", tdH8G3, "P31183\xff\xff", 0x1fef, 0x1908, g3p, g3v,
			mhz(136, 600), mhz(18, 107.999, 108, 136), false, false},
		{"td-tdh8g3ham", "TIDRADIO", "TD-H8-HAM", "TDH8_3rd_Gen_HAM", "TIDRADIO TD-H8 HAM (3ª generazione)", tdH8G3, "P31185\xff\xff", 0x1fef, 0x1908, g3p, g3v,
			hamTx, h3HamRx, false, false},
		{"td-tdh8g3gmrs", "TIDRADIO", "TD-H8-GMRS", "TDH8_3rd_Gen_GMRS", "TIDRADIO TD-H8 GMRS (3ª generazione)", tdH8G3, "P31184\xff\xff", 0x1fef, 0x1908, g3p, g3v,
			gmrsTx, mhz(18, 107.999, 108, 136), true, false},
		{"td-tdh3", "TIDRADIO", "TD-H3", "TDH3", "TIDRADIO TD-H3", tdH3, "P31183\xff\xff", 0x1fef, 0x1908, h3p, h3v, mhz(136, 600), mhz(18, 107.999, 108, 136), false, false},
		{"td-tdh3ham", "TIDRADIO", "TD-H3-HAM", "TDH3_HAM", "TIDRADIO TD-H3 HAM", tdH3, "P31185\xff\xff", 0x1fef, 0x1908, h3p, h3v, hamTx, h3HamRx, false, false},
		{"td-tdh3gmrs", "TIDRADIO", "TD-H3-GMRS", "TDH3_GMRS", "TIDRADIO TD-H3 GMRS", tdH3, "P31184\xff\xff", 0x1fef, 0x1908, h3p, h3v, gmrsTx,
			mhz(18, 107.999, 108, 136), true, false},
		{"rt-rt730", "Radtel", "RT-730", "RT730", "Radtel RT-730", rt730, "P31183\xff\xff", 0x6400, 0x1A08, rtp, rtv,
			mhz(136, 174, 174, 300, 300, 400, 400, 520, 520, 630), mhz(10, 108, 108, 136), false, true},
	}
	for _, x := range list {
		bands := append(append([]band{}, x.tx...), x.rx...)
		o := &tdh8Opt{magic: []byte(x.magic), ident: []byte(x.ident), memsize: x.memsize, flags: x.flags, gmrs: x.gmrs,
			pow: x.pow, tx: x.tx, rx: x.rx, rt730: x.rt, fwRead: true,
			gmrsDup: map[string]string{"td-tdh8gmrs": "3154", "td-tdh8g3gmrs": "all"}[x.id]}
		note := "Spegni la radio, collega il cavo, accendila con il volume alto, su un canale libero. Driver di CHIRP della famiglia TIDRADIO TD-H8. " +
			"Le versioni (base, HAM, GMRS) si distinguono dall'identificativo comunicato dalla radio: se non è quella giusta la radio non viene letta."
		if x.gmrs {
			note += " Radio GMRS (USA): canali 1-30 a frequenza fissa, 8-14 a banda stretta e bassa potenza."
		}
		genRegister(&genModel{ID: x.id, Vendor: x.vendor, Model: x.model, RClass: x.rclass,
			Name: x.name, Slots: 199, First: 1, NameMax: 8,
			Charset: charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./",
			Powers:  x.powers, Bands: bands, Narrow: true, HasScan: true,
			DCS: dtcsCodes, Baud: 38400, Fam: tdh8Fam{}, Opt: o, Notes: note})
	}
}

// Modelli del programma «TIDRadioCPS» (versioni dal 23/1 al 27/7/2026), che non sono
// in CHIRP. Il programma è uno solo per tutti: stesso saluto della TD-H3, canali da
// 0x0010, nomi da 0x0D40, canali usati a 0x1900 e scansione a 0x1920 (cioè la mappa
// della TD-H3 di CHIRP), tre potenze (0 bassa, 1 media, 2 alta), bit 3 del byte 14 a
// 1 = banda stretta. Identificativi presi dalla tabella del programma
// (GetSupportedVersions); i limiti di frequenza da ValidateFrequencyByDeviceModel
// (base: 108-174 e 200-600 MHz; HAM: trasmissione 144-148 e 420-450 MHz), con la
// banda aerea 108-136 MHz in sola ricezione come nella TD-H3. TD-H3 Plus e TD-H3 SE
// si presentano come la TD-H3 (P31183 / P31185 / P31184): la differenza è nelle tre
// potenze. I modelli H4 e H11 dello stesso programma (banda larga 18-800 MHz) non
// hanno un file di fabbrica nel pacchetto e non sono stati aggiunti.
func init() {
	const magic = "PVOJH\x5c\x14"
	mhz := func(l ...float64) []band {
		var out []band
		for k := 0; k+1 < len(l); k += 2 {
			out = append(out, band{int64(math.Round(l[k] * 1e6)), int64(math.Round(l[k+1] * 1e6))})
		}
		return out
	}
	rx := mhz(108, 174, 200, 600)
	type v struct{ id, model, name, n, h, g string }
	list := []v{
		{"td-h3plus", "TD-H3 Plus", "TIDRADIO TD-H3 Plus / TD-H3 SE", "P31183\xff\xff", "P31185\xff\xff", "P31184\xff\xff"},
		{"td-h7plus", "TD-H7 Plus", "TIDRADIO TD-H7 Plus", "TDH7PLUN", "TDH7PLUH", "TDH7PLUG"},
		{"td-h8g4", "TD-H8", "TIDRADIO TD-H8 (4ª generazione)", "TDH84GEN", "TDH84GEH", "TDH84GEG"},
		{"td-h9", "TD-H9", "TIDRADIO TD-H9", "TDH9\xff\xff\xffN", "TDH9\xff\xff\xffH", "TDH9\xff\xff\xffG"},
		{"td-638uv", "TD-638UV", "TIDRADIO TD-638UV", "TD638UVN", "TD638UVH", "TD638UVG"},
	}
	for _, x := range list {
		for k, ident := range []string{x.n, x.h, x.g} {
			id, name, tx := x.id, x.name, mhz(136, 174, 200, 600)
			switch k {
			case 1:
				id, name, tx = id+"ham", name+" HAM", mhz(144, 148, 420, 450)
			case 2:
				id, name = id+"gmrs", name+" GMRS"
			}
			o := &tdh8Opt{magic: []byte(magic), ident: []byte(ident), memsize: 0x1F30, flags: 0x1908,
				pow: []int64{0, 1, 2}, tx: tx, rx: rx, cps26: true}
			genRegister(&genModel{ID: id, Vendor: "TIDRADIO", Model: x.model, RClass: "TDH3", NoChirp: true,
				Name: name, Slots: 199, First: 1, NameMax: 8,
				Charset: charsetAlnum + "!@#$%^&*()+-=[]:\";'<>?,./",
				Powers:  []genPower{{"Bassa", 0, 2}, {"Media", 0, 1}, {"Alta", 0, 0}},
				Bands:   append(append([]band{}, tx...), rx...), Narrow: true, HasScan: true,
				DCS: dtcsCodes, Baud: 38400, Fam: tdh8Fam{}, Opt: o,
				Notes: "Spegni la radio, collega il cavo, accendila con il volume alto, su un canale libero. Non è in CHIRP: collegamento e memoria " +
					"vengono dal programma del costruttore «TIDRadioCPS» (2026), che usa la stessa mappa della TD-H3 di CHIRP, e sono provati con una radio " +
					"simulata e con i file di fabbrica del programma, non su una radio vera. Canali (frequenze, toni, potenza, banda, scansione, nome) e impostazioni generali; " +
					"blocco su canale occupato, PTT-ID, salto di frequenza, scrambler e modo di ricezione restano quelli del canale. Le versioni (base, HAM, " +
					"GMRS) si distinguono dall'identificativo comunicato dalla radio: se non è quella giusta la radio non viene letta."})
		}
	}
}

// pyRstrip: come str.rstrip() di Python (spazi e separatori ASCII 0x1C-0x1F)
func pyRstrip(s string) string {
	return strings.TrimRightFunc(s, func(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1C && r <= 0x1F) })
}
