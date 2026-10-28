// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Driver USB delle radio in modalità DFU (DM-1701 / Retevis RT84), 3.8:
// stato del collegamento e installazione assistita del driver del costruttore.
//
// Un driver di Windows (kernel) non si può includere nel programma né
// ridistribuire: si usa quello del pacchetto ufficiale «DM-1701 CPS» (cartella
// «DM-1701 USB Driver»), scelto dall'utente. Il programma avvia, con la
// richiesta di autorizzazione di Windows (UAC), l'installatore del driver
// (dpinst*.exe) oppure «pnputil /add-driver … /install» per un file .inf. In
// alternativa la radio funziona con il driver generico WinUSB (Zadig).

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// usbDriverStatus: radio DFU collegata e driver in uso
func usbDriverStatus() map[string]interface{} {
	st := map[string]interface{}{"supported": true, "present": false, "driver": ""}
	guid := windows.GUID{Data1: 0x3fe809ab, Data2: 0xfb91, Data3: 0x4cb5, Data4: [8]byte{0xa6, 0x43, 0x69, 0x67, 0x0d, 0x52, 0x36, 0x6e}}
	if len(usbDevicePaths(guid)) > 0 {
		st["present"], st["driver"] = true, "STTub30 (driver del costruttore)"
		return st
	}
	paths := usbPresent(0x0483, 0xDF11)
	if len(paths) == 0 {
		return st
	}
	st["present"] = true
	for _, p := range paths {
		if port, err := openWinUSBPort(p); err == nil {
			_ = port.Close()
			st["driver"] = "WinUSB"
			return st
		}
	}
	return st
}

// runDriverInstaller: avvia l'installatore scelto con i privilegi di amministratore
func runDriverInstaller(path string) error {
	base := strings.ToLower(filepath.Base(path))
	dir := filepath.Dir(path)
	var file, args string
	switch {
	case strings.HasSuffix(base, ".inf"):
		file = filepath.Join(windowsDir(), "System32", "pnputil.exe")
		args = fmt.Sprintf("/add-driver \"%s\" /install", path)
	case strings.HasPrefix(base, "dpinst") && strings.HasSuffix(base, ".exe"):
		file = path
	default:
		return errors.New("scegli l'installatore del driver (dpinst_amd64.exe o dpinst_x86.exe) o il file .inf del driver")
	}
	verb, _ := windows.UTF16PtrFromString("runas")
	f, _ := windows.UTF16PtrFromString(file)
	var a *uint16
	if args != "" {
		a, _ = windows.UTF16PtrFromString(args)
	}
	d, _ := windows.UTF16PtrFromString(dir)
	if err := windows.ShellExecute(0, verb, f, a, d, windows.SW_SHOWNORMAL); err != nil {
		if errors.Is(err, windows.ERROR_CANCELLED) {
			return errors.New("installazione annullata: Windows non ha avuto l'autorizzazione")
		}
		return fmt.Errorf("impossibile avviare l'installazione del driver: %v", err)
	}
	return nil
}

func windowsDir() string {
	if d, err := windows.GetSystemWindowsDirectory(); err == nil && d != "" {
		return d
	}
	return `C:\Windows`
}

// startProgram: avvia un programma del costruttore (come amministratore se richiesto)
func startProgram(path string, admin bool) error {
	v := "open"
	if admin {
		v = "runas"
	}
	verb, _ := windows.UTF16PtrFromString(v)
	f, _ := windows.UTF16PtrFromString(path)
	d, _ := windows.UTF16PtrFromString(filepath.Dir(path))
	if err := windows.ShellExecute(0, verb, f, nil, d, windows.SW_SHOWNORMAL); err != nil {
		if errors.Is(err, windows.ERROR_CANCELLED) {
			return errors.New("avvio annullato: Windows non ha avuto l'autorizzazione")
		}
		return fmt.Errorf("impossibile avviare %s: %v", filepath.Base(path), err)
	}
	return nil
}
