// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Riferimento incrociato: driver baofeng_digital.py di CHIRP (https://chirpmyradio.com,
// GNU GPL, autore Marcos Del Sol Vives) per W31D, BF-T20D e BF-V12D.

package main

// Tre famiglie di radio PMR446 / FRS / UHF senza display Baofeng ricavate dai
// programmi del costruttore pubblicati sui siti Baofeng e Radioddity.
//
// 1) Serie «D» (BF-V1D, 888SD, 1901D, 889D, V2D…V10D, BF-V12D, 1904D, 1909D,
//    W31D, BF-T20D): programma «BF Series CPS» (pacchetti «Baofeng_V1D_CPS» e
//    «Baofeng_BF-D_Series_CPS»). Protocollo:
//      «\x02prOGRAM» → 4 byte: regione, modello, versione (2 byte);
//      «\x02» + 2 byte del protocollo + «OGRAM» → 06 (80 82 per A/B/D, 20 22 per C);
//      02 → «P3107\xF7\x00\x00»; 06 → 06; «R 01 30 08» → «W 01 30 08» + 8×FF; 06 → 06;
//      lettura: 'R' indirizzo(2) 08 → 'W' indirizzo 08 + 8 byte; 06 → 06;
//      scrittura: dopo il preambolo 'E' → 'F', «\x02\x80\x82OGRAM» → 06,
//      02 → identificativo, 06 → 06; poi 'W' indirizzo 08 + 8 byte → 06. Fine: 'E'.
//    Aree lette: 0x000-0x10F (0x1EF con 30 canali), 0x2B0-0x2BF, 0x330-0x33F
//    (solo lettura), 0x380-0x3DF. Canale n (16 byte) a 0x010 + 16·n: frequenze
//    RX/TX (BCD, decine di Hz), toni RX/TX (BCD, 0x80 = DCS, 0x40 = invertito),
//    impostazioni: bit 7 salto di frequenza, bit 4 = 1 escluso dalla scansione
//    (il programma del costruttore mostra «aggiunto alla scansione: sì» con 0),
//    bit 3 potenza alta, bit 2 banda stretta, bit 1 canale digitale (criptato),
//    bit 0 = 1 senza blocco del canale occupato; byte 13: chiave di cifratura.
//    Nota: il driver di CHIRP legge il bit 4 al contrario (bit 4 = 1 in scansione);
//    qui si segue il programma del costruttore.
//
// 2) BF-T17 / BF-T18 (programmi «T17-T18 PMR446 PRO» e «T17-T18 FRS», .NET):
//    PMR, solo in lettura: «FE 03 30 08» → 12 byte (codice di salto), 'E' → 'F';
//    «\x02PROGPMR» (o «\x02PROGFRS») → 06; 02 → 8 byte; 06 → 06;
//    PMR: «R 07 B0 10» → 20 byte (password in 0x7B0: FF = nessuna); 06 → 06;
//    lettura a blocchi di 16 byte ('R' indirizzo 10 → 'W' indirizzo 10 + 16; 06 → 06)
//    fino a 0x660 (FRS: 0x000-0x22F e 0x630-0x65F); scrittura (PMR: prima 'E' → 'F'
//    e di nuovo l'aggancio) a blocchi di 8 byte → 06. Fine: 'E'.
//
//    La GT-18 PMR («BAOFENG GT18 PMR» v1.2.1) usa lo stesso programma della
//    T17 PMR446 (classe AnaCps.Data.DataProtocol identica) con «\x02PROGRAL».
//
// 3) BF-T9 (programma «BF-Ts» v1.04, pacchetto «BF-T9_Software», .NET): come
//    la FRS qui sopra ma con «\x02PROGRAL» (identificativo «P3107\xF4AM»), lettura
//    di tutta la memoria fino a 0x660 e scrittura a blocchi di 16 byte
//    ('W' indirizzo 10 + 16 byte → 06). 99 canali di 16 byte: i byte 13-15 restano
//    FF; impostazioni (byte 12): bit 4 = 1 escluso dalla scansione, bit 3 = 1
//    potenza ALTA, bit 2 = 1 banda stretta, bit 1 = 0 «beat shift», bit 0 = 0
//    blocco del canale occupato. In 0x650: «\x00\x00» + codice della gamma di
//    frequenze + «7013P» (13 gamme da 136-150 a 470-650 MHz).
//    Canale n (16 byte) a 16·n, 99 canali (FRS: 35): come sopra, ma bit 3 = 1
//    potenza BASSA.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

// ---------------------------------------------------------------- codifica comune

const pmrFormat = `
#seekto 0x%04X;
struct {
    lbcd rxfreq[4];
    lbcd txfreq[4];
    lbcd rxtone[2];
    lbcd txtone[2];
    u8 hopping:1,
       unused1:2,
       skip:1,
       power:1,
       narrow:1,
       digital:1,
       nobcl:1;
    u8 unused2:3,
       key:5;
    u8 unused3[2];
} memory[%d];
`

// pmrCodec: canali di 16 byte come quelli della BF-888S, con le varianti dei
// due programmi del costruttore.
type pmrCodec struct {
	base     int
	lowBit   bool // bit 3 = 1 potenza bassa (T17/T18); altrimenti potenza alta
	digital  bool // canali digitali (serie D)
	newTail  []byte
	newFlags byte
}

func (c pmrCodec) mem(m *genModel) *bwEl { return bwMust(fmt.Sprintf(pmrFormat, c.base, m.Slots)) }

func (c pmrCodec) rec(m *genModel, i int) *bwEl { return c.mem(m).F("memory").I(i) }

// tono: 0x0000 e 0xFFFF = nessuno
func pmrToneDec(raw []byte) string {
	if raw[0] == 0 && raw[1] == 0 {
		return ""
	}
	return h777ToneDec(raw)
}

func (c pmrCodec) get(m *genModel, img []byte, i int) *Chan {
	r := c.rec(m, i)
	rxr := r.F("rxfreq").Raw(img)
	if bytes.Equal(rxr, []byte{0, 0, 0, 0}) || bytes.Equal(rxr, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil
	}
	ch := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if t := r.F("txfreq").Raw(img); bytes.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF}) || bytes.Equal(t, []byte{0, 0, 0, 0}) {
		ch.Tx = ch.Rx
	} else {
		ch.Tx = r.F("txfreq").Int(img) * 10
	}
	ch.Narrow = r.F("narrow").Bool(img)
	hi := r.F("power").Bool(img)
	if c.lowBit {
		hi = !hi
	}
	if hi {
		ch.Power = m.powerSem(1)
	} else {
		ch.Power = m.powerSem(0)
	}
	ch.Scan = !r.F("skip").Bool(img)
	if c.digital && r.F("digital").Bool(img) {
		ch.Dig = true
		return ch
	}
	ch.TxTone = pmrToneDec(r.F("txtone").Raw(img))
	ch.RxTone = pmrToneDec(r.F("rxtone").Raw(img))
	return ch
}

func (c pmrCodec) set(m *genModel, img []byte, i int, ch *Chan) error {
	r := c.rec(m, i)
	if ch == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	if ch.Dig && !c.digital {
		return fmt.Errorf("questa radio non ha canali digitali")
	}
	tt, err := genToneParse(ch.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(ch.RxTone)
	if err != nil {
		return err
	}
	raw := r.Raw(img)
	rx := raw[0:4]
	if bytes.Equal(rx, []byte{0, 0, 0, 0}) || bytes.Equal(rx, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		// canale nuovo: valori del programma del costruttore
		for k := range raw {
			raw[k] = 0
		}
		raw[12] = c.newFlags
		copy(raw[13:], c.newTail)
	}
	r.F("rxfreq").SetInt(img, ch.Rx/10)
	r.F("txfreq").SetInt(img, ch.Tx/10)
	r.F("narrow").SetBool(img, ch.Narrow)
	hi := m.powerIndex(ch.Power) == 1
	if c.lowBit {
		hi = !hi
	}
	r.F("power").SetBool(img, hi)
	r.F("skip").SetBool(img, !ch.Scan)
	if c.digital {
		r.F("digital").SetBool(img, ch.Dig)
	}
	if ch.Dig {
		r.F("txtone").Fill(img, 0xFF)
		r.F("rxtone").Fill(img, 0xFF)
		return nil
	}
	h777ToneEnc(r.F("txtone").Raw(img), tt)
	h777ToneEnc(r.F("rxtone").Raw(img), rt)
	return nil
}

// ---------------------------------------------------------------- serie D

type bfdOpt struct {
	key    byte    // modello comunicato dalla radio
	magic  [2]byte // byte del protocollo dopo 0x02
	unlock bool    // sblocco prima dell'identificazione (BF-V12D)
}

type bfdFam struct{}

// modelli del programma «BF Series CPS» (codice, nome, protocollo, canali)
var bfdModels = []struct {
	key   byte
	label string
	proto byte
}{
	{0x00, "V1D", 'A'}, {0x01, "888SD", 'A'}, {0x02, "1901D", 'C'}, {0x03, "889D", 'B'},
	{0x04, "W31D (GMRS)", 'D'}, {0x05, "T20 / BF-T20D", 'D'}, {0x06, "V4D", 'B'}, {0x07, "MRA-1904 / 1904D", 'A'},
	{0x08, "1909D", 'B'}, {0x09, "UV8D", 'E'}, {0x0A, "BF-V12D", 'A'}, {0x0C, "T12 / T18 (FRS)", 'A'},
	{0x0D, "888H", 'A'}, {0x0E, "T12 / T18 (PMR)", 'A'}, {0x0F, "V8D", 'B'}, {0x10, "V9D", 'A'},
	{0x11, "V10D", 'A'}, {0x12, "V2D", 'A'}, {0x13, "V5D", 'A'}, {0x14, "V6D", 'A'},
}

func bfdName(key byte) string {
	for _, v := range bfdModels {
		if v.key == key {
			return v.label
		}
	}
	return fmt.Sprintf("modello sconosciuto (codice %02X)", key)
}

func bfdo(m *genModel) *bfdOpt { return m.Opt.(*bfdOpt) }

func bfdCodec(m *genModel) pmrCodec {
	return pmrCodec{base: 0x10, digital: true, newFlags: 0x01, newTail: []byte{0, 0, 0}}
}

func bfdRanges(m *genModel) [][3]int { // inizio, fine (esclusa), sola lettura
	end := 0x110
	if m.Slots == 30 {
		end = 0x1F0
	}
	return [][3]int{{0x000, end, 0}, {0x2B0, 0x2C0, 0}, {0x330, 0x340, 1}, {0x380, 0x3E0, 0}}
}

const bfdMemsize = 0x3E0

func init() {
	pw := []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}}
	pmr := []genPower{{"Bassa", 0.5, 2}, {"Alta", 0.5, 0}}
	band := []band{{400000000, 480000000}}
	mk := func(id, model, rclass, name string, slots int, o bfdOpt, p []genPower, nochirp bool, notes string) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "Baofeng", Model: model, RClass: rclass, Name: name, Slots: slots, First: 1,
			Powers: p, Bands: band, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: nochirp, DCS: dtcsAll(), Baud: 9600,
			Fam: bfdFam{}, Opt: &oo, Digital: "Digitale",
			Notes: "Radio senza display: i canali non hanno nome. Canali analogici o digitali (criptati, «DIG»): " +
				"per i canali digitali la chiave di cifratura resta quella impostata. " + notes}
	}
	a := bfdOpt{magic: [2]byte{0x80, 0x82}}
	c := bfdOpt{magic: [2]byte{0x20, 0x22}}
	with := func(o bfdOpt, key byte) bfdOpt { o.key = key; return o }
	cps := "Ricavato dal programma del costruttore «BF Series CPS» (stessa memoria della W31D / BF-V12D di CHIRP)."
	inChirp := "Presente anche in CHIRP (driver «Baofeng digital»), che però legge al contrario l'inclusione nella scansione: qui si segue il programma del costruttore."
	v12 := with(a, 0x0A)
	v12.unlock = true
	genRegister(
		// presenti anche in CHIRP (baofeng_digital.py): file .img compatibili, ma la
		// scansione segue il programma del costruttore (prova a parte in bfd_test.go)
		mk("bf-w31d", "W31D", "BaofengW31D", "Baofeng W31D (GMRS)", 30, with(a, 0x04), pw, true, inChirp),
		mk("bf-t20d", "BF-T20D", "BaofengBFT20D", "Baofeng BF-T20D / T20", 30, with(a, 0x05), pw, true, inChirp),
		mk("bf-v12d", "BF-V12D", "BaofengBFV12D", "Baofeng BF-V12D", 16, v12, pmr, true, inChirp),
		// solo nel programma del costruttore
		mk("bf-v1d", "V1D", "BaofengBFV12D", "Baofeng BF-V1D", 16, with(a, 0x00), pmr, true, cps),
		mk("bf-888sd", "888SD", "BaofengBFV12D", "Baofeng BF-888SD", 16, with(a, 0x01), pmr, true, cps),
		mk("bf-1901d", "1901D", "BaofengBFV12D", "Baofeng BF-1901D", 16, with(c, 0x02), pmr, true, cps),
		mk("bf-889d", "889D", "BaofengBFV12D", "Baofeng BF-889D", 16, with(a, 0x03), pmr, true, cps),
		mk("bf-v4d", "V4D", "BaofengBFV12D", "Baofeng BF-V4D", 16, with(a, 0x06), pmr, true, cps),
		mk("bf-1904d", "1904D", "BaofengBFV12D", "Baofeng MRA-1904 / BF-1904D", 16, with(a, 0x07), pmr, true, cps),
		mk("bf-1909d", "1909D", "BaofengBFV12D", "Baofeng BF-1909D", 16, with(a, 0x08), pmr, true, cps),
		mk("bf-v8d", "V8D", "BaofengBFV12D", "Baofeng BF-V8D", 16, with(a, 0x0F), pmr, true, cps),
		mk("bf-v9d", "V9D", "BaofengBFV12D", "Baofeng BF-V9D", 16, with(a, 0x10), pmr, true, cps),
		mk("bf-v10d", "V10D", "BaofengBFV12D", "Baofeng BF-V10D", 16, with(a, 0x11), pmr, true, cps),
		mk("bf-v2d", "V2D", "BaofengBFV12D", "Baofeng BF-V2D", 16, with(a, 0x12), pmr, true, cps),
		mk("bf-v5d", "V5D", "BaofengBFV12D", "Baofeng BF-V5D", 16, with(a, 0x13), pmr, true, cps),
		mk("bf-v6d", "V6D", "BaofengBFV12D", "Baofeng BF-V6D", 16, with(a, 0x14), pmr, true, cps),
	)
}

func (f bfdFam) layout(m *genModel) *bwEl { return bfdCodec(m).mem(m) }

func (f bfdFam) Get(m *genModel, img []byte, i int) *Chan { return bfdCodec(m).get(m, img, i) }

func (f bfdFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	return bfdCodec(m).set(m, img, i, c)
}

func (f bfdFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, bfdCodec(m).rec(m, i).Raw(img)...)
}

func (f bfdFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	bfdCodec(m).rec(m, i).SetRaw(img, raw)
}

func (f bfdFam) Match(m *genModel, img []byte) bool { return false }

// rxOnlySlot / prove: nessuno

type pmrSess struct {
	ctx context.Context
	p   Port
}

func (s *pmrSess) read(n int, to time.Duration) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, to); err != nil {
		return nil, err
	}
	return b, nil
}

// xfer invia e attende esattamente la risposta indicata
func (s *pmrSess) xfer(cmd, want []byte, what string) error {
	if err := s.p.Write(cmd); err != nil {
		return err
	}
	got, err := s.read(len(want), time.Second)
	if err != nil {
		if err == errTimeout {
			return rErr("nessuna risposta (%s)", what)
		}
		return err
	}
	if !bytes.Equal(got, want) {
		return rErr("risposta inattesa (%s): % X", what, got)
	}
	return nil
}

func (s *pmrSess) sleep(d time.Duration) error {
	select {
	case <-s.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

var bfdIdent = []byte("P3107\xF7\x00\x00")

// identificazione: risposta di 4 byte a «\x02prOGRAM» (tentativi come CHIRP)
func (f bfdFam) radioID(s *pmrSess, m *genModel) ([]byte, error) {
	var last error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			_ = s.p.Write([]byte{'E'})
			if err := s.sleep(time.Second); err != nil {
				return nil, err
			}
		}
		_ = s.p.Purge()
		if bfdo(m).unlock {
			// sblocco del programma del costruttore (password di lettura e scrittura predefinita)
			if last = s.xfer([]byte("\x02ucbfpwd"), []byte{0x06}, "sblocco"); last != nil {
				continue
			}
			if last = s.xfer([]byte{0x10, 0xC5, 0xEA, 0x35}, []byte{0x06}, "sblocco"); last != nil {
				continue
			}
			if last = s.xfer([]byte("Q041OTUwMA==\x00\x00\x00\x00"), []byte{0x06}, "sblocco"); last != nil {
				continue
			}
		}
		if err := s.p.Write([]byte("\x02prOGRAM")); err != nil {
			return nil, err
		}
		id, err := s.read(4, time.Second)
		if err == nil {
			return id, nil
		}
		if err == errCancelled {
			return nil, err
		}
		last = rErr("la radio non risponde")
	}
	return nil, last
}

func (f bfdFam) enter(s *pmrSess, m *genModel, write bool) (string, error) {
	o := bfdo(m)
	id, err := f.radioID(s, m)
	if err != nil {
		return "", err
	}
	info := fmt.Sprintf("modello %s, firmware %X.%02X, regione %02X", bfdName(id[1]), id[2], id[3], id[0])
	if id[1] != o.key {
		return info, rErr("la radio si identifica come Baofeng %s, non come %s: scegli il modello giusto", bfdName(id[1]), m.Name)
	}
	steps := []struct{ cmd, want []byte }{
		{append([]byte{0x02, o.magic[0], o.magic[1]}, "OGRAM"...), []byte{0x06}},
		{[]byte{0x02}, bfdIdent},
		{[]byte{0x06}, []byte{0x06}},
		{[]byte{'R', 0x01, 0x30, 0x08}, []byte("W\x01\x30\x08\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF")},
		{[]byte{0x06}, []byte{0x06}},
	}
	if write {
		steps = append(steps,
			struct{ cmd, want []byte }{[]byte{'E'}, []byte{'F'}},
			struct{ cmd, want []byte }{[]byte("\x02\x80\x82OGRAM"), []byte{0x06}},
			struct{ cmd, want []byte }{[]byte{0x02}, bfdIdent},
			struct{ cmd, want []byte }{[]byte{0x06}, []byte{0x06}},
		)
	}
	for k, st := range steps {
		if err := s.xfer(st.cmd, st.want, fmt.Sprintf("passo %d", k+1)); err != nil {
			return info, err
		}
	}
	return info, nil
}

func pmrOpen(ctx context.Context, port string) (*pmrSess, error) {
	p, err := openPortFunc(port, 9600)
	if err != nil {
		return nil, err
	}
	return &pmrSess{ctx: ctx, p: p}, nil
}

func (f bfdFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := pmrOpen(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	defer s.p.Write([]byte{'E'})
	prog(0, bfdMemsize, "Collegamento con la radio…")
	info, err := f.enter(s, m, false)
	if err != nil {
		return nil, info, genFail("Identificazione", err)
	}
	img := bytes.Repeat([]byte{0xFF}, bfdMemsize)
	for _, r := range bfdRanges(m) {
		for a := r[0]; a < r[1]; a += 8 {
			if ctx.Err() != nil {
				return nil, info, errCancelled
			}
			prog(a, bfdMemsize, "Lettura…")
			if err := s.p.Write([]byte{'R', byte(a >> 8), byte(a), 8}); err != nil {
				return nil, info, err
			}
			resp, err := s.read(12, time.Second)
			if err != nil {
				if err == errCancelled {
					return nil, info, err
				}
				return nil, info, rErr("lettura del blocco 0x%03X non riuscita", a)
			}
			if err := s.xfer([]byte{0x06}, []byte{0x06}, fmt.Sprintf("conferma del blocco 0x%03X", a)); err != nil {
				return nil, info, genFail("Lettura", err)
			}
			copy(img[a:], resp[4:])
		}
	}
	prog(bfdMemsize, bfdMemsize, "Lettura completata")
	return img, info, nil
}

func (f bfdFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < bfdMemsize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := pmrOpen(ctx, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	defer s.p.Write([]byte{'E'})
	prog(0, bfdMemsize, "Collegamento con la radio…")
	info, err := f.enter(s, m, true)
	if err != nil {
		return info, genFail("Identificazione", err)
	}
	for _, r := range bfdRanges(m) {
		if r[2] == 1 {
			continue
		}
		for a := r[0]; a < r[1]; a += 8 {
			if ctx.Err() != nil {
				return info, errCancelled
			}
			prog(a, bfdMemsize, "Scrittura…")
			cmd := append([]byte{'W', byte(a >> 8), byte(a), 8}, img[a:a+8]...)
			if err := s.xfer(cmd, []byte{0x06}, fmt.Sprintf("scrittura del blocco 0x%03X", a)); err != nil {
				return info, genFail("Scrittura", err)
			}
		}
	}
	prog(bfdMemsize, bfdMemsize, "Scrittura completata")
	return info, nil
}

// ---------------------------------------------------------------- BF-T17 / BF-T18

type t17Opt struct {
	cmd   string // «PROGPMR», «PROGFRS» o «PROGRAL» (BF-T9)
	pmr   bool   // versione PMR446: password e doppio aggancio in scrittura
	gaps  [][2]int
	wblk  int      // byte per blocco in scrittura (8; 16 per la BF-T9)
	codec pmrCodec // codifica dei canali
	rng   int      // indirizzo del codice della gamma di frequenze (0 = nessuno)
}

type t17Fam struct{}

func t17o(m *genModel) *t17Opt { return m.Opt.(*t17Opt) }

func t17Codec(m *genModel) pmrCodec { return t17o(m).codec }

// gamme della BF-T9 (codice in 0x652, tabella TAB_MODELTYPE del programma «BF-Ts»)
var t9Ranges = map[byte]band{
	0xF1: {136000000, 150000000}, 0xF2: {136000000, 174000000}, 0xF0: {150000000, 174000000}, 0xF3: {216000000, 223000000},
	0xFA: {350000000, 370000000}, 0xFB: {350000000, 390000000}, 0xFC: {370000000, 390000000}, 0xF8: {400000000, 420000000},
	0xF7: {400000000, 470000000}, 0xF9: {440000000, 480000000}, 0xF4: {450000000, 470000000}, 0xF5: {450000000, 490000000},
	0xF6: {470000000, 650000000},
}

// gamma di frequenze della radio letta (ok = false se l'immagine non la indica)
func t9Range(m *genModel, img []byte) (band, bool) {
	o := t17o(m)
	if o.rng == 0 || len(img) <= o.rng {
		return band{}, false
	}
	b, ok := t9Ranges[img[o.rng]]
	return b, ok
}

const t17Memsize = 0x660

func init() {
	pw := []genPower{{"Bassa", 0.5, 2}, {"Alta", 0.5, 0}}
	cod := pmrCodec{base: 0, lowBit: true, newFlags: 0x01, newTail: []byte{0xFF, 0xFF, 0xFF}}
	pmr := t17Opt{cmd: "PROGPMR", pmr: true, gaps: [][2]int{{0x000, 0x660}}, wblk: 8, codec: cod}
	frs := t17Opt{cmd: "PROGFRS", gaps: [][2]int{{0x000, 0x230}, {0x630, 0x660}}, wblk: 8, codec: cod}
	mk := func(id, name string, slots int, o t17Opt, b []band) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "Baofeng", Model: "BF-T17", Variant: o.cmd[4:], RClass: "IU0VYT_BFT17_" + o.cmd[4:], Name: name,
			Slots: slots, First: 1, Powers: pw, Bands: b, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
			DCS: dtcsAll(), Baud: 9600, Fam: t17Fam{}, Opt: &oo,
			Notes: "Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore «T17-T18 " +
				o.cmd[4:] + "» (v1.2.2); se sulla radio è impostata la password di lettura e scrittura, toglila prima con quel programma."}
	}
	genRegister(
		mk("bf-t17pmr", "Baofeng BF-T17 / BF-T18 (PMR446)", 99, pmr, []band{{400000000, 480000000}}),
		mk("bf-t17frs", "Baofeng BF-T17 / BF-T18 (FRS, USA)", 35, frs, []band{{400000000, 480000000}}),
	)
	// GT-18 PMR: programma «BAOFENG GT18 PMR» (v1.2.1), identico a quello della
	// T17 PMR446 tranne l'aggancio «\x02PROGRAL»
	gt := pmr
	gt.cmd = "PROGRAL"
	genRegister(&genModel{ID: "bf-gt18pmr", Vendor: "Baofeng", Model: "GT-18", Variant: "PMR", RClass: "IU0VYT_GT18_PMR", Name: "Baofeng GT-18 (PMR446)",
		Slots: 99, First: 1, Powers: pw, Bands: []band{{400000000, 480000000}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
		DCS: dtcsAll(), Baud: 9600, Fam: t17Fam{}, Opt: &gt,
		Notes: "Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore «BAOFENG GT18 PMR» (v1.2.1, stesso protocollo " +
			"della BF-T17 PMR446); se sulla radio è impostata la password di lettura e scrittura, toglila prima con quel programma."})
	// BF-T9: nuovo canale come nel programma «BF-Ts» (potenza alta, banda larga,
	// in scansione, senza blocco del canale occupato)
	t9 := t17Opt{cmd: "PROGRAL", gaps: [][2]int{{0x000, 0x660}}, wblk: 16, rng: 0x652,
		codec: pmrCodec{base: 0, newFlags: 0xEB, newTail: []byte{0xFF, 0xFF, 0xFF}}}
	genRegister(&genModel{ID: "bf-t9", Vendor: "Baofeng", Model: "BF-T9", RClass: "IU0VYT_BFT9", Name: "Baofeng BF-T9",
		Slots: 99, First: 1, Powers: []genPower{{"Bassa", 0.5, 2}, {"Alta", 2, 0}},
		Bands:  []band{{136000000, 174000000}, {216000000, 223000000}, {350000000, 650000000}},
		Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true, DCS: dtcsAll(), Baud: 9600, Fam: t17Fam{}, Opt: &t9,
		Notes: "Radio senza display: i canali non hanno nome. Ricavato dal programma del costruttore «BF-Ts» (v1.04, pacchetto «BF-T9 Software»). " +
			"La gamma di frequenze ammessa è quella memorizzata nella radio (es. 400-470 MHz) e viene controllata prima della scrittura."})
}

func (f t17Fam) layout(m *genModel) *bwEl { return t17Codec(m).mem(m) }

func (f t17Fam) Get(m *genModel, img []byte, i int) *Chan { return t17Codec(m).get(m, img, i) }

func (f t17Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if c != nil {
		if r, ok := t9Range(m, img); ok {
			for _, fq := range []int64{c.Rx, c.Tx} {
				if fq < r.Lo || fq > r.Hi {
					return fmt.Errorf("%s MHz fuori dalla gamma di questa radio (%s-%s MHz)", fmtMHzShort(fq), fmtMHzShort(r.Lo), fmtMHzShort(r.Hi))
				}
			}
		}
	}
	return t17Codec(m).set(m, img, i, c)
}

func (f t17Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, t17Codec(m).rec(m, i).Raw(img)...)
}

func (f t17Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	t17Codec(m).rec(m, i).SetRaw(img, raw)
}

func (f t17Fam) Match(m *genModel, img []byte) bool { return false }

func (f t17Fam) handshake(s *pmrSess, m *genModel) ([]byte, error) {
	if err := s.xfer(append([]byte{0x02}, t17o(m).cmd...), []byte{0x06}, "modalità di programmazione"); err != nil {
		return nil, err
	}
	if err := s.p.Write([]byte{0x02}); err != nil {
		return nil, err
	}
	ident, err := s.read(8, time.Second)
	if err != nil {
		if err == errTimeout {
			return nil, rErr("nessun identificativo dalla radio")
		}
		return nil, err
	}
	if err := s.xfer([]byte{0x06}, []byte{0x06}, "conferma dell'identificativo"); err != nil {
		return ident, err
	}
	return ident, nil
}

func (f t17Fam) enter(s *pmrSess, m *genModel, write bool) (string, error) {
	ident, err := f.handshake(s, m)
	info := fmt.Sprintf("identificativo % X", ident)
	if err != nil {
		return info, err
	}
	if !t17o(m).pmr {
		return info, nil
	}
	if err := s.p.Write([]byte{'R', 0x07, 0xB0, 0x10}); err != nil {
		return info, err
	}
	pw, err := s.read(20, time.Second)
	if err != nil {
		if err == errTimeout {
			return info, rErr("nessuna risposta alla richiesta della password")
		}
		return info, err
	}
	if !bytes.Equal(pw[4:10], bytes.Repeat([]byte{0xFF}, 6)) {
		return info, rErr("sulla radio è impostata una password di lettura e scrittura: toglila con il programma del costruttore")
	}
	if err := s.xfer([]byte{0x06}, []byte{0x06}, "conferma della password"); err != nil {
		return info, err
	}
	if write {
		if err := s.xfer([]byte{'E'}, []byte{'F'}, "conferma della scrittura"); err != nil {
			return info, err
		}
		if _, err := f.handshake(s, m); err != nil {
			return info, err
		}
	}
	return info, nil
}

func (f t17Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := pmrOpen(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, t17Memsize, "Collegamento con la radio…")
	jump := ""
	if t17o(m).pmr {
		// come il programma del costruttore: prima il codice di salto di frequenza
		if err := s.p.Write([]byte{0xFE, 0x03, 0x30, 0x08}); err != nil {
			return nil, "", err
		}
		r, err := s.read(12, time.Second)
		if err != nil {
			if err == errTimeout {
				err = rErr("la radio non risponde")
			}
			return nil, "", genFail("Identificazione", err)
		}
		jump = fmt.Sprintf(", codice di salto %d", int(r[5])<<8|int(r[4]))
		if err := s.xfer([]byte{'E'}, []byte{'F'}, "codice di salto"); err != nil {
			return nil, "", genFail("Identificazione", err)
		}
	}
	info, err := f.enter(s, m, false)
	info += jump
	if err != nil {
		return nil, info, genFail("Identificazione", err)
	}
	img := bytes.Repeat([]byte{0xFF}, t17Memsize)
	for _, g := range t17o(m).gaps {
		for a := g[0]; a < g[1]; a += 16 {
			if ctx.Err() != nil {
				return nil, info, errCancelled
			}
			prog(a, t17Memsize, "Lettura…")
			if err := s.p.Write([]byte{'R', byte(a >> 8), byte(a), 0x10}); err != nil {
				return nil, info, err
			}
			resp, err := s.read(20, time.Second)
			if err != nil {
				if err == errCancelled {
					return nil, info, err
				}
				return nil, info, rErr("lettura del blocco 0x%03X non riuscita", a)
			}
			if err := s.xfer([]byte{0x06}, []byte{0x06}, fmt.Sprintf("conferma del blocco 0x%03X", a)); err != nil {
				return nil, info, genFail("Lettura", err)
			}
			copy(img[a:], resp[4:])
		}
	}
	_ = s.p.Write([]byte{'E'})
	prog(t17Memsize, t17Memsize, "Lettura completata")
	return img, info, nil
}

func (f t17Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < t17Memsize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	s, err := pmrOpen(ctx, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, t17Memsize, "Collegamento con la radio…")
	info, err := f.enter(s, m, true)
	if err != nil {
		return info, genFail("Identificazione", err)
	}
	for _, g := range t17o(m).gaps {
		n := t17o(m).wblk
		for a := g[0]; a < g[1]; a += n {
			if ctx.Err() != nil {
				return info, errCancelled
			}
			prog(a, t17Memsize, "Scrittura…")
			cmd := append([]byte{'W', byte(a >> 8), byte(a), byte(n)}, img[a:a+n]...)
			if err := s.xfer(cmd, []byte{0x06}, fmt.Sprintf("scrittura del blocco 0x%03X", a)); err != nil {
				return info, genFail("Scrittura", err)
			}
		}
	}
	_ = s.p.Write([]byte{'E'})
	prog(t17Memsize, t17Memsize, "Scrittura completata")
	return info, nil
}
