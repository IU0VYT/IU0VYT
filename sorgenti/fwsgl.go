// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware .sgl delle radio DMR della famiglia Radioddity GD-77 (4.10): Baofeng
// RD-5R (anche BF-5R, DM-5R Tier II) e Baofeng DM-1801, via USB (HID 15A2:0073)
// con il caricatore della radio.
//
// Formato e sequenza riscritti da zero dalle descrizioni e dal codice del
// «FirmwareLoader» di OpenGD77 (Roger Clark VK3KYY e altri, GNU GPL v2; usato da
// anni per scrivere sia OpenGD77 sia i firmware del costruttore) e di radio_tool
// (v0l, GNU GPL v3), che concordano:
//   file: «SGL!»; a 0x0C quattro byte (in XOR con «SGL!»): posizione della
//   seconda intestazione (2 byte) e la sua chiave XOR (2 byte); seconda
//   intestazione (in XOR con la chiave a 2 byte): +0x06 lunghezza del firmware,
//   +0x32 famiglia (16 byte: «BF-5R», «BF-DMR»), +0x42 modello (8: «BF-5R»,
//   «1801»), +0x4A versione del protocollo (8), +0x5F chiave del modello (8:
//   «DV02…», «DV03…»); il firmware cifrato sono gli ultimi «lunghezza» byte e si
//   invia così com'è (la radio lo decifra);
//   radio: accesa nel caricatore (schermo spento, indicatore acceso);
//   rapporti HID come per i canali (01 00 lunghezza dati; risposta 03 00
//   lunghezza dati): DOWNLOAD → «#UPDATE?»; A → A; chiave del modello → i primi
//   4 byte («DV02»: altrimenti il firmware è per un'altra radio); F-PROG,
//   famiglia, modello, versione, F-ERASE, A, PROGRAM → A; poi blocchi da 32
//   byte (indirizzo 4 byte e lunghezza 2 byte, alto prima) e, a ogni KB e alla
//   fine, «END» FF + somma dei byte del KB (4 byte, basso prima) → A.

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
	sglMaxLen   = 0x7B000
	sglChunk    = 0x20
	sglSumBlock = 0x400
)

// radio con questo caricatore: famiglia nell'intestazione → radio
var sglTargets = []struct {
	group, model, key string
	target, confirm   string
	drivers           []string
}{
	{"BF-5R", "BF-5R", "DV02", "Baofeng RD-5R / BF-5R", "RD5R", []string{"bf-rd5r"}},
	{"BF-DMR", "1801", "DV03", "Baofeng DM-1801", "DM1801", []string{"bf-dm1801"}},
	{"SG-MD-760", "MD-760", "DV01", "Radioddity GD-77 / TYT MD-760", "GD77", nil},
	{"SG-MD-730", "MD-730", "DV02", "Radioddity GD-77S / TYT MD-730", "GD77S", nil},
}

// fwKnownSGL: file del costruttore provati (SHA-256; pubblicati dal costruttore
// e raccolti in github.com/lxa520/GD-77)
var fwKnownSGL = map[string]string{
	"d4e21559511666d84636fe802af11538bf112abff0d85e9c32c5837983c7e68a": "Baofeng RD-5R, firmware V2.0.9 del costruttore (BF-5R_V2.0.9.sgl)",
	"60f7afc08effeac6a008fc272014ad5ba0bf882993f94a094ab58d5857e84062": "Baofeng RD-5R, firmware V2.1.0 del costruttore (BF-5R_V2.1.0.sgl)",
	"c31c536e5c3e43f6f3f27fd0eb1731bf5003d112d6ee64d77e5e3bd8416f6561": "Baofeng RD-5R, firmware V2.1.6 del costruttore (BF-5R_V2.1.6.sgl)",
	"7b1e3771ee1c083750d4061d5d743bbca04816b80efc558c01fc2cfccc01b418": "Baofeng DM-1801, firmware V2.1.9 del costruttore (DM-1801_V2.1.9.sgl)",
}

type sglInfo struct {
	group, model, proto string
	key                 []byte // chiave del modello (8 byte)
	fw                  []byte // firmware come nel file (cifrato)
}

func sglField(b []byte) []byte { return bytes.TrimRight(b, "\xff\x00") }

// sglParse legge le intestazioni di un file .sgl
func sglParse(raw []byte) (*sglInfo, error) {
	if len(raw) < 0x400 || !bytes.HasPrefix(raw, []byte("SGL!")) {
		return nil, fmt.Errorf("non è un firmware .sgl («SGL!» mancante)")
	}
	var h [4]byte
	for i := range h {
		h[i] = raw[0x0C+i] ^ "SGL!"[i]
	}
	off := int(h[0]) | int(h[1])<<8
	key := [2]byte{h[2], h[3]}
	if off < 0x10 || off+0x67 > len(raw) {
		return nil, fmt.Errorf("seconda intestazione fuori dal file")
	}
	h2 := make([]byte, 0x67)
	for i := range h2 {
		h2[i] = raw[off+i] ^ key[i%2]
	}
	n := int(binary.LittleEndian.Uint32(h2[0x06:]))
	if n <= 0 || n > len(raw)-0x400 {
		return nil, fmt.Errorf("lunghezza del firmware non valida (%d)", n)
	}
	return &sglInfo{
		group: string(sglField(h2[0x32:0x42])), model: string(sglField(h2[0x42:0x4A])),
		proto: string(sglField(h2[0x4A:0x52])), key: append([]byte{}, h2[0x5F:0x67]...),
		fw: raw[len(raw)-n:],
	}, nil
}

var reSglVer = regexp.MustCompile(`(?i)V(\d+(?:\.\d+)+)`)

// analyzeSGL controlla un file .sgl per RD-5R / DM-1801
func analyzeSGL(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("sgl", "Baofeng RD-5R o DM-1801", "", name, origin, raw)
	info.Platform = "NXP MK22"
	info.Typed = true
	s, err := sglParse(raw)
	if err != nil {
		add("bad", "File non valido: %v", err)
		return img
	}
	var t *struct {
		group, model, key string
		target, confirm   string
		drivers           []string
	}
	for i := range sglTargets {
		if sglTargets[i].group == s.group {
			t = &sglTargets[i]
		}
	}
	if t == nil {
		add("bad", "Firmware per una radio sconosciuta (famiglia %q, modello %q)", s.group, s.model)
		return img
	}
	info.Target, info.Confirm = t.target, t.confirm
	if t.drivers == nil {
		add("bad", "Firmware per la %s, che non è nella galleria di questo programma", t.target)
		return img
	}
	info.Models = t.drivers
	add("ok", "Firmware per la %s (famiglia %q, modello %q, protocollo %q)", t.target, s.group, s.model, s.proto)
	if !bytes.HasPrefix(s.key, []byte(t.key)) {
		add("bad", "Chiave del modello %q diversa da quella della %s (%s…): file danneggiato o per un'altra radio", string(sglField(s.key)), t.target, t.key)
		return img
	}
	if len(s.fw) > sglMaxLen {
		add("bad", "Firmware troppo grande (%s): il caricatore accetta al massimo %s", fmtBytes(len(s.fw)), fmtBytes(sglMaxLen))
		return img
	}
	add("ok", "Firmware cifrato di %s, %d blocchi da 32 byte con la somma di controllo a ogni KB: si invia così com'è (la radio lo decifra)",
		fmtBytes(len(s.fw)), (len(s.fw)+sglChunk-1)/sglChunk)
	if m := reSglVer.FindStringSubmatch(path.Base(name)); m != nil {
		info.Version = "V" + m[1]
	}
	if strings.Contains(strings.ToLower(name), "opengd77") {
		add("warn", "Firmware libero OpenGD77, non del costruttore: dopo l'aggiornamento la memoria e la programmazione cambiano (usa il programma di OpenGD77)")
		info.Warnings = true
	}
	if k, ok := fwKnownSGL[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware provati da questo programma")
	}
	add("ok", "La radio controlla la chiave del modello prima di cancellare la memoria: con il firmware di un'altra radio l'aggiornamento si ferma senza modificarla")
	add("warn", "Scrivi il modello per confermare")
	add("ok", "Se l'invio si interrompe la radio resta nel caricatore: riaccendila nella modalità di aggiornamento e ripeti l'invio")
	info.Warnings = true
	img.Dec = raw
	info.Valid = true
	return img
}

// ---------------------------------------------------------------- invio

type sglSess struct {
	ctx context.Context
	p   Port
}

func (s *sglSess) xfer(data []byte) ([]byte, error) {
	var buf [42]byte
	buf[0], buf[1] = 1, 0
	buf[2], buf[3] = byte(len(data)), byte(len(data)>>8)
	copy(buf[4:], data)
	if err := s.p.Write(buf[:]); err != nil {
		return nil, err
	}
	var rb [42]byte
	if err := s.p.ReadFull(s.ctx, rb[:], 5*time.Second); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	n := int(rb[2]) | int(rb[3])<<8
	if rb[0] != 3 || rb[1] != 0 || n > 38 {
		return nil, rErr("risposta della radio non valida (% X)", rb[:4])
	}
	return append([]byte{}, rb[4:4+n]...), nil
}

func sglPad(b []byte, n int) []byte {
	out := bytes.Repeat([]byte{0xFF}, max(n, len(b)))
	copy(out, b)
	return out
}

func (s *sglSess) ok(data []byte, what string) error {
	r, err := s.xfer(data)
	if err != nil {
		return err
	}
	if len(r) < 1 || r[0] != 'A' {
		return rErr("la radio non ha confermato %s (% X)", what, r)
	}
	return nil
}

// sglUpdate scrive il firmware nella radio collegata via USB nel caricatore.
// started è chiamata prima della cancellazione.
func sglUpdate(ctx context.Context, raw []byte, prog Progress, logf func(string), started func()) (int64, error) {
	si, err := sglParse(raw)
	if err != nil {
		return 0, err
	}
	p, err := openPort(fmt.Sprintf("HID:%04X:%04X", rdtyVID, rdtyPID), 0)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	s := &sglSess{ctx: ctx, p: p}
	logf("Radio collegata via USB: richiesta di aggiornamento")
	r, err := s.xfer([]byte("DOWNLOAD"))
	if err != nil {
		return 0, rErr("la radio non risponde nel caricatore (%v): accendila nella modalità di aggiornamento (schermo spento). La radio non è stata modificata", err)
	}
	if string(r) != "#UPDATE?" {
		return 0, rErr("la radio non è nella modalità di aggiornamento (risposta %q): spegnila e riaccendila tenendo premuti i tasti indicati. La radio non è stata modificata", string(r))
	}
	if err := s.ok([]byte("A"), "la richiesta"); err != nil {
		return 0, err
	}
	r, err = s.xfer(sglPad(si.key, 8))
	if err != nil {
		return 0, err
	}
	if len(r) == 0 || !bytes.HasPrefix(si.key, r) {
		return 0, rErr("la radio ha risposto %q alla chiave del modello %q: il firmware è per un'altra radio. La radio non è stata modificata", string(r), string(si.key[:4]))
	}
	logf(fmt.Sprintf("Chiave del modello accettata (%s): famiglia %s, modello %s", string(r), si.group, si.model))
	for _, c := range []struct {
		data []byte
		what string
	}{
		{sglPad([]byte("F-PROG"), 8), "F-PROG"},
		{sglPad([]byte(si.group), 16), "la famiglia della radio"},
		{sglPad([]byte(si.model), 8), "il modello"},
		{sglPad([]byte(si.proto), 8), "la versione del protocollo"},
	} {
		if err := s.ok(c.data, c.what); err != nil {
			return 0, err
		}
	}
	started()
	logf("Cancellazione della memoria del programma")
	prog(0, len(si.fw), "Cancellazione della memoria del programma…")
	if err := s.ok(sglPad([]byte("F-ERASE"), 8), "la cancellazione"); err != nil {
		return 0, err
	}
	if err := s.ok([]byte("A"), "la fine della cancellazione"); err != nil {
		return 0, err
	}
	// PROGRAM con 0x0F in fondo, come il FirmwareLoader di OpenGD77
	if err := s.ok([]byte("PROGRAM\x0f"), "l'inizio della scrittura"); err != nil {
		return 0, err
	}
	fw := si.fw
	var sent int64
	start := 0
	for a := 0; a < len(fw); {
		if ctx.Err() != nil {
			return sent, errCancelled
		}
		n := min(sglChunk, len(fw)-a)
		pk := make([]byte, 6, 6+n)
		binary.BigEndian.PutUint32(pk, uint32(a))
		binary.BigEndian.PutUint16(pk[4:], uint16(n))
		pk = append(pk, fw[a:a+n]...)
		if err := s.ok(pk, fmt.Sprintf("il blocco a 0x%05X", a)); err != nil {
			return sent, err
		}
		a += n
		sent = int64(a)
		if a%sglSumBlock == 0 || a == len(fw) {
			var sum uint32
			for _, b := range fw[start:a] {
				sum += uint32(b)
			}
			c := []byte{'E', 'N', 'D', 0xFF, 0, 0, 0, 0}
			binary.LittleEndian.PutUint32(c[4:], sum)
			if err := s.ok(c, fmt.Sprintf("la somma del KB a 0x%05X", start)); err != nil {
				return sent, err
			}
			start = a
			prog(a, len(fw), "Invio del firmware…")
		}
	}
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %s, %d KB confermati dalla radio", fmtBytes(len(fw)), (len(fw)+sglSumBlock-1)/sglSumBlock))
	return sent, nil
}
