// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-69 GPS: protocollo e memoria ricavati dal programma del costruttore
// «RT-69 GPS» (889G.exe, .NET, progetto «UV8D»), pubblicato sulla pagina dei
// download di radtels.com.
//
// Protocollo (19200 baud, 8N1): gli stessi pacchetti in chiaro della Wouxun
// KG-UV8D (drv_kguv8.go), «7D comando FF lunghezza dati somma» (la somma non
// comprende il primo byte); la radio risponde «7D comando 00 lunghezza …»:
//   0x80 saluto → risposta con la somma giusta (il contenuto non si controlla);
//   0x82 lettura: indirizzo(2, big endian) lunghezza → indirizzo + dati;
//   0x83 scrittura: indirizzo(2) + dati → indirizzo; 0x81 fine.
// Senza risposta il programma aspetta 3 secondi e riprova tutto da capo fino a
// tre volte.
// Lettura: saluto, password (8 byte a 0x0870: se c'è, il programma la chiede),
// poi la memoria da 0x0480 a 0x7FFF in blocchi di 64 byte. Scrittura: saluto,
// password, limiti di frequenza della radio (0x0044 e 0x0054, 16 byte ognuno, e
// 0x02F0, 13 byte), controllo che ogni canale stia in quei limiti, poi gli stessi
// blocchi da 0x0480 a 0x7FFF. L'immagine è la memoria da 0x0000 a 0x7FFF (FF sotto
// 0x0480, che non si legge e non si scrive).
//
// Canale (16 byte da 0x0910, 999 canali; usato se il byte «valido» a 0x6701 +
// canale è 0x9E):
//   0-3 RX e 4-7 TX: numero di 32 bit big endian, decine di Hz (TX a FF = non
//       trasmette)
//   8-9 tono RX, 10-11 tono TX (big endian): 0 nessuno; 0x8000 + decimi di Hz;
//       0x4000 + codice DCS (in ottale), 0x2000 se invertito
//   12  bit 0-1 potenza (0 bassa, 1 media, 2 alta), bit 0x04 DCS «speciale»
//   13  bit 0x01 banda larga, bit 1-2 silenziamento (QT, QT+DTMF, QT*DTMF),
//       bit 0x20 in scansione
// Nome: 8 byte a 0x4788 + 8 × canale e altri 4 a 0x6FB0 + 4 × canale, con zeri
// in coda.
// Limiti della radio: se 0x0044 non è a FF, due bande come numeri di 32 bit
// (0x0044-0x004B e 0x0054-0x005B); altrimenti fino a tre bande in 0x02F0 (primo
// byte: bit 0, 1 e 2 = banda presente; poi inizio e fine in BCD, decimi di MHz).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	rt69Size    = 0x8000
	rt69From    = 0x0480 // primo indirizzo letto e scritto
	rt69Chan    = 0x0910
	rt69Name    = 0x4788
	rt69Name2   = 0x6FB0
	rt69Valid   = 0x6701
	rt69PwdAt   = 0x0870
	rt69Slots   = 999
	rt69LoHz    = 66025000 // limiti del programma
	rt69HiHz    = 599995000
	rt69Blocks  = (rt69Size - rt69From) / 64
	rt69Timeout = 3 * time.Second
)

// codici DCS dell'elenco del programma (114: i 104 più comuni e altri dieci)
var rt69DCS = []int{
	23, 25, 26, 31, 32, 36, 43, 47, 50, 51, 53, 54, 55, 65, 71, 72, 73, 74, 114, 115, 116, 122, 125, 131, 132,
	134, 135, 143, 145, 152, 155, 156, 162, 165, 172, 174, 205, 212, 217, 223, 225, 226, 243, 244, 245, 246,
	251, 252, 254, 255, 261, 263, 265, 266, 271, 274, 305, 306, 311, 315, 325, 331, 332, 343, 345, 346, 351,
	356, 364, 365, 371, 411, 412, 413, 423, 425, 431, 432, 445, 446, 452, 454, 455, 462, 464, 465, 466, 503,
	506, 516, 523, 526, 532, 534, 546, 565, 606, 612, 624, 627, 631, 632, 645, 654, 662, 664, 703, 712, 723,
	731, 732, 734, 743, 754,
}

// pacchetti della KG-UV8D, in chiaro
var rt69Wire = &wx8Opt{start: 0x7D}

const rt69Note = "999 canali con nomi di 12 caratteri, 19200 baud. Protocollo e memoria ricavati dal programma del costruttore «RT-69 GPS» e " +
	"provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, nome, toni, potenza, banda e scansione; " +
	"il resto del canale e le impostazioni della radio restano come sono. Prima di scrivere si controlla, come fa il programma, che " +
	"ogni canale stia nelle bande della radio. Se sulla radio è impostata una password va tolta prima con il programma del costruttore. " +
	"«Identifica» non riconosce questa radio: scegli tu la scheda giusta."

type rt69Opt struct{}

type rt69Fam struct{}

func init() {
	genRegister(&genModel{ID: "rtl-rt69", Vendor: "Radtel", Model: "RT-69 GPS", RClass: "IU0VYT_Radtel_RT69GPS", Name: "Radtel RT-69 GPS",
		Slots: rt69Slots, First: 1, NameMax: 12, Charset: charsetASCII,
		Powers: []genPower{{"Bassa", 0, 2}, {"Media", 0, 1}, {"Alta", 0, 0}},
		Bands:  []band{{rt69LoHz, rt69HiHz}},
		Narrow: true, HasScan: true, NoChirp: true, DCS: rt69DCS, Baud: 19200, Fam: rt69Fam{}, Opt: &rt69Opt{}, Notes: rt69Note})
}

func (rt69Fam) rec(img []byte, i int) []byte   { return img[rt69Chan+i*16 : rt69Chan+i*16+16] }
func (rt69Fam) name1(img []byte, i int) []byte { return img[rt69Name+i*8 : rt69Name+i*8+8] }
func (rt69Fam) name2(img []byte, i int) []byte { return img[rt69Name2+i*4 : rt69Name2+i*4+4] }

func (f rt69Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(img, i)...)
	out = append(out, f.name1(img, i)...)
	out = append(out, f.name2(img, i)...)
	return append(out, img[rt69Valid+i])
}

func (f rt69Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(f.rec(img, i), raw[:16])
	if len(raw) >= 29 {
		copy(f.name1(img, i), raw[16:24])
		copy(f.name2(img, i), raw[24:28])
		img[rt69Valid+i] = raw[28]
	}
}

func (rt69Fam) Match(m *genModel, img []byte) bool { return false }

// tono come lo legge il programma: 0 nessuno; bit 0x4000 = DCS (codice nei 9 bit
// bassi, in ottale; 0x2000 invertito); altrimenti decimi di Hz nei 12 bit bassi,
// da 0,1 a 254,1 Hz
func rt69ToneDec(v int) string {
	switch {
	case v == 0:
		return ""
	case v&0x4000 != 0:
		code := v & 0x1FF
		return genTone{Mode: "DTCS", Code: (code>>6)*100 + (code>>3&7)*10 + code&7, Inv: v&0x2000 != 0}.String()
	}
	if hz := v & 0x0FFF; hz > 0 && hz <= 2541 {
		return fmtTone(hz)
	}
	return ""
}

func rt69ToneEnc(t genTone) int {
	switch t.Mode {
	case "Tone":
		return 0x8000 | t.Hz10
	case "DTCS":
		v := 0x4000 | (t.Code/100)<<6 | (t.Code/10%10)<<3 | t.Code%10
		if t.Inv {
			v |= 0x2000
		}
		return v
	}
	return 0
}

func (f rt69Fam) Get(m *genModel, img []byte, i int) *Chan {
	if img[rt69Valid+i] != wx8Valid {
		return nil
	}
	r := f.rec(img, i)
	rx := int64(binary.BigEndian.Uint32(r[0:4]))
	if rx == 0xFFFFFFFF || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	if tx := int64(binary.BigEndian.Uint32(r[4:8])); tx == 0xFFFFFFFF || tx == 0 {
		c.TxOff = true
	} else {
		c.Tx = tx * 10
	}
	nm := append(append([]byte{}, f.name1(img, i)...), f.name2(img, i)...)
	if k := bytes.IndexByte(nm, 0); k >= 0 {
		nm = nm[:k]
	}
	c.Name = strings.TrimRight(string(nm), " ")
	c.RxTone = rt69ToneDec(int(binary.BigEndian.Uint16(r[8:10])))
	c.TxTone = rt69ToneDec(int(binary.BigEndian.Uint16(r[10:12])))
	// potenza come la mostra il programma: 0 bassa, 1 media, 2 e 3 alta
	c.Power = m.powerSem(min(int(r[12]&3), 2))
	c.Narrow = r[13]&0x01 == 0
	c.Scan = r[13]&0x20 != 0
	return c
}

// fit: i canali fuori da 66.025-599.995 MHz e i toni fuori da 60-254.1 Hz non si copiano
func (rt69Fam) fit(m *genModel, c *Chan) (bool, string) {
	for _, s := range []string{c.TxTone, c.RxTone} {
		if t, err := genToneParse(s); err == nil && t.Mode == "Tone" && (t.Hz10 < 600 || t.Hz10 > 2541) {
			return false, fmt.Sprintf("tono %s non disponibile su questa radio", fmtTone(t.Hz10))
		}
	}
	return genBandFit(m, c)
}

func (f rt69Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, n1, n2 := f.rec(img, i), f.name1(img, i), f.name2(img, i)
	if c == nil {
		// come «cancella riga» del programma
		copy(r, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF})
		copy(n1, make([]byte, 8))
		copy(n2, make([]byte, 4))
		img[rt69Valid+i] = 0
		return nil
	}
	hzs := []int64{c.Rx}
	if !c.TxOff {
		hzs = append(hzs, c.Tx)
	}
	for _, hz := range hzs {
		if !m.inBands(hz) {
			return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	for _, t := range []genTone{tt, rt} {
		if t.Mode == "DTCS" && idxInt(m.DCS, t.Code) < 0 {
			return fmt.Errorf("codice DCS %s non disponibile su questa radio", fmt.Sprintf("%03d", t.Code))
		}
		if t.Mode == "Tone" && (t.Hz10 < 600 || t.Hz10 > 2541) {
			return fmt.Errorf("tono %s non disponibile su questa radio", fmtTone(t.Hz10))
		}
	}
	old := f.Get(m, img, i)
	if old == nil {
		// canale nuovo come lo crea il programma: DCS «speciale», silenziamento QT
		copy(r, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0x04, 0x00, 0xFF, 0xFF})
		copy(n1, make([]byte, 8))
		copy(n2, make([]byte, 4))
	}
	binary.BigEndian.PutUint32(r[0:4], uint32(c.Rx/10))
	if c.TxOff {
		copy(r[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		binary.BigEndian.PutUint32(r[4:8], uint32(c.Tx/10))
	}
	binary.BigEndian.PutUint16(r[8:10], uint16(rt69ToneEnc(rt)))
	binary.BigEndian.PutUint16(r[10:12], uint16(rt69ToneEnc(tt)))
	r[12] = r[12]&0xFC | byte(m.powerIndex(c.Power))
	b := r[13] &^ 0x21
	if !c.Narrow {
		b |= 0x01
	}
	if c.Scan {
		b |= 0x20
	}
	r[13] = b
	// il nome non cambiato resta com'è (anche con caratteri che qui non si scrivono)
	if old == nil || c.Name != old.Name {
		nm := make([]byte, 12)
		copy(nm, c.Name)
		copy(n1, nm[:8])
		copy(n2, nm[8:])
	}
	img[rt69Valid+i] = wx8Valid
	return nil
}

// ---------------------------------------------------------------- protocollo

// rt69Reply: risposta della radio «7D comando 00 lunghezza dati somma», con
// l'intestazione e la somma controllate come fa il programma (3 secondi di attesa)
func (s *rawSess) rt69Reply(cmd byte) ([]byte, error) {
	h := readUpTo(s.ctx, s.p, 4, rt69Timeout)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(h) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if len(h) != 4 {
		return nil, rErr("intestazione della risposta incompleta")
	}
	body := readUpTo(s.ctx, s.p, int(h[3])+1, rt69Timeout)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(body) != int(h[3])+1 {
		return nil, rErr("risposta della radio incompleta")
	}
	n := len(body) - 1
	if h[0] != rt69Wire.start || h[1] != cmd || h[2] != 0 || body[n] != sum8(append(append([]byte{}, h[1:]...), body[:n]...)) {
		return nil, rErr("risposta inattesa")
	}
	return body[:n], nil
}

// rt69ReadAt: lettura di n byte all'indirizzo a (la risposta ripete l'indirizzo)
func (s *rawSess) rt69ReadAt(a, n int) ([]byte, error) {
	if err := s.wx8Write(rt69Wire, wx8CmdRd, []byte{byte(a >> 8), byte(a), byte(n)}); err != nil {
		return nil, err
	}
	resp, err := s.rt69Reply(wx8CmdRd)
	if err != nil {
		return nil, err
	}
	if len(resp) != 2+n || int(resp[0])<<8|int(resp[1]) != a {
		return nil, rErr("risposta inattesa")
	}
	return resp[2:], nil
}

// rt69Enter: saluto e password, come il programma
func (s *rawSess) rt69Enter() (string, error) {
	if err := s.wx8Write(rt69Wire, wx8CmdID, nil); err != nil {
		return "", err
	}
	id, err := s.rt69Reply(wx8CmdID)
	if err != nil {
		return "", err
	}
	pw, err := s.rt69ReadAt(rt69PwdAt, 8)
	if err != nil {
		return "", err
	}
	if pw[0] != 0 && pw[0] != 0xFF {
		_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
		return "", rErr("sulla radio è impostata una password: toglila prima con il programma del costruttore")
	}
	// nome del modello nella risposta al saluto, se è testo
	var name []byte
	for _, b := range id {
		if b < 0x20 || b > 0x7E {
			break
		}
		name = append(name, b)
	}
	return strings.TrimSpace(string(name)), nil
}

func (f rt69Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, rt69Blocks, "Collegamento con la radio…")
	id, err := s.rt69Enter()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := bytes.Repeat([]byte{0xFF}, rt69Size)
	for k := 0; k < rt69Blocks; k++ {
		if ctx.Err() != nil {
			_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
			return nil, "", errCancelled
		}
		prog(k, rt69Blocks, "Lettura…")
		a := rt69From + 64*k
		d, err := s.rt69ReadAt(a, 64)
		if err != nil {
			_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
			return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
		}
		copy(img[a:], d)
	}
	_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
	prog(rt69Blocks, rt69Blocks, "Lettura completata")
	info := ""
	if id != "" {
		info = fmt.Sprintf("identificativo %q", id)
	}
	return img, info, nil
}

// rt69Bcd: quattro cifre BCD (decimi di MHz) in Hz; falso se non sono cifre
func rt69Bcd(b []byte) (int64, bool) {
	v := int64(0)
	for _, x := range b[:2] {
		if x>>4 > 9 || x&0x0F > 9 {
			return 0, false
		}
		v = v*100 + int64(x>>4)*10 + int64(x&0x0F)
	}
	return v * 100000, true
}

// rt69Limits: bande della radio dai tre blocchi letti prima della scrittura
// (CheckFrequency del programma)
func rt69Limits(b1, b2, b3 []byte) ([]band, error) {
	var out []band
	if !bytes.Equal(b1[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		for _, b := range [][]byte{b1, b2} {
			out = append(out, band{int64(binary.BigEndian.Uint32(b[0:4])) * 10, int64(binary.BigEndian.Uint32(b[4:8])) * 10})
		}
		return out, nil
	}
	for k := 0; k < 3; k++ {
		if b3[0]&(1<<k) == 0 {
			continue
		}
		lo, ok1 := rt69Bcd(b3[1+4*k:])
		hi, ok2 := rt69Bcd(b3[3+4*k:])
		if !ok1 || !ok2 {
			return nil, rErr("risposta inattesa")
		}
		out = append(out, band{lo, hi})
	}
	return out, nil
}

// rt69Check: ogni frequenza dei canali usati deve stare in una banda della radio
func (f rt69Fam) rt69Check(m *genModel, img []byte, bands []band) error {
	in := func(hz int64) bool {
		for _, b := range bands {
			if hz >= b.Lo && hz <= b.Hi {
				return true
			}
		}
		return false
	}
	for i := 0; i < rt69Slots; i++ {
		c := f.Get(m, img, i)
		if c == nil {
			continue
		}
		if !in(c.Rx) {
			return rErr("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
		}
		if !c.TxOff && !in(c.Tx) {
			return rErr("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Tx))
		}
	}
	return nil
}

func (f rt69Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < rt69Size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), rt69Size)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, rt69Blocks, "Collegamento con la radio…")
	if _, err := s.rt69Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	fail := func(stage string, err error) (string, error) {
		_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
		return "", genFail(stage, err)
	}
	var lim [3][]byte
	for k, q := range [][2]int{{0x0044, 16}, {0x0054, 16}, {0x02F0, 13}} {
		if lim[k], err = s.rt69ReadAt(q[0], q[1]); err != nil {
			return fail("Identificazione", err)
		}
	}
	bands, err := rt69Limits(lim[0], lim[1], lim[2])
	if err != nil {
		return fail("Identificazione", err)
	}
	if err := f.rt69Check(m, img, bands); err != nil {
		return fail("Scrittura", err)
	}
	for k := 0; k < rt69Blocks; k++ {
		prog(k, rt69Blocks, "Scrittura…")
		a := rt69From + 64*k
		if err := s.wx8Write(rt69Wire, wx8CmdWr, append([]byte{byte(a >> 8), byte(a)}, img[a:a+64]...)); err != nil {
			return "", err
		}
		resp, err := s.rt69Reply(wx8CmdWr)
		if err == nil && (len(resp) != 2 || int(resp[0])<<8|int(resp[1]) != a) {
			err = rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
		if err != nil {
			return fail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
	}
	_ = s.wx8Write(rt69Wire, wx8CmdEnd, nil)
	prog(rt69Blocks, rt69Blocks, "Scrittura completata")
	return "", nil
}
