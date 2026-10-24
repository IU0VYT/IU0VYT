// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware con i caricatori "KDH" più recenti, ricavati dal
// codice dei programmi del costruttore (classe KDH_Bootloader.BootHelper):
//
//   kdhx    "Baofeng Bootloader Pro" (UV-25 Plus, file .kdhx)
//   bj9000  "BJ9000_UPDATE_V02" (Bajeton BJ-9000, file .Firmware; protocollo V3)
//   kdh9g   "Bootloader_BaudRate38400" (Baofeng UV-9G, file .kdh)
//
// kdhx e bj9000 usano pacchetti come il caricatore KDH V1.0 (vedi bfkdh.go):
//   kdhx:   AA cmd arg lenH lenL dati crcH crcL EF
//   bj9000: AA cmd argH argL lenH lenL dati crcH crcL 55
// CRC16 XMODEM su tutto tranne AA e la coda. Risposta: stesso comando con
// argomento 06 (accettato) o un codice d'errore (kdhx: nel comando; bj9000:
// nell'argomento): E1 codice d'accesso, E2 errore di controllo (nuovo invio),
// E3 indirizzo, E4 scrittura della flash, E5 comando, E6 (bj9000) modello errato.
// Attesa di ogni risposta 2 secondi, poi nuovo invio, al massimo 5 volte; il
// conteggio riparte quando arriva l'inizio (AA) di una risposta.
//
// Sequenza automatica (radio accesa):
//   kdhx:   9600 baud, "PROGROM"+"H802"+"U" → 06; "UPDATE" → 06; la porta si
//           chiude e si riapre a 115200 baud, pausa di 100 ms
//   bj9000: 115200 baud, "PROGRAMBT9000U" → 06; "UPDATE" → 06; pausa di 80 ms
//   poi 42 (avvio del caricatore) → conferma.
// Sequenza manuale (radio accesa con i tasti indicati dal costruttore, già nel
// caricatore): 115200 baud, si parte da qui:
//   kdhx 01 "BOOTLOADER" / bj9000 0A "BOOTLOADER_V3"   presentazione
//   bj9000: 02 [32 byte del file dall'offset 992]      controllo del modello
//   04 [numero di pacchetti]  (bj9000: 2 byte, numero-1 se più di uno)
//   03 pacchetti da 1024 byte (arg = numero; l'ultimo completato con 00)
//   45 fine (kdhx attende la conferma; bj9000 no)
//
// kdh9g: 9600 baud ".Program" → 06 (primo byte ricevuto); 'U'; la porta si
// riapre a 38400 baud, pausa di 500 ms (in modalità manuale si parte da qui,
// a 38400 baud); AA 4B 44 CC 48 EF → 06; poi blocchi da 38 byte
// AA addrH addrL 32 byte crcH crcL EF (CRC16 XMODEM su indirizzo e dati, ultimo
// blocco completato con 00) → 06; E1/E2/E3 (dati, indirizzo, flash) → nuovo
// invio, al massimo 3 volte; nessuna risposta in 1 secondo → nuovo invio, al
// massimo 5 volte; alla fine (o in caso di errore) un byte 'E' (45).

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type kdh2Variant struct {
	ID        string
	Hello     []byte // presentazione della sequenza automatica
	HelloBaud int
	BootBaud  int
	Reopen    bool          // la porta si riapre a BootBaud dopo "UPDATE"
	Settle    time.Duration // pausa prima del comando 42
	GreetCmd  byte
	Greet     string
	ModelAt   int  // posizione dei 32 byte del modello controllati dalla radio (-1: nessun controllo)
	V3        bool // argomento a 16 bit, coda 55, errori nell'argomento
	WaitEnd   bool // attesa della conferma del comando di fine
	MaxPkt    int
}

var (
	kdhxProto = &kdh2Variant{ID: "kdhx", Hello: []byte("PROGROMH802U"), HelloBaud: 9600, BootBaud: 115200, Reopen: true,
		Settle: 100 * time.Millisecond, GreetCmd: 0x01, Greet: "BOOTLOADER", ModelAt: -1, WaitEnd: true, MaxPkt: 255}
	bj9000Proto = &kdh2Variant{ID: "bj9000", Hello: []byte("PROGRAMBT9000U"), HelloBaud: 115200, BootBaud: 115200,
		Settle: 80 * time.Millisecond, GreetCmd: 0x0A, Greet: "BOOTLOADER_V3", ModelAt: 992, V3: true, MaxPkt: 65535}
)

const (
	kdh2Wait    = 2 * time.Second
	kdh2Retries = 5
	kdh2Fill    = 0x00
)

func (v *kdh2Variant) pack(cmd byte, arg uint16, data []byte) []byte {
	if !v.V3 {
		return kdhPack(cmd, byte(arg), data)
	}
	n := len(data)
	b := make([]byte, 6+n+3)
	b[0], b[1], b[2], b[3], b[4], b[5] = 0xAA, cmd, byte(arg>>8), byte(arg), byte(n>>8), byte(n)
	copy(b[6:], data)
	c := crc16x(b[1 : 6+n])
	b[6+n], b[7+n], b[8+n] = byte(c>>8), byte(c), 0x55
	return b
}

// hdrLen: byte dopo AA fino alla lunghezza compresa
func (v *kdh2Variant) hdrLen() int {
	if v.V3 {
		return 5
	}
	return 4
}

// countData: dati del comando 04
func (v *kdh2Variant) countData(n int) []byte {
	if !v.V3 {
		return []byte{byte(n)}
	}
	if n > 1 {
		return []byte{byte((n - 1) >> 8), byte(n - 1)}
	}
	return []byte{0, byte(n)}
}

var kdh2ErrText = map[byte]string{
	0xE1: "la radio ha rifiutato la presentazione (firmware o radio non compatibili)",
	0xE3: "la radio ha segnalato un indirizzo errato",
	0xE4: "la radio ha segnalato un errore di scrittura della memoria flash",
	0xE5: "la radio ha segnalato un comando errato",
	0xE6: "la radio ha rifiutato il firmware: è per un altro modello",
}

type kdh2Host struct {
	ctx     context.Context
	name    string
	p       Port
	v       *kdh2Variant
	last    []byte
	logf    func(string)
	started bool
	onStart func()
}

func (h *kdh2Host) sleep(d time.Duration) error {
	select {
	case <-h.ctx.Done():
		return errCancelled
	case <-time.After(d):
		return nil
	}
}

func (h *kdh2Host) write(b []byte) error {
	h.last = b
	return h.p.Write(b)
}

// reopen: come il programma originale chiude la porta e la riapre a un'altra velocità
func (h *kdh2Host) reopen(baud int) error {
	_ = h.p.Close()
	p, err := openPortFunc(h.name, baud)
	if err != nil {
		h.p = nopPort{}
		return fmt.Errorf("riapertura della porta %s a %d baud non riuscita: %v", h.name, baud, err)
	}
	h.p = p
	return nil
}

// resp attende una risposta: 2 secondi, che ripartono all'arrivo di AA.
// code è l'argomento (kdhx) o il suo byte basso (bj9000, come nell'originale).
func (h *kdh2Host) resp() (cmd, code byte, saw bool, err error) {
	deadline := time.Now().Add(kdh2Wait)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return 0, 0, false, errTimeout
		}
		b, err := readByte(h.ctx, h.p, left)
		if err != nil {
			return 0, 0, false, err
		}
		if b == 0xAA {
			break
		}
	}
	deadline = time.Now().Add(kdh2Wait)
	bad := func() (byte, byte, bool, error) {
		// l'originale resta in attesa fino allo scadere del tempo
		if err := h.sleep(time.Until(deadline)); err != nil {
			return 0, 0, true, err
		}
		_ = h.p.Purge()
		return 0, 0, true, errKdhBad
	}
	hl := h.v.hdrLen()
	hdr := make([]byte, hl)
	if err := h.p.ReadFull(h.ctx, hdr, time.Until(deadline)); err != nil {
		return 0, 0, true, err
	}
	n := int(hdr[hl-2])<<8 | int(hdr[hl-1])
	if n > 112 { // buffer di ricezione dell'originale: 128 byte
		return bad()
	}
	rest := make([]byte, n+3)
	if err := h.p.ReadFull(h.ctx, rest, time.Until(deadline)); err != nil {
		return 0, 0, true, err
	}
	c := crc16x(append(append([]byte{}, hdr...), rest[:n]...))
	if c != uint16(rest[n])<<8|uint16(rest[n+1]) {
		return bad()
	}
	return hdr[0], hdr[hl-3], true, nil
}

// exchange invia un pacchetto e attende la conferma del comando cmd.
func (h *kdh2Host) exchange(pkt []byte) error {
	cmd := pkt[1]
	if err := h.write(pkt); err != nil {
		return err
	}
	left, naks := kdh2Retries, 0
	for {
		rc, code, saw, err := h.resp()
		if saw {
			left = kdh2Retries
		}
		if err == nil {
			ecode := rc
			if h.v.V3 {
				ecode = code
			}
			switch {
			case code == 6 && rc == cmd:
				if !h.started {
					h.started = true
					if h.onStart != nil {
						h.onStart()
					}
				}
				return nil
			case code != 6 && ecode == 0xE2:
				naks++
				if naks > kdhMaxNak {
					return rErr("la radio continua a segnalare errori di controllo dei dati")
				}
				h.logf(fmt.Sprintf("La radio ha segnalato un errore di controllo: nuovo invio (%d)", naks))
				err = errKdhBad
				saw = false
			case code != 6 && kdh2ErrText[ecode] != "" && (h.v.V3 || ecode != 0xE6):
				return rErr("%s", kdh2ErrText[ecode])
			default:
				// conferma di un altro comando o risposta sconosciuta: come
				// l'originale si attende lo scadere del tempo
				if err := h.sleep(kdh2Wait); err != nil {
					return err
				}
				err = errKdhBad
			}
		}
		if err != errTimeout && err != errKdhBad {
			return err
		}
		if left <= 0 {
			return errTimeout
		}
		left--
		if saw || err == errTimeout {
			h.logf(fmt.Sprintf("Nessuna risposta valida: nuovo invio %d di %d", kdh2Retries-left, kdh2Retries))
		}
		if err := h.p.Write(h.last); err != nil {
			return err
		}
	}
}

// wait06 attende il byte 06 (gli altri sono scartati) entro il tempo indicato.
func (h *kdh2Host) wait06(d time.Duration) error {
	deadline := time.Now().Add(d)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		b, err := readByte(h.ctx, h.p, left)
		if err != nil {
			return err
		}
		if b == 6 {
			return nil
		}
	}
}

// handshake della sequenza automatica (HandShake_0)
func (h *kdh2Host) handshake() error {
	if err := h.write(h.v.Hello); err != nil {
		return err
	}
	left := kdh2Retries
	for {
		err := h.wait06(kdh2Wait)
		if err == nil {
			break
		}
		if err != errTimeout {
			return err
		}
		if left <= 0 {
			return rErr("la radio non risponde alla richiesta di aggiornamento")
		}
		left--
		h.logf(fmt.Sprintf("Nessuna risposta: nuovo invio %d di %d", kdh2Retries-left, kdh2Retries))
		if err := h.p.Write(h.v.Hello); err != nil {
			return err
		}
	}
	h.logf("La radio ha risposto: richiesta di aggiornamento…")
	if err := h.write([]byte("UPDATE")); err != nil {
		return err
	}
	// l'originale non ripete "UPDATE": attende fino a 12 secondi
	if err := h.wait06(time.Duration(kdh2Retries+1) * kdh2Wait); err != nil {
		if err == errTimeout {
			return rErr("la radio non ha accettato la richiesta di aggiornamento")
		}
		return err
	}
	if h.v.Reopen {
		if err := h.reopen(h.v.BootBaud); err != nil {
			return err
		}
	}
	return h.sleep(h.v.Settle)
}

// kdh2Update invia il firmware (kdhx o bj9000); restituisce i byte inviati.
// started viene chiamata alla prima conferma del caricatore.
func kdh2Update(ctx context.Context, port string, raw []byte, v *kdh2Variant, manual bool, prog Progress, logf func(string), started func()) (int64, error) {
	if logf == nil {
		logf = func(string) {}
	}
	n := (len(raw) + kdhPkt - 1) / kdhPkt
	if n == 0 || n > v.MaxPkt {
		return 0, errors.New("dimensione del firmware non valida per questo caricatore")
	}
	if v.ModelAt >= 0 && len(raw) < v.ModelAt+32 {
		return 0, errors.New("file del firmware troppo corto")
	}
	baud := v.HelloBaud
	if manual {
		baud = v.BootBaud
	}
	p, err := openPortFunc(port, baud)
	if err != nil {
		return 0, err
	}
	h := &kdh2Host{ctx: ctx, name: port, p: p, v: v, logf: logf, onStart: started}
	defer func() { _ = h.p.Close() }()
	total := n + 1
	prog(0, total, "Collegamento con la radio…")
	if !manual {
		logf(fmt.Sprintf("Richiesta di aggiornamento alla radio accesa (%d baud)…", baud))
		if err := h.handshake(); err != nil {
			return 0, err
		}
		logf("Avvio del caricatore della radio…")
		if err := h.exchange(v.pack(0x42, 0, nil)); err != nil {
			return 0, err
		}
		logf("Caricatore attivo")
	}
	logf("Presentazione al caricatore…")
	if err := h.exchange(v.pack(v.GreetCmd, 0, []byte(v.Greet))); err != nil {
		return 0, err
	}
	if v.ModelAt >= 0 {
		logf("Controllo del modello da parte della radio…")
		if err := h.exchange(v.pack(0x02, 0, raw[v.ModelAt:v.ModelAt+32])); err != nil {
			return 0, err
		}
		logf("La radio ha accettato il modello indicato nel firmware")
	}
	logf(fmt.Sprintf("Invio del firmware: %d pacchetti da 1024 byte", n))
	if err := h.exchange(v.pack(0x04, 0, v.countData(n))); err != nil {
		return 0, err
	}
	var sent int64
	for i := 0; i < n; i++ {
		blk := make([]byte, kdhPkt)
		for j := range blk {
			blk[j] = kdh2Fill
		}
		end := min((i+1)*kdhPkt, len(raw))
		copy(blk, raw[i*kdhPkt:end])
		prog(i, total, "Invio del firmware…")
		if err := h.exchange(v.pack(0x03, uint16(i), blk)); err != nil {
			return sent, err
		}
		sent += int64(end - i*kdhPkt)
	}
	prog(n, total, "Fine dell'aggiornamento…")
	if v.WaitEnd {
		if err := h.exchange(v.pack(0x45, 0, nil)); err != nil {
			if err == errTimeout {
				return sent, rErr("il firmware è stato inviato tutto, ma la radio non ha confermato la fine dell'aggiornamento")
			}
			return sent, err
		}
		logf("La radio ha confermato la fine dell'aggiornamento")
	} else {
		if err := h.p.Write(v.pack(0x45, 0, nil)); err != nil {
			return sent, err
		}
		// l'originale chiude subito la porta; qui si attende un attimo la conferma
		if rc, code, _, err := h.resp(); err == nil && rc == 0x45 && code == 6 {
			logf("La radio ha confermato la fine dell'aggiornamento")
		} else {
			logf("Comando di fine aggiornamento inviato")
		}
	}
	prog(total, total, "Aggiornamento completato")
	return sent, nil
}

// ---------------------------------------------------------------- UV-9G (.kdh)

const (
	k9gBaud    = 38400
	k9gWait    = time.Second
	k9gRetries = 5
	k9gErrors  = 3
	k9gBlock   = 32
	k9gMax     = 65536
)

var (
	k9gHello = []byte(".Program")
	k9gHs    = []byte{0xAA, 0x4B, 0x44, 0xCC, 0x48, 0xEF}
)

func k9gFrame(addr int, data []byte) []byte {
	b := make([]byte, 38)
	b[0], b[1], b[2] = 0xAA, byte(addr>>8), byte(addr)
	copy(b[3:35], data)
	c := crc16x(b[1:35])
	b[35], b[36], b[37] = byte(c>>8), byte(c), 0xEF
	return b
}

// k9gFirst attende il primo byte di una risposta e scarta il resto (l'originale
// legge tutti i byte arrivati e guarda il primo).
func k9gFirst(ctx context.Context, p Port, d time.Duration) (byte, error) {
	b, err := readByte(ctx, p, d)
	if err != nil {
		return 0, err
	}
	time.Sleep(5 * time.Millisecond)
	_ = p.Purge()
	return b, nil
}

// k9gWait06: attesa di 06 entro d (gli altri byte sono ignorati)
func k9gWait06(ctx context.Context, p Port, d time.Duration) error {
	deadline := time.Now().Add(d)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		b, err := k9gFirst(ctx, p, left)
		if err != nil {
			return err
		}
		if b == 6 {
			return nil
		}
	}
}

// kdh9gUpdate invia un file .kdh alla UV-9G (caricatore a 38400 baud).
func kdh9gUpdate(ctx context.Context, port string, raw []byte, manual bool, prog Progress, logf func(string), started func()) (int64, error) {
	if logf == nil {
		logf = func(string) {}
	}
	if len(raw) == 0 || len(raw) > k9gMax {
		return 0, errors.New("dimensione del firmware non valida per questo caricatore")
	}
	baud := 9600
	if manual {
		baud = k9gBaud
	}
	p, err := openPortFunc(port, baud)
	if err != nil {
		return 0, err
	}
	h := &kdh2Host{ctx: ctx, name: port, p: p, logf: logf}
	defer func() { _ = h.p.Close() }()
	n := (len(raw) + k9gBlock - 1) / k9gBlock
	prog(0, n, "Collegamento con la radio…")
	// tentativi con nuovo invio dopo 1 secondo
	try := func(pkt []byte, what string) error {
		if err := h.p.Write(pkt); err != nil {
			return err
		}
		for left := k9gRetries; ; left-- {
			err := k9gWait06(ctx, h.p, k9gWait)
			if err == nil {
				return nil
			}
			if err != errTimeout {
				return err
			}
			if left <= 0 {
				return rErr("%s", what)
			}
			logf(fmt.Sprintf("Nessuna risposta: nuovo invio %d di %d", k9gRetries-left+1, k9gRetries))
			if err := h.p.Write(pkt); err != nil {
				return err
			}
		}
	}
	if !manual {
		logf("Richiesta di aggiornamento alla radio accesa (9600 baud)…")
		if err := try(k9gHello, "la radio non risponde alla richiesta di aggiornamento"); err != nil {
			return 0, err
		}
		logf("La radio ha risposto: avvio del caricatore a 38400 baud…")
		if err := h.p.Write([]byte{'U'}); err != nil {
			return 0, err
		}
		if err := h.reopen(k9gBaud); err != nil {
			return 0, err
		}
		if err := h.sleep(500 * time.Millisecond); err != nil {
			return 0, err
		}
	}
	logf("Presentazione al caricatore…")
	if err := try(k9gHs, "il caricatore della radio non risponde"); err != nil {
		return 0, err
	}
	if started != nil {
		started()
	}
	logf(fmt.Sprintf("Invio del firmware: %d blocchi da 32 byte", n))
	fail := func(err error) error {
		_ = h.p.Write([]byte{'E'})
		return err
	}
	var sent int64
	for i := 0; i < n; i++ {
		end := min((i+1)*k9gBlock, len(raw))
		fr := k9gFrame(i*k9gBlock, raw[i*k9gBlock:end])
		if i%32 == 0 {
			prog(i, n, "Invio del firmware…")
		}
		if err := h.p.Write(fr); err != nil {
			return sent, err
		}
		left, errs := k9gRetries, k9gErrors
		deadline := time.Now().Add(k9gWait)
		for {
			b, err := k9gFirst(ctx, h.p, time.Until(deadline))
			if err == nil && b == 6 {
				break
			}
			if err == nil && b >= 0xE1 && b <= 0xE3 {
				what := map[byte]string{0xE1: "errore nei dati", 0xE2: "indirizzo fuori dall'area utente", 0xE3: "errore di scrittura della flash"}[b]
				if errs <= 0 {
					return sent, fail(rErr("la radio ha segnalato ripetutamente: %s", what))
				}
				errs--
				left = k9gRetries
				logf(fmt.Sprintf("La radio ha segnalato: %s; nuovo invio (%d)", what, k9gErrors-errs))
			} else if err == nil {
				continue // byte sconosciuto: si attende lo scadere del tempo
			} else if err != errTimeout {
				return sent, err
			} else {
				if left <= 0 {
					return sent, fail(errTimeout)
				}
				left--
				logf(fmt.Sprintf("Nessuna risposta: nuovo invio %d di %d", k9gRetries-left, k9gRetries))
			}
			if err := h.p.Write(fr); err != nil {
				return sent, err
			}
			deadline = time.Now().Add(k9gWait)
		}
		sent += int64(end - i*k9gBlock)
	}
	if err := h.p.Write([]byte{'E'}); err != nil {
		return sent, err
	}
	logf("Comando di fine aggiornamento inviato")
	prog(n, n, "Aggiornamento completato")
	return sent, nil
}
