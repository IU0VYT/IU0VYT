// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.47: modelli ricavati dai programmi dei costruttori (punto 4 dell'ordine dei
// lavori). Retevis B3B, B63B, RT22P, RT622P: famiglia della H777 nuova versione
// (drv_h777v4.go). B3B e RT22P sono le versioni FRS (USA), B63B e RT622P le
// versioni PMR446 (Europa), come dicono i file di configurazione dei loro programmi.

const rtv447 = "Stessa memoria dei canali della Retevis H777 nuova versione; protocollo ricavato dal programma del costruttore. " +
	"Canali (16, senza nomi) e impostazioni, password di programmazione compresa: una password già presente non ferma la lettura né la scrittura. " +
	"Non provata su una radio vera: prima la sola lettura."

const rtv447b = "Protocollo ricavato dal programma del costruttore. Canali (senza nomi) e impostazioni, password di programmazione compresa: " +
	"una password già presente non ferma la lettura né la scrittura. " +
	"Non provata su una radio vera: prima la sola lettura."

var catalog447 = []catItem{
	{ID: "rtv-b3b", Brand: "Retevis", Name: "B3B", Type: "gmrs", Drivers: []string{"rt-b3b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	{ID: "rtv-b63b", Brand: "Retevis", Name: "B63B", Type: "pmr", Drivers: []string{"rt-b63b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	{ID: "rtv-rt22p", Brand: "Retevis", Name: "RT22P", Type: "gmrs", Drivers: []string{"rt-rt22p"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	{ID: "rtv-rt622p", Brand: "Retevis", Name: "RT622P", Type: "pmr", Drivers: []string{"rt-rt622p"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	// RB666 e RB666P (PMR446): stessa famiglia, la radio conferma con 66 e può rimandare ciò che riceve
	{ID: "rtv-rb666", Brand: "Retevis", Name: "RB666", Type: "pmr", Drivers: []string{"rt-rb666"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	{ID: "rtv-rb666p", Brand: "Retevis", Name: "RB666P", Type: "pmr", Drivers: []string{"rt-rb666p"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447},
	// RB45 e RB47 (FRS, 22 canali), RB645 (PMR446, 16 canali): canali da 11 byte, come dicono i loro programmi
	{ID: "rtv-rb45", Brand: "Retevis", Name: "RB45", Type: "gmrs", Drivers: []string{"rt-rb45"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb645", Brand: "Retevis", Name: "RB645", Type: "pmr", Drivers: []string{"rt-rb645"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb47", Brand: "Retevis", Name: "RB47", Type: "gmrs", Drivers: []string{"rt-rb47"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	// B3H e RT68H (FRS), B63H, RB647 e RT668H (PMR446): blocchi da 8 byte, canali come nella famiglia della T18 (drv_small.go)
	{ID: "rtv-b3h", Brand: "Retevis", Name: "B3H", Type: "gmrs", Drivers: []string{"rt-b3h"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-b63h", Brand: "Retevis", Name: "B63H", Type: "pmr", Drivers: []string{"rt-b63h"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb647", Brand: "Retevis", Name: "RB647", Type: "pmr", Drivers: []string{"rt-rb647"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rt68h", Brand: "Retevis", Name: "RT68H", Type: "gmrs", Drivers: []string{"rt-rt68h"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rt668h", Brand: "Retevis", Name: "RT668H", Type: "pmr", Drivers: []string{"rt-rt668h"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-nr10", Brand: "Retevis", Name: "NR10", Type: "gmrs", Drivers: []string{"rt-nr10"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-nr610", Brand: "Retevis", Name: "NR610", Type: "pmr", Drivers: []string{"rt-nr610"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb66s", Brand: "Retevis", Name: "RB66S", Type: "gmrs", Drivers: []string{"rt-rb66s"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb666s", Brand: "Retevis", Name: "RB666S", Type: "pmr", Drivers: []string{"rt-rb666s"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	// RB65 (FRS, 22 canali): una sola frequenza e un solo tono per canale (drv_rb65.go)
	{ID: "rtv-rb65", Brand: "Retevis", Name: "RB65", Type: "gmrs", Drivers: []string{"rt-rb65"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	// RB66 e RB66P (FRS, 22 canali): una frequenza e un tono per canale, collegamento delle RB666 (drv_rb66.go)
	{ID: "rtv-rb66", Brand: "Retevis", Name: "RB66", Type: "gmrs", Drivers: []string{"rt-rb66"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	// RT68B (FRS), RT668B (PMR446) e RB669 (32 canali: PMR446 e, nel file di fabbrica, i canali liberi tedeschi a 149 MHz): famiglia della T18, programmi della generazione successiva
	{ID: "rtv-rt68b", Brand: "Retevis", Name: "RT68B", Type: "gmrs", Drivers: []string{"rt-rt68b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rt668b", Brand: "Retevis", Name: "RT668B", Type: "pmr", Drivers: []string{"rt-rt668b"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb669", Brand: "Retevis", Name: "RB669", Type: "pmr", Drivers: []string{"rt-rb669"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
	{ID: "rtv-rb66p", Brand: "Retevis", Name: "RB66P", Type: "gmrs", Drivers: []string{"rt-rb66p"}, Links: []catLink{info("Programmi Retevis", rtvDown)}, Notes: rtv447b},
}
