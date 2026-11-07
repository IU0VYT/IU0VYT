// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radioddity_r2.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Radioddity R2, Retevis RT24, RT24V e H777S (radioddity_r2.py di CHIRP)

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const r2FormatTpl = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown1:1,
    compand:1,
    scramb:1,
    scanadd:1,
    power:1,
    mode:1,
    unknown2:1,
    bclo:1;
  u8 reserved[3];
} memory[%d];

#seekto 0x03C0;
struct {
  u8 unknown3c08:1,
      scanmode:1,
      unknown3c06:1,
      unknown3c05:1,
      voice:2,
      save:1,
      beep:1;
  u8 squelch;
  u8 unknown3c2;
  u8 timeout;
  u8 voxgain;
  u8 specialcode;
  u8 unknown3c6;
  u8 voxdelay;
} settings;
`

type r2Fam struct{}

func init() {
	mk := func(id, vendor, model, rclass, name string, slots int, b band, notes string) *genModel {
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: slots, First: 1, NameMax: 0,
			Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 3 W", 3, 0}}, Bands: []band{b}, Narrow: true, HasScan: true,
			DCS: dtcsCodes, Baud: 9600, Fam: r2Fam{},
			Notes: "Accendi la radio con il cavo collegato. Canali senza nomi; blocco del canale occupato, scrambler e compander si conservano. " +
				"Se la radio ha una password la lettura si ferma (come in CHIRP). " + notes + "Driver di CHIRP."}
	}
	uhf := band{400000000, 470000000}
	genRegister(
		mk("rd-r2", "Radioddity", "R2", "RadioddityR2", "Radioddity R2", 16, uhf, ""),
		mk("rt-rt24", "Retevis", "RT24", "RetevisRT24", "Retevis RT24 (PMR446)", 16, uhf,
			"Venduta come radio PMR446: CHIRP permette tutta la gamma 400-470 MHz, in Europa usa solo le frequenze PMR446. "),
		mk("rt-rt24v", "Retevis", "RT24V", "RetevisRT24V", "Retevis RT24V (Freenet)", 6, band{136000000, 174000000},
			"Venduta come radio Freenet (Germania, 149 MHz): CHIRP permette tutta la gamma 136-174 MHz. "),
		mk("rt-h777s", "Retevis", "H777S", "RetevisH777S", "Retevis H777S", 16, uhf, ""),
	)
}

func (r2Fam) mem(m *genModel) *bwEl                { return bwMust(fmt.Sprintf(r2FormatTpl, m.Slots)) }
func (f r2Fam) layout(m *genModel) *bwEl           { return f.mem(m) }
func (f r2Fam) rec(m *genModel, i int) *bwEl       { return f.mem(m).F("memory").I(i) }
func (r2Fam) crossband() bool                      { return true }
func (f r2Fam) Match(m *genModel, img []byte) bool { return false }
func (f r2Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f r2Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f r2Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	rx := r.F("rx_freq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rx_freq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: r.F("tx_freq").Int(img) * 10}
	if c.Tx != rx && (c.Tx == 0 || c.Tx == 1666666650) {
		c.TxOff, c.Tx = true, rx
	}
	c.Narrow = r.F("mode").Bool(img)
	c.Scan = r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	dec := func(e *bwEl) string {
		if bytes.Equal(e.Raw(img), []byte{0xFF, 0xFF}) {
			return ""
		}
		return x3ToneDec(e.Int(img))
	}
	c.TxTone, c.RxTone = dec(r.F("tx_tone")), dec(r.F("rx_tone"))
	return c
}

func (f r2Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		img2 := bytes.Repeat([]byte{0xFF}, 13)
		copy(img[r.Off:], img2)
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
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	if err := x3ToneEnc(r.F("tx_tone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rx_tone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("scanadd").SetBool(img, c.Scan)
	r.F("mode").SetBool(img, c.Narrow)
	return nil
}

// r2Send: un byte alla volta (come CHIRP)
func (s *rawSess) r2Send(b []byte, gap time.Duration) error {
	for k := range b {
		if gap > 0 {
			if err := s.sleep(gap); err != nil {
				return err
			}
		}
		if err := s.send(b[k : k+1]); err != nil {
			return err
		}
	}
	return nil
}

func (s *rawSess) r2Enter() error {
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	ok := false
	var ack []byte
	for k := 0; k < 5 && !ok; k++ {
		if err := s.r2Send([]byte("TYOGRAM"), 0); err != nil {
			return err
		}
		ack = readUpTo(s.ctx, s.p, 1, time.Second)
		ok = len(ack) == 1 && ack[0] == 0x06
	}
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if !ok {
		if len(ack) == 0 {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non è entrata in modalità di programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
	}
	fail := func(f string, a ...interface{}) error { _ = s.send([]byte{0x06}); return rErr(f, a...) }
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, time.Second)
	if len(id) == 0 || (id[0] != 0x06 && id[0] != 0x53) {
		return fail("identificativo della radio sconosciuto (% X)", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return fail("la radio non è entrata in modalità di programmazione")
	}
	if err := s.send([]byte{0x05}); err != nil {
		return err
	}
	if pw := readUpTo(s.ctx, s.p, 6, time.Second); !bytes.Equal(pw, bytes.Repeat([]byte{0xFF}, 6)) {
		return fail("la radio è protetta da una password: toglila dal menu della radio o con il programma del costruttore")
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	// le RT24 recenti non confermano: si prosegue come il programma del costruttore
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) == 1 && b[0] != 0x06 {
		return fail("la radio non è entrata in modalità di programmazione")
	}
	return nil
}

func (f r2Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x3F0, "Collegamento con la radio…")
	if err := s.r2Enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x3F0; a += 8 {
		prog(a, 0x3F0, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), 8}
		if err := s.r2Send(cmd, 5*time.Millisecond); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 12, time.Second)
		if len(resp) != 12 || !bytes.Equal(resp[:4], []byte{'W', cmd[1], cmd[2], cmd[3]}) {
			_ = s.send([]byte{0x06})
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.sleep(5 * time.Millisecond); err != nil {
			return nil, "", err
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			_ = s.send([]byte{0x06})
			return nil, "", rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	_ = s.send([]byte{0x06})
	prog(0x3F0, 0x3F0, "Lettura completata")
	return data, "", nil
}

func (f r2Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x3F0 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x3F0)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x3F0, "Collegamento con la radio…")
	if err := s.r2Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x3F0; a += 8 {
		prog(a, 0x3F0, "Scrittura…")
		if err := s.r2Send(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...), 0); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			_ = s.send([]byte{0x06})
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	_ = s.send([]byte{0x06})
	prog(0x3F0, 0x3F0, "Scrittura completata")
	return "", nil
}
