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
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub><b>हिन्दी</b></sub></td>
</tr>
</table>
</div>

GitHub पर रिलीज़ पृष्ठ (अंग्रेज़ी में), डाउनलोड के साथ: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

Microsoft Windows 10/11 के लिए IU0VYT Flash V.4.11.0

### ज़ोन के नाम
- «ज़ोन 1», «ज़ोन 2»… की जगह हर ज़ोन का एक नाम हो सकता है, निर्माता के प्रोग्रामों की तरह: «✎ ज़ोन के नाम…» बटन या ज़ोन पर डबल क्लिक। ज़ोन पट्टी और नीचे की पंक्ति «संख्या · नाम» दिखाती है।
- रेडियो: Radtel RT-950 Pro, Bajeton BJ-9000 और Hiroyasu HI-UV850 Pro (फ़र्मवेयर 0.23 तक और 0.24 से), Radtel RT-950, Radtel RT-920, बिना Bluetooth के Radtel RT-910 / Bajeton BJ-7700; Baofeng के DMR रेडियो RD-5R, DM-1801 और DM-32UV के ज़ोन।
- नाम चैनलों के साथ केबल से पढ़े और लिखे जाते हैं और .img फ़ाइल में सहेजे जाते हैं। लंबाई और अक्षर निर्माता के प्रोग्रामों जैसे (RT-950 Pro पर 12 अक्षर, चीनी भी; RT-920 पर 10; DMR पर 16)।
- सुरक्षा: नामों वाला हिस्सा तभी लिखा जाता है जब नाम बदलें, और उससे पहले रेडियो से दोबारा पढ़ा जाता है; अगर वहाँ ऐसा डेटा है जो नाम नहीं है, तो कुछ नहीं लिखा जाता।

### एक साथ संपादन
- «✎ एक साथ संपादन…» (राइट-क्लिक से भी) चिह्नित (या चुने हुए) चैनलों की पावर, बैंडविड्थ, स्कैन, TX टोन, RX टोन, शिफ़्ट और ट्रांसमिशन एक साथ बदलता है। «न बदलें» पर छोड़े गए विकल्प जैसे हैं वैसे रहते हैं; Ctrl+Z सब कुछ वापस करता है।

### दो और भाषाएँ (कुल 11)
- कोरियाई, दक्षिण कोरिया के झंडे के साथ।
- अरबी, संयुक्त अरब अमीरात के झंडे के साथ और दाएँ से बाएँ पृष्ठ (आवृत्तियाँ और संख्याएँ बाएँ से दाएँ ही रहती हैं)।
- `lingue` फ़ोल्डर की सभी जानकारी फ़ाइलें कोरियाई और अरबी में भी अनूदित हैं और सभी भाषाओं में 4.11 तक अपडेट हैं।

### अन्य
- वेबसाइट: उपलब्ध भाषाएँ बताई गईं; फ़र्मवेयर फ़ाइलों की सूची हटाई गई।
- `go test` Windows पर पास होता है (i18n_test.go भी); एक नया परीक्षण जाँचता है कि LEGGIMI.txt, AUTORI_E_LICENZE.txt और लाइसेंस मुख्य फ़ोल्डर और «sorgenti» में एक जैसे हैं।

नए फ़ॉर्मैट निर्माता के प्रोग्रामों के कोड और CHIRP ड्राइवर से निकाले गए हैं और सिम्युलेटेड रेडियो पर जाँचे गए हैं; असली रेडियो पर इन्हें नहीं जाँचा गया: बैकअप हमेशा रखें।

गैलरी: 44 ब्रांडों के 489 मॉडल (451 चैनल के साथ, 396 सेटिंग्स के साथ, 57 फ़र्मवेयर अपडेट के साथ)।
