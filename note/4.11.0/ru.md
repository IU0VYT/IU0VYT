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

Страница выпуска на GitHub (на английском), со ссылкой для загрузки: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 для Microsoft Windows 10/11

### Названия зон
- Вместо «Зона 1», «Зона 2»… у каждой зоны может быть название, как в программах производителя: кнопка «✎ Названия зон…» или двойной щелчок по зоне. Панель зон и нижняя строка показывают «номер · название».
- Радиостанции: Radtel RT-950 Pro, Bajeton BJ-9000 и Hiroyasu HI-UV850 Pro (прошивка до 0.23 и начиная с 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 без Bluetooth; зоны DMR-радиостанций Baofeng RD-5R, DM-1801 и DM-32UV.
- Названия читаются и записываются по кабелю вместе с каналами и сохраняются в файле .img. Длина и символы — как в программах производителя (12 символов, в том числе китайские, на RT-950 Pro; 10 на RT-920; 16 на DMR).
- Безопасность: область названий записывается, только если названия изменились, после повторного чтения её из радиостанции; если там лежат данные, которые не являются названиями, ничего не записывается.

### Групповое изменение
- «✎ Групповое изменение…» (также через правую кнопку мыши) меняет сразу мощность, полосу, сканирование, тон TX, тон RX, сдвиг и передачу отмеченных (или выделенных) каналов. Параметры, оставленные на «не менять», остаются прежними; Ctrl+Z отменяет всё.

### Ещё два языка (всего 11)
- Корейский, с флагом Южной Кореи.
- Арабский, с флагом Объединённых Арабских Эмиратов и страницей справа налево (частоты и числа остаются слева направо).
- Все информационные файлы в папке `lingue` переведены и на корейский и арабский и обновлены до 4.11 на всех языках.

### Прочее
- Сайт: указаны доступные языки; убран список файлов прошивки.
- `go test` проходит в Windows (в том числе i18n_test.go); новый тест проверяет, что LEGGIMI.txt, AUTORI_E_LICENZE.txt и лицензия одинаковы в основной папке и в «sorgenti».

Новые форматы получены из кода программ производителя и из драйвера CHIRP и проверены на имитациях радиостанций; на настоящих радиостанциях они не проверялись: всегда сохраняйте резервные копии.

Галерея: 489 моделей 44 марок (451 с каналами, 396 с настройками, 57 с обновлением прошивки).
