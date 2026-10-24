// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura e scrittura della Bajeton 8300 via cavo, secondo il protocollo del
// programma originale BT-11_CPS (classe KDH.RWDataOperation: HandShake,
// ReadRadioData, WriteRadioData). Porta a 57600 baud 8N1, DTR e RTS attivi.
//
// Sequenza:
//   1. "PROGRAMBT11U"; la radio risponde 0x06.
//   2. 'F'; la radio invia 16 byte di identificazione.
//   3. pacchetto di 25 byte "SEND" + scelta casuale della chiave di cifratura
//      (una di 20 sequenze di 4 caratteri); risposta 0x06.
//   lettura: per ogni blocco da 64 byte 'R', indirizzo alto, basso, 0x40;
//            la radio invia 4 + 64 byte (dati cifrati).
//   scrittura: 'W', indirizzo, 0x40 + 64 byte cifrati; risposta 0x06.
//   fine: 'E'.
// Cifratura: XOR con la chiave a rotazione, saltando i byte 0x00, 0xFF,
// uguali alla chiave o al suo complemento, e le posizioni in cui la chiave
// è uno spazio.

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

var btSymbols = []string{"BHT ", "CO 7", "A ES", " EIY", "M PQ", "XN Y", "RVB ", " HQP", "W RC", "MS N",
	" SAT", "K DH", "ZO R", "C SL", "6RB ", " JCG", "PN V", "J PK", "EK L", "I LZ"}

type btInfo struct {
	ID string `json:"id"` // 16 byte inviati dalla radio dopo 'F' (non usati dal programma originale)
}

type btSession struct {
	ctx  context.Context
	p    Port
	key  []byte
	rnd  func(n int) int // rand.Intn; sostituibile nei test
	info btInfo
}

func btCrypt(data, key []byte) {
	k := 0
	for i := range data {
		s := key[k]
		b := data[i]
		if s != ' ' && b != 0 && b != 0xFF && b != s && b != s^0xFF {
			data[i] = b ^ s
		}
		k = (k + 1) % 4
	}
}

// waitByte: aspetta il byte v scartando gli altri (come il programma originale)
func (s *btSession) waitByte(v byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		b, err := readByte(s.ctx, s.p, left)
		if err != nil {
			return err
		}
		if b == v {
			return nil
		}
	}
}

func (s *btSession) handShake() error {
	// passo 1: fino a 4 invii, 1 secondo di attesa ciascuno
	ok := false
	for try := 0; try < 4 && !ok; try++ {
		if err := s.p.Write([]byte("PROGRAMBT11U")); err != nil {
			return err
		}
		err := s.waitByte(0x06, time.Second)
		if err == errTimeout {
			continue
		}
		if err != nil {
			return err
		}
		ok = true
	}
	if !ok {
		return errTimeout
	}
	// passo 2: 'F' e 16 byte di risposta (rinvio di 'F' a ogni tentativo)
	id := make([]byte, 16)
	ok = false
	for try := 0; try < 4 && !ok; try++ {
		if err := s.p.Write([]byte{'F'}); err != nil {
			return err
		}
		err := s.p.ReadFull(s.ctx, id, time.Second)
		if err == errTimeout {
			continue
		}
		if err != nil {
			return err
		}
		ok = true
	}
	if !ok {
		return errTimeout
	}
	s.info.ID = printable(id)
	// passo 3: scelta della chiave
	pkt := make([]byte, 25)
	copy(pkt, "SEND")
	hi := 1 + s.rnd(2) // random.Next(1, 3)
	lo := s.rnd(5)     // random.Next(0, 5)
	pkt[4] = byte(hi<<4 | lo)
	idx := 2 * lo
	if hi == 2 {
		idx++
	}
	idx++
	for j := 0; j < 19; j++ {
		pkt[5+j] = byte(s.rnd(20))
	}
	s.key = []byte(btSymbols[pkt[4+idx]])
	if err := s.p.Write(pkt); err != nil {
		return err
	}
	return s.waitByte(0x06, 4*time.Second)
}

func printable(b []byte) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if c >= 32 && c < 127 {
			out = append(out, c)
		} else {
			out = append(out, '.')
		}
	}
	return string(out)
}

func btOpenSession(ctx context.Context, port string) (*btSession, error) {
	p, err := openPortFunc(port, 57600)
	if err != nil {
		return nil, err
	}
	return &btSession{ctx: ctx, p: p, rnd: rand.Intn}, nil
}

func btFailMsg(step string, err error) error {
	if err == errCancelled {
		return err
	}
	if err == errTimeout {
		return fmt.Errorf("%s: la radio non risponde. Controlla che sia accesa, che il cavo sia inserito bene e che la porta sia quella giusta", step)
	}
	return fmt.Errorf("%s: %v", step, err)
}

// btReadImage legge la memoria della radio (immagine da 0xC300 byte, 0xFF
// dove non letta).
func btReadImage(ctx context.Context, port string, prog Progress, s *btSession) ([]byte, btInfo, error) {
	if s == nil {
		var err error
		if s, err = btOpenSession(ctx, port); err != nil {
			return nil, btInfo{}, err
		}
	}
	defer s.p.Close()
	blocks := btBlocks()
	prog(0, len(blocks), "Collegamento con la radio…")
	if err := s.handShake(); err != nil {
		return nil, s.info, btFailMsg("Collegamento", err)
	}
	img := make([]byte, btImgSize)
	for i := range img {
		img[i] = 0xFF
	}
	for n, addr := range blocks {
		if ctx.Err() != nil {
			return nil, s.info, errCancelled
		}
		prog(n, len(blocks), "Lettura…")
		req := []byte{'R', byte(addr >> 8), byte(addr), 0x40}
		var resp []byte
		for try := 0; ; try++ {
			if try > 3 {
				return nil, s.info, btFailMsg(fmt.Sprintf("Lettura (indirizzo 0x%04X)", addr), errTimeout)
			}
			if err := s.p.Write(req); err != nil {
				return nil, s.info, btFailMsg("Lettura", err)
			}
			buf := make([]byte, 68)
			err := s.p.ReadFull(s.ctx, buf, time.Second)
			if err == errTimeout {
				_ = s.p.Purge()
				continue
			}
			if err != nil {
				return nil, s.info, btFailMsg("Lettura", err)
			}
			if buf[1] != req[1] || buf[2] != req[2] {
				// risposta per un altro indirizzo: si richiede di nuovo
				_ = s.p.Purge()
				continue
			}
			resp = buf[4:]
			break
		}
		data := append([]byte{}, resp...)
		btCrypt(data, s.key)
		if addr < btImgSize {
			copy(img[addr:], data)
		}
	}
	if err := s.p.Write([]byte{'E'}); err != nil {
		return nil, s.info, btFailMsg("Chiusura", err)
	}
	prog(len(blocks), len(blocks), "Lettura completata")
	return img, s.info, nil
}

// btWriteImage scrive l'immagine nella radio, blocco per blocco, nello stesso
// ordine del programma originale (compreso il blocco finale a 0xFF).
func btWriteImage(ctx context.Context, port string, img []byte, prog Progress, s *btSession) (btInfo, error) {
	if len(img) < btImgSize {
		return btInfo{}, fmt.Errorf("dati da scrivere incompleti")
	}
	if s == nil {
		var err error
		if s, err = btOpenSession(ctx, port); err != nil {
			return btInfo{}, err
		}
	}
	defer s.p.Close()
	blocks := btBlocks()
	prog(0, len(blocks), "Collegamento con la radio…")
	if err := s.handShake(); err != nil {
		return s.info, btFailMsg("Collegamento", err)
	}
	for n, addr := range blocks {
		prog(n, len(blocks), "Scrittura…")
		pkt := make([]byte, 68)
		pkt[0], pkt[1], pkt[2], pkt[3] = 'W', byte(addr>>8), byte(addr), 0x40
		for i := 4; i < 68; i++ {
			pkt[i] = 0xFF
		}
		if addr < btImgSize {
			copy(pkt[4:], img[addr:addr+64])
		}
		btCrypt(pkt[4:], s.key)
		if err := s.p.Write(pkt); err != nil {
			return s.info, btFailMsg("Scrittura", err)
		}
		// il programma originale non ripete l'invio: aspetta la conferma fino a 4 secondi
		if err := s.waitByte(0x06, 4*time.Second); err != nil {
			return s.info, btFailMsg(fmt.Sprintf("Scrittura (indirizzo 0x%04X)", addr), err)
		}
	}
	if err := s.p.Write([]byte{'E'}); err != nil {
		return s.info, btFailMsg("Chiusura", err)
	}
	prog(len(blocks), len(blocks), "Scrittura completata")
	return s.info, nil
}

// btIdentify: solo il collegamento iniziale (come all'inizio di una lettura) e
// la chiusura con 'E'. Se la radio risponde al protocollo del programma
// BT-11_CPS è una Bajeton 8300 (o una radio con lo stesso protocollo); i 16
// byte di identificazione vengono mostrati all'utente.
func btIdentify(ctx context.Context, port string, prog Progress) (btInfo, error) {
	s, err := btOpenSession(ctx, port)
	if err != nil {
		return btInfo{}, err
	}
	defer s.p.Close()
	prog(0, 1, "Collegamento con la radio…")
	if err := s.handShake(); err != nil {
		return s.info, btFailMsg("Identificazione", err)
	}
	if err := s.p.Write([]byte{'E'}); err != nil {
		return s.info, btFailMsg("Chiusura", err)
	}
	prog(1, 1, "Radio identificata")
	return s.info, nil
}
