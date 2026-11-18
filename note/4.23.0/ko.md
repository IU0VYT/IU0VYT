# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.23.0

### 새 기능
- **Kenwood TK-481과 TK-981**(900 MHz, 채널이 시스템으로 나뉨)과 차량용 **TK-780과 TK-880**: 채널과 설정, CHIRP 드라이버 사용.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: 이제 무전기 설정도.
- **Radtel RT-13**: 채널(22개, 이름 없음), 제조사 프로그램에서 가져옴.
- **Radtel RT-69 GPS**: 채널(999개, 이름 12자), 제조사 프로그램에서 가져옴. 쓰기 전에 각 채널이 무전기 대역 안에 있는지 확인합니다.
- 손상된 파일에 더 강해짐: Baofeng UV-B5 / UV-B6과 UV-3R의 톤을 읽을 때 생기던 비정상 종료 두 가지를 수정.
- 변환: 대상 무전기가 저장할 수 없는 채널은 변환을 멈추지 않고 경고와 함께 건너뜁니다(IJV, nicFW, OpenGD77, UV-KX, R3b0rn 펌웨어, Radtel RT-480 / RT-68, RT-830 / RT-850).
- 갤러리: 47개 제조사 569개 모델(채널 515, 설정 433).

### 참고
- TK-481 / TK-981: 무전기에 이미 있는 시스템만 씁니다. 공유 자원(트렁킹) 시스템은 바꾸지 않습니다.
- RT-13과 RT-69 GPS: 무전기에 비밀번호가 설정되어 있으면 먼저 Radtel 프로그램으로 해제하세요. 제조사 프로그램에서 가져왔고 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보세요.
- 이 프로그램은 독립적이며 무전기 제조사와 관계가 없습니다.

---

이 버전에는 시험판으로만 공개되었던 4.13.0부터 4.22.0까지의 새 기능도 들어 있습니다. 최신 버전부터 정리했습니다.

## IU0VYT Flash 4.22.0

### 새 기능
- **Radtel RT-480과 RT-68**: 채널과 이름(256개 채널, 0부터 255까지 번호, 이름 12자), 제조사 프로그램에서 가져옴. 주파수, 이름, 출력, 대역폭, 스캔, 톤을 바꿀 수 있고 채널의 다른 항목은 그대로입니다.
- **Radtel RT-830과 RT-850**: 채널(128개, 이름 없음), 제조사 프로그램에서 가져옴. 주파수, 출력, 대역폭, 스캔, 톤을 바꿀 수 있습니다.
- **Radtel RT-20M**: Ruyage UV58Plus 탭으로 엽니다(iRadio UV-5118plus와 같은 메모리).

### 참고
- RT-480 / RT-68: 채널과 이름만 씁니다. 무전기의 설정, VFO, DTMF는 건드리지 않습니다. 주파수는 1.25 kHz 단위이며 모든 채널에 송신 주파수가 있어야 합니다.
- RT-830 / RT-850: 설정과 VFO는 원래대로 되돌립니다. 무전기 비밀번호는 절대 쓰지 않습니다. 비밀번호가 설정되어 있으면 먼저 Radtel 프로그램으로 해제하세요.
- 모두 제조사 프로그램에서 가져왔고 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보세요.

## IU0VYT Flash 4.21.0

### 새 기능
- **Radtel RT-5D(DMR)**: 채널(64개씩 16개 구역, 이름 12자), 제조사 프로그램에서 가져옴. 주파수, 이름, 출력, 대역폭, 스캔, 톤을 바꿀 수 있고 채널의 다른 항목은 그대로입니다.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E**와 차량용 **TK-7180, TK-8180**(E 포함): 채널, CHIRP 드라이버 사용. 목록에는 처음 32개 구역마다 250개 자리가 있으며, 무전기에 이미 있는 구역만 써서 모두 512개 채널입니다.

### 참고
- RT-5D: 읽기는 메모리 전체를 가져오고, 쓰기는 Radtel 프로그램처럼 파일에 있는 그대로 전부 되돌려 보냅니다. 쓰기 전에 새로 읽으세요: 오래된 파일은 그때의 연락처와 설정을 되돌려 놓습니다. 가상 무전기로만 시험했습니다.
- TK-x180: 채널만, 설정은 없음. 구역을 만들거나 지우지 않습니다. 무전기에 비표준(트렁킹) 구역이 있으면 읽기를 거부합니다.

## IU0VYT Flash 4.20.0

### 새 기능
- **Radtel RT-4D와 RT-6D(DMR), RT-1000 Pro**: 채널(1024개, 이름 16자)과 구역(250개), 제조사 프로그램에서 가져옴. 주파수, 이름, 출력, 대역폭, 톤, 그리고 RT-4D와 RT-6D에서는 스캔을 바꿀 수 있고 채널의 다른 항목은 그대로입니다. 채널과 구역만 쓰며 연락처, 그룹, 메시지, FM 라디오, 설정은 건드리지 않습니다.

### 참고
- 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보고 문제가 있으면 알려 주세요.
- «식별»은 연결 방식이 같은 다른 Radtel, Bajeton 무전기와 이 무전기를 구별하지 못합니다: 탭을 직접 고르세요.
- RT-1000 Pro: 스캔 제외는 다루지 않습니다(무전기에 있는 그대로).

## IU0VYT Flash 4.19.0

### 새 기능
- **Radtel RT-817A / RT-871A**: 채널(1024개, 이름 16자), 제조사 프로그램에서 가져옴. RT-880G와 연결 방식과 채널 형식이 같지만 구역은 없습니다. Radtel 프로그램이 쓰는 영역(설정, 채널, FM 라디오)만 읽고 쓰며 로고는 건드리지 않습니다.

### 참고
- 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보고 문제가 있으면 알려 주세요.
- «식별»은 RT-817A / RT-871A와 RT-880G를 구별하지 못합니다: 탭을 직접 고르세요.

## IU0VYT Flash 4.18.0

### 새 기능
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: 채널이 구역(그룹)으로 나뉘는 업무용 휴대 무전기, 채널과 설정. 목록에 메모리의 250개 자리가 보이며, 새 채널은 바로 앞 채널의 구역에 들어갑니다.
- **fagci의 R3b0rn 펌웨어를 쓰는 Quansheng UV-K5 / UV-K6**: 채널(원래 메모리로 298개, 더 큰 메모리로 최대 1024개). 설정, VFO, 프리셋, SSB 보정은 절대 쓰지 않습니다.
- **Radtel RT-330과 RT-400**: 채널, 제조사 프로그램에서 가져옴.

### 참고
- TK 휴대 무전기용 CHIRP 드라이버는 실험적이라고 밝혀져 있습니다. 프로그램에서 구역을 만들거나 지우지 않습니다.
- R3b0rn 펌웨어, RT-330, RT-400은 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보고 문제가 있으면 알려 주세요.
- R3b0rn 펌웨어에서는 이름 없는 채널이 주파수를 이름으로 갖습니다(이 펌웨어에는 이름 없는 채널이 없습니다).
- RT-400 탭은 RT-470, RT-630, RT-495, HI-8811로 자신을 밝히는 무전기를 거부합니다: RT-400에만 쓰세요.

## IU0VYT Flash 4.17.0

### 새 기능
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: 무전기를 평소처럼 켠 상태(9600보드)에서 채널을 하나씩 읽고 씁니다. 쓸 때는 바뀐 채널만 보냅니다.
- **Kenwood TK 업무용 휴대 무전기**: 채널과 설정.
  - 60 시리즈: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - 60G 시리즈: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: «Radtel RT-880G» 탭을 씁니다(같은 연결 방식, 같은 채널 메모리).

### 참고
- Kenwood TH-F6/F7, TH-G71, TH-D7, TH-K2는 CHIRP 드라이버의 명령을 쓰며 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보고 문제가 있으면 알려 주세요.
- TK 휴대 무전기용 CHIRP 드라이버는 실험적이라고 밝혀져 있습니다.
- RT-860은 제조사 프로그램과 비교해 만든 것으로 실제 무전기에서는 시험하지 않았습니다.

## IU0VYT Flash 4.16.0

### 새 기능
- **OpenGD77**: 자유 펌웨어 OpenGD77을 쓰는 DMR 무전기의 채널.
  - Radioddity GD-77과 GD-77S, Baofeng DM-1801과 RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600, MD-2017, Retevis RT3S와 RT84.
  - 무전기는 USB 직렬 포트로 보입니다. 바뀐 채널과 구역만 쓰며, 설정, 연락처, 그룹, 보정값은 절대 쓰지 않습니다.
- **Quansheng UV-K5, 다른 자유 펌웨어**:
  - kamilsss655의 **NUNU**: 채널과 설정.
  - joaquim.org의 **UV-KX**: 채널과 설정. 보정값은 절대 쓰지 않습니다.
  - «OSFW» 탭은 대역을 넓힌 개조 순정 펌웨어(uvmod-kitchen, Tunas1337의 모음)에도 쓸 수 있습니다.
- 순정 펌웨어의 GPS 없는 **Radtel RT-880**: «Radtel RT-880G» 탭을 씁니다.

### 참고
- OpenGD77은 자유 프로그램 qdmr에 설명된 프로토콜로 작성했고 가상 무전기로만 시험했습니다: 먼저 읽기만 해 보고, 백업을 보관하며, 문제가 있으면 알려 주세요.
- NUNU와 UV-KX는 각 작성자의 CHIRP용 드라이버와 바이트 단위로 비교했습니다.

## IU0VYT Flash 4.15.0

### 새 기능
- **D-STAR와 C4FM**: 아마추어 디지털 휴대 무전기의 채널을 케이블로 다루며 .img 파일은 CHIRP과 호환됩니다. «방식» 열에 «D-STAR» 또는 «C4FM»이 나옵니다.
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2(D-STAR)와 IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D(복사는 무전기에서 시작합니다: 안내는 케이블 창에 있습니다).
  - **Kenwood**(새 제조사, 모두 47곳) TH-D72, TH-D74, TH-D75, USB 케이블 사용.
  - IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D, TH-D72의 설정.
- **자유 펌웨어**:
  - TIDRADIO TD-H3의 **nicFW 2.x**: 채널과 주요 설정.
  - Radtel RT-880, RT-880G, RT-950 Pro의 **nicFW880 / nicFW950**: 채널.
  - Quansheng UV-K1과 UV-K5 V3의 **F4HWN Fusion 5.x**: 1024개 채널과 설정.
  - nicFW에서는 무전기 보정값을 절대 쓰지 않습니다.
- **갤러리의 Kenwood**: PMR446과 DMR/NXDN 휴대 무전기도 비호환 모델로 표시(대리점 전용 KPG 프로그램으로만 설정할 수 있음).

### 수정
- 설정: 한 번도 쓰지 않은 텍스트 칸은 «예상하지 않은 값» 대신 빈칸으로 읽습니다. 바뀌지 않은 값은 다시 쓰지 않습니다.
- Puxing PX-888K: 잘못된 값을 쓸 수 있던 설정 항목 6개를 없앴습니다(BOT/EOT 코드와 DTMF 시간).
- Yaesu VX-3: DTMF 코드에 공백을 더는 허용하지 않습니다(무전기에 공백이 없음).

### 참고
- 드라이버는 CHIRP 드라이버(GNU GPL), nicsure가 공개한 문서(nicFW), F4HWN 드라이버를 바탕으로 처음부터 작성했고 가상 무전기로 바이트 단위 비교를 했습니다. 이 버전의 실제 무전기는 아직 시험하지 않았습니다: 백업을 보관하고 문제가 있으면 알려 주세요.
- Icom, Yaesu, Kenwood는 펌웨어 업데이트가 없습니다(제조사 프로그램으로 합니다).

## IU0VYT Flash 4.14.0

### 새 기능
- **Yaesu와 Icom**: 새 제조사 두 곳(모두 46곳)과 무전기 13종 추가. 케이블로 채널을 다루며 .img 파일은 CHIRP과 호환됩니다.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R(SCU-35 케이블)과 FT-60R/E(무전기에서 시작하는 복사, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82(OPC-478 케이블).
  - IC-T70을 뺀 모두 설정 지원. 이 무전기들은 펌웨어 업데이트가 없습니다.
- **언어 2개 추가**: 말레이어(Bahasa Melayu)와 힌디어(हिन्दी), 모두 13개: 프로그램, 웹사이트, 안내서와 정보 파일, README, 보안 정책.
- Yaesu와 Icom으로의 **변환기와 «다른 무전기로 전달»**: 무전기에 없는 톤은 경고와 함께 없애고, 저장할 수 없는 채널은 경고와 함께 건너뜁니다. 전체 작업을 멈추지 않습니다.

### 수정
- 번역: 안내서의 굵은 글씨 다섯 단어가 모든 언어에서 이탈리아어로 남아 있었습니다. 한국어와 아랍어의 틀린 문장 두 개.
- Wouxun KG-UV8D: 내부 필드 이름을 지금의 CHIRP에 맞췄습니다(채널에는 영향 없음).

### 참고
- 드라이버는 CHIRP 드라이버(GNU GPL)를 바탕으로 처음부터 작성했고 가상 무전기로 CHIRP과 바이트 단위 비교를 했습니다. 실제 Yaesu, Icom 무전기는 아직 시험하지 않았습니다: 백업을 보관하고 문제가 있으면 알려 주세요.
- 말레이어와 힌디어 번역은 새것입니다: 알아보기 어려운 문장이 있으면 알려 주세요.

## IU0VYT Flash 4.13.1

### 새 기능
- Quansheng UV-K5, UV-K5 V3, UV-K1용 **IJV 펌웨어**(IU0IJV의 비순정 펌웨어, [ijvradio.com](https://www.ijvradio.com/)): 펌웨어 버전마다 하나씩 탭 다섯 개.
  - 기존 UV-K5(하드웨어 V1)의 **V3**, 200개 또는 999개 채널.
  - UV-K5 V3와 UV-K1의 **V4**, **V5(«FW5»)**, **V6**: 무전기에서 고른 뱅크의 500개 채널, 그룹과 뱅크 이름 포함.
  - 채널과 주요 설정(메뉴, 화면, VOX, 옆 키)을 읽고 씁니다. 쓰기 전에 무전기에서 같은 뱅크가 그대로 선택되어 있는지 확인하고 바뀐 부분만 씁니다. 보정값은 절대 건드리지 않습니다.
  - **V6(4.13.1의 새 기능)**: V6 데이터 연결, 4 KiB 섹터 단위로 쓰고 섹터마다 다시 읽어 확인. 무전기가 데이터 연결이 없다고 하면 RadioManager의 FLASH V6로 업데이트하세요.
- **펌웨어 업데이트**: IJV 버전이 든 .bin 파일을 알아보고, 채널에 쓸 탭을 알려 주며, V5와 V6에는 경고를 냅니다.
- 11개 언어 모두로 번역.

### 참고
- 프로토콜과 메모리는 IJV가 공개한 CHIRP 모듈과 RadioManager 4.3.2 분석에서 가져왔고, 코드는 처음부터 작성해 가상 무전기로 시험했습니다. IJV를 쓰는 실제 무전기는 아직 시험하지 않았습니다: 문제가 있으면 알려 주세요.
- IJV의 모듈, 항목, 라이선스, 펌웨어는 IJV의 RadioManager에 속합니다.

## IU0VYT Flash 4.13.0

### 새 기능
- Quansheng UV-K5, UV-K5 V3, UV-K1용 **IJV 펌웨어**(IU0IJV의 비순정 펌웨어, [ijvradio.com](https://www.ijvradio.com/)): 펌웨어 버전마다 하나씩 새 탭 네 개.
  - 기존 UV-K5(하드웨어 V1)의 **V3**, 200개 또는 999개 채널.
  - UV-K5 V3와 UV-K1의 **V4**와 **V5(«FW5»)**: 무전기에서 고른 뱅크의 500개 채널, 그룹과 뱅크 이름 포함.
  - 채널과 주요 설정(메뉴, 화면, VOX, 옆 키)을 읽고 씁니다. 쓰기 전에 무전기에서 같은 뱅크가 그대로 선택되어 있는지 확인하고 바뀐 블록만 씁니다. 보정값은 절대 건드리지 않습니다.
  - **V6**는 인식합니다: V6의 채널과 설정은 IJV의 RadioManager로 다룹니다.
- **펌웨어 업데이트**: IJV 버전이 든 .bin 파일을 알아보고, 채널에 쓸 탭을 알려 주며, V5와 V6에는 경고를 냅니다.
- 11개 언어 모두로 번역.

### 참고
- 프로토콜과 메모리는 IJV가 공개한 CHIRP 모듈과 RadioManager 4.3.2 분석에서 가져왔고, 코드는 처음부터 작성해 가상 무전기로 시험했습니다. IJV를 쓰는 실제 무전기는 아직 시험하지 않았습니다: 문제가 있으면 알려 주세요.
