//go:build !windows

// IU0VYT Flash - gestione canali, ponti e firmware di radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"
	"testing"
)

func ogdTestImg() []byte {
	img := bytes.Repeat([]byte{0xFF}, ogdSize)
	for b := 0; b < 8; b++ {
		copy(img[ogdBank(b):], make([]byte, 16))
	}
	copy(img[rdtyZoneTab:], make([]byte, 32))
	return img
}

// byte del canale come nella descrizione di qdmr (OpenGD77BaseCodeplug::ChannelElement)
func TestOgd77Channel(t *testing.T) {
	m := genByID["ogd77"]
	f := m.Fam.(ogdFam)
	img := ogdTestImg()
	c := &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D023I", Name: "R0 Monte", Power: 1, Scan: false}
	if err := f.Set(m, img, 0, c); err != nil {
		t.Fatal(err)
	}
	r := img[0x3780+16:][:56]
	if string(r[:8]) != "R0 Monte" || r[8] != 0xFF || binary.LittleEndian.Uint32(r[16:]) != 0x14560000 || binary.LittleEndian.Uint32(r[20:]) != 0x14500000 {
		t.Fatalf("nome o frequenze: % X", r[:24])
	}
	if r[24] != 0 || r[25] != 6 || binary.LittleEndian.Uint16(r[34:]) != 0x0885 || binary.LittleEndian.Uint16(r[32:]) != 0xC023 {
		t.Fatalf("modo, potenza o toni: % X", r[24:36])
	}
	if r[51] != 1<<1|1<<4 || img[0x3780] != 1 {
		t.Fatalf("byte 51 %02X, bitmap %02X", r[51], img[0x3780])
	}
	// canale 129: primo della flash (posizione nell'immagine 0xB000 + 0x1B0)
	d := &Chan{Rx: 430000000, Tx: 430000000, Dig: true, Name: "DMR", Scan: true}
	if err := f.Set(m, img, 128, d); err != nil {
		t.Fatal(err)
	}
	r = img[ogdEep+0x1B0+16:][:56]
	if img[ogdEep+0x1B0] != 1 || r[24] != 1 || r[25] != 9 || r[44] != 1 || r[51] != 0 || binary.LittleEndian.Uint16(r[32:]) != 0xFFFF {
		t.Fatalf("canale DMR: % X", r[24:56])
	}
	// potenza «quella della radio» (0): resta finché il livello non cambia
	r[25] = 0
	d2 := *f.Get(m, img, 128)
	d2.Name = "DMR 2"
	if err := f.Set(m, img, 128, &d2); err != nil || r[25] != 0 {
		t.Fatalf("potenza della radio non conservata: %d (%v)", r[25], err)
	}
	d2.Power = 3
	if err := f.Set(m, img, 128, &d2); err != nil || r[25] != 1 {
		t.Fatalf("potenza minima: %d", r[25])
	}
	// cancellazione: il canale esce dalle zone
	z := img[rdtyZoneTab+32:]
	img[rdtyZoneTab] |= 1
	copy(z, make([]byte, 16+2*ogdZoneMem))
	copy(z, "Zona")
	binary.LittleEndian.PutUint16(z[16:], 1)
	binary.LittleEndian.PutUint16(z[18:], 129)
	if err := f.Set(m, img, 0, nil); err != nil || f.Get(m, img, 0) != nil {
		t.Fatal("canale non cancellato")
	}
	if binary.LittleEndian.Uint16(z[16:]) != 129 || binary.LittleEndian.Uint16(z[18:]) != 0 {
		t.Fatalf("zona dopo la cancellazione: % X", z[16:22])
	}
}

// lettura, scrittura dei soli blocchi cambiati, rifiuto della radio dell'altra famiglia
func TestOgd77Clone(t *testing.T) {
	for _, id := range []string{"ogd77", "ogd77-uv380"} {
		m := genByID[id]
		f := m.Fam.(ogdFam)
		radio := ogdTestImg()
		for i := range radio { // contenuto qualsiasi fuori da canali e zone
			if !ogdWritable(i) {
				radio[i] = byte(i * 7)
			}
		}
		for _, r := range [][2]int{{0, 0xE0}, {0x6000, 0x7500}} {
			for i := r[0]; i < r[1]; i++ {
				radio[i] = 0xFF // parti della EEPROM che non si leggono
			}
		}
		for i := 0; i < 1024; i += 37 {
			if err := f.Set(m, radio, i, &Chan{Rx: 145000000 + int64(i)*12500, Tx: 145000000 + int64(i)*12500, Name: "CH", Scan: true}); err != nil {
				t.Fatal(err)
			}
		}
		e := f.emu(m, radio).(*ogdEmu)
		var img []byte
		emuRun(e, func() {
			var err error
			if img, _, err = f.Download(context.Background(), m, "EMU", func(int, int, string) {}); err != nil {
				t.Fatal(err)
			}
		})
		if !bytes.Equal(img, radio) {
			t.Fatalf("%s: lettura diversa dalla memoria della radio", id)
		}
		mod := append([]byte{}, img...)
		_ = f.Set(m, mod, 5, &Chan{Rx: 433500000, Tx: 433500000, Name: "NUOVO", Scan: true})
		_ = f.Set(m, mod, 900, &Chan{Rx: 438000000, Tx: 430400000, Dig: true, Name: "DMR", Scan: true})
		_ = f.Set(m, mod, 37, nil)
		mod[0x100] ^= 0xFF          // impostazioni: non devono arrivare alla radio
		mod[ogdEep+0x12000] ^= 0xFF // contatti: lo stesso
		emuRun(e, func() {
			if _, err := f.Upload(context.Background(), m, "EMU", mod, func(int, int, string) {}); err != nil {
				t.Fatal(err)
			}
		})
		if len(e.bad) != 0 {
			t.Fatalf("%s: scritture fuori da canali e zone: %v", id, e.bad)
		}
		want := append([]byte{}, mod...)
		want[0x100], want[ogdEep+0x12000] = img[0x100], img[ogdEep+0x12000]
		if got := e.image(); !bytes.Equal(got, want) {
			t.Fatalf("%s: memoria finale diversa: %s", id, diffRanges(got, want))
		}
		if e.wrote == 0 || e.wrote > 12 {
			t.Errorf("%s: blocchi scritti %d (attesi pochi)", id, e.wrote)
		}
		// la radio dell'altra famiglia è rifiutata con il nome della scheda giusta
		other := genByID[map[string]string{"ogd77": "ogd77-uv380", "ogd77-uv380": "ogd77"}[id]]
		eo := f.emu(other, radio).(*ogdEmu)
		emuRun(eo, func() {
			_, _, err := f.Download(context.Background(), m, "EMU", func(int, int, string) {})
			if err == nil || !strings.Contains(err.Error(), other.Name) {
				t.Errorf("%s: errore atteso con la scheda giusta, ottenuto %v", id, err)
			}
		})
	}
}
