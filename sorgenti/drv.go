// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: formato dei file .img (chirp_common.py, directory.py).

package main

// Driver delle radio "generiche" (tutti i modelli Baofeng e Bajeton oltre al
// K67 Pro e alla Bajeton 8300, che hanno le loro schede dedicate).
//
// Ogni famiglia di radio ha:
//   - il protocollo via cavo (lettura e scrittura della memoria della radio);
//   - la mappa della memoria e la codifica dei canali.
// Protocolli, mappe e codifiche sono quelli documentati dal progetto CHIRP
// (https://chirpmyradio.com, driver in Python, licenza GPL), riscritti qui in
// Go e verificati con i test automatici confrontando, sulle stesse immagini di
// memoria, i canali letti e i byte scritti da CHIRP e da questo programma, e i
// byte scambiati con radio simulate (vedi verifica/chirp).
//
// La memoria di una radio si salva in file .img nello stesso formato di CHIRP
// (memoria della radio + metadati): i file si aprono con CHIRP e viceversa.

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// livello di potenza della radio
type genPower struct {
	Label string  `json:"label"` // nome mostrato (es. "Alta 5 W")
	Watts float64 `json:"watts"`
	Sem   int     `json:"sem"` // significato nel programma: 0 alta, 1 media, 2 bassa
}

type genModel struct {
	ID      string // identificativo interno (es. "uv5r")
	Vendor  string // marca e modello come in CHIRP (per i file .img)
	Model   string
	Variant string
	RClass  string // classe di CHIRP (metadati dei file .img)
	Name    string // nome mostrato
	Slots   int    // numero di canali
	First   int    // numero del primo canale come lo mostra la radio (0 o 1)
	NameMax int
	Charset string // caratteri ammessi nei nomi ("" = ASCII stampabile)
	Powers  []genPower
	Bands   []band // gamme di frequenza accettate
	Narrow  bool   // banda stretta disponibile
	HasScan bool
	NoTxOff bool  // la radio non permette di disabilitare la trasmissione sul canale
	// NoChirp: modello non presente in CHIRP (protocollo ricavato dal programma del
	// costruttore); RClass è la classe di CHIRP con la stessa mappa della memoria
	NoChirp bool
	Zones   int   // canali divisi in zone (banchi) di Slots/Zones posizioni (solo visualizzazione)
	DCS     []int // codici DCS ammessi
	Baud    int
	Fam     genFamily
	Opt     interface{} // parametri della famiglia
	Notes   string      // avvertenze mostrate nella finestra del cavo
}

// genFamily: operazioni di una famiglia di radio.
type genFamily interface {
	// canale i (0 = primo); nil se la posizione è vuota
	Get(m *genModel, img []byte, i int) *Chan
	// scrive il canale (nil = cancella), come CHIRP set_memory
	Set(m *genModel, img []byte, i int, c *Chan) error
	// byte del canale (record ed eventuale nome) per conservarli negli spostamenti
	Rec(m *genModel, img []byte, i int) []byte
	PutRec(m *genModel, img []byte, i int, raw []byte)
	// lettura e scrittura via cavo
	Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error)
	Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error)
	// il contenuto (senza metadati) è un'immagine di questo modello?
	Match(m *genModel, img []byte) bool
}

var genModels []*genModel
var genByID = map[string]*genModel{}

func genRegister(ms ...*genModel) {
	for _, m := range ms {
		if _, dup := genByID[m.ID]; dup {
			panic("modello registrato due volte: " + m.ID)
		}
		if m.First == 0 && m.Slots == 0 {
			panic("modello senza canali: " + m.ID)
		}
		genModels = append(genModels, m)
		genByID[m.ID] = m
	}
}

func isGenKind(k string) bool { _, ok := genByID[k]; return ok }

// ---------------------------------------------------------------- potenza

func (m *genModel) powerIndex(sem int) int {
	best, bd := 0, 99
	for i, p := range m.Powers {
		d := p.Sem - sem
		if d < 0 {
			d = -d
		}
		if d < bd {
			best, bd = i, d
		}
	}
	return best
}

func (m *genModel) powerSem(idx int) int {
	if idx >= 0 && idx < len(m.Powers) {
		return m.Powers[idx].Sem
	}
	return 0
}

// ---------------------------------------------------------------- dati grezzi del canale

func genRaw(m *genModel, rec []byte) string { return "g:" + m.Fam2() + ":" + hex.EncodeToString(rec) }

// Fam2: identificativo della famiglia (i dati grezzi valgono tra modelli della stessa mappa)
func (m *genModel) Fam2() string {
	return fmt.Sprintf("%T", m.Fam) + "/" + m.ID
}

func genUnraw(m *genModel, s string) ([]byte, bool) {
	p := "g:" + m.Fam2() + ":"
	if !strings.HasPrefix(s, p) {
		return nil, false
	}
	b, err := hex.DecodeString(s[len(p):])
	return b, err == nil
}

// genGet: canale con i dati grezzi.
func genGet(m *genModel, img []byte, i int) *Chan {
	c := m.Fam.Get(m, img, i)
	if c != nil {
		c.Raw = genRaw(m, m.Fam.Rec(m, img, i))
	}
	return c
}

func genChans(m *genModel, img []byte) []*Chan {
	out := make([]*Chan, m.Slots)
	for i := range out {
		out[i] = genGet(m, img, i)
	}
	return out
}

func chanEqual(a, b *Chan) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return *a == *b
}

// genWrite applica i canali all'immagine base: modifica solo le posizioni
// cambiate (un file aperto e salvato senza modifiche resta identico).
func genWrite(m *genModel, base []byte, chs []*Chan, rep *Report) ([]byte, error) {
	out := append([]byte{}, base...)
	for i := 0; i < m.Slots; i++ {
		var want *Chan
		if i < len(chs) {
			want = chs[i]
		}
		cur := genGet(m, out, i)
		if chanEqual(cur, want) {
			continue
		}
		if lk, ok := m.Fam.(interface {
			Locked(m *genModel, i int, old, want *Chan) error
		}); ok {
			if err := lk.Locked(m, i, cur, want); err != nil {
				return nil, fmt.Errorf("canale %d: %v", i+m.First, err)
			}
		}
		if want != nil {
			if err := genValidate(m, want, i); err != nil {
				return nil, err
			}
			if raw, ok := genUnraw(m, want.Raw); ok && len(raw) == len(m.Fam.Rec(m, out, i)) {
				m.Fam.PutRec(m, out, i, raw)
			}
		}
		if err := m.Fam.Set(m, out, i, want); err != nil {
			return nil, fmt.Errorf("canale %d: %v", i+m.First, err)
		}
		if want != nil {
			if back := m.Fam.Get(m, out, i); back == nil || back.Rx != want.Rx || (!want.TxOff && back.Tx != want.Tx) {
				return nil, fmt.Errorf("errore interno: il canale %d non si rilegge uguale", i+m.First)
			}
		}
	}
	return out, nil
}

func genValidate(m *genModel, c *Chan, i int) error {
	if c.Rx <= 0 || c.Rx >= 1e9*2 {
		return fmt.Errorf("canale %d: frequenza RX non valida", i+m.First)
	}
	if !c.TxOff && (c.Tx <= 0 || c.Tx >= 1e9*2) {
		return fmt.Errorf("canale %d: frequenza TX non valida", i+m.First)
	}
	if c.TxOff && m.NoTxOff {
		return fmt.Errorf("canale %d: questa radio non permette di disabilitare la trasmissione sul canale", i+m.First)
	}
	for _, t := range []string{c.TxTone, c.RxTone} {
		if t == "" {
			continue
		}
		if _, ok := parseTone(t); ok {
			continue
		}
		code, _, ok := dcsParse(t)
		if !ok {
			return fmt.Errorf("canale %d: tono %q non valido", i+m.First, t)
		}
		if !m.dcsOK(code) {
			return fmt.Errorf("canale %d: codice DCS %s non disponibile su questa radio", i+m.First, code)
		}
	}
	return nil
}

func (m *genModel) dcsOK(code string) bool {
	var n int
	if _, err := fmt.Sscanf(code, "%d", &n); err != nil {
		return false
	}
	for _, d := range m.DCS {
		if d == n {
			return true
		}
	}
	return false
}

func (m *genModel) inBands(hz int64) bool {
	for _, b := range m.Bands {
		if hz >= b.Lo && hz <= b.Hi {
			return true
		}
	}
	return len(m.Bands) == 0
}

// ---------------------------------------------------------------- toni

// toni nel formato del programma: "" spento, "88.5" CTCSS, "D023N"/"D023I" DCS
type genTone struct {
	Mode string // "", "Tone", "DTCS"
	Hz10 int    // decimi di Hz
	Code int    // codice DCS (es. 23)
	Inv  bool
}

func genToneParse(s string) (genTone, error) {
	if strings.TrimSpace(s) == "" {
		return genTone{}, nil
	}
	if v, ok := parseTone(s); ok {
		if v == 0 {
			return genTone{}, nil
		}
		return genTone{Mode: "Tone", Hz10: v}, nil
	}
	code, inv, ok := dcsParse(s)
	if !ok {
		return genTone{}, fmt.Errorf("tono %q non valido", s)
	}
	var n int
	fmt.Sscanf(code, "%d", &n)
	return genTone{Mode: "DTCS", Code: n, Inv: inv}, nil
}

func (t genTone) String() string {
	switch t.Mode {
	case "Tone":
		return fmtTone(t.Hz10)
	case "DTCS":
		s := fmt.Sprintf("D%03d", t.Code)
		if t.Inv {
			return s + "I"
		}
		return s + "N"
	}
	return ""
}

// ---------------------------------------------------------------- nomi

// genNameFromBytes: come CHIRP, i caratteri 0xFF e 0x00 diventano spazi
// (oppure terminano il nome, a seconda della radio) e gli spazi finali spariscono.
func genNameFromBytes(b []byte, stopAt0 bool) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if stopAt0 && c == 0 {
			break
		}
		if c == 0xFF || c == 0 {
			c = ' '
		}
		out = append(out, c)
	}
	return strings.TrimRight(string(out), " ")
}

// pyName: nome come lo calcola CHIRP nella maggior parte dei driver: i byte
// 0xFF diventano spazi (se ff2sp), gli altri restano, e si tolgono gli spazi
// finali (str.rstrip di Python).
func pyName(b []byte, ff2sp bool) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if ff2sp && c == 0xFF {
			c = ' '
		}
		out = append(out, c)
	}
	return strings.TrimRight(string(out), " \t\n\r\x0b\x0c")
}

// filtro dei nomi per la radio (maiuscole e caratteri ammessi)
func (m *genModel) cleanName(s string) string {
	s = asciiName(s)
	if m.Charset != "" {
		up := true
		for _, c := range m.Charset {
			if c >= 'a' && c <= 'z' {
				up = false
			}
		}
		if up {
			s = strings.ToUpper(s)
		}
		b := []byte(s)
		for i, c := range b {
			if !strings.ContainsRune(m.Charset, rune(c)) {
				b[i] = ' '
			}
		}
		s = string(b)
	}
	if len(s) > m.NameMax {
		s = s[:m.NameMax]
	}
	return strings.TrimRight(s, " ")
}

// ---------------------------------------------------------------- file .img (formato CHIRP)

var chirpMagic = []byte("\x00\xffchirp\xeeimg\x00\x01")

type chirpMeta struct {
	RClass  string `json:"rclass"`
	Vendor  string `json:"vendor"`
	Model   string `json:"model"`
	Variant string `json:"variant"`
	Version string `json:"chirp_version,omitempty"`
}

// splitImg separa memoria e metadati di un file .img.
func splitImg(b []byte) ([]byte, map[string]interface{}) {
	i := bytes.Index(b, chirpMagic)
	if i < 0 {
		return b, nil
	}
	meta := map[string]interface{}{}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(b[i+len(chirpMagic):])))
	if err == nil {
		_ = json.Unmarshal(raw, &meta)
	}
	return b[:i], meta
}

// makeImg: file .img con i metadati di CHIRP (conservando quelli già presenti).
func makeImg(m *genModel, mem []byte, old map[string]interface{}) []byte {
	meta := map[string]interface{}{}
	for k, v := range old {
		meta[k] = v
	}
	meta["rclass"] = m.RClass
	meta["vendor"] = m.Vendor
	meta["model"] = m.Model
	meta["variant"] = m.Variant
	if _, ok := meta["chirp_version"]; !ok {
		meta["chirp_version"] = "IU0VYT-Flash-" + appVersion
	}
	// chiavi in ordine stabile
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		kb, _ := json.Marshal(k)
		vb, _ := json.Marshal(meta[k])
		buf.Write(kb)
		buf.WriteString(": ")
		buf.Write(vb)
	}
	buf.WriteByte('}')
	out := append([]byte{}, mem...)
	out = append(out, chirpMagic...)
	out = append(out, []byte(base64.StdEncoding.EncodeToString(buf.Bytes()))...)
	return out
}

// detectGen riconosce il modello di un file .img (metadati di CHIRP o contenuto).
func detectGen(b []byte, prefer string) (*genModel, []byte, map[string]interface{}, error) {
	mem, meta := splitImg(b)
	if meta != nil {
		rc, _ := meta["rclass"].(string)
		model, _ := meta["model"].(string)
		vendor, _ := meta["vendor"].(string)
		var cands []*genModel
		for _, m := range genModels {
			if m.RClass == rc {
				cands = append(cands, m)
			}
		}
		if len(cands) == 0 {
			for _, m := range genModels {
				if strings.EqualFold(m.Model, model) && strings.EqualFold(m.Vendor, vendor) {
					cands = append(cands, m)
				}
			}
		}
		for _, m := range cands {
			if m.ID == prefer {
				return m, mem, meta, nil
			}
		}
		for _, m := range cands {
			if m.Model == model {
				return m, mem, meta, nil
			}
		}
		if len(cands) > 0 {
			return cands[0], mem, meta, nil
		}
		if rc != "" || model != "" {
			return nil, nil, nil, fmt.Errorf("il file è di una radio non gestita da questo programma (%s %s, classe %s di CHIRP)", vendor, model, rc)
		}
	}
	if m, ok := genByID[prefer]; ok && m.Fam.Match(m, mem) {
		return m, mem, meta, nil
	}
	for _, m := range genModels {
		if m.Fam.Match(m, mem) {
			return m, mem, meta, nil
		}
	}
	return nil, nil, nil, errors.New("formato non riconosciuto: il file non è un'immagine .img di una radio gestita")
}

// ---------------------------------------------------------------- definizione per l'interfaccia

type genDef struct {
	Kind    string     `json:"kind"`
	Name    string     `json:"name"`
	Ext     string     `json:"ext"`
	Slots   int        `json:"slots"`
	First   int        `json:"first"`
	NameMax int        `json:"nameMax"`
	Powers  []genPower `json:"powerLevels"`
	Bands   []band     `json:"bands"`
	Narrow  bool       `json:"narrow"`
	HasScan bool       `json:"hasScan"`
	DCS     []string   `json:"dcs"`
	Charset string     `json:"charset"`
	Notes   string     `json:"notes"`
	NoTxOff bool       `json:"noTxOff"`
	Zones   int        `json:"zones"`
	PerZone int        `json:"perZone"`
	Vendor  string     `json:"vendor"`
	Model   string     `json:"model"`
}

func (m *genModel) def() genDef {
	d := genDef{Kind: m.ID, Name: m.Name, Ext: ".img", Slots: m.Slots, First: m.First, NameMax: m.NameMax,
		Powers: m.Powers, Bands: m.Bands, Narrow: m.Narrow, HasScan: m.HasScan, Charset: m.Charset, Notes: m.Notes,
		NoTxOff: m.NoTxOff, Zones: m.Zones, Vendor: m.Vendor, Model: m.Model}
	if m.Zones > 0 {
		d.PerZone = m.Slots / m.Zones
	}
	for _, c := range m.DCS {
		d.DCS = append(d.DCS, fmt.Sprintf("D%03dN", c), fmt.Sprintf("D%03dI", c))
	}
	return d
}

// codici DCS standard (come chirp_common.DTCS_CODES)
var dtcsCodes = []int{23, 25, 26, 31, 32, 36, 43, 47, 51, 53, 54, 65, 71, 72, 73, 74, 114, 115, 116, 122, 125, 131,
	132, 134, 143, 145, 152, 155, 156, 162, 165, 172, 174, 205, 212, 223, 225, 226, 243, 244, 245, 246, 251, 252,
	255, 261, 263, 265, 266, 271, 274, 306, 311, 315, 325, 331, 332, 343, 346, 351, 356, 364, 365, 371, 411, 412,
	413, 423, 431, 432, 445, 446, 452, 454, 455, 462, 464, 465, 466, 503, 506, 516, 523, 526, 532, 546, 565, 606,
	612, 624, 627, 631, 632, 654, 662, 664, 703, 712, 723, 731, 732, 734, 743, 754}

func withCodes(base []int, extra ...int) []int {
	out := append(append([]int{}, base...), extra...)
	sort.Ints(out)
	return out
}

func idxInt(l []int, v int) int {
	for i, x := range l {
		if x == v {
			return i
		}
	}
	return -1
}

// errore della radio (come errors.RadioError di CHIRP)
type radioErr struct{ msg string }

func (e radioErr) Error() string { return e.msg }

func rErr(f string, a ...interface{}) error { return radioErr{fmt.Sprintf(f, a...)} }
