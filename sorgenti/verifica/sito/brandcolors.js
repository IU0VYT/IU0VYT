// IU0VYT Flash — colori e ordine delle marche (da ui/gallery.js) per il sito
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva
// legge da ui/gallery.js BRAND_COLOR, brandColor e brandOrder; stampa "Marca|colore" in ordine
// 4.12: ui/gallery.js si legge sempre dalla cartella dei sorgenti (non da un percorso
// passato sulla riga di comando: segnalazione di CodeQL)
const fs = require('fs');
const path = require('path');
const src = fs.readFileSync(path.join(__dirname, '..', '..', 'ui', 'gallery.js'), 'utf8');
const pick = re => { const m = src.match(re); if (!m) throw new Error('non trovato ' + re); return m[0]; };
const code = pick(/const BRAND_COLOR = \{[\s\S]*?\};/) + '\n' + pick(/function brandColor\(b\) \{[\s\S]*?\n\}/) + '\n' + pick(/const brandOrder = .*;/);
const f = new Function(code + '\nreturn { brandColor, brandOrder };')();
const brands = fs.readFileSync(0, 'utf8').split('\n').map(s => s.trim()).filter(Boolean);
[...new Set(brands)].sort(f.brandOrder).forEach(b => console.log(b + '|' + f.brandColor(b)));
