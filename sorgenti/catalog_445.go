// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Catalogo 4.45: le radio DMR che il programma libero dmrconfig gestisce con lo
// stesso codice di radio già presenti: famiglia AnyTone (drv_anytone.go, la stessa
// delle BTECH DMR-6X2) e famiglia TYT MD-UV380 (drv_tyt.go, la stessa della
// Baofeng DM-1701).

const (
	dmrconfigSite = "https://github.com/OpenRTX/dmrconfig"
	at445         = "Radio DMR con 4000 canali, collegata con il suo cavo USB (porta seriale). Memoria e protocollo dei programmi liberi " +
		"dmrconfig e qdmr, verificati con il codice originale di dmrconfig. Si scrivono solo canali, zone e liste di scansione. " +
		"«Identifica» la prova dopo tutte le altre radio."
	tyt445 = "Con il firmware originale: stessa memoria e stesso protocollo della Baofeng DM-1701, via USB in modalità DFU (il programma " +
		"libero dmrconfig usa lo stesso codice per questi modelli)."
)

var catalog445 = []catItem{
	{ID: "at-d868uv", Brand: "AnyTone", Name: "AT-D868UV", Aliases: []string{"D868UV", "AT-D868"}, Type: "dmr", Drivers: []string{"at-d868uv"},
		Links: []catLink{info("Sito AnyTone", atWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite), info("qdmr (programma libero per radio DMR, GPL)", qdmrSite)},
		Notes: at445},
	{ID: "at-d878uv", Brand: "AnyTone", Name: "AT-D878UV / AT-D878UV Plus", Aliases: []string{"D878UV", "AT-D878UV Plus", "D878UV Plus"}, Models: []string{"AT-D878UV", "AT-D878UV Plus"},
		Type: "dmr", Drivers: []string{"at-d878uv"},
		Links: []catLink{info("Sito AnyTone", atWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite), info("qdmr (programma libero per radio DMR, GPL)", qdmrSite)},
		Notes: at445},
	{ID: "at-d878uv2", Brand: "AnyTone", Name: "AT-D878UV II Plus", Aliases: []string{"D878UVII", "AT-D878UVII Plus", "D878UV2"}, Type: "dmr", Drivers: []string{"at-d878uv2"},
		Links: []catLink{info("Sito AnyTone", atWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite), info("qdmr (programma libero per radio DMR, GPL)", qdmrSite)},
		Notes: at445},
	{ID: "tyt-md-uv380", Brand: "TYT", Name: "MD-UV380", Aliases: []string{"MDUV380"}, Type: "dmr", Drivers: []string{"tyt-mduv380", "ogd77-uv380"},
		Links: []catLink{info("Sito TYT", tytWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
	{ID: "tyt-md-uv390", Brand: "TYT", Name: "MD-UV390", Aliases: []string{"MDUV390"}, Type: "dmr", Drivers: []string{"tyt-mduv390", "ogd77-uv380"},
		Links: []catLink{info("Sito TYT", tytWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
	{ID: "tyt-md-2017", Brand: "TYT", Name: "MD-2017", Aliases: []string{"MD2017"}, Type: "dmr", Drivers: []string{"tyt-md2017", "ogd77-uv380"},
		Links: []catLink{info("Sito TYT", tytWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
	{ID: "tyt-md-9600", Brand: "TYT", Name: "MD-9600", Aliases: []string{"MD9600"}, Type: "dmr", Drivers: []string{"tyt-md9600", "ogd77-uv380"},
		Links: []catLink{info("Sito TYT", tytWeb), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
	{ID: "rtv-rt3s", Brand: "Retevis", Name: "RT3S", Aliases: []string{"RT-3S", "TYT MD-UV390"}, Type: "dmr", Drivers: []string{"tyt-mduv390", "ogd77-uv380"},
		Links: []catLink{info("Programmi Retevis", rtvDown), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
	{ID: "rtv-rt82", Brand: "Retevis", Name: "RT82", Aliases: []string{"RT-82", "TYT MD-2017"}, Type: "dmr", Drivers: []string{"tyt-md2017", "ogd77-uv380"},
		Links: []catLink{info("Programmi Retevis", rtvDown), info("dmrconfig (programma libero da riga di comando, licenza BSD)", dmrconfigSite)},
		Notes: tyt445 + ogd77Note},
}
