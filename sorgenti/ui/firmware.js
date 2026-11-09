'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Aggiornamento del firmware (Bajeton BJ-8300/7800/7700 .anc e BJ-9000 .Firmware, Baofeng UV-5RM/5RH .BF,
// 5RH Pro .dat, UV-25 Plus .kdhx, UV-9G .kdh): scelta del firmware (file del PC
// o Internet, anche dentro archivi .zip/.rar), identificazione della radio,
// conferme e avanzamento.

// 4.12: testo semplice di un pezzo di HTML del programma (per i messaggi di conferma),
// letto dal browser invece che togliendo i tag con un'espressione regolare
function htmlText(h) {
  const t = document.createElement('template');
  t.innerHTML = h;
  return t.content.textContent;
}

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
  // expect: tipo di firmware della radio da cui si apre la finestra ('anc' | 'bf' | 'rh5' | 'kdhx' | 'bj9000' | 'kdh9g' | 'k5' | 'tdh' | 'th8600' | 'btu' | 'rt890' | 'dm1701' | 'sgl'), radioKind: scheda d'origine
  open(expect, radioKind) {
    const TARGET = { anc: 'Bajeton BJ-8300 / BJ-7800 / BJ-7700', bf: 'Baofeng UV-5RM / UV-5RH e radio con lo stesso caricatore', rh5: 'Baofeng 5RH Pro / K67 Pro',
      kdhx: 'Baofeng UV-25 Plus, Radtel RT-470 / RT-495 / RT-630 e radio con il caricatore «Bootloader Pro»', bj9000: 'Bajeton BJ-9000', kdh9g: 'Baofeng UV-9G',
      k5: 'Quansheng UV-K5 / UV-K5(8) / UV-K6 / UV-5R Plus / UV-K5 V3, Retevis RA79', th8600: 'Retevis RA89 / RT85', tdh: 'TIDRADIO TD-H3 / TD-H8 o Radtel RT-730', rt890: 'Radtel RT-890', dm1701: 'Baofeng DM-1701 / Retevis RT84', sgl: 'Baofeng RD-5R / DM-1801',
      btu: 'Radtel RT-910 / RT-900 e Bajeton BJ-7700 senza Bluetooth' };
    const PROTO = {
      anc: 'quello del programma originale «BinTe ANC Bootloader» V1.1',
      bf: 'quello del programma originale «BF Upgrade Tool for TFT 330L Series» (KDH Bootloader V1.0)',
      rh5: 'quello del programma originale di aggiornamento della 5RH Pro (Radio.exe)',
      kdhx: 'quello del programma originale «Baofeng Bootloader Pro» (per le Radtel «JJCC Bootloader Pro», stesso programma con la piattaforma JJCC)',
      bj9000: 'quello del programma originale «BJ9000_EnUPDATE» (BJ9000_UPDATE_V02)',
      kdh9g: 'quello del programma originale «Bootloader_BaudRate38400» della UV-9G',
      btu: 'quello del programma originale «Radtel Bootload Tools» (caricatore KDH V1.0, presentazione «PROGRAMBT80U»)',
      k5: 'quello del caricatore della UV-K5 documentato dai programmi liberi k5prog e uvtools, riscritto da zero',
      th8600: 'quello del programma originale «RA89-update-software» di Retevis («TH8600-UpDAta Soft» 2.3)',
      th8600rt85: 'quello del programma originale «RT85 10W_5W_UpData» di Retevis («TH8600-UpDAta Soft» 2.1)',
      tdh: 'quello dei programmi originali «H8_IAP» di TIDRADIO e «RT730 Update tool» di Radtel (documentato da nicsure, TD-H3-Engineering), riscritto da zero',
      rt890: 'quello del caricatore della RT-890 documentato dal programma libero «rt_890_flasher» di JuantAldea, riscritto da zero',
      dm1701: 'quello del caricatore DFU della DM-1701 usato dal programma libero «radio_tool» (provato da quel progetto su radio vere), riscritto da zero',
      sgl: 'quello del caricatore della famiglia GD-77 usato dal «FirmwareLoader» di OpenGD77 e da radio_tool, riscritto da zero',
    };
    // caricatori KDH recenti: istruzioni del costruttore e velocità
    const BOOT = {
      kdhx: { manual: 'radio spenta: tieni premuti <b>SK2 e SK3</b> (i due tasti sotto il PTT; sulle Radtel RT-470 / RT-495 / RT-630 i <b>tasti laterali 2 e 3</b>) e accendila; l\'indicatore diventa verde. È la procedura indicata da Baofeng per la UV-25 Plus e da Radtel (per le altre radio segui le istruzioni fornite con il firmware)',
        auto: 'la radio accesa viene portata nel caricatore dal programma (pulsante «自动升级»); prima 9600 baud, poi 115200', speed: 'Velocità come nel programma originale: 9600 baud per la richiesta, 115200 per il firmware.' },
      bj9000: { manual: 'radio spenta: tieni premuti i <b>tasti laterali 3 e 4</b> e accendila; compare la schermata «Update» (pulsante «Flashing Mode»)',
        auto: 'la radio accesa viene portata nel caricatore dal programma (pulsante «Upgrade Mode»)', speed: 'Velocità fissa 115200 baud, come nel programma originale. La radio controlla che il firmware sia per il suo modello prima di cancellare la memoria.' },
      kdh9g: { manual: 'radio spenta: tieni premuti i <b>tasti 12 e 13</b> e accendila; si accende la torcia. È la procedura indicata da Baofeng',
        auto: 'la radio accesa viene portata nel caricatore dal programma (prima 9600 baud, poi 38400)', speed: 'Velocità come nel programma originale: 38400 baud per il firmware.' },
      btu: { manual: 'radio spenta: tieni premuti i <b>due tasti laterali</b> (sotto il PTT) e accendila: il display mostra «UPDATE» (pulsante «Update» con la scritta nera del programma originale). ' +
          'Secondo Radtel, se la radio non riparte dopo un aggiornamento: ripeti così; se non basta togli la batteria per 10 minuti e riprova',
        auto: 'la radio accesa viene portata nel caricatore dal programma (pulsante «Update» con la scritta rossa); prima 57600 baud, poi 115200',
        speed: 'Velocità come nel programma originale: 57600 baud per la richiesta, 115200 per il firmware. Il caricatore non controlla il modello: serve la conferma scritta.' },
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
          (es. BF_5RH_501_v2_0_9.dat), <b>.kdhx</b> della UV-25 Plus (UV25Pro_NRF_401+_V0.23_250522.kdhx) e delle Radtel RT-470 / RT-495 / RT-630, <b>.kdh</b> della UV-9G (T52GMRS_M8_V02_231004.kdh)
          <b>.bin</b> delle Quansheng UV-K5, UV-K5(8), UV-K6, UV-5R Plus e UV-K5 V3 e della Retevis RA79 (es. k5_v2.01.26_publish.bin, RA79 V3.bin),
          <b>.bin</b> della Retevis RA89 (Eng_RA89_Upadat_20250524_B91.38.bin), <b>.hex</b> della Retevis RT85 (RT85-Firmware-V1.53.hex)
          <b>.BIN</b> di TIDRADIO TD-H3 / TD-H8 e Radtel RT-730 (es. 730Step improvement.BIN)
          e <b>.BTU</b> delle Radtel RT-910 e RT-900 e della Bajeton BJ-7700 senza Bluetooth (es. RT910_H1.00G_V0.11_260305.BTU).
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

        <div data-panel="k5" style="display:none">
          <div class="fwbox"><b>Porta la radio nel caricatore</b> (procedura di Quansheng):
            <ol><li>Spegni la radio e collega il cavo di programmazione.</li>
            <li>Riaccendila <b>tenendo premuto il PTT</b>: la torcia si accende e lo schermo resta spento.</li>
            <li>Premi «Avvia aggiornamento» (il programma aspetta la radio per 30 secondi).</li></ol>
            Prima di scrivere il programma controlla che il caricatore sia del tipo del firmware (2 = UV-K5, 3 = UV-K5(8)/UV-K6, 4 = UV-5R Plus,
            7 = UV-K5 V3 / UV-K1 / Retevis RA79 V3, con la stessa versione principale del caricatore, es. 7.00):
            altrimenti la radio non viene toccata. Canali, impostazioni e taratura non vengono cancellati.</div>
          <p class="note">Velocità fissa 38400 baud. Caricatore 7.x (4.4): protocollo come pubblicato da quansheng_toolkit e Multi-UVTools,
            riscritto da zero e provato con una radio simulata.</p>
        </div>
        <div data-panel="th8600" style="display:none">
          <div class="fwbox" data-th="ra89"><b>Porta la radio in modalità di aggiornamento</b> come indicato da Retevis (radio spenta, accesa tenendo premuti i tasti
            indicati: l'indicatore diventa verde), collega il cavo e premi «Avvia aggiornamento».
            Il file è cifrato e la radio non controlla il modello: per questo si deve scrivere il modello.
            <b>Se l'invio si interrompe non accendere la radio normalmente</b> (Retevis avverte che si blocca): lasciala in modalità di aggiornamento
            e ripeti subito l'aggiornamento.</div>
          <div class="fwbox" data-th="rt85" style="display:none"><b>Porta la radio in modalità di aggiornamento</b> (procedura di Retevis):
            <ol><li>Radio spenta: tieni premuti <b>PTT e P1</b> e accendila; l'indicatore diventa <b>verde</b>.</li>
            <li>Collega il cavo, scegli la potenza della tua RT85 e premi «Avvia aggiornamento».</li></ol>
            Il file è cifrato e la radio non controlla il modello: per questo si deve scrivere il modello.
            <b>Se l'invio si interrompe non accendere la radio</b> (Retevis avverte che si blocca per sempre): lasciala in modalità di aggiornamento
            e ripeti subito l'aggiornamento.</div>
          <div class="row" data-th="rt85" style="display:none"><label class="l">Potenza della RT85</label>
            <select data-thpower><option value="">— scegli —</option><option value="5W">5 W</option><option value="10W">10 W</option></select>
            <span class="note">come nel programma di Retevis («5W» / «10W»): la radio risponde solo alla scelta giusta, altrimenti non viene toccata.
              La RT85 descritta sul sito di Retevis è da 5 W (come la TYT TH-UV88).</span></div>
          <p class="note" data-thnote>Velocità 57600 baud, blocchi da 128 byte confermati dalla radio uno per uno, come nel programma di Retevis
            (confrontato byte per byte con il suo codice).</p>
        </div>
        <div data-panel="tdh" style="display:none">
          <div class="fwbox"><b>Porta la radio nel caricatore</b> (procedura di TIDRADIO e Radtel):
            <ol><li>Collega il cavo di programmazione alla radio spenta.</li>
            <li>Premi «Avvia aggiornamento»: il programma aspetta la radio per 60 secondi.</li>
            <li>Accendi la radio <b>tenendo premuto il PTT</b>: l'invio comincia da solo.</li></ol>
            Il caricatore non controlla il modello del firmware (TD-H3, TD-H8 e RT-730 hanno file diversi): per questo si deve scrivere il modello.
            Se l'invio si interrompe la radio resta nel caricatore: riaccendila tenendo premuto il PTT e ripeti l'aggiornamento.</div>
          <p class="note">Velocità fissa 115200 baud, blocchi da 32 byte confermati dalla radio uno per uno.</p>
        </div>

        <div data-panel="rt890" style="display:none">
          <div class="fwbox"><b>Porta la radio nel caricatore</b>:
            <ol><li>Collega il cavo di programmazione alla radio spenta.</li>
            <li>Accendi la radio <b>tenendo premuti i due tasti laterali</b> (lo schermo resta spento).</li>
            <li>Premi «Avvia aggiornamento» (il programma aspetta la radio per 60 secondi).</li></ol>
            Il caricatore non controlla il modello del firmware: per questo si deve scrivere il modello. Le schede PCB 2.0 e PCB 2.1 (RT-890 II)
            usano firmware diversi. Se l'invio si interrompe la radio resta nel caricatore: riaccendila tenendo premuti i due tasti laterali e ripeti l'aggiornamento.
            Dopo un firmware libero (M7OCM) i canali si gestiscono con la scheda «Radtel RT-890 (firmware M7OCM)».</div>
          <p class="note">Velocità fissa 115200 baud: cancellazione della memoria del programma, poi blocchi da 128 byte confermati dalla radio uno per uno.</p>
        </div>

        <div data-panel="dm1701" style="display:none">
          <div class="fwbox"><b>Porta la radio nel caricatore</b>:
            <ol><li>Radio spenta: tieni premuti <b>il PTT e il tasto sopra il PTT</b> e accendila (l'indicatore lampeggia, lo schermo resta spento).</li>
            <li>Collega il cavo USB di programmazione (serve il driver USB della DM-1701, come per i canali).</li>
            <li>Premi «Avvia aggiornamento».</li></ol>
            Si scrive solo la memoria del programma: il caricatore della radio resta intatto, quindi se l'invio si interrompe basta riaccendere la radio
            nel caricatore e ripetere l'aggiornamento. Le parti del file per la memoria SPI (caratteri, immagini) non si scrivono.</div>
          <p class="note">Collegamento USB DFU (STMicroelectronics 0483:DF11): settori cancellati prima di scrivere, blocchi da 1024 byte.</p>
        </div>

        <div data-panel="sgl" style="display:none">
          <div class="fwbox"><b>Porta la radio nella modalità di aggiornamento</b> (lo schermo resta spento, l'indicatore si accende o lampeggia):
            <ul><li><b>RD-5R</b>: radio spenta, tieni premuti il tasto <b>arancione S1</b> e il tasto <b>nero Fn (S2)</b>, ai due lati del PTT, e accendila.</li>
            <li><b>DM-1801</b>: radio spenta, tieni premuti <b>i due tasti piccoli sotto il PTT</b> e accendila.</li></ul>
            Poi collega il cavo USB (nessun driver da installare: è un dispositivo HID) e premi «Avvia aggiornamento».
            La radio controlla la chiave del modello prima di cancellare la memoria. Se l'invio si interrompe la radio resta nel caricatore: ripeti l'aggiornamento.</div>
          <p class="note">Collegamento USB HID (15A2:0073): blocchi da 32 byte con la somma di controllo a ogni KB, confermati dalla radio.</p>
        </div>

        <h4 class="fwh"><span class="num">3</span> Conferma</h4>
        <div data-warnok style="display:none" class="row"><label><input type="checkbox" data-acceptwarn> <b>Ho letto le segnalazioni sul firmware</b> e voglio usarlo comunque</label></div>
        <div class="row"><label><input type="checkbox" data-confirm> <b data-conftext>Confermo il modello della radio collegata.</b>
          <span class="note">Un firmware per un'altra radio può renderla inutilizzabile.</span></label></div>
        <div class="row" data-skiprow><label><input type="checkbox" data-skip> salta l'identificazione automatica <span class="note">(solo se la radio funziona ma non risponde alla lettura via cavo)</span></label></div>
        <div class="row" data-typerow style="display:none"><label class="l" data-typelabel>Scrivi BJ8300</label><input type="text" data-typed maxlength="10" style="width:120px" autocomplete="off">
          <span class="note">conferma richiesta quando il programma non può verificare il modello della radio</span></div>
        <p class="fwwarn">⚠ Durante l'aggiornamento <b>non spegnere la radio e non scollegare il cavo</b>; usa una batteria carica. Il produttore non documenta la procedura:
          il protocollo è <span data-proto>quello del programma originale del costruttore</span><span data-protohow>, ricavato dal suo codice</span> e verificato con radio simulate.</p>
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
      kind() === 'kdh9g' || kind() === 'btu' || kind() === 'tdh' || kind() === 'rt890' || kind() === 'dm1701' || kind() === 'sgl' || kind() === 'th8600' || (kind() === 'kdhx' && btmode() === 'manual') || !!(fw && fw.typed) || otherModel();
    // radio scelta nella finestra diversa da quelle a cui è destinato il file
    const otherModel = () => {
      if (!fw || !fw.models || !fw.models.length) return false;
      const x = $(kind() === 'bf' ? '[data-bfmodel]' : '[data-btmodel]', el);
      return !!(x && x.value && !fw.models.includes(x.value));
    };
    const refresh = () => {
      const k = kind();
      // 4.10: DM-1701, RD-5R e DM-1801 via USB (nessuna porta seriale)
      const usb = k === 'dm1701' ? 'DFU:0483:DF11' : k === 'sgl' ? 'HID:15A2:0073' : '';
      if (usb && sel.value !== usb) {
        if (sel.dataset.prev == null) sel.dataset.prev = sel.innerHTML;
        sel.innerHTML = `<option value="${usb}">${k === 'sgl' ? 'USB (HID) — RD-5R / DM-1801' : 'USB (DFU) — DM-1701'}</option>`;
      } else if (!usb && sel.dataset.prev != null) {
        sel.innerHTML = sel.dataset.prev;
        delete sel.dataset.prev;
      }
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
      const rt85 = k === 'th8600' && !!(fw && fw.power);
      $$('[data-th]', el).forEach(x => x.style.display = (x.dataset.th === 'rt85') === rt85 ? '' : 'none');
      $('[data-thnote]', el).textContent = 'Velocità 57600 baud, blocchi da ' + (rt85 ? '64 byte (4 righe del file .hex con la loro somma)' : '128 byte') +
        ' confermati dalla radio uno per uno, come nel programma di Retevis (confrontato byte per byte con il suo codice).';
      $('[data-proto]', el).textContent = (rt85 ? PROTO.th8600rt85 : PROTO[k]) || 'quello del programma originale del costruttore';
      $('[data-protohow]', el).textContent = k === 'k5' || k === 'tdh' || k === 'rt890' || k === 'dm1701' || k === 'sgl' ? '' : ', ricavato dal suo codice';
      const target = fw ? fw.target : TARGET[k];
      $('[data-conftext]', el).textContent = k === 'kdh9g' || k === 'btu'
        ? `Confermo che la radio collegata è una ${target}${fw && fw.hw ? ' (' + fw.hw + ')' : ''}: il suo caricatore non controlla il firmware che riceve.`
        : k === 'bf'
        ? `Confermo che la radio collegata è una ${target}${fw && fw.platform ? ' con chip radio ' + fw.platform : ''} e che questo firmware è quello adatto.`
        : k === 'k5'
        ? `Confermo che la radio collegata è una ${target} ed è nel caricatore (accesa tenendo premuto il PTT).`
        : k === 'tdh'
        ? `Confermo che la radio collegata è una ${target}: dopo l'avvio la accenderò tenendo premuto il PTT.`
        : k === 'rt890'
        ? `Confermo che la radio collegata è una ${target} ed è nel caricatore (accesa tenendo premuti i due tasti laterali).`
        : k === 'dm1701'
        ? `Confermo che la radio collegata via USB è una ${target} ed è nel caricatore (accesa tenendo premuti il PTT e il tasto sopra il PTT).`
        : k === 'sgl'
        ? `Confermo che la radio collegata via USB è una ${target} ed è nella modalità di aggiornamento (schermo spento).`
        : k === 'th8600'
        ? `Confermo che la radio collegata è una ${target}${rt85 && $('[data-thpower]', el).value ? ' da ' + $('[data-thpower]', el).value.replace('W', ' W') : ''} ed è in modalità di aggiornamento.`
        : target ? `Confermo che la radio collegata è una ${target}.` : 'Confermo il modello della radio collegata.';
      const want = (fw && fw.confirm) || (k === 'anc' || !k ? 'BJ8300' : '');
      $('[data-typelabel]', el).textContent = want ? 'Scrivi ' + want : 'Scrivi il modello indicato dal file';
      const typedOk = !needTyped() || $('[data-typed]', el).value.toUpperCase().replace(/[\s\-_.]/g, '') === want;
      const ok = fw && fw.valid && sel.value && $('[data-confirm]', el).checked && typedOk && (!fw.warnings || $('[data-acceptwarn]', el).checked) &&
        (!rt85 || !!$('[data-thpower]', el).value);
      $('[data-ok]', el).disabled = !ok || running;
      $('[data-ident]', el).disabled = !sel.value || running || k === 'dm1701' || k === 'sgl';
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
      $('[data-thpower]', el).value = '';
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
      sel.innerHTML = r.ports.map(p => `<option value="${esc(p.name)}">${esc(portLabel(p))}</option>`).join('');
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
          : 'La radio deve essere già nel caricatore (' + htmlText(BOOT[k].manual) + ').\n';
      } else if (k === 'k5') {
        req = Object.assign(base, { mode: 'manual' });
        how = 'La radio deve essere nel caricatore (spenta e riaccesa tenendo premuto il PTT: torcia accesa, schermo spento).\n';
      } else if (k === 'th8600') {
        req = Object.assign(base, { mode: 'manual', confirmText: $('[data-typed]', el).value, power: fw.power ? $('[data-thpower]', el).value : '' });
        how = (fw.power ? `Potenza scelta: ${req.power.replace('W', ' W')}. La radio deve essere già in modalità di aggiornamento (accesa tenendo premuti PTT e P1: indicatore verde).`
          : 'La radio deve essere già in modalità di aggiornamento (indicatore verde).') + ' Se l\'invio si interrompe non accenderla normalmente: ripeti subito l\'aggiornamento.\n';
      } else if (k === 'sgl') {
        req = Object.assign(base, { mode: 'manual', confirmText: $('[data-typed]', el).value });
        how = 'La radio deve essere collegata via USB e accesa nella modalità di aggiornamento (schermo spento).\n';
      } else if (k === 'dm1701') {
        req = Object.assign(base, { mode: 'manual', confirmText: $('[data-typed]', el).value });
        how = 'La radio deve essere collegata via USB e accesa nel caricatore (tenendo premuti il PTT e il tasto sopra il PTT).\n';
      } else if (k === 'rt890') {
        req = Object.assign(base, { mode: 'manual', confirmText: $('[data-typed]', el).value });
        how = 'La radio deve essere nel caricatore (spenta e riaccesa tenendo premuti i due tasti laterali, schermo spento).\n';
      } else if (k === 'tdh') {
        req = Object.assign(base, { mode: 'manual', confirmText: $('[data-typed]', el).value });
        how = 'Dopo l\'avvio spegni la radio e riaccendila tenendo premuto il PTT (il programma la aspetta per 60 secondi).\n';
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
