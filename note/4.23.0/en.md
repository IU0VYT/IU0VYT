# IU0VYT Flash 4.23.0

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

Release page on GitHub, with the download: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 for Microsoft Windows 10/11

### What's new
- **Kenwood TK-481 and TK-981** (900 MHz, channels divided into systems) and the mobile **TK-780 and TK-880**: channels and settings, with the CHIRP drivers.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: now also the radio settings.
- **Radtel RT-13**: channels (22, without names), derived from the manufacturer's program.
- **Radtel RT-69 GPS**: channels (999, with 12-character names), derived from the manufacturer's program; before writing, each channel is checked to be within the radio's bands.
- More robust with damaged files: fixed two crashes when reading the tones of the Baofeng UV-B5 / UV-B6 and UV-3R.
- Conversions: a channel that the target radio cannot store is skipped with a warning instead of stopping the conversion (IJV, nicFW, OpenGD77, UV-KX, R3b0rn firmware, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Gallery: 569 models of 47 brands (515 with channels, 433 with settings).

### Notes
- TK-481 / TK-981: only the systems the radio already has are used; shared-resource (trunking) systems are not modified.
- RT-13 and RT-69 GPS: if a password is set on the radio, remove it first with the Radtel program. Derived from the manufacturer's programs and tested only with a simulated radio: do a read only first.
- Independent program, not affiliated with the radio manufacturers.

---

This version also includes the news of versions 4.13.0 to 4.22.0, published only as test versions. Here they are, from the most recent.

## IU0VYT Flash 4.22.0

### What's new
- **Radtel RT-480 and RT-68**: channels and names (256 channels, numbered 0 to 255, with 12-character names), derived from the manufacturer's program. Frequencies, name, power, bandwidth, scan and tones can be changed; the other channel items stay as they are.
- **Radtel RT-830 and RT-850**: channels (128, without names), derived from the manufacturer's programs. Frequencies, power, bandwidth, scan and tones can be changed.
- **Radtel RT-20M**: opens with the Ruyage UV58Plus tab (same memory as the iRadio UV-5118plus).

### Notes
- RT-480 / RT-68: only channels and names are written; the radio's settings, VFO and DTMF are not touched. Frequencies in 1.25 kHz steps; every channel must have a transmit frequency.
- RT-830 / RT-850: settings and VFO go back as they were; the radio password is never written. If a password is set on the radio, remove it first with the Radtel program.
- All derived from the manufacturer's programs and tested only with a simulated radio: do a read only first.

## IU0VYT Flash 4.21.0

### What's new
- **Radtel RT-5D (DMR)**: channels (16 zones of 64, with 12-character names), derived from the manufacturer's program. Frequencies, name, power, bandwidth, scan and tones can be changed; the other channel items stay as they are.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** and the mobile **TK-7180, TK-8180** (E too): channels, with the CHIRP drivers. The list has 250 positions for each of the first 32 zones; only the zones the radio already has are used, 512 channels in total.

### Notes
- RT-5D: the read takes the whole memory and the write sends it all back as it is in the file, as the Radtel program does. Before writing, do a fresh read: an old file would bring back the contacts and settings of that time. Tested only with a simulated radio.
- TK-x180: channels only, no settings; zones are not created or deleted. If the radio has non-conventional (trunking) zones, the read is refused.

## IU0VYT Flash 4.20.0

### What's new
- **Radtel RT-4D and RT-6D (DMR) and RT-1000 Pro**: channels (1024, with 16-character names) and zones (250), derived from the manufacturer's programs. Frequencies, name, power, bandwidth, tones and, on RT-4D and RT-6D, scan can be changed; the other channel items stay as they are. Only channels and zones are written: contacts, groups, messages, FM radio and settings are not touched.

### Notes
- Tested only with a simulated radio: do a read only first and report any problem.
- «Identify» does not tell these radios apart from the other Radtel and Bajeton radios with the same connection: the tab must be chosen by hand.
- RT-1000 Pro: scan exclusion is not handled (it stays as it is on the radio).

## IU0VYT Flash 4.19.0

### What's new
- **Radtel RT-817A / RT-871A**: channels (1024, with 16-character names), derived from the manufacturer's program. Same connection and same channel format as the RT-880G, without zones. Only the areas used by the Radtel program are read and written (settings, channels, FM radio); the logo is not touched.

### Notes
- Tested only with a simulated radio: do a read only first and report any problem.
- «Identify» does not tell the RT-817A / RT-871A apart from the RT-880G: the tab must be chosen by hand.

## IU0VYT Flash 4.18.0

### What's new
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: professional handhelds with channels divided into zones (groups), channels and settings. The list shows the 250 memory positions; a new channel goes into the zone of the channel before it.
- **Quansheng UV-K5 / UV-K6 with fagci's R3b0rn firmware**: channels (298 with the original memory, up to 1024 with a larger memory). Settings, VFO, presets and SSB correction are never written.
- **Radtel RT-330 and RT-400**: channels, derived from the manufacturer's programs.

### Notes
- The CHIRP drivers for the TK handhelds are declared experimental. Zones are not created or deleted by the program.
- The R3b0rn firmware, the RT-330 and the RT-400 are tested only with a simulated radio: do a read only first and report any problem.
- With the R3b0rn firmware a channel without a name takes the frequency as its name (for that firmware a channel without a name does not exist).
- The RT-400 tab refuses radios that identify themselves as RT-470, RT-630, RT-495 or HI-8811: use it only with the RT-400.

## IU0VYT Flash 4.17.0

### What's new
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: channels read and written one channel at a time, with the radio switched on normally (9600 baud). When writing, only the changed channels are sent.
- **Kenwood TK professional handhelds**: channels and settings.
  - 60 series: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - 60G series: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: use the «Radtel RT-880G» tab (same connection and same channel memory).

### Notes
- The Kenwood TH-F6/F7, TH-G71, TH-D7 and TH-K2 use the commands of the CHIRP drivers and are tested only with a simulated radio: do a read only first and report any problem.
- The CHIRP drivers for the TK handhelds are declared experimental.
- The RT-860 comes from comparison with the manufacturer's program and has not been tested on a real radio.

## IU0VYT Flash 4.16.0

### What's new
- **OpenGD77**: channels of the DMR radios with the free OpenGD77 firmware.
  - Radioddity GD-77 and GD-77S, Baofeng DM-1801 and RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 and MD-2017, Retevis RT3S and RT84.
  - The radio appears as a USB serial port; only the changed channels and zones are written, never settings, contacts, groups or calibration.
- **Quansheng UV-K5, more free firmware**:
  - kamilsss655's **NUNU**: channels and settings.
  - joaquim.org's **UV-KX**: channels and settings; the calibration is never written.
  - The «OSFW» tab also works for original firmware modified with extended bands (uvmod-kitchen, Tunas1337's collection).
- **Radtel RT-880** without GPS with the original firmware: use the «Radtel RT-880G» tab.

### Notes
- OpenGD77 is written from the protocol described in the free program qdmr and tested only with a simulated radio: do a read only first, keep the backup and report any problem.
- NUNU and UV-KX are compared byte for byte with their authors' CHIRP drivers.

## IU0VYT Flash 4.15.0

### What's new
- **D-STAR and C4FM**: channels via cable for digital amateur handhelds, with .img files compatible with CHIRP. The «Mode» column shows «D-STAR» or «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) and IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (the copy starts from the radio: the instructions are in the cable window).
  - **Kenwood** (new brand, 47 in total) TH-D72, TH-D74, TH-D75, with the USB cable.
  - Settings for IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D and TH-D72.
- **Free firmware**:
  - **nicFW 2.x** for the TIDRADIO TD-H3: channels and main settings.
  - **nicFW880 / nicFW950** for Radtel RT-880, RT-880G and RT-950 Pro: channels.
  - **F4HWN Fusion 5.x** for Quansheng UV-K1 and UV-K5 V3: 1024 channels and settings.
  - With nicFW the radio calibration is never written.
- **Kenwood in the gallery**: also the PMR446 and DMR/NXDN handhelds, among the incompatible ones (they can only be programmed with the KPG program reserved for dealers).

### Fixes
- Settings: a text field never used reads as empty instead of «unexpected value»; an unchanged value is not rewritten.
- Puxing PX-888K: removed six settings items that could write a wrong value (BOT/EOT codes and DTMF timings).
- Yaesu VX-3: the space is no longer allowed in DTMF codes (the radio does not have it).

### Notes
- Drivers written from scratch based on those of CHIRP (GNU GPL), on the documentation published by nicsure (nicFW) and on the F4HWN driver, and compared byte for byte using simulated radios. The real radios of this version have not been tested yet: keep the backup and report any problem.
- For Icom, Yaesu and Kenwood there is no firmware update (it is done with the manufacturers' programs).

## IU0VYT Flash 4.14.0

### What's new
- **Yaesu and Icom**: two new brands (46 in total) and 13 more radios, with channels via cable and .img files compatible with CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (SCU-35 cable) and FT-60R/E (copy started from the radio, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (OPC-478 cable).
  - Settings for all except the IC-T70. There is no firmware update for these radios.
- **Two more languages**: Malay (Bahasa Melayu) and Hindi (हिन्दी), 13 in total: program, website, guide and information files, README and security policy.
- **Converter and «Forward to another radio»** towards Yaesu and Icom: tones the radio does not have are removed with a warning and channels it cannot store are skipped with a warning, instead of stopping the whole operation.

### Fixes
- Translations: five bold words in the guides stayed in Italian in all languages; two wrong sentences in Korean and Arabic.
- Wouxun KG-UV8D: internal field names aligned with today's CHIRP (no effect on the channels).

### Notes
- Drivers written from scratch based on those of CHIRP (GNU GPL) and compared byte for byte with CHIRP using simulated radios. Real Yaesu and Icom radios have not been tested yet: keep the backup and report any problem.
- The Malay and Hindi translations are new: report any unclear text.

## IU0VYT Flash 4.13.1

### What's new
- **IJV firmware** for the Quansheng UV-K5, UV-K5 V3 and UV-K1 (unofficial firmware by IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): five tabs, one per firmware version.
  - **V3** with 200 or 999 channels on the classic UV-K5 (hardware V1).
  - **V4**, **V5 («FW5»)** and **V6** on the UV-K5 V3 and the UV-K1: the 500 channels of the bank chosen on the radio, with the names of groups and banks.
  - Channels and main settings (menu, display, VOX, side keys) are read and written. Before writing, the program checks that the same bank is still chosen on the radio and writes only the changed parts; the calibrations are never touched.
  - **V6 (new in 4.13.1)**: V6 data link, writing in 4 KiB sectors, each one read back and checked. If the radio says the data link is missing, update it with FLASH V6 of RadioManager.
- **Firmware update**: a .bin file with an IJV version is flagged, with the tab to use for the channels and a warning for V5 and V6.
- Translations in all 11 languages.

### Notes
- Protocol and memory derived from the CHIRP modules published by IJV and from the analysis of RadioManager 4.3.2; code written from scratch, tested with simulated radios. Real radios with IJV have not been tested yet: report any problem.
- IJV modules, items, licence and firmware remain with IJV's RadioManager.

## IU0VYT Flash 4.13.0

### What's new
- **IJV firmware** for the Quansheng UV-K5, UV-K5 V3 and UV-K1 (unofficial firmware by IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): four new tabs, one per firmware version.
  - **V3** with 200 or 999 channels on the classic UV-K5 (hardware V1).
  - **V4** and **V5 («FW5»)** on the UV-K5 V3 and the UV-K1: the 500 channels of the bank chosen on the radio, with the names of groups and banks.
  - Channels and main settings (menu, display, VOX, side keys) are read and written. Before writing, the program checks that the same bank is still chosen on the radio and writes only the changed blocks; the calibrations are never touched.
  - **V6** is recognised: V6 channels and settings are managed with IJV's RadioManager.
- **Firmware update**: a .bin file with an IJV version is flagged, with the tab to use for the channels and a warning for V5 and V6.
- Translations in all 11 languages.

### Notes
- Protocol and memory derived from the CHIRP modules published by IJV and from the analysis of RadioManager 4.3.2; code written from scratch, tested with simulated radios. Real radios with IJV have not been tested yet: report any problem.
