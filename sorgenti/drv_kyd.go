// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: kyd.py e kyd_IP620.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: KYD NC-630A (Plant-Tours MT-700) e KYD IP-620 (Argut A-36)

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const nc630aFormat = `
#seekto 0x0010;
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:3,
     bcl:2,
     unknown2:3;
  u8 unknown3:2,
     highpower:1,
     wide:1,
     unknown4:4;
  u8 unknown5[2];
} memory[16];

#seekto 0x012F;
struct {
  u8 voice;
  u8 tot;
  u8 totalert;
  u8 unknown1[2];
  u8 squelch;
  u8 save;
  u8 beep;
  u8 unknown2[3];
  u8 vox;
  u8 voxdelay;
} settings;

#seekto 0x017E;
u8 skipflags[2];
`

const ip620Format = `
#seekto 0x0000;
struct {
  lbcd rx_freq[4];
  lbcd tx_freq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown_1:4,
     busy_loc:2,
     n_a:2;
  u8 unknown_2:1,
     scan_add:1,
     n_a2:1,
     w_n:1,
     lout:1,
     n_a_:1,
     power:2;
  u8 unknown_3;
  u8 unknown_4;
} memory[200];
#seekto 0x1000;
struct {
  u8 chan_name[6];
  u8 unknown_1[10];
} chan_names[200];
#seekto 0x0C80;
struct {
  lbcd freq_a_rx[4];
  lbcd freq_a_tx[4];
  ul16 freq_a_rx_tone;
  ul16 freq_a_tx_tone;
  u8 unknown_1_5:4,
  freq_a_busy_loc:2,
  n_a:2;
  u8 unknown_1_6:3,
  freq_a_w_n:1,
  n_a2:1,
  na:1,
  freq_a_power:2;
  u8 unknown_1_7;
  u8 unknown_1_8;
} settings_freq_a;
#seekto 0x0E20;
struct {
  u8 chan_disp_way;
  u8 step_freq;
  u8 rf_sql;
  u8 bat_save;
  u8 chan_pri;
  u8 end_beep;
  u8 tot;
  u8 vox;
  u8 chan_pri_num;
  u8 n_a_2;
  u8 ch_mode;
  u8 n_a_3;
  u8 call_tone;
  u8 beep;
  u8 unknown_1_1[2];
  u8 unknown_1_2[8];
  u8 scan_rev;
  u8 unknown_1_3[2];
  u8 enc;
  u8 vox_dly;
  u8 wait_back_light;
  u8 unknown_1_4[2];
} settings;
#seekto 0x0E40;
struct {
  u8 fm_radio;
  u8 auto_lock;
  u8 unknown_1[8];
  u8 pon_msg[6];
} settings_misc;
#seekto 0x1C80;
struct {
  u8 unknown_1[16];
  u8 unknown_2[16];
} settings_radio_3;
`

type kydFam struct{}

type kydOpt struct {
	ip620 bool
}

func init() {
	genRegister(&genModel{ID: "kyd-nc630a", Vendor: "KYD", Model: "NC-630A", RClass: "NC630aRadio", Name: "KYD NC-630A (Plant-Tours MT-700)",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}},
		Bands: []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: kydFam{}, Opt: &kydOpt{},
		Notes: "Accendi la radio con il cavo collegato. 16 canali UHF senza nomi. Driver di CHIRP."})
	genRegister(&genModel{ID: "kyd-ip620", Vendor: "KYD", Model: "IP-620", RClass: "IP620Radio", Name: "KYD IP-620 (Argut A-36)",
		Slots: 200, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 2,5 W", 2.5, 1}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{136000000, 174000000}, {200000000, 260000000}, {300000000, 340000000}, {350000000, 390000000}, {400000000, 520000000}},
		Narrow: true, HasScan: false, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: kydFam{}, Opt: &kydOpt{ip620: true},
		Notes: "Accendi la radio con il cavo collegato. 200 canali senza nomi (CHIRP non li gestisce); blocco del canale occupato, " +
			"esclusione e scansione si conservano. Il driver di CHIRP è sperimentale e scrive la potenza alta come bassa: " +
			"qui la potenza si scrive com'è nella memoria della radio (bassa, media, alta). La gamma 300-340 MHz esiste solo sulla versione russa (Argut A-36)."})
}

func kydo(m *genModel) *kydOpt { return m.Opt.(*kydOpt) }

func (kydFam) mem(m *genModel) *bwEl {
	if kydo(m).ip620 {
		return bwMust(ip620Format)
	}
	return bwMust(nc630aFormat)
}
func (f kydFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f kydFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f kydFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f kydFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f kydFam) Match(m *genModel, img []byte) bool {
	if kydo(m).ip620 {
		return len(img) == 0x2000 && bytes.Equal(img[0xF7E:0xF80], []byte{0x01, 0xE2})
	}
	return (len(img) == 0x338 || len(img) == 0x3C8) && bytes.Equal(img[0x1B8:0x1BE], []byte("P32073"))
}

func (f kydFam) names(m *genModel) (rx, tx, wide string) {
	if kydo(m).ip620 {
		return "rx_freq", "tx_freq", "w_n"
	}
	return "rxfreq", "txfreq", "wide"
}

func (f kydFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	rxn, txn, wn := f.names(m)
	ff := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	if bytes.Equal(r.F(rxn).Raw(img), ff) {
		return nil
	}
	rx := r.F(rxn).Int(img) * 10
	if rx == 0 && !kydo(m).ip620 {
		return nil
	}
	c := &Chan{Rx: rx}
	switch {
	case r.F(rxn).Int(img) == r.F(txn).Int(img):
		c.Tx = rx
	case bytes.Equal(r.F(txn).Raw(img), ff):
		c.TxOff, c.Tx = true, rx
	default:
		c.Tx = r.F(txn).Int(img) * 10
	}
	c.Narrow = !r.F(wn).Bool(img)
	c.TxTone = kenwoodToneDec(r.F("txtone").Int(img), 0x2800, 0x8000)
	c.RxTone = kenwoodToneDec(r.F("rxtone").Int(img), 0x2800, 0x8000)
	if kydo(m).ip620 {
		c.Power = m.powerSem(int(r.F("power").Int(img)))
		c.Scan = true
	} else {
		c.Power = m.powerSem(int(r.F("highpower").Int(img)))
		sk := f.mem(m).F("skipflags").I(i / 8)
		c.Scan = sk.Int(img)&(1<<uint(i%8)) != 0
	}
	return c
}

func (f kydFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := kydo(m)
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
	if o.ip620 && c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	rxn, txn, wn := f.names(m)
	wasEmpty := f.Get(m, img, i) == nil
	if o.ip620 {
		var keep [3]int64
		if !wasEmpty {
			keep = [3]int64{r.F("lout").Int(img), r.F("busy_loc").Int(img), r.F("scan_add").Int(img)}
		}
		r.F("lout").SetInt(img, keep[0])
		r.F("busy_loc").SetInt(img, keep[1])
		r.F("scan_add").SetInt(img, keep[2])
		r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	} else {
		bcl := int64(0)
		if !wasEmpty {
			bcl = r.F("bcl").Int(img)
		}
		r.SetRaw(img, append(make([]byte, 14), 0xFF, 0xFF))
		r.F("bcl").SetInt(img, bcl)
		r.F("highpower").SetBool(img, m.powerIndex(c.Power) == 1)
		sk := f.mem(m).F("skipflags").I(i / 8)
		v := sk.Int(img)
		if c.Scan {
			v |= 1 << uint(i%8)
		} else {
			v &^= 1 << uint(i%8)
		}
		sk.SetInt(img, v)
	}
	r.F(rxn).SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F(txn).Fill(img, 0xFF)
	} else {
		r.F(txn).SetInt(img, c.Tx/10)
	}
	r.F(wn).SetBool(img, !c.Narrow)
	r.F("rxtone").SetInt(img, kenwoodToneEnc(rt, 0x2800, 0x8000, 0xFFFF))
	r.F("txtone").SetInt(img, kenwoodToneEnc(tt, 0x2800, 0x8000, 0xFFFF))
	return nil
}

// chirpInvalid (solo per i test): modifiche che CHIRP non ammette per l'IP-620
// (toni diversi tra trasmissione e ricezione, «split») o che scrive in modo
// sbagliato (potenza alta, scritta come bassa da set_memory di kyd_IP620.py)
func (f kydFam) chirpInvalid(m *genModel, c *Chan) bool {
	if !kydo(m).ip620 {
		return false
	}
	if m.powerIndex(c.Power) == 2 {
		return true
	}
	d := c.Tx - c.Rx
	if d > 70000000 || d < -70000000 {
		return true
	}
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	switch {
	case t.Mode == "" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
	default:
		return true
	}
	return false
}

// kydEnter: presentazione, identificativo di 8 byte, conferma
func (s *rawSess) kydEnter(m *genModel) error {
	if kydo(m).ip620 {
		if err := s.send([]byte("iUHOUXUN\x02")); err != nil {
			return err
		}
		if err := s.sleep(200 * time.Millisecond); err != nil {
			return err
		}
	} else if err := s.send([]byte("PROGRAM")); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 0x06 {
		return rErr("la radio non è entrata in modalità di programmazione")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return err
	}
	id := readUpTo(s.ctx, s.p, 8, time.Second)
	want := []byte("P32073")
	if kydo(m).ip620 {
		want = []byte("\x06KG67\x01V\xF8")
	}
	if !bytes.HasPrefix(id, want) {
		return rErr("identificativo della radio sconosciuto (% X): è il modello giusto?", id)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return err
	}
	if b := readUpTo(s.ctx, s.p, 1, time.Second); len(b) != 1 || b[0] != 0x06 {
		return rErr("la radio non ha confermato l'identificazione")
	}
	return nil
}

func kydSizes(m *genModel) (memsize, rbs, wend, wbs int) {
	if kydo(m).ip620 {
		return 0x2000, 0x40, 0x2000, 0x10
	}
	return 0x3C8, 0x08, 0x330, 0x08
}

func (f kydFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	n, rbs, _, _ := kydSizes(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, n, "Collegamento con la radio…")
	if err := s.kydEnter(m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data, err := wxDownload(ctx, s, 0, n, rbs, time.Second, prog, true)
	if err == nil && kydo(m).ip620 {
		_ = s.send([]byte{0x06})
	}
	return data, "", err
}

func (f kydFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	n, _, wend, wbs := kydSizes(m)
	if len(img) < n {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), n)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, wend, "Collegamento con la radio…")
	if err := s.kydEnter(m); err != nil {
		return "", genFail("Identificazione", err)
	}
	if err := wxUpload(ctx, s, img, 0, wend, wbs, time.Second, prog); err != nil {
		return "", err
	}
	if kydo(m).ip620 {
		_ = s.send([]byte{0x06})
	}
	return "", nil
}
