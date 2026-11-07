// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_ha1g.py, retevis_ha2.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: Retevis (Ailunce) HA1G, HA1UV e HA2 (retevis_ha1g.py e retevis_ha2.py di
// CHIRP). 115200 baud, protocollo «RDTP» a pacchetti con CRC16: la memoria è
// divisa in zone (impostazioni, zone, canali, scansioni, allarmi, DTMF…) lette e
// scritte una per volta. I canali presenti sono quelli nell'elenco degli indici;
// cancellando un canale si tolgono anche i suoi riferimenti (canale di
// accensione, zone, liste di scansione, allarmi), come fa CHIRP.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	haG1 = 1 + iota
	haUV
	haA2
)

type haRegion struct {
	name       string
	id         byte
	start, len int
}

type ha1gOpt struct {
	kind    int
	model   string // nome che la radio usa nel collegamento
	skipVfo int    // canali VFO in testa all'elenco degli indici
	reqVer  string
	regions []haRegion
	exclude map[string]bool
	size    int
}

type ha1gFam struct{}

func ha1go(m *genModel) *ha1gOpt { return m.Opt.(*ha1gOpt) }

var (
	haNameChars = charsetUpperNum + "-/;,._!?*#@$%&+=<>~(){}]'"
	haNameSet   = charsetAlnum + "-/;,._!? *#@$%&+=/<>~(){}]'"
)

func init() {
	r1 := []haRegion{{"radioHead", 2, 0, 14}, {"radioInfo", 3, 14, 68}, {"radioVer", 4, 82, 10}, {"settingData", 6, 92, 100},
		{"zoneData", 7, 192, 3202}, {"channelData", 8, 3394, 43136}, {"scanData", 11, 46530, 3618},
		{"vfoScanData", 12, 50148, 68}, {"alarmData", 13, 50244, 258}, {"dTMFData", 15, 51272, 842}}
	r2 := []haRegion{{"radioHead", 2, 0, 14}, {"radioInfo", 3, 14, 68}, {"radioVer", 4, 82, 10}, {"settingData", 6, 92, 152},
		{"zoneData", 7, 244, 2338}, {"channelData", 8, 2582, 43010}, {"vfoChannelData", 9, 45592, 128},
		{"scanData", 11, 45720, 3618}, {"vfoScanData", 12, 49338, 68}, {"alarmData", 13, 49406, 258},
		{"fmdata", 14, 49664, 1538}, {"dTMFData", 15, 51202, 842}, {"aprsData", 38, 52334, 19364},
		{"gnssData", 51, 71698, 76}, {"aliasData", 53, 71774, 302}}
	ex := map[string]bool{"radioHead": true, "radioInfo": true, "radioVer": true, "scanData": true, "alarmData": true}
	pw := []genPower{{"Bassa", 1, 2}, {"Alta", 5, 0}}
	air := band{108000000, 135999999}
	note := "Copia a 115200 baud con il cavo USB della radio. Come in CHIRP: niente canali in sola ricezione né " +
		"spostamenti oltre 70 MHz; TOT, permessi di trasmissione, squelch, allarme, DTMF, talk around, VOX, compander e " +
		"scrambler del canale si conservano; cancellando un canale si tolgono i suoi riferimenti (accensione, zone, " +
		"scansioni, allarmi). Driver di CHIRP."
	genRegister(
		&genModel{ID: "rt-ha1g", Vendor: "Retevis", Model: "HA1G", RClass: "HA1G", Name: "Retevis (Ailunce) HA1G (GMRS)",
			Slots: 256, First: 1, NameMax: 12, Charset: haNameChars, Powers: pw,
			Bands: []band{air, {136000000, 174000010}, {400000000, 480000010}}, Narrow: true, NoTxOff: true, DCS: dtcsCodes,
			Baud: 115200, Fam: ha1gFam{}, Opt: &ha1gOpt{kind: haG1, model: "HA1G", skipVfo: 3, reqVer: "v1.1.11.6",
				regions: r1, exclude: ex, size: 52114},
			Notes: "Radio GMRS per gli Stati Uniti (in Italia la banda GMRS non è ammessa): i canali 1-30 sono i canali GMRS " +
				"fissi (frequenza e passo non modificabili, 8-14 anche banda e potenza). Serve il firmware 1.1.11.6 o " +
				"successivo. " + note},
		&genModel{ID: "rt-ha1uv", Vendor: "Retevis", Model: "HA1UV", RClass: "HA1UV", Name: "Retevis (Ailunce) HA1UV",
			Slots: 1024, First: 1, NameMax: 12, Charset: haNameChars, Powers: pw,
			Bands: []band{air, {136000000, 174000010}, {400000000, 480000010}}, Narrow: true, NoTxOff: true, DCS: dtcsCodes,
			Baud: 115200, Fam: ha1gFam{}, Opt: &ha1gOpt{kind: haUV, model: "HA1UV", skipVfo: 3, reqVer: "v1.1.11.6",
				regions: r1, exclude: ex, size: 52114},
			Notes: "1024 canali con nomi di 12 caratteri. Serve il firmware 1.1.11.6 o successivo. " + note},
		&genModel{ID: "rt-ha2", Vendor: "Retevis", Model: "HA2", RClass: "HA2", Name: "Retevis (Ailunce) HA2",
			Slots: 1024, First: 1, NameMax: 12, Charset: haNameChars, Powers: pw,
			Bands: []band{air, {136000000, 600000001}}, Narrow: true, NoTxOff: true, DCS: dtcsCodes,
			Baud: 115200, Fam: ha1gFam{}, Opt: &ha1gOpt{kind: haA2, model: "Ailunce HA2", skipVfo: 0, reqVer: "v1.0.0.0",
				regions: r2, exclude: ex, size: 72076},
			Notes: "1024 canali con nomi di 12 caratteri (radio con GPS e APRS). " + note},
	)
}

func (f ha1gFam) mem(m *genModel) *bwEl {
	if ha1go(m).kind == haA2 {
		return bwMust(ha1gDefs + ha2Format)
	}
	return bwMust(ha1gDefs + ha1gFormat)
}
func (f ha1gFam) layout(m *genModel) *bwEl { return f.mem(m) }

// chIndex: posizione del canale nella tabella della radio
func (f ha1gFam) chIndex(m *genModel, i int) int {
	if ha1go(m).kind == haA2 {
		return i
	}
	return i + 2 + 1
}

func (f ha1gFam) rec(m *genModel, i int) *bwEl {
	b := f.mem(m)
	if ha1go(m).kind == haA2 {
		return b.F("channeldata").F("channels").I(f.chIndex(m, i))
	}
	return b.F("channels").I(f.chIndex(m, i))
}
func (f ha1gFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f ha1gFam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f ha1gFam) Match(m *genModel, img []byte) bool                { return false }
func (ha1gFam) crossband() bool                                     { return true }

// fit (4.9, convertitore): canali con uno spostamento oltre 70 MHz non copiati
// (la radio, come CHIRP, non li ammette)
func (ha1gFam) fit(m *genModel, c *Chan) (bool, string) {
	if d := c.Tx - c.Rx; !c.TxOff && (d > 70000000 || d < -70000000) {
		return false, "questa radio non ammette spostamenti oltre 70 MHz (split)"
	}
	return true, ""
}

// elenco degli indici dei canali presenti (get_ch_index)
func (f ha1gFam) list(m *genModel, img []byte) []int {
	d := f.mem(m).F("channeldata")
	n := int(d.F("chnum").Int(img))
	ci := d.F("chindex")
	seen := map[int]bool{}
	var out []int
	for k := 0; k < n && k < len(ci.Elems); k++ {
		v := int(ci.I(k).Int(img))
		if v != 0xFFFF && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func (f ha1gFam) setList(m *genModel, img []byte, l []int) {
	l = append([]int{}, l...)
	sort.Ints(l)
	d := f.mem(m).F("channeldata")
	d.F("chnum").SetInt(img, int64(len(l)))
	ci := d.F("chindex")
	for k := range ci.Elems {
		v := int64(0xFFFF)
		if k < len(l) {
			v = int64(l[k])
		}
		ci.I(k).SetInt(img, v)
	}
}

func intsHas(l []int, v int) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

// filtro dei nomi come filter_name di CHIRP (maiuscole, caratteri ammessi)
func haFilterName(s string) string {
	u := []byte(s)
	for k, c := range u {
		if c >= 'a' && c <= 'z' {
			u[k] = c - 32
		}
	}
	s = string(u)
	if len(s) > 12 {
		s = s[:12]
	}
	var b []byte
	for k := 0; k < len(s); k++ {
		if strings.IndexByte(haNameChars, s[k]) >= 0 {
			b = append(b, s[k])
		}
	}
	return string(b)
}

func (f ha1gFam) Get(m *genModel, img []byte, i int) *Chan {
	if !intsHas(f.list(m, img), f.chIndex(m, i)) {
		return nil
	}
	r := f.rec(m, i)
	rx := r.F("rxfreq").Int(img)
	if rx == 0 || rx == 0xFFFFFFFF {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx, Scan: true}
	c.Name = haFilterName(string(bytes.TrimRight(r.F("alias").Raw(img), " \t\n\r\x0b\x0c")))
	switch tx := r.F("txfreq").Int(img); {
	case tx == rx:
	case tx == 0xFFFFFFFF:
		c.TxOff = true
	default:
		c.Tx = tx
	}
	c.Narrow = r.F("bandwidth").Int(img) < 3
	tone := func(typ, val string) string {
		v := r.F(val).Int(img)
		switch r.F(typ).Int(img) {
		case 1:
			if idxInt(chirpTones, int(v)) >= 0 {
				return fmtTone(int(v))
			}
		case 2, 3:
			return genTone{Mode: "DTCS", Code: octDigits(int(v)), Inv: r.F(typ).Int(img) == 3}.String()
		}
		return ""
	}
	c.TxTone = tone("txctcvaluetype", "txctc")
	c.RxTone = tone("rxctcvaluetype", "rxctc")
	if r.F("power").Int(img) == 2 {
		c.Power = m.powerSem(1)
	} else {
		c.Power = m.powerSem(0)
	}
	return c
}

// haItems: elenco di allarmi o sistemi DTMF (get_alarm_item_list /
// get_dtmf_item_list): nome e numero
type haItem struct {
	name string
	id   int
}

func (f ha1gFam) items(m *genModel, img []byte, alarm bool) []haItem {
	b := f.mem(m)
	var info *bwEl
	var out []haItem
	var num, idxs, list *bwEl
	max := 8
	if alarm {
		info = b.F("alarms")
		out = []haItem{{"OFF", 255}}
		num, idxs, list = info.F("alarmnum"), info.F("alarmindex"), info.F("alarms")
	} else {
		info = b.F("dtmfinfos")
		out = []haItem{{"OFF", 15}}
		num, idxs, list = info.F("dtmfnum"), info.F("dtmfindex"), info.F("dtmfs")
		max = 4
	}
	n := int(num.Int(img))
	if n > max {
		n = max
	}
	for k := 0; k < n; k++ {
		x := int(idxs.I(k).Int(img))
		if x > max-1 {
			x = max - 1
		}
		nm := list.I(x).F("name").Raw(img)
		var s []byte
		for j := 0; j < len(nm) && j < 12; j++ {
			if strings.IndexByte(haNameSet, nm[j]) >= 0 {
				s = append(s, nm[j])
			}
		}
		out = append(out, haItem{string(s), x})
	}
	return out
}

// haMap: valore → voce dell'elenco → valore (come CHIRP: le voci sconosciute
// diventano la prima, e due voci con lo stesso nome danno il numero della prima)
func haMap(l []haItem, v int) int {
	k := 0
	for j, it := range l {
		if it.id == v {
			k = j
			break
		}
	}
	for _, it := range l {
		if it.name == l[k].name {
			return it.id
		}
	}
	return l[0].id
}

// haMarkItems: CHIRP, aprendo l'immagine, segna come attivi gli allarmi e i
// sistemi DTMF elencati (alarmstatus/dtmfstatus = 1)
func (f ha1gFam) haMarkItems(m *genModel, img []byte) {
	b := f.mem(m)
	for _, a := range [][4]string{{"alarms", "alarmnum", "alarmindex", "alarms"}, {"dtmfinfos", "dtmfnum", "dtmfindex", "dtmfs"}} {
		info := b.F(a[0])
		max := 8
		st := "alarmstatus"
		if a[0] == "dtmfinfos" {
			max, st = 4, "dtmfstatus"
		}
		n := int(info.F(a[1]).Int(img))
		if n > max {
			n = max
		}
		for k := 0; k < n; k++ {
			x := int(info.F(a[2]).I(k).Int(img))
			if x > max-1 {
				x = max - 1
			}
			info.F(a[3]).I(x).F(st).SetInt(img, 1)
		}
	}
}

func haClampRx(hz int64) int64 {
	f := hz / 10 * 10
	switch {
	case hz == 0:
		return 0
	case hz < 108000000:
		return 108000000
	case hz > 174000000 && hz < 400000000:
		return 174000000
	case hz > 480000000:
		return 480000000
	}
	return f
}

// removeRefs: riferimenti al canale cancellato (update_channel_indexes)
func (f ha1gFam) removeRefs(m *genModel, img []byte, ci int, l []int) {
	o := ha1go(m)
	b := f.mem(m)
	next := 0
	if len(l) > o.skipVfo {
		next = l[o.skipVfo]
	}
	st := b.F("settings")
	for _, k := range []string{"1", "2"} {
		if int(st.F("homepoweronch_"+k).Int(img)) == ci {
			st.F("homepoweronzone_"+k).SetInt(img, 0xFFF)
			st.F("homepoweronch_"+k).SetInt(img, int64(next))
		}
	}
	remove := func(arr, num *bwEl) {
		cnt := int(num.Int(img))
		at := -1
		for j := range arr.Elems {
			if int(arr.I(j).Int(img)) == ci {
				at = j
				break
			}
		}
		if at < 0 {
			return
		}
		for j := at; j < cnt-1; j++ {
			arr.I(j).SetInt(img, arr.I(j+1).Int(img))
		}
		if cnt-1 >= 0 && cnt-1 < len(arr.Elems) {
			arr.I(cnt-1).SetInt(img, 0xFFFF)
		} else if cnt-1 < 0 {
			arr.I(len(arr.Elems)+cnt-1).SetInt(img, 0xFFFF)
		}
		num.SetInt(img, int64(cnt-1)&(1<<(8*num.Size)-1))
	}
	z := b.F("zoneinfo")
	n := int(z.F("zonenum").Int(img))
	if n > 16 {
		n = 16
	}
	for k := 0; k < n; k++ {
		x := int(z.F("zoneindex").I(k).Int(img))
		if x == 0xFFFF || x >= 16 {
			break
		}
		zz := z.F("zones").I(x)
		remove(zz.F("chindex"), zz.F("chnum"))
	}
	s := b.F("scans")
	n = int(s.F("scannum").Int(img))
	if n > 16 {
		n = 16
	}
	for k := 0; k < n; k++ {
		x := int(s.F("scanindex").I(k).Int(img))
		if x == 0xFFFF || x >= 16 {
			break
		}
		sc := s.F("scans").I(x)
		if int(sc.F("specifych").Int(img)) == ci {
			sc.F("specifych").SetInt(img, int64(next))
		}
		for _, p := range []string{"PriorityCh1", "PriorityCh2"} {
			if int(sc.F(p).Int(img)) == ci {
				sc.F(p).SetInt(img, 0xFFFF)
			}
		}
		remove(sc.F("chindex"), sc.F("chnum"))
	}
	a := b.F("alarms")
	n = int(a.F("alarmnum").Int(img))
	if n > 8 {
		n = 8
	}
	for k := 0; k < n; k++ {
		x := int(a.F("alarmindex").I(k).Int(img))
		if x == 0xFFFF || x >= 8 {
			break
		}
		if int(a.F("alarms").I(x).F("jumpch").Int(img)) == ci {
			a.F("alarms").I(x).F("jumpch").SetInt(img, 0)
		}
	}
}

func (f ha1gFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := ha1go(m)
	if c == nil && o.kind == haG1 && i < 30 {
		return nil // come CHIRP: i canali GMRS fissi non si cancellano
	}
	var tt, rt genTone
	if c != nil {
		if c.TxOff {
			return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
		}
		if d := c.Tx - c.Rx; d > 70000000 || d < -70000000 {
			return fmt.Errorf("questa radio non ammette spostamenti oltre 70 MHz (split)")
		}
		var err error
		if tt, err = genToneParse(c.TxTone); err != nil {
			return err
		}
		if rt, err = genToneParse(c.RxTone); err != nil {
			return err
		}
	}
	f.haMarkItems(m, img)
	ci := f.chIndex(m, i)
	r := f.rec(m, i)
	l := f.list(m, img)
	wasEmpty := !intsHas(l, ci)
	// opzioni del canale come le legge CHIRP (anche da un canale vuoto)
	alarms, dtmfs := f.items(m, img, true), f.items(m, img, false)
	tot := r.F("tottime").Int(img)
	if tot == 0 {
		tot = 63
	}
	sq := r.F("rxsqlmode").Int(img)
	dt := int64(haMap(dtmfs, int(r.F("dtmfsignalinglist").Int(img))))
	if o.kind == haA2 && wasEmpty {
		tot, sq, dt = 36, 4, int64(haMap(dtmfs, 15))
	}
	keep := map[string]int64{"tottime": tot, "totpermissions": r.F("totpermissions").Int(img), "rxsqlmode": sq,
		"alarmlist": int64(haMap(alarms, int(r.F("alarmlist").Int(img)))), "dtmfsignalinglist": dt,
		"offlineorreversal": r.F("offlineorreversal").Int(img), "vox": r.F("vox").Int(img),
		"companding": r.F("companding").Int(img)}
	if o.kind == haA2 {
		fq := r.F("scrambleFreq").Int(img)
		if r.F("scramble").Int(img) == 0 || fq > 9 {
			keep["scrambleFreq"], keep["scramble"] = 255, 0
		} else {
			keep["scrambleFreq"], keep["scramble"] = fq, 1
		}
	} else {
		keep["autoscan"] = r.F("autoscan").Int(img)
		keep["scramble"] = r.F("scramble").Int(img)
	}
	var rxw int64
	if c != nil {
		rxw = c.Rx
		if o.kind != haA2 {
			rxw = haClampRx(c.Rx)
		}
	}
	switch in := intsHas(l, ci); {
	case !in && rxw != 0 && rxw != 0xFFFFFFFF:
		l = append(l, ci)
	case in && rxw <= 0:
		for k, v := range l {
			if v == ci {
				l = append(l[:k:k], l[k+1:]...)
				break
			}
		}
		f.removeRefs(m, img, ci, l)
	}
	f.setList(m, img, l)
	r.Fill(img, 0)
	if c == nil {
		if o.kind == haA2 {
			r.F("dtmfsignalinglist").SetInt(img, 15)
			r.F("tottime").SetInt(img, 36)
			r.F("rxsqlmode").SetInt(img, 4)
		}
		return nil
	}
	r.F("rxfreq").SetInt(img, rxw)
	nm := r.F("alias")
	for k := 0; k < 14; k++ {
		v := byte(' ')
		if k < len(c.Name) {
			v = c.Name[k]
		}
		img[nm.Off+k] = v
	}
	base := c.Rx
	if o.kind != haA2 {
		base = rxw
	}
	r.F("txfreq").SetInt(img, base+(c.Tx-c.Rx))
	if c.Narrow {
		r.F("bandwidth").SetInt(img, 1)
	} else {
		r.F("bandwidth").SetInt(img, 3)
	}
	if m.powerIndex(c.Power) == 1 {
		r.F("power").SetInt(img, 2)
	}
	enc := func(t genTone, typ, val string) {
		switch t.Mode {
		case "Tone":
			r.F(typ).SetInt(img, 1)
			r.F(val).SetInt(img, int64(t.Hz10))
		case "DTCS":
			if t.Inv {
				r.F(typ).SetInt(img, 3)
			} else {
				r.F(typ).SetInt(img, 2)
			}
			r.F(val).SetInt(img, int64(octValue(t.Code)))
		}
	}
	enc(rt, "rxctcvaluetype", "rxctc")
	enc(tt, "txctcvaluetype", "txctc")
	for k, v := range keep {
		r.F(k).SetInt(img, v)
	}
	return nil
}

// lockedSlotImg / Locked (HA1G): canali GMRS 1-30 fissi (immutable di CHIRP)
func (f ha1gFam) lockedSlotImg(m *genModel, img []byte, i int) bool {
	return ha1go(m).kind == haG1 && i < 30
}

func (f ha1gFam) Locked(m *genModel, i int, old, want *Chan) error {
	if ha1go(m).kind != haG1 || i >= 30 {
		return nil
	}
	n := i + 1
	switch {
	case want == nil:
		return fmt.Errorf("il canale %d è un canale GMRS fisso: non si può cancellare", n)
	case old == nil:
		return fmt.Errorf("il canale %d è una posizione GMRS fissa: non si può riempire", n)
	case want.Rx != old.Rx:
		return fmt.Errorf("il canale %d è il canale GMRS %s MHz: frequenza non modificabile", n, fmtMHz(old.Rx))
	case want.Tx-want.Rx != old.Tx-old.Rx || want.TxOff != old.TxOff:
		return fmt.Errorf("sul canale GMRS %d il passo è fisso", n)
	case n >= 8 && n <= 14 && (want.Narrow != old.Narrow || want.Power != old.Power):
		return fmt.Errorf("il canale GMRS %d ha banda e potenza fisse", n)
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func crc16IBMRev(b []byte) uint16 {
	var crc uint16
	for _, x := range b {
		crc ^= uint16(x)
		for k := 0; k < 8; k++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func haPacket(dtype, code byte, idx, count int, data []byte) []byte {
	p := append([]byte("RDTP\x01\x00\x00\x00\x00\x00\x00\x00"), 0, 0, dtype, code, 0, 0, 0, 0)
	binary.LittleEndian.PutUint16(p[12:], uint16(len(data)+6))
	binary.LittleEndian.PutUint16(p[16:], uint16(idx))
	binary.LittleEndian.PutUint16(p[18:], uint16(count))
	p = append(p, data...)
	c := crc16IBMRev(p)
	return append(p, byte(c), byte(c>>8), 0xFF)
}

// haRead: un pacchetto della radio (intestazione di 20 byte, dati, CRC, FF)
func haRead(ctx context.Context, p Port, timeout time.Duration) ([]byte, error) {
	h := readUpTo(ctx, p, 20, timeout)
	if ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(h) < 20 {
		return h, nil
	}
	n := int(binary.LittleEndian.Uint16(h[12:])) - 6
	if n < 0 {
		n = 0
	}
	rest := readUpTo(ctx, p, n+3, timeout)
	pkt := append(h, rest...)
	if len(rest) != n+3 || crc16IBMRev(pkt[:len(pkt)-3]) != binary.LittleEndian.Uint16(pkt[len(pkt)-3:]) {
		return pkt, rErr("pacchetto della radio non valido")
	}
	return pkt, nil
}

type haSess struct {
	*rawSess
	m *genModel
}

func (s *haSess) xfer(pkt []byte, item byte) ([]byte, error) {
	if err := s.send(pkt); err != nil {
		return nil, err
	}
	r, err := haRead(s.ctx, s.p, time.Second)
	if err != nil || len(r) <= 14 {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("la radio ha segnalato un errore")
	}
	if r[14] != item {
		return nil, rErr("la radio ha segnalato un errore")
	}
	return r, nil
}

func (s *haSess) handshake() error {
	o := ha1go(s.m)
	data := make([]byte, 41)
	copy(data, o.model+" ")
	hs := haPacket(0, 0, 0, 1, data)
	var r []byte
	for k := 0; k < 15; k++ {
		if err := s.send(hs); err != nil {
			return err
		}
		var err error
		r, err = haRead(s.ctx, s.p, time.Second)
		if err == errCancelled {
			return err
		}
		if err := s.sleep(50 * time.Millisecond); err != nil {
			return err
		}
		if err == nil && len(r) > 20 {
			break
		}
	}
	switch {
	case len(r) == 0:
		return rErr("la radio non risponde")
	case len(r) <= 20 || !bytes.Equal(r[14:16], []byte{0x00, 0x01}):
		return rErr("errore di comunicazione con la radio")
	case r[20] == 1:
		return rErr("la radio è protetta da password")
	case !bytes.HasPrefix(r[20:], []byte(o.model)):
		return rErr("modello della radio diverso")
	}
	return nil
}

func (s *haSess) readItem(id byte) ([]byte, error) {
	var out []byte
	count := 1
	for k := 0; k < count; k++ {
		r, err := s.xfer(haPacket(id, 2, 0, 1, []byte{byte(k), byte(k >> 8)}), id)
		if err != nil {
			return nil, err
		}
		if k == 0 {
			count = int(binary.LittleEndian.Uint16(r[18:]))
		}
		out = append(out, r[20:len(r)-3]...)
	}
	return out, nil
}

func (s *haSess) writeItem(id byte, data []byte) error {
	n := (len(data) + 1023) / 1024
	for k := 0; k < n; k++ {
		end := (k + 1) * 1024
		if end > len(data) {
			end = len(data)
		}
		if _, err := s.xfer(haPacket(id, 0, k, n, data[k*1024:end]), id); err != nil {
			return err
		}
	}
	return nil
}

func haVersion(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	return fmt.Sprintf("v%02d.%02d.%02d.%03d", b[2], b[3], b[4], b[5])
}

// haVerCmp: confronto delle versioni come compare_version di CHIRP
func haVerCmp(a, b string) int {
	pa := strings.Split(strings.TrimLeft(a, "v"), ".")
	pb := strings.Split(strings.TrimLeft(b, "v"), ".")
	for k := 0; k < len(pa) || k < len(pb); k++ {
		var x, y int
		if k < len(pa) {
			fmt.Sscanf(pa[k], "%d", &x)
		} else {
			return -1
		}
		if k < len(pb) {
			fmt.Sscanf(pb[k], "%d", &y)
		} else {
			return 1
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
}

func (s *haSess) checkVer(item []byte) (string, error) {
	v := haVersion(item)
	if haVerCmp(v, ha1go(s.m).reqVer) < 0 {
		return v, rErr("il firmware della radio è %s: serve il %s o successivo", v, ha1go(s.m).reqVer)
	}
	return v, nil
}

func haOpen(ctx context.Context, m *genModel, port string) (*haSess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &haSess{rawSess: s, m: m}, nil
}

func (s *haSess) exit() { s.send(haPacket(111, 0, 0, 1, []byte{0})) }

func (f ha1gFam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	o := ha1go(m)
	s, err := haOpen(ctx, m, port)
	if err != nil {
		return nil, nil, "", err
	}
	defer s.p.Close()
	defer s.exit()
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.handshake(); err != nil {
		return nil, nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, o.size)
	meta := map[string]interface{}{}
	done := 0
	for _, rg := range o.regions {
		b, err := s.readItem(rg.id)
		if err != nil {
			return nil, nil, "", genFail("Lettura ("+rg.name+")", err)
		}
		if rg.name == "radioVer" && len(b) > 0 {
			v, err := s.checkVer(b)
			if err != nil {
				return nil, nil, "", genFail("Versione del firmware", err)
			}
			meta["ha1g_firmware"] = v
		}
		n := len(b)
		if n > rg.len {
			n = rg.len
		}
		copy(img[rg.start:], b[:n])
		done += rg.len
		prog(done, o.size, "Lettura…")
	}
	prog(o.size, o.size, "Lettura completata")
	note := ""
	if v, ok := meta["ha1g_firmware"].(string); ok {
		note = "firmware " + v
	}
	return img, meta, note, nil
}

func (f ha1gFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	img, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return img, note, err
}

func (f ha1gFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := ha1go(m)
	if len(img) < o.size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.size)
	}
	s, err := haOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	defer s.exit()
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.handshake(); err != nil {
		return "", genFail("Identificazione", err)
	}
	ver := ""
	for _, rg := range o.regions {
		if rg.name != "radioVer" {
			continue
		}
		b, err := s.readItem(rg.id)
		if err != nil {
			return "", genFail("Versione del firmware", err)
		}
		if len(b) > 0 {
			if ver, err = s.checkVer(b); err != nil {
				return "", genFail("Versione del firmware", err)
			}
		}
	}
	// zone: HA1G e HA1UV dal firmware 1.1.13.1 (supports_banks di CHIRP)
	banks := o.kind == haA2 || haVerCmp(ver, "1.1.13.1") >= 0
	done := 0
	for _, rg := range o.regions {
		if o.exclude[rg.name] || rg.name == "zoneData" && !banks {
			continue
		}
		if err := s.writeItem(rg.id, img[rg.start:rg.start+rg.len]); err != nil {
			return "", genFail("Scrittura ("+rg.name+")", err)
		}
		done += rg.len
		prog(done, o.size, "Scrittura…")
	}
	prog(o.size, o.size, "Scrittura completata")
	return "firmware " + ver, nil
}
