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
<td align="center"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub><b>Русский</b></sub></td>
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

Страница выпуска на GitHub (на английском), со ссылкой для загрузки: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 для Microsoft Windows 10/11

### Исправления
- **Radtel RT-880G**: больше не обозначается как радиостанция GMRS / FRS. Это любительский трансивер VHF/UHF (с GPS и APRS), как RT-880 и RT-860; предупреждение «Радиостанция продаётся как GMRS» больше не появляется перед записью.
- **Проверка после записи** у Radtel RT-880G: она завершалась ошибкой, потому что к радиостанции, которая по окончании записи сохраняет данные и перезагружается, обращались слишком рано и слишком часто. Теперь программа ждёт 10 секунд, не беспокоя её, и повторяет попытку каждые 15 секунд, до 90 секунд. Те же задержки действуют для RT-880, RT-860 и RT-817A / RT-871A, которые используют тот же драйвер.
- Более строгие проверки значений, получаемых от AR8200, Kenwood и TK-180 (замечания CodeQL).

### Примечания
- Проверка опробована на настоящей RT-880G; RT-880, RT-860 и RT-817A / RT-871A на настоящих радиостанциях не проверялись. Для остальных радиостанций задержки прежние.
- В этой версии нет новых радиостанций.
- Независимая программа, не связанная с производителями радиостанций.
- Полный набор тестов пройден: общая часть; 915 сравнений структуры и 380 чтений / записей, идентичных CHIRP.
