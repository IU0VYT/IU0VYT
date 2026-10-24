'use strict';
// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Tabella con selezione righe, cella attiva e modifica in linea.

class Grid {
  // opts: { columns, getRows, onEdit, onSelChange, onKey, onDblRow, checkbox, onSort, rowClass }
  // colonna: { key, title, width, cls, html(row) | text(row), edit: { type: 'text'|'select'|'tone', options, value(row) } }
  constructor(wrap, opts) {
    this.wrap = wrap;
    this.o = opts;
    this.rows = [];
    this.sel = new Set();     // id delle righe selezionate
    this.cur = { r: 0, c: 0 };
    this.anchor = 0;
    this.editing = null;
    wrap.tabIndex = 0;
    wrap.addEventListener('mousedown', e => this.onMouseDown(e));
    wrap.addEventListener('dblclick', e => this.onDbl(e));
    wrap.addEventListener('keydown', e => this.onKeyDown(e));
    wrap.addEventListener('click', e => this.onClick(e));
  }

  get cols() { return this.o.columns.filter(c => !c.hidden || !c.hidden()); }

  render(keepScroll) {
    const top = this.wrap.scrollTop, left = this.wrap.scrollLeft;
    this.rows = this.o.getRows();
    if (this.pendingCur != null) {
      const k = this.rows.findIndex(r => r.id === this.pendingCur);
      if (k >= 0) { this.cur.r = k; this.anchor = k; }
      this.pendingCur = null;
    }
    const cols = this.cols;
    if (this.cur.r >= this.rows.length) this.cur.r = Math.max(0, this.rows.length - 1);
    if (this.cur.c >= cols.length) this.cur.c = 0;
    const out = [];
    const tw = cols.reduce((a, c) => a + (c.width || 100), this.o.checkbox ? 30 : 0);
    out.push(`<table class="grid" style="width:${tw}px"><colgroup>`);
    if (this.o.checkbox) out.push('<col style="width:30px">');
    for (const c of cols) out.push(`<col style="width:${c.width || 100}px">`);
    out.push('</colgroup><thead><tr>');
    if (this.o.checkbox) {
      const all = this.rows.length > 0 && this.rows.every(r => r.checked);
      out.push(`<th class="chk"><input type="checkbox" data-all ${all ? 'checked' : ''} title="Seleziona/deseleziona tutte le righe visibili"></th>`);
    }
    for (const c of cols) {
      const s = this.o.sortKey && this.o.sortKey() === c.key ? (this.o.sortDir() > 0 ? ' ▲' : ' ▼') : '';
      out.push(`<th class="${c.sort ? 'sortable' : ''}" data-key="${c.key}" title="${esc(c.tip || c.title)}">${esc(c.title)}${s}</th>`);
    }
    out.push('</tr></thead><tbody>');
    for (let i = 0; i < this.rows.length; i++) {
      const r = this.rows[i];
      if (r.head) {
        out.push(`<tr class="zonehead"><td colspan="${cols.length + (this.o.checkbox ? 1 : 0)}">${esc(r.head)}</td></tr>`);
        continue;
      }
      const cls = [r.cls || ''];
      if (this.sel.has(r.id)) cls.push('sel');
      out.push(`<tr data-i="${i}" class="${cls.join(' ')}">`);
      if (this.o.checkbox) out.push(`<td class="chk"><input type="checkbox" data-chk ${r.checked ? 'checked' : ''}></td>`);
      for (let j = 0; j < cols.length; j++) {
        const c = cols[j];
        let content = c.html ? c.html(r) : esc(c.text ? c.text(r) : '');
        let ccls = c.cls || '';
        if (c.bad && c.bad(r)) ccls += ' bad';
        if (i === this.cur.r && j === this.cur.c) ccls += ' cur';
        const tip = c.cellTip ? c.cellTip(r) : '';
        out.push(`<td class="${ccls}" data-j="${j}"${tip ? ` title="${esc(tip)}"` : ''}>${content}</td>`);
      }
      out.push('</tr>');
    }
    if (this.rows.length === 0 && this.o.emptyText) {
      out.push(`<tr><td colspan="${cols.length + (this.o.checkbox ? 1 : 0)}" style="padding:20px;color:var(--muted);white-space:normal">${esc(this.o.emptyText())}</td></tr>`);
    }
    out.push('</tbody></table>');
    this.editing = null;
    this.wrap.innerHTML = out.join('');
    if (keepScroll !== false) { this.wrap.scrollTop = top; this.wrap.scrollLeft = left; }
  }

  rowAt(e) {
    const tr = e.target.closest('tr[data-i]');
    return tr ? +tr.dataset.i : -1;
  }

  selectOnly(i) {
    this.sel.clear();
    if (this.rows[i] && !this.rows[i].head) this.sel.add(this.rows[i].id);
    this.anchor = i;
  }

  selectRange(a, b) {
    this.sel.clear();
    const [lo, hi] = a < b ? [a, b] : [b, a];
    for (let k = lo; k <= hi; k++) if (this.rows[k] && !this.rows[k].head) this.sel.add(this.rows[k].id);
  }

  onMouseDown(e) {
    if (e.target.closest('.editing')) return;
    if (e.target.matches('input[type=checkbox]')) return;
    const th = e.target.closest('th[data-key]');
    if (th) return;
    const i = this.rowAt(e);
    if (i < 0) return;
    e.preventDefault(); // la tabella viene ridisegnata: il fuoco resta sulla tabella
    const td = e.target.closest('td[data-j]');
    if (this.editing) this.commitEdit();
    if (e.shiftKey) this.selectRange(this.anchor, i);
    else if (e.ctrlKey || e.metaKey) {
      const id = this.rows[i].id;
      if (this.sel.has(id)) this.sel.delete(id); else this.sel.add(id);
      this.anchor = i;
    } else if (!(e.button === 2 && this.sel.has(this.rows[i].id))) this.selectOnly(i);
    this.cur.r = i;
    if (td) this.cur.c = +td.dataset.j;
    this.refreshSel();
    this.wrap.focus({ preventScroll: true });
    if (this.o.onSelChange) this.o.onSelChange();
  }

  // aggiorna evidenziazioni senza ridisegnare (il DOM resta stabile per clic e doppio clic)
  refreshSel() {
    for (const tr of this.wrap.querySelectorAll('tbody tr[data-i]')) {
      const r = this.rows[+tr.dataset.i];
      tr.classList.toggle('sel', !!r && this.sel.has(r.id));
    }
    const old = this.wrap.querySelector('td.cur');
    if (old) old.classList.remove('cur');
    const td = this.wrap.querySelector(`tr[data-i="${this.cur.r}"] td[data-j="${this.cur.c}"]`);
    if (td) td.classList.add('cur');
  }

  onClick(e) {
    const th = e.target.closest('th[data-key]');
    if (th && this.o.onSort) {
      const c = this.cols.find(x => x.key === th.dataset.key);
      if (c && c.sort) this.o.onSort(c.key);
      return;
    }
    if (e.target.matches('input[data-all]')) {
      this.o.onCheckAll(e.target.checked);
      return;
    }
    if (e.target.matches('input[data-chk]')) {
      const i = this.rowAt(e);
      if (i >= 0) {
        if (e.shiftKey && this.lastChk != null) {
          const [lo, hi] = this.lastChk < i ? [this.lastChk, i] : [i, this.lastChk];
          const ids = [];
          for (let k = lo; k <= hi; k++) if (this.rows[k] && !this.rows[k].head) ids.push(this.rows[k].id);
          this.o.onCheck(ids, e.target.checked);
        } else this.o.onCheck([this.rows[i].id], e.target.checked);
        this.lastChk = i;
      }
    }
  }

  onDbl(e) {
    if (e.target.closest('.editing') || e.target.closest('thead') || e.target.matches('input[type=checkbox]')) return;
    const i = this.rowAt(e);
    const td = e.target.closest('td[data-j]');
    if (i >= 0 && td) this.cur = { r: i, c: +td.dataset.j };
    this.startEdit();
  }

  ensureVisible() {
    const tr = this.wrap.querySelector(`tr[data-i="${this.cur.r}"]`);
    if (!tr) return;
    const head = this.wrap.querySelector('thead');
    const hh = head ? head.offsetHeight : 0;
    const top = tr.offsetTop, bottom = top + tr.offsetHeight;
    if (top - hh < this.wrap.scrollTop) this.wrap.scrollTop = top - hh;
    else if (bottom > this.wrap.scrollTop + this.wrap.clientHeight) this.wrap.scrollTop = bottom - this.wrap.clientHeight;
  }

  moveCursor(dr, dc, extend) {
    let r = this.cur.r + dr;
    while (this.rows[r] && this.rows[r].head) r += dr > 0 ? 1 : -1;
    r = Math.max(0, Math.min(this.rows.length - 1, r));
    if (this.rows[r] && this.rows[r].head) r = this.cur.r;
    const c = Math.max(0, Math.min(this.cols.length - 1, this.cur.c + dc));
    this.cur = { r, c };
    if (dr !== 0) {
      if (extend) this.selectRange(this.anchor, r); else this.selectOnly(r);
    }
    if (this.editing) this.render(); else this.refreshSel();
    this.ensureVisible();
    if (this.o.onSelChange) this.o.onSelChange();
  }

  onKeyDown(e) {
    if (this.editing) return;
    if (this.o.onKey && this.o.onKey(e) === true) { e.preventDefault(); return; }
    const k = e.key;
    const pg = Math.max(1, Math.floor(this.wrap.clientHeight / 26) - 1);
    if (k === 'ArrowDown' && !e.altKey) { e.preventDefault(); this.moveCursor(1, 0, e.shiftKey); }
    else if (k === 'ArrowUp' && !e.altKey) { e.preventDefault(); this.moveCursor(-1, 0, e.shiftKey); }
    else if (k === 'ArrowLeft') { e.preventDefault(); this.moveCursor(0, -1); }
    else if (k === 'ArrowRight') { e.preventDefault(); this.moveCursor(0, 1); }
    else if (k === 'PageDown') { e.preventDefault(); this.moveCursor(pg, 0, e.shiftKey); }
    else if (k === 'PageUp') { e.preventDefault(); this.moveCursor(-pg, 0, e.shiftKey); }
    else if (k === 'Home' && e.ctrlKey) { e.preventDefault(); this.moveCursor(-1e6, 0, e.shiftKey); }
    else if (k === 'End' && e.ctrlKey) { e.preventDefault(); this.moveCursor(1e6, 0, e.shiftKey); }
    else if (k === 'Tab') { e.preventDefault(); this.moveCursor(0, e.shiftKey ? -1 : 1); }
    else if (k === 'Enter' || k === 'F2') { e.preventDefault(); this.startEdit(); }
    else if (k.toLowerCase() === 'a' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      this.sel = new Set(this.rows.filter(r => !r.head).map(r => r.id));
      this.refreshSel();
      if (this.o.onSelChange) this.o.onSelChange();
    } else if (k.length === 1 && !e.ctrlKey && !e.altKey && !e.metaKey && k !== ' ') {
      const col = this.cols[this.cur.c];
      if (col && col.edit && col.edit.type !== 'select') { e.preventDefault(); this.startEdit(k); }
    } else if (k === ' ' && this.o.checkbox) {
      e.preventDefault();
      const ids = this.sel.size ? Array.from(this.sel) : (this.rows[this.cur.r] ? [this.rows[this.cur.r].id] : []);
      const row = this.rows.find(r => r.id === ids[0]);
      if (ids.length) this.o.onCheck(ids, !(row && row.checked));
    }
  }

  startEdit(initial) {
    const row = this.rows[this.cur.r];
    let col = this.cols[this.cur.c];
    if (row && col && this.o.editRedirect) {
      const k = this.o.editRedirect(row, col);
      if (k) {
        const j = this.cols.findIndex(c => c.key === k);
        if (j >= 0) {
          this.cur.c = j; col = this.cols[j];
          if (initial != null && k !== 'rx') initial = null;
          this.render();
        }
      }
    }
    if (!row || row.head || !col || !col.edit) return;
    if (col.edit.can && !col.edit.can(row)) return;
    const td = this.wrap.querySelector(`tr[data-i="${this.cur.r}"] td[data-j="${this.cur.c}"]`);
    if (!td) return;
    const ed = col.edit;
    let input;
    const val = ed.value ? ed.value(row) : '';
    if (ed.type === 'select') {
      const opts = typeof ed.options === 'function' ? ed.options(row) : ed.options;
      input = document.createElement('select');
      for (const o of opts) {
        const op = document.createElement('option');
        op.value = o.value; op.textContent = o.label;
        if (String(o.value) === String(val)) op.selected = true;
        input.appendChild(op);
      }
    } else {
      input = document.createElement('input');
      input.type = 'text';
      input.value = initial != null ? initial : val;
      if (ed.type === 'tone') input.setAttribute('list', 'ctcss-list');
      if (ed.maxlength) input.maxLength = ed.maxlength(row);
      input.spellcheck = false;
    }
    td.classList.add('editing');
    td.innerHTML = '';
    td.appendChild(input);
    this.editing = { input, row, col, r: this.cur.r, c: this.cur.c, td };
    input.focus();
    if (initial == null && input.select) input.select();
    input.addEventListener('keydown', ev => {
      ev.stopPropagation();
      if (ev.key === 'Escape') { ev.preventDefault(); this.cancelEdit(); }
      else if (ev.key === 'Enter') { ev.preventDefault(); this.commitEdit(1, 0); }
      else if (ev.key === 'Tab') { ev.preventDefault(); this.commitEdit(0, ev.shiftKey ? -1 : 1); }
      else if ((ev.key === 'ArrowDown' || ev.key === 'ArrowUp') && ed.type !== 'select' && ed.type !== 'tone') {
        ev.preventDefault(); this.commitEdit(ev.key === 'ArrowDown' ? 1 : -1, 0);
      }
    });
    input.addEventListener('blur', () => { setTimeout(() => { if (this.editing && this.editing.input === input) this.commitEdit(); }, 0); });
    if (ed.type === 'select') input.addEventListener('change', () => this.commitEdit());
  }

  cancelEdit() {
    this.editing = null;
    this.render();
    this.wrap.focus({ preventScroll: true });
  }

  async commitEdit(dr, dc) {
    const ed = this.editing;
    if (!ed) return;
    this.editing = null;
    const err = await this.o.onEdit(ed.row, ed.col, ed.input.value);
    if (err) {
      toast(err, true);
      this.render();
      this.cur = { r: ed.r, c: ed.c };
      this.render();
      this.startEdit();
      this.editing && (this.editing.input.value = ed.input.value);
      return;
    }
    this.render();
    if (dr || dc) {
      this.cur = { r: ed.r, c: ed.c };
      this.moveCursor(dr || 0, dc || 0);
    }
    this.wrap.focus({ preventScroll: true });
  }

  selectedIds() { return Array.from(this.sel); }
  // la cella attiva viene posizionata al prossimo render()
  setSelection(ids, cursorId) {
    this.sel = new Set(ids);
    if (cursorId != null) this.pendingCur = cursorId;
  }
  curId() { const r = this.rows[this.cur.r]; return r && !r.head ? r.id : null; }
}
