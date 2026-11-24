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
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub><b>日本語</b></sub></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0（Microsoft Windows 10/11 用）

### 新機能
- **Retevis、メーカーのプログラムから調べた新しい 24 機種**：B3B、B63B、RT22P、RT622P、RB666、RB666P（チャンネルメモリーは新バージョンの H777 と同じ）、RB45、RB645、RB47、B3H、B63H、RB647、RT68H、RT668H、NR10、NR610、RB66S、RB666S、RT68B、RT668B、RB669（T18 ファミリー）、RB65、RB66、RB66P（1 チャンネルに周波数 1 つとトーン 1 つ）。接続、読み込み、書き込みは一から書き直し、各プログラムの工場出荷データファイルで確認しました。RB66 と RB66P、および RB669、RT68H、RT668H は同じコードで応答するため、正しいカードは手動で選ぶ必要があります。
- **Retevis 24 機種の設定**：合計 241 項目（スケルチ、送信時間制限、VOX、スキャン、音声ガイド、バッテリーセーブ、ビープ、サイドキーなど）。その機種のプログラムが表示する項目だけを提供します。24 個の工場出荷ファイルの値はすべて許される選択肢でした。
- **プログラミング用パスワード**：Retevis の 24 機種と新バージョンの Retevis H777 では、CHIRP と同じく読み込みも書き込みも止めません（メーカーのプログラムは、パスワードがユーザーの入力と違うと停止します）。設定の 1 項目としてそのまま表示され、確認、変更、削除ができます。長さと使える文字はメーカーのプログラムと同じです：6 文字（RB45、RB645、RB47、RB66 は 3 文字、RB66P は 4 文字）、数字と英字（RB65 と RB66 は数字のみ）。B3H、B63H、RB45、RB645、RB66、RB666 はメーカーのプログラムにパスワードの入力欄が表示されないため、警告でその旨を知らせます。
- **これまでチャンネルだけだった無線機の設定**：AnyTone ファミリー（AT-D868UV、AT-D878UV、AT-D878UV II Plus、BTECH DMR-6X2、DMR-6X2 PRO：27 項目、qdmr と dmrconfig のメモリー記述による）、TYT MD-UV380、MD-UV390、MD-2017、MD-9600、Retevis RT3S、RT82（Baofeng DM-1701 と同じ項目）、nicFW880 / nicFW950 ファームウェアの Radtel RT-880 と RT-950 Pro（17 項目、作者の文書による）、TID TD-UV68（TD-H8 と同じ項目）、M7OCM ファームウェアの Radtel RT-890（53 項目）。
- **テスト版**：公開済みの最新の安定版より番号の大きいコピーは警告バー（「テスト版（ベータテスト）…」）を表示し、そのコピーだけが次のテスト版を案内します。テスト版は必ず GitHub からダウンロードされます。安定版を使っている人にテスト版は表示されません。
- ギャラリー：54 ブランド 697 機種（チャンネル対応 647、設定対応 596）。

### 注意
- Retevis の 24 機種は、チャンネルも設定も実機では未確認です。まず読み込みだけを行い、次に無害な設定を 1 つだけ変えてください。無線機が想定外の応答をした場合、読み込みはメッセージを出して止まり、無線機には何も書き込まれません。
- すでにプログラミング用パスワードがあっても、プログラムは CHIRP と同じく読み込みと書き込みを試みます。これは実機では未確認です。
- RB66 と RB66P、および RB669、RT68H、RT668H は同じコードで応答するため、正しいカードは手動で選ぶ必要があります。
- 本プログラムは独立したもので、無線機メーカーとは関係ありません。
- 全テストに合格：一般部分、915 件の構造比較と CHIRP と同一の 380 件の読み込み / 書き込み。
