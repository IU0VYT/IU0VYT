// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// vx1.py, yaesu_clone.py (4.41).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.41: Yaesu VX-1R (1997). La radio ha due «gruppi di configurazione» che
// tengono memorie diverse, più dieci memorie della banda delle onde medie. In
// CHIRP sono tre sotto-radio; qui una lista sola di 260 posizioni:
//
//	  1- 52  gruppo 1: memorie complete (scostamento, subtono o DCS)
//	 53- 72  gruppo 1: limiti di scansione L1/U1 … L10/U10
//	 73- 80  gruppo 1: canali «home» delle otto bande
//	 81-222  gruppo 2: memorie semplici (senza scostamento proprio né tono)
//	223-242  gruppo 2: limiti di scansione L1/U1 … L10/U10
//	243-250  gruppo 2: canali «home» delle otto bande
//	251-260  onde medie (da 0,5 a 1,7 MHz, solo AM)
//
// La radio usa un gruppo alla volta (menu «GRP»): le memorie dell'altro restano
// nella memoria ma non si vedono.
//
// Gruppo 2: il canale non ha uno scostamento suo né un tono suo. Ha solo il
// verso dello scostamento (quello normale della banda: 600 kHz sopra i 144 MHz,
// 5 MHz sopra i 420 MHz, come lo mostra CHIRP) e il tipo di tono, senza la
// frequenza. Qui un canale del gruppo 2 accetta solo quegli scostamenti e non
// accetta toni; il tipo di tono che c'è già resta com'è.
//
// Copia: tutta la memoria (4818 byte) in un blocco solo, a 9600 baud. In
// lettura la radio invia e il programma non risponde niente; in scrittura la
// radio conferma alla fine con 0x21 (alcune versioni mandano prima degli 00).
// Due somme di controllo (0x00A0 e 0x12D1).
//
// Differenze volute rispetto a CHIRP:
//   - il modo segue la frequenza quando il canale è nuovo o cambia frequenza: FM
//     larga da 76 a 108 MHz, AM da 108 a 137 MHz, FM altrove; il modo di un
//     canale resta se la frequenza non cambia (CHIRP scrive quello che si
//     sceglie a mano nella sua colonna, che qui non c'è);
//   - un canale nuovo mostra il nome se ne ha uno e non ha lo «spostamento
//     dell'orologio» (CHIRP copia i due bit da quello che c'era nella posizione
//     vuota);
//   - frequenze e scostamenti che la radio non può memorizzare vengono
//     rifiutati (CHIRP li tronca); uno scostamento non multiplo di 1 kHz si
//     scrive come frequenza di trasmissione separata;
//   - i caratteri speciali dei nomi che non esistono in ASCII (Δ Υ Σ ¥ e i
//     simboli grafici) si leggono come spazi; restano come sono finché il
//     canale non viene modificato.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type vx1Fam struct{}

const vx1Size = 4818

const vx1Format = `
#seekto 0x0000;
u8 match_mode;
u8 fixed_bits:2,
   wide_rx:1,
   cellular_rx:1,
   wide_tx:1,
   country:3;
#seekto 0x0002;
u8 band;
#seekto 0x0007;
u8 ars:1, atmd:1, resume:1, scnl:1, beep:1, bclo:1, bsyled:1, unknown0:1;
u8 rxsave:3, apo:3, lamp:2;
u8 bell:3, unknown1a:1, lock:3, lk:1;
u8 montc:1, dialm:1, tot:3, smtmd:1, artsbp:2;
u8 unknown2a:1, cwid:1, grp:1, unknown2b:5;
#seekto 0x000e;
u8 dtmfm:4, unknown3:4;
#seekto 0x000f;
u8 volume;
u8 squelch;
#seekto 0x0011;
struct {
  lbcd rx_freq[3];
  lbcd offset[3];
  u8 skip:1,
     name_flag:1,
     mode:2,
     shift:2,
     selcal:2;
  u8 unknown:3,
     power:1,
     clk_shift:1,
     step:3;
  u8 unknown2:2,
     ctcss:6;
  u8 dcs;
  u8 name[6];
} current;
#seekto 0x00a0;
u8 checksum1;
#seekto 0x00f1;
struct {
  lbcd rx_freq[3];
  lbcd offset[3];
  u8 skip:1,
     name_flag:1,
     mode:2,
     shift:2,
     selcal:2;
  u8 unknown:3,
     power:1,
     clk_shift:1,
     step:3;
  u8 unknown2:2,
     ctcss:6;
  u8 dcs;
  u8 name[6];
} smart_search[31];
#seekto 0x02f1;
u8 cwid_text[9];
#seekto 0x0311;
struct {
  u8 freq[2];
  u8 name[6];
} bc_vfo;
#seekto 0x0321;
struct {
  lbcd rx_freq[3];
  lbcd offset[3];
  u8 skip:1,
     name_flag:1,
     mode:2,
     shift:2,
     selcal:2;
  u8 unknown:3,
     power:1,
     clk_shift:1,
     step:3;
  u8 unknown2:2,
     ctcss:6;
  u8 dcs;
  u8 name[6];
} vfos[8];
#seekto 0x03a1;
u8 priority_channel;
#seekto 0x03d1;
struct {
  lbcd rx_freq[3];
  u8 unknown;
  u8 skip:1,
     name_flag:1,
     mode:2,
     shift:2,
     selcal:2;
  u8 unknown2:3,
     power:1,
     clk_shift:1,
     step:3;
  u8 name[6];
} cg2_memory[170];
#seekto 0x0bd1;
struct {
  lbcd rx_freq[3];
  lbcd offset[3];
  u8 skip:1,
     name_flag:1,
     mode:2,
     shift:2,
     selcal:2;
  u8 unknown:3,
     power:1,
     clk_shift:1,
     step:3;
  u8 unknown2:2,
     ctcss:6;
  u8 dcs;
  u8 name[6];
} cg1_memory[80];
#seekto 0x10d1;
lbit band_mask[2048];
#seekto 0x11f1;
struct {
  u8 freq[2];
  u8 name[6];
} bc_memory[10];
#seekto 0x1251;
struct {
  u8 dtmf[16];
} autodialer[8];
#seekto 0x12d1;
u8 checksum2;
`

// nomi: posizione del carattere nell'elenco della radio (63 caratteri; 0x3F =
// fine del nome). I caratteri che non esistono in ASCII sono segnati con 0
const vx1Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/\x00\x00\x00|-?%&_$\x00\\\x00<>\x00\x00\x00\x00"

// vx1Charset: caratteri che si possono scrivere
const vx1Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ ()+-=*/|?%&_$\\<>"

// vx1Bands: le otto bande della radio (indice = bit della maschera di banda)
var vx1Bands = [8][2]int64{{76000000, 108000000}, {108000000, 137000000}, {137000000, 170000000}, {170000000, 222000000},
	{222000000, 420000000}, {420000000, 470000000}, {470000000, 800000000}, {800000000, 999000000}}

func init() {
	genRegister(&genModel{ID: "ys-vx1", Vendor: "Yaesu", Model: "VX-1", RClass: "VX1Radio", Name: "Yaesu VX-1R",
		Slots: 260, First: 1, NameMax: 6, Charset: vx1Charset,
		Powers:  []genPower{{"Bassa 0,2 W", 0.2, 2}, {"Alta 1 W", 1, 0}},
		Bands:   []band{{76000000, 999000000}, {500000, 1700000}},
		HasScan: true, NoTxOff: true, DCS: dtcsCodes, OneDCS: true, Baud: 9600, Fam: vx1Fam{}, Opt: &vx1Opt{},
		Notes: "260 posizioni: 1-52 memorie del gruppo 1 (complete), 53-72 limiti di scansione e 73-80 canali «home» del gruppo 1; " +
			"81-222 memorie del gruppo 2 (senza scostamento proprio né tono), 223-242 limiti di scansione e 243-250 canali «home» del gruppo 2; " +
			"251-260 onde medie (da 0,5 a 1,7 MHz). La radio usa un gruppo alla volta (menu «GRP»). Nomi di 6 caratteri. " +
			"Il modo segue la frequenza: FM larga da 76 a 108 MHz, AM in banda aerea. Le potenze indicate valgono con l'alimentazione esterna (a batteria 0,5 e 0,05 W). " +
			"Per la copia: spegni la radio, collega il cavo alla presa MIC/SP e riaccendila tenendo premuto [FW]. Lettura: avvia la lettura e poi premi [DWN] sulla radio («CLN OUT»). " +
			"Scrittura: premi [UP] sulla radio («CLN IN») e poi avvia la scrittura. Se la radio non accetta la copia va riportata ai valori di fabbrica ([M/V] e [AR] premuti all'accensione, poi [FW]). Driver di CHIRP."})
}

type vx1Opt struct{}

func (vx1Fam) mem() *bwEl                 { return bwMust(vx1Format) }
func (f vx1Fam) layout(m *genModel) *bwEl { return f.mem() }

// noIdent: la copia parte dalla radio, che non risponde a nessuna richiesta
func (vx1Fam) noIdent() {}

// zone: gruppo della posizione i (1 = gruppo 1, 2 = gruppo 2, 0 = onde medie),
// indice nella tabella della radio e indice per la maschera di banda (-1 per i
// canali «home», che non l'hanno)
func (vx1Fam) zone(i int) (grp, idx, mask int) {
	switch {
	case i < 72:
		return 1, i + 8, i
	case i < 80:
		return 1, i - 72, -1
	case i < 242:
		return 2, i - 80 + 8, i - 80
	case i < 250:
		return 2, i - 242, -1
	}
	return 0, i - 250, -1
}

func (f vx1Fam) rec(i int) *bwEl {
	switch grp, idx, _ := f.zone(i); grp {
	case 1:
		return f.mem().F("cg1_memory").I(idx)
	case 2:
		return f.mem().F("cg2_memory").I(idx)
	default:
		return f.mem().F("bc_memory").I(idx)
	}
}

func (f vx1Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(i).Raw(img)...)
}

func (f vx1Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) == f.rec(i).ByteLen() {
		f.rec(i).SetRaw(img, raw)
	}
}

// Match: come CHIRP (lunghezza e la scritta «YAESU» a 0x00E6)
func (vx1Fam) Match(m *genModel, img []byte) bool {
	return len(img) == vx1Size && string(img[0xE6:0xEB]) == "YAESU"
}

// vx1Sums: le due somme di controllo (prima quella a 0x00A0, che entra nella seconda)
func vx1Sums(img []byte) (a, b byte) {
	for _, c := range img[0x0001:0x00A0] {
		a += c
	}
	var s byte
	for k := 0x0001; k <= 0x12D0; k++ {
		if k == 0x00A0 {
			s += a
			continue
		}
		s += img[k]
	}
	return a, -s
}

func (vx1Fam) fixSums(img []byte) {
	a, b := vx1Sums(img)
	img[0x00A0], img[0x12D1] = a, b
}

// vx1Band: banda della frequenza (-1 se fuori dalle otto bande)
func vx1Band(hz int64) int {
	for k, b := range vx1Bands {
		if hz >= b[0] && hz < b[1] {
			return k
		}
	}
	if hz == vx1Bands[7][1] {
		return 7 // 999,000 MHz: l'ultima frequenza della radio
	}
	return -1
}

// setMask: maschera di banda della memoria (un bit per banda; tutti a zero se
// band è -1). In ognuno degli otto blocchi da 32 byte: un byte con il bit
// della banda, 22 byte per il gruppo 2, un altro byte con il bit della banda,
// poi il gruppo 1; i bit si contano dal meno significativo
func (vx1Fam) setMask(img []byte, grp, k, band int) {
	base := 8
	if grp == 1 {
		base = 176 + 8
	}
	for b := 0; b < 8; b++ {
		bit := base + b*256 + k
		a, mk := 0x10D1+bit/8, byte(1)<<(bit%8)
		if b == band {
			img[a] |= mk
		} else {
			img[a] &^= mk
		}
	}
}

// stdOffset: scostamento normale della banda per le memorie del gruppo 2
// (come lo mostra CHIRP)
func vx1StdOffset(hz int64) int64 {
	switch {
	case hz >= 420000000:
		return 5000000
	case hz >= 144000000:
		return 600000
	}
	return 0
}

func vx1Name(raw []byte) string {
	var out []byte
	for _, b := range raw {
		if b == 0x3F {
			break
		}
		c := byte(' ')
		if int(b) < len(vx1Chars) && vx1Chars[b] != 0 {
			c = vx1Chars[b]
		}
		out = append(out, c)
	}
	return strings.TrimRight(string(out), " ")
}

func vx1PutName(m *genModel, name string) []byte {
	out := []byte{0x3F, 0x3F, 0x3F, 0x3F, 0x3F, 0x3F}
	for k, ch := range []byte(strings.TrimRight(m.cleanName(name), " ")) {
		if k >= len(out) {
			break
		}
		p := strings.IndexByte(vx1Chars, ch)
		if p < 0 {
			p = strings.IndexByte(vx1Chars, '_')
		}
		out[k] = byte(p)
	}
	return out
}

// vx1BC: frequenza delle memorie in onde medie (un byte: 0,5 MHz + n·1,2/254)
func vx1BCHz(b byte) int64 {
	return int64((0.5+float64(b)*(1.2/254.0))*1000000 + 0.5)
}

func vx1BCByte(hz int64) byte {
	v := int((float64(hz)/1000000.0-0.5)/(1.2/254.0) + 0.5)
	if v < 0 {
		v = 0
	}
	if v > 0xFE {
		v = 0xFE
	}
	return byte(v)
}

func (vx1Fam) toneOK(t, r genTone) bool { return vx5ToneOK(t, r) }
func (vx1Fam) fitBySet()                {}

// trySlot: le conversioni da altre radio si provano su una memoria completa
func (vx1Fam) trySlot(m *genModel, c *Chan) int {
	if c != nil && c.Rx < 1700000 {
		return 250
	}
	return 0
}

func (f vx1Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	return !vx5ToneOK(t, r)
}

// Locked: ogni zona accetta solo le sue frequenze
func (f vx1Fam) Locked(m *genModel, i int, old, want *Chan) error {
	if want == nil {
		return nil
	}
	grp, _, _ := f.zone(i)
	bc := want.Rx >= 500000 && want.Rx <= 1700000
	switch {
	case grp == 0 && !bc:
		return fmt.Errorf("le posizioni da 251 a 260 sono le memorie delle onde medie (da 0,5 a 1,7 MHz): %s MHz va nelle posizioni da 1 a 250", fmtMHz(want.Rx))
	case grp != 0 && bc:
		return fmt.Errorf("le onde medie (%s MHz) vanno nelle posizioni da 251 a 260", fmtMHz(want.Rx))
	case grp == 2 && (want.TxTone != "" || want.RxTone != ""):
		return fmt.Errorf("le memorie del gruppo 2 (posizioni da 81 a 250) non memorizzano il tono: usa una posizione da 1 a 80")
	}
	if d, std := want.Tx-want.Rx, vx1StdOffset(want.Rx); grp == 2 && d != 0 && (std == 0 || (d != std && d != -std)) {
		return fmt.Errorf("le memorie del gruppo 2 (posizioni da 81 a 250) hanno solo lo scostamento normale della banda (600 kHz in VHF, 5 MHz in UHF): usa una posizione da 1 a 80")
	}
	return nil
}

// roundRx: le onde medie si memorizzano in un byte (passi di circa 4,7 kHz):
// la frequenza scritta è la più vicina, come in CHIRP
func (f vx1Fam) roundRx(m *genModel, i int, hz int64) int64 {
	if grp, _, _ := f.zone(i); grp == 0 && hz >= 500000 && hz <= 1700000 {
		return vx1BCHz(vx1BCByte(hz))
	}
	return hz
}

// placeSkip: una conversione da un'altra radio riempie solo le memorie normali
// dei due gruppi e le onde medie, non i limiti di scansione né i canali «home»
func (vx1Fam) placeSkip(m *genModel, i int) bool {
	return (i >= 52 && i < 80) || (i >= 222 && i < 250)
}

func (f vx1Fam) Get(m *genModel, img []byte, i int) *Chan {
	grp, _, _ := f.zone(i)
	r := f.rec(i)
	raw := r.Raw(img)
	if raw[0] == 0xFF {
		return nil
	}
	if grp == 0 {
		hz := vx1BCHz(raw[0])
		return &Chan{Rx: hz, Tx: hz, Power: m.powerSem(0), Scan: true, Name: vx1Name(raw[2:8])}
	}
	khz := r.F("rx_freq").Int(img) * 1000
	c := &Chan{Rx: yaesuRound(khz)}
	if c.Rx <= 0 {
		c.Rx = khz
	}
	if c.Rx <= 0 {
		return nil
	}
	c.Tx = c.Rx
	shift := r.F("shift").Int(img)
	if grp == 1 {
		offs := r.F("offset").Int(img) * 1000
		switch shift {
		case 1:
			c.Tx = c.Rx - offs
		case 2:
			c.Tx = c.Rx + offs
		case 3:
			if c.Tx = yaesuRound(offs); c.Tx <= 0 {
				c.Tx = offs
			}
		}
		tone, dcs := "88.5", "D023N"
		if k := int(r.F("ctcss").Int(img)); k < len(chirpOldTones) {
			tone = fmtTone(chirpOldTones[k])
		}
		if k := int(r.F("dcs").Int(img)); k < len(dtcsCodes) {
			dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
		}
		switch r.F("selcal").Int(img) {
		case 1:
			c.TxTone = tone
		case 2:
			c.TxTone, c.RxTone = tone, tone
		case 3:
			c.TxTone, c.RxTone = dcs, dcs
		}
	} else {
		switch offs := vx1StdOffset(c.Rx); shift {
		case 1:
			c.Tx = c.Rx - offs
		case 2:
			c.Tx = c.Rx + offs
		}
	}
	if c.Tx <= 0 {
		c.Tx = c.Rx
	}
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Scan = !r.F("skip").Bool(img)
	c.Name = vx1Name(r.F("name").Raw(img))
	return c
}

func (f vx1Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	grp, _, mask := f.zone(i)
	r := f.rec(i)
	if c == nil {
		// come CHIRP: solo il primo byte a FF e la maschera di banda a zero
		raw := r.Raw(img)
		raw[0] = 0xFF
		r.SetRaw(img, raw)
		if mask >= 0 {
			f.setMask(img, grp, mask, -1)
		}
		return nil
	}
	if err := f.Locked(m, i, nil, c); err != nil {
		return err
	}
	if grp == 0 {
		if c.Tx != c.Rx || c.TxTone != "" || c.RxTone != "" {
			return fmt.Errorf("le memorie delle onde medie sono di sola ricezione: niente scostamento né toni")
		}
		raw := append([]byte{vx1BCByte(c.Rx), 0x40}, vx1PutName(m, c.Name)...)
		r.SetRaw(img, raw)
		return nil
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := vx5Tmode(t, rt)
	if tmode < 0 || !vx5ToneOK(t, rt) {
		return fmt.Errorf("la radio non ha questa combinazione di toni (subtono in trasmissione, lo stesso subtono o lo stesso codice DCS nei due versi; mai codici invertiti)")
	}
	khz := c.Rx / 1000
	if back := yaesuRound(khz * 1000); (back != c.Rx && khz*1000 != c.Rx) || khz > 999999 {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz", fmtMHz(c.Rx))
	}
	// senza una delle otto bande la memoria non avrebbe il suo bit nella
	// maschera di banda (CHIRP qui accetta da 1,7 MHz e lascia il bit com'era)
	if vx1Band(c.Rx) < 0 {
		return fmt.Errorf("la radio riceve da 76 a 999 MHz (e le onde medie nelle posizioni da 251 a 260): %s MHz non si può memorizzare", fmtMHz(c.Rx))
	}
	was := r.Raw(img)[0] != 0xFF
	oldRx := int64(0)
	if was {
		if old := f.Get(m, img, i); old != nil {
			oldRx = old.Rx
		}
	}
	// modo: resta se la frequenza non cambia, altrimenti segue la frequenza
	mode := int64(0)
	switch b := vx1Band(c.Rx); {
	case was && oldRx == c.Rx:
		if mode = r.F("mode").Int(img); mode > 2 {
			mode = 0
		}
	case b == 0:
		mode = 1
	case b == 1:
		mode = 2
	}
	// canale nuovo: mostra il nome se ne ha uno e non sposta il clock (i valori
	// che il driver di CHIRP dichiara come predefiniti; CHIRP in realtà riparte
	// dai due bit rimasti nella posizione vuota, vedi TestVX1ChirpZones)
	step := int64(0)
	nameFlag, clk := c.Name != "", false
	if was {
		step = r.F("step").Int(img)
		nameFlag = r.F("name_flag").Bool(img)
		if grp == 2 {
			clk = r.F("clk_shift").Bool(img)
		}
	}
	shift, offs := int64(0), int64(-1)
	d := c.Tx - c.Rx
	if grp == 1 {
		switch {
		case d == 0:
		case d%1000 == 0 && d > -1000000000 && d < 1000000000 && d > -70000000 && d < 70000000:
			shift, offs = 2, d/1000
			if d < 0 {
				shift, offs = 1, -d/1000
			}
		default:
			// frequenza di trasmissione separata
			tk := c.Tx / 1000
			if back := yaesuRound(tk * 1000); (back != c.Tx && tk*1000 != c.Tx) || tk > 999999 || tk <= 0 {
				return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz", fmtMHz(c.Tx))
			}
			shift, offs = 3, tk
		}
	} else {
		std := vx1StdOffset(c.Rx)
		switch {
		case tmode != 0:
			return fmt.Errorf("le memorie del gruppo 2 (posizioni da 81 a 250) non memorizzano il tono: usa una posizione da 1 a 80")
		case d == 0:
			// senza scostamento normale nella banda il verso che c'era resta
			if was && oldRx == c.Rx && std == 0 {
				if shift = r.F("shift").Int(img); shift > 2 {
					shift = 0
				}
			}
		case std != 0 && d == std:
			shift = 2
		case std != 0 && d == -std:
			shift = 1
		default:
			return fmt.Errorf("le memorie del gruppo 2 (posizioni da 81 a 250) hanno solo lo scostamento normale della banda (600 kHz in VHF, 5 MHz in UHF): usa una posizione da 1 a 80")
		}
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	selcal := int64(0)
	if grp == 2 && was {
		selcal = r.F("selcal").Int(img) // il tipo di tono del gruppo 2 resta com'è
	}
	r.F("rx_freq").SetInt(img, khz)
	r.F("name").SetRaw(img, vx1PutName(m, c.Name))
	r.F("mode").SetInt(img, mode)
	r.F("skip").SetBool(img, !c.Scan)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("step").SetInt(img, step)
	r.F("clk_shift").SetBool(img, clk)
	r.F("name_flag").SetBool(img, nameFlag)
	r.F("shift").SetInt(img, shift)
	if grp == 1 {
		// come CHIRP: senza tono il canale torna a 88,5 Hz e DCS 023; lo
		// scostamento di un canale senza scostamento resta quello che c'era
		// (600 kHz se il canale è nuovo)
		ctcss, dcs := int64(8), int64(0)
		switch tmode {
		case 1, 2:
			ctcss = int64(yaesuOldTone(t.Hz10))
			selcal = int64(tmode)
		case 3:
			dcs = int64(icomIndex(dtcsCodes, t.Code))
			selcal = 3
		}
		if offs < 0 {
			offs = 600
			if was {
				offs = r.F("offset").Int(img)
			}
		}
		r.F("offset").SetInt(img, offs)
		r.F("ctcss").SetInt(img, ctcss)
		r.F("dcs").SetInt(img, dcs)
	}
	r.F("selcal").SetInt(img, selcal)
	if mask >= 0 {
		if b := vx1Band(c.Rx); b >= 0 {
			f.setMask(img, grp, mask, b)
		}
	}
	return nil
}

// ---------------------------------------------------------------- copia

// vx1Answer: quello che la radio ha risposto dopo la scrittura, tolti l'eco del
// cavo (tutto quello che è stato inviato) e gli 00 che alcune versioni mandano
// prima della conferma. pending: i byte arrivati finora possono essere l'inizio
// dell'eco, bisogna aspettare ancora
func vx1Answer(img, got []byte) (rest []byte, pending bool) {
	switch {
	case len(got) >= len(img) && string(got[:len(img)]) == string(img):
		got = got[len(img):]
	case len(got) < len(img) && string(got) == string(img[:len(got)]):
		pending = true
	}
	for len(got) > 0 && got[0] == 0x00 {
		got = got[1:]
	}
	return got, pending
}

func (f vx1Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, vx1Size, "In attesa della radio: avvia l'invio sulla radio…")
	var data []byte
	// come CHIRP: si aspetta l'inizio dei dati, poi al massimo due secondi fra un pezzo e l'altro
	for try := 0; try < 30 && len(data) == 0; try++ {
		data = readUpTo(ctx, s.p, 32, time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
	}
	if len(data) == 0 {
		return nil, "", genFail("Collegamento", rErr("la radio non ha inviato nulla: mettila in modalità «CLONE» e premi %s dopo aver avviato la lettura", "[DWN]"))
	}
	for len(data) < vx1Size {
		n := vx1Size - len(data)
		if n > 32 {
			n = 32
		}
		b := readUpTo(ctx, s.p, n, 2*time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(b) == 0 {
			return nil, "", rErr("lettura incompleta: %d byte su %d", len(data), vx1Size)
		}
		data = append(data, b...)
		prog(len(data), vx1Size, "Lettura…")
	}
	if a, b := vx1Sums(data); data[0x00A0] != a || data[0x12D1] != b {
		return nil, "", rErr("somma di controllo errata: ripeti la lettura")
	}
	if !f.Match(m, data) {
		return nil, "", rErr("i dati ricevuti non sono quelli di una VX-1: controlla di aver scelto la radio giusta")
	}
	prog(vx1Size, vx1Size, "Lettura completata")
	return data, "", nil
}

func (f vx1Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida (%d byte)", len(img))
	}
	img = append([]byte{}, img[:vx1Size]...)
	f.fixSums(img)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, vx1Size, "Scrittura…")
	// come CHIRP: tutta la memoria in un blocco solo (cinque secondi a 9600 baud)
	if err := s.send(img); err != nil {
		return "", err
	}
	// poi la conferma (0x21). Con il cavo a un filo prima torna indietro tutto
	// quello che è stato inviato; alcune versioni della radio mandano degli 00
	// prima della conferma. Si raccoglie quello che arriva finché la linea tace
	var got, rest []byte
	for len(got) < vx1Size+64 {
		to := 7 * time.Second
		if len(got) > 0 {
			to = 1500 * time.Millisecond
		}
		b := readUpTo(ctx, s.p, vx1Size+64-len(got), to)
		if ctx.Err() != nil {
			return "", errCancelled
		}
		if len(b) == 0 {
			break
		}
		got = append(got, b...)
		prog(min(len(got), vx1Size), vx1Size, "Scrittura…")
		var pending bool
		if rest, pending = vx1Answer(img, got); !pending && len(rest) > 0 {
			break
		}
	}
	rest, _ = vx1Answer(img, got)
	switch {
	case len(rest) == 0:
		return "", genFail("Collegamento", rErr("la radio non conferma i dati: mettila in modalità «CLONE» e premi %s prima di avviare la scrittura", "[UP] («CLN IN»)"))
	case rest[0] != 0x21:
		return "", rErr("risposta inattesa della radio (%02X) durante la scrittura", rest[0])
	}
	prog(vx1Size, vx1Size, "Scrittura completata")
	return "", nil
}
