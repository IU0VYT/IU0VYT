// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni delle Baofeng della serie «D» (BF-V1D, 888SD, 1901D, 889D, V2D…V12D,
// 1904D, 1909D, W31D, BF-T20D), 4.42. Ricavate dal programma del costruttore
// «BF Series CPS» v1.5.1 (pacchetti «Baofeng_V1D_CPS» e «Baofeng BF-V12D Software»)
// e controllate sulla versione precedente «BF_FH_user» 1.3.19 (pacchetto
// «Baofeng_BF-D_Series_CPS»): è un programma Electron, il codice è JavaScript.
//
// La funzione che prepara la scrittura mette, dopo i canali:
//
//   0x2B0  guida vocale (0/1)              0x2B8  «RXEmergency» (0/1; non nel tipo E)
//   0x2B1  lingua della guida vocale       0x2B9  riduzione del rumore (0/1)
//   0x2B2  scansione (0/1; casella nascosta)
//   0x2B3  VOX (0/1)                       0x3C0  bit 0 suono dei tasti («Beep»)
//   0x2B4  livello del VOX (1-5)                  bit 1 risparmio della batteria
//   0x2B5  VOX spento in ricezione                bit 2 modalità dell'allarme (solo tipo C)
//          (casella assente)               0x3C1  squelch: 0, 1, 5 o 9
//   0x2B6  blocco TX a batteria scarica    0x3C2  tasto laterale (non nei tipi C ed E)
//   0x2B7  blocco TX a tensione alta       0x3C3  tempo massimo di trasmissione (0-10)
//
// È la mappa della BF-888S (e dei programmi «AnaCps», vedi settings_t17.go), con in
// più la riduzione del rumore; le due voci del blocco della trasmissione sono
// nell'ordine di CHIRP per la BF-888S (bassa in 0x2B6, alta in 0x2B7).
//
// Scelte: squelch «Normally Open» 0, «Convention» 1, «Strengthen» 5, «Strict» 9;
// TOT spento, 30…300 s; tasto laterale spento 0, monitor 1, alta/bassa potenza 2,
// allarme 3 (la scelta 2 manca nei tipi B e D); allarme locale 0 / remoto 1. La
// lingua ha due scelte i cui nomi dipendono dalla coppia di lingue della radio
// (inglese e cinese, cinese e russo, arabo e inglese…): qui «prima» e «seconda».
// Tipi (protocolCode del programma): A = V1D, 888SD, 1904D, V12D, V9D, V10D, V2D,
// V5D, V6D; B = 889D, V4D, 1909D, V8D; C = 1901D; D = W31D, T20.
//
// ATTENZIONE, incongruenza del programma del costruttore (uguale nelle due
// versioni): in scrittura la guida vocale va in 0x2B0 e il suono dei tasti nel bit 0
// di 0x3C0, ma la funzione che interpreta la lettura li assegna al contrario
// (0x2B0 al suono dei tasti, bit 0 di 0x3C0 alla guida vocale). Qui si segue la
// scrittura, che coincide con CHIRP per la BF-888S e con tutti i programmi «AnaCps»
// della stessa famiglia; le due voci hanno un avviso che lo dice.
//
// Non mostrate: scansione (nel programma la casella è nascosta), «VOX spento in
// ricezione» (nessuna casella), le voci del tipo E (UV8D, non gestita).

const wBfdSwap = "Nel programma del costruttore «Guida vocale» e «Suono dei tasti» risultano scambiate fra la lettura e la scrittura: " +
	"qui si segue la scrittura, come nelle altre radio della stessa famiglia. Se sulla radio l'effetto è scambiato, usa l'altra voce."

func bfdSetDefs(proto byte) []t17SetDef {
	g1, g2, g3 := "Ricezione e trasmissione", "Display, suoni e tasti", "DTMF, allarme e funzioni remote"
	t := t17SetDefsFor(0)
	pick := func(key string, off int, mask byte) t17SetDef {
		for _, d := range t {
			if d.key == key {
				d.off, d.mask = off, mask
				return d
			}
		}
		panic(key)
	}
	sql := pick("sql", 0x3C1, 0xFF)
	sql.opts, sql.vals = []string{"Sempre aperto", "Normale", "Rinforzato", "Rigoroso"}, []int{0, 1, 5, 9}
	sql.dsc = "Soglia di apertura dell'audio. Con «Sempre aperto» l'audio resta sempre aperto (si sente il fruscio); negli altri livelli serve un segnale via via più forte."
	vl := pick("vox_level", 0x2B4, 0xFF)
	vl.first = 1
	voice, beep := pick("voice", 0x2B0, 0xFF), pick("beep", 0x3C0, 0x01)
	voice.warn, beep.warn = wBfdSwap, wBfdSwap
	lang := pick("lang", 0x2B1, 0xFF)
	lang.opts = []string{"Prima lingua della radio", "Seconda lingua della radio"}
	lang.dsc = "Lingua della guida vocale. Ogni radio ha una coppia di lingue (di solito inglese e cinese): quale sia la prima dipende dalla versione della radio."
	d := []t17SetDef{
		sql, pick("tot", 0x3C3, 0xFF), pick("vox_on", 0x2B3, 0xFF), vl,
		pick("low_tx", 0x2B6, 0xFF), pick("high_tx", 0x2B7, 0xFF),
		{key: "noise", off: 0x2B9, mask: 0xFF, group: g1, label: "Riduzione del rumore",
			dsc: "Filtro che attenua il rumore di fondo dell'audio."},
		voice, lang, beep, pick("save", 0x3C0, 0x02),
	}
	switch proto {
	case 'A':
		d = append(d, t17SetDef{key: "side_key", off: 0x3C2, mask: 0xFF, opts: []string{"Nessuna", "Monitor", "Cambio di potenza", "Allarme"}, group: g2,
			label: "Tasto laterale", dsc: "Funzione del tasto laterale."})
	case 'B', 'D':
		d = append(d, t17SetDef{key: "side_key", off: 0x3C2, mask: 0xFF, opts: []string{"Nessuna", "Monitor", "Allarme"}, vals: []int{0, 1, 3}, group: g2,
			label: "Tasto laterale", dsc: "Funzione del tasto laterale."})
	}
	d = append(d, t17SetDef{key: "rx_emerg", off: 0x2B8, mask: 0xFF, group: g3, label: "Allarme («Rx Emergency»)",
		dsc: "Accende o spegne la funzione di allarme: nel programma del costruttore la voce si chiama «Rx Emergency» in inglese e «interruttore dell'allarme» in cinese."})
	if proto == 'C' {
		am := pick("alarm_mode", 0x3C0, 0x04)
		d = append(d, am)
	}
	return d
}

var bfdSettings = map[string]*setTable{}

func init() {
	src := t17Src("«BF Series CPS» v1.5.1")
	byProto := map[byte]*setTable{}
	for _, m := range genModels {
		if _, ok := m.Fam.(bfdFam); !ok {
			continue
		}
		var proto byte
		for _, v := range bfdModels {
			if v.key == bfdo(m).key {
				proto = v.proto
			}
		}
		if proto != 'A' && proto != 'B' && proto != 'C' && proto != 'D' {
			continue
		}
		if byProto[proto] == nil {
			byProto[proto] = t17SetTableSize(bfdSetDefs(proto), src, bfdMemsize)
		}
		bfdSettings[m.ID] = byProto[proto]
	}
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return bfdSettings[kind] != nil },
		get: func(kind string, img []byte) *setTable {
			t := bfdSettings[kind]
			if t == nil || len(img) < t.Size {
				return nil
			}
			return t
		},
	})
}
