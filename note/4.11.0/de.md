# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 für Microsoft Windows 10/11

### Zonennamen
- Statt «Zone 1», «Zone 2»… kann jede Zone einen Namen haben, wie in den Programmen des Herstellers: Schaltfläche «✎ Zonennamen…» oder Doppelklick auf die Zone. Die Zonenleiste und die untere Zeile zeigen «Nummer · Name».
- Funkgeräte: Radtel RT-950 Pro, Bajeton BJ-9000 und Hiroyasu HI-UV850 Pro (Firmware bis 0.23 und ab 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 ohne Bluetooth; Zonen der DMR-Funkgeräte Baofeng RD-5R, DM-1801 und DM-32UV.
- Die Namen werden zusammen mit den Kanälen per Kabel gelesen und geschrieben und in der .img-Datei gespeichert. Länge und Zeichen wie in den Programmen des Herstellers (12 Zeichen, auch chinesische, beim RT-950 Pro; 10 beim RT-920; 16 bei den DMR-Geräten).
- Sicherheit: Der Bereich der Namen wird nur geschrieben, wenn sich die Namen ändern, nachdem er erneut aus dem Funkgerät gelesen wurde; stehen dort Daten, die keine Namen sind, wird nichts geschrieben.

### Sammelbearbeitung
- «✎ Sammelbearbeitung…» (auch über die rechte Maustaste) ändert Leistung, Bandbreite, Suchlauf, TX-Ton, RX-Ton, Ablage und Senden der angehakten (oder markierten) Kanäle auf einmal. Auf «nicht ändern» belassene Einträge bleiben, wie sie sind; Strg+Z macht alles rückgängig.

### Zwei weitere Sprachen (insgesamt 11)
- Koreanisch, mit der Flagge Südkoreas.
- Arabisch, mit der Flagge der Vereinigten Arabischen Emirate und der Seite von rechts nach links (Frequenzen und Zahlen bleiben von links nach rechts).
- Alle Informationsdateien im Ordner `lingue` sind auch ins Koreanische und Arabische übersetzt und in allen Sprachen auf 4.11 aktualisiert.

### Sonstiges
- Website: Die verfügbaren Sprachen werden angegeben; die Liste der Firmware-Dateien wurde entfernt.
- `go test` läuft unter Windows durch (auch i18n_test.go); ein neuer Test prüft, dass LEGGIMI.txt, AUTORI_E_LICENZE.txt und die Lizenz im Hauptordner und in «sorgenti» identisch sind.

Die neuen Formate sind aus dem Code der Herstellerprogramme und aus dem CHIRP-Treiber abgeleitet und mit simulierten Funkgeräten getestet; mit echten Funkgeräten wurden sie nicht getestet: Bewahren Sie immer die Sicherungskopien auf.

Galerie: 489 Modelle von 44 Marken (451 mit Kanälen, 396 mit Einstellungen, 57 mit Firmware-Aktualisierung).
