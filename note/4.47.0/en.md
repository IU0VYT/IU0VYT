# IU0VYT Flash 4.47.0

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

Release page on GitHub, with the download: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 for Microsoft Windows 10/11

### What's new
- **Retevis, 24 new models derived from the manufacturer's programs**: B3B, B63B, RT22P, RT622P, RB666, RB666P (same channel memory as the new version of the H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (T18 family); RB65, RB66, RB66P (one frequency and one tone per channel). Connection, reading and writing are rewritten from scratch and checked against the factory data files of the programs. The RB66 and RB66P, and the RB669, RT68H and RT668H, answer with the same code: the right tab must be chosen by hand.
- **Settings of the 24 Retevis models**: 241 items in all (squelch, time-out timer, VOX, scan, voice prompt, battery save, beep, side keys…). Only the items that the program of the model shows are offered; the values of the 24 factory files are all allowed choices.
- **Programming password**: on the 24 Retevis models and on the new version of the Retevis H777 it does not stop reading or writing, as in CHIRP (the manufacturer's program stops instead, if the password is not the one typed by the user). It is an item of the settings, in clear: it can be seen, changed and removed. Length and characters are those of the manufacturer's program: 6 characters (3 for RB45, RB645, RB47 and RB66; 4 for RB66P), digits and letters (digits only for RB65 and RB66). For B3H, B63H, RB45, RB645, RB66 and RB666, whose program does not show the password box, a warning says so.
- **Settings for radios that only had channels**: AnyTone family (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 and DMR-6X2 PRO: 27 items, from the memory descriptions of qdmr and dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 and MD-9600, Retevis RT3S and RT82 (the items of the Baofeng DM-1701); Radtel RT-880 and RT-950 Pro with the nicFW880 / nicFW950 firmware (17 items, from the author's documentation); TID TD-UV68 (the items of the TD-H8); Radtel RT-890 with the M7OCM firmware (53 items).
- **Test versions**: a copy with a number higher than the latest published stable version shows a warning bar («Test version (beta testing)…») and, only then, offers the following test version, always downloaded from GitHub. Users of the stable version do not see test versions.
- Gallery: 697 models of 54 brands (647 with channels, 596 with settings).

### Notes
- None of the 24 Retevis radios has been tested on real hardware, neither for channels nor for settings: read only first, then one harmless setting. If the radio answers in an unexpected way the read stops with a message and nothing is written to the radio.
- With a programming password already set the program still tries to read and write, as CHIRP does: this has not been tested on a real radio.
- The RB66 and RB66P, and the RB669, RT68H and RT668H, answer with the same code: the right tab must be chosen by hand.
- Independent program, not affiliated with the radio manufacturers.
- Full test suite passed: general part; 915 structure comparisons and 380 reads / writes identical to CHIRP.
