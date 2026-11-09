# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub><b>日本語</b></sub></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-11-orange" alt="11 languages">
</p>

**バージョン 4.13.1** · Windows · 11 言語のインターフェース（italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語、한국어、العربية）· 実行ファイル 1 つだけ、インストール不要

アマチュア無線家と **44 ブランド**のトランシーバーの利用者のためのプログラムです（**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone）。プログラミングケーブルでチャンネル、周波数、設定を読み書きし、イタリアのレピーターのアーカイブを管理し、対応する無線機のファームウェアを更新します。

🌐 プログラムのサイト：**[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*このファイルは [README.it.md](../../README.it.md) の日本語訳です。イタリア語の原文が正式なテキストです。*

## 機能

- 451 機種の**チャンネルと周波数**（デジタル PMR や DMR 無線機を含み、USB 接続も可）：ケーブルによる読み書き、CHIRP 互換の `.img` ファイル、Baofeng K67 Pro（`.xlc`）と Bajeton 8300（`.dat`）の元のファイル。
- 397 機種の**無線機の設定**（メニュー、表示、VOX、スケルチ、TOT、DTMF、キー…）。各項目の説明があり、注意が必要な項目には確認付きの警告があります。
- **接続した無線機の識別**：このプログラムが自動で識別し、メーカー、機種、種類、ファームウェア、無線機が伝えるその他の情報を表示します（USB の DMR 無線機では DMR ID と名前も）。**自動識別**（「識別」タブのボタン、選択は記憶）を使えば、どのタブにいてもケーブルをつなぐだけです。すべてのケーブルのポートを同時に監視し、ケーブル接続後に電源を入れた無線機や、別々のケーブルにつないだ複数の無線機も識別します。過去に識別した無線機やよく使われる無線機を先に試し、ファームウェア更新モードの無線機にだまされず、それぞれのタブを読み込んだチャンネルとともに開きます。「無線機を識別中」の枠は応答する無線機のあるケーブルにだけ表示されます。別の無線機には右上の **「🔎 無線機を識別」** ボタンがあります。
- **別の無線機への転送**：ある無線機でチェックしたチャンネルを、メーカーと機種で選んだ接続中の別の無線機に直接書き込みます。互換性チェック、調整、バックアップ、確認付きです。
- **汎用コンバーター**：どの無線機やファイル（K67 `.xlc`、Bajeton `.dat`、CHIRP のものを含む `.img` イメージ、CHIRP の CSV、Excel/CSV）のチャンネルも、ほかの無線機の形式にその規則に合わせて変換します。
- **11 言語のインターフェース**：「ℹ Info」の隣の旗の選択で、プログラム全体をすぐに翻訳し（アラビア語は右から左）、言語は再起動後も記憶されます。説明ファイルの訳は `lingue` フォルダーにあります。
- 「ゾーン 1」「ゾーン 2」…の代わりの**ゾーン名**（Radtel RT-950 Pro、RT-950、RT-920、RT-910、Bajeton BJ-9000 と BJ-7700、Hiroyasu HI-UV850 Pro、DMR 無線機 RD-5R、DM-1801、DM-32UV のゾーン）。チャンネルと一緒にケーブルで読み書きします。
- チェックしたチャンネルの**一括編集**：出力、帯域幅、スキャン、サブトーン、シフト、送信をまとめて変更し、各チャンネルの残りはそのままです。
- 44 ブランド 489 機種の**ギャラリー**：対応と非対応、種類（アマチュア無線、DMR、GMRS/FRS、PMR446、子ども向け）、ブランドで分類し、検索、フィルター、公式プログラムやマニュアルへのリンク付き。
- [IK2ANE](https://www.ik2ane.it) の最新リストに基づく**レピーターのアーカイブ**：フィルターと無線機への入力、書き込み前のレピーターと無線機の互換性チェック付き。
- 57 機種の**ファームウェア更新**（Quansheng UV-K5、UV-K6、UV-5R Plus、Quansheng と Radtel の .bin ファイルによる Radtel RT-590 と RT-600、Radtel の .kdhx ファイルによる Radtel RT-470、RT-470L、RT-470X、RT-495、RT-630、メーカーの .BIN ファイルによる TIDRADIO TD-H3、TD-H8 と Radtel RT-730、.bin ファイルによる Retevis RA79 V3 と RA89、Retevis の .hex ファイルによる Retevis RT85、.BTF ファイルによる Radtel RT-950 Pro、RT-950 と Bajeton BJ-9000、.anc ファイルによる Radtel RT-920 と RT-910B、.BTU ファイルによる Radtel RT-910、RT-900 と Bluetooth なしの Bajeton BJ-7700、Radtel と M7OCM の .bin ファイルによる Radtel RT-890、USB で Baofeng の .bin ファイルによる Baofeng DM-1701、USB で .sgl ファイルによる Baofeng RD-5R と DM-1801 も）。ファイルのチェックと無線機のバックアップ付き。BTECH DMR-6X2 と DMR-6X2 PRO は公式プログラムによるガイド付き更新。
- 書き込みのたびに自動で**バックアップ**し、書き込み後に確認。
- 起動のたびにプログラムの**アップデートを確認**。

## 使い方

1. [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip)（最新版）をダウンロードして展開します。
2. `IU0VYT Flash.exe` を起動します（Windows 10/11。Windows の WebView2 コンポーネントを使います）。
3. **📻 機種** タブで無線機を選び、ケーブルをつないで **無線機から読み込む** を使います。

完全なガイドは `GUIDE.txt` ファイルにあります（イタリア語の原文：`LEGGIMI.txt`）。

## ビルド

Go 1.22 以降が必要です。`sorgenti` フォルダーで：

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

ライブラリは `vendor` フォルダーに含まれています。

## 注意事項

- 無線機のプロトコルは相互運用のために再構成し、模擬無線機で確認したものです。**バックアップは必ず保管してください**。このプログラムは一切の保証なしで提供されます。
- 独立したプログラムであり、無線機のメーカー（Baofeng、Bajeton、Quansheng、Retevis、Radtel、Radioddity、TIDRADIO など）とは**提携しておらず、承認も受けていません**。記載した商標はそれぞれの所有者に帰属し、対応する無線機を示すためだけに使っています。
- お住まいの国の周波数の利用に関する規則を守ってください。

## 謝辞

- [CHIRP](https://chirpmyradio.com) とその作者の皆さん：無線機のドライバーの多くは皆さんの成果に由来します。
- [IK2ANE](https://www.ik2ane.it)：イタリアのレピーターのリストを無料ですべての人に提供してくださっています。

構成要素の作者とライセンス：`AUTHORS_AND_LICENSES.txt`。

## ライセンス

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

**GNU 一般公衆利用許諾書バージョン 3 以降**（`GPL-3.0-or-later`）で配布するフリーソフトウェアです：`LICENSE.txt` と `LICENSE_NOTICE.txt` を参照してください。
