// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// alinco.py (4.40).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.40: Alinco. Due famiglie con lo stesso modo di parlare (comandi di testo
// «AL~F…», dati in esadecimale):
//
//   - alincoFam: palmari DJ-596 e DJ-175, veicolari DR-135T, DR-235T, DR-435T,
//     DR-03T e DR-06T: 100 canali (da 0 a 99) con nomi di 7 caratteri, 9600
//     baud, blocchi da 16 byte;
//   - alg7Fam: palmare tribanda DJ-G7 (DJ-G7EG e DJ-G7T): 1000 canali (da 0 a
//     999) con nomi di 16 caratteri, 57600 baud (19200 se la radio è impostata
//     così), blocchi da 64 byte con conferma.
//
// Collegamento. Il cavo Alinco a un filo rimanda indietro tutto quello che
// riceve: CHIRP conta su quell'eco e con un cavo che non la fa non funziona.
// Qui la risposta si legge a righe: se la prima riga è il comando appena
// inviato è l'eco e si passa oltre, altrimenti è già la risposta. I byte
// inviati alla radio sono gli stessi di CHIRP, con le stesse pause.
//
// Nella prima famiglia la radio non conferma i blocchi scritti (nemmeno con
// CHIRP): il controllo è la rilettura dei canali che il programma fa dopo
// ogni scrittura.
//
// Differenze volute rispetto a CHIRP:
//   - una frequenza o uno scostamento che la radio non può memorizzare (non
//     multipli di 100 Hz nella prima famiglia) vengono rifiutati (CHIRP li
//     tronca);
//   - DJ-175: CHIRP tiene l'esclusione dalla scansione nei byte da 0x130, che
//     in questa radio sono il nome di una delle memorie fuori elenco (quelle da
//     0x100, con la stessa forma di un canale): ogni canale scritto ne cambia
//     un bit. Qui quei byte non si toccano e la colonna della scansione non
//     c'è;
//   - DJ-G7: la banda aerea (da 108 a 137 MHz) si scrive in AM e la
//     radiodiffusione FM (da 76 a 108 MHz) in FM larga, perché in FM stretta o
//     normale quei canali non servirebbero; AM e FM larga di un canale restano
//     se la frequenza non cambia. Lo squelch «al contrario» (TSQL-R) di un
//     canale resta se non gli si dà un tono;
//   - toni che la radio non ha vengono rifiutati con un messaggio (in CHIRP
//     alcune combinazioni si scrivono in un altro modo senza dirlo).

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// alincoTones: i subtoni di CHIRP senza gli undici che le Alinco non hanno
var alincoTones = func() []int {
	no := map[int]bool{1598: true, 1655: true, 1713: true, 1773: true, 1835: true, 1899: true, 1966: true, 1995: true,
		2065: true, 2291: true, 2541: true}
	var out []int
	for _, t := range chirpTones {
		if !no[t] {
			out = append(out, t)
		}
	}
	return out
}()

const alincoToneDef = 8 // 88,5 Hz: il subtono di un canale nuovo, come in CHIRP

// alincoGap: pausa prima di ogni blocco nella prima famiglia (come CHIRP). Le
// prove con la radio simulata la accorciano; per le radio vere è alincoGapStd
const alincoGapStd = 100 * time.Millisecond

var alincoGap = alincoGapStd

const alincoCable = "Cavo Alinco ERW-4C / ERW-7 (o compatibile) nella presa dell'altoparlante. "

// alincoToneMode: 0 nessun tono, 1 subtono in trasmissione, 2 lo stesso subtono
// nei due versi, 3 lo stesso codice DCS nei due versi, 4 due subtoni diversi
// (solo DJ-G7); -1 se le Alinco non hanno la combinazione
func alincoToneMode(t, r genTone, cross bool) int {
	has := func(g genTone) bool { return g.Mode != "Tone" || icomIndex(alincoTones, g.Hz10) >= 0 }
	switch {
	case t.Inv || r.Inv || !has(t) || !has(r):
		return -1
	case t.Mode == "" && r.Mode == "":
		return 0
	case t.Mode == "Tone" && r.Mode == "":
		return 1
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
		return 2
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code && icomIndex(dtcsCodes, t.Code) >= 0:
		return 3
	case t.Mode == "Tone" && r.Mode == "Tone" && cross:
		return 4
	}
	return -1
}

var errAlincoTones = fmt.Errorf("la radio non ha questa combinazione di toni (subtono in trasmissione, lo stesso subtono o lo stesso codice DCS nei due versi; mai codici invertiti)")

// alincoSplit: scostamento del canale; ok falso se CHIRP lo tratterebbe come
// «split», che queste radio non hanno
func alincoFar(c *Chan) bool {
	d := c.Tx - c.Rx
	return d > 70000000 || d < -70000000
}

// ---------------------------------------------------------------- DJ-596, DJ-175, DR-x35

type alincoFam struct{}

type alincoOpt struct {
	ident string // nome a cui la radio risponde
	size  int
	sig   []byte // byte a 0x64 con cui CHIRP riconosce i file (nil: dalla lunghezza)
	// dj175: canale occupato scritto nel canale, potenza a tre livelli, risposte
	// senza indirizzo, nessuna esclusione dalla scansione
	dj175 bool
}

const alincoFormat = `
#seekto 0x0120;
u8 used_flags[25];
#seekto 0x0130;
u8 skips[25];
#seekto 0x0200;
struct {
  u8 new_used:1,
     unknown1:1,
     isnarrow:1,
     isdigital:1,
     ishigh:1,
     unknown2:3;
  u8 unknown3:6,
     duplex:2;
  u8 unknown4:4,
     tmode:4;
  u8 unknown5:4,
     step:4;
  bbcd freq[4];
  u8 unknown6[1];
  bbcd offset[3];
  u8 rtone;
  u8 ctone;
  u8 dtcs_tx;
  u8 dtcs_rx;
  u8 name[7];
  u8 unknown8[2];
  u8 unknown9:6,
     power:2;
  u8 unknownA[6];
} memory[100];
`

// nomi: cifre da 0x30, lettere da 0x3A, spazio 0x54 (come CHIRP)
const alincoChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ "

func init() {
	one := func(w float64) []genPower { return []genPower{{"Unica", w, 0}} }
	mk := func(id, model, rclass, ident string, size int, sig []byte, pw []genPower, bands []band, what string) *genModel {
		o := &alincoOpt{ident: ident, size: size, sig: sig, dj175: model == "DJ175"}
		note := "100 canali (da 0 a 99) con nomi di 7 caratteri (lettere maiuscole, cifre e spazio). " + alincoCable +
			"9600 baud; la radio va accesa normalmente. La radio non conferma i blocchi scritti: il controllo è la rilettura dei canali dopo la scrittura. " +
			"DCS con lo stesso codice nei due versi, mai invertito. " + what + "Driver di CHIRP."
		return &genModel{ID: id, Vendor: "Alinco", Model: model, RClass: rclass, Name: "Alinco " + alincoShown(model),
			Slots: 100, First: 0, NameMax: 7, Charset: alincoChars, Powers: pw, Bands: bands, Narrow: true, HasScan: !o.dj175,
			NoTxOff: true, DCS: dtcsCodes, OneDCS: true, Baud: 9600, Fam: alincoFam{}, Opt: o, Notes: note}
	}
	noPw := "La potenza non è memorizzata nel canale. "
	genRegister(
		mk("al-dj596", "DJ596", "DJ596Radio", "DJ596", 4096, []byte{0x45, 0x01},
			[]genPower{{"Bassa 1 W", 1, 2}, {"Alta 5 W", 5, 0}}, []band{{136000000, 174000000}, {400000000, 511000000}}, ""),
		mk("al-dj175", "DJ175", "DJ175Radio", "DJ175", 6896, nil,
			[]genPower{{"Bassa 0,5 W", 0.5, 2}, {"Media 2 W", 2, 1}, {"Alta 5 W", 5, 0}}, []band{{136000000, 174000000}, {400000000, 511000000}},
			"L'esclusione dalla scansione non si gestisce da qui. "),
		mk("al-dr135t", "DR135T", "DR135Radio", "DR135", 4096, []byte{0x01, 0x44}, one(50), []band{{118000000, 174000000}}, noPw),
		mk("al-dr235t", "DR235T", "DR235Radio", "DR235", 4096, []byte{0x02, 0x22}, one(25), []band{{216000000, 280000000}}, noPw),
		mk("al-dr435t", "DR435T", "DR435Radio", "DR435", 4096, []byte{0x04, 0x00}, one(35), []band{{350000000, 511000000}}, noPw),
		mk("al-dr03t", "DR03T", "DR03Radio", "DR135", 4096, []byte{0x00, 0x28}, one(10), []band{{28000000, 29695000}}, noPw),
		mk("al-dr06t", "DR06T", "DR06Radio", "DR435", 4096, []byte{0x00, 0x50}, one(50), []band{{50000000, 53995000}}, noPw),
	)
}

// alincoShown: nome come sulla radio (DJ-596, DR-135T) dal modello di CHIRP (DJ596, DR135T)
func alincoShown(model string) string {
	if len(model) > 2 && !strings.Contains(model, "-") {
		return model[:2] + "-" + model[2:]
	}
	return model
}

func alincoO(m *genModel) *alincoOpt { return m.Opt.(*alincoOpt) }

func (alincoFam) mem() *bwEl                 { return bwMust(alincoFormat) }
func (f alincoFam) layout(m *genModel) *bwEl { return f.mem() }
func (f alincoFam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

// used: la posizione è occupata (DJ-175: bit nel canale; le altre: elenco da 0x120)
func (f alincoFam) used(m *genModel, img []byte, i int) bool {
	if alincoO(m).dj175 {
		return f.rec(i).F("new_used").Bool(img)
	}
	return f.mem().F("used_flags").I(i/8).Int(img)&int64(0x80>>(i%8)) != 0
}

func (f alincoFam) setBit(el *bwEl, img []byte, i int, on bool) {
	v := el.I(i / 8).Int(img)
	if on {
		v |= int64(0x80 >> (i % 8))
	} else {
		v &^= int64(0x80 >> (i % 8))
	}
	el.I(i/8).SetInt(img, v)
}

func (f alincoFam) setUsed(m *genModel, img []byte, i int, on bool) {
	if alincoO(m).dj175 {
		f.rec(i).F("new_used").SetBool(img, on)
		return
	}
	f.setBit(f.mem().F("used_flags"), img, i, on)
}

func (f alincoFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}

func (f alincoFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) != f.rec(i).ByteLen() {
		return
	}
	f.rec(i).SetRaw(img, raw)
	f.setUsed(m, img, i, true)
}

// Match: come CHIRP (lunghezza e, dove ci sono, i due byte dei limiti a 0x64)
func (alincoFam) Match(m *genModel, img []byte) bool {
	o := alincoO(m)
	return len(img) == o.size && (o.sig == nil || bytes.Equal(img[0x64:0x66], o.sig))
}

func (alincoFam) toneOK(t, r genTone) bool { return alincoToneMode(t, r, false) >= 0 }

// identLast: in «Identifica» le Alinco si provano dopo tutte le altre radio (vedi
// identPlan): sono poco diffuse e così la ricerca delle altre non si allunga
func (alincoFam) identLast() {}
func (alg7Fam) identLast()   {}
func (alincoFam) fitBySet()  {}

func (f alincoFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !f.toneOK(t, r) || alincoFar(c)
}

func (f alincoFam) Get(m *genModel, img []byte, i int) *Chan {
	if !f.used(m, img, i) {
		return nil
	}
	o, r := alincoO(m), f.rec(i)
	c := &Chan{Rx: r.F("freq").Int(img) * 100}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx = c.Rx
	offs := r.F("offset").Int(img) * 100
	switch r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx - offs
	case 2:
		c.Tx = c.Rx + offs
	}
	tone := func(name string) string {
		if k := r.F(name).Int(img); k < int64(len(alincoTones)) {
			return fmtTone(alincoTones[k])
		}
		return ""
	}
	switch r.F("tmode").Int(img) {
	case 1:
		c.TxTone = tone("rtone")
	case 3:
		c.TxTone = tone("ctone")
		c.RxTone = c.TxTone
	case 12:
		if k := r.F("dtcs_tx").Int(img); k < int64(len(dtcsCodes)) {
			c.TxTone = fmt.Sprintf("D%03dN", dtcsCodes[k])
			c.RxTone = c.TxTone
		}
	}
	c.Narrow = r.F("isnarrow").Bool(img)
	switch {
	case o.dj175:
		c.Power = m.powerSem(int(r.F("power").Int(img)))
		c.Scan = true
	default:
		if len(m.Powers) > 1 {
			c.Power = m.powerSem(int(r.F("ishigh").Int(img)))
		}
		c.Scan = f.mem().F("skips").I(i/8).Int(img)&int64(0x80>>(i%8)) == 0
	}
	var name []byte
	for _, b := range r.F("name").Raw(img) {
		if b == 0x00 || b == 0xFF {
			break
		}
		if k := int(b) - 0x30; k >= 0 && k < len(alincoChars) {
			name = append(name, alincoChars[k])
		}
	}
	c.Name = strings.TrimRight(string(name), " ")
	return c
}

func (f alincoFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o, r := alincoO(m), f.rec(i)
	if c == nil {
		f.setUsed(m, img, i, false) // come CHIRP: il resto resta com'è
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := alincoToneMode(t, rt, false)
	if tmode < 0 {
		return errAlincoTones
	}
	if c.Rx%100 != 0 || c.Rx/100 > 99999999 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtoni 88,5 Hz, DCS 023, passo 5 kHz, scostamento 600 kHz)
	was := f.used(m, img, i)
	rtone, ctone, dcs, step, offs, digital := int64(alincoToneDef), int64(alincoToneDef), int64(0), int64(0), int64(600000), r.F("isdigital").Bool(img)
	if was {
		rtone, ctone, dcs = r.F("rtone").Int(img), r.F("ctone").Int(img), r.F("dtcs_tx").Int(img)
		step, offs = r.F("step").Int(img), r.F("offset").Int(img)*100
		if rtone >= int64(len(alincoTones)) {
			rtone = alincoToneDef
		}
		if ctone >= int64(len(alincoTones)) {
			ctone = alincoToneDef
		}
		if dcs >= int64(len(dtcsCodes)) {
			dcs = 0
		}
		if step > 6 {
			step = 0
		}
	}
	switch tmode {
	case 1:
		rtone = icomIndex(alincoTones, t.Hz10)
	case 2:
		rtone = icomIndex(alincoTones, t.Hz10)
		ctone = rtone
	case 3:
		dcs = icomIndex(dtcsCodes, t.Code)
	}
	dup := int64(0)
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if offs%100 != 0 {
		return fmt.Errorf("la radio non può memorizzare uno scostamento di %d Hz", offs)
	}
	if offs/100 > 999999 {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente.
	// Come CHIRP: un canale che c'era già viene prima azzerato (il modo
	// digitale resta), uno nuovo si scrive sopra quello che c'era nella posizione
	if was {
		r.Fill(img, 0x00)
		r.F("isdigital").SetBool(img, digital)
	}
	f.setUsed(m, img, i, true)
	r.F("freq").SetInt(img, c.Rx/100)
	r.F("rtone").SetInt(img, rtone)
	r.F("ctone").SetInt(img, ctone)
	r.F("duplex").SetInt(img, dup)
	r.F("offset").SetInt(img, offs/100)
	r.F("tmode").SetInt(img, []int64{0, 1, 3, 12}[tmode])
	r.F("dtcs_tx").SetInt(img, dcs)
	r.F("dtcs_rx").SetInt(img, dcs)
	r.F("step").SetInt(img, step)
	r.F("isnarrow").SetBool(img, c.Narrow)
	switch {
	case o.dj175:
		r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	case len(m.Powers) > 1:
		r.F("ishigh").SetBool(img, m.powerIndex(c.Power) == 1)
	}
	if !o.dj175 {
		f.setBit(f.mem().F("skips"), img, i, !c.Scan)
	}
	name := make([]byte, 7)
	for k, ch := range []byte(strings.TrimRight(m.cleanName(c.Name), " ")) {
		if p := strings.IndexByte(alincoChars, ch); p >= 0 && k < len(name) {
			name[k] = byte(0x30 + p)
		}
	}
	r.F("name").SetRaw(img, name)
	return nil
}

// ---------------------------------------------------------------- protocollo

// alSess: collegamento con una Alinco. Le risposte si leggono a righe; i
// ritorni a capo prima di una riga si saltano
type alSess struct {
	*rawSess
	echo  int  // 0 non ancora noto, 1 il cavo rimanda i dati, 2 non li rimanda
	wrong bool // un'eco è tornata diversa da quello che era stato inviato
}

// line: prossima riga non vuota (senza ritorno a capo); nil se non arriva
// entro il tempo indicato
func (s *alSess) line(to time.Duration) []byte {
	deadline := time.Now().Add(to)
	var out []byte
	b := make([]byte, 1)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return out
		}
		if err := s.p.ReadFull(s.ctx, b, left); err != nil {
			return out
		}
		if b[0] == '\r' || b[0] == '\n' {
			if len(out) > 0 {
				return out
			}
			continue
		}
		out = append(out, b[0])
	}
}

// cmd: invia un comando; restituisce la riga di risposta (nil se non ne
// arrivano). answer falso: comando senza risposta (si legge solo l'eco).
//
// Con un cavo che rimanda i dati, prima dell'eco possono arrivare righe rimaste
// indietro (una radio che conferma con «OK» anche i blocchi a cui CHIRP non
// chiede conferma): si saltano, come fa CHIRP senza saperlo. Se l'eco non
// arriva o arriva diversa, wrong diventa vero
func (s *alSess) cmd(cmd string, answer bool, to time.Duration) ([]byte, error) {
	if err := s.send([]byte(cmd)); err != nil {
		return nil, err
	}
	sent := []byte(strings.TrimRight(cmd, "\r\n"))
	var first []byte
	switch s.echo {
	case 0:
		first = s.line(to)
		switch {
		case bytes.Equal(first, sent):
			s.echo, first = 1, nil
		case len(first) > 0:
			s.echo = 2 // cavo senza eco: questa è già la risposta
		default:
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
			return nil, nil // non arriva niente, nemmeno l'eco
		}
	case 1:
		found := false
		var other []byte
		for k := 0; k < 4 && !found; k++ {
			l := s.line(to)
			if s.ctx.Err() != nil {
				return nil, errCancelled
			}
			if len(l) == 0 {
				break
			}
			if found = bytes.Equal(l, sent); !found {
				other = l
			}
		}
		if !found {
			// eco mancante o rovinata; l'ultima riga arrivata può essere la risposta
			s.wrong, first = true, other
		}
	}
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if !answer {
		return nil, nil
	}
	if len(first) > 0 {
		return first, nil
	}
	l := s.line(to)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	return l, nil
}

func alOpen(ctx context.Context, port string, baud int) (*alSess, error) {
	s, err := rawOpen(ctx, port, baud)
	if err != nil {
		return nil, err
	}
	return &alSess{rawSess: s}, nil
}

// hello: la radio risponde «OK» al proprio nome. Come CHIRP: fino a tre
// tentativi a un secondo l'uno dall'altro (uno solo durante «Identifica»)
func (s *alSess) hello(ident string, tries int) error {
	var last []byte
	for k := 0; k < tries; k++ {
		if k > 0 {
			if err := s.sleep(time.Second); err != nil {
				return err
			}
		}
		l, err := s.cmd(ident+"\r\n", true, time.Second)
		if err != nil {
			return err
		}
		if string(l) == "OK" {
			return nil
		}
		last = l
	}
	if len(last) == 0 {
		return errTimeout
	}
	return rErr("la radio non ha riconosciuto il modello (risposta «%s»): controlla di aver scelto la radio giusta", cleanText(last))
}

// cleanText: testo ricevuto dalla radio, solo caratteri stampabili
func cleanText(b []byte) string {
	var out []byte
	for _, c := range b {
		if c >= 0x20 && c < 0x7F && len(out) < 40 {
			out = append(out, c)
		}
	}
	return string(out)
}

func alTries(port string) int {
	if identProbing(port) {
		return 1
	}
	return 3
}

// alHex: dati di una risposta (cifre esadecimali, n byte)
func alHex(l []byte, n int) ([]byte, bool) {
	d, err := hex.DecodeString(string(l))
	return d, err == nil && len(d) == n
}

// end: fine della copia («AL~E»): la radio torna al funzionamento normale
func (s *alSess) end() {
	_, _ = s.cmd("AL~E\r\n", false, 500*time.Millisecond)
	readUpTo(s.ctx, s.p, 20, 300*time.Millisecond) // la risposta non serve (come CHIRP)
}

func (f alincoFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := alincoO(m)
	s, err := alOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.hello(o.ident, alTries(port)); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	data := make([]byte, 0, o.size)
	for a := 0; a < o.size; a += 16 {
		prog(a, o.size, "Lettura…")
		if err := s.sleep(alincoGap); err != nil {
			return nil, "", err
		}
		l, err := s.cmd(fmt.Sprintf("AL~F%04XR\r\n", a), true, time.Second)
		if err != nil {
			return nil, "", err
		}
		if len(l) == 0 {
			s.end()
			return nil, "", genFail(fmt.Sprintf("Lettura del blocco 0x%04X", a), errTimeout)
		}
		// «indirizzo:dati» (DJ-175: solo i dati)
		if k := bytes.IndexByte(l, ':'); k >= 0 {
			if !strings.EqualFold(string(l[:k]), fmt.Sprintf("%04X", a)) {
				s.end()
				return nil, "", rErr("lettura del blocco 0x%04X: la radio ha risposto per un altro indirizzo (%s)", a, cleanText(l[:k]))
			}
			l = l[k+1:]
		}
		d, ok := alHex(l, 16)
		if !ok {
			s.end()
			return nil, "", rErr("lettura del blocco 0x%04X non riuscita (risposta «%s»)", a, cleanText(l))
		}
		data = append(data, d...)
	}
	s.end()
	prog(o.size, o.size, "Lettura completata")
	return data, "", nil
}

func (f alincoFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := alincoO(m)
	if len(img) < o.size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := alOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.hello(o.ident, 3); err != nil {
		return "", genFail("Identificazione", err)
	}
	// come CHIRP: da 0x100 in poi (i primi 256 byte non si scrivono)
	for a := 0x100; a < o.size; a += 16 {
		prog(a, o.size, "Scrittura…")
		if err := s.sleep(alincoGap); err != nil {
			return "", err
		}
		if _, err := s.cmd(fmt.Sprintf("AL~F%04XW%s\r\n", a, strings.ToUpper(hex.EncodeToString(img[a:a+16]))), false, time.Second); err != nil {
			return "", err
		}
		if s.wrong {
			s.end()
			return "", rErr("scrittura del blocco 0x%04X: i dati tornati dal cavo sono diversi da quelli inviati (disturbi o cavo non adatto). Ripeti la scrittura", a)
		}
	}
	s.end()
	prog(o.size, o.size, "Scrittura completata")
	return "", nil
}

// ---------------------------------------------------------------- DJ-G7

type alg7Fam struct{}

const alg7Format = `
#seekto 0x200;
ul16 bank[50];
ul16 special_bank[7];
#seekto 0x1200;
struct {
  u8 empty;
  ul32 freq;
  u8 mode;
  u8 step;
  ul32 offset;
  u8 duplex;
  u8 squelch_type;
  u8 tx_tone;
  u8 rx_tone;
  u8 dcs;
  ul24 unknown1;
  u8 skip;
  ul32 unknown2;
  ul32 unknown3;
  ul32 unknown4;
  char name[32];
} memory[1000];
`

const alg7Size = 0x1A7C0

func init() {
	mk := func(id, model, rclass string) *genModel {
		return &genModel{ID: id, Vendor: "Alinco", Model: model, RClass: rclass, Name: "Alinco " + model,
			Slots: 1000, First: 0, NameMax: 16, Powers: []genPower{{"Unica", 5, 0}},
			Bands: []band{{500000, 1300000000}}, Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, OneDCS: true,
			Baud: 57600, Fam: alg7Fam{}, Opt: &alg7Opt{ident: "AL~" + model},
			Notes: "1000 canali (da 0 a 999) con nomi di 16 caratteri. Serve il firmware 4_10 o successivo. Per la copia: spegni la radio, collega il cavo, " +
				"riaccendila e, tenendo premuto [MONI], premi e rilascia tre volte [PTT]. 57600 baud (se la radio è impostata a 19200 il programma se ne accorge; " +
				"si cambia girando la manopola con [MONI] premuto). La potenza non è memorizzata nel canale. Il modo segue la frequenza: AM in banda aerea, " +
				"FM larga per la radiodiffusione. DCS con lo stesso codice nei due versi, mai invertito. Driver di CHIRP."}
	}
	genRegister(mk("al-djg7eg", "DJ-G7EG", "AlincoDJG7EG"), mk("al-djg7t", "DJ-G7T", "AlincoDJG7T"))
}

type alg7Opt struct{ ident string }

func (alg7Fam) mem() *bwEl                 { return bwMust(alg7Format) }
func (f alg7Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f alg7Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }

func (f alg7Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}

func (f alg7Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) == f.rec(i).ByteLen() {
		f.rec(i).SetRaw(img, raw)
	}
}

// Match: i file si riconoscono dai metadati (come in CHIRP)
func (alg7Fam) Match(m *genModel, img []byte) bool { return false }

func (alg7Fam) toneOK(t, r genTone) bool { return alincoToneMode(t, r, true) >= 0 }
func (alg7Fam) fitBySet()                {}

func (f alg7Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !f.toneOK(t, r) || alincoFar(c)
}

// alg7Flag: segno di canale occupato (come CHIRP): 2 se il canale si vede anche
// sulla banda principale (FM o FM stretta nelle bande dei radioamatori), 1 se no
func alg7Flag(hz int64, mode int64) int64 {
	if mode > 1 {
		return 1
	}
	if hz >= 136000000 && hz < 174000000 || hz >= 400000000 && hz < 470000000 || hz >= 1240000000 && hz < 1300000000 {
		return 2
	}
	return 1
}

// alg7Auto: modo che segue la frequenza (2 AM, 3 FM larga); 0 se nessuno
func alg7Auto(hz int64) int64 {
	switch {
	case hz >= 76000000 && hz < 108000000:
		return 3
	case hz >= 108000000 && hz < 137000000:
		return 2
	}
	return 0
}

func (f alg7Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if r.F("empty").Int(img) == 0 {
		return nil
	}
	c := &Chan{Rx: r.F("freq").Int(img)}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx = c.Rx
	switch offs := r.F("offset").Int(img); r.F("duplex").Int(img) {
	case 1:
		c.Tx = c.Rx + offs
	case 2:
		c.Tx = c.Rx - offs
	}
	tone := func(name string) string {
		if k := r.F(name).Int(img); k >= 1 && k <= int64(len(alincoTones)) {
			return fmtTone(alincoTones[k-1])
		}
		return ""
	}
	switch r.F("squelch_type").Int(img) {
	case 2:
		c.TxTone = tone("tx_tone")
	case 3:
		c.TxTone, c.RxTone = tone("tx_tone"), tone("rx_tone")
	case 5:
		if k := r.F("dcs").Int(img); k < int64(len(dtcsCodes)) {
			c.TxTone = fmt.Sprintf("D%03dN", dtcsCodes[k])
			c.RxTone = c.TxTone
		}
	}
	c.Narrow = r.F("mode").Int(img) == 0
	c.Scan = r.F("skip").Int(img) == 0
	raw := r.F("name").Raw(img)
	var name []byte
	for k := 0; k < len(raw); k += 2 {
		name = append(name, raw[k])
	}
	c.Name = strings.TrimRight(string(name), "\x00 ")
	return c
}

func (f alg7Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.Fill(img, 0xFF) // come CHIRP
		r.F("empty").SetInt(img, 0)
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := alincoToneMode(t, rt, true)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (subtono in trasmissione o nei due versi, lo stesso codice DCS nei due versi; mai codici invertiti)")
	}
	if c.Rx > 0xFFFFFFFF {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// canale nuovo: valori di CHIRP (subtoni 88,5 Hz, DCS 023, passo 5 kHz, scostamento 600 kHz)
	was := r.F("empty").Int(img) != 0
	txt, rxt, dcs, step, offs, sq := int64(alincoToneDef+1), int64(alincoToneDef+1), int64(0), int64(0), int64(600000), int64(0)
	oldMode, oldRx := int64(1), int64(0)
	if was {
		txt, rxt, dcs = r.F("tx_tone").Int(img), r.F("rx_tone").Int(img), r.F("dcs").Int(img)
		step, offs, sq = r.F("step").Int(img), r.F("offset").Int(img), r.F("squelch_type").Int(img)
		oldMode, oldRx = r.F("mode").Int(img), r.F("freq").Int(img)
		if txt < 1 || txt > int64(len(alincoTones)) {
			txt = alincoToneDef + 1
		}
		if rxt < 1 || rxt > int64(len(alincoTones)) {
			rxt = alincoToneDef + 1
		}
		if dcs >= int64(len(dtcsCodes)) {
			dcs = 0
		}
		if step > 15 {
			step = 0
		}
	}
	switch tmode {
	case 0:
		// squelch «al contrario» (4) e tipo 1, che qui non si vedono: restano
		if sq != 1 && sq != 4 {
			sq = 0
		}
	case 1:
		sq, txt = 2, icomIndex(alincoTones, t.Hz10)+1
	case 2:
		sq, txt = 3, icomIndex(alincoTones, t.Hz10)+1
		rxt = txt
	case 3:
		sq, dcs = 5, icomIndex(dtcsCodes, t.Code)
	case 4:
		sq, txt, rxt = 3, icomIndex(alincoTones, t.Hz10)+1, icomIndex(alincoTones, rt.Hz10)+1
	}
	dup := int64(0)
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 1, d
	case d < 0:
		dup, offs = 2, -d
	}
	if offs > 0xFFFFFFFF {
		return fmt.Errorf("scostamento fuori dai limiti della radio")
	}
	mode := int64(1)
	switch {
	case c.Narrow:
		mode = 0
	case was && oldRx == c.Rx && oldMode > 1 && oldMode < 4:
		mode = oldMode
	case alg7Auto(c.Rx) != 0 && !(was && oldRx == c.Rx):
		mode = alg7Auto(c.Rx)
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	r.F("empty").SetInt(img, alg7Flag(c.Rx, mode))
	r.F("freq").SetInt(img, c.Rx)
	r.F("mode").SetInt(img, mode)
	r.F("step").SetInt(img, step)
	r.F("offset").SetInt(img, offs)
	r.F("duplex").SetInt(img, dup)
	r.F("squelch_type").SetInt(img, sq)
	r.F("tx_tone").SetInt(img, txt)
	r.F("rx_tone").SetInt(img, rxt)
	r.F("dcs").SetInt(img, dcs)
	r.F("skip").SetBool(img, !c.Scan)
	name := make([]byte, 32)
	for k, ch := range []byte(strings.TrimRight(m.cleanName(c.Name), " ")) {
		if k < 16 {
			name[2*k] = ch
		}
	}
	r.F("name").SetRaw(img, name)
	// come CHIRP (senza questi valori il programma di Alinco non accetta il canale)
	r.F("unknown1").SetInt(img, 0x3E001C)
	r.F("unknown2").SetInt(img, 0x0A)
	r.F("unknown3").SetInt(img, 0)
	r.F("unknown4").SetInt(img, 0)
	return nil
}

// alg7Hello: 57600 baud, poi 19200 (la radio può essere impostata così)
func (s *alSess) alg7Hello(m *genModel, tries int) error {
	id := m.Opt.(*alg7Opt).ident
	err := s.hello(id, tries)
	if err == nil || err == errCancelled || s.ctx.Err() != nil {
		return err
	}
	if e := s.p.SetBaud(19200); e != nil {
		return e
	}
	s.echo = 0
	if e := s.hello(id, tries); e != nil {
		return e
	}
	return nil
}

func (f alg7Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := alOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, alg7Size, "Collegamento con la radio…")
	if err := s.alg7Hello(m, alTries(port)); err != nil {
		return nil, "", genFail("Identificazione", err)
	}
	// come CHIRP: i primi 0x200 byte non si leggono (nell'immagine sono a zero)
	data := make([]byte, 0x200, alg7Size)
	for a := 0x200; a < alg7Size; a += 0x40 {
		prog(a, alg7Size, "Lettura…")
		l, err := s.cmd(fmt.Sprintf("AL~F%05XR\r", a), true, time.Second)
		if err != nil {
			return nil, "", err
		}
		if len(l) == 0 {
			return nil, "", genFail(fmt.Sprintf("Lettura del blocco 0x%05X", a), errTimeout)
		}
		d, ok := alHex(l, 0x40)
		if !ok {
			return nil, "", rErr("lettura del blocco 0x%05X non riuscita (risposta «%s»)", a, cleanText(l))
		}
		data = append(data, d...)
	}
	prog(alg7Size, alg7Size, "Lettura completata")
	return data, "", nil
}

func (f alg7Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < alg7Size {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte)", len(img))
	}
	s, err := alOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, alg7Size, "Collegamento con la radio…")
	if err := s.alg7Hello(m, 3); err != nil {
		return "", genFail("Identificazione", err)
	}
	for a := 0x200; a < alg7Size; a += 0x40 {
		prog(a, alg7Size, "Scrittura…")
		l, err := s.cmd(fmt.Sprintf("AL~F%05XW%s\r", a, strings.ToUpper(hex.EncodeToString(img[a:a+0x40]))), true, 2*time.Second)
		if err != nil {
			return "", err
		}
		if len(l) == 0 {
			return "", genFail(fmt.Sprintf("Scrittura del blocco 0x%05X", a), errTimeout)
		}
		if string(l) != "OK" {
			return "", rErr("la radio non ha accettato il blocco 0x%05X (risposta «%s»)", a, cleanText(l))
		}
	}
	prog(alg7Size, alg7Size, "Scrittura completata")
	return "", nil
}
