// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// kenwood_live.py, kenwood_d7.py (4.17).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.17: palmari Kenwood che si programmano con i comandi del PC, un canale per
// volta (in CHIRP «live mode»): TH-F6A, TH-F7E, TH-G71, TH-D7 / TH-D7G, TH-K2.
//
// La radio accesa normalmente, col cavo PG-4Y (o equivalente) e la porta del PC
// a 9600 baud 8N1, risponde a righe chiuse da CR:
//   ID                      → «ID TH-F7»
//   AI 0                    spegne i messaggi automatici
//   TH-F6 / TH-F7   MR 0,nnn → «MR 0,nnn,frequenza(11),passo,spostamento,inverso,
//                   tono,CTCSS,DCS,n. tono(2),n. CTCSS(2),n. DCS(3),scostamento(9),
//                   modo,escluso»; MR 1,nnn → frequenza di trasmissione separata;
//                   MNA nnn → «MNA nnn,nome»; MW e MNA con gli stessi campi scrivono
//   TH-G71 / TH-D7  MR 0,0,nnn → «MR 0,0,nnn,frequenza,passo,spostamento,inverso,
//                   tono,CTCSS,,codice tono(2),,codice CTCSS(2),scostamento(9)
//                   [,modo],escluso»; MR 0,1,nnn; MNA 0,nnn
//   TH-K2           ME nn → «ME nn,frequenza(10),0,spostamento,inverso,tono,CTCSS,
//                   DCS,n. tono,n. CTCSS,n. DCS,scostamento(8),stretta,0,
//                   0000000000,0,escluso»; MN nn → «MN nn,nome»
// Un canale vuoto risponde «N»; la scrittura senza campi lo cancella.
//
// L'immagine del programma (formato di IU0VYT Flash, CHIRP non ha file per queste
// radio): 32 byte di intestazione e un blocco di 128 byte per canale con il testo
// della risposta della radio, il nome e la frequenza di trasmissione separata.
// In scrittura la radio viene riletta e si inviano solo i canali cambiati.

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	kwlHead    = 32
	kwlRec     = 128
	kwlSpecAt  = 4
	kwlSpecLen = 80
	kwlNameAt  = 84
	kwlNameLen = 16
	kwlSplitAt = 100
	kwlMagic   = "IU0VYT-KENWOOD-LIVE\x00"
)

// attesa tra la scrittura del canale e quella del nome (come CHIRP)
var kwlPause = 500 * time.Millisecond

const (
	kwlF6 = iota
	kwlG71
	kwlD7
	kwlK2
)

// posizione dei campi nel testo del canale (-1: il campo non c'è)
type kwlIdx struct {
	n                                int // numero dei campi
	num, freq, step, shift, rev      int
	tone, ctcss, dcs                 int
	tcode, ccode, dcode, offs        int
	mode, lock                       int
	numW, freqW, offsW, codeW, codeB int // cifre dei campi; codeB: primo codice dei toni
}

var kwlIdxs = map[int]*kwlIdx{
	kwlF6:  {n: 15, num: 1, freq: 2, step: 3, shift: 4, rev: 5, tone: 6, ctcss: 7, dcs: 8, tcode: 9, ccode: 10, dcode: 11, offs: 12, mode: 13, lock: 14, numW: 3, freqW: 11, offsW: 9, codeW: 2},
	kwlG71: {n: 15, num: 2, freq: 3, step: 4, shift: 5, rev: 6, tone: 7, ctcss: 8, dcs: -1, tcode: 10, ccode: 12, dcode: -1, offs: 13, mode: -1, lock: 14, numW: 3, freqW: 11, offsW: 9, codeW: 2, codeB: 1},
	kwlD7:  {n: 16, num: 2, freq: 3, step: 4, shift: 5, rev: 6, tone: 7, ctcss: 8, dcs: -1, tcode: 10, ccode: 12, dcode: -1, offs: 13, mode: 14, lock: 15, numW: 3, freqW: 11, offsW: 9, codeW: 2, codeB: 1},
	kwlK2:  {n: 17, num: 0, freq: 1, step: 2, shift: 3, rev: 4, tone: 5, ctcss: 6, dcs: 7, tcode: 8, ccode: 9, dcode: 10, offs: 11, mode: 12, lock: 16, numW: 2, freqW: 10, offsW: 8, codeW: 2},
}

type kwlOpt struct {
	id      string // risposta al comando ID
	kind    int
	tones   []int // subtoni in decimi di Hz, nell'ordine della radio
	steps   []int // passi in Hz (8333: 8,33 kHz)
	maxOffs int64 // scostamento massimo del ponte
	split   bool  // frequenza di trasmissione separata
	bauds   []int
}

type kwlFam struct{}

// subtoni delle Kenwood recenti (42) e di quelle vecchie (38 più il 69,3 Hz che
// la radio non usa)
var kwlTones = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230, 1273,
	1318, 1365, 1413, 1462, 1514, 1567, 1622, 1679, 1738, 1799, 1862, 1928, 2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}
var kwlOldTones = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230, 1273,
	1318, 1365, 1413, 1462, 1514, 1567, 1622, 1679, 1738, 1799, 1862, 1928, 2035, 2107, 2181, 2257, 2336, 2418, 2503}

var kwlStepsF6 = []int{5000, 6250, 8333, 9000, 10000, 12500, 15000, 20000, 25000, 30000, 50000, 100000}
var kwlStepsOld = []int{5000, 6250, 10000, 12500, 15000, 20000, 25000, 30000, 50000, 100000}

const (
	kwlCharsF6 = " !\"#$%&'()*+-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	kwlCharsG7 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-/ "
	kwlCharsD7 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz 1234567890/.-+*)('&%$#!~}|{"
)

func init() {
	const how = "Si programma con la radio accesa normalmente (non in modalità di copia), col cavo di programmazione e la porta " +
		"del PC della radio a 9600 baud: il programma legge e scrive un canale per volta con i comandi del PC. Prima di " +
		"scrivere la radio viene riletta e si inviano solo i canali cambiati. "
	const keep = "Passo e modo di un canale restano quelli impostati finché la frequenza non cambia. Comandi dei driver di CHIRP, " +
		"provati con una radio simulata."
	one := []genPower{{"Unica", 5, 0}}
	mk := func(id, model, name string, slots, nameMax int, charset string, bands []band, narrow bool, dcs []int, o kwlOpt, notes string) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "Kenwood", Model: model, RClass: "IU0VYT_Kenwood_" + o.id, Name: name,
			Slots: slots, First: 0, NameMax: nameMax, Charset: charset, Powers: one, Bands: bands, Narrow: narrow, HasScan: true,
			NoTxOff: true, NoChirp: true, DCS: dcs, Baud: 9600, Fam: kwlFam{}, Opt: &oo, Notes: notes}
	}
	all := []int{9600, 4800, 19200, 38400, 57600}
	f6 := kwlOpt{kind: kwlF6, tones: kwlTones, steps: kwlStepsF6, maxOffs: 59950000, split: true, bauds: all}
	f6a, f7 := f6, f6
	f6a.id, f7.id = "TH-F6", "TH-F7"
	wide := []band{{100000, 1300000000}}
	f6note := how + "400 canali (0-399); subtono in trasmissione, subtono o codice DCS uguale nei due versi; ponti fino a 59,95 MHz " +
		"di scostamento, oltre si usa la frequenza di trasmissione separata. " + keep
	oldNote := how + "200 canali (0-199); subtono in trasmissione o uguale nei due versi (niente DCS); ponti fino a 29,95 MHz di " +
		"scostamento, oltre si usa la frequenza di trasmissione separata. " + keep
	d7 := kwlOpt{kind: kwlD7, tones: kwlOldTones, steps: kwlStepsOld, maxOffs: 29950000, split: true, bauds: []int{9600}}
	d7a, d7g := d7, d7
	d7a.id, d7g.id = "TH-D7", "TH-D7G"
	d7bands := []band{{118000000, 174000000}, {400000000, 470000000}}
	genRegister(
		mk("kw-thf6", "TH-F6", "Kenwood TH-F6A", 400, 8, kwlCharsF6, wide, false, dtcsCodes, f6a, f6note),
		mk("kw-thf7", "TH-F7", "Kenwood TH-F7E", 400, 8, kwlCharsF6, wide, false, dtcsCodes, f7, f6note),
		mk("kw-thg71", "TH-G71", "Kenwood TH-G71A / TH-G71E", 200, 6, kwlCharsG7,
			[]band{{118000000, 174000000}, {320000000, 470000000}, {800000000, 949000000}}, false, nil,
			kwlOpt{id: "TH-G71", kind: kwlG71, tones: kwlOldTones, steps: kwlStepsOld, maxOffs: 29950000, split: true, bauds: all}, oldNote),
		mk("kw-thd7", "TH-D7", "Kenwood TH-D7A / TH-D7E", 200, 8, kwlCharsD7, d7bands, false, nil, d7a, oldNote),
		mk("kw-thd7g", "TH-D7G", "Kenwood TH-D7A(G) / TH-D7E(G)", 200, 8, kwlCharsD7, d7bands, false, nil, d7g, oldNote),
		mk("kw-thk2", "TH-K2", "Kenwood TH-K2AT / TH-K2E / TH-K2ET", 100, 6, kwlCharsG7,
			[]band{{136000000, 174000000}}, true, dtcsCodes,
			kwlOpt{id: "TH-K2", kind: kwlK2, tones: kwlTones, maxOffs: 69950000, bauds: all},
			how+"100 canali (0-99); i nomi ci sono solo se la radio è impostata a 50 canali con nome (in quel caso i canali da 50 "+
				"in su non esistono e la radio li rifiuta). Subtono in trasmissione, subtono o codice DCS uguale nei due versi; ponti "+
				"fino a 69,95 MHz di scostamento. Comandi del driver di CHIRP, provati con una radio simulata."),
	)
}

func kwlo(m *genModel) *kwlOpt { return m.Opt.(*kwlOpt) }

func kwlSize(m *genModel) int { return kwlHead + m.Slots*kwlRec }

// kwlBlank: immagine senza canali
func kwlBlank(m *genModel) []byte {
	img := make([]byte, kwlSize(m))
	copy(img, kwlMagic)
	copy(img[24:32], kwlo(m).id)
	return img
}

func (kwlFam) rec(img []byte, i int) []byte { return img[kwlHead+i*kwlRec : kwlHead+(i+1)*kwlRec] }

func (f kwlFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}
func (f kwlFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }
func (f kwlFam) Match(m *genModel, img []byte) bool {
	return len(img) == kwlSize(m) && bytes.HasPrefix(img, []byte(kwlMagic)) &&
		string(bytes.TrimRight(img[24:32], "\x00")) == kwlo(m).id
}

func kwlText(b []byte) string {
	if k := bytes.IndexByte(b, 0); k >= 0 {
		b = b[:k]
	}
	return string(b)
}

// kwlFields: campi del canale; nil se il blocco è vuoto o non si capisce
func kwlFields(m *genModel, r []byte) []string {
	if r[0] != 1 {
		return nil
	}
	ix := kwlIdxs[kwlo(m).kind]
	f := strings.Split(kwlText(r[kwlSpecAt:kwlSpecAt+kwlSpecLen]), ",")
	if len(f) < ix.n {
		return nil
	}
	if _, err := strconv.ParseInt(f[ix.freq], 10, 64); err != nil {
		return nil
	}
	return f
}

func kwlInt(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

func kwlSplitTx(r []byte) int64 {
	if r[1] != 1 {
		return 0
	}
	return kwlInt(kwlText(r[kwlSplitAt : kwlSplitAt+11]))
}

func (f kwlFam) Get(m *genModel, img []byte, i int) *Chan {
	r, o := f.rec(img, i), kwlo(m)
	ix := kwlIdxs[o.kind]
	fl := kwlFields(m, r)
	if fl == nil {
		return nil
	}
	c := &Chan{Rx: kwlInt(fl[ix.freq])}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx = c.Rx
	offs := kwlInt(fl[ix.offs])
	switch kwlInt(fl[ix.shift]) {
	case 1:
		c.Tx = c.Rx + offs
	case 2:
		if c.Rx-offs > 0 {
			c.Tx = c.Rx - offs
		}
	default:
		if tx := kwlSplitTx(r); tx > 0 {
			c.Tx = tx
		}
	}
	tone := func(k int) string {
		n := int(kwlInt(fl[k])) - ix.codeB
		if n < 0 || n >= len(o.tones) {
			return "88.5"
		}
		return fmtTone(o.tones[n])
	}
	switch {
	case ix.dcs >= 0 && kwlInt(fl[ix.dcs]) == 1:
		code := 23
		if n := int(kwlInt(fl[ix.dcode])); n >= 0 && n < len(dtcsCodes) {
			code = dtcsCodes[n]
		}
		c.TxTone = fmt.Sprintf("D%03dN", code)
		c.RxTone = c.TxTone
	case kwlInt(fl[ix.ctcss]) != 0:
		c.TxTone = tone(ix.ccode)
		c.RxTone = c.TxTone
	case kwlInt(fl[ix.tone]) != 0:
		c.TxTone = tone(ix.tcode)
	}
	if o.kind == kwlK2 {
		c.Narrow = kwlInt(fl[ix.mode]) == 1
	}
	c.Scan = kwlInt(fl[ix.lock]) != 1
	var sb strings.Builder
	for _, b := range []byte(kwlText(r[kwlNameAt : kwlNameAt+kwlNameLen])) {
		if b < 0x20 || b > 0x7E {
			b = '?'
		}
		sb.WriteByte(b)
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

// toneOK: combinazioni di toni di queste radio (il convertitore toglie il tono in
// ricezione, poi anche l'altro, invece di fermarsi)
func (kwlFam) toneOK(t, r genTone) bool {
	switch {
	case t.Mode == "" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "":
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code && !t.Inv && !r.Inv:
	default:
		return false
	}
	return true
}

// fit: prova di scrittura del canale su una memoria vuota. I toni che il modello
// non ha (DCS su TH-G71 e TH-D7, subtoni fuori tabella) si tolgono; un canale che
// la radio non può memorizzare non viene copiato
func (f kwlFam) fit(m *genModel, c *Chan) (bool, string) {
	n := *c
	err := f.Set(m, kwlBlank(m), 0, &n)
	if err == nil {
		return true, ""
	}
	if c.TxTone != "" || c.RxTone != "" {
		n.TxTone, n.RxTone = "", ""
		if f.Set(m, kwlBlank(m), 0, &n) == nil {
			why := fmt.Sprintf("tono «%s» non disponibile su questa radio, impostato OFF", c.TxTone)
			c.TxTone, c.RxTone = "", ""
			return true, why
		}
	}
	return false, err.Error()
}

// kwlStepOK: la frequenza è un multiplo del passo
func kwlStepOK(hz int64, step int) bool {
	if step == 8333 {
		return hz*3%25000 == 0
	}
	return step > 0 && hz%int64(step) == 0
}

// chirpInvalid: combinazioni che la radio non ha (i test non le provano)
func (f kwlFam) chirpInvalid(m *genModel, c *Chan) bool {
	_, _, _, _, _, _, err := f.tones(m, c, nil)
	return err != nil
}

// tones: (tono, CTCSS, DCS, codice del tono, codice del CTCSS, codice DCS); i
// codici non usati restano quelli del canale (88,5 Hz e 023 se è nuovo)
func (f kwlFam) tones(m *genModel, c *Chan, old []string) (tone, ctcss, dcs bool, tcode, ccode, dcode int, err error) {
	o := kwlo(m)
	ix := kwlIdxs[o.kind]
	find := func(hz10 int) (int, error) {
		for k, x := range o.tones {
			if x == hz10 && !(ix.codeB == 1 && hz10 == 693) {
				return k + ix.codeB, nil
			}
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	tcode, _ = find(885)
	ccode = tcode
	if old != nil {
		tcode, ccode = int(kwlInt(old[ix.tcode])), int(kwlInt(old[ix.ccode]))
		if ix.dcode >= 0 {
			dcode = int(kwlInt(old[ix.dcode]))
		}
	}
	t, e := genToneParse(c.TxTone)
	if e != nil {
		return false, false, false, 0, 0, 0, e
	}
	rt, e := genToneParse(c.RxTone)
	if e != nil {
		return false, false, false, 0, 0, 0, e
	}
	bad := fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
	if ix.dcs < 0 {
		bad = fmt.Errorf("la radio non ha questa combinazione di toni (mai il solo tono in ricezione o codici invertiti)")
	}
	switch {
	case t.Mode == "" && rt.Mode == "":
	case t.Mode == "Tone" && rt.Mode == "":
		tone = true
		if tcode, err = find(t.Hz10); err != nil {
			return
		}
	case t.Mode == "Tone" && rt.Mode == "Tone" && t.Hz10 == rt.Hz10:
		ctcss = true
		if ccode, err = find(t.Hz10); err != nil {
			return
		}
	case ix.dcs >= 0 && t.Mode == "DTCS" && rt.Mode == "DTCS" && t.Code == rt.Code && !t.Inv && !rt.Inv:
		dcs = true
		dcode = -1
		for k, x := range dtcsCodes {
			if x == t.Code {
				dcode = k
			}
		}
		if dcode < 0 {
			err = fmt.Errorf("tono non disponibile su questa radio")
			return
		}
	default:
		err = bad
	}
	return
}

func (f kwlFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, o := f.rec(img, i), kwlo(m)
	ix := kwlIdxs[o.kind]
	if c == nil {
		copy(r, make([]byte, kwlRec))
		return nil
	}
	old := kwlFields(m, r)
	var oldCh *Chan
	if old != nil {
		oldCh = f.Get(m, img, i)
	}
	tone, ctcss, dcs, tcode, ccode, dcode, err := f.tones(m, c, old)
	if err != nil {
		return err
	}
	if c.Rx <= 0 || c.Rx >= 10000000000 {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	// campi di partenza: quelli del canale, o un canale nuovo
	fl := make([]string, ix.n)
	for k := range fl {
		fl[k] = "0"
	}
	switch o.kind {
	case kwlG71, kwlD7:
		fl[9], fl[11] = "", ""
	case kwlK2:
		fl[14] = "0000000000"
	}
	if old != nil {
		copy(fl, old[:ix.n])
	}
	// set: il campo cambia solo se cambia il valore (il testo della radio resta com'è)
	set := func(k int, v int64, w int, base int) {
		if k < 0 {
			return
		}
		if cur, e := strconv.ParseInt(strings.TrimSpace(fl[k]), base, 64); e == nil && cur == v && old != nil {
			return
		}
		if base == 16 {
			fl[k] = strings.ToUpper(strconv.FormatInt(v, 16))
			return
		}
		fl[k] = fmt.Sprintf("%0*d", w, v)
	}
	if o.kind != kwlK2 {
		fl[0] = "0"
		if o.kind != kwlF6 {
			fl[1] = "0"
		}
	}
	set(ix.num, int64(i), ix.numW, 10)
	set(ix.freq, c.Rx, ix.freqW, 10)
	// passo: resta quello del canale se la frequenza è un suo multiplo
	if len(o.steps) > 0 {
		cur, e := strconv.ParseInt(fl[ix.step], 16, 64)
		if e != nil || old == nil || cur < 0 || int(cur) >= len(o.steps) || !kwlStepOK(c.Rx, o.steps[cur]) {
			k := -1
			for n, s := range o.steps {
				if kwlStepOK(c.Rx, s) {
					k = n
					break
				}
			}
			if k < 0 {
				return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
			}
			set(ix.step, int64(k), 1, 16)
		}
	} else if c.Rx%5000 != 0 && c.Rx%6250 != 0 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// ponte o frequenza di trasmissione separata
	d := c.Tx - c.Rx
	ad := d
	if ad < 0 {
		ad = -ad
	}
	split := int64(0)
	switch {
	case d == 0:
		set(ix.shift, 0, 1, 10)
		set(ix.offs, kwlInt(fl[ix.offs]), ix.offsW, 10) // lo scostamento impostato resta
	case ad <= o.maxOffs && ad%50000 == 0:
		if d > 0 {
			set(ix.shift, 1, 1, 10)
		} else {
			set(ix.shift, 2, 1, 10)
		}
		set(ix.offs, ad, ix.offsW, 10)
	case o.split:
		set(ix.shift, 0, 1, 10)
		set(ix.offs, kwlInt(fl[ix.offs]), ix.offsW, 10)
		split = c.Tx
	default:
		return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
	}
	b2i := func(b bool) int64 {
		if b {
			return 1
		}
		return 0
	}
	set(ix.tone, b2i(tone), 1, 10)
	set(ix.ctcss, b2i(ctcss), 1, 10)
	set(ix.dcs, b2i(dcs), 1, 10)
	set(ix.tcode, int64(tcode), ix.codeW, 10)
	set(ix.ccode, int64(ccode), ix.codeW, 10)
	set(ix.dcode, int64(dcode), 3, 10)
	// modo: resta quello del canale se la frequenza non cambia
	switch o.kind {
	case kwlF6:
		if oldCh == nil || oldCh.Rx != c.Rx {
			set(ix.mode, icx90Mode(c.Rx), 1, 10) // 0 FM, 1 FM larga, 2 AM
		}
	case kwlD7:
		if oldCh == nil || oldCh.Rx != c.Rx {
			set(ix.mode, b2i(icx90Mode(c.Rx) == 2), 1, 10) // 0 FM, 1 AM
		}
	case kwlK2:
		set(ix.mode, b2i(c.Narrow), 1, 10)
	}
	set(ix.lock, b2i(!c.Scan), 1, 10)
	spec := strings.Join(fl, ",")
	if len(spec) > kwlSpecLen {
		return fmt.Errorf("errore interno: il canale %d non si rilegge uguale", i+m.First)
	}
	name := c.Name
	if len(name) > m.NameMax {
		name = name[:m.NameMax]
	}
	var oldName []byte
	if oldCh != nil && oldCh.Name == name {
		oldName = append([]byte{}, r[kwlNameAt:kwlNameAt+kwlNameLen]...) // anche i caratteri non ASCII restano
	}
	copy(r, make([]byte, kwlRec))
	r[0] = 1
	copy(r[kwlSpecAt:], spec)
	if oldName != nil {
		copy(r[kwlNameAt:], oldName)
	} else {
		copy(r[kwlNameAt:kwlNameAt+kwlNameLen], name)
	}
	if split > 0 {
		r[1] = 1
		copy(r[kwlSplitAt:], fmt.Sprintf("%011d", split))
	}
	return nil
}

// ---------------------------------------------------------------- collegamento

// comandi del canale n: lettura, nome, frequenza di trasmissione separata
func kwlCmds(o *kwlOpt, n int) (get, name, split string) {
	switch o.kind {
	case kwlF6:
		return fmt.Sprintf("MR 0,%03d", n), fmt.Sprintf("MNA %03d", n), fmt.Sprintf("MR 1,%03d", n)
	case kwlK2:
		return fmt.Sprintf("ME %02d", n), fmt.Sprintf("MN %02d", n), ""
	}
	return fmt.Sprintf("MR 0,0,%03d", n), fmt.Sprintf("MNA 0,%03d", n), fmt.Sprintf("MR 0,1,%03d", n)
}

func kwlRefused(r string) bool { return r == "N" || r == "?" || r == "E" }

type kwlSess struct {
	*rawSess
	o *kwlOpt
}

func (s *kwlSess) line(cmd string) (string, error) {
	r, err := thd74Line(s.rawSess, cmd, 2*time.Second)
	if err != nil {
		return "", err
	}
	if r == "" {
		return "", errTimeout
	}
	return r, nil
}

func kwlOpen(ctx context.Context, m *genModel, port string) (*kwlSess, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	s := &kwlSess{rs, kwlo(m)}
	fail := func(e error) (*kwlSess, error) {
		rs.p.Close()
		return nil, e
	}
	id := ""
	for _, baud := range s.o.bauds {
		if err := rs.p.SetBaud(baud); err != nil {
			return fail(err)
		}
		if err := rs.send([]byte("\r")); err != nil {
			return fail(err)
		}
		readUpTo(ctx, rs.p, 32, 250*time.Millisecond)
		if ctx.Err() != nil {
			return fail(errCancelled)
		}
		for try := 0; try < 2 && id == ""; try++ {
			r, err := thd74Line(rs, "ID", time.Second)
			if err != nil {
				return fail(err)
			}
			if strings.HasPrefix(r, "ID ") {
				id = strings.TrimSpace(r[3:])
			} else if r != "?" {
				break
			}
		}
		if id != "" {
			break
		}
	}
	if id == "" {
		return fail(genFail("Collegamento", errTimeout))
	}
	if id != s.o.id {
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwlOpt); ok && xo.id == id {
				return fail(rErr("la radio è un altro modello: apri la scheda «%s»", x.Name))
			}
		}
		return fail(rErr("la radio è un altro modello (%s)", id))
	}
	if _, err := s.line("AI 0"); err != nil {
		return fail(genFail("Collegamento", err))
	}
	return s, nil
}

// readAll: tutti i canali della radio nel formato dell'immagine
func (s *kwlSess) readAll(m *genModel, prog Progress, what string) ([]byte, error) {
	img := kwlBlank(m)
	f := kwlFam{}
	ix := kwlIdxs[s.o.kind]
	for i := 0; i < m.Slots; i++ {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		prog(i, m.Slots, what)
		get, name, split := kwlCmds(s.o, i)
		r, err := s.line(get)
		if err != nil {
			return nil, genFail("Lettura", err)
		}
		if kwlRefused(r) {
			continue
		}
		k := strings.IndexByte(r, ' ')
		if k < 0 || r[:k] != get[:2] || len(r)-k-1 > kwlSpecLen {
			return nil, rErr("risposta inattesa della radio")
		}
		// una risposta con meno campi di quelli previsti (i driver di CHIRP la
		// ammettono) si completa con «0»: in scrittura i campi si mandano tutti
		spec := r[k+1:]
		for n := strings.Count(spec, ",") + 1; n < ix.n && n > ix.freq; n++ {
			spec += ",0"
		}
		if len(spec) > kwlSpecLen {
			return nil, rErr("risposta inattesa della radio")
		}
		rec := f.rec(img, i)
		rec[0] = 1
		copy(rec[kwlSpecAt:], spec)
		fl := kwlFields(m, rec)
		if fl == nil {
			return nil, rErr("risposta inattesa della radio")
		}
		if m.NameMax > 0 {
			nr, err := s.line(name)
			if err != nil {
				return nil, genFail("Lettura", err)
			}
			if k := strings.IndexByte(nr, ' '); k >= 0 && !kwlRefused(nr) {
				// il nome è quello che segue il numero del canale (alcune radio
				// rispondono «MNA 0,nnn,nome» anche alla richiesta «MNA nnn»)
				parts := strings.Split(nr[k+1:], ",")
				num := name[strings.LastIndexAny(name, " ,")+1:]
				for j, p := range parts {
					if p == num && j+1 < len(parts) {
						nm := strings.Join(parts[j+1:], ",")
						if len(nm) > kwlNameLen {
							nm = nm[:kwlNameLen]
						}
						copy(rec[kwlNameAt:], nm)
						break
					}
				}
			}
		}
		if s.o.split && kwlInt(fl[ix.shift]) == 0 {
			sr, err := s.line(split)
			if err != nil {
				return nil, genFail("Lettura", err)
			}
			if k := strings.IndexByte(sr, ' '); k >= 0 && !kwlRefused(sr) {
				parts := strings.Split(sr[k+1:], ",")
				if at := strings.Count(split, ",") + 1; len(parts) > at {
					if tx := kwlInt(parts[at]); tx > 0 && tx != kwlInt(fl[ix.freq]) {
						rec[1] = 1
						copy(rec[kwlSplitAt:], fmt.Sprintf("%011d", tx))
					}
				}
			}
		}
	}
	return img, nil
}

func (f kwlFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := kwlOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	img, err := s.readAll(m, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	prog(m.Slots, m.Slots, "Lettura completata")
	return img, "", nil
}

func (f kwlFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := kwlOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	cur, err := s.readAll(m, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		return "", err
	}
	var todo []int
	for i := 0; i < m.Slots; i++ {
		if !bytes.Equal(f.rec(cur, i), f.rec(img, i)) {
			todo = append(todo, i)
		}
	}
	send := func(cmd string) error {
		r, err := s.line(cmd)
		if err != nil {
			return genFail("Scrittura", err)
		}
		if kwlRefused(r) {
			return rErr("la radio ha rifiutato la scrittura (%q)", cmd)
		}
		return nil
	}
	ix := kwlIdxs[s.o.kind]
	for k, i := range todo {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(k, len(todo), "Scrittura…")
		get, name, split := kwlCmds(s.o, i)
		erase := "MW" + get[2:]
		if s.o.kind == kwlK2 {
			erase = get + ","
		}
		was, now := f.rec(cur, i), f.rec(img, i)
		if now[0] != 1 {
			if err := send(erase); err != nil {
				return "", err
			}
			continue
		}
		fl := kwlFields(m, now)
		if fl == nil {
			return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
		}
		if was[1] == 1 && now[1] != 1 {
			// la frequenza di trasmissione separata si toglie cancellando il canale
			if err := send(erase); err != nil {
				return "", err
			}
		}
		write := "MW "
		if s.o.kind == kwlK2 {
			write = "ME "
		}
		if err := send(write + kwlText(now[kwlSpecAt:kwlSpecAt+kwlSpecLen])); err != nil {
			return "", err
		}
		if m.NameMax > 0 {
			time.Sleep(kwlPause)
			r, err := s.line(name + "," + kwlText(now[kwlNameAt:kwlNameAt+kwlNameLen]))
			if err != nil {
				return "", genFail("Scrittura", err)
			}
			// TH-K2 a 100 canali: i nomi non ci sono e la radio li rifiuta
			if kwlRefused(r) && s.o.kind != kwlK2 {
				return "", rErr("la radio ha rifiutato la scrittura (%q)", name)
			}
		}
		if now[1] == 1 {
			step := "0"
			if s.o.kind == kwlG71 {
				step = fl[ix.step]
			}
			if err := send("MW" + split[2:] + "," + kwlText(now[kwlSplitAt:kwlSplitAt+11]) + "," + step); err != nil {
				return "", err
			}
		}
	}
	prog(len(todo), len(todo), "Scrittura completata")
	return "", nil
}
