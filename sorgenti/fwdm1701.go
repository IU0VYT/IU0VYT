// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware della Baofeng DM-1701 / Retevis RT84 (4.10), file .bin del costruttore
// («OutSecurityBin»), via USB nel caricatore DFU della radio.
//
// Formato e sequenza riscritti da zero dalla descrizione e dal codice di
// radio_tool (https://github.com/v0l/radio_tool, GNU GPL v3; il firmware della
// DM-1701 è stato scritto con quella sequenza su radio vere), confrontati con
// md380tools e dmrconfig (stesso collegamento DFU usato qui per i canali):
//   file: 0x000 «OutSecurityBin», 0x010 modello («DM1701»), 0x030 numero del
//   modello (lunghezza + byte: 01 0F = DM-1701), 0x07C numero di aree
//   (FFFFFFFF = 1), 0x080 tabella delle aree (indirizzo, lunghezza), 0x100 dati
//   delle aree uno dopo l'altro, cifrati: alla radio si inviano così come sono
//   (il caricatore li decifra da sé), in fondo «OutputBinDataEnd»;
//   radio: accesa tenendo premuti PTT e il tasto sopra il PTT (caricatore,
//   indicatore lampeggiante); lettura del modello (A2 01 + lettura), comando
//   91 31 (inizio dell'aggiornamento), cancellazione dei settori toccati dalle
//   aree (41 + indirizzo), poi per ogni settore: indirizzo (21 + indirizzo) e
//   blocchi da 1024 byte numerati da 2; ogni comando si conclude con
//   GETSTATUS (occupato, attesa indicata dalla radio, pronto).
// Memoria scrivibile: 0x0800C000-0x080FFFFF (settori da 16, 64 e 7 × 128 KB,
// come dichiara il caricatore: «@Internal Flash /0x0800C000/01*016Kg,01*064Kg,
// 07*128Kg»); il caricatore (0x08000000-0x0800BFFF) non si tocca mai. Le aree
// del file per la memoria SPI (caratteri, immagini) non si scrivono con questa
// procedura e si saltano.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	dm1701FwMagic   = "OutSecurityBin"
	dm1701FwEnd     = "OutputBinDataEnd"
	dm1701FwData    = 0x100
	dm1701FwBlock   = 1024
	dm1701FlashLo   = 0x0800C000
	dm1701FlashHi   = 0x08100000
	dm1701DfuPort   = "DFU:0483:DF11"
	dm1701FwConfirm = "DM1701"
)

// settori scrivibili (STM32F405: 3 da 16 KB, 4 da 64 KB, 5-11 da 128 KB)
var dm1701Sectors = func() [][2]uint32 {
	s := [][2]uint32{{0x0800C000, 0x4000}, {0x08010000, 0x10000}}
	for a := uint32(0x08020000); a < dm1701FlashHi; a += 0x20000 {
		s = append(s, [2]uint32{a, 0x20000})
	}
	return s
}()

// numeri dei modelli della famiglia TYT nei file del costruttore (radio_tool)
var tytFwMagics = []struct {
	magic []byte
	radio string
}{
	{[]byte{0x01, 0x0F}, "Baofeng DM-1701 / Retevis RT84"},
	{[]byte{0x02, 0x19, 0x0C}, "TYT MD-2017"}, {[]byte{0x02, 0x18, 0x0C}, "TYT MD-2017 GPS"},
	{[]byte{0x01, 0x19}, "TYT MD-2017"}, {[]byte{0x01, 0x18}, "TYT MD-2017 GPS"},
	{[]byte{0x01, 0x14}, "TYT MD-9600"}, {[]byte{0x02, 0x16, 0x0C}, "TYT MD-UV380 / UV390 GPS"},
	{[]byte{0x02, 0x17, 0x0C}, "TYT MD-UV380 / UV390"}, {[]byte{0x01, 0x10}, "TYT MD-390"},
	{[]byte{0x01, 0x0D}, "TYT MD-380 / MD-446"}, {[]byte{0x01, 0x1B}, "TYT MD-280"},
}

// fwKnownDM1701: file del costruttore provati (SHA-256)
var fwKnownDM1701 = map[string]string{
	"0314efd4c62625661a2d693261ae682dff8a1b075a30620f3dbda4be065de32d": "Baofeng DM-1701, firmware V02.06 con la rubrica dei contatti (CSV) (DM-1701(CSV)-V02.06.bin)",
	"4230c8d0aabf15f3f165c29a118a5f8f9b906a2d9b7d215c1d8a7bc0e3978bd9": "Baofeng DM-1701, firmware V02.06 senza rubrica (noCSV) (DM-1701(noCSV)-V02.06.bin)",
}

var reDM1701Ver = regexp.MustCompile(`(?i)V?(\d{1,2})\.(\d{2})`)

type dm1701Region struct {
	addr uint32
	data []byte // come nel file (cifrati)
}

// dm1701Parse legge le aree di un file «OutSecurityBin»
func dm1701Parse(raw []byte) (model string, magic []byte, regs []dm1701Region, err error) {
	if len(raw) < dm1701FwData+16 || !bytes.HasPrefix(raw, []byte(dm1701FwMagic)) {
		return "", nil, nil, fmt.Errorf("non è un firmware TYT («OutSecurityBin» mancante)")
	}
	model = strings.TrimRight(string(bytes.SplitN(raw[0x10:0x20], []byte{0}, 2)[0]), "\xff")
	n := int(raw[0x30])
	if n < 1 || n > 3 {
		return model, nil, nil, fmt.Errorf("numero del modello non valido")
	}
	magic = append([]byte{}, raw[0x30:0x31+n]...)
	cnt := binary.LittleEndian.Uint32(raw[0x7C:])
	if cnt == 0xFFFFFFFF {
		cnt = 1
	}
	if cnt == 0 || cnt > 16 {
		return model, magic, nil, fmt.Errorf("numero di aree non valido (%d)", cnt)
	}
	off := dm1701FwData
	for i := 0; i < int(cnt); i++ {
		a := binary.LittleEndian.Uint32(raw[0x80+8*i:])
		l := int(binary.LittleEndian.Uint32(raw[0x84+8*i:]))
		if l <= 0 || l > len(raw) || off+l > len(raw) {
			return model, magic, nil, fmt.Errorf("area %d fuori dal file", i+1)
		}
		regs = append(regs, dm1701Region{a, raw[off : off+l]})
		off += l
	}
	return model, magic, regs, nil
}

func dm1701Writable(r dm1701Region) bool {
	end := uint64(r.addr) + uint64(len(r.data))
	return r.addr >= dm1701FlashLo && end <= dm1701FlashHi
}

// analyzeDM1701 controlla un file .bin per la DM-1701
func analyzeDM1701(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("dm1701", "Baofeng DM-1701 / Retevis RT84", dm1701FwConfirm, name, origin, raw)
	info.Platform = "STM32F405"
	info.Typed = true
	info.Models = []string{"bf-dm1701"}
	model, magic, regs, err := dm1701Parse(raw)
	if err != nil {
		add("bad", "File non valido: %v", err)
		return img
	}
	radio := ""
	for _, t := range tytFwMagics {
		if bytes.Equal(t.magic, magic) {
			radio = t.radio
			break
		}
	}
	switch {
	case radio == "":
		add("bad", "Firmware TYT per una radio sconosciuta (modello %q, numero % X): non è per la DM-1701", model, magic)
		return img
	case !bytes.Equal(magic, tytFwMagics[0].magic):
		add("bad", "Firmware per la %s (modello %q): non è per la DM-1701. Il caricatore lo accetterebbe ma la radio non si avvierebbe più", radio, model)
		return img
	}
	add("ok", "Firmware del costruttore per la DM-1701 (modello %q, numero del modello % X)", model, magic)
	if !bytes.HasSuffix(bytes.TrimRight(raw, "\x00"), []byte(dm1701FwEnd)) {
		add("warn", "Manca la chiusura «%s» del file: potrebbe essere incompleto", dm1701FwEnd)
		info.Warnings = true
	}
	var w, skip int
	for _, r := range regs {
		if dm1701Writable(r) {
			w += len(r.data)
			add("ok", "Area 0x%08X-0x%08X (%s): memoria del programma, si scrive", r.addr, r.addr+uint32(len(r.data)), fmtBytes(len(r.data)))
		} else {
			skip++
			add("warn", "Area 0x%08X (%s) fuori dalla memoria del programma (es. caratteri e immagini nella memoria SPI): non si scrive con questa procedura", r.addr, fmtBytes(len(r.data)))
		}
	}
	if w == 0 {
		add("bad", "Nessuna area del file va nella memoria del programma della radio (0x%08X-0x%08X)", dm1701FlashLo, dm1701FlashHi-1)
		return img
	}
	if skip > 0 {
		add("warn", "Come radio_tool, si scrive solo il programma: caratteri e immagini della memoria SPI restano quelli della radio. Se il costruttore indica che la nuova versione li cambia, usa il suo programma «DMR Firmware DownLoad Setup»")
		info.Warnings = true
	}
	base := path.Base(name)
	if m := reDM1701Ver.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1] + "." + m[2]
	}
	switch u := strings.ToUpper(base); {
	case strings.Contains(u, "NOCSV") || strings.Contains(u, "NO CSV") || strings.Contains(u, "NO_CSV"):
		add("ok", "Versione senza rubrica dei contatti (noCSV)")
	case strings.Contains(u, "CSV"):
		add("ok", "Versione con la rubrica dei contatti (CSV)")
	}
	if k, ok := fwKnownDM1701[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware provati da questo programma")
	}
	add("warn", "Il caricatore della DM-1701 non controlla il modello del firmware: scrivi il modello per confermare")
	add("ok", "Il caricatore resta intatto: se l'invio si interrompe, riaccendi la radio nel caricatore (PTT e tasto sopra il PTT) e ripeti l'aggiornamento")
	info.Warnings = true
	if info.Version == "" {
		info.Version = strings.TrimSpace(model)
	}
	img.Dec = raw
	info.Valid = true
	return img
}

// ---------------------------------------------------------------- invio

// dm1701Dnload: DNLOAD come radio_tool (stato pronto, comando, GETSTATUS: occupato,
// attesa indicata dalla radio, GETSTATUS: pronto)
func (s *dfuSess) dm1701Dnload(value uint16, data []byte, what string) error {
	if err := s.dm1701Settle(); err != nil {
		return err
	}
	if _, err := s.ctl(false, dfuDnload, value, data, 0); err != nil {
		return err
	}
	for tries := 0; tries < 400; tries++ {
		st, err := s.ctl(true, dfuGetStatus, 0, nil, 6)
		if err != nil {
			return err
		}
		if len(st) < 5 {
			return rErr("stato DFU non valido (%s)", what)
		}
		if st[0] != 0 {
			return rErr("la radio ha rifiutato %s (errore DFU %d)", what, st[0])
		}
		switch st[4] {
		case 5, dfuStIdle: // dnload idle
			return nil
		case dfuStDnBusy:
			wait := time.Duration(int(st[1])|int(st[2])<<8|int(st[3])<<16) * time.Millisecond
			if wait < 5*time.Millisecond {
				wait = 5 * time.Millisecond
			}
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			time.Sleep(wait)
		case dfuStError:
			return rErr("la radio ha rifiutato %s", what)
		default:
			return rErr("stato DFU inatteso %d (%s)", st[4], what)
		}
		if s.ctx.Err() != nil {
			return errCancelled
		}
	}
	return rErr("la radio non ha concluso %s", what)
}

func (s *dfuSess) dm1701Settle() error {
	for tries := 0; tries < 20; tries++ {
		st, err := s.ctl(true, dfuGetState, 0, nil, 1)
		if err != nil {
			return err
		}
		if len(st) != 1 {
			return rErr("stato DFU non valido")
		}
		switch st[0] {
		case dfuStIdle, 5:
			return nil
		case dfuStError:
			_, err = s.ctl(false, dfuClrStatus, 0, nil, 0)
		default:
			_, err = s.ctl(false, dfuAbort, 0, nil, 0)
		}
		if err != nil {
			return err
		}
	}
	return rErr("la radio non torna pronta (DFU)")
}

func dm1701Addr(op byte, a uint32) []byte {
	return []byte{op, byte(a), byte(a >> 8), byte(a >> 16), byte(a >> 24)}
}

// dm1701Pieces: parti di un'area divise per settore
func dm1701Pieces(r dm1701Region) [][3]uint32 { // indirizzo, inizio nell'area, lunghezza
	var out [][3]uint32
	end := r.addr + uint32(len(r.data))
	for _, s := range dm1701Sectors {
		lo, hi := max(r.addr, s[0]), min(end, s[0]+s[1])
		if lo < hi {
			out = append(out, [3]uint32{lo, lo - r.addr, hi - lo})
		}
	}
	return out
}

// dm1701Update scrive il firmware nella radio collegata via USB nel caricatore.
// started è chiamata prima della prima cancellazione.
func dm1701Update(ctx context.Context, raw []byte, prog Progress, logf func(string), started func()) (int64, string, error) {
	_, _, regs, err := dm1701Parse(raw)
	if err != nil {
		return 0, "", err
	}
	var w []dm1701Region
	total := 0
	for _, r := range regs {
		if dm1701Writable(r) {
			w = append(w, r)
			total += len(r.data)
		}
	}
	if len(w) == 0 {
		return 0, "", rErr("nessuna area da scrivere")
	}
	p, err := openPort(dm1701DfuPort, 0)
	if err != nil {
		return 0, "", err
	}
	defer p.Close()
	s := &dfuSess{ctx: ctx, p: p}
	logf("Radio collegata via USB: lettura del modello")
	if err := s.dm1701Dnload(0, []byte{0xA2, 0x01}, "la lettura del modello"); err != nil {
		return 0, "", rErr("la radio non risponde nel caricatore USB (%v): accendila tenendo premuti il PTT e il tasto sopra il PTT. La radio non è stata modificata", err)
	}
	if err := s.dm1701Settle(); err != nil {
		return 0, "", err
	}
	idb, err := s.ctl(true, dfuUpload, 0, nil, 64)
	if err != nil {
		// come radio_tool: senza il modello si prosegue (conferma scritta)
		logf(fmt.Sprintf("Modello della radio non letto (%v): si prosegue con la conferma scritta", err))
		idb = nil
	}
	if n := bytes.IndexByte(idb, 0); n >= 0 {
		idb = idb[:n]
	}
	id := strings.TrimSpace(string(idb))
	if id != "" && !strings.Contains(strings.ReplaceAll(strings.ToUpper(id), "-", ""), "1701") {
		return 0, id, rErr("la radio si identifica come %q, non come DM-1701: il firmware non è per questa radio. La radio non è stata modificata", id)
	}
	logf(fmt.Sprintf("Radio: %q. Inizio dell'aggiornamento", id))
	if err := s.dm1701Dnload(0, []byte{0x91, 0x31}, "l'inizio dell'aggiornamento"); err != nil {
		return 0, id, err
	}
	started()
	for _, r := range w {
		for _, pc := range dm1701Pieces(r) {
			prog(0, total, "Cancellazione della memoria del programma…")
			logf(fmt.Sprintf("Cancellazione del settore a 0x%08X", pc[0]))
			if err := s.dm1701Dnload(0, dm1701Addr(0x41, pc[0]), fmt.Sprintf("la cancellazione a 0x%08X", pc[0])); err != nil {
				return 0, id, err
			}
		}
	}
	var sent int64
	for _, r := range w {
		for _, pc := range dm1701Pieces(r) {
			if err := s.dm1701Dnload(0, dm1701Addr(0x21, pc[0]), fmt.Sprintf("l'indirizzo 0x%08X", pc[0])); err != nil {
				return sent, id, err
			}
			part := r.data[pc[1] : pc[1]+pc[2]]
			for k := 0; k*dm1701FwBlock < len(part); k++ {
				blk := part[k*dm1701FwBlock : min((k+1)*dm1701FwBlock, len(part))]
				if err := s.dm1701Dnload(uint16(k+2), blk, fmt.Sprintf("il blocco a 0x%08X", pc[0]+uint32(k*dm1701FwBlock))); err != nil {
					return sent, id, err
				}
				sent += int64(len(blk))
				prog(int(sent), total, "Invio del firmware…")
			}
		}
	}
	prog(total, total, "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %s in %d aree", fmtBytes(int(sent)), len(w)))
	return sent, id, nil
}
