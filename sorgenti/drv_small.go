// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: uvb5.py, bf_t1.py, bf_t8.py, baofeng_uv3r.py, radtel_t18.py, retevis_rt22.py,
// retevis_rt18.py e th350.py (4.7).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio Baofeng "piccole" con protocolli propri: UV-B5/UV-B6 (uvb5.py di
// CHIRP) e BF-T1 (bf_t1.py). Protocolli e mappe come in CHIRP.

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------- comuni

type rawSess struct {
	ctx context.Context
	p   Port
}

func rawOpen(ctx context.Context, port string, baud int) (*rawSess, error) {
	p, err := openPort(port, baud)
	if err != nil {
		return nil, err
	}
	return &rawSess{ctx: ctx, p: p}, nil
}

func (s *rawSess) sleep(d time.Duration) error {
	select {
	case <-s.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// read: esattamente n byte; errTimeout diventa un errore "la radio non risponde"
func (s *rawSess) read(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, to); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

func (s *rawSess) send(b []byte) error { return s.p.Write(b) }

// tono come indice nelle liste di CHIRP: 0 spento, 1-50 CTCSS, 51… DCS
func idxToneDec(v int64, inv bool, dcs []int) string {
	switch {
	case v == 0:
		return ""
	case v < 51:
		if int(v-1) < len(chirpTones) {
			return fmtTone(chirpTones[v-1])
		}
	default:
		if int(v-51) < len(dcs) {
			return genTone{Mode: "DTCS", Code: dcs[v-51], Inv: inv}.String()
		}
	}
	return ""
}

func idxToneEnc(t genTone, dcs []int) (int64, bool, error) {
	switch t.Mode {
	case "Tone":
		k := idxInt(chirpTones, t.Hz10)
		if k < 0 {
			return 0, false, fmt.Errorf("subtono %s non disponibile su questa radio", fmtTone(t.Hz10))
		}
		return int64(k + 1), false, nil
	case "DTCS":
		k := idxInt(dcs, t.Code)
		if k < 0 {
			return 0, false, fmt.Errorf("codice DCS D%03d non disponibile su questa radio", t.Code)
		}
		return int64(k + 51), t.Inv, nil
	}
	return 0, false, nil
}

// ---------------------------------------------------------------- UV-B5 / UV-B6

const uvb5Format = `
struct memory {
  lbcd freq[4];
  lbcd offset[4];
  u8 unknown1:2,
     txpol:1,
     rxpol:1,
     compander:1,
     unknown2:3;
  u8 rxtone;
  u8 txtone;
  u8 pttid:1,
     scanadd:1,
     isnarrow:1,
     bcl:1,
     highpower:1,
     revfreq:1,
     duplex:2;
  u8 step;
  u8 unknown[3];
};

#seekto 0x0000;
char ident[32];
u8 blank[16];
struct memory vfo1;
struct memory channels[99];
#seekto 0x0850;
struct memory vfo2;

#seekto 0x09D0;
u16 fm_presets[16];

#seekto 0x0A30;
struct {
  u8 name[5];
} names[99];
`

// 4.7: TYT TH-350 (th350.py di CHIRP, classe derivata dalla UV-B5): 128 canali,
// toni in due byte (cifre decimali scritte come esadecimali), nessun passo
const th350Format = `
struct memory {
  lbcd freq[4];
  lbcd offset[4];
  u8 unknown1:2,
     txpol:1,
     rxpol:1,
     compander:1,
     scrambler:1,
     unknown2:2;
  u8 rxtoneb;
  u8 rxtonea;
  u8 txtoneb;
  u8 txtonea;
  u8 pttid:1,
     scanadd:1,
     isnarrow:1,
     bcl:1,
     highpower:1,
     revfreq:1,
     duplex:2;
  u8 unknown[2];
};

#seekto 0x0000;
char ident[32];
u8 blank[16];
struct memory vfo1;
struct memory channels[128];
#seekto 0x0840;
struct memory vfo3;
struct memory vfo2;

#seekto 0x09D0;
u16 fm_presets[16];

#seekto 0x0A30;
struct {
  u8 name[5];
} names[128];
`

const uvb5Charset = "0123456789- ABCDEFGHIJKLMNOPQRSTUVWXYZ/_+*"

type uvb5Fam struct{}

// uvb5Opt (4.7): varianti della famiglia; nil = Baofeng UV-B5 / UV-B6
type uvb5Opt struct {
	th350  bool   // TYT TH-350: formato, toni, saluto «TROGRAM», nessuna conferma dei blocchi letti
	header string // intestazione dell'immagine (48 byte con gli zeri)
	us     bool   // 4.9: TH-350 US (Th350USRadio): nomi di 8 caratteri a 0x0930
}

// th350USFormat (4.9): come th350Format, con i nomi di 8 caratteri a 0x0930
var th350USFormat = strings.Replace(th350Format, `#seekto 0x09D0;
u16 fm_presets[16];

#seekto 0x0A30;
struct {
  u8 name[5];
} names[128];`, `#seekto 0x0930;
struct {
  u8 name[8];
} names[128];`, 1)

func uvb5o(m *genModel) *uvb5Opt {
	if o, ok := m.Opt.(*uvb5Opt); ok {
		return o
	}
	return &uvb5Opt{header: "KT511 Radio Program data v1.08\x00\x00"}
}

func init() {
	genRegister(&genModel{ID: "bf-uvb5", Vendor: "Baofeng", Model: "UV-B5", RClass: "BaofengUVB5", Name: "Baofeng UV-B5 / UV-B6",
		Slots: 99, First: 1, NameMax: 5, Charset: uvb5Charset, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{130000000, 175000000}, {220000000, 269000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: uvb5Fam{},
		Notes: "Accendi la radio su un canale libero."})
	genRegister(&genModel{ID: "tyt-th350", Vendor: "TYT", Model: "TH-350", RClass: "Th350Radio", Name: "TYT TH-350",
		Slots: 128, First: 1, NameMax: 5, Charset: uvb5Charset, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{130000000, 175000000}, {220000000, 269000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: uvb5Fam{},
		Opt:   &uvb5Opt{th350: true, header: "TRI350 Radio Program data v1.08\x00"},
		Notes: "Accendi la radio su un canale libero. Driver di CHIRP dichiarato sperimentale dai suoi autori: tieni una copia della lettura."})
	// 4.9: TYT TH-350 versione USA (Th350USRadio): nomi di 8 caratteri
	genRegister(&genModel{ID: "tyt-th350us", Vendor: "TYT", Model: "TH-350", Variant: "US", RClass: "Th350USRadio", Name: "TYT TH-350 (versione USA, nomi di 8 caratteri)",
		Slots: 128, First: 1, NameMax: 8, Charset: uvb5Charset, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{130000000, 175000000}, {220000000, 269000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: uvb5Fam{},
		Opt: &uvb5Opt{th350: true, us: true, header: "TRI350 Radio Program data v1.08\x00"},
		Notes: "Versione USA con i nomi dei canali di 8 caratteri. Accendi la radio su un canale libero. Driver di CHIRP dichiarato " +
			"sperimentale dai suoi autori: tieni una copia della lettura."})
}

func (uvb5Fam) mem(m *genModel) *bwEl {
	if uvb5o(m).us {
		return bwMust(th350USFormat)
	}
	if uvb5o(m).th350 {
		return bwMust(th350Format)
	}
	return bwMust(uvb5Format)
}
func (f uvb5Fam) layout(m *genModel) *bwEl { return f.mem(m) }
func (f uvb5Fam) recs(m *genModel, i int) (*bwEl, *bwEl) {
	r := f.mem(m)
	return r.F("channels").I(i), r.F("names").I(i).F("name")
}
func (f uvb5Fam) Match(m *genModel, img []byte) bool {
	h := uvb5o(m).header
	return bytes.HasPrefix(img, []byte(h[:24])) && len(img) == 0x1030
}

func (f uvb5Fam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(m, i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}

func (f uvb5Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(m, i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}

// th350ToneDec: toni della TH-350 (_decode_tone di th350.py)
func th350ToneDec(a, b int64) string {
	switch {
	case a == 0xFF:
		return ""
	case a >= 0x80:
		pold := a / 16
		if pold != 0x8 && pold != 0xC {
			return ""
		}
		code := (a%16)*100 + b/16*10 + b%16
		return genTone{Mode: "DTCS", Code: int(code), Inv: pold == 0xC}.String()
	}
	// 107.2 → 0x10 0x72
	return fmtTone(int(a/16*1000 + a%16*100 + b/16*10 + b%16))
}

// th350ToneEnc: come _encode_tone di th350.py (stessi calcoli in virgola mobile)
func th350ToneEnc(t genTone) (int64, int64) {
	switch t.Mode {
	case "Tone":
		val := float64(t.Hz10) / 10
		a := int64(math.Floor(val/100))*16 + int64(math.Mod(math.Floor(val/10), 10))
		b := int64(math.Floor(math.Mod(val, 10)))*16 + int64(math.Mod(math.Floor(val*10), 10))
		return a, b
	case "DTCS":
		a := int64(0x80)
		if t.Inv {
			a = 0xC0
		}
		v := int64(t.Code)
		return a + v/100, (v/10)%10*16 + v%10
	}
	return 0xFF, 0xFF
}

func uvb5ToneDec(v int64, pol bool) string {
	switch {
	case v > int64(50+len(dtcsCodes)) || v <= 0: // 4.23: indice fuori dalla tabella dei codici DCS (file rovinato)
		return ""
	case v > 50:
		return genTone{Mode: "DTCS", Code: dtcsCodes[v-51], Inv: pol}.String()
	}
	return fmtTone(chirpTones[v-1])
}

func (f uvb5Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, nam := f.recs(m, i)
	if r.F("freq").Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("freq").Int(img) * 10}
	off := r.F("offset").Int(img) * 10
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - off
	case 2:
		c.Tx = c.Rx + off
	default:
		c.Tx = c.Rx
	}
	if uvb5o(m).th350 {
		c.TxTone = th350ToneDec(r.F("txtonea").Int(img), r.F("txtoneb").Int(img))
		c.RxTone = th350ToneDec(r.F("rxtonea").Int(img), r.F("rxtoneb").Int(img))
	} else {
		c.TxTone = uvb5ToneDec(r.F("txtone").Int(img), r.F("txpol").Bool(img))
		c.RxTone = uvb5ToneDec(r.F("rxtone").Int(img), r.F("rxpol").Bool(img))
	}
	c.Narrow = r.F("isnarrow").Bool(img)
	c.Scan = r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	var name strings.Builder
	for _, b := range nam.Raw(img) {
		if int(b) >= len(uvb5Charset) {
			break
		}
		name.WriteByte(uvb5Charset[b])
	}
	c.Name = strings.TrimRight(name.String(), " \t\n\r\x0b\x0c")
	return c
}

func (f uvb5Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, nam := f.recs(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	raw := r.Raw(img)
	wasEmpty := r.F("freq").Raw(img)[0] == 0xFF
	var keep [4]int64
	if !wasEmpty {
		keep = [4]int64{r.F("bcl").Int(img), r.F("revfreq").Int(img), r.F("pttid").Int(img), r.F("compander").Int(img)}
	}
	if bytes.Equal(raw, bytes.Repeat([]byte{0xFF}, 16)) {
		r.SetRaw(img, append(make([]byte, 13), 0xFF, 0xFF, 0xFF))
	}
	r.F("freq").SetInt(img, c.Rx/10)
	switch {
	case c.Tx == c.Rx:
		// CHIRP riscrive lo scostamento memorizzato (0,6 MHz per un canale nuovo)
		if wasEmpty {
			r.F("offset").SetInt(img, 600000/10)
		}
		r.F("duplex").SetInt(img, 0)
	case c.Tx < c.Rx:
		r.F("duplex").SetInt(img, 1)
		r.F("offset").SetInt(img, (c.Rx-c.Tx)/10)
	default:
		r.F("duplex").SetInt(img, 2)
		r.F("offset").SetInt(img, (c.Tx-c.Rx)/10)
	}
	for _, x := range []struct {
		t        string
		tone, pl string
	}{{c.TxTone, "txtone", "txpol"}, {c.RxTone, "rxtone", "rxpol"}} {
		t, err := genToneParse(x.t)
		if err != nil {
			return err
		}
		if uvb5o(m).th350 {
			// TH-350: la polarità è nel tono (i bit txpol/rxpol non si toccano)
			a, b := th350ToneEnc(t)
			r.F(x.tone+"a").SetInt(img, a)
			r.F(x.tone+"b").SetInt(img, b)
			continue
		}
		v, inv, err := idxToneEnc(t, dtcsCodes)
		if err != nil {
			return err
		}
		r.F(x.pl).SetBool(img, inv)
		r.F(x.tone).SetInt(img, v)
	}
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("scanadd").SetBool(img, c.Scan)
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	r.F("bcl").SetInt(img, keep[0])
	r.F("revfreq").SetInt(img, keep[1])
	r.F("pttid").SetInt(img, keep[2])
	r.F("compander").SetInt(img, keep[3])
	for k := 0; k < m.NameMax; k++ {
		if k < len(c.Name) {
			x := strings.IndexByte(uvb5Charset, c.Name[k])
			if x < 0 {
				return fmt.Errorf("carattere %q non ammesso nel nome", c.Name[k])
			}
			img[nam.Off+k] = byte(x)
		} else {
			img[nam.Off+k] = 0xFF
		}
	}
	return nil
}

func (s *rawSess) uvb5Ident(m *genModel) error {
	hello := "\x05PROGRAM"
	if uvb5o(m).th350 {
		hello = "\x05TROGRAM"
	}
	if err := s.send([]byte(hello)); err != nil {
		return err
	}
	saw, ok := false, false
	for x := 0; x < 10 && !ok; x++ {
		b, err := s.read(1, 3*time.Second)
		if err == errCancelled {
			return err
		}
		if err != nil {
			continue
		}
		saw = true
		ok = b[0] == 0x06
	}
	if !ok {
		if !saw {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non è entrata in modalità di programmazione")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, 3*time.Second)
	if !bytes.HasPrefix(id, []byte("HKT511")) {
		return rErr("modello non supportato (identificativo %q)", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b, err := s.read(1, 3*time.Second); err != nil || b[0] != 0x06 {
		if err == errCancelled {
			return err
		}
		return rErr("la radio non ha confermato l'identificazione")
	}
	return nil
}

func (f uvb5Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	if err := s.uvb5Ident(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := append([]byte(uvb5o(m).header), make([]byte, 16)...)
	th := uvb5o(m).th350
	var first []byte
	for a := 0; a < 0x1000; a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0x1000, "Lettura…")
		fr := []byte{'R', byte(a >> 8), byte(a), 16}
		if err := s.send(fr); err != nil {
			return nil, "", err
		}
		res, err := s.read(20, 3*time.Second)
		if err != nil {
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		if !bytes.Equal(res[1:4], fr[1:4]) {
			return nil, "", rErr("risposta non valida per l'indirizzo 0x%04X", a)
		}
		if th {
			data = append(data, res[4:]...) // TH-350: nessuna conferma dopo ogni blocco
			continue
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		ack := readUpTo(ctx, s.p, 1, 3*time.Second)
		if first == nil {
			first = ack
		} else if !bytes.Equal(ack, first) {
			return nil, "", rErr("risposta inattesa dalla radio")
		}
		data = append(data, res[4:]...)
	}
	prog(0x1000, 0x1000, "Lettura completata")
	return data, "", nil
}

func (f uvb5Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x1030 {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x1000, "Collegamento con la radio…")
	if err := s.uvb5Ident(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	data := img[0x30:]
	note := ""
	for a := 0; a < 0x1000; a += 16 {
		prog(a, 0x1000, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 16}, data[a:a+16]...)); err != nil {
			return "", err
		}
		ack := readUpTo(ctx, s.p, 1, 3*time.Second)
		if len(ack) != 1 || ack[0] != 0x06 {
			if a == 0x0F10 {
				// radio con 27 menu: non accetta le impostazioni di servizio
				note = "la radio non ha le impostazioni di servizio (versione con 27 menu)"
				break
			}
			return "", rErr("la radio ha rifiutato il blocco 0x%04X", a)
		}
	}
	prog(0x1000, 0x1000, "Scrittura completata")
	return note, nil
}

// ---------------------------------------------------------------- BF-T1

const bft1Format = `
struct channel {
  lbcd rxfreq[4];
  u8 rxtone;
  lbcd txoffset[4];
  u8 txtone;
  u8 noskip:1,
     wide:1,
     ttondinv:1,
     unA:1,
     rtondinv:1,
     unB:1,
     offplus:1,
     offminus:1;
  u8 empty[5];
};

#seekto 0x0000;
struct channel emg;
#seekto 0x0010;
struct channel channels[20];

#seekto 0x0150;
struct {
  lbcd vhfl[2];
  lbcd vhfh[2];
  lbcd uhfl[2];
  lbcd uhfh[2];
} settings;
`

type bft1Fam struct{}

func init() {
	genRegister(&genModel{ID: "bf-t1", Vendor: "Baofeng", Model: "BF-T1", RClass: "BFT1", Name: "Baofeng BF-T1",
		Slots: 20, First: 1, NameMax: 0, Powers: []genPower{{"Unica", 1, 0}},
		Bands:  []band{{130000000, 174000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: uv5rDTCS, Baud: 9600, Fam: bft1Fam{},
		Notes: "Radio senza nomi dei canali; la potenza è un'impostazione generale della radio."})
}

func (bft1Fam) mem() *bwEl                 { return bwMust(bft1Format) }
func (f bft1Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f bft1Fam) rec(i int) *bwEl          { return f.mem().F("channels").I(i) }
func (f bft1Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f bft1Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f bft1Fam) Match(m *genModel, img []byte) bool {
	return len(img) >= 0x0700 && string(img[0x06F8:0x0700]) == " BF9100S"
}

// gamme della radio scritte nella memoria (come CHIRP)
func (f bft1Fam) Bands(m *genModel, img []byte) []band {
	if img == nil || len(img) < 0x160 {
		return m.Bands
	}
	st := f.mem().F("settings")
	return []band{{st.F("vhfl").Int(img) * 100000, st.F("vhfh").Int(img) * 100000},
		{st.F("uhfl").Int(img) * 100000, st.F("uhfh").Int(img) * 100000}}
}

func (f bft1Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	off := r.F("txoffset").Int(img) * 10
	c.Tx = c.Rx
	if off != 0 {
		if r.F("offminus").Bool(img) {
			c.Tx = c.Rx - off
		}
		if r.F("offplus").Bool(img) {
			c.Tx = c.Rx + off
		}
	}
	c.Narrow = !r.F("wide").Bool(img)
	c.Scan = r.F("noskip").Bool(img)
	c.TxTone = idxToneDec(r.F("txtone").Int(img), r.F("ttondinv").Bool(img), uv5rDTCS)
	c.RxTone = idxToneDec(r.F("rxtone").Int(img), r.F("rtondinv").Bool(img), uv5rDTCS)
	return c
}

func (f bft1Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	wasEmpty := r.Raw(img)[0] == 0xFF
	r.F("rxfreq").SetInt(img, c.Rx/10)
	switch {
	case c.Tx == c.Rx:
		if wasEmpty {
			// scostamento predefinito di CHIRP per un canale nuovo
			r.F("txoffset").SetInt(img, 600000/10)
		}
		r.F("offplus").SetInt(img, 0)
		r.F("offminus").SetInt(img, 0)
	case c.Tx > c.Rx:
		r.F("txoffset").SetInt(img, (c.Tx-c.Rx)/10)
		r.F("offplus").SetInt(img, 1)
		r.F("offminus").SetInt(img, 0)
	default:
		r.F("txoffset").SetInt(img, (c.Rx-c.Tx)/10)
		r.F("offplus").SetInt(img, 0)
		r.F("offminus").SetInt(img, 1)
	}
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("noskip").SetBool(img, c.Scan)
	for _, x := range []struct{ t, tone, inv string }{{c.TxTone, "txtone", "ttondinv"}, {c.RxTone, "rxtone", "rtondinv"}} {
		t, err := genToneParse(x.t)
		if err != nil {
			return err
		}
		v, inv, err := idxToneEnc(t, uv5rDTCS)
		if err != nil {
			return err
		}
		r.F(x.tone).SetInt(img, v)
		r.F(x.inv).SetBool(img, inv)
	}
	return nil
}

const bft1Timeout = 250 * time.Millisecond // CHIRP usa 70 ms: qui un margine in più

// svuota i byte in arrivo (al massimo 101, come CHIRP)
func (s *rawSess) bft1Clean() error {
	total := 0
	for {
		d := readUpTo(s.ctx, s.p, 100, 70*time.Millisecond)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		total += len(d)
		if total > 101 {
			return rErr("dati continui sulla porta: controlla di aver scelto la porta giusta")
		}
		if len(d) == 0 {
			return nil
		}
	}
}

func (s *rawSess) bft1Ident(magic, id []byte) error {
	if err := s.bft1Clean(); err != nil {
		return err
	}
	ok := false
	for a := 0; a < 3 && !ok; a++ {
		if err := s.send(magic); err != nil {
			return err
		}
		b, err := s.read(1, bft1Timeout)
		if err != nil {
			return err
		}
		ok = b[0] == 0x06
	}
	if !ok {
		return rErr("la radio non è entrata in modalità di programmazione: è il modello giusto?")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	got := readUpTo(s.ctx, s.p, len(id), bft1Timeout)
	if len(got) == 0 {
		return rErr("la radio non risponde")
	}
	if !bytes.Equal(got, id) {
		return rErr("identificazione della radio non riuscita (%q)", got)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	b, err := s.read(1, bft1Timeout)
	if err != nil {
		return err
	}
	if b[0] != 0x06 {
		return rErr("la radio non ha confermato l'identificazione")
	}
	return nil
}

func (f bft1Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x800, "Collegamento con la radio…")
	if err := s.bft1Ident([]byte("\x05PROGRAM"), []byte(" BF9100S")); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	if err := s.bft1Clean(); err != nil {
		return nil, "", err
	}
	var data []byte
	for a := 0; a < 0x800; a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0x800, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 16}); err != nil {
			return nil, "", err
		}
		blk, err := s.read(20, bft1Timeout)
		if err != nil {
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		if !bytes.Equal(blk[:4], []byte{'W', byte(a >> 8), byte(a), 16}) {
			return nil, "", rErr("intestazione non valida per il blocco 0x%04X", a)
		}
		data = append(data, blk[4:]...)
	}
	if err := s.send([]byte{0x62}); err != nil {
		return nil, "", err
	}
	prog(0x800, 0x800, "Lettura completata")
	return data, "", nil
}

func (f bft1Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x180 {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x180, "Collegamento con la radio…")
	if err := s.bft1Ident([]byte("\x05PROGRAM"), []byte(" BF9100S")); err != nil {
		return "", genFail("Identificazione", err)
	}
	if err := s.bft1Clean(); err != nil {
		return "", err
	}
	for a := 0; a < 0x180; a += 16 {
		prog(a, 0x180, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 16}, img[a:a+16]...)); err != nil {
			return "", err
		}
		b, err := s.read(1, bft1Timeout)
		if err != nil {
			return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
		if b[0] != 0x06 {
			return "", rErr("la radio ha rifiutato il blocco 0x%04X", a)
		}
	}
	if err := s.send([]byte{0x62}); err != nil {
		return "", err
	}
	prog(0x180, 0x180, "Scrittura completata")
	return "", nil
}

// ---------------------------------------------------------------- BF-T8 / BF-U9

const bft8Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 rx_tmode;
  u8 rx_tone;
  u8 tx_tmode;
  u8 tx_tone;
  u8 unknown1:3,
     skip:1,
     unknown2:2,
     isnarrow:1,
     lowpower:1;
  u8 unknown3[3];
} memory[99];
`

// bft8FormatN (4.7): canali come in CHIRP (MEM_FORMAT % _upper)
func bft8FormatN(n int) string {
	return strings.Replace(bft8Format, "memory[99]", fmt.Sprintf("memory[%d]", n), 1)
}

// nomi dei canali (Retevis RB27 e simili, MEM_FORMAT_PT2 di bf_t8.py)
const bft8NamesFormat = `
#seekto 0x0D00;
struct {
  char name[6];
  u8 unknown1[2];
} names[99];
`

type bft8Fam struct{}

// bft8Opt: varianti della famiglia (classi di bf_t8.py)
type bft8Opt struct {
	spec  *rbSpec
	names bool  // nomi da 6 caratteri a 0x0D00 (RB27)
	dcs   []int // codici DCS nell'ordine della radio
	// 4.7: n: canali (_upper; 0 = 99); pmr: canali PMR446 a frequenza fissa
	// (RetevisRB627B di bf_t8.py)
	n   int
	pmr bool
	// 4.9: frs: canali FRS a frequenza fissa (Retevis RT16, RB27B, BTECH FRS-A1);
	// murs: canali MURS (Retevis RB27V)
	frs, murs bool
}

// bft8Fixed (4.9): frequenze fisse della radio (PMR446, FRS o MURS) e posizioni
// in cui valgono
func (o *bft8Opt) fixed() ([]int64, int) {
	switch {
	case o.pmr:
		return t18PMR, 16
	case o.frs:
		return t18FRS, 22
	case o.murs:
		return t18MURS, 5
	}
	return nil, 0
}

func init() {
	genRegister(&genModel{ID: "bf-t8", Vendor: "Baofeng", Model: "BF-T8", RClass: "BaofengBFT8Generic", Name: "Baofeng BF-T8 / BF-U9",
		Slots: 99, First: 1, NameMax: 0, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: false, DCS: dtcsCodes, Baud: 9600, Fam: bft8Fam{},
		Opt:   &bft8Opt{spec: bft8Spec, dcs: dtcsCodes},
		Notes: "Radio senza nomi dei canali."})
	// Baofeng G11S: è la Retevis RB27 (classe RetevisRB27 di CHIRP, tabella dei
	// modelli equivalenti di CHIRP); stesso protocollo della BF-T8, 99 canali con
	// nomi di 6 caratteri, banda 136-174 e 400-520 MHz.
	rb27dcs := withCodes(dtcsCodes, 645)
	genRegister(&genModel{ID: "bf-g11s", Vendor: "Retevis", Model: "RB27", RClass: "RetevisRB27", Name: "Baofeng G11S (Retevis RB27)",
		Slots: 99, First: 1, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{136000000, 174000000}, {400000000, 520000000}}, Narrow: true, HasScan: true, DCS: rb27dcs, Baud: 9600, Fam: bft8Fam{},
		Opt: &bft8Opt{spec: rb27Spec, names: true, dcs: rb27dcs},
		Notes: "Baofeng G11S = Retevis RB27 (equivalenza indicata da CHIRP): driver della RB27 di CHIRP; " +
			"nei file .img compare come Retevis RB27."})
	// 4.7: Retevis RB627B (PMR446, classe RetevisRB627B di bf_t8.py): 16 canali
	// a frequenza fissa, nomi di 6 caratteri, 0,5 W
	genRegister(&genModel{ID: "rt-rb627b", Vendor: "Retevis", Model: "RB627B", RClass: "RetevisRB627B", Name: "Retevis RB627B (PMR446)",
		Slots: 16, First: 1, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-",
		Powers: []genPower{{"Alta 0,5 W", 0.5, 0}, {"Bassa 0,5 W", 0.499, 2}},
		Bands:  []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: rb27dcs, Baud: 9600, Fam: bft8Fam{},
		Opt:   &bft8Opt{spec: rb27Spec, names: true, dcs: rb27dcs, n: 16, pmr: true},
		Notes: "PMR446: i 16 canali hanno le frequenze fisse PMR446 (come in CHIRP); si possono cambiare nome, toni e scansione."})
	// 4.9: Retevis RT16 (FRS, senza nomi), RB27B (FRS) e RB27V (MURS), BTECH
	// FRS-A1 (classi di bf_t8.py)
	frsNote := "Radio FRS (USA): i 22 canali hanno le frequenze FRS fisse (come in CHIRP); i canali 8-14 sono a bassa potenza."
	genRegister(&genModel{ID: "rt-rt16", Vendor: "Retevis", Model: "RT16", RClass: "RetevisRT16", Name: "Retevis RT16 (FRS)",
		Slots: 22, First: 1, NameMax: 0, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: false, DCS: dtcsCodes, Baud: 9600, Fam: bft8Fam{},
		Opt: &bft8Opt{spec: bft8Spec, dcs: dtcsCodes, n: 22, frs: true}, Notes: frsNote})
	genRegister(&genModel{ID: "rt-rb27b", Vendor: "Retevis", Model: "RB27B", RClass: "RetevisRB27B", Name: "Retevis RB27B (FRS)",
		Slots: 22, First: 1, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-",
		Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: rb27dcs, Baud: 9600, Fam: bft8Fam{},
		Opt: &bft8Opt{spec: rb27Spec, names: true, dcs: rb27dcs, n: 22, frs: true}, Notes: frsNote})
	genRegister(&genModel{ID: "rt-rb27v", Vendor: "Retevis", Model: "RB27V", RClass: "RetevisRB27V", Name: "Retevis RB27V (MURS)",
		Slots: 5, First: 1, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-",
		Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{136000000, 174000000}}, Narrow: true, HasScan: true, DCS: rb27dcs, Baud: 9600, Fam: bft8Fam{},
		Opt:   &bft8Opt{spec: rb27Spec, names: true, dcs: rb27dcs, n: 5, murs: true},
		Notes: "Radio MURS (USA): i 5 canali hanno le frequenze MURS fisse (come in CHIRP); i canali 1-3 sono a banda stretta."})
	genRegister(&genModel{ID: "bt-frsa1", Vendor: "BTECH", Model: "FRS-A1", RClass: "FRSA1Radio", Name: "BTECH FRS-A1 (FRS)",
		Slots: 22, First: 1, NameMax: 6, Charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890-",
		Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{400000000, 470000000}}, Narrow: true, HasScan: true, DCS: rb27dcs, Baud: 9600, Fam: bft8Fam{},
		Opt: &bft8Opt{spec: frsa1Spec, names: true, dcs: rb27dcs, n: 22, frs: true}, Notes: frsNote})
}

// 4.9: BTECH FRS-A1 (FRSA1Radio di bf_t8.py): identificativo «BF-T8A.», memoria
// fino a 0x0DC0, scrittura di 0x0000-0x0640 e dei nomi 0x0D00-0x0DC0
var frsa1Spec = &rbSpec{magic: []byte("\x02PROGRAM"), finger: []byte("BF-T8A\x2E"), memsize: 0x0DC0,
	ranges: [][2]int{{0x0000, 0x0640}, {0x0D00, 0x0DC0}}, block: 0x10, tries: 5, to: time.Second}

func bft8o(m *genModel) *bft8Opt {
	if o, ok := m.Opt.(*bft8Opt); ok {
		return o
	}
	return &bft8Opt{spec: bft8Spec, dcs: dtcsCodes}
}

func bft8Fmt(m *genModel) string {
	if n := bft8o(m).n; n > 0 {
		return bft8FormatN(n)
	}
	return bft8Format
}

func bft8NamesFmt(m *genModel) string {
	if n := bft8o(m).n; n > 0 {
		return strings.Replace(bft8NamesFormat, "names[99]", fmt.Sprintf("names[%d]", n), 1)
	}
	return bft8NamesFormat
}

func (bft8Fam) mem(m *genModel) *bwEl { return bwMust(bft8Fmt(m)) }
func (f bft8Fam) layout(m *genModel) *bwEl {
	if bft8o(m).names {
		return bwMust(bft8Fmt(m) + bft8NamesFmt(m))
	}
	return f.mem(m)
}
func (f bft8Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f bft8Fam) name(m *genModel, i int) *bwEl {
	return bwMust(bft8Fmt(m) + bft8NamesFmt(m)).F("names").I(i).F("name")
}
func (f bft8Fam) Rec(m *genModel, img []byte, i int) []byte {
	r := append([]byte{}, f.rec(m, i).Raw(img)...)
	if bft8o(m).names {
		r = append(r, f.name(m, i).Raw(img)...)
	}
	return r
}
func (f bft8Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(m, i).SetRaw(img, raw[:16])
	if bft8o(m).names && len(raw) >= 22 {
		f.name(m, i).SetRaw(img, raw[16:22])
	}
}
func (f bft8Fam) Match(m *genModel, img []byte) bool { return false }

func bft8ToneDec(mode, v int64, dcs []int) string {
	switch mode {
	case 1:
		if int(v) < len(chirpTones) {
			return fmtTone(chirpTones[v])
		}
	case 2, 3:
		if int(v) < len(dcs) {
			return genTone{Mode: "DTCS", Code: dcs[v], Inv: mode == 3}.String()
		}
	}
	return ""
}

func (f bft8Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("isnarrow").Bool(img)
	c.Scan = !r.F("skip").Bool(img)
	o := bft8o(m)
	c.TxTone = bft8ToneDec(r.F("tx_tmode").Int(img), r.F("tx_tone").Int(img), o.dcs)
	c.RxTone = bft8ToneDec(r.F("rx_tmode").Int(img), r.F("rx_tone").Int(img), o.dcs)
	lp := int(r.F("lowpower").Int(img))
	c.Power = m.powerSem(lp)
	if o.pmr && t18In(t18PMR, rx) {
		// 4.7 (RB627B): canale PMR446 fisso, come lo mostra CHIRP
		if n := i + 1; n <= 16 {
			c.Rx = t18PMR[n-1]
		}
		c.Tx, c.TxOff, c.Narrow = c.Rx, false, true
		if len(m.Powers) > 1 {
			c.Power = m.Powers[1].Sem
		}
	}
	if o.frs && t18In(t18FRS, rx) {
		// 4.9 (RT16, RB27B, FRS-A1): canale FRS fisso, come lo mostra CHIRP
		// (lo scostamento diventa 0; «off» resta)
		if n := i + 1; n <= 22 {
			c.Rx = t18FRS[n-1]
		}
		if !c.TxOff {
			c.Tx = c.Rx
		}
		c.Narrow = true
		if n := i + 1; n >= 8 && n <= 14 && len(m.Powers) > 1 {
			c.Power = m.Powers[1].Sem
		}
	}
	if o.murs && t18In(t18MURS, rx) {
		// 4.9 (RB27V): canale MURS fisso
		if n := i + 1; n <= 5 {
			c.Rx = t18MURS[n-1]
		}
		c.Tx, c.TxOff = c.Rx, false
		if i+1 <= 3 {
			c.Narrow = true
		}
	}
	if o.names {
		// come CHIRP: 0xFF (anche in mezzo al nome) = spazio, spazi finali tolti
		nb := []byte{}
		for _, b := range f.name(m, i).Raw(img) {
			if b == 0xFF {
				b = ' '
			}
			nb = append(nb, b)
		}
		c.Name = strings.TrimRight(string(nb), " \t\n\r\x0b\x0c")
	}
	return c
}

func (f bft8Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := bft8o(m)
	if c == nil {
		r.Fill(img, 0xFF)
		if o.names {
			f.name(m, i).Fill(img, 0xFF)
		}
		return nil
	}
	r.SetRaw(img, append(make([]byte, 13), 0xFF, 0xFF, 0xFF))
	if o.names {
		nb := bytes.Repeat([]byte{0xFF}, 6)
		copy(nb, c.Name)
		f.name(m, i).SetRaw(img, nb)
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("skip").SetBool(img, !c.Scan)
	for _, x := range []struct{ t, mode, tone string }{{c.TxTone, "tx_tmode", "tx_tone"}, {c.RxTone, "rx_tmode", "rx_tone"}} {
		t, err := genToneParse(x.t)
		if err != nil {
			return err
		}
		switch t.Mode {
		case "Tone":
			k := idxInt(chirpTones, t.Hz10)
			if k < 0 {
				return fmt.Errorf("subtono %s non disponibile su questa radio", fmtTone(t.Hz10))
			}
			r.F(x.mode).SetInt(img, 1)
			r.F(x.tone).SetInt(img, int64(k))
		case "DTCS":
			k := idxInt(o.dcs, t.Code)
			if k < 0 {
				return fmt.Errorf("codice DCS D%03d non disponibile su questa radio", t.Code)
			}
			md := int64(2)
			if t.Inv {
				md = 3
			}
			r.F(x.mode).SetInt(img, md)
			r.F(x.tone).SetInt(img, int64(k))
		}
	}
	if m.powerIndex(c.Power) == 1 {
		r.F("lowpower").SetInt(img, 1)
	}
	return nil
}

// lockedSlotImg (4.7, RB627B): canale PMR446 fisso (escluso dalle prove di scrittura)
func (f bft8Fam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	c := f.Get(m, img, i)
	fx, _ := bft8o(m).fixed()
	return c != nil && fx != nil && t18In(fx, c.Rx)
}

// Locked (4.7, RB627B): i canali PMR446 non si cancellano e non cambiano
// frequenza, banda e potenza (campi non modificabili di CHIRP)
func (f bft8Fam) Locked(m *genModel, i int, old, want *Chan) error {
	o := bft8o(m)
	fx, n := o.fixed()
	if fx == nil {
		return nil
	}
	what := "PMR446"
	switch {
	case o.frs:
		what = "FRS"
	case o.murs:
		what = "MURS"
	}
	if want != nil && t18In(fx, want.Rx) && i < n && want.Rx != fx[i] {
		return fmt.Errorf("nella posizione %d della %s la frequenza %s MHz non è ammessa (canali %s fissi)", i+1, m.Name, fmtMHz(want.Rx), what)
	}
	if old == nil || !t18In(fx, old.Rx) {
		return nil
	}
	// campi non modificabili di CHIRP: frequenza e duplex sempre; banda stretta
	// (PMR, FRS, MURS 1-3) e potenza (PMR, FRS 8-14)
	narrowFix := o.pmr || o.frs || (o.murs && i+1 <= 3)
	powerFix := o.pmr || (o.frs && i+1 >= 8 && i+1 <= 14)
	switch {
	case want == nil:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: non si può cancellare", i+1, m.Name, what)
	case want.Rx != old.Rx || want.Tx != old.Tx || want.TxOff != old.TxOff:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza non modificabile", i+1, m.Name, what)
	case narrowFix && want.Narrow != old.Narrow:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: sempre a banda stretta", i+1, m.Name, what)
	case powerFix && want.Power != old.Power:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: potenza non modificabile", i+1, m.Name, what)
	}
	return nil
}

// protocollo comune a BF-T8 e simili: comando, identificativo, blocchi 'R'/'W' con conferma, 'E'
type rbSpec struct {
	exit          []byte // comando di uscita ("E"; "b" sulle radio tipo T18; vuoto: nessun comando)
	quiet         bool   // in caso di errore di identificazione non invia il comando di uscita
	magic, finger []byte
	memsize       int
	ranges        [][2]int
	block         int
	tries         int
	to            time.Duration
	// 4.0 (famiglia T18 di CHIRP, radtel_t18.py):
	// fingers: identificativi ammessi (inizio degli 8 byte inviati dalla radio),
	// al posto di finger; noAck: nessuna conferma dopo ogni blocco letto
	// (ACK_BLOCK = False); echo: il cavo restituisce i byte inviati (_echo);
	// enterAck: risposta alla conferma dell'identificativo (0xF0 sulla RT647)
	fingers  [][]byte
	noAck    bool
	echo     bool
	enterAck byte
	// 4.2 (famiglia Retevis RT21 di CHIRP, retevis_rt21.py): wblock: blocchi
	// scritti più piccoli di quelli letti (BLOCK_SIZE_UP); noAckFirst: nessuna
	// conferma dopo il primo blocco letto (_ack_1st_block = False); ackReads: il
	// saluto si invia una volta sola e si leggono fino a ackReads byte cercando 0x06
	wblock     int
	noAckFirst bool
	ackReads   int
	// 4.7 (Retevis RT18, retevis_rt18.py di CHIRP): preDelay: attesa prima del
	// saluto (come il programma del costruttore); pwStep: dopo l'ingresso il comando
	// 0x05 restituisce i 6 byte della password di programmazione, confermati con 0x06
	preDelay time.Duration
	pwStep   bool
}

// rbSend: invia e, con i cavi che restituiscono i byte inviati, li scarta
func (s *rawSess) rbSend(sp *rbSpec, b []byte) error {
	if err := s.send(b); err != nil {
		return err
	}
	if sp.echo && len(b) > 0 {
		readUpTo(s.ctx, s.p, len(b), sp.to)
	}
	return nil
}

func (s *rawSess) rbEnter(sp *rbSpec) error {
	ok := false
	got := false
	if sp.preDelay > 0 {
		if err := sleepCtx(s.ctx, sp.preDelay); err != nil {
			return err
		}
	}
	if sp.ackReads > 0 {
		if err := s.rbSend(sp, sp.magic); err != nil {
			return err
		}
		for k := 0; k < sp.ackReads && !ok; k++ {
			b := readUpTo(s.ctx, s.p, 1, sp.to)
			if s.ctx.Err() != nil {
				return errCancelled
			}
			if len(b) == 1 {
				got = true
				ok = b[0] == 0x06
			}
		}
	}
	for k := 0; k < sp.tries && !ok && sp.ackReads == 0; k++ {
		if err := s.rbSend(sp, sp.magic); err != nil {
			return err
		}
		b := readUpTo(s.ctx, s.p, 1, sp.to)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 1 {
			got = true
			ok = b[0] == 0x06
		}
	}
	if !ok {
		if !sp.quiet {
			_ = s.send(sp.exitCmd())
		}
		if !got {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non è entrata in modalità di programmazione: controlla il cavo e riaccendi la radio")
	}
	if err := s.rbSend(sp, []byte{0x02}); err != nil {
		return err
	}
	var id []byte
	known := false
	if sp.fingers != nil {
		id = readUpTo(s.ctx, s.p, 8, sp.to)
		for _, f := range sp.fingers {
			known = known || bytes.HasPrefix(id, f)
		}
	} else {
		id = readUpTo(s.ctx, s.p, len(sp.finger), sp.to)
		known = bytes.Equal(id, sp.finger)
	}
	if !known {
		if !sp.quiet {
			_ = s.send(sp.exitCmd())
		}
		return rErr("la radio ha restituito un identificativo sconosciuto (%q)", id)
	}
	if err := s.rbSend(sp, []byte{0x06}); err != nil {
		return err
	}
	want := byte(0x06)
	if sp.enterAck != 0 {
		want = sp.enterAck
	}
	b := readUpTo(s.ctx, s.p, 1, sp.to)
	if len(b) != 1 || b[0] != want {
		if !sp.quiet {
			_ = s.send(sp.exitCmd())
		}
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if sp.pwStep {
		// la password (o 6 × 0xFF) è anche nella memoria (0x03B0): qui si legge e
		// basta, come fa CHIRP
		if err := s.rbSend(sp, []byte{0x05}); err != nil {
			return err
		}
		if pw := readUpTo(s.ctx, s.p, 6, sp.to); len(pw) != 6 {
			return rErr("la radio non ha risposto al passo della password")
		}
		if err := s.rbSend(sp, []byte{0x06}); err != nil {
			return err
		}
		if b := readUpTo(s.ctx, s.p, 1, sp.to); len(b) != 1 || b[0] != 0x06 {
			return rErr("conferma mancante dopo il passo della password")
		}
	}
	return nil
}

func rt18Spec(sp *rbSpec) *rbSpec {
	sp.preDelay, sp.pwStep = 150*time.Millisecond, true
	return sp
}

func (sp *rbSpec) exitCmd() []byte {
	if sp.exit == nil {
		return []byte{'E'}
	}
	return sp.exit
}

func rbDownload(ctx context.Context, port string, baud int, sp *rbSpec, prog Progress) ([]byte, error) {
	s, err := rawOpen(ctx, port, baud)
	if err != nil {
		return nil, err
	}
	defer s.p.Close()
	prog(0, sp.memsize, "Collegamento con la radio…")
	if err := s.rbEnter(sp); err != nil {
		return nil, genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < sp.memsize; a += sp.block {
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		prog(a, sp.memsize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), byte(sp.block)}
		if err := s.rbSend(sp, cmd); err != nil {
			return nil, err
		}
		resp := readUpTo(ctx, s.p, 4+sp.block, sp.to)
		if len(resp) != 4+sp.block || !bytes.Equal(resp[:4], append([]byte{'W'}, cmd[1:]...)) {
			if !sp.quiet {
				_ = s.send(sp.exitCmd())
			}
			if ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if !sp.noAck && (a != 0 || !sp.noAckFirst) {
			if err := s.rbSend(sp, []byte{0x06}); err != nil {
				return nil, err
			}
			ack := readUpTo(ctx, s.p, 1, sp.to)
			if len(ack) != 1 || ack[0] != 0x06 {
				if !sp.quiet {
					_ = s.send(sp.exitCmd())
				}
				return nil, rErr("nessuna conferma leggendo il blocco 0x%04X", a)
			}
		}
		data = append(data, resp[4:]...)
	}
	if err := s.rbSend(sp, sp.exitCmd()); err != nil {
		return nil, err
	}
	prog(sp.memsize, sp.memsize, "Lettura completata")
	return data, nil
}

func rbUpload(ctx context.Context, port string, baud int, sp *rbSpec, img []byte, prog Progress) error {
	if len(img) < sp.memsize {
		return fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := rawOpen(ctx, port, baud)
	if err != nil {
		return err
	}
	defer s.p.Close()
	prog(0, sp.memsize, "Collegamento con la radio…")
	if err := s.rbEnter(sp); err != nil {
		return genFail("Identificazione", err)
	}
	wb := sp.block
	if sp.wblock > 0 {
		wb = sp.wblock
	}
	for _, r := range sp.ranges {
		for a := r[0]; a < r[1]; a += wb {
			prog(a, sp.memsize, "Scrittura…")
			if err := s.rbSend(sp, append([]byte{'W', byte(a >> 8), byte(a), byte(wb)}, img[a:a+wb]...)); err != nil {
				return err
			}
			ack := readUpTo(ctx, s.p, 1, sp.to)
			if len(ack) != 1 || ack[0] != 0x06 {
				if !sp.quiet {
					_ = s.send(sp.exitCmd())
				}
				if ctx.Err() != nil {
					return errCancelled
				}
				return rErr("la radio ha rifiutato il blocco 0x%04X", a)
			}
		}
	}
	if err := s.rbSend(sp, sp.exitCmd()); err != nil {
		return err
	}
	prog(sp.memsize, sp.memsize, "Scrittura completata")
	return nil
}

var bft8Spec = &rbSpec{magic: []byte("\x02PROGRAM"), finger: []byte("\x2EBF-T6\x2E"), memsize: 0x0B60,
	ranges: [][2]int{{0x0000, 0x0B60}}, block: 0x10, tries: 5, to: time.Second}

// Retevis RB27 (Baofeng G11S): stesso identificativo, memoria fino a 0x1040,
// scrittura di canali/impostazioni (0x0000-0x0640) e nomi (0x0D00-0x1040)
var rb27Spec = &rbSpec{magic: []byte("\x02PROGRAM"), finger: []byte("\x2EBF-T6\x2E"), memsize: 0x1040,
	ranges: [][2]int{{0x0000, 0x0640}, {0x0D00, 0x1040}}, block: 0x10, tries: 5, to: time.Second}

func (f bft8Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	d, err := rbDownload(ctx, port, 9600, bft8o(m).spec, prog)
	return d, "", err
}

func (f bft8Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", rbUpload(ctx, port, 9600, bft8o(m).spec, img, prog)
}

// ---------------------------------------------------------------- UV-3R

const uv3rFormat = `
#seekto 0x0010;
struct {
  lbcd rx_freq[4];
  u8 rxtone;
  lbcd offset[4];
  u8 txtone;
  u8 ishighpower:1,
     iswide:1,
     dtcsinvt:1,
     unknown1:1,
     dtcsinvr:1,
     unknown2:1,
     duplex:2;
  u8 unknown;
  lbcd tx_freq[4];
} tx_memory[99];

#seekto 0x0810;
struct {
  lbcd rx_freq[4];
  u8 rxtone;
  lbcd offset[4];
  u8 txtone;
  u8 ishighpower:1,
     iswide:1,
     dtcsinvt:1,
     unknown1:1,
     dtcsinvr:1,
     unknown2:1,
     duplex:2;
  u8 unknown;
  lbcd tx_freq[4];
} rx_memory[99];
`

type uv3rFam struct{}

func init() {
	genRegister(&genModel{ID: "bf-uv3r", Vendor: "Baofeng", Model: "UV-3R", RClass: "UV3RRadio", Name: "Baofeng UV-3R",
		Slots: 99, First: 1, NameMax: 0, Powers: []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}},
		Bands: []band{{136000000, 235000000}, {400000000, 529000000}}, Narrow: true, HasScan: false, NoTxOff: true,
		DCS: dtcsCodes, OneDCS: true, Baud: 9600, Fam: uv3rFam{}, Notes: "Radio senza nomi dei canali."})
}

func (uv3rFam) mem() *bwEl                 { return bwMust(uv3rFormat) }
func (f uv3rFam) layout(m *genModel) *bwEl { return f.mem() }
func (f uv3rFam) recs(i int) (*bwEl, *bwEl) {
	r := f.mem()
	return r.F("tx_memory").I(i), r.F("rx_memory").I(i)
}
func (f uv3rFam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}
func (f uv3rFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}
func (f uv3rFam) Match(m *genModel, img []byte) bool { return len(img) == 0x0E40 }

func uv3rToneDec(v int64, inv bool) string {
	switch {
	case v == 0 || v == 0xFF:
		return ""
	case v < 0x33:
		return fmtTone(chirpTones[v-1])
	case v-0x33 >= int64(len(dtcsCodes)): // 4.23: indice fuori dalla tabella dei codici DCS (file rovinato)
		return ""
	}
	return genTone{Mode: "DTCS", Code: dtcsCodes[v-0x33], Inv: inv}.String()
}

// i toni sono letti dalla copia "rx" della memoria (come CHIRP); le due
// copie vengono sempre scritte uguali
func (f uv3rFam) Get(m *genModel, img []byte, i int) *Chan {
	_, r := f.recs(i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	off := r.F("offset").Int(img) * 10
	dup := r.F("duplex").Int(img)
	switch {
	case dup == 2:
		c.Tx = c.Rx + off
	case dup == 1:
		c.Tx = c.Rx - off
	case off > 60000000:
		c.Tx = off
	default:
		c.Tx = c.Rx
	}
	c.Power = m.powerSem(int(1 - r.F("ishighpower").Int(img)))
	c.Narrow = !r.F("iswide").Bool(img)
	c.Scan = true
	c.TxTone = uv3rToneDec(r.F("txtone").Int(img), r.F("dtcsinvt").Bool(img))
	c.RxTone = uv3rToneDec(r.F("rxtone").Int(img), r.F("dtcsinvr").Bool(img))
	return c
}

func (f uv3rFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	tm, rm := f.recs(i)
	if c == nil {
		tm.Fill(img, 0xFF)
		rm.Fill(img, 0xFF)
		return nil
	}
	// scostamento memorizzato (per i canali in isofrequenza CHIRP lo conserva;
	// 0,6 MHz per un canale nuovo)
	keepOff := int64(600000 / 10)
	if rm.Raw(img)[0] != 0xFF {
		keepOff = rm.F("offset").Int(img)
		if rm.F("duplex").Int(img) == 0 && keepOff*10 > 60000000 {
			keepOff = rm.F("offset").Int(img)
		}
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	enc := func(t genTone) (int64, error) {
		switch t.Mode {
		case "Tone":
			k := idxInt(chirpTones, t.Hz10)
			if k < 0 {
				return 0, fmt.Errorf("subtono %s non disponibile su questa radio", fmtTone(t.Hz10))
			}
			return int64(k + 1), nil
		case "DTCS":
			k := idxInt(dtcsCodes, t.Code)
			if k < 0 {
				return 0, fmt.Errorf("codice DCS D%03d non disponibile su questa radio", t.Code)
			}
			return int64(k + 0x33), nil
		}
		return 0, nil
	}
	tv, err := enc(tt)
	if err != nil {
		return err
	}
	rv, err := enc(rt)
	if err != nil {
		return err
	}
	for _, r := range []*bwEl{tm, rm} {
		r.F("rx_freq").SetInt(img, c.Rx/10)
		d := c.Tx - c.Rx
		if d < 0 {
			d = -d
		}
		switch {
		case d > 70000000:
			r.F("offset").SetInt(img, d/10)
			if c.Tx > c.Rx {
				r.F("duplex").SetInt(img, 2)
			} else {
				r.F("duplex").SetInt(img, 1)
			}
			r.F("tx_freq").Fill(img, 0xFF)
		case c.Tx == c.Rx:
			r.F("offset").SetInt(img, keepOff)
			r.F("duplex").SetInt(img, 0)
			r.F("tx_freq").SetInt(img, (c.Rx+keepOff*10)/10)
		default:
			r.F("offset").SetInt(img, d/10)
			if c.Tx > c.Rx {
				r.F("duplex").SetInt(img, 2)
			} else {
				r.F("duplex").SetInt(img, 1)
			}
			r.F("tx_freq").SetInt(img, (c.Rx+d)/10)
		}
		r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 0)
		r.F("iswide").SetBool(img, !c.Narrow)
		r.F("dtcsinvt").SetBool(img, tt.Mode == "DTCS" && tt.Inv)
		r.F("dtcsinvr").SetBool(img, rt.Mode == "DTCS" && rt.Inv)
		r.F("txtone").SetInt(img, tv)
		r.F("rxtone").SetInt(img, rv)
	}
	return nil
}

func (s *rawSess) uv3rPrep() error {
	var last error
	for k := 0; k < 10; k++ {
		err := func() error {
			if err := s.send([]byte("\x05PROGRAM")); err != nil {
				return err
			}
			if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				return rErr("la radio non ha risposto al primo comando")
			}
			if err := s.send([]byte{0x02}); err != nil {
				return err
			}
			if id := readUpTo(s.ctx, s.p, 8, time.Second); len(id) != 8 {
				return rErr("la radio non ha inviato l'identificativo")
			}
			if err := s.send([]byte{0x06}); err != nil {
				return err
			}
			if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				return rErr("la radio non ha confermato l'identificativo")
			}
			return nil
		}()
		if err == nil {
			return nil
		}
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if _, ok := err.(radioErr); !ok {
			return err
		}
		last = err
		if err := s.sleep(time.Second); err != nil {
			return err
		}
	}
	return last
}

func (f uv3rFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0xE40, "Collegamento con la radio…")
	if err := s.uv3rPrep(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < 0xE40; a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, 0xE40, "Lettura…")
		if err := s.send([]byte{'R', byte(a >> 8), byte(a), 16}); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 20, time.Second)
		if len(resp) != 20 {
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		readUpTo(ctx, s.p, 1, time.Second)
		data = append(data, resp[4:]...)
	}
	prog(0xE40, 0xE40, "Lettura completata")
	return data, "", nil
}

func (f uv3rFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0xE40 {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0xE40, "Collegamento con la radio…")
	if err := s.uv3rPrep(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < 0xE40; a += 16 {
		prog(a, 0xE40, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 16}, img[a:a+16]...)); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(0xE40, 0xE40, "Scrittura completata")
	return "", nil
}

// ---------------------------------------------------------------- BF-V8A, BF-T20FRS (driver radtel_t18.py)

const t18FormatTpl = `
#seekto 0x0010;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 speccode:1,
       compander:1,
       scramble:1,
       skip:1,
       highpower:1,
       narrow:1,
       unknown2:1,
       bcl:1;
    u8 unknown3[3];
} memory[%d];
`

const t20frsFormat = `
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 jumpcode:1,
       unknown1:2,
       skip:1,
       highpower:1,
       narrow:1,
       unknown2:1,
       bcl:1;
    u8 unknown3[3];
} memory[22];
`

type t18Opt struct {
	format string
	spec   *rbSpec
	// 4.0: varianti della famiglia (radtel_t18.py di CHIRP)
	// potenza come in CHIRP: "rw" letta e scritta; "r" letta, ma il canale scritto
	// è sempre ad alta potenza (RB18); "" non gestita (T18); "low" non letta,
	// il canale scritto è a bassa potenza (RB618); "keep" non letta, il canale
	// scritto conserva la potenza che aveva (nuovo: bassa) (RT647)
	power    string
	reset    byte     // primo dei 4 byte finali del canale nuovo (0xF9; 0x09 BF-V8A/BF-T20FRS; 0xF7 RB29/RB629)
	reserved bool     // gli ultimi 3 byte del canale si conservano anche cancellandolo (H777H)
	extras   []string // bit del canale conservati se il canale esisteva (impostazioni «extra» di CHIRP)
	lock     string   // canali con frequenza fissa: frs, frs16, murs, pmr, gmrs
}

type t18Fam struct{}

// frequenze fisse dei canali (come radtel_t18.py)
var (
	t18FRS1   = []int64{462562500, 462587500, 462612500, 462637500, 462662500, 462687500, 462712500}
	t18FRS2   = []int64{467562500, 467587500, 467612500, 467637500, 467662500, 467687500, 467712500}
	t18FRS3   = []int64{462550000, 462575000, 462600000, 462625000, 462650000, 462675000, 462700000, 462725000}
	t18FRS    = append(append(append([]int64{}, t18FRS1...), t18FRS2...), t18FRS3...)
	t18FRS16  = []int64{462562500, 462587500, 462612500, 462637500, 462662500, 462625000, 462725000, 462687500, 462712500, 462550000, 462575000, 462600000, 462650000, 462675000, 462700000, 462725000}
	t18GMRS   = append(append([]int64{}, t18FRS...), t18FRS3...)
	t18MURS   = []int64{151820000, 151880000, 151940000, 154570000, 154600000}
	t18PMR    = []int64{446006250, 446018750, 446031250, 446043750, 446056250, 446068750, 446081250, 446093750, 446106250, 446118750, 446131250, 446143750, 446156250, 446168750, 446181250, 446193750}
	t18Format = t18FormatN(16)
)

func t18In(list []int64, f int64) bool {
	for _, x := range list {
		if x == f {
			return true
		}
	}
	return false
}

// t18FormatN: canali della memoria «MEM_FORMAT» di radtel_t18.py
func t18FormatN(n int) string { return strings.Replace(t18FormatTpl, "%d", fmt.Sprint(n), 1) }

// memoria «MEM_FORMAT_RB18» (canali da 0x0000)
const rb18FormatTpl = `
#seekto 0x0000;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 jumpcode:1,
       unknown1:2,
       skip:1,
       highpower:1,
       narrow:1,
       unknown2:1,
       bcl:1;
    u8 unknown3[3];
} memory[%d];
`

func init() {
	pmr := func(h, l float64) []genPower {
		return []genPower{{fmt.Sprintf("Alta %g W", h), h, 0}, {fmt.Sprintf("Bassa %g W", l), l, 2}}
	}
	uhf := []band{{400000000, 470000000}}
	vhf := []band{{136000000, 174000000}}
	type v struct {
		id, vendor, model, rclass, name string
		slots                           int
		pw                              []genPower
		bands                           []band
		o                               t18Opt
	}
	spec := func(magic string, fingers [][]byte, block int, ack bool, exit string, echo bool, memsize int, ranges [][2]int) *rbSpec {
		return &rbSpec{exit: []byte(exit), quiet: true, magic: append([]byte{0x02}, magic...), fingers: fingers,
			memsize: memsize, ranges: ranges, block: block, tries: 1, to: time.Second, noAck: !ack, echo: echo}
	}
	all := [][2]int{{0x0000, 0x03F0}}
	smp := func(tail string) [][]byte { return [][]byte{[]byte("SMP558" + tail)} }
	six := [][]byte{[]byte("\x06\x00\x00\x00\x00\x00\x00\x00")}
	t18 := func(lock string, extras ...string) t18Opt {
		return t18Opt{power: "rw", reset: 0xF9, lock: lock, extras: append([]string{"bcl"}, extras...)}
	}
	sc := []string{"scramble", "compander", "speccode"}
	list := []v{
		{"bf-v8a", "Baofeng", "BF-V8A", "BFV8ARadio", "Baofeng BF-V8A", 16, pmr(2, 0.5), uhf,
			t18Opt{format: t18FormatN(16), spec: spec("PROGRAM", [][]byte{[]byte("P3107\xF7\x00\x00")}, 8, true, "b", false, 0x03F0, all),
				power: "rw", reset: 0x09, extras: []string{"bcl"}}},
		{"bf-t20frs", "Baofeng", "BF-T20FRS", "BFT20FRSRadio", "Baofeng BF-T20FRS", 22, pmr(2, 0.5), uhf,
			t18Opt{format: t20frsFormat, spec: spec("PROGRAM", [][]byte{[]byte("P3107\xF7\x00\x00")}, 8, true, "b", false, 0x03F0, [][2]int{{0x0000, 0x0160}, {0x02B0, 0x02D0}}),
				power: "rw", reset: 0x09, extras: []string{"bcl"}}},
		{"rt-t18", "Radtel", "T18", "T18Radio", "Radtel T18", 16, []genPower{{"Unica", 0.5, 0}}, uhf,
			t18Opt{spec: spec("1ROGRAM", smp("\x00\x00"), 8, true, "b", false, 0x03F0, all), reset: 0xF9, extras: append([]string{"bcl"}, sc...)}},
		// 4.7: Retevis RT18 (retevis_rt18.py): la potenza è nell'ordine di CHIRP (bit a 1 = bassa)
		{"rt-rt18", "Retevis", "RT18", "RT18Radio", "Retevis RT18", 16, []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}}, uhf,
			withSpec(t18("", sc...), rt18Spec(spec("JSOGRAP", [][]byte{{0x06}}, 8, false, "b", false, 0x03F0, all)))},
		{"rt-rt20", "Retevis", "RT20", "RT20Radio", "Retevis RT20", 16, pmr(2, 0.5), uhf,
			withSpec(t18("", sc...), spec("8AOGRAM", smp("\x02"), 8, true, "b", false, 0x03F0, all))},
		{"rt-rt22s", "Retevis", "RT22S", "RT22SRadio", "Retevis RT22S", 22, pmr(2, 0.5), uhf,
			withSpec(t18("frs", sc...), spec("9COGRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-rb18", "Retevis", "RB18", "RB18Radio", "Retevis RB18", 22, pmr(2, 0.5), uhf,
			t18Opt{format: strings.Replace(rb18FormatTpl, "%d", "22", 1), spec: spec("PROGRAL", [][]byte{[]byte("P3107\xF7")}, 16, true, "E", false, 0x0660, [][2]int{{0x0000, 0x0660}}),
				power: "r", reset: 0xF9, lock: "frs", extras: []string{"bcl"}}},
		{"rt-rb618", "Retevis", "RB618", "RB618Radio", "Retevis RB618", 16, []genPower{{"Unica 0,5 W", 0.5, 0}}, uhf,
			t18Opt{format: strings.Replace(rb18FormatTpl, "%d", "16", 1), spec: spec("PROGRAL", [][]byte{[]byte("P3107\xF7")}, 16, true, "E", false, 0x0660, [][2]int{{0x0000, 0x0660}}),
				power: "low", reset: 0xF9, lock: "pmr", extras: []string{"bcl"}}},
		{"rt-rt68", "Retevis", "RT68", "RT68Radio", "Retevis RT68", 16, pmr(2, 0.5), uhf, withSpec(t18("frs16", sc...), spec("83OGRAM", six, 8, false, "", false, 0x03F0, all))},
		{"rt-rt668", "Retevis", "RT668", "RT668Radio", "Retevis RT668", 16, pmr(2, 0.5), uhf, withSpec(t18("pmr", sc...), spec("83OGRAM", six, 8, false, "", false, 0x03F0, all))},
		{"rt-rb17", "Retevis", "RB17", "RB17Radio", "Retevis RB17", 16, pmr(2, 0.5), uhf,
			withSpec(t18("frs16", "scramble", "compander"), spec("A5OGRAM", [][]byte{[]byte("S\x00\x00\x00\x00\x00\x00\x00")}, 8, false, "", false, 0x03F0, all))},
		{"rt-rb617", "Retevis", "RB617", "RB617Radio", "Retevis RB617", 16, pmr(2, 0.5), uhf,
			withSpec(t18("pmr", "scramble", "compander"), spec("A5OGRAM", [][]byte{[]byte("S\x00\x00\x00\x00\x00\x00\x00")}, 8, false, "", false, 0x03F0, all))},
		{"rt-rb17v", "Retevis", "RB17V", "RB17VRadio", "Retevis RB17V (MURS)", 5, pmr(2, 0.5), vhf,
			withSpec(t18("murs", "scramble", "compander", "speccode"), spec("A5OGRAM", [][]byte{[]byte("S\x00\x00\x00\x00\x00\x00\x00")}, 8, false, "", false, 0x03F0, all))},
		{"rt-rb85", "Retevis", "RB85", "RB85Radio", "Retevis RB85", 16, pmr(10, 5), uhf, withSpec(t18("", sc...), spec("H19GRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-rb75", "Retevis", "RB75", "RB75Radio", "Retevis RB75", 30, pmr(5, 0.5), uhf,
			withSpec(t18("", "scramble", "compander"), spec("KVOGRAM", smp("\x00"), 8, false, "b", false, 0x03F0, all))},
		{"bt-frsb1", "BTECH", "FRS-B1", "FRSB1Radio", "BTECH FRS-B1", 22, pmr(2, 0.5), uhf,
			withSpec(t18("frs"), spec("PROGRAM", [][]byte{[]byte("P3107\xF7\x00")}, 8, true, "b", false, 0x03F0, all))},
		{"rt-rb19", "Retevis", "RB19", "RB19Radio", "Retevis RB19", 22, pmr(2, 0.5), uhf, withSpec(t18("frs", sc...), spec("9COGRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-rb19p", "Retevis", "RB19P", "RB19PRadio", "Retevis RB19P", 30, pmr(3, 0.5), uhf,
			withSpec(t18("gmrs", "scramble", "compander"), spec("70OGRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-rb619", "Retevis", "RB619", "RB619Radio", "Retevis RB619", 16, pmr(0.5, 0.499), uhf, withSpec(t18("pmr", sc...), spec("9COGRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-rt47", "Retevis", "RT47", "RT47Radio", "Retevis RT47", 16, pmr(2, 0.5), uhf, withSpec(t18("frs16", sc...), spec("47OGRAM", six, 8, false, "b", true, 0x03F0, all))},
		{"rt-rt47v", "Retevis", "RT47V", "RT47VRadio", "Retevis RT47V (MURS)", 5, pmr(2, 0.5), vhf, withSpec(t18("murs", "compander"), spec("47OGRAM", six, 8, false, "b", true, 0x03F0, all))},
		{"rt-rt647", "Retevis", "RT647", "RT647Radio", "Retevis RT647", 16, []genPower{{"Unica 0,5 W", 0.5, 0}}, uhf, withSpec(t18Pow(t18("pmr", sc...), "keep"), spec("47OGRAM", six, 8, false, "b", true, 0x03F0, all))},
		{"rt-rb29", "Retevis", "RB29", "RB29Radio", "Retevis RB29", 16, pmr(2, 0.5), uhf,
			t18Opt{spec: spec("S19GRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all), power: "rw", reset: 0xF7, lock: "frs16", extras: []string{"bcl"}}},
		{"rt-rb629", "Retevis", "RB629", "RB629Radio", "Retevis RB629", 16, pmr(0.5, 0.499), uhf,
			t18Opt{spec: spec("S19GRAM", smp("\x02"), 8, false, "b", false, 0x03F0, all), power: "rw", reset: 0xF7, lock: "pmr", extras: []string{"bcl"}}},
		{"rt-rt15", "Retevis", "RT15", "RT15Radio", "Retevis RT15", 16, pmr(2, 0.5), uhf,
			withSpec(t18("", sc...), spec("KAOGRAM", [][]byte{[]byte("\x06\x00\x00\x00\x00\x00\x00\x00"), []byte("\x06\x03\xe8\x08\xff\xff\xff\xff"), []byte("\x06MP558\xff\xff")}, 8, false, "b", false, 0x03F0, all))},
		{"rt-rb87", "Retevis", "RB87", "RB87Radio", "Retevis RB87", 30, pmr(5, 0.5), uhf,
			withSpec(t18("gmrs", "compander", "speccode"), spec("C8OGRAN", smp(""), 8, false, "", false, 0x03F0, all))},
		{"rt-h777h-frs", "Retevis", "H777H_FRS", "H777HFRSRadio", "Retevis H777H (FRS)", 16, pmr(2, 0.5), uhf,
			withSpec(t18Res(t18("frs16", sc...)), spec("C701RAD", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
		{"rt-h777h-pmr", "Retevis", "H777H_PMR", "H777HPMRRadio", "Retevis H777H (PMR446)", 16, pmr(0.5, 0.5), uhf,
			withSpec(t18Pow(t18Res(t18("pmr", sc...)), "high"), spec("C701RAD", smp("\x02"), 8, false, "b", false, 0x03F0, all))},
	}
	for _, x := range list {
		o := x.o
		if o.format == "" {
			o.format = t18FormatN(x.slots)
		}
		m := &genModel{ID: x.id, Vendor: x.vendor, Model: x.model, RClass: x.rclass, Name: x.name, Slots: x.slots, First: 1,
			Powers: x.pw, Bands: x.bands, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: t18Fam{}, Opt: &o, Notes: "Radio senza nomi dei canali."}
		if x.rclass == "RT647Radio" {
			o.spec.enterAck = 0xF0
		}
		genRegister(m)
	}
}

// t18Fixed: campi del canale che CHIRP non lascia modificare (immutable),
// per i canali che contengono una frequenza fissa FRS, PMR446, MURS o GMRS
func t18Fixed(o *t18Opt, n int, rx int64) (fixed, dup, mode, power bool) {
	switch o.lock {
	case "frs":
		if t18In(t18FRS, rx) {
			return true, true, true, n >= 8 && n <= 14
		}
	case "frs16":
		if t18In(t18FRS16, rx) {
			return true, true, true, false
		}
	case "murs":
		if t18In(t18MURS, rx) {
			return true, true, n <= 3, false
		}
	case "pmr":
		if t18In(t18PMR, rx) {
			return true, true, true, true
		}
	case "gmrs":
		if t18In(t18GMRS, rx) {
			return n <= 30, n <= 30, n >= 8 && n <= 14, n >= 8 && n <= 14
		}
	}
	return false, false, false, false
}

// lockedSlotImg: canale a frequenza fissa (escluso dalle prove di scrittura)
func (f t18Fam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	c := f.Get(m, img, i)
	if c == nil {
		return false
	}
	fx, _, _, _ := t18Fixed(t18o(m), i+1, c.Rx)
	return fx
}

// Locked: i canali con una frequenza fissa (FRS, PMR446, MURS, GMRS) non si
// cancellano e non cambiano frequenza, come in CHIRP
func (f t18Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if want != nil {
		// frequenza di un canale fisso nella posizione sbagliata: la radio (come
		// CHIRP) la sostituirebbe con quella della posizione
		if fx, _, _, _ := t18Fixed(t18o(m), i+1, want.Rx); fx {
			buf := bytes.Repeat([]byte{0xFF}, genMinSize(m))
			if err := f.Set(m, buf, i, want); err == nil {
				if back := f.Get(m, buf, i); back == nil || back.Rx != want.Rx || (!want.TxOff && back.Tx != want.Tx) {
					return fmt.Errorf("nella posizione %d della %s la frequenza %s MHz non è ammessa (canali fissi)", i+1, m.Name, fmtMHz(want.Rx))
				}
			}
		}
	}
	if old == nil {
		return nil
	}
	fx, dup, mode, power := t18Fixed(t18o(m), i+1, old.Rx)
	if !fx {
		return nil
	}
	what := map[string]string{"frs": "FRS", "frs16": "FRS", "murs": "MURS", "pmr": "PMR446", "gmrs": "GMRS"}[t18o(m).lock]
	switch {
	case want == nil:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: non si può cancellare", i+1, m.Name, what)
	case want.Rx != old.Rx:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza non modificabile", i+1, m.Name, what)
	case dup && (want.Tx != old.Tx || want.TxOff != old.TxOff):
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza di trasmissione non modificabile", i+1, m.Name, what)
	case mode && want.Narrow != old.Narrow:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: sempre a banda stretta", i+1, m.Name, what)
	case power && want.Power != old.Power:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: potenza non modificabile", i+1, m.Name, what)
	}
	return nil
}

func withSpec(o t18Opt, sp *rbSpec) t18Opt { o.spec = sp; return o }
func t18Pow(o t18Opt, p string) t18Opt     { o.power = p; return o }

// powerFixed: la potenza scelta non viene scritta così com'è (vedi t18Opt.power)
func (f t18Fam) powerFixed(m *genModel) bool { return t18o(m).power != "rw" }
func t18Res(o t18Opt) t18Opt                 { o.reserved = true; return o }

func t18o(m *genModel) *t18Opt { return m.Opt.(*t18Opt) }

func (t18Fam) mem(m *genModel) *bwEl          { return bwMust(t18o(m).format) }
func (f t18Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f t18Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f t18Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f t18Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f t18Fam) Match(m *genModel, img []byte) bool                { return false }

func t18ToneDec(raw []byte) string {
	v := bcdByte(raw[0]) + 100*bcdByte(raw[1])
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

func t18ToneEnc(dst []byte, t genTone) {
	switch t.Mode {
	case "":
		dst[0], dst[1] = 0xFF, 0xFF
	case "Tone":
		dst[0], dst[1] = decToBCD(int64(t.Hz10)%100), decToBCD(int64(t.Hz10)/100%100)
	case "DTCS":
		dst[0], dst[1] = decToBCD(int64(t.Code)%100), decToBCD(int64(t.Code)/100%100)
		if t.Inv {
			dst[1] |= 0xC0
		} else {
			dst[1] |= 0x80
		}
	}
}

func (f t18Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := t18o(m)
	r := f.rec(m, i)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = r.F("narrow").Bool(img)
	c.Scan = !r.F("skip").Bool(img)
	c.TxTone = t18ToneDec(r.F("txtone").Raw(img))
	c.RxTone = t18ToneDec(r.F("rxtone").Raw(img))
	if o.power == "rw" || o.power == "r" || o.power == "high" {
		c.Power = m.powerSem(int(1 - r.F("highpower").Int(img)))
	}
	// canali a frequenza fissa: come li mostra CHIRP (get_memory)
	n := i + 1
	simplex := func() { c.Tx, c.TxOff = c.Rx, false }
	low := func() {
		if len(m.Powers) > 1 {
			c.Power = m.Powers[1].Sem
		}
	}
	switch o.lock {
	case "frs":
		if t18In(t18FRS, rx) {
			if n <= 22 {
				c.Rx = t18FRS[n-1]
			}
			simplex()
			c.Narrow = true
			if n >= 8 && n <= 14 {
				low()
			}
		}
	case "frs16":
		if t18In(t18FRS16, rx) {
			if n <= 16 {
				c.Rx = t18FRS16[n-1]
			}
			simplex()
			c.Narrow = true
		}
	case "murs":
		if t18In(t18MURS, rx) {
			if n <= 5 {
				c.Rx = t18MURS[n-1]
			}
			simplex()
			if n <= 3 {
				c.Narrow = true
			}
		}
	case "pmr":
		if t18In(t18PMR, rx) {
			if n <= 16 {
				c.Rx = t18PMR[n-1]
			}
			simplex()
			c.Narrow = true
			low()
		}
	case "gmrs":
		if t18In(t18GMRS, rx) {
			if n <= 30 {
				c.Rx = t18GMRS[n-1]
			}
			switch {
			case n <= 7, n >= 15 && n <= 22:
				simplex()
			case n >= 8 && n <= 14:
				simplex()
				c.Narrow = true
				low()
			case n >= 23 && n <= 30:
				c.Tx, c.TxOff = c.Rx+5000000, false
			}
		}
	}
	return c
}

func (f t18Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := t18o(m)
	r := f.rec(m, i)
	raw := r.Raw(img)
	rsvd := append([]byte{}, raw[13:16]...)
	if c == nil {
		r.Fill(img, 0xFF)
		if o.reserved {
			copy(r.Raw(img)[13:], rsvd)
		}
		return nil
	}
	wasEmpty := f.Get(m, img, i) == nil
	oldHigh := r.F("highpower").Int(img)
	keep := map[string]int64{}
	for _, k := range o.extras {
		keep[k] = r.F(k).Int(img)
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	if o.reserved {
		r.SetRaw(img, append(make([]byte, 13), rsvd...))
	} else {
		r.SetRaw(img, append(make([]byte, 12), o.reset, 0xFF, 0xFF, 0xFF))
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	t18ToneEnc(r.F("txtone").Raw(img), tt)
	t18ToneEnc(r.F("rxtone").Raw(img), rt)
	switch o.power {
	case "rw":
		r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 0)
	case "low":
		r.F("highpower").SetInt(img, 0)
	case "high":
		// H777H PMR: le due potenze di CHIRP valgono entrambe 0,5 W e CHIRP
		// scrive sempre «alta»
		r.F("highpower").SetInt(img, 1)
	case "keep":
		if wasEmpty {
			oldHigh = 0
		}
		r.F("highpower").SetInt(img, oldHigh)
	}
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("skip").SetBool(img, !c.Scan)
	r.F("unknown2").SetInt(img, 1)
	if !wasEmpty {
		// impostazioni del canale conservate (come CHIRP)
		for k, v := range keep {
			r.F(k).SetInt(img, v)
		}
	}
	return nil
}

func (f t18Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	d, err := rbDownload(ctx, port, 9600, t18o(m).spec, prog)
	return d, "", err
}

func (f t18Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", rbUpload(ctx, port, 9600, t18o(m).spec, img, prog)
}

// ---------------------------------------------------------------- BF-T20, W31E (driver retevis_rt22.py)

const rt22Format = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1;
  u8 unknown3:2,
     highpower:1,
     wide:1,
     unknown4:2,
     signal:1,
     bcl:1;
  u8 unknown5[2];
} memory[16];

#seekto 0x017E;
u8 skipflags[2];
`

type rt22Opt struct {
	fileid  [][]byte
	memsize int
	ranges  [][3]int // inizio, fine, blocco
	anyID   bool     // identificativo non controllato (come il programma del costruttore)
	cpsFill bool     // byte non usati del canale a FF, come il programma del costruttore
}

type rt22Fam struct{}

var rt22DefaultIDs = [][]byte{[]byte("P32073"), []byte("P3\x00\x00\x003"), []byte("P3207!"), []byte("\x00\x00\x00\x00\x00\x00\xF8\xFF")}

func init() {
	pw := []genPower{{"Bassa 2 W", 2, 2}, {"Alta 5 W", 5, 0}}
	mk := func(id, model, rclass, name string, o *rt22Opt) *genModel {
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: 16, First: 1,
			Powers: pw, Bands: []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600,
			Fam: rt22Fam{}, Opt: o, Notes: "Radio senza nomi dei canali."}
	}
	genRegister(
		mk("bf-t20", "BF-T20", "BFT20", "Baofeng BF-T20", &rt22Opt{fileid: [][]byte{[]byte("P330h33"), []byte("P32073\xF8\xFF")}, memsize: 0x0400,
			ranges: [][3]int{{0x0000, 0x0180, 0x10}, {0x01B8, 0x01F8, 0x10}, {0x01F8, 0x0200, 0x08}, {0x0200, 0x0340, 0x10}}}),
		mk("bf-w31e", "W31E", "W31E", "Baofeng W31E", &rt22Opt{fileid: rt22DefaultIDs, memsize: 0x0200,
			ranges: [][3]int{{0x0000, 0x0200, 0x10}}}),
	)
	// Baofeng T22 (PMR446): programma del costruttore «Baofeng_T22_Programming_Software_PMR»,
	// stesso protocollo («PROGRGS») e stessa memoria della W31E (retevis_rt22 di CHIRP);
	// il programma non controlla l'identificativo e scrive FF nei byte non usati.
	t22 := mk("bf-t22", "T22", "W31E", "Baofeng T22 (PMR446)", &rt22Opt{memsize: 0x0200,
		ranges: [][3]int{{0x0000, 0x0200, 0x10}}, anyID: true, cpsFill: true})
	t22.NoChirp = true
	t22.Powers = []genPower{{"Bassa", 0.5, 2}, {"Alta", 0.5, 0}}
	t22.Bands = []band{{446000000, 446200000}}
	t22.Notes = "Radio PMR446 senza nomi dei canali. Ricavato dal programma del costruttore (versione PMR), " +
		"stessa memoria della W31E di CHIRP: nei file .img compare come W31E."
	genRegister(t22)

	// 4.0: gli altri modelli del driver retevis_rt22.py di CHIRP (stessa memoria e
	// stesso protocollo della Retevis RT22, come la BF-T20)
	rtr := [][3]int{{0x0000, 0x0180, 0x10}, {0x01B8, 0x01F8, 0x10}, {0x01F8, 0x0200, 0x08}, {0x0200, 0x0340, 0x10}}
	for _, v := range []struct{ id, vendor, model, rclass, name string }{
		{"rt-rt22", "Retevis", "RT22", "RT22Radio", "Retevis RT22"},
		{"rt-rt22frs", "Retevis", "RT22FRS", "RT22FRS", "Retevis RT22FRS"},
		{"rt-rt622", "Retevis", "RT622", "RT622", "Retevis RT622"},
		{"wln-kdc1", "WLN", "KD-C1", "KDC1", "WLN KD-C1"},
		{"zt-ztx6", "Zastone", "ZT-X6", "ZTX6", "Zastone ZT-X6"},
		{"lt-lt316", "LUITON", "LT-316", "LT316", "Luiton LT-316"},
		{"tid-tdm8", "TID", "TD-M8", "TDM8", "TID TD-M8"},
	} {
		ids := rt22DefaultIDs
		if v.rclass == "RT622" {
			ids = append(append([][]byte{}, rt22DefaultIDs...), []byte("\xFF\xFF\xF8\xFF"))
		}
		x := mk(v.id, v.model, v.rclass, v.name, &rt22Opt{fileid: ids, memsize: 0x0400, ranges: rtr})
		x.Vendor = v.vendor
		genRegister(x)
	}
}

func rt22o(m *genModel) *rt22Opt { return m.Opt.(*rt22Opt) }

func (rt22Fam) mem() *bwEl                 { return bwMust(rt22Format) }
func (f rt22Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f rt22Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f rt22Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}
func (f rt22Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(i).SetRaw(img, raw) }
func (f rt22Fam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x0408 && bytes.HasPrefix(img[0x400:0x408], []byte(m.Model))
}

// toni come il modello "Kenwood" di CHIRP: DCS 0x2800 + codice ottale, 0x8000 invertito
func kenwoodToneDec(v int64, dcsBase, polMask int64) string {
	if v == 0xFFFF || v == 0 {
		return ""
	}
	if v&dcsBase == dcsBase {
		code, _ := strconv.Atoi(fmt.Sprintf("%03o", v&0x07FF))
		return genTone{Mode: "DTCS", Code: code, Inv: v&polMask != 0}.String()
	}
	return fmtTone(int(v & 0x7FFF))
}

func kenwoodToneEnc(t genTone, dcsBase, polMask, init int64) int64 {
	switch t.Mode {
	case "DTCS":
		v, _ := strconv.ParseInt(fmt.Sprintf("%d", t.Code), 8, 64)
		v |= dcsBase
		if t.Inv {
			v += polMask
		}
		return v
	case "Tone":
		return int64(t.Hz10)
	}
	return init
}

func (f rt22Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	c := &Chan{Rx: rx}
	switch {
	case r.F("rxfreq").Int(img) == r.F("txfreq").Int(img):
		c.Tx = rx
	case bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}):
		c.TxOff, c.Tx = true, rx
	default:
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Narrow = !r.F("wide").Bool(img)
	c.TxTone = kenwoodToneDec(r.F("txtone").Int(img), 0x2800, 0x8000)
	c.RxTone = kenwoodToneDec(r.F("rxtone").Int(img), 0x2800, 0x8000)
	c.Power = m.powerSem(int(r.F("highpower").Int(img)))
	sk := f.mem().F("skipflags").I(i / 8)
	c.Scan = sk.Int(img)&(1<<uint(i%8)) != 0
	return c
}

func (f rt22Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF)
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
	r.Fill(img, 0)
	if rt22o(m).cpsFill {
		r.F("unknown1").SetInt(img, 0xFF)
		r.F("unknown5").Fill(img, 0xFF)
	} else {
		r.F("unknown5").I(0).SetInt(img, 0x80)
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("rxtone").SetInt(img, kenwoodToneEnc(rt, 0x2800, 0x8000, 0xFFFF))
	r.F("txtone").SetInt(img, kenwoodToneEnc(tt, 0x2800, 0x8000, 0xFFFF))
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
	sk := f.mem().F("skipflags").I(i / 8)
	v := sk.Int(img)
	if c.Scan {
		v |= 1 << uint(i%8)
	} else {
		v &^= 1 << uint(i%8)
	}
	sk.SetInt(img, v)
	return nil
}

// invio byte per byte con 5 ms tra l'uno e l'altro (come CHIRP)
func (s *rawSess) slowSend(b []byte) error {
	for _, c := range b {
		if err := s.sleep(5 * time.Millisecond); err != nil {
			return err
		}
		if err := s.send([]byte{c}); err != nil {
			return err
		}
	}
	return nil
}

func (s *rawSess) rt22Enter(o *rt22Opt) ([]byte, error) {
	var ack []byte
	ok := false
	for k := 0; k < 5 && !ok; k++ {
		if err := s.slowSend([]byte("PROGRGS")); err != nil {
			return nil, err
		}
		ack = readUpTo(s.ctx, s.p, 1, time.Second)
		ok = len(ack) == 1 && ack[0] == 0x06
	}
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if !ok {
		if len(ack) == 0 {
			return nil, rErr("la radio non risponde")
		}
		return nil, rErr("la radio non è entrata in modalità di programmazione: controlla il cavo e riaccendi la radio")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return nil, err
	}
	id := readUpTo(s.ctx, s.p, 8, time.Second)
	found := o.anyID && len(id) == 8
	for _, fp := range o.fileid {
		if bytes.Contains(id, fp) {
			found = true
		}
	}
	if !found {
		return id, rErr("identificazione della radio non riuscita (%q)", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return id, err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		_ = s.send([]byte{'E'})
		return id, rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if err := s.send([]byte{0x07}); err != nil {
		return id, err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x4E {
		_ = s.send([]byte{'E'})
		return id, rErr("la radio ha rifiutato la modalità di programmazione")
	}
	return id, nil
}

func (f rt22Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rt22o(m)
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	id, err := s.rt22Enter(o)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < o.memsize; a += 0x40 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, o.memsize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), 0x40}
		if err := s.slowSend(cmd); err != nil {
			return nil, "", err
		}
		resp := readUpTo(ctx, s.p, 4+0x40, time.Second)
		if len(resp) != 0x44 || !bytes.Equal(resp[:4], append([]byte{'W'}, cmd[1:]...)) {
			_ = s.send([]byte{'E'})
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if err := s.sleep(5 * time.Millisecond); err != nil {
			return nil, "", err
		}
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
			_ = s.send([]byte{'E'})
			return nil, "", rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	_ = s.send([]byte{'E'})
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, fmt.Sprintf("identificativo %q", id), nil
}

func (f rt22Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rt22o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := rawOpen(ctx, port, 9600)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	id, err := s.rt22Enter(o)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	patch := bytes.Equal(img[0x1B8:0x1C0], []byte{0, 0, 0, 0, 0, 0, 0xFF, 0xFF})
	for _, r := range o.ranges {
		for a := r[0]; a < r[1]; a += r[2] {
			prog(a, o.memsize, "Scrittura…")
			data := append([]byte{}, img[a:a+r[2]]...)
			if patch && a == 0x1B8 {
				// come CHIRP (RT-622, RT22FRS): l'identificativo letto dalla radio al posto di zeri
				for _, b := range id {
					if bytes.IndexByte(img[0:len(id)], b) >= 0 {
						data = append(append([]byte{}, img[0:len(id)]...), data[len(id):]...)
					}
				}
			}
			if err := s.slowSend([]byte{'W', byte(a >> 8), byte(a), byte(r[2])}); err != nil {
				return "", err
			}
			if err := s.slowSend(data); err != nil {
				return "", err
			}
			if b := readUpTo(ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
				_ = s.send([]byte{'E'})
				if ctx.Err() != nil {
					return "", errCancelled
				}
				return "", rErr("la radio ha rifiutato il blocco 0x%04X", a)
			}
		}
	}
	_ = s.send([]byte{'E'})
	prog(o.memsize, o.memsize, "Scrittura completata")
	return "", nil
}
