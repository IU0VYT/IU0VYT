# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub><b>Deutsch</b></sub></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="../../lingue/ja/README.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="../../lingue/hi/README.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-13-orange" alt="13 languages">
</p>

**Version 4.30.0** · Windows · Oberfläche in 13 Sprachen (italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية, Bahasa Melayu, हिन्दी) · eine einzige ausführbare Datei, keine Installation

Programm für Funkamateure und Nutzer von Funkgeräten von **52 Marken** (**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone, Yaesu, Icom, Kenwood): es liest und schreibt Kanäle, Frequenzen und Einstellungen über das Programmierkabel, verwaltet ein Archiv der italienischen Relais und aktualisiert die Firmware der unterstützten Funkgeräte.

🌐 Website des Programms: **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*Dies ist die deutsche Übersetzung von [README.it.md](../../README.it.md); das italienische Original ist der maßgebliche Text.*

## Funktionen

- **Kanäle und Frequenzen** von 571 Modellen (einschließlich digitaler PMR- und DMR-Funkgeräte, auch per USB): Lesen und Schreiben per Kabel, mit CHIRP verträgliche `.img`-Dateien, Originaldateien des Baofeng K67 Pro (`.xlc`) und der Bajeton 8300 (`.dat`).
- **Einstellungen der Funkgeräte** (Menüs, Display, VOX, Squelch, TOT, DTMF, Tasten…) für 491 Modelle, mit der Beschreibung jedes Eintrags und einer zu bestätigenden Warnung bei heiklen Einträgen.
- **Identifizierung des angeschlossenen Funkgeräts**: das Programm erkennt es selbst und zeigt Marke, Modell, Typ, Firmware und die übrigen Angaben, die es meldet (bei USB-DMR-Funkgeräten auch DMR-ID und Name). Mit der **automatischen Identifizierung** (Schaltfläche in der Registerkarte „Identifizieren“, Wahl gespeichert) genügt es, das Kabel anzuschließen, egal in welcher Registerkarte man ist: das Programm hört an allen Kabelanschlüssen zugleich, erkennt auch ein erst nach dem Anschließen eingeschaltetes Funkgerät und mehrere an verschiedenen Kabeln angeschlossene Funkgeräte, probiert zuerst die früher schon erkannten und die verbreitetsten, lässt sich von Funkgeräten im Firmware-Aktualisierungsmodus nicht täuschen und öffnet für jedes die Registerkarte mit den gelesenen Kanälen; das Feld „Funkgerät wird identifiziert“ erscheint nur für das Kabel mit einem antwortenden Funkgerät; für ein weiteres Funkgerät gibt es oben rechts die Schaltfläche **„🔎 Funkgerät identifizieren“**.
- **Weitergabe an ein anderes Funkgerät**: die in einem Funkgerät abgehakten Kanäle werden direkt in ein anderes angeschlossenes Funkgerät geschrieben, gewählt nach Marke und Modell, mit Verträglichkeitsprüfung, Anpassung, Sicherungskopie und Prüfung.
- **Universeller Konverter**: die Kanäle jedes Funkgeräts oder jeder Datei (K67 `.xlc`, Bajeton `.dat`, `.img`-Abbilder auch von CHIRP, CHIRP-CSV, Excel/CSV) im Format jedes anderen Funkgeräts, an dessen Regeln angepasst.
- **Oberfläche in 13 Sprachen** mit der Flaggenauswahl neben „ℹ Info“: sofortige Übersetzung des ganzen Programms (Arabisch von rechts nach links), Sprache beim Neustart gespeichert; Informationsdateien übersetzt im Ordner `lingue`.
- **Zonennamen** statt „Zone 1“, „Zone 2“… (Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000 und BJ-7700, Hiroyasu HI-UV850 Pro; Zonen der DMR-Funkgeräte RD-5R, DM-1801 und DM-32UV), per Kabel zusammen mit den Kanälen gelesen und geschrieben.
- **Sammelbearbeitung** der angehakten Kanäle: Leistung, Bandbreite, Suchlauf, Subtöne, Ablage oder Senden gemeinsam, der Rest jedes Kanals bleibt, wie er ist.
- **Galerie mit 625 Modellen** von 52 Marken, gegliedert in unterstützte und nicht unterstützte, nach Typ (Amateurfunk, DMR, GMRS/FRS, PMR446, für Kinder) und nach Marke, mit Suche, Filtern und Links zu den offiziellen Programmen und Handbüchern.
- **Relaisarchiv** auf Grundlage der aktuellen Liste von [IK2ANE](https://www.ik2ane.it), mit Filtern und Übernahme in die Funkgeräte, mit Verträglichkeitsprüfung zwischen Relais und Funkgerät vor dem Schreiben.
- **Firmware-Aktualisierung** von 57 Modellen (auch der Quansheng UV-K5, UV-K6 und UV-5R Plus, der Radtel RT-590 und RT-600 mit den .bin-Dateien von Quansheng und Radtel, der Radtel RT-470, RT-470L, RT-470X, RT-495 und RT-630 mit den .kdhx-Dateien von Radtel, der TIDRADIO TD-H3, TD-H8 und Radtel RT-730 mit den .BIN-Dateien des Herstellers, der Retevis RA79 V3 und RA89 mit .bin-Dateien und der Retevis RT85 mit den .hex-Dateien von Retevis, der Radtel RT-950 Pro und RT-950 und der Bajeton BJ-9000 mit .BTF-Dateien, der Radtel RT-920 und RT-910B mit .anc-Dateien, der Radtel RT-910 und RT-900 und der Bajeton BJ-7700 ohne Bluetooth mit .BTU-Dateien, der Radtel RT-890 mit den .bin-Dateien von Radtel und M7OCM, der Baofeng DM-1701 per USB mit den .bin-Dateien von Baofeng, der Baofeng RD-5R und DM-1801 per USB mit .sgl-Dateien), mit Prüfung der Datei und Sicherungskopie des Funkgeräts; geführte Aktualisierung mit dem offiziellen Programm für BTECH DMR-6X2 und DMR-6X2 PRO.
- Automatische **Sicherungskopie** vor jedem Schreiben und Prüfung danach.
- **Prüfung auf Aktualisierungen** des Programms bei jedem Start.

## Verwendung

1. [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (neueste Version) herunterladen und das Archiv entpacken.
2. `IU0VYT Flash.exe` starten (Windows 10/11; verwendet die WebView2-Komponente von Windows).
3. In der Registerkarte **📻 Modelle** das eigene Funkgerät wählen, das Kabel anschließen und **Aus Funkgerät lesen** verwenden.

Die vollständige Anleitung steht in der Datei `LIESMICH.txt` (italienisches Original: `LEGGIMI.txt`).

## Kompilieren

Benötigt wird Go 1.22 oder neuer. Im Ordner `sorgenti`:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

Die Bibliotheken sind im Ordner `vendor` enthalten.

## Hinweise

- Die Protokolle der Funkgeräte wurden für die Interoperabilität rekonstruiert und mit simulierten Funkgeräten überprüft: **bewahren Sie immer die Sicherungskopien auf**. Das Programm wird ohne jede Gewährleistung bereitgestellt.
- Unabhängiges Programm, **weder verbunden noch genehmigt** von den Herstellern der Funkgeräte (Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO und die übrigen). Die genannten Marken gehören ihren jeweiligen Inhabern und dienen nur zur Angabe der unterstützten Funkgeräte.
- Beachten Sie die Vorschriften Ihres Landes zur Nutzung der Frequenzen.

## Dank

- [CHIRP](https://chirpmyradio.com) und seinen Autoren: ein großer Teil der Treiber der Funkgeräte stammt aus ihrer Arbeit.
- [IK2ANE](https://www.ik2ane.it), der die Liste der italienischen Relais allen kostenlos zur Verfügung stellt.

Autoren und Lizenzen der Bestandteile: `AUTOREN_UND_LIZENZEN.txt`.

## Lizenz

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Freie Software unter der **GNU General Public License Version 3 oder neuer** (`GPL-3.0-or-later`): siehe `LICENSE.txt` und `LIZENZHINWEIS.txt`.
