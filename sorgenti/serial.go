// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Collegamento seriale con le radio (cavo di programmazione USB).
// Le funzioni dei protocolli lavorano su questa interfaccia: su Windows è la
// porta COM vera, nei test è un collegamento simulato.

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Port interface {
	Write(b []byte) error
	// ReadFull legge esattamente len(b) byte entro il tempo indicato.
	ReadFull(ctx context.Context, b []byte, timeout time.Duration) error
	// Purge scarta i byte in arrivo non ancora letti.
	Purge() error
	SetBaud(baud int) error
	Close() error
}

type PortInfo struct {
	Name string `json:"name"` // COM6
	Desc string `json:"desc"` // descrizione del dispositivo, se disponibile
}

var errTimeout = errors.New("la radio non risponde")

// openPortFunc apre una porta seriale 8N1 con DTR e RTS attivi, come i
// programmi originali. Impostata dalla versione Windows; nei test viene
// sostituita con un collegamento simulato.
var openPortFunc = func(name string, baud int) (Port, error) {
	return nil, fmt.Errorf("porte seriali non disponibili su questo sistema")
}

var listPortsFunc = func() []PortInfo { return nil }

// Progress riceve l'avanzamento: byte trasferiti, totale, messaggio.
type Progress func(done, total int, msg string)

func readByte(ctx context.Context, p Port, timeout time.Duration) (byte, error) {
	b := make([]byte, 1)
	if err := p.ReadFull(ctx, b, timeout); err != nil {
		return 0, err
	}
	return b[0], nil
}

func xorBytes(b []byte, k byte) []byte {
	out := make([]byte, len(b))
	for i, c := range b {
		out[i] = c ^ k
	}
	return out
}

// readUpTo legge fino a n byte entro il tempo indicato (come la lettura di
// pyserial con timeout: restituisce quelli arrivati, anche meno di n).
func readUpTo(ctx context.Context, p Port, n int, timeout time.Duration) []byte {
	deadline := time.Now().Add(timeout)
	var out []byte
	for len(out) < n {
		left := time.Until(deadline)
		if left <= 0 {
			break
		}
		b := make([]byte, 1)
		if err := p.ReadFull(ctx, b, left); err != nil {
			break
		}
		out = append(out, b[0])
	}
	return out
}
