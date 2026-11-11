//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func init() { kwlPause = 0 }

func kwlModel(t *testing.T, id string) *genModel {
	for _, m := range genModels {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("modello %s non registrato", id)
	return nil
}

// i canali si scrivono e si rileggono uguali; il testo ha il formato dei driver di CHIRP
func TestKwliveChannels(t *testing.T) {
	want := map[string]string{
		// TH-F7: canale 5, 145,600 -600 kHz con tono 88,5 Hz in trasmissione
		"kw-thf7": "0,005,00145600000,0,2,0,1,0,0,08,08,000,000600000,0,0",
		// TH-G71 e TH-D7: i codici dei toni partono da 1 e i campi del DCS sono vuoti
		"kw-thg71": "0,0,005,00145600000,0,2,0,1,0,,09,,09,000600000,0",
		"kw-thd7":  "0,0,005,00145600000,0,2,0,1,0,,09,,09,000600000,0,0",
		"kw-thk2":  "05,0145600000,0,2,0,1,0,0,08,08,000,00600000,0,0,0000000000,0,0",
	}
	for id, spec := range want {
		m := kwlModel(t, id)
		f := m.Fam.(kwlFam)
		img := kwlBlank(m)
		c := &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true, Name: "R0"}
		if err := f.Set(m, img, 5, c); err != nil {
			t.Fatal(id, err)
		}
		if got := kwlText(f.rec(img, 5)[kwlSpecAt:]); got != spec {
			t.Errorf("%s: testo del canale\n  %s\nvoluto\n  %s", id, got, spec)
		}
		if back := f.Get(m, img, 5); back == nil || *back != *c {
			t.Errorf("%s: riletto %+v, scritto %+v", id, back, c)
		}
		// subtono nei due versi, escluso dalla scansione, 12,5 kHz
		c = &Chan{Rx: 145612500, Tx: 145612500, TxTone: "123.0", RxTone: "123.0", Name: "TSQL"}
		if err := f.Set(m, img, 6, c); err != nil {
			t.Fatal(id, err)
		}
		if back := f.Get(m, img, 6); back == nil || *back != *c {
			t.Errorf("%s: riletto %+v, scritto %+v", id, back, c)
		}
		// combinazioni che la radio non ha
		for _, bad := range []*Chan{
			{Rx: 145500000, Tx: 145500000, RxTone: "88.5"},
			{Rx: 145500000, Tx: 145500000, TxTone: "88.5", RxTone: "100.0"},
			{Rx: 145500000, Tx: 145500000, TxTone: "D023I", RxTone: "D023I"},
			{Rx: 145501000, Tx: 145501000},
		} {
			if err := f.Set(m, img, 7, bad); err == nil {
				t.Errorf("%s: accettato %+v", id, bad)
			}
		}
		o := kwlo(m)
		dcs := &Chan{Rx: 145500000, Tx: 145500000, TxTone: "D754N", RxTone: "D754N", Scan: true}
		if err := f.Set(m, img, 8, dcs); (err == nil) != (len(m.DCS) > 0) {
			t.Errorf("%s: DCS: %v", id, err)
		} else if err == nil {
			if back := f.Get(m, img, 8); back == nil || *back != *dcs {
				t.Errorf("%s: riletto %+v, scritto %+v", id, back, dcs)
			}
		}
		// frequenza di trasmissione separata (o rifiuto, sulla TH-K2)
		sp := &Chan{Rx: 145500000, Tx: 433500000, Scan: true, Name: "SPLIT"}
		err := f.Set(m, img, 9, sp)
		if (err == nil) != o.split {
			t.Errorf("%s: frequenza separata: %v", id, err)
		} else if err == nil {
			if back := f.Get(m, img, 9); back == nil || *back != *sp {
				t.Errorf("%s: riletto %+v, scritto %+v", id, back, sp)
			}
			if r := f.rec(img, 9); r[1] != 1 || kwlText(r[kwlSplitAt:]) != "00433500000" {
				t.Errorf("%s: frequenza separata non memorizzata", id)
			}
		}
		// un canale spostato prende il numero della nuova posizione
		raw := f.Rec(m, img, 5)
		f.PutRec(m, img, 20, raw)
		moved := f.Get(m, img, 20)
		moved.Name = "MOVED"
		if err := f.Set(m, img, 20, moved); err != nil {
			t.Fatal(id, err)
		}
		ix := kwlIdxs[o.kind]
		if fl := kwlFields(m, f.rec(img, 20)); fl == nil || kwlInt(fl[ix.num]) != 20 {
			t.Errorf("%s: numero del canale spostato: %v", id, fl)
		}
		// cancellazione
		if err := f.Set(m, img, 5, nil); err != nil || f.Get(m, img, 5) != nil || !bytes.Equal(f.rec(img, 5), make([]byte, kwlRec)) {
			t.Errorf("%s: canale non cancellato", id)
		}
		if !f.Match(m, img) {
			t.Errorf("%s: immagine non riconosciuta", id)
		}
	}
}

// i campi che il programma non cambia restano scritti come li ha dati la radio;
// passo e modo restano finché la frequenza non cambia
func TestKwliveKeepsRadioText(t *testing.T) {
	m := kwlModel(t, "kw-thf7")
	f := m.Fam.(kwlFam)
	img := kwlBlank(m)
	r := f.rec(img, 3)
	r[0] = 1
	// 7,100 MHz in LSB (modo 3), passo 5 kHz, numero di tono fuori formato
	copy(r[kwlSpecAt:], "0,003,00007100000,0,0,0,0,0,0,8,8,0,000000000,3,0")
	copy(r[kwlNameAt:], "40M")
	c := f.Get(m, img, 3)
	if c == nil || c.Rx != 7100000 || c.Name != "40M" || c.TxTone != "" {
		t.Fatalf("letto %+v", c)
	}
	c.Name = "QRP"
	c.Scan = false
	if err := f.Set(m, img, 3, c); err != nil {
		t.Fatal(err)
	}
	if got, want := kwlText(r[kwlSpecAt:]), "0,003,00007100000,0,0,0,0,0,0,8,8,0,000000000,3,1"; got != want {
		t.Errorf("testo\n  %s\nvoluto\n  %s", got, want)
	}
	// la frequenza cambia: il modo segue la frequenza (FM), il passo resta se va bene
	c.Rx, c.Tx = 145500000, 145500000
	if err := f.Set(m, img, 3, c); err != nil {
		t.Fatal(err)
	}
	if got, want := kwlText(r[kwlSpecAt:]), "0,003,00145500000,0,0,0,0,0,0,8,8,0,000000000,0,1"; got != want {
		t.Errorf("testo\n  %s\nvoluto\n  %s", got, want)
	}
	// banda aerea a 8,33 kHz: passo 2 e AM
	c.Rx, c.Tx = 118008333, 118008333
	if err := f.Set(m, img, 3, c); err == nil {
		t.Errorf("accettata una frequenza che non è multipla di nessun passo")
	}
	c.Rx, c.Tx = 118025000, 118025000
	if err := f.Set(m, img, 3, c); err != nil {
		t.Fatal(err)
	}
	if got, want := kwlText(r[kwlSpecAt:]), "0,003,00118025000,0,0,0,0,0,0,8,8,0,000000000,2,1"; got != want {
		t.Errorf("testo\n  %s\nvoluto\n  %s", got, want)
	}
}

// in scrittura si inviano solo i canali cambiati; modello sbagliato e radio muta
func TestKwliveUpload(t *testing.T) {
	for _, id := range []string{"kw-thf6", "kw-thg71", "kw-thd7g", "kw-thk2"} {
		m := kwlModel(t, id)
		f := m.Fam.(kwlFam)
		base := kwlBlank(m)
		for i := 0; i < 12; i++ {
			c := &Chan{Rx: 145000000 + int64(i)*12500, Tx: 145000000 + int64(i)*12500, Scan: true, Name: "CH" + string(rune('A'+i))}
			if i == 4 && kwlo(m).split {
				c.Tx = 434000000
			}
			if err := f.Set(m, base, i, c); err != nil {
				t.Fatal(err)
			}
		}
		emu := f.emu(m, base).(*kwlEmu)
		old := openPortFunc
		openPortFunc = func(name string, baud int) (Port, error) {
			a, b := newPipe()
			a.baud = baud
			go emu.serve(b)
			return a, nil
		}
		got, _, err := f.Download(context.Background(), m, "COM7", noProg)
		if err != nil {
			t.Fatal(id, err)
		}
		if !bytes.Equal(got, base) {
			t.Fatalf("%s: lettura diversa dalla memoria della radio", id)
		}
		// la radio che risponde «MNA 0,nnn,nome» si legge uguale
		emu.zero = true
		if got2, _, err := f.Download(context.Background(), m, "COM7", noProg); err != nil || !bytes.Equal(got2, base) {
			t.Fatalf("%s: lettura con il nome in tre campi: %v", id, err)
		}
		emu.zero = false
		// nessuna modifica: nessuna scrittura
		if _, err := f.Upload(context.Background(), m, "COM7", got, noProg); err != nil || len(emu.log) != 0 {
			t.Fatalf("%s: scrittura senza modifiche: %v, %d comandi", id, err, len(emu.log))
		}
		want := append([]byte{}, got...)
		f.Set(m, want, 1, nil)                                                                         // cancellato
		f.Set(m, want, 2, &Chan{Rx: 145600000, Tx: 145000000, TxTone: "88.5", Scan: true, Name: "R0"}) // cambiato
		f.Set(m, want, 4, &Chan{Rx: 145050000, Tx: 145050000, Scan: true, Name: "CHE"})                // non più separato
		f.Set(m, want, 30, &Chan{Rx: 145500000, Tx: 145500000, Scan: true, Name: "NEW"})               // nuovo
		if kwlo(m).split {
			f.Set(m, want, 31, &Chan{Rx: 145500000, Tx: 430000000, Scan: true, Name: "SPL"})
		}
		if _, err := f.Upload(context.Background(), m, "COM7", want, noProg); err != nil {
			t.Fatal(id, err)
		}
		if !bytes.Equal(emu.image(), want) {
			t.Fatalf("%s: memoria della radio diversa da quella scritta", id)
		}
		for _, l := range emu.log {
			for _, n := range []string{"000", "003", "005", "011"} {
				if kwlo(m).kind == kwlK2 {
					n = n[1:]
				}
				if strings.Contains(l, " "+n+",") || strings.Contains(l, ","+n+",") && !strings.Contains(l, "00"+n) {
					t.Errorf("%s: scritto un canale non cambiato: %s", id, l)
				}
			}
		}
		// un'altra radio della stessa famiglia
		for _, x := range genModels {
			if xo, ok := x.Opt.(*kwlOpt); ok && x.ID != id && xo.id != kwlo(m).id {
				if _, _, err := x.Fam.Download(context.Background(), x, "COM7", noProg); err == nil || !strings.Contains(err.Error(), m.Name) {
					t.Errorf("%s letta come %s: %v", id, x.ID, err)
				}
				break
			}
		}
		if id == "kw-thd7g" {
			emu.mute = true
			if _, _, err := f.Download(context.Background(), m, "COM7", noProg); err == nil {
				t.Errorf("%s: radio muta letta", id)
			}
		}
		openPortFunc = old
	}
}
