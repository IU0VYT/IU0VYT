//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.13: firmware IJV con le radio simulate

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"
	"testing"
)

func ijvWithEmu(t *testing.T, e *ijvEmu, run func()) {
	t.Helper()
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	run()
}

func ijvNewEmu(id string, mod func(img []byte)) (*genModel, *ijvEmu) {
	m := genByID[id]
	img := ijvSynthImage(m)
	if mod != nil {
		mod(img)
	}
	return m, m.Fam.(ijvFam).emu(m, img).(*ijvEmu)
}

func TestIJVFamily(t *testing.T) {
	for fw, want := range map[string]string{
		"IJV V6 6.1": "V6", "IJV FW5 5.2": "FW5", "FW5.1.3": "FW5", "V-X4.0.4": "V4", "IJV V4.0.4": "V4", "FW4 4.1": "V4",
		"V-X3.60": "V3X", "X3.61": "V3X", "V-N3.60": "V3", "IJV 3.55": "V3",
		// firmware originali e liberi: non sono IJV
		"k5_2.01.32": "", "3.00.07": "", "4.00.07": "", "5.00.05": "", "7.00.07": "", "7.03.01": "", "EGZUMER v0.22": "",
		"F4HWN v4.3": "", "F4HWN V4.2": "", "LOSEHU131HS": "", "LOSEHU124H": "", "OSFW-bd90ca3": "", "": "",
	} {
		if got := ijvFamily(fw); got != want {
			t.Errorf("%q: famiglia %q, attesa %q", fw, got, want)
		}
	}
}

// record del canale: posizioni dei campi come nei moduli CHIRP e in RadioManager
func TestIJVRecord(t *testing.T) {
	for _, id := range []string{"qs-uvk5-ijv3", "qs-uvk5-ijv3x", "qs-uvk5v3-ijv4", "qs-uvk5v3-ijv5"} {
		m := genByID[id]
		f := m.Fam.(ijvFam)
		img := ijvSynthImage(m)
		c := &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", RxTone: "D023N", Power: 1, Narrow: true, Scan: false, Name: "RPT TEST"}
		if err := f.Set(m, img, 4, c); err != nil {
			t.Fatal(id, err)
		}
		r := f.rec(img, 4)
		if string(r[:10]) != "RPT TEST  " || !bytes.Equal(r[10:15], bytes.Repeat([]byte{0xEE}, 5)) {
			t.Fatalf("%s: nome o selettiva % X", id, r[:15])
		}
		if r[15] != 2 || binary.LittleEndian.Uint32(r[16:]) != 14560000 || binary.LittleEndian.Uint32(r[20:]) != 60000 {
			t.Fatalf("%s: banda/frequenza/scostamento % X", id, r[15:24])
		}
		// RX DCS 023 (indice 0, tipo 2), TX CTCSS 88,5 (indice 8, tipo 1)
		if r[24] != 0 || r[25] != 8 || r[26] != 0x12 {
			t.Fatalf("%s: toni % X", id, r[24:27])
		}
		// shift −, scansione esclusa, FM; potenza media, N 10k; passo 12,5 kHz, squelch 1
		if r[27] != 0x22 || r[28] != 1<<5|6<<1 || r[30] != 0x1B {
			t.Fatalf("%s: attributi % X", id, r[27:31])
		}
		g := f.Get(m, img, 4)
		if g.Rx != c.Rx || g.Tx != c.Tx || g.TxTone != c.TxTone || g.RxTone != c.RxTone || g.Power != 1 || !g.Narrow || g.Scan || g.Name != c.Name {
			t.Fatalf("%s: riletto %+v", id, g)
		}
		// campi IJV conservati modificando il canale (gruppo, AGC, compander,
		// selettiva, blocco TX, canale occupato)
		r[15] = 0x52
		r[29] = 0x2B
		r[10] = 0x12
		r[27] |= 0x80
		r[28] |= 0x80
		r[31] = 0x31
		c2 := *g
		c2.Tx, c2.TxTone, c2.Power, c2.Narrow, c2.Scan = c2.Rx, "", 0, false, true
		if err := f.Set(m, img, 4, &c2); err != nil {
			t.Fatal(err)
		}
		if r[15] != 0x52 || r[29] != 0x2B || r[10] != 0x12 || r[31] != 0x31 || r[27] != 0x80 || r[28] != 0x80|2<<5 {
			t.Fatalf("%s: campi IJV non conservati % X", id, r)
		}
		// aeronautica in AM, solo ricezione
		air := &Chan{Rx: 118100000, Tx: 118100000, TxOff: true, Scan: true, Name: "AIR"}
		if err := f.Set(m, img, 5, air); err != nil {
			t.Fatal(err)
		}
		r5 := f.rec(img, 5)
		if (r5[27]>>2)&7 != ijvModAM || r5[27]&3 != 2 || !bytes.Equal(r5[16:20], r5[20:24]) {
			t.Fatalf("%s: canale aeronautico % X", id, r5)
		}
		if g := f.Get(m, img, 5); g == nil || !g.TxOff || g.Rx != air.Rx {
			t.Fatalf("%s: canale aeronautico riletto %+v", id, g)
		}
		// cancellazione: 0xFF e selettiva vuota
		if err := f.Set(m, img, 5, nil); err != nil || f.Get(m, img, 5) != nil || r5[10] != 0xEE || r5[0] != 0xFF {
			t.Fatalf("%s: cancellazione % X", id, r5)
		}
		// fuori gamma
		if err := f.Set(m, img, 6, &Chan{Rx: 1400000000, Tx: 1400000000}); err == nil {
			t.Fatalf("%s: 1400 MHz accettati", id)
		}
	}
}

// V4 senza il protocollo in chiaro: si passa a quello di fabbrica; FW5: solo in chiaro
func TestIJVProtocolModes(t *testing.T) {
	m, e := ijvNewEmu("qs-uvk5v3-ijv4", nil)
	e.noPlain = true
	ijvWithEmu(t, e, func() {
		got, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
		if err != nil || !bytes.Equal(got, e.image()) || e.lastMode != "fabbrica" {
			t.Fatalf("V4 di fabbrica: %v (%s)", err, e.lastMode)
		}
	})
	m, e = ijvNewEmu("qs-uvk5v3-ijv5", nil)
	e.beacon = true
	ijvWithEmu(t, e, func() {
		got, note, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
		if err != nil || !bytes.Equal(got, e.image()) || e.lastMode != "chiaro" || !strings.Contains(note, "banco 3") {
			t.Fatalf("FW5: %v (%s) %q", err, e.lastMode, note)
		}
	})
}

// prima di scrivere si controlla il banco scelto sulla radio
func TestIJVBankMismatch(t *testing.T) {
	for _, id := range []string{"qs-uvk5v3-ijv4", "qs-uvk5v3-ijv5"} {
		m, e := ijvNewEmu(id, nil)
		ijvWithEmu(t, e, func() {
			img, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
			if err != nil {
				t.Fatal(err)
			}
			if err := m.Fam.Set(m, img, 0, &Chan{Rx: 433500000, Tx: 433500000, Scan: true, Name: "UHF"}); err != nil {
				t.Fatal(err)
			}
			// sulla radio si sceglie un altro banco
			if e.kind == ijvFW5 {
				e.flash[ijvCfg5Bank] = 5
			} else {
				e.img[0x0B] = 5
			}
			before := e.image()
			_, err = m.Fam.Upload(context.Background(), m, "COM7", img, noProg)
			if err == nil || !strings.Contains(err.Error(), "banco 6") || !bytes.Equal(before, e.image()) || e.writes != 0 {
				t.Fatalf("%s: %v (scritture %d)", id, err, e.writes)
			}
		})
	}
}

// solo i blocchi cambiati; impostazioni sì, tarature mai
func TestIJVWriteOnlyChanged(t *testing.T) {
	for _, id := range []string{"qs-uvk5v3-ijv4", "qs-uvk5v3-ijv5"} {
		m, e := ijvNewEmu(id, nil)
		ijvWithEmu(t, e, func() {
			img, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
			if err != nil {
				t.Fatal(err)
			}
			want := append([]byte{}, img...)
			if err := m.Fam.Set(m, want, 499, &Chan{Rx: 446006250, Tx: 446006250, Narrow: true, Scan: true, Name: "PMR 1"}); err != nil {
				t.Fatal(err)
			}
			want[0x0D] = 3    // retroilluminazione
			want[0x1E10] ^= 1 // taratura nell'immagine: non deve arrivare alla radio
			note, err := m.Fam.Upload(context.Background(), m, "COM7", want, noProg)
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			got := e.image()
			exp := append([]byte{}, want...)
			exp[0x1E10] = img[0x1E10]
			if id == "qs-uvk5v3-ijv4" {
				// riscrittura a 0x1000 e 0x2000 (stessi dati): nessuna differenza
				if e.writes != 4 {
					t.Fatalf("V4: %d scritture, attese 4 (canale, impostazione, chiusura dei settori)", e.writes)
				}
			} else if e.writes != 2 || e.flushes != 1 {
				t.Fatalf("FW5: %d scritture e %d salvataggi", e.writes, e.flushes)
			}
			if !bytes.Equal(got, exp) {
				t.Fatalf("%s: memoria della radio diversa da quella attesa", id)
			}
			if !strings.Contains(note, "blocchi cambiati e scritti: 2") {
				t.Fatalf("%s: nota %q", id, note)
			}
		})
	}
}

func TestIJVFW5Refused(t *testing.T) {
	m, e := ijvNewEmu("qs-uvk5v3-ijv5", nil)
	e.refuse = true
	ijvWithEmu(t, e, func() {
		img, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
		if err != nil {
			t.Fatal(err)
		}
		_ = m.Fam.Set(m, img, 0, &Chan{Rx: 145500000, Tx: 145500000, Scan: true})
		if _, err := m.Fam.Upload(context.Background(), m, "COM7", img, noProg); err == nil || !strings.Contains(err.Error(), "abilita sulla radio") {
			t.Fatalf("errore atteso, ottenuto %v", err)
		}
	})
}

// scheda sbagliata: il programma indica quella giusta e non scrive
func TestIJVWrongCard(t *testing.T) {
	_, e := ijvNewEmu("qs-uvk5v3-ijv5", nil)
	m4 := genByID["qs-uvk5v3-ijv4"]
	ijvWithEmu(t, e, func() {
		_, _, err := m4.Fam.Download(context.Background(), m4, "COM7", noProg)
		if err == nil || !strings.Contains(err.Error(), genByID["qs-uvk5v3-ijv5"].Name) {
			t.Fatalf("errore atteso, ottenuto %v", err)
		}
	})
	// scheda del firmware originale con una radio IJV V4 (risponde al saluto di fabbrica)
	_, e = ijvNewEmu("qs-uvk5v3-ijv4", nil)
	k5 := genByID["qs-uvk5"]
	ijvWithEmu(t, e, func() {
		_, _, err := k5.Fam.Download(context.Background(), k5, "COM7", noProg)
		if err == nil || !strings.Contains(err.Error(), m4.Name) {
			t.Fatalf("errore atteso, ottenuto %v", err)
		}
	})
}

func TestIJVV6(t *testing.T) {
	m, e := ijvNewEmu("qs-uvk5v3-ijv4", nil)
	e.v6 = true
	ijvWithEmu(t, e, func() {
		_, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
		if err == nil || !strings.Contains(err.Error(), "IJV V6 (UV-K5 V3, versione 6.1.2)") {
			t.Fatalf("errore atteso, ottenuto %v", err)
		}
	})
}

// impostazioni: valori nella memoria, nessuna nelle tarature
func TestIJVSettings(t *testing.T) {
	for id, tb := range ijvSetTables {
		m := genByID[id]
		img := ijvSynthImage(m)
		if settingsFor(id, img) != tb || !settingsAvailable(id) {
			t.Fatalf("%s: impostazioni non disponibili", id)
		}
		keys := map[string]bool{}
		for i := range tb.Settings {
			d := &tb.Settings[i]
			if keys[d.Key] {
				t.Fatalf("%s: chiave ripetuta %s", id, d.Key)
			}
			keys[d.Key] = true
			var offs []int
			if d.Enc.Kind == "str" {
				offs = d.Enc.Offs
			} else {
				o, _, _ := d.Enc.fields()
				offs = o
			}
			for _, o := range offs {
				if o >= ijvSettingEnd && o < ijvFixedEnd {
					t.Fatalf("%s: %s nelle tarature (0x%04X)", id, d.Key, o)
				}
			}
		}
		set := func(k, v string) {
			for i := range tb.Settings {
				if tb.Settings[i].Key == k {
					if err := writeSetting(&tb.Settings[i], img, v); err != nil {
						t.Fatalf("%s %s: %v", id, k, err)
					}
					got, err := readSetting(&tb.Settings[i], img)
					if err != nil || got != v {
						t.Fatalf("%s %s: riletto %q (%v)", id, k, got, err)
					}
					return
				}
			}
			t.Fatalf("%s: impostazione %s assente", id, k)
		}
		set("backlight", "7")
		set("tx_enable", "false")
		set("upconverter", "2")
		set("qrz", "IZ0ABC")
		set("group_name_3", "AIR")
		if img[0x0D] != 7 || img[0x17]&1 != 0 || img[0x1F]&0x30 != 0x20 || string(img[0x170:0x178]) != "IZ0ABC  " ||
			string(img[m.Fam.(ijvFam).groupNames()+24:][:8]) != "AIR     " {
			t.Fatalf("%s: memoria % X", id, img[:0x20])
		}
		if m.Fam.(ijvFam).kind == ijvV4 || m.Fam.(ijvFam).kind == ijvFW5 {
			set("battery_saver", "2")
			set("ptt_toggle", "true")
			set("bank_name_3", "HOME")
			if img[0x27] != 0x0A || string(img[ijvBankNames+16:][:8]) != "HOME    " {
				t.Fatalf("%s: memoria 0x27 %02X", id, img[0x27])
			}
		}
	}
}

// file .bin con la versione di un firmware IJV: avviso e scheda IJV per i canali
func TestIJVFirmwareFile(t *testing.T) {
	for ver, want := range map[string]string{"*IJV V4.0.4": "qs-uvk5v3-ijv4", "*IJV FW5 5.2": "qs-uvk5v3-ijv5", "2.01.26": "qs-uvk5"} {
		raw, _ := k5TestFirmware(ver, 0x8000)
		img := analyzeK5("prova.bin", "", raw)
		if len(img.Info.Models) == 0 || img.Info.Models[0] != want {
			t.Fatalf("%s: schede %v", ver, img.Info.Models)
		}
		if strings.Contains(want, "ijv") && !img.Info.Warnings {
			t.Fatalf("%s: avviso atteso", ver)
		}
	}
}
