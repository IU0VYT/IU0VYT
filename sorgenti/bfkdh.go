// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware delle Baofeng con caricatore "KDH Bootloader" V1.0
// e firmware .BF (serie con display TFT «330L»: UV-5RM, UV-5RH e derivate),
// secondo il programma del costruttore "BF_Upgrade_Tool_for_TFT_330L_Series"
// (升级工具_V1.0, classe KDH_Bootloader.BootHelper), ricavato dal suo codice:
// stessi pacchetti, stessa sequenza, stesse attese e stessi nuovi tentativi.
//
// Pacchetto: AA cmd arg lenH lenL dati crcH crcL EF; CRC16 XMODEM (0x1021,
// valore iniziale 0) su cmd, arg, lunghezza e dati. La radio risponde con lo
// stesso comando e arg = 06 (accettato) o con un codice d'errore nel comando
// (E1 codice di accesso errato, E2 errore di controllo → nuovo invio, E3
// indirizzo errato, E4 errore di scrittura della flash, E5 comando errato).
// Sequenza "automatica" (radio accesa): "PROGRAMBFNORMALU" → 06; "UPDATE" → 06;
// dopo 20 ms comando 42 (avvio del caricatore). Sequenza "manuale" (radio
// accesa tenendo premuto il PTT, già nel caricatore): si parte da qui.
//   01 "BOOTLOADER"            presentazione
//   04 [pacchetti dell'area 1] poi i pacchetti 03 (arg = numero, 1024 byte)
//   05 [pacchetti dell'area 2] poi di nuovo i pacchetti 03 (numerati da 0)
//   45                         fine (il programma originale non attende risposta)
// Attesa di ogni risposta: 1 secondo, poi nuovo invio dell'ultimo pacchetto, al
// massimo 5 volte; il conteggio riparte quando arriva l'inizio di una risposta.
// Il file .BF si invia così com'è (cifrato: lo decifra il caricatore). Come
// nell'originale, l'ultimo pacchetto dell'area 1 contiene anche i primi byte
// dell'area 2 (se l'area 1 non è un multiplo di 1024 byte) e l'ultimo pacchetto
// del file è completato con byte FF.

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	kdhBaud     = 115200
	kdhWait     = time.Second
	kdhRetries  = 5
	kdhMaxNak   = 10 // E2 ripetuti: l'originale riprova all'infinito
	kdhPkt      = 1024
	kdhHdr      = 16
	kdhHandshke = "PROGRAMBFNORMALU"
)

var errKdhBad = errors.New("risposta non valida")

func kdhPack(cmd, arg byte, data []byte) []byte {
	n := len(data)
	b := make([]byte, 5+n+3)
	b[0], b[1], b[2], b[3], b[4] = 0xAA, cmd, arg, byte(n>>8), byte(n)
	copy(b[5:], data)
	c := crc16x(b[1 : 5+n])
	b[5+n], b[6+n], b[7+n] = byte(c>>8), byte(c), 0xEF
	return b
}

type kdhOpts struct {
	Manual bool // radio già nel caricatore (accesa tenendo premuto il PTT)
}

type kdhHost struct {
	ctx     context.Context
	p       Port
	last    []byte
	logf    func(string)
	started bool
	onStart func()
}

func (h *kdhHost) write(b []byte) error {
	h.last = b
	return h.p.Write(b)
}

// resp attende una risposta: 1 secondo, che riparte all'arrivo di 0xAA.
func (h *kdhHost) resp() (cmd, arg byte, saw bool, err error) {
	deadline := time.Now().Add(kdhWait)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return 0, 0, false, errTimeout
		}
		b, err := readByte(h.ctx, h.p, left)
		if err != nil {
			return 0, 0, false, err
		}
		if b == 0xAA {
			break
		}
	}
	deadline = time.Now().Add(kdhWait)
	bad := func() (byte, byte, bool, error) {
		// l'originale resta in attesa fino allo scadere del tempo
		select {
		case <-h.ctx.Done():
			return 0, 0, true, errCancelled
		case <-time.After(time.Until(deadline)):
		}
		_ = h.p.Purge()
		return 0, 0, true, errKdhBad
	}
	hdr := make([]byte, 4)
	if err := h.p.ReadFull(h.ctx, hdr, time.Until(deadline)); err != nil {
		if err == errTimeout {
			return 0, 0, true, errKdhBad
		}
		return 0, 0, true, err
	}
	n := int(hdr[2])<<8 | int(hdr[3])
	if n > 120 { // buffer di ricezione dell'originale: 128 byte
		return bad()
	}
	rest := make([]byte, n+3)
	if err := h.p.ReadFull(h.ctx, rest, time.Until(deadline)); err != nil {
		if err == errTimeout {
			return 0, 0, true, errKdhBad
		}
		return 0, 0, true, err
	}
	c := crc16x(append(append([]byte{}, hdr...), rest[:n]...))
	if c != uint16(rest[n])<<8|uint16(rest[n+1]) {
		return bad()
	}
	return hdr[0], hdr[1], true, nil
}

var kdhErrText = map[byte]string{
	0xE1: "la radio ha rifiutato il codice di accesso (firmware o radio non compatibili)",
	0xE3: "la radio ha segnalato un indirizzo errato",
	0xE4: "la radio ha segnalato un errore di scrittura della memoria flash",
	0xE5: "la radio ha segnalato un comando errato",
}

// exchange invia un pacchetto e attende la conferma, con i nuovi tentativi
// dell'originale.
func (h *kdhHost) exchange(pkt []byte) error {
	cmd := pkt[1]
	if err := h.write(pkt); err != nil {
		return err
	}
	left, naks := kdhRetries, 0
	for {
		rc, ra, saw, err := h.resp()
		if err == nil {
			if ra == 6 {
				if rc != cmd {
					return fmt.Errorf("la radio ha confermato il comando %02X invece di %02X", rc, cmd)
				}
				if !h.started {
					h.started = true
					if h.onStart != nil {
						h.onStart()
					}
				}
				return nil
			}
			if t, ok := kdhErrText[rc]; ok {
				return rErr("%s", t)
			}
			if rc == 0xE2 {
				naks++
				if naks > kdhMaxNak {
					return rErr("la radio continua a segnalare errori di controllo dei dati")
				}
				left = kdhRetries - 1
				h.logf(fmt.Sprintf("La radio ha segnalato un errore di controllo: nuovo invio (%d)", naks))
				if err := h.p.Write(h.last); err != nil {
					return err
				}
				continue
			}
			// risposta sconosciuta: come l'originale si attende lo scadere del tempo
			select {
			case <-h.ctx.Done():
				return errCancelled
			case <-time.After(kdhWait):
			}
			err = errKdhBad
		}
		if err != errTimeout && err != errKdhBad {
			return err
		}
		if saw || err == errKdhBad {
			left = kdhRetries
		}
		if left <= 0 {
			return errTimeout
		}
		left--
		h.logf(fmt.Sprintf("Nessuna risposta valida: nuovo invio %d di %d", kdhRetries-left, kdhRetries))
		if err := h.p.Write(h.last); err != nil {
			return err
		}
	}
}

// wait06 attende il byte 06 (gli altri sono scartati) entro il tempo indicato.
func (h *kdhHost) wait06(d time.Duration) error {
	deadline := time.Now().Add(d)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		b, err := readByte(h.ctx, h.p, left)
		if err != nil {
			return err
		}
		if b == 6 {
			return nil
		}
	}
}

// handshake della sequenza automatica (HandShake_0)
func (h *kdhHost) handshake() error {
	hs := []byte(kdhHandshke)
	if err := h.write(hs); err != nil {
		return err
	}
	left := kdhRetries
	for {
		err := h.wait06(kdhWait)
		if err == nil {
			break
		}
		if err != errTimeout {
			return err
		}
		if left <= 0 {
			return rErr("la radio non risponde alla richiesta di aggiornamento")
		}
		left--
		h.logf(fmt.Sprintf("Nessuna risposta: nuovo invio %d di %d", kdhRetries-left, kdhRetries))
		if err := h.p.Write(hs); err != nil {
			return err
		}
	}
	h.logf("La radio ha risposto: richiesta di aggiornamento…")
	if err := h.write([]byte("UPDATE")); err != nil {
		return err
	}
	// l'originale non ripete "UPDATE": attende fino a 6 secondi
	if err := h.wait06(time.Duration(kdhRetries+1) * kdhWait); err != nil {
		if err == errTimeout {
			return rErr("la radio non ha accettato la richiesta di aggiornamento")
		}
		return err
	}
	select {
	case <-h.ctx.Done():
		return errCancelled
	case <-time.After(20 * time.Millisecond):
	}
	return nil
}

func be32(b []byte) int64 {
	return int64(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
}

// kdhUpdate invia il file .BF alla radio collegata a port; restituisce i byte di
// firmware inviati. started viene chiamata alla prima conferma del caricatore:
// da quel momento l'operazione non va interrotta.
func kdhUpdate(ctx context.Context, port string, raw []byte, o kdhOpts, prog Progress, logf func(string), started func()) (int64, error) {
	if logf == nil {
		logf = func(string) {}
	}
	if len(raw) < kdhHdr {
		return 0, errors.New("file del firmware troppo corto")
	}
	s1, s2 := be32(raw[1:5]), be32(raw[5:9])
	t1, t2 := int((s1+kdhPkt-1)/kdhPkt), int((s2+kdhPkt-1)/kdhPkt)
	total := t1 + t2
	if total <= 0 {
		total = 1
	}
	p, err := openPortFunc(port, kdhBaud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	h := &kdhHost{ctx: ctx, p: p, logf: logf, onStart: started}
	prog(0, total, "Collegamento con la radio…")
	if !o.Manual {
		logf("Richiesta di aggiornamento alla radio accesa…")
		if err := h.handshake(); err != nil {
			return 0, err
		}
		logf("Avvio del caricatore della radio…")
		if err := h.exchange(kdhPack(0x42, 0, nil)); err != nil {
			return 0, err
		}
		logf("Caricatore attivo")
	}
	logf("Presentazione al caricatore…")
	if err := h.exchange(kdhPack(0x01, 0, []byte("BOOTLOADER"))); err != nil {
		return 0, err
	}
	logf(fmt.Sprintf("Invio dell'area 1: %d pacchetti da 1024 byte", t1))
	if err := h.exchange(kdhPack(0x04, 0, []byte{byte(t1)})); err != nil {
		return 0, err
	}
	pos, region2, num, done := int64(kdhHdr), false, byte(0), 0
	var sent int64
	for {
		if !region2 && pos >= s1+kdhHdr && pos < s1+kdhHdr+kdhPkt {
			region2 = true
			pos = s1 + kdhHdr
			logf(fmt.Sprintf("Invio dell'area 2: %d pacchetti", t2))
			if err := h.exchange(kdhPack(0x05, 0, []byte{byte(t2)})); err != nil {
				return sent, err
			}
			num = 0
			continue
		}
		n := int64(len(raw)) - pos
		if n <= 0 {
			break
		}
		if n > kdhPkt {
			n = kdhPkt
		}
		blk := make([]byte, kdhPkt)
		for i := range blk {
			blk[i] = 0xFF
		}
		copy(blk, raw[pos:pos+n])
		pos += n
		prog(done, total, "Invio del firmware…")
		if err := h.exchange(kdhPack(0x03, num, blk)); err != nil {
			return sent, err
		}
		num++
		done++
		sent += n
	}
	if err := h.p.Write(kdhPack(0x45, 0, nil)); err != nil {
		return sent, err
	}
	// l'originale chiude subito la porta; qui si attende un attimo la conferma
	if rc, ra, _, err := h.resp(); err == nil && rc == 0x45 && ra == 6 {
		logf("La radio ha confermato la fine dell'aggiornamento")
	} else {
		logf("Comando di fine aggiornamento inviato")
	}
	prog(total, total, "Aggiornamento completato")
	return sent, nil
}
