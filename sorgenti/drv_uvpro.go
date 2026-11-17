// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
//
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// btech_uvpro.py (4.41).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.41: BTECH UV-PRO (in CHIRP modo «live»). La radio si programma dalla porta
// seriale Bluetooth (SPP) che il PC crea dopo l'abbinamento; la velocità della
// porta non conta. Si leggono e si scrivono i 30 canali del gruppo attivo
// sulla radio; il gruppo non si cambia mai da qui.
//
// Ogni blocco: FF 01, un byte di opzioni (bit 0: in coda c'è un byte di
// controllo), la lunghezza dei dati; poi il gruppo dei comandi (2 byte, sempre
// 2), il comando (2 byte; nella risposta ha il bit 0x8000) e i dati. Il primo
// byte dei dati di una risposta è l'esito (0 = riuscito). I comandi usati:
//
//	 4  informazioni (dati 03): costruttore, prodotto, versioni, numero dei gruppi
//	20  stato: da qui il gruppo attivo
//	73  nome di un gruppo (10 caratteri)
//	13  lettura di un canale (dati: il numero): 25 byte
//	14  scrittura di un canale (dati: i 25 byte)
//
// Il canale (25 byte): numero; trasmissione e ricezione (4 byte ciascuna: 2
// bit di modulazione, 0 FM 1 AM 2 DMR, e 30 bit di frequenza in Hz); tono di
// trasmissione e di ricezione (2 byte: 0 nessuno, sotto 6700 il codice DCS,
// altrimenti il subtono in centesimi di Hz); un byte di opzioni (0x80 in
// scansione, 0x40 potenza alta, 0x10 banda larga, 0x02 potenza media, 0x01
// trasmissione disabilitata); un altro byte di opzioni; il nome (10 caratteri).
//
// Come CHIRP: la scrittura è ammessa solo con la versione del firmware provata
// dall'autore del driver (147); ogni canale viene riletto prima di scriverlo e
// dopo, e se non coincide si rimette quello di prima.
//
// L'immagine del programma (formato di IU0VYT Flash): 64 byte di intestazione
// (le informazioni della radio, il gruppo attivo e il suo nome) e 32 byte per
// canale (stato e i 25 byte della radio).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	uvproHead    = 64
	uvproRec     = 32
	uvproMagic   = "IU0VYT-BTECH-UVPRO\x00"
	uvproInfoAt  = 32 // 9 byte di informazioni
	uvproGroupAt = 41 // gruppo attivo alla lettura (FF = non noto)
	uvproNameAt  = 42 // nome del gruppo (10 caratteri)
	uvproLen     = 25
	uvproSoftVer = 147
	uvproTimeout = 3 * time.Second
)

// stato di una posizione nell'immagine
const (
	uvproOK      = 0
	uvproNone    = 1 // la radio non l'ha data
	uvproForeign = 2 // dati di un altro tipo (lunghezza diversa)
)

type uvproFam struct{}

func init() {
	genRegister(&genModel{ID: "bt-uvpro", Vendor: "BTECH", Model: "UV-PRO", RClass: "IU0VYT_BTECH_UVPRO", Name: "BTECH UV-PRO",
		Slots: 30, First: 0, NameMax: 10,
		Powers: []genPower{{"Bassa 0,5 W", 0.5, 2}, {"Media 2 W", 2, 1}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{108000000, 135999999}, {136000000, 174000000}, {400000000, 520000000}},
		Narrow: true, HasScan: true, NoChirp: true, DCS: dtcsCodes, Baud: 115200, Fam: uvproFam{},
		Notes: "Si programma via Bluetooth, senza cavo: abbina la radio al PC e scegli la porta seriale Bluetooth in uscita della radio («SPP Dev»). " +
			"Si leggono e si scrivono i 30 canali (da 0 a 29) del gruppo attivo sulla radio: per un altro gruppo cambialo sulla radio e rileggi. " +
			"Nomi di 10 caratteri; codici DCS solo normali. Da 108 a 136 MHz la radio riceve soltanto (AM): quei canali si salvano in sola ricezione e senza toni. " +
			"La scrittura è ammessa solo con la versione del firmware provata dall'autore del driver di CHIRP (147): con un'altra versione la radio si legge ma non si scrive. " +
			"Ogni canale scritto viene riletto e, se non coincide, si rimette quello di prima. I canali DMR non si modificano da qui. " +
			"Comandi del driver di CHIRP (sperimentale), provati con una radio simulata."})
}

// noIdent: «Identifica» non prova le porte Bluetooth
func (uvproFam) noIdent() {}

func uvproSize(m *genModel) int { return uvproHead + m.Slots*uvproRec }

// uvproEmpty: posizione vuota (il numero e 24 zeri)
func uvproEmpty(i int) []byte { return append([]byte{byte(i)}, make([]byte, uvproLen-1)...) }

// uvproBlank: immagine con tutte le posizioni vuote, di una radio con il firmware provato
func uvproBlank(m *genModel) []byte {
	img := make([]byte, uvproSize(m))
	copy(img, uvproMagic)
	binary.BigEndian.PutUint16(img[uvproInfoAt+4:], uvproSoftVer)
	img[uvproInfoAt+8] = byte(m.Slots)
	img[uvproGroupAt] = 0xFF
	f := uvproFam{}
	for i := 0; i < m.Slots; i++ {
		copy(f.rec(img, i)[1:], uvproEmpty(i))
	}
	return img
}

func (uvproFam) rec(img []byte, i int) []byte {
	return img[uvproHead+i*uvproRec : uvproHead+(i+1)*uvproRec]
}

func (f uvproFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}
func (f uvproFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (uvproFam) Match(m *genModel, img []byte) bool {
	return len(img) == uvproSize(m) && bytes.HasPrefix(img, []byte(uvproMagic))
}

func (uvproFam) imgSize(m *genModel) int { return uvproSize(m) }
func (uvproFam) fitBySet()               {}

// ch: i 25 byte del canale (nil se la radio non l'ha dato o è di un altro tipo)
func (f uvproFam) ch(img []byte, i int) []byte {
	r := f.rec(img, i)
	if r[0] != uvproOK {
		return nil
	}
	return r[1 : 1+uvproLen]
}

func uvproDMR(p []byte) bool { return p[1]>>6 == 2 || p[5]>>6 == 2 }

// uvproTone: tono dal valore della radio
func uvproTone(raw uint16) string {
	switch {
	case raw == 0:
		return ""
	case raw < 6700:
		return fmt.Sprintf("D%03dN", raw)
	}
	return fmtTone(int(raw+5) / 10)
}

func (f uvproFam) Get(m *genModel, img []byte, i int) *Chan {
	p := f.ch(img, i)
	if p == nil {
		return nil
	}
	rx := int64(binary.BigEndian.Uint32(p[5:9]) & 0x3FFFFFFF)
	tx := int64(binary.BigEndian.Uint32(p[1:5]) & 0x3FFFFFFF)
	if rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: tx}
	if p[13]&0x01 != 0 || tx == 0 {
		c.Tx, c.TxOff = rx, true
	}
	am := p[5]>>6 == 1
	c.Narrow = p[13]&0x10 == 0 && !am
	if !uvproDMR(p) {
		c.TxTone = uvproTone(binary.BigEndian.Uint16(p[9:11]))
		c.RxTone = uvproTone(binary.BigEndian.Uint16(p[11:13]))
	}
	switch {
	case p[13]&0x40 != 0:
		c.Power = 0
	case p[13]&0x02 != 0:
		c.Power = 1
	default:
		c.Power = 2
	}
	c.Scan = p[13]&0x80 != 0
	name := p[15:25]
	if k := bytes.IndexByte(name, 0); k >= 0 {
		name = name[:k]
	}
	c.Name = strings.TrimRight(asciiName(string(name)), " ")
	return c
}

// LockedImg: posizioni che da qui non si modificano
func (f uvproFam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	switch r := f.rec(img, i); {
	case r[0] == uvproNone:
		return fmt.Errorf("la radio non ha dato questa posizione alla lettura: non si può scrivere")
	case r[0] != uvproOK || uvproDMR(r[1:1+uvproLen]):
		return fmt.Errorf("la posizione contiene un canale DMR: si modifica con l'applicazione del costruttore")
	}
	return nil
}

func uvproAir(hz int64) bool { return hz >= 108000000 && hz < 136000000 }

// uvproSub: valore della radio per il tono
func uvproSub(t genTone) (uint16, error) {
	switch t.Mode {
	case "":
		return 0, nil
	case "Tone":
		if icomIndex(chirpTones, t.Hz10) < 0 || t.Hz10 < 670 {
			return 0, fmt.Errorf("tono non disponibile su questa radio")
		}
		return uint16(t.Hz10 * 10), nil
	}
	if t.Inv {
		return 0, fmt.Errorf("la radio memorizza solo i codici DCS normali (N), non quelli invertiti")
	}
	if icomIndex(dtcsCodes, t.Code) < 0 {
		return 0, fmt.Errorf("codice DCS %03d non disponibile su questa radio", t.Code)
	}
	return uint16(t.Code), nil
}

func (f uvproFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return t.Inv || r.Inv
}

func (uvproFam) toneOK(t, r genTone) bool { return !t.Inv && !r.Inv }

func (f uvproFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if err := f.LockedImg(m, img, i, nil, c); err != nil {
		return err
	}
	r := f.rec(img, i)
	if c == nil {
		copy(r[1:], uvproEmpty(i))
		return nil
	}
	air := uvproAir(c.Rx)
	if !m.inBands(c.Rx) {
		return fmt.Errorf("frequenza %s MHz fuori dalle bande della radio (108-136 in sola ricezione, 136-174 e 400-520 MHz)", fmtMHz(c.Rx))
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	txSub, err := uvproSub(t)
	if err != nil {
		return err
	}
	rxSub, err := uvproSub(rt)
	if err != nil {
		return err
	}
	mod, tx, flags := uint32(0), uint32(c.Rx), r[1+13]&0x2C // restano i tre bit che il programma non gestisce
	switch {
	case air:
		// banda aerea: AM, sola ricezione (come la salva la radio: frequenza di
		// trasmissione a zero), banda larga
		if c.Tx != c.Rx && !c.TxOff {
			return fmt.Errorf("da 108 a 136 MHz la radio riceve soltanto: niente scostamento")
		}
		if txSub != 0 || rxSub != 0 {
			return fmt.Errorf("da 108 a 136 MHz la radio riceve in AM: niente toni")
		}
		mod, tx = 1, 0
		flags |= 0x01 | 0x10
	case c.TxOff:
		flags |= 0x01
	default:
		if c.Tx < 136000000 || !m.inBands(c.Tx) {
			return fmt.Errorf("frequenza di trasmissione %s MHz fuori dalle bande della radio (136-174 e 400-520 MHz)", fmtMHz(c.Tx))
		}
		tx = uint32(c.Tx)
	}
	if !air && !c.Narrow {
		flags |= 0x10
	}
	switch m.Powers[m.powerIndex(c.Power)].Sem {
	case 0:
		flags |= 0x40
	case 1:
		flags |= 0x02
	}
	if c.Scan {
		flags |= 0x80
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	p := r[1 : 1+uvproLen]
	p[0] = byte(i)
	binary.BigEndian.PutUint32(p[1:5], mod<<30|tx)
	binary.BigEndian.PutUint32(p[5:9], mod<<30|uint32(c.Rx))
	binary.BigEndian.PutUint16(p[9:11], txSub)
	binary.BigEndian.PutUint16(p[11:13], rxSub)
	p[13] = flags
	name := []byte(asciiName(c.Name))
	if len(name) > 10 {
		name = name[:10]
	}
	copy(p[15:25], make([]byte, 10))
	copy(p[15:25], name)
	return nil
}

// ---------------------------------------------------------------- collegamento

type uvproSess struct {
	*rawSess
	info   []byte // 9 byte
	group  int    // gruppo attivo (-1 = non noto)
	gname  string
	unseen int // blocchi non richiesti saltati
}

func uvproFrame(cmd uint16, body []byte) []byte {
	out := []byte{0xFF, 0x01, 0x00, byte(len(body)), 0x00, 0x02, byte(cmd >> 8), byte(cmd)}
	return append(out, body...)
}

// request: invia il comando e dà i dati della risposta (dal byte dell'esito);
// i blocchi che non c'entrano si saltano, come CHIRP
func (s *uvproSess) request(cmd uint16, body []byte) ([]byte, error) {
	if err := s.send(uvproFrame(cmd, body)); err != nil {
		return nil, err
	}
	for {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		h, err := s.read(4, uvproTimeout)
		if err != nil {
			return nil, err
		}
		if h[0] != 0xFF || h[1] != 0x01 {
			return nil, rErr("dati inattesi dalla radio (%02X %02X): la porta scelta è quella Bluetooth della radio?", h[0], h[1])
		}
		msg, err := s.read(4+int(h[3])+int(h[2]&1), uvproTimeout)
		if err != nil {
			return nil, err
		}
		msg = msg[:4+int(h[3])]
		if binary.BigEndian.Uint16(msg[0:2]) == 2 && binary.BigEndian.Uint16(msg[2:4]) == 0x8000|cmd {
			return msg[4:], nil
		}
		if s.unseen++; s.unseen > 500 {
			return nil, rErr("la radio non risponde ai comandi")
		}
	}
}

func uvproOpen(ctx context.Context, m *genModel, port string) (*uvproSess, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	s := &uvproSess{rawSess: rs, group: -1}
	fail := func(err error) (*uvproSess, error) {
		rs.p.Close()
		return nil, genFail("Collegamento", err)
	}
	r, err := s.request(4, []byte{0x03})
	if err != nil {
		return fail(err)
	}
	if len(r) < 1+9 || r[0] != 0 {
		return fail(rErr("la radio non ha dato le sue informazioni"))
	}
	s.info = append([]byte{}, r[1:10]...)
	// gruppo attivo e suo nome: se la radio non li dà si prosegue senza (come CHIRP)
	if r, err = s.request(20, nil); err == nil && len(r) >= 1+4 && r[0] == 0 {
		g := int(binary.BigEndian.Uint16(r[3:5]) >> 6 & 0x3F)
		s.group = g
		if r, err = s.request(73, []byte{byte(g)}); err == nil && len(r) >= 2 && r[0] == 0 && int(r[1]) == g {
			name := r[2:]
			if len(name) > 10 {
				name = name[:10]
			}
			if k := bytes.IndexByte(name, 0); k >= 0 {
				name = name[:k]
			}
			s.gname = strings.TrimRight(asciiName(string(name)), " ")
		}
	}
	if ctx.Err() != nil {
		rs.p.Close()
		return nil, errCancelled
	}
	return s, nil
}

func (s *uvproSess) softVer() int { return int(binary.BigEndian.Uint16(s.info[4:6])) }

// readCh: stato e dati del canale i
func (s *uvproSess) readCh(i int) (status byte, rec []byte, err error) {
	r, err := s.request(13, []byte{byte(i)})
	if err != nil {
		return 0, nil, err
	}
	if len(r) < 1 {
		return 0, nil, rErr("risposta vuota della radio alla lettura del canale %d", i)
	}
	return r[0], r[1:], nil
}

func (s *uvproSess) writeCh(i int, rec []byte) error {
	r, err := s.request(14, rec)
	if err != nil {
		return err
	}
	if len(r) < 1 || r[0] != 0 {
		st := -1
		if len(r) > 0 {
			st = int(r[0])
		}
		return rErr("la radio ha rifiutato la scrittura del canale %d (esito %d)", i, st)
	}
	if len(r) > 1 && int(r[1]) != i {
		return rErr("la radio ha confermato il canale %d al posto del %d", r[1], i)
	}
	return nil
}

func (s *uvproSess) readAll(m *genModel, prog Progress, what string) ([]byte, error) {
	f := uvproFam{}
	img := make([]byte, uvproSize(m))
	copy(img, uvproMagic)
	copy(img[uvproInfoAt:], s.info)
	img[uvproGroupAt] = 0xFF
	if s.group >= 0 {
		img[uvproGroupAt] = byte(s.group)
		copy(img[uvproNameAt:uvproNameAt+10], s.gname)
	}
	for i := 0; i < m.Slots; i++ {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		st, rec, err := s.readCh(i)
		if err != nil {
			return nil, genFail("Lettura", err)
		}
		r := f.rec(img, i)
		switch {
		case st != 0:
			r[0] = uvproNone
		case len(rec) == uvproLen:
			r[0] = uvproOK
			copy(r[1:], rec)
		default:
			r[0] = uvproForeign
		}
		prog(i+1, m.Slots, what)
	}
	return img, nil
}

func (f uvproFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := uvproOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	img, err := s.readAll(m, prog, "Lettura…")
	if err != nil {
		return nil, "", err
	}
	note := ""
	switch {
	case s.group >= 0 && s.gname != "":
		note = fmt.Sprintf("Letto il gruppo attivo sulla radio: %d («%s»).", s.group, s.gname)
	case s.group >= 0:
		note = fmt.Sprintf("Letto il gruppo attivo sulla radio: %d.", s.group)
	}
	if v := s.softVer(); v != uvproSoftVer {
		note = strings.TrimSpace(note + fmt.Sprintf(" La radio ha la versione %d del firmware: la scrittura è ammessa solo con la versione %d.", v, uvproSoftVer))
	}
	prog(m.Slots, m.Slots, "Lettura completata")
	return img, note, nil
}

func (f uvproFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, m.Slots, "Collegamento con la radio…")
	s, err := uvproOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	if v := s.softVer(); v != uvproSoftVer {
		return "", rErr("scrittura non ammessa: la radio ha la versione %d del firmware, il driver è stato provato solo con la %d. Nulla è stato scritto", v, uvproSoftVer)
	}
	if g := img[uvproGroupAt]; g != 0xFF && s.group >= 0 && int(g) != s.group {
		return "", rErr("la radio ora è sul gruppo %d, i canali erano stati letti dal gruppo %d: rimetti quel gruppo sulla radio oppure rileggi. Nulla è stato scritto", s.group, g)
	}
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
		rec := f.ch(img, i)
		if rec == nil {
			continue // posizione che la radio non aveva dato: da qui non si modifica
		}
		// come CHIRP: il canale si rilegge subito prima di scriverlo…
		st, before, err := s.readCh(i)
		if err != nil {
			return "", genFail("Scrittura", err)
		}
		if st != 0 || len(before) != uvproLen {
			return "", rErr("il canale %d non si può scrivere: la radio non lo dà alla lettura (esito %d, %d byte)", i, st, len(before))
		}
		if uvproDMR(before) {
			return "", rErr("il canale %d sulla radio è DMR: si modifica con l'applicazione del costruttore", i)
		}
		if bytes.Equal(before, rec) {
			continue
		}
		// …e subito dopo; se non coincide si rimette quello di prima
		werr := s.writeCh(i, rec)
		if werr == nil {
			var after []byte
			if st, after, werr = s.readCh(i); werr == nil && (st != 0 || !bytes.Equal(after, rec)) {
				werr = rErr("il canale %d riletto dalla radio non è quello scritto", i)
			}
		}
		if werr != nil {
			msg := werr.Error()
			if s.writeCh(i, before) == nil {
				if _, again, e := s.readCh(i); e == nil && bytes.Equal(again, before) {
					return "", rErr("%s; il canale di prima è stato rimesso e verificato", msg)
				}
			}
			return "", rErr("%s; NON è stato possibile rimettere il canale di prima: rileggi la radio prima di altre modifiche", msg)
		}
	}
	prog(len(todo), len(todo), "Scrittura completata")
	return "", nil
}
