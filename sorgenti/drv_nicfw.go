// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.15: firmware libero nicFW di Marcus Dudley «nicsure» (https://nicsure.co.uk).
// Codice scritto da zero dalla documentazione pubblicata dall'autore:
//   - TIDRADIO TD-H3, nicFW 2.x: https://github.com/nicsure/nicfw2docs
//     (memoria di 8 KiB a blocchi da 32 byte; 198 canali da 32 byte; numeri in
//     ordine «little endian» nelle versioni 2.0x e «big endian» nelle 2.5x, che
//     si riconoscono dal numero magico delle impostazioni a 0x1900);
//   - Radtel RT-880 / iRadio UV-98 e Radtel RT-950 Pro, nicFW880 / nicFW950:
//     https://github.com/nicsure/nicFW880_950_Docs (memoria «storage» a blocchi
//     da 128 byte; 999 canali da 62 byte a partire da 0x50200).
//
// Regole di prudenza di questo programma:
//   - la taratura della radio (TD-H3: da 0x1DE0 in poi; RT-880: 0x5FA00-0x5FAFF)
//     non si scrive mai;
//   - prima di scrivere si rilegge la radio e si scrivono solo i blocchi cambiati;
//   - la scrittura parte solo se la radio ha la stessa versione della memoria
//     dell'immagine (numero magico).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------- parti comuni

// nicfwTone: tono come lo scrive nicFW (0 = nessuno; subtono in decimi di Hz;
// DCS = 0x8000 + codice, 0x4000 = invertito)
func nicfwTone(s string) (uint16, error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, err
	}
	switch t.Mode {
	case "Tone":
		if t.Hz10 <= 0 || t.Hz10 > 3000 {
			return 0, fmt.Errorf("tono non disponibile su questa radio")
		}
		return uint16(t.Hz10), nil
	case "DTCS":
		// il codice (es. 754) si scrive come numero in base 8
		v := (t.Code/100%10)<<6 | (t.Code/10%10)<<3 | t.Code%10
		if v < 1 || v > 511 || t.Code > 777 {
			return 0, fmt.Errorf("tono non disponibile su questa radio")
		}
		w := uint16(0x8000 | v)
		if t.Inv {
			w |= 0x4000
		}
		return w, nil
	}
	return 0, nil
}

func nicfwToneStr(w uint16) string {
	switch {
	case w == 0:
		return ""
	case w&0x8000 != 0:
		v := int(w & 0x1FF)
		if v == 0 {
			return ""
		}
		pol := "N"
		if w&0x4000 != 0 {
			pol = "I"
		}
		return fmt.Sprintf("D%d%d%d%s", v>>6&7, v>>3&7, v&7, pol)
	case w <= 3000:
		return fmtTone(int(w))
	}
	return ""
}

func nicfwName(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		if c == 0 || c == 0xFF {
			break
		}
		if c >= 0x20 && c < 0x7F {
			sb.WriteByte(c)
		}
	}
	return strings.TrimRight(sb.String(), " ")
}

// nicfwPower: potenza del canale (numero da 1 a max) per «alta, media, bassa»;
// se il canale aveva già una potenza della stessa fascia, resta quella
func nicfwPower(sem int, old, max int) int {
	if max < 1 || max > 255 {
		max = 255
	}
	if old >= 1 && old <= 255 && nicfwSem(old, max) == sem {
		return old
	}
	p := max
	switch sem {
	case 1:
		p = (max + 1) / 2
	case 2:
		p = (max + 4) / 5
	}
	if p < 1 {
		p = 1
	}
	return p
}

func nicfwSem(p, max int) int {
	if max < 1 || max > 255 {
		max = 255
	}
	switch {
	case p*4 >= max*3:
		return 0
	case p*20 >= max*7:
		return 1
	}
	return 2
}

func nicfwSum(b []byte) byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return s
}

var nicfwPowers = []genPower{{"Alta (massima)", 5, 0}, {"Media (metà)", 2.5, 1}, {"Bassa (un quinto)", 1, 2}}

const nicfwPowerNote = "Potenza: alta (il massimo), media (metà) o bassa (un quinto); una potenza già impostata sulla radio resta com'è. "

const nicfwSafety = "La taratura della radio non viene mai scritta; prima di scrivere il programma rilegge la radio e scrive solo quello che è cambiato. " +
	"Dalla documentazione pubblicata dall'autore del firmware."

// ---------------------------------------------------------------- TD-H3, nicFW 2.x

type nicfwFam struct{}

const (
	nicfwSize   = 0x2000
	nicfwSetAt  = 0x1900
	nicfwCalib  = 0x1DE0 // da qui in poi: taratura e tabelle delle potenze
	nicfwChans  = 198
	nicfwMinHz  = 18000000
	nicfwMaxHz  = 1299999990
	nicfwMagic5 = 0xD82F // impostazioni delle versioni 2.5x (numeri «big endian»)
	nicfwMagic0 = 0x9BCF // impostazioni delle versioni 2.0x (numeri «little endian»)
)

func init() {
	genRegister(&genModel{ID: "td-tdh3-nicfw", Vendor: "TIDRADIO", Model: "TD-H3", Variant: "nicFW 2.x", RClass: "IU0VYT_TDH3_nicFW", Name: "TIDRADIO TD-H3 (firmware nicFW)",
		Slots: nicfwChans, First: 1, NameMax: 12, Charset: charsetASCII, Powers: nicfwPowers,
		Bands: []band{{nicfwMinHz, nicfwMaxHz}}, Narrow: true, NoChirp: true, DCS: dtcsCodes, Baud: 38400, Fam: nicfwFam{},
		Notes: "TD-H3 con il firmware libero nicFW 2.x (versioni 2.0x e 2.5x): 198 canali con nomi di 12 caratteri, 38400 baud, radio accesa. " + nicfwPowerNote + nicfwSafety})
}

func (nicfwFam) layout(m *genModel) *bwEl { return bwMust("#seekto 0x1FFF;\nu8 fine;") }

// be: numeri «big endian» (2.5x)? ok = false se la memoria non è di nicFW 2.x
func nicfwOrder(img []byte) (order binary.ByteOrder, ok bool) {
	if len(img) < nicfwSetAt+2 {
		return binary.BigEndian, false
	}
	switch {
	case binary.BigEndian.Uint16(img[nicfwSetAt:]) == nicfwMagic5:
		return binary.BigEndian, true
	case binary.LittleEndian.Uint16(img[nicfwSetAt:]) == nicfwMagic0:
		return binary.LittleEndian, true
	}
	return binary.BigEndian, false
}

func (nicfwFam) at(i int) int { return 0x40 + i*32 }

func (f nicfwFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, img[f.at(i):f.at(i)+32]...)
}

func (f nicfwFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) == 32 {
		copy(img[f.at(i):], raw)
	}
}

func (nicfwFam) Match(m *genModel, img []byte) bool {
	_, ok := nicfwOrder(img)
	return len(img) == nicfwSize && ok
}

// potenza massima tarata (VHF sotto i 300 MHz, UHF sopra)
func nicfwMaxPower(img []byte, hz int64) int {
	if hz < 300000000 {
		return int(img[0x1DFF])
	}
	return int(img[0x1DFD])
}

func (f nicfwFam) Get(m *genModel, img []byte, i int) *Chan {
	o, _ := nicfwOrder(img)
	r := img[f.at(i) : f.at(i)+32]
	rx := int64(o.Uint32(r[0:])) * 10
	if rx < nicfwMinHz || rx > nicfwMaxHz {
		return nil
	}
	c := &Chan{Rx: rx, Tx: int64(o.Uint32(r[4:])) * 10, Scan: true}
	if c.Tx < nicfwMinHz || c.Tx > nicfwMaxHz {
		c.Tx = rx
	}
	c.RxTone, c.TxTone = nicfwToneStr(o.Uint16(r[8:])), nicfwToneStr(o.Uint16(r[10:]))
	if p := int(r[12]); p == 0 {
		c.TxOff, c.Tx = true, rx
	} else {
		c.Power = nicfwSem(p, nicfwMaxPower(img, c.Tx))
	}
	c.Narrow = r[15]&1 != 0
	c.Name = nicfwName(r[20:32])
	return c
}

// fit (4.23): nelle conversioni un canale con una frequenza che la radio non può
// memorizzare non viene copiato (con un avviso), invece di fermare tutto
func (nicfwFam) fit(m *genModel, c *Chan) (bool, string) {
	return genRangeFit(c, nicfwMinHz, nicfwMaxHz)
}
func (nicfw880Fam) fit(m *genModel, c *Chan) (bool, string) {
	return genRangeFit(c, nicfwMinHz, nicfwMaxHz)
}

func (f nicfwFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o, ok := nicfwOrder(img)
	r := img[f.at(i) : f.at(i)+32]
	if c == nil {
		for k := range r {
			r[k] = 0
		}
		return nil
	}
	if !ok {
		return fmt.Errorf("la memoria non è quella del firmware nicFW 2.x")
	}
	if c.Rx%10 != 0 || c.Rx < nicfwMinHz || c.Rx > nicfwMaxHz {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	if tx%10 != 0 || tx < nicfwMinHz || tx > nicfwMaxHz {
		return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(tx))
	}
	rt, err := nicfwTone(c.RxTone)
	if err != nil {
		return err
	}
	tt, err := nicfwTone(c.TxTone)
	if err != nil {
		return err
	}
	oldRx := int64(o.Uint32(r[0:])) * 10
	if oldRx < nicfwMinHz || oldRx > nicfwMaxHz {
		// canale nuovo: nessun gruppo, modulazione automatica
		for k := range r {
			r[k] = 0
		}
	}
	old := int(r[12])
	o.PutUint32(r[0:], uint32(c.Rx/10))
	o.PutUint32(r[4:], uint32(tx/10))
	o.PutUint16(r[8:], rt)
	o.PutUint16(r[10:], tt)
	if c.TxOff {
		r[12] = 0
	} else {
		r[12] = byte(nicfwPower(c.Power, old, nicfwMaxPower(img, tx)))
	}
	r[15] &^= 1
	if c.Narrow {
		r[15] |= 1
	}
	if tx != c.Rx {
		r[15] &^= 0x80 // il blocco su canale occupato non vale con i ponti (regola del firmware)
	}
	name := m.cleanName(c.Name)
	for k := 0; k < 12; k++ {
		r[20+k] = 0
		if k < len(name) {
			r[20+k] = name[k]
		}
	}
	return nil
}

func nicfwAck(s *rawSess, cmd byte) error {
	if err := s.send([]byte{cmd}); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return errTimeout
	}
	if b[0] != cmd {
		return rErr("risposta inattesa della radio")
	}
	return nil
}

// nicfwLast: ultimo comando senza conferma (ripresa della radio, riavvio); una
// breve attesa lascia uscire i byte prima che la porta si chiuda
func nicfwLast(s *rawSess, cmd ...byte) {
	if s.send(cmd) == nil {
		_ = s.sleep(200 * time.Millisecond)
	}
}

func nicfwReadBlock(s *rawSess, blk int) ([]byte, error) {
	if err := s.send([]byte{0x30, byte(blk)}); err != nil {
		return nil, err
	}
	b := readUpTo(s.ctx, s.p, 34, 2*time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return nil, errTimeout
	}
	if len(b) != 34 || b[0] != 0x30 {
		return nil, rErr("blocco della radio non valido (0x%04X)", blk*32)
	}
	if nicfwSum(b[1:33]) != b[33] {
		return nil, rErr("somma di controllo errata nel blocco 0x%04X: ripeti la lettura", blk*32)
	}
	return b[1:33], nil
}

// nicfwReadAll: tutta la memoria; la radio viene fermata e poi riavviata
func nicfwReadAll(s *rawSess, prog Progress, what string) ([]byte, error) {
	if err := nicfwAck(s, 0x45); err != nil {
		return nil, genFail("Identificazione", err)
	}
	img := make([]byte, 0, nicfwSize)
	for blk := 0; blk < nicfwSize/32; blk++ {
		b, err := nicfwReadBlock(s, blk)
		if err != nil {
			nicfwLast(s, 0x46)
			return nil, err
		}
		img = append(img, b...)
		prog(len(img), nicfwSize, what)
	}
	return img, nil
}

func (f nicfwFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, nicfwSize, "Collegamento con la radio…")
	img, err := nicfwReadAll(s, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	if err := nicfwAck(s, 0x46); err != nil {
		return nil, "", err
	}
	if _, ok := nicfwOrder(img); !ok {
		return nil, "", rErr("la radio non ha il firmware nicFW 2.x (numero magico %02X%02X)", img[nicfwSetAt], img[nicfwSetAt+1])
	}
	prog(nicfwSize, nicfwSize, "Lettura completata")
	return img, "", nil
}

func (f nicfwFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, nicfwSize, "Collegamento con la radio…")
	now, err := nicfwReadAll(s, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		return "", err
	}
	if !bytes.Equal(now[nicfwSetAt:nicfwSetAt+2], img[nicfwSetAt:nicfwSetAt+2]) {
		nicfwLast(s, 0x46)
		return "", rErr("la radio ha un'altra versione del firmware nicFW rispetto all'immagine: leggi la radio e riporta i canali su quella lettura")
	}
	written := 0
	for blk := 0; blk < nicfwCalib/32; blk++ {
		a := blk * 32
		if bytes.Equal(now[a:a+32], img[a:a+32]) {
			continue
		}
		pkt := append([]byte{0x31, byte(blk)}, img[a:a+32]...)
		pkt = append(pkt, nicfwSum(img[a:a+32]))
		if err := s.send(pkt); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return "", errCancelled
		}
		if len(b) != 1 || b[0] != 0x31 {
			nicfwLast(s, 0x46)
			return "", rErr("la radio non ha confermato il blocco 0x%04X", a)
		}
		written++
		prog(a+32, nicfwSize, "Scrittura…")
	}
	if err := nicfwAck(s, 0x46); err != nil {
		return "", err
	}
	if written > 0 {
		// riavvio: le modifiche valgono dalla riaccensione (la radio non conferma)
		nicfwLast(s, 0x49)
	}
	prog(nicfwSize, nicfwSize, "Scrittura completata")
	return "", nil
}

// ---------------------------------------------------------------- RT-880 / RT-950 Pro, nicFW880

type nicfw880Fam struct{}

const (
	nicfw880Base   = 0x50200 // primo indirizzo letto (canali)
	nicfw880Size   = 0xFF00  // fino a 0x60100: canali, impostazioni, taratura, gruppi, tasti, piano delle bande
	nicfw880Rec    = 62
	nicfw880Chans  = 999
	nicfw880Magic  = 0xABC6 // canale in uso
	nicfw880SetAt  = 0x5F600 - nicfw880Base
	nicfw880CalLo  = 0x5FA00 - nicfw880Base // taratura e fasce di potenza: mai scritte
	nicfw880CalHi  = 0x5FB00 - nicfw880Base
	nicfw880ChanHi = 0x5F600 - nicfw880Base
)

func init() {
	mk := func(id, vendor, model, name, what string) *genModel {
		return &genModel{ID: id, Vendor: vendor, Model: model, Variant: "nicFW880", RClass: "IU0VYT_nicFW880", Name: name,
			Slots: nicfw880Chans, First: 1, NameMax: 30, Charset: charsetASCII, Powers: nicfwPowers,
			Bands: []band{{nicfwMinHz, nicfwMaxHz}}, Narrow: true, NoChirp: true, DCS: dtcsCodes, Baud: 38400, Fam: nicfw880Fam{},
			Notes: what + " con il firmware libero nicFW880 / nicFW950: 999 canali con nomi fino a 30 caratteri. Cavo di programmazione della radio, " +
				"38400 baud, radio accesa. Dopo la scrittura la radio si riavvia. " + nicfwPowerNote + nicfwSafety}
	}
	genRegister(
		mk("rtl-rt880-nicfw", "Radtel", "RT-880", "Radtel RT-880 / RT-880G (firmware nicFW880)", "Radtel RT-880 e RT-880G (e iRadio UV-98)"),
		mk("rtl-rt950pro-nicfw", "Radtel", "RT-950 Pro", "Radtel RT-950 Pro (firmware nicFW950)", "Radtel RT-950 Pro"),
	)
}

func (nicfw880Fam) layout(m *genModel) *bwEl { return bwMust("#seekto 0xFEFF;\nu8 fine;") }

// i canali cominciano dopo i tre VFO
func (nicfw880Fam) at(i int) int { return (i + 3) * nicfw880Rec }

func (f nicfw880Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, img[f.at(i):f.at(i)+nicfw880Rec]...)
}

func (f nicfw880Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) == nicfw880Rec {
		copy(img[f.at(i):], raw)
	}
}

// Match: i file si riconoscono dai metadati
func (nicfw880Fam) Match(m *genModel, img []byte) bool { return false }

func (f nicfw880Fam) Get(m *genModel, img []byte, i int) *Chan {
	le := binary.LittleEndian
	r := img[f.at(i) : f.at(i)+nicfw880Rec]
	rx := int64(le.Uint32(r[0:])) * 10
	if le.Uint16(r[24:]) != nicfw880Magic || rx < nicfwMinHz || rx > nicfwMaxHz {
		return nil
	}
	c := &Chan{Rx: rx, Tx: int64(le.Uint32(r[4:])) * 10, Scan: true}
	if c.Tx < nicfwMinHz || c.Tx > nicfwMaxHz {
		c.Tx = rx
	}
	c.RxTone, c.TxTone = nicfwToneStr(le.Uint16(r[12:])), nicfwToneStr(le.Uint16(r[14:]))
	if p := int(r[21]); p == 0 {
		c.TxOff, c.Tx = true, rx
	} else {
		c.Power = nicfwSem(p, 255)
	}
	c.Narrow = r[19] == 1
	c.Name = nicfwName(r[32:62])
	return c
}

func (f nicfw880Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	le := binary.LittleEndian
	r := img[f.at(i) : f.at(i)+nicfw880Rec]
	if c == nil {
		for k := range r {
			r[k] = 0xFF
		}
		return nil
	}
	if c.Rx%10 != 0 || c.Rx < nicfwMinHz || c.Rx > nicfwMaxHz {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	if tx%10 != 0 || tx < nicfwMinHz || tx > nicfwMaxHz {
		return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(tx))
	}
	rt, err := nicfwTone(c.RxTone)
	if err != nil {
		return err
	}
	tt, err := nicfwTone(c.TxTone)
	if err != nil {
		return err
	}
	old := 0
	if le.Uint16(r[24:]) != nicfw880Magic {
		// canale nuovo: nessun gruppo, modulazione automatica (3)
		for k := range r {
			r[k] = 0
		}
		r[18] = 3
	} else {
		old = int(r[21])
	}
	le.PutUint32(r[0:], uint32(c.Rx/10))
	le.PutUint32(r[4:], uint32(tx/10))
	le.PutUint16(r[12:], rt)
	le.PutUint16(r[14:], tt)
	r[19] = 0
	if c.Narrow {
		r[19] = 1
	}
	if c.TxOff {
		r[21] = 0
	} else {
		r[21] = byte(nicfwPower(c.Power, old, 255))
	}
	if tx != c.Rx {
		r[20] = 0 // il blocco su canale occupato non vale con i ponti
	}
	le.PutUint16(r[24:], nicfw880Magic)
	name := m.cleanName(c.Name)
	for k := 0; k < 30; k++ {
		r[32+k] = 0
		if k < len(name) {
			r[32+k] = name[k]
		}
	}
	return nil
}

func nicfw880Read(s *rawSess, addr int) ([]byte, error) {
	a := []byte{byte(addr), byte(addr >> 8), byte(addr >> 16), byte(addr >> 24)}
	if err := s.send(append([]byte{0xAA, 0x30}, a...)); err != nil {
		return nil, err
	}
	b := readUpTo(s.ctx, s.p, 6+128+1, 2*time.Second)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return nil, errTimeout
	}
	if len(b) != 135 || b[0] != 0xAA || b[1] != 0x30 || !bytes.Equal(b[2:6], a) {
		return nil, rErr("blocco della radio non valido (0x%04X)", addr)
	}
	if nicfwSum(b[6:134]) != b[134] {
		return nil, rErr("somma di controllo errata nel blocco 0x%04X: ripeti la lettura", addr)
	}
	return b[6:134], nil
}

func nicfw880ReadAll(s *rawSess, prog Progress, what string) ([]byte, error) {
	img := make([]byte, 0, nicfw880Size)
	for a := 0; a < nicfw880Size; a += 128 {
		b, err := nicfw880Read(s, nicfw880Base+a)
		if err != nil {
			if a == 0 {
				return nil, genFail("Identificazione", err)
			}
			return nil, err
		}
		img = append(img, b...)
		prog(len(img), nicfw880Size, what)
	}
	return img, nil
}

func (f nicfw880Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, nicfw880Size, "Collegamento con la radio…")
	img, err := nicfw880ReadAll(s, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	prog(nicfw880Size, nicfw880Size, "Lettura completata")
	return img, "", nil
}

func (f nicfw880Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != nicfw880Size {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, nicfw880Size, "Collegamento con la radio…")
	now, err := nicfw880ReadAll(s, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		return "", err
	}
	if !bytes.Equal(now[nicfw880SetAt:nicfw880SetAt+2], img[nicfw880SetAt:nicfw880SetAt+2]) {
		return "", rErr("la radio ha un'altra versione del firmware nicFW rispetto all'immagine: leggi la radio e riporta i canali su quella lettura")
	}
	written := 0
	for a := 0; a < nicfw880ChanHi; a += 128 {
		if bytes.Equal(now[a:a+128], img[a:a+128]) {
			continue
		}
		addr := nicfw880Base + a
		pkt := []byte{0xAA, 0x40, byte(addr), byte(addr >> 8), byte(addr >> 16), byte(addr >> 24)}
		pkt = append(pkt, img[a:a+128]...)
		pkt = append(pkt, nicfwSum(img[a:a+128]))
		if err := s.send(pkt); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 2, 3*time.Second)
		if ctx.Err() != nil {
			return "", errCancelled
		}
		if len(b) != 2 || b[0] != 0xAA || b[1] != 0x40 {
			return "", rErr("la radio non ha confermato il blocco 0x%04X", addr)
		}
		written++
		prog(a+128, nicfw880Size, "Scrittura…")
	}
	if written > 0 {
		// conferma finale (la radio salva i dati) e riavvio
		if err := s.send([]byte{0xAA, 0x48}); err != nil {
			return "", err
		}
		b := readUpTo(ctx, s.p, 2, 10*time.Second)
		if ctx.Err() != nil {
			return "", errCancelled
		}
		if len(b) != 2 || b[0] != 0xAA || b[1] != 0x48 {
			return "", rErr("la radio non ha confermato la scrittura")
		}
		nicfwLast(s, 0xAA, 0x50)
	}
	prog(nicfw880Size, nicfw880Size, "Scrittura completata")
	return "", nil
}
