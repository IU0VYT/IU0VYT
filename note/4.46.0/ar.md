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
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub><b>العربية</b></sub></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

صفحة الإصدار على GitHub (بالإنجليزية)، مع رابط التنزيل: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

<div dir="rtl">

IU0VYT Flash V.4.46.0 لنظام Microsoft Windows 10/11

### إصلاحات
- **Radtel RT-880G**: لم يعد يُعرض كجهاز GMRS / FRS. إنه جهاز إرسال واستقبال للهواة VHF/UHF (مع GPS وAPRS)، مثل RT-880 وRT-860؛ لم يعد التحذير «جهاز يُباع كـ GMRS» يظهر قبل الكتابة.
- **التحقق بعد الكتابة** في Radtel RT-880G: كان يفشل لأن الجهاز، الذي يحفظ البيانات ويُعاد تشغيله عند انتهاء الكتابة، كان يُتّصل به مبكرًا جدًا وكثيرًا جدًا. الآن ينتظر البرنامج 10 ثوانٍ دون إزعاجه ويعيد المحاولة كل 15 ثانية، حتى 90 ثانية. الأوقات نفسها تُطبَّق على RT-880 وRT-860 وRT-817A / RT-871A التي تستخدم التعريف نفسه.
- فحوص أكثر صرامة للقيم المستلمة من AR8200 وأجهزة Kenwood وTK-180 (تنبيهات CodeQL).

### ملاحظات
- جُرِّب التحقق على جهاز RT-880G حقيقي؛ ولم تُجرَّب RT-880 وRT-860 وRT-817A / RT-871A على أجهزة حقيقية. الأجهزة الأخرى تحتفظ بالأوقات السابقة.
- لا أجهزة جديدة في هذا الإصدار.
- برنامج مستقل، غير تابع للشركات المصنعة للأجهزة.
- اجتياز مجموعة الاختبارات كاملة: الجزء العام؛ و915 مقارنة للبنية و380 قراءة / كتابة مطابقة لـ CHIRP.

</div>
