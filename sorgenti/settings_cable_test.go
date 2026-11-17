//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"testing"
)

// cableImage: memoria letta dal driver dalla radio simulata che contiene img
func cableImage(t *testing.T, m *genModel, img []byte) []byte {
	t.Helper()
	e := emuFor(m, append([]byte{}, img...))
	if e == nil {
		return nil
	}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	got, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatalf("%s: lettura dalla radio simulata: %v", m.ID, err)
	}
	return got
}

// 4.43: le impostazioni devono esserci anche dopo aver letto la radio via cavo.
// Fino alla 4.42 per 33 radio di CHIRP (UV-17 Pro e parenti, RT22, TK-2140/3140,
// MP31, HA1G…) e per le 8 «gemelle» della 4.42 la tabella era data solo alle
// immagini lunghe quanto quella di CHIRP, che ha in coda il nome del modello o
// altre zone: la memoria letta dal nostro driver è più corta e la scheda
// «Impostazioni» rispondeva che la memoria non corrispondeva. Le prove usavano
// le immagini di CHIRP e non se ne accorgevano: qui si usa la memoria che il
// driver legge davvero.
func TestSettingsAfterCableRead(t *testing.T) {
	chirpDir(t)
	risk := 0
	for _, m := range genModels {
		if m.NoChirp || !settingsAvailable(m.ID) {
			continue
		}
		full := chirpSettings()[m.RClass]
		if full == nil || genMinSize(m) >= full.Size {
			continue // l'immagine del driver è lunga almeno quanto la tabella: nessun rischio
		}
		got := cableImage(t, m, loadImg(t, chirpImageFor(t, m)))
		if got == nil {
			continue
		}
		risk++
		tb := settingsFor(m.ID, got)
		if tb == nil {
			t.Errorf("%s (%s): nessuna tabella per la memoria letta via cavo (%d byte, tabella di CHIRP %d)", m.ID, m.RClass, len(got), full.Size)
			continue
		}
		for i := range tb.Settings {
			if end := setDefEnd(&tb.Settings[i]); end > len(got) {
				t.Errorf("%s: la voce %s arriva a %d, oltre la memoria letta (%d)", m.ID, tb.Settings[i].Key, end, len(got))
			}
		}
		// nessuna voce di CHIRP sta nella coda che il driver non legge
		if len(tb.Settings) != len(full.Settings) {
			t.Errorf("%s (%s): %d voci invece di %d", m.ID, m.RClass, len(tb.Settings), len(full.Settings))
		}
	}
	if risk < 30 {
		t.Fatalf("solo %d radio provate", risk)
	}
	// le gemelle (4.42 e 4.43): stessa cosa, con la memoria del loro driver
	for id := range twinSettings {
		m := genByID[id]
		got := cableImage(t, m, noChirpBase(t, m))
		if got == nil {
			t.Fatalf("%s: nessuna radio simulata", id)
		}
		tb, all := settingsFor(id, got), twinTablesMap()[id]
		if tb == nil || all == nil || len(tb.Settings) != len(all.Settings) {
			t.Errorf("%s: tabella assente o incompleta per la memoria letta via cavo (%d byte)", id, len(got))
		}
	}
	// una memoria più corta della mappa della radio resta senza tabella
	m := genByID["bf-uv17pro"]
	if settingsFor(m.ID, make([]byte, genMinSize(m)-1)) != nil {
		t.Fatal("tabella data a una memoria troppo corta")
	}
}
