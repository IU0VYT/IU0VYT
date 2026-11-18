# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.45.0

### 새 기능
- **AnyTone 계열 DMR 무전기**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus와 BTECH DMR-6X2, DMR-6X2 PRO(이 두 모델은 «비호환»이었음). 무전기의 USB 케이블로 채널과 구역을 다룹니다. 메모리와 프로토콜은 자유 프로그램 dmrconfig에서 가져왔고 그 원래 코드로 확인했습니다. 채널, 구역, 스캔 목록만 씁니다.
- 순정 펌웨어의 **TYT MD-UV380, MD-UV390, MD-2017, MD-9600과 Retevis RT3S, RT82**: Baofeng DM-1701과 같은 메모리와 프로토콜.
- **Baofeng BF-666S**(BF-888S 드라이버)와 **BF-T6**(실험적, BF-T8 드라이버).
- «비호환»으로 남은 카드 UV-8D, BF-UV8, DM-1702, DM-1703에는 이제 이유가 적혀 있습니다.
- 갤러리: 54개 제조사 673개 모델(채널 623, 설정 556).

### 참고
- 새 무전기는 모두 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요.
- BTECH DMR-6X2 PRO와 Baofeng BF-T6는 실험적입니다.
- 이 프로그램은 독립적이며 무전기 제조사와 관계가 없습니다.
- 전체 테스트 통과: 일반 부분, 구조 비교 918건과 CHIRP과 동일한 읽기 / 쓰기 381건, AnyTone 계열은 dmrconfig 원래 코드와 비교.

---

이 버전에는 별도의 안정 릴리스가 없었던 4.24.0부터 4.44.0까지의 새 기능도 들어 있습니다. 최신 버전부터 정리했습니다.

## IU0VYT Flash 4.44.0

### 새 기능
- 채널만 지원하던 무전기 8종에 **설정** 추가: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro와 Bajeton DM-5000, 제조사 프로그램 기반.
- **RT-480과 RT-68**: 쓸 때 일반 설정 블록도 이제 무전기에 들어갑니다.
- **RT-4D, RT-6D, RT-1000 Pro**: 무전기 파일에 설정도 들어 있습니다(이전에 저장한 파일도 그대로 쓸 수 있음).

### 참고
- 새 설정은 모두 실제 무전기로 시험하지 않았습니다.
- RT-69 GPS: 지금은 제조사 프로그램에서 하나씩 확인한 항목만 있습니다.

## IU0VYT Flash 4.43.0

### 새 기능
- **오류 수정**: 무전기 33종(Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22와 HA1G, Kenwood TK-2140 / TK-3140 등)과 4.42의 «쌍둥이» 8종에서, 케이블로 읽은 뒤에는 설정이 나오지 않고 CHIRP 파일을 열 때만 나왔습니다.
- Radtel 5종에 **설정** 추가: RT-817A / RT-871A, RT-830, RT-850, RT-330, RT-400, 제조사 프로그램 기반.

### 참고
- 새 설정은 모두 실제 무전기로 시험하지 않았습니다.
- Radtel RT-830과 RT-850: 무전기마다 항목 하나를 제조사 프로그램이 쓰는 방식과 다르게 읽습니다. 여기서는 쓰기 방식을 따르고 그 항목에 경고를 붙였습니다.

## IU0VYT Flash 4.42.0

### 새 기능
- **홈 탭**: 이제 창 안에 모두 들어갑니다. 제조사 상자가 넓어졌고 페이지가 창 크기에 맞춰 저절로 조정됩니다.
- 채널만 지원하던 30개 모델에 **설정** 추가: Baofeng D 시리즈(BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro와 Radtel RT-13.

### 참고
- 새 설정은 모두 실제 무전기로 시험하지 않았습니다.
- Baofeng D 시리즈: 제조사 프로그램에서는 «음성 안내»와 «키 소리»가 읽기와 쓰기에서 서로 바뀌어 있습니다. 여기서는 쓰기 방식을 따르고 두 항목에 경고를 붙였습니다.

## IU0VYT Flash 4.41.0

### 새 기능
- **새 무전기**(CHIRP에 있던 휴대 무전기 중 빠져 있던 것): Yaesu VX-1R(채널과 설정), Icom IC-91AD와 IC-92AD(2밴드, D-STAR도), BTECH UV-PRO(Bluetooth로, 케이블 없이), 수신기 AOR AR8200(새 제조사).
- 채널만 지원하던 무전기에 **설정** 추가: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### 참고
- 새 무전기는 모두 실제로 시험하지 않았습니다: 먼저 읽기만 해 보세요.
- BTECH UV-PRO: CHIRP 드라이버 작성자가 시험한 펌웨어 버전(147)에서만 쓰기를 허용합니다.

## IU0VYT Flash 4.40.0

### 새 기능
- **탭이 이미 열려 있을 때 더 빠른 자동 검색**: 짧은 확인(이전에는 Baofeng K67 Pro만)이 이제 Icom, Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, Quansheng UV-K5와 파생 모델, Bajeton BJ-9000, BJ-7800, BJ-7700, BJ-8300, Radtel RT-900 / RT-950 / RT-470, 같은 메모리를 쓰는 무전기에도 적용됩니다. 나머지는 전체 읽기 그대로입니다.
- **새 제조사: Alinco.** 휴대용 DJ-596, DJ-175, DJ-G7EG, DJ-G7T; 차량용 DR-135T, DR-235T, DR-435T, DR-03T, DR-06T. 채널만. 데이터를 되돌려 보내지 않는 케이블에서도 작동합니다.

### 참고
- Alinco는 모두 실제로 시험하지 않았습니다: 먼저 읽기만 해 보세요.
- 짧은 확인은 가상 무전기로 시험했습니다. 확인 뒤에 무전기가 정상 작동으로 돌아오지 않으면 알려 주세요.

## IU0VYT Flash 4.39.0

### 새 기능
- **새 Icom 무전기 4종**: ID-80H(D-STAR 휴대용), ID-880H(D-STAR 차량용), IC-W32A와 IC-W32E(두 부분으로 나뉜 2밴드 휴대용: 0부터 99는 VHF, 100부터 199는 UHF 채널). 채널만.

### 참고
- 새 무전기는 모두 실제로 시험하지 않았습니다: 먼저 읽기만 해 보세요.

## IU0VYT Flash 4.38.0

### 수정
- **식별**: Bajeton BJ-9000이 특히 처음 식별할 때 가끔 멈췄습니다. 4.35처럼 첫 인사 전에 1초 동안 포트를 다시 듣게 했으며, 그 케이블에서 이미 인식된 무전기도 마찬가지입니다. 4.36의 다른 개선 사항은 그대로입니다.

### 참고
- 이 버전에는 새 무전기가 없습니다.

## IU0VYT Flash 4.37.0

### 새 기능
- **새 Yaesu 휴대 무전기 8종**: VX-5R과 VX-7R(채널); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR, VX-8GE(채널과 설정; VX-8은 APRS도).
- 복사는 무전기에서 시작합니다(«CLONE» 모드): 탭은 직접 고르며 안내는 케이블 창에 있습니다.

### 참고
- 새 무전기는 모두 실제로 시험하지 않았습니다: 먼저 읽기만 해 보세요.
- 하지 않은 것: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### 수정
- **«식별»과 자동 식별**: 프로그램이 각 무전기를 어느 케이블에서 인식했는지 기억하고 그것부터 시작합니다. 첫 검색에서 듣기와 준비에 시간을 낭비하지 않으며, 케이블을 꽂은 뒤에 켠 무전기는 더 자주 다시 시도합니다.
- «무전기 식별 진행 중» 창은 무전기를 읽는 동안에만 모델과 진행 상황과 함께 나타납니다. 무전기가 연결되어 있지만 꺼져 있으면 켜라고 알려 주고, 켜자마자 인식합니다.
- 탭이 이미 열린 Baofeng K67 Pro: 메모리 전체를 다시 읽지 않고 몇 초 만에 인식.

### 새 기능
- «식별» 탭: **«식별 진단 저장»** 버튼(최근 검색의 시간과 첫 바이트를 텍스트 파일로 저장. 식별이 느리거나 틀리면 작성자에게 보낼 수 있으며, 채널과 개인 정보는 들어 있지 않음).

### 참고
- 이 버전에는 새 무전기가 없습니다.

## IU0VYT Flash 4.35.0

### 수정
- **«식별»과 자동 식별**: 4.31의 «빠른 한 바퀴»를 없앴습니다. 실제 무전기에서는 무전기가 프로그래밍 상태로 멈추고 검색이 느려질 수 있었습니다. 대기 시간은 예전대로 돌아왔습니다.
- 검색은 어느 PC에서든 같은 무전기 그룹을 시도합니다(이전에는 시작할 때마다 달라졌음).
- 무전기가 없는 케이블의 에코를 더는 무전기로 착각하지 않습니다.

### 참고
- 이 버전에는 새 무전기가 없습니다.

## IU0VYT Flash 4.34.0

### 새 기능
- **새 무전기**: Yaesu VX-170과 VX-177(휴대용), FT-7800 / FT-7900과 FT-8900(차량용), 채널만, CHIRP 드라이버 사용: 읽기, 쓰기, 무전기와 주고받는 바이트가 CHIRP과 동일.

### 참고
- 복사는 무전기에서 시작합니다(«CLONE» 모드): 탭을 직접 골라야 하며 «식별»은 이 무전기들을 시도하지 않습니다.

## IU0VYT Flash 4.33.0

### 새 기능
- **새 무전기**: Icom IC-T7H, IC-T8A, IC-Q7A, IC-P7(채널), CHIRP 드라이버 사용: 읽기, 쓰기, 무전기와 주고받는 바이트가 CHIRP과 동일.

### 참고
- IC-Q7A와 IC-P7: 모드(FM, 광대역 FM, AM)는 주파수를 따릅니다. 채널의 «P» 건너뛰기와 튜닝 단계는 유지합니다.

## IU0VYT Flash 4.32.0

### 새 기능
- **새 무전기**: FRS 버전 Baofeng T22, TIDRADIO TD-M8S, TD-V2, TD-H4, TD-H11(채널과 설정), 제조사 프로그램 기반.
- **«식별»**은 BF-888S처럼 자신을 밝히는 새 무전기도 검색 시간을 늘리지 않고 제안합니다: 그 탭을 열 때만 해당 드라이버로 무전기를 읽습니다.

### 수정
- 2026년 TIDRADIO(TD-H3 Plus / SE, TD-H7 Plus, 4세대 TD-H8, TD-H9, TD-638UV): 중간 출력은 실제로 그것이 있는 모델에서만 제공하며 GMRS 버전에서는 절대 제공하지 않습니다.
- Baofeng T22 PMR446: 연결 방식을 제조사 프로그램과 같게.
- Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS, TIDRADIO TD-M2 Pro의 안내: «식별»은 이 무전기들을 시도합니다. 반대로 적혀 있던 문장을 없앴습니다.

## IU0VYT Flash 4.31.0

### 새 기능
- **«식별»이 훨씬 빨라졌습니다**: 먼저 모든 무전기를 빠르게 한 바퀴 돌고, 아무것도 찾지 못했을 때만 예전처럼 전체를 한 바퀴 돕니다. 목록 끝에 있는 무전기도 거의 2분 대신 30초 안에 찾습니다.
- **«식별»이 이제 최근 버전에서 추가한 무전기도 시도합니다**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G, 화면 없는 TIDRADIO.

### 수정
- Retevis RB39 / RB639: 다른 나라용 탭을 쓰면 무전기를 끌 때까지 프로그래밍 상태로 남아 있었습니다.

## IU0VYT Flash 4.30.0

### 새 기능
- TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **4세대 TD-H8**, **TD-H9**, **TD-638UV**의 **일반 설정**: 프로그래밍 키, 스퀠치, VOX, 송신 제한 시간, 화면, 언어, 켤 때 메시지 등.
- **새 무전기(채널과 설정)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR**, **TD-M2 Pro**.

### 참고
- 새 무전기는 제조사 프로그램에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. «식별»은 이들을 서로 구별하지 못합니다: 탭을 직접 고르세요.

## IU0VYT Flash 4.29.0

### 새 기능
- **새 무전기(채널만)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **4세대 TD-H8**, **TD-H9**, **TD-638UV**, 각각 기본, HAM, GMRS 버전.
- TD-H3 Plus는 자신을 TD-H3라고 밝힙니다: 출력이 3단계이므로 «TD-H3 Plus» 탭을 쓰세요.

### 참고
- 이 무전기들은 제조사의 2026년 프로그램(«TIDRadioCPS»)에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. 일반 설정은 아직 다루지 않습니다.

## IU0VYT Flash 4.28.0

### 새 기능
- **새 무전기**(Radioddity 차량용): **UB15-G**(GMRS 15 W), **DB35-G**(GMRS 35 W), **DB35-UV**(VHF / UHF 35 W), **DB40-G**(GMRS 40 W, 199채널), 채널과 일반 설정. UB15-G는 펌웨어에 따라 탭이 두 개입니다: V1.10 이전(이름 없음)과 V1.10(이름 있음).
- **DB40-G**는 CHIRP에도 있습니다: 여기서도 같은 바이트와 같은 `.img` 파일을 읽고 씁니다.

### 수정
- 4.26.0부터 Windows에서 테스트가 컴파일되지 않았습니다(테스트 파일 세 개가 Windows에서 컴파일되지 않는 가상 무전기를 썼기 때문). 프로그램 자체에는 영향이 없었습니다.

### 참고
- UB15-G, DB35-G, DB35-UV는 제조사 프로그램과 그 설정 파일에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. «식별»은 이 무전기들을 인식하지 못합니다: 탭을 직접 고르세요.

## IU0VYT Flash 4.27.0

### 새 기능
- **새 무전기**: **Retevis C6**(FRS / GMRS, 30채널), **Retevis C66**(UHF, 16채널), **Radioddity Echo-DE / Echo-IT**(PMR446, 16채널), 채널과 일반 설정.
- 쓰기 전에 무전기의 대역, 모델, 지역이 쓸 메모리와 같은지 확인합니다. 무전기의 비밀번호와 보정값은 절대 쓰지 않습니다.

### 수정
- **Retevis RB39 / RB39P**: 이제 모델 탭에서 «GMRS / FRS(미국)» 무전기에 들어 있습니다.

### 참고
- 새 무전기는 제조사 프로그램의 코드에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. «식별»은 이들을 인식하지 못합니다: 탭을 직접 고르세요.
- 무전기에 읽기 비밀번호가 설정되어 있으면 먼저 제조사 프로그램으로 해제하세요.

## IU0VYT Flash 4.26.0

### 새 기능
- **새 무전기**: **Retevis RB39 / RB39P**(FRS, 22채널)와 **RB639 / RB639P**(PMR446, 16채널), 채널과 일반 설정. 무전기가 어느 나라용인지 알려 주므로, 잘못된 탭이면 읽기가 멈추고 올바른 탭을 알려 줍니다.
- **Baofeng BF-T17 / BF-T18**(PMR446과 FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3**, **PR-T3**의 **새 설정**: 스퀠치, 송신 제한 시간, VOX, 음성 안내, 키 소리, 배터리 절약 등 제조사 프로그램의 항목.

### 수정
- 12개 언어의 변경 기록: 4.25.0 날짜 수정(2026년 11월 12일).

### 참고
- 새 무전기는 제조사 프로그램의 코드에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. «식별»은 이들을 인식하지 못합니다: 탭을 직접 고르세요.

## IU0VYT Flash 4.25.0

### 새 기능
- **새 펌웨어의 Bajeton BJ-8300**: 99채널씩 10개 구역, 모두 990채널인 무전기용 새 탭(제조사 프로그램 BT-11_CPS v1.2.6): 채널, 구역 이름, 설정. 기존 «Bajeton 8300» 탭(960채널, 64채널씩 15개 구역)은 그대로이며, 이제 새 펌웨어를 발견하면 **경고**합니다.
- **새 무전기**: **Radioddity GS-10B Pro**(999채널과 설정), **Radioddity FS-T3**와 **PR-T3**(채널).
- 990채널 **Bajeton BJ-9000**: 설정 5개 추가(프로그램 v1.3.3의 항목).

### 수정
- **Baofeng BF-T17 / BF-T18, GT-18, Radioddity PR-T5**: 고출력과 저출력이 서로 바뀌어 있었습니다. 이전 버전으로 이 무전기에 쓴 적이 있다면 다시 읽고 출력을 다시 고르세요.
- «메뉴 자동 나가기»(BJ-8300, BJ-9000, Radtel RT-950): 마지막 선택지는 55초가 아니라 60초입니다.

### 참고
- 새 무전기와 탭은 제조사 프로그램의 코드에서 가져왔고 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요.
- BJ-8300: 어느 탭을 쓸지는 무전기에서 구역 수(64채널 15개 또는 99채널 10개)를 보고 정하세요.

## IU0VYT Flash 4.24.0

### 새 기능
- **더 보기 쉬운 모델 탭**: 무전기 종류 표시가 페이지 가운데에 있고, 호환과 비호환 무전기를 나누는 선이 빨갛고 더 두껍습니다.
- **무전기 26종 추가**, 이미 있는 같은 무전기의 드라이버로 다룸:
  - 제조사 프로그램을 비교해 알아낸 것: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - CHIRP의 동등 모델 표에서: Radioddity GA-5S, Retevis RT5(출력 3단계), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+와 TS-T9+, Rugged RH5R와 RH5X, Ansoko A-5R와 A-8S, Arcshell AR-5, AR-6, AR-8, Greaval GV-8S와 GV-9S, Tenway TW-325, UV-82 Pro, UV-5R Pro.

### 참고
- 제조사 프로그램을 비교해 알아낸 6종은 아직 실제 무전기로 시험하지 않았습니다: 먼저 읽기만 해 보세요. 모델이 맞지 않으면 무전기는 응답하지 않으며 바뀌지도 않습니다.
