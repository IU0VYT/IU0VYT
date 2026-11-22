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
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub><b>हिन्दी</b></sub></td>
</tr>
</table>
</div>

GitHub पर रिलीज़ पृष्ठ (अंग्रेज़ी में), डाउनलोड के साथ: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

Microsoft Windows 10/11 के लिए IU0VYT Flash V.4.46.0

### सुधार
- **Radtel RT-880G**: अब इसे GMRS / FRS रेडियो के रूप में नहीं दिखाया जाता। यह RT-880 और RT-860 की तरह शौकिया VHF/UHF ट्रांसीवर है (GPS और APRS के साथ); लिखने से पहले «रेडियो इस रूप में बेचा जाता है: GMRS» चेतावनी अब नहीं आती।
- **लिखने के बाद की जाँच** (Radtel RT-880G): यह विफल होती थी, क्योंकि लिखना पूरा होने पर डेटा सहेजकर फिर से चालू होने वाले रेडियो से बहुत जल्दी और बहुत बार संपर्क किया जाता था। अब प्रोग्राम रेडियो को छेड़े बिना 10 सेकंड रुकता है और हर 15 सेकंड पर दोबारा कोशिश करता है, 90 सेकंड तक। यही समय उसी ड्राइवर वाले RT-880, RT-860 और RT-817A / RT-871A पर भी लागू है।
- AR8200, Kenwood और TK-180 से मिलने वाले मानों की अधिक सख्त जाँच (CodeQL की सूचनाएँ)।

### नोट
- जाँच असली RT-880G पर परखी गई; RT-880, RT-860 और RT-817A / RT-871A असली उपकरण पर नहीं परखे गए। बाकी रेडियो के लिए समय पहले जैसा है।
- इस संस्करण में कोई नया रेडियो नहीं।
- स्वतंत्र प्रोग्राम, रेडियो निर्माताओं से संबद्ध नहीं।
- परीक्षणों का पूरा सेट पास: सामान्य हिस्सा; संरचना की 915 तुलनाएँ और CHIRP जैसे 380 पढ़ना / लिखना।
