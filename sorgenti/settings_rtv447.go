// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.47: impostazioni dei 24 modelli Retevis ricavati dai programmi del costruttore
// (famiglia «Analog1»: drv_h777v4.go, drv_rb66.go, drv_rb65.go e i modelli «vend»
// di drv_small.go). Ogni voce viene dal programma del suo modello: la posizione e
// il valore di ogni scelta sono stati letti due volte, nel metodo che riempie la
// finestra dalla memoria e nel gestore che scrive la memoria quando si cambia la
// voce, e si offrono solo le voci in cui le due letture coincidono e che il
// programma mostra (quelle nascoste nella finestra di un modello non si offrono).
// I nomi sono quelli dei file di lingua dei programmi, tradotti. Controprova: i
// valori dei file dati di fabbrica dei 24 programmi corrispondono tutti a una
// scelta ammessa. La scheda con il metodo è in ANALISI_Retevis_Analog1.md.
//
// Una voce: posizione nella memoria, maschera dei bit, valore (già nella maschera)
// di ogni scelta. Senza scelte è una casella (bit a 1 = attiva).
// Non si offrono: le voci con posizioni prese da variabili che non si sono potute
// seguire fino in fondo (RB47), la guida vocale dei modelli in cui il programma
// ha un elenco di una sola voce.
// La password di programmazione è una voce a parte, di testo (rtvPwDefs, più sotto).

import (
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
)

type rtvSet struct {
	key, label string
	off        int
	mask       byte
	vals       []byte
	opts       []string
}

// rtvHalf: 0,5 s, 1 s, 1,5 s… (ritardo del VOX)
func rtvHalf(n int) []string {
	o := make([]string, n)
	for i := range o {
		o[i] = strings.Replace(strconv.FormatFloat(0.5*float64(i+1), 'f', -1, 64), ".", ",", 1) + " s"
	}
	return o
}

var rtvSetDefs = map[string][]rtvSet{
	"rt-b3b": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, ijvNums(0, 8, 1, "")},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voice", "Guida vocale", 0x0D0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"voxsw", "VOX", 0x0D0, 0x10, nil, nil},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, append([]string{"Spento"}, ijvNums(1, 7, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-b63b": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, ijvNums(0, 8, 1, "")},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voice", "Guida vocale", 0x0D0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"voxsw", "VOX", 0x0D0, 0x10, nil, nil},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, append([]string{"Spento"}, ijvNums(1, 7, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-rt22p": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxsw", "VOX", 0x0D0, 0x10, nil, nil},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key", "Tasto laterale", 0x0D8, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Scansione", "Allarme di emergenza"}},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-rt622p": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxsw", "VOX", 0x0D0, 0x10, nil, nil},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key", "Tasto laterale", 0x0D8, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Scansione", "Allarme di emergenza"}},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-rb666": {
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"voice", "Guida vocale", 0x0D0, 0x0C, []byte{0x00, 0x08}, []string{"Spento", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0D9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0DA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
	},
	"rt-rb666p": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0D9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0DA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
	},
	"rt-rb45": {
		{"squelch", "Livello dello squelch", 0x0F3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0F2, 0x02, nil, nil},
		{"scanmode", "Modo di scansione", 0x0F2, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"beep", "Bip dei tasti", 0x0F2, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0F4, 0x0F, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxdelay", "Ritardo del VOX", 0x0F4, 0xF0, []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}, rtvHalf(6)},
		{"voxlevel", "Livello del VOX", 0x0F5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0F6, 0xF0, []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}, []string{"Spento", "Annuncio del canale", "Monitor", "Scansione", "VOX", "Allarme di emergenza"}},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0F6, 0x0F, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Annuncio del canale", "Monitor", "Scansione", "VOX", "Allarme di emergenza"}},
	},
	"rt-rb645": {
		{"squelch", "Livello dello squelch", 0x0F3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x0F2, 0x02, nil, nil},
		{"scanmode", "Modo di scansione", 0x0F2, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"beep", "Bip dei tasti", 0x0F2, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0F4, 0x0F, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxdelay", "Ritardo del VOX", 0x0F4, 0xF0, []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}, rtvHalf(6)},
		{"voxlevel", "Livello del VOX", 0x0F5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0F6, 0xF0, []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}, []string{"Spento", "Annuncio del canale", "Monitor", "Scansione", "VOX", "Allarme di emergenza"}},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0F6, 0x0F, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Annuncio del canale", "Monitor", "Scansione", "VOX", "Allarme di emergenza"}},
	},
	"rt-rb47": {
		{"save", "Risparmio della batteria", 0x0F2, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x0F3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x0F2, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"voxsw", "VOX", 0x0F2, 0x10, []byte{0x00, 0x10}, []string{"Spento", "Acceso"}},
		{"codesw", "Controllo del codice (Code switch)", 0x0F2, 0x80, []byte{0x00, 0x80}, []string{"Spento", "Acceso"}},
		{"voxlevel", "Livello del VOX", 0x0F5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
	},
	"rt-b3h": {
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxsw", "VOX", 0x3C0, 0x10, []byte{0x00, 0x10}, []string{"Spento", "Acceso"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(1, 9, 1, "")},
		{"key", "Tasto laterale", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Scansione", "Allarme di emergenza"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-b63h": {
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(1, 9, 1, "")},
		{"voxsw", "VOX", 0x3C0, 0x10, []byte{0x00, 0x10}, []string{"Spento", "Acceso"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key", "Tasto laterale", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Scansione", "Allarme di emergenza"}},
	},
	"rt-rb647": {
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"specmode", "Tipo di codice speciale (Spec mode)", 0x3C2, 0xFF, []byte{0x00, 0x01}, []string{"Spec Code 1", "Spec Code 2"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"keydown", "Tasto giù (pressione lunga)", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Scansione", "Allarme di emergenza"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"voxsw", "VOX", 0x3C0, 0x10, []byte{0x00, 0x10}, []string{"Spento", "Acceso"}},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x80, []byte{0x00, 0x80}, []string{"Spento", "Acceso"}},
	},
	"rt-rt68h": {
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C9, 0x01, nil, nil},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x02, 0x03, 0x04}, []string{"Spento", "Scansione", "VOX", "Blocco del canale"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C8, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "Monitor finché tieni premuto", "Scansione", "VOX", "Blocco del canale"}},
	},
	"rt-rt668h": {
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x08}, []string{"Nessuna", "Inglese"}},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C9, 0x01, nil, nil},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"specmode", "Tipo di codice speciale (Spec mode)", 0x3C5, 0xFF, []byte{0x00, 0x01}, []string{"Spec Code", "Cifratura"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x20, []byte{0x00, 0x20}, []string{"Spento", "Acceso"}},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x02, 0x03, 0x04}, []string{"Spento", "Scansione", "VOX", "Blocco del canale"}},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C8, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "Monitor finché tieni premuto", "Scansione", "VOX", "Blocco del canale"}},
	},
	"rt-nr10": {
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x04, 0x08}, []string{"Nessuna", "Cinese", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "VOX", "Scansione", "Monitor", "Riduzione del rumore"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "VOX", "Scansione", "Monitor", "Riduzione del rumore"}},
	},
	"rt-nr610": {
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x04, 0x08}, []string{"Nessuna", "Cinese", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "VOX", "Scansione", "Monitor", "Riduzione del rumore"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, []string{"Spento", "VOX", "Scansione", "Monitor", "Riduzione del rumore"}},
	},
	"rt-rb66s": {
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x04, 0x08}, []string{"Nessuna", "Cinese", "Inglese"}},
		{"beep", "Bip dei tasti", 0x3C0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C0, 0x20, nil, nil},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"keylock", "Blocco dei tasti", 0x3C0, 0x10, nil, nil},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x80, []byte{0x00, 0x80}, []string{"Spento", "Acceso"}},
		{"key2s", "Tasto laterale 2 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"abr", "Retroilluminazione", 0x3CA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(1, 10, 1, "")...)},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x3C8, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"lcd", "LCD", 0x3CB, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, ijvNums(1, 5, 1, "")},
	},
	"rt-rb666s": {
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x3C0, 0x0C, []byte{0x00, 0x04, 0x08}, []string{"Nessuna", "Cinese", "Inglese"}},
		{"beep", "Bip dei tasti", 0x3C0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C0, 0x20, nil, nil},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C5, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"keylock", "Blocco dei tasti", 0x3C0, 0x10, nil, nil},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x3C9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x80, []byte{0x00, 0x80}, []string{"Spento", "Acceso"}},
		{"key2s", "Tasto laterale 2 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"abr", "Retroilluminazione", 0x3CA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(1, 10, 1, "")...)},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x3C8, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, []string{"Spento", "Monitor", "Scansione", "VOX", "Allarme", "Torcia"}},
		{"lcd", "LCD", 0x3CB, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, ijvNums(1, 5, 1, "")},
	},
	"rt-rb65": {
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"beep", "Bip dei tasti", 0x3C0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
	},
	"rt-rb66": {
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"voice", "Guida vocale", 0x0D0, 0x0C, []byte{0x00, 0x04}, []string{"Nessuna", "Inglese"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0D9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor finché tieni premuto", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0DA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor finché tieni premuto", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
	},
	"rt-rb66p": {
		{"squelch", "Livello dello squelch", 0x0D1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"voice", "Guida vocale", 0x0D0, 0x0C, []byte{0x00, 0x04}, []string{"Spento", "Inglese"}},
		{"save", "Risparmio della batteria", 0x0D0, 0x02, nil, nil},
		{"beep", "Bip dei tasti", 0x0D0, 0x01, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x0D2, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"key1l", "Tasto laterale 1 (pressione lunga)", 0x0D9, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"key2l", "Tasto laterale 2 (pressione lunga)", 0x0DA, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, []string{"Spento", "Monitor", "Scansione", "VOX", "Blocco del canale", "Torcia", "Allarme di emergenza"}},
		{"voxlevel", "Livello del VOX", 0x0D3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"voxdelay", "Ritardo del VOX", 0x0D4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"scanmode", "Modo di scansione", 0x0D0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
	},
	"rt-rt68b": {
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"voice", "Guida vocale", 0x3C0, 0x04, []byte{0x00, 0x04}, []string{"Nessuna", "Inglese"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C0, 0x10, nil, nil},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"beep", "Bip dei tasti", 0x3C0, 0x01, nil, nil},
		{"voxsw", "VOX", 0x3C0, 0x20, nil, nil},
		{"chlock", "Blocco del canale", 0x3C0, 0x08, nil, nil},
	},
	"rt-rt668b": {
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, append([]string{"Spento"}, ijvNums(30, 10, 30, " s")...)},
		{"voice", "Guida vocale", 0x3C0, 0x04, []byte{0x00, 0x04}, []string{"Nessuna", "Inglese"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, append([]string{"Spento"}, ijvNums(1, 7, 1, "")...)},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"roger", "Tono di fine trasmissione (roger beep)", 0x3C0, 0x10, nil, nil},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"beep", "Bip dei tasti", 0x3C0, 0x01, nil, nil},
		{"voxsw", "VOX", 0x3C0, 0x20, nil, nil},
		{"chlock", "Blocco del canale", 0x3C0, 0x08, nil, nil},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x80, []byte{0x00, 0x80}, []string{"Spento", "Acceso"}},
	},
	"rt-rb669": {
		{"squelch", "Livello dello squelch", 0x3C1, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, ijvNums(0, 10, 1, "")},
		{"scanmode", "Modo di scansione", 0x3C0, 0x40, []byte{0x00, 0x40}, []string{"Portante (CO)", "A tempo"}},
		{"save", "Risparmio della batteria", 0x3C0, 0x02, nil, nil},
		{"tot", "Tempo massimo di trasmissione (TOT)", 0x3C3, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, append([]string{"Spento"}, ijvNums(30, 6, 30, " s")...)},
		{"specmode", "Tipo di codice speciale (Spec mode)", 0x3C8, 0xFF, []byte{0x00, 0x01}, []string{"Spec Code 1", "Spec Code 2"}},
		{"voxlevel", "Livello del VOX", 0x3C4, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, append([]string{"Spento"}, ijvNums(1, 9, 1, "")...)},
		{"codesw", "Controllo del codice (Code switch)", 0x3C0, 0x20, []byte{0x00, 0x20}, []string{"Spento", "Acceso"}},
		{"voxdelay", "Ritardo del VOX", 0x3C7, 0xFF, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, rtvHalf(6)},
		{"key1s", "Tasto laterale 1 (pressione breve)", 0x3C2, 0xFF, []byte{0x00, 0x01, 0x02}, []string{"Spento", "Monitor", "VOX"}},
	},
}

// 4.47, password di programmazione (decisione di Daniele del 23/11/2026: come CHIRP
// con la Retevis RT18, che la mostra fra le impostazioni): non ferma la lettura né
// la scrittura, si vede in chiaro e si può cambiare o togliere. Vincoli presi dai
// programmi del costruttore, letti due volte (il metodo che la scrive nella memoria
// prima della scrittura, e le due tabelle dei caratteri — da carattere a byte e da
// byte a carattere — che in ogni programma sono una l'inversa dell'altra):
//
//	off, n  posizione e numero di byte nella memoria che il programma scrive (il
//	        programma della RB66 ne tiene 3: gli altri 3 che la radio dà al comando 05
//	        stanno oltre la memoria che legge e scrive; RB66P 4; gruppo da 11 byte 3)
//	tab     tabella dei caratteri: 'A' il byte è la posizione in 0-9 A-Z a-z (0-61);
//	        'B' cifre 11-20, poi da 21 a 72 le lettere all'indietro (z…a, Z…A);
//	        'C' tabella mescolata di RB66S e RB666S (rtvPwTabC)
//	digits  la casella della nuova password del programma accetta solo cifre
//	hidden  il programma del modello tiene nascoste le caselle della password (con
//	        quel programma una radio con la password non si legge e non si scrive)
//
// In lettura un byte fuori dalla tabella vale «vuoto» e si salta (enc.Fmt «skip»);
// in scrittura i byte dopo l'ultimo carattere vanno a FF, come nel programma.
// La H777 nuova versione ha la tabella di CHIRP, che la password non la mostra: la
// voce le viene aggiunta in coda (rtvPwH777).
type rtvPw struct {
	off, n         int
	tab            byte
	digits, hidden bool
}

var rtvPwDefs = map[string]rtvPw{
	"rt-h777v4": {0xDD, 6, 'A', false, false},
	"rt-b3b":    {0xDD, 6, 'A', false, false},
	"rt-b63b":   {0xDD, 6, 'A', false, false},
	"rt-rt22p":  {0xDD, 6, 'A', false, false},
	"rt-rt622p": {0xDD, 6, 'A', false, false},
	"rt-rb666":  {0xDD, 6, 'A', false, true},
	"rt-rb666p": {0xDD, 6, 'A', false, false},
	"rt-rb45":   {0xFA, 3, 'A', false, true},
	"rt-rb645":  {0xFA, 3, 'A', false, true},
	"rt-rb47":   {0xFA, 3, 'A', false, false},
	"rt-rb66p":  {0xDD, 4, 'A', false, false},
	"rt-rb66":   {0xDD, 3, 'B', true, true},
	"rt-rb65":   {0x3B0, 6, 'B', true, false},
	"rt-b3h":    {0x3B0, 6, 'B', false, true},
	"rt-b63h":   {0x3B0, 6, 'B', false, true},
	"rt-rb647":  {0x3B0, 6, 'B', false, false},
	"rt-rt68h":  {0x3B0, 6, 'B', false, false},
	"rt-rt668h": {0x3B0, 6, 'B', false, false},
	"rt-nr10":   {0x3B0, 6, 'B', false, false},
	"rt-nr610":  {0x3B0, 6, 'B', false, false},
	"rt-rt68b":  {0x3B0, 6, 'B', false, false},
	"rt-rt668b": {0x3B0, 6, 'B', false, false},
	"rt-rb669":  {0x3B0, 6, 'B', false, false},
	"rt-rb66s":  {0x3B0, 6, 'C', false, false},
	"rt-rb666s": {0x3B0, 6, 'C', false, false},
}

const rtvPwChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// rtvPwTabC: byte dei 62 caratteri, nell'ordine di rtvPwChars (RB66S, RB666S)
var rtvPwTabC = [62]byte{48, 15, 20, 67, 28, 69, 6, 60, 122, 111,
	58, 33, 12, 13, 14, 1, 16, 17, 24, 19, 2, 21, 22, 23, 18, 25, 26, 27, 4, 29, 30, 31, 32, 11, 34, 35,
	36, 37, 128, 39, 40, 141, 42, 43, 44, 99, 46, 47, 0, 49, 50, 51, 52, 153, 54, 55, 156, 57, 10, 59, 7, 88}

// rtvPwByte: byte del carattere numero i di rtvPwChars nella tabella tab
func rtvPwByte(tab byte, i int) int {
	switch tab {
	case 'B':
		if i < 10 {
			return 11 + i
		}
		return 82 - i // «A» (10) = 72 … «z» (61) = 21
	case 'C':
		return int(rtvPwTabC[i])
	}
	return i
}

const (
	rtvPwLabel = "Password di programmazione"
	rtvPwDesc  = "Password che il programma del costruttore chiede prima di leggere o scrivere la radio; vuota = nessuna password. " +
		"Questo programma legge e scrive la radio anche quando c'è (come CHIRP): qui la vedi, la cambi o la togli, e la modifica arriva alla radio quando la scrivi. " +
		"Lunghezza e caratteri sono quelli del programma del costruttore (cifre e lettere, maiuscole e minuscole distinte)."
	rtvPwDescDigits = "Password che il programma del costruttore chiede prima di leggere o scrivere la radio; vuota = nessuna password. " +
		"Questo programma legge e scrive la radio anche quando c'è (come CHIRP): qui la vedi, la cambi o la togli, e la modifica arriva alla radio quando la scrivi. " +
		"Lunghezza e caratteri sono quelli del programma del costruttore (solo cifre)."
	rtvPwWarn = "Con una password impostata il programma del costruttore legge e scrive la radio solo se gli si dà la stessa password. " +
		"Da qui la si può sempre vedere e togliere."
	rtvPwWarnHidden = "Il programma del costruttore di questo modello non mostra la casella della password: se ne imposti una, quel programma " +
		"non leggerà e non scriverà più la radio finché non la togli da qui."
)

// rtvPwDef: la voce della password del modello kind (ok falso se non ne ha)
func rtvPwDef(kind string) (setDef, bool) {
	p, ok := rtvPwDefs[kind]
	if !ok {
		return setDef{}, false
	}
	pad := 0xFF
	enc := &setEnc{Kind: "str", Fmt: "skip", Map: map[string]int{}, Pad: &pad}
	for k := 0; k < p.n; k++ {
		enc.Offs = append(enc.Offs, p.off+k)
	}
	chars := rtvPwChars
	for i, c := range chars {
		enc.Map[string(c)] = rtvPwByte(p.tab, i) // in lettura si riconoscono sempre tutti e 62
	}
	d := setDef{Key: "progpw", Group: "Password", Label: rtvPwLabel, ItLabel: rtvPwLabel, Type: "str", MaxLen: p.n,
		Charset: chars, ItDesc: rtvPwDesc, ItWarn: rtvPwWarn, Enc: enc}
	if p.digits {
		d.Charset, d.ItDesc = chars[:10], rtvPwDescDigits
	}
	if p.hidden {
		d.ItWarn = rtvPwWarnHidden
	}
	return d, true
}

var (
	rtvSetMu    sync.Mutex
	rtvSetCache = map[string]*setTable{}
	rtvPwH777   = map[*setTable]*setTable{} // tabella di CHIRP della H777 nuova versione → la stessa con la password in coda
)

// rtvMemSize: memoria del modello (0 se non è uno dei modelli di questa tabella)
func rtvMemSize(m *genModel) int {
	switch o := m.Opt.(type) {
	case *h777v4Opt:
		return o.memsize
	case *t18Opt:
		return o.spec.memsize
	case *rb65Opt:
		return o.spec.memsize
	}
	return 0
}

func rtvSetTable(kind string) *setTable {
	rtvSetMu.Lock()
	defer rtvSetMu.Unlock()
	if t := rtvSetCache[kind]; t != nil {
		return t
	}
	m, defs := genByID[kind], rtvSetDefs[kind]
	if m == nil || defs == nil || rtvMemSize(m) == 0 {
		return nil
	}
	t := &setTable{Class: "RTV-" + kind, Size: rtvMemSize(m), Source: "programma del costruttore " + m.Name}
	for _, d := range defs {
		group := "Funzionamento"
		if strings.HasPrefix(d.key, "key") && d.key != "keylock" {
			group = "Tasti laterali"
		}
		enc := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(d.off), hex.EncodeToString([]byte{d.mask})}}}
		def := setDef{Key: d.key, Group: group, Label: d.label, ItLabel: d.label, Type: "bool", Enc: enc}
		vals := []byte{0, d.mask}
		if d.opts != nil {
			def.Type, def.ItOpts, vals = "list", d.opts, d.vals
			for i := range d.opts {
				def.Options = append(def.Options, strconv.Itoa(i))
			}
		}
		for _, v := range vals {
			s := hex.EncodeToString([]byte{v})
			enc.Table = append(enc.Table, &s)
		}
		t.Settings = append(t.Settings, def)
	}
	if d, ok := rtvPwDef(kind); ok {
		t.Settings = append(t.Settings, d) // sempre l'ultima voce
	}
	rtvSetCache[kind] = t
	return t
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return rtvSetDefs[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := rtvSetTable(kind)
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
	// H777 nuova versione: la tabella di CHIRP (adattata alla memoria aperta, come per
	// ogni radio di CHIRP) con in coda la password di programmazione
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool {
			m := genByID[kind]
			return kind == "rt-h777v4" && m != nil && chirpSettings()[m.RClass] != nil
		},
		get: func(kind string, img []byte) *setTable {
			m := genByID[kind]
			base := chirpSettings()[m.RClass]
			if base == nil {
				return nil
			}
			if base = fitSetTable(m, base, img); base == nil {
				return nil
			}
			d, ok := rtvPwDef(kind)
			if !ok || setDefEnd(&d) > len(img) {
				return base
			}
			rtvSetMu.Lock()
			defer rtvSetMu.Unlock()
			if t := rtvPwH777[base]; t != nil {
				return t
			}
			t := &setTable{Class: base.Class, Size: base.Size, Source: base.Source}
			t.Settings = append(append([]setDef{}, base.Settings...), d)
			rtvPwH777[base] = t
			return t
		},
	})
}
