// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: fd268.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Feidaxin FD-268A/B, FD-288A/B, FD-150A, FD-160A, FD-450A, FD-460A e
// FD-460UH (fd268.py di CHIRP): 99 canali, modello riconosciuto dai primi 4 byte

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const fd268Format = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown:4,
     scrambler:1,
     unknown1:1,
     unknown2:1,
     busy_lock:1;
  u8 unknown3[3];
} memory[99];

#seekto 0x0640;
struct {
  lbcd vrx_freq[4];
  lbcd vtx_freq[4];
  lbcd vrx_tone[2];
  lbcd vtx_tone[2];
  u8 shift_plus:1,
     shift_minus:1,
     unknown11:2,
     scramble:1,
     unknown12:1,
     unknown13:1,
     busy_lock:1;
  u8 unknown14[3];
} vfo;

#seekto 0x07B0;
struct {
  u8 ani_mode;
  char ani[3];
  u8 unknown21[12];
  u8 unknown22:5,
     bw1:1,             // twin setting of bw (LCD "romb")
     bs1:1,             // twin setting of bs (LCD "S")
     warning1:1;        // twin setting of warning (LCD "Tune")
  u8 sql;
  u8 monitorval;
  u8 tot;
  u8 powerrank;
  u8 unknown23[3];
  u8 unknown24[8];
  char model[8];
  u8 unknown26[8];
  u8 step;
  u8 unknown27:2,
     power:1,
     lamp:1,
     lamp_auto:1,
     key:1,
     monitor:1,
     bw:1;
  u8 unknown28:3,
     warning:1,
     bs:1,
     unknown29:1,
     wmem:1,
     wvfo:1;
  u8 active_ch;
  u8 unknown30[4];
  u8 unknown31[4];
  bbcd vfo_shift[4];
} settings;
`

type fd268Fam struct{}

type fd268Opt struct{ ident []byte }

func init() {
	for _, v := range []struct {
		model string
		b     band
		id    string
	}{
		{"FD-268A", band{136000000, 174000000}, "\xFF\xEE\x46\xFF"},
		{"FD-268B", band{400000000, 470000000}, "\xFF\xEE\x47\xFF"},
		{"FD-288A", band{136000000, 174000000}, "\xFF\xEE\x4B\xFF"},
		{"FD-288B", band{400000000, 470000000}, "\xFF\xEE\x4C\xFF"},
		{"FD-150A", band{136000000, 174000000}, "\xFF\xEE\x45\xFF"},
		{"FD-160A", band{136000000, 174000000}, "\xFF\xEE\x48\xFF"},
		{"FD-450A", band{400000000, 470000000}, "\xFF\xEE\x44\xFF"},
		{"FD-460A", band{400000000, 470000000}, "\xFF\xEE\x4A\xFF"},
		{"FD-460UH", band{400000000, 480000000}, "\xFF\xF4\x44\xFF"},
	} {
		id := "fdx-" + strings.ToLower(strings.ReplaceAll(v.model[3:], "-", ""))
		genRegister(&genModel{ID: id, Vendor: "Feidaxin", Model: v.model, RClass: strings.ReplaceAll(v.model, "-", "") + "Radio",
			Name: "Feidaxin " + v.model, Slots: 99, First: 1, NameMax: 0, Powers: []genPower{{"Unica", 5, 0}},
			Bands: []band{v.b}, Narrow: false, HasScan: false, DCS: dtcsCodes, Baud: 9600, Fam: fd268Fam{}, Opt: &fd268Opt{ident: []byte(v.id)},
			Notes: "Accendi la radio con il cavo collegato. 99 canali senza nomi, senza potenza e banda per canale (sono impostazioni generali); " +
				"blocco del canale occupato e scrambler si conservano. Driver di CHIRP (sperimentale)."})
	}
}

func (fd268Fam) mem() *bwEl                 { return bwMust(fd268Format) }
func (f fd268Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f fd268Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (fd268Fam) crossband() bool            { return true }
func (f fd268Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f fd268Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f fd268Fam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x800 && bytes.Equal(img[:4], m.Opt.(*fd268Opt).ident)
}

func (f fd268Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if img[r.Off] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10, Scan: true}
	if bytes.Equal(r.F("tx_freq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("tx_freq").Int(img) * 10
	}
	dec := func(e *bwEl) string {
		if bytes.Equal(e.Raw(img), []byte{0xFF, 0xFF}) {
			return ""
		}
		return x3ToneDec(e.Int(img))
	}
	c.TxTone, c.RxTone = dec(r.F("tx_tone")), dec(r.F("rx_tone"))
	return c
}

func (f fd268Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	r.F("rx_freq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("tx_freq").Fill(img, 0xFF)
	} else {
		r.F("tx_freq").SetInt(img, c.Tx/10)
	}
	if err := x3ToneEnc(r.F("tx_tone"), img, c.TxTone); err != nil {
		return err
	}
	return x3ToneEnc(r.F("rx_tone"), img, c.RxTone)
}

func (s *rawSess) fd268Prog() error {
	ok := false
	for k := 0; k < 8 && !ok; k++ {
		if err := s.r2Send([]byte("\x02PROGRA"), 0); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, time.Second)
		ok = len(b) == 1 && b[0] == 0x06
	}
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if !ok {
		return rErr("la radio non è entrata in modalità di programmazione: controlla il cavo e la radio, poi riprova")
	}
	if err := s.send([]byte("M\x02")); err != nil {
		return err
	}
	readUpTo(s.ctx, s.p, 8, time.Second)
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	return nil
}

func (f fd268Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x800, "Collegamento con la radio…")
	if err := s.fd268Prog(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, 0x800, 8, time.Second, prog, true)
	return data, "", err
}

func (f fd268Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x800 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x800)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x800, "Collegamento con la radio…")
	if err := s.fd268Prog(); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img, 0, 0x800, 8, time.Second, prog)
}
