# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.5.0

### Retevis RT85 펌웨어(펌웨어 업데이트 지원 44개 모델)
- **Retevis RT85**: Retevis의 .hex 파일(예: «RT85-Firmware-V1.53.hex»)을 사용합니다. Retevis 프로그램 «RT85 10W_5W_UpData»의 프로토콜을 쓰며, 이 프로그램을 가상 무전기로 실행해 보낸 블록과 명령이 바이트 단위까지 같음을 확인했습니다.
- Retevis 절차: 무전기를 끄고 **PTT와 P1**을 누른 채 켭니다(녹색 표시등). 그런 다음 무전기 출력(Retevis 프로그램처럼 **5 W 또는 10 W**)을 고르고 확인을 위해 RT85를 입력합니다.

### 수정: Retevis RA89 업데이트
- 4.4는 인사 단계에서 무전기에게 잘못된 응답을 기다렸기 때문에 실제 RA89에서는 업데이트가 시작되지 않았습니다(무전기는 변경되지 않음). 이제 기대하는 응답과 대기 시간이 Retevis 프로그램과 같습니다.

### 포트, 식별 탭과 변환 탭
- 케이블 포트는 어떤 시스템에서든 항상 번호 순서입니다: COM1, COM2, COM3… COM10. 자동 식별도 첫 번째 포트부터 시작합니다.
- 식별 탭: «예전에 식별한 무전기는 더 빨리 인식됩니다».
- 변환 탭: «Baofeng, Bajeton 파일, 다른 모든 제조사 무전기의 .img(CHIRP로 저장한 것 포함), CHIRP CSV, Excel 표 또는 CSV를 고르세요».

### 웹사이트
- 내려받기 버튼 아래의 릴리스 노트 문구를 없앴습니다.
- 제목이 더 위로 올라가 노트북 화면에서도 «PRESS START» 아래 버전 번호가 모두 보입니다.

갤러리: 27개 제조사 386개 모델(채널 348, 설정 302, 펌웨어 업데이트 44).
