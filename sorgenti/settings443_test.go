//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"bytes"
	"context"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// 4.43: impostazioni delle Radtel ricavate dai programmi del costruttore.

type twin443 struct {
	off  int
	mask byte
	n    int // scelte (0: non è un elenco)
}

// checkTwin443: la tabella della gemella ha esattamente le voci attese, alle
// posizioni lette nel programma del costruttore; ogni voce si scrive e si rilegge
// toccando solo i suoi bit
func checkTwin443(t *testing.T, kind string, want map[string]twin443) (*setTable, []byte) {
	t.Helper()
	m := genByID[kind]
	tw, ok := twinSettings[kind]
	if m == nil || !ok || !settingsAvailable(kind) {
		t.Fatalf("%s: impostazioni non disponibili", kind)
	}
	img := noChirpBase(t, m)
	tb := settingsFor(kind, img)
	if tb == nil {
		t.Fatalf("%s: nessuna tabella per un'immagine di %d byte", kind, len(img))
	}
	if !strings.HasPrefix(tb.Source, "CHIRP+programma del costruttore «") {
		t.Fatalf("%s: origine %q", kind, tb.Source)
	}
	if len(tb.Settings) != len(want) {
		var keys []string
		for i := range tb.Settings {
			keys = append(keys, tb.Settings[i].Key)
		}
		t.Fatalf("%s: %d impostazioni invece di %d: %v", kind, len(tb.Settings), len(want), keys)
	}
	rnd := rand.New(rand.NewSource(int64(len(kind)) * 7907))
	for i := range tb.Settings {
		d := &tb.Settings[i]
		w, ok := want[d.Key]
		if !ok {
			t.Fatalf("%s: voce inattesa %s", kind, d.Key)
		}
		offs, masks, err := d.Enc.fields()
		if err != nil || len(offs) != 1 || offs[0] != w.off || masks[0][0] != w.mask {
			t.Fatalf("%s %s: posizione %#x maschera %X (attese %#x %X)", kind, d.Key, offs, masks, w.off, w.mask)
		}
		if tw.within != [2]int{} && (w.off < tw.within[0] || w.off >= tw.within[1]) {
			t.Fatalf("%s %s: fuori dal blocco verificato", kind, d.Key)
		}
		if w.n != len(d.Options) {
			t.Fatalf("%s %s: %d scelte invece di %d", kind, d.Key, len(d.Options), w.n)
		}
		cur, _ := readSetting(d, img)
		for r := 0; r < 8; r++ {
			val, ok := randSetVal(rnd, d, cur)
			if !ok {
				t.Fatalf("%s %s: nessun valore di prova", kind, d.Key)
			}
			before := append([]byte(nil), img...)
			if err := writeSetting(d, img, val); err != nil {
				t.Fatalf("%s %s=%q: %v", kind, d.Key, val, err)
			}
			before[w.off] = before[w.off]&^w.mask | img[w.off]&w.mask
			if !bytes.Equal(before, img) {
				t.Fatalf("%s %s=%q: modificati bit fuori dalla maschera", kind, d.Key, val)
			}
			if back, err := readSetting(d, img); err != nil || back != val {
				t.Fatalf("%s %s: riletto %q invece di %q (%v)", kind, d.Key, back, val, err)
			}
			cur = val
		}
	}
	return tb, img
}

func set443(t *testing.T, tb *setTable, key string) *setDef {
	t.Helper()
	for i := range tb.Settings {
		if tb.Settings[i].Key == key {
			return &tb.Settings[i]
		}
	}
	t.Fatalf("voce %s assente", key)
	return nil
}

// Radtel RT-400 («JJCC-7700 CPS») e RT-330 («Radio CPS»): blocco a 0x9000, posizioni
// trascritte da GetFunCfgImf_Part1 / 2 / 3 dei due programmi
func TestSettings443JC(t *testing.T) {
	base := map[string]twin443{
		"sql": {0x9000, 0x0F, 0}, "save": {0x9001, 0x03, 4}, "vox": {0x9002, 0x0F, 10}, "abr": {0x9003, 0x0F, 9},
		"tdr": {0x9004, 0x01, 0}, "tot": {0x9005, 0x0F, 9}, "beep": {0x9006, 0x01, 0}, "voice": {0x9007, 0x01, 0},
		"language": {0x9008, 0x01, 2}, "dtmfst": {0x9009, 0x03, 4}, "screv": {0x900A, 0x03, 3}, "pttid": {0x900B, 0x03, 4},
		"pttlt": {0x900C, 0x07, 7}, "mdfa": {0x900D, 0x03, 3}, "mdfb": {0x900E, 0x03, 3},
		"autolk": {0x9010, 0x03, 4}, "almode": {0x9011, 0x03, 3}, "alarmsound": {0x9012, 0x01, 0}, "ste": {0x9014, 0x01, 0},
		"rpste": {0x9015, 0x0F, 11}, "rptrl": {0x9016, 0x0F, 11}, "roger": {0x9017, 0x03, 3}, "fmradio": {0x9019, 0x01, 0},
		"vfomra": {0x901A, 0x01, 2}, "vfomrb": {0x901A, 0x10, 2}, "kblock": {0x901B, 0x01, 0}, "ponmsg": {0x901C, 0x01, 2},
		"tone": {0x901E, 0x03, 4}, "voxd": {0x9020, 0x0F, 16}, "menuquit": {0x9021, 0x0F, 11},
	}
	cp := func(extra map[string]twin443) map[string]twin443 {
		w := map[string]twin443{}
		for k, v := range base {
			w[k] = v
		}
		for k, v := range extra {
			w[k] = v
		}
		return w
	}
	// RT-400: in più trasmissione del doppio ascolto, «QT save» e identificativo ANI
	w400 := cp(map[string]twin443{"dualtx": {0x9013, 0x03, 3}, "qtsave": {0x9029, 0x03, 3}, "ani": {0x902A, 0x3F, 0}})
	tb, img := checkTwin443(t, "rt-rt400", w400)
	// RT-330: fine trasmissione a due scelte
	w330 := cp(nil)
	w330["roger"] = twin443{0x9017, 0x03, 2}
	tb3, img3 := checkTwin443(t, "rt-rt330", w330)

	for _, x := range []struct {
		tb  *setTable
		img []byte
	}{{tb, img}, {tb3, img3}} {
		// TOT come nei programmi: valore 3 = 90 s, valore 8 = 240 s
		d := set443(t, x.tb, "tot")
		if d.Options[0] != "Off" || d.Options[3] != "90 seconds" || d.Options[8] != "240 seconds" {
			t.Fatalf("TOT: %v", d.Options)
		}
		x.img[0x9005] = 0x03
		if v, err := readSetting(d, x.img); err != nil || v != "90 seconds" {
			t.Fatalf("TOT 3: %q %v", v, err)
		}
		if err := writeSetting(d, x.img, "240 seconds"); err != nil || x.img[0x9005] != 0x08 {
			t.Fatalf("TOT 240 s: %v, byte %02X", err, x.img[0x9005])
		}
		// la tabella di CHIRP della RT-470 non è stata toccata
		for _, s := range chirpSettings()["RT470Radio"].Settings {
			if s.Key == "tot" && (len(s.Options) != 6 || s.Options[3] != "120 seconds") {
				t.Fatalf("la tabella di CHIRP è stata modificata: %v", s.Options)
			}
		}
		// tasti e voci non confermate assenti
		for i := range x.tb.Settings {
			if k := x.tb.Settings[i].Key; strings.Contains(k, "key") || k == "rxendtail" {
				t.Fatalf("voce %s presente", k)
			}
		}
	}
	// RT-400: «QT save» nell'ordine del programma (tutti, TX, RX)
	if d := set443(t, tb, "qtsave"); d.Options[1] != "TX" || d.Options[2] != "RX" {
		t.Fatalf("QT save: %v", d.Options)
	}
	// RT-330: luce, prima scelta «Off»; fine trasmissione spento / acceso
	if d := set443(t, tb3, "abr"); d.Options[0] != "Off" || d.Options[8] != "3 minutes" {
		t.Fatalf("luce: %v", d.Options)
	}
	if d := set443(t, tb3, "roger"); len(d.Options) != 2 || d.Options[0] != "OFF" {
		t.Fatalf("fine trasmissione: %v", d.Options)
	}
	// i blocchi stanno nelle aree che il driver legge e scrive
	for _, kind := range []string{"rt-rt400", "rt-rt330"} {
		o := genByID[kind].Opt.(*jc8810Opt)
		ok := false
		for _, r := range o.ranges {
			ok = ok || (r[0] <= 0x9000 && r[1] >= 0x9030)
		}
		if !ok || o.memsize < 0x9040 {
			t.Fatalf("%s: il blocco 0x9000 non è fra le aree del driver (%v)", kind, o.ranges)
		}
	}
	// le RT-470 di CHIRP hanno ancora la tabella intera
	m := genByID["rt-rt470"]
	if m == nil || m.NoChirp || settingsFor("rt-rt470", make([]byte, 0xB300)) != chirpSettings()[m.RClass] {
		t.Fatal("RT-470: tabella di CHIRP non più intera")
	}
}

// rs443: posizione di una voce della RT-871A (byte, larghezza) e numero di scelte
// (0 per le voci sì/no e per i numeri)
type rs443 struct{ off, width, n int }

// 4.43, Radtel RT-817A / RT-871A: le voci stanno dove le mette il programma del
// costruttore (posizioni riscritte qui a mano dall'analisi, non prese dalla
// tabella), ogni voce tocca solo i suoi byte, e ciò che si imposta arriva alla
// radio simulata e torna uguale
func TestSettings443RT871A(t *testing.T) {
	const kind = "rtl-rt871a"
	m := genByID[kind]
	f := m.Fam.(iradioFam)
	p := f.proto()
	if !settingsAvailable(kind) {
		t.Fatal("impostazioni non disponibili")
	}
	img := bytes.Repeat([]byte{0xFF}, p.size)
	if settingsFor(kind, img[:0xFFF]) != nil {
		t.Fatal("tabella offerta per un'immagine troppo corta")
	}
	for i := 0; i < 0x1000; i++ {
		img[i] = 0
	}
	tb := settingsFor(kind, img)
	if tb == nil || !strings.HasPrefix(tb.Source, "programma del costruttore «Radtel RT-871A CPS»") {
		t.Fatalf("tabella o origine: %+v", tb)
	}
	want := map[string]rs443{
		"saver_timer": {0x11, 1, 43}, "voice": {0x5C, 1, 0}, "beep": {0x5D, 1, 0}, "key_lock": {0x5E, 1, 0}, "lock_timer": {0x5F, 1, 43},
		"lcd_timer": {0x62, 1, 43}, "save_mode": {0x63, 1, 4}, "save_start": {0x64, 1, 43}, "menu_timer": {0x65, 1, 43},
		"dw": {0x66, 1, 0}, "talkaround": {0x67, 1, 3}, "alarm_type": {0x68, 1, 3}, "tx_pri": {0x7E, 1, 2}, "main_ptt": {0x7F, 1, 2},
		"step": {0x80, 1, 14}, "main_range": {0x83, 1, 2}, "mode_a": {0x84, 1, 3}, "show_a": {0x85, 1, 2}, "mode_b": {0x89, 1, 3},
		"show_b": {0x8A, 1, 2}, "lock1": {0x8E, 1, 3}, "lock1_start": {0x8F, 2, 0}, "lock1_end": {0x91, 2, 0},
		"lock2": {0x93, 1, 3}, "lock2_start": {0x94, 2, 0}, "lock2_end": {0x96, 2, 0}, "lock3": {0x98, 1, 3}, "lock3_start": {0x99, 2, 0},
		"lock3_end": {0x9B, 2, 0}, "lock4": {0x9D, 1, 3}, "lock4_start": {0x9E, 2, 0}, "lock4_end": {0xA0, 2, 0},
		"scan_dir": {0xA2, 1, 2}, "scan_mode": {0xA3, 1, 3}, "scan_return": {0xA4, 1, 2}, "scan_dwell": {0xA5, 1, 31},
		"voice_refresh": {0xA6, 2, 21}, "rssi_refresh": {0xA8, 2, 21}, "fs1_s": {0xAA, 1, 20}, "fs1_l": {0xAB, 1, 20},
		"fs2_s": {0xAC, 1, 20}, "fs2_l": {0xAD, 1, 20}, "freq_digits": {0xEA, 1, 2}, "tone": {0x100, 2, 0}, "sql": {0x102, 1, 0},
		"mic_gain": {0x105, 1, 0}, "rx_volume": {0x106, 1, 0}, "tx_start_beep": {0x10B, 1, 0}, "roger": {0x10C, 1, 5}, "vox": {0x10D, 1, 0},
		"vox_th": {0x10E, 1, 0}, "vox_delay": {0x10F, 1, 11}, "detect_range": {0x110, 1, 11}, "relay_delay": {0x111, 2, 21},
		"glitch": {0x114, 1, 0}, "dtmf_delay": {0x200, 1, 21}, "dtmf_duration": {0x201, 1, 18}, "dtmf_interval": {0x202, 1, 18},
		"dtmf_mode": {0x203, 1, 4}, "dtmf_select": {0x204, 1, 16}, "dtmf_display": {0x205, 1, 0}, "dtmf_gain": {0x206, 1, 0},
		"dtmf_th": {0x207, 1, 0}, "dtmf_remote": {0x208, 1, 0}, "time_cal": {0x209, 1, 0},
	}
	for i := 0; i < 10; i++ {
		want["key"+strconv.Itoa(i)] = rs443{0xB0 + i, 1, 20}
	}
	if len(tb.Settings) != len(want) {
		t.Fatalf("%d impostazioni invece di %d", len(tb.Settings), len(want))
	}
	used := map[int]string{}
	rnd := rand.New(rand.NewSource(871))
	for i := range tb.Settings {
		d := &tb.Settings[i]
		w, ok := want[d.Key]
		if !ok {
			t.Fatalf("voce inattesa %s", d.Key)
		}
		offs, masks, err := d.Enc.fields()
		if err != nil || len(offs) != 1 || offs[0] != w.off || len(masks[0]) != w.width || w.n != len(d.Options) || len(d.ItOpts) != w.n {
			t.Fatalf("%s: posizione %#x, %d byte, %d scelte (attesi %#x, %d, %d)", d.Key, offs, len(masks[0]), len(d.Options), w.off, w.width, w.n)
		}
		if d.ItLabel == "" || d.Group == "" {
			t.Fatalf("%s: senza nome o gruppo", d.Key)
		}
		for k := 0; k < w.width; k++ {
			if o := used[w.off+k]; o != "" {
				t.Fatalf("%s e %s usano lo stesso byte %#x", d.Key, o, w.off+k)
			}
			used[w.off+k] = d.Key
			if w.off+k >= rt871aRegions[0][1]*p.block {
				t.Fatalf("%s: fuori dal blocco delle impostazioni che il driver scrive", d.Key)
			}
		}
		cur, err := readSetting(d, img)
		if err != nil && !strings.HasPrefix(d.Key, "lock") { // le gamme partono da 18 MHz: lo zero non è un valore
			t.Fatalf("%s: memoria azzerata non leggibile: %v", d.Key, err)
		}
		for r := 0; r < 8; r++ {
			val, ok := randSetVal(rnd, d, cur)
			if !ok {
				t.Fatalf("%s: nessun valore di prova", d.Key)
			}
			before := append([]byte(nil), img...)
			if err := writeSetting(d, img, val); err != nil {
				t.Fatalf("%s=%q: %v", d.Key, val, err)
			}
			copy(before[w.off:w.off+w.width], img[w.off:w.off+w.width])
			if !bytes.Equal(before, img) {
				t.Fatalf("%s=%q: modificati byte di altre voci", d.Key, val)
			}
			if back, err := readSetting(d, img); err != nil || back != val {
				t.Fatalf("%s: riletto %q invece di %q (%v)", d.Key, back, val, err)
			}
			cur = val
		}
	}
	// valori come li scrive il programma: indice della scelta, numeri con il byte basso per primo
	put := func(key, val string) {
		t.Helper()
		if err := writeSetting(set443(t, tb, key), img, val); err != nil {
			t.Fatalf("%s=%s: %v", key, val, err)
		}
	}
	put("sql", "7")
	put("tone", "1750")
	put("relay_delay", "12")
	put("lock2", "1")
	put("lock2_start", "430")
	put("lock2_end", "440")
	put("lock_timer", "42")
	put("vox_th", "200")
	put("roger", "4")
	put("key9", "19")
	put("vox", "true")
	chk := map[int]byte{0x102: 7, 0x100: 0xD6, 0x101: 0x06, 0x111: 12, 0x112: 0, 0x93: 1, 0x94: 0xAE, 0x95: 0x01, 0x96: 0xB8, 0x97: 0x01,
		0x5F: 42, 0x10E: 200, 0x10C: 4, 0xB9: 19, 0x10D: 1}
	for a, v := range chk {
		if img[a] != v {
			t.Fatalf("byte %#x = %#x invece di %#x", a, img[a], v)
		}
	}
	if d := set443(t, tb, "lock_timer"); d.ItOpts[0] != "OFF" || d.ItOpts[1] != "5 s" || d.ItOpts[4] != "30 s" || d.ItOpts[42] != "600 s" {
		t.Fatalf("tempi: %v", d.ItOpts)
	}
	if d := set443(t, tb, "dtmf_duration"); d.ItOpts[0] != "30 ms" || d.ItOpts[17] != "200 ms" {
		t.Fatalf("durate DTMF: %v", d.ItOpts)
	}
	for _, bad := range [][2]string{{"sql", "11"}, {"tone", "20001"}, {"lock1_start", "17"}, {"lock1_end", "1001"}, {"mic_gain", "32"}, {"step", "14"}} {
		if err := writeSetting(set443(t, tb, bad[0]), img, bad[1]); err == nil {
			t.Fatalf("%s=%s accettato", bad[0], bad[1])
		}
	}
	// scrittura e rilettura con la radio simulata: le impostazioni arrivano e tornano uguali
	e := f.emu(m, bytes.Repeat([]byte{0xFF}, p.size)).(*iradioEmu)
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	if _, err := f.Upload(context.Background(), m, "COM7", img, noProg); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(e.img[:0x1000], img[:0x1000]) {
		t.Fatal("le impostazioni non sono arrivate alla radio simulata")
	}
	got, _, err := f.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	t2 := settingsFor(kind, got)
	for key, val := range map[string]string{"sql": "7", "tone": "1750", "relay_delay": "12", "lock2_start": "430", "vox": "true", "key9": "19"} {
		if back, err := readSetting(set443(t, t2, key), got); err != nil || back != val {
			t.Fatalf("dopo la rilettura %s = %q (%v)", key, back, err)
		}
	}
}

// checkProv443: tabella ricavata dal programma del costruttore: ha esattamente le
// voci attese alle posizioni attese (riscritte a mano nella prova), nessun bit è
// usato da due voci, ogni voce si scrive e si rilegge toccando solo i suoi bit
func checkProv443(t *testing.T, kind string, want map[string]twin443) (*setTable, []byte) {
	t.Helper()
	m := genByID[kind]
	if m == nil || !settingsAvailable(kind) {
		t.Fatalf("%s: impostazioni non disponibili", kind)
	}
	size := genMinSize(m)
	if o, ok := m.Opt.(*rt830Opt); ok {
		size = o.size()
	}
	if n, ok := provSize443[kind]; ok {
		size = n
	}
	img := make([]byte, size)
	if len(img) == 0 {
		t.Fatalf("%s: dimensione dell'immagine sconosciuta", kind)
	}
	tb := settingsFor(kind, img)
	if tb == nil || !strings.HasPrefix(tb.Source, "programma del costruttore «") {
		t.Fatalf("%s: tabella o origine (%d byte)", kind, len(img))
	}
	if len(tb.Settings) != len(want) {
		t.Fatalf("%s: %d impostazioni invece di %d", kind, len(tb.Settings), len(want))
	}
	used := map[int]byte{}
	rnd := rand.New(rand.NewSource(int64(len(kind)) * 104729))
	for i := range tb.Settings {
		d := &tb.Settings[i]
		w, ok := want[d.Key]
		if !ok {
			t.Fatalf("%s: voce inattesa %s", kind, d.Key)
		}
		offs, masks, err := d.Enc.fields()
		if err != nil || len(offs) != 1 || offs[0] != w.off || len(masks[0]) != 1 || masks[0][0] != w.mask {
			t.Fatalf("%s %s: posizione %#x maschera %X (attese %#x %X)", kind, d.Key, offs, masks, w.off, w.mask)
		}
		if used[w.off]&w.mask != 0 {
			t.Fatalf("%s %s: bit già usati da un'altra voce", kind, d.Key)
		}
		used[w.off] |= w.mask
		if w.n != len(d.Options) || len(d.ItOpts) != w.n || d.ItLabel == "" || d.Group == "" {
			t.Fatalf("%s %s: %d scelte invece di %d, o nome mancante", kind, d.Key, len(d.Options), w.n)
		}
		// le scelte stanno tutte nei bit della maschera
		if top := int(w.mask >> uint(bitsLow443(w.mask))); w.n > top+1 {
			t.Fatalf("%s %s: %d scelte in una maschera %X", kind, d.Key, w.n, w.mask)
		}
		cur, err := readSetting(d, img)
		if err != nil {
			t.Fatalf("%s %s: memoria azzerata non leggibile: %v", kind, d.Key, err)
		}
		for r := 0; r < 8; r++ {
			val, ok := randSetVal(rnd, d, cur)
			if !ok {
				t.Fatalf("%s %s: nessun valore di prova", kind, d.Key)
			}
			before := append([]byte(nil), img...)
			if err := writeSetting(d, img, val); err != nil {
				t.Fatalf("%s %s=%q: %v", kind, d.Key, val, err)
			}
			before[w.off] = before[w.off]&^w.mask | img[w.off]&w.mask
			if !bytes.Equal(before, img) {
				t.Fatalf("%s %s=%q: modificati bit fuori dalla maschera", kind, d.Key, val)
			}
			if back, err := readSetting(d, img); err != nil || back != val {
				t.Fatalf("%s %s: riletto %q invece di %q (%v)", kind, d.Key, back, val, err)
			}
			cur = val
		}
	}
	return tb, img
}

// provSize443: lunghezza dell'immagine delle radio senza mappa dichiarata;
// provRadio443: memoria di partenza della radio simulata (se non basta azzerata)
var (
	provSize443  = map[string]int{}
	provRadio443 = map[string][]byte{}
)

func bitsLow443(mask byte) int {
	n := 0
	for mask>>uint(n)&1 == 0 {
		n++
	}
	return n
}

// cable443: le voci impostate arrivano alla radio simulata e tornano uguali
func cable443(t *testing.T, kind string, img []byte, vals map[string]string) {
	t.Helper()
	m := genByID[kind]
	tb := settingsFor(kind, img)
	for k, v := range vals {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s %s=%s: %v", kind, k, v, err)
		}
	}
	radio := make([]byte, len(img))
	if base, ok := provRadio443[kind]; ok { // radio simulata con un contenuto di partenza preciso
		radio = append([]byte{}, base...)
	}
	if _, ok := m.Opt.(*rt830Opt); ok { // radio senza password: blocco della password a FF
		for i := 0; i < 16; i++ {
			radio[i], img[i] = 0xFF, 0xFF
		}
	}
	e := emuFor(m, radio)
	if e == nil {
		t.Fatalf("%s: nessuna radio simulata", kind)
	}
	old := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		a, b := newPipe()
		a.baud = baud
		go e.serve(b)
		return a, nil
	}
	defer func() { openPortFunc = old }()
	if _, err := m.Fam.Upload(context.Background(), m, "COM7", img, noProg); err != nil {
		t.Fatalf("%s: scrittura: %v", kind, err)
	}
	got, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatalf("%s: rilettura: %v", kind, err)
	}
	t2 := settingsFor(kind, got)
	if t2 == nil {
		t.Fatalf("%s: nessuna tabella per la memoria riletta (%d byte)", kind, len(got))
	}
	for k, v := range vals {
		if back, err := readSetting(set443(t, t2, k), got); err != nil || back != v {
			t.Fatalf("%s: dopo scrittura e rilettura %s = %q invece di %q (%v)", kind, k, back, v, err)
		}
	}
}

// 4.43, Radtel RT-830 e RT-850: posizioni e bit come nei programmi del costruttore
// (0x10 + posizione nel blocco delle impostazioni)
func TestSettings443RT830(t *testing.T) {
	w830 := map[string]twin443{
		"lang": {0x31, 0x80, 2}, "voice": {0x31, 0x40, 0}, "beep": {0x31, 0x20, 0}, "vox_on": {0x31, 0x10, 0}, "tx_pri": {0x31, 0x08, 2},
		"save": {0x31, 0x07, 6}, "dual": {0x32, 0x80, 0}, "area": {0x32, 0x40, 2}, "roger": {0x32, 0x30, 4}, "sql": {0x32, 0x0F, 10},
		"vox_level": {0x33, 0xF0, 10}, "vox_delay": {0x33, 0x0F, 10}, "display_mode": {0x34, 0x80, 2}, "ch_state": {0x34, 0x40, 2},
		"tot": {0x34, 0x3F, 41}, "offset_dir": {0x35, 0xC0, 3}, "auto_lock": {0x35, 0x3F, 41}, "light_on": {0x36, 0x80, 0},
		"light_time": {0x36, 0x3F, 41}, "step": {0x3B, 0xF0, 10}, "scramble": {0x3B, 0x0F, 9}, "ch_a": {0x3C, 0xFF, 128},
		"ch_b": {0x3D, 0xFF, 128}, "radio_type": {0x3E, 0x80, 2}, "radio_standby": {0x3E, 0x40, 0}, "tail": {0x3E, 0x20, 2},
	}
	tb, img := checkProv443(t, "rtl-rt830", w830)
	for i := range img {
		img[i] = 0
	}
	// come li scrive il programma: 0x21 = lingua<<7 + voce<<6 + beep<<5 + VOX<<4 + priorità<<3 + risparmio
	for k, v := range map[string]string{"lang": "1", "voice": "true", "vox_on": "true", "save": "5", "roger": "3", "sql": "9", "tot": "40",
		"offset_dir": "2", "light_time": "1", "step": "9", "scramble": "8", "ch_a": "127", "tail": "1"} {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s=%s: %v", k, v, err)
		}
	}
	for a, v := range map[int]byte{0x31: 0x80 | 0x40 | 0x10 | 5, 0x32: 3<<4 | 9, 0x34: 40, 0x35: 2 << 6, 0x36: 1, 0x3B: 9<<4 | 8, 0x3C: 127, 0x3E: 0x20} {
		if img[a] != v {
			t.Fatalf("RT-830 byte %#x = %#x invece di %#x", a, img[a], v)
		}
	}
	if d := set443(t, tb, "tot"); d.ItOpts[0] != "OFF" || d.ItOpts[1] != "15 s" || d.ItOpts[40] != "600 s" {
		t.Fatalf("tempi: %v", d.ItOpts)
	}
	cable443(t, "rtl-rt830", img, map[string]string{"sql": "4", "tot": "12", "dual": "true", "ch_b": "99", "radio_type": "1"})

	w850 := map[string]twin443{
		"screen_time": {0x39, 0xFF, 41}, "led_time": {0x3A, 0xFF, 41}, "auto_lock": {0x3B, 0xFF, 41}, "fs1_s": {0x3C, 0xFF, 7},
		"fs1_l": {0x3D, 0xFF, 7}, "fs2_s": {0x3E, 0xFF, 7}, "fs2_l": {0x3F, 0xFF, 7}, "voice": {0x41, 0xC0, 3}, "beep": {0x41, 0x20, 0},
		"roger": {0x41, 0x18, 4}, "save": {0x41, 0x07, 6}, "display_mode": {0x42, 0x80, 2}, "dual": {0x42, 0x40, 0}, "area": {0x42, 0x20, 2},
		"radio_standby": {0x42, 0x10, 0}, "sql": {0x42, 0x0F, 10}, "vox_level": {0x43, 0xF0, 10}, "vox_delay": {0x43, 0x0F, 10},
		"screen_light": {0x44, 0xC0, 3}, "offset_dir": {0x44, 0x20, 2}, "alarm": {0x44, 0x10, 2}, "scramble": {0x44, 0x0F, 9},
		"screen_on": {0x45, 0x80, 0}, "led_on": {0x45, 0x40, 0}, "key_lock": {0x45, 0x20, 0}, "radio_on": {0x45, 0x10, 0},
		"compander": {0x45, 0x08, 0}, "offset_on": {0x45, 0x04, 0}, "freq_lock": {0x45, 0x02, 0}, "vox_on": {0x45, 0x01, 0},
		"step": {0x46, 0xF0, 10}, "tx_pri": {0x46, 0x08, 2}, "tail": {0x46, 0x07, 4}, "ch_a": {0x4D, 0xFF, 128}, "ch_b": {0x4E, 0xFF, 128},
		"tot": {0x4F, 0xFF, 41},
	}
	tb, img = checkProv443(t, "rtl-rt850", w850)
	for i := range img {
		img[i] = 0
	}
	for k, v := range map[string]string{"voice": "2", "beep": "true", "roger": "3", "save": "4", "screen_light": "2", "alarm": "1", "scramble": "8",
		"compander": "true", "vox_on": "true", "step": "9", "tail": "3", "tot": "40", "fs2_l": "6", "radio_standby": "true", "area": "1"} {
		if err := writeSetting(set443(t, tb, k), img, v); err != nil {
			t.Fatalf("%s=%s: %v", k, v, err)
		}
	}
	for a, v := range map[int]byte{0x41: 2<<6 | 0x20 | 3<<3 | 4, 0x42: 0x20 | 0x10, 0x44: 2<<6 | 0x10 | 8, 0x45: 0x08 | 0x01, 0x46: 9<<4 | 3, 0x4F: 40, 0x3F: 6} {
		if img[a] != v {
			t.Fatalf("RT-850 byte %#x = %#x invece di %#x", a, img[a], v)
		}
	}
	cable443(t, "rtl-rt850", img, map[string]string{"sql": "7", "tot": "20", "fs1_s": "1", "ch_a": "64", "screen_time": "3"})
}

// 4.43: le radio nuove nell'interfaccia: apertura da file, elenco delle voci,
// modifica, salvataggio e riapertura
func TestSettings443API(t *testing.T) {
	blank := func(kind string) []byte {
		m := genByID[kind]
		if o, ok := m.Opt.(*rt830Opt); ok {
			b := make([]byte, o.size())
			for i := 0; i < 16; i++ {
				b[i] = 0xFF // nessuna password
			}
			for i := o.chOff(); i < len(b); i++ {
				b[i] = 0xFF // canali liberi
			}
			return b
		}
		if f, ok := m.Fam.(iradioFam); ok {
			b := bytes.Repeat([]byte{0xFF}, f.proto().size)
			for i := 0; i < 0x1000; i++ {
				b[i] = 0
			}
			return b
		}
		return noChirpBase(t, m)
	}
	apiCase442(t, "rtl-rt871a", "tone", "1750", 7, blank("rtl-rt871a"))
	apiCase442(t, "rtl-rt830", "tot", "12", 3, blank("rtl-rt830"))
	apiCase442(t, "rtl-rt850", "fs1_s", "3", 4, blank("rtl-rt850"))
	apiCase442(t, "rt-rt400", "tot", "90 seconds", 0, blank("rt-rt400"))
	apiCase442(t, "rt-rt330", "tot", "240 seconds", 0, blank("rt-rt330"))
}
