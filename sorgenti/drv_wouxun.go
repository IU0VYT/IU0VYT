// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: wouxun_common.py, th_uv3r.py e th_uv3r25.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: protocollo «Wouxun» (wouxun_common.py di CHIRP: blocchi 'R'/'W' con
// indirizzo a 16 bit) e TYT TH-UV3R / TH-UV3R-25 (th_uv3r.py, th_uv3r25.py).

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------- protocollo Wouxun

// wxDownload: 'R' indirizzo (16 bit) dimensione → 'W' indirizzo dimensione dati;
// conferma 06 e risposta di un byte (do_download di wouxun_common.py)
// strict: risposta con l'intestazione "W" indirizzo dimensione e conferma 06
// obbligatoria (driver kyd.py e kyd_IP620.py)
func wxDownload(ctx context.Context, s *rawSess, start, end, bs int, to time.Duration, prog Progress, strict ...bool) ([]byte, error) {
	st := len(strict) > 0 && strict[0]
	var data []byte
	for a := start; a < end; a += bs {
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		prog(a-start, end-start, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), byte(bs)}); err != nil {
			return nil, err
		}
		resp := readUpTo(ctx, s.p, 4+bs, to)
		if len(resp) != 4+bs {
			if ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, rErr("lettura del blocco 0x%04X non riuscita (%d byte invece di %d)", a, len(resp), 4+bs)
		}
		if st && (resp[0] != 'W' || resp[1] != byte(a>>8) || resp[2] != byte(a) || resp[3] != byte(bs)) {
			return nil, rErr("risposta inattesa leggendo il blocco 0x%04X", a)
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, err
		}
		if b := readUpTo(ctx, s.p, 1, to); st && (len(b) != 1 || b[0] != 0x06) {
			if ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	prog(end-start, end-start, "Lettura completata")
	return data, nil
}

// wxUpload: 'W' indirizzo dimensione dati → 06 (do_upload di wouxun_common.py)
func wxUpload(ctx context.Context, s *rawSess, img []byte, start, end, bs int, to time.Duration, prog Progress) error {
	for a := start; a < end; a += bs {
		prog(a-start, end-start, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), byte(bs)}, img[a:a+bs]...)); err != nil {
			return err
		}
		if b := readUpTo(ctx, s.p, 1, to); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return errCancelled
			}
			return rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(end-start, end-start, "Scrittura completata")
	return nil
}

// ---------------------------------------------------------------- TYT TH-UV3R

const thuv3rFormat = `
struct memory {
  ul24 duplex:2,
       bit:1,
       iswide:1,
       bits:2,
       is625:1,
       freq:17;
  ul16 offset;
  ul16 rx_tone;
  ul16 tx_tone;
  u8 unknown;
  u8 name[6];
};

#seekto 0x0010;
struct memory memory[128];

#seekto 0x0870;
u8 emptyflags[16];
u8 skipflags[16];
`

const thuv3r25Format = `
struct memory {
  ul32 rx_freq;
  ul32 tx_freq;
  ul16 rx_tone;
  ul16 tx_tone;
  u8 unknown1a:1,
     iswide:1,
     bclo_n:1,
     vox_n:1,
     tail:1,
     power_high:1,
     voice_mode:2;
  u8 name[6];
  u8 unknown2;
};

#seekto 0x0010;
struct memory memory[128];

#seekto 0x0A80;
u8 emptyflags[16];
u8 skipflags[16];
`

const thuv3rCharset = "0123456789 -*+ABCDEFGHIJKLMNOPQRSTUVWXYZ_/"

type thuv3rFam struct{}

type thuv3rOpt struct {
	v25     bool // TH-UV3R-25 (record di 20 byte, potenza)
	memsize int
}

func init() {
	common := "Accendi la radio con il cavo collegato. 128 canali con nomi di 6 caratteri (cifre, lettere maiuscole e - * + _ /). " +
		"In DCS lo stesso codice vale in trasmissione e in ricezione. Driver di CHIRP."
	genRegister(&genModel{ID: "tyt-thuv3r", Vendor: "TYT", Model: "TH-UV3R", RClass: "TYTUV3RRadio", Name: "TYT TH-UV3R",
		Slots: 128, First: 1, NameMax: 6, Charset: thuv3rCharset, Powers: []genPower{{"Unica", 2, 0}},
		Bands: []band{{136000000, 520000000}}, Narrow: true, HasScan: true, NoTxOff: true, OneDCS: true,
		DCS: dtcsCodes, Baud: 2400, Fam: thuv3rFam{}, Opt: &thuv3rOpt{memsize: 0x910}, Notes: common})
	genRegister(&genModel{ID: "tyt-thuv3r25", Vendor: "TYT", Model: "TH-UV3R-25", RClass: "TYTUV3R25Radio", Name: "TYT TH-UV3R-25 (passo 2,5 kHz)",
		Slots: 128, First: 1, NameMax: 6, Charset: thuv3rCharset, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,8 W", 0.8, 2}},
		Bands: []band{{136000000, 520000000}}, Narrow: true, HasScan: true, NoTxOff: true, OneDCS: true,
		DCS: dtcsCodes, Baud: 2400, Fam: thuv3rFam{}, Opt: &thuv3rOpt{v25: true, memsize: 0xB30},
		Notes: common + " Per canale anche blocco del canale occupato, VOX, eliminazione della coda e compander/scrambler (si conservano)."})
}

func thuv3ro(m *genModel) *thuv3rOpt { return m.Opt.(*thuv3rOpt) }

func (thuv3rFam) mem(m *genModel) *bwEl {
	if thuv3ro(m).v25 {
		return bwMust(thuv3r25Format)
	}
	return bwMust(thuv3rFormat)
}
func (f thuv3rFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f thuv3rFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f thuv3rFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f thuv3rFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f thuv3rFam) Match(m *genModel, img []byte) bool                { return len(img) == thuv3ro(m).memsize }

func (f thuv3rFam) flag(m *genModel, img []byte, name string, i int) bool {
	return img[f.mem(m).F(name).Off+i/8]&(1<<(i%8)) != 0
}

func (f thuv3rFam) setFlag(m *genModel, img []byte, name string, i int, v bool) {
	a := f.mem(m).F(name).Off + i/8
	if v {
		img[a] |= 1 << (i % 8)
	} else {
		img[a] &^= 1 << (i % 8)
	}
}

// valore di un tono come lo scrive la radio: cifre decimali lette come
// esadecimali (88,5 Hz = 0x0885), DCS con 0x8000 (0xC000 invertito)
type thTone struct {
	mode string // "", "Tone", "DTCS"
	inv  bool
	v    int // decimi di Hz o codice DCS
}

func thuv3rToneDec(v int64) (thTone, bool) {
	if v == 0xFFFF {
		return thTone{}, true
	}
	if v&0x8000 != 0 {
		n, err := strconv.Atoi(fmt.Sprintf("%x", v&0x0FFF))
		return thTone{mode: "DTCS", inv: v&0x4000 != 0, v: n}, err == nil
	}
	n, err := strconv.Atoi(fmt.Sprintf("%x", v))
	return thTone{mode: "Tone", v: n}, err == nil
}

func (t thTone) str(code int, inv bool) string {
	switch t.mode {
	case "Tone":
		return fmtTone(t.v)
	case "DTCS":
		return genTone{Mode: "DTCS", Code: code, Inv: inv}.String()
	}
	return ""
}

// thuv3rTones: toni come li mostra CHIRP (_decode_tone di th_uv3r.py)
func thuv3rTones(txv, rxv int64) (string, string) {
	tx, ok1 := thuv3rToneDec(txv)
	rx, ok2 := thuv3rToneDec(rxv)
	if !ok1 || !ok2 || (tx.mode == "" && rx.mode == "") {
		return "", ""
	}
	if tx.mode == "DTCS" && rx.mode == "DTCS" {
		tx.v = rx.v // CHIRP: un solo codice DCS
	}
	if (rx.mode == "" || rx.mode == tx.mode) && (rx.v == 0 || rx.v == tx.v) {
		switch {
		case rx.mode == "Tone":
			s := fmtTone(tx.v)
			return s, s
		case tx.mode == "DTCS":
			return tx.str(tx.v, tx.inv), tx.str(tx.v, rx.inv)
		default:
			return tx.str(tx.v, false), ""
		}
	}
	return tx.str(tx.v, tx.inv), rx.str(rx.v, rx.inv)
}

func thHex(n int) int64 { v, _ := strconv.ParseInt(strconv.Itoa(n), 16, 64); return v }

// thuv3rToneEnc: _encode_tone di th_uv3r.py con i toni del programma
func thuv3rToneEnc(txs, rxs string) (int64, int64, error) {
	t, err := genToneParse(txs)
	if err != nil {
		return 0, 0, err
	}
	r, err := genToneParse(rxs)
	if err != nil {
		return 0, 0, err
	}
	dcs := func(code int, inv bool) int64 {
		if inv {
			return thHex(code) | 0xC000
		}
		return thHex(code) | 0x8000
	}
	enc := func(x genTone) int64 {
		switch x.Mode {
		case "Tone":
			return thHex(x.Hz10)
		case "DTCS":
			return dcs(x.Code, x.Inv)
		}
		return 0xFFFF
	}
	if t.Mode == "DTCS" && r.Mode == "DTCS" {
		// CHIRP (senza codice DCS di ricezione separato): vale il codice di ricezione
		t.Code = r.Code
	}
	return enc(t), enc(r), nil
}

func (f thuv3rFam) Get(m *genModel, img []byte, i int) *Chan {
	if f.flag(m, img, "emptyflags", i) {
		return nil
	}
	r := f.rec(m, i)
	c := &Chan{}
	if thuv3ro(m).v25 {
		c.Rx = r.F("rx_freq").Int(img) * 10
		c.Tx = r.F("tx_freq").Int(img) * 10
		c.Power = m.powerSem(map[bool]int{true: 0, false: 1}[r.F("power_high").Bool(img)])
	} else {
		mult := int64(5000)
		if r.F("is625").Bool(img) {
			mult = 6250
		}
		c.Rx = r.F("freq").Int(img) * mult
		off := r.F("offset").Int(img) * 5000
		switch r.F("duplex").Int(img) {
		case 1:
			c.Tx = c.Rx + off
		case 2:
			c.Tx = c.Rx - off
		default:
			c.Tx = c.Rx
		}
	}
	c.TxTone, c.RxTone = thuv3rTones(r.F("tx_tone").Int(img), r.F("rx_tone").Int(img))
	c.Narrow = !r.F("iswide").Bool(img)
	c.Scan = !f.flag(m, img, "skipflags", i)
	var name strings.Builder
	for _, b := range r.F("name").Raw(img) {
		if int(b) < len(thuv3rCharset) {
			name.WriteByte(thuv3rCharset[b])
		}
	}
	c.Name = strings.TrimRight(name.String(), " \t\n\r\x0b\x0c")
	return c
}

func (f thuv3rFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := thuv3ro(m)
	if c == nil {
		f.setFlag(m, img, "emptyflags", i, true)
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	wasEmpty := f.flag(m, img, "emptyflags", i)
	txv, rxv, err := thuv3rToneEnc(c.TxTone, c.RxTone)
	if err != nil {
		return err
	}
	if o.v25 {
		if c.Rx/10 > 0xFFFFFFFF || c.Tx/10 > 0xFFFFFFFF {
			return fmt.Errorf("frequenza non valida")
		}
		// voice_mode 0 non è tra le scelte di CHIRP: diventa «None» (3) come quando
		// CHIRP riscrive il canale
		if !wasEmpty && r.F("voice_mode").Int(img) == 0 {
			r.F("voice_mode").SetInt(img, 3)
		}
	} else {
		off := c.Tx - c.Rx
		if off < 0 {
			off = -off
		}
		if off/5000 > 0xFFFF {
			return fmt.Errorf("scostamento troppo grande per questa radio (massimo %s MHz)", fmtMHzShort(0xFFFF*5000))
		}
	}
	f.setFlag(m, img, "emptyflags", i, false)
	if o.v25 {
		r.F("rx_freq").SetInt(img, c.Rx/10)
		r.F("tx_freq").SetInt(img, c.Tx/10)
		r.F("power_high").SetBool(img, m.powerIndex(c.Power) == 0)
	} else {
		mult := int64(5000)
		frac := c.Rx%5000 != 0 && c.Rx%6250 == 0
		if frac {
			mult = 6250
		}
		r.F("is625").SetBool(img, frac)
		r.F("freq").SetInt(img, c.Rx/mult)
		switch {
		case c.Tx == c.Rx:
			// CHIRP riscrive lo scostamento memorizzato (0,6 MHz per un canale nuovo)
			if wasEmpty {
				r.F("offset").SetInt(img, 600000/5000)
			}
			r.F("duplex").SetInt(img, 0)
		case c.Tx > c.Rx:
			r.F("duplex").SetInt(img, 1)
			r.F("offset").SetInt(img, (c.Tx-c.Rx)/5000)
		default:
			r.F("duplex").SetInt(img, 2)
			r.F("offset").SetInt(img, (c.Rx-c.Tx)/5000)
		}
	}
	r.F("iswide").SetBool(img, !c.Narrow)
	r.F("tx_tone").SetInt(img, txv)
	r.F("rx_tone").SetInt(img, rxv)
	f.setFlag(m, img, "skipflags", i, !c.Scan)
	nm := r.F("name")
	for k := 0; k < 6; k++ {
		ch := byte(' ')
		if k < len(c.Name) {
			ch = c.Name[k]
		}
		x := strings.IndexByte(thuv3rCharset, ch)
		if x < 0 {
			x = strings.IndexByte(thuv3rCharset, ' ')
		}
		img[nm.Off+k] = byte(x)
	}
	return nil
}

// thuv3rPrep: "PROGRAMa" → 06 (tyt_uv3r_prep di th_uv3r.py)
func (s *rawSess) thuv3rPrep() error {
	if err := s.send([]byte("PROGRAMa")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, 2*time.Second)
	if len(b) == 0 {
		if s.ctx.Err() != nil {
			return errCancelled
		}
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio non ha confermato il primo comando")
	}
	return nil
}

func (f thuv3rFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	n := thuv3ro(m).memsize
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, n, "Collegamento con la radio…")
	if err := s.thuv3rPrep(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, n, 0x10, 2*time.Second, prog)
	return data, "", err
}

func (f thuv3rFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	n := thuv3ro(m).memsize
	if len(img) < n {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), n)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, n, "Collegamento con la radio…")
	if err := s.thuv3rPrep(); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img, 0, n, 0x10, 2*time.Second, prog)
}

// crossband: CHIRP non ammette «split» (scostamento oltre 70 MHz) per queste
// radio; qui si scrive comunque come scostamento (TH-UV3R) o frequenza di
// trasmissione (TH-UV3R-25), che la memoria della radio permette
func (thuv3rFam) crossband() bool { return true }
