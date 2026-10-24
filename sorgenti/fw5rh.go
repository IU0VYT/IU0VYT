// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware della Baofeng 5RH Pro e delle radio con la stessa
// elettronica (es. K67 Pro, programmate con il CPS «5RH_PRO» e i file .xlc),
// secondo il programma del costruttore "Radio.exe" (classe PortComm,
// DoWriteData), ricavato dal suo codice.
//
// File .dat: intestazione di 80 byte ("BaoFeng", "BF_5RH", "V1.0.0.0" a 0x20,
// "V2.0.0.0" a 0x30 per l'elettronica nuova, lunghezza dei dati a 0x40 in
// big-endian), poi i dati del firmware.
// Protocollo (115200 baud predefiniti, 38400 in alternativa; radio accesa in
// modalità aggiornamento):
//   "DOWNLOAD" → "#UPDATE?" (elettronica V1) oppure "V2_00_00" (elettronica V2):
//                deve corrispondere al file, altrimenti ci si ferma;
//   "A" → "A";  "F-ERASE" FF + 8 byte fissi → "A" (cancellazione);
//   "PROGRAM1" → "A";  blocchi: indirizzo (4 byte BE) 00 + fino a 1024 byte → "A";
//   "END" FF FF + somma dei byte inviati (4 byte BE) → "A".
// Ogni attesa: 5 secondi; nessun nuovo tentativo (come l'originale).

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	rh5Head  = 0x50
	rh5Block = 1024
	rh5Wait  = 5 * time.Second
	rh5Max   = 524288
)

var (
	rh5Erase = []byte{'F', '-', 'E', 'R', 'A', 'S', 'E', 0xFF, 40, 6, 136, 25, 19, 3, 24, 32}
	rh5V1    = []byte("#UPDATE?")
	rh5V2    = []byte("V2_00_00")
)

// rh5Size: lunghezza dei dati indicata nell'intestazione
func rh5Size(raw []byte) int64 {
	if len(raw) < rh5Head {
		return 0
	}
	return be32(raw[0x40:0x44])
}

func rh5IsV2(raw []byte) bool { return len(raw) > 0x31 && raw[0x31] == '2' }

type rh5Host struct {
	ctx context.Context
	p   Port
}

func (h *rh5Host) cmd(b []byte, n int, what string) ([]byte, error) {
	_ = h.p.Purge()
	if err := h.p.Write(b); err != nil {
		return nil, err
	}
	r := make([]byte, n)
	if err := h.p.ReadFull(h.ctx, r, rh5Wait); err != nil {
		if err == errTimeout {
			return nil, rErr("%s: la radio non risponde", what)
		}
		return nil, err
	}
	return r, nil
}

func (h *rh5Host) ack(b []byte, what string) error {
	r, err := h.cmd(b, 1, what)
	if err != nil {
		return err
	}
	if r[0] != 'A' {
		return rErr("%s: la radio non ha confermato (risposta %02X)", what, r[0])
	}
	return nil
}

// rh5Update invia il firmware .dat; started è chiamata prima della
// cancellazione della memoria della radio (da lì l'operazione non va interrotta).
func rh5Update(ctx context.Context, port string, raw []byte, baud int, prog Progress, logf func(string), started func()) (int64, error) {
	if logf == nil {
		logf = func(string) {}
	}
	if baud != 38400 {
		baud = 115200
	}
	size := rh5Size(raw)
	if size <= 0 || size > rh5Max {
		return 0, errors.New("lunghezza del firmware non valida nell'intestazione")
	}
	img := make([]byte, rh5Max) // come l'originale: memoria di 512 KB azzerata
	copy(img, raw)
	blocks := int((size + rh5Block - 1) / rh5Block)
	p, err := openPortFunc(port, baud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	h := &rh5Host{ctx: ctx, p: p}
	prog(0, blocks, "Collegamento con la radio…")
	logf("Richiesta di aggiornamento («DOWNLOAD»)…")
	r, err := h.cmd([]byte("DOWNLOAD"), 8, "Collegamento")
	if err != nil {
		return 0, err
	}
	want, other := rh5V1, rh5V2
	if rh5IsV2(raw) {
		want, other = rh5V2, rh5V1
	}
	if !bytes.Equal(r, want) {
		if bytes.Equal(r, other) {
			hw := "V1 (firmware 1.x)"
			if bytes.Equal(r, rh5V2) {
				hw = "V2 (firmware 2.x)"
			}
			return 0, rErr("la radio ha l'elettronica %s, diversa da quella a cui è destinato questo firmware: aggiornamento NON eseguito", hw)
		}
		return 0, rErr("risposta inattesa dalla radio (% X): è in modalità aggiornamento?", r)
	}
	logf(fmt.Sprintf("La radio ha risposto «%s»", r))
	if err := h.ack([]byte{'A'}, "Conferma"); err != nil {
		return 0, err
	}
	if ctx.Err() != nil {
		return 0, errCancelled
	}
	if started != nil {
		started()
	}
	logf("Cancellazione della memoria del firmware…")
	if err := h.ack(rh5Erase, "Cancellazione"); err != nil {
		return 0, err
	}
	if err := h.ack([]byte("PROGRAM1"), "Avvio della scrittura"); err != nil {
		return 0, err
	}
	logf(fmt.Sprintf("Scrittura: %d blocchi", blocks))
	var sum uint32
	pos := int64(rh5Head)
	var a int64
	for k := 0; a < size; k++ {
		n := int64(rh5Block) - a%rh5Block
		if a+rh5Block > size {
			n = size - a
		}
		blk := make([]byte, 5+n)
		blk[0], blk[1], blk[2], blk[3] = byte(a>>24), byte(a>>16), byte(a>>8), byte(a)
		copy(blk[5:], img[pos:pos+n])
		for _, v := range blk[5:] {
			sum += uint32(v)
		}
		pos += n
		prog(k, blocks, "Invio del firmware…")
		if err := h.ack(blk, fmt.Sprintf("Blocco 0x%05X", a)); err != nil {
			return a, err
		}
		a += n
	}
	end := []byte{'E', 'N', 'D', 0xFF, 0xFF, byte(sum >> 24), byte(sum >> 16), byte(sum >> 8), byte(sum)}
	if err := h.ack(end, "Fine (controllo della somma)"); err != nil {
		return a, err
	}
	logf("La radio ha confermato la fine dell'aggiornamento")
	prog(blocks, blocks, "Aggiornamento completato")
	return a, nil
}
