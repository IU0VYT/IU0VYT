//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.41: ricevitore AOR AR8200 (drv_ar8200.go). CHIRP non ha file per questa
// radio: il suo driver parla con la radio simulata e deve leggere gli stessi
// canali con gli stessi comandi, e per le stesse modifiche scrivere le stesse
// righe. Poi le regole dei canali, la divisione dei banchi, le tre velocità,
// i due bit di stop e gli errori sul cavo.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func ar82Base(t *testing.T) (*genModel, []byte) {
	t.Helper()
	m := genByID["aor-ar8200"]
	f := ar82Fam{}
	img := ar82Blank(m)
	img[ar82SizesAt+1] = 30 // coppia B/b: 30 e 70 canali
	for n, hz := range []uint32{0, 600000, 1600000, 7600000, 5000000} {
		img[ar82OLAt+4*n], img[ar82OLAt+4*n+1], img[ar82OLAt+4*n+2], img[ar82OLAt+4*n+3] = byte(hz>>24), byte(hz>>16), byte(hz>>8), byte(hz)
	}
	for i, c := range map[int]*ar82Chan{
		0:   {freq: 145500000, step: 12500, auto: true, name: "S20"},                                                       // A00
		1:   {freq: 145750000, step: 12500, mode: 6, of: 1, dup: 2, name: "R6"},                                            // A01: FM stretta, -600 kHz
		49:  {freq: 118100000, step: 8330, mode: 2, att: true, of: 2, dup: 1, name: "TORRE"},                               // A49: AM, attenuatore
		50:  {freq: 14195000, step: 100, mode: 3, of: 4, dup: 1, name: "DX 20m"},                                           // a00: USB
		120: {freq: 98100000, step: 50000, auto: true, adjOn: true, adj: 25000, of: 3, dup: 2, name: "RADIO UNO"},          // B20
		131: {freq: 433500000, step: 25000, mode: 1, of: 2, dup: 2, sel: true, selIdx: 0, name: "LPD"},                     // b01
		999: {freq: 7030000, step: 50, mode: 5, pass: true, limit: true, protect: true, of: 1, dup: 1, name: "CW QRP 40m"}, // j49
	} {
		f.put(img, i, c)
	}
	return m, img
}

func TestAR8200Model(t *testing.T) {
	m, img := ar82Base(t)
	f := ar82Fam{}
	if m.Slots != 1000 || m.First != 0 || m.NameMax != 12 || m.Zones != 10 || !m.NoChirp || m.NoTxOff || !m.Narrow || m.Baud != 19200 ||
		len(m.DCS) != 0 || genMinSize(m) != 256+1000*64 || !f.Match(m, img) || f.Match(m, img[:len(img)-1]) {
		t.Errorf("dati del modello")
	}
	// posizione → banco e canale, con la divisione che dà la radio
	for _, c := range []struct {
		i    int
		bank byte
		ch   int
	}{{0, 'A', 0}, {49, 'A', 49}, {50, 'a', 0}, {99, 'a', 49}, {100, 'B', 0}, {129, 'B', 29}, {130, 'b', 0}, {199, 'b', 69},
		{200, 'C', 0}, {900, 'J', 0}, {950, 'j', 0}, {999, 'j', 49}} {
		if bank, ch := ar82Addr(img, c.i); bank != c.bank || ch != c.ch {
			t.Errorf("posizione %d: %c%02d, atteso %c%02d", c.i, bank, ch, c.bank, c.ch)
		}
	}
	// ogni posizione ha un canale suo
	seen := map[string]int{}
	for i := 0; i < m.Slots; i++ {
		bank, ch := ar82Addr(img, i)
		id := fmt.Sprintf("%c%02d", bank, ch)
		if o, dup := seen[id]; dup {
			t.Fatalf("posizioni %d e %d sullo stesso canale %s", o, i, id)
		}
		seen[id] = i
	}
	// la riga che scrive un canale, carattere per carattere (formato di CHIRP)
	for i, want := range map[int]string{
		0:   "MXA00 RF0145500000 MP0 ST012500 MD0 AT0 OF00 AU1 TMS20         ",
		1:   "MXA01 RF0145750000 MP0 ST012500 MD6 AT0 OF01- AU0 TMR6          ",
		49:  "MXA49 RF0118100000 MP0 ST008330 MD2 AT1 OF02+ AU0 TMTORRE       ",
		50:  "MXa00 RF0014195000 MP0 ST000100 MD3 AT0 OF04+ AU0 TMDX 20m      ",
		131: "MXb01 RF0433500000 MP0 ST025000 MD1 AT0 OF02- AU0 TMLPD         ",
		999: "MXj49 RF0007030000 MP1 ST000050 MD5 AT0 OF01+ AU0 TMCW QRP 40m  ",
	} {
		bank, ch := ar82Addr(img, i)
		if got := ar82Write(bank, ch, f.get(img, i)); got != want {
			t.Errorf("posizione %d:\n  %q\nvoluto\n  %q", i, got, want)
		}
	}
	// senza verso lo scostamento scritto è 00
	c := *f.get(img, 1)
	c.dup = 0
	if got := ar82Write('A', 1, &c); !strings.Contains(got, " OF00 ") {
		t.Errorf("canale senza verso: %q", got)
	}
	// i campi di un canale tornano uguali dall'immagine
	for i := 0; i < m.Slots; i++ {
		if c := f.get(img, i); c != nil {
			x := ar82Blank(m)
			f.put(x, 7, c)
			if back := f.get(x, 7); *back != *c {
				t.Errorf("posizione %d: %+v riletto %+v", i, *c, *back)
			}
		}
	}
}

func TestAR8200Channels(t *testing.T) {
	m, img := ar82Base(t)
	f := ar82Fam{}
	// lettura: sola ricezione, FM stretta, «pass»
	for i, want := range map[int]Chan{
		0:   {Rx: 145500000, Tx: 145500000, TxOff: true, Scan: true, Name: "S20"},
		1:   {Rx: 145750000, Tx: 145750000, TxOff: true, Narrow: true, Scan: true, Name: "R6"},
		999: {Rx: 7030000, Tx: 7030000, TxOff: true, Name: "CW QRP 40m"},
	} {
		if g := f.Get(m, img, i); g == nil || *g != want {
			t.Errorf("posizione %d: %+v", i, g)
		}
	}
	set := func(i int, c Chan) *ar82Chan {
		t.Helper()
		if err := f.Set(m, img, i, &c); err != nil {
			t.Fatalf("posizione %d, %+v: %v", i, c, err)
		}
		g := f.Get(m, img, i)
		want := c
		want.Tx, want.TxOff = c.Rx, true
		if g == nil || *g != want {
			t.Fatalf("posizione %d:\n  scritto %+v\n  riletto %+v", i, want, g)
		}
		return f.get(img, i)
	}
	// canale nuovo: modo automatico, passo da 5 kHz; con la FM stretta no
	if c := set(300, Chan{Rx: 145337500, Tx: 145337500, Scan: true, Name: "NUOVO"}); !c.auto || c.mode != 0 || c.step != 5000 || c.att || c.dup != 0 || c.limit || c.protect {
		t.Errorf("canale nuovo: %+v", *c)
	}
	if c := set(301, Chan{Rx: 446006250, Tx: 446006250, TxOff: true, Narrow: true, Name: "PMR 1"}); c.auto || c.mode != 6 || !c.pass {
		t.Errorf("canale nuovo in FM stretta: %+v", *c)
	}
	// nome e «pass» cambiati, frequenza no: tutto il resto resta
	before := *f.get(img, 999)
	c := set(999, Chan{Rx: 7030000, Tx: 7030000, Scan: true, Name: "QRP"})
	before.name, before.pass = "QRP", false
	if *c != before {
		t.Errorf("canale rinominato:\n  %+v\nvoluto\n  %+v", *c, before)
	}
	// frequenza cambiata: modo automatico; passo, attenuatore e scostamento restano
	before = *f.get(img, 49)
	c = set(49, Chan{Rx: 121500000, Tx: 121500000, Scan: true, Name: "TORRE"})
	before.freq, before.auto, before.mode = 121500000, true, 0
	if *c != before {
		t.Errorf("frequenza cambiata:\n  %+v\nvoluto\n  %+v", *c, before)
	}
	// FM stretta tolta: FM; messa su un canale in USB: FM stretta
	if c = set(1, Chan{Rx: 145750000, Tx: 145750000, Scan: true, Name: "R6"}); c.auto || c.mode != 1 {
		t.Errorf("FM stretta tolta: %+v", *c)
	}
	if c = set(50, Chan{Rx: 14195000, Tx: 14195000, Narrow: true, Scan: true, Name: "DX 20m"}); c.auto || c.mode != 6 {
		t.Errorf("FM stretta messa: %+v", *c)
	}
	// un canale automatico resta automatico se non cambia la frequenza
	if c = set(0, Chan{Rx: 145500000, Tx: 145500000, Name: "S20 bis"}); !c.auto || c.step != 12500 || !c.pass {
		t.Errorf("canale automatico: %+v", *c)
	}
	// con «sola ricezione» non spuntata ma senza scostamento il canale si salva lo stesso
	set(302, Chan{Rx: 145000000, Tx: 145000000, Scan: true, Name: "RX"})
	// rifiuti: l'immagine non cambia
	for _, c := range []Chan{
		{Rx: 145500000, Tx: 145500000, TxTone: "88.5"},
		{Rx: 145500000, Tx: 145500000, RxTone: "D023N"},
		{Rx: 145500000, Tx: 144900000},
		{Rx: 99000, Tx: 99000},
		{Rx: 2040000050, Tx: 2040000050},
		{Rx: 145500010, Tx: 145500010},
		{Rx: 118008333, Tx: 118008333},
	} {
		was := append([]byte{}, img...)
		if err := f.Set(m, img, 303, &c); err == nil {
			t.Errorf("%+v accettato", c)
		}
		if !bytes.Equal(was, img) {
			t.Fatalf("%+v rifiutato ma l'immagine è cambiata", c)
		}
	}
	// estremi della banda e nomi
	set(304, Chan{Rx: 100000, Tx: 100000, Scan: true, Name: "LW"})
	set(305, Chan{Rx: 2040000000, Tx: 2040000000, Scan: true, Name: "a/B-c d.e~12"})
	if err := f.Set(m, img, 306, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "Nome troppo lungo"}); err != nil {
		t.Fatal(err)
	}
	if g := f.Get(m, img, 306); g.Name != "Nome troppo" {
		t.Errorf("nome lungo riletto %q", g.Name)
	}
	// cancellazione
	if err := f.Set(m, img, 131, nil); err != nil || f.Get(m, img, 131) != nil || !bytes.Equal(f.rec(img, 131), make([]byte, 64)) {
		t.Errorf("cancellazione: %v", err)
	}
	// nelle conversioni da un ricetrasmettitore resta la ricezione, senza toni
	rep := &Report{}
	n := genAdapt(&Chan{Rx: 145750000, Tx: 145150000, TxTone: "88.5", RxTone: "D023N", Scan: true, Name: "R6"}, m, "canale 1", rep)
	if n == nil || n.Rx != 145750000 || n.Tx != n.Rx || !n.TxOff || n.TxTone != "" || n.RxTone != "" {
		t.Errorf("conversione: %+v (%v)", n, rep.Warnings)
	}
	if err := f.Set(m, img, 307, n); err != nil {
		t.Errorf("canale convertito rifiutato: %v", err)
	}
}

// comandi inviati (una riga per comando)
func ar82Cmds(tx []byte) []string {
	s := strings.Split(string(tx), "\r")
	return s[:len(s)-1]
}

var ar82ReSet = regexp.MustCompile(`^(MX|MQ[0-9]{2}$|SH[0-9]{6}|NL[01]$|PC[01]$|GD)`)

func TestAR8200Chirp(t *testing.T) {
	chirpDir(t)
	m, base := ar82Base(t)
	f := ar82Fam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	// lettura di tutte le 1000 posizioni
	var ops []map[string]interface{}
	for i := 0; i < m.Slots; i++ {
		ops = append(ops, map[string]interface{}{"op": "get", "n": i})
	}
	ce := f.emu(m, base).(*ar82Emu)
	res, chirpTx := liveSession(t, ce, "AR8200Radio", ops)
	modes := []string{"WFM", "FM", "AM", "USB", "LSB", "CW", "NFM", "AM", "AM"}
	full := 0
	for i := 0; i < m.Slots; i++ {
		var ref *struct {
			chirpChan
			Step float64 `json:"tuning_step"`
		}
		if err := json.Unmarshal(res[i], &ref); err != nil || strings.Contains(string(res[i]), `"error"`) {
			t.Fatalf("posizione %d: CHIRP risponde %s", i, res[i])
		}
		c := f.get(base, i)
		if (c == nil) != (ref == nil) {
			t.Fatalf("posizione %d: qui %v, CHIRP %s", i, c, res[i])
		}
		if c == nil {
			continue
		}
		full++
		mode := modes[c.mode]
		if c.auto {
			mode = "Auto"
		}
		tx := c.freq
		hz := int64(base[ar82OLAt+4*c.of])<<24 | int64(base[ar82OLAt+4*c.of+1])<<16 | int64(base[ar82OLAt+4*c.of+2])<<8 | int64(base[ar82OLAt+4*c.of+3])
		switch c.dup {
		case 1:
			tx += hz
		case 2:
			tx -= hz
		}
		if c.freq != ref.Rx || tx != ref.Tx || c.name != ref.Name || c.pass == ref.Scan || mode != ref.Mode || int(ref.Step*1000+0.5) != c.step {
			t.Errorf("posizione %d:\n  qui   %+v (%s, TX %d)\n  CHIRP %s", i, *c, mode, tx, res[i])
		}
	}
	if full != 7 {
		t.Fatalf("%d canali confrontati in lettura", full)
	}
	// gli stessi comandi di CHIRP, nello stesso ordine. Tolti: le righe vuote
	// (CHIRP ne manda una dopo ogni risposta che non capisce), le ripetizioni
	// (CHIRP ripete cinque volte GR per la riga che chiude l'elenco) e l'EX
	// finale, che CHIRP non manda
	ge := f.emu(m, base).(*ar82Emu)
	var got []byte
	err, gtx := alSession(t, ge, nil, func() error {
		var err error
		got, _, err = f.Download(ctx, m, "EMU", nop)
		return err
	})
	if err != nil || !bytes.Equal(got, base) || !bytes.Equal(got, ge.image()) {
		t.Fatalf("lettura: %v (immagine uguale: %v)", err, bytes.Equal(got, base))
	}
	norm := func(tx []byte) []string {
		var out []string
		for _, c := range ar82Cmds(tx) {
			if c == "" || c == "EX" || (len(out) > 0 && out[len(out)-1] == c) {
				continue
			}
			out = append(out, c)
		}
		return out
	}
	cn, gn := norm(chirpTx), norm(gtx)
	if strings.Join(cn, "|") != strings.Join(gn, "|") {
		for k := 0; k < len(cn) && k < len(gn); k++ {
			if cn[k] != gn[k] {
				t.Errorf("comando %d: CHIRP %q, qui %q", k, cn[k], gn[k])
				break
			}
		}
		t.Errorf("comandi di lettura diversi: CHIRP %d, qui %d", len(cn), len(gn))
	}
	// le stesse modifiche con CHIRP e da qui
	type ed struct {
		n    int
		c    *Chan
		mode string
		dup  int64 // scostamento che il canale ha già (per CHIRP: frequenza di trasmissione)
	}
	edits := []ed{
		{0, &Chan{Rx: 145500000, Tx: 145500000, Name: "S20 pass"}, "Auto", 0},              // nome e «pass»
		{1, &Chan{Rx: 145762500, Tx: 145762500, Scan: true, Name: "R6a"}, "Auto", -600000}, // frequenza: modo automatico
		{49, &Chan{Rx: 118100000, Tx: 118100000, Scan: true, Name: "TWR"}, "AM", 1600000},  // solo il nome: AM resta
		{50, &Chan{Rx: 14195000, Tx: 14195000, Narrow: true, Scan: true, Name: "DX 20m"}, "NFM", 5000000},
		{120, nil, "", 0},
		{300, &Chan{Rx: 433500000, Tx: 433500000, Scan: true, Name: "NUOVO"}, "Auto", 0},
		{301, &Chan{Rx: 446006250, Tx: 446006250, Narrow: true, Name: "PMR 1"}, "NFM", 0},
		{999, &Chan{Rx: 7030000, Tx: 7030000, Name: "CW 40m"}, "CW", 600000},
	}
	ops = nil
	mod := append([]byte{}, base...)
	for _, e := range edits {
		if e.c == nil {
			ops = append(ops, map[string]interface{}{"op": "get", "n": e.n}, map[string]interface{}{"op": "del", "n": e.n})
		} else {
			ops = append(ops, map[string]interface{}{"op": "set", "n": e.n, "chan": map[string]interface{}{
				"rx": e.c.Rx, "tx": e.c.Rx + e.dup, "txOff": false, "txTone": "", "rxTone": "", "power": -1,
				"mode": e.mode, "scan": e.c.Scan, "name": e.c.Name}})
		}
		if err := f.Set(m, mod, e.n, e.c); err != nil {
			t.Fatalf("posizione %d: %v", e.n, err)
		}
	}
	ce = f.emu(m, base).(*ar82Emu)
	res, chirpTx = liveSession(t, ce, "AR8200Radio", ops)
	// CHIRP non riesce a scrivere un canale in AM: ar8200.py confronta il nome
	// del modo con l'oggetto dell'impostazione «AM Mode» invece che con il suo
	// valore, e si ferma con un IndexError. Qui il canale si scrive (riga
	// controllata più sotto); per il confronto con CHIRP lo si toglie
	const amLine = "MXA49 RF0118100000 MP0 ST008330 MD2 AT1 OF02+ AU0 TMTWR         "
	for k, r := range res {
		s := string(r)
		if k == 2 {
			if !strings.Contains(s, "IndexError") {
				t.Errorf("CHIRP ora scrive i canali in AM (%s): il confronto va rifatto anche per quelli", s)
			}
			continue
		}
		if s != `"ok"` && !strings.HasPrefix(s, "{\"rx\"") {
			t.Errorf("CHIRP, operazione %d: %s", k, r)
		}
	}
	ge = f.emu(m, base).(*ar82Emu)
	err, gtx = alSession(t, ge, nil, func() error {
		_, err := f.Upload(ctx, m, "EMU", mod, nop)
		return err
	})
	if err != nil {
		t.Fatalf("scrittura: %v", err)
	}
	sets := func(tx []byte) []string {
		var out []string
		for _, c := range ar82Cmds(tx) {
			if ar82ReSet.MatchString(c) {
				out = append(out, c)
			}
		}
		sort.Strings(out)
		return out
	}
	cs, gs := sets(chirpTx), sets(gtx)
	// il canale in AM: la riga scritta da qui, poi tolto con i suoi tre comandi
	for _, drop := range []string{amLine, "SH000000", "NL0", "PC0"} {
		k := sort.SearchStrings(gs, drop)
		if k >= len(gs) || gs[k] != drop {
			t.Fatalf("comando %q non inviato per il canale in AM", drop)
		}
		gs = append(gs[:k], gs[k+1:]...)
	}
	if strings.Join(cs, "|") != strings.Join(gs, "|") {
		t.Errorf("comandi di scrittura diversi:\n  CHIRP %q\n  qui   %q", cs, gs)
	}
	ci, gi := ce.image(), ge.image()
	copy(f.rec(ci, 49), f.rec(gi, 49)) // il canale in AM che CHIRP non ha scritto
	if !bytes.Equal(ci, gi) || !bytes.Equal(gi, mod) {
		t.Errorf("la radio alla fine non ha gli stessi canali")
	}
	if ce.writes != 6 || ge.writes != 7 || ce.deletes != 1 || ge.deletes != 1 {
		t.Errorf("canali scritti %d/%d, cancellati %d/%d", ce.writes, ge.writes, ce.deletes, ge.deletes)
	}
	t.Logf("%d canali letti come CHIRP con gli stessi comandi; %d modifiche: gli stessi %d comandi di scrittura (più il canale in AM, che CHIRP non riesce a scrivere)", full, len(edits), len(gs))
}

func TestAR8200Cable(t *testing.T) {
	oldT, oldW, oldM := ar82Timeout, ar82Wait, ar82WaitMax
	defer func() { ar82Timeout, ar82Wait, ar82WaitMax = oldT, oldW, oldM }()
	ar82Timeout, ar82Wait, ar82WaitMax = 150*time.Millisecond, 5*time.Millisecond, 20*time.Millisecond
	m, base := ar82Base(t)
	f := ar82Fam{}
	ctx := context.Background()
	nop := func(int, int, string) {}
	mk := func() *ar82Emu { return f.emu(m, base).(*ar82Emu) }
	var pc *pipeEnd
	read := func(e *ar82Emu) ([]byte, error, []byte) {
		var img []byte
		e.pcBaud = func() int { return pc.baud }
		err, tx := alSession(t, e, func(p *pipeEnd) Port { pc = p; return p }, func() error {
			var err error
			img, _, err = f.Download(ctx, m, "EMU", nop)
			return err
		})
		return img, err, tx
	}
	write := func(e *ar82Emu, img []byte) (error, []byte) {
		e.pcBaud = func() int { return pc.baud }
		return alSession(t, e, func(p *pipeEnd) Port { pc = p; return p }, func() error {
			_, err := f.Upload(ctx, m, "EMU", img, nop)
			return err
		})
	}
	// lettura: i comandi, nell'ordine
	e := mk()
	img, err, tx := read(e)
	if err != nil || !bytes.Equal(img, base) || !bytes.Equal(img, e.image()) {
		t.Fatalf("lettura: %v", err)
	}
	cmds := ar82Cmds(tx)
	start := []string{"", "", "VR", "VF", "MC1", "MWA", "OL", "OL", "OL", "OL", "OL", "GR", "MAA", "MA", "MA", "MA", "MA", "MRA00", "OF", "SH", "NL", "PC", "MRA01", "OF"}
	if len(cmds) < len(start) || strings.Join(cmds[:len(start)], "|") != strings.Join(start, "|") {
		t.Errorf("inizio della lettura: %q", cmds[:min(len(cmds), len(start))])
	}
	count := map[string]int{}
	for _, c := range cmds {
		count[c[:min(len(c), 2)]]++
	}
	// 10 divisioni, 5 pagine di scostamenti, 100 pagine di canali, 7 canali presenti, un EX alla fine
	if count["MW"] != 10 || count["OL"] != 5 || count["GR"] != 1 || count["MA"] != 100 || count["MR"] != 7 || count["OF"] != 7 || count["SH"] != 7 ||
		count["NL"] != 7 || count["PC"] != 7 || cmds[len(cmds)-1] != "EX" || e.ex != 1 || e.writes+e.deletes != 0 {
		t.Errorf("comandi della lettura: %v", count)
	}
	if !pc.stop2 {
		t.Errorf("porta senza i due bit di stop")
	}
	// la radio a 9600 e a 4800 baud: si trova lo stesso (dopo i tentativi alle altre velocità)
	for _, baud := range []int{9600, 4800} {
		e = mk()
		e.baud = baud
		if img, err, _ = read(e); err != nil || !bytes.Equal(img, base) || pc.baud != baud {
			t.Errorf("radio a %d baud: %v (porta a %d)", baud, err, pc.baud)
		}
	}
	// radio spenta: i tentativi di CHIRP alle tre velocità (per ognuna 5 righe
	// vuote e 5 VR, più la riga vuota di ogni ripresa), poi l'errore; niente EX
	e = mk()
	e.baud = 1200
	img, err, tx = read(e)
	cmds = ar82Cmds(tx)
	count = map[string]int{}
	for _, c := range cmds {
		count[c]++
	}
	if err == nil || img != nil || !strings.Contains(err.Error(), "non risponde") || count["VR"] != 15 || count[""] != 1+3*(5+5+5) || len(count) != 2 {
		t.Errorf("radio spenta: %v, comandi %v", err, count)
	}
	// risposte nelle altre forme: «OF00 » senza verso, elenco vuoto della
	// scansione selettiva senza numero, XOFF e XON alla scrittura
	e = mk()
	e.ofSpace, e.grBare = true, true
	e.gr = nil
	nosel := append([]byte{}, base...)
	f.rec(nosel, 131)[28] = 0
	if img, err, _ = read(e); err != nil || !bytes.Equal(img, nosel) {
		t.Errorf("lettura con le altre forme di risposta: %v", err)
	}
	// una pagina rovinata: l'elenco del banco si richiede da capo
	e = mk()
	e.garble = 13
	img, err, tx = read(e)
	if err != nil || !bytes.Equal(img, base) || strings.Count(string(tx), "MAB\r") != 2 {
		t.Errorf("pagina rovinata: %v, %d richieste del banco", err, strings.Count(string(tx), "MAB\r"))
	}
	// scrittura: solo i canali cambiati
	mod := append([]byte{}, base...)
	for i, c := range map[int]*Chan{
		0:   {Rx: 145500000, Tx: 145500000, Name: "S20 pass"},
		131: nil,
		500: {Rx: 433500000, Tx: 433500000, Scan: true, Name: "NUOVO"},
		999: {Rx: 7030000, Tx: 7030000, Scan: true, Name: "CW 40m"},
	} {
		if err := f.Set(m, mod, i, c); err != nil {
			t.Fatal(err)
		}
	}
	for _, xon := range []bool{false, true} {
		e = mk()
		e.xon = xon
		err, tx = write(e, mod)
		if err != nil {
			t.Fatalf("scrittura (XON/XOFF %v): %v", xon, err)
		}
		// la radio alla fine: uguale all'immagine scritta (la scansione
		// selettiva del canale cancellato sparisce con lui)
		if !bytes.Equal(e.image(), mod) || e.writes != 3 || e.deletes != 1 || e.ex != 1 {
			t.Errorf("scrittura (XON/XOFF %v): %d canali scritti, %d cancellati", xon, e.writes, e.deletes)
		}
		var sets []string
		for _, c := range ar82Cmds(tx) {
			if ar82ReSet.MatchString(c) {
				sets = append(sets, c)
			}
		}
		want := []string{
			"MXA00 RF0145500000 MP1 ST012500 MD0 AT0 OF00 AU1 TMS20 pass    ", "NL0", "PC0",
			"MQ01",
			"MXF00 RF0433500000 MP0 ST005000 MD0 AT0 OF00 AU1 TMNUOVO       ", "NL0", "PC0",
			"MXj49 RF0007030000 MP0 ST000050 MD5 AT0 OF01+ AU0 TMCW 40m      ", "SH000000", "NL1", "PC1",
		}
		if strings.Join(sets, "|") != strings.Join(want, "|") {
			t.Errorf("scrittura: comandi\n  %q\nvoluti\n  %q", sets, want)
		}
	}
	// niente da cambiare: niente scritto
	e = mk()
	if err, _ = write(e, base); err != nil || e.writes+e.deletes != 0 {
		t.Errorf("scrittura senza modifiche: %v", err)
	}
	// la correzione del passo si commuta solo se serve
	adj := append([]byte{}, base...)
	c := f.get(adj, 120)
	c.auto, c.mode, c.name = false, 0, "RADIO 1"
	f.put(adj, 120, c)
	e = mk()
	if err, tx = write(e, adj); err != nil || !strings.Contains(string(tx), "\rSH025000\r") || strings.Contains(string(tx), "SH025000+") || !e.ch["B20"].adjOn {
		t.Errorf("correzione del passo già accesa: %v", err)
	}
	c.adj = 0
	f.put(adj, 120, c)
	e = mk()
	if err, tx = write(e, adj); err != nil || !strings.Contains(string(tx), "\rSH000000+\r") || e.ch["B20"].adjOn {
		t.Errorf("correzione del passo da spegnere: %v", err)
	}
	// la radio rifiuta un canale: cinque tentativi, poi l'errore con il canale
	e = mk()
	e.refuse = "F00"
	err, tx = write(e, mod)
	if err == nil || !strings.Contains(err.Error(), "F00") || !strings.Contains(err.Error(), "posizione 500") || strings.Count(string(tx), "MXF00 ") != 5 || e.writes != 1 || e.ex != 1 {
		t.Errorf("canale rifiutato: %v (%d tentativi)", err, strings.Count(string(tx), "MXF00 "))
	}
	// la divisione dei banchi o la tabella degli scostamenti cambiate dopo la lettura: niente scritto
	e = mk()
	e.sizes[4] = 70
	if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "divisione dei banchi") || e.writes+e.deletes != 0 {
		t.Errorf("divisione dei banchi cambiata: %v", err)
	}
	e = mk()
	e.ol[1] = 700000
	if err, _ = write(e, mod); err == nil || !strings.Contains(err.Error(), "scostamenti") || e.writes+e.deletes != 0 {
		t.Errorf("tabella degli scostamenti cambiata: %v", err)
	}
	// immagine non valida: niente inviato
	bad := append([]byte{}, mod...)
	f.rec(bad, 500)[9] = 9
	for _, x := range [][]byte{mod[:len(mod)-1], bad} {
		if err, tx = write(mk(), x); err == nil || len(tx) != 0 {
			t.Errorf("immagine non valida: %v, %d byte inviati", err, len(tx))
		}
	}
	// una porta senza i due bit di stop non parla con la radio
	e = mk()
	old := openPortFunc
	a, b := newPipe()
	go e.serve(b)
	openPortFunc = func(name string, baud int) (Port, error) { a.baud = baud; return noStop2Port{a}, nil }
	_, _, err = f.Download(ctx, m, "EMU", nop)
	openPortFunc = old
	a.Close()
	if err == nil {
		t.Errorf("lettura riuscita senza i due bit di stop")
	}
}

// noStop2Port: porta che non sa impostare i due bit di stop
type noStop2Port struct{ p *pipeEnd }

func (n noStop2Port) Write(b []byte) error { return n.p.Write(b) }
func (n noStop2Port) ReadFull(ctx context.Context, b []byte, to time.Duration) error {
	return n.p.ReadFull(ctx, b, to)
}
func (n noStop2Port) Purge() error        { return n.p.Purge() }
func (n noStop2Port) SetBaud(b int) error { return n.p.SetBaud(b) }
func (n noStop2Port) Close() error        { return n.p.Close() }
