//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Versione per Linux/macOS, usata per sviluppo e test automatici:
//   gestionecanali                      avvia l'interfaccia (indirizzo stampato a video)
//   gestionecanali converti IN OUT [BASE]  converte un file .xlc <-> .dat

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

var fileDialogNative = false

func main() {
	if len(os.Args) >= 4 && os.Args[1] == "converti" {
		src, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var base []byte
		if len(os.Args) > 4 {
			if base, err = os.ReadFile(os.Args[4]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		out, _, rep, err := Convert(src, base)
		if err != nil {
			fmt.Println("ERRORE:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(os.Args[3], out, 0o644); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, w := range rep.Warnings {
			fmt.Println("-", w)
		}
		return
	}
	st, err := openStore(appDataDir())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if os.Getenv("GESTIONE_CANALI_EMU") == "1" {
		setupEmulatedRadios()
	}
	// solo per le prove (versione Linux): pacchetto degli aggiornamenti simulato
	if u := os.Getenv("GESTIONE_CANALI_UPD_URL"); u != "" {
		updateSource = u
	}
	app := newApp(st)
	if t := os.Getenv("GESTIONE_CANALI_TOKEN"); t != "" {
		app.token = t
	}
	ln, url, err := app.listen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(url)
	_ = http.Serve(ln, app.handler())
}

// Radio simulate per le prove dell'interfaccia (GESTIONE_CANALI_EMU=1):
// COM3 = K67 Pro, COM4 = Bajeton 8300, COM5 = nessuna radio collegata,
// COM6 = Bajeton 8300 bloccata in modalità aggiornamento del firmware.
type slowPort struct{ Port }

func (s slowPort) Write(b []byte) error { time.Sleep(4 * time.Millisecond); return s.Port.Write(b) }

func setupEmulatedRadios() {
	k67 := &k67Emu{mem: xlcImageForRadio(baseXLC), model: "5RH-PRO"}
	if os.Getenv("GESTIONE_CANALI_EMU_PWD") != "" {
		k67.password = os.Getenv("GESTIONE_CANALI_EMU_PWD")
	}
	btImg, err := datToEEPROM(baseDAT)
	if err != nil {
		panic(err)
	}
	bt := newBtEmu(btImg)
	// aggiornamento del firmware: radio accesa (COM4) e radio bloccata in modalità aggiornamento (COM6)
	ancOn := newAncEmu(nil, false)
	ancOn.planFor(677306, 0)
	ancOn.reboot = "app"
	ancOff := newAncEmu(nil, true)
	ancOff.planFor(677306, 0)
	ancOff.reboot = "boot"
	// radio generica (driver di CHIRP) letta da un file .img: COM7
	var gen genEmu
	genName := ""
	if p := os.Getenv("GESTIONE_CANALI_EMU_IMG"); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			if m, mem, _, err := detectGen(b, os.Getenv("GESTIONE_CANALI_EMU_MODEL")); err == nil {
				gen, genName = genEmuFor(m, mem), m.Name
			}
		}
	}
	// firmware Baofeng: UV-5RM accesa (COM8, .BF) e 5RH Pro in modalità aggiornamento (COM9, .dat);
	// caricatori KDH più recenti, radio accese: UV-25 Plus (COM10), BJ-9000 (COM11), UV-9G (COM12).
	// Una radio simulata per porta, che resta la stessa tra una riapertura e l'altra
	// (cambi di velocità durante l'aggiornamento).
	kdhx, bj9k, k9g := newKdh2Emu(kdhxProto, nil, false), newKdh2Emu(bj9000Proto, nil, false), newK9gEmu(nil, false)
	var kdh *kdhEmu
	var rh5 *rh5Emu
	listPortsFunc = func() []PortInfo {
		l := []PortInfo{{"COM3", "USB-SERIAL CH340 (COM3) - K67 simulata"}, {"COM4", "Prolific USB-to-Serial (COM4) - Bajeton simulata"}, {"COM5", "Porta senza radio (COM5)"},
			{"COM6", "USB-SERIAL CH340 (COM6) - Bajeton simulata bloccata in modalità aggiornamento"}}
		if gen != nil {
			l = append(l, PortInfo{"COM7", "USB-SERIAL CH340 (COM7) - " + genName + " simulata"})
		}
		return append(l, PortInfo{"COM8", "USB-SERIAL CH340 (COM8) - UV-5RM simulata (firmware)"}, PortInfo{"COM9", "USB-SERIAL CH340 (COM9) - 5RH Pro simulata (firmware)"},
			PortInfo{"COM10", "USB-SERIAL CH340 (COM10) - UV-25 Plus simulata (firmware)"}, PortInfo{"COM11", "USB-SERIAL CH340 (COM11) - BJ-9000 simulata (firmware)"},
			PortInfo{"COM12", "USB-SERIAL CH340 (COM12) - UV-9G simulata (firmware)"})
	}
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		switch name {
		case "COM3":
			go k67.serve(b)
		case "COM4":
			if baud == 57600 && ancOn.idle() {
				go bt.serve(b)
			} else {
				go ancOn.serve(b, baud)
			}
		case "COM6":
			go ancOff.serve(b, baud)
		case "COM7":
			if gen == nil {
				return nil, fmt.Errorf("impossibile aprire la porta %s: porta inesistente", name)
			}
			go gen.serve(b)
		case "COM8":
			// la radio simulata riceve il firmware scelto (si confronta con il file inviato)
			kdh = newKdhEmu(nil, false)
			go func(e *kdhEmu) { e.serve(b, baud) }(kdh)
		case "COM9":
			rh5 = newRh5Emu(nil, true)
			go func(e *rh5Emu) { e.serve(b, baud) }(rh5)
		case "COM10":
			go kdhx.serve(b, baud)
		case "COM11":
			go bj9k.serve(b, baud)
		case "COM12":
			go k9g.serve(b, baud)
		case "COM5":
		default:
			return nil, fmt.Errorf("impossibile aprire la porta %s: porta inesistente", name)
		}
		return slowPort{a}, nil
	}
}
