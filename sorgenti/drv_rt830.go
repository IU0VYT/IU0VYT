// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-830 e RT-850: protocollo e memoria ricavati dai programmi del
// costruttore «Radtel RT830» («UV-83 Radio Program Software») e «Radtel RT-850
// programming software» («UV5118 Radio Program Software»), .NET, pubblicati
// sulla pagina dei download di radtels.com. I due programmi hanno la stessa
// struttura; cambiano gli indirizzi e l'ordine dei byte nel canale.
//
// Protocollo (9600 baud, 8N1), quello «iRadio» di drv_iradio.go:
//   39 33 05 10 81 → 0x06 (ingresso; il programma lo manda fino a tre volte);
//   lettura: 'R' indirizzo(2, big endian) somma → 'R' indirizzo 16 byte somma;
//   scrittura: 'W' indirizzo(2) 16 byte somma → 0x06; uscita: 39 33 05 EE 5F.
// La somma è quella dei byte precedenti, modulo 256.
// Lettura del programma: blocco della password (i primi 12 byte devono essere
// quelli scritti nella finestra: FF se non c'è password), poi impostazioni e
// canali; la RT-850 legge anche l'indirizzo 0x0000 e vuole 0x58 nel primo byte.
// Scrittura: blocco della password (stesso controllo), indirizzo 0x0000 (la
// RT-850 con lo stesso controllo), poi impostazioni e canali; la password si
// scrive solo se la si cambia: qui non si scrive mai.
//
//            password  impostazioni    canali (16 byte l'uno)
//   RT-830   0x01D0    0x01E0-0x020F   0x0210-0x0A3F: 128 canali, VFO A, VFO B e
//                                      un record letto ma non scritto
//   RT-850   0x0340    0x0350-0x038F   0x0390-0x0B8F: 128 canali
// L'immagine è la memoria dal blocco della password alla fine dei canali.
// Qui si scrivono le stesse aree del programma con i dati letti dalla radio:
// impostazioni e VFO tornano come sono, cambiano solo i canali modificati.
//
// Canale (RT-830 big endian, RT-850 little endian; libero = tutto a FF, cioè
// bit 0x20 del byte 12 acceso, o frequenza a zero):
//   0-3 RX e 6-9 TX in decine di Hz
//   4-5 tono RX, 10-11 tono TX nei 14 bit bassi: 0x3000 nessuno, 0x1000 + codice
//       ottale DCS normale, 0x2000 + codice DCS invertito, altrimenti decimi di
//       Hz; nei 2 bit alti il tipo di DCS (tono RX) e la cifratura (tono TX)
//   12  0x80 bassa potenza, 0x40 banda stretta, 0x02 limite di trasmissione,
//       0x01 in scansione; negli altri bit la gamma: RT-830 0x00 100-136 MHz,
//       0x04 136-174, 0x08 174-240, 0x0C 240-400, 0x10 400-520, 0x1C il resto;
//       RT-850 0x10 se RX è da 240 MHz in su, 0x08 se lo è TX
//   13-15 codice di silenziamento (big endian)

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
)

type rt830Opt struct {
	be    bool  // campi del canale big endian
	base  int   // indirizzo del blocco della password (inizio dell'immagine)
	cfg   int   // byte di impostazioni dopo la password
	recs  int   // record da 16 byte letti
	wrecs int   // record da 16 byte scritti
	model int   // primo byte atteso all'indirizzo 0x0000 (-1: non controllato)
	lo    int64 // limiti di frequenza del programma (Hz)
	hi    int64
}

func (o *rt830Opt) size() int  { return 16 + o.cfg + o.recs*16 }
func (o *rt830Opt) chOff() int { return 16 + o.cfg }

func (o *rt830Opt) u32(b []byte) int64 {
	if o.be {
		return int64(binary.BigEndian.Uint32(b))
	}
	return int64(binary.LittleEndian.Uint32(b))
}

func (o *rt830Opt) put32(b []byte, v int64) {
	if o.be {
		binary.BigEndian.PutUint32(b, uint32(v))
	} else {
		binary.LittleEndian.PutUint32(b, uint32(v))
	}
}

func (o *rt830Opt) u16(b []byte) int {
	if o.be {
		return int(binary.BigEndian.Uint16(b))
	}
	return int(binary.LittleEndian.Uint16(b))
}

func (o *rt830Opt) put16(b []byte, v int) {
	if o.be {
		binary.BigEndian.PutUint16(b, uint16(v))
	} else {
		binary.LittleEndian.PutUint16(b, uint16(v))
	}
}

// band: bit di gamma del byte 12, come li calcola il programma (frequenze in decine di Hz)
func (o *rt830Opt) band(rx, tx int64) byte {
	if !o.be {
		var b byte
		if rx >= 24000000 {
			b |= 0x10
		}
		if tx >= 24000000 {
			b |= 0x08
		}
		return b
	}
	switch {
	case rx >= 10000000 && rx < 13600000:
		return 0x00
	case rx >= 13600000 && rx <= 17400000:
		return 0x04
	case rx > 17400000 && rx < 24000000:
		return 0x08
	case rx >= 24000000 && rx < 40000000:
		return 0x0C
	case rx >= 40000000 && rx <= 52000000:
		return 0x10
	}
	return 0x1C
}

type rt830Fam struct{}

var (
	rt830Enter = []byte{0x39, 0x33, 0x05, 0x10, 0x81}
	rt830Exit  = []byte{0x39, 0x33, 0x05, 0xEE, 0x5F}
)

func init() {
	for _, x := range []struct {
		id, model, rclass string
		opt               rt830Opt
	}{
		{"rtl-rt830", "RT-830", "IU0VYT_Radtel_RT830",
			rt830Opt{be: true, base: 0x01D0, cfg: 0x30, recs: 131, wrecs: 130, model: -1, lo: 100000000, hi: 520000000}},
		{"rtl-rt850", "RT-850", "IU0VYT_Radtel_RT850",
			rt830Opt{base: 0x0340, cfg: 0x40, recs: 128, wrecs: 128, model: 0x58, lo: 20000000, hi: 660000000}},
	} {
		o := x.opt
		genRegister(&genModel{ID: x.id, Vendor: "Radtel", Model: x.model, RClass: x.rclass, Name: "Radtel " + x.model,
			Slots: 128, First: 1, NameMax: 0, Powers: []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}},
			Bands: []band{{o.lo, o.hi}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
			DCS: iradioDCS, Baud: 9600, Fam: rt830Fam{}, Opt: &o,
			Notes: "Radio senza nomi dei canali (128 canali). Protocollo e memoria ricavati dal programma del costruttore. " +
				"Qui si modificano frequenze, potenza, banda, scansione e toni; limite di trasmissione, cifratura, tipo di DCS e codice di " +
				"silenziamento del canale restano quelli impostati. Impostazioni e VFO della radio tornano come sono. " +
				"Se sulla radio è impostata una password, toglila prima con il programma del costruttore."})
	}
}

func rt830o(m *genModel) *rt830Opt { return m.Opt.(*rt830Opt) }

func (rt830Fam) rec(m *genModel, img []byte, i int) []byte {
	a := rt830o(m).chOff() + i*16
	return img[a : a+16]
}

// toni (SetSubaudio / SaveSubaudio del programma)
func rt830ToneDec(v int) string {
	switch v & 0x3000 {
	case 0x3000:
		return ""
	case 0x1000:
		return iradioDCSTone(int64(v&0x1FF), false)
	case 0x2000:
		return iradioDCSTone(int64(v&0x1FF), true)
	}
	if v&0xFFF == 0 {
		return ""
	}
	return fmtTone(v & 0xFFF)
}

func rt830ToneEnc(t genTone) (int, error) {
	switch t.Mode {
	case "Tone":
		if t.Hz10 <= 0 || t.Hz10 > 0xFFF {
			return 0, fmt.Errorf("tono %s non disponibile su questa radio", fmtTone(t.Hz10))
		}
		return t.Hz10, nil
	case "DTCS":
		if idxInt(iradioDCS, t.Code) < 0 {
			return 0, fmt.Errorf("codice DCS %s non disponibile su questa radio", fmt.Sprintf("%03d", t.Code))
		}
		if t.Inv {
			return int(iradioOct(t.Code)) | 0x2000, nil
		}
		return int(iradioOct(t.Code)) | 0x1000, nil
	}
	return 0x3000, nil
}

func (f rt830Fam) Get(m *genModel, img []byte, i int) *Chan {
	o := rt830o(m)
	r := f.rec(m, img, i)
	rx := o.u32(r[0:4])
	if r[12]&0x20 != 0 || rx == 0 || rx == 0xFFFFFFFF {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	if tx := o.u32(r[6:10]); tx == 0 || tx == 0xFFFFFFFF {
		c.TxOff = true
	} else {
		c.Tx = tx * 10
	}
	c.RxTone = rt830ToneDec(o.u16(r[4:6]))
	c.TxTone = rt830ToneDec(o.u16(r[10:12]))
	c.Power = m.powerSem(int(r[12] >> 7))
	c.Narrow = r[12]&0x40 != 0
	c.Scan = r[12]&0x01 != 0
	return c
}

func (f rt830Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, img, i)...)
}

func (f rt830Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(m, img, i), raw) }

func (f rt830Fam) Match(m *genModel, img []byte) bool { return false }

// fit (4.23): nelle conversioni i canali fuori dai limiti del programma non si
// copiano (con un avviso)
func (rt830Fam) fit(m *genModel, c *Chan) (bool, string) { return genBandFit(m, c) }

func (f rt830Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := rt830o(m)
	r := f.rec(m, img, i)
	if c == nil {
		copy(r, bytes.Repeat([]byte{0xFF}, 16))
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	for _, hz := range []int64{c.Rx, c.Tx} {
		if hz < o.lo || hz > o.hi {
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
	tv, err := rt830ToneEnc(tt)
	if err != nil {
		return err
	}
	rv, err := rt830ToneEnc(rt)
	if err != nil {
		return err
	}
	if f.Get(m, img, i) == nil {
		// canale nuovo come lo crea il programma: limite di trasmissione, cifratura
		// e tipo di DCS spenti, codice di silenziamento 0
		copy(r, make([]byte, 16))
	}
	rx, tx := c.Rx/10, c.Tx/10
	o.put32(r[0:4], rx)
	o.put32(r[6:10], tx)
	// i due bit alti dei toni (tipo di DCS, cifratura) restano
	o.put16(r[4:6], o.u16(r[4:6])&0xC000|rv)
	o.put16(r[10:12], o.u16(r[10:12])&0xC000|tv)
	// byte 12 come lo ricostruisce il programma
	b := r[12]&0x02 | o.band(rx, tx)
	if m.powerIndex(c.Power) == 1 {
		b |= 0x80
	}
	if c.Narrow {
		b |= 0x40
	}
	if c.Scan {
		b |= 0x01
	}
	r[12] = b
	return nil
}

// ---------------------------------------------------------------- protocollo

func rt830Proto() iradioProto {
	return iradioProto{enter: rt830Enter, exit: rt830Exit, enterTO: iradioFam{irAR518}.proto().enterTO, block: 0x10}
}

// rt830Start: ingresso in programmazione e controlli iniziali del programma
// (password; per la RT-850 anche il modello); restituisce il blocco della password
func (s *rawSess) rt830Start(o *rt830Opt, write bool) ([]byte, error) {
	if err := s.iradioEnter(rt830Proto()); err != nil {
		return nil, genFail("Ingresso in programmazione", err)
	}
	pw, err := s.iradioRead(o.base, 16)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pw[:12], bytes.Repeat([]byte{0xFF}, 12)) {
		return nil, rErr("sulla radio è impostata una password: toglila prima con il programma del costruttore")
	}
	if write || o.model >= 0 {
		id, err := s.iradioRead(0x0000, 16)
		if err != nil {
			return nil, err
		}
		if o.model >= 0 && int(id[0]) != o.model {
			return nil, rErr("identificativo della radio sconosciuto (% X): è il modello giusto?", id[:1])
		}
	}
	return pw, nil
}

func (f rt830Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rt830o(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.size(), "Collegamento con la radio…")
	data, err := s.rt830Start(o, false)
	if err != nil {
		_ = s.send(rt830Exit)
		return nil, "", err
	}
	defer s.send(rt830Exit)
	for a := 16; a < o.size(); a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a, o.size(), "Lettura…")
		b, err := s.iradioRead(o.base+a, 16)
		if err != nil {
			return nil, "", err
		}
		data = append(data, b...)
	}
	prog(o.size(), o.size(), "Lettura completata")
	return data, "", nil
}

func (f rt830Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rt830o(m)
	if len(img) < o.size() {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.size())
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	end := 16 + o.cfg + o.wrecs*16
	prog(0, end, "Collegamento con la radio…")
	if _, err := s.rt830Start(o, true); err != nil {
		_ = s.send(rt830Exit)
		return "", err
	}
	defer s.send(rt830Exit)
	for a := 16; a < end; a += 16 {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(a, end, "Scrittura…")
		if err := s.iradioWrite('W', o.base+a, img[a:a+16]); err != nil {
			return "", err
		}
	}
	prog(end, end, "Scrittura completata")
	return "", nil
}
