// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"strings"
	"testing"
)

// 4.14: i canali che le radio Yaesu e Icom non possono memorizzare si adattano
// (toni) o si scartano (frequenza) prima della scrittura, con un avviso, invece di
// fermare la conversione o l'inoltro.
func TestAdattaYaesuIcom(t *testing.T) {
	type caso struct {
		id             string
		rx, tx         int64
		txTone, rxTone string
		keep           bool   // il canale resta
		wantTx, wantRx string // toni dopo l'adattamento
		warn           string // testo che l'avviso deve contenere ("" = nessun avviso)
	}
	casi := []caso{
		{"ys-ft4xr", 145600000, 145000000, "88.5", "", true, "88.5", "", ""},
		{"ys-ft4xr", 145600000, 145000000, "D023I", "", true, "", "", "D023I"},
		{"ys-ft4xr", 145600000, 145000000, "88.5", "D023N", true, "88.5", "", "D023N"},
		{"ys-ft65r-s3", 145600000, 145000000, "D023N", "D023N", true, "D023N", "D023N", ""},
		{"ys-ft60", 145600000, 145000000, "88.5", "88.5", true, "88.5", "88.5", ""},
		{"ys-ft60", 145600000, 145000000, "", "88.5", true, "", "", "88.5"},
		{"ys-ft60", 446006250, 446006250, "", "", false, "", "", "non copiato"},
		{"ic-t10", 145600000, 145000000, "88.5", "100.0", true, "88.5", "", "100.0"},
		{"ic-t10", 145600000, 145000000, "D023N", "", true, "", "", "D023N"},
		{"ic-t10", 145600000, 145000000, "D023N", "D023N", true, "D023N", "D023N", ""},
		{"ic-v86", 145600000, 145000000, "88.5", "D023N", true, "88.5", "", "D023N"},
		{"ic-v80", 145600000, 145000000, "", "88.5", true, "", "", "88.5"},
		{"ic-t70", 433000000, 433000000, "D023N", "88.5", true, "", "", "D023N"},
		{"ic-v82", 145600000, 145000000, "88.5", "", true, "88.5", "", ""},
		{"ic-u82", 145600000, 145000000, "88.5", "", false, "", "", "non copiato"},
		{"ic-u82", 433000000, 433000000, "", "", true, "", "", ""},
	}
	for _, k := range casi {
		m := genByID[k.id]
		if m == nil {
			t.Fatalf("modello %s assente", k.id)
		}
		rep := &Report{}
		c := &Chan{Name: "PROVA", Rx: k.rx, Tx: k.tx, TxTone: k.txTone, RxTone: k.rxTone, Scan: true}
		out := genAdapt(c, m, "canale 1", rep)
		w := strings.Join(rep.Warnings, " | ")
		if (out != nil) != k.keep {
			t.Errorf("%s %d %q/%q: resta=%v, atteso %v (%s)", k.id, k.rx, k.txTone, k.rxTone, out != nil, k.keep, w)
			continue
		}
		if out != nil {
			if out.TxTone != k.wantTx || out.RxTone != k.wantRx {
				t.Errorf("%s %q/%q: toni %q/%q, attesi %q/%q", k.id, k.txTone, k.rxTone, out.TxTone, out.RxTone, k.wantTx, k.wantRx)
			}
			// il canale adattato deve essere accettato dal driver
			if err := genTrySet(m, out); err != nil {
				t.Errorf("%s %q/%q: il canale adattato è rifiutato: %v", k.id, k.txTone, k.rxTone, err)
			}
		}
		if k.warn == "" && w != "" {
			t.Errorf("%s %q/%q: avviso non atteso: %s", k.id, k.txTone, k.rxTone, w)
		}
		if k.warn != "" && !strings.Contains(w, k.warn) {
			t.Errorf("%s %q/%q: avviso %q senza %q", k.id, k.txTone, k.rxTone, w, k.warn)
		}
		if c.TxTone != k.txTone || c.RxTone != k.rxTone {
			t.Errorf("%s: il canale di partenza è stato modificato", k.id)
		}
	}
}
