'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda «Identifica» (3.8): cerca la radio collegata provando i protocolli di
// lettura di tutte le radio conosciute, la legge e mostra marca, modello, tipo,
// firmware e le altre informazioni che la radio comunica. Nessuna scrittura.

const Identify = {
  built: false,
  running: false,
  result: null,
  timer: null,

  async render(page) {
    page = page || $('#page-ident');
    if (this.built) { if (!this.running) this.loadPorts(page).catch(() => {}); return; }
    this.built = true;
    page.innerHTML = `<div class="infopage"><div class="infocard convcard">
      <div class="infohead"><div style="font-size:44px">🔎</div><div><h1>Identifica la radio</h1>
        <div class="note">Collega la radio al PC e accendila: il programma prova uno dopo l'altro i protocolli di <b>lettura</b> di tutte le radio
        che conosce finché una risponde, poi la legge e mostra marca, modello, tipo, firmware e le altre informazioni che la radio comunica.
        <b>Nessun dato viene scritto nella radio.</b></div></div></div>
      <h3>1. Collegamento</h3>
      <div class="convrow"><label class="l">Porta</label><select data-port style="min-width:330px"></select>
        <button class="btn" data-refresh title="Cerca di nuovo le porte (dopo aver collegato il cavo)">🔄 Aggiorna</button></div>
      <div class="note" data-portnote></div>
      <div class="convrow"><label class="l">Cerca tra</label>
        <select data-brand><option value="">tutte le marche</option><option>Baofeng</option><option>Bajeton</option></select>
        <select data-family><option value="">radio analogiche e DMR</option><option value="analog">solo radio analogiche</option><option value="dmr">solo radio DMR</option></select>
        <span class="note">restringere la ricerca la rende più veloce</span></div>
      <h3>2. Ricerca</h3>
      <div class="convrow"><button class="btn primary" data-go>🔎 Identifica la radio</button><button class="btn" data-stop style="display:none">Interrompi</button></div>
      <div data-run style="display:none">
        <div style="font-weight:600;margin:6px 0" data-phase></div>
        <div class="pbar"><div data-bar></div></div>
        <div class="note" data-msg style="margin-top:6px"></div>
      </div>
      <div data-res></div>
      <p class="note" style="margin-top:14px">Se la radio non viene trovata: controlla che sia accesa con il volume a metà, che lo spinotto del cavo sia
      inserito fino in fondo e che la porta sia quella del cavo; alcune radio (UV-17, UV-13Pro, GM-15 Pro) vanno accese <b>dopo</b> aver collegato il cavo.
      Le radio DMR con il cavo USB diretto (RD-5R, DM-1801, DM-1701) si cercano scegliendo la porta «USB».</p>
    </div></div>`;
    $('[data-refresh]', page).onclick = () => this.loadPorts(page).catch(reportError);
    $('[data-go]', page).onclick = () => this.start(page).catch(reportError);
    $('[data-stop]', page).onclick = () => api('cable/cancel', {}, { quiet: true }).catch(reportError);
    await this.loadPorts(page);
  },

  async loadPorts(page) {
    const r = await api('cable/ports', {}, { quiet: true });
    const sel = $('[data-port]', page);
    const cur = sel.value || (r.settings || {}).genPort || '';
    sel.innerHTML = r.ports.map(p => `<option value="${esc(p.name)}">${esc(p.desc && p.desc !== p.name ? p.desc : p.name)}</option>`).join('') +
      '<option value="USB">USB — radio DMR collegate direttamente (RD-5R, DM-1801, DM-1701)</option>';
    if ([...sel.options].some(o => o.value === cur)) sel.value = cur;
    $('[data-portnote]', page).textContent = r.ports.length ? `Le letture vengono salvate anche in: ${r.backupDir}`
      : 'Nessuna porta seriale trovata: collega il cavo e verifica che il driver (es. CH340 o Prolific) sia installato, poi premi Aggiorna.';
  },

  async start(page) {
    if (this.running) return;
    const port = $('[data-port]', page).value;
    if (!port) { toast('Scegli la porta del cavo di programmazione.', true); return; }
    this.running = true;
    this.result = null;
    $('[data-res]', page).innerHTML = '';
    $('[data-run]', page).style.display = '';
    $('[data-bar]', page).classList.remove('err');
    $('[data-bar]', page).style.width = '0%';
    $('[data-go]', page).disabled = true;
    $('[data-stop]', page).style.display = '';
    try {
      await api('identify/start', { port, brand: $('[data-brand]', page).value, family: $('[data-family]', page).value }, { quiet: true });
    } catch (e) {
      this.finish(page);
      throw e;
    }
    this.poll(page);
  },

  finish(page) {
    this.running = false;
    clearTimeout(this.timer);
    $('[data-go]', page).disabled = false;
    $('[data-stop]', page).style.display = 'none';
  },

  async poll(page) {
    let st;
    try { st = await api('cable/status', {}, { quiet: true }); } catch (e) { this.timer = setTimeout(() => this.poll(page), 500); return; }
    if (st.op !== 'identify') { this.finish(page); return; }
    $('[data-phase]', page).textContent = st.phase || '';
    const pct = st.total ? Math.round(st.done * 100 / st.total) : 0;
    $('[data-bar]', page).style.width = pct + '%';
    $('[data-msg]', page).textContent = st.msg || '';
    if (st.state === 'running') { this.timer = setTimeout(() => this.poll(page), 250); return; }
    this.finish(page);
    $('[data-run]', page).style.display = 'none';
    if (st.state !== 'done') {
      $('[data-res]', page).innerHTML = `<div class="infores bad">${st.state === 'cancelled' ? '⏹ ' : '❌ '}${esc(st.error || 'Ricerca non riuscita')}</div>`;
      return;
    }
    this.result = st.result;
    this.show(page, st.result);
  },

  show(page, r) {
    const box = $('[data-res]', page);
    const answers = (r.answers || []).length ? `<details style="margin-top:8px"><summary>La radio ha risposto a ${r.answers.length} ${r.answers.length === 1 ? 'protocollo' : 'protocolli'} senza completare la lettura</summary>
      <ul class="warnlist">${r.answers.map(a => '<li>' + esc(a) + '</li>').join('')}</ul></details>` : '';
    if (!r.found) {
      box.innerHTML = `<div class="infores bad">Nessuna radio riconosciuta dopo ${r.tried} tentativi (${r.seconds} s).
        ${(r.answers || []).length ? 'La radio ha risposto, ma con nessuno dei protocolli la lettura è andata a buon fine: vedi i dettagli.' : 'La radio non ha risposto a nessun protocollo.'}</div>${answers}`;
      return;
    }
    const others = r.others || [];
    const choose = others.length ? `<div class="convbase"><p>${r.ambiguous
      ? '<b>Modello non distinguibile via cavo:</b> questi modelli usano lo stesso collegamento e la stessa memoria e la radio non dice quale dei due è. Scegli il tuo:'
      : 'Altri modelli compatibili con la memoria letta (stesso collegamento):'}</p>
      <select data-kind><option value="${esc(r.kind)}">${esc(r.name)}</option>${others.map(o => `<option value="${esc(o.kind)}">${esc(o.name)}</option>`).join('')}</select></div>` : '';
    box.innerHTML = `<div class="infores ok"><span style="font-size:18px">✔ <b>${esc(r.name)}</b></span><br>${esc(r.type)} — trovata in ${r.seconds} s dopo ${r.tried} ${r.tried === 1 ? 'tentativo' : 'tentativi'}.</div>
      <table class="infotab" style="margin-top:10px">${(r.fields || []).map(f => `<tr><th>${esc(f.l)}</th><td>${esc(f.v)}</td></tr>`).join('')}
        <tr><th>Canali presenti</th><td>${r.count}</td></tr>
        ${r.backup ? `<tr><th>Copia della lettura</th><td>${esc(r.backup)}</td></tr>` : ''}</table>
      ${choose}
      <div class="convrow" style="margin-top:10px"><button class="btn primary" data-open>Apri i canali nella scheda della radio</button>
        <span class="note">come «Leggi dalla radio»: poi puoi modificarli e salvarli</span></div>${answers}`;
    $('[data-open]', box).onclick = () => this.open(page).catch(reportError);
  },

  async open(page) {
    const r = this.result;
    if (!r || !r.found) return;
    const ks = $('[data-kind]', page);
    const kind = ks ? ks.value : r.kind;
    const old = Views[kind];
    if (old && old.loaded && !(await old.confirmDiscard())) return;
    const res = await api('identify/open', { kind }, { busyText: 'Apertura…' });
    const v = App.openRadio(kind, false);
    v.load(res.radio);
    v.path = '';
    v.fromRadio = true;
    v.dirty = true;
    App.onDirty();
    App.showTab(kind);
    App.saveSession();
    toast('Canali letti dalla radio: ricordati di salvarli in un file.');
  },
};
