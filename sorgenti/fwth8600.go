// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware .bin della Retevis RA89 (4.4), con il protocollo del programma del
// costruttore «RA89-update-software» («TH8600-UpDAta Soft» Ver 2.3 20230621,
// lo stesso programma delle TYT TH-UV88 / TH-UV98 e della Retevis RT45P),
// ricavato dal suo codice e confrontato con lo stesso codice fatto girare su una
// radio simulata (verifica/ra89):
//   - 57600 baud 8N1 (scelta proposta; anche 115200), cavo a due fili (nessuna eco);
//   - prima di ogni invio una pausa (20 ms del programma, meno 3 a 57600 baud);
//   - «AA 09 EF» → la radio risponde «AA 01 EF» (modalità di aggiornamento);
//   - «AA 10 00 00 EF» → «AA 11 EF»;
//   - il file è inviato così com'è, a blocchi di 128 byte: «55», lunghezza (1 byte),
//     dati, somma dei dati (1 byte) → «AA 02 EF» a ogni blocco;
//   - «55 00 00» (fine) → «AA 02 ..»; poi «AA FF EF» e la porta si chiude.
// Ogni risposta è attesa al massimo 2,5 secondi; nessun nuovo tentativo (come
// nel programma originale). Il caricatore non controlla il file (cifrato):
// conferma scritta del modello.

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	thBlock = 128
	thMin   = 16 * 1024
	thMax   = 512 * 1024
)

var (
	thBaud     = 57600
	thDelay    = 17 * time.Millisecond // GbCommDelay: 20 ms, meno 3 a 57600 baud
	thAckWait  = 2500 * time.Millisecond
	thSleepMul = time.Duration(1) // solo per le prove (pause del programma originale)
)

// thTargets: radio con questo programma e i file .bin (RA89; UV88, UV98 e RT45P
// usano file .hex, preparati diversamente, e qui non sono gestite). Presentazione
// della RA89: «AA 09 EF» (EnterWriteModPro, come la UV98; UV88 «AA 01 EF», RT45P
// «AA 08 EF»)
const thHelloRA89 = 0x09

var thTargets = []fwTarget{
	{[]string{"RA89"}, "Retevis RA89", "RA89", []string{"rt-ra89"}},
}

var fwKnownTH8600 = map[string]string{
	"3cd181ed58251b5f6a8c066701d96ef9e05a4ebaeea46fd653d00d5df970c9ad": "Retevis RA89, versione B91.38 del 24/05/2025 (Eng_RA89_Upadat_20250524_B91.38.bin, retevis.com)",
}

var reThVer = regexp.MustCompile(`(?i)B(\d{2})\.(\d{2})`)

// thFile: il file è un firmware della RA89 (dal nome: il contenuto è cifrato)?
func thFile(name string) bool {
	return strings.Contains(normModel(path.Base(name)), "RA89")
}

// analyzeTH8600 controlla un file .bin della Retevis RA89
func analyzeTH8600(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("th8600", "Retevis RA89", "RA89", name, origin, raw)
	info.Typed = true
	t := fwTargetFromName(thTargets, name)
	if t == nil {
		t = &thTargets[0]
	}
	info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	if len(raw) < thMin || len(raw) > thMax {
		add("bad", "Dimensione non plausibile per un firmware della RA89 (%s)", fmtBytes(len(raw)))
		return img
	}
	if bytes.Count(raw, []byte{0xFF})*4 > len(raw) || bytes.Count(raw, []byte{0x00})*4 > len(raw) {
		add("bad", "Il contenuto del file non è un firmware cifrato della RA89 (troppi byte vuoti): file sbagliato o danneggiato")
		return img
	}
	n := (len(raw) + thBlock - 1) / thBlock
	add("ok", "Firmware di %s, %d blocchi da 128 byte (inviato così com'è, come nel programma di Retevis)", fmtBytes(len(raw)), n)
	base := strings.TrimSuffix(path.Base(name), path.Ext(name))
	if m := reThVer.FindStringSubmatch(base); m != nil {
		info.Version = "B" + m[1] + "." + m[2]
	}
	if m := reDate8.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/" + m[1]
	}
	add("warn", "Il file è cifrato e il caricatore della radio non controlla il modello: usa solo il firmware di Retevis per la RA89. Scrivi RA89 per confermare")
	add("warn", "Retevis avverte: se l'aggiornamento si interrompe NON accendere la radio normalmente, ma riportala subito in modalità di aggiornamento e ripeti l'invio")
	info.Warnings = true
	if k, ok := fwKnownTH8600[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware provati da questo programma")
	}
	img.Dec = raw
	info.Valid = true
	return img
}

var reDate8 = regexp.MustCompile(`(20\d{2})(0[1-9]|1[0-2])([0-3]\d)`)

// thPacket: blocco del firmware («55», lunghezza, dati, somma)
func thPacket(d []byte) []byte {
	p := append([]byte{0x55, byte(len(d))}, d...)
	var s byte
	for _, b := range d {
		s += b
	}
	return append(p, s)
}

// thUpdate invia il firmware alla radio in modalità di aggiornamento.
func thUpdate(ctx context.Context, port string, fw []byte, hello byte, prog Progress, logf func(string), started func()) (int64, error) {
	if len(fw) < thMin || len(fw) > thMax {
		return 0, fmt.Errorf("dimensione del firmware non valida")
	}
	p, err := openPort(port, thBaud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	sleep := func(d time.Duration) error { return sleepCtx(ctx, d*thSleepMul) }
	ack := make([]byte, 3)
	// SendDAtaPro: pausa, scarto dei byte arrivati, invio; poi 3 byte di risposta
	send := func(b []byte, want func([]byte) bool, what string) error {
		if err := sleep(thDelay); err != nil {
			return err
		}
		_ = p.Purge()
		if err := p.Write(b); err != nil {
			return err
		}
		if err := p.ReadFull(ctx, ack, thAckWait); err != nil {
			if err == errCancelled {
				return err
			}
			return rErr("%s: nessuna risposta della radio entro 2,5 secondi", what)
		}
		if !want(ack) {
			return rErr("%s: risposta inattesa della radio (% X)", what, ack)
		}
		return nil
	}
	is := func(b ...byte) func([]byte) bool { return func(r []byte) bool { return bytes.Equal(r, b) } }
	logf("Presentazione alla radio (la radio deve essere già in modalità di aggiornamento: indicatore verde)…")
	if err := sleep(600 * time.Millisecond); err != nil {
		return 0, err
	}
	if err := send([]byte{0xAA, hello, 0xEF}, is(0xAA, 0x01, 0xEF), "presentazione"); err != nil {
		if err == errCancelled {
			return 0, err
		}
		return 0, rErr("la radio non è in modalità di aggiornamento o non è una RA89 (%v): la radio non è stata modificata", err)
	}
	logf("Radio in modalità di aggiornamento")
	if err := sleep(100 * time.Millisecond); err != nil {
		return 0, err
	}
	if err := send([]byte{0xAA, 0x10, 0x00, 0x00, 0xEF}, is(0xAA, 0x11, 0xEF), "inizio"); err != nil {
		if err == errCancelled {
			return 0, err
		}
		return 0, rErr("la radio non ha accettato l'inizio dell'aggiornamento (%v): la radio non è stata modificata", err)
	}
	if err := sleep(100 * time.Millisecond); err != nil {
		return 0, err
	}
	started()
	n := (len(fw) + thBlock - 1) / thBlock
	var sent int64
	for i := 0; i < n; i++ {
		prog(i*thBlock, len(fw), "Invio del firmware…")
		// SendDataPro: 20 ms, poi SendDAtaPro
		if err := sleep(20 * time.Millisecond); err != nil {
			return sent, err
		}
		blk := fw[i*thBlock : min((i+1)*thBlock, len(fw))]
		if err := send(thPacket(blk), is(0xAA, 0x02, 0xEF), fmt.Sprintf("blocco %d di %d", i+1, n)); err != nil {
			return sent, err
		}
		sent += int64(len(blk))
	}
	if err := sleep(300 * time.Millisecond); err != nil {
		return sent, err
	}
	if err := send([]byte{0x55, 0x00, 0x00}, func(r []byte) bool { return r[0] == 0xAA && r[1] == 0x02 }, "fine dei dati"); err != nil {
		return sent, err
	}
	if err := sleep(80 * time.Millisecond); err != nil {
		return sent, err
	}
	// «AA FF EF» (uscita): il programma originale chiude la porta dopo 100 ms
	// senza attendere la risposta
	if err := sleep(thDelay); err != nil {
		return sent, err
	}
	_ = p.Purge()
	if err := p.Write([]byte{0xAA, 0xFF, 0xEF}); err != nil {
		return sent, err
	}
	_ = sleep(100 * time.Millisecond)
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi da 128 byte, tutti confermati dalla radio", n))
	return sent, nil
}
