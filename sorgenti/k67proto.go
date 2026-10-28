// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura e scrittura del Baofeng K67 Pro via cavo, secondo il protocollo del
// programma originale "5RH PRO" (classe DataProtocol: handShake, readDevInfo,
// readInfo, writeInfo e Timer_Elapsed). Tempi e tentativi sono gli stessi.
//
// Sequenza:
//   1. 16 byte (12 x 00 + 4 x FF), la radio risponde 'A'. Se risponde con un
//      altro byte, la porta passa a 115200 baud.
//   2. "PROGRAM" cifrato con un seme casuale + il seme; risposta 'A' cifrata.
//   3. password (8 byte, FF se vuota) cifrata; 'A' oppure password errata.
//   4. "INFORMATION" cifrato; la radio invia 16 byte: modello (es. "BF-K6"),
//      il byte 8 è la banda di frequenza della radio.
//   5. 'R' o 'W' cifrato; risposta 'A'.
//   lettura: per ogni blocco da 4096 byte (0x0000-0xBFFF) richiesta di 4 byte
//            'R', indirizzo alto, basso, 0; la radio invia 4 + 4096 byte.
//   scrittura: 'W', indirizzo, 0 + 4096 byte; risposta 'A'.
//   fine: "END\0", risposta 'A'.
// Tutto ciò che va e viene è cifrato con XOR del seme (anche i dati letti).

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var errCancelled = errors.New("operazione annullata")

type k67Opts struct {
	Password string
	Baud     int // 19200 o 115200 (il programma originale parte da quanto salvato, 19200 di serie)
	// per la scrittura: banda di frequenza attesa (ultimo byte del file) e
	// "ultima radio letta era un BF-K6" (stessi controlli del programma originale)
	FreqBand byte
	K6       bool
	seed     func() byte // solo per i test
}

type k67Info struct {
	Model    string `json:"model"`
	FreqBand byte   `json:"freqBand"`
	K6       bool   `json:"k6"`
	Baud     int    `json:"baud"`
}

type k67Session struct {
	ctx  context.Context
	p    Port
	o    *k67Opts
	seed byte
	baud int
	info k67Info
}

var (
	k67TInfo  = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF}
	k67Sync   = []byte("PROGRAM")
	k67Header = []byte("INFORMATION")
	k67End    = []byte("END\x00")
)

const (
	errK67Pwd  = "password di programmazione errata"
	errK67Band = "la banda di frequenza della radio è diversa da quella del file: il programma originale in questo caso non scrive. Leggi prima la radio e lavora sul file letto"
)

func (s *k67Session) readEnc(timeout time.Duration) (byte, error) {
	b, err := readByte(s.ctx, s.p, timeout)
	return b ^ s.seed, err
}

// waitA: aspetta un byte che, decifrato, sia 'A' (gli altri vengono scartati)
func (s *k67Session) waitA(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		b, err := s.readEnc(left)
		if err != nil {
			return err
		}
		if b == 'A' {
			return nil
		}
	}
}

// handShake: op 'R' o 'W'
func (s *k67Session) handShake(op byte) error {
	if s.o.seed != nil {
		s.seed = s.o.seed()
	} else {
		s.seed = byte(1 + rand.Intn(254)) // Random.Next(1, 255)
	}
	sync := make([]byte, 8)
	for i := 0; i < 7; i++ {
		sync[i] = s.seed ^ k67Sync[i]
	}
	sync[7] = s.seed
	// passo 1: fino a 26 invii, 200 ms di attesa ciascuno
	got := false
	for try := 0; try < 26 && !got; try++ {
		if err := s.p.Write(k67TInfo); err != nil {
			return err
		}
		b, err := readByte(s.ctx, s.p, 200*time.Millisecond)
		if err == errTimeout {
			continue
		}
		if err != nil {
			return err
		}
		got = true
		if b != 'A' {
			// risposta inattesa: il programma originale passa a 115200 baud
			if err := s.p.SetBaud(115200); err != nil {
				return err
			}
			s.baud = 115200
		}
	}
	if !got {
		return errTimeout
	}
	if err := s.p.Write(sync); err != nil {
		return err
	}
	// passi 2-6: 200 ms rinnovati 3 volte
	const hsWait = 800 * time.Millisecond
	if err := s.waitA(hsWait); err != nil {
		return err
	}
	pwd := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	pb := []byte(s.o.Password)
	if len(pb) > 8 {
		return fmt.Errorf("la password può avere al massimo 8 caratteri")
	}
	copy(pwd, pb)
	if err := s.p.Write(xorBytes(pwd, s.seed)); err != nil {
		return err
	}
	b, err := s.readEnc(hsWait)
	if err != nil {
		return err
	}
	if b != 'A' {
		return errors.New(errK67Pwd)
	}
	if err := s.p.Write(xorBytes(k67Header, s.seed)); err != nil {
		return err
	}
	resp := make([]byte, 16)
	if err := s.p.ReadFull(s.ctx, resp, hsWait); err != nil {
		return err
	}
	dec := xorBytes(resp, s.seed)
	n := 0
	for n < 16 && dec[n] != 0xFF {
		n++
	}
	s.info.Model = string(dec[:n])
	s.info.FreqBand = dec[8]
	s.info.K6 = s.info.Model == "BF-K6"
	if op == 'W' {
		if s.info.K6 != s.o.K6 {
			if s.info.K6 {
				return fmt.Errorf("la radio collegata si identifica come \"BF-K6\", ma l'ultima radio letta no: come il programma originale, " +
					"la scrittura parte solo verso lo stesso modello letto. Leggi prima questa radio (o attiva la copia di sicurezza prima della scrittura) e riprova")
			}
			return fmt.Errorf("la radio collegata (\"%s\") non è il modello dell'ultima radio letta (\"BF-K6\"): leggi prima questa radio e riprova", s.info.Model)
		}
		if dec[8] != s.o.FreqBand {
			return errors.New(errK67Band)
		}
	}
	if err := s.p.Write([]byte{s.seed ^ op}); err != nil {
		return err
	}
	return s.waitA(hsWait)
}

func (s *k67Session) end() error {
	if err := s.p.Write(xorBytes(k67End, s.seed)); err != nil {
		return err
	}
	return s.waitA(time.Second)
}

func (s *k67Session) readBlock(addr int) ([]byte, error) {
	req := []byte{'R', byte(addr >> 8), byte(addr), 0}
	if err := s.p.Write(xorBytes(req, s.seed)); err != nil {
		return nil, err
	}
	buf := make([]byte, 4100)
	// 4 secondi, rinnovati una volta
	if err := s.p.ReadFull(s.ctx, buf, 8*time.Second); err != nil {
		return nil, err
	}
	return buf[4:], nil
}

func k67Open(name string, o *k67Opts) (*k67Session, error) {
	baud := o.Baud
	if baud != 115200 {
		baud = 19200
	}
	p, err := openPortFunc(name, baud)
	if err != nil {
		return nil, err
	}
	return &k67Session{p: p, o: o, baud: baud}, nil
}

func k67Fail(step string, err error) error {
	if err == errCancelled {
		return err
	}
	if err == errTimeout {
		return fmt.Errorf("%s: la radio non risponde. Controlla che sia accesa, che il cavo sia inserito bene e che la porta sia quella giusta", step)
	}
	return fmt.Errorf("%s: %v", step, err)
}

// k67Read legge la memoria della radio (0xC000 byte, già decifrati).
func k67Read(ctx context.Context, port string, o *k67Opts, prog Progress) ([]byte, k67Info, error) {
	s, err := k67Open(port, o)
	if err != nil {
		return nil, k67Info{}, err
	}
	defer s.p.Close()
	s.ctx = ctx
	prog(0, xlcRadioLen, "Collegamento con la radio…")
	if err := s.handShake('R'); err != nil {
		return nil, s.info, k67Fail("Collegamento", err)
	}
	img := make([]byte, xlcRadioLen)
	for addr := 0; addr < xlcRadioLen; addr += 4096 {
		if ctx.Err() != nil {
			return nil, s.info, errCancelled
		}
		prog(addr, xlcRadioLen, "Lettura…")
		blk, err := s.readBlock(addr)
		if err != nil {
			return nil, s.info, k67Fail(fmt.Sprintf("Lettura (blocco 0x%04X)", addr), err)
		}
		copy(img[addr:], blk)
	}
	if err := s.end(); err != nil {
		return nil, s.info, k67Fail("Chiusura", err)
	}
	prog(xlcRadioLen, xlcRadioLen, "Lettura completata")
	s.info.Baud = s.baud
	return xorBytes(img, s.seed), s.info, nil
}

// k67Write scrive l'immagine (primi 0xC000 byte del file). Come il programma
// originale, prima legge i 128 byte di informazioni della radio e li mette al
// posto di quelli del file.
func k67Write(ctx context.Context, port string, img []byte, o *k67Opts, prog Progress) (k67Info, error) {
	if len(img) < xlcRadioLen {
		return k67Info{}, fmt.Errorf("dati da scrivere incompleti")
	}
	s, err := k67Open(port, o)
	if err != nil {
		return k67Info{}, err
	}
	defer s.p.Close()
	s.ctx = ctx
	full := make([]byte, xlcRadioLen)
	copy(full, img)
	prog(0, xlcRadioLen, "Collegamento con la radio…")
	if err := s.handShake('W'); err != nil {
		return s.info, k67Fail("Collegamento", err)
	}
	// readDevInfo
	blk, err := s.readBlock(0)
	if err != nil {
		return s.info, k67Fail("Lettura delle informazioni della radio", err)
	}
	copy(full[0:128], xorBytes(blk[:128], s.seed))
	if err := s.end(); err != nil {
		return s.info, k67Fail("Lettura delle informazioni della radio", err)
	}
	if allSame(full[0:128]) {
		return s.info, fmt.Errorf("le informazioni lette dalla radio non sono valide: scrittura annullata, nessun dato è stato scritto")
	}
	if err := s.handShake('W'); err != nil {
		return s.info, k67Fail("Collegamento (seconda fase)", err)
	}
	for addr := 0; addr < xlcRadioLen; addr += 4096 {
		prog(addr, xlcRadioLen, "Scrittura…")
		pkt := make([]byte, 4100)
		pkt[0], pkt[1], pkt[2] = 'W', byte(addr>>8), byte(addr)
		copy(pkt[4:], full[addr:addr+4096])
		if err := s.p.Write(xorBytes(pkt, s.seed)); err != nil {
			return s.info, k67Fail("Scrittura", err)
		}
		b, err := s.readEnc(8 * time.Second)
		if err != nil {
			return s.info, k67Fail(fmt.Sprintf("Scrittura (blocco 0x%04X)", addr), err)
		}
		if b != 'A' {
			return s.info, fmt.Errorf("Scrittura (blocco 0x%04X): la radio ha rifiutato i dati", addr)
		}
	}
	if err := s.end(); err != nil {
		return s.info, k67Fail("Chiusura (i dati sono stati inviati: rileggi la radio per controllare)", err)
	}
	prog(xlcRadioLen, xlcRadioLen, "Scrittura completata")
	s.info.Baud = s.baud
	return s.info, nil
}

func allSame(b []byte) bool {
	for _, x := range b {
		if x != b[0] {
			return false
		}
	}
	return true
}
