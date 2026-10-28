// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Formato .dat della Bajeton 8300 (software BT-11_CPS): oggetto KDH.RadioData
// serializzato con il BinaryFormatter .NET. I canali sono in
// channelData: KDH.Channel[15][64]; frequenze, toni e nome sono stringhe.
// Il programma modifica solo i campi che cambiano davvero: un file aperto e
// salvato senza modifiche resta identico byte per byte.

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const datZones, datPerZone = 15, 64

// campi interi conservati ma non modificabili dal programma
var datKeepInts = []string{"signallingGroup", "pttId", "scram", "encrypt", "busyLockout", "rxModulation", "learnFHSS"}

func isDAT(b []byte) bool {
	return len(b) > 32 && b[0] == 0 && bytes.Contains(b[:min(len(b), 4096)], []byte("KDH.RadioData"))
}

type datSlot struct {
	rec  *Rec
	meta *Rec
	idx  map[string]int
}

func datSlots(d *Doc) ([]*datSlot, error) {
	root := d.Objs[1]
	if root == nil || d.MetaOf(root) == nil || d.MetaOf(root).Name != "KDH.RadioData" {
		return nil, fmt.Errorf("non è un file Bajeton/BT-11 valido (KDH.RadioData mancante)")
	}
	m := d.MetaOf(root)
	ci := -1
	for i, n := range m.Names {
		if n == "channelData" {
			ci = i
		}
	}
	if ci < 0 {
		return nil, fmt.Errorf("sezione canali non trovata")
	}
	outer := d.Resolve(root.Vals[ci])
	if outer == nil || outer.T != 7 {
		return nil, fmt.Errorf("struttura canali inattesa")
	}
	els, _, err := datChannelVals(d, outer)
	if err != nil {
		return nil, err
	}
	var slots []*datSlot
	for _, ce := range els {
		c := d.Resolve(ce)
		if c == nil || (c.T != 1 && c.T != 5) {
			return nil, fmt.Errorf("canale mancante nel file: struttura non supportata")
		}
		mm := d.MetaOf(c)
		s := &datSlot{rec: c, meta: mm, idx: map[string]int{}}
		for i, n := range mm.Names {
			s.idx[n] = i
		}
		for _, need := range []string{"rxFreq", "txFreq", "rxQT", "txQT", "txPower", "bandWide", "scanAdd", "chName"} {
			if _, ok := s.idx[need]; !ok {
				return nil, fmt.Errorf("campo %s mancante: versione del software non supportata", need)
			}
		}
		slots = append(slots, s)
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("nessun canale nel file")
	}
	return slots, nil
}

// datChannelVals: i 960 canali del file nell'ordine della memoria della radio.
// BT-11_CPS fino alla v1.0.8 salva channelData come Channel[15][64] (15 zone da
// 64); dalla v1.0.9 come Channel[960] (flat = true). Le due versioni non aprono
// i file dell'altra: IU0VYT Flash le gestisce entrambe e conserva il formato.
func datChannelVals(d *Doc, outer *Rec) ([]Val, bool, error) {
	if len(outer.Els) == datZones*datPerZone {
		if c := d.Resolve(outer.Els[0]); c != nil && (c.T == 1 || c.T == 5) {
			return outer.Els, true, nil
		}
	}
	if len(outer.Els) != datZones {
		return nil, false, fmt.Errorf("struttura delle zone non supportata")
	}
	var out []Val
	for _, ze := range outer.Els {
		if ze.R == nil || ze.R.T == 13 || ze.R.T == 14 || ze.R.T == 10 {
			return nil, false, fmt.Errorf("zona vuota nel file: struttura non supportata")
		}
		z := d.Resolve(ze)
		if z == nil || z.T != 7 || len(z.Els) != datPerZone {
			return nil, false, fmt.Errorf("struttura zona inattesa")
		}
		out = append(out, z.Els...)
	}
	return out, false, nil
}

func (s *datSlot) str(d *Doc, f string) string {
	i, ok := s.idx[f]
	if !ok {
		return ""
	}
	return d.StrOf(s.rec.Vals[i])
}
func (s *datSlot) int(f string) int {
	i, ok := s.idx[f]
	if !ok {
		return 0
	}
	v := s.rec.Vals[i]
	if !v.Prim || len(v.Raw) != 4 {
		return 0
	}
	return int(int32(binary.LittleEndian.Uint32(v.Raw)))
}
func (s *datSlot) setStr(d *Doc, f, val string) {
	if i, ok := s.idx[f]; ok {
		if r := d.Resolve(s.rec.Vals[i]); r != nil && r.T == 6 && r.S == val {
			return // invariato
		}
		s.rec.Vals[i] = d.NewString(val)
	}
}
func (s *datSlot) setInt(f string, val int) {
	i, ok := s.idx[f]
	if !ok || !s.rec.Vals[i].Prim || len(s.rec.Vals[i].Raw) != 4 {
		return
	}
	raw := make([]byte, 4)
	binary.LittleEndian.PutUint32(raw, uint32(int32(val)))
	s.rec.Vals[i].Raw = raw
}

func cleanDatName(s string) string { return string(bytes.TrimRight([]byte(s), "\x00 ")) }

// datTone: stringa del file -> stringa del modello
func datTone(s string) string {
	t, ok := parseTone(s)
	if ok {
		return toneStr(t)
	}
	return strings.TrimSpace(s) // valore non CTCSS (es. DCS): conservato
}

type datRawT struct {
	Ints map[string]int `json:"i"`
	FHSS string         `json:"f,omitempty"`
}

func datRaw(d *Doc, s *datSlot) string {
	r := datRawT{Ints: map[string]int{}, FHSS: s.str(d, "fhssCode")}
	for _, f := range datKeepInts {
		if _, ok := s.idx[f]; ok {
			r.Ints[f] = s.int(f)
		}
	}
	b, _ := json.Marshal(r)
	return "dat:" + string(b)
}

func datUnraw(s string) (datRawT, bool) {
	var r datRawT
	if !strings.HasPrefix(s, "dat:") {
		return r, false
	}
	if json.Unmarshal([]byte(s[4:]), &r) != nil {
		return r, false
	}
	return r, true
}

func datSlotChan(d *Doc, s *datSlot) *Chan {
	rx, ok := parseMHz(s.str(d, "rxFreq"))
	if !ok || rx == 0 {
		return nil
	}
	tx, ok := parseMHz(s.str(d, "txFreq"))
	if !ok || tx == 0 {
		tx = rx
	}
	p := s.int("txPower")
	if p < 0 || p > 2 {
		p = 0
	}
	return &Chan{Rx: rx, Tx: tx, Power: p, Narrow: s.int("bandWide") == 1, Scan: s.int("scanAdd") != 0,
		Name: cleanDatName(s.str(d, "chName")), RxTone: datTone(s.str(d, "rxQT")), TxTone: datTone(s.str(d, "txQT"))}
}

func readDAT(b []byte, rep *Report) ([]*Chan, error) {
	d, err := ParseNRBF(b)
	if err != nil {
		return nil, err
	}
	slots, err := datSlots(d)
	if err != nil {
		return nil, err
	}
	chs := make([]*Chan, len(slots))
	for i, s := range slots {
		c := datSlotChan(d, s)
		if c == nil {
			continue
		}
		for k, t := range []string{c.RxTone, c.TxTone} {
			if _, ok := parseTone(t); !ok && btDCSIndex(t) < 0 {
				rep.warn("Bajeton zona %d n. %d (%s): tono %s \"%s\" non valido: verrà scritto OFF", i/datPerZone+1, i%datPerZone+1, c.Name, []string{"RX", "TX"}[k], t)
			}
		}
		c.Raw = datRaw(d, s)
		chs[i] = c
	}
	return chs, nil
}

func datToneOut(t string) (string, bool) {
	v, ok := parseTone(t)
	if ok {
		return fmtTone(v), true
	}
	if btDCSIndex(t) >= 0 {
		return t, true
	}
	return "OFF", false
}

func writeDAT(base []byte, chs []*Chan, rep *Report) ([]byte, error) {
	d, err := ParseNRBF(base)
	if err != nil {
		return nil, fmt.Errorf("file base Bajeton non leggibile: %v", err)
	}
	old := map[int32]*Rec{}
	for k, v := range d.Objs {
		old[k] = v
	}
	slots, err := datSlots(d)
	if err != nil {
		return nil, err
	}
	extra := 0
	for i, c := range chs {
		if i >= len(slots) && c != nil {
			extra++
		}
	}
	if extra > 0 {
		rep.warn("%d canali oltre la capacità della Bajeton (%d) esclusi", extra, len(slots))
	}
	for i, s := range slots {
		var c *Chan
		if i < len(chs) {
			c = chs[i]
		}
		cur := datSlotChan(d, s)
		if c == nil {
			if cur == nil && s.str(d, "rxFreq") == "" {
				continue // già vuoto
			}
			s.setStr(d, "rxFreq", "")
			s.setStr(d, "txFreq", "")
			s.setStr(d, "rxQT", "OFF")
			s.setStr(d, "txQT", "OFF")
			for _, f := range []string{"signallingGroup", "pttId", "txPower", "scram", "bandWide", "encrypt",
				"busyLockout", "scanAdd", "rxModulation", "learnFHSS"} {
				s.setInt(f, 0)
			}
			s.setStr(d, "fhssCode", "")
			s.setStr(d, "chName", "")
			continue
		}
		pos := fmt.Sprintf("Bajeton zona %d n. %d (%s)", i/datPerZone+1, i%datPerZone+1, c.Name)
		if c.Rx <= 0 || c.Tx <= 0 || c.Rx >= 10000000000 || c.Tx >= 10000000000 {
			rep.warn("%s: frequenza non valida, canale non scritto", pos)
			continue
		}
		if cur == nil || cur.Rx != c.Rx {
			s.setStr(d, "rxFreq", fmtMHz(c.Rx))
		}
		if cur == nil || cur.Tx != c.Tx {
			s.setStr(d, "txFreq", fmtMHz(c.Tx))
		}
		for _, x := range []struct{ f, v, curv, lbl string }{{"rxQT", c.RxTone, "", "RX"}, {"txQT", c.TxTone, "", "TX"}} {
			cv := ""
			if cur != nil {
				if x.f == "rxQT" {
					cv = cur.RxTone
				} else {
					cv = cur.TxTone
				}
			}
			if cur != nil && cv == x.v {
				continue
			}
			v, ok := datToneOut(x.v)
			if !ok {
				rep.warn("%s: tono %s \"%s\" non gestito sulla Bajeton, impostato OFF", pos, x.lbl, x.v)
			}
			s.setStr(d, x.f, v)
		}
		if r, ok := datUnraw(c.Raw); ok {
			for _, f := range datKeepInts {
				if v, ok := r.Ints[f]; ok {
					s.setInt(f, v)
				}
			}
			s.setStr(d, "fhssCode", r.FHSS)
		} else {
			for _, f := range datKeepInts {
				s.setInt(f, 0)
			}
			s.setStr(d, "fhssCode", "")
		}
		p := c.Power
		if p < 0 || p > 2 {
			p = 0
		}
		s.setInt("txPower", p)
		bw := 0
		if c.Narrow {
			bw = 1
		}
		s.setInt("bandWide", bw)
		sc := 0
		if c.Scan {
			sc = 1
		}
		s.setInt("scanAdd", sc)
		name := asciiName(c.Name)
		if sn := shortName(name, 12); sn != name {
			rep.warn("%s: nome \"%s\" accorciato in \"%s\"", pos, name, sn)
			name = sn
		}
		if cur == nil || cur.Name != name {
			s.setStr(d, "chName", name)
		}
	}
	if err := d.FixDangling(old); err != nil {
		return nil, err
	}
	out := d.Dump()
	// controllo: il file prodotto deve essere rileggibile
	if _, err := ParseNRBF(out); err != nil {
		return nil, fmt.Errorf("errore interno, file non valido: %v", err)
	}
	return out, nil
}

func blankDAT(base []byte) ([]byte, error) { return writeDAT(base, nil, &Report{}) }

// datVer: versione del programma BT-11_CPS che apre il file ("1.0.9" se i canali
// sono in Channel[960], altrimenti "1.0.8")
func datVer(b []byte) string {
	d, err := ParseNRBF(b)
	if err != nil {
		return ""
	}
	root := d.Objs[1]
	if root == nil || d.MetaOf(root) == nil {
		return ""
	}
	for i, n := range d.MetaOf(root).Names {
		if n == "channelData" {
			if outer := d.Resolve(root.Vals[i]); outer != nil {
				if _, flat, err := datChannelVals(d, outer); err == nil {
					if flat {
						return "1.0.9"
					}
					return "1.0.8"
				}
			}
		}
	}
	return ""
}

func datTemplate(ver string) []byte {
	if ver == "1.0.9" {
		return baseDAT109
	}
	return baseDAT
}

// datToVer converte il file nel formato della versione indicata del programma
// (tutto il contenuto passa per la memoria della radio, come in lettura e scrittura)
func datToVer(b []byte, ver string) ([]byte, error) {
	if datVer(b) == ver {
		return b, nil
	}
	img, err := datToEEPROM(b)
	if err != nil {
		return nil, err
	}
	return eepromToDAT(img, datTemplate(ver))
}

// bajTemplate: modello dei nuovi file .dat (formato scelto dall'utente)
func (a *App) bajTemplate() []byte {
	return datTemplate(a.store.Settings.BajVer)
}

func (a *App) apiDatVer(r *http.Request) (interface{}, error) {
	var req struct {
		Ver string `json:"ver"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Ver != "1.0.8" && req.Ver != "1.0.9" {
		return nil, fmt.Errorf("versione non valida: %q", req.Ver)
	}
	a.mu.Lock()
	sess := a.radios[kindDAT]
	var base []byte
	if sess != nil {
		base = sess.Base
	}
	a.mu.Unlock()
	if base != nil {
		out, err := datToVer(base, req.Ver)
		if err != nil {
			return nil, fmt.Errorf("conversione non riuscita: %v", err)
		}
		a.mu.Lock()
		if a.radios[kindDAT] == sess {
			sess.Base = out
		}
		a.mu.Unlock()
	}
	st := a.store.Settings
	st.BajVer = req.Ver
	_ = a.store.SaveSettings(st)
	return map[string]interface{}{"ver": req.Ver}, nil
}
