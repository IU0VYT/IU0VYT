// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-480 e RT-68: protocollo e memoria ricavati dal programma del
// costruttore «Radio Programing Software» (versioni 1.5.2 per la RT-480 e 1.5.3
// per la RT-68, .NET, progetto «SQ5R»), pubblicato sulla pagina dei download di
// radtels.com. Le due versioni hanno lo stesso identico codice.
//
// Protocollo (9600 baud, 8N1), della stessa famiglia della Radioddity GA-510
// (drv_ga510.go) ma con un'altra mappa della memoria:
//   «PROGROM» + «S19» + 'U' → 0x06; 'F' → 8 byte (il programma non li controlla);
//   lettura: 'R' indirizzo(2, big endian) 0x40 → 4 byte di intestazione (il
//   programma controlla che i byte 1 e 2 ripetano l'indirizzo) + 64 byte;
//   scrittura: 'W' indirizzo(2) 0x40 + 64 byte → 0x06;
//   fine: 'E' (il programma lo manda dopo la lettura).
// Senza risposta il programma aspetta fino a 6 secondi (un secondo, più cinque
// nuove attese); l'aggancio qui si ripete fino a 5 volte.
// Il programma legge 0x0000-0x1FFF (canali e nomi), 0x3A00-0x3A7F (impostazioni
// e VFO) e 0x3B00-0x3C3F (DTMF); l'immagine è la fila di queste tre aree. In
// scrittura il programma ricostruisce impostazioni, VFO e DTMF dalle sue
// finestre: qui si scrivono solo canali e nomi (0x0000-0x1FFF), con gli stessi
// pacchetti, e il resto della radio non si tocca.
//
// Canale (16 byte da 0x0000, 256 canali numerati da 0 a 255; libero = primo
// byte a FF):
//   0-3 RX e 4-7 TX in BCD little endian, decine di Hz (il programma accetta solo
//       multipli di 1,25 kHz e vuole sempre la frequenza di trasmissione)
//   8-9 tono RX, 10-11 tono TX: 0000 nessuno; secondo byte a zero = DCS, il primo
//       è la posizione (da 1) fra 105 codici normali seguiti dagli stessi 105
//       invertiti; altrimenti decimi di Hz (little endian)
//   12 codice di segnalazione (da zero), 13 PTT ID (0 no, 1 inizio, 2 fine, 3 entrambi)
//   14 potenza nei 4 bit bassi (0 alta, 1 bassa), scrambler nei 4 bit alti (0-9)
//   15 bit 0x40 banda stretta, 0x08 blocco su canale occupato, 0x04 in scansione
// Nome (16 byte da 0x1000): 12 byte GB2312, FF in coda.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const (
	rt480Slots   = 256
	rt480NameOff = 0x1000
	rt480ChEnd   = 0x2000 // fine di canali e nomi: l'area che si scrive
	rt480Hello   = "PROGROMS19U"
	rt480Timeout = 6 * time.Second
)

// aree lette dal programma, nell'ordine (inizio, fine esclusa)
var rt480Areas = [][2]int{{0x0000, 0x2000}, {0x3A00, 0x3A80}, {0x3B00, 0x3C40}}

func rt480Size() int {
	n := 0
	for _, a := range rt480Areas {
		n += a[1] - a[0]
	}
	return n
}

type rt480Opt struct{}

type rt480Fam struct{}

func init() {
	for _, x := range []struct {
		id, model, rclass string
		pw                []genPower
	}{
		{"rtl-rt480", "RT-480", "IU0VYT_Radtel_RT480", []genPower{{"Alta 5 W", 5, 0}, {"Bassa 1 W", 1, 2}}},
		{"rtl-rt68", "RT-68", "IU0VYT_Radtel_RT68", []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}}},
	} {
		genRegister(&genModel{ID: x.id, Vendor: "Radtel", Model: x.model, RClass: x.rclass, Name: "Radtel " + x.model,
			Slots: rt480Slots, First: 0, NameMax: 12, Charset: charsetASCII, Powers: x.pw,
			Bands:  []band{{134000000, 175999990}, {200000000, 259999990}, {300000000, 520000000}},
			Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true, DCS: uv5rDTCS, Baud: 9600, Fam: rt480Fam{}, Opt: &rt480Opt{},
			Notes: "Protocollo e memoria ricavati dal programma del costruttore «Radio Programing Software» (versioni 1.5.2 e 1.5.3). " +
				"Accendi la radio con il cavo collegato. Canali da 0 a 255, come nel programma del costruttore; frequenze a passi di 1,25 kHz. " +
				"Qui si modificano frequenze, nome, potenza, banda, scansione e toni; codice di segnalazione, PTT ID, scrambler e blocco " +
				"su canale occupato restano quelli impostati. Si scrivono solo canali e nomi: impostazioni, VFO e DTMF della radio non si toccano. " +
				"«Identifica» non riconosce questa radio: scegli tu la scheda giusta."})
	}
}

func (rt480Fam) rec(img []byte, i int) []byte { return img[i*16 : i*16+16] }
func (rt480Fam) name(img []byte, i int) []byte {
	return img[rt480NameOff+i*16 : rt480NameOff+i*16+16]
}

// frequenza come la legge il programma: vuota se l'ultimo byte è FF o zero
func rt480Freq(b []byte) (int64, bool) {
	if b[3] == 0xFF || b[3] == 0 {
		return 0, false
	}
	v, ok := bcdLE(b[:4])
	return v * 10, ok && v > 0
}

func (f rt480Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if r[0] == 0xFF {
		return nil
	}
	rx, ok := rt480Freq(r[0:4])
	if !ok {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if tx, ok := rt480Freq(r[4:8]); ok {
		c.Tx = tx
	} else {
		c.TxOff = true
	}
	c.Name = uvk5cnDecodeName(f.name(img, i)[:12])
	c.RxTone = uv5rToneDec(int64(r[8])|int64(r[9])<<8, true)
	c.TxTone = uv5rToneDec(int64(r[10])|int64(r[11])<<8, false)
	// il programma del costruttore conosce solo alta (0) e bassa: ogni altro
	// valore lo riscrive come bassa
	pw := 0
	if r[14]&0x0F != 0 {
		pw = 1
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r[15]&0x40 != 0
	c.Scan = r[15]&0x04 != 0
	return c
}

func (f rt480Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append(append([]byte{}, f.rec(img, i)...), f.name(img, i)...)
}

func (f rt480Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(f.rec(img, i), raw[:16])
	if len(raw) >= 32 {
		copy(f.name(img, i), raw[16:32])
	}
}

func (f rt480Fam) Match(m *genModel, img []byte) bool { return false }

// rt480Name: nome di 12 byte GB2312 con FF in coda (senza spezzare un carattere
// di due byte), in un campo di 16 byte a FF
func rt480Name(s string) []byte {
	out := bytes.Repeat([]byte{0xFF}, 16)
	n := 0
	for _, r := range s {
		e := uvk5cnEncodeChar(r)
		if n+len(e) > 12 {
			break
		}
		n += copy(out[n:], e)
	}
	return out
}

func (f rt480Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, nm := f.rec(img, i), f.name(img, i)
	if c == nil {
		copy(r, bytes.Repeat([]byte{0xFF}, 16))
		copy(nm, bytes.Repeat([]byte{0xFF}, 16))
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("questa radio non permette di disabilitare la trasmissione sul canale")
	}
	for _, hz := range []int64{c.Rx, c.Tx} {
		// limiti del programma: 134-176, 200-260 e 300-520 MHz
		if !m.inBands(hz) {
			return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
		if hz%1250 != 0 {
			return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 1,25 kHz)", fmtMHz(hz))
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
		if t.Mode == "DTCS" && idxInt(uv5rDTCS, t.Code) < 0 {
			return fmt.Errorf("codice DCS %s non disponibile su questa radio", fmt.Sprintf("%03d", t.Code))
		}
	}
	old := f.Get(m, img, i)
	if old == nil {
		// canale nuovo come lo crea il programma: segnalazione 1, PTT ID spento,
		// scrambler spento, blocco su canale occupato spento
		copy(r, make([]byte, 16))
		copy(nm, bytes.Repeat([]byte{0xFF}, 16))
	}
	putBCDLE(r[0:4], c.Rx/10)
	putBCDLE(r[4:8], c.Tx/10)
	rv, tv := uv5rToneEnc(rt), uv5rToneEnc(tt)
	r[8], r[9], r[10], r[11] = byte(rv), byte(rv>>8), byte(tv), byte(tv>>8)
	r[14] = r[14]&0xF0 | byte(m.powerIndex(c.Power))
	// byte 15 come lo ricostruisce il programma: solo banda, blocco e scansione
	b := r[15] & 0x08
	if c.Narrow {
		b |= 0x40
	}
	if c.Scan {
		b |= 0x04
	}
	r[15] = b
	// il nome non cambiato resta com'è (anche con caratteri che qui non si scrivono)
	if old == nil || c.Name != old.Name {
		copy(nm, rt480Name(c.Name))
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

// rt480Enter: aggancio del programma (HandShake)
func (s *rawSess) rt480Enter() ([]byte, error) {
	ok, any := false, false
	for k := 0; k < 5 && !ok; k++ {
		if err := s.send([]byte(rt480Hello)); err != nil {
			return nil, err
		}
		// il programma legge un byte per volta e aspetta lo 0x06
		end := time.Now().Add(time.Second)
		for !ok && time.Now().Before(end) {
			b := readUpTo(s.ctx, s.p, 1, time.Until(end))
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
			if len(b) == 1 {
				any = true
				ok = b[0] == 0x06
			}
		}
	}
	if !ok {
		if !any {
			return nil, rErr("la radio non risponde")
		}
		return nil, rErr("la radio non è entrata in programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
	}
	var id []byte
	for k := 0; k < 6 && len(id) < 8; k++ {
		if err := s.send([]byte{'F'}); err != nil {
			return nil, err
		}
		id = append(id, readUpTo(s.ctx, s.p, 8-len(id), time.Second)...)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
	}
	if len(id) < 8 {
		return nil, rErr("identificativo della radio incompleto")
	}
	return id, nil
}

func (f rt480Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	total := rt480Size()
	prog(0, total, "Collegamento con la radio…")
	id, err := s.rt480Enter()
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := make([]byte, 0, total)
	for _, ar := range rt480Areas {
		for a := ar[0]; a < ar[1]; a += 0x40 {
			if ctx.Err() != nil {
				_ = s.send([]byte{'E'})
				return nil, "", errCancelled
			}
			prog(len(data), total, "Lettura…")
			h := []byte{'R', byte(a >> 8), byte(a), 0x40}
			if err := s.send(h); err != nil {
				return nil, "", err
			}
			d := make([]byte, 0x44)
			err := s.p.ReadFull(ctx, d, rt480Timeout)
			if err == nil && (d[1] != h[1] || d[2] != h[2]) {
				err = rErr("risposta inattesa")
			}
			if err != nil {
				_ = s.send([]byte{'E'})
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			data = append(data, d[4:]...)
		}
	}
	_ = s.send([]byte{'E'})
	prog(total, total, "Lettura completata")
	return data, fmt.Sprintf("identificativo %q", id), nil
}

func (f rt480Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < rt480ChEnd {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), rt480Size())
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, rt480ChEnd, "Collegamento con la radio…")
	if _, err := s.rt480Enter(); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < rt480ChEnd; a += 0x40 {
		prog(a, rt480ChEnd, "Scrittura…")
		if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 0x40}, img[a:a+0x40]...)); err != nil {
			return "", err
		}
		ack, err := readByte(ctx, s.p, rt480Timeout)
		if err == nil && ack != 0x06 {
			err = rErr("la radio ha rifiutato il blocco")
		}
		if err != nil {
			_ = s.send([]byte{'E'})
			return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
		}
	}
	// uscita dalla programmazione, come dopo la lettura
	_ = s.send([]byte{'E'})
	prog(rt480ChEnd, rt480ChEnd, "Scrittura completata")
	return "", nil
}
