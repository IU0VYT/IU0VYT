// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// API del driver USB della DM-1701 (3.8): stato e installazione assistita.

import (
	"net/http"
)

func (a *App) apiUSBDriver(r *http.Request) (interface{}, error) {
	var req struct {
		Action string `json:"action"` // status | install
		Path   string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if req.Action != "install" {
		return usbDriverStatus(), nil
	}
	p := req.Path
	if p == "" {
		var err error
		if p, err = a.dialog(false, "Driver USB della DM-1701: scegli dpinst_amd64.exe (o il file .inf) nella cartella «DM-1701 USB Driver»",
			"Installatore del driver (dpinst*.exe, *.inf)|dpinst*.exe;*.inf|Tutti i file|*.*", "", ""); err != nil {
			return nil, err
		}
	}
	if err := runDriverInstaller(p); err != nil {
		return nil, err
	}
	return map[string]interface{}{"started": true, "path": p}, nil
}
