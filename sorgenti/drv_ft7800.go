// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ft7800.py, vx170.py (4.34).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.34: Yaesu VX-170 e VX-177 (palmari), FT-7800 / FT-7900 e FT-8900 (veicolari).
// 9600 baud. La copia parte dalla radio (modalità «CLONE»): in lettura la radio
// manda 8 byte con il modello, poi blocchi da 32 byte (VX) o 64 byte (FT) e la
// somma di controllo, ognuno confermato con 06; in scrittura il programma manda
// gli stessi blocchi e la radio li conferma. Il cavo rimanda indietro ogni byte
// inviato. Si gestiscono solo i canali.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare (passi da 6,25 kHz) viene
//     rifiutata invece di essere arrotondata;
//   - uno scostamento che non è multiplo di 50 kHz o supera 12,75 MHz si scrive
//     come frequenza di trasmissione separata (CHIRP lo tronca);
//   - nella gamma aeronautica (108-137 MHz) delle veicolari il canale resta in
//     AM; un canale in AM a cui non si cambia frequenza resta in AM;
//   - la sola ricezione non si può impostare (CHIRP non la conosce per queste
//     radio).
// Come CHIRP: il canale cancellato conserva i suoi dati (si toglie solo il segno
// «in uso»; nella FT-7800 esce anche dai banchi); il salto «P» diventa
// salto normale o nessun salto quando il canale viene modificato.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

type ftxFam struct{}

type ftxOpt struct {
	id    string // primi byte della memoria
	size  int    // byte della memoria (l'ultimo è la somma di controllo)
	blk   int    // byte per blocco
	bytes string // disposizione della memoria
	k89   bool   // FT-8900: nome, salto e scostamento dentro il canale
	banks bool   // FT-7800: banchi da liberare quando si cancella
	am    int64  // valore del campo «mode» per l'AM (-1 = la radio non ha l'AM)
	nfm   int64  // valore del campo «mode» per la banda stretta
	send  string // tasto che avvia l'invio sulla radio
	recv  string // tasto che mette la radio in ricezione
	name  string // nome breve per i messaggi
}

const ftxVX170Format = `
char model[6];
#seekto 0x0168;
struct {
  u8 used:1,
     unknown1:1,
     mode:1,
     unknown2:2,
     duplex:3;
  bbcd freq[3];
  u8 clockshift:1,
     tune_step:3,
     unknown5:1,
     tmode:3;
  bbcd split[3];
  u8 power:2,
     tone:6;
  u8 unknown6:1,
     dtcs:7;
  u8 unknown7[2];
  u8 offset;
  u8 unknown9[3];
} memory[200];
#seekto 0x0F28;
struct {
  u8 name[6];
  u8 enabled:1,
     unknown1:7;
  u8 used:1,
     unknown2:7;
} names[200];
#seekto 0x1768;
struct {
  u8 skip3:2,
     skip2:2,
     skip1:2,
     skip0:2;
} flags[50];
`

const ftx7800Format = `
#seekto 0x04C8;
struct {
  u8 used:1,
     unknown1:1,
     mode:2,
     unknown2:1,
     duplex:3;
  bbcd freq[3];
  u8 clockshift:1,
     tune_step:3,
     unknown5:1,
     tmode:3;
  bbcd split[3];
  u8 power:2,
     tone:6;
  u8 unknown6:1,
     dtcs:7;
  u8 unknown7[2];
  u8 offset;
  u8 unknown9[3];
} memory[1000];
#seekto 0x4988;
struct {
  u8 name[6];
  u8 enabled:1,
     unknown1:7;
  u8 used:1,
     unknown2:7;
} names[1000];
#seekto 0x6c48;
struct {
  u32 bitmap[32];
} bank_channels[20];
#seekto 0x7648;
struct {
  u8 skip0:2,
     skip1:2,
     skip2:2,
     skip3:2;
} flags[250];
#seekto 0x7B48;
u8 checksum;
`

const ftx8900Format = `
#seekto 0x0708;
struct {
  u8 used:1,
     skip:2,
     sub_used:1,
     unknown2:1,
     duplex:3;
  bbcd freq[3];
  u8 mode:2,
     nameused:1,
     unknown4:1,
     power:2,
     tmode:2;
  bbcd split[3];
  u8 unknown5:2,
     tone:6;
  u8 namevalid:1,
     dtcs:7;
  u8 name[6];
} memory[799];
`

// caratteri della radio nell'ordine dei codici (gli spazi oltre il primo sono
// simboli che da qui non si scrivono)
const ftxChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "          " + "*+,- /|      [ ] _"

const ftxCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ *+,-/|[]_"

func init() {
	vxNote := func(cls string) string {
		return "200 canali con nomi di 6 caratteri. La copia parte dalla radio: spegnila, collega il cavo alla presa MIC/SP, accendila tenendo premuto " +
			"[MONI], scegli «CLONE» nel menu e premi [F]: la radio riparte in modalità «CLONE». Per leggere: avvia la lettura qui e poi tieni premuto " +
			"[PTT] per un momento (compare «-TX-»). Per scrivere: premi [MONI] sulla radio (compare «-RX-») e poi avvia la scrittura qui. " +
			"Le frequenze si memorizzano a passi di 5 o 12,5 kHz. " + chirpNoteCh(cls)
	}
	vx := func(id string) *ftxOpt {
		return &ftxOpt{id: id, size: 6057, blk: 32, bytes: ftxVX170Format, am: -1, nfm: 1, send: "[PTT]", recv: "[MONI]", name: "VX-170 / VX-177"}
	}
	hp := []genPower{{"Alta 5 W", 5, 0}, {"Media 2 W", 2, 1}, {"Bassa 0,5 W", 0.5, 2}}
	genRegister(&genModel{ID: "ys-vx170", Vendor: "Yaesu", Model: "VX-170", RClass: "VX170Radio", Name: "Yaesu VX-170",
		Slots: 200, First: 1, NameMax: 6, Charset: ftxCharset, Powers: hp, Bands: []band{{137000000, 174000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ftxFam{}, Opt: vx("AH022$"), Notes: vxNote("VX170Radio")})
	genRegister(&genModel{ID: "ys-vx177", Vendor: "Yaesu", Model: "VX-177", RClass: "VX177Radio", Name: "Yaesu VX-177",
		Slots: 200, First: 1, NameMax: 6, Charset: ftxCharset, Powers: hp, Bands: []band{{420000000, 470000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ftxFam{}, Opt: vx("AH022U"), Notes: vxNote("VX177Radio")})
	mp := []genPower{{"Alta 50 W", 50, 0}, {"Media 20 W", 20, 1}, {"Media 10 W", 10, 2}, {"Bassa 5 W", 5, 3}}
	genRegister(&genModel{ID: "ys-ft7800", Vendor: "Yaesu", Model: "FT-7800/7900", RClass: "FT7800Radio", Name: "Yaesu FT-7800 / FT-7900",
		Slots: 999, First: 1, NameMax: 6, Charset: ftxCharset, Powers: mp, Bands: []band{{108000000, 520000000}, {700000000, 990000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ftxFam{},
		Opt: &ftxOpt{id: "AH016", size: 31561, blk: 64, bytes: ftx7800Format, banks: true, am: 1, nfm: 2, send: "[V/M(MW)]", recv: "[LOW(ACC)]", name: "FT-7800 / FT-7900"},
		Notes: "Radio veicolare. 999 canali con nomi di 6 caratteri. La copia parte dalla radio: spegnila, collega il cavo alla presa DATA, accendila " +
			"tenendo premuto [MHz(PRI)], scegli «F-7 CLONE» con la manopola e tieni premuto [BAND(SET)] finché compare «CLONE». Per leggere: avvia la " +
			"lettura qui e poi premi [V/M(MW)]. Per scrivere: premi [LOW(ACC)] (compare «--RX--») e poi avvia la scrittura qui. " +
			"La potenza alta indicata è quella in VHF; in UHF è più bassa. Le frequenze si memorizzano a passi di 5 o 12,5 kHz. " +
			chirpNoteCh("FT7800Radio")})
	genRegister(&genModel{ID: "ys-ft8900", Vendor: "Yaesu", Model: "FT-8900", RClass: "FT8900Radio", Name: "Yaesu FT-8900",
		Slots: 799, First: 1, NameMax: 6, Charset: ftxCharset, Powers: mp,
		Bands:  []band{{28000000, 29700000}, {50000000, 54000000}, {108000000, 180000000}, {320000000, 480000000}, {700000000, 985000000}},
		Narrow: true, HasScan: true, NoTxOff: true, DCS: dtcsCodes, Baud: 9600, Fam: ftxFam{},
		Opt: &ftxOpt{id: "AH008", size: 14793, blk: 64, bytes: ftx8900Format, k89: true, am: 2, nfm: 1, send: "[V/M] di sinistra", recv: "[LOW] di sinistra", name: "FT-8900"},
		Notes: "Radio veicolare a quattro bande. 799 canali con nomi di 6 caratteri. La copia parte dalla radio: spegnila, collega il cavo alla presa DATA, " +
			"accendila tenendo premuto il tasto [V/M] di sinistra, scegli «CLONE START» con la manopola di destra e premi [SET]: compare «CLONE». " +
			"Per leggere: avvia la lettura qui e poi premi il tasto [V/M] di sinistra. Per scrivere: premi il tasto [LOW] di sinistra (compare " +
			"«CLONE -RX-») e poi avvia la scrittura qui. La potenza alta indicata è quella in VHF; in UHF è più bassa. Le frequenze si memorizzano a passi di 5 o 12,5 kHz. " +
			chirpNoteCh("FT8900Radio")})
}

func ftxo(m *genModel) *ftxOpt { return m.Opt.(*ftxOpt) }

func (ftxFam) mem(m *genModel) *bwEl          { return bwMust(ftxo(m).bytes) }
func (f ftxFam) layout(m *genModel) *bwEl     { return f.mem(m) }
func (f ftxFam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

// skip: campo del salto in scansione del canale i
func (f ftxFam) skip(m *genModel, i int) *bwEl {
	if ftxo(m).k89 {
		return f.rec(m, i).F("skip")
	}
	return f.mem(m).F("flags").I(i / 4).F(fmt.Sprintf("skip%d", i%4))
}

func (f ftxFam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(m, i).Raw(img)...)
	if !ftxo(m).k89 {
		out = append(out, f.mem(m).F("names").I(i).Raw(img)...)
		out = append(out, byte(f.skip(m, i).Int(img)))
	}
	return out
}

func (f ftxFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if ftxo(m).k89 {
		if len(raw) == 16 {
			f.rec(m, i).SetRaw(img, raw)
		}
		return
	}
	if len(raw) != 16+8+1 {
		return
	}
	f.rec(m, i).SetRaw(img, raw[:16])
	f.mem(m).F("names").I(i).SetRaw(img, raw[16:24])
	f.skip(m, i).SetInt(img, int64(raw[24]))
}

func (ftxFam) Match(m *genModel, img []byte) bool {
	o := ftxo(m)
	return len(img) == o.size && bytes.HasPrefix(img, []byte(o.id))
}

func ftxSum(img []byte) byte {
	var s byte
	for _, b := range img[:len(img)-1] {
		s += b
	}
	return s
}

// frequenza: 6 cifre BCD in unità da 10 kHz; nel primo byte 0x80 = +5 kHz e
// 0x40 = +2,5 kHz
func ftxDecFreq(raw []byte) int64 {
	var v int64
	for _, b := range raw {
		v = v*100 + int64(b>>4)*10 + int64(b&0x0F)
	}
	hz := v * 10000
	if hz > 8000000000 {
		hz = hz - 8000000000 + 5000
	}
	if hz > 4000000000 {
		hz = hz - 4000000000 + 2500
	}
	return hz
}

func ftxEncFreq(hz int64) []byte {
	v := hz / 10000
	out := []byte{byte(v/100000%10<<4 | v/10000%10), byte(v/1000%10<<4 | v/100%10), byte(v/10%10<<4 | v%10)}
	if hz%1000 == 500 {
		out[0] |= 0x40
	}
	if hz%10000 >= 5000 {
		out[0] |= 0x80
	}
	return out
}

func ftxFreqOK(hz int64) bool {
	return hz > 0 && hz < 4000000000 && ftxDecFreq(ftxEncFreq(hz)) == hz
}

// off89: scostamento della FT-8900, nei due bit alti degli ultimi quattro
// caratteri del nome (unità da 50 kHz)
func (f ftxFam) off89(img []byte, r *bwEl) int64 {
	var v int64
	for _, b := range r.F("name").Raw(img)[2:6] {
		v = v<<2 | int64(b>>6)
	}
	return v
}

func (f ftxFam) setOff89(img []byte, r *bwEl, v int64) {
	n := append([]byte{}, r.F("name").Raw(img)...)
	for k := 5; k >= 2; k-- {
		n[k] = n[k]&0x3F | byte(v&3)<<6
		v >>= 2
	}
	r.F("name").SetRaw(img, n)
}

// duplex: frequenza di trasmissione e scostamento che CHIRP conserva quando il
// canale diventa simplex
func (f ftxFam) duplex(m *genModel, img []byte, r *bwEl) (tx, kept int64) {
	rx := ftxDecFreq(r.F("freq").Raw(img))
	if ftxo(m).k89 {
		kept = f.off89(img, r) * 50000
	} else {
		kept = r.F("offset").Int(img) * 50000
	}
	switch r.F("duplex").Int(img) {
	case 2:
		return rx - kept, kept
	case 3:
		return rx + kept, kept
	case 4:
		tx = ftxDecFreq(r.F("split").Raw(img))
		return tx, tx
	}
	return rx, kept
}

// tmode: modo dei toni → valore nella radio (VX e FT-7800: 0 nessuno, 1 tono,
// 2 tono e squelch, 3 squelch rovesciato, 4 DCS; FT-8900: 0, 1, 2, 3 DCS)
func (ftxFam) tmDCS(m *genModel) int64 {
	if ftxo(m).k89 {
		return 3
	}
	return 4
}

func (f ftxFam) Get(m *genModel, img []byte, i int) *Chan {
	o := ftxo(m)
	r := f.rec(m, i)
	if !r.F("used").Bool(img) {
		return nil
	}
	c := &Chan{Rx: ftxDecFreq(r.F("freq").Raw(img))}
	c.Tx, _ = f.duplex(m, img, r)
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img)); k < len(chirpTones) {
		tone = fmtTone(chirpTones[k])
	}
	if k := int(r.F("dtcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	switch tm := r.F("tmode").Int(img); {
	case tm == 1:
		c.TxTone = tone
	case tm == 2:
		c.TxTone, c.RxTone = tone, tone
	case tm == f.tmDCS(m):
		c.TxTone, c.RxTone = dcs, dcs
	}
	c.Power = m.powerSem(int(r.F("power").Int(img)))
	c.Narrow = r.F("mode").Int(img) == o.nfm
	c.Scan = f.skip(m, i).Int(img) != 1
	var raw []byte
	if o.k89 {
		if r.F("namevalid").Bool(img) {
			raw = r.F("name").Raw(img)
		}
	} else if n := f.mem(m).F("names").I(i); n.F("used").Bool(img) {
		raw = n.F("name").Raw(img)
	}
	var sb strings.Builder
	for _, b := range raw {
		if o.k89 {
			b &= 0x3F
		}
		if int(b) < len(ftxChars) {
			sb.WriteByte(ftxChars[b])
		}
	}
	c.Name = strings.TrimRight(sb.String(), " ")
	return c
}

// ftxTmode: modo dei toni per la coppia (trasmissione, ricezione); -1 se la
// radio non ha quella combinazione
func (f ftxFam) tmode(m *genModel, t, r genTone) int64 {
	switch {
	case t.Inv || r.Inv:
		return -1
	case t.Mode == "" && r.Mode == "":
		return 0
	case t.Mode == "Tone" && r.Mode == "":
		return 1
	case t.Mode == "Tone" && r.Mode == "Tone" && t.Hz10 == r.Hz10:
		return 2
	case t.Mode == "DTCS" && r.Mode == "DTCS" && t.Code == r.Code:
		if m == nil {
			return 4
		}
		return f.tmDCS(m)
	}
	return -1
}

// ftxSplit: lo scostamento va scritto come frequenza di trasmissione separata
func ftxSplit(d int64) bool {
	if d < 0 {
		d = -d
	}
	return d%50000 != 0 || d > 255*50000
}

// chirpInvalid: combinazioni di toni che la radio non ha e scostamenti che
// CHIRP tronca (qui diventano frequenza di trasmissione separata)
func (f ftxFam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	d := c.Tx - c.Rx
	if d < 0 {
		d = -d
	}
	return f.tmode(m, t, r) < 0 || (ftxSplit(d) && d <= 70000000) || (c.Rx >= 108000000 && c.Rx < 137000000 && ftxo(m).am >= 0)
}

// noIdent: la copia parte dalla radio e la radio non risponde a nessuna
// richiesta, quindi «Identifica» non la prova
func (ftxFam) noIdent() {}

// toneOK, fitBySet: vedi genAdapt
func (f ftxFam) toneOK(t, r genTone) bool { return f.tmode(nil, t, r) >= 0 }
func (ftxFam) fitBySet()                  {}

func (f ftxFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := ftxo(m)
	r := f.rec(m, i)
	if c == nil {
		r.F("used").SetBool(img, false)
		if o.k89 {
			r.F("sub_used").SetBool(img, false)
		}
		if o.banks {
			// come CHIRP: il canale cancellato esce da tutti i banchi
			for b := 0; b < 20; b++ {
				w := f.mem(m).F("bank_channels").I(b).F("bitmap").I(i / 32)
				w.SetInt(img, w.Int(img)&^(1<<uint(31-i%32)))
			}
		}
		return nil
	}
	if c.TxOff {
		return fmt.Errorf("su questa radio la trasmissione non si può disabilitare sul canale")
	}
	t, err := genToneParse(c.TxTone)
	if err != nil {
		return err
	}
	rt, err := genToneParse(c.RxTone)
	if err != nil {
		return err
	}
	tmode := f.tmode(m, t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione o in entrambi i versi, lo stesso DCS " +
			"in entrambi i versi; mai toni diversi nei due versi o codici invertiti)")
	}
	if !ftxFreqOK(c.Rx) {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (passi da 5 o 12,5 kHz)", fmtMHz(c.Rx))
	}
	// valori di CHIRP per un canale nuovo: scostamento 600 kHz, 88,5 Hz, D023
	kept, tone, dcs := int64(600000), int64(8), int64(0)
	was := r.F("used").Bool(img)
	wasAM := false
	if was {
		_, kept = f.duplex(m, img, r)
		tone, dcs = r.F("tone").Int(img), r.F("dtcs").Int(img)
		if int(tone) >= len(chirpTones) {
			tone = 8
		}
		if int(dcs) >= len(dtcsCodes) {
			dcs = 0
		}
		wasAM = o.am >= 0 && r.F("mode").Int(img) == o.am && ftxDecFreq(r.F("freq").Raw(img)) == c.Rx
	}
	find := func(list []int, v int) (int64, error) {
		for k, x := range list {
			if x == v {
				return int64(k), nil
			}
		}
		return 0, fmt.Errorf("tono non disponibile su questa radio")
	}
	if t.Mode == "Tone" {
		if tone, err = find(chirpTones, t.Hz10); err != nil {
			return err
		}
	}
	if t.Mode == "DTCS" {
		if dcs, err = find(dtcsCodes, t.Code); err != nil {
			return err
		}
	}
	dup, offs := int64(0), kept
	split := []byte{0, 0, 0}
	d := c.Tx - c.Rx
	switch {
	case d == 0:
	case d > 70000000 || d < -70000000 || ftxSplit(d):
		if !ftxFreqOK(c.Tx) {
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz (passi da 5 o 12,5 kHz)", fmtMHz(c.Tx))
		}
		dup, split = 4, ftxEncFreq(c.Tx)
	case d > 0:
		dup, offs = 3, d
	default:
		dup, offs = 2, -d
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	r.F("used").SetBool(img, true)
	r.F("freq").SetRaw(img, ftxEncFreq(c.Rx))
	r.F("tone").SetInt(img, tone)
	r.F("tmode").SetInt(img, tmode)
	mode := int64(0)
	switch {
	case o.am >= 0 && (wasAM || (c.Rx >= 108000000 && c.Rx < 137000000)):
		mode = o.am
	case c.Narrow:
		mode = o.nfm
	}
	r.F("mode").SetInt(img, mode)
	r.F("dtcs").SetInt(img, dcs)
	if !o.k89 && !was {
		r.F("tune_step").SetInt(img, 0)
	}
	r.F("duplex").SetInt(img, dup)
	r.F("split").SetRaw(img, split)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("unknown5").SetInt(img, 0)
	name := m.cleanName(c.Name)
	has := strings.TrimSpace(name) != ""
	enc := make([]byte, 6)
	for k := range enc {
		ch := byte(' ')
		if k < len(name) {
			ch = name[k]
		}
		p := strings.IndexByte(ftxChars, ch)
		if p < 0 {
			p = strings.IndexByte(ftxChars, ' ')
		}
		enc[k] = byte(p)
	}
	skip := int64(1)
	if c.Scan {
		skip = 0
	}
	if o.k89 {
		r.F("name").SetRaw(img, enc)
		r.F("namevalid").SetBool(img, true)
		r.F("nameused").SetBool(img, has)
		if dup != 4 {
			f.setOff89(img, r, offs/50000&0xFF)
		}
		r.F("skip").SetInt(img, skip)
		r.F("sub_used").SetBool(img, c.Rx >= 108000000 && c.Rx <= 480000000)
		return nil
	}
	n := f.mem(m).F("names").I(i)
	if has {
		n.F("name").SetRaw(img, enc)
	}
	n.F("used").SetBool(img, has)
	n.F("enabled").SetBool(img, has)
	if dup != 4 {
		r.F("offset").SetInt(img, offs/50000&0xFF)
	}
	f.skip(m, i).SetInt(img, skip)
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

// ---------------------------------------------------------------- protocollo

func (f ftxFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := ftxo(m)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.size, "In attesa della radio: avvia l'invio sulla radio…")
	// la radio comincia quando si preme il tasto: 8 byte con il modello
	var data []byte
	for try := 0; try < 30 && len(data) == 0; try++ {
		b := readUpTo(ctx, s.p, 8, time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(b) == 8 {
			data = b
		} else if len(b) > 0 {
			return nil, "", rErr("risposta della radio non valida (% X)", b)
		}
	}
	if len(data) == 0 {
		return nil, "", genFail("Collegamento", rErr("la radio non ha inviato nulla: mettila in modalità «CLONE» e premi %s dopo aver avviato la lettura", o.send))
	}
	if !bytes.HasPrefix(data, []byte(o.id)) {
		for _, x := range genModels {
			if _, ok := x.Fam.(ftxFam); ok && bytes.HasPrefix(data, []byte(ftxo(x).id)) {
				return nil, "", rErr("la radio è un altro modello: apri la scheda «%s»", x.Name)
			}
		}
		return nil, "", rErr("la radio non è una %s (identificativo %q)", o.name, data[:6])
	}
	if err := s.ft60Send([]byte{0x06}); err != nil {
		return nil, "", err
	}
	for len(data) < o.size {
		prog(len(data), o.size, "Lettura…")
		want := o.blk
		if o.size-1-len(data) < want {
			want = o.size - 1 - len(data)
		}
		if want == 0 {
			want = 1 // somma di controllo
		}
		b := readUpTo(ctx, s.p, want, 2*time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(b) != want {
			return nil, "", rErr("blocco incompleto all'indirizzo %d (%d byte su %d)", len(data), len(b), want)
		}
		data = append(data, b...)
		if err := s.ft60Send([]byte{0x06}); err != nil {
			return nil, "", err
		}
	}
	if ftxSum(data) != data[o.size-1] {
		return nil, "", rErr("somma di controllo errata: ripeti la lettura")
	}
	prog(o.size, o.size, "Lettura completata")
	return data, "", nil
}

func (f ftxFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := ftxo(m)
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida per la %s", o.name)
	}
	img = append([]byte{}, img...)
	img[o.size-1] = ftxSum(img)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, o.size, "Collegamento con la radio…")
	ack := func() error {
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return rErr("la radio non conferma i dati: mettila in modalità «CLONE» e premi %s prima di avviare la scrittura", o.recv)
		}
		if b[0] != 0x06 {
			return rErr("risposta inattesa della radio (%02X) durante la scrittura", b[0])
		}
		return nil
	}
	if err := s.ft60Send(img[:8]); err != nil {
		return "", genFail("Collegamento", err)
	}
	if err := ack(); err != nil {
		return "", genFail("Collegamento", err)
	}
	for a := 8; a < o.size; {
		e := a + o.blk
		if e > o.size-1 {
			e = o.size - 1
		}
		if a == o.size-1 {
			e = o.size // somma di controllo
		}
		prog(a, o.size, "Scrittura…")
		if err := s.ft60Send(img[a:e]); err != nil {
			return "", err
		}
		if err := ack(); err != nil {
			return "", err
		}
		// la radio non ha controllo di flusso (CHIRP aspetta 50 ms)
		if err := s.sleep(ftxBlockDelay); err != nil {
			return "", err
		}
		a = e
	}
	prog(o.size, o.size, "Scrittura completata")
	return "", nil
}

// ftxBlockDelay: pausa fra i blocchi in scrittura
var ftxBlockDelay = 50 * time.Millisecond
