// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Aggiornamento del firmware delle Quansheng UV-K5, UV-K5(8), UV-K6 e UV-5R Plus
// (4.1, file .bin pubblicati da Quansheng, es. k5_v2.01.26_publish.bin).
//
// Il protocollo del caricatore (bootloader) è documentato dai programmi liberi
// k5prog (Jacek Lipkowski SQ5BPF, GNU GPL v3) e uvtools (egzumer), che lo hanno
// ricostruito dal programma di aggiornamento di Quansheng; qui è riscritto da
// zero e confrontato con il codice di uvtools (verifica/k5).
//
// File .bin: firmware cifrato con uno XOR di 128 byte, con in fondo il CRC16
// XModem (2 byte, little endian) del contenuto cifrato. Decifrato, contiene a
// 0x2000 la versione (16 byte, es. "2.01.26"), che non viene scritta nella radio:
// la radio riceve il firmware senza quei 16 byte (al massimo 0xEFFF byte; da
// 0xF000 c'è il caricatore, mai toccato).
//
// Caricatore (radio spenta, accesa tenendo premuto il PTT: torcia accesa e
// schermo spento), 38400 baud, pacchetti AB CD lunghezza(2, LE) dati+CRC16
// cifrati con lo XOR di 16 byte del protocollo di programmazione, DC BA:
//   radio → 0x0518 ripetuto (identificativo e versione del caricatore a 0x14,
//           il cui primo carattere indica il tipo di radio: 2 UV-K5, 3 UV-K5(8)
//           e UV-K6, 4 UV-5R Plus);
//   PC    → 0x0530 con la versione del firmware (16 byte); radio → 0x0518;
//   PC    → 0x0519 per ogni blocco di 256 byte: 19 05 0C 01 8A 8D 9F 1D,
//           indirizzo (2, BE), fine arrotondata a 256 (2, BE), 01 00, 00 00, dati;
//   radio → 0x051A con lo stesso indirizzo nei byte 8-9.
// Il programma controlla che il primo carattere della versione del firmware sia
// quello del caricatore (come uvtools): un firmware della UV-K6 non viene inviato
// a una UV-K5. Le memorie dei canali e la taratura (EEPROM) non vengono toccate.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	k5FwVerOff = 0x2000
	k5FwVerLen = 16
	k5FwMax    = 0xEFFF // firmware (senza versione) al massimo; da 0xF000 il caricatore
	k5FwBlock  = 0x100
	// 4.4: radio con il caricatore 7.x (UV-K5 V3, UV-K1, Retevis RA79 V3:
	// microcontrollore PY32F071, programma da 0x08002800): firmware più grande
	k5v3FwMax = 0x18000
)

// fwKnownK5v3: file del costruttore per il caricatore 7.x provati (4.4)
var fwKnownK5v3 = map[string]string{
	"939f1cae8e26e10c748283e35bdc0e5ab06e24f87f2c35e33932099a6891d9cf": "Retevis RA79 V3, firmware 7.00.16 (RA79 V3.bin, retevis.com)",
}

// k5MaxFor: dimensione massima del firmware per il tipo di caricatore
func k5MaxFor(c byte) int {
	if c == '7' {
		return k5v3FwMax
	}
	return k5FwMax
}

// chiave XOR dei file .bin di Quansheng (128 byte)
var k5FwKey = []byte{
	0x47, 0x22, 0xc0, 0x52, 0x5d, 0x57, 0x48, 0x94, 0xb1, 0x60, 0x60, 0xdb, 0x6f, 0xe3, 0x4c, 0x7c,
	0xd8, 0x4a, 0xd6, 0x8b, 0x30, 0xec, 0x25, 0xe0, 0x4c, 0xd9, 0x00, 0x7f, 0xbf, 0xe3, 0x54, 0x05,
	0xe9, 0x3a, 0x97, 0x6b, 0xb0, 0x6e, 0x0c, 0xfb, 0xb1, 0x1a, 0xe2, 0xc9, 0xc1, 0x56, 0x47, 0xe9,
	0xba, 0xf1, 0x42, 0xb6, 0x67, 0x5f, 0x0f, 0x96, 0xf7, 0xc9, 0x3c, 0x84, 0x1b, 0x26, 0xe1, 0x4e,
	0x3b, 0x6f, 0x66, 0xe6, 0xa0, 0x6a, 0xb0, 0xbf, 0xc6, 0xa5, 0x70, 0x3a, 0xba, 0x18, 0x9e, 0x27,
	0x1a, 0x53, 0x5b, 0x71, 0xb1, 0x94, 0x1e, 0x18, 0xf2, 0xd6, 0x81, 0x02, 0x22, 0xfd, 0x5a, 0x28,
	0x91, 0xdb, 0xba, 0x5d, 0x64, 0xc6, 0xfe, 0x86, 0x83, 0x9c, 0x50, 0x1c, 0x73, 0x03, 0x11, 0xd6,
	0xaf, 0x30, 0xf4, 0x2c, 0x77, 0xb2, 0x7d, 0xbb, 0x3f, 0x29, 0x28, 0x57, 0x22, 0xd6, 0x92, 0x8b,
}

func k5FwXor(b []byte) []byte {
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = v ^ k5FwKey[i%len(k5FwKey)]
	}
	return out
}

// k5Unpack: file .bin → versione (16 byte) e firmware da inviare
func k5Unpack(b []byte) (ver, fw []byte, crcOK bool, err error) {
	if len(b) < k5FwVerOff+k5FwVerLen+2 {
		return nil, nil, false, fmt.Errorf("file troppo corto per un firmware della UV-K5 (%d byte)", len(b))
	}
	body := b[:len(b)-2]
	crcOK = binary.LittleEndian.Uint16(b[len(b)-2:]) == crc16XModem(body)
	dec := k5FwXor(body)
	ver = append([]byte{}, dec[k5FwVerOff:k5FwVerOff+k5FwVerLen]...)
	fw = append(append([]byte{}, dec[:k5FwVerOff]...), dec[k5FwVerOff+k5FwVerLen:]...)
	return ver, fw, crcOK, nil
}

// k5Pack: il contrario di k5Unpack (per le prove)
func k5Pack(ver, fw []byte) []byte {
	v := make([]byte, k5FwVerLen)
	copy(v, ver)
	dec := append(append(append([]byte{}, fw[:k5FwVerOff]...), v...), fw[k5FwVerOff:]...)
	enc := k5FwXor(dec)
	crc := make([]byte, 2)
	binary.LittleEndian.PutUint16(crc, crc16XModem(enc))
	return append(enc, crc...)
}

// k5VerString: versione leggibile («2.01.26»)
func k5VerString(v []byte) string {
	if i := bytes.IndexByte(v, 0); i >= 0 {
		v = v[:i]
	}
	return strings.TrimSpace(string(v))
}

func k5PlausibleVer(s string) bool {
	if s == "" || len(s) > 15 {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			return false
		}
	}
	return s[0] == '*' || (s[0] >= '0' && s[0] <= '9')
}

// k5IsPacked: il file è un firmware .bin della UV-K5 (CRC giusto e versione leggibile)?
func k5IsPacked(b []byte) bool {
	ver, _, ok, err := k5Unpack(b)
	return err == nil && ok && k5PlausibleVer(k5VerString(ver))
}

// k5Family: tipo di radio dal primo carattere della versione
func k5Family(c byte) string {
	switch c {
	case '2':
		return "UV-K5"
	case '3':
		return "UV-K5(8) / UV-K6"
	case '4':
		return "UV-5R Plus"
	case '7':
		return "UV-K5 V3 / UV-K1 / Retevis RA79 V3"
	case '*':
		return "qualsiasi (versione «*»)"
	}
	return ""
}

// analyzeK5 controlla un file .bin della UV-K5 e lo prepara per l'invio
func analyzeK5(name, origin string, raw []byte) *fwImage {
	img, info, add := newInfo("k5", "Quansheng UV-K5 / UV-K5(8) / UV-K6 / UV-5R Plus", "UVK5", name, origin, raw)
	ver, fw, crcOK, err := k5Unpack(raw)
	if err != nil {
		add("bad", "%v", err)
		return img
	}
	if !crcOK {
		add("bad", "Somma di controllo (CRC) del file errata: il file è danneggiato o non è un firmware .bin di Quansheng")
		return img
	}
	add("ok", "Somma di controllo (CRC16) del file corretta")
	vs := k5VerString(ver)
	if !k5PlausibleVer(vs) {
		add("bad", "Versione del firmware non leggibile: non è un firmware della UV-K5")
		return img
	}
	if len(fw) > k5MaxFor(vs[0]) {
		add("bad", "Firmware troppo grande (%s): il caricatore accetta al massimo %s", fmtBytes(len(fw)), fmtBytes(k5MaxFor(vs[0])))
		return img
	}
	if vs[0] == '7' {
		// caricatore 7.x: programma per PY32F071 (stack in RAM, avvio dopo il
		// caricatore a 0x08002800)
		sp, rv := binary.LittleEndian.Uint32(fw), binary.LittleEndian.Uint32(fw[4:])
		if sp>>16 != 0x2000 || rv&^0xFFFF != 0x08000000 || rv&^1 < 0x08002800 {
			add("bad", "Il contenuto del file non è un programma per le radio con il caricatore 7.x (avvio a 0x%08X): non è un firmware utilizzabile", rv)
			return img
		}
		add("ok", "Programma per il microcontrollore PY32F071 (avvio a 0x%08X, dopo il caricatore)", rv&^1)
	}
	info.Version = vs
	fam := k5Family(vs[0])
	if fam == "" {
		add("warn", "Versione «%s»: tipo di radio non riconosciuto (2 = UV-K5, 3 = UV-K5(8)/UV-K6, 4 = UV-5R Plus, 7 = UV-K5 V3 / UV-K1)", vs)
		info.Warnings = true
	} else {
		info.Target = "Quansheng " + fam
		if vs[0] == '*' {
			info.Target = "Quansheng UV-K5 / UV-K6 / UV-5R Plus"
			add("warn", "Versione «*»: il caricatore la accetta su tutte le radio, anche se il firmware non è per quella collegata")
			info.Warnings = true
		}
		add("ok", "Firmware versione %s per %s (%s, %d blocchi da 256 byte)", vs, fam, fmtBytes(len(fw)), (len(fw)+k5FwBlock-1)/k5FwBlock)
	}
	low := strings.ToLower(name)
	if !strings.Contains(low, "publish") && !strings.Contains(low, "k5") && !strings.Contains(low, "k6") && !strings.Contains(low, "5r") &&
		!strings.Contains(low, "k1") && !strings.Contains(low, "ra79") {
		add("warn", "Il nome del file non è quello dei firmware di Quansheng (es. k5_v2.01.26_publish.bin): verifica che sia quello giusto")
		info.Warnings = true
	}
	add("ok", "Prima dell'invio il programma controlla che il caricatore della radio sia del tipo indicato dalla versione: altrimenti la radio non viene toccata")
	if vs[0] == '7' {
		add("ok", "Caricatore 7.x: il firmware %s va solo sulle radio con il caricatore %s.xx (es. UV-K5 V3 e Retevis RA79 V3: 7.00; UV-K1: 7.02 e 7.03); le UV-K5 con il caricatore 2.x o 5.x vengono rifiutate", vs, vs[:min(4, len(vs))])
	}
	if k, ok := fwKnownK5v3[info.SHA256]; ok {
		info.Known = k
		add("ok", "File conosciuto: %s", k)
	}
	// 4.9: canali e impostazioni dei firmware liberi (schede egzumer, F4HWN, LOSEHU)
	up := strings.ToUpper(vs)
	switch {
	case strings.Contains(up, "F4HWN"):
		add("ok", "Firmware libero F4HWN: dopo l'aggiornamento canali e impostazioni si leggono con la scheda «Quansheng UV-K5 (firmware F4HWN)»")
		img.Info.Models = []string{"qs-uvk5-f4hwn"}
	case strings.Contains(up, "EGZUMER"):
		add("ok", "Firmware libero egzumer: dopo l'aggiornamento canali e impostazioni si leggono con la scheda «Quansheng UV-K5 (firmware egzumer)»")
		img.Info.Models = []string{"qs-uvk5-egz"}
	case strings.Contains(up, "LOSEHU") || strings.Contains(up, "UVE") || strings.Contains(up, "LTS"):
		add("ok", "Firmware cinese ampliato: dopo l'aggiornamento i canali si leggono con la scheda «Quansheng UV-K5 (firmware LOSEHU, LTS, UVE)» se la versione è tra quelle gestite")
		img.Info.Models = []string{"qs-uvk5-losehu"}
	case vs[0] == '*':
		add("warn", "Firmware non di Quansheng: i canali si leggono solo con le schede dei firmware liberi gestiti (egzumer, F4HWN, LOSEHU); con altri firmware il programma, come CHIRP, non li legge")
		info.Warnings = true
		img.Info.Models = []string{"qs-uvk5"}
	default:
		img.Info.Models = []string{"qs-uvk5"}
	}
	img.Dec = fw
	info.Hw = "caricatore " + strings.TrimSuffix(fam, " (versione «*»)")
	if vs[0] == '7' {
		img.Info.Models = []string{"rt-ra79", "qs-uvk1"}
	}
	img.Ver = ver
	info.Valid = true
	return img
}

// ---------------------------------------------------------------- protocollo

func k5BootPacket(data []byte) []byte {
	crc := make([]byte, 2)
	binary.LittleEndian.PutUint16(crc, crc16XModem(data))
	out := []byte{0xAB, 0xCD, byte(len(data)), byte(len(data) >> 8)}
	out = append(out, uvk5Xor(append(append([]byte{}, data...), crc...))...)
	return append(out, 0xDC, 0xBA)
}

// k5ReadFrame: prossimo pacchetto della radio (dati decifrati, senza CRC)
func k5ReadFrame(ctx context.Context, p Port, timeout time.Duration) ([]byte, error) {
	end := time.Now().Add(timeout)
	b := make([]byte, 1)
	for {
		left := time.Until(end)
		if left <= 0 {
			return nil, errTimeout
		}
		if err := p.ReadFull(ctx, b, left); err != nil {
			return nil, err
		}
		if b[0] != 0xAB {
			continue
		}
		if err := p.ReadFull(ctx, b, time.Second); err != nil {
			return nil, err
		}
		if b[0] != 0xCD {
			continue
		}
		h := make([]byte, 2)
		if err := p.ReadFull(ctx, h, time.Second); err != nil {
			return nil, err
		}
		n := int(binary.LittleEndian.Uint16(h))
		if n > 1024 {
			continue
		}
		d := make([]byte, n+4)
		if err := p.ReadFull(ctx, d, time.Second); err != nil {
			return nil, err
		}
		if d[n+2] != 0xDC || d[n+3] != 0xBA {
			continue
		}
		return uvk5Xor(d[:n+2])[:n], nil
	}
}

// k5BootVersion: versione del caricatore nel pacchetto 0x0518 (a 0x14)
func k5BootVersion(b []byte) string {
	var s []byte
	for i := 0x14; i < len(b) && i < 0x14+16; i++ {
		if b[i] < 0x20 || b[i] > 0x7E {
			break
		}
		s = append(s, b[i])
	}
	return string(s)
}

var (
	k5BootWait  = 30 * time.Second // attesa del caricatore
	k5BlockWait = 3 * time.Second
)

// k5Update invia il firmware alla radio nel caricatore. started è chiamata
// prima del primo blocco (da lì la radio va riaggiornata se si interrompe).
func k5Update(ctx context.Context, port string, ver, fw []byte, prog Progress, logf func(string), started func()) (int64, error) {
	if len(fw) == 0 || len(ver) == 0 || len(fw) > k5MaxFor(ver[0]) {
		return 0, fmt.Errorf("dimensione del firmware non valida")
	}
	p, err := openPort(port, 38400)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	logf("In attesa della radio nel caricatore (radio spenta, accesa tenendo premuto il PTT: torcia accesa, schermo spento)…")
	var beacon []byte
	end := time.Now().Add(k5BootWait)
	for beacon == nil {
		if ctx.Err() != nil {
			return 0, errCancelled
		}
		if time.Now().After(end) {
			return 0, rErr("la radio non si è presentata nel caricatore entro %d secondi: spegnila, tieni premuto il PTT e riaccendila (la torcia si accende, lo schermo resta spento)", int(k5BootWait.Seconds()))
		}
		f, err := k5ReadFrame(ctx, p, time.Second)
		if err == errCancelled {
			return 0, err
		}
		if err == nil && len(f) >= 0x14 && f[0] == 0x18 && f[1] == 0x05 {
			beacon = f
		}
	}
	bl := k5BootVersion(beacon)
	logf(fmt.Sprintf("Caricatore della radio: versione %s", bl))
	if ver[0] == '7' || strings.HasPrefix(bl, "7") {
		return k5v3Update(ctx, p, bl, ver, fw, prog, logf, started)
	}
	v := make([]byte, k5FwVerLen)
	copy(v, ver)
	if err := p.Write(k5BootPacket(append([]byte{0x30, 0x05, k5FwVerLen, 0x00}, v...))); err != nil {
		return 0, err
	}
	var resp []byte
	end = time.Now().Add(5 * time.Second)
	for resp == nil {
		if time.Now().After(end) {
			return 0, rErr("la radio non ha risposto alla versione del firmware")
		}
		f, err := k5ReadFrame(ctx, p, time.Second)
		if err == errCancelled {
			return 0, err
		}
		if err == nil && len(f) > 0x14 && f[0] == 0x18 {
			resp = f
		}
	}
	if v[0] != '*' && resp[0x14] != v[0] {
		return 0, rErr("il firmware (versione %s, per %s) non è per questa radio: il suo caricatore è il %s (%s). La radio non è stata modificata",
			k5VerString(v), k5Family(v[0]), bl, k5Family(resp[0x14]))
	}
	final := (len(fw) + 0xFF) &^ 0xFF
	started()
	var sent int64
	for a := 0; a < len(fw); a += k5FwBlock {
		if ctx.Err() != nil {
			return sent, errCancelled
		}
		prog(a, len(fw), "Invio del firmware…")
		blk := make([]byte, k5FwBlock)
		copy(blk, fw[a:])
		cmd := append([]byte{0x19, 0x05, 0x0C, 0x01, 0x8A, 0x8D, 0x9F, 0x1D,
			byte(a >> 8), byte(a), byte(final >> 8), 0x00, 0x01, 0x00, 0x00, 0x00}, blk...)
		if err := p.Write(k5BootPacket(cmd)); err != nil {
			return sent, err
		}
		ok := false
		bend := time.Now().Add(k5BlockWait)
		for !ok {
			if time.Now().After(bend) {
				return sent, rErr("nessuna conferma della radio per il blocco 0x%04X", a)
			}
			f, err := k5ReadFrame(ctx, p, k5BlockWait)
			if err == errCancelled {
				return sent, err
			}
			if err != nil {
				continue
			}
			if f[0] == 0x18 {
				continue // la radio ripete ancora il suo annuncio (succede dopo il primo blocco)
			}
			if len(f) >= 10 && f[0] == 0x1A && f[8] == byte(a>>8) && f[9] == byte(a) {
				ok = true
			}
		}
		sent += int64(min(k5FwBlock, len(fw)-a))
	}
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi", (len(fw)+k5FwBlock-1)/k5FwBlock))
	return sent, nil
}

// ---------------------------------------------------------------- caricatore 7.x

// k5v3Update (4.4): radio con il caricatore 7.x (UV-K5 V3, UV-K1, Retevis RA79 V3).
// Stessi pacchetti della UV-K5 ma, dopo l'annuncio 0x0518, il PC risponde tre
// volte con 0x0530 e le prime 4 lettere della versione del CARICATORE (non del
// firmware) e il caricatore non controlla il firmware: il controllo lo fa qui il
// programma (firmware 7.00.xx solo su caricatore 7.00.xx, ecc.; caricatori 2.x e
// 5.x delle UV-K5 V1/V2 rifiutati: firmware di un altro microcontrollore). Blocchi
// 0x0519: ora (4 byte), numero del blocco e numero dei blocchi (2 + 2 byte), 4
// byte a zero, 256 byte; conferma 0x051A con il numero del blocco e l'esito.
// Protocollo come pubblicato da quansheng_toolkit (kradio/dfu.py) e Multi-UVTools
// (js/flash-k1.js), riscritto da zero.
var k5v3Rounds = 3

func k5v3Update(ctx context.Context, p Port, bl string, ver, fw []byte, prog Progress, logf func(string), started func()) (int64, error) {
	vs := k5VerString(ver)
	if bl == "" {
		return 0, rErr("la radio non ha comunicato la versione del caricatore: la radio non è stata modificata")
	}
	switch {
	case !strings.HasPrefix(bl, "7"):
		return 0, rErr("il firmware %s è per le radio con il caricatore 7.x (UV-K5 V3, UV-K1, RA79 V3), ma il caricatore di questa radio è il %s (UV-K5 / UV-K6 / UV-5R Plus, altro microcontrollore): non è per questa radio. La radio non è stata modificata",
			vs, bl)
	case len(vs) < 4 || vs[0] != '7':
		return 0, rErr("il caricatore di questa radio è il %s (UV-K5 V3 / UV-K1): il firmware %s non è per questa radio. La radio non è stata modificata", bl, vs)
	case len(bl) < 4 || bl[:4] != vs[:4]:
		return 0, rErr("il firmware %s è per le radio con il caricatore %s.xx, ma quello di questa radio è il %s: non è per questa radio. La radio non è stata modificata", vs, vs[:4], bl)
	}
	// presentazione: una risposta a ciascuno dei prossimi annunci
	b4 := make([]byte, 4)
	copy(b4, bl)
	sent := 0
	end := time.Now().Add(5 * time.Second)
	for sent < k5v3Rounds {
		if ctx.Err() != nil {
			return 0, errCancelled
		}
		if time.Now().After(end) {
			return 0, rErr("la radio non ha accettato la presentazione del caricatore (%d di %d): la radio non è stata modificata", sent, k5v3Rounds)
		}
		f, err := k5ReadFrame(ctx, p, time.Second)
		if err == errCancelled {
			return 0, err
		}
		if err != nil || len(f) < 2 || f[0] != 0x18 || f[1] != 0x05 {
			continue
		}
		if err := p.Write(k5BootPacket(append([]byte{0x30, 0x05, 0x04, 0x00}, b4...))); err != nil {
			return 0, err
		}
		sent++
	}
	// annunci rimasti in arrivo
	for {
		if _, err := k5ReadFrame(ctx, p, 200*time.Millisecond); err != nil {
			if err == errCancelled {
				return 0, err
			}
			break
		}
	}
	n := (len(fw) + k5FwBlock - 1) / k5FwBlock
	ts := uint32(time.Now().Unix())
	started()
	var done int64
	for i := 0; i < n; i++ {
		prog(i*k5FwBlock, len(fw), "Invio del firmware…")
		cmd := make([]byte, 16+k5FwBlock)
		copy(cmd, []byte{0x19, 0x05, 0x0C, 0x01})
		binary.LittleEndian.PutUint32(cmd[4:], ts)
		binary.LittleEndian.PutUint16(cmd[8:], uint16(i))
		binary.LittleEndian.PutUint16(cmd[10:], uint16(n))
		copy(cmd[16:], fw[i*k5FwBlock:min((i+1)*k5FwBlock, len(fw))])
		ok := false
		for try := 0; try <= 3 && !ok; try++ {
			if ctx.Err() != nil {
				return done, errCancelled
			}
			if try > 0 {
				logf(fmt.Sprintf("Blocco %d di %d: nuovo tentativo (%d)", i+1, n, try))
			}
			if err := p.Write(k5BootPacket(cmd)); err != nil {
				return done, err
			}
			bend := time.Now().Add(k5BlockWait)
			for time.Now().Before(bend) {
				f, err := k5ReadFrame(ctx, p, time.Until(bend))
				if err == errCancelled {
					return done, err
				}
				if err != nil || len(f) < 12 || f[0] != 0x1A || f[1] != 0x05 || int(binary.LittleEndian.Uint16(f[8:])) != i {
					continue
				}
				if e := binary.LittleEndian.Uint16(f[10:]); e != 0 {
					logf(fmt.Sprintf("Blocco %d di %d: errore 0x%04X della radio", i+1, n, e))
					break
				}
				ok = true
				break
			}
		}
		if !ok {
			return done, rErr("la radio non ha confermato il blocco %d di %d", i+1, n)
		}
		done = int64(min((i+1)*k5FwBlock, len(fw)))
	}
	prog(len(fw), len(fw), "Firmware inviato")
	logf(fmt.Sprintf("Firmware inviato: %d blocchi da 256 byte, tutti confermati dalla radio", n))
	return done, nil
}
