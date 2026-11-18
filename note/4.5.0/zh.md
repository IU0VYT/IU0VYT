# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0，适用于 Microsoft Windows 10/11

### Retevis RT85 固件（44 个型号支持固件升级）
- **Retevis RT85**，使用 Retevis 的 .hex 文件，例如“RT85-Firmware-V1.53.hex”。采用 Retevis 程序“RT85 10W_5W_UpData”的协议，该程序已在模拟电台上运行：发送的数据块和命令逐字节一致。
- Retevis 的操作步骤：电台关机，按住 **PTT 和 P1** 开机（绿色指示灯）。然后选择电台功率（**5 W 或 10 W**，与 Retevis 程序相同），并输入 RT85 确认。

### 已修正：Retevis RA89 的升级
- 4.4 版在握手时期待电台给出错误的应答：使用真实的 RA89 时升级无法开始（电台未被改动）。现在期待的应答和停顿与 Retevis 程序一致。

### 端口、“识别”和“转换”标签页
- 数据线端口在任何系统上都按顺序排列：COM1、COM2、COM3… COM10。自动识别也从第一个端口开始。
- “识别”标签页：“以前识别过的电台会更快被识别”。
- “转换”标签页：“选择 Baofeng、Bajeton 文件，其他所有品牌电台的 .img 文件（包括用 CHIRP 保存的），CHIRP 的 CSV，Excel 表格或 CSV”。

### 网站
- 删除了下载按钮下方的版本说明文字。
- 标题位置更高：笔记本电脑屏幕上也能完整看到“PRESS START”下方的版本号。

型号库：27 个品牌的 386 个型号（348 个支持信道，302 个支持设置，44 个支持固件升级）。
