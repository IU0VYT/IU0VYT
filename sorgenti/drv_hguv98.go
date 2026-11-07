// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: hg_uv98.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Lanchonlh HG-UV98 (hg_uv98.py di CHIRP): 9600 baud con parità pari,
// blocchi da 64 byte con doppia conferma

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const hguv98Format = `
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:6,
     wide:1,
     highpower:1;
  u8 unknown2:5,
     bcl:1,
     scan:1,
     unknown3:1;
  u8 unknown4[2];
} memory[130];

#seekto 0x0a00;
struct {
  u8 abr;  // 0x0a00
  u8 save;
  u8 ch_a_step;
  u8 ch_b_step;
  u8 vox_grd;
  u8 ch_a_sql;
  u8 ch_b_sql;
  u8 roger;
  u8 ch_a_v_m;
  u8 ch_b_v_m;
  u8 ch_a_ch_mdf;
  u8 ch_b_ch_mdf;
  u8 tdr;
  u8 unknown5[3];
  u8 unknown6[5];  // 0x0a10
  u8 english;
  u8 beep;
  u8 voice;
  u8 night_mode;
  u8 abr_lv;  // backlight level
  u8 tot;
  u8 toa;
  u8 vox_dly;
  u8 sc_rev;
  u8 lockmode;
  u8 autolock;
  u8 unknown7;  // 0x0a20
  u8 pf1_short;
  u8 pf1_long;
  u8 pf2_short;
  u8 pf2_long;
  u8 top_short;
  u8 top_long;
  u8 rpt_rct;
  u8 sc_qt;
  u8 pri_ch;
  u8 pri_scn;
  u8 unknown8;
  u8 aprs_rx_band;
  u8 ch_a_mute;
  u8 ch_b_mute;
  u8 unknown9[7];  // 0x0a30
  u8 tx_priority;
  u8 aprs_rx_popup;
  u8 aprs_rx_tone;
  u8 aprs_tx_tone;
  u8 unknown10;
  u8 auto_lock_dly;
  u8 menu_dly;
  u8 beacon_exit_dly;
  u8 unknown11;
  u8 unknown12[2];        // 0x0a40
  u8 ch_a_mem_ch;
  u8 ch_b_mem_ch;
  u8 unknown13[12];
} settings;

#seekto 0x1000;
struct {
  char name[11];
  u8 unknown[5];
} name[128];

struct {
    char callsign[9];
    u8 null;
} aprs;

#seekto 0x1f80;
struct {
  ul32 unknown[8];
} unknown_settings;

`

type hguv98Fam struct{}

func init() {
	genRegister(&genModel{ID: "lch-hguv98", Vendor: "Lanchonlh", Model: "HG-UV98", RClass: "LanchonlhHG_UV98", Name: "Lanchonlh HG-UV98",
		Slots: 128, First: 1, NameMax: 8, Charset: charsetAlnum, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands: []band{{136000000, 174000000}, {400000000, 500000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600,
		Fam: hguv98Fam{}, Notes: "Accendi la radio con il cavo collegato (9600 baud, parità pari). 128 canali con nomi di 8 caratteri; " +
			"il blocco del canale occupato si conserva. Driver di CHIRP (versione «alpha»): controlla le impostazioni anche con il programma del costruttore."})
}

func (hguv98Fam) mem() *bwEl                 { return bwMust(hguv98Format) }
func (f hguv98Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f hguv98Fam) recs(i int) (*bwEl, *bwEl) {
	return f.mem().F("memory").I(i), f.mem().F("name").I(i).F("name")
}
func (f hguv98Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f hguv98Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 27 {
		n.SetRaw(img, raw[16:27])
	}
}
func (f hguv98Fam) Match(m *genModel, img []byte) bool { return false }
func (hguv98Fam) crossband() bool                      { return true }

func (f hguv98Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if bytes.HasPrefix(r.Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10, Tx: r.F("tx_freq").Int(img) * 10}
	c.TxTone = kenwoodToneDec(r.F("txtone").Int(img), 0x2800, 0x8000)
	c.RxTone = kenwoodToneDec(r.F("rxtone").Int(img), 0x2800, 0x8000)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Narrow = r.F("wide").Bool(img) // in CHIRP «wide» = 1 vuol dire NFM
	c.Scan = r.F("scan").Bool(img)
	nm, _, _ := bytes.Cut(n.Raw(img), []byte{0xFF})
	c.Name = strings.TrimRight(string(nm), " \t\n\r\x0b\x0c")
	return c
}

func (f hguv98Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	nm := []byte(c.Name)
	for len(nm) < 11 {
		nm = append(nm, 0xFF)
	}
	n.SetRaw(img, nm[:11])
	// come CHIRP: campi riservati a 1
	for _, k := range []string{"unknown1", "unknown2", "unknown3"} {
		r.F(k).SetInt(img, 0xFF)
	}
	r.F("unknown4").Fill(img, 0xFF)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	r.F("tx_freq").SetInt(img, c.Tx/10)
	r.F("rxtone").SetInt(img, kenwoodToneEnc(rt, 0x2800, 0x8000, 0xFFFF))
	r.F("txtone").SetInt(img, kenwoodToneEnc(tt, 0x2800, 0x8000, 0xFFFF))
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("wide").SetBool(img, c.Narrow)
	r.F("scan").SetBool(img, c.Scan)
	return nil
}

func (s *rawSess) hguv98Ident() error {
	if err := s.send([]byte("NiNHSG0N")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio non è entrata in modalità di programmazione")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, time.Second); !bytes.Equal(id, []byte("P3107\x00\x00\x00")) {
		return rErr("modello non corrispondente (identificativo % X)", id)
	}
	return s.ack2()
}

// ack2: 06 → 06
func (s *rawSess) ack2() error {
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		if s.ctx.Err() != nil {
			return errCancelled
		}
		return rErr("la radio non ha confermato")
	}
	return nil
}

func (f hguv98Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	if err := setEvenParity(s.p); err != nil {
		return nil, "", err
	}
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.hguv98Ident(); err != nil {
		_ = s.send([]byte{0x45})
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x2000; a += 64 {
		prog(a, 0x2000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 64}); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 4+64, time.Second)
		if len(resp) != 68 || int(resp[1])<<8|int(resp[2]) != a {
			_ = s.send([]byte{0x45})
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.ack2(); err != nil {
			return nil, "", err
		}
		data = append(data, resp[4:]...)
		if err := s.ack2(); err != nil {
			return nil, "", err
		}
	}
	_ = s.send([]byte{0x45})
	prog(0x2000, 0x2000, "Lettura completata")
	return data, "", nil
}

func (f hguv98Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2000 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2000)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	if err := setEvenParity(s.p); err != nil {
		return "", err
	}
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.hguv98Ident(); err != nil {
		_ = s.send([]byte{0x45})
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x2000; a += 64 {
		prog(a, 0x2000, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 64}, img[a:a+64]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			_ = s.send([]byte{0x45})
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
		if err := s.ack2(); err != nil {
			return "", err
		}
	}
	_ = s.send([]byte{0x45})
	prog(0x2000, 0x2000, "Scrittura completata")
	return "", nil
}
