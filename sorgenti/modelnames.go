// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Nomi esatti delle radio (4.6). Più radio della galleria possono avere la stessa
// gestione (stesso driver, stesso collegamento e stessa memoria): per esempio la
// Radtel RT-950 Pro e la Bajeton BJ-9000, la Baofeng K67 Pro e la 5RH Pro, la
// Radtel RT-910 e la Bajeton BJ-7700. Via cavo queste radio rispondono allo stesso
// modo e la memoria non dice quale sia, ma restano radio diverse: la scheda, la
// lettura, la scrittura e l'identificazione devono chiamarle con il loro nome.
//   - la scheda aperta dalla galleria porta il nome della radio scelta;
//   - l'identificazione propone tutte le radio con la stessa gestione e usa quella
//     scelta l'ultima volta (Settings.ModelCards); se non c'è una scelta
//     precedente e le radio sono più d'una lo dice e chiede quale sia;
//   - la scelta si ricorda per le volte successive.

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

// cardsForKind: radio della galleria gestite con la scheda kind (nell'ordine del catalogo)
func cardsForKind(kind string) []catItem {
	var out []catItem
	for _, c := range catalog {
		for _, d := range c.Drivers {
			if d == kind {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func cardByID(id string) (catItem, bool) {
	for _, c := range catalog {
		if c.ID == id {
			return c, true
		}
	}
	return catItem{}, false
}

// nameQualifier: precisazione tra parentesi alla fine del nome di un modello
// (es. «firmware 0.24 e successivi: 990 canali», «Bluetooth», «PMR446»)
func nameQualifier(n string) string {
	n = strings.TrimSpace(n)
	if !strings.HasSuffix(n, ")") {
		return ""
	}
	i := strings.LastIndex(n, "(")
	if i < 0 {
		return ""
	}
	return n[i+1 : len(n)-1]
}

// kindBaseName: nome della gestione (driver) kind
func kindBaseName(kind string) string {
	switch kind {
	case kindXLC:
		return "Baofeng K67 Pro"
	case kindDAT:
		return "Bajeton BJ-8300"
	}
	if m, ok := genByID[kind]; ok {
		return m.Name
	}
	return kind
}

// cardLabel: nome della radio c gestita con la scheda kind; se la radio ha più
// versioni (più driver) si aggiunge quella di kind
func cardLabel(c catItem, kind string) string {
	name := strings.TrimSpace(c.Brand + " " + c.Name)
	if len(c.Drivers) > 1 {
		if q := nameQualifier(kindBaseName(kind)); q != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(q)) {
			name += " (" + q + ")"
		}
	}
	return name
}

func normName(s string) string {
	return strings.NewReplacer(" ", "", "-", "", "/", "", ".", "").Replace(strings.ToLower(s))
}

// kindCard: radio della galleria per la scheda kind: quella scelta l'ultima volta
// (ok = true se c'è una scelta, o se la radio con questa gestione è una sola)
func (a *App) kindCard(kind string) (catItem, bool) {
	cards := cardsForKind(kind)
	if id := a.store.Get().ModelCards[kind]; id != "" {
		for _, c := range cards {
			if c.ID == id {
				return c, true
			}
		}
	}
	if len(cards) == 1 {
		return cards[0], true
	}
	// nessuna scelta: la radio che ha il nome della gestione, altrimenti la prima
	base := normName(kindBaseName(kind))
	for _, c := range cards {
		if n := normName(c.Brand + c.Name); n != "" && strings.HasPrefix(base, n) {
			return c, false
		}
	}
	if len(cards) > 0 {
		return cards[0], false
	}
	return catItem{}, false
}

// rememberCard: la radio scelta per la scheda kind
func (a *App) rememberCard(kind, card string) {
	if kind == "" || card == "" {
		return
	}
	ok := false
	for _, c := range cardsForKind(kind) {
		ok = ok || c.ID == card
	}
	if !ok {
		return
	}
	rememberMu.Lock()
	defer rememberMu.Unlock()
	st := a.store.Get()
	if st.ModelCards[kind] == card {
		return
	}
	mc := map[string]string{}
	for k, v := range st.ModelCards {
		mc[k] = v
	}
	mc[kind] = card
	st.ModelCards = mc
	_ = a.store.SaveSettings(st)
}

// identCardNames (4.6): nomi esatti nell'esito dell'identificazione. Il modello
// trovato prende il nome della radio scelta l'ultima volta per quella gestione;
// le altre radio con la stessa gestione (e quelle delle altre gestioni proposte)
// diventano scelte «kind#scheda». Senza una scelta precedente, con più radio
// possibili, l'esito è ambiguo: l'interfaccia chiede quale radio è collegata.
func (a *App) identCardNames(r *identResult) {
	if r == nil || !r.Found {
		return
	}
	cards := cardsForKind(r.Kind)
	if len(cards) == 0 {
		return
	}
	c, sure := a.kindCard(r.Kind)
	old := r.Name
	r.Name, r.Brand, r.Model = cardLabel(c, r.Kind), c.Brand, c.Name
	r.Card = c.ID
	for i := range r.Fields {
		if r.Fields[i].L == "Marca e modello" {
			r.Fields[i].V = r.Name
		}
	}
	var others []identOther
	seen := map[string]bool{r.Kind + "#" + c.ID: true}
	add := func(kind string, cs []catItem, fallback string) {
		if len(cs) == 0 {
			if !seen[kind] {
				seen[kind] = true
				others = append(others, identOther{kind, fallback})
			}
			return
		}
		for _, o := range cs {
			k := kind + "#" + o.ID
			if !seen[k] {
				seen[k] = true
				others = append(others, identOther{k, cardLabel(o, kind)})
			}
		}
	}
	add(r.Kind, cards, old)
	for _, o := range r.Others {
		if strings.Contains(o.Kind, "#") {
			continue
		}
		add(o.Kind, cardsForKind(o.Kind), o.Name)
	}
	sort.SliceStable(others, func(i, j int) bool { return others[i].Name < others[j].Name })
	r.Others = others
	if len(others) > 0 {
		var names []string
		for _, o := range others {
			names = append(names, o.Name)
		}
		if len(names) > 12 {
			names = append(names[:12], "…")
		}
		r.add("Stessa gestione di", "%s", strings.Join(names, ", "))
	}
	if !sure && len(cards) > 1 {
		r.Ambiguous = true
	}
}

// splitKindCard: «kind#scheda» → kind, scheda
func splitKindCard(s string) (string, string) {
	if i := strings.Index(s, "#"); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// radio/card: la radio della galleria scelta per una scheda (nome mostrato)
func (a *App) apiRadioCard(r *http.Request) (interface{}, error) {
	var req struct {
		Kind string `json:"kind"`
		Card string `json:"card"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if _, ok := cardByID(req.Card); !ok {
		return nil, errors.New("radio sconosciuta")
	}
	a.rememberCard(req.Kind, req.Card)
	return map[string]interface{}{"ok": true}, nil
}
