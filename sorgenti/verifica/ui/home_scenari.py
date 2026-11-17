#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
#
# Verifica dell'interfaccia (4.42): scenari sulla scheda Home: ridimensionamento, cambio di scheda e ritorno, cambio di lingua, finestra bassissima.
# Serve Python con «playwright» e un Chrome installato (percorso in CHROME, altrimenti
# /usr/bin/google-chrome). Si avvia il programma (su Linux stampa l'indirizzo) e poi:
#   home_scenari.py <indirizzo stampato dal programma>
# L'ultima riga deve essere «NON A POSTO: 0».
import os, sys
from playwright.sync_api import sync_playwright
url = sys.argv[1]
JS = """() => { const p = document.querySelector('#page-home'); const pr = p.getBoundingClientRect();
  const els = [...p.querySelectorAll('.home-meta > div, .home-links span, .home-start, .home-card, .hb, .neon')].map(e => e.getBoundingClientRect());
  const t = [...p.querySelectorAll('.neon-top .nl')].map(e => e.textContent).join('');
  const xs = [...p.querySelectorAll('.neon-top .nl')].map(e => e.getBoundingClientRect().left);
  return {fuori: els.filter(q => q.bottom > pr.bottom + .5 || q.right > pr.right + .5 || q.left < pr.left - .5).length,
    scroll: p.scrollHeight, client: p.clientHeight, zoom: p.querySelector('.home-inner').style.zoom || '1',
    titolo: t, ordine: xs.every((x, i) => i === 0 || x > xs[i - 1])}; }"""
bad = 0
def check(pg, what, zoom1=None):
    global bad
    pg.wait_for_timeout(400)
    m = pg.evaluate(JS)
    ok = m['fuori'] == 0 and m['scroll'] <= m['client'] and m['ordine'] and m['titolo'] == 'IUØVYT'
    if zoom1 is True: ok = ok and m['zoom'] == '1'
    if zoom1 is False: ok = ok and m['zoom'] != '1'
    bad += 0 if ok else 1
    print(('OK ' if ok else 'NO '), what, m)
with sync_playwright() as pw:
    b = pw.chromium.launch(executable_path=os.environ.get('CHROME', '/usr/bin/google-chrome'), args=['--no-sandbox'])
    errs = []
    pg = b.new_context(viewport={'width': 1920, 'height': 1009}).new_page()
    pg.on('pageerror', lambda e: errs.append(str(e)))
    pg.on('response', lambda r: print('   risposta', r.status, r.url.split('?')[0]) if r.status >= 400 else None)
    pg.on('console', lambda m: errs.append(m.text) if m.type == 'error' else None)
    pg.goto(url); pg.wait_for_selector('#page-home .home-small')
    pg.evaluate("I18N.choose('it')")
    check(pg, 'avvio 1920x1009', True)
    pg.set_viewport_size({'width': 1011, 'height': 609}); check(pg, 'rimpicciolita a 1011x609', False)
    pg.set_viewport_size({'width': 1920, 'height': 1009}); check(pg, 'di nuovo grande', True)
    pg.click('.tab[data-tab="gal"]'); pg.wait_for_timeout(300)
    pg.set_viewport_size({'width': 960, 'height': 540}); pg.wait_for_timeout(300)
    pg.click('.tab[data-tab="home"]'); check(pg, 'ridimensionata su una scheda diversa, poi ritorno alla Home', False)
    for l in ('ar', 'ja', 'hi', 'de', 'it'):
        pg.evaluate("l => I18N.choose(l)", l); pg.wait_for_timeout(500); check(pg, 'lingua ' + l + ' a 960x540', False)
    pg.set_viewport_size({'width': 900, 'height': 330}); pg.wait_for_timeout(400)
    m = pg.evaluate(JS)
    ok = m['zoom'] == '0.6' and m['scroll'] > m['client']
    bad += 0 if ok else 1
    print(('OK ' if ok else 'NO '), 'finestra bassissima: zoom minimo e pagina che scorre', m)
    pg.evaluate("document.querySelector('#page-home').scrollTop = 99999"); pg.wait_for_timeout(200)
    y = pg.evaluate("(() => { const p = document.querySelector('#page-home'); return [Math.round(p.querySelector('.home-small').getBoundingClientRect().bottom), Math.round(p.getBoundingClientRect().bottom)]; })()")
    ok = y[0] <= y[1] + 1; bad += 0 if ok else 1
    print(('OK ' if ok else 'NO '), 'scorrendo si arriva in fondo', y)
    pg.set_viewport_size({'width': 1536, 'height': 793}); check(pg, 'ritorno a 1536x793')
    pg.click('.home-start'); pg.wait_for_timeout(300)
    ok = pg.evaluate("document.querySelector('#page-gal').classList.contains('active')"); bad += 0 if ok else 1
    print(('OK ' if ok else 'NO '), 'PRESS START apre i modelli')
    print('errori della pagina:', errs); bad += len(errs)
    b.close()
print('NON A POSTO:', bad)
