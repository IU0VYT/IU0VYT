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

Страница выпуска на GitHub (на английском), со ссылкой для загрузки: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 для Microsoft Windows 10/11

### Что нового
- **Retevis, 24 новые модели, полученные из программ производителя**: B3B, B63B, RT22P, RT622P, RB666, RB666P (та же память каналов, что у новой версии H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (семейство T18); RB65, RB66, RB66P (одна частота и один тон на канал). Подключение, чтение и запись написаны с нуля и проверены по заводским файлам данных этих программ. RB66 и RB66P, а также RB669, RT68H и RT668H отвечают одним и тем же кодом: нужную карточку надо выбрать вручную.
- **Настройки 24 моделей Retevis**: всего 241 пункт (шумоподавитель, ограничение времени передачи, VOX, сканирование, голосовые подсказки, экономия батареи, звук клавиш, боковые кнопки…). Предлагаются только те пункты, которые показывает программа данной модели; значения из 24 заводских файлов — все допустимые варианты.
- **Пароль программирования**: у 24 моделей Retevis и у новой версии Retevis H777 он не мешает ни чтению, ни записи, как в CHIRP (программа производителя, наоборот, останавливается, если пароль не совпадает с введённым пользователем). Это пункт настроек, открытым текстом: его можно увидеть, изменить и снять. Длина и символы — как в программе производителя: 6 символов (3 у RB45, RB645, RB47 и RB66; 4 у RB66P), цифры и буквы (только цифры у RB65 и RB66). Для B3H, B63H, RB45, RB645, RB66 и RB666, программа которых не показывает поле пароля, об этом напоминает предупреждение.
- **Настройки для радиостанций, у которых были только каналы**: семейство AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 и DMR-6X2 PRO: 27 пунктов, по описаниям памяти из qdmr и dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 и MD-9600, Retevis RT3S и RT82 (пункты Baofeng DM-1701); Radtel RT-880 и RT-950 Pro с прошивкой nicFW880 / nicFW950 (17 пунктов, по документации автора); TID TD-UV68 (пункты TD-H8); Radtel RT-890 с прошивкой M7OCM (53 пункта).
- **Тестовые версии**: копия с номером выше последней опубликованной стабильной версии показывает полосу предупреждения («Тестовая версия (бета-тестирование)…») и — только она — предлагает следующую тестовую версию, которая всегда скачивается с GitHub. Тот, кто пользуется стабильной версией, тестовых версий не видит.
- Галерея: 697 моделей 54 марок (647 с каналами, 596 с настройками).

### Примечания
- Ни одна из 24 радиостанций Retevis не проверена на настоящем устройстве — ни каналы, ни настройки: сначала только чтение, затем одна безобидная настройка. Если радиостанция отвечает неожиданно, чтение останавливается с сообщением, и в радиостанцию ничего не записывается.
- Если пароль программирования уже задан, программа всё равно пытается читать и записывать, как CHIRP: на настоящей радиостанции это не проверено.
- RB66 и RB66P, а также RB669, RT68H и RT668H отвечают одним и тем же кодом: нужную карточку надо выбрать вручную.
- Независимая программа, не связанная с производителями радиостанций.
- Полный набор тестов пройден: общая часть; 915 сравнений структуры и 380 чтений / записей, идентичных CHIRP.
