// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// kguv8d.py, kguv8dplus.py, kguv8e.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Wouxun KG-UV8D, KG-UV8D Plus e KG-UV8E. 19200 baud, pacchetti
// «inizio comando FF lunghezza dati somma»; KG-UV8D Plus e KG-UV8E cifrano dati
// e somma con uno XOR a catena (primo byte con 0x57, poi ognuno con il
// precedente). Memoria di 32 KiB: 999 canali (la posizione 0 non si usa) con
// nomi di 8 caratteri e un byte «valido» (158) per canale.

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type wx8Fam struct{}

type wx8Opt struct {
	start    byte   // primo byte dei pacchetti
	enc      bool   // dati cifrati
	model    string // risposta all'identificazione
	tBase    int64  // toni (KenwoodToneModel): indicatore DCS, polarità, valore iniziale
	tPol     int64
	tInit    int64
	validOff int    // byte «valido» dei canali
	nameRec  int    // dimensione del record del nome
	scrReset bool   // scrambler e compander azzerati a ogni scrittura (KG-UV8D Plus, KG-UV8E)
	format   string // per la verifica della mappa
	ident    func(img []byte) bool
}

const wx8MemFmt = `
#seekto 0x0900;
struct {
    u32     rxfreq;
    u32     txfreq;
    u16     rxtone;
    u16     txtone;
    u8      %s
            power:1,
            unknown2:1;
    u8      unknown3:2,
            scan_add:1,
            %s
            mute_mode:2,
            iswide:1;
    u16     padding;
} memory[1000];
#seekto 0x4780;
struct {
    u8    name[8];
    %s
} names[1000];
#seekto 0x%04x;
u8          valid[1000];
`

const wx8Valid = 158

func wx8Lower(img []byte) []byte {
	if len(img) < 0x430 {
		return nil
	}
	b := bytes.ReplaceAll(bytes.ReplaceAll(img[0x426:0x430], []byte("("), nil), []byte(")"), nil)
	return append([]byte("kg"), bytes.ToLower(b)...)
}

func init() {
	notes := "Accendi la radio con il cavo collegato (19200 baud). 999 canali con nomi di 8 caratteri. Driver di CHIRP (sperimentale)."
	pw := []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}}
	mk := func(id, model, rclass string, o *wx8Opt) *genModel {
		return &genModel{ID: id, Vendor: "Wouxun", Model: model, RClass: rclass, Name: "Wouxun " + model, Slots: 999, First: 1,
			NameMax: 8, Charset: charsetASCII, Powers: pw, Bands: []band{{134000000, 175000000}, {400000000, 520000000}},
			Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 19200, Fam: wx8Fam{}, Opt: o, Notes: notes}
	}
	genRegister(
		mk("wx-kguv8d", "KG-UV8D", "KGUV8DRadio", &wx8Opt{start: 0x7D, model: "KG-UV8D", tBase: 0x2800, tPol: 0x8000, tInit: 0xFFFF,
			validOff: 0x6700, nameRec: 8, format: fmt.Sprintf(wx8MemFmt, "unknown1:6,", "unknown4:2,", "", 0x6700),
			ident: func(img []byte) bool { return len(img) >= 0x408 && bytes.Contains(img[0x400:0x408], []byte("KGUV8D")) }}),
		mk("wx-kguv8dplus", "KG-UV8D Plus", "KGUV8DPlusRadio", &wx8Opt{start: 0x7A, enc: true, model: "KG-UV8D", tBase: 0x4000, tPol: 0x2000,
			validOff: 0x7670, nameRec: 12, scrReset: true, format: fmt.Sprintf(wx8MemFmt, "scrambler:4,\n            unknown1:2,", "unknown4:1,\n            compander:1,", "u8    unknown[4];", 0x7670),
			ident: func(img []byte) bool { return bytes.Contains(wx8Lower(img), []byte("kguv8dplus")) }}),
		mk("wx-kguv8e", "KG-UV8E", "KGUV8ERadio", &wx8Opt{start: 0x7B, enc: true, model: "KG-UV8D-A", tBase: 0x2800, tPol: 0x8000,
			validOff: 0x7670, nameRec: 12, scrReset: true, format: fmt.Sprintf(wx8MemFmt, "scrambler:4,\n            unknown1:2,", "unknown4:1,\n            compander:1,", "u8    unknown[4];", 0x7670),
			ident: func(img []byte) bool { return bytes.Contains(wx8Lower(img), []byte("kguv8e")) }}),
	)
}

func wx8o(m *genModel) *wx8Opt { return m.Opt.(*wx8Opt) }

func (wx8Fam) mem(m *genModel) *bwEl      { return bwMust(wx8o(m).format) }
func (f wx8Fam) layout(m *genModel) *bwEl { return f.mem(m) }
func (f wx8Fam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	b := f.mem(m)
	return b.F("memory").I(i + 1), b.F("names").I(i + 1)
}
func (f wx8Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(m, i)
	return append(append(append([]byte{}, r.Raw(img)...), n.Raw(img)...), img[wx8o(m).validOff+i+1])
}
func (f wx8Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(m, i)
	nl := wx8o(m).nameRec
	r.SetRaw(img, raw[:16])
	if len(raw) >= 16+nl+1 {
		n.SetRaw(img, raw[16:16+nl])
		img[wx8o(m).validOff+i+1] = raw[16+nl]
	}
}
func (f wx8Fam) Match(m *genModel, img []byte) bool { return len(img) == 0x8000 && wx8o(m).ident(img) }
func (wx8Fam) crossband() bool                      { return false }

// wx8ToneEnc: _set_tone_val di KenwoodToneModel (CTCSS con l'indicatore 0x8000)
func wx8ToneEnc(t genTone, base, pol, init int64) int64 {
	switch t.Mode {
	case "DTCS":
		v, _ := strconv.ParseInt(fmt.Sprintf("%d", t.Code), 8, 64)
		v |= base
		if t.Inv {
			v += pol
		}
		return v
	case "Tone":
		return int64(t.Hz10) + 0x8000
	}
	return init
}

// wx8ToneDec: _get_tone_val di KenwoodToneModel
func wx8ToneDec(v, base, pol int64) string {
	if v == 0xFFFF || v == 0 {
		return ""
	}
	if v&base == base {
		code, _ := strconv.Atoi(fmt.Sprintf("%03o", v&0x07FF))
		return genTone{Mode: "DTCS", Code: code, Inv: v&pol != 0}.String()
	}
	return fmtTone(int(v & 0x7FFF))
}

func (f wx8Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := wx8o(m)
	if int(img[o.validOff+i+1]) != wx8Valid {
		return nil
	}
	r, n := f.recs(m, i)
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if tx := r.F("txfreq").Int(img); tx == 0xFFFFFFFF {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = tx * 10
	}
	var nm []rune
	for _, b := range n.F("name").Raw(img) {
		if b != 0 {
			nm = append(nm, rune(b))
		}
	}
	c.Name = strings.TrimRight(string(nm), " \t\n\r\x0b\x0c")
	c.TxTone = wx8ToneDec(r.F("txtone").Int(img), o.tBase, o.tPol)
	c.RxTone = wx8ToneDec(r.F("rxtone").Int(img), o.tBase, o.tPol)
	c.Scan = r.F("scan_add").Int(img) > 0
	if r.F("power").Int(img) > 0 {
		c.Power = 0
	} else {
		c.Power = 2
	}
	c.Narrow = r.F("iswide").Int(img) == 0
	return c
}

func (f wx8Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := wx8o(m)
	r, n := f.recs(m, i)
	if c == nil {
		r.Fill(img, 0)
		img[o.validOff+i+1] = 0
		n.Fill(img, 0)
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
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").SetInt(img, 0xFFFFFFFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("scan_add").SetBool(img, c.Scan)
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("rxtone").SetInt(img, wx8ToneEnc(rt, o.tBase, o.tPol, o.tInit))
	r.F("txtone").SetInt(img, wx8ToneEnc(tt, o.tBase, o.tPol, o.tInit))
	if o.scrReset {
		r.F("scrambler").SetInt(img, 0)
		r.F("compander").SetInt(img, 0)
	}
	r.F("power").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("mute_mode").SetInt(img, 0)
	nm := n.F("name")
	for k := 0; k < 8; k++ {
		v := byte(0)
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	img[o.validOff+i+1] = wx8Valid
	return nil
}

// ---------------------------------------------------------------- protocollo

const (
	wx8CmdID  = 128
	wx8CmdEnd = 129
	wx8CmdRd  = 130
	wx8CmdWr  = 131
)

func wx8Enc(d []byte) []byte {
	out := make([]byte, len(d))
	for k := range d {
		if k == 0 {
			out[k] = 0x57 ^ d[k]
		} else {
			out[k] = out[k-1] ^ d[k]
		}
	}
	return out
}

func wx8Dec(d []byte) []byte {
	out := make([]byte, len(d))
	for k := len(d) - 1; k > 0; k-- {
		out[k] = d[k] ^ d[k-1]
	}
	if len(d) > 0 {
		out[0] = d[0] ^ 0x57
	}
	return out
}

func sum8(b []byte) byte {
	s := 0
	for _, v := range b {
		s += int(v)
	}
	return byte(s)
}

func (s *rawSess) wx8Write(o *wx8Opt, cmd byte, payload []byte) error {
	h := []byte{o.start, cmd, 0xFF, byte(len(payload))}
	cs := sum8(append(append([]byte{}, h[1:]...), payload...))
	body := append(append([]byte{}, payload...), cs)
	if o.enc {
		body = wx8Enc(body)
	}
	return s.send(append(h, body...))
}

// wx8Read: un pacchetto della radio; ok = somma giusta
func (s *rawSess) wx8Read(o *wx8Opt) (data []byte, ok bool, err error) {
	h := readUpTo(s.ctx, s.p, 4, 2*time.Second)
	if s.ctx.Err() != nil {
		return nil, false, errCancelled
	}
	if len(h) == 0 {
		return nil, false, rErr("la radio non risponde")
	}
	if len(h) != 4 {
		return nil, false, rErr("intestazione della risposta incompleta")
	}
	pk := readUpTo(s.ctx, s.p, int(h[3]), 2*time.Second)
	if len(pk) != int(h[3]) {
		return nil, false, rErr("risposta della radio incompleta")
	}
	var last byte
	if o.enc && len(pk) > 0 {
		last = pk[len(pk)-1]
		pk = wx8Dec(pk)
	}
	cs := readUpTo(s.ctx, s.p, 1, 2*time.Second)
	if len(cs) != 1 {
		return nil, false, rErr("risposta della radio incompleta")
	}
	rcs := cs[0]
	if o.enc {
		rcs ^= last
	}
	want := byte((int(sum8(h[1:])) + int(sum8(pk))) % 256)
	return pk, rcs == want, nil
}

func (s *rawSess) wx8Ident(m *genModel) error { return s.wx8IdentO(wx8o(m)) }

func (s *rawSess) wx8IdentO(o *wx8Opt) error {
	for k := 0; k < 10; k++ {
		if err := s.wx8Write(o, wx8CmdID, nil); err != nil {
			return err
		}
		resp, ok, err := s.wx8Read(o)
		if err != nil {
			return err
		}
		if !ok {
			if err := s.sleep(100 * time.Millisecond); err != nil {
				return err
			}
			continue
		}
		if len(resp) >= len(o.model) && string(resp[:len(o.model)]) == o.model {
			return nil
		}
		return rErr("la radio risponde ma non è di questo modello (%q)", strings.TrimRight(string(resp[:min(len(resp), 10)]), "\x00"))
	}
	return rErr("identificazione non riuscita (somma di controllo errata)")
}

func (f wx8Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return f.dl(ctx, m, wx8o(m), port, prog)
}

// dl: lettura dei 32 KiB (anche per la famiglia KG-935G)
func (f wx8Fam) dl(ctx context.Context, m *genModel, o *wx8Opt, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.wx8IdentO(o); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x8000; a += 64 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0x8000, "Lettura…")
		if err := s.wx8Write(o, wx8CmdRd, []byte{byte(a >> 8), byte(a), 64}); err != nil {
			return nil, "", err
		}
		resp, ok, err := s.wx8Read(o)
		if err != nil {
			return nil, "", err
		}
		if !ok || len(resp) < 2+64 {
			return nil, "", rErr("errore di trasmissione leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[2:2+64]...)
	}
	_ = s.wx8Write(o, wx8CmdEnd, nil)
	prog(0x8000, 0x8000, "Lettura completata")
	return data, "", nil
}

func (f wx8Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := wx8o(m)
	if len(img) < 0x8000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x8000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x8000, "Collegamento con la radio…")
	if err := s.wx8Ident(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x8000; a += 64 {
		prog(a, 0x8000, "Scrittura…")
		if err := s.wx8Write(o, wx8CmdWr, append([]byte{byte(a >> 8), byte(a)}, img[a:a+64]...)); err != nil {
			return "", err
		}
		resp, ok, err := s.wx8Read(o)
		if err != nil {
			return "", err
		}
		if !ok || len(resp) < 2 || int(resp[0])<<8|int(resp[1]) != a {
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	_ = s.wx8Write(o, wx8CmdEnd, nil)
	prog(0x8000, 0x8000, "Scrittura completata")
	return "", nil
}
