'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Filtri sui ponti (banda, tipo, stato, regione, provincia, rete, testo)
// e selettore con spunte usato in importazione ed esportazione.

const BANDE = ['VHF', 'UHF', 'Altre'];
const STATI = ['Attivo', 'Temporaneamente non attivo', 'Non attivo', 'Non H24'];
const statoOf = p => p.stato || 'Attivo';
const tipiOf = p => (p.tipi && p.tipi.length ? p.tipi : ['Altro']);

function pontiText(p) {
  return [p.nome, p.sigla, p.localita, p.provincia, p.regione, p.rete, p.nominativo, p.nodo, p.locatore, p.note, p.tipoOrig,
    (p.tipi || []).join(' '), fmtMHz(p.rx), fmtMHz(p.tx), p.txTone, (p.avvisi || []).join(' ')].join(' ').toLowerCase();
}

class PontiFilter {
  // opts: { onChange, tipiDefault: [..] | null }
  constructor(container, opts) {
    this.el = container;
    this.o = opts || {};
    this.data = [];
    this.reset(true);
  }

  reset(silent) {
    this.st = {
      bande: new Set(BANDE),
      tipi: new Set(this.o.tipiDefault || (App.init ? App.init.tipi : [])),
      tipiAll: !this.o.tipiDefault,
      stati: new Set(this.o.statiDefault || STATI),
      regioni: null, province: null, reti: null,
      testo: '', soloAvvisi: false,
    };
    if (!silent) { this.render(); this.changed(); }
  }

  changed() { if (this.o.onChange) this.o.onChange(); }

  setData(data) {
    this.data = data;
    this.render();
  }

  match(p) {
    const s = this.st;
    if (!s.bande.has(bandOf(p.rx))) return false;
    if (!s.tipiAll && !tipiOf(p).some(t => s.tipi.has(t))) return false;
    if (!s.stati.has(statoOf(p))) return false;
    if (s.regioni && !s.regioni.has(p.regione || '')) return false;
    if (s.province && !s.province.has(p.provincia || '')) return false;
    if (s.reti && !s.reti.has((p.rete || '').trim())) return false;
    if (s.soloAvvisi && !(p.avvisi && p.avvisi.length)) return false;
    if (s.testo) {
      const words = s.testo.toLowerCase().split(/\s+/).filter(Boolean);
      const t = pontiText(p);
      if (!words.every(w => t.includes(w))) return false;
    }
    return true;
  }

  counts(fn) {
    const m = new Map();
    for (const p of this.data) for (const v of [].concat(fn(p))) m.set(v, (m.get(v) || 0) + 1);
    return m;
  }

  render() {
    const s = this.st;
    const cb = this.counts(p => bandOf(p.rx)), ct = this.counts(tipiOf), cs = this.counts(statoOf);
    const tipiList = (App.init ? App.init.tipi : []).filter(t => ct.get(t) || (!s.tipiAll && s.tipi.has(t)));
    const chip = (grp, v, on, n) => `<span class="chip ${on ? 'on' : ''}" data-g="${grp}" data-v="${esc(v)}">${esc(v)}<span class="n">${n || 0}</span></span>`;
    this.el.innerHTML = `
      <div class="fgroup"><span class="lbl">Banda</span>${BANDE.map(b => chip('bande', b, s.bande.has(b), cb.get(b))).join('')}</div>
      <div class="fgroup"><span class="lbl">Tipo</span><span class="chip ${s.tipiAll ? 'on' : ''}" data-g="tipiAll">Tutti</span>${tipiList.map(t => chip('tipi', t, !s.tipiAll && s.tipi.has(t), ct.get(t))).join('')}</div>
      <div class="fgroup"><span class="lbl">Stato</span>${STATI.filter(x => cs.get(x) || !s.stati.has(x)).map(x => chip('stati', x, s.stati.has(x), cs.get(x))).join('')}</div>
      <div class="fgroup">
        <span class="ms" data-ms="regioni"></span><span class="ms" data-ms="province"></span><span class="ms" data-ms="reti"></span>
      </div>
      <div class="fgroup"><input class="search" type="search" placeholder="Cerca (nome, località, nominativo…)" value="${esc(s.testo)}" data-testo>
        <label title="Mostra solo i ponti con valori dubbi da controllare"><input type="checkbox" data-avvisi ${s.soloAvvisi ? 'checked' : ''}> solo con avvisi ⚠</label>
        <button class="btn small" data-reset>Azzera filtri</button></div>`;
    $$('.chip', this.el).forEach(c => c.onclick = () => this.toggle(c.dataset.g, c.dataset.v));
    const t = $('[data-testo]', this.el);
    const onT = debounce(() => { s.testo = t.value.trim(); this.changed(); }, 200);
    t.addEventListener('input', onT);
    t.addEventListener('keydown', e => e.stopPropagation());
    $('[data-avvisi]', this.el).onchange = e => { s.soloAvvisi = e.target.checked; this.changed(); };
    $('[data-reset]', this.el).onclick = () => this.reset();
    this.renderMS('regioni', 'Regione', p => p.regione || '');
    this.renderMS('province', 'Provincia', p => p.provincia || '', p => !s.regioni || s.regioni.has(p.regione || ''));
    this.renderMS('reti', 'Rete', p => (p.rete || '').trim());
  }

  toggle(g, v) {
    const s = this.st;
    if (g === 'tipiAll') { s.tipiAll = true; s.tipi = new Set(App.init.tipi); }
    else if (g === 'tipi') {
      if (s.tipiAll) { s.tipiAll = false; s.tipi = new Set([v]); }
      else { if (s.tipi.has(v)) s.tipi.delete(v); else s.tipi.add(v); if (!s.tipi.size) { s.tipiAll = true; s.tipi = new Set(App.init.tipi); } }
    } else {
      const set = s[g];
      if (set.has(v)) set.delete(v); else set.add(v);
    }
    this.render();
    this.changed();
  }

  renderMS(key, label, val, pre) {
    const host = $(`[data-ms="${key}"]`, this.el);
    const s = this.st;
    const counts = new Map();
    for (const p of this.data) { if (pre && !pre(p)) continue; const v = val(p); counts.set(v, (counts.get(v) || 0) + 1); }
    const vals = Array.from(counts.keys()).sort((a, b) => a.localeCompare(b, 'it'));
    const sel = s[key];
    const lbl = !sel ? 'tutte' : sel.size === 1 ? (Array.from(sel)[0] || '(vuoto)') : sel.size + ' scelte';
    host.innerHTML = `<button class="btn small" data-open>${esc(label)}: ${esc(lbl)} ▾</button>
      <div class="ms-pop"><div class="top"><input type="text" placeholder="Cerca…" data-q><button class="btn small" data-all>Tutte</button><button class="btn small" data-none>Nessuna</button></div>
      <div class="list">${vals.map(v => `<label data-v="${esc(v)}"><input type="checkbox" ${!sel || sel.has(v) ? 'checked' : ''}> ${esc(v || '(vuoto)')}<span class="n">${counts.get(v)}</span></label>`).join('')}</div></div>`;
    const pop = $('.ms-pop', host);
    $('[data-open]', host).onclick = (e) => {
      e.stopPropagation();
      const was = pop.classList.contains('open');
      $$('.ms-pop.open').forEach(x => x.classList.remove('open'));
      if (!was) { pop.classList.add('open'); $('[data-q]', pop).focus(); }
    };
    pop.addEventListener('click', e => e.stopPropagation());
    pop.addEventListener('keydown', e => { e.stopPropagation(); if (e.key === 'Escape') pop.classList.remove('open'); });
    $('[data-q]', pop).addEventListener('input', e => {
      const q = e.target.value.toLowerCase();
      $$('label', pop).forEach(l => l.style.display = l.dataset.v.toLowerCase().includes(q) ? '' : 'none');
    });
    const apply = () => {
      const checked = $$('label', pop).filter(l => $('input', l).checked).map(l => l.dataset.v);
      s[key] = checked.length === vals.length ? null : new Set(checked);
      if (key === 'regioni') s.province = null;
      $('[data-open]', host).textContent = `${label}: ${!s[key] ? 'tutte' : s[key].size === 1 ? (checked[0] || '(vuoto)') : s[key].size + ' scelte'} ▾`;
      this.changed();
      if (key === 'regioni') this.renderMS('province', 'Provincia', p => p.provincia || '', p => !s.regioni || s.regioni.has(p.regione || ''));
    };
    $$('input[type=checkbox]', pop).forEach(c => c.onchange = apply);
    $('[data-all]', pop).onclick = () => { $$('label', pop).forEach(l => { if (l.style.display !== 'none') $('input', l).checked = true; }); apply(); };
    $('[data-none]', pop).onclick = () => { $$('label', pop).forEach(l => { if (l.style.display !== 'none') $('input', l).checked = false; }); apply(); };
  }
}
document.addEventListener('click', () => $$('.ms-pop.open').forEach(x => x.classList.remove('open')));

// colonne comuni per le tabelle di ponti
function pontiColumns(editable) {
  const ed = (o) => editable ? o : undefined;
  return [
    { key: 'warn', title: '⚠', width: 30, html: p => p.p.avvisi && p.p.avvisi.length ? '<span class="ic">⚠</span>' : '', cellTip: r => (r.p.avvisi || []).join('\n'), sort: true, tip: 'Valori dubbi da controllare' },
    { key: 'nome', title: 'Nome canale', width: 150, text: r => r.p.nome, sort: true,
      edit: ed({ type: 'text', value: r => r.p.nome, maxlength: () => 16 }) },
    { key: 'sigla', title: 'Sigla', width: 64, text: r => r.p.sigla, sort: true },
    { key: 'localita', title: 'Località', width: 200, text: r => r.p.localita, sort: true, cellTip: r => r.p.localita },
    { key: 'provincia', title: 'Prov.', width: 48, text: r => r.p.provincia, sort: true },
    { key: 'regione', title: 'Regione', width: 120, text: r => r.p.regione, sort: true },
    { key: 'rx', title: 'RX (MHz)', width: 92, cls: 'num', text: r => fmtMHz(r.p.rx), sort: true, edit: ed({ type: 'text', value: r => fmtMHz(r.p.rx) }) },
    { key: 'shift', title: 'Shift', width: 66, cls: 'num', text: r => fmtShift(r.p.rx, r.p.tx), sort: true, edit: ed({ type: 'text', value: r => fmtShift(r.p.rx, r.p.tx).replace('split', '') }) },
    { key: 'txTone', title: 'Tono', width: 60, cls: 'num', text: r => r.p.txTone, sort: true, tip: 'Subtono CTCSS da trasmettere', edit: ed({ type: 'tone', value: r => r.p.txTone }) },
    { key: 'tipi', title: 'Tipo', width: 150, text: r => (r.p.tipi || []).join(', '), sort: true, cellTip: r => r.p.tipoOrig ? 'Nel file: ' + r.p.tipoOrig : '' },
    { key: 'rete', title: 'Rete', width: 120, text: r => r.p.rete, sort: true },
    { key: 'nominativo', title: 'Nominativo', width: 90, text: r => r.p.nominativo, sort: true },
    { key: 'stato', title: 'Stato', width: 120, text: r => r.p.stato || '', sort: true },
    { key: 'avvisi', title: 'Avvisi e note', width: 340, text: r => [].concat(r.p.avvisi || [], r.p.note ? [r.p.note] : []).join(' · '), cellTip: r => [].concat(r.p.avvisi || [], r.p.note ? [r.p.note] : []).join('\n') },
  ];
}

function pontiSortFn(key, dir) {
  const get = {
    warn: p => (p.avvisi || []).length, rx: p => p.rx, shift: p => p.tx - p.rx, txTone: p => parseFloat(p.txTone) || 0,
    tipi: p => (p.tipi || []).join(','), stato: p => p.stato || '',
  }[key] || (p => String(p[key] || '').toLowerCase());
  return (a, b) => {
    const x = get(a), y = get(b);
    const c = typeof x === 'number' ? x - y : String(x).localeCompare(String(y), 'it');
    return (c || a.rx - b.rx) * dir;
  };
}

// Selettore con filtri e spunte riga per riga (importazione/esportazione)
class PontiPicker {
  // opts: { tipiDefault, statiDefault, onChange }
  constructor(host, opts) {
    this.o = opts || {};
    host.innerHTML = '<div class="filters" data-f></div><div class="gridwrap" data-g style="flex:1"></div>';
    this.list = [];
    this.checked = new Set();
    this.sortKey = null; this.sortDir = 1;
    this.filter = new PontiFilter($('[data-f]', host), {
      tipiDefault: this.o.tipiDefault, statiDefault: this.o.statiDefault,
      onChange: () => { this.grid.render(false); this.notify(); },
    });
    this.grid = new Grid($('[data-g]', host), {
      columns: pontiColumns(false),
      checkbox: true,
      getRows: () => this.visible().map(p => ({ id: p.id, p, checked: this.checked.has(p.id), cls: p.avvisi && p.avvisi.length ? 'warn' : (p.stato === 'Non attivo' ? 'off' : '') })),
      onEdit: () => null,
      onCheck: (ids, on) => { ids.forEach(id => on ? this.checked.add(id) : this.checked.delete(id)); this.grid.render(); this.notify(); },
      onCheckAll: (on) => { this.visible().forEach(p => on ? this.checked.add(p.id) : this.checked.delete(p.id)); this.grid.render(); this.notify(); },
      onSort: k => { if (this.sortKey === k) this.sortDir = -this.sortDir; else { this.sortKey = k; this.sortDir = 1; } this.grid.render(); },
      sortKey: () => this.sortKey, sortDir: () => this.sortDir,
      emptyText: () => 'Nessun ponte corrisponde ai filtri.',
    });
  }
  setList(list, allChecked) {
    this.list = list;
    this.checked = new Set(allChecked === false ? [] : list.map(p => p.id));
    this.filter.setData(list);
    this.grid.render(false);
    this.notify();
  }
  visible() {
    let v = this.list.filter(p => this.filter.match(p));
    if (this.sortKey) v = v.slice().sort(pontiSortFn(this.sortKey, this.sortDir));
    return v;
  }
  chosen() { return this.visible().filter(p => this.checked.has(p.id)); }
  notify() { if (this.o.onChange) this.o.onChange(); }
}
