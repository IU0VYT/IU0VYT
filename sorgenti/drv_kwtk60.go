// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tk270.py, tk760g.py (4.17).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.17: palmari professionali Kenwood delle serie 60 e 60G.
//
// Serie 60 (tk270.py di CHIRP): TK-260, TK-270, TK-272, TK-278 (VHF), TK-360,
// TK-370, TK-372, TK-378 (UHF). Memoria di 0x400 byte, fino a 32 canali senza
// nome. 9600 baud 8N1: «PROGRAM» → 06; 02 → sigla di 8 byte; 06 → 06; blocchi di
// 8 byte: 'R' indirizzo(2) 08 → 4 byte di intestazione e 8 di dati, 06 → 06;
// 'W' indirizzo(2) 08 dati → 06. Da 0x3D8 in poi la memoria non si scrive.
//
// Serie 60G (tk760g.py di CHIRP): TK-260G, TK-270G, TK-272G, TK-278G (VHF),
// TK-360G, TK-370G, TK-372G, TK-378G, TK-388G (UHF). Memoria di 0x8000 byte,
// fino a 128 canali con nome di 8 caratteri. 9600 baud, parità pari: «PROGRAM»
// → 06; 02 → sigla; 06 → 06; blocchi di 256 byte: 'R' numero(2) → 'W' dati somma
// (oppure 'Z' e un byte per un blocco vuoto), 06 → 06; 'W' numero(2) dati somma
// → 06 ('Z' numero(2) FF per un blocco vuoto); 'E' per uscire.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------- serie 60

type kwp60Fam struct{}

type kwtkOpt struct {
	typ  string // sigla del modello nella risposta della radio
	vars []kwVar
	ro52 bool // serie 60G: blocchi da 0x52 non scrivibili (TK-260G)
}

const kwp60Format = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown41:1,
     unknown42:1,
     power:1,
     shift:1,
     busy:1,
     unknown46:1,
     unknown47:1,
     unknown48:1;
  u8 rxen;
  u8 txen;
  u8 unknown7;
} memory[32];
#seekto 0x0338;
u8 scan[4];
u8 active[4];
`

const (
	kwp60Size  = 0x400
	kwp60ID    = 0x3D0 // sigla della radio, a rovescio
	kwp60Flags = 0x280 // segnalazione DTMF / 2 toni dei canali
	kwp60Top   = 0x3D0 // ultimo blocco che si scrive
)

var kwtkPowers = []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}}

const kwtkHow = "Radio professionale: si programma col cavo di programmazione Kenwood, accendendo la radio dopo aver collegato il cavo. " +
	"La versione della radio (banda e numero dei canali) si legge dalla sua memoria; i limiti di frequenza sono quelli del " +
	"costruttore allargati del 4%, come in CHIRP. "

// kwtkBands: limiti di tutte le versioni del modello, allargati del 4% come CHIRP
func kwtkBand(v kwVar) band {
	return band{int64(float64(v.lo*1000000) * 0.96), int64(float64(v.hi*1000000) * 1.04)}
}

func kwtkBands(vars []kwVar) []band {
	var out []band
	for _, v := range vars {
		b := kwtkBand(v)
		merged := false
		for k := range out {
			if b.Lo <= out[k].Hi && b.Hi >= out[k].Lo {
				if b.Lo < out[k].Lo {
					out[k].Lo = b.Lo
				}
				if b.Hi > out[k].Hi {
					out[k].Hi = b.Hi
				}
				merged = true
			}
		}
		if !merged {
			out = append(out, b)
		}
	}
	return out
}

func kwtkSlots(vars []kwVar) int {
	n := 0
	for _, v := range vars {
		if v.n > n {
			n = v.n
		}
	}
	return n
}

func init() {
	p60 := func(id, model, class, typ string) *genModel {
		vars := kwTkVariants[class]
		n := kwtkSlots(vars)
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: n, First: 1, Powers: kwtkPowers, Bands: kwtkBands(vars), HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kwp60Fam{}, Opt: &kwtkOpt{typ: typ, vars: vars},
			Notes: "Canali senza nome. " + kwtkHow + "Blocco del canale occupato e «beat shift» si conservano; " +
				"la segnalazione DTMF / 2 toni dei canali viene azzerata a ogni scrittura, come fa CHIRP. Driver di CHIRP (sperimentale)."}
	}
	genRegister(
		p60("kw-tk260", "TK-260", "TK260_Radio", "P0260"),
		p60("kw-tk270", "TK-270", "TK270_Radio", "P0270"),
		p60("kw-tk272", "TK-272", "TK272_Radio", "P0272"),
		p60("kw-tk278", "TK-278", "TK278_Radio", "P0278"),
		p60("kw-tk360", "TK-360", "TK360_Radio", "P0360"),
		p60("kw-tk370", "TK-370", "TK370_Radio", "P0370"),
		p60("kw-tk372", "TK-372", "TK372_Radio", "P0372"),
		p60("kw-tk378", "TK-378", "TK378_Radio", "P0378"),
	)
	g := func(id, model, class, typ string) *genModel {
		vars := kwTkVariants[class]
		n := kwtkSlots(vars)
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: class, Name: "Kenwood " + model,
			Slots: n, First: 1, NameMax: 8, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890_-*()/\\+=",
			Powers: kwtkPowers, Bands: kwtkBands(vars), Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: kw60gFam{}, Opt: &kwtkOpt{typ: typ, vars: vars, ro52: class == "TK260G_Radios"},
			Notes: "Canali con nomi di 8 caratteri. " + kwtkHow + "Gruppo del canale, compander, blocco del " +
				"canale occupato e «beat shift» si conservano; i canali nuovi vanno nel gruppo 1. A ogni modifica dei canali le " +
				"tabelle dei gruppi vengono ricostruite e la segnalazione DTMF / 2 toni dei canali azzerata, come fa CHIRP. " +
				"Driver di CHIRP (sperimentale)."}
	}
	genRegister(
		g("kw-tk260g", "TK-260G", "TK260G_Radios", "P2600"),
		g("kw-tk270g", "TK-270G", "TK270G_Radios", "P2700"),
		g("kw-tk272g", "TK-272G", "TK272G_Radios", "P2720"),
		g("kw-tk278g", "TK-278G", "TK278G_Radios", "P2780"),
		g("kw-tk360g", "TK-360G", "TK360G_Radios", "P3600"),
		g("kw-tk370g", "TK-370G", "TK370G_Radios", "P3700"),
		g("kw-tk372g", "TK-372G", "TK372G_Radios", "P3720"),
		g("kw-tk378g", "TK-378G", "TK378G_Radios", "P3780"),
		g("kw-tk388g", "TK-388G", "TK388G_Radios", "P3880"),
	)
}

func kwtko(m *genModel) *kwtkOpt { return m.Opt.(*kwtkOpt) }

// kwtkFind: versione con la sigla data
func kwtkFind(m *genModel, id []byte) *kwVar {
	o := kwtko(m)
	for k := range o.vars {
		if string(id) == o.vars[k].id {
			return &o.vars[k]
		}
	}
	return nil
}

func kwp60Ident(img []byte) []byte {
	id := make([]byte, 8)
	for k := range id {
		id[k] = img[kwp60ID+7-k]
	}
	return id
}

func (kwp60Fam) variant(m *genModel, img []byte) *kwVar {
	if len(img) < kwp60Size {
		return nil
	}
	return kwtkFind(m, kwp60Ident(img))
}

func (kwp60Fam) mem() *bwEl                 { return bwMust(kwp60Format) }
func (f kwp60Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f kwp60Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f kwp60Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f kwp60Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f kwp60Fam) Match(m *genModel, img []byte) bool {
	return len(img) == kwp60Size && f.variant(m, img) != nil
}

func (f kwp60Fam) Bands(m *genModel, img []byte) []band {
	if v := f.variant(m, img); v != nil {
		return []band{kwtkBand(*v)}
	}
	return m.Bands
}

func kwtkBit(img []byte, at, i int) bool { return img[at+i/8]>>(i%8)&1 == 1 }
func kwtkSetBit(img []byte, at, i int, v bool) {
	if v {
		img[at+i/8] |= 1 << (i % 8)
	} else {
		img[at+i/8] &^= 1 << (i % 8)
	}
}

func (f kwp60Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	raw := r.Raw(img)
	if raw[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	c.Tx = c.Rx
	if raw[4] == 0xFF || raw[14] == 0xFF {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = x3ToneDec(r.F("tx_tone").Int(img))
	c.RxTone = x3ToneDec(r.F("rx_tone").Int(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Scan = kwtkBit(img, f.mem().F("scan").Off, i)
	return c
}

func (f kwp60Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	act := f.mem().F("active").Off
	if c == nil {
		r.Fill(img, 0xFF)
		kwtkSetBit(img, act, i, true) // 1: canale non attivo
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	if r.Raw(img)[0] == 0xFF {
		r.Fill(img, 0xFF) // posizione vuota: via anche i resti di un canale cancellato
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	r.F("rxen").SetInt(img, 0)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
		r.F("txen").SetInt(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
		r.F("txen").SetInt(img, 0)
	}
	if err := x3ToneEnc(r.F("tx_tone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rx_tone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	kwtkSetBit(img, f.mem().F("scan").Off, i, c.Scan)
	kwtkSetBit(img, act, i, false)
	return nil
}

// kwtkEnter: «PROGRAM», sigla della radio, conferma. Restituisce la sigla.
func kwtkEnter(s *rawSess, m *genModel, exit bool) ([]byte, error) {
	fail := func(e error) ([]byte, error) {
		if exit {
			s.send([]byte{'E'})
		}
		return nil, e
	}
	s.p.Purge()
	if err := s.send([]byte("PROGRAM")); err != nil {
		return nil, err
	}
	b := readUpTo(s.ctx, s.p, 1, time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return fail(errTimeout)
	}
	if b[0] != 0x06 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
	}
	if err := s.send([]byte{0x02}); err != nil {
		return nil, err
	}
	id := readUpTo(s.ctx, s.p, 8, time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(id) != 8 {
		return fail(errTimeout)
	}
	typ := kwtko(m).typ
	if !strings.Contains(strings.ToUpper(string(id)), typ) {
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwtkOpt); ok && strings.Contains(strings.ToUpper(string(id)), xo.typ) {
				return fail(rErr("la radio è un altro modello: apri la scheda «%s»", x.Name))
			}
		}
		return fail(rErr("la radio è un altro modello (%s)", strings.TrimRight(string(id[:5]), "\xff\x00")))
	}
	if err := s.ack2(); err != nil {
		return fail(err)
	}
	return id, nil
}

func (f kwp60Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, kwp60Size, "Collegamento con la radio…")
	if _, err := kwtkEnter(s, m, false); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < kwp60Size; a += 8 {
		prog(a, kwp60Size, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 8}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 12, time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(b) != 12 || b[3] != 8 {
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.ack2(); err != nil {
			return nil, "", err
		}
		data = append(data, b[4:]...)
	}
	if f.variant(m, data) == nil {
		return nil, "", rErr("versione della radio sconosciuta (% X)", kwp60Ident(data))
	}
	prog(kwp60Size, kwp60Size, "Lettura completata")
	return data, "", nil
}

func (f kwp60Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	v := f.variant(m, img)
	if v == nil {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	out := append([]byte{}, img...)
	// come CHIRP: la segnalazione DTMF / 2 toni dei canali (5 byte per canale) si azzera
	for k := 0; k < 5*v.n; k++ {
		out[kwp60Flags+k] = 0xFF
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, kwp60Top+8, "Collegamento con la radio…")
	if _, err := kwtkEnter(s, m, false); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a <= kwp60Top; a += 8 {
		prog(a, kwp60Top+8, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 8}, out[a:a+8]...)); err != nil {
			return "", err
		}
		if err := s.sleep(20 * time.Millisecond); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 1, time.Second)
		if ctx.Err() != nil {
			return "", errCancelled
		}
		if len(b) != 1 || b[0] != 0x06 {
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(kwp60Top+8, kwp60Top+8, "Scrittura completata")
	return "", nil
}

// ---------------------------------------------------------------- serie 60G

type kw60gFam struct{}

const kw60gFormat = `
struct {
  u8 unknown0[14];
  u8 banks;
  u8 channels;
  ul16 tot;
  u8 tot_rekey;
  u8 unknown1;
  u8 tot_reset;
  u8 unknown2;
  u8 tot_alert;
  u8 unknown3[7];
  u8 sql_level;
  u8 battery_save;
  u8 unknown4[10];
  u8 unknown5:3,
     c2t:1,
     unknown6:4;
  u8 unknown7[5];
  u8 unknown8[16];
  u8 unknown9[16];
  u8 unknown10[16];
  u8 unknown11[16];
  u8 add[16];
} settings;
#seekto 0x2000;
struct {
  u8 bnumb;
  u8 bank;
  char name[8];
  u8 unknown20[2];
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 rx_unkw;
  u8 tx_unkw;
  ul16 rx_tone;
  ul16 tx_tone;
  u8 unknown23[5];
  u8 signaling;
  u8 ptt_id:2,
     beat_shift:1,
     unknown26:2,
     power:1,
     compander:1,
     wide:1;
  u8 unknown27:6,
     busy_lock:1,
     unknown28:1;
  u8 unknown29[14];
} memory[128];
`

const (
	kw60gSize  = 0x8000
	kw60gID    = 0xA7   // sigla della versione (7 byte)
	kw60gSoft  = 0xBB   // sigla del programma che ha scritto la radio (5 byte)
	kw60gTable = 0x300  // tabella gruppo / posizione / canale
	kw60gBanks = 0x1000 // gruppi: numero e canali di ognuno
	kw60gDTMF  = 0x7000 // segnalazione dei canali (5 byte per canale)
)

func (kw60gFam) variant(m *genModel, img []byte) *kwVar {
	if len(img) < kw60gSize {
		return nil
	}
	return kwtkFind(m, img[kw60gID:kw60gID+7])
}

func (kw60gFam) mem() *bwEl                 { return bwMust(kw60gFormat) }
func (f kw60gFam) layout(m *genModel) *bwEl { return f.mem() }
func (f kw60gFam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f kw60gFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f kw60gFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f kw60gFam) Match(m *genModel, img []byte) bool {
	return len(img) == kw60gSize && f.variant(m, img) != nil
}

func (f kw60gFam) Bands(m *genModel, img []byte) []band {
	if v := f.variant(m, img); v != nil {
		return []band{kwtkBand(*v)}
	}
	return m.Bands
}

// count: canali della versione della radio
func (f kw60gFam) count(m *genModel, img []byte) int {
	if v := f.variant(m, img); v != nil {
		return v.n
	}
	return m.Slots
}

// i canali oltre quelli della versione della radio non esistono
func (f kw60gFam) lockedSlotImg(m *genModel, img []byte, i int) bool { return i >= f.count(m, img) }

func (f kw60gFam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	if n := f.count(m, img); i >= n && want != nil {
		return fmt.Errorf("questa versione della radio ha %d canali", n)
	}
	return nil
}

func kw60gToneDec(v int64) string {
	switch {
	case v == 0xFFFF:
		return ""
	case v >= 0x2800:
		var code int
		fmt.Sscanf(fmt.Sprintf("%03o", v&0x7FF), "%d", &code)
		return genTone{Mode: "DTCS", Code: code, Inv: v&0x8000 != 0}.String()
	}
	return fmtTone(int(v))
}

func kw60gToneEnc(e *bwEl, img []byte, s string) error {
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
		v += 0x2800
		if t.Inv {
			v += 0xA000
		}
		e.SetInt(img, v)
	default:
		e.Fill(img, 0xFF)
	}
	return nil
}

func (f kw60gFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	raw := r.Raw(img)
	if raw[12] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	c.Tx = c.Rx
	if raw[16] == 0xFF {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = strings.TrimRight(r.F("name").Str(img), " \xff")
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = !r.F("wide").Bool(img)
	c.Scan = kwtkBit(img, f.mem().F("settings").F("add").Off, i)
	c.TxTone = kw60gToneDec(r.F("tx_tone").Int(img))
	c.RxTone = kw60gToneDec(r.F("rx_tone").Int(img))
	return c
}

func (f kw60gFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		f.prep(m, img)
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if r.F("rx_unkw").Int(img) == 0xFF {
		r.F("rx_unkw").SetInt(img, 0x35)
		r.F("tx_unkw").SetInt(img, 0x32)
	}
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
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
	kwtkSetBit(img, f.mem().F("settings").F("add").Off, i, c.Scan)
	// canale nuovo: gruppo 1. Un gruppo non valido (0 o oltre 127) diventa 1, come in CHIRP
	if r.F("bnumb").Int(img) == 0xFF {
		r.F("bnumb").SetInt(img, int64(i))
		r.F("bank").SetInt(img, 1)
	}
	if b := r.F("bank").Int(img); b == 0 || b > 127 {
		r.F("bank").SetInt(img, 1)
	}
	f.prep(m, img)
	return nil
}

// prep: tabelle dei gruppi e dei canali ricostruite dai canali (_prep_data di CHIRP)
func (f kw60gFam) prep(m *genModel, img []byte) {
	n := f.count(m, img)
	var order []int // gruppi nell'ordine in cui compaiono
	chans := map[int][]int{}
	total := 0
	for ch := 0; ch < n; ch++ {
		r := f.rec(ch)
		bnumb, bank := int(r.F("bnumb").Int(img)), int(r.F("bank").Int(img))
		if bnumb == 255 || bank == 255 || bank == 0 {
			continue
		}
		if _, ok := chans[bank]; !ok {
			order = append(order, bank)
		}
		chans[bank] = append(chans[bank], ch)
		total++
	}
	img[0x0F] = byte(total)
	img[0x0E] = byte(len(order))
	var fdata []byte
	for _, k := range order {
		for c, ch := range chans[k] {
			fdata = append(fdata, byte(k), byte(c+1), byte(k-1), byte(ch))
		}
	}
	fdata = append(fdata, bytes.Repeat([]byte{0xFF}, len(fdata)%256)...)
	copy(img[kw60gTable:], fdata)
	bdata := bytes.Repeat([]byte{0xFF}, 0x1000)
	sorted := append([]int{}, order...)
	for a := 1; a < len(sorted); a++ {
		for b := a; b > 0 && sorted[b] < sorted[b-1]; b-- {
			sorted[b], sorted[b-1] = sorted[b-1], sorted[b]
		}
	}
	for k, bank := range sorted {
		bdata[k*16], bdata[k*16+1] = byte(bank), byte(len(chans[bank]))
	}
	copy(img[kw60gBanks:], bdata)
	for ch := 0; ch < n; ch++ {
		rec := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		if ch < total {
			rec = []byte{0x00, 0xF0, 0xFF, 0xFF, 0xFF}
		}
		copy(img[kw60gDTMF+ch*5:], rec)
	}
}

func kwtkSum(b []byte) byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return s
}

// roBlock: blocchi che CHIRP non scrive (da 0x10 a 0x1E solo se vuoti)
func (f kw60gFam) skipBlock(m *genModel, blk int, data []byte) bool {
	empty := bytes.Equal(data, bytes.Repeat([]byte{0xFF}, 256))
	lo := 0x59
	if kwtko(m).ro52 {
		lo = 0x52
	}
	switch {
	case blk >= 0x10 && blk <= 0x1E:
		return empty
	case blk >= lo && blk <= 0x5E:
		return true
	}
	return false
}

func (f kw60gFam) open(ctx context.Context, m *genModel, port string) (*rawSess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	if err := setEvenParity(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	if _, err := kwtkEnter(s, m, true); err != nil {
		s.p.Close()
		return nil, genFail("Identificazione", err)
	}
	if err := s.sleep(time.Second); err != nil {
		s.p.Close()
		return nil, err
	}
	return s, nil
}

func (f kw60gFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, kw60gSize, "Collegamento con la radio…")
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
	for blk := 0; blk < kw60gSize/256; blk++ {
		prog(blk*256, kw60gSize, "Lettura…")
		s.p.Purge()
		if err := s.send([]byte{'R', byte(blk >> 8), byte(blk)}); err != nil {
			return nil, "", err
		}
		if err := s.sleep(100 * time.Millisecond); err != nil {
			return fail(err)
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
	s.send([]byte{'E'})
	if f.variant(m, data) == nil {
		return nil, "", rErr("versione della radio sconosciuta (% X)", data[kw60gID:kw60gID+7])
	}
	prog(kw60gSize, kw60gSize, "Lettura completata")
	return data, "", nil
}

func (f kw60gFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, kw60gSize, "Collegamento con la radio…")
	s, err := f.open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	for blk := 0; blk < kw60gSize/256; blk++ {
		if ctx.Err() != nil {
			s.send([]byte{'E'})
			return "", errCancelled
		}
		prog(blk*256, kw60gSize, "Scrittura…")
		data := img[blk*256 : blk*256+256]
		if f.skipBlock(m, blk, data) {
			continue
		}
		frame := []byte{'Z', byte(blk >> 8), byte(blk), 0xFF}
		if !bytes.Equal(data, bytes.Repeat([]byte{0xFF}, 256)) {
			frame = append([]byte{'W', byte(blk >> 8), byte(blk)}, data...)
			frame = append(frame, kwtkSum(data))
		}
		if err := s.send(frame); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if len(b) != 1 || b[0] != 0x06 {
			s.send([]byte{'E'})
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", blk*256)
		}
	}
	s.send([]byte{'E'})
	prog(kw60gSize, kw60gSize, "Scrittura completata")
	return "", nil
}
