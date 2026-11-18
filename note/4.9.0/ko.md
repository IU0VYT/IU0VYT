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
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub><b>한국어</b></sub></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.9.0

이 버전에는 시험판으로만 공개되었던 4.8.0, 4.7.0, 4.6.0의 새 기능도 들어 있습니다.

## 4.9.0의 새 기능

### 수정: Lanchonlh HG-UV98과 Quansheng TG-UV2+ 식별
- «식별» 중에는 포트가 늘 8N1로 통신했습니다. 이제 무전기 탭에서 읽을 때처럼 짝수 패리티(HG-UV98)와 정지 비트 2개(TG-UV2+)를 씁니다. 두 무전기가 이제 인식됩니다.

### 수정: HG-UV98 패리티
- 짝수 패리티일 때 Windows 포트가 이제 패리티 오류도 검사합니다. 이전에는 검사를 켰다가 곧바로 껐습니다.

### 자유 펌웨어를 쓰는 Quansheng UV-K5 / UV-K6 / UV-5R Plus
- egzumer, F4HWN, LOSEHU/LTS/UVE 펌웨어용 새 탭.
- F4HWN은 armel의 F4HWN 드라이버(CHIRP 밖)를 씁니다: 출력 8단계, 송신 잠금, 스캔 목록 3개, 펌웨어 설정.
- 각 탭은 자기 펌웨어만 읽고 씁니다. 펌웨어를 업데이트한 뒤에는 어느 탭을 써야 하는지 프로그램이 알려 줍니다.

### 새 무전기
- Wouxun(20개 모델): KG-UVD1P(KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350(미국)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890(Radtel 펌웨어와 자유 펌웨어 M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

CHIRP 드라이버에서 가져온 무전기는 채널과 설정을 CHIRP과 비교하는 자동 테스트로 확인했습니다. GMRS, FRS, MURS 무전기는 미국용입니다: 이탈리아에서는 그 주파수로 송신할 수 없습니다.

## 4.8.0의 새 기능

### .BTU 펌웨어 업데이트(Bluetooth 없는 Radtel RT-910과 RT-900, Bluetooth 없는 Bajeton BJ-7700)
- .BTU 펌웨어를 «Radtel Bootload Tools» 프로그램의 프로토콜로 업데이트합니다. 부트로더는 Baofeng UV-5RM과 같고 인사 문자열은 «PROGRAMBT80U»입니다.
- 자동 모드: 무전기를 평소처럼 켠 상태. 수동 모드: 무전기를 끄고 양쪽 옆 키를 누른 채 켜서 화면에 «UPDATE»가 나올 때까지.
- 부트로더는 모델을 확인하지 않습니다. 그래서 프로그램이 파일 이름으로 무전기와 하드웨어 버전을 알아내고, 경고를 보여 주며, 항상 모델명을 입력하게 합니다.
- 갤러리에서 RT-910(Bluetooth 없음, .BTU)과 RT-910B(Bluetooth 있음, .anc)가 이제 카드 두 장으로 나뉩니다. BJ-7700과 Bluetooth 없는 BJ-7700도 마찬가지입니다.

### 수정: 512채널 펌웨어를 쓰는 Bluetooth 없는 Radtel RT-900
- V1.09P까지는 512채널, V1.14P부터는 999채널입니다.
- 이제 채널 513-999는 수정했을 때만 씁니다. 그 전에 무전기에서 그 영역이 비어 있거나 채널이 들어 있는지 확인합니다.

### 새 무전기(CHIRP 드라이버)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4(H777의 새 버전), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

모두 채널과 설정을 CHIRP과 비교하는 자동 테스트로 확인했습니다.

## 4.7.0의 새 기능

### 수정: 자동 식별이 멈추거나 매우 느린 문제
- 펌웨어 업데이트 모드로 켠 무전기는 스스로 주기적인 메시지를 보냅니다: 이전에는 그것이 모든 모델에 응답하는 것처럼 보여 검색이 몇 분씩 걸렸습니다. 이제 프로그램이 시도하기 전에 포트를 들어 보고 속지 않으므로, 검색이 1분 안에 끝나고 식별 탭이 무슨 일이 일어나는지 설명합니다.
- 검색이 시작될 때 화면이 더는 멈추지 않으며, 많이 쓰이는 무전기부터 시도합니다.
- 앞서 닫은 탭에 대해 «무전기 탭이 이미 열려 있음»이 잘못 뜨는 문제를 수정.

### 같은 방식으로 다루는 무전기
- 선택 창과 식별 탭에 가능한 무전기를 모두 보여 줍니다(예: Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro). 고르는 메뉴와, 무전기 어디에서 펌웨어 버전을 볼 수 있는지 짧은 안내가 있습니다.

### 수정: 펌웨어 0.23까지의 Radtel RT-950 Pro / Bajeton BJ-9000
- 채널의 «송신 허용» 비트를 두 버전 모두에서 항상 씁니다(Radtel 프로그램 v1.2.x까지와 같음; v1.3.3은 이 비트를 무시). 옛 펌웨어에서는 990채널 탭에서 고친 채널이 송신하지 못할 수 있었습니다. 이 비트가 없던 채널 1-960에도 다시 넣습니다.
- Bajeton BJ-9000의 실제 버전 확인: 펌웨어 V0.22, V0.27, V0.29(Radtel과 동일), 프로그램 v1.1.1(960채널)과 v1.3.3(990채널, Radtel 프로그램). 채널과 설정은 Bajeton 프로그램의 코드로도 확인했습니다.

### 다른 무전기의 펌웨어 업데이트
- Radtel RT-950 Pro, Radtel RT-950, Hiroyasu HI-UV850 Pro(.BTF 파일, BJ-9000과 같은 부트로더).
- Radtel RT-920과 RT-910B(.anc 파일, BJ-7800, BJ-7700과 같은 프로토콜).

### 새 무전기(CHIRP 드라이버)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B(PMR446); 갤러리에는 Hiroyasu HI-UV850 Pro도.
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580의 설정도.

## 4.6.0의 새 기능

### 수정: 최신 펌웨어의 Radtel RT-950 Pro / Bajeton BJ-9000
- 펌웨어 **0.24**(2025년 12월)부터 무전기는 **99채널씩 10개 구역, 모두 990채널**(또는 990개 단일 목록)을 가집니다. **0.23**까지는 **64채널씩 15개 구역, 모두 960채널**(또는 960개 단일 목록)이었습니다. 프로그램이 이제 두 버전을 모두 다룹니다.
- 버전은 무전기에서 읽은 데이터로 알아냅니다(무전기가 직접 알려 주지 않음): 읽으면 올바른 탭에 채널이 열립니다. 데이터가 부족하면 무전기에서 버전을 확인하는 방법을 안내합니다.
- 채널 961-990은 바뀌었을 때만, 무전기에서 다시 읽은 뒤에 씁니다. 무전기에 그 채널이 없으면(펌웨어 0.23까지) 아무것도 쓰지 않습니다.
- 채널과 설정은 Radtel 프로그램 v1.2.0(960채널)과 v1.3.3(990채널)의 원래 코드로 확인했습니다.

### 새로 추가: Radtel RT-950(«Pro» 없음)
- 960채널(64채널씩 15개 구역) 또는 단일 목록, 일반 설정. «BT-9000_CPS» v1.3.0 프로그램의 원래 코드로 확인했습니다.

### 수정: 무전기의 정확한 이름
- 같은 방식으로 다루는 서로 다른 무전기(예: **Radtel RT-950 Pro**와 **Bajeton BJ-9000**, **Baofeng K67 Pro**와 **5RH Pro**, Radtel RT-910과 Bajeton BJ-7700)를 이제 탭, 읽기, 쓰기, 식별, 백업 파일에서 각자의 이름으로 부릅니다.
- 이 무전기 중 하나를 처음 식별하면 어떤 무전기인지 프로그램이 묻고 그 뒤로 기억합니다.

### 수정: 저장하지 않은 데이터 경고
- 읽기, 쓰기, 전달 뒤에는 채널이 무전기에 들어 있습니다: 프로그램을 닫을 때 실제로 바뀐 것이 있을 때만 경고가 뜹니다.

### 새 무전기(CHIRP 드라이버)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: 무전기와 같이 990채널을 99채널씩 10개 구역으로.

갤러리: 44개 제조사 489개 모델(채널 451, 설정 386, 펌웨어 업데이트 52). 새 제조사: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
