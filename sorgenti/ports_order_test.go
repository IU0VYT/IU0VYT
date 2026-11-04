// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.5: porte in ordine progressivo su ogni sistema

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestPortOrder(t *testing.T) {
	for _, want := range [][]string{
		{"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM9", "COM10", "COM11", "COM20", "COM100", "COM256"},
		{"COM3", "com4", "COM12", "/dev/ttyACM0", "/dev/ttyACM2", "/dev/ttyACM10", "/dev/ttyS0", "/dev/ttyS1", "/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyUSB2", "/dev/ttyUSB10", "/dev/ttyUSB11"},
		{"/dev/cu.usbserial-110", "/dev/cu.usbserial-140", "/dev/cu.usbserial-1110", "/dev/cu.wchusbserial1420", "/dev/cu.wchusbserial14210"},
	} {
		for seed := int64(1); seed <= 20; seed++ {
			got := append([]string{}, want...)
			rand.New(rand.NewSource(seed)).Shuffle(len(got), func(i, j int) { got[i], got[j] = got[j], got[i] })
			sort.Slice(got, func(i, j int) bool { return portLess(got[i], got[j]) })
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ordine: %v\natteso: %v", got, want)
			}
		}
	}
	for _, c := range [][2]string{{"COM2", "COM10"}, {"COM02", "COM3"}, {"COM9", "COMX"}, {"COM9", "/dev/ttyUSB0"}} {
		if !portLess(c[0], c[1]) || portLess(c[1], c[0]) {
			t.Fatalf("%s deve venire prima di %s", c[0], c[1])
		}
	}
	if portLess("COM5", "COM5") {
		t.Fatal("ordine non stretto")
	}
}
