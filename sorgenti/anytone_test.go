//go:build linux

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Famiglia AnyTone (4.45) confrontata con il codice di dmrconfig (anytone_ht.c,
// serial.c), compilato con verifica/dmrconfig/tcpdev.c:
//   DMRCONFIG=…/dmrconfig-tcp DMRCONFIG_EX=…/dmrconfig/examples
// 1) memoria di prova creata da dmrconfig con un suo file di esempio; 2) lettura:
// stessi canali; 3) modifiche fatte qui, poi dmrconfig stampa la configurazione
// della memoria modificata e la riapplica alla memoria di partenza: memoria
// identica; 4) protocollo: lettura di dmrconfig dalla radio simulata (attraverso
// una pseudo-terminale) uguale a quella di questo programma; scrittura: ogni byte
// scritto da questo programma è uguale a quello scritto da dmrconfig allo stesso
// indirizzo, e la radio finisce con la stessa memoria.
// Le altre prove (TestAnytone…) non hanno bisogno di dmrconfig.

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// filePort: porta della radio simulata sopra un file (lato principale della
// pseudo-terminale)
type filePort struct{ f *os.File }

func (p *filePort) Write(b []byte) error { _, err := p.f.Write(b); return err }
func (p *filePort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	_, err := io.ReadFull(p.f, b)
	return err
}
func (p *filePort) Purge() error      { return nil }
func (p *filePort) SetBaud(int) error { return nil }
func (p *filePort) Close() error      { return p.f.Close() }

// servePty: pseudo-terminale collegata alla radio simulata; restituisce il
// percorso da dare a dmrconfig (DMRCONFIG_SERIAL)
func servePty(t *testing.T, serve func(p Port)) (string, func()) {
	t.Helper()
	fd, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skip("pseudo-terminale non disponibile: ", err)
	}
	var n, unlock uint32
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); e != 0 {
		syscall.Close(fd)
		t.Skip("pseudo-terminale non disponibile: ", e)
	}
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n))); e != 0 {
		syscall.Close(fd)
		t.Skip("pseudo-terminale non disponibile: ", e)
	}
	// senza eco e senza trasformazioni sul lato della radio
	var tio syscall.Termios
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&tio))); e == 0 {
		tio.Iflag, tio.Oflag, tio.Lflag = 0, 0, 0
		tio.Cc[syscall.VMIN], tio.Cc[syscall.VTIME] = 1, 0
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&tio)))
	}
	f := os.NewFile(uintptr(fd), "ptmx")
	done := make(chan struct{})
	go func() {
		defer close(done)
		serve(&filePort{f})
	}()
	return fmt.Sprintf("/dev/pts/%d", n), func() { f.Close(); <-done }
}

// canale della tabella di dmrconfig confrontato con quello letto qui
func atSameAsConf(f []string, c *Chan) bool {
	rx, _ := strconv.ParseFloat(f[2], 64)
	off, _ := strconv.ParseFloat(f[3], 64)
	rxHz := int64(rx*1e5+0.5) * 10
	txHz := rxHz + int64(off*1e5+0.5*sign(off))*10
	pw := map[string]int{"Turbo": 0, "High": 0, "Mid": 1, "Low": 2}[f[4]]
	if c.TxOff {
		txHz = rxHz
	}
	if rxHz != c.Rx || txHz != c.Tx || strings.TrimRight(strings.ReplaceAll(f[1], "_", " "), " ") != c.Name || pw != c.Power ||
		(f[0] == "Digital") != c.Dig || (f[7] == "+") != c.TxOff {
		return false
	}
	if f[0] == "Analog" {
		tone := func(s string) string {
			if s == "-" {
				return ""
			}
			if strings.HasPrefix(s, "D") {
				return s
			}
			v, _ := strconv.ParseFloat(s, 64)
			return fmtTone(int(v*10 + 0.5))
		}
		return tone(f[10]) == c.RxTone && tone(f[11]) == c.TxTone && (f[12] == "12.5") == c.Narrow
	}
	return true
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func TestAnytoneDmrconfig(t *testing.T) {
	bin, ex := os.Getenv("DMRCONFIG"), os.Getenv("DMRCONFIG_EX")
	if bin == "" || ex == "" {
		t.Skip("DMRCONFIG o DMRCONFIG_EX non impostati")
	}
	for _, tc := range []struct{ id, conf string }{
		{"at-d868uv", "d868uv-rmham-2018-10-20.conf"},
		{"bt-dmr6x2", "dmr6x2-rmham-2018-10-20.conf"},
		{"at-d868uv", "d868uv-norcal-ka7qqv-2017-11-04.conf"},
	} {
		tc := tc
		t.Run(tc.id+"/"+tc.conf, func(t *testing.T) {
			m := genByID[tc.id]
			dir := t.TempDir()
			run := func(env []string, args ...string) string {
				cmd := exec.Command(bin, args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), env...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("dmrconfig %v: %v\n%s", args, err, out)
				}
				return string(out)
			}
			// 1) memoria di prova
			_ = os.WriteFile(filepath.Join(dir, "blank.img"), atSynthImage(m), 0o644)
			run(nil, "-c", "blank.img", filepath.Join(ex, tc.conf))
			base, err := os.ReadFile(filepath.Join(dir, "device.img"))
			if err != nil {
				t.Fatal(err)
			}
			_ = os.WriteFile(filepath.Join(dir, "base.img"), base, 0o644)
			if !m.Fam.Match(m, base) {
				t.Fatal("la memoria di dmrconfig non è riconosciuta")
			}
			// 2) lettura
			rows := rdtyDmrconfigChans(run(nil, "base.img"))
			n := 0
			for i := 0; i < m.Slots; i++ {
				c := m.Fam.Get(m, base, i)
				f, ok := rows[i+1]
				if (c != nil) != ok {
					t.Fatalf("canale %d: qui %v, dmrconfig %v", i+1, c, f)
				}
				if c == nil {
					continue
				}
				n++
				if !atSameAsConf(f, c) {
					t.Errorf("canale %d: dmrconfig %v, qui %+v", i+1, f, c)
				}
			}
			if n < 10 {
				t.Fatalf("solo %d canali nella memoria di prova", n)
			}
			// 3) modifiche: cancellazioni (anche di canali in zone e liste di
			// scansione), canali modificati (anche da DMR ad analogico e viceversa),
			// nuovi, solo ricezione, toni
			mine := append([]byte{}, base...)
			type edit struct {
				n int
				c *Chan
			}
			var edits []edit
			inZone, inScan, ana, dig, keep := -1, -1, -1, -1, -1
			for z := 0; z < atNZones && inZone < 0; z++ {
				if atZoneValid(base, z) {
					if v := atList(atZoneList(base, z)); len(v) > 2 {
						inZone = int(v[1])
					}
				}
			}
			for s := 0; s < atNScan && inScan < 0; s++ {
				if atScanValid(base, s) {
					if v := atList(atScanRec(base, s)[32:132]); len(v) > 1 && int(v[0]) != inZone {
						inScan = int(v[0])
					}
				}
			}
			for i := 0; i < m.Slots; i++ {
				c := m.Fam.Get(m, base, i)
				if c == nil || i == inZone || i == inScan {
					continue
				}
				switch {
				case c.Dig && dig < 0:
					dig = i
				case c.Dig && keep < 0:
					keep = i
				case !c.Dig && ana < 0:
					ana = i
				}
			}
			if ana < 0 { // nessun canale analogico: un altro canale DMR diventa analogico
				for i := keep + 1; i < m.Slots && ana < 0; i++ {
					if c := m.Fam.Get(m, base, i); c != nil && i != inZone && i != inScan {
						ana = i
					}
				}
			}
			var free []int
			for i := 3900; i < 4000 && len(free) < 4; i++ {
				if !atValid(base, i) {
					free = append(free, i)
				}
			}
			if ana < 0 || dig < 0 || keep < 0 || len(free) < 4 || inZone < 0 {
				t.Fatalf("memoria di prova inadatta: %d %d %d %d %d %v", ana, dig, keep, inZone, inScan, free)
			}
			kc := *m.Fam.Get(m, base, keep) // canale DMR modificato restando DMR: cambia solo il nome e la frequenza
			kc.Name, kc.Rx, kc.Tx = "DMR MODIFICATO", 438300000, 430700000
			if inScan >= 0 { // l'esempio «norcal» non ha liste di scansione
				edits = append(edits, edit{inScan, nil})
			}
			edits = append(edits,
				edit{inZone, nil},
				edit{ana, &Chan{Rx: 433500000, Tx: 433500000, RxTone: "D023N", TxTone: "88.5", Power: 1, Narrow: true, Name: "R0 ROMA"}},
				edit{dig, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "D754I", Power: 2, Name: "RV48"}},
				edit{keep, &kc},
				edit{free[0], &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Narrow: true, Power: 0, Name: "DMR IR0UGN"}},
				edit{free[1], &Chan{Rx: 145500000, Tx: 145500000, TxOff: true, Power: 2, Narrow: true, Name: "ASCOLTO"}},
				edit{free[2], &Chan{Rx: 145750000, Tx: 145150000, TxTone: "67.0", RxTone: "67.0", Power: 0, Name: "RV60"}},
				edit{free[3], &Chan{Rx: 431225000, Tx: 432825000, TxTone: "100.5", RxTone: "D754I", Power: 1, Name: "TONO LIBERO"}},
			)
			for _, e := range edits {
				if err := m.Fam.Set(m, mine, e.n, e.c); err != nil {
					t.Fatal(err)
				}
			}
			for _, e := range edits {
				g := m.Fam.Get(m, mine, e.n)
				if (g == nil) != (e.c == nil) || g != nil && (g.Rx != e.c.Rx || g.Tx != e.c.Tx || g.Name != e.c.Name || g.Dig != e.c.Dig ||
					g.TxTone != e.c.TxTone || g.RxTone != e.c.RxTone || g.TxOff != e.c.TxOff || g.Power != e.c.Power || g.Narrow != e.c.Narrow) {
					t.Errorf("canale %d riletto: %+v (scritto %+v)", e.n+1, g, e.c)
				}
			}
			_ = os.WriteFile(filepath.Join(dir, "mine.img"), mine, 0o644)
			conf := run(nil, "mine.img")
			if k := strings.Index(conf, "\nRadio:"); k >= 0 {
				conf = conf[k+1:] // via la riga «Read codeplug from file»
			}
			_ = os.WriteFile(filepath.Join(dir, "mod.conf"), []byte(conf), 0o644)
			run(nil, "-c", "base.img", "mod.conf")
			theirs, _ := os.ReadFile(filepath.Join(dir, "device.img"))
			// fuori dal confronto: il canale A e B di ogni zona (dmrconfig, creando le
			// zone, scrive quello della zona n nella posizione 2·n: errore non copiato)
			mask := func(b []byte) []byte {
				o := append([]byte{}, b...)
				for k := atOffZChA; k < atOffZChB+0x200; k++ {
					o[k] = 0
				}
				return o
			}
			if !bytes.Equal(mask(theirs), mask(mine)) {
				t.Fatalf("memoria diversa da quella scritta da dmrconfig: %s", diffRanges(mask(mine), mask(theirs)))
			}
			// nelle zone toccate dalle modifiche il canale A e B è un canale della zona
			for z := 0; z < atNZones; z++ {
				l := atList(atZoneList(mine, z))
				if !atZoneValid(mine, z) || len(l) == 0 || bytes.Equal(atZoneList(mine, z), atZoneList(base, z)) {
					continue
				}
				for _, o := range []int{atOffZChA, atOffZChB} {
					v, ok := binary.LittleEndian.Uint16(mine[o+2*z:]), false
					for _, x := range l {
						ok = ok || x == v
					}
					if !ok {
						t.Errorf("zona %d: il canale %d (A o B) non è nella zona", z+1, v+1)
					}
				}
			}
			// 4) protocollo
			old := openPortFunc
			defer func() { openPortFunc = old }()
			var cur genEmu
			openPortFunc = func(string, int) (Port, error) {
				a, b := newPipe()
				go cur.serve(b)
				return a, nil
			}
			emu := genEmuFor(m, base)
			tty, stop := servePty(t, func(p Port) { _ = emu.serve(p) })
			run([]string{"DMRCONFIG_SERIAL=" + tty}, "-r")
			stop()
			read, _ := os.ReadFile(filepath.Join(dir, "device.img"))
			cur = genEmuFor(m, base)
			got, info, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(read, got) || !bytes.Equal(got, base) {
				t.Fatalf("lettura di dmrconfig diversa: %s / %s", diffRanges(read, got), diffRanges(got, base))
			}
			// scrittura: dmrconfig su una radio vuota (scrive tutto), questo programma
			// sulla radio appena letta (scrive solo canali, zone e liste di scansione)
			eo := genEmuFor(m, atSynthImage(m)).(*atEmu)
			tty, stop = servePty(t, func(p Port) { _ = eo.serve(p) })
			run([]string{"DMRCONFIG_SERIAL=" + tty}, "-w", "mine.img")
			stop()
			ec := genEmuFor(m, base).(*atEmu)
			cur = ec
			if _, err := m.Fam.Upload(context.Background(), m, "COM7", mine, noProg); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(eo.image(), ec.image()) || !bytes.Equal(ec.image(), mine) {
				t.Fatalf("radio scritta da dmrconfig diversa: %s / %s", diffRanges(eo.image(), ec.image()), diffRanges(ec.image(), mine))
			}
			theirW := map[uint32]bool{}
			for _, a := range eo.writes {
				theirW[a] = true
			}
			for _, a := range ec.writes {
				// le aree qui scritte per intero e da dmrconfig solo in parte: nomi delle
				// zone non valide no (qui solo le valide); tutto il resto deve coincidere
				if !theirW[a] {
					t.Fatalf("indirizzo %08X scritto qui e non da dmrconfig", a)
				}
			}
			if ec.ended != 1 || eo.ended != 1 {
				t.Errorf("«END» inviati: qui %d, dmrconfig %d", ec.ended, eo.ended)
			}
			t.Logf("%s: %d canali letti uguali a dmrconfig; %d modifiche: memoria identica a quella di dmrconfig; lettura uguale; %d blocchi scritti qui (dmrconfig %d), tutti uguali",
				info, n, len(edits), len(ec.writes), len(eo.writes))
		})
	}
}

// atTestImg: memoria di prova senza dmrconfig: tre zone, due liste di scansione,
// venti canali, un contatto
func atTestImg(t *testing.T, m *genModel) []byte {
	t.Helper()
	img := atSynthImage(m)
	atSetBit(img, atOffContactMap, 0, false)
	copy(img[atOffContacts:], make([]byte, 100))
	img[atOffContacts] = 1
	copy(img[atOffContacts+1:], "Locale")
	for z := 0; z < 2; z++ {
		atSetBit(img, atOffZoneMap, z, true)
		copy(img[atOffZoneNames+z*32:], append([]byte(fmt.Sprintf("Zona %d", z+1)), make([]byte, 26)...))
	}
	for i := 0; i < 20; i++ {
		c := &Chan{Rx: 145000000 + int64(i)*12500, Tx: 145000000 + int64(i)*12500, Name: fmt.Sprintf("CANALE %d", i+1), Power: i % 3}
		if i%4 == 3 {
			c.Dig, c.Rx, c.Tx = true, 438000000+int64(i)*12500, 430400000+int64(i)*12500
		}
		if err := m.Fam.Set(m, img, i, c); err != nil {
			t.Fatal(err)
		}
	}
	// zona 2 con i canali 11-20 (la prima li ha presi tutti)
	atPutList(atZoneList(img, 0), []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	atPutList(atZoneList(img, 1), []uint16{10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
	binary.LittleEndian.PutUint16(img[atOffZChA+2:], 10)
	binary.LittleEndian.PutUint16(img[atOffZChB+2:], 11)
	for s := 0; s < 2; s++ {
		atSetBit(img, atOffScanMap, s, true)
		sl := atScanRec(img, s)
		for k := range sl {
			sl[k] = 0
		}
		copy(sl[15:], fmt.Sprintf("Lista %d", s+1))
		atPutList(sl[32:132], []uint16{uint16(s * 10), uint16(s*10 + 1), uint16(s*10 + 2)})
		sl[1] = 3
		binary.LittleEndian.PutUint16(sl[2:], uint16(s*10+1)) // canale prioritario 1: il primo della lista
		binary.LittleEndian.PutUint16(sl[4:], uint16(s*10+2))
	}
	return img
}

func atWithEmu(t *testing.T, e genEmu) {
	t.Helper()
	old := openPortFunc
	openPortFunc = func(string, int) (Port, error) {
		a, b := newPipe()
		go e.serve(b)
		return a, nil
	}
	t.Cleanup(func() { openPortFunc = old })
}

// cancellazioni e canali nuovi: zone, canale A e B, liste di scansione
func TestAnytoneZones(t *testing.T) {
	for _, id := range []string{"at-d878uv", "bt-dmr6x2pro"} {
		m := genByID[id]
		img := atTestImg(t, m)
		// cancello il canale 11: è il canale A della zona 2 e il primo della lista 2,
		// di cui è anche il canale prioritario 1
		if err := m.Fam.Set(m, img, 10, nil); err != nil {
			t.Fatal(err)
		}
		if l := atList(atZoneList(img, 1)); len(l) != 9 || l[0] != 11 {
			t.Fatalf("%s: zona 2 dopo la cancellazione: %v", id, l)
		}
		if a, b := binary.LittleEndian.Uint16(img[atOffZChA+2:]), binary.LittleEndian.Uint16(img[atOffZChB+2:]); a != 11 || b != 11 {
			t.Fatalf("%s: canale A %d e B %d della zona 2 (attesi 12 e 12)", id, a+1, b+1)
		}
		sl := atScanRec(img, 1)
		if l := atList(sl[32:132]); len(l) != 2 || l[0] != 11 || l[1] != 12 {
			t.Fatalf("%s: lista 2 dopo la cancellazione: %v", id, l)
		}
		if p1, p2 := binary.LittleEndian.Uint16(sl[2:]), binary.LittleEndian.Uint16(sl[4:]); p1 != 0xFFFF || p2 != 12 || sl[1] != 2 {
			t.Fatalf("%s: canali prioritari %04X %04X, scelta %d", id, p1, p2, sl[1])
		}
		if atValid(img, 10) || !bytes.Equal(atRec(img, 10), bytes.Repeat([]byte{0xFF}, 64)) {
			t.Fatalf("%s: il canale cancellato non è vuoto", id)
		}
		// la zona 1 resta com'era
		if l := atList(atZoneList(img, 0)); len(l) != 10 {
			t.Fatalf("%s: zona 1 cambiata: %v", id, l)
		}
		// canale nuovo: entra nella prima zona con posto
		if err := m.Fam.Set(m, img, 100, &Chan{Rx: 433000000, Tx: 433000000, Name: "NUOVO"}); err != nil {
			t.Fatal(err)
		}
		if l := atList(atZoneList(img, 0)); len(l) != 11 || l[10] != 100 {
			t.Fatalf("%s: il canale nuovo non è nella zona 1: %v", id, l)
		}
		// senza nessuna zona nasce «Zone 1»
		empty := atSynthImage(m)
		if err := m.Fam.Set(m, empty, 5, &Chan{Rx: 433000000, Tx: 433000000, Name: "PRIMO"}); err != nil {
			t.Fatal(err)
		}
		if !atZoneValid(empty, 0) || genNameFromBytes(empty[atOffZoneNames:atOffZoneNames+16], true) != "Zone 1" ||
			len(atList(atZoneList(empty, 0))) != 1 || binary.LittleEndian.Uint16(empty[atOffZChA:]) != 5 {
			t.Fatalf("%s: zona nuova non creata", id)
		}
		// zona piena (250 canali): il canale nuovo va nella successiva
		full := atSynthImage(m)
		atSetBit(full, atOffZoneMap, 0, true)
		var all []uint16
		for k := 0; k < 250; k++ {
			all = append(all, uint16(k))
			atSetBit(full, atOffChMap, k, true)
		}
		atPutList(atZoneList(full, 0), all)
		if err := m.Fam.Set(m, full, 300, &Chan{Rx: 433000000, Tx: 433000000, Name: "OLTRE"}); err != nil {
			t.Fatal(err)
		}
		if l := atList(atZoneList(full, 1)); !atZoneValid(full, 1) || len(l) != 1 || l[0] != 300 || len(atList(atZoneList(full, 0))) != 250 {
			t.Fatalf("%s: zona piena gestita male: %v", id, l)
		}
	}
}

// potenza «turbo», tono libero, modi A+D e D+A, canale DMR nuovo
func TestAnytoneChannel(t *testing.T) {
	m := genByID["at-d868uv"]
	x6 := genByID["bt-dmr6x2"]
	img := atTestImg(t, m)
	r := atRec(img, 0)
	before := append([]byte{}, r...)
	r[8] |= 3 << 2 // turbo
	c := m.Fam.Get(m, img, 0)
	if c.Power != 0 {
		t.Fatalf("turbo letto come %d", c.Power)
	}
	c.Name = "TURBO"
	if err := m.Fam.Set(m, img, 0, c); err != nil {
		t.Fatal(err)
	}
	if r[8]>>2&3 != 3 {
		t.Fatal("la potenza turbo non è stata conservata")
	}
	c.Power = 2
	_ = m.Fam.Set(m, img, 0, c)
	if r[8]>>2&3 != 0 {
		t.Fatal("potenza bassa non scritta")
	}
	c.Power = 0
	_ = m.Fam.Set(m, img, 0, c)
	if r[8]>>2&3 != 2 {
		t.Fatal("potenza alta non scritta")
	}
	// i byte non gestiti di un canale esistente non cambiano
	copy(r, before)
	r[18], r[24], r[26], r[52], r[60] = 7, 3, 0x11, 1, 0x5A
	c = m.Fam.Get(m, img, 0)
	c.Rx, c.Tx, c.TxTone, c.Narrow = 146000000, 145400000, "123.0", true
	if err := m.Fam.Set(m, img, 0, c); err != nil {
		t.Fatal(err)
	}
	if r[18] != 7 || r[24] != 3 || r[26] != 0x11 || r[52] != 1 || r[60] != 0x5A {
		t.Fatalf("byte non gestiti cambiati: % X", r)
	}
	if g := m.Fam.Get(m, img, 0); g.Rx != 146000000 || g.Tx != 145400000 || g.TxTone != "123.0" || !g.Narrow || r[8]>>6 != 2 {
		t.Fatalf("canale riletto %+v", g)
	}
	// tono libero: uno solo per canale
	if err := m.Fam.Set(m, img, 1, &Chan{Rx: 145000000, Tx: 145000000, TxTone: "100.5", RxTone: "100.5", Name: "LIBERO"}); err != nil {
		t.Fatal(err)
	}
	if g := m.Fam.Get(m, img, 1); g.TxTone != "100.5" || g.RxTone != "100.5" || atRec(img, 1)[10] != 51 || atRec(img, 1)[11] != 51 {
		t.Fatalf("tono libero riletto %+v", g)
	}
	if err := m.Fam.Set(m, img, 1, &Chan{Rx: 145000000, Tx: 145000000, TxTone: "100.5", RxTone: "101.5", Name: "DUE"}); err == nil {
		t.Fatal("due toni liberi diversi accettati")
	}
	if (atFam{}).toneOK(genTone{Mode: "Tone", Hz10: 1005}, genTone{Mode: "Tone", Hz10: 1015}) || !(atFam{}).toneOK(genTone{Mode: "Tone", Hz10: 885}, genTone{Mode: "Tone", Hz10: 1015}) {
		t.Fatal("toneOK")
	}
	// A+D resta A+D (analogico), D+A resta D+A (DMR)
	for _, md := range []byte{2, 3} {
		r := atRec(img, 2)
		r[8] = r[8]&^3 | md
		c := m.Fam.Get(m, img, 2)
		if c.Dig != (md == 3) {
			t.Fatalf("modo %d letto come DMR=%v", md, c.Dig)
		}
		c.Name = "MISTO"
		_ = m.Fam.Set(m, img, 2, c)
		if r[8]&3 != md {
			t.Fatalf("modo %d cambiato in %d", md, r[8]&3)
		}
	}
	// canale DMR nuovo: color code 1, slot 1, primo contatto, nessuna lista
	_ = m.Fam.Set(m, img, 50, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Name: "DMR"})
	r = atRec(img, 50)
	if r[8]&3 != 1 || r[32] != 1 || r[33] != 0 || binary.LittleEndian.Uint16(r[20:]) != 0 || r[27] != 0xFF || r[28] != 0xFF || r[8]>>6 != 2 {
		t.Fatalf("canale DMR nuovo: % X", r)
	}
	// DMR-6X2: le otto liste di scansione sono da 54
	i6 := atTestImg(t, x6)
	_ = x6.Fam.Set(x6, i6, 50, &Chan{Rx: 438300000, Tx: 430700000, Dig: true, Name: "DMR"})
	r = atRec(i6, 50)
	if r[27] != 0 || !bytes.Equal(r[54:62], bytes.Repeat([]byte{0xFF}, 8)) || r[62] != 0 {
		t.Fatalf("canale nuovo della DMR-6X2: % X", r)
	}
	// un canale DMR spostato in una posizione vuota conserva color code, slot e
	// contatto, diventa valido ed entra in una zona
	src := atRec(img, 50)
	src[32], src[33] = 7, 1
	binary.LittleEndian.PutUint16(src[20:], 0)
	chs := genChans(m, img)
	moved := *chs[50]
	chs[70], chs[50] = &moved, nil
	out, err := genWrite(m, img, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	if d := atRec(out, 70); !atValid(out, 70) || atValid(out, 50) || d[32] != 7 || d[33] != 1 || d[8]&3 != 1 {
		t.Fatalf("canale DMR spostato: % X", d)
	}
	inZone := false
	for z := 0; z < atNZones; z++ {
		for _, v := range atList(atZoneList(out, z)) {
			inZone = inZone || (atZoneValid(out, z) && v == 70)
			if v == 50 {
				t.Fatal("il canale spostato è rimasto in una zona con il vecchio numero")
			}
		}
	}
	if !inZone {
		t.Fatal("il canale spostato non è in nessuna zona")
	}
	// frequenze non memorizzabili
	if err := m.Fam.Set(m, img, 60, &Chan{Rx: 145000005, Tx: 145000005, Name: "X"}); err == nil {
		t.Fatal("frequenza con passo di 5 Hz accettata")
	}
	if keep, _ := (atFam{}).fit(m, &Chan{Rx: 145000005, Tx: 145000005}); keep {
		t.Fatal("fit")
	}
	// riconoscimento del file: il modello è in testa alla memoria
	for _, o := range []string{"at-d868uv", "at-d878uv", "at-d878uv2", "bt-dmr6x2", "bt-dmr6x2pro"} {
		om := genByID[o]
		b := atSynthImage(om)
		for _, p := range []string{"at-d868uv", "at-d878uv", "at-d878uv2", "bt-dmr6x2", "bt-dmr6x2pro"} {
			if genByID[p].Fam.Match(genByID[p], b) != (o == p) {
				t.Fatalf("la memoria della %s è presa per una %s", o, p)
			}
		}
	}
}

// protocollo: somma di controllo, modello diverso, scrittura rifiutata, aree scritte
func TestAnytoneProtocol(t *testing.T) {
	m := genByID["at-d878uv2"]
	base := atTestImg(t, m)
	e := genEmuFor(m, base).(*atEmu)
	e.badSum = 2
	atWithEmu(t, e)
	got, info, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
	if err != nil || !bytes.Equal(got, base) {
		t.Fatalf("lettura con due somme sbagliate: %v", err)
	}
	if !strings.Contains(info, "D878UV2") || e.ended != 1 {
		t.Fatalf("informazioni %q, END %d", info, e.ended)
	}
	e.badSum = 3
	if _, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg); err == nil {
		t.Fatal("tre somme sbagliate di fila accettate")
	}
	e.badSum = 0
	// la radio di un altro modello è rifiutata prima di leggere o scrivere
	other := genByID["at-d878uv"]
	w0 := len(e.writes)
	if _, _, err := other.Fam.Download(context.Background(), other, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "D878UV2") {
		t.Fatalf("modello diverso: %v", err)
	}
	oi := atTestImg(t, other)
	if _, err := other.Fam.Upload(context.Background(), other, "COM7", oi, noProg); err == nil || len(e.writes) != w0 {
		t.Fatalf("scrittura su un modello diverso: %v, %d blocchi scritti", err, len(e.writes)-w0)
	}
	// una memoria di un altro modello non si scrive
	if _, err := m.Fam.Upload(context.Background(), m, "COM7", oi, noProg); err == nil || len(e.writes) != w0 {
		t.Fatalf("memoria di un altro modello: %v", err)
	}
	// modifiche e scrittura: solo canali, zone, liste di scansione e tabelle
	chs := genChans(m, got)
	chs[3] = nil
	chs[30] = &Chan{Rx: 433500000, Tx: 433500000, Name: "NUOVO", TxTone: "88.5"}
	nc := *chs[5]
	nc.Name = "RINOMINATO"
	chs[5] = &nc
	want, err := genWrite(m, got, chs, &Report{})
	if err != nil {
		t.Fatal(err)
	}
	// la radio ha anche dati che il programma non deve toccare
	e.mu.Lock()
	copy(e.mem[atOffSettings+0x600:], "BENVENUTO")
	copy(want[atOffSettings+0x600:], "BENVENUTO")
	e.mu.Unlock()
	if _, err := m.Fam.Upload(context.Background(), m, "COM7", want, noProg); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(e.image(), want) {
		t.Fatalf("radio diversa dalla memoria scritta: %s", diffRanges(e.image(), want))
	}
	for _, a := range e.writes[w0:] {
		ok := a >= atAddrCh && a < 0x01000000 || // canali
			a >= atAddrZoneLists && a < atAddrZoneLists+250*512 ||
			a >= atAddrScan && a < 0x01640000 ||
			a >= atAddrZoneMap && a < atAddrZoneMap+0x60 ||
			a >= atAddrChMap && a < atAddrChMap+0x200 ||
			a >= atAddrZChA && a < atAddrZChB+0x200 ||
			a >= atAddrZoneNames && a < atAddrZoneNames+250*32
		if !ok {
			t.Fatalf("scritto l'indirizzo %08X, fuori da canali, zone e liste", a)
		}
		if a%16 != 0 {
			t.Fatalf("indirizzo %08X non allineato a 16 byte", a)
		}
	}
	// il canale cancellato non viene riscritto
	for _, a := range e.writes[w0:] {
		if a >= atAddrCh+3*64 && a < atAddrCh+4*64 {
			t.Fatal("scritto il canale cancellato")
		}
	}
	// scrittura rifiutata dalla radio: errore
	e.nak = atAddrChMap
	if _, err := m.Fam.Upload(context.Background(), m, "COM7", want, noProg); err == nil {
		t.Fatal("scrittura rifiutata non segnalata")
	}
	e.nak = 0
	// contatti: un contatto valido dopo uno vuoto si legge per intero
	e2 := genEmuFor(m, base).(*atEmu)
	atSetBit(e2.mem, atOffContactMap, 2, false)
	for k := 0; k < 100; k++ {
		e2.mem[atOffContacts+200+k] = byte(k + 1)
	}
	atWithEmu(t, e2)
	got2, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg)
	if err != nil {
		t.Fatal(err)
	}
	for k := 0; k < 100; k++ {
		if got2[atOffContacts+200+k] != byte(k+1) {
			t.Fatalf("contatto 3 letto male al byte %d", k)
		}
	}
	// la memoria che dichiara un altro modello è rifiutata
	e3 := genEmuFor(m, base).(*atEmu)
	copy(e3.mem, "D999XYZ")
	atWithEmu(t, e3)
	if _, _, err := m.Fam.Download(context.Background(), m, "COM7", noProg); err == nil || !strings.Contains(err.Error(), "D999XYZ") {
		t.Fatalf("memoria di un altro modello: %v", err)
	}
	// annullamento
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	atWithEmu(t, genEmuFor(m, base))
	if _, _, err := m.Fam.Download(ctx, m, "COM7", noProg); err == nil {
		t.Fatal("lettura annullata senza errore")
	}
}
