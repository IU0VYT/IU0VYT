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

Страница выпуска на GitHub (на английском), со ссылкой для загрузки: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 для Microsoft Windows 10/11

### Прошивка Retevis RT85 (44 модели с обновлением прошивки)
- **Retevis RT85** с файлами .hex от Retevis, например «RT85-Firmware-V1.53.hex». Используется протокол программы Retevis «RT85 10W_5W_UpData», запущенной с имитацией радиостанции: отправляемые блоки и команды совпадают байт в байт.
- Порядок Retevis: радиостанция выключена, включается с нажатыми **PTT и P1** (зелёный индикатор). Затем выбирается мощность радиостанции (**5 Вт или 10 Вт**, как в программе Retevis) и для подтверждения вводится RT85.

### Исправлено: обновление Retevis RA89
- Версия 4.4 ждала от радиостанции неверного ответа на приветствие: с настоящей RA89 обновление не начиналось (радиостанция не затрагивалась). Теперь ожидаемый ответ и паузы такие же, как в программе Retevis.

### Порты, вкладки «Опознать» и «Конвертировать»
- Порты кабелей всегда идут по возрастанию, в любой системе: COM1, COM2, COM3… COM10. Автоопознание тоже начинается с первого порта.
- Вкладка «Опознать»: «Радиостанции, опознанные ранее, распознаются быстрее».
- Вкладка «Конвертировать»: «Выберите файл Baofeng, Bajeton, .img любых других марок радиостанций (в том числе сохранённый в CHIRP), CSV из CHIRP, таблицы Excel или CSV».

### Сайт
- Убран текст примечаний к версии под кнопкой загрузки.
- Заголовок поднят выше: номер версии под «PRESS START» виден целиком и на экранах ноутбуков.

Галерея: 386 моделей 27 марок (348 с каналами, 302 с настройками, 44 с обновлением прошивки).
