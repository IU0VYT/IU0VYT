'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda «Converti» (3.7): i canali di qualsiasi file o scheda aperta (K67 .xlc,
// Bajeton .dat, immagini .img di tutte le radio anche di CHIRP, CSV di CHIRP,
// tabelle Excel/CSV) scritti nel formato di qualsiasi altra radio o in un formato
// universale, adattati alle regole della radio di destinazione.

const Converter = {
  src: null,      // { name, desc, kind, channels, warnings, fromTab }
  targets: [],
  target: '',

  // apre la scheda con i canali di una scheda radio già pronti come partenza
  fromView(view) {
    this.src = { name: view.path ? view.path.split(/[\\/]/).pop() : view.def.name.replace(/\s+/g, '_'), desc: 'Scheda aperta: ' + view.def.name,
      kind: view.kind, channels: view.chans.map(c => c || null), warnings: [], fromTab: view.kind };
    App.showTab('conv');
  },

  async render(page) {
    page = page || $('#page-conv');
    if (!this.targets.length || this._stale) {
      try { this.targets = (await api('convert/targets', {}, { quiet: true })).targets; } catch (e) { reportError(e); }
      this._stale = false;
    } else {
      // le schede aperte cambiano: aggiorna l'indicazione «scheda aperta»
      api('convert/targets', {}, { quiet: true }).then(r => { this.targets = r.targets; this.renderTargetInfo(page); }).catch(() => {});
    }
    const loaded = Object.values(Views).filter(v => v.loaded);
    const s = this.src;
    const n = s ? s.channels.filter(Boolean).length : 0;
    page.innerHTML = `<div class="infopage"><div class="infocard convcard">
      <div class="infohead"><div style="font-size:44px">🔁</div><div><h1>Converti</h1>
        <div class="note">I canali di qualsiasi radio o file, nel formato di qualsiasi altra radio: le frequenze, i toni, le potenze, i nomi e la
        larghezza vengono adattati alle regole della radio di destinazione, con l'elenco delle modifiche.</div></div></div>
      <h3>1. Canali di partenza</h3>
      <div class="convrow"><button class="btn primary" data-open>📂 Scegli un file…</button>
        <span class="note">K67 (.xlc), Bajeton (.dat v1.0.8 o v1.0.9), immagini .img di tutte le radio (anche salvate con CHIRP), CSV di CHIRP, tabelle Excel o CSV</span></div>
      ${loaded.length ? `<div class="convrow"><span>oppure i canali di una scheda aperta:</span>
        <select data-tab><option value="">—</option>${loaded.map(v => `<option value="${esc(v.kind)}" ${s && s.fromTab === v.kind ? 'selected' : ''}>${esc(v.def.name)} (${v.chans.filter(Boolean).length} canali)</option>`).join('')}</select></div>` : ''}
      <div class="convsrc" data-src>${s ? `<b>${esc(s.desc)}</b>${s.fromTab ? '' : ' — ' + esc(s.name)}: <b>${n}</b> ${n === 1 ? 'canale' : 'canali'}
        ${s.warnings && s.warnings.length ? `<details><summary>${s.warnings.length} segnalazioni nella lettura</summary><ul class="warnlist">${s.warnings.map(w => '<li>' + esc(w) + '</li>').join('')}</ul></details>` : ''}`
        : '<span class="note">Nessun file scelto.</span>'}</div>
      <h3>2. Formato di destinazione</h3>
      <div class="convrow"><input type="search" data-q placeholder="Cerca una radio o un formato…" style="width:260px">
        <select data-target style="flex:1;min-width:300px"></select></div>
      <div data-tinfo></div>
      <div class="convrow"><label><input type="checkbox" data-pack> compatta i canali (togli le posizioni vuote)</label>
        <span class="note">altrimenti ogni canale resta nella stessa posizione, se la radio di destinazione ce l'ha</span></div>
      <h3>3. Conversione</h3>
      <div class="convrow"><button class="btn primary" data-run ${n ? '' : 'disabled'}>🔁 Converti e salva…</button></div>
      <div data-res></div>
    </div></div>`;
    const sel = $('[data-target]', page);
    const fill = () => {
      const q = ($('[data-q]', page).value || '').trim().toLowerCase();
      const groups = {};
      for (const t of this.targets) {
        if (q && !t.name.toLowerCase().includes(q) && !t.id.includes(q)) continue;
        (groups[t.group] = groups[t.group] || []).push(t);
      }
      sel.innerHTML = Object.keys(groups).map(g => `<optgroup label="${esc(g)}">${groups[g].map(t =>
        `<option value="${esc(t.id)}" ${t.id === this.target ? 'selected' : ''}>${esc(t.name)}${t.tab ? ' — scheda aperta' : ''}</option>`).join('')}</optgroup>`).join('');
      if (!sel.value && sel.options.length) sel.selectedIndex = 0;
      this.target = sel.value;
      this.renderTargetInfo(page);
    };
    fill();
    $('[data-q]', page).addEventListener('input', fill);
    sel.onchange = () => { this.target = sel.value; this.renderTargetInfo(page); };
    $('[data-open]', page).onclick = async () => {
      try {
        const r = await api('convert/open', {}, { busyText: 'Lettura del file…' });
        this.src = { name: r.name, desc: r.desc, kind: r.kind, channels: r.channels, warnings: r.warnings || [] };
        this.render(page);
      } catch (e) { reportError(e); }
    };
    const tabSel = $('[data-tab]', page);
    if (tabSel) tabSel.onchange = () => { const v = Views[tabSel.value]; if (v) this.fromView(v); this.render(page); };
    $('[data-run]', page).onclick = () => this.run(page).catch(reportError);
  },

  targetById(id) { return this.targets.find(t => t.id === id); },

  renderTargetInfo(page) {
    const box = $('[data-tinfo]', page);
    if (!box) return;
    const t = this.targetById(this.target);
    if (!t) { box.innerHTML = ''; return; }
    const slots = t.slots ? `<span class="note">La radio ha ${t.slots} posizioni.</span>` : '';
    if (t.needsBase) {
      box.innerHTML = `<div class="convbase"><p>Per creare l'immagine <b>.img</b> serve un'<b>immagine di questa radio</b>, cioè la copia completa della
        sua memoria: si ottiene leggendo la radio con il cavo («Leggi dalla radio») oppure aprendo un file .img salvato prima da qui o da CHIRP.
        Contiene anche impostazioni e taratura della radio, che non si possono inventare: si cambiano solo i canali. ${slots}</p>
        <label><input type="radio" name="base" value="tab" ${t.tab ? 'checked' : 'disabled'}> immagine della scheda aperta della radio ${t.tab ? '' : '<span class="note">(la scheda non ha ancora dati)</span>'}</label><br>
        <label><input type="radio" name="base" value="file" ${t.tab ? '' : 'checked'}> un file .img della radio (salvato da qui o da CHIRP)…</label>
        ${t.tab ? '' : '<div style="margin-top:6px"><button class="btn" data-readbase>📻 Leggi adesso la radio collegata…</button> <span class="note">poi torna qui: l\'immagine letta sarà la base</span></div>'}</div>`;
      const rb = $('[data-readbase]', box);
      if (rb) rb.onclick = async () => {
        try {
          const v = App.openRadio(t.id, true);
          await v.cableRead();
          if (v.loaded) { this._stale = true; App.showTab('conv'); }
        } catch (e) { reportError(e); }
      };
    } else if (['k67', 'baj-1.0.8', 'baj-1.0.9'].includes(t.id)) {
      box.innerHTML = `<div class="convbase">
        <label><input type="radio" name="base" value="" ${t.tab ? '' : 'checked'}> file nuovo (impostazioni predefinite del programma del costruttore)</label><br>
        ${t.tab ? `<label><input type="radio" name="base" value="tab" checked> impostazioni del file aperto nella scheda della radio (consigliato)</label><br>` : ''}
        <label><input type="radio" name="base" value="file"> impostazioni di un altro file della radio…</label> ${slots}</div>`;
    } else {
      box.innerHTML = `<p class="note">${t.id === 'chirp'
        ? 'Formato universale: si apre con CHIRP (File → Importa) per qualsiasi radio e con «📥 Importa…» di questo programma.'
        : 'Tabella con tutti i dati dei canali, modificabile con Excel; si reimporta con «📥 Importa…» o con questo convertitore.'}</p>`;
    }
  },

  async run(page) {
    const s = this.src;
    if (!s) return;
    const t = this.targetById(this.target);
    if (!t) { toast('Scegli il formato di destinazione.'); return; }
    const baseEl = $('input[name=base]:checked', page);
    const base = baseEl ? baseEl.value : '';
    if (t.needsBase && !base) { toast('Scegli l\'immagine della radio da usare come base.'); return; }
    const r = await api('convert/run', { channels: s.channels, target: t.id, base, pack: $('[data-pack]', page).checked, name: s.name },
      { busyText: 'Conversione…', save: true, suggest: '' });
    const kind = t.id.startsWith('baj-') ? 'baj' : t.id;
    const canOpen = App.radioDef(kind) && !['chirp', 'xlsx', 'csv'].includes(t.id);
    $('[data-res]', page).innerHTML = `<div class="convok"><span>✔ Salvato <b>${esc(r.path)}</b>: ${r.count} ${r.count === 1 ? 'canale' : 'canali'}.</span>
      ${canOpen ? '<button class="btn" data-openres>Apri nella scheda della radio</button>' : ''}</div>
      ${r.warnings && r.warnings.length ? `<p>Adattamenti per la radio di destinazione (${r.warnings.length}):</p><ul class="warnlist">${r.warnings.map(w => '<li>' + esc(w) + '</li>').join('')}</ul>`
        : '<p class="note">Nessun adattamento necessario.</p>'}`;
    const ob = $('[data-openres]', page);
    if (ob) ob.onclick = async () => { try { await App.openRadio(kind, true).open(r.path); } catch (e) { reportError(e); } };
    toast('File convertito salvato.');
  },
};
