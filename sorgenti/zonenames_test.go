//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.11: nomi delle zone, con le radio simulate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZoneNameCodec(t *testing.T) {
	gb := &zoneNameSpec{Stride: 16, Max: 12, GB: true, Pad: 0xFF}
	as := &zoneNameSpec{Stride: 16, Max: 12, Pad: 0x00}
	for _, c := range []struct {
		s    *zoneNameSpec
		name string
		ok   bool
	}{
		{gb, "PONTI VHF", true}, {gb, "本地中继", true}, {gb, "123456789012", true}, {gb, "1234567890123", false},
		{gb, "中继中继中继中", false}, {gb, "Città", false}, {gb, "", true},
		{as, "PMR446", true}, {as, "Città", false}, {as, "中继", false}, {as, "ABCDEFGHIJKL", true}, {as, "ABCDEFGHIJKLM", false},
	} {
		rec, err := c.s.encode(c.name)
		if (err == nil) != c.ok {
			t.Errorf("%q: errore %v", c.name, err)
			continue
		}
		if err != nil {
			continue
		}
		if len(rec) != 16 {
			t.Fatalf("record di %d byte", len(rec))
		}
		got, ok := c.s.decode(rec)
		if !ok || got != c.name {
			t.Errorf("%q: riletto %q %v (% X)", c.name, got, ok, rec)
		}
		// riempimento come il programma del costruttore
		n := len(gb2312Enc(c.name))
		for _, b := range rec[n:] {
			if b != c.s.Pad {
				t.Errorf("%q: riempimento % X", c.name, rec)
				break
			}
		}
	}
	// record che non sono nomi
	for _, rec := range [][]byte{
		append([]byte("ABC\xFF\x41"), bytes.Repeat([]byte{0xFF}, 11)...), // dati dopo il riempimento
		append([]byte{0x01, 0x02}, bytes.Repeat([]byte{0xFF}, 14)...),    // caratteri di controllo
		[]byte("ABCDEFGHIJKLMNOP"),                                       // 16 caratteri (al massimo 12)
	} {
		if _, ok := gb.decode(rec); ok {
			t.Errorf("% X accettato come nome", rec)
		}
	}
	// i programmi Radtel scrivono anche 0x00 dopo il nome
	if n, ok := gb.decode(append([]byte("ZONA"), make([]byte, 12)...)); !ok || n != "ZONA" {
		t.Errorf("riempimento 0x00: %q %v", n, ok)
	}
}

// radio simulata della famiglia rt900: lettura e scrittura
func zoneEmuIO(t *testing.T, m *genModel, radio, img []byte) (*rt900Emu, []byte, error) {
	t.Helper()
	e := emuFor(m, radio).(*rt900Emu)
	a, b := newPipe()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) { return a, nil }
	defer func() { openPortFunc = old; _ = a.Close() }()
	go func() { _ = e.serve(b) }()
	ctx := context.Background()
	prog := func(int, int, string) {}
	if img == nil {
		data, _, err := m.Fam.Download(ctx, m, "EMU", prog)
		return e, data, err
	}
	_, err := m.Fam.Upload(ctx, m, "EMU", img, prog)
	return e, nil, err
}

func TestZoneNamesRT950(t *testing.T) {
	for _, id := range []string{"bj-9000", "bj-9000-960", "rtl-rt950", "bj-7700nbt"} {
		t.Run(id, func(t *testing.T) {
			m := genByID[id]
			zs := m.ZoneNames
			if zs == nil || zs.Area != 0x100 || zs.Count != m.Zones {
				t.Fatalf("zone %d, nomi %+v", m.Zones, zs)
			}
			if !m.def().ZoneNames {
				t.Fatal("l'interfaccia non vede i nomi delle zone")
			}
			o := rt900o(m)
			size := max(o.memsize, zs.Addr+zs.Area)
			radio := bytes.Repeat([]byte{0xFF}, size)
			if zs.Pad == 0 {
				for i := zs.Addr; i < zs.Addr+zs.Area; i++ {
					radio[i] = 0
				}
			}
			first := "PONTI"
			if zs.GB {
				first = "本地中继"
			}
			r0, _ := zs.encode(first)
			copy(radio[zs.Addr:], r0)

			// lettura: i nomi arrivano nel file
			_, data, err := zoneEmuIO(t, m, radio, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) < zs.Addr+zs.Area {
				t.Fatalf("letti %d byte: mancano i nomi delle zone", len(data))
			}
			names, mx, ok := zoneNamesOf(m, data)
			if !ok || len(names) != zs.Count || names[0] != first || names[1] != "" || mx != zs.Max {
				t.Fatalf("nomi letti: %q %d %v", names, mx, ok)
			}

			// nomi nuovi: si scrivono
			names[1], names[zs.Count-1] = "PMR446", "LPD"
			img := append([]byte{}, data...)
			if err := setZoneNames(m, img, names); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(img[:zs.Addr], data[:zs.Addr]) || !bytes.Equal(img[zs.Addr+zs.Area:], data[zs.Addr+zs.Area:]) {
				t.Fatal("cambiati byte fuori dall'area dei nomi")
			}
			e, _, err := zoneEmuIO(t, m, radio, img)
			if err != nil {
				t.Fatal(err)
			}
			got, _, _ := zoneNamesOf(m, e.image())
			if strings.Join(got, "|") != strings.Join(names, "|") {
				t.Fatalf("nella radio: %q", got)
			}
			// la zona 1 non cambiata resta byte per byte com'era
			if !bytes.Equal(e.image()[zs.Addr:zs.Addr+16], radio[zs.Addr:zs.Addr+16]) {
				t.Fatal("zona 1 riscritta")
			}

			// nomi non validi: nessuna modifica alla memoria
			for _, bad := range [][]string{append(append([]string{}, names[:zs.Count-1]...), "NOME TROPPO LUNGO"), names[:1]} {
				cp := append([]byte{}, img...)
				if err := setZoneNames(m, cp, bad); err == nil || !bytes.Equal(cp, img) {
					t.Fatalf("%q: %v", bad, err)
				}
			}

			// radio con altri dati nell'area dei nomi: i nomi cambiati non si
			// scrivono e la radio non cambia
			junk := append([]byte{}, radio...)
			for a := zs.Addr + 0x13; a < zs.Addr+0x80; a += 5 {
				junk[a] = 0x01
			}
			e, _, err = zoneEmuIO(t, m, junk, img)
			if err == nil || !strings.Contains(err.Error(), "Nessun dato è stato scritto") {
				t.Fatalf("scrittura con altri dati nell'area: %v", err)
			}
			if !bytes.Equal(e.image()[:len(junk)], junk) {
				t.Fatalf("la radio è stata modificata: %s", diffRanges(junk, e.image()[:len(junk)]))
			}
			// stessa radio, file letto da lei (area uguale): i canali si scrivono,
			// l'area resta com'è
			img2 := append([]byte{}, junk...)
			rt950Chan(t, m, img2, 0, 145500000, "UNO")
			e, _, err = zoneEmuIO(t, m, junk, img2)
			if err != nil {
				t.Fatal(err)
			}
			if c := genGet(m, e.image(), 0); c == nil || c.Name != "UNO" {
				t.Fatalf("canale 1: %+v", c)
			}
			if !bytes.Equal(e.image()[zs.Addr:zs.Addr+zs.Area], junk[zs.Addr:zs.Addr+zs.Area]) {
				t.Fatal("area dei nomi modificata")
			}
			if _, _, ok := zoneNamesOf(m, junk); ok {
				t.Fatal("altri dati mostrati come nomi")
			}
		})
	}
}

// Radtel RT-950: file salvato prima della 4.11 (senza l'area dei nomi)
func TestZoneNamesRT950OldFile(t *testing.T) {
	m := genByID["rtl-rt950"]
	o := rt900o(m)
	zs := m.ZoneNames
	radio := bytes.Repeat([]byte{0xFF}, zs.Addr+zs.Area)
	r0, _ := zs.encode("RADIO")
	copy(radio[zs.Addr:], r0)
	img := bytes.Repeat([]byte{0xFF}, o.memsize)
	rt950Chan(t, m, img, 0, 145500000, "UNO")
	if _, _, ok := zoneNamesOf(m, img); ok {
		t.Fatal("nomi in un file senza l'area")
	}
	if err := setZoneNames(m, append([]byte{}, img...), make([]string, 15)); err == nil {
		t.Fatal("nomi scritti in un file senza l'area")
	}
	e, _, err := zoneEmuIO(t, m, radio, img)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(e.image()[zs.Addr:], radio[zs.Addr:]) {
		t.Fatal("area dei nomi modificata")
	}
	if c := genGet(m, e.image(), 0); c == nil || c.Name != "UNO" {
		t.Fatalf("canale 1: %+v", c)
	}
}

// Radtel RT-920: area scritta sempre (come CHIRP), 10 caratteri
func TestZoneNamesRT920(t *testing.T) {
	m := genByID["rt-rt920"]
	zs := m.ZoneNames
	o := rt900o(m)
	if zs.Area != 0 || zs.Count != 10 || zs.Max != 10 {
		t.Fatalf("%+v", zs)
	}
	written := false
	for _, r := range o.ranges {
		if r[0] <= zs.Addr && zs.Addr+zs.Count*zs.Stride <= r[1] {
			written = true
		}
	}
	if !written {
		t.Fatal("l'area dei nomi non è tra quelle scritte")
	}
	img := bytes.Repeat([]byte{0xFF}, o.memsize)
	names, _, ok := zoneNamesOf(m, img)
	if !ok || len(names) != 10 {
		t.Fatalf("%q %v", names, ok)
	}
	names[0] = "ABCDEFGHIJ"
	if err := setZoneNames(m, img, names); err != nil {
		t.Fatal(err)
	}
	names[1] = "ABCDEFGHIJK"
	if err := setZoneNames(m, img, names); err == nil {
		t.Fatal("11 caratteri accettati")
	}
	if !bytes.Equal(img[zs.Addr:zs.Addr+16], append([]byte("ABCDEFGHIJ"), 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)) {
		t.Fatalf("% X", img[zs.Addr:zs.Addr+16])
	}
}

// radio DMR: nomi delle zone esistenti
func TestZoneNamesDMR(t *testing.T) {
	m := genByID["bf-rd5r"]
	o := rdtyo(m)
	img := bytes.Repeat([]byte{0xFF}, rdtyMemSize)
	copy(img, o.hdr)
	for i := rdtyZoneTab; i < rdtyZoneTab+32; i++ {
		img[i] = 0
	}
	if names, _, ok := zoneNamesOf(m, img); ok {
		t.Fatalf("zone in una memoria vuota: %q", names)
	}
	// zona creata come rdtyZoneAdd (un canale in una memoria senza zone)
	zr := img[rdtyZoneOff(o, 0):][:rdtyZoneLen(o)]
	for k := range zr {
		zr[k] = 0
	}
	copy(zr, append([]byte("Zone 1"), bytes.Repeat([]byte{0xFF}, 10)...))
	zr[16] = 1
	img[rdtyZoneTab] |= 1
	names, max, ok := zoneNamesOf(m, img)
	if !ok || len(names) == 0 || names[0] != "Zone 1" || max != 16 {
		t.Fatalf("%q %d %v", names, max, ok)
	}
	before := append([]byte{}, img...)
	names[0] = "RIPETITORI LAZIO"
	if err := setZoneNames(m, img, names); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := zoneNamesOf(m, img); got[0] != "RIPETITORI LAZIO" {
		t.Fatalf("%q", got)
	}
	for i := range img {
		off := rdtyZoneOff(o, 0)
		if (i < off || i >= off+16) && img[i] != before[i] {
			t.Fatalf("byte 0x%X cambiato fuori dal nome", i)
		}
	}
	for _, bad := range []string{"", "NOME DI DICIASSETTE", "Città"} {
		cp := append([]string{}, names...)
		cp[0] = bad
		if err := setZoneNames(m, append([]byte{}, img...), cp); err == nil {
			t.Fatalf("%q accettato", bad)
		}
	}
	if !m.def().ZoneNames {
		t.Fatal("l'interfaccia non vede i nomi delle zone")
	}

	// DM-32UV con il file di prova di CHIRP
	d := os.Getenv("CHIRP_DIR")
	if d == "" {
		return
	}
	dm := genByID["bf-dm32uv"]
	dimg := loadImg(t, filepath.Join(d, "tests", "images", "Baofeng_DM-32UV.img"))
	zn, _, ok := zoneNamesOf(dm, dimg)
	if !ok || len(zn) == 0 {
		t.Fatalf("DM-32UV: %q %v", zn, ok)
	}
	zn[0] = "LOCALI"
	if err := setZoneNames(dm, dimg, zn); err != nil {
		t.Fatal(err)
	}
	got, _, _ := zoneNamesOf(dm, dimg)
	if got[0] != "LOCALI" {
		t.Fatalf("DM-32UV: %q", got)
	}
	if ch := (dm32Fam{}).zoneList(dimg); len(ch) != len(zn) {
		t.Fatalf("DM-32UV: zone %d invece di %d", len(ch), len(zn))
	}
}
