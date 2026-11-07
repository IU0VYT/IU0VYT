// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// tdm11.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: TIDRADIO TD-M11 (tdm11.py di CHIRP): 22 FRS (firmware 0.9.4), 16 PMR
// (firmware 1.0.3) e 16 PMR con firmware 0.9.5. 9600 baud; chiave «STD-M1…»
// → 'A', password vuota (6 × FF) → 'A'; blocchi 'R'/'W' da 16 byte con
// indirizzo little endian e somma di controllo di un byte; 'E' per uscire.
//
// Differenza voluta da CHIRP: CHIRP, prima di scrivere, mette data e ora della
// programmazione in un punto sbagliato della memoria (subito dopo i canali: sul
// 16 PMR dentro i canali nascosti 19-20); qui la memoria si scrive com'è.
// Per lo stesso motivo (mappa delle impostazioni subito dopo i canali invece
// che a 0x300) le impostazioni di CHIRP per questa radio non si usano.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type tdm11Opt struct{ magic string }

var tdm11Order = []string{"STD-M11-", "STD-M12\xff", "STD-M11\xff"}

type tdm11Fam struct{ o *tdm11Opt }

const tdm11Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  lbcd rxtone[2];
  lbcd txtone[2];
  u8 unknown0:2,
     jumpfreq:1,
     scan:1,
     txpower:1,
     narrow:1,
     bcl:2;
  u8 unknown1:3,
     compand:1,
     scramble:4;
  u8 unknown2;
  u8 unknown3;
} memory[%d];
`

var tdm11Magics = map[string]string{
	"tdr-tdm11-22":     "STD-M11-",
	"tdr-tdm11-16":     "STD-M12\xff",
	"tdr-tdm11-16v095": "STD-M11\xff",
}

var tdm11Names = map[string]string{
	"STD-M11-":    "TD-M11 22 FRS (firmware 0.9.4)",
	"STD-M12\xff": "TD-M11 16 PMR (firmware 1.0.3)",
	"STD-M11\xff": "TD-M11 16 PMR (firmware 0.9.5)",
}

func init() {
	pw := []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}}
	note := "Radio senza nomi dei canali. Copia con il cavo sul connettore del microfono, radio accesa su un canale libero e " +
		"volume alto. Se la programmazione è protetta da password va tolta con il programma del costruttore. Blocco del " +
		"canale occupato, salto di frequenza, compander e scrambler si conservano. Driver di CHIRP (sperimentale)."
	mk := func(id, variant, rclass string, slots int, bands []band) {
		o := &tdm11Opt{magic: tdm11Magics[id]}
		genRegister(&genModel{ID: id, Vendor: "TIDRADIO", Model: "TD-M11", Variant: variant, RClass: rclass,
			Name: "TIDRADIO TD-M11 " + variant, Slots: slots, First: 1, NameMax: 0, Powers: pw, Bands: bands,
			Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600, Fam: tdm11Fam{o}, Notes: note})
	}
	mk("tdr-tdm11-22", "22 FRS", "TDM11_22", 22, []band{{136000000, 174000000}, {200000000, 260000000},
		{350000000, 390000000}, {400000000, 520000000}})
	mk("tdr-tdm11-16", "16 PMR", "TDM11_16", 16, []band{{136000000, 174000000}, {400000000, 520000000}})
	mk("tdr-tdm11-16v095", "16 PMR v0.9.5", "TDM11_16_1", 16, []band{{136000000, 174000000}, {400000000, 520000000}})
}

func (tdm11Fam) mem(m *genModel) *bwEl          { return bwMust(fmt.Sprintf(tdm11Format, m.Slots)) }
func (f tdm11Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f tdm11Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f tdm11Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f tdm11Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f tdm11Fam) Match(m *genModel, img []byte) bool {
	return len(img) == 0x450 && bytes.HasPrefix(img[0x400:], []byte("TD-M11"))
}
func (tdm11Fam) crossband() bool { return false }

func (f tdm11Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	rx := r.F("rxfreq").Int(img) * 10
	if rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff = true
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = tdm11ToneDec(r.F("txtone").Int(img))
	c.RxTone = tdm11ToneDec(r.F("rxtone").Int(img))
	c.Scan = r.F("scan").Bool(img)
	c.Power = m.powerSem(int(r.F("txpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

func tdm11ToneDec(v int64) string {
	if v == 0 {
		return ""
	}
	return x3ToneDec(v)
}

func (f tdm11Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	for _, s := range []string{c.TxTone, c.RxTone} {
		if _, err := genToneParse(s); err != nil {
			return err
		}
	}
	// opzioni del canale (come CHIRP: dal canale che c'era)
	var b12, b13 byte
	if f.Get(m, img, i) != nil {
		b12, b13 = r.Raw(img)[12]&0x23, r.Raw(img)[13]&0x1F
	}
	r.Fill(img, 0x00)
	img[r.Off+12], img[r.Off+13] = b12, b13
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	if err := x3ToneEnc(r.F("txtone"), img, c.TxTone); err != nil {
		return err
	}
	if err := x3ToneEnc(r.F("rxtone"), img, c.RxTone); err != nil {
		return err
	}
	r.F("scan").SetBool(img, c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("txpower").SetInt(img, int64(m.powerIndex(c.Power)))
	return nil
}

// ---------------------------------------------------------------- protocollo

func tdm11Sum(b []byte) byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return s
}

// tdm11Magic: la chiave fino a 5 volte (risposta 'A')
func (s *rawSess) tdm11Magic(magic string) (ok, saw bool, err error) {
	for k := 0; k < 5; k++ {
		if err := s.send([]byte(magic)); err != nil {
			return false, saw, err
		}
		b := readUpTo(s.ctx, s.p, 1, 100*time.Millisecond)
		if s.ctx.Err() != nil {
			return false, saw, errCancelled
		}
		if len(b) > 0 {
			saw = true
			if b[0] == 'A' {
				return true, true, nil
			}
		}
	}
	return false, saw, nil
}

func (s *rawSess) tdm11Enter(m *genModel, o *tdm11Opt, write, ident bool) error {
	// lettura: come CHIRP, le chiavi di tutte le versioni nell'ordine 22 FRS,
	// 16 PMR, 16 PMR 0.9.5 (riconoscimento del modello); scrittura: solo la sua
	keys := []string{o.magic}
	if !write {
		keys = tdm11Order
	}
	saw := false
	got := ""
	if ident {
		// 4.9, «Identifica»: le tre chiavi a turno, una per giro (fino a 5 giri),
		// così la 16 PMR 0.9.5 risponde entro 0,3 s e non entro 1 s: il tentativo
		// si chiude se la radio tace per 0,7 s
		for k := 0; k < 5 && got == ""; k++ {
			for _, mg := range keys {
				if err := s.send([]byte(mg)); err != nil {
					return err
				}
				b := readUpTo(s.ctx, s.p, 1, 100*time.Millisecond)
				if s.ctx.Err() != nil {
					return errCancelled
				}
				if len(b) > 0 {
					saw = true
					if b[0] == 'A' {
						got = mg
						break
					}
				}
			}
		}
	}
	for _, mg := range keys {
		if ident {
			break
		}
		ok, sw, err := s.tdm11Magic(mg)
		if err != nil {
			return err
		}
		saw = saw || sw
		if ok {
			got = mg
			break
		}
	}
	if got == "" {
		if !saw {
			return rErr("la radio non risponde")
		}
		return rErr("la radio non è entrata in modalità di programmazione")
	}
	if got != o.magic {
		s.send([]byte{'E'})
		return rErr("la radio è una %s: scegli quella scheda", tdm11Names[got])
	}
	if err := s.send(bytes.Repeat([]byte{0xFF}, 6)); err != nil {
		return err
	}
	b := readUpTo(s.ctx, s.p, 1, 500*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(b) == 0 {
		return rErr("la radio non risponde")
	}
	if b[0] != 'A' {
		return rErr("la programmazione della radio è protetta da password: toglierla con il programma del costruttore")
	}
	if write {
		// identificativo della radio (blocco 0x0310)
		if err := s.send([]byte{0x52, 0x10, 0x03, 0x10}); err != nil {
			return err
		}
		id := readUpTo(s.ctx, s.p, 0x11, 500*time.Millisecond)
		want := []byte{0x00, 0x40, 0x00, 0x52, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		if len(id) != 0x11 {
			if s.ctx.Err() != nil {
				return errCancelled
			}
			return rErr("errore di comunicazione con la radio")
		}
		if !bytes.Equal(id[:16], want) || id[16] != tdm11Sum(id[:16]) {
			return rErr("modello della radio non corrispondente")
		}
	}
	return nil
}

func (f tdm11Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, 0x450, "Collegamento con la radio…")
	if err := s.tdm11Enter(m, f.o, false, identProbing(port)); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	defer s.send([]byte{'E'})
	var data []byte
	for a := 0; a < 0x450; a += 0x10 {
		prog(a, 0x450, "Lettura…")
		if err := s.send([]byte{'R', byte(a), byte(a >> 8), 0x10}); err != nil {
			return nil, "", err
		}
		b := readUpTo(ctx, s.p, 0x11, 500*time.Millisecond)
		if len(b) != 0x11 {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		if b[16] != tdm11Sum(b[:16]) {
			return nil, "", rErr("somma di controllo errata nel blocco 0x%04X", a)
		}
		data = append(data, b[:16]...)
	}
	prog(0x450, 0x450, "Lettura completata")
	return data, "", nil
}

func (f tdm11Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < 0x450 {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), 0x450)
	}
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, 0x450, "Collegamento con la radio…")
	if err := s.tdm11Enter(m, f.o, true, false); err != nil {
		return "", genFail("Identificazione", err)
	}
	defer s.send([]byte{'E'})
	for a := 0; a < 0x450; a += 0x10 {
		prog(a, 0x450, "Scrittura…")
		d := img[a : a+0x10]
		if err := s.send(append(append([]byte{'W', byte(a), byte(a >> 8)}, d...), tdm11Sum(d))); err != nil {
			return "", err
		}
		if b := readUpTo(ctx, s.p, 1, 500*time.Millisecond); len(b) != 1 || b[0] != 'A' {
			if ctx.Err() != nil {
				return "", errCancelled
			}
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	prog(0x450, 0x450, "Scrittura completata")
	return "", nil
}
