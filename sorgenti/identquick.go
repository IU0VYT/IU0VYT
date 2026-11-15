// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// 4.40: controllo breve con la scheda già aperta, per le altre radio.
//
// Nella ricerca automatica, quando la scheda di una radio è già aperta, non
// serve rileggere tutta la memoria per dire «Radio già identificata!! Scheda
// Radio già aperta»: basta sapere che sul cavo c'è ancora quella radio. La
// 4.36 lo faceva solo per la Baofeng K67 Pro (k67Check).
//
// REGOLA: il controllo breve si dà solo alle famiglie in cui è fatto di passi
// che il protocollo prevede già così e che finiscono con l'uscita prevista dal
// protocollo. Per tutte le altre resta la lettura intera: un saluto lasciato a
// metà, su una radio vera, può lasciarla in programmazione, e da qui non si
// può provare. I casi ammessi sono tre:
//
//  1. una domanda che la radio riceve in funzionamento normale e che non la
//     mette in programmazione: Icom (richiesta del modello, la stessa con cui
//     CHIRP riconosce da solo le Icom), Kenwood «dal vivo» (comando ID),
//     Quansheng UV-K5 e derivate (saluto con la versione del firmware, lo
//     stesso che CHIRP usa per riconoscerle);
//  2. ingresso in programmazione, lettura del primo blocco e comando di
//     uscita, cioè l'inizio e la fine di ogni lettura normale: Bajeton
//     BJ-9000 / Radtel RT-900 / RT-470 e famiglia (rt900Fam, jc8810Fam) e
//     Bajeton BJ-8300. È la forma del controllo della K67 Pro, provato da
//     Daniele sulla radio vera, ed è anche quello che succede quando una
//     lettura viene annullata dopo il primo blocco;
//  3. la K67 Pro (k67Check, 4.36).
//
// Non ammessi, fra gli altri: UV-5R e derivate, BF-888S e famiglia H777,
// UV-17 Pro, TD-H8: nel loro protocollo l'identificazione è il primo passo di
// una lettura e non è noto che cosa fa la radio se la lettura non segue.
//
// Sicurezze: (a) il controllo breve si usa solo nella ricerca automatica e solo
// se la scheda di QUEL modello è aperta; (b) se la radio risponde ma il
// controllo non la conferma (altro modello, risposta inattesa) per quel
// modello il controllo breve non si usa più fino al riavvio del programma e
// la ricerca prosegue con la lettura normale; (c) le attese sono quelle della
// lettura normale, niente è accorciato.

import (
	"bytes"
	"context"
	"sync"
	"time"
)

// genQuicker: famiglia con il controllo breve
type genQuicker interface {
	quick(ctx context.Context, m *genModel, port string) error
}

var identQuick = struct {
	sync.Mutex
	off map[string]bool
}{off: map[string]bool{}}

// identQuickOff: per questo tipo di scheda il controllo breve non si usa più
// (fino al riavvio del programma)
func identQuickOff(kind string) {
	identQuick.Lock()
	identQuick.off[kind] = true
	identQuick.Unlock()
}

func identQuickAllowed(kind string) bool {
	identQuick.Lock()
	defer identQuick.Unlock()
	return !identQuick.off[kind]
}

func identQuickReset() {
	identQuick.Lock()
	identQuick.off = map[string]bool{}
	identQuick.Unlock()
}

// identQuickModel: il modello del gruppo che ha la scheda aperta e una
// famiglia con il controllo breve (nil se non c'è)
func (a *App) identQuickModel(j *cableJob, kinds []*genModel) *genModel {
	if !j.Auto {
		return nil
	}
	for _, m := range kinds {
		if _, ok := m.Fam.(genQuicker); ok && identQuickAllowed(m.ID) && a.autoTabOpen(m.ID) {
			return m
		}
	}
	return nil
}

// ---------------------------------------------------------------- Icom

// icomQuick: richiesta del modello (la radio resta in funzionamento normale)
func icomQuick(ctx context.Context, port string, baud int, o *icomOpt) error {
	rs, err := rawOpen(ctx, port, baud)
	if err != nil {
		return err
	}
	defer rs.p.Close()
	s := &icomSess{rawSess: rs}
	if err := s.ident(o); err != nil {
		return genFail("Identificazione", err)
	}
	return nil
}

func icomQuickOf(ctx context.Context, m *genModel, port string) error {
	o, ok := icomOptOf(m)
	if !ok {
		return rErr("controllo breve non disponibile")
	}
	return icomQuick(ctx, port, m.Baud, o)
}

func (id31Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (id880Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icw32Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (ict7Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icq7Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icp7Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icv86Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (ict10Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icv80Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (ict70Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icx8xFam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}
func (icx90Fam) quick(ctx context.Context, m *genModel, port string) error {
	return icomQuickOf(ctx, m, port)
}

// ---------------------------------------------------------------- Kenwood «dal vivo»

// quick: comando ID (kwlOpen controlla che il modello sia quello della scheda)
func (kwlFam) quick(ctx context.Context, m *genModel, port string) error {
	s, err := kwlOpen(ctx, m, port)
	if err != nil {
		return err
	}
	return s.p.Close()
}

// ---------------------------------------------------------------- Quansheng UV-K5

// quick: saluto con la versione del firmware (start controlla che sia fra
// quelle della scheda)
func (uvk5Fam) quick(ctx context.Context, m *genModel, port string) error {
	p, err := openPort(port, 38400)
	if err != nil {
		return err
	}
	defer p.Close()
	s := &uvk5Sess{ctx: ctx, p: p}
	_, err = s.start(m)
	return err
}

func (uvk5xFam) quick(ctx context.Context, m *genModel, port string) error {
	p, _, _, err := uvk5xFam{}.connect(ctx, m, port)
	if err != nil {
		return err
	}
	return p.Close()
}

func (uvkxFam) quick(ctx context.Context, m *genModel, port string) error {
	return uvk5xFam{}.quick(ctx, m, port)
}

// ---------------------------------------------------------------- BJ-9000 e famiglia

// rt900Quick: ingresso in programmazione, primo blocco, uscita
func rt900Quick(ctx context.Context, m *genModel, port string) error {
	o := rt900o(m)
	s, err := rt900Open(ctx, m, port)
	if err != nil {
		return err
	}
	defer s.p.Close()
	if err := s.enter(); err != nil {
		s.exit()
		return genFail("Identificazione", err)
	}
	a, bs := 0, o.blk()
	if len(o.reads) > 0 {
		a = o.reads[0][0]
	}
	cmd := []byte{'R', byte(a >> 8), byte(a), byte(bs)}
	if err := s.p.Write(cmd); err != nil {
		s.exit()
		return err
	}
	d, err := s.recv(4 + bs)
	if err == nil && !bytes.Equal(d[:4], cmd) {
		err = rErr("risposta inattesa")
	}
	s.exit()
	if err != nil {
		return genFail("Lettura del primo blocco", err)
	}
	return nil
}

func (rt900Fam) quick(ctx context.Context, m *genModel, port string) error {
	return rt900Quick(ctx, m, port)
}
func (jc8810Fam) quick(ctx context.Context, m *genModel, port string) error {
	return rt900Quick(ctx, m, port)
}

// ---------------------------------------------------------------- BJ-8300

// btCheck: collegamento, primo blocco, chiusura (come k67Check per la K67 Pro)
func btCheck(ctx context.Context, port string, prog Progress) (btInfo, error) {
	s, err := btOpenSession(ctx, port)
	if err != nil {
		return btInfo{}, err
	}
	defer s.p.Close()
	prog(0, 1, "Collegamento con la radio…")
	if err := s.handShake(); err != nil {
		return s.info, btFailMsg("Collegamento", err)
	}
	addr := btBlocks()[0]
	req := []byte{'R', byte(addr >> 8), byte(addr), 0x40}
	if err := s.p.Write(req); err != nil {
		return s.info, btFailMsg("Lettura", err)
	}
	buf := make([]byte, 68)
	err = s.p.ReadFull(s.ctx, buf, time.Second)
	if err == nil && (buf[1] != req[1] || buf[2] != req[2]) {
		err = rErr("risposta inattesa")
	}
	// la chiusura si manda comunque: la radio è entrata in programmazione
	if e := s.p.Write([]byte{'E'}); e != nil && err == nil {
		err = e
	}
	if err != nil {
		return s.info, btFailMsg("Lettura del primo blocco", err)
	}
	prog(1, 1, "Lettura completata")
	return s.info, nil
}
