//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"context"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// lentaPort: lato radio di una porta simulata. La prima risposta di ogni
// collegamento parte dopo un ritardo, come in una radio vera che deve entrare in
// programmazione; conta i collegamenti in cui la radio ha risposto.
type lentaPort struct {
	Port
	lat      time.Duration
	answered *int32
	done     bool
}

func (p *lentaPort) Write(b []byte) error {
	if !p.done {
		p.done = true
		time.Sleep(p.lat)
		atomic.AddInt32(p.answered, 1)
	}
	return p.Port.Write(b)
}

// 4.35: «Identifica» con una radio che risponde al saluto dopo 150-250 ms (tempi
// veri, tutto il piano di ricerca). La radio che ha ricevuto il suo saluto è
// entrata in programmazione: il programma deve aspettare la risposta e
// completare l'identificazione in quello stesso collegamento. Se abbandona il
// tentativo prima della risposta (come faceva il giro veloce delle versioni
// 4.31-4.34 dopo 100 ms) la radio resta bloccata in programmazione e i
// tentativi seguenti le mandano i saluti di altre radio.
func TestIdentifySlowRadio(t *testing.T) {
	type cas struct {
		name, kind string
		serve      func(p Port)
		lat        time.Duration
	}
	k67 := &k67Emu{mem: xlcImageForRadio(mustRead(t, sampleXLC)), model: "BF-K6"}
	btImg, err := datToEEPROM(mustRead(t, sampleDAT))
	if err != nil {
		t.Fatal(err)
	}
	bt := newBtEmu(btImg)
	cases := []cas{
		// la K67 Pro ripete la domanda ogni 200 ms, come il programma del
		// costruttore: la sua risposta deve arrivare prima
		{"Baofeng K67 Pro", kindXLC, func(p Port) { _ = k67.serve(p) }, 150 * time.Millisecond},
		{"Bajeton BJ-8300", kindDAT, func(p Port) { _ = bt.serve(p) }, 250 * time.Millisecond},
	}
	for _, id := range []string{"bj-9000", "bf-uv5r", "bf-uv17pro", "qs-uvk5", "bf-888s", "td-tdh8"} {
		m := genByID[id]
		mem := identModelBase(t, m)
		if mem == nil {
			t.Logf("%s: nessuna immagine di prova", m.Name)
			continue
		}
		emu := genEmuFor(m, mem)
		cases = append(cases, cas{m.Name, id, func(p Port) { emu.serve(p) }, 250 * time.Millisecond})
	}
	run := func(t *testing.T, cs cas, lat time.Duration) (found bool, kind string, answered int32, d time.Duration) {
		c := newAPIClient(t)
		var n int32
		identAttach(t, func(p Port) { cs.serve(&lentaPort{Port: p, lat: lat, answered: &n}) }, "")
		t0 := time.Now()
		c.must("identify/start", map[string]string{"port": "COM6"})
		st := identWait(c)
		if st["state"] != "done" {
			t.Fatalf("ricerca: %v %v", st["state"], st["error"])
		}
		res := st["result"].(map[string]interface{})
		k, _ := res["kind"].(string)
		return res["found"] == true, k, atomic.LoadInt32(&n), time.Since(t0)
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			f0, k0, a0, d0 := run(t, cs, 0)
			if !f0 {
				t.Fatalf("radio pronta: non identificata")
			}
			f1, k1, a1, d1 := run(t, cs, cs.lat)
			t.Logf("radio pronta: %s, %d collegamenti con risposta, %v; radio lenta (%v): %s, %d collegamenti, %v",
				k0, a0, d0.Round(100*time.Millisecond), cs.lat, k1, a1, d1.Round(100*time.Millisecond))
			if !f1 || k1 != k0 {
				t.Fatalf("radio lenta: non identificata come la radio pronta (%v %q, attesa %q)", f1, k1, k0)
			}
			if a1 != a0 {
				t.Errorf("radio lenta: ha risposto in %d collegamenti, la radio pronta in %d: un tentativo è stato abbandonato dopo che la radio era entrata in programmazione", a1, a0)
			}
			if d1 > d0+15*time.Second {
				t.Errorf("radio lenta: ricerca di %v contro %v", d1, d0)
			}
		})
	}
}

// identGroupLines: i gruppi di «Identifica» (modelli provati con lo stesso
// tentativo), una riga per gruppo: primo modello (quello che fa il tentativo) e
// poi tutti i modelli in ordine alfabetico
func identGroupLines(t *testing.T) []string {
	a := newApp(mustStore(t))
	identPortMu.Lock()
	plan := a.identPlan(&identReq{Port: "COM7"})
	identPortMu.Unlock()
	var out []string
	for _, p := range plan {
		if len(p.kinds) == 0 {
			continue
		}
		var ids []string
		for _, k := range p.kinds {
			ids = append(ids, k.ID)
		}
		sort.Strings(ids)
		out = append(out, p.kinds[0].ID+": "+strings.Join(ids, " "))
	}
	sort.Strings(out)
	return out
}

// 4.35: i «saluti» dei driver (la chiave con cui i modelli si raggruppano nella
// ricerca) non devono dipendere dal tempo: calcolati due volte, la seconda con
// il processore occupato, devono essere identici. Prima della 4.35 venti modelli
// cambiavano gruppo da un avvio all'altro.
func TestIdentTranscriptsStable(t *testing.T) {
	identPortMu.Lock()
	defer identPortMu.Unlock()
	reset := func() {
		identTxOnce = sync.Once{}
		identTxKeys = nil
	}
	defer reset()
	reset()
	t0 := time.Now()
	k1 := map[string]string{}
	for id, k := range identTranscripts() {
		k1[id] = k
	}
	d1, late1 := time.Since(t0), identTxLate
	stop := make(chan struct{})
	for w := 0; w < 16; w++ {
		go func() {
			for x := 0; ; x++ {
				select {
				case <-stop:
					return
				default:
				}
			}
		}()
	}
	reset()
	k2 := identTranscripts()
	close(stop)
	if late1 != 0 || identTxLate != 0 {
		t.Errorf("trascrizioni fermate dal tempo invece che dal conteggio: %d e %d", late1, identTxLate)
	}
	if d1 > 3*time.Second {
		t.Errorf("preparazione dei protocolli troppo lunga: %v", d1)
	}
	diff := 0
	for id, k := range k1 {
		if k2[id] == k {
			continue
		}
		// saluto con una parte casuale (famiglia Retevis C6 / C66 / Echo): la
		// chiave cambia per forza; ogni modello fa gruppo a sé
		if _, ok := genByID[id].Fam.(wltFam); ok {
			continue
		}
		diff++
		t.Errorf("%s: saluto diverso da un calcolo all'altro\n  %q\n  %q", id, k, k2[id])
	}
	t.Logf("%d saluti in %v, %d diversi", len(k1), d1.Round(time.Millisecond), diff)
}

// 4.35: i gruppi di «Identifica» sono fissati in testdata/ident_gruppi.txt. Chi
// aggiunge una radio deve guardare in che gruppo finisce (e chi fa il tentativo
// per il gruppo) e poi rigenerare il file con IU0VYT_SCRIVI_GRUPPI=1: così un
// cambiamento dei gruppi non passa inosservato.
func TestIdentGroupsGolden(t *testing.T) {
	got := identGroupLines(t)
	const file = "testdata/ident_gruppi.txt"
	text := strings.Join(got, "\n") + "\n"
	if os.Getenv("IU0VYT_SCRIVI_GRUPPI") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("scritto %s (%d gruppi)", file, len(got))
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	wl := strings.Split(strings.TrimRight(strings.ReplaceAll(string(want), "\r\n", "\n"), "\n"), "\n")
	have := map[string]bool{}
	for _, l := range got {
		have[l] = true
	}
	exp := map[string]bool{}
	for _, l := range wl {
		exp[l] = true
		if !have[l] {
			t.Errorf("gruppo atteso e non trovato: %s", l)
		}
	}
	for _, l := range got {
		if !exp[l] {
			t.Errorf("gruppo nuovo o cambiato: %s", l)
		}
	}
	// la famiglia Retevis RT22 / Baofeng T20 («PROGRGS») è un gruppo solo
	n := 0
	for _, l := range got {
		if strings.Contains(l, " rt-rt22 ") {
			n++
			for _, id := range []string{"bf-t20", "bf-w31e", "lt-lt316", "rt-rt22frs", "rt-rt622", "tid-tdm8", "wln-kdc1", "zt-ztx6"} {
				if !strings.Contains(l+" ", " "+id+" ") {
					t.Errorf("%s fuori dal gruppo della Retevis RT22: %s", id, l)
				}
			}
		}
	}
	if n != 1 {
		t.Errorf("la Retevis RT22 è in %d gruppi", n)
	}
}

// eco435Port: cavo che rimanda indietro tutto quello che riceve (nessuna radio)
type eco435Port struct {
	mu  sync.Mutex
	buf []byte
}

func (p *eco435Port) Write(b []byte) error {
	p.mu.Lock()
	p.buf = append(p.buf, b...)
	p.mu.Unlock()
	return nil
}
func (p *eco435Port) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.buf) < len(b) {
		return errTimeout
	}
	copy(b, p.buf)
	p.buf = p.buf[len(b):]
	return nil
}
func (p *eco435Port) Purge() error {
	p.mu.Lock()
	p.buf = nil
	p.mu.Unlock()
	return nil
}
func (p *eco435Port) SetBaud(int) error { return nil }
func (p *eco435Port) Close() error      { return nil }

// 4.35: un driver che manda pacchetti lunghi pieni di zeri (Radtel RT-5D: la
// richiesta di lettura ha la stessa forma della risposta) su un cavo con l'eco
// e senza radio. Dopo 4096 byte inviati il programma smetteva di ricordarli e
// l'eco dei pacchetti seguenti, con un byte perso, passava per la risposta
// sicura di una radio: riquadro «Identificazione Radio In Corso» senza radio e,
// alla fine, una radio «letta» dal solo eco. Con una radio vera la risposta
// resta sicura.
func TestIdentEchoLongFrames(t *testing.T) {
	identPortMu.Lock()
	defer identPortMu.Unlock()
	identSpontReset()
	frame := func(k byte) []byte {
		f := append([]byte{0xA5, 0x13, 0x00, k, 0x03, 0x20}, make([]byte, 0x320)...)
		return append(f, 0x23, 0x61+k)
	}
	run := func(port string, radio bool) *probeState {
		lp := &eco435Port{}
		old := openPortFunc
		defer func() { openPortFunc = old }()
		openPortFunc = func(string, int) (Port, error) { return lp, nil }
		st, _, _ := probe(context.Background(), port, func(ctx context.Context) error {
			p, err := openPort(port, 115200)
			if err != nil {
				return err
			}
			read := func(n int) error { return p.ReadFull(ctx, make([]byte, n), 50*time.Millisecond) }
			for k := byte(0); k < 8; k++ {
				f := frame(k)
				_ = p.Write(f)
				_ = read(1)
				_ = read(5)
				if k == 6 {
					// un disturbo in mezzo ai dati: la lettura resta indietro di un byte,
					// il driver scarta il resto e rimanda la richiesta
					lp.mu.Lock()
					lp.buf = append(append(append([]byte{}, lp.buf[:100]...), 0xF8), lp.buf[100:]...)
					lp.mu.Unlock()
					_ = read(0x322)
					_ = p.Purge()
					_ = p.Write(f)
					_ = read(1)
					_ = read(5)
				}
				_ = read(0x322)
			}
			if radio {
				// la radio risponde: conferma e identificativo
				_ = lp.Write([]byte{0x06, 0xAA, 0x42, 0x46, 0x42, 0x32, 0x39, 0x37, 0xDD})
				_ = read(1)
				_ = read(8)
			}
			return nil
		})
		return st
	}
	if st := run("COM31", false); st.radioSure() || !st.echoOnly() {
		t.Errorf("solo eco del cavo (pacchetti lunghi, un byte perso): presa per la risposta di una radio (% X)", st.clean)
	}
	if st := run("COM32", true); !st.radioSure() || st.echoOnly() {
		t.Errorf("radio vera dopo i pacchetti lunghi: risposta non riconosciuta come sicura")
	}
	// una radio vuota che manda solo conferme 0x06, come quelle del programma: i
	// byte arrivati sono tutti fra quelli inviati, ma non è un'eco
	st := &probeState{tx: []byte{0x53, 0x40, 0x06, 0x53, 0x40, 0x06, 0x53, 0x40, 0x06, 0x53, 0x40, 0x06, 0x53, 0x40, 0x06}, clean: []byte{0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06}}
	if st.echoOnly() {
		t.Errorf("sole conferme 0x06: scambiate per l'eco del cavo")
	}
}
