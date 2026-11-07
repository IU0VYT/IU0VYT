'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Avvio dell'interfaccia, schede (anche quelle delle radio aperte dalla
// galleria dei modelli), tasti globali, chiusura.

const Views = {};
let currentTab = 'home';
const FIXED_TABS = ['home', 'gal', 'ident', 'conv', 'arch', 'info'];

Object.assign(App, {
  catalog: [], defs: {},

  showTab(t) {
    if (!FIXED_TABS.includes(t) && !Views[t]) t = 'gal';
    if (Views[t]) { App.radioTab(t); App.lastRadio = t; }
    currentTab = t;
    $$('.tab').forEach(b => b.classList.toggle('active', b.dataset.tab === t));
    $$('.page').forEach(p => p.classList.toggle('active', p.id === 'page-' + t));
    try { sessionStorage.setItem('tab', t); } catch (e) { /* facoltativo */ }
    if (t === 'home') { Home.render($('#page-home')); return; }
    if (t === 'info') { Info.render($('#page-info')); return; }
    if (t === 'gal') { Gallery.render(); return; }
    if (t === 'conv') { Converter.render($('#page-conv')); return; }
    if (t === 'ident') { Identify.render($('#page-ident')); return; }
    const g = t === 'arch' ? Archive.grid : Views[t].grid;
    if (t === 'arch') Archive.render(); else Views[t].render();
    const btn = $(`.tab[data-tab="${t}"]`);
    if (btn) btn.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    setTimeout(() => g.wrap.focus({ preventScroll: true }), 0);
  },

  // definizione della radio "kind" (radio principali o driver generici)
  radioDef(kind) {
    if (RADIO_DEF[kind]) return RADIO_DEF[kind];
    const g = App.defs[kind];
    if (!g) return null;
    // 4.8: prima la scheda della galleria in cui la radio è il driver principale
    // (es. RT-910 senza Bluetooth: firmware .BTU, non quello .anc della BJ-8200)
    const item = App.catalog.find(c => (c.drivers || [])[0] === kind) || App.catalog.find(c => (c.drivers || []).includes(kind));
    return genRadioDef(g, item && item.fw);
  },

  // 4.6: nomi esatti delle radio. Più radio della galleria possono avere la stessa
  // gestione (es. Radtel RT-950 Pro e Bajeton BJ-9000): la scheda porta il nome
  // della radio scelta (dalla galleria, dall'identificazione o l'ultima volta)
  cardsFor(kind) { return (App.catalog || []).filter(c => (c.drivers || []).includes(kind)); },
  baseName(kind) { return RADIO_DEF[kind] ? (kind === 'baj' ? 'Bajeton BJ-8300' : RADIO_DEF[kind].name) : (App.defs[kind] ? App.defs[kind].name : kind); },
  cardLabel(c, kind) {
    let n = ((c.brand || '') + ' ' + (c.name || '')).trim();
    if ((c.drivers || []).length > 1) {
      const m = /\(([^()]*)\)\s*$/.exec(App.baseName(kind));
      if (m && !n.toLowerCase().includes(m[1].toLowerCase())) n += ' (' + m[1] + ')';
    }
    return n;
  },
  defaultCard(kind) {
    const cs = App.cardsFor(kind);
    const id = ((App.init && App.init.settings && App.init.settings.modelCards) || {})[kind];
    const norm = s => s.toLowerCase().replace(/[\s\-\/.]/g, '');
    return cs.find(c => c.id === id) || (cs.length === 1 ? cs[0] : null) ||
      cs.find(c => norm(App.baseName(kind)).startsWith(norm(c.brand + c.name))) || null;
  },
  retitle(kind) {
    const v = Views[kind], tab = $(`.tab[data-tab="${kind}"]`);
    if (!v || !tab) return;
    tab.title = v.def.name;
    const t = tab.firstChild;
    if (t && t.nodeType === 3) t.nodeValue = v.def.name.replace(/^(Baofeng|Bajeton) /, '') + ' ';
  },

  // linguetta della scheda di una radio (creata quando la radio si apre dalla galleria)
  radioTab(kind) {
    if ($(`.tab[data-tab="${kind}"]`)) return;
    const def = Views[kind] ? Views[kind].def : App.radioDef(kind);
    const tab = h(`<button class="tab gen" data-tab="${esc(kind)}" title="${esc(def.name)}">${esc(def.name.replace(/^(Baofeng|Bajeton) /, ''))} <span class="dot" id="dot-${esc(kind)}"></span><span class="x" title="Chiudi la scheda">×</span></button>`);
    $('#radiotabs').appendChild(tab);
    tab.onclick = (e) => { if (e.target.classList.contains('x')) { e.stopPropagation(); App.closeRadio(kind).catch(reportError); } else App.showTab(kind); };
  },

  // apre (o mostra) la scheda di una radio generica; card (4.6): radio della
  // galleria di cui la scheda porta il nome
  openRadio(kind, show, card) {
    if (Views[kind]) {
      if (card) Views[kind].setCard(card, true);
      if (show) App.showTab(kind);
      return Views[kind];
    }
    const def = App.radioDef(kind);
    if (!def) throw new Error('Radio sconosciuta: ' + kind);
    const page = h(`<section class="page" id="page-${esc(kind)}"></section>`);
    $('main').appendChild(page);
    Views[kind] = new RadioView(kind, page, def);
    if (card) Views[kind].setCard(card, true);
    App.radioTab(kind);
    if (show) App.showTab(kind);
    return Views[kind];
  },

  async closeRadio(kind) {
    const v = Views[kind];
    if (!v) return;
    if (App.lastRadio === kind) App.lastRadio = null;
    if (Archive.target === kind) Archive.target = null;
    if (!v.def.generic) {
      // K67 Pro e Bajeton 8300: si chiude solo la linguetta, i dati restano in memoria
      const t = $(`.tab[data-tab="${kind}"]`);
      if (t) t.remove();
      if (currentTab === kind) App.showTab('gal');
      return;
    }
    if (!(await v.confirmDiscard())) return;
    delete Views[kind];
    v.page.remove();
    const t = $(`.tab[data-tab="${kind}"]`);
    if (t) t.remove();
    App.onDirty();
    App.saveSession();
    if (currentTab === kind) App.showTab('gal');
  },

  anyDirty() { return Object.values(Views).some(v => v.dirty); },
  _lastDirty: null,
  onDirty() {
    const d = App.anyDirty();
    if (d !== App._lastDirty) {
      App._lastDirty = d;
      api('dirty', { dirty: d }, { quiet: true }).catch(() => {});
    }
  },

  // copia di lavoro in sessionStorage: sopravvive a un ricaricamento accidentale della pagina
  saveSession: debounce(() => {
    try {
      const gen = {};
      for (const [k, v] of Object.entries(Views)) if (v.def.generic) gen[k] = v.sessionData();
      sessionStorage.setItem('work-' + TOKEN, JSON.stringify({ k67: Views.k67.sessionData(), baj: Views.baj.sessionData(), gen }));
    } catch (e) { /* spazio esaurito: non indispensabile */ }
  }, 800),

  async quit() {
    if (App.anyDirty()) {
      const dirty = Object.values(Views).filter(v => v.dirty);
      const names = dirty.map(v => v.def.name).join(', ');
      const c = await chooseDlg('Modifiche non salvate', `Ci sono modifiche non salvate per: ${names}.`,
        [{ id: 'no', label: 'Annulla' }, { id: 'quit', label: 'Esci senza salvare' }, { id: 'save', label: 'Salva ed esci', primary: true }]);
      if (!c || c === 'no') return;
      if (c === 'save') {
        try {
          for (const v of dirty) await v.save(false);
        } catch (e) { reportError(e); return; }
      }
    }
    if (Archive._saving) await Archive._saving;
    try { sessionStorage.removeItem('work-' + TOKEN); } catch (e) { /* */ }
    await api('dirty', { dirty: false }, { quiet: true }).catch(() => {});
    await api('quit', {}, { quiet: true }).catch(() => {});
    document.body.innerHTML = '<div class="welcome"><h2>Programma chiuso</h2><p>Puoi chiudere questa finestra.</p></div>';
  },
});

async function start() {
  try {
    App.init = await api('init', {}, { quiet: true });
    const cat = await api('catalog', {}, { quiet: true });
    App.catalog = cat.catalog || [];
    App.defs = cat.defs || {};
    App.noDriver = cat.noDriver || '';
  } catch (e) {
    document.body.innerHTML = `<div class="welcome"><h2>Impossibile avviare</h2><p>${esc(e.message)}</p></div>`;
    return;
  }
  CTCSS = App.init.ctcss;
  App.nativeDlg = App.init.nativeDlg;
  const dl = document.createElement('datalist');
  dl.id = 'ctcss-list';
  dl.innerHTML = '<option value="OFF">' + CTCSS.map(t => `<option value="${t}">`).join('');
  document.body.appendChild(dl);

  Views.k67 = new RadioView('k67', $('#page-k67'));
  Views.baj = new RadioView('baj', $('#page-baj'));
  Gallery.build($('#page-gal'));
  Archive.build($('#page-arch'));
  Archive.setPonti(App.init.ponti || []);

  try {
    const s = JSON.parse(sessionStorage.getItem('work-' + TOKEN) || 'null');
    if (s) {
      Views.k67.restore(s.k67); Views.baj.restore(s.baj);
      for (const k of ['k67', 'baj']) if (s[k] && s[k].card) Views[k].setCard(s[k].card, false);
      for (const k of ['k67', 'baj']) if (Views[k].loaded) App.radioTab(k);
      for (const [k, d] of Object.entries(s.gen || {})) {
        if (App.defs[k] && d && d.loaded) App.openRadio(k, false, d.card || '').restore(d);
      }
      App.onDirty();
    }
  } catch (e) { /* */ }

  $$('.tab').forEach(b => { if (!b.classList.contains('gen')) b.onclick = () => App.showTab(b.dataset.tab); });
  $('#btn-help').onclick = () => Dialogs.help();
  $('#btn-upd').onclick = () => Updates.manual();
  $('#btn-quit').onclick = () => App.quit();
  // a ogni avvio si parte dalla schermata iniziale
  App.showTab('home');
  // controllo automatico degli aggiornamenti a ogni avvio (in silenzio se non c'è connessione)
  setTimeout(() => Updates.check({ auto: true }).catch(() => {}), 1200);
  AutoIdent.start();

  document.addEventListener('keydown', e => {
    const k = e.key, ctrl = e.ctrlKey || e.metaKey;
    // niente ricaricamento della pagina (si perderebbero le modifiche)
    if (k === 'F5' || (ctrl && k.toLowerCase() === 'r')) { e.preventDefault(); return; }
    if (modalOpen()) return;
    if (k === 'F1') { e.preventDefault(); Dialogs.help(); return; }
    if (['info', 'gal', 'home', 'ident', 'conv'].includes(currentTab)) return;
    if (ctrl && k.toLowerCase() === 's' && currentTab !== 'arch') { e.preventDefault(); Views[currentTab].action(e.shiftKey ? 'saveas' : 'save'); return; }
    if (ctrl && k.toLowerCase() === 'o' && currentTab !== 'arch') { e.preventDefault(); Views[currentTab].action('open'); return; }
    // tasti della tabella anche se il fuoco non è sulla tabella
    const inField = e.target.closest('input,select,textarea');
    if (!inField && !e.target.closest('.gridwrap')) {
      const g = currentTab === 'arch' ? Archive.grid : Views[currentTab].grid;
      const loaded = Views[currentTab] ? Views[currentTab].loaded : true;
      const navKey = ['ArrowUp', 'ArrowDown', 'Delete', 'Insert', 'PageUp', 'PageDown'].includes(k);
      const ctrlKey = ctrl && k.length === 1 && 'cvxzya'.includes(k.toLowerCase());
      if (g && loaded && (navKey || ctrlKey)) {
        g.wrap.focus({ preventScroll: true });
        g.onKeyDown(e);
      }
    }
  });
  window.addEventListener('beforeunload', e => {
    if (App.anyDirty()) { e.preventDefault(); e.returnValue = ''; }
  });
  // segnala al programma che l'interfaccia è viva
  setInterval(() => api('ping', {}, { quiet: true }).catch(() => {}), 60000);
}

start();
