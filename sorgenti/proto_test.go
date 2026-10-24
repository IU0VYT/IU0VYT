//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func noProg(int, int, string) {}

// il simulatore lavora in parallelo: attende che abbia finito
func waitFor(f func() bool) {
	for i := 0; i < 200 && !f(); i++ {
		time.Sleep(5 * time.Millisecond)
	}
}

// collega i protocolli a una radio simulata
func withEmu(t *testing.T, serve func(Port) error) func() {
	t.Helper()
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go serve(b)
		return a, nil
	}
	return func() { openPortFunc = old }
}

func k67Sample(t *testing.T) []byte {
	return xlcImageForRadio(mustRead(t, sampleXLC))
}

func TestK67ReadWrite(t *testing.T) {
	file := mustRead(t, sampleXLC)
	radio := &k67Emu{mem: append([]byte{}, xlcImageForRadio(file)...), model: "5RH-PRO", band: 0}
	defer withEmu(t, radio.serve)()
	img, info, err := k67Read(context.Background(), "COM1", &k67Opts{}, noProg)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(img, radio.mem) || info.Model != "5RH-PRO" || info.FreqBand != 0 || info.K6 {
		t.Fatalf("lettura errata: %+v", info)
	}
	// il file costruito dalla lettura coincide con il file di esempio, tranne
	// data di salvataggio e voci di zona 150-151 (come il programma originale)
	got := xlcFileFromRadio(img, info.FreqBand, info.K6, time.Date(2026, 10, 17, 17, 15, 0, 0, time.UTC))
	if d := diffRanges(file, got); d != "2 byte diversi: 0x7BBD-0x7BBD (95 -> 00); 0x7BBF-0x7BBF (96 -> 00)" {
		t.Fatalf("file ricostruito: %s", d)
	}
	// scrittura di un file modificato: la radio riceve i dati, tranne i 128 byte iniziali che restano i suoi
	chs := readXLC(file, &Report{})
	chs[0].Name = "SCRITTO"
	mod := writeXLC(file, chs, &Report{})
	radio.mem[0x10] ^= 0x01 // informazioni della radio diverse da quelle del file
	before := append([]byte{}, radio.mem[:128]...)
	if _, err := k67Write(context.Background(), "COM1", xlcImageForRadio(mod), &k67Opts{FreqBand: 0}, noProg); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(radio.mem[:128], before) {
		t.Fatal("informazioni della radio sovrascritte")
	}
	if !bytes.Equal(radio.mem[128:], xlcImageForRadio(mod)[128:]) || radio.writes != 12 {
		t.Fatalf("dati scritti diversi (blocchi %d)", radio.writes)
	}
	back := readXLC(xlcFileFromRadio(radio.mem, 0, false, time.Now()), &Report{})
	if back[0].Name != "SCRITTO" {
		t.Fatal("nome non scritto")
	}
}

func TestK67Errors(t *testing.T) {
	img := k67Sample(t)
	radio := &k67Emu{mem: append([]byte{}, img...), model: "5RH-PRO", band: 3, password: "1234"}
	defer withEmu(t, radio.serve)()
	if _, _, err := k67Read(context.Background(), "COM1", &k67Opts{}, noProg); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("password: %v", err)
	}
	if _, _, err := k67Read(context.Background(), "COM1", &k67Opts{Password: "1234"}, noProg); err != nil {
		t.Fatal(err)
	}
	// banda diversa: nessun blocco scritto
	if _, err := k67Write(context.Background(), "COM1", img, &k67Opts{Password: "1234", FreqBand: 0}, noProg); err == nil || !strings.Contains(err.Error(), "banda") {
		t.Fatalf("banda: %v", err)
	}
	// BF-K6 non letto prima: rifiutato
	radio.model, radio.band = "BF-K6", 0
	if _, err := k67Write(context.Background(), "COM1", img, &k67Opts{Password: "1234"}, noProg); err == nil || !strings.Contains(err.Error(), "BF-K6") {
		t.Fatalf("K6: %v", err)
	}
	if radio.writes != 0 {
		t.Fatal("scritture eseguite nonostante l'errore")
	}
	// rifiuto dei dati
	radio.model, radio.rejectW = "5RH-PRO", true
	if _, err := k67Write(context.Background(), "COM1", img, &k67Opts{Password: "1234"}, noProg); err == nil || !strings.Contains(err.Error(), "rifiutato") {
		t.Fatalf("rifiuto: %v", err)
	}
	// radio spenta
	openPortFunc = func(string, int) (Port, error) { a, _ := newPipe(); return a, nil }
	if _, _, err := k67Read(context.Background(), "COM1", &k67Opts{}, noProg); err == nil || !strings.Contains(err.Error(), "non risponde") {
		t.Fatalf("radio spenta: %v", err)
	}
}

func TestK67Baud115200(t *testing.T) {
	radio := &k67Emu{mem: k67Sample(t), model: "5RH-PRO", answerA: 'x'}
	var port *pipeEnd
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		port = a
		go radio.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	_, info, err := k67Read(context.Background(), "COM1", &k67Opts{}, noProg)
	if err != nil || info.Baud != 115200 || port.baud != 115200 {
		t.Fatalf("cambio velocità: %v %d", err, port.baud)
	}
}

func btSampleImage(t *testing.T) []byte {
	img, err := datToEEPROM(mustRead(t, sampleDAT))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestBtCrypt(t *testing.T) {
	key := []byte("A ES")
	for i := 0; i < 256; i++ {
		d := []byte{byte(i), byte(i), byte(i), byte(i)}
		e := append([]byte{}, d...)
		btCrypt(e, key)
		btCrypt(e, key)
		if !bytes.Equal(d, e) {
			t.Fatalf("cifratura non reversibile per %02x", i)
		}
	}
}

func TestBtReadWrite(t *testing.T) {
	img := btSampleImage(t)
	radio := newBtEmu(img)
	defer withEmu(t, radio.serve)()
	got, info, err := btReadImage(context.Background(), "COM1", noProg, nil)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(func() bool { return radio.ended })
	if !bytes.Equal(got, img) || !radio.ended || info.ID != "BT-11U  V1.00   " {
		t.Fatal("lettura diversa dall'immagine")
	}
	// il .dat ricostruito ha gli stessi canali del file di esempio
	dat, err := eepromToDAT(got, baseDAT)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := readDAT(mustRead(t, sampleDAT), &Report{})
	b, _ := readDAT(dat, &Report{})
	for i := range a {
		if (a[i] == nil) != (b[i] == nil) || (a[i] != nil && (a[i].Rx != b[i].Rx || a[i].Tx != b[i].Tx || a[i].Name != b[i].Name || a[i].TxTone != b[i].TxTone || a[i].RxTone != b[i].RxTone || a[i].Power != b[i].Power || a[i].Narrow != b[i].Narrow || a[i].Scan != b[i].Scan)) {
			t.Fatalf("canale %d: %+v / %+v", i, a[i], b[i])
		}
	}
	// scrittura: blocchi scritti tutti, compreso quello finale a 0xFF
	radio2 := newBtEmu(bytes.Repeat([]byte{0xFF}, btImgSize))
	openPortFunc = func(string, int) (Port, error) { x, y := newPipe(); go radio2.serve(y); return x, nil }
	if _, err := btWriteImage(context.Background(), "COM1", img, noProg, nil); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(radio2.mem[:btImgSize], img) || len(radio2.written) != len(btBlocks()) || !radio2.written[btImgSize] {
		t.Fatalf("scrittura errata: %d blocchi", len(radio2.written))
	}
	for _, x := range radio2.mem[btImgSize:] {
		if x != 0xFF {
			t.Fatal("blocco finale non a 0xFF")
		}
	}
}

// .dat -> memoria -> .dat: tutti i campi dell'esempio restano uguali
func TestBtEEPROMRoundTrip(t *testing.T) {
	src := mustRead(t, sampleDAT)
	img := btSampleImage(t)
	back, err := eepromToDAT(img, src)
	if err != nil {
		t.Fatal(err)
	}
	img2, err := datToEEPROM(back)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(img, img2) {
		t.Fatal("memoria diversa dopo andata e ritorno")
	}
	fa, fb := datFields(t, src), datFields(t, back)
	diff := 0
	for k, v := range fa {
		if fb[k] != v {
			diff++
			if diff < 20 {
				t.Logf("campo %s: %q -> %q", k, v, fb[k])
			}
		}
	}
	if diff > 0 {
		t.Errorf("%d campi diversi", diff)
	}
}

// datFields: tutti i valori del file .dat, per nome
func datFields(t *testing.T, dat []byte) map[string]string {
	t.Helper()
	d, err := ParseNRBF(dat)
	if err != nil {
		t.Fatal(err)
	}
	var br *btRadio
	func() {
		defer func() {
			if e := recover(); e != nil {
				t.Fatal(e)
			}
		}()
		br = btOpen(d)
	}()
	out := map[string]string{}
	dump := func(prefix string, o *nobj) {
		m := d.MetaOf(o.r)
		for i, n := range m.Names {
			v := o.r.Vals[i]
			key := prefix + "." + n
			if v.Prim {
				out[key] = fmt.Sprintf("%x", v.Raw)
				continue
			}
			if r := d.Resolve(v); r != nil && r.T == 6 {
				out[key] = "s:" + r.S
			}
		}
	}
	for i, c := range br.chans {
		dump(fmt.Sprintf("ch%03d", i), c)
	}
	dump("vfoA", br.vfo[0])
	dump("vfoB", br.vfo[1])
	dump("fun", br.fun)
	dump("dtmf", br.dtmf)
	for i := range br.codes.Els {
		out[fmt.Sprintf("dtmf.code%d", i)] = br.code(i)
	}
	dump("mod", br.mod)
	for i, c := range br.modChans {
		dump(fmt.Sprintf("mod%02d", i), c)
	}
	return out
}
