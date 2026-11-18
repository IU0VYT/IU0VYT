# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Release-Seite auf GitHub (auf Englisch), mit dem Download: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 für Microsoft Windows 10/11

Diese Version enthält auch die Neuerungen von 4.8.0, 4.7.0 und 4.6.0, die nur als Testversionen veröffentlicht wurden.

## Neu in 4.9.0

### Korrigiert: Identifizieren mit Lanchonlh HG-UV98 und Quansheng TG-UV2+
- Beim «Identifizieren» sprach der Anschluss immer 8N1. Jetzt verwendet er gerade Parität (HG-UV98) und zwei Stoppbits (TG-UV2+), wie beim Lesen aus der Registerkarte des Funkgeräts. Die beiden Funkgeräte werden jetzt erkannt.

### Korrigiert: Parität des HG-UV98
- Bei gerader Parität prüft der Windows-Anschluss jetzt auch Paritätsfehler. Vorher wurde die Prüfung eingeschaltet und sofort wieder ausgeschaltet.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus mit freier Firmware
- Neue Registerkarten für die Firmware egzumer, F4HWN und LOSEHU/LTS/UVE.
- F4HWN verwendet den F4HWN-Treiber von armel (außerhalb von CHIRP): 8 Leistungsstufen, Sendesperre, 3 Suchlauflisten und die Einstellungen der Firmware.
- Jede Registerkarte liest und schreibt nur ihre eigene Firmware. Nach der Firmware-Aktualisierung nennt das Programm die zu verwendende Registerkarte.

### Neue Funkgeräte
- Wouxun (20 Modelle): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (USA)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (Radtel-Firmware und freie Firmware M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Bei den Funkgeräten aus CHIRP-Treibern sind Kanäle und Einstellungen mit den automatischen Tests gegen CHIRP geprüft. GMRS-, FRS- und MURS-Funkgeräte sind für die USA bestimmt: In Italien ist das Senden auf diesen Frequenzen nicht erlaubt.

## Neu in 4.8.0

### Firmware-Aktualisierung .BTU (Radtel RT-910 und RT-900 ohne Bluetooth, Bajeton BJ-7700 ohne Bluetooth)
- .BTU-Firmware wird mit dem Protokoll des Programms «Radtel Bootload Tools» aktualisiert. Der Bootloader ist der des Baofeng UV-5RM, mit der Begrüßung «PROGRAMBT80U».
- Automatischer Modus mit normal eingeschaltetem Funkgerät. Manueller Modus: Funkgerät aus, mit gedrückten beiden Seitentasten einschalten, bis die Anzeige «UPDATE» zeigt.
- Der Bootloader prüft das Modell nicht. Deshalb erkennt das Programm Funkgerät und Hardwareversion am Dateinamen, zeigt die Warnungen und verlangt immer, das Modell einzugeben.
- In der Galerie haben RT-910 (ohne Bluetooth, .BTU) und RT-910B (mit Bluetooth, .anc) jetzt zwei Karten. Dasselbe gilt für BJ-7700 und BJ-7700 ohne Bluetooth.

### Korrigiert: Radtel RT-900 ohne Bluetooth mit der Firmware für 512 Kanäle
- Bis V1.09P hat das Funkgerät 512 Kanäle; ab V1.14P hat es 999.
- Jetzt werden die Kanäle 513-999 nur geschrieben, wenn Sie sie ändern. Zuvor prüft das Programm im Funkgerät, dass dieser Bereich leer ist oder Kanäle enthält.

### Neue Funkgeräte (CHIRP-Treiber)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (die neue Version des H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Bei allen sind Kanäle und Einstellungen mit den automatischen Tests gegen CHIRP geprüft.

## Neu in 4.7.0

### Korrigiert: automatische Identifizierung blockiert oder sehr langsam
- Ein im Firmware-Aktualisierungsmodus eingeschaltetes Funkgerät sendet von selbst regelmäßige Nachrichten: Früher schien es auf jedes Modell zu antworten und die Suche dauerte Minuten. Jetzt hört das Programm den Anschluss ab, bevor es probiert, und lässt sich nicht täuschen: Die Suche endet in weniger als einer Minute und die Registerkarte Identifizieren erklärt, was passiert.
- Die Oberfläche friert beim Start der Suchen nicht mehr ein; die verbreitetsten Funkgeräte werden zuerst versucht.
- Den falschen Hinweis «Registerkarte bereits geöffnet» für eine zuvor geschlossene Registerkarte korrigiert.

### Funkgeräte mit gleicher Verwaltung
- Das Auswahlfenster und die Registerkarte Identifizieren listen alle möglichen Funkgeräte auf (z. B. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), mit dem Menü zur Auswahl und einem kurzen Hinweis, wo man am Funkgerät die Firmware-Version sieht.

### Korrigiert: Radtel RT-950 Pro / Bajeton BJ-9000 mit Firmware bis 0.23
- Das Bit «Senden erlaubt» der Kanäle wird jetzt immer geschrieben, bei beiden Versionen (wie die Radtel-Programme bis v1.2.x; v1.3.3 ignoriert es). Mit der alten Firmware konnten in der Registerkarte mit 990 Kanälen geänderte Kanäle nicht senden. Das Bit wird auch in den Kanälen 1-960 gesetzt, denen es fehlte.
- Die echten Versionen des Bajeton BJ-9000 geprüft: Firmware V0.22, V0.27 und V0.29 (identisch mit der von Radtel), Programm v1.1.1 (960 Kanäle) und v1.3.3 (990 Kanäle, das Radtel-Programm). Kanäle und Einstellungen auch mit dem Code der Bajeton-Programme geprüft.

### Firmware-Aktualisierung für weitere Funkgeräte
- Radtel RT-950 Pro, Radtel RT-950 und Hiroyasu HI-UV850 Pro (.BTF-Dateien, gleicher Bootloader wie das BJ-9000).
- Radtel RT-920 und RT-910B (.anc-Dateien, gleiches Protokoll wie BJ-7800 und BJ-7700).

### Neue Funkgeräte (CHIRP-Treiber)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); in der Galerie auch Hiroyasu HI-UV850 Pro.
- Einstellungen auch für SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 und Anysecu AC-580.

## Neu in 4.6.0

### Korrigiert: Radtel RT-950 Pro / Bajeton BJ-9000 mit neuer Firmware
- Ab Firmware **0.24** (Dezember 2025) hat das Funkgerät **990 Kanäle in 10 Zonen zu 99** (oder eine einzige Liste mit 990); bis **0.23** hatte es **960 in 15 Zonen zu 64** (oder eine einzige Liste mit 960). Das Programm verwaltet jetzt beide Versionen.
- Die Version wird an den aus dem Funkgerät gelesenen Daten erkannt (das Funkgerät teilt sie nicht mit): Das Lesen öffnet die Kanäle in der richtigen Registerkarte. Reichen die Daten nicht aus, erklärt ein Hinweis, wie man die Version am Funkgerät prüft.
- Die Kanäle 961-990 werden nur geschrieben, wenn sie sich ändern, nachdem sie erneut aus dem Funkgerät gelesen wurden; hat das Funkgerät sie nicht (Firmware bis 0.23), wird nichts geschrieben.
- Kanäle und Einstellungen mit dem Originalcode der Radtel-Programme v1.2.0 (960 Kanäle) und v1.3.3 (990 Kanäle) geprüft.

### Neu: Radtel RT-950 (ohne «Pro»)
- 960 Kanäle (15 Zonen zu 64) oder eine einzige Liste, allgemeine Einstellungen; geprüft mit dem Originalcode des Programms «BT-9000_CPS» v1.3.0.

### Korrigiert: genauer Name des Funkgeräts
- Verschiedene Funkgeräte mit gleicher Verwaltung (z. B. **Radtel RT-950 Pro** und **Bajeton BJ-9000**, **Baofeng K67 Pro** und **5RH Pro**, Radtel RT-910 und Bajeton BJ-7700) heißen jetzt in Registerkarte, Lesen, Schreiben, Identifizierung und Sicherungsdateien bei ihrem eigenen Namen.
- Bei der ersten Identifizierung eines dieser Funkgeräte fragt das Programm, welches es ist, und merkt es sich.

### Korrigiert: Warnung vor nicht gespeicherten Daten
- Nach einem Lesen, Schreiben oder Weitergeben sind die Kanäle im Funkgerät: Beim Schließen des Programms erscheint die Warnung nur bei echten Änderungen.

### Neue Funkgeräte (CHIRP-Treiber)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: die 990 Kanäle in 10 Zonen zu 99, wie am Funkgerät.

Galerie: 489 Modelle von 44 Marken (451 mit Kanälen, 386 mit Einstellungen, 52 mit Firmware-Aktualisierung). Neue Marken: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
