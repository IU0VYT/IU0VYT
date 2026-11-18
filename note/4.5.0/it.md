# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub><b>Italiano</b></sub></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="es.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="pt.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="ru.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

Pagina della release su GitHub (in inglese), con il download: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 per Microsoft Windows 10/11

### Firmware Retevis RT85 (44 modelli con l'aggiornamento del firmware)
- **Retevis RT85** con i file .hex di Retevis, per esempio «RT85-Firmware-V1.53.hex». Usa il protocollo del programma di Retevis «RT85 10W_5W_UpData», che è stato fatto girare su una radio simulata: blocchi e comandi inviati sono identici byte per byte.
- Procedura di Retevis: radio spenta, accesa tenendo premuti **PTT e P1** (indicatore verde). Poi si sceglie la potenza della radio (**5 W o 10 W**, come nel programma di Retevis) e si scrive RT85 per confermare.

### Corretto: aggiornamento della Retevis RA89
- La 4.4 aspettava dalla radio una risposta sbagliata alla presentazione: con una RA89 vera l'aggiornamento non partiva (la radio non veniva toccata). Ora la risposta attesa e le pause sono quelle del programma di Retevis.

### Porte, schede Identifica e Converti
- Le porte dei cavi sono sempre in ordine progressivo, su ogni sistema: COM1, COM2, COM3… COM10. Anche l'identificazione automatica parte dalla prima porta.
- Scheda Identifica: «Le radio già identificate in passato vengono riconosciute più velocemente».
- Scheda Converti: «Scegli un file Baofeng, Bajeton, .img di tutte le altre marche di radio (anche salvate con CHIRP), CSV di CHIRP, tabelle Excel o CSV».

### Sito web
- Tolto il testo delle note della versione sotto il pulsante di download.
- Il titolo è più in alto: il numero di versione sotto «PRESS START» si vede intero anche sugli schermi dei portatili.

Galleria: 386 modelli di 27 marche (348 con canali, 302 con impostazioni, 44 con l'aggiornamento del firmware).
