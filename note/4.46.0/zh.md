# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 上的发布页面（英文），含下载：[V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0，适用于 Microsoft Windows 10/11

### 修正
- **Radtel RT-880G**：不再被标为 GMRS / FRS 电台。它是业余无线电 VHF/UHF 收发信机（带 GPS 和 APRS），与 RT-880 和 RT-860 相同；写入前不再出现“电台的销售类型为 GMRS”的警告。
- **写入后的校验**（Radtel RT-880G）：以前会失败，因为电台在写入结束时保存数据并重启，而程序联系它太早、太频繁。现在程序先等待 10 秒、不打扰电台，然后每 15 秒重试一次，最长 90 秒。同样的时间也用于使用同一驱动的 RT-880、RT-860 和 RT-817A / RT-871A。
- 对来自 AR8200、Kenwood 和 TK-180 的数值进行更严格的检查（CodeQL 报告）。

### 说明
- 校验已在真实的 RT-880G 上测试；RT-880、RT-860 和 RT-817A / RT-871A 未在真机上测试。其他电台保持原来的时间。
- 本版本没有新增电台。
- 本程序为独立程序，与各电台厂家无关联。
- 完整测试全部通过：通用部分；915 项结构对比和 380 次与 CHIRP 完全一致的读取 / 写入。
