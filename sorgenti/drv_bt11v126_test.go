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
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// 4.25: Radioddity GS-10B Pro e Bajeton BJ-8300 con il firmware nuovo
// (programmi GS-10B Pro_EnCPS v1.0.3 e BT-11_CPS v1.2.6).

func bt11v126Link(t *testing.T, m *genModel, radio []byte) (*rt900Emu, func()) {
	t.Helper()
	e := emuFor(m, radio).(*rt900Emu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	go func() { _ = e.serve(b) }()
	return e, func() { openPortFunc = old; _ = a.Close() }
}

// posizioni del record del canale, trascritte da Set/GetCHInfo_Parameter dei programmi
func TestBT11v126Channel(t *testing.T) {
	for _, tc := range []struct {
		id         string
		slots, hi  int
		lastOff    int
		zones      int
		readBlocks int
	}{
		{"rd-gs10bpro", 999, 0, 998 * 32, 0, (0x7D00 + 0x40 + 0x40 + 0x140 + 0xC0 + 0x300) / 0x40},
		{"bj-8300-990", 990, 0x7800, 989 * 32, 10, (0x7BC0+0x40+0x40+0x140+0xC0+0x300)/0x40 + 4},
	} {
		m := genByID[tc.id]
		if m == nil {
			t.Fatalf("%s: modello assente", tc.id)
		}
		o := rt900o(m)
		if m.Slots != tc.slots || m.Zones != tc.zones || m.Baud != 57600 || string(o.magic) != "PROGRAMBT11U" || o.blk() != 0x40 || o.memsize != 0xC300 || !m.NoChirp {
			t.Fatalf("%s: %d canali, %d zone, %d baud, %q", tc.id, m.Slots, m.Zones, m.Baud, o.magic)
		}
		if o.fp1 != nil || o.fp2 != nil || o.model != "" || o.plain {
			t.Fatalf("%s: il programma del costruttore non controlla l'identificativo", tc.id)
		}
		img := bytes.Repeat([]byte{0xFF}, o.memsize)
		c := &Chan{Rx: 145612500, Tx: 145012500, Name: "R1 MONTE", Power: m.Powers[2].Sem, Narrow: true, Scan: true, TxTone: "88.5", RxTone: "D023N"}
		if err := m.Fam.Set(m, img, tc.slots-1, c); err != nil {
			t.Fatal(err)
		}
		r := img[tc.lastOff : tc.lastOff+32]
		// frequenze BCD in decine di Hz, byte meno significativo per primo
		if !bytes.Equal(r[0:4], []byte{0x50, 0x12, 0x56, 0x14}) || !bytes.Equal(r[4:8], []byte{0x50, 0x12, 0x50, 0x14}) {
			t.Fatalf("%s: frequenze % X", tc.id, r[:8])
		}
		// toni: DCS = posizione nella tabella + 1; CTCSS = decimi di Hz
		if r[8] != 1 || r[9] != 0 || r[10] != 0x75 || r[11] != 0x03 {
			t.Fatalf("%s: toni % X", tc.id, r[8:12])
		}
		// byte 14: potenza nei 4 bit bassi (2 = bassa); byte 15: bit 6 banda
		// stretta, bit 2 scansione
		if r[14]&0x0F != 2 || r[15]&0x40 == 0 || r[15]&0x04 == 0 || r[15]&0x02 != 0 {
			t.Fatalf("%s: byte 14-15 % X", tc.id, r[14:16])
		}
		if !bytes.Equal(r[20:32], append([]byte("R1 MONTE"), 0xFF, 0xFF, 0xFF, 0xFF)) {
			t.Fatalf("%s: nome % X", tc.id, r[20:32])
		}
		g := genGet(m, img, tc.slots-1)
		if g == nil || g.Rx != c.Rx || g.Tx != c.Tx || g.Name != c.Name || !g.Narrow || !g.Scan || g.TxTone != "88.5" || g.RxTone != "D023N" || g.Power != c.Power {
			t.Fatalf("%s: riletto %+v", tc.id, g)
		}
		// banda aerea: AM (bit 1 del byte 15)
		if err := m.Fam.Set(m, img, 0, &Chan{Rx: 118100000, Tx: 118100000, TxOff: true, Name: "TWR", Power: m.Powers[0].Sem}); err != nil {
			t.Fatal(err)
		}
		if img[15]&0x02 == 0 || !bytes.Equal(img[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
			t.Fatalf("%s: canale in banda aerea % X", tc.id, img[:16])
		}

		// lettura e scrittura con la radio simulata: stesse aree del programma
		radio := bytes.Repeat([]byte{0xFF}, 0xC900)
		copy(radio, img)
		for a := 0x8000; a < 0x8040; a++ {
			radio[a] = byte(a)
		}
		radio[0x9000], radio[0x902E] = 0x05, 0x03
		if tc.zones > 0 {
			for a := 0x7BC0; a < 0x7C00; a++ {
				radio[a] = 0x5A
			}
		}
		e, done := bt11v126Link(t, m, radio)
		blocks := 0
		got, _, err := m.Fam.Download(context.Background(), m, "EMU", func(d, tot int, s string) {
			if s == "Lettura…" {
				blocks++
			}
		})
		done()
		if err != nil {
			t.Fatalf("%s: lettura: %v", tc.id, err)
		}
		want := tc.readBlocks
		if tc.zones > 0 {
			want -= 4 // i nomi delle zone si leggono dopo, a parte
		}
		if blocks != want {
			t.Fatalf("%s: %d blocchi letti invece di %d", tc.id, blocks, want)
		}
		if !bytes.Equal(got[:0x7BC0], radio[:0x7BC0]) || !bytes.Equal(got[0x8000:0x8040], radio[0x8000:0x8040]) || got[0x902E] != 3 {
			t.Fatalf("%s: memoria letta diversa: %s", tc.id, diffRanges(radio[:len(got)], got))
		}
		if tc.zones > 0 && len(got) != 0xC900 {
			t.Fatalf("%s: nomi delle zone non letti (%#x byte)", tc.id, len(got))
		}
		// modifica e scrittura
		if err := m.Fam.Set(m, got, 1, &Chan{Rx: 433500000, Tx: 433500000, Name: "DUE", Power: m.Powers[0].Sem}); err != nil {
			t.Fatal(err)
		}
		if err := m.Fam.Set(m, got, tc.slots-1, nil); err != nil {
			t.Fatal(err)
		}
		e, done = bt11v126Link(t, m, radio)
		_, err = m.Fam.Upload(context.Background(), m, "EMU", got, func(int, int, string) {})
		done()
		if err != nil {
			t.Fatalf("%s: scrittura: %v", tc.id, err)
		}
		after := e.image()
		if c := genGet(m, after, 1); c == nil || c.Name != "DUE" {
			t.Fatalf("%s: canale 2 non scritto", tc.id)
		}
		if genGet(m, after, tc.slots-1) != nil {
			t.Fatalf("%s: ultimo canale non cancellato", tc.id)
		}
		// fuori dalle aree del programma la radio non cambia
		outside := [][2]int{{0x7D00, 0x8000}, {0x8040, 0x9000}, {0x9040, 0xA000}, {0xA140, 0xB000}, {0xB0C0, 0xC000}, {0xC300, 0xC900}}
		if tc.zones > 0 {
			// BJ-8300: il blocco dopo il canale 990 non si legge e non si scrive
			outside = append(outside, [2]int{0x7BC0, 0x7D00})
			if !bytes.Equal(got[0x7BC0:0x7C00], bytes.Repeat([]byte{0xFF}, 0x40)) {
				t.Fatalf("%s: blocco 0x7BC0 letto", tc.id)
			}
		}
		for _, r := range outside {
			if r[1] > len(after) {
				r[1] = len(after) // la radio simulata della GS-10B Pro finisce a 0xC380
			}
			if !bytes.Equal(after[r[0]:r[1]], radio[r[0]:r[1]]) {
				t.Fatalf("%s: area %#x-%#x modificata", tc.id, r[0], r[1])
			}
		}
	}
}

// BJ-8300 con il firmware nuovo: canali 961-990 e nomi delle zone solo se la
// radio li ha; con il firmware precedente non si scrive nulla
func TestBJ8300NewFirmwareGuard(t *testing.T) {
	m := genByID["bj-8300-990"]
	o := rt900o(m)
	if o.hi != [2]int{0x7800, 0x7BC0} || !o.hiStrict || o.ranges[0] != [2]int{0, 0x7800} || o.reads[0] != [2]int{0, 0x7BC0} {
		t.Fatalf("aree: %v %v", o.hi, o.ranges)
	}
	zs := m.ZoneNames
	if zs == nil || zs.Addr != 0xC800 || zs.Count != 10 || zs.Stride != 16 || zs.Max != 12 || zs.Area != 0x100 || !zs.GB || zs.Pad != 0xFF {
		t.Fatalf("nomi delle zone: %+v", zs)
	}
	blank := func() []byte { return bytes.Repeat([]byte{0xFF}, 0xC900) }
	upload := func(radio, img []byte) (*rt900Emu, error) {
		e, done := bt11v126Link(t, m, radio)
		_, err := m.Fam.Upload(context.Background(), m, "EMU", img, func(int, int, string) {})
		done()
		return e, err
	}

	// firmware nuovo: canale 990 e nome della zona 3 si scrivono
	radio := blank()
	img := append([]byte{}, radio...)
	rt950Chan(t, m, img, 989, 439000000, "ULTIMO")
	names := make([]string, 10)
	names[2] = "PONTI VHF"
	if err := setZoneNames(m, img, names); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(img[0xC820:0xC830], append([]byte("PONTI VHF"), bytes.Repeat([]byte{0xFF}, 7)...)) {
		t.Fatalf("nome della zona 3: % X", img[0xC820:0xC830])
	}
	e, err := upload(radio, img)
	if err != nil {
		t.Fatal(err)
	}
	got := e.image()
	if c := genGet(m, got, 989); c == nil || c.Name != "ULTIMO" {
		t.Fatalf("canale 990 non scritto: %+v", c)
	}
	if n, _, ok := zoneNamesOf(m, got); !ok || n[2] != "PONTI VHF" {
		t.Fatalf("nome della zona non scritto: %v", n)
	}

	// firmware precedente (dati che non sono canali dopo il 960): nessuna
	// scrittura, nemmeno se l'area non è stata modificata; la lettura lo segnala
	radio = blank()
	for a := 0x7803; a < 0x7900; a += 7 {
		radio[a] = 0x3C
	}
	if _, why := genPickLayout(m, radio); !strings.Contains(why, "Bajeton 8300") {
		t.Fatalf("lettura di una radio con il firmware precedente: %q", why)
	}
	img = append([]byte{}, radio...)
	rt950Chan(t, m, img, 0, 145500000, "UNO")
	e, err = upload(radio, img)
	if err == nil || !strings.Contains(err.Error(), "firmware precedente") {
		t.Fatalf("scrittura in una radio con il firmware precedente: %v", err)
	}
	if got := e.image(); !bytes.Equal(got, radio) {
		t.Fatalf("la radio è stata modificata: %s", diffRanges(radio, got))
	}
	if _, why := genPickLayout(m, blank()); why != "" {
		t.Fatalf("avviso con una radio vuota: %q", why)
	}

	// area dei nomi delle zone con altri dati: nomi modificati non scritti
	radio = blank()
	for a := 0xC800; a < 0xC900; a += 3 {
		radio[a] = 0x07
	}
	img = append([]byte{}, radio...)
	copy(img[0xC800:], bytes.Repeat([]byte{0xFF}, 0x100))
	if err := setZoneNames(m, img, names); err != nil {
		t.Fatal(err)
	}
	e, err = upload(radio, img)
	if err == nil || !strings.Contains(err.Error(), "nomi delle zone") {
		t.Fatalf("nomi delle zone in una radio che non li ha: %v", err)
	}
	if got := e.image(); !bytes.Equal(got, radio) {
		t.Fatalf("la radio è stata modificata: %s", diffRanges(radio, got))
	}
}

// impostazioni: posizioni trascritte da Get/SetFunConfigInfo_Part1 e _Part2 dei
// due programmi (blocco 0x9000; Part2 parte da 0x9020), numero delle scelte dal
// modulo usato in lettura
func TestBT11v126Settings(t *testing.T) {
	type pos struct {
		off  int
		mask byte
		n    int
	}
	part1 := map[string]pos{"sql": {0x00, 0x0F, 10}, "save": {0x01, 0x0F, 4}, "vox": {0x02, 0x0F, 10}, "backlight": {0x03, 0x0F, 9},
		"tdr": {0x04, 0x0F, 2}, "tot": {0x05, 0x0F, 6}, "beep": {0x06, 0x0F, 2}, "voice": {0x07, 0x0F, 2}, "lang": {0x08, 0x0F, 2},
		"dtmf_st": {0x09, 0x0F, 4}, "scan_mode": {0x0A, 0x0F, 3}, "ptt_id": {0x0B, 0x0F, 4}, "id_delay": {0x0C, 0x0F, 7},
		"disp_a": {0x0D, 0x0F, 3}, "disp_b": {0x0E, 0x0F, 3}, "busy_lock": {0x0F, 0x0F, 2}, "auto_lock": {0x10, 0x0F, 4},
		"alarm_mode": {0x11, 0x0F, 3}, "alarm_sound": {0x12, 0x0F, 2}, "tdr_tx": {0x13, 0x0F, 3}, "tail": {0x14, 0x0F, 2},
		"rp_ste": {0x15, 0x0F, 11}, "rpt_rl": {0x16, 0x0F, 11}, "tx_end": {0x17, 0x0F, 3}, "fm_radio": {0x19, 0x0F, 2},
		"mode_a": {0x1A, 0x0F, 2}, "mode_b": {0x1A, 0xF0, 2}, "key_lock": {0x1B, 0x0F, 2}, "power_msg": {0x1C, 0x0F, 2},
		"bt_prog": {0x1D, 0x0F, 2}, "r_tone": {0x1E, 0x0F, 4}, "vox_delay": {0x20, 0x0F, 16}, "menu_time": {0x21, 0x0F, 11},
		"weather_ch": {0x27, 0x0F, 10}, "denoise": {0x32, 0x0F, 2}}
	with := func(extra map[string]pos) map[string]pos {
		out := map[string]pos{}
		for k, v := range part1 {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	for _, tc := range []struct {
		id   string
		want map[string]pos
	}{
		{"rd-gs10bpro", with(map[string]pos{"pf2_short": {0x2B, 0x0F, 8}, "pf2_long": {0x2C, 0x0F, 9}, "pf3_short": {0x2D, 0x0F, 8}, "pf3_long": {0x2E, 0x0F, 8}})},
		{"bj-8300-990", with(map[string]pos{"remote_kill": {0x24, 0x0F, 2}, "divide_ch": {0x28, 0x0F, 2},
			"pf2_short": {0x2B, 0x0F, 8}, "pf2_long": {0x2C, 0x0F, 7}, "pf3_short": {0x2D, 0x0F, 7}})},
	} {
		if !settingsAvailable(tc.id) {
			t.Fatalf("%s: impostazioni non disponibili", tc.id)
		}
		img := bytes.Repeat([]byte{0xFF}, 0xC300)
		tb := settingsFor(tc.id, img)
		if tb == nil || len(tb.Settings) != len(tc.want) {
			t.Fatalf("%s: %d impostazioni invece di %d", tc.id, len(tb.Settings), len(tc.want))
		}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			w, ok := tc.want[d.Key]
			if !ok || d.ItLabel == "" || d.ItDesc == "" || d.Group == "" {
				t.Fatalf("%s: impostazione inattesa o senza nome: %+v", tc.id, d)
			}
			offs, masks, err := d.Enc.fields()
			if err != nil || len(offs) != 1 || offs[0] != 0x9000+w.off || masks[0][0] != w.mask {
				t.Fatalf("%s %s: posizione %#x maschera %X (attese %#x %X)", tc.id, d.Key, offs, masks, 0x9000+w.off, w.mask)
			}
			n := len(d.Options)
			if d.Type == "bool" {
				n = 2
			}
			if n != w.n {
				t.Fatalf("%s %s: %d scelte invece di %d", tc.id, d.Key, n, w.n)
			}
			// ogni scelta si scrive e si rilegge; gli altri bit del byte restano
			img[offs[0]] = 0xFF
			for v := 0; v < n; v++ {
				val := strconv.Itoa(v)
				if d.Type == "bool" {
					val = strconv.FormatBool(v == 1)
				}
				if err := writeSetting(d, img, val); err != nil {
					t.Fatalf("%s %s=%s: %v", tc.id, d.Key, val, err)
				}
				raw := v
				if d.Key == "fm_radio" {
					raw = 1 - v // il programma mostra ON per 0
				}
				sh := 0
				if w.mask == 0xF0 {
					sh = 4
				}
				if got := int(img[offs[0]]&w.mask) >> sh; got != raw {
					t.Fatalf("%s %s=%s: byte %02X", tc.id, d.Key, val, img[offs[0]])
				}
				if img[offs[0]]|w.mask != 0xFF {
					t.Fatalf("%s %s: bit fuori dalla maschera modificati (%02X)", tc.id, d.Key, img[offs[0]])
				}
				if back, err := readSetting(d, img); err != nil || back != val {
					t.Fatalf("%s %s: riletto %q invece di %q (%v)", tc.id, d.Key, back, val, err)
				}
			}
		}
	}
	// uscita automatica dal menu: l'ultima scelta è 60 secondi (non 55), in tutti
	// i programmi della famiglia
	for _, id := range []string{"rd-gs10bpro", "bj-8300-990", "bj-9000", "bj-9000-960", "rtl-rt950"} {
		tb := settingsFor(id, make([]byte, 0xD300))
		found := false
		for _, d := range tb.Settings {
			if d.Key == "menu_time" {
				found = true
				if len(d.ItOpts) != 11 || d.ItOpts[9] != "50 s" || d.ItOpts[10] != "60 s" {
					t.Fatalf("%s: uscita dal menu %v", id, d.ItOpts)
				}
			}
		}
		if !found {
			t.Fatalf("%s: uscita dal menu assente", id)
		}
	}
	for _, b := range bajSetDefs {
		if b.key == "menu_time" && (len(b.opts) != 11 || b.opts[10] != "60 s") {
			t.Fatalf("BJ-8300 (file .dat): uscita dal menu %v", b.opts)
		}
	}

	// BJ-9000 con 990 canali: le cinque voci del programma v1.3.3 (posizioni di
	// Get/SetFunConfigInfo_Part2, byte 4, e _Part3, byte 5-8); non nelle altre versioni
	want := map[string][2]int{"pwr_delay": {0x9024, 15}, "fm_rx_break": {0x9045, 2}, "breath_led": {0x9046, 2}, "noaa_alarm": {0x9047, 2}, "fm_backlight": {0x9048, 2}}
	for _, id := range []string{"bj-9000", "bj-9000-960", "rtl-rt950"} {
		tb := settingsFor(id, make([]byte, 0xD300))
		n := 0
		for i := range tb.Settings {
			d := &tb.Settings[i]
			w, ok := want[d.Key]
			if !ok {
				continue
			}
			n++
			offs, masks, err := d.Enc.fields()
			cnt := len(d.Options)
			if d.Type == "bool" {
				cnt = 2
			}
			if err != nil || len(offs) != 1 || offs[0] != w[0] || masks[0][0] != 0x0F || cnt != w[1] || d.ItDesc == "" {
				t.Fatalf("%s %s: %v %X %d scelte", id, d.Key, offs, masks, cnt)
			}
			if d.Key == "pwr_delay" && (d.ItOpts[0] != "0.2 s" || d.ItOpts[14] != "3.0 s") {
				t.Fatalf("ritardo all'accensione: %v", d.ItOpts)
			}
		}
		if (id == "bj-9000") != (n == len(want)) || (id != "bj-9000" && n != 0) {
			t.Fatalf("%s: %d voci del programma v1.3.3", id, n)
		}
	}
	// «Channel Divide»: la descrizione non cita più le 15 zone nelle versioni da 10
	for _, id := range []string{"bj-9000", "bj-8300-990"} {
		for _, d := range settingsFor(id, make([]byte, 0xD300)).Settings {
			if d.Key == "divide_ch" && (strings.Contains(d.ItDesc, "15") || len(d.ItOpts) != 2 || !strings.Contains(d.ItOpts[0], "10 × 99")) {
				t.Fatalf("%s: divisione dei canali %q %v", id, d.ItDesc, d.ItOpts)
			}
		}
	}
}

// scheda «Bajeton 8300» (file .dat): alla fine della lettura riconosce il
// firmware nuovo dai canali oltre il 960 o dai nomi delle zone; l'immagine letta
// non cambia e una radio che non risponde a quelle letture non dà errore
func TestBtProbeNewFirmware(t *testing.T) {
	m := genByID["bj-8300-990"]
	img := btSampleImage(t)
	read := func(prep func(e *btEmu)) (btInfo, *btEmu) {
		e := newBtEmu(img)
		prep(e)
		defer withEmu(t, e.serve)()
		got, info, err := btReadImage(context.Background(), "COM1", noProg, nil)
		if err != nil {
			t.Fatal(err)
		}
		waitFor(func() bool { return e.ended })
		if !bytes.Equal(got, img) || info.ID != "BT-11U  V1.00   " {
			t.Fatal("lettura diversa dall'immagine")
		}
		return info, e
	}
	// radio senza canali oltre il 960 e senza nomi delle zone: nessun avviso
	info, e := read(func(*btEmu) {})
	if info.NewFw != "" {
		t.Fatalf("avviso senza motivo: %q", info.NewFw)
	}
	// le letture in più sono 15 blocchi dei canali 961-990 e 4 dei nomi delle zone
	extra := 0
	for _, l := range e.log {
		if strings.HasPrefix(l, "R:78") || strings.HasPrefix(l, "R:79") || strings.HasPrefix(l, "R:7A") || strings.HasPrefix(l, "R:7B") || strings.HasPrefix(l, "R:C8") {
			extra++
		}
	}
	if extra != 19 || e.log[len(e.log)-1] != "E" {
		t.Fatalf("%d letture in più, ultimo comando %q", extra, e.log[len(e.log)-1])
	}
	// un canale oltre il 960
	info, _ = read(func(e *btEmu) { rt950Chan(t, m, e.mem, 969, 439000000, "NUOVO") })
	if info.NewFw != "ha un canale oltre il 960" {
		t.Fatalf("canale 970: %q", info.NewFw)
	}
	info, _ = read(func(e *btEmu) {
		rt950Chan(t, m, e.mem, 960, 439000000, "A")
		rt950Chan(t, m, e.mem, 989, 439100000, "B")
	})
	if info.NewFw != "ha 2 canali oltre il 960" {
		t.Fatalf("due canali: %q", info.NewFw)
	}
	// nomi delle zone
	info, _ = read(func(e *btEmu) { copy(e.mem[0xC810:], "ZONA 2") })
	if info.NewFw != "ha i nomi delle zone" {
		t.Fatalf("nomi delle zone: %q", info.NewFw)
	}
	// dati che non sono canali dopo il 960 (firmware precedente): nessun avviso,
	// e i nomi delle zone non si leggono nemmeno
	info, e = read(func(e *btEmu) {
		for a := 0x7803; a < 0x7900; a += 7 {
			e.mem[a] = 0x3C
		}
		copy(e.mem[0xC810:], "ZONA 2")
	})
	if info.NewFw != "" {
		t.Fatalf("firmware precedente: %q", info.NewFw)
	}
	for _, l := range e.log {
		if strings.HasPrefix(l, "R:C8") {
			t.Fatal("nomi delle zone letti da una radio con il firmware precedente")
		}
	}
	// radio che non risponde alle letture da 0xC800: nessun errore, nessun avviso
	info, e = read(func(e *btEmu) { e.oldFw = true })
	if info.NewFw != "" || e.log[len(e.log)-1] != "E" {
		t.Fatalf("radio che non risponde: %q, ultimo comando %q", info.NewFw, e.log[len(e.log)-1])
	}
	// il testo dell'avviso nomina la scheda giusta
	if !strings.Contains(fmt.Sprintf(btNewFwNote, "ha un canale oltre il 960"), genByID["bj-8300-990"].Name) {
		t.Fatal("l'avviso non nomina la scheda del firmware nuovo")
	}
}

// l'avviso del firmware nuovo arriva una volta nella lettura e una volta nella
// scrittura con la copia di sicurezza (non con la rilettura di verifica); senza
// firmware nuovo non c'è
func TestCableBajetonNewFirmwareNote(t *testing.T) {
	count := func(st map[string]interface{}) int {
		n := 0
		notes, _ := st["notes"].([]interface{})
		for _, x := range notes {
			if s, _ := x.(string); strings.Contains(s, "firmware nuovo") {
				n++
			}
		}
		return n
	}
	for _, newFw := range []bool{false, true} {
		radio := newBtEmu(btSampleImage(t))
		if newFw {
			rt950Chan(t, genByID["bj-8300-990"], radio.mem, 975, 439000000, "NUOVO")
		}
		done := withEmu(t, radio.serve)
		c := newAPIClient(t)
		c.must("cable/start", map[string]interface{}{"op": "read", "radio": "baj", "port": "COM4"})
		st, _ := c.wait("")
		want := 0
		if newFw {
			want = 1
		}
		if st["state"] != "done" || count(st) != want {
			t.Fatalf("lettura (firmware nuovo = %v): %v", newFw, st["notes"])
		}
		chs := st["result"].(map[string]interface{})["radio"].(map[string]interface{})["channels"].([]interface{})
		chs[0].(map[string]interface{})["name"] = "CAVO"
		c.must("cable/start", map[string]interface{}{"op": "write", "radio": "baj", "port": "COM4", "backup": true, "verify": true, "channels": chs})
		st, _ = c.wait("")
		if st["state"] != "done" || count(st) != want {
			t.Fatalf("scrittura (firmware nuovo = %v): %v %v", newFw, st["state"], st["notes"])
		}
		done()
	}
}
