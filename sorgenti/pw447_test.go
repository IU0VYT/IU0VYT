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
	"testing"
)

// pwThrough447: decisione del 23/11/2026 — con la password di programmazione ci si
// comporta come CHIRP. Una radio (simulata) che ha la password pw alla posizione at
// si legge e si scrive lo stesso: la lettura dà la memoria con la password della
// radio; la memoria letta e modificata (edit) arriva alla radio, la password resta
// quella di prima e una nuova lettura dà la memoria appena scritta.
func pwThrough447(t *testing.T, m *genModel, el genEmu, at int, pw []byte, edit func(img []byte)) {
	t.Helper()
	nop := func(int, int, string) {}
	before := el.image()
	if len(before) < at+len(pw) || !bytes.Equal(before[at:at+len(pw)], pw) {
		t.Fatalf("%s: radio simulata senza la password di prova", m.ID)
	}
	var img []byte
	emuRun(el, func() {
		var err error
		if img, _, err = m.Fam.Download(context.Background(), m, "EMU", nop); err != nil {
			t.Fatalf("%s: lettura con la password: %v", m.ID, err)
		}
	})
	if len(img) < at+len(pw) || !bytes.Equal(img[at:at+len(pw)], pw) {
		t.Fatalf("%s: la password letta non è quella della radio", m.ID)
	}
	orig := append([]byte{}, img...)
	edit(img)
	if bytes.Equal(orig, img) || !bytes.Equal(img[at:at+len(pw)], pw) {
		t.Fatalf("%s: modifica di prova senza effetto, oppure password toccata dalla modifica", m.ID)
	}
	emuRun(el, func() {
		if _, err := m.Fam.Upload(context.Background(), m, "EMU", img, nop); err != nil {
			t.Fatalf("%s: scrittura con la password: %v", m.ID, err)
		}
	})
	after := el.image()
	if !bytes.Equal(after[at:at+len(pw)], pw) {
		t.Fatalf("%s: la scrittura ha cambiato la password della radio", m.ID)
	}
	if bytes.Equal(after, before) {
		t.Fatalf("%s: la scrittura con la password non è arrivata alla radio", m.ID)
	}
	emuRun(el, func() {
		if got, _, err := m.Fam.Download(context.Background(), m, "EMU", nop); err != nil || !bytes.Equal(got, img) {
			t.Fatalf("%s: rilettura con la password diversa dalla memoria scritta (%v)", m.ID, err)
		}
	})
}
