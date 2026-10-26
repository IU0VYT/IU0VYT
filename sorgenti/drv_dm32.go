// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// baofeng_dm32uv.py (Copyright 2026 chaffed).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Baofeng DM-32UV (radio DMR). Driver portato da CHIRP (baofeng_dm32uv.py, a sua
// volta ricavato dal programma del costruttore «BAOFENG DM-32UV CPS» v1.60 e
// provato su una radio vera). Qui si gestiscono i canali: frequenze, nome,
// potenza, banda e toni dei canali analogici; i canali DMR («DMR») si possono
// spostare, rinominare e cambiare di frequenza, mentre le impostazioni DMR
// (contatto, gruppo, slot, color code…), le zone, le liste e le impostazioni
// della radio restano quelle già presenti. Cancellando un canale lo si toglie
// anche dalle zone e dalle liste di scansione, come fa CHIRP.
//
// Lo stesso programma del costruttore (v1.60, stesso eseguibile «DMR CPS.exe»)
// è pubblicato anche per la DM-1801 Pro.
//
// Memoria: pagine flash da 4 KB il cui ultimo byte è un'etichetta (tag) che ne
// indica il contenuto; la radio le sposta quando le riscrive. L'immagine è
// logica: una pagina da 4 KB per ogni etichetta di dm32Tags, nell'ordine.
// Protocollo (115200 baud): «PSEARCH» → 06 + «DP570UV»; «PASSSTA» → 'P' + 2 byte
// (password); «SYSINFO» → 06; domande 'V' (informazioni, area della memoria);
// 'G' 00 00 00 00 01 → 'S' + 0x105 byte; FF FF FF FF 0C; «PROGRAM» → 06;
// 02 → 8 byte; 06 → 06. Lettura: 'R' indirizzo(3, LE) lunghezza(2, LE) → 'W' +
// gli stessi 5 byte + dati. Scrittura di una pagina intera: 'W' indirizzo(3)
// lunghezza(2) + 4096 byte → 06, poi rilettura di controllo.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	dm32Page       = 0x1000
	dm32ChSize     = 0x30
	dm32ChPerPage  = 85
	dm32ChCount    = 4000
	dm32ZonePerPg  = 28
	dm32ZoneCount  = 250
	dm32ZoneMem    = 64
	dm32ZoneRec    = 0x91
	dm32ScanTag    = 0x11
	dm32ScanMax    = 32
	dm32ScanRec    = 0x39
	dm32ScanName   = 11
	dm32ScanMem    = 16
	dm32ZoneTag0   = 0x5C
	dm32WorkMask   = 0x3E
	dm32ModelID    = "DP570UV"
	dm32ContactMax = 800
)

var dm32Tags = func() []byte {
	t := []byte{0x02, 0x03, 0x04, 0x06, 0x0A, 0x0B, 0x0F, 0x10, 0x11}
	for x := 0x12; x < 0x49; x++ {
		t = append(t, byte(x))
	}
	for x := 0x5C; x < 0x68; x++ {
		t = append(t, byte(x))
	}
	return t
}()

// pagine scritte dalla scrittura, nell'ordine (come CHIRP: prima le liste, poi i
// canali, le zone per ultime); 0x02 e 0x69 (taratura) non si scrivono mai
var dm32UploadTags = func() []byte {
	t := []byte{0x03, 0x04, 0x06, 0x67}
	for x := 0x44; x < 0x49; x++ {
		t = append(t, byte(x))
	}
	t = append(t, 0x0B, 0x0F, 0x11)
	for x := 0x12; x < 0x42; x++ {
		t = append(t, byte(x))
	}
	t = append(t, 0x42, 0x43)
	for x := 0x5C; x < 0x65; x++ {
		t = append(t, byte(x))
	}
	return t
}()

var dm32ScanDefaultOpts = []byte{0x03, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x7f}

func dm32Slot(tag byte) int { return bytes.IndexByte(dm32Tags, tag) * dm32Page }

var dm32Format = func() string {
	var b strings.Builder
	b.WriteString(`
struct chan {
  char name[16];
  lbcd rxfreq[4];
  lbcd txfreq[4];
  u8 chtype:4, forbid_tx:1, power:2, lone_work:1;
  u8 wide:1, auto_scan:1, scanlist:6;
  u8 forbid_talkaround:1, tx_admit:3, unknown1a:1, aprs_rx:1, offset_dir:2;
  u8 emerg_indicator:1, emerg_ack:1, unknown1b:1, emerg_system:5;
  u8 squelch:4, aprs_report:2, aprs_ptt_analog:1, aprs_ptt_digital:1;
  u8 private_confirm:1, short_data_confirm:1, tdma_direct:1, timeslot:1,
     colorcode:4;
  u8 privacy;
  u8 unknown1f:1, encrypt:1, rxgroup:6;
  u8 aprs_channel;
  ul16 rxtone;
  ul16 txtone;
  u8 unknown25:2, compander:1, vox:1, unknown25b:4;
  u8 ptt_id_display:1, rx_squelch_mode:3, signaling:3, unknown26:1;
  u8 rx_signal:4, tx_signal:4;
  u8 unknown28;
  u8 step:4, ptt_id:2, unknown29:2;
  u8 unknown2a;
  u8 radio_id;
  lbcd offset[4];
};
struct zone {
  char name[16];
  u8 count;
  ul16 members[64];
};
#seekto 0x09000;
ul16 ch_count;
#seekto 0x09010;
struct chan page0[84];
`)
	for p := 1; p < 47; p++ {
		fmt.Fprintf(&b, "#seekto 0x%05x;\nstruct chan page%d[85];\n", 0x09000+p*dm32Page, p)
	}
	b.WriteString(`#seekto 0x38000;
struct chan page47[6];
#seekto 0x40000;
struct {
  u8 count;
  u8 a_pos;
  u8 unknown2;
  u8 b_pos;
  u8 unknown4;
  u8 a_zone;
  u8 unknown6;
  u8 b_zone;
} zone_hdr;
#seekto 0x40010;
struct zone zpage0[28];
`)
	for p := 1; p < 9; p++ {
		fmt.Fprintf(&b, "#seekto 0x%05x;\nstruct zone zpage%d[28];\n", 0x40000+p*dm32Page, p)
	}
	return b.String()
}()

var dm32Steps = []float64{2.5, 5.0, 6.25, 10.0, 12.5, 25.0, 50.0, 100.0}

type dm32Fam struct{}

func init() {
	genRegister(&genModel{ID: "bf-dm32uv", Vendor: "Baofeng", Model: "DM-32UV", RClass: "DM32UV", Name: "Baofeng DM-32UV",
		Slots: dm32ChCount, First: 1, NameMax: 16,
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 2,5 W", 2.5, 1}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, DCS: dtcsCodes, Baud: 115200,
		Fam: dm32Fam{}, Digital: "DMR",
		Notes: "Radio DMR: qui si modificano frequenze, nome, potenza e toni; per i canali DMR contatto, gruppo, slot e color code " +
			"restano quelli impostati (per cambiarli usa il programma del costruttore o CHIRP). Zone, liste e impostazioni della radio " +
			"non si toccano; un canale cancellato viene tolto anche dalle zone e dalle liste di scansione. " +
			"Alcuni cavi con chip CH340 alterano i dati di questa radio: è consigliato un cavo con chip FTDI o CP2102."})
}

func (dm32Fam) mem() *bwEl { return bwMust(dm32Format) }

func (f dm32Fam) layout(m *genModel) *bwEl { return f.mem() }

// canale n (1-4000): pagina e posizione come le calcola il programma del costruttore
func dm32ChanOffset(n int) (int, int) {
	if n < dm32ChPerPage {
		return 0, n - 1
	}
	return n / dm32ChPerPage, n % dm32ChPerPage
}

func (f dm32Fam) ch(n int) *bwEl {
	p, i := dm32ChanOffset(n)
	return f.mem().F(fmt.Sprintf("page%d", p)).I(i)
}

func (f dm32Fam) count(img []byte) int {
	c := int(f.mem().F("ch_count").Int(img))
	if c > dm32ChCount {
		return 0
	}
	return c
}

// tono (ul16): 0xFFFF / 0x0000 nessuno, 0x8000 + codice DCS in BCD (0x4000
// invertito), altrimenti decimi di Hz in BCD
func dm32ToneDec(v int64) string {
	if v == 0xFFFF || v == 0 {
		return ""
	}
	if v&0x8000 != 0 {
		var code int
		fmt.Sscanf(fmt.Sprintf("%03x", v&0x07FF), "%d", &code)
		return genTone{Mode: "DTCS", Code: code, Inv: v&0x4000 != 0}.String()
	}
	var t int
	fmt.Sscanf(fmt.Sprintf("%04x", v&0x7FFF), "%d", &t)
	return fmtTone(t)
}

func dm32ToneEnc(t genTone) int64 {
	hex := func(n int) int64 {
		var v int64
		fmt.Sscanf(fmt.Sprintf("%d", n), "%x", &v)
		return v
	}
	switch t.Mode {
	case "Tone":
		return hex(t.Hz10)
	case "DTCS":
		v := hex(t.Code) | 0x8000
		if t.Inv {
			v += 0x4000
		}
		return v
	}
	return 0xFFFF
}

func (f dm32Fam) Get(m *genModel, img []byte, i int) *Chan {
	n := i + 1
	r := f.ch(n)
	rx := r.F("rxfreq").Raw(img)
	if n > f.count(img) || bytes.Equal(rx, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(rx, []byte{0, 0, 0, 0}) {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.Name = strings.TrimRight(string(r.F("name").Raw(img)), "\x00\xff ")
	ct := r.F("chtype").Int(img)
	c.Dig = ct == 1 || ct == 3
	c.Narrow = !r.F("wide").Bool(img)
	pw := int(r.F("power").Int(img))
	if pw > 2 {
		pw = 2
	}
	c.Power = m.powerSem(pw)
	if !c.Dig {
		c.TxTone = dm32ToneDec(r.F("txtone").Int(img))
		c.RxTone = dm32ToneDec(r.F("rxtone").Int(img))
	}
	return c
}

func (f dm32Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.ch(i+1).Raw(img)...)
}

func (f dm32Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.ch(i+1).SetRaw(img, raw) }

func (f dm32Fam) Match(m *genModel, img []byte) bool { return false }

// ---- pagine «grezze» (liste fuori dalla mappa)

func dm32PageOf(img []byte, tag byte) []byte {
	b := dm32Slot(tag)
	return img[b : b+dm32Page]
}

// put scrive nella pagina e la segna come presente (etichetta nell'ultimo byte)
func dm32Put(img []byte, tag byte, off int, data []byte) {
	p := dm32PageOf(img, tag)
	copy(p[off:], data)
	p[dm32Page-1] = tag
}

func dm32Text(raw []byte) []byte {
	if i := bytes.IndexByte(raw, 0xFF); i >= 0 {
		raw = raw[:i]
	}
	if i := bytes.IndexByte(raw, 0x00); i >= 0 {
		raw = raw[:i]
	}
	return raw
}

func dm32NameBytes(name []byte, n int) []byte {
	out := make([]byte, n)
	copy(out, name)
	return out
}

// ---- contatto di trasmissione (fuori dal canale, pagine 0x42/0x43)

func dm32TxcLoc(n int) (byte, int) {
	if n < 0x800 {
		return 0x42, 2 * (n - 1)
	}
	return 0x43, 2 * (n & 0x7FF)
}

func dm32TxContact(img []byte, n int) int {
	tag, off := dm32TxcLoc(n)
	p := dm32PageOf(img, tag)
	v := int(p[off]>>4)<<8 | int(p[off+1])
	if v > dm32ContactMax {
		return 0
	}
	return v
}

func dm32UpdateTxContact(img []byte, n, contact int, digital bool) {
	tag, off := dm32TxcLoc(n)
	p := dm32PageOf(img, tag)
	hi, lo := p[off], p[off+1]
	if hi == 0xFF && lo == 0xFF && contact == 0 {
		return
	}
	d := byte(0)
	if digital {
		d = 1
	}
	if dm32TxContact(img, n) != contact || hi&1 != d {
		h := hi & 0x0E
		if hi == 0xFF {
			h = 0
		}
		dm32Put(img, tag, off, []byte{byte(contact>>8)<<4 | h | d, byte(contact)})
	}
}

// ---- zone

func (f dm32Fam) zoneCount(img []byte) int {
	c := int(f.mem().F("zone_hdr").F("count").Int(img))
	if c > dm32ZoneCount {
		return 0
	}
	return c
}

func (f dm32Fam) zone(z int) *bwEl {
	return f.mem().F(fmt.Sprintf("zpage%d", (z-1)/dm32ZonePerPg)).I((z - 1) % dm32ZonePerPg)
}

func (f dm32Fam) zoneMembers(img []byte, z int) []int {
	if z > f.zoneCount(img) {
		return nil
	}
	zr := f.zone(z)
	cnt := int(zr.F("count").Int(img))
	if cnt > dm32ZoneMem {
		cnt = dm32ZoneMem
	}
	var out []int
	for k := 0; k < cnt; k++ {
		v := int(zr.F("members").I(k).Int(img))
		if v >= 1 && v <= dm32ChCount {
			out = append(out, v)
		}
	}
	return out
}

func (f dm32Fam) setZoneMembers(img []byte, z int, members []int) {
	zr := f.zone(z)
	for k := 0; k < dm32ZoneMem; k++ {
		v := 0
		if k < len(members) {
			v = members[k]
		}
		zr.F("members").I(k).SetInt(img, int64(v))
	}
	zr.F("count").SetInt(img, int64(len(members)))
}

// zone 1..count: nome (byte) e canali
type dm32Zone struct {
	name    []byte
	members []int
}

func (f dm32Fam) zoneList(img []byte) []dm32Zone {
	var out []dm32Zone
	for z := 1; z <= f.zoneCount(img); z++ {
		out = append(out, dm32Zone{dm32Text(f.zone(z).F("name").Raw(img)), f.zoneMembers(img, z)})
	}
	return out
}

func (f dm32Fam) dropEmptyLastZone(img []byte) {
	c := f.zoneCount(img)
	if c > 1 && len(f.zoneMembers(img, c)) == 0 {
		f.zone(c).F("name").Fill(img, 0xFF)
		f.mem().F("zone_hdr").F("count").SetInt(img, int64(c-1))
	}
}

// ---- liste di scansione (pagina 0x11)

type dm32Scan struct {
	name    []byte
	members []int
	opts    []byte
}

func dm32ScanLists(img []byte) []dm32Scan {
	p := dm32PageOf(img, dm32ScanTag)
	cnt := int(p[0])
	if cnt > dm32ScanMax {
		cnt = 0
	}
	var out []dm32Scan
	for n := 1; n <= cnt; n++ {
		rec := p[dm32ScanRec*n-0x38:][:dm32ScanRec]
		k := int(rec[0x0B])
		if k > dm32ScanMem {
			k = dm32ScanMem
		}
		var mem []int
		for i := 0; i < k; i++ {
			v := int(binary.LittleEndian.Uint16(rec[0x18+2*i:]))
			if v >= 1 && v <= dm32ChCount {
				mem = append(mem, v)
			}
		}
		out = append(out, dm32Scan{append([]byte{}, dm32Text(rec[:dm32ScanName])...), mem, append([]byte{}, rec[0x0C:0x18]...)})
	}
	return out
}

func dm32SetScanLists(img []byte, lists []dm32Scan) {
	def := dm32ScanDefaultOpts
	if len(lists) > 0 {
		def = lists[0].opts
	}
	old := len(dm32ScanLists(img))
	dm32Put(img, dm32ScanTag, 0, []byte{byte(len(lists))})
	for n := 1; n <= dm32ScanMax; n++ {
		base := dm32ScanRec*n - 0x38
		if n > len(lists) {
			if n <= old {
				dm32Put(img, dm32ScanTag, base, make([]byte, dm32ScanRec))
			}
			continue
		}
		l := lists[n-1]
		opts := append([]byte{}, l.opts...)
		if len(opts) != 12 {
			opts = append([]byte{}, def...)
		}
		mem := l.members
		if len(mem) > dm32ScanMem {
			mem = mem[:dm32ScanMem]
		}
		designed := int(binary.LittleEndian.Uint16(opts[3:5]))
		if len(mem) > 0 && idxInt(mem, designed) < 0 {
			binary.LittleEndian.PutUint16(opts[3:5], uint16(mem[0]))
		}
		rec := append(dm32NameBytes(l.name, dm32ScanName), byte(len(mem)))
		rec = append(rec, opts...)
		for _, v := range mem {
			rec = binary.LittleEndian.AppendUint16(rec, uint16(v))
		}
		rec = append(rec, make([]byte, 2*(dm32ScanMem-len(mem)))...)
		if len(rec) > dm32ScanRec {
			rec = rec[:dm32ScanRec]
		}
		dm32Put(img, dm32ScanTag, base, rec) // 0x38 byte: l'ultimo del record resta com'è (come CHIRP)
	}
}

func dm32RemoveFromScanLists(img []byte, number int) {
	lists := dm32ScanLists(img)
	changed := false
	var out []dm32Scan
	for _, l := range lists {
		var mem []int
		for _, v := range l.members {
			if v != number {
				mem = append(mem, v)
			} else {
				changed = true
			}
		}
		opts := l.opts
		if len(mem) == 0 && int(binary.LittleEndian.Uint16(opts[3:5])) == number {
			o := append([]byte{}, opts[:3]...)
			o = append(o, dm32ScanDefaultOpts[3:5]...)
			o = append(o, opts[5:]...)
			if !bytes.Equal(o, opts) {
				changed = true
			}
			opts = o
		}
		out = append(out, dm32Scan{l.name, mem, opts})
	}
	if changed {
		dm32SetScanLists(img, out)
	}
}

// ---- canali (come set_memory di CHIRP)

func (f dm32Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	n := i + 1
	r := f.ch(n)
	if c == nil {
		r.Fill(img, 0xFF)
		dm32UpdateTxContact(img, n, 0, false)
		// le zone non devono indicare un canale che non c'è più
		for z := 1; z <= f.zoneCount(img); z++ {
			mem := f.zoneMembers(img, z)
			if idxInt(mem, n) >= 0 {
				var keep []int
				for _, v := range mem {
					if v != n {
						keep = append(keep, v)
					}
				}
				f.setZoneMembers(img, z, keep)
			}
		}
		f.dropEmptyLastZone(img)
		dm32RemoveFromScanLists(img, n)
		if n == f.count(img) {
			cnt := n - 1
			for cnt > 0 && f.Get(m, img, cnt-1) == nil {
				cnt--
			}
			f.mem().F("ch_count").SetInt(img, int64(cnt))
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
	if c.Dig {
		tt, rt = genTone{}, genTone{}
	}
	txc := dm32TxContact(img, n)
	raw := r.Raw(img)
	// passo di sintonia: resta quello del canale; 5 kHz (valore predefinito di
	// CHIRP) per un canale nuovo o con un valore non valido
	step := int(r.F("step").Int(img))
	if f.Get(m, img, i) == nil {
		step = len(dm32Steps)
	}
	if bytes.Equal(raw, bytes.Repeat([]byte{0xFF}, dm32ChSize)) || bytes.Equal(raw, make([]byte, dm32ChSize)) {
		// valori dei canali con cui arriva la radio
		tpl := append(bytes.Repeat([]byte{0xFF}, 16), make([]byte, 8)...)
		tpl = append(tpl, 0, 0, 0, 0, 0x30, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0)
		tpl = append(tpl, make([]byte, 8)...)
		r.SetRaw(img, tpl)
	}
	if cnt := f.count(img); n > cnt {
		// i canali saltati entrano in 1..n: devono essere vuoti
		for g := cnt + 1; g < n; g++ {
			f.ch(g).Fill(img, 0xFF)
		}
		f.mem().F("ch_count").SetInt(img, int64(n))
	}
	name := make([]byte, 16)
	copy(name, c.Name)
	r.F("name").SetRaw(img, name)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	ct := r.F("chtype").Int(img)
	if c.Dig && ct != 1 && ct != 3 {
		r.F("chtype").SetInt(img, 1)
	} else if !c.Dig && ct != 0 && ct != 2 {
		r.F("chtype").SetInt(img, 0)
	}
	r.F("wide").SetBool(img, !c.Dig && !c.Narrow)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	if step >= len(dm32Steps) {
		r.F("step").SetInt(img, 1)
	}
	r.F("rxtone").SetInt(img, dm32ToneEnc(rt))
	r.F("txtone").SetInt(img, dm32ToneEnc(tt))
	ct = r.F("chtype").Int(img)
	dm32UpdateTxContact(img, n, txc, ct == 1 || ct == 3)
	return nil
}

// ---------------------------------------------------------------- protocollo

type dm32Link struct {
	ctx context.Context
	p   Port
}

func (l *dm32Link) send(d []byte) error {
	time.Sleep(10 * time.Millisecond)
	_ = l.p.Purge()
	return l.p.Write(d)
}

func (l *dm32Link) recv(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := l.p.ReadFull(l.ctx, b, to); err != nil {
		return nil, err
	}
	return b, nil
}

func (l *dm32Link) xfer(d []byte, n int) ([]byte, error) {
	if err := l.send(d); err != nil {
		return nil, err
	}
	r, err := l.recv(n, 500*time.Millisecond)
	if err == errTimeout {
		return nil, rErr("nessuna risposta della radio a %q", d[:min(7, len(d))])
	}
	return r, err
}

// avviso sui cavi che alterano il bit 7 (CH340)
const dm32Cable = "alcuni cavi con chip CH340 alterano i dati di questa radio: usa un cavo con chip FTDI o CP2102"

func (l *dm32Link) queryV(cmd []byte) ([]byte, error) {
	if err := l.send(cmd); err != nil {
		return nil, err
	}
	h, err := l.recv(3, 500*time.Millisecond)
	if err != nil || h[0] != 'V' {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("risposta errata alla domanda %X", cmd)
	}
	b, err := l.recv(int(h[2]), 500*time.Millisecond)
	if err != nil {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("risposta errata alla domanda %X", cmd)
	}
	return b, nil
}

func (l *dm32Link) readBlock(addr, n int, to time.Duration) ([]byte, error) {
	cmd := []byte{'R', byte(addr), byte(addr >> 8), byte(addr >> 16), byte(n), byte(n >> 8)}
	if err := l.send(cmd); err != nil {
		return nil, err
	}
	r, err := l.recv(6+n, to)
	if err != nil {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("lettura della memoria a %06X non riuscita", addr)
	}
	want := append([]byte{'W'}, cmd[1:]...)
	if !bytes.Equal(r[:6], want) {
		if bytes.Equal(dm32Strip7(r[:6]), want) {
			return nil, rErr("risposta alterata leggendo %06X: %s", addr, dm32Cable)
		}
		return nil, rErr("risposta errata leggendo la memoria a %06X", addr)
	}
	return r[6:], nil
}

func dm32Strip7(b []byte) []byte {
	out := make([]byte, len(b))
	for i, x := range b {
		out[i] = x & 0x7F
	}
	return out
}

func (l *dm32Link) search() ([]byte, error) {
	garbled := false
	for k := 0; k < 5; k++ {
		if err := l.send([]byte("PSEARCH")); err != nil {
			return nil, err
		}
		r, err := l.recv(8, 500*time.Millisecond)
		if err == errCancelled {
			return nil, err
		}
		if err == nil && r[0] == 0x06 {
			ok := true
			for _, x := range r[1:] {
				if x >= 0x80 {
					ok = false
				}
			}
			if ok {
				return r[1:], nil
			}
			garbled = true
		}
	}
	if garbled {
		return nil, rErr("risposta alterata dalla radio: %s", dm32Cable)
	}
	return nil, rErr("la radio non risponde: controlla che sia accesa e collegata")
}

func (l *dm32Link) identify(writing bool) (string, int, int, error) {
	id, err := l.search()
	if err != nil {
		return "", 0, 0, err
	}
	if string(id) != dm32ModelID {
		return "", 0, 0, rErr("la radio si identifica come %q, non come Baofeng DM-32UV (%q)", id, dm32ModelID)
	}
	r, err := l.xfer([]byte("PASSSTA"), 3)
	if err != nil {
		return "", 0, 0, err
	}
	if r[0] != 'P' {
		return "", 0, 0, rErr("risposta inattesa a PASSSTA")
	}
	if r[2] == 0xA5 {
		return "", 0, 0, rErr("sulla radio è impostata una password di lettura: toglila con il programma del costruttore")
	}
	if writing && r[1] == 0xA5 {
		return "", 0, 0, rErr("sulla radio è impostata una password di scrittura: toglila con il programma del costruttore")
	}
	if r, err := l.xfer([]byte("SYSINFO"), 1); err != nil || r[0] != 0x06 {
		if err != nil {
			return "", 0, 0, err
		}
		return "", 0, 0, rErr("la radio ha rifiutato SYSINFO")
	}
	if _, err := l.queryV([]byte{'V', 0, 0, 0x40, 0x0D}); err != nil {
		return "", 0, 0, err
	}
	info := map[int][]byte{}
	for i := 1; i <= 16; i++ {
		if i == 12 {
			continue
		}
		b, err := l.queryV([]byte{'V', 0, 0, 0, byte(i)})
		if err != nil {
			return "", 0, 0, err
		}
		info[i] = b
	}
	if len(info[10]) < 8 {
		return "", 0, 0, rErr("area della memoria non indicata dalla radio")
	}
	start := int(binary.LittleEndian.Uint32(info[10]))
	end := int(binary.LittleEndian.Uint32(info[10][4:]))
	if (start|end)>>24 != 0 || start >= end {
		return "", 0, 0, rErr("area della memoria inattesa")
	}
	return strings.TrimRight(string(info[1]), "\x00"), start, end, nil
}

func (l *dm32Link) enterProgram() error {
	r, err := l.xfer([]byte{'G', 0, 0, 0, 0, 1}, 0x106)
	if err != nil {
		return err
	}
	if r[0] != 'S' {
		return rErr("risposta inattesa a G")
	}
	if err := l.send([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0C}); err != nil {
		return err
	}
	if r, err := l.xfer([]byte("PROGRAM"), 1); err != nil || r[0] != 0x06 {
		if err != nil {
			return err
		}
		return rErr("la radio ha rifiutato la modalità di programmazione")
	}
	if _, err := l.xfer([]byte{0x02}, 8); err != nil {
		return err
	}
	if r, err := l.xfer([]byte{0x06}, 1); err != nil || r[0] != 0x06 {
		if err != nil {
			return err
		}
		return rErr("la radio non ha confermato")
	}
	return nil
}

// etichetta di ogni pagina: {etichetta: indirizzi}, pagine libere
func (l *dm32Link) scan(start, end int, prog func(i int)) (map[byte][]int, []int, error) {
	where := map[byte][]int{}
	k := 0
	for a := start; a+dm32Page-1 <= end; a += dm32Page {
		b, err := l.readBlock(a+dm32Page-1, 1, 500*time.Millisecond)
		if err != nil {
			return nil, nil, err
		}
		if b[0] >= 0x80 && b[0] < 0xFF {
			return nil, nil, rErr("etichetta di pagina impossibile %02X a %06X: %s", b[0], a, dm32Cable)
		}
		where[b[0]] = append(where[b[0]], a)
		k++
		prog(k)
	}
	free := where[0xFF]
	delete(where, 0xFF)
	return where, free, nil
}

func dm32Open(ctx context.Context, port string) (*dm32Link, error) {
	p, err := openPortFunc(port, 115200)
	if err != nil {
		return nil, err
	}
	return &dm32Link{ctx: ctx, p: p}, nil
}

func (f dm32Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	l, err := dm32Open(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer l.p.Close()
	prog(0, 100, "Collegamento con la radio…")
	fw, start, end, err := l.identify(false)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	info := "firmware " + fw
	if err := l.enterProgram(); err != nil {
		return nil, info, genFail("Programmazione", err)
	}
	pages := (end + 1 - start) / dm32Page
	total := pages + len(dm32Tags)
	where, _, err := l.scan(start, end, func(i int) { prog(i, total, "Ricerca delle pagine…") })
	if err != nil {
		return nil, info, genFail("Lettura", err)
	}
	img := bytes.Repeat([]byte{0xFF}, len(dm32Tags)*dm32Page)
	for i, tag := range dm32Tags {
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		if a := where[tag]; len(a) > 0 {
			d, err := l.readBlock(a[0], dm32Page, 5*time.Second)
			if err != nil {
				return nil, info, genFail("Lettura", err)
			}
			copy(img[i*dm32Page:], d)
		}
		prog(pages+i+1, total, "Lettura…")
	}
	if dup := dm32Dupes(where, dm32Tags); dup != "" {
		info += "; pagine doppie " + dup + " (si usa la prima; la scrittura sarà rifiutata finché la radio non le riordina)"
	}
	prog(total, total, "Lettura completata")
	return img, info, nil
}

func dm32Dupes(where map[byte][]int, tags []byte) string {
	var d []string
	for _, t := range tags {
		if len(where[t]) > 1 {
			d = append(d, fmt.Sprintf("%02X", t))
		}
	}
	return strings.Join(d, ", ")
}

func dm32ZoneRecord(rec []byte) dm32Zone {
	cnt := int(rec[16])
	if cnt > dm32ZoneMem {
		cnt = dm32ZoneMem
	}
	var mem []int
	for i := 0; i < cnt; i++ {
		mem = append(mem, int(binary.LittleEndian.Uint16(rec[17+2*i:])))
	}
	return dm32Zone{dm32Text(rec[:16]), mem}
}

func dm32ZoneEq(a, b dm32Zone) bool {
	if !bytes.Equal(a.name, b.name) || len(a.members) != len(b.members) {
		return false
	}
	for i := range a.members {
		if a.members[i] != b.members[i] {
			return false
		}
	}
	return true
}

// byte 1-7 dell'intestazione delle zone: zona e posizione mostrate sulle righe
// A/B, presi dalla radio e spostati dove si trovano nel nuovo elenco
func dm32FollowDisplay(want, current []byte, radioZone func(z int) (dm32Zone, bool), zones []dm32Zone) []byte {
	w := append([]byte{}, want...)
	copy(w[1:8], current[1:8])
	for _, pz := range [][2]int{{1, 5}, {3, 7}} {
		pos, zf := pz[0], pz[1]
		z := int(current[zf])
		var old dm32Zone
		has := false
		if z >= 1 && z <= dm32ZoneCount {
			old, has = radioZone(z)
		}
		nz := 0
		if has {
			for i, zn := range zones {
				if dm32ZoneEq(zn, old) {
					nz = i + 1
					break
				}
			}
			if nz == 0 {
				for i, zn := range zones {
					if bytes.Equal(zn.name, old.name) {
						nz = i + 1
						break
					}
				}
			}
		}
		if nz == 0 {
			if !has && z >= 1 && z <= len(zones) {
				nz = z
			} else {
				nz = 1
			}
		}
		var members []int
		if len(zones) > 0 {
			members = zones[nz-1].members
		}
		p := int(current[pos])
		if has {
			ch := -1
			if p >= 1 && p <= len(old.members) {
				ch = old.members[p-1]
			}
			if k := idxInt(members, ch); ch >= 0 && k >= 0 {
				p = k + 1
			} else {
				p = 1
			}
		}
		if p < 1 || p > max(len(members), 1) {
			p = 1
		}
		w[zf], w[pos] = byte(nz), byte(p)
	}
	return w
}

func (f dm32Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < len(dm32Tags)*dm32Page {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	l, err := dm32Open(ctx, port)
	if err != nil {
		return "", err
	}
	defer l.p.Close()
	prog(0, 100, "Collegamento con la radio…")
	fw, start, end, err := l.identify(true)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	info := "firmware " + fw
	if err := l.enterProgram(); err != nil {
		return info, genFail("Programmazione", err)
	}
	pages := (end + 1 - start) / dm32Page
	total := pages + len(dm32UploadTags)
	var where map[byte][]int
	radioZone := func(z int) (dm32Zone, bool) {
		pg, idx := (z-1)/dm32ZonePerPg, (z-1)%dm32ZonePerPg
		a := where[byte(dm32ZoneTag0+pg)]
		if len(a) == 0 {
			return dm32Zone{}, false
		}
		off := idx * dm32ZoneRec
		if pg == 0 {
			off += 0x10
		}
		rec, err := l.readBlock(a[0]+off, dm32ZoneRec, 5*time.Second)
		if err != nil {
			return dm32Zone{}, false
		}
		return dm32ZoneRecord(rec), true
	}
	var zones []dm32Zone
	for _, z := range f.zoneList(img) {
		zones = append(zones, dm32Zone{bytes.TrimRight(dm32NameBytes(z.name, 16), "\x00"), z.members})
	}
	where, free, err := l.scan(start, end, func(i int) { prog(i, total, "Ricerca delle pagine…") })
	if err != nil {
		return info, genFail("Scrittura", err)
	}
	if dup := dm32Dupes(where, dm32UploadTags); dup != "" {
		return info, fmt.Errorf("la radio ha più pagine con l'etichetta %s: nessuna scrittura (riprova dopo aver acceso e spento la radio)", dup)
	}
	written := 0
	ff := bytes.Repeat([]byte{0xFF}, dm32Page-1)
	for i, tag := range dm32UploadTags {
		if ctx.Err() != nil {
			return info, errCancelled
		}
		slot := dm32Slot(tag)
		data := img[slot : slot+dm32Page-1]
		if !bytes.Equal(data, ff) || img[slot+dm32Page-1] == tag && len(where[tag]) > 0 {
			want := append(append([]byte{}, data...), tag)
			addr := -1
			if a := where[tag]; len(a) > 0 {
				addr = a[0]
				cur, err := l.readBlock(addr, dm32Page, 5*time.Second)
				if err != nil {
					return info, genFail("Scrittura", err)
				}
				if tag == dm32ZoneTag0 {
					want = dm32FollowDisplay(want, cur, radioZone, zones)
				} else if tag == 0x04 {
					want[0x80] = want[0x80]&^dm32WorkMask | cur[0x80]&dm32WorkMask
				}
				if bytes.Equal(cur, want) {
					addr = -1
				}
			} else if len(free) > 0 {
				addr, free = free[0], free[1:]
			} else {
				return info, fmt.Errorf("nessuna pagina libera sulla radio")
			}
			if addr >= 0 {
				if err := l.writePage(addr, want, start, end); err != nil {
					return info, genFail("Scrittura", err)
				}
				written++
			}
		}
		prog(pages+i+1, total, "Scrittura…")
	}
	prog(total, total, "Scrittura completata")
	return fmt.Sprintf("%s; %d pagine scritte", info, written), nil
}

func (l *dm32Link) writePage(addr int, data []byte, start, end int) error {
	if len(data) != dm32Page || addr%dm32Page != 0 || addr < start || addr+dm32Page-1 > end || data[dm32Page-1] == 0x02 || data[dm32Page-1] == 0x69 {
		return rErr("scrittura non sicura a %06X: rifiutata", addr)
	}
	cmd := []byte{'W', byte(addr), byte(addr >> 8), byte(addr >> 16), byte(dm32Page & 0xFF), byte(dm32Page >> 8)}
	if err := l.send(append(cmd, data...)); err != nil {
		return err
	}
	ack, err := l.recv(1, 5*time.Second)
	if err != nil || ack[0] != 0x06 {
		if err == errCancelled {
			return err
		}
		return rErr("la pagina a %06X non è stata confermata: riprova o ripristina la copia di sicurezza", addr)
	}
	back, err := l.readBlock(addr, dm32Page, 5*time.Second)
	if err != nil {
		return err
	}
	if !bytes.Equal(back, data) {
		return rErr("la pagina a %06X non si rilegge come scritta: riprova o ripristina la copia di sicurezza", addr)
	}
	return nil
}
