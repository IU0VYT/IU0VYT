# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 für Microsoft Windows 10/11

Diese Version enthält auch die Neuerungen von 4.3.0, 4.2.0, 4.1.0 und 4.0.0, die keine eigene stabile Veröffentlichung hatten.

## Neu in 4.4.0

### Korrigiert: Fenster «Identifizierung des Funkgeräts läuft» ohne Funkgerät
- Das Fenster konnte auch bei einem Kabel ohne Funkgerät erscheinen: Leitungsstörungen und das Echo der gesendeten Befehle sahen aus wie ein antwortendes Funkgerät.
- Jetzt erscheint es **nur für das Kabel, an dem ein Funkgerät hängt, das wirklich antwortet**. Bei zwei Kabeln und einem Funkgerät (COM6 ohne Funkgerät, COM7 mit Funkgerät, oder umgekehrt) erscheint es nur für den Anschluss des Funkgeräts.
- Ohne Funkgerät ist die Runde der Versuche außerdem schneller.

### Obere Leiste, Start, Registerkarte Identifizieren
- Die Schaltfläche heißt **«🔎 Funkgerät identifizieren»** und steht rechts neben «Aktualisierungen».
- Bei vielen offenen Registerkarten wird die Schaltfläche von selbst schmaler, bis nur noch die Lupe bleibt.
- Start: Die Beschreibungen der Kästen aus 4.2 sind zurück. Unten bleiben Copyright, Website und Quellcode sichtbar, auch bei 1280×720.
- Registerkarte Identifizieren: kurze, klare Erklärungen.

### Retevis-Firmware (43 Modelle mit Firmware-Aktualisierung)
- **Retevis RA79 V3** (und UV-K5 V3 / UV-K1, Bootloader 7.x) mit den .bin-Dateien von Retevis, zum Beispiel «RA79 V3.bin». Die Firmware wird nur gesendet, wenn der Bootloader des Funkgeräts zur selben Familie gehört.
- **Retevis RA89** mit den .bin-Dateien von Retevis, zum Beispiel «Eng_RA89_Upadat_20250524_B91.38.bin». Verwendet das Protokoll des Retevis-Programms «RA89-update-software», das mit einem simulierten Funkgerät ausgeführt wurde. Die gesendeten Befehle sind Byte für Byte identisch.

### Sonstiges
- Die Tests mit simulierten Funkgeräten verhindern nicht mehr das Kompilieren der Tests unter Windows.

Galerie: 386 Modelle von 27 Marken (348 mit Kanälen, 302 mit Einstellungen, 43 mit Firmware-Aktualisierung).

## Neu in 4.3.0

### Automatische Identifizierung: schneller, mehr Funkgeräte, mehr Anschlüsse
- Eine vollständige Runde von Versuchen ohne Funkgerät dauert etwa **50 Sekunden** (vorher 135).
- Das Programm **hört ständig an allen Kabelanschlüssen**, jeder für sich und alle zusammen:
  - es erkennt auch ein Funkgerät, das ausgeschaltet angeschlossen und erst später eingeschaltet wird;
  - es erkennt **mehrere Funkgeräte an verschiedenen Kabeln** gleichzeitig, zeigt ihre Daten und öffnet alle ihre Registerkarten.
- **Bereits erkannte Funkgeräte** (z. B. K67 Pro, Bajeton 8300, BJ-9000) werden zuerst versucht, danach alle anderen.
- Ist die Registerkarte des Funkgeräts schon offen, erscheint der Hinweis «Funkgerät bereits identifiziert!! Registerkarte bereits geöffnet» (bei mehreren Funkgeräten «Funkgeräte bereits identifiziert!! Registerkarten bereits geöffnet»).
- Registerkarte Identifizieren: Anschlüsse in Reihenfolge (COM6, COM7… COM10) und angezeigt als «COM6 — USB-SERIAL CH340»; «Neues Funkgerät automatisch identifizieren» fett.

### Firmware-Aktualisierung: 8 Funkgeräte mehr (insgesamt 41)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** mit den .kdhx-Dateien von Radtel (Bootloader «JJCC Bootloader Pro»). RT-470: Firmware 1.x ist nur für die alte Platine, 2.x nur für die neue. Das Programm weist darauf hin und verlangt eine schriftliche Bestätigung.
- **TIDRADIO TD-H3, TD-H8 und Radtel RT-730** mit den .BIN-Dateien des Herstellers: Aktualisierung starten und das Funkgerät mit gedrückter PTT-Taste einschalten.

### Sonstiges
- **Bajeton BJ-9000**: wie das BJ-8300 alle 960 Kanäle in einer einzigen Liste ohne Zonen («Kanäle: alle (960)»).
- Kompakterer **Start**: Unten sind Version, Copyright, Website und Quellcode immer sichtbar.
- **Galerie**: Eine dicke Linie trennt die kompatiblen von den nicht kompatiblen Modellen.
- **LEGGIMI**: Ein langer, in 4.2.0 versehentlich wiederholter Teil wurde entfernt.

Galerie: 386 Modelle von 27 Marken (348 mit Kanälen, 302 mit Einstellungen, 41 mit Firmware-Aktualisierung).

## Neu in 4.2.0

- Automatische Identifizierung korrigiert: Das Fenster «Identifizierung des Funkgeräts läuft» erscheint nur, wenn ein angeschlossenes Funkgerät antwortet (nicht mehr beim Start mit dem Kabel allein), und verschwindet, wenn das Funkgerät erkannt oder nicht gefunden wird. Nach der Erkennung öffnen sich Daten und Registerkarte des Funkgeräts und die Suche endet, ohne die Funktion abzuschalten.
- Neue Schaltfläche «🔎 Neues Funkgerät automatisch identifizieren» in der oberen Leiste, zwischen «Info» und «Aktualisierungen»: Sie sucht nach denselben Regeln ein neues Funkgerät und ist grau, wenn die automatische Identifizierung abgeschaltet ist.
- 35 neue Funkgeräte (36 Varianten) und 8 weitere Marken: Familie TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), Familie Radtel RT-490 (mit JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 Funkgeräte der Familie Retevis RT21 (mit Abbree AR-63) und Quansheng TK11.
- Firmware-Aktualisierung auch für Radtel RT-590 / RT-590 Plus und RT-600 / RT-600 Plus mit den .bin-Dateien von Radtel.
- Einige verborgene Fehler beim Lesen von Kanalnamen mit Akzent- oder chinesischen Zeichen korrigiert.
- Galerie mit 386 Modellen von 27 Marken: 348 mit Kanälen und Relais, 302 mit Einstellungen, 33 mit Firmware-Aktualisierung.

## Neu in 4.1.0

- Automatische Identifizierung korrigiert: Sie funktioniert in jeder Registerkarte, und wenn Sie das Funkgerät abziehen, ausschalten oder wechseln, startet die Suche von selbst neu und erkennt das danach angeschlossene Funkgerät (auch wenn es dasselbe ist oder nach dem Kabel eingeschaltet wird). Vorher suchte sie nach der ersten Erkennung keine weiteren Funkgeräte mehr.
- Während der automatischen Suche erscheint das Fenster «Identifizierung des Funkgeräts läuft. Bitte warten.» mit der Schaltfläche «Abbrechen».
- 14 neue Funkgeräte und 3 weitere Marken: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 und RT-730, Hiroyasu HI-8811, Talkpod A36plus und A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2. und 3. Generation) und TD-H3, in den Ausführungen Basis, HAM und GMRS.
- Firmware-Aktualisierung von Quansheng UV-K5, UV-K5(8), UV-K6 und UV-5R Plus mit den .bin-Dateien von Quansheng, mit Prüfung, dass die Firmware zum angeschlossenen Funkgerätetyp passt.
- Startbildschirm und Website: «KANALVERWALTUNG - FREQUENZEN - RELAIS - FIRMWARE» und die Namen aller Marken in einem Rahmen im Stil der Videospiele der 80er.
- Galerie mit 349 Modellen von 19 Marken: 311 mit Kanälen und Relais, 265 mit Einstellungen, 31 mit Firmware-Aktualisierung.

## Neu in 4.0.0

- Programm für viele Marken: 335 Modelle von 16 Marken in der Galerie, 297 mit Kanälen und Frequenzen per Kabel und 251 mit den Einstellungen des Funkgeräts.
- Neue Funkgeräte: Quansheng UV-K5, UV-K6, UV-5R Plus und UV-K1; 31 Retevis-Modelle (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 und T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S und GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton und TID.
- Automatische Identifizierung (Registerkarte «Identifizieren»): Kabel anschließen, und das Funkgerät wird von selbst erkannt, mit geöffneter Registerkarte und gelesenen Kanälen. Die Wahl wird gespeichert und mit derselben Schaltfläche abgeschaltet.
- Modellgalerie mit Markenmenü und einer Farbe pro Marke.
- «Aus Funkgerät lesen» und «In Funkgerät schreiben» in Blau in den Registerkarten der Funkgeräte.
- Start, Info, Anleitung, Website und Dokumentation an das Programm für viele Marken angepasst.
