// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: retevis_rt1.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Retevis RT1 (retevis_rt1.py di CHIRP): versioni VHF e UHF riconosciute
// dall'identificativo a 0x0170

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const rt1Format = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rxtone[2];
  lbcd txtone[2];
  u8 bcl:1,         // Busy Lock
     epilogue:1,    // Epilogue (STE)
     scramble:1,    // Scramble
     compander:1,   // Compander
     skip:1,        // Scan Add
     wide:1,        // Bandwidth
     unknown1:1,
     highpower:1;   // Power Level
  u8 unknown2[3];
} memory[16];

#seekto 0x0120;
struct {
  u8 hivoltnotx:1,  // TX Inhibit when voltage too high
     lovoltnotx:1,  // TX Inhibit when voltage too low
     unknown1:1,
     alarm:1,       // Incept Alarm
     scan:1,        // Scan
     tone:1,        // Tone
     voice:2;       // Voice
  u8 unknown2:1,
     ssave:3,       // Super Battery Save
     unknown3:1,
     save:3;        // Battery Save
  u8 squelch;       // Squelch
  u8 tot;           // Time Out Timer
  u8 voxi:1,        // VOX Inhibit on Receive
     voxd:2,        // VOX Delay
     voxc:1,        // VOX Control
     voxg:4;        // VOX Gain
  u8 unknown4:4,
     scanspeed:4;   // Scan Speed
  u8 unknown5:3,
     scandelay:5;   // Scan Delay
  u8 unknown6:3,
     prioritych:5;  // Priority Channel
  u8 k1shortp;      // Key 1 Short Press
  u8 k2shortp;      // Key 2 Short Press
  u8 k1longp;       // Key 1 Long Press
  u8 k2longp;       // Key 2 Long Press
  u8 lpt;           // Long Press Time
  u8 unknown7:6,
     txtone:2;      // TX Tone
} settings;

#seekto 0x0170;
struct {
  char fp[8];
} fingerprint;
`

type rt1Fam struct{}

var (
	rt1VHF = []byte("PXT8K\xF0\x00\x00")
	rt1UHF = []byte("PXT8K\xF3\x00\x00")
)

func init() {
	genRegister(&genModel{ID: "rt-rt1", Vendor: "Retevis", Model: "RT1", RClass: "RT1Radio", Name: "Retevis RT1",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 5 W", 5, 2}, {"Alta 9 W", 9, 0}},
		Bands: []band{{134000000, 175000000}, {400000000, 521000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 2400, Fam: rt1Fam{},
		Notes: "Accendi la radio con il cavo collegato. 16 canali senza nomi; la radio è VHF oppure UHF (si riconosce dalla memoria) e la scrittura " +
			"si ferma se la memoria è dell'altra versione. Blocco del canale occupato, eliminazione della coda, compander e scrambler si conservano. Driver di CHIRP."})
}

func (rt1Fam) mem() *bwEl                 { return bwMust(rt1Format) }
func (f rt1Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt1Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f rt1Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f rt1Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f rt1Fam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x400 && bytes.HasPrefix(img[0x170:0x176], []byte("PXT8K"))
}

func (f rt1Fam) Bands(m *genModel, img []byte) []band {
	if len(img) >= 0x178 {
		switch {
		case bytes.Equal(img[0x170:0x178], rt1VHF):
			return m.Bands[:1]
		case bytes.Equal(img[0x170:0x178], rt1UHF):
			return m.Bands[1:]
		}
	}
	return m.Bands
}

func (f rt1Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	switch {
	case r.F("rxfreq").Int(img) == r.F("txfreq").Int(img):
	case bytes.Equal(r.F("txfreq").Raw(img), ff):
		c.TxOff = true
	default:
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = !r.F("wide").Bool(img)
	dec := func(e *bwEl) string {
		if bytes.Equal(e.Raw(img), []byte{0xFF, 0xFF}) {
			return ""
		}
		return x3ToneDec(e.Int(img))
	}
	c.TxTone, c.RxTone = dec(r.F("txtone")), dec(r.F("rxtone"))
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.Scan = !r.F("skip").Bool(img)
	return c
}

func (f rt1Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if _, err := genToneParse(c.TxTone); err != nil {
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

func (s *rawSess) rt1Enter() error {
	fail := func(f string, a ...interface{}) error { _ = s.send([]byte("E")); return rErr(f, a...) }
	for _, mg := range []string{"PROGRAMa", "PROGRAMb"} {
		if err := s.send([]byte(mg)); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, 2*time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return fail("la radio non risponde")
		}
		if b[0] != 0x06 {
			return fail("la radio non è entrata in modalità di programmazione")
		}
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 16, 2*time.Second); !bytes.HasPrefix(id, []byte("PXT8K")) {
		return fail("identificativo della radio sconosciuto (% X)", id)
	}
	if err := s.send([]byte("MXT8KCUMHS1X7BN/")); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 || b[0] != 0xB2 {
		return fail("la radio non è entrata in modalità di programmazione")
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 || b[0] != 0x06 {
		return fail("la radio non è entrata in modalità di programmazione")
	}
	return nil
}

func (f rt1Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x400, "Collegamento con la radio…")
	if err := s.rt1Enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, 0x400, 0x10, 2*time.Second, prog, true)
	_ = s.send([]byte("E"))
	return data, "", err
}

func (f rt1Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x400 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x400)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x190, "Collegamento con la radio…")
	if err := s.rt1Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	id, err := wxDownload(ctx, s, 0x170, 0x180, 0x10, 2*time.Second, func(int, int, string) {}, true)
	if err != nil {
		_ = s.send([]byte("E"))
		return "", err
	}
	if !bytes.Equal(id[:8], img[0x170:0x178]) {
		_ = s.send([]byte("E"))
		return "", rErr("la memoria è di una versione della radio diversa (VHF/UHF) da quella collegata: nessun dato è stato scritto")
	}
	err = wxUpload(ctx, s, img, 0, 0x190, 0x10, 2*time.Second, prog)
	_ = s.send([]byte("E"))
	return "", err
}
