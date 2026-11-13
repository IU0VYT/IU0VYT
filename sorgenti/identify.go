// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Identificazione della radio collegata (3.8, scheda «Identifica»).
//
// Il programma prova, uno dopo l'altro, i protocolli di LETTURA di tutte le radio
// che conosce (K67 Pro, Bajeton BJ-8300, poi i driver delle altre radio) finché
// una risponde; la radio che risponde viene letta per intero e riconosciuta dal
// contenuto della memoria, come fanno CHIRP e dmrconfig con i file.
//
// Per non provare decine di volte la stessa cosa, i modelli che all'inizio del
// collegamento inviano esattamente gli stessi byte alla stessa velocità (lo
// stesso «saluto») formano un gruppo e se ne prova uno solo: i byte si ricavano
// facendo «parlare» ogni driver con una radio finta che non risponde mai
// (identTranscript). Se la radio risponde al saluto ma la lettura non riesce
// (ad esempio la radio dice di essere un altro modello), si provano gli altri
// modelli del gruppo.
//
// Per abbreviare l'attesa quando la radio non risponde, ogni attesa della
// PRIMA risposta è limitata (identFirstWait) e il tentativo si abbandona se la
// radio non ha risposto entro identSilentMax. I byte che tornano uguali a
// quelli inviati (eco del cavo) non contano come risposta.
//
// Si inviano solo comandi di lettura: nessun dato viene scritto nella radio.

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
)

var (
	// 4.3: attese più brevi (le radio rispondono al saluto in poche decine di
	// millisecondi): la ricerca completa senza radio passa da circa 135 a meno di
	// 60 secondi; una radio che ha risposto continua con le attese normali
	identFirstWait = 350 * time.Millisecond  // attesa massima di ogni lettura prima della prima risposta
	identSilentMax = 700 * time.Millisecond  // tentativo abbandonato se la radio tace
	identPause     = 1000 * time.Millisecond // pausa dopo un tentativo a cui la radio ha risposto
	// 4.31: giro veloce. Una radio accesa risponde al saluto in poche decine di
	// millisecondi: prima del giro con le attese piene se ne fa uno con attese
	// brevi, che da solo trova quasi tutte le radio in un quarto del tempo. Il giro
	// con le attese piene resta, per le radio lente.
	identFastWait   = 100 * time.Millisecond // giro veloce: attesa della prima risposta
	identFastSilent = 320 * time.Millisecond // giro veloce: tentativo abbandonato se la radio tace
	// giro veloce: dopo una domanda rimasta senza risposta il driver ne fa subito
	// un'altra oppure il tentativo si chiude (niente ripetizioni né pause)
	identFastIdle = 30 * time.Millisecond
	// giro veloce: una domanda (stessi byte, stessa velocità) a cui la radio non ha
	// risposto in questo giro, ripetuta dallo stesso driver o da un altro, non
	// aspetta di nuovo: la prima volta ha già avuto tutta l'attesa. Il giro con le
	// attese piene resta com'era prima della 4.31
	identReplayWaitFast = 10 * time.Millisecond
	identPortMu         sync.Mutex // solo per i test: prove in sequenza
)

// ---------------------------------------------------------------- radio finta muta

// silentPort: radio che non risponde mai; registra i byte inviati dal driver.
type silentPort struct {
	tx  *bytes.Buffer
	max int
	mu  *sync.Mutex
}

var errIdentEnough = errors.New("trascrizione completa")

func (p *silentPort) Write(b []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tx.Write(b)
	if p.tx.Len() >= p.max {
		return errIdentEnough
	}
	return nil
}
func (p *silentPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if ctx.Err() != nil {
		return errCancelled
	}
	return errTimeout
}
func (p *silentPort) Purge() error { return nil }
func (p *silentPort) SetBaud(baud int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.tx, "\x00<baud %d>\x00", baud)
	return nil
}
func (p *silentPort) Close() error { return nil }

const identTxMax = 96

var (
	identTxOnce sync.Once
	identTxKeys map[string]string // modello → «saluto»
)

const identSilentName = "\x00IDENT:"

// identTranscripts: velocità e primi byte inviati dal driver di ogni modello
// collegato via cavo a una radio che non risponde (chiave del «saluto»). Si
// calcolano una volta sola, tutti insieme (4.3: con le porte simulate di probeReg,
// senza sostituire openPortFunc, anche mentre altre porte sono in ricerca).
func identTranscripts() map[string]string {
	identTxOnce.Do(func() {
		keys := map[string]string{}
		var mu sync.Mutex
		bufs := map[string]*bytes.Buffer{}
		var list []*genModel
		for _, m := range genModels {
			if m.Conn == "" {
				list = append(list, m)
				bufs[m.ID] = &bytes.Buffer{}
			}
		}
		probeReg.Lock()
		probeReg.silent = func(name string, baud int) (Port, error) {
			tx := bufs[strings.TrimPrefix(name, identSilentName)]
			if tx == nil {
				return nil, errors.New("porta sconosciuta")
			}
			mu.Lock()
			fmt.Fprintf(tx, "<%d>", baud)
			mu.Unlock()
			return &silentPort{tx: tx, max: identTxMax, mu: &mu}, nil
		}
		probeReg.Unlock()
		var wg sync.WaitGroup
		for _, m := range list {
			wg.Add(1)
			go func(m *genModel) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
				defer cancel()
				defer func() { _ = recover() }()
				_, _, _ = m.Fam.Download(ctx, m, identSilentName+m.ID, func(int, int, string) {})
			}(m)
		}
		wg.Wait()
		probeReg.Lock()
		probeReg.silent = nil
		probeReg.Unlock()
		for _, m := range list {
			mu.Lock()
			b := bufs[m.ID].Bytes()
			if len(b) > identTxMax+32 {
				b = b[:identTxMax+32]
			}
			keys[m.ID] = string(b)
			mu.Unlock()
		}
		identTxKeys = keys
	})
	return identTxKeys
}

// ---------------------------------------------------------------- porta controllata

// probePort: porta vera con le attese della prima risposta abbreviate; registra
// i primi byte ricevuti (identificativo della radio).
type probePort struct {
	Port
	st *probeState
}

// 4.9: la porta controllata passa alla porta vera la parità pari e i due bit di
// stop. Prima non lo faceva: durante «Identifica» la Lanchonlh HG-UV98 (parità
// pari) e la Quansheng TG-UV2+ (due bit di stop) parlavano con la porta in 8N1
// e non venivano riconosciute.
func (p probePort) SetEvenParity(on bool) error {
	p.st.lineSet('E', fmt.Sprint(on))
	if x, ok := p.Port.(interface{ SetEvenParity(bool) error }); ok {
		return x.SetEvenParity(on)
	}
	return nil
}

func (p probePort) SetTwoStopBits(on bool) error {
	p.st.lineSet('S', fmt.Sprint(on))
	if x, ok := p.Port.(interface{ SetTwoStopBits(bool) error }); ok {
		return x.SetTwoStopBits(on)
	}
	return nil
}

// SetRTS (4.9): RTS spento per la Retevis C2
func (p probePort) SetRTS(on bool) error {
	p.st.lineSet('R', fmt.Sprint(on))
	if x, ok := p.Port.(interface{ SetRTS(bool) error }); ok {
		return x.SetRTS(on)
	}
	return nil
}

// SetDTR (4.13): DTR spento per il firmware IJV V6
func (p probePort) SetDTR(on bool) error {
	p.st.lineSet('D', fmt.Sprint(on))
	if x, ok := p.Port.(interface{ SetDTR(bool) error }); ok {
		return x.SetDTR(on)
	}
	return nil
}

func (p probePort) unwrapPort() Port { return p.Port }

// SetBaud (4.31): la velocità fa parte della «domanda» (vedi identDead)
func (p probePort) SetBaud(baud int) error {
	p.st.lineSet('B', fmt.Sprint(baud))
	return p.Port.SetBaud(baud)
}

type probeState struct {
	mu      sync.Mutex
	key     string // porta (vedi probeKey)
	tx      []byte
	genuine bool // la radio ha risposto (byte «puliti» diversi dall'eco di quelli inviati)
	first   []byte
	// 4.4: risposta SICURA di una radio, per il riquadro «Identificazione Radio In
	// Corso»: byte che non sono disturbi della linea (0x00, 0xFF e simili, per es.
	// cavo senza radio o con la radio spenta) né l'eco di quanto inviato, anche nei
	// tentativi precedenti sulla stessa porta; almeno identSureBytes byte o due
	// letture distinte (es. conferma 0x06 e poi l'identificativo)
	clean    []byte
	hits     int
	sure     bool
	onAnswer func() // chiamata una volta, quando la risposta è sicura
	// spont (4.7): byte che la radio trasmette da sola (vedi identSpontSample)
	spont []byte
	// listened (4.7): tempo passato ad ascoltare la porta (identSpontSample), che
	// non conta nell'attesa massima del tentativo (identSilentMax); listenFrom:
	// inizio dell'ascolto in corso
	listened   time.Duration
	listenFrom time.Time
	ctx        context.Context // tentativo (l'ascolto si interrompe con la ricerca)
	// 4.31: giro veloce e domande ripetute (vedi identFastWait, identReplayWait)
	fast     bool
	line     string // velocità e impostazioni della linea, parte della «domanda»
	lineKV   map[byte]string
	stim     []byte    // byte inviati dopo l'ultima lettura
	pend     string    // domanda in attesa di risposta
	pendDead bool      // … già rimasta senza risposta in questo giro
	timeouts int       // letture rimaste senza risposta in questo tentativo
	deadAt   time.Time // ultima lettura rimasta senza risposta
	streak   int       // letture a vuoto di fila senza una domanda nuova
}

// identDead: per porta, le domande (linea + byte inviati) rimaste senza risposta
// nel giro in corso
var identDead struct {
	sync.Mutex
	m map[string]map[string]bool
}

// identDeadReset: comincia un giro nuovo sulla porta
func identDeadReset(key string) {
	identDead.Lock()
	if identDead.m == nil {
		identDead.m = map[string]map[string]bool{}
	}
	identDead.m[key] = map[string]bool{}
	identDead.Unlock()
}

func identDeadHas(key, q string) bool {
	identDead.Lock()
	defer identDead.Unlock()
	return identDead.m[key][q]
}

func identDeadAdd(key, q string) {
	identDead.Lock()
	if m := identDead.m[key]; m != nil && len(m) < 4096 {
		m[q] = true
	}
	identDead.Unlock()
}

// lineSet: impostazione della linea (B velocità, E parità, S bit di stop, R RTS,
// D DTR); una porta appena aperta (O) riparte dalle impostazioni normali
func (s *probeState) lineSet(k byte, v string) {
	s.mu.Lock()
	if k == 'O' {
		s.lineKV = map[byte]string{}
		k = 'B'
	}
	if s.lineKV == nil {
		s.lineKV = map[byte]string{}
	}
	s.lineKV[k] = v
	s.line = ""
	for _, c := range [...]byte{'B', 'E', 'S', 'R', 'D'} {
		if x, ok := s.lineKV[c]; ok {
			s.line += string(c) + x + "|"
		}
	}
	s.mu.Unlock()
}

// silence: tempo del tentativo senza l'ascolto della porta
func (s *probeState) silence(start time.Time) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := time.Since(start) - s.listened
	if !s.listenFrom.IsZero() {
		d -= time.Since(s.listenFrom)
	}
	return d
}

var identSureHook func(key string, clean, echo []byte) // solo per le prove

// identSureBytes: byte «puliti» che bastano da soli per una risposta sicura
const identSureBytes = 4

// lineNoise: byte tipici dei disturbi di una linea seriale senza nulla collegato
// (riposo, interruzione, trame incomplete: solo uni o solo zeri a gruppi)
func lineNoise(b byte) bool {
	switch b {
	case 0x00, 0xFF, 0x80, 0xC0, 0xE0, 0xF0, 0xF8, 0xFC, 0xFE, 0x01, 0x03, 0x07, 0x0F, 0x1F, 0x3F, 0x7F:
		return true
	}
	return false
}

func cleanBytes(b []byte) []byte {
	var out []byte
	for _, c := range b {
		if !lineNoise(c) {
			out = append(out, c)
		}
	}
	return out
}

// identSent: ultimi byte inviati su ogni porta durante le ricerche (eco ritardata
// di un tentativo precedente, letta all'inizio del successivo)
var identSent struct {
	sync.Mutex
	m map[string][]byte
}

const identSentMax = 2048

func identSentAdd(key string, b []byte) {
	identSent.Lock()
	defer identSent.Unlock()
	if identSent.m == nil {
		identSent.m = map[string][]byte{}
	}
	h := append(identSent.m[key], b...)
	if len(h) > identSentMax {
		h = append([]byte{}, h[len(h)-identSentMax:]...)
	}
	identSent.m[key] = h
}

// identEchoTail: byte inviati alla fine del tentativo precedente che possono
// tornare come eco ritardata all'inizio del successivo
const identEchoTail = 512

func identSentTail(key string, n int) []byte {
	h := identSentGet(key)
	return h[max(0, len(h)-n):]
}

// isSubsequence: tutti i byte di a compaiono in b nello stesso ordine
func isSubsequence(a, b []byte) bool {
	i := 0
	for _, c := range b {
		if i < len(a) && a[i] == c {
			i++
		}
	}
	return i == len(a)
}

func identSentGet(key string) []byte {
	identSent.Lock()
	defer identSent.Unlock()
	return append([]byte{}, identSent.m[key]...)
}

// radioSure: la radio ha risposto davvero (vedi probeState.sure)
func (s *probeState) radioSure() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sure
}

// identProbing (4.9): sulla porta è in corso un tentativo di «Identifica»
// (i driver che provano più saluti possono cambiarne l'ordine per rispondere
// entro identSilentMax)
func identProbing(port string) bool {
	probeReg.Lock()
	defer probeReg.Unlock()
	return probeReg.m[probeKey(port)] != nil
}

// identAnswerKey: valore del contesto con la funzione da chiamare quando una radio
// risponde durante la ricerca (vedi probeState.onAnswer)
type identAnswerKey struct{}

// identFastKey (4.31): valore del contesto che chiede un tentativo del giro veloce
type identFastKey struct{}

// idle (4.31, giro veloce): dopo una lettura a vuoto il driver non ha fatto una
// domanda nuova entro d, oppure continua a ripetere domande già fatte
func (s *probeState) idle(d time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.genuine || s.deadAt.IsZero() {
		return false
	}
	return s.streak >= 2 || (time.Since(s.deadAt) >= d && len(s.stim) == 0 && (s.pend == "" || s.pendDead))
}

func (s *probeState) answered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.genuine
}

func (p probePort) Write(b []byte) error {
	p.st.mu.Lock()
	if len(p.st.tx) < 4096 {
		p.st.tx = append(p.st.tx, b...)
	}
	if len(p.st.stim) < 512 {
		p.st.stim = append(p.st.stim, b...)
	}
	key := p.st.key
	p.st.mu.Unlock()
	identSentAdd(key, b)
	return p.Port.Write(b)
}

func (p probePort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	// 4.31, giro veloce: prima della prima risposta ogni lettura è l'attesa di una
	// «domanda» (i byte inviati dopo la lettura precedente). Attesa del giro veloce
	// la prima volta; quasi nulla se la stessa domanda è già rimasta senza risposta
	// in questo giro o se dopo una lettura a vuoto non è stato inviato niente di nuovo
	if !p.st.answered() && !p.st.fast && timeout > identFirstWait {
		timeout = identFirstWait
	}
	if !p.st.answered() && p.st.fast {
		s := p.st
		s.mu.Lock()
		first, replay := identFastWait, identReplayWaitFast
		if len(s.stim) > 0 {
			// domanda nuova: resta in sospeso finché una lettura non va a vuoto (in
			// mezzo può esserci la lettura dell'eco del cavo)
			s.pend = s.line + string(s.stim)
			s.stim = s.stim[:0]
			s.pendDead = identDeadHas(s.key, s.pend)
		}
		if s.pendDead || (s.pend == "" && s.timeouts > 0) {
			first = replay
		}
		s.mu.Unlock()
		if timeout > first {
			timeout = first
		}
	}
	err := p.Port.ReadFull(ctx, b, timeout)
	if err != nil && ctx.Err() == nil && p.st.fast && !p.st.answered() {
		s := p.st
		s.mu.Lock()
		s.timeouts++
		q, dead := s.pend, s.pendDead
		s.pend, s.pendDead = "", false
		s.deadAt = time.Now()
		if q != "" && !dead {
			s.streak = 0
		} else {
			s.streak++ // domanda ripetuta o lettura senza domanda
		}
		s.mu.Unlock()
		if q != "" && !dead {
			identDeadAdd(s.key, q)
		}
	}
	if err == nil && len(p.st.spont) > 0 {
		// 4.7: byte uguali a quelli che la radio trasmette da sola (vedi
		// identSpontSample): per il driver è come se la radio tacesse
		cyc := append(append([]byte{}, p.st.spont...), p.st.spont...)
		if bytes.Contains(cyc, b) || (identBeacon(p.st.spont) && identBeacon(b) && b[0] == p.st.spont[0]) {
			return errTimeout
		}
	}
	if err == nil {
		s := p.st
		s.mu.Lock()
		// 4.4: si considerano solo i byte che non sono disturbi della linea e non
		// sono l'eco di quanto inviato (anche nei tentativi precedenti): prima
		// bastava un byte qualsiasi diverso dall'eco, e con un cavo senza radio i
		// disturbi facevano durare ogni tentativo come con una radio
		c := cleanBytes(b)
		if len(s.spont) > 0 && len(c) > 0 {
			// 4.7: byte uguali a quelli che la radio trasmette da sola: non sono una risposta
			cyc := cleanBytes(append(append([]byte{}, s.spont...), s.spont...))
			if bytes.Contains(cyc, c) || isSubsequence(c, cyc) {
				c = nil
			}
		}
		if !s.sure && len(c) > 0 {
			s.clean = append(s.clean, c...)
			s.hits++
			// per proseguire il tentativo conta l'eco di questo tentativo (come prima
			// della 4.4); quella dei precedenti solo per il riquadro: altrimenti una
			// conferma 0x06 già inviata da un altro tentativo sembrerebbe un'eco
			if !bytes.Contains(cleanBytes(s.tx), s.clean) {
				s.genuine = true // la radio ha risposto: il tentativo prosegue
				echo := cleanBytes(append(identSentGet(s.key), s.tx...))
				// eco con dei byte persi (es. letture scartate dai driver): i byte
				// ricevuti compaiono, nello stesso ordine, tra quelli inviati in
				// questo tentativo o alla fine del precedente
				recent := cleanBytes(append(identSentTail(s.key, identEchoTail), s.tx...))
				if !bytes.Contains(echo, s.clean) && !isSubsequence(s.clean, recent) && (len(s.clean) >= identSureBytes || s.hits >= 2) {
					if identSureHook != nil {
						identSureHook(s.key, s.clean, echo)
					}
					s.sure = true // risposta sicura: riquadro visibile
					if s.onAnswer != nil {
						go s.onAnswer()
					}
				}
			}
		}
		if len(s.first) < 64 {
			s.first = append(s.first, b[:min(len(b), 64-len(s.first))]...)
		}
		s.mu.Unlock()
	}
	return err
}

// 4.3: porte controllate registrate per nome (prima l'identificazione sostituiva
// openPortFunc, quindi si poteva cercare una sola porta alla volta). I driver
// aprono le porte con openPort: se sulla porta è in corso un tentativo di
// identificazione, la porta aperta è quella controllata.
var probeReg struct {
	sync.Mutex
	m      map[string]*probeState
	silent func(name string, baud int) (Port, error) // trascrizioni dei saluti (identTranscripts)
}

// probeKey: le radio DMR via USB si aprono come «HID:…» o «DFU:…», cercate come «USB»
func probeKey(name string) string {
	if strings.HasPrefix(name, "HID:") || strings.HasPrefix(name, "DFU:") {
		return "USB"
	}
	return strings.ToUpper(name)
}

// openPort: apertura di una porta da parte dei driver (vedi probeReg)
func openPort(name string, baud int) (Port, error) {
	probeReg.Lock()
	silent, st := probeReg.silent, probeReg.m[probeKey(name)]
	probeReg.Unlock()
	if strings.HasPrefix(name, identSilentName) && silent != nil {
		return silent(name, baud)
	}
	p, err := openPortFunc(name, baud)
	if err != nil || st == nil {
		return p, err
	}
	// 4.7: radio che trasmette da sola (es. in modalità aggiornamento firmware, con
	// i suoi messaggi periodici): prima del tentativo si ascolta la porta senza
	// inviare nulla; quei byte non contano come risposta della radio, così ogni
	// tentativo finisce presto come con una porta senza radio (prima ogni modello
	// sembrava rispondere e la ricerca durava alcuni minuti)
	st.lineSet('O', fmt.Sprint(baud))
	st.mu.Lock()
	fast := st.fast
	st.mu.Unlock()
	// 4.31: nel giro veloce si ascolta solo dove qualcosa ha già risposto (o se
	// l'ascolto di meno di un minuto fa è ancora valido): senza byte in arrivo non
	// c'è nulla da scartare
	if fast && !identSpontNeeded(probeKey(name), baud) {
		return probePort{p, st}, nil
	}
	st.mu.Lock()
	st.listenFrom = time.Now()
	st.mu.Unlock()
	sp := identSpontSample(st.ctx, probeKey(name), baud, p)
	st.mu.Lock()
	st.spont = sp
	st.listened += time.Since(st.listenFrom) // il tempo d'ascolto non conta come silenzio della radio
	st.listenFrom = time.Time{}
	st.mu.Unlock()
	return probePort{p, st}, nil
}

// identSpont: byte trasmessi dalla radio senza che le sia stato inviato nulla,
// per porta e velocità (validi identSpontTTL)
var identSpont struct {
	sync.Mutex
	m map[string]identSpontRec
}

type identSpontRec struct {
	at   time.Time
	data []byte
}

var (
	identSpontTTL = time.Minute
	// ascolto: abbastanza per i messaggi periodici più lenti (il caricatore delle
	// Bajeton si presenta ogni 0,7 s); una volta al minuto per velocità
	identSpontListen     = time.Second
	identSpontListen9600 = time.Second
)

// identSpontSample: ascolto della porta appena aperta (una volta al minuto per
// velocità); i byte che non sono disturbi della linea si ricordano. Se la radio
// tace si torna subito (al massimo identSpontListen)
func identSpontSample(ctx context.Context, key string, baud int, p Port) []byte {
	k := fmt.Sprintf("%s|%d", key, baud)
	identSpont.Lock()
	if r, ok := identSpont.m[k]; ok && time.Since(r.at) < identSpontTTL {
		identSpont.Unlock()
		return r.data
	}
	identSpont.Unlock()
	var got []byte
	buf := make([]byte, 1)
	listen := identSpontListen
	if baud == 9600 {
		listen = identSpontListen9600
	}
	end := time.Now().Add(listen)
	for len(got) < 512 {
		left := time.Until(end)
		if left <= 0 {
			break
		}
		if err := p.ReadFull(ctx, buf, left); err != nil {
			break
		}
		got = append(got, buf[0])
	}
	if ctx.Err() != nil {
		return nil // ricerca interrotta: ascolto incompleto, non si ricorda
	}
	c := got
	if len(cleanBytes(got)) < identSureBytes && !identBeacon(got) {
		c = nil // qualche disturbo isolato: non è una trasmissione
	}
	_ = p.Purge()
	identSpont.Lock()
	if identSpont.m == nil {
		identSpont.m = map[string]identSpontRec{}
	}
	identSpont.m[k] = identSpontRec{time.Now(), c}
	identSpont.Unlock()
	return c
}

// identSpontWant (4.31): velocità a cui, nel giro veloce, qualcosa ha risposto: da
// lì in poi anche il giro veloce ascolta la porta prima dei tentativi
var identSpontWant struct {
	sync.Mutex
	m map[string]time.Time
}

func identSpontAsk(key string) {
	identSpontWant.Lock()
	if identSpontWant.m == nil {
		identSpontWant.m = map[string]time.Time{}
	}
	identSpontWant.m[key] = time.Now()
	identSpontWant.Unlock()
}

// identSpontNeeded: nel giro veloce, ascolto da fare (qualcosa ha risposto su
// questa porta da meno di identSpontTTL) o già fatto e ancora valido
func identSpontNeeded(key string, baud int) bool {
	identSpontWant.Lock()
	at, ok := identSpontWant.m[key]
	identSpontWant.Unlock()
	if ok && time.Since(at) < identSpontTTL {
		return true
	}
	identSpont.Lock()
	defer identSpont.Unlock()
	r, ok := identSpont.m[fmt.Sprintf("%s|%d", key, baud)]
	return ok && time.Since(r.at) < identSpontTTL
}

// identBeacon: un byte ripetuto (es. 0x01 del caricatore delle Bajeton in modalità
// aggiornamento, ogni 0,7 s a 9600 baud)
func identBeacon(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c != b[0] || c == 0x00 || c == 0xFF {
			return false
		}
	}
	return true
}

// identSpontSeen: la porta ha trasmesso da sola di recente (a una velocità qualsiasi)
func identSpontSeen(key string) bool {
	identSpont.Lock()
	defer identSpont.Unlock()
	for k, r := range identSpont.m {
		if strings.HasPrefix(k, key+"|") && len(r.data) > 0 && time.Since(r.at) < identSpontTTL {
			return true
		}
	}
	return false
}

// rememberRadio (4.3): il modello riconosciuto va in testa all'elenco delle radio
// note (al massimo identKnownMax): le ricerche successive cominciano da lì
const identKnownMax = 24

var rememberMu sync.Mutex // più porte possono riconoscere una radio nello stesso momento

func (a *App) rememberRadio(kind string) {
	rememberMu.Lock()
	defer rememberMu.Unlock()
	st := a.store.Get()
	list := []string{kind}
	for _, k := range st.KnownRadios {
		if k != kind && len(list) < identKnownMax {
			list = append(list, k)
		}
	}
	if strings.Join(list, ",") == strings.Join(st.KnownRadios, ",") {
		return
	}
	st.KnownRadios = list
	_ = a.store.SaveSettings(st)
}

// identPortLock: una sola ricerca alla volta per porta (porte diverse in parallelo)
var identPorts struct {
	sync.Mutex
	m map[string]*sync.Mutex
}

func identPortLock(port string) *sync.Mutex {
	identPorts.Lock()
	defer identPorts.Unlock()
	if identPorts.m == nil {
		identPorts.m = map[string]*sync.Mutex{}
	}
	k := probeKey(port)
	if identPorts.m[k] == nil {
		identPorts.m[k] = &sync.Mutex{}
	}
	return identPorts.m[k]
}

// probe: esegue fn con la porta controllata; se la radio non risponde entro
// identSilentMax il tentativo si interrompe (silent = true).
func probe(ctx context.Context, port string, fn func(ctx context.Context) error) (st *probeState, silent bool, err error) {
	key := probeKey(port)
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()
	st = &probeState{key: key, ctx: pctx}
	st.onAnswer, _ = ctx.Value(identAnswerKey{}).(func())
	st.fast, _ = ctx.Value(identFastKey{}).(bool)
	probeReg.Lock()
	if probeReg.m == nil {
		probeReg.m = map[string]*probeState{}
	}
	probeReg.m[key] = st
	probeReg.Unlock()
	defer func() {
		probeReg.Lock()
		if probeReg.m[key] == st {
			delete(probeReg.m, key)
		}
		probeReg.Unlock()
	}()
	stop := make(chan struct{})
	timedOut := false
	var tmu sync.Mutex
	silentMax := identSilentMax
	if st.fast {
		silentMax = identFastSilent
	}
	go func() {
		start := time.Now()
		wait := silentMax
		if st.fast {
			wait = 10 * time.Millisecond
		}
		for {
			select {
			case <-time.After(wait):
				if st.fast && st.silence(start) < silentMax && !st.idle(identFastIdle) {
					continue
				}
				if st.fast && !st.answered() {
					tmu.Lock()
					timedOut = true
					tmu.Unlock()
					cancel()
					return
				}
				// 4.7: l'ascolto della porta prima del tentativo non è silenzio della
				// radio (con un ascolto più lungo di identSilentMax il tentativo si
				// chiuderebbe prima di inviare il saluto)
				if d := st.silence(start); d < silentMax {
					wait = max(silentMax-d, 20*time.Millisecond)
					continue
				}
				if !st.answered() {
					tmu.Lock()
					timedOut = true
					tmu.Unlock()
					cancel()
				}
				return
			case <-stop:
				return
			}
		}
	}()
	func() {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("errore interno: %v", e)
			}
		}()
		err = fn(pctx)
	}()
	close(stop)
	if ctx.Err() != nil {
		return st, false, errCancelled
	}
	if st.fast && err != nil && st.answered() {
		// qualcosa ha risposto ma il tentativo è fallito: da qui, anche nel giro
		// veloce, si ascolta la porta prima dei tentativi (radio che trasmette da sola)
		identSpontAsk(key)
	}
	tmu.Lock()
	to := timedOut
	tmu.Unlock()
	if err != nil && (!st.answered() || to) {
		return st, true, err
	}
	return st, false, err
}

// ---------------------------------------------------------------- risultato

type identOther struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type identField struct {
	L string `json:"l"`
	V string `json:"v"`
}

type identResult struct {
	Found   bool         `json:"found"`
	Kind    string       `json:"kind"`  // scheda della radio (k67, baj o driver generico)
	Name    string       `json:"name"`  // nome completo
	Brand   string       `json:"brand"` // marca
	Model   string       `json:"model"`
	Type    string       `json:"type"`
	Fields  []identField `json:"fields"`
	Others  []identOther `json:"others"`  // altri modelli possibili (stesso collegamento e stessa memoria)
	Answers []string     `json:"answers"` // risposte della radio non andate a buon fine
	Tried   int          `json:"tried"`   // protocolli provati
	// Ambiguous: la radio ha la memoria di più modelli che non si distinguono
	// via cavo (stesso protocollo, nessun segno del modello nella memoria)
	Ambiguous bool `json:"ambiguous"`
	// Card (4.6): radio della galleria (scheda del catalogo) con il nome mostrato
	Card string `json:"card,omitempty"`
	// Spont (4.7): nessuna radio riconosciuta, ma la radio trasmetteva da sola
	// (tipico della modalità aggiornamento firmware)
	Spont bool `json:"spont,omitempty"`
	// FwHint (4.7): dove vedere sulla radio la versione del firmware (radio con
	// la stessa gestione)
	FwHint  string `json:"fwHint,omitempty"`
	Seconds int    `json:"seconds"`
	Backup  string `json:"backup"`
	Count   int    `json:"count"` // canali presenti
}

func (r *identResult) add(l, f string, a ...interface{}) {
	v := strings.TrimSpace(fmt.Sprintf(f, a...))
	if v != "" {
		r.Fields = append(r.Fields, identField{l, v})
	}
}

// radio letta in attesa di essere aperta nella sua scheda
type identSession struct {
	kind  string
	kinds []string // modelli proposti (il primo è quello identificato)
	sess  *radioSession
	mems  map[string][]byte // memoria letta con il driver di ogni modello proposto
	// 4.32: modelli proposti senza averli riletti (vedi identRun): la memoria si
	// legge con il loro driver quando l'utente apre quella scheda
	lazy map[string]bool
	port string
	at   time.Time // fine dell'identificazione
	mu   sync.Mutex
}

// ---------------------------------------------------------------- testo dei byte

// printable: testo leggibile dei byte (nil se non è testo)
func identText(b []byte) string {
	b = trimBytes(b, 0x00, 0xFF, ' ')
	if len(b) == 0 {
		return ""
	}
	n := 0
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			n++
		} else if c != 0 && c != 0xFF {
			return ""
		}
	}
	if n < 2 {
		return ""
	}
	var sb strings.Builder
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			sb.WriteByte(c)
		} else {
			sb.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

// identRuns: tratti di testo leggibile (almeno 4 caratteri) nei byte ricevuti
// dalla radio, es. «BT-11U V1.00» o «PROGRAMCOLORPROU»; "" se non ce ne sono
func identRuns(b []byte) string {
	var runs []string
	cur := []byte{}
	flush := func() {
		if t := strings.TrimSpace(string(cur)); len(t) >= 4 {
			runs = append(runs, strings.Join(strings.Fields(t), " "))
		}
		cur = cur[:0]
	}
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			cur = append(cur, c)
		} else {
			flush()
		}
	}
	flush()
	return strings.Join(runs, " · ")
}

func bandsText(bs []band) string {
	var out []string
	for _, b := range bs {
		out = append(out, fmt.Sprintf("%s–%s MHz", mhzText(b.Lo), mhzText(b.Hi)))
	}
	return strings.Join(out, ", ")
}

func mhzText(hz int64) string {
	s := fmt.Sprintf("%.4f", float64(hz)/1e6)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return strings.Replace(s, ".", ",", 1)
}

// ---------------------------------------------------------------- dati DMR

// BCD a 8 cifre (ID DMR della RD-5R e della DM-1801, come GET_ID di dmrconfig)
func bcdID(b []byte) (uint32, bool) {
	var v uint32
	for _, c := range b {
		hi, lo := c>>4, c&15
		if hi > 9 || lo > 9 {
			return 0, false
		}
		v = v*100 + uint32(hi)*10 + uint32(lo)
	}
	return v, true
}

func bcdDigits(b []byte) (string, bool) {
	var sb strings.Builder
	for _, c := range b {
		if c>>4 > 9 || c&15 > 9 {
			return "", false
		}
		sb.WriteByte('0' + c>>4)
		sb.WriteByte('0' + c&15)
	}
	return sb.String(), true
}

// dmrInfo: ID DMR, nome della radio, data dell'ultima programmazione e versione
// del programma che l'ha scritta, dalla memoria letta (stessi campi e stesse
// posizioni di dmrconfig: rd5r.c / gd77.c / dm1801.c print_id e print_version,
// uv380.c per la DM-1701).
func dmrInfo(m *genModel, mem []byte, r *identResult) {
	switch m.Fam.(type) {
	case rdtyFam:
		const ts, gs = 0x88, 0xE0
		if len(mem) < gs+12 {
			return
		}
		if id, ok := bcdID(mem[gs+8 : gs+12]); ok {
			r.add("ID DMR", "%d", id)
		}
		if n := mem[gs]; n != 0 && n != 0xFF {
			r.add("Nome della radio", "%s", identText(mem[gs:gs+8]))
		}
		if mem[ts] != 0xFF {
			if d, ok := bcdDigits(mem[ts : ts+6]); ok {
				r.add("Ultima programmazione", "%s-%s-%s %s:%s", d[0:4], d[4:6], d[6:8], d[8:10], d[10:12])
			}
		}
	case tytFam:
		const ts, gs = 0x2001, 0x2040
		if len(mem) < gs+0x90 {
			return
		}
		if id := uint32(mem[gs+0x44]) | uint32(mem[gs+0x45])<<8 | uint32(mem[gs+0x46])<<16; id != 0xFFFFFF {
			r.add("ID DMR", "%d", id)
		}
		if u := mem[gs+0x70 : gs+0x90]; !(u[0] == 0 && u[1] == 0) && !(u[0] == 0xFF && u[1] == 0xFF) {
			var w []uint16
			for i := 0; i+1 < len(u); i += 2 {
				c := uint16(u[i]) | uint16(u[i+1])<<8
				if c == 0 || c == 0xFFFF {
					break
				}
				w = append(w, c)
			}
			r.add("Nome della radio", "%s", strings.TrimSpace(string(utf16.Decode(w))))
		}
		if t := mem[ts : ts+11]; t[0] != 0xFF {
			if d, ok := bcdDigits(t[:7]); ok {
				r.add("Ultima programmazione", "%s-%s-%s %s:%s:%s", d[0:4], d[4:6], d[6:8], d[8:10], d[10:12], d[12:14])
			}
			const cm = "0123456789:;<=>?"
			if v := fmt.Sprintf("V%c%c.%c%c", cm[t[7]&15], cm[t[8]&15], cm[t[9]&15], cm[t[10]&15]); !strings.ContainsAny(v, ":;<=>?") {
				r.add("Versione del programma che l'ha scritta", "%s", v)
			}
		}
	}
}

// ---------------------------------------------------------------- descrizione

func genIdentFill(m *genModel, cands []*genModel, mem []byte, note string, st *probeState, port string, r *identResult) {
	r.Found, r.Kind, r.Name, r.Brand = true, m.ID, m.Name, m.Vendor
	r.Model = strings.TrimSpace(m.Model + " " + m.Variant)
	switch {
	case m.Digital != "":
		r.Type = "Ricetrasmettitore " + m.Digital + " (canali analogici FM e digitali " + m.Digital + ")"
	default:
		r.Type = "Ricetrasmettitore analogico FM"
	}
	if pmr := strings.Contains(m.Name, "PMR") || strings.Contains(m.Name, "FRS"); pmr {
		r.Type += ", uso libero " + map[bool]string{true: "PMR446", false: "FRS/GMRS"}[strings.Contains(m.Name, "PMR")]
	}
	r.add("Marca e modello", "%s", m.Name)
	r.add("Tipo", "%s", r.Type)
	bs := m.Bands
	if bf, ok := m.Fam.(interface {
		Bands(m *genModel, img []byte) []band
	}); ok {
		// gamme scritte nella memoria della radio (come CHIRP)
		func() {
			defer func() { _ = recover() }()
			if b := bf.Bands(m, mem); len(b) > 0 {
				bs = b
			}
		}()
	}
	r.add("Gamme di frequenza", "%s", bandsText(bs))
	r.add("Memorie", "%d canali", m.Slots)
	switch m.Conn {
	case "hid":
		r.add("Collegamento", "USB (dispositivo HID, senza driver)")
	case "dfu":
		r.add("Collegamento", "USB in modalità DFU (driver del costruttore o WinUSB)")
	default:
		r.add("Collegamento", "cavo di programmazione sulla porta %s, %d baud", port, m.Baud)
	}
	r.add("Firmware / identificativo", "%s", note)
	dmrInfo(m, mem, r)
	if st != nil {
		r.add("Testo inviato dalla radio", "%s", identRuns(st.first))
	}
	r.add("Numero di serie", "%s", "non disponibile: il protocollo di programmazione di questa radio non lo comunica")
	seen := map[*genModel]bool{m: true}
	for _, o := range cands {
		if !seen[o] {
			seen[o] = true
			r.Others = append(r.Others, identOther{o.ID, o.Name})
		}
	}
	sort.Slice(r.Others, func(i, j int) bool { return r.Others[i].Name < r.Others[j].Name })
	r.Count = len(packed(genChans(m, mem)))
}

// ---------------------------------------------------------------- scansione

type identReq struct {
	Port   string `json:"port"`   // porta COM o "USB"
	Brand  string `json:"brand"`  // "" tutte, oppure la marca (Baofeng, Bajeton, Quansheng, Retevis…)
	Family string `json:"family"` // "" tutte, analog, dmr
}

type identProbe struct {
	name  string // testo mostrato
	kinds []*genModel
	k67   bool
	baud  int // K67: velocità
	bt    bool
	prio  bool // 4.31: radio già riconosciute su questo PC o con una scheda aperta: attese piene anche nel giro veloce
	rep   bool // 4.31: ripetizione di una radio nota in mezzo al piano (ascolto continuo)
}

var (
	driverBrandsOnce sync.Once
	driverBrands     map[string]map[string]bool // driver → marche della galleria che lo usano
)

// modelBrands: marche (in minuscolo) delle radio della galleria gestite dal
// driver; senza voci nella galleria vale la marca indicata da CHIRP
func modelBrands(m *genModel) map[string]bool {
	driverBrandsOnce.Do(func() {
		driverBrands = map[string]map[string]bool{}
		for _, c := range catalog {
			for _, d := range c.Drivers {
				if driverBrands[d] == nil {
					driverBrands[d] = map[string]bool{}
				}
				driverBrands[d][strings.ToLower(c.Brand)] = true
			}
		}
	})
	if b := driverBrands[m.ID]; len(b) > 0 {
		return b
	}
	return map[string]bool{strings.ToLower(m.Vendor): true}
}

func identWanted(req *identReq, brand string, dmr bool) bool {
	return identWantedAny(req, map[string]bool{strings.ToLower(brand): true}, dmr)
}

func identWantedAny(req *identReq, brands map[string]bool, dmr bool) bool {
	if req.Brand != "" && !brands[strings.ToLower(req.Brand)] {
		return false
	}
	switch req.Family {
	case "analog":
		return !dmr
	case "dmr":
		return dmr
	}
	return true
}

// identPlan: tentativi nell'ordine (prima le radio con una scheda aperta)
func (a *App) identPlan(req *identReq) []identProbe {
	usb := strings.EqualFold(req.Port, "USB")
	var out []identProbe
	if !usb {
		if identWanted(req, "Baofeng", false) {
			// come il programma originale: prima la velocità usata l'ultima volta
			b1, b2 := 19200, 115200
			if a.store.Get().K67Baud == 115200 {
				b1, b2 = b2, b1
			}
			for _, b := range []int{b1, b2} {
				out = append(out, identProbe{name: fmt.Sprintf("Baofeng K67 Pro / 5RH Pro / BF-K6 (%d baud)", b), k67: true, baud: b})
			}
		}
		if identWanted(req, "Bajeton", false) {
			out = append(out, identProbe{name: "Bajeton BJ-8300 (BT-11)", bt: true})
		}
	}
	a.mu.Lock()
	open := map[string]bool{}
	for k := range a.radios {
		open[k] = true
	}
	a.mu.Unlock()
	groups := map[string]*identProbe{}
	var order []string
	var joins []*genModel
	var keys map[string]string
	if !usb {
		keys = identTranscripts()
	}
	for _, m := range genModels {
		if usb != (m.Conn != "") || !identWantedAny(req, modelBrands(m), m.Digital != "") {
			continue
		}
		// 4.14: radio che non rispondono a nessuna richiesta (es. Yaesu FT-60: la
		// copia parte dalla radio, con un tasto): «Identifica» non può riconoscerle
		if _, ok := m.Fam.(interface{ noIdent() }); ok {
			continue
		}
		key := "usb " + m.ID // via USB: un tentativo per modello (se la radio manca, l'apertura fallisce subito)
		if !usb {
			key = keys[m.ID]
			// 4.14: famiglie con lo stesso saluto di altre ma con un dialogo diverso
			// subito dopo (vedi ft4Fam.identGroup): gruppo a parte
			if f, ok := m.Fam.(interface{ identGroup() string }); ok {
				key += "|" + f.identGroup()
			}
		}
		g := groups[key]
		if g == nil {
			g = &identProbe{}
			groups[key] = g
			order = append(order, key)
		}
		if open[m.ID] {
			g.kinds = append([]*genModel{m}, g.kinds...)
		} else {
			g.kinds = append(g.kinds, m)
		}
		if j, ok := m.Opt.(interface{ identJoin() string }); ok && !usb && j.identJoin() != "" {
			joins = append(joins, m)
		}
	}
	// 4.32: modelli con un saluto proprio che possono rispondere anche a quello
	// di un altro gruppo (vedi rt22Opt.identJoin): restano fra le candidate di
	// quel gruppo, in coda
	for _, m := range joins {
		target := m.Opt.(interface{ identJoin() string }).identJoin()
		for _, k := range order {
			g, has, in := groups[k], false, false
			for _, o := range g.kinds {
				has = has || o.ID == target
				in = in || o == m
			}
			if has && !in {
				g.kinds = append(g.kinds, m)
			}
		}
	}
	for _, k := range order {
		g := groups[k]
		g.name = g.kinds[0].Name
		if n := len(g.kinds) - 1; n > 0 {
			g.name += fmt.Sprintf(" (e altri %d modelli con lo stesso collegamento)", n)
			if n == 1 {
				g.name = g.kinds[0].Name + " (e " + g.kinds[1].Name + ")"
			}
		}
		out = append(out, *g)
	}
	// 4.3: prima le radio già riconosciute (le più recenti prima), poi quelle con
	// una scheda aperta, poi le altre nell'ordine solito
	known := map[string]int{}
	for i, k := range a.store.Get().KnownRadios {
		if _, ok := known[k]; !ok {
			known[k] = i
		}
	}
	rank := func(kind string) int {
		if i, ok := known[kind]; ok {
			return i
		}
		if open[kind] {
			return 1000
		}
		if identPopular[kind] {
			return 1500
		}
		return 2000
	}
	probeRank := func(p identProbe) int {
		switch {
		case p.k67:
			return rank(kindXLC)
		case p.bt:
			return rank(kindDAT)
		}
		r := 2000
		for _, m := range p.kinds {
			r = min(r, rank(m.ID))
		}
		return r
	}
	// dentro un gruppo (stesso saluto) l'ordine resta quello solito, con prima le
	// schede aperte come nella 4.2: le radio conosciute spostano solo il gruppo.
	// Riordinarle anche qui cambierebbe quali memorie del gruppo si rileggono
	// (al massimo identMaxRereads) e una radio con la memoria di un altro modello
	// del gruppo (es. Retevis RT22 dopo una Baofeng T22 conosciuta) verrebbe
	// scambiata per quello
	sort.SliceStable(out, func(i, j int) bool { return probeRank(out[i]) < probeRank(out[j]) })
	for i := range out {
		out[i].prio = probeRank(out[i]) < 1500
	}
	return out
}

// identPopular (4.7): radio molto diffuse, provate prima delle altre alla prima
// identificazione (prima la Quansheng UV-K5, ultima del giro, arrivava dopo circa
// un minuto). Si sposta solo il gruppo, come per le radio già riconosciute
var identPopular = map[string]bool{
	kindXLC: true, kindDAT: true, // K67 Pro e BJ-8300 restano le prime
	"qs-uvk5": true, "bf-uv17pro": true, "bf-uv17progps": true, "bf-uv17": true,
	"bj-9000": true, "td-tdh8": true, "bj-7800": true, "bj-7700": true, "bf-uv5r": true, "bf-888s": true,
}

func (a *App) runIdentify(ctx context.Context, j *cableJob, req *identReq) (*identResult, error) {
	l := identPortLock(req.Port)
	l.Lock()
	defer l.Unlock()
	start := time.Now()
	res := &identResult{Fields: []identField{}, Others: []identOther{}, Answers: []string{}}
	j.set("Preparazione dei protocolli", true)
	plan := a.identPlan(req)
	if len(plan) == 0 {
		return nil, errors.New("nessuna radio corrisponde alla ricerca")
	}
	if j.Auto {
		plan = identInterleaveKnown(plan, a.knownProbes(plan))
	}
	res.Tried = 0
	// 4.31: giro veloce e poi giro con le attese piene (vedi identFastWait). Via
	// USB non serve (se la radio manca l'apertura fallisce subito). Nell'ascolto
	// continuo i giri si ripetono: uno sì e uno no si fa anche quello pieno
	passes := []bool{true, false}
	switch {
	case strings.EqualFold(req.Port, "USB"):
		passes = []bool{false}
	case j.Auto && identRoundNext(probeKey(req.Port))%2 == 1:
		passes = []bool{true}
	}
	var err error
	for n, fast := range passes {
		identDeadReset(probeKey(req.Port))
		res, err = a.identRun(ctx, j, req, plan, start, res, identPass{fast: fast, n: n, of: len(passes)})
		if err != nil || res == nil || res.Found {
			break
		}
	}
	if err == nil {
		// 4.6: nome esatto della radio (e delle altre con la stessa gestione)
		a.identCardNames(res)
		if res != nil && res.Found && j.sess != nil {
			kinds := []string{res.Kind}
			for _, o := range res.Others {
				k, _ := splitKindCard(o.Kind)
				kinds = append(kinds, k)
			}
			j.sess.kinds = append(j.sess.kinds, kinds...)
		}
	}
	return res, err
}

// identKnownEvery: nella ricerca automatica (ascolto continuo) le radio già
// riconosciute in passato si riprovano dopo ogni identKnownEvery tentativi: una di
// queste, accesa dopo il cavo, si trova in pochi secondi anche a metà giro
const identKnownEvery = 8

// knownProbes: i primi tentativi del piano che contengono radio già riconosciute
func (a *App) knownProbes(plan []identProbe) int {
	known := map[string]bool{}
	for _, k := range a.store.Get().KnownRadios {
		known[k] = true
	}
	n := 0
	for _, p := range plan {
		hit := (p.k67 && known[kindXLC]) || (p.bt && known[kindDAT])
		for _, m := range p.kinds {
			hit = hit || known[m.ID]
		}
		if !hit {
			break
		}
		n++
	}
	return n
}

// identInterleaveKnown: i primi n tentativi (radio note) ripetuti ogni
// identKnownEvery tentativi degli altri
func identInterleaveKnown(plan []identProbe, n int) []identProbe {
	if n == 0 || n >= len(plan) {
		return plan
	}
	n = min(n, 4)
	out := append([]identProbe{}, plan[:n]...)
	for i, p := range plan[n:] {
		if i > 0 && i%identKnownEvery == 0 {
			for _, k := range plan[:n] {
				k.rep = true
				out = append(out, k)
			}
		}
		out = append(out, p)
	}
	return out
}

// identPass (4.31): un giro della ricerca. fast: attese brevi; n di of: per
// l'avanzamento mostrato
type identPass struct {
	fast  bool
	n, of int
}

// identRounds: giri di ascolto continuo fatti su ogni porta
var identRounds struct {
	sync.Mutex
	m map[string]int
}

func identRoundNext(key string) int {
	identRounds.Lock()
	defer identRounds.Unlock()
	if identRounds.m == nil {
		identRounds.m = map[string]int{}
	}
	identRounds.m[key]++
	return identRounds.m[key]
}

func (a *App) identRun(ctx context.Context, j *cableJob, req *identReq, plan []identProbe, start time.Time, res *identResult, pass identPass) (*identResult, error) {
	total, base := len(plan)*pass.of, len(plan)*pass.n
	// tentativo veloce: solo il primo di ogni gruppo e non per le radio già
	// riconosciute su questo PC o con una scheda aperta, che hanno subito le attese piene
	pctxFor := func(fast bool) context.Context {
		if fast {
			return context.WithValue(ctx, identFastKey{}, true)
		}
		return ctx
	}
	answer := func(who string, err error) {
		if err != nil && err != errCancelled {
			line := who + ": " + err.Error()
			for _, x := range res.Answers {
				if x == line {
					return // già annotata nel giro precedente
				}
			}
			res.Answers = append(res.Answers, line)
			j.logf("%s ha risposto, ma: %v", who, err)
		}
	}
	for i, p := range plan {
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		if !pass.fast && pass.n > 0 && p.prio && !p.rep {
			continue // già provato con le attese piene nel giro veloce
		}
		// nel giro veloce le ripetizioni delle radio note sono veloci anche loro
		// (con le attese piene, ogni otto tentativi, lo allungherebbero di molto)
		fast := pass.fast && (!p.prio || p.rep)
		j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", base+i+1, total), true)
		j.prog(base+i, total, "Provo: "+p.name)
		res.Tried++
		switch {
		case p.k67 || p.bt:
			kind := kindXLC
			creq := &cableReq{Op: "read", Radio: kindXLC, Port: req.Port, Baud: p.baud, Password: a.store.Get().K67Pwd}
			if p.bt {
				kind, creq.Radio, creq.Baud, creq.Password = kindDAT, kindDAT, 0, ""
			}
			var file []byte
			var info interface{}
			var k6 bool
			st, silent, err := probe(pctxFor(fast), req.Port, func(pctx context.Context) error {
				var e error
				file, info, k6, e = a.cableRead(pctx, j, creq)
				if e == nil {
					return nil
				}
				return e
			})
			if err == errCancelled && ctx.Err() != nil {
				return nil, errCancelled
			}
			if silent {
				continue
			}
			if err != nil {
				answer(p.name, err)
				_ = sleepCtx(ctx, identPause)
				continue
			}
			a.identFillMain(kind, file, info, k6, st, req.Port, res)
			res.Seconds = int(time.Since(start).Seconds())
			if bp, err := a.saveBackup(kind, file, "identificata"); err == nil {
				res.Backup = bp
			}
			j.sess = &identSession{kind: kind, kinds: []string{kind}, sess: &radioSession{Base: file, K6: k6}}
			a.rememberRadio(kind)
			return res, nil
		default:
			readWith := func(m *genModel, fast bool) (mem []byte, note string, st *probeState, silent bool, err error) {
				j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", base+i+1, total), true)
				j.prog(base+i, total, "Provo: "+m.Name)
				st, silent, err = probe(pctxFor(fast), req.Port, func(pctx context.Context) error {
					var e error
					mem, note, e = m.Fam.Download(pctx, m, req.Port, func(done, total int, msg string) {
						j.mu.Lock()
						j.Phase = "Lettura: " + m.Name
						j.Done, j.Total, j.Msg = done, total, msg
						j.mu.Unlock()
					})
					if e == nil && len(mem) < genMinSize(m) {
						e = fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
					}
					return e
				})
				return
			}
			for k, m := range p.kinds {
				mem, note, st, silent, err := readWith(m, fast && k == 0)
				if err == errCancelled && ctx.Err() != nil {
					return nil, errCancelled
				}
				if silent {
					if k == 0 {
						break // nessuna risposta al saluto del gruppo
					}
					continue
				}
				if err != nil {
					answer(m.Name, err)
					_ = sleepCtx(ctx, identPause)
					continue
				}
				// letta. I modelli del gruppo con lo stesso saluto possono avere
				// memorie diverse (driver o dimensioni): si rilegge la radio una
				// volta per ogni tipo di memoria del gruppo e il modello è quello
				// che riconosce la propria (come CHIRP con i file). Se nessuno la
				// riconosce, i modelli non si distinguono via cavo: si propongono
				// tutti quelli con la stessa memoria di chi ha letto.
				type shapeRead struct {
					m    *genModel
					mem  []byte
					note string
					st   *probeState
				}
				reads := []shapeRead{{m, mem, note, st}}
				seen := map[string]bool{identShape(m): true}
				// 4.32: i tipi di memoria dei modelli che lo dichiarano (identLazy)
				// non si rileggono qui: una rilettura costa una pausa e una lettura
				// intera e per loro non direbbe niente di più (non si riconoscono
				// dalla memoria e il loro driver non controlla il modello). Si
				// propongono lo stesso, anche oltre identMaxRereads, e la memoria
				// si legge con il loro driver solo se l'utente apre quella scheda
				// (apiIdentifyOpen)
				lazyShape := map[string]bool{}
				for _, o := range p.kinds {
					if _, ok := lazyShape[identShape(o)]; !ok {
						lazyShape[identShape(o)] = true
					}
					if !identIsLazy(o) {
						lazyShape[identShape(o)] = false
					}
				}
				for _, o := range p.kinds[k+1:] {
					if seen[identShape(o)] || len(reads) > identMaxRereads || lazyShape[identShape(o)] {
						continue
					}
					seen[identShape(o)] = true
					_ = sleepCtx(ctx, identPause)
					mem2, note2, st2, silent2, err2 := readWith(o, false)
					if err2 == errCancelled && ctx.Err() != nil {
						return nil, errCancelled
					}
					if silent2 || err2 != nil {
						j.logf("%s: %v", o.Name, err2)
						continue
					}
					reads = append(reads, shapeRead{o, mem2, note2, st2})
				}
				// confermati: il modello riconosce la memoria letta con il suo driver;
				// possibili: modelli che non hanno segni riconoscibili nella memoria
				// (come in CHIRP) letti senza errori con il loro driver
				var best *genModel
				var confirmed, possible []*genModel
				for _, r := range reads {
					for _, o := range p.kinds {
						if identShape(o) != identShape(r.m) {
							continue
						}
						if !identCanMatch(o) {
							possible = append(possible, o)
							continue
						}
						if func() (ok bool) {
							defer func() { _ = recover() }()
							return len(r.mem) >= genMinSize(o) && o.Fam.Match(o, r.mem)
						}() {
							confirmed = append(confirmed, o)
							if best == nil || o == m {
								best, mem, note, st = o, r.mem, r.note, r.st
							}
						}
					}
				}
				if len(confirmed) == 0 {
					// nessun modello riconosce la memoria: sono possibili tutti quelli
					// che l'hanno letta senza errori
					for _, r := range reads {
						for _, o := range p.kinds {
							if identShape(o) == identShape(r.m) && identCanMatch(o) {
								possible = append(possible, o)
							}
						}
					}
				}
				if best == nil && len(possible) > 0 {
					best = possible[0]
					for _, o := range possible {
						if o == m {
							best = o
						}
					}
					for _, r := range reads {
						if identShape(r.m) == identShape(best) {
							mem, note, st = r.mem, r.note, r.st
						}
					}
				}
				if best == nil {
					// nessun modello riconosce la memoria: quello che l'ha letta
					best = m
					mem, note, st = reads[0].mem, reads[0].note, reads[0].st
				}
				// 4.6: versione della radio adatta alla memoria letta (BJ-9000 / RT-950
				// Pro con 960 o 990 canali); se i dati la dicono, l'altra non si propone
				if genLayoutOf(best) != 0 {
					b2, why := genPickLayout(best, mem)
					if l, _ := rt950Layout(mem); l != 0 {
						var keep []*genModel
						for _, o := range possible {
							if genLayoutOf(o) == 0 || genLayoutOf(o) == l {
								keep = append(keep, o)
							}
						}
						possible = keep
					}
					best = b2
					note = strings.TrimSpace(note + " " + why)
				}
				// memoria letta per ogni modello proposto (quella del suo tipo)
				mems := map[string][]byte{best.ID: mem}
				var cands []*genModel
				for _, o := range append(confirmed, possible...) {
					for _, r := range reads {
						if _, ok := mems[o.ID]; !ok && identShape(r.m) == identShape(o) {
							mems[o.ID] = r.mem
							cands = append(cands, o)
						}
					}
				}
				// 4.32: modelli dei tipi di memoria non riletti (identLazy):
				// proposti in coda, da leggere all'apertura della scheda
				lazy := map[string]bool{}
				for _, o := range p.kinds {
					if _, ok := mems[o.ID]; !ok && !seen[identShape(o)] && lazyShape[identShape(o)] {
						lazy[o.ID] = true
						cands = append(cands, o)
					}
				}
				genIdentFill(best, cands, mem, note, st, req.Port, res)
				res.Ambiguous = len(confirmed) == 0 && len(res.Others) > 0
				res.Seconds = int(time.Since(start).Seconds())
				if bp, err := a.saveGenBackup(best, mem, nil, "identificata"); err == nil {
					res.Backup = bp
				}
				kinds := []string{best.ID}
				for _, o := range res.Others {
					kinds = append(kinds, o.Kind)
				}
				j.sess = &identSession{kind: best.ID, kinds: kinds, sess: &radioSession{Base: mem}, mems: mems,
					lazy: lazy, port: req.Port, at: time.Now()}
				a.rememberRadio(best.ID)
				return res, nil
			}
		}
	}
	res.Seconds = int(time.Since(start).Seconds())
	// 4.7: la radio trasmetteva da sola (es. in modalità aggiornamento firmware)
	res.Spont = identSpontSeen(probeKey(req.Port))
	return res, nil
}

// identLazyRead (4.32): memoria della radio identificata per il modello scelto.
// Per i modelli proposti senza averli riletti (identSession.lazy) la radio si
// legge adesso con il loro driver; la lettura resta nella sessione.
func (a *App) identLazyRead(ctx context.Context, id *identSession, m *genModel) ([]byte, error) {
	id.mu.Lock()
	defer id.mu.Unlock()
	if b, ok := id.mems[m.ID]; ok {
		return b, nil
	}
	if !id.lazy[m.ID] {
		return id.sess.Base, nil
	}
	a.yieldAuto()
	l := identPortLock(id.port)
	l.Lock()
	defer l.Unlock()
	// la radio deve essere uscita dalla programmazione dell'ultima lettura
	if d := identPause - time.Since(id.at); d > 0 {
		if err := sleepCtx(ctx, d); err != nil {
			return nil, err
		}
	}
	mem, _, err := m.Fam.Download(ctx, m, id.port, func(int, int, string) {})
	id.at = time.Now()
	if err == nil && len(mem) < genMinSize(m) {
		err = fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
	}
	if err != nil {
		return nil, err
	}
	if id.mems == nil {
		id.mems = map[string][]byte{}
	}
	id.mems[m.ID] = mem
	return mem, nil
}

func famName(m *genModel) string { return fmt.Sprintf("%T", m.Fam) }

// identShape: tipo di memoria letta (driver e dimensione)
func identShape(m *genModel) string {
	n := genMinSize(m)
	switch o := m.Opt.(type) {
	case *h777Opt:
		n = o.memsize
	case *rt22Opt:
		n = o.memsize
	case *uv17Opt:
		n = o.total
	case *ga510Opt:
		n = o.memsize
	case *rt900Opt:
		n = o.memsize
	case *bft8Opt:
		if o.spec != nil {
			n = o.spec.memsize // 4.9: la RB27V (5 canali) ha la memoria della RB27
		}
	}
	return fmt.Sprintf("%s/%d", famName(m), n)
}

// identIsLazy (4.32): modello che «Identifica» propone senza rileggere la radio
// con il suo driver (vedi identRun)
func identIsLazy(m *genModel) bool {
	l, ok := m.Opt.(interface{ identLazy() bool })
	return ok && l.identLazy() && !identCanMatch(m)
}

// identCanMatch: il modello si riconosce dal contenuto della memoria (come
// match_model di CHIRP); per gli altri la memoria non dice di che modello è.
func identCanMatch(m *genModel) bool {
	switch o := m.Opt.(type) {
	case *uv5rOpt:
		return o.match
	case *bfcOpt:
		return o.match || o.fileid5x3
	case *h777Opt:
		return o.match
	}
	switch m.Fam.(type) {
	case rdtyFam, uvb5Fam, bft1Fam, rt22Fam, uv3rFam, tytFam:
		return true
	}
	return false
}

const identMaxRereads = 4

func sleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

// K67 Pro e Bajeton BJ-8300
func (a *App) identFillMain(kind string, file []byte, info interface{}, k6 bool, st *probeState, port string, r *identResult) {
	r.Found, r.Kind = true, kind
	if kind == kindXLC {
		ki, _ := info.(k67Info)
		r.Brand, r.Type = "Baofeng", "Ricetrasmettitore analogico FM bibanda"
		r.Model = "K67 Pro (5RH Pro)"
		if ki.K6 {
			r.Model = "BF-K6 (memoria della K67 Pro)"
		}
		r.Name = "Baofeng " + r.Model
		r.add("Marca e modello", "%s", r.Name)
		r.add("Tipo", "%s", r.Type)
		r.add("Modello comunicato dalla radio", "%s", ki.Model)
		r.add("Gamme di frequenza", "%s", bandsText(xlcBands(file, 0x10)))
		r.add("Codice della banda (programma originale)", "%d", ki.FreqBand)
		r.add("Memorie", "%d canali", xlcChannels)
		r.add("Collegamento", "cavo di programmazione sulla porta %s, %d baud", port, ki.Baud)
	} else {
		bi, _ := info.(btInfo)
		r.Brand, r.Type = "Bajeton", "Ricetrasmettitore analogico FM bibanda"
		r.Model = "BJ-8300"
		r.Name = "Bajeton BJ-8300"
		r.add("Marca e modello", "%s", r.Name)
		r.add("Tipo", "%s", r.Type)
		f := strings.Fields(bi.ID)
		if len(f) > 0 {
			r.add("Modello comunicato dalla radio", "%s", f[0])
		}
		if len(f) > 1 {
			r.add("Versione del firmware", "%s", strings.Join(f[1:], " "))
		}
		r.add("Memorie", "%d canali (%d zone da %d)", datZones*datPerZone, datZones, datPerZone)
		r.add("Collegamento", "cavo di programmazione sulla porta %s, 57600 baud", port)
		if v := datVer(file); v != "" {
			r.add("Formato del file", "BT-11_CPS v%s", v)
		}
	}
	r.add("Numero di serie", "%s", "non disponibile: il protocollo di programmazione di questa radio non lo comunica")
	if _, chs, err := readRadio(file, &Report{}); err == nil {
		r.Count = len(packed(chs))
	}
}

// ---------------------------------------------------------------- API

func (a *App) apiIdentifyStart(r *http.Request) (interface{}, error) {
	var req identReq
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Port) == "" {
		return nil, errors.New("scegli la porta del cavo di programmazione (o «USB» per le radio DMR collegate via USB)")
	}
	a.yieldAuto()
	j, err := a.startIdentify(&req, false)
	if err != nil {
		return nil, err
	}
	return j.snapshot(), nil
}

// startIdentify: avvia la ricerca della radio (dalla scheda «Identifica» o, con
// auto, dall'identificazione automatica al collegamento della radio)
func (a *App) startIdentify(req *identReq, auto bool) (*cableJob, error) {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ctx, cancel := context.WithCancel(context.Background())
	j := &cableJob{ID: hex.EncodeToString(b), Op: "identify", State: "running", cancel: cancel, started: time.Now(), Notes: []string{}, Log: []string{}, Auto: auto}
	ctx = context.WithValue(ctx, identAnswerKey{}, func() { j.answered.Store(true) })
	a.mu.Lock()
	if a.job != nil && a.job.isRunning() {
		a.mu.Unlock()
		cancel()
		return nil, errors.New("è già in corso un'operazione con la radio")
	}
	if !auto {
		// 4.3: le ricerche automatiche (una per porta, anche insieme) non sono
		// «l'operazione in corso» della finestra del cavo
		a.job = j
		a.ident = nil
	}
	a.mu.Unlock()
	if auto {
		a.auto.liveAdd(j, true)
	}
	if !auto && !strings.EqualFold(req.Port, "USB") {
		st := a.store.Get()
		st.GenPort = req.Port
		_ = a.store.SaveSettings(st)
	}
	go func() {
		defer cancel()
		if auto {
			defer a.auto.liveAdd(j, false)
		}
		var res *identResult
		var err error
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("errore interno: %v", e)
				}
			}()
			res, err = a.runIdentify(ctx, j, req)
		}()
		if err == nil && !auto && j.sess != nil {
			a.mu.Lock() // prima di j.mu (stesso ordine di autoStep)
			a.ident = j.sess
			a.mu.Unlock()
		}
		j.mu.Lock()
		defer j.mu.Unlock()
		switch {
		case err == errCancelled:
			j.State, j.Error = "cancelled", "Ricerca interrotta. Se la radio è rimasta in modalità programmazione, spegnila e riaccendila."
		case err != nil:
			j.State, j.Error = "error", err.Error()
		default:
			j.State, j.Result = "done", res
		}
		j.CanStop = false
	}()
	return j, nil
}

// identify/open: la radio identificata nella sua scheda (come dopo «Leggi dalla radio»)
func (a *App) apiIdentifyOpen(r *http.Request) (interface{}, error) {
	var req struct {
		Kind string `json:"kind"`
		Seq  int    `json:"seq"` // esito dell'identificazione automatica
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	a.mu.Lock()
	id := a.ident
	a.mu.Unlock()
	if req.Seq > 0 {
		id = a.autoIdentSession(req.Seq)
	}
	if id == nil {
		return nil, errors.New("nessuna radio identificata da aprire")
	}
	kind := id.kind
	// 4.6: «kind#scheda»: radio scelta tra quelle con la stessa gestione
	reqKind, card := splitKindCard(req.Kind)
	if reqKind != "" {
		ok := false
		for _, k := range id.kinds {
			ok = ok || k == reqKind
		}
		if !ok {
			return nil, errors.New("modello non compatibile con la radio identificata")
		}
		kind = reqKind
	}
	if card == "" {
		if c, sure := a.kindCard(kind); sure {
			card = c.ID
		}
	}
	if m, ok := genByID[kind]; ok {
		// 4.32: modello proposto senza averlo riletto: la radio si legge adesso,
		// con il suo driver (una volta sola)
		mem, err := a.identLazyRead(r.Context(), id, m)
		if err != nil {
			return nil, err
		}
		if len(mem) < genMinSize(m) {
			return nil, errors.New("la memoria letta non è quella di questo modello")
		}
		a.rememberCard(kind, card)
		a.mu.Lock()
		a.radios[kind] = &radioSession{Base: mem}
		a.mu.Unlock()
		return map[string]interface{}{"radio": genResp{Kind: m.ID, Channels: genChans(m, mem), Def: m.def(), Warnings: []string{}, Single: genSingle(m.ID, mem)}, "card": card}, nil
	}
	a.rememberCard(kind, card)
	rep := &Report{}
	kind, chs, err := readRadio(id.sess.Base, rep)
	if err != nil || kind != id.kind {
		return nil, fmt.Errorf("dati letti non validi: %v", err)
	}
	a.mu.Lock()
	a.radios[id.kind] = &radioSession{Base: id.sess.Base, K6: id.sess.K6}
	a.mu.Unlock()
	resp := radioResp{Kind: kind, Channels: chs, Warnings: rep.Warnings}
	if kind == kindXLC {
		resp.Bands = xlcBands(id.sess.Base, 0x10)
	} else {
		resp.DatVer, resp.Single = datVer(id.sess.Base), datSingleList(id.sess.Base)
	}
	return map[string]interface{}{"radio": resp, "card": card}, nil
}
