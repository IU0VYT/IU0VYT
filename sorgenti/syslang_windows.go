// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "golang.org/x/sys/windows"

// systemLangs (4.12): lingue di visualizzazione di Windows dell'utente, nell'ordine
// di preferenza (Impostazioni → Ora e lingua → Lingua), es. ["it-IT", "en-US"]
func systemLangs() []string {
	l, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil || len(l) == 0 {
		if l, err = windows.GetSystemPreferredUILanguages(windows.MUI_LANGUAGE_NAME); err != nil {
			return nil
		}
	}
	return l
}
