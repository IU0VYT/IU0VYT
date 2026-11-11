//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"
	"testing"
)

// 4.18, firmware R3b0rn di fagci: disposizione dei canali come nei sorgenti del
// firmware (settings.h, channels.c)
func TestFagciLayout(t *testing.T) {
	m := genByID["qs-uvk5-fagci"]
	f := m.Fam.(fagciFam)
	// canali per tipo di EEPROM: (fine − 0x52F) / 23, al massimo 1024; la correzione
	// SSB (15832 byte) sta in fondo solo se la EEPROM la contiene
	want := map[int][2]int{2: {8192, 298}, 3: {16384, 654}, 4: {32768 - 15832, 678}, 5: {65536 - 15832, 1024}, 6: {131072 - 15832, 1024}, 7: {262144 - 15832, 1024}}
	for typ, w := range want {
		img := fagciBlank(typ)
		if end := int(binary.LittleEndian.Uint32(img[20:])); end != w[0] || f.count(img) != w[1] {
			t.Errorf("EEPROM tipo %d: fine %d, %d canali (attesi %d, %d)", typ, end, f.count(img), w[0], w[1])
		}
		if !f.Match(m, img) || f.Get(m, img, 0) != nil || f.Get(m, img, w[1]) != nil || !f.lockedSlotImg(m, img, w[1]) || f.lockedSlotImg(m, img, w[1]-1) {
			t.Errorf("EEPROM tipo %d: immagine vuota non riconosciuta", typ)
		}
		if err := f.Set(m, img, w[1], &Chan{Rx: 145500000, Tx: 145500000}); err == nil {
			t.Errorf("EEPROM tipo %d: canale oltre la memoria accettato", typ)
		}
	}
	if fagciStart != 0x52F {
		t.Fatalf("inizio dei canali 0x%X", fagciStart)
	}
	img := fagciBlank(2)
	c := &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D023I", Name: "R0 Monte", Power: 1, Scan: true, Narrow: true}
	if err := f.Set(m, img, 0, c); err != nil {
		t.Fatal(err)
	}
	// canale 0 = ultimi 23 byte della EEPROM. Ricezione 14560000 (×10 Hz) con DCS
	// invertito (tipo 3, codice 0); trasmissione 14500000 con subtono (tipo 1,
	// codice 8 = 88,5 Hz); nome; liste 1; FM, stretta, potenza media; ricevitore automatico
	rec := img[len(img)-fagciRec:]
	wantRec := []byte{0x00, 0x2B, 0xDE, 0x18, 0x00, 0xA0, 0x40, 0xDD, 0x08, 0x08}
	wantRec = append(wantRec, []byte("R0 Monte\x00\x00")...)
	wantRec = append(wantRec, 0x01, 2<<6|1<<4, 0x03)
	if !bytes.Equal(rec, wantRec) {
		t.Fatalf("record del canale 0:\n% X\natteso\n% X", rec, wantRec)
	}
	if g := f.Get(m, img, 0); g == nil || *g != *c {
		t.Fatalf("riletto %+v", g)
	}
	// il canale 1 sta subito prima
	if err := f.Set(m, img, 1, &Chan{Rx: 433500000, Tx: 433500000, Name: "U", Power: 0}); err != nil {
		t.Fatal(err)
	}
	if r := img[len(img)-2*fagciRec : len(img)-fagciRec]; r[10] != 'U' || r[21] != 3<<6 || r[20] != 0 {
		t.Fatalf("record del canale 1: % X", r)
	}
	// senza nome: la frequenza (per il firmware un canale senza nome non esiste)
	if err := f.Set(m, img, 2, &Chan{Rx: 118100000, Tx: 118100000}); err != nil {
		t.Fatal(err)
	}
	r2 := f.rec(img, 2)
	if string(r2[10:19]) != "118.10000" || r2[19] != 0 || r2[21]&0x0F != 1 {
		t.Fatalf("canale senza nome o banda aerea: % X", r2)
	}
	if g := f.Get(m, img, 2); g == nil || g.Name != "" || g.Rx != 118100000 {
		t.Fatalf("canale senza nome riletto %+v", g)
	}
	// modulazione, ricevitore, liste e lista nera si conservano; larga resta «25k+»
	r2[20], r2[21], r2[22] = 0x84, 3<<6|3<<4|2, 1
	if err := f.Set(m, img, 2, &Chan{Rx: 118100000, Tx: 118100000, Name: "TWR", Power: 2, Scan: true}); err != nil {
		t.Fatal(err)
	}
	if r2[20] != 0x84 || r2[21] != 1<<6|3<<4|2 || r2[22] != 1 {
		t.Fatalf("campi non conservati: % X", r2[20:])
	}
	if err := f.Set(m, img, 2, &Chan{Rx: 118100000, Tx: 118100000, Name: "TWR", Power: 2}); err != nil || r2[20] != 0x80 {
		t.Fatalf("liste di scansione: %02X %v", r2[20], err)
	}
	// cancellazione come CHANNELS_Delete
	if err := f.Set(m, img, 0, nil); err != nil || f.Get(m, img, 0) != nil {
		t.Fatal("canale non cancellato")
	}
	if rec := img[len(img)-fagciRec:]; !bytes.Equal(rec[:21], make([]byte, 21)) || rec[21] != 0x40 || rec[22] != 3 {
		t.Fatalf("canale cancellato: % X", rec)
	}
	// frequenza di trasmissione a zero (canali scritti da altri programmi): resta
	// finché il canale non cambia frequenza, poi si scrive per intero
	r5 := f.rec(img, 5)
	fagciPutF(r5[0:], 145500000, 0, 0)
	copy(r5[10:], "ZERO")
	if g := f.Get(m, img, 5); g == nil || g.Tx != 145500000 {
		t.Fatalf("trasmissione a zero letta %+v", g)
	}
	if err := f.Set(m, img, 5, &Chan{Rx: 145500000, Tx: 145500000, Name: "ZERO2", Power: 2}); err != nil || binary.LittleEndian.Uint32(r5[5:])&0x07FFFFFF != 0 {
		t.Fatalf("trasmissione a zero non conservata: % X", r5)
	}
	if err := f.Set(m, img, 5, &Chan{Rx: 145500000, Tx: 144900000, Name: "ZERO2", Power: 2}); err != nil || binary.LittleEndian.Uint32(r5[5:])&0x07FFFFFF != 14490000 {
		t.Fatalf("trasmissione non scritta: % X", r5)
	}
	for _, bad := range []*Chan{{Rx: 145500005, Tx: 145500005}, {Rx: 100000, Tx: 100000}, {Rx: 145500000, Tx: 1400000000}, {Rx: 145500000, Tx: 145500000, TxTone: "60.0"}} {
		if err := f.Set(m, fagciBlank(2), 0, bad); err == nil {
			t.Errorf("canale non valido accettato: %+v", bad)
		}
	}
}

func fagciPipe(t *testing.T, e *fagciEmu) func() {
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	return func() { openPortFunc = old }
}

// lettura e scrittura con la radio simulata: si scrive solo la zona dei canali
func TestFagciClone(t *testing.T) {
	m := genByID["qs-uvk5-fagci"]
	f := m.Fam.(fagciFam)
	for _, typ := range []int{2, 3, 4, 5} {
		base := fagciBlank(typ)
		for k := fagciHdr + 21; k < fagciChanAt; k++ {
			base[k] = byte(k * 7) // VFO e preset qualsiasi
		}
		n := f.count(base)
		for _, i := range []int{0, 1, n / 2, n - 1} {
			if err := f.Set(m, base, i, &Chan{Rx: 145000000 + int64(i)*12500, Tx: 145000000 + int64(i)*12500, Name: "CH", Scan: true}); err != nil {
				t.Fatal(err)
			}
		}
		e := f.emu(m, base).(*fagciEmu)
		before := append([]byte{}, e.eep...)
		restore := fagciPipe(t, e)
		got, note, err := f.Download(context.Background(), m, "COM7", noProg)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, base) || !strings.Contains(note, "R3b0rn") {
			t.Fatalf("tipo %d: lettura diversa dalla memoria (%s)", typ, note)
		}
		want := append([]byte{}, got...)
		edits := map[int]*Chan{0: nil, 2: {Rx: 433500000, Tx: 438500000, TxTone: "123.0", Name: "RPT", Power: 0, Scan: true}, n - 1: {Rx: 446006250 - 6250, Tx: 446000000, Name: "LAST"}}
		for i, c := range edits {
			if err := f.Set(m, want, i, c); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := f.Upload(context.Background(), m, "COM7", want, noProg); err != nil {
			t.Fatal(err)
		}
		restore()
		if !bytes.Equal(e.image(), want) {
			t.Fatalf("tipo %d: memoria della radio diversa da quella scritta", typ)
		}
		end, _ := fagciGeom(len(e.eep))
		lo := end - n*fagciRec
		if !bytes.Equal(e.eep[:lo], before[:lo]) || !bytes.Equal(e.eep[end:], before[end:]) {
			t.Fatalf("tipo %d: scritto fuori dalla zona dei canali", typ)
		}
		if len(e.log) == 0 {
			t.Fatal("nessuna scrittura")
		}
		for _, w := range e.log {
			if w[0] < lo || w[0]+w[1] > end || w[1]%8 != 0 || w[1] == 0 {
				t.Fatalf("tipo %d: scrittura a 0x%X di %d byte fuori dalla zona dei canali (0x%X-0x%X)", typ, w[0], w[1], lo, end)
			}
		}
	}
}

// versioni del firmware o memorie diverse da quelle conosciute: nessuna scrittura
func TestFagciRefuse(t *testing.T) {
	m := genByID["qs-uvk5-fagci"]
	f := m.Fam.(fagciFam)
	img := fagciBlank(2)
	if err := f.Set(m, img, 0, &Chan{Rx: 145500000, Tx: 145500000, Name: "A"}); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		prep func(e *fagciEmu)
		want string
	}{
		{"firmware vecchio", func(e *fagciEmu) { e.old16 = true }, "non risponde ai comandi conosciuti"},
		{"numero dei preset", func(e *fagciEmu) { e.eep[9] = 29 }, "29 preset"},
		{"memoria non preparata", func(e *fagciEmu) { e.eep[0] = 0 }, "non è stata preparata"},
		{"memoria di un'altra dimensione", func(e *fagciEmu) { e.eep[0] = 3 }, "8 KiB"},
	}
	for _, c := range cases {
		e := f.emu(m, fagciBlank(2)).(*fagciEmu)
		if c.name == "memoria di un'altra dimensione" {
			e = f.emu(m, fagciBlank(3)).(*fagciEmu)
		}
		c.prep(e)
		restore := fagciPipe(t, e)
		_, err := f.Upload(context.Background(), m, "COM7", img, noProg)
		restore()
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: errore atteso «%s», ottenuto %v", c.name, c.want, err)
		}
		if len(e.log) != 0 {
			t.Errorf("%s: la radio è stata scritta", c.name)
		}
	}
	if _, err := f.Upload(context.Background(), m, "COM7", img[:len(img)-1], noProg); err == nil {
		t.Error("immagine tronca accettata")
	}
}

// 4.18, Radtel RT-400: usa il saluto delle RT-470 e non ha un identificativo
// conosciuto; la radio che si identifica come un altro modello viene rifiutata
func TestRT400RefusesOtherModels(t *testing.T) {
	m := genByID["rt-rt400"]
	for _, id := range []string{"rt-rt470", "rt-rt470x", "rt-rt630", "rt-rt495", "hy-hi8811"} {
		other := genByID[id]
		emu := genEmuFor(other, noChirpBase(t, m))
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			a, b := newPipe()
			a.baud = baud
			go emu.serve(b)
			return a, nil
		}
		_, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
		openPortFunc = old
		if err == nil || !strings.Contains(err.Error(), "altro modello") {
			t.Errorf("%s: errore atteso, ottenuto %v", id, err)
		}
	}
	// la RT-330 ha un saluto solo suo e 512 canali a 115200 baud
	if r := genByID["rt-rt330"]; r.Slots != 512 || r.Baud != 115200 || string(r.Opt.(*jc8810Opt).magic) != "PROGRAMYF68U" || !r.NoChirp {
		t.Errorf("RT-330: %d canali, %d baud", r.Slots, r.Baud)
	}
}
