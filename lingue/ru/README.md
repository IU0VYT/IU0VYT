# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub><b>Русский</b></sub></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="../../lingue/ja/README.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-11-orange" alt="11 languages">
</p>

**Версия 4.13.1** · Windows · интерфейс на 11 языках (italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية) · один исполняемый файл, без установки

Программа для радиолюбителей и пользователей радиостанций **44 марок** (**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone): читает и записывает каналы, частоты и настройки через кабель программирования, ведёт архив итальянских ретрансляторов и обновляет прошивку поддерживаемых радиостанций.

🌐 Сайт программы: **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*Это русский перевод [README.it.md](../../README.it.md); эталонным текстом является итальянский оригинал.*

## Возможности

- **Каналы и частоты** 451 модели (включая цифровые PMR и радиостанции DMR, в том числе по USB): чтение и запись по кабелю, файлы `.img`, совместимые с CHIRP, оригинальные файлы Baofeng K67 Pro (`.xlc`) и Bajeton 8300 (`.dat`).
- **Настройки радиостанций** (меню, дисплей, VOX, шумоподавление, TOT, DTMF, кнопки…) для 397 моделей, с описанием каждого пункта и предупреждением с подтверждением для важных пунктов.
- **Опознание подключённой радиостанции**: программа сама её распознаёт и показывает марку, модель, тип, прошивку и другие сообщаемые сведения (для USB-радиостанций DMR также DMR ID и имя). С **автоматическим опознанием** (кнопка на вкладке «Опознать», выбор запоминается) достаточно подключить кабель, на любой вкладке: программа слушает все порты кабелей одновременно, распознаёт и радиостанцию, включённую после подключения кабеля, и несколько радиостанций на разных кабелях, сначала пробует уже распознанные ранее и самые распространённые, не обманывается радиостанциями в режиме обновления прошивки и открывает вкладку каждой с прочитанными каналами; панель «Идёт опознание радиостанции» появляется только для кабеля с отвечающей радиостанцией; для другой радиостанции есть кнопка **«🔎 Опознать радиостанцию»** вверху справа.
- **Передача в другую радиостанцию**: отмеченные в одной радиостанции каналы записываются прямо в другую подключённую радиостанцию, выбранную по марке и модели, с проверкой совместимости, приспособлением, резервной копией и проверкой.
- **Универсальный конвертер**: каналы любой радиостанции или файла (K67 `.xlc`, Bajeton `.dat`, образы `.img`, в том числе из CHIRP, CSV из CHIRP, Excel/CSV) в формат любой другой радиостанции с учётом её правил.
- **Интерфейс на 11 языках** с переключателем флагов рядом с «ℹ Info»: мгновенный перевод всей программы (арабский справа налево), язык запоминается при перезапуске; информационные файлы переведены в папке `lingue`.
- **Названия зон** вместо «Зона 1», «Зона 2»… (Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000 и BJ-7700, Hiroyasu HI-UV850 Pro; зоны раций DMR RD-5R, DM-1801 и DM-32UV), читаются и записываются по кабелю вместе с каналами.
- **Групповое изменение** отмеченных каналов: мощность, полоса, сканирование, субтоны, сдвиг или передача сразу, остальное в каждом канале остаётся как есть.
- **Галерея из 489 моделей** 44 марок, разделённых на совместимые и несовместимые, по типам (любительские, DMR, GMRS/FRS, PMR446, детские) и маркам, с поиском, фильтрами и ссылками на официальные программы и руководства.
- **Архив ретрансляторов** на основе актуального списка [IK2ANE](https://www.ik2ane.it), с фильтрами и вставкой в радиостанции, с проверкой совместимости ретрансляторов и радиостанции перед записью.
- **Обновление прошивки** 57 моделей (в том числе Quansheng UV-K5, UV-K6 и UV-5R Plus, Radtel RT-590 и RT-600 с файлами .bin от Quansheng и Radtel, Radtel RT-470, RT-470L, RT-470X, RT-495 и RT-630 с файлами .kdhx от Radtel, TIDRADIO TD-H3, TD-H8 и Radtel RT-730 с файлами .BIN производителя, Retevis RA79 V3 и RA89 с файлами .bin и Retevis RT85 с файлами .hex от Retevis, Radtel RT-950 Pro и RT-950 и Bajeton BJ-9000 с файлами .BTF, Radtel RT-920 и RT-910B с файлами .anc, Radtel RT-910 и RT-900 и Bajeton BJ-7700 без Bluetooth с файлами .BTU, Radtel RT-890 с файлами .bin от Radtel и M7OCM, Baofeng DM-1701 по USB с файлами .bin от Baofeng, Baofeng RD-5R и DM-1801 по USB с файлами .sgl), с проверкой файла и резервной копией радиостанции; пошаговое обновление официальной программой для BTECH DMR-6X2 и DMR-6X2 PRO.
- Автоматическая **резервная копия** перед каждой записью и проверка после неё.
- **Проверка обновлений** программы при каждом запуске.

## Использование

1. Скачайте [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (последняя версия) и распакуйте архив.
2. Запустите `IU0VYT Flash.exe` (Windows 10/11; используется компонент Windows WebView2).
3. На вкладке **📻 Модели** выберите свою радиостанцию, подключите кабель и нажмите **Прочитать из радиостанции**.

Полное руководство — в файле `GUIDE.txt` (итальянский оригинал: `LEGGIMI.txt`).

## Сборка

Нужен Go 1.22 или новее. В папке `sorgenti`:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

Библиотеки включены в папку `vendor`.

## Предупреждения

- Протоколы радиостанций восстановлены для обеспечения совместимости и проверены на имитациях: **всегда храните резервные копии**. Программа предоставляется без каких-либо гарантий.
- Независимая программа, **не связанная с производителями радиостанций и не одобренная ими** (Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO и другие). Упомянутые товарные знаки принадлежат их владельцам и служат лишь для указания совместимых радиостанций.
- Соблюдайте правила использования частот в своей стране.

## Благодарности

- [CHIRP](https://chirpmyradio.com) и его авторам: большая часть драйверов радиостанций происходит из их работы.
- [IK2ANE](https://www.ik2ane.it), который бесплатно предоставляет всем список итальянских ретрансляторов.

Авторы и лицензии компонентов: `AUTHORS_AND_LICENSES.txt`.

## Лицензия

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Свободная программа, распространяемая на условиях **GNU General Public License версии 3 или более поздней** (`GPL-3.0-or-later`): см. `LICENSE.txt` и `LICENSE_NOTICE.txt`.
