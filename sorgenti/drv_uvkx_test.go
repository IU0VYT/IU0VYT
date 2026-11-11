// IU0VYT Flash - gestione canali, ponti e firmware di radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva

package main

import "testing"

// 4.16, UV-KX: differenze volute dal driver per CHIRP.
// La modulazione resta quella del canale se la frequenza non cambia ed è AM nella
// banda aerea per un canale nuovo (CHIRP scrive quella scelta nella sua tabella:
// le prove gli passano sempre FM); la taratura non si scrive mai.
func (f uvkxFam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	for i := 0; i < m.Slots; i++ {
		a := f.at(i) + 25
		ref[a] = ref[a]&0xF0 | out[a]&0x0F
	}
}

func (uvkxFam) chirpOwnWrite() bool { return true }

// canali scritti e riletti uguali; cancellazione; EEPROM mai scritta
func TestUvkxChannels(t *testing.T) {
	m := genByID["qs-uvk5-uvkx"]
	f := m.Fam.(uvkxFam)
	for _, fill := range []byte{0x00, 0xFF} {
		img := make([]byte, uvk5MemSize)
		for i := range img {
			img[i] = fill
		}
		cal := append([]byte{}, img[uvkxCal:]...)
		want := []*Chan{
			{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Name: "R0 Monte", Scan: true},
			{Rx: 433500000, Tx: 433500000, TxTone: "D754I", RxTone: "D023N", Name: "DCS", Power: 2, Scan: true},
			{Rx: 118100000, Tx: 118100000, TxOff: true, Name: "TWR", Power: 1, Scan: true},
			{Rx: 145500000, Tx: 435500000, RxTone: "123.0", Name: "SPLIT", Scan: true},
		}
		for i, c := range want {
			if err := f.Set(m, img, i*50, c); err != nil {
				t.Fatal(err)
			}
			g := f.Get(m, img, i*50)
			if g == nil || g.Rx != c.Rx || g.TxOff != c.TxOff || (!c.TxOff && g.Tx != c.Tx) || g.TxTone != c.TxTone || g.RxTone != c.RxTone ||
				g.Name != c.Name || g.Power != c.Power {
				t.Fatalf("riempimento %02X, canale %d: scritto %+v, riletto %+v", fill, i*50, c, g)
			}
		}
		// banda aerea: AM (1) nel mezzo byte basso; canale normale: FM (0)
		if img[f.at(100)+25]&0x0F != 1 || img[f.at(0)+25]&0x0F != 0 {
			t.Errorf("modulazione dei canali nuovi: %02X %02X", img[f.at(100)+25], img[f.at(0)+25])
		}
		if n := int(img[f.at(50)+22]) | int(img[f.at(50)+23])<<8; n != 51 {
			t.Errorf("numero del canale nel record: %d", n)
		}
		if err := f.Set(m, img, 0, nil); err != nil || f.Get(m, img, 0) != nil {
			t.Error("canale non cancellato")
		}
		for i := 0; i < m.Slots; i++ {
			if i%50 != 0 || i == 0 || i > 150 {
				if f.Get(m, img, i) != nil {
					t.Fatalf("canale %d non vuoto", i)
				}
			}
		}
		for k, b := range img[uvkxCal:] {
			if b != cal[k] {
				t.Fatal("taratura toccata")
			}
		}
	}
	if err := f.Set(m, make([]byte, uvk5MemSize), 0, &Chan{Rx: 145000005, Tx: 145000005}); err == nil {
		t.Error("frequenza non multipla di 10 Hz accettata")
	}
}
