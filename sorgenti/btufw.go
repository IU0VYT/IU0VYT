// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.8: firmware .BTU di Radtel RT-910 e RT-900 senza Bluetooth e della Bajeton
// BJ-7700 senza Bluetooth, per il programma «Radtel Bootload Tools» (ricavato dal
// suo codice: classe BootHelper, stesso caricatore KDH della serie Baofeng TFT
// 330L, vedi bfkdh.go). Differenze rispetto al .BF:
//   - presentazione "PROGRAM" + "BT80" + "U" a 57600 baud, poi "UPDATE"; la porta
//     si chiude e si riapre a 115200 baud, pausa di 20 ms, comando 42;
//   - modalità manuale (radio accesa tenendo premuti i due tasti laterali sotto
//     il PTT, schermata "UPDATE"; pulsante con la scritta nera): 115200 baud,
//     si parte dalla presentazione 01;
//   - intestazione di 16 byte: byte 0 = numero di aree (1 o 2), byte 1-4 =
//     lunghezza dell'area 1, byte 5-8 = lunghezza dell'area 2 solo se le aree
//     sono 2 (con un'area i byte 5-15 non sono lunghezze); dati in chiaro;
//   - ultimo pacchetto completato con FF; conferma della fine attesa.
// Il caricatore NON controlla il modello: RT-900 e RT-910 usano lo stesso
// programma e la stessa presentazione. Il modello e l'elettronica si riconoscono
// dal nome del file e dall'impronta; serve sempre la conferma scritta.

import (
	"encoding/binary"
	"path"
	"regexp"
	"strings"
)

var btuOpts = kdhOpts{Hello: "PROGRAMBT80U", HelloBaud: 57600, Reopen: true, WaitEnd: true}

// firmware pubblicati da Radtel (radtels.com) e Bajeton (bajeton.com)
var fwKnownBTU = map[string]string{
	"d8278c024da39544326d6a581c17f4ffd061d640104489a4b55fb9d562a7ec5b": "Radtel RT-910 senza Bluetooth, elettronica H1.00, V0.07 del 22/05/2025 (RT910_H1.00_V0.07_250522.BTU)",
	"10418c50f5df2123b332110cfb24df121dbd713cb1acdcae89c2a02ffbd5982b": "Radtel RT-910 senza Bluetooth, elettronica H1.00, V0.08 del 30/06/2025 (RT910_H1.00_V0.08_250630.BTU)",
	"50d5591c0fb5bd87771fff487b14e8d7b199beef6bbd168f632dbd0b588740de": "Radtel RT-910 senza Bluetooth, elettronica H1.00, V0.09 del 15/08/2025 (RT910_H1.00_V0.09_250815.BTU)",
	"91f936fa7a20e94cb8a905baf98558ed92223c8439d695522c4bf13d0aa67d2a": "Radtel RT-910 senza Bluetooth, elettronica H1.00G, V0.10 del 20/11/2025 (RT910_H1.00G_V0.10_251120.BTU)",
	"01666b016fbdfd174637aaf2215990696248eac081e8bd3bd6f26117283ffa25": "Radtel RT-910 senza Bluetooth, elettronica H1.00G, V0.11 del 05/03/2026 (RT910_H1.00G_V0.11_260305.BTU: 15 zone da 64 canali)",
	"3231f685e6e7379dcc6af7a160cdbbfde4e9ec6c9cd620e511094ebf43924edd": "Bajeton BJ-7700 senza Bluetooth (5 W), V1.07 del 03/11/2025 (BJ7700_5W_V1.07_251103.BTU)",
	"b6c10f5b9ae2e7a8aecdafec26449c3ffa4f4bb6429155f410ab873cac2dd897": "Radtel RT-900 senza Bluetooth, elettronica precedente (firmware da 0x08001000), V1.04P del 12/12/2024 (WT_RT_900_512_V1.04P_20241212.BTU, 512 canali)",
	"cef6566a908405f1aeefe090b61799c5d89ce0a1de2fccf1e77c4b6ee1b484b5": "Radtel RT-900 senza Bluetooth, firmware «OPEN» V1.05P del 02/01/2025 (WT_RT900_OP_512_V1.05P_250102.BTU, 512 canali)",
	"be83150aa5468c20c8fbf838c429239f4b353780b1f7ecdcd655e23a792ba4b6": "Radtel RT-900 senza Bluetooth, firmware «OPEN» V1.09P del 11/04/2025 (WT_RT900_OP_512_V1.09P_250411.BTU, 512 canali)",
	"a6763b51cc6063ae32c36542e53001b425c2e2e8f38f4ff57dd5bd4606357840": "Radtel RT-900 senza Bluetooth, elettronica H3.01, V1.14P del 03/11/2025 (RT900_H3.01_OPEN_V1.14P_251103.BTU, 999 canali)",
}

// radio dal nome del file (prima le sigle più lunghe)
var btuTargets = []fwTarget{
	{[]string{"BJ7700"}, "Bajeton BJ-7700 senza Bluetooth (= Radtel RT-910 senza Bluetooth)", "BJ7700", []string{"bj-7700nbt"}},
	{[]string{"RT910"}, "Radtel RT-910 senza Bluetooth (= Bajeton BJ-7700 senza Bluetooth)", "RT910", []string{"bj-7700nbt"}},
	{[]string{"RT900"}, "Radtel RT-900 senza Bluetooth", "RT900", []string{"rt-rt900"}},
}

var (
	reBTUVer   = regexp.MustCompile(`(?i)(?:^|[_\- ])V(\d+(?:\.\d+)+)(P?)`)
	reBTUDate8 = regexp.MustCompile(`(?:^|[_\- ])(20\d{2})(\d{2})(\d{2})(?:$|[_\-. ])`)
)

var reBTUHw = regexp.MustCompile(`(?i)(?:^|[_\- ])H(\d\.\d{2}G?)(?:$|[_\- ])`)

func analyzeBTU(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("btu", "Radtel RT-910 / RT-900 senza Bluetooth", "RT910", name, origin, raw)
	info.Platform = "BT80"
	// sempre la conferma scritta: il caricatore non controlla il modello
	info.Typed, info.Warnings = true, true
	base := strings.TrimSuffix(path.Base(name), path.Ext(name))
	if t := fwTargetFromName(btuTargets, base); t != nil {
		info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
		add("ok", "Il nome del file indica la %s", t.target)
	} else {
		info.Models = nil
		add("warn", "Il nome del file non indica né la RT-910 né la RT-900: il caricatore non controlla il modello, verifica che sia il firmware della tua radio")
	}
	valid := true
	if len(raw) < kdhHdr+kdhPkt || len(raw) > 255*kdhPkt+kdhHdr {
		add("bad", "Dimensione non plausibile per un firmware .BTU (%s)", fmtBytes(len(raw)))
		return img
	}
	areas, s1, s2 := raw[0], be32(raw[1:5]), int64(0)
	if areas == 2 {
		s2 = be32(raw[5:9])
	}
	switch {
	case (areas == 1 || areas == 2) && s1 > 0 && kdhHdr+s1+s2 == int64(len(raw)):
		if areas == 1 {
			add("ok", "Formato .BTU riconosciuto: un'area di firmware di %s", fmtBytes(int(s1)))
		} else {
			add("ok", "Formato .BTU riconosciuto: area del firmware di %s e seconda area di %s", fmtBytes(int(s1)), fmtBytes(int(s2)))
			add("warn", "File con due aree: i firmware conosciuti ne hanno una sola")
		}
	default:
		add("bad", "Intestazione non valida: le lunghezze indicate non corrispondono al file. Non è un firmware .BTU utilizzabile")
		return img
	}
	if (s1+kdhPkt-1)/kdhPkt > 255 || (s2+kdhPkt-1)/kdhPkt > 255 {
		add("bad", "Area troppo grande per il protocollo del caricatore (massimo 255 pacchetti)")
		valid = false
	}
	// tabella dei vettori del Cortex-M (in chiaro): da dove parte il firmware
	r1 := raw[kdhHdr : kdhHdr+s1]
	sp, rv := binary.LittleEndian.Uint32(r1[0:]), binary.LittleEndian.Uint32(r1[4:])
	var org uint32
	switch {
	case rv >= 0x08003000:
		org = 0x08003000
	case rv >= 0x08001000:
		org = 0x08001000
	}
	vecOK := org != 0 && sp >= 0x20000000 && sp <= 0x20020000 && rv&1 == 1 && int64(rv) < int64(org)+s1
	for i := 2; i < 16 && vecOK; i++ {
		v := binary.LittleEndian.Uint32(r1[i*4:])
		if v != 0 && (v&1 == 0 || v < org || int64(v) >= int64(org)+s1) {
			vecOK = false
		}
	}
	if !vecOK {
		add("bad", "L'inizio del file non è un programma valido per queste radio (tabella dei vettori non coerente)")
		valid = false
	} else {
		add("ok", "Programma per microcontrollore ARM caricato dal caricatore a 0x%08X (avvio a 0x%08X)", org, rv&^1)
		if org == 0x08001000 {
			info.Hw = "elettronica con il firmware da 0x08001000 (RT-900 precedente alla H3.01)"
			add("warn", "Firmware per l'elettronica precedente (caricatore più piccolo, firmware da 0x08001000), come la RT-900 V1.04P. "+
				"I firmware «OPEN» dalla V1.05P (Radtel indica l'elettronica H3.01 per la V1.14P) partono da 0x08003000: i due tipi NON sono intercambiabili")
		}
	}
	if m := reBTUHw.FindStringSubmatch(base); m != nil {
		if info.Hw == "" {
			info.Hw = "elettronica H" + strings.ToUpper(m[1])
		}
		add("warn", "Il nome del file indica l'elettronica H%s: verifica la versione della tua radio prima di aggiornare (sulla RT-900 è nel menu 47); "+
			"il costruttore avverte che il firmware vale solo per quella elettronica", strings.ToUpper(m[1]))
	} else {
		add("warn", "Il nome del file non indica la versione dell'elettronica: verifica con il rivenditore che sia quella della tua radio")
	}
	if strings.Contains(strings.ToUpper(base), "5W") {
		add("warn", "Il nome del file indica la versione da 5 W della radio")
	}
	if m := reBTUVer.FindStringSubmatch(base); m != nil {
		info.Version = "V" + m[1] + strings.ToUpper(m[2])
	}
	if m := reDate6.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/20" + m[1]
	} else if m := reBTUDate8.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/" + m[1]
	}
	if k, ok := fwKnownBTU[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else if valid {
		add("warn", "File non presente nell'elenco dei firmware conosciuti da questo programma")
	}
	add("warn", "Il caricatore di queste radio non controlla il modello del firmware: un file sbagliato può rendere la radio inutilizzabile fino a un nuovo aggiornamento. "+
		"La lettura della copia di sicurezza usa lo stesso protocollo su RT-900, RT-910 e BJ-7700, con e senza Bluetooth: conferma che la radio è di questa famiglia, "+
		"non quale modello sia. Una radio con Bluetooth (firmware .anc) ha un altro caricatore: se non risponde come quello delle .BTU l'aggiornamento si ferma prima di inviare dati")
	info.Valid = valid
	return img
}
