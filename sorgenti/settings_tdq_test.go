// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"testing"
)

// Riga delle impostazioni scritta a mano: voce inglese (2) e guadagno 5 → 52; squelch 4;
// TOT indice 2 = 30 s; preavviso 5 s; VOX spento; sensibilità indice 1 = «2»; ritardo
// indice 2 = 1,5 s; suono dei tasti e risparmio accesi; tasti 9 (Bluetooth), 14 (monitor),
// 12 (riduzione del rumore), 0.
func tdqSetImg() []byte {
	img := bytes.Repeat([]byte{0xFF}, tdqSize)
	copy(img[tdqSetAddr:], []byte{0x52, 4, 2, 5, 0, 0x51, 2, 0, 1, 1, 1, 9, 14, 12, 0, 1})
	return img
}

func TestTdqSettings(t *testing.T) {
	img := tdqSetImg()
	for id, n := range map[string]int{"td-m11plus": 14, "td-712plus": 14, "td-m2pro": 14, "td-frs": 12, "td-pmr": 12, "td-m11plus-22": 12, "td-777s-16": 14} {
		if !settingsAvailable(id) {
			t.Fatalf("%s: impostazioni non disponibili", id)
		}
		tab := settingsFor(id, img)
		if tab == nil || len(tab.Settings) != n {
			t.Fatalf("%s: tabella assente o con un numero di voci inatteso", id)
		}
		if settingsFor(id, bytes.Repeat([]byte{0xFF}, tdqSize)) != nil || settingsFor(id, make([]byte, tdqSize)) != nil {
			t.Fatalf("%s: immagine vuota, attesa nessuna tabella", id)
		}
	}
	tab := settingsFor("td-m11plus", img)
	for key, want := range map[string]string{"squelch": "4", "tot": "30 s", "tot_alert": "5 s", "vox": "false", "vox_level": "2", "vox_delay": "1,5 s",
		"mic_gain": "5", "save": "true", "voice": "Inglese", "beep": "true", "key1_short": "Bluetooth", "key1_long": "Monitor finché tieni premuto",
		"key2_short": "Riduzione del rumore", "key2_long": "Nessuna"} {
		if got, _ := tdcGet(t, tab, img, key); got != want {
			t.Errorf("%s = %q, atteso %q", key, got, want)
		}
	}
	// TD-M2 Pro: il valore 11 è la copia via radio, il 14 non c'è
	m2 := settingsFor("td-m2pro", img)
	img[tdqSetAddr+11] = 11
	if got, _ := tdcGet(t, m2, img, "key1_short"); got != "Copia via radio" {
		t.Errorf("TD-M2 Pro, tasto 11 = %q", got)
	}
	for i := range m2.Settings {
		if d := &m2.Settings[i]; d.Key == "key1_long" {
			if _, err := readSetting(d, img); err == nil {
				t.Error("TD-M2 Pro: funzione 14 letta senza errore")
			}
		}
	}
	// scrittura: cambia solo il byte (o i bit) della voce
	img = tdqSetImg()
	w := func(key, opt string) {
		t.Helper()
		_, d := tdcGet(t, tab, img, key)
		v := opt
		for i, o := range d.ItOpts {
			if o == opt {
				v = d.Options[i]
			}
		}
		if err := writeSetting(d, img, v); err != nil {
			t.Fatalf("%s=%s: %v", key, opt, err)
		}
	}
	w("squelch", "0")
	w("tot", "600 s")
	w("tot_alert", "OFF")
	w("vox", "true")
	w("vox_level", "5")
	w("vox_delay", "3 s")
	w("mic_gain", "9")
	w("voice", "OFF")
	w("beep", "false")
	w("key1_short", "Monitor (squelch aperto)")
	w("key2_long", "Riduzione del rumore")
	want := []byte{0x90, 0, 40, 0, 1, 0x54, 5, 0, 0, 1, 1, 14, 14, 12, 12, 1}
	if got := img[tdqSetAddr : tdqSetAddr+16]; !bytes.Equal(got, want) {
		t.Errorf("riga scritta: % X\n       attesa: % X", got, want)
	}
	rest := tdqSetImg()
	copy(rest[tdqSetAddr:], want)
	if !bytes.Equal(img, rest) {
		t.Error("toccati byte fuori dalla riga delle impostazioni")
	}
}
