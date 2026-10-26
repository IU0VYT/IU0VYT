// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Controlli dei file firmware Baofeng: .BF (caricatore KDH V1.0, serie TFT
// 330L: UV-5RM, UV-5RH) e .dat della 5RH Pro (anche K67 Pro). Servono a non
// inviare alla radio un file sbagliato o danneggiato.

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"
)

// firmware conosciuti: raccolti dalla comunità (non pubblicati dal costruttore)
var fwKnownBaofeng = map[string]string{
	"2b9aa883196cbc57c12e1ec5b62ccd062dc589253d5cfacb50eef6faec171441": "UV-5RM/5RH, chip BK4819, V0.14 del 23/11/2023 (5RH_BK_V0.14_231123.BF)",
	"db58ac9be488032ee7c3e0ea9283f17239df1e018a1095c2514541c38bb8602b": "UV-5RM/5RH, chip BK4819, V0.13 del 20/11/2023 (5RH_BK_V0.13_231120.BF)",
	"575a1a4deb551d242f88aad597196c8b84bc336e799ff14ef62190ae7d9ba61c": "UV-5RM/5RH, chip BK4819, V0.12 cinese tradizionale del 27/10/2023 (5RH_BK_Fangti_V0.12_231027.BF)",
	"d46830b56cfba51100f7ad91b5ec9895dca5ccce13320565476e8a063bd1b4bc": "UV-5RM/5RH, chip BK4819, V0.11 cinese tradizionale del 06/10/2023 (5RH_BK_Fangti_V0.11_231006.BF)",
	"ecc0eb191b2303a04bb2a9217e9cbe3fe83f525c0110be8a6b01c7a088253129": "UV-5RH «L», chip AT1846S, V0.07 del 18/09/2023 (5RH_AT1846S_V0.07_FangtiBlueBG_230918.BF)",
	"e9680597d870ebf71ff823c99f7080f10388b089a3be69a01a2adf2ffa920759": "UV-5RH, chip FD6818 (NRF), V0.13 del 21/11/2023 (5RH_NRF_V0.13_231121.BF)",
	"376c93ee3ae1462a77c3a2c84ada3a780a6e9dee551f2f1aca4e15714b545664": "UV-5RH, chip FD6818 (NRF), V0.14 cinese tradizionale del 24/11/2023 (5RH_NRF_Fangti_V0.14_231124.BF)",
	"07ed427a05939132f6e0aa34a51301ad2627e19dd64ee2ead6b684e3315051df": "5RH Pro, elettronica V1, firmware 1.0.92 (BF_5RH_501_v1_0_92.dat)",
	"6baaa6504a1281e877789692a22ef456177aeaa9c5df2e74a8605e6b94d43beb": "5RH Pro, elettronica V1, firmware 1.0.97 (BF_5RH_501_v1_0_97.dat)",
	"db3817c49b4ac63808c31913a2c83db4debae827a9464498ddd583eee28a47a5": "5RH Pro, elettronica V1, firmware 1.1.00 (BF_5RH_501_v1_1_00.dat)",
	"3b648891f2c5d6894a3b5fb7425b672f8db069f8d31740133bc9c5e8a2460c0a": "5RH Pro, elettronica V1, firmware 1.1.04 (BF_5RH_501_v1_1_04.dat)",
	"8103671917511f462124bf2b6036460074bf7aaeabdd8c407c913a090b0103d2": "5RH Pro, elettronica V2, firmware 2.0.7 (BF_5RH_501_v2_0_7.dat)",
	"db97ac883de493720e8dc1677fe1e77be2aa6d981d6d30a0033a47590ef84855": "5RH Pro, elettronica V2, firmware 2.0.9 (BF_5RH_501_v2_0_9.dat)",
}

var (
	reDate6   = regexp.MustCompile(`(?:^|[_\- ])(\d{2})(\d{2})(\d{2})(?:$|[_\-. ])`)
	reVerUnd  = regexp.MustCompile(`(?i)(?:^|[_\- ])v(\d+)_(\d+)_(\d+)`)
	reChipBK  = regexp.MustCompile(`(?i)(?:^|[_\- ])BK(?:4819)?(?:$|[_\- ])`)
	reChipAT  = regexp.MustCompile(`(?i)AT1846`)
	reChipNRF = regexp.MustCompile(`(?i)(?:^|[_\- ])(NRF|FD6818)(?:$|[_\- ])`)
)

// fwTarget: radio a cui è destinato un firmware, riconosciuta dal nome del file
// (i file del costruttore iniziano con la sigla del modello: «5RH_BK_…»,
// «UV25Pro_NRF_…»). keys: sigle normalizzate (maiuscole, senza - _ . e spazi).
type fwTarget struct {
	keys    []string
	target  string
	confirm string
	drivers []string
}

func fwTargetFromName(tbl []fwTarget, name string) *fwTarget {
	base := normModel(strings.TrimSuffix(path.Base(name), path.Ext(name)))
	for i := range tbl {
		for _, k := range tbl[i].keys {
			if strings.Contains(base, k) {
				return &tbl[i]
			}
		}
	}
	return nil
}

// radio della serie TFT 330L e derivate (stesso caricatore KDH V1.0 della
// UV-5RM secondo il programma del costruttore e CHIRP): l'ordine conta
// (prima le sigle più lunghe)
var bfTargets = []fwTarget{
	{[]string{"5RMPLUS"}, "Baofeng UV-5RM Plus", "5RMPLUS", []string{"bf-uv5rmplus"}},
	{[]string{"K5PLUS"}, "Baofeng K5 Plus", "K5PLUS", []string{"bf-k5plus"}},
	{[]string{"17RPLUS"}, "Baofeng UV-17R Plus", "UV17RPLUS", []string{"bf-uv17rplus"}},
	{[]string{"UV17R", "UV17"}, "Baofeng UV-17R", "UV17R", []string{"bf-uv17r"}},
	{[]string{"UV21L"}, "Baofeng UV-21L", "UV21L", []string{"bf-uv21l"}},
	{[]string{"UV21"}, "Baofeng UV-21R / UV-21", "UV21R", []string{"bf-uv21r"}},
	{[]string{"T60", "988UV"}, "Baofeng BF-T60 / TR-988UV", "T60", []string{"bf-uv17r"}},
	{[]string{"5RH", "5RM"}, "Baofeng UV-5RM / UV-5RH (serie TFT 330L)", "5RM", []string{"bf-uv5rm", "bf-uv5rh"}},
}

func newInfo(kind, target, confirm, name, origin string, raw []byte) (*fwImage, *fwInfo, func(level, f string, a ...interface{})) {
	img := &fwImage{Raw: raw, Dec: raw}
	img.Info = fwInfo{Kind: kind, Target: target, Confirm: confirm, Name: name, Origin: origin, Size: len(raw)}
	sum := sha256.Sum256(raw)
	img.Info.SHA256 = hex.EncodeToString(sum[:])
	info := &img.Info
	return img, info, func(level, f string, a ...interface{}) {
		info.Checks = append(info.Checks, fwCheck{Level: level, Text: fmt.Sprintf(f, a...)})
	}
}

// bfCrypt: cifratura dei file .BF (operazione simmetrica; strumento di OK2MOP e
// BD4VOW): blocchi da 1024 byte, i primi due e l'ultimo in chiaro; degli altri
// (tranne il penultimo) quelli con indice%3 = 1 con "KDHT", = 2 con "RBGI".
func bfCrypt(b []byte) []byte {
	out := append([]byte{}, b...)
	n := (len(b) + kdhPkt - 1) / kdhPkt
	for i := 2; i < n-1; i++ {
		var key []byte
		switch {
		case i%3 == 1 && i < n-2:
			key = ancKeyA
		case i%3 == 2 && i < n-2:
			key = ancKeyB
		}
		if key == nil {
			continue
		}
		end := min((i+1)*kdhPkt, len(b))
		for j := i * kdhPkt; j < end; j++ {
			v, s := out[j], key[(j-i*kdhPkt)%4]
			if v != 0 && v != 0xFF && v != s && v != s^0xFF {
				out[j] = v ^ s
			}
		}
	}
	return out
}

func countBXLR(b []byte) int {
	n := 0
	for i := 0; i+1 < len(b); i += 2 {
		if b[i] == 0x70 && b[i+1] == 0x47 {
			n++
		}
	}
	return n
}

// analyzeBF controlla un file .BF.
func analyzeBF(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("bf", "Baofeng UV-5RM / UV-5RH (serie TFT 330L)", "5RM", name, origin, raw)
	info.Models = []string{"bf-uv5rm", "bf-uv5rh"}
	_, known := fwKnownBaofeng[info.SHA256]
	if t := fwTargetFromName(bfTargets, name); t != nil && t.confirm != "5RM" && known {
		// un file conosciuto vale più del nome: è della UV-5RM/5RH
		add("warn", "Il nome del file indica la %s, ma il contenuto è un firmware conosciuto della UV-5RM/5RH: va usato solo su quelle radio", t.target)
		info.Warnings = true
	} else if t != nil {
		info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
		if t.confirm != "5RM" {
			// Baofeng non pubblica firmware .BF; quelli raccolti dalla comunità
			// sono tutti della UV-5RM/5RH: per le altre radio conferma scritta
			info.Typed = true
			add("warn", "Il nome del file indica la %s: questo programma conosce solo i firmware .BF della UV-5RM/5RH. Usalo solo se ti è stato fornito dal costruttore o dal rivenditore per la tua radio: il caricatore non controlla il modello", t.target)
			info.Warnings = true
		}
	}
	valid := true
	if len(raw) < kdhHdr+kdhPkt || len(raw) > 512*1024 {
		add("bad", "Dimensione non plausibile per un firmware .BF (%s)", fmtBytes(len(raw)))
		return img
	}
	s1, s2 := be32(raw[1:5]), be32(raw[5:9])
	if raw[0] == 2 && kdhHdr+s1+s2 == int64(len(raw)) && s1 > 0 {
		add("ok", "Formato .BF riconosciuto: area del firmware di %s e area di configurazione di %s", fmtBytes(int(s1)), fmtBytes(int(s2)))
	} else {
		add("bad", "Intestazione non valida: le lunghezze delle aree non corrispondono al file. Non è un firmware .BF utilizzabile")
		return img
	}
	if s1 > 255*kdhPkt || s2 > 255*kdhPkt {
		add("bad", "Aree troppo grandi per il protocollo del caricatore (massimo 255 pacchetti)")
		valid = false
	} else if s1 > 60*1024 || s2 > 4*1024 {
		add("warn", "Aree più grandi della memoria della UV-5RM/5RH (60 KB + 4 KB): forse il firmware di una radio diversa")
		info.Warnings = true
	}
	// tabella dei vettori del Cortex-M (primi 2 KB in chiaro)
	r1 := raw[kdhHdr : kdhHdr+s1]
	sp, rv := binary.LittleEndian.Uint32(r1[0:]), binary.LittleEndian.Uint32(r1[4:])
	vecOK := sp >= 0x20000000 && sp <= 0x20010000 && rv&1 == 1 && rv >= 0x08001000 && int64(rv) < 0x08001000+s1
	for i := 2; i < 16 && vecOK && len(r1) >= 64; i++ {
		v := binary.LittleEndian.Uint32(r1[i*4:])
		// gestori nella flash del firmware o nell'area 2 (0x1FFFE400, ex caricatore di sistema)
		inFw := v >= 0x08001000 && int64(v) < 0x08001000+s1
		inCfg := v >= 0x1FFFE400 && int64(v) < 0x1FFFE400+s2
		if v != 0 && (v&1 == 0 || !(inFw || inCfg)) {
			vecOK = false
		}
	}
	if vecOK {
		add("ok", "Programma per microcontrollore ARM caricato dopo il caricatore (avvio a 0x%08X, stack a 0x%08X)", rv&^1, sp)
	} else {
		add("bad", "L'inizio dell'area del firmware non è un programma valido per queste radio (tabella dei vettori non coerente)")
		valid = false
	}
	if s1 > 4*kdhPkt {
		dec := bfCrypt(r1)
		a, b := countBXLR(dec[2*kdhPkt:]), countBXLR(r1[2*kdhPkt:])
		if a > b {
			add("ok", "Parte cifrata coerente: dopo la decifratura contiene codice ARM riconoscibile")
		} else {
			add("warn", "La parte cifrata non contiene codice riconoscibile dopo la decifratura: file forse danneggiato")
			info.Warnings = true
		}
	}
	base := strings.TrimSuffix(name, path.Ext(name))
	switch {
	case reChipBK.MatchString(base):
		info.Platform = "BK4819"
	case reChipAT.MatchString(base):
		info.Platform = "AT1846S"
	case reChipNRF.MatchString(base):
		info.Platform = "FD6818 (NRF)"
	}
	if info.Platform != "" {
		info.Hw = "chip radio " + info.Platform
		add("warn", "Il nome del file indica le radio con chip %s: verifica che la tua sia uguale (etichetta sotto la batteria; le versioni con chip diversi NON sono intercambiabili)", info.Platform)
	} else {
		add("warn", "Il nome del file non indica il chip radio a cui è destinato: verifica con il rivenditore che sia quello giusto")
	}
	info.Warnings = true
	if m := reVersion.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1]
	}
	if m := reDate6.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/20" + m[1]
	}
	if k, ok := fwKnownBaofeng[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s (raccolto dalla comunità radioamatoriale; Baofeng non pubblica questi firmware)", k)
	} else if valid {
		add("warn", "File non presente nell'elenco dei firmware conosciuti da questo programma")
	}
	info.Valid = valid
	return img
}

// analyze5RH controlla un file .dat della 5RH Pro.
func analyze5RH(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("rh5", "Baofeng 5RH Pro (e K67 Pro / radio con il CPS «5RH_PRO»)", "5RHPRO", name, origin, raw)
	valid := true
	if len(raw) < rh5Head+rh5Block || len(raw) > rh5Head+rh5Max {
		add("bad", "Dimensione non plausibile per un firmware della 5RH Pro (%s)", fmtBytes(len(raw)))
		return img
	}
	if !strings.HasPrefix(string(raw), "BaoFeng") || string(raw[16:22]) != "BF_5RH" || raw[32] != 'V' {
		add("bad", "Intestazione «BaoFeng / BF_5RH» assente: non è un firmware della 5RH Pro")
		return img
	}
	size := rh5Size(raw)
	switch {
	case size <= 0 || size > rh5Max:
		add("bad", "Lunghezza dei dati non valida nell'intestazione")
		return img
	case rh5Head+size > int64(len(raw)):
		add("bad", "File incompleto: mancano %s di dati", fmtBytes(int(rh5Head+size-int64(len(raw)))))
		return img
	case rh5Head+size < int64(len(raw)):
		add("warn", "Il file contiene %s in più oltre i dati indicati nell'intestazione (non vengono inviati)", fmtBytes(int(int64(len(raw))-rh5Head-size)))
		info.Warnings = true
	default:
		add("ok", "Formato riconosciuto: firmware «BF_5RH» di %s", fmtBytes(int(size)))
	}
	if rh5IsV2(raw) {
		info.Hw = "elettronica V2 (firmware 2.x)"
	} else {
		info.Hw = "elettronica V1 (firmware 1.x)"
	}
	add("ok", "Destinato alle radio con %s: la radio dichiara la sua versione all'inizio e, se è diversa, l'aggiornamento non parte", info.Hw)
	base := strings.TrimSuffix(name, path.Ext(name))
	if m := reVerUnd.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1] + "." + m[2] + "." + m[3]
		if (m[1] == "2") != rh5IsV2(raw) {
			add("warn", "Il numero di versione nel nome (%s) non corrisponde all'elettronica indicata nel file", info.Version)
			info.Warnings = true
		}
	} else if m := reVersion.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1]
	}
	if k, ok := fwKnownBaofeng[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s (raccolto dalla comunità radioamatoriale)", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware conosciuti da questo programma")
		info.Warnings = true
	}
	info.Valid = valid
	return img
}
