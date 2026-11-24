# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 für Microsoft Windows 10/11

### Neu
- **Retevis, 24 neue Modelle, aus den Programmen des Herstellers ermittelt**: B3B, B63B, RT22P, RT622P, RB666, RB666P (gleicher Kanalspeicher wie die neue Version des H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (T18-Familie); RB65, RB66, RB66P (eine Frequenz und ein Ton pro Kanal). Verbindung, Lesen und Schreiben sind von Grund auf neu geschrieben und mit den Werksdatendateien der Programme geprüft. RB66 und RB66P sowie RB669, RT68H und RT668H antworten mit demselben Kürzel: Die richtige Karte muss von Hand gewählt werden.
- **Einstellungen der 24 Retevis-Modelle**: insgesamt 241 Einträge (Squelch, Sendezeitbegrenzung, VOX, Suchlauf, Sprachansage, Batteriesparfunktion, Tastenton, Seitentasten…). Angeboten werden nur die Einträge, die das Programm des Modells anzeigt; die Werte der 24 Werksdateien sind alle zulässige Auswahlen.
- **Programmierpasswort**: bei den 24 Retevis-Modellen und bei der neuen Version des Retevis H777 verhindert es weder das Lesen noch das Schreiben, wie in CHIRP (das Programm des Herstellers hält dagegen an, wenn das Passwort nicht das vom Benutzer eingegebene ist). Es ist ein Eintrag der Einstellungen, im Klartext: Man kann es sehen, ändern und entfernen. Länge und Zeichen entsprechen dem Programm des Herstellers: 6 Zeichen (3 bei RB45, RB645, RB47 und RB66; 4 bei RB66P), Ziffern und Buchstaben (nur Ziffern bei RB65 und RB66). Bei B3H, B63H, RB45, RB645, RB66 und RB666, deren Programm das Passwortfeld nicht anzeigt, weist ein Hinweis darauf hin.
- **Einstellungen für Funkgeräte, die bisher nur die Kanäle hatten**: AnyTone-Familie (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 und DMR-6X2 PRO: 27 Einträge, aus den Speicherbeschreibungen von qdmr und dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 und MD-9600, Retevis RT3S und RT82 (die Einträge des Baofeng DM-1701); Radtel RT-880 und RT-950 Pro mit der Firmware nicFW880 / nicFW950 (17 Einträge, aus der Dokumentation des Autors); TID TD-UV68 (die Einträge des TD-H8); Radtel RT-890 mit der M7OCM-Firmware (53 Einträge).
- **Testversionen**: Eine Kopie mit einer höheren Nummer als die letzte veröffentlichte stabile Version zeigt eine Warnleiste («Testversion (Betatest)…») und bietet – nur sie – die nächste Testversion an, die immer von GitHub heruntergeladen wird. Wer die stabile Version verwendet, sieht keine Testversionen.
- Galerie: 697 Modelle von 54 Marken (647 mit Kanälen, 596 mit Einstellungen).

### Hinweise
- Keines der 24 Retevis-Funkgeräte wurde an echter Hardware getestet, weder die Kanäle noch die Einstellungen: zuerst nur lesen, dann eine einzige harmlose Einstellung. Antwortet das Funkgerät unerwartet, bricht das Lesen mit einer Meldung ab und es wird nichts ins Funkgerät geschrieben.
- Ist bereits ein Programmierpasswort gesetzt, versucht das Programm trotzdem zu lesen und zu schreiben, wie CHIRP: An einem echten Funkgerät wurde das nicht getestet.
- RB66 und RB66P sowie RB669, RT68H und RT668H antworten mit demselben Kürzel: Die richtige Karte muss von Hand gewählt werden.
- Unabhängiges Programm, nicht mit den Funkgeräteherstellern verbunden.
- Vollständige Testreihe bestanden: allgemeiner Teil; 915 Strukturvergleiche und 380 mit CHIRP identische Lese- / Schreibvorgänge.
