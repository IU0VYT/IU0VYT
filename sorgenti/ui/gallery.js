'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Galleria dei modelli Baofeng e Bajeton: ricerca, filtri, schede con le
// funzioni disponibili (canali via cavo, file .img, firmware) e i collegamenti
// ai programmi e ai manuali dei costruttori.

const TYPE_LABEL = { analog: 'Radioamatoriale', dmr: 'DMR', gmrs: 'GMRS / FRS', pmr: 'PMR446 / uso libero', kids: 'Per bambini' };
// divisioni per tipologia nella galleria e negli elenchi (in quest'ordine)
const TYPE_ORDER = ['analog', 'dmr', 'gmrs', 'pmr', 'kids'];
const TYPE_GROUP = {
  analog: 'Radioamatoriali (VHF/UHF)', dmr: 'DMR (digitali)', gmrs: 'GMRS / FRS (USA)', pmr: 'PMR446 e radio di uso libero', kids: 'Per bambini',
};
// ordine alfabetico «naturale» (UV-5R prima di UV-17)
const byModelName = (a, b) => a.name.localeCompare(b.name, 'it', { numeric: true, sensitivity: 'base' }) || a.brand.localeCompare(b.brand);
const FW_LABEL = { anc: 'firmware .anc', bf: 'firmware .BF', rh5: 'firmware .dat', kdhx: 'firmware .kdhx', bj9000: 'firmware .Firmware', kdh9g: 'firmware .kdh' };
const FW_KINDS = Object.keys(FW_LABEL);
const FW_DESC = {
  anc: 'file .anc del costruttore (protocollo del programma «BinTe ANC Bootloader»)',
  bf: 'file .BF (caricatore KDH, programma «BF Upgrade Tool for TFT 330L Series»; anche file .kdhx forniti dal costruttore)',
  rh5: 'file .dat «BF_5RH» (programma di aggiornamento della 5RH Pro)',
  kdhx: 'file .kdhx (programma «Baofeng Bootloader Pro» del costruttore, piattaforma H802)',
  bj9000: 'file .Firmware (programma «BJ9000_EnUPDATE» del costruttore)',
  kdh9g: 'file .kdh (programma «Bootloader_BaudRate38400» del costruttore)',
};

// disegno stilizzato della radio (per tipo e marca)
function radioIcon(item, size) {
  const s = size || 64;
  const body = item.Brand === 'Bajeton' ? '#e2711d' : '#1c64d8';
  const dark = '#243044';
  let ant, h = 34, w = 22, y = 22, scr = true, keys = true;
  switch (item.type) {
    case 'pmr': case 'kids': ant = '<rect x="27" y="10" width="4" height="14" rx="2" fill="' + dark + '"/>'; h = 30; y = 24; scr = item.type !== 'kids'; keys = false; break;
    case 'dmr': ant = '<rect x="26" y="6" width="5" height="18" rx="2.5" fill="' + dark + '"/>'; break;
    default: ant = '<rect x="27" y="3" width="3.5" height="21" rx="1.7" fill="' + dark + '"/>';
  }
  const kids = item.type === 'kids' ? `<circle cx="32" cy="${y + 20}" r="6" fill="#ffd54a" stroke="${dark}" stroke-width="1.5"/>` : '';
  const screen = scr ? `<rect x="${32 - w / 2 + 3}" y="${y + 4}" width="${w - 6}" height="9" rx="1.5" fill="#cfe6ff" stroke="${dark}" stroke-width="1"/>` : '';
  let kp = '';
  if (keys) for (let r = 0; r < 3; r++) for (let c = 0; c < 3; c++) kp += `<rect x="${32 - w / 2 + 4 + c * 5.2}" y="${y + 16 + r * 4.6}" width="3.6" height="3" rx="0.8" fill="#fff" opacity=".85"/>`;
  return `<svg viewBox="0 0 64 64" width="${s}" height="${s}" aria-hidden="true">
    ${ant}<rect x="${32 - w / 2 - 2}" y="${y + 6}" width="2.5" height="8" rx="1" fill="${dark}"/>
    <rect x="${32 - w / 2}" y="${y}" width="${w}" height="${h}" rx="4" fill="${body}" stroke="${dark}" stroke-width="1.5"/>
    ${screen}${kp}${kids}</svg>`;
}

const Gallery = {
  page: null,
  q: '', brand: '', type: '', onlyChan: false, onlyFw: false,

  build(page) {
    this.page = page;
    page.innerHTML = `<div class="galhead">
        <div><h2>Modelli Baofeng e Bajeton</h2>
          <div class="note">Scegli la tua radio per gestire canali e ponti (via cavo o con i file .img, compatibili con CHIRP) o aggiornarne il firmware.</div></div>
        <div class="galtools">
          <input type="search" data-q placeholder="Cerca modello (es. UV-5R, 888, BJ-7800)…" autocomplete="off">
          <div class="seg" data-brand><button data-v="" class="on">Tutte</button><button data-v="Baofeng">Baofeng</button><button data-v="Bajeton">Bajeton</button></div>
          <div class="seg" data-type><button data-v="" class="on">Tutti</button>${Object.entries(TYPE_LABEL).map(([k, v]) => `<button data-v="${k}">${v}</button>`).join('')}</div>
          <label><input type="checkbox" data-chan> con canali via cavo</label>
          <label><input type="checkbox" data-fw> con firmware</label>
        </div>
      </div>
      <div class="galbody" data-body></div>`;
    const q = $('[data-q]', page);
    q.addEventListener('input', debounce(() => { this.q = q.value; this.renderCards(); }, 120));
    for (const [sel, key] of [['[data-brand]', 'brand'], ['[data-type]', 'type']]) {
      const seg = $(sel, page);
      $$('button', seg).forEach(b => b.onclick = () => {
        $$('button', seg).forEach(x => x.classList.toggle('on', x === b));
        this[key] = b.dataset.v;
        this.renderCards();
      });
    }
    $('[data-chan]', page).onchange = e => { this.onlyChan = e.target.checked; this.renderCards(); };
    $('[data-fw]', page).onchange = e => { this.onlyFw = e.target.checked; this.renderCards(); };
    this.renderCards();
  },

  // filtro «con firmware» (anche dalla schermata iniziale)
  setOnlyFw(v) {
    this.onlyFw = !!v;
    const cb = this.page && $('[data-fw]', this.page);
    if (cb) cb.checked = this.onlyFw;
    this.renderCards();
  },

  render() { this.renderCards(); },

  matches(c) {
    if (this.brand && c.brand !== this.brand) return false;
    if (this.type && c.type !== this.type) return false;
    if (this.onlyChan && !(c.drivers || []).length) return false;
    if (this.onlyFw && !FW_KINDS.concat(['link']).includes(c.fw)) return false;
    const q = this.q.trim().toLowerCase().replace(/[\s\-_]/g, '');
    if (!q) return true;
    const hay = [c.brand, c.name].concat(c.aliases || []).join(' ').toLowerCase().replace(/[\s\-_]/g, '');
    return hay.includes(q);
  },

  card(c) {
    const it = { Brand: c.brand, type: c.type };
    const n = (c.drivers || []).length;
    const badges = [];
    if (n) badges.push('<span class="badge ch">Canali</span>');
    if (FW_KINDS.includes(c.fw)) badges.push(`<span class="badge fw" title="${esc(FW_LABEL[c.fw])}">Firmware</span>`);
    else if (c.fw === 'link') badges.push('<span class="badge ln" title="Aggiornamento con il programma del costruttore">Firmware (link)</span>');
    if (!n && (c.links || []).some(l => l.kind === 'cps')) badges.push('<span class="badge ln">Programma costruttore</span>');
    const open = n && c.drivers.some(d => Views[d] && Views[d].loaded);
    return `<button class="galcard ${n ? '' : 'nodrv'}" data-id="${esc(c.id)}">
      <div class="gicon">${radioIcon(it, 58)}</div>
      <div class="gtext"><div class="gname"><span class="gbrand">${esc(c.brand)}</span> ${esc(c.name)}${open ? ' <span class="gopen" title="Scheda aperta">●</span>' : ''}</div>
        ${c.aliases && c.aliases.length ? `<div class="galias">${esc(c.aliases.join(' · '))}</div>` : ''}
        <div class="gtype">${esc(TYPE_LABEL[c.type] || '')}</div>
        <div class="gbadges">${badges.join('')}</div></div></button>`;
  },

  renderCards() {
    if (!this.page) return;
    const body = $('[data-body]', this.page);
    const list = App.catalog.filter(c => this.matches(c));
    const withDrv = list.filter(c => (c.drivers || []).length);
    const without = list.filter(c => !(c.drivers || []).length);
    // in ogni sezione una divisione per tipologia, con i modelli in ordine alfabetico
    // dentro ogni tipologia: una parte per marca (Baofeng, Bajeton), in ordine alfabetico
    const brands = its => {
      const bs = [...new Set(its.map(c => c.brand))].sort((a, b) => (a === 'Baofeng' ? 0 : a === 'Bajeton' ? 1 : 2) - (b === 'Baofeng' ? 0 : b === 'Bajeton' ? 1 : 2) || a.localeCompare(b));
      return bs.map(br => {
        const bi = its.filter(c => c.brand === br);
        return `<div class="galbrand galbrand-${esc(br.toLowerCase())}"><div class="galbrand-head"><span class="galbrand-name">${esc(br)}</span>
          <span class="note">${bi.length} ${bi.length === 1 ? 'modello' : 'modelli'}</span></div>
          <div class="galgrid">${bi.map(c => this.card(c)).join('')}</div></div>`;
      }).join('');
    };
    const groups = items => TYPE_ORDER.concat([...new Set(items.map(c => c.type))].filter(t => !TYPE_ORDER.includes(t))).map(type => {
      const its = items.filter(c => c.type === type).sort(byModelName);
      return its.length ? `<div class="galarea galtype-${esc(type)}"><div class="galarea-head"><span class="galarea-name">${esc(TYPE_GROUP[type] || type)}</span>
        <span class="note">${its.length} ${its.length === 1 ? 'modello' : 'modelli'}</span></div>
        ${brands(its)}</div>` : '';
    }).join('');
    const sec = (title, note, items, cls) => `<section class="galsecbox ${cls}"><h3 class="galsec">${esc(title)}: ${items.length}</h3>
      ${note ? `<p class="note galnote">${note}</p>` : ''}${items.length ? groups(items) : '<p class="note galnote">Nessun modello con i filtri attuali.</p>'}</section>`;
    body.innerHTML = list.length ? sec('MODELLI COMPATIBILI', 'Canali e ponti via cavo o con i file .img; dove indicato, anche aggiornamento del firmware. Ogni modello ha la sua scheda, anche quando la gestione è la stessa di un altro.', withDrv, 'ok') +
      sec('MODELLI NON COMPATIBILI', 'Si programmano solo con il programma del costruttore (protocollo non documentato) o con programmi liberi: qui trovi i collegamenti ai programmi e ai manuali ufficiali e, dove esistono, ai programmi liberi alternativi.', without, 'no')
      : '<div class="welcome"><h2>Nessun modello trovato</h2><p>Prova a cambiare la ricerca o i filtri.</p></div>';
    $$('.galcard', body).forEach(b => b.onclick = () => this.detail(App.catalog.find(c => c.id === b.dataset.id)));
  },

  driverLine(kind) {
    if (kind === 'k67') return 'File .xlc del programma «5RH_PRO» · 640 canali · lettura e scrittura via cavo';
    if (kind === 'baj') return 'File .dat del programma «BT-11» · 15 zone da 64 canali · lettura e scrittura via cavo';
    const g = App.defs[kind];
    if (!g) return '';
    const bands = (g.bands || []).map(b => `${Math.round(b.lo / 1e5) / 10}–${Math.round(b.hi / 1e5) / 10}`).join(', ');
    return `${g.slots} canali · nomi ${g.nameMax ? 'fino a ' + g.nameMax + ' caratteri' : 'non memorizzati'} · ${(g.powerLevels || []).slice().sort((a, b) => a.sem - b.sem).map(p => p.label).join(', ')}${bands ? ' · ' + bands + ' MHz' : ''}`;
  },

  detail(c) {
    if (!c) return;
    const drivers = c.drivers || [];
    const linkIcon = { cps: '💾', fw: '⬆', manual: '📖', info: 'ℹ' };
    const m = openModal(`<div class="dialog mid galdlg"><header>${esc(c.brand)} ${esc(c.name)}</header><div class="body">
      <div class="galdtop"><div class="gicon">${radioIcon({ Brand: c.brand, type: c.type }, 92)}</div><div>
        <div class="gtype">${esc(TYPE_LABEL[c.type] || '')}</div>
        ${c.aliases && c.aliases.length ? `<div class="note">Anche: ${esc(c.aliases.join(', '))}</div>` : ''}
        ${c.notes ? `<p>${esc(c.notes)}</p>` : ''}
        ${!drivers.length ? `<p class="note">${esc(App.noDriver)}</p>` : ''}</div></div>
      ${drivers.length ? `<h4 class="fwh">Canali e ponti</h4>
        ${drivers.length > 1 ? '<p class="note">Questa radio esiste in più versioni: scegli quella giusta. Se non è quella, la radio non risponde alla lettura e non viene modificata.</p>' : ''}
        <div class="galacts">${drivers.map(d => `<div class="galact"><button class="btn primary" data-open="${esc(d)}">📻 ${esc(d === 'k67' || d === 'baj' ? 'Vai alla scheda ' + RADIO_DEF[d].name : 'Apri la scheda ' + (App.defs[d] ? App.defs[d].name : d))}</button>
          <span class="note">${esc(this.driverLine(d))}</span></div>`).join('')}</div>` : ''}
      ${c.guide === '6x2' ? `<h4 class="fwh">Firmware</h4><div class="galact"><button class="btn" data-guide>🧭 Aggiornamento guidato…</button>
        <span class="note">controllo dei file del pacchetto ufficiale, procedura passo per passo e avvio del programma BTECH</span></div>` : ''}
      ${FW_KINDS.includes(c.fw) ? `<h4 class="fwh">Firmware</h4><div class="galact"><button class="btn" data-fw>⬆ Aggiorna il firmware…</button>
        <span class="note">${esc(FW_DESC[c.fw])}</span></div>` : ''}
      ${(c.links || []).length ? `<h4 class="fwh">Programmi e documenti</h4><ul class="gallinks">${c.links.map(l => `<li>${linkIcon[l.kind] || '🔗'} <a href="#" data-link="${esc(l.url)}">${esc(l.label)}</a> <span class="note">${esc(l.url.replace(/^https?:\/\//, '').split('/')[0])}</span></li>`).join('')}</ul>` : ''}
      </div><footer><button class="btn" data-no>Chiudi</button></footer></div>`);
    const el = m.el;
    $('[data-no]', el).onclick = () => m.close();
    $$('[data-link]', el).forEach(a => a.onclick = (e) => { e.preventDefault(); Info.open(a.dataset.link); });
    $$('[data-open]', el).forEach(b => b.onclick = () => {
      m.close();
      const k = b.dataset.open;
      try {
        if (Views[k]) App.showTab(k); else App.openRadio(k, true);
      } catch (e) { reportError(e); }
    });
    const gb = $('[data-guide]', el);
    if (gb) gb.onclick = () => { m.close(); Fw6x2.open(c); };
    const fb = $('[data-fw]', el);
    if (fb) fb.onclick = () => { m.close(); Firmware.open(c.fw, drivers[0]); };
  },
};
