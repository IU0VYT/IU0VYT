'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Archivio ponti condiviso, salvato sul PC.

const Archive = {
  ponti: [],
  undoStack: [], redoStack: [],
  sortKey: null, sortDir: 1,
  page: null, grid: null, filter: null,

  build(page) {
    this.page = page;
    page.innerHTML = `
      <div class="toolbar">
        <button class="btn primary" data-a="import" title="Importa ponti da Excel (.xls/.xlsx), CSV o da Internet">⬇ Importa…</button>
        <button class="btn" data-a="export" title="Esporta in Excel, CSV o CSV per CHIRP">⬆ Esporta…</button>
        <span class="sep"></span>
        <button class="btn" data-a="new">＋ Nuovo ponte</button>
        <button class="btn" data-a="del" title="Elimina i ponti selezionati (Canc)">Elimina</button>
        <button class="btn" data-a="undo" title="Annulla (Ctrl+Z)">↶</button>
        <button class="btn" data-a="redo" title="Ripeti (Ctrl+Y)">↷</button>
        <span class="sep"></span>
        <button class="btn" data-a="tok67" title="Inserisce i ponti selezionati nel K67">→ K67…</button>
        <button class="btn" data-a="tobaj" title="Inserisce i ponti selezionati nella Bajeton">→ Bajeton…</button>
        <button class="btn" data-a="toother" title="Inserisce i ponti selezionati in un'altra radio aperta dalla galleria dei modelli">→ Altra radio… ▾</button>
        <button class="btn" data-a="copy" title="Copia i ponti selezionati: poi incollali in una radio con Ctrl+V">Copia</button>
        <span class="sep"></span>
        <button class="btn" data-a="names" title="Ricalcola i nomi dei canali con un modello">Rigenera nomi…</button>
        <button class="btn" data-a="verified" title="Toglie gli avvisi dai ponti selezionati dopo averli controllati">Segna come verificati</button>
        <button class="btn" data-a="dups" title="Trova i ponti con stessa frequenza, shift e località">Doppioni…</button>
        <span class="sep"></span>
        <span class="info" data-info></span>
      </div>
      <div class="filters" data-f></div>
      <div class="content"><div class="gridwrap" data-g></div><div class="side" data-side></div></div>
      <div class="statusbar" data-status></div>`;
    $$('[data-a]', page).forEach(b => b.addEventListener('click', () => this.action(b.dataset.a)));
    this.filter = new PontiFilter($('[data-f]', page), { onChange: () => { this.grid.sel.clear(); this.render(); } });
    this.grid = new Grid($('[data-g]', page), {
      columns: pontiColumns(true),
      getRows: () => this.visible().map(p => ({ id: p.id, p, cls: p.avvisi && p.avvisi.length ? 'warn' : ((p.stato === 'Non attivo' || p.stato === 'Temporaneamente non attivo') ? 'off' : '') })),
      onEdit: (row, col, val) => this.edit(row.p, col.key, val),
      onKey: e => this.onKey(e),
      onSelChange: () => { this.renderSide(); this.renderStatus(); },
      onSort: k => { if (this.sortKey === k) this.sortDir = -this.sortDir; else { this.sortKey = k; this.sortDir = 1; } this.grid.render(); },
      sortKey: () => this.sortKey, sortDir: () => this.sortDir,
      emptyText: () => this.ponti.length ? 'Nessun ponte corrisponde ai filtri.' : 'L\'archivio è vuoto. Usa «Importa…» per caricare l\'elenco di IK2ANE (da Internet o da file) o un tuo file Excel/CSV, oppure aggiungi i ponti a mano.',
    });
  },

  setPonti(list) {
    this.ponti = list || [];
    this.filter.setData(this.ponti);
    this.render();
  },

  visible() {
    let v = this.ponti.filter(p => this.filter.match(p));
    if (this.sortKey) v = v.slice().sort(pontiSortFn(this.sortKey, this.sortDir));
    return v;
  },

  render() {
    this.grid.render();
    this.renderSide();
    this.renderStatus();
    $('#arch-count').textContent = this.ponti.length ? '(' + this.ponti.length + ')' : '';
    $('[data-info]', this.page).textContent = App.init ? 'Salvato in: ' + App.init.archivePath : '';
    $('[data-a="undo"]', this.page).disabled = !this.undoStack.length;
    $('[data-a="redo"]', this.page).disabled = !this.redoStack.length;
  },

  renderStatus() {
    const vis = this.grid.rows.length;
    const sel = this.grid.sel.size;
    const warn = this.ponti.filter(p => p.avvisi && p.avvisi.length).length;
    $('[data-status]', this.page).innerHTML = `<span>Ponti in archivio: ${this.ponti.length} · visibili con i filtri: ${vis}${sel ? ' · selezionati: ' + sel : ''}</span>
      <span>${warn ? warn + ' con avvisi da controllare (⚠)' : ''}</span><span>Ctrl+A seleziona tutti i visibili · doppio clic per modificare</span>`;
  },

  selectedPonti() {
    const ids = new Set(this.grid.selectedIds());
    return this.visible().filter(p => ids.has(p.id));
  },

  snapshot() {
    this.undoStack.push(JSON.stringify(this.ponti));
    if (this.undoStack.length > 40) this.undoStack.shift();
    this.redoStack = [];
  },
  undo() {
    if (!this.undoStack.length) return;
    this.redoStack.push(JSON.stringify(this.ponti));
    this.ponti = JSON.parse(this.undoStack.pop());
    this.changed(true);
  },
  redo() {
    if (!this.redoStack.length) return;
    this.undoStack.push(JSON.stringify(this.ponti));
    this.ponti = JSON.parse(this.redoStack.pop());
    this.changed(true);
  },
  changed(refreshFilter) {
    if (refreshFilter) this.filter.setData(this.ponti);
    this.render();
    this.save();
  },
  // salvataggio immediato; se ne è già in corso uno, ne parte un altro subito dopo
  _saving: null, _again: false,
  save() {
    if (this._saving) { this._again = true; return this._saving; }
    this._saving = (async () => {
      try {
        do {
          this._again = false;
          await api('archive/save', { ponti: this.ponti }, { quiet: true });
        } while (this._again);
      } catch (e) { reportError(new Error('Archivio non salvato: ' + e.message)); }
      finally { this._saving = null; }
    })();
    return this._saving;
  },

  async action(a) {
    try {
      switch (a) {
        case 'import': return await Dialogs.importPonti();
        case 'export': return await Dialogs.exportPonti();
        case 'new': return this.newPonte();
        case 'del': return await this.del();
        case 'undo': return this.undo();
        case 'redo': return this.redo();
        case 'tok67': return await this.toRadio('k67');
        case 'tobaj': return await this.toRadio('baj');
        case 'toother': {
          const gens = Object.values(Views).filter(v => v.def.generic);
          const items = gens.map(v => ({ label: v.def.name + (v.loaded ? '' : ' (da leggere o aprire)'), run: () => this.toRadio(v.kind) }));
          items.push({ label: 'Scegli il modello dalla galleria…', run: () => { this.copy(); App.showTab('gal'); } });
          return showMenuAt($('[data-a="toother"]'), items);
        }
        case 'copy': return this.copy();
        case 'names': return await Dialogs.renamePonti();
        case 'verified': return this.markVerified();
        case 'dups': return await this.duplicates();
      }
    } catch (e) { reportError(e); }
  },

  newPonte() {
    this.snapshot();
    const p = { id: 'n' + Date.now().toString(36) + Math.random().toString(36).slice(2, 6), nome: 'NUOVO PONTE', sigla: '', localita: '', provincia: '', regione: '',
      rx: 145500000, tx: 145500000, txTone: '', rxTone: '', narrow: false, power: 0, tipi: ['Analogico'], rete: '', nominativo: '', nodo: '', locatore: '',
      stato: '', avvisi: [], note: '', fonte: 'Inserito a mano' };
    this.ponti.unshift(p);
    this.filter.reset(true);
    this.filter.setData(this.ponti);
    this.grid.setSelection([p.id], p.id);
    this.changed();
    toast('Nuovo ponte aggiunto in cima: completa i dati nel pannello a destra.');
  },

  async del() {
    const sel = this.selectedPonti();
    if (!sel.length) { toast('Seleziona prima i ponti da eliminare.'); return; }
    if (sel.length > 1 && !(await confirmDlg('Elimina ponti', `Eliminare ${sel.length} ponti dall'archivio? (puoi annullare con Ctrl+Z)`, 'Elimina'))) return;
    this.snapshot();
    const ids = new Set(sel.map(p => p.id));
    this.ponti = this.ponti.filter(p => !ids.has(p.id));
    this.grid.sel.clear();
    this.changed(true);
    toast(`Eliminati ${sel.length} ponti.`);
  },

  copy() {
    const sel = this.selectedPonti();
    if (!sel.length) { toast('Seleziona prima i ponti da copiare.'); return; }
    Clipboard.kind = 'ponti';
    Clipboard.items = sel.map(pontiToChan);
    toast(`Copiati ${sel.length} ponti: vai nella scheda di una radio, scegli la posizione e premi Incolla (Ctrl+V).`);
  },

  async toRadio(kind) {
    const sel = this.selectedPonti();
    if (!sel.length) { toast('Seleziona prima i ponti (clic, Maiusc+clic, Ctrl+clic o Ctrl+A).'); return; }
    await Dialogs.insertInto(Views[kind], sel.map(pontiToChan), 'ponti');
  },

  markVerified() {
    const sel = this.selectedPonti().filter(p => p.avvisi && p.avvisi.length);
    if (!sel.length) { toast('Nessun ponte selezionato con avvisi.'); return; }
    this.snapshot();
    for (const p of sel) {
      p.note = [p.note, 'Verificato: ' + p.avvisi.join('; ')].filter(Boolean).join('; ');
      p.avvisi = [];
    }
    this.changed(true);
    toast(`${sel.length} ponti segnati come verificati (gli avvisi sono stati spostati nelle note).`);
  },

  async duplicates() {
    const key = chiavePonte;
    const seen = new Map(), dup = [];
    for (const p of this.ponti) { const k = key(p); if (seen.has(k)) dup.push(p); else seen.set(k, p); }
    if (!dup.length) { await alertDlg('Doppioni', 'Nessun doppione trovato (stessa frequenza, stesso TX, stessa località e sigla).'); return; }
    const c = await chooseDlg('Doppioni', `Trovati ${dup.length} ponti doppi (stessa frequenza, stesso TX, stessa località e sigla di un altro ponte).`,
      [{ id: 'no', label: 'Annulla' }, { id: 'show', label: 'Seleziona i doppioni' }, { id: 'del', label: `Elimina i ${dup.length} doppioni`, primary: true }]);
    if (c === 'del') {
      this.snapshot();
      const ids = new Set(dup.map(p => p.id));
      this.ponti = this.ponti.filter(p => !ids.has(p.id));
      this.changed(true);
      toast(`Eliminati ${dup.length} doppioni (Ctrl+Z per annullare).`);
    } else if (c === 'show') {
      this.filter.reset(true);
      this.filter.setData(this.ponti);
      this.grid.setSelection(dup.map(p => p.id), dup[0].id);
      this.render();
    }
  },

  addFromChannels(chans, from) {
    this.snapshot();
    const added = chans.map(c => ({ id: 'c' + Math.random().toString(36).slice(2, 12), nome: c.name, sigla: '', localita: '', provincia: '', regione: '',
      rx: c.rx, tx: c.tx, txTone: c.txTone && !c.txTone.startsWith('K67:') && normCTCSS(c.txTone) != null ? c.txTone : '', rxTone: c.rxTone && normCTCSS(c.rxTone) != null ? c.rxTone : '',
      narrow: !!c.narrow, power: c.power || 0, tipi: ['Analogico'], rete: '', nominativo: '', nodo: '', locatore: '', stato: '', avvisi: [], note: '', fonte: 'Da ' + from }));
    this.ponti = added.concat(this.ponti);
    this.changed(true);
    toast(`Aggiunti ${added.length} canali all'archivio ponti.`);
  },

  edit(p, key, val) {
    const before = JSON.stringify(p);
    const np = clone(p);
    switch (key) {
      case 'nome': np.nome = asciiName(val).slice(0, 16); break;
      case 'rx': {
        const hz = parseFreq(val);
        if (hz == null) return 'Frequenza non valida';
        const off = np.tx - np.rx; np.rx = hz; np.tx = Math.abs(off) > 50e6 ? np.tx : hz + off; break;
      }
      case 'shift': {
        const off = parseShift(val);
        if (off == null) return 'Shift non valido (es. -0.6, +1.6, +5, 0)';
        np.tx = np.rx + off; break;
      }
      case 'txTone': {
        const t = normCTCSS(val);
        if (t == null) return 'Tono CTCSS non valido (es. 67.0) oppure vuoto';
        np.txTone = t; break;
      }
      default: return null;
    }
    if (JSON.stringify(np) === before) return null;
    this.snapshot();
    Object.assign(p, np);
    this.changed();
    return null;
  },

  renderSide() {
    const side = $('[data-side]', this.page);
    const sel = this.selectedPonti();
    if (sel.length !== 1) { side.classList.remove('open'); return; }
    const p = sel[0];
    side.classList.add('open');
    const tipi = App.init.tipi;
    side.innerHTML = `<h3>${esc(p.nome || '(senza nome)')}</h3>
      ${p.avvisi && p.avvisi.length ? `<div class="avvisi"><b>Da controllare:</b><ul>${p.avvisi.map(a => '<li>' + esc(a) + '</li>').join('')}</ul>
        <button class="btn small" data-verif style="margin-top:6px">Segna come verificato</button></div>` : ''}
      ${p.orig ? `<div class="note" style="margin-bottom:8px">Valori originali nel file: ${Object.entries(p.orig).map(([k, v]) => esc(k) + ' «' + esc(v) + '»').join(', ')}</div>` : ''}
      <div class="form">
        <label>Nome canale</label><input data-k="nome" maxlength="16" value="${esc(p.nome)}">
        <label>Sigla</label><input data-k="sigla" value="${esc(p.sigla)}">
        <label>Località</label><input data-k="localita" value="${esc(p.localita)}">
        <label>Provincia</label><input data-k="provincia" value="${esc(p.provincia)}">
        <label>Regione</label><input data-k="regione" value="${esc(p.regione)}">
        <label>RX (MHz)</label><input data-k="rx" value="${esc(fmtMHz(p.rx))}">
        <label>TX (MHz)</label><input data-k="tx" value="${esc(fmtMHz(p.tx))}">
        <label>Shift</label><input data-k="shift" value="${esc(fmtShift(p.rx, p.tx))}" placeholder="es. -0.6 / +1.6 / +5">
        <label>Tono TX</label><input data-k="txTone" list="ctcss-list" value="${esc(p.txTone)}" placeholder="OFF">
        <label>Tono RX</label><input data-k="rxTone" list="ctcss-list" value="${esc(p.rxTone)}" placeholder="OFF">
        <label>Banda</label><select data-k="narrow"><option value="0" ${!p.narrow ? 'selected' : ''}>Larga (25 kHz)</option><option value="1" ${p.narrow ? 'selected' : ''}>Stretta (12,5 kHz)</option></select>
        <label>Potenza</label><select data-k="power">${POWERS.map((x, i) => `<option value="${i}" ${p.power === i ? 'selected' : ''}>${x}</option>`).join('')}</select>
        <label>Tipo</label><div class="chips" data-tipi>${tipi.map(t => `<span class="chip ${(p.tipi || []).includes(t) ? 'on' : ''}" data-t="${esc(t)}">${esc(t)}</span>`).join('')}</div>
        <label>Rete</label><input data-k="rete" value="${esc(p.rete)}">
        <label>Nominativo</label><input data-k="nominativo" value="${esc(p.nominativo)}">
        <label>Nodo EchoLink</label><input data-k="nodo" value="${esc(p.nodo)}">
        <label>Locatore</label><input data-k="locatore" value="${esc(p.locatore)}">
        <label>Stato</label><select data-k="stato">${['', 'Temporaneamente non attivo', 'Non attivo', 'Non H24'].map(s => `<option value="${esc(s)}" ${p.stato === s ? 'selected' : ''}>${esc(s || 'Attivo')}</option>`).join('')}</select>
        <label class="full">Note</label><textarea class="full" data-k="note">${esc(p.note)}</textarea>
        <label>Fonte</label><div class="note">${esc(p.fonte || '')}</div>
      </div>`;
    const commit = (k, el) => {
      const v = el.value;
      const np = clone(p);
      let err = null;
      if (k === 'rx' || k === 'tx') { const hz = parseFreq(v); if (hz == null) err = 'Frequenza non valida'; else if (k === 'rx') { const off = np.tx - np.rx; np.rx = hz; np.tx = Math.abs(off) > 50e6 ? np.tx : hz + off; } else np.tx = hz; }
      else if (k === 'shift') { const off = parseShift(v); if (off == null) err = 'Shift non valido'; else np.tx = np.rx + off; }
      else if (k === 'txTone' || k === 'rxTone') { const t = normCTCSS(v); if (t == null) err = 'Tono CTCSS non valido'; else np[k] = t; }
      else if (k === 'narrow') np.narrow = v === '1';
      else if (k === 'power') np.power = +v;
      else if (k === 'nome') np.nome = asciiName(v).slice(0, 16);
      else if (k === 'provincia' || k === 'nominativo' || k === 'locatore') np[k] = v.trim().toUpperCase();
      else np[k] = v;
      if (err) { toast(err, true); el.focus(); return; }
      if (JSON.stringify(np) === JSON.stringify(p)) return;
      this.snapshot();
      Object.assign(p, np);
      const filterChange = ['regione', 'provincia', 'rete', 'stato'].includes(k);
      this.changed(filterChange);
    };
    $$('[data-k]', side).forEach(el => {
      el.addEventListener('change', () => commit(el.dataset.k, el));
      el.addEventListener('keydown', e => { e.stopPropagation(); if (e.key === 'Enter' && el.tagName === 'INPUT') el.blur(); });
    });
    $$('[data-t]', side).forEach(c => c.onclick = () => {
      this.snapshot();
      const t = c.dataset.t, set = new Set(p.tipi || []);
      if (set.has(t)) set.delete(t); else set.add(t);
      p.tipi = tipi.filter(x => set.has(x));
      this.changed(true);
    });
    const v = $('[data-verif]', side);
    if (v) v.onclick = () => this.markVerified();
  },

  onKey(e) {
    const k = e.key, ctrl = e.ctrlKey || e.metaKey;
    if (k === 'Delete') { this.del().catch(reportError); return true; }
    if (ctrl && k.toLowerCase() === 'c') { this.copy(); return true; }
    if (ctrl && k.toLowerCase() === 'z' && !e.shiftKey) { this.undo(); return true; }
    if (ctrl && (k.toLowerCase() === 'y' || (k.toLowerCase() === 'z' && e.shiftKey))) { this.redo(); return true; }
    return false;
  },
};

function pontiToChan(p) {
  return { rx: p.rx, tx: p.tx, txTone: p.txTone || '', rxTone: p.rxTone || '', power: p.power || 0, narrow: !!p.narrow, scan: true, name: p.nome || '' };
}
