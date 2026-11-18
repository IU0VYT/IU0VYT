# IU0VYT Flash 4.5.0

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

Release page on GitHub, with the download: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 for Microsoft Windows 10/11

### Retevis RT85 firmware (44 models with firmware update)
- **Retevis RT85** with Retevis .hex files, for example «RT85-Firmware-V1.53.hex». It uses the protocol of the Retevis program «RT85 10W_5W_UpData», which was run against a simulated radio: the blocks and commands sent are identical byte for byte.
- Retevis procedure: radio off, switched on while holding **PTT and P1** (green indicator). Then you choose the radio power (**5 W or 10 W**, as in the Retevis program) and type RT85 to confirm.

### Fixed: Retevis RA89 update
- Version 4.4 expected a wrong answer from the radio to the handshake: with a real RA89 the update did not start (the radio was not touched). Now the expected answer and the pauses are those of the Retevis program.

### Ports, Identify and Convert tabs
- Cable ports are always in ascending order, on every system: COM1, COM2, COM3… COM10. Automatic identification also starts from the first port.
- Identify tab: «Radios identified in the past are recognised faster».
- Convert tab: «Choose a Baofeng or Bajeton file, an .img of any other radio brand (also saved with CHIRP), a CHIRP CSV, Excel tables or CSV».

### Website
- Removed the release-notes text under the download button.
- The title sits higher: the version number under «PRESS START» is fully visible on laptop screens too.

Gallery: 386 models of 27 brands (348 with channels, 302 with settings, 44 with firmware update).
