// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da dmrconfig (https://github.com/OpenRTX/dmrconfig, licenza BSD a
// 3 clausole, Copyright (C) 2018 Serge Vakulenko KK6ABQ): anytone_ht.c,
// anytone_ht-map.h, serial.c; confrontate con qdmr (https://github.com/hmatuschek/qdmr,
// GNU GPL v3, Hannes Matuschek DM3MAT): anytone_interface.cc, anytone_codeplug.hh,
// d868uv_codeplug.cc, dmr6x2uv_codeplug.hh.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Famiglia AnyTone (4.45): AnyTone AT-D868UV, AT-D878UV, AT-D878UV II e le BTECH
// DMR-6X2 e DMR-6X2 PRO, che ne derivano. La radio si collega con il suo cavo USB
// e compare come porta seriale (USB 28E9:018A).
//
// Protocollo (serial.c di dmrconfig, anytone_interface.cc di qdmr):
//   «PROGRAM» → «QX» 06; 02 → «I» + modello (7 caratteri) + 00 + versione + 06;
//   lettura: «R» indirizzo(4, dal byte alto) 10 → «W» indirizzo(4) 10 dati(16)
//   somma 06 (somma = somma dei byte da indirizzo a dati);
//   scrittura: «W» indirizzo(4) 10 dati(16) somma 06 → 06; «END» → 06.
// dmrconfig legge 64 byte per volta; qui 16, come qdmr e come in scrittura.
//
// La memoria della radio è sparsa su più di 64 MB di indirizzi: l'immagine è la
// stessa di dmrconfig (file .img di 1 606 528 byte, le aree di anytone_ht-map.h una
// dopo l'altra), così i file si scambiano con dmrconfig. Come dmrconfig non si
// leggono i canali, le zone, le liste di scansione e i contatti non validi (restano
// a FF). Differenza voluta: dmrconfig decide a blocchi di 64 byte guardando solo il
// contatto in cui il blocco comincia (i contatti sono lunghi 100 byte) e può
// lasciare a FF l'inizio di un contatto valido che segue uno vuoto; qui un blocco
// si salta solo se tutti i contatti che tocca sono vuoti.
//
// Canale (64 byte): 0 RX e 4 scostamento TX (BCD, 8 cifre, 10 Hz); 8 bit 0-1 modo
// (0 analogico, 1 DMR, 2 A+D, 3 D+A), bit 2-3 potenza (0 bassa, 1 media, 2 alta,
// 3 turbo), bit 4 banda (1 = 25 kHz), bit 6-7 verso dello scostamento (1 +, 2 −);
// 9 bit 0 CTCSS RX, 1 DCS RX, 2 CTCSS TX, 3 DCS TX, 5 solo ricezione; 10 indice
// CTCSS TX, 11 indice CTCSS RX (51 = valore libero in 16-17); 12 DCS TX e 14 DCS
// RX (codice in ottale, +512 invertito); 20 contatto; 25 bit 4 squelch a toni;
// 27 lista di scansione; 28 lista dei gruppi; 32 color code; 33 bit 0 slot 2;
// 35 nome (16 caratteri). DMR-6X2: otto liste di scansione da 54.
//
// SCRITTURA: qui si scrivono SOLO i canali validi, la tabella dei canali validi, le
// zone (elenchi, nomi, tabella, canale A e B di ogni zona) e le liste di scansione;
// dalla 4.47 anche le impostazioni generali, le righe di accensione e il primo ID
// (settings_447.go), ma solo le righe diverse da quelle della radio.
// Contatti, gruppi e messaggi non si toccano
// (dmrconfig riscrive tutto e ricostruisce l'indice dei contatti). Non è copiato
// l'errore di dmrconfig che, creando le zone, scrive il canale A e B della zona n
// nella posizione 2·n della tabella.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	atSize      = 1606528
	atNCh       = 4000
	atNZones    = 250
	atNScan     = 250
	atNContacts = 10000
	// posizioni nell'immagine
	atOffCh         = 0x000040
	atOffZoneLists  = 0x03E8C0 // 250 elenchi da 250 canali (16 bit, FFFF = libero)
	atOffScan       = 0x05DCC0 // 250 liste da 192 byte
	atOffZoneMap    = 0x070940
	atOffScanMap    = 0x070980
	atOffChMap      = 0x070A40
	atOffSettings   = 0x071600
	atOffZChA       = 0x071700 // canale A di ogni zona (16 bit)
	atOffZChB       = 0x071900
	atOffZoneNames  = 0x071DC0 // 250 nomi da 32 byte
	atOffRadioID    = 0x073D00
	atOffContactMap = 0x080140 // bit a 1 = contatto NON valido
	atOffContacts   = 0x080640 // 10000 contatti da 100 byte
	// indirizzi nella radio
	atAddrCh        = 0x00800000 // banchi da 128 canali ogni 0x40000
	atAddrZoneLists = 0x01000000
	atAddrScan      = 0x01080000 // 16 liste ogni 0x40000, una ogni 0x200
	atAddrZoneMap   = 0x024C1300
	atAddrScanMap   = 0x024C1340
	atAddrChMap     = 0x024C1500
	atAddrZChA      = 0x02500100
	atAddrZChB      = 0x02500300
	atAddrZoneNames = 0x02540000
)

// toni CTCSS della radio (decimi di Hz), nell'ordine degli indici
var atCTCSS = []int{625, 670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230,
	1273, 1318, 1365, 1413, 1462, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835, 1862, 1899, 1928, 1966,
	1995, 2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

type atRegion struct {
	addr uint32
	size int
	off  int // posizione nell'immagine
}

// aree della memoria nell'ordine dell'immagine (anytone_ht-map.h di dmrconfig)
var atRegions = func() []atRegion {
	var out []atRegion
	off := 0
	add := func(addr uint32, size, n int, stride uint32) {
		for k := 0; k < n; k++ {
			out = append(out, atRegion{addr + uint32(k)*stride, size, off})
			off += size
		}
	}
	add(0x02FA0010, 0x40, 1, 0)         // nome del modello e data
	add(atAddrCh, 0x2000, 31, 0x40000)  // canali 1-3968
	add(0x00FC0000, 0x880, 1, 0)        // canali 3969-4000, VFO A e B
	add(atAddrZoneLists, 0x1F400, 1, 0) // elenchi dei canali delle zone
	for b := uint32(0); b < 15; b++ {   // liste di scansione 1-240
		add(atAddrScan+b*0x40000, 0xC0, 16, 0x200)
	}
	add(atAddrScan+15*0x40000, 0xC0, 10, 0x200) // liste di scansione 241-250
	add(0x01640000, 0x640, 1, 0)
	add(0x01640800, 0xC0, 1, 0)
	add(0x02140000, 0x800, 12, 0x40000) // messaggi 1-96
	add(0x02440000, 0x400, 1, 0)        // messaggi 97-100
	add(0x02480000, 0x1C0, 1, 0)
	add(0x02480200, 0x40, 1, 0)
	add(0x024C0000, 0x40, 1, 0)
	add(0x024C0C80, 0x40, 1, 0)
	add(0x024C0D00, 0x200, 1, 0)
	add(0x024C1000, 0x140, 1, 0)
	add(0x024C1280, 0x40, 1, 0)
	add(atAddrZoneMap, 0x80, 1, 0) // zone valide e, da +0x40, liste di scansione valide
	add(0x024C1400, 0x80, 1, 0)
	add(atAddrChMap, 0x240, 1, 0) // canali validi
	add(0x024C1800, 0x500, 1, 0)
	add(0x024C2000, 0x440, 1, 0)
	add(0x024C2600, 0x40, 1, 0)
	add(0x02500000, 0x640, 1, 0) // impostazioni generali (con il canale A e B delle zone)
	add(0x02501000, 0x40, 1, 0)
	add(0x02501100, 0x140, 1, 0)       // messaggio GPS
	add(atAddrZoneNames, 0x1F40, 1, 0) // nomi delle zone
	add(0x02580000, 0x1F40, 1, 0)      // ID della radio
	add(0x025C0000, 0x880, 1, 0)
	add(0x025C0B00, 0x40, 1, 0)
	add(0x02600000, 0x9C40, 1, 0)  // elenco dei contatti validi
	add(0x02640000, 0x500, 1, 0)   // tabella dei contatti
	add(0x02680000, 0xF4240, 1, 0) // contatti
	add(0x02900000, 0x80, 2, 0x100)
	add(0x02940000, 0x180, 1, 0)       // contatti analogici
	add(0x02980000, 0x140, 250, 0x200) // liste dei gruppi
	if off != atSize {
		panic(off) // la mappa non corrisponde alla dimensione dell'immagine
	}
	return out
}()

type atOpt struct {
	ident string // modello dichiarato dalla radio (e scritto in testa alla memoria)
	x6    bool   // DMR-6X2: otto liste di scansione per canale
}

func ato(m *genModel) *atOpt { return m.Opt.(*atOpt) }

type atFam struct{}

const atNoteCommon = "Si collega con il cavo USB della radio, che compare come porta seriale (COM); se Windows non la crea serve il driver " +
	"«USB Virtual Com Port» del costruttore. Si modificano frequenze, nomi (16 caratteri), modo (analogico / DMR), potenza, banda, toni e " +
	"solo ricezione; nei canali DMR color code, slot, contatto e lista dei gruppi restano quelli impostati. Un canale nuovo entra nella " +
	"prima zona con un posto libero. Si scrivono canali, zone e liste di scansione; ID, nome, testo di accensione e impostazioni generali " +
	"solo se li hai cambiati (finestra «Impostazioni»). Contatti, gruppi e messaggi non si toccano. Non provata su una radio vera: prima la sola lettura."

func init() {
	mk := func(id, vendor, model, name, ident string, x6 bool, extra string) *genModel {
		return &genModel{ID: id, Vendor: vendor, Model: model, RClass: "IU0VYT_ANYTONE_" + ident, Name: name,
			Slots: atNCh, First: 1, NameMax: 16, Charset: charsetASCII,
			Powers: []genPower{{"Alta", 5, 0}, {"Media", 2.5, 1}, {"Bassa", 1, 2}},
			Bands:  []band{{136000000, 174000000}, {400000000, 480000000}}, Narrow: true, HasScan: false, NoChirp: true,
			DCS: dtcsAll(), Baud: 115200, Fam: atFam{}, Opt: &atOpt{ident: ident, x6: x6}, Digital: "DMR",
			Notes: extra + atNoteCommon}
	}
	genRegister(
		mk("at-d868uv", "AnyTone", "AT-D868UV", "AnyTone AT-D868UV", "D868UVE", false, ""),
		mk("at-d878uv", "AnyTone", "AT-D878UV", "AnyTone AT-D878UV / AT-D878UV Plus", "D878UV", false, ""),
		mk("at-d878uv2", "AnyTone", "AT-D878UV II", "AnyTone AT-D878UV II Plus", "D878UV2", false, ""),
		mk("bt-dmr6x2", "BTECH", "DMR-6X2", "BTECH DMR-6X2", "D6X2UV", true, ""),
		mk("bt-dmr6x2pro", "BTECH", "DMR-6X2 PRO", "BTECH DMR-6X2 PRO", "D6X2UV2", true,
			"Sperimentale: modello non presente in dmrconfig, indirizzi del programma libero qdmr. "),
	)
}

// atFam non passa da «Identifica» con le altre: è provata per ultima (la radio ha
// la sua porta USB e non usa il cavo di programmazione delle altre)
func (atFam) identLast() {}

// ---------------------------------------------------------------- canali

func atRec(img []byte, i int) []byte { return img[atOffCh+i*64:][:64] }

func atBit(img []byte, off, i int) bool { return img[off+i/8]>>(uint(i)&7)&1 == 1 }

func atSetBit(img []byte, off, i int, on bool) {
	if on {
		img[off+i/8] |= 1 << (uint(i) & 7)
	} else {
		img[off+i/8] &^= 1 << (uint(i) & 7)
	}
}

func atValid(img []byte, i int) bool { return atBit(img, atOffChMap, i) }

// frequenza in BCD a 8 cifre, unità 10 Hz, dalla cifra più alta
func atFreq(b []byte) (int64, bool) {
	var v int64
	for _, c := range b[:4] {
		if c>>4 > 9 || c&15 > 9 {
			return 0, false
		}
		v = v*100 + int64(c>>4)*10 + int64(c&15)
	}
	return v * 10, true
}

func atPutFreq(b []byte, hz int64) {
	v := hz / 10
	for k := 3; k >= 0; k-- {
		d := v % 100
		b[k] = byte(d/10)<<4 | byte(d%10)
		v /= 100
	}
}

// modello scritto in testa alla memoria (fino al primo carattere non stampabile)
func atHeadName(img []byte) string {
	n := 0
	for n < 8 && n < len(img) && img[n] > 0x20 && img[n] < 0x7F {
		n++
	}
	return string(img[:n])
}

func (f atFam) Match(m *genModel, img []byte) bool {
	return len(img) == atSize && atHeadName(img) == ato(m).ident
}

func atToneGet(ctcss, dcs bool, idx byte, code uint16, custom uint16) string {
	switch {
	case ctcss:
		if int(idx) < len(atCTCSS) {
			return fmtTone(atCTCSS[idx])
		}
		if custom == 0 || custom == 0xFFFF {
			return ""
		}
		return fmtTone(int(custom))
	case dcs:
		t := genTone{Mode: "DTCS", Code: int(code>>6&7)*100 + int(code>>3&7)*10 + int(code&7), Inv: code>>9&1 == 1}
		return t.String()
	}
	return ""
}

func (f atFam) Get(m *genModel, img []byte, i int) *Chan {
	if len(img) != atSize || i < 0 || i >= atNCh || !atValid(img, i) {
		return nil
	}
	r := atRec(img, i)
	rx, ok1 := atFreq(r[0:])
	off, ok2 := atFreq(r[4:])
	if !ok1 || !ok2 || rx == 0 {
		return nil
	}
	c := &Chan{Rx: rx, Tx: rx, Name: genNameFromBytes(r[35:51], true)}
	switch r[8] >> 6 {
	case 1:
		c.Tx = rx + off
	case 2:
		c.Tx = rx - off
	}
	mode := r[8] & 3
	c.Dig = mode == 1 || mode == 3
	c.Narrow = r[8]>>4&1 == 0
	switch r[8] >> 2 & 3 {
	case 0:
		c.Power = 2
	case 1:
		c.Power = 1
	default: // alta e turbo
		c.Power = 0
	}
	if r[9]>>5&1 == 1 {
		c.TxOff = true
		c.Tx = rx
	}
	if !c.Dig {
		custom := binary.LittleEndian.Uint16(r[16:])
		c.RxTone = atToneGet(r[9]&1 != 0, r[9]&2 != 0, r[11], binary.LittleEndian.Uint16(r[14:]), custom)
		c.TxTone = atToneGet(r[9]&4 != 0, r[9]&8 != 0, r[10], binary.LittleEndian.Uint16(r[12:]), custom)
	}
	return c
}

// indice del tono nella tabella della radio (51 = valore libero)
func atCTCSSIdx(hz10 int) int {
	for k, v := range atCTCSS {
		if v == hz10 {
			return k
		}
	}
	return len(atCTCSS)
}

func atCustom(t genTone) bool { return t.Mode == "Tone" && atCTCSSIdx(t.Hz10) == len(atCTCSS) }

// la radio ha un solo valore CTCSS libero per canale: due toni fuori tabella
// devono essere uguali
func (atFam) toneOK(t, r genTone) bool {
	if atCustom(t) && atCustom(r) && t.Hz10 != r.Hz10 {
		return false
	}
	for _, x := range []genTone{t, r} {
		if x.Mode == "Tone" && (x.Hz10 < 600 || x.Hz10 > 2600) {
			return false
		}
		if x.Mode == "DTCS" && (x.Code > 777 || x.Code%10 > 7 || x.Code/10%10 > 7) {
			return false
		}
	}
	return true
}

func (atFam) fit(m *genModel, c *Chan) (bool, string) {
	if c.Rx%10 != 0 || c.Tx%10 != 0 {
		return false, "la radio memorizza le frequenze a passi di 10 Hz"
	}
	return genBandFit(m, c)
}

func atDCS(t genTone) uint16 {
	v := uint16(t.Code/100)<<6 | uint16(t.Code/10%10)<<3 | uint16(t.Code%10)
	if t.Inv {
		v |= 1 << 9
	}
	return v
}

func atPutName(r []byte, name string) {
	for k := 35; k < 52; k++ {
		r[k] = 0
	}
	copy(r[35:51], asciiName(name))
}

func (f atFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	if len(img) != atSize {
		return fmt.Errorf("i dati non sono una memoria della %s", m.Name)
	}
	if i < 0 || i >= atNCh {
		return fmt.Errorf("canale %d fuori dalla memoria", i+1)
	}
	r := atRec(img, i)
	if c == nil {
		if atValid(img, i) {
			atSetBit(img, atOffChMap, i, false)
			for k := range r {
				r[k] = 0xFF
			}
			atUnlink(img, i)
		}
		return nil
	}
	if c.Rx%10 != 0 || (!c.TxOff && c.Tx%10 != 0) {
		return fmt.Errorf("la radio memorizza le frequenze a passi di 10 Hz")
	}
	if c.Rx <= 0 || c.Rx >= 1000000000 {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	if !c.TxOff && (c.Tx <= 0 || c.Tx >= 1000000000) {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Tx))
	}
	var tt, rt genTone
	if !c.Dig {
		var err error
		if tt, err = genToneParse(c.TxTone); err != nil {
			return err
		}
		if rt, err = genToneParse(c.RxTone); err != nil {
			return err
		}
		if !(atFam{}).toneOK(tt, rt) {
			return fmt.Errorf("tono non disponibile su questa radio")
		}
	}
	isNew := !atValid(img, i)
	mode := r[8] & 3
	wasDig := mode == 1 || mode == 3
	fresh := isNew || wasDig != c.Dig
	if fresh {
		// canale nuovo, o passato da analogico a DMR o viceversa: come setup_channel
		// di dmrconfig, si parte da un canale azzerato
		for k := range r {
			r[k] = 0
		}
		binary.LittleEndian.PutUint16(r[16:], 2511) // CTCSS libero: 251,1 Hz
		binary.LittleEndian.PutUint16(r[20:], 0xFFFF)
		r[28] = 0xFF // nessuna lista dei gruppi
		if c.Dig {
			r[8] = 1
			r[32] = 1 // color code 1, slot 1
			if !atBit(img, atOffContactMap, 0) {
				binary.LittleEndian.PutUint16(r[20:], 0) // primo contatto
			}
		}
		if ato(m).x6 {
			for k := 54; k < 62; k++ { // nessuna delle otto liste di scansione
				r[k] = 0xFF
			}
		} else {
			r[27] = 0xFF // nessuna lista di scansione
		}
		r[8] |= 2 << 2 // potenza alta, poi corretta sotto
	}
	atPutName(r, c.Name)
	atPutFreq(r[0:], c.Rx)
	r[8] &^= 3 << 6
	switch {
	case c.TxOff || c.Tx == c.Rx:
		copy(r[4:8], []byte{0x00, 0x01, 0x00, 0x00}) // scostamento predefinito di dmrconfig
	case c.Tx > c.Rx:
		r[8] |= 1 << 6
		atPutFreq(r[4:], c.Tx-c.Rx)
	default:
		r[8] |= 2 << 6
		atPutFreq(r[4:], c.Rx-c.Tx)
	}
	// potenza: alta e turbo si leggono «alta»; il valore cambia solo se cambia il livello
	cur := 0
	switch r[8] >> 2 & 3 {
	case 0:
		cur = 2
	case 1:
		cur = 1
	}
	if want := m.powerSem(m.powerIndex(c.Power)); want != cur {
		r[8] = r[8]&^(3<<2) | []byte{2, 1, 0}[want]<<2
	}
	if c.TxOff {
		r[9] |= 1 << 5
	} else {
		r[9] &^= 1 << 5
	}
	if c.Dig {
		if fresh {
			r[8] &^= 1 << 4 // 12,5 kHz, come dmrconfig per i canali DMR
		}
		return atFinish(img, i, isNew)
	}
	if c.Narrow {
		r[8] &^= 1 << 4
	} else {
		r[8] |= 1 << 4
	}
	r[9] &^= 0x0F
	r[10], r[11] = 0, 0
	binary.LittleEndian.PutUint16(r[12:], 0)
	binary.LittleEndian.PutUint16(r[14:], 0)
	switch rt.Mode {
	case "Tone":
		r[9] |= 1
		r[11] = byte(atCTCSSIdx(rt.Hz10))
		if atCustom(rt) {
			binary.LittleEndian.PutUint16(r[16:], uint16(rt.Hz10))
		}
	case "DTCS":
		r[9] |= 2
		binary.LittleEndian.PutUint16(r[14:], atDCS(rt))
	}
	if rt.Mode != "" {
		r[25] |= 1 << 4 // squelch a toni
	} else {
		r[25] &^= 1 << 4
	}
	switch tt.Mode {
	case "Tone":
		r[9] |= 4
		r[10] = byte(atCTCSSIdx(tt.Hz10))
		if atCustom(tt) {
			binary.LittleEndian.PutUint16(r[16:], uint16(tt.Hz10))
		}
	case "DTCS":
		r[9] |= 8
		binary.LittleEndian.PutUint16(r[12:], atDCS(tt))
	}
	return atFinish(img, i, isNew)
}

func atFinish(img []byte, i int, isNew bool) error {
	if isNew {
		atSetBit(img, atOffChMap, i, true)
		atZoneAdd(img, i)
	}
	return nil
}

func (f atFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, atRec(img, i)...)
}

// PutRec: i byte di un canale spostato in un'altra posizione (così un canale DMR
// conserva color code, slot, contatto e lista dei gruppi); se la posizione era
// vuota il canale diventa valido ed entra in una zona
func (f atFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(img) != atSize || i < 0 || i >= atNCh || len(raw) != 64 {
		return
	}
	if _, ok := atFreq(raw); !ok { // non sono i byte di un canale
		return
	}
	copy(atRec(img, i), raw)
	if !atValid(img, i) {
		atSetBit(img, atOffChMap, i, true)
		atZoneAdd(img, i)
	}
}

// ---------------------------------------------------------------- zone e liste di scansione

func atZoneValid(img []byte, z int) bool  { return atBit(img, atOffZoneMap, z) }
func atScanValid(img []byte, s int) bool  { return atBit(img, atOffScanMap, s) }
func atZoneList(img []byte, z int) []byte { return img[atOffZoneLists+z*512:][:500] }
func atScanRec(img []byte, s int) []byte  { return img[atOffScan+s*192:][:192] }

// elenco di canali a 16 bit (FFFF = libero)
func atList(b []byte) []uint16 {
	var out []uint16
	for k := 0; k+1 < len(b); k += 2 {
		if v := binary.LittleEndian.Uint16(b[k:]); v != 0xFFFF {
			out = append(out, v)
		}
	}
	return out
}

func atPutList(b []byte, v []uint16) {
	for k := 0; k+1 < len(b); k += 2 {
		x := uint16(0xFFFF)
		if k/2 < len(v) {
			x = v[k/2]
		}
		binary.LittleEndian.PutUint16(b[k:], x)
	}
}

// toglie il canale dall'elenco; true se c'era
func atDrop(b []byte, ch uint16) bool {
	old := atList(b)
	var keep []uint16
	for _, v := range old {
		if v != ch {
			keep = append(keep, v)
		}
	}
	if len(keep) == len(old) {
		return false
	}
	atPutList(b, keep)
	return true
}

// atZoneAB: se il canale A o B della zona non è più nell'elenco, diventa il primo
// (B il secondo, se c'è) come quando dmrconfig crea la zona
func atZoneAB(img []byte, z int) {
	l := atList(atZoneList(img, z))
	if len(l) == 0 {
		return
	}
	in := func(v uint16) bool {
		for _, x := range l {
			if x == v {
				return true
			}
		}
		return false
	}
	a, b := img[atOffZChA+2*z:], img[atOffZChB+2*z:]
	if !in(binary.LittleEndian.Uint16(a)) {
		binary.LittleEndian.PutUint16(a, l[0])
	}
	if !in(binary.LittleEndian.Uint16(b)) {
		second := l[0]
		if len(l) > 1 {
			second = l[1]
		}
		binary.LittleEndian.PutUint16(b, second)
	}
}

// atUnlink toglie il canale i da zone e liste di scansione
func atUnlink(img []byte, i int) {
	ch := uint16(i)
	for z := 0; z < atNZones; z++ {
		if atZoneValid(img, z) && atDrop(atZoneList(img, z), ch) {
			atZoneAB(img, z)
		}
	}
	for s := 0; s < atNScan; s++ {
		if !atScanValid(img, s) {
			continue
		}
		sl := atScanRec(img, s)
		atDrop(sl[32:132], ch)
		// canali prioritari: numero del canale (da 1), 0 = canale corrente, FFFF = nessuno
		for k, p := range []int{2, 4} {
			if binary.LittleEndian.Uint16(sl[p:]) == ch+1 {
				binary.LittleEndian.PutUint16(sl[p:], 0xFFFF)
				sl[1] &^= 1 << uint(k)
			}
		}
	}
}

// atZoneAdd: canale nuovo nella prima zona con posto libero, altrimenti in una
// zona nuova («Zone n»)
func atZoneAdd(img []byte, i int) {
	ch := uint16(i)
	for z := 0; z < atNZones; z++ {
		if !atZoneValid(img, z) {
			continue
		}
		for _, v := range atList(atZoneList(img, z)) {
			if v == ch {
				return
			}
		}
	}
	for z := 0; z < atNZones; z++ {
		if !atZoneValid(img, z) {
			continue
		}
		if l := atList(atZoneList(img, z)); len(l) < 250 {
			atPutList(atZoneList(img, z), append(l, ch))
			atZoneAB(img, z)
			return
		}
	}
	for z := 0; z < atNZones; z++ {
		if atZoneValid(img, z) {
			continue
		}
		atSetBit(img, atOffZoneMap, z, true)
		name := img[atOffZoneNames+z*32:][:32]
		for k := range name {
			name[k] = 0
		}
		copy(name[:16], fmt.Sprintf("Zone %d", z+1))
		full := img[atOffZoneLists+z*512:][:512]
		for k := range full {
			full[k] = 0xFF
		}
		atPutList(atZoneList(img, z), []uint16{ch})
		binary.LittleEndian.PutUint16(img[atOffZChA+2*z:], ch)
		binary.LittleEndian.PutUint16(img[atOffZChB+2*z:], ch)
		return
	}
}

// ---------------------------------------------------------------- protocollo

// blocchi dell'immagine (64 byte) da non leggere: canali, zone, liste di scansione
// e contatti non validi
func atSkip(img []byte, off int) bool {
	switch {
	case off >= atOffCh && off < atOffCh+atNCh*64:
		return !atValid(img, (off-atOffCh)/64)
	case off >= atOffZoneLists && off < atOffZoneLists+atNZones*512:
		return !atZoneValid(img, (off-atOffZoneLists)/512)
	case off >= atOffScan && off < atOffScan+atNScan*192:
		return !atScanValid(img, (off-atOffScan)/192)
	case off >= atOffContacts && off < atOffContacts+atNContacts*100:
		lo, hi := (off-atOffContacts)/100, (off+63-atOffContacts)/100
		for c := lo; c <= hi && c < atNContacts; c++ {
			if !atBit(img, atOffContactMap, c) {
				return false
			}
		}
		return true
	}
	return false
}

type atSess struct {
	*rawSess
}

const atTimeout = 1500 * time.Millisecond

// atEnter: ingresso in programmazione; restituisce modello e versione dichiarati
func (s *atSess) enter() (model, version string, err error) {
	var last error
	for try := 0; try < 2; try++ {
		if try > 0 {
			if err := s.sleep(500 * time.Millisecond); err != nil {
				return "", "", err
			}
		}
		_ = s.p.Purge()
		if err := s.send([]byte("PROGRAM")); err != nil {
			return "", "", err
		}
		ack, err := s.read(3, atTimeout)
		if err != nil {
			last = err
			if s.ctx.Err() != nil {
				return "", "", errCancelled
			}
			continue
		}
		if !bytes.Equal(ack, []byte("QX\x06")) {
			last = rErr("risposta inattesa della radio (% X)", ack)
			continue
		}
		if err := s.send([]byte{0x02}); err != nil {
			return "", "", err
		}
		id, err := s.read(16, atTimeout)
		if err != nil {
			last = err
			continue
		}
		if id[0] != 'I' || id[15] != 0x06 {
			last = rErr("risposta inattesa della radio (% X)", id)
			continue
		}
		cut := func(b []byte) string {
			if k := bytes.IndexByte(b, 0); k >= 0 {
				b = b[:k]
			}
			return strings.TrimSpace(string(b))
		}
		return cut(id[1:8]), cut(id[9:15]), nil
	}
	return "", "", last
}

func (s *atSess) end() {
	if s.send([]byte("END")) == nil {
		_, _ = s.read(1, atTimeout)
	}
}

func atSum(b []byte) (sum byte) {
	for _, c := range b {
		sum += c
	}
	return
}

func (s *atSess) read16(addr uint32) ([]byte, error) {
	cmd := []byte{'R', byte(addr >> 24), byte(addr >> 16), byte(addr >> 8), byte(addr), 16}
	var last error
	for try := 0; try < 3; try++ {
		if err := s.send(cmd); err != nil {
			return nil, err
		}
		r, err := s.read(24, atTimeout)
		if err != nil {
			return nil, err
		}
		if r[0] != 'W' || r[23] != 0x06 || !bytes.Equal(r[1:6], cmd[1:6]) {
			return nil, rErr("risposta inattesa alla lettura del blocco %08X", addr)
		}
		if atSum(r[1:22]) != r[22] {
			last = rErr("dati del blocco %08X non validi (somma di controllo)", addr)
			continue
		}
		return r[6:22], nil
	}
	return nil, last
}

func (s *atSess) write16(addr uint32, data []byte) error {
	cmd := append([]byte{'W', byte(addr >> 24), byte(addr >> 16), byte(addr >> 8), byte(addr), 16}, data[:16]...)
	cmd = append(cmd, atSum(cmd[1:]), 0x06)
	if err := s.send(cmd); err != nil {
		return err
	}
	ack, err := s.read(1, atTimeout)
	if err != nil {
		return err
	}
	if ack[0] != 0x06 {
		return rErr("la radio ha rifiutato il blocco %08X (risposta %02X)", addr, ack[0])
	}
	return nil
}

func atOpen(ctx context.Context, m *genModel, port string) (*atSess, string, error) {
	rs, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	s := &atSess{rs}
	model, ver, err := s.enter()
	if err != nil {
		s.p.Close()
		return nil, "", genFail("Identificazione", err)
	}
	info := model + " " + ver
	if model != ato(m).ident {
		s.end()
		s.p.Close()
		return nil, info, genFail("Identificazione", rErr("modello errato: la radio si identifica come %q (attesa %q)", model, ato(m).ident))
	}
	return s, info, nil
}

// legge un'area dell'immagine a blocchi di 16 byte
func (s *atSess) readSpan(img []byte, addr uint32, off, n int) error {
	for k := 0; k < n; k += 16 {
		d, err := s.read16(addr + uint32(k))
		if err != nil {
			return err
		}
		copy(img[off+k:], d)
	}
	return nil
}

func (f atFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	prog(0, 1, "Collegamento con la radio…")
	s, info, err := atOpen(ctx, m, port)
	if err != nil {
		return nil, info, err
	}
	defer s.p.Close()
	fail := func(err error) ([]byte, string, error) {
		s.end()
		if ctx.Err() != nil {
			return nil, info, errCancelled
		}
		return nil, info, genFail("Lettura", err)
	}
	img := bytes.Repeat([]byte{0xFF}, atSize)
	// prima le tabelle di zone, liste di scansione, canali e contatti validi
	first := map[uint32]bool{atAddrZoneMap: true, atAddrChMap: true, 0x02640000: true}
	for _, rg := range atRegions {
		if first[rg.addr] {
			if err := s.readSpan(img, rg.addr, rg.off, rg.size); err != nil {
				return fail(err)
			}
		}
	}
	total := 0
	for off := 0; off < atSize; off += 64 {
		if !atSkip(img, off) {
			total++
		}
	}
	done := 0
	for _, rg := range atRegions {
		for k := 0; k < rg.size; k += 64 {
			if atSkip(img, rg.off+k) {
				continue
			}
			if ctx.Err() != nil {
				return fail(ctx.Err())
			}
			prog(done, total, "Lettura…")
			done++
			if first[rg.addr] {
				continue
			}
			if err := s.readSpan(img, rg.addr+uint32(k), rg.off+k, 64); err != nil {
				return fail(err)
			}
		}
	}
	s.end()
	if h := atHeadName(img); h != ato(m).ident {
		// la memoria dichiara un altro modello: nessuna modifica è stata fatta alla radio
		return nil, info, genFail("Lettura", rErr("modello errato: la radio si identifica come %q (attesa %q)", h, ato(m).ident))
	}
	prog(total, total, "Lettura completata")
	return img, info, nil
}

type atSpan struct {
	addr uint32
	off  int
	n    int
}

// aree scritte nella radio: canali validi, tabelle, zone e liste di scansione
func atWriteSpans(img []byte) []atSpan {
	var out []atSpan
	for i := 0; i < atNCh; i++ {
		if atValid(img, i) {
			out = append(out, atSpan{atAddrCh + uint32(i/128)*0x40000 + uint32(i%128)*64, atOffCh + i*64, 64})
		}
	}
	for z := 0; z < atNZones; z++ {
		if atZoneValid(img, z) {
			out = append(out, atSpan{atAddrZoneLists + uint32(z)*512, atOffZoneLists + z*512, 512},
				atSpan{atAddrZoneNames + uint32(z)*32, atOffZoneNames + z*32, 32})
		}
	}
	for sl := 0; sl < atNScan; sl++ {
		if atScanValid(img, sl) {
			out = append(out, atSpan{atAddrScan + uint32(sl/16)*0x40000 + uint32(sl%16)*0x200, atOffScan + sl*192, 192})
		}
	}
	out = append(out,
		atSpan{atAddrZChA, atOffZChA, 0x200},
		atSpan{atAddrZChB, atOffZChB, 0x200},
		atSpan{atAddrZoneMap, atOffZoneMap, 0x20},
		atSpan{atAddrScanMap, atOffScanMap, 0x20},
		atSpan{atAddrChMap, atOffChMap, 0x200})
	return out
}

func (f atFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("i dati non sono una memoria della %s", m.Name)
	}
	spans := atWriteSpans(img)
	total := 0
	for _, sp := range spans {
		total += sp.n / 16
	}
	prog(0, total, "Collegamento con la radio…")
	s, info, err := atOpen(ctx, m, port)
	if err != nil {
		return info, err
	}
	defer s.p.Close()
	done := 0
	for _, sp := range spans {
		for k := 0; k < sp.n; k += 16 {
			if ctx.Err() != nil {
				s.end()
				return info, errCancelled
			}
			prog(done, total, "Scrittura…")
			if err := s.write16(sp.addr+uint32(k), img[sp.off+k:sp.off+k+16]); err != nil {
				s.end()
				return info, genFail("Scrittura", err)
			}
			done++
		}
	}
	// 4.47: impostazioni generali, righe di accensione e primo ID della radio.
	// Si rileggono dalla radio e si scrivono solo le righe diverse: se le
	// impostazioni non sono state cambiate qui non si scrive nulla.
	for _, sp := range atSetSpans {
		for k := 0; k < sp.n; k += 16 {
			if ctx.Err() != nil {
				s.end()
				return info, errCancelled
			}
			want := img[sp.off+k : sp.off+k+16]
			have, err := s.read16(sp.addr + uint32(k))
			if err != nil {
				s.end()
				return info, genFail("Scrittura", err)
			}
			if bytes.Equal(have, want) {
				continue
			}
			if err := s.write16(sp.addr+uint32(k), want); err != nil {
				s.end()
				return info, genFail("Scrittura", err)
			}
		}
	}
	s.end()
	prog(total, total, "Scrittura completata")
	return info, nil
}
