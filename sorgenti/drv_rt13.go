// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-13: protocollo e memoria ricavati dal programma del costruttore
// «RT-13» versione 1.7 (.NET, progetto «AnaCps», lo stesso dei programmi delle
// Baofeng BF-888S / GT-22), pubblicato sulla pagina dei download di radtels.com.
//
// Protocollo (9600 baud, 8 bit, due bit di stop), parente di quello della
// BF-888S (drv_h777.go) ma con un altro saluto, 22 canali e la password:
//   0x6A + «PWPGJC2» (otto byte insieme) → 0x06; 0x02 → 8 byte (il programma
//   non li controlla); 0x06 → 0x06;
//   password: 'R' 0x0180 8 → 4 byte di intestazione + 8 byte; i primi sei devono
//   essere tutti FF (nessuna password), altrimenti il programma si ferma; 0x06 → 0x06;
//   lettura: 'R' indirizzo(2, big endian) 8 → 4 byte di intestazione (non
//   controllati) + 8 byte; 0x06 → 0x06; alla fine 'E';
//   scrittura: dopo la password 'E' → 'F' e di nuovo il saluto (senza la
//   password); 'W' indirizzo(2) 8 + 8 byte → 0x06; alla fine 'E' → 'F'.
// Senza risposta il programma rinuncia dopo un secondo (cinque scatti di 200 ms).
// Il programma legge 0x0000-0x016F (canali), 0x02B0-0x02BF e 0x03C0 (impostazioni),
// 0x0330-0x033F e 0x0380-0x03DF; scrive 0x0000-0x016F, 0x02B0-0x02BF e
// 0x0380-0x03DF. L'immagine è la memoria da 0x0000 a 0x03DF (FF dove non si legge).
// In scrittura il programma ricostruisce le impostazioni dalle sue finestre e
// mette FF nei byte che non conosce: qui si riscrivono i byte letti dalla radio,
// con i canali modificati, negli stessi blocchi.
//
// Canale (16 byte da 0x0010, 22 canali; libero = quarto byte a FF):
//   0-3 RX e 4-7 TX in BCD little endian, decine di Hz (TX a FF = non trasmette)
//   8-9 tono RX, 10-11 tono TX: FFFF nessuno; bit 0x80 del secondo byte = DCS
//       (0xC0 invertito) con il codice in cifre; altrimenti decimi di Hz in BCD
//   12  bit 0x80 salto di frequenza, 0x40 «cambio di frequenza», 0x10 fuori dalla
//       scansione, 0x08 potenza alta, 0x04 banda stretta, 0x01 blocco su canale
//       occupato spento
//   13  scrambler (0 spento, 1-8); 14-15 FF
// Nelle due griglie del programma (cinese e inglese) e nella finestra di modifica
// in cinese le voci sono «bassa, alta» e «larga, stretta» (posizione = valore del
// bit); solo la traduzione inglese della finestra di modifica le ha al contrario.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const (
	rt13Size    = 0x03E0
	rt13Base    = 0x0010
	rt13Slots   = 22
	rt13PwdAddr = 0x0180
	rt13Timeout = time.Second
)

var (
	rt13Hello = []byte("\x6APWPGJC2")
	// aree lette e scritte dal programma (inizio, fine esclusa), nell'ordine
	rt13ReadAreas  = [][2]int{{0x0000, 0x0170}, {0x02B0, 0x02C0}, {0x0330, 0x0340}, {0x0380, 0x03E0}}
	rt13WriteAreas = [][2]int{{0x0000, 0x0170}, {0x02B0, 0x02C0}, {0x0380, 0x03E0}}
)

const rt13Note = "22 canali senza nome, 400-480 MHz, 9600 baud. Protocollo e memoria ricavati dal programma del costruttore «RT-13» 1.7 e " +
	"provati con una radio simulata, non su una radio vera. Qui si modificano frequenze, toni, potenza, banda e scansione; salto di " +
	"frequenza, scrambler e blocco su canale occupato restano quelli impostati. Se sulla radio è impostata una password va tolta prima " +
	"con il programma del costruttore."

type rt13Opt struct{}

type rt13Fam struct{}

func init() {
	genRegister(&genModel{ID: "rtl-rt13", Vendor: "Radtel", Model: "RT-13", RClass: "IU0VYT_Radtel_RT13", Name: "Radtel RT-13",
		Slots: rt13Slots, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
		Bands:  []band{{400000000, 480000000}},
		Narrow: true, HasScan: true, NoChirp: true, DCS: dtcsAll(), Baud: 9600, Fam: rt13Fam{}, Opt: &rt13Opt{}, Notes: rt13Note})
}

func (rt13Fam) rec(img []byte, i int) []byte { return img[rt13Base+i*16 : rt13Base+i*16+16] }

func (f rt13Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f rt13Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw[:16]) }

func (rt13Fam) Match(m *genModel, img []byte) bool { return false }

// tono come lo legge il programma: senza tono se il secondo byte è zero o il
// valore non è un numero in BCD
func rt13ToneDec(raw []byte) string {
	if raw[1]&0x80 == 0 {
		if raw[1] == 0 || raw[0]&0x0F > 9 || raw[0]>>4 > 9 || raw[1]&0x0F > 9 || raw[1]>>4 > 9 {
			return ""
		}
		return fmtTone(int(bcdByte(raw[0]) + 100*bcdByte(raw[1])))
	}
	if raw[1] == 0xFF {
		return ""
	}
	return h777ToneDec(raw)
}

func (f rt13Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	if r[3] == 0xFF {
		return nil
	}
	v, ok := bcdLE(r[0:4])
	if !ok || v == 0 {
		return nil
	}
	c := &Chan{Rx: v * 10, Tx: v * 10}
	if tx, ok := bcdLE(r[4:8]); r[7] != 0xFF && ok && tx > 0 {
		c.Tx = tx * 10
	} else {
		c.TxOff = true
	}
	c.RxTone = rt13ToneDec(r[8:10])
	c.TxTone = rt13ToneDec(r[10:12])
	c.Scan = r[12]&0x10 == 0
	c.Power = m.powerSem(int(r[12] >> 3 & 1))
	c.Narrow = r[12]&0x04 != 0
	return c
}

// genBandFit (4.23): nelle conversioni e negli inoltri un canale con una frequenza
// fuori dalle bande della radio non viene copiato (con un avviso), invece di
// fermare tutto con l'errore della scrittura
func genBandFit(m *genModel, c *Chan) (bool, string) {
	hzs := []int64{c.Rx}
	if !c.TxOff {
		hzs = append(hzs, c.Tx)
	}
	for _, hz := range hzs {
		if !m.inBands(hz) {
			return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
	}
	return true, ""
}

// genRangeFit (4.23): come genBandFit, per le radio che ammettono ogni frequenza
// multipla di 10 Hz fra due limiti
func genRangeFit(c *Chan, lo, hi int64) (bool, string) {
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	for _, hz := range []int64{c.Rx, tx} {
		if hz%10 != 0 || hz < lo || hz > hi {
			return false, fmt.Sprintf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(hz))
		}
	}
	return true, ""
}

func (rt13Fam) fit(m *genModel, c *Chan) (bool, string) { return genBandFit(m, c) }

func (f rt13Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if c == nil {
		// canale vuoto come lo scrive il programma: frequenze e toni a FF e i
		// valori di partenza (potenza alta, blocco su canale occupato spento)
		copy(r, append(bytes.Repeat([]byte{0xFF}, 12), 0x09, 0x00, 0xFF, 0xFF))
		return nil
	}
	hzs := []int64{c.Rx}
	if !c.TxOff {
		hzs = append(hzs, c.Tx)
	}
	for _, hz := range hzs {
		// limiti del programma per la RT-13 (banda 400-480 MHz)
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
	}
	// salto di frequenza, «cambio di frequenza», blocco su canale occupato e
	// scrambler: quelli del canale; canale nuovo con i valori di partenza del
	// programma (tutto spento)
	keep, scr := byte(0x01), byte(0)
	if f.Get(m, img, i) != nil {
		keep, scr = r[12]&0xC1, r[13]
	}
	putBCDLE(r[0:4], c.Rx/10)
	if c.TxOff {
		copy(r[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		putBCDLE(r[4:8], c.Tx/10)
	}
	r[8], r[9], r[10], r[11] = 0, 0, 0, 0
	h777ToneEnc(r[8:10], rt)
	h777ToneEnc(r[10:12], tt)
	b := keep
	if !c.Scan {
		b |= 0x10
	}
	if m.powerIndex(c.Power) == 1 {
		b |= 0x08
	}
	if c.Narrow {
		b |= 0x04
	}
	r[12], r[13], r[14], r[15] = b, scr, 0xFF, 0xFF
	return nil
}

// ---------------------------------------------------------------- protocollo

// rt13Ack: aspetta il byte di conferma; come il programma, gli altri byte si
// scartano e si rinuncia dopo un secondo
func (s *rawSess) rt13Ack(want byte) (any bool, err error) {
	end := time.Now().Add(rt13Timeout)
	for time.Now().Before(end) {
		b := readUpTo(s.ctx, s.p, 1, time.Until(end))
		if s.ctx.Err() != nil {
			return any, errCancelled
		}
		if len(b) == 1 {
			if b[0] == want {
				return true, nil
			}
			any = true
		}
	}
	return any, errTimeout
}

// rt13Hand: saluto del programma (handShake, stati da 1 a 4)
func (s *rawSess) rt13Hand() ([]byte, error) {
	if err := s.send(rt13Hello); err != nil {
		return nil, err
	}
	if any, err := s.rt13Ack(0x06); err != nil {
		if err != errTimeout {
			return nil, err
		}
		if any {
			return nil, rErr("la radio ha rifiutato la modalità di programmazione")
		}
		return nil, rErr("la radio non risponde")
	}
	if err := s.send([]byte{0x02}); err != nil {
		return nil, err
	}
	id := readUpTo(s.ctx, s.p, 8, rt13Timeout)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(id) < 8 {
		return nil, rErr("identificativo della radio incompleto")
	}
	if err := s.send([]byte{0x06}); err != nil {
		return nil, err
	}
	if _, err := s.rt13Ack(0x06); err != nil {
		if err != errTimeout {
			return nil, err
		}
		return nil, rErr("nessuna conferma dopo l'identificazione")
	}
	return id, nil
}

// rt13Block: lettura di un blocco di 8 byte ('R' → 12 byte, 0x06 → 0x06)
func (s *rawSess) rt13Block(a int) ([]byte, error) {
	if err := s.send([]byte{'R', byte(a >> 8), byte(a), 8}); err != nil {
		return nil, err
	}
	d := make([]byte, 12)
	if err := s.p.ReadFull(s.ctx, d, rt13Timeout); err != nil {
		if err == errCancelled {
			return nil, err
		}
		return nil, rErr("lettura del blocco 0x%04X non riuscita", a)
	}
	if err := s.send([]byte{0x06}); err != nil {
		return nil, err
	}
	if _, err := s.rt13Ack(0x06); err != nil {
		if err != errTimeout {
			return nil, err
		}
		return nil, rErr("nessuna conferma leggendo il blocco 0x%04X", a)
	}
	return d[4:], nil
}

// rt13Enter: saluto e controllo della password; in scrittura il programma esce
// ('E' → 'F') e rifà il saluto, questa volta senza la password
func (s *rawSess) rt13Enter(write bool) ([]byte, error) {
	id, err := s.rt13Hand()
	if err != nil {
		return nil, err
	}
	pw, err := s.rt13Block(rt13PwdAddr)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pw[:6], bytes.Repeat([]byte{0xFF}, 6)) {
		_ = s.send([]byte{'E'})
		return nil, rErr("sulla radio è impostata una password: toglila prima con il programma del costruttore")
	}
	if !write {
		return id, nil
	}
	if err := s.send([]byte{'E'}); err != nil {
		return nil, err
	}
	if _, err := s.rt13Ack('F'); err != nil {
		if err != errTimeout {
			return nil, err
		}
		return nil, rErr("nessuna conferma dopo l'identificazione")
	}
	return s.rt13Hand()
}

func rt13Open(ctx context.Context, m *genModel, port string) (*rawSess, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	// due bit di stop, come il programma
	if err := setTwoStopBits(s.p); err != nil {
		s.p.Close()
		return nil, err
	}
	return s, nil
}

func rt13Total(areas [][2]int) (n int) {
	for _, a := range areas {
		n += a[1] - a[0]
	}
	return n
}

func (f rt13Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rt13Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	total, done := rt13Total(rt13ReadAreas), 0
	prog(0, total, "Collegamento con la radio…")
	id, err := s.rt13Enter(false)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	img := bytes.Repeat([]byte{0xFF}, rt13Size)
	for _, ar := range rt13ReadAreas {
		for a := ar[0]; a < ar[1]; a += 8 {
			if ctx.Err() != nil {
				_ = s.send([]byte{'E'})
				return nil, "", errCancelled
			}
			prog(done, total, "Lettura…")
			d, err := s.rt13Block(a)
			if err != nil {
				_ = s.send([]byte{'E'})
				return nil, "", genFail("Lettura", err)
			}
			copy(img[a:], d)
			done += 8
		}
	}
	_ = s.send([]byte{'E'})
	prog(total, total, "Lettura completata")
	return img, fmt.Sprintf("identificativo %q", id), nil
}

func (f rt13Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < rt13Size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), rt13Size)
	}
	s, err := rt13Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	total, done := rt13Total(rt13WriteAreas), 0
	prog(0, total, "Collegamento con la radio…")
	if _, err := s.rt13Enter(true); err != nil {
		return "", genFail("Identificazione", err)
	}
	for _, ar := range rt13WriteAreas {
		for a := ar[0]; a < ar[1]; a += 8 {
			prog(done, total, "Scrittura…")
			if err := s.send(append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...)); err != nil {
				return "", err
			}
			if _, err := s.rt13Ack(0x06); err != nil {
				if err == errTimeout {
					err = rErr("la radio non ha accettato il blocco 0x%04X", a)
				}
				_ = s.send([]byte{'E'})
				return "", genFail("Scrittura", err)
			}
			done += 8
		}
	}
	if err := s.send([]byte{'E'}); err != nil {
		return "", err
	}
	if _, err := s.rt13Ack('F'); err != nil {
		if err != errTimeout {
			return "", err
		}
		return "", rErr("la radio non ha confermato la fine della scrittura")
	}
	prog(total, total, "Scrittura completata")
	return "", nil
}
