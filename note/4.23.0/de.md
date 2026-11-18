# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 für Microsoft Windows 10/11

### Neu
- **Kenwood TK-481 und TK-981** (900 MHz, Kanäle in Systeme aufgeteilt) und die Mobilgeräte **TK-780 und TK-880**: Kanäle und Einstellungen, mit den CHIRP-Treibern.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: jetzt auch die Einstellungen des Funkgeräts.
- **Radtel RT-13**: Kanäle (22, ohne Namen), aus dem Programm des Herstellers abgeleitet.
- **Radtel RT-69 GPS**: Kanäle (999, mit Namen aus 12 Zeichen), aus dem Programm des Herstellers abgeleitet; vor dem Schreiben wird geprüft, dass jeder Kanal in den Bändern des Funkgeräts liegt.
- Robuster bei beschädigten Dateien: zwei Abstürze beim Lesen der Töne von Baofeng UV-B5 / UV-B6 und UV-3R behoben.
- Konvertierungen: Ein Kanal, den das Zielfunkgerät nicht speichern kann, wird mit einer Warnung übersprungen, statt die Konvertierung abzubrechen (Firmware IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galerie: 569 Modelle von 47 Marken (515 mit Kanälen, 433 mit Einstellungen).

### Hinweise
- TK-481 / TK-981: Es werden nur die Systeme verwendet, die das Funkgerät bereits hat; Systeme mit gemeinsamen Ressourcen (Trunking) werden nicht verändert.
- RT-13 und RT-69 GPS: Ist am Funkgerät ein Passwort gesetzt, entfernen Sie es zuerst mit dem Radtel-Programm. Aus den Herstellerprogrammen abgeleitet und nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen.
- Unabhängiges Programm, nicht mit den Funkgeräteherstellern verbunden.

---

Diese Version enthält auch die Neuerungen der Versionen 4.13.0 bis 4.22.0, die nur als Testversionen veröffentlicht wurden. Hier sind sie, von der neuesten an.

## IU0VYT Flash 4.22.0

### Neu
- **Radtel RT-480 und RT-68**: Kanäle und Namen (256 Kanäle, nummeriert von 0 bis 255, mit Namen aus 12 Zeichen), aus dem Programm des Herstellers abgeleitet. Änderbar sind Frequenzen, Name, Leistung, Bandbreite, Suchlauf und Töne; die übrigen Einträge des Kanals bleiben, wie sie sind.
- **Radtel RT-830 und RT-850**: Kanäle (128, ohne Namen), aus den Programmen des Herstellers abgeleitet. Änderbar sind Frequenzen, Leistung, Bandbreite, Suchlauf und Töne.
- **Radtel RT-20M**: wird mit der Registerkarte Ruyage UV58Plus geöffnet (gleicher Speicher wie das iRadio UV-5118plus).

### Hinweise
- RT-480 / RT-68: Es werden nur Kanäle und Namen geschrieben; Einstellungen, VFO und DTMF des Funkgeräts bleiben unberührt. Frequenzen in Schritten von 1,25 kHz; jeder Kanal muss eine Sendefrequenz haben.
- RT-830 / RT-850: Einstellungen und VFO kommen unverändert zurück; das Passwort des Funkgeräts wird nie geschrieben. Ist am Funkgerät ein Passwort gesetzt, entfernen Sie es zuerst mit dem Radtel-Programm.
- Alle aus den Herstellerprogrammen abgeleitet und nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen.

## IU0VYT Flash 4.21.0

### Neu
- **Radtel RT-5D (DMR)**: Kanäle (16 Zonen zu 64, mit Namen aus 12 Zeichen), aus dem Programm des Herstellers abgeleitet. Änderbar sind Frequenzen, Name, Leistung, Bandbreite, Suchlauf und Töne; die übrigen Einträge des Kanals bleiben, wie sie sind.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** und die Mobilgeräte **TK-7180, TK-8180** (auch E): Kanäle, mit den CHIRP-Treibern. Die Liste hat 250 Positionen für jede der ersten 32 Zonen; es werden nur die Zonen verwendet, die das Funkgerät bereits hat, insgesamt 512 Kanäle.

### Hinweise
- RT-5D: Das Lesen holt den ganzen Speicher, und das Schreiben sendet ihn vollständig so zurück, wie er in der Datei steht, wie es das Radtel-Programm tut. Vor dem Schreiben frisch lesen: Eine alte Datei würde die damaligen Kontakte und Einstellungen zurückbringen. Nur mit einem simulierten Funkgerät getestet.
- TK-x180: nur Kanäle, keine Einstellungen; Zonen werden weder angelegt noch gelöscht. Hat das Funkgerät nicht konventionelle Zonen (Trunking), wird das Lesen abgelehnt.

## IU0VYT Flash 4.20.0

### Neu
- **Radtel RT-4D und RT-6D (DMR) und RT-1000 Pro**: Kanäle (1024, mit Namen aus 16 Zeichen) und Zonen (250), aus den Programmen des Herstellers abgeleitet. Änderbar sind Frequenzen, Name, Leistung, Bandbreite, Töne und beim RT-4D und RT-6D der Suchlauf; die übrigen Einträge des Kanals bleiben, wie sie sind. Es werden nur Kanäle und Zonen geschrieben: Kontakte, Gruppen, Nachrichten, UKW-Radio und Einstellungen bleiben unberührt.

### Hinweise
- Nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen und jedes Problem melden.
- «Identifizieren» unterscheidet diese Funkgeräte nicht von den anderen Radtel- und Bajeton-Geräten mit derselben Verbindung: Die Registerkarte muss von Hand gewählt werden.
- RT-1000 Pro: Der Ausschluss vom Suchlauf wird nicht verwaltet (er bleibt, wie er am Funkgerät ist).

## IU0VYT Flash 4.19.0

### Neu
- **Radtel RT-817A / RT-871A**: Kanäle (1024, mit Namen aus 16 Zeichen), aus dem Programm des Herstellers abgeleitet. Gleiche Verbindung und gleiches Kanalformat wie das RT-880G, ohne Zonen. Gelesen und geschrieben werden nur die Bereiche, die das Radtel-Programm nutzt (Einstellungen, Kanäle, UKW-Radio); das Logo bleibt unberührt.

### Hinweise
- Nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen und jedes Problem melden.
- «Identifizieren» unterscheidet das RT-817A / RT-871A nicht vom RT-880G: Die Registerkarte muss von Hand gewählt werden.

## IU0VYT Flash 4.18.0

### Neu
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: professionelle Handfunkgeräte mit in Zonen (Gruppen) aufgeteilten Kanälen, Kanäle und Einstellungen. Die Liste zeigt die 250 Speicherplätze; ein neuer Kanal kommt in die Zone des vorhergehenden Kanals.
- **Quansheng UV-K5 / UV-K6 mit der Firmware R3b0rn von fagci**: Kanäle (298 mit dem Originalspeicher, bis 1024 mit größerem Speicher). Einstellungen, VFO, Voreinstellungen und SSB-Korrektur werden nie geschrieben.
- **Radtel RT-330 und RT-400**: Kanäle, aus den Programmen des Herstellers abgeleitet.

### Hinweise
- Die CHIRP-Treiber der TK-Handfunkgeräte sind als experimentell gekennzeichnet. Zonen werden im Programm weder angelegt noch gelöscht.
- Die Firmware R3b0rn, das RT-330 und das RT-400 sind nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen und jedes Problem melden.
- Mit der Firmware R3b0rn erhält ein Kanal ohne Namen die Frequenz als Namen (für diese Firmware gibt es keinen Kanal ohne Namen).
- Die Registerkarte RT-400 lehnt Funkgeräte ab, die sich als RT-470, RT-630, RT-495 oder HI-8811 melden: Nur mit dem RT-400 verwenden.

## IU0VYT Flash 4.17.0

### Neu
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: Kanäle werden einzeln gelesen und geschrieben, mit normal eingeschaltetem Funkgerät (9600 Baud). Beim Schreiben werden nur die geänderten Kanäle gesendet.
- **Professionelle Kenwood-TK-Handfunkgeräte**: Kanäle und Einstellungen.
  - Serie 60: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Serie 60G: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: Man verwendet die Registerkarte «Radtel RT-880G» (gleiche Verbindung und gleicher Kanalspeicher).

### Hinweise
- Die Kenwood TH-F6/F7, TH-G71, TH-D7 und TH-K2 verwenden die Befehle der CHIRP-Treiber und sind nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen und jedes Problem melden.
- Die CHIRP-Treiber der TK-Handfunkgeräte sind als experimentell gekennzeichnet.
- Das RT-860 stammt aus dem Vergleich mit dem Herstellerprogramm und wurde nicht an einem echten Funkgerät getestet.

## IU0VYT Flash 4.16.0

### Neu
- **OpenGD77**: Kanäle der DMR-Funkgeräte mit der freien Firmware OpenGD77.
  - Radioddity GD-77 und GD-77S, Baofeng DM-1801 und RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 und MD-2017, Retevis RT3S und RT84.
  - Das Funkgerät erscheint als serieller USB-Anschluss; geschrieben werden nur die geänderten Kanäle und Zonen, nie Einstellungen, Kontakte, Gruppen und Kalibrierung.
- **Quansheng UV-K5, weitere freie Firmware**:
  - **NUNU** von kamilsss655: Kanäle und Einstellungen.
  - **UV-KX** von joaquim.org: Kanäle und Einstellungen; die Kalibrierung wird nie geschrieben.
  - Die Registerkarte «OSFW» gilt auch für Originalfirmware, die mit erweiterten Bändern verändert wurde (uvmod-kitchen, Sammlung von Tunas1337).
- **Radtel RT-880** ohne GPS mit der Originalfirmware: Man verwendet die Registerkarte «Radtel RT-880G».

### Hinweise
- OpenGD77 ist nach dem im freien Programm qdmr beschriebenen Protokoll geschrieben und nur mit einem simulierten Funkgerät getestet: Zuerst nur lesen, die Sicherungskopie aufbewahren und jedes Problem melden.
- NUNU und UV-KX sind Byte für Byte mit den CHIRP-Treibern ihrer Autoren verglichen.

## IU0VYT Flash 4.15.0

### Neu
- **D-STAR und C4FM**: Kanäle per Kabel für digitale Amateurfunk-Handgeräte, mit .img-Dateien, die mit CHIRP kompatibel sind. Die Spalte «Modus» zeigt «D-STAR» oder «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) und IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (die Kopie startet am Funkgerät: Die Anleitung steht im Kabelfenster).
  - **Kenwood** (neue Marke, insgesamt 47) TH-D72, TH-D74, TH-D75, mit dem USB-Kabel.
  - Einstellungen für IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D und TH-D72.
- **Freie Firmware**:
  - **nicFW 2.x** des TIDRADIO TD-H3: Kanäle und wichtigste Einstellungen.
  - **nicFW880 / nicFW950** von Radtel RT-880, RT-880G und RT-950 Pro: Kanäle.
  - **F4HWN Fusion 5.x** von Quansheng UV-K1 und UV-K5 V3: 1024 Kanäle und Einstellungen.
  - Mit nicFW wird die Kalibrierung des Funkgeräts nie geschrieben.
- **Kenwood in der Galerie**: auch die PMR446- und DMR/NXDN-Handgeräte, unter den nicht kompatiblen (sie lassen sich nur mit dem Händlern vorbehaltenen Programm KPG programmieren).

### Korrekturen
- Einstellungen: Ein nie benutztes Textfeld wird leer gelesen statt als «unerwarteter Wert»; ein unveränderter Wert wird nicht neu geschrieben.
- Puxing PX-888K: sechs Einstellungen entfernt, die einen falschen Wert schreiben konnten (BOT/EOT-Codes und DTMF-Zeiten).
- Yaesu VX-3: In DTMF-Codes ist das Leerzeichen nicht mehr erlaubt (das Funkgerät hat es nicht).

### Hinweise
- Treiber von Grund auf neu geschrieben, ausgehend von denen von CHIRP (GNU GPL), von der von nicsure veröffentlichten Dokumentation (nicFW) und vom F4HWN-Treiber, und mit simulierten Funkgeräten Byte für Byte verglichen. Die echten Funkgeräte dieser Version wurden noch nicht getestet: Sicherungskopie aufbewahren und jedes Problem melden.
- Für Icom, Yaesu und Kenwood gibt es keine Firmware-Aktualisierung (sie erfolgt mit den Herstellerprogrammen).

## IU0VYT Flash 4.14.0

### Neu
- **Yaesu und Icom**: zwei neue Marken (insgesamt 46) und 13 Funkgeräte mehr, mit Kanälen per Kabel und .img-Dateien, die mit CHIRP kompatibel sind.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (Kabel SCU-35) und FT-60R/E (Kopie am Funkgerät gestartet, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (Kabel OPC-478).
  - Einstellungen für alle außer dem IC-T70. Für diese Funkgeräte gibt es keine Firmware-Aktualisierung.
- **Zwei weitere Sprachen**: Malaiisch (Bahasa Melayu) und Hindi (हिन्दी), insgesamt 13: Programm, Website, Anleitung und Informationsdateien, README und Sicherheitsrichtlinie.
- **Konverter und «An ein anderes Funkgerät weitergeben»** zu Yaesu und Icom: Töne, die das Funkgerät nicht hat, werden mit einer Warnung entfernt, und Kanäle, die es nicht speichern kann, werden mit einer Warnung übersprungen, statt den ganzen Vorgang abzubrechen.

### Korrekturen
- Übersetzungen: Fünf fett gedruckte Wörter der Anleitungen blieben in allen Sprachen italienisch; zwei falsche Sätze auf Koreanisch und Arabisch.
- Wouxun KG-UV8D: interne Feldnamen an das heutige CHIRP angepasst (keine Auswirkung auf die Kanäle).

### Hinweise
- Treiber von Grund auf neu geschrieben, ausgehend von denen von CHIRP (GNU GPL), und mit simulierten Funkgeräten Byte für Byte mit CHIRP verglichen. Echte Yaesu- und Icom-Funkgeräte wurden noch nicht getestet: Sicherungskopie aufbewahren und jedes Problem melden.
- Die Übersetzungen ins Malaiische und ins Hindi sind neu: Melden Sie unklare Texte.

## IU0VYT Flash 4.13.1

### Neu
- **IJV-Firmware** für Quansheng UV-K5, UV-K5 V3 und UV-K1 (inoffizielle Firmware von IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): fünf Registerkarten, eine pro Firmware-Version.
  - **V3** mit 200 oder 999 Kanälen auf dem klassischen UV-K5 (Hardware V1).
  - **V4**, **V5 («FW5»)** und **V6** auf UV-K5 V3 und UV-K1: die 500 Kanäle der am Funkgerät gewählten Bank, mit den Namen der Gruppen und Bänke.
  - Gelesen und geschrieben werden Kanäle und wichtigste Einstellungen (Menü, Anzeige, VOX, Seitentasten). Vor dem Schreiben prüft das Programm, dass am Funkgerät noch dieselbe Bank gewählt ist, und schreibt nur die geänderten Teile; die Kalibrierungen bleiben immer unberührt.
  - **V6 (neu in 4.13.1)**: Datenverbindung der V6, Schreiben in Sektoren zu 4 KiB, jeder erneut gelesen und geprüft. Meldet das Funkgerät, dass die Datenverbindung fehlt, aktualisieren Sie es mit FLASH V6 von RadioManager.
- **Firmware-Aktualisierung**: Eine .bin-Datei mit einer IJV-Version wird gemeldet, mit der für die Kanäle zu verwendenden Registerkarte und einer Warnung für V5 und V6.
- Übersetzungen in allen 11 Sprachen.

### Hinweise
- Protokoll und Speicher aus den von IJV veröffentlichten CHIRP-Modulen und aus der Analyse von RadioManager 4.3.2 abgeleitet; Code von Grund auf neu geschrieben, mit simulierten Funkgeräten getestet. Echte Funkgeräte mit IJV wurden noch nicht getestet: Melden Sie jedes Problem.
- Module, Einträge, Lizenz und Firmware von IJV bleiben bei RadioManager von IJV.

## IU0VYT Flash 4.13.0

### Neu
- **IJV-Firmware** für Quansheng UV-K5, UV-K5 V3 und UV-K1 (inoffizielle Firmware von IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): vier neue Registerkarten, eine pro Firmware-Version.
  - **V3** mit 200 oder 999 Kanälen auf dem klassischen UV-K5 (Hardware V1).
  - **V4** und **V5 («FW5»)** auf UV-K5 V3 und UV-K1: die 500 Kanäle der am Funkgerät gewählten Bank, mit den Namen der Gruppen und Bänke.
  - Gelesen und geschrieben werden Kanäle und wichtigste Einstellungen (Menü, Anzeige, VOX, Seitentasten). Vor dem Schreiben prüft das Programm, dass am Funkgerät noch dieselbe Bank gewählt ist, und schreibt nur die geänderten Blöcke; die Kalibrierungen bleiben immer unberührt.
  - Die **V6** wird erkannt: Kanäle und Einstellungen der V6 werden mit RadioManager von IJV verwaltet.
- **Firmware-Aktualisierung**: Eine .bin-Datei mit einer IJV-Version wird gemeldet, mit der für die Kanäle zu verwendenden Registerkarte und einer Warnung für V5 und V6.
- Übersetzungen in allen 11 Sprachen.

### Hinweise
- Protokoll und Speicher aus den von IJV veröffentlichten CHIRP-Modulen und aus der Analyse von RadioManager 4.3.2 abgeleitet; Code von Grund auf neu geschrieben, mit simulierten Funkgeräten getestet. Echte Funkgeräte mit IJV wurden noch nicht getestet: Melden Sie jedes Problem.
