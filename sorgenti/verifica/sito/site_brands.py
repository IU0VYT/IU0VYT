# IU0VYT Flash — riquadro delle marche del sito (docs/index.html)
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva
# Uso: vedi «SITO: RIQUADRO DELLE MARCHE» in verifica/LEGGIMI_VERIFICA.txt
# rigenera il riquadro delle marche del sito da un elenco "Marca|colore" (stdin)
import sys, re, html
brands=[l.strip().split('|') for l in sys.stdin if l.strip()]
inv_rows=['00100000100','00010001000','00111111100','01101110110','11111111111','10111111101','10100000101','00011011000']
def inv(cls):
    r=''.join(f'<rect x="{x}" y="{y}" width="1" height="1"/>' for y,row in enumerate(inv_rows) for x,c in enumerate(row) if c=='1')
    return f'<svg class="{cls}" viewBox="0 0 11 8" width="22" height="16" aria-hidden="true" shape-rendering="crispEdges">{r}</svg>'
items=''.join(f'<li><span class="hb-dot" style="background:{c}"></span>{html.escape(b)}</li>' for b,c in brands)
block=(f'<!-- marche: generato con sorgenti/verifica/sito dal catalogo del programma (controllato da TestSitoMarche) -->\n'
 f'    <div class="hb" role="group" aria-label="Marche di radio gestite: {len(brands)}"><div class="hb-frame"><div class="hb-in">\n'
 f'      <div class="hb-title">{inv("hb-inv")}<span>PLAYER SELECT · <span data-n="marche">{len(brands)}</span> MARCHE</span>{inv("hb-inv hb-inv2")}</div>\n'
 f'      <ul class="hb-grid">{items}</ul>\n'
 f'    </div></div></div>\n    <!-- /marche -->')
p=sys.argv[1]; s=open(p).read()
if '<!-- marche:' in s:
    s=re.sub(r'<!-- marche:.*?<!-- /marche -->', lambda m: block, s, flags=re.S)
else:
    a='    <div class="tag">CANALI · PONTI · FIRMWARE<span class="tag-sep"> — </span><span class="tag-b">RADIO DI <span data-n="marche">16</span> MARCHE</span></div>\n'
    assert a in s
    s=s.replace(a,'    <div class="tag">GESTIONE CANALI<span class="tag-sep"> - </span><span class="tag-b">FREQUENZE - PONTI RADIO - FIRMWARE</span></div>\n    '+block+'\n')
open(p,'w').write(s)
