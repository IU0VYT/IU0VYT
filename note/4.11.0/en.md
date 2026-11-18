# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub><b>English</b></sub></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
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

Release page on GitHub, with the download: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 for Microsoft Windows 10/11

### Zone names
- Instead of «Zone 1», «Zone 2»… each zone can have a name, as in the manufacturer's programs: the «✎ Zone names…» button or a double click on the zone. The zone bar and the bottom line show «number · name».
- Radios: Radtel RT-950 Pro, Bajeton BJ-9000 and Hiroyasu HI-UV850 Pro (firmware up to 0.23 and from 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 without Bluetooth; zones of the Baofeng DMR radios RD-5R, DM-1801 and DM-32UV.
- The names are read and written via cable together with the channels and saved in the .img file. Length and characters as in the manufacturer's programs (12 characters, Chinese too, on the RT-950 Pro; 10 on the RT-920; 16 on the DMR radios).
- Safety: the names area is written only if the names change, after reading it back from the radio; if the radio holds data there that are not names, nothing is written.

### Bulk edit
- «✎ Bulk edit…» (also from the right-click menu) changes power, bandwidth, scan, TX tone, RX tone, shift and transmission of the checked (or selected) channels together. The items left on «do not change» stay as they are; Ctrl+Z undoes everything.

### Two more languages (11 in total)
- Korean, with the flag of South Korea.
- Arabic, with the flag of the United Arab Emirates and the page running right to left (frequencies and numbers stay left to right).
- All the information files in the `lingue` folder are translated into Korean and Arabic too and updated to 4.11 in all languages.

### Other
- Website: the available languages are listed; the list of firmware files was removed.
- `go test` passes on Windows (i18n_test.go too); a new test checks that LEGGIMI.txt, AUTORI_E_LICENZE.txt and the licence are identical in the main folder and in «sorgenti».

The new formats are derived from the code of the manufacturer's programs and from the CHIRP driver and tested with simulated radios; they have not been tested with real radios: always keep the backups.

Gallery: 489 models of 44 brands (451 with channels, 396 with settings, 57 with firmware update).
