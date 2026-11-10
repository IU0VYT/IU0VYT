// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// ft60.py (4.14).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.14: Yaesu FT-60R/E (cavo CT-91 o compatibile nella presa MIC/SP, 9600 baud).
// La copia parte dalla radio (modalità «F8 CLONE»): in lettura la radio manda 8
// byte, poi 447 blocchi da 64 byte e la somma di controllo, ognuno confermato
// con 06; in scrittura il programma manda gli stessi blocchi e la radio li
// conferma. Il cavo rimanda indietro ogni byte inviato.
//
// Differenze volute rispetto a CHIRP (controllate nelle prove):
//   - una frequenza che la radio non può memorizzare (es. passi da 6,25 kHz) viene
//     rifiutata invece di essere arrotondata;
//   - DCS in trasmissione con subtono in ricezione: il subtono scritto è quello
//     di ricezione (CHIRP lascia il valore precedente);
//   - uno scostamento che non è multiplo di 50 kHz si scrive come frequenza di
//     trasmissione separata;
//   - nella gamma aeronautica (108-137 MHz) il canale resta in AM.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

type ft60Fam struct{}

const ft60Format = `
struct mem {
  u8 used:1,
     unknown1:1,
     isnarrow:1,
     isam:1,
     duplex:4;
  bbcd freq[3];
  u8 unknown2:1,
     step:3,
     unknown2_1:1,
     tmode:3;
  bbcd tx_freq[3];
  u8 power:2,
     tone:6;
  u8 unknown4:1,
     dtcs:7;
  u8 unknown5;
  u16 unknown5_1:1,
      offset:15;
  u8 unknown6[3];
};
#seekto 0x0248;
struct mem memory[1000];
#seekto 0x40c8;
struct mem pms[100];
#seekto 0x6EC8;
struct {
  u8 skip3:2,
     skip2:2,
     skip1:2,
     skip0:2;
} flags[250];
#seekto 0x4708;
struct {
  u8 name[6];
  u8 use_name:1,
     unknown1:7;
  u8 valid:1,
     unknown2:7;
} names[1000];
#seekto 0x6FC8;
u8 checksum;
`

const (
	ft60Size = 28617
	// caratteri della radio nell'ordine dei codici (le posizioni di «`», «o» e «u»
	// sono simboli che non si scrivono da qui)
	ft60Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ !`o$%&'()*+,-./|;/=>?@[u]^_"
)

var ft60Steps = []int64{5000, 10000, 12500, 15000, 20000, 25000, 50000, 100000}

func init() {
	genRegister(&genModel{ID: "ys-ft60", Vendor: "Yaesu", Model: "FT-60", RClass: "FT60Radio", Name: "Yaesu FT-60R / FT-60E",
		Slots: 1000, First: 1, NameMax: 6, Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ !$%&'()*+,-./|;=>?@[]^_",
		Powers: []genPower{{"Alta 5 W", 5, 0}, {"Media 2 W", 2, 1}, {"Bassa 0,5 W", 0.5, 2}},
		Bands:  []band{{108000000, 520000000}, {700000000, 999990000}}, Narrow: true, HasScan: true, DCS: dtcsCodes, Baud: 9600, Fam: ft60Fam{},
		Notes: "1000 canali con nomi di 6 caratteri. La copia parte dalla radio: spegnila, collega il cavo alla presa MIC/SP, accendila tenendo premuto " +
			"[MONI], scegli «F8 CLONE» con la manopola e premi [F/W]. Per leggere: avvia la lettura qui e poi tieni premuto [PTT] per un secondo. " +
			"Per scrivere: premi [MONI] sulla radio (compare «--RX--») e poi avvia la scrittura qui. Banchi, canali PMS e passo di sintonia si conservano. " +
			"Le frequenze si memorizzano a passi di 5 o 12,5 kHz. Driver di CHIRP."})
}

func (ft60Fam) mem() *bwEl                 { return bwMust(ft60Format) }
func (f ft60Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f ft60Fam) rec(i int) *bwEl          { return f.mem().F("memory").I(i) }
func (f ft60Fam) skip(i int) *bwEl         { return f.mem().F("flags").I(i / 4).F(fmt.Sprintf("skip%d", i%4)) }

func (f ft60Fam) Rec(m *genModel, img []byte, i int) []byte {
	out := append([]byte{}, f.rec(i).Raw(img)...)
	out = append(out, f.mem().F("names").I(i).Raw(img)...)
	return append(out, byte(f.skip(i).Int(img)))
}

func (f ft60Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	if len(raw) != 16+8+1 {
		return
	}
	f.rec(i).SetRaw(img, raw[:16])
	f.mem().F("names").I(i).SetRaw(img, raw[16:24])
	f.skip(i).SetInt(img, int64(raw[24]))
}

func (ft60Fam) Match(m *genModel, img []byte) bool {
	return len(img) == ft60Size && bytes.HasPrefix(img, []byte("AH017"))
}

func ft60Checksum(img []byte) {
	if len(img) == ft60Size {
		var s byte
		for _, b := range img[:0x6FC8] {
			s += b
		}
		img[0x6FC8] = s
	}
}

// frequenza: 6 cifre BCD in unità da 10 kHz; nel primo byte 0x80 = +5 kHz e
// 0x40 = passo da 12,5 kHz (si aggiungono 2,5 kHz fino al multiplo di 12,5)
func ft60DecFreq(raw []byte) int64 {
	var v int64
	for _, b := range raw {
		v = v*100 + int64(b>>4)*10 + int64(b&0x0F)
	}
	hz := v * 10000
	if hz > 8000000000 {
		hz = hz - 8000000000 + 5000
	}
	if hz > 4000000000 {
		hz -= 4000000000
		for k := 0; k < 3; k++ {
			hz += 2500
			if hz%5000 != 0 && hz%12500 == 0 {
				break
			}
		}
	}
	return hz
}

func ft60EncFreq(hz int64) []byte {
	v := hz / 10000
	out := []byte{byte(v/100000%10<<4 | v/10000%10), byte(v/1000%10<<4 | v/100%10), byte(v/10%10<<4 | v%10)}
	if (hz/1000)%10 >= 5 {
		out[0] |= 0x80
	}
	if hz%5000 != 0 && (hz%12500 == 0 || hz%6250 == 0) {
		out[0] |= 0x40
	}
	return out
}

func ft60FreqOK(hz int64) bool { return hz > 0 && ft60DecFreq(ft60EncFreq(hz)) == hz }

func ft60Zero(b []byte) bool { return bytes.Equal(b, make([]byte, len(b))) }

// duplex: frequenza di trasmissione, trasmissione disabilitata e scostamento che
// CHIRP conserva quando il canale diventa simplex
func (f ft60Fam) duplex(img []byte, r *bwEl) (tx int64, off bool, kept int64) {
	rx := ft60DecFreq(r.F("freq").Raw(img))
	kept = r.F("offset").Int(img) * 50000
	switch r.F("duplex").Int(img) {
	case 2:
		return rx - kept, false, kept
	case 3:
		return rx + kept, false, kept
	case 4:
		if ft60Zero(r.F("tx_freq").Raw(img)) {
			return rx, true, kept
		}
		tx = ft60DecFreq(r.F("tx_freq").Raw(img))
		return tx, false, tx
	case 5:
		return rx, true, kept
	}
	return rx, false, kept
}

func (f ft60Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(i)
	if !r.F("used").Bool(img) {
		return nil
	}
	c := &Chan{Rx: ft60DecFreq(r.F("freq").Raw(img))}
	c.Tx, c.TxOff, _ = f.duplex(img, r)
	tone, dcs := "88.5", "D023N"
	if k := int(r.F("tone").Int(img)); k < len(chirpTones) {
		tone = fmtTone(chirpTones[k])
	}
	if k := int(r.F("dtcs").Int(img)); k < len(dtcsCodes) {
		dcs = fmt.Sprintf("D%03dN", dtcsCodes[k])
	}
	switch r.F("tmode").Int(img) {
	case 1:
		c.TxTone = tone
	case 2:
		c.TxTone, c.RxTone = tone, tone
	case 4:
		c.TxTone, c.RxTone = dcs, dcs
	case 5:
		c.TxTone = dcs
	case 6:
		c.TxTone, c.RxTone = tone, dcs
	case 7:
		c.TxTone, c.RxTone = dcs, tone
	}
	p := int(r.F("power").Int(img))
	if p > 2 {
		p = 0
	}
	c.Power = m.powerSem(p)
	c.Narrow = r.F("isnarrow").Bool(img) && !r.F("isam").Bool(img)
	c.Scan = f.skip(i).Int(img) != 1
	n := f.mem().F("names").I(i)
	if n.F("use_name").Bool(img) && n.F("valid").Bool(img) {
		var sb strings.Builder
		for _, b := range n.F("name").Raw(img) {
			if b == 0xFF {
				break
			}
			if int(b) < len(ft60Chars) {
				sb.WriteByte(ft60Chars[b])
			}
		}
		c.Name = strings.TrimRight(sb.String(), " ")
	}
	return c
}

// ft60Tmode: modo dei toni della radio per la coppia (trasmissione, ricezione);
// -1 se la radio non ha quella combinazione
func ft60Tmode(t, r genTone) int64 {
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
		return 4
	case t.Mode == "DTCS" && r.Mode == "":
		return 5
	case t.Mode == "Tone" && r.Mode == "DTCS":
		return 6
	case t.Mode == "DTCS" && r.Mode == "Tone":
		return 7
	}
	return -1
}

// chirpInvalid: combinazioni che la radio non ha, più quella che CHIRP scrive
// in modo diverso (DCS in trasmissione con subtono in ricezione)
func (ft60Fam) chirpInvalid(m *genModel, c *Chan) bool {
	t, _ := genToneParse(c.TxTone)
	r, _ := genToneParse(c.RxTone)
	k := ft60Tmode(t, r)
	return k < 0 || k == 7
}

// noIdent (4.14): la copia parte dalla radio ([PTT] in «F8 CLONE») e la radio non
// risponde a nessuna richiesta, quindi «Identifica» non la prova
func (ft60Fam) noIdent() {}

// toneOK (4.14): combinazioni di toni che la radio ha (vedi genAdapt)
func (ft60Fam) toneOK(t, r genTone) bool { return ft60Tmode(t, r) >= 0 }

func (f ft60Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := f.rec(i)
	if c == nil {
		r.F("used").SetBool(img, false)
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
	tmode := ft60Tmode(t, rt)
	if tmode < 0 {
		return fmt.Errorf("la radio non ha questa combinazione di toni (ammessi: subtono in trasmissione o in entrambi i versi, DCS in " +
			"trasmissione o in entrambi i versi, subtono e DCS misti; mai il solo tono in ricezione o codici invertiti)")
	}
	if !ft60FreqOK(c.Rx) {
		return fmt.Errorf("la radio non può memorizzare la frequenza %s MHz (passi da 5 o 12,5 kHz)", fmtMHz(c.Rx))
	}
	kept, tone, dcs := int64(600000), int64(8), int64(0) // valori di CHIRP per un canale nuovo
	was := r.F("used").Bool(img)
	if was {
		_, _, kept = f.duplex(img, r)
		tone, dcs = r.F("tone").Int(img), r.F("dtcs").Int(img)
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
	if rt.Mode == "Tone" && tmode == 7 {
		if tone, err = find(chirpTones, rt.Hz10); err != nil {
			return err
		}
	}
	for _, g := range []genTone{t, rt} {
		if g.Mode == "DTCS" {
			if dcs, err = find(dtcsCodes, g.Code); err != nil {
				return err
			}
		}
	}
	dup, offs := int64(0), kept
	var txf []byte = []byte{0, 0, 0}
	d := c.Tx - c.Rx
	switch {
	case c.TxOff:
		dup, offs = 5, 0
	case d == 0:
	case d > 70000000 || d < -70000000 || d%50000 != 0:
		if !ft60FreqOK(c.Tx) {
			return fmt.Errorf("la radio non può memorizzare la frequenza di trasmissione %s MHz (passi da 5 o 12,5 kHz)", fmtMHz(c.Tx))
		}
		dup, offs, txf = 4, 0, ft60EncFreq(c.Tx)
	case d > 0:
		dup, offs = 3, d
	default:
		dup, offs = 2, -d
	}
	// da qui in poi nessun errore: il canale si scrive tutto o per niente
	if !was {
		r.Fill(img, 0)
		r.F("used").SetBool(img, true)
	}
	r.F("freq").SetRaw(img, ft60EncFreq(c.Rx))
	r.F("tx_freq").SetRaw(img, txf)
	r.F("offset").SetInt(img, offs/50000)
	r.F("duplex").SetInt(img, dup)
	r.F("tmode").SetInt(img, tmode)
	r.F("tone").SetInt(img, tone)
	r.F("dtcs").SetInt(img, dcs)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	r.F("isnarrow").SetBool(img, c.Narrow)
	r.F("isam").SetBool(img, c.Rx >= 108000000 && c.Rx < 137000000)
	if c.Scan {
		f.skip(i).SetInt(img, 0)
	} else {
		f.skip(i).SetInt(img, 1)
	}
	name := m.cleanName(c.Name)
	n := f.mem().F("names").I(i)
	enc := make([]byte, 6)
	for k := range enc {
		ch := byte(' ')
		if k < len(name) {
			ch = name[k]
		}
		p := strings.IndexByte(ft60Chars, ch)
		if p < 0 {
			p = strings.IndexByte(ft60Chars, ' ')
		}
		enc[k] = byte(p)
	}
	n.F("name").SetRaw(img, enc)
	has := strings.TrimSpace(name) != ""
	n.F("use_name").SetBool(img, has)
	n.F("valid").SetBool(img, has)
	return nil // la somma di controllo si aggiorna quando si scrive nella radio (come CHIRP)
}

// ---------------------------------------------------------------- protocollo

// invio con lettura dell'eco del cavo
func (s *rawSess) ft60Send(b []byte) error {
	if err := s.send(b); err != nil {
		return err
	}
	echo := readUpTo(s.ctx, s.p, len(b), 2*time.Second)
	if s.ctx.Err() != nil {
		return errCancelled
	}
	if len(echo) == 0 {
		return errFt4NoEcho
	}
	if !bytes.Equal(echo, b) {
		return rErr("il cavo non rimanda indietro i dati inviati come dovrebbe: cavo non adatto o disturbi")
	}
	return nil
}

func (f ft60Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, ft60Size, "In attesa della radio: tieni premuto [PTT] per un secondo…")
	// la radio comincia quando si preme [PTT]: 8 byte con il modello
	var data []byte
	for try := 0; try < 20 && len(data) == 0; try++ {
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
		return nil, "", genFail("Collegamento", rErr("la radio non ha inviato nulla: mettila in «F8 CLONE» e tieni premuto [PTT] dopo aver avviato la lettura"))
	}
	if !bytes.HasPrefix(data, []byte("AH017")) {
		return nil, "", rErr("la radio non è una FT-60 (identificativo %q)", data[:5])
	}
	if err := s.ft60Send([]byte{0x06}); err != nil {
		return nil, "", err
	}
	for k := 0; k < 448; k++ {
		prog(len(data), ft60Size, "Lettura…")
		want := 64
		if k == 447 {
			want = 1
		}
		b := readUpTo(ctx, s.p, want, 2*time.Second)
		if ctx.Err() != nil {
			return nil, "", errCancelled
		}
		if len(b) != want {
			return nil, "", rErr("blocco %d incompleto (%d byte)", k, len(b))
		}
		data = append(data, b...)
		if err := s.ft60Send([]byte{0x06}); err != nil {
			return nil, "", err
		}
	}
	var sum byte
	for _, b := range data[:0x6FC8] {
		sum += b
	}
	if sum != data[0x6FC8] {
		return nil, "", rErr("somma di controllo errata: ripeti la lettura")
	}
	prog(ft60Size, ft60Size, "Lettura completata")
	return data, "", nil
}

func (f ft60Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if !f.Match(m, img) {
		return "", fmt.Errorf("immagine della memoria non valida per la FT-60")
	}
	img = append([]byte{}, img...)
	ft60Checksum(img)
	s, err := rawOpen(ctx, port, m.Baud)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, ft60Size, "Collegamento con la radio…")
	ack := func() error {
		b := readUpTo(ctx, s.p, 1, 2*time.Second)
		if ctx.Err() != nil {
			return errCancelled
		}
		if len(b) == 0 {
			return rErr("la radio non conferma i dati: mettila in «F8 CLONE» e premi [MONI] («--RX--») prima di avviare la scrittura")
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
	for k := 0; k < 448; k++ {
		a := 8 + k*64
		e := a + 64
		if e > len(img) {
			e = len(img)
		}
		prog(a, ft60Size, "Scrittura…")
		if err := s.ft60Send(img[a:e]); err != nil {
			return "", err
		}
		if err := ack(); err != nil {
			return "", err
		}
	}
	prog(ft60Size, ft60Size, "Scrittura completata")
	return "", nil
}
