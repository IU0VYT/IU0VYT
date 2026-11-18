# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0，适用于 Microsoft Windows 10/11

本版本还包含 4.8.0、4.7.0 和 4.6.0 的新内容，这些版本只作为测试版发布过。

## 4.9.0 的新内容

### 已修正：识别 Lanchonlh HG-UV98 和 Quansheng TG-UV2+
- “识别”时端口始终使用 8N1。现在改用偶校验（HG-UV98）和两个停止位（TG-UV2+），与在电台标签页中读取时一致。这两款电台现在可以被识别。

### 已修正：HG-UV98 的校验
- 使用偶校验时，Windows 端口现在也会检查校验错误。之前该检查刚打开就又被关闭了。

### 使用开源固件的 Quansheng UV-K5 / UV-K6 / UV-5R Plus
- 为 egzumer、F4HWN 和 LOSEHU/LTS/UVE 固件新增标签页。
- F4HWN 使用 armel 的 F4HWN 驱动（不在 CHIRP 中）：8 档功率、发射锁定、3 个扫描列表以及固件设置。
- 每个标签页只读写自己的固件。固件升级后，程序会提示应使用哪个标签页。

### 新电台
- Wouxun（20 个型号）：KG-UVD1P（KG-UV2D、KG-UV3D）、KG-UV6D、KG-816、KG-818、KG-805G、KG-UV8D、KG-UV8D Plus、KG-UV8E、KG-UV9D Plus、KG-UV9PX、KG-UV9GX、KG-UV9K、KG-UV9G Pro、KG-935G、KG-935G Plus、KG-935H、KG-UV8H、KG-XS20G、KG-XS20G Plus、KG-XS20H
- TYT TH-UV8000D/E、UVL-15W、TH-350（美国）
- Abbree AR-518、Ruyage UV58Plus、Radtel RT-880G、Radtel RT-890（Radtel 固件和开源固件 M7OCM）
- Retevis C2、HA1G、HA1UV、HA2、RB15、RB615、RB17P、RB28、RB628、RB48、RB648、RT23、RT26、RT76P、RT87、RT16、RB27B、RB27V
- BTECH GMRS-V1、GMRS-V2、MURS-V1、MURS-V2、FRS-A1
- AnyTone TERMN-8R、OBLTR-8R
- Puxing PX-888K、QYT KT-8R、Rugged RH5R-V2、Baojie BJ-UV55、MTC UV-5R-3、Radioddity MU-5、Zartek ZA-758、TIDRADIO TD-M11、TDXone TD-Q8A

对于来自 CHIRP 驱动的电台，信道和设置都通过与 CHIRP 对比的自动测试进行了校验。GMRS、FRS 和 MURS 电台用于美国：在意大利不允许在这些频率上发射。

## 4.8.0 的新内容

### .BTU 固件升级（无蓝牙的 Radtel RT-910 和 RT-900，无蓝牙的 Bajeton BJ-7700）
- .BTU 固件使用“Radtel Bootload Tools”程序的协议升级。引导程序与 Baofeng UV-5RM 相同，握手为“PROGRAMBT80U”。
- 自动模式：电台正常开机。手动模式：电台关机，按住两个侧键开机，直到显示屏出现“UPDATE”。
- 引导程序不检查型号。因此程序会根据文件名识别电台和硬件版本，显示警告，并且总是要求输入型号。
- 型号库中 RT-910（无蓝牙，.BTU）和 RT-910B（有蓝牙，.anc）现在分为两张卡片。BJ-7700 和无蓝牙的 BJ-7700 也是如此。

### 已修正：使用 512 信道固件的无蓝牙 Radtel RT-900
- V1.09P 及以前电台有 512 个信道；从 V1.14P 起有 999 个。
- 现在只有在你修改了 513-999 号信道时才会写入它们。写入前，程序会先在电台中检查该区域为空或存放的是信道。

### 新电台（CHIRP 驱动）
- TYT TH-UV3R、TH-UV3R-25、TH-UVF1、TH-UVF8D
- KYD NC-630A、IP-620
- KSUN M6
- Puxing PX-777、PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24、RT24V、H777S、RT1、H777 V4（H777 的新版本）、RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A、FD-268B、FD-288A、FD-288B、FD-150A、FD-160A、FD-450A、FD-460A、FD-460UH

以上所有电台的信道和设置都通过与 CHIRP 对比的自动测试进行了校验。

## 4.7.0 的新内容

### 已修正：自动识别卡住或非常慢
- 处于固件升级模式的电台会自行周期性地发送消息：以前它看起来像在回应每一个型号，搜索要持续好几分钟。现在程序在尝试前先监听端口，不会再被误导：搜索在一分钟内结束，“识别”标签页会说明正在发生什么。
- 启动搜索时界面不再卡住；最常见的电台优先尝试。
- 修正了对已关闭的标签页误报“电台标签页已打开”的问题。

### 管理方式相同的电台
- 选择窗口和“识别”标签页会列出所有可能的电台（例如 Radtel RT-950 Pro、Bajeton BJ-9000、Hiroyasu HI-UV850 Pro），提供选择菜单，并简要说明在电台上哪里可以看到固件版本。

### 已修正：固件 0.23 及以前的 Radtel RT-950 Pro / Bajeton BJ-9000
- 信道的“允许发射”位现在总是会写入，两种版本都如此（与 Radtel 程序 v1.2.x 及以前相同；v1.3.3 忽略该位）。使用旧固件时，在 990 信道标签页中修改的信道可能无法发射。缺少该位的 1-960 号信道也会补上。
- 核实了 Bajeton BJ-9000 的实际版本：固件 V0.22、V0.27 和 V0.29（与 Radtel 的相同），程序 v1.1.1（960 信道）和 v1.3.3（990 信道，即 Radtel 程序）。信道和设置也用 Bajeton 程序的代码进行了校验。

### 更多电台支持固件升级
- Radtel RT-950 Pro、Radtel RT-950 和 Hiroyasu HI-UV850 Pro（.BTF 文件，引导程序与 BJ-9000 相同）。
- Radtel RT-920 和 RT-910B（.anc 文件，协议与 BJ-7800 和 BJ-7700 相同）。

### 新电台（CHIRP 驱动）
- Radtel RT-620、Radioddity GA-510 V2、TYT TH-350、Retevis RT18、Retevis RB627B（PMR446）；型号库中还有 Hiroyasu HI-UV850 Pro。
- 还为 SenhaiX 8800、Radioddity GS-5B、Cignus XTR-5 和 Anysecu AC-580 增加了设置。

## 4.6.0 的新内容

### 已修正：使用新固件的 Radtel RT-950 Pro / Bajeton BJ-9000
- 从固件 **0.24**（2025 年 12 月）起，电台有 **990 个信道，分为 10 个区域、每区 99 个**（或一个 990 个信道的列表）；**0.23** 及以前有 **960 个，分为 15 个区域、每区 64 个**（或一个 960 个信道的列表）。程序现在支持这两种版本。
- 版本根据从电台读取的数据识别（电台本身不会报告版本）：读取后会在正确的标签页中打开信道。如果数据不足，会有说明告诉你如何在电台上查看版本。
- 961-990 号信道只有在改变时才写入，并且写入前会先从电台回读；如果电台没有这些信道（固件 0.23 及以前），则不写入任何内容。
- 信道和设置已用 Radtel 程序 v1.2.0（960 信道）和 v1.3.3（990 信道）的原始代码校验。

### 新增：Radtel RT-950（不带“Pro”）
- 960 个信道（15 个区域、每区 64 个）或一个列表，以及常规设置；已用“BT-9000_CPS”v1.3.0 程序的原始代码校验。

### 已修正：电台的准确名称
- 管理方式相同的不同电台（例如 **Radtel RT-950 Pro** 和 **Bajeton BJ-9000**、**Baofeng K67 Pro** 和 **5RH Pro**、Radtel RT-910 和 Bajeton BJ-7700）现在在标签页、读取、写入、识别和备份文件中都使用各自的名称。
- 第一次识别到其中一款电台时，程序会询问是哪一款，之后会记住。

### 已修正：未保存数据的提醒
- 读取、写入或转发之后，信道已经在电台中：关闭程序时，只有存在真正的修改才会出现提醒。

### 新电台（CHIRP 驱动）
- SenhaiX 8800、Radioddity GS-5B、Cignus XTR-5、Anysecu AC-580、TID TD-UV68。
- Radtel RT-920：990 个信道分为 10 个区域、每区 99 个，与电台上相同。

型号库：44 个品牌的 489 个型号（451 个支持信道，386 个支持设置，52 个支持固件升级）。新品牌：Wouxun、AnyTone、QYT、Rugged、Baojie、MTC、Zartek、Ruyage、TDXone。
