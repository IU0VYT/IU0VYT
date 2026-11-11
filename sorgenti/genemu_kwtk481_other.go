//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radio simulata Kenwood TK-481 / TK-981: quella delle TK-280 / TK-380 (stesso protocollo).

func (kwx481Fam) emu(m *genModel, img []byte) genEmu { return kwx80Fam{}.emu(m, img) }
