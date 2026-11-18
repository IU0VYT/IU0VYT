# IU0VYT Flash 4.10.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.10.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.10.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.10.0)

IU0VYT Flash V.4.10.0，适用于 Microsoft Windows 10/11

### 程序支持 9 种语言
- 新的带国旗的语言选择器，就在“ℹ 信息”右侧：italiano、English、Français、Deutsch、Español、Português、Русский、中文、日本語。
- 切换语言立即生效，无需重启：标签页、按钮、菜单、窗口、指南、消息和警告（包括来自电台的）、型号库卡片、设置，以及 Windows 的“打开”和“保存”窗口。
- 最后选择的语言在重启后仍会保留。
- 说明文件已翻译并放在 `lingue` 文件夹中（每种语言一个子文件夹）：完整指南、README、包含 4.10 新内容的 CHANGELOG、作者与许可证、版本以及许可证说明。有效的许可证仍是 GNU GPL 的英文文本（LICENSE.txt）。

### 顶部栏
- 删除了“更新”按钮：每次启动时仍会自动检查，也可在“信息”标签页中使用“🔄 检查更新”。
- “🔎 识别电台”现在位于“? 指南”旁边。

### 固件升级：新增 5 款电台（共 57 款）
- Radtel RT-890：Radtel 的 .bin 文件和开源固件（M7OCM、OEFW）；刷入 M7OCM 固件后，信道使用其专用标签页管理。
- Baofeng DM-1701 / Retevis RT84 通过 USB，使用 Baofeng 的 .bin 文件：只写入程序存储区，不改动电台的引导程序。
- Baofeng RD-5R 和 DM-1801 通过 USB（无需驱动），使用厂家的 .sgl 文件：电台在擦除存储器之前会检查型号。

### 设置：新增 10 款电台（共 396 款）
- Feidaxin FD-288A、FD-150A、FD-160A、FD-450A、FD-460A、FD-460UH 和 Retevis RT76P（来自 CHIRP）。
- Baofeng RD-5R 和 DM-1801（31 项），Baofeng DM-1701（名称、DMR ID、开机文字）。

这些协议均为从零重写，并已使用模拟电台和厂家的真实文件测试；尚未在真实电台上测试：请始终保留备份。

型号库：44 个品牌的 489 个型号（451 个支持信道，396 个支持设置，57 个支持固件升级）。
