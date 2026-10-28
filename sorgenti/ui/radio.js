'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda di una radio: tabella dei canali, zone (Bajeton), operazioni.

const Clipboard = { kind: null, items: [] }; // kind: 'k67' | 'baj' | 'ponti'

const RADIO_DEF = {
  k67: { kind: 'k67', name: 'Baofeng K67 Pro', ext: '.xlc', slots: 640, nameMax: 16, powers: [0, 2], hasScan: true, zones: 0, perZone: 0, first: 1, narrow: true, fw: 'rh5', settings: true },
  baj: { kind: 'baj', name: 'Bajeton 8300', ext: '.dat', slots: 960, nameMax: 12, powers: [0, 1, 2], hasScan: true, zones: 15, perZone: 64, first: 1, narrow: true, fw: 'anc', settings: true },
};

// definizione di una radio gestita con i driver di CHIRP (dalla risposta di /api/catalog)
function genRadioDef(g, fw) {
  const powers = [], labels = {};
  for (const p of g.powerLevels || []) {
    if (!(p.sem in labels)) { powers.push(p.sem); labels[p.sem] = p.label; }
  }
  powers.sort((a, b) => a - b);
  return {
    kind: g.kind, name: g.name, ext: '.img', slots: g.slots, nameMax: g.nameMax || 0, powers, powerLabels: labels,
    hasScan: g.hasScan, zones: g.zones || 0, perZone: g.perZone || 0, first: g.first, narrow: g.narrow, bands: g.bands || [],
    dcs: new Set(g.dcs || []), charset: g.charset || '', noTxOff: g.noTxOff, notes: g.notes || '', generic: true, fw: fw || '', digital: g.digital || '',
    conn: g.conn || '', settings: !!g.settings,
  };
}

function powerLabel(d, p) { return (d.powerLabels && d.powerLabels[p]) || POWERS[p] || 'Alta'; }

// nome adattato ai caratteri ammessi dalla radio
function cleanNameFor(d, s) {
  let n = asciiName(s);
  if (d.charset) {
    if (!/[a-z]/.test(d.charset)) n = n.toUpperCase();
    n = Array.from(n).map(c => d.charset.includes(c) ? c : ' ').join('').replace(/\s+$/, '');
  }
  if (d.nameMax && n.length > d.nameMax) n = n.slice(0, d.nameMax);
  return n;
}

class RadioView {
  constructor(kind, page, def) {
    this.def = def || RADIO_DEF[kind];
    this.kind = kind;
    this.page = page;
    this.loaded = false;
    this.path = '';
    this.chans = [];
    this.bands = [];
    this.dirty = false;
    this.undoStack = [];
    this.redoStack = [];
    this.zone = 0;
    this.showEmpty = true;
    this.build();
  }

  // altre radio in cui copiare i canali: le schede aperte e le due radio principali
  others() { return Object.values(Views).filter(v => v !== this && (v.loaded || !v.def.generic)); }

  // ------------------------------------------------------------ interfaccia
  build() {
    const d = this.def;
    this.page.innerHTML = `
      <div class="toolbar">
        <button class="btn" data-a="open" title="Apri un file ${d.ext}${d.generic ? ' (anche salvato con CHIRP)' : ''} (Ctrl+O)">📂 Apri…</button>
        ${d.generic ? '' : '<button class="btn" data-a="new" title="Nuovo file vuoto con le impostazioni predefinite">Nuovo</button>'}
        <button class="btn primary" data-a="save" title="Salva (Ctrl+S)">💾 Salva</button>
        <button class="btn" data-a="saveas">Salva con nome…</button>
        <span class="sep"></span>
        <button class="btn" data-a="cread" title="Legge i canali dalla radio collegata con il cavo di programmazione">📻 Leggi dalla radio…</button>
        <button class="btn" data-a="cwrite" title="Scrive i canali di questa tabella nella radio collegata con il cavo">📻 Scrivi nella radio…</button>
        ${({ anc: '.anc', bf: '.BF', rh5: '.dat', kdhx: '.kdhx', bj9000: '.Firmware', kdh9g: '.kdh' })[d.fw] ? `<button class="btn" data-a="firmware" title="Aggiorna il firmware della radio (file ${({ anc: '.anc', bf: '.BF', rh5: '.dat', kdhx: '.kdhx', bj9000: '.Firmware', kdh9g: '.kdh' })[d.fw]}, anche dentro archivi .zip o .rar, dal PC o da Internet)">⬆ Aggiorna firmware…</button>` : ''}
        ${d.settings ? '<button class="btn" data-a="settings" title="Impostazioni della radio (menu, display, VOX, DTMF…), con la descrizione di ogni voce">⚙ Impostazioni…</button>' : ''}
        ${this.kind === 'baj' ? `<label class="note" style="white-space:nowrap" title="Il programma Bajeton BT-11_CPS v1.0.9 salva i file .dat in un formato diverso dalla v1.0.8 e le due versioni non aprono i file dell'altra. Scegli la versione installata sul tuo PC: il file aperto viene convertito (canali e impostazioni restano gli stessi) e i nuovi file e le letture dalla radio useranno questo formato.">File per BT-11_CPS
          <select data-datver><option value="1.0.8">v1.0.8</option><option value="1.0.9">v1.0.9</option></select></label>` : ''}
        <span class="sep"></span>
        <span class="info" data-info></span>
      </div>
      <div class="toolbar" data-edit>
        <button class="btn" data-a="add" title="Aggiunge un canale nella prima posizione libera">＋ Aggiungi</button>
        <button class="btn" data-a="insert" title="Inserisce una riga vuota spostando in basso i canali (Ins)">Inserisci riga</button>
        <button class="btn" data-a="del" title="Cancella i canali selezionati lasciando la posizione vuota (Canc)">Cancella</button>
        <button class="btn" data-a="delc" title="Elimina i canali selezionati e fa risalire quelli successivi">Elimina e compatta</button>
        <span class="sep"></span>
        <button class="btn" data-a="up" title="Sposta su (Alt+↑)">↑</button>
        <button class="btn" data-a="down" title="Sposta giù (Alt+↓)">↓</button>
        <button class="btn" data-a="sort">Ordina ▾</button>
        <button class="btn" data-a="compact" title="Elimina le posizioni vuote tra i canali">Compatta</button>
        ${d.zones ? '<button class="btn" data-a="tozone">Sposta in zona ▾</button>' : ''}
        <span class="sep"></span>
        <button class="btn" data-a="copy" title="Copia (Ctrl+C)">Copia</button>
        <button class="btn" data-a="cut" title="Taglia (Ctrl+X)">Taglia</button>
        <button class="btn" data-a="paste" title="Incolla nella posizione selezionata (Ctrl+V)">Incolla</button>
        <span class="sep"></span>
        <button class="btn" data-a="undo" title="Annulla (Ctrl+Z)">↶</button>
        <button class="btn" data-a="redo" title="Ripeti (Ctrl+Y)">↷</button>
        <span class="sep"></span>
        <button class="btn" data-a="copyto" title="Copia i canali selezionati in un'altra radio">Copia in… ▾</button>
        <button class="btn" data-a="toarch" title="Aggiunge i canali selezionati all'archivio ponti">→ Archivio</button>
        <button class="btn" data-a="import" title="Inserisce i canali di un altro file: K67 (.xlc), Bajeton (.dat), immagini .img di qualsiasi radio (anche di CHIRP), CSV di CHIRP, tabelle Excel/CSV">📥 Importa…</button>
        <button class="btn" data-a="export" title="Esporta i canali in Excel, CSV o CSV di CHIRP, oppure nel formato di un'altra radio">Esporta…</button>
        <label style="margin-left:6px;white-space:nowrap"><input type="checkbox" data-empty checked> mostra posizioni vuote</label>
      </div>
      ${d.zones ? '<div class="zonebar" data-zones></div>' : ''}
      <div class="content"><div class="gridwrap" data-grid></div></div>
      <div class="statusbar" data-status></div>`;
    $$('[data-a]', this.page).forEach(b => {
      const a = b.dataset.a;
      if (a === 'sort') {
        attachMenu(b, [
          { label: 'Per frequenza', run: () => this.sort('rx') },
          { label: 'Per nome', run: () => this.sort('name') },
          { label: 'Per tono', run: () => this.sort('tone') },
        ]);
      } else if (a === 'copyto') {
        b.addEventListener('click', () => this.copyToMenu(b));
      } else if (a === 'tozone') {
        const items = [];
        for (let z = 0; z < d.zones; z++) items.push({ label: 'Zona ' + (z + 1), run: () => this.moveToZone(z) });
        attachMenu(b, items);
      } else b.addEventListener('click', () => this.action(a));
    });
    $('[data-empty]', this.page).addEventListener('change', e => { this.showEmpty = e.target.checked; this.render(); });
    const dv = $('[data-datver]', this.page);
    if (dv) {
      dv.value = (App.init.settings && App.init.settings.bajVer) || '1.0.8';
      dv.addEventListener('change', async () => {
        const ver = dv.value;
        try {
          await api('radio/datver', { ver }, { busyText: 'Conversione del file…' });
          if (App.init.settings) App.init.settings.bajVer = ver;
          if (this.loaded && this.datVer && this.datVer !== ver) {
            this.datVer = ver;
            this.dirty = true;
            App.onDirty();
            this.render();
            toast(`File convertito per BT-11_CPS v${ver}: salvalo per usarlo con quella versione del programma.`);
          } else toast(`I nuovi file e le letture dalla radio saranno nel formato di BT-11_CPS v${ver}.`);
        } catch (e) { dv.value = this.datVer || '1.0.8'; reportError(e); }
      });
    }
    this.grid = new Grid($('[data-grid]', this.page), {
      columns: this.columns(),
      getRows: () => this.rows(),
      onEdit: (row, col, val) => this.edit(row, col, val),
      onKey: e => this.onKey(e),
      onSelChange: () => this.renderStatus(),
      editRedirect: (row, col) => (!row.ch && col.key !== 'rx') ? 'rx' : null,
      emptyText: () => this.loaded ? 'Nessun canale in questa vista.' : '',
    });
    this.grid.wrap.addEventListener('contextmenu', e => this.contextMenu(e));
    this.render();
  }

  columns() {
    const d = this.def;
    const tones = { type: 'tone', value: r => r.ch ? r.ch[this._tk] : '' };
    const cols = [
      { key: 'pos', title: d.zones ? 'Pos.' : 'N°', width: 58, cls: 'pos', text: r => this.posText(r.id) },
      { key: 'name', title: 'Nome', width: d.nameMax * 9 + 30, cls: 'name', text: r => r.ch ? r.ch.name : '',
        edit: { type: 'text', value: r => r.ch ? r.ch.name : '', maxlength: () => d.nameMax } },
      { key: 'rx', title: 'RX (MHz)', width: 100, cls: 'num', text: r => r.ch ? fmtMHz(r.ch.rx) : '',
        bad: r => r.ch && this.outOfBand(r.ch.rx), cellTip: r => r.ch && this.outOfBand(r.ch.rx) ? 'Fuori dalle gamme della radio' : '',
        edit: { type: 'text', value: r => r.ch ? fmtMHz(r.ch.rx) : '' } },
      { key: 'shift', title: 'Shift', width: 72, cls: 'num', text: r => r.ch && !r.ch.txOff ? fmtShift(r.ch.rx, r.ch.tx) : '', tip: 'Shift in MHz (es. -0.6, +1.6, +5); vuoto = isofrequenza',
        edit: { type: 'text', value: r => r.ch ? fmtShift(r.ch.rx, r.ch.tx).replace('split', '') : '' } },
      { key: 'tx', title: 'TX (MHz)', width: 100, cls: 'num', text: r => r.ch ? (r.ch.txOff ? 'OFF' : fmtMHz(r.ch.tx)) : '',
        tip: d.generic && !d.noTxOff ? 'Frequenza di trasmissione; OFF = canale di sola ricezione' : '',
        bad: r => r.ch && !r.ch.txOff && this.outOfBand(r.ch.tx), edit: { type: 'text', value: r => r.ch ? (r.ch.txOff ? 'OFF' : fmtMHz(r.ch.tx)) : '' } },
      { key: 'txTone', title: 'Tono TX', width: 76, cls: 'num', text: r => r.ch ? (r.ch.dig ? '—' : fmtTone(r.ch.txTone)) : '', tip: 'Subtono trasmesso (per aprire il ponte): CTCSS es. 88.5, oppure DCS es. D023N / D023I',
        edit: { type: 'tone', value: r => r.ch ? r.ch.txTone : '' } },
      { key: 'rxTone', title: 'Tono RX', width: 76, cls: 'num', text: r => r.ch ? (r.ch.dig ? '—' : fmtTone(r.ch.rxTone)) : '', tip: 'Subtono in ricezione (squelch a tono): CTCSS es. 88.5, oppure DCS es. D023N / D023I',
        edit: { type: 'tone', value: r => r.ch ? r.ch.rxTone : '' } },
      { key: 'power', title: 'Potenza', width: d.powerLabels ? 96 : 72, text: r => r.ch ? powerLabel(d, r.ch.power) : '',
        edit: { type: 'select', options: () => d.powers.map(p => ({ value: p, label: powerLabel(d, p) })), value: r => r.ch ? r.ch.power : 0 } },
    ];
    if (d.digital) cols.splice(2, 0, { key: 'mode', title: 'Modo', width: 86, text: r => r.ch ? (r.ch.dig ? d.digital : 'Analogico') : '',
      tip: 'Canale analogico (FM) o digitale (' + d.digital + '); sui canali digitali i subtoni non si usano e le impostazioni digitali restano quelle della radio',
      edit: { type: 'select', options: [{ value: '0', label: 'Analogico' }, { value: '1', label: d.digital }], value: r => r.ch && r.ch.dig ? '1' : '0' } });
    if (d.narrow !== false) cols.push({ key: 'narrow', title: 'Banda', width: 70, text: r => r.ch ? (r.ch.narrow ? 'Stretta' : 'Larga') : '', tip: 'Larghezza di banda: Larga 25 kHz, Stretta 12,5 kHz',
      edit: { type: 'select', options: [{ value: '0', label: 'Larga' }, { value: '1', label: 'Stretta' }], value: r => r.ch && r.ch.narrow ? '1' : '0' } });
    if (d.hasScan) cols.push({ key: 'scan', title: 'Scan', width: 52, text: r => r.ch ? (r.ch.scan ? 'Sì' : 'No') : '',
      edit: { type: 'select', options: [{ value: '1', label: 'Sì' }, { value: '0', label: 'No' }], value: r => r.ch && r.ch.scan ? '1' : '0' } });
    return cols;
  }

  posText(i) {
    const d = this.def;
    if (!d.zones) return String(i + (d.first == null ? 1 : d.first));
    const z = Math.floor(i / d.perZone) + 1, n = i % d.perZone + 1;
    return this.zone < 0 ? z + '/' + n : String(n);
  }

  range() {
    const d = this.def;
    if (!d.zones || this.zone < 0) return [0, d.slots];
    return [this.zone * d.perZone, (this.zone + 1) * d.perZone];
  }

  rows() {
    if (!this.loaded) return [];
    const [a, b] = this.range();
    const out = [];
    for (let i = a; i < b; i++) {
      if (this.def.zones && this.zone < 0 && i % this.def.perZone === 0) {
        const n = this.usedIn(i, i + this.def.perZone);
        if (this.showEmpty || n) out.push({ head: 'Zona ' + (i / this.def.perZone + 1) + ' — ' + n + ' canali', id: 'z' + i });
      }
      const ch = this.chans[i];
      if (!ch && !this.showEmpty) continue;
      out.push({ id: i, ch, cls: ch ? '' : 'empty' });
    }
    return out;
  }

  usedIn(a, b) { let n = 0; for (let i = a; i < b; i++) if (this.chans[i]) n++; return n; }

  outOfBand(hz) {
    const bands = this.def.generic ? this.def.bands : this.kind === 'k67' ? this.bands : null;
    if (!bands || !bands.length) return false;
    return !bands.some(b => hz >= b.lo && hz <= b.hi);
  }

  render() {
    const info = $('[data-info]', this.page);
    if (this.loaded) {
      info.innerHTML = this.path ? `File: <b>${esc(this.path)}</b>${this.dirty ? ' — <span style="color:#b26a00">modificato, non salvato</span>' : ''}`
        : this.fromRadio ? `<b>Canali letti dalla radio</b> <span style="color:#b26a00">(non ancora salvati in un file)</span>` : `<b>Nuovo file</b> (non ancora salvato)`;
    } else info.textContent = 'Nessun file aperto';
    $$('[data-edit] .btn', this.page).forEach(b => b.disabled = !this.loaded);
    $$('[data-a="save"],[data-a="saveas"],[data-a="cwrite"]', this.page).forEach(b => b.disabled = !this.loaded);
    if (this.loaded) {
      $('[data-a="undo"]', this.page).disabled = !this.undoStack.length;
      $('[data-a="redo"]', this.page).disabled = !this.redoStack.length;
    }
    this.renderZones();
    if (!this.loaded && this.def.generic) {
      const d = this.def;
      this.grid.wrap.innerHTML = `<div class="welcome"><h2>${esc(d.name)}</h2>
        <p>Leggi i canali dalla radio con il cavo di programmazione, oppure apri un file <b>.img</b> della radio (anche salvato con CHIRP).
        Le impostazioni della radio restano quelle lette: qui si modificano canali e ponti, poi li puoi scrivere nella radio e salvare in un file .img
        (apribile anche con CHIRP).</p>
        <p class="note">${d.slots} canali · nomi fino a ${d.nameMax} caratteri · potenze: ${esc(d.powers.map(p => powerLabel(d, p)).join(', '))}${d.notes ? '<br>' + esc(d.notes) : ''}</p>
        <div class="actions"><button class="btn primary" data-w="cread">📻 Leggi dalla radio…</button><button class="btn" data-w="open">📂 Apri file .img…</button></div></div>`;
      $('[data-w="cread"]', this.grid.wrap).onclick = () => this.action('cread');
      $('[data-w="open"]', this.grid.wrap).onclick = () => this.action('open');
    } else if (!this.loaded) {
      this.grid.wrap.innerHTML = `<div class="welcome"><h2>${esc(this.def.name)}</h2>
        <p>Leggi i canali direttamente dalla radio con il cavo di programmazione, oppure apri un file ${esc(this.def.ext)}
        salvato con il programma della radio, oppure crea un file nuovo. Dopo le modifiche puoi scrivere i canali nella radio
        e salvarli in un file ${esc(this.def.ext)}.</p>
        <div class="actions"><button class="btn primary" data-w="cread">📻 Leggi dalla radio…</button><button class="btn" data-w="open">📂 Apri file ${esc(this.def.ext)}…</button><button class="btn" data-w="new">Nuovo file vuoto</button></div></div>`;
      $('[data-w="cread"]', this.grid.wrap).onclick = () => this.action('cread');
      $('[data-w="open"]', this.grid.wrap).onclick = () => this.action('open');
      $('[data-w="new"]', this.grid.wrap).onclick = () => this.action('new');
    } else this.grid.render();
    this.renderStatus();
    const dot = document.getElementById('dot-' + this.kind);
    if (dot) dot.className = 'dot' + (this.dirty ? ' dirty' : '');
  }

  renderZones() {
    const bar = $('[data-zones]', this.page);
    if (!bar) return;
    if (!this.loaded) { bar.innerHTML = ''; return; }
    const d = this.def;
    let html = '';
    for (let z = 0; z < d.zones; z++) {
      const n = this.usedIn(z * d.perZone, (z + 1) * d.perZone);
      html += `<button class="zonebtn ${this.zone === z ? 'active' : ''}" data-z="${z}">Zona ${z + 1}<span class="n">${n}</span></button>`;
    }
    html += `<button class="zonebtn ${this.zone < 0 ? 'active' : ''}" data-z="-1">Tutte</button>`;
    bar.innerHTML = html;
    $$('[data-z]', bar).forEach(b => b.onclick = () => { this.zone = +b.dataset.z; this.grid.sel.clear(); this.grid.cur.r = 0; this.render(); this.grid.wrap.focus(); });
  }

  renderStatus() {
    const st = $('[data-status]', this.page);
    if (!this.loaded) { st.textContent = ''; return; }
    const used = this.chans.filter(Boolean).length;
    const sel = this.grid.sel.size;
    let s = `Canali programmati: ${used} su ${this.def.slots}`;
    if (this.def.zones && this.zone >= 0) s += ` · Zona ${this.zone + 1}: ${this.usedIn(...this.range())} su ${this.def.perZone}`;
    if (sel) s += ` · Selezionati: ${sel}`;
    st.innerHTML = `<span>${esc(s)}</span><span>Doppio clic o Invio per modificare · tasto destro per il menu</span>`;
  }

  // ------------------------------------------------------------ stato e annulla
  snapshot() {
    this.undoStack.push(JSON.stringify(this.chans));
    if (this.undoStack.length > 60) this.undoStack.shift();
    this.redoStack = [];
  }
  changed() {
    this.dirty = true;
    App.onDirty();
    this.render();
    App.saveSession();
  }
  undo() {
    if (!this.undoStack.length) return;
    this.redoStack.push(JSON.stringify(this.chans));
    this.chans = JSON.parse(this.undoStack.pop());
    this.changed();
  }
  redo() {
    if (!this.redoStack.length) return;
    this.undoStack.push(JSON.stringify(this.chans));
    this.chans = JSON.parse(this.redoStack.pop());
    this.changed();
  }

  load(resp) {
    this.loaded = true;
    this.path = resp.path || '';
    this.fromRadio = false;
    this.chans = resp.channels.map(c => c || null);
    while (this.chans.length < this.def.slots) this.chans.push(null);
    this.bands = resp.bands || [];
    this.datVer = resp.datVer || '';
    const dv = $('[data-datver]', this.page);
    if (dv && this.datVer) dv.value = this.datVer;
    this.dirty = false;
    this.undoStack = []; this.redoStack = [];
    this.zone = this.def.zones ? 0 : 0;
    this.grid.sel.clear();
    this.grid.cur = { r: 0, c: 1 };
    App.onDirty();
    this.render();
    App.saveSession();
  }

  // ------------------------------------------------------------ azioni
  async action(a) {
    try {
      switch (a) {
        case 'open': return await this.open();
        case 'new': return await this.newFile();
        case 'save': return await this.save(false);
        case 'saveas': return await this.save(true);
        case 'add': return this.add();
        case 'insert': return this.insertRow();
        case 'del': return this.deleteSel(false);
        case 'delc': return this.deleteSel(true);
        case 'up': return this.move(-1);
        case 'down': return this.move(1);
        case 'compact': return this.compact();
        case 'copy': return this.copy();
        case 'cut': return this.cut();
        case 'paste': return await this.paste();
        case 'undo': return this.undo();
        case 'redo': return this.redo();
        case 'copyto': return await this.copyTo();
        case 'toarch': return await this.toArchive();
        case 'export': return await Dialogs.exportRadio(this);
        case 'import': return await this.importFile();
        case 'cread': return await this.cableRead();
        case 'cwrite': return await this.cableWrite();
        case 'firmware': return Firmware.open(this.def.fw, this.kind);
        case 'settings': return await RadioSettings.open(this);
      }
    } catch (e) { reportError(e); }
  }

  // canali da un file qualsiasi (3.7): adattati a questa radio e inseriti dove scegli
  async importFile() {
    let r;
    try { r = await api('radio/import', { kind: this.kind }, { busyText: 'Lettura del file…' }); } catch (e) { if (!e.cancel) reportError(e); return; }
    if (r.warnings && r.warnings.length) await alertDlg('Importa canali', `${r.desc}: alcune righe non sono state lette.`, r.warnings);
    toast(`${r.desc}: ${r.channels.length} canali da inserire.`);
    await Dialogs.insertInto(this, r.channels, r.kind);
  }

  async confirmDiscard() {
    if (!this.dirty) return true;
    return confirmDlg('Modifiche non salvate', `Le modifiche ai canali del ${this.def.name} non sono state salvate. Vuoi abbandonarle?`, 'Abbandona le modifiche');
  }

  async open(path) {
    if (!(await this.confirmDiscard())) return;
    const r = await api('radio/open', { radio: this.kind, path: path || '' }, { busyText: 'Apertura del file…' });
    const v = Views[r.kind] || App.openRadio(r.kind);
    if (r.kind !== this.kind) {
      if (!(await v.confirmDiscard())) return;
      toast(`Il file è della ${v.def.name}: aperto nella sua scheda.`);
    }
    v.load(r);
    App.showTab(r.kind);
    if (r.warnings && r.warnings.length) await alertDlg('File aperto con alcune segnalazioni', `${v.def.name}: controlla questi canali.`, r.warnings);
  }

  async newFile() {
    if (!(await this.confirmDiscard())) return;
    const r = await api('radio/new', { radio: this.kind });
    this.load(r);
    this.dirty = true;
    App.onDirty();
    this.render();
  }

  async save(saveAs) {
    if (!this.loaded) return;
    if (this.grid.editing) await this.grid.commitEdit();
    const r = await api('radio/save', { radio: this.kind, channels: this.chans, saveAs: !!saveAs }, { save: true, suggest: this.path, busyText: 'Salvataggio…' });
    this.path = r.path;
    this.dirty = false;
    this.fromRadio = false;
    App.onDirty();
    this.render();
    App.saveSession();
    if (r.warnings && r.warnings.length) await alertDlg('File salvato', `Salvato in ${r.path}. Segnalazioni:`, r.warnings);
    else toast('Salvato: ' + r.path);
  }

  // ------------------------------------------------------------ radio via cavo
  async cableRead() {
    if (!(await this.confirmDiscard())) return;
    const r = await Dialogs.cable(this, 'read');
    if (!r || !r.radio) return;
    this.load(r.radio);
    // dati letti dalla radio: non ancora salvati in un file
    this.path = '';
    this.fromRadio = true;
    this.dirty = true;
    App.onDirty();
    this.render();
    App.saveSession();
  }

  async cableWrite() {
    if (!this.loaded) return;
    if (this.grid.editing) await this.grid.commitEdit();
    if (!this.chans.some(Boolean)) { toast('Non ci sono canali da scrivere nella radio.', true); return; }
    await Dialogs.cable(this, 'write');
  }

  selected() {
    return this.grid.selectedIds().filter(i => typeof i === 'number').sort((a, b) => a - b);
  }
  cursorSlot() {
    const id = this.grid.curId();
    if (typeof id === 'number') return id;
    const s = this.selected();
    return s.length ? s[0] : this.range()[0];
  }
  setSel(ids, cur) { this.grid.setSelection(ids, cur != null ? cur : ids[0]); }

  freeSlots(from, n) {
    const [a, b] = this.range();
    const out = [];
    for (let i = Math.max(from, a); i < b && out.length < n; i++) if (!this.chans[i]) out.push(i);
    return out;
  }

  add() {
    const [a, b] = this.range();
    let slot = this.freeSlots(this.cursorSlot(), 1)[0];
    if (slot == null) slot = this.freeSlots(a, 1)[0];
    if (slot == null) { toast(this.def.zones && this.zone >= 0 ? 'Nessuna posizione libera in questa zona.' : 'Nessuna posizione libera.', true); return; }
    if (!this.showEmpty) { this.showEmpty = true; $('[data-empty]', this.page).checked = true; }
    this.setSel([slot], slot);
    this.grid.cur.c = 2;
    this.render();
    this.grid.ensureVisible();
    this.grid.startEdit();
  }

  insertRow() {
    const [a, b] = this.range();
    const c = this.cursorSlot();
    let gap = -1;
    for (let i = c; i < b; i++) if (!this.chans[i]) { gap = i; break; }
    if (gap < 0) { toast('Non ci sono posizioni libere dopo quella selezionata per far scorrere i canali.', true); return; }
    this.snapshot();
    for (let i = gap; i > c; i--) this.chans[i] = this.chans[i - 1];
    this.chans[c] = null;
    this.setSel([c], c);
    this.changed();
  }

  deleteSel(compactAfter) {
    const s = this.selected().filter(i => this.chans[i] || compactAfter);
    if (!s.length) { toast('Seleziona prima uno o più canali.'); return; }
    this.snapshot();
    if (!compactAfter) {
      for (const i of s) this.chans[i] = null;
    } else {
      const [a, b] = this.range();
      const set = new Set(s);
      const first = s[0];
      const rest = [];
      for (let i = first; i < b; i++) if (!set.has(i)) rest.push(this.chans[i]);
      for (let i = first; i < b; i++) this.chans[i] = i - first < rest.length ? rest[i - first] : null;
    }
    this.setSel([s[0]], s[0]);
    this.changed();
  }

  move(dir) {
    const s = this.selected();
    if (!s.length) return;
    const [a, b] = this.range();
    if (dir < 0 && s[0] - 1 < a) return;
    if (dir > 0 && s[s.length - 1] + 1 >= b) return;
    this.snapshot();
    const order = dir < 0 ? s : s.slice().reverse();
    for (const i of order) {
      const j = i + dir;
      const t = this.chans[j]; this.chans[j] = this.chans[i]; this.chans[i] = t;
    }
    const ns = s.map(i => i + dir);
    this.setSel(ns, this.cursorSlot() + dir);
    this.changed();
    this.grid.ensureVisible();
  }

  sortKey(key) {
    if (key === 'name') return (x, y) => (x.name || '').localeCompare(y.name || '', 'it') || x.rx - y.rx;
    if (key === 'tone') return (x, y) => (parseFloat(x.txTone) || 0) - (parseFloat(y.txTone) || 0) || x.rx - y.rx;
    return (x, y) => x.rx - y.rx || (x.name || '').localeCompare(y.name || '', 'it');
  }

  sort(key) {
    if (!this.loaded) return;
    const s = this.selected().filter(i => this.chans[i]);
    this.snapshot();
    if (s.length >= 2) {
      const items = s.map(i => this.chans[i]).sort(this.sortKey(key));
      s.forEach((pos, k) => this.chans[pos] = items[k]);
      toast(`Ordinati ${s.length} canali selezionati (le posizioni occupate restano le stesse).`);
    } else {
      const [a, b] = this.range();
      const items = this.chans.slice(a, b).filter(Boolean).sort(this.sortKey(key));
      for (let i = a; i < b; i++) this.chans[i] = items[i - a] || null;
      toast(this.def.zones && this.zone >= 0 ? `Zona ${this.zone + 1} ordinata e compattata.` : 'Canali ordinati e compattati.');
    }
    this.changed();
  }

  compact() {
    const [a, b] = this.range();
    const items = this.chans.slice(a, b).filter(Boolean);
    this.snapshot();
    for (let i = a; i < b; i++) this.chans[i] = items[i - a] || null;
    this.changed();
  }

  copy() {
    const s = this.selected().filter(i => this.chans[i]);
    if (!s.length) { toast('Seleziona prima i canali da copiare.'); return false; }
    Clipboard.kind = this.kind;
    Clipboard.items = s.map(i => clone(this.chans[i]));
    const tsv = ['Nome\tRX (MHz)\tTX (MHz)\tTono TX\tTono RX\tPotenza\tBanda']
      .concat(Clipboard.items.map(c => [c.name, fmtMHz(c.rx), fmtMHz(c.tx), c.txTone, c.rxTone, POWERS[c.power], c.narrow ? 'Stretta' : 'Larga'].join('\t'))).join('\r\n');
    try { navigator.clipboard && navigator.clipboard.writeText(tsv).catch(() => {}); } catch (e) { /* facoltativo */ }
    toast(`Copiati ${s.length} canali. Seleziona la posizione di destinazione e premi Incolla (Ctrl+V).`);
    return true;
  }

  cut() {
    if (!this.copy()) return;
    this.deleteSel(false);
  }

  // prepara canali provenienti da altra radio o archivio
  async adaptFrom(kind, items, positions) {
    if (kind === this.kind) return { channels: items.map(clone), warnings: [] };
    return api('adapt', { to: this.kind, channels: items, pos: positions || [] });
  }

  // posiziona items a partire dalla cella attiva, chiedendo cosa fare se occupato
  async placeItems(items, srcKind, startSlot) {
    const [a, b] = this.range();
    const start = startSlot != null ? startSlot : this.cursorSlot();
    const n = items.length;
    const targets = [];
    for (let i = start; i < start + n; i++) targets.push(i);
    const fits = start + n <= b;
    const occupied = fits ? targets.filter(i => this.chans[i]).length : 0;
    let mode = 'over';
    if (!fits || occupied) {
      const where = this.def.zones && this.zone >= 0 ? 'in questa zona' : '';
      const choices = [{ id: 'cancel', label: 'Annulla' }, { id: 'free', label: 'Usa solo le posizioni libere', primary: true }, { id: 'insert', label: 'Inserisci spostando in basso' }];
      if (fits) choices.push({ id: 'over', label: `Sovrascrivi ${occupied} canali` });
      const msg = fits ? `${occupied} delle ${n} posizioni a partire dalla ${this.posText(start)} sono già occupate.`
        : `Dalla posizione ${this.posText(start)} non ci sono ${n} posizioni ${where} fino alla fine.`;
      mode = await chooseDlg('Posizioni occupate', msg + '\nCome vuoi procedere?', choices);
      if (!mode || mode === 'cancel') return null;
    }
    let pos;
    if (mode === 'over') pos = targets;
    else if (mode === 'free') {
      pos = this.freeSlots(start, n);
      if (pos.length < n) {
        const tot = this.freeSlots(a, n * 2).length;
        throw new Error(`Posizioni libere insufficienti ${this.def.zones && this.zone >= 0 ? 'in questa zona' : ''} dalla posizione ${this.posText(start)}: ne servono ${n}, ne restano ${pos.length}` + (tot > pos.length ? ' (altre sono libere più in alto)' : '') + '.');
      }
    } else {
      // inserimento: gli esistenti scorrono in basso occupando i vuoti
      const seq = this.chans.slice(start, b);
      const merged = new Array(n).fill(null).concat(seq);
      // come «Inserisci riga»: i canali scendono fino alle prime posizioni vuote
      let excess = merged.length - (b - start);
      for (let k = n; k < merged.length && excess > 0;) {
        if (!merged[k]) { merged.splice(k, 1); excess--; } else k++;
      }
      if (excess > 0) throw new Error('Spazio insufficiente per inserire spostando in basso: libera qualche posizione.');
      pos = targets;
      const res = await this.adaptFrom(srcKind, items, pos);
      this.snapshot();
      for (let k = 0; k < merged.length; k++) this.chans[start + k] = merged[k];
      pos.forEach((p, k) => this.chans[p] = res.channels[k] || null);
      return { pos, warnings: res.warnings };
    }
    const res = await this.adaptFrom(srcKind, items, pos);
    this.snapshot();
    pos.forEach((p, k) => this.chans[p] = res.channels[k] || this.chans[p]);
    return { pos, warnings: res.warnings };
  }

  async paste() {
    if (!this.loaded) return;
    if (!Clipboard.items.length) { toast('Niente da incollare: prima copia dei canali o dei ponti.'); return; }
    const r = await this.placeItems(Clipboard.items, Clipboard.kind);
    if (!r) return;
    this.setSel(r.pos, r.pos[0]);
    this.changed();
    this.grid.ensureVisible();
    if (r.warnings && r.warnings.length) await alertDlg('Canali incollati con alcune modifiche', `Incollati ${r.pos.length} canali nel ${this.def.name}.`, r.warnings);
    else toast(`Incollati ${r.pos.length} canali.`);
  }

  moveToZone(z) {
    const s = this.selected().filter(i => this.chans[i]);
    if (!s.length) { toast('Seleziona prima i canali da spostare.'); return; }
    const d = this.def;
    const free = [];
    for (let i = z * d.perZone; i < (z + 1) * d.perZone && free.length < s.length; i++) if (!this.chans[i] || s.includes(i)) free.push(i);
    if (free.length < s.length) { toast(`Nella zona ${z + 1} ci sono solo ${free.length} posizioni libere.`, true); return; }
    this.snapshot();
    const items = s.map(i => this.chans[i]);
    s.forEach(i => this.chans[i] = null);
    free.forEach((p, k) => this.chans[p] = items[k]);
    this.zone = z;
    this.setSel(free, free[0]);
    this.changed();
    toast(`Spostati ${s.length} canali nella zona ${z + 1}.`);
  }

  copyToMenu(btn) {
    const s = this.selected().filter(i => this.chans[i]);
    if (!s.length) { toast('Seleziona prima i canali da copiare.'); return; }
    const items = this.others().map(v => ({ label: v.def.name + (v.loaded ? '' : ' (file da aprire)'), run: () => this.copyTo(v) }));
    items.push({ label: 'Altra radio dalla galleria dei modelli…', run: () => { Clipboard.kind = this.kind; Clipboard.items = s.map(i => clone(this.chans[i])); App.showTab('gal'); toast('Canali copiati: apri la scheda della radio dalla galleria e usa Incolla (Ctrl+V).'); } });
    showMenuAt(btn, items);
  }

  async copyTo(target) {
    const s = this.selected().filter(i => this.chans[i]);
    if (!s.length) { toast('Seleziona prima i canali da copiare.'); return; }
    await Dialogs.insertInto(target, s.map(i => clone(this.chans[i])), this.kind);
  }

  async toArchive() {
    const s = this.selected().filter(i => this.chans[i]);
    if (!s.length) { toast('Seleziona prima i canali da aggiungere all\'archivio.'); return; }
    Archive.addFromChannels(s.map(i => this.chans[i]), this.def.name);
  }

  // ------------------------------------------------------------ modifica celle
  async edit(row, col, val) {
    const d = this.def;
    const i = row.id;
    const old = this.chans[i];
    let ch = old ? clone(old) : null;
    val = String(val);
    switch (col.key) {
      case 'rx': {
        if (val.trim() === '' && !old) return null;
        const hz = parseFreq(val);
        if (hz == null) return 'Frequenza non valida (es. 145.6000 oppure 430.0250)';
        if (hz >= 1e9 && !d.generic) return 'Frequenza troppo alta per i file della radio';
        if (!ch) ch = { rx: hz, tx: hz, txTone: '', rxTone: '', power: d.powers[0] || 0, narrow: false, scan: true, name: '' };
        else if (ch.txOff) { ch.rx = hz; ch.tx = hz; }
        else { const off = ch.tx - ch.rx; ch.rx = hz; ch.tx = Math.abs(off) > 50e6 ? ch.tx : hz + off; }
        break;
      }
      case 'shift': {
        if (!ch) return null;
        if (val.trim().toLowerCase() === 'split') return 'Per un TX in un\'altra banda scrivi la frequenza nella colonna TX';
        const off = parseShift(val);
        if (off == null) return 'Shift non valido (es. -0.6, +1.6, +5, 0)';
        ch.tx = ch.rx + off;
        if (ch.tx <= 0 || ch.tx >= 1e9) return 'Shift non valido per questa frequenza';
        break;
      }
      case 'tx': {
        if (!ch) return null;
        if (/^\s*(off|no|-|—)\s*$/i.test(val)) {
          if (!d.generic || d.noTxOff) return 'Questa radio non permette canali di sola ricezione';
          ch.txOff = true; ch.tx = ch.rx;
          break;
        }
        delete ch.txOff;
        const hz = val.trim() === '' ? ch.rx : parseFreq(val);
        if (hz == null) return 'Frequenza non valida';
        if (hz >= 1e9) return 'Frequenza troppo alta per i file della radio';
        ch.tx = hz;
        break;
      }
      case 'txTone': case 'rxTone': {
        if (!ch) return null;
        if (ch.dig) return 'Canale digitale: i subtoni si usano solo sui canali analogici (cambia prima il modo)';
        if (val === fmtTone(ch[col.key])) return null;
        const t = normTone(val);
        if (t == null) return 'Tono non valido: scrivi un subtono CTCSS standard (es. 67.0, 88.5), un codice DCS (es. D023N, D023I invertito) oppure lascia vuoto per OFF';
        if (t.startsWith('D') && this.kind === 'baj' && !(App.init.dcsBaj || []).includes(t)) return 'Codice DCS non disponibile sulla Bajeton 8300 (sono ammessi i 104 codici standard, es. D023N, D754I)';
        if (t.startsWith('D') && d.generic && !d.dcs.has(t)) return `Codice DCS non disponibile sulla ${d.name}`;
        ch[col.key] = t;
        break;
      }
      case 'name': {
        if (!ch) return null;
        if (d.generic && !d.nameMax) return 'Questa radio non memorizza i nomi dei canali';
        let n = d.generic ? cleanNameFor(d, val) : asciiName(val);
        if (d.generic && n !== asciiName(val).trimEnd()) toast(`Nome adattato alla radio: «${n}» (max ${d.nameMax} caratteri ammessi)`);
        if (n.length > d.nameMax) { n = n.slice(0, d.nameMax); toast(`Nome accorciato a ${d.nameMax} caratteri.`); }
        ch.name = n;
        break;
      }
      case 'power': if (!ch) return null; ch.power = +val; break;
      case 'mode':
        if (!ch) return null;
        if (val === '1') { ch.dig = true; ch.txTone = ''; ch.rxTone = ''; } else delete ch.dig;
        break;
      case 'narrow': if (!ch) return null; ch.narrow = val === '1'; break;
      case 'scan': if (!ch) return null; ch.scan = val === '1'; break;
      default: return null;
    }
    if (JSON.stringify(ch) === JSON.stringify(old)) return null;
    this.snapshot();
    this.chans[i] = ch;
    this.dirty = true;
    App.onDirty();
    this.render();
    App.saveSession();
    return null;
  }

  onKey(e) {
    const k = e.key, ctrl = e.ctrlKey || e.metaKey;
    if (!this.loaded) return false;
    if (k === 'Delete') { this.deleteSel(e.shiftKey); return true; }
    if (k === 'Insert') { this.insertRow(); return true; }
    if (ctrl && k.toLowerCase() === 'c') { this.copy(); return true; }
    if (ctrl && k.toLowerCase() === 'x') { this.cut(); return true; }
    if (ctrl && k.toLowerCase() === 'v') { this.paste().catch(reportError); return true; }
    if (ctrl && k.toLowerCase() === 'z' && !e.shiftKey) { this.undo(); return true; }
    if (ctrl && (k.toLowerCase() === 'y' || (k.toLowerCase() === 'z' && e.shiftKey))) { this.redo(); return true; }
    if (e.altKey && k === 'ArrowUp') { this.move(-1); return true; }
    if (e.altKey && k === 'ArrowDown') { this.move(1); return true; }
    return false;
  }

  contextMenu(e) {
    if (!this.loaded) return;
    e.preventDefault();
    $$('.ctxmenu').forEach(m => m.remove());
    const items = [
      ['Copia', () => this.copy()], ['Taglia', () => this.cut()], ['Incolla', () => this.paste()],
      ['Cancella', () => this.deleteSel(false)], ['Elimina e compatta', () => this.deleteSel(true)], ['Inserisci riga', () => this.insertRow()],
      ['Sposta su', () => this.move(-1)], ['Sposta giù', () => this.move(1)],
      ...this.others().map(v => [`Copia in ${v.def.name}…`, () => this.copyTo(v)]), ['Aggiungi all\'archivio', () => this.toArchive()],
    ];
    const m = h('<div class="menu open ctxmenu" style="position:fixed"></div>');
    for (const [label, fn] of items) {
      const b = h(`<button>${esc(label)}</button>`);
      b.onclick = () => { m.remove(); Promise.resolve().then(fn).catch(reportError); };
      m.appendChild(b);
    }
    document.body.appendChild(m);
    const x = Math.min(e.clientX, innerWidth - 240), y = Math.min(e.clientY, innerHeight - m.offsetHeight - 8);
    m.style.left = x + 'px'; m.style.top = y + 'px';
    setTimeout(() => document.addEventListener('mousedown', function f(ev) { if (!m.contains(ev.target)) { m.remove(); document.removeEventListener('mousedown', f); } }), 0);
  }

  sessionData() {
    return { loaded: this.loaded, path: this.path, chans: this.chans, bands: this.bands, dirty: this.dirty, zone: this.zone, fromRadio: !!this.fromRadio };
  }
  restore(s) {
    if (!s || !s.loaded) return;
    this.loaded = true; this.path = s.path; this.fromRadio = !!s.fromRadio; this.chans = s.chans; this.bands = s.bands || []; this.dirty = s.dirty; this.zone = s.zone || 0;
    this.render();
  }
}
