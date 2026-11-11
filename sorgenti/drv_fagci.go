// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.18: Quansheng UV-K5 / UV-K6 con il firmware libero «R3b0rn» di fagci
// (https://github.com/fagci/uvk5-fagci-reborn).
//
// La disposizione della memoria è ricavata dai sorgenti del firmware
// (src/settings.h, src/helper/channels.c, src/driver/uart.c) e riscritta in Go:
//
//   - impostazioni (21 byte), due VFO (13 byte l'uno) e 40 «preset» (32 byte
//     l'uno) da 0x0000 a 0x052E;
//   - canali da 23 byte in fondo alla EEPROM, all'indietro: il canale n (da 0) sta
//     a «fine − (n+1)·23», dove «fine» è la dimensione della EEPROM meno i 15832
//     byte della correzione SSB quando la EEPROM è abbastanza grande da contenerla;
//   - il tipo di EEPROM (da 8 KiB a 256 KiB) sta nei tre bit bassi del primo byte:
//     da lì il numero dei canali (298 con la EEPROM originale, fino a 1024);
//   - un canale esiste se il primo carattere del nome è stampabile.
//
// Comandi: quelli della UV-K5 a 38400 baud, ma con indirizzi a 32 bit (lettura
// 0x051B, scrittura 0x051D a gruppi di 8 byte).
//
// Scelte di prudenza: si leggono e si scrivono solo i canali (impostazioni, VFO,
// preset e correzione SSB non si scrivono mai); il firmware che non risponde alla
// lettura con indirizzi a 32 bit, o che non ha 40 preset, non si legge e non si
// scrive perché la sua memoria è disposta in un altro modo.
//
// Differenze volute dal driver per CHIRP pubblicato con il firmware: la frequenza
// di trasmissione si scrive sempre per intero come fa la radio quando salva un
// canale (quel driver scrive 0 per i canali senza scostamento, e il firmware
// sceglie il ricevitore dalla frequenza di trasmissione); il nome ha al massimo 9
// caratteri, quelli che la radio stessa scrive.

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
)

type fagciFam struct{}

const (
	fagciMagic   = "IU0VYT-FAGCI-R3\x00" // 16 byte
	fagciHdr     = 32
	fagciRec     = 23
	fagciPresets = 40
	fagciStart   = 21 + 2*13 + fagciPresets*32 // 0x52F: primo byte dopo i preset
	fagciHead    = fagciStart                  // impostazioni, VFO e preset
	fagciChanAt  = fagciHdr + fagciHead
	fagciPatch   = 15832
	fagciMax     = 1024
	fagciNameMax = 9
	fagciMinHz   = 153000
	fagciMaxHz   = 1340000000
)

// dimensione della EEPROM per tipo (EEPROM_SIZES del firmware)
var fagciSizes = [8]int{8192, 8192, 8192, 16384, 32768, 65536, 131072, 262144}

func init() {
	genRegister(&genModel{ID: "qs-uvk5-fagci", Vendor: "Quansheng", Model: "UV-K5", Variant: "fagci R3b0rn", RClass: "IU0VYT_Quansheng_R3b0rn",
		Name:  "Quansheng UV-K5 / UV-K6 (firmware R3b0rn di fagci)",
		Slots: fagciMax, First: 1, NameMax: fagciNameMax, Charset: charsetASCII,
		Powers: []genPower{{"Minima", 0.01, 3}, {"Bassa", 0.5, 2}, {"Media", 2, 1}, {"Alta", 5, 0}},
		Bands:  []band{{fagciMinHz, fagciMaxHz}}, Narrow: true, HasScan: true, NoTxOff: true, NoChirp: true,
		DCS: dtcsCodes, Baud: 38400, Fam: fagciFam{},
		Opt: &uvk5Opt{approve: func(fw string) bool { return strings.Contains(strings.ToUpper(fw), "R3B0RN") }, emuFW: "R3b0rn"},
		Notes: "Firmware libero R3b0rn di fagci (la radio risponde «R3b0rn»). Accendi la radio, poi collega il cavo al connettore " +
			"microfono/altoparlante. I canali dipendono dalla memoria montata: 298 con quella originale, fino a 1024 con una memoria " +
			"più grande; le posizioni oltre quelle della radio non si possono usare. Nomi di 9 caratteri; un canale senza nome prende " +
			"come nome la frequenza, perché per il firmware un canale senza nome non esiste. «Scan» mette il canale nella " +
			"lista di scansione 1 (le altre liste si conservano). Modulazione e ricevitore del canale restano quelli impostati sulla " +
			"radio (FM per un canale nuovo, AM nella banda aerea). Si leggono e si scrivono solo i canali: impostazioni, VFO, " +
			"preset e correzione SSB non si toccano. Le versioni vecchie del firmware, con la memoria disposta in altro modo, " +
			"vengono rifiutate. Ricavato dai sorgenti del firmware."})
}

// fagciGeom: fine della zona dei canali e numero dei canali per una EEPROM di
// size byte (getChannelsEnd e CHANNELS_GetCountMax del firmware)
func fagciGeom(size int) (end, n int) {
	end = size
	if size >= fagciStart+fagciRec+fagciPatch {
		end = size - fagciPatch
	}
	n = (end - fagciStart) / fagciRec
	if n > fagciMax {
		n = fagciMax
	}
	return end, n
}

// fagciNew: immagine vuota per una EEPROM del tipo indicato (2-7), con la testa
// (impostazioni, VFO, preset) passata
func fagciNew(head []byte, typ int) []byte {
	size := fagciSizes[typ&7]
	end, n := fagciGeom(size)
	img := make([]byte, fagciChanAt+n*fagciRec)
	copy(img, fagciMagic)
	binary.LittleEndian.PutUint32(img[16:], uint32(size))
	binary.LittleEndian.PutUint32(img[20:], uint32(end))
	binary.LittleEndian.PutUint32(img[24:], uint32(n))
	copy(img[fagciHdr:fagciChanAt], head)
	return img
}

// count: canali della radio da cui viene l'immagine (0 se non è valida)
func (fagciFam) count(img []byte) int {
	if len(img) < fagciChanAt || string(img[:16]) != fagciMagic {
		return 0
	}
	size := int(binary.LittleEndian.Uint32(img[16:]))
	end, n := fagciGeom(size)
	if size < 8192 || int(binary.LittleEndian.Uint32(img[20:])) != end || int(binary.LittleEndian.Uint32(img[24:])) != n ||
		len(img) != fagciChanAt+n*fagciRec {
		return 0
	}
	return n
}

// rec: i 23 byte del canale i (nil se la radio non ha quella posizione). I canali
// stanno all'indietro: il canale 0 è l'ultimo record
func (f fagciFam) rec(img []byte, i int) []byte {
	n := f.count(img)
	if i < 0 || i >= n {
		return nil
	}
	a := fagciChanAt + (n-1-i)*fagciRec
	return img[a : a+fagciRec]
}

func (f fagciFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f fagciFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if r := f.rec(img, i); r != nil && len(raw) == fagciRec {
		copy(r, raw)
	}
}

func (f fagciFam) Match(m *genModel, img []byte) bool { return f.count(img) > 0 }

// le posizioni oltre i canali della radio non esistono
func (f fagciFam) lockedSlotImg(m *genModel, img []byte, i int) bool { return i >= f.count(img) }

func (f fagciFam) LockedImg(m *genModel, img []byte, i int, old, want *Chan) error {
	if n := f.count(img); i >= n && want != nil {
		return fmt.Errorf("la memoria di questa radio ha %d canali", n)
	}
	return nil
}

// fagciAuto: nome dato a un canale senza nome (la frequenza, al massimo 9 caratteri)
func fagciAuto(rx int64) string {
	s := fmtMHz(rx)
	if len(s) > fagciNameMax {
		s = s[:fagciNameMax]
	}
	return s
}

func fagciName(r []byte) string {
	var name []byte
	for _, b := range r[10:20] {
		if b < 0x20 || b > 0x7E {
			break
		}
		name = append(name, b)
	}
	return string(name)
}

// frequenza (27 bit, unità di 10 Hz) e tipo del tono (4 bit) nei primi 4 byte, poi il codice
func fagciF(b []byte) (hz int64, typ, code byte) {
	v := binary.LittleEndian.Uint32(b)
	return int64(v&0x07FFFFFF) * 10, byte(v>>27) & 0x0F, b[4]
}

func fagciPutF(b []byte, hz int64, typ, code byte) {
	v := binary.LittleEndian.Uint32(b)&0x80000000 | uint32(typ&0x0F)<<27 | uint32(hz/10)&0x07FFFFFF
	binary.LittleEndian.PutUint32(b, v)
	b[4] = code
}

func (f fagciFam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(img, i)
	// CHANNELS_Existing del firmware: primo carattere del nome stampabile
	if r == nil || r[10] < 0x20 || r[10] > 0x7E {
		return nil
	}
	rx, rt, rc := fagciF(r[0:])
	tx, tt, tc := fagciF(r[5:])
	if rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx}
	if tx != 0 {
		c.Tx = tx
	}
	c.RxTone, c.TxTone = uvkxToneDec(rt, rc), uvkxToneDec(tt, tc)
	c.Name = strings.TrimSpace(fagciName(r))
	if c.Name == fagciAuto(rx) {
		c.Name = ""
	}
	c.Power = m.powerSem(int(r[21] >> 6))
	bw := r[21] >> 4 & 3
	c.Narrow = bw == 1 || bw == 2
	c.Scan = r[20]&0x7F != 0
	return c
}

func (f fagciFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(img, i)
	if r == nil {
		if c == nil {
			return nil
		}
		return fmt.Errorf("la memoria di questa radio ha %d canali", f.count(img))
	}
	if c == nil {
		// CHANNELS_Delete del firmware: tutto a zero, potenza bassa, ricevitore automatico
		for k := range r {
			r[k] = 0
		}
		r[21] = 1 << 6
		r[22] = 3
		return nil
	}
	if c.Rx%10 != 0 || c.Rx < fagciMinHz || c.Rx > fagciMaxHz {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	if c.Tx%10 != 0 || c.Tx < fagciMinHz || c.Tx > fagciMaxHz {
		return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
	}
	rt, rc, err := uvkxToneEnc(c.RxTone)
	if err != nil {
		return err
	}
	tt, tc, err := uvkxToneEnc(c.TxTone)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(c.Name)
	for _, ch := range name {
		if ch < 0x20 || ch > 0x7E {
			return fmt.Errorf("carattere non ammesso nel nome")
		}
	}
	if len(name) > fagciNameMax {
		name = name[:fagciNameMax]
	}
	if name == "" {
		name = fagciAuto(c.Rx)
	}
	exists := r[10] >= 0x20 && r[10] <= 0x7E
	oldRx, _, _ := fagciF(r[0:])
	oldTx, _, _ := fagciF(r[5:])
	if !exists {
		for k := range r {
			r[k] = 0
		}
		// canale nuovo: FM (AM nella banda aerea), ricevitore scelto dalla radio
		if c.Rx >= 108000000 && c.Rx < 137000000 {
			r[21] = 1
		}
		r[22] = 3
		oldRx, oldTx = 0, 0
	}
	fagciPutF(r[0:], c.Rx, rt, rc)
	if exists && oldTx == 0 && oldRx == c.Rx && c.Tx == c.Rx {
		// canale già senza frequenza di trasmissione (trasmette su quella di ricezione): resta com'è
		fagciPutF(r[5:], 0, tt, tc)
	} else {
		fagciPutF(r[5:], c.Tx, tt, tc)
	}
	copy(r[10:20], make([]byte, 10))
	copy(r[10:20], name)
	if c.Scan {
		if r[20]&0x7F == 0 {
			r[20] |= 1
		}
	} else {
		r[20] &= 0x80
	}
	bw := r[21] >> 4 & 3
	if c.Narrow && bw != 1 && bw != 2 {
		bw = 1
	} else if !c.Narrow && (bw == 1 || bw == 2) {
		bw = 0
	}
	r[21] = byte(m.powerIndex(c.Power))<<6 | bw<<4 | r[21]&0x0F
	return nil
}

// ---------------------------------------------------------------- collegamento

// lettura con indirizzo a 32 bit (CMD_051B del firmware)
func fagciRead(s *uvk5Sess, off, n int) ([]byte, error) {
	cmd := []byte{0x1b, 0x05, 0x0c, 0x00, byte(off), byte(off >> 8), byte(off >> 16), byte(off >> 24), byte(n), 0, 0, 0, 0x6a, 0x39, 0x57, 0x64}
	if err := s.cmd(cmd); err != nil {
		return nil, err
	}
	rep, err := s.reply()
	if err != nil {
		return nil, err
	}
	if len(rep) < 12+n || rep[0] != 0x1c || rep[4] != byte(off) || rep[5] != byte(off>>8) || rep[6] != byte(off>>16) || rep[8] != byte(n) {
		return nil, rErr("lettura della memoria a 0x%05X incompleta", off)
	}
	return rep[12 : 12+n], nil
}

// scrittura con indirizzo a 32 bit (CMD_051D): la radio scrive a gruppi di 8 byte
func fagciWrite(s *uvk5Sess, off int, data []byte) error {
	if len(data)%8 != 0 {
		return fmt.Errorf("blocco di scrittura non multiplo di 8 byte")
	}
	cmd := []byte{0x1d, 0x05, byte(len(data) + 12), 0x00, byte(off), byte(off >> 8), byte(off >> 16), byte(off >> 24), byte(len(data)), 0, 0, 1,
		0x6a, 0x39, 0x57, 0x64}
	if err := s.cmd(append(cmd, data...)); err != nil {
		return err
	}
	rep, err := s.reply()
	if err != nil {
		return err
	}
	if len(rep) < 8 || rep[0] != 0x1e || rep[4] != byte(off) || rep[5] != byte(off>>8) || rep[6] != byte(off>>16) {
		return rErr("la radio non ha confermato la scrittura a 0x%05X", off)
	}
	return nil
}

// fagciOpen: collegamento, lettura della testa (impostazioni, VFO, preset) e
// controlli sulla disposizione della memoria
func (f fagciFam) open(ctx context.Context, m *genModel, port string) (Port, *uvk5Sess, string, []byte, error) {
	p, s, fw, err := uvk5xFam{}.connect(ctx, m, port)
	if err != nil {
		return nil, nil, fw, nil, err
	}
	fail := func(e error) (Port, *uvk5Sess, string, []byte, error) {
		p.Close()
		if ctx.Err() != nil {
			return nil, nil, fw, nil, errCancelled
		}
		return nil, nil, fw, nil, e
	}
	var head []byte
	for a := 0; a < fagciHead; a += uvk5Block {
		k := fagciHead - a
		if k > uvk5Block {
			k = uvk5Block
		}
		b, err := fagciRead(s, a, k)
		if err != nil {
			if a == 0 {
				return fail(rErr("questa versione del firmware R3b0rn non risponde ai comandi conosciuti: per sicurezza non si legge e non si scrive"))
			}
			return fail(err)
		}
		head = append(head, b...)
	}
	if head[0]&7 < 2 {
		return fail(rErr("la memoria della radio non è stata preparata dal firmware R3b0rn: fai prima il «reset» dal menu della radio"))
	}
	if head[9] != fagciPresets {
		return fail(rErr("questa versione del firmware R3b0rn ha la memoria disposta in un altro modo (%d preset invece di %d): per sicurezza non si legge e non si scrive",
			head[9], fagciPresets))
	}
	return p, s, fw, head, nil
}

func (f fagciFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, 1, "Collegamento con la radio…")
	p, s, fw, head, err := f.open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer p.Close()
	img := fagciNew(head, int(head[0]&7))
	n := f.count(img)
	end := int(binary.LittleEndian.Uint32(img[20:]))
	lo, total := end-n*fagciRec, n*fagciRec
	for done := 0; done < total; {
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		prog(done, total, "Lettura…")
		k := total - done
		if k > uvk5Block {
			k = uvk5Block
		}
		b, err := fagciRead(s, lo+done, k)
		if err != nil {
			return nil, "", err
		}
		copy(img[fagciChanAt+done:], b)
		done += k
	}
	prog(total, total, "Lettura completata")
	return img, fmt.Sprintf("Firmware della radio: %s; memoria di %d KiB, %d canali", fw, fagciSizes[head[0]&7]/1024, n), nil
}

func (f fagciFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	n := f.count(img)
	if n == 0 {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	prog(0, 1, "Collegamento con la radio…")
	p, s, fw, head, err := f.open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer p.Close()
	if fagciSizes[head[0]&7] != int(binary.LittleEndian.Uint32(img[16:])) {
		return "", rErr("la memoria di questa radio (%d KiB) è diversa da quella da cui vengono i canali (%d KiB): leggi prima questa radio",
			fagciSizes[head[0]&7]/1024, binary.LittleEndian.Uint32(img[16:])/1024)
	}
	end := int(binary.LittleEndian.Uint32(img[20:]))
	lo, total := end-n*fagciRec, n*fagciRec
	data := img[fagciChanAt : fagciChanAt+total]
	// solo la zona dei canali, a blocchi multipli di 8 byte: l'ultimo blocco parte
	// un po' prima e ripete byte già scritti (la radio non riscrive i byte uguali)
	for done := 0; done < total; {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(done, total, "Scrittura…")
		k := total - done
		if k > uvk5Block {
			k = uvk5Block
		}
		at := done
		if k%8 != 0 {
			k = (k + 7) &^ 7
			at = total - k
		}
		if err := fagciWrite(s, lo+at, data[at:at+k]); err != nil {
			return "", err
		}
		done = at + k
	}
	_ = s.cmd([]byte{0xdd, 0x05, 0x00, 0x00})
	prog(total, total, "Scrittura completata")
	return "Firmware della radio: " + fw, nil
}

// fagciBlank: immagine di una radio appena preparata dal firmware (per le prove e
// per la radio simulata): EEPROM del tipo indicato, 40 preset, canali cancellati
func fagciBlank(typ int) []byte {
	head := make([]byte, fagciHead)
	head[0] = byte(typ & 7)
	head[9] = fagciPresets
	img := fagciNew(head, typ)
	f := fagciFam{}
	for i := 0; i < f.count(img); i++ {
		_ = f.Set(nil, img, i, nil)
	}
	return img
}
