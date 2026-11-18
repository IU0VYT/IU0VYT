# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0，适用于 Microsoft Windows 10/11

### 区域名称
- 每个区域都可以有自己的名称，取代“区域 1”“区域 2”……与厂家程序相同：使用“✎ 区域名称…”按钮，或双击区域。区域栏和底部信息行显示“编号 · 名称”。
- 电台：Radtel RT-950 Pro、Bajeton BJ-9000 和 Hiroyasu HI-UV850 Pro（固件 0.23 及以前和 0.24 及以后），Radtel RT-950、Radtel RT-920、无蓝牙的 Radtel RT-910 / Bajeton BJ-7700；以及 Baofeng DMR 电台 RD-5R、DM-1801 和 DM-32UV 的区域。
- 名称与信道一起通过数据线读写，并保存在 .img 文件中。长度和字符与厂家程序相同（RT-950 Pro 为 12 个字符，可用中文；RT-920 为 10 个；DMR 电台为 16 个）。
- 安全：只有名称改变时才写入名称区域，并且写入前会先从电台回读；如果电台在该位置存放的不是名称数据，则不写入任何内容。

### 批量修改
- “✎ 批量修改…”（也可在右键菜单中使用）可同时修改已勾选（或已选中）信道的功率、带宽、扫描、发射亚音、接收亚音、频差和发射。保持“不更改”的项目不变；Ctrl+Z 可全部撤销。

### 新增两种语言（共 11 种）
- 韩语，使用韩国国旗。
- 阿拉伯语，使用阿联酋国旗，页面从右向左排列（频率和数字仍从左向右）。
- `lingue` 文件夹中的所有说明文件也已译成韩语和阿拉伯语，并在所有语言中更新到 4.11。

### 其他
- 网站：列出了可用语言；删除了固件文件列表。
- `go test` 在 Windows 上通过（包括 i18n_test.go）；新增测试，检查主文件夹和“sorgenti”中的 LEGGIMI.txt、AUTORI_E_LICENZE.txt 和许可证是否完全相同。

新的文件格式来自厂家程序的代码和 CHIRP 驱动，并已用模拟电台测试；尚未在真实电台上测试：请始终保留备份。

型号库：44 个品牌的 489 个型号（451 个支持信道，396 个支持设置，57 个支持固件升级）。
