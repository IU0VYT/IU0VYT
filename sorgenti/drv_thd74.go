// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// thd74.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: Kenwood TH-D74A/E e TH-D75A/E (cavo USB; la radio compare come porta
// seriale). 1000 canali con nomi di 16 caratteri, analogici o digitali D-STAR
// (DV / DR). Collegamento: «ID» (la radio dichiara il modello), «0M PROGRAM»,
// poi a 57600 baud blocchi da 256 byte: lettura «R blocco 0 0» → «W blocco 0 0»
// e i dati, confermati con 06; scrittura «W blocco 0 0» e i dati → 06; fine «E».
// Gli ultimi due blocchi della memoria non si scrivono (come CHIRP).
//
// Il modo (AM, SSB, CW) non è una colonna del programma: resta quello del canale
// se la frequenza non cambia, altrimenti il canale è FM (AM in banda aerea e
// sotto i 30 MHz). Un canale D-STAR nuovo nasce con la chiamata «CQCQCQ».

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

type thd74Fam struct{}

type thd74Opt struct {
	id   string
	size int  // byte della memoria (0 = quella della TH-D74)
	d72  bool // TH-D72: numero del blocco scritto al contrario, RTS attivo in programmazione
}

func (o *thd74Opt) bytes() int {
	if o.size != 0 {
		return o.size
	}
	return thd74Size
}

// intestazione di un blocco: comando e numero del blocco
func (o *thd74Opt) head(cmd byte, blk int) []byte {
	if o.d72 {
		return []byte{cmd, 0, byte(blk), byte(blk >> 8), 0}
	}
	return []byte{cmd, byte(blk >> 8), byte(blk), 0, 0}
}

const thd74Size = 0x7A300

const thd74Format = `
#seekto 0x2000;
struct {
  u8 used;
  u8 unknown1:7,
     lockout:1;
  u8 group;
  u8 unknownFF;
} flags[1200];

#seekto 0x4000;
struct memory {
    ul32 freq;
    ul32 offset;
    u8 tuning_step:4,
       split_tuning_step:3,
       unknown2:1;
    u8 unknown3_0:1,
       mode:3,
       narrow:1,
       fine_mode:1,
       fine_step:2;
    u8 tone_mode:1,
       ctcss_mode:1,
       dtcs_mode:1,
       cross_mode:1,
       unknown4_0:1,
       split:1,
       duplex:2;
    u8 rtone;
    u8 unknownctone:2,
       ctone:6;
    u8 unknowndtcs:1,
       dtcs_code:7;
    u8 unknown5_1:2,
       cross_mode_mode:2,
       unknown5_2:2,
       dig_squelch:2;
    char dv_urcall[8];
    char dv_rpt1call[8];
    char dv_rpt2call[8];
    u8 unknown9:1,
       dv_code:7;
};

struct {
  struct memory memories[6];
  u8 pad[16];
} memgroups[192];

struct {
  char name[16];
} names[1200];
`

const thdCableNote = "Cavo USB della radio (la radio compare come porta seriale); durante la copia la radio mostra «PROG MCP». "

func init() {
	mk := func(id, model, name, ident, rclass string) *genModel {
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: rclass, Name: name,
			Slots: 1000, First: 0, NameMax: 16, Powers: []genPower{{"Unica", 5, 0}}, Bands: []band{{100000, 470000000}},
			Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: thd74Fam{}, Opt: &thd74Opt{id: ident}, Digital: "D-STAR",
			Notes: "1000 canali (da 0 a 999) con nomi di 16 caratteri, analogici o D-STAR. " + thdCableNote +
				"Un canale D-STAR nuovo nasce con «CQCQCQ». Driver di CHIRP."}
	}
	genRegister(
		mk("kw-thd74", "TH-D74 (clone mode)", "Kenwood TH-D74A / TH-D74E", "TH-D74", "THD74Radio"),
		mk("kw-thd75", "TH-D75", "Kenwood TH-D75A / TH-D75E", "TH-D75", "THD75Radio"),
	)
}

func (thd74Fam) mem() *bwEl                 { return bwMust(thd74Format) }
func (f thd74Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f thd74Fam) rec(i int) *bwEl          { return f.mem().F("memgroups").I(i / 6).F("memories").I(i % 6) }
func (f thd74Fam) flag(i int) *bwEl         { return f.mem().F("flags").I(i) }
func (f thd74Fam) name(i int) *bwEl         { return f.mem().F("names").I(i).F("name") }

func (f thd74Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(i).Raw(img)...)
	out = append(out, f.flag(i).Raw(img)...)
	return append(out, f.name(i).Raw(img)...)
}

func (f thd74Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) != 40+4+16 {
		return
	}
	f.rec(i).SetRaw(img, raw[:40])
	f.flag(i).SetRaw(img, raw[40:44])
	f.name(i).SetRaw(img, raw[44:])
}

// Match: TH-D74 e TH-D75 hanno la memoria della stessa grandezza: i file si
// riconoscono dai metadati
func (thd74Fam) Match(m *genModel, img []byte) bool { return false }

// thd74Tmode: modo dei toni: 0 nessuno, 1 subtono in trasmissione, 2 subtono nei
// due versi, 3 DCS nei due versi, 4…7 misti (DCS in trasmissione; subtono e DCS;
// DCS e subtono; due subtoni diversi); -1 se la radio non ha la combinazione
func thd74Tmode(t, r genTone) int {
	switch {
	case t.Inv || r.Inv:
		return -1
	case t.Mode == "" && r.Mode == "":
		return 0
	case t.Mode == "Tone" && r.Mode == "":
		return 1
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
		return 2
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
		return 3
	case t.Mode == "DTCS" && r.Mode == "":
		return 4
	case t.Mode == "Tone" && r.Mode == "DTCS":
		return 5
	case t.Mode == "DTCS" && r.Mode == "Tone":
		return 6
	case t.Mode == "Tone" && r.Mode == "Tone":
		return 7
	}
	return -1
}

func (thd74Fam) toneOK(t, r genTone) bool { return thd74Tmode(t, r) >= 0 }
func (thd74Fam) fitBySet()                {}

func (thd74Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return thd74Tmode(t, r) < 0
}

func (f thd74Fam) Get(m *genModel, img []byte, i int) *Chan {
	fl, r := f.flag(i), f.rec(i)
	if fl.F("used").Int(img) == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("freq").Int(img)}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img)
	switch {
	case r.F("split").Bool(img):
		c.Tx = offs
	case r.F("duplex").Int(img) == 1:
		c.Tx = c.Rx + offs
	case r.F("duplex").Int(img) == 2:
		c.Tx = c.Rx - offs
	}
	mode := r.F("mode").Int(img)
	c.Dig = mode == 1 || mode == 7
	c.Narrow = mode == 6
	if !c.Dig {
		tone := func(k int64) string {
			if k >= 0 && int(k) < len(chirpTones) {
				return fmtTone(chirpTones[k])
			}
			return "88.5"
		}
		dcs := "D023N"
		if k := int(r.F("dtcs_code").Int(img)); k < len(dtcsCodes) {
			dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
		}
		rtone, ctone := tone(r.F("rtone").Int(img)), tone(r.F("ctone").Int(img))
		switch {
		case r.F("tone_mode").Bool(img):
			c.TxTone = rtone
		case r.F("ctcss_mode").Bool(img):
			c.TxTone, c.RxTone = ctone, ctone
		case r.F("dtcs_mode").Bool(img):
			c.TxTone, c.RxTone = dcs, dcs
		case r.F("cross_mode").Bool(img):
			switch r.F("cross_mode_mode").Int(img) {
			case 0:
				c.TxTone = dcs
			case 1:
				c.TxTone, c.RxTone = rtone, dcs
			case 2:
				c.TxTone, c.RxTone = dcs, ctone
			case 3:
				c.TxTone, c.RxTone = rtone, ctone
			}
		}
	}
	c.Scan = !fl.F("lockout").Bool(img)
	c.Name = strings.TrimRight(strings.Trim(strings.TrimRight(f.name(i).Str(img), " "), "\x00"), " ")
	return c
}

// passo di sintonia della frequenza di trasmissione separata (indice nell'elenco
// della radio); -1 se la radio non può memorizzarla
func thd74SplitStep(hz int64) int64 {
	switch {
	case hz%5000 == 0:
		return 0
	case hz%12500 == 0:
		return 5
	case hz%6250 == 0:
		return 1
	}
	switch hz % 25000 {
	case 8330, 8333, 16660, 16666:
		return 2
	}
	return -1
}

func (f thd74Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	fl, r, nm := f.flag(i), f.rec(i), f.name(i)
	wasEmpty := fl.F("used").Int(img) == 0xFF
	if c == nil {
		fl.F("used").SetInt(img, 0xFF)
		fl.F("lockout").SetInt(img, 0)
		fl.F("group").SetInt(img, 0)
		nm.Fill(img, 0)
		r.Fill(img, 0xFF)
		return nil
	}
	var t, rt genTone
	if !c.Dig {
		var err error
		if t, err = genToneParse(c.TxTone); err != nil {
			return err
		}
		if rt, err = genToneParse(c.RxTone); err != nil {
			return err
		}
	}
	tmode := thd74Tmode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (mai il solo tono in ricezione o codici invertiti)")
	}
	if c.Rx <= 0 || c.Rx > 0xFFFFFFFF {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtoni 88,5 Hz, DCS 023, scostamento 600 kHz, passo 5 kHz)
	rtone, ctone, dcs, offs, step, oldMode, oldRx := int64(8), int64(8), int64(0), int64(600000), int64(0), int64(0), int64(0)
	var calls []byte
	code := int64(0)
	if !wasEmpty {
		oldMode, oldRx = r.F("mode").Int(img), r.F("freq").Int(img)
		offs, step = r.F("offset").Int(img), r.F("tuning_step").Int(img)
		if oldMode == 1 || oldMode == 7 {
			calls = append([]byte{}, img[r.F("dv_urcall").Off:r.F("dv_urcall").Off+24]...)
			code = r.F("dv_code").Int(img)
		} else {
			rtone, ctone, dcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs_code").Int(img)
		}
	}
	if step > 11 {
		step = 0
	}
	find := func(list []int, v int) (int64, error) {
		for k, x := range list {
			if x == v {
				return int64(k), nil
			}
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	var err error
	switch tmode {
	case 1, 5, 7:
		rtone, err = find(chirpTones, t.Hz10)
	case 2:
		if rtone, err = find(chirpTones, t.Hz10); err == nil {
			ctone = rtone
		}
	}
	if err == nil && (tmode == 6 || tmode == 7) {
		ctone, err = find(chirpTones, rt.Hz10)
	}
	if err == nil && (tmode == 3 || tmode == 4 || tmode == 6) {
		dcs, err = find(dtcsCodes, t.Code)
	}
	if err == nil && tmode == 5 {
		dcs, err = find(dtcsCodes, rt.Code)
	}
	if err != nil {
		return err
	}
	dup, split, sstep := int64(0), false, int64(0)
	band := c.Rx
	d := c.Tx - c.Rx
	switch {
	case d == 0:
	case d > 70000000 || d < -70000000:
		if sstep = thd74SplitStep(c.Tx); sstep < 0 || c.Tx > 0xFFFFFFFF {
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
		}
		split, offs, band = true, c.Tx, c.Tx
	case d > 0:
		dup, offs = 1, d
	default:
		dup, offs = 2, -d
	}
	// modo: 0 FM, 1 DV, 2 AM, 6 FM stretta, 7 DR; gli altri (SSB, CW) restano se
	// la frequenza non cambia
	mode := int64(0)
	switch {
	case c.Dig:
		mode = 1
		if oldMode == 7 {
			mode = 7
		}
	case !wasEmpty && oldRx == c.Rx && oldMode >= 2 && oldMode <= 5:
		mode = oldMode
	case icx90Mode(c.Rx) == 2:
		mode = 2
	case c.Narrow:
		mode = 6
	}
	if stale := r.F("mode").Int(img); calls == nil && (c.Dig || wasEmpty && (stale == 1 || stale == 7)) {
		// canale D-STAR nuovo: chiamata generale, nessun ripetitore (CHIRP la scrive
		// anche nei canali analogici nuovi, dove non ha effetto)
		calls = make([]byte, 24)
		copy(calls, "CQCQCQ")
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	used := int64(2)
	switch {
	case band < 150000000:
		used = 0
	case band < 400000000:
		used = 1
	}
	fl.F("used").SetInt(img, used)
	r.Fill(img, 0)
	if wasEmpty {
		fl.F("group").SetInt(img, 0)
	}
	r.F("freq").SetInt(img, c.Rx)
	nm.SetStr(img, m.cleanName(c.Name), ' ')
	r.F("offset").SetInt(img, offs)
	r.F("split").SetBool(img, split)
	r.F("duplex").SetInt(img, dup)
	r.F("split_tuning_step").SetInt(img, sstep)
	r.F("tuning_step").SetInt(img, step)
	r.F("mode").SetInt(img, mode)
	r.F("narrow").SetBool(img, mode == 6)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("dtcs_code").SetInt(img, dcs)
	r.F("tone_mode").SetBool(img, tmode == 1)
	r.F("ctcss_mode").SetBool(img, tmode == 2)
	r.F("dtcs_mode").SetBool(img, tmode == 3)
	r.F("cross_mode").SetBool(img, tmode >= 4)
	if tmode >= 4 {
		r.F("cross_mode_mode").SetInt(img, int64(tmode-4))
	}
	fl.F("lockout").SetBool(img, !c.Scan)
	if calls != nil {
		copy(img[r.F("dv_urcall").Off:], calls)
		r.F("dv_code").SetInt(img, code)
	}
	return nil
}

// ---------------------------------------------------------------- collegamento

func thd74Line(s *rawSess, cmd string, to time.Duration) (string, error) {
	if err := s.send([]byte(cmd + "\r")); err != nil {
		return "", err
	}
	var buf []byte
	deadline := time.Now().Add(to)
	for !bytes.HasSuffix(buf, []byte("\r")) && time.Now().Before(deadline) {
		b := readUpTo(s.ctx, s.p, 1, time.Until(deadline))
		if s.ctx.Err() != nil {
			return "", errCancelled
		}
		buf = append(buf, b...)
	}
	return strings.TrimSpace(string(buf)), nil
}

// enter: la radio dichiara il modello (a qualunque velocità) ed entra in
// programmazione; da lì in poi si parla a 57600 baud
func thd74Enter(s *rawSess, m *genModel) error {
	id := ""
	for _, baud := range []int{9600, 19200, 38400, 57600} {
		if err := s.p.SetBaud(baud); err != nil {
			return err
		}
		if err := s.send([]byte("\r\r")); err != nil {
			return err
		}
		readUpTo(s.ctx, s.p, 32, 250*time.Millisecond)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		r, err := thd74Line(s, "ID", time.Second)
		if err != nil {
			return err
		}
		if strings.HasPrefix(r, "ID ") {
			id = strings.TrimSpace(r[3:])
			break
		}
	}
	if id == "" {
		return rErr("la radio non risponde")
	}
	if want := m.Opt.(*thd74Opt).id; id != want {
		for _, x := range genModels {
			if o, ok := x.Opt.(*thd74Opt); ok && o.id == id {
				return rErr("la radio è un altro modello: apri la scheda «%s»", x.Name)
			}
		}
		return rErr("la radio è un altro modello (%s)", id)
	}
	r, err := thd74Line(s, "0M PROGRAM", time.Second)
	if err != nil {
		return err
	}
	if r == "" {
		return rErr("la radio non risponde")
	}
	if r != "0M" {
		return rErr("risposta inattesa della radio")
	}
	if err := s.p.SetBaud(57600); err != nil {
		return err
	}
	if m.Opt.(*thd74Opt).d72 {
		if x, ok := s.p.(interface{ SetRTS(bool) error }); ok {
			_ = x.SetRTS(true)
		}
	}
	readUpTo(s.ctx, s.p, 1, 250*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	return nil
}

func (f thd74Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	return thd74Download(ctx, m, port, prog)
}

func (f thd74Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return thd74Upload(ctx, m, port, img, prog)
}

func thd74Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := m.Opt.(*thd74Opt)
	size := o.bytes()
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, size, "Collegamento con la radio…")
	if err := thd74Enter(s, m); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	defer s.send([]byte("E"))
	img := make([]byte, 0, size)
	for blk := 0; blk < size/256; blk++ {
		if err := s.send(o.head('R', blk)); err != nil {
			return nil, "", err
		}
		h := readUpTo(ctx, s.p, 5, 2*time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(h) != 5 {
			return nil, "", rErr("la radio ha smesso di inviare dati prima della fine (0x%04X)", blk*256)
		}
		if !bytes.Equal(h, o.head('W', blk)) {
			return nil, "", rErr("blocco della radio non valido (0x%04X)", blk*256)
		}
		d := readUpTo(ctx, s.p, 256, 3*time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(d) != 256 {
			return nil, "", rErr("la radio ha smesso di inviare dati prima della fine (0x%04X)", blk*256)
		}
		img = append(img, d...)
		if err := s.send([]byte{0x06}); err != nil {
			return nil, "", err
		}
		if a := readUpTo(ctx, s.p, 1, 2*time.Second); len(a) != 1 || a[0] != 0x06 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("la radio non ha confermato il blocco 0x%04X", blk*256)
		}
		prog(len(img), size, "Lettura…")
	}
	prog(size, size, "Lettura completata")
	return img, "", nil
}

func thd74Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := m.Opt.(*thd74Opt)
	size := o.bytes()
	if len(img) != size {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, size, "Collegamento con la radio…")
	if err := thd74Enter(s, m); err != nil {
		return "", genFail("Identificazione", err)
	}
	defer s.send([]byte("E"))
	n := size/256 - 2
	for blk := 0; blk < n; blk++ {
		pkt := append(o.head('W', blk), img[blk*256:blk*256+256]...)
		if err := s.send(pkt); err != nil {
			return "", err
		}
		ok := false
		for try := 0; try < 10 && !ok; try++ {
			a := readUpTo(ctx, s.p, 1, time.Second)
			if ctx.Err() != nil {
				return "", errCancelled
			}
			ok = len(a) == 1 && a[0] == 0x06
		}
		if !ok {
			return "", rErr("la radio non ha confermato il blocco 0x%04X", blk*256)
		}
		prog((blk+1)*256, size, "Scrittura…")
	}
	prog(size, size, "Scrittura completata")
	return "", nil
}
