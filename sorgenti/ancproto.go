// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware della Bajeton 8300 (BJ-8300) via cavo, secondo il
// programma originale del costruttore "BinTe ANC Bootloader" V1.1 (BinTalk,
// classe KDH_Bootloader.BootHelper), ricavato dal suo codice: stessi pacchetti,
// stesse velocità, stesse attese, stessi nuovi tentativi.
//
// Pacchetto: AA 55, lunghezza dei dati (2 byte, LE), dati, CRC16 XMODEM
// (polinomio 0x1021, valore iniziale 0) calcolato su AA 55 + lunghezza + dati,
// scritto LE. Il primo byte dei dati è il comando:
//   06 PC→radio  entra in modalità aggiornamento (radio accesa normalmente)
//   01 PC→radio  velocità: [01, baud LE32];  radio→PC: risposta/presenza
//   02 radio→PC  richiesta dati: [02, indirizzo LE32, lunghezza LE32]
//      PC→radio  [02, indirizzo, byte validi LE32, dati]; i dati hanno sempre la
//                lunghezza richiesta (zeri oltre la fine del file)
//   04 radio→PC  dati finiti → PC→radio: 03 (stop)
//   03 radio→PC  aggiornamento completato
// Sequenza "Update" (radio accesa): porta a 9600 baud, 06 → 01;
// [01,10000] e passaggio a 10000 baud → 01; [01,10000] → 01; [01,baud] e
// passaggio a baud (115200 predefinita, o 57600); da qui la radio chiede i dati.
// Sequenza di recupero (pulsante "Click Here If Your Radio Can't Work"): porta a
// 9600 baud senza inviare nulla; quando la radio si presenta con 01 si invia
// [01,baud] e si passa a baud.
// Attesa di ogni risposta: 2 secondi, poi nuovo invio dell'ultimo pacchetto, al
// massimo 5 volte (il conteggio riparte a ogni risposta della radio).
// Il file .anc è cifrato a blocchi di 1024 byte: il programma lo decifra in
// memoria e invia alla radio i dati in chiaro (vedi ancDecrypt).

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	ancBaudStart = 9600
	ancBaudMid   = 10000
	ancWait      = 2 * time.Second
	ancRetries   = 5
	ancMaxRx     = 10240 // buffer di ricezione del programma originale
)

var (
	ancKeyA = []byte("KDHT")
	ancKeyB = []byte("RBGI")
)

var errAncBad = errors.New("pacchetto non valido")

// ancDecrypt decifra (o cifra: l'operazione è simmetrica) un file .anc come
// BootHelper.Decrypt: blocchi da 1024 byte; i primi due e gli ultimi due non
// sono cifrati; degli altri, a rotazione, uno resta in chiaro, uno usa la chiave
// "KDHT" e uno "RBGI". XOR byte per byte con la chiave a rotazione (la posizione
// nella chiave prosegue da un blocco cifrato al successivo), saltando i byte 0x00,
// 0xFF, uguali alla chiave o al suo complemento.
func ancDecrypt(src []byte) []byte {
	b := append([]byte{}, src...)
	n := len(b)
	blocks := n / 1024
	if n%1024 != 0 {
		blocks++
	}
	k, sel := 0, 0
	for i, off := 0, 0; off < n; i, off = i+1, off+1024 {
		end := off + 1024
		if end > n {
			end = n
		}
		if !(i < 2 || i == blocks-1 || i == blocks-2) {
			var key []byte
			switch sel {
			case 1:
				key = ancKeyA
			case 2:
				key = ancKeyB
			}
			if key != nil {
				for j := off; j < end; j++ {
					v, s := b[j], key[k]
					if v != 0 && v != 0xFF && v != s && v != s^0xFF {
						b[j] = v ^ s
					}
					k = (k + 1) % 4
				}
			}
		}
		sel = (sel + 1) % 3
	}
	return b
}

func crc16x(b []byte) uint16 {
	var c uint16
	for _, v := range b {
		c ^= uint16(v) << 8
		for i := 0; i < 8; i++ {
			if c&0x8000 != 0 {
				c = c<<1 ^ 0x1021
			} else {
				c <<= 1
			}
		}
	}
	return c
}

// ancPack: PackageFmt.Packing
func ancPack(payload []byte) []byte {
	n := len(payload)
	out := make([]byte, n+6)
	out[0], out[1], out[2], out[3] = 0xAA, 0x55, byte(n), byte(n>>8)
	copy(out[4:], payload)
	c := crc16x(out[:n+4])
	out[n+4], out[n+5] = byte(c), byte(c>>8)
	return out
}

func le32(b []byte) int64 {
	return int64(int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24))
}

type ancOpts struct {
	Recovery bool // radio già in modalità aggiornamento (non risponde normalmente)
	Baud     int  // velocità del trasferimento: 115200 (predefinita) o 57600
}

type ancHost struct {
	ctx     context.Context
	name    string
	p       Port
	fw      []byte // firmware decifrato
	last    []byte // ultimo pacchetto inviato, ripetuto nei nuovi tentativi
	sent    int64
	total   int64
	prog    Progress
	logf    func(string)
	started func()
}

func (h *ancHost) send(payload []byte) error {
	b := ancPack(payload)
	h.last = b
	return h.p.Write(b)
}

// reopen: come il programma originale chiude la porta e la riapre a un'altra
// velocità (la chiusura attende che i byte in uscita siano stati trasmessi).
func (h *ancHost) reopen(baud int) error {
	_ = h.p.Close()
	p, err := openPortFunc(h.name, baud)
	if err != nil {
		h.p = nopPort{}
		return fmt.Errorf("riapertura della porta %s a %d baud non riuscita: %v", h.name, baud, err)
	}
	h.p = p
	return nil
}

func (h *ancHost) sleep(d time.Duration) error {
	select {
	case <-h.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// readPacket attende un pacchetto (2 secondi dall'ultimo segnale della radio).
// sawHeader indica che è arrivato almeno l'inizio di un pacchetto (0xAA).
func (h *ancHost) readPacket() (cmd byte, data []byte, sawHeader bool, err error) {
	deadline := time.Now().Add(ancWait)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return 0, nil, false, errTimeout
		}
		b, err := readByte(h.ctx, h.p, left)
		if err != nil {
			return 0, nil, false, err
		}
		if b == 0xAA {
			break
		}
	}
	// come il programma originale: all'arrivo di 0xAA l'attesa riparte da capo
	deadline = time.Now().Add(ancWait)
	bad := func() (byte, []byte, bool, error) {
		// il programma originale resta in attesa fino allo scadere dei 2 secondi
		if err := h.sleep(time.Until(deadline)); err != nil {
			return 0, nil, true, err
		}
		_ = h.p.Purge()
		return 0, nil, true, errAncBad
	}
	hdr := make([]byte, 3)
	if err := h.p.ReadFull(h.ctx, hdr, time.Until(deadline)); err != nil {
		if err == errTimeout {
			return 0, nil, true, errAncBad
		}
		return 0, nil, true, err
	}
	n := int(hdr[1]) | int(hdr[2])<<8
	if n == 0 || n+6 > ancMaxRx {
		return bad()
	}
	rest := make([]byte, n+2)
	if err := h.p.ReadFull(h.ctx, rest, time.Until(deadline)); err != nil {
		if err == errTimeout {
			return 0, nil, true, errAncBad
		}
		return 0, nil, true, err
	}
	frame := append([]byte{0xAA}, hdr...)
	frame = append(frame, rest[:n]...)
	if crc16x(frame) != uint16(rest[n])|uint16(rest[n+1])<<8 {
		return bad()
	}
	return rest[0], rest[:n], true, nil
}

// receive: risposta della radio con i nuovi tentativi del programma originale
// (fino a 5 nuovi invii dell'ultimo pacchetto; il conteggio riparte quando la
// radio inizia a rispondere).
func (h *ancHost) receive() (byte, []byte, error) {
	left := ancRetries
	for {
		cmd, data, saw, err := h.readPacket()
		if err == nil {
			return cmd, data, nil
		}
		if err != errTimeout && err != errAncBad {
			return 0, nil, err
		}
		if saw {
			left = ancRetries
		}
		if left == 0 {
			return 0, nil, errTimeout
		}
		left--
		n := ancRetries - left
		if h.last != nil {
			h.logf(fmt.Sprintf("Nessuna risposta valida: nuovo invio %d di %d", n, ancRetries))
			if err := h.p.Write(h.last); err != nil {
				return 0, nil, err
			}
		} else {
			h.logf(fmt.Sprintf("In attesa della radio… (%d di %d)", n, ancRetries))
		}
	}
}

func (h *ancHost) progress(msg string) {
	done := h.sent
	if h.total > 0 && done > h.total*99/100 {
		done = h.total * 99 / 100
	}
	h.prog(int(done), int(h.total), msg)
}

// ancUpdate invia il firmware (già decifrato) alla radio collegata a port.
// started viene chiamata alla prima risposta della radio: da quel momento
// l'operazione non va interrotta.
func ancUpdate(ctx context.Context, port string, fw []byte, o ancOpts, prog Progress, logf func(string), started func()) (int64, error) {
	baud := o.Baud
	if baud != 57600 {
		baud = 115200
	}
	if logf == nil {
		logf = func(string) {}
	}
	if started == nil {
		started = func() {}
	}
	p, err := openPortFunc(port, ancBaudStart)
	if err != nil {
		return 0, err
	}
	h := &ancHost{ctx: ctx, name: port, p: p, fw: fw, total: int64(len(fw) / 6), prog: prog, logf: logf}
	if h.total <= 0 {
		h.total = 1
	}
	defer func() { _ = h.p.Close() }()

	step := 0
	pkt := []byte{0}
	if o.Recovery {
		step = 2
		logf("In attesa che la radio si presenti in modalità aggiornamento…")
		h.progress("In attesa della radio…")
	} else {
		logf("Richiesta alla radio di entrare in modalità aggiornamento…")
		h.progress("Collegamento con la radio…")
		pkt = []byte{6}
		if err := h.send(pkt); err != nil {
			return 0, err
		}
	}
	first := true
	for {
		cmd, rx, err := h.receive()
		if err != nil {
			return h.sent, err
		}
		if first {
			first = false
			started()
		}
		switch cmd {
		case 1:
			switch step {
			case 0:
				logf("La radio è in modalità aggiornamento")
			case 1:
				logf("Velocità di comunicazione verificata, avvio del caricatore…")
			case 2:
				logf(fmt.Sprintf("Caricatore della radio attivo: invio del firmware a %d baud", baud))
				h.progress("Invio del firmware…")
			default:
				h.progress("Invio del firmware…")
			}
			switch step {
			case 0, 1:
				pkt = []byte{1, 0x10, 0x27, 0, 0} // 10000 baud
			case 2:
				pkt = []byte{1, byte(baud), byte(baud >> 8), byte(baud >> 16), byte(baud >> 24)}
			}
			// passo 3: il programma originale ripete l'ultimo pacchetto costruito
			if err := h.send(pkt); err != nil {
				return h.sent, err
			}
			switch step {
			case 0:
				if err := h.reopen(ancBaudMid); err != nil {
					return h.sent, err
				}
				if err := h.sleep(20 * time.Millisecond); err != nil {
					return h.sent, err
				}
				step = 1
			case 1:
				step = 2
			case 2:
				if err := h.sleep(20 * time.Millisecond); err != nil {
					return h.sent, err
				}
				if err := h.reopen(baud); err != nil {
					return h.sent, err
				}
				if err := h.sleep(10 * time.Millisecond); err != nil {
					return h.sent, err
				}
				step = 3
			}
		case 2:
			if len(rx) < 9 {
				return h.sent, fmt.Errorf("richiesta di dati non valida dalla radio")
			}
			addr, n := le32(rx[1:5]), le32(rx[5:9])
			if addr < 0 || n < 0 || n+9 > 0xFFFF {
				return h.sent, fmt.Errorf("richiesta di dati non valida dalla radio (indirizzo %d, lunghezza %d)", addr, n)
			}
			pkt = make([]byte, 9+n)
			cnt := int64(0)
			if addr < int64(len(fw)) {
				cnt = int64(copy(pkt[9:], fw[addr:]))
			}
			pkt[0] = 2
			copy(pkt[1:5], rx[1:5])
			pkt[5], pkt[6], pkt[7], pkt[8] = byte(cnt), byte(cnt>>8), byte(cnt>>16), byte(cnt>>24)
			if err := h.send(pkt); err != nil {
				return h.sent, err
			}
			h.sent += cnt
			h.progress("Invio del firmware…")
		case 4:
			logf("Dati inviati: la radio sta completando l'aggiornamento…")
			pkt = []byte{3}
			if err := h.send(pkt); err != nil {
				return h.sent, err
			}
		case 3:
			logf("La radio ha confermato: aggiornamento completato")
			prog(int(h.total), int(h.total), "Aggiornamento completato")
			return h.sent, nil
		default:
			// il programma originale qui si fermerebbe senza più rispondere: si attende
			// la risposta successiva (con i nuovi tentativi)
			logf(fmt.Sprintf("Risposta sconosciuta dalla radio (comando %d): ignorata", cmd))
		}
	}
}

// porta fittizia dopo una riapertura non riuscita
type nopPort struct{}

func (nopPort) Write([]byte) error { return errors.New("porta chiusa") }
func (nopPort) ReadFull(context.Context, []byte, time.Duration) error {
	return errors.New("porta chiusa")
}
func (nopPort) Purge() error      { return nil }
func (nopPort) SetBaud(int) error { return nil }
func (nopPort) Close() error      { return nil }
