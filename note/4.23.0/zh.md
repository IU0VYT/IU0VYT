# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0，适用于 Microsoft Windows 10/11

### 新内容
- **Kenwood TK-481 和 TK-981**（900 MHz，信道分为多个系统）以及车载台 **TK-780 和 TK-880**：信道和设置，使用 CHIRP 驱动。
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**：现在也支持电台设置。
- **Radtel RT-13**：信道（22 个，无名称），根据厂家程序得出。
- **Radtel RT-69 GPS**：信道（999 个，名称最多 12 个字符），根据厂家程序得出；写入前会检查每个信道是否在电台的频段内。
- 对损坏文件更稳健：修正了读取 Baofeng UV-B5 / UV-B6 和 UV-3R 亚音时的两处崩溃。
- 转换：目标电台无法存储的信道会被跳过并给出警告，而不再中止转换（IJV、nicFW、OpenGD77、UV-KX、R3b0rn 固件，Radtel RT-480 / RT-68、RT-830 / RT-850）。
- 型号库：47 个品牌的 569 个型号（515 个支持信道，433 个支持设置）。

### 说明
- TK-481 / TK-981：只使用电台中已有的系统；共享资源（集群）系统不做修改。
- RT-13 和 RT-69 GPS：如果电台设置了密码，请先用 Radtel 程序取消。根据厂家程序得出，只在模拟电台上测试过：请先只做读取。
- 本程序为独立程序，与各电台厂家无关联。

---

本版本还包含 4.13.0 至 4.22.0 各版本的新内容，这些版本只作为测试版发布过。以下从最新的开始列出。

## IU0VYT Flash 4.22.0

### 新内容
- **Radtel RT-480 和 RT-68**：信道和名称（256 个信道，编号 0 到 255，名称最多 12 个字符），根据厂家程序得出。可修改频率、名称、功率、带宽、扫描和亚音；信道的其他项目保持不变。
- **Radtel RT-830 和 RT-850**：信道（128 个，无名称），根据厂家程序得出。可修改频率、功率、带宽、扫描和亚音。
- **Radtel RT-20M**：使用 Ruyage UV58Plus 标签页打开（存储器与 iRadio UV-5118plus 相同）。

### 说明
- RT-480 / RT-68：只写入信道和名称；电台的设置、VFO 和 DTMF 不会改动。频率步进为 1.25 kHz；每个信道都必须有发射频率。
- RT-830 / RT-850：设置和 VFO 原样写回；电台密码从不写入。如果电台设置了密码，请先用 Radtel 程序取消。
- 以上均根据厂家程序得出，只在模拟电台上测试过：请先只做读取。

## IU0VYT Flash 4.21.0

### 新内容
- **Radtel RT-5D（DMR）**：信道（16 个区域、每区 64 个，名称最多 12 个字符），根据厂家程序得出。可修改频率、名称、功率、带宽、扫描和亚音；信道的其他项目保持不变。
- **Kenwood TK-2180、TK-3180K、TK-3180K2、TK-3180E** 以及车载台 **TK-7180、TK-8180**（含 E 版）：信道，使用 CHIRP 驱动。列表为前 32 个区域各提供 250 个位置；只使用电台中已有的区域，共 512 个信道。

### 说明
- RT-5D：读取时获取全部存储器，写入时按文件中的内容整体写回，与 Radtel 程序的做法相同。写入前请重新读取一次：旧文件会把当时的联系人和设置写回去。只在模拟电台上测试过。
- TK-x180：只有信道，没有设置；不会新建或删除区域。如果电台有非常规（集群）区域，会拒绝读取。

## IU0VYT Flash 4.20.0

### 新内容
- **Radtel RT-4D 和 RT-6D（DMR）以及 RT-1000 Pro**：信道（1024 个，名称最多 16 个字符）和区域（250 个），根据厂家程序得出。可修改频率、名称、功率、带宽、亚音，RT-4D 和 RT-6D 还可修改扫描；信道的其他项目保持不变。只写入信道和区域：联系人、组、短信、FM 收音机和设置不会改动。

### 说明
- 只在模拟电台上测试过：请先只做读取，并报告任何问题。
- “识别”无法把这些电台与连接方式相同的其他 Radtel 和 Bajeton 电台区分开：需要手动选择标签页。
- RT-1000 Pro：不处理扫描排除（保持电台上的原样）。

## IU0VYT Flash 4.19.0

### 新内容
- **Radtel RT-817A / RT-871A**：信道（1024 个，名称最多 16 个字符），根据厂家程序得出。连接方式和信道格式与 RT-880G 相同，但没有区域。只读写 Radtel 程序使用的区域（设置、信道、FM 收音机）；不会改动开机图标。

### 说明
- 只在模拟电台上测试过：请先只做读取，并报告任何问题。
- “识别”无法区分 RT-817A / RT-871A 和 RT-880G：需要手动选择标签页。

## IU0VYT Flash 4.18.0

### 新内容
- **Kenwood TK-2140、TK-3140、TK-280、TK-380**：信道分为区域（组）的专业手台，支持信道和设置。列表显示存储器的 250 个位置；新信道会进入前一个信道所在的区域。
- **使用 fagci 的 R3b0rn 固件的 Quansheng UV-K5 / UV-K6**：信道（原装存储器为 298 个，更大的存储器最多 1024 个）。设置、VFO、预设和 SSB 校正从不写入。
- **Radtel RT-330 和 RT-400**：信道，根据厂家程序得出。

### 说明
- TK 手台的 CHIRP 驱动被标注为实验性。程序不会新建或删除区域。
- R3b0rn 固件、RT-330 和 RT-400 只在模拟电台上测试过：请先只做读取，并报告任何问题。
- 使用 R3b0rn 固件时，没有名称的信道会以频率作为名称（对该固件来说不存在没有名称的信道）。
- RT-400 标签页会拒绝自报为 RT-470、RT-630、RT-495 或 HI-8811 的电台：只用于 RT-400。

## IU0VYT Flash 4.17.0

### 新内容
- **Kenwood TH-F6A、TH-F7E、TH-G71、TH-D7、TH-K2**：电台正常开机（9600 波特）时逐个读写信道。写入时只发送有改动的信道。
- **Kenwood TK 专业手台**：信道和设置。
  - 60 系列：TK-260、TK-270、TK-272、TK-278、TK-360、TK-370、TK-372、TK-378。
  - 60G 系列：TK-260G、TK-270G、TK-272G、TK-278G、TK-360G、TK-370G、TK-372G、TK-378G、TK-388G。
- **Radtel RT-860**：使用“Radtel RT-880G”标签页（连接方式和信道存储器相同）。

### 说明
- Kenwood TH-F6/F7、TH-G71、TH-D7 和 TH-K2 使用 CHIRP 驱动的命令，只在模拟电台上测试过：请先只做读取，并报告任何问题。
- TK 手台的 CHIRP 驱动被标注为实验性。
- RT-860 来自与厂家程序的对比，尚未在真实电台上测试。

## IU0VYT Flash 4.16.0

### 新内容
- **OpenGD77**：使用开源 OpenGD77 固件的 DMR 电台的信道。
  - Radioddity GD-77 和 GD-77S，Baofeng DM-1801 和 RD-5R。
  - Baofeng DM-1701，TYT MD-UV380、MD-380、MD-9600 和 MD-2017，Retevis RT3S 和 RT84。
  - 电台显示为 USB 串口；只写入有改动的信道和区域，从不写入设置、联系人、组和校准数据。
- **Quansheng UV-K5，更多开源固件**：
  - kamilsss655 的 **NUNU**：信道和设置。
  - joaquim.org 的 **UV-KX**：信道和设置；从不写入校准数据。
  - “OSFW”标签页也适用于扩展了频段的修改版原厂固件（uvmod-kitchen、Tunas1337 的合集）。
- **无 GPS 的 Radtel RT-880**（原厂固件）：使用“Radtel RT-880G”标签页。

### 说明
- OpenGD77 根据开源程序 qdmr 中描述的协议编写，只在模拟电台上测试过：请先只做读取，保留备份，并报告任何问题。
- NUNU 和 UV-KX 已与其作者为 CHIRP 编写的驱动逐字节对比。

## IU0VYT Flash 4.15.0

### 新内容
- **D-STAR 和 C4FM**：通过数据线管理业余数字手台的信道，.img 文件与 CHIRP 兼容。“方式”列显示“D-STAR”或“C4FM”。
  - **Icom** ID-31A/E、ID-51A/E、ID-51 Plus、ID-51 Plus2（D-STAR）以及 IC-E90 / IC-T90。
  - **Yaesu** VX-6、VX-3、FT-70D、FT1D、FT2D、FT3D、FT5D（从电台端开始复制：操作说明在数据线窗口中）。
  - **Kenwood**（新品牌，共 47 个）TH-D72、TH-D74、TH-D75，使用 USB 数据线。
  - IC-E90、VX-6、VX-3、FT-70D、FT1D、FT2D、FT3D、FT5D 和 TH-D72 支持设置。
- **开源固件**：
  - TIDRADIO TD-H3 的 **nicFW 2.x**：信道和主要设置。
  - Radtel RT-880、RT-880G 和 RT-950 Pro 的 **nicFW880 / nicFW950**：信道。
  - Quansheng UV-K1 和 UV-K5 V3 的 **F4HWN Fusion 5.x**：1024 个信道和设置。
  - 使用 nicFW 时从不写入电台的校准数据。
- **型号库中的 Kenwood**：还有 PMR446 和 DMR/NXDN 手台，列在不兼容型号中（只能用仅供经销商使用的 KPG 程序编程）。

### 修正
- 设置：从未使用过的文本字段读出为空，而不是“意外的值”；未改动的值不会重写。
- Puxing PX-888K：删除了六个可能写入错误数据的设置项（BOT/EOT 代码和 DTMF 时间）。
- Yaesu VX-3：DTMF 代码中不再允许空格（电台没有空格）。

### 说明
- 驱动以 CHIRP 的驱动（GNU GPL）、nicsure 发布的文档（nicFW）和 F4HWN 的驱动为基础从零编写，并用模拟电台逐字节对比。本版本涉及的真实电台尚未测试：请保留备份，并报告任何问题。
- Icom、Yaesu 和 Kenwood 不支持固件升级（需使用厂家程序）。

## IU0VYT Flash 4.14.0

### 新内容
- **Yaesu 和 Icom**：两个新品牌（共 46 个）和 13 款新电台，可通过数据线管理信道，.img 文件与 CHIRP 兼容。
  - **Yaesu** FT-4XR、FT-4XE、FT-4VR、FT-65R、FT-65E、FT-25R（SCU-35 数据线）以及 FT-60R/E（从电台端开始复制，“F8 CLONE”）。
  - **Icom** IC-V86、IC-T10、IC-V80、IC-T70、IC-V82、IC-U82（OPC-478 数据线）。
  - 除 IC-T70 外都支持设置。这些电台不支持固件升级。
- **新增两种语言**：马来语（Bahasa Melayu）和印地语（हिन्दी），共 13 种：程序、网站、指南和说明文件、README 以及安全策略。
- **转换器和“转发到其他电台”**支持 Yaesu 和 Icom：电台没有的亚音会被删除并给出警告，电台无法存储的信道会被跳过并给出警告，而不再中止整个操作。

### 修正
- 翻译：指南中有五个粗体词在所有语言中仍是意大利语；韩语和阿拉伯语中各有一句译错。
- Wouxun KG-UV8D：内部字段名称与当前 CHIRP 保持一致（对信道无影响）。

### 说明
- 驱动以 CHIRP 的驱动（GNU GPL）为基础从零编写，并用模拟电台与 CHIRP 逐字节对比。真实的 Yaesu 和 Icom 电台尚未测试：请保留备份，并报告任何问题。
- 马来语和印地语的翻译是新的：请报告不清楚的文字。

## IU0VYT Flash 4.13.1

### 新内容
- 适用于 Quansheng UV-K5、UV-K5 V3 和 UV-K1 的 **IJV 固件**（IU0IJV 的非原厂固件，[ijvradio.com](https://www.ijvradio.com/)）：五个标签页，每个固件版本一个。
  - 经典 UV-K5（硬件 V1）上的 **V3**，200 或 999 个信道。
  - UV-K5 V3 和 UV-K1 上的 **V4**、**V5（“FW5”）**和 **V6**：电台上所选存储组的 500 个信道，带有组名和存储组名称。
  - 读写信道和主要设置（菜单、显示、VOX、侧键）。写入前，程序会检查电台上仍选择着同一个存储组，并且只写入有改动的部分；从不改动校准数据。
  - **V6（4.13.1 新增）**：V6 的数据连接，按 4 KiB 扇区写入，每个扇区都会回读并检查。如果电台提示缺少数据连接，请用 RadioManager 的 FLASH V6 升级。
- **固件升级**：带 IJV 版本的 .bin 文件会被标记，并提示管理信道应使用的标签页，对 V5 和 V6 给出警告。
- 已翻译成全部 11 种语言。

### 说明
- 协议和存储器结构来自 IJV 发布的 CHIRP 模块和对 RadioManager 4.3.2 的分析；代码从零编写，已用模拟电台测试。使用 IJV 固件的真实电台尚未测试：请报告任何问题。
- IJV 的模块、项目、许可证和固件仍归 IJV 的 RadioManager 所有。

## IU0VYT Flash 4.13.0

### 新内容
- 适用于 Quansheng UV-K5、UV-K5 V3 和 UV-K1 的 **IJV 固件**（IU0IJV 的非原厂固件，[ijvradio.com](https://www.ijvradio.com/)）：四个新标签页，每个固件版本一个。
  - 经典 UV-K5（硬件 V1）上的 **V3**，200 或 999 个信道。
  - UV-K5 V3 和 UV-K1 上的 **V4** 和 **V5（“FW5”）**：电台上所选存储组的 500 个信道，带有组名和存储组名称。
  - 读写信道和主要设置（菜单、显示、VOX、侧键）。写入前，程序会检查电台上仍选择着同一个存储组，并且只写入有改动的数据块；从不改动校准数据。
  - 可以识别 **V6**：V6 的信道和设置使用 IJV 的 RadioManager 管理。
- **固件升级**：带 IJV 版本的 .bin 文件会被标记，并提示管理信道应使用的标签页，对 V5 和 V6 给出警告。
- 已翻译成全部 11 种语言。

### 说明
- 协议和存储器结构来自 IJV 发布的 CHIRP 模块和对 RadioManager 4.3.2 的分析；代码从零编写，已用模拟电台测试。使用 IJV 固件的真实电台尚未测试：请报告任何问题。
