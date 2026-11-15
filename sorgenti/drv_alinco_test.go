//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.40: Alinco, aggiunte per i confronti con CHIRP (TestChirp*).

import (
	"encoding/binary"
	"strings"
	"time"
)

func init() {
	// con la radio simulata la pausa prima di ogni blocco non serve
	alincoGap = time.Millisecond
	// CHIRP ha l'immagine di prova della sola DR-235T: per le altre della serie
	// è la stessa, con i due byte a 0x64 da cui CHIRP riconosce il modello
	for _, id := range []string{"al-dr135t", "al-dr435t", "al-dr03t", "al-dr06t"} {
		sig := alincoO(genByID[id]).sig
		chirpTestImage[id] = "Alinco_DR235T.img"
		chirpImagePatch[id] = func(b []byte) { copy(b[0x64:], sig) }
	}
	// l'immagine di prova della DJ-G7 ha 205 canali scritti ma segnati tutti come
	// vuoti (segno a zero): qui si rimette il segno che CHIRP scrive per un
	// canale occupato, così il confronto di lettura ha dei canali da confrontare
	g7 := func(b []byte) {
		for i := 0; i < 1000; i++ {
			r := b[0x1200+64*i : 0x1240+64*i]
			if hz := int64(binary.LittleEndian.Uint32(r[1:5])); r[5] < 4 && hz >= 500000 && hz < 1300000000 {
				r[0] = byte(alg7Flag(hz, int64(r[5])))
			}
		}
	}
	chirpTestImage["al-djg7t"] = "Alinco_DJ-G7EG.img"
	chirpImagePatch["al-djg7eg"], chirpImagePatch["al-djg7t"] = g7, g7
}

// DJ-175: CHIRP cambia un bit da 0x130 a ogni canale scritto (vedi drv_alinco.go);
// qui quei byte restano quelli dell'immagine di partenza
func (alincoFam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	if alincoO(m).dj175 {
		copy(ref[0x130:0x13D], base[0x130:0x13D])
	}
}

// DJ-G7: CHIRP lascia gli spazi in fondo al nome; qui si tolgono
func (alg7Fam) chirpRefFix(m *genModel, g *Chan, c *chirpChan) {
	c.Name = strings.TrimRight(c.Name, " ")
}
