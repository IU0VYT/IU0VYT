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
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub><b>हिन्दी</b></sub></td>
</tr>
</table>
</div>

GitHub पर रिलीज़ पृष्ठ (अंग्रेज़ी में), डाउनलोड के साथ: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

Microsoft Windows 10/11 के लिए IU0VYT Flash V.4.9.0

इस संस्करण में 4.8.0, 4.7.0 और 4.6.0 की नई बातें भी शामिल हैं, जो केवल परीक्षण संस्करण के रूप में प्रकाशित हुए थे।

## 4.9.0 में नया

### ठीक किया गया: Lanchonlh HG-UV98 और Quansheng TG-UV2+ के साथ «पहचानें»
- «पहचानें» के दौरान पोर्ट हमेशा 8N1 में बात करता था। अब यह सम पैरिटी (HG-UV98) और दो स्टॉप बिट (TG-UV2+) इस्तेमाल करता है, जैसे रेडियो के टैब से पढ़ते समय। दोनों रेडियो अब पहचाने जाते हैं।

### ठीक किया गया: HG-UV98 की पैरिटी
- सम पैरिटी के साथ Windows पोर्ट अब पैरिटी की गलतियाँ भी जाँचता है। पहले जाँच चालू होकर तुरंत बंद हो जाती थी।

### मुक्त फ़र्मवेयर वाले Quansheng UV-K5 / UV-K6 / UV-5R Plus
- egzumer, F4HWN और LOSEHU/LTS/UVE फ़र्मवेयर के लिए नए टैब।
- F4HWN, armel का F4HWN ड्राइवर (CHIRP से बाहर) इस्तेमाल करता है: पावर के 8 स्तर, ट्रांसमिशन लॉक, 3 स्कैन सूचियाँ और फ़र्मवेयर की सेटिंग्स।
- हर टैब केवल अपना फ़र्मवेयर पढ़ता और लिखता है। फ़र्मवेयर अपडेट के बाद प्रोग्राम बताता है कि कौन-सा टैब इस्तेमाल करना है।

### नए रेडियो
- Wouxun (20 मॉडल): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (अमेरिका)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (Radtel फ़र्मवेयर और मुक्त फ़र्मवेयर M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

CHIRP ड्राइवरों से आए रेडियो के लिए चैनल और सेटिंग्स CHIRP के मुकाबले स्वचालित परीक्षणों से जाँचे गए हैं। GMRS, FRS और MURS रेडियो अमेरिका के लिए हैं: इटली में उन आवृत्तियों पर ट्रांसमिट करने की अनुमति नहीं है।

## 4.8.0 में नया

### .BTU फ़र्मवेयर अपडेट (बिना Bluetooth के Radtel RT-910 और RT-900, बिना Bluetooth के Bajeton BJ-7700)
- .BTU फ़र्मवेयर «Radtel Bootload Tools» प्रोग्राम के प्रोटोकॉल से अपडेट होते हैं। बूटलोडर Baofeng UV-5RM वाला है, «PROGRAMBT80U» शुरुआती संपर्क के साथ।
- स्वचालित मोड: रेडियो सामान्य रूप से चालू। मैन्युअल मोड: रेडियो बंद, दोनों साइड बटन दबाए रखकर चालू करें, जब तक डिस्प्ले पर «UPDATE» न दिखे।
- बूटलोडर मॉडल नहीं जाँचता। इसलिए प्रोग्राम फ़ाइल के नाम से रेडियो और हार्डवेयर संस्करण पहचानता है, चेतावनियाँ दिखाता है और हमेशा मॉडल लिखने को कहता है।
- गैलरी में RT-910 (बिना Bluetooth, .BTU) और RT-910B (Bluetooth के साथ, .anc) के अब दो कार्ड हैं। BJ-7700 और बिना Bluetooth के BJ-7700 के लिए भी यही।

### ठीक किया गया: 512 चैनल वाले फ़र्मवेयर के साथ बिना Bluetooth का Radtel RT-900
- V1.09P तक रेडियो में 512 चैनल हैं; V1.14P से 999।
- अब चैनल 513-999 तभी लिखे जाते हैं जब आप उन्हें बदलें। पहले प्रोग्राम रेडियो में जाँचता है कि वह हिस्सा खाली है या उसमें चैनल हैं।

### नए रेडियो (CHIRP ड्राइवर)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (H777 का नया संस्करण), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

सभी के लिए चैनल और सेटिंग्स CHIRP के मुकाबले स्वचालित परीक्षणों से जाँचे गए हैं।

## 4.7.0 में नया

### ठीक किया गया: स्वचालित पहचान अटकना या बहुत धीमी होना
- फ़र्मवेयर अपडेट मोड में चालू रेडियो अपने आप समय-समय पर संदेश भेजता है: पहले ऐसा लगता था कि वह हर मॉडल को जवाब दे रहा है और खोज में कई मिनट लगते थे। अब प्रोग्राम कोशिश से पहले पोर्ट सुनता है और धोखा नहीं खाता: खोज एक मिनट से कम में पूरी होती है और पहचान टैब बताता है कि क्या हो रहा है।
- खोज शुरू होते समय इंटरफ़ेस अब नहीं अटकता; सबसे आम रेडियो पहले आज़माए जाते हैं।
- पहले बंद किए गए टैब के लिए झूठी सूचना «रेडियो टैब पहले से खुला है» ठीक की गई।

### एक ही तरह से संभाले जाने वाले रेडियो
- चयन विंडो और पहचान टैब सभी संभावित रेडियो दिखाते हैं (जैसे Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), चुनने के मेनू और इस छोटे संकेत के साथ कि रेडियो पर फ़र्मवेयर संस्करण कहाँ दिखता है।

### ठीक किया गया: फ़र्मवेयर 0.23 तक वाले Radtel RT-950 Pro / Bajeton BJ-9000
- चैनलों का «ट्रांसमिशन सक्षम» बिट अब हमेशा लिखा जाता है, दोनों संस्करणों में (Radtel के प्रोग्राम v1.2.x तक की तरह; v1.3.3 इसे अनदेखा करता है)। पुराने फ़र्मवेयर के साथ 990 चैनल वाले टैब में बदले गए चैनल ट्रांसमिट नहीं कर पाते थे। जिन चैनल 1-960 में यह बिट नहीं था, उनमें भी यह वापस लगाया जाता है।
- Bajeton BJ-9000 के असली संस्करण जाँचे गए: फ़र्मवेयर V0.22, V0.27 और V0.29 (Radtel वाले जैसा), प्रोग्राम v1.1.1 (960 चैनल) और v1.3.3 (990 चैनल, यह Radtel का प्रोग्राम है)। चैनल और सेटिंग्स Bajeton के प्रोग्रामों के कोड से भी जाँचे गए।

### दूसरे रेडियो के लिए फ़र्मवेयर अपडेट
- Radtel RT-950 Pro, Radtel RT-950 और Hiroyasu HI-UV850 Pro (.BTF फ़ाइलें, BJ-9000 वाला ही बूटलोडर)।
- Radtel RT-920 और RT-910B (.anc फ़ाइलें, BJ-7800 और BJ-7700 वाला ही प्रोटोकॉल)।

### नए रेडियो (CHIRP ड्राइवर)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); गैलरी में Hiroyasu HI-UV850 Pro भी।
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 और Anysecu AC-580 के लिए सेटिंग्स भी।

## 4.6.0 में नया

### ठीक किया गया: नए फ़र्मवेयर वाले Radtel RT-950 Pro / Bajeton BJ-9000
- फ़र्मवेयर **0.24** (दिसंबर 2025) से रेडियो में **99-99 के 10 ज़ोन में 990 चैनल** हैं (या 990 की एक सूची); **0.23** तक **64-64 के 15 ज़ोन में 960** थे (या 960 की एक सूची)। प्रोग्राम अब दोनों संस्करण संभालता है।
- संस्करण रेडियो से पढ़े गए डेटा से पहचाना जाता है (रेडियो खुद इसे नहीं बताता): पढ़ने पर चैनल सही टैब में खुलते हैं। अगर डेटा काफ़ी न हो, तो एक नोट बताता है कि रेडियो पर संस्करण कैसे जाँचें।
- चैनल 961-990 तभी लिखे जाते हैं जब वे बदलें, रेडियो से दोबारा पढ़ने के बाद; अगर रेडियो में वे नहीं हैं (फ़र्मवेयर 0.23 तक), तो कुछ नहीं लिखा जाता।
- चैनल और सेटिंग्स Radtel के प्रोग्राम v1.2.0 (960 चैनल) और v1.3.3 (990 चैनल) के मूल कोड से जाँचे गए।

### नया: Radtel RT-950 (बिना «Pro»)
- 960 चैनल (64-64 के 15 ज़ोन) या एक सूची, सामान्य सेटिंग्स; «BT-9000_CPS» v1.3.0 प्रोग्राम के मूल कोड से जाँचा गया।

### ठीक किया गया: रेडियो का सही नाम
- एक ही तरह से संभाले जाने वाले अलग-अलग रेडियो (जैसे **Radtel RT-950 Pro** और **Bajeton BJ-9000**, **Baofeng K67 Pro** और **5RH Pro**, Radtel RT-910 और Bajeton BJ-7700) अब टैब, पढ़ने, लिखने, पहचान और बैकअप फ़ाइलों में अपने नाम से बुलाए जाते हैं।
- इनमें से किसी रेडियो की पहली पहचान पर प्रोग्राम पूछता है कि वह कौन-सा है और फिर उसे याद रखता है।

### ठीक किया गया: बिना सहेजे डेटा की चेतावनी
- पढ़ने, लिखने या फ़ॉरवर्ड करने के बाद चैनल रेडियो में होते हैं: प्रोग्राम बंद करते समय चेतावनी तभी दिखती है जब सचमुच बदलाव हों।

### नए रेडियो (CHIRP ड्राइवर)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68।
- Radtel RT-920: 990 चैनल 99-99 के 10 ज़ोन में, जैसे रेडियो पर।

गैलरी: 44 ब्रांडों के 489 मॉडल (451 चैनल के साथ, 386 सेटिंग्स के साथ, 52 फ़र्मवेयर अपडेट के साथ)। नए ब्रांड: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone।
