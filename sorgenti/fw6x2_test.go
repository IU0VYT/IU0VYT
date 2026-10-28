// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Controllo dei file di aggiornamento BTECH DMR-6X2 / DMR-6X2 PRO (3.8):
// pacchetti ufficiali (BTECH_6X2_DIR = cartella con DMR-6X2-221b e
// DMR-6X2-PRO-121e estratti) e file danneggiati.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFw6x2Check(t *testing.T) {
	d := os.Getenv("BTECH_6X2_DIR")
	if d == "" {
		t.Skip("BTECH_6X2_DIR non impostata")
	}
	want := map[string]string{
		"DMR_6x2_V2.21a_20250319.spi":   "main|DMR-6X2|2.21a|19/03/2025",
		"DMR_6X2II_V1.21a_20250319.spi": "main|DMR-6X2 PRO|1.21a|19/03/2025",
		"DMR_6X2_PIC_V2.00.spi":         "icons||2.00|",
		"BHT_BOARD_V1.06_20221108.spi":  "board||1.06|08/11/2022",
	}
	seen := map[string]bool{}
	fs, _ := filepath.Glob(filepath.Join(d, "DMR-6X2-*", "DMR-6X2-*", "*", "*.spi"))
	for _, f := range fs {
		in, err := check6x2(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		got := strings.Join([]string{in.Kind, in.Model, in.Version, in.Date}, "|")
		if w := want[filepath.Base(f)]; got != w || len(in.Errors) > 0 {
			t.Errorf("%s: %s (atteso %s), errori %v", filepath.Base(f), got, w, in.Errors)
		}
		seen[filepath.Base(f)] = true
		t.Logf("%s: %s; %v", filepath.Base(f), in.Title, in.Checks)
	}
	if len(seen) != len(want) {
		t.Fatalf("file trovati: %v", seen)
	}
	// file danneggiati: .CDD troncato, modello scambiato, .CDI mancante
	src := filepath.Join(d, "DMR-6X2-221b", "DMR-6X2-221b", "DMR_6X2 Firmware Update 2.21", "DMR_6x2_V2.21a_20250319")
	tmp := filepath.Join(t.TempDir(), "DMR_6x2_V2.21a_20250319")
	cp := func(ext string, f func([]byte) []byte) {
		b, err := os.ReadFile(src + ext)
		if err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(tmp+ext, f(b), 0o644)
	}
	same := func(b []byte) []byte { return b }
	cp(".spi", same)
	cp(".CDI", same)
	cp(".CDD", func(b []byte) []byte { return b[:len(b)-100] })
	if in, err := check6x2(tmp + ".spi"); err != nil || len(in.Errors) == 0 {
		t.Fatalf("CDD troncato non segnalato: %v %+v", err, in)
	}
	cp(".CDD", same)
	cp(".spi", func(b []byte) []byte { copy(b[14:], "D6X2UV2"); return b })
	if in, err := check6x2(tmp + ".spi"); err != nil || len(in.Errors) == 0 {
		t.Fatalf("modello scambiato non segnalato: %v %+v", err, in)
	}
	cp(".spi", same)
	cp(".CDD", func(b []byte) []byte { b[3] = 0x55; return b })
	if in, err := check6x2(tmp + ".spi"); err != nil || len(in.Errors) == 0 {
		t.Fatalf("inizio del firmware errato non segnalato: %v %+v", err, in)
	}
	_ = os.Remove(tmp + ".CDI")
	if _, err := check6x2(tmp + ".spi"); err == nil || !strings.Contains(err.Error(), ".CDI") {
		t.Fatalf("CDI mancante: %v", err)
	}
}
