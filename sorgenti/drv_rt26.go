// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_rt26.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis RT26 (retevis_rt26.py di CHIRP): 16 canali UHF senza nome.
// 4800 baud; «PROGRAMa» → 06, «PROGRAMb» → 06, 02 → «PDK80…», chiave di 16
// caratteri → B2, 06 → 06; blocchi 'R'/'W' da 16 byte con conferma; 'E' per uscire.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type rt26Fam struct{}

const rt26Format = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rxtone[2];
  lbcd txtone[2];
  u8 unknown1:1,
     pttid:2,
     skip:1,
     wide:1,
     bcl:1,
     epilogue:1,
     highpower:1;
  u8 unknown2[3];
} memory[16];
`

func init() {
	genRegister(&genModel{ID: "rt-rt26", Vendor: "Retevis", Model: "RT26", RClass: "RT26Radio", Name: "Retevis RT26",
		Slots: 16, First: 1, Powers: []genPower{{"Bassa 5 W", 5, 2}, {"Alta 10 W", 10, 0}},
		Bands: []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 4800, Fam: rt26Fam{},
		Notes: "16 canali UHF senza nome. Copia a 4800 baud. Blocco del canale occupato, eliminazione della coda (STE) e " +
			"PTT ID si conservano. Driver di CHIRP."})
}

func (rt26Fam) mem() *bwEl                 { return bwMust(rt26Format) }
func (f rt26Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt26Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f rt26Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f rt26Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f rt26Fam) Match(m *genModel, img []byte) bool {
	return len(img) >= 0x176 && bytes.HasPrefix(img[0x170:], []byte("PDK80"))
}

func (f rt26Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	rx := r.F("rxfreq").Int(img)
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff) {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	switch tx := r.F("txfreq").Int(img); {
	case tx == rx:
	case bytes.Equal(r.F("txfreq").Raw(img), ff):
		c.TxOff = true
	default:
		c.Tx = tx * 10
	}
	c.Narrow = !r.F("wide").Bool(img)
	c.TxTone = x3ToneDec(r.F("txtone").Int(img))
	c.RxTone = x3ToneDec(r.F("rxtone").Int(img))
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Scan = !r.F("skip").Bool(img)
	return c
}

func (f rt26Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
		return err
	}
	if _, err := genToneParse(c.RxTone); err != nil {
		return err
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("wide").SetBool(img, !c.Narrow)
	if err := x3ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("skip").SetBool(img, !c.Scan)
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rt26Enter() error {
	fail := func(e error) error {
		s.send([]byte{'E'})
		return e
	}
	for _, mg := range []string{"PROGRAMa", "PROGRAMb"} {
		if err := s.send([]byte(mg)); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return fail(rErr("la radio non risponde"))
		}
		if b[0] != 0x06 {
			return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
		}
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 16, time.Second); !bytes.HasPrefix(id, []byte("PDK80")) {
		return fail(rErr("identificativo sconosciuto (% X)", id))
	}
	if err := s.send([]byte("MXT8KCUMHS1X7BN/")); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0xB2 {
		return fail(rErr("la radio ha rifiutato la modalità di programmazione"))
	}
	if err := s.ack2(); err != nil {
		return fail(err)
	}
	return nil
}

func (f rt26Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x400, "Collegamento con la radio…")
	if err := s.rt26Enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x400; a += 0x10 {
		prog(a, 0x400, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x10}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x14, time.Second)
		if len(b) != 0x14 || !bytes.Equal(b[:4], []byte{'W', byte(a >> 8), byte(a), 0x10}) {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			s.send([]byte{'E'})
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.ack2(); err != nil {
			s.send([]byte{'E'})
			return nil, "", err
		}
		data = append(data, b[4:]...)
	}
	s.send([]byte{'E'})
	prog(0x400, 0x400, "Lettura completata")
	return data, "", nil
}

func (f rt26Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x190 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x190, "Collegamento con la radio…")
	if err := s.rt26Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x190; a += 0x10 {
		prog(a, 0x190, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x10}, img[a:a+0x10]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			s.send([]byte{'E'})
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	s.send([]byte{'E'})
	prog(0x190, 0x190, "Scrittura completata")
	return "", nil
}
