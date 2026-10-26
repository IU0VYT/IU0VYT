// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: radtel_rt900.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// ottobre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// Bajeton BJ7800 (driver radtel_rt900.py di CHIRP, classe BJ7800, derivata
// dalla Radtel RT-920; stesso apparato della SHU H28Y Pro) e Bajeton BJ-7700
// (Radtel RT-910: il programma del costruttore «BJ-7700_EnCPS» è il
// «BT-RT910_CPS», per «RT910» e «BJ-7700»; classi RT910BT e RT910 di CHIRP).
// 960 canali in 15 zone da 64, record di 32 byte, 57600 baud.
//
// Protocollo (mml_jc8810 + radtel_rt900 di CHIRP):
//   "PROGRAMBT11U" (BJ-7800) o "PROGRAMBT80U" (BJ-7700) → 0x06 (fino a 5 tentativi);
//   'F' → identificativo di 8 byte, seguito da un secondo identificativo di 8 byte;
//   pacchetto "SEND …" (impostazione della cifratura) → 0x06;
//   lettura: 'R' indirizzo(2, BE) 0x40 → 'R' indirizzo 0x40 + 64 byte;
//   scrittura: 'W' indirizzo 0x40 + 64 byte → 0x06;
//   uscita: 'E'.
// Dati cifrati con btCrypt (chiave "CO 7") sotto l'indirizzo 0xF000 (sopra c'è
// la taratura, letta in chiaro e mai riscritta).
// L'immagine copre 0x0000-0xF27F: CHIRP legge a blocchi di 64 byte fino a
// 0xF250 e l'ultimo blocco va oltre.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

const rt900Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:4,
     scode:4;
  u8 unknown2:6,
     pttid:2;
  u8 scramble:6,
     txpower:2;
  u8 unknown4:1,
     narrow:1,
     encrypt:2,
     bcl:1,
     scan:1,
     am_modulation:1,
     learning:1;
  ul24 code;
  u8 code_flag;
  char name[12];
} memory[%d];
`

// BJ-9000 (Radtel RT-950 Pro, programma «BT-RT950PRO_CPS»): stesso record di
// 32 byte, ma nel byte 15 l'apprendimento è il bit 7, la modulazione AM il bit 0
// e il bit 1 vale sempre 1.
const rt950Format = `
struct {
  lbcd rxfreq[4];
  lbcd txfreq[4];
  ul16 rxtone;
  ul16 txtone;
  u8 unknown1:4,
     scode:4;
  u8 unknown2:6,
     pttid:2;
  u8 scramble:6,
     txpower:2;
  u8 learning:1,
     narrow:1,
     encrypt:2,
     bcl:1,
     scan:1,
     txen:1,
     am_modulation:1;
  ul24 code;
  u8 code_flag;
  char name[12];
} memory[%d];
`

type rt900Opt struct {
	magic   []byte
	fp1     [][]byte // risposta al comando 'F' (nil: non controllata)
	fp2     [][]byte // secondo identificativo
	crypt   []byte   // pacchetto "SEND …"
	memsize int      // area letta (CHIRP _memsize)
	ranges  [][2]int // aree scritte (CHIRP _ranges, nell'ordine)
	airband band     // gamma aeronautica: canali in AM
	format  string   // record dei canali (predefinito rt900Format)
	model   string   // risposta attesa al comando 'M' (12 byte, BJ-9000)
	block   int      // byte per blocco (predefinito 0x40)
	reads   [][2]int // aree lette (nil: da 0 a memsize)
	txen    bool     // bit "txen" sempre a 1 (BJ-9000)
}

func (o *rt900Opt) blk() int {
	if o.block == 0 {
		return 0x40
	}
	return o.block
}

type rt900Fam struct{}

func init() {
	o := &rt900Opt{
		magic: []byte("PROGRAMBT11U"),
		fp1:   [][]byte{{0x01, 0x36, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20}, {0x01, 0x00, 0x01, 0x80, 0x04, 0x00, 0x05, 0x20}},
		fp2:   [][]byte{{0x02, 0x00, 0x02, 0x60, 0x01, 0x03, 0x30, 0x04}},
		crypt: append([]byte("SEND \x01\x01"), make([]byte, 18)...),
		// 0x0000-0x77FF: 15 zone da 64 canali di 32 byte
		ranges: [][2]int{{0x0000, 0x7800}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xB000, 0xB400},
			{0xC000, 0xC400}, {0xA000, 0xA140}, {0xD000, 0xD040}},
		memsize: 0xF250,
		airband: band{108000000, 136975000},
	}
	genRegister(&genModel{ID: "bj-7800", Vendor: "Bajeton", Model: "BJ7800", RClass: "BJ7800",
		Name: "Bajeton BJ-7800", Slots: 960, First: 1, NameMax: 12, Zones: 15,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato, su un canale libero. " +
			"Canali 1-960 = 15 zone da 64 (zona 1: canali 1-64, zona 2: 65-128, …). " +
			"Driver di CHIRP per il firmware V0.15P."})

	// BJ-7700 = Radtel RT-910 (stesso protocollo della RT-900 BT, memoria della RT-910)
	o910 := &rt900Opt{
		magic:   []byte("PROGRAMBT80U"),
		fp1:     o.fp1,
		fp2:     o.fp2,
		crypt:   o.crypt,
		ranges:  [][2]int{{0x0000, 0x7800}, {0x8000, 0x8040}, {0x9000, 0x9040}, {0xA000, 0xA140}, {0xD000, 0xD040}},
		memsize: 0xF250,
		airband: o.airband,
	}
	for _, v := range []struct{ id, model, rclass, name, fw string }{
		{"bj-7700", "RT-910_BT", "RT910BT", "Bajeton BJ-7700 (Bluetooth)", "V0.24P della RT-910 BT"},
		{"bj-7700nbt", "RT-910", "RT910", "Bajeton BJ-7700 (senza Bluetooth)", "V0.11 della RT-910"},
	} {
		genRegister(&genModel{ID: v.id, Vendor: "Radtel", Model: v.model, RClass: v.rclass,
			Name: v.name, Slots: 960, First: 1, NameMax: 12, Zones: 15,
			Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
			Powers:  []genPower{{"Alta", 8, 0}, {"Media", 4, 1}, {"Bassa", 1, 2}},
			Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
			DCS: uv5rDTCS, Baud: 57600, Fam: rt900Fam{}, Opt: o910,
			Notes: "Accendi la radio con il cavo collegato, su un canale libero. " +
				"Canali 1-960 = 15 zone da 64 (zona 1: canali 1-64, zona 2: 65-128, …). " +
				"La BJ-7700 è la Radtel RT-910 (il programma del costruttore è lo stesso): driver di CHIRP per il firmware " + v.fw + "; " +
				"nei file .img compare come Radtel " + v.model + "."})
	}
}

func init() {
	// BJ-9000 = Radtel RT-950 Pro (programma del costruttore «BT-RT950PRO_CPS» v1.1.1):
	// 115200 baud, "PROGRAMBT9000U", 'F' (16 byte), 'M' → "RT-950", blocchi da 128 byte.
	// Si scrivono solo i canali (0x0000-0x77FF): le altre aree restano quelle della radio.
	o := &rt900Opt{
		magic:   []byte("PROGRAMBT9000U"),
		crypt:   append([]byte("SEND \x01\x01"), make([]byte, 18)...),
		model:   "RT-950",
		block:   0x80,
		reads:   [][2]int{{0x0000, 0x7800}, {0x8000, 0x8080}, {0x9000, 0x9080}, {0xA000, 0xA180}, {0xB000, 0xB100}, {0xD000, 0xD300}},
		ranges:  [][2]int{{0x0000, 0x7800}},
		memsize: 0xD300,
		airband: band{108000000, 136975000},
		format:  rt950Format,
		txen:    true,
	}
	genRegister(&genModel{ID: "bj-9000", Vendor: "Bajeton", Model: "BJ-9000", RClass: "RT950Pro", NoChirp: true,
		Name: "Bajeton BJ-9000", Slots: 960, First: 1, NameMax: 12, Zones: 15,
		Charset: charsetAlnum + "`~!@#$%^&*()-=_+[]\\{}|;':\",./<>?",
		Powers:  []genPower{{"Alta 8 W", 8, 0}, {"Media 4 W", 4, 1}, {"Bassa 1 W", 1, 2}},
		Bands:   []band{{18000000, 1000000000}}, Narrow: true, HasScan: true,
		DCS: uv5rDTCS, Baud: 115200, Fam: rt900Fam{}, Opt: o,
		Notes: "Accendi la radio con il cavo collegato. Canali 1-960 = 15 zone da 64. " +
			"Ricavato dal programma del costruttore «BJ-9000_CPS» v1.1.1 (la BJ-9000 è la Radtel RT-950 Pro): " +
			"si scrivono solo i canali, le impostazioni e l'APRS restano quelli della radio. Non compatibile con CHIRP."})
}

// rt900Identify: ingresso in programmazione e uscita (identificazione della radio)
func rt900Identify(ctx context.Context, m *genModel, port string) error {
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return err
	}
	defer s.p.Close()
	err = s.enter()
	s.exit()
	if err != nil {
		return genFail("Identificazione", err)
	}
	return nil
}

func rt900o(m *genModel) *rt900Opt { return m.Opt.(*rt900Opt) }

func (rt900Fam) mem(m *genModel) *bwEl {
	f := rt900o(m).format
	if f == "" {
		f = rt900Format
	}
	return bwMust(fmt.Sprintf(f, m.Slots))
}

func (f rt900Fam) layout(m *genModel) *bwEl { return f.mem(m) }

func (f rt900Fam) rec(m *genModel, i int) *bwEl { return f.mem(m).F("memory").I(i) }

func (f rt900Fam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, f.rec(m, i).Raw(img)...)
}

func (f rt900Fam) PutRec(m *genModel, img []byte, i int, raw []byte) { f.rec(m, i).SetRaw(img, raw) }

func (f rt900Fam) Match(m *genModel, img []byte) bool { return false }

func (f rt900Fam) Get(m *genModel, img []byte, i int) *Chan {
	r := f.rec(m, i)
	if r.Raw(img)[0] == 0xFF {
		return nil
	}
	c := &Chan{Rx: r.F("rxfreq").Int(img) * 10}
	if c.Rx == 0 {
		return nil
	}
	if bytes.Equal(r.F("txfreq").Raw(img), []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		c.TxOff, c.Tx = true, c.Rx
	} else {
		c.Tx = r.F("txfreq").Int(img) * 10
	}
	c.TxTone = uv5rToneDec(r.F("txtone").Int(img), false)
	c.RxTone = uv5rToneDec(r.F("rxtone").Int(img), true)
	c.Scan = r.F("scan").Bool(img)
	pw := int(r.F("txpower").Int(img)) // 0 alta, 1 media, 2 bassa (3: non valido)
	if pw > 2 {
		pw = 0
	}
	c.Power = m.powerSem(pw)
	c.Narrow = r.F("narrow").Bool(img)
	// come CHIRP: 0xFF → spazio, spazi finali tolti, poi via i NUL
	c.Name = strings.ReplaceAll(pyName(r.F("name").Raw(img), true), "\x00", "")
	return c
}

func (f rt900Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	o := rt900o(m)
	r := f.rec(m, i)
	if c == nil {
		r.Fill(img, 0xFF)
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
	// impostazioni del canale che CHIRP conserva (mem.extra)
	wasEmpty := r.Raw(img)[0] == 0xFF
	var encrypt, scramble, bcl, learning, pttid, scode, code int64
	oldAM, oldRx := false, int64(-1)
	code = 0xFFFFFF
	if !wasEmpty {
		encrypt, scramble = r.F("encrypt").Int(img), r.F("scramble").Int(img)
		bcl, learning = r.F("bcl").Int(img), r.F("learning").Int(img)
		pttid, scode = r.F("pttid").Int(img), r.F("scode").Int(img)
		if v := r.F("code").Int(img); v != 0xFFFFFF {
			code = v & 0x7FFFFF
		}
		oldAM, oldRx = r.F("am_modulation").Bool(img), r.F("rxfreq").Int(img)*10
	}
	raw := append(bytes.Repeat([]byte{0}, 16), bytes.Repeat([]byte{0xFF}, 16)...)
	r.SetRaw(img, raw)
	r.F("rxfreq").SetInt(img, c.Rx/10)
	if c.TxOff {
		r.F("txfreq").Fill(img, 0xFF)
	} else {
		r.F("txfreq").SetInt(img, c.Tx/10)
	}
	r.F("name").SetStr(img, c.Name, 0xFF)
	r.F("txtone").SetInt(img, uv5rToneEnc(tt))
	r.F("rxtone").SetInt(img, uv5rToneEnc(rt))
	r.F("scan").SetBool(img, c.Scan)
	r.F("narrow").SetBool(img, c.Narrow)
	r.F("txpower").SetInt(img, int64(m.powerIndex(c.Power)))
	// AM: obbligatoria in banda aerea; altrimenti si conserva quella del
	// canale se la frequenza non cambia (es. canali CB in AM)
	am := (c.Rx >= o.airband.Lo && c.Rx <= o.airband.Hi) || (oldAM && oldRx == c.Rx)
	r.F("am_modulation").SetBool(img, am)
	if o.txen {
		r.F("txen").SetInt(img, 1)
	}
	if !wasEmpty {
		r.F("encrypt").SetInt(img, encrypt)
		r.F("scramble").SetInt(img, scramble)
		r.F("bcl").SetInt(img, bcl)
		r.F("learning").SetInt(img, learning)
		r.F("pttid").SetInt(img, pttid)
		r.F("scode").SetInt(img, scode)
		r.F("code").SetInt(img, code)
		if code == 0xFFFFFF {
			r.F("code_flag").SetInt(img, 0xFF)
		} else {
			r.F("code_flag").SetInt(img, 0xA0)
		}
	}
	return nil
}

// ---------------------------------------------------------------- protocollo

type rt900Sess struct {
	ctx context.Context
	p   Port
	m   *genModel
}

const rt900Timeout = 1500 * time.Millisecond

func (s *rt900Sess) recv(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, rt900Timeout); err != nil {
		if err == errTimeout {
			return nil, rErr("la radio non risponde")
		}
		return nil, err
	}
	return b, nil
}

func bytesIn(l [][]byte, b []byte) bool {
	for _, x := range l {
		if bytes.Equal(x, b) {
			return true
		}
	}
	return false
}

// ingresso in programmazione (_enter_programming_mode di radtel_rt900/mml_jc8810)
func (s *rt900Sess) enter() error {
	o := rt900o(s.m)
	ok, saw := false, false
	for k := 0; k < 5 && !ok; k++ {
		if err := s.p.Write(o.magic); err != nil {
			return err
		}
		b, err := s.recv(1)
		if err == errCancelled {
			return err
		}
		if err == nil {
			saw = true
			ok = b[0] == 0x06
		}
	}
	if !ok {
		if saw {
			return rErr("la radio non ha accettato la modalità di programmazione dopo 5 tentativi: controlla il cavo e riaccendi la radio")
		}
		return rErr("la radio non risponde")
	}
	if err := s.p.Write([]byte{'F'}); err != nil {
		return err
	}
	id, err := s.recv(8)
	if err != nil {
		return err
	}
	if o.fp1 != nil && !bytesIn(o.fp1, id) {
		return rErr("identificativo della radio sconosciuto (% X): è il modello giusto?", id)
	}
	id2, err := s.recv(8)
	if err != nil {
		return err
	}
	if o.model != "" {
		// modello (programma del costruttore: 'M' → 12 byte)
		if err := s.p.Write([]byte{'M'}); err != nil {
			return err
		}
		mb, err := s.recv(12)
		if err != nil {
			return err
		}
		if got := strings.TrimSpace(strings.Trim(string(mb), "\x00\xff")); got != o.model {
			return rErr("la radio si è identificata come «%s», non come %s: è il modello giusto?", got, o.model)
		}
	}
	if err := s.p.Write(o.crypt); err != nil {
		return err
	}
	ack, err := s.recv(1)
	if err != nil {
		return err
	}
	if ack[0] != 0x06 {
		return rErr("errore nell'impostazione della cifratura")
	}
	if o.fp2 != nil && !bytesIn(o.fp2, id2) {
		return rErr("identificativo secondario della radio sconosciuto (% X)", id2)
	}
	return nil
}

func (s *rt900Sess) exit() { _ = s.p.Write([]byte{'E'}) }

func rt900Open(ctx context.Context, m *genModel, port string) (*rt900Sess, error) {
	p, err := openPortFunc(port, m.Baud)
	if err != nil {
		return nil, err
	}
	return &rt900Sess{ctx: ctx, p: p, m: m}, nil
}

func (f rt900Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	o := rt900o(m)
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return nil, "", err
	}
	defer s.p.Close()
	prog(0, o.memsize, "Collegamento con la radio…")
	err = s.enter()
	if err != nil {
		s.exit()
		return nil, "", genFail("Identificazione", err)
	}
	key := []byte(btSymbols[1])
	bs := o.blk()
	reads := o.reads
	if reads == nil {
		reads = [][2]int{{0, (o.memsize + bs - 1) / bs * bs}}
	}
	total := 0
	for _, r := range reads {
		total += r[1] - r[0]
	}
	data := bytes.Repeat([]byte{0xFF}, o.memsize)
	if o.reads == nil {
		data = nil
	}
	done := 0
	for _, r := range reads {
		for a := r[0]; a < r[1]; a += bs {
			if ctx.Err() != nil {
				s.exit()
				return nil, "", errCancelled
			}
			prog(done, total, "Lettura…")
			cmd := []byte{'R', byte(a >> 8), byte(a), byte(bs)}
			if err := s.p.Write(cmd); err != nil {
				return nil, "", err
			}
			d, err := s.recv(4 + bs)
			if err == nil && !bytes.Equal(d[:4], cmd) {
				err = rErr("risposta inattesa")
			}
			if err != nil {
				s.exit()
				return nil, "", genFail(fmt.Sprintf("Lettura (0x%04X)", a), err)
			}
			blk := append([]byte{}, d[4:]...)
			if a < 0xF000 {
				btCrypt(blk, key)
			}
			if o.reads == nil {
				data = append(data, blk...)
			} else {
				copy(data[a:], blk)
			}
			done += bs
		}
	}
	s.exit()
	prog(total, total, "Lettura completata")
	return data, "", nil
}

func (f rt900Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	o := rt900o(m)
	if len(img) < o.memsize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), o.memsize)
	}
	total := 0
	for _, r := range o.ranges {
		total += r[1] - r[0]
	}
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer s.p.Close()
	prog(0, total, "Collegamento con la radio…")
	if err := s.enter(); err != nil {
		s.exit()
		return "", genFail("Identificazione", err)
	}
	key := []byte(btSymbols[1])
	bs := o.blk()
	done := 0
	for _, r := range o.ranges {
		for a := r[0]; a < r[1]; a += bs {
			prog(done, total, "Scrittura…")
			blk := append([]byte{}, img[a:a+bs]...)
			if a < 0xF000 {
				btCrypt(blk, key)
			}
			if err := s.p.Write(append([]byte{'W', byte(a >> 8), byte(a), byte(bs)}, blk...)); err != nil {
				return "", err
			}
			ack, err := s.recv(1)
			if err == nil && ack[0] != 0x06 {
				err = rErr("la radio ha rifiutato il blocco")
			}
			if err != nil {
				s.exit()
				return "", genFail(fmt.Sprintf("Scrittura (0x%04X)", a), err)
			}
			done += bs
		}
	}
	s.exit()
	prog(total, total, "Scrittura completata")
	return "", nil
}
