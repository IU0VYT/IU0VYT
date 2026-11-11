//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

// 4.18: prove dei Kenwood TK-280 / TK-380. CHIRP ha l'immagine di prova della
// TK-280: per la TK-380 si usa la stessa memoria con la sua sigla.
func init() {
	chirpFamilyImage["main.kwx80Fam"] = "Kenwood_TK-280.img"
	for _, m := range genModels {
		o, ok := m.Opt.(*kwx80Opt)
		if !ok || m.ID == "kw-tk280" {
			continue
		}
		typ := o.typ
		chirpImagePatch[m.ID] = func(b []byte) { copy(b[kwx80ID:], typ) }
	}
}
