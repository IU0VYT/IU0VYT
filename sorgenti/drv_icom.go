// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// icf.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: copia («clone») delle radio Icom con il cavo di programmazione (OPC-478
// e simili, 9600 baud). Pacchetti «FE FE mittente destinatario comando dati FD»
// (PC = EE, radio = EF); il cavo a un filo rimanda indietro i pacchetti inviati.
//   E0 00 00 00 00      → E1 modello (4 byte) e altri dati
//   E2 modello          → la radio manda i blocchi E4 e poi E5
//   E3 modello, blocchi E4, E5 con la firma del modello → E6 00 (accettato)
// Blocco E4: indirizzo (2 byte, oppure 4 se la memoria supera 64 KiB), lunghezza,
// dati e somma di controllo (complemento a due), scritti come cifre esadecimali.

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type icomOpt struct {
	model  [4]byte
	end    string   // firma del modello nel pacchetto di chiusura
	size   int      // byte della memoria
	ranges [][3]int // zone da scrivere: inizio, fine, lunghezza dei blocchi
	twice  bool     // la radio vuole l'identificazione due volte prima della scrittura
	fast   bool     // copia veloce: dopo l'identificazione si passa a 38400 baud
}

const (
	icomPC, icomRadio                                = 0xEE, 0xEF
	icomID, icomModel, icomOut, icomIn, icomDat      = 0xE0, 0xE1, 0xE2, 0xE3, 0xE4
	icomEnd, icomOK                             byte = 0xE5, 0xE6
)

const icomFast byte = 0xE8 // avvio della copia veloce

type icomSess struct {
	*rawSess
	buf  []byte
	echo bool // il cavo rimanda indietro i pacchetti inviati
}

func icomFrame(src, dst, cmd byte, payload []byte) []byte {
	out := append([]byte{0xFE, 0xFE, src, dst, cmd}, payload...)
	return append(out, 0xFD)
}

func icomSum(b []byte) byte {
	var s byte
	for _, x := range b {
		s += x
	}
	return -s
}

func (s *icomSess) sendCmd(cmd byte, payload []byte) error {
	return s.sendFrame(icomFrame(icomPC, icomRadio, cmd, payload))
}

// sendFrame: con il cavo a un filo il pacchetto torna indietro: lo si aspetta,
// così la radio riceve i dati al suo ritmo e un cavo difettoso si nota subito
func (s *icomSess) sendFrame(frame []byte) error {
	if err := s.send(frame); err != nil {
		return err
	}
	if !s.echo {
		return nil
	}
	for len(s.buf) < len(frame) {
		b := readUpTo(s.ctx, s.p, len(frame)-len(s.buf), 2*time.Second)
		if s.ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			break
		}
		s.buf = append(s.buf, b...)
	}
	if !bytes.HasPrefix(s.buf, frame) {
		return rErr("il cavo non rimanda indietro i dati inviati come dovrebbe: cavo non adatto o disturbi")
	}
	s.buf = s.buf[len(frame):]
	return nil
}

// frame: prossimo pacchetto della radio (quelli del PC tornati indietro si saltano);
// nil alla scadenza
func (s *icomSess) frame(to time.Duration) (cmd byte, payload []byte, ok bool, err error) {
	deadline := time.Now().Add(to)
	for {
		if k := bytes.IndexByte(s.buf, 0xFD); k >= 0 {
			f := s.buf[:k]
			s.buf = s.buf[k+1:]
			for len(f) > 2 && f[2] == 0xFE {
				f = f[1:]
			}
			if len(f) < 5 || f[0] != 0xFE || f[1] != 0xFE {
				return 0, nil, false, rErr("dati della radio non riconosciuti: controlla il cavo")
			}
			if f[2] == icomPC && f[3] == icomRadio {
				s.echo = true
				continue
			}
			return f[4], f[5:], true, nil
		}
		left := time.Until(deadline)
		if left <= 0 {
			return 0, nil, false, nil
		}
		if left > 250*time.Millisecond {
			left = 250 * time.Millisecond
		}
		b := readUpTo(s.ctx, s.p, 1, left)
		if s.ctx.Err() != nil {
			return 0, nil, false, errCancelled
		}
		s.buf = append(s.buf, b...)
		if len(s.buf) > 4096 {
			return 0, nil, false, rErr("dati della radio non riconosciuti: controlla il cavo")
		}
	}
}

// ident: la radio dichiara il modello
func (s *icomSess) ident(o *icomOpt) error {
	if err := s.sendCmd(icomID, []byte{0, 0, 0, 0}); err != nil {
		return err
	}
	cmd, p, ok, err := s.frame(2 * time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return rErr("la radio non risponde")
	}
	if cmd != icomModel || len(p) < 4 {
		return rErr("risposta inattesa della radio")
	}
	if !bytes.Equal(p[:4], o.model[:]) {
		for _, x := range genModels {
			if xo, isIcom := icomOptOf(x); isIcom && bytes.Equal(p[:4], xo.model[:]) {
				return rErr("la radio è un altro modello: apri la scheda «%s»", x.Name)
			}
		}
		return rErr("la radio è un altro modello (identificativo % X)", p[:4])
	}
	return nil
}

// start: avvio della copia (cmd = E2 lettura, E3 scrittura). Con la copia veloce
// la radio riceve prima il pacchetto E8 a 9600 baud, poi tutto continua a 38400.
func (s *icomSess) start(o *icomOpt, cmd byte) error {
	if !o.fast {
		return s.sendCmd(cmd, o.model[:])
	}
	pre := bytes.Repeat([]byte{0xFE}, 20)
	hs := icomFrame(icomPC, icomRadio, icomFast, append(append([]byte{}, o.model[:]...), 0x00, 0x00, 0x02, 0x01, 0xFD))
	if err := s.send(append(pre, hs...)); err != nil {
		return err
	}
	// eco ed eventuale risposta a 9600 baud: non servono
	readUpTo(s.ctx, s.p, 128, 400*time.Millisecond)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	s.buf = nil
	if err := s.p.SetBaud(38400); err != nil {
		return err
	}
	pre = bytes.Repeat([]byte{0xFE}, 14)
	fr := icomFrame(icomPC, icomRadio, cmd, []byte{o.model[0], o.model[1], o.model[2], 0x00})
	return s.sendFrame(append(pre, fr...))
}

func icomDownload(ctx context.Context, port string, baud int, o *icomOpt, prog Progress) ([]byte, error) {
	rs, err := rawOpen(ctx, port, baud)
	if err != nil {
		return nil, err
	}
	defer rs.p.Close()
	s := &icomSess{rawSess: rs}
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.ident(o); err != nil {
		return nil, genFail("Identificazione", err)
	}
	if err := s.start(o, icomOut); err != nil {
		return nil, err
	}
	img := make([]byte, o.size)
	alen := 2
	if o.size >= 0x10000 {
		alen = 4
	}
	addr := 0
	for {
		cmd, p, ok, err := s.frame(10 * time.Second)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, rErr("la radio ha smesso di inviare dati prima della fine (0x%04X)", addr)
		}
		if cmd == icomEnd {
			break
		}
		if cmd != icomDat {
			continue
		}
		d := make([]byte, len(p)/2)
		if _, err := hex.Decode(d, p[:len(d)*2]); err != nil || len(d) < alen+2 {
			return nil, rErr("blocco della radio non valido (0x%04X)", addr)
		}
		a := 0
		for _, b := range d[:alen] {
			a = a<<8 | int(b)
		}
		n := int(d[alen])
		if len(d) != alen+1+n+1 || icomSum(d[:len(d)-1]) != d[len(d)-1] {
			return nil, rErr("somma di controllo errata nel blocco 0x%04X: ripeti la lettura", a)
		}
		if a+n > len(img) {
			return nil, rErr("la radio ha inviato più dati del previsto (0x%04X)", a)
		}
		copy(img[a:], d[alen+1:alen+1+n])
		addr = a + n
		prog(addr, o.size, "Lettura…")
	}
	if addr != o.size {
		return nil, rErr("lettura incompleta: %d byte su %d", addr, o.size)
	}
	prog(o.size, o.size, "Lettura completata")
	return img, nil
}

func icomUpload(ctx context.Context, port string, baud int, o *icomOpt, img []byte, prog Progress) error {
	if len(img) != o.size {
		return fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	rs, err := rawOpen(ctx, port, baud)
	if err != nil {
		return err
	}
	defer rs.p.Close()
	s := &icomSess{rawSess: rs}
	prog(0, o.size, "Collegamento con la radio…")
	if err := s.ident(o); err != nil {
		return genFail("Identificazione", err)
	}
	if o.twice {
		if err := s.ident(o); err != nil {
			return genFail("Identificazione", err)
		}
	}
	if err := s.start(o, icomIn); err != nil {
		return err
	}
	alen := 2
	if o.size >= 0x10000 {
		alen = 4
	}
	for _, rg := range o.ranges {
		for a := rg[0]; a < rg[1]; a += rg[2] {
			n := rg[2]
			if a+n > rg[1] {
				n = rg[1] - a
			}
			var d []byte
			if alen == 4 {
				d = []byte{byte(a >> 24), byte(a >> 16), byte(a >> 8), byte(a), byte(n)}
			} else {
				d = []byte{byte(a >> 8), byte(a), byte(n)}
			}
			d = append(d, img[a:a+n]...)
			d = append(d, icomSum(d))
			frame := icomFrame(icomPC, icomRadio, icomDat, []byte(strings.ToUpper(hex.EncodeToString(d))))
			if err := s.sendFrame(frame); err != nil {
				return err
			}
			if !s.echo {
				if err := s.sleep(20 * time.Millisecond); err != nil {
					return err
				}
			}
			prog(a+n, o.size, "Scrittura…")
		}
	}
	if err := s.sendCmd(icomEnd, []byte(o.end)); err != nil {
		return err
	}
	for {
		cmd, p, ok, err := s.frame(5 * time.Second)
		if err != nil {
			return err
		}
		if !ok {
			return rErr("la radio non ha confermato la scrittura")
		}
		if cmd != icomOK {
			continue
		}
		if len(p) == 0 || p[0] != 0 {
			return rErr("la radio ha rifiutato i dati scritti")
		}
		break
	}
	prog(o.size, o.size, "Scrittura completata")
	return nil
}

// famiglie Icom: i parametri della copia
func icomOptOf(m *genModel) (*icomOpt, bool) {
	if f, ok := m.Fam.(interface{ icom(*genModel) *icomOpt }); ok {
		return f.icom(m), true
	}
	return nil, false
}

const icomCableNote = "Cavo Icom OPC-478 (o compatibile) nella presa dell'altoparlante, 9600 baud; la radio va accesa normalmente " +
	"e durante la copia mostra «CLONE». "
