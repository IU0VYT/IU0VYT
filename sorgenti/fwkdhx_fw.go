// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Controlli dei file firmware per i caricatori KDH più recenti (fwkdhx.go):
// UV-25 Plus (.kdhx), Bajeton BJ-9000 (.Firmware) e UV-9G (.kdh).

import (
	"encoding/binary"
	"path"
	"regexp"
	"strings"
)

var fwKnownKDH2 = map[string]string{
	"ff199d6b56ac4833d01925de20f897fa5f58f558c510088adf5f1e293f0e95a5": "UV-25 Plus, V0.23 del 22/05/2025 (UV25Pro_NRF_401+_V0.23_250522.kdhx, pubblicato da Baofeng)",
	"cbaad8d2ca956f481d2e9524f102d0281a19da1d74cb6cb048d9e4ee549c263d": "BJ-9000, V0.27 del 13/03/2026 (BJ9000_V0.27_260313.Firmware, pubblicato da Bajeton)",
	"c873df7d189b51b692eee266725c8d732081cf21060243b3b335337b0cd1717e": "UV-9G, V02 del 04/10/2023 (T52GMRS_M8_V02_231004.kdh, pubblicato da Baofeng: trasmissione sul canale 15)",
}

var (
	reVerShort = regexp.MustCompile(`(?i)(?:^|[_\- ])V(\d+)(?:$|[_\- ])`)
	reHwPlus   = regexp.MustCompile(`(?:^|[_\- ])(\d{3})\+(?:$|[_\- ])`)
)

// armVectors controlla la tabella dei vettori di un programma per Cortex-M
// all'inizio del file (in chiaro anche nei file cifrati).
func armVectors(b []byte, n int) (sp, reset uint32, ok bool) {
	if len(b) < n*4 {
		return 0, 0, false
	}
	sp, reset = binary.LittleEndian.Uint32(b), binary.LittleEndian.Uint32(b[4:])
	inFlash := func(v uint32) bool { return v&1 == 1 && v >= 0x08000000 && v < 0x08000000+uint32(len(b))+0x10000 }
	if sp < 0x20000000 || sp > 0x20040000 || !inFlash(reset) {
		return sp, reset, false
	}
	for i := 2; i < n; i++ {
		if v := binary.LittleEndian.Uint32(b[i*4:]); v != 0 && !inFlash(v) {
			return sp, reset, false
		}
	}
	return sp, reset, true
}

func nameVersionDate(info *fwInfo, name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	if m := reVersion.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1]
	} else if m := reVerShort.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1]
	}
	if m := reDate6.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/20" + m[1]
	}
	return base
}

func knownKDH2(info *fwInfo, add func(level, f string, a ...interface{})) {
	if k, ok := fwKnownKDH2[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware conosciuti da questo programma")
		info.Warnings = true
	}
}

// analyzeKDHX controlla un firmware .kdhx della UV-25 Plus.
func analyzeKDHX(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("kdhx", "Baofeng UV-25 Plus", "UV25PLUS", name, origin, raw)
	valid := true
	if len(raw) < 8*kdhPkt || len(raw) > 255*kdhPkt {
		add("bad", "Dimensione non plausibile per un firmware della UV-25 Plus (%s; il caricatore accetta al massimo 255 KB)", fmtBytes(len(raw)))
		return img
	}
	sp, rv, ok := armVectors(raw, 48)
	if ok {
		add("ok", "Programma per microcontrollore ARM caricato dopo il caricatore (avvio a 0x%08X, stack a 0x%08X)", rv&^1, sp)
	} else {
		add("bad", "L'inizio del file non è un programma valido per questa radio (tabella dei vettori non coerente): non è un firmware .kdhx utilizzabile")
		return img
	}
	if len(raw) > 4*kdhPkt {
		a, b := countBXLR(bfCrypt(raw)[2*kdhPkt:]), countBXLR(raw[2*kdhPkt:])
		if a > b {
			add("ok", "Parte cifrata coerente: dopo la decifratura contiene codice ARM riconoscibile")
		} else {
			add("warn", "La parte cifrata non contiene codice riconoscibile dopo la decifratura: file forse danneggiato")
			info.Warnings = true
		}
	}
	base := nameVersionDate(info, name)
	up := strings.ToUpper(base)
	if !strings.Contains(up, "UV25") && !strings.Contains(up, "UV-25") {
		add("warn", "Il nome del file non indica la UV-25 Plus: verifica che sia il firmware giusto")
		info.Warnings = true
	}
	if reChipNRF.MatchString(base) {
		info.Platform = "NRF"
	}
	if m := reHwPlus.FindStringSubmatch(base); m != nil {
		info.Hw = "elettronica «" + m[1] + "+»"
		add("warn", "Il nome del file contiene «%s+»: secondo la denominazione del costruttore indica le radio a partire da quella versione dell'elettronica. Usalo solo per una UV-25 Plus come indicato da Baofeng", m[1])
		info.Warnings = true
	}
	add("warn", "Nella modalità automatica la radio controlla il proprio codice di modello (H802) prima di aggiornarsi; nella modalità manuale (radio accesa con SK2+SK3) il caricatore non fa controlli: usa solo il firmware della UV-25 Plus")
	info.Warnings = true
	knownKDH2(info, add)
	info.Valid = valid
	return img
}

// bjModel: modello dichiarato nel firmware della BJ-9000 (controllato dalla radio)
func bjModel(raw []byte) (string, bool) {
	if len(raw) < 1024 {
		return "", false
	}
	s := strings.TrimRight(string(raw[992:1004]), " \x00")
	if len(s) < 3 {
		return "", false
	}
	for _, c := range []byte(s) {
		if c < 0x20 || c > 0x7E {
			return "", false
		}
	}
	return s, true
}

// analyzeBJ9000 controlla un firmware .Firmware della Bajeton BJ-9000.
func analyzeBJ9000(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("bj9000", "Bajeton BJ-9000", "BJ9000", name, origin, raw)
	if len(raw) < 16*kdhPkt || len(raw) > 2*1024*1024 {
		add("bad", "Dimensione non plausibile per un firmware della BJ-9000 (%s)", fmtBytes(len(raw)))
		return img
	}
	sp, rv, ok := armVectors(raw, 48)
	if !ok {
		add("bad", "L'inizio del file non è un programma valido per questa radio (tabella dei vettori non coerente)")
		return img
	}
	add("ok", "Programma per microcontrollore ARM (avvio a 0x%08X, stack a 0x%08X)", rv&^1, sp)
	model, mok := bjModel(raw)
	if !mok {
		add("bad", "Il file non contiene il codice del modello che la radio controlla prima di aggiornarsi: non è un firmware della BJ-9000")
		return img
	}
	info.Hw = "modello «" + model + "»"
	add("ok", "Codice del modello nel firmware: «%s». La radio lo confronta con il proprio e, se è diverso, rifiuta l'aggiornamento prima di cancellare la memoria", model)
	nameVersionDate(info, name)
	if !strings.Contains(strings.ToUpper(name), "9000") {
		add("warn", "Il nome del file non indica la BJ-9000: verifica che sia il firmware giusto")
		info.Warnings = true
	}
	knownKDH2(info, add)
	info.Valid = true
	return img
}

// analyzeKDH9G controlla un firmware .kdh della UV-9G.
func analyzeKDH9G(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("kdh9g", "Baofeng UV-9G", "UV9G", name, origin, raw)
	if len(raw) < 4*1024 || len(raw) > k9gMax {
		add("bad", "Dimensione non plausibile per un firmware della UV-9G (%s; il caricatore accetta al massimo 64 KB)", fmtBytes(len(raw)))
		return img
	}
	add("ok", "Firmware di %s (%d blocchi da 32 byte)", fmtBytes(len(raw)), (len(raw)+k9gBlock-1)/k9gBlock)
	base := nameVersionDate(info, name)
	up := strings.ToUpper(base)
	if !strings.Contains(up, "T52") && !strings.Contains(up, "9G") {
		add("warn", "Il nome del file non indica la UV-9G (i firmware del costruttore si chiamano T52GMRS_…): verifica che sia quello giusto")
		info.Warnings = true
	}
	add("warn", "Il caricatore della UV-9G non controlla il modello della radio né il contenuto del file: un firmware sbagliato la rende inutilizzabile. Usa solo il file pubblicato da Baofeng per la UV-9G")
	info.Warnings = true
	knownKDH2(info, add)
	info.Valid = true
	return img
}
