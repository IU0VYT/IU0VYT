# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.11.0

### 구역 이름
- «구역 1», «구역 2»… 대신 제조사 프로그램처럼 구역마다 이름을 붙일 수 있습니다: «✎ 구역 이름…» 버튼 또는 구역을 두 번 클릭. 구역 막대와 아래 줄에 «번호 · 이름»이 표시됩니다.
- 대상: Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro(펌웨어 0.23까지 및 0.24부터), Radtel RT-950, Radtel RT-920, Bluetooth 없는 Radtel RT-910 / Bajeton BJ-7700, Baofeng DMR 무전기 RD-5R, DM-1801, DM-32UV의 구역.
- 이름은 채널과 함께 케이블로 읽고 쓰며 .img 파일에 저장됩니다. 길이와 문자는 제조사 프로그램과 같습니다(RT-950 Pro는 중국어를 포함해 12자, RT-920은 10자, DMR은 16자).
- 안전: 이름 영역은 이름이 바뀌었을 때만, 무전기에서 다시 읽은 뒤에 씁니다. 그 자리에 이름이 아닌 데이터가 있으면 아무것도 쓰지 않습니다.

### 일괄 편집
- «✎ 일괄 편집…»(오른쪽 클릭으로도 가능)은 체크한(또는 선택한) 채널의 출력, 대역폭, 스캔, 송신 톤, 수신 톤, 시프트, 송신을 한꺼번에 바꿉니다. «바꾸지 않음»으로 둔 항목은 그대로이며 Ctrl+Z로 모두 되돌릴 수 있습니다.

### 언어 2개 추가(모두 11개)
- 한국어, 대한민국 국기.
- 아랍어, 아랍에미리트 국기와 오른쪽에서 왼쪽으로 쓰는 페이지(주파수와 숫자는 왼쪽에서 오른쪽 그대로).
- `lingue` 폴더의 모든 정보 파일을 한국어와 아랍어로도 번역했고 모든 언어에서 4.11로 업데이트했습니다.

### 기타
- 웹사이트: 지원 언어를 표시하고 펌웨어 파일 목록을 없앴습니다.
- `go test`가 Windows에서 통과합니다(i18n_test.go 포함). 새 테스트는 LEGGIMI.txt, AUTORI_E_LICENZE.txt와 라이선스가 기본 폴더와 «sorgenti»에서 같은지 확인합니다.

새 형식은 제조사 프로그램의 코드와 CHIRP 드라이버에서 가져왔고 가상 무전기로 시험했지만 실제 무전기로는 시험하지 않았습니다. 백업은 항상 보관하세요.

갤러리: 44개 제조사 489개 모델(채널 451, 설정 396, 펌웨어 업데이트 57).
