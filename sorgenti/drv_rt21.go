// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: retevis_rt21.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.2: famiglia Retevis RT21 (driver retevis_rt21.py di CHIRP), 19 modelli:
// Retevis RT21, RT21V, RB17A, RB23, RB26, RB626, RB89, RT76, RT29 (UHF e VHF),
// RT19, RT619, RT40B, RB28B, RB628B, RT86, RT86S, RT86A e Abbree AR-63.
// Radio senza nomi dei canali; canali di 16 byte (frequenze BCD, toni a 16 bit:
// CTCSS in decimi di Hz, DCS con 0x2000 + 0x0800 + codice in ottale, 0x8000 se
// invertito). Ogni modello ha la sua mappa (MEM_FORMAT… di retevis_rt21.py).
//
// Protocollo (_enter_programming_mode, _read_block, _write_block): comando del
// modello (es. «PRMZUNE», «PHOGR\x01\x30») → 0x06 (fino a 4 byte letti); 0x02 →
// identificativo di 8 byte; 0x06 → 0x06; lettura: 'R' indirizzo(2, BE) n → 'W'
// indirizzo n + n byte, poi 0x06 → 0x06 (non dopo il primo blocco sui modelli
// recenti); scrittura a blocchi di 16 byte: 'W' indirizzo 0x10 + 16 byte → 0x06;
// fine: 'E'. Sulla RT40B il cavo restituisce i byte inviati.
//
// Canali fissi come in CHIRP: FRS (RB28B e RT19, canali 1-22), PMR446 (RB628B e
// RT619, canali 1-16), GMRS (RB17A, RB23, RB26, RT76, RT40B, canali 1-30).

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type rt21Opt struct {
	model    string // MODEL di CHIRP (le scelte di retevis_rt21.py dipendono dal modello)
	format   string
	spec     *rbSpec
	skip     bool   // mappa di bit dei canali in scansione (_skipflags)
	reserved bool   // ultimi 3 byte del canale conservati (_reserved)
	lock     string // "frs", "pmr", "gmrs": canali fissi
	power    string // "" bit highpower; "rt29" campo txpower; "rt86s" tabella levels
}

type rt21Fam struct{}

func rt21o(m *genModel) *rt21Opt { return m.Opt.(*rt21Opt) }

// frequenze fisse (FRS_FREQS, GMRS_FREQS, PMR_FREQS di retevis_rt21.py)
var (
	rt21FRS  = append(append(append([]int64{}, t18FRS1...), t18FRS2...), t18FRS3...)
	rt21GMRS = append(append([]int64{}, rt21FRS...), t18FRS3...)
	rt21PMR  = t18PMR
)

const rt21RecFields = `
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rx_tone;
  ul16 tx_tone;
`

func rt21Formats() map[string]func(n int) string {
	rec := func(tail string) string { return "struct {" + rt21RecFields + tail + "\n}" }
	return map[string]func(int) string{
		"rt21": func(n int) string {
			return "#seekto 0x0010;\n" + rec(`  u8 unknown1:3,
     bcl:2,
     unknown2:3;
  u8 unknown3:2,
     highpower:1,
     wide:1,
     unknown4:4;
  u8 unknown7:1,
     scramble_type:3,
     unknown5:4;
  u8 unknown6:5,
     scramble_type2:3;`) + fmt.Sprintf(" memory[%d];\n#seekto 0x017E;\nu8 skipflags[2];\n", n)
		},
		"rb17a": func(int) string {
			return "struct memory {" + rt21RecFields + `  u8 unknown1:1,
     compander:1,
     bcl:2,
     cdcss:1,
     scramble_type:3;
  u8 unknown2:4,
     middlepower:1,
     unknown3:1,
     highpower:1,
     wide:1;
  u8 unknown4;
  u8 unknown5;
};
#seekto 0x0010;
struct memory lomems[16];
#seekto 0x0200;
struct memory himems[14];
#seekto 0x017E;
u8 skipflags[4];
`
		},
		"rb26": func(n int) string {
			return "#seekto 0x0000;\n" + rec(`  u8 compander:1,
     unknown1:1,
     highpower:1,
     wide:1,
     bcl:1,
     unknown2:3;
  u8 reserved[3];`) + fmt.Sprintf(" memory[%d];\n", n)
		},
		"rt76": func(n int) string {
			return "#seekto 0x0000;\n" + rec(`  u8 compander:1,
     hop:1,
     highpower:1,
     wide:1,
     scramble:4;
  u8 reserved[3];`) + fmt.Sprintf(" memory[%d];\n#seekto 0x004E;\nu8 skipflags[2];\n", n)
		},
		"rt29": func(int) string {
			return "#seekto 0x0010;\n" + rec(`  u8 unknown1:2,
     compander:1,
     bcl:2,
     unknown2:3;
  u8 unknown3:1,
     txpower:2,
     wide:1,
     unknown4:3,
     cdcss:1;
  u8 unknown5;
  u8 unknown6:5,
     scramble_type:3;`) + " memory[16];\n#seekto 0x017E;\nu8 skipflags[2];\n"
		},
		"rt19": func(n int) string {
			hop := 0x160
			if n == 16 {
				hop = 0x100
			}
			return "#seekto 0x0000;\n" + rec(`  u8 function:2,
     highpower:1,
     wide:1,
     unknown_1:1,
     scramble_type:3;
  u8 reserved[3];`) + fmt.Sprintf(" memory[%d];\n#seekto 0x%X;\nstruct {\n  u8 freqhop;\n} freqhops[%d];\n", n, hop, n)
		},
		"rt40b": func(n int) string {
			return "#seekto 0x0000;\n" + rec(`  u8 compander:1,
     unknown1:1,
     highpower:1,
     wide:1,
     unknown2:4;
  u8 reserved[3];`) + fmt.Sprintf(" memory[%d];\n#seekto 0x00AD;\nu8 skipflags[3];\n", n)
		},
		"rb28b": func(int) string {
			return "#seekto 0x0000;\n" + rec(`  u8 compander:1,
     unused_1:1,
     highpower:1,
     wide:1,
     bcl:1,
     unused_2:3;
  u8 reserved[3];`) + " memory[30];\n"
		},
		"rt86": func(n int) string {
			s := "#seekto 0x0000;\n" + rec(`  u8 audio:2,
     highpower:1,
     wide:1,
     bcl:1,
     scramble_type:3;
  u8 reserved[3];`) + fmt.Sprintf(" memory[%d];\n", n)
			if n == 16 {
				return s + "#seekto 0x004E;\nu8 skipflags[2];\n#seekto 0x0100;\nstruct {\n  u8 freqhop;\n} freqhops[16];\n"
			}
			return s + "#seekto 0x01F0;\nu8 skipflags[4];\n#seekto 0x0100;\nstruct {\n  u8 freqhop;\n} freqhops[16];\n" +
				"#seekto 0x0330;\nstruct {\n  u8 txpower;\n} levels[30];\n"
		},
	}
}

func init() {
	fm := rt21Formats()
	type v struct {
		id, vendor, model, rclass, name string
		slots                           int
		format                          string
		spec                            *rbSpec
		pw                              []genPower
		bands                           []band
		dcs                             []int
		skip, reserved                  bool
		lock, power, note               string
	}
	uhf := []band{{400000000, 480000000}}
	p2 := func(h, l float64) []genPower {
		return []genPower{{fmt.Sprintf("Alta %s W", wStr(h)), h, 0}, {fmt.Sprintf("Bassa %s W", wStr(l)), l, 2}}
	}
	p3 := func(h, md, l float64) []genPower {
		return []genPower{{fmt.Sprintf("Alta %s W", wStr(h)), h, 0}, {fmt.Sprintf("Media %s W", wStr(md)), md, 1}, {fmt.Sprintf("Bassa %s W", wStr(l)), l, 2}}
	}
	base := withCodes(dtcsCodes, 17, 50, 645)
	extra := withCodes(dtcsCodes, 645)
	rt29dcs := withCodes(dtcsCodes, 17, 50, 55, 135, 217, 254, 305, 345, 425, 466, 534, 645, 765)
	sp := func(magic string, finger []string, block int, ack1st bool, memsize int, end int) *rbSpec {
		var fs [][]byte
		for _, f := range finger {
			fs = append(fs, []byte(f))
		}
		return &rbSpec{magic: []byte(magic), fingers: fs, block: block, wblock: 0x10, noAckFirst: !ack1st, ackReads: 4,
			memsize: memsize, ranges: [][2]int{{0, end}}, tries: 1, to: time.Second, quiet: true}
	}
	f2 := []string{"P32073\x02\xFF"}
	list := []v{
		{"rt-rt21", "Retevis", "RT21", "RT21Radio", "Retevis RT21", 16, fm["rt21"](16), sp("PRMZUNE", []string{"P3207s\xF8\xFF"}, 0x10, true, 0x400, 0x400),
			p2(2.5, 1), uhf, base, true, false, "", "", ""},
		{"rt-rt21v", "Retevis", "RT21V", "RT21VRadio", "Retevis RT21V", 5, fm["rt21"](5), sp("PRMZUNE", []string{"P2207\x01\xF8\xFF"}, 0x10, true, 0x400, 0x400),
			p2(2, 0.5), []band{{137000000, 174000000}}, base, true, false, "", "", " Versione VHF (venduta per la banda MURS), 5 canali."},
		{"rt-rb17a", "Retevis", "RB17A", "RB17ARadio", "Retevis RB17A", 30, fm["rb17a"](30), sp("PROA8US", []string{"P3217s\xF8\xFF"}, 0x40, true, 0x300, 0x300),
			p2(5, 0.5), uhf, base, true, false, "gmrs", "", ""},
		{"rt-rb23", "Retevis", "RB23", "RB23Radio", "Retevis RB23", 30, fm["rb26"](30) + "#seekto 0x01F0;\nu8 skipflags[4];\n", sp("PHOGR\x010", f2, 0x20, false, 0x320, 0x320),
			p2(5, 0.5), uhf, base, true, true, "gmrs", "", ""},
		{"rt-rb26", "Retevis", "RB26", "RB26Radio", "Retevis RB26", 30, fm["rb26"](30) + "#seekto 0x01F0;\nu8 skipflags[4];\n", sp("PHOGR\x010", f2, 0x20, false, 0x800, 0x320),
			p2(3, 0.5), uhf, extra, true, true, "gmrs", "", ""},
		{"rt-rb626", "Retevis", "RB626", "RB626", "Retevis RB626", 32, fm["rb26"](32) + "#seekto 0x0200;\nu8 skipflags[4];\n", sp("PHOGR\x010", f2, 0x20, false, 0x800, 0x340),
			p2(3, 0.5), uhf, extra, true, true, "", "", ""},
		{"rt-rb89", "Retevis", "RB89", "RB89Radio", "Retevis RB89", 30, fm["rb26"](30) + "#seekto 0x01F0;\nu8 skipflags[4];\n", sp("PHOGR\x010", []string{"P32073\x01\xFF"}, 0x20, false, 0x340, 0x330),
			p2(5, 0.5), uhf, base, true, true, "", "", " Venduta come radio GMRS, ma CHIRP la gestisce senza canali fissi."},
		{"rt-rt76", "Retevis", "RT76", "RT76Radio", "Retevis RT76", 30, strings.Replace(fm["rt76"](30), "#seekto 0x004E;\nu8 skipflags[2];\n", "", 1), sp("PHOGR\x14\xD4", f2, 0x20, false, 0x1E0, 0x1E0),
			p2(5, 0.5), uhf, extra, false, true, "gmrs", "", ""},
		{"rt-rt29u", "Retevis", "RT29_UHF", "RT29UHFRadio", "Retevis RT29 UHF", 16, fm["rt29"](16), sp("PROHRAM", []string{"P3207\x13\xF8\xFF", "P3207s\xF8\xFF"}, 0x40, true, 0x400, 0x300),
			p3(10, 5, 1), uhf, rt29dcs, true, false, "", "rt29", ""},
		{"rt-rt29v", "Retevis", "RT29_VHF", "RT29VHFRadio", "Retevis RT29 VHF", 16, fm["rt29"](16), sp("PROHRAM", []string{"P2207\x01\xF8\xFF"}, 0x40, true, 0x400, 0x300),
			p3(10, 5, 1), []band{{136000000, 174000000}}, rt29dcs, true, false, "", "rt29", ""},
		{"rt-rt19", "Retevis", "RT19", "RT19Radio", "Retevis RT19", 22, fm["rt19"](22), sp("PHOGRQ^", f2, 0x20, false, 0x180, 0x180),
			p2(2, 0.5), uhf, base, false, true, "frs", "", ""},
		{"rt-rt619", "Retevis", "RT619", "RT619Radio", "Retevis RT619", 16, fm["rt19"](16), sp("PHOGRS]", f2, 0x20, false, 0x120, 0x120),
			p2(0.5, 0.49), uhf, base, false, true, "pmr", "", ""},
		{"ab-ar63", "Abbree", "AR-63", "AR63Radio", "Abbree AR-63", 16, fm["rt76"](16), sp("PHOGR\xF5\x9A", []string{"P32073\x02\xFF", "P32073\x03\xFF"}, 0x20, false, 0x140, 0x140),
			p2(3, 1), uhf, base, true, true, "", "", ""},
		{"rt-rt40b", "Retevis", "RT40B", "RT40BRadio", "Retevis RT40B", 22, fm["rt40b"](22), sp("PHOGRH\x5C", f2, 0x20, false, 0x160, 0x160),
			p2(2, 0.5), uhf, extra, true, true, "gmrs", "", ""},
		{"rt-rb28b", "Retevis", "RB28B", "RB28BRadio", "Retevis RB28B", 22, fm["rb28b"](22), sp("PHOGR\x08\xB2", f2, 0x20, false, 0x1F0, 0x1F0),
			p2(2, 0.5), uhf, extra, false, true, "frs", "", ""},
		{"rt-rb628b", "Retevis", "RB628B", "RB628BRadio", "Retevis RB628B", 16, fm["rb28b"](16), sp("PHOGR\x09\xB2", f2, 0x20, false, 0x1F0, 0x1F0),
			p2(0.5, 0.5), uhf, extra, false, true, "pmr", "", ""},
		{"rt-rt86", "Retevis", "RT86", "RT86Radio", "Retevis RT86", 16, fm["rt86"](16), sp("PHOGR\xCD\x91", f2, 0x20, false, 0x1A0, 0x1A0),
			p2(2, 0.5), uhf, extra, true, true, "", "", ""},
		{"rt-rt86s", "Retevis", "RT86S", "RT86SRadio", "Retevis RT86S", 30, fm["rt86"](30), sp("PHOGR\xA7\xB8", []string{"P32073\x00\xFF", "P32073\x03\xFF"}, 0x20, false, 0x360, 0x350),
			p3(2, 1, 0.5), []band{{400000000, 520000000}}, extra, true, true, "", "rt86s", ""},
		// RT86A: in CHIRP la potenza usa il solo bit «alta» (le scelte di
		// retevis_rt21.py controllano il nome «RT86S»): alta o media
		{"rt-rt86a", "Retevis", "RT86A", "RT86ARadio", "Retevis RT86A", 30, fm["rt86"](30), sp("PHOGR6\x83", f2, 0x20, false, 0x360, 0x350),
			p3(2, 1, 0.5)[:2], []band{{400000000, 520000000}}, extra, true, true, "", "", " Potenza alta o media (come CHIRP)."},
	}
	for _, x := range list {
		if x.id == "rt-rt40b" {
			x.spec.echo = true
		}
		note := "Accendi la radio con il cavo collegato. Radio senza nomi dei canali (driver di CHIRP della famiglia Retevis RT21)."
		switch x.lock {
		case "frs":
			note += " Canali 1-22 FRS fissi: frequenza e banda non modificabili (come CHIRP)."
		case "pmr":
			note += " Canali 1-16 PMR446 fissi: frequenza, banda e potenza non modificabili (come CHIRP)."
		case "gmrs":
			note += fmt.Sprintf(" Canali 1-%d GMRS fissi: frequenza non modificabile (come CHIRP).", min(x.slots, 30))
		}
		genRegister(&genModel{ID: x.id, Vendor: x.vendor, Model: x.model, RClass: x.rclass, Name: x.name, Slots: x.slots, First: 1,
			Powers: x.pw, Bands: x.bands, Narrow: true, HasScan: x.skip, DCS: x.dcs, Baud: 9600, Fam: rt21Fam{},
			Opt:   &rt21Opt{model: x.model, format: x.format, spec: x.spec, skip: x.skip, reserved: x.reserved, lock: x.lock, power: x.power},
			Notes: note + x.note})
	}
}

// wStr: watt senza zeri inutili, con la virgola
func wStr(w float64) string {
	return strings.Replace(strconv.FormatFloat(w, 'f', -1, 64), ".", ",", 1)
}

func (rt21Fam) mem(m *genModel) *bwEl { return bwMust(rt21o(m).format) }

func (f rt21Fam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f rt21Fam) rec(m *genModel, i int) *bwEl {
	b := f.mem(m)
	if rt21o(m).model == "RB17A" {
		if i < 16 {
			return b.F("lomems").I(i)
		}
		return b.F("himems").I(i - 16)
	}
	return b.F("memory").I(i)
}

func (f rt21Fam) skipOff(m *genModel, i int) (int, byte) {
	return f.mem(m).F("skipflags").Off + i/8, byte(1) << (i % 8)
}

// record del canale, poi (se esistono) il bit di scansione e il salto di frequenza
func (f rt21Fam) Rec(m *genModel, img []byte, i int) []byte {
	o := rt21o(m)
	out := append([]byte{}, f.rec(m, i).Raw(img)...)
	if o.skip {
		off, b := f.skipOff(m, i)
		out = append(out, img[off]&b)
	}
	if f.mem(m).Has("freqhops") && i < f.mem(m).F("freqhops").Len() {
		out = append(out, f.mem(m).F("freqhops").I(i).Raw(img)...)
	}
	return out
}

func (f rt21Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	o := rt21o(m)
	f.rec(m, i).SetRaw(img, raw[:16])
	k := 16
	if o.skip {
		off, b := f.skipOff(m, i)
		if raw[k] != 0 {
			img[off] |= b
		} else {
			img[off] &^= b
		}
		k++
	}
	if f.mem(m).Has("freqhops") && i < f.mem(m).F("freqhops").Len() && k < len(raw) {
		f.mem(m).F("freqhops").I(i).SetRaw(img, raw[k:k+1])
	}
}

func (f rt21Fam) Match(m *genModel, img []byte) bool { return false }

// _get_dcs: 0x0800 = codice in ottale, altrimenti in esadecimale (se non è un
// codice valido si legge in ottale, come CHIRP per le memorie scritte male)
func rt21DCS(v int64) genTone {
	c := v & 0x07FF
	var code int
	if v&0x0800 != 0 {
		code, _ = strconv.Atoi(fmt.Sprintf("%03o", c))
	} else {
		n, err := strconv.Atoi(fmt.Sprintf("%03x", c))
		if err != nil || idxInt(dtcsCodes, n) < 0 {
			n, _ = strconv.Atoi(fmt.Sprintf("%03o", c))
		}
		code = n
	}
	return genTone{Mode: "DTCS", Code: code, Inv: v&0x8000 != 0}
}

func rt21ToneDec(v int64) string {
	switch {
	case v == 0xFFFF:
		return ""
	case v > 0x2000:
		return rt21DCS(v).String()
	}
	if v == 0 {
		return "0.0"
	}
	return fmtTone(int(v))
}

func rt21ToneEnc(t genTone) (int64, error) {
	switch t.Mode {
	case "Tone":
		return int64(t.Hz10), nil
	case "DTCS":
		v, err := strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		if err != nil {
			return 0, fmt.Errorf("codice DCS %03d non valido", t.Code)
		}
		v |= 0x2000 | 0x0800
		if t.Inv {
			v |= 0x8000
		}
		return v, nil
	}
	return 0xFFFF, nil
}

// rt21Fixed: canale fisso (CHIRP: immutable) e campi bloccati
func rt21Fixed(o *rt21Opt, n int) (fixed, dup, mode, power bool) {
	switch o.lock {
	case "frs":
		if n >= 1 && n <= 22 {
			return true, true, true, n >= 8 && n <= 14
		}
	case "pmr":
		if n >= 1 && n <= 16 {
			return true, true, true, true
		}
	case "gmrs":
		if n >= 1 && n <= 30 {
			return true, true, n >= 8 && n <= 14, n >= 8 && n <= 14
		}
	}
	return false, false, false, false
}

func (f rt21Fam) lockedSlot(m *genModel, i int) bool {
	fx, _, _, _ := rt21Fixed(rt21o(m), i+1)
	return fx
}

// Locked: i canali fissi non si cancellano e non cambiano frequenza (come
// CHIRP). Una posizione fissa vuota si può programmare solo con la sua frequenza:
// CHIRP (e la radio) mostrerebbero comunque quella.
func (f rt21Fam) Locked(m *genModel, i int, old, want *Chan) error {
	o := rt21o(m)
	fx, dup, mode, power := rt21Fixed(o, i+1)
	if !fx {
		return nil
	}
	what := map[string]string{"frs": "FRS", "pmr": "PMR446", "gmrs": "GMRS"}[o.lock]
	if want == nil {
		if old != nil {
			return fmt.Errorf("il canale %d della %s è un canale %s fisso: non si può cancellare", i+1, m.Name, what)
		}
		return nil
	}
	if old != nil && dup && (want.TxOff != old.TxOff || (!want.TxOff && want.Tx != old.Tx)) {
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza di trasmissione non modificabile", i+1, m.Name, what)
	}
	// come lo rileggerebbe CHIRP
	buf := bytes.Repeat([]byte{0xFF}, genMinSize(m))
	if err := f.Set(m, buf, i, want); err != nil {
		return err
	}
	back := f.Get(m, buf, i)
	switch {
	case back == nil || back.Rx != want.Rx:
		fixedRx := rt21FRS
		if o.lock == "pmr" {
			fixedRx = rt21PMR
		} else if o.lock == "gmrs" {
			fixedRx = rt21GMRS
		}
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza %s MHz non modificabile", i+1, m.Name, what, fmtMHz(fixedRx[i]))
	case dup && (back.Tx != want.Tx || back.TxOff != want.TxOff):
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: frequenza di trasmissione non modificabile", i+1, m.Name, what)
	case mode && back.Narrow != want.Narrow:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: sempre a banda stretta", i+1, m.Name, what)
	case power && back.Power != want.Power:
		return fmt.Errorf("il canale %d della %s è un canale %s fisso: potenza non modificabile", i+1, m.Name, what)
	}
	return nil
}

func (f rt21Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := rt21o(m)
	r := f.rec(m, i)
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 || bytes.Equal(r.F("rxfreq").Raw(img), ff4) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	txv := r.F("txfreq").Int(img) * 10
	switch {
	case txv == rx:
	case bytes.Equal(r.F("txfreq").Raw(img), ff4):
		c.TxOff = true
	default:
		c.Tx = txv
	}
	c.Narrow = !r.F("wide").Bool(img)
	c.TxTone = rt21ToneDec(r.F("tx_tone").Int(img))
	c.RxTone = rt21ToneDec(r.F("rx_tone").Int(img))
	pidx := 0 // CHIRP: livello non riconosciuto (nessuna potenza): la prima
	switch o.power {
	case "rt29":
		pidx = map[int64]int{1: 0, 0: 1, 2: 2}[r.F("txpower").Int(img)]
	case "rt86s":
		pidx = map[int64]int{2: 0, 1: 1, 0: 2}[f.mem(m).F("levels").I(i).F("txpower").Int(img)]
	default:
		pidx = 1 - int(r.F("highpower").Int(img))
	}
	c.Power = m.powerSem(pidx)
	c.Scan = true
	if o.skip {
		off, b := f.skipOff(m, i)
		c.Scan = img[off]&b != 0
	}
	// canali fissi: come li mostra CHIRP (frequenza della posizione; il verso dello
	// spostamento resta quello letto)
	n := i + 1
	dir := int64(0)
	if !c.TxOff && c.Tx != c.Rx {
		dir = 1
		if c.Tx < c.Rx {
			dir = -1
		}
	}
	switch o.lock {
	case "frs":
		if n <= 22 {
			c.Rx, c.Tx, c.Narrow = rt21FRS[n-1], rt21FRS[n-1], true
			if n >= 8 && n <= 14 {
				c.Power = m.powerSem(1)
			}
		}
	case "pmr":
		if n <= 16 {
			c.Rx, c.Tx, c.TxOff, c.Narrow = rt21PMR[n-1], rt21PMR[n-1], false, true
			c.Power = m.powerSem(1)
		}
	case "gmrs":
		if n <= 30 {
			c.Rx, c.Tx = rt21GMRS[n-1], rt21GMRS[n-1]
			if n >= 8 && n <= 14 {
				c.Narrow = true
				c.Power = m.powerSem(1)
			}
			if n >= 23 {
				c.Tx = c.Rx + dir*5000000
			}
		}
	}
	return c
}

// impostazioni del canale lette e riscritte da CHIRP (mem.extra) per modello
func rt21Extras(model string) []string {
	switch model {
	case "RT21", "RT21V":
		return []string{"bcl", "scramble_type"}
	case "RB17A", "RT29_UHF", "RT29_VHF":
		return []string{"bcl", "scramble_type", "cdcss", "compander"}
	case "RB23", "RB26", "RB89", "RB626":
		return []string{"bcl", "compander"}
	case "RT76", "RT40B":
		return []string{"compander"}
	case "AR-63":
		return []string{"compander", "scramble", "hop"}
	case "RT19", "RT619":
		return []string{"function", "scramble_type"}
	case "RB28B", "RB628B":
		return []string{"compander", "bcl"}
	case "RT86":
		return []string{"audio", "scramble_type", "bcl"}
	}
	return nil
}

func (f rt21Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := rt21o(m)
	r := f.rec(m, i)
	raw := r.Raw(img)
	rsvd := append([]byte{}, raw[13:16]...)
	byteC := raw[12]
	if c == nil {
		switch o.model {
		case "RB23", "RB26", "RB626", "RT40B", "RT76", "RT86", "AR-63":
			r.SetRaw(img, append(bytes.Repeat([]byte{0xFF}, 13), rsvd...))
		case "RT19", "RT619":
			r.SetRaw(img, append(bytes.Repeat([]byte{0xFF}, 13), rsvd...))
			f.mem(m).F("freqhops").I(i).SetRaw(img, []byte{0})
		default:
			r.Fill(img, 0xFF)
		}
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
	txv, err := rt21ToneEnc(tt)
	if err != nil {
		return err
	}
	rxv, err := rt21ToneEnc(rt)
	if err != nil {
		return err
	}
	was := f.Get(m, img, i) != nil
	keep := map[string]int64{}
	for _, k := range rt21Extras(o.model) {
		keep[k] = r.F(k).Int(img)
	}
	switch {
	case o.model == "RB17A":
		r.SetRaw(img, append(make([]byte, 14), 0xFF, 0xFF))
	case o.reserved && o.model == "RT86S":
		r.SetRaw(img, append(append(make([]byte, 12), byteC), rsvd...))
	case o.reserved:
		r.SetRaw(img, append(make([]byte, 13), rsvd...))
	default:
		r.SetRaw(img, append(make([]byte, 13), 0x30, 0x8F, 0xF8))
	}
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("wide").SetBool(img, !c.Narrow)
	r.F("rx_tone").SetInt(img, rxv)
	r.F("tx_tone").SetInt(img, txv)
	pidx := m.powerIndex(c.Power)
	switch o.power {
	case "rt29":
		r.F("txpower").SetInt(img, []int64{1, 0, 2}[pidx])
	case "rt86s":
		f.mem(m).F("levels").I(i).F("txpower").SetInt(img, []int64{2, 1, 0}[pidx])
	default:
		// CHIRP: mem.power == POWER_LEVELS[0] confronta i watt (RB628B: due
		// livelli da 0,5 W, sempre «alta»)
		r.F("highpower").SetBool(img, m.Powers[pidx].Watts == m.Powers[0].Watts)
	}
	if o.skip {
		off, b := f.skipOff(m, i)
		if c.Scan {
			img[off] |= b
		} else {
			img[off] &^= b
		}
	}
	if was {
		for k, v := range keep {
			r.F(k).SetInt(img, v)
		}
		if o.model == "RT21" || o.model == "RT21V" {
			r.F("scramble_type2").SetInt(img, keep["scramble_type"])
		}
	}
	return nil
}

func (f rt21Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	d, err := rbDownload(ctx, port, m.Baud, rt21o(m).spec, prog)
	return d, "", err
}

func (f rt21Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	return "", rbUpload(ctx, port, m.Baud, rt21o(m).spec, img, prog)
}
