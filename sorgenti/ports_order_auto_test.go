//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.5: porte in ordine progressivo su ogni sistema

import (
	"reflect"
	"testing"
)

// TestPortOrderAuto: l'identificazione automatica cerca e mostra le porte in
// ordine progressivo, anche se il sistema le elenca in disordine
func TestPortOrderAuto(t *testing.T) {
	c := newAPIClient(t)
	a := c.app
	oldL := listPortsFunc
	t.Cleanup(func() { listPortsFunc = oldL })
	listPortsFunc = func() []PortInfo {
		var l []PortInfo
		for _, n := range []string{"COM10", "COM4", "COM1", "COM7", "COM3", "COM2", "COM6", "COM5"} {
			l = append(l, PortInfo{n, "USB-SERIAL CH340 (" + n + ")"})
		}
		return l
	}
	want := []string{"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM10"}
	r := c.must("cable/ports", nil)
	var got []string
	for _, p := range r["ports"].([]interface{}) {
		got = append(got, p.(map[string]interface{})["name"].(string))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cable/ports: %v", got)
	}
	a.auto.mu.Lock()
	a.auto.ports = map[string]*autoPort{}
	for _, n := range want {
		a.auto.ports[n] = &autoPort{name: n}
	}
	for k := 0; k < 20; k++ {
		got = got[:0]
		for _, p := range a.autoAllLocked() {
			got = append(got, p.name)
		}
		if !reflect.DeepEqual(got, want) {
			a.auto.mu.Unlock()
			t.Fatalf("ordine delle ricerche: %v", got)
		}
	}
	a.auto.ports = nil
	a.auto.mu.Unlock()
}
