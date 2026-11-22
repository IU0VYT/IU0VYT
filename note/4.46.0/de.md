# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub><b>Deutsch</b></sub></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 für Microsoft Windows 10/11

### Korrekturen
- **Radtel RT-880G**: wird nicht mehr als GMRS- / FRS-Funkgerät angezeigt. Es ist ein VHF/UHF-Amateurfunkgerät (mit GPS und APRS), wie das RT-880 und das RT-860; die Warnung «Funkgerät verkauft als GMRS» erscheint vor dem Schreiben nicht mehr.
- **Kontrolle nach dem Schreiben** beim Radtel RT-880G: sie schlug fehl, weil das Funkgerät, das am Ende des Schreibens die Daten speichert und neu startet, zu früh und zu oft angesprochen wurde. Jetzt wartet das Programm 10 Sekunden, ohne es zu stören, und versucht es alle 15 Sekunden erneut, bis zu 90 Sekunden. Dieselben Zeiten gelten für RT-880, RT-860 und RT-817A / RT-871A, die denselben Treiber verwenden.
- Strengere Prüfungen der Werte, die von AR8200, Kenwood und TK-180 empfangen werden (CodeQL-Meldungen).

### Hinweise
- Die Kontrolle wurde mit einem echten RT-880G getestet; RT-880, RT-860 und RT-817A / RT-871A wurden nicht mit echten Geräten getestet. Die anderen Funkgeräte behalten die bisherigen Zeiten.
- Keine neuen Funkgeräte in dieser Version.
- Unabhängiges Programm, nicht mit den Funkgeräteherstellern verbunden.
- Vollständige Testreihe bestanden: allgemeiner Teil; 915 Strukturvergleiche und 380 mit CHIRP identische Lese- / Schreibvorgänge.
