// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da dmrconfig (https://github.com/OpenRTX/dmrconfig, licenza BSD a
// 3 clausole, Copyright (C) 2018 Serge Vakulenko KK6ABQ): uv380.c, dfu-libusb.c,
// dfu-windows.c; confrontate con qdmr (https://github.com/hmatuschek/qdmr, GNU GPL
// v3, Hannes Matuschek DM3MAT): dm1701_codeplug.cc, tyt_interface.cc.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Baofeng DM-1701 (anche Retevis RT84): memoria della famiglia TYT MD-UV380,
// collegamento USB in modalità DFU (STMicroelectronics 0483:DF11), come in
// dmrconfig (uv380.c, dfu-libusb.c, dfu-windows.c).
//
// Protocollo DFU (richieste di controllo USB; con il driver «STTub30» installato
// dal programma del costruttore, come dfu-windows.c):
//   GETSTATE / GETSTATUS / CLRSTATUS / ABORT / DETACH per tornare a «dfuIDLE»;
//   DNLOAD [91 01] modalità di programmazione; DNLOAD [A2 01] + UPLOAD 64 byte:
//   identificazione («DM-1701»); DNLOAD [21 indirizzo(4)] indirizzo 0;
//   UPLOAD/DNLOAD con valore = blocco+2: 1024 byte (i blocchi da 256 in su
//   sono spostati di 832: memoria estesa da 0x110000);
//   DNLOAD [41 indirizzo(4)] cancella 64 KB (prima di scrivere: 0x00000-0x3FFFF
//   e 0x110000-0x1DFFFF); DNLOAD [91 05] alla fine: la radio si riavvia.
//
// Memoria (0xD0000 byte): canali 3000 da 64 byte a 0x40000 (validi se il nome,
// Unicode a 16 caratteri, non è vuoto); zone 250 da 64 byte a 0x149E0 (nome e
// 16 canali) con l'estensione da 224 byte a 0x31000 (altri 48 canali della lista
// A, 64 della lista B); liste di scansione 250 da 104 byte a 0x18860.
// Canale: 0 bit 0-1 modo (1 analogico, 2 DMR), bit 2-3 banda (0 12,5, 1 20,
// 2 25 kHz); 1 bit 1 solo ricezione, bit 2-3 slot, bit 4-7 color code; 6 contatto;
// 8 TOT; 11 lista di scansione; 12 lista dei gruppi; 15 squelch; 16 RX e 20 TX
// (BCD, 10 Hz); 24 tono RX e 26 tono TX (come la RD-5R); 30 bit 0-1 potenza
// (3 alta, 2 media, 0 bassa); 32 nome.
//
// Qui si gestiscono i canali (frequenze, nome, modo, potenza, banda, toni, solo
// ricezione); le impostazioni DMR dei canali esistenti, i contatti, i gruppi, i
// messaggi, la rubrica e le impostazioni della radio restano come sono.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"
)

const (
	tytMemSize    = 0xD0000
	tytChOff      = 0x40000
	tytChSize     = 64
	tytNCh        = 3000
	tytZoneOff    = 0x149E0
	tytZoneExtOff = 0x31000
	tytNZones     = 250
	tytScanOff    = 0x18860
	tytScanSize   = 104
	tytNScan      = 250
	tytVID        = 0x0483
	tytPID        = 0xDF11
)

type tytOpt struct {
	ident string // risposta all'identificazione
}

type tytFam struct{}

func init() {
	genRegister(&genModel{ID: "bf-dm1701", Vendor: "Baofeng", Model: "DM-1701", RClass: "IU0VYT_DM1701", Name: "Baofeng DM-1701 / Retevis RT84",
		Slots: tytNCh, First: 1, NameMax: 16, Charset: charsetASCII,
		Powers: []genPower{{"Alta", 5, 0}, {"Media", 2.5, 1}, {"Bassa", 1, 2}},
		Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: false, NoChirp: true,
		DCS: dtcsCodes, Baud: 0, Fam: tytFam{}, Opt: &tytOpt{ident: "DM-1701"}, Digital: "DMR", Conn: "dfu",
		Notes: "Radio DMR collegata via USB in modalità DFU: serve il driver USB del costruttore (pacchetto «DM-1701 CPS», cartella " +
			"«DM-1701 USB Driver», lo stesso usato dal suo programma: si installa con il pulsante «Installa il driver USB…» della finestra " +
			"«Leggi dalla radio») oppure il driver generico WinUSB installato con Zadig. Memoria e protocollo come nei programmi liberi dmrconfig e qdmr, " +
			"confrontati con il codice di dmrconfig. Si modificano frequenze, nomi (16 caratteri), modo (analogico / DMR), potenza, banda, " +
			"toni e solo ricezione; per i canali DMR color code, slot, contatto e lista dei gruppi restano quelli impostati (i canali DMR " +
			"nuovi: color code 1, slot 1, senza contatto). Contatti, gruppi, messaggi, rubrica e impostazioni della radio non si toccano. " +
			"La scrittura cancella e riscrive la memoria dei canali (circa 1 minuto)."})
}

func tytRec(img []byte, i int) []byte { return img[tytChOff+i*tytChSize:][:tytChSize] }

func tytStr(b []byte, n int) string {
	var u []uint16
	for k := 0; k < n; k++ {
		v := binary.LittleEndian.Uint16(b[2*k:])
		if v == 0 || v == 0xFFFF {
			break
		}
		u = append(u, v)
	}
	return strings.TrimRight(string(utf16.Decode(u)), " ")
}

// come utf8_decode di dmrconfig: caratteri e poi zeri
func tytPutStr(b []byte, n int, s string) {
	u := utf16.Encode([]rune(s))
	for k := 0; k < n; k++ {
		var v uint16
		if k < len(u) {
			v = u[k]
		}
		binary.LittleEndian.PutUint16(b[2*k:], v)
	}
}

func tytValid(img []byte, i int) bool {
	v := binary.LittleEndian.Uint16(tytRec(img, i)[32:])
	return v != 0 && v != 0xFFFF
}

func (f tytFam) Get(m *genModel, img []byte, i int) *Chan {
	if len(img) != tytMemSize || i < 0 || i >= tytNCh || !tytValid(img, i) {
		return nil
	}
	r := tytRec(img, i)
	rx, ok1 := rdtyFreq(r[16:])
	tx, ok2 := rdtyFreq(r[20:])
	if !ok1 || !ok2 || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: tx, Name: tytStr(r[32:], 16)}
	c.Dig = r[0]&3 == 2
	c.Narrow = r[0]>>2&3 == 0
	switch r[30] & 3 {
	case 3:
		c.Power = 0
	case 2:
		c.Power = 1
	default:
		c.Power = 2
	}
	c.TxOff = r[1]>>1&1 == 1
	c.Scan = r[11] != 0
	if !c.Dig {
		c.RxTone = rdtyTone(binary.LittleEndian.Uint16(r[24:]))
		c.TxTone = rdtyTone(binary.LittleEndian.Uint16(r[26:]))
	}
	return c
}

// tytErase: canale cancellato come erase_channel di dmrconfig
func tytErase(r []byte) {
	copy(r, []byte{
		1 | 3<<5,    // analogico, 12,5 kHz, _unused1 = 3
		1<<2 | 1<<4, // slot 1, color code 1
		0,
		6<<4 | 1<<7, // _unused4 = 6, display PTT ID disattivato
		1<<2 | 1<<5, // _unused5 = 1, _unused6 = 1
		3 << 6,      // nessuna frequenza di spegnimento
		0, 0,        // contatto
		4, 0, // TOT 60 s, ritardo
		0, 0, 0, 0, 0, 1, // emergenza, lista di scansione, gruppi, squelch 1
	})
	binary.LittleEndian.PutUint32(r[16:], 0x40000000)
	binary.LittleEndian.PutUint32(r[20:], 0x40000000)
	binary.LittleEndian.PutUint16(r[24:], 0xFFFF)
	binary.LittleEndian.PutUint16(r[26:], 0xFFFF)
	r[28], r[29] = 0, 0
	r[30] = 3 | 0x3F<<2            // potenza alta
	r[31] = 7 | 1<<3 | 1<<4 | 7<<5 // DCDM disattivato, MS
	tytPutStr(r[32:], 16, "")
}

func (f tytFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if i < 0 || i >= tytNCh {
		return fmt.Errorf("canale %d fuori dalla memoria", i+1)
	}
	r := tytRec(img, i)
	if c == nil {
		if tytValid(img, i) {
			tytErase(r)
			tytUnlink(img, i)
		}
		return nil
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
	isNew := !tytValid(img, i)
	wasDig := r[0]&3 == 2
	if isNew {
		tytErase(r)
	}
	name := c.Name
	if strings.TrimSpace(name) == "" {
		// la radio considera vuoto un canale senza nome
		name = fmt.Sprintf("CH %d", i+1)
	}
	tytPutStr(r[32:], 16, name)
	rdtyPutFreq(r[16:], c.Rx)
	tx := c.Tx
	if c.TxOff {
		tx = c.Rx
	}
	rdtyPutFreq(r[20:], tx)
	if c.Dig {
		r[0] = r[0]&^3 | 2
		binary.LittleEndian.PutUint16(r[24:], 0xFFFF)
		binary.LittleEndian.PutUint16(r[26:], 0xFFFF)
		if isNew || !wasDig {
			r[0] &^= 3 << 2 // 12,5 kHz, come dmrconfig per i canali DMR
		}
		if r[4]>>6 == 2 {
			r[4] &^= 3 << 6 // ammissione «tono» non valida per un canale DMR
		}
	} else {
		r[0] = r[0]&^3 | 1
		binary.LittleEndian.PutUint16(r[24:], rdtyToneEnc(rt))
		binary.LittleEndian.PutUint16(r[26:], rdtyToneEnc(tt))
		bw := r[0] >> 2 & 3
		if c.Narrow {
			bw = 0
		} else if bw == 0 || bw == 3 {
			bw = 2
		}
		r[0] = r[0]&^(3<<2) | bw<<2
		if r[4]>>6 == 3 {
			r[4] &^= 3 << 6 // ammissione «color code» non valida in analogico
		}
		if wasDig && !isNew {
			// come setup_channel per un canale analogico: color code 1,
			// slot 1, nessun contatto e nessuna lista dei gruppi
			r[1] = r[1]&^0xFC | 1<<2 | 1<<4
			r[6], r[7], r[12] = 0, 0, 0
		}
	}
	switch m.powerIndex(c.Power) {
	case 0:
		r[30] = r[30]&^3 | 3
	case 1:
		r[30] = r[30]&^3 | 2
	default:
		r[30] &^= 3
	}
	if c.TxOff {
		r[1] |= 1 << 1
	} else {
		r[1] &^= 1 << 1
	}
	if isNew {
		tytZoneAdd(img, i)
	}
	return nil
}

// ---------------------------------------------------------------- zone e liste

func tytZone(img []byte, z int) []byte    { return img[tytZoneOff+z*64:][:64] }
func tytZoneExt(img []byte, z int) []byte { return img[tytZoneExtOff+z*224:][:224] }
func tytZoneValid(img []byte, z int) bool {
	v := binary.LittleEndian.Uint16(tytZone(img, z))
	return v != 0 && v != 0xFFFF
}
func tytScan(img []byte, s int) []byte { return img[tytScanOff+s*tytScanSize:][:tytScanSize] }
func tytScanValid(img []byte, s int) bool {
	v := binary.LittleEndian.Uint16(tytScan(img, s))
	return v != 0 && v != 0xFFFF
}

// elenco a più pezzi (zona: 16 + 48 canali): letto e riscritto compatto
type tytList [][]byte

func (l tytList) get() []uint16 {
	var out []uint16
	for _, p := range l {
		for k := 0; k+1 < len(p); k += 2 {
			if v := binary.LittleEndian.Uint16(p[k:]); v != 0 {
				out = append(out, v)
			}
		}
	}
	return out
}

func (l tytList) put(v []uint16) {
	n := 0
	for _, p := range l {
		for k := 0; k+1 < len(p); k += 2 {
			var x uint16
			if n < len(v) {
				x = v[n]
			}
			binary.LittleEndian.PutUint16(p[k:], x)
			n++
		}
	}
}

func (l tytList) size() int {
	n := 0
	for _, p := range l {
		n += len(p) / 2
	}
	return n
}

func tytZoneA(img []byte, z int) tytList {
	return tytList{tytZone(img, z)[32:64], tytZoneExt(img, z)[0:96]}
}
func tytZoneB(img []byte, z int) tytList { return tytList{tytZoneExt(img, z)[96:224]} }

func tytDrop(l tytList, cnum uint16) {
	old := l.get()
	var keep []uint16
	for _, v := range old {
		if v != cnum {
			keep = append(keep, v)
		}
	}
	if len(keep) != len(old) {
		l.put(keep)
	}
}

// tytUnlink toglie il canale i (numero i+1) da zone e liste di scansione
func tytUnlink(img []byte, i int) {
	cnum := uint16(i + 1)
	for z := 0; z < tytNZones; z++ {
		if tytZoneValid(img, z) {
			tytDrop(tytZoneA(img, z), cnum)
			tytDrop(tytZoneB(img, z), cnum)
		}
	}
	for s := 0; s < tytNScan; s++ {
		if !tytScanValid(img, s) {
			continue
		}
		sl := tytScan(img, s)
		tytDrop(tytList{sl[42:104]}, cnum)
		for _, p := range []int{32, 34, 36} {
			if binary.LittleEndian.Uint16(sl[p:]) == cnum {
				binary.LittleEndian.PutUint16(sl[p:], 0xFFFF)
			}
		}
	}
}

// tytZoneAdd: canale nuovo nella lista A della prima zona con posto libero,
// altrimenti in una zona nuova («Zone n»)
func tytZoneAdd(img []byte, i int) {
	cnum := uint16(i + 1)
	for z := 0; z < tytNZones; z++ {
		if !tytZoneValid(img, z) {
			continue
		}
		for _, v := range append(tytZoneA(img, z).get(), tytZoneB(img, z).get()...) {
			if v == cnum {
				return
			}
		}
	}
	for z := 0; z < tytNZones; z++ {
		if !tytZoneValid(img, z) {
			continue
		}
		a := tytZoneA(img, z)
		if v := a.get(); len(v) < a.size() {
			a.put(append(v, cnum))
			return
		}
	}
	for z := 0; z < tytNZones; z++ {
		if tytZoneValid(img, z) {
			continue
		}
		zr, ze := tytZone(img, z), tytZoneExt(img, z)
		for k := range zr {
			zr[k] = 0
		}
		for k := range ze {
			ze[k] = 0
		}
		tytPutStr(zr, 16, fmt.Sprintf("Zone %d", z+1))
		binary.LittleEndian.PutUint16(zr[32:], cnum)
		return
	}
}

func (f tytFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, tytRec(img, i)...)
}

func (f tytFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	was := tytValid(img, i)
	copy(tytRec(img, i), raw)
	if !was && tytValid(img, i) {
		tytZoneAdd(img, i)
	}
}

func (f tytFam) Match(m *genModel, img []byte) bool { return len(img) == tytMemSize }

// ---------------------------------------------------------------- protocollo DFU

// Ogni richiesta di controllo passa sul Port (openPortFunc("DFU:0483:DF11")):
// 8 byte (bmRequestType, bRequest, wValue, wIndex, wLength, little endian) e i
// dati in uscita; risposta: 4 byte (esito, intero con segno: byte trasferiti o
// errore negativo) e i dati in entrata. Su Windows il Port traduce in richieste
// al driver STTub30 (usb_windows.go); nelle prove è la radio simulata.

const (
	dfuDetach    = 0
	dfuDnload    = 1
	dfuUpload    = 2
	dfuGetStatus = 3
	dfuClrStatus = 4
	dfuGetState  = 5
	dfuAbort     = 6

	dfuStIdle         = 2
	dfuStDnBusy       = 4
	dfuStAppIdle      = 0
	dfuStAppDetach    = 1
	dfuStManifestWait = 8
	dfuStError        = 10
)

type dfuSess struct {
	ctx context.Context
	p   Port
}

func (s *dfuSess) ctl(in bool, req byte, value uint16, out []byte, inLen int) ([]byte, error) {
	if s.ctx.Err() != nil {
		return nil, errCancelled
	}
	rt := byte(0x21)
	n := len(out)
	if in {
		rt, n = 0xA1, inLen
	}
	hdr := []byte{rt, req, byte(value), byte(value >> 8), 0, 0, byte(n), byte(n >> 8)}
	if err := s.p.Write(append(hdr, out...)); err != nil {
		return nil, err
	}
	var rb [4]byte
	if err := s.p.ReadFull(s.ctx, rb[:], 10*time.Second); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	res := int32(binary.LittleEndian.Uint32(rb[:]))
	if res < 0 {
		return nil, rErr("richiesta USB rifiutata dalla radio (errore %d)", res)
	}
	if !in || res == 0 {
		return nil, nil
	}
	if int(res) > inLen {
		return nil, rErr("risposta USB troppo lunga")
	}
	data := make([]byte, res)
	if err := s.p.ReadFull(s.ctx, data, 10*time.Second); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *dfuSess) getStatus() error { _, err := s.ctl(true, dfuGetStatus, 0, nil, 6); return err }

func (s *dfuSess) waitIdle() error {
	for tries := 0; tries < 300; tries++ {
		st, err := s.ctl(true, dfuGetState, 0, nil, 1)
		if err != nil {
			return err
		}
		if len(st) != 1 {
			return rErr("stato DFU non valido")
		}
		switch st[0] {
		case dfuStIdle:
			return nil
		case dfuStAppIdle:
			_, err = s.ctl(false, dfuDetach, 5000, nil, 0)
		case dfuStError:
			_, err = s.ctl(false, dfuClrStatus, 0, nil, 0)
		case dfuStAppDetach, dfuStDnBusy, dfuStManifestWait:
			time.Sleep(100 * time.Millisecond)
			continue
		default:
			_, err = s.ctl(false, dfuAbort, 0, nil, 0)
		}
		if err != nil {
			return err
		}
	}
	return rErr("la radio non torna pronta (DFU)")
}

func (s *dfuSess) command(a, b byte) error {
	if _, err := s.ctl(false, dfuDnload, 0, []byte{a, b}, 0); err != nil {
		return err
	}
	if err := s.getStatus(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return s.waitIdle()
}

func (s *dfuSess) addrCmd(op byte, addr uint32) error {
	if _, err := s.ctl(false, dfuDnload, 0, []byte{op, byte(addr), byte(addr >> 8), byte(addr >> 16), byte(addr >> 24)}, 0); err != nil {
		return err
	}
	if err := s.getStatus(); err != nil {
		return err
	}
	return s.waitIdle()
}

func (s *dfuSess) enter() (string, error) {
	if err := s.waitIdle(); err != nil {
		return "", err
	}
	if err := s.command(0x91, 0x01); err != nil {
		return "", err
	}
	if err := s.command(0xA2, 0x01); err != nil {
		return "", err
	}
	id, err := s.ctl(true, dfuUpload, 0, nil, 64)
	if err != nil {
		return "", err
	}
	if err := s.getStatus(); err != nil {
		return "", err
	}
	if n := bytes.IndexByte(id, 0); n >= 0 {
		id = id[:n]
	}
	if err := s.addrCmd(0x21, 0); err != nil {
		return "", err
	}
	return string(id), nil
}

func tytPhys(bno int) uint16 {
	if bno >= 256 && bno < 2048 {
		bno += 832
	}
	return uint16(bno + 2)
}

func (s *dfuSess) reboot() error {
	if err := s.waitIdle(); err != nil {
		return err
	}
	if _, err := s.ctl(false, dfuDnload, 0, []byte{0x91, 0x05}, 0); err != nil {
		return err
	}
	return s.getStatus()
}

func tytOpen(ctx context.Context, m *genModel) (*dfuSess, string, error) {
	p, err := openPortFunc(fmt.Sprintf("DFU:%04X:%04X", tytVID, tytPID), 0)
	if err != nil {
		return nil, "", err
	}
	s := &dfuSess{ctx: ctx, p: p}
	id, err := s.enter()
	if err != nil {
		p.Close()
		return nil, id, genFail("Identificazione", err)
	}
	if want := m.Opt.(*tytOpt).ident; id != want {
		p.Close()
		return nil, id, genFail("Identificazione", rErr("modello errato: la radio si identifica come %q (attesa %q)", id, want))
	}
	return s, id, nil
}

func (f tytFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	nb := tytMemSize / 1024
	prog(0, nb, "Collegamento con la radio…")
	s, info, err := tytOpen(ctx, m)
	if err != nil {
		return nil, info, err
	}
	defer s.p.Close()
	img := make([]byte, tytMemSize)
	for bno := 0; bno < nb; bno++ {
		prog(bno, nb, "Lettura…")
		d, err := s.ctl(true, dfuUpload, tytPhys(bno), nil, 1024)
		if err != nil {
			return nil, info, err
		}
		if len(d) != 1024 {
			return nil, info, rErr("blocco %d incompleto", bno)
		}
		copy(img[bno*1024:], d)
		if err := s.getStatus(); err != nil {
			return nil, info, err
		}
	}
	if err := s.reboot(); err != nil {
		return nil, info, err
	}
	prog(nb, nb, "Lettura completata")
	return img, info, nil
}

func (f tytFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) != tytMemSize {
		return "", fmt.Errorf("i dati non sono una memoria della %s", m.Name)
	}
	nb := tytMemSize / 1024
	blocks := []uint32{0, 0x10000, 0x20000, 0x30000}
	for a := uint32(0x110000); a <= 0x1D0000; a += 0x10000 {
		blocks = append(blocks, a)
	}
	total := len(blocks) + nb
	prog(0, total, "Collegamento con la radio…")
	s, info, err := tytOpen(ctx, m)
	if err != nil {
		return info, err
	}
	defer s.p.Close()
	// cancellazione (come dfu_erase di dmrconfig)
	if err := s.getStatus(); err != nil {
		return info, err
	}
	if err := s.waitIdle(); err != nil {
		return info, err
	}
	if err := s.command(0x91, 0x01); err != nil {
		return info, err
	}
	time.Sleep(100 * time.Millisecond)
	for k, a := range blocks {
		prog(k, total, "Cancellazione della memoria dei canali…")
		if err := s.addrCmd(0x41, a); err != nil {
			return info, err
		}
	}
	if err := s.addrCmd(0x21, 0); err != nil {
		return info, err
	}
	for bno := 0; bno < nb; bno++ {
		prog(len(blocks)+bno, total, "Scrittura…")
		if _, err := s.ctl(false, dfuDnload, tytPhys(bno), img[bno*1024:(bno+1)*1024], 0); err != nil {
			return info, err
		}
		if err := s.getStatus(); err != nil {
			return info, err
		}
		if err := s.waitIdle(); err != nil {
			return info, err
		}
	}
	if err := s.reboot(); err != nil {
		return info, err
	}
	prog(total, total, "Scrittura completata")
	return info, nil
}
