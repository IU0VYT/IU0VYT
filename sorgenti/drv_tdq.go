// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Famiglia «tdq» (4.30): i TIDRADIO senza display programmati dai programmi
// «TD-FRS» / «TD-PMR» (TID-712Plus, TD-M7, TD-M11 Plus, TD-BR09, TD-777S, TD-V900,
// TD-FRS, TD-PMR) e «TD-M2Pro». Protocollo e memoria sono stati ricavati leggendo
// quei programmi (solo lettura statica) e riscritti da zero:
//
//   - 38400 baud, 8N1; il PC manda la chiave di 7 lettere («QROGRAM», per il
//     TD-M2Pro «QROGRAO») e la radio risponde 07; poi «123456» + 7 byte a FF,
//     risposta 55; poi 08, e la radio risponde con un byte;
//   - lettura: 52 + indirizzo (2 byte, prima il più alto) + 40; la radio risponde
//     con 4 byte (il 2° e il 3° ripetono l'indirizzo) + 64 di dati; fra un blocco
//     e l'altro il PC manda 08 e aspetta 08; alla fine manda 45;
//   - scrittura: 57 + indirizzo + 10 + 16 byte di dati, risposta 08;
//   - memoria 0x000-0xE7F: 200 canali di 16 byte, poi le impostazioni (0xC80),
//     i codici DTMF e, da 0xCD0, i codici dello scrambler di ogni canale;
//   - canale: ricezione e trasmissione in BCD (4 byte, dal meno significativo,
//     unità di 10 Hz), toni (2 byte: CTCSS in decimi di Hz, DCS 0x2800 + codice,
//     0xA800 se invertito), byte 12 = potenza (bit 7-6), funzione (5-4),
//     scrambler (3-1), banda larga (bit 0), byte 13 = scansione (bit 6), blocco
//     su canale occupato (5), salto di frequenza (4), PTT-ID (3-2), modo di
//     ricezione (1-0). I canali non hanno nome.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	tdqSize    = 0xE80 // memoria letta e scritta dal programma
	tdqChans   = 0xC80 // fine dei canali, inizio delle impostazioni
	tdqSetAddr = 0xC80
)

type tdqOpt struct {
	magic   string
	pow     []int // valore nella radio di ogni livello di m.Powers
	toneOff byte  // come il programma scrive «nessun tono»: 00 00 oppure FF FF
	prog    string
}

type tdqFam struct{}

func tdqO(m *genModel) *tdqOpt { return m.Opt.(*tdqOpt) }

// tdqNote: nota della scheda. I modelli con la chiave «QROGRAM» hanno lo stesso
// collegamento e «Identifica» non li distingue; la TD-M2 Pro («QROGRAO») è la
// sola con la sua (4.32: la frase non compare più nella sua nota). Due testi
// interi, perché ognuno ha la sua traduzione.
func tdqNote(magic, prog, extra string) string {
	if magic != "QROGRAM" {
		return strings.TrimSpace(fmt.Sprintf("38400 baud. Protocollo e memoria ricavati dal programma del costruttore %s e provati con una radio simulata, non su una radio vera. "+
			"Qui si modificano frequenze (a passi di 1,25 kHz), toni, potenza, banda e scansione dei canali (non hanno nome) e le impostazioni generali; "+
			"il resto del canale (scrambler, PTT-ID, blocco su canale occupato) non si tocca. %s", prog, extra))
	}
	return fmt.Sprintf("38400 baud. Protocollo e memoria ricavati dal programma del costruttore %s e provati con una radio simulata, non su una radio vera. "+
		"Qui si modificano frequenze (a passi di 1,25 kHz), toni, potenza, banda e scansione dei canali (non hanno nome) e le impostazioni generali; "+
		"il resto del canale (scrambler, PTT-ID, blocco su canale occupato) non si tocca. %s"+
		"«Identifica» non distingue i modelli di questa famiglia: scegli tu la scheda.", prog, extra)
}

func init() {
	lmh := []genPower{{"Bassa", 0, 2}, {"Media", 0, 1}, {"Alta", 0, 0}}
	lh := []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}}
	vu := []band{{136000000, 174000000}, {400000000, 520000000}}
	uhf := []band{{400000000, 520000000}}
	frs := "«TD-FRS» / «TD-PMR» (versione del 25/6/2026)"
	mk := func(id, model, name string, slots int, bands []band, pw []genPower, pow []int, magic string, off byte, prog, extra string) *genModel {
		return &genModel{ID: id, Vendor: "TIDRADIO", Model: model, RClass: "IU0VYT_TIDRADIO_" + id[3:], Name: name, Slots: slots, First: 1,
			Powers: pw, Bands: bands, Narrow: true, HasScan: true, NoChirp: true, DCS: dtcsAll(), Baud: 38400, Fam: tdqFam{},
			Opt:   &tdqOpt{magic: magic, pow: pow, toneOff: off, prog: prog},
			Notes: tdqNote(magic, prog, extra)}
	}
	free := "È una radio di libero uso: fuori dai suoi canali e dalla sua potenza di fabbrica non è più a norma. "
	genRegister(
		mk("td-712plus", "TID-712Plus", "TIDRADIO TID-712Plus (TD-V712 Plus)", 199, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-m7", "TD-M7", "TIDRADIO TD-M7", 199, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-m11plus", "TD-M11 Plus", "TIDRADIO TD-M11 Plus (199 canali)", 199, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-m11plus-16", "TD-M11 Plus", "TIDRADIO TD-M11 Plus (16 canali)", 16, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-m11plus-22", "TD-M11 Plus", "TIDRADIO TD-M11 Plus (22 canali)", 22, vu, lh, []int{0, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-br09", "TD-BR09", "TIDRADIO TD-BR09", 199, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-777s-16", "TD-777S", "TIDRADIO TD-777S (16 canali)", 16, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-777s-22", "TD-777S", "TIDRADIO TD-777S (22 canali)", 22, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-v900", "TD-V900", "TIDRADIO TD-V900", 199, vu, lmh, []int{0, 1, 2}, "QROGRAM", 0x00, frs, ""),
		mk("td-frs", "TD-FRS", "TIDRADIO TD-FRS (22 canali FRS)", 22, vu, lh, []int{0, 2}, "QROGRAM", 0x00, frs, free),
		mk("td-pmr", "TD-PMR", "TIDRADIO TD-PMR (16 canali PMR446)", 16, vu, lh, []int{0, 2}, "QROGRAM", 0x00, frs, free),
		mk("td-m2pro", "TD-M2 Pro", "TIDRADIO TD-M2 Pro", 199, uhf, lmh, []int{0, 1, 2}, "QROGRAO", 0xFF, "«TD-M2Pro» (V1.0 e V1.2)",
			"Il programma accetta solo frequenze UHF. "),
	)
}

func (tdqFam) rec(img []byte, i int) []byte { return img[i*16 : (i+1)*16] }

func (f tdqFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f tdqFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (tdqFam) Match(m *genModel, img []byte) bool { return false }

// tdqFreq: frequenza in Hz da 4 byte BCD (dal meno significativo, unità di 10 Hz); 0 se non valida
func tdqFreq(b []byte) int64 {
	var v int64
	for i := 3; i >= 0; i-- {
		hi, lo := int64(b[i]>>4), int64(b[i]&0x0F)
		if hi > 9 || lo > 9 {
			return 0
		}
		v = v*100 + hi*10 + lo
	}
	return v * 10
}

func tdqPutFreq(b []byte, hz int64) {
	v := hz / 10
	for i := 0; i < 4; i++ {
		d := v % 100
		b[i] = byte(d/10<<4 | d%10)
		v /= 100
	}
}

func tdqToneDec(b []byte) string {
	v := int(b[1])<<8 | int(b[0])
	switch {
	case b[1]>>4 == 0: // CTCSS in decimi di Hz
		if b[0] != 0 && b[0] != 0xFF && v >= 600 && v <= 2600 {
			return genTone{Mode: "Tone", Hz10: v}.String()
		}
	case b[0] != 0 && b[0] <= 0xF0: // DCS: tre cifre ottali nei 9 bit bassi; 2 = normale, altrimenti invertito
		n := v & 0x1FF
		return genTone{Mode: "DTCS", Code: (n>>6)*100 + ((n>>3)&7)*10 + (n & 7), Inv: b[1]>>4 != 2}.String()
	}
	return ""
}

func tdqToneEnc(dst []byte, s string, off byte) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch t.Mode {
	case "Tone":
		if t.Hz10 < 600 || t.Hz10 > 2600 {
			return fmt.Errorf("tono %s non previsto dalla radio", s)
		}
		dst[0], dst[1] = byte(t.Hz10), byte(t.Hz10>>8)
	case "DTCS":
		d2, d1, d0 := t.Code/100, t.Code/10%10, t.Code%10
		if t.Code <= 0 || d2 > 7 || d1 > 7 || d0 > 7 {
			return fmt.Errorf("codice DCS %s non previsto dalla radio", s)
		}
		v := 0x2800 | d2<<6 | d1<<3 | d0
		if t.Inv {
			v |= 0x8000
		}
		dst[0], dst[1] = byte(v), byte(v>>8)
	default:
		dst[0], dst[1] = off, off
	}
	return nil
}

func (f tdqFam) Get(m *genModel, img []byte, i int) *Chan {
	o := tdqO(m)
	r := f.rec(img, i)
	if r[0] == 0xFF || r[3] == 0xFF {
		return nil
	}
	rx := tdqFreq(r[0:4])
	if !m.inBands(rx) {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if tx := tdqFreq(r[4:8]); r[7] != 0xFF && m.inBands(tx) {
		c.Tx = tx
	} else {
		c.TxOff = true
	}
	c.RxTone, c.TxTone = tdqToneDec(r[8:10]), tdqToneDec(r[10:12])
	pw := 0 // un valore che il modello non prevede si vede come il primo livello, come nel programma
	for k, v := range o.pow {
		if v == int(r[12]>>6) {
			pw = k
		}
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r[12]&1 == 0
	c.Scan = r[13]&0x40 != 0
	return c
}

func (f tdqFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := tdqO(m)
	r := f.rec(img, i)
	if c == nil { // canale vuoto come lo scrive il programma
		for k := range r {
			r[k] = 0xFF
		}
		return nil
	}
	if ok, why := f.fit(m, c); !ok {
		return fmt.Errorf("%s", why)
	}
	var rt, tt [2]byte
	if err := tdqToneEnc(rt[:], c.RxTone, o.toneOff); err != nil {
		return err
	}
	if err := tdqToneEnc(tt[:], c.TxTone, o.toneOff); err != nil {
		return err
	}
	if f.Get(m, img, i) == nil { // canale nuovo: i valori di partenza del programma
		for k := range r {
			r[k] = 0xFF
		}
		r[12] = 7 << 1 // scrambler 8 (spento: funzione «nessuna»)
		r[13] = 0x01   // modo di ricezione 1
	}
	tdqPutFreq(r[0:4], c.Rx)
	if c.TxOff {
		copy(r[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		tdqPutFreq(r[4:8], c.Tx)
	}
	copy(r[8:10], rt[:])
	copy(r[10:12], tt[:])
	r[12] = r[12]&0x3E | byte(o.pow[m.powerIndex(c.Power)])<<6
	if !c.Narrow {
		r[12] |= 0x01
	}
	r[13] &^= 0x40
	if c.Scan {
		r[13] |= 0x40
	}
	return nil
}

func (tdqFam) fit(m *genModel, c *Chan) (bool, string) {
	fs := []int64{c.Rx}
	if !c.TxOff {
		fs = append(fs, c.Tx)
	}
	for _, hz := range fs {
		if !m.inBands(hz) {
			return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
		if hz%1250 != 0 {
			return false, fmt.Sprintf("frequenza %s MHz: la radio accetta solo passi di 1,25 kHz", fmtMHz(hz))
		}
	}
	return true, ""
}

// chirpInvalid: modifiche di prova che il modello non può memorizzare
func (f tdqFam) chirpInvalid(m *genModel, c *Chan) bool {
	ok, _ := f.fit(m, c)
	return !ok
}

// ---------------------------------------------------------------- protocollo

// tdqWait: attesa di ogni risposta; le prove la accorciano
var tdqWait = 2 * time.Second

var tdqPass = []byte("123456\xFF\xFF\xFF\xFF\xFF\xFF\xFF")

type tdqSess struct {
	ctx context.Context
	p   Port
	o   *tdqOpt
}

func (s *tdqSess) xfer(send []byte, n int) ([]byte, error) {
	if err := s.p.Write(send); err != nil {
		return nil, err
	}
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, tdqWait); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *tdqSess) expect(send []byte, want byte, what string) error {
	got, err := s.xfer(send, 1)
	if err != nil {
		return err
	}
	if got[0] != want {
		return rErr("%s: risposta inattesa della radio (%02X invece di %02X)", what, got[0], want)
	}
	return nil
}

func (s *tdqSess) enter() error {
	_ = s.p.Purge()
	if err := s.expect([]byte(s.o.magic), 0x07, "aggancio"); err != nil {
		return err
	}
	if err := s.expect(tdqPass, 0x55, "chiave"); err != nil {
		return err
	}
	_, err := s.xfer([]byte{0x08}, 1) // il programma accetta qualunque byte
	return err
}

func tdqOpen(ctx context.Context, m *genModel, port string) (*tdqSess, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &tdqSess{ctx: ctx, p: p, o: tdqO(m)}, nil
}

func (f tdqFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := tdqOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	const blk = 0x40
	total := tdqSize / blk
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, tdqSize)
	for addr := 0; addr < tdqSize; addr += blk {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(addr/blk, total, "Lettura…")
		if addr > 0 {
			if err := s.expect([]byte{0x08}, 0x08, "conferma"); err != nil {
				return nil, "", genFail("Lettura", err)
			}
		}
		b, err := s.xfer([]byte{0x52, byte(addr >> 8), byte(addr), blk}, blk+4)
		if err != nil {
			return nil, "", genFail("Lettura", err)
		}
		if b[1] != byte(addr>>8) || b[2] != byte(addr) {
			return nil, "", genFail("Lettura", rErr("risposta inattesa alla lettura a %#04x (% X)", addr, b[:4]))
		}
		copy(img[addr:], b[4:])
	}
	_ = s.p.Write([]byte{0x45})
	prog(total, total, "Lettura completata")
	return img, "", nil
}

func (f tdqFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < tdqSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), tdqSize)
	}
	end := tdqSize
	if tdqBlankSet(img) {
		end = tdqChans // immagine mai letta da una radio: le impostazioni della radio non si toccano
	}
	s, err := tdqOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	total := end / 16
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for addr := 0; addr < end; addr += 16 {
		prog(addr/16, total, "Scrittura…")
		pkt := append([]byte{0x57, byte(addr >> 8), byte(addr), 0x10}, img[addr:addr+16]...)
		if err := s.expect(pkt, 0x08, fmt.Sprintf("scrittura a %#04x", addr)); err != nil {
			return "", genFail("Scrittura", err)
		}
	}
	prog(total, total, "Scrittura completata")
	return "", nil
}

// tdqBlankSet: riga delle impostazioni mai letta
func tdqBlankSet(img []byte) bool {
	return len(img) < tdqSize || bytes.Count(img[tdqSetAddr:tdqSetAddr+16], []byte{0xFF}) == 16 || bytes.Count(img[tdqSetAddr:tdqSetAddr+16], []byte{0}) == 16
}
