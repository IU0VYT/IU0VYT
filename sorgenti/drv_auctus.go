// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Retevis RB39 / RB39P (FRS, 22 canali) e RB639 / RB639P (PMR446, 16 canali)
// (4.26). Non sono nel CHIRP: protocollo e dati ricavati leggendo il codice dei
// programmi del costruttore «RB39 CPS» e «RB639» v1.0.7 e «Retevis Analog CPS»
// v1.1.2 (classe Auctus_V1._0.Data.WriFreq: lo stesso programma per le quattro
// radio; cambiano il paese atteso, «FCC» o «CE», e nelle versioni «P»
// l'interruttore del Bluetooth).
//
// 2400 baud 8N1. Pacchetti: AA, lunghezza totale, comando, sottocomando, dati,
// somma di controllo (somma a 8 bit dei byte dalla lunghezza all'ultimo dato),
// BB.
//   versione:        AA 06 01 01 08 BB → testo («AU147…» o «AU148…»)
//   inizio lettura:  AA 06 02 01 09 BB → testo che contiene il paese
//   inizio scrittura AA 06 02 00 08 BB → testo che contiene il paese
//   lettura canale n AA 07 03 01 n chk BB → 16 byte del canale
//   lettura sistema  AA 06 04 01 0B BB → 9 byte (10 nelle versioni «P»)
//   scrittura canale AA 16 03 00 + 16 byte + chk BB → risposta
//   scrittura sistema AA 0F 04 00 + 9 byte + chk BB (AA 10 … con 10 byte)
//   fine:            AA 06 07 01 0E BB (lettura), AA 06 07 00 0D BB (scrittura)
// Canale (16 byte): numero, TX e RX (4 byte ciascuna, prima il più
// significativo, frequenza in decine di Hz), tono RX, tono TX (CTCSS: posizione
// nella tabella, 0 = nessuno; DCS: posizione + 1), tipo dei toni (bit 0: RX è
// DCS, bit 1: TX è DCS), potenza (0 bassa, 1 alta), banda (0 stretta, 1 larga),
// blocco su canale occupato, canale attivo.
//
// Immagine (formato di questo programma): 22 canali di 16 byte, poi 16 byte con
// i dati di sistema (primo byte: quanti sono) e due testi di 32 byte (versione e
// risposta all'inizio della lettura). I dati di sistema si riscrivono come
// sono nell'immagine: li modifica solo la scheda delle impostazioni
// (settings_auctus.go).

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	auctusSlots  = 22
	auctusSys    = auctusSlots * 16 // lunghezza e dati di sistema
	auctusVer    = auctusSys + 16   // testo della versione
	auctusRegion = auctusVer + 32   // testo della risposta all'inizio della lettura
	auctusSize   = auctusRegion + 32
)

// toni CTCSS nell'ordine del programma (posizione 0 = nessun tono)
var auctusCTCSS = []string{"", "67.0", "69.3", "71.9", "74.4", "77.0", "79.7", "82.5", "85.4", "88.5", "91.5", "94.8", "97.4", "100.0",
	"103.5", "107.2", "110.9", "114.8", "118.8", "123.0", "127.3", "131.8", "136.5", "141.3", "146.2", "151.4", "156.7", "159.8", "162.2",
	"165.5", "167.9", "171.3", "173.8", "177.3", "179.9", "183.5", "186.2", "189.9", "192.8", "196.6", "199.5", "203.5", "206.5", "210.7",
	"218.1", "225.7", "229.1", "233.6", "241.8", "250.3", "254.1"}

// codici DCS nell'ordine del programma (prima i 104 normali, poi gli invertiti)
var auctusDCS = []int{23, 25, 26, 31, 32, 36, 43, 47, 51, 53, 54, 65, 71, 72, 73, 74, 114, 115, 116, 122, 125, 131, 132, 134, 143, 145, 152,
	155, 156, 162, 165, 172, 174, 205, 212, 223, 225, 226, 243, 244, 245, 246, 251, 252, 255, 261, 263, 265, 266, 271, 274, 306, 311, 315,
	325, 331, 332, 343, 346, 351, 356, 364, 365, 371, 411, 412, 413, 423, 431, 432, 445, 446, 452, 454, 455, 462, 464, 465, 466, 503, 506,
	516, 523, 526, 532, 546, 565, 606, 612, 624, 627, 631, 632, 654, 662, 664, 703, 712, 723, 731, 732, 734, 743, 754}

type auctusOpt struct {
	country string // «FCC» o «CE»: deve comparire nella risposta della radio
}

type auctusFam struct{}

func (auctusFam) noIdent() {}

func auctusO(m *genModel) *auctusOpt { return m.Opt.(*auctusOpt) }

func init() {
	note := func(prog, other string) string {
		return "Canali senza nome, 400-470 MHz, 2400 baud. Protocollo e dati ricavati dal programma del costruttore " + prog +
			" e provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, toni, potenza e banda dei canali e le impostazioni generali " +
			"(tempo massimo di trasmissione, VOX, guida vocale, tasto laterale e le altre della scheda Impostazioni). La radio dice per quale paese è fatta: se non è quello di questa scheda la lettura si ferma e ti indica la scheda " +
			other + ". «Identifica» non riconosce questa radio: scegli tu la scheda giusta."
	}
	pw := []genPower{{"Bassa", 0.5, 2}, {"Alta", 2, 0}}
	genRegister(
		&genModel{ID: "rt-rb39", Vendor: "Retevis", Model: "RB39", RClass: "IU0VYT_Retevis_RB39", Name: "Retevis RB39 / RB39P (FRS)",
			Slots: 22, First: 1, Powers: pw, Bands: []band{{400000000, 470000000}},
			Narrow: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 2400, Fam: auctusFam{}, Opt: &auctusOpt{country: "FCC"},
			Notes: note("«RB39 CPS» v1.0.7 / «Retevis Analog CPS» v1.1.2", "«Retevis RB639 / RB639P (PMR446)»")},
		&genModel{ID: "rt-rb639", Vendor: "Retevis", Model: "RB639", RClass: "IU0VYT_Retevis_RB639", Name: "Retevis RB639 / RB639P (PMR446)",
			Slots: 16, First: 1, Powers: pw, Bands: []band{{400000000, 470000000}},
			Narrow: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 2400, Fam: auctusFam{}, Opt: &auctusOpt{country: "CE"},
			Notes: note("«RB639» v1.0.7 / «Retevis Analog CPS» v1.1.2", "«Retevis RB39 / RB39P (FRS)»")},
	)
}

func (auctusFam) rec(img []byte, i int) []byte { return img[i*16 : i*16+16] }

func (f auctusFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f auctusFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(f.rec(img, i), raw[:16])
	f.rec(img, i)[0] = byte(i) // il numero del canale resta quello della posizione
}

func (auctusFam) Match(m *genModel, img []byte) bool { return false }

func auctusToneDec(v byte, dcs bool) string {
	if !dcs {
		if int(v) < len(auctusCTCSS) {
			return auctusCTCSS[v]
		}
		return ""
	}
	k := int(v) - 1
	switch {
	case k < 0 || k >= 2*len(auctusDCS):
		return ""
	case k < len(auctusDCS):
		return fmt.Sprintf("D%03dN", auctusDCS[k])
	}
	return fmt.Sprintf("D%03dI", auctusDCS[k-len(auctusDCS)])
}

func auctusToneEnc(t string) (v byte, dcs bool, err error) {
	if t == "" {
		return 0, false, nil
	}
	for i, s := range auctusCTCSS {
		if i > 0 && s == t {
			return byte(i), false, nil
		}
	}
	if len(t) == 5 && t[0] == 'D' && (t[4] == 'N' || t[4] == 'I') {
		for i, c := range auctusDCS {
			if fmt.Sprintf("%03d", c) == t[1:4] {
				if t[4] == 'I' {
					i += len(auctusDCS)
				}
				return byte(i + 1), true, nil
			}
		}
	}
	return 0, false, fmt.Errorf("tono %s non disponibile su questa radio", t)
}

func (f auctusFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	rx, tx := int64(binary.BigEndian.Uint32(r[5:9])), int64(binary.BigEndian.Uint32(r[1:5]))
	if r[15] != 1 || rx == 0 || rx == 0xFFFFFFFF {
		return nil // canale spento o senza frequenza
	}
	c := &Chan{Rx: rx * 10, Tx: tx * 10}
	if tx == 0 || tx == 0xFFFFFFFF {
		c.Tx = c.Rx
	}
	c.RxTone = auctusToneDec(r[9], r[11]&1 != 0)
	c.TxTone = auctusToneDec(r[10], r[11]&2 != 0)
	c.Power = m.powerSem(int(r[12] & 1))
	c.Narrow = r[13] == 0
	return c
}

func (f auctusFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	r[0] = byte(i)
	if c == nil {
		r[15] = 0 // canale spento: gli altri dati restano
		return nil
	}
	rt, rd, err := auctusToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	tt, td, err := auctusToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(r[1:5], uint32(c.Tx/10))
	binary.BigEndian.PutUint32(r[5:9], uint32(c.Rx/10))
	r[9], r[10], r[11] = rt, tt, 0
	if rd {
		r[11] |= 1
	}
	if td {
		r[11] |= 2
	}
	r[12] = 0
	if m.powerIndex(c.Power) == 1 {
		r[12] = 1
	}
	r[13] = 1
	if c.Narrow {
		r[13] = 0
	}
	if r[14] > 1 {
		r[14] = 0 // blocco su canale occupato: resta quello del canale, se valido
	}
	r[15] = 1
	return nil
}

// ---------------------------------------------------------------- protocollo

type auctusSess struct {
	ctx context.Context
	p   Port
}

func auctusSum(b []byte) (s byte) {
	for _, x := range b {
		s += x
	}
	return s
}

// auctusPkt: pacchetto con comando, sottocomando e dati
func auctusPkt(cmd, sub byte, data []byte) []byte {
	p := append([]byte{0xAA, byte(len(data) + 6), cmd, sub}, data...)
	return append(p, auctusSum(p[1:]), 0xBB)
}

// recv: un pacchetto dalla radio; restituisce i dati (dopo comando e sottocomando)
func (s *auctusSess) recv() ([]byte, error) {
	h := make([]byte, 2)
	if err := s.p.ReadFull(s.ctx, h, 2*time.Second); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	if h[0] != 0xAA || h[1] < 6 {
		_ = s.p.Purge()
		return nil, rErr("risposta inattesa della radio (%s)", fmt.Sprintf("% X", h))
	}
	rest := make([]byte, int(h[1])-2)
	if err := s.p.ReadFull(s.ctx, rest, 2*time.Second); err != nil {
		if err == errTimeout {
			return nil, rErr("risposta inattesa")
		}
		return nil, err
	}
	n := len(rest)
	if rest[n-1] != 0xBB || auctusSum(append([]byte{h[1]}, rest[:n-2]...)) != rest[n-2] {
		return nil, rErr("somma di controllo errata nella risposta")
	}
	return rest[2 : n-2], nil
}

func (s *auctusSess) ask(pkt []byte) ([]byte, error) {
	if err := s.p.Write(pkt); err != nil {
		return nil, err
	}
	return s.recv()
}

// enter: versione e inizio della lettura o della scrittura; controlla il paese
func (s *auctusSess) enter(m *genModel, write bool) (ver, region string, err error) {
	v, err := s.ask(auctusPkt(0x01, 0x01, nil))
	if err != nil {
		return "", "", err
	}
	ver = strings.TrimRight(string(v), "\x00 ")
	sub := byte(1)
	if write {
		sub = 0
	}
	r, err := s.ask(auctusPkt(0x02, sub, nil))
	if err != nil {
		return ver, "", err
	}
	region = strings.TrimRight(string(r), "\x00 ")
	want := auctusO(m).country
	if !strings.Contains(region, want) {
		other := map[string]string{"FCC": "«Retevis RB639 / RB639P (PMR446)»", "CE": "«Retevis RB39 / RB39P (FRS)»"}[want]
		if strings.Contains(region, map[string]string{"FCC": "CE", "CE": "FCC"}[want]) {
			return ver, region, rErr("la radio è la versione per un altro paese (%q): apri la scheda %s", region, other)
		}
		return ver, region, rErr("la radio non ha indicato il paese atteso (%s): risposta %q", want, region)
	}
	return ver, region, nil
}

func auctusOpen(ctx context.Context, m *genModel, port string) (*auctusSess, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &auctusSess{ctx: ctx, p: p}, nil
}

func (f auctusFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := auctusOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	total := m.Slots + 1
	prog(0, total, "Collegamento con la radio…")
	ver, region, err := s.enter(m, false)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, auctusSize)
	for i := 0; i < auctusSlots; i++ {
		img[i*16] = byte(i) // posizioni oltre i canali della radio: spente
	}
	copy(img[auctusVer:auctusVer+32], ver)
	copy(img[auctusRegion:auctusRegion+32], region)
	for i := 0; i < m.Slots; i++ {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(i, total, "Lettura…")
		d, err := s.ask(auctusPkt(0x03, 0x01, []byte{byte(i)}))
		if err == nil && len(d) < 16 {
			err = rErr("dati letti incompleti (%d byte)", len(d))
		}
		if err != nil {
			return nil, "", genFail("Lettura", err)
		}
		copy(img[i*16:], d[:16])
	}
	prog(m.Slots, total, "Lettura…")
	d, err := s.ask(auctusPkt(0x04, 0x01, nil))
	if err == nil && (len(d) < 9 || len(d) > 15) {
		err = rErr("dati letti incompleti (%d byte)", len(d))
	}
	if err != nil {
		return nil, "", genFail("Lettura", err)
	}
	img[auctusSys] = byte(len(d))
	copy(img[auctusSys+1:], d)
	_ = s.p.Write(auctusPkt(0x07, 0x01, nil))
	prog(total, total, "Lettura completata")
	return img, fmt.Sprintf("identificativo %q", ver+" / "+region), nil
}

func (f auctusFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < auctusSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), auctusSize)
	}
	n := int(img[auctusSys])
	if n < 9 || n > 15 {
		return "", fmt.Errorf("immagine non adatta a questa radio: leggi prima la radio, modifica i canali e scrivili nella stessa radio")
	}
	s, err := auctusOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	total := m.Slots + 1
	prog(0, total, "Collegamento con la radio…")
	ver, region, err := s.enter(m, true)
	info := fmt.Sprintf("identificativo %q", ver+" / "+region)
	if err != nil {
		return info, genFail("Identificazione", err)
	}
	for i := 0; i < m.Slots; i++ {
		prog(i, total, "Scrittura…")
		d := append([]byte{}, img[i*16:i*16+16]...)
		d[0] = byte(i)
		if _, err := s.ask(auctusPkt(0x03, 0x00, d)); err != nil {
			return info, genFail("Scrittura", err)
		}
	}
	prog(m.Slots, total, "Scrittura…")
	if _, err := s.ask(auctusPkt(0x04, 0x00, img[auctusSys+1:auctusSys+1+n])); err != nil {
		return info, genFail("Scrittura", err)
	}
	_ = s.p.Write(auctusPkt(0x07, 0x00, nil))
	prog(total, total, "Scrittura completata")
	return info, nil
}
