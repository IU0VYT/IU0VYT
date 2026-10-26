// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: uvb5.py, bf_t1.py, bf_t8.py, baofeng_uv3r.py, radtel_t18.py, retevis_rt22.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Radio Baofeng "piccole" con protocolli propri: UV-B5/UV-B6 (uvb5.py di
// CHIRP) e BF-T1 (bf_t1.py). Protocolli e mappe come in CHIRP.

import (
	"bytes"
	"context"
	"fmt"
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
	p, err := openPortFunc(port, baud)
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

const uvb5Charset = "0123456789- ABCDEFGHIJKLMNOPQRSTUVWXYZ/_+*"

type uvb5Fam struct{}

func init() {
	genRegister(&genModel{ID: "bf-uvb5", Vendor: "Baofeng", Model: "UV-B5", RClass: "BaofengUVB5", Name: "Baofeng UV-B5 / UV-B6",
		Slots: 99, First: 1, NameMax: 5, Charset: uvb5Charset, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{130000000, 175000000}, {220000000, 269000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: uvb5Fam{},
		Notes: "Accendi la radio su un canale libero."})
}

func (uvb5Fam) mem() *bwEl                 { return bwMust(uvb5Format) }
func (f uvb5Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f uvb5Fam) recs(i int) (*bwEl, *bwEl) {
	r := f.mem()
	return r.F("channels").I(i), r.F("names").I(i).F("name")
}
func (f uvb5Fam) Match(m *genModel, img []byte) bool {
	return bytes.HasPrefix(img, []byte("KT511 Radio Program data")) && len(img) == 0x1030
}

func (f uvb5Fam) Rec(m *genModel, img []byte, i int) []byte {
	a, b := f.recs(i)
	return append(append([]byte{}, a.Raw(img)...), b.Raw(img)...)
}

func (f uvb5Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	a, b := f.recs(i)
	a.SetRaw(img, raw[:16])
	b.SetRaw(img, raw[16:])
}

func uvb5ToneDec(v int64, pol bool) string {
	switch {
	case v > 155 || v == 0:
		return ""
	case v > 50:
		return genTone{Mode: "DTCS", Code: dtcsCodes[v-51], Inv: pol}.String()
	}
	return fmtTone(chirpTones[v-1])
}

func (f uvb5Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, nam := f.recs(i)
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
	c.TxTone = uvb5ToneDec(r.F("txtone").Int(img), r.F("txpol").Bool(img))
	c.RxTone = uvb5ToneDec(r.F("rxtone").Int(img), r.F("rxpol").Bool(img))
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
	r, nam := f.recs(i)
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
	for k := 0; k < 5; k++ {
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

func (s *rawSess) uvb5Ident() error {
	if err := s.send([]byte("\x05PROGRAM")); err != nil {
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
	if err := s.uvb5Ident(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := append([]byte("KT511 Radio Program data v1.08\x00\x00"), make([]byte, 16)...)
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
	if err := s.uvb5Ident(); err != nil {
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
}

func bft8o(m *genModel) *bft8Opt {
	if o, ok := m.Opt.(*bft8Opt); ok {
		return o
	}
	return &bft8Opt{spec: bft8Spec, dcs: dtcsCodes}
}

func (bft8Fam) mem() *bwEl { return bwMust(bft8Format) }
func (f bft8Fam) layout(m *genModel) *bwEl {
	if bft8o(m).names {
		return bwMust(bft8Format + bft8NamesFormat)
	}
	return f.mem()
}
func (f bft8Fam) rec(i int) *bwEl { return f.mem().F("memory").I(i) }
func (f bft8Fam) name(i int) *bwEl {
	return bwMust(bft8Format + bft8NamesFormat).F("names").I(i).F("name")
}
func (f bft8Fam) Rec(m *genModel, img []byte, i int) []byte {
	r := append([]byte{}, f.rec(i).Raw(img)...)
	if bft8o(m).names {
		r = append(r, f.name(i).Raw(img)...)
	}
	return r
}
func (f bft8Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	f.rec(i).SetRaw(img, raw[:16])
	if bft8o(m).names && len(raw) >= 22 {
		f.name(i).SetRaw(img, raw[16:22])
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
	r := f.rec(i)
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
	if o.names {
		// come CHIRP: 0xFF (anche in mezzo al nome) = spazio, spazi finali tolti
		nb := []byte{}
		for _, b := range f.name(i).Raw(img) {
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
	r := f.rec(i)
	o := bft8o(m)
	if c == nil {
		r.Fill(img, 0xFF)
		if o.names {
			f.name(i).Fill(img, 0xFF)
		}
		return nil
	}
	r.SetRaw(img, append(make([]byte, 13), 0xFF, 0xFF, 0xFF))
	if o.names {
		nb := bytes.Repeat([]byte{0xFF}, 6)
		copy(nb, c.Name)
		f.name(i).SetRaw(img, nb)
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

// protocollo comune a BF-T8 e simili: comando, identificativo, blocchi 'R'/'W' con conferma, 'E'
type rbSpec struct {
	exit          []byte // comando di uscita ("E"; "b" sulle radio tipo T18)
	quiet         bool   // in caso di errore di identificazione non invia il comando di uscita
	magic, finger []byte
	memsize       int
	ranges        [][2]int
	block         int
	tries         int
	to            time.Duration
}

func (s *rawSess) rbEnter(sp *rbSpec) error {
	ok := false
	got := false
	for k := 0; k < sp.tries && !ok; k++ {
		if err := s.send(sp.magic); err != nil {
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
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, len(sp.finger), sp.to)
	if !bytes.Equal(id, sp.finger) {
		if !sp.quiet {
			_ = s.send(sp.exitCmd())
		}
		return rErr("la radio ha restituito un identificativo sconosciuto (%q)", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, sp.to)
	if len(b) != 1 || b[0] != 0x06 {
		if !sp.quiet {
			_ = s.send(sp.exitCmd())
		}
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	return nil
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
		if err := s.send(cmd); err != nil {
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
		if err := s.send([]byte{0x06}); err != nil {
			return nil, err
		}
		ack := readUpTo(ctx, s.p, 1, sp.to)
		if len(ack) != 1 || ack[0] != 0x06 {
			if !sp.quiet {
				_ = s.send(sp.exitCmd())
			}
			return nil, rErr("nessuna conferma leggendo il blocco 0x%04X", a)
		}
		data = append(data, resp[4:]...)
	}
	if err := s.send(sp.exitCmd()); err != nil {
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
	for _, r := range sp.ranges {
		for a := r[0]; a < r[1]; a += sp.block {
			prog(a, sp.memsize, "Scrittura…")
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), byte(sp.block)}, img[a:a+sp.block]...)); err != nil {
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
	if err := s.send(sp.exitCmd()); err != nil {
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
		DCS: dtcsCodes, Baud: 9600, Fam: uv3rFam{}, Notes: "Radio senza nomi dei canali."})
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

const t18Format = `
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
} memory[16];
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
}

type t18Fam struct{}

func init() {
	pw := []genPower{{"Alta 2 W", 2, 0}, {"Bassa 0,5 W", 0.5, 2}}
	mk := func(id, model, rclass, name string, slots int, o *t18Opt) *genModel {
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: slots, First: 1,
			Powers: pw, Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600,
			Fam: t18Fam{}, Opt: o, Notes: "Radio senza nomi dei canali."}
	}
	spec := func(ranges [][2]int) *rbSpec {
		return &rbSpec{exit: []byte("b"), quiet: true, magic: []byte("\x02PROGRAM"), finger: []byte("P3107\xF7\x00\x00"),
			memsize: 0x03F0, ranges: ranges, block: 8, tries: 1, to: time.Second}
	}
	genRegister(
		mk("bf-v8a", "BF-V8A", "BFV8ARadio", "Baofeng BF-V8A", 16, &t18Opt{format: t18Format, spec: spec([][2]int{{0x0000, 0x03F0}})}),
		mk("bf-t20frs", "BF-T20FRS", "BFT20FRSRadio", "Baofeng BF-T20FRS", 22, &t18Opt{format: t20frsFormat, spec: spec([][2]int{{0x0000, 0x0160}, {0x02B0, 0x02D0}})}),
	)
}

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
	c.Power = m.powerSem(int(1 - r.F("highpower").Int(img)))
	return c
}

func (f t18Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	wasEmpty := f.Get(m, img, i) == nil
	bcl := r.F("bcl").Int(img)
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	r.SetRaw(img, append(make([]byte, 12), 0x09, 0xFF, 0xFF, 0xFF))
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	t18ToneEnc(r.F("txtone").Raw(img), tt)
	t18ToneEnc(r.F("rxtone").Raw(img), rt)
	r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 0)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("skip").SetBool(img, !c.Scan)
	r.F("unknown2").SetInt(img, 1)
	if !wasEmpty {
		// impostazione del canale conservata (come CHIRP)
		r.F("bcl").SetInt(img, bcl)
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
