# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0，适用于 Microsoft Windows 10/11

### 新内容
- **Retevis，新增 24 款取自厂家程序的型号**：B3B、B63B、RT22P、RT622P、RB666、RB666P（信道存储与新版 H777 相同）；RB45、RB645、RB47；B3H、B63H、RB647、RT68H、RT668H、NR10、NR610、RB66S、RB666S、RT68B、RT668B、RB669（T18 系列）；RB65、RB66、RB66P（每个信道一个频率和一个亚音）。连接、读取和写入均从零重写，并用各程序的出厂数据文件核对。RB66 与 RB66P，以及 RB669、RT68H 与 RT668H 的应答代号相同：需要手动选择正确的卡片。
- **24 款 Retevis 型号的设置**：共 241 项（静噪、发射限时、VOX、扫描、语音提示、省电、按键音、侧键……）。只提供该型号的厂家程序显示的项目；24 个出厂文件中的数值全部是允许的选项。
- **编程密码**：在这 24 款 Retevis 型号和新版 Retevis H777 上，它不会阻止读取或写入，与 CHIRP 相同（厂家程序则会在密码与用户输入的不一致时停止）。它是设置中的一项，以明文显示：可以查看、修改和清除。长度和字符与厂家程序一致：6 个字符（RB45、RB645、RB47 和 RB66 为 3 个；RB66P 为 4 个），数字和字母（RB65 和 RB66 仅限数字）。B3H、B63H、RB45、RB645、RB66 和 RB666 的厂家程序不显示密码输入框，对此有一条警告提示。
- **为原先只有信道的电台增加设置**：AnyTone 系列（AT-D868UV、AT-D878UV、AT-D878UV II Plus、BTECH DMR-6X2 和 DMR-6X2 PRO：27 项，依据 qdmr 和 dmrconfig 的存储说明）；TYT MD-UV380、MD-UV390、MD-2017 和 MD-9600、Retevis RT3S 和 RT82（与 Baofeng DM-1701 相同的项目）；使用 nicFW880 / nicFW950 固件的 Radtel RT-880 和 RT-950 Pro（17 项，依据固件作者的文档）；TID TD-UV68（与 TD-H8 相同的项目）；使用 M7OCM 固件的 Radtel RT-890（53 项）。
- **测试版**：版本号高于已发布的最新稳定版的程序会显示一条警告栏（“测试版（beta 测试）…”），并且只有它才会提示下一个测试版，测试版始终从 GitHub 下载。使用稳定版的用户看不到测试版。
- 图库：54 个品牌的 697 个型号（647 个支持信道，596 个支持设置）。

### 说明
- 这 24 款 Retevis 电台的信道和设置均未在真机上测试：请先只读取，然后只改一项无害的设置。如果电台的应答出乎意料，读取会停止并给出提示，不会向电台写入任何内容。
- 电台已设有编程密码时，程序仍会尝试读取和写入，与 CHIRP 相同：这一点未在真机上测试。
- RB66 与 RB66P，以及 RB669、RT68H 与 RT668H 的应答代号相同：需要手动选择正确的卡片。
- 本程序为独立程序，与各电台厂家无关联。
- 完整测试全部通过：通用部分；915 项结构对比和 380 次与 CHIRP 完全一致的读取 / 写入。
