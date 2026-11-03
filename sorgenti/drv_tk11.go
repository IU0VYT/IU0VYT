// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver: tk11.py.
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.2: Quansheng TK11 (driver tk11.py di CHIRP). 999 canali di 64 byte da
// 0x0000 (frequenze in decine di Hz, spostamento separato, nome di 16 caratteri),
// tabella dei canali in uso da 0x11000 (0xFF = vuoto, altrimenti la gamma di
// frequenza); memoria di 0x23000 byte, 38400 baud.
//
// Protocollo (do_download, do_upload, encapsulate_message di tk11.py): pacchetti
// AB CD lunghezza(2, LE) dati CRC16-XModem(2, LE) DC BA, con dati e CRC cifrati
// con la stessa chiave XOR di 16 byte della UV-K5:
//   collegamento F4 01 04 00 + codice di sessione(4) → 64 byte (tipo 0x01F5 e
//   versione del firmware);
//   lettura FB 01 0C 00 indirizzo(4, LE, 0x080000 + posizione) 00 02 00 00 +
//   codice → 0x214 byte (tipo 0x01FC, 12 byte di intestazione e 512 di dati);
//   scrittura FD 01 0C 00 indirizzo lunghezza 00 00 codice + 512 byte → 16 byte
//   (tipo 0x01FE); dopo la scrittura il riavvio (DD 05 00 00).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const tk11Format = `
struct {
    ul32 rx_freq;
    ul32 freq_diff;
    u8 tx_non_standard_1;
    u8 tx_non_standard_2;
    u8 rx_qt_type;
    u8 tx_qt_type;
    u8 freq_dir;
    u8 band;
    u8 step;
    u8 encrypt;
    u8 power;
    u8 busy;
    u8 reverse;
    u8 dtmf_decode_flag;
    u8 ptt_id;
    u8 mode;
    u8 scan_list;
    u8 sq;
    char name[16];
    ul32 rx_qt;
    ul32 tx_qt;
    u8 unknown[8];
    ul32 tx_qt2;
    u8 signal;
    u8 unknown_2[3];
} channels[999];
#seekto 0x11000;
struct {
    u8 flag;
} channels_usage[1011];
#seekto 0x22280;
struct {
    u8 value;
} noaa_same_events_control[128];
`

const (
	tk11MemSize = 0x23000
	tk11Chunk   = 0x200
	tk11Magic   = 5 // codice di sessione (come CHIRP: un valore fisso)
	tk11Wait    = 2 * time.Second
)

// gamme di frequenza della tabella dei canali in uso (TX_FREQUENCIES)
var tk11Ranges = [][2]int64{{150000, 1800000}, {1800000, 18000000}, {18000000, 32000000}, {32000000, 76000000},
	{108000000, 136000000}, {136000000, 174000000}, {174000000, 350000000}, {350000000, 400000000},
	{400000000, 470000000}, {470000000, 580000000}, {580000000, 760000000}, {760000000, 1000000000},
	{1000000000, 1160000000}}

type tk11Fam struct{}

func init() {
	genRegister(&genModel{ID: "qs-tk11", Vendor: "Quansheng", Model: "TK11", RClass: "TK11", Name: "Quansheng TK11",
		Slots: 999, First: 1, NameMax: 16,
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 5 W", 5, 1}, {"Alta 10 W", 10, 0}},
		Bands:  []band{{150000, 1160000000}}, Narrow: true, NoTxOff: true, DCS: dtcsCodes, Baud: 38400, Fam: tk11Fam{},
		Notes: "Accendi la radio con il cavo collegato. 999 canali con nome di 16 caratteri (driver di CHIRP della Quansheng TK11). " +
			"I canali in AM, SSB e CW restano nel loro modo. La radio non permette di disabilitare la trasmissione sul canale. " +
			"Dopo la scrittura la radio si riavvia."})
}

func (tk11Fam) mem() *bwEl                 { return bwMust(tk11Format) }
func (f tk11Fam) layout(m *genModel) *bwEl { return f.mem() }
func (f tk11Fam) crossband() bool          { return true }

func (f tk11Fam) recs(i int) (*bwEl, *bwEl) {
	b := f.mem()
	return b.F("channels").I(i), b.F("channels_usage").I(i)
}

func (f tk11Fam) Rec(m *genModel, img []byte, i int) []byte {
	r, u := f.recs(i)
	return append(append([]byte{}, r.Raw(img)...), u.Raw(img)...)
}

func (f tk11Fam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	r, u := f.recs(i)
	r.SetRaw(img, raw[:64])
	u.SetRaw(img, raw[64:65])
}

func (f tk11Fam) Match(m *genModel, img []byte) bool { return false }

func tk11Range(hz int64) int64 {
	if hz == 0 {
		return 0xFF
	}
	n := int64(0)
	for _, r := range tk11Ranges {
		if r[0] > hz || hz > r[1] {
			n++
		} else {
			break
		}
	}
	return n
}

// tono: 1 CTCSS (se non è un tono della lista: 63,0 Hz, come CHIRP), 2 DCS
// normale, 3 DCS invertito (codice in ottale), altro nessuno
func tk11ToneDec(typ, qt int64) string {
	switch typ {
	case 1:
		if qt != 630 && idxInt(chirpTones, int(qt)) < 0 {
			qt = 630
		}
		return fmtTone(int(qt))
	case 2, 3:
		code, _ := strconv.Atoi(strconv.FormatInt(qt, 8))
		return genTone{Mode: "DTCS", Code: code, Inv: typ == 3}.String()
	}
	return ""
}

func tk11ToneEnc(t genTone) (typ, qt int64, err error) {
	switch t.Mode {
	case "Tone":
		return 1, int64(t.Hz10), nil
	case "DTCS":
		v, err := strconv.ParseInt(strconv.Itoa(t.Code), 8, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("codice DCS %03d non valido", t.Code)
		}
		if t.Inv {
			return 3, v, nil
		}
		return 2, v, nil
	}
	return 0, 0, nil
}

func (f tk11Fam) Get(m *genModel, img []byte, i int) *Chan {
	r, u := f.recs(i)
	rx := r.F("rx_freq").Int(img)
	if rx == 0 || u.F("flag").Int(img) == 0xFF {
		return nil
	}
	c := &Chan{Rx: rx * 10, Tx: rx * 10}
	if off := r.F("freq_diff").Int(img) * 10; off != 0 {
		switch r.F("freq_dir").Int(img) {
		case 1:
			c.Tx = c.Rx + off
		case 2:
			c.Tx = c.Rx - off
		}
	}
	pw := int(r.F("power").Int(img))
	if pw > 2 {
		pw = 2 // valore sconosciuto: alta, come CHIRP
	}
	c.Power = m.powerSem(pw)
	mode := r.F("mode").Int(img)
	c.Narrow = mode == 5 || (mode == 0 && r.F("band").Int(img)&15 == 1)
	// nome: vuoto se inizia con 0xFF o non è ASCII; spazi e 0x00 finali tolti
	nb := r.F("name").Raw(img)
	if nb[0] != 0xFF {
		nb = trimBytes(nb, ' ', 0x00)
		if bytes.IndexFunc(nb, func(r rune) bool { return r >= 0x80 || r == 0xFFFD }) < 0 {
			c.Name = string(nb)
		}
	}
	c.TxTone = tk11ToneDec(r.F("tx_qt_type").Int(img), r.F("tx_qt").Int(img))
	c.RxTone = tk11ToneDec(r.F("rx_qt_type").Int(img), r.F("rx_qt").Int(img))
	c.Scan = true
	return c
}

func (f tk11Fam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r, u := f.recs(i)
	old := f.Get(m, img, i)
	if c == nil {
		// come CHIRP (erase_memory): canale libero nella tabella, record a zero
		u.F("flag").SetInt(img, 0xFF)
		r.Fill(img, 0)
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
	txT, txQ, err := tk11ToneEnc(tt)
	if err != nil {
		return err
	}
	rxT, rxQ, err := tk11ToneEnc(rt)
	if err != nil {
		return err
	}
	for _, ch := range c.Name {
		if ch >= 0x80 {
			return fmt.Errorf("nome %q: la radio accetta solo lettere senza accenti", c.Name)
		}
	}
	u.F("flag").SetInt(img, tk11Range(c.Rx))
	r.F("rx_freq").SetInt(img, c.Rx/10)
	// impostazioni del canale lette e riscritte da CHIRP (mem.extra); canale
	// nuovo: i valori predefiniti di CHIRP
	msw, scan, enc, sq, busy, dtmf, sig, ptt := int64(0), int64(0xFF), int64(0), int64(4), int64(0), int64(0), int64(0), int64(0)
	mode := int64(0)
	if old != nil {
		msw = r.F("band").Int(img) >> 4
		scan, enc, sq = r.F("scan_list").Int(img), r.F("encrypt").Int(img), r.F("sq").Int(img)
		busy, dtmf = b2i(r.F("busy").Int(img) != 0), b2i(r.F("dtmf_decode_flag").Int(img) != 0)
		sig, ptt = r.F("signal").Int(img), r.F("ptt_id").Int(img)
		// AM, LSB, USB, CW: il programma non ha questi modi nei canali, restano
		if md := r.F("mode").Int(img); md >= 1 && md <= 4 && !c.Narrow {
			mode = md
		}
	}
	r.F("mode").SetInt(img, mode)
	r.F("band").SetInt(img, b2i(c.Narrow)|msw<<4)
	r.F("power").SetInt(img, int64(m.powerIndex(c.Power)))
	nb := append([]byte(c.Name), make([]byte, 16)...)[:16]
	r.F("name").SetRaw(img, nb)
	switch {
	case c.Tx == c.Rx:
		r.F("freq_dir").SetInt(img, 0)
		r.F("freq_diff").SetInt(img, 0)
	case c.Tx > c.Rx:
		r.F("freq_dir").SetInt(img, 1)
		r.F("freq_diff").SetInt(img, (c.Tx-c.Rx)/10)
	default:
		r.F("freq_dir").SetInt(img, 2)
		r.F("freq_diff").SetInt(img, (c.Rx-c.Tx)/10)
	}
	r.F("scan_list").SetInt(img, scan)
	r.F("encrypt").SetInt(img, enc)
	r.F("sq").SetInt(img, sq)
	r.F("busy").SetInt(img, busy)
	r.F("dtmf_decode_flag").SetInt(img, dtmf)
	r.F("signal").SetInt(img, sig)
	r.F("ptt_id").SetInt(img, ptt)
	r.F("tx_qt_type").SetInt(img, txT)
	r.F("tx_qt").SetInt(img, txQ)
	r.F("rx_qt_type").SetInt(img, rxT)
	r.F("rx_qt").SetInt(img, rxQ)
	return nil
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// chirpInvalid (solo per i test): CHIRP non ammette un tono in sola ricezione
// o un DCS in sola trasmissione su questa radio (VALID_CROSS_MODES) e non li
// scrive
func (f tk11Fam) chirpInvalid(m *genModel, c *Chan) bool {
	tx, _ := genToneParse(c.TxTone)
	rx, _ := genToneParse(c.RxTone)
	return (tx.Mode == "" && rx.Mode != "") || (tx.Mode == "DTCS" && rx.Mode == "")
}

// ---------------------------------------------------------------- protocollo

func tk11Packet(data []byte) []byte {
	crc := make([]byte, 2)
	binary.LittleEndian.PutUint16(crc, crc16XModem(data))
	out := []byte{0xAB, 0xCD, byte(len(data)), byte(len(data) >> 8)}
	out = append(out, uvk5Xor(append(append([]byte{}, data...), crc...))...)
	return append(out, 0xDC, 0xBA)
}

type tk11Sess struct {
	ctx context.Context
	p   Port
}

// reply: risposta di n byte in tutto (read_serial di tk11.py), dati decifrati
func (s *tk11Sess) reply(n int) ([]byte, error) {
	b := make([]byte, n)
	if err := s.p.ReadFull(s.ctx, b, tk11Wait); err != nil {
		return nil, err
	}
	if b[0] != 0xAB || b[1] != 0xCD || b[n-2] != 0xDC || b[n-1] != 0xBA {
		return nil, rErr("risposta della radio non valida")
	}
	l := int(binary.LittleEndian.Uint16(b[2:]))
	if 4+l > n {
		return nil, rErr("risposta della radio non valida")
	}
	return uvk5Xor(b[4 : 4+l]), nil
}

func tk11Req(typ uint16, addr, n int) []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint16(b[0:], typ)
	binary.LittleEndian.PutUint16(b[2:], 0x0C)
	binary.LittleEndian.PutUint32(b[4:], uint32(0x080000+addr))
	binary.LittleEndian.PutUint16(b[8:], uint16(n))
	binary.LittleEndian.PutUint32(b[12:], tk11Magic)
	return b
}

func (s *tk11Sess) connect() (string, error) {
	req := make([]byte, 8)
	binary.LittleEndian.PutUint16(req[0:], 0x01F4)
	binary.LittleEndian.PutUint16(req[2:], 4)
	binary.LittleEndian.PutUint32(req[4:], tk11Magic)
	if err := s.p.Write(tk11Packet(req)); err != nil {
		return "", err
	}
	rep, err := s.reply(64)
	if err != nil {
		return "", err
	}
	if len(rep) < 20 || binary.LittleEndian.Uint16(rep) != 0x01F5 {
		return "", rErr("risposta al collegamento non valida: è una Quansheng TK11?")
	}
	ver := rep[4:20]
	if k := bytes.IndexByte(ver, 0); k >= 0 {
		ver = ver[:k]
	}
	return strings.TrimSpace(string(ver)), nil
}

func (f tk11Fam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return mem, note, err
}

func (f tk11Fam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	p, err := openPortFunc(port, m.Baud)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	s := &tk11Sess{ctx: ctx, p: p}
	prog(0, tk11MemSize, "Collegamento con la radio…")
	fw, err := s.connect()
	if err != nil {
		return nil, nil, "", genFail("Collegamento", err)
	}
	data := make([]byte, 0, tk11MemSize)
	for a := 0; a < tk11MemSize; a += tk11Chunk {
		if ctx.Err() != nil {
			return nil, nil, "", errCancelled
		}
		prog(a, tk11MemSize, "Lettura…")
		if err := p.Write(tk11Packet(tk11Req(0x01FB, a, tk11Chunk))); err != nil {
			return nil, nil, "", err
		}
		rep, err := s.reply(tk11Chunk + 20)
		if err == nil && (len(rep) < 12+tk11Chunk || binary.LittleEndian.Uint16(rep) != 0x01FC) {
			err = rErr("risposta inattesa")
		}
		if err != nil {
			return nil, nil, "", genFail(fmt.Sprintf("Lettura (0x%05X)", a), err)
		}
		data = append(data, rep[12:12+tk11Chunk]...)
	}
	prog(tk11MemSize, tk11MemSize, "Lettura completata")
	return data, map[string]interface{}{"tk11_firmware_version": fw}, "Firmware della radio: " + fw, nil
}

func (f tk11Fam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < tk11MemSize {
		return "", fmt.Errorf("immagine della memoria incompleta (%d byte invece di %d)", len(img), tk11MemSize)
	}
	img = img[:tk11MemSize]
	p, err := openPortFunc(port, m.Baud)
	if err != nil {
		return "", err
	}
	defer p.Close()
	s := &tk11Sess{ctx: ctx, p: p}
	prog(0, tk11MemSize, "Collegamento con la radio…")
	for a := 0; a < tk11MemSize; a += tk11Chunk {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(a, tk11MemSize, "Scrittura…")
		pk := append(tk11Req(0x01FD, a, tk11Chunk), img[a:a+tk11Chunk]...)
		if err := p.Write(tk11Packet(pk)); err != nil {
			return "", err
		}
		rep, err := s.reply(16)
		if err == nil && (len(rep) < 2 || binary.LittleEndian.Uint16(rep) != 0x01FE) {
			err = rErr("la radio non ha confermato la scrittura")
		}
		if err != nil {
			return "", genFail(fmt.Sprintf("Scrittura (0x%05X)", a), err)
		}
	}
	// riavvio della radio
	req := make([]byte, 4)
	binary.LittleEndian.PutUint16(req, 0x05DD)
	_ = p.Write(tk11Packet(req))
	prog(tk11MemSize, tk11MemSize, "Scrittura completata")
	return "", nil
}
