// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bajeton DM-5000: byte dei canali e delle zone come li scrive il programma del
// costruttore (SaveSingle, InitDigital, InitAnalog, SaveSubaudio, salvataggio
// delle zone di «BAJETON DM-5000 CPS» 250726, valori calcolati a mano dal codice).

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDm5000Codec(t *testing.T) {
	m := genByID["bj-dm5000"]
	img := bytes.Repeat([]byte{0xFF}, dm5kSize())
	ff := func(n int) string { return hex.EncodeToString(bytes.Repeat([]byte{0xFF}, n)) }
	cases := []struct {
		i    int
		c    *Chan
		want string
	}{
		// analogico 433.500, TX 88.5, banda stretta, alta, in scansione, «TEST»
		{0, &Chan{Rx: 433500000, Tx: 433500000, TxTone: "88.5", Power: 0, Narrow: true, Scan: true, Name: "TEST"},
			"00010101" + "0000" + "f0779502" + "f0779502" + "7513" + "01" + "00" + "00" + "00" + "000000000000000000000000" + "54455354" + ff(12)},
		// analogico con DCS: RX D754I, TX D023N, banda larga, bassa, escluso dalla scansione
		{1, &Chan{Rx: 145600000, Tx: 145000000, RxTone: "D754I", TxTone: "D023N", Power: 2, Scan: false, Name: "R0 ROMA"},
			"00010100" + "ec31" + "002bde00" + "a040dd00" + "1320" + "00" + "00" + "00" + "80" + "000000000000000000000000" + "523020524f4d41" + ff(9)},
		// digitale DMR, alta, in scansione: valori predefiniti del programma
		{2, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Power: 0, Scan: true, Name: "DMR"},
			"00010000" + "0000" + "f0ca9c02" + "30329102" + "00ff" + "01" + "00" + "ff" + "00" + "0000" + "0000" + "0000" + "0000" + "01000000" + "444d52" + ff(13)},
	}
	for _, k := range cases {
		if err := m.Fam.Set(m, img, k.i, k.c); err != nil {
			t.Fatal(err)
		}
		got := hex.EncodeToString(dm5kFam{}.rec(img, k.i))
		if got != k.want {
			t.Errorf("canale %d:\n got %s\nwant %s", k.i+1, got, k.want)
		}
		g := m.Fam.Get(m, img, k.i)
		if g == nil || g.Rx != k.c.Rx || g.Tx != k.c.Tx || g.TxTone != k.c.TxTone || g.RxTone != k.c.RxTone || g.Name != k.c.Name ||
			g.Dig != k.c.Dig || g.Scan != k.c.Scan || m.powerIndex(g.Power) != m.powerIndex(k.c.Power) || !k.c.Dig && g.Narrow != k.c.Narrow {
			t.Errorf("canale %d riletto diverso: %+v", k.i+1, g)
		}
	}
	// zona 1: i tre canali nuovi, nome predefinito «Zone-001»
	z := dm5kFam{}.zone(img, 0)
	if want := "0303ffff" + hex.EncodeToString([]byte("Zone-001")) + ff(8) + "000001000200" + ff(197*2); hex.EncodeToString(z[:420]) != want {
		t.Errorf("zona 1:\n got %x\nwant %s", z[:420], want)
	}
	// da digitale ad analogico: valori predefiniti analogici
	if err := m.Fam.Set(m, img, 2, &Chan{Rx: 438300000, Tx: 430700000, Power: 0, Scan: true, Name: "DMR"}); err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(dm5kFam{}.rec(img, 2)[:32]); got != "00010100"+"0000"+"f0ca9c02"+"30329102"+"0000"+"01"+"00"+"00"+"00"+"000000000000000000000000" {
		t.Errorf("canale 3 analogico: %s", got)
	}
	// cancellazione: canale a FF e tolto dalla zona (elenco compatto)
	if err := m.Fam.Set(m, img, 1, nil); err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(dm5kFam{}.rec(img, 1)); got != ff(48) {
		t.Errorf("canale 2 cancellato: %s", got)
	}
	if got := (dm5kFam{}).zoneMembers(img, 0); len(got) != 2 || got[0] != 0 || got[1] != 2 || z[0] != 2 || z[1] != 2 {
		t.Errorf("zona 1 dopo la cancellazione: %v (numero %d)", got, z[0])
	}
}
