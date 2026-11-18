# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 für Microsoft Windows 10/11

### Neu
- **DMR-Funkgeräte der AnyTone-Familie**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus und BTECH DMR-6X2 und DMR-6X2 PRO (diese beiden waren «nicht kompatibel»). Kanäle und Zonen mit dem USB-Kabel des Funkgeräts; Speicher und Protokoll aus dem freien Programm dmrconfig, mit dessen Originalcode geprüft. Geschrieben werden nur Kanäle, Zonen und Suchlauflisten.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 und Retevis RT3S und RT82** mit der Originalfirmware: gleicher Speicher und gleiches Protokoll wie das Baofeng DM-1701.
- **Baofeng BF-666S** (Treiber des BF-888S) und **BF-T6** (experimentell, Treiber des BF-T8).
- Für die Karten, die «nicht kompatibel» bleiben – UV-8D, BF-UV8, DM-1702 und DM-1703 – steht jetzt der Grund dabei.
- Galerie: 673 Modelle von 54 Marken (623 mit Kanälen, 556 mit Einstellungen).

### Hinweise
- Keines der neuen Funkgeräte wurde an einem echten Gerät getestet: Zuerst nur lesen.
- BTECH DMR-6X2 PRO und Baofeng BF-T6 sind experimentell.
- Unabhängiges Programm, nicht mit den Funkgeräteherstellern verbunden.
- Vollständige Testreihe bestanden: allgemeiner Teil; 918 Strukturvergleiche und 381 mit CHIRP identische Lese- / Schreibvorgänge; AnyTone-Familie mit dem Originalcode von dmrconfig verglichen.

---

Diese Version enthält auch die Neuerungen der Versionen 4.24.0 bis 4.44.0, die keine eigene stabile Veröffentlichung hatten. Hier sind sie, von der neuesten an.

## IU0VYT Flash 4.44.0

### Neu
- **Einstellungen** für acht weitere Funkgeräte, die nur Kanäle hatten: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro und Bajeton DM-5000, aus den Programmen der Hersteller.
- **RT-480 und RT-68**: Beim Schreiben erreicht jetzt auch der Block der allgemeinen Einstellungen das Funkgerät.
- **RT-4D, RT-6D, RT-1000 Pro**: Die Datei des Funkgeräts enthält auch die Einstellungen (früher gespeicherte Dateien bleiben gültig).

### Hinweise
- Keine der neuen Einstellungen wurde an einem echten Funkgerät getestet.
- RT-69 GPS: vorerst nur die Einträge, die einzeln am Herstellerprogramm geprüft wurden.

## IU0VYT Flash 4.43.0

### Neu
- **Fehler behoben**: Bei 33 Funkgeräten (darunter Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 und HA1G, Kenwood TK-2140 / TK-3140) und bei den 8 «Zwillingen» aus 4.42 erschienen die Einstellungen nach dem Lesen per Kabel nicht, sondern nur beim Öffnen einer CHIRP-Datei.
- **Einstellungen** für fünf weitere Radtel: RT-817A / RT-871A, RT-830, RT-850, RT-330 und RT-400, aus den Programmen des Herstellers.

### Hinweise
- Keine der neuen Einstellungen wurde an einem echten Funkgerät getestet.
- Radtel RT-830 und RT-850: Ein Eintrag pro Funkgerät wird vom Herstellerprogramm anders gelesen als geschrieben; hier wird dem Schreiben gefolgt und der Eintrag hat eine Warnung.

## IU0VYT Flash 4.42.0

### Neu
- **Registerkarte Start**: Sie passt jetzt vollständig ins Fenster. Der Kasten der Marken ist breiter und die Seite passt sich von selbst der Fenstergröße an.
- **Einstellungen** für 30 Modelle, die nur Kanäle hatten: Baofeng D-Serie (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro und Radtel RT-13.

### Hinweise
- Keine der neuen Einstellungen wurde an einem echten Funkgerät getestet.
- Baofeng D-Serie: Im Herstellerprogramm sind «Sprachansage» und «Tastenton» zwischen Lesen und Schreiben vertauscht; hier wird dem Schreiben gefolgt und die beiden Einträge haben eine Warnung.

## IU0VYT Flash 4.41.0

### Neu
- **Neue Funkgeräte** (die fehlenden Handgeräte aus CHIRP): Yaesu VX-1R (Kanäle und Einstellungen), Icom IC-91AD und IC-92AD (zwei Bänder, auch D-STAR), BTECH UV-PRO (über Bluetooth, ohne Kabel) und der Empfänger AOR AR8200 (neue Marke).
- **Einstellungen** für Funkgeräte, die nur Kanäle hatten: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Hinweise
- Keines der neuen Funkgeräte wurde in der Praxis getestet: Zuerst nur lesen.
- BTECH UV-PRO: Schreiben ist nur mit der Firmware-Version erlaubt, die der Autor des CHIRP-Treibers getestet hat (147).

## IU0VYT Flash 4.40.0

### Neu
- **Schnellere automatische Suche bei bereits geöffneter Registerkarte**: Die kurze Prüfung (vorher nur für das Baofeng K67 Pro) gilt jetzt auch für die Icom-Geräte, die Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, die Quansheng UV-K5 und Abkömmlinge, die Bajeton BJ-9000, BJ-7800, BJ-7700 und BJ-8300, die Radtel RT-900 / RT-950 / RT-470 und die Funkgeräte mit demselben Speicher. Für die anderen bleibt das vollständige Lesen.
- **Neue Marke: Alinco.** Handgeräte DJ-596, DJ-175, DJ-G7EG und DJ-G7T; Mobilgeräte DR-135T, DR-235T, DR-435T, DR-03T und DR-06T. Nur Kanäle. Funktioniert auch mit Kabeln, die die Daten nicht zurücksenden.

### Hinweise
- Kein Alinco wurde in der Praxis getestet: Zuerst nur lesen.
- Die kurze Prüfung ist mit den simulierten Funkgeräten getestet; kehrt ein Funkgerät nach der Prüfung nicht in den normalen Betrieb zurück, melden Sie es.

## IU0VYT Flash 4.39.0

### Neu
- **Vier neue Icom-Funkgeräte**: ID-80H (D-STAR-Handgerät), ID-880H (D-STAR-Mobilgerät), IC-W32A und IC-W32E (Dualband-Handgeräte mit zwei getrennten Hälften: von 0 bis 99 die VHF-Kanäle, von 100 bis 199 die UHF-Kanäle). Nur Kanäle.

### Hinweise
- Keines der neuen Funkgeräte wurde in der Praxis getestet: Zuerst nur lesen.

## IU0VYT Flash 4.38.0

### Korrekturen
- **Identifizierung**: Das Bajeton BJ-9000 blockierte ab und zu, vor allem bei der ersten Identifizierung. Das Programm hört wieder eine Sekunde lang den Anschluss ab, bevor es zum ersten Mal grüßt, wie in 4.35, auch für das an diesem Kabel bereits erkannte Funkgerät. Die übrigen Verbesserungen aus 4.36 bleiben.

### Hinweise
- Keine neuen Funkgeräte in dieser Version.

## IU0VYT Flash 4.37.0

### Neu
- **Acht neue Yaesu-Handgeräte**: VX-5R und VX-7R (Kanäle); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR und VX-8GE (Kanäle und Einstellungen; bei den VX-8 auch APRS).
- Die Kopie startet am Funkgerät (Modus «CLONE»): Die Registerkarte wird von Hand gewählt und die Anleitung steht im Kabelfenster.

### Hinweise
- Keines der neuen Funkgeräte wurde in der Praxis getestet: Zuerst nur lesen.
- Nicht umgesetzt: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Korrekturen
- **«Identifizieren» und automatische Identifizierung**: Das Programm merkt sich, an welchem Kabel es jedes Funkgerät erkannt hat, und beginnt mit diesen; die erste Suche verliert keine Sekunden mehr mit Lauschen und Vorbereitung; ein nach dem Kabel eingeschaltetes Funkgerät wird öfter erneut versucht.
- Das Fenster «Identifizierung des Funkgeräts läuft» erscheint nur, während das Funkgerät gelesen wird, mit Modell und Fortschritt. Ist das Funkgerät angeschlossen, aber aus, fordert das Programm zum Einschalten auf und erkennt es, sobald es eingeschaltet ist.
- Baofeng K67 Pro bei bereits geöffneter Registerkarte: in ein paar Sekunden erkannt, ohne den ganzen Speicher neu zu lesen.

### Neu
- Registerkarte «Identifizieren»: Schaltfläche **«Diagnose der Identifizierung speichern»** (Zeiten und erste Bytes der letzten Suchen in einer Textdatei, die man dem Autor schicken kann, wenn die Identifizierung langsam ist oder sich irrt; sie enthält weder Kanäle noch persönliche Daten).

### Hinweise
- Keine neuen Funkgeräte in dieser Version.

## IU0VYT Flash 4.35.0

### Korrekturen
- **«Identifizieren» und automatische Identifizierung**: Die «schnelle Runde» aus 4.31 wurde entfernt; mit echten Funkgeräten konnte sie das Funkgerät im Programmiermodus hängen lassen und die Suche verlangsamen. Die Wartezeiten sind wieder wie vorher.
- Die Suche probiert auf jedem PC dieselben Gruppen von Funkgeräten (vorher änderten sie sich von einem Start zum nächsten).
- Das Echo eines Kabels ohne Funkgerät wird nicht mehr für ein Funkgerät gehalten.

### Hinweise
- Keine neuen Funkgeräte in dieser Version.

## IU0VYT Flash 4.34.0

### Neu
- **Neue Funkgeräte**: Yaesu VX-170 und VX-177 (Handgeräte), FT-7800 / FT-7900 und FT-8900 (Mobilgeräte), nur Kanäle, mit den CHIRP-Treibern: Lesen, Schreiben und mit dem Funkgerät ausgetauschte Bytes identisch mit CHIRP.

### Hinweise
- Die Kopie startet am Funkgerät (Modus «CLONE»): Die Registerkarte muss von Hand gewählt werden, «Identifizieren» probiert diese Funkgeräte nicht.

## IU0VYT Flash 4.33.0

### Neu
- **Neue Funkgeräte**: Icom IC-T7H, IC-T8A, IC-Q7A und IC-P7 (Kanäle), mit den CHIRP-Treibern: Lesen, Schreiben und mit dem Funkgerät ausgetauschte Bytes identisch mit CHIRP.

### Hinweise
- IC-Q7A und IC-P7: Der Modus (FM, breites FM, AM) folgt der Frequenz; Überspringen «P» und Abstimmschritt des Kanals bleiben erhalten.

## IU0VYT Flash 4.32.0

### Neu
- **Neue Funkgeräte**: Baofeng T22 in der FRS-Version, TIDRADIO TD-M8S, TD-V2, TD-H4 und TD-H11 (Kanäle und Einstellungen), aus den Programmen der Hersteller.
- **«Identifizieren»** schlägt auch die neuen Funkgeräte vor, die sich wie das BF-888S melden, ohne die Suche zu verlängern: Das Funkgerät wird mit ihrem Treiber erst gelesen, wenn man diese Registerkarte öffnet.

### Korrekturen
- TIDRADIO von 2026 (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 4. Generation, TD-H9, TD-638UV): Die mittlere Leistung wird nur bei den Modellen angeboten, die sie wirklich haben, und nie in den GMRS-Versionen.
- Baofeng T22 PMR446: Verbindung wie im Herstellerprogramm.
- Hinweise zu Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS und TIDRADIO TD-M2 Pro: «Identifizieren» probiert sie; der Satz, der das Gegenteil sagte, wurde entfernt.

## IU0VYT Flash 4.31.0

### Neu
- **«Identifizieren» ist viel schneller**: Zuerst macht es eine schnelle Runde über alle Funkgeräte und erst danach, wenn es nichts gefunden hat, die vollständige wie bisher. Ein Funkgerät am Ende der Liste wird in weniger als einer halben Minute statt in fast zwei gefunden.
- **«Identifizieren» probiert jetzt auch die in den letzten Versionen hinzugefügten Funkgeräte**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G und die TIDRADIO ohne Display.

### Korrekturen
- Retevis RB39 / RB639: Mit der Registerkarte des falschen Landes blieb das Funkgerät bis zum Ausschalten im Programmiermodus.

## IU0VYT Flash 4.30.0

### Neu
- **Allgemeine Einstellungen** für TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 der vierten Generation**, **TD-H9** und **TD-638UV**: programmierbare Tasten, Rauschsperre, VOX, maximale Sendezeit, Anzeige, Sprache, Einschaltmeldungen und mehr.
- **Neue Funkgeräte (Kanäle und Einstellungen)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** und **TD-M2 Pro**.

### Hinweise
- Die neuen Funkgeräte stammen aus den Programmen des Herstellers und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. «Identifizieren» unterscheidet sie nicht voneinander: Wählen Sie die Registerkarte selbst.

## IU0VYT Flash 4.29.0

### Neu
- **Neue Funkgeräte (nur Kanäle)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 der vierten Generation**, **TD-H9** und **TD-638UV**, jeweils in den Ausführungen Basis, HAM und GMRS.
- Das TD-H3 Plus meldet sich wie das TD-H3: Verwenden Sie die Registerkarten «TD-H3 Plus», denn es hat drei Leistungsstufen.

### Hinweise
- Diese Funkgeräte stammen aus dem Herstellerprogramm von 2026 («TIDRadioCPS») und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. Die allgemeinen Einstellungen werden noch nicht verwaltet.

## IU0VYT Flash 4.28.0

### Neu
- **Neue Funkgeräte** (Radioddity-Mobilgeräte): **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) und **DB40-G** (GMRS 40 W, 199 Kanäle), mit Kanälen und allgemeinen Einstellungen. Das UB15-G hat je nach Firmware zwei Registerkarten: vor V1.10 (ohne Namen) und V1.10 (mit Namen).
- Das **DB40-G** gibt es auch in CHIRP: Hier liest und schreibt es dieselben Bytes und dieselben `.img`-Dateien.

### Korrekturen
- Die Tests ließen sich seit 4.26.0 unter Windows nicht kompilieren (drei Testdateien verwendeten die simulierten Funkgeräte, die unter Windows nicht kompiliert werden). Das Programm war nicht betroffen.

### Hinweise
- UB15-G, DB35-G und DB35-UV stammen aus dem Herstellerprogramm und seiner Konfigurationsdatei und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. «Identifizieren» erkennt diese Funkgeräte nicht: Die Registerkarte muss von Hand gewählt werden.

## IU0VYT Flash 4.27.0

### Neu
- **Neue Funkgeräte**: **Retevis C6** (FRS / GMRS, 30 Kanäle), **Retevis C66** (UHF, 16 Kanäle) und **Radioddity Echo-DE / Echo-IT** (PMR446, 16 Kanäle), mit Kanälen und allgemeinen Einstellungen.
- Vor dem Schreiben prüft das Programm, dass Band, Modell und Region des Funkgeräts zum zu schreibenden Speicher passen; Passwort und Kalibrierungen des Funkgeräts werden nie geschrieben.

### Korrekturen
- **Retevis RB39 / RB39P**: steht jetzt in der Registerkarte Modelle bei den Funkgeräten «GMRS / FRS (USA)».

### Hinweise
- Die neuen Funkgeräte stammen aus dem Code der Herstellerprogramme und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. «Identifizieren» erkennt sie nicht: Die Registerkarte muss von Hand gewählt werden.
- Ist im Funkgerät ein Lesepasswort gesetzt, muss es zuerst mit dem Herstellerprogramm entfernt werden.

## IU0VYT Flash 4.26.0

### Neu
- **Neue Funkgeräte**: **Retevis RB39 / RB39P** (FRS, 22 Kanäle) und **RB639 / RB639P** (PMR446, 16 Kanäle), mit Kanälen und allgemeinen Einstellungen. Das Funkgerät gibt an, für welches Land es gebaut ist: Mit der falschen Registerkarte bricht das Lesen ab und nennt die richtige.
- **Neue Einstellungen** für **Baofeng BF-T17 / BF-T18** (PMR446 und FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** und **PR-T3**: Rauschsperre, maximale Sendezeit, VOX, Sprachansage, Tastenton, Batteriesparmodus und die übrigen des Herstellerprogramms.

### Korrekturen
- Änderungsprotokolle in den 12 Sprachen: Datum der 4.25.0 korrigiert (12.11.2026).

### Hinweise
- Die neuen Funkgeräte stammen aus dem Code der Herstellerprogramme und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. «Identifizieren» erkennt sie nicht: Die Registerkarte muss von Hand gewählt werden.

## IU0VYT Flash 4.25.0

### Neu
- **Bajeton BJ-8300 mit der neuen Firmware**: neue Registerkarte für Funkgeräte mit 990 Kanälen in 10 Zonen zu 99 (Herstellerprogramm BT-11_CPS v1.2.6): Kanäle, Zonennamen und Einstellungen. Die bisherige Registerkarte «Bajeton 8300» (960 Kanäle, 15 Zonen zu 64) bleibt, wie sie ist, und **warnt** jetzt, wenn sie die neue Firmware findet.
- **Neue Funkgeräte**: **Radioddity GS-10B Pro** (999 Kanäle und Einstellungen), **Radioddity FS-T3** und **PR-T3** (Kanäle).
- **Bajeton BJ-9000** mit 990 Kanälen: fünf weitere Einstellungen (die des Programms v1.3.3).

### Korrekturen
- **Baofeng BF-T17 / BF-T18, GT-18 und Radioddity PR-T5**: Hohe und niedrige Leistung waren vertauscht. Wer diese Funkgeräte mit einer früheren Version beschrieben hat, sollte sie neu lesen und die Leistung neu wählen.
- «Automatisches Verlassen des Menüs» (BJ-8300, BJ-9000, Radtel RT-950): Die letzte Wahl ist 60 Sekunden, nicht 55.

### Hinweise
- Die neuen Funkgeräte und die neue Registerkarte stammen aus dem Code der Herstellerprogramme und sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen.
- BJ-8300: Um zu wissen, welche Registerkarte zu verwenden ist, sehen Sie am Funkgerät nach, wie viele Zonen es gibt (15 mit 64 Kanälen oder 10 mit 99).

## IU0VYT Flash 4.24.0

### Neu
- **Besser lesbare Registerkarte Modelle**: Die Bezeichnungen der Funkgerätetypen stehen in der Mitte der Seite, und die Linie zwischen kompatiblen und nicht kompatiblen Funkgeräten ist rot und dicker.
- **26 weitere Funkgeräte**, verwaltet mit den Treibern bereits vorhandener baugleicher Funkgeräte:
  - durch Vergleich der Herstellerprogramme erkannt: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - aus der CHIRP-Tabelle gleichwertiger Modelle: Radioddity GA-5S, Retevis RT5 (drei Leistungsstufen), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ und TS-T9+, Rugged RH5R und RH5X, Ansoko A-5R und A-8S, Arcshell AR-5, AR-6 und AR-8, Greaval GV-8S und GV-9S, Tenway TW-325, UV-82 Pro und UV-5R Pro.

### Hinweise
- Die sechs durch Vergleich der Herstellerprogramme erkannten Funkgeräte sind noch nicht mit einem echten Funkgerät getestet: Zuerst nur lesen. Ist das Modell nicht das richtige, antwortet das Funkgerät nicht und wird nicht verändert.
