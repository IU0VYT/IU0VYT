//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: BTECH UV-PRO (drv_uvpro.go). CHIRP non ha file per questa radio: il
// suo driver parla con la radio simulata e deve leggere gli stessi canali e,
// per le stesse modifiche, scrivere gli stessi 25 byte. Poi le regole dei
// canali, il blocco della scrittura con un firmware non provato, la rilettura
// dopo ogni canale e il ripristino quando non coincide.

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
)

func (uvproFam) blankImg(m *genModel) []byte { return uvproBlank(m) }

func uvproBase(t *testing.T) (*genModel, []byte) {
	t.Helper()
	m := genByID["bt-uvpro"]
	f := uvproFam{}
	img := uvproBlank(m)
	img[uvproGroupAt] = 2
	copy(img[uvproNameAt:], "GRUPPO 3")
	for i, c := range map[int]*Chan{
		0:  {Rx: 145500000, Tx: 145500000, Power: 0, Scan: true, Name: "S20"},
		1:  {Rx: 145750000, Tx: 145150000, TxTone: "88.5", Power: 1, Scan: true, Name: "R6"},
		2:  {Rx: 430100000, Tx: 431700000, TxTone: "123.0", RxTone: "123.0", Power: 2, Name: "RU2 Roma"},
		3:  {Rx: 433500000, Tx: 433500000, TxTone: "D754N", RxTone: "D023N", Power: 0, Narrow: true, Scan: true, Name: "DCS"},
		4:  {Rx: 121500000, Tx: 121500000, TxOff: true, Power: 0, Scan: true, Name: "EMERGENZA"},
		5:  {Rx: 446006250, Tx: 446006250, TxOff: true, Power: 2, Narrow: true, Scan: true, Name: "PMR 1 RX"},
		6:  {Rx: 145200000, Tx: 435000000, RxTone: "D071N", Power: 0, Scan: true, Name: "CROSSBAND"},
		7:  {Rx: 162550000, Tx: 162550000, TxTone: "67.0", RxTone: "254.1", Power: 1, Scan: true, Name: "TONI"},
		29: {Rx: 439975000, Tx: 439975000, Power: 0, Scan: true, Name: "ULTIMO"},
	} {
		if err := f.Set(m, img, i, c); err != nil {
			t.Fatalf("posizione %d: %v", i, err)
		}
	}
	return m, img
}

func TestUVProModel(t *testing.T) {
	m := genByID["bt-uvpro"]
	if m == nil {
		t.Fatal("modello mancante")
	}
	f := uvproFam{}
	if m.Slots != 30 || m.First != 0 || m.NameMax != 10 || !m.NoChirp || m.NoTxOff || !m.HasScan || !m.Narrow || len(m.Powers) != 3 ||
		m.Powers[0].Sem != 2 || m.Powers[2].Sem != 0 || genMinSize(m) != 64+30*32 {
		t.Errorf("dati del modello")
	}
	img := uvproBlank(m)
	if !f.Match(m, img) || f.Match(m, img[1:]) || binary.BigEndian.Uint16(img[uvproInfoAt+4:]) != 147 {
		t.Errorf("immagine vuota")
	}
	for i := 0; i < m.Slots; i++ {
		if f.Get(m, img, i) != nil || !bytes.Equal(f.ch(img, i), uvproEmpty(i)) {
			t.Fatalf("posizione %d non vuota", i)
		}
	}
	// i blocchi: FF 01 00 lunghezza, gruppo 2, comando, dati
	if got := uvproFrame(13, []byte{7}); !bytes.Equal(got, []byte{0xFF, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x0D, 0x07}) {
		t.Errorf("blocco di lettura: %x", got)
	}
	if got := uvproFrame(4, []byte{3}); !bytes.Equal(got, []byte{0xFF, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x04, 0x03}) {
		t.Errorf("blocco delle informazioni: %x", got)
	}
	// toni: 0 nessuno, sotto 6700 il codice DCS, altrimenti centesimi di Hz
	for raw, want := range map[uint16]string{0: "", 23: "D023N", 754: "D754N", 6700: "67.0", 8850: "88.5", 25410: "254.1", 10000: "100.0"} {
		if got := uvproTone(raw); got != want {
			t.Errorf("tono %d: %q, atteso %q", raw, got, want)
		}
		if want != "" {
			tn, _ := genToneParse(want)
			if back, err := uvproSub(tn); err != nil || back != raw {
				t.Errorf("tono %q: valore %d (%v), atteso %d", want, back, err, raw)
			}
		}
	}
}

func TestUVProChannels(t *testing.T) {
	m, img := uvproBase(t)
	f := uvproFam{}
	// un canale byte per byte: R6, 145,750 -600 kHz, 88,5 Hz in trasmissione,
	// potenza media, in scansione, banda larga
	want := []byte{0x01, 0x08, 0xA6, 0xD0, 0x30, 0x08, 0xAF, 0xF7, 0xF0, 0x22, 0x92, 0x00, 0x00, 0x92, 0x00, 'R', '6', 0, 0, 0, 0, 0, 0, 0, 0}
	if got := f.ch(img, 1); !bytes.Equal(got, want) {
		t.Errorf("canale R6\n  %x\nvoluto\n  %x", got, want)
	}
	// banda aerea: AM, frequenza di trasmissione a zero, trasmissione disabilitata, banda larga
	p := f.ch(img, 4)
	if binary.BigEndian.Uint32(p[1:5]) != 0x40000000 || binary.BigEndian.Uint32(p[5:9]) != 0x40000000|121500000 || p[13] != 0x80|0x40|0x10|0x01 {
		t.Errorf("canale in banda aerea: %x", p)
	}
	// sola ricezione fuori dalla banda aerea: la frequenza resta, FM
	p = f.ch(img, 5)
	if binary.BigEndian.Uint32(p[1:5]) != 446006250 || p[13] != 0x80|0x01 {
		t.Errorf("canale in sola ricezione: %x", p)
	}
	rt := func(i int, c Chan) {
		t.Helper()
		if err := f.Set(m, img, i, &c); err != nil {
			t.Fatalf("posizione %d, %+v: %v", i, c, err)
		}
		g := f.Get(m, img, i)
		if g == nil || *g != c {
			t.Errorf("posizione %d:\n  scritto %+v\n  riletto %+v", i, c, g)
		}
		if p := f.ch(img, i); int(p[0]) != i {
			t.Errorf("posizione %d: numero del canale %d", i, p[0])
		}
	}
	for _, i := range []int{0, 15, 29} {
		for pw := 0; pw < 3; pw++ {
			rt(i, Chan{Rx: 145500000, Tx: 145500000, Power: pw, Scan: true, Name: "POTENZA"})
		}
		rt(i, Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", Power: 0, Scan: true, Name: "R6"})
		rt(i, Chan{Rx: 430100000, Tx: 431700000, TxTone: "123.0", RxTone: "123.0", Power: 0, Name: "0123456789"})
		rt(i, Chan{Rx: 433500000, Tx: 433500000, RxTone: "100.0", Power: 0, Narrow: true, Scan: true, Name: "solo RX"})
		rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D754N", Power: 0, Scan: true, Name: "a/B-c_d"})
		rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "88.5", Power: 0, Scan: true})
		rt(i, Chan{Rx: 433500000, Tx: 433500000, TxTone: "88.5", RxTone: "D023N", Power: 0, Scan: true})
		rt(i, Chan{Rx: 145200000, Tx: 435000000, Power: 0, Scan: true, Name: "SAT"})
		rt(i, Chan{Rx: 446006250, Tx: 446006250, TxOff: true, Power: 2, Narrow: true, Scan: true, Name: "RX"})
		rt(i, Chan{Rx: 118100000, Tx: 118100000, TxOff: true, Power: 0, Scan: true, Name: "TWR"})
		rt(i, Chan{Rx: 135975000, Tx: 135975000, TxOff: true, Power: 2, Name: "AIR MAX"})
	}
	// banda aerea senza «sola ricezione»: si salva comunque in sola ricezione
	if err := f.Set(m, img, 10, &Chan{Rx: 120500000, Tx: 120500000, Power: 0, Narrow: true, Scan: true, Name: "APP"}); err != nil {
		t.Fatal(err)
	}
	if g := f.Get(m, img, 10); g == nil || !g.TxOff || g.Narrow || f.ch(img, 10)[13]&0x11 != 0x11 {
		t.Errorf("banda aerea: %+v", g)
	}
	// i tre bit che il programma non gestisce e il secondo byte di opzioni restano
	p = f.ch(img, 11)
	p[13], p[14] = 0x2C, 0xF0
	rt(11, Chan{Rx: 145500000, Tx: 145500000, Power: 2, Name: "BIT"})
	if p = f.ch(img, 11); p[13] != 0x2C|0x10 || p[14] != 0xF0 {
		t.Errorf("opzioni non gestite: %02x %02x", p[13], p[14])
	}
	// rifiuti: la memoria non cambia
	for _, c := range []Chan{
		{Rx: 145500000, Tx: 145500000, TxTone: "D023I", RxTone: "D023N"},
		{Rx: 145500000, Tx: 145500000, RxTone: "D754I"},
		{Rx: 145500000, Tx: 145500000, TxTone: "88.6"},
		{Rx: 145500000, Tx: 145500000, TxTone: "D999N"},
		{Rx: 120500000, Tx: 120500000, TxTone: "88.5", TxOff: true},
		{Rx: 120500000, Tx: 120500000, RxTone: "D023N"},
		{Rx: 120500000, Tx: 121100000},
		{Rx: 107900000, Tx: 107900000},
		{Rx: 175000000, Tx: 175000000},
		{Rx: 399000000, Tx: 399000000},
		{Rx: 521000000, Tx: 521000000},
		{Rx: 145500000, Tx: 120500000},
		{Rx: 145500000, Tx: 300000000},
		{Rx: 0, Tx: 0},
	} {
		before := append([]byte{}, img...)
		c.Scan = true
		if err := f.Set(m, img, 12, &c); err == nil {
			t.Errorf("%+v accettato", c)
		}
		if !bytes.Equal(before, img) {
			t.Fatalf("%+v rifiutato ma l'immagine è cambiata", c)
		}
	}
	// cancellazione: il numero e 24 zeri
	if err := f.Set(m, img, 2, nil); err != nil {
		t.Fatal(err)
	}
	if f.Get(m, img, 2) != nil || !bytes.Equal(f.ch(img, 2), uvproEmpty(2)) {
		t.Errorf("posizione cancellata: %x", f.ch(img, 2))
	}
	// posizioni che la radio non ha dato, o con un canale DMR: non si toccano
	lk := genLockedFn(m, img)
	f.rec(img, 20)[0] = uvproNone
	f.rec(img, 21)[0] = uvproForeign
	rt(22, Chan{Rx: 145500000, Tx: 145500000, Power: 0, Scan: true, Name: "DMR"})
	f.ch(img, 22)[5] |= 0x80
	for _, i := range []int{20, 21, 22} {
		before := append([]byte{}, img...)
		for _, c := range []*Chan{nil, {Rx: 145500000, Tx: 145500000, Scan: true}} {
			if f.Set(m, img, i, c) == nil || lk(i, nil, c) == nil {
				t.Errorf("posizione %d modificata", i)
			}
		}
		if !bytes.Equal(before, img) {
			t.Fatalf("posizione %d: l'immagine è cambiata", i)
		}
	}
	if g := f.Get(m, img, 22); g == nil || g.Rx != 145500000 || g.Name != "DMR" || f.Get(m, img, 20) != nil || f.Get(m, img, 21) != nil {
		t.Errorf("lettura delle posizioni non modificabili")
	}
	if lk(0, nil, nil) != nil {
		t.Errorf("posizione normale bloccata")
	}
}

// uvproWrites: i canali inviati con il comando di scrittura (14)
func uvproWrites(tx []byte) [][]byte {
	var out [][]byte
	for len(tx) >= 8 {
		if tx[0] != 0xFF || tx[1] != 0x01 {
			tx = tx[1:]
			continue
		}
		n := int(tx[3])
		if len(tx) < 8+n {
			break
		}
		if tx[7] == 14 && tx[6] == 0 {
			out = append(out, append([]byte{}, tx[8:8+n]...))
		}
		tx = tx[8+n:]
	}
	return out
}

func TestUVProChirp(t *testing.T) {
	chirpDir(t)
	m, base := uvproBase(t)
	f := uvproFam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	// lettura di tutti i canali
	var ops []map[string]interface{}
	for i := 0; i < m.Slots; i++ {
		ops = append(ops, map[string]interface{}{"op": "get", "n": i})
	}
	res, _ := liveSession(t, f.emu(m, base), "BTECHUVPRORadio", ops)
	full := 0
	for i := 0; i < m.Slots; i++ {
		var ref *chirpChan
		if err := json.Unmarshal(res[i], &ref); err != nil {
			t.Fatalf("posizione %d: CHIRP risponde %s", i, res[i])
		}
		g := f.Get(m, base, i)
		if (g == nil) != (ref == nil) {
			t.Errorf("posizione %d: qui %v, CHIRP %s", i, g, res[i])
			continue
		}
		if g == nil {
			continue
		}
		full++
		mode := "FM"
		switch {
		case uvproAir(g.Rx):
			mode = "AM"
		case g.Narrow:
			mode = "NFM"
		}
		if g.Rx != ref.Rx || g.Tx != ref.Tx || g.TxOff != ref.TxOff || g.TxTone != ref.TxTone || g.RxTone != ref.RxTone || g.Name != ref.Name ||
			g.Scan != ref.Scan || mode != ref.Mode || m.powerIndex(g.Power) != ref.Power {
			t.Errorf("posizione %d:\n  qui   %+v (%s)\n  CHIRP %s", i, *g, mode, res[i])
		}
	}
	if full != 9 {
		t.Fatalf("%d canali confrontati in lettura", full)
	}
	// le stesse modifiche con CHIRP e da qui
	type ed struct {
		n int
		c *Chan
	}
	edits := []ed{
		{0, &Chan{Rx: 145525000, Tx: 145525000, Power: 2, Scan: true, Name: "S21"}},
		{1, &Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", RxTone: "88.5", Power: 1, Name: "R6"}},
		{2, nil},
		{3, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "D023N", RxTone: "D023N", Power: 0, Scan: true, Name: "DCS"}},
		{4, &Chan{Rx: 118100000, Tx: 118100000, TxOff: true, Power: 0, Scan: true, Name: "TORRE"}},
		{5, &Chan{Rx: 446006250, Tx: 446006250, Power: 2, Narrow: true, Scan: true, Name: "PMR 1"}},
		{6, &Chan{Rx: 145200000, Tx: 435000000, TxTone: "D071N", RxTone: "100.0", Power: 0, Scan: true, Name: "CROSSBAND"}},
		{10, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "71.9", Power: 1, Scan: true, Name: "NUOVO"}},
		{11, &Chan{Rx: 432500000, Tx: 432500000, RxTone: "D754N", Power: 2, Narrow: true, Name: "nuovo 2"}},
		{12, &Chan{Rx: 433000000, Tx: 438000000, TxTone: "D023N", RxTone: "D754N", Power: 0, Scan: true}},
		{13, &Chan{Rx: 145000000, Tx: 145000000, TxOff: true, Power: 0, Scan: true, Name: "SOLO RX"}},
		{14, &Chan{Rx: 127500000, Tx: 127500000, TxOff: true, Power: 2, Scan: true, Name: "AEREA"}},
		{29, nil},
	}
	ops = nil
	mod := append([]byte{}, base...)
	for _, e := range edits {
		if e.c == nil {
			ops = append(ops, map[string]interface{}{"op": "del", "n": e.n})
		} else {
			mode := "FM"
			switch {
			case uvproAir(e.c.Rx):
				mode = "AM"
			case e.c.Narrow:
				mode = "NFM"
			}
			ops = append(ops, map[string]interface{}{"op": "set", "n": e.n, "chan": map[string]interface{}{
				"rx": e.c.Rx, "tx": e.c.Tx, "txOff": e.c.TxOff, "txTone": e.c.TxTone, "rxTone": e.c.RxTone,
				"power": m.powerIndex(e.c.Power), "mode": mode, "scan": e.c.Scan, "name": e.c.Name}})
		}
		if err := f.Set(m, mod, e.n, e.c); err != nil {
			t.Fatalf("posizione %d: %v", e.n, err)
		}
	}
	ce := f.emu(m, base).(*uvproEmu)
	res, chirpTx := liveSession(t, ce, "BTECHUVPRORadio", ops)
	for k, r := range res {
		if string(r) != `"ok"` {
			t.Errorf("CHIRP sulla posizione %d: %s", edits[k].n, r)
		}
	}
	ge := f.emu(m, base).(*uvproEmu)
	err, gtx := alSession(t, ge, nil, func() error {
		_, err := f.Upload(ctx, m, "EMU", mod, nop)
		return err
	})
	if err != nil {
		t.Fatalf("scrittura: %v", err)
	}
	cw, gw := uvproWrites(chirpTx), uvproWrites(gtx)
	if len(cw) != len(edits) || len(gw) != len(edits) {
		t.Fatalf("canali scritti: CHIRP %d, qui %d, modifiche %d", len(cw), len(gw), len(edits))
	}
	for k := range cw {
		if !bytes.Equal(cw[k], gw[k]) {
			t.Errorf("posizione %d:\n  CHIRP %x\n  qui   %x", edits[k].n, cw[k], gw[k])
		}
	}
	if !bytes.Equal(ce.image(), ge.image()) || !bytes.Equal(ge.image(), mod) {
		t.Errorf("la radio alla fine non ha gli stessi canali")
	}
	// e gli stessi comandi, uno per uno (CHIRP legge il canale prima della
	// modifica, prima della scrittura e dopo; qui prima si rilegge tutta la
	// radio, poi ogni canale prima e dopo la scrittura)
	if ce.writes != ge.writes || ge.reads != m.Slots+2*len(edits) {
		t.Errorf("scritture %d/%d, letture qui %d", ce.writes, ge.writes, ge.reads)
	}
	t.Logf("%d canali letti come CHIRP; %d modifiche: gli stessi 25 byte di CHIRP per ogni canale", full, len(edits))
}

func TestUVProCable(t *testing.T) {
	m, base := uvproBase(t)
	f := uvproFam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	mk := func() *uvproEmu { return f.emu(m, base).(*uvproEmu) }
	read := func(e *uvproEmu) ([]byte, string, error, []byte) {
		var img []byte
		var note string
		err, tx := alSession(t, e, nil, func() error {
			var err error
			img, note, err = f.Download(ctx, m, "EMU", nop)
			return err
		})
		return img, note, err, tx
	}
	write := func(e *uvproEmu, img []byte) (error, []byte) {
		return alSession(t, e, nil, func() error {
			_, err := f.Upload(ctx, m, "EMU", img, nop)
			return err
		})
	}
	// lettura: informazioni, stato, nome del gruppo, poi i 30 canali
	e := mk()
	img, note, err, tx := read(e)
	if err != nil || !bytes.Equal(img, base) || !bytes.Equal(img, e.image()) {
		t.Fatalf("lettura: %v", err)
	}
	if !strings.Contains(note, "GRUPPO 3") || !strings.Contains(note, "2") || strings.Contains(note, "firmware") {
		t.Errorf("lettura: nota %q", note)
	}
	wantTx := append(append(append(uvproFrame(4, []byte{3}), uvproFrame(20, nil)...), uvproFrame(73, []byte{2})...), uvproFrame(13, []byte{0})...)
	if !bytes.HasPrefix(tx, wantTx) || len(tx) != 9+8+9+30*9 || len(uvproWrites(tx)) != 0 {
		t.Errorf("lettura: %d byte inviati, inizio %x", len(tx), tx[:min(len(tx), 40)])
	}
	// avvisi non richiesti della radio (con il byte di controllo): stessa immagine
	e = mk()
	e.stray = true
	if img, _, err, _ = read(e); err != nil || !bytes.Equal(img, base) {
		t.Errorf("lettura con avvisi della radio: %v", err)
	}
	// la radio non dà lo stato: si legge lo stesso, senza il gruppo
	e = mk()
	e.noState = true
	img, note, err, _ = read(e)
	if err != nil || img[uvproGroupAt] != 0xFF || note != "" || !bytes.Equal(img, e.image()) {
		t.Errorf("lettura senza stato: %v, nota %q", err, note)
	}
	// posizioni che la radio non dà, o di un altro tipo
	e = mk()
	e.missing[8] = true
	e.recs[9] = make([]byte, 40)
	img, _, err, _ = read(e)
	if err != nil || f.rec(img, 8)[0] != uvproNone || f.rec(img, 9)[0] != uvproForeign || f.Get(m, img, 8) != nil || f.Get(m, img, 9) != nil {
		t.Errorf("posizioni mancanti: %v", err)
	}
	part := img
	// firmware non provato: si legge (con l'avviso), non si scrive
	e = mk()
	binary.BigEndian.PutUint16(e.info[4:], 150)
	img, note, err, _ = read(e)
	if err != nil || !strings.Contains(note, "150") || !strings.Contains(note, "147") {
		t.Errorf("lettura con un altro firmware: %v, nota %q", err, note)
	}
	mod := append([]byte{}, base...)
	for i, c := range map[int]*Chan{
		0:  {Rx: 145525000, Tx: 145525000, Power: 0, Scan: true, Name: "S21"},
		1:  nil,
		20: {Rx: 145787500, Tx: 145187500, TxTone: "82.5", Power: 1, Scan: true, Name: "R7a"},
	} {
		if err := f.Set(m, mod, i, c); err != nil {
			t.Fatal(err)
		}
	}
	e = mk()
	binary.BigEndian.PutUint16(e.info[4:], 150)
	if err, tx = write(e, mod); err == nil || !strings.Contains(err.Error(), "150") || e.writes != 0 || e.reads != 0 || len(uvproWrites(tx)) != 0 {
		t.Errorf("scrittura con un altro firmware: %v (%d scritture)", err, e.writes)
	}
	// la radio su un altro gruppo: non si scrive
	e = mk()
	e.group = 4
	if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "gruppo 4") || e.writes != 0 || e.reads != 0 {
		t.Errorf("scrittura con un altro gruppo attivo: %v", err)
	}
	// scrittura: solo i canali cambiati, ognuno riletto prima e dopo
	e = mk()
	if err, tx = write(e, mod); err != nil {
		t.Fatalf("scrittura: %v", err)
	}
	if !bytes.Equal(e.image(), mod) || e.writes != 3 || e.reads != 30+6 {
		t.Errorf("scrittura: %d canali scritti, %d letture", e.writes, e.reads)
	}
	if w := uvproWrites(tx); len(w) != 3 || !bytes.Equal(w[1], uvproEmpty(1)) || !bytes.Equal(w[0], f.ch(mod, 0)) || !bytes.Equal(w[2], f.ch(mod, 20)) {
		t.Errorf("scrittura: canali inviati %x", w)
	}
	// niente da cambiare: niente scritto
	e = mk()
	if err, _ = write(e, base); err != nil || e.writes != 0 || e.reads != 30 {
		t.Errorf("scrittura senza modifiche: %v, %d scritture", err, e.writes)
	}
	// immagine letta senza gruppo: si scrive (non si può controllare)
	nog := append([]byte{}, mod...)
	nog[uvproGroupAt] = 0xFF
	e = mk()
	e.group = 4
	if err, _ = write(e, nog); err != nil || e.writes != 3 {
		t.Errorf("scrittura senza gruppo noto: %v", err)
	}
	// posizioni che la radio non aveva dato restano come sono, le altre si scrivono
	pm := append([]byte{}, part...)
	if err := f.Set(m, pm, 0, &Chan{Rx: 145525000, Tx: 145525000, Power: 0, Scan: true, Name: "S21"}); err != nil {
		t.Fatal(err)
	}
	e = mk()
	if err, _ = write(e, pm); err != nil || e.writes != 1 || !bytes.Equal(e.recs[8], f.ch(base, 8)) || !bytes.Equal(e.recs[9], f.ch(base, 9)) {
		t.Errorf("scrittura con posizioni non lette: %v, %d scritture", err, e.writes)
	}
	// la radio rifiuta un canale: si rimette quello di prima
	e = mk()
	e.refuse = 2
	err, _ = write(e, mod)
	if err == nil || !strings.Contains(err.Error(), "rifiutato la scrittura del canale 1") || !strings.Contains(err.Error(), "rimesso e verificato") ||
		!bytes.Equal(e.recs[1], f.ch(base, 1)) || !bytes.Equal(e.recs[0], f.ch(mod, 0)) || !bytes.Equal(e.recs[20], f.ch(base, 20)) {
		t.Errorf("canale rifiutato: %v", err)
	}
	// la radio salva un canale diverso da quello inviato: errore e ripristino
	e = mk()
	e.flipAt = 1
	err, _ = write(e, mod)
	if err == nil || !strings.Contains(err.Error(), "canale 0 riletto") || !strings.Contains(err.Error(), "rimesso e verificato") || !bytes.Equal(e.recs[0], f.ch(base, 0)) || e.writes != 2 {
		t.Errorf("canale salvato diverso: %v (%d scritture)", err, e.writes)
	}
	// …e se nemmeno il ripristino riesce lo dice
	e = mk()
	e.flipAt, e.stuck = 1, true
	err, _ = write(e, mod)
	if err == nil || !strings.Contains(err.Error(), "NON è stato possibile") || e.writes != 2 {
		t.Errorf("ripristino non riuscito: %v", err)
	}
	// conferma con un altro numero di canale
	e = mk()
	e.ackOff = true
	if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "confermato il canale 1 al posto del 0") {
		t.Errorf("conferma sbagliata: %v", err)
	}
	// il canale sulla radio nel frattempo è diventato DMR: non si tocca
	e = mk()
	e.recs[0][5] |= 0x80
	if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "DMR") || e.writes != 0 {
		t.Errorf("canale DMR sulla radio: %v", err)
	}
	// radio muta a metà e porta che non è quella della radio
	e = mk()
	e.mute = 12
	if img, _, err, _ = read(e); err == nil || img != nil || !strings.Contains(err.Error(), "non risponde") {
		t.Errorf("radio muta a metà: %v", err)
	}
	e = mk()
	e.garbage = true
	if img, _, err, _ = read(e); err == nil || img != nil || !strings.Contains(err.Error(), "Bluetooth") {
		t.Errorf("porta sbagliata: %v", err)
	}
	// immagine non valida: niente inviato
	if err, tx = write(mk(), mod[:len(mod)-1]); err == nil || len(tx) != 0 {
		t.Errorf("immagine corta: %v, %d byte inviati", err, len(tx))
	}
}
