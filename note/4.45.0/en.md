# IU0VYT Flash 4.45.0

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

Release page on GitHub, with the download: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 for Microsoft Windows 10/11

### What's new
- **AnyTone family DMR radios**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus and BTECH DMR-6X2 and DMR-6X2 PRO (these two were «not compatible»). Channels and zones with the radio's USB cable; memory and protocol from the free program dmrconfig, verified with its original code. Only channels, zones and scan lists are written.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 and Retevis RT3S and RT82** with the original firmware: same memory and same protocol as the Baofeng DM-1701.
- **Baofeng BF-666S** (BF-888S driver) and **BF-T6** (experimental, BF-T8 driver).
- For the cards that stay «not compatible» UV-8D, BF-UV8, DM-1702 and DM-1703 the reason is now written.
- Gallery: 673 models of 54 brands (623 with channels, 556 with settings).

### Notes
- None of the new radios has been tested on a real radio: do a read only first.
- BTECH DMR-6X2 PRO and Baofeng BF-T6 are experimental.
- Independent program, not affiliated with the radio manufacturers.
- Full test suite passed: general part; 918 structure comparisons and 381 reads / writes identical to CHIRP; AnyTone family compared with the original dmrconfig code.

---

This version also includes the news of versions 4.24.0 to 4.44.0, which did not have a stable release of their own. Here they are, from the most recent.

## IU0VYT Flash 4.44.0

### What's new
- **Settings** for eight more radios that had only channels: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro and Bajeton DM-5000, from the manufacturers' programs.
- **RT-480 and RT-68**: when writing, the general settings block now also reaches the radio.
- **RT-4D, RT-6D, RT-1000 Pro**: the radio file also contains the settings (files saved before remain valid).

### Notes
- None of the new settings has been tested on a real radio.
- RT-69 GPS: for now only the items verified one by one on the manufacturer's program.

## IU0VYT Flash 4.43.0

### What's new
- **Bug fixed**: for 33 radios (including Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 and HA1G, Kenwood TK-2140 / TK-3140) and for the 8 «twins» of 4.42 the settings did not appear after a read via cable, only when opening a CHIRP file.
- **Settings** for five more Radtel radios: RT-817A / RT-871A, RT-830, RT-850, RT-330 and RT-400, from the manufacturer's programs.

### Notes
- None of the new settings has been tested on a real radio.
- Radtel RT-830 and RT-850: one item per radio is read by the manufacturer's program differently from how it is written; here the write is followed and the item has a warning.

## IU0VYT Flash 4.42.0

### What's new
- **Home tab**: it now fits entirely in the window. The brands box is wider and the page adapts by itself to the window size.
- **Settings** for 30 models that had only channels: Baofeng D series (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro and Radtel RT-13.

### Notes
- None of the new settings has been tested on a real radio.
- Baofeng D series: in the manufacturer's program «voice prompt» and «key beep» are swapped between read and write; here the write is followed and the two items have a warning.

## IU0VYT Flash 4.41.0

### What's new
- **New radios** (the CHIRP handhelds that were missing): Yaesu VX-1R (channels and settings), Icom IC-91AD and IC-92AD (dual band, D-STAR too), BTECH UV-PRO (via Bluetooth, no cable) and the AOR AR8200 receiver (new brand).
- **Settings** for radios that had only channels: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Notes
- None of the new radios has been tested for real: do a read only first.
- BTECH UV-PRO: writing is allowed only with the firmware version tested by the author of the CHIRP driver (147).

## IU0VYT Flash 4.40.0

### What's new
- **Faster automatic search with the tab already open**: the short check (before only for the Baofeng K67 Pro) now also applies to the Icom radios, the Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, the Quansheng UV-K5 and derivatives, the Bajeton BJ-9000, BJ-7800, BJ-7700 and BJ-8300, the Radtel RT-900 / RT-950 / RT-470 and the radios with the same memory. For the others the full read remains.
- **New brand: Alinco.** Handhelds DJ-596, DJ-175, DJ-G7EG and DJ-G7T; mobiles DR-135T, DR-235T, DR-435T, DR-03T and DR-06T. Channels only. It also works with cables that do not echo the data back.

### Notes
- No Alinco has been tested for real: do a read only first.
- The short check is tested with the simulated radios; if after the check a radio does not return to normal operation, report it.

## IU0VYT Flash 4.39.0

### What's new
- **Four new Icom radios**: ID-80H (D-STAR handheld), ID-880H (D-STAR mobile), IC-W32A and IC-W32E (dual-band handhelds with two separate halves: channels 0 to 99 VHF, 100 to 199 UHF). Channels only.

### Notes
- None of the new radios has been tested for real: do a read only first.

## IU0VYT Flash 4.38.0

### Fixes
- **Identification**: the Bajeton BJ-9000 sometimes froze, especially at the first identification. The program again listens to the port for one second before the first greeting, as in 4.35, also for the radio already recognised on that cable. The other improvements of 4.36 remain.

### Notes
- No new radios in this version.

## IU0VYT Flash 4.37.0

### What's new
- **Eight new Yaesu handhelds**: VX-5R and VX-7R (channels); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR and VX-8GE (channels and settings; APRS too for the VX-8).
- The copy starts from the radio («CLONE» mode): the tab is chosen by hand and the instructions are in the cable window.

### Notes
- None of the new radios has been tested for real: do a read only first.
- Not done: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Fixes
- **«Identify» and automatic identification**: the program remembers on which cable it recognised each radio and starts from those; the first search no longer wastes seconds listening and preparing; a radio switched on after the cable is retried more often.
- The «Radio Identification In Progress» box appears only while the radio is being read, with the model and the progress. If the radio is connected but off, the program tells you to switch it on and recognises it as soon as it is on.
- Baofeng K67 Pro with the tab already open: recognised in a couple of seconds, without reading the whole memory again.

### What's new
- «Identify» tab: **«Save the identification diagnostics»** button (timings and first bytes of the latest searches in a text file, to send to the author if identification is slow or wrong; it contains no channels and no personal data).

### Notes
- No new radios in this version.

## IU0VYT Flash 4.35.0

### Fixes
- **«Identify» and automatic identification**: removed the «quick round» of 4.31, which with real radios could leave the radio stuck in programming mode and make the search slow. The waits are again those of before.
- The search tries the same groups of radios on every PC (before they changed from one start to the next).
- The echo of a cable without a radio is no longer mistaken for a radio.

### Notes
- No new radios in this version.

## IU0VYT Flash 4.34.0

### What's new
- **New radios**: Yaesu VX-170 and VX-177 (handhelds), FT-7800 / FT-7900 and FT-8900 (mobiles), channels only, with the CHIRP drivers: read, write and bytes exchanged with the radio identical to CHIRP.

### Notes
- The copy starts from the radio («CLONE» mode): the tab must be chosen by hand, «Identify» does not try these radios.

## IU0VYT Flash 4.33.0

### What's new
- **New radios**: Icom IC-T7H, IC-T8A, IC-Q7A and IC-P7 (channels), with the CHIRP drivers: read, write and bytes exchanged with the radio identical to CHIRP.

### Notes
- IC-Q7A and IC-P7: the mode (FM, wide FM, AM) follows the frequency; the channel's «P» skip and tuning step are kept.

## IU0VYT Flash 4.32.0

### What's new
- **New radios**: Baofeng T22 in the FRS version, TIDRADIO TD-M8S, TD-V2, TD-H4 and TD-H11 (channels and settings), from the manufacturers' programs.
- **«Identify»** also suggests the new radios that present themselves like the BF-888S, without making the search longer: the radio is read with their driver only when that tab is opened.

### Fixes
- 2026 TIDRADIO (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 4th generation, TD-H9, TD-638UV): the medium power is offered only by the models that really have it, and never in the GMRS versions.
- Baofeng T22 PMR446: connection like the manufacturer's program.
- Notes of Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS and TIDRADIO TD-M2 Pro: «Identify» does try them, the sentence that said otherwise was removed.

## IU0VYT Flash 4.31.0

### What's new
- **«Identify» is much faster**: first it makes a quick round of all the radios and only then, if it has found nothing, the full one as before. A radio at the bottom of the list is found in less than half a minute instead of almost two.
- **«Identify» now also tries the radios added in the latest versions**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G and the TIDRADIO radios without a display.

### Fixes
- Retevis RB39 / RB639: with the tab for the wrong country the radio stayed in programming mode until switched off.

## IU0VYT Flash 4.30.0

### What's new
- **General settings** for TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 4th generation**, **TD-H9** and **TD-638UV**: programmable keys, squelch, VOX, transmit time-out, display, language, power-on messages and more.
- **New radios (channels and settings)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** and **TD-M2 Pro**.

### Notes
- The new radios come from the manufacturer's programs and have not been tested with a real radio yet: do a read only first. «Identify» does not tell them apart: choose the tab yourself.

## IU0VYT Flash 4.29.0

### What's new
- **New radios (channels only)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 4th generation**, **TD-H9** and **TD-638UV**, each in the base, HAM and GMRS versions.
- The TD-H3 Plus presents itself as the TD-H3: use the «TD-H3 Plus» tabs, because it has three power levels.

### Notes
- These radios come from the manufacturer's 2026 program («TIDRadioCPS») and have not been tested with a real radio yet: do a read only first. General settings are not handled yet.

## IU0VYT Flash 4.28.0

### What's new
- **New radios** (Radioddity mobiles): **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) and **DB40-G** (GMRS 40 W, 199 channels), with channels and general settings. The UB15-G has two tabs, depending on the firmware: before V1.10 (without names) and V1.10 (with names).
- The **DB40-G** is also in CHIRP: here it reads and writes the same bytes and the same `.img` files.

### Fixes
- The tests did not compile on Windows since 4.26.0 (three test files used the simulated radios, which are not compiled on Windows). The program was not affected.

### Notes
- UB15-G, DB35-G and DB35-UV come from the manufacturer's program and its configuration file and have not been tested with a real radio yet: do a read only first. «Identify» does not recognise these radios: the tab must be chosen by hand.

## IU0VYT Flash 4.27.0

### What's new
- **New radios**: **Retevis C6** (FRS / GMRS, 30 channels), **Retevis C66** (UHF, 16 channels) and **Radioddity Echo-DE / Echo-IT** (PMR446, 16 channels), with channels and general settings.
- Before writing, the program checks that band, model and area of the radio are those of the memory to be written; the radio's password and calibrations are never written.

### Fixes
- **Retevis RB39 / RB39P**: it is now among the «GMRS / FRS (USA)» radios in the Models tab.

### Notes
- The new radios come from the code of the manufacturers' programs and have not been tested with a real radio yet: do a read only first. «Identify» does not recognise them: the tab must be chosen by hand.
- If a read password is set in the radio, remove it first with the manufacturer's program.

## IU0VYT Flash 4.26.0

### What's new
- **New radios**: **Retevis RB39 / RB39P** (FRS, 22 channels) and **RB639 / RB639P** (PMR446, 16 channels), with channels and general settings. The radio states which country it is made for: with the wrong tab the read stops and points to the right one.
- **New settings** for **Baofeng BF-T17 / BF-T18** (PMR446 and FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** and **PR-T3**: squelch, transmit time-out, VOX, voice prompt, key beep, battery saver and the others of the manufacturer's program.

### Fixes
- Changelogs in the 12 languages: date of 4.25.0 corrected (12/11/2026).

### Notes
- The new radios come from the code of the manufacturers' programs and have not been tested with a real radio yet: do a read only first. «Identify» does not recognise them: the tab must be chosen by hand.

## IU0VYT Flash 4.25.0

### What's new
- **Bajeton BJ-8300 with the new firmware**: new tab for the radios with 990 channels in 10 zones of 99 (manufacturer's program BT-11_CPS v1.2.6): channels, zone names and settings. The usual «Bajeton 8300» tab (960 channels, 15 zones of 64) stays as it is and now **warns** when it finds the new firmware.
- **New radios**: **Radioddity GS-10B Pro** (999 channels and settings), **Radioddity FS-T3** and **PR-T3** (channels).
- **Bajeton BJ-9000** with 990 channels: five more settings (those of program v1.3.3).

### Fixes
- **Baofeng BF-T17 / BF-T18, GT-18 and Radioddity PR-T5**: high and low power were swapped. If you wrote these radios with an earlier version, read them again and choose the power again.
- «Automatic menu exit» (BJ-8300, BJ-9000, Radtel RT-950): the last choice is 60 seconds, not 55.

### Notes
- The new radios and tab come from the code of the manufacturers' programs and have not been tested with a real radio yet: do a read only first.
- BJ-8300: to know which tab to use, look on the radio at how many zones there are (15 of 64 channels, or 10 of 99).

## IU0VYT Flash 4.24.0

### What's new
- **More readable Models tab**: the labels of the radio types are in the centre of the page and the line between compatible and incompatible radios is red and thicker.
- **26 more radios**, handled with the drivers of identical radios already present:
  - recognised by comparing the manufacturers' programs: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - from the CHIRP table of equivalent models: Radioddity GA-5S, Retevis RT5 (three power levels), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ and TS-T9+, Rugged RH5R and RH5X, Ansoko A-5R and A-8S, Arcshell AR-5, AR-6 and AR-8, Greaval GV-8S and GV-9S, Tenway TW-325, UV-82 Pro and UV-5R Pro.

### Notes
- The six radios recognised by comparing the manufacturers' programs have not been tested with a real radio yet: do a read only first. If the model is not the right one the radio does not answer and is not modified.
