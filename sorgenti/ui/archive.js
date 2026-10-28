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
        <button class="btn" data-a="checkall" title="Spunta tutti i ponti visibili con i filtri attuali">☑ Seleziona tutti</button>
        <button class="btn" data-a="uncheckall" title="Toglie la spunta a tutti i ponti dell'archivio, anche a quelli nascosti dai filtri">☐ Deseleziona tutti</button>
        <button class="btn primary" data-a="forward" title="Scrive nella radio i ponti spuntati (o, se non ce ne sono, quelli selezionati)">📻 Inoltra alla radio</button>
        <button class="btn" data-a="copy" title="Copia i ponti spuntati (o selezionati): poi incollali in una radio con Ctrl+V">Copia</button>
        <span class="sep"></span>
        <button class="btn" data-a="names" title="Ricalcola i nomi dei canali con un modello">Rigenera nomi…</button>
        <button class="btn" data-a="verified" title="Toglie gli avvisi dai ponti selezionati dopo averli controllati">Segna come verificati</button>
        <button class="btn" data-a="dups" title="Trova i ponti con stessa frequenza, shift e località">Doppioni…</button>
        <span class="sep"></span>
        <span class="info" data-info></span>
      </div>
      <div class="ik2ane-credit">📡 <b>Grazie a IK2ANE.</b> L'archivio dei ponti è possibile grazie al lavoro di <b>IK2ANE</b>, che cura e mette a
        disposizione di tutti, gratuitamente, l'elenco aggiornato dei ponti radio italiani (<a href="#" data-link="https://www.ik2ane.it">www.ik2ane.it</a>).
        Usa «⬇ Importa…» per scaricare l'ultima versione.</div>
      <div class="arch-target" data-target></div>
      <div class="filters" data-f></div>
      <div class="content"><div class="gridwrap" data-g></div><div class="side" data-side></div></div>
      <div class="statusbar" data-status></div>`;
    $$('[data-a]', page).forEach(b => b.addEventListener('click', () => this.action(b.dataset.a)));
    $$('[data-link]', page).forEach(x => x.onclick = (e) => { e.preventDefault(); Info.open(x.dataset.link); });
    this.filter = new PontiFilter($('[data-f]', page), { onChange: () => { this.grid.sel.clear(); this.render(); } });
    this.grid = new Grid($('[data-g]', page), {
      columns: pontiColumns(true),
      checkbox: true,
      onCheck: (ids, on) => this.setChecked(ids, on),
      onCheckAll: on => this.setChecked(this.visible().map(p => p.id), on),
      getRows: () => this.visible().map(p => ({ id: p.id, p, checked: !!p.scelto, cls: p.avvisi && p.avvisi.length ? 'warn' : ((p.stato === 'Non attivo' || p.stato === 'Temporaneamente non attivo') ? 'off' : '') })),
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
    this.renderTarget();
    $('#arch-count').textContent = this.ponti.length ? '(' + this.ponti.length + ')' : '';
    $('[data-info]', this.page).textContent = App.init ? 'Salvato in: ' + App.init.archivePath : '';
    $('[data-a="undo"]', this.page).disabled = !this.undoStack.length;
    $('[data-a="redo"]', this.page).disabled = !this.redoStack.length;
  },

  renderStatus() {
    const vis = this.grid.rows.length;
    const sel = this.grid.sel.size;
    const warn = this.ponti.filter(p => p.avvisi && p.avvisi.length).length;
    const chk = this.ponti.filter(p => p.scelto).length;
    const hid = chk ? chk - this.grid.rows.filter(r => r.checked).length : 0;
    $('[data-status]', this.page).innerHTML = `<span>Ponti in archivio: ${this.ponti.length} · visibili con i filtri: ${vis}${chk ? ' · spuntati: ' + chk : ''}${hid ? ` (<b>${hid} nascosti dai filtri</b>: «Deseleziona tutti» toglie anche quelli)` : ''}${sel ? ' · selezionati: ' + sel : ''}</span>
      <span>${warn ? warn + ' con avvisi da controllare (⚠)' : ''}</span><span>Ctrl+A seleziona tutti i visibili · doppio clic per modificare</span>`;
  },

  // spunte: i ponti scelti per essere inoltrati alla radio (salvate con l'archivio)
  setChecked(ids, on) {
    const set = new Set(ids);
    let n = 0;
    for (const p of this.ponti) if (set.has(p.id) && !!p.scelto !== on) { if (on) p.scelto = true; else delete p.scelto; n++; }
    if (!n) return;
    this.changed();
  },

  // «Seleziona tutti»: i ponti visibili con i filtri attuali; «Deseleziona tutti»:
  // tutti i ponti spuntati, anche quelli nascosti dai filtri (prima della 3.6 solo i
  // visibili: chi spuntava dei ponti e poi cambiava filtro non riusciva più a
  // toglierli senza azzerare i filtri)
  checkVisible(on) {
    if (!on) {
      const all = this.ponti.filter(p => p.scelto);
      if (!all.length) { toast('Nessun ponte spuntato.'); return; }
      const hidden = all.length - this.visible().filter(p => p.scelto).length;
      this.setChecked(all.map(p => p.id), false);
      this.grid.sel.clear();
      this.render();
      toast(`Tolta la spunta a ${all.length} ponti${hidden ? ` (${hidden} erano nascosti dai filtri)` : ''}.`);
      return;
    }
    const v = this.visible();
    if (!v.length) { toast('Nessun ponte visibile con i filtri attuali.'); return; }
    this.setChecked(v.map(p => p.id), on);
    toast(`Spuntati ${v.length} ponti (quelli visibili con i filtri attuali).`);
  },

  // ponti da inoltrare: quelli spuntati, altrimenti quelli selezionati
  chosenPonti() {
    const chk = this.ponti.filter(p => p.scelto);
    if (chk.length) {
      const order = this.visible();
      const vis = new Set(order.map(p => p.id));
      // prima nell'ordine della tabella, poi gli spuntati nascosti dai filtri
      return order.filter(p => p.scelto).concat(chk.filter(p => !vis.has(p.id)));
    }
    return this.selectedPonti();
  },

  // radio di destinazione: quella scelta l'ultima volta, oppure l'ultima radio aperta
  target: null, targetLabel: '', targetType: '',
  targetKind() {
    if (this.target) return this.target;
    const t = App.lastRadio;
    return t && Views[t] ? t : null;
  },

  // nome mostrato della radio di destinazione (il modello scelto nell'elenco, se c'è)
  targetName(kind) {
    if (kind === this.target && this.targetLabel) return this.targetLabel;
    const v = Views[kind];
    return v ? v.def.name : ((RADIO_DEF[kind] || App.defs[kind] || {}).name || kind);
  },

  renderTarget() {
    const el = $('[data-target]', this.page);
    if (!el) return;
    const t = this.targetKind();
    el.innerHTML = t
      ? `📻 Radio di destinazione: <b>${esc(this.targetName(t))}</b> <button class="btn small" data-change>Cambia radio…</button>`
      : '📻 Nessuna radio di destinazione: con «Inoltra alla radio» potrai sceglierla dall\'elenco dei modelli compatibili.';
    const b = $('[data-change]', el);
    if (b) b.onclick = () => this.pickTarget().catch(reportError);
  },

  async pickTarget() {
    const r = await Dialogs.pickRadio('Scegli la radio di destinazione', 'Su quale radio vuoi scrivere i ponti scelti?');
    if (!r) return null;
    this.target = r.kind;
    this.targetLabel = r.label;
    this.targetType = r.type;
    this.renderTarget();
    return r.kind;
  },

  async forward() {
    let sel = this.chosenPonti();
    if (!sel.length) { toast('Spunta prima i ponti da inoltrare alla radio (o selezionali con clic, Maiusc+clic, Ctrl+clic).'); return; }
    let what = this.ponti.some(p => p.scelto) ? 'spuntati' : 'selezionati';
    const nVis = new Set(this.visible().map(p => p.id));
    const nHid = sel.filter(p => !nVis.has(p.id)).length;
    if (nHid) what += ` (${nHid} di questi sono nascosti dai filtri attuali)`;
    let kind = this.targetKind();
    if (!kind) kind = await this.pickTarget();
    else {
      const c = await chooseDlg('Inoltra alla radio', `Radio selezionata: ${this.targetName(kind)}\n\nScrivere i ${sel.length} ponti ${what} in questa radio?`,
        [{ id: 'no', label: 'Annulla' }, { id: 'other', label: 'Scegli un\'altra radio…' }, { id: 'ok', label: 'Conferma e scrivi', primary: true }]);
      if (!c || c === 'no') return;
      if (c === 'other') kind = await this.pickTarget();
    }
    // controllo preliminare: i dati devono essere compatibili con la radio scelta
    while (kind) {
      const chk = radioCompat(kind, sel, kind === this.target ? this.targetType : '');
      if (!chk.bad.length && !chk.warn.length) break;
      const c = await compatBanner(this.targetName(kind), chk);
      if (c === 'some') { const keep = new Set(chk.ok.concat(chk.warn.map(w => w.p))); sel = sel.filter(p => keep.has(p)); break; }
      if (c !== 'other') return;
      kind = await this.pickTarget();
    }
    if (!kind) return;
    await this.toRadio(kind, sel);
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
        case 'checkall': return this.checkVisible(true);
        case 'uncheckall': return this.checkVisible(false);
        case 'forward': return await this.forward();
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
    const sel = this.chosenPonti();
    if (!sel.length) { toast('Seleziona prima i ponti da copiare.'); return; }
    Clipboard.kind = 'ponti';
    Clipboard.items = sel.map(p => pontiToChan(p));
    toast(`Copiati ${sel.length} ponti: vai nella scheda di una radio, scegli la posizione e premi Incolla (Ctrl+V).`);
  },

  async toRadio(kind, sel) {
    sel = sel || this.chosenPonti();
    if (!sel.length) { toast('Spunta o seleziona prima i ponti.'); return; }
    const v = Views[kind] || App.openRadio(kind);
    App.radioTab(kind);
    if (this.target !== kind) { this.targetLabel = ''; this.targetType = ''; }
    this.target = kind;
    this.renderTarget();
    await Dialogs.insertInto(v, sel.map(p => pontiToChan(p, v.def)), 'ponti');
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

// ponte → canale; sulle radio DMR un ponte solo DMR diventa un canale DMR
function pontiToChan(p, def) {
  const c = { rx: p.rx, tx: p.tx, txTone: p.txTone || '', rxTone: p.rxTone || '', power: p.power || 0, narrow: !!p.narrow, scan: true, name: p.nome || '' };
  if (def && def.digital === 'DMR' && ponteSoloDigitale(p) && (p.tipi || []).includes('DMR')) {
    c.dig = true;
    c.txTone = c.rxTone = '';
  }
  return c;
}

// ------------------------------------------------------------ compatibilità ponti / radio

const TIPI_DIGITALI = ['DMR', 'C4FM', 'D-Star', 'NXDN'];
// ponte senza modo analogico (solo DMR, C4FM, D-Star, NXDN)
function ponteSoloDigitale(p) {
  const t = p.tipi || [];
  return t.length > 0 && t.every(x => TIPI_DIGITALI.includes(x));
}

// gamme delle due radio con file propri (5RH Pro / K67 Pro e Bajeton 8300: tre bande)
const BANDS_FILE = [{ lo: 136e6, hi: 174e6 }, { lo: 220e6, hi: 260e6 }, { lo: 400e6, hi: 520e6 }];
const PMR446 = { lo: 446000000, hi: 446200000 };
const GMRS = { lo: 462550000, hi: 467725000 };

// tipologia della radio: quella del modello scelto, altrimenti quella dei modelli
// del catalogo gestiti da quel driver (se sono tutti PMR / GMRS la radio è di uso libero)
function radioCategory(kind, chosenType) {
  if (chosenType) return chosenType;
  const types = new Set(App.catalog.filter(c => (c.drivers || []).includes(kind)).map(c => c.type));
  if (!types.size) return 'analog';
  if ([...types].every(t => t === 'pmr' || t === 'kids')) return 'pmr';
  if ([...types].every(t => t === 'gmrs')) return 'gmrs';
  return types.has('dmr') && types.size === 1 ? 'dmr' : 'analog';
}

// controllo preliminare prima di scrivere i ponti in una radio:
// { ok: [...], warn: [{ p, why }] (si possono scrivere con conferma), bad: [{ p, why }] }
function radioCompat(kind, ponti, chosenType) {
  const def = (Views[kind] && Views[kind].def) || App.defs[kind] || RADIO_DEF[kind] || {};
  const bands = def.bands && def.bands.length ? def.bands : BANDS_FILE;
  const cat = radioCategory(kind, chosenType);
  const inBand = f => bands.some(b => f >= b.lo && f <= b.hi);
  const inR = (f, r) => f >= r.lo && f <= r.hi;
  const mhz = f => (f / 1e6).toFixed(4).replace(/0+$/, '').replace(/\.$/, '');
  const ok = [], warn = [], bad = [];
  const gamme = bands.map(b => mhz(b.lo) + '-' + mhz(b.hi)).join(', ');
  for (const p of ponti) {
    let why = '', soft = '';
    if (ponteSoloDigitale(p) && !(def.digital === 'DMR' && (p.tipi || []).includes('DMR'))) {
      why = `ponte solo digitale (${(p.tipi || []).join(', ')}): ${def.digital === 'DMR' ? 'la radio gestisce solo DMR' : 'la radio è solo analogica'}`;
    } else if (!inBand(p.rx) || !inBand(p.tx)) {
      const fq = mhz(inBand(p.rx) ? p.tx : p.rx);
      why = cat === 'pmr' || cat === 'kids'
        ? `apparato PMR446 / di uso libero che non accetta ${fq} MHz (gamme della radio: ${gamme} MHz): i ponti radio non si possono caricare`
        : cat === 'gmrs'
          ? `apparato GMRS / FRS che non accetta ${fq} MHz (gamme della radio: ${gamme} MHz): i ponti radio non si possono caricare`
          : `frequenza ${fq} MHz fuori dalle gamme della radio (${gamme} MHz)`;
    } else if ((cat === 'pmr' || cat === 'kids') && !(inR(p.rx, PMR446) && inR(p.tx, PMR446))) {
      // venduta come PMR446, ma la radio (e il suo programma) accetta queste frequenze
      soft = `apparato venduto come PMR446 / di uso libero, ma la radio accetta queste frequenze (gamme: ${gamme} MHz)`;
    } else if (cat === 'gmrs' && !(inR(p.rx, GMRS) && inR(p.tx, GMRS))) {
      soft = `apparato venduto come GMRS / FRS, ma la radio accetta queste frequenze (gamme: ${gamme} MHz; il driver può comunque limitare la trasmissione)`;
    }
    if (why) bad.push({ p, why }); else if (soft) warn.push({ p, why: soft }); else ok.push(p);
  }
  return { ok, warn, bad, cat };
}

// banner: dispositivo non compatibile con i dati da inviare (o venduto come PMR446 /
// GMRS ma in grado di usarli). Restituisce 'no' | 'other' | 'some' (solo i ponti
// compatibili e quelli accettati dalla radio)
function compatBanner(name, chk) {
  const n = chk.bad.length + chk.ok.length + chk.warn.length;
  const list = arr => {
    const reasons = new Map();
    for (const b of arr) reasons.set(b.why, (reasons.get(b.why) || []).concat(b.p));
    return [...reasons].map(([why, ps]) => `<li><b>${ps.length} ${ps.length === 1 ? 'ponte' : 'ponti'}</b>: ${esc(why)}
      <span class="note">(${esc(ps.slice(0, 6).map(p => p.nome || p.sigla || (p.rx / 1e6).toFixed(4)).join(', '))}${ps.length > 6 ? ', …' : ''})</span></li>`).join('');
  };
  const usable = chk.ok.length + chk.warn.length;
  const onlyWarn = !chk.bad.length;
  const head = onlyWarn
    ? `<div class="compat-banner warnonly"><b>La radio scelta è venduta come apparato ${chk.cat === 'gmrs' ? 'GMRS / FRS' : 'PMR446 / di uso libero'}, ma accetta le frequenze dei ponti
        radio che stai inviando.</b><br>Radio scelta: ${esc(name)}. Puoi scriverli se sei radioamatore, rispettando le norme sull'uso delle frequenze
        (in Italia un apparato PMR446 omologato non si può usare fuori dalla banda PMR446). Nella radio non è ancora stato scritto nulla.</div>`
    : `<div class="compat-banner"><b>${usable ? `Hai scelto un dispositivo non compatibile con ${chk.bad.length} dei ${n} ponti che stai cercando di inviare alla radio.`
        : 'Hai scelto un dispositivo non compatibile con i dati che stai cercando di inviare alla radio.'}</b><br>
        Radio scelta: ${esc(name)}. Nella radio non è stato scritto nulla.</div>`;
  const okBtn = onlyWarn ? (usable === 1 ? 'Scrivi comunque il ponte' : `Scrivi comunque i ${usable} ponti`)
    : usable === 1 ? `Invia solo il ponte ${chk.warn.length ? 'utilizzabile' : 'compatibile'}` : `Invia solo i ${usable} ${chk.warn.length ? 'utilizzabili' : 'compatibili'}`;
  return new Promise(res => {
    const m = openModal(`<div class="dialog mid"><header>⚠ ${onlyWarn ? 'Radio venduta come ' + (chk.cat === 'gmrs' ? 'GMRS' : 'PMR446') : 'Radio non compatibile con i dati da inviare'}</header><div class="body">
      ${head}
      ${chk.bad.length ? `<p>Non compatibili:</p><ul class="compat-list">${list(chk.bad)}</ul>` : ''}
      ${chk.warn.length ? `<p>${onlyWarn ? '' : 'Accettati dalla radio, anche se venduta come apparato di uso libero:'}</p><ul class="compat-list">${list(chk.warn)}</ul>` : ''}
      ${!onlyWarn && chk.ok.length ? `<p class="note">${chk.ok.length} ponti sono compatibili.</p>` : ''}
      </div><footer><button class="btn" data-c="no">Annulla</button><button class="btn" data-c="other">Scegli un'altra radio…</button>
      ${usable ? `<button class="btn primary" data-c="some">${okBtn}</button>` : ''}</footer></div>`,
      { onEsc: () => { m.close(); res('no'); } });
    $$('[data-c]', m.el).forEach(b => b.onclick = () => { m.close(); res(b.dataset.c); });
  });
}
