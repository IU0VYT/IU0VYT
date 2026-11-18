# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0，适用于 Microsoft Windows 10/11

本版本还包含 4.3.0、4.2.0、4.1.0 和 4.0.0 的新内容，这些版本没有单独发布稳定版。

## 4.4.0 的新内容

### 已修正：没有电台时出现“正在识别电台”窗口
- 只连接数据线而没有电台时，该窗口也可能出现：线路干扰和发送命令的回声看起来像电台在应答。
- 现在**只有连接着真正应答的电台的那根数据线**才会出现该窗口。两根数据线只接一台电台时（COM6 没有电台，COM7 有电台，或反过来），只对电台所在的端口显示。
- 没有电台时，一轮尝试也更快了。

### 顶部栏、首页、“识别”标签页
- 按钮名为 **“🔎 识别电台”**，位于右侧“更新”旁边。
- 打开很多标签页时，按钮会自动变窄，最后只剩放大镜图标。
- 首页：恢复了 4.2 中各方框的说明。底部的版权、网站和源代码始终可见，在 1280×720 下也是如此。
- “识别”标签页：说明简短清楚。

### Retevis 固件（43 个型号支持固件升级）
- **Retevis RA79 V3**（以及 UV-K5 V3 / UV-K1，引导程序 7.x），使用 Retevis 的 .bin 文件，例如“RA79 V3.bin”。只有当电台的引导程序属于同一系列时才会发送固件。
- **Retevis RA89**，使用 Retevis 的 .bin 文件，例如“Eng_RA89_Upadat_20250524_B91.38.bin”。采用 Retevis 程序“RA89-update-software”的协议，该程序已在模拟电台上运行。发送的命令逐字节一致。

### 其他
- 使用模拟电台的测试不再妨碍在 Windows 上编译测试。

型号库：27 个品牌的 386 个型号（348 个支持信道，302 个支持设置，43 个支持固件升级）。

## 4.3.0 的新内容

### 自动识别：更快、更多电台、更多端口
- 没有电台时，完整一轮尝试约需 **50 秒**（之前为 135 秒）。
- 程序**持续监听所有数据线端口**，每个端口各自独立，也同时进行：
  - 关机连接、之后才开机的电台也能识别；
  - 可同时识别**不同数据线上的多台电台**，显示它们的数据并打开它们的所有标签页。
- **已识别过的电台**（例如 K67 Pro、Bajeton 8300、BJ-9000）优先尝试，然后才是其他电台。
- 如果电台的标签页已经打开，会显示“电台已识别！！电台标签页已打开”（多台电台时为“电台已识别！！电台标签页已打开”）。
- “识别”标签页：端口按顺序排列（COM6、COM7… COM10），显示为“COM6 — USB-SERIAL CH340”；“自动识别新电台”用粗体显示。

### 固件升级：新增 8 款电台（共 41 款）
- **Radtel RT-470、RT-470L、RT-470X、RT-495、RT-630**，使用 Radtel 的 .kdhx 文件（引导程序“JJCC Bootloader Pro”）。RT-470：1.x 固件只适用于旧主板，2.x 只适用于新主板。程序会提示并要求书面确认。
- **TIDRADIO TD-H3、TD-H8 和 Radtel RT-730**，使用厂家的 .BIN 文件：开始升级后，按住 PTT 开启电台。

### 其他
- **Bajeton BJ-9000**：与 BJ-8300 相同，全部 960 个信道放在一个不分区域的列表中（“信道：全部（960）”）。
- **首页**更紧凑：底部始终显示版本、版权、网站和源代码。
- **型号库**：用一条粗线把兼容型号和不兼容型号分开。
- **LEGGIMI**：删除了 4.2.0 中误重复的一大段内容。

型号库：27 个品牌的 386 个型号（348 个支持信道，302 个支持设置，41 个支持固件升级）。

## 4.2.0 的新内容

- 修正自动识别：只有当连接的电台应答时才会出现“正在识别电台”窗口（不再在只连数据线启动时出现），电台被识别或未找到时窗口消失。识别后会打开电台的数据和标签页，搜索停止，但功能不会被关闭。
- 顶部栏“信息”和“更新”之间新增“🔎 自动识别新电台”按钮：按同样的规则搜索新电台，自动识别关闭时按钮为灰色。
- 新增 35 款电台（36 个版本）和 8 个品牌：TYT TH-UV88 系列（TH-UV88、TH-UV98、Retevis RT85、RA89、P2、P62、Explorer QRZ-1），Radtel RT-490 系列（含 JC-8629、FB-8629、Jianpai 8800 Plus、Boristone 8RS、Abbree AR-869、HamGeek HG-590），Retevis RT21 系列的 19 款电台（含 Abbree AR-63）以及 Quansheng TK11。
- Radtel RT-590 / RT-590 Plus 和 RT-600 / RT-600 Plus 也支持使用 Radtel 的 .bin 文件升级固件。
- 修正了读取带重音字母或中文的信道名称时的一些隐藏缺陷。
- 型号库：27 个品牌的 386 个型号：348 个支持信道和中继台，302 个支持设置，33 个支持固件升级。

## 4.1.0 的新内容

- 修正自动识别：在任何标签页中都能工作；当你拔下、关闭或更换电台时，搜索会自动重新开始，并识别之后连接的电台（即使是同一台，或在接好数据线后才开机）。之前在第一次识别后就不再搜索其他电台。
- 自动搜索期间会显示“正在识别电台，请稍候。”窗口，并带有“中止”按钮。
- 新增 14 款电台和 3 个品牌：Radtel RT-470、RT-470L、RT-470X、RT-470X Plus、RT-630、RT-495 和 RT-730，Hiroyasu HI-8811，Talkpod A36plus 和 A36plus 8 W，Anysecu UV-A37，Abbree AR-730，TIDRADIO TD-H8（第 2 代和第 3 代）和 TD-H3，均有基础版、HAM 版和 GMRS 版。
- 可使用 Quansheng 的 .bin 文件升级 Quansheng UV-K5、UV-K5(8)、UV-K6 和 UV-5R Plus 的固件，并检查固件是否适用于所连接的电台类型。
- 首页和网站：“信道管理 - 频率 - 中继台 - 固件”以及所有品牌名称，放在 80 年代电子游戏风格的边框中。
- 型号库：19 个品牌的 349 个型号：311 个支持信道和中继台，265 个支持设置，31 个支持固件升级。

## 4.0.0 的新内容

- 多品牌程序：型号库中有 16 个品牌的 335 个型号，297 个可通过数据线管理信道和频率，251 个支持电台设置。
- 新电台：Quansheng UV-K5、UV-K6、UV-5R Plus 和 UV-K1；31 个 Retevis 型号（H777、RT22、RT22S、RB18、RB19、RB29、RT47、RT68、RB85、RA79、RA685…）；Radtel RT-900、RT-900 BT、RT-920 和 T18；Radioddity GM-30、GM-30 Plus、GM-30 Pro、GS-10B、GA-2S 和 GA-510；TIDRADIO TD-H6、MaxTalker、Abbree、Intek、Maverick、BTECH、WLN、Zastone、Luiton 和 TID。
- 自动识别（“识别”标签页）：接上数据线，电台会被自动识别，并打开其标签页、读取信道。该选择会被记住，可用同一个按钮关闭。
- 型号库带有品牌菜单，每个品牌一种颜色。
- 电台标签页中的“从电台读取”和“写入电台”改为蓝色。
- 首页、信息、指南、网站和文档已更新为多品牌程序。
