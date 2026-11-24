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
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub><b>हिन्दी</b></sub></td>
</tr>
</table>
</div>

GitHub पर रिलीज़ पृष्ठ (अंग्रेज़ी में), डाउनलोड के साथ: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

Microsoft Windows 10/11 के लिए IU0VYT Flash V.4.47.0

### नया
- **Retevis, निर्माता के प्रोग्रामों से लिए गए 24 नए मॉडल**: B3B, B63B, RT22P, RT622P, RB666, RB666P (चैनल मेमोरी H777 के नए संस्करण जैसी); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (T18 परिवार); RB65, RB66, RB66P (हर चैनल में एक फ़्रीक्वेंसी और एक टोन)। कनेक्शन, पढ़ना और लिखना शून्य से दोबारा लिखे गए हैं और प्रोग्रामों की फ़ैक्टरी डेटा फ़ाइलों से जाँचे गए हैं। RB66 और RB66P, तथा RB669, RT68H और RT668H एक ही कोड से जवाब देते हैं: सही कार्ड हाथ से चुनना होगा।
- **24 Retevis मॉडलों की सेटिंग्स**: कुल 241 मदें (स्क्वेल्च, प्रसारण समय सीमा, VOX, स्कैन, आवाज़ मार्गदर्शन, बैटरी बचत, बीप, साइड कुंजियाँ…)। केवल वही मदें दी जाती हैं जो उस मॉडल का प्रोग्राम दिखाता है; 24 फ़ैक्टरी फ़ाइलों के सभी मान स्वीकृत विकल्प हैं।
- **प्रोग्रामिंग पासवर्ड**: 24 Retevis मॉडलों और Retevis H777 के नए संस्करण में यह पढ़ने या लिखने को नहीं रोकता, CHIRP की तरह (निर्माता का प्रोग्राम तब रुक जाता है जब पासवर्ड उपयोगकर्ता के लिखे पासवर्ड से अलग हो)। यह सेटिंग्स की एक मद है, साफ़ दिखती है: इसे देखा, बदला और हटाया जा सकता है। लंबाई और अक्षर निर्माता के प्रोग्राम जैसे ही हैं: 6 अक्षर (RB45, RB645, RB47 और RB66 में 3; RB66P में 4), अंक और अक्षर (RB65 और RB66 में केवल अंक)। B3H, B63H, RB45, RB645, RB66 और RB666 में, जिनका प्रोग्राम पासवर्ड का बॉक्स नहीं दिखाता, एक चेतावनी यह याद दिलाती है।
- **उन रेडियो की सेटिंग्स जिनमें अब तक केवल चैनल थे**: AnyTone परिवार (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 और DMR-6X2 PRO: 27 मदें, qdmr और dmrconfig के मेमोरी विवरणों से); TYT MD-UV380, MD-UV390, MD-2017 और MD-9600, Retevis RT3S और RT82 (Baofeng DM-1701 वाली मदें); nicFW880 / nicFW950 फ़र्मवेयर वाले Radtel RT-880 और RT-950 Pro (17 मदें, लेखक के दस्तावेज़ों से); TID TD-UV68 (TD-H8 वाली मदें); M7OCM फ़र्मवेयर वाला Radtel RT-890 (53 मदें)।
- **परीक्षण संस्करण**: जिस प्रति का नंबर प्रकाशित नवीनतम स्थिर संस्करण से बड़ा है वह एक चेतावनी पट्टी («परीक्षण संस्करण (बीटा परीक्षण)…») दिखाती है और केवल वही अगला परीक्षण संस्करण सुझाती है, जो हमेशा GitHub से डाउनलोड होता है। स्थिर संस्करण के उपयोगकर्ताओं को परीक्षण संस्करण नहीं दिखते।
- गैलरी: 54 ब्रांडों के 697 मॉडल (647 चैनल के साथ, 596 सेटिंग्स के साथ)।

### नोट
- 24 Retevis रेडियो में से कोई भी असली उपकरण पर परखा नहीं गया, न चैनल न सेटिंग्स: पहले केवल पढ़ें, फिर केवल एक हानिरहित सेटिंग बदलें। यदि रेडियो अनपेक्षित ढंग से जवाब देता है तो पढ़ना एक संदेश के साथ रुक जाता है और रेडियो में कुछ नहीं लिखा जाता।
- पहले से प्रोग्रामिंग पासवर्ड होने पर भी प्रोग्राम CHIRP की तरह पढ़ने और लिखने का प्रयास करता है: यह असली रेडियो पर परखा नहीं गया।
- RB66 और RB66P, तथा RB669, RT68H और RT668H एक ही कोड से जवाब देते हैं: सही कार्ड हाथ से चुनना होगा।
- स्वतंत्र प्रोग्राम, रेडियो निर्माताओं से संबद्ध नहीं।
- परीक्षणों का पूरा सेट पास: सामान्य हिस्सा; संरचना की 915 तुलनाएँ और CHIRP जैसे 380 पढ़ना / लिखना।
