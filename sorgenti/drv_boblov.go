// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: boblov_x3plus.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Boblov X3 Plus (radio da casco, boblov_x3plus.py di CHIRP). La radio
// ripete ogni byte che riceve: l'eco si legge e si confronta prima della risposta.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const x3plusFormat = `
#seekto 0x0010;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 unknown1:1,
       compander:1,
       scramble:1,
       skip:1,
       highpower:1,
       narrow:1,
       unknown2:1,
       bcl:1;
    u8 unknown3[3];
} memory[16];
#seekto 0x03C0;
struct {
    u8 unknown1:4,
       voiceprompt:2,
       batterysaver:1,
       beep:1;
    u8 squelchlevel;
    u8 unknown2;
    u8 timeouttimer;
    u8 voxlevel;
    u8 unknown3;
    u8 unknown4;
    u8 voxdelay;
} settings;
`

type x3plusFam struct{}

func init() {
	genRegister(&genModel{ID: "bob-x3plus", Vendor: "Boblov", Model: "X3Plus", RClass: "BoblovX3Plus", Name: "Boblov X3 Plus (radio da casco)",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600, Fam: x3plusFam{},
		Notes: "Radio da casco per moto e bicicletta: 16 canali UHF senza nomi. Come in CHIRP (driver sperimentale) la potenza dei canali si legge " +
			"ma non si cambia; blocco del canale occupato, scrambler e compander si conservano."})
}

func (x3plusFam) mem() *bwEl                    { return bwMust(x3plusFormat) }
func (f x3plusFam) layout(m *genModel) *bwEl    { return f.mem() }
func (f x3plusFam) rec(i int) *bwEl             { return f.mem().F("memory").I(i) }
func (f x3plusFam) powerFixed(m *genModel) bool { return true }
func (f x3plusFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f x3plusFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f x3plusFam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x3F0 && bytes.Contains(img[0x3D0:0x3D8], []byte("P310"))
}

// toni in BCD: FFFF spento, 8000+ DCS, 12000+ DCS invertito
func x3ToneDec(v int64) string {
	switch {
	case v == 16665:
		return ""
	case v >= 12000:
		return genTone{Mode: "DTCS", Code: int(v - 12000), Inv: true}.String()
	case v >= 8000:
		return genTone{Mode: "DTCS", Code: int(v - 8000)}.String()
	}
	return fmtTone(int(v))
}

func x3ToneEnc(e *bwEl, img []byte, s string) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch t.Mode {
	case "Tone":
		e.SetInt(img, int64(t.Hz10))
	case "DTCS":
		e.SetInt(img, int64(t.Code))
		if t.Inv {
			img[e.Off+1] |= 0xC0
		} else {
			img[e.Off+1] |= 0x80
		}
	default:
		e.Fill(img, 0xFF)
	}
	return nil
}

func (f x3plusFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), ff) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = !r.F("skip").Bool(img)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	c.TxTone = x3ToneDec(r.F("txtone").Int(img))
	c.RxTone = x3ToneDec(r.F("rxtone").Int(img))
	return c
}

func (f x3plusFam) Set(m *genModel, img []byte, i int, c *Chan) error {
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
		r.F("txtone").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
		if err := x3ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
			return err
		}
	}
	if err := x3ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("skip").SetBool(img, !c.Scan)
	return nil
}

type x3Sess struct{ *rawSess }

// write: invio e lettura dell'eco (_write di boblov_x3plus.py)
func (s x3Sess) write(b []byte) error {
	if err := s.send(b); err != nil {
		return err
	}
	e := readUpTo(s.ctx, s.p, len(b), 3*time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if !bytes.Equal(e, b) {
		return rErr("la radio non ha ripetuto i dati inviati (eco)")
	}
	return nil
}

func (s x3Sess) enter() error {
	if err := s.write([]byte(".VKOGRAM")); err != nil {
		return err
	}
	if err := s.sleep(500 * time.Millisecond); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, 3*time.Second)
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio non è entrata in modalità di programmazione")
	}
	if err := s.write([]byte{0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, 3*time.Second); !bytes.HasPrefix(id, []byte("SMP558")) {
		return rErr("identificativo della radio sconosciuto (% X)", id)
	}
	if err := s.write([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 3*time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio non ha confermato l'identificazione")
	}
	return nil
}

func (f x3plusFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer rs.p.Close()
	s := x3Sess{rs}
	prog(0, 0x3F0, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0x3F0; a += 8 {
		prog(a, 0x3F0, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), 8}
		if err := s.write(cmd); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 12, 3*time.Second)
		if len(resp) != 12 || !bytes.Equal(resp[:4], []byte{'W', cmd[1], cmd[2], cmd[3]}) {
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.write([]byte{0x06}); err != nil {
			return nil, "", err
		}
		if b := readUpTo(ctx, s.p, 1, 3*time.Second); len(b) != 1 || b[0] != 0x06 {
			return nil, "", rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	_ = s.write([]byte{'b'})
	prog(0x3F0, 0x3F0, "Lettura completata")
	return data, "", nil
}

func (f x3plusFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x3F0 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x3F0)
	}
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer rs.p.Close()
	s := x3Sess{rs}
	prog(0, 0x3F0, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0x3F0; a += 8 {
		prog(a, 0x3F0, "Scrittura…")
		if err := s.write(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, 3*time.Second); len(b) != 1 || b[0] != 0x06 {
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	_ = s.write([]byte{'b'})
	prog(0x3F0, 0x3F0, "Scrittura completata")
	return "", nil
}
