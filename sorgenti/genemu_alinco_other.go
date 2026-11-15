//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.40: radio Alinco simulate per le prove (comandi di testo «AL~F…»). Come il
// cavo a un filo, la radio simulata rimanda indietro ogni comando ricevuto
// (noEcho: cavo che non lo fa). Le risposte sono quelle della radio simulata
// delle prove di CHIRP (tests/unit/test_alinco_clone.py).

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type alincoEmu struct {
	mu     sync.Mutex
	img    []byte
	ident  string
	g7     bool // DJ-G7: blocchi da 64 byte, scritture confermate, comandi chiusi da CR
	bare   bool // DJ-175: risposte senza indirizzo
	noEcho bool // cavo che non rimanda i dati
	deaf   bool // la radio non risponde al proprio nome (spenta): torna solo l'eco
	writes int  // blocchi scritti
	known  bool // la radio ha riconosciuto il proprio nome
	// badAt: blocco (indirizzo) a cui la radio risponde con dati rovinati; -1 nessuno
	badAt int
	// nakAt: blocco (indirizzo) che la DJ-G7 non accetta in scrittura; -1 nessuno
	nakAt int
	// ackW: la radio conferma con «OK» anche i blocchi da 16 byte (CHIRP non lo
	// chiede e non lo legge: non si sa se le radio vere lo fanno)
	ackW bool
}

func (alincoFam) emu(m *genModel, img []byte) genEmu {
	o := alincoO(m)
	mem := make([]byte, o.size)
	if img == nil && o.sig != nil {
		copy(mem[0x64:], o.sig)
	}
	copy(mem, img)
	return &alincoEmu{img: mem, ident: o.ident, bare: o.dj175, badAt: -1, nakAt: -1}
}

func (alg7Fam) emu(m *genModel, img []byte) genEmu {
	mem := make([]byte, alg7Size)
	copy(mem, img)
	return &alincoEmu{img: mem, ident: m.Opt.(*alg7Opt).ident, g7: true, badAt: -1, nakAt: -1}
}

func (e *alincoEmu) image() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]byte{}, e.img...)
}

// command: prossimo comando com'è arrivato (con il ritorno a capo)
func (e *alincoEmu) command(p Port) ([]byte, error) {
	var raw []byte
	for {
		b, err := emuByte(p)
		if err != nil {
			return nil, err
		}
		raw = append(raw, b)
		if b != '\r' {
			continue
		}
		// la DJ-G7 chiude i comandi di lettura e scrittura con il solo CR
		if e.g7 && strings.HasPrefix(string(raw), "AL~F") {
			return raw, nil
		}
		b, err = emuByte(p)
		if err != nil {
			return nil, err
		}
		return append(raw, b), nil
	}
}

func (e *alincoEmu) serve(p Port) error {
	blk, alen := 16, 4
	if e.g7 {
		blk, alen = 64, 5
	}
	for {
		raw, err := e.command(p)
		if err != nil {
			return nil // fine della sessione
		}
		if !e.noEcho {
			p.Write(raw)
		}
		cmd := strings.TrimRight(string(raw), "\r\n")
		switch {
		case cmd == e.ident:
			if e.deaf {
				continue
			}
			e.known = true
			p.Write([]byte("OK\r\n\r\n"))
		case !e.known:
			// prima del nome la radio non risponde a niente
		case cmd == "AL~E":
			p.Write([]byte(strings.Repeat("OK", 10)))
			return nil
		case strings.HasPrefix(cmd, "AL~F") && len(cmd) > 4+alen:
			a64, err := strconv.ParseInt(cmd[4:4+alen], 16, 32)
			a := int(a64)
			if err != nil || a%blk != 0 || a+blk > len(e.img) {
				return fmt.Errorf("indirizzo non valido in %q", cmd)
			}
			switch cmd[4+alen] {
			case 'R':
				if len(cmd) != 5+alen {
					return fmt.Errorf("comando di lettura non valido %q", cmd)
				}
				e.mu.Lock()
				h := strings.ToUpper(hex.EncodeToString(e.img[a : a+blk]))
				e.mu.Unlock()
				if a == e.badAt {
					h = h[:len(h)-3] + "ZZ" + h[len(h)-1:]
				}
				switch {
				case e.bare:
					p.Write([]byte(h + "\r\n"))
				case e.g7:
					p.Write([]byte("\r\n" + h + "\r\n"))
				default:
					p.Write([]byte(fmt.Sprintf("\r\n%04X:%s\r\n", a, h)))
				}
			case 'W':
				d, err := hex.DecodeString(cmd[5+alen:])
				if err != nil || len(d) != blk {
					return fmt.Errorf("blocco di scrittura non valido %q", cmd)
				}
				if e.g7 && a == e.nakAt {
					p.Write([]byte("NG\r\n\r\n"))
					continue
				}
				e.mu.Lock()
				copy(e.img[a:], d)
				e.writes++
				e.mu.Unlock()
				if e.g7 || e.ackW {
					p.Write([]byte("OK\r\n\r\n"))
				}
			default:
				return fmt.Errorf("comando sconosciuto %q", cmd)
			}
		default:
			// un altro nome o un comando che la radio non ha: nessuna risposta
		}
	}
}
