// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: tg_uv2p.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Quansheng TG-UV2+ (tg_uv2p.py di CHIRP): 9600 baud con due bit di stop;
// l'immagine inizia con 48 byte di intestazione come i file di CHIRP

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

const tguv2pFormat = `
struct memory {
  bbcd freq[4];
  bbcd offset[4];
  u8 rxtone;
  u8 txtone;
  u8 unknown1:2,
     txtmode:2,
     unknown2:2,
     rxtmode:2;
  u8 duplex;
  u8 unknown3:3,
     isnarrow:1,
     unknown4:2,
     not_scramble:1,
     not_revfreq:1;
  u8 flag3;
  u8 step;
  u8 power;
};

struct bandflag {
    u8 scanadd:1,
        unknown1:3,
        band:4;
};

struct tguv2_config {
    u8 unknown1;
    u8 squelch;
    u8 time_out_timer;
    u8 priority_channel;

    u8 unknown2:7,
        keyunlocked:1;
    u8 busy_lockout;
    u8 vox;
    u8 unknown3;

    u8 beep_tone_disabled;
    u8 display;
    u8 step;
    u8 unknown4;

    u8 unknown5;
    u8 rxmode;
    u8 unknown6:7,
        not_end_tone_elim:1;
    u8 vfo_mode;
};

struct vfo {
    u8 current;
    u8 chan;
    u8 memno;
};

struct name {
  u8 name[6];
  u8 unknown1[10];
};

#seekto 0x0000;
char ident[32];
u8 blank[16];

struct memory channels[200];
struct memory bands[5];

#seekto 0x0D30;
struct bandflag bandflags[200];

#seekto 0x0E30;
struct tguv2_config settings;
struct vfo vfos[2];
u8 unk5;
u8 reserved2[9];
u8 band_restrict;
u8 txen350390;

#seekto 0x0F30;
struct name names[200];

`

const tguv2pCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_|* +-"

var tguv2pSteps = []float64{5, 6.25, 10, 12.5, 15, 20, 25, 30, 50, 100}

var tguv2pBands = []band{{88000000, 108000000}, {136000000, 174000000}, {350000000, 390000000}, {400000000, 470000000}, {470000000, 520000000}}

type tguv2pFam struct{}

func init() {
	genRegister(&genModel{ID: "qs-tguv2p", Vendor: "Quansheng", Model: "TG-UV2+", RClass: "QuanshengTGUV2P", Name: "Quansheng TG-UV2+ / TG-UV2",
		Slots: 200, First: 0, NameMax: 6, Charset: tguv2pCharset, Powers: []genPower{{"Alta 10 W", 10, 0}, {"Media 5 W", 5, 1}, {"Bassa 1 W", 1, 2}},
		Bands: tguv2pBands, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: tguv2pFam{},
		Notes: "Accendi la radio con il cavo collegato (9600 baud, due bit di stop). 200 canali (da 0 a 199) con nomi di 6 caratteri; " +
			"passo, scrambler e inversione del ponte di ogni canale si conservano. Driver di CHIRP (sperimentale)."})
}

func (tguv2pFam) mem() *bwEl                 { return bwMust(tguv2pFormat) }
func (f tguv2pFam) layout(m *genModel) *bwEl { return f.mem() }
func (f tguv2pFam) recs(i int) (*bwEl, *bwEl, *bwEl) {
	b := f.mem()
	return b.F("channels").I(i), b.F("bandflags").I(i), b.F("names").I(i).F("name")
}
func (f tguv2pFam) Rec(m *genModel, img []byte, i int) []byte {
	r, bf, n := f.recs(i)
	return append(append(append([]byte{}, r.Raw(img)...), bf.Raw(img)...), n.Raw(img)...)
}
func (f tguv2pFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, bf, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 23 {
		bf.SetRaw(img, raw[16:17])
		n.SetRaw(img, raw[17:23])
	}
}
func (f tguv2pFam) Match(m *genModel, img []byte) bool {
	return bytes.HasPrefix(img, []byte("TG-UV2+ Radio Program Data")) && len(img) == 0x2030
}
func (tguv2pFam) crossband() bool { return true }

func tguv2pToneDec(v, mode int64) string {
	switch {
	case v > 104 || mode > 3 || mode == 0:
		return ""
	case mode == 1:
		if int(v) < len(chirpTones) {
			return fmtTone(chirpTones[v])
		}
		return ""
	}
	return genTone{Mode: "DTCS", Code: dtcsCodes[v], Inv: mode == 3}.String()
}

func tguv2pToneEnc(s string) (int64, int64, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, 0, err
	}
	switch t.Mode {
	case "Tone":
		k := slices.Index(chirpTones, t.Hz10)
		if k < 0 {
			return 0, 0, fmt.Errorf("tono %s non ammesso", fmtTone(t.Hz10))
		}
		return int64(k), 1, nil
	case "DTCS":
		k := slices.Index(dtcsCodes, t.Code)
		if k < 0 {
			return 0, 0, fmt.Errorf("codice DCS %03d non ammesso", t.Code)
		}
		if t.Inv {
			return int64(k), 3, nil
		}
		return int64(k), 2, nil
	}
	return 0, 0, nil
}

func (f tguv2pFam) Get(m *genModel, img []byte, i int) *Chan {
	r, bf, n := f.recs(i)
	if img[r.Off] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("freq").Int(img) * 10}
	off := int64(0)
	if img[r.F("offset").Off] != 0xFF {
		off = r.F("offset").Int(img) * 10
	}
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx + off
	case 2:
		c.Tx = c.Rx - off
	default:
		c.Tx = c.Rx
	}
	c.TxTone = tguv2pToneDec(r.F("txtone").Int(img), r.F("txtmode").Int(img))
	c.RxTone = tguv2pToneDec(r.F("rxtone").Int(img), r.F("rxtmode").Int(img))
	c.Narrow = r.F("isnarrow").Bool(img)
	c.Scan = bf.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	var nm strings.Builder
	for _, b := range n.Raw(img) {
		if int(b) >= len(tguv2pCharset) {
			break
		}
		nm.WriteByte(tguv2pCharset[b])
	}
	c.Name = strings.TrimRight(nm.String(), " \t\n\r\x0b\x0c")
	return c
}

func (f tguv2pFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, bf, n := f.recs(i)
	if c == nil {
		bf.Fill(img, 0xFF)
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	tv, tm, err := tguv2pToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rv, rm, err := tguv2pToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	var nm []byte
	for k := 0; k < 6; k++ {
		if k >= len(c.Name) {
			nm = append(nm, 0xFF)
			continue
		}
		x := strings.IndexByte(tguv2pCharset, c.Name[k])
		if x < 0 {
			return fmt.Errorf("carattere %q non ammesso nel nome", c.Name[k])
		}
		nm = append(nm, byte(x))
	}
	old := f.Get(m, img, i)
	// valori conservati come in CHIRP (get_memory): scrambler, inversione,
	// passo (oltre i 10 passi → 5 kHz), scostamento memorizzato
	ns, nr, step, off := int64(1), int64(1), int64(0), int64(600000)
	if old != nil {
		ns, nr = r.F("not_scramble").Int(img), r.F("not_revfreq").Int(img)
		if step = r.F("step").Int(img); step > int64(len(tguv2pSteps)) {
			step = 0
		}
		off = 0
		if img[r.F("offset").Off] != 0xFF {
			off = r.F("offset").Int(img) * 10
		}
	}
	bf.Fill(img, 0xFF)
	r.SetRaw(img, append(append(make([]byte, 12), 0xFF, 0xFF), 0, 0))
	bf.F("scanadd").SetBool(img, c.Scan)
	bf.F("band").SetInt(img, 0x0F)
	for k, b := range tguv2pBands {
		if c.Rx >= b.Lo && c.Rx <= b.Hi {
			bf.F("band").SetInt(img, int64(k))
			break
		}
	}
	r.F("freq").SetInt(img, c.Rx/10)
	switch {
	case c.Tx > c.Rx:
		off = c.Tx - c.Rx
		r.F("duplex").SetInt(img, 1)
	case c.Tx < c.Rx:
		off = c.Rx - c.Tx
		r.F("duplex").SetInt(img, 2)
	}
	r.F("offset").SetInt(img, off/10)
	r.F("txtone").SetInt(img, tv)
	r.F("txtmode").SetInt(img, tm)
	r.F("rxtone").SetInt(img, rv)
	r.F("rxtmode").SetInt(img, rm)
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("step").SetInt(img, step)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	n.SetRaw(img, nm)
	r.F("not_scramble").SetInt(img, ns)
	r.F("not_revfreq").SetInt(img, nr)
	return nil
}

func (s *rawSess) tguv2pProg() error {
	if err := s.send([]byte("\x02PnOGdAM")); err != nil {
		return err
	}
	for k := 0; k < 10; k++ {
		b := readUpTo(s.ctx, s.p, 1, 3*time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 1 && b[0] == 0x06 {
			return nil
		}
	}
	return rErr("la radio non è entrata in modalità di programmazione")
}

func (s *rawSess) tguv2pIdent() error {
	if err := s.tguv2pProg(); err != nil {
		return err
	}
	if err := s.send([]byte{0x4D, 0x02}); err != nil {
		return err
	}
	if id := readUpTo(s.ctx, s.p, 8, 3*time.Second); !bytes.HasPrefix(id, []byte("P5555")) {
		if err := s.tguv2pProg(); err != nil {
			return err
		}
		if err := s.send([]byte{0x40, 0x02}); err != nil {
			return err
		}
		if id := readUpTo(s.ctx, s.p, 8, 3*time.Second); !bytes.HasPrefix(id, []byte("P5555")) {
			return rErr("modello non supportato (identificativo % X)", id)
		}
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, 3*time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio non ha confermato l'identificazione")
	}
	return nil
}

func tguv2pOpen(ctx context.Context, m *genModel, port string) (*rawSess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	if err := setTwoStopBits(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	return s, nil
}

func (f tguv2pFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := tguv2pOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.tguv2pIdent(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := append([]byte("TG-UV2+ Radio Program Data v1.0\x00"), make([]byte, 16)...)
	var first []byte
	for a := 0; a < 0x2000; a += 8 {
		prog(a, 0x2000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 8}); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 12, 3*time.Second)
		if len(resp) != 12 || resp[0] != 'W' || resp[1] != byte(a>>8) || resp[2] != byte(a) || resp[3] != 8 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("risposta non valida per l'indirizzo 0x%04X", a)
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		ack := readUpTo(ctx, s.p, 1, 3*time.Second)
		if first == nil {
			first = ack
		} else if !bytes.Equal(ack, first) {
			return nil, "", rErr("risposta inattesa leggendo l'indirizzo 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	prog(0x2000, 0x2000, "Lettura completata")
	return data, "", nil
}

func (f tguv2pFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x2030 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x2030)
	}
	s, err := tguv2pOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x2000, "Collegamento con la radio…")
	if err := s.tguv2pIdent(); err != nil {
		return "", genFail("Identificazione", err)
	}
	return "", wxUpload(ctx, s, img[0x30:], 0, 0x2000, 8, 3*time.Second, prog)
}
