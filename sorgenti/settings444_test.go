//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"testing"
)

// 4.44: impostazioni ricavate dai programmi dei costruttori (seguito della 4.43).
// Le posizioni sono riscritte qui a mano dall'analisi dei programmi.

// Radtel RT-480 e RT-68: blocco 0x3A00 della radio = 0x2000 dell'immagine; il
// driver ora lo scrive (prima scriveva solo canali e nomi)
func TestSettings444RT480(t *testing.T) {
	const o = 0x2000
	want := map[string]twin443{
		"sql": {o + 0x00, 0xFF, 10}, "save": {o + 0x01, 0xFF, 0}, "vox": {o + 0x02, 0xFF, 11}, "light": {o + 0x03, 0xFF, 5},
		"tdr": {o + 0x04, 0xFF, 0}, "tot": {o + 0x05, 0xFF, 6}, "beep": {o + 0x06, 0xFF, 0}, "voice": {o + 0x07, 0xFF, 0},
		"lang": {o + 0x08, 0xFF, 2}, "dtmf_st": {o + 0x09, 0xFF, 4}, "scan_mode": {o + 0x0A, 0xFF, 3}, "ptt_id": {o + 0x0B, 0xFF, 4},
		"ptt_delay": {o + 0x0C, 0xFF, 7}, "disp_a": {o + 0x0D, 0xFF, 3}, "disp_b": {o + 0x0E, 0xFF, 3}, "bcl": {o + 0x0F, 0xFF, 0},
		"auto_lock": {o + 0x10, 0xFF, 4}, "alarm_mode": {o + 0x11, 0xFF, 3}, "alarm_sound": {o + 0x12, 0xFF, 0}, "tdr_tx": {o + 0x13, 0xFF, 3},
		"ste": {o + 0x14, 0xFF, 0}, "rp_ste": {o + 0x15, 0xFF, 11}, "rpt_rl": {o + 0x16, 0xFF, 11}, "roger": {o + 0x17, 0xFF, 0},
		"fm_radio": {o + 0x19, 0xFF, 0}, "work_mode": {o + 0x1A, 0xFF, 2}, "key_lock": {o + 0x1B, 0xFF, 0}, "pwr_msg": {o + 0x1C, 0x03, 3},
		"tx_beep": {o + 0x1D, 0x01, 0}, "auto_off": {o + 0x1D, 0xF0, 5}, "tone": {o + 0x1E, 0xFF, 4}, "ani": {o + 0x25, 0x01, 0},
	}
	for _, kind := range []string{"rtl-rt480", "rtl-rt68"} {
		provSize443[kind] = rt480Size()
		tb, img := checkProv443(t, kind, want)
		for i := range img {
			img[i] = 0
		}
		// bit che non appartengono a nessuna voce: devono restare come sono
		img[o+0x1C], img[o+0x1F], img[o+0x25] = 0x40, 0xFF, 0xA0
		for k, v := range map[string]string{"sql": "5", "vox": "10", "tot": "5", "fm_radio": "true", "pwr_msg": "2", "tx_beep": "true",
			"auto_off": "4", "tone": "2", "ani": "true", "tdr": "true", "light": "4"} {
			if err := writeSetting(set443(t, tb, k), img, v); err != nil {
				t.Fatalf("%s %s=%s: %v", kind, k, v, err)
			}
		}
		for a, v := range map[int]byte{0x00: 5, 0x02: 10, 0x05: 5, 0x19: 0, 0x1C: 0x42, 0x1D: 0x41, 0x1E: 2, 0x1F: 0xFF, 0x25: 0xA1, 0x04: 1, 0x03: 4} {
			if img[o+a] != v {
				t.Fatalf("%s: byte +%#x = %#x invece di %#x", kind, a, img[o+a], v)
			}
		}
		// radio FM: nel programma 0 = abilitata
		if err := writeSetting(set443(t, tb, "fm_radio"), img, "false"); err != nil || img[o+0x19] != 1 {
			t.Fatalf("%s: radio FM spenta = %#x (%v)", kind, img[o+0x19], err)
		}
		cable443(t, kind, img, map[string]string{"sql": "7", "tot": "3", "fm_radio": "true", "auto_off": "2", "rp_ste": "10", "alarm_mode": "1"})
	}
	apiCase442(t, "rtl-rt480", "tone", "2", 4, make([]byte, rt480Size()))
}

// Radtel RT-5D: i 64 byte delle «funzioni opzionali», dopo DTMF, chiavi,
// contatti, gruppi, canali e VFO
func TestSettings444RT5D(t *testing.T) {
	const o = 272 + 264 + 800*80 + 1024*4 + 1024*64 + 128
	if rt5dOff("funzioni") != o {
		t.Fatalf("sezione delle funzioni a %d, attesa %d", rt5dOff("funzioni"), o)
	}
	want := map[string]twin443{
		"sql": {o + 0x00, 0xFF, 10}, "save": {o + 0x01, 0xFF, 5}, "vox": {o + 0x02, 0xFF, 10}, "light": {o + 0x03, 0xFF, 9},
		"tdr": {o + 0x04, 0xFF, 0}, "tot": {o + 0x05, 0xFF, 6}, "beep": {o + 0x06, 0xFF, 0}, "voice": {o + 0x07, 0xFF, 0},
		"lang": {o + 0x08, 0xFF, 2}, "dtmf_st": {o + 0x09, 0xFF, 4}, "scan_mode": {o + 0x0A, 0xFF, 3}, "ptt_id": {o + 0x0B, 0xFF, 4},
		"ptt_delay": {o + 0x0C, 0xFF, 7}, "disp_a": {o + 0x0D, 0xFF, 3}, "disp_b": {o + 0x0E, 0xFF, 3}, "auto_lock": {o + 0x10, 0xFF, 4},
		"alarm_mode": {o + 0x11, 0xFF, 3}, "alarm_sound": {o + 0x12, 0xFF, 0}, "tdr_tx": {o + 0x13, 0xFF, 3}, "ste": {o + 0x14, 0xFF, 0},
		"rp_ste": {o + 0x15, 0xFF, 11}, "rpt_rl": {o + 0x16, 0xFF, 11}, "roger": {o + 0x17, 0xFF, 0}, "band": {o + 0x18, 0xFF, 2},
		"fm_radio": {o + 0x19, 0xFF, 0}, "mode_a": {o + 0x1A, 0x0F, 2}, "mode_b": {o + 0x1A, 0xF0, 2}, "key_lock": {o + 0x1B, 0xFF, 0},
		"boot": {o + 0x1C, 0xFF, 2}, "tone": {o + 0x1E, 0xFF, 4}, "tx_start": {o + 0x1F, 0xFF, 0}, "vox_delay": {o + 0x20, 0xFF, 16},
		"menu_quit": {o + 0x21, 0xFF, 11}, "dsql": {o + 0x22, 0xFF, 10}, "ste_freq": {o + 0x26, 0xFF, 2}, "wx_ch": {o + 0x27, 0xFF, 10},
		"top_key": {o + 0x2A, 0xFF, 7}, "key2_s": {o + 0x2B, 0xFF, 8}, "key2_l": {o + 0x2C, 0xFF, 7}, "key3_s": {o + 0x2D, 0xFF, 7},
		"key3_l": {o + 0x2E, 0xFF, 7}, "keep_call": {o + 0x31, 0x1F, 20}, "tdr_back": {o + 0x33, 0xFF, 11},
	}
	provSize443["rtl-rt5d"] = rt5dSize()
	tb, img := checkProv443(t, "rtl-rt5d", want)
	for i := range img {
		img[i] = 0
	}
	for k, v := range map[string]string{"sql": "9", "mode_a": "1", "mode_b": "1", "fm_radio": "false", "key2_s": "7", "keep_call": "19",
		"vox_delay": "15", "menu_quit": "10", "tdr_back": "10", "light": "8"} {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s=%s: %v", k, v, err)
		}
	}
	for a, v := range map[int]byte{0x00: 9, 0x1A: 0x11, 0x19: 1, 0x2B: 7, 0x31: 19, 0x20: 15, 0x21: 10, 0x33: 10, 0x03: 8} {
		if img[o+a] != v {
			t.Fatalf("byte +%#x = %#x invece di %#x", a, img[o+a], v)
		}
	}
	for key, w := range map[string][2]string{"vox_delay": {"0,5 s", "2,0 s"}, "menu_quit": {"5 s", "60 s"}, "tdr_back": {"0,5 s", "10 s"},
		"keep_call": {"1 s", "20 s"}, "rp_ste": {"0 ms", "1000 ms"}} {
		d := set443(t, tb, key)
		if d.ItOpts[0] != w[0] || d.ItOpts[len(d.ItOpts)-1] != w[1] {
			t.Fatalf("%s: scelte da %q a %q", key, d.ItOpts[0], d.ItOpts[len(d.ItOpts)-1])
		}
	}
	cable443(t, "rtl-rt5d", img, map[string]string{"sql": "4", "tot": "5", "fm_radio": "true", "key3_l": "6", "dsql": "3", "mode_b": "0"})
	apiCase442(t, "rtl-rt5d", "tone", "2", 5, make([]byte, rt5dSize()))
}

// Radtel RT-69 GPS: gemella della Wouxun KG-UV8D Plus di CHIRP, con le sole voci
// verificate sul programma del costruttore (indirizzi delle due griglie di 889G)
func TestSettings444RT69(t *testing.T) {
	want := map[string]twin443{
		"voice": {0x815, 0x01, 0}, "channel_menu": {0x818, 0x01, 0}, "power_save": {0x819, 0x01, 0}, "autolock": {0x81A, 0x01, 0},
		"keylock": {0x81B, 0x01, 0}, "beep": {0x81C, 0x01, 0}, "stopwatch": {0x81D, 0x01, 0}, "scan_rev": {0x81F, 0x03, 3},
		"roger_beep": {0x821, 0x03, 4}, "ani_sw": {0x830, 0x01, 0}, "dtmf_st": {0x838, 0x03, 4}, "bcl_a": {0x839, 0x01, 0},
		"bcl_b": {0x83A, 0x01, 0}, "ptt_id": {0x83B, 0x03, 3}, "prich_sw": {0x83C, 0x01, 0}, "rpt_spk": {0x83E, 0x01, 0},
		"rpt_ptt": {0x83F, 0x01, 0}, "alert": {0x840, 0x03, 4}, "language": {0x856, 0x01, 2}, "rpt_tone": {0x864, 0x01, 0},
		"scan_det": {0x866, 0x01, 0},
	}
	tb, img := checkTwin443(t, "rtl-rt69", want)
	for a := 0x800; a < 0x880; a++ {
		img[a] = 0
	}
	base := append([]byte{}, img...)
	// valori come li mostra il programma: ascolto DTMF 0 spento, 1 tastiera, 2
	// identificativo, 3 entrambi; lingua 0 cinese, 1 inglese; tono 0 = 1750, 1 = 2100
	for k, v := range map[string]string{"dtmf_st": "ANI-ST", "language": "English", "alert": "2100", "roger_beep": "Both", "ptt_id": "Both",
		"scan_rev": "SE", "voice": "true", "bcl_b": "true"} {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s=%s: %v", k, v, err)
		}
	}
	for a, v := range map[int]byte{0x838: 2, 0x856: 1, 0x840: 1, 0x821: 3, 0x83B: 2, 0x81F: 2, 0x815: 1, 0x83A: 1} {
		if img[a] != v {
			t.Fatalf("byte %#x = %#x invece di %#x", a, img[a], v)
		}
	}
	if d := set443(t, tb, "dtmf_st"); d.Options[0] != "Off" || d.Options[3] != "DT-ANI" {
		t.Fatalf("ascolto DTMF: %v", d.Options)
	}
	// la tabella di CHIRP della KG-UV8D Plus non è toccata
	for _, d := range chirpSettings()["KGUV8DPlusRadio"].Settings {
		if d.Key == "dtmf_st" && d.Options[0] != "DT-ST" {
			t.Fatalf("tabella di CHIRP modificata: %v", d.Options)
		}
	}
	provRadio443["rtl-rt69"] = base
	cable443(t, "rtl-rt69", img, map[string]string{"dtmf_st": "DT-ST", "language": "Chinese", "beep": "true", "roger_beep": "End"})
	apiCase442(t, "rtl-rt69", "alert", "1000", 0, base)
}

// Radtel RT-4D, RT-6D, RT-1000 Pro: il blocco delle impostazioni (blocco 8) ora è
// in coda all'immagine; posizioni lette nei tre programmi del costruttore
func TestSettings444RT4D(t *testing.T) {
	const o = (48 + 128) * 1024
	if rt4dCfgOff != o || rt4dSizeCfg != o+1024 {
		t.Fatalf("blocco delle impostazioni a %d", rt4dCfgOff)
	}
	want := map[string]twin443{
		"boot_logo": {o + 0x10, 0xFF, 0}, "boot_tone": {o + 0x13, 0xFF, 0}, "boot_text": {o + 0x14, 0xFF, 0}, "voice": {o + 0x5C, 0xFF, 0},
		"beep": {o + 0x5D, 0xFF, 0}, "lock_timer": {o + 0x5F, 0xFF, 43}, "light": {o + 0x61, 0xFF, 5}, "lcd_timer": {o + 0x62, 0xFF, 43},
		"save_mode": {o + 0x63, 0xFF, 4}, "menu_timer": {o + 0x65, 0xFF, 43}, "talkaround": {o + 0x67, 0xFF, 3}, "alarm_type": {o + 0x68, 0xFF, 3},
		"tx_pri": {o + 0x7E, 0xFF, 2}, "scan_mode": {o + 0xA3, 0xFF, 3}, "scan_return": {o + 0xA4, 0xFF, 2}, "scan_dwell": {o + 0xA5, 0xFF, 31},
		"freq_digits": {o + 0xEA, 0xFF, 2}, "sql": {o + 0x102, 0xFF, 0}, "mic_gain": {o + 0x105, 0xFF, 0}, "rx_volume": {o + 0x106, 0xFF, 0},
		"tx_start_beep": {o + 0x10B, 0xFF, 0}, "roger": {o + 0x10C, 0xFF, 5}, "detect_range": {o + 0x110, 0xFF, 11}, "relay_delay": {o + 0x111, 0xFF, 21},
		"glitch": {o + 0x114, 0xFF, 0}, "dtmf_delay": {o + 0x200, 0xFF, 21}, "dtmf_duration": {o + 0x201, 0xFF, 18}, "dtmf_interval": {o + 0x202, 0xFF, 18},
		"dtmf_mode": {o + 0x203, 0xFF, 4}, "dtmf_select": {o + 0x204, 0xFF, 16}, "dtmf_display": {o + 0x205, 0xFF, 0}, "dtmf_gain": {o + 0x206, 0xFF, 0},
		"dtmf_th": {o + 0x207, 0xFF, 0}, "dtmf_remote": {o + 0x208, 0xFF, 0},
	}
	if len(rt4dCfgFields()) != len(want) {
		t.Fatalf("%d campi scritti dal driver, %d voci", len(rt4dCfgFields()), len(want))
	}
	for _, kind := range []string{"rtl-rt4d", "rtl-rt6d", "rtl-rt1000pro"} {
		m := genByID[kind]
		f := m.Fam.(rt4dFam)
		provSize443[kind] = rt4dSizeCfg
		tb, img := checkProv443(t, kind, want)
		if settingsFor(kind, img[:rt4dSize]) != nil {
			t.Fatalf("%s: tabella data a un'immagine senza il blocco delle impostazioni", kind)
		}
		// la tabella della RT-871A, da cui vengono 28 voci, non è toccata
		if d := set443(t, rt871aSettings, "sql"); d.Enc.Fields[0][0].(float64) != 0x102 {
			t.Fatalf("tabella della RT-871A modificata")
		}
		for i := range img {
			img[i] = 0
		}
		// nell'immagine il resto del blocco è diverso da quello della radio: non deve arrivarci
		for a := 0; a < 1024; a++ {
			img[o+a] = 0x33
		}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			v := "1"
			if d.Type == "bool" {
				v = "true"
			}
			if err := writeSetting(d, img, v); err != nil {
				t.Fatalf("%s %s: %v", kind, d.Key, err)
			}
		}
		for k, v := range map[string]string{"sql": "10", "roger": "3", "relay_delay": "20", "lock_timer": "42", "light": "4", "dtmf_gain": "127"} {
			if err := writeSetting(set443(t, tb, k), img, v); err != nil {
				t.Fatalf("%s %s=%s: %v", kind, k, v, err)
			}
		}
		for a, v := range map[int]byte{0x102: 10, 0x10C: 3, 0x111: 20, 0x5F: 42, 0x61: 4, 0x206: 127, 0x5C: 1, 0x112: 0x33} {
			if img[o+a] != v {
				t.Fatalf("%s: byte +%#x = %#x invece di %#x", kind, a, img[o+a], v)
			}
		}
		// radio simulata con un altro contenuto nel blocco 8
		radioCfg := make([]byte, 1024)
		for a := range radioCfg {
			radioCfg[a] = byte(a*7 + 1)
		}
		copy(radioCfg, bytes.Repeat([]byte{0xFF}, 12)) // nessuna password
		start := make([]byte, rt4dSizeCfg)
		copy(start[o:], radioCfg)
		e := f.emu(m, start).(*rt4dEmu)
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			a, b := newPipe()
			a.baud = baud
			go e.serve(b)
			return a, nil
		}
		if _, err := f.Upload(context.Background(), m, "COM7", img, noProg); err != nil {
			openPortFunc = old
			t.Fatalf("%s: %v", kind, err)
		}
		known := map[int]bool{}
		for _, fl := range rt4dCfgFields() {
			known[fl[0]] = true
		}
		for a := 0; a < 1024; a++ {
			exp := radioCfg[a]
			switch {
			case known[a]:
				exp = img[o+a]
			case a == 14 || a == 15:
				exp = 0xFF
			}
			if e.blocks[rt4dBlkCfg][a] != exp {
				openPortFunc = old
				t.Fatalf("%s: blocco delle impostazioni, byte %#x = %#x invece di %#x", kind, a, e.blocks[rt4dBlkCfg][a], exp)
			}
		}
		// immagine del vecchio formato (senza il blocco): la radio tiene le sue impostazioni
		e2 := f.emu(m, start).(*rt4dEmu)
		openPortFunc = func(name string, baud int) (Port, error) {
			a, b := newPipe()
			a.baud = baud
			go e2.serve(b)
			return a, nil
		}
		if _, err := f.Upload(context.Background(), m, "COM7", img[:rt4dSize], noProg); err != nil {
			openPortFunc = old
			t.Fatalf("%s: immagine del vecchio formato: %v", kind, err)
		}
		openPortFunc = old
		ref := append([]byte{}, radioCfg...)
		ref[14], ref[15] = 0xFF, 0xFF
		if !bytes.Equal(e2.blocks[rt4dBlkCfg], ref) {
			t.Fatalf("%s: con un'immagine del vecchio formato le impostazioni della radio sono cambiate", kind)
		}
		// lettura dopo la scrittura: le voci tornano
		clean := make([]byte, rt4dSizeCfg)
		cable443(t, kind, clean, map[string]string{"sql": "6", "roger": "4", "voice": "true", "dtmf_select": "15", "scan_dwell": "30"})
	}
	apiCase442(t, "rtl-rt4d", "light", "3", 5, make([]byte, rt4dSizeCfg))
}

// Bajeton DM-5000: le voci che il suo programma ha alla stessa posizione e con
// le stesse scelte della RT-871A; il blocco è all'inizio dell'immagine
func TestSettings444DM5000(t *testing.T) {
	const kind = "bj-dm5000"
	m := genByID[kind]
	img := make([]byte, dm5kSize())
	if !settingsAvailable(kind) || settingsFor(kind, img[:4*1024-1]) != nil {
		t.Fatal("tabella assente o data a un'immagine troppo corta")
	}
	tb := settingsFor(kind, img)
	if tb == nil || len(tb.Settings) != 53 {
		t.Fatalf("tabella: %+v", tb)
	}
	// posizioni lette nel programma «BAJETON DM-5000 CPS» (posizione, byte, scelte)
	want := map[string]rs443{
		"saver_timer": {0x11, 1, 43}, "voice": {0x5C, 1, 0}, "key_lock": {0x5E, 1, 0}, "lcd_timer": {0x62, 1, 43}, "save_start": {0x64, 1, 43},
		"dw": {0x66, 1, 0}, "tx_pri": {0x7E, 1, 2}, "main_ptt": {0x7F, 1, 2}, "main_range": {0x83, 1, 2}, "mode_a": {0x84, 1, 3},
		"mode_b": {0x89, 1, 3}, "lock1": {0x8E, 1, 3}, "lock1_start": {0x8F, 2, 0}, "lock1_end": {0x91, 2, 0}, "lock3": {0x98, 1, 3},
		"lock4_end": {0xA0, 2, 0}, "scan_dir": {0xA2, 1, 2}, "scan_dwell": {0xA5, 1, 31}, "freq_digits": {0xEA, 1, 2}, "tone": {0x100, 2, 0},
		"sql": {0x102, 1, 0}, "mic_gain": {0x105, 1, 0}, "rx_volume": {0x106, 1, 0}, "tx_start_beep": {0x10B, 1, 0}, "roger": {0x10C, 1, 5},
		"detect_range": {0x110, 1, 11}, "relay_delay": {0x111, 2, 21}, "glitch": {0x114, 1, 0}, "dtmf_delay": {0x200, 1, 21},
		"dtmf_select": {0x204, 1, 16}, "dtmf_gain": {0x206, 1, 0}, "dtmf_remote": {0x208, 1, 0}, "time_cal": {0x209, 1, 0},
	}
	used := map[int]string{}
	for i := range tb.Settings {
		d := &tb.Settings[i]
		offs, masks, err := d.Enc.fields()
		if err != nil || len(offs) != 1 || offs[0] < 0x10 || offs[0]+len(masks[0]) > 960 {
			t.Fatalf("%s: posizione %v fuori dalla parte del blocco che il driver scrive dall'immagine", d.Key, offs)
		}
		for k := range masks[0] {
			if o := used[offs[0]+k]; o != "" {
				t.Fatalf("%s e %s usano lo stesso byte", d.Key, o)
			}
			used[offs[0]+k] = d.Key
		}
		if w, ok := want[d.Key]; ok && (offs[0] != w.off || len(masks[0]) != w.width || len(d.Options) != w.n) {
			t.Fatalf("%s: posizione %#x, %d byte, %d scelte (attesi %#x, %d, %d)", d.Key, offs[0], len(masks[0]), len(d.Options), w.off, w.width, w.n)
		}
		delete(want, d.Key)
		switch d.Key { // voci che nel programma della DM-5000 sono diverse o mancano
		case "step", "show_a", "show_b", "vox", "vox_th", "vox_delay", "fs1_s", "key0", "voice_refresh", "rssi_refresh":
			t.Fatalf("voce %s non verificata sul programma della DM-5000", d.Key)
		}
	}
	if len(want) != 0 {
		t.Fatalf("voci attese e assenti: %v", want)
	}
	if d := set443(t, tb, "roger"); d.ItOpts[3] != "MDC1200" {
		t.Fatalf("tono di fine: %v", d.ItOpts)
	}
	for k, v := range map[string]string{"sql": "8", "tone": "1750", "relay_delay": "20", "lock1": "2", "lock1_start": "144", "lock1_end": "146", "roger": "3"} {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s=%s: %v", k, v, err)
		}
	}
	for a, v := range map[int]byte{0x102: 8, 0x100: 0xD6, 0x101: 0x06, 0x111: 20, 0x112: 0, 0x8E: 2, 0x8F: 144, 0x90: 0, 0x91: 146, 0x10C: 3} {
		if img[a] != v {
			t.Fatalf("byte %#x = %#x invece di %#x", a, img[a], v)
		}
	}
	cable443(t, kind, img, map[string]string{"sql": "3", "voice": "true", "dtmf_select": "9", "scan_dwell": "12", "tone": "2100"})
	apiCase442(t, kind, "lock_timer", "5", 6, make([]byte, dm5kSize()))
	_ = m
}
