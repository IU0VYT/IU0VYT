// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// vx5.py (4.37).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.37: Yaesu VX-5R: 220 canali con nomi di 8 caratteri, ricevitore a larga
// banda, 9600 baud. La copia parte dalla radio ed è in tre blocchi (10, 16 e
// 8097 byte), i primi due confermati (vedi drv_yaesuclone.go). La radio non
// dichiara il modello: la memoria si riconosce dalla lunghezza, come in CHIRP.
// Il modo (FM, AM, FM larga) non è una colonna del programma: resta quello del
// canale se la frequenza non cambia, altrimenti segue la frequenza.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare viene rifiutata invece di
//     essere troncata al kHz;
//   - uno scostamento che non è un numero intero di kHz si scrive come
//     frequenza di trasmissione separata (CHIRP lo tronca);
//   - il canale con il salto «P» resta tale quando lo si modifica senza
//     toglierlo dalla scansione.

import (
	"context"
	"fmt"
	"strings"
)

type vx5Fam struct{}

const vx5Format = `
struct mem_struct {
  u8 unknown1:2,
     half_deviation:1,
     cpu_shift:1,
     unknown2:1,
     sp_rx1:1,
     unknown3:1,
     sp_rx2:1;
  u8 unknown4:4,
     tuning_step:4;
  bbcd freq[3];
  u8 icon:6,
     mode:2;
  char name[8];
  bbcd offset[3];
  u8 tmode:4,
     power:2,
     duplex:2;
  u8 unknown5:2,
     tone:6;
  u8 unknown6:1,
     dtcs:7;
  u8 unknown7;
  u8 unknown8;
};

struct flag_struct {
  u8 zeros:4,
     pskip:1,
     skip:1,
     visible:1,
     used:1;
};

#seekto 0x002A;
struct {
  u8 current_member;
} bank_used[5];

#seekto 0x0032;
struct {
  struct {
    u8 status;
    u8 channel;
  } members[24];
} bank_groups[5];

#seekto 0x012A;
struct flag_struct flag[220];
struct flag_struct specialflag[20];

#seekto 0x026A;
struct mem_struct memory[220];
struct mem_struct special[50];

#seekto 0x1D03;
u8 current_bank;
`

var vx5Clone = &yaesuOpt{model: "", size: 8123, head: 10, blocks: []int{10, 16, 8097}, chunk: 8,
	sums: [][3]int{{0x0000, 0x1FB9, 0x1FBA}}, send: "[VFO(DW)SC]", recv: "[MR(SKP)SC] («CLONE WAIT»)"}

// chirpOldTones: i 39 subtoni delle radio più vecchie (OLD_TONES di CHIRP), in
// decimi di Hz
var chirpOldTones = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148,
	1188, 1230, 1273, 1318, 1365, 1413, 1462, 1514, 1567, 1622, 1679, 1738, 1799, 1862, 1928, 2035, 2107, 2181,
	2257, 2336, 2418, 2503}

// yaesuASCII: tutti i caratteri ASCII stampabili (nomi dei canali di VX-5 e VX-150)
var yaesuASCII = func() string {
	var sb strings.Builder
	for c := byte(' '); c <= '~'; c++ {
		sb.WriteByte(c)
	}
	return sb.String()
}()

// yaesuNote: canali, lunghezza dei nomi e velocità (una sola frase per tutti i modelli)
func yaesuNote(slots, name, baud int) string {
	return fmt.Sprintf("%d canali con nomi di %d caratteri, cavo di programmazione Yaesu, %d baud. ", slots, name, baud)
}

func init() {
	genRegister(
		&genModel{ID: "ys-vx5", Vendor: "Yaesu", Model: "VX-5", RClass: "VX5Radio", Name: "Yaesu VX-5R",
			Slots: 220, First: 1, NameMax: 8, Charset: yaesuASCII,
			Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 1 W", 1, 2}, {"Minima 0,05 W", 0.05, 3}},
			Bands:  []band{{500000, 16000000}, {48000000, 729000000}, {800000000, 999000000}},
			Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: vx5Fam{}, Opt: vx5Clone,
			Notes: yaesuNote(220, 8, 9600) + yaesuCloneHow("MIC/EAR", "[F/W]", "[VFO(DW)SC]", "[MR(SKP)SC]") + "Driver di CHIRP."},
	)
}

func (vx5Fam) yaesu(m *genModel) *yaesuOpt { return m.Opt.(*yaesuOpt) }
func (vx5Fam) mem() *bwEl                  { return bwMust(vx5Format) }
func (f vx5Fam) layout(m *genModel) *bwEl  { return f.mem() }
func (f vx5Fam) rec(i int) *bwEl           { return f.mem().F("memory").I(i) }
func (f vx5Fam) flag(i int) *bwEl          { return f.mem().F("flag").I(i) }

func (f vx5Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(i).Raw(img)...), f.flag(i).Raw(img)...)
}

func (f vx5Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(i).ByteLen()
	if len(raw) != n+1 {
		return
	}
	f.rec(i).SetRaw(img, raw[:n])
	f.flag(i).SetRaw(img, raw[n:])
	f.flag(i).F("used").SetBool(img, true)
	f.flag(i).F("visible").SetBool(img, true)
}

func (f vx5Fam) Match(m *genModel, img []byte) bool { return yaesuMatch(f.yaesu(m), img) }

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (vx5Fam) noIdent() {}

// vx5Tmode: modo dei toni (0 nessuno, 1 subtono in trasmissione, 2 lo stesso
// subtono nei due versi, 3 lo stesso codice DCS nei due versi); -1 se la radio
// non ha quella combinazione
func vx5Tmode(t, r genTone) int64 {
	switch k := ft60Tmode(t, r); k {
	case 0, 1, 2:
		return k
	case 4:
		return 3
	}
	return -1
}

func yaesuOldTone(hz10 int) int64 {
	for k, x := range chirpOldTones {
		if x == hz10 {
			return int64(k)
		}
	}
	return -1
}

// vx5ToneOK: combinazione e subtoni che la radio ha (39 subtoni)
func vx5ToneOK(t, r genTone) bool {
	if vx5Tmode(t, r) < 0 {
		return false
	}
	for _, g := range []genTone{t, r} {
		if g.Mode == "Tone" && yaesuOldTone(g.Hz10) < 0 {
			return false
		}
	}
	return true
}

func (vx5Fam) toneOK(t, r genTone) bool { return vx5ToneOK(t, r) }
func (vx5Fam) fitBySet()                {}

func (vx5Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !vx5ToneOK(t, r)
}

func (f vx5Fam) Get(m *genModel, img []byte, i int) *Chan {
	fl := f.flag(i)
	if !fl.F("visible").Bool(img) || !fl.F("used").Bool(img) {
		return nil
	}
	r := f.rec(i)
	c := &Chan{Rx: yaesuRound(r.F("freq").Int(img) * 1000)}
	if c.Rx <= 0 {
		return nil
	}
	off := r.F("offset").Int(img) * 1000
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - off
	case 2:
		c.Tx = c.Rx + off
	case 3:
		c.Tx = yaesuRound(off)
	default:
		c.Tx = c.Rx
	}
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img)); k < len(chirpOldTones) {
		tone = fmtTone(chirpOldTones[k])
	}
	if k := int(r.F("dtcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	switch r.F("tmode").Int(img) & 3 {
	case 1:
		c.TxTone = tone
	case 2:
		c.TxTone, c.RxTone = tone, tone
	case 3:
		c.TxTone, c.RxTone = dcs, dcs
	}
	c.Power = m.powerSem(int(3 - r.F("power").Int(img)))
	c.Narrow = r.F("half_deviation").Bool(img) && r.F("mode").Int(img) == 0
	c.Scan = !fl.F("skip").Bool(img)
	c.Name = strings.TrimRight(yaesuCleanASCII(r.F("name").Raw(img)), " ")
	return c
}

// yaesuCleanASCII: nome scritto in ASCII; i byte non stampabili diventano spazi
func yaesuCleanASCII(b []byte) string {
	out := make([]byte, len(b))
	for k, c := range b {
		if c < ' ' || c > '~' {
			c = ' '
		}
		out[k] = c
	}
	return string(out)
}

// wipeBanks: il canale i esce dai banchi in cui si trova (come
// _wipe_memory_banks di CHIRP con il modello dei banchi della VX-5)
func (f vx5Fam) wipeBanks(img []byte, i int) {
	root := f.mem()
	for b := 0; b < 5; b++ {
		used := root.F("bank_used").I(b).F("current_member")
		if used.Int(img) == 0xFF {
			continue
		}
		mem := root.F("bank_groups").I(b).F("members")
		in := false
		for k := 0; k < mem.Len(); k++ {
			if mem.I(k).F("status").Int(img) != 0xFF && mem.I(k).F("channel").Int(img) == int64(i) {
				in = true
			}
		}
		if !in {
			continue
		}
		remaining := 0
		for k := 0; k < mem.Len(); k++ {
			if mem.I(k).F("status").Int(img) != 0x00 {
				continue
			}
			if mem.I(k).F("channel").Int(img) == int64(i) {
				mem.I(k).F("status").SetInt(img, 0xFF)
			} else {
				remaining++
			}
		}
		if remaining == 0 {
			used.SetInt(img, 0xFF)
		}
	}
}

func (f vx5Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, fl := f.rec(i), f.flag(i)
	used, visible := fl.F("used").Bool(img), fl.F("visible").Bool(img)
	if c == nil {
		// come CHIRP: il canale visibile viene nascosto, quello già nascosto liberato
		if used && !visible {
			fl.F("used").SetBool(img, false)
			return nil
		}
		fl.F("visible").SetBool(img, false)
		f.wipeBanks(img, i)
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := vx5Tmode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	kept, tone, dcs, step, oldRx, oldMode := int64(600), int64(8), int64(0), int64(0), int64(0), int64(0)
	if used {
		kept = r.F("offset").Int(img)
		if r.F("duplex").Int(img) == 3 {
			kept = yaesuRound(kept*1000) / 1000
		}
		if k := r.F("tone").Int(img); int(k) < len(chirpOldTones) {
			tone = k
		}
		if k := r.F("dtcs").Int(img); int(k) < len(dtcsCodes) {
			dcs = k
		}
		if k := r.F("tuning_step").Int(img); k < 8 {
			step = k
		}
		oldRx, oldMode = yaesuRound(r.F("freq").Int(img)*1000), r.F("mode").Int(img)
	}
	if t.Mode == "Tone" {
		if tone = yaesuOldTone(t.Hz10); tone < 0 {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	if t.Mode == "DTCS" {
		dcs = -1
		for k, x := range dtcsCodes {
			if x == t.Code {
				dcs = int64(k)
			}
		}
		if dcs < 0 {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	dup, offs := int64(0), kept
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	switch {
	case d == 0:
	case ad > 70000000 || ad%1000 != 0:
		tk, ok := yaesuKHz(c.Tx)
		if !ok {
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
		}
		dup, offs = 3, tk
	case d > 0:
		dup, offs = 2, ad/1000
	default:
		dup, offs = 1, ad/1000
	}
	// modo: resta quello del canale se la frequenza non cambia, altrimenti segue la frequenza
	mode := []int64{0, 2, 1}[icx90Mode(c.Rx)] // 0 FM, 1 AM, 2 FM larga
	if used && visible && oldRx == c.Rx && (oldMode == 1 || oldMode == 2) {
		mode = oldMode
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !used {
		for _, k := range []string{"unknown1", "unknown2", "unknown3", "unknown4", "unknown5", "unknown6", "unknown7", "unknown8",
			"cpu_shift", "sp_rx1", "sp_rx2"} {
			r.F(k).SetInt(img, 0)
		}
		r.F("icon").SetInt(img, 12) // l'icona dello schedario, come CHIRP
	}
	wasP := used && visible && fl.F("pskip").Bool(img)
	fl.F("used").SetBool(img, true)
	fl.F("visible").SetBool(img, true)
	r.F("freq").SetInt(img, fk)
	r.F("duplex").SetInt(img, dup)
	name := m.cleanName(c.Name)
	if len(name) > 8 {
		name = name[:8]
	}
	r.F("name").SetStr(img, name, ' ')
	r.F("mode").SetInt(img, mode)
	r.F("half_deviation").SetBool(img, c.Narrow && mode == 0)
	r.F("tuning_step").SetInt(img, step)
	r.F("offset").SetInt(img, offs)
	r.F("power").SetInt(img, 3-int64(m.powerIndex(c.Power)))
	r.F("tmode").SetInt(img, tmode)
	r.F("tone").SetInt(img, tone)
	r.F("dtcs").SetInt(img, dcs)
	fl.F("skip").SetBool(img, !c.Scan)
	fl.F("pskip").SetBool(img, wasP && c.Scan)
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

func (f vx5Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := yaesuDownload(ctx, m, port, f.yaesu(m), prog)
	return img, "", err
}

func (f vx5Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", yaesuUpload(ctx, m, port, f.yaesu(m), img, prog)
}
