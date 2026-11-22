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
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub><b>日本語</b></sub></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0（Microsoft Windows 10/11 用）

### 修正
- **Radtel RT-880G**：GMRS / FRS 無線機として表示されなくなりました。RT-880 や RT-860 と同じアマチュア無線用の VHF/UHF トランシーバー（GPS・APRS 付き）で、書き込みの前に「販売時の区分 GMRS」の警告が出なくなりました。
- **書き込み後の確認**（Radtel RT-880G）：書き込みの終わりにデータを保存して再起動する無線機に、早すぎるタイミングで何度も問い合わせていたため失敗していました。現在は 10 秒間無線機に触れずに待ち、15 秒ごとに再試行します（最大 90 秒）。同じドライバーを使う RT-880、RT-860、RT-817A / RT-871A にも同じ待ち時間を使います。
- AR8200、Kenwood、TK-180 から受け取る値のチェックを厳しくしました（CodeQL の指摘）。

### 注意
- 確認は実機の RT-880G で試験済みです。RT-880、RT-860、RT-817A / RT-871A は実機では未確認です。ほかの無線機の待ち時間は従来どおりです。
- このバージョンでの新しい無線機はありません。
- 本プログラムは独立したもので、無線機メーカーとは関係ありません。
- 全テストに合格：一般部分、915 件の構造比較と CHIRP と同一の 380 件の読み込み / 書き込み。
