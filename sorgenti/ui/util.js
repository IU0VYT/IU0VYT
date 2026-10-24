'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Utilità comuni: chiamate al programma, formati, finestre di dialogo.

const TOKEN = new URLSearchParams(location.search).get('t') || '';
const App = { init: null, nativeDlg: false };

function esc(s) {
  return String(s == null ? '' : s).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}
function $(sel, root) { return (root || document).querySelector(sel); }
function $$(sel, root) { return Array.from((root || document).querySelectorAll(sel)); }
function h(html) { const t = document.createElement('template'); t.innerHTML = html.trim(); return t.content.firstElementChild; }
function clone(o) { return o == null ? o : JSON.parse(JSON.stringify(o)); }

// ------------------------------------------------------------ messaggi
let toastTimer = null;
function toast(msg, isErr) {
  const t = $('#toast');
  t.textContent = msg;
  t.className = 'show' + (isErr ? ' err' : '');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.className = ''; }, isErr ? 6000 : 3200);
}
let busyCount = 0, busyTimer = null;
function busy(on, text) {
  busyCount += on ? 1 : -1;
  if (busyCount < 0) busyCount = 0;
  $('#busy-text').textContent = text || 'Attendere…';
  clearTimeout(busyTimer);
  if (busyCount > 0) busyTimer = setTimeout(() => $('#busy').classList.add('show'), 250);
  else $('#busy').classList.remove('show');
}

// ------------------------------------------------------------ chiamate al programma
class Cancelled extends Error { constructor() { super('annullato'); this.cancel = true; } }

async function api(name, body, opts) {
  opts = opts || {};
  if (!opts.quiet) busy(true, opts.busyText);
  try {
    for (let attempt = 0; attempt < 2; attempt++) {
      const r = await fetch('/api/' + name, {
        method: 'POST', headers: { 'Content-Type': 'application/json', 'X-Token': TOKEN },
        body: JSON.stringify(body || {})
      });
      let data = null;
      try { data = await r.json(); } catch (e) { data = null; }
      if (r.ok) return data;
      if (data && data.cancel) throw new Cancelled();
      if (data && data.needPath && attempt === 0) {
        // senza finestre native (solo in prova su Linux): percorso digitato a mano
        const save = !!opts.save;
        if (!opts.quiet) busy(false);
        const p = await promptDlg(save ? 'Salva come (percorso completo)' : 'Apri file (percorso completo)', 'Percorso del file:', opts.suggest || '');
        if (!opts.quiet) busy(true, opts.busyText);
        if (!p) throw new Cancelled();
        body = Object.assign({}, body, { path: p });
        continue;
      }
      if (r.status === 403) throw new Error('Sessione non valida: riavvia il programma.');
      throw new Error((data && data.error) || ('Errore ' + r.status));
    }
    throw new Cancelled();
  } finally {
    if (!opts.quiet) busy(false);
  }
}

function reportError(e) {
  if (e && e.cancel) return;
  console.error(e);
  toast(e && e.message ? e.message : String(e), true);
}

// ------------------------------------------------------------ frequenze e toni
function fmtMHz(hz) {
  if (hz == null || hz === 0) return '';
  const neg = hz < 0; hz = Math.abs(hz);
  let s = Math.floor(hz / 1e6) + '.' + String(hz % 1e6).padStart(6, '0');
  while (s.endsWith('0') && s.length - s.indexOf('.') > 4) s = s.slice(0, -1);
  return (neg ? '-' : '') + s;
}
function fmtShift(rx, tx) {
  if (!rx) return '';
  const d = tx - rx;
  if (d === 0) return '';
  if (Math.abs(d) > 50e6) return 'split';
  return (d > 0 ? '+' : '-') + fmtMHz(Math.abs(d));
}
function parseNum(s) {
  s = String(s).trim().replace(/\s+/g, '').replace('−', '-');
  if (s === '') return NaN;
  const hasDot = s.includes('.'), hasComma = s.includes(',');
  if (hasDot && hasComma) {
    if (s.lastIndexOf(',') > s.lastIndexOf('.')) s = s.replace(/\./g, '').replace(',', '.');
    else s = s.replace(/,/g, '');
  } else if (hasComma) s = s.replace(',', '.');
  if (!/^[+-]?\d*\.?\d+$|^[+-]?\d+\.$/.test(s)) return NaN;
  return parseFloat(s);
}
// "145.6", "145,6", "145600" (kHz), "MHz" facoltativo -> Hz
function parseFreq(s) {
  s = String(s).toLowerCase().replace('mhz', '').trim();
  let v = parseNum(s);
  if (!isFinite(v) || v <= 0) return null;
  if (v < 3000) v *= 1e6; else if (v < 1e7) v *= 1e3;
  const hz = Math.round(v / 10) * 10;
  if (hz < 1e6 || hz > 3e9) return null;
  return hz;
}
// shift in MHz: "+0.6", "-1.6", "0" -> Hz con segno
function parseShift(s) {
  s = String(s).toLowerCase().replace('mhz', '').trim();
  if (s === '' || s === '0' || s === 'simplex' || s === 'iso') return 0;
  let v = parseNum(s);
  if (!isFinite(v)) return null;
  if (Math.abs(v) >= 100) v = v / 1000; // scritto in kHz
  return Math.round(v * 1e5) * 10;
}
let CTCSS = [];
function normTone(s) {
  s = String(s == null ? '' : s).trim().toUpperCase();
  if (s === '' || s === 'OFF' || s === '-' || s === 'NO') return '';
  // DCS: D023N (normale) o D023I (invertito); «D023» = normale
  const m = /^D([0-7]{3})([NI])?$/.exec(s.replace(/\s+/g, ''));
  if (m) return 'D' + m[1] + (m[2] || 'N');
  if (/^D/.test(s)) return null;
  const v = parseNum(s);
  if (!isFinite(v)) return null;
  const t = (Math.round(v * 10) / 10).toFixed(1);
  return CTCSS.includes(t) ? t : null;
}
// solo CTCSS (archivio ponti)
function normCTCSS(s) { const t = normTone(s); return t && t.startsWith('D') ? null : t; }
function fmtTone(t) {
  if (!t) return '';
  if (t.startsWith('K67:')) return 'DCS? (' + t.slice(4) + ')';
  return t;
}
const POWERS = ['Alta', 'Media', 'Bassa'];
const ACC = { 'à': 'a', 'è': 'e', 'é': 'e', 'ì': 'i', 'ò': 'o', 'ù': 'u', 'À': 'A', 'È': 'E', 'É': 'E', 'Ì': 'I', 'Ò': 'O', 'Ù': 'U', '’': "'", '‘': "'" };
function asciiName(s) {
  return String(s || '').replace(/[àèéìòùÀÈÉÌÒÙ’‘]/g, c => ACC[c]).replace(/[^\x20-\x7e]/g, '?').replace(/\s+$/, '');
}
function bandOf(hz) {
  if (hz >= 136e6 && hz <= 174e6) return 'VHF';
  if (hz >= 400e6 && hz <= 480e6) return 'UHF';
  return 'Altre';
}

// ------------------------------------------------------------ finestre di dialogo
function openModal(html, opts) {
  opts = opts || {};
  const ov = h('<div class="overlay"></div>');
  const dlg = h(html);
  ov.appendChild(dlg);
  $('#modal-root').appendChild(ov);
  const close = () => { ov.remove(); document.removeEventListener('keydown', onKey, true); };
  const onKey = (e) => {
    if ($('#modal-root').lastElementChild !== ov) return;
    if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); if (opts.onEsc) opts.onEsc(); else close(); }
    e.stopPropagation();
  };
  document.addEventListener('keydown', onKey, true);
  setTimeout(() => { const f = dlg.querySelector('[autofocus]') || dlg.querySelector('input,select,button.primary'); if (f) f.focus(); }, 30);
  return { el: dlg, close };
}
function modalOpen() { return $('#modal-root').children.length > 0; }

function alertDlg(title, text, list) {
  return new Promise(res => {
    const m = openModal(`<div class="dialog mid"><header>${esc(title)}</header><div class="body"><p>${esc(text).replace(/\n/g, '<br>')}</p>
      ${list && list.length ? '<ul class="warnlist">' + list.map(x => '<li>' + esc(x) + '</li>').join('') + '</ul>' : ''}</div>
      <footer><button class="btn primary" data-ok>OK</button></footer></div>`, { onEsc: () => { m.close(); res(); } });
    m.el.querySelector('[data-ok]').onclick = () => { m.close(); res(); };
  });
}
// choices: [{id, label, primary}]
function chooseDlg(title, text, choices, list) {
  return new Promise(res => {
    const m = openModal(`<div class="dialog mid"><header>${esc(title)}</header><div class="body"><p>${esc(text).replace(/\n/g, '<br>')}</p>
      ${list && list.length ? '<ul class="warnlist">' + list.map(x => '<li>' + esc(x) + '</li>').join('') + '</ul>' : ''}</div>
      <footer>${choices.map(c => `<button class="btn ${c.primary ? 'primary' : ''}" data-id="${esc(c.id)}">${esc(c.label)}</button>`).join('')}</footer></div>`,
      { onEsc: () => { m.close(); res(null); } });
    $$('footer button', m.el).forEach(b => b.onclick = () => { m.close(); res(b.dataset.id); });
  });
}
async function confirmDlg(title, text, okLabel) {
  return (await chooseDlg(title, text, [{ id: 'no', label: 'Annulla' }, { id: 'ok', label: okLabel || 'OK', primary: true }])) === 'ok';
}
function promptDlg(title, label, value) {
  return new Promise(res => {
    const m = openModal(`<div class="dialog mid"><header>${esc(title)}</header><div class="body">
      <div class="row"><label class="l">${esc(label)}</label><input type="text" style="flex:1" autofocus value="${esc(value || '')}"></div></div>
      <footer><button class="btn" data-no>Annulla</button><button class="btn primary" data-ok>OK</button></footer></div>`,
      { onEsc: () => { m.close(); res(null); } });
    const inp = m.el.querySelector('input');
    const ok = () => { m.close(); res(inp.value); };
    m.el.querySelector('[data-ok]').onclick = ok;
    m.el.querySelector('[data-no]').onclick = () => { m.close(); res(null); };
    inp.addEventListener('keydown', e => { if (e.key === 'Enter') ok(); });
  });
}

// ------------------------------------------------------------ menu a tendina
function attachMenu(btn, items) {
  const wrap = h('<span class="menu-wrap"></span>');
  btn.replaceWith(wrap);
  wrap.appendChild(btn);
  const menu = h('<div class="menu"></div>');
  for (const it of items) {
    const b = h(`<button>${esc(it.label)}</button>`);
    b.onclick = (e) => { e.stopPropagation(); menu.classList.remove('open'); it.run(); };
    menu.appendChild(b);
  }
  wrap.appendChild(menu);
  btn.addEventListener('click', (e) => {
    e.stopPropagation();
    const was = menu.classList.contains('open');
    $$('.menu.open').forEach(m => m.classList.remove('open'));
    if (!was) menu.classList.add('open');
  });
  return wrap;
}
document.addEventListener('click', () => $$('.menu.open').forEach(m => m.classList.remove('open')));

// menu a comparsa sotto un elemento (voci create al momento)
function showMenuAt(anchor, items) {
  $$('.ctxmenu').forEach(m => m.remove());
  const m = h('<div class="menu open ctxmenu" style="position:fixed"></div>');
  for (const it of items) {
    const b = h(`<button>${esc(it.label)}</button>`);
    b.onclick = (e) => { e.stopPropagation(); m.remove(); Promise.resolve().then(it.run).catch(reportError); };
    m.appendChild(b);
  }
  document.body.appendChild(m);
  const r = anchor.getBoundingClientRect();
  m.style.left = Math.min(r.left, innerWidth - m.offsetWidth - 8) + 'px';
  m.style.top = Math.min(r.bottom + 2, innerHeight - m.offsetHeight - 8) + 'px';
  setTimeout(() => document.addEventListener('mousedown', function f(ev) { if (!m.contains(ev.target)) { m.remove(); document.removeEventListener('mousedown', f); } }), 0);
}

function debounce(fn, ms) {
  let t = null;
  return function () { const a = arguments; clearTimeout(t); t = setTimeout(() => fn.apply(this, a), ms); };
}

function plural(n, one, many) { return n + ' ' + (n === 1 ? one : many); }
