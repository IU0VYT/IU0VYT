# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

GitHub 릴리스 페이지(영어), 내려받기 포함: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

Microsoft Windows 10/11용 IU0VYT Flash V.4.46.0

### 수정
- **Radtel RT-880G**: 더 이상 GMRS / FRS 무전기로 표시되지 않습니다. RT-880, RT-860과 같은 아마추어 무선용 VHF/UHF 송수신기(GPS, APRS 포함)이며, 쓰기 전에 «다음으로 판매되는 무전기 GMRS» 경고가 더 이상 나타나지 않습니다.
- **쓰기 후 확인**(Radtel RT-880G): 쓰기가 끝나면 데이터를 저장하고 다시 시작하는 무전기에 너무 일찍, 너무 자주 연결하려 해서 실패했습니다. 이제 프로그램은 무전기를 건드리지 않고 10초 기다린 뒤 15초마다 다시 시도하며, 최대 90초까지 기다립니다. 같은 드라이버를 쓰는 RT-880, RT-860, RT-817A / RT-871A에도 같은 대기 시간이 적용됩니다.
- AR8200, Kenwood, TK-180에서 받는 값을 더 엄격하게 검사합니다(CodeQL 지적).

### 참고
- 확인은 실제 RT-880G로 시험했습니다. RT-880, RT-860, RT-817A / RT-871A는 실제 장비로 시험하지 않았습니다. 다른 무전기의 대기 시간은 이전과 같습니다.
- 이 버전에는 새 무전기가 없습니다.
- 이 프로그램은 독립적이며 무전기 제조사와 관계가 없습니다.
- 전체 테스트 통과: 일반 부분, 구조 비교 915건과 CHIRP과 동일한 읽기 / 쓰기 380건.
