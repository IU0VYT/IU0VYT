# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 per Microsoft Windows 10/11

### Nomi delle zone
- Al posto di «Zona 1», «Zona 2»… ogni zona può avere un nome, come nei programmi del costruttore: pulsante «✎ Nomi delle zone…» oppure doppio clic sulla zona. La barra delle zone e la riga in basso mostrano «numero · nome».
- Radio: Radtel RT-950 Pro, Bajeton BJ-9000 e Hiroyasu HI-UV850 Pro (firmware fino a 0.23 e dalla 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 senza Bluetooth; zone delle radio DMR Baofeng RD-5R, DM-1801 e DM-32UV.
- I nomi si leggono e si scrivono via cavo insieme ai canali e si salvano nel file .img. Lunghezza e caratteri come nei programmi del costruttore (12 caratteri, anche cinesi, sulla RT-950 Pro; 10 sulla RT-920; 16 sulle DMR).
- Sicurezza: l'area dei nomi si scrive solo se i nomi cambiano, dopo averla riletta dalla radio; se la radio ha lì dati che non sono nomi, non si scrive nulla.

### Modifica in blocco
- «✎ Modifica in blocco…» (anche dal tasto destro) cambia insieme potenza, banda, scansione, tono TX, tono RX, shift e trasmissione dei canali spuntati (o selezionati). Le voci lasciate su «non cambiare» restano come sono; Ctrl+Z annulla tutto.

### Due lingue in più (11 in tutto)
- Coreano, con la bandiera della Corea del Sud.
- Arabo, con la bandiera degli Emirati Arabi Uniti e la pagina da destra a sinistra (frequenze e numeri restano da sinistra a destra).
- Tutti i file informativi della cartella `lingue` sono tradotti anche in coreano e in arabo e aggiornati alla 4.11 in tutte le lingue.

### Altro
- Sito: indicate le lingue disponibili; tolto l'elenco dei file di firmware.
- `go test` passa su Windows (anche i18n_test.go); un nuovo test controlla che LEGGIMI.txt, AUTORI_E_LICENZE.txt e la licenza siano identici nella cartella principale e in «sorgenti».

I nuovi formati sono ricavati dal codice dei programmi del costruttore e dal driver di CHIRP e provati con radio simulate; non sono stati provati con radio vere: conserva sempre le copie di sicurezza.

Galleria: 489 modelli di 44 marche (451 con canali, 396 con impostazioni, 57 con l'aggiornamento del firmware).
