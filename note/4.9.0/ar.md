# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

صفحة الإصدار على GitHub (بالإنجليزية)، مع رابط التنزيل: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

<div dir="rtl">

IU0VYT Flash V.4.9.0 لنظام Microsoft Windows 10/11

يتضمن هذا الإصدار أيضًا جديد الإصدارات 4.8.0 و4.7.0 و4.6.0، التي نُشرت كإصدارات تجريبية فقط.

## جديد الإصدار 4.9.0

### تم الإصلاح: «التعرّف» مع Lanchonlh HG-UV98 وQuansheng TG-UV2+
- أثناء «التعرّف» كان المنفذ يتكلم دائمًا بنمط 8N1. الآن يستخدم التكافؤ الزوجي (HG-UV98) وبتّي توقف (TG-UV2+)، كما في القراءة من علامة تبويب الجهاز. أصبح الجهازان يُتعرّف عليهما.

### تم الإصلاح: تكافؤ HG-UV98
- مع التكافؤ الزوجي يتحقق منفذ Windows الآن من أخطاء التكافؤ أيضًا. سابقًا كان التحقق يُفعّل ثم يُعطّل فورًا.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus مع البرامج الثابتة الحرة
- علامات تبويب جديدة لبرامج egzumer وF4HWN وLOSEHU/LTS/UVE الثابتة.
- يستخدم F4HWN تعريف F4HWN من armel (خارج CHIRP): 8 مستويات للقدرة، وقفل الإرسال، و3 قوائم مسح، وإعدادات البرنامج الثابت.
- كل علامة تبويب تقرأ وتكتب برنامجها الثابت فقط. بعد تحديث البرنامج الثابت يبيّن البرنامج علامة التبويب التي يجب استخدامها.

### أجهزة جديدة
- Wouxun (20 طرازًا): KG-UVD1P ‏(KG-UV2D وKG-UV3D) وKG-UV6D وKG-816 وKG-818 وKG-805G وKG-UV8D وKG-UV8D Plus وKG-UV8E وKG-UV9D Plus وKG-UV9PX وKG-UV9GX وKG-UV9K وKG-UV9G Pro وKG-935G وKG-935G Plus وKG-935H وKG-UV8H وKG-XS20G وKG-XS20G Plus وKG-XS20H
- TYT TH-UV8000D/E وUVL-15W وTH-350 (الولايات المتحدة)
- Abbree AR-518 وRuyage UV58Plus وRadtel RT-880G وRadtel RT-890 (برنامج Radtel الثابت والبرنامج الثابت الحر M7OCM)
- Retevis C2 وHA1G وHA1UV وHA2 وRB15 وRB615 وRB17P وRB28 وRB628 وRB48 وRB648 وRT23 وRT26 وRT76P وRT87 وRT16 وRB27B وRB27V
- BTECH GMRS-V1 وGMRS-V2 وMURS-V1 وMURS-V2 وFRS-A1
- AnyTone TERMN-8R وOBLTR-8R
- Puxing PX-888K وQYT KT-8R وRugged RH5R-V2 وBaojie BJ-UV55 وMTC UV-5R-3 وRadioddity MU-5 وZartek ZA-758 وTIDRADIO TD-M11 وTDXone TD-Q8A

بالنسبة إلى الأجهزة المأخوذة من تعريفات CHIRP، تم التحقق من القنوات والإعدادات بالاختبارات التلقائية مقارنةً بـ CHIRP. أجهزة GMRS وFRS وMURS مخصصة للولايات المتحدة: في إيطاليا لا يُسمح بالإرسال على تلك الترددات.

## جديد الإصدار 4.8.0

### تحديث البرنامج الثابت ‎.BTU ‏(Radtel RT-910 وRT-900 بدون Bluetooth، وBajeton BJ-7700 بدون Bluetooth)
- تُحدَّث البرامج الثابتة ‎.BTU ببروتوكول برنامج «Radtel Bootload Tools». محمّل الإقلاع هو نفسه في Baofeng UV-5RM، مع التحية «PROGRAMBT80U».
- الوضع التلقائي: الجهاز مشغّل بشكل عادي. الوضع اليدوي: الجهاز مطفأ، ثم يُشغَّل مع الضغط على المفتاحين الجانبيين حتى تُظهر الشاشة «UPDATE».
- محمّل الإقلاع لا يتحقق من الطراز. لذلك يتعرّف البرنامج على الجهاز وإصدار العتاد من اسم الملف، ويعرض التحذيرات، ويطلب دائمًا كتابة الطراز.
- في المعرض أصبح لـ RT-910 (بدون Bluetooth، ‎.BTU) وRT-910B (مع Bluetooth، ‎.anc) بطاقتان. والأمر نفسه لـ BJ-7700 وBJ-7700 بدون Bluetooth.

### تم الإصلاح: Radtel RT-900 بدون Bluetooth مع البرنامج الثابت ذي 512 قناة
- حتى V1.09P للجهاز 512 قناة؛ ومن V1.14P صار له 999.
- الآن لا تُكتب القنوات 513-999 إلا إذا عدّلتها. قبل ذلك يتحقق البرنامج في الجهاز من أن تلك المنطقة فارغة أو تحتوي على قنوات.

### أجهزة جديدة (تعريفات CHIRP)
- TYT TH-UV3R وTH-UV3R-25 وTH-UVF1 وTH-UVF8D
- KYD NC-630A وIP-620
- KSUN M6
- Puxing PX-777 وPX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24 وRT24V وH777S وRT1 وH777 V4 (النسخة الجديدة من H777) وRT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A وFD-268B وFD-288A وFD-288B وFD-150A وFD-160A وFD-450A وFD-460A وFD-460UH

لجميعها تم التحقق من القنوات والإعدادات بالاختبارات التلقائية مقارنةً بـ CHIRP.

## جديد الإصدار 4.7.0

### تم الإصلاح: تعرّف تلقائي متوقف أو بطيء جدًا
- الجهاز المشغّل في وضع تحديث البرنامج الثابت يرسل من تلقاء نفسه رسائل دورية: سابقًا كان يبدو كأنه يرد على كل طراز وكان البحث يستغرق دقائق. الآن يستمع البرنامج إلى المنفذ قبل المحاولة ولا ينخدع: ينتهي البحث في أقل من دقيقة، وتشرح علامة تبويب التعرّف ما يحدث.
- لم تعد الواجهة تتجمد أثناء بدء عمليات البحث؛ وتُجرَّب الأجهزة الأكثر انتشارًا أولًا.
- إصلاح التنبيه الخاطئ «علامة تبويب الجهاز مفتوحة من قبل» لعلامة تبويب أُغلقت سابقًا.

### أجهزة تُدار بالطريقة نفسها
- تعرض نافذة الاختيار وعلامة تبويب التعرّف كل الأجهزة الممكنة (مثل Radtel RT-950 Pro وBajeton BJ-9000 وHiroyasu HI-UV850 Pro)، مع قائمة للاختيار وإشارة قصيرة إلى مكان رؤية إصدار البرنامج الثابت على الجهاز.

### تم الإصلاح: Radtel RT-950 Pro / Bajeton BJ-9000 مع البرنامج الثابت حتى 0.23
- تُكتب بتة «الإرسال مسموح» للقنوات الآن دائمًا، مع الإصدارين (كبرامج Radtel حتى v1.2.x؛ أما v1.3.3 فيتجاهلها). مع البرنامج الثابت القديم قد لا ترسل القنوات المعدّلة في علامة تبويب الـ 990 قناة. وتُعاد البتة أيضًا في القنوات 1-960 التي كانت تفتقدها.
- التحقق من الإصدارات الفعلية لـ Bajeton BJ-9000: البرنامج الثابت V0.22 وV0.27 وV0.29 (مطابق لبرنامج Radtel)، والبرنامج v1.1.1 ‏(960 قناة) وv1.3.3 ‏(990 قناة، وهو برنامج Radtel). تم التحقق من القنوات والإعدادات أيضًا بشيفرة برامج Bajeton.

### تحديث البرنامج الثابت لأجهزة أخرى
- Radtel RT-950 Pro وRadtel RT-950 وHiroyasu HI-UV850 Pro (ملفات ‎.BTF، محمّل الإقلاع نفسه في BJ-9000).
- Radtel RT-920 وRT-910B (ملفات ‎.anc، البروتوكول نفسه في BJ-7800 وBJ-7700).

### أجهزة جديدة (تعريفات CHIRP)
- Radtel RT-620 وRadioddity GA-510 V2 وTYT TH-350 وRetevis RT18 وRetevis RB627B ‏(PMR446)؛ وفي المعرض أيضًا Hiroyasu HI-UV850 Pro.
- إعدادات أيضًا لـ SenhaiX 8800 وRadioddity GS-5B وCignus XTR-5 وAnysecu AC-580.

## جديد الإصدار 4.6.0

### تم الإصلاح: Radtel RT-950 Pro / Bajeton BJ-9000 مع البرنامج الثابت الحديث
- منذ البرنامج الثابت **0.24** (ديسمبر 2025) للجهاز **990 قناة في 10 مناطق من 99** (أو قائمة واحدة من 990)؛ وحتى **0.23** كان له **960 في 15 منطقة من 64** (أو قائمة واحدة من 960). يدير البرنامج الآن الإصدارين.
- يُتعرّف على الإصدار من البيانات المقروءة من الجهاز (الجهاز لا يصرّح به): تفتح القراءة القنوات في علامة التبويب الصحيحة. إذا لم تكفِ البيانات تشرح ملاحظة كيفية التحقق من الإصدار على الجهاز.
- لا تُكتب القنوات 961-990 إلا إذا تغيّرت، بعد إعادة قراءتها من الجهاز؛ وإذا لم تكن في الجهاز (البرنامج الثابت حتى 0.23) فلا يُكتب شيء.
- تم التحقق من القنوات والإعدادات بالشيفرة الأصلية لبرامج Radtel ‏v1.2.0 ‏(960 قناة) وv1.3.3 ‏(990 قناة).

### جديد: Radtel RT-950 (بدون «Pro»)
- 960 قناة (15 منطقة من 64) أو قائمة واحدة، وإعدادات عامة؛ تم التحقق منه بالشيفرة الأصلية لبرنامج «BT-9000_CPS» ‏v1.3.0.

### تم الإصلاح: الاسم الدقيق للجهاز
- الأجهزة المختلفة التي تُدار بالطريقة نفسها (مثل **Radtel RT-950 Pro** و**Bajeton BJ-9000**، و**Baofeng K67 Pro** و**5RH Pro**، وRadtel RT-910 وBajeton BJ-7700) صارت تُسمّى باسمها في علامة التبويب والقراءة والكتابة والتعرّف وملفات النسخ.
- عند أول تعرّف على أحد هذه الأجهزة يسأل البرنامج أيّها هو ثم يتذكّره.

### تم الإصلاح: تحذير البيانات غير المحفوظة
- بعد القراءة أو الكتابة أو التمرير تكون القنوات في الجهاز: عند إغلاق البرنامج لا يظهر التحذير إلا إذا وُجدت تعديلات حقيقية.

### أجهزة جديدة (تعريفات CHIRP)
- SenhaiX 8800 وRadioddity GS-5B وCignus XTR-5 وAnysecu AC-580 وTID TD-UV68.
- Radtel RT-920: القنوات الـ 990 في 10 مناطق من 99، كما على الجهاز.

المعرض: 489 طرازًا من 44 علامة تجارية (451 مع القنوات، 386 مع الإعدادات، 52 مع تحديث البرنامج الثابت). علامات تجارية جديدة: Wouxun وAnyTone وQYT وRugged وBaojie وMTC وZartek وRuyage وTDXone.

</div>
