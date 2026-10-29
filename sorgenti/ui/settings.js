// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Finestra delle impostazioni della radio (menu, display, VOX, DTMF…): i valori
// si leggono dalla memoria della radio aperta nella scheda; «Applica» li scrive in
// quella memoria, che viene salvata nel file e scritta nella radio con i canali.
// Ogni impostazione ha l'icona ⓘ (descrizione: passaggio del mouse o clic) e, se
// delicata, l'icona ⚠ con l'avviso, che va confermato prima di applicare.

const RadioSettings = {
  async open(view) {
    let r;
    try {
      r = await api('radio/settings', { kind: view.kind }, { busyText: 'Lettura delle impostazioni…' });
    } catch (e) { reportError(e); return; }
    if (!r.available) { await alertDlg('Impostazioni della radio', r.reason || 'Impostazioni non disponibili.'); return; }
    let groups = r.groups;
    const changes = {};
    const src = r.source === 'CHIRP'
      ? 'Tabella ricavata dal codice di CHIRP per questa radio e verificata con esso.'
      : (r.source ? 'Tabella ricavata dal ' + r.source + '.' : '');
    const m = openModal(`<div class="dialog big setdlg"><header>⚙ Impostazioni — ${esc(view.def.name)}
        <input type="search" data-q placeholder="Cerca un'impostazione…" style="margin-left:auto;width:260px"></header>
      <div class="body flush"><div class="setwrap">
        <nav class="setnav" data-nav></nav>
        <div class="setlist" data-list></div>
      </div></div>
      <footer><span class="left note" data-foot>${esc(src)} Le modifiche vengono scritte nella memoria della scheda: salva il file o scrivi nella radio per conservarle.</span>
        <button class="btn" data-reset title="Annulla le modifiche non ancora applicate">Annulla modifiche</button>
        <button class="btn" data-no>Chiudi</button><button class="btn primary" data-ok disabled>Applica</button></footer></div>`,
      { onEsc: () => close() });
    const el = m.el;
    const close = async () => {
      if (Object.keys(changes).length && !(await confirmDlg('Modifiche non applicate', `Hai cambiato ${Object.keys(changes).length} impostazioni senza applicarle. Vuoi chiudere comunque?`, 'Chiudi senza applicare'))) return;
      m.close();
    };
    const list = $('[data-list]', el), nav = $('[data-nav]', el);

    const control = (it) => {
      const cur = Object.prototype.hasOwnProperty.call(changes, it.key) ? changes[it.key] : it.value;
      const dis = it.ro ? 'disabled' : '';
      switch (it.type) {
        case 'bool':
          return `<label class="setbool"><input type="checkbox" data-k="${esc(it.key)}" ${cur === 'true' ? 'checked' : ''} ${dis}> ${cur === 'true' ? 'Attivo' : 'Spento'}</label>`;
        case 'list':
          return `<select data-k="${esc(it.key)}" ${dis}>${it.unknown && !Object.prototype.hasOwnProperty.call(changes, it.key) ? '<option value="" selected>(valore non previsto)</option>' : ''}${it.options.map((o, i) =>
            `<option value="${esc(o)}" ${o === cur ? 'selected' : ''}>${esc(it.optLabels ? it.optLabels[i] : o)}</option>`).join('')}</select>`;
        case 'int': case 'float': case 'numstr':
          return `<input type="number" data-k="${esc(it.key)}" value="${esc(cur)}" ${it.type !== 'numstr' ? `min="${it.min}" max="${it.max}"` : ''} step="${it.type === 'int' ? (it.step || 1) : (it.dec ? Math.pow(10, -it.dec) : 'any')}" ${dis} style="width:140px">
            ${it.type !== 'numstr' ? `<span class="note">${esc(String(it.min))}–${esc(String(it.max))}</span>` : ''}`;
        case 'str':
          return `<input type="text" data-k="${esc(it.key)}" value="${esc(cur)}" maxlength="${it.maxlen || 32}" ${dis} style="width:220px" spellcheck="false">
            <span class="note">max ${it.maxlen} caratteri</span>`;
        default:
          return `<span class="setinfo">${esc(cur)}</span>`;
      }
    };
    const row = (it) => `<div class="setrow ${it.critical ? 'crit' : ''} ${Object.prototype.hasOwnProperty.call(changes, it.key) ? 'changed' : ''}" data-row="${esc(it.key)}">
        <div class="setlabel">
          <span class="seticon" tabindex="0" title="${esc(it.desc)}${it.orig && it.orig !== it.label ? '\n(' + esc(it.orig) + ')' : ''}" data-info="${esc(it.key)}">ⓘ</span>
          ${it.critical ? `<span class="setwarn" title="${esc(it.warn)}">⚠</span>` : ''}
          <span>${esc(it.label)}</span>${it.ro ? ' <span class="note">(sola lettura)</span>' : ''}
        </div>
        <div class="setctl">${control(it)}${it.unknown && it.type !== 'list' ? ' <span class="note">valore in memoria non previsto</span>' : ''}</div>
        <div class="setdesc" data-desc="${esc(it.key)}" style="display:none">${esc(it.desc)}${it.critical ? `<div class="setalert">⚠ ${esc(it.warn)}</div>` : ''}</div>
      </div>`;
    const render = () => {
      const q = ($('[data-q]', el).value || '').trim().toLowerCase();
      const match = it => !q || (it.label + ' ' + it.desc + ' ' + (it.orig || '')).toLowerCase().includes(q);
      const gs = groups.map(g => ({ name: g.name, items: g.items.filter(match) })).filter(g => g.items.length);
      nav.innerHTML = gs.map((g, i) => `<a href="#" data-g="${i}">${esc(g.name)} <span class="note">${g.items.length}</span></a>`).join('');
      list.innerHTML = gs.length ? gs.map((g, i) => `<section class="setgroup" data-sec="${i}"><h4>${esc(g.name)}</h4>${g.items.map(row).join('')}</section>`).join('')
        : '<p class="note" style="padding:16px">Nessuna impostazione trovata.</p>';
      $$('[data-g]', nav).forEach(a => a.onclick = (e) => { e.preventDefault(); const s = $(`[data-sec="${a.dataset.g}"]`, list); if (s) s.scrollIntoView({ block: 'start' }); });
      $$('[data-info]', list).forEach(ic => {
        const show = () => { const d = $(`[data-desc="${CSS.escape(ic.dataset.info)}"]`, list); d.style.display = d.style.display === 'none' ? '' : 'none'; };
        ic.onclick = show;
        ic.onkeydown = (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); show(); } };
      });
      $$('[data-k]', list).forEach(inp => {
        const ev = inp.type === 'checkbox' || inp.tagName === 'SELECT' ? 'change' : 'input';
        inp.addEventListener(ev, () => {
          const it = byKey[inp.dataset.k];
          const v = inp.type === 'checkbox' ? String(inp.checked) : inp.value;
          if (v === it.value) delete changes[it.key]; else changes[it.key] = v;
          const rw = inp.closest('.setrow');
          rw.classList.toggle('changed', Object.prototype.hasOwnProperty.call(changes, it.key));
          if (inp.type === 'checkbox') inp.parentElement.lastChild.textContent = ' ' + (inp.checked ? 'Attivo' : 'Spento');
          updFoot();
        });
      });
      updFoot();
    };
    let byKey = {};
    const index = () => { byKey = {}; groups.forEach(g => g.items.forEach(it => { byKey[it.key] = it; })); };
    const updFoot = () => {
      const n = Object.keys(changes).length;
      $('[data-ok]', el).disabled = !n;
      $('[data-ok]', el).textContent = n ? `Applica (${n})` : 'Applica';
    };
    index();
    render();
    $('[data-q]', el).addEventListener('input', render);
    $('[data-no]', el).onclick = () => close();
    $('[data-reset]', el).onclick = () => { for (const k of Object.keys(changes)) delete changes[k]; render(); };
    $('[data-ok]', el).onclick = async () => {
      const keys = Object.keys(changes);
      if (!keys.length) return;
      const crit = keys.map(k => byKey[k]).filter(it => it && it.critical);
      const lines = keys.map(k => { const it = byKey[k]; return `${it.label}: ${fmtVal(it, it.value)} → ${fmtVal(it, changes[k])}`; });
      if (crit.length) {
        const ok = await confirmDlg('Attenzione: impostazioni delicate',
          'Stai modificando impostazioni che possono rendere la radio difficile da usare o non conforme alle norme:\n\n' +
          crit.map(it => '⚠ ' + it.label + ': ' + it.warn).join('\n\n') + '\n\nVuoi applicare comunque le modifiche?', 'Applica comunque');
        if (!ok) return;
      }
      try {
        const res = await api('radio/settings/set', { kind: view.kind, values: changes }, { busyText: 'Applicazione delle impostazioni…' });
        groups = res.groups;
        index();
        for (const k of keys) delete changes[k];
        render();
        view.dirty = true;
        App.onDirty();
        if (res.single !== undefined) view.setLayout(res.single);
        view.render();
        toast(`${keys.length} impostazioni applicate alla memoria della radio: salva il file o scrivi nella radio per conservarle.`);
        if (lines.length > 1) console.log(lines.join('\n'));
      } catch (e) { reportError(e); }
    };
    const fmtVal = (it, v) => it.type === 'bool' ? (v === 'true' ? 'attivo' : 'spento')
      : it.type === 'list' && it.optLabels ? (it.optLabels[it.options.indexOf(v)] || v) : (v === '' ? '(vuoto)' : v);
  },
};
