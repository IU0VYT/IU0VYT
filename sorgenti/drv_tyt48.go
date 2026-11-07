// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: thuv1f.py e th_uvf8d.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: TYT TH-UVF1 (thuv1f.py di CHIRP) e TH-UVF8D (th_uvf8d.py)

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const thuvf1Format = `
struct mem {
  bbcd rx_freq[4];
  bbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];
  u8 unknown1:1,
     pttid:2,
     unknown2:2,
     ishighpower:1,
     unknown3:2;
  u8 unknown4:4,
     isnarrow:1,
     vox:1,
     bcl:2;
  u8 unknown5:1,
     scan:1,
     unknown6:3,
     scramble_code:3;
  u8 unknown7;
};

struct name {
  char name[7];
};

#seekto 0x0020;
struct mem memory[128];

#seekto 0x0840;
struct {
  u8 scans:2,
     autolk:1,
     unknown1:5;
  u8 light:2,
     unknown6:2,
     disnm:1,
     voice:1,
     beep:1,
     rxsave:1;
  u8 led:2,
     unknown5:3,
     ani:1,
     roger:1,
     dw:1;
  u8 opnmsg:2,
     unknown4:1,
     dwait:1,
     unknown9:4;
  u8 squelch;
  u8 unknown2:4,
     tot:4;
  u8 unknown3:4,
     vox_level:4;
  u8 pad[10];
  char ponmsg[6];
} settings;

#seekto 0x08D0;
struct name names[128];
`

type thuvf1Fam struct{}

func init() {
	genRegister(&genModel{ID: "tyt-thuvf1", Vendor: "TYT", Model: "TH-UVF1", RClass: "TYTTHUVF1Radio", Name: "TYT TH-UVF1",
		Slots: 128, First: 1, NameMax: 7, Charset: charsetAlnum + "-/", Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}},
		Bands: []band{{136000000, 174000000}, {420000000, 470000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
		Fam: thuvf1Fam{}, Notes: "Accendi la radio con il cavo collegato. 128 canali con nomi di 7 caratteri; PTT ID, VOX, blocco del canale occupato " +
			"e scrambler di ogni canale si conservano. La scrittura controlla che la radio sia quella letta (identificativo). Driver di CHIRP."})
}

func (thuvf1Fam) mem() *bwEl                 { return bwMust(thuvf1Format) }
func (f thuvf1Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f thuvf1Fam) recs(i int) (*bwEl, *bwEl) {
	return f.mem().F("memory").I(i), f.mem().F("names").I(i).F("name")
}
func (f thuvf1Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}
func (f thuvf1Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	if len(raw) >= 23 {
		n.SetRaw(img, raw[16:23])
	}
}

var thuvf1IDs = [][]byte{
	[]byte("\x13\x60\x17\x40\x40\x00\x48\x00\x35\x00\x39\x00\x47\x00\x52\x00"), // TH-UVF1
	[]byte("\x14\x40\x14\x80\x43\x00\x45\x00\x13\x60\x17\x40\x40\x00\x47\x00"), // TH-UVF1 V2
}

func (f thuvf1Fam) Match(m *genModel, img []byte) bool {
	for _, id := range thuvf1IDs {
		if bytes.HasPrefix(img, id) {
			return true
		}
	}
	return false
}

// toni (thuv1f.py): byte basso a 0xFF = spento; DCS con 0x8000 (0xC000
// invertito) e codice in cifre esadecimali; CTCSS in BCD
func thuvf1ToneDec(v int64) string {
	switch {
	case v&0xFF == 0xFF:
		return ""
	case v&0xC000 == 0xC000, v&0x8000 != 0:
		t, ok := thuv3rToneDec(v&0x3FFF | 0x8000)
		if !ok {
			return ""
		}
		return genTone{Mode: "DTCS", Code: t.v, Inv: v&0xC000 == 0xC000}.String()
	}
	t, ok := thuv3rToneDec(v)
	if !ok {
		return ""
	}
	return fmtTone(t.v)
}

// le16: valore grezzo di due byte (il primo è il meno significativo)
func le16(b []byte) int64 { return int64(b[0]) | int64(b[1])<<8 }

func thuvf1ToneEnc(s string) (int64, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, err
	}
	switch t.Mode {
	case "Tone":
		return thHex(t.Hz10), nil
	case "DTCS":
		v := thHex(t.Code) | 0x8000
		if t.Inv {
			v |= 0x4000
		}
		return v, nil
	}
	return 0xFFFF, nil
}

func (f thuvf1Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if bytes.HasPrefix(r.Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	if bytes.Equal(r.F("tx_freq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("tx_freq").Int(img) * 10
	}
	c.TxTone = thuvf1ToneDec(le16(r.F("tx_tone").Raw(img)))
	c.RxTone = thuvf1ToneDec(le16(r.F("rx_tone").Raw(img)))
	c.Name = strings.TrimRight(strings.ReplaceAll(string(n.Raw(img)), "\xFF", " "), " \t\n\r\x0b\x0c")
	c.Scan = r.F("scan").Bool(img)
	c.Narrow = r.F("isnarrow").Bool(img)
	c.Power = m.powerSem(int(1 - r.F("ishighpower").Int(img)))
	return c
}

func (f thuvf1Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	tx, err := thuvf1ToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rx, err := thuvf1ToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, 16)) {
		r.Fill(img, 0)
	}
	r.F("rx_freq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("tx_freq").Fill(img, 0xFF)
	} else {
		r.F("tx_freq").SetInt(img, c.Tx/10)
	}
	r.F("tx_tone").SetRaw(img, []byte{byte(tx), byte(tx >> 8)})
	r.F("rx_tone").SetRaw(img, []byte{byte(rx), byte(rx >> 8)})
	nm := []byte(c.Name)
	for len(nm) < 7 {
		nm = append(nm, 0xFF)
	}
	n.SetRaw(img, nm[:7])
	r.F("scan").SetBool(img, c.Scan)
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 0)
	return nil
}

// thuvf1Ident: "PROG333" → 06; 02 → identificativo di 16 byte; 06 → 06
func (s *rawSess) thuvf1Ident() ([]byte, error) {
	if err := s.send([]byte("PROG333")); err != nil {
		return nil, err
	}
	b := readUpTo(s.ctx, s.p, 1, 2*time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return nil, rErr("risposta inattesa della radio")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return nil, err
	}
	id := readUpTo(s.ctx, s.p, 16, 2*time.Second)
	if err := s.send([]byte{0x06}); err != nil {
		return nil, err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); len(b) != 1 || b[0] != 0x06 {
		return nil, rErr("la radio non ha confermato l'identificazione")
	}
	return id, nil
}

func (f thuvf1Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	id, err := s.thuvf1Ident()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, 0x1000, 0x10, 2*time.Second, prog, true)
	if err != nil {
		return nil, "", err
	}
	_ = s.send([]byte{0x45})
	return append(id, data...), "", nil
}

func (f thuvf1Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x1010 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x1010)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	id, err := s.thuvf1Ident()
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	if !bytes.Equal(id, img[:16]) {
		return "", rErr("la radio collegata non è quella da cui è stata letta questa memoria (identificativo diverso): nessun dato è stato scritto")
	}
	if err := wxUpload(ctx, s, img[0x10:], 0, 0x1000, 0x10, time.Second, prog); err != nil {
		return "", err
	}
	_ = s.send([]byte{0x45})
	return "", nil
}

// ---------------------------------------------------------------- TH-UVF8D

const thuvf8dFormat = `
struct memory {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  lbcd rx_tone[2];
  lbcd tx_tone[2];

  u8 apro:4,
     rpt_md:2,
     unknown1:2;
  u8 bclo:2,
     wideband:1,
     ishighpower:1,
     scanadd:1,
     vox:1,
     pttid:2;
  u8 unknown3:8;

  u8 unknown4:6,
     duplex:2;

  lbcd offset[4];

  char unknown5[4];

  char name[7];

  char unknown6[1];
};

struct fm_broadcast_memory {
  lbcd freq[3];
  u8 unknown;
};

struct enable_flags {
  bit flags[8];
};

#seekto 0x0020;
struct memory channels[128];

#seekto 0x2020;
struct memory vfo1;
struct memory vfo2;

#seekto 0x2060;
struct {
  u8 unknown2060:4,
     tot:4;
  u8 unknown2061;
  u8 squelch;
  u8 unknown2063:4,
     vox_level:4;
  u8 tuning_step;
  char unknown12;
  u8 lamp_t;
  char unknown11;
  u8 unknown2068;
  u8 ani:1,
     scan_mode:2,
     unknown2069:2,
     beep:1,
     tx_sel:1,
     roger:1;
  u8 light:2,
     led:2,
     taileliminate:1,
     autolk:1,
     unknown206ax:2;
  u8 unknown206b:1,
     b_display:2,
     a_display:2,
     ab_switch:1,
     dwait:1,
     mode:1;
  u8 dw:1,
     unknown206c:6,
     voice:1;
  u8 unknown206d:2,
     rxsave:2,
     opnmsg:2,
     lock_mode:2;
  u8 a_work_area:1,
     b_work_area:1,
     unknown206ex:6;
  u8 a_channel;
  u8 b_channel;
  u8 pad3[15];
  char ponmsg[7];
} settings;

#seekto 0x2E60;
struct enable_flags enable[16];
struct enable_flags skip[16];

#seekto 0x2FA0;
struct fm_broadcast_memory fm_current;

#seekto 0x2FA8;
struct fm_broadcast_memory fm_memories[20];
`

type thuvf8dFam struct{}

func init() {
	genRegister(&genModel{ID: "tyt-thuvf8d", Vendor: "TYT", Model: "TH-UVF8D", RClass: "TYTUVF8DRadio", Name: "TYT TH-UVF8D",
		Slots: 128, First: 1, NameMax: 7, Charset: charsetUpperNum + "-", Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands: []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600,
		Fam: thuvf8dFam{}, Notes: "Accendi la radio con il cavo collegato. 128 canali (banco A) con nomi di 7 caratteri; PTT ID, VOX, blocco del canale occupato, " +
			"scrambler/compander e modalità ripetitore di ogni canale si conservano. La scrittura controlla che la radio sia quella letta. Driver di CHIRP."})
}

func (thuvf8dFam) mem() *bwEl                 { return bwMust(thuvf8dFormat) }
func (f thuvf8dFam) layout(m *genModel) *bwEl { return f.mem() }
func (f thuvf8dFam) rec(i int) *bwEl          { return f.mem().F("channels").I(i) }
func (f thuvf8dFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f thuvf8dFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f thuvf8dFam) Match(m *genModel, img []byte) bool {
	return bytes.HasPrefix(img, []byte("TYT-F10\x00"))
}
func (thuvf8dFam) crossband() bool { return true }

// indicatori «canale presente» e «da non saltare»: un bit per canale (bit 0 = primo)
func (f thuvf8dFam) flag(img []byte, name string, i int) bool {
	return img[f.mem().F(name).Off+i/8]&(1<<(i%8)) != 0
}

func (f thuvf8dFam) setFlag(img []byte, name string, i int, v bool) {
	a := f.mem().F(name).Off + i/8
	if v {
		img[a] |= 1 << (i % 8)
	} else {
		img[a] &^= 1 << (i % 8)
	}
}

func (f thuvf8dFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if bytes.HasPrefix(r.Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) || !f.flag(img, "enable", i) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	off := r.F("offset").Int(img) * 10
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - off
	case 2:
		c.Tx = c.Rx + off
	default:
		c.Tx = c.Rx
	}
	c.TxTone = thuvf1ToneDec(le16(r.F("tx_tone").Raw(img)))
	c.RxTone = thuvf1ToneDec(le16(r.F("rx_tone").Raw(img)))
	c.Name = strings.TrimRight(string(r.F("name").Raw(img)), "\xFF ")
	c.Scan = f.flag(img, "skip", i)
	c.Narrow = !r.F("wideband").Bool(img)
	c.Power = m.powerSem(int(1 - r.F("ishighpower").Int(img)))
	return c
}

func (f thuvf8dFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		f.setFlag(img, "enable", i, false)
		f.setFlag(img, "skip", i, false)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	tx, err := thuvf1ToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	rx, err := thuvf1ToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	wasEmpty := f.Get(m, img, i) == nil
	f.setFlag(img, "enable", i, true)
	if bytes.Equal(r.Raw(img), bytes.Repeat([]byte{0xFF}, 32)) {
		r.Fill(img, 0)
	}
	r.F("rx_freq").SetInt(img, c.Rx/10)
	r.F("tx_freq").SetInt(img, c.Tx/10)
	switch {
	case c.Tx == c.Rx:
		// CHIRP riscrive lo scostamento memorizzato (0,6 MHz per un canale nuovo)
		r.F("duplex").SetInt(img, 0)
		if wasEmpty {
			r.F("offset").SetInt(img, 600000/10)
		}
	case c.Tx < c.Rx:
		r.F("duplex").SetInt(img, 1)
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
	default:
		r.F("duplex").SetInt(img, 2)
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
	}
	r.F("tx_tone").SetRaw(img, []byte{byte(tx), byte(tx >> 8)})
	r.F("rx_tone").SetRaw(img, []byte{byte(rx), byte(rx >> 8)})
	nm := []byte(strings.TrimRight(c.Name, " "))
	for len(nm) < 7 {
		nm = append(nm, 0xFF)
	}
	r.F("name").SetRaw(img, nm[:7])
	f.setFlag(img, "skip", i, c.Scan)
	r.F("scanadd").SetBool(img, c.Scan)
	r.F("wideband").SetBool(img, !c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 0)
	return nil
}

// thuvf8dIdent: "\x02PROGRAM" → "PG"; 02 → identificativo di 32 byte; "A" → "A"
func (s *rawSess) thuvf8dIdent() ([]byte, error) {
	// fino a 3 tentativi (CHIRP ne fa uno): con una radio che risponde subito i byte
	// scambiati sono gli stessi; il saluto ripetuto distingue questa radio, durante
	// l'identificazione automatica, da quelle che inviano lo stesso «\x02PROGRAM»
	// (BF-888S e simili), che scambierebbero la risposta «PG» per l'eco del cavo
	var b []byte
	for k := 0; k < 3 && len(b) == 0; k++ {
		if err := s.send([]byte("\x02PROGRAM")); err != nil {
			return nil, err
		}
		b = readUpTo(s.ctx, s.p, 2, 2*time.Second)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
	}
	if len(b) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if string(b) != "PG" {
		return nil, rErr("la radio non ha confermato il primo comando")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return nil, err
	}
	id := readUpTo(s.ctx, s.p, 32, 2*time.Second)
	if err := s.send([]byte("A")); err != nil {
		return nil, err
	}
	if b := readUpTo(s.ctx, s.p, 1, 2*time.Second); string(b) != "A" {
		return nil, rErr("la radio non ha confermato l'identificazione")
	}
	return id, nil
}

// blocchi da 32 byte confermati con "A" (tyt_uvf8d_download / _upload)
func (f thuvf8dFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x4000, "Collegamento con la radio…")
	id, err := s.thuvf8dIdent()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := id
	for a := 0; a < 0x4000; a += 0x20 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0x4000, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 0x20}); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 0x24, 2*time.Second)
		if len(resp) != 0x24 {
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.send([]byte("A")); err != nil {
			return nil, "", err
		}
		if b := readUpTo(ctx, s.p, 1, 2*time.Second); string(b) != "A" {
			return nil, "", rErr("la radio non ha confermato il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	_ = s.send([]byte("ENDR"))
	prog(0x4000, 0x4000, "Lettura completata")
	return data, "", nil
}

func (f thuvf8dFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x4020 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x4020)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x4000, "Collegamento con la radio…")
	id, err := s.thuvf8dIdent()
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	if !bytes.Equal(id, img[:32]) {
		return "", rErr("la radio collegata non è quella da cui è stata letta questa memoria (identificativo diverso): nessun dato è stato scritto")
	}
	for a := 0; a < 0x4000; a += 0x20 {
		prog(a, 0x4000, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x20}, img[a+0x20:a+0x40]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); string(b) != "A" {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	_ = s.send([]byte("ENDW"))
	readUpTo(ctx, s.p, 3, time.Second)
	prog(0x4000, 0x4000, "Scrittura completata")
	return "", nil
}
