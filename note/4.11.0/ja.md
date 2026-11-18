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
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub><b>日本語</b></sub></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0（Microsoft Windows 10/11 用）

### ゾーン名
- 「ゾーン 1」「ゾーン 2」…の代わりに、メーカーのプログラムと同じく各ゾーンに名前を付けられます：「✎ ゾーン名…」ボタン、またはゾーンをダブルクリック。ゾーンのバーと下の行には「番号 · 名前」が表示されます。
- 対象：Radtel RT-950 Pro、Bajeton BJ-9000、Hiroyasu HI-UV850 Pro（ファームウェア 0.23 まで、および 0.24 以降）、Radtel RT-950、Radtel RT-920、Bluetooth なしの Radtel RT-910 / Bajeton BJ-7700、Baofeng の DMR 無線機 RD-5R、DM-1801、DM-32UV のゾーン。
- 名前はチャンネルと一緒にケーブルで読み書きされ、.img ファイルに保存されます。長さと文字はメーカーのプログラムと同じです（RT-950 Pro は中国語も含め 12 文字、RT-920 は 10 文字、DMR は 16 文字）。
- 安全：名前の領域は、名前が変わったときだけ、無線機から読み直したうえで書き込みます。その場所に名前ではないデータがある場合は、何も書き込みません。

### 一括編集
- 「✎ 一括編集…」（右クリックからも）で、チェックした（または選択した）チャンネルの出力、帯域幅、スキャン、送信トーン、受信トーン、シフト、送信をまとめて変更できます。「変更しない」のままの項目はそのまま残り、Ctrl+Z ですべて元に戻せます。

### 言語を 2 つ追加（合計 11）
- 韓国語（韓国の国旗）。
- アラビア語（アラブ首長国連邦の国旗）、ページは右から左（周波数と数字は左から右のまま）。
- `lingue` フォルダーのすべての説明ファイルを韓国語とアラビア語にも翻訳し、全言語で 4.11 に更新しました。

### その他
- ウェブサイト：利用できる言語を表示し、ファームウェアファイルの一覧を削除。
- `go test` が Windows で通るように（i18n_test.go も）。新しいテストで、LEGGIMI.txt、AUTORI_E_LICENZE.txt とライセンスがメインフォルダーと「sorgenti」で同一かを確認します。

新しい形式はメーカーのプログラムのコードと CHIRP のドライバーから導き、模擬無線機で試験しましたが、実機では試験していません。必ずバックアップを保存してください。

ギャラリー：44 メーカー 489 機種（チャンネル対応 451、設定対応 396、ファームウェア更新対応 57）。
