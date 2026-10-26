// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radioddity_gm30.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Baofeng GM-15 Pro (radio GMRS): driver radioddity_gm30.py di CHIRP, classe
// BaofengGM15Pro (stessa radio della Radioddity GM-30, gamme 136-174 e 400-512 MHz).
// 250 canali di 16 byte e nomi di 6 caratteri; 57600 baud.
//
// Immagine (0xA000 byte): dieci aree da 0x1000 nell'ordine dei "tipi" 02 04 06 16
// 17 18 19 24 25 26 (impostazioni 0x1000, DTMF 0x2000, canali 0x3000, nomi 0x7000).
//
// Protocollo (do_download / do_upload di radioddity_gm30.py):
//   "PSEARCH" → 06 + "P13GMRS"; "PASSSTA" → 50 00 00; "SYSINFO" → 06;
//   56 + indirizzo (00000A0D, 00100A0D, 00200A0D → 13 byte; 0000000A → 11 byte), poi 06 → 06;
//   FF FF FF FF 0C; "P13GMRS" → 06; 02 → 8 byte, poi 06 → 06;
//   mappa dei tipi: per i = 1…15, 52 FF (i<<4|0F) 00 01 → 57 + eco + tipo, 06 → 06
//   (la radio sposta le aree in memoria: la posizione vale solo per la sessione);
//   lettura: 52 blocco posizione+i 00 40 → 57 + eco + 64 byte, 06 → 06;
//   scrittura (impostazioni, DTMF, canali, nomi): 57 blocco posizione+i 00 40 + 64 byte → 06;
//   fine: 06 06 00.
//
// Canali GMRS come in CHIRP: 1-7 e 15-22 TX = RX; 8-14 solo ricezione, banda e
// potenza fisse; 23-54 ripetitori (+5 MHz, 31-54 solo frequenze GMRS dei ponti);
// 1-30 frequenze fisse; 55-250 solo ricezione.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const gm30Format = `
struct{
    u8     unknown02[256];
} unknown02[16];
struct{
    u8      bootscrmode;
    u8      bsmodepad[15];
    u8      bootscreen1[10];
    u8      bs1pad[6];
    u8      bootscreen2[10];
    u8      bs2pad[6];
    u8      unused[16];
    u8      timeout;
    u8      squelch;
    u8      vox_level;
    u8      batt_save:4,
            unk_bits:2,
            work_mode:1,
            voice_alert:1;
    u8      backlight;
    u8      beep_tone:1,
            auto_key_lock:1,
            unk_bit_2:1,
            ctcss_revert:1,
            scan_type:2,
            side_tone:2;
    u8      unk_bit_3:1,
            standby:1,
            roger:1,
            alarm_mode:2,
            alarm_sound:1,
            fm_radio:1,
            unk_bit_4:1;
    u8      tail_revert;
    u8      tail_delay;
    u8      tbst;
    u8      unk_sett[6];
    u8      unk_bits_5:6,
            b_ch_disp:1,
            a_ch_disp:1;
    u8      unk_sett_2[24];
    u8      passw_w_ena;
    u8      passw_r_ena;
    u8      passw_w_val[8];
    u8      passw_r_val[8];
    u8      unk_sett_3[5];
} settings;
struct{
    u8     unusedsettings[32];
} unusedsettings[124];
struct{
    u8     entry[5];
} dtmf_list[15];
struct{
    u8      dtmf_l_pad[5];
    u8      radio_id[5];
    u8      unk_dtmf[11];
    u8      unk_dtmf_bits:6,
            press_send:1,
            release_send:1;
    u8      delay_time;
    u8      digit_dur;
    u8      inter_dur;
    u8      unk_dtmf_2[28];
} dtmf;
struct{
    u8     unkdtmf[128];
} unuseddtmf[31];
struct{
    u8        unk_mem[16];
} tuning;
struct {
    lbcd      rx_freq[4];
    lbcd      tx_freq[4];
    u8        mem_changed;
    lbcd      rx_tone[2];
    lbcd      tx_tone[2];
    u8      unk_mem2:1,
            busy_lock:1,
            ptt_id:2,
            unk_mem3:1,
            mode:1,
            power:1,
            unk_mem4:1;
            u8 signal:4,
            unk_mem5:2,
            freq_hop:1,
            scan:1;
            u8 step;
} vfomemory[2];
struct {
    lbcd      rx_freq[4];
    lbcd      tx_freq[4];
    u8        mem_changed;
    lbcd      rx_tone[2];
    lbcd      tx_tone[2];
    u8      unk_mem2:1,
            busy_lock:1,
            ptt_id:2,
            unk_mem3:1,
            mode:1,
            power:1,
            unk_mem4:1;
            u8 signal:4,
            unk_mem5:2,
            freq_hop:1,
            scan:1;
            u8 step;
} memory[250];
struct{
    u8      unused_memch[16];
} unused_mem[3];
struct{
    u8     unknown17[256];
} unknown17[16];
struct{
    u8     unknown18[256];
} unknown18[16];
struct{
    u8     unknown19[256];
} unknown19[16];
struct{
    char      name[6];
    u8      pad[5];
} memnames[250];
struct{
    u8      unknames[64];
} unknownnames [21];
struct{
    u8      unknown25[256];
} unknown25[16];
struct{
    u8      unknown26[256];
} unknown26[16];
`

// tipi di dati nell'ordine dell'immagine (TYPE_MAP) e quelli scritti (WRITE_MAP)
var gm30Types = []byte{0x02, 0x04, 0x06, 0x16, 0x17, 0x18, 0x19, 0x24, 0x25, 0x26}
var gm30Write = []struct {
	typ byte
	off int
}{{0x04, 0x1000}, {0x06, 0x2000}, {0x16, 0x3000}, {0x24, 0x7000}}

var gm30Rptr = []int64{462550000, 462575000, 462600000, 462625000, 462650000, 462675000, 462700000, 462725000}

const gm30Size = 0xA000

type gm30Opt struct {
	ident []byte
}

type gm30Fam struct{}

func init() {
	genRegister(&genModel{ID: "bf-gm15pro", Vendor: "Baofeng", Model: "GM-15Pro", RClass: "BaofengGM15Pro",
		Name: "Baofeng GM-15 Pro", Slots: 250, First: 1, NameMax: 6,
		Powers: []genPower{{"Bassa", 0.5, 2}, {"Alta", 3, 0}},
		Bands:  []band{{136000000, 174000000}, {400000000, 512000000}}, Narrow: true, HasScan: true,
		DCS: dtcsCodes, Baud: 57600, Fam: gm30Fam{}, Opt: &gm30Opt{ident: []byte("P13GMRS")},
		Notes: "Radio GMRS (USA). Canali come in CHIRP: 1-30 frequenze GMRS fisse (8-14 solo ricezione), " +
			"23-54 ponti GMRS (+5 MHz; 31-54 solo frequenze GMRS dei ponti), 55-250 solo ricezione. " +
			"Driver di CHIRP (radioddity_gm30, GM-15Pro)."})
}

func gm30o(m *genModel) *gm30Opt { return m.Opt.(*gm30Opt) }

func (gm30Fam) mem() *bwEl                 { return bwMust(gm30Format) }
func (f gm30Fam) layout(m *genModel) *bwEl { return f.mem() }

func (f gm30Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("memory").I(i), b.F("memnames").I(i)
}

func (f gm30Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, n := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), n.Raw(img)...)
}

func (f gm30Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, n := f.recs(i)
	r.SetRaw(img, raw[:16])
	n.SetRaw(img, raw[16:27])
}

func (f gm30Fam) Match(m *genModel, img []byte) bool { return false }

// numero del canale (1…250) → regole GMRS di get_memory
func gm30Fixed(i int) bool    { return i < 30 } // frequenze non modificabili
func gm30RxOnly(i int) bool   { n := i + 1; return n >= 8 && n <= 14 || n >= 55 }
func gm30Repeater(i int) bool { n := i + 1; return n >= 23 && n <= 54 }

// lockedSlot: canali 1-54 con frequenze o scostamento imposti (esclusi dalle prove)
func (f gm30Fam) lockedSlot(m *genModel, i int) bool { return i < 54 }

// rxOnlySlot: canali in sola ricezione (nelle prove si scrivono senza TX)
func (f gm30Fam) rxOnlySlot(m *genModel, i int) bool { return gm30RxOnly(i) }

// tono: lbcd[2], bit 0x80 (DCS) e 0x40 (invertito) del secondo byte (get_tone)
func gm30ToneDec(raw []byte) string {
	if raw[1] == 0xFF || raw[1] == 0x00 {
		return ""
	}
	bcd := func(b byte) int { return int(b>>4)*10 + int(b&0x0F) }
	v := bcd(raw[1]&0x3F)*100 + bcd(raw[0])
	if raw[1]&0x80 != 0 {
		t := genTone{Mode: "DTCS", Code: v, Inv: raw[1]&0x40 != 0}
		return t.String()
	}
	return fmtTone(v)
}

func gm30ToneEnc(dst []byte, t genTone) {
	enc := func(v int) (byte, byte) {
		return byte((v/10%10)<<4 | v%10), byte((v/1000%10)<<4 | v/100%10)
	}
	switch t.Mode {
	case "Tone":
		dst[0], dst[1] = enc(t.Hz10)
	case "DTCS":
		dst[0], dst[1] = enc(t.Code)
		dst[1] |= 0x80
		if t.Inv {
			dst[1] |= 0x40
		}
	default:
		dst[0], dst[1] = 0xFF, 0xFF
	}
}

func (f gm30Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, n := f.recs(i)
	if bytes.Equal(r.F("rx_freq").Raw(img), ff4) {
		return nil
	}
	c := &Chan{Rx: r.F("rx_freq").Int(img) * 10}
	switch {
	case gm30RxOnly(i):
		c.TxOff, c.Tx = true, c.Rx
	case gm30Repeater(i):
		c.Tx = c.Rx + 5000000
	default:
		c.Tx = c.Rx
	}
	var nb []byte
	for _, b := range n.F("name").Raw(img) {
		if b >= 0x20 && b <= 0x7E {
			nb = append(nb, b)
		}
	}
	c.Name = strings.TrimRight(string(nb), " \t\n\r\x0b\x0c")
	c.TxTone = gm30ToneDec(r.F("tx_tone").Raw(img))
	c.RxTone = gm30ToneDec(r.F("rx_tone").Raw(img))
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("mode").Int(img) == 0
	c.Scan = r.F("scan").Bool(img)
	return c
}

func (f gm30Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n := f.recs(i)
	name := ""
	if c != nil {
		name = c.Name
	}
	if len(name) > 6 {
		name = name[:6]
	}
	n.F("name").SetStr(img, name, 0x00)
	if c == nil {
		r.SetRaw(img, append(bytes.Repeat([]byte{0xFF}, 13), 0x06, 0x11, 0x00))
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
	existed := !bytes.Equal(r.F("rx_freq").Raw(img), ff4)
	r.F("rx_freq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("tx_freq").Fill(img, 0xFF)
	} else {
		r.F("tx_freq").SetInt(img, c.Tx/10)
	}
	mode := int64(1)
	if c.Narrow {
		mode = 0
	}
	r.F("mode").SetInt(img, mode)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("scan").SetBool(img, c.Scan)
	rec := r.Raw(img)
	gm30ToneEnc(rec[11:13], tt)
	gm30ToneEnc(rec[9:11], rt)
	// CHIRP riscrive le impostazioni extra di un canale esistente: "DTMF ID" 0 diventa 1
	if existed && r.F("signal").Int(img) == 0 {
		r.F("signal").SetInt(img, 1)
	}
	return nil
}

// Locked: regole dei canali GMRS (immutable e validate_memory di CHIRP)
func (f gm30Fam) Locked(m *genModel, i int, old, want *Chan) error {
	n := i + 1
	if n <= 54 && want == nil {
		return fmt.Errorf("i canali GMRS 1-54 della GM-15 Pro non si possono cancellare")
	}
	if want == nil {
		return nil
	}
	if gm30Fixed(i) && (old == nil || old.Rx != want.Rx) {
		return fmt.Errorf("i canali 1-30 della GM-15 Pro sono canali GMRS fissi: frequenza non modificabile")
	}
	if n >= 31 && n <= 54 {
		ok := false
		for _, f := range gm30Rptr {
			ok = ok || want.Rx == f
		}
		if !ok {
			return fmt.Errorf("sui canali 31-54 della GM-15 Pro sono ammesse solo le frequenze GMRS dei ponti (462,550-462,725 MHz)")
		}
	}
	switch {
	case gm30RxOnly(i):
		if !want.TxOff {
			if n <= 14 {
				return fmt.Errorf("i canali GMRS 8-14 della GM-15 Pro sono solo in ricezione: imposta TX OFF")
			}
			return fmt.Errorf("i canali 55-250 della GM-15 Pro sono solo in ricezione: imposta TX OFF")
		}
		if n <= 14 && old != nil && (old.Narrow != want.Narrow || old.Power != want.Power) {
			return fmt.Errorf("i canali GMRS 8-14 della GM-15 Pro hanno banda e potenza fisse")
		}
	case gm30Repeater(i):
		if want.TxOff || want.Tx != want.Rx+5000000 {
			return fmt.Errorf("i canali 23-54 della GM-15 Pro sono canali per ponti GMRS: TX = RX + 5 MHz")
		}
	default:
		if want.TxOff || want.Tx != want.Rx {
			return fmt.Errorf("sui canali GMRS 1-7 e 15-22 della GM-15 Pro TX e RX coincidono")
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

const gm30Timeout = 1500 * time.Millisecond

type gm30Sess struct {
	*rawSess
}

func (s gm30Sess) xfer(out []byte, n int) ([]byte, error) {
	if err := s.send(out); err != nil {
		return nil, err
	}
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, gm30Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

func (s gm30Sess) ackAck() error {
	b, err := s.xfer([]byte{0x06}, 1)
	if err == nil && b[0] != 0x06 {
		err = rErr("conferma 06 attesa, ricevuto %02X", b[0])
	}
	return err
}

func (s gm30Sess) readCmd(cmd []byte, n int) ([]byte, error) {
	b, err := s.xfer(append([]byte{0x52}, cmd...), n+len(cmd)+1)
	if err != nil {
		return nil, err
	}
	if b[0] != 0x57 {
		return nil, rErr("risposta inattesa al comando di lettura (% X)", b[:min(len(b), 8)])
	}
	if err := s.ackAck(); err != nil {
		return nil, err
	}
	return b[len(cmd)+1:], nil
}

// handshake e mappa dei tipi (enter_prog … do_read_tlmap)
func (s gm30Sess) start(o *gm30Opt) (map[byte]int, error) {
	b, err := s.xfer([]byte("PSEARCH"), 8)
	if err != nil {
		return nil, err
	}
	if b[0] != 0x06 {
		return nil, rErr("ingresso in programmazione rifiutato (% X)", b)
	}
	if !bytes.Equal(b[1:], o.ident) {
		return nil, rErr("la radio si è identificata come %q, non come %q: è il modello giusto?", b[1:], o.ident)
	}
	if b, err = s.xfer([]byte("PASSSTA"), 3); err != nil {
		return nil, err
	}
	if !bytes.Equal(b, []byte{0x50, 0, 0}) {
		return nil, rErr("risposta inattesa a PASSSTA (% X): radio protetta da password?", b)
	}
	if b, err = s.xfer([]byte("SYSINFO"), 1); err != nil {
		return nil, err
	}
	if b[0] != 0x06 {
		return nil, rErr("risposta inattesa a SYSINFO (%02X)", b[0])
	}
	for _, c := range []struct {
		addr uint32
		n    int
	}{{0x00000a0d, 13}, {0x00100a0d, 13}, {0x00200a0d, 13}, {0x0000000a, 11}} {
		cmd := []byte{0x56, byte(c.addr >> 24), byte(c.addr >> 16), byte(c.addr >> 8), byte(c.addr)}
		if _, err := s.xfer(cmd, c.n); err != nil {
			return nil, err
		}
		if err := s.ackAck(); err != nil {
			return nil, err
		}
	}
	if err := s.send([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0C}); err != nil {
		return nil, err
	}
	if b, err = s.xfer(o.ident, 1); err != nil {
		return nil, err
	}
	if b[0] != 0x06 {
		return nil, rErr("risposta inattesa all'identificativo (%02X)", b[0])
	}
	if _, err = s.xfer([]byte{0x02}, 8); err != nil {
		return nil, err
	}
	if err := s.ackAck(); err != nil {
		return nil, err
	}
	tl := map[byte]int{}
	for _, t := range gm30Types {
		tl[t] = 0
	}
	for i := 1; i < 16; i++ {
		r, err := s.readCmd([]byte{0xFF, byte(i<<4 | 0x0F), 0x00, 0x01}, 1)
		if err != nil {
			return nil, err
		}
		if _, ok := tl[r[0]]; ok {
			tl[r[0]] = i << 4
		}
	}
	return tl, nil
}

func (s gm30Sess) exit() { _ = s.send([]byte{0x06, 0x06, 0x00}) }

func (f gm30Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer rs.p.Close()
	s := gm30Sess{rs}
	prog(0, gm30Size, "Collegamento con la radio…")
	tl, err := s.start(gm30o(m))
	if err != nil {
		s.exit()
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for _, t := range gm30Types {
		loc := tl[t]
		if loc == 0 {
			s.exit()
			return nil, "", genFail("Mappa della memoria", rErr("area %02X non trovata", t))
		}
		for i := 0; i < 16; i++ {
			for pre := 0; pre <= 0xC0; pre += 0x40 {
				if ctx.Err() != nil {
					s.exit()
					return nil, "", errCancelled
				}
				prog(len(data), gm30Size, "Lettura…")
				d, err := s.readCmd([]byte{byte(pre), byte(loc + i), 0x00, 0x40}, 0x40)
				if err != nil {
					s.exit()
					return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", len(data)), err)
				}
				data = append(data, d...)
			}
		}
	}
	s.exit()
	prog(gm30Size, gm30Size, "Lettura completata")
	return data, "", nil
}

func (f gm30Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < gm30Size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer rs.p.Close()
	s := gm30Sess{rs}
	total := len(gm30Write) * 0x1000
	prog(0, total, "Collegamento con la radio…")
	tl, err := s.start(gm30o(m))
	if err != nil {
		s.exit()
		return "", genFail("Identificazione", err)
	}
	done := 0
	for _, w := range gm30Write {
		loc := tl[w.typ]
		if loc == 0 {
			s.exit()
			return "", genFail("Mappa della memoria", rErr("area %02X non trovata", w.typ))
		}
		for i := 0; i < 16; i++ {
			for pre := 0; pre <= 0xC0; pre += 0x40 {
				if ctx.Err() != nil {
					s.exit()
					return "", errCancelled
				}
				prog(done, total, "Scrittura…")
				x := w.off + i*0x100 + pre
				cmd := append([]byte{0x57, byte(pre), byte(loc + i), 0x00, 0x40}, img[x:x+0x40]...)
				b, err := s.xfer(cmd, 1)
				if err == nil && b[0] != 0x06 {
					err = rErr("la radio ha rifiutato il blocco")
				}
				if err != nil {
					s.exit()
					return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", x), err)
				}
				done += 0x40
			}
		}
	}
	s.exit()
	prog(total, total, "Scrittura completata")
	return "", nil
}
