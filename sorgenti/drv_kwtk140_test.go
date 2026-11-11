//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

// 4.18: prove dei Kenwood TK-2140 / TK-3140. CHIRP ha l'immagine di prova della
// TK-3140K: per gli altri modelli si usa la stessa memoria con la loro sigla.
func init() {
	chirpFamilyImage["main.kwx140Fam"] = "Kenwood_TK-3140K.img"
	for _, m := range genModels {
		o, ok := m.Opt.(*kwx140Opt)
		if !ok || m.ID == "kw-tk3140k" {
			continue
		}
		id := o.id
		chirpImagePatch[m.ID] = func(b []byte) { copy(b[kwx140ID:], id) }
	}
}
