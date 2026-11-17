// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
//
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ic9x.py, ic9x_ll.py (4.41).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.41: Icom IC-91A / IC-91AD / IC-92AD (in CHIRP «IC-91/92AD», modo «live»).
// La radio accesa normalmente, col cavo dati del costruttore, si programma un
// canale per volta. Le due bande sono due radio distinte (come in CHIRP):
//
//	banda A: 800 memorie (0-799), 50 limiti di scansione (800-849), 2 canali di
//	         chiamata; FM, FM larga, AM
//	banda B: 400 memorie (0-399), 50 limiti di scansione (400-449), 2 canali di
//	         chiamata; FM, FM stretta, AM e D-STAR (DV)
//
// Il collegamento: a 4800 baud 160 byte FE e «FE FE 01 80 19 FD»; la radio
// risponde con «80 01 19» e da lì si parla a 38400 baud. Dopo 4 secondi senza
// comandi il risveglio va ripetuto. Ogni comando è «FE FE banda 80 1A 00 dati
// FD» (banda: 1 = A, 2 = B), la risposta «FE FE 80 banda …»:
//
//	lettura        dati = tipo (1 memoria, 2 canale di chiamata) e numero in
//	               BCD (2 byte); risposta: gli stessi 3 byte e il canale (34
//	               byte in tutto nella banda A, 60 nella B) oppure FF = vuoto
//	scrittura      dati = il canale; risposta «FE FE 80 banda FB FD»
//	cancellazione  dati = tipo, numero, FF; risposta FB
//
// Il canale: tipo, numero (BCD), frequenza in Hz (5 byte BCD dal meno
// significativo), scostamento in Hz (4 byte), 1 byte, subtono in trasmissione
// e subtono in ricezione in decimi di Hz (2 byte BCD), codice DCS (2 byte
// BCD), 2 byte, modo e passo, tipo di tono e verso dello scostamento, polarità
// del DCS e salti, banco (lettera) e posizione nel banco, nome (8 caratteri);
// nella banda B anche 1 byte, il codice digitale e i nominativi RPT2, RPT1, UR.
//
// L'immagine del programma (formato di IU0VYT Flash, CHIRP non ha file per
// queste radio): 32 byte di intestazione e 64 byte per posizione (stato,
// lunghezza e il canale come lo dà la radio). In scrittura la radio viene
// riletta e si inviano solo le posizioni cambiate.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	ic9xHead   = 32
	ic9xRec    = 64
	ic9xMagic  = "IU0VYT-ICOM-IC9X\x00"
	ic9xBandAt = 24
	ic9xLenA   = 34
	ic9xLenB   = 60
)

// tempo dopo il quale la radio vuole di nuovo il risveglio (come CHIRP)
var ic9xStale = 4 * time.Second

// passi di sintonia della radio in Hz (8,33 kHz = 25/3)
var ic9xSteps = []int64{5000, 6250, 8333, 9000, 10000, 12500, 15000, 20000, 25000, 30000, 50000, 100000, 125000, 200000}

// canale nuovo: il blocco valido che usa CHIRP (146,010 MHz FM, 600 kHz, 88,5
// Hz, DCS 023), con il primo ripetitore vuoto al posto del nominativo che
// CHIRP ha nel suo esempio
var ic9xTemplate = append([]byte{
	0x01, 0x00, 0x03, 0x00, 0x00, 0x01, 0x46, 0x01,
	0x00, 0x00, 0x60, 0x00, 0x00, 0x08, 0x85, 0x08,
	0x85, 0x00, 0x23, 0x22, 0x80, 0x06, 0x00, 0x00,
	0x00, 0x00, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x00, 0x00}, []byte("                CQCQCQ  ")...)

type ic9xFam struct{}

type ic9xOpt struct {
	band   byte // 1 = A, 2 = B
	normal int  // memorie normali più limiti di scansione; poi i due canali di chiamata
	mem    int  // memorie normali
}

func ic9xo(m *genModel) *ic9xOpt { return m.Opt.(*ic9xOpt) }

const ic9xCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func init() {
	const how = "Si programma con la radio accesa normalmente (non in modalità di copia) e il cavo dati del costruttore: il programma " +
		"legge e scrive un canale per volta. Prima di scrivere la radio viene riletta e si inviano solo le posizioni cambiate. " +
		"Le due bande si leggono e si scrivono separatamente (come in CHIRP): scegli la banda nell'elenco delle radio. "
	const keep = "Passo, modo e banco di un canale restano quelli impostati finché la frequenza non cambia; un canale nuovo prende AM in banda aerea " +
		"e sotto i 30 MHz. I banchi e gli elenchi dei nominativi non si modificano da qui. Comandi del driver di CHIRP, provati con una radio simulata."
	one := []genPower{{"Unica", 5, 0}}
	genRegister(&genModel{ID: "ic-9x-a", Vendor: "Icom", Model: "IC-91/92AD", Variant: "Banda A", RClass: "IU0VYT_Icom_IC9xA",
		Name: "Icom IC-91AD / IC-92AD (banda A)", Slots: 852, First: 0, NameMax: 8, Charset: ic9xCharset, Powers: one,
		Bands: []band{{495000, 999990000}}, HasScan: true, NoTxOff: true, NoChirp: true, DCS: dtcsCodes, Baud: 4800,
		Fam: ic9xFam{}, Opt: &ic9xOpt{band: 1, normal: 850, mem: 800},
		Notes: "Banda A (ricevitore a larga banda): posizioni 0-799 memorie, 800-849 limiti di scansione (0A, 0B … 24A, 24B), 850 e 851 canali di chiamata. " +
			"Nomi di 8 caratteri; FM, FM larga (da 76 a 108 MHz) e AM. " + how + keep})
	genRegister(&genModel{ID: "ic-9x-b", Vendor: "Icom", Model: "IC-91/92AD", Variant: "Banda B", RClass: "IU0VYT_Icom_IC9xB",
		Name: "Icom IC-91AD / IC-92AD (banda B)", Slots: 452, First: 0, NameMax: 8, Charset: ic9xCharset, Powers: one,
		Bands: []band{{118000000, 174000000}, {350000000, 470000000}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
		DCS: dtcsCodes, Baud: 4800, Digital: "D-STAR", Fam: ic9xFam{}, Opt: &ic9xOpt{band: 2, normal: 450, mem: 400},
		Notes: "Banda B (144/430 MHz, anche D-STAR): posizioni 0-399 memorie, 400-449 limiti di scansione, 450 e 451 canali di chiamata. " +
			"Nomi di 8 caratteri; canali analogici (FM, FM stretta, AM) o D-STAR: dei canali D-STAR si conservano i nominativi (UR, RPT1, RPT2) " +
			"e il codice digitale, un canale D-STAR nuovo nasce con «CQCQCQ». " + how + keep})
}

// noIdent: «Identifica» non prova queste radio (il risveglio sono 166 byte a
// 4800 baud, che a un'altra radio non vanno mandati)
func (ic9xFam) noIdent() {}

func ic9xSize(m *genModel) int { return ic9xHead + m.Slots*ic9xRec }

func ic9xBlank(m *genModel) []byte {
	img := make([]byte, ic9xSize(m))
	copy(img, ic9xMagic)
	img[ic9xBandAt] = ic9xo(m).band
	return img
}

func (ic9xFam) rec(img []byte, i int) []byte { return img[ic9xHead+i*ic9xRec : ic9xHead+(i+1)*ic9xRec] }

func (f ic9xFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}
func (f ic9xFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (ic9xFam) Match(m *genModel, img []byte) bool {
	return len(img) == ic9xSize(m) && bytes.HasPrefix(img, []byte(ic9xMagic)) && img[ic9xBandAt] == ic9xo(m).band
}

// ic9xLen: lunghezza del canale nella banda
func ic9xLen(band byte) int {
	if band == 1 {
		return ic9xLenA
	}
	return ic9xLenB
}

// ic9xAddr: tipo (1 memoria, 2 canale di chiamata) e numero della posizione i
func ic9xAddr(m *genModel, i int) (kind byte, num int) {
	if o := ic9xo(m); i >= o.normal {
		return 2, i - o.normal
	}
	return 1, i
}

// special: limiti di scansione e canali di chiamata (senza salto di scansione
// né banco, come in CHIRP)
func ic9xSpecial(m *genModel, i int) bool { return i >= ic9xo(m).mem }

// payload: il canale della posizione i come lo dà la radio (nil se vuota o
// non valida)
func (f ic9xFam) payload(m *genModel, img []byte, i int) []byte {
	r := f.rec(img, i)
	if r[0] != 1 || int(r[1]) != ic9xLen(ic9xo(m).band) {
		return nil
	}
	return r[2 : 2+int(r[1])]
}

// ic9xValid: i campi del canale sono leggibili (cifre BCD, indici nei limiti)
func ic9xValid(p []byte, band byte) bool {
	if len(p) != ic9xLen(band) || (p[0] != 1 && p[0] != 2) {
		return false
	}
	for _, r := range [][2]int{{1, 3}, {3, 8}, {8, 12}, {13, 15}, {15, 17}, {17, 19}} {
		for _, b := range p[r[0]:r[1]] {
			if b>>4 > 9 || b&0x0F > 9 {
				return false
			}
		}
	}
	mode, step, tmode, dup := p[21]>>4&7, p[21]&0x0F, p[22]>>2&7, p[22]&3
	okMode := mode == 0 || mode == 3 || (band == 1 && mode == 2) || (band == 2 && (mode == 1 || mode == 4))
	return okMode && int(step) < len(ic9xSteps) && tmode <= 5 && dup <= 2
}

func (f ic9xFam) Get(m *genModel, img []byte, i int) *Chan {
	p := f.payload(m, img, i)
	if p == nil || !ic9xValid(p, ic9xo(m).band) {
		return nil
	}
	rx, _ := bcdLE(p[3:8])
	offs, _ := bcdLE(p[8:12])
	if rx <= 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx, Scan: true}
	switch p[22] & 3 {
	case 1:
		c.Tx = rx - offs
	case 2:
		c.Tx = rx + offs
	}
	if c.Tx <= 0 {
		c.Tx = rx
	}
	mode := p[21] >> 4 & 7
	c.Narrow, c.Dig = mode == 1, mode == 4
	if !c.Dig {
		rt, _ := bcdBE(p[13:15])
		ct, _ := bcdBE(p[15:17])
		dc, _ := bcdBE(p[17:19])
		switch p[22] >> 2 & 7 {
		case 1:
			c.TxTone = fmtTone(rt)
		case 2, 3:
			c.TxTone, c.RxTone = fmtTone(ct), fmtTone(ct)
		case 4, 5:
			pol := p[23] >> 2 & 3
			c.TxTone = fmt.Sprintf("D%03d%c", dc, "NI"[pol>>1&1])
			c.RxTone = fmt.Sprintf("D%03d%c", dc, "NI"[pol&1])
		}
	}
	if !ic9xSpecial(m, i) {
		c.Scan = p[23]&3 == 0
	}
	c.Name = strings.TrimRight(string(p[26:34]), " ")
	return c
}

// ic9xStepFor: passo di sintonia per la frequenza (quello che c'era se la
// divide, altrimenti 5, 6,25, 9 o 8,33 kHz); -1 se la radio non può averla
func ic9xStepFor(hz int64, old int) int {
	ok := func(k int) bool {
		if k == 2 {
			r := hz * 3 % 25000
			return r <= 2 || r >= 24998
		}
		return hz%ic9xSteps[k] == 0
	}
	if old >= 0 && old < len(ic9xSteps) && ok(old) {
		return old
	}
	for _, k := range []int{0, 1, 3, 2} {
		if ok(k) {
			return k
		}
	}
	return -1
}

func (f ic9xFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	return icomToneMode(t, r) < 0 || d > 99999999 || d < -99999999 || ic9xStepFor(c.Rx, -1) < 0 || !m.inBands(c.Rx)
}

func (ic9xFam) toneOK(t, r genTone) bool { return icomToneMode(t, r) >= 0 }

// imgSize, fitBySet: nelle conversioni da un'altra radio il canale si prova con
// Set (passo della frequenza, scostamento), vedi genAdapt
func (ic9xFam) imgSize(m *genModel) int { return ic9xSize(m) }
func (ic9xFam) fitBySet()               {}

func (f ic9xFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := ic9xo(m)
	r := f.rec(img, i)
	if c == nil {
		for k := range r {
			r[k] = 0
		}
		return nil
	}
	if c.Dig && o.band != 2 {
		return fmt.Errorf("il D-STAR è solo nella banda B della radio")
	}
	old := f.payload(m, img, i)
	if old != nil && !ic9xValid(old, o.band) {
		old = nil
	}
	p := append([]byte{}, ic9xTemplate[:ic9xLen(o.band)]...)
	oldRx, oldMode, oldStep := int64(0), byte(0), -1
	if old != nil {
		copy(p, old)
		oldRx, _ = bcdLE(old[3:8])
		oldMode, oldStep = old[21]>>4&7, int(old[21]&0x0F)
	}
	if c.Rx <= 0 || !m.inBands(c.Rx) {
		if o.band == 2 {
			return fmt.Errorf("frequenza %s MHz fuori dalla banda B della radio (118-174 e 350-470 MHz): usa la banda A", fmtMHz(c.Rx))
		}
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	step := ic9xStepFor(c.Rx, oldStep)
	if step < 0 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (multipli di 5, di 6,25, di 8,33 o di 9 kHz)", fmtMHz(c.Rx))
	}
	// lo scostamento di un canale senza scostamento resta quello che c'era (come in CHIRP)
	dup, offs := byte(0), int64(-1)
	switch d := c.Tx - c.Rx; {
	case d > 0:
		dup, offs = 2, d
	case d < 0:
		dup, offs = 1, -d
	}
	if offs > 99999999 {
		return fmt.Errorf("scostamento fuori dai limiti della radio (meno di 100 MHz)")
	}
	tmode, pol := byte(0), byte(0)
	rt, _ := bcdBE(p[13:15])
	ct, _ := bcdBE(p[15:17])
	dc, _ := bcdBE(p[17:19])
	if !c.Dig {
		t, err := genToneParse(c.TxTone)
		if err != nil {
			return err
		}
		rx, err := genToneParse(c.RxTone)
		if err != nil {
			return err
		}
		switch icomToneMode(t, rx) {
		case -1:
			return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione, lo stesso subtono nei due versi, lo stesso codice DCS nei due versi)")
		case 1:
			tmode, rt = 1, t.Hz10
		case 2:
			// come CHIRP: il subtono va nei due campi
			tmode, rt, ct = 2, t.Hz10, t.Hz10
		case 3:
			tmode, dc = 4, t.Code
			if t.Inv {
				pol |= 2
			}
			if rx.Inv {
				pol |= 1
			}
		}
		if (tmode == 1 || tmode == 2) && icomIndex(chirpTones, rt) < 0 || tmode == 4 && icomIndex(dtcsCodes, dc) < 0 {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	// modo: resta se la frequenza non cambia, altrimenti segue la frequenza
	mode := byte(0)
	same := old != nil && oldRx == c.Rx
	narrow := c.Narrow && o.band == 2 // la FM stretta è solo nella banda B
	switch {
	case c.Dig:
		mode = 4
	case same && oldMode != 4 && !narrow:
		if mode = oldMode; mode == 1 {
			mode = 0 // da FM stretta a FM
		}
	case c.Rx >= 118000000 && c.Rx < 137000000 && !narrow, c.Rx < 30000000:
		mode = 3
	case o.band == 1 && c.Rx >= 76000000 && c.Rx < 108000000:
		mode = 2
	case narrow:
		mode = 1
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	kind, num := ic9xAddr(m, i)
	p[0] = kind
	putBCDBE(p[1:3], num)
	putBCDLE(p[3:8], c.Rx)
	if offs >= 0 {
		putBCDLE(p[8:12], offs)
	}
	putBCDBE(p[13:15], rt)
	putBCDBE(p[15:17], ct)
	putBCDBE(p[17:19], dc)
	p[21] = p[21]&0x80 | mode<<4 | byte(step)
	p[22] = p[22]&0xE0 | tmode<<2 | dup
	flags := p[23] & 0xF0
	if tmode == 4 {
		flags |= pol << 2
	}
	if !ic9xSpecial(m, i) {
		switch {
		case c.Scan:
		case old != nil && old[23]&3 == 2:
			flags |= 2 // il salto «P» resta
		default:
			flags |= 1
		}
	}
	p[23] = flags
	copy(p[26:34], fmt.Sprintf("%-8.8s", m.cleanName(c.Name)))
	if o.band == 2 && c.Dig && !(old != nil && oldMode == 4) {
		// canale D-STAR nuovo: chiamata generale, nessun ripetitore
		p[35] = 0
		copy(p[36:], "                CQCQCQ  ")
	}
	for k := range r {
		r[k] = 0
	}
	r[0], r[1] = 1, byte(len(p))
	copy(r[2:], p)
	return nil
}

// ---------------------------------------------------------------- collegamento

type ic9xSess struct {
	*rawSess
	band byte
	last time.Time
}

func ic9xFrame(band byte, data []byte) []byte {
	return append(append([]byte{0xFE, 0xFE, band, 0x80, 0x1A, 0x00}, data...), 0xFD)
}

// wake: il risveglio, come CHIRP (cinque tentativi; i tempi sono i suoi)
func (s *ic9xSess) wake() error {
	if err := s.p.SetBaud(4800); err != nil {
		return err
	}
	buf := append(bytes.Repeat([]byte{0xFE}, 160), 0xFE, 0xFE, 0x01, 0x80, 0x19, 0xFD)
	for try := 0; try < 5; try++ {
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if err := s.send(buf); err != nil {
			return err
		}
		if err := s.sleep(100 * time.Millisecond); err != nil {
			return err
		}
		// CHIRP legge 5 byte entro un secondo; qui si legge finché arriva la
		// risposta (anche se il cavo rimanda indietro quello che è stato inviato)
		var got []byte
		end := time.Now().Add(time.Second)
		for time.Now().Before(end) && !bytes.Contains(got, []byte{0x80, 0x01, 0x19}) {
			b := readUpTo(s.ctx, s.p, 1, time.Until(end))
			if len(b) == 0 {
				break
			}
			got = append(got, b...)
		}
		if bytes.Contains(got, []byte{0x80, 0x01, 0x19}) {
			if err := s.p.SetBaud(38400); err != nil {
				return err
			}
			// quello che la radio manda dopo il risveglio si scarta (fino a un secondo di silenzio)
			for len(readUpTo(s.ctx, s.p, 64, time.Second)) > 0 {
				if s.ctx.Err() != nil {
					return errCancelled
				}
			}
			s.last = time.Now()
			return nil
		}
		if err := s.sleep(time.Second); err != nil {
			return err
		}
	}
	return genFail("Collegamento", rErr("la radio non risponde: controlla che sia accesa, il cavo dati e la porta"))
}

// recv: il prossimo blocco della radio (senza FE FE e FD), nil se non arriva
// niente entro un secondo
func (s *ic9xSess) recv() []byte {
	var data []byte
	for len(data) < 256 {
		b := readUpTo(s.ctx, s.p, 1, time.Second)
		if len(b) == 0 {
			return nil
		}
		data = append(data, b[0])
		if b[0] == 0xFD {
			break
		}
	}
	k := bytes.Index(data, []byte{0xFE, 0xFE})
	if k < 0 {
		return []byte{}
	}
	for k+2 < len(data) && data[k+2] == 0xFE {
		k++
	}
	return data[k+2 : len(data)-1]
}

// cmd: invia un comando e dà la risposta della radio (banda, 80 tolti: dal
// byte del comando in poi). Come CHIRP salta i blocchi che non c'entrano; in
// più salta l'eco del cavo
func (s *ic9xSess) cmd(data []byte) ([]byte, error) {
	if s.last.IsZero() || time.Since(s.last) > ic9xStale {
		if err := s.wake(); err != nil {
			return nil, err
		}
	}
	if err := s.send(ic9xFrame(s.band, data)); err != nil {
		return nil, err
	}
	for skipped := 0; skipped < 50; skipped++ {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		fr := s.recv()
		if fr == nil {
			return nil, rErr("la radio non risponde")
		}
		if len(fr) < 3 || fr[0] != 0x80 {
			continue // eco del cavo o dati rovinati
		}
		if len(fr) > 4 && fr[4] != data[0] {
			continue // blocco non richiesto (come CHIRP)
		}
		s.last = time.Now()
		return fr[2:], nil
	}
	return nil, rErr("risposta inattesa della radio")
}

// get: il canale della posizione i (nil se vuota)
func (s *ic9xSess) get(m *genModel, i int) ([]byte, error) {
	kind, num := ic9xAddr(m, i)
	req := []byte{kind, 0, 0}
	putBCDBE(req[1:3], num)
	r, err := s.cmd(req)
	if err != nil {
		return nil, err
	}
	if len(r) < 2+4 || r[0] != 0x1A {
		return nil, rErr("risposta inattesa della radio alla lettura della posizione %d", i+m.First)
	}
	p := r[2:]
	if got, ok := bcdBE(p[1:3]); !ok || got != num {
		return nil, rErr("la radio ha risposto con un'altra posizione (%d al posto di %d)", got, num)
	}
	if p[3] == 0xFF {
		return nil, nil
	}
	if len(p) != ic9xLen(s.band) {
		return nil, rErr("la posizione %d ha una lunghezza inattesa (%d byte): la banda scelta è quella giusta?", i+m.First, len(p))
	}
	return p, nil
}

func ic9xOpen(ctx context.Context, m *genModel, port string) (*ic9xSess, error) {
	rs, err := rawOpen(ctx, port, 4800)
	if err != nil {
		return nil, err
	}
	s := &ic9xSess{rawSess: rs, band: ic9xo(m).band}
	if err := s.wake(); err != nil {
		rs.p.Close()
		return nil, err
	}
	return s, nil
}

func (s *ic9xSess) readAll(m *genModel, prog Progress, what string) ([]byte, error) {
	f := ic9xFam{}
	img := ic9xBlank(m)
	for i := 0; i < m.Slots; i++ {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		p, err := s.get(m, i)
		if err != nil {
			return nil, genFail("Lettura", err)
		}
		if p != nil {
			r := f.rec(img, i)
			r[0], r[1] = 1, byte(len(p))
			copy(r[2:], p)
		}
		if i%10 == 0 {
			prog(i, m.Slots, what)
		}
	}
	return img, nil
}

func (f ic9xFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := ic9xOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	img, err := s.readAll(m, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	prog(m.Slots, m.Slots, "Lettura completata")
	return img, "", nil
}

func (f ic9xFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	// prima di collegarsi: ogni posizione dell'immagine dev'essere vuota o un canale valido
	for i := 0; i < m.Slots; i++ {
		if r := f.rec(img, i); r[0] != 0 && (f.payload(m, img, i) == nil || !ic9xValid(f.payload(m, img, i), ic9xo(m).band)) {
			return "", fmt.Errorf("immagine della memoria non valida (posizione %d)", i+m.First)
		}
	}
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := ic9xOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	cur, err := s.readAll(m, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		return "", err
	}
	var todo []int
	for i := 0; i < m.Slots; i++ {
		if !bytes.Equal(f.rec(cur, i), f.rec(img, i)) {
			todo = append(todo, i)
		}
	}
	for k, i := range todo {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(k, len(todo), "Scrittura…")
		data := f.payload(m, img, i)
		if data == nil {
			kind, num := ic9xAddr(m, i)
			data = []byte{kind, 0, 0, 0xFF}
			putBCDBE(data[1:3], num)
		}
		r, err := s.cmd(data)
		if err != nil {
			return "", genFail("Scrittura", err)
		}
		if len(r) < 1 || r[0] != 0xFB {
			return "", rErr("la radio ha rifiutato la posizione %d: i canali già scritti restano, gli altri no", i+m.First)
		}
	}
	prog(len(todo), len(todo), "Scrittura completata")
	return "", nil
}
