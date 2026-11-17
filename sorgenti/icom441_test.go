//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: Icom IC-91AD / IC-92AD (drv_ic9x.go). CHIRP non ha file per queste
// radio: il confronto si fa sul cavo. Il driver di CHIRP parla con la radio
// simulata e deve leggere gli stessi canali che legge questo programma; le
// stesse modifiche fatte con CHIRP e da qui devono mandare alla radio gli
// stessi blocchi. Poi le regole dei canali, i limiti e gli errori sul cavo.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// immagine vuota per le prove generali
func (ic9xFam) blankImg(m *genModel) []byte { return ic9xBlank(m) }

// lockedSlot: le prove generali modificano solo le memorie normali (limiti di
// scansione e canali di chiamata non hanno il salto di scansione)
func (ic9xFam) lockedSlot(m *genModel, i int) bool { return ic9xSpecial(m, i) }

// liveSession: il driver «live» di CHIRP esegue le operazioni sulla radio
// simulata; dà il risultato di ognuna e i byte che CHIRP ha inviato
func liveSession(t *testing.T, e genEmu, class string, ops []map[string]interface{}) ([]json.RawMessage, []byte) {
	t.Helper()
	chirpDir(t)
	var rec *recPort
	var wg sync.WaitGroup
	wg.Add(1)
	addr, stop := serveTCP(func(p Port) error {
		defer wg.Done()
		rec = &recPort{Port: p}
		return e.serve(rec)
	})
	defer stop()
	oj, _ := json.Marshal(ops)
	of := filepath.Join(t.TempDir(), "ops.json")
	if err := os.WriteFile(of, oj, 0o644); err != nil {
		t.Fatal(err)
	}
	var res []json.RawMessage
	if err := json.Unmarshal(chirpRef(t, "live", class, addr, of), &res); err != nil {
		t.Fatal(err)
	}
	if len(res) != len(ops) {
		t.Fatalf("CHIRP ha eseguito %d operazioni su %d", len(res), len(ops))
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("la radio simulata non ha terminato la sessione con CHIRP")
	}
	return res, rec.got()
}

// ic9xFrames: i blocchi di scrittura e di cancellazione (i dati dopo «FE FE
// banda 80 1A 00») fra i byte inviati alla radio
func ic9xFrames(tx []byte) [][]byte {
	var out [][]byte
	for {
		k := bytes.Index(tx, []byte{0x80, 0x1A, 0x00})
		if k < 0 {
			return out
		}
		tx = tx[k+3:]
		end := bytes.IndexByte(tx, 0xFD)
		if end < 0 {
			return out
		}
		if end > 3 {
			out = append(out, append([]byte{}, tx[:end]...))
		}
		tx = tx[end:]
	}
}

func ic9xBase(t *testing.T, m *genModel) []byte {
	t.Helper()
	f := ic9xFam{}
	img := ic9xBlank(m)
	b := ic9xo(m).band
	chans := map[int]*Chan{
		0:   {Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"},
		1:   {Rx: 145750000, Tx: 145150000, TxTone: "88.5", Scan: true, Name: "R6"},
		2:   {Rx: 430100000, Tx: 431700000, TxTone: "123.0", RxTone: "123.0", Name: "RU 2"},
		7:   {Rx: 433500000, Tx: 433500000, TxTone: "D754N", RxTone: "D754I", Scan: true, Name: "DCS"},
		120: {Rx: 120500000, Tx: 120500000, Scan: true, Name: "TWR"},
		398: {Rx: 439975000, Tx: 439975000, Scan: true, Name: "Last-1"},
	}
	if b == 1 {
		chans[30] = &Chan{Rx: 98100000, Tx: 98100000, Scan: true, Name: "RADIO 1"}
		chans[31] = &Chan{Rx: 999000, Tx: 999000, Scan: true, Name: "OM"}
		chans[799] = &Chan{Rx: 850000000, Tx: 850000000, Name: "800 MHz"}
		chans[800] = &Chan{Rx: 144000000, Tx: 144000000, Scan: true, Name: "0A"}
		chans[801] = &Chan{Rx: 146000000, Tx: 146000000, Scan: true, Name: "0B"}
		chans[850] = &Chan{Rx: 145000000, Tx: 145000000, Scan: true, Name: "CALL V"}
	} else {
		chans[10] = &Chan{Rx: 145612500, Tx: 145012500, Narrow: true, Scan: true, Name: "NFM"}
		chans[20] = &Chan{Rx: 438400000, Tx: 430800000, Dig: true, Scan: true, Name: "DSTAR"}
		chans[400] = &Chan{Rx: 430000000, Tx: 430000000, Scan: true, Name: "0A"}
		chans[451] = &Chan{Rx: 433000000, Tx: 433000000, Scan: true, Name: "CALL U"}
	}
	for i, c := range chans {
		if err := f.Set(m, img, i, c); err != nil {
			t.Fatalf("%s posizione %d: %v", m.ID, i, err)
		}
	}
	if b == 2 {
		// il canale D-STAR con i nominativi di un ripetitore
		p := f.payload(m, img, 20)
		p[35] = 7
		copy(p[36:], "IR0UAA GIR0UAA BCQCQCQ  ")
	}
	return img
}

func TestIC9xModels(t *testing.T) {
	a, b := genByID["ic-9x-a"], genByID["ic-9x-b"]
	if a == nil || b == nil {
		t.Fatal("modelli mancanti")
	}
	f := ic9xFam{}
	for _, c := range []struct {
		m                  *genModel
		slots, normal, mem int
		band               byte
	}{{a, 852, 850, 800, 1}, {b, 452, 450, 400, 2}} {
		o := ic9xo(c.m)
		if c.m.Slots != c.slots || o.normal != c.normal || o.mem != c.mem || o.band != c.band || c.m.First != 0 || c.m.NameMax != 8 ||
			!c.m.NoChirp || !c.m.NoTxOff || !c.m.HasScan || c.m.Baud != 4800 || (c.m.Digital == "D-STAR") != (c.band == 2) || c.m.Narrow != (c.band == 2) {
			t.Errorf("%s: dati del modello", c.m.ID)
		}
		if genMinSize(c.m) > ic9xSize(c.m) || ic9xSize(c.m) != 32+c.slots*64 {
			t.Errorf("%s: dimensione dell'immagine", c.m.ID)
		}
		for _, x := range []struct {
			i    int
			kind byte
			num  int
			spec bool
		}{{0, 1, 0, false}, {c.mem - 1, 1, c.mem - 1, false}, {c.mem, 1, c.mem, true}, {c.normal - 1, 1, c.normal - 1, true},
			{c.normal, 2, 0, true}, {c.normal + 1, 2, 1, true}} {
			kind, num := ic9xAddr(c.m, x.i)
			if kind != x.kind || num != x.num || ic9xSpecial(c.m, x.i) != x.spec {
				t.Errorf("%s posizione %d: tipo %d numero %d", c.m.ID, x.i, kind, num)
			}
		}
		img := ic9xBlank(c.m)
		if !f.Match(c.m, img) || f.Match(c.m, img[:len(img)-1]) {
			t.Errorf("%s: riconoscimento dell'immagine", c.m.ID)
		}
		for i := 0; i < c.m.Slots; i++ {
			if f.Get(c.m, img, i) != nil {
				t.Fatalf("%s: posizione %d non vuota nell'immagine vuota", c.m.ID, i)
			}
		}
	}
	// ogni banda riconosce solo la sua immagine
	if f.Match(a, ic9xBlank(b)) || f.Match(b, ic9xBlank(a)) {
		t.Errorf("immagine di una banda accettata dall'altra")
	}
	// il canale di partenza è valido in tutte e due le bande ed è quello di CHIRP
	// (146,010 MHz, 600 kHz, 88,5 Hz, DCS 023), con il primo ripetitore vuoto
	if len(ic9xTemplate) != 60 || !ic9xValid(ic9xTemplate[:34], 1) || !ic9xValid(ic9xTemplate, 2) || ic9xValid(ic9xTemplate, 1) || ic9xValid(ic9xTemplate[:34], 2) {
		t.Fatalf("canale di partenza non valido")
	}
	if rx, _ := bcdLE(ic9xTemplate[3:8]); rx != 146010000 {
		t.Errorf("canale di partenza: %d Hz", rx)
	}
	if string(ic9xTemplate[36:]) != "                CQCQCQ  " {
		t.Errorf("canale di partenza: nominativi %q", ic9xTemplate[36:])
	}
	if len(ic9xSteps) != 14 || ic9xSteps[0] != 5000 || ic9xSteps[6] != 15000 || ic9xSteps[13] != 200000 {
		t.Errorf("passi di sintonia: %v", ic9xSteps)
	}
	for _, c := range []struct {
		hz        int64
		old, want int
	}{{145500000, -1, 0}, {145500000, 6, 6}, {145512500, -1, 1}, {145512500, 5, 5}, {145512500, 6, 1}, {999000, -1, 3},
		{118008333, -1, 2}, {118016667, -1, 2}, {145501000, -1, -1}, {433000000, 13, 13}, {433005000, 13, 0}, {145500000, 99, 0}} {
		if got := ic9xStepFor(c.hz, c.old); got != c.want {
			t.Errorf("passo per %d Hz (prima %d): %d, atteso %d", c.hz, c.old, got, c.want)
		}
	}
}

func TestIC9xChannels(t *testing.T) {
	f := ic9xFam{}
	for _, id := range []string{"ic-9x-a", "ic-9x-b"} {
		m := genByID[id]
		o := ic9xo(m)
		img := ic9xBlank(m)
		// un canale nuovo, byte per byte: R6 con subtono in trasmissione
		c := &Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", Scan: true, Name: "R6"}
		if err := f.Set(m, img, 12, c); err != nil {
			t.Fatal(err)
		}
		want := append([]byte{}, ic9xTemplate[:ic9xLen(o.band)]...)
		copy(want, []byte{0x01, 0x00, 0x12, 0x00, 0x00, 0x75, 0x45, 0x01, 0x00, 0x00, 0x60, 0x00})
		want[21], want[22] = 0x00, 0x05 // FM, passo da 5 kHz; subtono in trasmissione, scostamento negativo
		copy(want[26:], "R6      ")
		if got := f.payload(m, img, 12); !bytes.Equal(got, want) {
			t.Errorf("%s: canale nuovo\n  %x\nvoluto\n  %x", id, got, want)
		}
		rt := func(i int, c Chan) {
			t.Helper()
			if err := f.Set(m, img, i, &c); err != nil {
				t.Fatalf("%s posizione %d, %+v: %v", id, i, c, err)
			}
			g := f.Get(m, img, i)
			if g == nil {
				t.Fatalf("%s posizione %d vuota dopo la scrittura", id, i)
			}
			if *g != c {
				t.Errorf("%s posizione %d:\n  scritto %+v\n  riletto %+v", id, i, c, *g)
			}
			if p := f.payload(m, img, i); !ic9xValid(p, o.band) {
				t.Errorf("%s posizione %d: canale non valido per la radio", id, i)
			}
		}
		for _, i := range []int{0, o.mem - 1} {
			rt(i, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "Simplex"})
			rt(i, Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", Scan: true, Name: "R6"})
			rt(i, Chan{Rx: 430100000, Tx: 431700000, TxTone: "123.0", RxTone: "123.0", Name: "RU2 [x]"})
			rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D023N", Scan: true, Name: "DCS"})
			rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D754I", RxTone: "D754N", Scan: true, Name: "DCS IN"})
			rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D754N", RxTone: "D754I", Scan: true, Name: "DCS NI"})
			rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D754I", RxTone: "D754I", Scan: true, Name: "DCS II"})
			rt(i, Chan{Rx: 145337500, Tx: 145337500, Scan: true, Name: "12k5"})
			rt(i, Chan{Rx: 120500000, Tx: 120500000, Scan: true, Name: "AM"})
			rt(i, Chan{Rx: 145000000, Tx: 244999999, Scan: true, Name: "+99.9"})
		}
		// limiti di scansione e canali di chiamata: niente salto di scansione
		for _, i := range []int{o.mem, o.normal - 1, o.normal, o.normal + 1} {
			rt(i, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "EDGE"})
			if err := f.Set(m, img, i, &Chan{Rx: 145525000, Tx: 145525000, Name: "EDGE"}); err != nil {
				t.Fatal(err)
			}
			if p := f.payload(m, img, i); p[23]&3 != 0 || !f.Get(m, img, i).Scan {
				t.Errorf("%s posizione %d: salto di scansione su una posizione speciale", id, i)
			}
			kind, num := ic9xAddr(m, i)
			if p := f.payload(m, img, i); p[0] != kind || int(bcdByte(p[1]))*100+int(bcdByte(p[2])) != num {
				t.Errorf("%s posizione %d: tipo e numero %x", id, i, p[:3])
			}
		}
		// modo: segue la frequenza quando cambia, resta quando non cambia
		modeOf := func(i int) byte { return f.payload(m, img, i)[21] >> 4 & 7 }
		set := func(i int, c Chan) {
			t.Helper()
			if err := f.Set(m, img, i, &c); err != nil {
				t.Fatalf("%s posizione %d, %+v: %v", id, i, c, err)
			}
		}
		set(5, Chan{Rx: 145500000, Tx: 145500000, Scan: true})
		if modeOf(5) != 0 {
			t.Errorf("%s: 145,5 MHz modo %d", id, modeOf(5))
		}
		set(5, Chan{Rx: 121500000, Tx: 121500000, Scan: true})
		if modeOf(5) != 3 {
			t.Errorf("%s: banda aerea modo %d", id, modeOf(5))
		}
		f.payload(m, img, 5)[21] &^= 0x70 // FM scelta sulla radio in banda aerea
		set(5, Chan{Rx: 121500000, Tx: 121500000, Scan: true, Name: "FM"})
		if modeOf(5) != 0 {
			t.Errorf("%s: modo scelto sulla radio perso modificando il nome", id)
		}
		set(5, Chan{Rx: 145500000, Tx: 145500000, Scan: true})
		f.payload(m, img, 5)[21] |= 0x30 // AM scelta sulla radio a 145,5 MHz
		set(5, Chan{Rx: 145500000, Tx: 145500000, Name: "AM"})
		if modeOf(5) != 3 {
			t.Errorf("%s: AM scelta sulla radio persa modificando il nome", id)
		}
		set(5, Chan{Rx: 145525000, Tx: 145525000, Scan: true})
		if modeOf(5) != 0 {
			t.Errorf("%s: modo dopo il cambio di frequenza %d", id, modeOf(5))
		}
		if o.band == 1 {
			set(6, Chan{Rx: 98100000, Tx: 98100000, Scan: true})
			set(7, Chan{Rx: 999000, Tx: 999000, Scan: true})
			set(8, Chan{Rx: 7100000, Tx: 7100000, Scan: true})
			if modeOf(6) != 2 || modeOf(7) != 3 || modeOf(8) != 3 || f.payload(m, img, 7)[21]&0x0F != 3 {
				t.Errorf("banda A: modi %d %d %d, passo delle onde medie %d", modeOf(6), modeOf(7), modeOf(8), f.payload(m, img, 7)[21]&0x0F)
			}
			// nella banda A niente D-STAR; la FM stretta diventa FM
			if err := f.Set(m, img, 9, &Chan{Rx: 145500000, Tx: 145500000, Dig: true}); err == nil {
				t.Errorf("banda A: canale D-STAR accettato")
			}
			set(9, Chan{Rx: 145500000, Tx: 145500000, Narrow: true, Scan: true})
			if modeOf(9) != 0 || f.Get(m, img, 9).Narrow {
				t.Errorf("banda A: FM stretta")
			}
		} else {
			rt(6, Chan{Rx: 145612500, Tx: 145012500, Narrow: true, Scan: true, Name: "NFM"})
			set(6, Chan{Rx: 121500000, Tx: 121500000, Narrow: true, Scan: true})
			if modeOf(6) != 1 {
				t.Errorf("banda B: FM stretta in banda aerea, modo %d", modeOf(6))
			}
			// D-STAR: un canale nuovo nasce con CQCQCQ; i nominativi di uno esistente restano
			rt(20, Chan{Rx: 438400000, Tx: 430800000, Dig: true, Scan: true, Name: "DSTAR"})
			p := f.payload(m, img, 20)
			if modeOf(20) != 4 || string(p[36:]) != "                CQCQCQ  " || p[35] != 0 {
				t.Errorf("canale D-STAR nuovo: modo %d, nominativi %q", modeOf(20), p[36:])
			}
			p[35] = 7
			copy(p[36:], "IR0UAA GIR0UAA BCQCQCQ  ")
			rt(20, Chan{Rx: 438412500, Tx: 430812500, Dig: true, Name: "DSTAR 2"})
			if p = f.payload(m, img, 20); string(p[36:]) != "IR0UAA GIR0UAA BCQCQCQ  " || p[35] != 7 {
				t.Errorf("canale D-STAR modificato: nominativi %q, codice %d", p[36:], p[35])
			}
			// i toni di un canale D-STAR non si usano (e non fanno rifiutare il canale)
			set(20, Chan{Rx: 438412500, Tx: 430812500, Dig: true, TxTone: "D023N", RxTone: "88.5", Name: "DSTAR 2"})
			if g := f.Get(m, img, 20); g.TxTone != "" || g.RxTone != "" || !g.Dig {
				t.Errorf("canale D-STAR con toni: %+v", g)
			}
			// da D-STAR ad analogico e ritorno: i nominativi ripartono da CQCQCQ
			set(20, Chan{Rx: 438412500, Tx: 430812500, Scan: true, Name: "FM"})
			set(20, Chan{Rx: 438412500, Tx: 430812500, Dig: true, Scan: true, Name: "DV"})
			if p = f.payload(m, img, 20); string(p[52:]) != "CQCQCQ  " || string(p[36:52]) != "                " {
				t.Errorf("canale tornato D-STAR: nominativi %q", p[36:])
			}
		}
		// il salto «P» scelto sulla radio resta finché il canale è escluso
		set(40, Chan{Rx: 145500000, Tx: 145500000, Name: "SKIP"})
		p := f.payload(m, img, 40)
		if p[23]&3 != 1 {
			t.Errorf("%s: canale escluso, bit %d", id, p[23]&3)
		}
		p[23] = p[23]&^3 | 2
		set(40, Chan{Rx: 145500000, Tx: 145500000, Name: "PSKIP"})
		if p = f.payload(m, img, 40); p[23]&3 != 2 || f.Get(m, img, 40).Scan {
			t.Errorf("%s: salto «P» perso", id)
		}
		set(40, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "PSKIP"})
		if p = f.payload(m, img, 40); p[23]&3 != 0 {
			t.Errorf("%s: canale rimesso in scansione, bit %d", id, p[23]&3)
		}
		// lo scostamento, il banco e il passo di un canale restano
		set(41, Chan{Rx: 145750000, Tx: 144150000, Scan: true})
		p = f.payload(m, img, 41)
		p[24], p[25], p[21] = 'C', 0x42, p[21]&0xF0|8 // banco C posizione 42, passo da 25 kHz
		set(41, Chan{Rx: 145750000, Tx: 145750000, Scan: true, Name: "KEEP"})
		p = f.payload(m, img, 41)
		if offs, _ := bcdLE(p[8:12]); offs != 1600000 || p[24] != 'C' || p[25] != 0x42 || p[21]&0x0F != 8 || p[22]&3 != 0 {
			t.Errorf("%s: scostamento %d, banco %c/%02x, passo %d", id, offs, p[24], p[25], p[21]&0x0F)
		}
		// rifiuti: la memoria non cambia
		for _, c := range []Chan{
			{Rx: 145500000, Tx: 145500000, RxTone: "88.5"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "100.0"},
			{Rx: 145500000, Tx: 145500000, TxTone: "D023N", RxTone: "D754N"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "D023N"},
			{Rx: 145500000, Tx: 145500000, TxTone: "D023N"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.6"},
			{Rx: 145501000, Tx: 145501000},
			{Rx: 145500000, Tx: 245500000},
			{Rx: 345500000, Tx: 245500000},
			{Rx: 0, Tx: 0},
		} {
			before := append([]byte{}, img...)
			c.Scan = true
			if err := f.Set(m, img, 41, &c); err == nil {
				t.Errorf("%s: %+v accettato", id, c)
			}
			if !bytes.Equal(before, img) {
				t.Fatalf("%s: %+v rifiutato ma l'immagine è cambiata", id, c)
			}
		}
		// cancellazione: la posizione torna vuota, le vicine restano
		before := append([]byte{}, img...)
		if err := f.Set(m, img, 41, nil); err != nil {
			t.Fatal(err)
		}
		if f.Get(m, img, 41) != nil || !bytes.Equal(f.rec(img, 41), make([]byte, 64)) {
			t.Errorf("%s: posizione cancellata ma non vuota", id)
		}
		copy(f.rec(before, 41), make([]byte, 64))
		if !bytes.Equal(before, img) {
			t.Errorf("%s: la cancellazione ha toccato altre posizioni", id)
		}
		// canale letto dalla radio con campi non leggibili: vuoto, e si può riscrivere
		set(42, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "BAD"})
		f.payload(m, img, 42)[4] = 0xAB
		if f.Get(m, img, 42) != nil {
			t.Errorf("%s: canale con frequenza non BCD letto come valido", id)
		}
		rt(42, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "GOOD"})
		// nomi: 8 caratteri, maiuscole e minuscole, segni
		for _, name := range []string{"AbCdEfGh", "a/b-c_d~", "[x]{y}", "A"} {
			rt(43, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: name})
		}
		set(43, Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "TroppoLungo"})
		if g := f.Get(m, img, 43); g.Name != "TroppoLu" {
			t.Errorf("%s: nome lungo riletto %q", id, g.Name)
		}
	}
}

// il driver di CHIRP con la radio simulata: stessi canali letti, stessi
// blocchi inviati per le stesse modifiche
func TestIC9xChirp(t *testing.T) {
	chirpDir(t)
	f := ic9xFam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-9x-a", "ic-9x-b"} {
		m := genByID[id]
		o := ic9xo(m)
		class := map[byte]string{1: "ic9x.IC9xRadioA", 2: "ic9x.IC9xRadioB"}[o.band]
		base := ic9xBase(t, m)
		// lettura: le posizioni piene e alcune vuote (CHIRP nella banda B non
		// arriva ai limiti di scansione né ai canali di chiamata, e nella A ai
		// canali di chiamata: errori del suo codice, non della radio)
		var ops []map[string]interface{}
		var nums []int
		top := o.mem
		if o.band == 1 {
			top = o.normal
		}
		for i := 0; i < top-1; i++ {
			if f.Get(m, base, i) != nil || i%97 == 3 {
				nums = append(nums, i)
				ops = append(ops, map[string]interface{}{"op": "get", "n": i})
			}
		}
		res, _ := liveSession(t, f.emu(m, base), class, ops)
		full := 0
		for k, i := range nums {
			var ref *struct {
				chirpChan
				Ur   string `json:"dv_urcall"`
				R1   string `json:"dv_rpt1call"`
				R2   string `json:"dv_rpt2call"`
				Code int    `json:"dv_code"`
			}
			if err := json.Unmarshal(res[k], &ref); err != nil {
				t.Fatalf("%s posizione %d: CHIRP risponde %s", id, i, res[k])
			}
			g := f.Get(m, base, i)
			if (g == nil) != (ref == nil) {
				t.Errorf("%s posizione %d: qui %v, CHIRP %s", id, i, g, res[k])
				continue
			}
			if g == nil {
				continue
			}
			full++
			mode := "FM"
			switch f.payload(m, base, i)[21] >> 4 & 7 {
			case 1:
				mode = "NFM"
			case 2:
				mode = "WFM"
			case 3:
				mode = "AM"
			case 4:
				mode = "DV"
			}
			txT, rxT := g.TxTone, g.RxTone
			if g.Rx != ref.Rx || g.Tx != ref.Tx || txT != ref.TxTone || rxT != ref.RxTone || g.Name != ref.Name || g.Scan != ref.Scan || mode != ref.Mode {
				t.Errorf("%s posizione %d:\n  qui   %+v (%s)\n  CHIRP %s", id, i, *g, mode, res[k])
			}
			if g.Dig && (ref.Ur != "CQCQCQ" || ref.R1 != "IR0UAA B" || ref.R2 != "IR0UAA G" || ref.Code != 7) {
				t.Errorf("%s posizione %d: nominativi letti da CHIRP %s", id, i, res[k])
			}
		}
		if full < 6 {
			t.Fatalf("%s: solo %d canali confrontati in lettura", id, full)
		}
		// modifiche: le stesse con CHIRP e da qui
		type ed struct {
			n    int
			c    *Chan
			mode string
		}
		edits := []ed{
			{0, &Chan{Rx: 145525000, Tx: 145525000, Scan: true, Name: "S21"}, "FM"},                                   // frequenza e nome
			{1, &Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", RxTone: "88.5", Scan: true, Name: "R6"}, "FM"},    // subtono nei due versi
			{2, &Chan{Rx: 430100000, Tx: 431700000, Scan: true, Name: "RU 2"}, "FM"},                                  // senza tono, in scansione
			{7, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023I", RxTone: "D023N", Scan: true, Name: "DCS"}, "FM"}, // altro codice e polarità
			{50, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "71.9", Scan: true, Name: "NUOVO"}, "FM"},                // canale nuovo
			{51, &Chan{Rx: 432500000, Tx: 432500000, TxTone: "D754N", RxTone: "D754N", Name: "new-2"}, "FM"},          // nuovo, escluso
			{52, &Chan{Rx: 145800000, Tx: 145800000, Scan: true}, "FM"},                                               // nuovo senza nome
			{120, nil, ""}, // cancellazione
			{398, nil, ""},
		}
		if o.band == 1 {
			edits = append(edits,
				ed{60, &Chan{Rx: 99900000, Tx: 99900000, Scan: true, Name: "FM 99.9"}, "WFM"},
				ed{61, &Chan{Rx: 119700000, Tx: 119700000, Scan: true, Name: "APP"}, "AM"},
				ed{800, &Chan{Rx: 144500000, Tx: 144500000, Scan: true, Name: "0A"}, "FM"},
				ed{848, &Chan{Rx: 430000000, Tx: 430000000, Scan: true, Name: "24A"}, "FM"})
		} else {
			edits = append(edits,
				ed{60, &Chan{Rx: 145600000, Tx: 145000000, Narrow: true, Scan: true, Name: "NFM"}, "NFM"},
				ed{20, &Chan{Rx: 438450000, Tx: 430850000, Dig: true, Name: "DV edit"}, "DV"})
		}
		ops = nil
		mod := append([]byte{}, base...)
		for _, e := range edits {
			if e.c == nil {
				ops = append(ops, map[string]interface{}{"op": "del", "n": e.n})
			} else {
				ops = append(ops, map[string]interface{}{"op": "set", "n": e.n, "chan": map[string]interface{}{
					"rx": e.c.Rx, "tx": e.c.Tx, "txOff": false, "txTone": e.c.TxTone, "rxTone": e.c.RxTone, "power": -1,
					"mode": e.mode, "scan": e.c.Scan, "name": e.c.Name}})
			}
			if err := f.Set(m, mod, e.n, e.c); err != nil {
				t.Fatalf("%s posizione %d: %v", id, e.n, err)
			}
		}
		ce := f.emu(m, base).(*ic9xEmu)
		res, chirpTx := liveSession(t, ce, class, ops)
		for k, r := range res {
			if string(r) != `"ok"` {
				t.Errorf("%s: CHIRP sulla posizione %d: %s", id, edits[k].n, r)
			}
		}
		ge := f.emu(m, base).(*ic9xEmu)
		err, gtx := alSession(t, ge, nil, func() error {
			_, err := f.Upload(ctx, m, "EMU", mod, nop)
			return err
		})
		if err != nil {
			t.Fatalf("%s: scrittura: %v", id, err)
		}
		// i blocchi inviati: CHIRP li manda nell'ordine delle modifiche, qui in
		// ordine di posizione. Nei canali non D-STAR della banda B CHIRP manda i
		// nominativi del suo esempio, qui restano quelli della radio (o vuoti)
		norm := func(frames [][]byte) map[string]bool {
			out := map[string]bool{}
			for _, fr := range frames {
				fr = append([]byte{}, fr...)
				if len(fr) == 60 && fr[21]>>4&7 != 4 {
					copy(fr[34:], make([]byte, 26))
				}
				out[string(fr)] = true
			}
			return out
		}
		cf, gf := norm(ic9xFrames(chirpTx)), norm(ic9xFrames(gtx))
		if len(cf) != len(edits) || len(gf) != len(edits) {
			t.Errorf("%s: blocchi inviati: CHIRP %d, qui %d, modifiche %d", id, len(cf), len(gf), len(edits))
		}
		for fr := range cf {
			if !gf[fr] {
				t.Errorf("%s: blocco inviato solo da CHIRP: %x", id, fr)
			}
		}
		for fr := range gf {
			if !cf[fr] {
				t.Errorf("%s: blocco inviato solo da qui:   %x", id, fr)
			}
		}
		// e la radio alla fine ha gli stessi canali (letti da qui)
		ci, gi := ce.image(), ge.image()
		for i := 0; i < m.Slots; i++ {
			a, b := f.Get(m, ci, i), f.Get(m, gi, i)
			if (a == nil) != (b == nil) || (a != nil && *a != *b) {
				t.Errorf("%s posizione %d: dopo CHIRP %+v, dopo questo programma %+v", id, i, a, b)
			}
		}
		if !bytes.Equal(gi, mod) {
			t.Errorf("%s: la radio non ha l'immagine scritta", id)
		}
		if ce.writes != ge.writes || ce.clears != ge.clears || ge.clears != 2 {
			t.Errorf("%s: scritture %d/%d, cancellazioni %d/%d", id, ce.writes, ge.writes, ce.clears, ge.clears)
		}
		t.Logf("%s: %d canali letti come CHIRP; %d modifiche: %d blocchi inviati, uguali a quelli di CHIRP", id, full, len(edits), len(gf))
	}
}

// ic9xPort: cavo che ricorda a che velocità è stato inviato ogni blocco
type ic9xPort struct {
	*pipeEnd
	mu    sync.Mutex
	baud  int
	sends []int
}

func (p *ic9xPort) SetBaud(b int) error {
	p.mu.Lock()
	p.baud = b
	p.mu.Unlock()
	return p.pipeEnd.SetBaud(b)
}

func (p *ic9xPort) Write(b []byte) error {
	p.mu.Lock()
	p.sends = append(p.sends, p.baud)
	p.mu.Unlock()
	return p.pipeEnd.Write(b)
}

func TestIC9xCable(t *testing.T) {
	old := ic9xStale
	defer func() { ic9xStale = old }()
	f := ic9xFam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"ic-9x-a", "ic-9x-b"} {
		m := genByID[id]
		base := ic9xBase(t, m)
		mk := func() *ic9xEmu { return f.emu(m, base).(*ic9xEmu) }
		read := func(e *ic9xEmu, wrap func(p *pipeEnd) Port) ([]byte, error, []byte) {
			var img []byte
			err, tx := alSession(t, e, wrap, func() error {
				var err error
				img, _, err = f.Download(ctx, m, "EMU", nop)
				return err
			})
			return img, err, tx
		}
		// lettura: immagine uguale alla radio; il risveglio a 4800 baud, il resto a 38400
		var port *ic9xPort
		e := mk()
		img, err, tx := read(e, func(p *pipeEnd) Port {
			port = &ic9xPort{pipeEnd: p, baud: p.baud}
			return port
		})
		if err != nil || !bytes.Equal(img, base) || !bytes.Equal(img, e.image()) {
			t.Fatalf("%s: lettura: %v", id, err)
		}
		if e.wakes != 1 || e.reads != m.Slots || e.writes+e.clears != 0 {
			t.Errorf("%s: lettura: %d risvegli, %d letture", id, e.wakes, e.reads)
		}
		if len(port.sends) != 1+m.Slots || port.sends[0] != 4800 {
			t.Errorf("%s: %d invii, il primo a %d baud", id, len(port.sends), port.sends[0])
		}
		for _, b := range port.sends[1:] {
			if b != 38400 {
				t.Fatalf("%s: comando inviato a %d baud", id, b)
			}
		}
		wake := append(bytes.Repeat([]byte{0xFE}, 162), 0x01, 0x80, 0x19, 0xFD)
		if !bytes.HasPrefix(tx, wake) || !bytes.HasPrefix(tx[len(wake):], []byte{0xFE, 0xFE, ic9xo(m).band, 0x80, 0x1A, 0x00, 0x01, 0x00, 0x00, 0xFD}) {
			t.Errorf("%s: inizio della sessione %x", id, tx[:min(len(tx), 180)])
		}
		// cavo con l'eco e blocchi non richiesti dalla radio: stessa immagine
		e = mk()
		e.echo, e.stray = true, true
		if img, err, _ = read(e, nil); err != nil || !bytes.Equal(img, base) {
			t.Errorf("%s: lettura con eco e blocchi non richiesti: %v", id, err)
		}
		// radio spenta: cinque tentativi come CHIRP, poi l'errore
		e = mk()
		e.deaf = true
		img, err, tx = read(e, nil)
		if err == nil || img != nil || !strings.Contains(err.Error(), "non risponde") || bytes.Count(tx, []byte{0x01, 0x80, 0x19, 0xFD}) != 5 || len(tx) != 5*len(wake) {
			t.Errorf("%s: radio spenta: %v, %d byte inviati", id, err, len(tx))
		}
		// la radio smette di rispondere a metà, o risponde con un'altra posizione
		e = mk()
		e.mute = 100
		if img, err, _ = read(e, nil); err == nil || img != nil || !strings.Contains(err.Error(), "non risponde") {
			t.Errorf("%s: radio muta a metà: %v", id, err)
		}
		e = mk()
		e.wrong = true
		if img, err, _ = read(e, nil); err == nil || img != nil || !strings.Contains(err.Error(), "altra posizione") {
			t.Errorf("%s: posizione sbagliata: %v", id, err)
		}
		// scrittura: solo le posizioni cambiate
		mod := append([]byte{}, base...)
		for i, c := range map[int]*Chan{
			0:               {Rx: 145525000, Tx: 145525000, Scan: true, Name: "S21"},
			1:               nil,
			300:             {Rx: 145787500, Tx: 145187500, TxTone: "82.5", Scan: true, Name: "R7a"},
			ic9xo(m).normal: {Rx: 145500000, Tx: 145500000, Scan: true, Name: "CALL"},
		} {
			if err := f.Set(m, mod, i, c); err != nil {
				t.Fatal(err)
			}
		}
		write := func(e *ic9xEmu, img []byte) (error, []byte) {
			return alSession(t, e, nil, func() error {
				_, err := f.Upload(ctx, m, "EMU", img, nop)
				return err
			})
		}
		for _, echo := range []bool{false, true} {
			e = mk()
			e.echo, e.stray = echo, echo
			if err, _ = write(e, mod); err != nil {
				t.Fatalf("%s: scrittura (eco %v): %v", id, echo, err)
			}
			if !bytes.Equal(e.image(), mod) || e.writes != 3 || e.clears != 1 || e.reads != m.Slots {
				t.Errorf("%s: scrittura (eco %v): %d canali scritti, %d cancellati, %d letture", id, echo, e.writes, e.clears, e.reads)
			}
		}
		// niente da cambiare: niente scritto
		e = mk()
		if err, tx = write(e, base); err != nil || e.writes+e.clears != 0 || len(ic9xFrames(tx)) != 0 {
			t.Errorf("%s: scrittura senza modifiche: %v, %d scritture", id, err, e.writes+e.clears)
		}
		// la radio rifiuta un canale: errore con la posizione
		e = mk()
		e.nakAt = 2
		if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "rifiutato la posizione 300") || e.writes != 1 {
			t.Errorf("%s: canale rifiutato: %v (%d scritti)", id, err, e.writes)
		}
		// immagine dell'altra banda o rovinata: niente inviato
		other := genByID[map[string]string{"ic-9x-a": "ic-9x-b", "ic-9x-b": "ic-9x-a"}[id]]
		if err, tx = write(mk(), ic9xBlank(other)); err == nil || len(tx) != 0 {
			t.Errorf("%s: immagine dell'altra banda: %v, %d byte inviati", id, err, len(tx))
		}
		bad := append([]byte{}, mod...)
		f.rec(bad, 300)[2+21] = 0x7F
		if err, tx = write(mk(), bad); err == nil || len(tx) != 0 || !strings.Contains(err.Error(), "300") {
			t.Errorf("%s: immagine rovinata: %v, %d byte inviati", id, err, len(tx))
		}
		// radio spenta: scrittura mai cominciata
		e = mk()
		e.deaf = true
		if err, _ = write(e, mod); err == nil || e.writes+e.clears+e.reads != 0 {
			t.Errorf("%s: scrittura con la radio spenta: %v", id, err)
		}
	}
	// la radio che si riaddormenta: il risveglio si ripete (come CHIRP dopo 4
	// secondi senza comandi; qui i tempi sono accorciati)
	m := genByID["ic-9x-b"]
	base := ic9xBase(t, m)
	e := f.emu(m, base).(*ic9xEmu)
	ic9xStale = 150 * time.Millisecond
	var got, again []byte
	err, _ := alSession(t, e, nil, func() error {
		s, err := ic9xOpen(ctx, m, "EMU")
		if err != nil {
			return err
		}
		defer s.p.Close()
		if got, err = s.get(m, 0); err != nil {
			return err
		}
		e.mu.Lock()
		e.awake = false // la radio è tornata al funzionamento normale
		e.mu.Unlock()
		time.Sleep(200 * time.Millisecond)
		again, err = s.get(m, 1)
		return err
	})
	if err != nil || e.wakes != 2 || !bytes.Equal(got, f.payload(m, base, 0)) || !bytes.Equal(again, f.payload(m, base, 1)) {
		t.Errorf("radio riaddormentata: %v, %d risvegli", err, e.wakes)
	}
}
