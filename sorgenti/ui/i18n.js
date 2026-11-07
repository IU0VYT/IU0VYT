'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Lingue (4.10): l'interfaccia è scritta in italiano; con un'altra lingua ogni
// testo che compare nella pagina (anche quelli che arrivano dal programma:
// messaggi, avvisi, schede delle radio) si sostituisce con la sua traduzione.
// Le traduzioni sono in lang/<lingua>.tsv: «chiave<TAB>testo», dove la chiave è
// l'impronta FNV-1a a 32 bit (8 cifre esadecimali) del testo italiano con gli
// spazi ridotti a uno; lang/it.tsv contiene i testi italiani (servono per i testi
// con parti variabili, scritti con {0}, {1}…). Senza traduzione resta l'italiano.

const I18N = {
  langs: [
    { code: 'it', name: 'Italiano' },
    { code: 'en', name: 'English' },
    { code: 'fr', name: 'Français' },
    { code: 'de', name: 'Deutsch' },
    { code: 'es', name: 'Español' },
    { code: 'pt', name: 'Português' },
    { code: 'ru', name: 'Русский' },
    { code: 'zh', name: '中文' },
    { code: 'ja', name: '日本語' },
  ],
  lang: 'it',
  dict: null,        // chiave -> traduzione (lingua corrente)
  pats: null,        // testi italiani con parti variabili: { re, key, n }
  patIdx: null,      // parola -> pattern che la contengono
  itText: null,      // chiave -> testo italiano
  cache: new Map(),  // testo -> traduzione (o null)
  loaded: {},        // lingua -> { dict }
  textRec: new WeakMap(),  // nodo di testo -> { orig, set }
  attrRec: new WeakMap(),  // elemento -> { attributo: { orig, set } }
  observer: null,
  ATTRS: ['title', 'placeholder', 'aria-label', 'alt'],

  // impronta FNV-1a 32 bit del testo (UTF-8), come in i18n.go
  hash(s) {
    const b = new TextEncoder().encode(s);
    let h = 0x811c9dc5;
    for (let i = 0; i < b.length; i++) {
      h ^= b[i];
      h = Math.imul(h, 0x01000193) >>> 0;
    }
    return ('0000000' + h.toString(16)).slice(-8);
  },
  norm(s) { return String(s).replace(/\s+/g, ' ').trim(); },

  parseTSV(txt) {
    const m = new Map();
    for (const line of txt.split('\n')) {
      const i = line.indexOf('\t');
      if (i !== 8) continue;
      m.set(line.slice(0, 8), line.slice(9).replace(/\r$/, '').replace(/\\n/g, '\n').replace(/\\t/g, '\t'));
    }
    return m;
  },
  async fetchTSV(code) {
    const r = await fetch('lang/' + code + '.tsv', { cache: 'no-store' });
    if (!r.ok) throw new Error('lang/' + code + '.tsv: ' + r.status);
    return this.parseTSV(await r.text());
  },

  // testi italiani con parti variabili -> espressioni regolari
  async loadPatterns() {
    if (this.pats) return;
    this.itText = await this.fetchTSV('it');
    const pats = [], idx = new Map();
    const reEsc = s => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    for (const [key, t] of this.itText) {
      // parti variabili: {0}, {1}… (messaggi) e $1, $2… (nomi delle impostazioni)
      if (!/\{\d+\}|\$\d/.test(t)) continue;
      const parts = t.split(/(\{\d+\}|\$\d)/);
      let re = '^', lit = '', order = [];
      for (let i = 0; i < parts.length; i++) {
        // una parte variabile può anche essere vuota (pezzi facoltativi del messaggio)
        if (i % 2) { re += '(.*?)'; order.push(parts[i]); } else { re += reEsc(parts[i]); lit += ' ' + parts[i]; }
      }
      const words = (lit.match(/[\p{L}]{3,}/gu) || []);
      if (!words.length) continue; // solo parti variabili: non si traduce
      // «{0} canale {1}»: una sola parola fra due parti variabili prenderebbe
      // qualsiasi frase che la contiene
      if (words.length < 2 && /^(\{\d+\}|\$\d)/.test(t) && /(\{\d+\}|\$\d)$/.test(t)) continue;
      const p = { re: new RegExp(re + '$', 'u'), key, order, len: lit.length };
      pats.push(p);
      // indice: la parola più lunga del testo fisso
      const w = words.reduce((a, b) => (b.length > a.length ? b : a)).toLowerCase();
      if (!idx.has(w)) idx.set(w, []);
      idx.get(w).push(p);
    }
    for (const l of idx.values()) l.sort((a, b) => b.len - a.len);
    this.pats = pats;
    this.patIdx = idx;
  },

  async use(code) {
    if (!this.langs.some(l => l.code === code)) code = 'it';
    if (code !== 'it') {
      await this.loadPatterns();
      if (!this.loaded[code]) this.loaded[code] = { dict: await this.fetchTSV(code) };
      this.dict = this.loaded[code].dict;
    } else {
      this.dict = null;
    }
    this.lang = code;
    this.cache = new Map();
    document.documentElement.lang = code;
    this.applyAll();
    this.renderSelector();
  },

  // traduzione di un testo (null se non c'è): esatta, con parti variabili, a pezzi
  tr(text, depth) {
    if (!this.dict) return null;
    const n = this.norm(text);
    if (!n || !/[\p{L}]{2,}/u.test(n)) return null;
    if (this.cache.has(n)) return this.cache.get(n);
    let res = this.dict.get(this.hash(n));
    // frase senza il punto finale (il punto è nel testo che la contiene) o con un
    // punto in più (aggiunto dal testo che la contiene)
    if (res === undefined && !/[.!?…:;]$/.test(n)) {
      const t = this.dict.get(this.hash(n + '.'));
      if (t !== undefined) res = t.replace(/[.。]$/, '');
    } else if (res === undefined && /[^.]\.$/.test(n)) {
      const t = this.dict.get(this.hash(n.slice(0, -1)));
      if (t !== undefined) res = /[.!?。！？]$/.test(t) ? t : t + (/[\u3000-\u9fff]$/.test(t) ? '。' : '.');
    }
    if (res === undefined) res = this.trPattern(n, depth || 0);
    if (res == null && (depth || 0) < 3) res = this.trSplit(n, depth || 0);
    if (res === undefined) res = null;
    if (this.cache.size > 20000) this.cache.clear();
    this.cache.set(n, res);
    return res;
  },
  trPattern(n, depth) {
    // fra i testi con parti variabili che corrispondono vince quello con più testo fisso
    const seen = new Set();
    let best = null, bm = null;
    for (const w of (n.match(/[\p{L}]{3,}/gu) || [])) {
      const l = this.patIdx.get(w.toLowerCase());
      if (!l || seen.has(l)) continue;
      seen.add(l);
      for (const p of l) {
        if (best && p.len <= best.len) break;
        if (!this.dict.has(p.key)) continue;
        const m = p.re.exec(n);
        if (m) { best = p; bm = m; break; }
      }
    }
    {
      const p = best, m = bm;
      if (p) {
        const t = this.dict.get(p.key);
        // le parti variabili si traducono a loro volta se sono testi noti
        return t.replace(/\{\d+\}|\$\d/g, x => {
          const j = p.order.indexOf(x);
          if (j < 0) return x;
          const v = m[j + 1];
          return depth < 2 ? this.trPart(v, depth + 1) : v;
        });
      }
    }
    return null;
  },
  // parte variabile di un testo: tradotta se è un testo noto, anche se le manca il
  // punto finale (preso dal testo che la contiene) o se comincia con « — », «: »…
  trPart(v, depth) {
    let t = this.tr(v, depth);
    if (t != null) return t;
    if (/[^.]$/.test(v)) {
      t = this.tr(v + '.', depth);
      if (t != null) return t.replace(/[.。]$/, '');
    }
    const mm = v.match(/^([\s—–·:;,.(\-]+)(.*\S.*)$/su);
    if (mm) {
      t = this.tr(mm[2], depth);
      if (t != null) return mm[1] + t;
    }
    return v;
  },
  // messaggi composti («Lettura non riuscita: la radio non risponde»)
  trSplit(n, depth) {
    // si prova ogni punto di divisione: prima quelli in cui si traducono entrambe
    // le parti (frasi messe una dopo l'altra), poi il primo con almeno una parte
    let half = null;
    for (const sep of ['. ', ': ', '; ', ' — ', ' – ', ' · ', ', ', ' (', ' - ']) {
      for (let i = n.indexOf(sep), k = 0; i > 0 && k < (depth ? 2 : 12); i = n.indexOf(sep, i + 1), k++) {
        const a = n.slice(0, i);
        const b = sep === ' (' ? '(' + n.slice(i + 2) : n.slice(i + sep.length);
        const punct = /^[:;.,] $/.test(sep) ? sep[0] : '';
        const tb = this.tr(b, depth + 1);
        // il testo a sinistra con la sua punteggiatura («Lettura non riuscita:»)
        let ta = punct ? this.tr(a + punct, depth + 1) : null, joined = null;
        if (ta != null) joined = (x, y) => x + (x && y ? ' ' : '') + y;
        else {
          ta = this.tr(a, depth + 1);
          joined = (x, y) => x + (sep === ' (' ? ' ' : sep) + y;
        }
        if (ta != null && tb != null) return joined(ta, tb);
        if (!half && (ta != null || tb != null)) half = joined(ta ?? a, tb ?? b);
      }
    }
    return half;
  },
  // per il codice: testo tradotto (o l'originale)
  t(s) { return this.tr(s) ?? s; },

  skip(el) {
    return !el || !!el.closest('script,style,textarea,[contenteditable="true"],[data-notr],.notr');
  },
  applyText(node) {
    const el = node.parentElement;
    if (this.skip(el)) return;
    let rec = this.textRec.get(node);
    const cur = node.nodeValue;
    if (!rec || cur !== rec.set) { rec = { orig: cur, set: cur }; this.textRec.set(node, rec); }
    let out = rec.orig;
    if (this.dict) {
      const t = this.tr(rec.orig);
      if (t != null) {
        const lead = rec.orig.match(/^\s*/)[0], trail = rec.orig.match(/\s*$/)[0];
        out = lead + t + trail;
      }
    }
    if (out !== cur) node.nodeValue = out;
    rec.set = out;
  },
  applyAttr(el, a) {
    if (!el.hasAttribute(a) || this.skip(el)) return;
    let recs = this.attrRec.get(el);
    if (!recs) { recs = {}; this.attrRec.set(el, recs); }
    const cur = el.getAttribute(a);
    let rec = recs[a];
    if (!rec || cur !== rec.set) { rec = { orig: cur, set: cur }; recs[a] = rec; }
    const out = this.dict ? (this.tr(rec.orig) ?? rec.orig) : rec.orig;
    if (out !== cur) el.setAttribute(a, out);
    rec.set = out;
  },
  applyEl(el) {
    for (const a of this.ATTRS) this.applyAttr(el, a);
    if (el.tagName === 'INPUT' && /^(button|submit|reset)$/i.test(el.type)) this.applyAttr(el, 'value');
  },
  applyTree(root) {
    if (root.nodeType === 3) { this.applyText(root); return; }
    if (root.nodeType !== 1 && root.nodeType !== 11) return;
    if (root.nodeType === 1) { if (this.skip(root)) return; this.applyEl(root); }
    const w = document.createTreeWalker(root, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT);
    let n;
    while ((n = w.nextNode())) {
      if (n.nodeType === 3) this.applyText(n);
      else this.applyEl(n);
    }
  },
  applyAll() {
    this.applyTree(document.body);
    if (!this.titleOrig) this.titleOrig = document.title;
    document.title = this.dict ? (this.tr(this.titleOrig) ?? this.titleOrig) : this.titleOrig;
  },

  observe() {
    if (this.observer) return;
    this.observer = new MutationObserver(list => {
      if (!this.dict && this.lang === 'it') return;
      for (const m of list) {
        if (m.type === 'childList') m.addedNodes.forEach(n => this.applyTree(n));
        else if (m.type === 'characterData') this.applyText(m.target);
        else if (m.type === 'attributes') {
          const t = m.target;
          if (m.attributeName !== 'value') this.applyAttr(t, m.attributeName);
          else if (t.tagName === 'INPUT' && /^(button|submit|reset)$/i.test(t.type)) this.applyAttr(t, 'value');
        }
      }
    });
    this.observer.observe(document.body, {
      childList: true, subtree: true, characterData: true,
      attributes: true, attributeFilter: this.ATTRS.concat(['value']),
    });
  },

  // ---------------------------------------------------------------- selettore
  flag(code) {
    const f = {
      it: '<rect width="8" height="16" fill="#009246"/><rect x="8" width="8" height="16" fill="#fff"/><rect x="16" width="8" height="16" fill="#ce2b37"/>',
      en: '<rect width="24" height="16" fill="#012169"/><path d="M0 0L24 16M24 0L0 16" stroke="#fff" stroke-width="3.2"/><path d="M0 0L24 16M24 0L0 16" stroke="#c8102e" stroke-width="1.2"/><path d="M12 0V16M0 8H24" stroke="#fff" stroke-width="5"/><path d="M12 0V16M0 8H24" stroke="#c8102e" stroke-width="3"/>',
      fr: '<rect width="8" height="16" fill="#002395"/><rect x="8" width="8" height="16" fill="#fff"/><rect x="16" width="8" height="16" fill="#ed2939"/>',
      de: '<rect width="24" height="5.34" fill="#000"/><rect y="5.33" width="24" height="5.34" fill="#dd0000"/><rect y="10.66" width="24" height="5.34" fill="#ffce00"/>',
      es: '<rect width="24" height="16" fill="#aa151b"/><rect y="4" width="24" height="8" fill="#f1bf00"/><rect x="5" y="6" width="3" height="4" rx=".6" fill="#aa151b" opacity=".75"/>',
      pt: '<rect width="24" height="16" fill="#ff0000"/><rect width="9.6" height="16" fill="#006600"/><circle cx="9.6" cy="8" r="3.4" fill="#ffcc00"/><circle cx="9.6" cy="8" r="2.2" fill="#ff0000"/><rect x="8.7" y="6.6" width="1.8" height="2.6" rx=".4" fill="#fff"/>',
      ru: '<rect width="24" height="5.34" fill="#fff"/><rect y="5.33" width="24" height="5.34" fill="#0039a6"/><rect y="10.66" width="24" height="5.34" fill="#d52b1e"/>',
      zh: '<rect width="24" height="16" fill="#de2910"/><path d="M4 1.6l.9 2.7h2.8l-2.3 1.7.9 2.7L4 7l-2.3 1.7.9-2.7L.3 4.3h2.8z" fill="#ffde00"/><circle cx="8.6" cy="1.8" r=".7" fill="#ffde00"/><circle cx="10" cy="3.6" r=".7" fill="#ffde00"/><circle cx="10" cy="6" r=".7" fill="#ffde00"/><circle cx="8.6" cy="7.8" r=".7" fill="#ffde00"/>',
      ja: '<rect width="24" height="16" fill="#fff"/><circle cx="12" cy="8" r="4.8" fill="#bc002d"/>',
    }[code] || '';
    return `<svg class="flag" viewBox="0 0 24 16" width="22" height="15" aria-hidden="true">${f}<rect width="24" height="16" fill="none" stroke="rgba(0,0,0,.25)" stroke-width=".8"/></svg>`;
  },
  renderSelector() {
    const box = document.getElementById('langsel');
    if (!box) return;
    const cur = this.langs.find(l => l.code === this.lang) || this.langs[0];
    box.innerHTML = `<button class="langbtn" type="button" aria-haspopup="listbox" aria-expanded="false" data-notr
        title="${esc(cur.name)}">${this.flag(cur.code)}<span class="langcaret" aria-hidden="true">▾</span></button>`;
    const btn = box.firstElementChild;
    btn.onclick = (e) => { e.stopPropagation(); this.toggleMenu(btn); };
  },
  toggleMenu(btn) {
    let menu = document.getElementById('langmenu');
    if (menu) { this.closeMenu(); return; }
    menu = document.createElement('div');
    menu.id = 'langmenu';
    menu.setAttribute('role', 'listbox');
    menu.setAttribute('data-notr', '');
    menu.innerHTML = this.langs.map(l => `<button type="button" role="option" class="langopt${l.code === this.lang ? ' cur' : ''}"
        data-lang="${l.code}" aria-selected="${l.code === this.lang}">${this.flag(l.code)}<span>${esc(l.name)}</span></button>`).join('');
    document.body.appendChild(menu);
    const r = btn.getBoundingClientRect();
    menu.style.top = Math.round(r.bottom + 4) + 'px';
    menu.style.left = Math.round(Math.max(4, Math.min(r.left, window.innerWidth - menu.offsetWidth - 4))) + 'px';
    btn.setAttribute('aria-expanded', 'true');
    menu.querySelectorAll('.langopt').forEach(o => o.onclick = (e) => {
      e.stopPropagation();
      this.closeMenu();
      this.choose(o.dataset.lang);
    });
    const cur = menu.querySelector('.cur') || menu.firstElementChild;
    if (cur) cur.focus();
    menu.onkeydown = (e) => {
      const opts = [...menu.querySelectorAll('.langopt')];
      const i = opts.indexOf(document.activeElement);
      if (e.key === 'ArrowDown') { e.preventDefault(); opts[(i + 1) % opts.length].focus(); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); opts[(i - 1 + opts.length) % opts.length].focus(); }
      else if (e.key === 'Escape') { e.preventDefault(); this.closeMenu(); btn.focus(); }
    };
    setTimeout(() => document.addEventListener('click', this._close = () => this.closeMenu(), { once: true }), 0);
  },
  closeMenu() {
    const m = document.getElementById('langmenu');
    if (m) m.remove();
    const b = document.querySelector('#langsel .langbtn');
    if (b) b.setAttribute('aria-expanded', 'false');
    if (this._close) { document.removeEventListener('click', this._close); this._close = null; }
  },
  async choose(code) {
    try {
      await this.use(code);
    } catch (e) {
      console.error(e);
      return;
    }
    // la scelta si ricorda nelle impostazioni del programma (e vale al prossimo avvio)
    try { await api('lang', { lang: code }, { quiet: true }); } catch (e) { /* */ }
    try { localStorage.setItem('iu0vyt-lang', code); } catch (e) { /* */ }
  },

  // all'avvio: la lingua salvata (impostazioni del programma)
  async start(saved) {
    let code = saved;
    if (!code) { try { code = localStorage.getItem('iu0vyt-lang'); } catch (e) { /* */ } }
    this.observe();
    try { await this.use(code || 'it'); } catch (e) { console.error(e); await this.use('it'); }
  },
};
