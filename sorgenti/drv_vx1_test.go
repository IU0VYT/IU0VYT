//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: Yaesu VX-1, aggiunte per i confronti con CHIRP (TestChirp*).

// lockedSlot: nelle prove generali le modifiche si fanno sulle memorie complete
// del gruppo 1 (posizioni 1-52); il gruppo 2, i canali speciali e le onde medie
// hanno regole loro e si confrontano con CHIRP in TestYaesuVX1Chirp
func (vx1Fam) lockedSlot(m *genModel, i int) bool { return i >= 52 }
