// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Radtel RT-4D e RT-6D (DMR) e RT-1000 Pro (analogica): protocollo e memoria
// ricavati dai programmi del costruttore «RT-4D CPS V2.07» (260210), «DMR RT-6D
// V2.06» (261020) e «Radtel RT-1000 Pro CPS» (260806), tutti .NET, progetto
// «DM550_NOLCD», pubblicati sulla pagina dei download di radtels.com. Sono una
// generazione successiva a quella della Bajeton DM-5000 (drv_dm5000.go): stesso
// collegamento, canale e zona diversi.
//
// Protocollo (115200 baud, 8N1), uguale a quello della DM-5000:
//   «34 52 05 10 9B» → 06; lettura 'R' n(2) somma → 'R' n(2) + 1024 byte + somma;
//   scrittura comando n(2) + 1024 byte + somma → 06; fine «34 52 05 EE 79».
// Il programma scrive, nell'ordine: 0x90 blocco 0 (impostazioni, blocco letto 8),
// 0x91 blocchi 0-47 (canali, letti 16-63), 0x92 blocco 0 (VFO, letto 0x70), 0x93
// blocchi 0-127 (zone, letti 0x78-0xF7), poi contatti, gruppi, cifratura,
// messaggi, radio FM. Qui la scrittura si ferma dopo le zone: impostazioni e VFO
// vengono riscritti con i dati appena letti dalla radio (come fa il programma,
// byte 14 e 15 delle impostazioni a FF), il resto non si tocca.
// Prima di scrivere il programma legge il blocco 0 e il blocco 8 (password nei
// byte 0-11).
//
// Canale (48 byte; libero = tutto FF):
//   0  bit 7-6 tipo (0 digitale, 1 analogico; 2 e 3 = libero), bit 5-4 RX+TX /
//      solo RX / solo TX, bit 3 ID (radio / canale), bit 2 modo DMR, bit 1 slot,
//      bit 0 monitor digitale
//   1  bit 7-4 color code, bit 3-0 scrambler
//   2  bit 7-6 potenza (0 bassa, 1 alta), bit 5-0 TOT (indice)
//   3  bit 7 escluso dalla scansione, bit 6-5 priorità di chiamata, bit 4-3
//      priorità TX, bit 2-0 tono di coda
//   4  bit 7-6 banda (0 larga, 1 stretta), bit 5-4 FM / AM / SSB, bit 3-1 tipo di
//      subtono
//   5-8 RX e 9-12 TX in decine di Hz (little endian)
//   13-14 tono RX, 15-16 tono TX (come la DM-5000: 0x1000 | decimi di Hz,
//      0x2000 | codice DCS ottale, 0x3000 | codice invertito, 0 nessuno)
//   17-18 contatto, 19 lista dei gruppi, 20-21 cifratura, 22-25 ID del canale
//   (BCD), 26-29 codice di silenzio, 32-47 nome (GBK, FF in coda)
// Zona (520 byte, 250 zone): 0-1 e 2-3 canale scelto per A e per B, 4-19 nome;
// RT-1000 Pro: 0-1, 2-3, 4-5 canale scelto per A, B e C, 6-19 nome; da 20 fino a
// 250 canali (16 bit little endian, numero del canale da 0; 0xFFFF libero).
//
// Non certo, quindi non fatto: il significato esatto dei canali scelti per A/B/C
// (il programma li salva come numero di canale ma li rilegge come posizione
// nell'elenco): quei byte si conservano. RT-1000 Pro: il programma scrive
// l'esclusione dalla scansione nel bit 7 del byte 3 ma la rilegge dal bit 6:
// la scansione non si gestisce e il byte 3 dei canali esistenti si conserva.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	rt4dChSize   = 48
	rt4dSlots    = 1024
	rt4dChBlocks = 48  // blocchi dei canali: letti da 16, scritti con 0x91
	rt4dZnBlocks = 128 // blocchi delle zone: letti da 0x78, scritti con 0x93
	rt4dZoneSize = 520
	rt4dZones    = 250
	rt4dZoneMax  = 250
	rt4dZoneOff  = rt4dChBlocks * dm5kBlock
	rt4dSize     = (rt4dChBlocks + rt4dZnBlocks) * dm5kBlock

	rt4dBlkCfg  = 8
	rt4dBlkCh   = 16
	rt4dBlkVfo  = 0x70
	rt4dBlkZone = 0x78
)

type rt4dOpt struct {
	digital bool // canali DMR oltre agli analogici (RT-4D, RT-6D)
	zoneSel int  // canali scelti all'inizio della zona (2: A e B; 3: A, B e C)
}

func (o *rt4dOpt) zoneName() (off, n int) { return 2 * o.zoneSel, 20 - 2*o.zoneSel }

type rt4dFam struct{}

const rt4dNoteEnd = " Un canale nuovo viene aggiunto alla zona 1 se non è in nessuna zona; un canale cancellato viene tolto dalle zone. " +
	"Se sulla radio è impostata una password, toglila prima con il programma del costruttore. " +
	"«Identifica» non distingue questa radio dalle altre Radtel e Bajeton con lo stesso collegamento: scegli tu la scheda giusta."

func init() {
	for _, x := range []struct {
		id, model, rclass string
		opt               rt4dOpt
	}{
		{"rtl-rt4d", "RT-4D", "IU0VYT_Radtel_RT4D", rt4dOpt{digital: true, zoneSel: 2}},
		{"rtl-rt6d", "RT-6D", "IU0VYT_Radtel_RT6D", rt4dOpt{digital: true, zoneSel: 2}},
		{"rtl-rt1000pro", "RT-1000 Pro", "IU0VYT_Radtel_RT1000Pro", rt4dOpt{zoneSel: 3}},
	} {
		o := x.opt
		m := &genModel{ID: x.id, Vendor: "Radtel", Model: x.model, RClass: x.rclass, Name: "Radtel " + x.model,
			Slots: rt4dSlots, First: 1, NameMax: 16, Charset: charsetASCII,
			Powers: []genPower{{"Alta", 0, 0}, {"Bassa", 0, 2}},
			Bands:  []band{{18000000, 660000000}, {840000000, 1000000000}}, Narrow: true, HasScan: o.digital, NoTxOff: true, NoChirp: true,
			DCS: dtcsCodes, Baud: 115200, Fam: rt4dFam{}, Opt: &o}
		if o.digital {
			m.Digital = "DMR"
			m.Notes = "Radio DMR. Protocollo e memoria ricavati dai programmi del costruttore «RT-4D CPS» (versione 2.07) e «DMR RT-6D» (versione 2.06). " +
				"Qui si modificano frequenze, nome, potenza, banda, scansione e toni; le altre voci del canale restano quelle impostate " +
				"(un canale DMR nuovo prende i valori predefiniti del programma: primo contatto, slot 1, color code 1). " +
				"Si scrivono solo canali e zone: contatti, gruppi, messaggi e impostazioni della radio non si toccano." + rt4dNoteEnd
		} else {
			m.Notes = "Protocollo e memoria ricavati dal programma del costruttore «Radtel RT-1000 Pro CPS» (versione del 6 agosto 2026). " +
				"Qui si modificano frequenze, nome, potenza, banda e toni; le altre voci del canale, scansione compresa, restano quelle impostate. " +
				"Si scrivono solo canali e zone: radio FM e impostazioni della radio non si toccano." + rt4dNoteEnd
		}
		genRegister(m)
	}
}

func (rt4dFam) rec(img []byte, i int) []byte { return img[i*rt4dChSize : (i+1)*rt4dChSize] }

func (rt4dFam) used(o *rt4dOpt, r []byte) bool {
	t := r[0] >> 6
	if o.digital {
		return t <= 1
	}
	return t == 1
}

func (f rt4dFam) Get(m *genModel, img []byte, i int) *Chan {
	o := m.Opt.(*rt4dOpt)
	r := f.rec(img, i)
	if !f.used(o, r) {
		return nil
	}
	c := &Chan{Rx: int64(binary.LittleEndian.Uint32(r[5:])) * 10, Tx: int64(binary.LittleEndian.Uint32(r[9:])) * 10}
	c.Name = strings.TrimRight(string(trimBytes(r[32:48], 0xFF, 0x00)), " ")
	if (r[2]>>6)&3 == 1 {
		c.Power = m.powerSem(0)
	} else {
		c.Power = m.powerSem(1)
	}
	c.Scan = !m.HasScan || r[3]&0x80 == 0
	if r[0]>>6 == 0 {
		c.Dig = true
		return c
	}
	c.Narrow = (r[4]>>6)&3 == 1
	c.RxTone = dm5kToneDec(binary.LittleEndian.Uint16(r[13:]))
	c.TxTone = dm5kToneDec(binary.LittleEndian.Uint16(r[15:]))
	return c
}

func (f rt4dFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(img, i)...)
}

func (f rt4dFam) PutRec(m *genModel, img []byte, i int, raw []byte) { copy(f.rec(img, i), raw) }

func (f rt4dFam) Match(m *genModel, img []byte) bool { return false }

// ---- zone

func (rt4dFam) zone(img []byte, z int) []byte {
	o := rt4dZoneOff + z*rt4dZoneSize
	return img[o : o+rt4dZoneSize]
}

func (f rt4dFam) zoneMembers(img []byte, z int) []int {
	r := f.zone(img, z)
	var out []int
	for j := 0; j < rt4dZoneMax; j++ {
		if v := int(binary.LittleEndian.Uint16(r[20+2*j:])); v < rt4dSlots {
			out = append(out, v)
		}
	}
	return out
}

// elenco compatto, posizioni libere a FFFF; i canali scelti per A/B/C si
// conservano (a 0, come il programma, in una zona mai salvata)
func (f rt4dFam) setZoneMembers(o *rt4dOpt, img []byte, z int, mem []int) {
	r := f.zone(img, z)
	for j := 0; j < rt4dZoneMax; j++ {
		v := uint16(0xFFFF)
		if j < len(mem) {
			v = uint16(mem[j])
		}
		binary.LittleEndian.PutUint16(r[20+2*j:], v)
	}
	off, n := o.zoneName()
	if bytes.Equal(r[:20], bytes.Repeat([]byte{0xFF}, 20)) {
		// zona mai salvata: nome predefinito del programma («Zone1»)
		for k := 0; k < off; k++ {
			r[k] = 0
		}
		nm := []byte(fmt.Sprintf("Zone%d", z+1))
		copy(r[off:off+n], append(nm, bytes.Repeat([]byte{0xFF}, n-len(nm))...))
	}
}

func (f rt4dFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := m.Opt.(*rt4dOpt)
	r := f.rec(img, i)
	if c == nil {
		copy(r, bytes.Repeat([]byte{0xFF}, rt4dChSize))
		for z := 0; z < rt4dZones; z++ {
			mem := f.zoneMembers(img, z)
			if k := idxInt(mem, i); k >= 0 {
				f.setZoneMembers(o, img, z, append(mem[:k:k], mem[k+1:]...))
			}
		}
		return nil
	}
	tt, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	isNew := !f.used(o, r)
	typ := byte(1)
	if c.Dig && o.digital {
		typ = 0
	}
	oldRx := int64(binary.LittleEndian.Uint32(r[5:])) * 10
	if isNew || r[0]>>6 != typ {
		// valori predefiniti del programma per un canale nuovo: RX+TX, TOT 180 s
		// (indice 14), color code 1, tono di coda 1, ID del canale 1 (BCD), FM, larga
		copy(r, make([]byte, rt4dChSize))
		r[1], r[2], r[3] = 0x10, 0x0E, 0x01
		binary.LittleEndian.PutUint32(r[22:], 1)
		oldRx = -1
	}
	r[0] = r[0]&0x3F | typ<<6
	binary.LittleEndian.PutUint32(r[5:], uint32(c.Rx/10))
	binary.LittleEndian.PutUint32(r[9:], uint32(c.Tx/10))
	r[2] &= 0x3F
	if m.powerIndex(c.Power) == 0 {
		r[2] |= 1 << 6
	}
	if m.HasScan {
		r[3] &^= 0x80
		if !c.Scan {
			r[3] |= 0x80
		}
	}
	if typ == 1 {
		r[4] &= 0x3F
		if c.Narrow {
			r[4] |= 1 << 6
		}
		// FM / AM / SSB: in gamma aeronautica AM; altrimenti il modo resta se la
		// frequenza non cambia, FM se cambia
		switch {
		case c.Rx >= 108000000 && c.Rx <= 135999999:
			r[4] = r[4]&^0x30 | 1<<4
		case oldRx != c.Rx:
			r[4] &^= 0x30
		}
		binary.LittleEndian.PutUint16(r[13:], dm5kToneEnc(rt))
		binary.LittleEndian.PutUint16(r[15:], dm5kToneEnc(tt))
	}
	name := bytes.Repeat([]byte{0xFF}, 16)
	copy(name, c.Name)
	copy(r[32:], name)
	if isNew {
		in := false
		for z := 0; z < rt4dZones && !in; z++ {
			in = idxInt(f.zoneMembers(img, z), i) >= 0
		}
		if !in {
			mem := f.zoneMembers(img, 0)
			if len(mem) < rt4dZoneMax {
				f.setZoneMembers(o, img, 0, append(mem, i))
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

func (f rt4dFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	l, err := dm5kOpen(ctx, port)
	if err != nil {
		return nil, "", err
	}
	defer l.p.Close()
	prog(0, rt4dSize, "Collegamento con la radio…")
	if err := l.connect(); err != nil {
		return nil, "", genFail("Collegamento", err)
	}
	// come il programma: per primo il blocco 8 (password)
	b8, err := l.readBlock(rt4dBlkCfg)
	if err != nil {
		return nil, "", genFail("Lettura", err)
	}
	if err := dm5kCheckPwd(b8); err != nil {
		_ = l.p.Write(dm5kEnd)
		return nil, "", genFail("Lettura", err)
	}
	img := make([]byte, 0, rt4dSize)
	for _, a := range []struct {
		name     string
		first, n int
	}{{"canali", rt4dBlkCh, rt4dChBlocks}, {"zone", rt4dBlkZone, rt4dZnBlocks}} {
		for k := 0; k < a.n; k++ {
			if ctx.Err() != nil {
				return nil, "", errCancelled
			}
			prog(len(img), rt4dSize, "Lettura: "+a.name+"…")
			d, err := l.readBlock(a.first + k)
			if err != nil {
				return nil, "", genFail("Lettura", err)
			}
			img = append(img, d...)
		}
	}
	_ = l.p.Write(dm5kEnd)
	prog(rt4dSize, rt4dSize, "Lettura completata")
	return img, "", nil
}

func (f rt4dFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < rt4dSize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	l, err := dm5kOpen(ctx, port)
	if err != nil {
		return "", err
	}
	defer l.p.Close()
	prog(0, rt4dSize, "Collegamento con la radio…")
	if err := l.connect(); err != nil {
		return "", genFail("Collegamento", err)
	}
	// come il programma: blocco 0, poi blocco 8 (password); in più il blocco del
	// VFO, da riscrivere com'è
	if _, err := l.readBlock(0); err != nil {
		return "", genFail("Identificazione", err)
	}
	b8, err := l.readBlock(rt4dBlkCfg)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	if err := dm5kCheckPwd(b8); err != nil {
		_ = l.p.Write(dm5kEnd)
		return "", genFail("Identificazione", err)
	}
	vfo, err := l.readBlock(rt4dBlkVfo)
	if err != nil {
		return "", genFail("Identificazione", err)
	}
	cfg := append([]byte{}, b8...)
	cfg[14], cfg[15] = 0xFF, 0xFF // come il programma
	if err := l.writeBlock(0x90, 0, cfg); err != nil {
		return "", genFail("Scrittura", err)
	}
	for k := 0; k < rt4dChBlocks; k++ {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(k*dm5kBlock, rt4dSize, "Scrittura: canali…")
		if err := l.writeBlock(0x91, k, img[k*dm5kBlock:(k+1)*dm5kBlock]); err != nil {
			return "", genFail("Scrittura", err)
		}
	}
	if err := l.writeBlock(0x92, 0, vfo); err != nil {
		return "", genFail("Scrittura", err)
	}
	for k := 0; k < rt4dZnBlocks; k++ {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(rt4dZoneOff+k*dm5kBlock, rt4dSize, "Scrittura: zone…")
		if err := l.writeBlock(0x93, k, img[rt4dZoneOff+k*dm5kBlock:rt4dZoneOff+(k+1)*dm5kBlock]); err != nil {
			return "", genFail("Scrittura", err)
		}
	}
	_ = l.p.Write(dm5kEnd)
	prog(rt4dSize, rt4dSize, "Scrittura completata")
	return "", nil
}
