// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware .BIN delle radio TIDRADIO TD-H3 / TD-H8 e Radtel RT-730 (4.3).
//
// Protocollo del caricatore (scritto da zero da: documentazione pubblica di
// nicsure, «TD-H3-Engineering/firmware_flash_protocol.txt», con la cattura
// delle porte del programma del costruttore; codice di «TID Station»; stessa
// presentazione di 36 byte nei programmi di aggiornamento del costruttore
// «H8_IAP» di TIDRADIO e «RT730 Update tool» (HW-8800.exe) di Radtel):
//   - 115200 baud 8N1; la radio accesa tenendo premuto il PTT invia di continuo
//     il byte 0xA5;
//   - il programma invia la presentazione A0 EE 74 71 07 74 + 30 × 0x55 (il
//     quarto byte è la somma dei byte 4…35);
//   - la radio continua a inviare 0xA5; quando smette il programma ha circa
//     mezzo secondo per cominciare;
//   - blocchi da 32 byte: A1 (A2 per l'ultimo), numero del blocco (2 byte, alto
//     prima), somma dei 32 byte, dati (l'ultimo completato con 00); la radio
//     risponde 0xA3 a ogni blocco.
// Il caricatore non controlla il modello del firmware: conferma scritta.

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	tdhBaud   = 115200
	tdhBlock  = 32
	tdhMinLen = 10000
	tdhMaxLen = 65536
	tdhReady  = 0xA5
	tdhAck    = 0xA3
)

// tdhHello: presentazione dei programmi del costruttore
var tdhHello = func() []byte {
	h := []byte{0xA0, 0xEE, 0x74, 0x00, 0x07, 0x74}
	h = append(h, bytes.Repeat([]byte{0x55}, 30)...)
	var s byte
	for _, b := range h[4:] {
		s += b
	}
	h[3] = s
	return h
}()

var (
	tdhBootWait = 60 * time.Second       // attesa della radio accesa con il PTT
	tdhQuiet    = 250 * time.Millisecond // fine dei byte 0xA5 dopo la presentazione
	tdhAckWait  = 2 * time.Second
)

// radio con questo caricatore (l'ordine conta)
var tdhTargets = []fwTarget{
	{[]string{"RT730"}, "Radtel RT-730", "RT730", []string{"rt-rt730"}},
	{[]string{"TDH3", "H3"}, "TIDRADIO TD-H3", "TDH3", []string{"td-tdh3", "td-tdh3ham", "td-tdh3gmrs"}},
	{[]string{"TDH8", "H8"}, "TIDRADIO TD-H8", "TDH8", []string{"td-tdh8", "td-tdh8ham", "td-tdh8gmrs", "td-tdh8g3", "td-tdh8g3ham", "td-tdh8g3gmrs"}},
}

// fwKnownTDH: file del costruttore provati (SHA-256)
var fwKnownTDH = map[string]string{
	"358970c5aa86af42e4a722d11804dfb8779c3950a8fd6a5574a35025eee9a995": "Radtel RT-730, firmware del 31/07/2025 con il passo di 8,33 kHz migliorato (730Step improvement.BIN, radtels.com)",
}

// tdhTarget: radio indicata dal nome (i file di Radtel cominciano con «730»)
func tdhTarget(name string) *fwTarget {
	if strings.HasPrefix(normModel(path.Base(name)), "730") {
		return &tdhTargets[0]
	}
	return fwTargetFromName(tdhTargets, name)
}

// tdhPlausible: programma per microcontrollore 8051 (salto LJMP all'inizio)
func tdhPlausible(b []byte) bool {
	if len(b) < tdhMinLen || len(b) > tdhMaxLen || b[0] != 0x02 {
		return false
	}
	t := int(b[1])<<8 | int(b[2])
	return t >= 3 && t < len(b)
}

var reTdhDate = regexp.MustCompile(`(?:^|[^0-9])(2[0-9])(0[1-9]|1[0-2])([0-2][0-9]|3[01])(?:[^0-9]|$)`)

// analyzeTDH controlla un file .BIN per TD-H3 / TD-H8 / RT-730
func analyzeTDH(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("tdh", "TIDRADIO TD-H3 / TD-H8 o Radtel RT-730", "TDH", name, origin, raw)
	info.Platform = "8051"
	info.Typed = true
	if len(raw) < tdhMinLen || len(raw) > tdhMaxLen {
		add("bad", "Dimensione non plausibile (%s): il caricatore di queste radio accetta da 10 KB a 64 KB", fmtBytes(len(raw)))
		return img
	}
	if !tdhPlausible(raw) {
		add("bad", "L'inizio del file non è un programma per il microcontrollore di queste radio (manca il salto iniziale): non è un firmware utilizzabile")
		return img
	}
	add("ok", "Programma per microcontrollore 8051 (avvio a 0x%04X), %s, %d blocchi da 32 byte", int(raw[1])<<8|int(raw[2]), fmtBytes(len(raw)), (len(raw)+tdhBlock-1)/tdhBlock)
	known, isKnown := fwKnownTDH[info.SHA256]
	t := tdhTarget(name)
	if isKnown {
		t = tdhTarget(strings.Fields(strings.TrimPrefix(known, "Radtel "))[0])
	}
	if t != nil {
		info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	}
	// data della versione: nel nome (H3_240530) o nel testo del firmware (:250731)
	base := strings.TrimSuffix(path.Base(name), path.Ext(name))
	m := reTdhDate.FindStringSubmatch(base)
	if m == nil {
		if i := bytes.Index(raw, []byte(":2")); i >= 0 && i+7 <= len(raw) {
			m = reTdhDate.FindStringSubmatch(string(raw[i+1 : i+7]))
		}
	}
	if m != nil {
		info.Date = m[3] + "/" + m[2] + "/20" + m[1]
		info.Version = m[1] + m[2] + m[3]
	}
	switch {
	case t == nil:
		add("warn", "Il nome del file non indica la radio (730, H3 o H8): verifica che sia il firmware della tua radio")
		info.Warnings = true
	default:
		add("ok", "Firmware per %s", t.target)
	}
	if strings.Contains(strings.ToUpper(base), "NIC") {
		add("warn", "Firmware modificato da terzi (nicsure): non è del costruttore e questo programma, come CHIRP, potrebbe non leggerne più i canali")
		info.Warnings = true
	}
	add("warn", "Il caricatore di queste radio non controlla a quale modello è destinato il firmware: TD-H3, TD-H8 e RT-730 usano file diversi. Scrivi il modello per confermare")
	add("ok", "Se l'aggiornamento si interrompe la radio resta nel caricatore: riaccendila tenendo premuto il PTT e ripeti l'invio")
	info.Warnings = true
	if isKnown {
		info.Known = known
		add("ok", "File conosciuto: %s", known)
	} else {
		add("warn", "File non presente nell'elenco dei firmware provati da questo programma")
	}
	img.Dec = raw
	info.Valid = true
	return img
}

// tdhPacket: blocco n del firmware (ultimo: A2)
func tdhPacket(fw []byte, n int) []byte {
	p := make([]byte, 4+tdhBlock)
	off := n * tdhBlock
	copy(p[4:], fw[off:min(off+tdhBlock, len(fw))])
	p[0] = 0xA1
	if off+tdhBlock >= len(fw) {
		p[0] = 0xA2
	}
	p[1], p[2] = byte(n>>8), byte(n)
	var s byte
	for _, b := range p[4:] {
		s += b
	}
	p[3] = s
	return p
}

// tdhUpdate invia il firmware alla radio accesa nel caricatore. started è
// chiamata prima del primo blocco.
func tdhUpdate(ctx context.Context, port string, fw []byte, prog Progress, logf func(string), started func()) (int64, error) {
	if len(fw) < tdhMinLen || len(fw) > tdhMaxLen {
		return 0, fmt.Errorf("dimensione del firmware non valida")
	}
	p, err := openPort(port, tdhBaud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	logf("In attesa della radio nel caricatore (radio spenta e riaccesa tenendo premuto il PTT)…")
	b := make([]byte, 1)
	end := time.Now().Add(tdhBootWait)
	for {
		if ctx.Err() != nil {
			return 0, errCancelled
		}
		if time.Now().After(end) {
			return 0, rErr("la radio non si è presentata nel caricatore entro %d secondi: spegnila e riaccendila tenendo premuto il PTT", int(tdhBootWait.Seconds()))
		}
		err := p.ReadFull(ctx, b, 500*time.Millisecond)
		if err == errCancelled {
			return 0, err
		}
		if err == nil && b[0] == tdhReady {
			break
		}
	}
	logf("Radio nel caricatore: invio della presentazione")
	if err := p.Write(tdhHello); err != nil {
		return 0, err
	}
	// la radio ripete 0xA5 ancora per poco: si comincia quando smette
	qend := time.Now().Add(5 * time.Second)
	for {
		err := p.ReadFull(ctx, b, tdhQuiet)
		if err == errCancelled {
			return 0, err
		}
		if err != nil {
			break // silenzio: la radio aspetta il firmware
		}
		if b[0] != tdhReady {
			return 0, rErr("risposta inattesa della radio alla presentazione (0x%02X): non è una radio con questo caricatore. La radio non è stata modificata", b[0])
		}
		if time.Now().After(qend) {
			return 0, rErr("la radio non ha accettato la presentazione (continua a inviare 0xA5): non è una radio con questo caricatore. La radio non è stata modificata")
		}
	}
	started()
	var sent int64
	nb := (len(fw) + tdhBlock - 1) / tdhBlock
	for n := 0; n < nb; n++ {
		if ctx.Err() != nil {
			return sent, errCancelled
		}
		prog(n*tdhBlock, len(fw), "Invio del firmware…")
		if err := p.Write(tdhPacket(fw, n)); err != nil {
			return sent, err
		}
		aend := time.Now().Add(tdhAckWait)
		for {
			left := time.Until(aend)
			if left <= 0 {
				return sent, rErr("la radio non ha confermato il blocco %d di %d", n+1, nb)
			}
			if err := p.ReadFull(ctx, b, left); err != nil {
				if err == errCancelled {
					return sent, err
				}
				return sent, rErr("la radio non ha confermato il blocco %d di %d", n+1, nb)
			}
			if b[0] == tdhAck {
				break
			}
		}
		sent = int64(min((n+1)*tdhBlock, len(fw)))
	}
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi da 32 byte, tutti confermati dalla radio", nb))
	return sent, nil
}
