'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Aggiornamento guidato del firmware BTECH DMR-6X2 / DMR-6X2 PRO (3.8): controllo
// dei file del pacchetto ufficiale, procedura dalle istruzioni BTECH («STOP and
// README FIRST.pdf») e avvio del programma ufficiale Firmware_Update. La
// scrittura del firmware la fa il programma BTECH.

const Fw6x2 = {
  steps: {
    main: [
      'Con il CPS della radio salva il codeplug in un file: l\'aggiornamento cancella la programmazione.',
      'Installa il driver «USB Virtual Com Port Driver» (cartella «USB Drivers» del pacchetto: versione x64 per Windows a 64 bit) e il CPS del pacchetto.',
      'Spegni la radio. Tieni premuti il <b>PTT</b> e il tasto <b>Alarm</b> (il tasto arancione sopra il display) e accendi la radio: il LED lampeggia in rosso (modalità aggiornamento). Rilascia i tasti.',
      'Collega la radio al PC con il cavo USB.',
      'Avvia il programma ufficiale (pulsante qui sotto, oppure dal CPS: Tool → Firmware Upgrade; se non parte, avvialo come amministratore).',
      '«Open Update File»: scegli il file .spi controllato qui sopra → messaggio «File opened successfully» → OK.',
      'Porta: quella del «USB Virtual Com Port»; velocità <b>921600</b>; spunta <b>Duplex</b>. Premi <b>Write</b> e attendi: alla fine la radio si riavvia. Se compare la schermata di impostazione dell\'orologio l\'aggiornamento potrebbe non essere riuscito: ripeti dal punto 3.',
      'Se richiesto dalle note di versione, aggiorna anche le icone (file DMR_6X2_PIC_V2.00.spi: controllalo qui per vedere la sua procedura).',
      'Reset della radio (obbligatorio dopo firmware o icone): spegni, tieni premuti <b>PTT</b> e <b>PF1</b> (il tasto sotto il PTT) e accendi; conferma l\'inizializzazione, non spegnere durante il riavvio, poi imposta data e ora.',
      'Con il CPS della nuova versione apri il codeplug salvato al punto 1 e riscrivilo nella radio.',
    ],
    icons: [
      'Con il CPS salva il codeplug in un file.',
      'Spegni la radio. Tieni premuti il <b>PTT</b> e il tasto <b>PF2</b> (il tasto laterale in basso) e accendi: il display mostra «UPDATE MODE».',
      'Collega la radio al PC con il cavo USB e avvia il programma ufficiale (pulsante qui sotto o dal CPS: Tool → Firmware Upgrade).',
      '«Open Update File»: scegli DMR_6X2_PIC_V2.00.spi; velocità <b>921600</b>, spunta <b>Duplex</b>, porta del «USB Virtual Com Port», poi <b>Write</b>.',
      'Al termine riavvia la radio e fai il reset (spegni, <b>PTT</b> + <b>PF1</b> all\'accensione, conferma), poi imposta data e ora e riscrivi il codeplug.',
    ],
    board: [
      'Aggiornamento della scheda APRS / baseband: BTECH lo indica come NON necessario, salvo che la radio non sia mai stata aggiornata. Segui le istruzioni del pacchetto e, in caso di dubbi, chiedi all\'assistenza BTECH.',
    ],
  },

  open(c) {
    const pro = /PRO/i.test(c.name);
    const m = openModal(`<div class="dialog big"><header>🧭 Aggiornamento guidato del firmware — ${esc(c.brand)} ${esc(c.name)}</header><div class="body">
      <p>Il firmware si scrive con il programma ufficiale BTECH (<b>Firmware_Update</b>, nella cartella Update del CPS installato). Qui si controllano i file
      del pacchetto e si segue la procedura delle istruzioni ufficiali; il protocollo di scrittura non è pubblicato e non viene usato da questo programma.</p>
      <h4 class="fwh">1. File dell'aggiornamento</h4>
      <div class="convrow"><button class="btn primary" data-check>📂 Controlla il file .spi…</button>
        <span class="note">nel pacchetto ${pro ? 'DMR-6X2-PRO-121e: cartella «DMR-6X2 PRO Firmware Update 1.21»' : 'DMR-6X2-221b: cartella «DMR_6X2 Firmware Update 2.21»'}</span></div>
      <div data-res></div>
      <h4 class="fwh">2. Procedura</h4>
      <ol class="fw6steps" data-steps></ol>
      <div class="convrow"><button class="btn" data-launch>▶ Avvia il programma ufficiale…</button>
        <label><input type="checkbox" data-admin> come amministratore</label></div>
      </div><footer><button class="btn" data-no>Chiudi</button></footer></div>`);
    const el = m.el;
    const showSteps = kind => { $('[data-steps]', el).innerHTML = this.steps[kind].map(s => '<li>' + s + '</li>').join(''); };
    showSteps('main');
    $('[data-no]', el).onclick = () => m.close();
    $('[data-check]', el).onclick = async () => {
      try {
        const r = await api('fw6x2', { action: 'check' }, { busyText: 'Controllo dei file…' });
        const wrong = r.kind === 'main' && (r.model === 'DMR-6X2 PRO') !== pro;
        const bad = (r.errors || []).length || wrong;
        $('[data-res]', el).innerHTML = `<div class="infores ${bad ? 'bad' : 'ok'}"><b>${esc(r.title)}</b>${r.date ? ' del ' + esc(r.date) : ''}<br>
          ${(r.checks || []).map(x => '✔ ' + esc(x)).join('<br>')}
          ${(r.errors || []).map(x => '<br>✖ ' + esc(x)).join('')}
          ${wrong ? `<br>✖ Questo file è per la <b>${esc(r.model)}</b>, non per la ${esc(c.name)}: non usarlo.` : ''}
          ${!bad ? '<br>Il file è adatto: nel programma ufficiale apri <b>' + esc(r.path.split(/[\\/]/).pop()) + '</b>.' : ''}</div>`;
        showSteps(r.kind);
      } catch (e) { reportError(e); }
    };
    $('[data-launch]', el).onclick = async () => {
      try {
        await api('fw6x2', { action: 'launch', admin: $('[data-admin]', el).checked }, { busyText: 'Avvio…' });
        toast('Programma ufficiale avviato: segui la procedura.');
      } catch (e) { reportError(e); }
    };
  },
};
