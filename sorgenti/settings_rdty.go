// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni generali di Baofeng RD-5R e DM-1801 (4.10), famiglia Radioddity
// GD-77: blocco di 0x28 byte a 0x00E0 e testi di accensione a 0x7518-0x755F.
// Posizioni e codifiche riscritte da zero dalla descrizione di qdmr
// (radioddity_codeplug.cc/.hh, GeneralSettingsElement e BootSettingsElement,
// GNU GPL v3, Hannes Matuschek DM3MAT) e di dmrconfig (rd5r.c, dm1801.c: nome,
// ID e testi di accensione, licenza BSD, Serge Vakulenko KK6ABQ). Le password
// (programmazione e accensione) non si toccano: un errore bloccherebbe la radio.

import (
	"encoding/hex"
	"strconv"
)

const rdtySetBase = 0x00E0

func rdtyTbl(off int, mask byte, vals []byte) *setEnc {
	t := make([]*string, len(vals))
	for i, v := range vals {
		s := hex.EncodeToString([]byte{v & mask})
		t[i] = &s
	}
	return &setEnc{Kind: "table", Fields: [][2]interface{}{{float64(off), hex.EncodeToString([]byte{mask})}}, Table: t}
}

func rdtyList(key, group, label, desc string, off int, mask byte, vals []byte, opts []string) setDef {
	codes := make([]string, len(opts))
	for i := range opts {
		codes[i] = strconv.Itoa(i)
	}
	return setDef{Key: key, Group: group, Label: label, Type: "list", Options: codes, ItOpts: opts, ItLabel: label, ItDesc: desc, Enc: rdtyTbl(off, mask, vals)}
}

func rdtyBool(key, group, label, desc string, off int, bit uint, inv bool) setDef {
	vals := []byte{0, 1 << bit}
	if inv {
		vals = []byte{1 << bit, 0}
	}
	return setDef{Key: key, Group: group, Label: label, Type: "bool", ItLabel: label, ItDesc: desc, Enc: rdtyTbl(off, 1<<bit, vals)}
}

// rdtyNum: byte che vale «unit» per passo (es. 60 ms)
func rdtyNum(key, group, label, desc string, off int, unit, max int) setDef {
	return setDef{Key: key, Group: group, Label: label, Type: "int", Min: 0, Max: float64(max * unit), Step: float64(unit),
		ItLabel: label, ItDesc: desc,
		Enc: &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(off), "ff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: int64(unit)}}
}

func rdtyStr(key, group, label, desc string, off, n int) setDef {
	m := map[string]int{}
	for _, c := range asciiPrintable {
		m[string(c)] = int(c)
	}
	offs := make([]int, n)
	for i := range offs {
		offs[i] = off + i
	}
	p := 0xFF
	return setDef{Key: key, Group: group, Label: label, Type: "str", MaxLen: n, Charset: asciiPrintable, ItLabel: label, ItDesc: desc,
		Enc: &setEnc{Kind: "str", Offs: offs, Map: m, Pad: &p}}
}

var rdtySettings = func() *setTable {
	g1, g2, g3, g4 := "Radio e identificativo DMR", "Chiamate e tempi", "Toni e segnalazioni", "Scansione, risparmio e display"
	b := rdtySetBase
	t := &setTable{Class: "RDTY", Size: rdtyMemSize, Source: "descrizione della memoria di qdmr e dmrconfig (impostazioni generali della famiglia Radioddity GD-77)"}
	t.Settings = []setDef{
		rdtyStr("radio_name", g1, "Nome della radio", "Nome mostrato dalla radio (di solito il nominativo), fino a 8 caratteri.", b, 8),
		{Key: "radio_id", Group: g1, Label: "ID DMR della radio", Type: "int", Min: 0, Max: 16776415, ItLabel: "ID DMR della radio",
			ItDesc: "Numero DMR personale (ID radio), assegnato da RadioID.net al nominativo. Le altre stazioni e i ponti lo vedono come mittente.",
			Enc:    &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(b + 8), "ffffffff"}}, Endian: "be", Fmt: "bcd", Num: 1, Den: 1}},
		rdtyStr("intro1", g1, "Testo di accensione, riga 1", "Prima riga del messaggio mostrato all'accensione (16 caratteri).", 0x7540, 16),
		rdtyStr("intro2", g1, "Testo di accensione, riga 2", "Seconda riga del messaggio mostrato all'accensione (16 caratteri).", 0x7550, 16),
		rdtyList("boot_disp", g1, "Schermata di accensione", "Che cosa mostra la radio all'accensione: il logo del costruttore o le due righe di testo.", 0x7518, 0xFF, []byte{0, 1}, []string{"Logo", "Testo (le due righe)"}),
		rdtyNum("preamble", g2, "Durata del preambolo DMR (ms)", "Durata del segnale che apre la trasmissione digitale, a passi di 60 ms (predefinito 360 ms).", b+0x11, 60, 255),
		rdtyList("monitor", g2, "Tipo di monitor", "Tasto monitor: apre lo squelch (si sente il fruscio) oppure ascolto silenzioso.", b+0x12, 0xFF, []byte{0, 1}, []string{"Apre lo squelch", "Silenzioso"}),
		rdtyList("vox", g2, "Sensibilità del VOX", "Livello di voce che fa partire la trasmissione con il VOX (1 = più sensibile); «Spento» lo disattiva.", b+0x13, 0xFF,
			seqBytes(11, 0, 1, 0), append([]string{"Spento"}, strNums(10, 1, 1, "")...)),
		rdtyNum("lowbat", g2, "Avviso di batteria scarica (s)", "Ogni quanti secondi la radio ripete l'avviso di batteria scarica, a passi di 5 s.", b+0x14, 5, 127),
		rdtyNum("call_alert", g2, "Durata dell'avviso di chiamata (s)", "Per quanto tempo suona l'avviso di chiamata, a passi di 5 s (0 = continuo).", b+0x15, 5, 240),
		rdtyNum("grp_hang", g2, "Attesa dopo una chiamata di gruppo (ms)", "Tempo in cui si può rispondere nello stesso gruppo dopo la fine di una chiamata, a passi di 500 ms.", b+0x18, 500, 14),
		rdtyNum("priv_hang", g2, "Attesa dopo una chiamata privata (ms)", "Tempo in cui si può rispondere alla stessa stazione dopo una chiamata privata, a passi di 500 ms.", b+0x19, 500, 14),
		rdtyBool("ch_down_vfo", g4, "Tasto giù: modalità VFO", "Nel modo canali il tasto giù passa alla frequenza libera (VFO).", b+0x1A, 0, false),
		rdtyBool("ch_up_vfo", g4, "Tasto su: modalità VFO", "Nel modo canali il tasto su passa alla frequenza libera (VFO).", b+0x1A, 1, false),
		rdtyBool("reset_tone", g3, "Tono di ripristino", "Tono all'arrivo di un comando di ripristino.", b+0x1A, 2, false),
		rdtyBool("unknown_tone", g3, "Tono per numero sconosciuto", "Tono quando chiama una stazione che non è nella rubrica.", b+0x1A, 3, false),
		rdtyList("arts", g3, "Tono ARTS", "Segnalazione acustica del controllo automatico del collegamento (ARTS).", b+0x1A, 0xF0, []byte{0, 4 << 4, 8 << 4}, []string{"Spento", "Una volta", "Sempre"}),
		rdtyBool("dig_permit", g3, "Tono di permesso (digitale)", "Tono che indica che si può parlare dopo aver premuto il PTT in DMR.", b+0x1B, 0, false),
		rdtyBool("ana_permit", g3, "Tono di permesso (analogico)", "Tono che indica che si può parlare dopo aver premuto il PTT in FM.", b+0x1B, 1, false),
		rdtyBool("selftest", g3, "Tono di autodiagnosi", "Tono all'accensione, dopo il controllo della radio.", b+0x1B, 2, false),
		rdtyBool("free_tone", g3, "Tono di canale libero", "Tono quando il canale torna libero.", b+0x1B, 3, false),
		rdtyBool("all_tones", g3, "Tutti i toni", "Spento: la radio non emette nessun tono.", b+0x1B, 5, true),
		rdtyBool("batsave_rx", g4, "Risparmio energetico in ricezione", "In ascolto la radio spegne il ricevitore a intervalli per consumare meno.", b+0x1B, 6, false),
		rdtyBool("batsave_pre", g4, "Risparmio con il preambolo", "Usa il preambolo per svegliare le radio in risparmio energetico.", b+0x1B, 7, false),
		rdtyBool("all_leds", g4, "Indicatori luminosi (LED)", "Spento: la radio non accende nessun LED.", b+0x1C, 5, true),
		rdtyBool("tx_exit", g3, "Tono di fine trasmissione", "Tono quando si rilascia il PTT.", b+0x1D, 3, false),
		rdtyBool("tx_active", g4, "Trasmetti sul canale attivo", "Durante la scansione o il doppio ascolto trasmette sul canale su cui la radio si è fermata.", b+0x1D, 4, false),
		rdtyBool("animation", g4, "Animazione all'accensione", "Mostra l'animazione all'accensione.", b+0x1D, 5, false),
		rdtyList("scan_mode", g4, "Modo di scansione", "Quando la scansione riprende: dopo un tempo, alla fine del segnale (portante) o si ferma (ricerca).", b+0x1D, 0xC0, []byte{0, 1 << 6, 2 << 6}, []string{"A tempo", "A portante", "Si ferma"}),
		rdtyList("rpt_delay", g2, "Ritardo di fine ripetitore", "Attesa dopo la fine della trasmissione di un ponte (0-15).", b+0x1E, 0x0F, seqBytes(16, 0, 1, 0), strNums(16, 0, 1, "")),
		rdtyList("rpt_ste", g2, "Eliminazione della coda del ponte (STE)", "Eliminazione del rumore alla fine della trasmissione di un ponte (0-15).", b+0x1E, 0xF0, seqBytes(16, 0, 1, 4), strNums(16, 0, 1, "")),
	}
	return t
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "bf-rd5r" || kind == "bf-dm1801" },
		get: func(kind string, img []byte) *setTable {
			if len(img) < rdtySettings.Size {
				return nil
			}
			return rdtySettings
		},
	})
}

// ---------------------------------------------------------------- DM-1701

// Impostazioni generali della Baofeng DM-1701 / Retevis RT84 (famiglia TYT
// MD-UV380): blocco a 0x2040 dell'immagine. Solo i campi su cui concordano
// dmrconfig (uv380.c) e qdmr (tyt_codeplug.cc): nome della radio (16 caratteri
// Unicode a 0x20B0), ID DMR (24 bit a 0x2084) e righe del testo di accensione
// (10 caratteri Unicode a 0x2040 e 0x2054); per gli altri bit i due programmi
// non concordano e non si toccano.

var zeroPad = 0 // il resto del testo a 0x0000

func tytUStr(key, group, label, desc string, off, n int) setDef {
	offs := make([]int, n)
	for i := range offs {
		offs[i] = off + 2*i
	}
	return setDef{Key: key, Group: group, Label: label, Type: "str", MaxLen: n, Charset: asciiPrintable + "àèéìòùÀÈÉÌÒÙçÇñÑäöüÄÖÜß°", ItLabel: label, ItDesc: desc,
		Enc: &setEnc{Kind: "str", Offs: offs, Fmt: "utf16le", Pad: &zeroPad}}
}

var tytSettings = func() *setTable {
	g := "Radio e identificativo DMR"
	t := &setTable{Class: "DM1701", Size: tytMemSize, Source: "descrizione della memoria di dmrconfig e qdmr (impostazioni generali della famiglia TYT MD-UV380)"}
	t.Settings = []setDef{
		tytUStr("radio_name", g, "Nome della radio", "Nome mostrato dalla radio (di solito il nominativo), fino a 16 caratteri.", 0x20B0, 16),
		{Key: "radio_id", Group: g, Label: "ID DMR della radio", Type: "int", Min: 1, Max: 16776415, ItLabel: "ID DMR della radio",
			ItDesc: "Numero DMR personale (ID radio), assegnato da RadioID.net al nominativo.",
			Enc:    &setEnc{Kind: "num", Fields: [][2]interface{}{{float64(0x2084), "ffffff"}}, Endian: "le", Fmt: "bin", Num: 1, Den: 1}},
		tytUStr("intro1", g, "Testo di accensione, riga 1", "Prima riga del messaggio mostrato all'accensione (10 caratteri).", 0x2040, 10),
		tytUStr("intro2", g, "Testo di accensione, riga 2", "Seconda riga del messaggio mostrato all'accensione (10 caratteri).", 0x2054, 10),
	}
	return t
}()

func init() {
	settingsProviders = append(settingsProviders, settingsProvider{
		has: func(kind string) bool { return kind == "bf-dm1701" },
		get: func(kind string, img []byte) *setTable {
			if len(img) < tytSettings.Size {
				return nil
			}
			return tytSettings
		},
	})
}
