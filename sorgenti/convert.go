// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed templates/k67_base.xlc
var baseXLC []byte

//go:embed templates/bajeton_base.dat
var baseDAT []byte

const (
	kindXLC = "k67"
	kindDAT = "baj"
)

func detect(b []byte) string {
	switch {
	case isDAT(b):
		return kindDAT
	case isXLC(b):
		return kindXLC
	}
	return ""
}

func radioName(kind string) string {
	if m, ok := genByID[kind]; ok {
		return m.Name
	}
	if kind == kindXLC {
		return "Baofeng K67 Pro"
	}
	return "Bajeton 8300"
}

func radioSlots(kind string) int {
	if m, ok := genByID[kind]; ok {
		return m.Slots
	}
	if kind == kindXLC {
		return xlcChannels
	}
	return datZones * datPerZone
}

func readRadio(b []byte, rep *Report) (string, []*Chan, error) {
	switch detect(b) {
	case kindXLC:
		return kindXLC, readXLC(b, rep), nil
	case kindDAT:
		chs, err := readDAT(b, rep)
		return kindDAT, chs, err
	}
	return "", nil, fmt.Errorf("formato non riconosciuto: il file non sembra né un .xlc del Baofeng K67 Pro né un .dat della Bajeton 8300")
}

func writeRadio(kind string, base []byte, chs []*Chan, rep *Report) ([]byte, error) {
	if kind == kindXLC {
		if base == nil {
			base = baseXLC
		}
		return writeXLC(base, chs, rep), nil
	}
	if base == nil {
		base = baseDAT
	}
	return writeDAT(base, chs, rep)
}

func posLabel(kind string, i int) string {
	if m, ok := genByID[kind]; ok {
		return fmt.Sprintf("%s canale %d", m.Model, i+m.First)
	}
	if kind == kindXLC {
		return fmt.Sprintf("K67 canale %d", i+1)
	}
	return fmt.Sprintf("Bajeton zona %d n. %d", i/datPerZone+1, i%datPerZone+1)
}

// adaptChan prepara un canale (proveniente dall'altra radio o dall'archivio)
// per la radio "to". La logica è quella del convertitore originale.
// bands: gamme del K67 (se note) per il controllo fuori banda.
func adaptChan(c *Chan, to string, bands []band, label string, rep *Report) *Chan {
	n := *c
	n.Raw = "" // i dati non interpretati appartengono all'altra radio
	if n.TxOff {
		rep.warn("%s: canale di sola ricezione (trasmissione disabilitata), non rappresentabile sulla %s: canale non copiato", label, radioName(to))
		return nil
	}
	name := asciiName(n.Name)
	if name != n.Name {
		rep.warn("%s: nome \"%s\" → \"%s\" (caratteri non ammessi sostituiti)", label, n.Name, name)
	}
	if to == kindDAT {
		if sn := shortName(name, 12); sn != name {
			rep.warn("%s: nome \"%s\" accorciato in \"%s\" (max 12 caratteri)", label, name, sn)
			name = sn
		}
	} else {
		if len(name) > 16 {
			rep.warn("%s: nome \"%s\" accorciato a 16 caratteri", label, name)
			name = name[:16]
		}
		if n.Power == 1 {
			rep.warn("%s: potenza Media non disponibile sul K67, impostata Alta", label)
			n.Power = 0
		}
	}
	n.Name = name
	if n.Power < 0 || n.Power > 2 {
		n.Power = 0
	}
	for _, t := range []*string{&n.TxTone, &n.RxTone} {
		lbl := "TX"
		if t == &n.RxTone {
			lbl = "RX"
		}
		if *t == "" {
			continue
		}
		if v, ok := parseTone(*t); ok {
			*t = toneStr(v)
			continue
		}
		if dcsSupported(*t, to) {
			*t = strings.ToUpper(strings.TrimSpace(*t))
			continue
		}
		rep.warn("%s: tono %s \"%s\" non trasferibile, impostato OFF", label, lbl, *t)
		*t = ""
	}
	if to == kindXLC && len(bands) > 0 {
		if !inBands(n.Rx, bands) {
			rep.warn("%s: %s MHz fuori dalle gamme della radio", label, fmtMHz(n.Rx))
		}
	}
	if n.Rx >= 1000000000 && to == kindXLC {
		rep.warn("%s: %s MHz non rappresentabile nel file del K67", label, fmtMHz(n.Rx))
	}
	return &n
}

// Convert trasforma un file nell'altro formato (usata da test e riga di comando).
func Convert(src, base []byte) ([]byte, string, *Report, error) {
	rep := &Report{}
	kind, chs, err := readRadio(src, rep)
	if err != nil {
		return nil, "", nil, err
	}
	tgt := kindDAT
	if kind == kindDAT {
		tgt = kindXLC
	}
	if base != nil && detect(base) != tgt {
		return nil, "", nil, fmt.Errorf("il file base scelto non è un %s", radioName(tgt))
	}
	var bands []band
	if tgt == kindXLC {
		b := base
		if b == nil {
			b = baseXLC
		}
		bands = xlcBands(b, 0x10)
	}
	out := make([]*Chan, len(chs))
	for i, c := range chs {
		if c != nil {
			out[i] = adaptChan(c, tgt, bands, posLabel(tgt, i), rep)
		}
	}
	if base == nil {
		if tgt == kindXLC {
			base = blankXLC(baseXLC)
		} else {
			base, err = blankDAT(baseDAT)
			if err != nil {
				return nil, "", nil, err
			}
		}
	}
	res, err := writeRadio(tgt, base, out, rep)
	return res, tgt, rep, err
}

// dcsSupported: codice DCS ("D023N", "D754I") accettato dalla radio indicata.
func dcsSupported(t, kind string) bool {
	code, inv, ok := dcsParse(t)
	if !ok {
		return false
	}
	if kind == kindXLC {
		return true // il K67 accetta qualsiasi codice a tre cifre ottali
	}
	_ = inv
	return btDCSIndex(strings.ToUpper(strings.TrimSpace(t))) >= 0 && code != ""
}
