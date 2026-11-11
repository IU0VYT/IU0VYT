// IU0VYT Flash - gestione canali, ponti e firmware di radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
//
// Questo programma è software libero: puoi ridistribuirlo e/o modificarlo secondo
// i termini della GNU General Public License pubblicata dalla Free Software
// Foundation, versione 3 della licenza o (a tua scelta) qualsiasi versione
// successiva. È distribuito SENZA ALCUNA GARANZIA: vedi il file LICENSE.

// 4.16: radio DMR con il firmware libero OpenGD77 (Radioddity GD-77 e GD-77S,
// Baofeng DM-1801 e RD-5R) e OpenUV380 / OpenDM1701 (Baofeng DM-1701, TYT
// MD-UV380, MD-380, MD-9600, MD-2017, Retevis RT3S, RT84).
//
// Protocollo e disposizione della memoria sono quelli descritti nel programma
// libero qdmr di Hannes Matuschek DM3MAT (GNU GPL v3: lib/opengd77_interface.hh,
// opengd77_codeplug.hh, openuv380_codeplug.hh, opengd77base_codeplug.hh), riscritti
// in Go. Il formato del canale è quello della famiglia GD-77 (vedi drv_rdty.go).
//
// Collegamento: la radio accesa normalmente compare come porta seriale USB
// (VID 1FC9, PID 0094), 115200 baud.
//   lettura    'R' zona(1 flash, 2 EEPROM, 9 informazioni) indirizzo(4, big endian)
//              lunghezza(2) → 'R' lunghezza(2) dati
//   EEPROM     'W' 4 indirizzo(4) lunghezza(2) dati(32) → 'W' 4
//   flash      'W' 1 settore(3) → 'W' 1 (il settore di 4096 byte va nel buffer);
//              'W' 2 indirizzo(4) lunghezza(2) dati → 'W' 2;  'W' 3 → 'W' 3 (scrive)
//   comandi    'C' codice x y carattere allineamento inverso testo(16) → '-'
//              (0 schermo del programma, 1 pulisci, 2 testo, 3 mostra, 5 chiudi,
//              6 azione: 0 salva le impostazioni, 2 salva impostazioni e VFO,
//              3 led verde, 4 led rosso)
// Le radio della famiglia MD-UV380 usano 'X' al posto di 'W'.
//
// Memoria: EEPROM 0x00E0-0x6000 e 0x7500-0xB000 (canali 1-128 a 0x3780, zone a
// 0x8010); flash, 0x13E60 byte da 0x7B000 (GD-77) o 0x9B000 (MD-UV380): canali
// 129-1024 da +0x1B0, poi contatti e gruppi. L'immagine del programma è la EEPROM
// (0xB000 byte, le parti non lette a FF) seguita dalla zona della flash.
//
// Si scrivono solo i canali e le zone, e solo i blocchi cambiati rispetto alla
// radio (riletta prima di scrivere): impostazioni, contatti, gruppi, taratura e
// tutto il resto non si toccano mai.

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	ogdEep      = 0xB000  // byte di EEPROM nell'immagine
	ogdFlashLen = 0x13E60 // byte di flash nell'immagine
	ogdSize     = ogdEep + ogdFlashLen
	ogdBank0    = 0x3780         // canali 1-128 (EEPROM)
	ogdBank1    = ogdEep + 0x1B0 // canali 129-1024 nell'immagine
	ogdBankLen  = 16 + 128*rdtyChSize
	ogdZones    = 68
	ogdZoneMem  = 80
	ogdZoneEnd  = rdtyZoneTab + 32 + ogdZones*(16+2*ogdZoneMem)
	ogdBlock    = 32
	ogdSector   = 4096
)

// zone della EEPROM lette (come qdmr)
var ogdEepRanges = [][2]int{{0x00E0, 0x6000}, {0x7500, 0xB000}}

type ogdOpt struct {
	wr    byte   // 'W' (GD-77, DM-1801, RD-5R) oppure 'X' (famiglia MD-UV380)
	flash int    // indirizzo nella flash dell'inizio della zona letta
	types []byte // tipi di radio (informazioni del firmware) di questa scheda
	what  string
}

type ogdFam struct{}

var ogdZoneOpt = &rdtyOpt{zoneMem: ogdZoneMem, nzones: ogdZones}

// tipi di radio nelle informazioni del firmware
var ogdTypeName = map[byte]string{0: "Radioddity GD-77", 1: "Radioddity GD-77S", 2: "Baofeng DM-1801", 3: "Baofeng RD-5R",
	4: "Baofeng DM-1801A", 5: "TYT MD-9600", 6: "TYT MD-UV380", 7: "TYT MD-380", 8: "Baofeng DM-1701", 9: "TYT MD-2017", 10: "Baofeng DM-1701 (RGB)"}

func init() {
	mk := func(id, model, name string, o ogdOpt, bands []band) *genModel {
		oo := o
		return &genModel{ID: id, Vendor: "OpenGD77", Model: model, RClass: "IU0VYT_" + model, Name: name,
			Slots: 1024, First: 1, NameMax: 16, Charset: charsetASCII,
			Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2 W", 2, 1}, {"Bassa 500 mW", 0.5, 2}, {"Minima 50 mW", 0.05, 3}},
			Bands:  bands, Narrow: true, HasScan: true, NoChirp: true,
			DCS: dtcsCodes, Baud: 115200, Fam: ogdFam{}, Opt: &oo, Digital: "DMR",
			Notes: "Firmware libero " + o.what + ": la radio accesa normalmente compare come porta seriale USB (115200 baud). " +
				"1024 canali con nomi di 16 caratteri, analogici o DMR. Si modificano frequenze, nome, modo, potenza, banda, toni, " +
				"sola ricezione ed esclusione dalla scansione; per i canali DMR color code, slot, contatto e gruppo restano quelli " +
				"impostati (i canali DMR nuovi: color code 1, slot 1). Una potenza già impostata sulla radio (anche «quella della " +
				"radio») resta com'è finché non la cambi. Un canale cancellato esce anche dalle zone; i canali nuovi si trovano " +
				"nella zona «Tutti i canali» della radio. Prima di scrivere il programma rilegge la radio e scrive solo i blocchi " +
				"cambiati di canali e zone: impostazioni, contatti, gruppi e taratura non si toccano mai. " +
				"Protocollo e memoria come nel programma libero qdmr."}
	}
	vu := []band{{136000000, 174000000}, {400000000, 480000000}}
	vu220 := []band{{136000000, 174000000}, {200000000, 260000000}, {400000000, 480000000}}
	genRegister(
		mk("ogd77", "OpenGD77", "OpenGD77 (Radioddity GD-77 / GD-77S, Baofeng DM-1801 / RD-5R)",
			ogdOpt{wr: 'W', flash: 0x7B000, types: []byte{0, 1, 2, 3, 4}, what: "OpenGD77"}, vu220),
		mk("ogd77-uv380", "OpenUV380", "OpenGD77 (Baofeng DM-1701, TYT MD-UV380 / MD-380 / MD-9600 / MD-2017, Retevis RT3S / RT84)",
			ogdOpt{wr: 'X', flash: 0x9B000, types: []byte{5, 6, 7, 8, 9, 10}, what: "OpenGD77 per la famiglia MD-UV380 (OpenUV380, OpenDM1701)"}, vu),
	)
}

func ogdo(m *genModel) *ogdOpt { return m.Opt.(*ogdOpt) }

func ogdBank(b int) int {
	if b == 0 {
		return ogdBank0
	}
	return ogdBank1 + (b-1)*ogdBankLen
}

func ogdRecOff(i int) int { return ogdBank(i>>7) + 16 + (i%128)*rdtyChSize }

func ogdValid(img []byte, i int) bool { return img[ogdBank(i>>7)+(i%128)/8]>>(i&7)&1 == 1 }

func ogdSetValid(img []byte, i int, v bool) {
	p := ogdBank(i>>7) + (i%128)/8
	if v {
		img[p] |= 1 << (i & 7)
	} else {
		img[p] &^= 1 << (i & 7)
	}
}

// potenza del canale (byte 25): 0 quella della radio, 1 = 50 mW … 9 = 5 W, 10 = massima
func ogdPowerSem(raw byte) int {
	switch {
	case raw == 1:
		return 3
	case raw == 2 || raw == 3:
		return 2
	case raw >= 4 && raw <= 6:
		return 1
	}
	return 0
}

var ogdPowerRaw = [4]byte{9, 6, 3, 1}

func (ogdFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, img[ogdRecOff(i):][:rdtyChSize]...)
}

func (ogdFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(img[ogdRecOff(i):][:rdtyChSize], raw)
	ogdSetValid(img, i, true)
}

func (ogdFam) Match(m *genModel, img []byte) bool { return false }

func (f ogdFam) Get(m *genModel, img []byte, i int) *Chan {
	if len(img) < ogdSize || !ogdValid(img, i) {
		return nil
	}
	r := img[ogdRecOff(i):][:rdtyChSize]
	rx, ok1 := rdtyFreq(r[16:])
	tx, ok2 := rdtyFreq(r[20:])
	if !ok1 || !ok2 || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: tx, Name: rdtyName(r[0:16])}
	c.Dig = r[24] == 1
	c.Narrow = r[51]>>1&1 == 0
	c.Power = ogdPowerSem(r[25])
	c.TxOff = r[51]>>2&1 == 1
	c.Scan = r[51]>>4&1 == 0
	if !c.Dig {
		c.RxTone = rdtyTone(binary.LittleEndian.Uint16(r[32:]))
		c.TxTone = rdtyTone(binary.LittleEndian.Uint16(r[34:]))
	}
	return c
}

func (f ogdFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if len(img) < ogdSize {
		return fmt.Errorf("immagine della memoria incompleta")
	}
	r := img[ogdRecOff(i):][:rdtyChSize]
	if c == nil {
		if ogdValid(img, i) {
			for k := range r {
				r[k] = 0
			}
			copy(r[0:16], bytes.Repeat([]byte{0xFF}, 16))
			r[32], r[33], r[34], r[35] = 0xFF, 0xFF, 0xFF, 0xFF
			ogdSetValid(img, i, false)
			rdtyUnlink(ogdZoneOpt, img, i)
		}
		return nil
	}
	if c.Rx%10 != 0 || c.Rx <= 0 || c.Rx > 999999990 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	if tx%10 != 0 || tx <= 0 || tx > 999999990 {
		return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(tx))
	}
	for _, ch := range c.Name {
		if ch < 0x20 || ch > 0x7E {
			return fmt.Errorf("carattere non ammesso nel nome")
		}
	}
	var tt, rt genTone
	if !c.Dig {
		var err error
		if tt, err = genToneParse(c.TxTone); err != nil {
			return err
		}
		if rt, err = genToneParse(c.RxTone); err != nil {
			return err
		}
	}
	isNew := !ogdValid(img, i)
	if isNew {
		// canale nuovo come ChannelElement::clear() di qdmr: tutto a zero, squelch e
		// potenza scelti qui sotto
		for k := range r {
			r[k] = 0
		}
		if c.Dig {
			r[44] = 1 // color code 1, slot 1, nessun contatto
		}
	}
	copy(r[0:16], bytes.Repeat([]byte{0xFF}, 16))
	copy(r[0:16], c.Name)
	rdtyPutFreq(r[16:], c.Rx)
	rdtyPutFreq(r[20:], tx)
	if c.Dig {
		r[24] = 1
		binary.LittleEndian.PutUint16(r[32:], 0xFFFF)
		binary.LittleEndian.PutUint16(r[34:], 0xFFFF)
		if isNew {
			r[51] &^= 1 << 1 // 12,5 kHz
		}
	} else {
		r[24] = 0
		binary.LittleEndian.PutUint16(r[32:], rdtyToneEnc(rt))
		binary.LittleEndian.PutUint16(r[34:], rdtyToneEnc(tt))
		if c.Narrow {
			r[51] &^= 1 << 1
		} else {
			r[51] |= 1 << 1
		}
	}
	// potenza: quella che c'era resta se il livello non cambia (anche «quella
	// della radio», 0)
	if isNew || ogdPowerSem(r[25]) != c.Power {
		p := c.Power
		if p < 0 || p > 3 {
			p = 0
		}
		r[25] = ogdPowerRaw[p]
	}
	if c.TxOff {
		r[51] |= 1 << 2
	} else {
		r[51] &^= 1 << 2
	}
	if c.Scan {
		r[51] &^= 1 << 4
	} else {
		r[51] |= 1 << 4
	}
	ogdSetValid(img, i, true)
	return nil
}

// ---------------------------------------------------------------- collegamento

type ogdSess struct {
	ctx context.Context
	p   Port
	wr  byte
}

func (s *ogdSess) get(n int, d time.Duration) ([]byte, error) {
	b := readUpTo(s.ctx, s.p, n, d)
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	if len(b) == 0 {
		return nil, errTimeout
	}
	return b, nil
}

// comando 'C': la radio risponde '-'
func (s *ogdSess) cmd(code, x, y, font, align, inv byte, text string) error {
	b := make([]byte, 23)
	b[0], b[1], b[2], b[3], b[4], b[5], b[6] = 'C', code, x, y, font, align, inv
	copy(b[7:], text)
	if err := s.p.Write(b); err != nil {
		return err
	}
	r, err := s.get(1, 2*time.Second)
	if err != nil {
		return err
	}
	if r[0] != '-' {
		return rErr("risposta della radio non valida")
	}
	return nil
}

func (s *ogdSess) read(zone byte, addr, n int) ([]byte, error) {
	req := []byte{'R', zone, byte(addr >> 24), byte(addr >> 16), byte(addr >> 8), byte(addr), byte(n >> 8), byte(n)}
	if err := s.p.Write(req); err != nil {
		return nil, err
	}
	r, err := s.get(3+n, 2*time.Second)
	if err != nil {
		return nil, err
	}
	if len(r) != 3+n || r[0] != 'R' || int(r[1])<<8|int(r[2]) != n {
		return nil, rErr("risposta della radio non valida per il blocco 0x%05X", addr)
	}
	return r[3:], nil
}

func (s *ogdSess) write(code byte, body []byte) error {
	if err := s.p.Write(append([]byte{s.wr, code}, body...)); err != nil {
		return err
	}
	r, err := s.get(2, 3*time.Second)
	if err != nil {
		return err
	}
	if len(r) != 2 || r[0] != s.wr || r[1] != code {
		return rErr("la radio non ha confermato la scrittura")
	}
	return nil
}

func ogdAddrLen(addr, n int) []byte {
	return []byte{byte(addr >> 24), byte(addr >> 16), byte(addr >> 8), byte(addr), byte(n >> 8), byte(n)}
}

// start: informazioni del firmware e controllo del tipo di radio
func (s *ogdSess) start(m *genModel) (string, error) {
	info, err := s.read(9, 0, 46)
	if err != nil {
		return "", genFail("Collegamento", err)
	}
	typ := byte(binary.LittleEndian.Uint32(info[4:]))
	rev := string(bytes.TrimRight(info[8:24], "\x00 "))
	name := ogdTypeName[typ]
	if name == "" {
		name = fmt.Sprintf("%d", typ)
	}
	note := fmt.Sprintf("Firmware della radio: %s", "OpenGD77 "+rev+" ("+name+")")
	if bytes.IndexByte(ogdo(m).types, typ) < 0 {
		for _, o := range genModels {
			if oo, ok := o.Opt.(*ogdOpt); ok && o != m && bytes.IndexByte(oo.types, typ) >= 0 {
				return note, rErr("la radio è un altro modello (%s): apri la scheda «%s» o quella della stessa serie", name, o.Name)
			}
		}
		return note, rErr("la radio è un altro modello (%s)", name)
	}
	return note, nil
}

// screen: schermo del programma sulla radio, come qdmr
func (s *ogdSess) screen(line string, led byte) error {
	steps := []func() error{
		func() error { return s.cmd(0, 0, 0, 0, 0, 0, "") },
		func() error { return s.cmd(1, 0, 0, 0, 0, 0, "") },
		func() error { return s.cmd(2, 0, 0, 3, 1, 0, "IU0VYT Flash") },
		func() error { return s.cmd(2, 0, 16, 3, 1, 0, line) },
		func() error { return s.cmd(3, 0, 0, 0, 0, 0, "") },
		func() error { return s.cmd(6, led, 0, 0, 0, 0, "") },
		func() error { return s.cmd(6, 2, 0, 0, 0, 0, "") },
	}
	for _, f := range steps {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// blocchi dell'immagine: posizione nell'immagine, zona (2 EEPROM, 1 flash) e indirizzo
type ogdBlk struct {
	off  int
	zone byte
	addr int
}

func ogdBlocks(o *ogdOpt) []ogdBlk {
	var out []ogdBlk
	for _, r := range ogdEepRanges {
		for a := r[0]; a < r[1]; a += ogdBlock {
			out = append(out, ogdBlk{a, 2, a})
		}
	}
	for a := 0; a < ogdFlashLen; a += ogdBlock {
		out = append(out, ogdBlk{ogdEep + a, 1, o.flash + a})
	}
	return out
}

// ogdWritable: il blocco appartiene ai canali o alle zone
func ogdWritable(off int) bool {
	switch {
	case off >= ogdBank0 && off < ogdBank0+ogdBankLen:
		return true
	case off >= rdtyZoneTab && off < ogdZoneEnd:
		return true
	case off >= ogdBank1 && off < ogdBank1+7*ogdBankLen:
		return true
	}
	return false
}

func (s *ogdSess) readAll(o *ogdOpt, prog Progress, what string) ([]byte, error) {
	img := bytes.Repeat([]byte{0xFF}, ogdSize)
	bl := ogdBlocks(o)
	for k, b := range bl {
		if s.ctx.Err() != nil {
			return nil, errCancelled
		}
		d, err := s.read(b.zone, b.addr, ogdBlock)
		if err != nil {
			return nil, err
		}
		copy(img[b.off:], d)
		if k%16 == 0 {
			prog(k*ogdBlock, len(bl)*ogdBlock, what)
		}
	}
	return img, nil
}

func ogdOpen(ctx context.Context, m *genModel, port string) (*ogdSess, string, error) {
	p, err := openPort(port, 115200)
	if err != nil {
		return nil, "", err
	}
	s := &ogdSess{ctx: ctx, p: p, wr: ogdo(m).wr}
	note, err := s.start(m)
	if err != nil {
		p.Close()
		return nil, note, err
	}
	return s, note, nil
}

func (f ogdFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, ogdSize, "Collegamento con la radio…")
	s, note, err := ogdOpen(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	if err := s.screen("Lettura", 3); err != nil {
		return nil, "", err
	}
	img, err := s.readAll(ogdo(m), prog, "Lettura…")
	_ = s.cmd(5, 0, 0, 0, 0, 0, "")
	if err != nil {
		return nil, "", err
	}
	prog(ogdSize, ogdSize, "Lettura completata")
	return img, note, nil
}

func (f ogdFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != ogdSize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	o := ogdo(m)
	prog(0, ogdSize, "Collegamento con la radio…")
	s, note, err := ogdOpen(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	if err := s.screen("Scrittura", 4); err != nil {
		return "", err
	}
	// la radio si rilegge: si scrivono solo i blocchi cambiati di canali e zone
	cur, err := s.readAll(o, prog, "Lettura della radio prima della scrittura…")
	if err != nil {
		_ = s.cmd(5, 0, 0, 0, 0, 0, "")
		return "", err
	}
	var todo []ogdBlk
	for _, b := range ogdBlocks(o) {
		if ogdWritable(b.off) && !bytes.Equal(cur[b.off:b.off+ogdBlock], img[b.off:b.off+ogdBlock]) {
			todo = append(todo, b)
		}
	}
	sector := -1
	finish := func() error {
		if sector < 0 {
			return nil
		}
		sector = -1
		return s.write(3, nil)
	}
	for k, b := range todo {
		if ctx.Err() != nil {
			_ = finish()
			_ = s.cmd(5, 0, 0, 0, 0, 0, "")
			return "", errCancelled
		}
		prog(k, len(todo), "Scrittura…")
		data := img[b.off : b.off+ogdBlock]
		if b.zone == 2 {
			if err = s.write(4, append(ogdAddrLen(b.addr, ogdBlock), data...)); err != nil {
				break
			}
			continue
		}
		if sec := b.addr / ogdSector; sec != sector {
			if err = finish(); err != nil {
				break
			}
			if err = s.write(1, []byte{byte(sec >> 16), byte(sec >> 8), byte(sec)}); err != nil {
				break
			}
			sector = sec
		}
		if err = s.write(2, append(ogdAddrLen(b.addr, ogdBlock), data...)); err != nil {
			break
		}
	}
	if e2 := finish(); err == nil {
		err = e2
	}
	_ = s.cmd(5, 0, 0, 0, 0, 0, "")
	if err != nil {
		return "", err
	}
	_ = s.cmd(6, 0, 0, 0, 0, 0, "") // salva e ricarica le impostazioni
	prog(len(todo), len(todo), "Scrittura completata")
	return note, nil
}
