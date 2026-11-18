# IU0VYT Flash 4.10.0

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

Release page on GitHub, with the download: [V.4.10.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.10.0)

IU0VYT Flash V.4.10.0 for Microsoft Windows 10/11

### Program in 9 languages
- New language selector with flags, just to the right of «ℹ Info»: italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語.
- The translation is immediate, with no restart: tabs, buttons, menus, windows, guide, messages and warnings (including those coming from the radios), gallery cards, settings and the Windows «Open» and «Save» dialogs.
- The last language chosen is remembered at restart.
- The information files are translated in the `lingue` folder (one subfolder per language): full guide, README, CHANGELOG with the news of 4.10, authors and licences, version and a note on the licence. The valid licence remains the English text of the GNU GPL (LICENSE.txt).

### Top bar
- Removed the «Updates» button: the check stays automatic at every start and with «🔄 Check for updates» in the Info tab.
- «🔎 Identify Radio» is now next to «? Guide».

### Firmware update: 5 more radios (57 in total)
- Radtel RT-890: Radtel .bin files and free firmware (M7OCM, OEFW); after the M7OCM firmware the channels are managed with its own tab.
- Baofeng DM-1701 / Retevis RT84 via USB, with Baofeng .bin files: only the program memory is written, the radio bootloader is not touched.
- Baofeng RD-5R and DM-1801 via USB (no driver), with the manufacturer's .sgl files: the radio checks the model before erasing the memory.

### Settings: 10 more radios (396 in total)
- Feidaxin FD-288A, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH and Retevis RT76P (from CHIRP).
- Baofeng RD-5R and DM-1801 (31 items), Baofeng DM-1701 (name, DMR ID, power-on texts).

The protocols are rewritten from scratch and tested with simulated radios and with the manufacturers' real files; they have not been tested with real radios: always keep the backups.

Gallery: 489 models of 44 brands (451 with channels, 396 with settings, 57 with firmware update).
