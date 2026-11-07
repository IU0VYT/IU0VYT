// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: ksun_m6.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: KSUN M6 (ksun_m6.py di CHIRP): 80 canali UHF, frequenze in unità di
// 500 Hz (20 bit ricezione + 20 bit trasmissione), blocchi con somma di controllo

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"time"
)

const ksunM6Format = `
struct {
  u8 _unk1;
  u8 voice: 2,
     beep: 1,
     _unk2: 1,
     vox: 4;
  u8 led_timeout: 4,
     led: 2,
     _unk3: 2;
  u8 _unk4: 4,
     sq: 4;
  u8 _unk5: 2,
     tot: 6;
  u8 _unk6: 4,
     lock_timeout: 4;
  u8 _unk7: 1,
     channel: 7;
  u8 _unk8: 5,
     bat_save: 3;
  u8 _unk9[2];
  u8 pass[6];
} settings;

struct {
  u8 freq[5];
  u16 low_pwr: 1,
      no_tx: 1,
      rx_tone: 2,
      rx_code: 12;
  u16 nfm: 1,
      skip: 1,
      tx_tone: 2,
      tx_code: 12;
  u8 compander: 2,
     hopping: 2,
     scrambler: 4;
} memory[80];
`

type ksunFam struct{}

const ksunStart, ksunSize = 0x0050, 0x10 + 10*80

func init() {
	genRegister(&genModel{ID: "ksun-m6", Vendor: "KSUN", Model: "M6", RClass: "KSunM6Radio", Name: "KSUN M6",
		Slots: 80, First: 1, NameMax: 0, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands: []band{{400000000, 480000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 4800, Fam: ksunFam{},
		Notes: "Accendi la radio con il cavo collegato. 80 canali UHF senza nomi; frequenze a passi di 500 Hz. " +
			"Salto di frequenza, compander e scrambler di ogni canale si conservano. Driver di CHIRP."})
}

func (ksunFam) mem() *bwEl                 { return bwMust(ksunM6Format) }
func (f ksunFam) layout(m *genModel) *bwEl { return f.mem() }
func (f ksunFam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f ksunFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f ksunFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f ksunFam) Match(m *genModel, img []byte) bool                { return false }

var ksunTones = []string{"Tone", "DTCS_N", "DTCS_I", ""}

func ksunToneDec(mode, code int64) string {
	switch mode {
	case 0:
		return fmtTone(int(code))
	case 1, 2:
		n, _ := strconv.Atoi(strconv.FormatInt(code, 8))
		return genTone{Mode: "DTCS", Code: n, Inv: mode == 2}.String()
	}
	return ""
}

func ksunToneEnc(s string) (int64, int64, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, 0, err
	}
	switch t.Mode {
	case "Tone":
		return 0, int64(t.Hz10), nil
	case "DTCS":
		v, _ := strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		if t.Inv {
			return 2, v, nil
		}
		return 1, v, nil
	}
	return 3, 0, nil
}

func (f ksunFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	fr := r.F("freq").Raw(img)
	if bytes.Equal(fr, bytes.Repeat([]byte{0xFF}, 5)) {
		return nil
	}
	rx := int64(fr[0])<<12 | int64(fr[1])<<4 | int64(fr[2])>>4
	tx := int64(fr[2]&0x0F)<<16 | int64(fr[3])<<8 | int64(fr[4])
	c := &Chan{Rx: rx * 1000000 / 2000, Tx: tx * 1000000 / 2000}
	if r.F("no_tx").Bool(img) {
		c.TxOff, c.Tx = true, c.Rx
	}
	c.TxTone = ksunToneDec(r.F("tx_tone").Int(img), r.F("tx_code").Int(img))
	c.RxTone = ksunToneDec(r.F("rx_tone").Int(img), r.F("rx_code").Int(img))
	c.Narrow = r.F("nfm").Bool(img)
	c.Power = m.powerSem(int(r.F("low_pwr").Int(img)))
	c.Scan = !r.F("skip").Bool(img)
	return c
}

func ksunUnits(hz int64) int64 { return int64(math.RoundToEven(float64(hz) / 1000000 * 2000)) }

func (f ksunFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	tm, tc, err := ksunToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rm, rc, err := ksunToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	wasEmpty := f.Get(m, img, i) == nil
	var hop, comp, scr int64
	if !wasEmpty {
		hop, comp, scr = r.F("hopping").Int(img), r.F("compander").Int(img), r.F("scrambler").Int(img)
	}
	rx, tx := ksunUnits(c.Rx), ksunUnits(c.Tx)
	if c.TxOff {
		tx = rx
	}
	if rx > 0xFFFFF || tx > 0xFFFFF {
		return fmt.Errorf("frequenza non valida per questa radio")
	}
	r.F("freq").SetRaw(img, []byte{byte(rx >> 12), byte(rx >> 4), byte(rx<<4) | byte(tx>>16)&0x0F, byte(tx >> 8), byte(tx)})
	r.F("tx_tone").SetInt(img, tm)
	r.F("tx_code").SetInt(img, tc)
	r.F("rx_tone").SetInt(img, rm)
	r.F("rx_code").SetInt(img, rc)
	r.F("no_tx").SetBool(img, c.TxOff)
	r.F("nfm").SetBool(img, c.Narrow)
	r.F("skip").SetBool(img, !c.Scan)
	r.F("low_pwr").SetBool(img, m.powerIndex(c.Power) == 1)
	// come CHIRP: salto di frequenza e compander sì/no, scrambler 0-8
	r.F("hopping").SetBool(img, hop != 0)
	r.F("compander").SetBool(img, comp != 0)
	if scr > 8 {
		scr = 0
	}
	r.F("scrambler").SetInt(img, scr)
	return nil
}

// chirpInvalid (solo per i test): CHIRP ammette solo «split», non «+» e «-»
func (f ksunFam) chirpInvalid(m *genModel, c *Chan) bool {
	d := c.Tx - c.Rx
	return !c.TxOff && d != 0 && d <= 70000000 && d >= -70000000
}

func ksunSum(b []byte) byte {
	cs := 2
	for _, x := range b {
		cs += int(x)
	}
	return byte(cs)
}

func ksunCmd(b ...byte) []byte { return append(b, ksunSum(b)) }

func (s *rawSess) ksunRead(a int) ([]byte, error) {
	cmd := []byte{'R', byte(a >> 8), byte(a)}
	if err := s.send(ksunCmd(cmd...)); err != nil {
		return nil, err
	}
	res := readUpTo(s.ctx, s.p, 3+0x10+1, time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(res) != 0x14 || !bytes.Equal(res[:3], cmd) {
		return nil, rErr("risposta inattesa leggendo il blocco 0x%04X", a)
	}
	if res[0x13] != ksunSum(res[:0x13]) {
		return nil, rErr("somma di controllo errata nel blocco 0x%04X", a)
	}
	return res[3:0x13], nil
}

// ksunVerify: come CHIRP, fino a 3 letture del primo blocco
func (s *rawSess) ksunVerify() error {
	var err error
	for k := 0; k < 3; k++ {
		if _, err = s.ksunRead(ksunStart); err == nil || err == errCancelled {
			return err
		}
	}
	return rErr("la radio non risponde (%v)", err)
}

func (f ksunFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, ksunSize, "Collegamento con la radio…")
	if err := s.ksunVerify(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < ksunSize; a += 0x10 {
		prog(a, ksunSize, "Lettura…")
		b, err := s.ksunRead(ksunStart + a)
		if err != nil {
			return nil, "", err
		}
		data = append(data, b...)
	}
	prog(ksunSize, ksunSize, "Lettura completata")
	return data, "", nil
}

func (f ksunFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < ksunSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), ksunSize)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, ksunSize, "Collegamento con la radio…")
	if err := s.ksunVerify(); err != nil {
		return "", genFail("Identificazione", err)
	}
	if err := s.send(ksunCmd(0x32, 0x31, 0x05, 0x10)); err != nil {
		return "", err
	}
	if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return "", rErr("la radio non è entrata in modalità di programmazione")
	}
	for a := 0; a < ksunSize; a += 0x10 {
		prog(a, ksunSize, "Scrittura…")
		ad := ksunStart + a
		if err := s.send(ksunCmd(append([]byte{'W', byte(ad >> 8), byte(ad)}, img[a:a+0x10]...)...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", ad)
		}
	}
	// la radio si riavvia senza rispondere
	_ = s.send(ksunCmd(0x32, 0x31, 0x05, 0xEE))
	prog(ksunSize, ksunSize, "Scrittura completata")
	return "", nil
}

// fit: frequenze arrotondate al passo di 500 Hz della radio
func (f ksunFam) fit(m *genModel, c *Chan) (bool, string) {
	rx, tx := ksunUnits(c.Rx)*500, ksunUnits(c.Tx)*500
	if rx == c.Rx && tx == c.Tx {
		return true, ""
	}
	why := fmt.Sprintf("frequenze arrotondate al passo di 500 Hz della radio (%s / %s MHz)", fmtMHzShort(rx), fmtMHzShort(tx))
	c.Rx, c.Tx = rx, tx
	return true, why
}
