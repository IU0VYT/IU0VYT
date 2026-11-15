// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// vx7.py (4.37).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.37: Yaesu VX-7R: 450 canali con nomi di 8 caratteri (anche minuscole),
// ricevitore a larga banda, 19200 baud. La copia parte dalla radio ed è in tre
// blocchi (10, 8 e 16193 byte), i primi due confermati; tre somme di controllo.
// La radio non dichiara il modello: la memoria si riconosce dalla lunghezza,
// come in CHIRP. Niente banda stretta. Il modo (FM, AM, FM larga, automatico)
// non è una colonna del programma: resta quello del canale se la frequenza non
// cambia, altrimenti segue la frequenza.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare viene rifiutata invece di
//     essere troncata al kHz;
//   - uno scostamento che non è un numero intero di kHz si scrive come
//     frequenza di trasmissione separata (CHIRP lo tronca);
//   - DCS in trasmissione con subtono in ricezione: il subtono scritto è quello
//     di ricezione (CHIRP lascia il valore precedente);
//   - fra 222 e 225 MHz la radio ha solo le due potenze più basse: una potenza
//     più alta diventa la seconda (CHIRP rifiuta il canale);
//   - il canale con il salto «P» resta tale quando lo si modifica senza
//     toglierlo dalla scansione.

import (
	"context"
	"fmt"
	"strings"
)

type vx7Fam struct{}

const vx7Format = `
#seekto 0x0611;
u8 checksum1;

#seekto 0x0691;
u8 checksum2;

#seekto 0x0742;
struct {
  u16 in_use;
} bank_used[9];

#seekto 0x0EA2;
struct {
  u16 members[48];
} bank_members[9];

#seekto 0x3F52;
u8 checksum3;

#seekto 0x1202;
struct {
  u8 even_pskip:1,
     even_skip:1,
     even_valid:1,
     even_masked:1,
     odd_pskip:1,
     odd_skip:1,
     odd_valid:1,
     odd_masked:1;
} flags[225];

#seekto 0x1322;
struct {
  u8   unknown1;
  u8   power:2,
       duplex:2,
       tune_step:4;
  bbcd freq[3];
  u8   zeros1:2,
       ones:2,
       zeros2:2,
       mode:2;
  u8   name[8];
  u8 zero;
  bbcd offset[3];
  u8   zeros3:2,
       tone:6;
  u8   zeros4:1,
       dcs:7;
  u8   zeros5:5,
       is_split_tone:1,
       tmode:2;
  u8   charset;
} memory[450];
`

var vx7Clone = &yaesuOpt{model: "", size: 16211, head: 10, blocks: []int{10, 8, 16193}, chunk: 8,
	sums: [][3]int{{0x0592, 0x0610, 0x0611}, {0x0612, 0x0690, 0x0691}, {0x0000, 0x3F51, 0x3F52}},
	send: "[BAND]", recv: "[V/M] («CLONE WAIT»)"}

// caratteri della radio nell'ordine dei codici (il punto e il punto di domanda
// compaiono due volte: si scrive il primo codice, come CHIRP)
const vx7Chars = "0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz.,:;!\"#$%&'()*+-.=<>?@[?]^_\\{|}"

func init() {
	genRegister(
		&genModel{ID: "ys-vx7", Vendor: "Yaesu", Model: "VX-7", RClass: "VX7Radio", Name: "Yaesu VX-7R",
			Slots: 450, First: 1, NameMax: 8,
			Charset: "0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz.,:;!\"#$%&'()*+-=<>?@[]^_\\{|}",
			Powers:  []genPower{{"Alta 5 W", 5, 0}, {"Media 2,5 W", 2.5, 1}, {"Bassa 1 W", 1, 2}, {"Minima 0,05 W", 0.05, 3}},
			Bands:   []band{{500000, 999000000}}, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 19200, Fam: vx7Fam{}, Opt: vx7Clone,
			Notes: yaesuNote(450, 8, 19200) + yaesuCloneHow("MIC/SP", "[MON-F]", "[BAND]", "[V/M]") + "Driver di CHIRP."},
	)
}

func (vx7Fam) yaesu(m *genModel) *yaesuOpt { return m.Opt.(*yaesuOpt) }
func (vx7Fam) mem() *bwEl                  { return bwMust(vx7Format) }
func (f vx7Fam) layout(m *genModel) *bwEl  { return f.mem() }
func (f vx7Fam) rec(i int) *bwEl           { return f.mem().F("memory").I(i) }

// flag: campo dei bit del canale (i pari: metà «odd», i dispari: metà «even»)
func (f vx7Fam) flag(i int, what string) *bwEl {
	half := "odd_"
	if i%2 == 1 {
		half = "even_"
	}
	return f.mem().F("flags").I(i / 2).F(half + what)
}

func (f vx7Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(i).Raw(img)...)
	return append(out, icomB(f.flag(i, "pskip").Bool(img)), icomB(f.flag(i, "skip").Bool(img)))
}

func (f vx7Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	n := f.rec(i).ByteLen()
	if len(raw) != n+2 {
		return
	}
	f.rec(i).SetRaw(img, raw[:n])
	f.flag(i, "pskip").SetBool(img, raw[n] != 0)
	f.flag(i, "skip").SetBool(img, raw[n+1] != 0)
	f.flag(i, "valid").SetBool(img, true)
	f.flag(i, "masked").SetBool(img, true)
}

func (f vx7Fam) Match(m *genModel, img []byte) bool { return yaesuMatch(f.yaesu(m), img) }

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (vx7Fam) noIdent() {}

func (vx7Fam) toneOK(t, r genTone) bool { return ft60Tmode(t, r) >= 0 }
func (vx7Fam) fitBySet()                {}

func (vx7Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	k := ft60Tmode(t, r)
	return k < 0 || k == 7
}

// chirpPower (solo prove): CHIRP elenca le potenze dalla più bassa
func (vx7Fam) chirpPower(m *genModel, img []byte, i int, sem int) int { return 3 - m.powerIndex(sem) }

func vx7Is220(hz int64) bool { return hz >= 222000000 && hz <= 225000000 }

func (f vx7Fam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.flag(i, "masked").Bool(img) || !f.flag(i, "valid").Bool(img) {
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
	if k := int(r.F("tone").Int(img)); k < len(chirpTones) {
		tone = fmtTone(chirpTones[k])
	}
	if k := int(r.F("dcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	k := r.F("tmode").Int(img)
	if r.F("is_split_tone").Bool(img) {
		switch k {
		case 0:
			c.TxTone = dcs
		case 1:
			c.TxTone, c.RxTone = tone, dcs
		case 2:
			c.TxTone, c.RxTone = dcs, tone
		}
	} else {
		switch k {
		case 1:
			c.TxTone = tone
		case 2:
			c.TxTone, c.RxTone = tone, tone
		case 3:
			c.TxTone, c.RxTone = dcs, dcs
		}
	}
	p := r.F("power").Int(img)
	if vx7Is220(c.Rx) && p > 1 {
		p = 0 // come CHIRP: valore non ammesso fra 222 e 225 MHz
	}
	c.Power = m.powerSem(int(3 - p))
	c.Scan = !f.flag(i, "skip").Bool(img)
	var sb strings.Builder
	for _, b := range r.F("name").Raw(img) {
		if int(b) < len(vx7Chars) {
			sb.WriteByte(vx7Chars[b])
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

// wipeBanks: il canale i esce dai banchi in cui si trova (come
// _wipe_memory_banks di CHIRP con il modello dei banchi della VX-7)
func (f vx7Fam) wipeBanks(img []byte, i int) {
	root := f.mem()
	for b := 0; b < 9; b++ {
		used := root.F("bank_used").I(b).F("in_use")
		if used.Int(img) == 0xFFFF {
			continue
		}
		mem := root.F("bank_members").I(b).F("members")
		found, remaining := false, 0
		for k := 0; k < mem.Len(); k++ {
			switch v := mem.I(k).Int(img); {
			case v == int64(i):
				found = true
			case v != 0xFFFF:
				remaining++
			}
		}
		if !found {
			continue
		}
		for k := 0; k < mem.Len(); k++ {
			if mem.I(k).Int(img) == int64(i) {
				mem.I(k).SetInt(img, 0xFFFF)
			}
		}
		if remaining == 0 {
			used.SetInt(img, 0xFFFF)
		}
	}
}

func (f vx7Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	valid, used := f.flag(i, "valid").Bool(img), f.flag(i, "masked").Bool(img)
	if c == nil {
		// come CHIRP: il canale visibile viene nascosto, quello già nascosto liberato
		if valid && !used {
			f.flag(i, "valid").SetBool(img, false)
			return nil
		}
		f.flag(i, "masked").SetBool(img, false)
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
	k := ft60Tmode(t, rt)
	if k < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione o in entrambi i versi, DCS in " +
			"trasmissione o in entrambi i versi, subtono e DCS misti; mai il solo tono in ricezione o codici invertiti)")
	}
	fk, ok := yaesuKHz(c.Rx)
	if !ok {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtono 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	kept, tone, dcs, step, oldRx, oldMode := int64(600), int64(8), int64(0), int64(0), int64(0), int64(0)
	if valid {
		kept = r.F("offset").Int(img)
		if v := r.F("tone").Int(img); int(v) < len(chirpTones) {
			tone = v
		}
		if v := r.F("dcs").Int(img); int(v) < len(dtcsCodes) {
			dcs = v
		}
		if v := r.F("tune_step").Int(img); v < 9 {
			step = v
		}
		oldRx, oldMode = yaesuRound(r.F("freq").Int(img)*1000), r.F("mode").Int(img)
	}
	find := func(list []int, v int) (int64, error) {
		for k, x := range list {
			if x == v {
				return int64(k), nil
			}
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	if t.Mode == "Tone" {
		if tone, err = find(chirpTones, t.Hz10); err != nil {
			return err
		}
	}
	if rt.Mode == "Tone" && k == 7 {
		if tone, err = find(chirpTones, rt.Hz10); err != nil {
			return err
		}
	}
	for _, g := range []genTone{t, rt} {
		if g.Mode == "DTCS" {
			if dcs, err = find(dtcsCodes, g.Code); err != nil {
				return err
			}
		}
	}
	// modo dei toni: i tre misti hanno il segno «is_split_tone»
	split, tmode := false, k
	switch k {
	case 4:
		tmode = 3
	case 5, 6, 7:
		split, tmode = true, k-5
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
	mode := []int64{0, 2, 1}[icx90Mode(c.Rx)] // 0 FM, 1 AM, 2 FM larga, 3 automatico
	if valid && used && oldRx == c.Rx && oldMode != 0 {
		mode = oldMode
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !valid {
		r.Fill(img, 0)
		r.F("unknown1").SetInt(img, 0x05)
		r.F("ones").SetInt(img, 0x03)
		f.wipeBanks(img, i)
	}
	wasP := valid && used && f.flag(i, "pskip").Bool(img)
	f.flag(i, "valid").SetBool(img, true)
	f.flag(i, "masked").SetBool(img, true)
	r.F("freq").SetInt(img, fk)
	r.F("offset").SetInt(img, offs)
	r.F("tone").SetInt(img, tone)
	r.F("is_split_tone").SetBool(img, split)
	r.F("tmode").SetInt(img, tmode)
	r.F("duplex").SetInt(img, dup)
	r.F("mode").SetInt(img, mode)
	r.F("dcs").SetInt(img, dcs)
	r.F("tune_step").SetInt(img, step)
	p := 3 - int64(m.powerIndex(c.Power))
	if vx7Is220(c.Rx) && p > 1 {
		p = 1
	}
	r.F("power").SetInt(img, p)
	f.flag(i, "skip").SetBool(img, !c.Scan)
	f.flag(i, "pskip").SetBool(img, wasP && c.Scan)
	name := m.cleanName(c.Name)
	enc := make([]byte, 8)
	for k := range enc {
		ch := byte(' ')
		if k < len(name) {
			ch = name[k]
		}
		p := strings.IndexByte(vx7Chars, ch)
		if p < 0 {
			p = strings.IndexByte(vx7Chars, ' ')
		}
		enc[k] = byte(p)
	}
	r.F("name").SetRaw(img, enc)
	return nil // le somme di controllo si aggiornano quando si scrive nella radio (come CHIRP)
}

func (f vx7Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, err := yaesuDownload(ctx, m, port, f.yaesu(m), prog)
	return img, "", err
}

func (f vx7Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", yaesuUpload(ctx, m, port, f.yaesu(m), img, prog)
}
