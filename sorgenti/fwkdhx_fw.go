// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Controlli dei file firmware per i caricatori KDH più recenti (fwkdhx.go):
// UV-25 Plus (.kdhx), Bajeton BJ-9000 (.Firmware) e UV-9G (.kdh).

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

// radio con il caricatore «Baofeng Bootloader Pro» (piattaforma H802): la
// UV-25 Plus (firmware pubblicato da Baofeng) e le altre radio del programma
// «BF-UVs Pro CPS», lo stesso della UV-25 Plus
var kdhxTargets = []fwTarget{
	{[]string{"UV25"}, "Baofeng UV-25 Plus", "UV25PLUS", []string{"bf-uv25plus"}},
	{[]string{"UV28"}, "Baofeng UV-28 Pro", "UV28PRO", []string{"bf-uv28pro"}},
	{[]string{"UV29", "SW9"}, "Baofeng UV-29 Pro / SW-9", "UV29PRO", []string{"bf-uv29pro"}},
	{[]string{"AR152"}, "Baofeng AR-152 Pro", "AR152PRO", []string{"bf-ar152pro"}},
	{[]string{"UV7B"}, "Baofeng UV-7B", "UV7B", []string{"bf-uv7b"}},
	{[]string{"UV32"}, "Baofeng UV-32", "UV32", []string{"bf-uv32"}},
	{[]string{"UV17"}, "Baofeng UV-17R", "UV17R", []string{"bf-uv17r"}},
	{[]string{"UV21"}, "Baofeng UV-21R / UV-21L / UV-21", "UV21", []string{"bf-uv21r", "bf-uv21l"}},
	{[]string{"T60", "988UV"}, "Baofeng BF-T60 / TR-988UV", "T60", []string{"bf-uv17r"}},
	{[]string{"K5PLUS"}, "Baofeng K5 Plus", "K5PLUS", []string{"bf-k5plus"}},
	{[]string{"5RMPLUS"}, "Baofeng UV-5RM Plus", "5RMPLUS", []string{"bf-uv5rmplus"}},
}

// radtelKdhxTargets (4.3): firmware .kdhx di Radtel per il caricatore «JJCC
// Bootloader Pro» (file pubblicati su radtels.com); l'ordine conta
var radtelKdhxTargets = []fwTarget{
	{[]string{"RT495", "495NUM"}, "Radtel RT-495", "RT495", []string{"rt-rt495"}},
	{[]string{"RT630"}, "Radtel RT-630", "RT630", []string{"rt-rt630"}},
	{[]string{"470SERIAL", "RT470"}, "Radtel RT-470 / RT-470L / RT-470X", "RT470", []string{"rt-rt470", "rt-rt470l", "rt-rt470x"}},
}

// fwKnownRadtel: file .kdhx di Radtel scaricati da radtels.com e provati (4.3)
var fwKnownRadtel = map[string]string{
	"98290d76b7963cd7f225807f8dbe48d3f899c1a25e65fe803fef50494704fea4": "RT-470 / RT-470L / RT-470X scheda nuova (FW 2.x), V2.13C del 25/04/2024 (470 serial MFM_V2.13C_20240425.kdhx)",
	"daf267d76d7fae95941d37fefd66cac01810255ab62ce4c423267a1e53a2f6e4": "RT-470 10 W scheda vecchia (FW 1.x), V1.27A del 23/05/2024 (RT470_AllBandTx_V1.27A_20240523.kdhx)",
	"6b24b01a152bdf27c60e3ef9838bf31f9b2257813bd680c842c67825e7897d5b": "RT-495, V0.07 del 07/12/2023 (495_NUM_ALLBAND_V0.07_20231207.kdhx)",
	"507bb9060c2c609e7a9fb009209870d8bcf05ba37f94ae215f121a9c50736fe4": "RT-630, banda aerea 137 MHz, passo 8,33 kHz e AM/FM (rt-630 8.33AM137_AMFM.kdhx)",
}

// radtelKdhx: il file .kdhx è un firmware di Radtel (caricatore «JJCC»)? Un
// file conosciuto vale più del nome; altrimenti decide il nome (quelli della
// famiglia Baofeng, UV-xx o H802, restano alla UV-25 Plus)
func radtelKdhx(name string, raw []byte) *fwTarget {
	sum := sha256.Sum256(raw)
	if k, ok := fwKnownRadtel[hex.EncodeToString(sum[:])]; ok {
		return fwTargetFromName(radtelKdhxTargets, strings.Fields(k)[0])
	}
	base := normModel(strings.TrimSuffix(path.Base(name), path.Ext(name)))
	if strings.Contains(base, "UV") || strings.Contains(base, "H802") || strings.Contains(base, "BF") {
		return nil
	}
	return fwTargetFromName(radtelKdhxTargets, name)
}

// analyzeKDHX controlla un firmware .kdhx (caricatore «Bootloader Pro»).
func analyzeKDHX(name, origin string, raw []byte) *fwImage {
	if rt := radtelKdhx(name, raw); rt != nil {
		return analyzeRadtelKDHX(name, origin, raw, rt)
	}
	img, info, add := newInfo("kdhx", "radio Baofeng con caricatore «Bootloader Pro» (H802)", "H802", name, origin, raw)
	t := fwTargetFromName(kdhxTargets, name)
	uv25 := t != nil && t.confirm == "UV25PLUS"
	// un file conosciuto vale più del nome: il firmware ufficiale della UV-25
	// Plus resta tale anche se rinominato
	renamed := false
	if k, ok := fwKnownKDH2[info.SHA256]; ok && strings.HasPrefix(k, "UV-25 Plus") && !uv25 {
		t, uv25, renamed = &kdhxTargets[0], true, true
	}
	if t != nil {
		info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	}
	info.Typed = !uv25
	valid := true
	if len(raw) < 8*kdhPkt || len(raw) > 255*kdhPkt {
		add("bad", "Dimensione non plausibile per un firmware .kdhx (%s; il caricatore accetta al massimo 255 KB)", fmtBytes(len(raw)))
		return img
	}
	sp, rv, ok := armVectors(raw, 48)
	if ok {
		add("ok", "Programma per microcontrollore ARM caricato dopo il caricatore (avvio a 0x%08X, stack a 0x%08X)", rv&^1, sp)
	} else {
		add("bad", "L'inizio del file non è un programma valido per queste radio (tabella dei vettori non coerente): non è un firmware .kdhx utilizzabile")
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
	switch {
	case renamed:
		add("warn", "Il nome del file non corrisponde al contenuto: è il firmware ufficiale della UV-25 Plus. Va usato solo su una UV-25 Plus")
		info.Warnings = true
	case t == nil:
		add("warn", "Il nome del file non indica il modello della radio: verifica con il costruttore o il rivenditore che sia il firmware giusto. Scrivi H802 per confermare")
		info.Warnings = true
	case !uv25:
		add("warn", "Il nome del file indica la %s. Baofeng pubblica il firmware .kdhx solo per la UV-25 Plus: usa questo file solo se ti è stato fornito dal costruttore o dal rivenditore per la tua radio", t.target)
		info.Warnings = true
	}
	if reChipNRF.MatchString(base) {
		info.Platform = "NRF"
	}
	if m := reHwPlus.FindStringSubmatch(base); m != nil {
		info.Hw = "elettronica «" + m[1] + "+»"
		add("warn", "Il nome del file contiene «%s+»: secondo la denominazione del costruttore indica le radio a partire da quella versione dell'elettronica. Usalo solo per la radio indicata da Baofeng", m[1])
		info.Warnings = true
	}
	add("warn", "Nella modalità automatica la radio controlla solo il codice della piattaforma (H802), non il modello; nella modalità manuale (radio accesa con SK2+SK3) il caricatore non fa controlli: usa solo il firmware della tua radio")
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

// analyzeRadtelKDHX (4.3): firmware .kdhx di Radtel. Stessi controlli del file
// della UV-25 Plus; il caricatore è il «JJCC Bootloader Pro» (presentazione
// automatica PROGROMJJCCU). La RT-470 ha due schede diverse: il firmware 1.x è
// solo per la scheda vecchia, il 2.x solo per la nuova (Radtel: un firmware
// sbagliato blocca la radio), quindi si chiede di confermare il tipo di scheda.
func analyzeRadtelKDHX(name, origin string, raw []byte, t *fwTarget) *fwImage {
	img, info, add := newInfo("kdhx", t.target+" (caricatore «JJCC Bootloader Pro» di Radtel)", "JJCC", name, origin, raw)
	info.Platform = "JJCC"
	info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	info.Typed = true
	if len(raw) < 8*kdhPkt || len(raw) > 255*kdhPkt {
		add("bad", "Dimensione non plausibile per un firmware .kdhx (%s; il caricatore accetta al massimo 255 KB)", fmtBytes(len(raw)))
		return img
	}
	sp, rv, ok := armVectors(raw, 48)
	if !ok {
		add("bad", "L'inizio del file non è un programma valido per queste radio (tabella dei vettori non coerente): non è un firmware .kdhx utilizzabile")
		return img
	}
	add("ok", "Programma per microcontrollore ARM caricato dopo il caricatore (avvio a 0x%08X, stack a 0x%08X)", rv&^1, sp)
	if len(raw) > 4*kdhPkt {
		a, b := countBXLR(bfCrypt(raw)[2*kdhPkt:]), countBXLR(raw[2*kdhPkt:])
		if a > b {
			add("ok", "Parte cifrata coerente: dopo la decifratura contiene codice ARM riconoscibile")
		} else {
			add("warn", "La parte cifrata non contiene codice riconoscibile dopo la decifratura: file forse danneggiato")
			info.Warnings = true
		}
	}
	nameVersionDate(info, name)
	known, isKnown := fwKnownRadtel[info.SHA256]
	if i := strings.LastIndex(known, "("); isKnown && info.Version == "" && i > 0 {
		nameVersionDate(info, strings.TrimSuffix(known[i+1:], ")")) // file conosciuto ma rinominato
	}
	if t.confirm == "RT470" {
		v := strings.TrimPrefix(info.Version, "V")
		switch {
		case strings.HasPrefix(v, "1"):
			info.Hw = "scheda vecchia (firmware 1.x)"
			add("warn", "Firmware %s per la RT-470 con la scheda VECCHIA (menu 55: FW 1.xx). Su una radio con la scheda nuova (FW 2.0 o successivo) la blocca: controlla il menu 55 prima di continuare", info.Version)
		case strings.HasPrefix(v, "2"):
			info.Hw = "scheda nuova (firmware 2.x)"
			add("warn", "Firmware %s per la RT-470 / RT-470L / RT-470X con la scheda NUOVA (menu 55: FW 2.0 o successivo). Su una radio con la scheda vecchia (FW 1.xx) la blocca: controlla il menu 55 prima di continuare", info.Version)
		default:
			add("bad", "Dal nome del file non si capisce se il firmware è per la scheda vecchia (1.x) o nuova (2.x) della RT-470: usa il file con il nome originale di Radtel")
			return img
		}
		info.Warnings = true
	}
	add("warn", "Nella modalità automatica la radio controlla solo il codice della piattaforma (JJCC); nella modalità manuale (radio accesa tenendo premuti i tasti laterali 2 e 3) il caricatore non fa controlli: usa solo il firmware della tua radio")
	info.Warnings = true
	if isKnown {
		info.Known = known
		add("ok", "File conosciuto: %s", known)
	} else {
		add("warn", "File non presente nell'elenco dei firmware di Radtel conosciuti da questo programma")
	}
	info.Valid = true
	return img
}
