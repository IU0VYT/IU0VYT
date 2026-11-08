// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Nomi delle zone (4.11): al posto di «Zona 1», «Zona 2»… si può dare un nome a
// ogni zona, come nei programmi del costruttore. I nomi sono in un'area fissa
// della memoria, un record per zona; ricavati, per l'interoperabilità, dal codice
// dei programmi del costruttore e dal driver di CHIRP:
//   - Radtel RT-950 Pro / Bajeton BJ-9000 / Hiroyasu HI-UV850 Pro (programmi
//     RT-950PRO_CPS v1.2.0-v1.3.3 e BJ-9000_CPS v1.3.3) e Radtel RT-950 (BT-9000_CPS
//     v1.3.0): 0xC000, 16 byte per zona, fino a 12 byte GB2312, il resto 0xFF
//     (RWDataOperation.GetZoneNameInfo_Parameter / SetZoneNameInfo_Parameter);
//   - Radtel RT-920 (RT-920_EnCPS v1.2.4 e radtel_rt900.py di CHIRP, firmware
//     V0.17 e successivi): 0xC800, 10 zone da 16 byte; si usano 10 caratteri,
//     come il driver di CHIRP (il programma ne permette 12);
//   - Radtel RT-910 / Bajeton BJ-7700 senza Bluetooth (RT910_EnCPS v1.0.3):
//     0xA200, 16 byte per zona, fino a 12 caratteri ASCII, il resto 0x00.
// Un nome vuoto lascia la zona senza nome (la radio mostra il suo nome predefinito).
// Le radio DMR con le zone come elenchi di canali (Baofeng RD-5R, DM-1801,
// DM-32UV) usano invece l'interfaccia zoneNamer della famiglia: si rinominano le
// zone che ci sono, con nomi fino a 16 caratteri.

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

type zoneNameSpec struct {
	Addr   int  // inizio dell'area
	Count  int  // zone con nome
	Stride int  // byte per zona
	Max    int  // byte del nome
	GB     bool // GB2312 (altrimenti ASCII)
	Pad    byte // riempimento dopo il nome
	// Area: byte scritti nella radio (multiplo del blocco); 0 = l'area è già tra
	// quelle scritte sempre (RT-920, come CHIRP)
	Area int
}

// zoneNamer: famiglie con zone di dimensione libera (radio DMR) che hanno un nome
type zoneNamer interface {
	ZoneNames(m *genModel, img []byte) ([]string, int, bool)
	SetZoneNames(m *genModel, img []byte, names []string) error
}

// zoneNamesOf: nomi letti dall'immagine; ok = false se l'immagine non contiene
// l'area (file salvato prima della 4.11) o se l'area non contiene nomi validi
func zoneNamesOf(m *genModel, img []byte) (names []string, max int, ok bool) {
	if zn, is := m.Fam.(zoneNamer); is {
		return zn.ZoneNames(m, img)
	}
	s := m.ZoneNames
	if s == nil || len(img) < s.Addr+s.Count*s.Stride {
		return nil, 0, false
	}
	for z := 0; z < s.Count; z++ {
		n, good := s.decode(img[s.Addr+z*s.Stride : s.Addr+(z+1)*s.Stride])
		if !good {
			return nil, 0, false
		}
		names = append(names, n)
	}
	return names, s.Max, true
}

// decode: nome di un record (fino al primo 0x00 o 0xFF); good = false se il record
// non è un nome (caratteri non validi, o dati dopo il riempimento)
func (s *zoneNameSpec) decode(rec []byte) (string, bool) {
	end := len(rec)
	for i, b := range rec {
		if b == 0x00 || b == 0xFF {
			end = i
			break
		}
	}
	if end > s.Max {
		return "", false
	}
	for _, b := range rec[end:] {
		if b != 0x00 && b != 0xFF {
			return "", false
		}
	}
	raw := rec[:end]
	for _, b := range raw {
		if b < 0x20 || b == 0x7F || (!s.GB && b > 0x7E) {
			return "", false
		}
	}
	if s.GB {
		str, ok := gb2312Dec(raw)
		return strings.TrimSpace(str), ok
	}
	return strings.TrimSpace(string(raw)), true
}

// zoneRuneOK: carattere ammesso in un nome GB2312: ASCII stampabile, caratteri
// cinesi e punteggiatura cinese (come i programmi del costruttore)
func zoneRuneOK(r rune) bool {
	return (r >= 0x20 && r <= 0x7E) || unicode.Is(unicode.Han, r) || (r >= 0x3000 && r <= 0x303F) || (r >= 0xFF01 && r <= 0xFF5E)
}

// encode: record del nome (errore se troppo lungo o con caratteri non ammessi)
func (s *zoneNameSpec) encode(name string) ([]byte, error) {
	name = strings.TrimSpace(name)
	var raw []byte
	if s.GB {
		bad := false
		for _, r := range name {
			bad = bad || !zoneRuneOK(r)
		}
		raw = gb2312Enc(name)
		if dec, _ := gb2312Dec(raw); bad || dec != name {
			return nil, fmt.Errorf("«%s»: la radio accetta solo lettere senza accenti, cifre, simboli e caratteri cinesi", name)
		}
	} else {
		for _, r := range name {
			if r < 0x20 || r > 0x7E {
				return nil, fmt.Errorf("«%s»: la radio accetta solo lettere senza accenti, cifre e simboli", name)
			}
		}
		raw = []byte(name)
	}
	if len(raw) > s.Max {
		if s.GB && len(raw) != utf8.RuneCountInString(name) {
			return nil, fmt.Errorf("«%s»: troppo lungo, al massimo %d byte (ogni carattere cinese ne occupa 2)", name, s.Max)
		}
		return nil, fmt.Errorf("«%s»: troppo lungo, al massimo %d caratteri", name, s.Max)
	}
	rec := make([]byte, s.Stride)
	for i := range rec {
		rec[i] = s.Pad
	}
	copy(rec, raw)
	return rec, nil
}

// setZoneNames: scrive i nomi nell'immagine (solo se tutti sono validi)
func setZoneNames(m *genModel, img []byte, names []string) error {
	if zn, is := m.Fam.(zoneNamer); is {
		return zn.SetZoneNames(m, img, names)
	}
	s := m.ZoneNames
	if s == nil {
		return errors.New("questa radio non ha nomi delle zone")
	}
	if len(img) < s.Addr+s.Count*s.Stride {
		return errors.New("il file aperto non contiene i nomi delle zone: leggi di nuovo la radio via cavo")
	}
	if len(names) != s.Count {
		return fmt.Errorf("servono %d nomi, ricevuti %d", s.Count, len(names))
	}
	var recs [][]byte
	for _, n := range names {
		rec, err := s.encode(n)
		if err != nil {
			return err
		}
		recs = append(recs, rec)
	}
	for z, rec := range recs {
		a := s.Addr + z*s.Stride
		// un nome invariato resta byte per byte com'è (riempimento compreso)
		if old, ok := s.decode(img[a : a+s.Stride]); ok && old == strings.TrimSpace(names[z]) {
			continue
		}
		copy(img[a:], rec)
	}
	return nil
}

// zoneAreaPlausible: l'area dei nomi letta dalla radio contiene solo nomi validi o
// posizioni vuote (altrimenti la radio usa quei byte per altro: non si scrive)
func zoneAreaPlausible(s *zoneNameSpec, area []byte) bool {
	for z := 0; z < s.Count && (z+1)*s.Stride <= len(area); z++ {
		if _, ok := s.decode(area[z*s.Stride : (z+1)*s.Stride]); !ok {
			return false
		}
	}
	return true
}

// API: nomi delle zone della radio aperta
func (a *App) apiZoneNames(r *http.Request) (interface{}, error) {
	var req struct {
		Kind  string   `json:"kind"`
		Names []string `json:"names"`
		Set   bool     `json:"set"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	m, err := genModelOf(req.Kind)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sess := a.radios[req.Kind]
	if sess == nil {
		return map[string]interface{}{"available": false, "reason": "Per dare un nome alle zone leggi prima la radio via cavo (o apri un suo file)."}, nil
	}
	if req.Set {
		img := append([]byte{}, sess.Base...)
		if err := setZoneNames(m, img, req.Names); err != nil {
			return nil, err
		}
		sess.Base = img
	}
	names, max, ok := zoneNamesOf(m, sess.Base)
	if !ok {
		reason := "La memoria aperta non contiene i nomi delle zone (file salvato con una versione precedente o firmware senza nomi delle zone): leggi di nuovo la radio via cavo."
		return map[string]interface{}{"available": false, "reason": reason}, nil
	}
	_, required := m.Fam.(zoneNamer) // radio DMR: ogni zona ha un nome
	return map[string]interface{}{"available": true, "names": names, "max": max, "gb": m.ZoneNames != nil && m.ZoneNames.GB, "required": required}, nil
}

// ---- radio DMR: zone come elenchi di canali, con un nome di 16 caratteri

// dmrZoneName: nome ASCII di 16 byte (riempimento 0xFF, come le radio)
var dmrZoneName = &zoneNameSpec{Stride: 16, Max: 16, Pad: 0xFF}

// zoneNamesAt: nomi dei record di nome agli indirizzi dati
func zoneNamesAt(img []byte, offs []int) ([]string, int, bool) {
	if len(offs) == 0 {
		return nil, 0, false
	}
	var out []string
	for _, a := range offs {
		if a+16 > len(img) {
			return nil, 0, false
		}
		n, ok := dmrZoneName.decode(img[a : a+16])
		if !ok {
			// nome con caratteri che la radio mostra ma che qui non si scrivono:
			// si mostra la parte leggibile
			n = strings.TrimSpace(strings.Map(func(r rune) rune {
				if r < 0x20 || r > 0x7E {
					return -1
				}
				return r
			}, string(dm32Text(img[a:a+16]))))
		}
		out = append(out, n)
	}
	return out, 16, true
}

func setZoneNamesAt(img []byte, offs []int, names []string) error {
	if len(names) != len(offs) {
		return fmt.Errorf("servono %d nomi, ricevuti %d", len(offs), len(names))
	}
	cur, _, _ := zoneNamesAt(img, offs)
	var recs [][]byte
	for z, n := range names {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("zona %d: il nome non può essere vuoto", z+1)
		}
		rec, err := dmrZoneName.encode(n)
		if err != nil {
			return err
		}
		recs = append(recs, rec)
	}
	for z, rec := range recs {
		if z < len(cur) && cur[z] == strings.TrimSpace(names[z]) {
			continue
		}
		copy(img[offs[z]:offs[z]+16], rec)
	}
	return nil
}

// Baofeng RD-5R / DM-1801 (zone valide nella tabella a 0x8010)
func (f rdtyFam) zoneOffs(m *genModel, img []byte) []int {
	o := rdtyo(m)
	if len(img) < rdtyMemSize {
		return nil
	}
	var offs []int
	for z := 0; z < o.nzones; z++ {
		if rdtyZoneValid(img, z) {
			offs = append(offs, rdtyZoneOff(o, z))
		}
	}
	return offs
}

func (f rdtyFam) ZoneNames(m *genModel, img []byte) ([]string, int, bool) {
	return zoneNamesAt(img, f.zoneOffs(m, img))
}

func (f rdtyFam) SetZoneNames(m *genModel, img []byte, names []string) error {
	offs := f.zoneOffs(m, img)
	if len(offs) == 0 {
		return errors.New("nella memoria non ci sono zone")
	}
	return setZoneNamesAt(img, offs, names)
}

// Baofeng DM-32UV (zone 1..count nelle pagine 0x5C-0x64)
func (f dm32Fam) zoneOffs(img []byte) []int {
	var offs []int
	for z := 1; z <= f.zoneCount(img); z++ {
		offs = append(offs, f.zone(z).F("name").Off)
	}
	return offs
}

func (f dm32Fam) ZoneNames(m *genModel, img []byte) ([]string, int, bool) {
	return zoneNamesAt(img, f.zoneOffs(img))
}

func (f dm32Fam) SetZoneNames(m *genModel, img []byte, names []string) error {
	offs := f.zoneOffs(img)
	if len(offs) == 0 {
		return errors.New("nella memoria non ci sono zone")
	}
	return setZoneNamesAt(img, offs, names)
}
