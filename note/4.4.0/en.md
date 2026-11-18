# IU0VYT Flash 4.4.0

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

Release page on GitHub, with the download: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 for Microsoft Windows 10/11

This version also includes the news of 4.3.0, 4.2.0, 4.1.0 and 4.0.0, which did not have a stable release of their own.

## What's new in 4.4.0

### Fixed: «Radio Identification In Progress» box without a radio
- The box could appear even with a cable without a radio: line noise and the echo of the commands sent looked like a radio answering.
- Now it appears **only for the cable connected to a radio that really answers**. With two cables and one radio (COM6 without a radio, COM7 with the radio, or the other way round) it appears only for the radio's port.
- Without a radio the round of attempts is also faster.

### Top bar, Home, Identify tab
- The button is called **«🔎 Identify Radio»** and is on the right, next to «Updates».
- With many tabs open the button shrinks by itself, down to the magnifying glass alone.
- Home: the descriptions of the boxes from 4.2 are back. At the bottom, copyright, website and source code stay visible, even at 1280×720.
- Identify tab: short and clear explanations.

### Retevis firmware (43 models with firmware update)
- **Retevis RA79 V3** (and UV-K5 V3 / UV-K1, bootloader 7.x) with Retevis .bin files, for example «RA79 V3.bin». The firmware is sent only if the radio bootloader belongs to the same family.
- **Retevis RA89** with Retevis .bin files, for example «Eng_RA89_Upadat_20250524_B91.38.bin». It uses the protocol of the Retevis program «RA89-update-software», which was run against a simulated radio. The commands sent are identical byte for byte.

### Other
- The tests with simulated radios no longer prevent the tests from compiling on Windows.

Gallery: 386 models of 27 brands (348 with channels, 302 with settings, 43 with firmware update).

## What's new in 4.3.0

### Automatic identification: faster, more radios, more ports
- A full round of attempts without a radio takes about **50 seconds** (135 before).
- The program **listens continuously on all cable ports**, each on its own and all together:
  - it also recognises a radio connected while off and switched on later;
  - it recognises **several radios on different cables** at the same time, shows their data and opens all their tabs.
- **Radios already recognised** (e.g. K67 Pro, Bajeton 8300, BJ-9000) are tried first, then all the others.
- If the radio tab is already open, the notice «Radio already identified!! Radio tab already open» appears (with several radios «Radios already identified!! Radio tabs already open»).
- Identify tab: ports in order (COM6, COM7… COM10) and shown as «COM6 — USB-SERIAL CH340»; «Identify New Radio Automatically» in bold.

### Firmware update: 8 more radios (41 in total)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** with Radtel .kdhx files (bootloader «JJCC Bootloader Pro»). RT-470: firmware 1.x is only for the old board, 2.x only for the new one. The program says so and asks for written confirmation.
- **TIDRADIO TD-H3, TD-H8 and Radtel RT-730** with the manufacturer's .BIN files: start the update and switch the radio on while holding the PTT.

### Other
- **Bajeton BJ-9000**: like the BJ-8300, all 960 channels in a single list without zones («Channels: all (960)»).
- More compact **Home**: at the bottom, version, copyright, website and source code are always visible.
- **Gallery**: a thick line separates the compatible models from the incompatible ones.
- **LEGGIMI**: removed a long part repeated by mistake in 4.2.0.

Gallery: 386 models of 27 brands (348 with channels, 302 with settings, 41 with firmware update).

## What's new in 4.2.0

- Fixed automatic identification: the «Radio Identification In Progress» box appears only when a connected radio answers (no longer at start-up with just the cable) and disappears when the radio is recognised or not found. After recognition the radio data and tab open and the search stops, without disabling the feature.
- New «🔎 Identify New Radio Automatically» button in the top bar, between «Info» and «Updates»: it looks for a new radio with the same rules and is grey if automatic identification is disabled.
- 35 new radios (36 versions) and 8 more brands: TYT TH-UV88 family (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), Radtel RT-490 family (with JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 radios of the Retevis RT21 family (with Abbree AR-63) and Quansheng TK11.
- Firmware update also for Radtel RT-590 / RT-590 Plus and RT-600 / RT-600 Plus with Radtel .bin files.
- Fixed some hidden defects in reading channel names with accented or Chinese letters.
- Gallery of 386 models of 27 brands: 348 with channels and repeaters, 302 with settings, 33 with firmware update.

## What's new in 4.1.0

- Fixed automatic identification: it works in any tab and, when you disconnect the radio, switch it off or change it, the search restarts by itself and recognises the radio connected next (even if it is the same one or you switch it on after the cable). Before, after the first recognition, it no longer looked for other radios.
- During the automatic search the box «Radio Identification In Progress. Please wait.» appears, with the «Stop» button.
- 14 new radios and 3 more brands: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 and RT-730, Hiroyasu HI-8811, Talkpod A36plus and A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2nd and 3rd generation) and TD-H3, in the base, HAM and GMRS versions.
- Firmware update of the Quansheng UV-K5, UV-K5(8), UV-K6 and UV-5R Plus with Quansheng .bin files, with a check that the firmware is for the type of radio connected.
- Home screen and website: «CHANNEL MANAGEMENT - FREQUENCIES - REPEATERS - FIRMWARE» and the names of all brands in an 80s video-game style frame.
- Gallery of 349 models of 19 brands: 311 with channels and repeaters, 265 with settings, 31 with firmware update.

## What's new in 4.0.0

- Multi-brand program: 335 models of 16 brands in the gallery, 297 with channels and frequencies via cable and 251 with the radio settings.
- New radios: Quansheng UV-K5, UV-K6, UV-5R Plus and UV-K1; 31 Retevis models (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 and T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S and GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton and TID.
- Automatic identification («Identify» tab): connect the cable and the radio is recognised by itself, with its tab open and the channels read. The choice is remembered and is switched off with the same button.
- Model gallery with the brand menu and a colour for each brand.
- «Read from radio» and «Write to radio» in blue in the radio tabs.
- Home, Info, guide, website and documentation updated to the multi-brand program.
