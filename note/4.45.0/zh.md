# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="es.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="pt.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="ru.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub><b>中文</b></sub></td>
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub 上的发布页面（英文），含下载：[V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0，适用于 Microsoft Windows 10/11

### 新内容
- **AnyTone 系列 DMR 电台**：AnyTone AT-D868UV、AT-D878UV / AT-D878UV Plus、AT-D878UV II Plus，以及 BTECH DMR-6X2 和 DMR-6X2 PRO（这两款以前是“不兼容”）。通过电台的 USB 数据线管理信道和区域；存储器和协议来自开源程序 dmrconfig，并已用其原始代码校验。只写入信道、区域和扫描列表。
- 使用原厂固件的 **TYT MD-UV380、MD-UV390、MD-2017、MD-9600 以及 Retevis RT3S 和 RT82**：存储器和协议与 Baofeng DM-1701 相同。
- **Baofeng BF-666S**（使用 BF-888S 的驱动）和 **BF-T6**（实验性，使用 BF-T8 的驱动）。
- 对仍为“不兼容”的 UV-8D、BF-UV8、DM-1702 和 DM-1703 卡片，现在写明了原因。
- 型号库：54 个品牌的 673 个型号（623 个支持信道，556 个支持设置）。

### 说明
- 所有新电台都尚未在真实电台上测试：请先只做读取。
- BTECH DMR-6X2 PRO 和 Baofeng BF-T6 为实验性支持。
- 本程序为独立程序，与各电台厂家无关联。
- 完整测试全部通过：通用部分；918 项结构对比和 381 次与 CHIRP 完全一致的读取 / 写入；AnyTone 系列已与 dmrconfig 的原始代码对比。

---

本版本还包含 4.24.0 至 4.44.0 各版本的新内容，这些版本没有单独发布稳定版。以下从最新的开始列出。

## IU0VYT Flash 4.44.0

### 新内容
- 为另外八款原本只支持信道的电台增加**设置**：Radtel RT-480、RT-68、RT-5D、RT-69 GPS、RT-4D、RT-6D、RT-1000 Pro 和 Bajeton DM-5000，依据厂家程序。
- **RT-480 和 RT-68**：写入时，常规设置数据块现在也会写入电台。
- **RT-4D、RT-6D、RT-1000 Pro**：电台文件中也包含设置（以前保存的文件仍然有效）。

### 说明
- 所有新设置都尚未在真实电台上测试。
- RT-69 GPS：目前只包含在厂家程序上逐项核实过的项目。

## IU0VYT Flash 4.43.0

### 新内容
- **已修正错误**：对 33 款电台（包括 Baofeng UV-17 Pro、UV-17 Pro GPS、UV-21 Pro、UV-25、UV-32、UV-5RH、K5 Plus，Retevis RT22 和 HA1G，Kenwood TK-2140 / TK-3140）以及 4.42 的 8 款“孪生”电台，通过数据线读取后不显示设置，只有打开 CHIRP 文件时才显示。
- 为另外五款 Radtel 增加**设置**：RT-817A / RT-871A、RT-830、RT-850、RT-330 和 RT-400，依据厂家程序。

### 说明
- 所有新设置都尚未在真实电台上测试。
- Radtel RT-830 和 RT-850：每款电台有一个项目，厂家程序读取的方式与写入的方式不同；这里以写入为准，并对该项目给出警告。

## IU0VYT Flash 4.42.0

### 新内容
- **首页标签页**：现在可以完整显示在窗口中。品牌框更宽，页面会自动适应窗口大小。
- 为 30 个原本只支持信道的型号增加**设置**：Baofeng D 系列（BF-V1D … BF-V12D、888SD、889D、1901D、1904D、1909D、W31D、BF-T20D）、BF-88ST、BF-C2 / BF-V2、GT-68、GT-22、T22、UV-17R、UV-21R、UV-21L、UV-25 Plus、UV-28 Pro、UV-29 Pro、UV-7B、AR-152 Pro 和 Radtel RT-13。

### 说明
- 所有新设置都尚未在真实电台上测试。
- Baofeng D 系列：在厂家程序中，“语音提示”和“按键音”在读取和写入时是对调的；这里以写入为准，并对这两项给出警告。

## IU0VYT Flash 4.41.0

### 新内容
- **新电台**（CHIRP 中原先缺少的手台）：Yaesu VX-1R（信道和设置），Icom IC-91AD 和 IC-92AD（双频段，也支持 D-STAR），BTECH UV-PRO（通过蓝牙，无需数据线），以及 AOR AR8200 接收机（新品牌）。
- 为原本只支持信道的电台增加**设置**：Yaesu FT-7800 / FT-7900、Icom IC-Q7A、TIDRADIO TD-M11。

### 说明
- 所有新电台都尚未实机测试：请先只做读取。
- BTECH UV-PRO：只允许在 CHIRP 驱动作者测试过的固件版本（147）上写入。

## IU0VYT Flash 4.40.0

### 新内容
- **标签页已打开时，自动搜索更快**：快速检查（以前只用于 Baofeng K67 Pro）现在也适用于 Icom，Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2，Quansheng UV-K5 及其衍生型号，Bajeton BJ-9000、BJ-7800、BJ-7700 和 BJ-8300，Radtel RT-900 / RT-950 / RT-470，以及存储器相同的电台。其他电台仍为完整读取。
- **新品牌：Alinco。** 手台 DJ-596、DJ-175、DJ-G7EG 和 DJ-G7T；车载台 DR-135T、DR-235T、DR-435T、DR-03T 和 DR-06T。只支持信道。对不回送数据的数据线也能使用。

### 说明
- 所有 Alinco 都尚未实机测试：请先只做读取。
- 快速检查已用模拟电台测试；如果检查后电台不能恢复正常工作，请报告。

## IU0VYT Flash 4.39.0

### 新内容
- **四款新的 Icom 电台**：ID-80H（D-STAR 手台）、ID-880H（D-STAR 车载台）、IC-W32A 和 IC-W32E（双频手台，分为两个独立部分：0 到 99 为 VHF 信道，100 到 199 为 UHF 信道）。只支持信道。

### 说明
- 所有新电台都尚未实机测试：请先只做读取。

## IU0VYT Flash 4.38.0

### 修正
- **识别**：Bajeton BJ-9000 有时会卡住，尤其是第一次识别时。程序恢复为在第一次握手前先监听端口一秒钟，与 4.35 相同，对已在该数据线上识别过的电台也是如此。4.36 的其他改进保留。

### 说明
- 本版本没有新增电台。

## IU0VYT Flash 4.37.0

### 新内容
- **八款新的 Yaesu 手台**：VX-5R 和 VX-7R（信道）；VX-2R、FT-50R、VX-150、VX-8R、VX-8DR 和 VX-8GE（信道和设置；VX-8 还支持 APRS）。
- 从电台端开始复制（“CLONE”模式）：需手动选择标签页，操作说明在数据线窗口中。

### 说明
- 所有新电台都尚未实机测试：请先只做读取。
- 未实现：Yaesu VX-1。

## IU0VYT Flash 4.36.0

### 修正
- **“识别”和自动识别**：程序会记住每台电台是在哪根数据线上识别的，并优先从这些开始；第一次搜索不再浪费时间在监听和准备上；接好数据线后才开机的电台会更频繁地重新尝试。
- “正在识别电台”窗口只在读取电台时出现，并显示型号和进度。如果电台已连接但未开机，程序会提示开机，并在开机后立即识别。
- 标签页已打开的 Baofeng K67 Pro：几秒钟内即可识别，无需重新读取整个存储器。

### 新内容
- “识别”标签页：**“保存识别诊断”**按钮（把最近几次搜索的时间和前几个字节保存到文本文件中，识别慢或出错时可发给作者；文件中不含信道和个人数据）。

### 说明
- 本版本没有新增电台。

## IU0VYT Flash 4.35.0

### 修正
- **“识别”和自动识别**：删除了 4.31 的“快速一轮”，它在真实电台上可能让电台卡在编程状态，并使搜索变慢。等待时间恢复为以前的设置。
- 每台电脑上搜索尝试的电台分组都相同（以前每次启动都会变化）。
- 不再把没有电台的数据线的回声误认为电台。

### 说明
- 本版本没有新增电台。

## IU0VYT Flash 4.34.0

### 新内容
- **新电台**：Yaesu VX-170 和 VX-177（手台），FT-7800 / FT-7900 和 FT-8900（车载台），只支持信道，使用 CHIRP 驱动：读取、写入以及与电台交换的字节均与 CHIRP 完全一致。

### 说明
- 从电台端开始复制（“CLONE”模式）：需手动选择标签页，“识别”不会尝试这些电台。

## IU0VYT Flash 4.33.0

### 新内容
- **新电台**：Icom IC-T7H、IC-T8A、IC-Q7A 和 IC-P7（信道），使用 CHIRP 驱动：读取、写入以及与电台交换的字节均与 CHIRP 完全一致。

### 说明
- IC-Q7A 和 IC-P7：模式（FM、宽带 FM、AM）随频率而定；信道的“P”跳过和调谐步进会保留。

## IU0VYT Flash 4.32.0

### 新内容
- **新电台**：FRS 版 Baofeng T22，TIDRADIO TD-M8S、TD-V2、TD-H4 和 TD-H11（信道和设置），依据厂家程序。
- **“识别”**也会提示那些自报为 BF-888S 的新电台，而且不会延长搜索时间：只有打开相应标签页时，才会用它们的驱动读取电台。

### 修正
- 2026 年的 TIDRADIO（TD-H3 Plus / SE、TD-H7 Plus、第 4 代 TD-H8、TD-H9、TD-638UV）：只有真正具备中功率的型号才提供中功率，GMRS 版本一律不提供。
- Baofeng T22 PMR446：连接方式与厂家程序相同。
- Radtel RT-13、RT-480 / RT-68、RT-830 / RT-850、RT-69 GPS 和 TIDRADIO TD-M2 Pro 的说明：“识别”会尝试它们，删除了说法相反的句子。

## IU0VYT Flash 4.31.0

### 新内容
- **“识别”快了很多**：先对所有电台快速走一轮，只有什么都没找到时，才进行以前那样的完整一轮。列表末尾的电台不到半分钟就能找到，以前要将近两分钟。
- **“识别”现在也会尝试最近几个版本新增的电台**：Retevis RB39 / RB639、C6、C66，Radioddity Echo、UB15-G、DB35-G、DB35-UV、DB40-G，以及没有显示屏的 TIDRADIO。

### 修正
- Retevis RB39 / RB639：使用错误国家的标签页时，电台会一直停留在编程状态，直到关机。

## IU0VYT Flash 4.30.0

### 新内容
- TIDRADIO **TD-H3 Plus**、**TD-H3 SE**、**TD-H7 Plus**、**第四代 TD-H8**、**TD-H9** 和 **TD-638UV** 的**常规设置**：可编程按键、静噪、VOX、发射限时、显示、语言、开机信息等。
- **新电台（信道和设置）**：TIDRADIO **TID-712Plus**、**TD-M7**、**TD-M11 Plus**、**TD-BR09**、**TD-777S**、**TD-V900**、**TD-FRS**、**TD-PMR** 和 **TD-M2 Pro**。

### 说明
- 新电台来自厂家程序，尚未在真实电台上测试：请先只做读取。“识别”无法区分它们：请自行选择标签页。

## IU0VYT Flash 4.29.0

### 新内容
- **新电台（只支持信道）**：TIDRADIO **TD-H3 Plus**、**TD-H3 SE**、**TD-H7 Plus**、**第四代 TD-H8**、**TD-H9** 和 **TD-638UV**，每款都有基础版、HAM 版和 GMRS 版。
- TD-H3 Plus 自报为 TD-H3：请使用“TD-H3 Plus”标签页，因为它有三档功率。

### 说明
- 这些电台来自厂家 2026 年的程序（“TIDRadioCPS”），尚未在真实电台上测试：请先只做读取。常规设置暂不支持。

## IU0VYT Flash 4.28.0

### 新内容
- **新电台**（Radioddity 车载台）：**UB15-G**（GMRS 15 W）、**DB35-G**（GMRS 35 W）、**DB35-UV**（VHF / UHF 35 W）和 **DB40-G**（GMRS 40 W，199 个信道），支持信道和常规设置。UB15-G 根据固件分为两个标签页：V1.10 之前（无名称）和 V1.10（有名称）。
- **DB40-G** 在 CHIRP 中也有：这里读写的字节和 `.img` 文件与 CHIRP 相同。

### 修正
- 自 4.26.0 起，测试在 Windows 上无法编译（三个测试文件使用了模拟电台，而模拟电台在 Windows 上不参与编译）。程序本身不受影响。

### 说明
- UB15-G、DB35-G 和 DB35-UV 来自厂家程序及其配置文件，尚未在真实电台上测试：请先只做读取。“识别”无法识别这些电台：需要手动选择标签页。

## IU0VYT Flash 4.27.0

### 新内容
- **新电台**：**Retevis C6**（FRS / GMRS，30 个信道）、**Retevis C66**（UHF，16 个信道）和 **Radioddity Echo-DE / Echo-IT**（PMR446，16 个信道），支持信道和常规设置。
- 写入前，程序会检查电台的频段、型号和地区与要写入的存储器是否一致；电台的密码和校准数据从不写入。

### 修正
- **Retevis RB39 / RB39P**：现在归入“型号”标签页中的“GMRS / FRS（美国）”电台。

### 说明
- 新电台来自厂家程序的代码，尚未在真实电台上测试：请先只做读取。“识别”无法识别它们：需要手动选择标签页。
- 如果电台设置了读取密码，请先用厂家程序取消。

## IU0VYT Flash 4.26.0

### 新内容
- **新电台**：**Retevis RB39 / RB39P**（FRS，22 个信道）和 **RB639 / RB639P**（PMR446，16 个信道），支持信道和常规设置。电台会声明自己是为哪个国家制造的：使用错误的标签页时，读取会停止并指出正确的标签页。
- 为 **Baofeng BF-T17 / BF-T18**（PMR446 和 FRS）、**GT-18 PMR446**、**Radioddity PR-T5**、**FS-T3** 和 **PR-T3** 增加**新设置**：静噪、发射限时、VOX、语音提示、按键音、省电以及厂家程序中的其他设置。

### 修正
- 12 种语言的更新日志：更正了 4.25.0 的日期（2026 年 11 月 12 日）。

### 说明
- 新电台来自厂家程序的代码，尚未在真实电台上测试：请先只做读取。“识别”无法识别它们：需要手动选择标签页。

## IU0VYT Flash 4.25.0

### 新内容
- **使用新固件的 Bajeton BJ-8300**：为 990 个信道、分为 10 个区域、每区 99 个的电台新增标签页（厂家程序 BT-11_CPS v1.2.6）：信道、区域名称和设置。原有的“Bajeton 8300”标签页（960 个信道，15 个区域、每区 64 个）保持不变，现在发现新固件时会**发出警告**。
- **新电台**：**Radioddity GS-10B Pro**（999 个信道和设置），**Radioddity FS-T3** 和 **PR-T3**（信道）。
- 990 信道的 **Bajeton BJ-9000**：新增五项设置（来自 v1.3.3 程序）。

### 修正
- **Baofeng BF-T17 / BF-T18、GT-18 和 Radioddity PR-T5**：高功率和低功率颠倒了。用以前版本写过这些电台的用户，请重新读取并重新选择功率。
- “自动退出菜单”（BJ-8300、BJ-9000、Radtel RT-950）：最后一个选项是 60 秒，不是 55 秒。

### 说明
- 新电台和新标签页来自厂家程序的代码，尚未在真实电台上测试：请先只做读取。
- BJ-8300：要知道该用哪个标签页，请在电台上查看有几个区域（15 个、每区 64 个信道，或 10 个、每区 99 个）。

## IU0VYT Flash 4.24.0

### 新内容
- **“型号”标签页更易读**：电台类型的标签位于页面中央，兼容和不兼容电台之间的分隔线改为红色并加粗。
- **新增 26 款电台**，使用已有的相同电台的驱动管理：
  - 通过对比厂家程序识别：**Bajeton BT-8000**、**Radioddity GC-5**、**Trivox**、**PR-T5**、**GT-82**、**Retevis RT28**；
  - 来自 CHIRP 的等效型号表：Radioddity GA-5S、Retevis RT5（三档功率）、RT5R、RT5RV、RT5，TIDRADIO TD-UV5R TriPower，TechSide TI-F8+ 和 TS-T9+，Rugged RH5R 和 RH5X，Ansoko A-5R 和 A-8S，Arcshell AR-5、AR-6 和 AR-8，Greaval GV-8S 和 GV-9S，Tenway TW-325，UV-82 Pro 和 UV-5R Pro。

### 说明
- 通过对比厂家程序识别的六款电台尚未在真实电台上测试：请先只做读取。如果型号不对，电台不会应答，也不会被改动。
