# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 per Microsoft Windows 10/11

### Correzioni
- **Radtel RT-880G**: non è più indicata come radio GMRS / FRS. È un ricetrasmettitore radioamatoriale VHF/UHF (con GPS e APRS), come la RT-880 e la RT-860; prima della scrittura non compare più l'avviso «Radio venduta come GMRS».
- **Verifica dopo la scrittura** della Radtel RT-880G: falliva perché la radio, che a fine scrittura salva i dati e si riavvia, veniva cercata troppo presto e troppo spesso. Ora il programma aspetta 10 secondi senza disturbarla e riprova ogni 15 secondi, fino a 90 secondi. Gli stessi tempi valgono per RT-880, RT-860 e RT-817A / RT-871A, che usano lo stesso driver.
- Controlli più severi sui valori ricevuti da AR8200, Kenwood e TK-180 (segnalazioni CodeQL).

### Note
- La verifica è stata provata su una RT-880G vera; RT-880, RT-860 e RT-817A / RT-871A non sono state provate dal vero. Le altre radio mantengono i tempi di prima.
- Nessuna radio nuova in questa versione.
- Programma indipendente, non affiliato ai costruttori delle radio.
- Suite completa superata: parte generale; 915 confronti di struttura e 380 letture / scritture identiche a CHIRP.
