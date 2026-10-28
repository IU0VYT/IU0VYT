// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// BTECH DMR-6X2 / DMR-6X2 PRO (3.8): aggiornamento GUIDATO del firmware con il
// programma ufficiale.
//
// Il firmware si scrive con «Firmware_Update.exe» del pacchetto BTECH (menu
// Tool → Firmware Upgrade del CPS, o cartella Update del CPS installato):
// radio in modalità aggiornamento, porta COM del «USB Virtual Com Port Driver»,
// 921600 baud, Duplex, Write (istruzioni «STOP and README FIRST.pdf» dei
// pacchetti DMR-6X2-221b e DMR-6X2-PRO-121e). Il protocollo di scrittura non è
// pubblicato e il programma ufficiale è compilato (Visual Basic 6): qui NON si
// scrive il firmware. Il programma controlla i file scelti (modello, tipo di
// aggiornamento, versione, coerenza delle dimensioni), mostra la procedura
// giusta e avvia il programma ufficiale.
//
// File del pacchetto (struttura ricavata dai file ufficiali):
//   .spi  u32 0x20, u16 numero di file, u32 dimensione totale, zeri, poi (solo
//         nel firmware principale) 36 byte di identificativo, es.
//         «D6X2UV  V100DMR-6X2         V210    » (DMR-6X2) o
//         «D6X2UV2 V100D6X2UV2         V120    » (DMR-6X2 PRO);
//   .CDI  un elemento da 278 byte per file: nome (256 byte), u32 indirizzo,
//         u32 dimensione, …;
//   .CDD  i dati, uno dopo l'altro (dimensione = somma dei file).

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const cdiEntry = 278

type fw6x2File struct {
	Name string `json:"name"`
	Addr uint32 `json:"addr"`
	Size uint32 `json:"size"`
}

type fw6x2Info struct {
	Path    string      `json:"path"`
	Kind    string      `json:"kind"`  // main | icons | board
	Model   string      `json:"model"` // DMR-6X2 | DMR-6X2 PRO | "" (icone: entrambe)
	Version string      `json:"version"`
	Date    string      `json:"date"`
	Ident   string      `json:"ident"`
	Files   []fw6x2File `json:"files"`
	Total   int         `json:"total"`
	Checks  []string    `json:"checks"` // controlli superati
	Errors  []string    `json:"errors"` // problemi
	Title   string      `json:"title"`
}

var fw6x2NameRe = regexp.MustCompile(`(?i)_V(\d+)\.(\d+)([a-z]?)(?:_(\d{4})(\d{2})(\d{2}))?`)

// check6x2: controlla i tre file (.spi, .CDI, .CDD) con lo stesso nome
func check6x2(p string) (*fw6x2Info, error) {
	stem := strings.TrimSuffix(p, filepath.Ext(p))
	read := func(ext string) ([]byte, error) {
		for _, e := range []string{ext, strings.ToLower(ext), strings.ToUpper(ext)} {
			if b, err := os.ReadFile(stem + e); err == nil {
				return b, nil
			}
		}
		return nil, fmt.Errorf("manca il file %s accanto a %s: servono i tre file .spi, .CDI e .CDD del pacchetto", filepath.Base(stem)+ext, filepath.Base(p))
	}
	spi, err := read(".spi")
	if err != nil {
		return nil, err
	}
	cdi, err := read(".CDI")
	if err != nil {
		return nil, err
	}
	cdd, err := read(".CDD")
	if err != nil {
		return nil, err
	}
	in := &fw6x2Info{Path: stem + ".spi", Files: []fw6x2File{}, Checks: []string{}, Errors: []string{}}
	if len(spi) < 10 || binary.LittleEndian.Uint32(spi) != 0x20 {
		return nil, errors.New("il file .spi non è un file di aggiornamento delle radio DMR-6X2")
	}
	n := int(binary.LittleEndian.Uint16(spi[4:]))
	total := binary.LittleEndian.Uint32(spi[6:])
	if len(spi) >= 51 {
		in.Ident = strings.Join(strings.Fields(string(bytes.TrimRight(spi[14:50], "\x00 "))), " ")
	}
	if n == 0 || len(cdi) != n*cdiEntry {
		in.Errors = append(in.Errors, fmt.Sprintf("il file .CDI ha %d byte, attesi %d (%d file)", len(cdi), n*cdiEntry, n))
	}
	var sum uint32
	for i := 0; i < n && (i+1)*cdiEntry <= len(cdi); i++ {
		e := cdi[i*cdiEntry:]
		name := string(bytes.TrimRight(bytes.SplitN(e[:256], []byte{0}, 2)[0], " "))
		f := fw6x2File{Name: name, Addr: binary.LittleEndian.Uint32(e[256:]), Size: binary.LittleEndian.Uint32(e[260:])}
		in.Files = append(in.Files, f)
		sum += f.Size
	}
	in.Total = len(cdd)
	if sum == total && int(total) == len(cdd) {
		in.Checks = append(in.Checks, fmt.Sprintf("dimensioni coerenti: %d file, %d byte (.spi, .CDI e .CDD concordano)", n, len(cdd)))
	} else {
		in.Errors = append(in.Errors, fmt.Sprintf("dimensioni non coerenti: .spi %d byte, somma del .CDI %d, .CDD %d: file incompleto o danneggiato, non usarlo", total, sum, len(cdd)))
	}
	if m := fw6x2NameRe.FindStringSubmatch(filepath.Base(stem)); m != nil {
		in.Version = strings.TrimLeft(m[1], "0") + "." + m[2] + m[3]
		if strings.HasPrefix(in.Version, ".") {
			in.Version = "0" + in.Version
		}
		if m[4] != "" {
			in.Date = m[6] + "/" + m[5] + "/" + m[4]
		}
	}
	switch {
	case len(in.Files) == 1 && strings.EqualFold(in.Files[0].Name, "DMR_6X2UV.bin"):
		in.Kind, in.Model = "main", "DMR-6X2"
	case len(in.Files) == 1 && strings.EqualFold(in.Files[0].Name, "DMR_6X2UV_II.bin"):
		in.Kind, in.Model = "main", "DMR-6X2 PRO"
	case len(in.Files) > 1 && strings.HasPrefix(strings.ToUpper(in.Files[0].Name), "GBK"):
		in.Kind = "icons"
	case len(in.Files) == 1 && strings.EqualFold(in.Files[0].Name, "APRS_REC.bin"):
		in.Kind = "board"
	default:
		return nil, fmt.Errorf("aggiornamento non riconosciuto (%s): questo controllo conosce solo i pacchetti BTECH DMR-6X2 e DMR-6X2 PRO", in.Ident)
	}
	if in.Kind == "main" {
		// identificativo nel .spi coerente con il file del .CDI
		pro := strings.HasPrefix(in.Ident, "D6X2UV2")
		if pro != (in.Model == "DMR-6X2 PRO") {
			in.Errors = append(in.Errors, fmt.Sprintf("identificativo %q non coerente con il file %s", in.Ident, in.Files[0].Name))
		} else {
			in.Checks = append(in.Checks, fmt.Sprintf("identificativo del modello nel .spi: %q", in.Ident))
		}
		// tabella dei vettori del microcontrollore ARM: stack in RAM e avvio dentro il firmware
		f := in.Files[0]
		if len(cdd) >= 8 {
			sp, rv := binary.LittleEndian.Uint32(cdd), binary.LittleEndian.Uint32(cdd[4:])
			if sp>>24 == 0x20 && rv&^1 >= f.Addr && rv&^1 < f.Addr+f.Size {
				in.Checks = append(in.Checks, fmt.Sprintf("inizio del firmware valido (caricato a 0x%08X, avvio a 0x%08X)", f.Addr, rv&^1))
			} else {
				in.Errors = append(in.Errors, "l'inizio del firmware non è quello di un programma per il microcontrollore della radio")
			}
		}
	}
	switch in.Kind {
	case "main":
		in.Title = "Firmware " + in.Model
	case "icons":
		in.Title = "Icone (DMR-6X2 e DMR-6X2 PRO)"
	case "board":
		in.Title = "Scheda APRS / baseband (solo se indicato da BTECH)"
	}
	if in.Version != "" {
		in.Title += " versione " + in.Version
	}
	return in, nil
}

func (a *App) apiFw6x2(r *http.Request) (interface{}, error) {
	var req struct {
		Action string `json:"action"` // check | launch
		Path   string `json:"path"`
		Admin  bool   `json:"admin"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	switch req.Action {
	case "check":
		p := req.Path
		if p == "" {
			var err error
			if p, err = a.dialog(false, "File di aggiornamento DMR-6X2 (.spi)", "Aggiornamento DMR-6X2 (*.spi)|*.spi|Tutti i file|*.*", "", ""); err != nil {
				return nil, err
			}
		}
		return check6x2(p)
	case "launch":
		p := req.Path
		if p == "" {
			init := ""
			for _, d := range []string{`C:\DMR_6X2_2.21\Update`, `C:\DMR-6X2-PRO-1.21\Update`, `C:\DMR_6X2_1.21\Update`} {
				if dirExists(d) {
					init = d
					break
				}
			}
			var err error
			if p, err = a.dialog(false, "Programma ufficiale di aggiornamento (Firmware_Update.exe, cartella Update del CPS)",
				"Firmware_Update.exe|Firmware_Update*.exe|Tutti i file|*.*", init, ""); err != nil {
				return nil, err
			}
		}
		if !strings.HasPrefix(strings.ToLower(filepath.Base(p)), "firmware_update") || !strings.EqualFold(filepath.Ext(p), ".exe") {
			return nil, errors.New("scegli Firmware_Update.exe nella cartella Update del CPS della DMR-6X2 (es. C:\\DMR_6X2_2.21\\Update)")
		}
		if err := startProgram(p, req.Admin); err != nil {
			return nil, err
		}
		return map[string]interface{}{"started": true, "path": p}, nil
	}
	return nil, errors.New("operazione sconosciuta")
}
