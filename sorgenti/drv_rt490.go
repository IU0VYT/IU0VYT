// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radtel_rt490.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.2: famiglia Radtel RT-490 (driver radtel_rt490.py di CHIRP): Radtel RT-490,
// JC-8629 di MMLradio, JJCC e Socotran, Socotran FB-8629, Jianpai 8800 Plus,
// Boristone 8RS, Abbree AR-869, HamGeek HG-590 (stessa radio, stessa memoria).
// 256 canali (1-256) di 16 byte da 0x0000, nomi di 12 caratteri (GB2312) da
// 0x1000, codici privati dei canali da 0x2000; memoria di 16 KB, 9600 baud.
//
// Protocollo (_enter_programming_mode, do_download, do_upload di radtel_rt490.py):
//   «PROGROMJJCCU» → 0x06 (fino a 5 tentativi); 'F' → 8 byte (le gamme della
//   radio, non controllate);
//   lettura: 'R' indirizzo(2, BE) 0x40 → 'R' indirizzo 0x40 + 64 byte (0x0000-0x3FFF);
//   scrittura: 'W' indirizzo 0x40 + 64 byte → 0x06, nelle zone 0x0000-0x23FF,
//   0x3400-0x3C3F e 0x3F80-0x3FFF (come CHIRP); fine: 'E'.

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const rt490Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 signal;
  u8 pttid;
  u8 dcp:4,
     power:4;
  u8 unknown3_0:1,
     narrow:1,
     unknown3_1:1,
     unknown3_2:1,
     bcl:1,
     scan:1,
     tx_enable:1,
     learn:1;
} memory[256];
struct {
  char name[12];
  u8 ffpad[4];
} memname[256];
struct {
  u8 code[4];
} memcode[256];
`

const (
	rt490MemSize = 0x4000
	rt490Block   = 0x40
	rt490Wait    = time.Second
)

var rt490Ranges = [][2]int{{0x0000, 0x2400}, {0x3400, 0x3C40}, {0x3F80, 0x4000}}

type rt490Fam struct{}

func init() {
	pw := []genPower{{"Alta 5 W", 5, 0}, {"Bassa 3 W", 3, 2}}
	bands := []band{{108000000, 136000000}, {136000000, 180000000}, {200000000, 260000000}, {350000000, 400000000}, {400000000, 520000000}}
	add := func(id, vendor, model, rclass, name, note string) {
		genRegister(&genModel{ID: id, Vendor: vendor, Model: model, RClass: rclass, Name: name, Slots: 256, First: 1, NameMax: 12,
			Powers: pw, Bands: bands, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: rt490Fam{},
			Notes: "Accendi la radio con il cavo collegato. Canali 1-256. Driver di CHIRP della famiglia Radtel RT-490 " +
				"(«sperimentale» in CHIRP). I canali in AM restano in AM se la frequenza resta nella banda aerea (108-136 MHz)." + note})
	}
	same := " Stessa radio della Radtel RT-490: la radio non comunica la marca, scegli il modello giusto."
	add("rtl-rt490", "Radtel", "RT-490", "RT490Radio", "Radtel RT-490", "")
	add("mml-jc8629", "MMLradio", "JC-8629", "MML8629Radio", "MMLradio JC-8629", same)
	add("jjcc-jc8629", "JJCC", "JC-8629", "JJCC8629Radio", "JJCC JC-8629", same)
	add("soc-jc8629", "Socotran", "JC-8629", "SocotranJC8629Radio", "Socotran JC-8629", same)
	add("soc-fb8629", "Socotran", "FB-8629", "SocotranFB8629Radio", "Socotran FB-8629", same)
	add("jp-8800plus", "Jianpai", "8800_Plus", "Jianpai8629Radio", "Jianpai 8800 Plus", same)
	add("bor-8rs", "Boristone", "8RS", "Boristone8RSRadio", "Boristone 8RS", same)
	add("ab-ar869", "Abbree", "AR-869", "AbbreeAR869Radio", "Abbree AR-869", same)
	add("hg-hg590", "HamGeek", "HG-590", "HamGeekHG590Radio", "HamGeek HG-590", same)
}

func (rt490Fam) mem() *bwEl { return bwMust(rt490Format) }

func (f rt490Fam) layout(m *genModel) *bwEl { return f.mem() }

func (f rt490Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("memname").I(i)
}

func (f rt490Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}

func (f rt490Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	n.SetRaw(img, raw[16:32])
}

func (f rt490Fam) Match(m *genModel, img []byte) bool { return false }

// gb2312Dec: come bytes.decode('gb2312') di Python (EUC-CN): byte < 0x80 da soli,
// altrimenti coppie con il primo byte 0xA1-0xF7 e il secondo 0xA1-0xFE; un
// errore dà ok = false
func gb2312Dec(b []byte) (string, bool) {
	for i := 0; i < len(b); i++ {
		if b[i] < 0x80 {
			continue
		}
		if b[i] < 0xA1 || b[i] > 0xF7 || i+1 >= len(b) || b[i+1] < 0xA1 || b[i+1] > 0xFE {
			return "", false
		}
		i++
	}
	s, err := simplifiedchinese.GBK.NewDecoder().Bytes(b)
	if err != nil || bytes.ContainsRune(s, '�') {
		return "", false
	}
	return string(s), true
}

// gb2312Enc: caratteri ASCII e cinesi della GB2312 (gli altri: «?»)
func gb2312Enc(s string) []byte {
	enc := simplifiedchinese.GBK.NewEncoder()
	var out []byte
	for _, r := range s {
		if r < 0x80 {
			out = append(out, byte(r))
			continue
		}
		b, err := enc.Bytes([]byte(string(r)))
		if err != nil || len(b) != 2 || b[0] < 0xA1 || b[0] > 0xF7 || b[1] < 0xA1 {
			b = []byte{'?'}
		}
		out = append(out, b...)
	}
	return out
}

// tono: 0 o 0xFFFF spento, sotto 670 DCS (indice+1 nella lista con il 645, +105
// se invertito), altrimenti CTCSS in decimi di Hz (_decode_tone / _encode_tone)
func rt490ToneDec(v int64) string {
	switch {
	case v == 0 || v == 0xFFFF:
		return ""
	case v < 670:
		v--
		idx := v % int64(len(uv5rDTCS))
		return genTone{Mode: "DTCS", Code: uv5rDTCS[idx], Inv: idx != v}.String()
	}
	return fmtTone(int(v))
}

func rt490ToneEnc(t genTone) int64 {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10)
	case "DTCS":
		v := int64(idxInt(uv5rDTCS, t.Code))
		if t.Inv {
			v += int64(len(uv5rDTCS))
		}
		return v + 1
	}
	return 0
}

func (f rt490Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if bytes.Equal(r.F("rxfreq").Raw(img), ff4) {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), ff4) || !r.F("tx_enable").Bool(img) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	// come CHIRP: byte 0xFF finali tolti, poi GB2312 (nome vuoto se non si decodifica)
	c.Name, _ = gb2312Dec(trimBytes(n.F("name").Raw(img), 0xFF))
	c.TxTone = rt490ToneDec(r.F("txtone").Int(img))
	c.RxTone = rt490ToneDec(r.F("rxtone").Int(img))
	pw := int(r.F("power").Int(img))
	if pw >= len(m.Powers) {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = r.F("scan").Bool(img)
	return c
}

func (f rt490Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		n.Fill(img, 0xFF)
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
	was := !bytes.Equal(r.F("rxfreq").Raw(img), ff4)
	am := was && r.F("unknown3_0").Bool(img)
	if !was {
		r.Fill(img, 0) // nuovo canale (come CHIRP)
	}
	nb := gb2312Enc(c.Name)
	nb = append(nb, bytes.Repeat([]byte{0xFF}, 12)...)[:12]
	n.F("name").SetRaw(img, nb)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	r.F("tx_enable").SetBool(img, !c.TxOff)
	if c.TxOff {
		r.F("txfreq").SetInt(img, c.Rx/10) // come il programma del costruttore
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("txtone").SetInt(img, rt490ToneEnc(tt))
	r.F("rxtone").SetInt(img, rt490ToneEnc(rt))
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	// AM: CHIRP lo scrive solo sotto i 300 MHz; il programma non ha il modo AM nei
	// canali, quindi un canale in AM resta in AM nella banda aerea
	r.F("unknown3_0").SetBool(img, am && c.Rx >= 108000000 && c.Rx < 136000000)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("scan").SetBool(img, c.Scan)
	if was {
		// impostazioni del canale lette e riscritte da CHIRP (_get_extra / _set_extra)
		if r.F("dcp").Int(img) > 4 {
			r.F("dcp").SetInt(img, 0)
		}
		if r.F("pttid").Int(img) > 3 {
			r.F("pttid").SetInt(img, 0)
		}
		if r.F("signal").Int(img) < 15 && r.F("learn").Bool(img) {
			// «Learn» solo con un codice privato del canale
			code := f.mem().F("memcode").I(i).F("code").Raw(img)
			r.F("learn").SetBool(img, code[3] == 0xA0)
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (s *rawSess) rt490Start() ([]byte, error) {
	var ack []byte
	for try := 0; try < 5; try++ {
		if err := s.send([]byte("PROGROMJJCCU")); err != nil {
			return nil, err
		}
		ack = readUpTo(s.ctx, s.p, 1, rt490Wait)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(ack) == 1 && ack[0] == 0x06 {
			break
		}
	}
	if len(ack) == 0 {
		return nil, errTimeout
	}
	if ack[0] != 0x06 {
		return nil, rErr("la radio non è entrata in programmazione dopo cinque tentativi (risposta %02X)", ack[0])
	}
	if err := s.send([]byte{'F'}); err != nil {
		return nil, err
	}
	id := make([]byte, 8)
	if err := s.p.ReadFull(s.ctx, id, rt490Wait); err != nil {
		if err == errTimeout {
			return nil, rErr("identificativo della radio incompleto")
		}
		return nil, err
	}
	return id, nil
}

func (f rt490Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, rt490MemSize, "Collegamento con la radio…")
	id, err := s.rt490Start()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := make([]byte, 0, rt490MemSize)
	for a := 0; a < rt490MemSize; a += rt490Block {
		if ctx.Err() != nil {
			_ = s.send([]byte{'E'})
			return nil, "", errCancelled
		}
		prog(a, rt490MemSize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), rt490Block}
		if err := s.send(cmd); err != nil {
			return nil, "", err
		}
		d := make([]byte, 4+rt490Block)
		err := s.p.ReadFull(ctx, d, rt490Wait)
		if err == nil && !bytes.Equal(d[:4], cmd) {
			err = rErr("risposta inattesa")
		}
		if err != nil {
			_ = s.send([]byte{'E'})
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		data = append(data, d[4:]...)
	}
	_ = s.send([]byte{'E'})
	prog(rt490MemSize, rt490MemSize, "Lettura completata")
	return data, fmt.Sprintf("gamme della radio % X", id), nil
}

func (f rt490Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < rt490MemSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), rt490MemSize)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, rt490MemSize, "Collegamento con la radio…")
	if _, err := s.rt490Start(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, rg := range rt490Ranges {
		for a := rg[0]; a < rg[1]; a += rt490Block {
			if ctx.Err() != nil {
				_ = s.send([]byte{'E'})
				return "", errCancelled
			}
			prog(a, rt490MemSize, "Scrittura…")
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), rt490Block}, img[a:a+rt490Block]...)); err != nil {
				return "", err
			}
			ack, err := readByte(ctx, s.p, rt490Wait)
			if err == nil && ack != 0x06 {
				err = rErr("la radio ha rifiutato il blocco")
			}
			if err != nil {
				_ = s.send([]byte{'E'})
				return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
		}
	}
	_ = s.send([]byte{'E'})
	prog(rt490MemSize, rt490MemSize, "Scrittura completata")
	return "", nil
}
