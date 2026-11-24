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
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub><b>한국어</b></sub></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.47.0

### 새 기능
- **Retevis, 제조사 프로그램에서 알아낸 새 모델 24종**: B3B, B63B, RT22P, RT622P, RB666, RB666P(채널 메모리는 H777 새 버전과 같음); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669(T18 계열); RB65, RB66, RB66P(채널마다 주파수 하나와 톤 하나). 연결, 읽기, 쓰기는 처음부터 새로 작성했고 각 프로그램의 공장 데이터 파일로 확인했습니다. RB66과 RB66P, 그리고 RB669, RT68H, RT668H는 같은 코드로 응답하므로 올바른 카드를 직접 골라야 합니다.
- **Retevis 24개 모델의 설정**: 모두 241개 항목(스켈치, 송신 시간 제한, VOX, 스캔, 음성 안내, 배터리 절약, 비프, 측면 키…). 해당 모델의 프로그램이 보여 주는 항목만 제공합니다. 공장 파일 24개의 값은 모두 허용된 선택지입니다.
- **프로그래밍 비밀번호**: Retevis 24개 모델과 Retevis H777 새 버전에서는 CHIRP와 같이 읽기와 쓰기를 막지 않습니다(제조사 프로그램은 비밀번호가 사용자가 입력한 것과 다르면 멈춥니다). 설정의 한 항목으로 그대로 표시되며 보고, 바꾸고, 지울 수 있습니다. 길이와 문자는 제조사 프로그램과 같습니다: 6자(RB45, RB645, RB47, RB66은 3자, RB66P는 4자), 숫자와 영문자(RB65와 RB66은 숫자만). B3H, B63H, RB45, RB645, RB66, RB666은 제조사 프로그램에 비밀번호 입력란이 보이지 않으며, 경고가 이를 알려 줍니다.
- **지금까지 채널만 있던 무전기의 설정**: AnyTone 계열(AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2, DMR-6X2 PRO: 27개 항목, qdmr과 dmrconfig의 메모리 설명 기준); TYT MD-UV380, MD-UV390, MD-2017, MD-9600, Retevis RT3S, RT82(Baofeng DM-1701과 같은 항목); nicFW880 / nicFW950 펌웨어의 Radtel RT-880 및 RT-950 Pro(17개 항목, 펌웨어 제작자의 문서 기준); TID TD-UV68(TD-H8과 같은 항목); M7OCM 펌웨어의 Radtel RT-890(53개 항목).
- **테스트 버전**: 공개된 최신 안정 버전보다 번호가 높은 사본은 경고 막대(«테스트 버전(베타 테스트)…»)를 표시하고, 그 사본만 다음 테스트 버전을 알려 줍니다. 테스트 버전은 항상 GitHub에서 내려받습니다. 안정 버전 사용자에게는 테스트 버전이 보이지 않습니다.
- 갤러리: 54개 브랜드 697개 모델(채널 지원 647개, 설정 지원 596개).

### 참고
- Retevis 무전기 24종은 채널도 설정도 실제 장비로 시험하지 않았습니다: 먼저 읽기만 하고, 그다음 무해한 설정 하나만 바꿔 보세요. 무전기가 예상과 다르게 응답하면 읽기는 메시지와 함께 멈추고 무전기에는 아무것도 쓰지 않습니다.
- 프로그래밍 비밀번호가 이미 있어도 프로그램은 CHIRP처럼 읽기와 쓰기를 시도합니다: 실제 무전기로는 시험하지 않았습니다.
- RB66과 RB66P, 그리고 RB669, RT68H, RT668H는 같은 코드로 응답하므로 올바른 카드를 직접 골라야 합니다.
- 이 프로그램은 독립적이며 무전기 제조사와 관계가 없습니다.
- 전체 테스트 통과: 일반 부분, 구조 비교 915건과 CHIRP과 동일한 읽기 / 쓰기 380건.
