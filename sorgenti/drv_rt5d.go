// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-5D (radio DMR): protocollo e memoria ricavati dal programma del
// costruttore «RT-5D CPS» versione 2.0.5 («JJCC-888DMR_CPS.exe», .NET, progetto
// «KDH»), pubblicato sulla pagina dei download di radtels.com.
//
// Protocollo (115200 baud, 8N1): pacchetti
//   A5 comando parametro(2, big endian) lunghezza(2, big endian) dati CRC(2)
// con il CRC16 XMODEM (big endian) calcolato da «comando» alla fine dei dati.
// La radio risponde con un pacchetto dello stesso tipo; il comando EE nella
// risposta è un rifiuto. Senza risposta in 1,2 secondi il programma rimanda il
// pacchetto, al massimo 3 volte.
// Sequenza del programma (il parametro è il numero del pacchetto nella sezione):
//   02 «PROGRAMJC8810DU»; 05 FF×6 (password); 46 con 128 byte a zero (versione);
//   lettura: per ogni sezione il comando con tanti byte a zero quanti ne ha un
//   pacchetto, la risposta porta i dati; scrittura: comando + dati;
//   01 con 2 byte a zero (fine).
// Sezioni, nell'ordine (comando di lettura / di scrittura, byte × pacchetti):
//   DTMF 16/36 272×1; chiavi di cifratura 15/35 264×1; contatti 13/33 800×80;
//   gruppi 14/34 1024×4; canali 10/30 1024×64; VFO 11/31 128×1; funzioni
//   opzionali 12/32 64×1; dati della radio 19/39 64×1 (il programma li legge
//   sempre e li scrive solo con un'opzione apposita: qui non si scrivono mai).
// L'immagine è la fila delle sezioni così come le manda la radio.
//
// Canale (64 byte, 16 zone da 64 canali; libero = primi 4 byte a FF o a zero):
//   0-3 RX e 4-7 TX in decine di Hz (little endian; TX a FF = solo ricezione)
//   8-9 tono RX, 10-11 tono TX: 0000 (o FF) nessuno; secondo byte a zero = DCS,
//       il primo è la posizione (da 1) in una tabella di 105 codici normali
//       seguiti dagli stessi 105 invertiti; altrimenti decimi di Hz (little endian)
//   12 codice DTMF, 13 PTT ID
//   14 e 15 tipo: 0,0 digitale; 1,0 analogico largo; 1,1 analogico stretto
//   16 potenza (0 alta, 1 media, 2 bassa), 17 scrambler, 18 cifratura, 19 blocco
//   su canale occupato, 20 scansione (1 = in scansione), 21 slot, 22 color code,
//   23 gruppo di ricezione, 25 chiave, 26 modo DMR, 27 FHSS, 28-31 codice FHSS,
//   32-43 nome (GB2312, FF in coda), 44-45 contatto; gli altri byte a FF.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type rt5dSec struct {
	name   string
	rd, wr byte // comandi di lettura e di scrittura (wr 0: non si scrive)
	size   int  // byte di un pacchetto
	n      int  // pacchetti
}

var rt5dSecs = []rt5dSec{
	{"DTMF", 0x16, 0x36, 0x110, 1}, {"cifratura", 0x15, 0x35, 0x108, 1}, {"contatti", 0x13, 0x33, 0x320, 80},
	{"gruppi", 0x14, 0x34, 0x400, 4}, {"canali", 0x10, 0x30, 0x400, 64}, {"VFO", 0x11, 0x31, 0x80, 1},
	{"funzioni", 0x12, 0x32, 0x40, 1}, {"dati della radio", 0x19, 0, 0x40, 1},
}

const (
	rt5dChSize = 64
	rt5dSlots  = 1024
	rt5dHello  = "PROGRAMJC8810DU"
)

func rt5dOff(name string) int {
	o := 0
	for _, s := range rt5dSecs {
		if s.name == name {
			return o
		}
		o += s.size * s.n
	}
	panic("sezione sconosciuta " + name)
}

func rt5dSize() int {
	n := 0
	for _, s := range rt5dSecs {
		n += s.size * s.n
	}
	return n
}

// codici DCS del programma: quelli di CHIRP più 645, in ordine
var rt5dDCS = withCodes(dtcsCodes, 645)

type rt5dOpt struct{}

type rt5dFam struct{}

func init() {
	genRegister(&genModel{ID: "rtl-rt5d", Vendor: "Radtel", Model: "RT-5D", RClass: "IU0VYT_Radtel_RT5D", Name: "Radtel RT-5D",
		Slots: rt5dSlots, First: 1, NameMax: 12, Charset: charsetASCII, Zones: 16,
		Powers: []genPower{{"Alta", 0, 0}, {"Media", 0, 1}, {"Bassa", 0, 2}},
		Bands:  []band{{18000000, 999999990}}, Narrow: true, HasScan: true, NoChirp: true,
		DCS: rt5dDCS, Baud: 115200, Fam: rt5dFam{}, Opt: &rt5dOpt{}, Digital: "DMR",
		Notes: "Radio DMR: 16 zone da 64 canali. Protocollo e memoria ricavati dal programma del costruttore «RT-5D CPS» (versione 2.0.5). " +
			"Qui si modificano frequenze, nome, potenza, banda, scansione e toni; le altre voci del canale restano quelle impostate " +
			"(un canale DMR nuovo prende i valori predefiniti del programma: primo contatto, slot 1, color code 1). " +
			"La lettura prende tutta la memoria e la scrittura la rimanda tutta com'è nel file, come fa il programma del costruttore: " +
			"contatti, gruppi, chiavi e impostazioni non si modificano qui, ma scrivendo un file vecchio tornano quelli di allora. " +
			"Nome e identificativo del modello non vengono mai scritti."})
}

func (rt5dFam) rec(img []byte, i int) []byte {
	o := rt5dOff("canali") + i*rt5dChSize
	return img[o : o+rt5dChSize]
}

// come il programma: libero se i primi quattro byte sono FF o i primi tre zero
func rt5dEmpty(r []byte) bool {
	return r[0] == 0xFF && r[1] == 0xFF && r[2] == 0xFF && r[3] == 0xFF || r[0] == 0 && r[1] == 0 && r[2] == 0
}

func rt5dToneDec(b []byte) string {
	switch {
	case b[0] == 0xFF, b[0] == 0 && b[1] == 0:
		return ""
	case b[1] == 0:
		k := int(b[0]) - 1
		n := len(rt5dDCS)
		if k >= 2*n {
			return ""
		}
		return genTone{Mode: "DTCS", Code: rt5dDCS[k%n], Inv: k >= n}.String()
	}
	return fmtTone(int(binary.LittleEndian.Uint16(b)))
}

func rt5dToneEnc(t genTone) ([]byte, error) {
	switch t.Mode {
	case "Tone":
		return []byte{byte(t.Hz10), byte(t.Hz10 >> 8)}, nil
	case "DTCS":
		k := idxInt(rt5dDCS, t.Code)
		if k < 0 {
			return nil, fmt.Errorf("codice DCS %03d non ammesso da questa radio", t.Code)
		}
		if t.Inv {
			k += len(rt5dDCS)
		}
		return []byte{byte(k + 1), 0}, nil
	}
	return []byte{0, 0}, nil
}

func (f rt5dFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if rt5dEmpty(r) {
		return nil
	}
	c := &Chan{Rx: int64(binary.LittleEndian.Uint32(r[0:])) * 10}
	if rt5dEmpty(r[4:8]) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = int64(binary.LittleEndian.Uint32(r[4:])) * 10
	}
	c.Name = strings.TrimRight(string(trimBytes(r[32:44], 0xFF, 0x00)), " ")
	c.Power = m.powerSem(int(r[16]&0x0F) % 3)
	c.Scan = r[20]&1 == 1
	// tipo come lo legge il programma: (byte 14 & 1) + (byte 15 & 1)
	switch r[14]&1 + r[15]&1 {
	case 0:
		c.Dig = true
		return c
	case 2:
		c.Narrow = true
	}
	c.RxTone = rt5dToneDec(r[8:10])
	c.TxTone = rt5dToneDec(r[10:12])
	return c
}

func (f rt5dFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f rt5dFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (f rt5dFam) Match(m *genModel, img []byte) bool { return false }

func (f rt5dFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if c == nil {
		copy(r, bytes.Repeat([]byte{0xFF}, rt5dChSize))
		return nil
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tb, err := rt5dToneEnc(tt)
	if err != nil {
		return err
	}
	rb, err := rt5dToneEnc(rt)
	if err != nil {
		return err
	}
	if rt5dEmpty(r) {
		// canale nuovo come lo scrive il programma: tutto a FF, poi le voci con i
		// valori predefiniti (color code 1, il resto a zero); codice FHSS vuoto
		copy(r, bytes.Repeat([]byte{0xFF}, rt5dChSize))
		for _, k := range []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 23, 25, 26, 27, 44, 45} {
			r[k] = 0
		}
		r[22] = 1
	}
	binary.LittleEndian.PutUint32(r[0:], uint32(c.Rx/10))
	if c.TxOff {
		copy(r[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		binary.LittleEndian.PutUint32(r[4:], uint32(c.Tx/10))
	}
	if c.Dig {
		r[14], r[15] = 0, 0
	} else {
		r[14], r[15] = 1, 0
		if c.Narrow {
			r[15] = 1
		}
		copy(r[8:10], rb)
		copy(r[10:12], tb)
	}
	r[16] = byte(m.powerIndex(c.Power))
	r[20] = 0
	if c.Scan {
		r[20] = 1
	}
	name := bytes.Repeat([]byte{0xFF}, 12)
	copy(name, c.Name)
	copy(r[32:44], name)
	return nil
}

// ---------------------------------------------------------------- protocollo

func rt5dFrame(cmd byte, par int, data []byte) []byte {
	f := append([]byte{0xA5, cmd, byte(par >> 8), byte(par), byte(len(data) >> 8), byte(len(data))}, data...)
	c := crc16XModem(f[1:])
	return append(f, byte(c>>8), byte(c))
}

type rt5dLink struct {
	ctx context.Context
	p   Port
}

// una risposta: si scartano i byte fino ad A5 (come il programma), poi
// intestazione, dati e CRC
func (l *rt5dLink) reply(to time.Duration) (byte, []byte, error) {
	end := time.Now().Add(to)
	left := func() time.Duration {
		if d := time.Until(end); d > 10*time.Millisecond {
			return d
		}
		return 10 * time.Millisecond
	}
	b := make([]byte, 1)
	for {
		if err := l.p.ReadFull(l.ctx, b, left()); err != nil {
			return 0, nil, err
		}
		if b[0] == 0xA5 {
			break
		}
		if time.Now().After(end) {
			return 0, nil, errTimeout
		}
	}
	h := make([]byte, 5)
	if err := l.p.ReadFull(l.ctx, h, left()); err != nil {
		return 0, nil, err
	}
	n := int(h[3])<<8 | int(h[4])
	if n > 0x800 {
		return 0, nil, errTimeout // pacchetto non valido: si rimanda la richiesta
	}
	d := make([]byte, n+2)
	if err := l.p.ReadFull(l.ctx, d, left()+time.Duration(n)*time.Millisecond); err != nil {
		return 0, nil, err
	}
	if crc16XModem(append(h, d[:n]...)) != uint16(d[n])<<8|uint16(d[n+1]) {
		return 0, nil, errTimeout // CRC errato: il programma ignora il pacchetto e rimanda la richiesta
	}
	return h[0], d[:n], nil
}

// invio di un pacchetto e risposta, con i nuovi invii del programma
func (l *rt5dLink) xfer(cmd byte, par int, data []byte) ([]byte, error) {
	f := rt5dFrame(cmd, par, data)
	var last error
	for try := 0; try < 4; try++ {
		if l.ctx.Err() != nil {
			return nil, errCancelled
		}
		if try > 0 {
			_ = l.p.Purge()
		}
		if err := l.p.Write(f); err != nil {
			return nil, err
		}
		rc, d, err := l.reply(1200 * time.Millisecond)
		if err == nil {
			if rc == 0xEE {
				return nil, rErr("la radio ha rifiutato il comando %02X", cmd)
			}
			return d, nil
		}
		if err != errTimeout {
			return nil, err
		}
		last = err
	}
	return nil, last
}

func rt5dOpen(ctx context.Context, port string) (*rt5dLink, error) {
	p, err := openPort(port, 115200)
	if err != nil {
		return nil, err
	}
	return &rt5dLink{ctx: ctx, p: p}, nil
}

// ingresso in programmazione: saluto, password, versione
func (l *rt5dLink) enter() error {
	_ = l.p.Purge()
	if _, err := l.xfer(0x02, 0, []byte(rt5dHello)); err != nil {
		return err
	}
	if _, err := l.xfer(0x05, 0, bytes.Repeat([]byte{0xFF}, 6)); err != nil {
		return err
	}
	_, err := l.xfer(0x46, 0, make([]byte, 0x80))
	return err
}

func (l *rt5dLink) leave() { _, _ = l.xfer(0x01, 0, make([]byte, 2)) }

func (f rt5dFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	l, err := rt5dOpen(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer l.p.Close()
	size := rt5dSize()
	prog(0, size, "Collegamento con la radio…")
	if err := l.enter(); err != nil {
		return nil, "", genFail("Collegamento", err)
	}
	img := make([]byte, 0, size)
	for _, s := range rt5dSecs {
		for k := 0; k < s.n; k++ {
			prog(len(img), size, "Lettura: "+s.name+"…")
			d, err := l.xfer(s.rd, k, make([]byte, s.size))
			if err != nil {
				return nil, "", genFail("Lettura", err)
			}
			if len(d) < s.size {
				return nil, "", genFail("Lettura", rErr("risposta della radio troppo corta (%d byte invece di %d)", len(d), s.size))
			}
			img = append(img, d[:s.size]...)
		}
	}
	l.leave()
	prog(size, size, "Lettura completata")
	return img, "", nil
}

func (f rt5dFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	size := rt5dSize()
	if len(img) < size {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	l, err := rt5dOpen(ctx, port)
	if err != nil {
		return "", err
	}
	defer l.p.Close()
	prog(0, size, "Collegamento con la radio…")
	if err := l.enter(); err != nil {
		return "", genFail("Collegamento", err)
	}
	off := 0
	for _, s := range rt5dSecs {
		for k := 0; k < s.n; k++ {
			if s.wr != 0 {
				prog(off, size, "Scrittura: "+s.name+"…")
				if _, err := l.xfer(s.wr, k, img[off:off+s.size]); err != nil {
					return "", genFail("Scrittura", err)
				}
			}
			off += s.size
		}
	}
	l.leave()
	prog(size, size, "Scrittura completata")
	return "", nil
}
