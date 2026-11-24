// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: retevis_h777v4.py (4.8).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.8: Retevis H777 V4 (nuova versione della H777, comando «C777HAM») e Retevis
// RT21H («T210GAM», frequenze a 24 bit da 0x2000000)

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"
)

const h777v4Format = `
struct {
  ul32 rxfreq;        // 0-3
  ul32 txfreq;        // 4-7
  lbcd rx_tone[2];    // 8-9
  lbcd tx_tone[2];    // A-B
  u8 unknown0:2,      // C
     scramb:1,        //     Scramble  0 = Off, 1 = On
     scanadd:1,       //     Scan Add  0 = Scan, 1 = Skip
     ishighpower:1,   //     Power Level  0 = Low, 1 = High
     narrow:1,        //     Bandwidth  0 = Wide, 1 = Narrow
     unknown1:1,      //
     bcl:1;           //     Busy Channel Lockout  0 = On, 1 = Off
} memory[16];

struct {
  u8 codesw:1,        // 00D0  Code Switch
     scanm:1,         //       Scan Mode
     voxs:1,          //       VOX Switch
     roger:1,         //       Roger
     voice:1,         //       Voice Annunciation
     unknown_0:1,
     save:1,          //       Battery Save
     beep:1;          //       Beep Tone
  u8 squelch;         // 00D1  Squelch Level
  u8 tot;             // 00D2  Time-out Timer
  u8 vox;             // 00D3  VOX Level
  u8 voxd;            // 00D4  Vox Delay
  u8 unknown_1;
  u8 unknown_2;
  u8 skey2;           // 00D7  Side Key 2 (long)
  u8 unknown_3[5];    // 00D8 - 00DC
  u8 password[6];     // 00DD - 00E2  Password
} settings;
`

const rt21hFormat = `
struct {
  ul24 rxfreq;        // 0-2
  ul24 txfreq;        // 3-5
  lbcd rx_tone[2];    // 6-7
  lbcd tx_tone[2];    // 8-9
  u8 speccode:1,      // A   Spec Code  0 = On, 1 = Off
     compand:1,       //     Compander  0 = On, 1 = Off
     scramb:1,        //     Scramble  0 = Off, 1 = On
     scanadd:1,       //     Scan Add  0 = Scan, 1 = Skip
     ishighpower:1,   //     Power Level  0 = Low, 1 = High
     narrow:1,        //     Bandwidth  0 = Wide, 1 = Narrow
     unknown1:1,      //
     bcl:1;           //     Busy Channel Lockout  0 = On, 1 = Off
  u8 scrambopt;       // B   Scramble Options
} memory[16];

struct {
  u8 unknown_4:1,                  // 00C0
     scanm:1,                      //       Scan Mode
     roger:1,                      //       Roger
     lowbattwarn:1,                //       Low Battery Warning
     unknown_0:1,
     voice:1,                      //       Voice Annunciation
     save:1,                       //       Battery Save
     voxs:1;                       //       VOX Switch
  u8 squelch;                      // 00C1  Squelch Level
  u8 tot;                          // 00C2  Time-out Timer
  u8 voxd;                         // 00C3  Vox Delay
  u8 vox_level;                    // 00C4  VOX Level
  u8 unknown_1;
  u8 unknown_2;
  u8 skey1L:4,                     // 00C7  Side Key 1 (long)
     skey1S:4;                     //       Side Key 1 (short)
  u8 unknown_3:5,                  // 00C8
     removectdcs:1,                //       Remove CT/DCS
     beep:1,                       //       Beep Tone
     specmode:1;                   //       Spec Mode
} settings;
`

// 4.47: Retevis RB45, RB645 e RB47 (programmi «Analog1» con blocchi da 11 byte).
// Canale come quello della RT21H senza l'ultimo byte: frequenze a 24 bit da
// 0x2000000 in unità di 10 Hz, toni, un byte di opzioni. 22 canali (la RB645 ne
// usa 16), poi il blocco delle impostazioni a 0xF2 con la password di
// programmazione negli ultimi 3 byte. Il bit 1 delle opzioni a 0 segna un canale
// che il programma del costruttore chiama «cifrato» (toni nascosti).
const rb45Format = `
struct {
  ul24 rxfreq;        // 0-2
  ul24 txfreq;        // 3-5
  lbcd rx_tone[2];    // 6-7
  lbcd tx_tone[2];    // 8-9
  u8 speccode:1,      // A   Spec Code  0 = On, 1 = Off
     compand:1,       //     Compander  0 = Yes, 1 = No
     scramb:1,        //     Scramble  0 = Yes, 1 = No
     scanadd:1,       //     Scan Add  0 = Yes, 1 = No
     ishighpower:1,   //     TX Power  0 = Low, 1 = High
     narrow:1,        //     W/N  0 = Wide, 1 = Narrow
     plain:1,         //     0 = canale «cifrato»
     bcl:1;           //     Busy Lock  0 = On, 1 = Off
} memory[22];

struct {
  u8 unknown_0[8];    // 00F2  impostazioni (settings_rtv447.go)
  u8 password[3];     // 00FA - 00FC  Password
} settings;
`

type h777v4Fam struct{}

type h777v4Opt struct {
	rt21h   bool
	cmd     string
	idents  [][]byte // risposte ammesse all'identificazione (nessuna: non si controlla, come il programma del costruttore)
	pass    bool     // 4.47: dopo l'identificazione chiede la password di programmazione (comando 05), come il programma del costruttore; non ferma mai la lettura né la scrittura
	ack     byte     // 4.47: conferma della radio (0: 06; RB666 e RB666P: 66)
	echo    bool     // 4.47: la radio può rimandare ciò che riceve (cavo a un filo): si riconosce al collegamento
	bye     bool     // 4.47: dialogo del programma del costruttore: 02 e sigla insieme, pausa prima dei blocchi, chiusura con «b»
	rb45    bool     // 4.47: RB45, RB645, RB47: canali da 11 byte (rb45Format), password di 3 byte, nessuna chiusura dopo la lettura
	dcsDig  bool     // 4.47: codice DCS con le tre cifre nei mezzi byte (RB47) invece che come numero
	rb66    bool     // 4.47: RB66 e RB66P: canali da 8 byte con una frequenza e un tono (drv_rb66.go)
	pwN     int      // 4.47: byte della password che la radio dà al comando 05 (0: 6, oppure 3 con rb45)
	bs      int
	memsize int
	base    int64 // CHIRP base_freq_offset
	bits    int   // bit della frequenza
}

func init() {
	genRegister(&genModel{ID: "rt-h777v4", Vendor: "Retevis", Model: "H777", Variant: "V4", RClass: "H777V4", Name: "Retevis H777 V4 (nuova versione)",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 520000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
		// dialogo di CHIRP, senza la richiesta della password (05) che fa il programma del costruttore
		Opt: &h777v4Opt{cmd: "C777HAM", idents: [][]byte{bytes.Repeat([]byte{0}, 6), bytes.Repeat([]byte{0xFF}, 6)}, bs: 0x0D, memsize: 0xEA, bits: 32},
		Notes: "Nuova versione della Retevis H777 (versioni FRS e PMR446, CHIRP permette tutta la gamma 400-520 MHz): 16 canali senza nomi; " +
			"blocco del canale occupato e scrambler si conservano. Driver di CHIRP."})
	genRegister(&genModel{ID: "rt-rt21h", Vendor: "Retevis", Model: "RT21H", RClass: "RT21HRadio", Name: "Retevis RT21H",
		Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Alta 2 W", 2, 0}},
		Bands: []band{{400000000, 502000000}}, Narrow: true, HasScan: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
		Opt: &h777v4Opt{rt21h: true, cmd: "T210GAM", idents: [][]byte{[]byte("SMP558\x02\xFF"), []byte("SMP558\x02\x00")}, bs: 0x0C, memsize: 0xD8,
			base: 0x2000000, bits: 24},
		Notes: "Versioni FRS e PMR446 (CHIRP permette tutta la gamma 400-502 MHz): 16 canali senza nomi; blocco del canale occupato, scrambler " +
			"(con le 8 varianti), compander e codice speciale si conservano. Driver di CHIRP."})
}

// 4.47: Retevis B3B, B63B, RT22P e RT622P. Il programma del costruttore (famiglia
// «Analog1», .NET) è lo stesso della H777 nuova versione compilato con un'altra
// sigla di collegamento («C160RAM»): stesso dialogo, stessi blocchi da 13 byte
// (lettura 0x00-0xE9, scrittura degli stessi 18 blocchi), stessa posizione e
// stessa codifica di tutti i campi del canale, in lettura e in scrittura
// (confronto metodo per metodo dei programmi; scheda in ANALISI_Retevis_Analog1.md).
// Diverso dal driver della H777: come fa il programma, dopo l'identificazione si
// chiede la password di programmazione (05 → 6 byte; un byte da 0 a 61 è un
// carattere, gli altri valori sono «vuoto») e poi si ripete la conferma (06 → 06).
// Il programma del costruttore, se la radio ha una password diversa da quella
// scritta dall'utente, si ferma; qui no (decisione del 23/11/2026: come CHIRP):
// la risposta si legge e si va avanti, in lettura e in scrittura. La password
// resta nella memoria (6 byte a 0xDD) e si vede e si cambia fra le impostazioni
// (settings_rtv447.go).
// La risposta di identificazione non si controlla (il programma del costruttore
// non la guarda): a riconoscere il modello è la sigla, che una radio diversa
// rifiuta. Le impostazioni (blocco a 0xD0) hanno voci diverse dalla H777: sono
// in settings_rtv447.go.
const h777v4Note447 = "Radio senza licenza con la stessa memoria dei canali della Retevis H777 nuova versione e una sigla di collegamento sua. " +
	"Protocollo ricavato dal programma del costruttore: 16 canali senza nomi; le impostazioni si cambiano dalla scheda «Impostazioni», password di programmazione compresa. " +
	"Una password di programmazione già presente non ferma la lettura né la scrittura (come in CHIRP). Non provata su una radio vera: prima la sola lettura."

func init() {
	for _, x := range []struct{ id, model, name string }{
		{"rt-b3b", "B3B", "Retevis B3B"},
		{"rt-b63b", "B63B", "Retevis B63B"},
		{"rt-rt22p", "RT22P", "Retevis RT22P"},
		{"rt-rt622p", "RT622P", "Retevis RT622P"},
	} {
		genRegister(&genModel{ID: x.id, Vendor: "Retevis", Model: x.model, RClass: "IU0VYT_Retevis_" + x.model, Name: x.name,
			Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
			Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, NoChirp: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
			Opt: &h777v4Opt{cmd: "C160RAM", pass: true, bye: true, bs: 0x0D, memsize: 0xEA, bits: 32}, Notes: h777v4Note447})
	}
	// 4.47: Retevis RB666 e RB666P (sigla «DLHGRA5»). I loro programmi sono lo stesso
	// «Analog1» dei quattro modelli qui sopra: stessa tabella dei comandi (cambia solo
	// la sigla), stessi blocchi (lettura e scrittura da 0x00 a 0xE9), stessa mappa del
	// canale in lettura e stesse scritture con le stesse maschere. Due differenze nel
	// dialogo: la radio conferma con 66 invece di 06 (collegamento, dopo
	// l'identificazione, prima dei blocchi e a ogni blocco scritto) e può rimandare
	// ciò che riceve (il programma della RB666 se lo aspetta sempre, quello della
	// RB666P accetta i due casi): qui si riconosce al collegamento e poi si scarta.
	// Il programma della RB666 conosce anche una serie più vecchia, che al
	// collegamento conferma con 06 e tiene le frequenze in BCD: quel programma la
	// legge ma non riesce a scriverla (aspetta 66 a ogni blocco) e non c'è modo di
	// verificarne la scrittura, quindi qui ci si ferma con un messaggio chiaro.
	for _, x := range []struct{ id, model, name string }{
		{"rt-rb666", "RB666", "Retevis RB666"},
		{"rt-rb666p", "RB666P", "Retevis RB666P"},
	} {
		genRegister(&genModel{ID: x.id, Vendor: "Retevis", Model: x.model, RClass: "IU0VYT_Retevis_" + x.model, Name: x.name,
			Slots: 16, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
			Bands: []band{{400000000, 470000000}}, Narrow: true, HasScan: true, NoChirp: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
			Opt: &h777v4Opt{cmd: "DLHGRA5", pass: true, ack: 0x66, echo: true, bye: true, bs: 0x0D, memsize: 0xEA, bits: 32}, Notes: h777v4Note447})
	}
}

// 4.47: Retevis RB45, RB645 (sigla «M59GRAM») e RB47 («182GRAM»). Programmi
// «Analog1» con blocchi da 11 byte: stesso dialogo dei modelli qui sopra
// (02 e sigla, 02, 06, 05, 06, poi i blocchi), con tre differenze lette nei
// programmi: la password di programmazione è di 3 byte; i blocchi sono 23
// (0x00-0xFC: 22 canali da 11 byte e le impostazioni a 0xF2); dopo la lettura
// il programma non manda la chiusura «b» (dopo la scrittura sì). Il canale è
// quello descritto in rb45Format; RB45 e RB645 scrivono il codice DCS come
// numero (D243N = A3 80), la RB47 con le tre cifre (D243N = 43 82): verificato
// anche sui file dati di fabbrica dei tre programmi. La RB645 è la RB45 con 16
// canali (PMR446). Le impostazioni sono in settings_rtv447.go.
const rb45Note447 = "Radio senza licenza; protocollo ricavato dal programma del costruttore: canali senza nomi; le impostazioni si cambiano dalla scheda «Impostazioni», " +
	"password di programmazione compresa. Una password di programmazione già presente non ferma la lettura né la scrittura (come in CHIRP). " +
	"Un canale che il programma del costruttore mostra come «cifrato» resta com'è finché non lo si modifica. " +
	"Non provata su una radio vera: prima la sola lettura."

func init() {
	for _, x := range []struct {
		id, model, cmd string
		slots          int
		dig            bool
	}{
		{"rt-rb45", "RB45", "M59GRAM", 22, false},
		{"rt-rb645", "RB645", "M59GRAM", 16, false},
		{"rt-rb47", "RB47", "182GRAM", 22, true},
	} {
		genRegister(&genModel{ID: x.id, Vendor: "Retevis", Model: x.model, RClass: "IU0VYT_Retevis_" + x.model, Name: "Retevis " + x.model,
			Slots: x.slots, First: 1, NameMax: 0, Powers: []genPower{{"Bassa", 0, 2}, {"Alta", 0, 0}},
			Bands: []band{{400000000, 480000000}}, Narrow: true, HasScan: true, NoChirp: true, DCS: uv5rDTCS, Baud: 9600, Fam: h777v4Fam{},
			Opt:   &h777v4Opt{cmd: x.cmd, pass: true, bye: true, rb45: true, dcsDig: x.dig, bs: 0x0B, memsize: 0xFD, base: 0x2000000, bits: 24},
			Notes: rb45Note447})
	}
}

func h777v4o(m *genModel) *h777v4Opt { return m.Opt.(*h777v4Opt) }

func (h777v4Fam) mem(m *genModel) *bwEl {
	if h777v4o(m).rt21h {
		return bwMust(rt21hFormat)
	}
	if h777v4o(m).rb45 {
		return bwMust(rb45Format)
	}
	return bwMust(h777v4Format)
}
func (f h777v4Fam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f h777v4Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }
func (f h777v4Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}
func (f h777v4Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }
func (f h777v4Fam) Match(m *genModel, img []byte) bool                { return false }

// h777v4ToneDec: dig = codice DCS scritto con le tre cifre nei mezzi byte (RB47)
// invece che come numero (gli altri modelli)
func h777v4ToneDec(b []byte, dig bool) string {
	if b[0] == 0xFF {
		return ""
	}
	raw := int64(b[1])<<8 | int64(b[0])
	if b[1]&0x80 != 0 {
		n, _ := strconv.Atoi(strconv.FormatInt(raw&0x1FF, 8))
		if dig {
			n, _ = strconv.Atoi(strconv.FormatInt(raw&0xFFF, 16)) // una cifra che non è un numero dà 0, fuori dall'elenco
		}
		if !slices.Contains(uv5rDTCS, n) {
			return ""
		}
		return genTone{Mode: "DTCS", Code: n, Inv: b[1]&0xC0 == 0xC0}.String()
	}
	v := int(b[1]>>4)*1000 + int(b[1]&0xF)*100 + int(b[0]>>4)*10 + int(b[0]&0xF)
	if !slices.Contains(chirpTones, v) {
		return ""
	}
	return fmtTone(v)
}

func h777v4ToneEnc(s string, dig bool) ([]byte, error) {
	t, err := genToneParse(s)
	if err != nil {
		return nil, err
	}
	var v int64 = 0xFFFF
	switch t.Mode {
	case "Tone":
		v = thHex(t.Hz10)
	case "DTCS":
		if dig {
			v = thHex(t.Code)
		} else {
			v, _ = strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		}
		v |= 0x8000
		if t.Inv {
			v |= 0x4000
		}
	}
	return []byte{byte(v), byte(v >> 8)}, nil
}

func (f h777v4Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	o := h777v4o(m)
	if img[r.Off] == 0xFF && (!o.rb45 || r.F("rxfreq").Int(img) == 1<<o.bits-1) {
		// canale libero; nei modelli da 11 byte, come nel loro programma, solo se
		// tutti e tre i byte della frequenza sono FF
		return nil
	}
	c := &Chan{Rx: (r.F("rxfreq").Int(img) + o.base) * 10}
	if r.F("txfreq").Int(img) == 1<<o.bits-1 {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = (r.F("txfreq").Int(img) + o.base) * 10
	}
	if o.rb45 && !r.F("plain").Bool(img) {
		// canale «cifrato»: il programma del costruttore mostra la sola frequenza
		// (la stessa in trasmissione) e nasconde i toni
		c.TxOff, c.Tx = false, c.Rx
	} else {
		c.TxTone = h777v4ToneDec(r.F("tx_tone").Raw(img), o.dcsDig)
		c.RxTone = h777v4ToneDec(r.F("rx_tone").Raw(img), o.dcsDig)
	}
	c.Scan = !r.F("scanadd").Bool(img)
	c.Power = m.powerSem(int(r.F("ishighpower").Int(img)))
	c.Narrow = r.F("narrow").Bool(img)
	return c
}

func (f h777v4Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(m, i)
	o := h777v4o(m)
	if c == nil {
		r.Fill(img, 0xFF)
		return nil
	}
	enc := func(hz int64) (int64, error) {
		v := hz/10 - o.base
		if v < 0 || v > 1<<o.bits-1 {
			return 0, fmt.Errorf("frequenza %s MHz fuori dalla memoria della radio", fmtMHzShort(hz))
		}
		return v, nil
	}
	rx, err := enc(c.Rx)
	if err != nil {
		return err
	}
	tx := int64(1<<o.bits - 1)
	if !c.TxOff {
		if tx, err = enc(c.Tx); err != nil {
			return err
		}
	}
	tt, err := h777v4ToneEnc(c.TxTone, o.dcsDig)
	if err != nil {
		return err
	}
	rt, err := h777v4ToneEnc(c.RxTone, o.dcsDig)
	if err != nil {
		return err
	}
	if o.rb45 {
		if old := f.Get(m, img, i); old != nil && !r.F("plain").Bool(img) {
			if old.Rx == c.Rx && old.Tx == c.Tx && old.TxOff == c.TxOff && old.RxTone == c.RxTone && old.TxTone == c.TxTone &&
				old.Scan == c.Scan && old.Narrow == c.Narrow && m.powerIndex(old.Power) == m.powerIndex(c.Power) {
				return nil // canale «cifrato» non modificato: resta com'è, byte per byte
			}
		}
		// come il programma del costruttore quando si cambiano i toni: il canale non è più «cifrato»
		r.F("plain").SetBool(img, true)
	}
	if o.rt21h && f.Get(m, img, i) != nil && r.F("scrambopt").Int(img) > 7 {
		// come CHIRP: variante dello scrambler fuori dalle 8 ammesse → la prima
		r.F("scrambopt").SetInt(img, 0)
	}
	r.F("rxfreq").SetInt(img, rx)
	r.F("txfreq").SetInt(img, tx)
	r.F("tx_tone").SetRaw(img, tt)
	r.F("rx_tone").SetRaw(img, rt)
	r.F("scanadd").SetBool(img, !c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("ishighpower").SetBool(img, m.powerIndex(c.Power) == 1)
	return nil
}

// h777v4Sess: collegamento con una radio della famiglia. ack è la conferma che la
// radio manda (06, oppure 66 nelle RB666); echo dice che la radio rimanda ciò che
// riceve (cavo a un filo), e allora prima di ogni risposta tornano i byte mandati.
type h777v4Sess struct {
	*rawSess
	ack  byte
	echo bool
}

// xfer: manda tx e legge fino a want byte di risposta (dopo l'eventuale eco)
func (s *h777v4Sess) xfer(tx []byte, want int, to time.Duration) ([]byte, error) {
	if err := s.send(tx); err != nil {
		return nil, err
	}
	n := want
	if s.echo {
		n += len(tx)
	}
	b := readUpTo(s.ctx, s.p, n, to)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if s.echo {
		if len(b) < len(tx) || !bytes.Equal(b[:len(tx)], tx) {
			return nil, rErr("la radio non risponde")
		}
		b = b[len(tx):]
	}
	return b, nil
}

// h777v4Enter: come _h777_enter_programming_mode di h777.py con il comando del
// modello; con bye, 02 e sigla partono insieme come nel programma del costruttore
func h777v4Enter(rs *rawSess, m *genModel) (*h777v4Sess, error) {
	o := h777v4o(m)
	s := &h777v4Sess{rawSess: rs, ack: 0x06}
	if o.ack != 0 {
		s.ack = o.ack
	}
	if o.bye {
		if err := s.send(append([]byte{0x02}, o.cmd...)); err != nil {
			return nil, err
		}
	} else {
		if err := s.send([]byte{0x02}); err != nil {
			return nil, err
		}
		if err := s.sleep(100 * time.Millisecond); err != nil {
			return nil, err
		}
		if err := s.send([]byte(o.cmd)); err != nil {
			return nil, err
		}
	}
	b := readUpTo(s.ctx, s.p, 1, 500*time.Millisecond)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if o.echo && b[0] == 0x02 {
		// eco: tornano gli 8 byte mandati (il primo è già arrivato), poi la conferma
		r := readUpTo(s.ctx, s.p, 8, 500*time.Millisecond)
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(r) != 8 || string(r[:7]) != o.cmd {
			return nil, rErr("la radio non risponde")
		}
		s.echo, b = true, r[7:]
	}
	if b[0] != s.ack {
		if s.ack == 0x66 && b[0] == 0x06 && !o.rb66 { // la serie più vecchia è nota solo per le RB666
			return nil, rErr("questa radio è di una serie più vecchia, con la memoria in un altro formato: non è ancora compatibile")
		}
		return nil, rErr("la radio ha rifiutato la modalità di programmazione")
	}
	id, err := s.xfer([]byte{0x02}, 8, 500*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if len(id) == 0 {
		return nil, rErr("la radio non risponde")
	}
	if b, err := s.xfer([]byte{0x06}, 1, 500*time.Millisecond); err != nil {
		return nil, err
	} else if len(b) != 1 || b[0] != s.ack {
		return nil, rErr("nessuna conferma dopo l'identificazione")
	}
	ok := len(o.idents) == 0
	for _, x := range o.idents {
		if bytes.Contains(id, x) {
			ok = true
		}
	}
	if !ok {
		return nil, rErr("modello errato: la radio si identifica come %q", id)
	}
	if o.pass {
		// 4.47: password di programmazione, come il programma del costruttore
		n, to := 6, 500*time.Millisecond
		if o.rb45 {
			n, to = 3, time.Second // 3 byte; il programma del costruttore aspetta 200 ms prima di leggerli
		}
		if o.pwN != 0 {
			n = o.pwN
		}
		pw, err := s.xfer([]byte{0x05}, n, to)
		if err != nil {
			return nil, err
		}
		if len(pw) != n {
			return nil, rErr("la radio non risponde")
		}
		if n < 6 && o.rb66 {
			// RB66P: il programma del costruttore aspetta 4 byte, poi 30 ms, e scarta
			// quello che arriva in più; qui lo stesso, perché un byte rimasto nella
			// porta farebbe leggere male la conferma che segue
			_ = readUpTo(s.ctx, s.p, 6-n, 60*time.Millisecond)
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
		}
		// la password non ferma niente (come CHIRP): è nella memoria e si vede fra le impostazioni
		if b, err := s.xfer([]byte{0x06}, 1, 500*time.Millisecond); err != nil {
			return nil, err
		} else if len(b) != 1 || b[0] != s.ack {
			return nil, rErr("nessuna conferma dopo l'identificazione")
		}
	}
	if o.bye {
		if err := s.sleep(50 * time.Millisecond); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (f h777v4Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := h777v4o(m)
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer rs.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	s, err := h777v4Enter(rs, m)
	if err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	var data []byte
	for a := 0; a < o.memsize; a += o.bs {
		prog(a, o.memsize, "Lettura…")
		cmd := []byte{'R', byte(a >> 8), byte(a), byte(o.bs)}
		resp, err := s.xfer(cmd, 4+o.bs, time.Second)
		if err != nil {
			return nil, "", err
		}
		if o.rb45 && len(resp) == 4+o.bs {
			// il programma del costruttore non guarda l'intestazione della risposta, e nei
			// suoi file dati (presi da una radio) il byte alto dell'indirizzo non è sempre
			// quello chiesto: qui si controllano gli altri tre byte
			resp[1] = cmd[1]
		}
		if len(resp) != 4+o.bs || !bytes.Equal(resp[:4], []byte{'W', cmd[1], cmd[2], cmd[3]}) {
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita", a)
		}
		data = append(data, resp[4:]...)
	}
	if o.bye && !o.rb45 { // i programmi dei modelli da 11 byte dopo la lettura non mandano niente
		if err := s.send([]byte{'b'}); err != nil {
			return nil, "", err
		}
	}
	prog(o.memsize, o.memsize, "Lettura completata")
	return data, "", nil
}

func (f h777v4Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := h777v4o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.memsize)
	}
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer rs.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	s, err := h777v4Enter(rs, m)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0; a < o.memsize; a += o.bs {
		prog(a, o.memsize, "Scrittura…")
		b, err := s.xfer(append([]byte{'W', byte(a >> 8), byte(a), byte(o.bs)}, img[a:a+o.bs]...), 1, time.Second)
		if err != nil {
			return "", err
		}
		if len(b) != 1 || b[0] != s.ack {
			return "", rErr("la radio non ha accettato il blocco 0x%04X", a)
		}
	}
	if o.bye {
		if err := s.send([]byte{'b'}); err != nil {
			return "", err
		}
	}
	prog(o.memsize, o.memsize, "Scrittura completata")
	return "", nil
}

// fit: frequenze che la memoria della radio può contenere (RT21H: da 335,54432 MHz)
func (f h777v4Fam) fit(m *genModel, c *Chan) (bool, string) {
	o := h777v4o(m)
	ok := func(hz int64) bool { v := hz/10 - o.base; return v >= 0 && v < 1<<o.bits-1 }
	if !ok(c.Rx) || (!c.TxOff && !ok(c.Tx)) {
		return false, "frequenza che questa radio non può memorizzare"
	}
	return true, ""
}
