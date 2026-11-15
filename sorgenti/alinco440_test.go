//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.40: Alinco (drv_alinco.go). Lettura, modifica e copia sono confrontate con
// CHIRP nelle prove TestChirp*; qui i casi che quelle non toccano: i limiti dei
// canali, le differenze volute, i cavi con e senza eco, gli errori sul cavo, la
// velocità della DJ-G7 e il posto delle Alinco nel giro di «Identifica».

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

var alincoIDs = []string{"al-dj596", "al-dj175", "al-dr135t", "al-dr235t", "al-dr435t", "al-dr03t", "al-dr06t"}

// memoria vuota di una Alinco della prima famiglia (zeri, con i byte del modello)
func alincoBlank(m *genModel) []byte {
	o := alincoO(m)
	img := make([]byte, o.size)
	copy(img[0x64:], o.sig)
	return img
}

func TestAlincoModels(t *testing.T) {
	if alincoGapStd != 100*time.Millisecond {
		t.Fatalf("pausa prima di ogni blocco: %v, quella di CHIRP è 100 ms", alincoGapStd)
	}
	if len(alincoTones) != 39 || alincoTones[alincoToneDef] != 885 || alincoTones[38] != 2503 || icomIndex(alincoTones, 1598) >= 0 || icomIndex(alincoTones, 2541) >= 0 {
		t.Fatalf("subtoni delle Alinco: %v", alincoTones)
	}
	idents := map[string]string{"al-dj596": "DJ596", "al-dj175": "DJ175", "al-dr135t": "DR135", "al-dr235t": "DR235", "al-dr435t": "DR435",
		"al-dr03t": "DR135", "al-dr06t": "DR435"}
	for _, id := range alincoIDs {
		m := genByID[id]
		if m == nil {
			t.Fatalf("%s: modello mancante", id)
		}
		o := alincoO(m)
		if m.Vendor != "Alinco" || m.Slots != 100 || m.First != 0 || m.NameMax != 7 || m.Baud != 9600 || !m.NoTxOff || !m.Narrow || !m.OneDCS ||
			m.NoChirp || o.ident != idents[id] || m.HasScan == o.dj175 || o.dj175 != (id == "al-dj175") || (o.sig == nil) != o.dj175 {
			t.Errorf("%s: dati del modello", id)
		}
		if !strings.Contains(m.Name, "-") {
			t.Errorf("%s: nome mostrato %q", id, m.Name)
		}
		// ogni modello riconosce solo la propria memoria
		img := alincoBlank(m)
		for _, other := range alincoIDs {
			om := genByID[other]
			if got, want := om.Fam.Match(om, img), other == id; got != want {
				t.Errorf("memoria di %s riconosciuta da %s: %v", id, other, got)
			}
		}
		if !identCanMatch(m) && !o.dj175 {
			t.Errorf("%s: «Identifica» non lo distingue dalla memoria", id)
		}
	}
	for _, id := range []string{"al-djg7eg", "al-djg7t"} {
		m := genByID[id]
		if m == nil || m.Slots != 1000 || m.First != 0 || m.NameMax != 16 || m.Baud != 57600 || !m.NoTxOff || !m.Narrow || !m.HasScan ||
			m.Opt.(*alg7Opt).ident != "AL~"+m.Model || genMinSize(m) != 0x1200+1000*64 || genMinSize(m) > alg7Size {
			t.Errorf("%s: dati del modello", id)
		}
	}
}

func TestAlincoChannels(t *testing.T) {
	for _, id := range alincoIDs {
		m := genByID[id]
		f, o := m.Fam.(alincoFam), alincoO(m)
		img := alincoBlank(m)
		for i := 0; i < 100; i++ {
			if f.Get(m, img, i) != nil {
				t.Fatalf("%s: memoria vuota, canale %d presente", id, i)
			}
		}
		lo := m.Bands[0].Lo
		hi, lowP := m.Powers[len(m.Powers)-1].Sem, m.Powers[0].Sem
		chs := []*Chan{
			{Rx: lo + 500000, Tx: lo + 500000, Scan: true, Name: "SIMPLEX", Power: hi},
			{Rx: lo + 600000, Tx: lo, TxTone: "88.5", Scan: true, Name: "RPT 1", Power: lowP},
			{Rx: lo + 12500, Tx: lo + 612500, TxTone: "250.3", RxTone: "250.3", Scan: !o.dj175 == false, Name: "TSQL", Narrow: true, Power: hi},
			{Rx: lo + 25000, Tx: lo + 25000, TxTone: "D754N", RxTone: "D754N", Scan: true, Power: hi},
		}
		chs[2].Scan = o.dj175 // senza colonna della scansione il canale si rilegge sempre «in scansione»
		pos := []int{0, 7, 8, 99}
		before := append([]byte{}, img...)
		for k, c := range chs {
			if err := f.Set(m, img, pos[k], c); err != nil {
				t.Fatalf("%s canale %d: %v", id, pos[k], err)
			}
		}
		for k, c := range chs {
			if g := f.Get(m, img, pos[k]); g == nil || *g != *c {
				t.Fatalf("%s canale %d riletto diverso: %+v / %+v", id, pos[k], g, c)
			}
		}
		// fuori dai canali cambiano solo gli elenchi dei canali occupati e della
		// scansione; nella DJ-175 niente (da 0x100 a 0x1FF ci sono altre memorie)
		for a := 0; a < 0x200; a++ {
			if img[a] != before[a] && (o.dj175 || a < 0x120 || a >= 0x13D) {
				t.Fatalf("%s: byte 0x%03X cambiato scrivendo dei canali (%02X -> %02X)", id, a, before[a], img[a])
			}
		}
		if !o.dj175 && (img[0x120] != 0x81 || img[0x121] != 0x80 || img[0x12C] != 0x10 || img[0x131] != 0x80) {
			t.Errorf("%s: elenchi dei canali occupati e della scansione: % X / % X", id, img[0x120:0x12D], img[0x130:0x13D])
		}
		// nomi: maiuscole, cifre e spazio; sette caratteri
		if err := f.Set(m, img, 1, &Chan{Rx: lo, Tx: lo, Scan: true, Name: "ab-c 12xyz"}); err != nil {
			t.Fatal(err)
		}
		if g := f.Get(m, img, 1); g == nil || g.Name != "AB C 12" {
			t.Errorf("%s: nome %q", id, g.Name)
		}
		// rifiuti: frequenza o scostamento non multipli di 100 Hz, toni che la radio non ha
		bad := []*Chan{
			{Rx: lo + 6250 + 25, Tx: lo + 6250 + 25},
			{Rx: lo, Tx: lo + 600050},
			{Rx: lo, Tx: lo, TxTone: "159.8"},
			{Rx: lo, Tx: lo, RxTone: "88.5"},
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "100.0"},
			{Rx: lo, Tx: lo, TxTone: "D023I", RxTone: "D023I"},
			{Rx: lo, Tx: lo, TxTone: "D023N", RxTone: "D025N"},
			{Rx: lo, Tx: lo, TxTone: "D023N"},
			{Rx: lo, Tx: lo, TxTone: "88.5", RxTone: "D023N"},
		}
		snap := append([]byte{}, img...)
		for k, c := range bad {
			if err := f.Set(m, img, 7, c); err == nil {
				t.Errorf("%s: canale non valido %d accettato", id, k)
			}
			if !bytes.Equal(img, snap) {
				t.Fatalf("%s: canale non valido %d: memoria cambiata", id, k)
			}
		}
		// un canale che c'era già: il modo digitale resta, il resto si riscrive
		// (come CHIRP); uno nuovo si scrive sopra quello che c'era nella posizione
		r := f.rec(7)
		r.F("isdigital").SetBool(img, true)
		r.F("unknown3").SetInt(img, 0x15)
		if err := f.Set(m, img, 7, &Chan{Rx: lo + 100000, Tx: lo + 100000, Scan: true, Name: "EDIT"}); err != nil {
			t.Fatal(err)
		}
		if !r.F("isdigital").Bool(img) || r.F("unknown3").Int(img) != 0 || r.F("rtone").Int(img) != alincoToneDef {
			t.Errorf("%s: canale modificato: % X", id, r.Raw(img))
		}
		r = f.rec(50)
		r.F("unknown3").SetInt(img, 0x15)
		r.F("unknownA").Fill(img, 0xA5)
		if err := f.Set(m, img, 50, &Chan{Rx: lo, Tx: lo, Scan: true}); err != nil {
			t.Fatal(err)
		}
		if r.F("unknown3").Int(img) != 0x15 || r.F("unknownA").Raw(img)[5] != 0xA5 || r.F("offset").Int(img) != 6000 || r.F("rtone").Int(img) != alincoToneDef {
			t.Errorf("%s: canale nuovo: % X", id, r.Raw(img))
		}
		// cancellazione: solo il segno di canale occupato
		raw := append([]byte{}, f.rec(8).Raw(img)...)
		if err := f.Set(m, img, 8, nil); err != nil || f.Get(m, img, 8) != nil {
			t.Fatalf("%s: cancellazione", id)
		}
		if o.dj175 {
			raw[0] &^= 0x80
		}
		if !bytes.Equal(f.rec(8).Raw(img), raw) {
			t.Errorf("%s: la cancellazione ha cambiato il canale", id)
		}
		// spostamento con i dati del canale (PutRec): il canale torna com'era
		rec := f.Rec(m, img, 99)
		f.PutRec(m, img, 30, rec)
		if g, w := f.Get(m, img, 30), f.Get(m, img, 99); g == nil || *g != *w {
			t.Errorf("%s: canale spostato diverso: %+v / %+v", id, g, w)
		}
	}
}

func TestAlincoG7Channels(t *testing.T) {
	m := genByID["al-djg7eg"]
	f := m.Fam.(alg7Fam)
	img := bytes.Repeat([]byte{0xFF}, alg7Size)
	for i := 0; i < 1000; i++ {
		f.rec(i).F("empty").SetInt(img, 0)
	}
	type row struct {
		pos        int
		c          *Chan
		flag, mode int64
	}
	rows := []row{
		{0, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20"}, 2, 1},
		{1, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true, Name: "R0 sedici caratt"}, 2, 1},
		{2, &Chan{Rx: 430012500, Tx: 431612500, TxTone: "123.0", RxTone: "123.0", Narrow: true, Name: "TSQL"}, 2, 0},
		{3, &Chan{Rx: 1297500000, Tx: 1269500000, TxTone: "77.0", RxTone: "203.5", Scan: true, Name: "23cm"}, 2, 1},
		{4, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D023N", Scan: true}, 2, 1},
		{5, &Chan{Rx: 118100000, Tx: 118100000, Scan: true, Name: "TWR"}, 1, 2},       // banda aerea: AM
		{6, &Chan{Rx: 98500000, Tx: 98500000, Scan: true, Name: "RADIO"}, 1, 3},       // radiodiffusione: FM larga
		{7, &Chan{Rx: 27205000, Tx: 27205000, Scan: true, Name: "CB"}, 1, 1},          // fuori dalle bande principali
		{999, &Chan{Rx: 470000000, Tx: 470000000, Scan: true, Name: "LAST"}, 1, 1},    // 470 MHz è già fuori
		{998, &Chan{Rx: 136000000, Tx: 136000000, Scan: true, Name: "AIR END"}, 1, 2}, // AM anche se la banda principale arriva qui
	}
	for _, r := range rows {
		if err := f.Set(m, img, r.pos, r.c); err != nil {
			t.Fatalf("canale %d: %v", r.pos, err)
		}
		el := f.rec(r.pos)
		if g := f.Get(m, img, r.pos); g == nil || *g != *r.c {
			t.Fatalf("canale %d riletto diverso: %+v / %+v", r.pos, g, r.c)
		}
		if el.F("empty").Int(img) != r.flag || el.F("mode").Int(img) != r.mode {
			t.Errorf("canale %d: segno %d (atteso %d), modo %d (atteso %d)", r.pos, el.F("empty").Int(img), r.flag, el.F("mode").Int(img), r.mode)
		}
		if el.F("unknown1").Int(img) != 0x3E001C || el.F("unknown2").Int(img) != 0x0A || el.F("unknown3").Int(img) != 0 || el.F("unknown4").Int(img) != 0 {
			t.Errorf("canale %d: valori fissi", r.pos)
		}
	}
	// nome: un carattere ogni due byte
	if raw := f.rec(0).F("name").Raw(img); !bytes.Equal(raw[:8], []byte{'S', 0, '2', 0, '0', 0, 0, 0}) {
		t.Errorf("nome: % X", raw)
	}
	// l'AM di un canale resta se la frequenza non cambia; cambia con la frequenza
	if err := f.Set(m, img, 5, &Chan{Rx: 118100000, Tx: 118100000, Scan: false, Name: "TOWER"}); err != nil || f.rec(5).F("mode").Int(img) != 2 {
		t.Errorf("AM non conservata")
	}
	f.rec(0).F("mode").SetInt(img, 2) // AM scelta sulla radio per un canale a 145,5 MHz
	if err := f.Set(m, img, 0, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "S20 AM"}); err != nil || f.rec(0).F("mode").Int(img) != 2 || f.rec(0).F("empty").Int(img) != 1 {
		t.Errorf("AM scelta sulla radio non conservata")
	}
	if err := f.Set(m, img, 0, &Chan{Rx: 145525000, Tx: 145525000, Scan: true, Name: "S21"}); err != nil || f.rec(0).F("mode").Int(img) != 1 || f.rec(0).F("empty").Int(img) != 2 {
		t.Errorf("modo non aggiornato con la frequenza")
	}
	// FM scelta sulla radio in banda aerea: resta finché la frequenza non cambia
	f.rec(5).F("mode").SetInt(img, 1)
	if err := f.Set(m, img, 5, &Chan{Rx: 118100000, Tx: 118100000, Scan: true, Name: "TWR FM"}); err != nil || f.rec(5).F("mode").Int(img) != 1 {
		t.Errorf("FM scelta sulla radio non conservata")
	}
	// squelch «al contrario»: resta se non si dà un tono, si sostituisce se lo si dà
	f.rec(4).F("squelch_type").SetInt(img, 4)
	if g := f.Get(m, img, 4); g == nil || g.TxTone != "" || g.RxTone != "" {
		t.Fatalf("squelch al contrario: %+v", g)
	}
	if err := f.Set(m, img, 4, &Chan{Rx: 433500000, Tx: 433500000, Scan: true, Name: "TSQL-R"}); err != nil || f.rec(4).F("squelch_type").Int(img) != 4 {
		t.Errorf("squelch al contrario non conservato")
	}
	if err := f.Set(m, img, 4, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "100.0", Scan: true}); err != nil || f.rec(4).F("squelch_type").Int(img) != 2 {
		t.Errorf("squelch al contrario non sostituito")
	}
	if err := f.Set(m, img, 4, &Chan{Rx: 433500000, Tx: 433500000, Scan: true}); err != nil || f.rec(4).F("squelch_type").Int(img) != 0 {
		t.Errorf("tono non tolto")
	}
	// rifiuti
	snap := append([]byte{}, img...)
	for k, c := range []*Chan{
		{Rx: 145000000, Tx: 145000000, TxTone: "159.8"},
		{Rx: 145000000, Tx: 145000000, RxTone: "88.5"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023I", RxTone: "D023I"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023N", RxTone: "D025N"},
		{Rx: 145000000, Tx: 145000000, TxTone: "88.5", RxTone: "D023N"},
		{Rx: 145000000, Tx: 145000000, TxTone: "D023N", RxTone: "88.5"},
	} {
		if err := f.Set(m, img, 10, c); err == nil {
			t.Errorf("canale non valido %d accettato", k)
		}
		if !bytes.Equal(img, snap) {
			t.Fatalf("canale non valido %d: memoria cambiata", k)
		}
	}
	// cancellazione: canale a FF con il segno a zero (come CHIRP)
	if err := f.Set(m, img, 3, nil); err != nil || f.Get(m, img, 3) != nil {
		t.Fatal("cancellazione")
	}
	if raw := f.rec(3).Raw(img); raw[0] != 0 || !bytes.Equal(raw[1:], bytes.Repeat([]byte{0xFF}, 63)) {
		t.Errorf("canale cancellato: % X", raw)
	}
	// fuori dai canali non cambia niente
	if !bytes.Equal(img[:0x1200], bytes.Repeat([]byte{0xFF}, 0x1200)) {
		t.Errorf("byte fuori dai canali cambiati")
	}
}

// alSession: una sessione con la radio simulata; wrap (se c'è) sta tra il
// programma e il cavo
func alSession(t *testing.T, e genEmu, wrap func(p *pipeEnd) Port, run func() error) (error, []byte) {
	t.Helper()
	a, b := newPipe()
	rec := &recPort{Port: b}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a.baud = baud
		if wrap != nil {
			return wrap(a), nil
		}
		return a, nil
	}
	defer func() { openPortFunc = old }()
	done := make(chan error, 1)
	go func() { done <- e.serve(rec) }()
	err := run()
	_ = a.Close()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("la radio simulata non ha terminato la sessione")
	}
	return err, rec.got()
}

// alCorrupt: cavo che rovina un byte (il numero n fra quelli ricevuti)
type alCorrupt struct {
	Port
	mu sync.Mutex
	n  int
}

func (p *alCorrupt) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	err := p.Port.ReadFull(ctx, b, to)
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		for k := range b {
			if p.n--; p.n == 0 {
				b[k] ^= 0x01
			}
		}
	}
	return err
}

// alBaud: radio che sente solo a una velocità (alle altre i dati si perdono)
type alBaud struct {
	Port
	mu    sync.Mutex
	cur   int
	want  int
	bauds []int
}

func (p *alBaud) SetBaud(b int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cur = b
	p.bauds = append(p.bauds, b)
	return nil
}

func (p *alBaud) Write(b []byte) error {
	p.mu.Lock()
	ok := p.cur == p.want
	p.mu.Unlock()
	if !ok {
		return nil
	}
	return p.Port.Write(b)
}

func TestAlincoCable(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"al-dj596", "al-dj175", "al-djg7eg"} {
		m := genByID[id]
		t.Run(id, func(t *testing.T) {
			_, g7 := m.Fam.(alg7Fam)
			size, first := alg7Size, 0x200
			if !g7 {
				size, first = alincoO(m).size, 0x100
			}
			base := make([]byte, size)
			for k := range base {
				base[k] = byte(k*7 + k>>8)
			}
			mk := func() *alincoEmu { return genEmuFor(m, base).(*alincoEmu) }
			// lettura con il cavo che rimanda i dati e con quello che non lo fa:
			// stessa immagine, stessi byte inviati alla radio
			var sent [2][]byte
			for k, noEcho := range []bool{false, true} {
				e := mk()
				e.noEcho = noEcho
				var got []byte
				err, tx := alSession(t, e, nil, func() error {
					var err error
					got, _, err = m.Fam.Download(ctx, m, "EMU", nop)
					return err
				})
				if err != nil {
					t.Fatalf("lettura (senza eco: %v): %v", noEcho, err)
				}
				want := append([]byte{}, base...)
				if g7 {
					copy(want, make([]byte, 0x200)) // i primi 0x200 byte non si leggono
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("lettura (senza eco: %v): immagine diversa: %s", noEcho, diffRanges(want, got))
				}
				sent[k] = tx
			}
			if !bytes.Equal(sent[0], sent[1]) {
				t.Errorf("lettura: byte inviati diversi con e senza eco")
			}
			// scrittura, con e senza eco
			mod := append([]byte{}, base...)
			for k := first; k < size; k++ {
				mod[k] ^= 0x5A
			}
			for k, noEcho := range []bool{false, true} {
				e := mk()
				e.noEcho = noEcho
				err, tx := alSession(t, e, nil, func() error {
					_, err := m.Fam.Upload(ctx, m, "EMU", mod, nop)
					return err
				})
				if err != nil {
					t.Fatalf("scrittura (senza eco: %v): %v", noEcho, err)
				}
				fin := e.image()
				if !bytes.Equal(fin[first:], mod[first:]) || !bytes.Equal(fin[:first], base[:first]) {
					t.Fatalf("scrittura (senza eco: %v): memoria finale diversa", noEcho)
				}
				if e.writes != (size-first)/map[bool]int{true: 64, false: 16}[g7] {
					t.Errorf("scrittura: %d blocchi", e.writes)
				}
				sent[k] = tx
			}
			if !bytes.Equal(sent[0], sent[1]) {
				t.Errorf("scrittura: byte inviati diversi con e senza eco")
			}
			// la fine della copia: «AL~E» nella prima famiglia, niente nella DJ-G7
			if has := bytes.HasSuffix(sent[0], []byte("AL~E\r\n")); has == g7 {
				t.Errorf("fine della copia: AL~E inviato=%v", has)
			}
			// un blocco letto male: errore con l'indirizzo, niente immagine
			e := mk()
			e.badAt = 0x400
			err, tx := alSession(t, e, nil, func() error {
				_, _, err := m.Fam.Download(ctx, m, "EMU", nop)
				return err
			})
			if err == nil || !strings.Contains(err.Error(), "400") {
				t.Errorf("blocco rovinato: %v", err)
			}
			if !g7 && !bytes.HasSuffix(tx, []byte("AL~E\r\n")) {
				t.Errorf("blocco rovinato: la radio non è stata riportata al funzionamento normale")
			}
			// radio che non risponde al suo nome (spenta, o un altro modello): tre
			// tentativi come CHIRP (per la DJ-G7 anche a 19200 baud), poi l'errore
			e = mk()
			e.deaf = true
			err, tx = alSession(t, e, nil, func() error {
				_, _, err := m.Fam.Download(ctx, m, "EMU", nop)
				return err
			})
			tries := bytes.Count(tx, []byte(e.ident+"\r\n"))
			if err == nil || !strings.Contains(err.Error(), "non risponde") || tries != map[bool]int{true: 6, false: 3}[g7] || len(tx) != tries*(len(e.ident)+2) {
				t.Errorf("radio che non risponde: %v, %d tentativi, %d byte inviati", err, tries, len(tx))
			}
			// scrittura mai cominciata se la radio non risponde
			e = mk()
			e.deaf = true
			err, _ = alSession(t, e, nil, func() error {
				_, err := m.Fam.Upload(ctx, m, "EMU", mod, nop)
				return err
			})
			if err == nil || e.writes != 0 || !bytes.Equal(e.image(), base) {
				t.Errorf("scrittura con la radio che non risponde: %v, %d blocchi", err, e.writes)
			}
		})
	}
}

// prima famiglia: la radio non conferma i blocchi, ma se il cavo rimanda dei
// dati diversi da quelli inviati la scrittura si ferma e lo dice
func TestAlincoEchoWrong(t *testing.T) {
	m := genByID["al-dj596"]
	img := alincoBlank(m)
	e := genEmuFor(m, img).(*alincoEmu)
	err, tx := alSession(t, e, func(p *pipeEnd) Port { return &alCorrupt{Port: p, n: 600} }, func() error {
		_, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "diversi da quelli inviati") {
		t.Fatalf("eco rovinata: %v", err)
	}
	if !bytes.HasSuffix(tx, []byte("AL~E\r\n")) || e.writes > 20 {
		t.Errorf("eco rovinata: scrittura non fermata subito (%d blocchi) o radio lasciata in copia", e.writes)
	}
}

// prima famiglia: non si sa se le radio vere confermano i blocchi scritti (CHIRP
// non lo chiede e non lo legge). Se lo fanno, la scrittura deve riuscire lo
// stesso, con il cavo che rimanda i dati e con quello che non lo fa; e una
// lettura con un'eco rovinata deve riuscire lo stesso
func TestAlincoWriteAck(t *testing.T) {
	ctx := context.Background()
	nop := func(int, int, string) {}
	for _, id := range []string{"al-dj596", "al-dj175"} {
		m := genByID[id]
		o := alincoO(m)
		img := alincoBlank(m)
		for k := 0x100; k < len(img); k++ {
			img[k] = byte(k * 5)
		}
		for _, noEcho := range []bool{false, true} {
			e := genEmuFor(m, alincoBlank(m)).(*alincoEmu)
			e.ackW, e.noEcho = true, noEcho
			err, _ := alSession(t, e, nil, func() error {
				_, err := m.Fam.Upload(ctx, m, "EMU", img, nop)
				return err
			})
			if err != nil {
				t.Fatalf("%s, radio che conferma i blocchi (senza eco: %v): %v", id, noEcho, err)
			}
			if fin := e.image(); !bytes.Equal(fin[0x100:], img[0x100:]) || e.writes != (o.size-0x100)/16 {
				t.Fatalf("%s, radio che conferma i blocchi (senza eco: %v): memoria finale diversa (%d blocchi)", id, noEcho, e.writes)
			}
		}
		// lettura con un byte dell'eco rovinato (il quinto dell'eco del sesto
		// comando di lettura): l'immagine è quella giusta. Un byte rovinato nei
		// dati, invece, non si può riconoscere: il protocollo non ha somme di
		// controllo (vale anche per CHIRP)
		per := 11 + 41
		if o.dj175 {
			per = 11 + 34
		}
		e := genEmuFor(m, img).(*alincoEmu)
		var got []byte
		err, _ := alSession(t, e, func(p *pipeEnd) Port { return &alCorrupt{Port: p, n: 13 + 5*per + 5} }, func() error {
			var err error
			got, _, err = m.Fam.Download(ctx, m, "EMU", nop)
			return err
		})
		if err != nil || !bytes.Equal(got, e.image()) {
			t.Errorf("%s, lettura con l'eco rovinata: %v", id, err)
		}
	}
}

// DJ-G7: la radio conferma ogni blocco; se non lo accetta la scrittura si ferma
func TestAlincoG7Refused(t *testing.T) {
	m := genByID["al-djg7t"]
	img := make([]byte, alg7Size)
	e := genEmuFor(m, img).(*alincoEmu)
	e.nakAt = 0x1240
	err, _ := alSession(t, e, nil, func() error {
		_, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "01240") || e.writes != (0x1240-0x200)/64 {
		t.Fatalf("blocco rifiutato: %v, %d blocchi scritti", err, e.writes)
	}
}

// DJ-G7 impostata a 19200 baud: dopo i tentativi a 57600 il programma passa a
// 19200 (come CHIRP) e legge
func TestAlincoG7Baud(t *testing.T) {
	m := genByID["al-djg7eg"]
	base := make([]byte, alg7Size)
	for k := range base {
		base[k] = byte(k * 3)
	}
	var port *alBaud
	var got []byte
	err, tx := alSession(t, genEmuFor(m, base), func(p *pipeEnd) Port {
		port = &alBaud{Port: p, cur: p.baud, want: 19200}
		return port
	}, func() error {
		var err error
		got, _, err = m.Fam.Download(context.Background(), m, "EMU", func(int, int, string) {})
		return err
	})
	if err != nil {
		t.Fatalf("lettura a 19200 baud: %v", err)
	}
	if len(port.bauds) != 1 || port.bauds[0] != 19200 {
		t.Errorf("velocità impostate: %v", port.bauds)
	}
	if !bytes.Equal(got[0x200:], base[0x200:]) {
		t.Errorf("immagine diversa")
	}
	if n := bytes.Count(tx, []byte("AL~DJ-G7EG\r\n")); n != 1 {
		t.Errorf("la radio ha ricevuto %d volte il nome (attesa: una, a 19200 baud)", n)
	}
}

// in «Identifica» le Alinco si provano dopo tutte le altre radio; una Alinco già
// riconosciuta su quella porta, invece, per prima
func TestAlincoIdentLast(t *testing.T) {
	isAl := func(p identProbe) bool {
		for _, m := range p.kinds {
			if m.Vendor == "Alinco" {
				return true
			}
		}
		return false
	}
	st := mustStore(t)
	a := newApp(st)
	identPortMu.Lock()
	plan := a.identPlan(&identReq{Port: "COM7"})
	identPortMu.Unlock()
	first, last, n := -1, -1, 0
	for i, p := range plan {
		if isAl(p) {
			n++
			if first < 0 {
				first = i
			}
		} else {
			last = i
		}
	}
	if n != 7 || first < last {
		t.Fatalf("gruppi Alinco: %d, il primo al posto %d, l'ultimo delle altre radio al posto %d (su %d)", n, first, last, len(plan))
	}
	a.rememberRadio("COM7", "al-dj596")
	identPortMu.Lock()
	plan = a.identPlan(&identReq{Port: "COM7"})
	identPortMu.Unlock()
	if len(plan) == 0 || len(plan[0].kinds) == 0 || plan[0].kinds[0].ID != "al-dj596" {
		t.Errorf("Alinco già riconosciuta: non è la prima del giro")
	}
}

// «Identifica» con i tempi veri trova la DJ-G7 sia a 57600 baud sia impostata a
// 19200: dopo il tentativo a 57600 senza risposta il driver passa a 19200 nello
// stesso collegamento
func TestAlincoG7Identify(t *testing.T) {
	m := genByID["al-djg7eg"]
	for _, want := range []int{57600, 19200} {
		a := newApp(mustStore(t))
		identPortMu.Lock()
		plan := a.identPlan(&identReq{Port: "COM7"})
		identPortMu.Unlock()
		var grp []identProbe
		for _, p := range plan {
			for _, k := range p.kinds {
				if k == m {
					grp = append(grp, p)
				}
			}
		}
		if len(grp) != 1 || len(grp[0].kinds) != 1 {
			t.Fatalf("gruppi della DJ-G7EG: %d", len(grp))
		}
		emu := genEmuFor(m, make([]byte, alg7Size))
		var mu sync.Mutex
		var ports []*alBaud
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			x, y := newPipe()
			x.baud = baud
			go emu.serve(y)
			ab := &alBaud{Port: x, cur: baud, want: want}
			mu.Lock()
			ports = append(ports, ab)
			mu.Unlock()
			return ab, nil
		}
		j := &cableJob{Notes: []string{}, Log: []string{}}
		identPortMu.Lock()
		res, err := a.identRun(context.Background(), j, &identReq{Port: "COM7"}, grp, time.Now(),
			&identResult{Fields: []identField{}, Others: []identOther{}, Answers: []string{}})
		identPortMu.Unlock()
		openPortFunc = old
		if err != nil || res == nil || !res.Found || res.Kind != m.ID {
			t.Fatalf("radio a %d baud: non identificata: %v\n%s", want, err, strings.Join(j.Log, "\n"))
		}
		mu.Lock()
		last := ports[len(ports)-1]
		mu.Unlock()
		last.mu.Lock()
		cur := last.cur
		last.mu.Unlock()
		if cur != want {
			t.Errorf("radio a %d baud: identificata con la porta a %d", want, cur)
		}
	}
}
