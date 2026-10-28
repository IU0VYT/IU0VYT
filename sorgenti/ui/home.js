'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda «Home»: schermata iniziale mostrata a ogni avvio (titolo a neon in stile
// sala giochi anni '80, funzioni principali, versione, copyright e licenza).

const Home = {
  built: false,

  async render(page) {
    if (this.built) return;
    this.built = true;
    let d = { name: 'IU0VYT Flash', version: '', buildDate: '', author: 'Daniele "IU0VYT" Bonifazi', license: 'GNU GPL v3 o successiva' };
    try { d = await api('info', {}, { quiet: true }); } catch (e) { /* si mostra comunque */ }
    const nChan = App.catalog.filter(c => (c.drivers || []).length).length;
    const nFw = App.catalog.filter(c => FW_KINDS.includes(c.fw)).length;
    const nAll = App.catalog.length;
    // zero barrato, come nei nominativi radioamatoriali
    const letters = (s, cls) => [...s].map((ch, i) => `<span class="${cls}" style="--i:${i}">${ch === '0' ? 'Ø' : esc(ch)}</span>`).join('');
    page.innerHTML = `<div class="home">
      <div class="home-sky" aria-hidden="true"><div class="home-sun"></div><div class="home-grid"></div></div>
      <div class="home-inner">
        <h1 class="neon" aria-label="IU0VYT Flash">
          <span class="neon-top">${letters('IU0VYT', 'nl')}</span>
          <span class="neon-bottom">${letters('FLASH', 'nl')}</span>
        </h1>
        <div class="home-tag">CANALI · PONTI · FIRMWARE — BAOFENG &amp; BAJETON</div>
        <button class="home-start" data-go="gal">▶ PRESS START</button>
        <div class="home-cards">
          <button class="home-card" data-go="gal"><b>📻 ${nChan} radio</b><span>Canali e frequenze letti e scritti via cavo, file compatibili con CHIRP. Galleria di ${nAll} modelli Baofeng e Bajeton con programmi e manuali ufficiali.</span></button>
          <button class="home-card" data-go="ident"><b>🔎 Identifica</b><span>Collega una radio: il programma la riconosce da solo e mostra marca, modello, firmware e le altre informazioni che comunica.</span></button>
          <button class="home-card" data-go="arch"><b>📡 Ponti radio</b><span>Archivio dei ponti italiani (elenco gratuito di IK2ANE), filtri, nomi automatici e inserimento nelle radio.</span></button>
          <button class="home-card" data-go="gal-fw"><b>⬆ Firmware</b><span>Aggiornamento del firmware di ${nFw} modelli con i protocolli dei programmi ufficiali, con controllo del file e copia di sicurezza.</span></button>
          <button class="home-card" data-go="info"><b>🔄 Aggiornamenti</b><span>Il programma controlla da solo a ogni avvio se esiste una versione più recente.</span></button>
        </div>
        <div class="home-meta">
          <div>Versione <b>${esc(d.version || '')}</b>${d.buildDate && d.buildDate !== 'non indicata' ? ` · compilata il ${esc(d.buildDate)}` : ''}</div>
          <div>Copyright © 2026 ${esc(d.author)}</div>
          <div>Software libero: ${esc(d.license || '')} — senza alcuna garanzia</div>
          <div class="home-small">Programma indipendente, non affiliato a Baofeng o Bajeton. I marchi appartengono ai rispettivi titolari.</div>
        </div>
      </div>
    </div>`;
    $$('[data-go]', page).forEach(b => b.onclick = () => {
      const g = b.dataset.go;
      if (g === 'gal-fw') {
        App.showTab('gal');
        Gallery.setOnlyFw(true);
        return;
      }
      App.showTab(g);
    });
  },
};
