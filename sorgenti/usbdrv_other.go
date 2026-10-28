//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import "errors"

func usbDriverStatus() map[string]interface{} {
	return map[string]interface{}{"supported": false, "present": false, "driver": ""}
}

func runDriverInstaller(path string) error {
	return errors.New("l'installazione dei driver USB è disponibile solo su Windows")
}

func startProgram(path string, admin bool) error {
	return errors.New("i programmi del costruttore si avviano solo su Windows")
}
