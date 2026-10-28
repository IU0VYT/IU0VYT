'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda «Info»: versione, autore, licenza, collegamenti e verifica degli aggiornamenti.
// Aggiornamenti: il programma confronta la sua versione con quella pubblicata
// (indirizzo fisso, nel programma) a ogni avvio e con il pulsante «Verifica
// aggiornamenti». La verifica legge solo il numero di versione; il pacchetto si
// scarica solo se è più recente e solo quando lo chiedi.

const Updates = {
  last: null,      // esito dell'ultima verifica
  running: null,   // verifica in corso (promessa)
  listeners: [],

  // verifica: { auto } = controllo automatico all'avvio (nessun messaggio d'errore)
  async check(opts) {
    opts = opts || {};
    if (this.running) return this.running;
    this.running = (async () => {
      try {
        const r = await api('update/check', {}, { quiet: true });
        this.last = r;
        if (r.cmp > 0) this.showBanner(r, !!opts.auto);
        else this.hideBanner();
        this.listeners.forEach(f => f(r));
        return r;
      } finally {
        this.running = null;
        const b = $('#btn-upd');
        if (b) b.classList.remove('spin');
      }
    })();
    const b = $('#btn-upd');
    if (b) b.classList.add('spin');
    return this.running;
  },

  // pulsante della barra in alto
  async manual() {
    try {
      const r = await this.check();
      if (r.cmp > 0) {
        toast(`È disponibile la versione ${r.remote}: usa «Scarica l'aggiornamento».`);
      } else {
        toast(`✔ Il software è aggiornato: la versione ${r.local} è la più recente.`);
      }
    } catch (e) { reportError(new Error('Verifica degli aggiornamenti non riuscita: ' + e.message)); }
  },

  showBanner(r, auto) {
    // chiuso dall'utente in questa sessione per questa versione: all'avvio non si ripropone
    try { if (auto && sessionStorage.getItem('upd-hide') === r.remote) return; } catch (e) { /* */ }
    let b = $('#updbanner');
    if (!b) {
      b = h('<div id="updbanner" role="status"></div>');
      $('#topbar').after(b);
    }
    const built = r.remoteBuild ? ` del ${esc(r.remoteBuild)}` : '';
    b.innerHTML = `<span class="updicon">⬆</span>
      <span class="updtext"><b>È disponibile una nuova versione di IU0VYT Flash: la ${esc(r.remote)}</b>${built}. Stai usando la ${esc(r.local)}.</span>
      <button class="btn primary" data-dl>⬇ Scarica l'aggiornamento (${esc(fmtSize(r.size))})</button>
      <button class="btn" data-more>Dettagli</button>
      <button class="updx" data-x title="Nascondi (il controllo si ripete al prossimo avvio)">×</button>`;
    $('[data-dl]', b).onclick = () => this.download();
    $('[data-more]', b).onclick = () => App.showTab('info');
    $('[data-x]', b).onclick = () => {
      try { sessionStorage.setItem('upd-hide', r.remote); } catch (e) { /* */ }
      this.hideBanner();
    };
  },

  hideBanner() {
    const b = $('#updbanner');
    if (b) b.remove();
  },

  async download() {
    try {
      const r = await api('update/save', {}, { save: true, suggest: '', busyText: 'Download dell\'aggiornamento…' });
      this.hideBanner();
      const m = openModal(`<div class="dialog mid"><header>⬇ Aggiornamento scaricato</header><div class="body">
        <p>La versione <b>${esc(r.version)}</b> è stata salvata in<br><b>${esc(r.path)}</b></p>
        <ol><li>Chiudi IU0VYT Flash.</li><li>Estrai l'archivio .zip.</li>
        <li>Sostituisci il vecchio «IU0VYT Flash.exe» con quello nuovo.</li></ol>
        <p class="note">Archivio ponti, impostazioni e copie di sicurezza restano dove sono.</p></div>
        <footer><button class="btn" data-show>📂 Mostra nella cartella</button><button class="btn primary" data-no>OK</button></footer></div>`);
      $('[data-no]', m.el).onclick = () => m.close();
      $('[data-show]', m.el).onclick = async () => {
        try { await api('open', { file: r.path }, { quiet: true }); } catch (er) { reportError(er); }
      };
    } catch (e) { reportError(e); }
  },
};

const Info = {
  built: false,
  data: null,

  // apre un collegamento nel browser predefinito (dal programma su Windows, altrimenti dalla pagina)
  async open(url) {
    try {
      const r = await api('open', { url }, { quiet: true });
      if (r && r.opened) return;
    } catch (e) { /* si prova dalla pagina */ }
    window.open(url, '_blank', 'noopener');
  },

  async render(page) {
    if (this.built) return;
    this.built = true;
    let d;
    try { d = this.data = await api('info', {}, { quiet: true }); } catch (e) { reportError(e); this.built = false; return; }
    page.innerHTML = `<div class="infopage"><div class="infocard">
      <div class="infohead"><img src="icona.png" alt="" class="infoicon">
        <div><h1>${esc(d.name)}</h1><div class="note">Canali e ponti delle radio Baofeng e Bajeton · archivio ponti · aggiornamento del firmware</div></div></div>
      <table class="infotab">
        <tr><th>Versione</th><td><b>${esc(d.version)}</b></td></tr>
        <tr><th>Data di compilazione</th><td>${esc(d.buildDate)}</td></tr>
        <tr><th>Autore</th><td>${esc(d.author)}</td></tr>
        <tr><th>E-mail</th><td><button class="btn" data-mail>✉ Scrivi all'autore</button>
          <span class="note">si apre il tuo programma di posta con il destinatario già impostato</span>
          <div class="note" data-mailhelp style="display:none;margin-top:6px">Non si è aperto nessun programma di posta?
            <a href="#" data-mailalt>Altri modi per scrivere</a></div>
          <div class="mailalt" data-mailbox style="display:none">
            <p>Se sul PC non c'è un programma di posta, usa quella del browser (Gmail, Outlook.com…): copia l'indirizzo e incollalo nel campo «A».
              L'indirizzo non è scritto nella pagina, per proteggerlo da spam e phishing.</p>
            <button class="btn primary" data-mailcopy>📋 Copia l'indirizzo</button> <button class="btn" data-mailshow>👁 Mostra l'indirizzo</button>
            <div data-mailimg style="margin-top:8px"></div></div></td></tr>
        <tr><th>QRZ</th><td><a href="#" data-link="${esc(d.qrz)}">${esc(d.qrz)}</a></td></tr>
        <tr><th>Codice sorgente</th><td><a href="#" data-link="${esc(d.repo)}">${esc(d.repo)}</a> <span class="note">(repository GitHub: sorgenti, rilasci e aggiornamenti)</span></td></tr>
        <tr><th>Licenza</th><td>${esc(d.license)} — software libero, <b>senza alcuna garanzia</b> (testo completo nel file LICENSE.txt del pacchetto)</td></tr>
        <tr><th>Cartella dei dati</th><td>${d.dataDirShort ? `<b>${esc(d.dataDirShort)}</b><br>` : ''}<span class="note">${esc(d.dataDir)}</span>
          <button class="btn" data-folder style="margin-left:6px">📂 Apri</button>
          <div class="note">cartella dell'utente di Windows che usa il programma, su questo PC: archivio dei ponti, impostazioni e copie delle radio</div></td></tr>
      </table>
      <h3>Aggiornamenti</h3>
      <div class="updbox">
        <button class="btn primary big" data-check>🔄 Verifica aggiornamenti</button>
        <div class="note">Il programma controlla da solo a ogni avvio se esiste una versione più recente di quella in uso (${esc(d.version)}).
          Il controllo legge solo il numero di versione pubblicato: il programma nuovo si scarica solo se è più recente e solo quando lo chiedi.</div>
      </div>
      <div data-res class="infores note" style="display:none"></div>
      <h3>Note legali</h3>
      <p class="note">Copyright © 2026 ${esc(d.author)}. Questo programma è software libero: puoi ridistribuirlo e/o modificarlo secondo i termini della
        GNU General Public License pubblicata dalla Free Software Foundation, versione 3 o (a tua scelta) successiva. È distribuito nella speranza che sia utile,
        ma SENZA ALCUNA GARANZIA, neppure implicita di COMMERCIABILITÀ o IDONEITÀ A UN FINE PARTICOLARE.</p>
      <p class="note">I driver delle radio derivano in gran parte da CHIRP (chirpmyradio.com, © gli autori di CHIRP, GNU GPL). Altri componenti:
        WebView2 (Microsoft), go-webview2 (licenza MIT), go-winloader (licenza ISC), rardecode (archivi .rar, licenza BSD), librerie Go golang.org/x (licenza BSD).
        Elenco completo in AUTORI_E_LICENZE.txt.</p>
      <p class="note">Programma indipendente, non affiliato né approvato da Baofeng o da Bajeton: i nomi dei
        modelli servono solo a indicare con quali radio funziona. I marchi appartengono ai rispettivi titolari. I protocolli delle radio sono ricostruiti
        dai programmi originali dei costruttori per l'interoperabilità: conserva sempre le copie di sicurezza.</p>
    </div></div>`;
    $$('[data-link]', page).forEach(a => a.onclick = (e) => { e.preventDefault(); this.open(a.dataset.link); });
    $('[data-check]', page).onclick = () => this.check(page);
    this.contact(page);
    $('[data-folder]', page).onclick = async () => {
      try { const r = await api('contact', { action: 'folder' }, { quiet: true }); if (!r.opened) toast('Impossibile aprire la cartella.', true); } catch (e) { reportError(e); }
    };
    Updates.listeners.push(r => this.show(page, r));
    if (Updates.last) this.show(page, Updates.last);
  },

  // contatto con l'autore senza scrivere l'indirizzo nella pagina (3.9)
  contact(page) {
    const box = $('[data-mailbox]', page), help = $('[data-mailhelp]', page);
    const openBox = () => { box.style.display = ''; help.style.display = 'none'; };
    $('[data-mail]', page).onclick = async () => {
      try {
        const r = await api('contact', { action: 'mail' }, { quiet: true });
        if (r.opened) { toast('Si apre il programma di posta…'); help.style.display = ''; } else openBox();
      } catch (e) { openBox(); }
    };
    $('[data-mailalt]', page).onclick = (e) => { e.preventDefault(); openBox(); };
    $('[data-mailcopy]', page).onclick = async () => {
      try {
        const r = await api('contact', { action: 'copy' }, { quiet: true });
        if (r.copied) { toast('Indirizzo copiato negli appunti: incollalo nel campo «A» della tua posta (Ctrl+V).'); return; }
      } catch (e) { /* si mostra l'immagine */ }
      toast('Impossibile usare gli appunti: l\'indirizzo è mostrato come immagine.', true);
      this.mailImage(page);
    };
    $('[data-mailshow]', page).onclick = () => this.mailImage(page);
  },

  // l'indirizzo come immagine (non è testo della pagina: non si seleziona e non si legge dal codice della pagina)
  async mailImage(page) {
    const r = await api('contact', { action: 'show' }, { quiet: true });
    const c = document.createElement('canvas');
    const dpr = window.devicePixelRatio || 1;
    const ctx = c.getContext('2d');
    const font = '600 17px "Segoe UI", system-ui, sans-serif';
    ctx.font = font;
    const parts = [r.user, '@', r.domain];
    const w = Math.ceil(parts.reduce((t, p) => t + ctx.measureText(p).width, 0)) + 24;
    c.width = w * dpr; c.height = 34 * dpr; c.style.width = w + 'px'; c.style.height = '34px';
    c.className = 'mailcanvas';
    ctx.scale(dpr, dpr);
    ctx.fillStyle = '#f3f8ff'; ctx.fillRect(0, 0, w, 34);
    ctx.font = font; ctx.textBaseline = 'middle';
    let x = 12;
    parts.forEach((p, i) => { ctx.fillStyle = i === 1 ? '#c0392b' : '#1c3d73'; ctx.fillText(p, x, 18); x += ctx.measureText(p).width; });
    const holder = $('[data-mailimg]', page);
    holder.innerHTML = '';
    holder.appendChild(c);
    holder.appendChild(h('<div class="note">Scrivi l\'indirizzo a mano nel campo «A» della tua posta.</div>'));
  },

  show(page, r) {
    const res = $('[data-res]', page);
    if (!res) return;
    res.style.display = '';
    const built = r.remoteBuild ? ` (compilata il ${esc(r.remoteBuild)})` : '';
    if (r.cmp <= 0) {
      res.className = 'infores ok';
      res.innerHTML = r.cmp === 0
        ? `✔ <b>Il software è aggiornato</b>: stai usando la versione ${esc(r.local)}, la più recente${built}. Non c'è nulla da scaricare.`
        : `✔ <b>Il software è aggiornato</b>: la versione in uso (${esc(r.local)}) è più recente di quella pubblicata (${esc(r.remote)}). Non c'è nulla da scaricare.`;
    } else {
      res.className = 'infores new';
      res.innerHTML = `⬆ <b>È disponibile la versione ${esc(r.remote)}</b>${built}; stai usando la ${esc(r.local)}.
        <div style="margin-top:8px"><button class="btn primary" data-save>⬇ Scarica l'aggiornamento (${esc(fmtSize(r.size))})</button></div>`;
      $('[data-save]', res).onclick = () => Updates.download();
    }
  },

  async check(page) {
    const res = $('[data-res]', page);
    const btn = $('[data-check]', page);
    btn.disabled = true;
    res.style.display = '';
    res.className = 'infores note';
    res.textContent = 'Verifica in corso…';
    try {
      await Updates.check(); // l'esito arriva a show() tramite listeners
    } catch (e) {
      res.className = 'infores bad';
      res.textContent = '✖ Verifica non riuscita: ' + e.message;
    } finally { btn.disabled = false; }
  },
};
