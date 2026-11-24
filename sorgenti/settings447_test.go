//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
)

// ogni voce della tabella: tutte le scelte (o gli estremi) si scrivono e si
// rileggono uguali, e toccano solo i byte della voce
func checkTable447(t *testing.T, tb *setTable, img []byte) {
	t.Helper()
	for i := range tb.Settings {
		d := &tb.Settings[i]
		offs, masks, err := d.Enc.fields()
		if d.Enc.Kind == "str" {
			offs, err = d.Enc.Offs, nil
			masks = make([][]byte, len(offs))
			for k := range masks {
				masks[k] = []byte{0xFF}
			}
		}
		if err != nil || len(offs) == 0 {
			t.Fatalf("%s: campi %v (%v)", d.Key, offs, err)
		}
		var vals []string
		switch d.Type {
		case "bool":
			vals = []string{"true", "false", "true"}
		case "list":
			vals = append(append([]string{}, d.Options...), d.Options[0])
		case "int":
			vals = []string{itoa447(int(d.Min)), itoa447(int(d.Max)), itoa447(int(d.Min) + 1)}
		case "str":
			vals = []string{"IU0VYT", "", "Prova 123"[:min(9, d.MaxLen)]}
		default:
			t.Fatalf("%s: tipo %s", d.Key, d.Type)
		}
		for _, v := range vals {
			before := append([]byte{}, img...)
			if err := writeSetting(d, img, v); err != nil {
				t.Fatalf("%s=%q: %v", d.Key, v, err)
			}
			if got, err := readSetting(d, img); err != nil || got != v {
				t.Fatalf("%s: scritto %q, riletto %q (%v)", d.Key, v, got, err)
			}
			for k := range before {
				before[k] ^= img[k]
			}
			for a, o := range offs {
				for k, mk := range masks[a] {
					before[o+k] &^= mk
				}
			}
			if !bytes.Equal(before, make([]byte, len(before))) {
				t.Fatalf("%s=%q: cambiati byte fuori dalla voce", d.Key, v)
			}
		}
	}
}

func itoa447(v int) string {
	b := []byte{}
	if v == 0 {
		return "0"
	}
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

// 4.47, famiglia AnyTone: le voci stanno dove le mettono qdmr e dmrconfig
// (posizioni riscritte qui a mano), le DMR-6X2 hanno scambiati il tono dei tasti
// e il tono di canale libero, ciò che si imposta arriva alla radio simulata e, se
// non si cambia nulla, nelle tre parti delle impostazioni non si scrive niente
func TestSettings447Anytone(t *testing.T) {
	for _, kind := range []string{"at-d868uv", "at-d878uv", "at-d878uv2", "bt-dmr6x2", "bt-dmr6x2pro"} {
		m := genByID[kind]
		x6 := ato(m).x6
		radio := atTestImg(t, m)
		for k := 0; k < 0x60; k++ { // impostazioni «di fabbrica» diverse da zero
			radio[atOffSettings+k] = byte(k % 2)
		}
		copy(radio[atOffRadioID:], []byte{0x02, 0x22, 0x12, 0x34, 0})
		copy(radio[atOffRadioID+5:], "VECCHIO\x00")
		if !settingsAvailable(kind) || settingsFor(kind, radio[:atSize-1]) != nil {
			t.Fatalf("%s: tabella assente o data a un'immagine di un'altra lunghezza", kind)
		}
		tb := settingsFor(kind, radio)
		if tb == nil || len(tb.Settings) != 27 {
			t.Fatalf("%s: tabella assente o con un altro numero di voci", kind)
		}
		keyTone, idleTone := 0x00, 0x36
		if x6 {
			keyTone, idleTone = 0x36, 0x00
		}
		want := map[string]rs443{
			"radio_id": {0x073D00, 4, 0}, "boot_disp": {0x071606, 1, 3}, "work_mode": {0x071634, 1, 2}, "disp_mode": {0x071601, 1, 2},
			"brightness": {0x071626, 1, 5}, "show_clock": {0x071651, 1, 0}, "auto_keylock": {0x071602, 1, 0}, "auto_off": {0x071603, 1, 5},
			"power_save": {0x07160B, 1, 3}, "squelch_a": {0x071609, 1, 6}, "squelch_b": {0x07160A, 1, 6}, "vox": {0x07160C, 1, 4},
			"mic_gain": {0x07160F, 1, 5}, "enh_audio": {0x071657, 1, 0}, "gps": {0x071628, 1, 0}, "key_tone": {0x071600 + keyTone, 1, 0},
			"idle_tone": {0x071600 + idleTone, 1, 0}, "permit_dmr": {0x071631, 1, 0}, "permit_fm": {0x071631, 1, 0}, "reset_tone": {0x071632, 1, 0},
			"startup_tone": {0x071639, 1, 0}, "call_end": {0x07163A, 1, 0}, "sms_alert": {0x071629, 1, 0}, "call_alert": {0x07162F, 1, 0},
		}
		inSpan := func(o, n int) bool {
			for _, sp := range atSetSpans {
				if o >= sp.off && o+n <= sp.off+sp.n {
					return true
				}
			}
			return false
		}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			if d.Enc.Kind == "str" {
				if !inSpan(d.Enc.Offs[0], len(d.Enc.Offs)) {
					t.Fatalf("%s %s: testo fuori dalle parti scritte nella radio", kind, d.Key)
				}
				continue
			}
			offs, masks, err := d.Enc.fields()
			if err != nil || len(offs) != 1 || !inSpan(offs[0], len(masks[0])) {
				t.Fatalf("%s %s: posizione %v fuori dalle parti scritte nella radio", kind, d.Key, offs)
			}
			w, ok := want[d.Key]
			if !ok || offs[0] != w.off || len(masks[0]) != w.width || len(d.Options) != w.n {
				t.Fatalf("%s %s: posizione %#x, %d byte, %d scelte (attesi %+v)", kind, d.Key, offs[0], len(masks[0]), len(d.Options), w)
			}
			delete(want, d.Key)
		}
		if len(want) != 0 {
			t.Fatalf("%s: voci attese e assenti: %v", kind, want)
		}
		if s := set443(t, tb, "intro2").Enc.Offs; s[0] != 0x071C10 || len(s) != 14 {
			t.Fatalf("seconda riga di accensione: %#x, %d", s[0], len(s))
		}
		if s := set443(t, tb, "radio_name").Enc.Offs; s[0] != 0x073D05 || len(s) != 16 {
			t.Fatalf("nome della radio: %#x, %d", s[0], len(s))
		}
		if v, err := readSetting(set443(t, tb, "radio_id"), radio); err != nil || v != "2221234" {
			t.Fatalf("ID letto %q (%v)", v, err)
		}
		checkTable447(t, tb, append([]byte{}, radio...))

		e := m.Fam.(atFam).emu(m, radio).(*atEmu)
		atWithEmu(t, e)
		// 1. senza cambiare le impostazioni: nelle tre parti non si scrive nulla
		img := append([]byte{}, radio...)
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
			t.Fatal(err)
		}
		isSet := func(a uint32) bool {
			for _, sp := range atSetSpans {
				if a >= sp.addr && a < sp.addr+uint32(sp.n) {
					return true
				}
			}
			return false
		}
		for _, a := range e.writes {
			if isSet(a) {
				t.Fatalf("%s: impostazioni scritte (%#x) senza averle cambiate", kind, a)
			}
		}
		// 2. con le impostazioni cambiate: arrivano alla radio, e solo le righe cambiate
		for k, v := range map[string]string{"radio_id": "2229999", "radio_name": "IU0VYT", "intro1": "CIAO", "key_tone": "true",
			"idle_tone": "false", "squelch_a": "3", "auto_off": "2", "permit_fm": "true", "work_mode": "0"} {
			if err := writeSetting(set443(t, tb, k), img, v); err != nil {
				t.Fatalf("%s=%s: %v", k, v, err)
			}
		}
		if !bytes.Equal(img[atOffRadioID:atOffRadioID+5], []byte{0x02, 0x22, 0x99, 0x99, 0}) || string(img[atOffRadioID+5:atOffRadioID+12]) != "IU0VYT\x00" ||
			string(img[atOffBoot:atOffBoot+5]) != "CIAO\x00" || img[atOffSettings+keyTone] != 1 || img[atOffSettings+idleTone] != 0 ||
			img[atOffSettings+9] != 3 || img[atOffSettings+3] != 2 || img[atOffSettings+0x31] != 2|byte(0x31%2) || img[atOffSettings+0x34] != 0 {
			t.Fatalf("%s: byte scritti nell'immagine diversi dagli attesi: ID % X, nome %q, riga %q, impostazioni % X", kind,
				img[atOffRadioID:atOffRadioID+5], img[atOffRadioID+5:atOffRadioID+12], img[atOffBoot:atOffBoot+5], img[atOffSettings:atOffSettings+0x38])
		}
		e.mu.Lock()
		e.writes = nil
		e.mu.Unlock()
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, a := range e.writes {
			if isSet(a) {
				n++
			}
			if a >= atAddrBoot+0x20 && a < atAddrBoot+0x40 {
				t.Fatalf("%s: scritta la riga delle password", kind)
			}
		}
		if n == 0 || n > 8 {
			t.Fatalf("%s: righe delle impostazioni scritte %d", kind, n)
		}
		got := e.image()
		for _, sp := range atSetSpans {
			if !bytes.Equal(got[sp.off:sp.off+sp.n], img[sp.off:sp.off+sp.n]) {
				t.Fatalf("%s: impostazioni non arrivate alla radio (%#x)", kind, sp.addr)
			}
		}
		// il resto delle impostazioni generali (da 0x60 in poi) resta quello della radio
		if !bytes.Equal(got[atOffSettings+atSetLen:atOffBoot], radio[atOffSettings+atSetLen:atOffBoot]) {
			t.Fatalf("%s: toccata la parte delle impostazioni non gestita", kind)
		}
	}
}

// 4.47, nicFW880 / nicFW950: campi di 16 bit alle posizioni della struttura
// settingsInfo; le impostazioni cambiate arrivano alla radio, la taratura mai
func TestSettings447Nicfw880(t *testing.T) {
	for _, kind := range []string{"rtl-rt880-nicfw", "rtl-rt950pro-nicfw"} {
		m := genByID[kind]
		radio := bytes.Repeat([]byte{0}, nicfw880Size)
		if !settingsAvailable(kind) || settingsFor(kind, radio) != nil {
			t.Fatal("tabella assente, o data a una memoria senza il numero di controllo")
		}
		binary.LittleEndian.PutUint16(radio[nicfw880SetAt:], 0xABC6)
		for i := nicfw880CalLo; i < nicfw880CalHi; i++ {
			radio[i] = byte(i)
		}
		tb := settingsFor(kind, radio)
		if tb == nil || len(tb.Settings) != 17 {
			t.Fatal("tabella assente o con un altro numero di voci")
		}
		// posizioni nella struttura (contate a mano dalla documentazione)
		want := map[string]int{"squelch": 4, "noise_trigger": 6, "noise_ceiling": 8, "noise_hyst": 10, "sq_throttle": 12, "mic_gain": 18,
			"lcd_timeout": 26, "heartbeat": 28, "lcd_inverted": 30, "lcd_gamma": 32, "vox": 36, "ste": 40, "tone_time": 232,
			"lcd_day": 258, "lcd_night": 260, "lcd_mode": 262, "noise_gate": 270}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			offs, masks, err := d.Enc.fields()
			w, ok := want[d.Key]
			if err != nil || !ok || len(offs) != 1 || offs[0] != nicfw880SetAt+w || len(masks[0]) != 2 || offs[0]+2 > nicfw880CalLo {
				t.Fatalf("%s: posizione %v", d.Key, offs)
			}
			delete(want, d.Key)
		}
		if len(want) != 0 {
			t.Fatalf("voci assenti: %v", want)
		}
		checkTable447(t, tb, append([]byte{}, radio...))

		e := m.Fam.(nicfw880Fam).emu(m, radio).(*nicfw880Emu)
		img := append([]byte{}, radio...)
		for k, v := range map[string]string{"squelch": "5", "lcd_timeout": "300", "noise_gate": "true", "lcd_night": "7"} {
			if err := writeSetting(set443(t, tb, k), img, v); err != nil {
				t.Fatalf("%s=%s: %v", k, v, err)
			}
		}
		s := img[nicfw880SetAt:]
		if s[4] != 5 || s[5] != 0 || s[26] != 0x2C || s[27] != 0x01 || s[270] != 1 || s[271] != 0 || s[260] != 7 {
			t.Fatal("byte scritti nell'immagine diversi dagli attesi")
		}
		for i := nicfw880CalLo; i < nicfw880CalHi; i++ {
			img[i] ^= 0x55 // una taratura diversa nell'immagine non deve arrivare alla radio
		}
		emuRun(e, func() {
			if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {}); err != nil {
				t.Fatal(err)
			}
		})
		got := e.image()
		if !bytes.Equal(got[:nicfw880CalLo], img[:nicfw880CalLo]) {
			t.Fatal("impostazioni non arrivate alla radio")
		}
		if !bytes.Equal(got[nicfw880CalLo:], radio[nicfw880CalLo:]) {
			t.Fatal("taratura o parti successive toccate")
		}
		if len(e.writes) != 2 { // due blocchi da 128 byte: 0-127 e 256-383 delle impostazioni
			t.Fatalf("blocchi scritti: %v", e.writes)
		}
	}
}

// 4.47: TD-UV68 con la tabella della TD-H8; TYT MD-UV380 e parenti con quella
// della DM-1701
func TestSettings447Twins(t *testing.T) {
	m := genByID["td-uv68"]
	h8 := genByID["td-tdh8"]
	if m == nil || h8 == nil || genMinSize(m) != genMinSize(h8) {
		t.Fatal("TD-UV68 e TD-H8 non hanno la stessa memoria")
	}
	img := make([]byte, genMinSize(m))
	a, b := settingsFor("td-uv68", img), settingsFor("td-tdh8", img)
	if !settingsAvailable("td-uv68") || a == nil || b == nil || len(a.Settings) == 0 || len(a.Settings) != len(b.Settings) {
		t.Fatalf("TD-UV68: tabella %v, TD-H8 %v", a != nil, b != nil)
	}
	for _, kind := range []string{"bf-dm1701", "tyt-mduv380", "tyt-mduv390", "tyt-md2017", "tyt-md9600"} {
		img := make([]byte, tytMemSize)
		tb := settingsFor(kind, img)
		if !settingsAvailable(kind) || tb == nil || len(tb.Settings) != 4 || settingsFor(kind, img[:tytMemSize-1]) != nil {
			t.Fatalf("%s: tabella assente", kind)
		}
		if err := writeSetting(set443(t, tb, "radio_id"), img, "2221234"); err != nil || img[0x2084] != 0xB2 || img[0x2085] != 0xE4 || img[0x2086] != 0x21 {
			t.Fatalf("%s: ID scritto % X (%v)", kind, img[0x2084:0x2087], err)
		}
	}
}
