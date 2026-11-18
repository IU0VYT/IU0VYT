# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub のリリースページ（英語）、ダウンロードはこちら：[V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0（Microsoft Windows 10/11 用）

このバージョンには、テスト版としてのみ公開された 4.8.0、4.7.0、4.6.0 の新機能も含まれます。

## 4.9.0 の新機能

### 修正：Lanchonlh HG-UV98 と Quansheng TG-UV2+ の「識別」
- 「識別」の間、ポートは常に 8N1 で通信していました。今は無線機のタブからの読み込みと同じく、偶数パリティ（HG-UV98）と 2 ストップビット（TG-UV2+）を使います。この 2 機種が認識されるようになりました。

### 修正：HG-UV98 のパリティ
- 偶数パリティのとき、Windows のポートがパリティエラーも確認するようになりました。以前は確認をオンにした直後にオフに戻していました。

### フリーのファームウェアを入れた Quansheng UV-K5 / UV-K6 / UV-5R Plus
- egzumer、F4HWN、LOSEHU/LTS/UVE のファームウェア用の新しいタブ。
- F4HWN は armel による F4HWN のドライバー（CHIRP の外）を使います：出力 8 段階、送信ロック、スキャンリスト 3 つ、ファームウェアの設定。
- 各タブは自分のファームウェアだけを読み書きします。ファームウェアを更新した後、使うべきタブをプログラムが示します。

### 新しい無線機
- Wouxun（20 機種）：KG-UVD1P（KG-UV2D、KG-UV3D）、KG-UV6D、KG-816、KG-818、KG-805G、KG-UV8D、KG-UV8D Plus、KG-UV8E、KG-UV9D Plus、KG-UV9PX、KG-UV9GX、KG-UV9K、KG-UV9G Pro、KG-935G、KG-935G Plus、KG-935H、KG-UV8H、KG-XS20G、KG-XS20G Plus、KG-XS20H
- TYT TH-UV8000D/E、UVL-15W、TH-350（米国）
- Abbree AR-518、Ruyage UV58Plus、Radtel RT-880G、Radtel RT-890（Radtel のファームウェアとフリーの M7OCM ファームウェア）
- Retevis C2、HA1G、HA1UV、HA2、RB15、RB615、RB17P、RB28、RB628、RB48、RB648、RT23、RT26、RT76P、RT87、RT16、RB27B、RB27V
- BTECH GMRS-V1、GMRS-V2、MURS-V1、MURS-V2、FRS-A1
- AnyTone TERMN-8R、OBLTR-8R
- Puxing PX-888K、QYT KT-8R、Rugged RH5R-V2、Baojie BJ-UV55、MTC UV-5R-3、Radioddity MU-5、Zartek ZA-758、TIDRADIO TD-M11、TDXone TD-Q8A

CHIRP のドライバーから作った無線機は、チャンネルと設定を CHIRP との自動比較テストで確認しています。GMRS、FRS、MURS の無線機は米国向けで、イタリアではそれらの周波数での送信は認められていません。

## 4.8.0 の新機能

### .BTU ファームウェアの更新（Bluetooth なしの Radtel RT-910 と RT-900、Bluetooth なしの Bajeton BJ-7700）
- .BTU ファームウェアを「Radtel Bootload Tools」プログラムのプロトコルで更新します。ブートローダーは Baofeng UV-5RM と同じで、ハンドシェイクは「PROGRAMBT80U」です。
- 自動モード：無線機は普通に電源を入れた状態。手動モード：無線機の電源を切り、両方のサイドキーを押しながら、ディスプレイに「UPDATE」と出るまで電源を入れます。
- ブートローダーは機種を確認しません。そのためプログラムはファイル名から無線機と基板のバージョンを判断し、警告を表示して、必ず機種名の入力を求めます。
- ギャラリーでは RT-910（Bluetooth なし、.BTU）と RT-910B（Bluetooth あり、.anc）のカードが 2 枚に分かれました。BJ-7700 と Bluetooth なしの BJ-7700 も同様です。

### 修正：512 チャンネルのファームウェアを入れた Bluetooth なしの Radtel RT-900
- V1.09P までは 512 チャンネル、V1.14P からは 999 チャンネルです。
- チャンネル 513-999 は、変更したときだけ書き込むようになりました。その前に、その領域が空か、チャンネルが入っているかを無線機で確認します。

### 新しい無線機（CHIRP のドライバー）
- TYT TH-UV3R、TH-UV3R-25、TH-UVF1、TH-UVF8D
- KYD NC-630A、IP-620
- KSUN M6
- Puxing PX-777、PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24、RT24V、H777S、RT1、H777 V4（H777 の新バージョン）、RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A、FD-268B、FD-288A、FD-288B、FD-150A、FD-160A、FD-450A、FD-460A、FD-460UH

いずれもチャンネルと設定を CHIRP との自動比較テストで確認しています。

## 4.7.0 の新機能

### 修正：自動識別が止まる、またはとても遅い
- ファームウェア更新モードで電源を入れた無線機は、自分から定期的にメッセージを送ります。以前はそれがすべての機種への応答のように見え、検索に何分もかかっていました。今は試す前にポートを聞いてだまされないので、検索は 1 分以内に終わり、「識別」タブが状況を説明します。
- 検索の開始時に画面が固まらなくなりました。よく使われる無線機から先に試します。
- 前に閉じたタブに対して「無線機のタブはすでに開いています」と誤って表示される問題を修正。

### 同じ方法で扱う無線機
- 選択ウィンドウと「識別」タブに、考えられる無線機をすべて表示します（例：Radtel RT-950 Pro、Bajeton BJ-9000、Hiroyasu HI-UV850 Pro）。選ぶためのメニューと、無線機のどこでファームウェアのバージョンを見られるかの短い説明付きです。

### 修正：ファームウェア 0.23 までの Radtel RT-950 Pro / Bajeton BJ-9000
- チャンネルの「送信許可」ビットを、どちらのバージョンでも常に書き込むようにしました（Radtel のプログラム v1.2.x までと同じ。v1.3.3 はこのビットを無視します）。古いファームウェアでは、990 チャンネルのタブで変更したチャンネルが送信できないことがありました。このビットがなかったチャンネル 1-960 にも設定し直します。
- Bajeton BJ-9000 の実際のバージョンを確認：ファームウェア V0.22、V0.27、V0.29（Radtel と同一）、プログラム v1.1.1（960 チャンネル）と v1.3.3（990 チャンネル、Radtel のプログラム）。チャンネルと設定は Bajeton のプログラムのコードでも確認しました。

### ほかの無線機のファームウェア更新
- Radtel RT-950 Pro、Radtel RT-950、Hiroyasu HI-UV850 Pro（.BTF ファイル、BJ-9000 と同じブートローダー）。
- Radtel RT-920 と RT-910B（.anc ファイル、BJ-7800 と BJ-7700 と同じプロトコル）。

### 新しい無線機（CHIRP のドライバー）
- Radtel RT-620、Radioddity GA-510 V2、TYT TH-350、Retevis RT18、Retevis RB627B（PMR446）。ギャラリーには Hiroyasu HI-UV850 Pro も。
- SenhaiX 8800、Radioddity GS-5B、Cignus XTR-5、Anysecu AC-580 の設定にも対応。

## 4.6.0 の新機能

### 修正：新しいファームウェアの Radtel RT-950 Pro / Bajeton BJ-9000
- ファームウェア **0.24**（2025 年 12 月）から、無線機は **99 チャンネルずつ 10 ゾーン、合計 990 チャンネル**（または 990 の 1 つのリスト）になりました。**0.23** までは **64 チャンネルずつ 15 ゾーン、合計 960**（または 960 の 1 つのリスト）でした。プログラムは両方のバージョンに対応しました。
- バージョンは無線機から読み込んだデータで判断します（無線機自身は知らせません）。読み込むと正しいタブでチャンネルが開きます。データが足りない場合は、無線機でバージョンを確かめる方法を説明が示します。
- チャンネル 961-990 は、変更があったときだけ、無線機から読み直したうえで書き込みます。無線機にそれらがない場合（ファームウェア 0.23 まで）は何も書き込みません。
- チャンネルと設定は Radtel のプログラム v1.2.0（960 チャンネル）と v1.3.3（990 チャンネル）の元のコードで確認しました。

### 新規：Radtel RT-950（「Pro」なし）
- 960 チャンネル（64 ずつ 15 ゾーン）または 1 つのリスト、全般設定。「BT-9000_CPS」v1.3.0 プログラムの元のコードで確認しました。

### 修正：無線機の正確な名前
- 同じ方法で扱う別々の無線機（例：**Radtel RT-950 Pro** と **Bajeton BJ-9000**、**Baofeng K67 Pro** と **5RH Pro**、Radtel RT-910 と Bajeton BJ-7700）を、タブ、読み込み、書き込み、識別、バックアップファイルでそれぞれの名前で呼ぶようになりました。
- これらの無線機を初めて識別したとき、どれなのかをプログラムが尋ね、その後は覚えています。

### 修正：保存していないデータの警告
- 読み込み、書き込み、転送の後はチャンネルが無線機に入っています。プログラムを閉じるときの警告は、本当に変更があるときだけ表示されます。

### 新しい無線機（CHIRP のドライバー）
- SenhaiX 8800、Radioddity GS-5B、Cignus XTR-5、Anysecu AC-580、TID TD-UV68。
- Radtel RT-920：無線機と同じく、990 チャンネルを 99 ずつ 10 ゾーンに。

ギャラリー：44 メーカー 489 機種（チャンネル対応 451、設定対応 386、ファームウェア更新対応 52）。新しいメーカー：Wouxun、AnyTone、QYT、Rugged、Baojie、MTC、Zartek、Ruyage、TDXone。
