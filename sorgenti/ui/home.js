'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Scheda «Home»: schermata iniziale mostrata a ogni avvio (titolo a neon in stile
// sala giochi anni '80, funzioni principali, versione, copyright e licenza).

// indirizzi mostrati anche se l'API info non risponde (uguali a appSite e appRepo di info.go)
const SITE_URL = 'https://iu0vyt.github.io/IU0VYT/';
const REPO_URL = 'https://github.com/IU0VYT/IU0VYT';

// riquadro delle marche (4.1) in stile cabinato da sala giochi anni '80: cornice a
// pixel, titolo «PLAYER SELECT» con due invasori, marche in 4 colonne
function pixelInvader(cls) {
  // invasore 11 x 8 pixel
  const rows = ['00100000100', '00010001000', '00111111100', '01101110110', '11111111111', '10111111101', '10100000101', '00011011000'];
  let r = '';
  rows.forEach((row, y) => [...row].forEach((c, x) => { if (c === '1') r += `<rect x="${x}" y="${y}" width="1" height="1"/>`; }));
  return `<svg class="${cls}" viewBox="0 0 11 8" width="22" height="16" aria-hidden="true" shape-rendering="crispEdges">${r}</svg>`;
}
function brandMarquee(brands, where) {
  return `<div class="hb hb-${where}" role="group" aria-label="Marche di radio gestite: ${brands.length}">
    <div class="hb-frame"><div class="hb-in">
      <div class="hb-title">${pixelInvader('hb-inv')}<span>PLAYER SELECT · ${brands.length} MARCHE</span>${pixelInvader('hb-inv hb-inv2')}</div>
      <ul class="hb-grid">${brands.map((b, i) => `<li style="--i:${i}"><span class="hb-dot" style="background:${brandColor(b)}"></span>${esc(b)}</li>`).join('')}</ul>
    </div></div></div>`;
}

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
    const nBrands = catalogBrands().length;
    // zero barrato, come nei nominativi radioamatoriali
    const letters = (s, cls) => [...s].map((ch, i) => `<span class="${cls}" style="--i:${i}">${ch === '0' ? 'Ø' : esc(ch)}</span>`).join('');
    page.innerHTML = `<div class="home">
      <div class="home-sky" aria-hidden="true"><div class="home-sun"></div><div class="home-grid"></div></div>
      <div class="home-inner">
        <h1 class="neon" aria-label="IU0VYT Flash">
          <span class="neon-top">${letters('IU0VYT', 'nl')}</span>
          <span class="neon-bottom">${letters('FLASH', 'nl')}</span>
        </h1>
        <div class="home-tag">GESTIONE CANALI - FREQUENZE - PONTI RADIO - FIRMWARE</div>
        ${brandMarquee(catalogBrands(), 'home')}
        <button class="home-start" data-go="gal">▶ PRESS START</button>
        <div class="home-cards">
          <button class="home-card" data-go="gal"><b>📻 ${nChan} radio</b><span>Canali via cavo e file CHIRP. Galleria di ${nAll} modelli, ${nBrands} marche.</span></button>
          <button class="home-card" data-go="ident"><b>🔎 Identifica</b><span>Riconosce da sé la radio collegata, anche più radio insieme.</span></button>
          <button class="home-card" data-go="arch"><b>📡 Ponti radio</b><span>Ponti italiani (IK2ANE) da inserire nelle radio.</span></button>
          <button class="home-card" data-go="gal-fw"><b>⬆ Firmware</b><span>Aggiornamento di ${nFw} modelli, con controllo del file.</span></button>
          <button class="home-card" data-go="info"><b>🔄 Aggiornamenti</b><span>Nuove versioni controllate a ogni avvio.</span></button>
        </div>
        <div class="home-meta">
          <div>Versione <b>${esc(d.version || '')}</b>${d.buildDate && d.buildDate !== 'non indicata' ? ` · compilata il ${esc(d.buildDate)}` : ''} · Copyright © 2026 ${esc(d.author)}</div>
          <div>Software libero: ${esc(d.license || '')} — senza alcuna garanzia</div>
          <div class="home-links">
            <span>🌐 Sito web: <a href="#" data-link="${esc(d.site || SITE_URL)}">${esc(d.site || SITE_URL)}</a></span>
            <span>💻 Sorgenti: <a href="#" data-link="${esc(d.repo || REPO_URL)}">${esc(d.repo || REPO_URL)}</a></span>
          </div>
          <div class="home-small">Programma indipendente, non affiliato ai costruttori delle radio. I marchi appartengono ai rispettivi titolari.</div>
        </div>
      </div>
    </div>`;
    $$('[data-link]', page).forEach(x => x.onclick = (e) => { e.preventDefault(); Info.open(x.dataset.link); });
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
