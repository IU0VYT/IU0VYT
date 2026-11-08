# IU0VYT Flash

<p align="center">
🇮🇹 <a href="../../">Italiano</a> · 🇬🇧 <a href="../en/">English</a> · 🇫🇷 <a href="../fr/">Français</a> · 🇩🇪 <a href="../de/">Deutsch</a> · 🇪🇸 <a href="../es/">Español</a> · 🇵🇹 <a href="../pt/">Português</a> · 🇷🇺 <a href="../ru/">Русский</a> · 🇨🇳 <a href="../zh/">中文</a> · 🇯🇵 <a href="../ja/">日本語</a> · 🇰🇷 <b>한국어</b> · 🇦🇪 <a href="../ar/">العربية</a>
</p>

**버전 4.11.0** · Windows · 11개 언어의 인터페이스(italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية) · 실행 파일 하나, 설치 필요 없음

**44개 브랜드**(**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone)의 무전기를 쓰는 아마추어 무선사와 사용자를 위한 프로그램입니다. 프로그래밍 케이블로 채널, 주파수, 설정을 읽고 쓰며, 이탈리아 중계기 보관함을 관리하고, 지원하는 무전기의 펌웨어를 업데이트합니다.

🌐 프로그램 사이트: **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*이 파일은 README.md의 한국어 번역입니다. 이탈리아어 원문이 기준이 되는 글입니다.*

## 기능

- 451개 모델의 **채널과 주파수**(디지털 PMR과 DMR 무전기 포함, USB로도): 케이블로 읽기와 쓰기, CHIRP와 호환되는 `.img` 파일, Baofeng K67 Pro(`.xlc`)와 Bajeton 8300(`.dat`)의 원래 파일.
- 396개 모델의 **무전기 설정**(메뉴, 디스플레이, VOX, 스켈치, TOT, DTMF, 키…), 모든 항목의 설명과 민감한 항목의 확인 경고와 함께.
- **연결된 무전기 식별**: 프로그램이 스스로 알아보고 제조사, 모델, 종류, 펌웨어와 무전기가 알려 주는 다른 정보를 보여 줍니다(USB DMR 무전기는 DMR ID와 이름도). **자동 식별**(«식별» 탭의 버튼, 선택 기억)을 켜면 어느 탭에 있든 케이블만 연결하면 됩니다. 프로그램은 모든 케이블 포트를 함께 듣고, 케이블을 연결한 뒤에 켠 무전기와 서로 다른 케이블에 연결한 여러 무전기도 알아보며, 예전에 알아본 무전기와 가장 흔한 무전기를 먼저 시도하고, 펌웨어 업데이트 모드의 무전기에 속지 않으며, 읽은 채널과 함께 각 무전기의 탭을 엽니다. «무전기를 식별하는 중» 상자는 응답하는 무전기가 있는 케이블에만 나옵니다. 다른 무전기를 찾으려면 오른쪽 위의 **«🔎 무전기 식별»** 버튼이 있습니다.
- **다른 무전기로 보내기**: 한 무전기에서 체크한 채널을 제조사와 모델로 고른 다른 연결된 무전기에 바로 씁니다. 호환성 확인, 맞추기, 백업 사본, 확인과 함께.
- **범용 변환기**: 어떤 무전기나 파일(K67 `.xlc`, Bajeton `.dat`, CHIRP의 것도 포함한 `.img` 이미지, CHIRP CSV, Excel/CSV)의 채널을 다른 어떤 무전기의 형식으로, 그 규칙에 맞추어 바꿉니다.
- **11개 언어의 인터페이스**: «ℹ 정보» 옆의 국기 선택기로 프로그램 전체를 즉시 번역하며(아랍어는 오른쪽에서 왼쪽으로), 언어는 다시 시작해도 기억됩니다. 안내 파일의 번역은 `lingue` 폴더에 있습니다.
- «구역 1», «구역 2»… 대신 **구역 이름**(Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000과 BJ-7700, Hiroyasu HI-UV850 Pro, DMR 무전기 RD-5R, DM-1801, DM-32UV의 구역). 채널과 함께 케이블로 읽고 씁니다.
- 체크한 채널의 **일괄 편집**: 출력, 대역폭, 스캔, 서브톤, 시프트 또는 송신을 한꺼번에 바꾸며, 각 채널의 나머지는 그대로입니다.
- 44개 브랜드 **489개 모델의 갤러리**: 호환과 비호환으로, 종류(아마추어, DMR, GMRS/FRS, PMR446, 어린이용)와 브랜드별로 나뉘며, 검색, 필터, 공식 프로그램과 설명서 링크가 있습니다.
- [IK2ANE](https://www.ik2ane.it)의 최신 목록을 바탕으로 한 **중계기 보관함**: 필터와 무전기에 넣기, 쓰기 전 중계기와 무전기 사이 호환성 확인.
- 57개 모델의 **펌웨어 업데이트**(Quansheng UV-K5, UV-K6, UV-5R Plus, Quansheng과 Radtel의 .bin 파일로 Radtel RT-590과 RT-600, Radtel의 .kdhx 파일로 Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630, 제조사의 .BIN 파일로 TIDRADIO TD-H3, TD-H8과 Radtel RT-730, .bin 파일로 Retevis RA79 V3와 RA89, Retevis의 .hex 파일로 Retevis RT85, .BTF 파일로 Radtel RT-950 Pro와 RT-950, Bajeton BJ-9000, .anc 파일로 Radtel RT-920와 RT-910B, .BTU 파일로 블루투스 없는 Radtel RT-910과 RT-900, Bajeton BJ-7700, Radtel과 M7OCM의 .bin 파일로 Radtel RT-890, Baofeng의 .bin 파일로 USB를 통한 Baofeng DM-1701, .sgl 파일로 USB를 통한 Baofeng RD-5R와 DM-1801 포함), 파일 확인과 무전기 백업 사본과 함께. BTECH DMR-6X2와 DMR-6X2 PRO는 공식 프로그램으로 안내에 따른 업데이트.
- 쓰기 전마다 자동 **백업 사본**, 쓴 뒤에는 확인.
- 시작할 때마다 프로그램 **업데이트 확인**.

## 사용법

1. [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip)(최신 버전)을 내려받아 압축을 풉니다.
2. `IU0VYT Flash.exe`를 실행합니다(Windows 10/11. Windows의 WebView2 구성 요소를 씁니다).
3. **📻 모델** 탭에서 무전기를 고르고, 케이블을 연결한 뒤 **무전기에서 읽기**를 씁니다.

전체 안내는 `GUIDE.txt` 파일에 있습니다(이탈리아어 원문: `LEGGIMI.txt`).

## 빌드

Go 1.22 이상이 필요합니다. `sorgenti` 폴더에서:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

라이브러리는 `vendor` 폴더에 들어 있습니다.

## 주의

- 무전기 프로토콜은 상호 운용성을 위해 재구성했고 시뮬레이션 무전기로 확인했습니다: **백업 사본을 항상 보관하세요**. 프로그램은 어떠한 보증도 없이 제공됩니다.
- 무전기 제조사(Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO 등)와 **제휴하거나 그들이 보증하지 않은** 독립 프로그램입니다. 언급한 상표는 각 소유자의 것이며 호환되는 무전기를 나타내는 데만 씁니다.
- 자기 나라의 주파수 사용 규정을 지키세요.

## 감사의 말

- [CHIRP](https://chirpmyradio.com)와 그 작성자들: 무전기 드라이버 대부분이 그들의 작업에서 왔습니다.
- 이탈리아 중계기 목록을 모두에게 무료로 제공하는 [IK2ANE](https://www.ik2ane.it).

구성 요소의 작성자와 라이선스: `AUTHORS_AND_LICENSES.txt`.

## 라이선스

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

**GNU General Public License 버전 3 또는 그 이후 버전**(`GPL-3.0-or-later`)으로 배포하는 자유 소프트웨어: `LICENSE.txt`와 `LICENSE_NOTICE.txt`를 보세요.
