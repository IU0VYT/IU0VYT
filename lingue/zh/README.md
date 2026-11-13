# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub><b>中文</b></sub></td>
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

**版本 4.33.0** · Windows · 界面支持 13 种语言（italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語、한국어、العربية、Bahasa Melayu、हिन्दी）· 只有一个可执行文件，无需安装

面向业余无线电爱好者和 **52 个品牌**对讲机用户的程序（**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone, Yaesu, Icom, Kenwood）：通过编程线读取和写入信道、频率和设置，管理意大利中继台档案，并为支持的电台升级固件。

🌐 程序网站：**[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*本文件是 [README.it.md](../../README.it.md) 的中文译本；以意大利文原文为准。*

## 功能

- 579 个型号的**信道和频率**（包括数字 PMR 和 DMR 电台，也支持 USB）：通过数据线读写，兼容 CHIRP 的 `.img` 文件，Baofeng K67 Pro（`.xlc`）和 Bajeton 8300（`.dat`）的原始文件。
- 496 个型号的**电台设置**（菜单、显示、VOX、静噪、TOT、DTMF、按键……），每一项都有说明，敏感项需确认警告。
- **识别连接的电台**：程序会自动识别，并显示品牌、型号、类型、固件以及电台报告的其他信息（USB DMR 电台还有 DMR ID 和名称）。使用**自动识别**（“识别”标签页中的按钮，选择会被记住），无论在哪个标签页，只需连接数据线：程序同时监听所有数据线端口，也能识别连接数据线后才开机的电台以及连接在不同数据线上的多台电台，优先尝试以前识别过的和最常见的电台，不会被处于固件升级模式的电台误导，并为每台电台打开显示已读取信道的标签页；“正在识别电台”框只在有电台响应的数据线上出现；要识别另一台电台，可使用右上角的 **“🔎 识别电台”** 按钮。
- **转发到其他电台**：一台电台中勾选的信道可直接写入另一台已连接的电台（按品牌和型号选择），带兼容性检查、调整、备份和校验。
- **通用转换器**：任何电台或文件（K67 `.xlc`、Bajeton `.dat`、包括 CHIRP 的 `.img` 映像、CHIRP 的 CSV、Excel/CSV）的信道，转换为任何其他电台的格式，并按其规则调整。
- **13 种语言的界面**，旗帜选择器位于“ℹ Info”旁：整个程序立即翻译（阿拉伯语从右到左），重启后记住语言；说明文件的译本位于 `lingue` 文件夹。
- **区域名称**，代替“区域 1”“区域 2”……（Radtel RT-950 Pro、RT-950、RT-920、RT-910，Bajeton BJ-9000 和 BJ-7700，Hiroyasu HI-UV850 Pro；DMR 电台 RD-5R、DM-1801 和 DM-32UV 的区域），与信道一起通过数据线读取和写入。
- **批量修改**已勾选的信道：同时修改功率、带宽、扫描、亚音、频差或发射，每个信道的其余部分保持不变。
- **型号库**收录 52 个品牌的 633 个型号，分为兼容和不兼容，按类型（业余、DMR、GMRS/FRS、PMR446、儿童）和品牌分类，带搜索、筛选以及官方程序和手册的链接。
- **中继台档案**，基于 [IK2ANE](https://www.ik2ane.it) 的最新列表，带筛选和写入电台功能，写入前检查中继台与电台的兼容性。
- 57 个型号的**固件升级**（包括 Quansheng UV-K5、UV-K6 和 UV-5R Plus，使用 Quansheng 和 Radtel .bin 文件的 Radtel RT-590 和 RT-600，使用 Radtel .kdhx 文件的 Radtel RT-470、RT-470L、RT-470X、RT-495 和 RT-630，使用厂家 .BIN 文件的 TIDRADIO TD-H3、TD-H8 和 Radtel RT-730，使用 .bin 文件的 Retevis RA79 V3 和 RA89，以及使用 Retevis .hex 文件的 Retevis RT85，使用 .BTF 文件的 Radtel RT-950 Pro、RT-950 和 Bajeton BJ-9000，使用 .anc 文件的 Radtel RT-920 和 RT-910B，使用 .BTU 文件的 Radtel RT-910、RT-900 和无蓝牙 Bajeton BJ-7700，使用 Radtel 和 M7OCM .bin 文件的 Radtel RT-890，通过 USB 使用 Baofeng .bin 文件的 Baofeng DM-1701，通过 USB 使用 .sgl 文件的 Baofeng RD-5R 和 DM-1801），带文件检查和电台备份；BTECH DMR-6X2 和 DMR-6X2 PRO 可使用官方程序进行引导式升级。
- 每次写入前自动**备份**，写入后校验。
- 每次启动时**检查程序更新**。

## 使用方法

1. 下载 [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip)（最新版本）并解压。
2. 运行 `IU0VYT Flash.exe`（Windows 10/11；使用 Windows 的 WebView2 组件）。
3. 在 **📻 型号** 标签页中选择你的电台，连接数据线，然后使用 **从电台读取**。

完整指南见 `GUIDE.txt` 文件（意大利文原文：`LEGGIMI.txt`）。

## 编译

需要 Go 1.22 或更高版本。在 `sorgenti` 文件夹中：

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

库已包含在 `vendor` 文件夹中。

## 注意事项

- 电台协议是为互操作性而重建的，并用模拟电台验证过：**请务必保留备份**。本程序不提供任何担保。
- 独立程序，与电台厂家（Baofeng、Bajeton、Quansheng、Retevis、Radtel、Radioddity、TIDRADIO 等）**没有关联，也未获其认可**。提及的商标归各自所有者所有，仅用于表示兼容的电台。
- 请遵守你所在国家关于频率使用的规定。

## 致谢

- [CHIRP](https://chirpmyradio.com) 及其作者：大部分电台驱动源自他们的工作。
- [IK2ANE](https://www.ik2ane.it)，他免费向所有人提供意大利中继台列表。

组件的作者和许可证：`AUTHORS_AND_LICENSES.txt`。

## 许可证

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

自由软件，按照 **GNU 通用公共许可证第 3 版或更高版本**（`GPL-3.0-or-later`）发布：见 `LICENSE.txt` 和 `LICENSE_NOTICE.txt`。
