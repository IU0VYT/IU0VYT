'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Finestre: inserimento in una radio, importazione, esportazione, nomi.

const MODELLI = [
  { v: '{LOC}', l: 'Località (es. M.ALTO)' },
  { v: '{SIGLA} {LOC}', l: 'Sigla + località (es. RV M.ALTO)' },
  { v: '{NOMINATIVO} {LOC}', l: 'Nominativo + località (es. IR0XX M.ALTO)' },
  { v: '{NOMINATIVO}', l: 'Solo nominativo (es. IR0XX)' },
  { v: '{LOC} {SIGLA}', l: 'Località + sigla (es. M.ALTO RV)' },
];

function modelloField(value) {
  return `<input type="text" data-modello list="modelli-list" value="${esc(value || '{LOC}')}" style="width:220px">
    <datalist id="modelli-list">${MODELLI.map(m => `<option value="${esc(m.v)}">${esc(m.l)}</option>`).join('')}</datalist>
    <span class="note">Segnaposto: {LOC} località, {SIGLA} sigla (R0, RU12…), {NOMINATIVO}, {PROV}. Max 16 caratteri (12 sulla Bajeton, accorciati automaticamente).</span>`;
}

function normKeyJS(s) {
  return String(s || '').toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/[^a-z0-9]/g, '');
}
function pulisciLoc(s) { return String(s || '').replace(/\s+/g, ' ').replace(/\s*\([A-Za-z]{2,3}\)\s*$/, '').trim(); }
// stessa frequenza, stesso TX, stessa località e sigla/nominativo (o nome, se mancano: es. elenchi CHIRP)
function chiavePonte(p) {
  const loc = normKeyJS(pulisciLoc(p.localita)), id = normKeyJS((p.sigla || '') + (p.nominativo || ''));
  return [p.rx, p.tx, loc, id, loc || id ? '' : normKeyJS(p.nome)].join('|');
}

const Dialogs = {
  // ------------------------------------------------------------ scelta della radio
  // elenco di tutte le radio compatibili, una voce per modello (anche quando più
  // modelli hanno la stessa gestione) e, se la radio esiste in più versioni, una
  // voce per versione; diviso per marca e tipologia, in ordine alfabetico
  radioChoices() {
    const seen = new Set(), out = [];
    for (const c of App.catalog) {
      const drv = (c.drivers || []).filter(d => d === 'k67' || d === 'baj' || App.defs[d]);
      for (const d of drv) {
        const def = RADIO_DEF[d] || App.defs[d];
        const version = drv.length > 1 ? def.name.replace(/^(Baofeng|Bajeton|Radioddity)\s+/, '') : '';
        const key = c.brand + '|' + c.name + '|' + d;
        if (seen.has(key)) continue;
        seen.add(key);
        out.push({ kind: d, name: c.name, brand: c.brand, type: c.type, version, driver: def.name, id: c.id });
      }
    }
    const bOrder = b => b === 'Baofeng' ? 0 : b === 'Bajeton' ? 1 : 2;
    const tOrder = t => { const i = TYPE_ORDER.indexOf(t); return i < 0 ? 99 : i; };
    out.sort((a, b) => bOrder(a.brand) - bOrder(b.brand) || a.brand.localeCompare(b.brand) || tOrder(a.type) - tOrder(b.type) ||
      byModelName(a, b) || a.version.localeCompare(b.version, 'it', { numeric: true }));
    return out;
  },

  // restituisce { kind, label } della voce scelta (o null)
  pickRadio(title, text) {
    const all = Dialogs.radioChoices();
    return new Promise(res => {
      const m = openModal(`<div class="dialog mid"><header>${esc(title)}</header><div class="body">
        <p>${esc(text)}</p>
        <div class="row"><input type="search" data-q placeholder="Cerca il modello (es. UV-5R, K67, BJ-7700)…" style="flex:1" autofocus></div>
        <div class="radiopick" data-list></div>
        <p class="note">Le radio già aperte nel programma sono indicate con ●. Se la radio non è ancora stata letta, potrai leggerla via cavo o aprirne un file.
          Prima di scrivere viene controllato che i dati scelti siano compatibili con la radio.</p>
        </div><footer><button class="btn" data-no>Annulla</button></footer></div>`, { onEsc: () => { m.close(); res(null); } });
      const list = $('[data-list]', m.el);
      const draw = () => {
        const q = normKeyJS($('[data-q]', m.el).value);
        const its = all.filter(r => !q || normKeyJS(r.name + ' ' + r.version + ' ' + r.driver).includes(q));
        let html = '', brand = '', type = '';
        its.forEach((r, i) => {
          if (r.brand !== brand) { brand = r.brand; type = ''; html += `<div class="radiopick-brand">${esc(brand)}</div>`; }
          if (r.type !== type) { type = r.type; html += `<div class="radiopick-type">${esc(TYPE_GROUP[type] || type)}</div>`; }
          const open = Views[r.kind] && (Views[r.kind].loaded || r.kind === App.lastRadio);
          const same = normKeyJS(r.driver).includes(normKeyJS(r.name));
          const extra = r.version ? `versione ${r.version}` : same ? '' : `gestione: ${r.driver}`;
          html += `<button class="radiopick-item" data-i="${i}">${open ? '● ' : ''}${esc(r.name)}${extra ? ` <span class="note">(${esc(extra)})</span>` : ''}</button>`;
        });
        list.innerHTML = html || '<p class="note">Nessun modello trovato.</p>';
        $$('[data-i]', list).forEach(b => b.onclick = () => {
          const r = its[+b.dataset.i];
          m.close();
          res({ kind: r.kind, label: r.brand + ' ' + r.name + (r.version ? ' (' + r.version + ')' : ''), type: r.type });
        });
      };
      $('[data-q]', m.el).addEventListener('input', draw);
      $('[data-no]', m.el).onclick = () => { m.close(); res(null); };
      draw();
    });
  },

  // ------------------------------------------------------------ inserimento in una radio
  async insertInto(view, chans, srcKind) {
    if (!chans.length) return;
    if (!view.loaded) {
      const c = view.def.generic
        ? await chooseDlg(view.def.name, `Nella scheda ${view.def.name} non ci sono ancora canali: leggi prima la radio via cavo oppure apri un suo file .img.`,
          [{ id: 'no', label: 'Annulla' }, { id: 'open', label: 'Apri file .img…' }, { id: 'cread', label: 'Leggi dalla radio…', primary: true }])
        : await chooseDlg(view.def.name, `Nella scheda ${view.def.name} non c'è ancora un file aperto. Apri il file della radio (consigliato) oppure creane uno nuovo.`,
          [{ id: 'no', label: 'Annulla' }, { id: 'new', label: 'Nuovo file' }, { id: 'open', label: 'Apri file ' + view.def.ext + '…', primary: true }]);
      if (c === 'open') await view.open();
      else if (c === 'new') await view.newFile();
      else if (c === 'cread') await view.cableRead();
      if (!view.loaded) return;
    }
    const d = view.def;
    const n = chans.length;
    const zoneOpts = () => {
      let s = '';
      for (let z = 0; z < d.zones; z++) s += `<option value="${z}" ${z === Math.max(0, view.zone) ? 'selected' : ''}>Zona ${z + 1} — libere ${d.perZone - view.usedIn(z * d.perZone, (z + 1) * d.perZone)}</option>`;
      return s;
    };
    const m = openModal(`<div class="dialog mid"><header>Inserisci ${plural(n, 'canale', 'canali')} nel ${esc(d.name)}</header><div class="body">
      ${d.zones ? `<div class="row"><label class="l">Zona</label><select data-zone>${zoneOpts()}</select></div>` : ''}
      <div class="row"><label class="l">A partire dalla posizione</label><input type="number" data-start min="1" max="${d.zones ? d.perZone : d.slots}" value="1" style="width:90px">
        <span class="note">${d.zones ? `numero nella zona (1–${d.perZone})` : `1–${d.slots}` + (d.first === 0 ? ' (la radio numera i canali da 0: posizione 1 = canale 0)' : '')}</span></div>
      <div class="row radiolist"><label class="l">Modo</label><div>
        <label><input type="radio" name="mode" value="free" checked> solo nelle posizioni libere ${d.zones ? '(se la zona si riempie si prosegue nelle zone successive)' : ''}</label>
        <label><input type="radio" name="mode" value="over"> in posizioni consecutive, sovrascrivendo quelle occupate</label></div></div>
      <div data-prev class="note"></div>
      <div data-warn></div>
      </div><footer><span class="left" data-sum></span><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>Inserisci</button></footer></div>`);
    const el = m.el;
    // posizione di partenza predefinita: prima libera
    const first = () => {
      const [a] = d.zones ? [(+$('[data-zone]', el).value) * d.perZone] : [0];
      const b = d.zones ? a + d.perZone : d.slots;
      for (let i = a; i < b; i++) if (!view.chans[i]) return i - a + 1;
      return 1;
    };
    $('[data-start]', el).value = first();
    let plan = null;
    const compute = async () => {
      const mode = $('input[name=mode]:checked', el).value;
      const base = d.zones ? (+$('[data-zone]', el).value) * d.perZone : 0;
      const start = base + Math.max(1, +$('[data-start]', el).value || 1) - 1;
      let pos = [];
      if (mode === 'free') { for (let i = start; i < d.slots && pos.length < n; i++) if (!view.chans[i]) pos.push(i); }
      else for (let i = start; i < start + n && i < d.slots; i++) pos.push(i);
      const ok = pos.length === n;
      const over = mode === 'over' ? pos.filter(i => view.chans[i]).length : 0;
      const label = i => view.def.zones ? `zona ${Math.floor(i / d.perZone) + 1} n. ${i % d.perZone + 1}` : `n. ${i + (d.first == null ? 1 : d.first)}`;
      $('[data-prev]', el).innerHTML = ok
        ? `Posizioni: da ${esc(label(pos[0]))} a ${esc(label(pos[pos.length - 1]))}${over ? ` — <b style="color:var(--bad)">${over} canali verranno sovrascritti</b>` : ''}.`
        : `<b style="color:var(--bad)">Spazio insufficiente: ${pos.length} posizioni disponibili su ${n} necessarie.</b>`;
      $('[data-ok]', el).disabled = !ok;
      plan = null;
      if (!ok) { $('[data-warn]', el).innerHTML = ''; return; }
      const res = await view.adaptFrom(srcKind, chans, pos);
      // canali non trasferibili (es. sola ricezione) restano fuori
      const keep = res.channels.map((c, k) => c ? k : -1).filter(k => k >= 0);
      plan = keep.length ? { pos: keep.map(k => pos[k]), chans: keep.map(k => res.channels[k]) } : null;
      if (!plan) $('[data-ok]', el).disabled = true;
      $('[data-warn]', el).innerHTML = res.warnings && res.warnings.length
        ? `<p style="margin-top:10px">Adattamenti per il ${esc(d.name)}:</p><ul class="warnlist">${res.warnings.map(w => '<li>' + esc(w) + '</li>').join('')}</ul>` : '<p class="note" style="margin-top:10px">Nessun adattamento necessario.</p>';
      $('[data-sum]', el).textContent = plural(n, 'canale', 'canali');
    };
    const recompute = () => compute().catch(reportError);
    $$('input,select', el).forEach(x => x.addEventListener('change', () => { if (x.matches('[data-zone]')) $('[data-start]', el).value = first(); recompute(); }));
    $('[data-start]', el).addEventListener('input', debounce(recompute, 300));
    await compute().catch(reportError);
    return new Promise(res => {
      $('[data-no]', el).onclick = () => { m.close(); res(false); };
      $('[data-ok]', el).onclick = () => {
        if (!plan) return;
        view.snapshot();
        plan.pos.forEach((p, k) => view.chans[p] = plan.chans[k]);
        if (d.zones) view.zone = Math.floor(plan.pos[0] / d.perZone);
        view.setSel(plan.pos, plan.pos[0]);
        view.changed();
        m.close();
        App.showTab(view.kind);
        view.grid.ensureVisible();
        toast(`Inseriti ${plural(plan.pos.length, 'canale', 'canali')} nel ${d.name}` + (plan.pos.length < n ? ` (${n - plan.pos.length} non trasferibili)` : '') + '. Ricordati di salvare il file.');
        res(true);
      };
    });
  },

  // ------------------------------------------------------------ esportazione canali di una radio
  async exportRadio(view) {
    const sel = view.selected().filter(i => view.chans[i]);
    const m = openModal(`<div class="dialog mid"><header>Esporta i canali del ${esc(view.def.name)}</header><div class="body">
      <div class="row radiolist"><label class="l">Canali</label><div>
        <label><input type="radio" name="scope" value="all" checked> tutti (${view.chans.filter(Boolean).length})</label>
        <label><input type="radio" name="scope" value="sel" ${sel.length ? '' : 'disabled'}> solo i selezionati (${sel.length})</label></div></div>
      ${formatChoice()}
      <p class="note" style="margin-top:12px">Per creare il file di un'altra radio (K67, Bajeton, immagini .img di qualsiasi modello) usa il convertitore.</p>
      </div><footer><button class="btn" style="margin-right:auto" data-conv title="Apre la scheda «Converti» con questi canali">🔁 Per un'altra radio…</button>
      <button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>Esporta…</button></footer></div>`);
    $('[data-no]', m.el).onclick = () => m.close();
    $('[data-conv]', m.el).onclick = () => { m.close(); Converter.fromView(view); };
    $('[data-ok]', m.el).onclick = async () => {
      try {
        const scope = $('input[name=scope]:checked', m.el).value;
        const idx = scope === 'sel' ? sel : view.chans.map((c, i) => c ? i : -1).filter(i => i >= 0);
        const fmt = $('input[name=fmt]:checked', m.el).value;
        const r = await api('export', { what: 'radio', radio: view.kind, channels: view.chans, sel: idx, format: fmt }, { save: true, suggest: '' });
        m.close();
        toast('Esportato: ' + r.path);
      } catch (e) { reportError(e); }
    };
  },

  // ------------------------------------------------------------ radio via cavo
  // op: 'read' | 'write'. Restituisce il risultato dell'operazione (o null).
  cable(view, op) {
    const k67 = view.kind === 'k67';
    const write = op === 'write';
    // radio collegate via USB (HID / DFU): nessuna porta COM da scegliere
    const usb = view.def.conn || '';
    const n = view.chans.filter(Boolean).length;
    const title = (write ? 'Scrivi nella radio' : 'Leggi dalla radio') + ' — ' + view.def.name;
    return new Promise(resolve => {
      let running = false, result = null, timer = null, finished = false;
      const m = openModal(`<div class="dialog mid" data-cable><header>📻 ${esc(title)}</header><div class="body">
        <div data-setup>
          <p>${usb ? 'Collega il <b>cavo USB</b> della radio al PC' : 'Collega il cavo di programmazione al PC e alla radio'}, poi <b>accendi la radio</b>${usb ? '' : ' e tieni il volume a metà'}.
          ${write ? `Saranno scritti <b>${n} canali</b>: i canali presenti nella radio verranno sostituiti da quelli della tabella.` : 'I canali letti compariranno nella tabella: poi potrai modificarli e salvarli in un file.'}</p>
          ${usb ? `<div class="row"><label class="l">Collegamento</label><span><b>USB</b> — ${usb === 'dfu'
            ? 'serve il driver USB del costruttore (quello del suo programma, cartella «DM-1701 USB Driver»): vedi la nota della scheda'
            : 'dispositivo HID, Windows lo riconosce senza driver aggiuntivi'}</span></div>` : ''}
          <div class="row" ${usb ? 'style="display:none"' : ''}><label class="l">Porta del cavo</label><select data-port style="min-width:330px"></select>
            <button class="btn" data-refresh title="Cerca di nuovo le porte (dopo aver collegato il cavo)">🔄 Aggiorna</button></div>
          <div class="row note" data-portnote style="margin-top:-6px"></div>
          ${k67 ? `<div class="row"><label class="l">Velocità</label><select data-baud><option value="19200">19200 baud (predefinita)</option><option value="115200">115200 baud</option></select>
            <span class="note">come nel programma originale: la radio può passare da sola a 115200</span></div>
          <div class="row"><label class="l">Password radio</label><input type="text" data-pwd maxlength="8" style="width:120px" autocomplete="off">
            <span class="note">solo se nella radio è impostata una password di programmazione</span></div>` : ''}
          ${write ? `<div class="row"><label class="l">Sicurezza</label><div>
            <label style="display:block"><input type="checkbox" data-backup> prima leggi la radio e salva una copia di sicurezza</label>
            <label style="display:block"><input type="checkbox" data-verify> dopo la scrittura rileggi la radio e verifica i canali</label></div></div>
            <p class="note">Insieme ai canali vengono scritte anche le impostazioni generali contenute nel file aperto${k67 ? ' (menu, zone, tasti…)' : ' (menu, VFO, DTMF…)'}.
            Per non cambiarle, prima leggi la radio (o apri il file salvato da essa) e poi modifica i canali.</p>` : ''}
          ${view.def.generic && view.def.notes ? `<p class="note">${esc(view.def.notes)}</p>` : ''}
          ${view.def.generic ? '<p class="note">Protocollo del driver di CHIRP per questa radio, verificato byte per byte con CHIRP su radio simulate.</p>' : ''}
        </div>
        <div data-run style="display:none">
          <div style="font-weight:600;margin-bottom:6px" data-phase></div>
          <div class="pbar"><div data-bar></div></div>
          <div class="note" data-msg style="margin-top:6px"></div>
          <p data-final style="margin-top:12px;display:none"></p>
          <ul class="warnlist" data-notes style="display:none"></ul>
        </div>
        </div><footer><span class="left note" data-foot></span><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>${write ? 'Scrivi nella radio' : 'Leggi dalla radio'}</button></footer></div>`,
        { onEsc: () => { if (!running) close(); } });
      const el = m.el;
      const close = () => { clearTimeout(timer); m.close(); resolve(finished ? result : null); };
      const sel = $('[data-port]', el);
      let settings = {};

      const loadPorts = async () => {
        const r = await api('cable/ports', {}, { quiet: true });
        settings = r.settings || {};
        if (usb) {
          sel.innerHTML = '<option value="USB">USB</option>';
          $('[data-portnote]', el).textContent = `Copie di sicurezza e letture vengono salvate in: ${r.backupDir}`;
          $('[data-ok]', el).disabled = false;
          return r;
        }
        const last = k67 ? settings.k67Port : view.def.generic ? settings.genPort : settings.bajPort;
        const cur = sel.value || last;
        sel.innerHTML = r.ports.map(p => `<option value="${esc(p.name)}">${esc(p.desc && p.desc !== p.name ? p.desc : p.name)}</option>`).join('');
        if (r.ports.some(p => p.name === cur)) sel.value = cur;
        $('[data-portnote]', el).textContent = r.ports.length ? `Copie di sicurezza e letture vengono salvate in: ${r.backupDir}`
          : 'Nessuna porta seriale trovata: collega il cavo e verifica che il driver (es. CH340 o Prolific) sia installato, poi premi Aggiorna.';
        $('[data-ok]', el).disabled = !r.ports.length;
        return r;
      };
      loadPorts().then(() => {
        if (k67) {
          $('[data-baud]', el).value = String(settings.k67Baud === 115200 ? 115200 : 19200);
          $('[data-pwd]', el).value = settings.k67Pwd || '';
        }
        if (write) {
          $('[data-backup]', el).checked = settings.cableBackup !== false;
          $('[data-verify]', el).checked = settings.cableVerify !== false;
        }
      }).catch(reportError);
      $('[data-refresh]', el).onclick = () => loadPorts().catch(reportError);

      const showRun = () => { $('[data-setup]', el).style.display = 'none'; $('[data-run]', el).style.display = ''; };
      const setNotes = (notes) => {
        const ul = $('[data-notes]', el);
        ul.style.display = notes && notes.length ? '' : 'none';
        ul.innerHTML = (notes || []).map(x => '<li>' + esc(x) + '</li>').join('');
      };
      const done = (st) => {
        running = false;
        const ok = st.state === 'done';
        const fin = $('[data-final]', el);
        fin.style.display = '';
        $('[data-bar]', el).style.width = ok ? '100%' : $('[data-bar]', el).style.width;
        $('[data-bar]', el).classList.toggle('err', !ok);
        if (ok) {
          result = st.result; finished = true;
          if (write) {
            const v = st.result && st.result.verified;
            fin.innerHTML = v === false ? '⚠ <b>Scrittura completata, ma la verifica ha trovato differenze</b>: controlla le segnalazioni.'
              : '✅ <b>Scrittura completata.</b> ' + (v ? 'I canali riletti dalla radio corrispondono.' : '') + ' Puoi spegnere e riaccendere la radio.';
          } else {
            fin.innerHTML = '✅ <b>Lettura completata.</b> I canali sono nella tabella: ricordati di salvarli in un file.';
          }
          $('[data-phase]', el).textContent = 'Operazione completata';
        } else {
          fin.innerHTML = (st.state === 'cancelled' ? '⏹ ' : '❌ ') + esc(st.error || 'Operazione non riuscita');
          if (write && st.state === 'error' && st.phase === 'Scrittura nella radio')
            fin.innerHTML += '<br><span class="note">La scrittura si è interrotta: la radio potrebbe avere dati incompleti. Controlla il collegamento e ripeti la scrittura' +
              (st.result && st.result.backup ? '' : '') + '.</span>';
        }
        setNotes(st.notes);
        $('[data-msg]', el).textContent = '';
        $('[data-foot]', el).textContent = '';
        const no = $('[data-no]', el);
        no.textContent = 'Chiudi'; no.disabled = false; no.classList.add('primary');
        $('[data-ok]', el).style.display = ok ? 'none' : '';
        $('[data-ok]', el).textContent = 'Riprova';
        $('[data-ok]', el).disabled = false;
        if (ok && !write && (!st.notes || st.notes.length <= 1) && !(st.result.radio.warnings || []).length) { close(); toast('Canali letti dalla radio: ricordati di salvarli in un file.'); return; }
        if (ok && !write && st.result.radio.warnings && st.result.radio.warnings.length) setNotes((st.notes || []).concat(st.result.radio.warnings));
      };
      const poll = async () => {
        let st;
        try { st = await api('cable/status', {}, { quiet: true }); } catch (e) { timer = setTimeout(poll, 500); return; }
        $('[data-phase]', el).textContent = st.phase || '';
        const pct = st.total ? Math.round(st.done * 100 / st.total) : 0;
        $('[data-bar]', el).style.width = pct + '%';
        $('[data-msg]', el).textContent = (st.msg || '') + (st.total ? ` ${pct}%` : '');
        const no = $('[data-no]', el);
        no.disabled = !st.canStop;
        $('[data-foot]', el).textContent = st.state === 'running' && !st.canStop ? 'Non spegnere la radio e non scollegare il cavo!' : '';
        if (st.state === 'running') { timer = setTimeout(poll, 200); return; }
        done(st);
      };

      $('[data-no]', el).onclick = async () => {
        if (!running) { close(); return; }
        try { await api('cable/cancel', {}, { quiet: true }); } catch (e) { reportError(e); }
      };
      $('[data-ok]', el).onclick = async () => {
        if (running) return;
        const port = sel.value;
        if (!port) { toast('Scegli la porta del cavo di programmazione.', true); return; }
        const req = { op, radio: view.kind, port, baud: k67 ? +$('[data-baud]', el).value : 0, password: k67 ? $('[data-pwd]', el).value : '',
          backup: write ? $('[data-backup]', el).checked : false, verify: write ? $('[data-verify]', el).checked : false };
        if (write) {
          const c = await chooseDlg('Conferma la scrittura', `Vuoi scrivere ${n} canali nel ${view.def.name} collegato a ${usb ? 'USB' : port}?\n` +
            'I canali attualmente nella radio verranno sostituiti. Durante la scrittura non spegnere la radio e non scollegare il cavo.' +
            (req.backup ? '' : '\nAttenzione: hai disattivato la copia di sicurezza della radio.'),
            [{ id: 'no', label: 'Annulla' }, { id: 'ok', label: 'Scrivi', primary: true }]);
          if (c !== 'ok') return;
          req.channels = view.chans;
        }
        try {
          running = true;
          showRun();
          $('[data-final]', el).style.display = 'none';
          $('[data-bar]', el).classList.remove('err');
          $('[data-bar]', el).style.width = '0%';
          setNotes([]);
          $('[data-ok]', el).disabled = true;
          const no = $('[data-no]', el); no.textContent = 'Interrompi'; no.classList.remove('primary');
          await api('cable/start', req, { quiet: true });
          poll();
        } catch (e) {
          running = false;
          $('[data-ok]', el).disabled = false;
          $('[data-no]', el).textContent = 'Annulla';
          $('[data-setup]', el).style.display = ''; $('[data-run]', el).style.display = 'none';
          reportError(e);
        }
      };
    });
  },

  // ------------------------------------------------------------ esportazione archivio
  async exportPonti() {
    if (!Archive.ponti.length) { toast('L\'archivio è vuoto.'); return; }
    const m = openModal(`<div class="dialog big"><header>Esporta ponti dall'archivio</header>
      <div class="body flush" data-pick style="flex:1"></div>
      <div class="infobar">${formatChoice(true)}</div>
      <footer><span class="left" data-sum></span><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>Esporta…</button></footer></div>`);
    const picker = new PontiPicker($('[data-pick]', m.el), {
      onChange: () => { const k = picker.chosen().length; $('[data-sum]', m.el).textContent = `${k} ponti selezionati verranno esportati (spunta o togli la spunta riga per riga)`; $('[data-ok]', m.el).disabled = !k; },
    });
    picker.setList(Archive.ponti);
    $('[data-no]', m.el).onclick = () => m.close();
    $('[data-ok]', m.el).onclick = async () => {
      try {
        const fmt = $('input[name=fmt]:checked', m.el).value;
        const r = await api('export', { what: 'ponti', ponti: picker.chosen(), format: fmt }, { save: true });
        m.close();
        toast('Esportati ' + picker.chosen().length + ' ponti in ' + r.path);
      } catch (e) { reportError(e); }
    };
  },

  // ------------------------------------------------------------ importazione
  async importPonti() {
    const st = App.init.settings;
    const urls = Array.from(new Set([App.init.defaultURL].concat(st.urls || [])));
    const m = openModal(`<div class="dialog big"><header>Importa ponti nell'archivio</header>
      <div class="infobar" style="flex-direction:column;align-items:stretch;gap:8px">
        <div class="row" style="margin:0">
          <label><input type="radio" name="src" value="url" checked> Da Internet:</label>
          <input type="text" data-url list="url-list" value="${esc(st.url || App.init.defaultURL)}" style="flex:1;min-width:320px">
          <datalist id="url-list">${urls.map(u => `<option value="${esc(u)}">`).join('')}</datalist>
          <button class="btn small" data-ik2ane title="Ripristina l'indirizzo dell'elenco di IK2ANE">IK2ANE</button>
        </div>
        <div class="ik2ane-credit small">📡 L'elenco dei ponti radio italiani è curato e messo a disposizione gratuitamente da <b>IK2ANE</b>
          (<a href="#" data-link="https://www.ik2ane.it">www.ik2ane.it</a>): grazie a lui per questo servizio alla comunità radioamatoriale.</div>
        <div class="row" style="margin:0">
          <label><input type="radio" name="src" value="file"> Da file sul PC (.xls, .xlsx, .csv)</label>
          <span class="note" data-filelabel></span>
        </div>
        <div class="row" style="margin:0">
          <label>Nome canale:</label>${modelloField(st.modello)}
        </div>
        <div class="row" style="margin:0">
          <label><input type="checkbox" data-tonerx> usa il tono anche in ricezione (squelch a tono)</label>
          <span class="spacer"></span>
          <button class="btn primary" data-load>Carica elenco</button>
        </div>
      </div>
      <div data-res class="infobar" style="display:none"></div>
      <details class="scartate" data-scart style="display:none"><summary></summary><ul></ul></details>
      <div class="body flush" data-pick style="flex:1"></div>
      <footer><span class="left" data-sum>Carica un elenco per vedere i ponti.</span>
        <select data-merge title="Cosa fare con i ponti già presenti nell'archivio">
          <option value="add">Aggiungi solo i ponti nuovi</option>
          <option value="update">Aggiungi i nuovi e aggiorna quelli già presenti</option>
          <option value="replace">Sostituisci tutto l'archivio</option>
        </select>
        <button class="btn" data-no>Chiudi</button><button class="btn primary" data-ok disabled>Importa</button></footer></div>`);
    const el = m.el;
    let picker = null, result = null, label = '';
    $$('[data-link]', el).forEach(x => x.onclick = (e) => { e.preventDefault(); Info.open(x.dataset.link); });
    $('[data-ik2ane]', el).onclick = () => { $('[data-url]', el).value = App.init.defaultURL; $('input[name=src][value=url]', el).checked = true; };
    $('[data-url]', el).addEventListener('focus', () => { $('input[name=src][value=url]', el).checked = true; });
    const opts = () => ({ modello: $('[data-modello]', el).value.trim() || '{LOC}', toneRx: $('[data-tonerx]', el).checked });
    const show = (r) => {
      result = r.result; label = r.label;
      result.ponti = result.ponti || []; result.scartate = result.scartate || []; result.sheets = result.sheets || [];
      const res = $('[data-res]', el);
      res.style.display = '';
      const dub = result.ponti.filter(p => p.avvisi && p.avvisi.length).length;
      res.innerHTML = `<span><b>${esc(label)}</b></span>
        ${result.sheets.length > 1 ? `<label>Foglio: <select data-sheet>${result.sheets.map(s => `<option ${s === result.sheet ? 'selected' : ''}>${esc(s)}</option>`).join('')}</select></label>` : ''}
        <span class="tag">Formato: ${esc(result.formato)}</span><span class="tag">${result.ponti.length} ponti letti</span>
        ${dub ? `<span class="tag" style="background:var(--warn)">⚠ ${dub} con valori dubbi, interpretati e segnalati</span>` : ''}
        ${result.scartate.length ? `<span class="tag">${result.scartate.length} righe non importabili</span>` : ''}`;
      const sc = $('[data-scart]', el);
      sc.style.display = result.scartate.length ? '' : 'none';
      $('summary', sc).textContent = `Righe senza una frequenza valida, non importabili (${result.scartate.length}) — clicca per vederle`;
      $('ul', sc).innerHTML = result.scartate.map(s => `<li>Riga ${s.riga}: ${esc(s.motivo)} — ${esc(s.testo)}</li>`).join('');
      const sh = $('[data-sheet]', res);
      if (sh) sh.onchange = () => load('cache', sh.value);
      if (!picker) {
        picker = new PontiPicker($('[data-pick]', el), {
          tipiDefault: ['Analogico'], statiDefault: ['Attivo', 'Temporaneamente non attivo', 'Non H24'],
          onChange: () => {
            const k = picker.chosen().length;
            $('[data-sum]', el).innerHTML = `<b>${k}</b> ponti selezionati per l'importazione · filtri preimpostati: solo ponti utilizzabili in analogico, esclusi i non attivi (modificabili)`;
            $('[data-ok]', el).disabled = !k;
          },
        });
      }
      picker.setList(result.ponti);
    };
    const load = async (source, sheet) => {
      try {
        const src = source || $('input[name=src]:checked', el).value;
        const body = Object.assign({ source: src, url: $('[data-url]', el).value.trim(), sheet: sheet || '' }, opts());
        const r = await api('import/load', body, { busyText: src === 'url' ? 'Download dell\'elenco…' : 'Lettura del file…' });
        if (src === 'file') $('[data-filelabel]', el).textContent = r.label;
        show(r);
        await api('settings', Object.assign({}, App.init.settings, { modello: opts().modello }), { quiet: true }).then(s => App.init.settings = s).catch(() => {});
        if (src === 'url') { App.init.settings.url = r.label; }
      } catch (e) { reportError(e); }
    };
    const srcLabel = () => { $('[data-load]', el).textContent = $('input[name=src]:checked', el).value === 'file' ? 'Scegli il file e carica…' : 'Scarica elenco'; };
    $$('input[name=src]', el).forEach(r => r.addEventListener('change', srcLabel));
    $('[data-url]', el).addEventListener('focus', srcLabel);
    srcLabel();
    $('[data-load]', el).onclick = () => load();
    $('[data-modello]', el).addEventListener('change', () => { if (result) load('cache', result.sheet); });
    $('[data-tonerx]', el).addEventListener('change', () => { if (result) load('cache', result.sheet); });
    $('[data-no]', el).onclick = () => m.close();
    $('[data-ok]', el).onclick = async () => {
      const chosen = picker ? picker.chosen() : [];
      if (!chosen.length) return;
      const mode = $('[data-merge]', el).value;
      if (mode === 'replace' && Archive.ponti.length && !(await confirmDlg('Sostituisci archivio', `L'archivio attuale (${Archive.ponti.length} ponti) verrà sostituito da ${chosen.length} ponti. Puoi annullare con Ctrl+Z nella scheda Archivio.`, 'Sostituisci'))) return;
      const items = clone(chosen).map(p => { delete p.riga; return p; });
      Archive.snapshot();
      let added = 0, updated = 0, same = 0;
      if (mode === 'replace') { Archive.ponti = items; added = items.length; }
      else {
        const idx = new Map(Archive.ponti.map(p => [chiavePonte(p), p]));
        const fresh = [];
        for (const p of items) {
          const old = idx.get(chiavePonte(p));
          if (!old) { fresh.push(p); idx.set(chiavePonte(p), p); added++; continue; }
          if (mode === 'update') {
            const keepName = old.nome, keepId = old.id;
            Object.assign(old, p, { nome: keepName, id: keepId });
            updated++;
          } else same++;
        }
        Archive.ponti = Archive.ponti.concat(fresh);
      }
      Archive.changed(true);
      m.close();
      App.showTab('arch');
      await alertDlg('Importazione completata', `Ponti aggiunti: ${added}` + (mode === 'update' ? `\nPonti aggiornati: ${updated}` : '') + (same ? `\nGià presenti nell'archivio o doppi nell'elenco (non aggiunti): ${same}` : '') +
        `\n\nI ponti con valori dubbi sono evidenziati in giallo con ⚠: controllali e usa «Segna come verificati».`);
    };
  },

  // ------------------------------------------------------------ nomi
  async renamePonti() {
    const sel = Archive.selectedPonti();
    const vis = Archive.visible();
    const m = openModal(`<div class="dialog mid"><header>Rigenera i nomi dei canali</header><div class="body">
      <div class="row"><label class="l">Modello</label>${modelloField(App.init.settings.modello)}</div>
      <div class="row radiolist"><label class="l">Ponti</label><div>
        <label><input type="radio" name="sc" value="sel" ${sel.length ? 'checked' : 'disabled'}> selezionati (${sel.length})</label>
        <label><input type="radio" name="sc" value="vis" ${sel.length ? '' : 'checked'}> tutti quelli visibili con i filtri (${vis.length})</label></div></div>
      <p class="note">I nomi che hai modificato a mano verranno sostituiti (puoi annullare con Ctrl+Z).</p>
      </div><footer><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>Rigenera</button></footer></div>`);
    $('[data-no]', m.el).onclick = () => m.close();
    $('[data-ok]', m.el).onclick = async () => {
      try {
        const list = $('input[name=sc]:checked', m.el).value === 'sel' ? sel : vis;
        const modello = $('[data-modello]', m.el).value.trim() || '{LOC}';
        const r = await api('rename', { ponti: list, modello });
        Archive.snapshot();
        list.forEach((p, i) => p.nome = r.names[i]);
        Archive.changed();
        m.close();
        toast(`Rinominati ${list.length} ponti.`);
      } catch (e) { reportError(e); }
    };
  },

  help() {
    const m = openModal(`<div class="dialog mid help"><header>Guida rapida</header><div class="body">
      <p>Con il <b>cavo di programmazione</b> puoi leggere i canali dalla radio («📻 Leggi dalla radio»), modificarli e scriverli («📻 Scrivi nella radio»).
      Puoi anche lavorare sui <b>file</b> del programma originale (K67 Pro: <b>.xlc</b>; Bajeton 8300: <b>.dat</b>): aprili, modificali e salvali.
      Prima di salvare sopra un file ne viene conservata una copia <b>.bak</b>.</p>
      <h4>Modelli Baofeng e Bajeton</h4><ul>
        <li>All'avvio compare la schermata «🏠 Home». Le schede delle radio (anche Baofeng K67 Pro e Bajeton 8300) si aprono dalla galleria dei modelli
          e si chiudono con ×.</li>
        <li>La scheda «📻 Modelli» elenca tutte le radio Baofeng e Bajeton, divise tra «MODELLI COMPATIBILI» e «MODELLI NON COMPATIBILI» (con il numero di modelli) e,
          in ciascuna parte, per tipologia (Radioamatoriali, DMR, GMRS / FRS, PMR446 e uso libero, Per bambini); dentro ogni tipologia i modelli sono separati tra <b>Baofeng</b> e <b>Bajeton</b>,
          in ordine alfabetico. Ogni modello ha la sua
          scheda, anche quando la gestione è la stessa di un altro (es. UV-5RA come UV-5R; i modelli equivalenti indicati dalla tabella di CHIRP lo dicono nella scheda).
          Cerca il modello e filtra per marca, tipo e funzioni. Clic sulla radio per aprirne la scheda dei canali, aggiornarne il firmware o scaricare programmi e manuali del costruttore.</li>
        <li>Le radio gestite con i driver di <b>CHIRP</b> (oltre 70 modelli) si leggono e si scrivono via cavo e si salvano in file <b>.img</b>, apribili anche con CHIRP;
          si aprono anche i file .img salvati con CHIRP. Le impostazioni della radio restano quelle lette: si modificano canali e ponti.</li>
        <li>Se una radio esiste in più versioni (es. UV-5RX3) scegli quella giusta: se non lo è, la radio non risponde e non viene modificata.</li>
        <li>Canali di sola ricezione: nella colonna TX scrivi OFF (se la radio lo permette). I canali si copiano tra tutte le radio aperte con «Copia in…».</li>
      </ul>
      <h4>Radio via cavo</h4><ul>
        <li>Scegli la porta COM del cavo (se non compare, collega il cavo, installa il driver e premi «Aggiorna»). Accendi la radio prima di iniziare.</li>
        <li>Prima di scrivere, di serie il programma legge la radio e ne salva una copia di sicurezza; dopo la scrittura rilegge la radio e verifica i canali.
          Le copie si trovano nella cartella «copie_radio» accanto all'archivio.</li>
        <li>Con i canali vengono scritte anche le impostazioni generali del file aperto: per conservare quelle della radio, leggi prima la radio.</li>
        <li>K67 Pro: se nella radio è impostata una password di programmazione, scrivila nella finestra. Come nel programma originale, la scrittura
          viene rifiutata se la banda di frequenza della radio è diversa da quella del file: in quel caso leggi prima la radio e lavora sui canali letti.</li>
      </ul>
      <h4>Aggiornamento del firmware Baofeng</h4><ul>
        <li><b>UV-5RM / UV-5RH</b> (file <b>.BF</b>, caricatore KDH): modalità automatica con la radio accesa, oppure manuale con la radio accesa tenendo premuto il PTT.
          Il firmware deve essere quello del chip radio della tua radio (BK4819, AT1846S…): le versioni non sono intercambiabili.
          Lo stesso caricatore si usa per le radio della stessa famiglia (K5 Plus, UV-5RM Plus, UV-17R, UV-17R Plus, UV-21R, UV-21L, UV-21, BF-T60 / TR-988UV)
          con i file forniti dal costruttore: il modello è riconosciuto dal nome del file e va confermato scrivendolo.</li>
        <li><b>UV-25 Plus</b> e radio del programma «BF-UVs Pro CPS» (<b>UV-28 Pro, UV-29 Pro / SW-9, AR-152 Pro, UV-7B, UV-32</b>; file <b>.kdhx</b>, caricatore
          «Baofeng Bootloader Pro», piattaforma H802). Baofeng pubblica il firmware solo per la UV-25 Plus: per le altre radio servono i file del costruttore
          e il modello indicato dal nome del file va scritto per confermare. <b>UV-9G</b> (file .kdh) e <b>Bajeton BJ-9000</b> (file .Firmware) con i loro caricatori.</li>
        <li><b>5RH Pro / K67 Pro</b> (file <b>.dat</b> «BF_5RH»): la radio deve essere in modalità aggiornamento; dichiara la sua elettronica (V1 o V2) e, se il file
          è per l'altra, l'aggiornamento non parte.</li>
        <li>I protocolli sono quelli dei programmi originali dei costruttori, confrontati byte per byte con il loro codice su radio simulate.</li>
      </ul>
      <h4>Aggiornamento del firmware (Bajeton BJ-8300 e BJ-7800)</h4><ul>
        <li>La BJ-7800 usa lo stesso programma di aggiornamento della BJ-8300 (file <b>.anc</b>, es. BJ7800_CN_EN_V0.17_20251126.anc): si aggiorna allo stesso modo,
          dalla sua scheda o dalla galleria dei modelli. Il modello è riconosciuto dal nome del file.</li>
        <li>Nella scheda Bajeton 8300 (dalla galleria dei modelli) premi «⬆ Aggiorna firmware». Scegli il file <b>.anc</b> dal PC o scaricalo da Internet (anche link di Google Drive
          o Dropbox): può essere dentro archivi <b>.zip</b> o <b>.rar</b>, anche uno dentro l'altro. Il programma controlla il formato del file, la piattaforma,
          il nome e l'impronta dei firmware ufficiali conosciuti; i firmware scaricati vengono conservati nella cartella «firmware» accanto all'archivio.</li>
        <li>Prima di inviare il firmware la radio viene <b>identificata</b> con il protocollo di programmazione della Bajeton 8300 (e, di serie, ne viene salvata
          una copia dei canali); se non risponde come una Bajeton 8300 l'aggiornamento non parte. Devi comunque confermare che la radio è una BJ-8300.</li>
        <li>Modalità di recupero, per una radio che non si avvia: la radio non può essere identificata, quindi va confermata scrivendo BJ8300.</li>
        <li>Durante l'aggiornamento non spegnere la radio e non scollegare il cavo. Il protocollo è quello del programma originale «BinTe ANC Bootloader» V1.1.</li>
      </ul>
      <h4>Canali</h4><ul>
        <li>Doppio clic, <kbd>Invio</kbd> o <kbd>F2</kbd> per modificare una cella; si può anche iniziare a scrivere direttamente. <kbd>Esc</kbd> annulla.</li>
        <li>Frequenze in MHz (145.600 o 145,600). Lo shift si scrive come -0.6, +1.6, +5: il TX viene calcolato. I toni sono subtoni CTCSS (es. 88.5) o codici DCS (es. D023N, D023I invertito); vuoto = OFF.</li>
        <li>Selezione: clic, <kbd>Maiusc</kbd>+clic per un intervallo, <kbd>Ctrl</kbd>+clic per aggiungere, <kbd>Ctrl</kbd>+<kbd>A</kbd> tutti.</li>
        <li><kbd>Ctrl</kbd>+<kbd>C</kbd> / <kbd>X</kbd> / <kbd>V</kbd> copia, taglia, incolla (anche tra le due radio e dall'archivio, con gli adattamenti segnalati);
          <kbd>Canc</kbd> cancella; <kbd>Ins</kbd> inserisce una riga; <kbd>Alt</kbd>+<kbd>↑</kbd>/<kbd>↓</kbd> sposta; <kbd>Ctrl</kbd>+<kbd>Z</kbd>/<kbd>Y</kbd> annulla/ripeti; <kbd>Ctrl</kbd>+<kbd>S</kbd> salva.</li>
        <li>Bajeton: 15 zone da 64 canali; scegli la zona dalla barra sopra la tabella, «Tutte» per vederle insieme, «Sposta in zona» per spostare i selezionati.</li>
        <li>Nomi: massimo 16 caratteri sul K67 e 12 sulla Bajeton. Potenza Media solo sulla Bajeton (sul K67 diventa Alta).</li>
      </ul>
      <h4>Converti, importa ed esporta</h4><ul>
        <li>Scheda «🔁 Converti»: scegli un file (K67 .xlc, Bajeton .dat v1.0.8 o v1.0.9, immagini .img di qualsiasi radio anche salvate con CHIRP, CSV di CHIRP,
          tabelle Excel/CSV) o una scheda aperta, poi il formato di destinazione: CSV di CHIRP (universale), Excel, CSV, il file .xlc del K67, il .dat della Bajeton
          o l'immagine .img di qualsiasi radio compatibile. I canali vengono adattati alla radio di destinazione e le modifiche elencate.</li>
        <li>Per l'immagine .img di una radio serve un'immagine di quella radio (letta via cavo o aperta da file): contiene impostazioni e taratura, che non si
          possono inventare. Si cambiano solo i canali.</li>
        <li>«📥 Importa…» nella scheda di una radio inserisce i canali di qualsiasi di questi file, adattati, nelle posizioni che scegli.</li>
        <li>«Esporta…» salva i canali in Excel, CSV o CSV di CHIRP; «🔁 Per un'altra radio…» apre il convertitore con i canali della scheda.</li>
      </ul>
      <h4>Archivio ponti</h4><ul>
        <li>L'elenco aggiornato dei ponti radio italiani è messo a disposizione di tutti, gratuitamente, da <b>IK2ANE</b> (www.ik2ane.it): è grazie al suo lavoro che
          l'archivio ponti è possibile.</li>
        <li>«Importa» legge l'elenco di IK2ANE da Internet (indirizzo già impostato, modificabile) o da file .xls/.xlsx/.csv, anche esportati da CHIRP o da questo programma.</li>
        <li>Shift e toni scritti in modo irregolare vengono interpretati; le righe dubbie sono in giallo con ⚠ e l'avviso spiega cosa è stato deciso.</li>
        <li>Filtra per banda, tipo, stato, regione, provincia, rete o testo, poi spunta i ponti che ti interessano (casella a sinistra, oppure «☑ Seleziona tutti» /
          «☐ Deseleziona tutti»: «Seleziona tutti» spunta i ponti visibili con i filtri, «Deseleziona tutti» toglie la spunta a tutti, anche a quelli nascosti dai filtri;
          la barra in basso indica quanti spuntati sono nascosti; le spunte restano salvate nell'archivio).</li>
        <li>«📻 Inoltra alla radio» scrive i ponti spuntati (o, se non ne hai spuntati, quelli selezionati) nella radio di destinazione indicata sopra la tabella:
          se c'è già una radio scelta ti chiede la conferma, altrimenti puoi sceglierla dall'elenco di tutte le radio compatibili, diviso per marca e per tipologia,
          con ogni modello elencato singolarmente in ordine alfabetico.</li>
        <li>Prima di scrivere viene controllato che i ponti siano compatibili con la radio scelta: frequenze dentro le gamme della radio, nessun ponte radioamatoriale
          su un apparato PMR446 / di uso libero o GMRS, nessun ponte solo digitale (DMR, C4FM, D-Star, NXDN) su una radio solo analogica (sulle radio DMR i ponti solo
          DMR diventano canali DMR). Se qualcosa non va compare un avviso: puoi annullare, scegliere un'altra radio o, se una parte dei ponti è compatibile, inviare solo quella.
          Le radio vendute come PMR446 / uso libero o GMRS che accettano anche le frequenze radioamatoriali (gamme del driver ricavate dal programma del
          costruttore) ricevono un avviso giallo invece del blocco: puoi scrivere comunque i ponti, sia con «Inoltra alla radio» sia dalla scheda della radio.
          Rispetta le norme: in Italia la trasmissione sulle frequenze radioamatoriali richiede la licenza e un apparato PMR446 omologato non va modificato.</li>
        <li>Canali digitali (DMR sulla DM-32UV, digitali criptati sulle PMR della serie D): colonna «Modo»; i subtoni valgono solo per i canali analogici e i canali
          digitali non si copiano sulle radio solo analogiche.</li>
        <li>L'archivio si salva da solo in: <span class="note">${esc(App.init.archivePath)}</span></li>
      </ul>
      <p class="note">IU0VYT Flash, versione ${esc(App.init.version)} (verifica degli aggiornamenti con il pulsante «🔄 Aggiornamenti» in alto o nella scheda «ℹ Info»; licenza GNU GPL v3 o successiva) · i formati dei file e i protocolli via cavo sono stati ricostruiti dai programmi originali dei costruttori e verificati con radio simulate: conserva sempre le copie di sicurezza.</p>
      </div><footer><button class="btn primary" data-ok>Chiudi</button></footer></div>`);
    $('[data-ok]', m.el).onclick = () => m.close();
  },
};

function formatChoice(inline) {
  const f = [['xlsx', 'Excel (.xlsx)'], ['xls', 'Excel 97-2003 (.xls)'], ['csv', 'CSV per Excel (separatore ;)'], ['chirp', 'CSV per CHIRP']];
  return `<div class="row radiolist" style="margin:0"><label class="l">Formato</label><div style="display:flex;gap:14px;flex-wrap:wrap">
    ${f.map((x, i) => `<label style="margin:0"><input type="radio" name="fmt" value="${x[0]}" ${i === 0 ? 'checked' : ''}> ${x[1]}</label>`).join('')}</div></div>`;
}
