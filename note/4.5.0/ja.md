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
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub><b>日本語</b></sub></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0（Microsoft Windows 10/11 用）

### Retevis RT85 のファームウェア（ファームウェア更新対応は 44 機種）
- **Retevis RT85**：Retevis の .hex ファイル（例：「RT85-Firmware-V1.53.hex」）に対応。Retevis のプログラム「RT85 10W_5W_UpData」のプロトコルを使います。このプログラムは模擬無線機で動かし、送られるブロックとコマンドが 1 バイト単位で一致することを確認しました。
- Retevis の手順：無線機の電源を切り、**PTT と P1** を押しながら電源を入れます（緑のランプ）。次に無線機の出力（Retevis のプログラムと同じく **5 W または 10 W**）を選び、確認のため RT85 と入力します。

### 修正：Retevis RA89 の更新
- 4.4 はハンドシェイクに対して無線機から誤った応答を待っていたため、実機の RA89 では更新が始まりませんでした（無線機は変更されていません）。応答と待ち時間を Retevis のプログラムと同じにしました。

### ポート、「識別」タブと「変換」タブ
- ケーブルのポートはどのシステムでも常に番号順です：COM1、COM2、COM3… COM10。自動識別も最初のポートから始まります。
- 「識別」タブ：「以前に識別した無線機はより速く認識されます」。
- 「変換」タブ：「Baofeng や Bajeton のファイル、その他すべてのメーカーの無線機の .img（CHIRP で保存したものも）、CHIRP の CSV、Excel の表または CSV を選んでください」。

### ウェブサイト
- ダウンロードボタンの下にあったリリースノートの文章を削除。
- タイトルを上に移動：ノートパソコンの画面でも「PRESS START」の下のバージョン番号が全部見えます。

ギャラリー：27 メーカー 386 機種（チャンネル対応 348、設定対応 302、ファームウェア更新対応 44）。
