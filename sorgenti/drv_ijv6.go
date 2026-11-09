// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.13.1: firmware IJV V6 (UV-K5 V3, UV-K1). La memoria è quella della FW5
// (area fissa a 0x000000, banco n a 0x020000 + n·0x4000, «CFG5» a 0x1550), ma il
// collegamento è il «collegamento dati» della V6, ricavato dall'analisi di
// RadioManager 4.3.2 (V6DataLink) e scritto da zero:
//   - DTR e RTS bassi, 38400 baud;
//   - richiesta «A5 5A cmd seq indirizzo32 lunghezza16 dati CRC32» (CRC-32 dei
//     byte da cmd alla fine dei dati); risposta di 12 byte «5A A5 stato seq
//     valore32 CRC32» con lo stesso seq;
//   - 8 informazioni (modello, variante, servizi, collegamento dati presente),
//     18 versione, 9 apertura (numero delle richieste che seguono), 10 lettura
//     (al massimo 256 byte: dopo la risposta arrivano i dati e il loro CRC32),
//     11 pagina di 256 byte da scrivere, 12 conferma del settore di 4 KiB,
//     5 fine, 19 riavvio;
//   - si scrive solo a settori interi di 4 KiB: si legge il settore, si cambiano
//     solo i byte permessi, si scrive e si rilegge per verifica.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	ijv6CmdInfo = 8
	ijv6Open    = 9
	ijv6Read    = 10
	ijv6Stage   = 11
	ijv6Commit  = 12
	ijv6End     = 5
	ijv6Version = 18
	ijv6Reset   = 19
	ijv6Sector  = 0x1000
	ijv6Page    = 0x100
)

// zone che la radio tiene per sé (RadioOwned di RadioManager): non si
// confrontano e si lasciano come le ha la radio
var ijv6RadioOwned = [][2]int{{0x1C10, 0x1C18}, {0x1D00, 0x1E00}}

type ijv6Link struct {
	ctx    context.Context
	p      Port
	seq    byte
	open   bool
	done   int
	total  int
	toOpen int
	wrote  bool
}

type ijv6Info struct {
	model, variant string
	services       int
	dataLink       bool
	version        string
}

// reply: risposta di 12 byte con il seq atteso (si scartano i byte estranei)
func (l *ijv6Link) reply(seq byte, timeout time.Duration) (byte, uint32, error) {
	deadline := time.Now().Add(timeout)
	one := make([]byte, 1)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return 0, 0, errTimeout
		}
		if err := l.p.ReadFull(l.ctx, one, left); err != nil {
			return 0, 0, err
		}
		if one[0] != 0x5A {
			continue
		}
		if err := l.p.ReadFull(l.ctx, one, time.Until(deadline)); err != nil {
			return 0, 0, err
		}
		if one[0] != 0xA5 {
			continue
		}
		b := make([]byte, 10)
		if err := l.p.ReadFull(l.ctx, b, time.Until(deadline)); err != nil {
			return 0, 0, err
		}
		if crc32IEEE(b[:6]) != binary.LittleEndian.Uint32(b[6:]) || b[1] != seq {
			continue
		}
		return b[0], binary.LittleEndian.Uint32(b[2:]), nil
	}
}

// ask: richiesta con risposta (si ripete la stessa richiesta se la radio non
// risponde o segnala un CRC sbagliato, stato 1)
func (l *ijv6Link) ask(cmd byte, addr uint32, data []byte, timeout time.Duration, tries int) (byte, uint32, error) {
	if l.ctx.Err() != nil {
		return 0, 0, errCancelled
	}
	if l.open {
		l.done++
	}
	l.seq++
	req := ijvV6Request(cmd, l.seq, addr, data)
	var err error
	for t := 0; t < tries; t++ {
		if err = l.p.Write(req); err != nil {
			return 0, 0, err
		}
		var st byte
		var v uint32
		st, v, err = l.reply(l.seq, timeout)
		if err == nil && st != 1 {
			return st, v, nil
		}
	}
	if err == nil {
		err = rErr("la radio V6 segnala dati ricevuti male")
	}
	return 0, 0, err
}

func (l *ijv6Link) probe() (*ijv6Info, error) {
	st, v, err := l.ask(ijv6CmdInfo, 0, nil, 800*time.Millisecond, 2)
	if err != nil || st != 0 {
		return nil, rErr("la radio non risponde al collegamento della V6")
	}
	b := v >> 24
	in := &ijv6Info{services: int((v >> 8) & 0xFF), dataLink: b&0x80 != 0}
	in.model = map[uint32]string{1: "UV-K5 V3", 2: "UV-K1", 3: "UV-K10"}[(b>>4)&3]
	if in.model == "" {
		in.model = "UV-K5"
	}
	in.variant = map[uint32]string{0: "standard", 1: "si4732", 2: "si4732v50", 3: "iotcu"}[b&15]
	if in.variant == "" {
		in.variant = fmt.Sprintf("variante %d", b&15)
	}
	in.version = "6.0"
	if in.services >= 24 {
		if st, v, err := l.ask(ijv6Version, 0, nil, 800*time.Millisecond, 1); err == nil && st == 0 {
			in.version = fmt.Sprintf("%d.%d.%d", (v>>16)&0xFF, (v>>8)&0xFF, v&0xFF)
		}
	}
	return in, nil
}

// openFor: la radio vuole sapere quante richieste seguono (Open/Announce di
// RadioManager); se le richieste annunciate finiscono si annunciano le altre
func (l *ijv6Link) openFor(n int) error {
	if n <= 0 {
		return nil
	}
	var count int
	if !l.open {
		count = max(l.toOpen, n+1)
	} else if l.done+1+n+1 > l.total {
		// una richiesta per la nuova apertura, n richieste e una di margine
		// (Announce di RadioManager)
		count = l.done + 1 + n + 1
	} else {
		return nil
	}
	st, _, err := l.ask(ijv6Open, uint32(count), nil, 3*time.Second, 3)
	if err != nil {
		return rErr("la radio V6 ha smesso di rispondere")
	}
	if st != 0 {
		return rErr("il firmware V6 sulla radio non ha ancora il collegamento dati: aggiornala con FLASH V6 di RadioManager")
	}
	if !l.open {
		l.done, l.toOpen = 0, 0
	}
	l.total = count
	l.open = true
	return nil
}

func ijv6Reads(n int) int { return (n + ijv6Page - 1) / ijv6Page }

// read: lettura a blocchi di 256 byte (dati seguiti dal loro CRC32)
func (l *ijv6Link) read(addr, n int) ([]byte, error) {
	if err := l.openFor(ijv6Reads(n)); err != nil {
		return nil, err
	}
	out := make([]byte, 0, n)
	for len(out) < n {
		k := min(ijv6Page, n-len(out))
		a := addr + len(out)
		var got []byte
		for t := 0; t < 3 && got == nil; t++ {
			st, v, err := l.ask(ijv6Read, uint32(a), []byte{byte(k), byte(k >> 8)}, 3*time.Second, 3)
			if err != nil {
				return nil, rErr("la radio V6 ha smesso di rispondere (lettura a 0x%06X)", a)
			}
			if st != 0 || int(v) != k {
				return nil, rErr("la radio V6 non ha letto la memoria a 0x%06X (stato %d)", a, st)
			}
			b := make([]byte, k+4)
			if err := l.p.ReadFull(l.ctx, b, 3*time.Second); err != nil {
				_ = l.p.Purge()
				continue
			}
			if crc32IEEE(b[:k]) == binary.LittleEndian.Uint32(b[k:]) {
				got = b[:k]
			}
		}
		if got == nil {
			return nil, rErr("lettura della memoria a 0x%06X non riuscita", a)
		}
		out = append(out, got...)
	}
	return out, nil
}

// writeSector: settore di 4 KiB in 16 pagine e conferma (SendSector)
func (l *ijv6Link) writeSector(addr int, data []byte) error {
	if err := l.openFor(ijv6Sector/ijv6Page + 1); err != nil {
		return err
	}
	for try := 0; ; try++ {
		var st byte = 0xFF
		for off := 0; off < ijv6Sector; off += ijv6Page {
			s, _, err := l.ask(ijv6Stage, uint32(addr+off), data[off:off+ijv6Page], 10*time.Second, 3)
			if err != nil {
				st = 0xFF
				break
			}
			if st = s; st != 0 {
				break
			}
		}
		if st == 0 {
			s, _, err := l.ask(ijv6Commit, uint32(addr), nil, 10*time.Second, 1)
			if err != nil {
				st = 0xFF
			} else {
				st = s
			}
		}
		switch {
		case st == 0:
			l.wrote = true
			return nil
		case st == 3:
			l.wrote = true
			return rErr("la radio V6 non è riuscita a scrivere nella sua memoria (settore 0x%06X)", addr)
		case st == 2 && addr == 0x1000:
			return rErr("la radio V6 ha rifiutato le impostazioni: il blocco di configurazione non è valido")
		case st == 2:
			return rErr("la radio V6 ha rifiutato la scrittura a 0x%06X: aggiornala con FLASH V6 di RadioManager e riprova", addr)
		case try >= 1:
			if st == 0xFF {
				return rErr("la radio V6 ha smesso di rispondere (scrittura a 0x%06X)", addr)
			}
			return rErr("la radio V6 non è riuscita a scrivere nella sua memoria (settore 0x%06X)", addr)
		}
	}
}

// end: fine della sessione dati (5); dopo una scrittura RadioManager riavvia
// la radio (19) se il firmware lo permette (servizi dal 27)
func (l *ijv6Link) end(services int) {
	if !l.open {
		return
	}
	l.open = false
	if l.wrote {
		l.seq++
		_ = l.p.Write(ijvV6Request(ijv6End, l.seq, 0, nil))
		if services >= 27 {
			time.Sleep(200 * time.Millisecond)
			_ = l.p.Purge()
			_, _, _ = l.ask(ijv6Reset, 0, nil, 3*time.Second, 2)
		}
		return
	}
	_, _, _ = l.ask(ijv6End, 0, nil, 3*time.Second, 1)
}

func ijv6Connect(ctx context.Context, port string) (Port, *ijv6Link, *ijv6Info, error) {
	p, err := openPort(port, 38400)
	if err != nil {
		return nil, nil, nil, err
	}
	if x, ok := p.(interface{ SetDTR(bool) error }); ok {
		_ = x.SetDTR(false)
	}
	_ = setRTSOff(p)
	time.Sleep(120 * time.Millisecond)
	_ = p.Purge()
	l := &ijv6Link{ctx: ctx, p: p}
	in, err := l.probe()
	if err != nil {
		in, err = l.probe()
	}
	if err != nil {
		return p, nil, nil, err
	}
	if !in.dataLink {
		return p, nil, nil, rErr("il firmware V6 sulla radio non ha ancora il collegamento dati: aggiornala con FLASH V6 di RadioManager")
	}
	return p, l, in, nil
}

// v6Start: radio V6 con il collegamento dati; se non risponde, il saluto della
// UV-K5 dice quale scheda aprire
func (f ijvFam) v6Start(ctx context.Context, m *genModel, port string) (Port, *ijv6Link, *ijv6Info, error) {
	p, l, in, err := ijv6Connect(ctx, port)
	if err == nil {
		return p, l, in, nil
	}
	if p == nil {
		return nil, nil, nil, err
	}
	if x, ok := p.(interface{ SetDTR(bool) error }); ok {
		_ = x.SetDTR(true)
	}
	// non è una V6 (o non risponde): il saluto della UV-K5 dice quale firmware ha
	for _, plain := range []bool{true, false} {
		_ = p.Purge()
		s := &ijvSess{ctx: ctx, p: p, plain: plain}
		if fw, e := s.hello(); e == nil && fw != "" {
			p.Close()
			for _, o := range genModels {
				if op, ok := o.Opt.(*uvk5Opt); ok && o != m && op.approve(fw) {
					return nil, nil, nil, rErr("la radio ha il firmware %s: apri la scheda «%s» (questa scheda è per un altro firmware e non legge e non scrive)", fw, o.Name)
				}
			}
			return nil, nil, nil, rErr("il firmware della radio (%s) non è %s: per sicurezza non si legge e non si scrive", fw, "IJV V6")
		}
	}
	p.Close()
	return nil, nil, nil, genFail("Collegamento", err)
}

func (f ijvFam) v6Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	prog(0, ijvAppSize, "Collegamento con la radio…")
	p, l, in, err := f.v6Start(ctx, m, port)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	defer l.end(in.services)
	// annuncio di tutte le letture: area fissa e banco
	if err := l.openFor(ijv6Reads(ijvFixedEnd) + ijv6Reads(ijvBankSize)); err != nil {
		return nil, nil, "", err
	}
	data := make([]byte, 0, ijvAppSize)
	for a := 0; a < ijvFixedEnd; a += ijv6Page {
		prog(a, ijvAppSize, "Lettura…")
		b, err := l.read(a, ijv6Page)
		if err != nil {
			return nil, nil, "", err
		}
		data = append(data, b...)
	}
	bk, err := f.imgBank(data)
	if err != nil {
		return nil, nil, "", err
	}
	for a := 0; a < ijvBankSize; a += ijv6Page {
		prog(ijvFixedEnd+a, ijvAppSize, "Lettura…")
		b, err := l.read(ijvBankBase+bk*ijvBankSize+a, ijv6Page)
		if err != nil {
			return nil, nil, "", err
		}
		data = append(data, b...)
	}
	prog(ijvAppSize, ijvAppSize, "Lettura completata")
	fw := "IJV V6 " + in.model + " " + in.variant
	meta := map[string]interface{}{"uvk5_firmware": fw, "ijv_bank": bk}
	note := fmt.Sprintf("Firmware della radio: %s, versione %s — banco %d «%s»", fw, in.version, bk+1, ijvBankLabel(data, bk))
	return data, meta, note, nil
}

// ijv6Allowed: byte che si possono cambiare (impostazioni fino a 0x1540 e banco)
func ijv6Allowed(logical int) bool {
	return logical < 0x1540 || logical >= ijvFixedEnd
}

func ijv6Owned(phys int) bool {
	for _, r := range ijv6RadioOwned {
		if phys >= r[0] && phys < r[1] {
			return true
		}
	}
	return false
}

func (f ijvFam) v6Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	want, err := f.imgBank(img)
	if err != nil {
		return "", err
	}
	prog(0, ijvAppSize, "Collegamento con la radio…")
	p, l, in, err := f.v6Start(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer p.Close()
	defer l.end(in.services)
	cfg, err := l.read(ijvCfg5, 32)
	if err != nil {
		return "", err
	}
	bk, err := ijvCfgBank(cfg)
	if err != nil {
		return "", err
	}
	if bk != want {
		return "", rErr("sulla radio è scelto il banco %d, ma i canali da scrivere sono del banco %d: scegli sulla radio il banco %d "+
			"(o rileggi la radio). La radio non è stata modificata", bk+1, want+1, want+1)
	}
	// settori da controllare: area fissa (0x0000, 0x1000) e i 4 del banco
	type sect struct{ logical, phys int }
	secs := []sect{{0, 0}, {0x1000, 0x1000}}
	for k := 0; k < ijvBankSize; k += ijv6Sector {
		secs = append(secs, sect{ijvFixedEnd + k, ijvBankBase + bk*ijvBankSize + k})
	}
	wrote := 0
	for i, s := range secs {
		prog(i*ijv6Sector, len(secs)*ijv6Sector, "Scrittura…")
		cur, err := l.read(s.phys, ijv6Sector)
		if err != nil {
			return "", err
		}
		next := append([]byte{}, cur...)
		for k := 0; k < ijv6Sector; k++ {
			if ijv6Allowed(s.logical+k) && !ijv6Owned(s.phys+k) {
				next[k] = img[s.logical+k]
			}
		}
		if bytes.Equal(cur, next) {
			continue
		}
		if err := l.writeSector(s.phys, next); err != nil {
			return "", err
		}
		back, err := l.read(s.phys, ijv6Sector)
		if err != nil {
			return "", err
		}
		for k := 0; k < ijv6Sector; k++ {
			if back[k] != next[k] && !ijv6Owned(s.phys+k) {
				return "", rErr("verifica non riuscita: sulla radio V6 non c'è quello che è stato scritto (0x%06X)", s.phys+k)
			}
		}
		wrote++
	}
	prog(len(secs)*ijv6Sector, len(secs)*ijv6Sector, "Scrittura completata")
	return fmt.Sprintf("Firmware della radio: IJV V6 %s %s, versione %s — settori cambiati e scritti: %d", in.model, in.variant, in.version, wrote), nil
}
