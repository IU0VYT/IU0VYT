'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Aggiornamento del firmware (Bajeton BJ-8300/7800/7700 .anc e BJ-9000 .Firmware, Baofeng UV-5RM/5RH .BF,
// 5RH Pro .dat, UV-25 Plus .kdhx, UV-9G .kdh): scelta del firmware (file del PC
// o Internet, anche dentro archivi .zip/.rar), identificazione della radio,
// conferme e avanzamento.

const Firmware = {
  async choose(choices) {
    return new Promise(res => {
      const m = openModal(`<div class="dialog mid"><header>Più firmware trovati</header><div class="body">
        <p>Negli archivi ci sono più file firmware. Scegli quello da usare:</p>
        <div class="radiolist">${choices.map((c, i) => `<label><input type="radio" name="fwc" value="${i}" ${i === 0 ? 'checked' : ''}>
          <b>${esc(c.name)}</b> <span class="note">${esc(fmtSize(c.size))} — ${esc(c.where)}</span></label>`).join('')}</div>
        </div><footer><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>Usa questo firmware</button></footer></div>`,
        { onEsc: () => { m.close(); res(null); } });
      $('[data-no]', m.el).onclick = () => { m.close(); res(null); };
      $('[data-ok]', m.el).onclick = () => { const v = +$('input[name=fwc]:checked', m.el).value; m.close(); res(v); };
    });
  },

  // carica il firmware: { source: 'file' | 'url', url }
  async load(req) {
    let r = await api('fw/load', req, { busyText: req.source === 'url' ? 'Scaricamento del firmware…' : 'Lettura del file…' });
    if (r.state === 'choose') {
      const i = await this.choose(r.choices);
      if (i == null) return null;
      r = await api('fw/pick', { index: i }, { busyText: 'Controllo del firmware…' });
    }
    return r.state === 'ready' ? r : null;
  },

  // ------------------------------------------------------------ finestra principale
  // expect: tipo di firmware della radio da cui si apre la finestra ('anc' | 'bf' | 'rh5' | 'kdhx' | 'bj9000' | 'kdh9g'), radioKind: scheda d'origine
  open(expect, radioKind) {
    const TARGET = { anc: 'Bajeton BJ-8300 / BJ-7800 / BJ-7700', bf: 'Baofeng UV-5RM / UV-5RH e radio con lo stesso caricatore', rh5: 'Baofeng 5RH Pro / K67 Pro',
      kdhx: 'Baofeng UV-25 Plus e radio con il caricatore «Bootloader Pro»', bj9000: 'Bajeton BJ-9000', kdh9g: 'Baofeng UV-9G' };
    const PROTO = {
      anc: 'quello del programma originale «BinTe ANC Bootloader» V1.1',
      bf: 'quello del programma originale «BF Upgrade Tool for TFT 330L Series» (KDH Bootloader V1.0)',
      rh5: 'quello del programma originale di aggiornamento della 5RH Pro (Radio.exe)',
      kdhx: 'quello del programma originale «Baofeng Bootloader Pro»',
      bj9000: 'quello del programma originale «BJ9000_EnUPDATE» (BJ9000_UPDATE_V02)',
      kdh9g: 'quello del programma originale «Bootloader_BaudRate38400» della UV-9G',
    };
    // caricatori KDH recenti: istruzioni del costruttore e velocità
    const BOOT = {
      kdhx: { manual: 'radio spenta: tieni premuti <b>SK2 e SK3</b> (i due tasti sotto il PTT) e accendila; l\'indicatore diventa verde. È la procedura indicata da Baofeng per la UV-25 Plus (per le altre radio segui le istruzioni fornite con il firmware)',
        auto: 'la radio accesa viene portata nel caricatore dal programma (pulsante «自动升级»); prima 9600 baud, poi 115200', speed: 'Velocità come nel programma originale: 9600 baud per la richiesta, 115200 per il firmware.' },
      bj9000: { manual: 'radio spenta: tieni premuti i <b>tasti laterali 3 e 4</b> e accendila; compare la schermata «Update» (pulsante «Flashing Mode»)',
        auto: 'la radio accesa viene portata nel caricatore dal programma (pulsante «Upgrade Mode»)', speed: 'Velocità fissa 115200 baud, come nel programma originale. La radio controlla che il firmware sia per il suo modello prima di cancellare la memoria.' },
      kdh9g: { manual: 'radio spenta: tieni premuti i <b>tasti 12 e 13</b> e accendila; si accende la torcia. È la procedura indicata da Baofeng',
        auto: 'la radio accesa viene portata nel caricatore dal programma (prima 9600 baud, poi 38400)', speed: 'Velocità come nel programma originale: 38400 baud per il firmware.' },
    };
    const isBoot = k => !!BOOT[k];
    // radio per la copia di sicurezza: .BF e .kdhx insieme (stesse famiglie di radio)
    const fwFam = k => (k === 'bf' || k === 'kdhx') ? ['bf', 'kdhx'] : [k];
    const fwModels = k => {
      const seen = new Map();
      App.catalog.filter(c => fwFam(k).includes(c.fw) && (c.drivers || []).some(d => App.defs[d])).forEach(c => {
        const id = c.drivers.find(d => App.defs[d]);
        seen.set(id, seen.has(id) ? seen.get(id) + ' / ' + c.name : c.brand + ' ' + c.name);
      });
      return [...seen].map(([id, name]) => ({ id, name }));
    };
    const bfModels = fwModels('bf');
    let fw = null, running = false, timer = null, identified = null, settings = {};
    const kind = () => (fw && fw.kind) || expect || '';
    const radioName = () => (radioKind && App.defs[radioKind] && App.defs[radioKind].name) || TARGET[expect] || 'questa radio';
    const m = openModal(`<div class="dialog mid fwdlg"><header>⬆ Aggiornamento del firmware${expect ? ' — ' + esc(TARGET[expect]) : ''}</header><div class="body">
      <div data-setup>
        <h4 class="fwh"><span class="num">1</span> Firmware</h4>
        <div class="row"><button class="btn" data-file>📂 File sul PC…</button>
          <span class="note">oppure da Internet:</span>
          <input type="text" data-url placeholder="https://… (anche link di Google Drive o Dropbox)" style="flex:1;min-width:260px">
          <button class="btn" data-down>🌐 Scarica</button></div>
        <p class="note" style="margin-top:-4px">Firmware riconosciuti: <b>.anc</b> delle Bajeton BJ-8300, BJ-7800 e BJ-7700 con Bluetooth (es. BJ8300_V0.12_20251112.anc, BJ7700PRO_V0.06_251124.anc),
          <b>.Firmware</b> della Bajeton BJ-9000 (BJ9000_V0.27_260313.Firmware), <b>.BF</b> delle Baofeng UV-5RM/UV-5RH (es. 5RH_BK_V0.14_231123.BF), <b>.dat</b> della Baofeng 5RH Pro
          (es. BF_5RH_501_v2_0_9.dat), <b>.kdhx</b> della UV-25 Plus (UV25Pro_NRF_401+_V0.23_250522.kdhx) e <b>.kdh</b> della UV-9G (T52GMRS_M8_V02_231004.kdh).
          Possono essere anche dentro archivi <b>.zip</b> o <b>.rar</b>, pure uno dentro l'altro.</p>
        <div data-fwcard class="fwcard" style="display:none"></div>

        <h4 class="fwh"><span class="num">2</span> Radio</h4>
        <div data-nokind class="note">Scegli prima il file del firmware: le opzioni dipendono dalla radio a cui è destinato.</div>
        <div data-common style="display:none">
          <div class="row"><label class="l">Porta del cavo</label><select data-port style="min-width:300px"></select>
            <button class="btn" data-refresh title="Cerca di nuovo le porte">🔄 Aggiorna</button></div>
        </div>
        <div data-panel="anc" style="display:none">
          <div class="row radiolist"><label class="l">Stato della radio</label><div>
            <label><input type="radio" name="fwmode" value="update" checked> <b>Funziona normalmente</b> — viene identificata e portata in modalità aggiornamento
              <span class="note">(pulsante «Update» del programma originale)</span></label>
            <label><input type="radio" name="fwmode" value="recovery"> <b>Non si avvia</b> o un aggiornamento precedente si è interrotto — modalità di recupero
              <span class="note">(pulsante «Click Here If Your Radio Can't Work»)</span></label></div></div>
          <div data-upd>
            <div class="row"><label class="l">Identificazione</label><button class="btn" data-ident>🔍 Identifica la radio</button>
              <span data-identres class="note">la radio viene comunque identificata all'avvio dell'aggiornamento</span></div>
            <div class="row"><label class="l">Sicurezza</label><label><input type="checkbox" data-backup> prima dell'aggiornamento leggi la radio e salva una copia dei canali e delle impostazioni (.dat per la BJ-8300, .img per BJ-7800 e BJ-7700)</label></div>
          </div>
          <div data-rec class="fwbox" style="display:none">
            <b>Modalità di recupero.</b> La radio non può essere identificata: il programma (come quello originale) attende per circa 12 secondi
            che la radio si presenti in modalità aggiornamento.
            <ol><li>Collega il cavo alla radio <b>spenta</b>.</li>
            <li>Premi «Avvia aggiornamento».</li>
            <li>Subito dopo porta la radio in modalità aggiornamento secondo le istruzioni del rivenditore (di solito accendendola tenendo premuto un tasto, spesso il PTT).</li></ol>
          </div>
          <div class="row"><label class="l">Velocità</label><select data-baud><option value="115200">115200 baud (predefinita del programma originale)</option><option value="57600">57600 baud</option></select>
            <span class="note">velocità del trasferimento dei dati</span></div>
        </div>
        <div data-panel="bf" style="display:none">
          <div class="row radiolist"><label class="l">Stato della radio</label><div>
            <label><input type="radio" name="bfmode" value="auto" checked> <b>Accesa normalmente</b> — viene portata nel caricatore dal programma
              <span class="note">(pulsante «自动升级», aggiornamento automatico)</span></label>
            <label><input type="radio" name="bfmode" value="manual"> <b>Già nel caricatore</b>: spenta e riaccesa tenendo premuto il <b>PTT</b> (schermo spento) — anche per una radio che non si avvia
              <span class="note">(pulsante «手动升级», aggiornamento manuale)</span></label></div></div>
          <div class="row" data-bfbackup><label class="l">Sicurezza</label><label><input type="checkbox" data-bfbk> prima leggi la radio e salva una copia dei canali (.img) come</label>
            <select data-bfmodel>${bfModels.map(b => `<option value="${esc(b.id)}">${esc(b.name)}</option>`).join('')}</select></div>
          <div data-bfman class="fwbox" style="display:none"><b>Modalità manuale.</b>
            <ol><li>Spegni la radio e collega il cavo.</li><li>Riaccendila tenendo premuto il <b>PTT</b>: lo schermo resta spento (procedura della UV-5RM/5RH; per le altre radio segui le istruzioni fornite con il firmware).</li><li>Premi «Avvia aggiornamento».</li></ol></div>
          <p class="note">Velocità fissa 115200 baud, come nel programma originale.</p>
        </div>
        <div data-panel="rh5" style="display:none">
          <div class="fwbox"><b>Porta la radio in modalità aggiornamento</b> come indicato dal costruttore (di solito: radio spenta, accenderla tenendo premuti i tasti
            laterali e il PTT; lo schermo resta spento), collega il cavo e poi premi «Avvia aggiornamento». La radio dichiara la sua elettronica (V1 o V2):
            se non è quella del firmware scelto l'aggiornamento non parte.</div>
          <div class="row"><label class="l">Velocità</label><select data-rhbaud><option value="115200">115200 baud (predefinita)</option><option value="38400">38400 baud</option></select></div>
          <p class="note">La copia dei canali, se ti serve, va fatta prima dalla scheda della radio («Leggi dalla radio»), con la radio accesa normalmente.</p>
        </div>

        <div data-panel="boot" style="display:none">
          <div class="row radiolist"><label class="l">Stato della radio</label><div>
            <label><input type="radio" name="btmode" value="manual" checked> <b>Nel caricatore</b> — <span data-btman></span>
              <span class="note">(aggiornamento manuale; anche per una radio che non si avvia)</span></label>
            <label><input type="radio" name="btmode" value="auto"> <b>Accesa normalmente</b> — <span data-btauto></span></label></div></div>
          <div class="row" data-btbackup><label class="l">Sicurezza</label><label><input type="checkbox" data-btbk> prima leggi la radio e salva una copia dei canali (.img) come</label>
            <select data-btmodel></select></div>
          <div data-btmanbox class="fwbox"><b>Modalità manuale.</b>
            <ol><li>Collega il cavo alla radio spenta.</li><li data-btmanstep></li><li>Premi «Avvia aggiornamento».</li></ol></div>
          <p class="note" data-btspeed></p>
        </div>

        <h4 class="fwh"><span class="num">3</span> Conferma</h4>
        <div data-warnok style="display:none" class="row"><label><input type="checkbox" data-acceptwarn> <b>Ho letto le segnalazioni sul firmware</b> e voglio usarlo comunque</label></div>
        <div class="row"><label><input type="checkbox" data-confirm> <b data-conftext>Confermo il modello della radio collegata.</b>
          <span class="note">Un firmware per un'altra radio può renderla inutilizzabile.</span></label></div>
        <div class="row" data-skiprow><label><input type="checkbox" data-skip> salta l'identificazione automatica <span class="note">(solo se la radio funziona ma non risponde alla lettura via cavo)</span></label></div>
        <div class="row" data-typerow style="display:none"><label class="l" data-typelabel>Scrivi BJ8300</label><input type="text" data-typed maxlength="10" style="width:120px" autocomplete="off">
          <span class="note">conferma richiesta quando il programma non può verificare il modello della radio</span></div>
        <p class="fwwarn">⚠ Durante l'aggiornamento <b>non spegnere la radio e non scollegare il cavo</b>; usa una batteria carica. Il produttore non documenta la procedura:
          il protocollo è <span data-proto>quello del programma originale del costruttore</span>, ricavato dal suo codice e verificato con radio simulate.</p>
      </div>
      <div data-run style="display:none">
        <div style="font-weight:600;margin-bottom:6px" data-phase></div>
        <div class="pbar"><div data-bar></div></div>
        <div class="note" data-msg style="margin-top:6px"></div>
        <pre class="fwlog" data-log></pre>
        <p data-final style="margin-top:12px;display:none"></p>
        <ul class="warnlist" data-notes style="display:none"></ul>
      </div>
      </div><footer><span class="left note" data-foot></span><button class="btn" data-no>Chiudi</button><button class="btn primary" data-ok disabled>⬆ Avvia aggiornamento</button></footer></div>`,
      { onEsc: () => { if (!running) close(); } });
    const el = m.el;
    const close = () => { clearTimeout(timer); m.close(); };
    const sel = $('[data-port]', el);
    const mode = () => $('input[name=fwmode]:checked', el).value;
    const bfmode = () => $('input[name=bfmode]:checked', el).value;
    const btmode = () => $('input[name=btmode]:checked', el).value;
    if (radioKind && bfModels.some(b => b.id === radioKind)) $('[data-bfmodel]', el).value = radioKind;
    let btFor = '';
    // pannello dei caricatori KDH recenti per il tipo di firmware k
    const setupBoot = (k) => {
      if (btFor === k || !isBoot(k)) return;
      btFor = k;
      const b = BOOT[k];
      $('[data-btman]', el).innerHTML = b.manual;
      $('[data-btmanstep]', el).innerHTML = b.manual.charAt(0).toUpperCase() + b.manual.slice(1) + '.';
      $('[data-btauto]', el).innerHTML = b.auto;
      $('[data-btspeed]', el).textContent = b.speed;
      const ms = fwModels(k);
      $('[data-btmodel]', el).innerHTML = ms.map(x => `<option value="${esc(x.id)}">${esc(x.name)}</option>`).join('');
      if (radioKind && ms.some(x => x.id === radioKind)) $('[data-btmodel]', el).value = radioKind;
      $('[data-btbackup]', el).dataset.n = ms.length;
    };

    const needTyped = () => (kind() === 'anc' && (mode() === 'recovery' || $('[data-skip]', el).checked)) ||
      kind() === 'kdh9g' || (kind() === 'kdhx' && btmode() === 'manual') || !!(fw && fw.typed) || otherModel();
    // radio scelta nella finestra diversa da quelle a cui è destinato il file
    const otherModel = () => {
      if (!fw || !fw.models || !fw.models.length) return false;
      const x = $(kind() === 'bf' ? '[data-bfmodel]' : '[data-btmodel]', el);
      return !!(x && x.value && !fw.models.includes(x.value));
    };
    const refresh = () => {
      const k = kind();
      $('[data-nokind]', el).style.display = k ? 'none' : '';
      $('[data-common]', el).style.display = k ? '' : 'none';
      setupBoot(k);
      $$('[data-panel]', el).forEach(p => p.style.display = p.dataset.panel === (isBoot(k) ? 'boot' : k) ? '' : 'none');
      const rec = mode() === 'recovery';
      $('[data-upd]', el).style.display = rec ? 'none' : '';
      $('[data-rec]', el).style.display = rec ? '' : 'none';
      $('[data-skiprow]', el).style.display = k === 'anc' && !rec ? '' : 'none';
      $('[data-typerow]', el).style.display = needTyped() ? '' : 'none';
      $('[data-bfbackup]', el).style.display = bfmode() === 'auto' && bfModels.length ? '' : 'none';
      $('[data-bfman]', el).style.display = bfmode() === 'manual' ? '' : 'none';
      $('[data-btbackup]', el).style.display = btmode() === 'auto' && +($('[data-btbackup]', el).dataset.n || 0) ? '' : 'none';
      $('[data-btmanbox]', el).style.display = btmode() === 'manual' ? '' : 'none';
      $('[data-warnok]', el).style.display = fw && fw.warnings ? '' : 'none';
      $('[data-proto]', el).textContent = PROTO[k] || 'quello del programma originale del costruttore';
      const target = fw ? fw.target : TARGET[k];
      $('[data-conftext]', el).textContent = k === 'kdh9g'
        ? `Confermo che la radio collegata è una ${target}: il suo caricatore non controlla il firmware che riceve.`
        : k === 'bf'
        ? `Confermo che la radio collegata è una ${target}${fw && fw.platform ? ' con chip radio ' + fw.platform : ''} e che questo firmware è quello adatto.`
        : target ? `Confermo che la radio collegata è una ${target}.` : 'Confermo il modello della radio collegata.';
      const want = (fw && fw.confirm) || 'BJ8300';
      $('[data-typelabel]', el).textContent = 'Scrivi ' + want;
      const typedOk = !needTyped() || $('[data-typed]', el).value.toUpperCase().replace(/[\s\-_.]/g, '') === want;
      const ok = fw && fw.valid && sel.value && $('[data-confirm]', el).checked && typedOk && (!fw.warnings || $('[data-acceptwarn]', el).checked);
      $('[data-ok]', el).disabled = !ok || running;
      $('[data-ident]', el).disabled = !sel.value || running;
    };
    $$('input,select', el).forEach(x => x.addEventListener('change', refresh));
    // cambiando porta l'esito dell'identificazione precedente non vale più
    sel.addEventListener('change', () => {
      identified = null;
      const res = $('[data-identres]', el);
      res.className = 'note';
      res.textContent = 'la radio viene comunque identificata all\'avvio dell\'aggiornamento';
    });
    $('[data-typed]', el).addEventListener('input', refresh);

    const showFw = (r) => {
      fw = r.firmware;
      const card = $('[data-fwcard]', el);
      const icon = { ok: '✔', warn: '⚠', bad: '✖' };
      card.style.display = '';
      card.className = 'fwcard ' + (!fw.valid ? 'bad' : fw.warnings ? 'warn' : 'ok');
      // radio della scheda d'origine diversa da quella indicata dal file
      const other = radioKind && fw.models && fw.models.length ? !fw.models.includes(radioKind) : expect && fw.kind !== expect;
      card.innerHTML = `<div class="fwtitle">${fw.valid ? (fw.warnings ? '⚠' : '✅') : '❌'} <b>${esc(fw.name)}</b>
          ${fw.version ? `<span class="tag">${esc(fw.version)}</span>` : ''}${fw.date ? `<span class="tag">${esc(fw.date)}</span>` : ''}</div>
        <div class="note">Per: <b>${esc(fw.target)}</b>${fw.hw ? ' · ' + esc(fw.hw) : ''}</div>
        ${other ? `<div class="fwbad">Attenzione: questo firmware è per ${esc(fw.target)}, non per ${esc(radioName())} (la radio della scheda da cui hai aperto questa finestra).</div>` : ''}
        <div class="note">${esc(fmtSize(fw.size))} · da ${esc(fw.origin)}${fw.saved ? ' · copia salvata in ' + esc(fw.saved) : ''}</div>
        <div class="note">SHA-256 ${esc(fw.sha256)}</div>
        <ul class="fwchecks">${fw.checks.map(c => `<li class="${c.level}"><span>${icon[c.level]}</span> ${esc(c.text)}</li>`).join('')}</ul>
        ${r.skipped && r.skipped.length ? `<div class="note">File non usati: ${esc(r.skipped.join('; '))}</div>` : ''}
        ${!fw.valid ? '<p style="margin:6px 0 0"><b>Questo file non può essere inviato alla radio.</b></p>' : ''}`;
      $('[data-acceptwarn]', el).checked = false;
      $('[data-confirm]', el).checked = false;
      $('[data-typed]', el).value = '';
      refresh();
      // copia di sicurezza: la radio indicata dal file, se non si è partiti da una scheda
      const pre = (fw.models || []).find(id => id === radioKind) || (fw.models || [])[0];
      for (const s of ['[data-bfmodel]', '[data-btmodel]']) {
        const x = $(s, el);
        if (pre && x && [...x.options].some(o => o.value === pre) && !(radioKind && [...x.options].some(o => o.value === radioKind))) x.value = pre;
      }
    };
    const loadFw = async (req) => {
      try {
        const r = await Firmware.load(req);
        if (r) showFw(r);
      } catch (e) { reportError(e); }
    };
    $('[data-file]', el).onclick = () => loadFw({ source: 'file' });
    $('[data-down]', el).onclick = () => {
      const u = $('[data-url]', el).value.trim();
      if (!u) { toast('Scrivi l\'indirizzo del firmware.', true); return; }
      loadFw({ source: 'url', url: u });
    };
    $('[data-url]', el).addEventListener('keydown', e => { if (e.key === 'Enter') $('[data-down]', el).click(); });

    const loadPorts = async () => {
      const r = await api('cable/ports', {}, { quiet: true });
      settings = r.settings || {};
      const cur = sel.value || (kind() === 'anc' ? settings.bajPort : kind() === 'rh5' ? settings.k67Port || settings.genPort : settings.genPort) || settings.bajPort;
      sel.innerHTML = r.ports.map(p => `<option value="${esc(p.name)}">${esc(p.desc && p.desc !== p.name ? p.desc : p.name)}</option>`).join('');
      if (r.ports.some(p => p.name === cur)) sel.value = cur;
      if (!r.ports.length) sel.innerHTML = '<option value="">nessuna porta trovata: collega il cavo e premi Aggiorna</option>';
      refresh();
    };
    loadPorts().then(() => {
      $('[data-url]', el).value = settings.fwURL || '';
      $('[data-baud]', el).value = String(settings.fwBaud === 57600 ? 57600 : 115200);
      $('[data-backup]', el).checked = !settings.fwNoBackup;
      $('[data-bfbk]', el).checked = !settings.fwNoBackup;
      $('[data-btbk]', el).checked = !settings.fwNoBackup;
    }).catch(reportError);
    $('[data-refresh]', el).onclick = () => loadPorts().catch(reportError);

    // ---------------------------------------------------------- operazioni in background
    const poll = (onDone, onTick) => {
      const tick = async () => {
        let st;
        try { st = await api('cable/status', {}, { quiet: true }); } catch (e) { timer = setTimeout(tick, 500); return; }
        if (onTick) onTick(st);
        if (st.state === 'running') { timer = setTimeout(tick, 250); return; }
        onDone(st);
      };
      tick();
    };

    $('[data-ident]', el).onclick = async () => {
      const res = $('[data-identres]', el);
      try {
        running = true; refresh();
        res.className = 'note'; res.textContent = 'Identificazione in corso: accendi la radio…';
        await api('fw/identify', { port: sel.value }, { quiet: true });
        poll(st => {
          running = false;
          if (st.state === 'done') {
            identified = st.result;
            res.className = 'fwok';
            res.textContent = `✔ Risponde al protocollo ${st.result.proto || 'Bajeton BT-11'} su ${st.result.port} — identificativo «${st.result.id}»`;
          } else {
            identified = null;
            res.className = 'fwbad';
            res.textContent = '✖ ' + (st.error || 'Identificazione non riuscita');
          }
          refresh();
        });
      } catch (e) { running = false; refresh(); res.textContent = ''; reportError(e); }
    };

    const setNotes = (notes) => {
      const ul = $('[data-notes]', el);
      ul.style.display = notes && notes.length ? '' : 'none';
      ul.innerHTML = (notes || []).map(x => '<li>' + esc(x) + '</li>').join('');
    };
    const showLog = (log) => {
      const pre = $('[data-log]', el);
      const atEnd = pre.scrollTop + pre.clientHeight >= pre.scrollHeight - 4;
      pre.textContent = (log || []).join('\n');
      if (atEnd) pre.scrollTop = pre.scrollHeight;
    };
    const tick = (st) => {
      $('[data-phase]', el).textContent = st.phase || '';
      const pct = st.total ? Math.round(st.done * 100 / st.total) : 0;
      $('[data-bar]', el).style.width = pct + '%';
      $('[data-msg]', el).textContent = (st.msg || '') + (st.total > 1 ? ` ${pct}%` : '');
      showLog(st.log);
      const no = $('[data-no]', el);
      no.disabled = !st.canStop;
      no.textContent = 'Interrompi';
      $('[data-foot]', el).textContent = st.state === 'running' && !st.canStop ? 'Non spegnere la radio e non scollegare il cavo!' : '';
    };
    const done = (st) => {
      running = false;
      tick(st);
      const ok = st.state === 'done';
      const fin = $('[data-final]', el);
      fin.style.display = '';
      $('[data-bar]', el).classList.toggle('err', !ok);
      if (ok) {
        $('[data-bar]', el).style.width = '100%';
        $('[data-phase]', el).textContent = 'Aggiornamento completato';
        fin.innerHTML = '✅ <b>Aggiornamento del firmware completato.</b> Scollega il cavo, spegni e riaccendi la radio e controlla la versione all\'accensione.' +
          (st.result && st.result.backup ? '<br><span class="note">Se canali o impostazioni fossero cambiati, scrivi nella radio la copia salvata prima dell\'aggiornamento.</span>' : '');
      } else {
        fin.innerHTML = (st.state === 'cancelled' ? '⏹ ' : '❌ ') + esc(st.error || 'Operazione non riuscita');
      }
      setNotes(st.notes);
      $('[data-msg]', el).textContent = '';
      $('[data-foot]', el).textContent = '';
      const no = $('[data-no]', el);
      no.textContent = 'Chiudi'; no.disabled = false; no.classList.add('primary');
      const okb = $('[data-ok]', el);
      okb.style.display = ok ? 'none' : '';
      okb.textContent = '↩ Torna alle impostazioni';
      okb.disabled = false;
      okb.dataset.back = '1';
    };

    $('[data-no]', el).onclick = async () => {
      if (!running) { close(); return; }
      try { await api('cable/cancel', {}, { quiet: true }); } catch (e) { reportError(e); }
    };
    $('[data-ok]', el).onclick = async () => {
      const okb = $('[data-ok]', el);
      if (okb.dataset.back) {
        delete okb.dataset.back;
        okb.textContent = '⬆ Avvia aggiornamento';
        $('[data-setup]', el).style.display = ''; $('[data-run]', el).style.display = 'none';
        const no = $('[data-no]', el); no.textContent = 'Chiudi'; no.classList.remove('primary');
        refresh();
        return;
      }
      if (running || !fw) return;
      const k = fw.kind;
      const base = { port: sel.value, sha256: fw.sha256, confirmModel: $('[data-confirm]', el).checked, acceptWarnings: $('[data-acceptwarn]', el).checked };
      let req, how;
      if (k === 'bf') {
        const auto = bfmode() === 'auto';
        req = Object.assign(base, { mode: bfmode(), backup: auto && bfModels.length > 0 && $('[data-bfbk]', el).checked, model: $('[data-bfmodel]', el).value, confirmText: $('[data-typed]', el).value });
        how = auto ? 'La radio accesa verrà portata nel caricatore' + (req.backup ? ' dopo averne salvato una copia dei canali' : '') + '.\n'
          : 'La radio deve essere già nel caricatore (accesa tenendo premuto il PTT, schermo spento).\n';
      } else if (isBoot(k)) {
        const auto = btmode() === 'auto';
        req = Object.assign(base, { mode: btmode(), backup: auto && +($('[data-btbackup]', el).dataset.n || 0) > 0 && $('[data-btbk]', el).checked,
          model: $('[data-btmodel]', el).value, confirmText: $('[data-typed]', el).value });
        how = auto ? 'La radio accesa verrà portata nel caricatore' + (req.backup ? ' dopo averne salvato una copia dei canali' : '') + '.\n'
          : 'La radio deve essere già nel caricatore (' + BOOT[k].manual.replace(/<[^>]+>/g, '') + ').\n';
      } else if (k === 'rh5') {
        req = Object.assign(base, { mode: 'update', baud: +$('[data-rhbaud]', el).value });
        how = 'La radio deve essere già in modalità aggiornamento.\n';
      } else {
        const rec = mode() === 'recovery';
        req = Object.assign(base, { mode: mode(), baud: +$('[data-baud]', el).value, backup: !rec && $('[data-backup]', el).checked,
          skipIdentify: !rec && $('[data-skip]', el).checked, confirmText: $('[data-typed]', el).value });
        how = rec ? 'Modalità di recupero: dopo aver premuto «Aggiorna», porta subito la radio in modalità aggiornamento.\n'
          : 'La radio verrà prima identificata' + (req.skipIdentify ? ' (identificazione saltata su tua richiesta)' : '') + (req.backup ? ' e ne verrà salvata una copia' : '') + '; poi verrà portata in modalità aggiornamento.\n';
      }
      const what = `${fw.name}${fw.version ? ' (' + fw.version + ')' : ''}`;
      const c = await chooseDlg('Conferma l\'aggiornamento del firmware',
        `Vuoi installare ${what} nella ${fw.target} collegata a ${req.port}?\n` + how +
        'Durante l\'aggiornamento non spegnere la radio e non scollegare il cavo.',
        [{ id: 'no', label: 'Annulla' }, { id: 'ok', label: 'Aggiorna', primary: true }]);
      if (c !== 'ok') return;
      try {
        running = true;
        $('[data-setup]', el).style.display = 'none'; $('[data-run]', el).style.display = '';
        $('[data-final]', el).style.display = 'none';
        $('[data-bar]', el).classList.remove('err');
        $('[data-bar]', el).style.width = '0%';
        $('[data-log]', el).textContent = '';
        setNotes([]);
        okb.disabled = true;
        $('[data-no]', el).classList.remove('primary');
        await api('fw/start', req, { quiet: true });
        poll(done, tick);
      } catch (e) {
        running = false;
        $('[data-setup]', el).style.display = ''; $('[data-run]', el).style.display = 'none';
        $('[data-no]', el).textContent = 'Chiudi';
        refresh();
        reportError(e);
      }
    };
    refresh();
  },
};

function fmtSize(n) {
  if (n >= 1048576) return (n / 1048576).toFixed(2).replace('.', ',') + ' MB';
  if (n >= 1024) return (n / 1024).toFixed(1).replace('.', ',') + ' KB';
  return n + ' byte';
}
