//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

// 4.17: prove dei palmari Kenwood TK delle serie 60 e 60G. CHIRP ha un'immagine
// di prova per serie (TK-278 e TK-272G): per gli altri modelli si usa la stessa
// memoria con la sigla della prima versione del modello.
func init() {
	chirpFamilyImage["main.kwp60Fam"] = "Kenwood_TK-278.img"
	chirpFamilyImage["main.kw60gFam"] = "Kenwood_TK-272G.img"
	for _, m := range genModels {
		if _, ok := m.Opt.(*kwtkOpt); !ok || m.ID == "kw-tk278" || m.ID == "kw-tk272g" {
			continue
		}
		id := kwtkTestVar(m).id
		if _, p60 := m.Fam.(kwp60Fam); p60 {
			chirpImagePatch[m.ID] = func(b []byte) {
				for k := 0; k < 8; k++ {
					b[kwp60ID+7-k] = id[k]
				}
			}
		} else {
			chirpImagePatch[m.ID] = func(b []byte) { copy(b[kw60gID:], id) }
		}
	}
}

// CHIRP, modificando un canale già presente, ne azzera gruppo e posizione (le
// impostazioni «extra» del canale vengono riscritte tutte con un valore logico)
// e non ricostruisce le tabelle dopo una cancellazione: qui gruppo e posizione
// restano quelli del canale e le tabelle si ricostruiscono sempre.
func (f kw60gFam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	for i := 0; i < f.count(m, base); i++ {
		r := f.rec(i)
		if f.Get(m, ref, i) == nil {
			continue
		}
		r.F("bnumb").SetInt(ref, r.F("bnumb").Int(out))
		r.F("bank").SetInt(ref, r.F("bank").Int(out))
	}
	f.prep(m, ref)
}

// versione usata nelle prove: la prima con tutti i canali del modello
func kwtkTestVar(m *genModel) kwVar {
	o := m.Opt.(*kwtkOpt)
	for _, v := range o.vars {
		if v.n == m.Slots {
			return v
		}
	}
	return o.vars[0]
}

// modifiche che CHIRP non ammette: frequenze fuori dai limiti della versione
func kwtkOut(m *genModel, c *Chan) bool {
	b := kwtkBand(kwtkTestVar(m))
	if d := c.Tx - c.Rx; d > 70000000 || d < -70000000 {
		return true // per CHIRP è «split», che queste radio non hanno
	}
	return c.Rx < b.Lo || c.Rx > b.Hi || !c.TxOff && (c.Tx < b.Lo || c.Tx > b.Hi)
}
func (kwp60Fam) chirpInvalid(m *genModel, c *Chan) bool { return kwtkOut(m, c) }
func (kw60gFam) chirpInvalid(m *genModel, c *Chan) bool { return kwtkOut(m, c) }

// CHIRP, modificando un canale già presente (anche uno appena creato), rovescia
// «blocco del canale occupato» e «beat shift» (li riscrive con il valore
// mostrato, che è il contrario del bit): qui restano com'erano.
func (f kwp60Fam) chirpEncodeFix(m *genModel, base, ref, out []byte) {
	for i := 0; i < m.Slots; i++ {
		if f.Get(m, ref, i) == nil {
			continue
		}
		a := f.rec(i).Off + 12
		ref[a] = ref[a]&^0x18 | out[a]&0x18
	}
}
