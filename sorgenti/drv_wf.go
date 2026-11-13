// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radioddity UB15-G, DB-35G, DB35-UV e DB40-G (4.28). Il programma del
// costruttore è lo stesso per tutte («wfintegrationFramework») e
// prende velocità, aggancio, indirizzi e posizione di ogni campo da un file
// «model.xml» che sta accanto all'exe; il resto (ordine dei byte, pacchetti) è
// stato letto nel codice del programma (modelReadAsync, handShakeAsync,
// readChannelAsync, writeChannelAsync, readParamAsync, writeParamAsync).
//
//   tipo       55 01 00 00 → 4 byte con il tipo di radio (00 06 02 00 per
//              UB15-G, DB-35G e DB35-UV; 00 08 09 03 per la DB40-G)
//   aggancio   02 «PROGRAM» → 06;  02 → «SMP558» FF FF;  06 → 06
//   lettura    52 indirizzo (2 byte, prima l'alto) lunghezza → 4 byte + dati
//   scrittura  57 indirizzo (2 byte) lunghezza + dati → 06
//   fine       62
//
// L'indirizzo di un canale è indirizzo iniziale + numero della riga × 16, anche
// quando la riga è lunga 24 o 32 byte: è il numero del record, non una posizione
// in byte. Le impostazioni sono una sola riga all'indirizzo E0 00.
//
// Riga di un canale: ricezione e trasmissione in 4 byte (Hz, prima il byte
// alto; trasmissione 0 = solo ricezione; ricezione fuori banda = riga vuota),
// tono di ricezione e di trasmissione in 2 byte, prima l'alto (FF FF nessuno;
// CTCSS in decimi di Hz; DCS 0x2800 + codice, 0xA800 + codice se invertito),
// poi i bit di potenza, banda, scansione, blocco su canale occupato e priorità
// e, dove c'è, il nome di 8 caratteri maiuscoli riempito con spazi.
//
// La DB40-G dal 9/11/2026 è anche nel CHIRP (radioddity_db40g.py, provato dal
// suo autore su una radio vera): stessa memoria e stessi pacchetti. Per quel
// modello si fa byte per byte come il CHIRP, così i file sono gli stessi e le
// prove confrontano i due: niente domanda del tipo prima dell'aggancio, davanti
// ai dati letti si controlla l'eco «52 indirizzo lunghezza», nomi riempiti con
// 00 (come li ha la radio di fabbrica), riga cancellata tutta a 00, banda e
// potenza in un byte intero. Per le altre tre la risposta alla lettura non è
// nota (il programma del costruttore salta i primi 4 byte senza guardarli) e
// si fa come il programma del costruttore.
//
// Nell'immagine: le righe dei canali una dopo l'altra, poi la riga delle
// impostazioni.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type wfBit struct {
	off   int // byte della riga
	shift uint
	bits  uint
}

func (b *wfBit) get(r []byte) int { return int(r[b.off]>>b.shift) & (1<<b.bits - 1) }
func (b *wfBit) put(r []byte, v int) {
	m := byte(1<<b.bits-1) << b.shift
	r[b.off] = r[b.off]&^m | byte(v)<<b.shift&m
}

type wfOpt struct {
	page  int    // byte di una riga
	start int    // indirizzo della prima riga
	ver   []byte // tipo di radio atteso
	name  int    // posizione del nome (8 byte), -1 se non c'è
	power wfBit
	band  wfBit // 0 stretta, 1 larga
	scan  *wfBit
	chirp bool // DB40-G: byte per byte come il driver del CHIRP
	prog  string
	set   func() []wfSetDef
}

const wfParamAddr = 0xE000

// wfCharset: i nomi sono in maiuscolo (il programma del costruttore li converte)
const wfCharset = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`{|}~"

func (o *wfOpt) size(m *genModel) int  { return (m.Slots + 1) * o.page }
func (o *wfOpt) param(m *genModel) int { return m.Slots * o.page }

// tipi di radio che il programma riconosce dalla risposta alla prima domanda
var wfKinds = map[string]string{
	"00000002": "H2", "00000005": "H5", "00060000": "D1", "00060100": "C1",
	"00060200": "R1 (UB15-G, DB-35G, DB35-UV)", "00080903": "DB40-G",
}

type wfFam struct{}

func wfO(m *genModel) *wfOpt { return m.Opt.(*wfOpt) }

func init() {
	note := func(baud int, o *wfOpt, extra string) string {
		if o.name >= 0 {
			return fmt.Sprintf("%d baud. Protocollo e memoria ricavati dal programma del costruttore %s e dal suo file di configurazione, "+
				"e provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, toni, potenza, banda, scansione e nomi dei canali "+
				"e le impostazioni generali; blocco su canale occupato e priorità restano quelli del canale. %s", baud, o.prog, extra)
		}
		return fmt.Sprintf("%d baud. Protocollo e memoria ricavati dal programma del costruttore %s e dal suo file di configurazione, "+
			"e provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, toni, potenza, banda e scansione dei canali (non hanno nome) "+
			"e le impostazioni generali; blocco su canale occupato e priorità restano quelli del canale. %s", baud, o.prog, extra)
	}
	lowHigh := []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}}
	r1 := []byte{0x00, 0x06, 0x02, 0x00}
	bit := func(off int, shift, bits uint) wfBit { return wfBit{off, shift, bits} }
	ub := &wfOpt{page: 16, ver: r1, name: -1, power: bit(12, 0, 2), band: bit(12, 2, 1), scan: &wfBit{12, 4, 1},
		prog: "«UB-15G» (versione del 21/8/2026)", set: wfSetUB15}
	ub110 := &wfOpt{page: 24, ver: r1, name: 13, power: bit(12, 0, 2), band: bit(12, 2, 1), scan: &wfBit{12, 4, 1},
		prog: "«UB15-G» per il firmware V1.10", set: wfSetUB15v110}
	db35g := &wfOpt{page: 16, ver: r1, name: -1, power: bit(12, 0, 2), band: bit(12, 2, 1), scan: &wfBit{12, 4, 1},
		prog: "«DB-35G»", set: wfSetDB35G}
	db35uv := &wfOpt{page: 16, ver: r1, name: -1, power: bit(12, 0, 2), band: bit(12, 2, 1), scan: &wfBit{12, 4, 1},
		prog: "«DB35-UV» (V6 e V8)", set: wfSetDB35UV}
	db40 := &wfOpt{page: 32, start: 0x0010, ver: []byte{0x00, 0x08, 0x09, 0x03}, name: 14, power: bit(13, 0, 8), band: bit(12, 0, 8),
		chirp: true, prog: "«DB40-G»", set: wfSetDB40G}
	ubRows := "Le 88 posizioni sono, nell'ordine della radio: i canali 1-15, poi «15RP», 16, «16RP» e così via fino a 22 e «22RP» " +
		"(RP = lo stesso canale con la trasmissione verso il ponte), poi i canali 23-80. "
	uhf := []band{{400000000, 470000000}}
	vu := []band{{136000000, 174000000}, {400000000, 470000000}}
	mk := func(id, model, name string, slots int, bands []band, pw []genPower, baud int, o *wfOpt, extra string) *genModel {
		m := &genModel{ID: id, Vendor: "Radioddity", Model: model, RClass: "IU0VYT_Radioddity_" + strings.NewReplacer("-", "", " ", "_").Replace(id[3:]),
			Name: name, Slots: slots, First: 1, Powers: pw, Bands: bands, Narrow: true, HasScan: o.scan != nil,
			NoChirp: true, DCS: dtcsAll(), Baud: baud, Fam: wfFam{}, Opt: o}
		if o.name >= 0 {
			m.NameMax, m.Charset = 8, wfCharset
		}
		m.Notes = strings.TrimSpace(note(baud, o, extra))
		if o.chirp { // dal novembre 2026 è anche nel CHIRP: stesso file, e le prove confrontano i due
			m.RClass, m.NoChirp = "RadioddityDB40G", false
			m.Notes = fmt.Sprintf("%d baud. Stessa memoria e stessi pacchetti del driver del CHIRP (che ha questo modello dal novembre 2026), "+
				"confrontati anche con il programma del costruttore %s e con il suo file di configurazione. Qui si modificano frequenze, toni, potenza, "+
				"banda e nomi dei canali e le impostazioni generali.", baud, o.prog)
		}
		return m
	}
	genRegister(
		mk("rd-ub15g", "UB15-G", "Radioddity UB15-G (firmware prima della V1.10)", 88, uhf, lowHigh, 115200, ub,
			ubRows+"Con il firmware V1.10 la memoria è diversa (ci sono i nomi): usa l'altra scheda. "),
		mk("rd-ub15g-110", "UB15-G", "Radioddity UB15-G (firmware V1.10, con i nomi)", 88, uhf, lowHigh, 115200, ub110,
			ubRows+"Con un firmware precedente la memoria è diversa (senza nomi): usa l'altra scheda. "),
		mk("rd-db35g", "DB35-G", "Radioddity DB35-G (DB-35G)", 80, vu, lowHigh, 115200, db35g, ""),
		mk("rd-db35uv", "DB35-UV", "Radioddity DB35-UV", 80, vu, lowHigh, 115200, db35uv, ""),
		mk("rd-db40g", "DB40-G", "Radioddity DB40-G", 199, []band{{400000000, 480000000}},
			[]genPower{{"Bassa", 5, 2}, {"Media", 20, 1}, {"Alta", 40, 0}}, 19200, db40, ""),
	)
}

func (wfFam) rec(m *genModel, img []byte, i int) []byte {
	p := wfO(m).page
	return img[i*p : i*p+p]
}

func (f wfFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, img, i)...)
}

func (f wfFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(f.rec(m, img, i), raw)
}

func (wfFam) Match(m *genModel, img []byte) bool { return false }

func wfToneDec(b []byte) string {
	v := int(binary.BigEndian.Uint16(b))
	switch {
	case v == 0xFFFF || v == 0:
		return ""
	case v&0x2000 != 0: // DCS (il programma scrive 0x2800 + codice; come il CHIRP basta il bit 0x2000)
		n := v & 0x1FF // codice in ottale: tre cifre da 3 bit
		return genTone{Mode: "DTCS", Code: (n>>6)*100 + ((n>>3)&7)*10 + (n & 7), Inv: v&0x8000 != 0}.String()
	case v >= 600 && v <= 2600:
		return genTone{Mode: "Tone", Hz10: v}.String()
	}
	return ""
}

func wfToneEnc(dst []byte, s string) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch t.Mode {
	case "Tone":
		if t.Hz10 < 600 || t.Hz10 > 2600 {
			return fmt.Errorf("tono %s non previsto dalla radio", s)
		}
		binary.BigEndian.PutUint16(dst, uint16(t.Hz10))
	case "DTCS":
		d2, d1, d0 := t.Code/100, t.Code/10%10, t.Code%10
		if t.Code < 0 || d2 > 7 || d1 > 7 || d0 > 7 {
			return fmt.Errorf("codice DCS %s non previsto dalla radio", s)
		}
		v := 0x2800 | d2<<6 | d1<<3 | d0
		if t.Inv {
			v |= 0x8000
		}
		binary.BigEndian.PutUint16(dst, uint16(v))
	default:
		dst[0], dst[1] = 0xFF, 0xFF
	}
	return nil
}

func (f wfFam) Get(m *genModel, img []byte, i int) *Chan {
	o := wfO(m)
	r := f.rec(m, img, i)
	rx := int64(binary.BigEndian.Uint32(r[0:4]))
	if !m.inBands(rx) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if tx := int64(binary.BigEndian.Uint32(r[4:8])); m.inBands(tx) {
		c.Tx = tx
	} else {
		c.TxOff = true
	}
	c.RxTone, c.TxTone = wfToneDec(r[8:10]), wfToneDec(r[10:12])
	pw := o.power.get(r)
	if pw >= len(m.Powers) {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = o.band.get(r) == 0
	if o.scan != nil {
		c.Scan = o.scan.get(r) == 1
	}
	if o.name >= 0 {
		c.Name = string(bytes.TrimRight(r[o.name:o.name+8], "\x00\xFF "))
	}
	return c
}

func (f wfFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := wfO(m)
	r := f.rec(m, img, i)
	if c == nil { // riga vuota come la scrive il programma
		for k := range r {
			r[k] = 0
		}
		if o.chirp {
			return nil
		}
		r[8], r[9], r[10], r[11] = 0xFF, 0xFF, 0xFF, 0xFF
		if o.name >= 0 {
			copy(r[o.name:o.name+8], "        ")
		}
		return nil
	}
	if ok, why := f.fit(m, c); !ok {
		return fmt.Errorf("%s", why)
	}
	var rt, tt [2]byte
	if err := wfToneEnc(rt[:], c.RxTone); err != nil {
		return err
	}
	if err := wfToneEnc(tt[:], c.TxTone); err != nil {
		return err
	}
	name := strings.ToUpper(c.Name)
	if o.name >= 0 {
		if len(name) > 8 {
			return fmt.Errorf("nome «%s» troppo lungo (al massimo 8 caratteri)", c.Name)
		}
		for _, ch := range name {
			if ch < 0x20 || ch > 0x7E {
				return fmt.Errorf("nome «%s»: sono ammessi solo lettere, cifre e segni semplici", c.Name)
			}
		}
	}
	if !o.chirp && !m.inBands(int64(binary.BigEndian.Uint32(r[0:4]))) { // riga nuova: si parte da zero come il programma
		for k := range r {
			r[k] = 0
		}
	}
	binary.BigEndian.PutUint32(r[0:4], uint32(c.Rx))
	if c.TxOff {
		binary.BigEndian.PutUint32(r[4:8], 0)
	} else {
		binary.BigEndian.PutUint32(r[4:8], uint32(c.Tx))
	}
	copy(r[8:10], rt[:])
	copy(r[10:12], tt[:])
	o.power.put(r, m.powerIndex(c.Power))
	if c.Narrow {
		o.band.put(r, 0)
	} else {
		o.band.put(r, 1)
	}
	if o.scan != nil {
		v := 0
		if c.Scan {
			v = 1
		}
		o.scan.put(r, v)
	}
	if o.name >= 0 {
		pad := "        "
		if o.chirp {
			pad = "\x00\x00\x00\x00\x00\x00\x00\x00"
		}
		copy(r[o.name:o.name+8], pad)
		copy(r[o.name:o.name+8], name)
	}
	return nil
}

func (wfFam) fit(m *genModel, c *Chan) (bool, string) {
	if !m.inBands(c.Rx) {
		return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	if !c.TxOff && !m.inBands(c.Tx) {
		return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Tx))
	}
	return true, ""
}

// chirpInvalid: modifiche di prova che il modello non può memorizzare (fuori banda)
func (f wfFam) chirpInvalid(m *genModel, c *Chan) bool {
	ok, _ := f.fit(m, c)
	return !ok
}

// ---------------------------------------------------------------- protocollo

// wfWait: attesa di ogni risposta (2 secondi nel file di configurazione); le prove la accorciano
var wfWait = 2 * time.Second

type wfSess struct {
	ctx context.Context
	p   Port
	o   *wfOpt
}

func (s *wfSess) xfer(send []byte, n int) ([]byte, error) {
	if err := s.p.Write(send); err != nil {
		return nil, err
	}
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, wfWait); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *wfSess) expect(send, want []byte, what string) error {
	got, err := s.xfer(send, len(want))
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return rErr("%s: risposta inattesa della radio (% X invece di % X)", what, got, want)
	}
	return nil
}

func (s *wfSess) enter() error {
	_ = s.p.Purge()
	if !s.o.chirp {
		kind, err := s.xfer([]byte{0x55, 0x01, 0x00, 0x00}, 4)
		if err != nil {
			return err
		}
		if !bytes.Equal(kind, s.o.ver) {
			if k, ok := wfKinds[fmt.Sprintf("%x", kind)]; ok {
				return rErr("la radio dichiara un altro tipo (%s): non è quella di questa scheda", k)
			}
			return rErr("la radio dichiara un tipo sconosciuto (% X): non è quella di questa scheda", kind)
		}
	}
	if err := s.expect([]byte("\x02PROGRAM"), []byte{0x06}, "aggancio"); err != nil {
		return err
	}
	if err := s.expect([]byte{0x02}, []byte("SMP558\xFF\xFF"), "sigla"); err != nil {
		return err
	}
	return s.expect([]byte{0x06}, []byte{0x06}, "conferma")
}

func (s *wfSess) read(addr int) ([]byte, error) {
	cmd := []byte{0x52, byte(addr >> 8), byte(addr), byte(s.o.page)}
	b, err := s.xfer(cmd, s.o.page+4)
	if err != nil {
		return nil, err
	}
	if s.o.chirp && !bytes.Equal(b[:4], cmd) {
		return nil, rErr("risposta inattesa alla lettura a %#04x (% X)", addr, b[:4])
	}
	return b[4:], nil
}

func (s *wfSess) write(addr int, d []byte) error {
	pkt := append([]byte{0x57, byte(addr >> 8), byte(addr), byte(s.o.page)}, d...)
	return s.expect(pkt, []byte{0x06}, fmt.Sprintf("scrittura a %#04x", addr))
}

// wfBlank: riga tutta a 00 o tutta a FF
func wfBlank(r []byte) bool {
	return bytes.Count(r, []byte{0}) == len(r) || bytes.Count(r, []byte{0xFF}) == len(r)
}

func wfOpen(ctx context.Context, m *genModel, port string) (*wfSess, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &wfSess{ctx: ctx, p: p, o: wfO(m)}, nil
}

func (f wfFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := wfOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	o := s.o
	total := m.Slots + 1
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, o.size(m))
	for i := 0; i <= m.Slots; i++ {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(i, total, "Lettura…")
		addr := o.start + i*16
		if i == m.Slots {
			addr = wfParamAddr
		}
		d, err := s.read(addr)
		if err != nil {
			return nil, "", genFail("Lettura", err)
		}
		copy(img[i*o.page:], d)
	}
	_ = s.p.Write([]byte{0x62})
	prog(total, total, "Lettura completata")
	return img, "", nil
}

func (f wfFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := wfO(m)
	if len(img) < o.size(m) {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.size(m))
	}
	s, err := wfOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	total := m.Slots + 1
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for i := 0; i <= m.Slots; i++ {
		prog(i, total, "Scrittura…")
		addr := o.start + i*16
		if i == m.Slots {
			addr = wfParamAddr
		}
		row := img[i*o.page : (i+1)*o.page]
		if i == m.Slots && wfBlank(row) {
			continue // immagine mai letta da una radio: le impostazioni della radio non si toccano
		}
		if err := s.write(addr, row); err != nil {
			return "", genFail("Scrittura", err)
		}
	}
	_ = s.p.Write([]byte{0x62})
	prog(total, total, "Scrittura completata")
	return "", nil
}
