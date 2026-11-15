// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// yaesu_clone.py (4.15).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.15: copia («CLONE») dei palmari Yaesu VX e FT digitali. La copia parte dalla
// radio: in lettura la radio manda i primi byte con il modello, il programma
// conferma con 06 e la radio manda tutto il resto; in scrittura il programma
// manda i primi byte, la radio conferma con 06 e il resto parte a pezzi. Con i
// cavi a un filo ogni byte inviato torna indietro (eco). Alla fine la radio non
// conferma: l'esito si vede sul suo schermo.

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type yaesuOpt struct {
	model string   // primi 5 byte della memoria
	size  int      // byte della memoria
	head  int      // byte del primo blocco
	chunk int      // pezzi del secondo blocco in scrittura
	sums  [][3]int // somme di controllo: primo byte, ultimo byte, indirizzo della somma
	send  string   // tasto che avvia l'invio sulla radio
	recv  string   // tasto che mette la radio in ricezione
	// blocks (4.37): lunghezze dei blocchi della copia quando sono più di due
	// (_block_lengths di CHIRP: VX-5 10+16+8097, VX-7 10+8+16193, VX-2
	// 10+8+32577, FT-50 otto blocchi). Tutti i blocchi tranne l'ultimo vanno
	// confermati con 06. Vuoto: due blocchi, head e il resto
	blocks []int
	// byteEcho (4.37, FT-50): in scrittura un byte alla volta, aspettando che la
	// radio lo rimandi indietro (come fa CHIRP con questa radio)
	byteEcho bool
}

// blockLens: lunghezze dei blocchi della copia
func (o *yaesuOpt) blockLens() []int {
	if len(o.blocks) > 0 {
		return o.blocks
	}
	return []int{o.head, o.size - o.head}
}

// yaesuByteDelay (4.37): pause della scrittura un byte alla volta (FT-50): prima
// di ogni byte e prima di leggerne l'eco, come CHIRP
var yaesuByteDelay = [2]time.Duration{10 * time.Millisecond, 15 * time.Millisecond}

// yaesuChunkDelay: pausa tra i pezzi in scrittura (la radio non ha controllo di flusso)
var yaesuChunkDelay = 30 * time.Millisecond

func yaesuOptOf(m *genModel) (*yaesuOpt, bool) {
	if f, ok := m.Fam.(interface{ yaesu(*genModel) *yaesuOpt }); ok {
		return f.yaesu(m), true
	}
	return nil, false
}

func yaesuSums(o *yaesuOpt, img []byte) {
	for _, c := range o.sums {
		var s byte
		for _, b := range img[c[0] : c[1]+1] {
			s += b
		}
		img[c[2]] = s
	}
}

func yaesuSumsOK(o *yaesuOpt, img []byte) bool {
	for _, c := range o.sums {
		var s byte
		for _, b := range img[c[0] : c[1]+1] {
			s += b
		}
		if img[c[2]] != s {
			return false
		}
	}
	return true
}

func yaesuMatch(o *yaesuOpt, img []byte) bool {
	return len(img) == o.size && bytes.HasPrefix(img, []byte(o.model))
}

func yaesuDownload(ctx context.Context, m *genModel, port string, o *yaesuOpt, prog Progress) ([]byte, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, err
	}
	defer s.p.Close()
	prog(0, o.size, "In attesa della radio: avvia l'invio sulla radio…")
	var data []byte
	for try := 0; try < 30 && len(data) < o.head; try++ {
		b := readUpTo(ctx, s.p, o.head-len(data), time.Second)
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(b) == 0 && len(data) > 0 {
			break
		}
		data = append(data, b...)
	}
	if len(data) == 0 {
		return nil, genFail("Collegamento", rErr("la radio non ha inviato nulla: mettila in modalità «CLONE» e premi %s dopo aver avviato la lettura", o.send))
	}
	if len(data) < o.head {
		return nil, rErr("risposta della radio non valida (% X)", data)
	}
	if !bytes.HasPrefix(data, []byte(o.model)) {
		for _, x := range genModels {
			if xo, ok := yaesuOptOf(x); ok && xo.model != "" && bytes.HasPrefix(data, []byte(xo.model)) {
				return nil, rErr("la radio è un altro modello: apri la scheda «%s»", x.Name)
			}
		}
		return nil, rErr("la radio è un altro modello (identificativo % X)", data[:5])
	}
	if err := s.send([]byte{0x06}); err != nil {
		return nil, err
	}
	// 4.37: blocchi intermedi, ognuno confermato con 06. Con il cavo a un filo
	// la conferma appena inviata torna indietro prima dei dati: se il blocco
	// comincia con 06 si guarda se arriva un byte in più (la radio, finché non
	// riceve la conferma, non manda altro)
	lens := o.blockLens()
	for _, n := range lens[1 : len(lens)-1] {
		b := readUpTo(ctx, s.p, n, 2*time.Second)
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		if len(b) == n && b[0] == 0x06 {
			if x := readUpTo(ctx, s.p, 1, 300*time.Millisecond); len(x) == 1 {
				b = append(b[1:], x[0])
			}
			if ctx.Err() != nil {
				return nil, errCancelled
			}
		}
		if len(b) < n {
			return nil, rErr("lettura incompleta: %d byte su %d", len(data)+len(b), o.size)
		}
		data = append(data, b...)
		prog(len(data), o.size, "Lettura…")
		if err := s.send([]byte{0x06}); err != nil {
			return nil, err
		}
	}
	got := len(data)
	// con il cavo a un filo il 06 torna indietro prima dei dati: in quel caso
	// arrivano size+1 byte, e il primo si scarta
	rest := make([]byte, 0, o.size-got+1)
	want := o.size - got
	for len(rest) < want+1 {
		to := 2 * time.Second
		if len(rest) >= want {
			if rest[0] != 0x06 {
				break
			}
			to = 400 * time.Millisecond
		}
		n := want + 1 - len(rest)
		if n > 32 {
			n = 32
		}
		b := readUpTo(ctx, s.p, n, to)
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		rest = append(rest, b...)
		if len(b) < n {
			break
		}
		prog(got+len(rest), o.size, "Lettura…")
	}
	if len(rest) == want+1 && rest[0] == 0x06 {
		rest = rest[1:]
	}
	if len(rest) < want {
		return nil, rErr("lettura incompleta: %d byte su %d", got+len(rest), o.size)
	}
	data = append(data, rest[:want]...)
	if !yaesuSumsOK(o, data) {
		return nil, rErr("somma di controllo errata: ripeti la lettura")
	}
	prog(o.size, o.size, "Lettura completata")
	return data, nil
}

func yaesuUpload(ctx context.Context, m *genModel, port string, o *yaesuOpt, img []byte, prog Progress) error {
	if !yaesuMatch(o, img) {
		return fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	img = append([]byte{}, img...)
	yaesuSums(o, img)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return err
	}
	defer s.p.Close()
	prog(0, o.size, "Collegamento con la radio…")
	noAck := genFail("Collegamento", rErr("la radio non conferma i dati: mettila in modalità «CLONE» e premi %s prima di avviare la scrittura", o.recv))
	// 4.37: tutti i blocchi tranne l'ultimo vanno confermati dalla radio
	lens := o.blockLens()
	pos := 0
	for _, n := range lens[:len(lens)-1] {
		if o.byteEcho {
			if err := yaesuSendEcho(ctx, s, img[pos:pos+n]); err != nil {
				if pos == 0 && err != errCancelled {
					return noAck
				}
				return err
			}
		} else if err := s.send(img[pos : pos+n]); err != nil {
			return err
		}
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return noAck
		}
		if b[0] != 0x06 || (img[pos] == 0x06 && !o.byteEcho) {
			// eco del cavo, poi la conferma. Se il blocco comincia con 06 il byte
			// letto può essere l'eco oppure già la conferma: si guarda se arriva altro
			to := 2 * time.Second
			if b[0] == 0x06 {
				to = 300 * time.Millisecond
			}
			b2 := readUpTo(ctx, s.p, n, to)
			if ctx.Err() != nil {
				return errCancelled
			}
			if b[0] != 0x06 || len(b2) > 0 {
				if len(b2) < n {
					return noAck
				}
				if b2[len(b2)-1] != 0x06 {
					return rErr("risposta inattesa della radio (%02X) durante la scrittura", b2[len(b2)-1])
				}
			}
		}
		pos += n
		prog(pos, o.size, "Scrittura…")
	}
	if o.byteEcho {
		// FT-50: anche l'ultimo blocco un byte alla volta; la radio non lo conferma
		if err := yaesuSendEcho(ctx, s, img[pos:o.size]); err != nil {
			return err
		}
		prog(o.size, o.size, "Scrittura completata")
		return nil
	}
	for a := pos; a < o.size; a += o.chunk {
		e := a + o.chunk
		if e > o.size {
			e = o.size
		}
		if err := s.send(img[a:e]); err != nil {
			return err
		}
		if err := s.sleep(yaesuChunkDelay); err != nil {
			return err
		}
		prog(e, o.size, "Scrittura…")
	}
	// eco del cavo: non serve
	readUpTo(ctx, s.p, o.size, 300*time.Millisecond)
	if ctx.Err() != nil {
		return errCancelled
	}
	prog(o.size, o.size, "Scrittura completata")
	return nil
}

// yaesuSendEcho (4.37, FT-50): invio un byte alla volta; la radio rimanda
// indietro ogni byte ricevuto e il byte successivo parte solo dopo l'eco
func yaesuSendEcho(ctx context.Context, s *rawSess, data []byte) error {
	for _, c := range data {
		if err := s.sleep(yaesuByteDelay[0]); err != nil {
			return err
		}
		if err := s.send([]byte{c}); err != nil {
			return err
		}
		if err := s.sleep(yaesuByteDelay[1]); err != nil {
			return err
		}
		b := readUpTo(ctx, s.p, 1, time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return rErr("la radio non rimanda indietro i dati: la scrittura si è fermata")
		}
	}
	return nil
}
