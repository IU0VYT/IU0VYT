'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
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
        <div class="note">Riconosce la radio collegata e ne mostra marca, modello, tipo e firmware. Usa solo comandi di lettura:
        <b>nella radio non viene scritto nulla.</b></div></div></div>
      <h3>1. Collegamento</h3>
      <div class="convrow"><label class="l">Porta</label><select data-port style="min-width:330px"></select>
        <button class="btn" data-refresh title="Cerca di nuovo le porte (dopo aver collegato il cavo)">🔄 Aggiorna</button></div>
      <div class="note" data-portnote></div>
      <div class="convrow"><label class="l">Cerca tra</label>
        <select data-brand><option value="">tutte le marche</option>${catalogBrands().filter(b => App.catalog.some(c => c.brand === b && (c.drivers || []).length)).map(b => `<option>${esc(b)}</option>`).join('')}</select>
        <select data-family><option value="">radio analogiche e DMR</option><option value="analog">solo radio analogiche</option><option value="dmr">solo radio DMR</option></select>
        <span class="note">meno radio = ricerca più veloce</span></div>
      <h3>Identificazione automatica</h3>
      <div class="convrow"><button class="btn primary" data-auto>🔄 Abilita identificazione automatica</button>
        <span class="note" data-autost></span></div>
      <ul class="note idlist">
        <li>Collega il cavo: la radio viene riconosciuta da sola, anche se la accendi dopo, e si apre la sua scheda con i canali letti.</li>
        <li>Più cavi: ogni radio viene cercata sul suo cavo, tutte insieme.</li>
        <li>Le radio già identificate in passato vengono riconosciute più velocemente.</li>
        <li>Mentre una radio risponde compare in basso «Identificazione Radio In Corso».</li>
        <li>Per un'altra radio sullo stesso cavo premi <b>«🔎 Identifica Radio»</b> in alto a destra.</li>
        <li>La scelta resta memorizzata.</li>
      </ul>
      <h3>2. Ricerca</h3>
      <div class="convrow"><button class="btn primary" data-go>🔎 Identifica la radio</button><button class="btn" data-stop style="display:none">Interrompi</button></div>
      <div data-run style="display:none">
        <div style="font-weight:600;margin:6px 0" data-phase></div>
        <div class="pbar"><div data-bar></div></div>
        <div class="note" data-msg style="margin-top:6px"></div>
      </div>
      <div data-res></div>
      <p class="note" style="margin-top:14px"><b>Radio non trovata?</b> Accendila con il volume a metà, inserisci lo spinotto fino in fondo e
      controlla la porta. UV-17, UV-13Pro e GM-15 Pro vanno accese <b>dopo</b> aver collegato il cavo. Radio DMR con cavo USB diretto
      (RD-5R, DM-1801, DM-1701): porta «USB».</p>
    </div></div>`;
    $('[data-refresh]', page).onclick = () => this.loadPorts(page).catch(reportError);
    $('[data-go]', page).onclick = () => this.start(page).catch(reportError);
    $('[data-stop]', page).onclick = () => api('cable/cancel', {}, { quiet: true }).catch(reportError);
    $('[data-auto]', page).onclick = () => AutoIdent.toggle().catch(reportError);
    AutoIdent.paint();
    await this.loadPorts(page);
  },

  async loadPorts(page) {
    const r = await api('cable/ports', {}, { quiet: true });
    const sel = $('[data-port]', page);
    const cur = sel.value || (r.settings || {}).genPort || '';
    sel.innerHTML = r.ports.map(p => `<option value="${esc(p.name)}">${esc(portLabel(p))}</option>`).join('') +
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
    if (st.op !== 'identify' || st.auto) { this.finish(page); return; }
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

// ------------------------------------------------------------ identificazione automatica (4.0-4.2)
// Controllo in sottofondo (ogni secondo e mezzo), in qualsiasi scheda: stato nella
// scheda «Identifica»; il riquadro «Identificazione Radio In Corso» compare solo
// quando una radio collegata risponde e sparisce quando è riconosciuta o la
// ricerca finisce. Quando il programma riconosce una radio ne mostra i dati e apre
// la sua scheda, poi la ricerca su quel cavo si ferma: il pulsante «Identifica
// Radio» della barra in alto (attivo solo con l'opzione attiva) la fa ripartire.
const AutoIdent = {
  st: null,
  seq: null, // ultimo esito già mostrato
  busy: false,

  start() {
    const stop = $('#autobanner-stop');
    if (stop) stop.onclick = () => this.tick({ stop: true }).then(() => toast('Ricerca interrotta. Per riprenderla premi «Identifica Radio».')).catch(reportError);
    const nb = $('#btn-autonew');
    if (nb) nb.onclick = () => this.rescan().catch(reportError);
    const tick = () => this.tick().catch(() => {}).finally(() => setTimeout(tick, 1500));
    tick();
  },

  async tick(body) {
    const r = await api('identify/auto', body || {}, { quiet: true });
    this.st = r;
    if (this.seq === null) this.seq = r.seq; // esiti precedenti all'apertura della pagina: non si ripetono
    this.paint();
    // 4.3: più radio (una per cavo) possono essere riconosciute insieme
    const evs = (r.events || []).filter(e => e.seq > this.seq);
    if (evs.length && !this.busy) {
      this.seq = Math.max(...evs.map(e => e.seq));
      this.busy = true;
      try { await this.found(evs); } catch (e) { reportError(e); } finally { this.busy = false; }
    }
  },

  // «Identifica Radio»: nuova ricerca su tutti i cavi
  async rescan() {
    if (!(this.st && this.st.enabled)) return;
    await this.tick({ rescan: true });
    const n = this.st.rescanned || 0;
    toast(n ? 'Ricerca di una nuova radio sul cavo di programmazione: collega la radio e accendila.'
      : 'Nessun cavo di programmazione collegato: collegalo al PC e alla radio.', !n);
  },

  async toggle() {
    const on = !(this.st && this.st.enabled);
    await this.tick({ enabled: on });
    toast(on ? 'Identificazione automatica attiva: collega la radio con il cavo di programmazione.' : 'Identificazione automatica disattivata.');
  },

  paint() {
    const st = this.st;
    const nb = $('#btn-autonew');
    if (nb && st) {
      nb.disabled = !st.enabled;
      nb.title = st.enabled ? 'Cerca una nuova radio collegata al cavo di programmazione (identificazione automatica)'
        : 'Attiva prima l\'identificazione automatica nella scheda «Identifica»';
    }
    const ban = $('#autobanner');
    if (ban) {
      const show = !!(st && st.enabled && st.searching);
      ban.hidden = !show;
      if (show) $('.ab-port', ban).textContent = (st.answering || []).length ? `(${st.answering.join(', ')})` : '';
    }
    const page = $('#page-ident');
    const btn = page && $('[data-auto]', page);
    if (!btn || !st) return;
    const on = st.enabled;
    btn.textContent = on ? '⏹ Disattiva identificazione automatica' : '🔄 Abilita identificazione automatica';
    $('[data-autost]', page).textContent = on ? ('Attiva. ' + (st.status || 'In attesa del collegamento di una radio')) : 'Non attiva.';
  },

  // found: radio riconosciute (una o più insieme). Le schede già aperte non si
  // toccano: compare l'avviso «Radio già identificata!! Scheda Radio già aperta»
  // (al plurale con due o più radio); le altre si aprono con i canali letti.
  async found(evs) {
    const page = $('#page-ident');
    if (Identify.built && !Identify.running) {
      Identify.result = evs[evs.length - 1].result;
      Identify.show(page, Identify.result);
    }
    const already = [], opened = [], kept = [];
    for (const ev of evs) {
      const r = ev.result, kind = r.kind;
      const old = Views[kind];
      if (old && old.loaded) { already.push(ev); continue; }
      const res = await api('identify/open', { kind, seq: ev.seq }, { busyText: 'Apertura…' });
      const v = App.openRadio(kind, false);
      v.load(res.radio);
      v.path = '';
      v.fromRadio = true;
      v.dirty = true;
      opened.push(ev);
    }
    if (opened.length) {
      App.onDirty();
      App.showTab(opened[0].result.kind);
      App.saveSession();
    } else if (already.length) {
      App.showTab(already[0].result.kind);
    }
    if (already.length) {
      idBanner(already.length > 1 ? 'Radio già identificate!! Schede radio già aperte' : 'Radio già identificata!! Scheda Radio già aperta',
        already.map(ev => ev.result.name + ' (' + ev.port + ')').join(' · '));
    }
    if (!opened.length) return;
    const list = [];
    for (const ev of opened) {
      const r = ev.result;
      if (opened.length > 1) list.push('— ' + r.name + ' (' + r.type + ', porta ' + ev.port + ')');
      for (const f of (r.fields || [])) list.push(f.l + ': ' + f.v);
      list.push('Canali presenti: ' + r.count);
      if (r.backup) list.push('Copia della lettura: ' + r.backup);
      if ((r.others || []).length) list.push('Altri modelli possibili (stesso collegamento): ' + r.others.map(o => o.name).join(', ') + ' — puoi sceglierli nella scheda «Identifica».');
    }
    list.push('La ricerca automatica su questi cavi ora è ferma: per un\'altra radio premi «Identifica Radio» in alto a destra.');
    if (opened.length === 1) {
      const ev = opened[0];
      await alertDlg('Radio collegata riconosciuta: ' + ev.result.name,
        `${ev.result.type} — porta ${ev.port}. Ho aperto la sua scheda con i canali letti dalla radio: ricordati di salvarli in un file.`, list);
    } else {
      await alertDlg(opened.length + ' radio collegate riconosciute: ' + opened.map(ev => ev.result.name).join(', '),
        'Ho aperto le loro schede con i canali letti dalle radio: ricordati di salvarli in un file.', list);
    }
  },
};

// idBanner (4.3): avviso ben visibile in alto, si chiude da solo dopo alcuni secondi
function idBanner(title, detail) {
  let b = $('#idbanner');
  if (!b) {
    b = document.createElement('div');
    b.id = 'idbanner';
    b.setAttribute('role', 'alert');
    document.body.appendChild(b);
  }
  b.innerHTML = `<b>${esc(title)}</b>${detail ? `<span>${esc(detail)}</span>` : ''}`;
  b.hidden = false;
  clearTimeout(idBanner.t);
  idBanner.t = setTimeout(() => { b.hidden = true; }, 7000);
  b.onclick = () => { b.hidden = true; };
}
