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
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub><b>العربية</b></sub></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

صفحة الإصدار على GitHub (بالإنجليزية)، مع رابط التنزيل: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

<div dir="rtl">

IU0VYT Flash V.4.11.0 لنظام Microsoft Windows 10/11

### أسماء المناطق
- بدلًا من «المنطقة 1» و«المنطقة 2»… يمكن أن يكون لكل منطقة اسم، كما في برامج الشركة المصنعة: زر «✎ أسماء المناطق…» أو نقرة مزدوجة على المنطقة. يعرض شريط المناطق والسطر السفلي «الرقم · الاسم».
- الأجهزة: Radtel RT-950 Pro وBajeton BJ-9000 وHiroyasu HI-UV850 Pro (البرنامج الثابت حتى 0.23 ومن 0.24)، وRadtel RT-950 وRadtel RT-920 وRadtel RT-910 / Bajeton BJ-7700 بدون Bluetooth؛ ومناطق أجهزة DMR من Baofeng ‏RD-5R وDM-1801 وDM-32UV.
- تُقرأ الأسماء وتُكتب عبر الكابل مع القنوات وتُحفظ في ملف ‎.img. الطول والأحرف كما في برامج الشركة المصنعة (12 حرفًا، حتى الصينية، على RT-950 Pro؛ و10 على RT-920؛ و16 على أجهزة DMR).
- الأمان: لا تُكتب منطقة الأسماء إلا إذا تغيّرت الأسماء، بعد إعادة قراءتها من الجهاز؛ وإذا كانت في ذلك الموضع بيانات ليست أسماء، فلا يُكتب شيء.

### تعديل جماعي
- «✎ تعديل جماعي…» (من زر الفأرة الأيمن أيضًا) يغيّر معًا القدرة وعرض النطاق والمسح ونغمة الإرسال ونغمة الاستقبال والإزاحة والإرسال للقنوات المحددة (أو المختارة). تبقى البنود المتروكة على «دون تغيير» كما هي؛ ويتراجع Ctrl+Z عن كل شيء.

### لغتان إضافيتان (11 في المجموع)
- الكورية، بعلم كوريا الجنوبية.
- العربية، بعلم الإمارات العربية المتحدة وصفحة من اليمين إلى اليسار (تبقى الترددات والأرقام من اليسار إلى اليمين).
- كل الملفات الإعلامية في مجلد `lingue` مترجمة أيضًا إلى الكورية والعربية ومحدّثة إلى 4.11 في جميع اللغات.

### أخرى
- الموقع: ذُكرت اللغات المتاحة، وأُزيلت قائمة ملفات البرنامج الثابت.
- يجتاز `go test` على Windows (بما في ذلك i18n_test.go)؛ واختبار جديد يتحقق من أن LEGGIMI.txt وAUTORI_E_LICENZE.txt والترخيص متطابقة في المجلد الرئيسي وفي «sorgenti».

الصيغ الجديدة مستخلصة من شيفرة برامج الشركة المصنعة ومن تعريف CHIRP ومختبرة بأجهزة محاكاة؛ ولم تُختبر بأجهزة حقيقية: احتفظ دائمًا بالنسخ الاحتياطية.

المعرض: 489 طرازًا من 44 علامة تجارية (451 مع القنوات، 396 مع الإعدادات، 57 مع تحديث البرنامج الثابت).

</div>
