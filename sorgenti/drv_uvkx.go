// IU0VYT Flash - gestione canali, ponti e firmware di radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
//
// Questo programma è software libero: puoi ridistribuirlo e/o modificarlo secondo
// i termini della GNU General Public License pubblicata dalla Free Software
// Foundation, versione 3 della licenza o (a tua scelta) qualsiasi versione
// successiva. È distribuito SENZA ALCUNA GARANZIA: vedi il file LICENSE.

// 4.16: Quansheng UV-K5 / UV-K6 con il firmware libero «UV-KX» di joaquim.org
// (https://github.com/joaquimorg/UV-KX). Memoria, regole dei canali e comandi
// sono ricavati dal driver per CHIRP pubblicato con il firmware
// (chirp/CHIRP_uvkx.py, Copyright 2024 joaquim.org, licenza Apache 2.0, a sua
// volta derivato da uvk5.py di Jacek Lipkowski SQ5BPF) e riscritti in Go.
//
// EEPROM di 8 KiB: impostazioni e VFO a 0x0000-0x004F, 230 canali da 32 byte da
// 0x0050, taratura da 0x1E00. Differenza voluta dal driver per CHIRP: lui scrive
// tutta la memoria, taratura compresa; qui si scrive solo fino a 0x1DFF.

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"
)

type uvkxFam struct{}

const uvkxFormat = `
#seekto 0x0000;
struct {
    ul16 version;
    u8 battery_type:2,
       busy_lockout:1,
       beep:1,
       backlight_level:4;
    u8 backlight_time:4,
       mic_gain_db:4;
    u8 lcd_contrast:4,
       tx_timeout:4;
    u8 battery_save:4,
       backlight_mode:2,
       vfo_selected:2;
    ul16 memory[2];
    struct {
        ul32 rx_frequency;
        u8   rx_code_type;
        u8   rx_code;
        ul32 tx_frequency;
        u8   tx_code_type;
        u8   tx_code;
        char name[10];
        ul16 channel_id;
        u8 squelch:4,
           step:4;
        u8 modulation:4,
           bandwidth:4;
        u8 power:2,
           shift:2,
           repeater_ste:1,
           ste:1,
           compander:2;
        u8 roger:4,
           pttid:4;
        u8 rxagc:6,
           vfo_reserved_bits:2;
        u8 reserved_bytes[3];
    } vfo[2];
    u8 show_vfo[2];
    u8 settings_reserved[4];
} settings;

#seekto 0x0050;
struct {
    ul32 rx_frequency;
    u8   rx_code_type;
    u8   rx_code;
    ul32 tx_frequency;
    u8   tx_code_type;
    u8   tx_code;
    char name[10];
    ul16 channel_id;
    u8 squelch:4,
       step:4;
    u8 modulation:4,
       bandwidth:4;
    u8 power:2,
       shift:2,
       repeater_ste:1,
       ste:1,
       compander:2;
    u8 roger:4,
       pttid:4;
    u8 rxagc:6,
       channel_reserved_bits:2;
    u8 reserved_bytes[3];
} channel[230];
`

const (
	uvkxChans   = 230
	uvkxChanAt  = 0x0050
	uvkxRec     = 32
	uvkxCal     = 0x1E00 // taratura: mai scritta
	uvkxMinHz   = 18000000
	uvkxMaxHz   = 1300000000
	uvkxStepNew = 3 // 5 kHz nella tabella dei passi del driver
)

func init() {
	genRegister(&genModel{ID: "qs-uvk5-uvkx", Vendor: "Quansheng", Model: "UV-Kx", Variant: "joaquim.org", RClass: "UVKxRadio",
		Name:  "Quansheng UV-K5 / UV-K6 (firmware UV-KX di joaquim.org)",
		Slots: uvkxChans, First: 1, NameMax: 10, Charset: charsetASCII,
		Powers: []genPower{{"Bassa 1 W", 1, 2}, {"Media 2,5 W", 2.5, 1}, {"Alta 5 W", 5, 0}},
		Bands:  []band{{uvkxMinHz, uvkxMaxHz}}, Narrow: false, HasScan: false,
		DCS: dtcsCodes, Baud: 115200, Fam: uvkxFam{},
		Opt: &uvk5Opt{approve: func(fw string) bool { return strings.HasPrefix(strings.ToUpper(fw), "JOAQUIM") }, emuFW: "JOAQUIM V0.0.1",
			bauds: []int{115200}},
		Notes: "Firmware libero UV-KX di joaquim.org (la radio risponde «JOAQUIM …»): 230 canali con nomi di 10 caratteri, 115200 baud. " +
			"Accendi la radio, poi collega il cavo al connettore microfono/altoparlante. Larghezza di banda, squelch e guadagno di " +
			"ogni canale restano quelli impostati sulla radio; la modulazione resta quella del canale se la frequenza non cambia " +
			"(AM nella banda aerea per un canale nuovo). La taratura non si scrive mai (il driver per CHIRP la scrive). " +
			"Dal driver per CHIRP pubblicato con il firmware."})
}

func (uvkxFam) mem() *bwEl                 { return bwMust(uvkxFormat) }
func (f uvkxFam) layout(m *genModel) *bwEl { return f.mem() }
func (uvkxFam) at(i int) int               { return uvkxChanAt + i*uvkxRec }

func (f uvkxFam) Rec(m *genModel, img []byte, i int) []byte {
	return append([]byte{}, img[f.at(i):f.at(i)+uvkxRec]...)
}

func (f uvkxFam) PutRec(m *genModel, img []byte, i int, raw []byte) {
	copy(img[f.at(i):f.at(i)+uvkxRec], raw)
}

func (uvkxFam) Match(m *genModel, img []byte) bool { return false }

// subtoni della radio nell'ordine dei codici (CTCSS_TONES del driver: 50 toni)
var uvkxCtcss = []int{670, 693, 719, 744, 770, 797, 825, 854, 885, 915, 948, 974, 1000, 1035, 1072, 1109, 1148, 1188, 1230, 1273,
	1318, 1365, 1413, 1462, 1514, 1567, 1598, 1622, 1655, 1679, 1713, 1738, 1773, 1799, 1835, 1862, 1899, 1928, 1966, 1995,
	2035, 2065, 2107, 2181, 2257, 2291, 2336, 2418, 2503, 2541}

// tono: tipo (0 nessuno, 1 CTCSS, 2 DCS, 3 DCS invertito) e indice nella tabella
func uvkxToneDec(typ, code byte) string {
	switch typ {
	case 1:
		if int(code) < len(uvkxCtcss) {
			return fmtTone(uvkxCtcss[code])
		}
	case 2, 3:
		if int(code) < len(dtcsCodes) {
			return genTone{Mode: "DTCS", Code: dtcsCodes[code], Inv: typ == 3}.String()
		}
	}
	return ""
}

func uvkxToneEnc(s string) (typ, code byte, err error) {
	t, err := genToneParse(s)
	if err != nil {
		return 0, 0, err
	}
	switch t.Mode {
	case "Tone":
		for k, v := range uvkxCtcss {
			if v == t.Hz10 {
				return 1, byte(k), nil
			}
		}
		return 0, 0, fmt.Errorf("tono %s non disponibile su questa radio", t.String())
	case "DTCS":
		for k, v := range dtcsCodes {
			if v == t.Code {
				if t.Inv {
					return 3, byte(k), nil
				}
				return 2, byte(k), nil
			}
		}
		return 0, 0, fmt.Errorf("tono %s non disponibile su questa radio", t.String())
	}
	return 0, 0, nil
}

func (f uvkxFam) Get(m *genModel, img []byte, i int) *Chan {
	r := img[f.at(i) : f.at(i)+uvkxRec]
	le := binary.LittleEndian
	raw := le.Uint32(r[0:])
	// canale vuoto: frequenza 0 (come il driver); anche la EEPROM mai scritta (FF)
	if raw == 0 || raw == 0xFFFFFFFF {
		return nil
	}
	rx := int64(raw) * 10
	c := &Chan{Rx: rx, Tx: rx, Scan: true}
	tx := int64(le.Uint32(r[6:])) * 10
	shift := (r[26] >> 4) & 3
	switch {
	case tx == 0:
		c.TxOff = true
	case shift == 1:
		c.Tx = rx + abs64(tx-rx)
	case shift == 2:
		c.Tx = rx - abs64(rx-tx)
	default:
		c.Tx = tx
	}
	c.RxTone, c.TxTone = uvkxToneDec(r[4], r[5]), uvkxToneDec(r[10], r[11])
	switch p := r[26] >> 6; {
	case p == 0:
		c.Power = 2
	case p == 1:
		c.Power = 1
	default:
		c.Power = 0
	}
	var name []byte
	for _, b := range r[12:22] {
		if b == 0 || b == 0xFF {
			break
		}
		name = append(name, b)
	}
	c.Name = strings.TrimSpace(string(name))
	return c
}

func (f uvkxFam) Set(m *genModel, img []byte, i int, c *Chan) error {
	r := img[f.at(i) : f.at(i)+uvkxRec]
	le := binary.LittleEndian
	if c == nil {
		// _clear_channel: tutto a zero tranne i bit e i byte di riserva
		for k := 0; k < 28; k++ {
			r[k] = 0
		}
		r[28] &= 0x03
		return nil
	}
	if c.Rx%10 != 0 || c.Rx < uvkxMinHz || c.Rx > uvkxMaxHz {
		return fmt.Errorf("frequenza %s MHz fuori dai limiti della radio", fmtMHz(c.Rx))
	}
	if !c.TxOff && (c.Tx%10 != 0 || c.Tx < uvkxMinHz || c.Tx > uvkxMaxHz) {
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
	for _, ch := range c.Name {
		if ch < 0x20 || ch > 0x7E {
			return fmt.Errorf("carattere non ammesso nel nome")
		}
	}
	oldRaw := le.Uint32(r[0:])
	empty := oldRaw == 0 || oldRaw == 0xFFFFFFFF
	if oldRaw == 0xFFFFFFFF {
		// EEPROM mai scritta: si parte da un canale azzerato
		for k := range r {
			r[k] = 0
		}
	}
	oldMod := r[25] & 0x0F
	name := strings.TrimSpace(c.Name)
	if len(name) > 10 {
		name = name[:10]
	}
	copy(r[12:22], make([]byte, 10))
	copy(r[12:22], name)
	le.PutUint16(r[22:], uint16(i+1))
	le.PutUint32(r[0:], uint32(c.Rx/10))
	shift := byte(0)
	switch {
	case c.TxOff:
		le.PutUint32(r[6:], 0)
	case c.Tx == c.Rx:
		le.PutUint32(r[6:], uint32(c.Rx/10))
	case abs64(c.Tx-c.Rx) > 70000000:
		le.PutUint32(r[6:], uint32(c.Tx/10)) // frequenza di trasmissione separata
	case c.Tx > c.Rx:
		le.PutUint32(r[6:], uint32(c.Tx/10))
		shift = 1
	default:
		le.PutUint32(r[6:], uint32(c.Tx/10))
		shift = 2
	}
	// modulazione (nel driver è nel campo «bandwidth»): quella del canale se la
	// frequenza non cambia, AM nella banda aerea, altrimenti FM
	mod := byte(0)
	if !empty && int64(oldRaw)*10 == c.Rx {
		mod = oldMod
	} else if c.Rx >= 108000000 && c.Rx < 137000000 {
		mod = 1
	}
	if empty {
		// canale nuovo, come il driver: passo 5 kHz; l'indice della larghezza di
		// banda prende il valore che c'era nel mezzo byte basso (al massimo 9)
		r[24] = uvkxStepNew<<4 | r[24]&0x0F
		bw := oldMod
		if bw > 9 {
			bw = 9
		}
		r[25] = bw<<4 | mod
	} else {
		r[25] = r[25]&0xF0 | mod
	}
	p := byte(0)
	switch c.Power {
	case 0:
		p = 2
	case 1:
		p = 1
	}
	r[26] = p<<6 | shift<<4 | r[26]&0x0F
	r[4], r[5], r[10], r[11] = rt, rc, tt, tc
	return nil
}

func (f uvkxFam) Download(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, string, error) {
	mem, _, note, err := f.DownloadMeta(ctx, m, port, prog)
	return mem, note, err
}

func (f uvkxFam) DownloadMeta(ctx context.Context, m *genModel, port string, prog Progress) ([]byte, map[string]interface{}, string, error) {
	prog(0, uvk5MemSize, "Collegamento con la radio…")
	p, s, fw, err := uvk5xFam{}.connect(ctx, m, port)
	if err != nil {
		return nil, nil, "", err
	}
	defer p.Close()
	var data []byte
	for a := 0; a < uvk5MemSize; a += uvk5Block {
		if ctx.Err() != nil {
			return nil, nil, "", errCancelled
		}
		prog(a, uvk5MemSize, "Lettura…")
		b, err := s.readMem(a, uvk5Block)
		if err != nil {
			return nil, nil, "", err
		}
		data = append(data, b...)
	}
	prog(uvk5MemSize, uvk5MemSize, "Lettura completata")
	return data, map[string]interface{}{"uvk5_firmware": fw}, "Firmware della radio: " + fw, nil
}

func (f uvkxFam) Upload(ctx context.Context, m *genModel, port string, img []byte, prog Progress) (string, error) {
	if len(img) < uvk5MemSize {
		return "", fmt.Errorf("immagine della memoria incompleta")
	}
	prog(0, uvkxCal, "Collegamento con la radio…")
	p, s, fw, err := uvk5xFam{}.connect(ctx, m, port)
	if err != nil {
		return "", err
	}
	defer p.Close()
	// impostazioni e canali: la taratura (da 0x1E00) non si scrive mai
	for a := 0; a < uvkxCal; a += uvk5Block {
		if ctx.Err() != nil {
			return "", errCancelled
		}
		prog(a, uvkxCal, "Scrittura…")
		if err := s.writeMem(a, img[a:a+uvk5Block]); err != nil {
			return "", err
		}
	}
	_ = s.cmd([]byte{0xdd, 0x05, 0x00, 0x00})
	prog(uvkxCal, uvkxCal, "Scrittura completata")
	return "Firmware della radio: " + fw, nil
}
