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
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub><b>العربية</b></sub></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

صفحة الإصدار على GitHub (بالإنجليزية)، مع رابط التنزيل: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

<div dir="rtl">

IU0VYT Flash V.4.47.0 لنظام Microsoft Windows 10/11

### الجديد
- **Retevis، 24 طرازاً جديداً مأخوذة من برامج الشركة المصنّعة**: B3B، B63B، RT22P، RT622P، RB666، RB666P (ذاكرة القنوات نفسها في الإصدار الجديد من H777)؛ RB45، RB645، RB47؛ B3H، B63H، RB647، RT68H، RT668H، NR10، NR610، RB66S، RB666S، RT68B، RT668B، RB669 (عائلة T18)؛ RB65، RB66، RB66P (تردد واحد ونغمة واحدة لكل قناة). أُعيدت كتابة الاتصال والقراءة والكتابة من الصفر وجرى التحقق منها بملفات بيانات المصنع الخاصة بالبرامج. RB66 و RB66P، وكذلك RB669 و RT68H و RT668H، تجيب بالرمز نفسه: يجب اختيار البطاقة الصحيحة يدوياً.
- **إعدادات طرازات Retevis الـ 24**: 241 عنصراً في المجموع (الإسكات، مؤقّت حدّ الإرسال، VOX، المسح، الإرشاد الصوتي، توفير البطارية، صوت المفاتيح، المفاتيح الجانبية…). لا تُعرض إلا العناصر التي يُظهرها برنامج الطراز؛ وقيم ملفات المصنع الـ 24 كلها خيارات مسموح بها.
- **كلمة مرور البرمجة**: في طرازات Retevis الـ 24 وفي الإصدار الجديد من Retevis H777 لا تمنع القراءة ولا الكتابة، كما في CHIRP (أما برنامج الشركة المصنّعة فيتوقف إذا لم تكن كلمة المرور هي التي كتبها المستخدم). هي عنصر من الإعدادات، بنص واضح: يمكن رؤيتها وتغييرها وإزالتها. الطول والمحارف هي نفسها في برنامج الشركة المصنّعة: 6 محارف (3 في RB45 و RB645 و RB47 و RB66؛ 4 في RB66P)، أرقام وحروف (أرقام فقط في RB65 و RB66). وفي B3H و B63H و RB45 و RB645 و RB66 و RB666، التي لا يعرض برنامجها خانة كلمة المرور، يذكّر تحذير بذلك.
- **إعدادات لأجهزة لم يكن لها سوى القنوات**: عائلة AnyTone (AT-D868UV، AT-D878UV، AT-D878UV II Plus، BTECH DMR-6X2 و DMR-6X2 PRO: 27 عنصراً، من أوصاف الذاكرة في qdmr و dmrconfig)؛ و TYT MD-UV380 و MD-UV390 و MD-2017 و MD-9600، و Retevis RT3S و RT82 (عناصر Baofeng DM-1701)؛ و Radtel RT-880 و RT-950 Pro ببرنامج nicFW880 / nicFW950 الثابت (17 عنصراً، من توثيق المؤلف)؛ و TID TD-UV68 (عناصر TD-H8)؛ و Radtel RT-890 ببرنامج M7OCM الثابت (53 عنصراً).
- **الإصدارات التجريبية**: النسخة التي رقمها أعلى من آخر إصدار مستقر منشور تعرض شريط تحذير («إصدار تجريبي (اختبار بيتا)…»)، وهي وحدها تقترح الإصدار التجريبي التالي، ويُنزَّل دائماً من GitHub. من يستخدم الإصدار المستقر لا يرى الإصدارات التجريبية.
- المعرض: 697 طرازًا من 54 علامة (647 مع القنوات، 596 مع الإعدادات).

### ملاحظات
- لم يُجرَّب أيٌّ من أجهزة Retevis الـ 24 على جهاز حقيقي، لا القنوات ولا الإعدادات: ابدأ بالقراءة فقط، ثم غيّر إعداداً واحداً غير مؤذٍ. إذا أجاب الجهاز بطريقة غير متوقعة تتوقف القراءة برسالة ولا يُكتب شيء في الجهاز.
- عند وجود كلمة مرور للبرمجة مسبقاً يحاول البرنامج القراءة والكتابة مع ذلك، مثل CHIRP: لم يُجرَّب هذا على جهاز حقيقي.
- RB66 و RB66P، وكذلك RB669 و RT68H و RT668H، تجيب بالرمز نفسه: يجب اختيار البطاقة الصحيحة يدوياً.
- برنامج مستقل، غير تابع للشركات المصنعة للأجهزة.
- اجتياز مجموعة الاختبارات كاملة: الجزء العام؛ و915 مقارنة للبنية و380 قراءة / كتابة مطابقة لـ CHIRP.

</div>
