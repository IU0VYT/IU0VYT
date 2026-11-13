// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"strconv"
	"testing"
)

// tdcImg: immagine di prova con valori scritti a mano (indirizzi della radio + 8)
func tdcImg(t *testing.T, id string) ([]byte, *setTable) {
	t.Helper()
	tab := tdcSettings[id]
	if tab == nil {
		t.Fatalf("%s: impostazioni assenti", id)
	}
	img := make([]byte, tab.Size)
	set := func(addr int, b ...byte) { copy(img[8+addr:], b) }
	set(0xC90, 0, 2, 1, 5, 3, 6)          // tasti: sup. breve, lat.1 breve, lat.2 breve, sup. lungo, lat.1 lungo, lat.2 lungo
	set(0xC98, 0, 0, 0, 2, 0, 1)          // … luminosità = 1 (livello 4)
	set(0xCA0, 0xCA, 0x45, 0x69, 0x84)    // 0xCA2: 0x40 è un bit che il programma non usa
	set(0xCA4, 1, 1, 1, 0x43, 0x35, 4, 2) // VOX 3 + coda; passi A = 10 kHz, B = 25 kHz; squelch 4; TOT 60 s
	set(0xCAB, 0x52, 4, 3, 1, 0x91)       // roger tono 1 + TX 200 + bit 0x02 non usato; 1:4; 15 s; VOX 2 s; solo canali + lampeggio 5 s + allarme in aria
	set(0x1C00, []byte("CIAO")...)
	set(0x1F20, 5)
	set(0x1F28, 6, 1, 3, 0, 0, 0, 0, 5) // italiano, doppio, giallo, ripresa 3,0 s
	return img, tab
}

func tdcGet(t *testing.T, tab *setTable, img []byte, key string) (string, *setDef) {
	t.Helper()
	for i := range tab.Settings {
		if d := &tab.Settings[i]; d.Key == key {
			v, err := readSetting(d, img)
			if err != nil {
				t.Fatalf("%s: %v", key, err)
			}
			if d.Type == "list" {
				n, _ := strconv.Atoi(v)
				return d.ItOpts[n], d
			}
			return v, d
		}
	}
	t.Fatalf("%s: voce assente", key)
	return "", nil
}

func TestTdcpsSettingsRead(t *testing.T) {
	img, tab := tdcImg(t, "td-h9")
	for key, want := range map[string]string{
		"key_top_short": "Nessuna", "key_side1_short": "GNSS (posizione)", "key_side2_short": "Radio FM",
		"key_top_long": "Allarme", "key_side1_long": "Monitor (squelch aperto)", "key_side2_long": "Meteo",
		"bl_level": "4", "tx_led": "true", "rx_led": "true", "dtmf_st": "true", "pri_tx": "Banda selezionata",
		"scan_mode": "CO", "autolock": "false", "beep": "true", "voice": "true",
		"tone_burst": "1750 Hz", "fm_rx": "true", "name_a": "false", "work_a": "VFO e canali",
		"boot": "Immagine", "name_b": "false", "dual_watch": "true", "work_b": "VFO (frequenza)",
		"tail": "true", "vox_level": "3", "step_a": "10 kHz", "step_b": "25 kHz", "squelch": "4", "tot": "60 s",
		"roger": "Tono 1", "tx200": "true", "tx350": "false", "tx500": "false", "save": "1:4", "bl_time": "15 s",
		"vox_delay": "2 s", "only_ch": "true", "breath": "5 s", "alarm": "Trasmesso in aria",
		"boot_msg1": "CIAO", "boot_msg2": "", "mic_gain": "5", "lang": "Italiano", "style": "Doppio",
		"menu_color": "Giallo", "scan_hang": "3,0 s",
	} {
		if got, _ := tdcGet(t, tab, img, key); got != want {
			t.Errorf("%s = %q, atteso %q", key, got, want)
		}
	}
	// sulla pressione breve la funzione 3 non esiste: il valore non va inventato
	img[8+0xC91] = 3
	for i := range tab.Settings {
		if d := &tab.Settings[i]; d.Key == "key_side1_short" {
			if _, err := readSetting(d, img); err == nil {
				t.Error("tasto breve con valore 3: atteso un errore")
			}
		}
	}
}

func TestTdcpsSettingsModels(t *testing.T) {
	want := map[string]struct {
		n, lang int
		top     bool
		k2      string
	}{"td-h3plus": {45, 8, false, "Torcia"}, "td-h7plus": {45, 8, false, "Torcia"}, "td-h8g4": {47, 8, true, "Torcia"},
		"td-h9": {47, 9, true, "GNSS (posizione)"}, "td-638uv": {47, 9, true, "GNSS (posizione)"}}
	for id, w := range want {
		for _, suf := range []string{"", "ham", "gmrs"} {
			k := id + suf
			if !settingsAvailable(k) {
				t.Errorf("%s: impostazioni non disponibili", k)
				continue
			}
			img, tab := tdcImg(t, k)
			if got := settingsFor(k, img); got != tab {
				t.Errorf("%s: tabella non restituita", k)
			}
			if settingsFor(k, make([]byte, tab.Size)) != nil {
				t.Errorf("%s: immagine vuota, attesa nessuna tabella", k)
			}
			if len(tab.Settings) != w.n {
				t.Errorf("%s: %d voci, attese %d", k, len(tab.Settings), w.n)
			}
			top := false
			used := map[int]byte{}
			for i := range tab.Settings {
				d := &tab.Settings[i]
				top = top || d.Key == "key_top_short"
				switch d.Key {
				case "lang":
					if len(d.ItOpts) != w.lang {
						t.Errorf("%s: %d lingue, attese %d", k, len(d.ItOpts), w.lang)
					}
				case "key_side1_short":
					if d.ItOpts[2] != w.k2 || len(d.ItOpts) != 8 {
						t.Errorf("%s: funzioni del tasto breve %v", k, d.ItOpts)
					}
				}
				if d.ItLabel == "" || d.ItDesc == "" || d.Group == "" {
					t.Errorf("%s/%s: testi mancanti", k, d.Key)
				}
				// nessuna voce deve toccare i bit di un'altra
				if d.Enc.Kind == "table" {
					off := int(d.Enc.Fields[0][0].(float64))
					m, _ := strconv.ParseUint(d.Enc.Fields[0][1].(string), 16, 8)
					if used[off]&byte(m) != 0 {
						t.Errorf("%s/%s: bit già usati da un'altra voce", k, d.Key)
					}
					used[off] |= byte(m)
				} else {
					for _, off := range d.Enc.Offs {
						if used[off] != 0 {
							t.Errorf("%s/%s: byte già usato", k, d.Key)
						}
						used[off] = 0xFF
					}
				}
			}
			if top != w.top {
				t.Errorf("%s: tasto superiore = %v", k, top)
			}
		}
	}
}

func TestTdcpsSettingsWrite(t *testing.T) {
	base, tab := tdcImg(t, "td-638uv")
	for i := range tab.Settings {
		d := &tab.Settings[i]
		if d.Enc.Kind != "table" {
			continue
		}
		off := int(d.Enc.Fields[0][0].(float64))
		m64, _ := strconv.ParseUint(d.Enc.Fields[0][1].(string), 16, 8)
		mask := byte(m64)
		vals := []string{"false", "true"}
		if d.Type == "list" {
			vals = d.Options
		}
		for _, v := range vals {
			img := append([]byte{}, base...)
			if err := writeSetting(d, img, v); err != nil {
				t.Fatalf("%s=%s: %v", d.Key, v, err)
			}
			if got, err := readSetting(d, img); err != nil || got != v {
				t.Errorf("%s: scritto %s, riletto %q (%v)", d.Key, v, got, err)
			}
			// cambia soltanto i bit della voce
			if img[off]&^mask != base[off]&^mask {
				t.Errorf("%s=%s: toccati bit di altre voci (%02X → %02X)", d.Key, v, base[off], img[off])
			}
			img[off] = base[off]
			if !bytes.Equal(img, base) {
				t.Errorf("%s=%s: toccati altri byte", d.Key, v)
			}
		}
	}
	// valori scritti a mano
	img := append([]byte{}, base...)
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
	w("key_side1_short", "OD PTT")
	w("key_side2_short", "Tono")
	w("key_top_long", "Monitor (squelch aperto)")
	w("bl_level", "1")
	w("step_a", "8,33 kHz")
	w("step_b", "50 kHz")
	w("scan_mode", "SE")
	w("roger", "Tono 2")
	w("tx200", "false")
	w("breath", "30 s")
	w("alarm", "Solo sulla radio")
	w("only_ch", "false")
	w("scan_hang", "10,0 s")
	w("lang", "ภาษาไทย")
	w("boot_msg1", "IU0VYT Flash")
	w("boot_msg2", "0123456789ABCDEF")
	for addr, want := range map[int][]byte{
		0xC90:  {0, 8, 4, 3, 3, 6},
		0xC9D:  {4},
		0xCA1:  {0x85}, // scansione SE; il resto resta
		0xCA2:  {0x69}, // il bit non usato (0x40) resta
		0xCA8:  {0x86}, // 8,33 kHz e 50 kHz
		0xCAB:  {0x82}, // roger tono 2, TX 200 spento, bit 0x02 non usato conservato
		0xCAF:  {0x40}, // lampeggio 30 s, allarme solo sulla radio, non solo canali
		0x1C00: append([]byte("IU0VYT Flash"), 0, 0, 0, 0),
		0x1C10: []byte("0123456789ABCDEF"),
		0x1F28: {8},
		0x1F2F: {19},
	} {
		if got := img[8+addr : 8+addr+len(want)]; !bytes.Equal(got, want) {
			t.Errorf("%04X: % X, atteso % X", addr, got, want)
		}
	}
	_, d := tdcGet(t, tab, img, "boot_msg3")
	for _, bad := range []string{"0123456789ABCDEFG", "perché"} {
		if err := writeSetting(d, append([]byte{}, img...), bad); err == nil {
			t.Errorf("messaggio %q: atteso un errore", bad)
		}
	}
}
