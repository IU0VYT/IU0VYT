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
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub><b>한국어</b></sub></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.4.0

이 버전에는 별도의 안정 릴리스가 없었던 4.3.0, 4.2.0, 4.1.0, 4.0.0의 새 기능도 들어 있습니다.

## 4.4.0의 새 기능

### 수정: 무전기가 없는데 «무전기 식별 진행 중» 창이 뜨는 문제
- 무전기가 없는 케이블에서도 창이 뜰 수 있었습니다: 선로 잡음과 보낸 명령의 에코가 무전기의 응답처럼 보였기 때문입니다.
- 이제는 **정말로 응답하는 무전기가 연결된 케이블에만** 뜹니다. 케이블 두 개에 무전기 한 대(COM6은 무전기 없음, COM7은 무전기 있음, 또는 그 반대)인 경우 무전기의 포트에만 뜹니다.
- 무전기가 없을 때 한 바퀴 시도도 더 빨라졌습니다.

### 위쪽 막대, 홈, 식별 탭
- 버튼 이름은 **«🔎 무전기 식별»**이며 오른쪽 «업데이트» 옆에 있습니다.
- 탭을 많이 열면 버튼이 저절로 좁아져 돋보기만 남습니다.
- 홈: 4.2의 상자 설명이 돌아왔습니다. 아래쪽의 저작권, 웹사이트, 소스 코드는 1280×720에서도 보입니다.
- 식별 탭: 짧고 분명한 설명.

### Retevis 펌웨어(펌웨어 업데이트 지원 43개 모델)
- **Retevis RA79 V3**(그리고 UV-K5 V3 / UV-K1, 부트로더 7.x), Retevis .bin 파일 사용(예: «RA79 V3.bin»). 무전기의 부트로더가 같은 계열일 때만 펌웨어를 보냅니다.
- **Retevis RA89**, Retevis .bin 파일 사용(예: «Eng_RA89_Upadat_20250524_B91.38.bin»). Retevis 프로그램 «RA89-update-software»의 프로토콜을 쓰며, 이 프로그램을 가상 무전기로 실행했습니다. 보내는 명령은 바이트 단위까지 같습니다.

### 기타
- 가상 무전기를 쓰는 테스트 때문에 Windows에서 테스트가 컴파일되지 않는 일이 더는 없습니다.

갤러리: 27개 제조사 386개 모델(채널 348, 설정 302, 펌웨어 업데이트 43).

## 4.3.0의 새 기능

### 자동 식별: 더 빠르게, 더 많은 무전기, 더 많은 포트
- 무전기 없이 한 바퀴 시도하는 데 약 **50초**(이전 135초).
- 프로그램이 **모든 케이블 포트를 계속 듣습니다**. 포트마다 따로, 그리고 모두 함께:
  - 꺼진 채 연결했다가 나중에 켠 무전기도 인식합니다;
  - **서로 다른 케이블의 여러 무전기**를 동시에 인식하고 데이터를 보여 주며 탭을 모두 엽니다.
- **이미 인식했던 무전기**(예: K67 Pro, Bajeton 8300, BJ-9000)를 먼저 시도하고 그다음 나머지를 시도합니다.
- 무전기 탭이 이미 열려 있으면 «무전기가 이미 식별됨!! 무전기 탭이 이미 열려 있음»이 나타납니다(여러 대일 때는 «무전기들이 이미 식별됨!! 무전기 탭들이 이미 열려 있음»).
- 식별 탭: 포트를 순서대로(COM6, COM7… COM10) «COM6 — USB-SERIAL CH340» 형식으로 표시; «새 무전기 자동 식별»은 굵게.

### 펌웨어 업데이트: 무전기 8종 추가(모두 41종)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630**, Radtel .kdhx 파일 사용(부트로더 «JJCC Bootloader Pro»). RT-470: 펌웨어 1.x는 구형 보드 전용, 2.x는 신형 보드 전용입니다. 프로그램이 이를 알려 주고 글로 확인을 받습니다.
- **TIDRADIO TD-H3, TD-H8과 Radtel RT-730**, 제조사 .BIN 파일 사용: 업데이트를 시작한 뒤 PTT를 누른 채 무전기를 켜세요.

### 기타
- **Bajeton BJ-9000**: BJ-8300처럼 960개 채널 전부를 구역 없는 단일 목록으로(«채널: 전체(960)»).
- 더 간결한 **홈**: 아래쪽에 버전, 저작권, 웹사이트, 소스 코드가 항상 보입니다.
- **갤러리**: 두꺼운 선으로 호환 모델과 비호환 모델을 나눕니다.
- **LEGGIMI**: 4.2.0에서 실수로 반복된 긴 부분을 없앴습니다.

갤러리: 27개 제조사 386개 모델(채널 348, 설정 302, 펌웨어 업데이트 41).

## 4.2.0의 새 기능

- 자동 식별 수정: «무전기 식별 진행 중» 창은 연결된 무전기가 응답할 때만 뜨고(케이블만 꽂고 시작할 때는 더 이상 뜨지 않음), 무전기를 인식했거나 찾지 못하면 사라집니다. 인식한 뒤에는 무전기의 데이터와 탭이 열리고, 기능을 끄지 않은 채 검색이 멈춥니다.
- 위쪽 막대의 «정보»와 «업데이트» 사이에 새 버튼 «🔎 새 무전기 자동 식별»: 같은 규칙으로 새 무전기를 찾으며, 자동 식별이 꺼져 있으면 회색입니다.
- 새 무전기 35종(36개 버전)과 제조사 8곳 추가: TYT TH-UV88 계열(TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), Radtel RT-490 계열(JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590 포함), Retevis RT21 계열 19종(Abbree AR-63 포함), Quansheng TK11.
- Radtel RT-590 / RT-590 Plus와 RT-600 / RT-600 Plus도 Radtel .bin 파일로 펌웨어 업데이트 가능.
- 악센트 문자나 중국어가 들어간 채널 이름을 읽을 때의 숨은 결함 몇 가지를 수정.
- 갤러리: 27개 제조사 386개 모델: 채널과 중계기 348, 설정 302, 펌웨어 업데이트 33.

## 4.1.0의 새 기능

- 자동 식별 수정: 어느 탭에서든 작동하며, 무전기를 빼거나 끄거나 바꾸면 검색이 저절로 다시 시작되어 다음에 연결한 무전기를 인식합니다(같은 무전기여도, 케이블을 꽂은 뒤에 켜도). 이전에는 처음 인식한 뒤로 다른 무전기를 더 찾지 않았습니다.
- 자동 검색 중에는 «무전기 식별 진행 중. 잠시 기다려 주세요.» 창이 «중지» 버튼과 함께 나타납니다.
- 새 무전기 14종과 제조사 3곳 추가: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495, RT-730, Hiroyasu HI-8811, Talkpod A36plus와 A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8(2세대와 3세대)과 TD-H3. 모두 기본, HAM, GMRS 버전.
- Quansheng UV-K5, UV-K5(8), UV-K6, UV-5R Plus의 펌웨어를 Quansheng .bin 파일로 업데이트하며, 펌웨어가 연결된 무전기 종류에 맞는지 확인합니다.
- 홈 화면과 웹사이트: «채널 관리 - 주파수 - 중계기 - 펌웨어»와 모든 제조사 이름을 80년대 비디오 게임 스타일의 테두리 안에.
- 갤러리: 19개 제조사 349개 모델: 채널과 중계기 311, 설정 265, 펌웨어 업데이트 31.

## 4.0.0의 새 기능

- 여러 제조사를 지원하는 프로그램: 갤러리에 16개 제조사 335개 모델, 케이블로 채널과 주파수를 다루는 모델 297개, 무전기 설정을 다루는 모델 251개.
- 새 무전기: Quansheng UV-K5, UV-K6, UV-5R Plus, UV-K1; Retevis 31개 모델(H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920, T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S, GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID.
- 자동 식별(식별 탭): 케이블을 꽂으면 무전기가 저절로 인식되고, 그 탭이 열리며 채널을 읽어 옵니다. 선택은 기억되며 같은 버튼으로 끌 수 있습니다.
- 제조사 메뉴와 제조사별 색이 있는 모델 갤러리.
- 무전기 탭의 «무전기에서 읽기»와 «무전기에 쓰기»를 파란색으로.
- 홈, 정보, 안내서, 웹사이트, 문서를 여러 제조사용 프로그램에 맞게 업데이트.
