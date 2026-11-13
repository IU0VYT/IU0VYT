// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Retevis C6 (FRS / GMRS, 30 canali), Retevis C66 (UHF, 16 canali) e Radioddity
// Echo-DE / Echo-IT (PMR446, 16 canali) (4.27). Non sono nel CHIRP: protocollo e
// memoria ricavati leggendo il codice dei programmi dei costruttori «C6.exe»,
// «C66.exe» e «Radioddity ECHO.exe» (lo stesso programma, classi Form1, Form3 e
// Form4: cambiano la chiave dell'aggancio, il numero dei canali e gli indirizzi).
//
// Collegamento: 9600 baud, 8N1, DTR e RTS alti.
//
//   aggancio   il programma sceglie due numeri a caso, r1 (1-254) e r2 (0-254), e
//              manda 'S', poi gli 8 byte della chiave ciascuno in XOR con r1
//              spostato a sinistra della propria posizione (solo gli 8 bit bassi),
//              inserendo r1 dopo i primi (r2>>3)&7 byte, e in coda r2.
//              La radio risponde con 9 byte: v = ((r1<<2)&0xF0) | ((r2>>3)&0x0F),
//              poi la chiave in XOR con v spostato allo stesso modo.
//   password   6 byte (FF se non c'è) → 'A'
//   lettura    'R' indirizzo (2 byte, prima il basso) lunghezza → dati + somma
//   scrittura  'W' indirizzo (2 byte) + 16 dati + somma dei dati → 'A'
//   fine       'E'
//
// La prima lettura è sempre il blocco della gamma (gamma di frequenza in 4 byte,
// indice del modello, codice dell'area): l'indice del modello deve essere 0. Prima
// di scrivere i primi 6 byte di quel blocco devono essere uguali a quelli
// dell'immagine. Si legge da 0 fino al blocco della password compreso; si scrive
// da 0 fino al blocco della password escluso (la password non si scrive mai). Le
// tarature, oltre la password, non si leggono e non si scrivono.
//
// Memoria (C6; fra parentesi C66 ed Echo):
//   0x000  canali da 16 byte: 30 (16)
//   0x1E0  (0x100) codici del salto di frequenza, non gestiti
//   0x600  (0x300) impostazioni generali (settings_wlt.go)
//   0x607  (0x307) versione del firmware, 3 byte
//   0x610  (0x310) gamma, indice del modello, area
//   0x620  (0x320) password
//
// Canale: ricezione e trasmissione in 4 byte BCD (prima le cifre basse, decine di
// Hz; FF FF FF FF nella ricezione = vuoto), tono di ricezione e di trasmissione
// in 2 byte (FF FF nessuno; CTCSS in BCD, decimi di Hz; DCS: 0x8000 + codice in
// BCD, 0xC000 se invertito), byte 12: bit 0-1 blocco su canale occupato, bit 2
// banda stretta, bit 3 potenza alta, bit 4 in scansione, bit 5-7 salto di
// frequenza; byte 13: bit 0-3 scrambler, bit 4 compander; byte 14: salto di
// frequenza con scansione (con i valori 1, 2, 3 e 5 il programma non mostra i
// toni).
//
// Frequenze: il programma del C6 accetta alla tastiera solo 462-467 MHz, ma i suoi
// canali predefiniti (campo defaultData) arrivano a 467,725 MHz: qui si accetta
// 462-468 MHz. Il programma dell'Echo prevede quattro gamme (136-174, 200-260,
// 350-390, 400-520 MHz) scelte dalla gamma che la radio dichiara: l'Echo-DE /
// Echo-IT è PMR446, quindi qui vale solo 400-520 MHz. In più ogni frequenza
// deve stare nella gamma dichiarata dalla radio.

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type wltOpt struct {
	key     []byte // chiave dell'aggancio (8 byte)
	set     int    // indirizzo delle impostazioni generali
	newFlag byte   // byte 12 di un canale nuovo, come lo mette il programma
	prog    string
}

func (o *wltOpt) rng() int  { return o.set + 0x10 } // gamma, modello, area
func (o *wltOpt) pwd() int  { return o.set + 0x20 } // password
func (o *wltOpt) size() int { return o.set + 0x30 }

type wltFam struct{}

func wltO(m *genModel) *wltOpt { return m.Opt.(*wltOpt) }

func init() {
	note := func(o *wltOpt, extra string) string {
		return "Canali senza nome, 9600 baud. Protocollo e memoria ricavati dal programma del costruttore " + o.prog +
			" e provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, toni, potenza, banda e scansione dei canali " +
			"e le impostazioni generali; scrambler, compander, blocco su canale occupato e salto di frequenza restano quelli del canale. " + extra +
			"Se nella radio è impostata una password di lettura, toglila prima con il programma del costruttore."
	}
	pw := []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}}
	c6 := &wltOpt{key: []byte("XMU=E8\xFF\xFF"), set: 0x600, newFlag: 0x18, prog: "«C6»"}
	c66 := &wltOpt{key: []byte("XMU=E88\xFF"), set: 0x300, newFlag: 0x18, prog: "«C66»"}
	echo := &wltOpt{key: []byte("XK=N23C\xFF"), set: 0x300, newFlag: 0x08, prog: "«Radioddity ECHO»"}
	genRegister(
		&genModel{ID: "rt-c6", Vendor: "Retevis", Model: "C6", RClass: "IU0VYT_Retevis_C6", Name: "Retevis C6",
			Slots: 30, First: 1, Powers: pw, Bands: []band{{462000000, 468000000}},
			Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 9600, Fam: wltFam{}, Opt: c6,
			Notes: note(c6, "Frequenze fra 462 e 468 MHz: i canali di fabbrica sono i 22 FRS / GMRS più 8 canali per i ponti GMRS (trasmissione fra 467,55 e 467,725 MHz). ")},
		&genModel{ID: "rt-c66", Vendor: "Retevis", Model: "C66", RClass: "IU0VYT_Retevis_C66", Name: "Retevis C66",
			Slots: 16, First: 1, Powers: pw, Bands: []band{{400000000, 480000000}},
			Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 9600, Fam: wltFam{}, Opt: c66,
			Notes: note(c66, "")},
		&genModel{ID: "rd-echo", Vendor: "Radioddity", Model: "Echo", RClass: "IU0VYT_Radioddity_Echo", Name: "Radioddity Echo-DE / Echo-IT",
			Slots: 16, First: 1, Powers: pw, Bands: []band{{400000000, 520000000}},
			Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 9600, Fam: wltFam{}, Opt: echo,
			Notes: note(echo, "Frequenze fra 400 e 520 MHz, entro la gamma che la radio dichiara: "+
				"ricorda che in PMR446 sono ammessi solo i 16 canali fra 446,00625 e 446,19375 MHz. ")},
	)
}

func (wltFam) rec(img []byte, i int) []byte { return img[i*16 : i*16+16] }

func (f wltFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f wltFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(f.rec(img, i), raw)
}

func (wltFam) Match(m *genModel, img []byte) bool { return false }

func wltToneDec(b []byte) string {
	v := int(b[0]) | int(b[1])<<8
	if v == 0xFFFF {
		return ""
	}
	if v&0x8000 != 0 {
		code, ok := bcdLE([]byte{b[0], b[1] & 0x0F})
		if !ok {
			return ""
		}
		return genTone{Mode: "DTCS", Code: int(code), Inv: v&0x4000 != 0}.String()
	}
	hz10, ok := bcdLE(b[:2])
	if !ok || hz10 == 0 {
		return ""
	}
	return genTone{Mode: "Tone", Hz10: int(hz10)}.String()
}

func wltToneEnc(dst []byte, s string) error {
	t, err := genToneParse(s)
	if err != nil {
		return err
	}
	switch t.Mode {
	case "Tone":
		if t.Hz10 <= 0 || t.Hz10 > 7999 {
			return fmt.Errorf("tono %s non previsto dalla radio", s)
		}
		putBCDLE(dst[:2], int64(t.Hz10))
	case "DTCS":
		if t.Code < 0 || t.Code > 777 {
			return fmt.Errorf("codice DCS %s non previsto dalla radio", s)
		}
		putBCDLE(dst[:2], int64(t.Code))
		dst[1] |= 0x80
		if t.Inv {
			dst[1] |= 0x40
		}
	default:
		dst[0], dst[1] = 0xFF, 0xFF
	}
	return nil
}

// wltHidesTones: con questi valori del byte 14 (salto di frequenza con scansione)
// il programma mostra «None» al posto dei toni
func wltHidesTones(b byte) bool { return b == 1 || b == 2 || b == 3 || b == 5 }

func (f wltFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if bytes.Equal(r[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	rx, ok := bcdLE(r[0:4])
	if !ok || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	if tx, ok := bcdLE(r[4:8]); ok && tx != 0 {
		c.Tx = tx * 10
	}
	if !wltHidesTones(r[14]) {
		c.RxTone, c.TxTone = wltToneDec(r[8:10]), wltToneDec(r[10:12])
	}
	c.Power = m.powerSem(int(r[12] >> 3 & 1))
	c.Narrow = r[12]&0x04 != 0
	c.Scan = r[12]&0x10 != 0
	return c
}

// wltRange: gamma di frequenza dichiarata dalla radio (decimi di MHz in BCD), se leggibile
func wltRange(m *genModel, img []byte) (lo, hi int64, ok bool) {
	a := wltO(m).rng()
	if len(img) < a+4 {
		return 0, 0, false
	}
	l, ok1 := bcdLE(img[a : a+2])
	h, ok2 := bcdLE(img[a+2 : a+4])
	if !ok1 || !ok2 || l == 0 || h <= l {
		return 0, 0, false
	}
	return l * 100000, h * 100000, true
}

func (f wltFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if c == nil {
		for k := 0; k < 14; k++ { // come il programma: i byte 14 e 15 restano
			r[k] = 0xFF
		}
		return nil
	}
	for _, hz := range []int64{c.Rx, c.Tx} {
		if hz%10 != 0 || !m.inBands(hz) {
			return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
		if lo, hi, ok := wltRange(m, img); ok && (hz < lo || hz > hi) {
			return fmt.Errorf("frequenza %s MHz fuori dalla gamma dichiarata dalla radio (%s-%s MHz)", fmtMHz(hz), fmtMHz(lo), fmtMHz(hi))
		}
	}
	var rt, tt [2]byte
	if err := wltToneEnc(rt[:], c.RxTone); err != nil {
		return err
	}
	if err := wltToneEnc(tt[:], c.TxTone); err != nil {
		return err
	}
	if bytes.Equal(r[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) { // canale nuovo: valori del programma
		r[12], r[13], r[14] = wltO(m).newFlag, 0, 0
	}
	putBCDLE(r[0:4], c.Rx/10)
	putBCDLE(r[4:8], c.Tx/10)
	if !wltHidesTones(r[14]) {
		copy(r[8:10], rt[:])
		copy(r[10:12], tt[:])
	} else if c.RxTone != "" || c.TxTone != "" {
		// il canale usa il salto di frequenza con scansione, che non ha toni: se ne
		// scegli uno il canale torna normale (i toni nascosti non vengono persi
		// finché non lo fai)
		copy(r[8:10], rt[:])
		copy(r[10:12], tt[:])
		r[14] = 0
	}
	r[12] &^= 0x1C
	if m.powerIndex(c.Power) == 1 {
		r[12] |= 0x08
	}
	if c.Narrow {
		r[12] |= 0x04
	}
	if c.Scan {
		r[12] |= 0x10
	}
	return nil
}

// fit: frequenze nella banda del modello e a passi di 10 Hz (la radio non ne
// memorizza di più fini)
func (wltFam) fit(m *genModel, c *Chan) (bool, string) {
	for _, hz := range []int64{c.Rx, c.Tx} {
		if hz%10 != 0 || !m.inBands(hz) {
			return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
	}
	return true, ""
}

// ---------------------------------------------------------------- protocollo

// wltWait: attesa della risposta all'aggancio (il programma aspetta un secondo); le
// altre risposte hanno il doppio del tempo. Le prove la accorciano.
var wltWait = time.Second

// wltRand: i due numeri casuali dell'aggancio (sostituibile nelle prove)
var wltRand = func() (r1, r2 byte) { return byte(1 + rand.Intn(254)), byte(rand.Intn(255)) }

func wltSum(b []byte) (s byte) {
	for _, x := range b {
		s += x
	}
	return s
}

// wltHello: pacchetto dell'aggancio e risposta attesa
func wltHello(key []byte, r1, r2 byte) (pkt, want []byte) {
	k := int(r2>>3) & 7
	pkt = append(pkt, 'S')
	for i, c := range key {
		if i == k {
			pkt = append(pkt, r1)
		}
		pkt = append(pkt, c^byte(int(r1)<<uint(i)))
	}
	pkt = append(pkt, r2)
	v := (r1<<2)&0xF0 | (r2>>3)&0x0F
	want = append(want, v)
	for i, c := range key {
		want = append(want, c^byte(int(v)<<uint(i)))
	}
	return pkt, want
}

type wltSess struct {
	ctx context.Context
	p   Port
	o   *wltOpt
}

func (s *wltSess) ack(what string) error {
	b := make([]byte, 1)
	if err := s.p.ReadFull(s.ctx, b, 2*wltWait); err != nil {
		return err
	}
	if b[0] != 'A' {
		return rErr("%s: risposta inattesa della radio (%02X)", what, b[0])
	}
	return nil
}

func (s *wltSess) read(addr, n int) ([]byte, error) {
	if err := s.p.Write([]byte{'R', byte(addr), byte(addr >> 8), byte(n)}); err != nil {
		return nil, err
	}
	b := make([]byte, n+1)
	if err := s.p.ReadFull(s.ctx, b, 2*wltWait); err != nil {
		return nil, err
	}
	if wltSum(b[:n]) != b[n] {
		return nil, rErr("somma di controllo errata nel blocco %#04x", addr)
	}
	return b[:n], nil
}

// enter: aggancio (tre tentativi come il programma), password e blocco della gamma
func (s *wltSess) enter(pwd []byte) ([]byte, error) {
	var err error
	for try := 0; try < 3; try++ {
		_ = s.p.Purge()
		r1, r2 := wltRand()
		pkt, want := wltHello(s.o.key, r1, r2)
		if err = s.p.Write(pkt); err != nil {
			return nil, err
		}
		got := make([]byte, len(want))
		if err = s.p.ReadFull(s.ctx, got, wltWait); err != nil {
			if err != errTimeout {
				return nil, err
			}
			continue
		}
		if !bytes.Equal(got, want) {
			err = rErr("la radio ha risposto all'aggancio in modo diverso da quello atteso: non è il modello di questa scheda")
			continue
		}
		err = nil
		break
	}
	if err != nil {
		return nil, err
	}
	if err := s.p.Write(pwd); err != nil {
		return nil, err
	}
	if err := s.ack("password"); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non ha accettato la password: se ne è impostata una, toglila con il programma del costruttore")
		}
		return nil, err
	}
	rg, err := s.read(s.o.rng(), 16)
	if err != nil {
		return nil, err
	}
	if rg[4] != 0 {
		return nil, rErr("la radio dichiara un altro modello (indice %d): non è quella di questa scheda", rg[4])
	}
	return rg, nil
}

func wltOpen(ctx context.Context, m *genModel, port string) (*wltSess, error) {
	p, err := openPort(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &wltSess{ctx: ctx, p: p, o: wltO(m)}, nil
}

func wltInfo(o *wltOpt, img []byte) string {
	v := img[o.set+7 : o.set+10]
	if v[0] == 0xFF || v[1] == 0xFF || v[2] == 0xFF {
		return "versione del firmware non indicata"
	}
	return fmt.Sprintf("firmware v%d.%d.%d", v[0], v[1], v[2])
}

func (f wltFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := wltOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	o := s.o
	total := o.size() / 16
	prog(0, total, "Collegamento con la radio…")
	if _, err := s.enter(bytes.Repeat([]byte{0xFF}, 6)); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := make([]byte, o.size())
	for a := 0; a < o.size(); a += 16 {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(a/16, total, "Lettura…")
		d, err := s.read(a, 16)
		if err != nil {
			return nil, "", genFail("Lettura", err)
		}
		copy(img[a:], d)
	}
	_ = s.p.Write([]byte{'E'})
	prog(total, total, "Lettura completata")
	return img, wltInfo(o, img), nil
}

func (f wltFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := wltO(m)
	if len(img) < o.size() {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.size())
	}
	if _, _, ok := wltRange(m, img); !ok {
		return "", fmt.Errorf("immagine non adatta a questa radio: leggi prima la radio, modifica i canali e scrivili nella stessa radio")
	}
	s, err := wltOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	total := o.pwd() / 16
	prog(0, total, "Collegamento con la radio…")
	rg, err := s.enter(img[o.pwd() : o.pwd()+6])
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	if !bytes.Equal(rg[:6], img[o.rng():o.rng()+6]) {
		return "", genFail("Identificazione", rErr("la gamma di frequenza, il modello o l'area della radio (% X) non sono quelli dell'immagine (% X): "+
			"l'immagine viene da un'altra radio", rg[:6], img[o.rng():o.rng()+6]))
	}
	for a := 0; a < o.pwd(); a += 16 {
		prog(a/16, total, "Scrittura…")
		pkt := append([]byte{'W', byte(a), byte(a >> 8)}, img[a:a+16]...)
		pkt = append(pkt, wltSum(img[a:a+16]))
		if err := s.p.Write(pkt); err != nil {
			return "", genFail("Scrittura", err)
		}
		if err := s.ack(fmt.Sprintf("blocco %#04x", a)); err != nil {
			return "", genFail("Scrittura", err)
		}
	}
	_ = s.p.Write([]byte{'E'})
	prog(total, total, "Scrittura completata")
	return wltInfo(o, img), nil
}

// chirpInvalid: modifiche di prova che il modello non può memorizzare (fuori banda)
func (f wltFam) chirpInvalid(m *genModel, c *Chan) bool {
	ok, _ := f.fit(m, c)
	return !ok
}
