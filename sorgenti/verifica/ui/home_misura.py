#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
#
# Verifica dell'interfaccia (4.42): misura la scheda Home a 14 dimensioni di finestra e nelle lingue indicate: tutto deve stare nella pagina, senza scorrimento e senza testi tagliati.
# Serve Python con «playwright» e un Chrome installato (percorso in CHROME, altrimenti
# /usr/bin/google-chrome). Si avvia il programma (su Linux stampa l'indirizzo) e poi:
#   home_misura.py <indirizzo stampato dal programma> [cartella per le schermate] [lingue separate da virgole]
# L'ultima riga deve essere «NON A POSTO: 0».
import os, sys
from playwright.sync_api import sync_playwright
url = sys.argv[1]
out = sys.argv[2] if len(sys.argv) > 2 else None
langs = sys.argv[3].split(',') if len(sys.argv) > 3 else ['it']
# finestra ingrandita: schermo / scala di Windows, meno barra del titolo e barra delle applicazioni
sizes = [(1920, 1009), (1536, 793), (1280, 665), (1366, 697), (1280, 649), (1600, 829), (2560, 1369),
         (1024, 697), (1264, 761), (1011, 609), (884, 561), (1097, 577), (3840, 2089), (960, 500)]
JS = """() => {
  const p = document.querySelector('#page-home');
  const last = p.querySelector('.home-small');
  const r = last ? last.getBoundingClientRect() : null;
  const pr = p.getBoundingClientRect();
  const cut = [...p.querySelectorAll('.hb-grid li, .home-card span, .home-card b')].filter(e => e.scrollWidth > e.clientWidth + 1).length;
  const cards = [...p.querySelectorAll('.home-card')].map(c => Math.round(c.getBoundingClientRect().top));
  const inner = p.querySelector('.home-inner');
  const meta = [...p.querySelectorAll('.home-meta > div, .home-links span, .home-start, .home-card, .hb')].map(e => e.getBoundingClientRect());
  const fuori = meta.filter(q => q.bottom > pr.bottom + 0.5 || q.top < pr.top - 0.5 || q.left < pr.left - 0.5 || q.right > pr.right + 0.5).length;
  const fs = parseFloat(getComputedStyle(p.querySelector('.home-card span')).fontSize);
  return {zoom: inner.style.zoom || '1', fuori: fuori, lang: document.documentElement.lang, font: fs, righeMarche: new Set([...p.querySelectorAll('.hb-grid li')].map(e => Math.round(e.getBoundingClientRect().top))).size,
          client: p.clientHeight, scroll: p.scrollHeight, bottom: r ? Math.round(r.bottom - pr.top) : -1,
          tagliati: cut, righeCard: new Set(cards).size, sw: p.scrollWidth, cw: p.clientWidth,
          brands: p.querySelectorAll('.hb-grid li').length};
}"""
bad = 0
with sync_playwright() as pw:
    b = pw.chromium.launch(executable_path=os.environ.get('CHROME', '/usr/bin/google-chrome'), args=['--no-sandbox'])
    for lang in langs:
        for (w, h) in sizes:
            ctx = b.new_context(viewport={'width': w, 'height': h})
            pg = ctx.new_page()
            pg.goto(url)
            pg.wait_for_selector('#page-home .home-small', timeout=20000)
            if True:
                pg.evaluate("l => I18N.choose(l)", lang)
                pg.wait_for_timeout(900)
            pg.wait_for_timeout(500)
            m = pg.evaluate(JS)
            ok = m['scroll'] <= m['client'] and m['sw'] <= m['cw'] and m['tagliati'] == 0 and m['fuori'] == 0
            bad += 0 if ok else 1
            print(f"{lang} {w}x{h}: {'OK ' if ok else 'NO '} alto {m['scroll']}/{m['client']} fine testo {m['bottom']} largo {m['sw']}/{m['cw']} tagliati {m['tagliati']} righe riquadri {m['righeCard']} marche {m['brands']} su {m['righeMarche']} righe; zoom {m['zoom']} fuori {m['fuori']} lingua {m['lang']}")
            if out and lang in ('it', 'de', 'ar') :
                pg.screenshot(path=f"{out}/home_{lang}_{w}x{h}.png")
            ctx.close()
    b.close()
print('NON A POSTO:', bad)
