# IU0VYT Flash 4.46.0

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

Release page on GitHub, with the download: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 for Microsoft Windows 10/11

### Fixes
- **Radtel RT-880G**: no longer shown as a GMRS / FRS radio. It is an amateur VHF/UHF transceiver (with GPS and APRS), like the RT-880 and RT-860; the «Radio sold as GMRS» warning no longer appears before writing.
- **Verification after writing** on the Radtel RT-880G: it failed because the radio, which saves the data and restarts at the end of the write, was contacted too early and too often. Now the program waits 10 seconds without disturbing it and retries every 15 seconds, up to 90 seconds. The same timings apply to the RT-880, RT-860 and RT-817A / RT-871A, which use the same driver.
- Stricter checks on the values received from AR8200, Kenwood and TK-180 (CodeQL findings).

### Notes
- The verification was tested on a real RT-880G; the RT-880, RT-860 and RT-817A / RT-871A have not been tested on real radios. The other radios keep the previous timings.
- No new radios in this version.
- Independent program, not affiliated with the radio manufacturers.
- Full test suite passed: general part; 915 structure comparisons and 380 reads / writes identical to CHIRP.
