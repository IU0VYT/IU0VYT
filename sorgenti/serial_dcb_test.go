// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"strings"
	"testing"
)

// TestDCBParity (4.9): con la parità pari il bit fParity resta impostato, e
// serial_windows.go non riscrive Flags dopo averlo calcolato.
func TestDCBParity(t *testing.T) {
	if f := dcbFlags(true, false); f&dcbFParity == 0 || f&dcbFBinary == 0 || f&dcbDTREnable == 0 || f&dcbRTSEnable == 0 {
		t.Fatalf("parità pari: Flags %#x", f)
	}
	if f := dcbFlags(false, false); f&dcbFParity != 0 || f != dcbFlagsSimple {
		t.Fatalf("senza parità: Flags %#x", f)
	}
	if dcbFlags(true, false)&dcbFlowMask != 0 {
		t.Fatal("controllo di flusso attivo")
	}
	// 4.9: RTS spento (Retevis C2): gli altri bit restano
	if f := dcbFlags(true, true); f&dcbRTSEnable != 0 || f&dcbFParity == 0 || f&dcbDTREnable == 0 || f&dcbFBinary == 0 {
		t.Fatalf("RTS spento: Flags %#x", f)
	}
	src, err := os.ReadFile("serial_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if strings.Count(s, "dcb.Flags") != 1 || !strings.Contains(s, "dcb.Flags = dcbFlags(p.even, p.noRTS)") {
		t.Fatal("serial_windows.go deve impostare Flags una volta sola, con dcbFlags")
	}
}
