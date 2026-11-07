// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware .bin della Radtel RT-890 (4.10): firmware di Radtel e firmware liberi
// (M7OCM, OEFW di DualTachyon e della comunità) con il caricatore della radio.
//
// Protocollo (riscritto da zero dalla descrizione e dal programma
// «rt_890_flasher.py» di JuantAldea, GNU GPL v3, e dalle note del firmware
// libero di DualTachyon «radtel-rt-890-oefw»):
//   - 115200 baud 8N1; la radio si accende nel caricatore tenendo premuti i due
//     tasti laterali (schermo spento);
//   - ogni comando termina con la somma dei suoi byte (modulo 256);
//   - lettura 52 00 00: nel caricatore la radio risponde con 0xFF;
//   - cancellazione 39 00 00 55: la radio risponde 0x06;
//   - scrittura 57, indirizzo (2 byte, alto prima), 128 byte: risponde 0x06;
//     l'ultimo blocco si completa con 00;
//   - memoria del programma: 0xEC00 byte da 0x08000000 (microcontrollore
//     Artery AT32F421, programma ARM Cortex-M4 in chiaro); il caricatore resta
//     in fondo alla memoria e non viene toccato.
// Il caricatore non controlla il modello: conferma scritta (RT890).

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"
)

const (
	rt890Baud     = 115200
	rt890Block    = 0x80
	rt890MaxLen   = 0xEC00
	rt890MinLen   = 8 * 1024
	rt890Ack      = 0x06
	rt890InBoot   = 0xFF
	rt890CmdRead  = 0x52
	rt890CmdErase = 0x39
	rt890CmdWrite = 0x57
)

var (
	rt890BootWait  = 60 * time.Second // attesa della radio accesa nel caricatore
	rt890EraseWait = 15 * time.Second
	rt890AckWait   = 3 * time.Second
)

// radio con questo caricatore
var rt890Targets = []fwTarget{
	{[]string{"RT890", "890"}, "Radtel RT-890", "RT890", []string{"ry-uv58plus", "rtl-rt890-m7ocm"}},
}

// fwKnownRT890: file provati (SHA-256)
var fwKnownRT890 = map[string]string{
	"d7cae27219bff4eb48c734157e92fd112570a240616d1cc87b6be9cfadcde749": "firmware libero M7OCM v3.2 per la scheda PCB 2.0 (m7ocm-v3.2-pcb2.0.bin, github.com/M7OCM/890)",
}

func rt890Sum(b []byte) []byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return append(b, s)
}

func rtU32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// rt890Plausible: programma ARM in chiaro per l'AT32F421 (tabella dei vettori:
// pila nella RAM da 16 KB, avvio e guasti nella memoria del programma)
func rt890Plausible(b []byte) bool {
	if len(b) < rt890MinLen || len(b) > rt890MaxLen {
		return false
	}
	sp := rtU32(b[0:])
	if sp < 0x20000400 || sp > 0x20004000 {
		return false
	}
	for _, off := range []int{4, 8, 12} { // avvio, NMI, guasto grave
		v := rtU32(b[off:])
		if v&1 == 0 || v < 0x08000000 || v >= 0x08000000+uint32(len(b)) {
			return false
		}
	}
	return true
}

// analyzeRT890 controlla un file .bin per la Radtel RT-890
func analyzeRT890(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("rt890", "Radtel RT-890", "RT890", name, origin, raw)
	info.Platform = "AT32F421"
	info.Typed = true
	if len(raw) < rt890MinLen || len(raw) > rt890MaxLen {
		add("bad", "Dimensione non plausibile (%s): il caricatore della RT-890 accetta al massimo %s", fmtBytes(len(raw)), fmtBytes(rt890MaxLen))
		return img
	}
	if !rt890Plausible(raw) {
		add("bad", "Il file non è un programma per il microcontrollore della RT-890 (tabella dei vettori non valida): non è un firmware utilizzabile")
		return img
	}
	add("ok", "Programma ARM in chiaro per il microcontrollore AT32F421 (avvio a 0x%08X), %s, %d blocchi da 128 byte",
		rtU32(raw[4:])&^1, fmtBytes(len(raw)), (len(raw)+rt890Block-1)/rt890Block)
	t := &rt890Targets[0]
	info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	low := strings.ToLower(path.Base(name))
	known, isKnown := fwKnownRT890[info.SHA256]
	free := bytes.Contains(raw, []byte("M7OCM")) || bytes.Contains(raw, []byte("OEFW")) || bytes.Contains(raw, []byte("DualTachyon")) ||
		strings.Contains(low, "m7ocm") || strings.Contains(low, "oefw") || strings.Contains(low, "mono")
	if free {
		add("warn", "Firmware libero (M7OCM / OEFW), non di Radtel: dopo l'aggiornamento i canali si gestiscono con la scheda «Radtel RT-890 (firmware M7OCM)» (memoria diversa dal firmware di Radtel)")
		info.Warnings = true
	}
	switch {
	case strings.Contains(low, "pcb2.0") || strings.Contains(low, "pcb20"):
		add("warn", "File per la scheda PCB 2.0: la RT-890 con la scheda PCB 2.1 (RT-890 II) usa un firmware diverso")
		info.Warnings = true
	case strings.Contains(low, "pcb2.1") || strings.Contains(low, "pcb21") || strings.Contains(low, "890-ii") || strings.Contains(low, "890ii"):
		add("warn", "File per la scheda PCB 2.1 (RT-890 II): la RT-890 con la scheda PCB 2.0 usa un firmware diverso")
		info.Warnings = true
	}
	if !strings.Contains(strings.ToUpper(normModel(low)), "890") && !isKnown && !free {
		add("warn", "Il nome del file non indica la RT-890: verifica che sia il firmware della tua radio")
		info.Warnings = true
	}
	add("warn", "Il caricatore della RT-890 non controlla a quale radio è destinato il firmware: scrivi il modello per confermare")
	add("ok", "Se l'aggiornamento si interrompe la radio resta nel caricatore: riaccendila tenendo premuti i due tasti laterali e ripeti l'invio")
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

// rt890Cmd invia un comando e aspetta la risposta di un byte
func rt890Cmd(ctx context.Context, p Port, cmd []byte, wait time.Duration) (byte, error) {
	if err := p.Write(rt890Sum(append([]byte{}, cmd...))); err != nil {
		return 0, err
	}
	b := make([]byte, 1)
	if err := p.ReadFull(ctx, b, wait); err != nil {
		return 0, err
	}
	return b[0], nil
}

// rt890Update invia il firmware alla RT-890 nel caricatore. started è chiamata
// prima della cancellazione (da lì la radio è modificata).
func rt890Update(ctx context.Context, port string, fw []byte, prog Progress, logf func(string), started func()) (int64, error) {
	if !rt890Plausible(fw) {
		return 0, fmt.Errorf("firmware non valido per la RT-890")
	}
	p, err := openPort(port, rt890Baud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	logf("In attesa della radio nel caricatore (radio spenta e riaccesa tenendo premuti i due tasti laterali)…")
	end := time.Now().Add(rt890BootWait)
	for {
		if ctx.Err() != nil {
			return 0, errCancelled
		}
		if time.Now().After(end) {
			return 0, rErr("la radio non ha risposto dal caricatore entro %d secondi: spegnila e riaccendila tenendo premuti i due tasti laterali", int(rt890BootWait.Seconds()))
		}
		r, err := rt890Cmd(ctx, p, []byte{rt890CmdRead, 0, 0}, 700*time.Millisecond)
		if err == errCancelled {
			return 0, err
		}
		if err == nil {
			// il resto della risposta (4 byte in tutto)
			readUpTo(ctx, p, 8, 200*time.Millisecond)
			if r == rt890InBoot {
				break
			}
			return 0, rErr("la radio ha risposto 0x%02X invece di 0x%02X: non è nel caricatore della RT-890. La radio non è stata modificata", r, rt890InBoot)
		}
	}
	logf("Radio nel caricatore: cancellazione della memoria del programma")
	started()
	r, err := rt890Cmd(ctx, p, []byte{rt890CmdErase, 0, 0, 0x55}, rt890EraseWait)
	if err != nil {
		if err == errCancelled {
			return 0, err
		}
		return 0, rErr("la radio non ha confermato la cancellazione")
	}
	if r != rt890Ack {
		return 0, rErr("la radio ha rifiutato la cancellazione (0x%02X)", r)
	}
	padded := append(append([]byte{}, fw...), make([]byte, (rt890Block-len(fw)%rt890Block)%rt890Block)...)
	nb := len(padded) / rt890Block
	var sent int64
	for n := 0; n < nb; n++ {
		if ctx.Err() != nil {
			return sent, errCancelled
		}
		prog(n*rt890Block, len(fw), "Invio del firmware…")
		off := n * rt890Block
		cmd := append([]byte{rt890CmdWrite, byte(off >> 8), byte(off)}, padded[off:off+rt890Block]...)
		r, err := rt890Cmd(ctx, p, cmd, rt890AckWait)
		if err != nil {
			if err == errCancelled {
				return sent, err
			}
			return sent, rErr("la radio non ha confermato il blocco %d di %d", n+1, nb)
		}
		if r != rt890Ack {
			return sent, rErr("la radio ha rifiutato il blocco %d di %d (0x%02X)", n+1, nb, r)
		}
		sent = int64(min(off+rt890Block, len(fw)))
	}
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi da 128 byte, tutti confermati dalla radio", nb))
	return sent, nil
}
