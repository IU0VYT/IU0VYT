// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Firmware della Retevis RA89 (.bin, 4.4) e della Retevis RT85 (.hex, 4.5), con il
// protocollo dei programmi del costruttore «TH8600-UpDAta Soft» (RA89:
// «RA89-update-software», Ver 2.3 20230621, lo stesso delle TYT TH-UV88 / TH-UV98 e
// della Retevis RT45P; RT85: «RT85 10W_5W_UpData», Ver 2.1 20200701), ricavato dal
// loro codice e confrontato con lo stesso codice fatto girare su una radio
// simulata (verifica/ra89):
//   - 57600 baud 8N1 (scelta proposta; anche 115200), cavo a due fili (nessuna eco);
//   - prima di ogni invio una pausa: 2 ms a 57600 baud, 22 ms a 115200 (4.5: i valori
//     che il programma imposta scegliendo la velocità; la 4.4 usava 17 e 20 ms);
//   - presentazione «AA xx EF» → la radio risponde con gli stessi tre byte (4.5: la
//     4.4 aspettava «AA 01 EF» anche dalla RA89, che risponde «AA 09 EF»):
//     RA89 «AA 09 EF»; RT85 5 W «AA 01 EF», RT85 10 W «AA 09 EF»;
//   - «AA 10 00 00 EF» → «AA 11 EF»;
//   - i dati a blocchi: «55», lunghezza (1 byte), dati, somma dei dati (1 byte) →
//     «AA 02 EF» a ogni blocco. RA89: il file .bin così com'è, a blocchi di 128
//     byte. RT85: dal file .hex (Intel HEX) i dati di 4 righe alla volta (64 byte)
//     seguiti dalla loro somma, come li prepara il programma (FileDataDetailPro);
//   - «55 00 00» (fine) → «AA 02 ..»; poi «AA FF EF» e la porta si chiude.
// Le pause tra un passo e l'altro e l'attesa delle risposte (2,5 secondi per la
// RA89, 1,5 per la RT85) sono quelle dei due programmi; nessun nuovo tentativo
// (come negli originali). Il caricatore non controlla il file (cifrato): conferma
// scritta del modello.

import (
	"bytes"
	"context"
	"encoding/hex"
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
	thSleepMul = time.Duration(1) // solo per le prove (pause del programma originale)
	thAckTest  time.Duration      // solo per le prove: attesa delle risposte (0: quella del programma)
)

// thDelayFor: pausa prima di ogni invio (GbCommDelay): la scelta della velocità
// scrive 5 (57600) o 22 (115200) nella casella del ritardo, e all'avvio a 57600
// baud il programma toglie 3 ms
func thDelayFor(baud int) time.Duration {
	if baud == 115200 {
		return 22 * time.Millisecond
	}
	return 2 * time.Millisecond
}

// thProfile: le pause dei due programmi (EnterWriteModPro, SendAckPro,
// SendDataPro, CommunicatingPro, TimeRestartPro)
type thProfile struct {
	name    string
	hello   time.Duration // prima della presentazione
	start   time.Duration // dopo la presentazione, prima di «AA 10 00 00 EF»
	first   time.Duration // dopo «AA 11 EF», prima del primo blocco
	block   time.Duration // prima di ogni blocco
	end     time.Duration // prima di «55 00 00»
	exit    time.Duration // prima di «AA FF EF»
	ackWait time.Duration
	ackSecs string // per i messaggi
}

var (
	thRA89 = thProfile{"RA89", 600 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 20 * time.Millisecond,
		300 * time.Millisecond, 80 * time.Millisecond, 2500 * time.Millisecond, "2,5"}
	thRT85 = thProfile{"RT85", 300 * time.Millisecond, 100 * time.Millisecond, 50 * time.Millisecond, 5 * time.Millisecond,
		20 * time.Millisecond, 50 * time.Millisecond, 1500 * time.Millisecond, "1,5"}
)

// presentazioni (EnterWriteModPro e comboBox1_SelectedIndexChanged): RA89 come la
// TH-UV98 «AA 09 EF» (UV88 «AA 01 EF», RT45P «AA 08 EF»); RT85 5 W «AA 01 EF»
// (come la TH-UV88, di cui è la versione Retevis), 10 W «AA 09 EF»
const (
	thHelloRA89    = 0x09
	thHelloRT85_5W = 0x01
	thHelloRT85_10 = 0x09
)

// thTargets: radio con questi programmi (UV88, UV98 e RT45P con i loro file non
// sono gestite)
var thTargets = []fwTarget{
	{[]string{"RA89"}, "Retevis RA89", "RA89", []string{"rt-ra89"}},
	{[]string{"RT85"}, "Retevis RT85", "RT85", []string{"rt-rt85"}},
}

var fwKnownTH8600 = map[string]string{
	"3cd181ed58251b5f6a8c066701d96ef9e05a4ebaeea46fd653d00d5df970c9ad": "Retevis RA89, versione B91.38 del 24/05/2025 (Eng_RA89_Upadat_20250524_B91.38.bin, retevis.com)",
	"374a6c6d4af49c521f82beff3330a8471ec409cb291c8fd4960d725a8a9a9aea": "Retevis RT85, versione 1.53 (RT85-Firmware-V1.53.hex, retevis.com)",
}

var reThVer = regexp.MustCompile(`(?i)B(\d{2})\.(\d{2})`)

// thFile: il file è un firmware della RA89 (.bin) o della RT85 (.hex), dal nome
// (il contenuto è cifrato)?
func thFile(name string) bool {
	n := normModel(path.Base(name))
	switch strings.ToLower(path.Ext(name)) {
	case ".bin":
		return strings.Contains(n, "RA89")
	case ".hex":
		return strings.Contains(n, "RT85")
	}
	return false
}

// thIsRT85: file della RT85 (.hex)
func thIsRT85(name string) bool { return strings.EqualFold(path.Ext(name), ".hex") }

// analyzeTH8600 controlla un file .bin della Retevis RA89 o .hex della RT85
func analyzeTH8600(name, origin string, raw []byte) *fwImage {
	rt85 := thIsRT85(name)
	t := &thTargets[0]
	if rt85 {
		t = &thTargets[1]
	}
	img, info, add := newInfo("th8600", t.target, t.confirm, name, origin, raw)
	info.Typed = true
	info.Target, info.Confirm, info.Models = t.target, t.confirm, t.drivers
	data := raw
	n := 0
	if rt85 {
		blocks, d, err := thHexBlocks(raw)
		if err != nil {
			add("bad", "%s", err)
			return img
		}
		data, n = d, len(blocks)
		info.Power = true
	}
	if len(data) < thMin || len(data) > thMax {
		add("bad", "Dimensione non plausibile per un firmware della %s (%s)", t.confirm, fmtBytes(len(data)))
		return img
	}
	if bytes.Count(data, []byte{0xFF})*4 > len(data) || bytes.Count(data, []byte{0x00})*4 > len(data) {
		add("bad", "Il contenuto del file non è un firmware cifrato della %s (troppi byte vuoti): file sbagliato o danneggiato", t.confirm)
		return img
	}
	base := strings.TrimSuffix(path.Base(name), path.Ext(name))
	if rt85 {
		add("ok", "File Intel HEX con %s di firmware, %d blocchi da 64 byte (preparati come nel programma di Retevis)", fmtBytes(len(data)), n)
		if m := reThVerV.FindStringSubmatch(base); m != nil {
			info.Version = m[1]
		}
	} else {
		n = (len(raw) + thBlock - 1) / thBlock
		add("ok", "Firmware di %s, %d blocchi da 128 byte (inviato così com'è, come nel programma di Retevis)", fmtBytes(len(raw)), n)
		if m := reThVer.FindStringSubmatch(base); m != nil {
			info.Version = "B" + m[1] + "." + m[2]
		}
	}
	if m := reDate8.FindStringSubmatch(base); m != nil {
		info.Date = m[3] + "/" + m[2] + "/" + m[1]
	}
	add("warn", "Il file è cifrato e il caricatore della radio non controlla il modello: usa solo il firmware di Retevis per la %s. Scrivi %s per confermare", t.confirm, t.confirm)
	if rt85 {
		add("warn", "Scegli la potenza della tua RT85 (5 W o 10 W, come nel programma di Retevis): la radio risponde solo alla presentazione del suo modello")
	}
	add("warn", "Retevis avverte: se l'aggiornamento si interrompe NON accendere la radio normalmente, ma riportala subito in modalità di aggiornamento e ripeti l'invio")
	info.Warnings = true
	if k, ok := fwKnownTH8600[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	} else {
		add("warn", "File non presente nell'elenco dei firmware provati da questo programma")
	}
	img.Dec = data
	info.Valid = true
	return img
}

var reThVerV = regexp.MustCompile(`(?i)V(\d+\.\d+)`)

// thHexBlocks prepara i blocchi di un file .hex della RT85 come il programma di
// Retevis (FileDataDetailPro): per ogni gruppo di 4 righe di dati una riga con i
// loro byte (in genere 64) seguiti dalla somma; alla riga di fine file
// («:00000001FF») si aggiunge la somma della riga in corso (se il gruppo era già
// completo la riga ha così due somme e il blocco inviato contiene la prima: come
// nel programma). Si invia ogni riga con «55», la lunghezza senza l'ultimo byte,
// e i suoi byte: cioè il blocco (la riga senza l'ultima somma) e la sua somma. Gli indirizzi non vengono inviati: il
// programma li ignora, quindi qui si accettano solo file continui da 0 con righe
// di dati e la riga di fine. Le somme di controllo delle righe dei file di
// Retevis non sono quelle dell'Intel HEX (calcolate sui dati prima della
// cifratura): come il programma, non si usano.
// Restituisce anche i dati del firmware (senza le somme).
func thHexBlocks(raw []byte) ([][]byte, []byte, error) {
	txt := strings.ReplaceAll(string(raw), "\r\n", "\n")
	txt = strings.ReplaceAll(txt, "\r", "\n")
	lines := strings.Split(txt, "\n")
	var groups [][]byte
	var data []byte
	addr := 0
	sum := func(b []byte) byte {
		var s byte
		for _, c := range b {
			s += c
		}
		return s
	}
	for i, l := range lines {
		l = strings.TrimSpace(l)
		bad := func(f string, a ...interface{}) ([][]byte, []byte, error) {
			return nil, nil, fmt.Errorf("riga %d del file .hex: %s", i+1, fmt.Sprintf(f, a...))
		}
		if !strings.HasPrefix(l, ":") || len(l) < 11 || len(l)%2 == 0 {
			return bad("non è una riga Intel HEX")
		}
		rec, err := hex.DecodeString(l[1:])
		if err != nil {
			return bad("caratteri non validi")
		}
		if int(rec[0])+5 != len(rec) {
			return bad("lunghezza non valida")
		}
		switch rec[3] {
		case 0x00:
		case 0x01:
			if len(groups) == 0 {
				return bad("nessun dato prima della fine del file")
			}
			g := &groups[len(groups)-1]
			*g = append(*g, sum(*g))
			// ogni riga preparata finisce con la somma dei byte precedenti, che
			// è anche la somma del pacchetto (thPacket): il blocco è il resto
			blocks := make([][]byte, len(groups))
			for k, g := range groups {
				if len(g)-1 > 255 {
					return nil, nil, fmt.Errorf("righe del file .hex troppo lunghe per il programma della RT85")
				}
				blocks[k] = g[:len(g)-1]
			}
			return blocks, data, nil
		default:
			return bad("record di tipo %02X non usato dai firmware della RT85", rec[3])
		}
		if a := int(rec[1])<<8 | int(rec[2]); a != addr {
			return bad("indirizzo %04X invece di %04X: il file non è continuo (non è un firmware della RT85)", a, addr)
		}
		d := rec[4 : len(rec)-1]
		addr += len(d)
		if addr > 0x10000 {
			return bad("indirizzo oltre 64 KB")
		}
		data = append(data, d...)
		if i%4 == 0 {
			groups = append(groups, nil)
		}
		g := &groups[len(groups)-1]
		*g = append(*g, d...)
		if i%4 == 3 {
			*g = append(*g, sum(*g))
		}
	}
	return nil, nil, fmt.Errorf("manca la riga di fine del file .hex (file incompleto)")
}

// thFwBlocks: i blocchi da inviare (RA89: il file a blocchi di 128 byte)
func thFwBlocks(img *fwImage) ([][]byte, error) {
	if thIsRT85(img.Info.Name) {
		b, _, err := thHexBlocks(img.Raw)
		return b, err
	}
	var out [][]byte
	for i := 0; i < len(img.Dec); i += thBlock {
		out = append(out, img.Dec[i:min(i+thBlock, len(img.Dec))])
	}
	return out, nil
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

// thUpdate invia i blocchi del firmware alla radio in modalità di aggiornamento.
func thUpdate(ctx context.Context, port string, pr thProfile, hello byte, blocks [][]byte, prog Progress, logf func(string), started func()) (int64, error) {
	total := 0
	for _, b := range blocks {
		if len(b) == 0 || len(b) > 255 {
			return 0, fmt.Errorf("blocco del firmware non valido")
		}
		total += len(b)
	}
	if total < thMin || total > thMax+thMax/64 {
		return 0, fmt.Errorf("dimensione del firmware non valida")
	}
	p, err := openPort(port, thBaud)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	delay := thDelayFor(thBaud)
	sleep := func(d time.Duration) error { return sleepCtx(ctx, d*thSleepMul) }
	ack := make([]byte, 3)
	// SendDAtaPro: pausa, scarto dei byte arrivati, invio; poi 3 byte di risposta
	send := func(b []byte, want func([]byte) bool, what string) error {
		if err := sleep(delay); err != nil {
			return err
		}
		_ = p.Purge()
		if err := p.Write(b); err != nil {
			return err
		}
		wait := pr.ackWait
		if thAckTest > 0 {
			wait = thAckTest
		}
		if err := p.ReadFull(ctx, ack, wait); err != nil {
			if err == errCancelled {
				return err
			}
			return rErr("%s: nessuna risposta della radio entro %s secondi", what, pr.ackSecs)
		}
		if !want(ack) {
			return rErr("%s: risposta inattesa della radio (% X)", what, ack)
		}
		return nil
	}
	is := func(b ...byte) func([]byte) bool { return func(r []byte) bool { return bytes.Equal(r, b) } }
	logf("Presentazione alla radio (la radio deve essere già in modalità di aggiornamento: indicatore verde)…")
	if err := sleep(pr.hello); err != nil {
		return 0, err
	}
	if err := send([]byte{0xAA, hello, 0xEF}, is(0xAA, hello, 0xEF), "presentazione"); err != nil {
		if err == errCancelled {
			return 0, err
		}
		other := ""
		if pr.name == "RT85" {
			other = " o è una RT85 di potenza diversa da quella scelta"
		}
		return 0, rErr("la radio non è in modalità di aggiornamento, non è una %s%s (%v)", pr.name, other, err)
	}
	logf("Radio in modalità di aggiornamento")
	if err := sleep(pr.start); err != nil {
		return 0, err
	}
	if err := send([]byte{0xAA, 0x10, 0x00, 0x00, 0xEF}, is(0xAA, 0x11, 0xEF), "inizio"); err != nil {
		if err == errCancelled {
			return 0, err
		}
		return 0, rErr("la radio non ha accettato l'inizio dell'aggiornamento (%v)", err)
	}
	if err := sleep(pr.first); err != nil {
		return 0, err
	}
	started()
	n := len(blocks)
	var sent int64
	for i, blk := range blocks {
		prog(int(sent), total, "Invio del firmware…")
		// SendDataPro: pausa, poi SendDAtaPro (il primo blocco dopo la pausa di SendAckPro)
		if err := sleep(pr.block); err != nil {
			return sent, err
		}
		if err := send(thPacket(blk), is(0xAA, 0x02, 0xEF), fmt.Sprintf("blocco %d di %d", i+1, n)); err != nil {
			return sent, err
		}
		sent += int64(len(blk))
	}
	if err := sleep(pr.end); err != nil {
		return sent, err
	}
	if err := send([]byte{0x55, 0x00, 0x00}, func(r []byte) bool { return r[0] == 0xAA && r[1] == 0x02 }, "fine dei dati"); err != nil {
		return sent, err
	}
	if err := sleep(pr.exit); err != nil {
		return sent, err
	}
	// «AA FF EF» (uscita): il programma originale chiude la porta dopo 100 ms
	// senza attendere la risposta
	if err := sleep(delay); err != nil {
		return sent, err
	}
	_ = p.Purge()
	if err := p.Write([]byte{0xAA, 0xFF, 0xEF}); err != nil {
		return sent, err
	}
	_ = sleep(100 * time.Millisecond)
	prog(total, total, "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi, tutti confermati dalla radio", n))
	return sent, nil
}
