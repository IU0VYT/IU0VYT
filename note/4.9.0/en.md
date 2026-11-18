# IU0VYT Flash 4.9.0

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

Release page on GitHub, with the download: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 for Microsoft Windows 10/11

This version also includes the news of 4.8.0, 4.7.0 and 4.6.0, published only as test versions.

## What's new in 4.9.0

### Fixed: Identify with Lanchonlh HG-UV98 and Quansheng TG-UV2+
- During «Identify» the port always spoke 8N1. Now it uses even parity (HG-UV98) and two stop bits (TG-UV2+), like the read from the radio tab. The two radios are now recognised.

### Fixed: HG-UV98 parity
- With even parity the Windows port now also checks parity errors. Before, the check was switched on and immediately off again.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus with free firmware
- New tabs for the egzumer, F4HWN and LOSEHU/LTS/UVE firmware.
- F4HWN uses armel's F4HWN driver (outside CHIRP): 8 power levels, transmit lock, 3 scan lists and the firmware settings.
- Each tab reads and writes only its own firmware. After a firmware update the program tells you which tab to use.

### New radios
- Wouxun (20 models): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (USA)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (Radtel firmware and free M7OCM firmware)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

For the radios from CHIRP drivers, channels and settings are verified with the automatic tests against CHIRP. GMRS, FRS and MURS radios are for the United States: in Italy they are not allowed to transmit on those frequencies.

## What's new in 4.8.0

### .BTU firmware update (Radtel RT-910 and RT-900 without Bluetooth, Bajeton BJ-7700 without Bluetooth)
- .BTU firmware is updated with the protocol of the «Radtel Bootload Tools» program. The bootloader is the one of the Baofeng UV-5RM, with the «PROGRAMBT80U» handshake.
- Automatic mode, with the radio switched on normally. Manual mode: radio off, switched on while holding the two side keys, until the display shows «UPDATE».
- The bootloader does not check the model. That is why the program recognises the radio and the hardware version from the file name, shows the warnings and always asks you to type the model.
- In the gallery RT-910 (without Bluetooth, .BTU) and RT-910B (with Bluetooth, .anc) now have two cards. The same goes for BJ-7700 and BJ-7700 without Bluetooth.

### Fixed: Radtel RT-900 without Bluetooth with the 512-channel firmware
- Up to V1.09P the radio has 512 channels; from V1.14P it has 999.
- Now channels 513-999 are written only if you modify them. First the program checks in the radio that the area is empty or holds channels.

### New radios (CHIRP drivers)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (the new version of the H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

For all of them, channels and settings are verified with the automatic tests against CHIRP.

## What's new in 4.7.0

### Fixed: automatic identification stuck or very slow
- A radio switched on in firmware-update mode sends periodic messages by itself: before, it seemed to answer every model and the search lasted minutes. Now the program listens to the port before trying and is not fooled: the search ends in less than a minute and the Identify tab explains what is happening.
- The interface no longer freezes while the searches start; the most common radios are tried first.
- Fixed the false «Radio tab already open» notice for a tab closed earlier.

### Radios handled the same way
- The selection window and the Identify tab list all the possible radios (e.g. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), with the menu to choose and a short hint on where to see the firmware version on the radio.

### Fixed: Radtel RT-950 Pro / Bajeton BJ-9000 with firmware up to 0.23
- The «transmission enabled» bit of the channels is now always written, with both versions (like the Radtel programs up to v1.2.x; v1.3.3 ignores it). With the old firmware, channels modified in the 990-channel tab could fail to transmit. The bit is also restored in channels 1-960 that lacked it.
- Verified the real versions of the Bajeton BJ-9000: firmware V0.22, V0.27 and V0.29 (identical to the Radtel one), program v1.1.1 (960 channels) and v1.3.3 (990 channels, it is the Radtel program). Channels and settings also verified with the code of the Bajeton programs.

### Firmware update for other radios
- Radtel RT-950 Pro, Radtel RT-950 and Hiroyasu HI-UV850 Pro (.BTF files, same bootloader as the BJ-9000).
- Radtel RT-920 and RT-910B (.anc files, same protocol as BJ-7800 and BJ-7700).

### New radios (CHIRP drivers)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); in the gallery also Hiroyasu HI-UV850 Pro.
- Settings also for SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 and Anysecu AC-580.

## What's new in 4.6.0

### Fixed: Radtel RT-950 Pro / Bajeton BJ-9000 with recent firmware
- From firmware **0.24** (December 2025) the radio has **990 channels in 10 zones of 99** (or a single list of 990); up to **0.23** it had **960 in 15 zones of 64** (or a single list of 960). The program now handles both versions.
- The version is recognised from the data read from the radio (the radio does not report it): the read opens the channels in the right tab. If the data are not enough, a note explains how to check the version on the radio.
- Channels 961-990 are written only if they change, after reading them back from the radio; if the radio does not have them (firmware up to 0.23) nothing is written.
- Channels and settings verified with the original code of the Radtel programs v1.2.0 (960 channels) and v1.3.3 (990 channels).

### New: Radtel RT-950 (without «Pro»)
- 960 channels (15 zones of 64) or a single list, general settings; verified with the original code of the «BT-9000_CPS» v1.3.0 program.

### Fixed: exact radio name
- Different radios handled the same way (e.g. **Radtel RT-950 Pro** and **Bajeton BJ-9000**, **Baofeng K67 Pro** and **5RH Pro**, Radtel RT-910 and Bajeton BJ-7700) are now called by their own name in the tab, read, write, identification and in the backup files.
- The first time one of these radios is identified, the program asks which one it is and then remembers it.

### Fixed: unsaved data warning
- After a read, a write or a forward the channels are in the radio: when closing the program the warning appears only if there are real changes.

### New radios (CHIRP drivers)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: the 990 channels in 10 zones of 99, as on the radio.

Gallery: 489 models of 44 brands (451 with channels, 386 with settings, 52 with firmware update). New brands: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
