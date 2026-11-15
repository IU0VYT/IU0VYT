//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countPort: conta i byte che la radio simulata manda al programma
type countPort struct {
	Port
	n *atomic.Int64
}

func (c countPort) Write(b []byte) error {
	c.n.Add(int64(len(b)))
	return c.Port.Write(b)
}

// 4.40: controllo breve con la scheda già aperta (identquick.go). Per ogni
// famiglia che lo ha: la prima identificazione legge tutta la memoria; con la
// scheda aperta bastano pochi byte e l'esito è «già identificata»; se sul cavo
// c'è un'altra radio il controllo breve non la scambia per quella della scheda,
// si passa alla lettura normale e per quel modello non si usa più.
func TestIdentQuickOpenTab(t *testing.T) {
	identFast(t)
	identFirstWait, identSilentMax = 80*time.Millisecond, 250*time.Millisecond
	identSpontListen, identSpontListen9600 = 100*time.Millisecond, 100*time.Millisecond
	oldS, oldF, oldL, oldQ := autoSettle, autoRetryFast, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { autoSettle, autoRetryFast, listPortsFunc, autoQuiet = oldS, oldF, oldL, oldQ })
	listPortsFunc = func() []PortInfo { return []PortInfo{{"COM6", "USB-SERIAL CH340 (COM6)"}} }

	type radio struct {
		m   *genModel
		mem []byte
	}
	var rmu sync.Mutex
	var cur *radio
	var sent atomic.Int64
	old := openPortFunc
	t.Cleanup(func() { openPortFunc = old })
	openPortFunc = func(name string, baud int) (Port, error) {
		x, y := newPipe()
		x.baud = baud
		rmu.Lock()
		r := cur
		rmu.Unlock()
		if r != nil {
			go genEmuFor(r.m, append([]byte{}, r.mem...)).serve(countPort{y, &sent})
		}
		return x, nil
	}
	// modello con la scheda aperta e un'altra radio: dello stesso gruppo di
	// «Identifica» (stesso saluto: il controllo breve la riconosce come diversa)
	// oppure, con "altro saluto", di un altro gruppo (al controllo breve non
	// risponde nemmeno)
	cases := [][3]string{
		{"ic-id880", "ic-id80"}, {"ic-t7h", "ic-w32a"}, {"ic-w32e", "ic-p7"},
		{"bj-9000", "rt-rt900", "altro saluto"}, {"rt-rt470", "tp-a36plus", "altro saluto"},
		{"qs-uvk5", "qs-uvk5-f4hwn"}, {"qs-uvk5-egz", "qs-uvk5"},
		{"kw-thf6", "kw-thg71"},
	}
	for _, cs := range cases {
		m, om, same := genByID[cs[0]], genByID[cs[1]], cs[2] == ""
		if m == nil || om == nil {
			t.Fatalf("modelli %v", cs)
		}
		if _, ok := m.Fam.(genQuicker); !ok {
			t.Fatalf("%s: senza controllo breve", m.ID)
		}
		mem, omem := identModelBase(t, m), identModelBase(t, om)
		if mem == nil || omem == nil {
			t.Skip("manca l'immagine di prova di CHIRP (CHIRP_DIR)")
		}
		t.Run(m.ID, func(t *testing.T) {
			identDiagReset()
			identSpontReset()
			identQuickReset()
			c := newAPIClient(t)
			a := c.app
			defer func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() }()
			wait := func(seq int, open []string) map[string]interface{} {
				t.Helper()
				t0 := time.Now()
				for {
					a.autoStep(time.Now())
					r := c.must("identify/auto", map[string]interface{}{"open": open})
					if int(r["seq"].(float64)) >= seq {
						return r["last"].(map[string]interface{})["result"].(map[string]interface{})
					}
					if time.Since(t0) > 90*time.Second {
						t.Fatalf("esito %d non arrivato (stato %v)\n%s", seq, r["status"], identDiag.text())
					}
					time.Sleep(20 * time.Millisecond)
				}
			}
			has := func(r map[string]interface{}, id string) bool {
				if r["kind"] == id {
					return true
				}
				os, _ := r["others"].([]interface{})
				for _, o := range os {
					if o.(map[string]interface{})["kind"] == id {
						return true
					}
				}
				return false
			}
			rmu.Lock()
			cur = &radio{m, mem}
			rmu.Unlock()
			// 1. scheda chiusa: memoria letta per intero
			sent.Store(0)
			c.must("identify/auto", map[string]interface{}{"enabled": true, "open": []string{}})
			r := wait(1, []string{})
			full := sent.Load()
			if !has(r, m.ID) || r["already"] == true || a.autoIdentSession(1) == nil {
				t.Fatalf("prima identificazione: %v\n%s", r, identDiag.text())
			}
			kind := r["kind"].(string)
			if kind != m.ID {
				// gruppo che non si distingue via cavo: si apre la scheda proposta
				if _, ok := genByID[kind].Fam.(genQuicker); !ok {
					t.Fatalf("%s proposta al posto di %s e senza controllo breve", kind, m.ID)
				}
			}
			c.must("identify/open", map[string]interface{}{"kind": kind, "seq": 1})
			// 2. scheda aperta: controllo breve
			sent.Store(0)
			c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{kind}})
			r = wait(2, []string{kind})
			quick := sent.Load()
			if r["kind"] != kind || r["already"] != true || a.autoIdentSession(2) != nil {
				t.Fatalf("scheda aperta: %v\n%s", r, identDiag.text())
			}
			if quick <= 0 || quick > 600 || quick*8 > full {
				t.Fatalf("scheda aperta: %d byte dalla radio (lettura intera %d)", quick, full)
			}
			if !strings.Contains(identDiag.text(), "controllo breve riuscito") {
				t.Errorf("diagnosi senza il controllo breve:\n%s", identDiag.text())
			}
			// 3. un'altra radio sullo stesso cavo, con la scheda della prima aperta:
			// il controllo breve non la conferma e la ricerca la riconosce per quello che è
			rmu.Lock()
			cur = &radio{om, omem}
			rmu.Unlock()
			c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{kind}})
			r = wait(3, []string{kind})
			if r["already"] == true || r["found"] != true || r["kind"] == kind || (same && !has(r, om.ID)) {
				t.Fatalf("altra radio: %v\n%s", r, identDiag.text())
			}
			if identQuickAllowed(kind) == same {
				t.Errorf("controllo breve per %s dopo l'altra radio: attivo = %v", kind, !same)
			}
			// 4. di nuovo la prima radio: dopo un controllo non riuscito il
			// controllo breve non si usa più (lettura normale); se l'altra radio
			// non aveva nemmeno risposto, si usa ancora
			rmu.Lock()
			cur = &radio{m, mem}
			rmu.Unlock()
			sent.Store(0)
			c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{kind}})
			r = wait(4, []string{kind})
			if same && (r["already"] == true || !has(r, m.ID) || sent.Load() < full/2) {
				t.Fatalf("dopo un controllo non riuscito: %v, %d byte (lettura intera %d)", r, sent.Load(), full)
			}
			if !same && (r["already"] != true || r["kind"] != kind || sent.Load() > 600) {
				t.Fatalf("di nuovo la radio della scheda: %v, %d byte", r, sent.Load())
			}
		})
	}
}

// famiglie senza controllo breve: restano con la lettura intera (vedi la regola
// in identquick.go). L'elenco è scritto qui perché aggiungerne una sia una scelta.
func TestIdentQuickFamilies(t *testing.T) {
	want := map[string]bool{
		"id31Fam": true, "id880Fam": true, "icw32Fam": true, "ict7Fam": true, "icq7Fam": true, "icp7Fam": true,
		"icv86Fam": true, "ict10Fam": true, "icv80Fam": true, "ict70Fam": true, "icx8xFam": true, "icx90Fam": true,
		"kwlFam": true, "uvk5Fam": true, "uvk5xFam": true, "uvkxFam": true, "rt900Fam": true, "jc8810Fam": true,
	}
	got := map[string]bool{}
	n := 0
	for _, m := range genModels {
		k := strings.TrimPrefix(fmt.Sprintf("%T", m.Fam), "main.")
		if _, ok := m.Fam.(genQuicker); ok {
			got[k] = true
			n++
			if _, no := m.Fam.(interface{ noIdent() }); no {
				t.Errorf("%s: controllo breve su una famiglia che non passa da «Identifica»", m.ID)
			}
		}
		if _, isIcom := icomOptOf(m); isIcom && !got[k] {
			t.Errorf("%s (%s): famiglia Icom senza controllo breve", m.ID, k)
		}
	}
	for k := range want {
		if !got[k] {
			t.Errorf("%s: controllo breve mancante", k)
		}
	}
	for k := range got {
		if !want[k] {
			t.Errorf("%s: controllo breve non previsto dall'elenco", k)
		}
	}
	if n < 60 {
		t.Errorf("solo %d modelli con il controllo breve", n)
	}
	t.Logf("%d modelli con il controllo breve", n)
}

// 4.40: Bajeton BJ-8300 con la scheda aperta: collegamento, primo blocco,
// chiusura (btCheck); senza la scheda aperta, lettura intera
func TestIdentQuickBJ8300(t *testing.T) {
	identFast(t)
	identFirstWait, identSilentMax = 80*time.Millisecond, 250*time.Millisecond
	identSpontListen, identSpontListen9600 = 100*time.Millisecond, 100*time.Millisecond
	oldS, oldF, oldL, oldQ := autoSettle, autoRetryFast, listPortsFunc, autoQuiet
	autoSettle, autoRetryFast, autoQuiet = 10*time.Millisecond, 20*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { autoSettle, autoRetryFast, listPortsFunc, autoQuiet = oldS, oldF, oldL, oldQ })
	identDiagReset()
	identQuickReset()
	listPortsFunc = func() []PortInfo { return []PortInfo{{"COM6", "USB-SERIAL CH340 (COM6)"}} }
	bt := newBtEmu(mustBtImage(t))
	identAttach(t, func(p Port) { _ = bt.serve(p) }, "")
	c := newAPIClient(t)
	a := c.app
	t.Cleanup(func() { st := a.store.Get(); st.AutoIdent = false; _ = a.store.SaveSettings(st); a.yieldAuto() })
	wait := func(seq int, open []string) map[string]interface{} {
		t.Helper()
		t0 := time.Now()
		for {
			a.autoStep(time.Now())
			r := c.must("identify/auto", map[string]interface{}{"open": open})
			if int(r["seq"].(float64)) >= seq {
				return r["last"].(map[string]interface{})["result"].(map[string]interface{})
			}
			if time.Since(t0) > 30*time.Second {
				t.Fatalf("esito %d non arrivato (stato %v)\n%s", seq, r["status"], identDiag.text())
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	session := func() []string {
		bt.mu.Lock()
		defer bt.mu.Unlock()
		out := append([]string{}, bt.log...)
		bt.log = nil
		return out
	}
	c.must("identify/auto", map[string]interface{}{"enabled": true, "open": []string{}})
	r := wait(1, []string{})
	full := session()
	if r["kind"] != kindDAT || r["already"] == true || len(full) < 100 || a.autoIdentSession(1) == nil {
		t.Fatalf("prima identificazione: %v, %d passi", r, len(full))
	}
	c.must("identify/open", map[string]interface{}{"kind": kindDAT, "seq": 1})
	c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{kindDAT}})
	r = wait(2, []string{kindDAT})
	quick := session()
	if r["kind"] != kindDAT || r["already"] != true || a.autoIdentSession(2) != nil {
		t.Fatalf("scheda aperta: %v\n%s", r, identDiag.text())
	}
	// collegamento, un blocco, chiusura: niente altro
	if len(quick) != 3 || quick[0] != "HS" || !strings.HasPrefix(quick[1], "R:0000:") || quick[2] != "E" {
		t.Fatalf("scheda aperta: passi %v", quick)
	}
	// scheda chiusa: di nuovo la lettura intera
	c.must("identify/auto", map[string]interface{}{"rescan": true, "open": []string{}})
	r = wait(3, []string{})
	if again := session(); r["already"] == true || len(again) != len(full) {
		t.Fatalf("scheda chiusa: %v, %d passi (lettura intera %d)", r, len(again), len(full))
	}
}
