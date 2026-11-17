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
	"encoding/hex"
	"strings"
	"sync"
)

type twinSet struct {
	class string
	prog  string
	only  []string
	drop  []string
	limit map[string]int
	// 4.43: gemelle con una classe propria (RClass diversa) ma con il blocco delle
	// impostazioni allo stesso indirizzo di una classe di CHIRP
	anyClass bool
	within   [2]int              // se non zero: solo le voci con tutti i byte in [inizio, fine)
	size     int                 // se non zero: lunghezza minima dell'immagine (al posto di quella di CHIRP)
	repl     map[string]twinRepl // scelte del programma del costruttore al posto di quelle di CHIRP
}

// twinRepl: le scelte di una voce come le ha il programma del costruttore (valori
// 0, 1, 2… nei bit della maschera)
type twinRepl struct {
	mask byte
	opts []string
}

var (
	twinUVsDrop  = []string{"settings.gpstimezone", "settings.key2short"}
	twinUVsLimit = map[string]int{"settings.beep": 2, "settings.pttdly": 7}
	twinT6       = twinSet{class: "UV17Pro", prog: "«T6UV Series EN CPS»"}
	twinUVs      = twinSet{class: "UV17ProGPS", prog: "«BF-UVs CPS»", drop: twinUVsDrop, limit: twinUVsLimit}
	twinUVsBase  = twinSet{class: "UV17ProGPS", prog: "«BF-UVs CPS»", only: []string{"settings."}, drop: twinUVsDrop, limit: twinUVsLimit}
)

// Radtel RT-400 e RT-330 (4.43): i programmi del costruttore («JJCC-7700 CPS» e «Radio
// CPS») leggono il blocco delle impostazioni a 0x9000 in tre righe di 16 byte
// (GetFunCfgImf_Part1 / 2 / 3) alle stesse posizioni della famiglia MML JC-8810 di
// CHIRP (classe RT470Radio): squelch 0, risparmio 1, VOX 2, luce 3, doppio ascolto 4,
// TOT 5, beep 6, voce 7 e 8, toni DTMF 9, scansione 0x0A, PTT-ID 0x0B e 0x0C, display
// 0x0D e 0x0E; blocco tasti 0x10, allarme 0x11 e 0x12, coda 0x14-0x16, fine
// trasmissione 0x17, radio FM 0x19, VFO / canali 0x1A, tastiera 0x1B, messaggio
// d'accensione 0x1C, tono 0x1E; ritardo del VOX 0x20, uscita dal menu 0x21.
// Differenze trovate confrontando voce per voce e scelta per scelta:
//   - TOT: nei due programmi le scelte sono spento, 30, 60 … 240 s (nove, a passi di
//     30 s); nella tabella di CHIRP sono sei con altri valori: si usano quelle dei
//     programmi;
//   - tasti laterali e superiore (0x2B-0x30): CHIRP usa codici non in fila, i
//     programmi elenchi diversi per modello (nella RT-400 0x2F è il tasto laterale 2
//     premuto a lungo, nella RT-330 il banco corrente; in CHIRP è «RX end tail»): tolti;
//   - RT-400: «QT save» ha l'ordine tutti, TX, RX (CHIRP: tutti, RX, TX);
//   - RT-330: non ha «QT save» né la trasmissione del doppio ascolto; la fine
//     trasmissione ha due scelte; l'identificativo ANI ha 20 valori invece di 60
//     (tolto); la prima scelta della luce è «OFF» (in CHIRP «On»).
//
// Le altre aree (VFO a 0x8000, DTMF a 0xA000, nomi ANI a 0xB000) non sono state
// confrontate e non si mostrano.
var (
	jcSetBlock = [2]int{0x9000, 0x9030}
	jcKeys     = []string{"skey2_sp", "skey2_lp", "skey3_sp", "skey3_lp", "topkey_sp", "rxendtail"}
	jcTOT      = []string{"Off", "30 seconds", "60 seconds", "90 seconds", "120 seconds", "150 seconds", "180 seconds", "210 seconds", "240 seconds"}
)

var twinSettings = map[string]twinSet{
	"bf-uv17r": twinT6, "bf-uv21r": twinT6, "bf-uv21l": twinT6,
	"bf-uv25plus": twinUVs, "bf-uv28pro": twinUVs,
	"bf-uv29pro": twinUVsBase, "bf-uv7b": twinUVsBase, "bf-ar152pro": twinUVsBase,
	"rt-rt400": {class: "RT470Radio", prog: "«JJCC-7700 CPS» v1.0.8", anyClass: true, within: jcSetBlock, size: 0x9040,
		drop: jcKeys, repl: map[string]twinRepl{"tot": {0x0F, jcTOT}, "qtsave": {0x03, []string{"All", "TX", "RX"}}}},
	"rt-rt330": {class: "RT470Radio", prog: "«Radio CPS» v1.1.0 (RT-330)", anyClass: true, within: jcSetBlock, size: 0x9040,
		drop:  append([]string{"ani", "qtsave", "dualtx"}, jcKeys...),
		limit: map[string]int{"roger": 2},
		repl: map[string]twinRepl{"tot": {0x0F, jcTOT},
			"abr": {0x0F, []string{"Off", "5 seconds", "10 seconds", "15 seconds", "20 seconds", "30 seconds", "1 minute", "2 minutes", "3 minutes"}}}},
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
	if tw.size > 0 {
		t.Size = tw.size
	}
	for _, d := range src.Settings {
		if len(tw.only) > 0 && !twinMatch(d.Key, tw.only) {
			continue
		}
		if twinMatch(d.Key, tw.drop) {
			continue
		}
		if tw.within != [2]int{} {
			offs, _, err := d.Enc.fields()
			in := err == nil && len(offs)+len(d.Enc.Offs) > 0
			for _, a := range append(offs, d.Enc.Offs...) {
				in = in && a >= tw.within[0] && a < tw.within[1]
			}
			if !in {
				continue
			}
		}
		if r, ok := tw.repl[d.Key]; ok {
			offs, _, err := d.Enc.fields()
			if err != nil || len(offs) != 1 || d.Type != "list" {
				panic("settings_twin: " + d.Key)
			}
			sh := 0
			for r.mask>>sh&1 == 0 {
				sh++
			}
			e := &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(offs[0]), hex.EncodeToString([]byte{r.mask})}}}
			for i := range r.opts {
				v := hex.EncodeToString([]byte{byte(i<<sh) & r.mask})
				e.Table = append(e.Table, &v)
			}
			d.Enc, d.Options = e, append([]string(nil), r.opts...)
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
			if m == nil || !m.NoChirp || (m.RClass != tw.class && !tw.anyClass) {
				continue // la gemella deve dichiarare la stessa classe di CHIRP
			}
			k := id
			if !tw.anyClass {
				k = tw.class + "|" + strings.Join(tw.only, ",") + "|" + strings.Join(tw.drop, ",") + "|" + tw.prog
			}
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
			if t == nil {
				return nil
			}
			// 4.43: l'immagine letta dal driver è più corta di quella di CHIRP (vedi fitSetTable)
			return fitSetTable(genByID[kind], t, img)
		},
	})
}
