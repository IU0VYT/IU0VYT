// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Diagnosi dell'identificazione (4.36).
//
// Le radio simulate delle prove rispondono all'istante e sempre nello stesso
// modo; una radio vera no (tempi, radio spenta ma collegata, cavo che rimanda
// indietro i dati). Quando «Identifica» si comporta male su una radio vera, per
// capire che cosa è successo servono i byte e i tempi di quella prova. Il
// programma li tiene in memoria (solo le ultime righe, niente su disco) e la
// scheda «Identifica» li salva in un file di testo quando l'utente lo chiede.
//
// Che cosa si registra: per ogni tentativo il nome del modello provato, la
// velocità, i primi byte inviati e ricevuti (in esadecimale, pochi per
// tentativo), le attese andate a vuoto e l'esito, con l'ora al millisecondo.
// Che cosa NON si registra: la memoria della radio (solo i primi byte delle
// prime risposte), la password di programmazione della K67 Pro (dopo il saluto
// degli invii si scrive solo la lunghezza), nomi di file o di cartelle.
//
// I testi della diagnosi restano in italiano in tutte le lingue (li legge chi
// sviluppa il programma): per questo sono tutti in questo file, o dentro le
// chiamate a identDiag.add e diagNote, che la raccolta dei testi da tradurre
// salta (i18n_test.go, extractGo).

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// identDiagMax: righe tenute in memoria (le più vecchie si perdono)
const identDiagMax = 8000

type identDiagBuf struct {
	mu    sync.Mutex
	lines []string
}

var identDiag identDiagBuf

func (d *identDiagBuf) add(port, f string, a ...interface{}) {
	line := time.Now().Format("15:04:05.000") + "  " + port + "  " + fmt.Sprintf(f, a...)
	d.mu.Lock()
	d.lines = append(d.lines, line)
	if len(d.lines) > identDiagMax {
		d.lines = append([]string{}, d.lines[len(d.lines)-identDiagMax*3/4:]...)
	}
	d.mu.Unlock()
}

// diagLine: riga con l'ora di adesso (per le righe tenute da parte, vedi
// probeState.dbuf)
func diagLine(port, f string, a ...interface{}) string {
	return time.Now().Format("15:04:05.000") + "  " + port + "  " + fmt.Sprintf(f, a...)
}

func (d *identDiagBuf) addLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	d.mu.Lock()
	d.lines = append(d.lines, lines...)
	if len(d.lines) > identDiagMax {
		d.lines = append([]string{}, d.lines[len(d.lines)-identDiagMax*3/4:]...)
	}
	d.mu.Unlock()
}

func (d *identDiagBuf) text() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return strings.Join(d.lines, "\r\n")
}

func (d *identDiagBuf) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.lines)
}

// righe per tentativo: oltre questi numeri si conta soltanto
const (
	identDiagTx   = 14 // invii
	identDiagRx   = 14 // letture riuscite
	identDiagTo   = 6  // attese senza risposta
	identDiagHex  = 24 // byte mostrati per riga
	identDiagTail = 4  // con più byte: anche gli ultimi
)

func diagHex(b []byte) string {
	if len(b) <= identDiagHex+identDiagTail {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b[:identDiagHex]) + "…" + hex.EncodeToString(b[len(b)-identDiagTail:])
}

// diagTx / diagRx / diagWait / diagNote: invii, letture, attese a vuoto e note
// di un tentativo (chiamate con s.mu NON preso). Le righe restano da parte
// (s.dbuf) fino alla fine del tentativo: si scrivono solo se dalla porta è
// arrivato almeno un byte. Di un tentativo rimasto senza nessuna risposta
// resta la riga finale: così in memoria ci stanno molti più giri di ricerca.
func (s *probeState) diagTx(b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dTx++
	show := true
	if s.secret {
		// K67 Pro: dopo il saluto (16 byte ripetuti, poi 8 byte con il seme)
		// viene la password: da lì in poi solo la lunghezza
		if len(b) != len(k67TInfo) {
			s.dPlain++
		}
		show = s.dPlain <= 1
	}
	switch {
	case s.dTx > identDiagTx:
	case show:
		s.dbuf = append(s.dbuf, diagLine(s.key, "    → %d byte: %s", len(b), diagHex(b)))
	default:
		s.dbuf = append(s.dbuf, diagLine(s.key, "    → %d byte", len(b)))
	}
}

func (s *probeState) diagRx(b []byte, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dRx++
	switch {
	case s.dRx > identDiagRx:
	case s.secret && len(b) > 64:
		// K67 Pro: i blocchi della memoria (che contiene anche la password) no
		s.dbuf = append(s.dbuf, diagLine(s.key, "    ← %d byte dopo %d ms", len(b), d.Milliseconds()))
	default:
		s.dbuf = append(s.dbuf, diagLine(s.key, "    ← %d byte dopo %d ms: %s", len(b), d.Milliseconds(), diagHex(b)))
	}
}

func (s *probeState) diagWait(want int, d time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dTo++
	if s.dTo <= identDiagTo {
		what := "niente"
		if err != nil && err != errTimeout {
			what = err.Error()
		}
		s.dbuf = append(s.dbuf, diagLine(s.key, "    ← %s (attesi %d byte, %d ms)", what, want, d.Milliseconds()))
	}
}

func (s *probeState) diagNote(f string, a ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.dbuf) < 4*(identDiagTx+identDiagRx) {
		s.dbuf = append(s.dbuf, diagLine(s.key, f, a...))
	}
}

// diagEnd: riga finale del tentativo
func (s *probeState) diagEnd(start time.Time, outcome string) {
	s.mu.Lock()
	tx, rx, to, key := s.dTx, s.dRx, s.dTo, s.key
	clean := append([]byte{}, s.clean...)
	gen, sure := s.genuine, s.sure
	buf := s.dbuf
	s.dbuf = nil
	s.mu.Unlock()
	if rx > 0 {
		identDiag.addLines(buf)
	}
	extra := ""
	if tx > identDiagTx || rx > identDiagRx || to > identDiagTo {
		extra = fmt.Sprintf(" [in tutto %d invii, %d letture, %d attese a vuoto]", tx, rx, to)
	}
	ans := "nessuna risposta"
	switch {
	case sure:
		ans = "risposta sicura"
	case gen:
		ans = "risposta non sicura"
	case len(clean) > 0:
		ans = "solo eco"
	case rx > 0:
		ans = "solo disturbi di linea"
	}
	identDiag.add(key, "  = %s; %s; %d ms%s", outcome, ans, time.Since(start).Milliseconds(), extra)
}

// identDiagReport: testo completo (intestazione + righe registrate)
func (a *App) identDiagReport() string {
	var sb strings.Builder
	w := func(f string, x ...interface{}) { fmt.Fprintf(&sb, f+"\r\n", x...) }
	w("IU0VYT Flash %s — diagnosi dell'identificazione", appVersion)
	w("Salvata il %s; sistema %s/%s", time.Now().Format("2006-01-02 15:04:05"), runtime.GOOS, runtime.GOARCH)
	w("Attese: prima risposta %d ms, silenzio %d ms, pausa %d ms, ascolto %d ms",
		identFirstWait.Milliseconds(), identSilentMax.Milliseconds(), identPause.Milliseconds(), identSpontListen.Milliseconds())
	st := a.store.Get()
	w("Identificazione automatica: %v; velocità ricordata della K67 Pro: %d baud", st.AutoIdent, st.K67Baud)
	w("Radio già riconosciute (la più recente per prima): %s", strings.Join(st.KnownRadios, ", "))
	var ports []string
	for p := range st.PortRadios {
		ports = append(ports, p)
	}
	sort.Slice(ports, func(i, j int) bool { return portLess(ports[i], ports[j]) })
	for _, p := range ports {
		w("Radio già riconosciute su %s: %s", p, strings.Join(st.PortRadios[p], ", "))
	}
	for _, p := range listPortsFunc() {
		ok, strong := cablePortKind(p)
		w("Porta %s: «%s» (cavo di programmazione: %v, adattatore noto: %v)", p.Name, p.Desc, ok, strong)
	}
	a.auto.mu.Lock()
	w("Stato dell'identificazione automatica: %s", a.auto.status)
	a.auto.mu.Unlock()
	w("Righe registrate: %d (al massimo %d, le più vecchie si perdono)", identDiag.count(), identDiagMax)
	w("Legenda: «→» byte inviati dal programma, «←» byte ricevuti; i byte sono in esadecimale.")
	w("")
	sb.WriteString(identDiag.text())
	sb.WriteString("\r\n")
	return sb.String()
}

// note dell'apertura della porta (vedi identSpontSample)
const (
	diagNoteKnown = " (radio già nota: senza ascolto)"
	diagNoteCut   = " (ascolto interrotto)"
)

func diagNoteSpont(data []byte) string {
	return " (la porta trasmette da sola: " + diagHex(data) + ")"
}

func diagNoteListen(d time.Duration, got []byte) string {
	if len(got) == 0 {
		return fmt.Sprintf(" (ascolto di %d ms: nessun byte)", d.Milliseconds())
	}
	return fmt.Sprintf(" (ascolto di %d ms: %d byte %s)", d.Milliseconds(), len(got), diagHex(got))
}

// diagOutcome: esito di un tentativo (vedi probe)
func diagOutcome(err error, silent bool) string {
	switch {
	case err == errCancelled:
		return "interrotto"
	case silent && err != nil:
		return "abbandonato (" + err.Error() + ")"
	case silent:
		return "abbandonato"
	case err != nil:
		return "errore: " + err.Error()
	}
	return "lettura riuscita"
}

// diagMode: tipo di ricerca (vedi runIdentify)
func diagMode(auto, knownOnly bool) string {
	switch {
	case auto && knownOnly:
		return "automatica, solo le radio già note"
	case auto:
		return "automatica"
	}
	return "manuale"
}
