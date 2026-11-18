# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0（Microsoft Windows 10/11 用）

このバージョンには、単独の安定版リリースがなかった 4.3.0、4.2.0、4.1.0、4.0.0 の新機能も含まれます。

## 4.4.0 の新機能

### 修正：無線機がないのに「無線機を識別中です」の枠が出る
- 無線機のないケーブルでも枠が出ることがありました：回線のノイズや送ったコマンドのエコーが、無線機の応答のように見えていたためです。
- 今は**本当に応答する無線機がつながっているケーブルだけ**に出ます。ケーブル 2 本に無線機 1 台（COM6 は無線機なし、COM7 は無線機あり、またはその逆）の場合、無線機のポートだけに出ます。
- 無線機がないときの一巡の試行も速くなりました。

### 上部のバー、ホーム、「識別」タブ
- ボタンの名前は **「🔎 無線機を識別」** で、右側の「アップデート」の隣にあります。
- タブを多く開くとボタンは自動で縮み、最後は虫めがねだけになります。
- ホーム：4.2 の枠の説明が戻りました。下部の著作権表示、ウェブサイト、ソースコードは 1280×720 でも見えたままです。
- 「識別」タブ：短く分かりやすい説明。

### Retevis のファームウェア（ファームウェア更新対応は 43 機種）
- **Retevis RA79 V3**（および UV-K5 V3 / UV-K1、ブートローダー 7.x）：Retevis の .bin ファイル（例：「RA79 V3.bin」）に対応。無線機のブートローダーが同じファミリーの場合だけファームウェアを送ります。
- **Retevis RA89**：Retevis の .bin ファイル（例：「Eng_RA89_Upadat_20250524_B91.38.bin」）に対応。Retevis のプログラム「RA89-update-software」のプロトコルを使い、このプログラムを模擬無線機で動かしました。送るコマンドは 1 バイト単位で一致します。

### その他
- 模擬無線機を使うテストのせいで、Windows でテストがコンパイルできなくなることはなくなりました。

ギャラリー：27 メーカー 386 機種（チャンネル対応 348、設定対応 302、ファームウェア更新対応 43）。

## 4.3.0 の新機能

### 自動識別：より速く、より多くの無線機、より多くのポート
- 無線機なしでの一巡の試行は約 **50 秒**（以前は 135 秒）。
- プログラムは**すべてのケーブルのポートを常に監視**します。ポートごとに独立して、しかも同時に：
  - 電源を切ったまま接続し、あとで電源を入れた無線機も認識します；
  - **別々のケーブルにつないだ複数の無線機**を同時に認識し、それぞれのデータを表示してタブをすべて開きます。
- **すでに認識した無線機**（例：K67 Pro、Bajeton 8300、BJ-9000）を先に試し、その後でほかのすべてを試します。
- 無線機のタブがすでに開いている場合は「無線機は識別済みです!! 無線機のタブはすでに開いています」と表示されます（複数の場合は「無線機は識別済みです!! 無線機のタブはすでに開いています」）。
- 「識別」タブ：ポートは番号順（COM6、COM7… COM10）で「COM6 — USB-SERIAL CH340」の形で表示。「新しい無線機を自動識別」は太字。

### ファームウェア更新：8 機種追加（合計 41）
- **Radtel RT-470、RT-470L、RT-470X、RT-495、RT-630**：Radtel の .kdhx ファイル（ブートローダー「JJCC Bootloader Pro」）。RT-470：ファームウェア 1.x は旧基板専用、2.x は新基板専用です。プログラムがそれを示し、文字入力での確認を求めます。
- **TIDRADIO TD-H3、TD-H8 と Radtel RT-730**：メーカーの .BIN ファイル。更新を始めてから、PTT を押したまま無線機の電源を入れます。

### その他
- **Bajeton BJ-9000**：BJ-8300 と同じく、960 チャンネルすべてをゾーンなしの 1 つのリストに（「チャンネル：すべて（960）」）。
- **ホーム**をコンパクトに：下部にバージョン、著作権表示、ウェブサイト、ソースコードが常に見えます。
- **ギャラリー**：太い線で対応機種と非対応機種を分けました。
- **LEGGIMI**：4.2.0 で誤って繰り返していた長い部分を削除。

ギャラリー：27 メーカー 386 機種（チャンネル対応 348、設定対応 302、ファームウェア更新対応 41）。

## 4.2.0 の新機能

- 自動識別を修正：「無線機を識別中です」の枠は、接続した無線機が応答したときだけ表示され（ケーブルだけで起動したときには出なくなりました）、無線機を認識したとき、または見つからなかったときに消えます。認識後は無線機のデータとタブが開き、機能を無効にすることなく検索が止まります。
- 上部のバーの「情報」と「アップデート」の間に新しいボタン「🔎 新しい無線機を自動識別」：同じルールで新しい無線機を探します。自動識別が無効のときは灰色です。
- 新しい無線機 35 機種（36 バージョン）と 8 メーカーを追加：TYT TH-UV88 ファミリー（TH-UV88、TH-UV98、Retevis RT85、RA89、P2、P62、Explorer QRZ-1）、Radtel RT-490 ファミリー（JC-8629、FB-8629、Jianpai 8800 Plus、Boristone 8RS、Abbree AR-869、HamGeek HG-590 を含む）、Retevis RT21 ファミリーの 19 機種（Abbree AR-63 を含む）、Quansheng TK11。
- Radtel RT-590 / RT-590 Plus と RT-600 / RT-600 Plus も、Radtel の .bin ファイルでファームウェアを更新できます。
- アクセント付きの文字や中国語を含むチャンネル名の読み込みで、隠れていた不具合をいくつか修正。
- ギャラリー：27 メーカー 386 機種。チャンネルとレピーター対応 348、設定対応 302、ファームウェア更新対応 33。

## 4.1.0 の新機能

- 自動識別を修正：どのタブでも動き、無線機を外す、電源を切る、別の無線機に替えると、検索が自動でやり直され、次につないだ無線機を認識します（同じ無線機でも、ケーブルの後で電源を入れても）。以前は最初の認識の後、ほかの無線機を探さなくなっていました。
- 自動検索中は「無線機を識別中です。お待ちください。」の枠が「中止」ボタン付きで表示されます。
- 新しい無線機 14 機種と 3 メーカーを追加：Radtel RT-470、RT-470L、RT-470X、RT-470X Plus、RT-630、RT-495、RT-730、Hiroyasu HI-8811、Talkpod A36plus と A36plus 8 W、Anysecu UV-A37、Abbree AR-730、TIDRADIO TD-H8（第 2・第 3 世代）と TD-H3。いずれも標準、HAM、GMRS の各バージョン。
- Quansheng UV-K5、UV-K5(8)、UV-K6、UV-5R Plus のファームウェアを Quansheng の .bin ファイルで更新。ファームウェアが接続した無線機の種類に合っているかを確認します。
- ホーム画面とウェブサイト：「チャンネル管理 - 周波数 - レピーター - ファームウェア」と、すべてのメーカー名を 80 年代のビデオゲーム風の枠に。
- ギャラリー：19 メーカー 349 機種。チャンネルとレピーター対応 311、設定対応 265、ファームウェア更新対応 31。

## 4.0.0 の新機能

- 複数メーカー対応のプログラム：ギャラリーに 16 メーカー 335 機種。ケーブルでのチャンネルと周波数対応 297、無線機の設定対応 251。
- 新しい無線機：Quansheng UV-K5、UV-K6、UV-5R Plus、UV-K1、Retevis 31 機種（H777、RT22、RT22S、RB18、RB19、RB29、RT47、RT68、RB85、RA79、RA685…）、Radtel RT-900、RT-900 BT、RT-920、T18、Radioddity GM-30、GM-30 Plus、GM-30 Pro、GS-10B、GA-2S、GA-510、TIDRADIO TD-H6、MaxTalker、Abbree、Intek、Maverick、BTECH、WLN、Zastone、Luiton、TID。
- 自動識別（「識別」タブ）：ケーブルをつなぐと無線機が自動で認識され、そのタブが開いてチャンネルが読み込まれます。設定は記憶され、同じボタンで無効にできます。
- メーカーのメニューと、メーカーごとの色が付いた機種のギャラリー。
- 無線機のタブの「無線機から読み込む」と「無線機に書き込む」を青色に。
- ホーム、情報、ガイド、ウェブサイト、ドキュメントを複数メーカー対応のプログラムに合わせて更新。
