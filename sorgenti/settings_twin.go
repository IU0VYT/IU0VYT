// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle «gemelle» (4.42): radio che CHIRP non ha (NoChirp) ma che il
// programma del costruttore tratta con la stessa mappa della memoria di una radio di
// CHIRP (RClass). Fino alla 4.41 avevano solo i canali, perché la tabella di CHIRP
// vale per i modelli di CHIRP. Qui la tabella della classe di CHIRP si usa solo dopo
// averla confrontata, voce per voce, con il programma del costruttore del modello, e
// solo per le voci su cui i due sono d'accordo:
//
//   only   si tengono solo i gruppi di voci controllati sul programma;
//   drop   voci tolte: il programma non le ha o le scrive in modo diverso da CHIRP;
//   limit  scelte ridotte alle prime n: il programma ne ha meno di CHIRP.
//
// Baofeng UV-17R, UV-21R, UV-21L: il programma «T6UV Series EN CPS» dei tre pacchetti
// è, byte per byte, lo stesso file dei pacchetti di UV-17 Pro, UV-5RM, UV-5RH e
// K5 Plus, cioè delle radio per cui è nata la classe UV17Pro di CHIRP: tabella intera.
// La funzione Read del programma legge il blocco delle impostazioni (64 byte) alle
// stesse posizioni della struttura «settings» di CHIRP (squelch 0, risparmio 1, VOX 2,
// luce 3, doppio ascolto 4, TOT 5, beep 6, voce 7 e 8, … «hang-up» 0x39).
//
// Baofeng UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro: programma «BF-UVs CPS»
// (cinque versioni diverse). Il blocco delle impostazioni è letto e scritto alle
// stesse posizioni della classe UV17ProGPS in tutte e cinque, con tre differenze da
// CHIRP, uguali in tutte: il tasto laterale 2 (0x34) è scritto come numero d'ordine
// 0-3 e non con i codici dei tasti di CHIRP; il beep ha due scelte (spento, acceso) e
// non quattro; il ritardo dell'identificativo (0x0C) in lettura è ridotto a 0-6; il
// fuso orario del GPS non c'è. Quelle voci sono tolte o ridotte.
//   - UV-25 Plus e UV-28 Pro: il file del programma è lo stesso, byte per byte, di
//     quello della UV-32 e della UV-28 Plus (classi di CHIRP UV32 e UV28Plus, che
//     derivano da UV17ProGPS): restano anche VFO, DTMF e nomi dei banchi.
//   - UV-29 Pro, UV-7B e AR-152 Pro: le funzioni dei VFO e dei DTMF sono diverse da
//     quelle del programma della UV-32 (l'AR-152 Pro ha anche un blocco DTMF in più):
//     restano solo le voci del blocco delle impostazioni.
//
// Baofeng T22 (PMR446): il programma «Baofeng T22 CPS» legge e scrive il blocco a
// 0x100 alle posizioni della classe W31E di CHIRP (voce 0x12F, TOT 0x130, squelch
// 0x134, risparmio 0x135, beep 0x136, VOX 0x139, livello 0x13A, ritardo 0x13B, tasto
// 0x13E). Tolte «voice» (nel programma l'ordine è spento, inglese, cinese; in CHIRP
// spento, cinese, inglese) e «pf2key» (quattro scelte diverse dalle tre di CHIRP);
// il livello del VOX arriva a 5 e il ritardo ha sei valori (0,5-3 s).

import (
	"strings"
	"sync"
)

type twinSet struct {
	class string
	prog  string
	only  []string
	drop  []string
	limit map[string]int
}

var (
	twinUVsDrop  = []string{"settings.gpstimezone", "settings.key2short"}
	twinUVsLimit = map[string]int{"settings.beep": 2, "settings.pttdly": 7}
	twinT6       = twinSet{class: "UV17Pro", prog: "«T6UV Series EN CPS»"}
	twinUVs      = twinSet{class: "UV17ProGPS", prog: "«BF-UVs CPS»", drop: twinUVsDrop, limit: twinUVsLimit}
	twinUVsBase  = twinSet{class: "UV17ProGPS", prog: "«BF-UVs CPS»", only: []string{"settings."}, drop: twinUVsDrop, limit: twinUVsLimit}
)

var twinSettings = map[string]twinSet{
	"bf-uv17r": twinT6, "bf-uv21r": twinT6, "bf-uv21l": twinT6,
	"bf-uv25plus": twinUVs, "bf-uv28pro": twinUVs,
	"bf-uv29pro": twinUVsBase, "bf-uv7b": twinUVsBase, "bf-ar152pro": twinUVsBase,
	"bf-t22": {class: "W31E", prog: "«Baofeng T22 CPS» (PMR446)",
		only:  []string{"squelch", "tot", "vox", "voxgain", "voxdelay", "save", "beep"},
		limit: map[string]int{"voxgain": 6, "voxdelay": 6}},
}

// twinMatch: p è una chiave intera oppure, se finisce con «.» o «/», un prefisso
func twinMatch(key string, ps []string) bool {
	for _, p := range ps {
		if key == p || ((strings.HasSuffix(p, ".") || strings.HasSuffix(p, "/")) && strings.HasPrefix(key, p)) {
			return true
		}
	}
	return false
}

func twinTable(tw twinSet) *setTable {
	src := chirpSettings()[tw.class]
	if src == nil {
		return nil
	}
	// «CHIRP+»: l'interfaccia scrive che la tabella è di CHIRP, verificata sul programma
	t := &setTable{Class: src.Class, Size: src.Size, Source: "CHIRP+" + t17Src(tw.prog)}
	for _, d := range src.Settings {
		if len(tw.only) > 0 && !twinMatch(d.Key, tw.only) {
			continue
		}
		if twinMatch(d.Key, tw.drop) {
			continue
		}
		if n, ok := tw.limit[d.Key]; ok && d.Type == "list" && d.Enc != nil && n < len(d.Options) {
			e := *d.Enc
			e.Table = e.Table[:n:n]
			d.Enc = &e
			d.Options = d.Options[:n:n]
		}
		t.Settings = append(t.Settings, d)
	}
	return t
}

var (
	twinOnce   sync.Once
	twinTables map[string]*setTable
)

// twinTablesMap: le tabelle si preparano alla prima richiesta (serve la tabella di
// CHIRP, che si legge una volta sola)
func twinTablesMap() map[string]*setTable {
	twinOnce.Do(func() {
		twinTables = map[string]*setTable{}
		byDef := map[string]*setTable{}
		for id, tw := range twinSettings {
			m := genByID[id]
			if m == nil || !m.NoChirp || m.RClass != tw.class {
				continue // la gemella deve dichiarare la stessa classe di CHIRP
			}
			k := tw.class + "|" + strings.Join(tw.only, ",") + "|" + strings.Join(tw.drop, ",") + "|" + tw.prog
			if byDef[k] == nil {
				byDef[k] = twinTable(tw)
			}
			if byDef[k] != nil {
				twinTables[id] = byDef[k]
			}
		}
	})
	return twinTables
}

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool {
			if _, ok := twinSettings[kind]; !ok {
				return false
			}
			return twinTablesMap()[kind] != nil
		},
		get: func(kind string, img []byte) *setTable {
			if _, ok := twinSettings[kind]; !ok {
				return nil
			}
			t := twinTablesMap()[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
}
