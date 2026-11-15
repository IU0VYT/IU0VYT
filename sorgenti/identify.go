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
	"sync/atomic"
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
	identPortMu    sync.Mutex                // solo per i test: prove in sequenza
)

// ---------------------------------------------------------------- radio finta muta

// silentPort: radio che non risponde mai; registra i byte inviati dal driver.
//
// 4.35: la trascrizione finisce per CONTEGGIO e non più per tempo. Prima ogni
// driver veniva fermato dopo 1,5 secondi di orologio: i driver che ripetono il
// saluto con delle pause, o che aspettano la radio fino a una scadenza, venivano
// tagliati in un punto diverso a ogni avvio e secondo il carico del PC (478
// driver insieme), e modelli con lo stesso identico saluto finivano in gruppi di
// ricerca diversi (famiglia Retevis RT22 / Baofeng T20: da 4 a 6 gruppi invece
// di uno; Kenwood TK-x80). Su un PC lento il piano di ricerca poteva cambiare
// ancora di più. Ora la trascrizione si chiude dopo identTxMax byte inviati,
// dopo identTxReads letture rimaste senza risposta o quando il driver rinuncia
// da solo: lo stesso risultato su ogni PC.
type silentPort struct {
	st *silentState
}

// silentState: trascrizione di un modello (resta la stessa se il driver riapre
// la porta)
type silentState struct {
	mu     sync.Mutex
	tx     bytes.Buffer
	reads  int
	done   bool
	slept  time.Duration // pause chieste dal driver (tempo «virtuale»)
	cancel func()        // ferma il driver quando la trascrizione è completa
}

// finish (con st.mu preso): trascrizione completa. Il driver viene fermato: molti
// aspettano la radio rileggendo la porta fino a una scadenza e, con una porta che
// risponde subito «niente», terrebbero occupato il processore per secondi
func (st *silentState) finish() {
	st.done = true
	if st.cancel != nil {
		st.cancel()
	}
}

var errIdentEnough = errors.New("trascrizione completa")

// identTxReads: letture senza risposta dopo le quali la trascrizione è completa.
// Un driver che aspetta la radio fino a una scadenza (ciclo di letture a tempo)
// le supera in pochi millisecondi; i driver che contano i tentativi non ci
// arrivano mai
const identTxReads = 400

// identTxSleep: pause sommate dopo le quali la trascrizione è completa (un driver
// che fra un saluto e l'altro aspetta a lungo, come quello della UV-5R, resta
// nel gruppo del suo primo saluto: è il comportamento di sempre)
const identTxSleep = 1500 * time.Millisecond

func (p *silentPort) Write(b []byte) error {
	st := p.st
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.done {
		return errIdentEnough
	}
	st.tx.Write(b)
	if st.tx.Len() >= identTxMax {
		st.finish()
		return errIdentEnough
	}
	return nil
}
func (p *silentPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if ctx.Err() != nil {
		return errCancelled
	}
	st := p.st
	st.mu.Lock()
	defer st.mu.Unlock()
	st.reads++
	if st.done || st.reads > identTxReads {
		st.finish()
		return errIdentEnough
	}
	return errTimeout
}
func (p *silentPort) Purge() error { return nil }
func (p *silentPort) SetBaud(baud int) error {
	st := p.st
	st.mu.Lock()
	defer st.mu.Unlock()
	if !st.done {
		fmt.Fprintf(&st.tx, "\x00<baud %d>\x00", baud)
	}
	return nil
}
func (p *silentPort) Close() error { return nil }

// identSilentPort: la porta è quella di una trascrizione
func identSilentPort(p Port) bool {
	_, ok := p.(*silentPort)
	return ok
}

// drvSleep: pausa di un driver, interrotta se l'operazione viene annullata.
// Durante la trascrizione del saluto (identTranscripts) non c'è nessuna radio e
// le pause non si aspettano: così la trascrizione non dipende dal tempo
func drvSleep(ctx context.Context, p Port, d time.Duration) error {
	if sp, ok := p.(*silentPort); ok {
		// le pause si sommano senza aspettare: oltre identTxSleep la trascrizione
		// è completa (come quando si fermava a tempo, ma senza dipendere dal PC)
		st := sp.st
		st.mu.Lock()
		st.slept += d
		if st.slept >= identTxSleep && !st.done {
			st.finish()
		}
		st.mu.Unlock()
		if ctx.Err() != nil {
			return errCancelled
		}
		return nil
	}
	select {
	case <-ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

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
		sts := map[string]*silentState{}
		var list []*genModel
		for _, m := range genModels {
			if m.Conn == "" {
				list = append(list, m)
				sts[m.ID] = &silentState{}
			}
		}
		probeReg.Lock()
		probeReg.silent = func(name string, baud int) (Port, error) {
			st := sts[strings.TrimPrefix(name, identSilentName)]
			if st == nil {
				return nil, errors.New("porta sconosciuta")
			}
			st.mu.Lock()
			defer st.mu.Unlock()
			if st.done {
				return nil, errIdentEnough
			}
			fmt.Fprintf(&st.tx, "<%d>", baud)
			return &silentPort{st: st}, nil
		}
		probeReg.Unlock()
		late := int32(0)
		var wg sync.WaitGroup
		for _, m := range list {
			wg.Add(1)
			go func(m *genModel) {
				defer wg.Done()
				// la scadenza è solo una rete di sicurezza: la trascrizione finisce
				// per conteggio (vedi silentPort)
				ctx, cancel := context.WithTimeout(context.Background(), identTxTimeout)
				defer cancel()
				st := sts[m.ID]
				st.mu.Lock()
				st.cancel = cancel
				st.mu.Unlock()
				func() {
					defer func() { _ = recover() }()
					_, _, _ = m.Fam.Download(ctx, m, identSilentName+m.ID, func(int, int, string) {})
				}()
				st.mu.Lock()
				if ctx.Err() != nil && !st.done {
					atomic.AddInt32(&late, 1) // fermata dal tempo: chiave non garantita
				}
				st.done = true
				st.mu.Unlock()
			}(m)
		}
		wg.Wait()
		probeReg.Lock()
		probeReg.silent = nil
		probeReg.Unlock()
		for _, m := range list {
			st := sts[m.ID]
			st.mu.Lock()
			b := st.tx.Bytes()
			if len(b) > identTxMax+32 {
				b = b[:identTxMax+32]
			}
			keys[m.ID] = string(b)
			st.mu.Unlock()
		}
		identTxKeys = keys
		identTxLate = int(atomic.LoadInt32(&late))
	})
	return identTxKeys
}

// identWarmUp (4.36): i saluti dei driver si preparano poco dopo l'avvio del
// programma, in sottofondo. Prima si preparavano alla prima ricerca, che per
// questo partiva in ritardo (da mezzo secondo a qualche secondo, secondo il PC)
func identWarmUp() {
	go func() {
		time.Sleep(1500 * time.Millisecond)
		identTranscripts()
	}()
}

// identTxTimeout: tempo massimo di una trascrizione (rete di sicurezza);
// identTxLate: trascrizioni fermate dal tempo invece che dal conteggio (deve
// restare zero: lo controlla una prova)
var (
	identTxTimeout = 5 * time.Second
	identTxLate    int
)

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
	if x, ok := p.Port.(interface{ SetEvenParity(bool) error }); ok {
		return x.SetEvenParity(on)
	}
	return nil
}

func (p probePort) SetTwoStopBits(on bool) error {
	if x, ok := p.Port.(interface{ SetTwoStopBits(bool) error }); ok {
		return x.SetTwoStopBits(on)
	}
	return nil
}

// SetRTS (4.9): RTS spento per la Retevis C2
func (p probePort) SetRTS(on bool) error {
	if x, ok := p.Port.(interface{ SetRTS(bool) error }); ok {
		return x.SetRTS(on)
	}
	return nil
}

// SetDTR (4.13): DTR spento per il firmware IJV V6
func (p probePort) SetDTR(on bool) error {
	if x, ok := p.Port.(interface{ SetDTR(bool) error }); ok {
		return x.SetDTR(on)
	}
	return nil
}

func (p probePort) unwrapPort() Port { return p.Port }

type probeState struct {
	mu      sync.Mutex
	key     string // porta (vedi probeKey)
	tx      []byte // byte inviati in questo tentativo, senza i disturbi di linea (vedi cleanBytes)
	prev    []byte // ultimi byte inviati sulla porta prima di questo tentativo (eco ritardata)
	genuine bool   // la radio ha risposto (byte «puliti» diversi dall'eco di quelli inviati)
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
	// known (4.36): tentativo di una radio già riconosciuta su questa porta: non
	// si ascolta la porta prima di cominciare (vedi openPort)
	known bool
	// diagnosi (4.36, vedi identdiag.go): righe scritte per questo tentativo;
	// secret: tentativo della K67 Pro (la password non va nella diagnosi)
	secret                bool
	dTx, dRx, dTo, dPlain int
	dbuf                  []string
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

// identTxEcho: byte (senza disturbi) ricordati per ogni tentativo
const identTxEcho = 1 << 16

func identSentAdd(key string, b []byte) {
	identSent.Lock()
	defer identSent.Unlock()
	if identSent.m == nil {
		identSent.m = map[string][]byte{}
	}
	h := append(identSent.m[key], cleanBytes(b)...) // 4.35: senza i disturbi di linea
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

// echoOnly (4.35): tutto quello che è arrivato nel tentativo è l'eco, anche con
// dei byte persi, di quello che è stato inviato: nessuna radio ha risposto. Vale
// solo con abbastanza byte diversi fra loro (una radio vuota che risponde solo
// con delle conferme 0x06, uguali a quelle inviate dal programma, non è un'eco)
func (s *probeState) echoOnly() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sure || len(s.clean) < 8 {
		return false
	}
	seen := map[byte]bool{}
	for _, c := range s.clean {
		seen[c] = true
	}
	return len(seen) >= 4 && isSubsequence(s.clean, append(append([]byte{}, s.prev...), s.tx...))
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

func (s *probeState) answered() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.genuine
}

func (p probePort) Write(b []byte) error {
	p.st.mu.Lock()
	// 4.35: si ricordano TUTTI i byte inviati nel tentativo, senza i disturbi di
	// linea (prima i primi 4096 byte grezzi: un driver che manda pacchetti lunghi,
	// pieni di zeri, usciva dal conto e la sua eco sembrava la risposta di una radio)
	if len(p.st.tx) < identTxEcho {
		p.st.tx = append(p.st.tx, cleanBytes(b)...)
	}
	key := p.st.key
	p.st.mu.Unlock()
	identSentAdd(key, b)
	p.st.diagTx(b)
	return p.Port.Write(b)
}

// SetBaud (4.36): il cambio di velocità va nella diagnosi
func (p probePort) SetBaud(baud int) error {
	p.st.diagNote("    velocità: %d baud", baud)
	return p.Port.SetBaud(baud)
}

func (p probePort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if !p.st.answered() && timeout > identFirstWait {
		timeout = identFirstWait
	}
	t0 := time.Now()
	err := p.Port.ReadFull(ctx, b, timeout)
	if err == nil {
		p.st.diagRx(b, time.Since(t0))
	} else if err != errCancelled {
		p.st.diagWait(len(b), time.Since(t0), err)
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
			if !bytes.Contains(s.tx, s.clean) {
				s.genuine = true // la radio ha risposto: il tentativo prosegue
				echo := append(identSentGet(s.key), s.tx...)
				// eco con dei byte persi (es. letture scartate dai driver): i byte
				// ricevuti compaiono, nello stesso ordine, tra quelli inviati in
				// questo tentativo o alla fine del precedente (s.prev: presi
				// all'inizio del tentativo)
				recent := append(append([]byte{}, s.prev...), s.tx...)
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
		if st != nil {
			identDiag.add(st.key, "    apertura a %d baud non riuscita: %v", baud, err)
		}
		return p, err
	}
	// 4.7: radio che trasmette da sola (es. in modalità aggiornamento firmware, con
	// i suoi messaggi periodici): prima del tentativo si ascolta la porta senza
	// inviare nulla; quei byte non contano come risposta della radio, così ogni
	// tentativo finisce presto come con una porta senza radio (prima ogni modello
	// sembrava rispondere e la ricerca durava alcuni minuti)
	st.mu.Lock()
	st.listenFrom = time.Now()
	st.mu.Unlock()
	sp, note := identSpontSample(st.ctx, probeKey(name), baud, p)
	st.mu.Lock()
	st.spont = sp
	st.listened += time.Since(st.listenFrom) // il tempo d'ascolto non conta come silenzio della radio
	st.listenFrom = time.Time{}
	st.mu.Unlock()
	if note != "" {
		identDiag.add(st.key, "    apertura a %d baud%s", baud, note)
	} else {
		st.diagNote("    apertura a %d baud", baud)
	}
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

// identSpontSample: ascolto della porta appena aperta; i byte che non sono
// disturbi della linea si ricordano (validi identSpontTTL). Restituisce anche
// una nota per la diagnosi.
//
// L'ascolto dura identSpontListen (un secondo) e si fa una volta al minuto per
// ogni porta e velocità, PRIMA di inviare qualcosa. Oltre a scoprire le radio
// che trasmettono da sole, è anche il silenzio che la radio trova fra
// l'apertura della porta e il primo saluto.
//
// 4.38: la 4.36 aveva tolto questo ascolto per la radio già riconosciuta sulla
// porta e, dopo un ascolto senza byte, per tutte le altre velocità. Con la
// Bajeton BJ-9000 vera (prova del 14/11/2026) la radio ogni tanto si bloccava,
// soprattutto alla prima identificazione; con la 4.35, che ascoltava sempre,
// non succedeva. L'ascolto è tornato quello della 4.35: è un'attesa verso la
// radio e, come le altre, non si accorcia senza una misura su radio vere.
func identSpontSample(ctx context.Context, key string, baud int, p Port) ([]byte, string) {
	k := fmt.Sprintf("%s|%d", key, baud)
	identSpont.Lock()
	if r, ok := identSpont.m[k]; ok && time.Since(r.at) < identSpontTTL {
		identSpont.Unlock()
		if len(r.data) > 0 {
			return r.data, diagNoteSpont(r.data)
		}
		return r.data, ""
	}
	identSpont.Unlock()
	var got []byte
	buf := make([]byte, 1)
	listen := identSpontListen
	if baud == 9600 {
		listen = identSpontListen9600
	}
	t0 := time.Now()
	end := t0.Add(listen)
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
		return nil, diagNoteCut // ricerca interrotta: ascolto incompleto, non si ricorda
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
	return c, diagNoteListen(time.Since(t0), got)
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
// note (al massimo identKnownMax): le ricerche successive cominciano da lì.
//
// 4.36: si ricorda anche SU QUALE PORTA la radio è stata riconosciuta
// (Settings.PortRadios, al massimo identPortMax modelli per porta). Chi usa il
// programma ha poche radio e le collega sempre agli stessi cavi: sulla porta di
// un cavo si provano per prime le radio già riconosciute lì (l'ultima per
// prima), poi le altre radio note, poi tutto il resto. Prima l'elenco era uno
// solo: con tre radio su due cavi, la radio collegata poteva essere provata per
// terza, dopo due tentativi a vuoto su radio che su quel cavo non ci sono mai
// state.
const (
	identKnownMax = 24
	identPortMax  = 6
	identPortsMax = 48 // porte ricordate
)

var rememberMu sync.Mutex // più porte possono riconoscere una radio nello stesso momento

// moveFirst: elenco con v in testa, senza ripetizioni, al massimo n voci
func moveFirst(list []string, v string, n int) []string {
	out := []string{v}
	for _, k := range list {
		if k != v && len(out) < n {
			out = append(out, k)
		}
	}
	return out
}

func (a *App) rememberRadio(port, kind string) {
	rememberMu.Lock()
	defer rememberMu.Unlock()
	st := a.store.Get()
	key := probeKey(port)
	list := moveFirst(st.KnownRadios, kind, identKnownMax)
	plist := moveFirst(st.PortRadios[key], kind, identPortMax)
	if strings.Join(list, ",") == strings.Join(st.KnownRadios, ",") && strings.Join(plist, ",") == strings.Join(st.PortRadios[key], ",") {
		return
	}
	// la mappa delle impostazioni è condivisa con le altre copie: se ne fa una nuova
	pm := map[string][]string{}
	for k, v := range st.PortRadios {
		pm[k] = v
	}
	pm[key] = plist
	if len(pm) > identPortsMax {
		// troppe porte ricordate (adattatori spostati da una presa all'altra):
		// si tolgono quelle che ora non ci sono
		present := map[string]bool{key: true}
		for _, p := range listPortsFunc() {
			present[probeKey(p.Name)] = true
		}
		for k := range pm {
			if !present[k] && len(pm) > identPortsMax {
				delete(pm, k)
			}
		}
	}
	st.KnownRadios, st.PortRadios = list, pm
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
//
// 4.36: opts: tentativo di una radio già nota sulla porta (senza ascolto
// iniziale) e tentativo con dati da non scrivere nella diagnosi.
type probeOpt struct {
	known, secret bool
}

func probe(ctx context.Context, port string, fn func(ctx context.Context) error, opts ...probeOpt) (st *probeState, silent bool, err error) {
	key := probeKey(port)
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()
	st = &probeState{key: key, ctx: pctx, prev: identSentTail(key, identEchoTail)}
	for _, o := range opts {
		st.known, st.secret = o.known, o.secret
	}
	t0 := time.Now()
	defer func() { st.diagEnd(t0, diagOutcome(err, silent)) }()
	st.onAnswer, _ = ctx.Value(identAnswerKey{}).(func())
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
	go func() {
		start := time.Now()
		wait := silentMax
		for {
			select {
			case <-time.After(wait):
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
	// Already (4.36, ricerca automatica): la scheda di questa radio è già aperta
	// e la memoria non è stata riletta (vedi k67Check): non c'è niente da aprire
	Already bool `json:"already,omitempty"`
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
	if mem != nil {
		r.add("Firmware / identificativo", "%s", note)
		dmrInfo(m, mem, r)
	}
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
	if mem == nil {
		return // 4.40: controllo breve, memoria non letta
	}
	r.Count = len(packed(genChans(m, mem)))
}

// ---------------------------------------------------------------- scansione

type identReq struct {
	Port   string `json:"port"`   // porta COM o "USB"
	Brand  string `json:"brand"`  // "" tutte, oppure la marca (Baofeng, Bajeton, Quansheng, Retevis…)
	Family string `json:"family"` // "" tutte, analog, dmr
	// KnownOnly (4.36, solo ricerca automatica): si provano soltanto le radio
	// già riconosciute (vedi autoRetryKnown)
	KnownOnly bool `json:"-"`
}

type identProbe struct {
	name  string // testo mostrato
	kinds []*genModel
	k67   bool
	baud  int // K67: velocità
	bt    bool
	// known (4.36): radio già riconosciuta, provata senza ascolto iniziale
	known bool
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
			// 4.35: lo stesso, deciso dal modello (vedi kwlOpt.identGroup)
			if o, ok := m.Opt.(interface{ identGroup() string }); ok {
				key += "|" + o.identGroup()
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
	// una scheda aperta, poi le altre nell'ordine solito (4.36: vedi rememberRadio)
	known := map[string]int{}
	for i, k := range a.store.Get().KnownRadios {
		if _, ok := known[k]; !ok {
			known[k] = i
		}
	}
	// 4.36: prima di tutte le radio già riconosciute su QUESTA porta
	onPort := map[string]int{}
	for i, k := range a.store.Get().PortRadios[probeKey(req.Port)] {
		if _, ok := onPort[k]; !ok {
			onPort[k] = i
		}
	}
	rank := func(kind string) int {
		if i, ok := onPort[kind]; ok {
			return i
		}
		if i, ok := known[kind]; ok {
			return 100 + i
		}
		if open[kind] {
			return 1000
		}
		if identPopular[kind] {
			return 1500
		}
		// 4.40: famiglie che lo dichiarano (Alinco) in coda al giro: per le altre
		// radio la ricerca resta quella di prima, con gli stessi tempi
		if m := genByID[kind]; m != nil {
			if _, ok := m.Fam.(interface{ identLast() }); ok {
				return 2500
			}
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
		r := 9999
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
	// 4.36: tentativi delle radio già riconosciute (senza ascolto iniziale); nella
	// ricerca automatica si ripetono in mezzo agli altri
	idx, every := a.knownProbes(plan, req.Port)
	for _, i := range idx {
		plan[i].known = true
	}
	if j.Auto {
		if req.KnownOnly {
			var kn []identProbe
			for _, i := range idx {
				kn = append(kn, plan[i])
			}
			plan = kn
		} else {
			plan = identInterleaveKnown(plan, idx, every)
		}
	}
	var first []string
	for _, p := range plan[:min(len(plan), 6)] {
		first = append(first, p.name)
	}
	identDiag.add(probeKey(req.Port), "=== ricerca %s: %d tentativi; primi: %s", diagMode(j.Auto, req.KnownOnly), len(plan), strings.Join(first, " | "))
	res.Tried = 0
	res, err := a.identRun(ctx, j, req, plan, start, res)
	switch {
	case err != nil:
		identDiag.add(probeKey(req.Port), "=== fine dopo %d ms: %v", time.Since(start).Milliseconds(), err)
	case res != nil && res.Found:
		identDiag.add(probeKey(req.Port), "=== fine dopo %d ms: riconosciuta %s (%s) al tentativo %d", time.Since(start).Milliseconds(), res.Name, res.Kind, res.Tried)
	default:
		identDiag.add(probeKey(req.Port), "=== fine dopo %d ms: nessuna radio riconosciuta (%d tentativi, %d risposte non andate a buon fine)",
			time.Since(start).Milliseconds(), res.Tried, len(res.Answers))
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
// queste, accesa dopo il cavo, si trova in pochi secondi anche a metà giro.
//
// 4.36: se su QUELLA porta sono già state riconosciute delle radio si
// riprovano solo quelle (di solito una o due, al massimo identKnownPort) e più
// spesso, ogni identKnownEveryPort tentativi: la radio accesa dopo il cavo
// aspetta al massimo quattro secondi circa invece di otto. Su una porta nuova resta la
// regola di prima con le radio note in generale (al massimo identKnownAny).
const (
	identKnownEvery     = 8
	identKnownAny       = 4
	identKnownEveryPort = 4
	identKnownPort      = 3
)

// knownProbes: posizioni nel piano dei tentativi delle radio già riconosciute
// (sono i primi del piano, vedi identPlan) e ogni quanti altri tentativi vanno
// ripetuti. Della K67 Pro, che ha due tentativi (due velocità), conta il primo:
// la velocità giusta è quella ricordata.
func (a *App) knownProbes(plan []identProbe, port string) (idx []int, every int) {
	st := a.store.Get()
	pick := func(list []string, limit int) []int {
		set := map[string]bool{}
		for _, k := range list {
			set[k] = true
		}
		var out []int
		k67 := false
		for i, p := range plan {
			hit := (p.k67 && set[kindXLC]) || (p.bt && set[kindDAT])
			for _, m := range p.kinds {
				hit = hit || set[m.ID]
			}
			if !hit {
				break
			}
			if p.k67 {
				if k67 {
					continue
				}
				k67 = true
			}
			if len(out) < limit {
				out = append(out, i)
			}
		}
		return out
	}
	if idx = pick(st.PortRadios[probeKey(port)], identKnownPort); len(idx) > 0 {
		return idx, identKnownEveryPort
	}
	return pick(st.KnownRadios, identKnownAny), identKnownEvery
}

// identInterleaveKnown: i tentativi delle radio note (posizioni idx) all'inizio
// e poi ripetuti ogni every tentativi degli altri
func identInterleaveKnown(plan []identProbe, idx []int, every int) []identProbe {
	if len(idx) == 0 || len(idx) >= len(plan) || every <= 0 {
		return plan
	}
	isKnown := map[int]bool{}
	for _, i := range idx {
		isKnown[i] = true
	}
	var kn, others []identProbe
	for i, p := range plan {
		if isKnown[i] {
			kn = append(kn, p)
		} else {
			others = append(others, p)
		}
	}
	out := append([]identProbe{}, kn...)
	for i, p := range others {
		if i > 0 && i%every == 0 {
			out = append(out, kn...)
		}
		out = append(out, p)
	}
	return out
}

func (a *App) identRun(ctx context.Context, j *cableJob, req *identReq, plan []identProbe, start time.Time, res *identResult) (*identResult, error) {
	answer := func(who string, err error) {
		if err != nil && err != errCancelled {
			res.Answers = append(res.Answers, who+": "+err.Error())
			j.logf("%s ha risposto, ma: %v", who, err)
		}
	}
	for i, p := range plan {
		if ctx.Err() != nil {
			return nil, errCancelled
		}
		j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", i+1, len(plan)), true)
		j.prog(i, len(plan), "Provo: "+p.name)
		res.Tried++
		identDiag.add(probeKey(req.Port), "tentativo %d di %d: %s", i+1, len(plan), p.name)
		switch {
		case p.k67 || p.bt:
			kind := kindXLC
			j.readStart("Baofeng K67 Pro")
			if p.bt {
				j.readStart("Bajeton BJ-8300")
			}
			creq := &cableReq{Op: "read", Radio: kindXLC, Port: req.Port, Baud: p.baud, Password: a.store.Get().K67Pwd}
			if p.bt {
				kind, creq.Radio, creq.Baud, creq.Password = kindDAT, kindDAT, 0, ""
			}
			var file []byte
			var info interface{}
			var k6 bool
			// 4.36: ricerca automatica con la scheda della K67 Pro già aperta:
			// basta il controllo breve (vedi k67Check); l'interfaccia dirà
			// «Radio già identificata!! Scheda Radio già aperta» senza aspettare
			// la lettura di tutta la memoria, che non userebbe
			// 4.40: lo stesso per la Bajeton BJ-8300 (btCheck, vedi identquick.go)
			quick := j.Auto && identQuickAllowed(kind) && a.autoTabOpen(kind)
			st, silent, err := probe(ctx, req.Port, func(pctx context.Context) error {
				var e error
				if quick && p.bt {
					info, e = btCheck(pctx, req.Port, j.prog)
					return e
				}
				if quick {
					var ki k67Info
					ki, e = k67Check(pctx, req.Port, &k67Opts{Password: creq.Password, Baud: creq.Baud}, j.prog)
					info, k6 = ki, ki.K6
					return e
				}
				file, info, k6, e = a.cableRead(pctx, j, creq)
				// 4.36: memoria della K67 Pro senza le informazioni della radio
				// (tutti i byte uguali): non è una lettura buona. Il programma
				// originale in questo caso non scrive; qui la radio non si dà per
				// riconosciuta (può succedere con una radio spenta che risponde)
				if e == nil && kind == kindXLC && len(file) >= 96 && allSame(file[:96]) {
					e = errors.New("le informazioni lette dalla radio non sono valide")
				}
				return e
			}, probeOpt{known: p.known, secret: p.k67})
			if err != nil || silent {
				j.readStop()
			}
			if err == errCancelled && ctx.Err() != nil {
				return nil, errCancelled
			}
			if silent {
				continue
			}
			if err != nil {
				answer(p.name, err)
				if quick && st != nil && st.radioSure() {
					// 4.40: la radio ha risposto ma il controllo breve non è andato a
					// buon fine: da qui in poi, per lei, solo la lettura normale
					identQuickOff(kind)
					identDiag.add(probeKey(req.Port), "  controllo breve non riuscito: per questa radio non si usa più fino al riavvio")
				}
				if j.Auto && p.known && st != nil && st.radioSure() {
					j.knownFail.Store(true) // vedi cableJob.knownFail
				}
				_ = sleepCtx(ctx, identPause)
				continue
			}
			// 4.36: la velocità a cui la K67 Pro ha risposto si ricorda (come fa il
			// programma del costruttore): la prossima ricerca parte da quella
			if ki, ok := info.(k67Info); ok && kind == kindXLC && (ki.Baud == 19200 || ki.Baud == 115200) {
				if s := a.store.Get(); s.K67Baud != ki.Baud {
					s.K67Baud = ki.Baud
					_ = a.store.SaveSettings(s)
				}
			}
			a.identFillMain(kind, file, info, k6, st, req.Port, res)
			res.Seconds = int(time.Since(start).Seconds())
			if quick {
				// nessuna memoria letta: niente copia e niente da aprire
				res.Already = true
				identDiag.add(probeKey(req.Port), "  scheda già aperta: controllo breve, memoria non riletta")
				a.rememberRadio(req.Port, kind)
				return res, nil
			}
			if bp, err := a.saveBackup(kind, file, "identificata"); err == nil {
				res.Backup = bp
			}
			j.sess = &identSession{kind: kind, kinds: []string{kind}, sess: &radioSession{Base: file, K6: k6}}
			a.rememberRadio(req.Port, kind)
			return res, nil
		default:
			readWith := func(m *genModel) (mem []byte, note string, st *probeState, silent bool, err error) {
				j.set(fmt.Sprintf("Ricerca della radio (%d di %d)", i+1, len(plan)), true)
				j.prog(i, len(plan), "Provo: "+m.Name)
				if m != p.kinds[0] {
					identDiag.add(probeKey(req.Port), "  stesso saluto, altro modello: %s", m.Name)
				}
				j.readStart(m.Name)
				st, silent, err = probe(ctx, req.Port, func(pctx context.Context) error {
					var e error
					mem, note, e = m.Fam.Download(pctx, m, req.Port, func(done, total int, msg string) {
						j.mu.Lock()
						j.Phase = "Lettura: " + m.Name
						j.Done, j.Total, j.Msg = done, total, msg
						j.mu.Unlock()
						if done > 0 {
							j.reading.Store(true)
						}
					})
					if e == nil && len(mem) < genMinSize(m) {
						e = fmt.Errorf("dati letti incompleti (%d byte)", len(mem))
					}
					return e
				}, probeOpt{known: p.known})
				if err != nil || silent {
					j.readStop()
				}
				return
			}
			// 4.40: ricerca automatica con la scheda di un modello del gruppo già
			// aperta e una famiglia che ha il controllo breve (identquick.go):
			// basta sapere che sul cavo c'è ancora quella radio
			if qm := a.identQuickModel(j, p.kinds); qm != nil {
				j.prog(i, len(plan), "Provo: "+qm.Name)
				identDiag.add(probeKey(req.Port), "  scheda già aperta: controllo breve di %s", qm.Name)
				st, silent, err := probe(ctx, req.Port, func(pctx context.Context) error {
					return qm.Fam.(genQuicker).quick(pctx, qm, req.Port)
				}, probeOpt{known: p.known})
				if err == errCancelled && ctx.Err() != nil {
					return nil, errCancelled
				}
				if silent {
					continue // nessuna risposta al saluto del gruppo
				}
				if err == nil {
					genIdentFill(qm, nil, nil, "", st, req.Port, res)
					res.Already = true
					res.Seconds = int(time.Since(start).Seconds())
					identDiag.add(probeKey(req.Port), "  controllo breve riuscito: memoria non riletta")
					a.rememberRadio(req.Port, qm.ID)
					return res, nil
				}
				// la radio ha risposto ma non è (più) quella della scheda, o non
				// ha risposto come previsto: per questo modello il controllo breve
				// non si usa più e si prosegue con la lettura normale
				identQuickOff(qm.ID)
				j.logf("%s: controllo breve non riuscito (%v): si passa alla lettura normale", qm.Name, err)
				identDiag.add(probeKey(req.Port), "  controllo breve non riuscito (%v): lettura normale, e per questo modello non si usa più fino al riavvio", err)
				_ = sleepCtx(ctx, identPause)
			}
			for k, m := range p.kinds {
				mem, note, st, silent, err := readWith(m)
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
					if j.Auto && p.known && st != nil && st.radioSure() {
						j.knownFail.Store(true) // vedi cableJob.knownFail
					}
					_ = sleepCtx(ctx, identPause)
					continue
				}
				// 4.35: lettura «riuscita» in cui tutti i byte arrivati erano l'eco,
				// anche con dei byte persi, di quelli inviati (vedi echoOnly).
				// Succede con un cavo che rimanda indietro i dati, senza
				// radio, e un driver che accetta come risposta un pacchetto uguale alla
				// sua richiesta (Radtel RT-5D): non è una radio
				if !strings.EqualFold(req.Port, "USB") && st != nil && st.echoOnly() {
					j.logf("%s: lettura fatta solo dell'eco del cavo, nessuna radio", m.Name)
					identDiag.add(probeKey(req.Port), "  lettura fatta solo dell'eco del cavo: nessuna radio")
					j.readStop()
					if k == 0 {
						break
					}
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
					mem2, note2, st2, silent2, err2 := readWith(o)
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
				a.rememberRadio(req.Port, best.ID)
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
	case alincoFam:
		// 4.40: DR-135T e DR-03T, DR-435T e DR-06T rispondono allo stesso nome: si
		// distinguono dai limiti di banda scritti nella memoria (come CHIRP con i file)
		return alincoO(m).sig != nil
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
		if file != nil {
			r.add("Gamme di frequenza", "%s", bandsText(xlcBands(file, 0x10)))
		}
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
		if file != nil {
			if v := datVer(file); v != "" {
				r.add("Formato del file", "BT-11_CPS v%s", v)
			}
		}
	}
	r.add("Numero di serie", "%s", "non disponibile: il protocollo di programmazione di questa radio non lo comunica")
	if file == nil {
		return // 4.36: controllo breve, memoria non letta
	}
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
